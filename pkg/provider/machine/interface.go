package machine

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/thoas/go-funk"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	ReasonWaiting      = "Waiting"
	ReasonSkip         = "Skip"
	ReasonFailedInit   = "FailedInit"
	ReasonFailedUpdate = "FailedUpdate"
	ReasonFailedDelete = "FailedDelete"

	ConditionTypeDone = "EnsureDone"
)

// Provider defines a set of response interfaces for specific machine
// types in machine management.
type Provider interface {
	Name() string

	Validate(machine *devopsv1.Machine) field.ErrorList

	PreCreate(machine *devopsv1.Machine) error
	AfterCreate(machine *devopsv1.Machine) error

	OnCreate(ctx *common.ClusterContext, machine *devopsv1.Machine) error
	OnUpdate(ctx *common.ClusterContext, machine *devopsv1.Machine) error
	OnDelete(ctx *common.ClusterContext, machine *devopsv1.Machine) error
}

var _ Provider = &DelegateProvider{}

type Handler func(*common.ClusterContext, *devopsv1.Machine) error

type DelegateProvider struct {
	ProviderName string

	ValidateFunc    func(machine *devopsv1.Machine) field.ErrorList
	PreCreateFunc   func(machine *devopsv1.Machine) error
	AfterCreateFunc func(machine *devopsv1.Machine) error

	CreateHandlers []Handler
	DeleteHandlers []Handler
	UpdateHandlers []Handler
}

func (p *DelegateProvider) Name() string {
	if p.ProviderName == "" {
		return "unknown"
	}
	return p.ProviderName
}

func (h Handler) Name() string {
	name := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	i := strings.Index(name, "Ensure")
	if i == -1 {
		return ""
	}
	return strings.TrimSuffix(name[i:], "-fm")
}

func (p *DelegateProvider) Validate(machine *devopsv1.Machine) field.ErrorList {
	if p.ValidateFunc != nil {
		return p.ValidateFunc(machine)
	}

	return nil
}

func (p *DelegateProvider) PreCreate(machine *devopsv1.Machine) error {
	if p.PreCreateFunc != nil {
		return p.PreCreateFunc(machine)
	}

	return nil
}

func (p *DelegateProvider) AfterCreate(machine *devopsv1.Machine) error {
	if p.AfterCreateFunc != nil {
		return p.AfterCreateFunc(machine)
	}

	return nil
}

func (p *DelegateProvider) OnCreate(ctx *common.ClusterContext, machine *devopsv1.Machine) error {
	condition, err := p.getCreateCurrentCondition(machine)
	if err != nil {
		return err
	}

	now := metav1.Now()
	if ctx.Cluster.Spec.Features.SkipConditions != nil &&
		funk.ContainsString(ctx.Cluster.Spec.Features.SkipConditions, condition.Type) {
		machine.SetCondition(devopsv1.MachineCondition{
			Type:               condition.Type,
			Status:             devopsv1.ConditionTrue,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Reason:             ReasonSkip,
			Message:            "Skip current condition",
		})
	} else {
		f := p.getCreateHandler(condition.Type)
		if f == nil {
			return fmt.Errorf("can't get handler by %s", condition.Type)
		}
		handlerName := f.Name()
		ctx.Info("OnCreate", "handlerName", handlerName)
		err = f(ctx, machine)
		if err != nil {
			ctx.Error(err, " OnCreate", "handlerName", handlerName)
			machine.SetCondition(devopsv1.MachineCondition{
				Type:          condition.Type,
				Status:        devopsv1.ConditionFalse,
				LastProbeTime: now,
				Message:       err.Error(),
				Reason:        ReasonFailedInit,
			})

			return err
		}

		machine.SetCondition(devopsv1.MachineCondition{
			Type:               condition.Type,
			Status:             devopsv1.ConditionTrue,
			LastProbeTime:      now,
			LastTransitionTime: now,
		})
	}

	nextConditionType := p.getNextConditionType(condition.Type)
	if nextConditionType == ConditionTypeDone {
		machine.Status.Phase = devopsv1.MachineRunning
	} else {
		machine.SetCondition(devopsv1.MachineCondition{
			Type:               nextConditionType,
			Status:             devopsv1.ConditionUnknown,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Message:            "waiting execute",
			Reason:             ReasonWaiting,
		})
	}
	return nil
}

func (p *DelegateProvider) OnUpdate(ctx *common.ClusterContext, machine *devopsv1.Machine) error {
	for _, f := range p.UpdateHandlers {
		ctx.Info("OnUpdate", "handlerName", f.Name())
		err := f(ctx, machine)
		if err != nil {
			return err
		}
	}

	machine.Status.Reason = ""
	machine.Status.Message = ""
	return nil
}

func (p *DelegateProvider) OnDelete(ctx *common.ClusterContext, machine *devopsv1.Machine) error {
	for _, f := range p.DeleteHandlers {
		ctx.Info("OnDelete", "handlerName", f.Name())
		err := f(ctx, machine)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *DelegateProvider) getNextConditionType(conditionType string) string {
	var (
		i int
		f Handler
	)
	for i, f = range p.CreateHandlers {
		name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		if strings.Contains(name, conditionType) {
			break
		}
	}
	if i == len(p.CreateHandlers)-1 {
		return ConditionTypeDone
	}
	next := p.CreateHandlers[i+1]

	return next.Name()
}

func (p *DelegateProvider) getCreateHandler(conditionType string) Handler {
	for _, f := range p.CreateHandlers {
		if conditionType == f.Name() {
			return f
		}
	}

	return nil
}

func (p *DelegateProvider) getCreateCurrentCondition(c *devopsv1.Machine) (*devopsv1.MachineCondition, error) {
	if c.Status.Phase == devopsv1.MachineRunning {
		return nil, errors.New("machine phase is running now")
	}
	if len(p.CreateHandlers) == 0 {
		return nil, errors.New("no create handlers")
	}

	if len(c.Status.Conditions) == 0 {
		return &devopsv1.MachineCondition{
			Type:          p.CreateHandlers[0].Name(),
			Status:        devopsv1.ConditionUnknown,
			LastProbeTime: metav1.Now(),
			Message:       "waiting process",
			Reason:        ReasonWaiting,
		}, nil
	}

	for _, condition := range c.Status.Conditions {
		if condition.Status == devopsv1.ConditionFalse || condition.Status == devopsv1.ConditionUnknown {
			return &condition, nil
		}
	}

	return nil, errors.New("no condition need process")
}

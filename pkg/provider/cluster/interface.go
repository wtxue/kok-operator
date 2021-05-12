package cluster

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/thoas/go-funk"
	devopsv1 "github.com/wtxue/kok-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kok-operator/pkg/constants"
	"github.com/wtxue/kok-operator/pkg/controllers/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/server/mux"
)

const (
	ReasonFailedProcess     = "FailedProcess"
	ReasonWaitingProcess    = "WaitingProcess"
	ReasonSuccessfulProcess = "SuccessfulProcess"
	ReasonSkipProcess       = "SkipProcess"

	ConditionTypeDone = "EnsureDone"
)

// Provider defines a set of response interfaces for specific cluster
// types in cluster management.
type Provider interface {
	Name() string

	RegisterHandler(mux *mux.PathRecorderMux)

	Validate(ctx *common.ClusterContext) field.ErrorList

	PreCreate(ctx *common.ClusterContext) error
	AfterCreate(ctx *common.ClusterContext) error

	OnCreate(ctx *common.ClusterContext) error
	OnUpdate(ctx *common.ClusterContext) error
	OnDelete(ctx *common.ClusterContext) error
}

var _ Provider = &DelegateProvider{}

type Handler func(ctx *common.ClusterContext) error

type DelegateProvider struct {
	ProviderName string

	ValidateFunc    func(ctx *common.ClusterContext) field.ErrorList
	PreCreateFunc   func(ctx *common.ClusterContext) error
	AfterCreateFunc func(ctx *common.ClusterContext) error

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

func (p *DelegateProvider) RegisterHandler(mux *mux.PathRecorderMux) {
}

func (p *DelegateProvider) Validate(ctx *common.ClusterContext) field.ErrorList {
	if p.ValidateFunc != nil {
		return p.ValidateFunc(ctx)
	}

	return nil
}

func (p *DelegateProvider) PreCreate(ctx *common.ClusterContext) error {
	if p.PreCreateFunc != nil {
		return p.PreCreateFunc(ctx)
	}

	return nil
}

func (p *DelegateProvider) AfterCreate(ctx *common.ClusterContext) error {
	if p.AfterCreateFunc != nil {
		return p.AfterCreateFunc(ctx)
	}

	return nil
}

func (p *DelegateProvider) OnCreate(ctx *common.ClusterContext) error {
	condition, err := p.getCreateCurrentCondition(ctx)
	if err != nil {
		return err
	}

	now := metav1.Now()
	if ctx.Cluster.Spec.Features.SkipConditions != nil &&
		funk.ContainsString(ctx.Cluster.Spec.Features.SkipConditions, condition.Type) {
		ctx.Cluster.SetCondition(devopsv1.ClusterCondition{
			Type:               condition.Type,
			Status:             devopsv1.ConditionTrue,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Reason:             ReasonSkipProcess,
		})
	} else {
		f := p.getCreateHandler(condition.Type)
		if f == nil {
			return fmt.Errorf("can't get handler by %s", condition.Type)
		}

		handlerName := f.Name()
		ctx.Info("onCreate", "handlerName", handlerName)
		if err = f(ctx); err != nil {
			ctx.Error(err, "OnCreate err", "handlerName", handlerName)
			ctx.Cluster.SetCondition(devopsv1.ClusterCondition{
				Type:          condition.Type,
				Status:        devopsv1.ConditionFalse,
				LastProbeTime: now,
				Message:       err.Error(),
				Reason:        ReasonFailedProcess,
			})
			ctx.Cluster.Status.Reason = ReasonFailedProcess
			ctx.Cluster.Status.Message = err.Error()
			return nil
		}

		ctx.Cluster.SetCondition(devopsv1.ClusterCondition{
			Type:               condition.Type,
			Status:             devopsv1.ConditionTrue,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Reason:             ReasonSuccessfulProcess,
		})
	}

	nextConditionType := p.getNextConditionType(condition.Type)
	if nextConditionType == ConditionTypeDone {
		ctx.Cluster.Status.Phase = devopsv1.ClusterRunning
	} else {
		ctx.Cluster.SetCondition(devopsv1.ClusterCondition{
			Type:               nextConditionType,
			Status:             devopsv1.ConditionUnknown,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Message:            "waiting process",
			Reason:             ReasonWaitingProcess,
		})
	}

	return nil
}

func tryFindHandler(handlerName string, handlers []string) bool {
	for i := range handlers {
		if strings.Contains(handlers[i], handlerName) {
			return true
		}
	}

	return false
}

func (p *DelegateProvider) OnUpdate(ctx *common.ClusterContext) error {
	if ctx.Cluster.Annotations == nil {
		return nil
	}

	var key string
	var ok bool
	if key, ok = ctx.Cluster.Annotations[constants.ClusterUpdateStep]; !ok {
		return nil
	}

	Handlers := strings.Split(key, ",")
	for _, f := range p.UpdateHandlers {
		handlerName := f.Name()
		if !tryFindHandler(handlerName, Handlers) {
			continue
		}

		ctx.Info("onUpdate", "handlerName", handlerName)
		now := metav1.Now()
		if err := f(ctx); err != nil {
			ctx.Error(err, "onUpdate err", "handlerName", handlerName)
			ctx.Cluster.SetCondition(devopsv1.ClusterCondition{
				Type:          handlerName,
				Status:        devopsv1.ConditionFalse,
				LastProbeTime: now,
				Message:       err.Error(),
				Reason:        ReasonFailedProcess,
			})
			ctx.Cluster.Status.Reason = ReasonFailedProcess
			ctx.Cluster.Status.Message = err.Error()
			return nil
		}

		ctx.Cluster.SetCondition(devopsv1.ClusterCondition{
			Type:               handlerName,
			Status:             devopsv1.ConditionTrue,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Reason:             ReasonSuccessfulProcess,
		})
	}

	return nil
}

func (p *DelegateProvider) OnDelete(ctx *common.ClusterContext) error {
	for _, f := range p.DeleteHandlers {
		handlerName := f.Name()
		ctx.Info("OnDelete", "handlerName", handlerName)
		err := f(ctx)
		if err != nil {
			ctx.Error(err, "OnDelete err", "handlerName", handlerName)
			return err
		}
	}

	return nil
}

func (h Handler) Name() string {
	name := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	i := strings.Index(name, "Ensure")
	if i == -1 {
		return ""
	}
	return strings.TrimSuffix(name[i:], "-fm")
}

func (p *DelegateProvider) getNextConditionType(conditionType string) string {
	var (
		i int
		f Handler
	)
	for i, f = range p.CreateHandlers {
		name := f.Name()
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

func (p *DelegateProvider) getCreateCurrentCondition(ctx *common.ClusterContext) (*devopsv1.ClusterCondition, error) {
	if ctx.Cluster.Status.Phase == devopsv1.ClusterRunning {
		return nil, errors.New("cluster phase is running now")
	}

	if len(p.CreateHandlers) == 0 {
		return nil, errors.New("no create handlers")
	}

	if len(ctx.Cluster.Status.Conditions) == 0 {
		return &devopsv1.ClusterCondition{
			Type:          p.CreateHandlers[0].Name(),
			Status:        devopsv1.ConditionUnknown,
			LastProbeTime: metav1.Now(),
			Message:       "waiting process",
			Reason:        ReasonWaitingProcess,
		}, nil
	}

	for _, condition := range ctx.Cluster.Status.Conditions {
		if condition.Status == devopsv1.ConditionFalse || condition.Status == devopsv1.ConditionUnknown {
			return &condition, nil
		}
	}

	if len(ctx.Cluster.Status.Conditions) < len(p.CreateHandlers) {
		return &devopsv1.ClusterCondition{
			Type:          p.CreateHandlers[len(ctx.Cluster.Status.Conditions)].Name(),
			Status:        devopsv1.ConditionUnknown,
			LastProbeTime: metav1.Now(),
			Message:       "waiting process",
			Reason:        ReasonWaitingProcess,
		}, nil
	}

	return nil, errors.New("no condition need process")
}

/*
Copyright 2020 wtxue.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"time"

	"github.com/thoas/go-funk"
	devopsv1 "github.com/wtxue/kube-on-kube-operator/pkg/apis/devops/v1"
	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/controllers/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/klog"
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

	Validate(cluster *common.Cluster) field.ErrorList

	PreCreate(cluster *common.Cluster) error
	AfterCreate(cluster *common.Cluster) error

	OnCreate(ctx context.Context, cluster *common.Cluster) error
	OnUpdate(ctx context.Context, cluster *common.Cluster) error
	OnDelete(ctx context.Context, cluster *common.Cluster) error
}

var _ Provider = &DelegateProvider{}

type Handler func(context.Context, *common.Cluster) error

type DelegateProvider struct {
	ProviderName string

	ValidateFunc    func(cluster *common.Cluster) field.ErrorList
	PreCreateFunc   func(cluster *common.Cluster) error
	AfterCreateFunc func(cluster *common.Cluster) error

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

func (p *DelegateProvider) Validate(cluster *common.Cluster) field.ErrorList {
	if p.ValidateFunc != nil {
		return p.ValidateFunc(cluster)
	}

	return nil
}

func (p *DelegateProvider) PreCreate(cluster *common.Cluster) error {
	if p.PreCreateFunc != nil {
		return p.PreCreateFunc(cluster)
	}

	return nil
}

func (p *DelegateProvider) AfterCreate(cluster *common.Cluster) error {
	if p.AfterCreateFunc != nil {
		return p.AfterCreateFunc(cluster)
	}

	return nil
}

func (p *DelegateProvider) OnCreate(ctx context.Context, cluster *common.Cluster) error {
	condition, err := p.getCreateCurrentCondition(cluster)
	if err != nil {
		return err
	}

	now := metav1.Now()
	if cluster.Spec.Features.SkipConditions != nil &&
		funk.ContainsString(cluster.Spec.Features.SkipConditions, condition.Type) {
		cluster.SetCondition(devopsv1.ClusterCondition{
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
		klog.Infof("clusterName: %s OnCreate handler: %s", cluster.Name, handlerName)
		err = f(ctx, cluster)
		if err != nil {
			klog.Errorf("cluster: %s OnCreate handler: %s err: %+v", cluster.Name, handlerName, err)
			cluster.SetCondition(devopsv1.ClusterCondition{
				Type:          condition.Type,
				Status:        devopsv1.ConditionFalse,
				LastProbeTime: now,
				Message:       err.Error(),
				Reason:        ReasonFailedProcess,
			})
			cluster.Cluster.Status.Reason = ReasonFailedProcess
			cluster.Cluster.Status.Message = err.Error()
			return nil
		}

		cluster.SetCondition(devopsv1.ClusterCondition{
			Type:               condition.Type,
			Status:             devopsv1.ConditionTrue,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Reason:             ReasonSuccessfulProcess,
		})
	}

	nextConditionType := p.getNextConditionType(condition.Type)
	if nextConditionType == ConditionTypeDone {
		cluster.Cluster.Status.Phase = devopsv1.ClusterRunning
	} else {
		cluster.SetCondition(devopsv1.ClusterCondition{
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

func tryFindHandler(handlerName string, handlers []string, cluster *common.Cluster) bool {
	var obj *devopsv1.ClusterCondition
	for idx := range cluster.Cluster.Status.Conditions {
		c := &cluster.Cluster.Status.Conditions[idx]
		if c.Type == handlerName {
			ltime := c.LastProbeTime
			if c.Status == devopsv1.ConditionTrue && ltime.Add(2*time.Minute).After(time.Now()) {
				obj = c
			}
			break
		}
	}

	for _, name := range handlers {
		if name == handlerName {
			if obj == nil {
				return true
			} else {
				return false
			}
		}
	}

	return false
}

func (p *DelegateProvider) OnUpdate(ctx context.Context, cluster *common.Cluster) error {
	if cluster.Cluster.Annotations == nil {
		return nil
	}

	var key string
	var ok bool
	if key, ok = cluster.Cluster.Annotations[constants.ClusterAnnotationAction]; !ok {
		return nil
	}

	Handlers := strings.Split(key, ",")
	for _, f := range p.UpdateHandlers {
		handlerName := f.Name()
		if !tryFindHandler(handlerName, Handlers, cluster) {
			continue
		}

		klog.Infof("clusterName: %s OnUpdate handler: %s", cluster.Name, handlerName)
		now := metav1.Now()
		err := f(ctx, cluster)
		if err != nil {
			klog.Errorf("cluster: %s OnUpdate handler: %s err: %+v", cluster.Name, handlerName, err)
			cluster.SetCondition(devopsv1.ClusterCondition{
				Type:          handlerName,
				Status:        devopsv1.ConditionFalse,
				LastProbeTime: now,
				Message:       err.Error(),
				Reason:        ReasonFailedProcess,
			})
			cluster.Cluster.Status.Reason = ReasonFailedProcess
			cluster.Cluster.Status.Message = err.Error()
			return nil
		}

		cluster.SetCondition(devopsv1.ClusterCondition{
			Type:               handlerName,
			Status:             devopsv1.ConditionTrue,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Reason:             ReasonSuccessfulProcess,
		})
	}

	return nil
}

func (p *DelegateProvider) OnDelete(ctx context.Context, cluster *common.Cluster) error {
	for _, f := range p.DeleteHandlers {
		klog.Infof("clusterName: %s OnDelete handler: %s", cluster.Name, f.Name())
		err := f(ctx, cluster)
		if err != nil {
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

func (p *DelegateProvider) getCreateCurrentCondition(c *common.Cluster) (*devopsv1.ClusterCondition, error) {
	if c.Cluster.Status.Phase == devopsv1.ClusterRunning {
		return nil, errors.New("cluster phase is running now")
	}

	if len(p.CreateHandlers) == 0 {
		return nil, errors.New("no create handlers")
	}

	if len(c.Cluster.Status.Conditions) == 0 {
		return &devopsv1.ClusterCondition{
			Type:          p.CreateHandlers[0].Name(),
			Status:        devopsv1.ConditionUnknown,
			LastProbeTime: metav1.Now(),
			Message:       "waiting process",
			Reason:        ReasonWaitingProcess,
		}, nil
	}

	for _, condition := range c.Cluster.Status.Conditions {
		if condition.Status == devopsv1.ConditionFalse || condition.Status == devopsv1.ConditionUnknown {
			return &condition, nil
		}
	}

	if len(c.Cluster.Status.Conditions) < len(p.CreateHandlers) {
		return &devopsv1.ClusterCondition{
			Type:          p.CreateHandlers[len(c.Cluster.Status.Conditions)].Name(),
			Status:        devopsv1.ConditionUnknown,
			LastProbeTime: metav1.Now(),
			Message:       "waiting process",
			Reason:        ReasonWaitingProcess,
		}, nil
	}

	return nil, errors.New("no condition need process")
}

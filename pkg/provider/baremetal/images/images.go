package images

import (
	"reflect"
	"sort"

	"github.com/wtxue/kube-on-kube-operator/pkg/constants"
	"github.com/wtxue/kube-on-kube-operator/pkg/util/containerregistry"
)

type Components struct {
	ETCD               containerregistry.Image
	CoreDNS            containerregistry.Image
	Pause              containerregistry.Image
	NvidiaDevicePlugin containerregistry.Image
	Keepalived         containerregistry.Image

	GPUManager        containerregistry.Image
	Busybox           containerregistry.Image
	GPUQuotaAdmission containerregistry.Image
}

func (c Components) Get(name string) *containerregistry.Image {
	v := reflect.ValueOf(c)
	for i := 0; i < v.NumField(); i++ {
		v, _ := v.Field(i).Interface().(containerregistry.Image)
		if v.Name == name {
			return &v
		}
	}
	return nil
}

var components = Components{
	ETCD:       containerregistry.Image{Name: "etcd", Tag: "v3.3.18"},
	CoreDNS:    containerregistry.Image{Name: "coredns", Tag: "1.6.7"},
	Pause:      containerregistry.Image{Name: "pause", Tag: "3.2"},
	Keepalived: containerregistry.Image{Name: "keepalived", Tag: "2.0.16-r0"},

	Busybox: containerregistry.Image{Name: "busybox", Tag: "1.31.0"},
}

func List() []string {
	var items []string

	for _, version := range constants.K8sVersionsWithV {
		for _, name := range []string{"kube-apiserver", "kube-controller-manager", "kube-scheduler", "kube-proxy"} {
			items = append(items, containerregistry.Image{Name: name, Tag: version}.BaseName())
		}
	}

	v := reflect.ValueOf(components)
	for i := 0; i < v.NumField(); i++ {
		v, _ := v.Field(i).Interface().(containerregistry.Image)
		items = append(items, v.BaseName())
	}
	sort.Strings(items)

	return items
}

func Get() Components {
	return components
}

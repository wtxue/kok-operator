package cluster

import (
	"fmt"
	"sort"
	"sync"

	"k8s.io/apiserver/pkg/server/mux"
)

type CpManager struct {
	sync.RWMutex
	Cp map[string]Provider
}

func New() *CpManager {
	return &CpManager{
		Cp: make(map[string]Provider),
	}
}

// Register makes a provider available by the provided name.
// If Register is called twice with the same name or if provider is nil,
// it panics.
func (p *CpManager) Register(name string, provider Provider) {
	p.Lock()
	defer p.Unlock()
	if provider == nil {
		panic("cluster: Register provider is nil")
	}
	if _, dup := p.Cp[name]; dup {
		panic("cluster: Register called twice for provider " + name)
	}
	p.Cp[name] = provider
}

// RegisterHandler register all provider's hanlder.
func (p *CpManager) RegisterHandler(mux *mux.PathRecorderMux) {
	for _, p := range p.Cp {
		p.RegisterHandler(mux)
	}
}

// Providers returns a sorted list of the names of the registered providers.
func (p *CpManager) Providers() []string {
	p.RLock()
	defer p.RUnlock()
	var list []string
	for name := range p.Cp {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// GetProvider returns provider by name
func (p *CpManager) GetProvider(name string) (Provider, error) {
	p.RLock()
	defer p.RUnlock()
	provider, ok := p.Cp[name]
	if !ok {
		return nil, fmt.Errorf("cluster: unknown provider %q (forgotten import?)", name)

	}

	return provider, nil
}

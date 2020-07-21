package machine

import (
	"fmt"
	"sort"
	"sync"
)

type MpManager struct {
	sync.RWMutex
	Mp map[string]Provider
}

func New() *MpManager {
	return &MpManager{
		Mp: make(map[string]Provider),
	}
}

// Register makes a provider available by the provided name.
// If Register is called twice with the same name or if provider is nil,
// it panics.
func (p *MpManager) Register(name string, provider Provider) {
	p.Lock()
	defer p.Unlock()
	if provider == nil {
		panic("machine: Register provider is nil")
	}
	if _, dup := p.Mp[name]; dup {
		panic("machine: Register called twice for provider " + name)
	}
	p.Mp[name] = provider
}

// Providers returns a sorted list of the names of the registered providers.
func (p *MpManager) Providers() []string {
	p.RLock()
	defer p.RUnlock()
	var list []string
	for name := range p.Mp {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// GetProvider returns provider by name
func (p *MpManager) GetProvider(name string) (Provider, error) {
	p.RLock()
	defer p.RUnlock()
	provider, ok := p.Mp[name]
	if !ok {
		return nil, fmt.Errorf("machine: unknown provider %q (forgotten import?)", name)

	}

	return provider, nil
}

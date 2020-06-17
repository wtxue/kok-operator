package machine

import (
	"fmt"
	"sort"
	"sync"
)

var (
	providersMu sync.RWMutex
	providers   = make(map[string]Provider)
)

// Register makes a provider available by the provided name.
// If Register is called twice with the same name or if provider is nil,
// it panics.
func Register(name string, provider Provider) {
	providersMu.Lock()
	defer providersMu.Unlock()
	if provider == nil {
		panic("machine: Register provider is nil")
	}
	if _, dup := providers[name]; dup {
		panic("machine: Register called twice for provider " + name)
	}
	providers[name] = provider
}

// Providers returns a sorted list of the names of the registered providers.
func Providers() []string {
	providersMu.RLock()
	defer providersMu.RUnlock()
	var list []string
	for name := range providers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// GetProvider returns provider by name
func GetProvider(name string) (Provider, error) {
	providersMu.RLock()
	provider, ok := providers[name]
	providersMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("machine: unknown provider %q (forgotten import?)", name)

	}

	return provider, nil
}

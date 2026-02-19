// Package registry provides a type registry for goconfygen.
//
// Target projects register their config types via Provider implementations,
// which the CLI then uses to generate YAML templates and validate configs.
package registry

import "sync"

// Provider describes a config type that can be used by the generator.
type Provider interface {
	// ID returns the unique identifier for this config type.
	ID() string
	// New returns a pointer to a new zero-value instance of the config struct.
	New() any
}

var (
	mu        sync.RWMutex
	providers = make(map[string]Provider)
)

// Register adds a provider to the global registry.
// If a provider with the same ID already exists, it is overwritten.
func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers[p.ID()] = p
}

// Get retrieves a provider by its ID.
func Get(id string) (Provider, bool) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := providers[id]
	return p, ok
}

// List returns all registered provider IDs in no particular order.
func List() []string {
	mu.RLock()
	defer mu.RUnlock()
	ids := make([]string, 0, len(providers))
	for id := range providers {
		ids = append(ids, id)
	}
	return ids
}

package database

import (
	"sort"
	"sync"
)

// DriverFactory is a function that creates a new DBDriver instance
type DriverFactory func() DBDriver

// Registry manages database driver registration
type Registry struct {
	mu      sync.RWMutex
	drivers map[string]DriverFactory
}

// global registry instance
var globalRegistry = &Registry{
	drivers: make(map[string]DriverFactory),
}

// Register registers a driver factory with a name
func Register(name string, factory DriverFactory) error {
	return globalRegistry.Register(name, factory)
}

// Get returns a new driver instance by name
func Get(name string) (DBDriver, error) {
	return globalRegistry.Get(name)
}

// List returns all registered driver names
func List() []string {
	return globalRegistry.List()
}

// Has checks if a driver is registered
func Has(name string) bool {
	return globalRegistry.Has(name)
}

// Register registers a driver factory with a name
func (r *Registry) Register(name string, factory DriverFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.drivers[name]; exists {
		return ErrDriverAlreadyRegistered
	}

	r.drivers[name] = factory
	return nil
}

// Get returns a new driver instance by name
func (r *Registry) Get(name string) (DBDriver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.drivers[name]
	if !exists {
		return nil, ErrDriverNotFound
	}

	return factory(), nil
}

// List returns all registered driver names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.drivers))
	for name := range r.drivers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Has checks if a driver is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.drivers[name]
	return exists
}

// MustRegister registers a driver and panics on error
func MustRegister(name string, factory DriverFactory) {
	if err := Register(name, factory); err != nil {
		panic(err)
	}
}

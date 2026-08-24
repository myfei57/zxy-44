package turbine

import (
	"fmt"
	"sort"
	"sync"
)

// Registry maps turbine ids to their runtime state.
type Registry struct {
	mu    sync.RWMutex
	units map[string]*Turbine
}

// NewRegistry creates an empty turbine registry.
func NewRegistry() *Registry {
	return &Registry{units: make(map[string]*Turbine)}
}

// Register adds a turbine, rejecting duplicate ids.
func (r *Registry) Register(t *Turbine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.units[t.ID]; ok {
		return fmt.Errorf("turbine %s already registered", t.ID)
	}
	r.units[t.ID] = t
	return nil
}

// Get returns a turbine by id.
func (r *Registry) Get(id string) (*Turbine, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.units[id]
	return t, ok
}

// List returns all turbines sorted by name.
func (r *Registry) List() []*Turbine {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Turbine, 0, len(r.units))
	for _, t := range r.units {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Count returns the number of registered turbines.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.units)
}

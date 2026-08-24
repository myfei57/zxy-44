package ns

import (
	"fmt"
	"sort"
	"sync"
)

// Registry is the in-memory namespace that maps farms and turbine units to
// stable identifiers.
type Registry struct {
	mu    sync.RWMutex
	farms map[string]*Farm
	units map[string]*Unit
}

// NewRegistry creates an empty namespace registry.
func NewRegistry() *Registry {
	return &Registry{
		farms: make(map[string]*Farm),
		units: make(map[string]*Unit),
	}
}

// AddFarm registers a new wind farm.
func (r *Registry) AddFarm(name string) (*Farm, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	farm, err := NewFarm(name)
	if err != nil {
		return nil, err
	}
	r.farms[farm.ID] = farm
	return farm, nil
}

// AddUnit registers a new turbine unit under an existing farm.
func (r *Registry) AddUnit(farmID, name string) (*Unit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	farm, ok := r.farms[farmID]
	if !ok {
		return nil, fmt.Errorf("farm %s does not exist", farmID)
	}
	unit, err := NewUnit(farmID, name)
	if err != nil {
		return nil, err
	}
	r.units[unit.ID] = unit
	farm.AddUnit(unit.ID)
	return unit, nil
}

// LookupFarm returns a farm by id.
func (r *Registry) LookupFarm(id string) (*Farm, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	farm, ok := r.farms[id]
	return farm, ok
}

// LookupUnit returns a unit by id.
func (r *Registry) LookupUnit(id string) (*Unit, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	unit, ok := r.units[id]
	return unit, ok
}

// UnitsByFarm returns the turbine units of one farm sorted by name.
func (r *Registry) UnitsByFarm(farmID string) []*Unit {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Unit
	for _, unit := range r.units {
		if unit.FarmID == farmID {
			out = append(out, unit)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Farms returns all farms sorted by name.
func (r *Registry) Farms() []*Farm {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Farm, 0, len(r.farms))
	for _, farm := range r.farms {
		out = append(out, farm)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// FarmCount returns the number of registered farms.
func (r *Registry) FarmCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.farms)
}

// UnitCount returns the number of registered turbine units.
func (r *Registry) UnitCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.units)
}

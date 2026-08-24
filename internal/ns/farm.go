package ns

import (
	"fmt"

	"github.com/google/uuid"
)

// Farm groups the wind turbines of one wind farm under a stable identifier.
type Farm struct {
	ID    string
	Name  string
	Units []string
}

// NewFarm creates a farm with a fresh identifier.
func NewFarm(name string) (*Farm, error) {
	if name == "" {
		return nil, fmt.Errorf("farm name must not be empty")
	}
	return &Farm{ID: uuid.NewString(), Name: name}, nil
}

// AddUnit records a turbine as belonging to the farm.
func (f *Farm) AddUnit(unitID string) {
	for _, id := range f.Units {
		if id == unitID {
			return
		}
	}
	f.Units = append(f.Units, unitID)
}

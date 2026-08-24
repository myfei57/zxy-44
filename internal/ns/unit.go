package ns

import (
	"fmt"

	"github.com/google/uuid"
)

// Unit is a logical wind turbine within a farm.
type Unit struct {
	ID     string
	Name   string
	FarmID string
}

// NewUnit creates a turbine unit with a fresh identifier.
func NewUnit(farmID, name string) (*Unit, error) {
	if farmID == "" {
		return nil, fmt.Errorf("farm id must not be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("unit name must not be empty")
	}
	return &Unit{ID: uuid.NewString(), Name: name, FarmID: farmID}, nil
}

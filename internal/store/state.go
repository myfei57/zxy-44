package store

import (
	"time"
)

// StateSnapshot is the recoverable control state of the whole plant.
type StateSnapshot struct {
	UnitStates        map[string]string `json:"unit_states"`
	ActiveStrategies  map[string]string `json:"active_strategies"`
	SpeedSources      map[string]string `json:"speed_sources"`
	SavedAt           time.Time         `json:"saved_at"`
}

// SaveState persists the plant state snapshot for crash recovery.
func (s *Store) SaveState(snapshot StateSnapshot) error {
	return s.WriteJSON("state/plant.json", snapshot)
}

// LoadState reads the most recent plant state snapshot.
func (s *Store) LoadState() (StateSnapshot, error) {
	var snapshot StateSnapshot
	err := s.ReadJSON("state/plant.json", &snapshot)
	return snapshot, err
}

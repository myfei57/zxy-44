package store

import (
	"errors"
	"os"
)

// Recover loads the persisted plant state at startup. A missing snapshot is not
// an error: the controller falls back to an empty state and the console stays
// available for manual commissioning.
func (s *Store) Recover() (StateSnapshot, error) {
	snapshot, err := s.LoadState()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return StateSnapshot{
				UnitStates:       make(map[string]string),
				ActiveStrategies: make(map[string]string),
				SpeedSources:     make(map[string]string),
			}, nil
		}
		return StateSnapshot{}, err
	}
	if snapshot.UnitStates == nil {
		snapshot.UnitStates = make(map[string]string)
	}
	if snapshot.ActiveStrategies == nil {
		snapshot.ActiveStrategies = make(map[string]string)
	}
	if snapshot.SpeedSources == nil {
		snapshot.SpeedSources = make(map[string]string)
	}
	return snapshot, nil
}

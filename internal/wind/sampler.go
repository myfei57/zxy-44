package wind

import (
	"fmt"
	"sync"
	"time"

	"windctl/internal/quota"
)

// Sampler periodically measures the wind speed of one unit and keeps a rolling
// history for the control strategy.
type Sampler struct {
	unitID  string
	source  SourceFunc
	quota   *quota.Quota
	mu      sync.RWMutex
	history *History
}

// NewSampler creates a sampler for a unit backed by an anemometer source.
func NewSampler(unitID string, source SourceFunc, q *quota.Quota) *Sampler {
	return &Sampler{
		unitID:  unitID,
		source:  source,
		quota:   q,
		history: NewHistory(240),
	}
}

// Sample takes one wind measurement, honouring the sampling quota.
func (s *Sampler) Sample() (Sample, error) {
	if !s.quota.Acquire(s.unitID, 1) {
		return Sample{}, fmt.Errorf("sampling quota exhausted for %s", s.unitID)
	}
	speed, err := s.source()
	if err != nil {
		s.quota.Release(s.unitID, 1)
		return Sample{}, fmt.Errorf("read wind speed: %w", err)
	}
	sample := Sample{UnitID: s.unitID, Speed: speed, At: time.Now().UTC()}
	s.mu.Lock()
	s.history.Add(sample)
	s.mu.Unlock()
	return sample, nil
}

// Latest returns the most recent measurement of the sampler.
func (s *Sampler) Latest() (Sample, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.history.Latest()
}

// Average returns the rolling average wind speed.
func (s *Sampler) Average() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.history.Average()
}

// History exposes the retained samples for statistics.
func (s *Sampler) History() *History {
	return s.history
}

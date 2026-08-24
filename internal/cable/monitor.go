package cable

import (
	"fmt"
	"sync"
	"time"

	"windctl/internal/audit"
	"windctl/internal/store"
)

// Monitor tracks the accumulated cable twist of one turbine and raises an
// alert when the accumulated angle exceeds the configured threshold.
type Monitor struct {
	unitID     string
	mu         sync.Mutex
	accumulated float64
	baseline   float64
	threshold  float64
	buffer     *Buffer
	store      *store.Store
	auditor    *audit.Auditor
}

// NewMonitor creates a twist monitor for a unit.
func NewMonitor(unitID string, threshold float64, st *store.Store, a *audit.Auditor) *Monitor {
	return &Monitor{
		unitID:    unitID,
		threshold: threshold,
		buffer:    NewBuffer(128),
		store:     st,
		auditor:   a,
	}
}

// AddAngle accumulates a yaw rotation delta and records it in the buffer.
func (m *Monitor) AddAngle(delta float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accumulated += delta
	m.buffer.Add(m.accumulated, time.Now().UTC())
}

// Tripped reports whether the current twist exceeds the threshold.
func (m *Monitor) Tripped() bool {
	return ThresholdExceeded(m.Twist(), m.threshold)
}

// Evaluate checks the twist condition and writes an audit alert on trip.
func (m *Monitor) Evaluate() (bool, error) {
	tripped := m.Tripped()
	if tripped {
		meta := map[string]string{
			"accumulated": fmt.Sprintf("%.1f", m.Accumulated()),
			"baseline":    fmt.Sprintf("%.1f", m.Baseline()),
			"threshold":   fmt.Sprintf("%.1f", m.threshold),
		}
		if err := m.auditor.Record(m.unitID, audit.KindCable, "cable-twist-alert", meta); err != nil {
			return tripped, err
		}
	}
	return tripped, nil
}

// Accumulated returns the raw accumulated twist angle.
func (m *Monitor) Accumulated() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.accumulated
}

// Baseline returns the twist baseline after the last unwind.
func (m *Monitor) Baseline() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.baseline
}

// Buffer exposes the buffered angle samples.
func (m *Monitor) Buffer() *Buffer {
	return m.buffer
}

// UnitID returns the monitored turbine unit.
func (m *Monitor) UnitID() string {
	return m.unitID
}

package cable

import "fmt"

// Twist returns the effective twist angle relative to the unwind baseline.
// After an unwind the baseline equals the accumulated angle, so the effective
// twist starts from zero again.
func (m *Monitor) Twist() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.accumulated - m.baseline
}

// ResetBaseline marks the current accumulated angle as the new zero point of
// the twist measurement and durably acknowledges the unwind.
func (m *Monitor) ResetBaseline() error {
	m.mu.Lock()
	m.baseline = m.accumulated
	m.mu.Unlock()
	m.buffer.DropBeforeUnwind()
	if err := m.store.WriteAck("cable", "baseline-"+m.unitID); err != nil {
		return fmt.Errorf("ack cable baseline: %w", err)
	}
	if err := m.auditor.Record(m.unitID, "cable", "twist-baseline-reset", nil); err != nil {
		return fmt.Errorf("audit cable baseline: %w", err)
	}
	return nil
}

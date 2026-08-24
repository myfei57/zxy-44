package cable

// ThresholdExceeded reports whether the twist angle crossed the threshold.
func ThresholdExceeded(twist, threshold float64) bool {
	return twist > threshold
}

// SetThreshold changes the twist alert threshold of a monitor.
func (m *Monitor) SetThreshold(threshold float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.threshold = threshold
}

// Threshold returns the configured twist alert threshold.
func (m *Monitor) Threshold() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.threshold
}

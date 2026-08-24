package rotor

import "fmt"

// TripSink receives overspeed events from the rotor monitor.
type TripSink interface {
	Trip(reason string) error
}

// Monitor samples the rotor on a tick and raises the safety chain when the
// overspeed limit is exceeded.
type Monitor struct {
	rotor    *Rotor
	detector *OverspeedDetector
	sink     TripSink
}

// NewMonitor wires a rotor, detector and safety sink together.
func NewMonitor(r *Rotor, d *OverspeedDetector, sink TripSink) *Monitor {
	return &Monitor{rotor: r, detector: d, sink: sink}
}

// Tick reads one speed sample and trips the safety sink on overspeed.
func (m *Monitor) Tick() (float64, error) {
	value, err := m.rotor.Read()
	if err != nil {
		return 0, fmt.Errorf("rotor monitor tick: %w", err)
	}
	if m.detector.Trip() {
		if err := m.sink.Trip("overspeed"); err != nil {
			return value, fmt.Errorf("raise safety trip: %w", err)
		}
	}
	return value, nil
}

// Detector exposes the overspeed detector used by the monitor.
func (m *Monitor) Detector() *OverspeedDetector {
	return m.detector
}

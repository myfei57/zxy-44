package rotor

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// SourceFunc reads the current rotor speed from one sensor.
type SourceFunc func() (float64, error)

// SpeedSource is one named rotor speed sensor.
type SpeedSource struct {
	name string
	read SourceFunc
}

// NewSpeedSource wraps a sensor under a stable name.
func NewSpeedSource(name string, read SourceFunc) *SpeedSource {
	return &SpeedSource{name: name, read: read}
}

// Name returns the sensor name.
func (s *SpeedSource) Name() string {
	return s.name
}

// Read returns the current speed value of the sensor.
func (s *SpeedSource) Read() (float64, error) {
	return s.read()
}

// ErrNoActiveSource is returned when the rotor has no selected speed source.
var ErrNoActiveSource = errors.New("no active rotor speed source")

// Rotor keeps the primary and backup speed sensors of one turbine and selects
// the source the control strategy should read.
type Rotor struct {
	unitID  string
	mu      sync.Mutex
	primary *SpeedSource
	backup  *SpeedSource
	active  *SpeedSource
	speed   float64
	history *SpeedHistory
	brake   *Brake
}

// NewRotor creates a rotor whose active source starts as the primary sensor.
func NewRotor(unitID string, primary, backup *SpeedSource) *Rotor {
	return &Rotor{
		unitID:  unitID,
		primary: primary,
		backup:  backup,
		active:  primary,
		history: NewSpeedHistory(240),
		brake:   NewBrake(),
	}
}

// Switch rotates the active speed source to the other sensor.
func (r *Rotor) Switch() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active == r.primary {
		r.active = r.backup
	} else {
		r.active = r.primary
	}
	return nil
}

// Current returns the speed source the strategy should currently read.
func (r *Rotor) Current() *SpeedSource {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.primary
}

// Read samples the active speed source and records the value.
func (r *Rotor) Read() (float64, error) {
	source := r.Current()
	if source == nil {
		return 0, ErrNoActiveSource
	}
	value, err := source.Read()
	if err != nil {
		return 0, fmt.Errorf("read rotor speed: %w", err)
	}
	r.mu.Lock()
	r.speed = value
	r.history.Add(SpeedRecord{
		UnitID: r.unitID,
		Source: source.Name(),
		Value:  value,
		At:     time.Now().UTC(),
	})
	r.mu.Unlock()
	return value, nil
}

// Speed returns the most recently read rotor speed.
func (r *Rotor) Speed() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.speed
}

// SourceName returns the name of the currently selected sensor.
func (r *Rotor) SourceName() string {
	source := r.Current()
	if source == nil {
		return "none"
	}
	return source.Name()
}

// Brake exposes the mechanical brake of the rotor.
func (r *Rotor) Brake() *Brake {
	return r.brake
}

// SpeedHistory exposes the retained speed records.
func (r *Rotor) SpeedHistory() *SpeedHistory {
	return r.history
}

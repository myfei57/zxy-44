package turbine

import (
	"sync"

	"github.com/google/uuid"
)

// State is the operational state of a turbine.
type State int

const (
	StateIdle State = iota
	StateRunning
	StateFaulted
	StateTripped
	StateStopped
)

// String returns the human readable state name.
func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StateFaulted:
		return "faulted"
	case StateTripped:
		return "tripped"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// Turbine is the runtime state of one wind turbine.
type Turbine struct {
	ID       string
	Name     string
	FarmID   string
	mu       sync.Mutex
	state    State
	speed    float64
	pitch    float64
	azimuth  float64
	faults   *FaultLog
	resetSeq uint64
}

// NewTurbine creates a turbine with a stable identifier.
func NewTurbine(id, name, farmID string) *Turbine {
	if id == "" {
		id = uuid.NewString()
	}
	return &Turbine{
		ID:     id,
		Name:   name,
		FarmID: farmID,
		state:  StateIdle,
		faults: NewFaultLog(),
	}
}

// SetSpeed records the latest rotor speed of the turbine.
func (t *Turbine) SetSpeed(value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.speed = value
}

// SetPitch records the latest blade pitch angle.
func (t *Turbine) SetPitch(value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pitch = value
}

// SetAzimuth records the latest nacelle azimuth.
func (t *Turbine) SetAzimuth(value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.azimuth = value
}

// State returns the current operational state.
func (t *Turbine) State() State {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state
}

// Enter transitions the turbine into a new state.
func (t *Turbine) Enter(state State) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state = state
}

// Faults exposes the fault log of the turbine.
func (t *Turbine) Faults() *FaultLog {
	return t.faults
}

// ResetSeq returns how many successful resets happened.
func (t *Turbine) ResetSeq() uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.resetSeq
}

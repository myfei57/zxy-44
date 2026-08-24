package rotor

import (
	"errors"
	"sync"
)

// ErrNotFeathered is returned when the mechanical brake is engaged before the
// blades reached the feathered position.
var ErrNotFeathered = errors.New("blades are not feathered")

// Brake is the mechanical rotor brake with a feather interlock.
type Brake struct {
	mu        sync.Mutex
	applied   bool
	feathered bool
	seq       int
}

// NewBrake creates a released brake with no feather state.
func NewBrake() *Brake {
	return &Brake{}
}

// SetFeathered records whether the blades reached the feathered position.
func (b *Brake) SetFeathered(v bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.feathered = v
}

// Engage applies the mechanical brake. The blades must be feathered first so
// the gearbox does not take the full rotor impact.
func (b *Brake) Engage() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.applied = true
	b.seq++
	return nil
}

// Release disengages the mechanical brake.
func (b *Brake) Release() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.applied = false
	b.seq++
	return nil
}

// Applied reports whether the brake is currently engaged.
func (b *Brake) Applied() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.applied
}

// Feathered reports whether the blades were marked as feathered.
func (b *Brake) Feathered() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.feathered
}

// Sequence returns how many brake operations happened.
func (b *Brake) Sequence() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.seq
}

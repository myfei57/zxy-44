package turbine

import "errors"

// ErrSafetyLatched is returned when a reset is attempted while the safety
// chain is still latched.
var ErrSafetyLatched = errors.New("safety chain still latched")

// Reset brings a tripped turbine back to running. The caller must confirm the
// safety chain released; a reset with a latched chain is rejected and the
// fault log is only cleared as part of an accepted reset.
func (t *Turbine) Reset(released bool) error {
	if !released || t.State() != StateTripped {
		return ErrSafetyLatched
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state = StateRunning
	t.faults.Clear()
	t.resetSeq++
	return nil
}

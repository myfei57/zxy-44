package safety

import (
	"errors"
	"fmt"

	"windctl/internal/audit"
	"windctl/internal/turbine"
)

// ErrNotReleased is returned when a reset is attempted before the safety chain
// released.
var ErrNotReleased = errors.New("safety chain has not released")

// DoReset coordinates a safety reset across the chain and the turbine. The
// fault log is cleared only as part of an accepted reset, and a concurrent
// fault report is never lost.
func DoReset(c *Chain, t *turbine.Turbine) error {
	if !c.Released() {
		return ErrNotReleased
	}
	if err := t.Reset(true); err != nil {
		return fmt.Errorf("reset turbine: %w", err)
	}
	return c.MarkReset()
}

// MarkReset records one successful reset on the chain and acknowledges it.
func (c *Chain) MarkReset() error {
	c.mu.Lock()
	c.resetCount++
	err := c.persistLocked()
	c.mu.Unlock()
	if err != nil {
		return fmt.Errorf("persist safety reset: %w", err)
	}
	if err := c.store.WriteAck("safety", "reset-"+c.unitID); err != nil {
		return fmt.Errorf("ack safety reset: %w", err)
	}
	if err := c.auditor.Record(c.unitID, audit.KindReset, "safety-reset", nil); err != nil {
		return fmt.Errorf("audit safety reset: %w", err)
	}
	return nil
}

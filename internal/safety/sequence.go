package safety

import (
	"fmt"

	"windctl/internal/audit"
)

// StopExecutor performs the physical actions of an emergency stop.
type StopExecutor interface {
	Feather() error
	Brake() error
}

// RunStop executes the emergency stop sequence: the blades feather first to
// shed the rotor load, then the mechanical brake engages. Reversing the order
// lets the gearbox take the full rotor impact.
func RunStop(e StopExecutor, c *Chain, reason string) error {
	if err := e.Feather(); err != nil {
		return fmt.Errorf("feather on stop: %w", err)
	}
	if err := e.Brake(); err != nil {
		return fmt.Errorf("brake on stop: %w", err)
	}
	if err := c.auditor.Record(c.UnitID(), audit.KindBrake, "emergency-stop", map[string]string{"reason": reason}); err != nil {
		return fmt.Errorf("audit emergency stop: %w", err)
	}
	return nil
}

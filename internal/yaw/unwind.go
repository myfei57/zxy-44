package yaw

import (
	"fmt"

	"windctl/internal/audit"
)

// Unwind runs the cable unwind sequence: the twist baseline is reset so the
// protection starts counting from zero again, then the action is acknowledged
// and audited.
func (y *YawController) Unwind() error {
	if err := y.cable.ResetBaseline(); err != nil {
		return fmt.Errorf("reset cable baseline: %w", err)
	}
	if err := y.store.WriteAck("yaw", "unwind-"+y.unitID); err != nil {
		return fmt.Errorf("ack yaw unwind: %w", err)
	}
	if err := y.auditor.Record(y.unitID, audit.KindUnwind, "yaw-unwind", map[string]string{
		"azimuth": fmt.Sprintf("%.1f", y.Azimuth()),
	}); err != nil {
		return fmt.Errorf("audit yaw unwind: %w", err)
	}
	return nil
}

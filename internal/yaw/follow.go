package yaw

import (
	"fmt"

	"windctl/internal/audit"
	"windctl/internal/strategy"
)

// Follow rotates the nacelle towards the wind direction. The rotation is the
// shortest ring-aware turn, and the cable twist limit must be respected first.
func (y *YawController) Follow(windDir float64) error {
	if y.cable.Tripped() {
		return ErrCableTwist
	}
	target := strategy.NormalizeAzimuth(windDir)
	current := y.Azimuth()
	direction := strategy.Direction(current, target)
	y.SetAzimuth(current + direction)
	y.SetTarget(target)
	y.SetMoving(direction != 0)
	y.cable.AddAngle(direction)
	if err := y.auditor.Record(y.unitID, audit.KindYaw, "yaw-follow", map[string]string{
		"from":      fmt.Sprintf("%.1f", current),
		"to":        fmt.Sprintf("%.1f", target),
		"direction": fmt.Sprintf("%.1f", direction),
	}); err != nil {
		return fmt.Errorf("audit yaw follow: %w", err)
	}
	return nil
}

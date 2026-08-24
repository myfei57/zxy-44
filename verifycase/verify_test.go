package verifycase

import (
	"testing"

	"windctl/internal/audit"
	"windctl/internal/cable"
	"windctl/internal/store"
	"windctl/internal/yaw"
)

// TestWcCableAngleIgnoresPreUnwind verifies the twist evaluation filters the
// buffered angles recorded before the last unwind.
func TestWcCableAngleIgnoresPreUnwind(t *testing.T) {
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a := audit.New(st)
	mon := cable.NewMonitor("unit-01", 150, st, a)
	yawCtl := yaw.NewYawController("unit-01", 0, mon, a, st)

	mon.AddAngle(90)
	mon.AddAngle(90)
	if err := yawCtl.Unwind(); err != nil {
		t.Fatal(err)
	}
	mon.AddAngle(-90)

	peak, tripped, err := yawCtl.EvaluateTwist()
	if err != nil {
		t.Fatal(err)
	}
	if tripped {
		t.Fatalf("twist evaluation must ignore pre-unwind buffered angles, peak=%v", peak)
	}
	if peak > 100 {
		t.Fatalf("evaluated peak %v must come from post-unwind samples", peak)
	}
}

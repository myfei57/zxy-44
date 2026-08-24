package verifycase

import (
	"testing"

	"windctl/internal/audit"
	"windctl/internal/cable"
	"windctl/internal/store"
	"windctl/internal/strategy"
	"windctl/internal/yaw"
)

// TestWcAzimuthWrapDirectionStable verifies the azimuth stays on the ring and
// the yaw direction is the shortest turn across the 360 degree wrap.
func TestWcAzimuthWrapDirectionStable(t *testing.T) {
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a := audit.New(st)
	mon := cable.NewMonitor("unit-01", 150, st, a)
	yawCtl := yaw.NewYawController("unit-01", 355, mon, a, st)

	yawCtl.SetAzimuth(365)
	if got := yawCtl.Azimuth(); got != 5 {
		t.Fatalf("azimuth must wrap into the [0,360) ring, got %v", got)
	}
	d := strategy.Direction(355, 5)
	if d != 10 {
		t.Fatalf("direction must be the shortest ring-aware turn, got %v", d)
	}
	if d < -180 || d > 180 {
		t.Fatalf("direction %v is not a shortest turn", d)
	}
}

package verifycase

import (
	"testing"

	"windctl/internal/audit"
	"windctl/internal/safety"
	"windctl/internal/store"
	"windctl/internal/turbine"
)

// TestWcSafetyLatchReleasedOnRecover verifies the safety latch clears when the
// overspeed condition recovers and the turbine can reset afterwards.
func TestWcSafetyLatchReleasedOnRecover(t *testing.T) {
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a := audit.New(st)
	chain := safety.NewChain("unit-01", st, a)
	turb := turbine.NewTurbine("unit-01", "WTG-01", "farm-01")
	turb.Enter(turbine.StateTripped)

	if err := chain.Trip("overspeed"); err != nil {
		t.Fatal(err)
	}
	if !chain.Latched() {
		t.Fatal("chain must latch after a trip")
	}
	if err := chain.Release(); err != nil {
		t.Fatal(err)
	}
	if chain.Latched() {
		t.Fatal("safety latch must be cleared once the overspeed condition recovers")
	}
	if err := safety.DoReset(chain, turb); err != nil {
		t.Fatal(err)
	}
	if turb.State() != turbine.StateRunning {
		t.Fatalf("turbine must be running after a reset, got %s", turb.State())
	}
}

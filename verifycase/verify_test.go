package verifycase

import (
	"testing"

	"windctl/internal/rotor"
	"windctl/internal/strategy"
)

// TestWcStrategyUsesCurrentSpeedSource verifies the overspeed evaluation reads
// the currently selected speed source after a source switch.
func TestWcStrategyUsesCurrentSpeedSource(t *testing.T) {
	primary := rotor.NewSpeedSource("primary", func() (float64, error) { return 10, nil })
	backup := rotor.NewSpeedSource("backup", func() (float64, error) { return 80, nil })
	rot := rotor.NewRotor("unit-01", primary, backup)

	if _, err := rot.Read(); err != nil {
		t.Fatal(err)
	}
	e := strategy.NewEvaluator(rot)
	if err := rot.Switch(); err != nil {
		t.Fatal(err)
	}
	if got := rot.SourceName(); got != "backup" {
		t.Fatalf("active speed source must be backup after the switch, got %s", got)
	}
	if _, err := rot.Read(); err != nil {
		t.Fatal(err)
	}
	if !e.Check(50) {
		t.Fatal("overspeed evaluation must read the current backup source after the switch")
	}
}

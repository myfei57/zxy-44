package verifycase

import (
	"os"
	"testing"

	"windctl/internal/audit"
	"windctl/internal/blade"
	"windctl/internal/rotor"
	"windctl/internal/store"
	"windctl/internal/strategy"
)

type testSink struct {
	st *store.Store
}

func (s testSink) PersistConfig(cfg strategy.Config) error {
	return s.st.WriteConfig(cfg.Name, cfg.Payload)
}

func (s testSink) LoadConfig(name string) ([]byte, error) {
	return s.st.LoadConfig(name)
}

func (s testSink) HasConfig(name string) bool {
	return s.st.HasConfig(name)
}

// TestWcStrategyAfterConfigDurable verifies the active strategy only switches
// after its configuration is durably persisted.
func TestWcStrategyAfterConfigDurable(t *testing.T) {
	dir := t.TempDir()
	st, err := store.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir, []byte("occupied"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := audit.New(st)
	rot := rotor.NewRotor("unit-01",
		rotor.NewSpeedSource("primary", func() (float64, error) { return 10, nil }),
		rotor.NewSpeedSource("backup", func() (float64, error) { return 10, nil }),
	)
	ctl := blade.NewPitchController("unit-01", 0)
	sc := strategy.NewController("unit-01", testSink{st: st}, rot, ctl, a)

	next := strategy.Config{Name: "limit-x", Mode: strategy.ModeLimitPower, MaxPower: 60}
	if err := strategy.SwitchActive(sc, next); err == nil {
		t.Fatal("switch must fail when the config cannot be persisted durably")
	}
	if sc.Active().Name != "default" {
		t.Fatalf("active strategy must stay unchanged when the config write fails, got %s", sc.Active().Name)
	}
}

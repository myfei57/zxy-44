package verifycase

import (
	"sync"
	"testing"

	"windctl/internal/audit"
	"windctl/internal/blade"
	"windctl/internal/rotor"
	"windctl/internal/safety"
	"windctl/internal/store"
)

type stopRecorder struct {
	mu    sync.Mutex
	brake *rotor.Brake
	ctl   *blade.PitchController
	order []string
}

func (r *stopRecorder) Feather() error {
	r.mu.Lock()
	r.order = append(r.order, "feather")
	r.mu.Unlock()
	r.brake.SetFeathered(true)
	return blade.Feather(r.ctl)
}

func (r *stopRecorder) Brake() error {
	r.mu.Lock()
	r.order = append(r.order, "brake")
	r.mu.Unlock()
	return r.brake.Engage()
}

// TestWcStopFeathersBeforeBrake verifies the emergency stop feathers the
// blades before engaging the mechanical brake.
func TestWcStopFeathersBeforeBrake(t *testing.T) {
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a := audit.New(st)
	chain := safety.NewChain("unit-01", st, a)
	ctl := blade.NewPitchController("unit-01", 0)
	brake := rotor.NewBrake()
	rec := &stopRecorder{brake: brake, ctl: ctl}

	if err := safety.RunStop(rec, chain, "manual"); err != nil {
		t.Fatal(err)
	}
	if len(rec.order) != 2 || rec.order[0] != "feather" || rec.order[1] != "brake" {
		t.Fatalf("stop sequence must feather before braking, got %v", rec.order)
	}
	if !brake.Applied() {
		t.Fatal("mechanical brake must be engaged after the stop sequence")
	}
}

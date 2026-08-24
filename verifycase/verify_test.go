package verifycase

import (
	"fmt"
	"sync"
	"testing"

	"windctl/internal/audit"
	"windctl/internal/safety"
	"windctl/internal/store"
	"windctl/internal/turbine"
)

// TestWcConcurrentResetKeepsFault verifies a concurrent reset never corrupts
// the fault log while a fresh fault is being reported.
func TestWcConcurrentResetKeepsFault(t *testing.T) {
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
	if err := chain.Release(); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = safety.DoReset(chain, turb)
		}()
		go func(i int) {
			defer wg.Done()
			turb.RecordFault("E", fmt.Sprintf("fresh fault %d", i))
		}(i)
	}
	wg.Wait()

	seen := make(map[uint64]bool)
	for _, f := range turb.Faults().Recent(500) {
		if seen[f.Seq] {
			t.Fatalf("fault seq %d duplicated by concurrent reset", f.Seq)
		}
		seen[f.Seq] = true
	}
}

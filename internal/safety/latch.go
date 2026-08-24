package safety

import (
	"fmt"
	"time"

	"windctl/internal/audit"
	"windctl/internal/store"
)

// Trip latches the safety chain when a dangerous condition is detected.
func (c *Chain) Trip(reason string) error {
	c.mu.Lock()
	c.latched = true
	c.recovered = false
	c.tripCount++
	c.lastTrip = time.Now().UTC()
	err := c.persistLocked()
	c.mu.Unlock()
	if err != nil {
		return fmt.Errorf("persist safety trip: %w", err)
	}
	if err := c.auditor.Record(c.unitID, audit.KindSafety, "safety-trip", map[string]string{"reason": reason}); err != nil {
		return fmt.Errorf("audit safety trip: %w", err)
	}
	return nil
}

// Release clears the safety latch once the rotor conditions recover. After a
// successful release the turbine is allowed to reset again.
func (c *Chain) Release() error {
	c.mu.Lock()
	c.latched = false
	c.recovered = true
	err := c.persistLocked()
	c.mu.Unlock()
	if err != nil {
		return fmt.Errorf("persist safety release: %w", err)
	}
	if err := c.auditor.Record(c.unitID, audit.KindSafety, "safety-release", nil); err != nil {
		return fmt.Errorf("audit safety release: %w", err)
	}
	return nil
}

// Restore loads a previously persisted chain state. A missing snapshot yields
// a default released chain.
func Restore(unitID string, st *store.Store, a *audit.Auditor) (*Chain, error) {
	chain := NewChain(unitID, st, a)
	var state struct {
		Latched    bool      `json:"latched"`
		Recovered  bool      `json:"recovered"`
		TripCount  int       `json:"trip_count"`
		ResetCount int       `json:"reset_count"`
		UpdatedAt  time.Time `json:"updated_at"`
	}
	if err := st.ReadJSON("safety/"+unitID+".json", &state); err != nil {
		return chain, err
	}
	chain.mu.Lock()
	chain.latched = state.Latched
	chain.recovered = state.Recovered
	chain.tripCount = state.TripCount
	chain.resetCount = state.ResetCount
	chain.lastTrip = state.UpdatedAt
	chain.mu.Unlock()
	return chain, nil
}

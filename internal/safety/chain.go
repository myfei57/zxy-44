package safety

import (
	"sync"
	"time"

	"windctl/internal/audit"
	"windctl/internal/store"
)

// Chain is the safety chain state machine of one turbine. It moves between
// released and latched as rotor conditions recover or degrade.
type Chain struct {
	unitID     string
	mu         sync.Mutex
	latched    bool
	recovered  bool
	tripCount  int
	resetCount int
	lastTrip   time.Time
	store      *store.Store
	auditor    *audit.Auditor
}

// NewChain creates a released safety chain for a unit.
func NewChain(unitID string, st *store.Store, a *audit.Auditor) *Chain {
	return &Chain{
		unitID:   unitID,
		store:    st,
		auditor:  a,
		recovered: true,
	}
}

// Latched reports whether the chain is currently tripped.
func (c *Chain) Latched() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.latched
}

// Recovered reports whether the rotor conditions recovered after a trip.
func (c *Chain) Recovered() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.recovered
}

// Released reports whether the chain is both unlatched and recovered.
func (c *Chain) Released() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.latched && c.recovered
}

// TripCount returns how many trips happened since start.
func (c *Chain) TripCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tripCount
}

// ResetCount returns how many resets happened since start.
func (c *Chain) ResetCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.resetCount
}

// LastTrip returns when the chain last tripped.
func (c *Chain) LastTrip() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastTrip
}

// PersistState stores the latch state for crash recovery.
func (c *Chain) PersistState() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.persistLocked()
}

func (c *Chain) persistLocked() error {
	return c.store.WriteJSON("safety/"+c.unitID+".json", map[string]any{
		"latched":     c.latched,
		"recovered":   c.recovered,
		"trip_count":  c.tripCount,
		"reset_count": c.resetCount,
		"updated_at":  time.Now().UTC(),
	})
}

// UnitID returns the turbine unit of the chain.
func (c *Chain) UnitID() string {
	return c.unitID
}

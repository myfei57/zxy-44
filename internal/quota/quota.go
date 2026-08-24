package quota

import (
	"fmt"
	"sync"
	"time"

	"windctl/internal/store"
)

// Quota enforces the sampling budget per turbine unit and persists the ledger
// so a restart does not silently reset an exhausted cycle.
type Quota struct {
	mu        sync.Mutex
	policy    Policy
	remaining map[string]int
	cycle     int64
	store     *store.Store
}

// New creates a quota manager with a persisted ledger.
func New(policy Policy, st *store.Store) (*Quota, error) {
	if !policy.Valid() {
		return nil, fmt.Errorf("invalid quota policy %+v", policy)
	}
	q := &Quota{
		policy:    policy,
		remaining: make(map[string]int),
		cycle:     time.Now().UTC().Unix() / int64(policy.CycleSeconds),
		store:     st,
	}
	var saved Ledger
	if loadErr := st.ReadJSON("quota/ledger.json", &saved); loadErr == nil && saved.Cycle == q.cycle {
		q.remaining = saved.Remaining
	}
	return q, nil
}

// Acquire consumes n sampling units for a turbine when the budget allows it.
func (q *Quota) Acquire(unitID string, n int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	remaining, ok := q.remaining[unitID]
	if !ok {
		remaining = q.policy.MaxPerCycle
		q.remaining[unitID] = remaining
	}
	if remaining < n {
		return false
	}
	q.remaining[unitID] = remaining - n
	return true
}

// Release returns n sampling units to the turbine budget.
func (q *Quota) Release(unitID string, n int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.remaining[unitID] += n
}

// Remaining returns the budget still available for a turbine.
func (q *Quota) Remaining(unitID string) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.remaining[unitID]
}

// ResetCycle starts a fresh sampling cycle and persists the ledger.
func (q *Quota) ResetCycle() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	cycle := time.Now().UTC().Unix() / int64(q.policy.CycleSeconds)
	if cycle == q.cycle {
		return nil
	}
	q.cycle = cycle
	q.remaining = make(map[string]int)
	return q.persistLocked()
}

// Persist writes the current ledger to the store.
func (q *Quota) Persist() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.persistLocked()
}

func (q *Quota) persistLocked() error {
	ledger := NewLedger(q.cycle)
	ledger.Remaining = q.remaining
	return q.store.WriteJSON("quota/ledger.json", ledger)
}

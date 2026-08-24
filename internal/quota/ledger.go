package quota

import "time"

// Ledger is the persisted remaining-budget snapshot of one sampling cycle.
type Ledger struct {
	Remaining map[string]int `json:"remaining"`
	Cycle     int64          `json:"cycle"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// NewLedger initialises an empty ledger for a cycle.
func NewLedger(cycle int64) *Ledger {
	return &Ledger{
		Remaining: make(map[string]int),
		Cycle:     cycle,
		UpdatedAt: time.Now().UTC(),
	}
}

package audit

import (
	"encoding/json"
	"fmt"
	"sync"

	"windctl/internal/store"
)

// Auditor persists control events to the append-only audit log and serves the
// most recent events to the console.
type Auditor struct {
	store *store.Store
	mu    sync.Mutex
}

// New builds an auditor on top of the given store.
func New(st *store.Store) *Auditor {
	return &Auditor{store: st}
}

// Record appends one event to the control audit log.
func (a *Auditor) Record(unitID, kind, message string, meta map[string]string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	entry := NewEntry(unitID, kind, message, meta)
	return a.store.AppendAudit(entry)
}

// Recent returns up to limit events, optionally filtered by kind, in reverse
// chronological order.
func (a *Auditor) Recent(kind string, limit int) ([]Entry, error) {
	lines, err := a.store.ReadAuditLines()
	if err != nil {
		return nil, fmt.Errorf("read audit log: %w", err)
	}
	var all []Entry
	for _, line := range lines {
		var entry Entry
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, fmt.Errorf("parse audit line: %w", err)
		}
		all = append(all, entry)
	}
	if kind != "" {
		all = Filter(all, kind, "")
	}
	if len(all) > limit {
		all = all[len(all)-limit:]
	}
	reversed := make([]Entry, 0, len(all))
	for i := len(all) - 1; i >= 0; i-- {
		reversed = append(reversed, all[i])
	}
	return reversed, nil
}

// Count returns the number of events currently stored.
func (a *Auditor) Count() (int, error) {
	lines, err := a.store.ReadAuditLines()
	if err != nil {
		return 0, err
	}
	return len(lines), nil
}

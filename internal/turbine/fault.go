package turbine

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Fault is one fault report of a turbine.
type Fault struct {
	ID      string    `json:"id"`
	Code    string    `json:"code"`
	Message string    `json:"message"`
	At      time.Time `json:"at"`
	Seq     uint64    `json:"seq"`
}

// FaultLog keeps the ordered fault reports of one turbine. It is shared by the
// safety reset path and the fault reporting path, so every mutation must be
// serialised.
type FaultLog struct {
	mu     sync.Mutex
	faults []Fault
	seq    uint64
}

// NewFaultLog creates an empty fault log.
func NewFaultLog() *FaultLog {
	return &FaultLog{}
}

// Record appends a fault report and returns the stored fault.
func (f *FaultLog) Record(code, message string) Fault {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seq++
	fault := Fault{
		ID:      uuid.NewString(),
		Code:    code,
		Message: message,
		At:      time.Now().UTC(),
		Seq:     f.seq,
	}
	f.faults = append(f.faults, fault)
	return fault
}

// RecordFault marks the turbine as faulted and stores a new fault report.
func (t *Turbine) RecordFault(code, message string) Fault {
	t.mu.Lock()
	t.state = StateFaulted
	t.mu.Unlock()
	return t.faults.Record(code, message)
}

// Clear removes every recorded fault and returns what was cleared.
func (f *FaultLog) Clear() []Fault {
	f.mu.Lock()
	defer f.mu.Unlock()
	cleared := f.faults
	f.faults = nil
	return cleared
}

// Recent returns up to limit of the most recent faults, newest first.
func (f *FaultLog) Recent(limit int) []Fault {
	f.mu.Lock()
	defer f.mu.Unlock()
	if limit <= 0 || limit > len(f.faults) {
		limit = len(f.faults)
	}
	out := make([]Fault, 0, limit)
	for i := len(f.faults) - 1; i >= len(f.faults)-limit; i-- {
		out = append(out, f.faults[i])
	}
	return out
}

// Count returns the number of recorded faults.
func (f *FaultLog) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.faults)
}

// Last returns the most recent fault, if any.
func (f *FaultLog) Last() (Fault, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.faults) == 0 {
		return Fault{}, false
	}
	return f.faults[len(f.faults)-1], true
}

package turbine

import "time"

// Status is the console-facing snapshot of one turbine.
type Status struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	FarmID     string    `json:"farm_id"`
	State      string    `json:"state"`
	Speed      float64   `json:"speed"`
	Pitch      float64   `json:"pitch"`
	Azimuth    float64   `json:"azimuth"`
	FaultCount int       `json:"fault_count"`
	ResetSeq   uint64    `json:"reset_seq"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Snapshot builds a consistent status view of the turbine.
func Snapshot(t *Turbine) Status {
	t.mu.Lock()
	status := Status{
		ID:         t.ID,
		Name:       t.Name,
		FarmID:     t.FarmID,
		State:      t.state.String(),
		Speed:      t.speed,
		Pitch:      t.pitch,
		Azimuth:    t.azimuth,
		FaultCount: t.faults.Count(),
		ResetSeq:   t.resetSeq,
		UpdatedAt:  time.Now().UTC(),
	}
	t.mu.Unlock()
	return status
}

package blade

import "time"

// Feedback is the console-facing pitch state of one turbine.
type Feedback struct {
	UnitID   string    `json:"unit_id"`
	Angle    float64   `json:"angle"`
	Target   float64   `json:"target"`
	MaxStep  float64   `json:"max_step"`
	Sequence uint64    `json:"sequence"`
	Journal  []Command `json:"journal"`
	At       time.Time `json:"at"`
}

// Report builds a pitch feedback snapshot.
func Report(ctl *PitchController) Feedback {
	return Feedback{
		UnitID:   ctl.unitID,
		Angle:    ctl.Current(),
		Target:   ctl.Target(),
		MaxStep:  ctl.MaxStep(),
		Sequence: ctl.Sequence(),
		Journal:  ctl.Journal(),
		At:       time.Now().UTC(),
	}
}

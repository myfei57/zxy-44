package wind

import "time"

// Sample is one wind speed measurement of a turbine unit.
type Sample struct {
	UnitID string    `json:"unit_id"`
	Speed  float64   `json:"speed"`
	At     time.Time `json:"at"`
	Seq    uint64    `json:"seq"`
}

// SourceFunc produces the current wind speed measured by an anemometer.
type SourceFunc func() (float64, error)

package rotor

// OverspeedDetector decides whether the rotor is turning faster than the
// configured limit, feeding the safety chain.
type OverspeedDetector struct {
	rotor *Rotor
	limit float64
}

// NewOverspeedDetector creates a detector around a rotor.
func NewOverspeedDetector(r *Rotor, limit float64) *OverspeedDetector {
	return &OverspeedDetector{rotor: r, limit: limit}
}

// Trip reports whether the rotor exceeds the overspeed limit.
func (d *OverspeedDetector) Trip() bool {
	return d.rotor.Speed() > d.limit
}

// SetLimit changes the overspeed threshold.
func (d *OverspeedDetector) SetLimit(limit float64) {
	d.limit = limit
}

// Limit returns the current overspeed threshold.
func (d *OverspeedDetector) Limit() float64 {
	return d.limit
}

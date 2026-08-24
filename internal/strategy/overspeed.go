package strategy

import "windctl/internal/rotor"

// Evaluator checks the overspeed condition against the rotor. It must always
// read the currently selected speed source.
type Evaluator struct {
	rotor *rotor.Rotor
}

// NewEvaluator creates an overspeed evaluator around a rotor.
func NewEvaluator(r *rotor.Rotor) *Evaluator {
	return &Evaluator{rotor: r}
}

// Check reports whether the rotor speed exceeds limit.
func (e *Evaluator) Check(limit float64) bool {
	return e.rotor.Speed() > limit
}

// LimitForMode returns the overspeed limit appropriate for a strategy mode.
func LimitForMode(mode string) float64 {
	switch mode {
	case ModeLimitPower:
		return 14
	default:
		return 18
	}
}

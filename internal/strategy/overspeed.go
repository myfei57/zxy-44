package strategy

import "windctl/internal/rotor"

// Evaluator checks the overspeed condition against the rotor. It resolves the
// speed source on every check so it follows a source switch immediately
// instead of reading a stale, cached source.
type Evaluator struct {
	rotor *rotor.Rotor
}

// NewEvaluator creates an overspeed evaluator around a rotor.
func NewEvaluator(r *rotor.Rotor) *Evaluator {
	return &Evaluator{rotor: r}
}

// Check reports whether the rotor speed exceeds limit.
func (e *Evaluator) Check(limit float64) bool {
	source := e.rotor.Current()
	if source == nil {
		return false
	}
	value, err := source.Read()
	return err == nil && value > limit
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

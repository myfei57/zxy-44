package quota

// Policy controls how many samples each turbine unit may take per cycle.
type Policy struct {
	MaxPerCycle  int
	CycleSeconds int
}

// DefaultPolicy allows 600 samples per ten minute cycle per unit.
func DefaultPolicy() Policy {
	return Policy{MaxPerCycle: 600, CycleSeconds: 600}
}

// Valid reports whether the policy is usable.
func (p Policy) Valid() bool {
	return p.MaxPerCycle > 0 && p.CycleSeconds > 0
}

package blade

import "fmt"

// Apply moves the blade pitch towards target by at most one step and records
// the transition. Applying a transition is atomic: concurrent callers observe
// either the old angle or the new angle, never an interleaved value.
func (p *PitchController) Apply(target float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.angle = moveToward(p.angle, target, p.maxStep)
	p.target = target
	p.seq++
	if p.angle > 90 || p.angle < -2 {
		return fmt.Errorf("pitch angle %v out of range", p.angle)
	}
	return nil
}

// moveToward changes value towards target by at most step.
func moveToward(value, target, step float64) float64 {
	delta := target - value
	if delta > step {
		return value + step
	}
	if delta < -step {
		return value - step
	}
	return target
}

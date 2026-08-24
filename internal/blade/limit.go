package blade

// PowerAngle maps a maximum power value to a pitch angle. Higher power limits
// keep the blades closer to the feathered position being smaller.
func PowerAngle(maxPower float64) float64 {
	angle := 90 - (maxPower / 100 * 80)
	if angle < 0 {
		return 0
	}
	if angle > 90 {
		return 90
	}
	return angle
}

// LimitPower commands the pitch angle corresponding to a power limit.
func LimitPower(ctl *PitchController, maxPower float64) error {
	return Execute(ctl, []Step{{Command: CmdLimitPower, Target: PowerAngle(maxPower)}})
}

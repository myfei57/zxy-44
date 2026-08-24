package blade

// FeatherAngle is the blade angle at which the blades stop generating lift.
const FeatherAngle = 90.0

// Feather moves the blade to the feathered position.
func Feather(ctl *PitchController) error {
	return Execute(ctl, []Step{{Command: CmdFeather, Target: FeatherAngle}})
}

package blade

// Command is one step of a pitch sequence.
type Command int

const (
	CmdFeather Command = iota
	CmdLimitPower
	CmdTrackWind
	CmdHold
)

// String returns the command name.
func (c Command) String() string {
	switch c {
	case CmdFeather:
		return "feather"
	case CmdLimitPower:
		return "limit-power"
	case CmdTrackWind:
		return "track-wind"
	case CmdHold:
		return "hold"
	default:
		return "unknown"
	}
}

// Step is one ordered pitch command.
type Step struct {
	Command Command
	Target  float64
}

// Execute applies a pitch sequence in the exact order the strategy produced.
// The journal records every command so the console can audit the sequence.
func Execute(ctl *PitchController, steps []Step) error {
	for _, step := range steps {
		ctl.mu.Lock()
		ctl.record(step.Command)
		ctl.mu.Unlock()
		if err := ctl.Apply(step.Target); err != nil {
			return err
		}
	}
	return nil
}

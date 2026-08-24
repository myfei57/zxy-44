package strategy

import (
	"windctl/internal/blade"
)

// OverspeedWindSpeed is the wind speed at which the strategy treats the rotor
// as overspeed and feathers the blades.
const OverspeedWindSpeed = 25.0

// Decide produces the ordered pitch commands for the current conditions. In an
// overspeed sequence the blades must feather first and only then apply the
// limit-power adjustment, otherwise the rotor keeps accelerating.
func Decide(c *Controller, windSpeed float64, overspeed bool) ([]blade.Step, error) {
	if overspeed {
		return []blade.Step{
			{Command: blade.CmdFeather, Target: blade.FeatherAngle},
			{Command: blade.CmdLimitPower, Target: AngleForPower(c.Active(), c.Active().MaxPower)},
		}, nil
	}
	if windSpeed > OverspeedWindSpeed {
		return []blade.Step{
			{Command: blade.CmdLimitPower, Target: AngleForPower(c.Active(), c.Active().MaxPower)},
		}, nil
	}
	return []blade.Step{
		{Command: blade.CmdHold, Target: c.Pitch().Current()},
	}, nil
}

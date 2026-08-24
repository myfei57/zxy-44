package strategy

import (
	"fmt"

	"windctl/internal/blade"
)

// Control modes of a strategy configuration.
const (
	ModeMaxCapture = "max-capture"
	ModeLimitPower = "limit-power"
	ModeIdle       = "idle"
)

// Config is one versioned control strategy configuration.
type Config struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  int    `json:"version"`
	Mode     string `json:"mode"`
	MaxPower float64 `json:"max_power"`
	Payload  []byte `json:"payload,omitempty"`
}

// Validate rejects strategy configurations that cannot be activated.
func Validate(cfg Config) error {
	if cfg.Name == "" {
		return fmt.Errorf("strategy name must not be empty")
	}
	switch cfg.Mode {
	case ModeMaxCapture, ModeLimitPower, ModeIdle:
	default:
		return fmt.Errorf("unsupported strategy mode %q", cfg.Mode)
	}
	if cfg.MaxPower < 0 {
		return fmt.Errorf("max power must not be negative")
	}
	return nil
}

// AngleForPower maps a power demand to the pitch angle the blades should hold.
func AngleForPower(cfg Config, power float64) float64 {
	return blade.PowerAngle(power)
}

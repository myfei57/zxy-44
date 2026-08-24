package strategy

import (
	"sync"

	"windctl/internal/audit"
	"windctl/internal/blade"
	"windctl/internal/rotor"
)

// ConfigSink persists strategy configurations durably.
type ConfigSink interface {
	PersistConfig(Config) error
	LoadConfig(string) ([]byte, error)
	HasConfig(string) bool
}

// Controller owns the active strategy, its version history and the dispatcher
// that serialises pitch commands for one turbine.
type Controller struct {
	unitID     string
	mu         sync.Mutex
	active     Config
	versions   []Config
	sink       ConfigSink
	rotor      *rotor.Rotor
	ctl        *blade.PitchController
	auditor    *audit.Auditor
	dispatcher *Dispatcher
}

// NewController creates a controller running a default max-capture strategy.
func NewController(unitID string, sink ConfigSink, r *rotor.Rotor, ctl *blade.PitchController, a *audit.Auditor) *Controller {
	defaultConfig := Config{
		ID:       "default",
		Name:     "default",
		Version:  1,
		Mode:     ModeMaxCapture,
		MaxPower: 100,
	}
	c := &Controller{
		unitID:   unitID,
		active:   defaultConfig,
		versions: []Config{defaultConfig},
		sink:     sink,
		rotor:    r,
		ctl:      ctl,
		auditor:  a,
	}
	c.dispatcher = NewDispatcher(ctl)
	return c
}

// Active returns the currently active strategy.
func (c *Controller) Active() Config {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.active
}

// Versions returns the version history, oldest first.
func (c *Controller) Versions() []Config {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Config, len(c.versions))
	copy(out, c.versions)
	return out
}

// Rotor exposes the rotor used by the strategy.
func (c *Controller) Rotor() *rotor.Rotor {
	return c.rotor
}

// Pitch exposes the blade controller used by the strategy.
func (c *Controller) Pitch() *blade.PitchController {
	return c.ctl
}

// Dispatcher returns the serialised pitch command dispatcher.
func (c *Controller) Dispatcher() *Dispatcher {
	return c.dispatcher
}

// Sink returns the durable config sink.
func (c *Controller) Sink() ConfigSink {
	return c.sink
}

// UnitID returns the turbine unit of the controller.
func (c *Controller) UnitID() string {
	return c.unitID
}

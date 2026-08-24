package strategy

import (
	"sync"

	"windctl/internal/blade"
)

// Loop is one periodic control loop of a turbine. The power loop and the
// protection loop run concurrently; both must dispatch through the shared
// dispatcher so pitch commands to the same blade apply serially.
type Loop struct {
	name   string
	c      *Controller
	decide func() ([]blade.Step, error)
}

// NewLoop creates a named control loop.
func NewLoop(name string, c *Controller, decide func() ([]blade.Step, error)) *Loop {
	return &Loop{name: name, c: c, decide: decide}
}

// Tick runs one decision and dispatches the resulting pitch sequence.
func (l *Loop) Tick() error {
	steps, err := l.decide()
	if err != nil {
		return err
	}
	return l.c.Dispatcher().Dispatch(steps)
}

// Name returns the loop name.
func (l *Loop) Name() string {
	return l.name
}

// Dispatcher serialises pitch sequences from concurrent control loops.
type Dispatcher struct {
	mu  sync.Mutex
	ctl *blade.PitchController
}

// NewDispatcher creates a dispatcher for one blade controller.
func NewDispatcher(ctl *blade.PitchController) *Dispatcher {
	return &Dispatcher{ctl: ctl}
}

// Dispatch applies the steps under the dispatcher lock so concurrent loops
// never interleave on the same blade state.
func (d *Dispatcher) Dispatch(steps []blade.Step) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return blade.Execute(d.ctl, steps)
}

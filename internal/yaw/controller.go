package yaw

import (
	"errors"
	"sync"

	"windctl/internal/audit"
	"windctl/internal/cable"
	"windctl/internal/store"
)

// ErrCableTwist is returned when the yaw controller tries to follow the wind
// while the cable twist limit is already reached.
var ErrCableTwist = errors.New("cable twist limit reached")

// YawController steers the nacelle towards the wind and coordinates with the
// cable twist monitor.
type YawController struct {
	unitID   string
	mu       sync.Mutex
	azimuth  float64
	targetAz float64
	moving   bool
	cable    *cable.Monitor
	auditor  *audit.Auditor
	store    *store.Store
}

// NewYawController creates a yaw controller for a unit.
func NewYawController(unitID string, initial float64, m *cable.Monitor, a *audit.Auditor, st *store.Store) *YawController {
	return &YawController{
		unitID:   unitID,
		azimuth:  initial,
		targetAz: initial,
		cable:    m,
		auditor:  a,
		store:    st,
	}
}

// Azimuth returns the current nacelle azimuth.
func (y *YawController) Azimuth() float64 {
	y.mu.Lock()
	defer y.mu.Unlock()
	return y.azimuth
}

// Target returns the azimuth the nacelle is heading to.
func (y *YawController) Target() float64 {
	y.mu.Lock()
	defer y.mu.Unlock()
	return y.targetAz
}

// Moving reports whether the nacelle is currently rotating.
func (y *YawController) Moving() bool {
	y.mu.Lock()
	defer y.mu.Unlock()
	return y.moving
}

// Cable exposes the cable twist monitor of the unit.
func (y *YawController) Cable() *cable.Monitor {
	return y.cable
}

// UnitID returns the turbine unit of the controller.
func (y *YawController) UnitID() string {
	return y.unitID
}

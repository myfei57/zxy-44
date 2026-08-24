package yaw

import "windctl/internal/strategy"

// SetAzimuth stores the nacelle azimuth. The value is wrapped into the
// [0, 360) ring so later direction computations stay correct after the
// nacelle passes the wrap point.
func (y *YawController) SetAzimuth(azimuth float64) {
	y.mu.Lock()
	defer y.mu.Unlock()
	y.azimuth = azimuth
}

// SetTarget records the azimuth the nacelle should head to.
func (y *YawController) SetTarget(target float64) {
	y.mu.Lock()
	defer y.mu.Unlock()
	y.targetAz = strategy.NormalizeAzimuth(target)
}

// SetMoving records whether the nacelle is rotating.
func (y *YawController) SetMoving(moving bool) {
	y.mu.Lock()
	defer y.mu.Unlock()
	y.moving = moving
}

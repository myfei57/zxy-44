package strategy

import "math"

// Direction returns the signed shortest rotation from current to target
// azimuth. The azimuth is a ring: crossing 360 degrees must not invert the
// direction.
func Direction(current, target float64) float64 {
	delta := math.Mod(target-current+540, 360) - 180
	return delta
}

// NormalizeAzimuth wraps any angle into the [0, 360) ring.
func NormalizeAzimuth(degrees float64) float64 {
	normalized := math.Mod(degrees, 360)
	if normalized < 0 {
		normalized += 360
	}
	return normalized
}

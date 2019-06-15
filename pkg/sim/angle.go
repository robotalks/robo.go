package sim

import "math"

// AngleFromDegrees creates Angle from degrees.
func AngleFromDegrees(d float64) Angle {
	return Angle(normalizeRadians(d * math.Pi / 180.0))
}

// AngleFromRadians creates Angle from radians.
func AngleFromRadians(r float64) Angle {
	return Angle(normalizeRadians(r))
}

// Add adds an Angle.
func (a Angle) Add(a1 Angle) Angle {
	return Angle(normalizeRadians(float64(a) + float64(a1)))
}

// AddRadians adds radians to current angle.
func (a Angle) AddRadians(r float64) Angle {
	return Angle(normalizeRadians(float64(a) + r))
}

// AddDegrees adds degress to current angle.
func (a Angle) AddDegrees(d float64) Angle {
	return a.AddRadians(d * math.Pi / 180.0)
}

// Radians gets angle in radians.
func (a Angle) Radians() float64 {
	return float64(a)
}

// Degrees gets angle in degrees.
func (a Angle) Degrees() float64 {
	return float64(a) * 180 / math.Pi
}

// Cos wraps math.Cos.
func (a Angle) Cos() float64 {
	return math.Cos(float64(a))
}

// Sin wraps math.Sin.
func (a Angle) Sin() float64 {
	return math.Sin(float64(a))
}

// Tan wraps math.Tan.
func (a Angle) Tan() float64 {
	return math.Tan(float64(a))
}

// Project projects distance into X and Y.
func (a Angle) Project(dist float64) Pos2D {
	return Pos2D{X: dist * a.Cos(), Y: dist * a.Sin()}
}

func normalizeRadians(r float64) float64 {
	if r >= 2*math.Pi || r <= -2*math.Pi {
		r = math.Remainder(r, 2*math.Pi)
	}
	if r > math.Pi {
		r -= 2 * math.Pi
	} else if r < -math.Pi {
		r += 2 * math.Pi
	}
	return r
}

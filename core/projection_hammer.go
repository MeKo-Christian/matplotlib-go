package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

// hammerDataTransform implements the Hammer-Aitoff equal-area projection.
//
// Reference: third_party/matplotlib/lib/matplotlib/projections/geo.py:301.
//
// Forward:
//
//	α = sqrt(1 + cos(lat) cos(lon/2))
//	x = 2√2 cos(lat) sin(lon/2) / α
//	y =  √2 sin(lat) / α
//
// Output is normalised to [0,1] using the same x: 4√2, y: 2√2 spans as
// mollweideDataTransform so all geo projections share the elliptical frame.
type hammerDataTransform struct{}

func (hammerDataTransform) Apply(p geom.Pt) geom.Pt {
	lon := clamp(p.X, -math.Pi, math.Pi)
	lat := clamp(p.Y, -math.Pi/2, math.Pi/2)
	half := lon / 2
	alpha := math.Sqrt(1 + math.Cos(lat)*math.Cos(half))
	if alpha < 1e-12 {
		alpha = 1e-12
	}
	x := 2 * math.Sqrt2 * math.Cos(lat) * math.Sin(half) / alpha
	y := math.Sqrt2 * math.Sin(lat) / alpha
	return geom.Pt{
		X: 0.5 + x/(4*math.Sqrt2),
		Y: 0.5 + y/(2*math.Sqrt2),
	}
}

func (hammerDataTransform) Invert(p geom.Pt) (geom.Pt, bool) {
	x := (p.X - 0.5) * 4 * math.Sqrt2
	y := (p.Y - 0.5) * 2 * math.Sqrt2
	zsq := 1 - (x/4)*(x/4) - (y/2)*(y/2)
	if zsq < 0 {
		return geom.Pt{}, false
	}
	z := math.Sqrt(zsq)
	denom := 2*z*z - 1
	if math.Abs(denom) < 1e-12 {
		return geom.Pt{}, false
	}
	lon := 2 * math.Atan2(z*x/2, denom)
	lat := math.Asin(clamp(y*z, -1, 1))
	return geom.Pt{X: lon, Y: lat}, true
}

func newHammerProjection() *geoProjection {
	return &geoProjection{name: "hammer", transform: hammerDataTransform{}}
}

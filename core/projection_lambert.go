package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

// lambertDataTransform implements the Lambert azimuthal equal-area projection
// centered on (centerLon, centerLat).
//
// Reference: third_party/matplotlib/lib/matplotlib/projections/geo.py:416.
//
// Forward:
//
//	k = sqrt(2 / (1 + sin(c_lat) sin(lat) + cos(c_lat) cos(lat) cos(Δlon)))
//	x = k cos(lat) sin(Δlon)
//	y = k (cos(c_lat) sin(lat) − sin(c_lat) cos(lat) cos(Δlon))
//
// Natural extents are x,y ∈ [−2, 2]; we normalise to [0, 1] symmetrically so
// the shared geo elliptical frame still encloses the projected disc.
type lambertDataTransform struct {
	centerLon, centerLat float64
}

func (t lambertDataTransform) Apply(p geom.Pt) geom.Pt {
	lon := clamp(p.X, -math.Pi, math.Pi)
	lat := clamp(p.Y, -math.Pi/2, math.Pi/2)
	dlon := lon - t.centerLon
	cl := math.Cos(t.centerLat)
	sl := math.Sin(t.centerLat)
	cosLat := math.Cos(lat)
	sinLat := math.Sin(lat)
	cosDLon := math.Cos(dlon)
	denom := 1 + sl*sinLat + cl*cosLat*cosDLon
	if denom < 1e-15 {
		denom = 1e-15
	}
	k := math.Sqrt(2 / denom)
	x := k * cosLat * math.Sin(dlon)
	y := k * (cl*sinLat - sl*cosLat*cosDLon)
	return geom.Pt{
		X: 0.5 + x/4,
		Y: 0.5 + y/4,
	}
}

func (t lambertDataTransform) Invert(p geom.Pt) (geom.Pt, bool) {
	x := (p.X - 0.5) * 4
	y := (p.Y - 0.5) * 4
	rho := math.Hypot(x, y)
	if rho < 1e-15 {
		return geom.Pt{X: t.centerLon, Y: t.centerLat}, true
	}
	if rho > 2 {
		return geom.Pt{}, false
	}
	c := 2 * math.Asin(rho/2)
	sinC := math.Sin(c)
	cosC := math.Cos(c)
	cl := math.Cos(t.centerLat)
	sl := math.Sin(t.centerLat)
	lat := math.Asin(clamp(cosC*sl+(y*sinC*cl)/rho, -1, 1))
	lon := t.centerLon + math.Atan2(x*sinC, rho*cl*cosC-y*sl*sinC)
	return geom.Pt{X: lon, Y: lat}, true
}

func newLambertProjection() *geoProjection {
	return &geoProjection{name: "lambert", transform: lambertDataTransform{}}
}

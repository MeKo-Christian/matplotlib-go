package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

// aitoffDataTransform implements the Aitoff projection.
//
// Reference: third_party/matplotlib/lib/matplotlib/projections/geo.py:255.
//
// Forward:
//
//	α = arccos(cos(lat) cos(lon/2))
//	sinc(α) = sin(α) / α   (1 at α=0)
//	x = cos(lat) sin(lon/2) / sinc(α)
//	y = sin(lat) / sinc(α)
//
// Inverse is intentionally unsupported, matching matplotlib's AitoffAxes
// (the upstream implementation also returns no inverse).
type aitoffDataTransform struct{}

func (aitoffDataTransform) Apply(p geom.Pt) geom.Pt {
	lon := clamp(p.X, -math.Pi, math.Pi)
	lat := clamp(p.Y, -math.Pi/2, math.Pi/2)
	half := lon / 2
	alpha := math.Acos(clamp(math.Cos(lat)*math.Cos(half), -1, 1))
	sinc := 1.0
	if math.Abs(alpha) > 1e-12 {
		sinc = math.Sin(alpha) / alpha
	}
	x := math.Cos(lat) * math.Sin(half) / sinc
	y := math.Sin(lat) / sinc
	// Natural extents: x ∈ [−π/2, π/2] (equator endpoints) and
	// y ∈ [−π/2, π/2] (poles). Both axes share the same scale so that
	// the projection fills the axes box edge-to-edge, matching the
	// matplotlib_go convention used for Mollweide and Hammer.
	return geom.Pt{
		X: 0.5 + x/math.Pi,
		Y: 0.5 + y/math.Pi,
	}
}

func (aitoffDataTransform) Invert(geom.Pt) (geom.Pt, bool) {
	return geom.Pt{}, false
}

func newAitoffProjection() *geoProjection {
	return &geoProjection{name: "aitoff", transform: aitoffDataTransform{}}
}

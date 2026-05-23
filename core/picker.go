package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// PickInfo carries optional extra detail about a successful pick — for example
// the index of the nearest vertex in a polyline. Backends may attach this
// payload to the resulting PickEvent.
type PickInfo struct {
	// Index is the data-space index of the closest vertex or segment when the
	// artist has an obvious enumeration (lines, scatter). Zero when not
	// applicable.
	Index int
	// Distance is the pixel distance from the cursor to the picked geometry.
	Distance float64
	// Extra carries artist-specific payload.
	Extra any
}

// ArtistPicker is an optional Artist extension that reports whether a
// figure-pixel point hits the artist. The default pick radius is 5 pixels,
// matching Matplotlib's Artist.pickradius default.
//
// Coordinates are figure-pixel; implementations transform their data using the
// provided DrawContext. A nil context means the artist should ignore the test.
type ArtistPicker interface {
	Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo)
}

// DefaultPickRadius is the cursor tolerance, in pixels, used when an artist
// does not specify its own pick radius. It mirrors Matplotlib's
// Artist.pickradius default of 5 points / pixels.
const DefaultPickRadius = 5.0

// PickRadiusProvider lets an artist override the default cursor tolerance.
type PickRadiusProvider interface {
	PickRadius() float64
}

// distancePointToSegment returns the perpendicular distance from p to the
// segment [a,b].
func distancePointToSegment(a, b, p geom.Pt) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if dx == 0 && dy == 0 {
		return math.Hypot(p.X-a.X, p.Y-a.Y)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	proj := geom.Pt{X: a.X + t*dx, Y: a.Y + t*dy}
	return math.Hypot(p.X-proj.X, p.Y-proj.Y)
}

// ResolvePickRadius reads PickRadius from the artist or falls back to the
// default tolerance.
func ResolvePickRadius(art any) float64 {
	if pr, ok := art.(PickRadiusProvider); ok {
		if r := pr.PickRadius(); r > 0 {
			return r
		}
	}
	return DefaultPickRadius
}

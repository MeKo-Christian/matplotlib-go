package core

import "github.com/cwbudde/matplotlib-go/geom"

// Public aliases for the most common geometry types used by the core plotting
// API. These are exact type aliases, so values are interchangeable with
// geom.Rect / geom.Pt at no runtime cost.
type (
	// Pt is a 2D point matching geom.Pt.
	Pt = geom.Pt
	// Rect is an axis-aligned rectangle matching geom.Rect.
	Rect = geom.Rect
)

// NewRect builds a Rect from (left, bottom, width, height) in the same
// convention matplotlib uses for figure-relative axes rectangles.
func NewRect(left, bottom, width, height float64) Rect {
	return Rect{
		Min: Pt{X: left, Y: bottom},
		Max: Pt{X: left + width, Y: bottom + height},
	}
}

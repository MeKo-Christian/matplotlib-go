package core

import (
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// MarkerType defines the shape of markers in scatter plots.
type MarkerType uint8

const (
	MarkerCircle MarkerType = iota
	MarkerSquare
	MarkerTriangle
	MarkerDiamond
	MarkerPlus
	MarkerCross
)

// Scatter2D renders points with configurable markers.
type Scatter2D struct {
	XY     []geom.Pt      // data space points
	Sizes  []float64      // marker sizes (radius in pixels), if nil uses Size
	Colors []render.Color // marker colors, if nil uses Color
	Size   float64        // default marker size (radius in pixels)
	Color  render.Color   // default marker color
	Marker MarkerType     // marker shape
	z      float64        // z-order
}

// Draw renders scatter points by creating filled paths for each marker.
func (s *Scatter2D) Draw(r render.Renderer, ctx *DrawContext) {
	if len(s.XY) == 0 {
		return // nothing to draw
	}

	for i, pt := range s.XY {
		// Transform to pixel coordinates
		pixelPt := ctx.DataToPixel.Apply(pt)

		// Get size for this point
		size := s.Size
		if s.Sizes != nil && i < len(s.Sizes) {
			size = s.Sizes[i]
		}

		// Get color for this point
		color := s.Color
		if s.Colors != nil && i < len(s.Colors) {
			color = s.Colors[i]
		}

		// Create marker path
		markerPath := s.createMarkerPath(pixelPt, size)
		if len(markerPath.C) == 0 {
			continue // skip invalid markers
		}

		// Draw filled marker
		paint := render.Paint{
			Fill: color,
		}
		r.Path(markerPath, &paint)
	}
}

// createMarkerPath creates a filled path for the given marker type at the specified position and size.
func (s *Scatter2D) createMarkerPath(center geom.Pt, radius float64) geom.Path {
	switch s.Marker {
	case MarkerCircle:
		return s.createCirclePath(center, radius)
	case MarkerSquare:
		return s.createSquarePath(center, radius)
	case MarkerTriangle:
		return s.createTrianglePath(center, radius)
	case MarkerDiamond:
		return s.createDiamondPath(center, radius)
	case MarkerPlus:
		return s.createPlusPath(center, radius)
	case MarkerCross:
		return s.createCrossPath(center, radius)
	default:
		return s.createCirclePath(center, radius) // default to circle
	}
}

// createCirclePath creates a circular marker using a polygon approximation.
func (s *Scatter2D) createCirclePath(center geom.Pt, radius float64) geom.Path {
	const numSegments = 16 // Good balance of smoothness and performance
	path := geom.Path{}

	for i := 0; i < numSegments; i++ {
		angle := 2 * math.Pi * float64(i) / numSegments
		x := center.X + radius*math.Cos(angle)
		y := center.Y + radius*math.Sin(angle)

		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, geom.Pt{X: x, Y: y})
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createSquarePath creates a square marker centered at the given point.
func (s *Scatter2D) createSquarePath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Square vertices
	vertices := []geom.Pt{
		{X: center.X - radius, Y: center.Y - radius}, // bottom-left
		{X: center.X + radius, Y: center.Y - radius}, // bottom-right
		{X: center.X + radius, Y: center.Y + radius}, // top-right
		{X: center.X - radius, Y: center.Y + radius}, // top-left
	}

	for i, v := range vertices {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createTrianglePath creates an upward-pointing triangle marker.
func (s *Scatter2D) createTrianglePath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Triangle vertices (equilateral triangle pointing up)
	height := radius * math.Sqrt(3) / 2
	vertices := []geom.Pt{
		{X: center.X, Y: center.Y + height},            // top
		{X: center.X - radius, Y: center.Y - height/2}, // bottom-left
		{X: center.X + radius, Y: center.Y - height/2}, // bottom-right
	}

	for i, v := range vertices {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createDiamondPath creates a diamond (rotated square) marker.
func (s *Scatter2D) createDiamondPath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Diamond vertices
	vertices := []geom.Pt{
		{X: center.X, Y: center.Y + radius}, // top
		{X: center.X + radius, Y: center.Y}, // right
		{X: center.X, Y: center.Y - radius}, // bottom
		{X: center.X - radius, Y: center.Y}, // left
	}

	for i, v := range vertices {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createPlusPath creates a plus sign marker.
func (s *Scatter2D) createPlusPath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Plus is made of two rectangles: horizontal and vertical
	thickness := radius * 0.3 // thickness of the plus arms

	// Horizontal bar
	hBar := []geom.Pt{
		{X: center.X - radius, Y: center.Y - thickness},
		{X: center.X + radius, Y: center.Y - thickness},
		{X: center.X + radius, Y: center.Y + thickness},
		{X: center.X - radius, Y: center.Y + thickness},
	}

	for i, v := range hBar {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	// Vertical bar
	vBar := []geom.Pt{
		{X: center.X - thickness, Y: center.Y - radius},
		{X: center.X + thickness, Y: center.Y - radius},
		{X: center.X + thickness, Y: center.Y + radius},
		{X: center.X - thickness, Y: center.Y + radius},
	}

	for i, v := range vBar {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createCrossPath creates a cross (X) marker.
func (s *Scatter2D) createCrossPath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Cross is made of two diagonal rectangles
	thickness := radius * 0.3
	offset := thickness / math.Sqrt(2) // offset for rotated rectangle

	// First diagonal bar (\)
	diag1 := []geom.Pt{
		{X: center.X - radius + offset, Y: center.Y - radius - offset},
		{X: center.X - radius - offset, Y: center.Y - radius + offset},
		{X: center.X + radius - offset, Y: center.Y + radius - offset},
		{X: center.X + radius + offset, Y: center.Y + radius + offset},
	}

	for i, v := range diag1 {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	// Second diagonal bar (/)
	diag2 := []geom.Pt{
		{X: center.X + radius - offset, Y: center.Y - radius - offset},
		{X: center.X + radius + offset, Y: center.Y - radius + offset},
		{X: center.X - radius + offset, Y: center.Y + radius - offset},
		{X: center.X - radius - offset, Y: center.Y + radius + offset},
	}

	for i, v := range diag2 {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// Z returns the z-order for sorting.
func (s *Scatter2D) Z() float64 {
	return s.z
}

// Bounds returns an empty rect for now (will be enhanced in later phases).
func (s *Scatter2D) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}

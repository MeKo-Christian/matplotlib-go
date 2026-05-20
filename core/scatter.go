package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
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
	ArtistRasterization
	XY          []geom.Pt      // data space points
	Sizes       []float64      // marker areas in points^2, if nil uses Size
	Colors      []render.Color // marker colors, if nil uses Color
	EdgeColors  []render.Color // edge colors for marker outlines, if nil uses EdgeColor
	MarkerPath  geom.Path      // optional custom marker path in normalized collection space
	Size        float64        // default marker area in points^2
	Color       render.Color   // default marker color
	EdgeColor   render.Color   // default edge color for marker outlines
	EdgeWidth   float64        // edge width in pixels (0 means no edge)
	Alpha       float64        // alpha transparency (0-1), applied to both fill and edge
	PathEffects []render.PathEffect
	Marker      MarkerType // marker shape
	Label       string     // series label for legend
	z           float64    // z-order
}

var stemMarkerScale = math.Sqrt(math.Pi)

// Draw renders scatter points by creating filled paths for each marker.
func (s *Scatter2D) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || len(s.XY) == 0 {
		return
	}
	s.toPathCollection(ctx).Draw(r, ctx)
}

// createMarkerPath creates a filled path for the given marker type at the specified position and size.
func (s *Scatter2D) createMarkerPath(center geom.Pt, radius float64) geom.Path {
	if radius <= 0 {
		return geom.Path{}
	}
	return scaleAndTranslatePath(s.markerPrototypePath(), radius*stemMarkerScale, center)
}

func (s *Scatter2D) markerPrototypePath() geom.Path {
	if len(s.MarkerPath.C) > 0 {
		return s.MarkerPath
	}

	switch s.Marker {
	case MarkerCircle:
		return s.createCirclePath(geom.Pt{}, 1)
	case MarkerSquare:
		return s.createSquarePath(geom.Pt{}, 1)
	case MarkerTriangle:
		return s.createTrianglePath(geom.Pt{}, 1)
	case MarkerDiamond:
		return s.createDiamondPath(geom.Pt{}, 1)
	case MarkerPlus:
		return s.createPlusPath(geom.Pt{}, 1)
	case MarkerCross:
		return s.createCrossPath(geom.Pt{}, 1)
	default:
		return s.createCirclePath(geom.Pt{}, 1)
	}
}

func (s *Scatter2D) toPathCollection(ctx *DrawContext) *PathCollection {
	alpha := s.Alpha
	if alpha <= 0 {
		alpha = 1
	}

	lineOnly := s.Marker == MarkerPlus || s.Marker == MarkerCross
	lineWidth := s.EdgeWidth
	if lineOnly && lineWidth <= 0 && ctx != nil && ctx.RC.LineWidth > 0 {
		lineWidth = ctx.RC.LineWidth
	}

	size := scatterAreaScale(s.Size, ctx)
	var sizes []float64
	if len(s.Sizes) > 0 {
		sizes = make([]float64, len(s.XY))
		for i := range s.XY {
			itemSize := s.Size
			if i < len(s.Sizes) {
				itemSize = s.Sizes[i]
			}
			sizes[i] = scatterAreaScale(itemSize, ctx)
		}
	}

	faceColor := s.Color
	edgeColor := s.EdgeColor
	if lineOnly && edgeColor.A <= 0 {
		edgeColor = faceColor
	}

	return &PathCollection{
		Collection: Collection{
			Label:       s.Label,
			Alpha:       alpha,
			z:           s.z,
			PathEffects: cloneRenderPathEffects(s.PathEffects),
		},
		Path:          s.markerPrototypePath(),
		Offsets:       append([]geom.Pt(nil), s.XY...),
		Size:          size,
		Sizes:         sizes,
		PathInDisplay: true,
		FaceColors:    append([]render.Color(nil), s.Colors...),
		FaceColor:     faceColor,
		EdgeColors:    append([]render.Color(nil), s.EdgeColors...),
		EdgeColor:     edgeColor,
		EdgeWidth:     lineWidth,
		LineJoin:      s.markerLineJoin(),
		LineJoinSet:   true,
		LineCap:       s.markerLineCap(),
		LineCapSet:    true,
		LineOnly:      lineOnly,
	}
}

func (s *Scatter2D) markerLineJoin() render.LineJoin {
	switch s.Marker {
	case MarkerSquare, MarkerTriangle, MarkerDiamond:
		return render.JoinMiter
	default:
		return render.JoinRound
	}
}

func (s *Scatter2D) markerLineCap() render.LineCap {
	return render.CapButt
}

func scatterAreaScale(area float64, ctx *DrawContext) float64 {
	if area <= 0 || math.IsNaN(area) || math.IsInf(area, 0) {
		return 0
	}
	dpi := 72.0
	if ctx != nil && ctx.RC.DPI > 0 {
		dpi = ctx.RC.DPI
	}
	return math.Sqrt(area) * dpi / 72.0
}

// ScatterAreaFromRadius converts a marker radius in pixels to Matplotlib's
// scatter area unit, points^2, for the given DPI.
func ScatterAreaFromRadius(radiusPx, dpi float64) float64 {
	if radiusPx <= 0 || dpi <= 0 || math.IsNaN(radiusPx) || math.IsNaN(dpi) || math.IsInf(radiusPx, 0) || math.IsInf(dpi, 0) {
		return 0
	}
	radiusPt := radiusPx * 72.0 / dpi
	return math.Pi * radiusPt * radiusPt
}

// createCirclePath creates the same cubic unit-circle marker Matplotlib uses.
func (s *Scatter2D) createCirclePath(center geom.Pt, radius float64) geom.Path {
	const segments = 8
	const control = 0.2652031
	r := radius * 0.5
	delta := 2 * math.Pi / segments

	point := func(theta float64) geom.Pt {
		return geom.Pt{
			X: center.X + r*math.Cos(theta),
			Y: center.Y + r*math.Sin(theta),
		}
	}
	tangent := func(theta float64) geom.Pt {
		return geom.Pt{X: -math.Sin(theta), Y: math.Cos(theta)}
	}

	path := geom.Path{}
	theta0 := -math.Pi / 2
	path.MoveTo(point(theta0))
	for i := 0; i < segments; i++ {
		theta1 := theta0 + delta
		p0 := point(theta0)
		p1 := point(theta1)
		t0 := tangent(theta0)
		t1 := tangent(theta1)
		path.CubicTo(
			geom.Pt{X: p0.X + control*r*t0.X, Y: p0.Y + control*r*t0.Y},
			geom.Pt{X: p1.X - control*r*t1.X, Y: p1.Y - control*r*t1.Y},
			p1,
		)
		theta0 = theta1
	}
	path.Close()
	return path
}

// createSquarePath creates a square marker centered at the given point.
func (s *Scatter2D) createSquarePath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Square vertices
	half := 0.5 * radius
	vertices := []geom.Pt{
		{X: center.X - half, Y: center.Y - half}, // bottom-left
		{X: center.X + half, Y: center.Y - half}, // bottom-right
		{X: center.X + half, Y: center.Y + half}, // top-right
		{X: center.X - half, Y: center.Y + half}, // top-left
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

	// Triangle vertices matching matplotlib's '^' marker geometry.
	half := 0.5 * radius
	vertices := []geom.Pt{
		{X: center.X, Y: center.Y - half}, // top
		{X: center.X - half, Y: center.Y + half},
		{X: center.X + half, Y: center.Y + half},
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

	// Diamond vertices matching matplotlib's 'D' marker geometry.
	half := radius / math.Sqrt2
	vertices := []geom.Pt{
		{X: center.X, Y: center.Y - half}, // top
		{X: center.X + half, Y: center.Y}, // right
		{X: center.X, Y: center.Y + half}, // bottom
		{X: center.X - half, Y: center.Y}, // left
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

	// Plus is a line marker in matplotlib.
	half := 0.5 * radius
	hBar := []geom.Pt{
		{X: center.X - half, Y: center.Y},
		{X: center.X + half, Y: center.Y},
	}

	for i, v := range hBar {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	// Vertical bar
	vBar := []geom.Pt{
		{X: center.X, Y: center.Y - half},
		{X: center.X, Y: center.Y + half},
	}

	for i, v := range vBar {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	return path
}

// createCrossPath creates a cross (X) marker.
func (s *Scatter2D) createCrossPath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// First diagonal bar (\)
	diag1 := []geom.Pt{
		{X: center.X - 0.5*radius, Y: center.Y - 0.5*radius},
		{X: center.X + 0.5*radius, Y: center.Y + 0.5*radius},
	}

	for i, v := range diag1 {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	// Second diagonal bar (/)
	diag2 := []geom.Pt{
		{X: center.X - 0.5*radius, Y: center.Y + 0.5*radius},
		{X: center.X + 0.5*radius, Y: center.Y - 0.5*radius},
	}

	for i, v := range diag2 {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	return path
}

// Z returns the z-order for sorting.
func (s *Scatter2D) Z() float64 {
	return zOrDefault(s.z, defaultPatchZ)
}

// Bounds returns the data-space bounding box of all marker centers.
func (s *Scatter2D) Bounds(*DrawContext) geom.Rect {
	if len(s.XY) == 0 {
		return geom.Rect{}
	}

	// Initialize bounds with first point
	bounds := geom.Rect{
		Min: s.XY[0],
		Max: s.XY[0],
	}

	// Expand bounds to include all points
	for _, pt := range s.XY[1:] {
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}

	return bounds
}

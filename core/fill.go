package core

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// FillOrientation controls whether the independent coordinate runs along x or y.
type FillOrientation uint8

const (
	FillVertical FillOrientation = iota
	FillHorizontal
)

// Fill2D creates filled areas between curves or from curves to baselines.
type Fill2D struct {
	ArtistRasterization
	X           []float64       // independent coordinates (x for vertical fills, y for horizontal fills)
	Y1          []float64       // first dependent curve (y for vertical fills, x for horizontal fills)
	Y2          []float64       // second dependent curve, if nil uses Baseline
	Baseline    float64         // baseline value when Y2 is nil
	Orientation FillOrientation // vertical (fill_between) or horizontal (fill_betweenx)
	Color       render.Color    // fill color
	EdgeColor   render.Color    // edge color for outline (0 alpha means no edge)
	EdgeWidth   float64         // edge width in pixels (0 means no edge)
	Alpha       float64         // alpha transparency override (0-1), if 0 uses Color.A
	Label       string          // series label for legend
	z           float64         // z-order
}

// Draw renders the filled area by creating a closed path.
func (f *Fill2D) Draw(r render.Renderer, ctx *DrawContext) {
	if len(f.X) == 0 || len(f.Y1) == 0 {
		return // nothing to draw
	}

	// Use minimum length across all arrays
	n := len(f.X)
	if len(f.Y1) < n {
		n = len(f.Y1)
	}
	if f.Y2 != nil && len(f.Y2) < n {
		n = len(f.Y2)
	}

	if n < 2 {
		return // need at least 2 points for area
	}

	// Create the fill path
	fillPath := f.createFillPath(n, ctx)
	if len(fillPath.C) == 0 {
		return // invalid path
	}

	// Get fill color with alpha
	fillColor := f.Color
	if f.Alpha > 0 && f.Alpha <= 1 {
		fillColor.A = f.Alpha
	}

	// Get edge color with alpha
	edgeColor := f.EdgeColor
	if f.Alpha > 0 && f.Alpha <= 1 {
		edgeColor.A = f.Alpha
	}

	// Create paint for fill area
	paint := render.Paint{
		Fill: fillColor,
	}

	// Add stroke if edge width is specified and edge color has alpha > 0
	if f.EdgeWidth > 0 && edgeColor.A > 0 {
		paint.Stroke = edgeColor
		paint.LineWidth = f.EdgeWidth
		paint.LineJoin = render.JoinMiter
		paint.LineCap = render.CapButt
	}

	// Draw fill area with optional edge
	r.Path(fillPath, &paint)
}

// createFillPath creates a closed path for the fill area.
func (f *Fill2D) createFillPath(n int, ctx *DrawContext) geom.Path {
	path := geom.Path{}

	second := func(i int) float64 {
		if f.Y2 != nil {
			return f.Y2[i]
		}
		return f.Baseline
	}

	// Matplotlib's fill_between PolyCollection starts on the second boundary,
	// walks the first boundary forward, then walks the second boundary backward.
	// The duplicate endpoint vertices are intentional: they match upstream's
	// closed polygon path and affect antialiased edge coverage.
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, f.pixelPoint(0, second(0), ctx))

	for i := 0; i < n; i++ {
		pt := f.pixelPoint(i, f.Y1[i], ctx)
		path.C = append(path.C, geom.LineTo)
		path.V = append(path.V, pt)
	}

	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, f.pixelPoint(n-1, second(n-1), ctx))

	for i := n - 1; i >= 0; i-- {
		pt := f.pixelPoint(i, second(i), ctx)
		path.C = append(path.C, geom.LineTo)
		path.V = append(path.V, pt)
	}

	// Close the path
	path.C = append(path.C, geom.ClosePath)

	return path
}

// Z returns the z-order for sorting.
func (f *Fill2D) Z() float64 {
	return zOrDefault(f.z, defaultPatchZ)
}

// Bounds returns the bounding box of the fill area.
func (f *Fill2D) Bounds(*DrawContext) geom.Rect {
	if len(f.X) == 0 || len(f.Y1) == 0 {
		return geom.Rect{}
	}

	// Use minimum length across all arrays
	n := len(f.X)
	if len(f.Y1) < n {
		n = len(f.Y1)
	}
	if f.Y2 != nil && len(f.Y2) < n {
		n = len(f.Y2)
	}

	if n == 0 {
		return geom.Rect{}
	}

	// Initialize bounds with first point
	first := f.dataPoint(f.X[0], f.Y1[0])
	bounds := geom.Rect{Min: first, Max: first}

	// Expand bounds to include all independent and first-boundary values.
	for i := 0; i < n; i++ {
		bounds = expandRect(bounds, f.dataPoint(f.X[i], f.Y1[i]))
	}

	// Include the second boundary or baseline in bounds.
	if f.Y2 != nil {
		for i := 0; i < n; i++ {
			bounds = expandRect(bounds, f.dataPoint(f.X[i], f.Y2[i]))
		}
	} else {
		for i := 0; i < n; i++ {
			bounds = expandRect(bounds, f.dataPoint(f.X[i], f.Baseline))
		}
	}

	return bounds
}

// FillBetween creates a Fill2D for the area between two curves.
func FillBetween(x, y1, y2 []float64, color render.Color) *Fill2D {
	return &Fill2D{
		X:     x,
		Y1:    y1,
		Y2:    y2,
		Color: color,
	}
}

// FillToBaseline creates a Fill2D for the area from a curve to a baseline.
func FillToBaseline(x, y []float64, baseline float64, color render.Color) *Fill2D {
	return &Fill2D{
		X:        x,
		Y1:       y,
		Y2:       nil,
		Baseline: baseline,
		Color:    color,
	}
}

// FillBetweenX creates a Fill2D for the area between two x-curves across y values.
func FillBetweenX(y, x1, x2 []float64, color render.Color) *Fill2D {
	return &Fill2D{
		X:           y,
		Y1:          x1,
		Y2:          x2,
		Orientation: FillHorizontal,
		Color:       color,
	}
}

// FillToBaselineX creates a Fill2D from an x-curve to a vertical baseline across y values.
func FillToBaselineX(y, x []float64, baseline float64, color render.Color) *Fill2D {
	return &Fill2D{
		X:           y,
		Y1:          x,
		Baseline:    baseline,
		Orientation: FillHorizontal,
		Color:       color,
	}
}

func (f *Fill2D) pixelPoint(index int, dep float64, ctx *DrawContext) geom.Pt {
	return ctx.DataToPixel.Apply(f.dataPoint(f.X[index], dep))
}

func (f *Fill2D) dataPoint(primary, dependent float64) geom.Pt {
	if f.Orientation == FillHorizontal {
		return geom.Pt{X: dependent, Y: primary}
	}
	return geom.Pt{X: primary, Y: dependent}
}

func expandRect(rect geom.Rect, pt geom.Pt) geom.Rect {
	if pt.X < rect.Min.X {
		rect.Min.X = pt.X
	}
	if pt.X > rect.Max.X {
		rect.Max.X = pt.X
	}
	if pt.Y < rect.Min.Y {
		rect.Min.Y = pt.Y
	}
	if pt.Y > rect.Max.Y {
		rect.Max.Y = pt.Y
	}
	return rect
}

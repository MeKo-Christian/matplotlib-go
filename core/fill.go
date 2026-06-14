package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// FillOrientation controls whether the independent coordinate runs along x or y.
type FillOrientation uint8

const (
	FillVertical FillOrientation = iota
	FillHorizontal
)

// FillStep controls step interpolation for fill_between-style polygons.
type FillStep uint8

const (
	FillStepNone FillStep = iota
	FillStepPre
	FillStepPost
	FillStepMid
)

// Fill2D creates filled areas between curves or from curves to baselines.
type Fill2D struct {
	ArtistRasterization
	X           []float64       // independent coordinates (x for vertical fills, y for horizontal fills)
	Y1          []float64       // first dependent curve (y for vertical fills, x for horizontal fills)
	Y2          []float64       // second dependent curve, if nil uses Baseline
	Where       []bool          // per-point mask; regions fill only where adjacent points are true
	Interpolate bool            // extend masked regions to curve intersections
	Step        FillStep        // optional step mode
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
		Snap: render.SnapAuto,
	}

	// Add stroke if edge width is specified and edge color has alpha > 0
	if f.EdgeWidth > 0 && edgeColor.A > 0 {
		paint.Stroke = edgeColor
		paint.LineWidth = f.EdgeWidth
		paint.LineJoin = render.JoinRound
		paint.LineCap = render.CapButt
		if f.EdgeWidth <= 1.5 {
			paint.Snap = render.SnapOn
		}
	}

	regions := f.fillRegions()
	for _, region := range regions {
		fillPath := f.createFillPathForRegion(region, ctx)
		if len(fillPath.C) == 0 {
			continue
		}
		r.Path(fillPath, &paint)
	}
}

// createFillPath creates a closed path for the fill area.
func (f *Fill2D) createFillPath(n int, ctx *DrawContext) geom.Path {
	if n <= 0 {
		return geom.Path{}
	}
	return f.createFillPathForRegion(fillRegion{start: 0, end: n}, ctx)
}

func (f *Fill2D) createFillPathForRegion(region fillRegion, ctx *DrawContext) geom.Path {
	path := geom.Path{}
	samples := f.regionSamples(region)
	n := len(samples.x)
	if n < 2 {
		return path
	}

	// Matplotlib's fill_between PolyCollection starts on the second boundary,
	// walks the first boundary forward, then walks the second boundary backward.
	// The duplicate endpoint vertices are intentional: they match upstream's
	// closed polygon path and affect antialiased edge coverage.
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, ctx.DataToPixel.Apply(f.dataPoint(samples.start.X, samples.start.Y)))

	for i := 0; i < n; i++ {
		pt := ctx.DataToPixel.Apply(f.dataPoint(samples.x[i], samples.y1[i]))
		path.C = append(path.C, geom.LineTo)
		path.V = append(path.V, pt)
	}

	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, ctx.DataToPixel.Apply(f.dataPoint(samples.end.X, samples.end.Y)))

	for i := n - 1; i >= 0; i-- {
		pt := ctx.DataToPixel.Apply(f.dataPoint(samples.x[i], samples.y2[i]))
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

	var bounds geom.Rect
	have := false
	for _, region := range f.fillRegions() {
		samples := f.regionSamples(region)
		if len(samples.x) < 2 {
			continue
		}
		for _, pt := range samples.dataPoints(f.Orientation) {
			if !have {
				bounds = geom.Rect{Min: pt, Max: pt}
				have = true
				continue
			}
			bounds = expandRect(bounds, pt)
		}
	}
	if !have {
		return geom.Rect{}
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

type fillRegion struct {
	start int
	end   int
}

type fillRegionSamples struct {
	x     []float64
	y1    []float64
	y2    []float64
	start geom.Pt
	end   geom.Pt
}

func (s fillRegionSamples) dataPoints(orientation FillOrientation) []geom.Pt {
	points := make([]geom.Pt, 0, len(s.x)*2+2)
	points = append(points, orientFillPoint(s.start, orientation))
	for i := range s.x {
		points = append(points, orientFillPoint(geom.Pt{X: s.x[i], Y: s.y1[i]}, orientation))
	}
	points = append(points, orientFillPoint(s.end, orientation))
	for i := range s.x {
		points = append(points, orientFillPoint(geom.Pt{X: s.x[i], Y: s.y2[i]}, orientation))
	}
	return points
}

func orientFillPoint(pt geom.Pt, orientation FillOrientation) geom.Pt {
	if orientation == FillHorizontal {
		return geom.Pt{X: pt.Y, Y: pt.X}
	}
	return pt
}

func (f *Fill2D) fillRegions() []fillRegion {
	n := f.validCount()
	if n < 2 {
		return nil
	}
	mask := make([]bool, n)
	for i := 0; i < n; i++ {
		mask[i] = true
		if len(f.Where) > 0 {
			mask[i] = i < len(f.Where) && f.Where[i]
		}
		mask[i] = mask[i] && isFinite(f.X[i]) && isFinite(f.Y1[i]) && isFinite(f.secondAt(i))
	}

	regions := []fillRegion{}
	for i := 0; i < n; {
		for i < n && !mask[i] {
			i++
		}
		start := i
		for i < n && mask[i] {
			i++
		}
		if i-start >= 2 {
			regions = append(regions, fillRegion{start: start, end: i})
		}
	}
	return regions
}

func (f *Fill2D) validCount() int {
	n := len(f.X)
	if len(f.Y1) < n {
		n = len(f.Y1)
	}
	if f.Y2 != nil && len(f.Y2) < n {
		n = len(f.Y2)
	}
	if len(f.Where) > 0 && len(f.Where) < n {
		n = len(f.Where)
	}
	return n
}

func (f *Fill2D) secondAt(i int) float64 {
	if f.Y2 != nil {
		return f.Y2[i]
	}
	return f.Baseline
}

func (f *Fill2D) regionSamples(region fillRegion) fillRegionSamples {
	x := append([]float64(nil), f.X[region.start:region.end]...)
	y1 := append([]float64(nil), f.Y1[region.start:region.end]...)
	y2 := make([]float64, len(x))
	for i := range y2 {
		y2[i] = f.secondAt(region.start + i)
	}
	if f.Step != FillStepNone {
		x, y1, y2 = stepFillSamples(x, y1, y2, f.Step)
	}

	start := geom.Pt{X: x[0], Y: y2[0]}
	end := geom.Pt{X: x[len(x)-1], Y: y2[len(y2)-1]}
	if f.Interpolate {
		start = f.interpolatingPoint(region.start)
		end = f.interpolatingPoint(region.end)
	}
	return fillRegionSamples{x: x, y1: y1, y2: y2, start: start, end: end}
}

func (f *Fill2D) interpolatingPoint(idx int) geom.Pt {
	n := f.validCount()
	if n == 0 {
		return geom.Pt{}
	}
	if idx <= 0 {
		return geom.Pt{X: f.X[0], Y: f.Y1[0]}
	}
	if idx >= n {
		last := n - 1
		return geom.Pt{X: f.X[last], Y: f.Y1[last]}
	}
	prev := idx - 1
	x0, x1 := f.X[prev], f.X[idx]
	y10, y11 := f.Y1[prev], f.Y1[idx]
	y20, y21 := f.secondAt(prev), f.secondAt(idx)
	d0 := y10 - y20
	d1 := y11 - y21
	if d0 == d1 {
		return geom.Pt{X: x1, Y: y11}
	}
	t := -d0 / (d1 - d0)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	x := x0 + (x1-x0)*t
	y := y10 + (y11-y10)*t
	return geom.Pt{X: x, Y: y}
}

func stepFillSamples(x, y1, y2 []float64, step FillStep) ([]float64, []float64, []float64) {
	if len(x) < 2 {
		return x, y1, y2
	}
	switch step {
	case FillStepPre:
		outX, outY1, outY2 := make([]float64, 0, len(x)*2-1), make([]float64, 0, len(x)*2-1), make([]float64, 0, len(x)*2-1)
		outX = append(outX, x[0])
		outY1 = append(outY1, y1[0])
		outY2 = append(outY2, y2[0])
		for i := 1; i < len(x); i++ {
			outX = append(outX, x[i-1], x[i])
			outY1 = append(outY1, y1[i], y1[i])
			outY2 = append(outY2, y2[i], y2[i])
		}
		return outX, outY1, outY2
	case FillStepPost:
		outX, outY1, outY2 := make([]float64, 0, len(x)*2-1), make([]float64, 0, len(x)*2-1), make([]float64, 0, len(x)*2-1)
		for i := 0; i < len(x)-1; i++ {
			outX = append(outX, x[i], x[i+1])
			outY1 = append(outY1, y1[i], y1[i])
			outY2 = append(outY2, y2[i], y2[i])
		}
		outX = append(outX, x[len(x)-1])
		outY1 = append(outY1, y1[len(y1)-1])
		outY2 = append(outY2, y2[len(y2)-1])
		return outX, outY1, outY2
	case FillStepMid:
		outX, outY1, outY2 := make([]float64, 0, len(x)*2), make([]float64, 0, len(x)*2), make([]float64, 0, len(x)*2)
		outX = append(outX, x[0])
		outY1 = append(outY1, y1[0])
		outY2 = append(outY2, y2[0])
		for i := 0; i < len(x)-1; i++ {
			mid := (x[i] + x[i+1]) / 2
			outX = append(outX, mid, mid)
			outY1 = append(outY1, y1[i], y1[i+1])
			outY2 = append(outY2, y2[i], y2[i+1])
		}
		outX = append(outX, x[len(x)-1])
		outY1 = append(outY1, y1[len(y1)-1])
		outY2 = append(outY2, y2[len(y2)-1])
		return outX, outY1, outY2
	default:
		return x, y1, y2
	}
}

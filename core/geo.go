package core

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

const (
	geoFrameSegments    = 160
	geoGridSegments     = 75
	geoLongitudeGridCap = 5 * math.Pi / 12
	geoXAxisLabelPadPx  = 4.0
	geoYAxisLabelPadPx  = 8.0
)

type geoProjection struct {
	name      string
	transform transform.T
}

func newMollweideProjection() *geoProjection {
	return &geoProjection{name: "mollweide", transform: mollweideDataTransform{}}
}

func isGeoProjection(proj Projection) bool {
	_, ok := proj.(*geoProjection)
	return ok
}

func (p *geoProjection) Name() string {
	if p == nil || p.name == "" {
		return "mollweide"
	}
	return p.name
}

func (p *geoProjection) CloneProjection() Projection {
	if p == nil {
		return newMollweideProjection()
	}
	clone := *p
	return &clone
}

func (p *geoProjection) ConfigureAxes(ax *Axes) {
	if ax == nil {
		return
	}

	ax.XScale = transform.NewLinear(-math.Pi, math.Pi)
	ax.YScale = transform.NewLinear(-math.Pi/2, math.Pi/2)
	boxAspect := 0.5
	if p.Name() == "lambert" {
		boxAspect = 1
	}
	_ = ax.SetBoxAspect(boxAspect)
	ax.XAxis = NewXAxis()
	ax.YAxis = NewYAxis()
	ax.XAxisTop = nil
	ax.YAxisRight = nil
	ax.ShowFrame = false

	longitudeTicks := []float64{-5 * math.Pi / 6, -2 * math.Pi / 3, -math.Pi / 2, -math.Pi / 3, -math.Pi / 6, 0, math.Pi / 6, math.Pi / 3, math.Pi / 2, 2 * math.Pi / 3, 5 * math.Pi / 6}
	latitudeTicks := []float64{-5 * math.Pi / 12, -math.Pi / 3, -math.Pi / 4, -math.Pi / 6, -math.Pi / 12, 0, math.Pi / 12, math.Pi / 6, math.Pi / 4, math.Pi / 3, 5 * math.Pi / 12}

	ax.XAxis.Locator = FixedLocator{TicksList: longitudeTicks}
	ax.XAxis.MinorLocator = NullLocator{}
	ax.XAxis.Formatter = FuncFormatter(formatGeoDegreeLabel)
	ax.XAxis.ShowSpine = true
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = true

	ax.YAxis.Locator = FixedLocator{TicksList: latitudeTicks}
	ax.YAxis.MinorLocator = NullLocator{}
	ax.YAxis.Formatter = FuncFormatter(formatGeoDegreeLabel)
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = true
	if p.Name() == "lambert" {
		ax.YAxis.Formatter = NullFormatter{}
	}
}

func (p *geoProjection) DataToAxes(*Axes) transform.T {
	if p == nil || p.transform == nil {
		return mollweideDataTransform{}
	}
	return p.transform
}

func (p *geoProjection) FramePath(clip geom.Rect) geom.Path {
	return geoEllipsePath(clip)
}

func (p *geoProjection) ContainsDisplayPoint(clip geom.Rect, pt geom.Pt) bool {
	center := geom.Pt{X: clip.Min.X + clip.W()/2, Y: clip.Min.Y + clip.H()/2}
	rx := clip.W() / 2
	ry := clip.H() / 2
	if rx <= 0 || ry <= 0 {
		return false
	}
	dx := (pt.X - center.X) / rx
	dy := (pt.Y - center.Y) / ry
	return dx*dx+dy*dy <= 1+1e-12
}

type mollweideDataTransform struct{}

func (mollweideDataTransform) Apply(p geom.Pt) geom.Pt {
	lon := clamp(p.X, -math.Pi, math.Pi)
	lat := clamp(p.Y, -math.Pi/2, math.Pi/2)
	theta := mollweideTheta(lat)
	x := 2 * math.Sqrt2 / math.Pi * lon * math.Cos(theta)
	y := math.Sqrt2 * math.Sin(theta)
	return geom.Pt{
		X: 0.5 + x/(4*math.Sqrt2),
		Y: 0.5 + y/(2*math.Sqrt2),
	}
}

func (mollweideDataTransform) Invert(p geom.Pt) (geom.Pt, bool) {
	x := (p.X - 0.5) * 4 * math.Sqrt2
	y := (p.Y - 0.5) * 2 * math.Sqrt2
	if math.Abs(y) > math.Sqrt2+1e-9 {
		return geom.Pt{}, false
	}
	y = clamp(y, -math.Sqrt2, math.Sqrt2)
	theta := math.Asin(y / math.Sqrt2)
	cosTheta := math.Cos(theta)

	lon := 0.0
	if math.Abs(cosTheta) > 1e-12 {
		lon = math.Pi * x / (2 * math.Sqrt2 * cosTheta)
	}
	latArg := (2*theta + math.Sin(2*theta)) / math.Pi
	lat := math.Asin(clamp(latArg, -1, 1))

	if lon < -math.Pi-1e-9 || lon > math.Pi+1e-9 {
		return geom.Pt{}, false
	}
	return geom.Pt{
		X: clamp(lon, -math.Pi, math.Pi),
		Y: lat,
	}, true
}

func mollweideTheta(lat float64) float64 {
	if approx(lat, math.Pi/2, 1e-14) {
		return math.Pi / 2
	}
	if approx(lat, -math.Pi/2, 1e-14) {
		return -math.Pi / 2
	}

	theta := lat
	target := math.Pi * math.Sin(lat)
	for range 12 {
		f := 2*theta + math.Sin(2*theta) - target
		fp := 2 + 2*math.Cos(2*theta)
		if math.Abs(fp) < 1e-14 {
			break
		}
		next := theta - f/fp
		if math.Abs(next-theta) < 1e-14 {
			return next
		}
		theta = next
	}
	return theta
}

func geoEllipsePath(clip geom.Rect) geom.Path {
	center := geom.Pt{X: clip.Min.X + clip.W()/2, Y: clip.Min.Y + clip.H()/2}
	rx := clip.W() / 2
	ry := clip.H() / 2
	path := geom.Path{}
	if rx <= 0 || ry <= 0 {
		return path
	}
	for i := 0; i <= geoFrameSegments; i++ {
		t := 2 * math.Pi * float64(i) / float64(geoFrameSegments)
		pt := geom.Pt{X: center.X + rx*math.Cos(t), Y: center.Y + ry*math.Sin(t)}
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}

func drawGeoGridLine(r render.Renderer, ctx *DrawContext, axis AxisSide, tick float64, paint render.Paint) {
	if ctx == nil {
		return
	}
	xMin, xMax := ctx.DataToPixel.XScale.Domain()
	path := geom.Path{}

	for i := 0; i <= geoGridSegments; i++ {
		t := float64(i) / float64(geoGridSegments)
		var data geom.Pt
		switch axis {
		case AxisBottom, AxisTop:
			yMin := -geoLongitudeGridCap
			yMax := geoLongitudeGridCap
			data = geom.Pt{X: tick, Y: yMin + (yMax-yMin)*t}
		default:
			data = geom.Pt{X: xMin + (xMax-xMin)*t, Y: tick}
		}
		pt := ctx.DataToPixel.Apply(data)
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	r.Path(path, &paint)
}

func (a *Axis) drawGeoTickLabels(r render.Renderer, ctx *DrawContext) {
	textRen, ok := r.(render.TextDrawer)
	if !ok || a == nil || ctx == nil || !a.ShowLabels || a.Locator == nil || a.Formatter == nil {
		return
	}

	domainMin, domainMax, isXAxis, ok := geoAxisDomain(a, ctx)
	if !ok {
		return
	}
	ticks := visibleTicks(a.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
	fontSize := tickLabelFontSize(a, ctx)
	for i, tick := range ticks {
		label := formatTickLabel(a.Formatter, tick, i, ticks)
		if label == "" {
			continue
		}
		layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		origin, ok := geoTickLabelOrigin(a, ctx, tick, layout)
		if !ok {
			continue
		}
		drawDisplayText(textRen, label, origin, fontSize, a.Color, ctx.RC.FontKey, ctx.RC.UseTeX)
	}
}

func (a *Axis) geoTickLabelBounds(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if a == nil || ctx == nil || !a.ShowLabels || a.Locator == nil || a.Formatter == nil {
		return geom.Rect{}, false
	}
	domainMin, domainMax, isXAxis, ok := geoAxisDomain(a, ctx)
	if !ok {
		return geom.Rect{}, false
	}
	ticks := visibleTicks(a.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
	fontSize := tickLabelFontSize(a, ctx)

	var union geom.Rect
	var have bool
	for i, tick := range ticks {
		label := formatTickLabel(a.Formatter, tick, i, ticks)
		if label == "" {
			continue
		}
		layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		origin, ok := geoTickLabelOrigin(a, ctx, tick, layout)
		if !ok {
			continue
		}
		bounds, ok := textInkRect(origin, layout)
		if !ok {
			continue
		}
		if !have {
			union = bounds
			have = true
			continue
		}
		union = geom.Rect{
			Min: geom.Pt{X: math.Min(union.Min.X, bounds.Min.X), Y: math.Min(union.Min.Y, bounds.Min.Y)},
			Max: geom.Pt{X: math.Max(union.Max.X, bounds.Max.X), Y: math.Max(union.Max.Y, bounds.Max.Y)},
		}
	}
	return union, have
}

func geoAxisDomain(a *Axis, ctx *DrawContext) (minVal, maxVal float64, isXAxis bool, ok bool) {
	switch a.Side {
	case AxisBottom, AxisTop:
		minVal, maxVal = ctx.DataToPixel.XScale.Domain()
		return minVal, maxVal, true, true
	case AxisLeft, AxisRight:
		minVal, maxVal = ctx.DataToPixel.YScale.Domain()
		return minVal, maxVal, false, true
	default:
		return 0, 0, false, false
	}
}

func geoTickLabelOrigin(a *Axis, ctx *DrawContext, tick float64, layout singleLineTextLayout) (geom.Pt, bool) {
	switch a.Side {
	case AxisBottom:
		pos := ctx.DataToPixel.Apply(geom.Pt{X: tick, Y: 0})
		anchor := geom.Pt{X: pos.X, Y: pos.Y + geoXAxisLabelPadPx}
		return alignedSingleLineOrigin(anchor, layout, TextAlignCenter, textLayoutVAlignBottom), true
	case AxisTop:
		pos := ctx.DataToPixel.Apply(geom.Pt{X: tick, Y: 0})
		anchor := geom.Pt{X: pos.X, Y: pos.Y + geoXAxisLabelPadPx}
		return alignedSingleLineOrigin(anchor, layout, TextAlignCenter, textLayoutVAlignTop), true
	case AxisLeft:
		pos := geoLatitudeLabelPoint(ctx, -math.Pi, tick)
		anchor := geom.Pt{X: pos.X - geoYAxisLabelPadPx, Y: pos.Y}
		return alignedSingleLineOrigin(anchor, layout, TextAlignRight, textLayoutVAlignCenter), true
	case AxisRight:
		pos := geoLatitudeLabelPoint(ctx, math.Pi, tick)
		anchor := geom.Pt{X: pos.X + geoYAxisLabelPadPx, Y: pos.Y}
		return alignedSingleLineOrigin(anchor, layout, TextAlignLeft, textLayoutVAlignCenter), true
	default:
		return geom.Pt{}, false
	}
}

func geoLatitudeLabelPoint(ctx *DrawContext, lon, lat float64) geom.Pt {
	if ctx == nil {
		return geom.Pt{}
	}
	axesPt := ctx.TransProjection().Apply(geom.Pt{X: lon, Y: lat})
	axesPt.Y = 0.5 + (axesPt.Y-0.5)*1.1
	return ctx.TransAxes().Apply(axesPt)
}

func formatGeoDegreeLabel(rad float64) string {
	deg := rad * 180 / math.Pi
	if approx(deg, math.Round(deg), 1e-9) {
		deg = math.Round(deg)
	}
	if approx(deg, 0, 1e-9) {
		return "0"
	}
	return fmt.Sprintf("%.0f", deg)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

// Grid renders grid lines at tick positions.
type Grid struct {
	Axis           AxisSide     // which axis to use for tick positions
	Color          render.Color // major grid line color
	LineWidth      float64      // width of major grid lines
	Dashes         []float64    // dash pattern for major grid (nil = solid)
	Alpha          float64      // alpha override (0-1), if 0 uses Color.A
	Major          bool         // draw grid at major ticks
	Minor          bool         // draw grid at minor ticks
	MinorColor     render.Color // minor grid line color (zero value uses Color with lower alpha)
	MinorLineWidth float64      // width of minor grid lines (0 uses LineWidth*0.5)
	MinorDashes    []float64    // dash pattern for minor grid (nil = solid)
	Locator        Locator      // major tick locator (nil = owning axis locator, then LinearLocator)
	MinorLocator   Locator      // minor tick locator (nil = owning axis minor locator, then MinorLinearLocator{N:5})
	z              float64      // z-order (should be behind data)
}

// NewGrid creates a new grid for the specified axis.
func NewGrid(axis AxisSide) *Grid {
	rc := style.CurrentDefaults()
	return &Grid{
		Axis:           axis,
		Color:          rc.GridColor,
		LineWidth:      rc.GridLineWidth,
		MinorColor:     rc.MinorGridColor,
		MinorLineWidth: rc.MinorGridLineWidth,
		Alpha:          0, // use Color.A
		Major:          true,
		Minor:          false,
		z:              -1000, // behind everything else
	}
}

// Draw renders grid lines at tick positions.
func (g *Grid) Draw(r render.Renderer, ctx *DrawContext) {
	if !g.Major && !g.Minor {
		return
	}
	if isPolarProjection(ctx.Projection) {
		g.drawPolar(r, ctx)
		return
	}
	if isGeoProjection(ctx.Projection) {
		g.drawGeo(r, ctx)
		return
	}
	if projectionGridUsesSampling(ctx) {
		g.drawCurvelinear(r, ctx)
		return
	}

	var domainMin, domainMax float64
	var isXAxis bool

	switch g.Axis {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
	}

	axis := g.axisForContext(ctx)
	majorColor := g.Color
	if g.Alpha > 0 && g.Alpha <= 1 {
		majorColor.A = g.Alpha
	}

	// Draw minor grid first (behind major)
	if g.Minor {
		minorLoc := g.MinorLocator
		if minorLoc == nil && axis != nil && axis.MinorLocator != nil {
			minorLoc = axis.MinorLocator
		}
		if minorLoc == nil {
			minorLoc = MinorLinearLocator{N: 5}
		}
		minorTicks := visibleTicks(minorLoc.Ticks(domainMin, domainMax, g.cartesianTargetTickCount(axis, true, ctx, isXAxis)), domainMin, domainMax)

		minorColor := g.MinorColor
		if minorColor == (render.Color{}) {
			minorColor = majorColor
			minorColor.A = majorColor.A * 0.4
		}
		minorWidth := g.MinorLineWidth
		if minorWidth <= 0 {
			minorWidth = g.LineWidth * 0.5
		}

		for _, v := range minorTicks {
			g.drawLine(r, ctx, v, isXAxis, minorColor, minorWidth, g.MinorDashes)
		}
	}

	// Draw major grid
	if g.Major {
		loc := g.Locator
		if loc == nil && axis != nil && axis.Locator != nil {
			loc = axis.Locator
		}
		if loc == nil {
			loc = LinearLocator{}
		}
		ticks := visibleTicks(loc.Ticks(domainMin, domainMax, g.cartesianTargetTickCount(axis, false, ctx, isXAxis)), domainMin, domainMax)

		for _, v := range ticks {
			g.drawLine(r, ctx, v, isXAxis, majorColor, g.LineWidth, g.Dashes)
		}
	}
}

func (g *Grid) drawGeo(r render.Renderer, ctx *DrawContext) {
	var domainMin, domainMax float64
	switch g.Axis {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
	}

	majorColor := g.Color
	if g.Alpha > 0 && g.Alpha <= 1 {
		majorColor.A = g.Alpha
	}

	if g.Minor {
		minorLoc := g.MinorLocator
		if minorLoc == nil {
			minorLoc = MinorLinearLocator{N: 2}
		}
		minorColor := g.MinorColor
		if minorColor == (render.Color{}) {
			minorColor = majorColor
			minorColor.A = majorColor.A * 0.4
		}
		minorWidth := g.MinorLineWidth
		if minorWidth <= 0 {
			minorWidth = g.LineWidth * 0.5
		}
		paint := polarTickPaint(minorColor, minorWidth, g.MinorDashes)
		for _, tick := range visibleTicks(minorLoc.Ticks(domainMin, domainMax, 20), domainMin, domainMax) {
			drawGeoGridLine(r, ctx, g.Axis, tick, paint)
		}
	}

	if g.Major {
		loc := g.Locator
		if loc == nil {
			switch g.Axis {
			case AxisBottom, AxisTop:
				loc = ctx.Axes.effectiveXAxis().Locator
			default:
				loc = ctx.Axes.effectiveYAxis().Locator
			}
		}
		paint := polarTickPaint(majorColor, g.LineWidth, g.Dashes)
		for _, tick := range visibleTicks(loc.Ticks(domainMin, domainMax, 8), domainMin, domainMax) {
			drawGeoGridLine(r, ctx, g.Axis, tick, paint)
		}
	}
}

func (g *Grid) drawCurvelinear(r render.Renderer, ctx *DrawContext) {
	var domainMin, domainMax float64
	switch g.Axis {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
		if minVal, maxVal, ok := skewXGridXTickDomain(ctx); ok {
			domainMin, domainMax = minVal, maxVal
		}
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
	}

	axis := g.axisForContext(ctx)
	majorColor := g.Color
	if g.Alpha > 0 && g.Alpha <= 1 {
		majorColor.A = g.Alpha
	}

	if g.Minor {
		minorLoc := g.curvelinearLocator(axis, true)
		minorColor := g.MinorColor
		if minorColor == (render.Color{}) {
			minorColor = majorColor
			minorColor.A = majorColor.A * 0.4
		}
		minorWidth := g.MinorLineWidth
		if minorWidth <= 0 {
			minorWidth = g.LineWidth * 0.5
		}
		paint := polarTickPaint(minorColor, minorWidth, g.MinorDashes)
		for _, tick := range visibleTicks(minorLoc.Ticks(domainMin, domainMax, g.curvelinearTargetTickCount(axis, true)), domainMin, domainMax) {
			drawSampledGridLine(r, ctx, g.Axis, tick, geoGridSegments, paint)
		}
	}

	if g.Major {
		loc := g.curvelinearLocator(axis, false)
		paint := polarTickPaint(majorColor, g.LineWidth, g.Dashes)
		for _, tick := range visibleTicks(loc.Ticks(domainMin, domainMax, g.curvelinearTargetTickCount(axis, false)), domainMin, domainMax) {
			drawSampledGridLine(r, ctx, g.Axis, tick, geoGridSegments, paint)
		}
	}
}

func (g *Grid) drawPolar(r render.Renderer, ctx *DrawContext) {
	majorColor := g.Color
	if g.Alpha > 0 && g.Alpha <= 1 {
		majorColor.A = g.Alpha
	}

	if g.Minor {
		minorColor := g.MinorColor
		if minorColor == (render.Color{}) {
			minorColor = majorColor
			minorColor.A = majorColor.A * 0.4
		}
		minorWidth := g.MinorLineWidth
		if minorWidth <= 0 {
			minorWidth = g.LineWidth * 0.5
		}
		g.drawPolarLines(r, ctx, true, minorColor, minorWidth, g.MinorDashes)
	}

	if g.Major {
		g.drawPolarLines(r, ctx, false, majorColor, g.LineWidth, g.Dashes)
	}
}

func (g *Grid) drawPolarLines(r render.Renderer, ctx *DrawContext, minor bool, color render.Color, width float64, dashes []float64) {
	center, outerRadius := polarCenterAndRadius(ctx.Clip)
	paint := polarTickPaint(color, width, dashes)

	switch g.Axis {
	case AxisBottom, AxisTop:
		axis := axisForPolarGrid(ctx, true)
		ticks := polarThetaTicks(axis, ctx.DataToPixel.XScale, minor)
		if locator := g.polarLocator(axis, minor); locator != nil && ctx.DataToPixel.XScale != nil {
			minVal, maxVal := ctx.DataToPixel.XScale.Domain()
			ticks = visibleTicks(locator.Ticks(minVal, maxVal, polarTargetTickCount(axis, minor)), minVal, maxVal)
		}
		for _, tick := range ticks {
			angle := polarAngleForTheta(ctx.Projection, ctx.DataToPixel.XScale, tick)
			path := geom.Path{}
			path.MoveTo(center)
			path.LineTo(polarPixelPoint(center, outerRadius, angle))
			// Axis-aligned spokes (e.g. the radar north spoke) snap in
			// matplotlib's renderer; diagonal spokes pass through.
			r.Path(snapPathToPixels(path, paint.LineWidth), &paint)
		}
	case AxisLeft, AxisRight:
		axis := axisForPolarGrid(ctx, false)
		ticks := polarRadialTicks(axis, ctx.DataToPixel.YScale, minor)
		if locator := g.polarLocator(axis, minor); locator != nil && ctx.DataToPixel.YScale != nil {
			minVal, maxVal := ctx.DataToPixel.YScale.Domain()
			ticks = visibleTicks(locator.Ticks(minVal, maxVal, polarTargetTickCount(axis, minor)), minVal, maxVal)
		}
		for _, tick := range ticks {
			radius := outerRadius * ctx.DataToPixel.YScale.Fwd(tick)
			if radius <= 0 {
				continue
			}
			path := polarCirclePath(center, radius)
			if sides := radarFrameSidesForProjection(ctx.Projection); sides >= 3 {
				path = polarPolygonFramePath(ctx.Projection, center, radius, sides)
			}
			r.Path(snapPathToPixels(path, paint.LineWidth), &paint)
		}
	}
}

func axisForPolarGrid(ctx *DrawContext, isTheta bool) *Axis {
	if ctx == nil || ctx.Axes == nil {
		return nil
	}
	if isTheta {
		return ctx.Axes.effectiveXAxis()
	}
	return ctx.Axes.effectiveYAxis()
}

func (g *Grid) polarLocator(axis *Axis, minor bool) Locator {
	if minor {
		if g.MinorLocator != nil {
			return g.MinorLocator
		}
		if axis != nil {
			return axis.MinorLocator
		}
		return nil
	}
	if g.Locator != nil {
		return g.Locator
	}
	if axis != nil {
		return axis.Locator
	}
	return nil
}

func polarTargetTickCount(axis *Axis, minor bool) int {
	if axis == nil {
		if minor {
			return 30
		}
		return 6
	}
	if minor {
		return axis.minorTickTargetCount()
	}
	return axis.majorTickTargetCount()
}

func projectionGridUsesSampling(ctx *DrawContext) bool {
	if ctx == nil || ctx.Projection == nil || isPolarProjection(ctx.Projection) || isGeoProjection(ctx.Projection) {
		return false
	}
	if _, ok := ctx.TransProjection().(transform.Separable); ok {
		return false
	}
	return true
}

func (g *Grid) axisForContext(ctx *DrawContext) *Axis {
	if g == nil || ctx == nil || ctx.Axes == nil {
		return nil
	}
	switch g.Axis {
	case AxisTop:
		if axis := ctx.Axes.effectiveTopAxis(); axis != nil {
			return axis
		}
		return ctx.Axes.effectiveXAxis()
	case AxisRight:
		if axis := ctx.Axes.effectiveRightAxis(); axis != nil {
			return axis
		}
		return ctx.Axes.effectiveYAxis()
	case AxisLeft:
		return ctx.Axes.effectiveYAxis()
	default:
		return ctx.Axes.effectiveXAxis()
	}
}

func (g *Grid) curvelinearLocator(axis *Axis, minor bool) Locator {
	if minor {
		if g.MinorLocator != nil {
			return g.MinorLocator
		}
		if axis != nil && axis.MinorLocator != nil {
			return axis.MinorLocator
		}
		return MinorLinearLocator{N: 5}
	}
	if g.Locator != nil {
		return g.Locator
	}
	if axis != nil && axis.Locator != nil {
		return axis.Locator
	}
	return LinearLocator{}
}

func (g *Grid) curvelinearTargetTickCount(axis *Axis, minor bool) int {
	if axis == nil {
		if minor {
			return 30
		}
		return 8
	}
	if minor {
		return axis.minorTickTargetCount()
	}
	return axis.majorTickTargetCount()
}

func (g *Grid) cartesianTargetTickCount(axis *Axis, minor bool, ctx *DrawContext, isXAxis bool) int {
	if axis == nil {
		if minor {
			return 30
		}
		return 8
	}
	if minor {
		return axis.minorTickTargetCountForContext(ctx, isXAxis)
	}
	return axis.majorTickTargetCountForContext(ctx, isXAxis)
}

// drawLine draws a single grid line.
func (g *Grid) drawLine(r render.Renderer, ctx *DrawContext, tickValue float64, isXAxis bool, color render.Color, width float64, dashes []float64) {
	var p1, p2 geom.Pt

	if isXAxis {
		yMin, yMax := ctx.DataToPixel.YScale.Domain()
		p1 = ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: yMin})
		p2 = ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: yMax})
		// Matplotlib gridlines are Line2D artists with snap=None; AGG's
		// PathSnapper then auto-snaps horizontal/vertical paths to pixel centers
		// (third_party/matplotlib/src/path_converters.h).
		x := math.Round(p1.X) + 0.5
		p1.X = x
		p2.X = x
	} else {
		xMin, xMax := ctx.DataToPixel.XScale.Domain()
		p1 = ctx.DataToPixel.Apply(geom.Pt{X: xMin, Y: tickValue})
		p2 = ctx.DataToPixel.Apply(geom.Pt{X: xMax, Y: tickValue})
		y := math.Round(p1.Y) - 0.5
		p1.Y = y
		p2.Y = y
	}

	path := geom.Path{}
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, p1)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, p2)

	paint := render.Paint{
		LineWidth: width,
		Stroke:    color,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
		Dashes:    scaleGridDashes(dashes, width),
	}
	r.Path(path, &paint)
}

// snapPathToPixels ports matplotlib's PathSnapper in SNAP_AUTO mode
// (third_party/matplotlib/src/path_converters.h): a path of at most 1024
// vertices consisting solely of horizontal or vertical line segments has every
// vertex snapped to floor(v+0.5)+s, where s is 0.5 when the stroke width
// rounds to an odd integer and 0 otherwise. Any diagonal segment or curve
// returns the path unchanged. matplotlib applies this in the renderer to every
// path drawn with snap mode "auto" (the Line2D default); here it is applied at
// the gridline call sites, which draw plain polylines.
func snapPathToPixels(path geom.Path, strokeWidth float64) geom.Path {
	if len(path.V) == 0 || len(path.V) > 1024 {
		return path
	}
	vi := 0
	var prev geom.Pt
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			prev = path.V[vi]
			vi++
		case geom.LineTo:
			cur := path.V[vi]
			vi++
			if math.Abs(prev.X-cur.X) >= 1e-4 && math.Abs(prev.Y-cur.Y) >= 1e-4 {
				return path
			}
			prev = cur
		case geom.ClosePath:
			// The implicit closing segment is not inspected, matching the
			// upstream vertex-stream behavior.
		default:
			// QuadTo/CubicTo: curves are never snapped.
			return path
		}
	}
	snapValue := 0.0
	if int(math.Round(strokeWidth))%2 != 0 {
		snapValue = 0.5
	}
	out := geom.Path{C: append([]geom.Cmd(nil), path.C...), V: make([]geom.Pt, len(path.V))}
	for i, v := range path.V {
		out.V[i] = geom.Pt{X: math.Floor(v.X+0.5) + snapValue, Y: math.Floor(v.Y+0.5) + snapValue}
	}
	return out
}

// scaleGridDashes mirrors matplotlib's Line2D dash scaling
// (rcParams["lines.scale_dashes"] is True by default): the on/off pattern is
// multiplied by the line width. Because both the pattern and width are already
// in pixels here, the rendered dash length equals pattern*lineWidthPx — the
// same result matplotlib produces from a points-based pattern at this DPI.
func scaleGridDashes(dashes []float64, width float64) []float64 {
	if len(dashes) == 0 || width <= 0 {
		return dashes
	}
	scaled := make([]float64, len(dashes))
	for i, d := range dashes {
		scaled[i] = d * width
	}
	return scaled
}

func drawSampledGridLine(r render.Renderer, ctx *DrawContext, axis AxisSide, tick float64, segments int, paint render.Paint) {
	if ctx == nil || segments <= 0 {
		return
	}
	xMin, xMax := ctx.DataToPixel.XScale.Domain()
	yMin, yMax := ctx.DataToPixel.YScale.Domain()
	path := geom.Path{}

	for i := 0; i <= segments; i++ {
		t := float64(i) / float64(segments)
		var data geom.Pt
		switch axis {
		case AxisBottom, AxisTop:
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
	r.Path(snapPathToPixels(path, paint.LineWidth), &paint)
}

// Z returns the z-order for sorting.
func (g *Grid) Z() float64 {
	return g.z
}

// Bounds returns an empty rect for now.
func (g *Grid) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}

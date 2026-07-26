package plot3d

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

func (a *Axes3D) frameSegments(mins, maxs vec3) [][]geom.Pt {
	return a.frameSegmentsProjected(mins, maxs, mins, maxs, mins, maxs, 9)
}

func (a *Axes3D) frameSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs vec3, nbins int) [][]geom.Pt {
	return a.frameGridSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs, nbins)
}

func (a *Axes3D) activePanePolygons(mins, maxs vec3) [][]geom.Pt {
	return a.activePanePolygonsProjected(mins, maxs, mins, maxs)
}

func (a *Axes3D) activePanePolygonsProjected(mins, maxs, projMins, projMaxs vec3) [][]geom.Pt {
	planes := [6][4][3]int{
		{{0, 0, 0}, {0, 1, 0}, {0, 1, 1}, {0, 0, 1}},
		{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}},
		{{0, 1, 0}, {1, 1, 0}, {1, 1, 1}, {0, 1, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}},
		{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}},
	}
	project := func(corner [3]int) geom.Pt {
		x := mins[0]
		if corner[0] == 1 {
			x = maxs[0]
		}
		y := mins[1]
		if corner[1] == 1 {
			y = maxs[1]
		}
		z := mins[2]
		if corner[2] == 1 {
			z = maxs[2]
		}
		return a.project3DPointWithState(x, y, z, projMins, projMaxs)
	}

	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	panes := make([][]geom.Pt, 0, 3)
	for axis := range 3 {
		planeIndex := 2 * axis
		if highs[axis] {
			planeIndex++
		}
		plane := planes[planeIndex]
		polygon := make([]geom.Pt, len(plane))
		for i, corner := range plane {
			polygon[i] = project(corner)
		}
		panes = append(panes, polygon)
	}
	return panes
}

func (a *Axes3D) axisLineSegmentsProjected(mins, maxs, projMins, projMaxs vec3) [][]geom.Pt {
	project := func(p vec3) geom.Pt {
		return a.project3DPointWithState(p[0], p[1], p[2], projMins, projMaxs)
	}
	pairs := a.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)
	segments := make([][]geom.Pt, 0, len(pairs))
	for _, pair := range pairs {
		segments = append(segments, []geom.Pt{project(pair[0]), project(pair[1])})
	}
	return segments
}

func (a *Axes3D) axisTickSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs vec3, nbins int) [][]geom.Pt {
	project := func(p vec3) geom.Pt {
		return a.project3DPointWithState(p[0], p[1], p[2], projMins, projMaxs)
	}
	pairs := a.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)
	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	tickDirs := [3]int{1, 0, 0}
	segments := make([][]geom.Pt, 0, 24)
	for axis, pair := range pairs {
		tickDir := tickDirs[axis]
		tickDelta := (tickMaxs[tickDir] - tickMins[tickDir]) / 12
		if !highs[tickDir] {
			tickDelta = -tickDelta
		}
		outward := pair[0][tickDir] + 0.1*tickDelta
		inward := pair[0][tickDir] - 0.2*tickDelta
		for _, tick := range frameAxisTicks(tickMins[axis], tickMaxs[axis], nbins) {
			p0 := pair[0]
			p1 := pair[0]
			p0[axis] = tick
			p1[axis] = tick
			p0[tickDir] = outward
			p1[tickDir] = inward
			segments = append(segments, []geom.Pt{project(p0), project(p1)})
		}
	}
	return segments
}

func (a *Axes3D) axisLineEdgePointPairs(mins, maxs, projMins, projMaxs vec3) [][2]vec3 {
	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	minmax := vec3{}
	maxmin := vec3{}
	for i := range 3 {
		if highs[i] {
			minmax[i] = maxs[i]
			maxmin[i] = mins[i]
		} else {
			minmax[i] = mins[i]
			maxmin[i] = maxs[i]
		}
	}

	juggled := [3][3]int{
		{1, 0, 2},
		{0, 1, 2},
		{0, 2, 1},
	}
	pairs := make([][2]vec3, 0, 3)
	for axis := range 3 {
		p0 := minmax
		p0[juggled[axis][0]] = maxmin[juggled[axis][0]]
		p1 := p0
		p1[juggled[axis][1]] = maxmin[juggled[axis][1]]
		pairs = append(pairs, [2]vec3{p0, p1})
	}
	return pairs
}

func (a *Axes3D) frameGridSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs vec3, nbins int) [][]geom.Pt {
	project := func(p vec3) geom.Pt {
		return a.project3DPointWithState(p[0], p[1], p[2], projMins, projMaxs)
	}
	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	minmax := vec3{}
	maxmin := vec3{}
	for i := range 3 {
		if highs[i] {
			minmax[i] = maxs[i]
			maxmin[i] = mins[i]
		} else {
			minmax[i] = mins[i]
			maxmin[i] = maxs[i]
		}
	}

	segments := make([][]geom.Pt, 0, 18)
	limits := [][2]float64{
		{tickMins[0], tickMaxs[0]},
		{tickMins[1], tickMaxs[1]},
		{tickMins[2], tickMaxs[2]},
	}
	for index, lim := range limits {
		for _, tick := range frameAxisTicks(lim[0], lim[1], nbins) {
			p0 := minmax
			p1 := minmax
			p2 := minmax
			p0[index] = tick
			p1[index] = tick
			p2[index] = tick
			first := (index + 1) % 3
			second := (index + 2) % 3
			p0[first] = maxmin[first]
			p2[second] = maxmin[second]
			segments = append(segments, []geom.Pt{project(p0), project(p1), project(p2)})
		}
	}
	return segments
}

func (a *Axes3D) activePaneHighs(mins, maxs vec3) [3]bool {
	return a.activePaneHighsProjected(mins, maxs, mins, maxs)
}

func (a *Axes3D) activePaneHighsProjected(mins, maxs, projMins, projMaxs vec3) [3]bool {
	planes := [6][4]int{
		{0, 3, 7, 4},
		{1, 2, 6, 5},
		{0, 1, 5, 4},
		{3, 2, 6, 7},
		{0, 1, 2, 3},
		{4, 5, 6, 7},
	}
	corners := [8]vec3{
		{mins[0], mins[1], mins[2]},
		{maxs[0], mins[1], mins[2]},
		{maxs[0], maxs[1], mins[2]},
		{mins[0], maxs[1], mins[2]},
		{mins[0], mins[1], maxs[2]},
		{maxs[0], mins[1], maxs[2]},
		{maxs[0], maxs[1], maxs[2]},
		{mins[0], maxs[1], maxs[2]},
	}
	depths := [8]float64{}
	for i, corner := range corners {
		depths[i] = a.projectPointDepthWithProjectionLimits(corner[0], corner[1], corner[2], projMins, projMaxs)
	}

	means0 := [3]float64{}
	means1 := [3]float64{}
	highs := [3]bool{}
	equals := [3]bool{}
	equalCount := 0
	for axis := range 3 {
		means0[axis] = meanPlaneDepth(depths, planes[2*axis])
		means1[axis] = meanPlaneDepth(depths, planes[2*axis+1])
		highs[axis] = means0[axis] < means1[axis]
		if math.Abs(means0[axis]-means1[axis]) <= math.Nextafter(1, 2)-1 {
			equals[axis] = true
			equalCount++
		}
	}
	if equalCount == 2 {
		vertical := -1
		for i := range equals {
			if !equals[i] {
				vertical = i
				break
			}
		}
		switch vertical {
		case 2:
			highs[0], highs[1] = true, true
		case 1:
			highs[0], highs[2] = true, false
		case 0:
			highs[1], highs[2] = false, false
		}
	}
	return highs
}

func meanPlaneDepth(depths [8]float64, plane [4]int) float64 {
	return (depths[plane[0]] + depths[plane[1]] + depths[plane[2]] + depths[plane[3]]) / 4
}

func computed3DCollectionZ(projectedDepth float64) float64 {
	if math.IsNaN(projectedDepth) || math.IsInf(projectedDepth, 0) {
		return defaultPatchZ
	}
	return default3DComputedZ - projectedDepth
}

func (a *Axes3D) points3DCollectionZ(x, y, z []float64) float64 {
	n := minLen(x, y, z)
	depth := math.Inf(1)
	for i := 0; i < n; i++ {
		if !isFinite3D(x[i], y[i], z[i]) {
			continue
		}
		_, zDepth := a.projectPointDepth(x[i], y[i], z[i])
		if zDepth < depth {
			depth = zDepth
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) grid3DCollectionZ(x, y []float64, z [][]float64) float64 {
	if a == nil || len(z) == 0 {
		return defaultPatchZ
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return defaultPatchZ
	}
	depth := math.Inf(1)
	for row := 0; row < rows; row++ {
		if len(z[row]) != cols {
			return computed3DCollectionZ(depth)
		}
		for col := 0; col < cols; col++ {
			if !isFinite3D(x[col], y[row], z[row][col]) {
				continue
			}
			_, zDepth := a.projectPointDepth(x[col], y[row], z[row][col])
			if zDepth < depth {
				depth = zDepth
			}
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) bar3DCollectionZ(x, y, z, dx, dy, dz []float64) float64 {
	n := minLen(x, y, z, dx, dy, dz)
	depth := math.Inf(1)
	for i := 0; i < n; i++ {
		x0, x1 := x[i], x[i]+dx[i]
		y0, y1 := y[i], y[i]+dy[i]
		z0, z1 := z[i], z[i]+dz[i]
		corners := [8][3]float64{
			{x0, y0, z0},
			{x1, y0, z0},
			{x1, y1, z0},
			{x0, y1, z0},
			{x0, y0, z1},
			{x1, y0, z1},
			{x1, y1, z1},
			{x0, y1, z1},
		}
		for _, corner := range corners {
			if !isFinite3D(corner[0], corner[1], corner[2]) {
				continue
			}
			_, zDepth := a.projectPointDepth(corner[0], corner[1], corner[2])
			if zDepth < depth {
				depth = zDepth
			}
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) projectPointDepthWithProjectionLimits(x, y, z float64, mins, maxs vec3) float64 {
	state := a.projectionState()
	if a.distance <= 0 {
		return z
	}
	m := default3DProjectionMatrix(a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs, state)
	_, _, tz := transform3DPoint(m, x, y, z)
	return tz
}

func (a *Axes3D) draw3DTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *core.DrawContext, mins, maxs, tickMins, tickMaxs vec3) {
	fontSize := ctx.RC.TickLabelSize("x")
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	centers, deltas := axes3DLabelCentersDeltas(ctx, mins, maxs)
	labelDeltas := vec3{}
	for i := range 3 {
		labelDeltas[i] = (3.5 + 8) * deltas[i]
	}
	axisLines := a.axisLineEdgePointPairs(mins, maxs, tickMins, tickMaxs)
	tickDirs := [3]int{1, 0, 0}
	nbins := axes3DTickBins(ctx)

	if a.showXLabels {
		xTicks := frameAxisTicks(tickMins[0], tickMaxs[0], nbins)
		for i, tick := range xTicks {
			pos := axisLines[0][0]
			pos[0] = tick
			pos[tickDirs[0]] = axisLines[0][0][tickDirs[0]]
			anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 0), tickMins, tickMaxs)
			draw3DTextAtAnchorAligned(textRen, r, ctx, format3DTick(tick, i, xTicks), anchor, fontSize, textColor, core.TextVAlignTop)
		}
	}
	if a.showYLabels {
		yTicks := frameAxisTicks(tickMins[1], tickMaxs[1], nbins)
		for i, tick := range yTicks {
			pos := axisLines[1][0]
			pos[1] = tick
			pos[tickDirs[1]] = axisLines[1][0][tickDirs[1]]
			anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 1), tickMins, tickMaxs)
			draw3DTextAtAnchorAligned(textRen, r, ctx, format3DTick(tick, i, yTicks), anchor, fontSize, textColor, core.TextVAlignTop)
		}
	}
	zTicks := frameAxisTicks(tickMins[2], tickMaxs[2], nbins)
	if a.showZLabels {
		for i, tick := range zTicks {
			pos := axisLines[2][0]
			pos[2] = tick
			pos[tickDirs[2]] = axisLines[2][0][tickDirs[2]]
			anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 2), tickMins, tickMaxs)
			draw3DTextAtAnchorAligned(textRen, r, ctx, format3DTick(tick, i, zTicks), anchor, fontSize, textColor, core.TextVAlignTop)
		}
	}
}

func (a *Axes3D) draw3DAxisLabels(textRen render.TextDrawer, r render.Renderer, ctx *core.DrawContext, mins, maxs vec3) {
	fontSize := ctx.RC.AxisLabelSize()
	textColor := ctx.RC.DefaultAxesLabelColor()
	projMins, projMaxs := a.projectionLimits()
	centers, deltas := axes3DLabelCentersDeltas(ctx, mins, maxs)
	labelPads := [3]float64{a.XLabelPad(), a.YLabelPad(), ctx.RC.Axes.LabelPad}
	labelDeltas := vec3{}
	for i := range 3 {
		labelDeltas[i] = (labelPads[i] + 21) * deltas[i]
	}
	axisLines := a.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)
	if a.XLabel() != "" {
		pos := midpoint3D(axisLines[0][0], axisLines[0][1])
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 0), projMins, projMaxs)
		draw3DTextAtAnchor(textRen, r, ctx, a.XLabel(), anchor, fontSize, textColor, a.XLabelFontKey(ctx))
	}
	if a.YLabel() != "" {
		pos := midpoint3D(axisLines[1][0], axisLines[1][1])
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 1), projMins, projMaxs)
		draw3DTextAtAnchor(textRen, r, ctx, a.YLabel(), anchor, fontSize, textColor, a.YLabelFontKey(ctx))
	}
	if a.zLabel != "" {
		pos := midpoint3D(axisLines[2][0], axisLines[2][1])
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 2), projMins, projMaxs)
		draw3DTextAtAnchor(textRen, r, ctx, a.zLabel, anchor, fontSize, textColor, core.FontKeyWithWeight(ctx.RC.FontKey, ctx.RC.Axes.LabelWeight))
	}
}

func axes3DLabelCentersDeltas(ctx *core.DrawContext, mins, maxs vec3) (vec3, vec3) {
	centers := vec3{}
	deltas := vec3{}
	dpi := 100.0
	clipWidth := 1.0
	clipHeight := 1.0
	if ctx != nil {
		if ctx.RC.DPI > 0 {
			dpi = ctx.RC.DPI
		}
		if ctx.Clip.W() > 0 {
			clipWidth = ctx.Clip.W()
		}
		if ctx.Clip.H() > 0 {
			clipHeight = ctx.Clip.H()
		}
	}
	deltasPerPoint := 48 / (72 * (clipWidth + clipHeight) / dpi)
	// matplotlib uses scale = 1/12 * 24/25 = 0.08 (not 1/12) to match automargin behavior
	const scale = 0.08
	for i := range 3 {
		centers[i] = (mins[i] + maxs[i]) / 2
		deltas[i] = (maxs[i] - mins[i]) * scale * deltasPerPoint
	}
	return centers, deltas
}

func move3DLabelFromCenter(pos, centers, deltas vec3, axis int) vec3 {
	for i := range 3 {
		if i == axis {
			continue
		}
		if pos[i] < centers[i] {
			pos[i] -= deltas[i]
		} else {
			pos[i] += deltas[i]
		}
	}
	return pos
}

func midpoint3D(a, b vec3) vec3 {
	return vec3{(a[0] + b[0]) / 2, (a[1] + b[1]) / 2, (a[2] + b[2]) / 2}
}

func (a *Axes3D) project3DLabelAnchor(ctx *core.DrawContext, pos, projMins, projMaxs vec3) geom.Pt {
	projected := a.project3DPointWithState(pos[0], pos[1], pos[2], projMins, projMaxs)
	return ctx.TransformFor(core.Coords(core.CoordData)).Apply(projected)
}

func draw3DTextAtAnchor(_ render.TextDrawer, r render.Renderer, ctx *core.DrawContext, label string, anchor geom.Pt, fontSize float64, textColor render.Color, fontKey string) {
	if label == "" {
		return
	}
	core.DrawAlignedTextWithFont(r, ctx, anchor, label, fontSize, textColor, core.TextAlignCenter, core.TextVAlignMiddle, fontKey)
}

func draw3DTextAtAnchorAligned(_ render.TextDrawer, r render.Renderer, ctx *core.DrawContext, label string, anchor geom.Pt, fontSize float64, textColor render.Color, vAlign core.TextVerticalAlign) {
	if label == "" {
		return
	}
	core.DrawAlignedText(r, ctx, anchor, label, fontSize, textColor, core.TextAlignCenter, vAlign)
}

// axes3DTickSpace ports matplotlib XAxis.get_tick_space, which every 3D axis
// uses (axis3d.Axis subclasses maxis.XAxis): the estimated number of ticks
// that fit along the axes-box width in points, assuming tick text has at most
// a 3:1 aspect ratio relative to the x tick label size.
func axes3DTickSpace(ctx *core.DrawContext) int {
	if ctx == nil {
		return 9
	}
	dpi := 100.0
	if ctx.RC.DPI > 0 {
		dpi = ctx.RC.DPI
	}
	lengthPt := ctx.Clip.W() / dpi * 72
	size := ctx.RC.TickLabelSize("x") * 3
	if size <= 0 {
		return math.MaxInt32
	}
	return int(math.Floor(lengthPt / size))
}

// axes3DTickBins resolves AutoLocator's nbins='auto' the way
// MaxNLocator._raw_ticks does: clip(axis.get_tick_space(), max(1,
// min_n_ticks-1), 9) with AutoLocator's min_n_ticks=2.
func axes3DTickBins(ctx *core.DrawContext) int {
	n := axes3DTickSpace(ctx)
	if n < 1 {
		n = 1
	}
	if n > 9 {
		n = 9
	}
	return n
}

func frameAxisTicks(minVal, maxVal float64, nbins int) []float64 {
	lo, hi := minVal, maxVal
	if hi < lo {
		lo, hi = hi, lo
	}
	if nbins <= 0 {
		nbins = 9
	}
	ticks := ticker.MaxNLocator{
		N:            nbins,
		Steps:        []float64{1, 2, 2.5, 5, 10},
		RawStepScale: 23.0 / 24.0,
	}.Ticks(lo, hi, nbins)
	if len(ticks) == 0 {
		return nil
	}
	eps := 1e-12 * math.Max(1, math.Abs(hi-lo))
	filtered := ticks[:0]
	for _, tick := range ticks {
		if tick >= lo-eps && tick <= hi+eps {
			filtered = append(filtered, tick)
		}
	}
	return filtered
}

func format3DTick(v float64, i int, ticks []float64) string {
	step := 0.0
	if len(ticks) > 1 {
		if i+1 < len(ticks) {
			step = math.Abs(ticks[i+1] - ticks[i])
		} else {
			step = math.Abs(ticks[i] - ticks[i-1])
		}
	}
	return strings.ReplaceAll((ticker.ScalarFormatter{Prec: 3}).FormatStep(v, step), "-", "\u2212")
}

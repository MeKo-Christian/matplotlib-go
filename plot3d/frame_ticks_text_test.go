package plot3d

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestAxes3DSetZLabelRenders3DLabel(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetXLabel("X")
	ax.SetYLabel("Y")
	ax.SetZLabel("Z")
	if got := ax.ZLabel(); got != "Z" {
		t.Fatalf("ZLabel = %q, want %q", got, "Z")
	}

	mins, maxs := ax.projectionLimits()
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}
	ax.draw3DAxisLabels(r, r, ctx, mins, maxs)

	if !containsString(r.texts, "Z") {
		t.Fatalf("expected z-axis label text in draw calls, got %v", r.texts)
	}
}

func TestAxes3DFrameSegmentsUseMatplotlibActiveGridPlanes(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	segments := ax.frameSegments(vec3{0, 0, 0}, vec3{1, 1, 1})
	want := []Pt{
		ax.project3DPointWithState(0.2, 0, 0, vec3{0, 0, 0}, vec3{1, 1, 1}),
		ax.project3DPointWithState(0.2, 1, 0, vec3{0, 0, 0}, vec3{1, 1, 1}),
		ax.project3DPointWithState(0.2, 1, 1, vec3{0, 0, 0}, vec3{1, 1, 1}),
	}
	if !contains3DSegment(segments, want, 1e-12) {
		t.Fatalf("missing Matplotlib-style x gridline through active panes; want %+v in %+v", want, segments)
	}
}

func TestAxes3DFrameSegmentsDoNotAddSeparateCubeBox(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-0.05, -0.05, -0.1}
	maxs := vec3{1.05, 1.05, 2.1}
	frameMins, frameMaxs := mins, maxs
	segments := ax.frameSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs, 9)
	gridSegments := ax.frameGridSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs, 9)
	if got, want := len(segments), len(gridSegments); got != want {
		t.Fatalf("frame segment count = %d, want grid-only Matplotlib frame count %d", got, want)
	}
}

func TestAxes3DFrameGridSegmentsUseAxisTickLocations(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-3.45, -3.45, -0.88}
	maxs := vec3{3.45, 3.45, 0.78}
	segments := ax.frameSegments(mins, maxs)
	highs := ax.activePaneHighs(mins, maxs)
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

	p0 := minmax
	p1 := minmax
	p2 := minmax
	p0[0], p1[0], p2[0] = -3, -3, -3
	p0[1] = maxmin[1]
	p2[2] = maxmin[2]
	want := []Pt{
		project3DPointWithLimits(p0[0], p0[1], p0[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(p1[0], p1[1], p1[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(p2[0], p2[1], p2[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
	}
	if !contains3DSegment(segments, want, 1e-12) {
		t.Fatalf("missing gridline at Matplotlib AutoLocator tick -3; want %+v in %+v", want, segments)
	}
}

func TestAxes3DFrameAxisTicksMatchMatplotlibDensity(t *testing.T) {
	ticks := frameAxisTicks(-0.1, 2.1, 9)
	if !containsFloat64Approx(ticks, 0.25, 1e-12) {
		t.Fatalf("3D frame ticks = %v, want Matplotlib-like 0.25 z tick", ticks)
	}
}

func TestAxes3DFrameAxisTicksHandleInvertedLimitsLikeMatplotlib(t *testing.T) {
	ticks := frameAxisTicks(1, 0, 9)
	if !containsFloat64Approx(ticks, 0, 1e-12) ||
		!containsFloat64Approx(ticks, 0.2, 1e-12) ||
		!containsFloat64Approx(ticks, 1, 1e-12) {
		t.Fatalf("3D inverted axis ticks = %v, want ascending Matplotlib tick locations within 0..1", ticks)
	}
}

func TestAxes3DAxisLineSegmentsUseMatplotlibCameraFacingEdges(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-0.05, -0.05, -0.1}
	maxs := vec3{1.05, 1.05, 2.1}
	frameMins, frameMaxs := mins, maxs
	segments := ax.axisLineSegmentsProjected(frameMins, frameMaxs, mins, maxs)
	if got, want := len(segments), 3; got != want {
		t.Fatalf("axis line count = %d, want %d", got, want)
	}

	highs := ax.activePaneHighsProjected(frameMins, frameMaxs, mins, maxs)
	minmax := vec3{}
	maxmin := vec3{}
	for i := range 3 {
		if highs[i] {
			minmax[i] = frameMaxs[i]
			maxmin[i] = frameMins[i]
		} else {
			minmax[i] = frameMins[i]
			maxmin[i] = frameMaxs[i]
		}
	}
	x0 := minmax
	x0[1] = maxmin[1]
	x1 := x0
	x1[0] = maxmin[0]
	wantX := []Pt{
		project3DPointWithLimits(x0[0], x0[1], x0[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(x1[0], x1[1], x1[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
	}
	if !pointsEqual(segments[0], wantX, 1e-12) {
		t.Fatalf("x axis line = %+v, want Matplotlib camera-facing edge %+v", segments[0], wantX)
	}
}

func TestAxes3DTickSegmentsUseMatplotlibInwardOutwardFactors(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-0.05, -0.05, -0.1}
	maxs := vec3{1.05, 1.05, 2.1}
	frameMins, frameMaxs := mins, maxs
	segments := ax.axisTickSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs, 9)
	if len(segments) == 0 {
		t.Fatal("axis tick segments are empty")
	}

	pair := ax.axisLineEdgePointPairs(frameMins, frameMaxs, mins, maxs)[0]
	tick := frameAxisTicks(mins[0], maxs[0], 9)[0]
	tickDir := 1
	tickDelta := (maxs[tickDir] - mins[tickDir]) / 12
	if !ax.activePaneHighsProjected(frameMins, frameMaxs, mins, maxs)[tickDir] {
		tickDelta = -tickDelta
	}
	p0 := pair[0]
	p1 := pair[0]
	p0[0] = tick
	p1[0] = tick
	p0[tickDir] = pair[0][tickDir] + 0.1*tickDelta
	p1[tickDir] = pair[0][tickDir] - 0.2*tickDelta
	want := []Pt{
		project3DPointWithLimits(p0[0], p0[1], p0[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(p1[0], p1[1], p1[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
	}
	if !pointsEqual(segments[0], want, 1e-12) {
		t.Fatalf("first x tick segment = %+v, want Matplotlib inward/outward segment %+v", segments[0], want)
	}
}

func TestAxes3DPanePolygonsUseMatplotlibActivePanes(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{0, 0, 0}
	maxs := vec3{1, 1, 1}
	panes := ax.activePanePolygons(mins, maxs)
	if got, want := len(panes), 3; got != want {
		t.Fatalf("pane count = %d, want %d", got, want)
	}
	highs := ax.activePaneHighs(mins, maxs)
	expectedPlanes := [6][4][3]int{
		{{0, 0, 0}, {0, 1, 0}, {0, 1, 1}, {0, 0, 1}},
		{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}},
		{{0, 1, 0}, {1, 1, 0}, {1, 1, 1}, {0, 1, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}},
		{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}},
	}
	for axis := range 3 {
		planeIndex := 2 * axis
		if highs[axis] {
			planeIndex++
		}
		want := projectPlaneCorners(ax, expectedPlanes[planeIndex], mins, maxs)
		if !pointsEqual(panes[axis], want, 1e-12) {
			t.Fatalf("pane %d = %+v, want active Matplotlib pane %+v", axis, panes[axis], want)
		}
	}
}

func TestAxes3DPaneFaceColorsMatchMatplotlibDefaults(t *testing.T) {
	got := axes3DPaneFaceColors()
	want := []render.Color{
		{R: 0.95, G: 0.95, B: 0.95, A: 0.5},
		{R: 0.90, G: 0.90, B: 0.90, A: 0.5},
		{R: 0.925, G: 0.925, B: 0.925, A: 0.5},
	}
	if len(got) != len(want) {
		t.Fatalf("pane face colors = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pane face color %d = %+v, want Matplotlib default %+v", i, got[i], want[i])
		}
	}
}

func TestAxes3DTickCountAdaptsToAxesWidthLikeMatplotlib(t *testing.T) {
	// matplotlib's 3D axes all inherit XAxis, so AutoLocator's nbins='auto'
	// resolves via XAxis.get_tick_space: floor(axes_width_pt / (3 * xtick
	// labelsize)), clipped to [max(1, min_n_ticks-1), 9] in
	// MaxNLocator._raw_ticks.
	narrow := &DrawContext{
		RC:   style.RC{DPI: 100, XTickLabelFontSize: 10},
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 200, Y: 300}},
	}
	// 200 px @ 100 dpi = 144 pt; 144 / 30 = 4.8 -> 4 bins.
	if got, want := axes3DTickBins(narrow), 4; got != want {
		t.Fatalf("narrow axes tick bins = %d, want %d", got, want)
	}
	wide := &DrawContext{
		RC:   style.RC{DPI: 100, XTickLabelFontSize: 10},
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 600, Y: 300}},
	}
	// 600 px @ 100 dpi = 432 pt; 432 / 30 = 14.4 -> clipped to 9 bins.
	if got, want := axes3DTickBins(wide), 9; got != want {
		t.Fatalf("wide axes tick bins = %d, want %d", got, want)
	}

	// nbins=4 over [-30, 30]: raw step 60/4 * 23/24 = 14.375 -> step 20 ->
	// visible ticks -20, 0, 20 (matplotlib gallery-size 3D panels).
	ticks := frameAxisTicks(-30, 30, 4)
	want := []float64{-20, 0, 20}
	if len(ticks) != len(want) {
		t.Fatalf("frameAxisTicks(-30, 30, 4) = %v, want %v", ticks, want)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-9 {
			t.Fatalf("frameAxisTicks(-30, 30, 4) = %v, want %v", ticks, want)
		}
	}
}

func TestFormat3DTickUsesUnicodeMinusLikeMatplotlib(t *testing.T) {
	ticks := []float64{-0.5, -0.25, 0}
	if got := format3DTick(-0.25, 1, ticks); got != "\u22120.25" {
		t.Fatalf("3D negative tick label = %q, want Unicode minus like Matplotlib", got)
	}
}

func TestAxes3DDrawsYAxisEndpointTickLabels(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	mins, maxs := ax.projectionLimits()
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}

	ax.draw3DTickLabels(r, r, ctx, mins, maxs, mins, maxs)

	xTicks := frameAxisTicks(mins[0], maxs[0], 9)
	yTicks := frameAxisTicks(mins[1], maxs[1], 9)
	zTicks := frameAxisTicks(mins[2], maxs[2], 9)
	if got, want := len(r.texts), len(xTicks)+len(yTicks)+len(zTicks); got != want {
		t.Fatalf("3D tick label count = %d, want x+y+z endpoint labels included (%d)", got, want)
	}
}

func TestAxes3DTickLabelsRespectVisibilityToggles(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)
	ax.SetZLim(0, 1)
	mins, maxs := ax.projectionLimits()
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	drawLabels := func() []string {
		r := &axes3DTextRecorder{}
		ax.draw3DTickLabels(r, r, ctx, mins, maxs, mins, maxs)
		return append([]string(nil), r.texts...)
	}
	all := drawLabels()
	if countString(all, "0.0") != 3 {
		t.Fatalf("visible 3D tick labels = %v, want one 0.0 label on x, y, and z", all)
	}

	ax.SetShowXTickLabels(false)
	xHidden := drawLabels()
	if countString(xHidden, "0.0") != 2 {
		t.Fatalf("x-hidden 3D tick labels = %v, want x-axis labels removed", xHidden)
	}
	ax.SetShowYTickLabels(false)
	yHidden := drawLabels()
	if countString(yHidden, "0.0") != 1 {
		t.Fatalf("x/y-hidden 3D tick labels = %v, want only z-axis labels", yHidden)
	}
	ax.SetShowZTickLabels(false)
	zHidden := drawLabels()
	if len(zHidden) != 0 {
		t.Fatalf("all hidden 3D tick labels = %v, want none", zHidden)
	}
}

func TestAxes3DTickLabelsDrawForInvertedExplicitLimits(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(1, 0)
	ax.SetYLim(0, 1)
	ax.SetZLim(0, 1)
	mins, maxs := ax.projectionLimits()
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}

	ax.draw3DTickLabels(r, r, ctx, mins, maxs, mins, maxs)

	if !containsString(r.texts, "0.2") || !containsString(r.texts, "1.0") {
		t.Fatalf("inverted x tick labels = %v, want Matplotlib-style labels from numeric 0..1 range", r.texts)
	}
}

func TestAxes3DFrameTextDrawsBeforeDataCollectionsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(420, 320)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
	)

	r := &axes3DDrawOrderRecorder{}
	DrawFigure(fig, r)

	textAt, dataAt := -1, -1
	for i, event := range r.events {
		if event == "text" && textAt < 0 {
			textAt = i
		}
		if event == "data" && dataAt < 0 {
			dataAt = i
		}
	}
	if textAt < 0 || dataAt < 0 {
		t.Fatalf("draw events = %v, want both 3D frame text and data collection draw events", r.events)
	}
	if textAt >= dataAt {
		t.Fatalf("draw events = %v, want 3D axis/tick text before data collections like Matplotlib Axes3D.draw", r.events)
	}
}

func TestAxes3DFrameUsesRCLineWidthsLikeMatplotlib(t *testing.T) {
	gridWidth := 2.2
	axisWidth := 1.7
	fig := NewFigure(
		420, 320,
		style.WithGridLineWidths(gridWidth, gridWidth),
		style.WithAxisLineWidth(axisWidth),
	)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
	)

	r := &axes3DLineWidthRecorder{}
	DrawFigure(fig, r)

	if !containsFloat64(r.widths, pointsToPixels(fig.RC, gridWidth)) {
		t.Fatalf("3D frame stroke widths = %v, want grid linewidth from RC %.3g", r.widths, gridWidth)
	}
	if !containsFloat64(r.widths, pointsToPixels(fig.RC, axisWidth)) {
		t.Fatalf("3D frame stroke widths = %v, want axis linewidth from RC %.3g", r.widths, axisWidth)
	}
}

func TestAxes3DFrameUsesRCColorsLikeMatplotlib(t *testing.T) {
	gridColor := render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}
	axisColor := render.Color{R: 0.6, G: 0.1, B: 0.2, A: 1}
	fig := NewFigure(
		420, 320,
		style.WithGridColors(gridColor, gridColor),
		style.WithAxesEdgeColor(axisColor),
	)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
	)

	r := &axes3DLineWidthRecorder{}
	DrawFigure(fig, r)

	if !containsColor(r.colors, gridColor) {
		t.Fatalf("3D frame stroke colors = %+v, want grid color from RC %+v", r.colors, gridColor)
	}
	if !containsColor(r.colors, axisColor) {
		t.Fatalf("3D frame stroke colors = %+v, want axes edge color from RC %+v", r.colors, axisColor)
	}
}

func TestAxes3DFrameUsesRCGridDashesLikeMatplotlib(t *testing.T) {
	gridWidth := 2.0
	fig := NewFigure(420, 320, style.WithGridLineWidths(gridWidth, gridWidth))
	fig.RC.GridDashes = []float64{3, 1.5}
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
	)

	r := &axes3DLineWidthRecorder{}
	DrawFigure(fig, r)

	want := scaleGridDashes(fig.RC.GridDashes, gridWidth)
	if !containsDashPattern(r.dashes, want) {
		t.Fatalf("3D frame grid dashes = %v, want Matplotlib grid linestyle dashes %v", r.dashes, want)
	}
}

func TestAxes3DTickLabelsUseMatplotlibDataSpaceOffset(t *testing.T) {
	fig := NewFigure(760, 560)
	ax, err := AddAxes(fig, geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.10},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetTitle("3D Toolkit Scaffold")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetView(30, -60)
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	ax.Scatter3D([]float64{0.5, 0.7}, []float64{0.2, 0.9}, []float64{0.1, 0.3})
	z := [][]float64{{0, 1}, {1, 2}}
	ax.Wireframe([]float64{0, 1}, []float64{0, 1}, z)
	ax.Surface([]float64{0, 1}, []float64{0, 1}, z)

	mins, maxs := ax.projectionLimits()
	frameMins, frameMaxs := mins, maxs
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}

	ax.draw3DTickLabels(r, r, ctx, frameMins, frameMaxs, mins, maxs)

	if len(r.texts) == 0 {
		t.Fatal("expected 3D tick labels to be drawn")
	}
	xTicks := frameAxisTicks(mins[0], maxs[0], 9)
	label := format3DTick(xTicks[0], 0, xTicks)
	if got := r.texts[0]; got != label {
		t.Fatalf("first tick label = %q, want first x tick %q", got, label)
	}
	fontSize := ctx.RC.TickLabelSize("x")
	expectedAnchor := expectedMatplotlib3DTickLabelAnchor(ax, ctx, 0, xTicks[0], frameMins, frameMaxs, mins, maxs)
	layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey)
	want := alignedSingleLineOrigin(expectedAnchor, layout, TextAlignCenter, textLayoutVAlignTop)
	if !approx(r.positions[0].X, want.X, 1e-9) || !approx(r.positions[0].Y, want.Y, 1e-9) {
		t.Fatalf("first x tick label origin = %+v, want Matplotlib top-aligned data-space offset origin %+v", r.positions[0], want)
	}
}

func TestAxes3DTickLabelAnchorsMatchMatplotlibBasicFixture(t *testing.T) {
	fig := NewFigure(760, 560)
	ax, err := AddAxes(fig, geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.14},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetView(30, -60)
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	ax.Scatter3D([]float64{0.5, 0.7}, []float64{0.2, 0.9}, []float64{0.1, 0.3})
	z := [][]float64{{0, 1}, {1, 2}}
	ax.Wireframe([]float64{0, 1}, []float64{0, 1}, z)
	ax.Surface([]float64{0, 1}, []float64{0, 1}, z)
	ax.Contour([]float64{0, 1}, []float64{0, 1}, z)
	ax.Bar3D([]float64{0.2}, []float64{0.3}, []float64{0.4}, []float64{0.2}, []float64{0.2}, []float64{0.3})
	ax.Text3D(0.2, 0.8, 0.6, "demo point")

	mins, maxs := ax.projectionLimits()
	frameMins, frameMaxs := mins, maxs
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}

	ax.draw3DTickLabels(r, r, ctx, frameMins, frameMaxs, mins, maxs)

	fontSize := ctx.RC.TickLabelSize("x")
	for _, tc := range []struct {
		label      string
		occurrence int
		want       geom.Pt
	}{
		// Display space is y-up: Y is the flip of Matplotlib's top-down pixel (H=560).
		{label: "0.0", occurrence: 1, want: geom.Pt{X: 207.736, Y: 154.162}},
		{label: "0.0", occurrence: 2, want: geom.Pt{X: 463.314, Y: 94.691}},
		{label: "0.00", occurrence: 1, want: geom.Pt{X: 586.431, Y: 245.513}},
	} {
		idx := -1
		count := 0
		for i, label := range r.texts {
			if label == tc.label {
				count++
				if count == tc.occurrence {
					idx = i
					break
				}
			}
		}
		if idx == -1 {
			t.Fatalf("tick label %q was not drawn; got labels %v", tc.label, r.texts)
		}
		layout := measureSingleLineTextLayout(r, tc.label, fontSize, ctx.RC.FontKey)
		got := geom.Pt{
			X: r.positions[idx].X + textHorizontalOriginOffset(layout, TextAlignCenter),
			Y: r.positions[idx].Y - textBaselineOffset(layout, textLayoutVAlignTop),
		}
		if !approx(got.X, tc.want.X, 0.01) || !approx(got.Y, tc.want.Y, 0.01) {
			t.Errorf("%q occurrence %d tick label anchor = %+v, want Matplotlib %+v", tc.label, tc.occurrence, got, tc.want)
		}
	}
}

func TestAxes3DText3DProjectsInput(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetDistance(0)
	ax.SetView(0, 0)
	text := ax.Text3D(1, 2, 3, "hello")
	if text == nil || text.Content != "hello" {
		t.Fatalf("Text3D returned unexpected value: %#v", text)
	}
	if !approx(text.Position.X, 1, 1e-12) || !approx(text.Position.Y, 2, 1e-12) {
		t.Fatalf("Text position = %+v, want {1 2}", text.Position)
	}
}

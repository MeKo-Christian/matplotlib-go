package core

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestAxes_TickLabelPositionHelpers(t *testing.T) {
	axes := &Axes{XAxis: NewXAxis(), YAxis: NewYAxis()}

	if err := axes.SetXTickLabelPosition("top"); err != nil {
		t.Fatalf("SetXTickLabelPosition(top): %v", err)
	}
	if axes.XAxis.ShowLabels {
		t.Fatal("bottom x-axis labels should be hidden when top labels are requested")
	}
	if axes.XAxisTop == nil || !axes.XAxisTop.ShowLabels {
		t.Fatal("top x-axis labels should be visible after SetXTickLabelPosition(top)")
	}

	if err := axes.SetYTickLabelPosition("both"); err != nil {
		t.Fatalf("SetYTickLabelPosition(both): %v", err)
	}
	if !axes.YAxis.ShowLabels || axes.YAxisRight == nil || !axes.YAxisRight.ShowLabels {
		t.Fatal("both y-axis labels should be visible after SetYTickLabelPosition(both)")
	}
}

func TestAxes_LabelPositionHelpers(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	if err := ax.SetXLabelPosition("top"); err != nil {
		t.Fatalf("SetXLabelPosition(top): %v", err)
	}
	if err := ax.SetYLabelPosition("right"); err != nil {
		t.Fatalf("SetYLabelPosition(right): %v", err)
	}

	if ax.effectiveXLabelSide() != AxisTop {
		t.Fatalf("effective x label side = %v, want %v", ax.effectiveXLabelSide(), AxisTop)
	}
	if ax.effectiveYLabelSide() != AxisRight {
		t.Fatalf("effective y label side = %v, want %v", ax.effectiveYLabelSide(), AxisRight)
	}
}

func TestAxes_TwinAxes(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	twinX := ax.TwinX()
	if twinX == nil {
		t.Fatal("TwinX() returned nil")
	}
	if twinX.shareX != ax.xScaleRoot() {
		t.Fatal("TwinX() should share the x-scale root")
	}
	if twinX.YAxisRight == nil {
		t.Fatal("TwinX() should expose a right-side y-axis")
	}
	if twinX.XAxis.ShowTicks || twinX.XAxis.ShowLabels {
		t.Fatal("TwinX() should hide the duplicate primary x-axis")
	}
	if !twinX.ShowFrame || !twinX.XAxis.ShowSpine || !twinX.YAxis.ShowSpine {
		t.Fatal("TwinX() should keep the foreground frame spines visible like Matplotlib")
	}

	twinY := ax.TwinY()
	if twinY == nil {
		t.Fatal("TwinY() returned nil")
	}
	if twinY.shareY != ax.yScaleRoot() {
		t.Fatal("TwinY() should share the y-scale root")
	}
	if twinY.XAxisTop == nil {
		t.Fatal("TwinY() should expose a top-side x-axis")
	}
	if twinY.YAxis.ShowTicks || twinY.YAxis.ShowLabels {
		t.Fatal("TwinY() should hide the duplicate primary y-axis")
	}
	if !twinY.ShowFrame || !twinY.XAxis.ShowSpine || !twinY.YAxis.ShowSpine {
		t.Fatal("TwinY() should keep the foreground frame spines visible like Matplotlib")
	}
}

func TestAxes_SecondaryAxesUseLinkedScale(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 10)

	secondaryX, err := ax.SecondaryXAxis(
		AxisTop,
		func(v float64) float64 { return v*1.8 + 32 },
		func(v float64) (float64, bool) { return (v - 32) / 1.8, true },
	)
	if err != nil {
		t.Fatalf("SecondaryXAxis: %v", err)
	}
	dMin, dMax := secondaryX.XScale.Domain()
	if dMin != 32 || dMax != 212 {
		t.Fatalf("secondary x domain = (%v, %v), want (32, 212)", dMin, dMax)
	}
	if got := secondaryX.XScale.Fwd(122); got != ax.XScale.Fwd(50) {
		t.Fatalf("secondary x forward mapping mismatch: got %v want %v", got, ax.XScale.Fwd(50))
	}
	if secondaryX.XAxisTop == nil {
		t.Fatal("SecondaryXAxis should expose a top x-axis")
	}
	if secondaryX.XAxisTop.ShowSpine {
		t.Fatal("SecondaryXAxis should not draw an overlay spine over the parent frame")
	}

	secondaryY, err := ax.SecondaryYAxis(
		AxisRight,
		func(v float64) float64 { return v * 1000 },
		func(v float64) (float64, bool) { return v / 1000, true },
	)
	if err != nil {
		t.Fatalf("SecondaryYAxis: %v", err)
	}
	dMin, dMax = secondaryY.YScale.Domain()
	if dMin != 0 || dMax != 10000 {
		t.Fatalf("secondary y domain = (%v, %v), want (0, 10000)", dMin, dMax)
	}
	if secondaryY.YAxisRight == nil {
		t.Fatal("SecondaryYAxis should expose a right y-axis")
	}
	if secondaryY.YAxisRight.ShowSpine {
		t.Fatal("SecondaryYAxis should not draw an overlay spine over the parent frame")
	}
}

func TestAxis_DrawTickLabels_UsesStepPrecisionForScalarFormatter(t *testing.T) {
	axis := NewYAxis()
	axis.Locator = LinearLocator{}
	axis.MajorTickCount = 6
	ctx := createTestDrawContext()
	ctx.DataToPixel.YScale = transform.NewLinear(0, 0.196)

	var r axisLabelRecordingRenderer
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	want := []string{"0.00", "0.05", "0.10", "0.15"}
	if len(r.texts) != len(want) {
		t.Fatalf("unexpected tick label count: got %v want %v", r.texts, want)
	}
	for i := range want {
		if r.texts[i] != want[i] {
			t.Fatalf("tick label %d mismatch: got %q want %q", i, r.texts[i], want[i])
		}
	}
}

func TestAxis_DrawTickLabels_OmitsXLabelsOutsideViewLimits(t *testing.T) {
	axis := NewXAxis()
	ctx := createTestDrawContext()
	ctx.DataToPixel.XScale = transform.NewLinear(0.8, 10.8)

	var r axisLabelRecordingRenderer
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	want := []string{"2", "4", "6", "8", "10"}
	if len(r.texts) != len(want) {
		t.Fatalf("unexpected tick label count: got %v want %v", r.texts, want)
	}
	for i := range want {
		if r.texts[i] != want[i] {
			t.Fatalf("tick label %d mismatch: got %q want %q", i, r.texts[i], want[i])
		}
	}
}

func TestAxis_DrawTickLabels_DrawsConciseDateOffsetText(t *testing.T) {
	axis := NewXAxis()
	start := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	ticks := []float64{
		timeToDateNumber(start),
		timeToDateNumber(start.Add(6 * time.Hour)),
		timeToDateNumber(start.Add(12 * time.Hour)),
		timeToDateNumber(start.Add(18 * time.Hour)),
	}
	axis.Locator = FixedLocator{TicksList: ticks}
	axis.Formatter = ConciseDateFormatter{Location: time.UTC}

	ctx := createTestDrawContext()
	ctx.DataToPixel.XScale = transform.NewLinear(ticks[0], ticks[len(ticks)-1])

	var r axisLabelRecordingRenderer
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	want := []string{"Jan-02", "06:00", "12:00", "18:00", "2024-Jan-02"}
	if fmt.Sprint(r.texts) != fmt.Sprint(want) {
		t.Fatalf("drawn concise labels = %v, want %v", r.texts, want)
	}
	if len(r.origins) != len(want) {
		t.Fatalf("drawn origins = %d, want %d", len(r.origins), len(want))
	}
	offset := r.origins[len(r.origins)-1]
	first := r.origins[0]
	if !(offset.X > first.X && offset.Y < first.Y) {
		t.Fatalf("offset origin = %+v, want bottom-right below labels starting at %+v", offset, first)
	}
}

func TestAxis_DrawTickLabels_PositionsOffsetTextFromTickLabelBounds(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{0, 10}
	axis.Formatter = offsetFormatterForTest{
		labels: []string{"left", "right"},
		offset: "offset",
	}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 100
	axis.TickSize = pointsToPixels(ctx.RC, defaultTickSizePt)
	bottom := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: getSpinePosition(axis, ctx)}).Y
	ctx.Clip.Min.Y = bottom
	ctx.Clip.Max.X = 500

	var r axisLabelRecordingRenderer
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.texts) != 3 {
		t.Fatalf("drawn labels = %v, want two tick labels plus offset", r.texts)
	}

	fontSize := tickLabelFontSize(axis, ctx)
	labelPad := tickLabelPadPx(axis, ctx)
	tickLabelLineHeight := pointsToPixels(ctx.RC, fontSize)
	offsetPad := pointsToPixels(ctx.RC, 3)
	offsetAscent := fontSize * 0.8
	offsetWidth := float64(len("offset")) * fontSize * 0.5
	want := geom.Pt{
		X: ctx.Clip.Max.X - offsetWidth,
		Y: bottom - labelPad - tickLabelLineHeight - offsetPad - offsetAscent,
	}
	got := r.origins[len(r.origins)-1]
	if math.Abs(got.X-want.X) > 1e-9 || math.Abs(got.Y-want.Y) > 1e-9 {
		t.Fatalf("offset origin = %+v, want Matplotlib ticklabel-bounds origin %+v", got, want)
	}
}

func TestTickLabelPositionUsesBoundsForBottomXAxis(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = ScalarFormatter{Prec: 0}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"2": {X: 1, Y: -8, W: 5, H: 10},
	}
	r.useFontHeights = true
	r.fontHeights = render.FontHeightMetrics{Ascent: 8, Descent: 2}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.origins) != 1 {
		t.Fatalf("expected one tick label draw, got %d", len(r.origins))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: 2, Y: getSpinePosition(axis, ctx)})
	labelPad := tickLabelPadPx(axis, ctx)
	// Display space is y-up: the bottom label sits below the tick (anchor
	// tickPos.Y-labelPad) with a Top vAlign baseline offset of -Ascent.
	want := geom.Pt{
		X: tickPos.X - 5.0/2.0,
		Y: tickPos.Y - labelPad - 8,
	}
	if r.origins[0] != want {
		t.Fatalf("bottom x tick origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestTickLabelPadMatchesMatplotlibOutsideTickPadding(t *testing.T) {
	axis := NewXAxis()
	axis.TickSize = 10

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if got, want := tickLabelPadPx(axis, ctx), 13.5; math.Abs(got-want) > 1e-9 {
		t.Fatalf("outward tick label pad = %v, want %v", got, want)
	}

	axis.TickDirection = TickDirectionInOut
	if got, want := tickLabelPadPx(axis, ctx), 8.5; math.Abs(got-want) > 1e-9 {
		t.Fatalf("inout tick label pad = %v, want %v", got, want)
	}

	axis.TickDirection = TickDirectionIn
	if got, want := tickLabelPadPx(axis, ctx), 3.5; math.Abs(got-want) > 1e-9 {
		t.Fatalf("inward tick label pad = %v, want %v", got, want)
	}
}

func TestTickLabelPositionUsesBoundsForLeftYAxis(t *testing.T) {
	axis := NewYAxis()
	axis.Locator = staticLocator{4}
	axis.Formatter = ScalarFormatter{Prec: 0}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"4": {X: 1, Y: -8, W: 5, H: 10},
	}
	r.useFontHeights = true
	r.fontHeights = render.FontHeightMetrics{Ascent: 8, Descent: 2}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.origins) != 1 {
		t.Fatalf("expected one tick label draw, got %d", len(r.origins))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: getSpinePosition(axis, ctx), Y: 4})
	labelPad := tickLabelPadPx(axis, ctx)
	// Display space is y-up: the y-axis label is center-baseline aligned, so the
	// baseline offset is -Ascent/2 = -4.
	want := geom.Pt{
		X: tickPos.X - labelPad - 5.0,
		Y: tickPos.Y - 4,
	}
	if r.origins[0] != want {
		t.Fatalf("left y tick origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestTickLabelPositionUsesFontHeightMetricsForBottomXAxis(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = ScalarFormatter{Prec: 0}

	var r axisLabelRecordingRenderer
	r.useFontHeights = true
	r.fontHeights = render.FontHeightMetrics{Ascent: 8, Descent: 2}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.origins) != 1 {
		t.Fatalf("expected one tick label draw, got %d", len(r.origins))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: 2, Y: getSpinePosition(axis, ctx)})
	labelPad := tickLabelPadPx(axis, ctx)
	layout := measureSingleLineTextLayout(&r, "2", tickLabelFontSize(axis, ctx), ctx.RC.FontKey)
	// Display space is y-up: bottom label below the tick with Top vAlign (-Ascent).
	want := geom.Pt{
		X: tickPos.X - layout.Width/2,
		Y: tickPos.Y - labelPad - 8,
	}
	if r.origins[0] != want {
		t.Fatalf("bottom x tick origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestTickLabelPositionUsesBottomAlignmentForTopXAxis(t *testing.T) {
	axis := NewXAxis()
	axis.Side = AxisTop
	axis.Locator = staticLocator{2}
	axis.Formatter = ScalarFormatter{Prec: 0}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"2": {X: 1, Y: -6, W: 5, H: 8},
	}
	r.useFontHeights = true
	r.fontHeights = render.FontHeightMetrics{Ascent: 8, Descent: 2}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.origins) != 1 {
		t.Fatalf("expected one tick label draw, got %d", len(r.origins))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: 2, Y: getSpinePosition(axis, ctx)})
	labelPad := tickLabelPadPx(axis, ctx)
	// Display space is y-up: the top label sits above the tick (anchor
	// tickPos.Y+labelPad) with a Bottom vAlign baseline offset of +Descent.
	want := geom.Pt{
		X: tickPos.X - 5.0/2.0,
		Y: tickPos.Y + labelPad + 2,
	}
	if r.origins[0] != want {
		t.Fatalf("top x tick origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestTickLabelPositionUsesCenterBaselineForRightYAxis(t *testing.T) {
	axis := NewYAxis()
	axis.Side = AxisRight
	axis.Locator = staticLocator{4}
	axis.Formatter = ScalarFormatter{Prec: 0}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"4": {X: 1, Y: -6, W: 5, H: 8},
	}
	r.useFontHeights = true
	r.fontHeights = render.FontHeightMetrics{Ascent: 8, Descent: 2}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.origins) != 1 {
		t.Fatalf("expected one tick label draw, got %d", len(r.origins))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: getSpinePosition(axis, ctx), Y: 4})
	labelPad := tickLabelPadPx(axis, ctx)
	// Display space is y-up: center-baseline y label baseline offset is -Ascent/2 = -4.
	want := geom.Pt{
		X: tickPos.X + labelPad,
		Y: tickPos.Y - 4,
	}
	if r.origins[0] != want {
		t.Fatalf("right y tick origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestAxis_DrawTickLabels_UsesRotatedDrawerWhenRequested(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = ScalarFormatter{Prec: 0}
	axis.MajorLabelStyle = TickLabelStyle{Rotation: 45, AutoAlign: true}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"2": {X: 1, Y: -8, W: 5, H: 10},
	}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.texts) != 0 {
		t.Fatalf("expected rotated tick labels to bypass DrawText, got %v", r.texts)
	}
	if len(r.rotatedAnchors) != 1 {
		t.Fatalf("expected one rotated tick label draw, got %d", len(r.rotatedAnchors))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: 2, Y: getSpinePosition(axis, ctx)})
	labelPad := tickLabelPadPx(axis, ctx)
	// Display space is y-up: bottom label below the tick with Top vAlign (-Ascent).
	origin := geom.Pt{
		X: tickPos.X - 5.0/2.0,
		Y: tickPos.Y - labelPad - 8,
	}
	layout := measureSingleLineTextLayout(&r, "2", tickLabelFontSize(axis, ctx), ctx.RC.FontKey)
	want := tickLabelRotationAnchor(origin, layout, TextAlignCenter, textLayoutVAlignTop, math.Pi/4)
	if r.rotatedAnchors[0] != want {
		t.Fatalf("rotated tick label anchor = %+v, want %+v", r.rotatedAnchors[0], want)
	}
}

func TestAxis_TickLabelBoundsIncludeRotatedLayout(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = FixedFormatter{Labels: []string{"rotated-label"}}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"rotated-label": {X: 0, Y: -10, W: 80, H: 10},
	}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	unrotated, ok := axisTickLabelBounds(axis, &r, ctx)
	if !ok {
		t.Fatal("expected unrotated tick label bounds")
	}

	axis.MajorLabelStyle = TickLabelStyle{Rotation: 45, AutoAlign: true}
	rotated, ok := axisTickLabelBounds(axis, &r, ctx)
	if !ok {
		t.Fatal("expected rotated tick label bounds")
	}

	// matplotlib's rotation_mode="default" with valign="top" (the AutoAlign
	// value for a bottom x-axis) pins the rotated bounding box's top edge to the
	// anchor and lets the label extend downward and sideways — it never grows
	// above the unrotated top. (Verified against matplotlib: a rotated bottom
	// tick label keeps its top y unchanged and only its bottom y increases.)
	// In this y-up coordinate the pinned top edge is Max.Y and the downward
	// extent is Min.Y.
	if rotated.Max.Y > unrotated.Max.Y+1 {
		t.Fatalf("rotated bounds extended above the pinned top edge: rotated=%+v unrotated=%+v", rotated, unrotated)
	}
	if rotated.Min.Y >= unrotated.Min.Y-10 {
		t.Fatalf("rotated bounds did not extend below the unrotated layout: rotated=%+v unrotated=%+v", rotated, unrotated)
	}
	if (rotated.Max.Y - rotated.Min.Y) <= (unrotated.Max.Y - unrotated.Min.Y) {
		t.Fatalf("rotated bounds were not taller than unrotated: rotated=%+v unrotated=%+v", rotated, unrotated)
	}
}

func TestAxis_DrawTickLabels_RendersFullMathAsPathsWhenRotated(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = FixedFormatter{Labels: []string{`$\\frac{1}{2}$`}}
	axis.MajorLabelStyle = TickLabelStyle{Rotation: 45, AutoAlign: true}

	var r axisLabelRecordingRenderer
	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.texts) != 0 {
		t.Fatalf("expected rotated math tick labels to bypass DrawText, got %v", r.texts)
	}
	if len(r.rotatedText) != 0 {
		t.Fatalf("expected rotated math tick labels to bypass DrawTextRotated, got %v", r.rotatedText)
	}
	if !containsString(r.textPathCalls, "1") || !containsString(r.textPathCalls, "2") {
		t.Fatalf("expected fraction runs to resolve through TextPath, got %v", r.textPathCalls)
	}
	if r.pathCount < 3 {
		t.Fatalf("expected fraction rule plus glyph paths, got %d paths", r.pathCount)
	}
}

func TestAxis_ExtraTickLevelsDrawAdditionalLabels(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = FixedFormatter{Labels: []string{"major"}}
	axis.ClearTickLevels()
	axis.AddTickLevel(TickLevel{
		Locator:    staticLocator{2},
		Formatter:  FixedFormatter{Labels: []string{"minor row"}},
		ShowLabels: true,
		LabelStyle: TickLabelStyle{Pad: 14, AutoAlign: true},
	})

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"major":     {X: 1, Y: -8, W: 20, H: 10},
		"minor row": {X: 1, Y: -8, W: 35, H: 10},
	}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.texts) != 2 {
		t.Fatalf("expected major and extra tick labels, got %v", r.texts)
	}
	if r.texts[0] != "major" || r.texts[1] != "minor row" {
		t.Fatalf("unexpected tick label sequence: %v", r.texts)
	}
	// Display space is y-up: a bottom-axis extra level farther below sits at a
	// smaller Y than the major level.
	if !(r.origins[1].Y < r.origins[0].Y) {
		t.Fatalf("expected extra tick level to be farther from the axis: major=%+v extra=%+v", r.origins[0], r.origins[1])
	}
}

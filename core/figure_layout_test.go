package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestDrawFigure_RendersFigureLevelLabels(t *testing.T) {
	fig := NewFigure(800, 600)
	fig.SetSuptitle("Overview")
	fig.SetSupXLabel("time")
	fig.SetSupYLabel("value")

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.92, Y: 0.82},
	})
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowLabels = false

	var r figureLayoutRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "Overview") {
		t.Fatalf("missing suptitle draw: %v", r.texts)
	}
	titleOrigin, ok := r.textOrigin("Overview")
	if !ok {
		t.Fatalf("missing suptitle origin: %v", r.texts)
	}
	ctx := newFigureDrawContext(fig, fig.DisplayRect())
	titleLayout := measureSingleLineTextLayout(&r, "Overview", titleFontSize(ctx), fig.RC.FontKey, fig.RC.UseTeX)
	// Display space is y-up: the suptitle sits near the top edge (Max.Y), its
	// baseline dropping below the top inset by the ascent.
	wantY := fig.DisplayRect().Max.Y - figureLabelTopInsetPx(fig, ctx) - titleLayout.Ascent
	if !floatApprox(titleOrigin.Y, wantY, 1e-12) {
		t.Fatalf("suptitle origin Y = %v, want Matplotlib default y=0.98 position %v", titleOrigin.Y, wantY)
	}
	if !containsString(r.texts, "time") {
		t.Fatalf("missing supxlabel draw: %v", r.texts)
	}
	xlabelOrigin, ok := r.textOrigin("time")
	if !ok {
		t.Fatalf("missing supxlabel origin: %v", r.texts)
	}
	xlabelLayout := measureSingleLineTextLayout(&r, "time", figureLabelFontSize(ctx), fig.RC.FontKey, fig.RC.UseTeX)
	// Display space is y-up: the supxlabel sits near the bottom edge (Min.Y), its
	// baseline rising above the bottom inset by the descent.
	wantXLabelY := fig.DisplayRect().Min.Y + figureLabelBottomInsetPx(fig, ctx) + xlabelLayout.Descent
	if !floatApprox(xlabelOrigin.Y, wantXLabelY, 1e-12) {
		t.Fatalf("supxlabel origin Y = %v, want Matplotlib default y=0.01 position %v", xlabelOrigin.Y, wantXLabelY)
	}
	if !containsString(r.rotatedText, "value") {
		t.Fatalf("missing supylabel draw: %v", r.rotatedText)
	}
	anchor, ok := r.rotatedAnchor("value")
	if !ok {
		t.Fatalf("missing supylabel anchor: %v", r.rotatedText)
	}
	ylabelLayout := measureSingleLineTextLayout(&r, "value", figureLabelFontSize(ctx), fig.RC.FontKey, fig.RC.UseTeX)
	wantYLabelAnchor := geom.Pt{
		X: fig.DisplayRect().Min.X + figureLabelLeftInsetPx(fig, ctx) + ylabelLayout.Height,
		Y: fig.DisplayRect().Min.Y + fig.DisplayRect().H()/2,
	}
	if !floatApprox(anchor.X, wantYLabelAnchor.X, 1e-12) || !floatApprox(anchor.Y, wantYLabelAnchor.Y, 1e-12) {
		t.Fatalf("supylabel anchor = %+v, want Matplotlib default x=0.02/y=0.5 anchor %+v", anchor, wantYLabelAnchor)
	}
}

func TestFigureLegendCollectsAcrossAxes(t *testing.T) {
	fig := NewFigure(800, 600)
	left := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.15},
		Max: geom.Pt{X: 0.45, Y: 0.85},
	})
	right := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.55, Y: 0.15},
		Max: geom.Pt{X: 0.90, Y: 0.85},
	})

	left.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	right.Scatter([]float64{0.5}, []float64{0.5}, ScatterOptions{Label: "samples"})
	fig.AddLegend()

	var r figureLayoutRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "signal") || !containsString(r.texts, "samples") {
		t.Fatalf("unexpected figure legend labels: %v", r.texts)
	}
	if r.pathCount < 4 {
		t.Fatalf("expected figure legend box and samples, got %d paths", r.pathCount)
	}
}

func TestFigureLegendStacksBelowSuptitle(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.15},
		Max: geom.Pt{X: 0.90, Y: 0.85},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	fig.SetSuptitle("Figure Title")
	fig.AddLegend()

	var r figureLayoutRecordingRenderer
	DrawFigure(fig, &r)

	title, ok := r.textOrigin("Figure Title")
	if !ok {
		t.Fatalf("missing suptitle draw, texts=%v", r.texts)
	}
	label, ok := r.textOrigin("signal")
	if !ok {
		t.Fatalf("missing figure legend label draw, texts=%v", r.texts)
	}
	// Display space is y-up: the legend stacks below the suptitle at a smaller Y.
	if label.Y >= title.Y {
		t.Fatalf("figure legend should stack below suptitle, got title=%+v legend=%+v", title, label)
	}
}

func TestAnchoredTextBoxDrawsFigureAndAxesBoxes(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.90, Y: 0.88},
	})
	ax.AddAnchoredText("Peak\nstable")
	fig.AddAnchoredText("Global", AnchoredTextOptions{Location: LegendLowerRight})

	var r figureLayoutRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "Peak") || !containsString(r.texts, "stable") || !containsString(r.texts, "Global") {
		t.Fatalf("unexpected anchored text draws: %v", r.texts)
	}
	if r.pathCount < 2 {
		t.Fatalf("expected anchored text boxes to draw borders, got %d paths", r.pathCount)
	}
}

func TestDrawFigure_AlignsYLabelsAcrossColumn(t *testing.T) {
	fig := NewFigure(800, 600)
	top := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.55},
		Max: geom.Pt{X: 0.90, Y: 0.90},
	})
	bottom := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.10},
		Max: geom.Pt{X: 0.90, Y: 0.45},
	})

	top.SetYLabel("Top Y")
	bottom.SetYLabel("Bottom Y")
	top.YAxis.Locator = staticLocator{1000}
	top.YAxis.Formatter = ScalarFormatter{Prec: 0}
	bottom.YAxis.Locator = staticLocator{1}
	bottom.YAxis.Formatter = ScalarFormatter{Prec: 0}

	r := figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"1000": {X: 1, Y: -8, W: 24, H: 10},
			"1":    {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	DrawFigure(fig, &r)

	topAnchor, okTop := r.rotatedAnchor("Top Y")
	bottomAnchor, okBottom := r.rotatedAnchor("Bottom Y")
	if !okTop || !okBottom {
		t.Fatalf("missing ylabel draws: %v", r.rotatedText)
	}
	if topAnchor.X != bottomAnchor.X {
		t.Fatalf("expected aligned ylabel anchors, got top=%+v bottom=%+v", topAnchor, bottomAnchor)
	}
}

func TestDrawFigure_AlignsXLabelsAcrossRow(t *testing.T) {
	fig := NewFigure(800, 600)
	left := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.18},
		Max: geom.Pt{X: 0.45, Y: 0.82},
	})
	right := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.55, Y: 0.18},
		Max: geom.Pt{X: 0.90, Y: 0.82},
	})

	left.SetXLabel("Left X")
	right.SetXLabel("Right X")
	left.XAxis.Locator = staticLocator{1000}
	left.XAxis.Formatter = ScalarFormatter{Prec: 0}
	right.XAxis.Locator = staticLocator{1}
	right.XAxis.Formatter = ScalarFormatter{Prec: 0}

	r := figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"1000": {X: 1, Y: -14, W: 24, H: 20},
			"1":    {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	DrawFigure(fig, &r)

	leftOrigin, okLeft := r.textOrigin("Left X")
	rightOrigin, okRight := r.textOrigin("Right X")
	if !okLeft || !okRight {
		t.Fatalf("missing xlabel draws: %v", r.texts)
	}
	if leftOrigin.Y != rightOrigin.Y {
		t.Fatalf("expected aligned xlabel origins, got left=%+v right=%+v", leftOrigin, rightOrigin)
	}
}

func TestDrawFigure_AddAxesTitlesAreNotAlignedByDefault(t *testing.T) {
	fig := NewFigure(800, 600)
	left := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.18},
		Max: geom.Pt{X: 0.45, Y: 0.82},
	})
	right := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.55, Y: 0.18},
		Max: geom.Pt{X: 0.90, Y: 0.82},
	})

	left.SetTitle("Left Title")
	right.SetTitle("Right Title")
	if err := left.SetXTickLabelPosition("top"); err != nil {
		t.Fatalf("left SetXTickLabelPosition(top): %v", err)
	}
	if err := right.SetXTickLabelPosition("top"); err != nil {
		t.Fatalf("right SetXTickLabelPosition(top): %v", err)
	}
	left.TopAxis().Locator = staticLocator{1000}
	left.TopAxis().Formatter = ScalarFormatter{Prec: 0}
	right.TopAxis().Locator = staticLocator{1}
	right.TopAxis().Formatter = ScalarFormatter{Prec: 0}

	r := figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"1000": {X: 1, Y: -14, W: 24, H: 20},
			"1":    {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	DrawFigure(fig, &r)

	leftOrigin, okLeft := r.textOrigin("Left Title")
	rightOrigin, okRight := r.textOrigin("Right Title")
	if !okLeft || !okRight {
		t.Fatalf("missing title draws: %v", r.texts)
	}
	if leftOrigin.Y == rightOrigin.Y {
		t.Fatalf("add_axes titles should use local top extents by default, got left=%+v right=%+v", leftOrigin, rightOrigin)
	}
}

func TestDrawFigure_TitleClearsSecondaryXAxisTickLabels(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.20, Y: 0.20},
		Max: geom.Pt{X: 0.80, Y: 0.75},
	})
	ax.SetTitle("Parent Title")
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)

	sec, err := ax.SecondaryXAxis(
		AxisTop,
		func(x float64) float64 { return x * 100 },
		func(x float64) (float64, bool) { return x / 100, true },
	)
	if err != nil {
		t.Fatalf("SecondaryXAxis: %v", err)
	}
	if axis := sec.TopAxis(); axis != nil {
		axis.Locator = staticLocator{100}
		axis.Formatter = ScalarFormatter{Prec: 0}
	}

	r := figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"100": {X: 1, Y: -8, W: 18, H: 10},
		},
	}
	DrawFigure(fig, &r)

	titleOrigin, ok := r.textOrigin("Parent Title")
	if !ok {
		t.Fatalf("missing title draw: %v", r.texts)
	}
	// Display space is y-up: the title clears the secondary top labels by sitting
	// near the top edge (high Y), i.e. above 245 on the 300px figure.
	if titleOrigin.Y <= 245 {
		t.Fatalf("title did not clear secondary top tick labels: origin=%+v", titleOrigin)
	}
}

func TestDrawFigure_TightLayoutRespondsToMeasuredTextExtents(t *testing.T) {
	small := newTightLayoutProbeFigure()
	small.Children[0].SetXLabel("")
	small.TightLayout()
	DrawFigure(small, &figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"edge": {X: 0, Y: -8, W: 18, H: 10},
		},
	})

	large := newTightLayoutProbeFigure()
	large.Children[0].SetXLabel("")
	large.TightLayout()
	DrawFigure(large, &figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"edge": {X: 0, Y: -28, W: 72, H: 40},
		},
	})

	smallAx := small.Children[0]
	largeAx := large.Children[0]
	if largeAx.RectFraction.Min.X <= smallAx.RectFraction.Min.X {
		t.Fatalf("expected larger left margin for larger labels, got %v <= %v", largeAx.RectFraction.Min.X, smallAx.RectFraction.Min.X)
	}
}

func TestDrawFigure_ConstrainedLayoutRespondsToNeighboringDecorations(t *testing.T) {
	small := newConstrainedLayoutProbeFigure()
	small.ConstrainedLayout()
	DrawFigure(small, &figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"edge": {X: 0, Y: -8, W: 18, H: 10},
		},
	})

	large := newConstrainedLayoutProbeFigure()
	large.ConstrainedLayout()
	DrawFigure(large, &figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"edge": {X: 0, Y: -10, W: 64, H: 14},
		},
	})

	smallGap := small.Children[1].RectFraction.Min.X - small.Children[0].RectFraction.Max.X
	largeGap := large.Children[1].RectFraction.Min.X - large.Children[0].RectFraction.Max.X
	if largeGap <= smallGap {
		t.Fatalf("expected larger constrained-layout gap for larger labels, got %v <= %v", largeGap, smallGap)
	}
}

func TestConstrainedLayoutKeepsNestedYAxisTickDensityReadable(t *testing.T) {
	fig := NewFigure(960, 640)
	fig.ConstrainedLayout()

	outer := fig.GridSpec(
		2,
		2,
		WithGridSpecPadding(0.08, 0.96, 0.10, 0.92),
		WithGridSpecSpacing(0.06, 0.08),
		WithGridSpecWidthRatios(2, 1),
	)
	outer.Span(0, 0, 2, 1).AddAxes()
	nested := outer.Cell(0, 1).GridSpec(2, 1, WithGridSpecSpacing(0, 0.12))
	top := nested.Cell(0, 0).AddAxes()
	top.Plot([]float64{0, 1, 2, 3}, []float64{3.4, 2.6, 2.9, 1.8})
	top.AutoScale(0.10)
	bottom := nested.Cell(1, 0).AddAxes(WithSharedX(top))
	bottom.Plot([]float64{0, 1, 2, 3}, []float64{1.0, 1.6, 1.3, 2.2})
	bottom.AutoScale(0.10)
	outer.Cell(1, 1).SubFigure().AddSubplot(1, 1, 1)

	var r figureLayoutRecordingRenderer
	DrawFigure(fig, &r)

	ctx := newAxesDrawContext(top, fig, fig.DisplayRect(), top.adjustedLayout(fig))
	yMin, yMax := ctx.DataToPixel.YScale.Domain()
	ticks := visibleTicks(top.YAxis.Locator.Ticks(yMin, yMax, top.YAxis.majorTickTargetCountForContext(ctx, false)), yMin, yMax)
	if len(ticks) > 4 {
		t.Fatalf("nested y-axis tick density too high for small panel: ticks=%v clip=%+v", ticks, ctx.Clip)
	}
}

func TestConstrainedLayoutReservesColorbarSpaceAndTracksParent(t *testing.T) {
	fig := NewFigure(1000, 700)
	fig.ConstrainedLayout()
	if got, want := layoutPadPx(fig, LayoutEngineConstrained), pointsToPixels(fig.RC, 3); !floatApprox(got, want, 1e-12) {
		t.Fatalf("constrained layout pad = %v, want matplotlib default %v", got, want)
	}
	ax := fig.AddSubplot(1, 1, 1)
	img := ax.Image([][]float64{{0, 1}, {2, 3}})
	cb := fig.AddColorbar(ax, img, ColorbarOptions{Label: "Intensity"})
	if cb == nil {
		t.Fatal("expected colorbar axes")
	}

	var r figureLayoutRecordingRenderer
	DrawFigure(fig, &r)

	if cb.RectFraction.Max.X > 1 {
		t.Fatalf("colorbar overflowed figure: %+v", cb.RectFraction)
	}
	if cb.RectFraction.Min.X <= ax.RectFraction.Max.X {
		t.Fatalf("colorbar did not stay to the right of parent: parent=%+v colorbar=%+v", ax.RectFraction, cb.RectFraction)
	}
	base := cb.colorbarBase
	if got, want := cb.RectFraction.Min.X-ax.RectFraction.Max.X, resolvedColorbarPadding(base, cb.colorbarPadding)+constrainedColorbarSlotOffset(fig, base); !floatApprox(got, want, 1e-9) {
		t.Fatalf("colorbar did not track parent padding: got %v want %v", got, want)
	}
	if ax.RectFraction.Max.X >= 0.90 {
		t.Fatalf("constrained layout did not reserve right margin for colorbar: parent=%+v colorbar=%+v", ax.RectFraction, cb.RectFraction)
	}
}

func TestConstrainedLayoutMeasuresBottomXLabelLineBox(t *testing.T) {
	fig := NewFigure(1000, 700)
	fig.ConstrainedLayout()
	ax := fig.AddSubplot(1, 1, 1)
	ax.SetXLabel("x")
	img := ax.Image([][]float64{{0, 1}, {2, 3}})
	if cb := fig.AddColorbar(ax, img, ColorbarOptions{Label: "Intensity"}); cb == nil {
		t.Fatal("expected colorbar axes")
	}

	var r figureLayoutRecordingRenderer
	DrawFigure(fig, &r)

	px := ax.layout(fig)
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), px)
	wantMinBottom := layoutPadPx(fig, LayoutEngineConstrained) +
		tickLabelPadPx(ax.XAxis, ctx) +
		pointsToPixels(ctx.RC, tickLabelFontSize(ax.XAxis, ctx)) +
		axisLabelPadPx(ctx) +
		pointsToPixels(ctx.RC, axisLabelFontSize(ctx))
	// Display space is y-up: the bottom margin (where the x-label stack lives) is
	// the axes' lower edge px.Min.Y above the figure bottom.
	gotBottom := px.Min.Y
	if gotBottom+0.25 < wantMinBottom {
		t.Fatalf("bottom margin = %v, want at least full x-label line-box stack %v", gotBottom, wantMinBottom)
	}
}

func TestFigureLabelTightHeightUsesLineBoxWhenInkBoundsAreShort(t *testing.T) {
	r := &figureLayoutRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"time": {X: 0, Y: -4, W: 20, H: 6},
		},
	}

	got := figureLabelTightHeight(r, "time", 12, "", false)
	if !floatApprox(got, 12, 1e-9) {
		t.Fatalf("figureLabelTightHeight = %v, want line-box height 12", got)
	}
}

func newTightLayoutProbeFigure() *Figure {
	fig := NewFigure(800, 600)
	ax := fig.AddSubplot(1, 1, 1)
	ax.SetTitle("Overview")
	ax.SetXLabel("time")
	ax.SetYLabel("value")
	ax.XAxis.Locator = staticLocator{1}
	ax.XAxis.Formatter = FixedFormatter{Labels: []string{"edge"}}
	ax.YAxis.Locator = staticLocator{1}
	ax.YAxis.Formatter = FixedFormatter{Labels: []string{"edge"}}
	return fig
}

func newConstrainedLayoutProbeFigure() *Figure {
	fig := NewFigure(800, 600)
	grid := fig.Subplots(1, 2)
	grid[0][0].RightAxis().Locator = staticLocator{1}
	grid[0][0].RightAxis().Formatter = FixedFormatter{Labels: []string{"edge"}}
	grid[0][1].YAxis.Locator = staticLocator{1}
	grid[0][1].YAxis.Formatter = FixedFormatter{Labels: []string{"edge"}}
	return fig
}

type figureLayoutRecordingRenderer struct {
	render.NullRenderer
	bounds         map[string]render.TextBounds
	texts          []string
	origins        []geom.Pt
	rotatedText    []string
	rotatedAnchors []geom.Pt
	pathCount      int
}

func (r *figureLayoutRecordingRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.pathCount++
}

func (r *figureLayoutRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	metrics := render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
	if bounds, ok := r.bounds[text]; ok {
		if bounds.W > metrics.W {
			metrics.W = bounds.W
		}
		ascent := -bounds.Y
		if ascent > metrics.Ascent {
			metrics.Ascent = ascent
		}
		descent := bounds.Y + bounds.H
		if descent > metrics.Descent {
			metrics.Descent = descent
		}
		metrics.H = metrics.Ascent + metrics.Descent
	}
	return metrics
}

func (r *figureLayoutRecordingRenderer) MeasureTextBounds(text string, _ float64, _ string) (render.TextBounds, bool) {
	b, ok := r.bounds[text]
	return b, ok
}

func (r *figureLayoutRecordingRenderer) DrawText(text string, origin geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, origin)
}

func (r *figureLayoutRecordingRenderer) DrawTextRotated(text string, anchor geom.Pt, _ float64, _ float64, _ render.Color) {
	r.rotatedText = append(r.rotatedText, text)
	r.rotatedAnchors = append(r.rotatedAnchors, anchor)
}

func (r *figureLayoutRecordingRenderer) textOrigin(text string) (geom.Pt, bool) {
	for i, candidate := range r.texts {
		if candidate == text {
			return r.origins[i], true
		}
	}
	return geom.Pt{}, false
}

func (r *figureLayoutRecordingRenderer) rotatedAnchor(text string) (geom.Pt, bool) {
	for i, candidate := range r.rotatedText {
		if candidate == text {
			return r.rotatedAnchors[i], true
		}
	}
	return geom.Pt{}, false
}

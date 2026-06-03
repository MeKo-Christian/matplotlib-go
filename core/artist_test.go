package core

import (
	"image"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestRCEffectivePrecedence(t *testing.T) {
	fig := NewFigure(800, 600, style.WithDPI(110))
	axInherit := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	if got := axInherit.effectiveRC(fig).DPI; got != 110 {
		t.Fatalf("expected axes inherit figure DPI, got %v", got)
	}
	axOverride := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.2, Y: 0.2}, Max: geom.Pt{X: 0.8, Y: 0.8}}, style.WithDPI(200))
	if got := axOverride.effectiveRC(fig).DPI; got != 200 {
		t.Fatalf("expected axes override DPI=200, got %v", got)
	}
}

// artist with custom z
type zArtist struct {
	z   float64
	id  int
	hit *[]int
}

func (a zArtist) Draw(_ render.Renderer, _ *DrawContext) { *a.hit = append(*a.hit, a.id) }
func (a zArtist) Z() float64                             { return a.z }
func (a zArtist) Bounds(*DrawContext) geom.Rect          { return geom.Rect{} }

type rasterizedTestArtist struct {
	zArtist
	ArtistRasterization
}

type metadataTestArtist struct {
	ArtistRasterization
	draws *int
}

func (a *metadataTestArtist) Draw(render.Renderer, *DrawContext) {
	*a.draws++
}

func (a *metadataTestArtist) Z() float64 { return 0 }

func (a *metadataTestArtist) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

type rasterizationRecordingRenderer struct {
	render.NullRenderer
	events  []string
	options []render.Rasterization
}

func (r *rasterizationRecordingRenderer) StartRasterized(options render.Rasterization) bool {
	r.events = append(r.events, "start")
	r.options = append(r.options, options)
	return true
}

func (r *rasterizationRecordingRenderer) StopRasterized() bool {
	r.events = append(r.events, "stop")
	return true
}

func (r *rasterizationRecordingRenderer) Path(path geom.Path, paint *render.Paint) {
	render.DrawPathWithEffects(r, path, paint, func(geom.Path, *render.Paint) {})
}

type rasterizationFilterRecordingRenderer struct {
	rasterizationRecordingRenderer
	filterStarts int
	filterStops  int
}

func (r *rasterizationFilterRecordingRenderer) StartFilter() {
	r.filterStarts++
}

func (r *rasterizationFilterRecordingRenderer) StopFilter(func(*image.RGBA, float64) (*image.RGBA, geom.Pt)) {
	r.filterStops++
}

func (r *rasterizationFilterRecordingRenderer) Path(path geom.Path, paint *render.Paint) {
	render.DrawPathWithEffects(r, path, paint, func(geom.Path, *render.Paint) {})
}

type rasterizationNativeFilterRecordingRenderer struct {
	rasterizationRecordingRenderer
	nativeFilters int
}

func (r *rasterizationNativeFilterRecordingRenderer) SupportsPathEffectFilter(effect render.PathEffect) bool {
	return effect.Filter == "blur"
}

func (r *rasterizationNativeFilterRecordingRenderer) DrawPathEffectFilter(path geom.Path, paint render.Paint, effect render.PathEffect, draw func(geom.Path, *render.Paint)) bool {
	if effect.Filter != "blur" {
		return false
	}
	r.nativeFilters++
	draw(path, &paint)
	return true
}

func (r *rasterizationNativeFilterRecordingRenderer) Path(path geom.Path, paint *render.Paint) {
	render.DrawPathWithEffects(r, path, paint, func(geom.Path, *render.Paint) {})
}

func TestRasterizedArtistDrawIsBracketedWhenRendererSupportsMixedOutput(t *testing.T) {
	fig := NewFigure(100, 100, style.WithDPI(144))
	var order []int
	vectorBefore := zArtist{z: 1, id: 1, hit: &order}
	rasterized := &rasterizedTestArtist{
		zArtist: zArtist{z: 2, id: 2, hit: &order},
	}
	vectorAfter := zArtist{z: 3, id: 3, hit: &order}
	rasterized.SetRasterized(true)

	fig.Add(vectorBefore)
	fig.Add(rasterized)
	fig.Add(vectorAfter)

	ren := &rasterizationRecordingRenderer{}
	DrawFigure(fig, ren)

	wantOrder := []int{1, 2, 3}
	for i := range wantOrder {
		if order[i] != wantOrder[i] {
			t.Fatalf("draw order = %v, want %v", order, wantOrder)
		}
	}
	wantEvents := []string{"start", "stop"}
	if len(ren.events) != len(wantEvents) {
		t.Fatalf("raster events = %v, want %v", ren.events, wantEvents)
	}
	for i := range wantEvents {
		if ren.events[i] != wantEvents[i] {
			t.Fatalf("raster events = %v, want %v", ren.events, wantEvents)
		}
	}
	if len(ren.options) != 1 {
		t.Fatalf("raster options count = %d, want 1", len(ren.options))
	}
	if ren.options[0].Mode != render.RasterizeAlways {
		t.Fatalf("raster mode = %v, want RasterizeAlways", ren.options[0].Mode)
	}
	if ren.options[0].DPI != 144 {
		t.Fatalf("raster DPI = %v, want 144", ren.options[0].DPI)
	}
}

func TestCommonArtistsExposeRasterizedFlag(t *testing.T) {
	type rasterizable interface {
		SetRasterized(bool)
		Rasterization() render.Rasterization
	}

	var _ rasterizable = (*Line2D)(nil)
	var _ rasterizable = (*Scatter2D)(nil)
	var _ rasterizable = (*Image2D)(nil)
	var _ rasterizable = (*ContourSet)(nil)
	var _ rasterizable = (*Collection)(nil)
	var _ rasterizable = (*Bar2D)(nil)
	var _ rasterizable = (*Fill2D)(nil)
	var _ rasterizable = (*Patch)(nil)
	var _ rasterizable = (*Rectangle)(nil)
	var _ rasterizable = (*Text)(nil)
	var _ rasterizable = (*Annotation)(nil)
}

func TestArtistMetadataVisibilitySkipsDrawing(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	draws := 0

	figArtist := &metadataTestArtist{draws: &draws}
	axesArtist := &metadataTestArtist{draws: &draws}
	figArtist.SetVisible(false)
	axesArtist.SetVisible(false)

	fig.Add(figArtist)
	ax.Add(axesArtist)
	DrawFigure(fig, &render.NullRenderer{})

	if draws != 0 {
		t.Fatalf("invisible artists drew %d times, want 0", draws)
	}
	if !figArtist.Stale() || !axesArtist.Stale() {
		t.Fatal("metadata changes should mark artists stale")
	}
}

func TestDrawFigureSkipsAnimatedArtistsByDefault(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	draws := 0
	figArtist := &metadataTestArtist{draws: &draws}
	axesArtist := &metadataTestArtist{draws: &draws}
	figArtist.SetAnimated(true)
	axesArtist.SetAnimated(true)
	fig.Add(figArtist)
	ax.Add(axesArtist)

	DrawFigure(fig, &render.NullRenderer{})
	if draws != 0 {
		t.Fatalf("animated artists drew %d times under default options, want 0", draws)
	}
}

func TestDrawFigureWithOptionsOnlyAnimatedDrawsAnimatedOnly(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	animatedDraws := 0
	staticDraws := 0
	animated := &metadataTestArtist{draws: &animatedDraws}
	static := &metadataTestArtist{draws: &staticDraws}
	animated.SetAnimated(true)
	ax.Add(animated)
	ax.Add(static)

	DrawFigureWithOptions(fig, &render.NullRenderer{}, DrawOptions{AnimatedFilter: AnimatedFilterOnlyAnimated})
	if animatedDraws != 1 || staticDraws != 0 {
		t.Fatalf("only-animated pass drew animated=%d static=%d, want 1/0", animatedDraws, staticDraws)
	}
}

func TestDrawFigureWithOptionsAllDrawsBoth(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	animatedDraws := 0
	staticDraws := 0
	animated := &metadataTestArtist{draws: &animatedDraws}
	static := &metadataTestArtist{draws: &staticDraws}
	animated.SetAnimated(true)
	ax.Add(animated)
	ax.Add(static)

	DrawFigureWithOptions(fig, &render.NullRenderer{}, DrawOptions{AnimatedFilter: AnimatedFilterAll})
	if animatedDraws != 1 || staticDraws != 1 {
		t.Fatalf("all pass drew animated=%d static=%d, want 1/1", animatedDraws, staticDraws)
	}
}

func TestArtistMetadataDefaults(t *testing.T) {
	var metadata ArtistRasterization
	if !metadata.Visible() {
		t.Fatal("zero-value artist metadata should be visible")
	}
	if got := metadata.EffectiveAlpha(0); got != 1 {
		t.Fatalf("zero-value effective alpha = %v, want 1", got)
	}
	if !metadata.InLayout() {
		t.Fatal("zero-value artist metadata should participate in layout")
	}
}

func TestDenseScatterAutoRasterizesWithFigureDPI(t *testing.T) {
	fig := NewFigure(100, 100, style.WithDPI(180))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	x := make([]float64, autoRasterizeScatterPointThreshold)
	y := make([]float64, autoRasterizeScatterPointThreshold)
	for i := range x {
		x[i] = float64(i) / float64(len(x))
		y[i] = float64(i%10) / 10
	}
	ax.Scatter(x, y)

	ren := &rasterizationRecordingRenderer{}
	DrawFigure(fig, ren)

	if len(ren.options) != 1 {
		t.Fatalf("raster option count = %d, want 1", len(ren.options))
	}
	if got := ren.options[0]; got.Mode != render.RasterizeAuto || got.DPI != 180 {
		t.Fatalf("raster options = %+v, want auto at 180dpi", got)
	}
}

func TestSmallScatterDoesNotAutoRasterize(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.Scatter([]float64{0, 1}, []float64{0, 1})

	ren := &rasterizationRecordingRenderer{}
	DrawFigure(fig, ren)

	if len(ren.options) != 0 {
		t.Fatalf("small scatter rasterized unexpectedly: %+v", ren.options)
	}
}

func TestUnsupportedFilterPathEffectAutoRasterizes(t *testing.T) {
	fig := NewFigure(100, 100, style.WithDPI(96))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	line := ax.Plot([]float64{0, 1}, []float64{0, 1})
	line.PathEffects = []render.PathEffect{
		render.FilterPathEffect(
			render.Color{R: 1, A: 1},
			render.Color{B: 1, A: 1},
			2,
			"blur",
			2,
			geom.Pt{X: 1, Y: 1},
		),
	}

	ren := &rasterizationRecordingRenderer{}
	DrawFigure(fig, ren)

	if len(ren.options) != 1 {
		t.Fatalf("filter path effect raster option count = %d, want 1", len(ren.options))
	}
	if got := ren.options[0]; got.Mode != render.RasterizeAuto || got.DPI != 96 {
		t.Fatalf("filter path effect raster options = %+v, want auto at 96dpi", got)
	}
}

func TestDenseContourAutoRasterizesWithFigureDPI(t *testing.T) {
	fig := NewFigure(100, 100, style.WithDPI(150))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	segments := make([][]geom.Pt, autoRasterizeContourPathThreshold)
	for i := range segments {
		x := float64(i) / float64(len(segments))
		segments[i] = []geom.Pt{{X: x, Y: 0}, {X: x, Y: 1}}
	}
	ax.Add(&ContourSet{
		Lines: &LineCollection{
			Segments:  segments,
			Color:     render.Color{A: 1},
			LineWidth: 1,
		},
	})

	ren := &rasterizationRecordingRenderer{}
	DrawFigure(fig, ren)

	if len(ren.options) != 1 {
		t.Fatalf("dense contour raster option count = %d, want 1", len(ren.options))
	}
	if got := ren.options[0]; got.Mode != render.RasterizeAuto || got.DPI != 150 {
		t.Fatalf("dense contour raster options = %+v, want auto at 150dpi", got)
	}
}

func TestFilterCapableRendererKeepsFilterPathEffectVector(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	line := ax.Plot([]float64{0, 1}, []float64{0, 1})
	line.PathEffects = []render.PathEffect{
		render.FilterPathEffect(
			render.Color{R: 1, A: 1},
			render.Color{B: 1, A: 1},
			2,
			"blur",
			2,
			geom.Pt{X: 1, Y: 1},
		),
	}

	ren := &rasterizationFilterRecordingRenderer{}
	DrawFigure(fig, ren)

	if len(ren.options) != 0 {
		t.Fatalf("filter-capable renderer rasterized unexpectedly: %+v", ren.options)
	}
	if ren.filterStarts == 0 || ren.filterStops == 0 {
		t.Fatalf("filter-capable renderer did not use filter path: starts=%d stops=%d", ren.filterStarts, ren.filterStops)
	}
}

func TestUnsupportedNativeFilterPathEffectAutoRasterizes(t *testing.T) {
	fig := NewFigure(100, 100, style.WithDPI(110))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	line := ax.Plot([]float64{0, 1}, []float64{0, 1})
	line.PathEffects = []render.PathEffect{
		render.FilterPathEffect(
			render.Color{R: 1, A: 1},
			render.Color{B: 1, A: 1},
			2,
			"emboss",
			2,
			geom.Pt{X: 1, Y: 1},
		),
	}

	ren := &rasterizationNativeFilterRecordingRenderer{}
	DrawFigure(fig, ren)

	if len(ren.options) != 1 {
		t.Fatalf("unsupported native filter raster option count = %d, want 1", len(ren.options))
	}
	if ren.nativeFilters != 0 {
		t.Fatalf("unsupported filter should not use native filter hook, got %d calls", ren.nativeFilters)
	}
}

func TestZOrderStableSortAndTraversal(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	var order []int
	// Insertion order: ids 1..5, with equal z for 2 and 3
	ax.Add(zArtist{z: 0, id: 1, hit: &order})
	ax.Add(zArtist{z: 1, id: 2, hit: &order})
	ax.Add(zArtist{z: 1, id: 3, hit: &order})
	ax.Add(zArtist{z: -1, id: 4, hit: &order})
	ax.Add(zArtist{z: 2, id: 5, hit: &order})

	var r render.NullRenderer
	DrawFigure(fig, &r)

	// Expected draw order: z=-1 (id4), z=0 (id1), z=1 (ids 2 then 3), z=2 (id5)
	want := []int{4, 1, 2, 3, 5}
	if len(order) != len(want) {
		t.Fatalf("draw count mismatch: got %d want %d", len(order), len(want))
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order mismatch at %d: got %v want %v (full=%v)", i, order[i], want[i], order)
		}
	}
}

func TestTitleFontSizeUsesTitleOnlyCompensation(t *testing.T) {
	ctx := &DrawContext{RC: style.RC{FontSize: 12}}

	got := titleFontSize(ctx)
	want := 14.4

	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("titleFontSize() = %v, want %v", got, want)
	}
}

type axesLabelRecordingRenderer struct {
	render.NullRenderer
	bounds         map[string]render.TextBounds
	texts          []string
	origins        []geom.Pt
	sizes          []float64
	rotatedText    []string
	rotatedAnchors []geom.Pt
	rotatedSizes   []float64
}

func (r *axesLabelRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	switch text {
	case "4":
		return render.TextMetrics{W: 5, H: 10, Ascent: 8, Descent: 2}
	case "Value":
		return render.TextMetrics{W: size * 2, H: 10, Ascent: 8, Descent: 2}
	default:
		return render.TextMetrics{W: float64(len(text)) * size * 0.5, H: size, Ascent: size * 0.8, Descent: size * 0.2}
	}
}

func (r *axesLabelRecordingRenderer) MeasureTextBounds(text string, _ float64, _ string) (render.TextBounds, bool) {
	b, ok := r.bounds[text]
	return b, ok
}

func (r *axesLabelRecordingRenderer) DrawText(text string, origin geom.Pt, size float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, origin)
	r.sizes = append(r.sizes, size)
}

func (r *axesLabelRecordingRenderer) DrawTextRotated(text string, anchor geom.Pt, size float64, _ float64, _ render.Color) {
	r.rotatedText = append(r.rotatedText, text)
	r.rotatedAnchors = append(r.rotatedAnchors, anchor)
	r.rotatedSizes = append(r.rotatedSizes, size)
}

func TestDrawAxesLabels_YLabelUsesTickBoundsAndLabelPad(t *testing.T) {
	ax := &Axes{
		YAxis:  NewYAxis(),
		YLabel: "Value",
	}
	ax.YAxis.Locator = staticLocator{4}
	ax.YAxis.Formatter = ScalarFormatter{Prec: 0}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72
	px := geom.Rect{
		Min: geom.Pt{X: 50, Y: 350},
		Max: geom.Pt{X: 150, Y: 450},
	}

	r := &axesLabelRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"4": {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	defer r.End()

	drawAxesLabels(ax, r, ctx, px, figureTextAlignment{})

	if len(r.rotatedText) != 1 || r.rotatedText[0] != "Value" {
		t.Fatalf("unexpected rotated text draws: %v", r.rotatedText)
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: getSpinePosition(ax.YAxis, ctx), Y: 4})
	tickLabelMinX := tickPos.X - tickLabelPadPx(ax.YAxis, ctx) - (1 + 5.0) + 1
	// P is matplotlib's label anchor: min(spine, tick bounds) - labelpad,
	// vertically centered. The spine extent already includes its line width.
	p := geom.Pt{
		X: math.Min(spinePixelX(AxisLeft, px), tickLabelMinX) - axisLabelPadPx(ctx),
		Y: px.Min.Y + px.H()/2,
	}
	// matplotlib draws the left y-label at rotation=90, rotation_mode="anchor",
	// ha="center", va="bottom"; the backend pivot is derived via _get_layout.
	layout := measureSingleLineTextLayout(r, "Value", axisLabelFontSize(ctx), ctx.RC.FontKey)
	want := rotatedTextBackendAnchorFromP(p, layout, TextAlignCenter, textLayoutVAlignBottom, math.Pi/2, true)
	if r.rotatedAnchors[0] != want {
		t.Fatalf("ylabel anchor = %+v, want %+v", r.rotatedAnchors[0], want)
	}
}

func TestDrawAxesLabels_XLabelUsesTickBoundsAndLabelPad(t *testing.T) {
	ax := &Axes{
		XAxis:  NewXAxis(),
		XLabel: "Group",
	}
	ax.XAxis.Locator = staticLocator{2}
	ax.XAxis.Formatter = ScalarFormatter{Prec: 0}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72
	px := geom.Rect{
		Min: geom.Pt{X: 50, Y: 350},
		Max: geom.Pt{X: 150, Y: 450},
	}

	r := &axesLabelRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"2": {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	defer r.End()

	drawAxesLabels(ax, r, ctx, px, figureTextAlignment{})

	if len(r.texts) != 1 || r.texts[0] != "Group" {
		t.Fatalf("unexpected text draws: %v", r.texts)
	}

	layout := measureSingleLineTextLayout(r, "Group", axisLabelFontSize(ctx), ctx.RC.FontKey)
	// Display space is y-up: the bottom x-label sits below the axis at the
	// lowest extent minus the pad (matches drawAxesLabels).
	bottomExtent := spinePixelY(AxisBottom, px)
	if tickBounds, ok := axisTickLabelBounds(ax.XAxis, r, ctx); ok {
		bottomExtent = math.Min(bottomExtent, tickBounds.Min.Y)
	}
	want := alignedSingleLineOrigin(
		geom.Pt{
			X: ctx.TransAxes().Apply(geom.Pt{X: 0.5, Y: 0}).X,
			Y: bottomExtent - axisLabelPadPx(ctx),
		},
		layout,
		TextAlignCenter,
		textLayoutVAlignTop,
	)
	if r.origins[0] != want {
		t.Fatalf("xlabel origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestDrawAxesLabels_YLabelRightUsesRightTickBounds(t *testing.T) {
	ax := &Axes{
		YAxis:      NewYAxis(),
		YAxisRight: NewYAxis(),
		YLabel:     "Value",
		yLabelSide: AxisRight,
	}
	ax.YAxisRight.Side = AxisRight
	ax.YAxisRight.Locator = staticLocator{4}
	ax.YAxisRight.Formatter = ScalarFormatter{Prec: 0}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72
	px := geom.Rect{
		Min: geom.Pt{X: 50, Y: 350},
		Max: geom.Pt{X: 150, Y: 450},
	}

	r := &axesLabelRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"4": {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	defer r.End()

	drawAxesLabels(ax, r, ctx, px, figureTextAlignment{})

	if len(r.rotatedText) != 1 || r.rotatedText[0] != "Value" {
		t.Fatalf("unexpected rotated text draws: %v", r.rotatedText)
	}

	rightExtent := spinePixelX(AxisRight, px)
	if tickBounds, ok := axisTickLabelBounds(ax.YAxisRight, r, ctx); ok {
		rightExtent = math.Max(rightExtent, tickBounds.Max.X)
	}
	// P is matplotlib's label anchor: max(spine, tick bounds) + labelpad, centered.
	p := geom.Pt{
		X: rightExtent + axisLabelPadPx(ctx),
		Y: px.Min.Y + px.H()/2,
	}
	// matplotlib draws the right y-label at rotation=90, rotation_mode="anchor",
	// ha="center", va="top"; the backend pivot is derived via _get_layout.
	layout := measureSingleLineTextLayout(r, "Value", axisLabelFontSize(ctx), ctx.RC.FontKey)
	want := rotatedTextBackendAnchorFromP(p, layout, TextAlignCenter, textLayoutVAlignTop, math.Pi/2, true)
	if r.rotatedAnchors[0] != want {
		t.Fatalf("right ylabel anchor = %+v, want %+v", r.rotatedAnchors[0], want)
	}
}

func TestDrawAxesLabels_YLabelUsesTickPaddingWhenFormatterSuppressesLabels(t *testing.T) {
	ax := &Axes{
		YAxis:  NewYAxis(),
		YLabel: "Value",
	}
	ax.YAxis.Locator = staticLocator{0.5}
	ax.YAxis.Formatter = NullFormatter{}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72
	px := geom.Rect{
		Min: geom.Pt{X: 50, Y: 350},
		Max: geom.Pt{X: 150, Y: 450},
	}

	r := &axesLabelRecordingRenderer{}
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	defer r.End()

	drawAxesLabels(ax, r, ctx, px, figureTextAlignment{})

	if len(r.rotatedText) != 1 || r.rotatedText[0] != "Value" {
		t.Fatalf("unexpected rotated text draws: %v", r.rotatedText)
	}

	p := geom.Pt{
		X: spinePixelX(AxisLeft, px) - tickLabelPadPx(ax.YAxis, ctx) - axisLabelPadPx(ctx),
		Y: px.Min.Y + px.H()/2,
	}
	layout := measureSingleLineTextLayout(r, "Value", axisLabelFontSize(ctx), ctx.RC.FontKey)
	want := rotatedTextBackendAnchorFromP(p, layout, TextAlignCenter, textLayoutVAlignBottom, math.Pi/2, true)
	if r.rotatedAnchors[0] != want {
		t.Fatalf("ylabel anchor = %+v, want tick-padded %+v", r.rotatedAnchors[0], want)
	}
}

func TestDrawAxesLabels_TopXLabelUsesTopTickBoundsAndLabelPad(t *testing.T) {
	ax := &Axes{
		XAxis:      NewXAxis(),
		XAxisTop:   NewXAxis(),
		XLabel:     "Group",
		xLabelSide: AxisTop,
	}
	ax.XAxisTop.Side = AxisTop
	ax.XAxisTop.Locator = staticLocator{2}
	ax.XAxisTop.Formatter = ScalarFormatter{Prec: 0}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72
	px := geom.Rect{
		Min: geom.Pt{X: 50, Y: 350},
		Max: geom.Pt{X: 150, Y: 450},
	}

	r := &axesLabelRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"2": {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	defer r.End()

	drawAxesLabels(ax, r, ctx, px, figureTextAlignment{})

	if len(r.texts) != 1 || r.texts[0] != "Group" {
		t.Fatalf("unexpected text draws: %v", r.texts)
	}

	layout := measureSingleLineTextLayout(r, "Group", axisLabelFontSize(ctx), ctx.RC.FontKey)
	// Display space is y-up: the top x-label sits above the axis at the highest
	// extent plus the pad (matches drawAxesLabels).
	topExtent := spinePixelY(AxisTop, px)
	if tickBounds, ok := axisTickLabelBounds(ax.XAxisTop, r, ctx); ok {
		topExtent = math.Max(topExtent, tickBounds.Max.Y)
	}
	want := alignedSingleLineOrigin(
		geom.Pt{
			X: ctx.TransAxes().Apply(geom.Pt{X: 0.5, Y: 0}).X,
			Y: topExtent + axisLabelPadPx(ctx),
		},
		layout,
		TextAlignCenter,
		textLayoutVAlignBaseline,
	)
	if r.origins[0] != want {
		t.Fatalf("top xlabel origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestDrawAxesLabels_TitleClearsTopXLabel(t *testing.T) {
	ax := &Axes{
		XAxis:      NewXAxis(),
		XAxisTop:   NewXAxis(),
		Title:      "Title",
		XLabel:     "Group",
		xLabelSide: AxisTop,
	}
	ax.XAxisTop.Side = AxisTop
	ax.XAxisTop.Locator = staticLocator{2}
	ax.XAxisTop.Formatter = ScalarFormatter{Prec: 0}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72
	px := geom.Rect{
		Min: geom.Pt{X: 50, Y: 350},
		Max: geom.Pt{X: 150, Y: 450},
	}

	r := &axesLabelRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"2":     {X: 1, Y: -8, W: 5, H: 10},
			"Group": {X: 0, Y: -8, W: 24, H: 10},
		},
	}
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	defer r.End()

	drawAxesLabels(ax, r, ctx, px, figureTextAlignment{})

	if len(r.texts) != 2 || r.texts[0] != "Title" || r.texts[1] != "Group" {
		t.Fatalf("unexpected text draws: %v", r.texts)
	}
	titleLayout := measureSingleLineTextLayout(r, "Title", titleFontSize(ctx), ctx.RC.FontKey)
	titleBounds, ok := textInkRect(r.origins[0], titleLayout)
	if !ok {
		t.Fatal("expected title bounds")
	}
	xlabelLayout := measureSingleLineTextLayout(r, "Group", axisLabelFontSize(ctx), ctx.RC.FontKey)
	xlabelBounds, ok := textInkRect(r.origins[1], xlabelLayout)
	if !ok {
		t.Fatal("expected xlabel bounds")
	}
	// Display space is y-up: the title clears the top x-label when its bottom
	// edge (Min.Y) sits at or above the x-label's top edge (Max.Y).
	if titleBounds.Min.Y < xlabelBounds.Max.Y {
		t.Fatalf("title overlaps top xlabel: title=%+v xlabel=%+v", titleBounds, xlabelBounds)
	}
}

func TestDrawAxesLabels_TitleAboveTopXLabelUsesMatplotlibSecondAdjustment(t *testing.T) {
	ax := &Axes{
		XAxis:      NewXAxis(),
		XAxisTop:   NewXAxis(),
		Title:      "Title",
		XLabel:     "Group",
		xLabelSide: AxisTop,
	}
	ax.XAxisTop.Side = AxisTop
	ax.XAxisTop.Locator = staticLocator{2}
	ax.XAxisTop.Formatter = ScalarFormatter{Prec: 0}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 100
	px := geom.Rect{
		Min: geom.Pt{X: 50, Y: 350},
		Max: geom.Pt{X: 150, Y: 450},
	}

	r := &axesLabelRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"2":     {X: 1, Y: -8, W: 5, H: 10},
			"Group": {X: 0, Y: -8, W: 24, H: 10},
			"Title": {X: 0, Y: -10, W: 30, H: 12},
		},
	}
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	defer r.End()

	drawAxesLabels(ax, r, ctx, px, figureTextAlignment{})

	if len(r.texts) != 2 || r.texts[0] != "Title" || r.texts[1] != "Group" {
		t.Fatalf("unexpected text draws: %v", r.texts)
	}
	titleLayout := measureSingleLineTextLayout(r, "Title", titleFontSize(ctx), ctx.RC.FontKey)
	titleBounds, ok := textInkRect(r.origins[0], titleLayout)
	if !ok {
		t.Fatal("expected title bounds")
	}
	topExtent := titleTopExtent(ax, r, ctx, px)
	wantBounds, ok := textInkRect(geom.Pt{
		X: r.origins[0].X,
		Y: topExtent + pointsToPixels(ctx.RC, 6) + 1,
	}, titleLayout)
	if !ok {
		t.Fatal("expected adjusted title bounds")
	}
	want := wantBounds.Min.Y
	if math.Abs(titleBounds.Min.Y-want) > 1e-9 {
		t.Fatalf("title bottom = %v, want %v after Matplotlib second adjustment", titleBounds.Min.Y, want)
	}
}

func TestDrawAxesLabels_UsesSameFontSizeForXAndYLabels(t *testing.T) {
	ax := &Axes{
		XAxis:  NewXAxis(),
		YAxis:  NewYAxis(),
		XLabel: "Group",
		YLabel: "Value",
	}
	ax.XAxis.Locator = staticLocator{2}
	ax.XAxis.Formatter = ScalarFormatter{Prec: 0}
	ax.YAxis.Locator = staticLocator{4}
	ax.YAxis.Formatter = ScalarFormatter{Prec: 0}

	ctx := createTestDrawContext()
	r := &axesLabelRecordingRenderer{
		bounds: map[string]render.TextBounds{
			"2": {X: 1, Y: -8, W: 5, H: 10},
			"4": {X: 1, Y: -8, W: 5, H: 10},
		},
	}
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	defer r.End()

	drawAxesLabels(ax, r, ctx, geom.Rect{
		Min: geom.Pt{X: 50, Y: 350},
		Max: geom.Pt{X: 150, Y: 450},
	}, figureTextAlignment{})

	if len(r.sizes) != 1 || len(r.rotatedSizes) != 1 {
		t.Fatalf("unexpected label draw sizes: text=%v rotated=%v", r.sizes, r.rotatedSizes)
	}
	if r.sizes[0] != r.rotatedSizes[0] {
		t.Fatalf("x/y label font sizes differ: x=%v y=%v", r.sizes[0], r.rotatedSizes[0])
	}
	if r.sizes[0] != axisLabelFontSize(ctx) {
		t.Fatalf("axis label font size = %v, want %v", r.sizes[0], axisLabelFontSize(ctx))
	}
}

func TestDrawContextTransformsExposeCoordinateSpaces(t *testing.T) {
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(-5, 5),
			DataToAxes:  transform.NewScaleTransform(transform.NewLinear(0, 10), transform.NewLinear(-5, 5)),
			AxesToPixel: transform.NewDisplayRectTransform(geom.Rect{Min: geom.Pt{X: 50, Y: 100}, Max: geom.Pt{X: 250, Y: 300}}),
		},
		Clip:       geom.Rect{Min: geom.Pt{X: 50, Y: 100}, Max: geom.Pt{X: 250, Y: 300}},
		FigureRect: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 400, Y: 500}},
	}

	if got := ctx.TransData().Apply(geom.Pt{X: 2.5, Y: 0}); got != (geom.Pt{X: 100, Y: 200}) {
		t.Fatalf("transData point = %+v, want {100 200}", got)
	}
	// Display space is y-up: fraction (0,1) maps to (Min.Y, Max.Y).
	if got := ctx.TransAxes().Apply(geom.Pt{X: 0.25, Y: 0.75}); got != (geom.Pt{X: 100, Y: 250}) {
		t.Fatalf("transAxes point = %+v, want {100 250}", got)
	}
	if got := ctx.TransFigure().Apply(geom.Pt{X: 0.25, Y: 0.75}); got != (geom.Pt{X: 100, Y: 375}) {
		t.Fatalf("transFigure point = %+v, want {100 375}", got)
	}
	if got := ctx.TransformFor(BlendCoords(CoordFigure, CoordAxes)).Apply(geom.Pt{X: 0.5, Y: 0.25}); got != (geom.Pt{X: 200, Y: 150}) {
		t.Fatalf("blended transform point = %+v, want {200 150}", got)
	}
}

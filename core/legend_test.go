package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestLegendCollectEntries(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "line"})
	ax.Scatter([]float64{0.5}, []float64{0.5}, ScatterOptions{Label: "points"})
	ax.Bar([]float64{1}, []float64{2}, BarOptions{Label: "bars"})
	ax.Plot([]float64{0, 1}, []float64{1, 0})

	legend := ax.AddLegend()
	entries := legend.collectEntries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 legend entries, got %d", len(entries))
	}

	if entries[0].Label != "line" || entries[0].kind != legendEntryLine {
		t.Fatalf("unexpected first legend entry: %+v", entries[0])
	}
	if entries[1].Label != "points" || entries[1].kind != legendEntryMarker {
		t.Fatalf("unexpected second legend entry: %+v", entries[1])
	}
	if entries[2].Label != "bars" || entries[2].kind != legendEntryPatch {
		t.Fatalf("unexpected third legend entry: %+v", entries[2])
	}
}

func TestLegendCollectsLineMarkers(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	marker := MarkerDiamond
	face := render.Color{R: 1, A: 0.7}
	edge := render.Color{B: 1, A: 0.5}
	edgeWidth := 2.0
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{
		Label:           "line markers",
		Marker:          &marker,
		MarkerFaceColor: &face,
		MarkerEdgeColor: &edge,
		MarkerEdgeWidth: &edgeWidth,
	})

	entries := ax.AddLegend().collectEntries()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	entry := entries[0]
	if entry.kind != legendEntryLine || !entry.lineMarkerSet {
		t.Fatalf("legend entry should be combined line marker, got %+v", entry)
	}
	if entry.marker != marker || entry.markerFill != face || entry.markerEdge != edge || entry.markerEdgeWidth != pointsToPixels(style.Default, edgeWidth) {
		t.Fatalf("legend marker metadata = %+v", entry)
	}
}

func TestLegendDrawRendersLabelsAndSamples(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	ax.Scatter([]float64{0.5}, []float64{0.5}, ScatterOptions{Label: "samples"})
	ax.AddLegend()

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "signal") || !containsString(r.texts, "samples") {
		t.Fatalf("unexpected legend labels: %v", r.texts)
	}
	if r.pathCount < 4 {
		t.Fatalf("expected legend to draw box and sample paths, got %d paths", r.pathCount)
	}
}

func TestLegendDrawSupportsMultipleColumns(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "a"})
	ax.Plot([]float64{0, 1}, []float64{1, 2}, PlotOptions{Label: "b"})
	ax.Plot([]float64{0, 1}, []float64{2, 3}, PlotOptions{Label: "c"})
	ax.Plot([]float64{0, 1}, []float64{3, 4}, PlotOptions{Label: "d"})
	legend := ax.AddLegend()
	legend.Location = LegendUpperLeft
	legend.NumColumns = 2

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	a := r.textOrigin("a")
	b := r.textOrigin("b")
	c := r.textOrigin("c")
	d := r.textOrigin("d")
	if !floatApprox(a.X, b.X, 1e-9) {
		t.Fatalf("first column labels should share x origin, got a=%+v b=%+v", a, b)
	}
	if !floatApprox(c.X, d.X, 1e-9) || c.X <= a.X {
		t.Fatalf("second column labels should share a later x origin, got a=%+v c=%+v d=%+v", a, c, d)
	}
	if !floatApprox(a.Y, c.Y, 1e-9) || !floatApprox(b.Y, d.Y, 1e-9) {
		t.Fatalf("multi-column rows should align, got a=%+v b=%+v c=%+v d=%+v", a, b, c, d)
	}
}

func TestLegendMarkerSampleScaleAndScatterPoints(t *testing.T) {
	entry := legendEntryFromMarker("points", MarkerCircle, geom.Path{}, render.Color{A: 1}, render.Color{A: 1}, 1)
	sample := geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 70, Y: 30}}

	var base legendRecordingRenderer
	(&Legend{}).drawSample(&base, entry, sample)
	if got := len(base.paths); got != 1 {
		t.Fatalf("default marker legend sample paths = %d, want 1", got)
	}
	baseBounds := pathBoundsForLegendTest(base.paths[0])

	var scaled legendRecordingRenderer
	(&Legend{MarkerScale: 2, ScatterPoints: 3}).drawSample(&scaled, entry, sample)
	if got := len(scaled.paths); got != 3 {
		t.Fatalf("scaled scatter legend sample paths = %d, want 3", got)
	}
	scaledBounds := pathBoundsForLegendTest(scaled.paths[0])
	if scaledBounds.W() <= baseBounds.W()*1.5 {
		t.Fatalf("scaled marker width = %g, want larger than default width %g", scaledBounds.W(), baseBounds.W())
	}
	if !(pathCenterX(scaled.paths[0]) < pathCenterX(scaled.paths[1]) && pathCenterX(scaled.paths[1]) < pathCenterX(scaled.paths[2])) {
		t.Fatalf("scatter sample marker centers should advance left-to-right: %+v", scaled.paths)
	}
}

func TestLegendDrawsErrorBarSampleWithCaps(t *testing.T) {
	entry, ok := (&ErrorBar{
		Label:     "errs",
		YErr:      []float64{0.2},
		CapSize:   6,
		Color:     render.Color{R: 0.1, G: 0.2, B: 0.7, A: 1},
		LineWidth: 2,
	}).legendEntry()
	if !ok {
		t.Fatal("ErrorBar legendEntry returned !ok")
	}

	var r legendRecordingRenderer
	(&Legend{}).drawSample(&r, entry, geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 70, Y: 30}})

	if !containsVerticalLegendPath(r.paths) {
		t.Fatalf("errorbar legend sample should include vertical error stem, got paths %+v", r.paths)
	}
	if countHorizontalLegendSegments(r.paths) < 3 {
		t.Fatalf("errorbar legend sample should include line and two caps, got paths %+v", r.paths)
	}
}

func TestLegendCollectsStemAsSingleCombinedSample(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Stem([]float64{0, 1, 2}, []float64{1, 3, 2}, StemOptions{Label: "stem"})

	entries := ax.AddLegend().collectEntries()
	if len(entries) != 1 {
		t.Fatalf("stem legend entries = %d, want one combined sample: %+v", len(entries), entries)
	}
	entry := entries[0]
	if entry.Label != "stem" || entry.kind != legendEntryErrorBar || !entry.errorbarY || !entry.lineMarkerSet {
		t.Fatalf("stem legend entry = %+v, want combined stem line+marker sample", entry)
	}
}

func TestLegendDrawSupportsTitle(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	legend := ax.AddLegend()
	legend.Location = LegendUpperLeft
	legend.Title = "Series"

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "Series") || !containsString(r.texts, "signal") {
		t.Fatalf("legend title and label should be drawn, got %v", r.texts)
	}
	title := r.textOrigin("Series")
	label := r.textOrigin("signal")
	// Display space is y-up: the title sits above the first entry at a larger Y.
	if title.Y <= label.Y {
		t.Fatalf("legend title should be above first entry label, got title=%+v label=%+v", title, label)
	}

	withoutTitle := NewLegend(ax)
	withTitle := NewLegend(ax)
	withTitle.Title = "Series"
	boxWithout, okWithout := withoutTitle.boxRect(&r, &DrawContext{RC: fig.RC, Clip: geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 400, Y: 300}}})
	boxWith, okWith := withTitle.boxRect(&r, &DrawContext{RC: fig.RC, Clip: geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 400, Y: 300}}})
	if !okWithout || !okWith {
		t.Fatalf("expected legend boxes for titled and untitled legends, got %v %v", okWithout, okWith)
	}
	if boxWith.H() <= boxWithout.H() {
		t.Fatalf("titled legend height = %g, want larger than untitled height %g", boxWith.H(), boxWithout.H())
	}
}

func TestLegendFrameOnFalseSkipsFrameOnly(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	legend := ax.AddLegend()
	legend.Location = LegendUpperLeft
	legend.FrameOn = false

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "signal") {
		t.Fatalf("legend label should still be drawn when frame is disabled, got %v", r.texts)
	}
	if r.hasLegendFramePaint(legend) {
		t.Fatalf("legend frame paint should not be drawn when FrameOn is false")
	}
	if len(r.paths) == 0 {
		t.Fatalf("legend samples should still be drawn when frame is disabled")
	}
}

func TestLegendAddEntryDrawsProxyPatchSample(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	legend := ax.AddLegend()
	proxyFill := render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1}
	proxyEdge := render.Color{R: 0.05, G: 0.1, B: 0.2, A: 1}
	legend.AddEntry("proxy", LegendEntryOptions{
		Sample:    LegendSamplePatch,
		FaceColor: proxyFill,
		EdgeColor: proxyEdge,
		EdgeWidth: 2,
	})

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "proxy") {
		t.Fatalf("explicit proxy legend entry should be drawn, got labels %v", r.texts)
	}
	if !r.hasFillColor(proxyFill) {
		t.Fatalf("explicit proxy legend patch sample should use fill color %+v, got paints %+v", proxyFill, r.paints)
	}
}

func TestLegendSetHandlerOverridesCollectedArtistSample(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	line := ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "custom"})
	legend := ax.AddLegend()
	overrideFill := render.Color{R: 0.7, G: 0.2, B: 0.1, A: 1}
	legend.SetHandler(line, LegendEntryOptions{
		Sample:    LegendSamplePatch,
		FaceColor: overrideFill,
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 1,
	})

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "custom") {
		t.Fatalf("legend should still collect the artist label, got labels %v", r.texts)
	}
	if !r.hasFillColor(overrideFill) {
		t.Fatalf("custom legend handler sample should use fill color %+v, got paints %+v", overrideFill, r.paints)
	}
}

func TestLegendBestPlacementAvoidsLineAndScatterPoints(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	line := &Line2D{
		XY:    []geom.Pt{{X: 0.86, Y: 0.94}, {X: 0.96, Y: 0.94}},
		Label: "line",
	}
	scatter := &Scatter2D{
		XY:    []geom.Pt{{X: 0.9, Y: 0.95}},
		Label: "scatter",
	}
	ax.Artists = []Artist{line, scatter, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendUpperLeft, legend.Inset)
	if box != want {
		t.Fatalf("best legend box = %+v, want upper-left %+v when line/scatter occupy upper-right", box, want)
	}
}

func TestLegendBestPlacementAvoidsImageExtent(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	image := &Image2D{
		XMin: 0.85,
		XMax: 0.95,
		YMin: 0.85,
		YMax: 0.95,
	}
	ax.Artists = []Artist{image, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendUpperLeft, legend.Inset)
	if box != want {
		t.Fatalf("best legend box = %+v, want upper-left %+v when image occupies upper-right", box, want)
	}
}

func TestLegendBestPlacementAvoidsAnnotationAnchors(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	annotation := &Annotation{
		Point:   geom.Pt{X: 0.9, Y: 0.95},
		Coords:  Coords(CoordData),
		OffsetX: 8,
		OffsetY: 8,
	}
	boxPosition := geom.Pt{X: 0.9, Y: 0.95}
	annotationBox := &AnnotationBbox{
		Point:       geom.Pt{X: 0.88, Y: 0.92},
		XYCoords:    Coords(CoordData),
		BoxCoords:   Coords(CoordData),
		BoxPosition: &boxPosition,
	}
	ax.Artists = []Artist{annotation, annotationBox, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendUpperLeft, legend.Inset)
	if box != want {
		t.Fatalf("best legend box = %+v, want upper-left %+v when annotations occupy upper-right", box, want)
	}
}

func TestLegendDefaultsMatchMatplotlibSpacing(t *testing.T) {
	fig := NewFigure(800, 600)
	legend := fig.AddLegend()
	fontPx := pointsToPixels(fig.RC, fig.RC.LegendSize())

	if !floatApprox(legend.Padding, 0.4*fontPx, 1e-9) {
		t.Fatalf("legend padding = %v, want %v", legend.Padding, 0.4*fontPx)
	}
	if !floatApprox(legend.Inset, 0.5*fontPx, 1e-9) {
		t.Fatalf("legend inset = %v, want %v", legend.Inset, 0.5*fontPx)
	}
	if !floatApprox(legend.SampleWidth, 2.0*fontPx, 1e-9) {
		t.Fatalf("legend sample width = %v, want %v", legend.SampleWidth, 2.0*fontPx)
	}
	if !floatApprox(legend.SampleTextGap, 0.8*fontPx, 1e-9) {
		t.Fatalf("legend sample-text gap = %v, want %v", legend.SampleTextGap, 0.8*fontPx)
	}
	if legend.CornerRadius <= 0 {
		t.Fatalf("legend corner radius = %v, want rounded Matplotlib fancybox", legend.CornerRadius)
	}
}

func legendBestPlacementTestContext() *DrawContext {
	return &DrawContext{
		DataToPixel: Transform2D{
			XScale: transform.NewLinear(0, 1),
			YScale: transform.NewLinear(0, 1),
			// Display space is y-up: data y grows upward (no device flip here; the
			// backend owns that), so data (0.9,0.95) lands at the upper-right corner.
			AxesToPixel: transform.NewAffine(geom.Affine{A: 500, D: 500, F: 0}),
		},
		RC:   style.Default,
		Clip: geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 500, Y: 500}},
	}
}

type legendRecordingRenderer struct {
	render.NullRenderer
	pathCount   int
	paths       []geom.Path
	paints      []render.Paint
	texts       []string
	textOrigins map[string]geom.Pt
}

func (r *legendRecordingRenderer) Path(path geom.Path, paint *render.Paint) {
	r.pathCount++
	r.paths = append(r.paths, path)
	if paint == nil {
		r.paints = append(r.paints, render.Paint{})
		return
	}
	r.paints = append(r.paints, *paint)
}

func (r *legendRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *legendRecordingRenderer) DrawText(text string, origin geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	if r.textOrigins == nil {
		r.textOrigins = map[string]geom.Pt{}
	}
	r.textOrigins[text] = origin
}

func (r *legendRecordingRenderer) textOrigin(text string) geom.Pt {
	if r.textOrigins == nil {
		return geom.Pt{}
	}
	return r.textOrigins[text]
}

func (r *legendRecordingRenderer) hasLegendFramePaint(legend *Legend) bool {
	for _, paint := range r.paints {
		if paint.Fill == legend.BackgroundColor && paint.Stroke == legend.BorderColor && paint.LineWidth == legend.BorderWidth {
			return true
		}
	}
	return false
}

func (r *legendRecordingRenderer) hasFillColor(color render.Color) bool {
	for _, paint := range r.paints {
		if paint.Fill == color {
			return true
		}
	}
	return false
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func pathBoundsForLegendTest(path geom.Path) geom.Rect {
	if len(path.V) == 0 {
		return geom.Rect{}
	}
	bounds := geom.Rect{Min: path.V[0], Max: path.V[0]}
	for _, pt := range path.V[1:] {
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

func pathCenterX(path geom.Path) float64 {
	bounds := pathBoundsForLegendTest(path)
	return (bounds.Min.X + bounds.Max.X) / 2
}

func containsVerticalLegendPath(paths []geom.Path) bool {
	for _, path := range paths {
		if len(path.V) == 2 && floatApprox(path.V[0].X, path.V[1].X, 1e-9) && !floatApprox(path.V[0].Y, path.V[1].Y, 1e-9) {
			return true
		}
	}
	return false
}

func countHorizontalLegendSegments(paths []geom.Path) int {
	count := 0
	for _, path := range paths {
		if len(path.V) == 2 && floatApprox(path.V[0].Y, path.V[1].Y, 1e-9) && !floatApprox(path.V[0].X, path.V[1].X, 1e-9) {
			count++
		}
	}
	return count
}

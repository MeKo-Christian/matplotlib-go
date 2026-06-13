package core

import (
	"math"
	"reflect"
	"testing"

	matcolor "github.com/cwbudde/matplotlib-go/color"
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

func TestLegendLineMarkerSampleUsesOriginalMarkerSize(t *testing.T) {
	line := &Line2D{
		Label:      "line markers",
		Col:        render.Color{A: 1},
		W:          1.5,
		Marker:     MarkerCircle,
		MarkerSet:  true,
		MarkerSize: 12,
	}
	entry, ok := line.legendEntry()
	if !ok {
		t.Fatal("line marker legend entry not collected")
	}

	legend := NewLegend(nil)
	var r legendRecordingRenderer
	legend.drawSample(&r, entry, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 40, Y: 20},
	})

	if len(r.paths) < 2 {
		t.Fatalf("legend sample paths = %d, want line plus marker", len(r.paths))
	}
	markerBounds, ok := r.paths[1].Bounds()
	if !ok {
		t.Fatal("legend marker path has no bounds")
	}
	want := pointsToPixels(style.Default, line.MarkerSize)
	if got := markerBounds.W(); !floatApprox(got, want, 1e-9) {
		t.Fatalf("legend marker diameter = %v, want original Line2D markersize %v", got, want)
	}
}

func TestLegendLineMarkerSampleCopiesMarkerStrokeStyle(t *testing.T) {
	tuple := NewTupleMarkerStyle(5, MarkerTupleStar, 18)
	line := &Line2D{
		Label:       "tuple star",
		Col:         render.Color{R: 0.5, A: 1},
		W:           1.4,
		MarkerStyle: tuple,
		MarkerSize:  11,
	}
	entry, ok := line.legendEntry()
	if !ok {
		t.Fatal("line marker legend entry not collected")
	}

	legend := NewLegend(nil)
	var r legendRecordingRenderer
	legend.drawSample(&r, entry, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 40, Y: 20},
	})

	if len(r.paints) < 2 {
		t.Fatalf("legend sample paints = %d, want line plus marker", len(r.paints))
	}
	markerPaint := r.paints[1]
	if markerPaint.LineJoin != render.JoinBevel {
		t.Fatalf("legend tuple-star marker join = %v, want Matplotlib bevel join", markerPaint.LineJoin)
	}
	if markerPaint.LineCap != render.CapButt {
		t.Fatalf("legend tuple-star marker cap = %v, want Matplotlib butt cap", markerPaint.LineCap)
	}
}

func TestLegendLineMarkerSampleCopiesMatplotlibSnapPolicy(t *testing.T) {
	for _, tc := range []struct {
		name   string
		marker MarkerType
		want   render.SnapMode
	}{
		{name: "square", marker: MarkerSquare, want: render.SnapAuto},
		{name: "circle", marker: MarkerCircle, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			line := &Line2D{
				Label:      tc.name,
				Col:        render.Color{A: 1},
				W:          1.5,
				Marker:     tc.marker,
				MarkerSet:  true,
				MarkerSize: 9,
			}
			entry, ok := line.legendEntry()
			if !ok {
				t.Fatal("line marker legend entry not collected")
			}

			legend := NewLegend(nil)
			var r legendRecordingRenderer
			legend.drawSample(&r, entry, geom.Rect{
				Min: geom.Pt{X: 0, Y: 0},
				Max: geom.Pt{X: 40, Y: 20},
			})

			if len(r.paints) < 2 {
				t.Fatalf("legend sample paints = %d, want line plus marker", len(r.paints))
			}
			if got := r.paints[1].Snap; got != tc.want {
				t.Fatalf("legend %s marker snap = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestLegendLineMathTextMarkerDefersPathToRenderer(t *testing.T) {
	mathMarker := NewMathTextMarkerStyle("$f$")
	line := &Line2D{
		Label:       "mathtext",
		Col:         render.Color{A: 1},
		W:           1.5,
		MarkerStyle: mathMarker,
		MarkerSize:  12,
	}
	entry, ok := line.legendEntry()
	if !ok {
		t.Fatal("line marker legend entry not collected")
	}
	if entry.markerStyle.MathText != "$f$" {
		t.Fatalf("legend marker style = %+v, want mathtext marker retained", entry.markerStyle)
	}
	if len(entry.markerPath.C) != 0 {
		t.Fatalf("mathtext legend marker path was resolved without a renderer; commands=%d", len(entry.markerPath.C))
	}
}

func TestLegendLineSampleCopiesLine2DStrokeCaps(t *testing.T) {
	line := &Line2D{
		Label: "line",
		Col:   render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1},
		W:     2,
	}
	entry, ok := line.legendEntry()
	if !ok {
		t.Fatal("line legend entry not collected")
	}

	legend := NewLegend(nil)
	var r legendRecordingRenderer
	legend.drawSample(&r, entry, geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 40, Y: 32},
	})

	if len(r.paints) == 0 {
		t.Fatal("legend sample did not draw a line")
	}
	if got := r.paints[0].LineCap; got != render.CapButt {
		t.Fatalf("legend line cap = %v, want Line2D default %v", got, render.CapButt)
	}
	if got := r.paints[0].LineJoin; got != render.JoinRound {
		t.Fatalf("legend line join = %v, want Line2D default %v", got, render.JoinRound)
	}
}

func TestLegendErrorBarMarkerSampleUsesOriginalMarkerSize(t *testing.T) {
	errBar := &ErrorBar{
		Label:      "errorbar markers",
		Color:      render.Color{A: 1},
		LineWidth:  1.5,
		Marker:     MarkerSquare,
		MarkerSet:  true,
		MarkerSize: 12,
	}
	entry, ok := errBar.legendEntry()
	if !ok {
		t.Fatal("errorbar legend entry not collected")
	}
	if !entry.lineMarkerSet {
		t.Fatalf("errorbar legend entry should have marker metadata: %+v", entry)
	}

	legend := NewLegend(nil)
	var r legendRecordingRenderer
	legend.drawSample(&r, entry, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 40, Y: 20},
	})

	if len(r.paths) < 2 {
		t.Fatalf("legend sample paths = %d, want errorbar sample plus marker", len(r.paths))
	}
	markerBounds, ok := r.paths[len(r.paths)-1].Bounds()
	if !ok {
		t.Fatal("legend marker path has no bounds")
	}
	want := pointsToPixels(style.Default, errBar.MarkerSize)
	if got := markerBounds.W(); !floatApprox(got, want, 1e-9) {
		t.Fatalf("legend errorbar marker diameter = %v, want original markersize %v", got, want)
	}
}

func TestLegendErrorBarMarkerEdgeWidthUsesMarkerDefault(t *testing.T) {
	errBar := &ErrorBar{
		Label:      "errorbar markers",
		Color:      render.Color{A: 1},
		LineWidth:  2.0,
		Marker:     MarkerSquare,
		MarkerSet:  true,
		MarkerSize: 5,
	}
	entry, ok := errBar.legendEntry()
	if !ok {
		t.Fatal("errorbar legend entry not collected")
	}

	want := pointsToPixels(style.Default, 1)
	if !floatApprox(entry.markerEdgeWidth, want, 1e-9) {
		t.Fatalf("legend errorbar marker edge width = %v, want Matplotlib lines.markeredgewidth 1 pt = %v px", entry.markerEdgeWidth, want)
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

func TestLegendScatterSampleCentersUseMatplotlibOffsets(t *testing.T) {
	legend := NewLegend(nil)
	legend.ScatterPoints = 3
	sample := geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 50, Y: 40}}
	center := geom.Pt{X: 30, Y: 30}

	got := legend.markerSampleCenters(sample, center)
	want := []geom.Pt{
		{X: 16, Y: 28.25},
		{X: 30, Y: 30.00},
		{X: 44, Y: 27.375},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scatter sample centers = %+v, want Matplotlib offsets %+v", got, want)
	}
}

func TestLegendScatterSampleUsesSourceCollectionSize(t *testing.T) {
	scatter := &Scatter2D{
		Label:     "scatter",
		Size:      36,
		Color:     render.Color{A: 1},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 1,
		Marker:    MarkerCircle,
	}
	entry, ok := scatter.legendEntry()
	if !ok {
		t.Fatal("scatter legend entry not collected")
	}

	legend := NewLegend(nil)
	legend.MarkerScale = 1.8
	var r legendRecordingRenderer
	legend.drawSample(&r, entry, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 40, Y: 20},
	})

	if len(r.paths) != 1 {
		t.Fatalf("legend scatter sample paths = %d, want one marker", len(r.paths))
	}
	markerBounds, ok := r.paths[0].Bounds()
	if !ok {
		t.Fatal("legend scatter marker path has no bounds")
	}
	want := pointsToPixels(style.Default, math.Sqrt(scatter.Size)) * legend.MarkerScale
	if got := markerBounds.W(); !floatApprox(got, want, 1e-9) {
		t.Fatalf("legend scatter marker diameter = %v, want source collection size %v", got, want)
	}
}

func TestLegendScatterSampleUsesMatplotlibVariableSizeRepresentative(t *testing.T) {
	scatter := &Scatter2D{
		Label:     "scatter",
		Sizes:     []float64{16, 100, 36},
		Color:     render.Color{A: 1},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 1,
		Marker:    MarkerCircle,
	}
	entry, ok := scatter.legendEntry()
	if !ok {
		t.Fatal("scatter legend entry not collected")
	}

	legend := NewLegend(nil)
	var r legendRecordingRenderer
	legend.drawSample(&r, entry, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 40, Y: 20},
	})

	if len(r.paths) != 1 {
		t.Fatalf("legend scatter sample paths = %d, want one marker", len(r.paths))
	}
	markerBounds, ok := r.paths[0].Bounds()
	if !ok {
		t.Fatal("legend scatter marker path has no bounds")
	}
	wantArea := 0.5 * (16.0 + 100.0)
	want := pointsToPixels(style.Default, math.Sqrt(wantArea))
	if got := markerBounds.W(); !floatApprox(got, want, 1e-9) {
		t.Fatalf("legend scatter variable-size marker diameter = %v, want Matplotlib representative size %v", got, want)
	}
}

func TestLegendMathLabelWidthUsesMeasuredTextWidth(t *testing.T) {
	layout := singleLineTextLayout{
		TextLineLayout: render.TextLineLayout{Width: 81},
		MathLayout:     &MathTextLayout{},
	}

	if got, want := legendLabelWidth(layout), layout.Width; got != want {
		t.Fatalf("math legend label width = %v, want measured width %v", got, want)
	}
}

func TestLegendDrawKeepsCollectionOrderAfterZSorting(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "line"})
	ax.Scatter([]float64{0.5}, []float64{0.5}, ScatterOptions{Label: "scatter"})
	ax.Plot([]float64{0, 1}, []float64{1, 2}, PlotOptions{Label: "handler"})
	legend := ax.AddLegend()
	legend.Location = LegendUpperLeft
	legend.NumColumns = 2

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	entries := legend.collectEntries()
	labels := make([]string, len(entries))
	for i, entry := range entries {
		labels[i] = entry.Label
	}
	want := []string{"line", "scatter", "handler"}
	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("legend collection order after draw = %v, want insertion order %v", labels, want)
	}
}

func TestLegendCollectsErrorBarsAfterPlainArtistsLikeMatplotlibContainers(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "line"})
	ax.ErrorBar([]float64{0.5}, []float64{0.5}, nil, []float64{0.1}, ErrorBarOptions{Label: "errorbar"})
	ax.Plot([]float64{0, 1}, []float64{1, 2}, PlotOptions{Label: "handler"})

	entries := ax.AddLegend().collectEntries()
	labels := make([]string, len(entries))
	for i, entry := range entries {
		labels[i] = entry.Label
	}
	want := []string{"line", "handler", "errorbar"}
	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("legend collection order = %v, want Matplotlib child/container order %v", labels, want)
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
	wantCenters := []float64{19.5, 40.5, 61.5}
	for i, want := range wantCenters {
		if got := pathCenterX(scaled.paths[i]); !floatApprox(got, want, 1e-9) {
			t.Fatalf("scatter sample marker %d center x = %g, want Matplotlib padded position %g", i, got, want)
		}
	}
}

func TestLegendPatchSampleFillsMatplotlibHandleBox(t *testing.T) {
	entry := legendEntryFromPatchStyle(
		"patch",
		render.Color{R: 1, A: 1},
		render.Color{A: 1},
		1,
		"",
		render.Color{},
		0,
	)
	sample := geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 70, Y: 30}}

	var r legendRecordingRenderer
	(&Legend{}).drawSample(&r, entry, sample)
	if got := len(r.paths); got != 1 {
		t.Fatalf("patch legend sample paths = %d, want 1", got)
	}
	bounds := pathBoundsForLegendTest(r.paths[0])
	if !floatApprox(bounds.Min.X, sample.Min.X, 1e-9) || !floatApprox(bounds.Max.X, sample.Max.X, 1e-9) {
		t.Fatalf("patch sample width = [%g, %g], want full handle width [%g, %g]", bounds.Min.X, bounds.Max.X, sample.Min.X, sample.Max.X)
	}
	if !floatApprox(bounds.H(), sample.H()*0.7, 1e-9) {
		t.Fatalf("patch sample height = %g, want Matplotlib handleheight %g", bounds.H(), sample.H()*0.7)
	}
}

func TestLegendPatchSampleDefaultsHatchStyleLikeMatplotlib(t *testing.T) {
	edge := render.Color{R: 0.45, G: 0.30, B: 0.08, A: 1}
	entry := legendEntryFromOptions("proxy", LegendEntryOptions{
		Sample:    LegendSamplePatch,
		FaceColor: render.Color{R: 0.93, G: 0.77, B: 0.33, A: 0.92},
		EdgeColor: edge,
		EdgeWidth: 1.2,
		Hatch:     "xx",
	})

	if entry.patchHatch != "xx" {
		t.Fatalf("patch legend hatch = %q, want xx", entry.patchHatch)
	}
	if entry.patchHatchColor != edge {
		t.Fatalf("patch legend hatch color = %+v, want edge color %+v", entry.patchHatchColor, edge)
	}
	if !floatApprox(entry.patchHatchWidth, 100.0/72.0, 1e-9) {
		t.Fatalf("patch legend hatch linewidth = %g, want Matplotlib default %g", entry.patchHatchWidth, 100.0/72.0)
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
	sample := geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 50, Y: 30}}
	(&Legend{}).drawSample(&r, entry, sample)

	if !containsVerticalLegendPath(r.paths) {
		t.Fatalf("errorbar legend sample should include vertical error stem, got paths %+v", r.paths)
	}
	if countHorizontalLegendSegments(r.paths) < 3 {
		t.Fatalf("errorbar legend sample should include line and two caps, got paths %+v", r.paths)
	}
	if !floatApprox(r.paths[0].V[0].X, sample.Min.X, 1e-9) || !floatApprox(r.paths[0].V[1].X, sample.Max.X, 1e-9) {
		t.Fatalf("errorbar legend line = %+v, want Matplotlib full handle span [%g, %g]", r.paths[0], sample.Min.X, sample.Max.X)
	}
	fontPx := pointsToPixels(style.Default, style.Default.LegendSize())
	centerY := sample.Min.Y + sample.H()/2
	if !floatApprox(r.paths[1].V[0].Y, centerY-fontPx/2, 1e-9) || !floatApprox(r.paths[1].V[1].Y, centerY+fontPx/2, 1e-9) {
		t.Fatalf("errorbar legend stem = %+v, want Matplotlib HandlerErrorbar 0.5-font half extent around %g", r.paths[1], centerY)
	}
	if got := math.Abs(r.paths[2].V[1].X - r.paths[2].V[0].X); !floatApprox(got, 12, 1e-9) {
		t.Fatalf("errorbar legend cap length = %g, want Matplotlib 2*capsize marker length", got)
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

func TestLegendFrameUsesMatplotlibSnapAuto(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	legend := ax.AddLegend()

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	for i, paint := range r.paints {
		if paint.Fill == legend.BackgroundColor && paint.Stroke == legend.BorderColor && paint.LineWidth == legend.BorderWidth {
			if paint.Snap != render.SnapAuto {
				t.Fatalf("legend frame paint %d snap = %v, want Matplotlib SnapAuto", i, paint.Snap)
			}
			return
		}
	}
	t.Fatal("legend frame paint was not drawn")
}

func TestLegendFrameUsesMatplotlibRoundBoxStyle(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	legend := ax.AddLegend()

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	for i, paint := range r.paints {
		if paint.Fill == legend.BackgroundColor && paint.Stroke == legend.BorderColor && paint.LineWidth == legend.BorderWidth {
			if got := countPathCmd(r.paths[i], geom.QuadTo); got != 4 {
				t.Fatalf("legend frame quadratic corners = %d, want Matplotlib BoxStyle.Round with four CURVE3 corners; commands=%v", got, r.paths[i].C)
			}
			return
		}
	}
	t.Fatal("legend frame path was not drawn")
}

func TestLegendSampleCentersUseMatplotlibHandleBaseline(t *testing.T) {
	legend := NewLegend(nil)
	legend.Location = LegendUpperLeft
	legend.entries = []legendEntry{
		legendEntryFromLine("signal", render.Color{R: 0.1, G: 0.2, B: 0.7, A: 1}, 2, nil),
	}
	ctx := &DrawContext{
		RC:   style.Default,
		Clip: geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 300, Y: 200}},
	}
	var r legendRecordingRenderer
	legend.draw(&r, ctx)

	labelOrigin := r.textOrigin("signal")
	fontPx := pointsToPixels(ctx.RC, legend.FontSize)
	wantY := labelOrigin.Y + 0.35*fontPx
	for i, path := range r.paths {
		if i >= len(r.paints) || r.paints[i].Stroke != legend.entries[0].lineColor || len(path.V) < 2 {
			continue
		}
		gotY := path.V[0].Y
		if !floatApprox(gotY, wantY, 1e-9) {
			t.Fatalf("legend sample center Y = %v, want Matplotlib baseline + 0.35*fontsize = %v", gotY, wantY)
		}
		return
	}
	t.Fatalf("legend line sample was not drawn; paths=%+v", r.paths)
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

func TestLegendCollectionSamplesUseScalarMappedColors(t *testing.T) {
	cmapName := "legend-collection-scalar"
	low := render.Color{R: 1, A: 1}
	high := render.Color{B: 1, A: 1}
	matcolor.RegisterColormap(cmapName, matcolor.NewColormap(cmapName, []matcolor.ColorStop{
		{Pos: 0, Color: low},
		{Pos: 1, Color: high},
	}))

	pathCollection := &PathCollection{
		Collection: Collection{
			Label:        "mapped",
			Colormap:     cmapName,
			VMin:         0,
			VMax:         10,
			ScalarValues: []float64{0, 10},
		},
		Path:      markerCirclePath(1),
		Offsets:   []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 1,
	}
	pathEntry, ok := pathCollection.legendEntry()
	if !ok {
		t.Fatal("scalar-mapped path collection legend entry not collected")
	}
	if got := pathEntry.markerFill; got != low {
		t.Fatalf("legend marker fill = %+v, want first scalar-mapped face %+v", got, low)
	}

	explicit := render.Color{G: 1, A: 1}
	pathCollection.FaceColor = explicit
	pathEntry, ok = pathCollection.legendEntry()
	if !ok {
		t.Fatal("explicit-colored path collection legend entry not collected")
	}
	if got := pathEntry.markerFill; got != explicit {
		t.Fatalf("legend marker fill with explicit face = %+v, want explicit color %+v", got, explicit)
	}
	pathCollection.FaceColor = render.Color{}

	patchCollection := &PatchCollection{
		Collection: Collection{
			Label:        "mapped patch",
			Colormap:     cmapName,
			VMin:         0,
			VMax:         10,
			ScalarValues: []float64{10},
		},
		Paths:     []geom.Path{polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true)},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 1,
	}
	patchEntry, ok := patchCollection.legendEntry()
	if !ok {
		t.Fatal("scalar-mapped patch collection legend entry not collected")
	}
	if got := patchEntry.patchFill; got != high {
		t.Fatalf("legend patch fill = %+v, want first scalar-mapped face %+v", got, high)
	}

	lineCollection := &LineCollection{
		Collection: Collection{
			Label:        "mapped line",
			Colormap:     cmapName,
			VMin:         0,
			VMax:         10,
			ScalarValues: []float64{10},
		},
		Segments:  [][]geom.Pt{{{X: 0, Y: 0}, {X: 1, Y: 1}}},
		LineWidth: 1,
	}
	lineEntry, ok := lineCollection.legendEntry()
	if !ok {
		t.Fatal("scalar-mapped line collection legend entry not collected")
	}
	if got := lineEntry.lineColor; got != high {
		t.Fatalf("legend line color = %+v, want first scalar-mapped stroke %+v", got, high)
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
	if !approxRect(box, want, 1e-9) {
		t.Fatalf("best legend box = %+v, want upper-left %+v when line/scatter occupy upper-right", box, want)
	}
}

func TestLegendBestPlacementCountsLineIntersections(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	crossingLowerLeft := &Line2D{
		XY:    []geom.Pt{{X: -0.1, Y: 0.06}, {X: 0.3, Y: 0.06}},
		Label: "crossing",
	}
	pointsInOtherCorners := &Scatter2D{
		XY: []geom.Pt{
			{X: 0.9, Y: 0.95},
			{X: 0.1, Y: 0.95},
			{X: 0.9, Y: 0.06},
		},
		Label: "points",
	}
	ax.Artists = []Artist{crossingLowerLeft, pointsInOtherCorners, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendRight, legend.Inset)
	if !approxRect(box, want, 1e-9) {
		t.Fatalf("best legend box = %+v, want center-right %+v when lower-left is crossed by a line segment", box, want)
	}
}

func TestLegendBestPlacementMatchesMathtextInlineLabels(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.92, Y: 0.88},
	})
	ax.SetXLim(0, 5)
	ax.SetYLim(0.165, 0.925)

	const n = 90
	x := make([]float64, n)
	y1 := make([]float64, n)
	y2 := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1) * 5
		x[i] = t
		y1[i] = 0.55 + 0.35*math.Sin(1.5*t)
		y2[i] = 0.48 + 0.28*math.Cos(1.5*t+0.45)
	}
	ax.Plot(x, y1, PlotOptions{Label: `state $x_i(t)$`})
	ax.Plot(x, y2, PlotOptions{Label: `state $y_i(t)$`})
	ax.Text(0.03, 0.88, `peak $\alpha_i^2$`, TextOptions{
		Coords: Coords(CoordAxes),
		HAlign: TextAlignLeft,
		VAlign: TextVAlignTop,
	})
	ax.Text(0.97, 0.08, `ratio $\frac{a}{b}$`, TextOptions{
		Coords: Coords(CoordAxes),
		HAlign: TextAlignRight,
		VAlign: TextVAlignBottom,
	})
	legend := ax.AddLegend()

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	var r legendRecordingRenderer
	box, ok := legend.boxRect(&r, ctx)
	if !ok {
		t.Fatal("legend.boxRect returned !ok")
	}
	want := anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendLowerLeft, legend.Inset)
	if !approxRect(box, want, 1e-9) {
		data := legend.legendAvoidanceData(ctx)
		t.Fatalf("best legend box = %+v, want lower-left %+v; scores UR=%d UL=%d LL=%d LR=%d UC=%d C=%d lines=%+v",
			box, want,
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendUpperRight, legend.Inset), data),
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendUpperLeft, legend.Inset), data),
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendLowerLeft, legend.Inset), data),
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendLowerRight, legend.Inset), data),
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendUpperCenter, legend.Inset), data),
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendCenter, legend.Inset), data),
			legendDebugPathBounds(data.lines))
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

func TestLegendBestPlacementAvoidsPatchBounds(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	patch := &Rectangle{
		XY:     geom.Pt{X: 0.84, Y: 0.86},
		Width:  0.14,
		Height: 0.12,
		Coords: Coords(CoordData),
	}
	ax.Artists = []Artist{patch, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendUpperLeft, legend.Inset)
	if box != want {
		t.Fatalf("best legend box = %+v, want upper-left %+v when patch bbox occupies upper-right", box, want)
	}
}

func TestLegendBestPlacementAvoidsPatchPaths(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	patch := &PathPatch{
		Path: geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{{X: 0.84, Y: 0.94}, {X: 0.99, Y: 0.94}},
		},
		Coords: Coords(CoordData),
	}
	ax.Artists = []Artist{patch, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendUpperLeft, legend.Inset)
	if box != want {
		t.Fatalf("best legend box = %+v, want upper-left %+v when patch path occupies upper-right", box, want)
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
	if !floatApprox(legend.BorderWidth, pointsToPixels(fig.RC, 1), 1e-9) {
		t.Fatalf("legend border width = %v, want Matplotlib 1 point linewidth %v", legend.BorderWidth, pointsToPixels(fig.RC, 1))
	}
	if legend.CornerRadius <= 0 {
		t.Fatalf("legend corner radius = %v, want rounded Matplotlib fancybox", legend.CornerRadius)
	}
}

func TestAxesLegendDrawsOutsideAxesClip(t *testing.T) {
	fig := NewFigure(240, 240)
	ax := fig.AddPolarAxes(unitRect())
	ax.SetYLim(0, 1)
	color := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	ax.Plot([]float64{0, 1}, []float64{0.2, 0.8}, PlotOptions{
		Color: &color,
		Label: "legend label",
	})
	ax.AddLegend()

	r := &legendClipTrackingRenderer{}
	DrawFigure(fig, r)

	if clipped, ok := r.textClipped["legend label"]; !ok {
		t.Fatalf("legend label was not drawn; saw texts %v", r.texts)
	} else if clipped {
		t.Fatal("legend label was drawn while the axes clip was active")
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

func legendDebugPathBounds(paths []geom.Path) []geom.Rect {
	out := make([]geom.Rect, 0, len(paths))
	for _, path := range paths {
		if bounds, ok := pathBounds(path); ok {
			out = append(out, bounds)
		}
	}
	return out
}

type legendRecordingRenderer struct {
	render.NullRenderer
	pathCount   int
	paths       []geom.Path
	paints      []render.Paint
	texts       []string
	textOrigins map[string]geom.Pt
}

type legendClipTrackingRenderer struct {
	legendRecordingRenderer
	clipStack   []bool
	clipActive  bool
	textClipped map[string]bool
}

func (r *legendClipTrackingRenderer) Save() {
	r.clipStack = append(r.clipStack, r.clipActive)
	r.legendRecordingRenderer.Save()
}

func (r *legendClipTrackingRenderer) Restore() {
	if len(r.clipStack) > 0 {
		r.clipActive = r.clipStack[len(r.clipStack)-1]
		r.clipStack = r.clipStack[:len(r.clipStack)-1]
	}
	r.legendRecordingRenderer.Restore()
}

func (r *legendClipTrackingRenderer) ClipRect(rect geom.Rect) {
	r.clipActive = true
	r.legendRecordingRenderer.ClipRect(rect)
}

func (r *legendClipTrackingRenderer) ClipPath(path geom.Path) {
	r.clipActive = true
	r.legendRecordingRenderer.ClipPath(path)
}

func (r *legendClipTrackingRenderer) DrawText(text string, origin geom.Pt, size float64, color render.Color) {
	r.legendRecordingRenderer.DrawText(text, origin, size, color)
	if r.textClipped == nil {
		r.textClipped = map[string]bool{}
	}
	r.textClipped[text] = r.clipActive
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

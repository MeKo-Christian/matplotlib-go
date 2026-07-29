package core

import (
	"math"
	"reflect"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

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

func TestLegendLineMarkerSampleCentersOnHandleBox(t *testing.T) {
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

	sample := geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 40, Y: 20},
	}
	legend := NewLegend(nil)
	var r legendRecordingRenderer
	legend.drawSample(&r, entry, sample)

	if len(r.paths) < 2 {
		t.Fatalf("legend sample paths = %d, want line plus marker", len(r.paths))
	}
	markerBounds, ok := r.paths[1].Bounds()
	if !ok {
		t.Fatal("legend marker path has no bounds")
	}
	got := geom.Pt{
		X: markerBounds.Min.X + markerBounds.W()/2,
		Y: markerBounds.Min.Y + markerBounds.H()/2,
	}
	want := geom.Pt{
		X: sample.Min.X + sample.W()/2,
		Y: sample.Min.Y + sample.H()/2,
	}
	if !floatApprox(got.X, want.X, 1e-9) || !floatApprox(got.Y, want.Y, 1e-9) {
		t.Fatalf("legend marker center = %+v, want handle center %+v", got, want)
	}
}

func TestLegendLineMarkerSampleUsesMarkerBatchWhenAvailable(t *testing.T) {
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
	var r legendMarkerBatchRecordingRenderer
	legend.drawSample(&r, entry, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 40, Y: 20},
	})

	if len(r.markerBatches) != 1 {
		t.Fatalf("legend marker batches = %d, want 1", len(r.markerBatches))
	}
	batch := r.markerBatches[0]
	if len(batch.Items) != 1 {
		t.Fatalf("legend marker batch items = %d, want 1", len(batch.Items))
	}
	if got, want := batch.Items[0].Offset, (geom.Pt{X: 20, Y: 10}); got != want {
		t.Fatalf("legend marker batch offset = %+v, want handle center %+v", got, want)
	}
	if got, want := batch.Items[0].Transform.A, pointsToPixels(style.Default, line.MarkerSize); !floatApprox(got, want, 1e-9) {
		t.Fatalf("legend marker batch scale = %v, want marker size %v", got, want)
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
	if got := r.paints[0].LineCap; got != render.CapSquare {
		t.Fatalf("legend solid line cap = %v, want Line2D default %v", got, render.CapSquare)
	}
	if got := r.paints[0].LineJoin; got != render.JoinRound {
		t.Fatalf("legend line join = %v, want Line2D default %v", got, render.JoinRound)
	}
}

func TestLegendDashedLineSampleUsesMatplotlibButtCap(t *testing.T) {
	line := &Line2D{
		Label:  "dashed",
		Col:    render.Color{A: 1},
		W:      2,
		Dashes: PixelDashes(4, 2),
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
		t.Fatalf("legend dashed line cap = %v, want Matplotlib dash cap %v", got, render.CapButt)
	}
}

func TestLegendLineSampleUsesMatplotlibAutoSnap(t *testing.T) {
	line := &Line2D{
		Label: "line",
		Col:   render.Color{A: 1},
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
	if got := r.paints[0].Snap; got != render.SnapAuto {
		t.Fatalf("legend line snap = %v, want Matplotlib auto snap %v", got, render.SnapAuto)
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

	want := 1.0 // points; converted at the Paint sink
	if !floatApprox(entry.markerEdgeWidth, want, 1e-9) {
		t.Fatalf("legend errorbar marker edge width = %v, want Matplotlib lines.markeredgewidth 1 pt = %v px", entry.markerEdgeWidth, want)
	}
}

func TestLegendErrorBarCapWidthUsesMarkerDefault(t *testing.T) {
	errBar := &ErrorBar{
		Label:     "errorbar caps",
		Color:     render.Color{A: 1},
		LineWidth: 2.0,
		XErr:      []float64{0.2},
		YErr:      []float64{0.4},
		CapSize:   6,
	}
	entry, ok := errBar.legendEntry()
	if !ok {
		t.Fatal("errorbar legend entry not collected")
	}

	legend := NewLegend(nil)
	var r legendRecordingRenderer
	legend.drawSample(&r, entry, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 40, Y: 20},
	})

	want := pointsToPixels(style.Default, 1)
	for _, idx := range []int{2, 3, 5, 6} {
		if got := r.paints[idx].LineWidth; !floatApprox(got, want, 1e-9) {
			t.Fatalf("legend errorbar cap paint %d line width = %v, want Matplotlib markeredgewidth 1 pt = %v px", idx, got, want)
		}
	}
}

func TestLegendErrorBarSampleUsesMatplotlibAutoSnap(t *testing.T) {
	errBar := &ErrorBar{
		Label:     "errorbar snap",
		Color:     render.Color{A: 1},
		LineWidth: 2.0,
		XErr:      []float64{0.2},
		YErr:      []float64{0.4},
		CapSize:   6,
	}
	entry, ok := errBar.legendEntry()
	if !ok {
		t.Fatal("errorbar legend entry not collected")
	}

	legend := NewLegend(nil)
	var r legendRecordingRenderer
	legend.drawSample(&r, entry, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 40, Y: 20},
	})

	for i, paint := range r.paints {
		if got := paint.Snap; got != render.SnapAuto {
			t.Fatalf("legend errorbar paint %d snap = %v, want Matplotlib auto snap %v", i, got, render.SnapAuto)
		}
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
	// HandlerNpoints.get_xdata: linspace(pad, width-pad, scatterpoints) with
	// pad = 0.3*fontsize = 0.15*width at the default handlelength of 2. No
	// half-pixel offset — HandlerPathCollection places the handle through
	// draw_path_collection, which keeps the subpixel part of every offset.
	wantCenters := []float64{19, 40, 61}
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
		CapSize:   12,
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

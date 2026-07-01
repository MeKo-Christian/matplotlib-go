package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

type recordingRenderer struct {
	render.NullRenderer
	pathCalls []recordedPathCall
}

type recordedPathCall struct {
	path  geom.Path
	paint render.Paint
}

func (r *recordingRenderer) Path(p geom.Path, paint *render.Paint) {
	if paint == nil {
		r.pathCalls = append(r.pathCalls, recordedPathCall{path: p})
		return
	}
	r.pathCalls = append(r.pathCalls, recordedPathCall{
		path:  p,
		paint: *paint,
	})
}

func TestLine2D_EmptyData(t *testing.T) {
	line := &Line2D{
		XY:  []geom.Pt{}, // empty data
		W:   2.0,
		Col: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	// Should not panic with empty data
	var r render.NullRenderer
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	// This should not panic
	line.Draw(&r, ctx)
}

func TestLine2D_DefaultSolidCapstyleMatchesMatplotlib(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		W:   2.0,
		Col: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	r := &recordingRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 10),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	line.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one Path call, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.LineCap != render.CapSquare {
		t.Fatalf("expected Matplotlib projecting solid cap %v, got %v", render.CapSquare, r.pathCalls[0].paint.LineCap)
	}
	if r.pathCalls[0].paint.LineJoin != render.JoinRound {
		t.Fatalf("expected default line join %v, got %v", render.JoinRound, r.pathCalls[0].paint.LineJoin)
	}
	if r.pathCalls[0].paint.Snap != render.SnapAuto {
		t.Fatalf("expected default line snap %v, got %v", render.SnapAuto, r.pathCalls[0].paint.Snap)
	}
}

func TestLine2D_DashedCapstyleMatchesMatplotlib(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		W:      2.0,
		Col:    render.Color{R: 1, G: 0, B: 0, A: 1},
		Dashes: []float64{3, 2},
	}

	r := &recordingRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 10),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	line.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one Path call, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.LineCap != render.CapButt {
		t.Fatalf("expected Matplotlib butt dash cap %v, got %v", render.CapButt, r.pathCalls[0].paint.LineCap)
	}
}

func TestLine2D_ExplicitSolidCapstyleOverridesDefault(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		W:          2.0,
		Col:        render.Color{R: 1, G: 0, B: 0, A: 1},
		LineCap:    render.CapButt,
		LineCapSet: true,
	}

	r := &recordingRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 10),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	line.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one Path call, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.LineCap != render.CapButt {
		t.Fatalf("explicit solid cap = %v, want Matplotlib butt cap override", r.pathCalls[0].paint.LineCap)
	}
}

func TestLine2DArtistAlphaMultipliesStrokeAlpha(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		W:   2,
		Col: render.Color{R: 1, G: 0, B: 0, A: 0.8},
	}
	line.SetAlpha(0.5)

	r := &recordingRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC: style.Default,
	}
	line.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	if got, want := r.pathCalls[0].paint.Stroke.A, 0.4; got != want {
		t.Fatalf("stroke alpha = %v, want %v", got, want)
	}

	line.SetAlpha(0)
	r.pathCalls = nil
	line.Draw(r, ctx)
	if len(r.pathCalls) != 1 {
		t.Fatalf("transparent path calls = %d, want 1", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Stroke.A; got != 0 {
		t.Fatalf("transparent stroke alpha = %v, want 0", got)
	}
}

func TestLine2DDrawsMarkersWithFaceEdgeStyles(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		W:               2,
		Col:             render.Color{R: 1, A: 1},
		Marker:          MarkerSquare,
		MarkerSet:       true,
		MarkerSize:      8,
		MarkerFaceColor: render.Color{G: 1, A: 0.8},
		MarkerEdgeColor: render.Color{B: 1, A: 0.6},
		MarkerEdgeWidth: 3,
	}

	r := &recordingRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC: style.Default,
	}
	line.Draw(r, ctx)

	if len(r.pathCalls) != 3 {
		t.Fatalf("path calls = %d, want line plus two markers", len(r.pathCalls))
	}
	markerPaint := r.pathCalls[1].paint
	if got, want := markerPaint.Fill, line.MarkerFaceColor; got != want {
		t.Fatalf("marker fill = %+v, want %+v", got, want)
	}
	if got, want := markerPaint.Stroke, line.MarkerEdgeColor; got != want {
		t.Fatalf("marker edge = %+v, want %+v", got, want)
	}
	if got, want := markerPaint.LineWidth, pointsToPixels(ctx.RC, line.MarkerEdgeWidth); got != want {
		t.Fatalf("marker edge width = %v, want %v", got, want)
	}
}

func TestLine2DMarkerColorSentinels(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		W:               2,
		Col:             render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1},
		Marker:          MarkerCircle,
		MarkerSet:       true,
		MarkerFaceColor: render.Color{R: 1, A: 0.25},
	}
	line.SetMarkerEdgeColorAuto()

	r := &recordingRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC: style.Default,
	}
	line.Draw(r, ctx)

	if len(r.pathCalls) != 3 {
		t.Fatalf("path calls = %d, want line plus two markers", len(r.pathCalls))
	}
	if got, want := r.pathCalls[1].paint.Stroke, (render.Color{R: 0.2, G: 0.4, B: 0.6, A: 0.25}); got != want {
		t.Fatalf("auto edge color = %+v, want line RGB with face alpha %+v", got, want)
	}

	line.SetMarkerFaceColorNone()
	r.pathCalls = nil
	line.Draw(r, ctx)
	if got := r.pathCalls[1].paint.Fill.A; got != 0 {
		t.Fatalf("marker face alpha = %v, want none", got)
	}
	if got, want := r.pathCalls[1].paint.Stroke.A, 1.0; got != want {
		t.Fatalf("auto edge alpha with face none = %v, want %v", got, want)
	}
}

func TestLine2DLineOnlyMarkerDrawsStrokeOnly(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		W:               2,
		Col:             render.Color{R: 1, A: 1},
		Marker:          MarkerPlus,
		MarkerSet:       true,
		MarkerFaceColor: render.Color{G: 1, A: 1},
	}

	r := &recordingRenderer{}
	line.Draw(r, &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC: style.Default,
	})

	if len(r.pathCalls) != 3 {
		t.Fatalf("path calls = %d, want line plus two markers", len(r.pathCalls))
	}
	if got := r.pathCalls[1].paint.Fill.A; got != 0 {
		t.Fatalf("line-only marker fill alpha = %v, want 0", got)
	}
	if got := r.pathCalls[1].paint.Stroke.A; got <= 0 {
		t.Fatalf("line-only marker stroke alpha = %v, want visible stroke", got)
	}
}

func TestLine2DHalfFilledMarkerDrawsSplitHalvesWithEdges(t *testing.T) {
	markerStyle := NewMarkerStyle(MarkerCircle)
	markerStyle.FillStyle = MarkerFillLeft
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
		},
		Col:             render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1},
		MarkerStyle:     markerStyle,
		MarkerFaceColor: render.Color{R: 1, A: 1},
		MarkerEdgeColor: render.Color{B: 1, A: 1},
		MarkerEdgeWidth: 2,
	}
	line.SetMarkerFaceColorAlt(render.Color{G: 1, A: 0.75})

	r := &recordingRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC: style.Default,
	}
	line.drawMarkers(r, ctx)

	if len(r.pathCalls) != 2 {
		t.Fatalf("path calls = %d, want primary and alternate half markers", len(r.pathCalls))
	}
	if got, want := r.pathCalls[0].paint.Fill, line.MarkerFaceColor; got != want {
		t.Fatalf("primary half fill = %+v, want %+v", got, want)
	}
	if got, want := r.pathCalls[1].paint.Fill, (render.Color{G: 1, A: 0.75}); got != want {
		t.Fatalf("alternate half fill = %+v, want %+v", got, want)
	}
	for i := range r.pathCalls {
		if got, want := r.pathCalls[i].paint.Stroke, line.MarkerEdgeColor; got != want {
			t.Fatalf("half marker %d stroke = %+v, want Matplotlib edge %+v", i, got, want)
		}
		if got := r.pathCalls[i].paint.LineWidth; got != pointsToPixels(ctx.RC, line.MarkerEdgeWidth) {
			t.Fatalf("half marker %d edge width = %v, want %v", i, got, pointsToPixels(ctx.RC, line.MarkerEdgeWidth))
		}
	}
}

func TestLine2DMarkEveryDrawsEveryNthMarker(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
			{X: 2, Y: 0},
			{X: 3, Y: 1},
			{X: 4, Y: 0},
		},
		W:         2,
		Col:       render.Color{A: 1},
		Marker:    MarkerCircle,
		MarkerSet: true,
		MarkEvery: 2,
	}

	r := &recordingRenderer{}
	line.Draw(r, &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 4),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC: style.Default,
	})

	if len(r.pathCalls) != 4 {
		t.Fatalf("path calls = %d, want line plus 3 markers", len(r.pathCalls))
	}
}

func TestLine2DMarkerSizeUsesMatplotlibPointDiameter(t *testing.T) {
	line := &Line2D{MarkerSize: 8}
	ctx := &DrawContext{RC: style.Default}
	ctx.RC.DPI = 144

	got := line.resolvedMarkerSize(ctx)
	want := 8.0 * 144.0 / 72.0
	if got != want {
		t.Fatalf("resolved marker scale = %v, want full Matplotlib point diameter %v", got, want)
	}
}

func TestLine2DDataAccessorsCloneAndMarkStale(t *testing.T) {
	line := &Line2D{}
	x := []float64{0, 1, 2}
	y := []float64{3, 4, 5}
	line.SetData(x, y)

	if !line.Stale() {
		t.Fatal("SetData did not mark line stale")
	}
	x[0] = 99
	y[1] = 88
	gotX, gotY := line.Data()
	if gotX[0] != 0 || gotY[1] != 4 {
		t.Fatalf("SetData reused caller storage, got x=%v y=%v", gotX, gotY)
	}

	gotX[1] = 77
	gotY[2] = 66
	againX, againY := line.Data()
	if againX[1] != 1 || againY[2] != 5 {
		t.Fatalf("Data reused line storage, got x=%v y=%v", againX, againY)
	}

	line.SetStale(false)
	line.SetXData([]float64{10, 11})
	if !line.Stale() {
		t.Fatal("SetXData did not mark line stale")
	}
	gotX, gotY = line.Data()
	if len(gotX) != 2 || gotX[0] != 10 || gotX[1] != 11 || gotY[0] != 3 || gotY[1] != 4 {
		t.Fatalf("SetXData result x=%v y=%v", gotX, gotY)
	}

	line.SetStale(false)
	line.SetYData([]float64{20, 21, 22})
	if !line.Stale() {
		t.Fatal("SetYData did not mark line stale")
	}
	gotX, gotY = line.Data()
	if len(gotY) != 2 || gotX[0] != 10 || gotX[1] != 11 || gotY[0] != 20 || gotY[1] != 21 {
		t.Fatalf("SetYData result x=%v y=%v", gotX, gotY)
	}
}

func TestLine2DInvalidPointsBreakPathAndBounds(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
			{X: math.NaN(), Y: 0},
			{X: 2, Y: -1},
			{X: 3, Y: 0.5},
			{X: math.Inf(1), Y: 2},
			{X: 4, Y: 2},
		},
		W:   2,
		Col: render.Color{A: 1},
	}

	bounds := line.Bounds(nil)
	if bounds.Min.X != 0 || bounds.Min.Y != -1 || bounds.Max.X != 4 || bounds.Max.Y != 2 {
		t.Fatalf("bounds = %+v, want finite data bounds", bounds)
	}

	r := &recordingRenderer{}
	line.Draw(r, &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 4),
			YScale:      transform.NewLinear(-1, 2),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC: style.Default,
	})

	if len(r.pathCalls) != 1 {
		t.Fatalf("path calls = %d, want 1", len(r.pathCalls))
	}
	got := r.pathCalls[0].path.C
	want := []geom.Cmd{geom.MoveTo, geom.LineTo, geom.MoveTo, geom.LineTo, geom.MoveTo}
	if len(got) != len(want) {
		t.Fatalf("path commands = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("path commands = %v, want %v", got, want)
		}
	}
}

func TestLine2DGeoProjectionInterpolatesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(400, 400)
	ax, err := fig.AddAxesProjection(unitRect(), "lambert")
	if err != nil {
		t.Fatalf("AddAxesProjection(lambert): %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	line := &Line2D{
		XY: []geom.Pt{
			{X: -math.Pi / 2, Y: 0},
			{X: math.Pi / 2, Y: 0.35},
		},
	}

	path := line.displayPath(ctx)
	if got, want := len(path.C), geoGridSegments+1; got != want {
		t.Fatalf("geo line command count = %d, want %d after Matplotlib RESOLUTION interpolation", got, want)
	}
	if got, want := len(path.V), geoGridSegments+1; got != want {
		t.Fatalf("geo line vertex count = %d, want %d after Matplotlib RESOLUTION interpolation", got, want)
	}
	tMid := float64(geoGridSegments/2) / float64(geoGridSegments)
	wantMidData := geom.Pt{
		X: -math.Pi/2 + math.Pi*tMid,
		Y: 0.35 * tMid,
	}
	wantMid := ctx.DataToPixel.Apply(wantMidData)
	gotMid := path.V[geoGridSegments/2]
	if !approx(gotMid.X, wantMid.X, 1e-9) || !approx(gotMid.Y, wantMid.Y, 1e-9) {
		t.Fatalf("geo line midpoint = %+v, want projection of interpolated data midpoint %+v", gotMid, wantMid)
	}
}

func TestLine2DGapColorDrawsInverseDashPass(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
		},
		W:   2,
		Col: render.Color{B: 1, A: 1},
	}
	line.SetDashes(4, 2, 1, 3)
	line.SetGapColor(render.Color{R: 1, A: 0.5})

	r := &recordingRenderer{}
	line.Draw(r, &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Affine{A: 100, D: 100}),
		},
		RC: style.Default,
	})

	if len(r.pathCalls) != 2 {
		t.Fatalf("path calls = %d, want gap pass plus line pass", len(r.pathCalls))
	}
	if got, want := r.pathCalls[0].paint.Stroke, (render.Color{R: 1, A: 0.5}); got != want {
		t.Fatalf("gap stroke = %+v, want %+v", got, want)
	}
	if got := r.pathCalls[0].paint.Dashes; len(got) != 0 {
		t.Fatalf("gap pass should draw extracted gap path without dashes, got %v", got)
	}
	if got := r.pathCalls[0].path.C; len(got) == 0 || got[0] != geom.MoveTo {
		t.Fatalf("gap path commands = %v, want extracted path", got)
	}
	if got, want := r.pathCalls[1].paint.Dashes, []float64{8 * 100.0 / 72.0, 4 * 100.0 / 72.0, 2 * 100.0 / 72.0, 6 * 100.0 / 72.0}; len(got) != len(want) || !floatApprox(got[0], want[0], 1e-9) || !floatApprox(got[1], want[1], 1e-9) || !floatApprox(got[2], want[2], 1e-9) || !floatApprox(got[3], want[3], 1e-9) {
		t.Fatalf("line dashes = %v, want %v", got, want)
	}
}

func TestLine2DMarkEverySpecForms(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
			{X: 2, Y: 0},
			{X: 3, Y: 1},
			{X: 4, Y: 0},
			{X: 5, Y: 1},
		},
	}

	line.SetMarkEvery(StartStepMarkers(1, 2))
	if got := line.markerPoints(); len(got) != 3 || got[0].X != 1 || got[1].X != 3 || got[2].X != 5 {
		t.Fatalf("start/step markers = %v", got)
	}

	line.SetMarkEvery(IndexedMarkers(0, -1, 99))
	if got := line.markerPoints(); len(got) != 2 || got[0].X != 0 || got[1].X != 5 {
		t.Fatalf("indexed markers = %v", got)
	}

	line.SetMarkEvery(SliceMarkers(2, 5, 2))
	if got := line.markerPoints(); len(got) != 2 || got[0].X != 2 || got[1].X != 4 {
		t.Fatalf("slice markers = %v", got)
	}
}

func TestAxesPlotTypedLineStyle(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})

	// "--" resolves to matplotlib's dashed pattern (3.7, 1.6) scaled by the
	// line width in pixels (1.5 pt at 100 DPI).
	line := ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{LineStyle: LineStyleDashed})
	if line == nil {
		t.Fatal("Plot returned nil")
	}
	widthPx := 1.5 * 100.0 / 72.0
	want := []float64{3.7 * widthPx, 1.6 * widthPx}
	if len(line.Dashes) != 2 || math.Abs(line.Dashes[0]-want[0]) > 1e-12 || math.Abs(line.Dashes[1]-want[1]) > 1e-12 {
		t.Fatalf("LineStyleDashed dashes = %v, want %v", line.Dashes, want)
	}

	// An explicit dash pattern overrides the typed style.
	line = ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{LineStyle: LineStyleDotted, Dashes: []float64{9, 9}})
	if len(line.Dashes) != 2 || line.Dashes[0] != 9 || line.Dashes[1] != 9 {
		t.Fatalf("explicit Dashes = %v, want [9 9]", line.Dashes)
	}

	// "none" suppresses the line stroke entirely (markers-only plot).
	line = ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{LineStyle: LineStyleNone})
	if line.W != 0 || line.Dashes != nil {
		t.Fatalf("LineStyleNone line: W = %v, Dashes = %v; want 0 and nil", line.W, line.Dashes)
	}

	// Solid and unset both keep a solid stroke.
	for _, ls := range []LineStyle{"", LineStyleSolid} {
		line = ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{LineStyle: ls})
		if line.W != 1.5 || line.Dashes != nil {
			t.Fatalf("LineStyle %q: W = %v, Dashes = %v; want 1.5 and nil", ls, line.W, line.Dashes)
		}
	}
}

func TestAxesPlotConfiguresLineMarkers(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	marker := MarkerStar
	size := 9.0
	face := render.Color{R: 1, A: 0.8}
	edge := render.Color{B: 1, A: 0.6}
	edgeWidth := 2.0

	line := ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{
		Marker:          &marker,
		MarkerSize:      &size,
		MarkerFaceColor: &face,
		MarkerEdgeColor: &edge,
		MarkerEdgeWidth: &edgeWidth,
		MarkEvery:       3,
	})

	if line == nil {
		t.Fatal("Plot returned nil")
	}
	if !line.MarkerSet || line.Marker != marker || line.MarkerSize != size || line.MarkerFaceColor != face || line.MarkerEdgeColor != edge || line.MarkerEdgeWidth != edgeWidth || line.MarkEvery != 3 {
		t.Fatalf("line marker config not applied: %+v", line)
	}
}

func TestLine2D_UsesRCPathOptimizationSettings(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
		},
		W:   1.0,
		Col: render.Color{A: 1},
	}

	r := &recordingRenderer{}
	rc := style.Default
	rc.PathSimplify = true
	rc.PathSimplifyThreshold = 0.25
	rc.AggPathChunkSize = 4096
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   rc,
		Clip: geom.Rect{},
	}

	line.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one Path call, got %d", len(r.pathCalls))
	}
	paint := r.pathCalls[0].paint
	if !paint.Simplify {
		t.Fatalf("line paint should enable simplification from RC: %+v", paint)
	}
	if paint.SimplifyThreshold != 0.25 {
		t.Fatalf("line simplify threshold = %v, want 0.25", paint.SimplifyThreshold)
	}
	if paint.MaxChunkVertices != 4096 {
		t.Fatalf("line max chunk vertices = %d, want 4096", paint.MaxChunkVertices)
	}
}

func TestLine2D_SetDashesUsesMatplotlibUnits(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
		},
		W:   3.0,
		Col: render.Color{A: 1},
	}
	line.SetDashes(10, 4)

	r := &recordingRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	line.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one Path call, got %d", len(r.pathCalls))
	}
	got := r.pathCalls[0].paint.Dashes
	want := []float64{30 * 100.0 / 72.0, 12 * 100.0 / 72.0}
	if len(got) != len(want) || !floatApprox(got[0], want[0], 1e-9) || !floatApprox(got[1], want[1], 1e-9) {
		t.Fatalf("paint dashes = %v, want %v", got, want)
	}
}

func TestLine2D_SingletonData(t *testing.T) {
	line := &Line2D{
		XY:  []geom.Pt{{X: 5, Y: 0.5}}, // single point
		W:   2.0,
		Col: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	// Should not panic with singleton data
	var r render.NullRenderer
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	// This should not panic
	line.Draw(&r, ctx)
}

func TestLine2D_BasicFunctionality(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 0.2},
			{X: 3, Y: 0.9},
			{X: 6, Y: 0.4},
			{X: 10, Y: 0.8},
		},
		W:   2.0,
		Col: render.Color{R: 0, G: 0, B: 0, A: 1},
		z:   1.0,
	}

	// Test Z() method
	if line.Z() != 1.0 {
		t.Errorf("Expected Z() = 1.0, got %f", line.Z())
	}

	// Test Bounds() returns the data bounding box of the line
	bounds := line.Bounds(nil)
	if bounds.Min.X != 0 || bounds.Min.Y != 0 || bounds.Max.X != 10 || bounds.Max.Y != 0.9 {
		t.Errorf("Unexpected bounds, got %+v", bounds)
	}

	// Test Draw() method doesn't panic
	var r render.NullRenderer
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC:   style.Default,
		Clip: geom.Rect{},
	}

	// This should not panic
	line.Draw(&r, ctx)
}

func TestLine2D_AsArtist(t *testing.T) {
	line := &Line2D{
		XY:  []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}},
		W:   1.0,
		Col: render.Color{R: 1, G: 1, B: 1, A: 1},
	}

	// Test that Line2D implements Artist interface
	var _ Artist = line

	// Test integration with Axes
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.15}, Max: geom.Pt{X: 0.95, Y: 0.9}})
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 1)
	ax.Add(line)

	// Test that the figure can be drawn without panic
	var r render.NullRenderer
	DrawFigure(fig, &r)
}

func TestLine2D_Bounds(t *testing.T) {
	line := &Line2D{
		XY: []geom.Pt{{X: 2, Y: -1}, {X: 5, Y: 3}, {X: 8, Y: 0}},
	}
	b := line.Bounds(nil)
	if b.Min.X != 2 || b.Min.Y != -1 || b.Max.X != 8 || b.Max.Y != 3 {
		t.Errorf("unexpected bounds: %v", b)
	}
}

func TestLine2D_BoundsEmpty(t *testing.T) {
	line := &Line2D{}
	b := line.Bounds(nil)
	if b.W() != 0 || b.H() != 0 {
		t.Errorf("empty line should have zero bounds: %v", b)
	}
}

func TestAutoScale(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.Add(&Line2D{XY: []geom.Pt{{X: 2, Y: -1}, {X: 8, Y: 5}}})
	ax.Add(&Line2D{XY: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 3}}})

	ax.AutoScale(0)
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()

	if xMin != 0 || xMax != 10 {
		t.Errorf("x limits: got [%v, %v], want [0, 10]", xMin, xMax)
	}
	if yMin != -1 || yMax != 5 {
		t.Errorf("y limits: got [%v, %v], want [-1, 5]", yMin, yMax)
	}
}

func TestAutoScaleWithMargin(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.Add(&Line2D{XY: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 10}}})

	ax.AutoScale(0.1) // 10% margin
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()

	// 10% of span=10 is 1, so limits should be [-1, 11]
	if xMin != -1 || xMax != 11 {
		t.Errorf("x limits: got [%v, %v], want [-1, 11]", xMin, xMax)
	}
	if yMin != -1 || yMax != 11 {
		t.Errorf("y limits: got [%v, %v], want [-1, 11]", yMin, yMax)
	}
}

func TestAutoScaleRespectsManualLimitsLikeMatplotlibMargins(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.Plot([]float64{0, 10}, []float64{0.08, 0.48})
	ax.SetYLim(0, 1)
	ax.AutoScale(0.04)

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if !floatApprox(xMin, -0.4, 1e-12) || !floatApprox(xMax, 10.4, 1e-12) {
		t.Fatalf("x limits = [%v, %v], want [-0.4, 10.4]", xMin, xMax)
	}
	if yMin != 0 || yMax != 1 {
		t.Fatalf("manual y limits = [%v, %v], want [0, 1]", yMin, yMax)
	}
}

func TestAutoScaleNoData(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.XScale = transform.NewLinear(0, 1)
	ax.YScale = transform.NewLinear(0, 1)

	// AutoScale with no artists should not change limits
	ax.AutoScale(0.05)
	xMin, xMax := ax.XScale.Domain()
	if xMin != 0 || xMax != 1 {
		t.Errorf("limits should be unchanged with no data: got [%v, %v]", xMin, xMax)
	}
}

func TestPlotAutoScalesWithDefaultMargin(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.Plot([]float64{0, 10}, []float64{-1, 3})

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if !floatApprox(xMin, -0.5, 1e-12) || !floatApprox(xMax, 10.5, 1e-12) {
		t.Fatalf("x limits = [%v, %v], want [-0.5, 10.5]", xMin, xMax)
	}
	if !floatApprox(yMin, -1.2, 1e-12) || !floatApprox(yMax, 3.2, 1e-12) {
		t.Fatalf("y limits = [%v, %v], want [-1.2, 3.2]", yMin, yMax)
	}
}

func TestPlotAutoScaleExpandsAcrossLines(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.Plot([]float64{0, 1}, []float64{0, 1})
	ax.Plot([]float64{-2, 4}, []float64{-3, 5})

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if !floatApprox(xMin, -2.3, 1e-12) || !floatApprox(xMax, 4.3, 1e-12) {
		t.Fatalf("x limits = [%v, %v], want [-2.3, 4.3]", xMin, xMax)
	}
	if !floatApprox(yMin, -3.4, 1e-12) || !floatApprox(yMax, 5.4, 1e-12) {
		t.Fatalf("y limits = [%v, %v], want [-3.4, 5.4]", yMin, yMax)
	}
}

func TestPlotAutoScalePreservesExplicitLimits(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.SetXLim(0, 10)
	ax.SetYLim(-2, 2)
	ax.Plot([]float64{-100, 100}, []float64{-50, 50})

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if xMin != 0 || xMax != 10 {
		t.Fatalf("x limits = [%v, %v], want [0, 10]", xMin, xMax)
	}
	if yMin != -2 || yMax != 2 {
		t.Fatalf("y limits = [%v, %v], want [-2, 2]", yMin, yMax)
	}
}

func TestPlotAutoScalePreservesOnlyExplicitAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.SetXLim(0, 10)
	ax.Plot([]float64{-100, 100}, []float64{-50, 50})

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if xMin != 0 || xMax != 10 {
		t.Fatalf("x limits = [%v, %v], want [0, 10]", xMin, xMax)
	}
	if yMin != -55 || yMax != 55 {
		t.Fatalf("y limits = [%v, %v], want [-55, 55]", yMin, yMax)
	}
}

func TestPerAxesMargins(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.Plot([]float64{0, 10}, []float64{0, 10})

	ax.SetXMargin(0.1)
	ax.SetYMargin(0)

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if !floatApprox(xMin, -1, 1e-12) || !floatApprox(xMax, 11, 1e-12) {
		t.Fatalf("x limits with 0.1 margin = [%v, %v], want [-1, 11]", xMin, xMax)
	}
	if !floatApprox(yMin, 0, 1e-12) || !floatApprox(yMax, 10, 1e-12) {
		t.Fatalf("y limits with 0 margin = [%v, %v], want [0, 10]", yMin, yMax)
	}
}

func TestRoundNumbersAutolimit(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.Plot([]float64{0.3, 9.7}, []float64{0.3, 9.7})
	ax.SetAutolimitMode("round_numbers")

	xMin, xMax := ax.XScale.Domain()
	t.Logf("round_numbers x limits = [%v, %v]", xMin, xMax)
	if xMin > 0.3 || xMax < 9.7 {
		t.Fatalf("round_numbers limits [%v, %v] must contain data [0.3, 9.7]", xMin, xMax)
	}
	if xMin != math.Trunc(xMin) || xMax != math.Trunc(xMax) {
		t.Fatalf("round_numbers limits [%v, %v] should be whole numbers", xMin, xMax)
	}
}

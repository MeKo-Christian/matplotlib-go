package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestErrorBar_Draw_Basic(t *testing.T) {
	errBar := &ErrorBar{
		XY: []geom.Pt{
			{X: 1, Y: 2},
			{X: 2, Y: 3},
			{X: 3, Y: 2.5},
		},
		XErr:      []float64{0.2, 0.3, 0.25},
		YErr:      []float64{0.4, 0.2, 0.3},
		LineWidth: 1.2,
		CapSize:   6,
		Color:     render.Color{R: 0, G: 0, B: 0, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	errBar.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}
}

func TestErrorBar_Draw_Empty(t *testing.T) {
	errBar := &ErrorBar{}
	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	errBar.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}
}

func TestErrorBar_Draw_BroadcastError(t *testing.T) {
	errBar := &ErrorBar{
		XY: []geom.Pt{
			{X: 1, Y: 2},
			{X: 2, Y: 3},
			{X: 3, Y: 4},
		},
		XErr:      []float64{0.3},
		YErr:      []float64{0.1},
		LineWidth: 1,
		CapSize:   4,
		Color:     render.Color{R: 0, G: 0, B: 1, A: 1},
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	if err := renderer.Begin(geom.Rect{}); err != nil {
		t.Fatal(err)
	}
	errBar.Draw(renderer, ctx)
	if err := renderer.End(); err != nil {
		t.Fatal(err)
	}
}

func TestErrorBarDrawsMatplotlibDefaultDataLine(t *testing.T) {
	errBar := &ErrorBar{
		XY: []geom.Pt{
			{X: 1, Y: 2},
			{X: 2, Y: 3},
			{X: 3, Y: 2.5},
		},
		YErr:      []float64{0.4, 0.2, 0.3},
		LineWidth: 1.2,
		CapSize:   6,
		Color:     render.Color{R: 0, G: 0.5, B: 0, A: 1},
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()

	errBar.Draw(r, ctx)

	if !hasErrorBarDataLine(r.pathCalls, ctx, errBar.XY) {
		t.Fatalf("errorbar should draw Matplotlib's default data line through %v, got paths %+v", errBar.XY, r.pathCalls)
	}
}

func TestErrorBarCanSuppressDataLineLikeFmtNone(t *testing.T) {
	errBar := &ErrorBar{
		XY: []geom.Pt{
			{X: 1, Y: 2},
			{X: 2, Y: 3},
			{X: 3, Y: 2.5},
		},
		YErr:       []float64{0.4, 0.2, 0.3},
		LineWidth:  1.2,
		CapSize:    6,
		Color:      render.Color{A: 1},
		NoDataLine: true,
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()

	errBar.Draw(r, ctx)

	if hasErrorBarDataLine(r.pathCalls, ctx, errBar.XY) {
		t.Fatalf("errorbar with NoDataLine should match Matplotlib fmt='none'; got data line in paths %+v", r.pathCalls)
	}
}

func TestErrorBarDefaultLineWidthMatchesMatplotlib(t *testing.T) {
	errBar := &ErrorBar{
		XY: []geom.Pt{
			{X: 1, Y: 2},
			{X: 2, Y: 3},
		},
		YErr: []float64{0.4, 0.2},
		Color: render.Color{
			A: 1,
		},
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()

	errBar.Draw(r, ctx)

	if len(r.pathCalls) == 0 {
		t.Fatal("errorbar drew no paths")
	}
	want := pointsToPixels(ctx.RC, 1.5)
	if got := r.pathCalls[0].paint.LineWidth; !floatApprox(got, want, 1e-9) {
		t.Fatalf("default errorbar line width = %v, want Matplotlib 1.5 pt = %v px", got, want)
	}
}

func TestErrorBarMarkerEdgeWidthDefaultsToMatplotlibMarkerEdgeWidth(t *testing.T) {
	errBar := &ErrorBar{
		XY:         []geom.Pt{{X: 1, Y: 2}},
		Color:      render.Color{A: 1},
		Marker:     MarkerCircle,
		MarkerSet:  true,
		MarkerSize: 4.5,
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()

	errBar.Draw(r, ctx)

	if len(r.pathCalls) == 0 {
		t.Fatal("errorbar marker drew no paths")
	}
	want := pointsToPixels(ctx.RC, 1)
	got := r.pathCalls[len(r.pathCalls)-1].paint.LineWidth
	if !floatApprox(got, want, 1e-9) {
		t.Fatalf("default errorbar marker edge width = %v, want Matplotlib 1 pt = %v px", got, want)
	}
}

func TestErrorBarCapSizeIsTotalMarkerLength(t *testing.T) {
	errBar := &ErrorBar{
		XY:        []geom.Pt{{X: 1, Y: 2}},
		XErr:      []float64{0.2},
		YErr:      []float64{0.4},
		LineWidth: 1.2,
		CapSize:   6,
		Color:     render.Color{A: 1},
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()

	errBar.Draw(r, ctx)

	if len(r.pathCalls) != 6 {
		t.Fatalf("path calls = %d, want x/y stems plus four cap paths", len(r.pathCalls))
	}
	for _, idx := range []int{1, 2, 4, 5} {
		path := r.pathCalls[idx].path
		if len(path.V) != 2 {
			t.Fatalf("cap path %d vertices = %d, want 2", idx, len(path.V))
		}
		got := math.Hypot(path.V[1].X-path.V[0].X, path.V[1].Y-path.V[0].Y)
		if math.Abs(got-6) > 1e-9 {
			t.Fatalf("cap path %d length = %.6f px, want 6 px total cap marker length", idx, got)
		}
	}
}

func TestErrorBarLimitCaretDrawsWhenCapsDisabled(t *testing.T) {
	errBar := &ErrorBar{
		XY:        []geom.Pt{{X: 1, Y: 1}},
		YErrUpper: []float64{1},
		LoLimits:  []bool{true},
		CapSize:   0,
		LineWidth: 1,
		Color:     render.Color{A: 1},
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()

	errBar.Draw(r, ctx)

	if len(r.pathCalls) != 2 {
		t.Fatalf("path calls = %d, want stem and Matplotlib limit caret even with capsize=0", len(r.pathCalls))
	}
	caret := r.pathCalls[1].path.V
	if len(caret) != 3 {
		t.Fatalf("caret vertices = %d, want 3", len(caret))
	}
	endpoint := ctx.DataToPixel.Apply(geom.Pt{X: 1, Y: 2})
	if caret[0].Y != endpoint.Y || caret[2].Y != endpoint.Y {
		t.Fatalf("caret base y = %.3f, %.3f; want endpoint y %.3f", caret[0].Y, caret[2].Y, endpoint.Y)
	}
	if got, want := math.Abs(caret[2].X-caret[0].X), pointsToPixels(ctx.RC, 6); !floatApprox(got, want, 1e-9) {
		t.Fatalf("default limit caret base width = %.3f px, want Matplotlib lines.markersize %.3f px", got, want)
	}
}

func TestErrorBarCapWidthUsesMatplotlibMarkerEdgeWidth(t *testing.T) {
	errBar := &ErrorBar{
		XY:        []geom.Pt{{X: 1, Y: 2}},
		XErr:      []float64{0.2},
		YErr:      []float64{0.4},
		LineWidth: 2.0,
		CapSize:   6,
		Color:     render.Color{A: 1},
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()

	errBar.Draw(r, ctx)

	want := pointsToPixels(ctx.RC, 1)
	for _, idx := range []int{1, 2, 4, 5} {
		if got := r.pathCalls[idx].paint.LineWidth; !floatApprox(got, want, 1e-9) {
			t.Fatalf("cap path %d line width = %v, want Matplotlib markeredgewidth 1 pt = %v px", idx, got, want)
		}
	}
}

func TestErrorBarSegmentsUseMatplotlibSnapAuto(t *testing.T) {
	errBar := &ErrorBar{
		XY:        []geom.Pt{{X: 1, Y: 2}},
		XErr:      []float64{0.2},
		YErr:      []float64{0.4},
		LineWidth: 1.2,
		CapSize:   6,
		Color:     render.Color{A: 1},
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()

	errBar.Draw(r, ctx)

	// The bar segments rely on the backend's SnapAuto pixel snapper, exactly
	// like Matplotlib's vlines/hlines. The caps are drawn as pre-snapped marker
	// geometry (see faithfulCapPath) and must therefore opt out of the generic
	// per-vertex snapper with SnapOff, mirroring Matplotlib's draw_markers which
	// floors the marker centre and snaps the marker half-extent separately.
	var bars, caps int
	for i, call := range r.pathCalls {
		switch call.paint.Snap {
		case render.SnapAuto:
			bars++
		case render.SnapOff:
			caps++
		default:
			t.Fatalf("errorbar path %d snap mode = %v, want SnapAuto (bar) or SnapOff (cap)", i, call.paint.Snap)
		}
	}
	if bars != 2 {
		t.Fatalf("errorbar SnapAuto bar segments = %d, want 2 (x and y bars)", bars)
	}
	if caps != 4 {
		t.Fatalf("errorbar SnapOff cap segments = %d, want 4 (two x caps, two y caps)", caps)
	}
}

func TestErrorBarLimitCaretUsesEndpointAsBase(t *testing.T) {
	errBar := &ErrorBar{
		XY:        []geom.Pt{{X: 1, Y: 1}},
		YErrUpper: []float64{1},
		LoLimits:  []bool{true},
		CapSize:   8,
		LineWidth: 1,
		Color:     render.Color{R: 0, G: 0, B: 0, A: 1},
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()

	errBar.Draw(r, ctx)

	if len(r.pathCalls) != 3 {
		t.Fatalf("path calls = %d, want stem, caret, and cap marker", len(r.pathCalls))
	}
	caret := r.pathCalls[1].path.V
	if len(caret) != 3 {
		t.Fatalf("caret vertices = %d, want 3", len(caret))
	}
	if got, want := r.pathCalls[1].paint.Fill, errBar.Color; got != want {
		t.Fatalf("limit caret fill = %+v, want Matplotlib Agg marker color %+v", got, want)
	}
	if got := r.pathCalls[1].paint.LineJoin; got != render.JoinMiter {
		t.Fatalf("limit caret join = %v, want Matplotlib marker miter join", got)
	}
	if cmds := r.pathCalls[1].path.C; len(cmds) == 0 || cmds[len(cmds)-1] == geom.ClosePath {
		t.Fatalf("limit caret commands = %v, want open Matplotlib marker path", cmds)
	}
	endpoint := ctx.DataToPixel.Apply(geom.Pt{X: 1, Y: 2})
	if caret[0].Y != endpoint.Y || caret[2].Y != endpoint.Y {
		t.Fatalf("caret base y = %.3f, %.3f; want endpoint y %.3f", caret[0].Y, caret[2].Y, endpoint.Y)
	}
	if got, want := math.Abs(caret[2].X-caret[0].X), 8.0; math.Abs(got-want) > 1e-9 {
		t.Fatalf("caret base width = %.3f px, want %.3f px from cap marker size", got, want)
	}
	if caret[1].Y <= endpoint.Y {
		t.Fatalf("lower-limit caret tip y = %.3f, want below endpoint %.3f before backend flip", caret[1].Y, endpoint.Y)
	}
}

func TestErrorBarUpperLimitCaretPointsDownFromEndpoint(t *testing.T) {
	errBar := &ErrorBar{
		XY:        []geom.Pt{{X: 1, Y: 2}},
		YErrLower: []float64{1},
		UpLimits:  []bool{true},
		CapSize:   8,
		LineWidth: 1,
		Color:     render.Color{A: 1},
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()

	errBar.Draw(r, ctx)

	if len(r.pathCalls) != 3 {
		t.Fatalf("path calls = %d, want stem, caret, and cap marker", len(r.pathCalls))
	}
	caret := r.pathCalls[1].path.V
	endpoint := ctx.DataToPixel.Apply(geom.Pt{X: 1, Y: 1})
	if caret[0].Y != endpoint.Y || caret[2].Y != endpoint.Y {
		t.Fatalf("caret base y = %.3f, %.3f; want endpoint y %.3f", caret[0].Y, caret[2].Y, endpoint.Y)
	}
	if caret[1].Y >= endpoint.Y {
		t.Fatalf("upper-limit caret tip y = %.3f, want above endpoint %.3f before backend flip", caret[1].Y, endpoint.Y)
	}
}

func hasErrorBarDataLine(calls []recordedPathCall, ctx *DrawContext, points []geom.Pt) bool {
	if len(points) == 0 {
		return false
	}
	for _, call := range calls {
		if len(call.path.V) != len(points) || len(call.path.C) != len(points) {
			continue
		}
		if call.path.C[0] != geom.MoveTo {
			continue
		}
		matches := true
		for i, point := range points {
			if i > 0 && call.path.C[i] != geom.LineTo {
				matches = false
				break
			}
			if call.path.V[i] != ctx.DataToPixel.Apply(point) {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func TestErrorBar_ZOrder(t *testing.T) {
	errBar := &ErrorBar{z: 1.25}
	if got := errBar.Z(); got != 1.25 {
		t.Errorf("expected Z() = 1.25, got %v", got)
	}
}

func TestErrorBar_Bounds(t *testing.T) {
	errBar := &ErrorBar{
		XY: []geom.Pt{
			{X: 2, Y: 3},
			{X: 5, Y: 5},
		},
		XErr: []float64{0.5},
		YErr: []float64{0.4, 0.6},
	}
	bounds := errBar.Bounds(nil)
	if bounds.Min.X != 1.5 || bounds.Max.X != 5.5 || bounds.Min.Y != 2.6 || bounds.Max.Y != 5.6 {
		t.Errorf("unexpected bounds: %v", bounds)
	}
}

func TestAxes_ErrorBar(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	errBar := ax.ErrorBar(
		[]float64{1, 2, 3},
		[]float64{1.1, 2.2, 3.3},
		[]float64{0.1},
		nil,
	)
	if errBar == nil {
		t.Fatal("ErrorBar should return non-nil for non-empty data")
	}
}

func TestAxesErrorBarDefaultsMatchMatplotlib(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	errBar := ax.ErrorBar(
		[]float64{1, 2},
		[]float64{3, 4},
		nil,
		[]float64{0.2},
	)
	if errBar == nil {
		t.Fatal("expected non-nil error bar")
	}
	if errBar.LineWidth != 0 {
		t.Fatalf("default Axes.ErrorBar line width = %v, want renderer Matplotlib default", errBar.LineWidth)
	}
	if errBar.CapSize != 0 {
		t.Fatalf("default Axes.ErrorBar cap size = %v, want Matplotlib errorbar.capsize=0", errBar.CapSize)
	}
}

func TestAxes_ErrorBar_Options(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	col := render.Color{R: 1, G: 0, B: 0, A: 1}
	lineWidth := 2.0
	capSize := 6.0
	alpha := 0.8
	errBar := ax.ErrorBar(
		[]float64{1, 2},
		[]float64{3, 4},
		nil,
		[]float64{0.2},
		ErrorBarOptions{
			Color:      &col,
			LineWidth:  &lineWidth,
			CapSize:    &capSize,
			Alpha:      &alpha,
			NoDataLine: true,
			Label:      "test",
		},
	)

	if errBar == nil {
		t.Fatal("expected non-nil error bar")
	}
	if errBar.Label != "test" {
		t.Errorf("expected label 'test', got %q", errBar.Label)
	}
	if errBar.LineWidth != lineWidth {
		t.Errorf("expected line width %v, got %v", lineWidth, errBar.LineWidth)
	}
	if want := pointsToPixels(fig.RC, 2*capSize); errBar.CapSize != want {
		t.Errorf("expected cap size %v px, got %v", want, errBar.CapSize)
	}
	// The explicit alpha is baked into the stroke color; the Alpha field stays
	// at the 0 "unset" sentinel.
	if errBar.Alpha != 0 {
		t.Errorf("expected Alpha sentinel 0, got %v", errBar.Alpha)
	}
	if errBar.Color.A != alpha {
		t.Errorf("expected baked color alpha %v, got %v", alpha, errBar.Color.A)
	}
	if !errBar.NoDataLine {
		t.Error("expected NoDataLine option to be applied")
	}
}

func TestAxesErrorBarCapSizeUsesMatplotlibMarkerLength(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	capSize := 6.0

	errBar := ax.ErrorBar(
		[]float64{1},
		[]float64{2},
		nil,
		[]float64{0.2},
		ErrorBarOptions{CapSize: &capSize},
	)
	if errBar == nil {
		t.Fatal("expected non-nil error bar")
	}
	if want := pointsToPixels(fig.RC, 2*capSize); errBar.CapSize != want {
		t.Fatalf("cap marker length = %v px, want Matplotlib markersize=2*capsize = %v px", errBar.CapSize, want)
	}
}

func TestAxes_ErrorBar_AsymmetricLimitsAndValidation(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	errBar := ax.ErrorBar(
		[]float64{10},
		[]float64{5},
		[]float64{1},
		[]float64{2},
		ErrorBarOptions{
			XErrLower: []float64{2},
			XErrUpper: []float64{3},
			YErrLower: []float64{1},
			YErrUpper: []float64{4},
			XLoLimits: []bool{true},
			UpLimits:  []bool{true},
		},
	)
	if errBar == nil {
		t.Fatal("expected asymmetric errorbar")
	}
	if got, want := errBar.XErrLower[0], 2.0; got != want {
		t.Fatalf("x lower = %v, want %v", got, want)
	}
	bounds := errBar.Bounds(nil)
	if bounds.Min.X != 10 || bounds.Max.X != 13 || bounds.Min.Y != 4 || bounds.Max.Y != 5 {
		t.Fatalf("bounds = %+v, want x[10,13] y[4,5]", bounds)
	}

	if got := ax.ErrorBar([]float64{1, 2}, []float64{1, 2}, []float64{-1}, nil); got != nil {
		t.Fatal("negative symmetric errors should be rejected")
	}
	if got := ax.ErrorBar([]float64{1, 2}, []float64{1, 2}, nil, nil, ErrorBarOptions{YErrUpper: []float64{1, 2, 3}}); got != nil {
		t.Fatal("asymmetric errors with invalid length should be rejected")
	}
}

func TestErrorBarErrorEverySkipsErrorStemsButKeepsDataLine(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	every := 2
	errBar := ax.ErrorBar(
		[]float64{0, 1, 2, 3, 4},
		[]float64{1, 2, 3, 4, 5},
		nil,
		[]float64{0.2},
		ErrorBarOptions{ErrorEvery: every},
	)
	if errBar == nil {
		t.Fatal("expected errorbar")
	}

	r := &recordingRenderer{}
	ctx := createTestDrawContext()
	errBar.Draw(r, ctx)

	if !hasErrorBarDataLine(r.pathCalls, ctx, errBar.XY) {
		t.Fatal("errorevery should not thin the data line")
	}
	if got, want := countNonDataLinePaths(r.pathCalls, ctx, errBar.XY), 3; got != want {
		t.Fatalf("errorbar-only path count = %d, want stems for indices 0,2,4 with default capsize=0 (%d)", got, want)
	}
}

func TestErrorBarErrorEveryStartMatchesTupleForm(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	every := 2
	start := 1
	errBar := ax.ErrorBar(
		[]float64{0, 1, 2, 3, 4},
		[]float64{1, 2, 3, 4, 5},
		nil,
		[]float64{0.2},
		ErrorBarOptions{ErrorEvery: every, ErrorEveryStart: start},
	)
	if errBar == nil {
		t.Fatal("expected errorbar")
	}

	r := &recordingRenderer{}
	ctx := createTestDrawContext()
	errBar.Draw(r, ctx)

	if got, want := countNonDataLinePaths(r.pathCalls, ctx, errBar.XY), 2; got != want {
		t.Fatalf("errorbar-only path count = %d, want stems for indices 1,3 with default capsize=0 (%d)", got, want)
	}
	if got := ax.ErrorBar([]float64{0, 1}, []float64{1, 2}, nil, []float64{0.1}, ErrorBarOptions{ErrorEvery: -1}); got != nil {
		t.Fatal("negative ErrorEvery should be rejected")
	}
}

func countNonDataLinePaths(calls []recordedPathCall, ctx *DrawContext, points []geom.Pt) int {
	count := 0
	for _, call := range calls {
		if hasErrorBarDataLine([]recordedPathCall{call}, ctx, points) {
			continue
		}
		count++
	}
	return count
}

func TestErrorBarCapThickSetsField(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	capThick := 3.0
	capSize := 5.0
	bar := ax.ErrorBar([]float64{1, 2}, []float64{1, 2}, nil, []float64{0.5, 0.5},
		ErrorBarOptions{CapThick: &capThick, CapSize: &capSize})
	if bar == nil {
		t.Fatal("ErrorBar returned nil")
	}
	// CapThick is stored in points now; pixel conversion happens at the sink.
	if bar.CapThick != capThick {
		t.Fatalf("bar.CapThick = %v, want %v", bar.CapThick, capThick)
	}
}

func TestErrorBarDrawUsesCapThick(t *testing.T) {
	bar := &ErrorBar{
		XY:        []geom.Pt{{X: 1, Y: 2}},
		YErr:      []float64{0.5},
		LineWidth: 2,
		CapSize:   10,
		CapThick:  7,
		Color:     render.Color{A: 1},
	}
	r := &recordingRenderer{}
	ctx := createTestDrawContext()
	bar.Draw(r, ctx)
	// CapThick is points; the rendered cap stroke is converted to device pixels.
	wantWidth := pointsToPixels(ctx.RC, 7)
	found := false
	for _, c := range r.pathCalls {
		if c.paint.LineWidth == wantWidth {
			found = true
		}
	}
	if !found {
		t.Fatalf("no cap path drawn with CapThick width %v; got %d paths", wantWidth, len(r.pathCalls))
	}
}

func TestErrorBarDrawCapThickDefaultsToOnePoint(t *testing.T) {
	bar := &ErrorBar{
		XY:        []geom.Pt{{X: 1, Y: 2}},
		YErr:      []float64{0.5},
		LineWidth: 2,
		CapSize:   10,
		Color:     render.Color{A: 1},
	}
	ctx := createTestDrawContext()
	r := &recordingRenderer{}
	bar.Draw(r, ctx)
	want := pointsToPixels(ctx.RC, 1)
	found := false
	for _, c := range r.pathCalls {
		if c.paint.LineWidth == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("no cap path drawn with default width %v; got %d paths", want, len(r.pathCalls))
	}
}

package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
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
	if got := r.pathCalls[1].paint.Fill; got.A <= 0 {
		t.Fatalf("limit caret fill alpha = %v, want filled Matplotlib cap marker", got.A)
	}
	if cmds := r.pathCalls[1].path.C; len(cmds) == 0 || cmds[len(cmds)-1] != geom.ClosePath {
		t.Fatalf("limit caret commands = %v, want closed filled marker path", cmds)
	}
	endpoint := ctx.DataToPixel.Apply(geom.Pt{X: 1, Y: 2})
	if caret[0].Y != endpoint.Y || caret[2].Y != endpoint.Y {
		t.Fatalf("caret base y = %.3f, %.3f; want endpoint y %.3f", caret[0].Y, caret[2].Y, endpoint.Y)
	}
	if caret[1].Y >= endpoint.Y {
		t.Fatalf("lower-limit caret tip y = %.3f, want above endpoint %.3f in display space", caret[1].Y, endpoint.Y)
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
	if errBar.CapSize != capSize {
		t.Errorf("expected cap size %v, got %v", capSize, errBar.CapSize)
	}
	if errBar.Alpha != alpha {
		t.Errorf("expected alpha %v, got %v", alpha, errBar.Alpha)
	}
	if !errBar.NoDataLine {
		t.Error("expected NoDataLine option to be applied")
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

package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

func createFractionalDrawContext() *DrawContext {
	return &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 10),
			AxesToPixel: transform.NewAffine(geom.Affine{A: 95.3, D: -97.7, E: 50.2, F: 449.7}),
		},
		RC:   style.Default,
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1024, Y: 1024}},
	}
}

func TestSnapPixelRect(t *testing.T) {
	rect, ok := snapPixelRect(geom.Rect{
		Min: geom.Pt{X: 107.38, Y: 229.875},
		Max: geom.Pt{X: 183.62, Y: 449.7},
	}, false)
	if !ok {
		t.Fatal("expected snapped fill rect")
	}
	if rect.Min.X != 107 || rect.Max.X != 184 || rect.Min.Y != 230 || rect.Max.Y != 450 {
		t.Fatalf("unexpected snapped fill rect: %+v", rect)
	}

	strokeRect, ok := snapPixelRect(geom.Rect{
		Min: geom.Pt{X: 107.38, Y: 229.875},
		Max: geom.Pt{X: 183.62, Y: 449.7},
	}, true)
	if !ok {
		t.Fatal("expected snapped stroke rect")
	}
	// Display space is y-up: the stroke half-pixel offset points the opposite way
	// in Y (round - 0.5), so the snapped Y edges are 1px lower than the X form.
	if strokeRect.Min.X != 107.5 || strokeRect.Max.X != 184.5 || strokeRect.Min.Y != 229.5 || strokeRect.Max.Y != 449.5 {
		t.Fatalf("unexpected snapped stroke rect: %+v", strokeRect)
	}
}

func TestBar2D_Draw_SnapsFillAndStrokeRects(t *testing.T) {
	bar := &Bar2D{
		X:           []float64{1},
		Heights:     []float64{2.25},
		Width:       0.8,
		Color:       render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1},
		EdgeColor:   render.Color{R: 0, G: 0, B: 0, A: 1},
		EdgeWidth:   1.0,
		Orientation: BarVertical,
	}

	r := &recordingRenderer{}
	bar.Draw(r, createFractionalDrawContext())

	if len(r.pathCalls) != 2 {
		t.Fatalf("expected fill and stroke path calls, got %d", len(r.pathCalls))
	}

	fill := r.pathCalls[0].path.V
	if len(fill) != 4 {
		t.Fatalf("expected 4 fill vertices, got %d", len(fill))
	}
	if fill[0] != (geom.Pt{X: 56, Y: 428}) || fill[2] != (geom.Pt{X: 64, Y: 450}) {
		t.Fatalf("unexpected snapped fill vertices: %+v", fill)
	}

	stroke := r.pathCalls[1].path.V
	if len(stroke) != 4 {
		t.Fatalf("expected 4 stroke vertices, got %d", len(stroke))
	}
	// y-up stroke snap offsets Y by -0.5 (round - 0.5).
	if stroke[0] != (geom.Pt{X: 56.5, Y: 427.5}) || stroke[2] != (geom.Pt{X: 64.5, Y: 449.5}) {
		t.Fatalf("unexpected snapped stroke vertices: %+v", stroke)
	}
}

func TestHist2D_Draw_UsesMatplotlibPatchSnapAuto(t *testing.T) {
	hist := &Hist2D{
		Data:      []float64{1, 1.5},
		BinEdges:  []float64{0, 2},
		Color:     render.Color{R: 0.3, G: 0.6, B: 0.9, A: 1},
		EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
		EdgeWidth: 1.0,
	}

	r := &recordingRenderer{}
	hist.Draw(r, createFractionalDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one patch path call, got %d", len(r.pathCalls))
	}

	path := r.pathCalls[0].path.V
	if len(path) != 4 {
		t.Fatalf("expected 4 patch vertices, got %d", len(path))
	}
	if !pointAlmostEqual(path[0], geom.Pt{X: 50.2, Y: 430.16}) || !pointAlmostEqual(path[2], geom.Pt{X: 69.26, Y: 449.7}) {
		t.Fatalf("unexpected unsnapped histogram patch vertices: %+v", path)
	}
	if got := r.pathCalls[0].paint.Snap; got != render.SnapAuto {
		t.Fatalf("histogram patch snap = %v, want SnapAuto", got)
	}
}

func pointAlmostEqual(got, want geom.Pt) bool {
	return almostEqualFloat(got.X, want.X) && almostEqualFloat(got.Y, want.Y)
}

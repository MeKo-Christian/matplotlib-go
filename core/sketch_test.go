package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

func sketchTestContext() *DrawContext {
	return &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC: style.Default,
	}
}

func TestLine2DSketchOverridePropagatesToPaint(t *testing.T) {
	line := &Line2D{
		XY:     []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}},
		W:      1,
		Col:    render.Color{A: 1},
		Sketch: render.SketchParams{Scale: 2, Length: 50, Randomness: 3},
	}
	r := &recordingRenderer{}
	line.Draw(r, sketchTestContext())

	if len(r.pathCalls) == 0 {
		t.Fatal("expected at least one Path call")
	}
	if got := r.pathCalls[0].paint.Sketch; got != line.Sketch {
		t.Fatalf("paint.Sketch = %+v, want %+v", got, line.Sketch)
	}
}

func TestLine2DNoSketchLeavesPaintZero(t *testing.T) {
	line := &Line2D{
		XY:  []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}},
		W:   1,
		Col: render.Color{A: 1},
	}
	r := &recordingRenderer{}
	line.Draw(r, sketchTestContext())

	if len(r.pathCalls) == 0 {
		t.Fatal("expected at least one Path call")
	}
	if got := r.pathCalls[0].paint.Sketch; got != (render.SketchParams{}) {
		t.Fatalf("paint.Sketch = %+v, want zero (figure default applies at renderer)", got)
	}
}

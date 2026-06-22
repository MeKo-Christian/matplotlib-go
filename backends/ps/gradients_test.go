package ps

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func gradientTestRenderer(t *testing.T) *Renderer {
	t.Helper()
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	return r
}

func unitSquare() geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 90})
	p.LineTo(geom.Pt{X: 10, Y: 90})
	p.Close()
	return p
}

func TestSupportsGradientAndPatternFill(t *testing.T) {
	r := newTestRenderer(t)
	if !r.SupportsGradientFill() {
		t.Fatal("PS backend should report native gradient support")
	}
	if !r.SupportsPatternFill() {
		t.Fatal("PS backend should report native pattern support")
	}
	if _, ok := any(r).(render.GradientFiller); !ok {
		t.Fatal("PS backend should implement render.GradientFiller")
	}
	if _, ok := any(r).(render.PatternFiller); !ok {
		t.Fatal("PS backend should implement render.PatternFiller")
	}
}

func TestLinearGradientEmitsAxialShading(t *testing.T) {
	r := gradientTestRenderer(t)
	r.Path(unitSquare(), &render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 10, Y: 10},
			End:   geom.Pt{X: 90, Y: 10},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{B: 1, A: 1}},
			},
		},
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := string(r.document)
	for _, want := range []string{
		"/ShadingType 2",
		"/Coords [10 10 90 10]",
		"/FunctionType 2",
		"shfill",
		"clip",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("linear gradient output missing %q in\n%s", want, doc)
		}
	}
}

func TestRadialGradientEmitsRadialShadingWithStitching(t *testing.T) {
	r := gradientTestRenderer(t)
	r.Path(unitSquare(), &render.Paint{
		FillGradient: render.GradientFill{
			Kind:   render.RadialGradient,
			Center: geom.Pt{X: 50, Y: 50},
			Radius: 40,
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 0.5, Color: render.Color{G: 1, A: 1}},
				{Offset: 1, Color: render.Color{B: 1, A: 1}},
			},
		},
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := string(r.document)
	for _, want := range []string{
		"/ShadingType 3",
		"/Coords [50 50 0 50 50 40]",
		"/FunctionType 3",
		"/Bounds [0.5]",
		"shfill",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("radial gradient output missing %q in\n%s", want, doc)
		}
	}
}

func TestGradientFillWithStrokeEmitsOutline(t *testing.T) {
	r := gradientTestRenderer(t)
	r.Path(unitSquare(), &render.Paint{
		Stroke:    render.Color{A: 1},
		LineWidth: 2,
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 10, Y: 10},
			End:   geom.Pt{X: 10, Y: 90},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{B: 1, A: 1}},
			},
		},
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := string(r.document)
	if !strings.Contains(doc, "shfill") {
		t.Fatalf("expected gradient shfill in\n%s", doc)
	}
	if !strings.Contains(doc, "stroke") {
		t.Fatalf("expected stroke outline after gradient in\n%s", doc)
	}
	if strings.Index(doc, "shfill") > strings.Index(doc, "stroke") {
		t.Fatalf("stroke must follow the gradient fill in\n%s", doc)
	}
}

func TestPatternFillTilesCellInClip(t *testing.T) {
	r := gradientTestRenderer(t)
	var dot geom.Path
	dot.MoveTo(geom.Pt{X: 0, Y: 0})
	dot.LineTo(geom.Pt{X: 4, Y: 0})
	dot.LineTo(geom.Pt{X: 4, Y: 4})
	dot.LineTo(geom.Pt{X: 0, Y: 4})
	dot.Close()

	r.Path(unitSquare(), &render.Paint{
		FillPattern: render.PatternFill{
			ID:         "dots",
			Cell:       geom.Rect{Max: geom.Pt{X: 20, Y: 20}},
			Path:       dot,
			Foreground: render.Color{R: 1, A: 1},
			Background: render.Color{R: 0.9, G: 0.9, B: 0.9, A: 1},
		},
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := r.document
	if !bytes.Contains(doc, []byte("clip")) {
		t.Fatalf("pattern fill must clip to the path in\n%s", doc)
	}
	// Background + foreground both painted: two distinct setrgbcolor sources and
	// multiple fill ops from the tiling loop.
	if n := bytes.Count(doc, []byte("fill\n")); n < 4 {
		t.Fatalf("expected several tiled fills, got %d in\n%s", n, doc)
	}
}

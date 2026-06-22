package pgf

import (
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestPGFSupportsGradientAndPatternFill(t *testing.T) {
	r, err := New(120, 80, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if !r.SupportsGradientFill() {
		t.Fatal("PGF backend should report native gradient support")
	}
	if !r.SupportsPatternFill() {
		t.Fatal("PGF backend should report native pattern support")
	}
	for name, ok := range map[string]bool{
		"GradientFiller":         implements[render.GradientFiller](r),
		"PatternFiller":          implements[render.PatternFiller](r),
		"ClipPathTransformer":    implements[render.ClipPathTransformer](r),
		"FontVerticalTextDrawer": implements[render.FontVerticalTextDrawer](r),
	} {
		if !ok {
			t.Fatalf("PGF backend should implement render.%s", name)
		}
	}
}

func implements[T any](v any) bool {
	_, ok := v.(T)
	return ok
}

func TestPGFLinearGradientEmitsHorizontalShading(t *testing.T) {
	doc := renderPGFDocument(t, func(r *Renderer) {
		r.Path(testRectPath(), &render.Paint{
			FillGradient: render.GradientFill{
				Kind:  render.LinearGradient,
				Start: geom.Pt{X: 10, Y: 10},
				End:   geom.Pt{X: 50, Y: 10},
				Stops: []render.GradientStop{
					{Offset: 0, Color: render.Color{R: 1, A: 1}},
					{Offset: 1, Color: render.Color{B: 1, A: 1}},
				},
			},
		})
	})
	s := string(doc)
	for _, want := range []string{
		"\\pgfdeclarehorizontalshading{mplgpgfshading1}{100bp}",
		"\\pgfshadepath{mplgpgfshading1}{0}",
		"color(0bp)=(",
		"color(100bp)=(",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("linear gradient output missing %q in\n%s", want, s)
		}
	}
}

func TestPGFLinearGradientRotatesForDiagonalAxis(t *testing.T) {
	doc := renderPGFDocument(t, func(r *Renderer) {
		r.Path(testRectPath(), &render.Paint{
			FillGradient: render.GradientFill{
				Kind:  render.LinearGradient,
				Start: geom.Pt{X: 10, Y: 10},
				End:   geom.Pt{X: 50, Y: 40},
				Stops: []render.GradientStop{
					{Offset: 0, Color: render.Color{R: 1, A: 1}},
					{Offset: 1, Color: render.Color{B: 1, A: 1}},
				},
			},
		})
	})
	// Start (10,10) -> End (50,40): atan2(30,40) = 36.87°, so the shading is
	// drawn at that non-zero angle via \pgfshadepath's rotation argument.
	if !strings.Contains(string(doc), "\\pgfshadepath{mplgpgfshading1}{36") {
		t.Fatalf("diagonal linear gradient should rotate the shading by ~36.87° in\n%s", doc)
	}
}

func TestPGFRadialGradientEmitsRadialShading(t *testing.T) {
	doc := renderPGFDocument(t, func(r *Renderer) {
		r.Path(testRectPath(), &render.Paint{
			FillGradient: render.GradientFill{
				Kind:   render.RadialGradient,
				Center: geom.Pt{X: 30, Y: 25},
				Radius: 20,
				Stops: []render.GradientStop{
					{Offset: 0, Color: render.Color{R: 1, A: 1}},
					{Offset: 1, Color: render.Color{B: 1, A: 1}},
				},
			},
		})
	})
	s := string(doc)
	for _, want := range []string{
		"\\pgfdeclareradialshading{mplgpgfshading1}",
		"\\pgfuseshading{mplgpgfshading1}",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("radial gradient output missing %q in\n%s", want, s)
		}
	}
}

func TestPGFPatternFillTilesInClip(t *testing.T) {
	var dot geom.Path
	dot.MoveTo(geom.Pt{X: 0, Y: 0})
	dot.LineTo(geom.Pt{X: 3, Y: 0})
	dot.LineTo(geom.Pt{X: 3, Y: 3})
	dot.LineTo(geom.Pt{X: 0, Y: 3})
	dot.Close()

	doc := renderPGFDocument(t, func(r *Renderer) {
		r.Path(testRectPath(), &render.Paint{
			FillPattern: render.PatternFill{
				ID:         "dots",
				Cell:       geom.Rect{Max: geom.Pt{X: 10, Y: 10}},
				Path:       dot,
				Foreground: render.Color{R: 1, A: 1},
				Background: render.Color{R: 0.9, G: 0.9, B: 0.9, A: 1},
			},
		})
	})
	s := string(doc)
	if !strings.Contains(s, "\\pgfscope") || !strings.Contains(s, "\\pgfusepath{clip}") {
		t.Fatalf("pattern fill must clip within a scope in\n%s", s)
	}
	if n := strings.Count(s, "\\pgfusepath{fill}"); n < 4 {
		t.Fatalf("expected several tiled fills, got %d in\n%s", n, s)
	}
}

func TestPGFVerticalTextStacksGlyphs(t *testing.T) {
	doc := renderPGFDocument(t, func(r *Renderer) {
		r.DrawTextVertical("Ab", geom.Pt{X: 30, Y: 40}, 12, render.Color{A: 1})
	})
	s := string(doc)
	// Two glyphs => two \pgftext invocations, stacked (no rotation).
	if n := strings.Count(s, "\\pgftext["); n != 2 {
		t.Fatalf("expected one \\pgftext per stacked glyph, got %d in\n%s", n, s)
	}
	if strings.Contains(s, "rotate=") {
		t.Fatalf("vertical (stacked) text should not be rotated in\n%s", s)
	}
}

func TestPGFClipPathTransformedAppliesAffine(t *testing.T) {
	doc := renderPGFDocument(t, func(r *Renderer) {
		r.ClipPathTransformed(testRectPath(), geom.Affine{A: 1, D: 1, E: 100, F: 5})
	})
	s := string(doc)
	if !strings.Contains(s, "\\pgfusepath{clip}") {
		t.Fatalf("ClipPathTransformed must emit a clip in\n%s", s)
	}
	// The x-translation of 100 should shift the first moveto from 10 to 110.
	if !strings.Contains(s, "\\pgfpathmoveto{\\pgfpoint{110pt}{15pt}}") {
		t.Fatalf("ClipPathTransformed did not apply the affine translation in\n%s", s)
	}
}

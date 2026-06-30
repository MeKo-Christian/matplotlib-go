package gobasic

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/render"
)

// TestPathDeviceWarnsOnUnsupportedFill verifies that a paint carrying a gradient
// fill (which gobasic cannot render) produces a one-shot diagnostic rather than
// a silent blank fill.
func TestPathDeviceWarnsOnUnsupportedFill(t *testing.T) {
	var warnings []string
	restore := diag.SetHandler(func(m string) { warnings = append(warnings, m) })
	defer restore()

	r := New(8, 8, render.Color{R: 1, G: 1, B: 1, A: 1})
	_ = r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 8, Y: 8}})
	p := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{{X: 1, Y: 1}, {X: 7, Y: 1}, {X: 7, Y: 7}},
	}
	paint := render.Paint{
		Fill: render.Color{R: 1, A: 1},
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Stops: []render.GradientStop{{Offset: 0, Color: render.Color{A: 1}}},
		},
	}
	r.Path(p, &paint)
	_ = r.End()

	if len(warnings) == 0 {
		t.Fatal("expected a diagnostic for an unsupported gradient fill, got none")
	}
}

// TestPathDeviceCarriesFullPaintState confirms the quantized paint preserves
// fields the old reconstruction dropped (Alpha, CompositeMode) and does not
// mutate the caller's dash slice.
func TestPathDeviceCarriesFullPaintState(t *testing.T) {
	dashes := []float64{3.3, 1.1}
	orig := append([]float64(nil), dashes...)
	paint := render.Paint{
		Stroke:    render.Color{R: 1, A: 1},
		LineWidth: 1.0,
		Dashes:    dashes,
	}
	r := New(8, 8, render.Color{R: 1, G: 1, B: 1, A: 1})
	_ = r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 8, Y: 8}})
	p := geom.Path{C: []geom.Cmd{geom.MoveTo, geom.LineTo}, V: []geom.Pt{{X: 1, Y: 1}, {X: 7, Y: 7}}}
	r.Path(p, &paint)
	_ = r.End()

	for i := range dashes {
		if dashes[i] != orig[i] {
			t.Fatalf("caller dash slice was mutated: %v, want %v", dashes, orig)
		}
	}
}

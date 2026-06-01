package core_test

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestLayoutMathTextSuppressesOperatorSpacingInLimitScripts(t *testing.T) {
	r, err := agg.New(300, 160, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatal(err)
	}
	r.SetResolution(100)
	if _, ok := r.MeasureMathGlyphRun("x", 23, "DejaVu Sans"); !ok {
		t.Skip("pixel-exact MathText glyph metrics unavailable")
	}

	layout, ok := core.LayoutMathText(r, `\lim_{x\to 0}`, 23, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	x := mathTextRunByIndex(t, layout.Runs, "x", 0)
	arrow := mathTextRunByIndex(t, layout.Runs, "→", 0)
	zero := mathTextRunByIndex(t, layout.Runs, "0", 0)

	// Matplotlib 3.10.9 parses \lim as an over/under function and suppresses
	// spaced-symbol padding inside its limit script. VectorParse at 23 pt / 100 dpi:
	// x=1.000000, arrow=14.256250, zero=33.025000.
	if got, want := arrow.Offset.X-x.Offset.X, 13.25625; math.Abs(got-want) > 0.75 {
		t.Fatalf(`\to spacing in limit script = %.3f, want %.3f; runs=%+v`, got, want, layout.Runs)
	}
	if got, want := zero.Offset.X-arrow.Offset.X, 18.76875; math.Abs(got-want) > 0.75 {
		t.Fatalf(`0 spacing after \to in limit script = %.3f, want %.3f; runs=%+v`, got, want, layout.Runs)
	}
}

func TestDebugMathTextMatrixLayout(t *testing.T) {
	r, err := agg.New(300, 160, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatal(err)
	}
	r.SetResolution(100)
	for _, tc := range []struct {
		expr string
		size float64
	}{
		{`\genfrac{(}{)}{0}{0}{a\quad b}{c\quad d}`, 25},
		{`\genfrac{[}{]}{0}{0}{1\quad 0}{0\quad 1}`, 25},
		{`\genfrac{(}{)}{0}{0}{x}{y}`, 24},
		{`\left\langle\genfrac{}{}{0}{0}{u}{v}\right\rangle`, 24},
	} {
		layout, ok := core.LayoutMathText(r, tc.expr, tc.size, "DejaVu Sans")
		if !ok {
			t.Fatalf("LayoutMathText(%q) returned !ok", tc.expr)
		}
		t.Logf("expr=%q width=%.6f ascent=%.6f descent=%.6f height=%.6f", tc.expr, layout.Width, layout.Ascent, layout.Descent, layout.Height)
		for _, run := range layout.Runs {
			t.Logf(" glyph %q size=%.6g x=%.6f y=%.6f font=%q", run.Text, run.FontSize, run.Offset.X, run.Offset.Y, run.FontKey)
		}
	}
}

func mathTextRunByIndex(t *testing.T, runs []core.MathTextLayoutRun, text string, index int) core.MathTextLayoutRun {
	t.Helper()
	seen := 0
	for _, run := range runs {
		if run.Text != text {
			continue
		}
		if seen == index {
			return run
		}
		seen++
	}
	t.Fatalf("missing run %q at index %d in %+v", text, index, runs)
	return core.MathTextLayoutRun{}
}

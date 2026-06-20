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

func TestLayoutMathTextLogitOneMinusLabelMatchesMatplotlib(t *testing.T) {
	r, err := agg.New(120, 80, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatal(err)
	}
	r.SetResolution(100)
	if _, ok := r.MeasureMathGlyphRun("1", 10, "DejaVu Sans"); !ok {
		t.Skip("pixel-exact MathText glyph metrics unavailable")
	}

	layout, ok := core.LayoutMathText(r, `\mathdefault{1-10^{-1}}`, 10, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	one := mathTextRunByIndex(t, layout.Runs, "1", 0)
	binaryMinus := mathTextRunByIndex(t, layout.Runs, "−", 0)
	secondOne := mathTextRunByIndex(t, layout.Runs, "1", 1)
	zero := mathTextRunByIndex(t, layout.Runs, "0", 0)
	exponentMinus := mathTextRunByIndex(t, layout.Runs, "−", 1)
	exponentOne := mathTextRunByIndex(t, layout.Runs, "1", 2)

	// Matplotlib 3.10.9 MathTextParser("path") VectorParse for
	// $\mathdefault{1-10^{-1}}$ at 10 pt / 100 dpi:
	// width=59, glyph x positions 0, 11.530884, 25.859802, 34.6875,
	// 43.649182, 51.787195. The exponent's unary minus is intentionally not
	// padded as a binary operator.
	check := func(name string, got, want float64) {
		t.Helper()
		if math.Abs(got-want) > 0.75 {
			t.Fatalf("%s x = %.3f, want %.3f; width=%.3f runs=%+v", name, got, want, layout.Width, layout.Runs)
		}
	}
	check("leading 1", one.Offset.X, 0)
	check("binary minus", binaryMinus.Offset.X, 11.530884)
	check("second 1", secondOne.Offset.X, 25.859802)
	check("zero", zero.Offset.X, 34.6875)
	check("exponent minus", exponentMinus.Offset.X, 43.649182)
	check("exponent 1", exponentOne.Offset.X, 51.787195)
	if math.Abs(layout.Width-59) > 1.0 {
		t.Fatalf("layout width = %.3f, want 59; runs=%+v", layout.Width, layout.Runs)
	}
}

func TestLayoutMathTextRulelessGenfracVectorMetricsMatchMatplotlib(t *testing.T) {
	r, err := agg.New(300, 160, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatal(err)
	}
	r.SetResolution(100)
	if _, ok := r.MeasureMathGlyphRun("x", 25, "DejaVu Sans"); !ok {
		t.Skip("pixel-exact MathText glyph metrics unavailable")
	}

	layout, ok := core.LayoutMathText(r, `\genfrac{(}{)}{0}{0}{a\quad b}{c\quad d}`, 25, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	// Matplotlib 3.10.9 VectorParse at 25 pt / 100 dpi:
	// width=122, height=64, depth=23. VectorParse's height includes depth, so
	// the layout ascent is approximately height-depth.
	if math.Abs(layout.Width-122) > 1.25 ||
		math.Abs(layout.Ascent-41) > 1.25 ||
		math.Abs(layout.Descent-23) > 1.25 {
		t.Fatalf("ruleless genfrac vector metrics = width %.3f ascent %.3f descent %.3f; want 122/41/23; runs=%+v",
			layout.Width, layout.Ascent, layout.Descent, layout.Runs)
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

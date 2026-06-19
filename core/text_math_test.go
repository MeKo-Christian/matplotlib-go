package core

import (
	"math"
	"strings"
	"testing"

	mt "github.com/cwbudde/mathtext"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestNormalizeDisplayText_ReplacesBasicMathTokens(t *testing.T) {
	got := normalizeDisplayText(`\\mu = 1.2, \\Delta x \\rightarrow \\pi`)
	want := "μ = 1.2, Δ x → π"
	if got != want {
		t.Fatalf("unexpected normalized text: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_ParsesInlineMath(t *testing.T) {
	got := normalizeDisplayText(`signal $\\alpha^2 + \\beta_i$ peak`)
	want := "signal α² + βᵢ peak"
	if got != want {
		t.Fatalf("unexpected inline math normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_FormatsFractionsAndRoots(t *testing.T) {
	got := normalizeDisplayText(`$\\frac{1}{\\sqrt{2}}$`)
	want := "1⁄√2"
	if got != want {
		t.Fatalf("unexpected fraction/root normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_HandlesGroupedScripts(t *testing.T) {
	got := normalizeDisplayText(`$x_{\\mathrm{max}}$`)
	want := "xₘₐₓ"
	if got != want {
		t.Fatalf("unexpected grouped subscript normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_HandlesAccentsAndOperators(t *testing.T) {
	got := normalizeDisplayText(`$\\hat{x} + \\sin(\\theta) \\leq \\overline{AB}$`)
	want := "x̂ + sin(θ) ≤ A̅B̅"
	if got != want {
		t.Fatalf("unexpected accent/operator normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_PreservesUnmatchedDollar(t *testing.T) {
	got := normalizeDisplayText(`cost is $5`)
	want := "cost is $5"
	if got != want {
		t.Fatalf("unexpected unmatched dollar normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_IgnoresLimitModifiers(t *testing.T) {
	got := normalizeDisplayText(`$\\displaystyle \\sum\\limits_{i=1}^n$`)
	want := "∑ᵢ₌₁ⁿ"
	if got != want {
		t.Fatalf("unexpected limit-modifier normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_HandlesMatrixEnvironment(t *testing.T) {
	got := normalizeDisplayText(`$\\begin{pmatrix} a & b \\\\ c & d \\end{pmatrix}$`)
	want := "(a b; c d)"
	if got != want {
		t.Fatalf("unexpected matrix normalization: got %q want %q", got, want)
	}
}

func TestNormalizeDisplayText_HandlesMiddleDelimiter(t *testing.T) {
	got := normalizeDisplayText(`$\\left\\langle{a}\\middle|b\\right\\rangle$`)
	want := "⟨a|b⟩"
	if got != want {
		t.Fatalf("unexpected middle-delimiter normalization: got %q want %q", got, want)
	}
}

func TestLayoutMathTextScripts(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `x_i^2`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if layout.Width <= 0 || layout.Ascent <= 0 || layout.Descent <= 0 || layout.Height != layout.Ascent+layout.Descent {
		t.Fatalf("invalid layout metrics: %+v", layout)
	}
	if len(layout.Rules) != 0 {
		t.Fatalf("unexpected rules in script-only layout: %+v", layout.Rules)
	}
	if !containsMathRun(layout.Runs, "x", 20) || !containsMathRun(layout.Runs, "i", 14) || !containsMathRun(layout.Runs, "2", 14) {
		t.Fatalf("missing expected script runs: %+v", layout.Runs)
	}

	var subY, superY float64
	for _, run := range layout.Runs {
		switch run.Text {
		case "i":
			subY = run.Offset.Y
		case "2":
			superY = run.Offset.Y
		}
	}
	if subY <= 0 || superY >= 0 {
		t.Fatalf("script baselines not shifted as expected: sub=%v super=%v runs=%+v", subY, superY, layout.Runs)
	}
}

func TestLayoutMathTextFractionAddsRule(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\\frac{1}{2}`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if len(layout.Rules) != 1 {
		t.Fatalf("expected one fraction rule, got %+v", layout.Rules)
	}
	if layout.Rules[0].Rect.Max.X <= layout.Rules[0].Rect.Min.X {
		t.Fatalf("unexpected fraction rule rect: %+v", layout.Rules[0].Rect)
	}
}

func TestLayoutMathTextSqrtHasVinculum(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\\sqrt[3]{x + 1}`, 18, "")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if len(layout.Rules) != 1 {
		t.Fatalf("expected sqrt rule, got %+v", layout.Rules)
	}
	// The radical is an AutoHeightChar (size/font depends on the renderer's
	// metrics); the root index is shrunk twice (SHRINK_FACTOR^2 = 0.49): 18*0.49 = 8.82.
	if !containsMathRunText(layout.Runs, "√") || !containsMathRun(layout.Runs, "3", 8.82) {
		t.Fatalf("missing sqrt/index runs: %+v", layout.Runs)
	}
	if layout.Rules[0].Rect.Min.X <= 0 || layout.Rules[0].Rect.Max.X <= layout.Rules[0].Rect.Min.X {
		t.Fatalf("unexpected sqrt rule rect: %+v", layout.Rules[0].Rect)
	}
}

func TestMathTextRasterMetricsUseMatplotlibShipCoordinates(t *testing.T) {
	r := mathRasterMetricRenderer{}
	layout := MathTextLayout{
		Width:   191.7080706490,
		Ascent:  35.9166666667,
		Descent: 7.0000000000,
		Height:  42.9166666667,
		Runs: []MathTextLayoutRun{
			{Text: "√", Offset: mt.Pt{X: 0.6923, Y: -6.0000}, FontSize: 12.6433},
			{Text: "x", Offset: mt.Pt{X: 23.1829, Y: 0.0799}, FontSize: 23.0000},
			{Text: "+", Offset: mt.Pt{X: 48.3704, Y: 0.0799}, FontSize: 23.0000},
			{Text: "1", Offset: mt.Pt{X: 81.4329, Y: 0.0799}, FontSize: 23.0000},
			{Text: "3", Offset: mt.Pt{X: 0.0000, Y: -20.4000}, FontSize: 11.2700},
			{Text: "+", Offset: mt.Pt{X: 112.0353, Y: 0.0000}, FontSize: 23.0000},
			{Text: "√", Offset: mt.Pt{X: 145.0978, Y: 1.0000}, FontSize: 13.4100},
			{Text: "y", Offset: mt.Pt{X: 168.7775, Y: 0.0799}, FontSize: 23.0000},
		},
		Rules: []MathTextLayoutRule{
			{Rect: mt.Rect{Min: mt.Pt{X: 19.1899, Y: -33.9201}, Max: mt.Pt{X: 105.7853, Y: -31.9236}}},
			{Rect: mt.Rect{Min: mt.Pt{X: 164.7845, Y: -28.9201}, Max: mt.Pt{X: 191.7081, Y: -26.9236}}},
		},
	}

	_, ascent, descent, ok := mathLayoutImageMetrics(&r, layout, "DejaVu Sans")
	if !ok {
		t.Fatal("mathLayoutImageMetrics returned !ok")
	}

	// Matplotlib 3.10.9 _mathtext.Output.to_raster for
	// $\sqrt[3]{x + 1} + \sqrt{y}$ at 23 pt / 100 dpi reports RasterParse
	// height=47.0763888889 and depth=9.0798611111, so the ascent is
	// height-depth=37.9965277778. The bbox is computed in ship coordinates:
	// glyph/rule y values are offset by box.height while the origin 0 is not.
	if math.Abs(ascent-37.9965277778) > 0.01 || math.Abs(descent-9.0798611111) > 0.01 {
		t.Fatalf("raster metrics ascent/descent = %.10f/%.10f, want 37.9965277778/9.0798611111", ascent, descent)
	}
}

func TestMixedMathTextLineMetricsIncludePlainLPDescent(t *testing.T) {
	r := inlineMathLineMetricRenderer{}
	layout := measureSingleLineTextLayoutParseMath(&r, `time $t$`, 10, "DejaVu Sans", true)

	// Matplotlib Text._get_layout asks the renderer for "lp" metrics with
	// ismath=False and applies h=max(line_h, lp_h), d=max(line_d, lp_d). For
	// "time $t$" at 10 pt / 100 dpi this keeps height=15 and raises descent
	// from the raw mathtext 2 px to the plain-font 3 px.
	if math.Abs(layout.Height-15) > 0.01 || math.Abs(layout.Descent-3) > 0.01 || math.Abs(layout.Ascent-12) > 0.01 {
		t.Fatalf("mixed math line metrics = height %.2f ascent %.2f descent %.2f, want 15/12/3",
			layout.Height, layout.Ascent, layout.Descent)
	}
}

func TestLayoutMathTextStacksLargeOperatorLimits(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\\sum\\limits_{i=1}^n`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if !containsMathRun(layout.Runs, "∑", 20) || !containsMathRun(layout.Runs, "i", 14) || !containsMathRun(layout.Runs, "=", 14) || !containsMathRun(layout.Runs, "1", 14) || !containsMathRun(layout.Runs, "n", 14) {
		t.Fatalf("missing expected limit runs: %+v", layout.Runs)
	}

	sumW := r.MeasureText("∑", 20, "DejaVu Sans Display").W
	superW := r.MeasureText("n", 14, "DejaVu Sans").W

	var sumX, subMinX, subMaxX, superX, subY, superY float64
	sawSub := false
	for _, run := range layout.Runs {
		switch run.Text {
		case "∑":
			sumX = run.Offset.X
		case "i", "=", "1":
			runW := r.MeasureText(run.Text, run.FontSize, run.FontKey).W
			if !sawSub || run.Offset.X < subMinX {
				subMinX = run.Offset.X
			}
			if run.Offset.X+runW > subMaxX {
				subMaxX = run.Offset.X + runW
			}
			subY = run.Offset.Y
			sawSub = true
		case "n":
			superX = run.Offset.X
			superY = run.Offset.Y
		}
	}

	if subY <= 0 || superY >= 0 {
		t.Fatalf("large-operator limits not stacked vertically: sub=%v super=%v runs=%+v", subY, superY, layout.Runs)
	}
	sumCenter := sumX + sumW/2
	subCenter := (subMinX + subMaxX) / 2
	superCenter := superX + superW/2
	// matplotlib rounds the HCentered glue per row (round(glue_set*cur_glue)), so
	// each row's center may differ from the operator center by up to ~0.5px.
	const centerTol = 0.6
	if math.Abs(subCenter-sumCenter) > centerTol || math.Abs(superCenter-sumCenter) > centerTol {
		t.Fatalf("large-operator limits not centered over operator: sumCenter=%v subCenter=%v superCenter=%v runs=%+v", sumCenter, subCenter, superCenter, layout.Runs)
	}
}

func TestLayoutMathTextSupportsIntegralLimits(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\int_0^\infty`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if !containsMathRun(layout.Runs, "∫", 20) || !containsMathRun(layout.Runs, "0", 14) || !containsMathRun(layout.Runs, "∞", 14) {
		t.Fatalf("missing expected integral runs: %+v", layout.Runs)
	}

	var intX, intW, subX, subY, superX, superY float64
	for _, run := range layout.Runs {
		switch run.Text {
		case "∫":
			intX = run.Offset.X
			intW = r.MeasureText(run.Text, run.FontSize, run.FontKey).W
		case "0":
			subX = run.Offset.X
			subY = run.Offset.Y
		case "∞":
			superX = run.Offset.X
			superY = run.Offset.Y
		}
	}

	if subY <= 0 || superY >= 0 {
		t.Fatalf("integral scripts should sit below/above the baseline: sub=%v super=%v runs=%+v", subY, superY, layout.Runs)
	}
	// Matplotlib applies a small negative kern for slanted drop-subscript
	// operators such as integrals; the scripts still live at the side, but may
	// begin just before the operator advance.
	if subX < intX+intW-1 || superX < intX+intW-1 {
		t.Fatalf("integral scripts should be side scripts, not stacked limits: intX=%v intW=%v subX=%v superX=%v runs=%+v", intX, intW, subX, superX, layout.Runs)
	}
}

func TestLayoutMathTextUsesInkBoundsForStackedMath(t *testing.T) {
	plain := textRecordingRenderer{}
	bounded := mathInkBoundsRenderer{}

	plainLayout, ok := LayoutMathText(&plain, `\genfrac{}{}{0}{0}{x}{y}`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("plain LayoutMathText returned !ok")
	}
	boundedLayout, ok := LayoutMathText(&bounded, `\genfrac{}{}{0}{0}{x}{y}`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("bounded LayoutMathText returned !ok")
	}

	var plainTopY, plainBottomY, boundedTopY, boundedBottomY float64
	for _, run := range plainLayout.Runs {
		switch run.Text {
		case "x":
			plainTopY = run.Offset.Y
		case "y":
			plainBottomY = run.Offset.Y
		}
	}
	for _, run := range boundedLayout.Runs {
		switch run.Text {
		case "x":
			boundedTopY = run.Offset.Y
		case "y":
			boundedBottomY = run.Offset.Y
		}
	}

	plainGap := plainBottomY - plainTopY
	boundedGap := boundedBottomY - boundedTopY
	// Ink-bounds measurement still tightens the x/y stack; the faithful 3.8.4
	// genfrac derives the rule thickness from the measured x-height, so the
	// synthetic plain/bounded delta is smaller than the old heuristic's.
	if boundedGap >= plainGap-1.5 {
		t.Fatalf("ink bounds should tighten stacked math: plain gap=%v bounded gap=%v plain=%+v bounded=%+v", plainGap, boundedGap, plainLayout.Runs, boundedLayout.Runs)
	}
}

func TestLayoutMathTextSupportsFencedDelimiters(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\left(\frac{1}{2}\right)`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	var leftSize, rightSize float64
	for _, run := range layout.Runs {
		if !strings.HasPrefix(run.FontKey, "STIXSize") {
			continue
		}
		switch {
		case leftSize == 0:
			leftSize = run.FontSize
		default:
			rightSize = run.FontSize
		}
	}
	if leftSize <= 20 || rightSize <= 20 {
		t.Fatalf("expected stretched delimiters larger than base size: left=%v right=%v runs=%+v", leftSize, rightSize, layout.Runs)
	}
}

func TestLayoutMathTextSupportsMiddleDelimiters(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\left\langle \frac{1}{2} \middle| x \right\rangle`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	var leftSize, rightSize float64
	var leftX, rightX float64
	for _, run := range layout.Runs {
		if !strings.HasPrefix(run.FontKey, "STIXSize") {
			continue
		}
		switch {
		case leftSize == 0:
			leftSize = run.FontSize
			leftX = run.Offset.X
		default:
			rightSize = run.FontSize
			rightX = run.Offset.X
		}
	}

	middleX := -1.0
	for _, rule := range layout.Rules {
		if rule.Rect.H() > 15 && rule.Rect.W() < 5 {
			middleX = rule.Rect.Min.X
			break
		}
	}

	if leftSize <= 20 || rightSize <= 20 || middleX < 0 {
		t.Fatalf("expected stretched fence delimiters: left=%v middleX=%v right=%v runs=%+v rules=%+v", leftSize, middleX, rightSize, layout.Runs, layout.Rules)
	}
	if leftX >= middleX || middleX >= rightX {
		t.Fatalf("expected middle delimiter to be between outer delimiters: left=%v middle=%v right=%v runs=%+v rules=%+v", leftX, middleX, rightX, layout.Runs, layout.Rules)
	}
}

func TestLayoutMathTextSupportsOmittedFenceDelimiters(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\left. x \right|`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	for _, run := range layout.Runs {
		if run.Text == "." {
			t.Fatalf("null delimiter should not render as a visible glyph: %+v", layout.Runs)
		}
	}

	sawX := false
	for _, run := range layout.Runs {
		if strings.TrimSpace(run.Text) == "x" {
			sawX = true
		}
	}
	sawBar := false
	for _, rule := range layout.Rules {
		if rule.Rect.H() > 15 && rule.Rect.W() < 5 {
			sawBar = true
			break
		}
	}
	if !sawX || !sawBar {
		t.Fatalf("missing expected fence output: runs=%+v rules=%+v", layout.Runs, layout.Rules)
	}
}

func TestLayoutMathTextSupportsStyleSwitches(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\mathrm{r} + \mathsf{s} + \mathtt{t}`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	var romanKey, sansKey, monoKey string
	for _, run := range layout.Runs {
		switch run.Text {
		case "r":
			romanKey = strings.ToLower(run.FontKey)
		case "s":
			sansKey = strings.ToLower(run.FontKey)
		case "t":
			monoKey = strings.ToLower(run.FontKey)
		}
	}

	if romanKey == "" || sansKey == "" || monoKey == "" {
		t.Fatalf("missing styled run font keys: %+v", layout.Runs)
	}
	if !strings.Contains(romanKey, "serif") {
		t.Fatalf("roman style did not resolve serif font key: %q", romanKey)
	}
	if !strings.Contains(sansKey, "sans") {
		t.Fatalf("sans style did not resolve sans font key: %q", sansKey)
	}
	if !strings.Contains(monoKey, "mono") {
		t.Fatalf("monospace style did not resolve mono font key: %q", monoKey)
	}
}

func TestLayoutMathTextUsesFontPropertiesMathFontFamily(t *testing.T) {
	var r textRecordingRenderer
	fontKey := render.FontPropertiesKey(render.FontProperties{
		Families:       []string{"DejaVu Sans"},
		MathFontFamily: "dejavuserif",
	})
	layout, ok := LayoutMathText(&r, `x + y`, 20, fontKey)
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	var xKey string
	for _, run := range layout.Runs {
		if strings.TrimSpace(run.Text) == "x" {
			xKey = strings.ToLower(run.FontKey)
			break
		}
	}
	if xKey == "" {
		t.Fatalf("missing x run: %+v", layout.Runs)
	}
	if !strings.Contains(xKey, "dejavuserif") && !strings.Contains(xKey, "dejavu serif") {
		t.Fatalf("math font family did not route default math through DejaVu Serif: %q", xKey)
	}
}

func TestLayoutMathTextSupportsSpacingCommands(t *testing.T) {
	var r textRecordingRenderer
	compact, ok := LayoutMathText(&r, `ab`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("compact LayoutMathText returned !ok")
	}
	wide, ok := LayoutMathText(&r, `a\quad b`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("wide LayoutMathText returned !ok")
	}
	if wide.Width <= compact.Width+8 {
		t.Fatalf("spacing command did not widen layout enough: compact=%v wide=%v", compact.Width, wide.Width)
	}
}

func TestLayoutMathTextSupportsMatrixEnvironments(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\begin{pmatrix} a & b \\\\ c & d \end{pmatrix}`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	var leftX, rightX, firstColTopY, firstColBottomY, secondColTopX float64
	var sawLeft, sawRight, sawA, sawB, sawC, sawD bool
	for _, run := range layout.Runs {
		text := strings.TrimSpace(run.Text)
		switch {
		case strings.HasPrefix(run.FontKey, "STIXSize") && !sawLeft:
			leftX = run.Offset.X
			sawLeft = true
		case strings.HasPrefix(run.FontKey, "STIXSize"):
			rightX = run.Offset.X
			sawRight = true
		case text == "a":
			firstColTopY = run.Offset.Y
			sawA = true
		case text == "b":
			if !sawA {
				t.Fatalf("expected a to be laid out before b: %+v", layout.Runs)
			}
			secondColTopX = run.Offset.X
			sawB = true
		case text == "c":
			firstColBottomY = run.Offset.Y
			sawC = true
		case text == "d":
			sawD = true
		}
	}

	if !sawLeft || !sawRight || !sawA || !sawB || !sawC || !sawD {
		t.Fatalf("missing expected matrix runs: %+v", layout.Runs)
	}
	if rightX <= leftX {
		t.Fatalf("expected right delimiter after left delimiter: left=%v right=%v runs=%+v", leftX, rightX, layout.Runs)
	}
	if secondColTopX <= leftX {
		t.Fatalf("expected second matrix column to be offset to the right: left=%v secondCol=%v runs=%+v", leftX, secondColTopX, layout.Runs)
	}
	if firstColBottomY <= firstColTopY {
		t.Fatalf("expected second matrix row below first row: top=%v bottom=%v runs=%+v", firstColTopY, firstColBottomY, layout.Runs)
	}
}

func TestLayoutMathTextSupportsArrayEnvironments(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := LayoutMathText(&r, `\begin{array}{cc} a & b \\\\ c & d \end{array}`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	for _, run := range layout.Runs {
		if run.Text == "(" || run.Text == ")" || run.Text == "[" || run.Text == "]" {
			t.Fatalf("array environment should not add fences: %+v", layout.Runs)
		}
	}

	var sawA, sawD bool
	for _, run := range layout.Runs {
		switch strings.TrimSpace(run.Text) {
		case "a":
			sawA = true
		case "d":
			sawD = true
		}
	}
	if !sawA || !sawD {
		t.Fatalf("missing expected array runs: %+v", layout.Runs)
	}
}

func TestLayoutDisplayTextMixedInlineMath(t *testing.T) {
	var r textRecordingRenderer
	layout, ok := layoutDisplayText(&r, `phase $\\frac{1}{2}$ peak`, 20, "DejaVu Sans")
	if !ok {
		t.Fatal("layoutDisplayText returned !ok")
	}
	if layout.Width <= 0 || layout.Ascent <= 0 || layout.Descent <= 0 {
		t.Fatalf("invalid layout metrics: %+v", layout)
	}
	if len(layout.Rules) != 1 {
		t.Fatalf("expected one fraction rule, got %+v", layout.Rules)
	}
	if !containsMathRun(layout.Runs, "phase ", 20) || !containsMathRun(layout.Runs, "1", 14) || !containsMathRun(layout.Runs, "2", 14) || !containsMathRun(layout.Runs, " peak", 20) {
		t.Fatalf("missing expected mixed inline runs: %+v", layout.Runs)
	}
}

func TestDrawDisplayTextUsesTeXRendererWhenEnabled(t *testing.T) {
	r := &texRecordingRenderer{}
	layout := measureSingleLineTextLayout(r, `signal $\alpha$`, 20, "DejaVu Sans", true)
	if layout.Width != 123 || layout.Ascent != 17 || layout.Descent != 5 {
		t.Fatalf("measureSingleLineTextLayout did not use TeX metrics: %+v", layout)
	}

	drawDisplayText(r, `signal $\alpha$`, geom.Pt{X: 10, Y: 20}, 20, render.Color{A: 1}, "DejaVu Sans", true)
	if len(r.texDraws) != 1 {
		t.Fatalf("expected one TeX draw, got %+v", r.texDraws)
	}
	if len(r.texts) != 0 {
		t.Fatalf("TeX-enabled draw should not fall back to DrawText, got %+v", r.texts)
	}
}

func TestDrawDisplayTextRotatedUsesTeXRendererWhenEnabled(t *testing.T) {
	r := &texRecordingRenderer{}

	drawDisplayTextRotated(r, `signal $\alpha$`, geom.Pt{X: 10, Y: 20}, 20, math.Pi/6, render.Color{A: 1}, "DejaVu Sans", true)

	if len(r.texRotatedDraws) != 1 {
		t.Fatalf("expected one rotated TeX draw, got %+v", r.texRotatedDraws)
	}
	if len(r.texts) != 0 {
		t.Fatalf("TeX-enabled rotated draw should not fall back to DrawText, got %+v", r.texts)
	}
}

func TestAlignedTextOrigin(t *testing.T) {
	anchor := geom.Pt{X: 100, Y: 50}
	metrics := render.TextMetrics{W: 40, Ascent: 8, Descent: 2}

	// Display space is y-up: top alignment puts the text top at the anchor, so the
	// baseline drops by the ascent (smaller Y): 50 - 8 = 42.
	got := alignedTextOrigin(anchor, metrics, TextAlignCenter, TextVAlignTop)
	if got.X != 80 || got.Y != 42 {
		t.Fatalf("unexpected text origin: %+v", got)
	}
}

func TestAxesTextDrawsNormalizedContent(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.5, 0.5, `\\alpha + \\beta`, TextOptions{
		HAlign:   TextAlignCenter,
		VAlign:   TextVAlignMiddle,
		FontSize: 12,
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 1 {
		t.Fatalf("expected exactly one text draw, got %d (%v)", len(r.texts), r.texts)
	}
	if r.texts[0] != "α + β" {
		t.Fatalf("unexpected rendered text %q", r.texts[0])
	}
}

func TestAxesTextDrawsFullMathLayoutRuns(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.5, 0.5, `$\\frac{1}{2}$`, TextOptions{
		HAlign:   TextAlignCenter,
		VAlign:   TextVAlignMiddle,
		FontSize: 12,
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if containsTextString(r.texts, "1⁄2") {
		t.Fatalf("full math expression fell back to normalized text draw: %v", r.texts)
	}
	if !containsTextString(r.texts, "1") || !containsTextString(r.texts, "2") {
		t.Fatalf("expected structured math runs for fraction, got %v", r.texts)
	}
	if r.pathCount == 0 {
		t.Fatalf("expected fraction rule path, got %d paths", r.pathCount)
	}
}

func TestAxesTextDrawsMixedInlineMathLayoutRuns(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.5, 0.5, `phase $\\frac{1}{2}$ peak`, TextOptions{
		HAlign:   TextAlignCenter,
		VAlign:   TextVAlignMiddle,
		FontSize: 12,
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if containsTextString(r.texts, "phase 1⁄2 peak") {
		t.Fatalf("mixed inline math fell back to normalized text draw: %v", r.texts)
	}
	if !containsTextString(r.texts, "phase ") || !containsTextString(r.texts, "1") || !containsTextString(r.texts, "2") || !containsTextString(r.texts, " peak") {
		t.Fatalf("expected stitched mixed inline runs, got %v", r.texts)
	}
	if r.pathCount == 0 {
		t.Fatalf("expected fraction rule path, got %d paths", r.pathCount)
	}
}

func TestDrawDisplayTextVerticalFullMathUsesPaths(t *testing.T) {
	var r verticalMathTextRecordingRenderer
	drawDisplayTextVertical(&r, `$\\frac{1}{2}$`, geom.Pt{X: 100, Y: 60}, 12, render.Color{A: 1}, "DejaVu Sans")

	if len(r.verticalTexts) != 0 {
		t.Fatalf("full math expression unexpectedly used DrawTextVertical fallback: %v", r.verticalTexts)
	}
	if !containsTextString(r.textPathCalls, "1") || !containsTextString(r.textPathCalls, "2") {
		t.Fatalf("expected fraction runs to resolve through TextPath, got %v", r.textPathCalls)
	}
	if r.pathCount < 3 {
		t.Fatalf("expected fraction rule plus glyph paths, got %d paths", r.pathCount)
	}
}

func TestDrawDisplayTextRotatedMixedInlineMathUsesPaths(t *testing.T) {
	var r rotatedMathTextRecordingRenderer
	drawDisplayTextRotated(&r, `amp $\\frac{1}{2}$`, geom.Pt{X: 100, Y: 60}, 12, math.Pi/2, render.Color{A: 1}, "DejaVu Sans")

	if len(r.rotatedTexts) != 0 {
		t.Fatalf("mixed inline math unexpectedly used DrawTextRotated fallback: %v", r.rotatedTexts)
	}
	if !containsTextString(r.textPathCalls, "amp ") || !containsTextString(r.textPathCalls, "1") || !containsTextString(r.textPathCalls, "2") {
		t.Fatalf("expected mixed inline math runs to resolve through TextPath, got %v", r.textPathCalls)
	}
	if r.pathCount < 3 {
		t.Fatalf("expected fraction rule plus glyph paths, got %d paths", r.pathCount)
	}
}

func TestAxesLabelsDrawMathTextAccordingToExpressionScope(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle(`$\\alpha^2$`)
	ax.SetXLabel(`phase $\\theta$`)
	ax.SetYLabel(`amp $\\frac{1}{2}$`)

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if containsTextString(r.texts, "α²") {
		t.Fatalf("full math title unexpectedly collapsed to normalized text: %v", r.texts)
	}
	if !containsTextString(r.texts, "α") || !containsTextString(r.texts, "2") {
		t.Fatalf("missing structured title math runs: %v", r.texts)
	}
	if containsTextString(r.texts, "phase θ") {
		t.Fatalf("mixed inline xlabel unexpectedly collapsed to normalized text: %v", r.texts)
	}
	if !containsTextString(r.texts, "phase ") || !containsTextString(r.texts, "θ") {
		t.Fatalf("missing structured xlabel runs: %v", r.texts)
	}
	if containsTextString(r.texts, "amp 1⁄2") {
		t.Fatalf("mixed inline ylabel unexpectedly collapsed to normalized text: %v", r.texts)
	}
	if !containsTextString(r.texts, "amp ") || !containsTextString(r.texts, "1") || !containsTextString(r.texts, "2") {
		t.Fatalf("missing structured ylabel runs: %v", r.texts)
	}
}

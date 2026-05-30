package core

import (
	"image"
	"math"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
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
	if math.Abs(subCenter-sumCenter) > 0.01 || math.Abs(superCenter-sumCenter) > 0.01 {
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

func TestDrawDisplayTextUsesExplicitFontDrawer(t *testing.T) {
	r := &fontAwareTextRecordingRenderer{}

	drawDisplayText(r, "plain", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1}, "DejaVu Sans")

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware text draw, got %+v", r.fontTextCalls)
	}
	if r.fontTextCalls[0].fontKey != "DejaVu Sans" {
		t.Fatalf("font-aware draw fontKey = %q, want DejaVu Sans", r.fontTextCalls[0].fontKey)
	}
	if len(r.texts) != 0 {
		t.Fatalf("legacy DrawText should not be used when font-aware draw is available, got %+v", r.texts)
	}
}

func TestDrawDisplayTextRotatedUsesExplicitFontDrawer(t *testing.T) {
	r := &fontAwareTextRecordingRenderer{}

	drawDisplayTextRotated(r, "rotated", geom.Pt{X: 10, Y: 20}, 12, math.Pi/8, render.Color{A: 1}, "DejaVu Sans")

	if len(r.fontRotatedCalls) != 1 {
		t.Fatalf("expected one font-aware rotated text draw, got %+v", r.fontRotatedCalls)
	}
	if r.fontRotatedCalls[0].fontKey != "DejaVu Sans" {
		t.Fatalf("font-aware rotated draw fontKey = %q, want DejaVu Sans", r.fontRotatedCalls[0].fontKey)
	}
	if len(r.texts) != 0 {
		t.Fatalf("legacy DrawTextRotated should not be used when font-aware draw is available, got %+v", r.texts)
	}
}

func TestDrawDisplayTextVerticalUsesExplicitFontDrawer(t *testing.T) {
	r := &fontAwareTextRecordingRenderer{}

	drawDisplayTextVertical(r, "vertical", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1}, "DejaVu Sans")

	if len(r.fontVerticalCalls) != 1 {
		t.Fatalf("expected one font-aware vertical text draw, got %+v", r.fontVerticalCalls)
	}
	if r.fontVerticalCalls[0].fontKey != "DejaVu Sans" {
		t.Fatalf("font-aware vertical draw fontKey = %q, want DejaVu Sans", r.fontVerticalCalls[0].fontKey)
	}
	if len(r.verticalCalls) != 0 {
		t.Fatalf("legacy DrawTextVertical should not be used when font-aware draw is available, got %+v", r.verticalCalls)
	}
}

func TestTextArtistUsesTeXRendererWhenRCUseTeX(t *testing.T) {
	fig := NewFigure(320, 240)
	fig.RC.UseTeX = true
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
	ax.Text(0.5, 0.5, `signal $\alpha$`, TextOptions{FontSize: 20})

	r := &texRecordingRenderer{}
	DrawFigure(fig, r)

	if len(r.texDraws) != 1 {
		t.Fatalf("expected text artist to draw through TeX renderer, got %+v", r.texDraws)
	}
	if len(r.texts) != 0 {
		t.Fatalf("TeX-enabled text artist should not fall back to DrawText, got %+v", r.texts)
	}
}

func TestTextArtistCanDisableMathParsing(t *testing.T) {
	ctx := createTestDrawContext()
	parseMath := false
	text := &Text{
		Position:  geom.Pt{X: 1, Y: 1},
		Content:   `signal $\alpha$`,
		FontSize:  12,
		ParseMath: &parseMath,
		ClipOn:    true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one plain text draw, got %+v", r.fontTextCalls)
	}
	if got, want := r.fontTextCalls[0].text, `signal $\alpha$`; got != want {
		t.Fatalf("parse_math disabled text = %q, want %q", got, want)
	}
}

func TestTextArtistAlphaAppliesToDrawnText(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "plain",
		FontSize: 12,
		Color:    render.Color{R: 0.2, G: 0.4, B: 0.6, A: 0.8},
		ClipOn:   true,
	}
	text.SetAlpha(0.5)
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.textColors) != 1 {
		t.Fatalf("expected one drawn text color, got %+v", r.textColors)
	}
	if !approx(r.textColors[0].A, 0.4, 1e-12) {
		t.Fatalf("text alpha = %v, want local alpha multiplied by artist alpha", r.textColors[0].A)
	}
}

func TestTextArtistWrapWidthUsesMultilineLayoutAndBBox(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:  geom.Pt{X: 1, Y: 1},
		Content:   "alpha beta gamma",
		FontSize:  10,
		WrapWidth: 52,
		ClipOn:    true,
		BBox: &TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{A: 1},
			Padding:   1,
		},
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.texts) != 2 || r.texts[0] != "alpha beta" || r.texts[1] != "gamma" {
		t.Fatalf("wrapped text lines = %v, want [alpha beta] [gamma]", r.texts)
	}
	if len(r.pathCalls) == 0 {
		t.Fatal("expected wrapped text bbox path")
	}
	bounds, ok := pathBounds(r.pathCalls[0].path)
	if !ok {
		t.Fatalf("missing bbox path bounds: %+v", r.pathCalls[0].path)
	}
	if bounds.W() > text.WrapWidth {
		t.Fatalf("wrapped bbox width = %v, want <= wrap width %v", bounds.W(), text.WrapWidth)
	}
}

func TestTextArtistWrapUsesFigureBoxWidth(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.FigureRect = geom.Rect{Max: geom.Pt{X: 100, Y: 100}}
	text := &Text{
		Position: geom.Pt{X: 0.4, Y: 0.5},
		Coords:   Coords(CoordFigure),
		Content:  "alpha beta gamma",
		FontSize: 10,
		HAlign:   TextAlignLeft,
		Wrap:     true,
		ClipOn:   true,
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.texts) != 2 || r.texts[0] != "alpha beta" || r.texts[1] != "gamma" {
		t.Fatalf("auto-wrapped text lines = %v, want [alpha beta] [gamma]", r.texts)
	}
}

func TestRotatedTextBBoxRotatesWithText(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "tilt",
		FontSize: 10,
		Angle:    45,
		ClipOn:   true,
		BBox: &TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{A: 1},
			Padding:   1,
		},
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.pathCalls) == 0 {
		t.Fatal("expected rotated text bbox path")
	}
	path := r.pathCalls[0].path
	if len(path.V) < 4 {
		t.Fatalf("bbox path vertices = %+v, want rotated rectangle vertices", path.V)
	}
	if approx(path.V[0].Y, path.V[1].Y, 1e-9) || approx(path.V[1].X, path.V[2].X, 1e-9) {
		t.Fatalf("rotated text bbox remained axis-aligned: %+v", path.V[:4])
	}
}

func TestRotatedTextBBoxUsesDisplayRotationSign(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "tilt",
		FontSize: 10,
		Angle:    45,
		ClipOn:   true,
		BBox: &TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{A: 1},
			Padding:   1,
		},
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.pathCalls) == 0 || len(r.pathCalls[0].path.V) < 2 {
		t.Fatalf("expected rotated bbox path, got %+v", r.pathCalls)
	}
	edge := geom.Pt{
		X: r.pathCalls[0].path.V[1].X - r.pathCalls[0].path.V[0].X,
		Y: r.pathCalls[0].path.V[1].Y - r.pathCalls[0].path.V[0].Y,
	}
	if edge.X <= 0 || edge.Y >= 0 {
		t.Fatalf("positive text rotation should tilt bbox upward in display coordinates, edge=%+v path=%+v", edge, r.pathCalls[0].path.V[:4])
	}
}

func TestTextBBoxUsesLineBoxWhenInkBoundsAreShort(t *testing.T) {
	ctx := createTestDrawContext()
	r := &mathInkBoundsRenderer{}
	layout := measureSingleLineTextLayout(r, "bbox", 10, "", false, ctx.RC.UseTeX)
	origin := geom.Pt{X: 100, Y: 100}

	got, ok := textBBoxRect(origin, layout, &TextBBoxOptions{
		FaceColor: render.Color{A: 1},
		EdgeColor: render.Color{A: 1},
		Padding:   1,
	}, ctx, 10)

	if !ok {
		t.Fatal("textBBoxRect returned !ok")
	}
	want := geom.Rect{
		Min: geom.Pt{X: 99, Y: 97},
		Max: geom.Pt{X: 121, Y: 109},
	}
	if !approxRect(got, want, 1e-9) {
		t.Fatalf("text bbox = %+v, want line box %+v", got, want)
	}
}

func TestTextRotationModeAnchorRotatesAroundAlignedTextBox(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:     geom.Pt{X: 1, Y: 1},
		Content:      "tilt",
		FontSize:     10,
		HAlign:       TextAlignLeft,
		VAlign:       TextVAlignTop,
		Angle:        45,
		RotationMode: TextRotationModeAnchor,
		ClipOn:       true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 1 {
		t.Fatalf("expected one rotated text draw, got %+v", r.fontRotatedCalls)
	}
	anchor := transformedPoint(ctx, text.Coords, text.Position, text.OffsetX, text.OffsetY)
	layout := measureSingleLineTextLayoutParseMath(r, text.Content, text.FontSize, text.FontKey, true, ctx.RC.UseTeX)
	origin := alignedSingleLineOrigin(anchor, layout, text.HAlign, layoutVerticalAlign(text.VAlign, false))
	want := geom.Pt{
		X: origin.X + layout.Width/2,
		Y: origin.Y + layout.Descent,
	}
	if !approx(r.fontRotatedCalls[0].anchor.X, want.X, 1e-9) || !approx(r.fontRotatedCalls[0].anchor.Y, want.Y, 1e-9) {
		t.Fatalf("rotation_mode anchor draw anchor = %+v, want pre-rotation bottom-center %+v", r.fontRotatedCalls[0].anchor, want)
	}
	defaultAnchor := tickLabelRotationAnchor(origin, layout, text.HAlign, layoutVerticalAlign(text.VAlign, false), text.Angle*math.Pi/180)
	if approx(r.fontRotatedCalls[0].anchor.X, defaultAnchor.X, 1e-9) && approx(r.fontRotatedCalls[0].anchor.Y, defaultAnchor.Y, 1e-9) {
		t.Fatalf("rotation_mode anchor unexpectedly matched default rotated-bbox anchor %+v", defaultAnchor)
	}
}

func TestTextRotationModeXTickAdjustsHorizontalAlignment(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:     geom.Pt{X: 1, Y: 1},
		Content:      "tick",
		FontSize:     10,
		HAlign:       TextAlignCenter,
		VAlign:       TextVAlignBottom,
		Angle:        45,
		RotationMode: TextRotationModeXTick,
		ClipOn:       true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 1 {
		t.Fatalf("expected one rotated text draw, got %+v", r.fontRotatedCalls)
	}
	anchor := transformedPoint(ctx, text.Coords, text.Position, text.OffsetX, text.OffsetY)
	layout := measureSingleLineTextLayoutParseMath(r, text.Content, text.FontSize, text.FontKey, true, ctx.RC.UseTeX)
	vAlign := layoutVerticalAlign(text.VAlign, false)
	wantOrigin := alignedSingleLineOrigin(anchor, layout, TextAlignLeft, vAlign)
	want := tickLabelRotationAnchor(wantOrigin, layout, TextAlignLeft, vAlign, text.Angle*math.Pi/180)
	if !approx(r.fontRotatedCalls[0].anchor.X, want.X, 1e-9) || !approx(r.fontRotatedCalls[0].anchor.Y, want.Y, 1e-9) {
		t.Fatalf("xtick rotation anchor = %+v, want left-aligned anchor %+v", r.fontRotatedCalls[0].anchor, want)
	}
}

func TestTextRotationModeYTickAdjustsVerticalAlignment(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:     geom.Pt{X: 1, Y: 1},
		Content:      "tick",
		FontSize:     10,
		HAlign:       TextAlignLeft,
		VAlign:       TextVAlignMiddle,
		Angle:        45,
		RotationMode: TextRotationModeYTick,
		ClipOn:       true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 1 {
		t.Fatalf("expected one rotated text draw, got %+v", r.fontRotatedCalls)
	}
	anchor := transformedPoint(ctx, text.Coords, text.Position, text.OffsetX, text.OffsetY)
	layout := measureSingleLineTextLayoutParseMath(r, text.Content, text.FontSize, text.FontKey, true, ctx.RC.UseTeX)
	wantVAlign := textLayoutVAlignBaseline
	wantOrigin := alignedSingleLineOrigin(anchor, layout, TextAlignLeft, wantVAlign)
	want := tickLabelRotationAnchor(wantOrigin, layout, TextAlignLeft, wantVAlign, text.Angle*math.Pi/180)
	if !approx(r.fontRotatedCalls[0].anchor.X, want.X, 1e-9) || !approx(r.fontRotatedCalls[0].anchor.Y, want.Y, 1e-9) {
		t.Fatalf("ytick rotation anchor = %+v, want baseline-aligned anchor %+v", r.fontRotatedCalls[0].anchor, want)
	}
}

func TestTextCenterBaselineVerticalAlignment(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 2},
		Content:  "center",
		FontSize: 10,
		HAlign:   TextAlignLeft,
		VAlign:   TextVAlignCenterBaseline,
		Coords:   Coords(CoordData),
		ClipOn:   true,
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.origins) != 1 {
		t.Fatalf("text draws = %d, want 1", len(r.origins))
	}
	anchor := ctx.DataToPixel.Apply(text.Position)
	layout := measureSingleLineTextLayout(r, text.Content, text.FontSize, "")
	// Display space is y-up: centering the baseline lowers the origin by half the
	// ascent (smaller Y), so the offset is subtracted.
	want := geom.Pt{X: anchor.X, Y: anchor.Y - layout.Ascent/2}
	if !approx(r.origins[0].X, want.X, 1e-9) || !approx(r.origins[0].Y, want.Y, 1e-9) {
		t.Fatalf("center-baseline origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestTextArtistFontKeyOverridesRCFontKey(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.FontKey = "RC Font"
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "plain",
		FontSize: 12,
		FontKey:  "Artist Font",
		ClipOn:   true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware text draw, got %+v", r.fontTextCalls)
	}
	if got := r.fontTextCalls[0].fontKey; got != "Artist Font" {
		t.Fatalf("text fontKey = %q, want artist override", got)
	}
}

func TestTextArtistFontPropertiesOverrideRCFontKey(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.FontKey = "RC Font"
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "plain",
		FontSize: 12,
		FontProperties: &render.FontProperties{
			Families: []string{"DejaVu Serif"},
			Style:    render.FontStyleItalic,
			Weight:   700,
		},
		ClipOn: true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware text draw, got %+v", r.fontTextCalls)
	}
	props := render.ParseFontProperties(r.fontTextCalls[0].fontKey)
	if props.Style != render.FontStyleItalic || props.Weight != 700 || len(props.Families) != 1 || props.Families[0] != "DejaVu Serif" {
		t.Fatalf("text font properties = %+v, want DejaVu Serif italic 700", props)
	}
}

func TestTextArtistFontPropertiesRouteFeatureOptions(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.FontKey = "RC Font"
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "fi",
		FontSize: 12,
		FontProperties: &render.FontProperties{
			Families: []string{"DejaVu Sans"},
			Stretch:  "condensed",
			Variant:  "small-caps",
			Language: "de",
			Features: []render.TextFeature{{Tag: "liga", Value: 0}},
		},
		ClipOn: true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware text draw, got %+v", r.fontTextCalls)
	}
	props := render.ParseFontProperties(r.fontTextCalls[0].fontKey)
	if props.Stretch != "condensed" || props.Variant != "small-caps" || props.Language != "de" {
		t.Fatalf("text extended font properties = %+v, want condensed small-caps de", props)
	}
	if len(props.Features) != 1 || props.Features[0] != (render.TextFeature{Tag: "liga", Value: 0}) {
		t.Fatalf("text font features = %+v, want liga=0", props.Features)
	}
}

func TestAnnotationFontKeyOverridesRCFontKey(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.FontKey = "RC Font"
	annotation := &Annotation{
		Point:    geom.Pt{X: 1, Y: 1},
		Content:  "note",
		OffsetX:  10,
		OffsetY:  -8,
		FontSize: 12,
		FontKey:  "Annotation Font",
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware annotation draw, got %+v", r.fontTextCalls)
	}
	if got := r.fontTextCalls[0].fontKey; got != "Annotation Font" {
		t.Fatalf("annotation fontKey = %q, want annotation override", got)
	}
}

func TestAnnotationFontPropertiesOverrideRCFontKey(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.FontKey = "RC Font"
	annotation := &Annotation{
		Point:    geom.Pt{X: 1, Y: 1},
		Content:  "note",
		OffsetX:  10,
		OffsetY:  -8,
		FontSize: 12,
		FontProperties: &render.FontProperties{
			Families: []string{"DejaVu Sans Mono"},
			Style:    render.FontStyleOblique,
			Weight:   600,
		},
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware annotation draw, got %+v", r.fontTextCalls)
	}
	props := render.ParseFontProperties(r.fontTextCalls[0].fontKey)
	if props.Style != render.FontStyleOblique || props.Weight != 600 || len(props.Families) != 1 || props.Families[0] != "DejaVu Sans Mono" {
		t.Fatalf("annotation font properties = %+v, want DejaVu Sans Mono oblique 600", props)
	}
}

func TestAnnotationCanDisableMathParsing(t *testing.T) {
	ctx := createTestDrawContext()
	parseMath := false
	annotation := &Annotation{
		Point:     geom.Pt{X: 1, Y: 1},
		Content:   `note $\beta$`,
		OffsetX:   10,
		OffsetY:   -8,
		FontSize:  12,
		ParseMath: &parseMath,
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one plain annotation draw, got %+v", r.fontTextCalls)
	}
	if got, want := r.fontTextCalls[0].text, `note $\beta$`; got != want {
		t.Fatalf("parse_math disabled annotation = %q, want %q", got, want)
	}
}

func TestAnnotationAlphaAppliesToTextAndArrow(t *testing.T) {
	ctx := createTestDrawContext()
	arrowStyle, _ := ArrowStyleFromString("-|>")
	connectionStyle, _ := ConnectionStyleFromString("arc3")
	annotation := &Annotation{
		Point:           geom.Pt{X: 1, Y: 1},
		Content:         "note",
		OffsetX:         10,
		OffsetY:         -8,
		FontSize:        12,
		Color:           render.Color{R: 0.2, G: 0.4, B: 0.6, A: 0.6},
		ArrowColor:      render.Color{R: 0.8, G: 0.1, B: 0.1, A: 0.8},
		ArrowWidth:      1.25,
		ArrowHeadSize:   8,
		ArrowStyle:      arrowStyle,
		ConnectionStyle: connectionStyle,
	}
	annotation.SetAlpha(0.5)
	r := &textRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.textColors) != 1 {
		t.Fatalf("expected one annotation text color, got %+v", r.textColors)
	}
	if !approx(r.textColors[0].A, 0.3, 1e-12) {
		t.Fatalf("annotation text alpha = %v, want local alpha multiplied by artist alpha", r.textColors[0].A)
	}
	if !hasPaintAlpha(r.pathPaints, 0.4) {
		t.Fatalf("annotation arrow paints should include artist-multiplied alpha 0.4, got %+v", r.pathPaints)
	}
}

func TestAnnotationClipSkipsOutsideAnnotatedPoint(t *testing.T) {
	ctx := createTestDrawContext()
	clip := true
	annotation := &Annotation{
		Point:          geom.Pt{X: 100, Y: 100},
		Content:        "outside",
		OffsetX:        10,
		OffsetY:        -8,
		FontSize:       12,
		AnnotationClip: &clip,
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontTextCalls) != 0 || len(r.texts) != 0 {
		t.Fatalf("clipped annotation should not draw text, got font=%+v text=%+v", r.fontTextCalls, r.texts)
	}
}

func TestAnnotationClipFalseDrawsOutsideAnnotatedPoint(t *testing.T) {
	ctx := createTestDrawContext()
	clip := false
	annotation := &Annotation{
		Point:          geom.Pt{X: 100, Y: 100},
		Content:        "outside",
		OffsetX:        10,
		OffsetY:        -8,
		FontSize:       12,
		AnnotationClip: &clip,
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("annotation_clip=false should draw text, got %+v", r.fontTextCalls)
	}
}

func TestAnnotationClipDefaultMatchesMatplotlibDataOnlyPolicy(t *testing.T) {
	ctx := createTestDrawContext()
	dataAnnotation := &Annotation{
		Point:    geom.Pt{X: 100, Y: 100},
		Content:  "outside data",
		OffsetX:  10,
		OffsetY:  -8,
		FontSize: 12,
		Coords:   Coords(CoordData),
	}
	axesAnnotation := &Annotation{
		Point:    geom.Pt{X: 1.5, Y: 1.5},
		Content:  "outside axes",
		OffsetX:  10,
		OffsetY:  -8,
		FontSize: 12,
		Coords:   Coords(CoordAxes),
	}
	r := &fontAwareTextRecordingRenderer{}

	dataAnnotation.DrawOverlay(r, ctx)
	axesAnnotation.DrawOverlay(r, ctx)

	if containsFontTextCall(r.fontTextCalls, "outside data") || containsTextString(r.texts, "outside data") {
		t.Fatalf("annotation_clip default should clip outside data-coordinate annotations, got font=%+v text=%+v", r.fontTextCalls, r.texts)
	}
	if !containsFontTextCall(r.fontTextCalls, "outside axes") && !containsTextString(r.texts, "outside axes") {
		t.Fatalf("annotation_clip default should draw outside non-data annotations, got font=%+v text=%+v", r.fontTextCalls, r.texts)
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

func TestAnnotationDrawOverlayRendersArrowAndText(t *testing.T) {
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

	ax.Annotate("peak", 0.5, 0.5)

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 1 || r.texts[0] != "peak" {
		t.Fatalf("unexpected annotation texts: %v", r.texts)
	}
	if r.pathCount < 2 {
		t.Fatalf("expected annotation arrow line and head, got %d paths", r.pathCount)
	}
}

func TestAnnotationDrawsTextBBox(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	face := render.Color{R: 1, G: 0.95, B: 0.8, A: 1}
	edge := render.Color{R: 0.2, G: 0.1, B: 0.05, A: 1}
	ax.Annotate("boxed", 0.5, 0.5, AnnotationOptions{
		OffsetX: 10,
		OffsetY: -8,
		BBox: &TextBBoxOptions{
			FaceColor:    face,
			EdgeColor:    edge,
			LineWidth:    2,
			Padding:      3,
			CornerRadius: 4,
		},
	})

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	if !containsTextString(r.texts, "boxed") {
		t.Fatalf("expected annotation text to draw, got %v", r.texts)
	}
	if !hasPathPaint(r.pathPaints, face, edge, 2) {
		t.Fatalf("annotation bbox paint not found in %+v", r.pathPaints)
	}
}

func TestAnnotationArrowHeadSizeUsesPointMutationScale(t *testing.T) {
	ctx := createTestDrawContext()
	arrow, _ := ArrowStyleFromString("-|>,head_length=0.35,head_width=0.20")
	annotation := &Annotation{
		ArrowWidth:      1.2,
		ArrowHeadSize:   9,
		ArrowStyle:      arrow,
		ConnectionStyle: ConnectionStyle{Name: "arc3"},
	}
	r := &textRecordingRenderer{}

	annotation.drawArrow(r, ctx, geom.Pt{X: 100, Y: 100}, geom.Pt{X: 200, Y: 100})

	if len(r.pathCalls) < 2 {
		t.Fatalf("expected line and arrow head paths, got %+v", r.pathCalls)
	}
	headBounds, ok := pathBounds(r.pathCalls[len(r.pathCalls)-1].path)
	if !ok {
		t.Fatalf("arrow head path has no bounds: %+v", r.pathCalls[len(r.pathCalls)-1].path)
	}
	scale := pointsToPixels(ctx.RC, 9)
	wantWidth := 0.35 * scale
	wantHeight := 2 * 0.20 * scale
	if !approx(headBounds.Max.X-headBounds.Min.X, wantWidth, 1e-9) ||
		!approx(headBounds.Max.Y-headBounds.Min.Y, wantHeight, 1e-9) {
		t.Fatalf("arrow head bounds = %+v, want width %.12g height %.12g from point mutation scale", headBounds, wantWidth, wantHeight)
	}
}

func TestAnnotationArrowDefaultShrinkUsesPointUnits(t *testing.T) {
	ctx := createTestDrawContext()
	annotation := &Annotation{
		ArrowWidth:      1,
		ArrowHeadSize:   9,
		ArrowStyle:      ArrowStyle{Name: "-"},
		ConnectionStyle: ConnectionStyle{Name: "arc3"},
	}
	r := &textRecordingRenderer{}
	start := geom.Pt{X: 100, Y: 100}
	target := geom.Pt{X: 200, Y: 100}

	annotation.drawArrow(r, ctx, start, target)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one plain connection path, got %+v", r.pathCalls)
	}
	path := r.pathCalls[0].path
	if len(path.V) != 3 {
		t.Fatalf("connection path vertices = %+v, want quadratic path", path.V)
	}
	shrink := pointsToPixels(ctx.RC, 2)
	if !approx(path.V[0].X, start.X+shrink, 1e-9) || !approx(path.V[0].Y, start.Y, 1e-9) {
		t.Fatalf("shrunk start = %+v, want %+v", path.V[0], geom.Pt{X: start.X + shrink, Y: start.Y})
	}
	if !approx(path.V[2].X, target.X-shrink, 1e-9) || !approx(path.V[2].Y, target.Y, 1e-9) {
		t.Fatalf("shrunk end = %+v, want %+v", path.V[2], geom.Pt{X: target.X - shrink, Y: target.Y})
	}
}

func TestAnnotationArrowStartsFromTextBoxRelposBeforePatchClip(t *testing.T) {
	ctx := createTestDrawContext()
	annotation := &Annotation{
		Point:           geom.Pt{X: 0.2, Y: 0.8},
		Content:         "box",
		OffsetX:         120,
		OffsetY:         80,
		FontSize:        10,
		Coords:          Coords(CoordFigure),
		ArrowWidth:      1,
		ArrowHeadSize:   9,
		ArrowStyle:      ArrowStyle{Name: "-"},
		ConnectionStyle: ConnectionStyle{Name: "arc3"},
		HAlign:          TextAlignCenter,
		VAlign:          TextVAlignMiddle,
		BBox: &TextBBoxOptions{
			Padding:   4,
			LineWidth: 1,
		},
	}
	r := &textRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.pathCalls) == 0 {
		t.Fatal("expected annotation arrow path")
	}
	target := transformedPoint(ctx, annotation.Coords, annotation.Point, 0, 0)
	anchor := transformedPoint(ctx, annotation.Coords, annotation.Point, annotation.OffsetX, annotation.OffsetY)
	layout := measureSingleLineTextLayout(r, annotation.Content, annotation.FontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	origin := alignedSingleLineOrigin(anchor, layout, annotation.HAlign, layoutVerticalAlign(annotation.VAlign, false))
	box, ok := textBBoxRect(origin, layout, annotation.BBox, ctx, annotation.FontSize)
	if !ok {
		t.Fatal("expected annotation text bbox")
	}
	raw := annotation.ConnectionStyle.connect(rectCenter(box), target, 0, 0)
	boundary, ok := connectionPatchBoundaryPoint(raw, pixelRectPath(box).Interpolated(8).V, true)
	if !ok {
		t.Fatalf("expected center-to-target path to leave bbox: box=%+v path=%+v", box, raw)
	}
	raw.V[0] = boundary
	want := shrinkPathEndpoints(raw, arrowShrinkPixels(ctx, 2), arrowShrinkPixels(ctx, 2))
	got := r.pathCalls[0].path
	if len(got.V) != len(want.V) || distance(got.V[0], want.V[0]) > 1e-9 {
		t.Fatalf("annotation arrow start = %+v, want clipped relpos start %+v (box=%+v target=%+v)", got.V, want.V, box, target)
	}
}

func TestAnnotationAngleUsesRotatedTextDrawer(t *testing.T) {
	ctx := createTestDrawContext()
	annotation := &Annotation{
		Point:    geom.Pt{X: 1, Y: 1},
		Content:  "tilt",
		OffsetX:  10,
		FontSize: 10,
		Angle:    35,
		Coords:   Coords(CoordData),
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontRotatedCalls) != 1 {
		t.Fatalf("expected one rotated annotation text draw, got %+v", r.fontRotatedCalls)
	}
	if len(r.fontTextCalls) != 0 || len(r.texts) != 0 {
		t.Fatalf("rotated annotation should not use unrotated text draws, font=%+v legacy=%+v", r.fontTextCalls, r.texts)
	}
}

func TestAnnotationMultilineSplitsTextDraws(t *testing.T) {
	ctx := createTestDrawContext()
	annotation := &Annotation{
		Point:    geom.Pt{X: 1, Y: 1},
		Content:  "top\nbottom",
		OffsetX:  10,
		FontSize: 10,
		Coords:   Coords(CoordData),
	}
	annotation.SetAlpha(0.5)
	r := &textRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.texts) != 2 || r.texts[0] != "top" || r.texts[1] != "bottom" {
		t.Fatalf("annotation multiline draws = %v, want [top bottom]", r.texts)
	}
	for i, col := range r.textColors {
		if !approx(col.A, 0.5, 1e-12) {
			t.Fatalf("annotation multiline text alpha[%d] = %g, want 0.5", i, col.A)
		}
	}
}

func TestAnnotateRespectsConfiguredCoordinateSpaces(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Annotate("data", 0.25, 0.75, AnnotationOptions{
		Coords:  Coords(CoordData),
		OffsetX: 10,
		OffsetY: -15,
	})
	ax.Annotate("axes", 0.5, 0.5, AnnotationOptions{
		Coords:  Coords(CoordAxes),
		OffsetX: -12,
		OffsetY: 6,
	})
	ax.Annotate("figure", 0.2, 0.3, AnnotationOptions{
		Coords:  Coords(CoordFigure),
		OffsetX: 7,
		OffsetY: 4,
	})

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	dataToPx := ctx.TransformFor(Coords(CoordData))
	axesToPx := ctx.TransformFor(Coords(CoordAxes))
	figureToPx := ctx.TransformFor(Coords(CoordFigure))
	if dataToPx == nil || axesToPx == nil || figureToPx == nil {
		t.Fatalf("unexpected nil transform from coordinate-space transform helpers")
	}

	dataTarget := dataToPx.Apply(geom.Pt{X: 0.25, Y: 0.75})
	dataAnchor := transform.NewOffset(dataToPx, geom.Pt{X: 10, Y: -15}).Apply(geom.Pt{X: 0.25, Y: 0.75})
	axesTarget := axesToPx.Apply(geom.Pt{X: 0.5, Y: 0.5})
	axesAnchor := transform.NewOffset(axesToPx, geom.Pt{X: -12, Y: 6}).Apply(geom.Pt{X: 0.5, Y: 0.5})
	figureTarget := figureToPx.Apply(geom.Pt{X: 0.2, Y: 0.3})
	figureAnchor := transform.NewOffset(figureToPx, geom.Pt{X: 7, Y: 4}).Apply(geom.Pt{X: 0.2, Y: 0.3})

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	var connections []geom.Path
	for _, path := range r.pathCalls {
		if len(path.path.C) == 2 && path.path.C[0] == geom.MoveTo && (path.path.C[1] == geom.LineTo || path.path.C[1] == geom.QuadTo) && len(path.path.V) >= 2 {
			connections = append(connections, path.path)
		}
	}
	if len(connections) != 3 {
		t.Fatalf("expected 3 annotation connection paths, got %d", len(connections))
	}

	expectConnection := func(got geom.Path, anchor, target geom.Pt) {
		if len(got.V) < 2 {
			t.Fatalf("annotation connection path vertices = %d, want at least 2", len(got.V))
		}
		end := got.V[len(got.V)-1]
		if math.Hypot(end.X-target.X, end.Y-target.Y) > 12 {
			t.Fatalf("annotation connection end = %+v, want near target %+v", end, target)
		}
		if math.Hypot(got.V[0].X-anchor.X, got.V[0].Y-anchor.Y) > 40 {
			t.Fatalf("annotation connection start = %+v, expected near label anchor %+v", got.V[0], anchor)
		}
		if !containsPathPointNearForTextTest(r.pathCalls, target, 12) {
			t.Fatalf("annotation arrow head should land near target %+v, got paths %+v", target, r.pathCalls)
		}
	}

	expectConnection(connections[0], dataAnchor, dataTarget)
	expectConnection(connections[1], axesAnchor, axesTarget)
	expectConnection(connections[2], figureAnchor, figureTarget)
}

func TestAnnotationBboxDrawsTextFrameAndArrow(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	boxPos := geom.Pt{X: 0.7, Y: 0.25}
	align := geom.Pt{X: 0.5, Y: 0.5}
	frameOn := true
	frameFill := render.Color{R: 1, G: 1, B: 1, A: 1}
	frameEdge := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	textColor := render.Color{R: 0.9, G: 0.1, B: 0.2, A: 1}
	arrowColor := render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}

	if got := ax.AnnotationBbox("box", 0.25, 0.75, AnnotationBboxOptions{
		XYCoords:      Coords(CoordData),
		BoxCoords:     Coords(CoordAxes),
		BoxPosition:   &boxPos,
		BoxAlignment:  &align,
		FrameOn:       &frameOn,
		Padding:       4,
		FaceColor:     frameFill,
		EdgeColor:     frameEdge,
		LineWidth:     1.5,
		TextColor:     textColor,
		Arrow:         true,
		ArrowColor:    arrowColor,
		ArrowWidth:    1.25,
		ArrowHeadSize: 8,
	}); got == nil {
		t.Fatal("AnnotationBbox returned nil")
	}

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	target := ctx.TransformFor(Coords(CoordData)).Apply(geom.Pt{X: 0.25, Y: 0.75})
	boxAnchor := ctx.TransformFor(Coords(CoordAxes)).Apply(boxPos)

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	if len(r.texts) != 1 || r.texts[0] != "box" {
		t.Fatalf("unexpected annotation-box texts: %v", r.texts)
	}
	layout := measureSingleLineTextLayout(r, "box", resolvedFontSize(0, ctx), ctx.RC.FontKey, ctx.RC.UseTeX)
	wantOrigin := alignedSingleLineOrigin(boxAnchor, layout, TextAlignCenter, textLayoutVAlignCenter)
	if !approx(r.origins[0].X, wantOrigin.X, 1e-9) || !approx(r.origins[0].Y, wantOrigin.Y, 1e-9) {
		t.Fatalf("annotation-box text origin = %+v, want %+v", r.origins[0], wantOrigin)
	}
	if !hasPathPaint(r.pathPaints, frameFill, frameEdge, 1.5) {
		t.Fatalf("annotation-box frame paint not found in %+v", r.pathPaints)
	}

	foundArrowToTarget := false
	for _, call := range r.pathCalls {
		if pathHasPointNearForTextTest(call.path, target, 12) {
			foundArrowToTarget = true
			break
		}
	}
	if !foundArrowToTarget {
		t.Fatalf("annotation-box arrow should land near annotated point %+v, got paths %+v", target, r.pathCalls)
	}
}

func TestAnnotationBboxArrowStartsFromBoxRelposBeforePatchClip(t *testing.T) {
	ctx := createTestDrawContext()
	boxPos := geom.Pt{X: 0.44, Y: 0.64}
	box := &AnnotationBbox{
		Point:           geom.Pt{X: 0.2, Y: 0.8},
		Content:         "box",
		XYCoords:        Coords(CoordFigure),
		BoxCoords:       Coords(CoordFigure),
		BoxPosition:     &boxPos,
		BoxAlignment:    geom.Pt{X: 0.5, Y: 0.5},
		FrameOn:         true,
		Padding:         4,
		FontSize:        10,
		Arrow:           true,
		ArrowWidth:      1,
		ArrowHeadSize:   9,
		ArrowStyle:      ArrowStyle{Name: "-"},
		ConnectionStyle: ConnectionStyle{Name: "arc3"},
	}
	r := &textRecordingRenderer{}

	box.DrawOverlay(r, ctx)

	if len(r.pathCalls) == 0 {
		t.Fatal("expected annotation-box arrow path")
	}
	target := transformedPoint(ctx, box.XYCoords, box.Point, 0, 0)
	layout := measureSingleLineTextLayout(r, box.Content, box.FontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	boxAnchor := box.boxAnchor(ctx)
	contentBox := annotationBoxRect(boxAnchor, box.contentSize(layout, ctx), box.BoxAlignment)
	frame := expandAnchoredRect(contentBox, box.resolvedPadding(box.FontSize, ctx))
	raw := box.ConnectionStyle.connect(rectCenter(frame), target, 0, 0)
	boundary, ok := connectionPatchBoundaryPoint(raw, pixelRectPath(frame).Interpolated(8).V, true)
	if !ok {
		t.Fatalf("expected center-to-target path to leave annotation box: box=%+v path=%+v", frame, raw)
	}
	raw.V[0] = boundary
	want := shrinkPathEndpoints(raw, arrowShrinkPixels(ctx, 2), arrowShrinkPixels(ctx, 2))
	got := r.pathCalls[0].path
	if len(got.V) != len(want.V) || distance(got.V[0], want.V[0]) > 1e-9 {
		t.Fatalf("annotation-box arrow start = %+v, want clipped relpos start %+v (box=%+v target=%+v)", got.V, want.V, frame, target)
	}
}

func TestAnnotationBboxDrawsImageContent(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	img := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 10, 6)))
	boxPos := geom.Pt{X: 0.4, Y: 0.6}
	align := geom.Pt{X: 0, Y: 1}
	ax.AnnotationBbox("", 0.1, 0.2, AnnotationBboxOptions{
		BoxCoords:    Coords(CoordAxes),
		BoxPosition:  &boxPos,
		BoxAlignment: &align,
		Image:        img,
		ImageZoom:    2,
		Padding:      3,
	})

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	anchor := ctx.TransformFor(Coords(CoordAxes)).Apply(boxPos)

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	if got := len(r.imageDsts); got != 1 {
		t.Fatalf("annotation image draw count = %d, want 1", got)
	}
	scale := pointsToPixels(fig.RC, 1)
	wantDst := geom.Rect{
		Min: anchor,
		Max: geom.Pt{X: anchor.X + 20*scale, Y: anchor.Y + 12*scale},
	}
	if !approxRect(r.imageDsts[0], wantDst, 1e-9) {
		t.Fatalf("annotation image dst = %+v, want %+v", r.imageDsts[0], wantDst)
	}
}

func TestAnnotationBboxImageZoomScalesByDPI(t *testing.T) {
	fig := NewFigure(800, 600)
	fig.RC = style.Apply(fig.RC, style.WithDPI(144))
	ax := fig.AddAxes(unitRect())
	img := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 10, 6)))
	boxPos := geom.Pt{X: 0.4, Y: 0.6}
	align := geom.Pt{X: 0, Y: 1}
	ax.AnnotationBbox("", 0.1, 0.2, AnnotationBboxOptions{
		BoxCoords:    Coords(CoordAxes),
		BoxPosition:  &boxPos,
		BoxAlignment: &align,
		Image:        img,
		ImageZoom:    2,
		Padding:      0,
	})

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	anchor := ctx.TransformFor(Coords(CoordAxes)).Apply(boxPos)

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	wantDst := geom.Rect{
		Min: anchor,
		Max: geom.Pt{X: anchor.X + 40, Y: anchor.Y + 24},
	}
	if len(r.imageDsts) != 1 || !approxRect(r.imageDsts[0], wantDst, 1e-9) {
		t.Fatalf("annotation image dst = %+v, want [%+v]", r.imageDsts, wantDst)
	}
}

type textRecordingRenderer struct {
	render.NullRenderer
	pathCount  int
	pathPaints []render.Paint
	pathCalls  []recordedPathCall
	texts      []string
	textColors []render.Color
	textSizes  []float64
	origins    []geom.Pt
	imageDsts  []geom.Rect
}

func (r *textRecordingRenderer) Path(p geom.Path, paint *render.Paint) {
	r.pathCount++
	call := recordedPathCall{path: p}
	if paint != nil {
		call.paint = *paint
		r.pathPaints = append(r.pathPaints, call.paint)
	}
	r.pathCalls = append(r.pathCalls, call)
}

func (r *textRecordingRenderer) Image(_ render.Image, dst geom.Rect) {
	r.imageDsts = append(r.imageDsts, dst)
}

func (r *textRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

type mathInkBoundsRenderer struct {
	textRecordingRenderer
}

func (r *mathInkBoundsRenderer) MeasureTextBounds(text string, size float64, _ string) (render.TextBounds, bool) {
	if text == "" || size <= 0 {
		return render.TextBounds{}, false
	}
	return render.TextBounds{
		X: 0,
		Y: -size * 0.55,
		W: float64(len(text)) * size * 0.5,
		H: size * 0.70,
	}, true
}

type fontMetricTextRecordingRenderer struct {
	textRecordingRenderer
	fontHeights render.FontHeightMetrics
}

func (r *fontMetricTextRecordingRenderer) MeasureFontHeights(float64, string) (render.FontHeightMetrics, bool) {
	return r.fontHeights, true
}

func (r *textRecordingRenderer) DrawText(text string, origin geom.Pt, size float64, col render.Color) {
	r.texts = append(r.texts, text)
	r.textColors = append(r.textColors, col)
	r.textSizes = append(r.textSizes, size)
	r.origins = append(r.origins, origin)
}

func hasPathPaint(paints []render.Paint, fill, stroke render.Color, lineWidth float64) bool {
	for _, paint := range paints {
		if paint.Fill == fill && paint.Stroke == stroke && approx(paint.LineWidth, lineWidth, 1e-12) {
			return true
		}
	}
	return false
}

func hasPaintAlpha(paints []render.Paint, alpha float64) bool {
	for _, paint := range paints {
		if approx(paint.Fill.A, alpha, 1e-12) || approx(paint.Stroke.A, alpha, 1e-12) {
			return true
		}
	}
	return false
}

type recordedFontTextCall struct {
	text    string
	anchor  geom.Pt
	fontKey string
}

type fontAwareTextRecordingRenderer struct {
	textRecordingRenderer
	fontTextCalls     []recordedFontTextCall
	fontRotatedCalls  []recordedFontTextCall
	fontVerticalCalls []recordedFontTextCall
	verticalCalls     []string
}

func (r *fontAwareTextRecordingRenderer) DrawTextWithFont(text string, _ geom.Pt, _ float64, _ render.Color, fontKey string) {
	r.fontTextCalls = append(r.fontTextCalls, recordedFontTextCall{text: text, fontKey: fontKey})
}

func (r *fontAwareTextRecordingRenderer) DrawTextRotated(text string, anchor geom.Pt, _ float64, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, anchor)
}

func (r *fontAwareTextRecordingRenderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, _ float64, _ float64, _ render.Color, fontKey string) {
	r.fontRotatedCalls = append(r.fontRotatedCalls, recordedFontTextCall{text: text, anchor: anchor, fontKey: fontKey})
}

func (r *fontAwareTextRecordingRenderer) DrawTextVertical(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.verticalCalls = append(r.verticalCalls, text)
}

func (r *fontAwareTextRecordingRenderer) DrawTextVerticalWithFont(text string, _ geom.Pt, _ float64, _ render.Color, fontKey string) {
	r.fontVerticalCalls = append(r.fontVerticalCalls, recordedFontTextCall{text: text, fontKey: fontKey})
}

type verticalMathTextRecordingRenderer struct {
	textRecordingRenderer
	verticalTexts []string
	textPathCalls []string
}

func (r *verticalMathTextRecordingRenderer) DrawTextVertical(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.verticalTexts = append(r.verticalTexts, text)
}

func (r *verticalMathTextRecordingRenderer) TextPath(text string, origin geom.Pt, _ float64, _ string) (geom.Path, bool) {
	r.textPathCalls = append(r.textPathCalls, text)
	return patchRectPath(geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - 4},
		Max: geom.Pt{X: origin.X + 4, Y: origin.Y},
	}), true
}

type texRecordingRenderer struct {
	textRecordingRenderer
	texDraws        []string
	texRotatedDraws []string
}

func (r *texRecordingRenderer) MeasureTeX(text string, size float64, fontKey string) (render.TextMetrics, bool) {
	return render.TextMetrics{W: 123, H: 22, Ascent: 17, Descent: 5}, true
}

func (r *texRecordingRenderer) DrawTeX(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) bool {
	r.texDraws = append(r.texDraws, text)
	return true
}

func (r *texRecordingRenderer) DrawTeXRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) bool {
	r.texRotatedDraws = append(r.texRotatedDraws, text)
	return true
}

func (r *texRecordingRenderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, anchor)
}

func TestAxesTextSupportsAxesAndBlendedCoordinates(t *testing.T) {
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

	ax.Text(0.25, 0.75, "axes", TextOptions{
		Coords:  Coords(CoordAxes),
		OffsetX: 5,
		OffsetY: -7,
	})
	ax.Text(0.25, 0.75, "blend", TextOptions{
		Coords: BlendCoords(CoordFigure, CoordAxes),
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 2 {
		t.Fatalf("expected 2 text draws, got %d", len(r.texts))
	}

	// Display space is y-up: axes y=0.75 -> 60+0.75*480=420, plus OffsetY=-7 = 413.
	wantAxes := geom.Pt{X: 245, Y: 413}
	if r.origins[0] != wantAxes {
		t.Fatalf("axes coords origin = %+v, want %+v", r.origins[0], wantAxes)
	}

	wantBlend := geom.Pt{X: 200, Y: 420}
	if r.origins[1] != wantBlend {
		t.Fatalf("blended coords origin = %+v, want %+v", r.origins[1], wantBlend)
	}
}

func TestTextBBoxDrawsBehindAxesAndFigureText(t *testing.T) {
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

	box := &TextBBoxOptions{
		FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor: render.Color{R: 0.7, G: 0.7, B: 0.7, A: 1},
	}
	ax.Text(0.02, 0.98, "axes note", TextOptions{
		Coords: Coords(CoordAxes),
		HAlign: TextAlignLeft,
		VAlign: TextVAlignTop,
		BBox:   box,
	})
	fig.Text(0.98, 0.02, "figure note", TextOptions{
		HAlign: TextAlignRight,
		VAlign: TextVAlignBottom,
		BBox:   box,
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if !containsTextString(r.texts, "axes note") || !containsTextString(r.texts, "figure note") {
		t.Fatalf("missing bbox text draws: %v", r.texts)
	}
	if r.pathCount < 2 {
		t.Fatalf("expected text bbox paths, got %d", r.pathCount)
	}
	if len(r.pathPaints) < 2 || r.pathPaints[0].Fill.A == 0 || r.pathPaints[0].Stroke.A == 0 {
		t.Fatalf("expected visible bbox fill and stroke, got %+v", r.pathPaints)
	}
}

func TestMultilineTextSplitsDrawsAndUsesBlockBBox(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.1, 0.9, "top\nbottom", TextOptions{
		Coords: Coords(CoordAxes),
		VAlign: TextVAlignTop,
		BBox: &TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
		},
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 2 {
		t.Fatalf("expected two text draws, got %d: %v", len(r.texts), r.texts)
	}
	if r.texts[0] != "top" || r.texts[1] != "bottom" {
		t.Fatalf("unexpected multiline draw order: %v", r.texts)
	}
	if r.pathCount != 1 {
		t.Fatalf("expected one multiline bbox path, got %d", r.pathCount)
	}
	// Display space is y-up: the second line sits below the first at a smaller Y.
	if !(r.origins[1].Y < r.origins[0].Y) {
		t.Fatalf("expected second line below first, got origins %v", r.origins)
	}
}

func TestMultilineTextPathEffectsUseGlyphPaths(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "top\nbottom",
		FontSize: 10,
		ClipOn:   true,
		PathEffects: []render.PathEffect{
			render.StrokePathEffect(render.Color{R: 1, A: 1}, 2, geom.Pt{}),
			render.NormalPathEffect(),
		},
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.texts) != 0 {
		t.Fatalf("path-effect multiline text should draw glyph paths, got text draws %v", r.texts)
	}
	if len(r.pathPaints) < 2 {
		t.Fatalf("expected one effect-painted glyph path per line, got %d path paints", len(r.pathPaints))
	}
	for i, paint := range r.pathPaints {
		if len(paint.PathEffects) != 2 {
			t.Fatalf("path paint %d effects = %+v, want stroke + normal effects", i, paint.PathEffects)
		}
	}
}

func TestRotatedMultilineTextBBoxRotatesWithText(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "top\nbottom",
		FontSize: 10,
		Angle:    45,
		ClipOn:   true,
		BBox: &TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{A: 1},
			Padding:   1,
		},
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 2 {
		t.Fatalf("expected two rotated multiline text draws, got %+v", r.fontRotatedCalls)
	}
	if len(r.pathCalls) == 0 {
		t.Fatal("expected rotated multiline text bbox path")
	}
	path := r.pathCalls[0].path
	if len(path.V) < 4 {
		t.Fatalf("bbox path vertices = %+v, want rotated rectangle vertices", path.V)
	}
	if approx(path.V[0].Y, path.V[1].Y, 1e-9) || approx(path.V[1].X, path.V[2].X, 1e-9) {
		t.Fatalf("rotated multiline text bbox remained axis-aligned: %+v", path.V[:4])
	}
}

func TestTextMultiAlignmentControlsLineAlignmentWithinBlock(t *testing.T) {
	ctx := createTestDrawContext()
	multiAlign := TextAlignLeft
	text := &Text{
		Position:       geom.Pt{X: 1, Y: 1},
		Content:        "narrow\nmuch wider",
		FontSize:       10,
		HAlign:         TextAlignRight,
		VAlign:         TextVAlignTop,
		MultiAlignment: &multiAlign,
		ClipOn:         true,
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.origins) != 2 {
		t.Fatalf("expected two multiline origins, got %d: %+v", len(r.origins), r.origins)
	}
	if !approx(r.origins[0].X, r.origins[1].X, 1e-12) {
		t.Fatalf("left multialignment should keep line origins equal inside right-aligned block, got %+v", r.origins)
	}
	blockRight := transformedPoint(ctx, text.Coords, text.Position, text.OffsetX, text.OffsetY).X
	if !approx(r.origins[1].X+float64(len("much wider"))*text.FontSize*0.5, blockRight, 1e-12) {
		t.Fatalf("right-aligned multiline block no longer ends at anchor: origins=%+v anchorX=%v", r.origins, blockRight)
	}
}

func TestMultilineTextLinespacingControlsBaselineAdvance(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:    geom.Pt{X: 1, Y: 1},
		Content:     "first\nsecond",
		FontSize:    10,
		Linespacing: 2,
		ClipOn:      true,
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.origins) != 2 {
		t.Fatalf("multiline draw origins = %d, want 2", len(r.origins))
	}
	wantAdvance := text.FontSize * text.Linespacing
	gotAdvance := r.origins[0].Y - r.origins[1].Y // y-up: next line is below at smaller Y
	if !approx(gotAdvance, wantAdvance, 1e-9) {
		t.Fatalf("multiline baseline advance = %v, want %v", gotAdvance, wantAdvance)
	}
}

func TestMultilineTextNormalLinespacingUsesFontLineGap(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "first\nsecond",
		FontSize: 10,
		ClipOn:   true,
	}
	r := &fontMetricTextRecordingRenderer{
		fontHeights: render.FontHeightMetrics{Ascent: 12, Descent: 4, LineGap: 3},
	}

	text.Draw(r, ctx)

	if len(r.origins) != 2 {
		t.Fatalf("multiline draw origins = %d, want 2", len(r.origins))
	}
	wantAdvance := r.fontHeights.Ascent + r.fontHeights.Descent + r.fontHeights.LineGap
	gotAdvance := r.origins[0].Y - r.origins[1].Y // y-up: next line is below at smaller Y
	if !approx(gotAdvance, wantAdvance, 1e-9) {
		t.Fatalf("normal multiline baseline advance = %v, want font height + gap %v", gotAdvance, wantAdvance)
	}
}

func TestMultilineTextNumericLinespacingUsesFontHeight(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:    geom.Pt{X: 1, Y: 1},
		Content:     "first\nsecond",
		FontSize:    10,
		Linespacing: 1.5,
		ClipOn:      true,
	}
	r := &fontMetricTextRecordingRenderer{
		fontHeights: render.FontHeightMetrics{Ascent: 12, Descent: 4, LineGap: 3},
	}

	text.Draw(r, ctx)

	if len(r.origins) != 2 {
		t.Fatalf("multiline draw origins = %d, want 2", len(r.origins))
	}
	wantAdvance := text.Linespacing * (r.fontHeights.Ascent + r.fontHeights.Descent)
	gotAdvance := r.origins[0].Y - r.origins[1].Y // y-up: next line is below at smaller Y
	if !approx(gotAdvance, wantAdvance, 1e-9) {
		t.Fatalf("numeric multiline baseline advance = %v, want linespacing * font height %v", gotAdvance, wantAdvance)
	}
}

func TestMultilineTextAngleUsesRotatedTextDrawer(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "top\nbottom",
		FontSize: 10,
		Angle:    30,
		ClipOn:   true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 2 {
		t.Fatalf("expected two rotated multiline text draws, got %+v", r.fontRotatedCalls)
	}
	if len(r.fontTextCalls) != 0 || len(r.texts) != 0 {
		t.Fatalf("multiline angle should not use unrotated text draws, font=%+v legacy=%+v", r.fontTextCalls, r.texts)
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

func containsTextString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsFontTextCall(calls []recordedFontTextCall, want string) bool {
	for _, call := range calls {
		if call.text == want {
			return true
		}
	}
	return false
}

func containsPathPointForTextTest(calls []recordedPathCall, want geom.Pt) bool {
	for _, call := range calls {
		if pathHasPointForTextTest(call.path, want) {
			return true
		}
	}
	return false
}

func containsPathPointNearForTextTest(calls []recordedPathCall, want geom.Pt, tol float64) bool {
	for _, call := range calls {
		if pathHasPointNearForTextTest(call.path, want, tol) {
			return true
		}
	}
	return false
}

func pathHasPointForTextTest(path geom.Path, want geom.Pt) bool {
	for _, pt := range path.V {
		if approx(pt.X, want.X, 1e-9) && approx(pt.Y, want.Y, 1e-9) {
			return true
		}
	}
	return false
}

func pathHasPointNearForTextTest(path geom.Path, want geom.Pt, tol float64) bool {
	for _, pt := range path.V {
		if math.Hypot(pt.X-want.X, pt.Y-want.Y) <= tol {
			return true
		}
	}
	return false
}

func approxRect(got, want geom.Rect, tol float64) bool {
	return approx(got.Min.X, want.Min.X, tol) &&
		approx(got.Min.Y, want.Min.Y, tol) &&
		approx(got.Max.X, want.Max.X, tol) &&
		approx(got.Max.Y, want.Max.Y, tol)
}

func containsMathRunText(runs []MathTextLayoutRun, text string) bool {
	for _, run := range runs {
		if run.Text == text {
			return true
		}
	}
	return false
}

func containsMathRun(runs []MathTextLayoutRun, text string, size float64) bool {
	for _, run := range runs {
		if run.Text == text && almostEqualFloat(run.FontSize, size) {
			return true
		}
	}
	return false
}

func almostEqualFloat(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 1e-9
}

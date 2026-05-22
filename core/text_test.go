package core

import (
	"math"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
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
	if !containsMathRun(layout.Runs, "√", 10.44) || !containsMathRun(layout.Runs, "3", 9.9) {
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
	if boundedGap >= plainGap-4 {
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

func TestAlignedTextOrigin(t *testing.T) {
	anchor := geom.Pt{X: 100, Y: 50}
	metrics := render.TextMetrics{W: 40, Ascent: 8, Descent: 2}

	got := alignedTextOrigin(anchor, metrics, TextAlignCenter, TextVAlignTop)
	if got.X != 80 || got.Y != 58 {
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
		if !approx(end.X, target.X, 1e-9) || !approx(end.Y, target.Y, 1e-9) {
			t.Fatalf("annotation connection end = %+v, want target %+v", end, target)
		}
		if math.Hypot(got.V[0].X-anchor.X, got.V[0].Y-anchor.Y) > 40 {
			t.Fatalf("annotation connection start = %+v, expected near label anchor %+v", got.V[0], anchor)
		}
	}

	expectConnection(connections[0], dataAnchor, dataTarget)
	expectConnection(connections[1], axesAnchor, axesTarget)
	expectConnection(connections[2], figureAnchor, figureTarget)
}

type textRecordingRenderer struct {
	render.NullRenderer
	pathCount  int
	pathPaints []render.Paint
	pathCalls  []recordedPathCall
	texts      []string
	textColors []render.Color
	origins    []geom.Pt
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

func (r *textRecordingRenderer) DrawText(text string, origin geom.Pt, _ float64, col render.Color) {
	r.texts = append(r.texts, text)
	r.textColors = append(r.textColors, col)
	r.origins = append(r.origins, origin)
}

type recordedFontTextCall struct {
	text    string
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

func (r *fontAwareTextRecordingRenderer) DrawTextRotatedWithFont(text string, _ geom.Pt, _ float64, _ float64, _ render.Color, fontKey string) {
	r.fontRotatedCalls = append(r.fontRotatedCalls, recordedFontTextCall{text: text, fontKey: fontKey})
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

	wantAxes := geom.Pt{X: 245, Y: 173}
	if r.origins[0] != wantAxes {
		t.Fatalf("axes coords origin = %+v, want %+v", r.origins[0], wantAxes)
	}

	wantBlend := geom.Pt{X: 200, Y: 180}
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
	if !(r.origins[1].Y > r.origins[0].Y) {
		t.Fatalf("expected second line below first, got origins %v", r.origins)
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

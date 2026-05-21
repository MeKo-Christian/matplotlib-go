package mathtext

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type testMeasurer struct{}

func (testMeasurer) MeasureText(text string, size float64, _ string) Metrics {
	return Metrics{
		W:       float64(len([]rune(text))) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

type shapingMeasurer struct{}

func (shapingMeasurer) MeasureText(text string, size float64, fontKey string) Metrics {
	shaped, ok := render.ShapeText(text, geom.Pt{}, size*100/72, render.TextShapingOptions{FontKey: fontKey})
	if !ok {
		return Metrics{}
	}
	ascent := 0.0
	descent := 0.0
	if shaped.Bounds.H > 0 {
		ascent = math.Max(0, -shaped.Bounds.Y)
		descent = math.Max(0, shaped.Bounds.Y+shaped.Bounds.H)
	}
	return Metrics{
		W:       shaped.Advance.X,
		H:       ascent + descent,
		Ascent:  ascent,
		Descent: descent,
		BoundsY: shaped.Bounds.Y,
		BoundsH: shaped.Bounds.H,
	}
}

type countingMeasurer struct {
	calls int
	scale float64
}

func (m *countingMeasurer) MeasureText(text string, size float64, _ string) Metrics {
	m.calls++
	return Metrics{
		W:       float64(len([]rune(text))) * size * m.scale,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

type recordingResolver struct {
	requests []FontRequest
}

func (r *recordingResolver) ResolveMathFontKey(_ string, request FontRequest) string {
	r.requests = append(r.requests, request)
	if len(request.Families) > 0 {
		return "resolved:" + strings.Join(request.Families, ",")
	}
	if request.Style != "" {
		return "style:" + string(request.Style)
	}
	return ""
}

type shapingFontResolver struct{}

func (shapingFontResolver) ResolveMathFontKey(base string, request FontRequest) string {
	props := render.ParseFontProperties(base)
	if len(request.Families) > 0 {
		props.File = ""
		props.Families = append([]string(nil), request.Families...)
	}
	if request.Style != "" {
		props.Style = render.FontStyle(request.Style)
	}
	if request.Weight > 0 {
		props.Weight = request.Weight
	}
	if face, ok := render.DefaultFontManager().FindFont(props); ok && face.Path != "" {
		return face.Path
	}
	if len(props.Families) > 0 {
		return strings.Join(props.Families, ", ")
	}
	return base
}

func TestNormalizeDisplayParsesInlineMath(t *testing.T) {
	got := NormalizeDisplay(`signal $\\alpha_i^2$ peak`)
	if got != "signal αᵢ² peak" {
		t.Fatalf("NormalizeDisplay = %q, want %q", got, "signal αᵢ² peak")
	}
}

func TestSplitDisplaySegmentsRejectsUnbalancedMath(t *testing.T) {
	if _, _, ok := SplitDisplaySegments(`cost is $5`); ok {
		t.Fatal("SplitDisplaySegments returned ok for unbalanced math")
	}
}

func TestLayoutDisplayBuildsMixedRuns(t *testing.T) {
	layout, ok := LayoutDisplay(testMeasurer{}, `phase $\\frac{1}{2}$ peak`, 20, "base", Options{})
	if !ok {
		t.Fatal("LayoutDisplay returned !ok")
	}
	if layout.Width <= 0 || len(layout.Runs) < 3 || len(layout.Rules) == 0 {
		t.Fatalf("unexpected layout: %+v", layout)
	}
}

func TestLayoutMathTextDelegatesStyleFontResolution(t *testing.T) {
	resolver := &recordingResolver{}
	layout, ok := LayoutMathText(testMeasurer{}, `\mathsf{s}`, 20, "base", Options{FontResolver: resolver})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if len(resolver.requests) != 1 || len(resolver.requests[0].Families) == 0 {
		t.Fatalf("font resolver was not called with family override: %+v", resolver.requests)
	}
	if len(layout.Runs) != 1 || !strings.HasPrefix(layout.Runs[0].FontKey, "resolved:") {
		t.Fatalf("style font key was not applied to layout run: %+v", layout.Runs)
	}
}

func TestLayoutMathTextUsesItalicLatinVariablesByDefault(t *testing.T) {
	resolver := &recordingResolver{}
	layout, ok := LayoutMathText(testMeasurer{}, `x+\mathrm{x}+\sin x`, 20, "base", Options{FontResolver: resolver})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	italicRuns := 0
	romanRuns := 0
	for _, run := range layout.Runs {
		if run.Text != "x" {
			continue
		}
		switch {
		case strings.HasPrefix(run.FontKey, "resolved:"):
			romanRuns++
		case run.FontKey == "style:italic":
			italicRuns++
		default:
			t.Fatalf("unexpected font key for x run %q in %+v", run.FontKey, layout.Runs)
		}
	}
	if italicRuns != 2 || romanRuns != 1 {
		t.Fatalf("unexpected variable styles: italic=%d roman=%d runs=%+v", italicRuns, romanRuns, layout.Runs)
	}

	var sawItalicRequest bool
	for _, request := range resolver.requests {
		if request.Style == FontStyleItalic {
			sawItalicRequest = true
		}
	}
	if !sawItalicRequest {
		t.Fatalf("font resolver was not asked for an italic variable face: %+v", resolver.requests)
	}
}

func TestLayoutMathTextUsesItalicLatinVariablesInMatrices(t *testing.T) {
	resolver := &recordingResolver{}
	layout, ok := LayoutMathText(testMeasurer{}, `\begin{pmatrix}x&y\end{pmatrix}`, 20, "base", Options{FontResolver: resolver})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}

	italicRuns := 0
	for _, run := range layout.Runs {
		if (run.Text == "x" || run.Text == "y") && run.FontKey == "style:italic" {
			italicRuns++
		}
	}
	if italicRuns != 2 {
		t.Fatalf("expected matrix variables to use implicit italic style, got runs: %+v", layout.Runs)
	}
}

func TestLayoutMathTextUsesRuleDelimitersForStretchyBars(t *testing.T) {
	layout, ok := LayoutMathText(testMeasurer{}, `\left| \frac{1}{2} \right|`, 20, "base", Options{})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if len(layout.Rules) < 3 {
		t.Fatalf("expected fraction rule plus two stretchy bar rules, got %d rules: %+v", len(layout.Rules), layout.Rules)
	}
	barRules := 0
	for _, rule := range layout.Rules {
		if rule.Rect.H() > 20 && rule.Rect.W() < 5 {
			barRules++
		}
	}
	if barRules < 2 {
		t.Fatalf("expected at least two tall bar delimiter rules, got %d in %+v", barRules, layout.Rules)
	}
}

func TestLayoutMathTextUsesSizedGlyphsForStretchyBrackets(t *testing.T) {
	layout, ok := LayoutMathText(testMeasurer{}, `\left[\frac{1}{2}\right]`, 20, "base", Options{})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if len(layout.Rules) != 1 {
		t.Fatalf("expected only the fraction rule, got %d rules: %+v", len(layout.Rules), layout.Rules)
	}
	if len(layout.Runs) != 4 {
		t.Fatalf("expected bracket glyphs plus numerator and denominator text runs, got %+v", layout.Runs)
	}
	if !isTestSTIXSizeFont(layout.Runs[0].FontKey) || !isTestSTIXSizeFont(layout.Runs[len(layout.Runs)-1].FontKey) {
		t.Fatalf("expected stretchy bracket glyphs from STIX size fonts, got %+v", layout.Runs)
	}
}

func TestLayoutMathTextSupportsRulelessDelimitedFractions(t *testing.T) {
	layout, ok := LayoutMathText(testMeasurer{}, `\binom{n}{k}`, 20, "base", Options{})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if len(layout.Rules) != 0 {
		t.Fatalf("binom should not draw a fraction rule, got %+v", layout.Rules)
	}
	if len(layout.Runs) != 4 {
		t.Fatalf("expected left delimiter, numerator, denominator, right delimiter runs; got %+v", layout.Runs)
	}
	if !isTestSTIXSizeFont(layout.Runs[0].FontKey) || !isTestSTIXSizeFont(layout.Runs[len(layout.Runs)-1].FontKey) {
		t.Fatalf("binom did not add sized parenthesized delimiters: %+v", layout.Runs)
	}

	var numY, denY float64
	for _, run := range layout.Runs {
		switch run.Text {
		case "n":
			numY = run.Offset.Y
		case "k":
			denY = run.Offset.Y
		}
	}
	if numY >= 0 || denY <= 0 {
		t.Fatalf("expected numerator above and denominator below baseline: numY=%v denY=%v runs=%+v", numY, denY, layout.Runs)
	}
}

func TestLayoutMathTextSupportsGenfracDelimitersAndRuleSize(t *testing.T) {
	layout, ok := LayoutMathText(testMeasurer{}, `\genfrac{[}{]}{0}{0}{n}{k}`, 20, "base", Options{})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	for _, rule := range layout.Rules {
		if rule.Rect.Min.Y < 1 && rule.Rect.Max.Y > -1 && rule.Rect.W() > 10 {
			t.Fatalf("zero-rule genfrac should not draw a central fraction rule, got %+v", layout.Rules)
		}
	}
	if len(layout.Rules) != 0 {
		t.Fatalf("zero-rule genfrac should only draw sized bracket glyphs, got rules: %+v", layout.Rules)
	}
	if len(layout.Runs) != 4 || !isTestSTIXSizeFont(layout.Runs[0].FontKey) || !isTestSTIXSizeFont(layout.Runs[len(layout.Runs)-1].FontKey) {
		t.Fatalf("genfrac did not apply requested bracket delimiters as STIX size glyphs: %+v", layout.Runs)
	}
	if !containsTestRun(layout.Runs, "n", 20) || !containsTestRun(layout.Runs, "k", 20) {
		t.Fatalf("display-style genfrac should keep numerator and denominator at base size: %+v", layout.Runs)
	}
}

func isTestSTIXSizeFont(fontKey string) bool {
	return strings.HasPrefix(fontKey, "STIXSize")
}

func TestLayoutMathTextSupportsDisplayStyleFractions(t *testing.T) {
	frac, ok := LayoutMathText(testMeasurer{}, `\frac{n}{k}`, 20, "base", Options{})
	if !ok {
		t.Fatal("frac LayoutMathText returned !ok")
	}
	dfrac, ok := LayoutMathText(testMeasurer{}, `\dfrac{n}{k}`, 20, "base", Options{})
	if !ok {
		t.Fatal("dfrac LayoutMathText returned !ok")
	}
	if dfrac.Height <= frac.Height {
		t.Fatalf("dfrac should use a display-style vertical layout: frac=%+v dfrac=%+v", frac, dfrac)
	}
	if !containsTestRun(dfrac.Runs, "n", 20) || !containsTestRun(dfrac.Runs, "k", 20) {
		t.Fatalf("dfrac should keep numerator and denominator at display size: %+v", dfrac.Runs)
	}
}

func TestLayoutMathTextMatchesMatplotlibFixtureMetrics(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		size        float64
		wantWidth   float64
		wantAscent  float64
		wantDescent float64
	}{
		{
			name:        "binom",
			expr:        `\binom{n}{k} = \frac{n!}{k!(n-k)!}`,
			size:        23,
			wantWidth:   202,
			wantAscent:  31,
			wantDescent: 16,
		},
		{
			name:        "genfrac matrix",
			expr:        `\genfrac{(}{)}{0}{0}{a\quad b}{c\quad d}`,
			size:        25,
			wantWidth:   122,
			wantAscent:  40,
			wantDescent: 23,
		},
		{
			name:        "sum limits",
			expr:        `\sum_{i=1}^{n} i^2`,
			size:        26,
			wantWidth:   84,
			wantAscent:  58,
			wantDescent: 36,
		},
		{
			name:        "integral side scripts",
			expr:        `\int_0^\infty e^{-x}\,dx = 1`,
			size:        24,
			wantWidth:   214,
			wantAscent:  40,
			wantDescent: 18,
		},
	}

	for _, tt := range tests {
		layout, ok := LayoutMathText(shapingMeasurer{}, tt.expr, tt.size, "DejaVu Sans", Options{})
		if !ok {
			t.Fatalf("%s: LayoutMathText returned !ok", tt.name)
		}
		if math.Abs(layout.Width-tt.wantWidth) > 8 ||
			math.Abs(layout.Ascent-tt.wantAscent) > 4 ||
			math.Abs(layout.Descent-tt.wantDescent) > 4 {
			t.Errorf("%s metrics = width %.2f ascent %.2f descent %.2f, want near %.2f %.2f %.2f",
				tt.name, layout.Width, layout.Ascent, layout.Descent, tt.wantWidth, tt.wantAscent, tt.wantDescent)
		}
	}
}

func TestLayoutMathTextSupportsRicherSpacingCommands(t *testing.T) {
	compact, ok := LayoutMathText(testMeasurer{}, `ab`, 20, "base", Options{})
	if !ok {
		t.Fatal("compact LayoutMathText returned !ok")
	}
	wide, ok := LayoutMathText(testMeasurer{}, `a\enspace b\hspace{0.5em}c`, 20, "base", Options{})
	if !ok {
		t.Fatal("wide LayoutMathText returned !ok")
	}
	if wide.Width <= compact.Width+18 {
		t.Fatalf("spacing commands did not widen expression enough: compact=%v wide=%v", compact.Width, wide.Width)
	}
	tight, ok := LayoutMathText(testMeasurer{}, `a\negthinspace b`, 20, "base", Options{})
	if !ok {
		t.Fatal("tight LayoutMathText returned !ok")
	}
	plain, ok := LayoutMathText(testMeasurer{}, `a b`, 20, "base", Options{})
	if !ok {
		t.Fatal("plain LayoutMathText returned !ok")
	}
	if tight.Width >= plain.Width {
		t.Fatalf("negative spacing did not tighten expression: tight=%v plain=%v", tight.Width, plain.Width)
	}
}

func TestLayoutMathTextUsesMatplotlibQuadBasedSpacing(t *testing.T) {
	tests := []struct {
		expr      string
		wantWidth float64
	}{
		{expr: `1+x`, wantWidth: 82},
		{expr: `1 + x`, wantWidth: 82},
		{expr: `a\quad b`, wantWidth: 75},
		{expr: `a\,b`, wantWidth: 47},
	}

	for _, tt := range tests {
		layout, ok := LayoutMathText(shapingMeasurer{}, tt.expr, 24, "DejaVu Sans", Options{})
		if !ok {
			t.Fatalf("%s: LayoutMathText returned !ok", tt.expr)
		}
		if math.Abs(layout.Width-tt.wantWidth) > 4 {
			t.Errorf("%s width = %.2f, want near %.2f", tt.expr, layout.Width, tt.wantWidth)
		}
	}
}

func TestLayoutMathTextUsesMatplotlibFractionRuleThickness(t *testing.T) {
	layout, ok := LayoutMathText(shapingMeasurer{}, `\frac{1}{2}`, 24, "DejaVu Sans", Options{})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if len(layout.Rules) != 1 {
		t.Fatalf("expected one fraction rule, got %+v", layout.Rules)
	}
	if got, want := layout.Rules[0].Rect.H(), 2.08; math.Abs(got-want) > 0.35 {
		t.Fatalf("fraction rule thickness = %.2f, want near %.2f", got, want)
	}
	if got, want := layout.Rules[0].Rect.W(), 14.85; math.Abs(got-want) > 1.0 {
		t.Fatalf("fraction rule width = %.2f, want near %.2f", got, want)
	}
	if got, want := layout.Width, 20.0; math.Abs(got-want) > 1.5 {
		t.Fatalf("fraction layout width = %.2f, want near %.2f", got, want)
	}
}

func TestLayoutMathTextDoesNotPadRulelessGenfracHorizontally(t *testing.T) {
	layout, ok := LayoutMathText(shapingMeasurer{}, `\genfrac{(}{)}{0}{0}{a\quad b}{c\quad d}`, 25, "DejaVu Sans", Options{})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	var leftParen, firstBody *MathTextLayoutRun
	for i := range layout.Runs {
		switch layout.Runs[i].Text {
		case "(":
			leftParen = &layout.Runs[i]
		case "a":
			if firstBody == nil {
				firstBody = &layout.Runs[i]
			}
		}
	}
	if leftParen == nil || firstBody == nil {
		t.Fatalf("missing delimiter/body runs: %+v", layout.Runs)
	}
	leftWidth := shapingMeasurer{}.MeasureText(leftParen.Text, leftParen.FontSize, leftParen.FontKey).W
	if got := firstBody.Offset.X - (leftParen.Offset.X + leftWidth); math.Abs(got) > 1.0 {
		t.Fatalf("ruleless genfrac inserted horizontal padding %.2f; runs=%+v", got, layout.Runs)
	}
}

func TestLayoutMathTextUsesMatplotlibPunctuationSpacing(t *testing.T) {
	layout, ok := LayoutMathText(shapingMeasurer{}, `k!(n-k)!`, 16.1, "DejaVu Sans", Options{FontResolver: shapingFontResolver{}})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if got, want := layout.Width, 112.0; math.Abs(got-want) > 2.0 {
		t.Fatalf("punctuation-spaced width = %.2f, want near %.2f; runs=%+v", got, want, layout.Runs)
	}
}

func TestLayoutMathTextUsesMatplotlibFractionAxisAlignment(t *testing.T) {
	layout, ok := LayoutMathText(shapingMeasurer{}, `\left[\frac{1}{1+x}\right]`, 24, "DejaVu Sans", Options{FontResolver: shapingFontResolver{}})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if got, want := layout.Width, 90.0; math.Abs(got-want) > 1.5 {
		t.Fatalf("bracketed fraction width = %.2f, want near %.2f", got, want)
	}
	if got, want := layout.Ascent, 32.0; math.Abs(got-want) > 1.0 {
		t.Fatalf("bracketed fraction ascent = %.2f, want near %.2f", got, want)
	}
	if got, want := layout.Descent, 13.0; math.Abs(got-want) > 1.0 {
		t.Fatalf("bracketed fraction descent = %.2f, want near %.2f", got, want)
	}
}

func TestLayoutMathTextUsesMatplotlibOverUnderGap(t *testing.T) {
	layout, ok := LayoutMathText(shapingMeasurer{}, `\sum_{i=1}^{n} i^2`, 26, "DejaVu Sans", Options{FontResolver: shapingFontResolver{}})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if got, want := layout.Ascent, 58.0; math.Abs(got-want) > 1.5 {
		t.Fatalf("sum ascent = %.2f, want near %.2f", got, want)
	}
	if got, want := layout.Descent, 36.0; math.Abs(got-want) > 1.5 {
		t.Fatalf("sum descent = %.2f, want near %.2f", got, want)
	}
}

func TestLayoutMathTextUsesMatplotlibSqrtGeometry(t *testing.T) {
	layout, ok := LayoutMathText(shapingMeasurer{}, `\sqrt{y}`, 23, "DejaVu Sans", Options{})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if got, want := layout.Width, 47.0; math.Abs(got-want) > 2.5 {
		t.Fatalf("sqrt width = %.2f, want near %.2f", got, want)
	}
	if got, want := layout.Ascent, 30.0; math.Abs(got-want) > 3.0 {
		t.Fatalf("sqrt ascent = %.2f, want near %.2f", got, want)
	}
	if got, want := layout.Descent, 7.0; math.Abs(got-want) > 2.0 {
		t.Fatalf("sqrt descent = %.2f, want near %.2f", got, want)
	}
	if len(layout.Rules) != 1 {
		t.Fatalf("expected one sqrt rule, got %+v", layout.Rules)
	}
	if got, want := layout.Rules[0].Rect.H(), 2.0; math.Abs(got-want) > 0.35 {
		t.Fatalf("sqrt rule thickness = %.2f, want near %.2f", got, want)
	}
	if len(layout.Runs) < 2 || !strings.HasPrefix(layout.Runs[0].FontKey, "STIXSize") {
		t.Fatalf("sqrt radical should use STIX size font, got runs %+v", layout.Runs)
	}
}

func TestLayoutMathTextUsesMatplotlibDisplayFontForLargeOperators(t *testing.T) {
	tests := []struct {
		expr string
		text string
		size float64
	}{
		{expr: `\sum_{i=1}^{n} i^2`, text: "∑", size: 26},
		{expr: `\prod_{k=1}^{m} k`, text: "∏", size: 26},
		{expr: `\int_0^\infty e^{-x}\,dx = 1`, text: "∫", size: 24},
	}

	for _, tt := range tests {
		layout, ok := LayoutMathText(shapingMeasurer{}, tt.expr, tt.size, "DejaVu Sans", Options{})
		if !ok {
			t.Fatalf("%s: LayoutMathText returned !ok", tt.expr)
		}
		var got *MathTextLayoutRun
		for i := range layout.Runs {
			if layout.Runs[i].Text == tt.text {
				got = &layout.Runs[i]
				break
			}
		}
		if got == nil {
			t.Fatalf("%s: missing operator %q in runs %+v", tt.expr, tt.text, layout.Runs)
		}
		if got.FontKey != "DejaVu Sans Display" || math.Abs(got.FontSize-tt.size) > 0.01 {
			t.Fatalf("%s: operator run = %+v, want DejaVu Sans Display at %.2f", tt.expr, *got, tt.size)
		}
	}
}

func TestLayoutMathTextAddsMathOperatorSpacing(t *testing.T) {
	compact, ok := LayoutMathText(testMeasurer{}, `1+x`, 20, "base", Options{})
	if !ok {
		t.Fatal("compact LayoutMathText returned !ok")
	}
	spaced, ok := LayoutMathText(testMeasurer{}, `1 + x`, 20, "base", Options{})
	if !ok {
		t.Fatal("spaced LayoutMathText returned !ok")
	}
	if compact.Width != spaced.Width {
		t.Fatalf("raw spaces should not change math-mode operator spacing: compact=%v spaced=%v", compact.Width, spaced.Width)
	}
	if compact.Width <= 30 {
		t.Fatalf("binary operator spacing did not widen expression enough: %+v", compact)
	}

	unary, ok := LayoutMathText(testMeasurer{}, `-x`, 20, "base", Options{})
	if !ok {
		t.Fatal("unary LayoutMathText returned !ok")
	}
	if unary.Width >= compact.Width {
		t.Fatalf("unary minus should not use binary spacing: unary=%v compact=%v", unary.Width, compact.Width)
	}
}

func TestNormalizeDisplayHandlesExplicitSpacingCommands(t *testing.T) {
	got := NormalizeDisplay(`$a\\hspace{0.5em}b\\negthinspace c$`)
	if got != "a b c" {
		t.Fatalf("NormalizeDisplay = %q, want %q", got, "a b c")
	}
}

func TestLayoutMathTextCacheReusesMeasuredLayout(t *testing.T) {
	cache := NewCache()
	measurer := &countingMeasurer{scale: 0.5}
	opts := Options{Cache: cache, MeasurementKey: "renderer-a"}

	first, ok := LayoutMathText(measurer, `\frac{1}{2}`, 20, "base", opts)
	if !ok {
		t.Fatal("first LayoutMathText returned !ok")
	}
	firstCalls := measurer.calls
	if firstCalls == 0 {
		t.Fatal("first layout did not measure text")
	}

	first.Runs[0].Text = "mutated"
	second, ok := LayoutMathText(measurer, `\frac{1}{2}`, 20, "base", opts)
	if !ok {
		t.Fatal("second LayoutMathText returned !ok")
	}
	if measurer.calls != firstCalls {
		t.Fatalf("cached layout remeasured text: first calls=%d second calls=%d", firstCalls, measurer.calls)
	}
	if second.Runs[0].Text == "mutated" {
		t.Fatalf("cached layout returned mutable run slice: %+v", second.Runs)
	}
	parsed, layouts := cache.Stats()
	if parsed != 1 || layouts != 1 {
		t.Fatalf("cache stats = parsed %d layouts %d, want 1/1", parsed, layouts)
	}
}

func TestLayoutMathTextCacheSeparatesMeasurementKeys(t *testing.T) {
	cache := NewCache()
	narrow := &countingMeasurer{scale: 0.4}
	wide := &countingMeasurer{scale: 0.8}

	narrowLayout, ok := LayoutMathText(narrow, `ab`, 20, "base", Options{Cache: cache, MeasurementKey: "narrow"})
	if !ok {
		t.Fatal("narrow LayoutMathText returned !ok")
	}
	wideLayout, ok := LayoutMathText(wide, `ab`, 20, "base", Options{Cache: cache, MeasurementKey: "wide"})
	if !ok {
		t.Fatal("wide LayoutMathText returned !ok")
	}
	if wideLayout.Width <= narrowLayout.Width {
		t.Fatalf("measurement keys reused incompatible layout: narrow=%v wide=%v", narrowLayout.Width, wideLayout.Width)
	}
	parsed, layouts := cache.Stats()
	if parsed != 1 || layouts != 2 {
		t.Fatalf("cache stats = parsed %d layouts %d, want 1/2", parsed, layouts)
	}
}

func TestLayoutMathTextCacheWithoutMeasurementKeyOnlyCachesParse(t *testing.T) {
	cache := NewCache()
	measurer := &countingMeasurer{scale: 0.5}
	opts := Options{Cache: cache}

	if _, ok := LayoutMathText(measurer, `ab`, 20, "base", opts); !ok {
		t.Fatal("first LayoutMathText returned !ok")
	}
	firstCalls := measurer.calls
	if _, ok := LayoutMathText(measurer, `ab`, 20, "base", opts); !ok {
		t.Fatal("second LayoutMathText returned !ok")
	}
	if measurer.calls <= firstCalls {
		t.Fatalf("layout cache was used without measurement key: first=%d second=%d", firstCalls, measurer.calls)
	}
	parsed, layouts := cache.Stats()
	if parsed != 1 || layouts != 0 {
		t.Fatalf("cache stats = parsed %d layouts %d, want 1/0", parsed, layouts)
	}
}

func TestLayoutMathTextCacheEvictsOldestEntriesWhenBounded(t *testing.T) {
	cache := NewCacheWithConfig(CacheConfig{MaxParsed: 1, MaxLayouts: 1})
	measurer := &countingMeasurer{scale: 0.5}

	if _, ok := LayoutMathText(measurer, `ab`, 20, "base", Options{Cache: cache, MeasurementKey: "renderer"}); !ok {
		t.Fatal("first LayoutMathText returned !ok")
	}
	if _, ok := LayoutMathText(measurer, `cd`, 20, "base", Options{Cache: cache, MeasurementKey: "renderer"}); !ok {
		t.Fatal("second LayoutMathText returned !ok")
	}
	parsed, layouts := cache.Stats()
	if parsed != 1 || layouts != 1 {
		t.Fatalf("bounded cache stats = parsed %d layouts %d, want 1/1", parsed, layouts)
	}

	calls := measurer.calls
	if _, ok := LayoutMathText(measurer, `ab`, 20, "base", Options{Cache: cache, MeasurementKey: "renderer"}); !ok {
		t.Fatal("third LayoutMathText returned !ok")
	}
	if measurer.calls <= calls {
		t.Fatalf("oldest layout entry was reused after eviction: before=%d after=%d", calls, measurer.calls)
	}
}

func TestCacheSaveLoadFileReusesLayoutAcrossProcesses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mathtext-cache.json")
	cache := NewCache()
	measurer := &countingMeasurer{scale: 0.5}

	first, ok := LayoutMathText(measurer, `\frac{1}{2}`, 20, "base", Options{
		Cache:          cache,
		MeasurementKey: "renderer",
	})
	if !ok {
		t.Fatal("first LayoutMathText returned !ok")
	}
	if measurer.calls == 0 {
		t.Fatal("first layout did not measure text")
	}
	if err := cache.SaveFile(path); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}
	firstBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if err := cache.SaveFile(path); err != nil {
		t.Fatalf("second SaveFile: %v", err)
	}
	secondBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("second ReadFile: %v", err)
	}
	if string(firstBytes) != string(secondBytes) {
		t.Fatal("SaveFile output is not deterministic")
	}

	loaded := NewCache()
	if err := loaded.LoadFile(path); err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	measurer.calls = 0
	second, ok := LayoutMathText(measurer, `\frac{1}{2}`, 20, "base", Options{
		Cache:          loaded,
		MeasurementKey: "renderer",
	})
	if !ok {
		t.Fatal("second LayoutMathText returned !ok")
	}
	if measurer.calls != 0 {
		t.Fatalf("loaded layout cache remeasured text: calls=%d", measurer.calls)
	}
	if second.Width != first.Width || len(second.Rules) != len(first.Rules) || len(second.Runs) != len(first.Runs) {
		t.Fatalf("loaded layout mismatch: first=%+v second=%+v", first, second)
	}
}

func containsTestRun(runs []MathTextLayoutRun, text string, size float64) bool {
	for _, run := range runs {
		if run.Text == text && run.FontSize == size {
			return true
		}
	}
	return false
}

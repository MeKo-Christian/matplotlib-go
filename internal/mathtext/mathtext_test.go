package mathtext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	return ""
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

func TestLayoutMathTextUsesRuleDelimitersForStretchyBrackets(t *testing.T) {
	layout, ok := LayoutMathText(testMeasurer{}, `\left[\frac{1}{2}\right]`, 20, "base", Options{})
	if !ok {
		t.Fatal("LayoutMathText returned !ok")
	}
	if len(layout.Rules) < 7 {
		t.Fatalf("expected fraction rule plus bracket rule pieces, got %d rules: %+v", len(layout.Rules), layout.Rules)
	}
	if len(layout.Runs) != 2 {
		t.Fatalf("expected only numerator and denominator text runs, got %+v", layout.Runs)
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
	if layout.Runs[0].Text != "(" || layout.Runs[len(layout.Runs)-1].Text != ")" {
		t.Fatalf("binom did not add parenthesized delimiters: %+v", layout.Runs)
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
	if len(layout.Rules) < 4 {
		t.Fatalf("genfrac did not apply requested bracket delimiters as rule boxes: %+v", layout.Rules)
	}
	if !containsTestRun(layout.Runs, "n", 20) || !containsTestRun(layout.Runs, "k", 20) {
		t.Fatalf("display-style genfrac should keep numerator and denominator at base size: %+v", layout.Runs)
	}
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

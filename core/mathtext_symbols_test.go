package core

import (
	"strings"
	"testing"
)

// joinMathRunText concatenates the text of every layout run, so tests can assert
// on the rendered glyph content regardless of how the layout splits runs by
// font or kerning.
func joinMathRunText(runs []MathTextLayoutRun) string {
	var b strings.Builder
	for _, run := range runs {
		b.WriteString(run.Text)
	}
	return b.String()
}

// TestLayoutMathTextResolvesExpandedSymbols verifies that the expanded tex2uni
// table (Phase 8) lets common symbols render as their Unicode glyph instead of
// echoing the literal command.
func TestLayoutMathTextResolvesExpandedSymbols(t *testing.T) {
	var r textRecordingRenderer
	cases := []struct {
		expr  string
		glyph string
	}{
		{`\Rightarrow`, "⇒"},
		{`a \otimes b`, "⊗"},
		{`\cdots`, "⋯"},
		{`\vdash`, "⊢"},
		{`\aleph`, "ℵ"},
		{`\nabla \times F`, "∇"},
		{`x \leq y \geq z`, "≤"},
	}
	for _, tc := range cases {
		layout, ok := LayoutMathText(&r, tc.expr, 20, "DejaVu Sans")
		if !ok {
			t.Fatalf("LayoutMathText(%q) returned !ok", tc.expr)
		}
		joined := joinMathRunText(layout.Runs)
		if !strings.Contains(joined, tc.glyph) {
			t.Errorf("expr %q: missing glyph %q in runs %q", tc.expr, tc.glyph, joined)
		}
		if strings.Contains(joined, `\`) {
			t.Errorf("expr %q: literal-echo in runs %q", tc.expr, joined)
		}
	}
}

// TestLayoutMathTextResolvesMathAlphabets verifies the \mathbb / \mathcal /
// \mathfrak / \boldsymbol alphabets translate to the Unicode Mathematical
// Alphanumeric block (resolved via the per-glyph font fallback at draw time).
func TestLayoutMathTextResolvesMathAlphabets(t *testing.T) {
	var r textRecordingRenderer
	cases := []struct {
		expr  string
		glyph string
	}{
		{`\mathbb{R}`, "ℝ"},     // Letterlike-Symbols hole
		{`\mathbb{D}`, "𝔻"},     // contiguous double-struck
		{`\mathbb{1}`, "𝟙"},     // double-struck digit
		{`\mathcal{L}`, "ℒ"},    // script hole
		{`\mathfrak{g}`, "𝔤"},   // fraktur
		{`\boldsymbol{x}`, "𝒙"}, // bold italic letter
		{`\boldsymbol{2}`, "𝟐"}, // bold digit
	}
	for _, tc := range cases {
		layout, ok := LayoutMathText(&r, tc.expr, 20, "DejaVu Sans")
		if !ok {
			t.Fatalf("LayoutMathText(%q) returned !ok", tc.expr)
		}
		joined := joinMathRunText(layout.Runs)
		if !strings.Contains(joined, tc.glyph) {
			t.Errorf("expr %q: missing glyph %q in runs %q", tc.expr, tc.glyph, joined)
		}
		if strings.Contains(joined, `\`) {
			t.Errorf("expr %q: literal-echo in runs %q", tc.expr, joined)
		}
	}
}

// TestNormalizeDisplayTextResolvesExpandedSymbols guards the plain-text
// (Unicode fallback) path for the same expanded coverage.
func TestNormalizeDisplayTextResolvesExpandedSymbols(t *testing.T) {
	cases := map[string]string{
		`$A \Rightarrow B$`:               "A ⇒ B",
		`$\mathbb{Z} \subset \mathbb{R}$`: "ℤ ⊂ ℝ",
		`$\boldsymbol{v} \cdot w$`:        "𝒗 · w",
	}
	for in, want := range cases {
		if got := normalizeDisplayText(in); got != want {
			t.Errorf("normalizeDisplayText(%q) = %q, want %q", in, got, want)
		}
	}
}

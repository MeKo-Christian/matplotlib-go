package render

import "testing"

// TestResolveTextRunsMathFallback verifies the per-glyph fallback chain reaches
// the math-capable font (STIXGeneral) for glyphs the primary text font lacks,
// while glyphs the primary supports stay on the primary face. This is what lets
// expanded MathText symbols and the \mathbb/\mathfrak/\boldsymbol alphabets
// render instead of tofu.
func TestResolveTextRunsMathFallback(t *testing.T) {
	m := NewFontManager()

	// U+1D431 MATHEMATICAL BOLD SMALL X is absent from DejaVu Sans but present
	// in STIXGeneral, so it must resolve via the appended math fallback.
	runs, ok := m.ResolveTextRuns("\U0001D431", "DejaVu Sans")
	if !ok || len(runs) == 0 {
		t.Fatalf("math glyph not resolved: ok=%v runs=%d", ok, len(runs))
	}
	if got := runs[0].Face.Family; got != fontFamilyMathFallback {
		t.Errorf("math glyph resolved to %q, want %q", got, fontFamilyMathFallback)
	}

	// A plain ASCII run must stay on the primary face (no spurious fallback).
	runs, ok = m.ResolveTextRuns("abc", "DejaVu Sans")
	if !ok || len(runs) != 1 {
		t.Fatalf("ascii run: ok=%v runs=%d", ok, len(runs))
	}
	if got := runs[0].Face.Family; got != "DejaVu Sans" {
		t.Errorf("ascii resolved to %q, want DejaVu Sans", got)
	}

	// A mixed run splits into primary + fallback faces.
	runs, ok = m.ResolveTextRuns("a\U0001D431b", "DejaVu Sans")
	if !ok || len(runs) != 3 {
		t.Fatalf("mixed run: ok=%v runs=%d (want 3)", ok, len(runs))
	}
	if runs[1].Face.Family != fontFamilyMathFallback {
		t.Errorf("middle run face = %q, want %q", runs[1].Face.Family, fontFamilyMathFallback)
	}
}

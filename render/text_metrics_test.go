package render

import (
	"math"
	"testing"
)

// TestMeasureTextMetricsRealFont verifies the shared pure-Go metrics return
// font-derived, non-degenerate values (not the old crude char-count estimate).
func TestMeasureTextMetricsRealFont(t *testing.T) {
	const size = 12.0
	m := MeasureTextMetrics("Hg", size, "DejaVu Sans")
	if m.W <= 0 {
		t.Fatalf("width = %v, want > 0", m.W)
	}
	if m.Ascent <= 0 || m.Descent <= 0 {
		t.Fatalf("ascent/descent = %v/%v, want both > 0 ('g' descends, 'H' ascends)", m.Ascent, m.Descent)
	}
	// The crude legacy estimate was 0.6*size*len = 14.4 for "Hg"; the real
	// DejaVu advance differs, confirming we are no longer using the stub.
	if math.Abs(m.W-0.6*size*2) < 1e-9 {
		t.Fatalf("width %v matches the crude estimate; expected real font advance", m.W)
	}
	if got := m.Ascent + m.Descent; math.Abs(got-m.H) > 1e-9 {
		t.Fatalf("H = %v, want ascent+descent = %v", m.H, got)
	}
}

// TestMeasureTextMetricsDescenderSensitivity confirms descent now reflects the
// actual glyphs (a string without descenders reports less descent than one with).
func TestMeasureTextMetricsDescenderSensitivity(t *testing.T) {
	const size = 16.0
	withDesc := MeasureTextMetrics("pgjy", size, "DejaVu Sans")
	noDesc := MeasureTextMetrics("ABCD", size, "DejaVu Sans")
	if withDesc.Descent <= noDesc.Descent {
		t.Fatalf("descender string descent %v should exceed cap-only descent %v", withDesc.Descent, noDesc.Descent)
	}
}

func TestMeasureTextMetricsEmptyAndDegenerate(t *testing.T) {
	if got := MeasureTextMetrics("", 12, "DejaVu Sans"); got != (TextMetrics{}) {
		t.Fatalf("empty text = %+v, want zero", got)
	}
	if got := MeasureTextMetrics("x", 0, "DejaVu Sans"); got != (TextMetrics{}) {
		t.Fatalf("size 0 = %+v, want zero", got)
	}
}

func TestMeasureFontHeightMetrics(t *testing.T) {
	const size = 20.0
	fh, ok := MeasureFontHeightMetrics(size, "DejaVu Sans")
	if !ok {
		t.Skip("DejaVu Sans not resolvable in this environment")
	}
	if fh.Ascent <= 0 || fh.Descent <= 0 {
		t.Fatalf("font heights = %+v, want positive ascent/descent", fh)
	}
	// Sanity: font-wide ascent should not exceed the em size by an absurd factor.
	if fh.Ascent > 4*size {
		t.Fatalf("ascent %v implausibly large for %vpx em", fh.Ascent, size)
	}
}

func TestMeasureTextInkBounds(t *testing.T) {
	b, ok := MeasureTextInkBounds("Hg", 14, "DejaVu Sans")
	if !ok {
		t.Skip("DejaVu Sans not resolvable in this environment")
	}
	if b.W <= 0 || b.H <= 0 {
		t.Fatalf("ink bounds = %+v, want positive extent", b)
	}
	// Y is the baseline-relative top (negative above the baseline).
	if b.Y >= 0 {
		t.Fatalf("ink bounds Y = %v, want negative (ink rises above baseline)", b.Y)
	}
}

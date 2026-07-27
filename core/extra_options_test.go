package core

import "testing"

// The options-model pass finished what the variadic-tail rejection started: every plotting
// entry point now takes exactly one option value, so passing a second one is a
// compile error rather than a run-time rejection. The tests that exercised that
// rejection are gone with the tails they guarded; what remains is the guard
// against the opposite mistake — an entry point that refuses the one value it
// is supposed to accept.

// TestSingleOptionValueIsStillAccepted guards against the rejection firing one
// value too early.
func TestSingleOptionValueIsStillAccepted(t *testing.T) {
	x := []float64{0, 1, 2}
	y := []float64{0, 1, 2}

	ax := newAlphaTestAxes()
	if line, err := ax.Plot(x, y, PlotOptions{}); err != nil || line == nil {
		t.Fatalf("Plot() = (%v, %v), want an artist and no error", line, err)
	}
	if line := ax.SemilogX(x, y, PlotOptions{}); line == nil {
		t.Fatal("SemilogX() = nil, want an artist")
	}
	if img := ax.ImShow([][]float64{{0, 1}, {1, 2}}, ImShowOptions{}); img == nil {
		t.Fatal("ImShow() = nil, want an artist")
	}
	if txt := ax.Text(0, 0, "t", TextOptions{}); txt == nil {
		t.Fatal("Text() = nil, want an artist")
	}
}

package core

import "testing"

// TestFigureSizePxMatchesMatplotlibInchProduct pins the figure display extent to
// the value matplotlib derives, float64 noise included. Matplotlib sizes a
// figure in inches and computes pixels as figsize*dpi, so a figure asked for in
// pixels round-trips through px/dpi*dpi — which is not the identity. Losing this
// moves any tick label whose centred origin lands on an exact .5.
func TestFigureSizePxMatchesMatplotlibInchProduct(t *testing.T) {
	// Written as runtime float64 arithmetic on purpose: Go folds a constant
	// expression like 9.8*100 at arbitrary precision and would yield exactly 980,
	// which is the value this test exists to reject.
	inches := func(px int) float64 { return float64(px) / 100 * 100 }

	tests := []struct {
		name   string
		w, h   int
		dpi    float64
		wantX  float64
		wantY  float64
		exactX bool
	}{
		{name: "inexact width", w: 980, h: 620, dpi: 100, wantX: inches(980), wantY: inches(620)},
		{name: "exact both", w: 640, h: 360, dpi: 100, wantX: 640, wantY: 360, exactX: true},
		{name: "inexact width 930", w: 930, h: 340, dpi: 100, wantX: inches(930), wantY: inches(340)},
		{name: "zero dpi falls back", w: 980, h: 620, dpi: 0, wantX: 980, wantY: 620, exactX: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := figureSizePx(tc.w, tc.h, tc.dpi)
			if got.X != tc.wantX || got.Y != tc.wantY {
				t.Fatalf("figureSizePx(%d, %d, %v) = (%v, %v), want (%v, %v)",
					tc.w, tc.h, tc.dpi, got.X, got.Y, tc.wantX, tc.wantY)
			}
			if tc.exactX && got.X != float64(tc.w) {
				t.Fatalf("figureSizePx(%d, ...).X = %v, want exactly %d", tc.w, got.X, tc.w)
			}
		})
	}
}

// TestNewFigureUsesDerivedSizePx checks the constructor routes through the
// derivation rather than casting the integer directly.
func TestNewFigureUsesDerivedSizePx(t *testing.T) {
	fig := NewFigure(980, 620)
	if fig.RC.DPI != 100 {
		t.Skipf("default DPI is %v, not the 100 this case is calibrated for", fig.RC.DPI)
	}
	want := float64(980) / 100 * 100
	if fig.SizePx.X != want {
		t.Fatalf("NewFigure(980, 620).SizePx.X = %v, want %v", fig.SizePx.X, want)
	}
	if fig.SizePx.X == 980 {
		t.Fatal("NewFigure(980, 620).SizePx.X is exactly 980; matplotlib computes 980.0000000000001")
	}
	if fig.SizePx.Y != 620 {
		t.Fatalf("NewFigure(980, 620).SizePx.Y = %v, want exactly 620", fig.SizePx.Y)
	}
}

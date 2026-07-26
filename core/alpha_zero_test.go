package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

func newAlphaTestAxes() *Axes {
	fig := NewFigure(800, 600)
	return fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
}

// maxFillAlpha returns the largest fill alpha across all recorded path calls,
// or 0 if nothing with a fill was drawn.
func maxFillAlpha(r *recordingRenderer) float64 {
	maxA := 0.0
	for _, c := range r.pathCalls {
		if c.paint.Fill.A > maxA {
			maxA = c.paint.Fill.A
		}
	}
	return maxA
}

// An explicit alpha of 0 must produce a fully transparent fill, not be treated
// as "unset" and silently promoted to opaque. A fully transparent bar draws no
// opaque fill at all; an opaque control proves the bar would otherwise paint.
func TestBarHonorsExplicitZeroAlpha(t *testing.T) {
	red := render.Color{R: 1, A: 1}
	ctx := createTestDrawContext()

	zero := 0.0
	transparent := &recordingRenderer{}
	transparentBar, err := newAlphaTestAxes().
		Bar([]float64{0, 1, 2}, []float64{1, 2, 3}, BarOptions{Color: optional.Of(red), Alpha: optional.Of(zero)})
	if err != nil {
		t.Fatalf("Bar() returned error: %v", err)
	}
	transparentBar.Draw(transparent, ctx)
	if got := maxFillAlpha(transparent); got != 0 {
		t.Fatalf("Bar alpha=0 should draw no opaque fill, got max fill alpha %v", got)
	}

	opaque := &recordingRenderer{}
	opaqueBar, err := newAlphaTestAxes().
		Bar([]float64{0, 1, 2}, []float64{1, 2, 3}, BarOptions{Color: optional.Of(red)})
	if err != nil {
		t.Fatalf("Bar() returned error: %v", err)
	}
	opaqueBar.Draw(opaque, ctx)
	if maxFillAlpha(opaque) == 0 {
		t.Fatal("control opaque bar drew no fill — harness problem, not a real pass")
	}
}

func TestBarHalfAlphaApplied(t *testing.T) {
	ax := newAlphaTestAxes()
	half := 0.5
	red := render.Color{R: 1, A: 1}
	bar, err := ax.Bar([]float64{0}, []float64{1}, BarOptions{Color: optional.Of(red), Alpha: optional.Of(half)})
	if err != nil {
		t.Fatalf("Bar() returned error: %v", err)
	}

	r := &recordingRenderer{}
	bar.Draw(r, createTestDrawContext())

	if got := maxFillAlpha(r); got != 0.5 {
		t.Fatalf("Bar alpha=0.5 fill A=%v, want 0.5", got)
	}
}

func TestBarNilAlphaPreservesColorAlpha(t *testing.T) {
	ax := newAlphaTestAxes()
	semi := render.Color{R: 1, A: 0.8}
	bar, err := ax.Bar([]float64{0}, []float64{1}, BarOptions{Color: optional.Of(semi)})
	if err != nil {
		t.Fatalf("Bar() returned error: %v", err)
	}

	r := &recordingRenderer{}
	bar.Draw(r, createTestDrawContext())

	if got := maxFillAlpha(r); got != 0.8 {
		t.Fatalf("Bar nil alpha should preserve color A=0.8, got %v", got)
	}
}

func TestFillBetweenHonorsExplicitZeroAlpha(t *testing.T) {
	red := render.Color{R: 1, A: 1}
	ctx := createTestDrawContext()

	zero := 0.0
	transparent := &recordingRenderer{}
	transparentFill, err := newAlphaTestAxes().
		FillBetween([]float64{0, 1, 2}, []float64{0, 0, 0}, []float64{1, 2, 1}, FillOptions{Color: optional.Of(red), Alpha: optional.Of(zero)})
	if err != nil {
		t.Fatalf("FillBetween() returned error: %v", err)
	}
	transparentFill.Draw(transparent, ctx)
	if got := maxFillAlpha(transparent); got != 0 {
		t.Fatalf("FillBetween alpha=0 should draw no opaque fill, got max fill alpha %v", got)
	}

	opaque := &recordingRenderer{}
	opaqueFill, err := newAlphaTestAxes().
		FillBetween([]float64{0, 1, 2}, []float64{0, 0, 0}, []float64{1, 2, 1}, FillOptions{Color: optional.Of(red)})
	if err != nil {
		t.Fatalf("FillBetween() returned error: %v", err)
	}
	opaqueFill.Draw(opaque, ctx)
	if maxFillAlpha(opaque) == 0 {
		t.Fatal("control opaque fill drew nothing — harness problem")
	}
}

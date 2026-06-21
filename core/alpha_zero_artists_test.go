package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

// maxAnyAlpha returns the largest non-zero alpha across every recorded fill and
// stroke, or 0 if nothing visible was drawn. Use it for artists that paint via
// strokes (error bars, steps) as well as fills.
func maxAnyAlpha(r *recordingRenderer) float64 {
	maxA := 0.0
	for _, c := range r.pathCalls {
		if c.paint.Fill.A > maxA {
			maxA = c.paint.Fill.A
		}
		if c.paint.Stroke.A > maxA {
			maxA = c.paint.Stroke.A
		}
	}
	return maxA
}

// Hist must honor an explicit alpha of 0 (fully transparent) instead of
// promoting the silent 0 sentinel to opaque.
func TestHistHonorsExplicitZeroAlpha(t *testing.T) {
	red := render.Color{R: 1, A: 1}
	ctx := createTestDrawContext()
	data := []float64{1, 2, 2, 3, 3, 3}

	zero := 0.0
	transparent := &recordingRenderer{}
	newAlphaTestAxes().
		Hist(data, HistOptions{Color: &red, Alpha: &zero}).
		Draw(transparent, ctx)
	if got := maxAnyAlpha(transparent); got != 0 {
		t.Fatalf("Hist alpha=0 should draw nothing opaque, got max alpha %v", got)
	}

	opaque := &recordingRenderer{}
	newAlphaTestAxes().
		Hist(data, HistOptions{Color: &red}).
		Draw(opaque, ctx)
	if maxAnyAlpha(opaque) == 0 {
		t.Fatal("control opaque hist drew nothing — harness problem")
	}
}

// ErrorBar must honor an explicit alpha of 0.
func TestErrorBarHonorsExplicitZeroAlpha(t *testing.T) {
	red := render.Color{R: 1, A: 1}
	ctx := createTestDrawContext()
	x := []float64{0, 1, 2}
	y := []float64{1, 2, 3}
	e := []float64{0.2, 0.2, 0.2}

	zero := 0.0
	transparent := &recordingRenderer{}
	newAlphaTestAxes().
		ErrorBar(x, y, e, e, ErrorBarOptions{Color: &red, Alpha: &zero}).
		Draw(transparent, ctx)
	if got := maxAnyAlpha(transparent); got != 0 {
		t.Fatalf("ErrorBar alpha=0 should draw nothing opaque, got max alpha %v", got)
	}

	opaque := &recordingRenderer{}
	newAlphaTestAxes().
		ErrorBar(x, y, e, e, ErrorBarOptions{Color: &red}).
		Draw(opaque, ctx)
	if maxAnyAlpha(opaque) == 0 {
		t.Fatal("control opaque error bar drew nothing — harness problem")
	}
}

// BoxPlot must honor an explicit alpha of 0 for its box fill/edge.
func TestBoxPlotHonorsExplicitZeroAlpha(t *testing.T) {
	red := render.Color{R: 1, A: 1}
	ctx := createTestDrawContext()
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}

	zero := 0.0
	transparent := &recordingRenderer{}
	box := newAlphaTestAxes().BoxPlot(data, BoxPlotOptions{Color: &red, EdgeColor: &red, Alpha: &zero})
	box.Draw(transparent, ctx)
	maxFill := 0.0
	for _, c := range transparent.pathCalls {
		if c.paint.Fill.A > maxFill {
			maxFill = c.paint.Fill.A
		}
	}
	if maxFill != 0 {
		t.Fatalf("BoxPlot alpha=0 should draw no opaque box fill, got max fill alpha %v", maxFill)
	}

	opaque := &recordingRenderer{}
	newAlphaTestAxes().BoxPlot(data, BoxPlotOptions{Color: &red, EdgeColor: &red}).Draw(opaque, ctx)
	ctlFill := 0.0
	for _, c := range opaque.pathCalls {
		if c.paint.Fill.A > ctlFill {
			ctlFill = c.paint.Fill.A
		}
	}
	if ctlFill == 0 {
		t.Fatal("control opaque box plot drew no fill — harness problem")
	}
}

// Step delegates to Plot, which already bakes an explicit alpha (including 0)
// into the line color. This locks that behavior in as a regression guard.
func TestStepHonorsExplicitZeroAlpha(t *testing.T) {
	red := render.Color{R: 1, A: 1}
	ctx := createTestDrawContext()
	x := []float64{0, 1, 2}
	y := []float64{1, 2, 3}

	zero := 0.0
	transparent := &recordingRenderer{}
	newAlphaTestAxes().
		Step(x, y, StepOptions{Color: &red, Alpha: &zero}).
		Draw(transparent, ctx)
	if got := maxAnyAlpha(transparent); got != 0 {
		t.Fatalf("Step alpha=0 should draw nothing opaque, got max alpha %v", got)
	}

	opaque := &recordingRenderer{}
	newAlphaTestAxes().
		Step(x, y, StepOptions{Color: &red}).
		Draw(opaque, ctx)
	if maxAnyAlpha(opaque) == 0 {
		t.Fatal("control opaque step drew nothing — harness problem")
	}
}

package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

// drawNull renders the figure through a no-op renderer so the persistent
// transform graph is exercised exactly as in a real draw.
func drawNull(fig *Figure) {
	var r render.NullRenderer
	DrawFigure(fig, &r)
}

func TestAxesTransformReusedWhenSizeUnchanged(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{})

	drawNull(fig)
	if ax.axesBbox == nil {
		t.Fatal("axesBbox was not initialized by the draw")
	}
	v1 := ax.axesBbox.Version()

	// Redrawing at the same size must not invalidate the axes->pixel transform.
	drawNull(fig)
	if v2 := ax.axesBbox.Version(); v2 != v1 {
		t.Fatalf("redraw at same size invalidated axesBbox: version %d -> %d", v1, v2)
	}
}

func TestAxesTransformInvalidatedOnResize(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{})

	drawNull(fig)
	v1 := ax.axesBbox.Version()
	before := ax.transData.Apply(geom.Pt{X: 1, Y: 1})

	// Resize the figure: the axes pixel rectangle changes, so the persistent
	// transform must be invalidated and recompute a new mapping.
	fig.SizePx = geom.Pt{X: 800, Y: 600}
	drawNull(fig)

	if v2 := ax.axesBbox.Version(); v2 == v1 {
		t.Fatalf("resize did not invalidate axesBbox (version stayed %d)", v1)
	}
	after := ax.transData.Apply(geom.Pt{X: 1, Y: 1})
	if approxPtCore(before, after, 1e-9) {
		t.Fatalf("transData did not change after resize: %+v == %+v", before, after)
	}
}

// TestRefreshDataTransformInvalidationStage asserts the invalidation stage fired
// to dataNode matches the kind of data leg: an affine leg fires InvalidAffine,
// while a non-affine leg (e.g. a log scale) fires InvalidNonAffine so a
// split-aware consumer re-runs its non-affine vertex pass instead of reusing a
// stale projection (Phase 13).
func TestRefreshDataTransformInvalidationStage(t *testing.T) {
	ax := &Axes{}
	ax.ensureTransforms()

	// A probe dependent observes exactly the stage propagated from dataNode.
	var probe transform.TransformNode
	ax.dataNode.AddDependent(&probe)

	linear := transform.NewScaleTransform(transform.NewLinear(0, 1), transform.NewLinear(0, 1))
	logLeg := transform.NewScaleTransform(transform.NewLog(1, 10, 10), transform.NewLog(1, 10, 10))

	// Sanity: the log leg is genuinely non-affine.
	if _, ok := transform.AsAffine(logLeg); ok {
		t.Fatal("log scale leg unexpectedly reported as affine")
	}

	probe.ClearInvalid()
	ax.refreshDataTransform(linear)
	if got := probe.Invalid(); !got.Has(transform.InvalidAffine) || got.Has(transform.InvalidNonAffine) {
		t.Fatalf("affine leg: want InvalidAffine only, got %v", got)
	}

	probe.ClearInvalid()
	ax.refreshDataTransform(logLeg)
	if got := probe.Invalid(); !got.Has(transform.InvalidNonAffine) {
		t.Fatalf("non-affine leg: want InvalidNonAffine, got %v", got)
	}

	// Re-applying the same non-affine leg must NOT re-fire: an unchanged leg lets
	// a split-aware consumer reuse its cached projection across draws (Phase 13
	// leg-change detection). A structurally-equal rebuilt leg counts as unchanged.
	probe.ClearInvalid()
	logLegAgain := transform.NewScaleTransform(transform.NewLog(1, 10, 10), transform.NewLog(1, 10, 10))
	ax.refreshDataTransform(logLegAgain)
	if got := probe.Invalid(); got.Has(transform.InvalidNonAffine) {
		t.Fatalf("unchanged non-affine leg: want no InvalidNonAffine, got %v", got)
	}

	// Changing the leg (different log domain) must re-fire InvalidNonAffine so the
	// stale projection is rebuilt.
	probe.ClearInvalid()
	logLegChanged := transform.NewScaleTransform(transform.NewLog(1, 100, 10), transform.NewLog(1, 10, 10))
	ax.refreshDataTransform(logLegChanged)
	if got := probe.Invalid(); !got.Has(transform.InvalidNonAffine) {
		t.Fatalf("changed non-affine leg: want InvalidNonAffine, got %v", got)
	}
}

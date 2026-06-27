package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

// countingTransform is a non-affine transform.T (it is not one of the affine
// types splitAffine recognizes, so it is treated as the non-affine remainder)
// that counts how many times Apply runs. It is used to observe whether the
// cached non-affine projection is re-run across a redraw.
type countingTransform struct{ applies *int }

func (c countingTransform) Apply(p geom.Pt) geom.Pt {
	if c.applies != nil {
		*c.applies++
	}
	return geom.Pt{X: p.X*2 + 1, Y: p.Y*3 - 2}
}

func (c countingTransform) Invert(p geom.Pt) (geom.Pt, bool) {
	return geom.Pt{X: (p.X - 1) / 2, Y: (p.Y + 2) / 3}, true
}

// newCacheTestAxesContext builds a minimal axes with its persistent transform
// graph pointed at dataToAxes and a DrawContext that resolves data coordinates
// through that graph (composed == ax.transData).
func newCacheTestAxesContext(dataToAxes transform.T, rect geom.Rect) (*Axes, *DrawContext) {
	ax := &Axes{}
	ax.ensureTransforms()
	ax.axesBbox.Set(rect)
	ax.refreshDataTransform(dataToAxes)
	ctx := &DrawContext{
		Axes: ax,
		DataToPixel: Transform2D{
			DataToAxes:  dataToAxes,
			AxesToPixel: ax.transAxes,
			composed:    ax.transData,
		},
	}
	return ax, ctx
}

func pathsEqualExact(a, b geom.Path) bool {
	if len(a.C) != len(b.C) || len(a.V) != len(b.V) {
		return false
	}
	for i := range a.C {
		if a.C[i] != b.C[i] {
			return false
		}
	}
	for i := range a.V {
		if a.V[i] != b.V[i] {
			return false
		}
	}
	return true
}

// TestLine2DDisplayPathCacheParity asserts the cached data-coordinate draw path
// produces byte-identical output to the direct per-vertex transform, across
// affine and non-affine legs, step styles, and NaN gaps. The reference is the
// historical algorithm (transformAndSegment).
func TestLine2DDisplayPathCacheParity(t *testing.T) {
	nan := math.NaN()
	rect := geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 410, Y: 320}}

	linear := transform.NewScaleTransform(transform.NewLinear(0, 4), transform.NewLinear(0, 3))
	logLeg := transform.NewScaleTransform(transform.NewLog(1, 1000, 10), transform.NewLog(1, 1000, 10))
	nonAffine := countingTransform{} // genuine non-affine remainder

	legs := []struct {
		name string
		leg  transform.T
	}{
		{"linear-affine", linear},
		{"log-nonaffine", logLeg},
		{"counting-nonaffine", nonAffine},
	}

	inputs := []struct {
		name      string
		xy        []geom.Pt
		drawStyle LineDrawStyle
	}{
		{"default", []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 0.5}, {X: 3, Y: 2}}, LineDrawStyleDefault},
		{"nan-gap", []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: nan}, {X: 2, Y: 0.5}, {X: 3, Y: 2}}, LineDrawStyleDefault},
		{"steps-pre", []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 0.5}}, LineDrawStyleStepsPre},
		{"steps-mid", []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 0.5}}, LineDrawStyleStepsMid},
		{"steps-post", []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 0.5}}, LineDrawStyleStepsPost},
	}

	for _, leg := range legs {
		for _, in := range inputs {
			t.Run(leg.name+"/"+in.name, func(t *testing.T) {
				_, ctx := newCacheTestAxesContext(leg.leg, rect)
				l := &Line2D{XY: append([]geom.Pt(nil), in.xy...), DrawStyle: in.drawStyle}

				got := l.displayPath(ctx)

				// Reference: the direct per-vertex algorithm on the same prepared
				// points and resolved transform.
				tr := artistTransformFor(ctx, l, Coords(CoordData))
				want := transformAndSegment(l.pathPoints(), tr)

				if !pathsEqualExact(got, want) {
					t.Fatalf("cached path diverged from direct path\n got: %+v\nwant: %+v", got, want)
				}
			})
		}
	}
}

// TestLine2DDisplayPathReusesNonAffineProjection drives Line2D.displayPath
// directly with a non-affine data leg and a live bbox, then changes only the
// bbox (an affine-only redraw) and asserts the non-affine projection is reused
// (Apply count unchanged) while the output reflects the new affine. This proves
// the Phase 13 cache split at the artist level; driving displayPath directly
// avoids the every-draw refreshDataTransform that re-projects non-affine legs
// end-to-end (the deferred leg-change-detection follow-up).
func TestLine2DDisplayPathReusesNonAffineProjection(t *testing.T) {
	applies := 0
	counting := countingTransform{applies: &applies}
	ax, ctx := newCacheTestAxesContext(counting, geom.Rect{Max: geom.Pt{X: 100, Y: 100}})

	l := &Line2D{XY: []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 0.5}, {X: 3, Y: 2}}}
	n := len(l.XY)

	first := l.displayPath(ctx)
	if len(first.V) != n {
		t.Fatalf("first draw produced %d vertices, want %d", len(first.V), n)
	}
	if applies != n {
		t.Fatalf("first draw: non-affine pass ran %d times, want %d", applies, n)
	}

	// Affine-only change: resize the axes bbox. The data leg is untouched.
	ax.axesBbox.Set(geom.Rect{Max: geom.Pt{X: 200, Y: 150}})
	second := l.displayPath(ctx)

	if applies != n {
		t.Fatalf("after affine-only redraw: non-affine pass ran %d times total, want %d (projection should be reused)", applies, n)
	}
	if pathsEqualExact(first, second) {
		t.Fatal("resized redraw produced an identical path; trailing affine was not refreshed")
	}
}

// TestLine2DDisplayPathReusesProjectionThroughRefresh is the end-to-end companion
// to the artist-level probe above: it drives the every-draw refreshDataTransform
// between draws (as a full figure draw does). With leg-change detection, an
// unchanged non-affine leg no longer fires InvalidNonAffine, so a resize that only
// moves the axes bbox reuses the cached projection; only an actual leg change
// re-projects (Phase 13 leg-change detection).
func TestLine2DDisplayPathReusesProjectionThroughRefresh(t *testing.T) {
	applies := 0
	counting := countingTransform{applies: &applies}
	ax, ctx := newCacheTestAxesContext(counting, geom.Rect{Max: geom.Pt{X: 100, Y: 100}})

	l := &Line2D{XY: []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 0.5}, {X: 3, Y: 2}}}
	n := len(l.XY)

	first := l.displayPath(ctx)
	if applies != n {
		t.Fatalf("first draw: non-affine pass ran %d times, want %d", applies, n)
	}

	// Second "draw": resize the bbox AND re-run refreshDataTransform with an
	// unchanged (structurally-equal) leg, exactly as the figure pipeline does.
	ax.axesBbox.Set(geom.Rect{Max: geom.Pt{X: 200, Y: 150}})
	ax.refreshDataTransform(counting)
	second := l.displayPath(ctx)

	if applies != n {
		t.Fatalf("after resize + refresh with unchanged leg: non-affine pass ran %d times total, want %d (projection should be reused)", applies, n)
	}
	if pathsEqualExact(first, second) {
		t.Fatal("resized redraw produced an identical path; trailing affine was not refreshed")
	}

	// Third "draw": the leg genuinely changes (different non-affine map), which
	// must re-project.
	ax.refreshDataTransform(scaledCountingTransform{applies: &applies})
	_ = l.displayPath(ctx)
	if applies != n+n {
		t.Fatalf("after leg change: non-affine pass ran %d times total, want %d (projection should be rebuilt)", applies, n+n)
	}
}

// scaledCountingTransform is a second, distinct non-affine transform so a leg
// change is observable by reflect.DeepEqual (different dynamic type than
// countingTransform).
type scaledCountingTransform struct{ applies *int }

func (c scaledCountingTransform) Apply(p geom.Pt) geom.Pt {
	if c.applies != nil {
		*c.applies++
	}
	return geom.Pt{X: p.X*4 + 3, Y: p.Y*5 - 1}
}

func (c scaledCountingTransform) Invert(p geom.Pt) (geom.Pt, bool) {
	return geom.Pt{X: (p.X - 3) / 4, Y: (p.Y + 1) / 5}, true
}

// TestLine2DDisplayPathInvalidatesOnDataChange asserts that replacing the source
// data re-runs the non-affine projection (the cache is keyed on the point
// backing, not stale across SetData).
func TestLine2DDisplayPathInvalidatesOnDataChange(t *testing.T) {
	applies := 0
	counting := countingTransform{applies: &applies}
	_, ctx := newCacheTestAxesContext(counting, geom.Rect{Max: geom.Pt{X: 100, Y: 100}})

	l := &Line2D{XY: []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}}}
	_ = l.displayPath(ctx)
	if applies != 2 {
		t.Fatalf("first draw: applied %d times, want 2", applies)
	}

	// Replace the data: SetData allocates a fresh backing slice, so the cache
	// must reproject.
	l.SetData([]float64{0, 1, 2}, []float64{0, 1, 2})
	_ = l.displayPath(ctx)
	if applies != 2+3 {
		t.Fatalf("after SetData: applied %d times total, want %d", applies, 2+3)
	}
}

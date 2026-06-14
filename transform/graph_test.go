package transform

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestRectTransformRoundTrip(t *testing.T) {
	src := geom.Rect{Min: geom.Pt{X: -2, Y: -1}, Max: geom.Pt{X: 4, Y: 3}}
	dst := geom.Rect{Min: geom.Pt{X: 10, Y: 100}, Max: geom.Pt{X: 70, Y: 20}}

	tr := NewRectTransform(src, dst)
	pt := geom.Pt{X: 1, Y: 2}

	got := tr.Apply(pt)
	want := geom.Pt{X: 40, Y: 40}
	if !approxPt(got, want, 1e-9) {
		t.Fatalf("Apply() = %+v, want %+v", got, want)
	}

	inv, ok := tr.Invert(got)
	if !ok {
		t.Fatal("Invert() failed")
	}
	if !approxPt(inv, pt, 1e-9) {
		t.Fatalf("Invert(Apply(pt)) = %+v, want %+v", inv, pt)
	}
}

func TestBlendUsesIndependentAxes(t *testing.T) {
	x := NewUnitRectTransform(geom.Rect{
		Min: geom.Pt{X: 20, Y: 10},
		Max: geom.Pt{X: 120, Y: 210},
	})
	y := NewScaleTransform(NewLinear(0, 10), NewLinear(-5, 5))

	tr := Blend(x, y)
	got := tr.Apply(geom.Pt{X: 0.25, Y: 0})
	want := geom.Pt{X: 45, Y: 0.5}
	if !approxPt(got, want, 1e-9) {
		t.Fatalf("Apply() = %+v, want %+v", got, want)
	}

	inv, ok := tr.Invert(got)
	if !ok {
		t.Fatal("Invert() failed")
	}
	if !approxPt(inv, geom.Pt{X: 0.25, Y: 0}, 1e-9) {
		t.Fatalf("Invert(Apply(pt)) = %+v", inv)
	}
}

func TestOffsetTransformRoundTrip(t *testing.T) {
	base := NewUnitRectTransform(geom.Rect{
		Min: geom.Pt{X: 50, Y: 150},
		Max: geom.Pt{X: 250, Y: 350},
	})
	tr := NewOffset(base, geom.Pt{X: 12, Y: -8})

	got := tr.Apply(geom.Pt{X: 0.5, Y: 0.25})
	want := geom.Pt{X: 162, Y: 192}
	if !approxPt(got, want, 1e-9) {
		t.Fatalf("Apply() = %+v, want %+v", got, want)
	}

	inv, ok := tr.Invert(got)
	if !ok {
		t.Fatal("Invert() failed")
	}
	if !approxPt(inv, geom.Pt{X: 0.5, Y: 0.25}, 1e-9) {
		t.Fatalf("Invert(Apply(pt)) = %+v", inv)
	}
}

func TestFrozenTransformSnapshotsCachedGraph(t *testing.T) {
	var source TransformNode
	builds := 0
	cached := NewCachedTransform(func() T {
		builds++
		return NewOffset(
			NewSeparable(
				ScaleAxis{Scale: NewLinear(0, float64(builds))},
				LinearAxis{Scale: float64(builds), Offset: 1},
			),
			geom.Pt{X: float64(builds), Y: 0},
		)
	}, &source)

	frozen := Frozen(cached)
	before := frozen.Apply(geom.Pt{X: 0.5, Y: 2})

	source.Invalidate(InvalidAffine)
	afterCached := cached.Apply(geom.Pt{X: 0.5, Y: 2})
	afterFrozen := frozen.Apply(geom.Pt{X: 0.5, Y: 2})

	if before != afterFrozen {
		t.Fatalf("frozen transform changed after source invalidation: before=%+v after=%+v", before, afterFrozen)
	}
	if afterCached == afterFrozen {
		t.Fatalf("cached transform did not diverge from frozen snapshot: cached=%+v frozen=%+v", afterCached, afterFrozen)
	}
}

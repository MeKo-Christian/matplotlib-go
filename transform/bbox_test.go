package transform

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func rectFor(minX, minY, maxX, maxY float64) geom.Rect {
	return geom.Rect{Min: geom.Pt{X: minX, Y: minY}, Max: geom.Pt{X: maxX, Y: maxY}}
}

func wantPt(t *testing.T, got, want geom.Pt, tol float64) {
	t.Helper()
	if !approxPt(got, want, tol) {
		t.Fatalf("point mismatch: got %v want %v", got, want)
	}
}

func TestBboxSetInvalidatesOnlyWhenChanged(t *testing.T) {
	box := NewBbox(rectFor(0, 0, 10, 20))
	start := box.Version()

	box.Set(rectFor(0, 0, 10, 20)) // same rect
	if box.Version() != start {
		t.Fatalf("setting the same rect must not invalidate: version %d -> %d", start, box.Version())
	}

	box.Set(rectFor(0, 0, 30, 40)) // different rect
	if box.Version() == start {
		t.Fatalf("changing the rect must invalidate (version unchanged at %d)", box.Version())
	}
	if got := box.Rect(); got != rectFor(0, 0, 30, 40) {
		t.Fatalf("Rect() = %v, want %v", got, rectFor(0, 0, 30, 40))
	}
}

func TestBboxTransformToMatchesUnitRectTransform(t *testing.T) {
	rect := rectFor(2, 3, 12, 23)
	box := NewBbox(rect)
	live := NewBboxTransformTo(box)
	want := NewUnitRectTransform(rect)

	for _, p := range []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 0.5, Y: 0.25}} {
		wantPt(t, live.Apply(p), want.Apply(p), 1e-12)
	}
}

func TestBboxTransformToRecomputesAfterResize(t *testing.T) {
	box := NewBbox(rectFor(0, 0, 10, 10))
	live := NewBboxTransformTo(box)

	wantPt(t, live.Apply(geom.Pt{X: 1, Y: 1}), geom.Pt{X: 10, Y: 10}, 1e-12)

	box.Set(rectFor(0, 0, 100, 50))
	wantPt(t, live.Apply(geom.Pt{X: 1, Y: 1}), geom.Pt{X: 100, Y: 50}, 1e-12)
	wantPt(t, live.Apply(geom.Pt{X: 0.5, Y: 0}), geom.Pt{X: 50, Y: 0}, 1e-12)
}

func TestBboxTransformFromMatchesRectTransform(t *testing.T) {
	rect := rectFor(2, 4, 12, 24)
	box := NewBbox(rect)
	live := NewBboxTransformFrom(box)
	unit := rectFor(0, 0, 1, 1)
	want := NewRectTransform(rect, unit)

	for _, p := range []geom.Pt{{X: 2, Y: 4}, {X: 12, Y: 24}, {X: 7, Y: 14}} {
		wantPt(t, live.Apply(p), want.Apply(p), 1e-12)
	}
}

func TestBboxTransformBetweenMatchesRectTransform(t *testing.T) {
	in := NewBbox(rectFor(0, 0, 10, 10))
	out := NewBbox(rectFor(100, 200, 300, 600))
	live := NewBboxTransform(in, out)
	want := NewRectTransform(in.Rect(), out.Rect())

	for _, p := range []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 10}, {X: 5, Y: 2}} {
		wantPt(t, live.Apply(p), want.Apply(p), 1e-9)
	}

	// Resizing either bbox updates the live transform.
	out.Set(rectFor(0, 0, 20, 20))
	wantPt(t, live.Apply(geom.Pt{X: 10, Y: 10}), geom.Pt{X: 20, Y: 20}, 1e-9)
}

func TestBboxTransformToInvertRoundTrip(t *testing.T) {
	box := NewBbox(rectFor(3, 7, 53, 107))
	live := NewBboxTransformTo(box)
	p := geom.Pt{X: 0.3, Y: 0.8}
	mapped := live.Apply(p)
	back, ok := live.Invert(mapped)
	if !ok {
		t.Fatal("Invert returned ok=false")
	}
	wantPt(t, back, p, 1e-9)
}

func TestBboxTransformToIsSeparable(t *testing.T) {
	box := NewBbox(rectFor(2, 3, 12, 23))
	var live T = NewBboxTransformTo(box)
	sep, ok := live.(Separable)
	if !ok {
		t.Fatal("BboxTransformTo must implement Separable")
	}
	want := NewUnitRectTransform(box.Rect())
	if got := sep.XAxis().Forward(0.5); got != want.XAxis().Forward(0.5) {
		t.Fatalf("XAxis mismatch: %v vs %v", got, want.XAxis().Forward(0.5))
	}
	if got := sep.YAxis().Forward(0.5); got != want.YAxis().Forward(0.5) {
		t.Fatalf("YAxis mismatch: %v vs %v", got, want.YAxis().Forward(0.5))
	}
}

func TestFrozenAndAsAffineResolveBboxTransformTo(t *testing.T) {
	box := NewBbox(rectFor(2, 3, 12, 23))
	live := NewBboxTransformTo(box)

	frozen := Frozen(live)
	if _, ok := frozen.(SeparableT); !ok {
		t.Fatalf("Frozen(BboxTransformTo) = %T, want SeparableT", frozen)
	}

	aff, ok := AsAffine(live)
	if !ok {
		t.Fatal("AsAffine(BboxTransformTo) returned ok=false")
	}
	// Affine of unit->rect: scale (w,h), translate (minX,minY).
	wantPt(t, aff.Apply(geom.Pt{X: 1, Y: 1}), geom.Pt{X: 12, Y: 23}, 1e-9)
}

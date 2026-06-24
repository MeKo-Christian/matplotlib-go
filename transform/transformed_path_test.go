package transform

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestTransformedPathCachesAndInvalidates(t *testing.T) {
	var source TransformNode
	builds := 0
	cachedTransform := NewCachedTransform(func() T {
		builds++
		return NewOffset(nil, geom.Pt{X: float64(builds), Y: 0})
	}, &source)

	var path geom.Path
	path.MoveTo(geom.Pt{X: 1, Y: 2})
	path.LineTo(geom.Pt{X: 3, Y: 4})

	transformed := NewTransformedPath(path, cachedTransform, &source)
	first := transformed.Transformed()
	second := transformed.Transformed()
	if builds != 1 {
		t.Fatalf("transform builds = %d, want 1 before invalidation", builds)
	}
	if first.V[0] != second.V[0] {
		t.Fatalf("cached transformed path changed without invalidation: first=%+v second=%+v", first, second)
	}

	first.V[0].X = 99
	third := transformed.Transformed()
	if third.V[0].X == 99 {
		t.Fatal("Transformed should return clone-safe path copies")
	}

	source.Invalidate(InvalidAffine)
	after := transformed.Transformed()
	if builds != 2 {
		t.Fatalf("transform builds = %d, want 2 after invalidation", builds)
	}
	if after.V[0].X != 3 || after.V[0].Y != 2 {
		t.Fatalf("transformed path after invalidation = %+v", after.V)
	}
}

func TestTransformedPathSetPathAndTransformInvalidate(t *testing.T) {
	var path geom.Path
	path.MoveTo(geom.Pt{X: 1, Y: 1})

	tp := NewTransformedPath(path, NewOffset(nil, geom.Pt{X: 1, Y: 1}))
	initial := tp.Transformed()
	if initial.V[0] != (geom.Pt{X: 2, Y: 2}) {
		t.Fatalf("initial transformed path = %+v", initial.V)
	}

	var next geom.Path
	next.MoveTo(geom.Pt{X: 2, Y: 3})
	tp.SetPath(next)
	tp.SetTransform(NewOffset(nil, geom.Pt{X: -1, Y: 4}))
	updated := tp.Transformed()
	if updated.V[0] != (geom.Pt{X: 1, Y: 7}) {
		t.Fatalf("updated transformed path = %+v", updated.V)
	}
}

// logSeparable builds a separable transform with a log-scaled x axis and a
// linear y axis, which is genuinely non-affine.
func logSeparable() SeparableT {
	return NewSeparable(ScaleAxis{Scale: NewLog(1, 1000, 10)}, LinearAxis{Scale: 2, Offset: 3})
}

func TestSplitAffine(t *testing.T) {
	samples := []geom.Pt{{X: 1, Y: 2}, {X: 10, Y: -3}, {X: 100, Y: 0.5}}

	affine := NewAffine(geom.Affine{A: 2, D: 3, E: 5, F: 7})
	linearSep := NewSeparable(LinearAxis{Scale: 4, Offset: 1}, LinearAxis{Scale: -2, Offset: 6})
	bbox := NewBbox(geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 110, Y: 220}})
	bboxAffine := NewBboxTransformTo(bbox)
	provider := affineProviderStub{m: geom.Affine{A: 3, D: 5, E: 1, F: 2}, affine: true}

	cases := []struct {
		name       string
		t          T
		wantFull   bool // expect fullyAffine (nil non-affine remainder)
		wantPoints bool // expect the non-affine remainder to be a no-op (nil) too
	}{
		{name: "nil", t: nil, wantFull: true, wantPoints: true},
		{name: "affine", t: affine, wantFull: true, wantPoints: true},
		{name: "linear separable", t: linearSep, wantFull: true, wantPoints: true},
		{name: "log separable", t: logSeparable(), wantFull: false, wantPoints: false},
		{name: "offset over affine", t: NewOffset(affine, geom.Pt{X: 3, Y: 4}), wantFull: true, wantPoints: true},
		{name: "offset over log", t: NewOffset(logSeparable(), geom.Pt{X: 3, Y: 4}), wantFull: false, wantPoints: false},
		{name: "chain affine,affine", t: Chain{A: linearSep, B: affine}, wantFull: true, wantPoints: true},
		{name: "chain log,bbox-affine", t: Chain{A: logSeparable(), B: bboxAffine}, wantFull: false, wantPoints: false},
		{name: "chain affine,log", t: Chain{A: affine, B: logSeparable()}, wantFull: false, wantPoints: false},
		{name: "provider affine", t: provider, wantFull: true, wantPoints: true},
		{name: "provider non-affine", t: affineProviderStub{affine: false}, wantFull: false, wantPoints: false},
		{name: "opaque", t: opaqueT{}, wantFull: false, wantPoints: false},
		{name: "axes2d linear", t: NewAxes2D(NewLinear(0, 1), NewLinear(0, 1), affine), wantFull: false, wantPoints: false},
		{name: "cached wrapper", t: NewCachedTransform(func() T { return linearSep }), wantFull: true, wantPoints: true},
		{name: "bbox transform to", t: bboxAffine, wantFull: true, wantPoints: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nonAffine, trailing, fullyAffine := splitAffine(tc.t)
			if fullyAffine != tc.wantFull {
				t.Fatalf("fullyAffine = %v, want %v", fullyAffine, tc.wantFull)
			}
			if (nonAffine == nil) != tc.wantPoints {
				t.Fatalf("nonAffine nil = %v, want %v", nonAffine == nil, tc.wantPoints)
			}
			if fullyAffine && nonAffine != nil {
				t.Fatalf("fullyAffine must imply nil non-affine remainder")
			}
			// Reconstruction identity: trailing.Apply(nonAffine.Apply(p)) == t.Apply(p).
			for _, p := range samples {
				want := p
				if tc.t != nil {
					want = tc.t.Apply(p)
				}
				mid := p
				if nonAffine != nil {
					mid = nonAffine.Apply(p)
				}
				got := trailing.Apply(mid)
				if !approxPt(got, want, 1e-9) {
					t.Fatalf("reconstruction mismatch for %v: got %+v want %+v", p, got, want)
				}
			}
		})
	}
}

// countingT is a non-affine transform (it deliberately does not implement
// AffineProvider) that records how many times Apply was invoked, so a test can
// assert the non-affine vertex pass is reused across an affine-only change.
type countingT struct{ applies *int }

func (c countingT) Apply(p geom.Pt) geom.Pt {
	*c.applies++
	return geom.Pt{X: p.X * p.X, Y: p.Y} // nonlinear so it is genuinely non-affine
}

func (c countingT) Invert(p geom.Pt) (geom.Pt, bool) { return p, true }

func TestTransformedPathAffineNonAffineSplit(t *testing.T) {
	applies := 0
	bbox := NewBbox(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 10, Y: 10}})
	bboxAffine := NewBboxTransformTo(bbox)

	var path geom.Path
	path.MoveTo(geom.Pt{X: 2, Y: 1})
	path.LineTo(geom.Pt{X: 3, Y: 4})

	// Chain a non-affine counter with a live bbox affine; wire the path to the
	// bbox node so Bbox.Set fires InvalidAffine through to the cache.
	tr := Chain{A: countingT{applies: &applies}, B: bboxAffine}
	tp := NewTransformedPath(path, tr, bbox.Node())

	pts, aff := tp.TransformedPointsAndAffine()
	if applies != len(path.V) {
		t.Fatalf("non-affine applies = %d, want %d on first build", applies, len(path.V))
	}
	// Non-affine points are x*x, y (the counter), before the bbox affine.
	if pts.V[0] != (geom.Pt{X: 4, Y: 1}) {
		t.Fatalf("non-affine points = %+v", pts.V)
	}
	first := tp.Transformed()

	// Resize the axes: affine-only invalidation. The non-affine pass must NOT
	// re-run, but the trailing affine and full path must reflect the new bbox.
	bbox.Set(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 20, Y: 20}})
	pts2, aff2 := tp.TransformedPointsAndAffine()
	if applies != len(path.V) {
		t.Fatalf("non-affine applies = %d after affine-only resize, want %d (vertex pass must be reused)", applies, len(path.V))
	}
	if aff2 == aff {
		t.Fatalf("trailing affine should change on resize: %+v", aff2)
	}
	if pts2.V[0] != (geom.Pt{X: 4, Y: 1}) {
		t.Fatalf("non-affine points changed on affine-only resize: %+v", pts2.V)
	}
	resized := tp.Transformed()
	if resized.V[0] == first.V[0] {
		t.Fatalf("full transformed path should change on resize: %+v", resized.V)
	}
	// The full path must equal the freshly composed transform.
	if want := tr.Apply(path.V[0]); !approxPt(resized.V[0], want, 1e-9) {
		t.Fatalf("resized full path = %+v, want %+v", resized.V[0], want)
	}

	// A non-affine invalidation must re-run the vertex pass.
	before := applies
	tp.Invalidate(InvalidNonAffine)
	tp.TransformedPointsAndAffine()
	if applies-before != len(path.V) {
		t.Fatalf("non-affine applies delta = %d after InvalidNonAffine, want %d", applies-before, len(path.V))
	}
}

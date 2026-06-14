package transform

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestTransformNodeInvalidationPropagatesToDependents(t *testing.T) {
	var source TransformNode
	var dependent TransformNode
	source.AddDependent(&dependent)

	source.Invalidate(InvalidNonAffine)

	if !source.Invalid().Has(InvalidNonAffine) {
		t.Fatalf("source invalidation = %v, want non-affine", source.Invalid())
	}
	if !dependent.Invalid().Has(InvalidNonAffine) {
		t.Fatalf("dependent invalidation = %v, want non-affine", dependent.Invalid())
	}
	if source.Version() == 0 || dependent.Version() == 0 {
		t.Fatalf("versions not advanced: source=%d dependent=%d", source.Version(), dependent.Version())
	}
}

func TestCachedTransformRebuildsOnlyWhenInvalidated(t *testing.T) {
	var source TransformNode
	builds := 0
	cached := NewCachedTransform(func() T {
		builds++
		return NewOffset(nil, geom.Pt{X: float64(builds), Y: 0})
	}, &source)

	first := cached.Apply(geom.Pt{X: 10, Y: 5})
	second := cached.Apply(geom.Pt{X: 10, Y: 5})
	if first != second {
		t.Fatalf("cached transform changed without invalidation: first=%+v second=%+v", first, second)
	}
	if builds != 1 {
		t.Fatalf("builds = %d, want 1", builds)
	}

	source.Invalidate(InvalidAffine)
	third := cached.Apply(geom.Pt{X: 10, Y: 5})
	if third.X != 12 || third.Y != 5 {
		t.Fatalf("rebuilt transform output = %+v, want {12 5}", third)
	}
	if builds != 2 {
		t.Fatalf("builds = %d, want 2", builds)
	}
	if cached.Invalid() != InvalidNone {
		t.Fatalf("cached invalidation = %v, want cleared after rebuild", cached.Invalid())
	}
}

package transform

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
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

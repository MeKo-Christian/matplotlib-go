package transform

import "github.com/cwbudde/matplotlib-go/internal/geom"

// TransformedPath caches a path transformed through a transform graph.
//
// It mirrors the useful static-rendering part of Matplotlib's TransformedPath:
// callers get clone-safe path access, and cache invalidation can be wired to
// transform nodes without making geom.Path itself mutable or graph-aware.
//
// The current Go transform model applies every transform through a single T
// interface. InvalidAffine and InvalidNonAffine are preserved as dependency
// invalidation hints, but TransformedPath intentionally keeps one full-path
// cache until a visible renderer path needs Matplotlib's affine/non-affine
// split cache.
type TransformedPath struct {
	TransformNode
	path          geom.Path
	transform     T
	cached        geom.Path
	cachedValid   bool
	cachedVersion uint64
}

// NewTransformedPath creates a transformed path cache.
func NewTransformedPath(path geom.Path, transform T, dependencies ...*TransformNode) *TransformedPath {
	tp := &TransformedPath{
		path:      path.Clone(),
		transform: transform,
	}
	for _, dependency := range dependencies {
		if dependency != nil {
			dependency.AddDependent(&tp.TransformNode)
		}
	}
	tp.Invalidate(InvalidAll)
	return tp
}

// Path returns a clone of the original untransformed path.
func (tp *TransformedPath) Path() geom.Path {
	if tp == nil {
		return geom.Path{}
	}
	return tp.path.Clone()
}

// Transform returns the transform used by this cache.
func (tp *TransformedPath) Transform() T {
	if tp == nil {
		return nil
	}
	return tp.transform
}

// SetPath replaces the source path and invalidates the cached transformed copy.
func (tp *TransformedPath) SetPath(path geom.Path) {
	if tp == nil {
		return
	}
	tp.path = path.Clone()
	tp.Invalidate(InvalidAll)
}

// SetTransform replaces the transform and invalidates the cache.
func (tp *TransformedPath) SetTransform(transform T) {
	if tp == nil {
		return
	}
	tp.transform = transform
	tp.Invalidate(InvalidAll)
}

// Transformed returns a clone of the fully transformed path.
func (tp *TransformedPath) Transformed() geom.Path {
	if tp == nil {
		return geom.Path{}
	}
	if !tp.cachedValid || tp.Invalid() != InvalidNone {
		tp.cached = transformPath(tp.path, tp.transform)
		tp.cachedValid = true
		tp.cachedVersion = tp.Version()
		tp.ClearInvalid()
	}
	return tp.cached.Clone()
}

// CachedVersion reports the invalidation version used for the cached path.
func (tp *TransformedPath) CachedVersion() uint64 {
	if tp == nil {
		return 0
	}
	return tp.cachedVersion
}

func transformPath(path geom.Path, transform T) geom.Path {
	out := path.Clone()
	if transform == nil {
		return out
	}
	for i, v := range out.V {
		out.V[i] = transform.Apply(v)
	}
	return out
}

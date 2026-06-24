package transform

import "github.com/cwbudde/matplotlib-go/geom"

// TransformedPath caches a path transformed through a transform graph.
//
// It mirrors Matplotlib's TransformedPath, including the affine/non-affine cache
// split: the expensive non-affine remainder of the transform is applied to the
// vertices once and cached, while the cheap trailing affine is re-fetched and
// re-applied each time the path is requested. An InvalidAffine invalidation
// (e.g. an axes resize/pan/zoom via Bbox.Set) refreshes only the trailing affine
// and reuses the cached non-affine points; an InvalidNonAffine/InvalidAll
// invalidation re-runs the full vertex pass.
//
// Cache invalidation is wired to transform nodes without making geom.Path itself
// mutable or graph-aware, and all accessors return clones for caller safety.
type TransformedPath struct {
	TransformNode
	path      geom.Path
	transform T

	// nonAffine holds the source vertices after the non-affine remainder of the
	// transform has been applied. It is recomputed only on an InvalidNonAffine
	// (or first build); an affine-only change leaves it untouched.
	nonAffine      geom.Path
	nonAffineValid bool

	// affine is the maximal trailing affine of the transform, refreshed on any
	// affine or non-affine invalidation (cheap: it walks the transform structure,
	// not the vertices).
	affine geom.Affine

	// cached is the fully transformed path (affine applied over nonAffine),
	// derived lazily from the two sub-caches.
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
	tp.revalidate()
	if !tp.cachedValid {
		tp.cached = applyAffine(tp.nonAffine, tp.affine)
		tp.cachedValid = true
	}
	return tp.cached.Clone()
}

// TransformedPointsAndAffine returns a clone of the path with the non-affine part
// of the transform already applied, along with the trailing affine needed to
// finish the transformation. It mirrors Matplotlib's
// get_transformed_points_and_affine: callers that can apply the affine cheaply
// (or hand it to a backend) avoid re-running the non-affine projection when only
// the affine changed.
func (tp *TransformedPath) TransformedPointsAndAffine() (geom.Path, geom.Affine) {
	if tp == nil {
		return geom.Path{}, geom.Identity()
	}
	tp.revalidate()
	return tp.nonAffine.Clone(), tp.affine
}

// Affine returns the current trailing affine of the transform.
func (tp *TransformedPath) Affine() geom.Affine {
	if tp == nil {
		return geom.Identity()
	}
	tp.revalidate()
	return tp.affine
}

// revalidate refreshes the non-affine and affine sub-caches according to the
// pending invalidation, mirroring Matplotlib's TransformedPath._revalidate: the
// vertex pass re-runs only on a non-affine invalidation, while an affine-only
// invalidation just refreshes the trailing affine.
func (tp *TransformedPath) revalidate() {
	inv := tp.Invalid()
	switch {
	case !tp.nonAffineValid || inv.Has(InvalidNonAffine):
		nonAffine, affine, _ := splitAffine(tp.transform)
		tp.nonAffine = applyNonAffine(tp.path, nonAffine)
		tp.affine = affine
		tp.nonAffineValid = true
		tp.cachedValid = false
	case inv.Has(InvalidAffine):
		_, affine, _ := splitAffine(tp.transform)
		tp.affine = affine
		tp.cachedValid = false
	}
	tp.cachedVersion = tp.Version()
	tp.ClearInvalid()
}

// CachedVersion reports the invalidation version used for the cached path.
func (tp *TransformedPath) CachedVersion() uint64 {
	if tp == nil {
		return 0
	}
	return tp.cachedVersion
}

// applyNonAffine applies the non-affine remainder of a split transform to every
// vertex. A nil remainder (a fully affine transform) leaves the points unchanged.
func applyNonAffine(path geom.Path, nonAffine T) geom.Path {
	out := path.Clone()
	if nonAffine == nil {
		return out
	}
	for i, v := range out.V {
		out.V[i] = nonAffine.Apply(v)
	}
	return out
}

// applyAffine applies a trailing affine to every vertex of an already
// non-affine-transformed path.
func applyAffine(path geom.Path, affine geom.Affine) geom.Path {
	out := path.Clone()
	for i, v := range out.V {
		out.V[i] = affine.Apply(v)
	}
	return out
}

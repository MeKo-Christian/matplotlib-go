package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

// displayPathCache holds the persistent transform.TransformedPath for one artist
// (sub-)path so a redraw that changes only the trailing affine (axes
// resize/pan/zoom) reuses the cached non-affine projection instead of
// re-projecting every vertex. It is the patch/collection analogue of the Line2D
// cache.
//
// Unlike Line2D — whose source is the stable l.XY backing — patches and
// collections rebuild their source path fresh each draw (e.g. rectanglePath,
// polygonPath, per-element scaleAndTranslatePath), so change detection is
// value-based (pathsEqualValue) rather than slice-header identity. The cache only
// engages for genuine non-affine legs; the caller applies that parity gate (see
// DrawContext.dataLegIsNonAffine) before calling build.
type displayPathCache struct {
	tp  *transform.TransformedPath
	src geom.Path   // value snapshot of the last projected source path
	tr  transform.T // identity of the transform the projection was built for
}

// build returns the display path for src under tr, reusing the cached non-affine
// projection when only the trailing affine moved. It reports ok=false when the
// persistent axes invalidation graph is unavailable, in which case the caller
// falls back to the direct per-vertex transform.
//
// The fully-affine parity gate is applied by the caller, so build assumes tr has
// a non-affine remainder: TransformedPath.Transformed() then equals the direct
// applyTransformPath byte-for-byte (the trailing affine is the single bbox matrix
// and the composition is exact).
func (c *displayPathCache) build(ctx *DrawContext, src geom.Path, tr transform.T) (geom.Path, bool) {
	if c == nil || ctx == nil || tr == nil {
		return geom.Path{}, false
	}
	deps := ctx.dataTransformDeps()
	if len(deps) == 0 {
		return geom.Path{}, false
	}

	switch {
	case c.tp == nil || c.tr != tr:
		c.tp = transform.NewTransformedPath(src, tr, deps...)
		c.tr = tr
		c.src = src.Clone()
	case !pathsEqualValue(c.src, src):
		c.tp.SetPath(src)
		c.src = src.Clone()
	}

	// Apply the trailing affine ourselves rather than via Transformed(): for a
	// shear-free affine (the separable axes bbox) we apply it per axis so a vertex
	// outside the data domain (e.g. NaN under a log scale) keeps NaN local to one
	// axis, matching the direct separable chain. A full-matrix apply would let the
	// zero cross-term contaminate the finite axis (0*NaN == NaN).
	nonAffine, affine := c.tp.TransformedPointsAndAffine()
	return applyTrailingAffinePath(nonAffine, affine), true
}

// applyTrailingAffinePath applies the trailing affine of a split transform to a
// path's vertices. A shear-free affine (B==C==0, i.e. the separable axes->pixel
// bbox) is applied per axis to stay byte-identical to the direct separable
// transform, including NaN locality; any other affine uses the full matrix.
func applyTrailingAffinePath(path geom.Path, aff geom.Affine) geom.Path {
	out := path.Clone()
	if aff.B == 0 && aff.C == 0 {
		for i, v := range out.V {
			out.V[i] = geom.Pt{X: aff.A*v.X + aff.E, Y: aff.D*v.Y + aff.F}
		}
		return out
	}
	for i, v := range out.V {
		out.V[i] = aff.Apply(v)
	}
	return out
}

// displayPathCacher is implemented by single-path artists that own one persistent
// display-path cache (every type embedding Patch). Collections do not implement
// it — they keep one cache per element and call buildCachedDisplayPath directly.
type displayPathCacher interface {
	displayPathCacheSlot() *displayPathCache
}

// pathsEqualValue reports whether two paths are element-wise identical. It is a
// value comparison (sources are rebuilt each draw), so a path containing NaN
// vertices always compares unequal and conservatively re-projects.
func pathsEqualValue(a, b geom.Path) bool {
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

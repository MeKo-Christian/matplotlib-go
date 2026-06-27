package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

// ensureTransforms lazily builds the persistent transform graph for the axes.
//
// transAxes is a live transform.BboxTransformTo over axesBbox (the axes pixel
// rectangle). transData composes the current data->axes leg with transAxes and
// is invalidated when either axesBbox or the data leg changes, so it is reused
// across draws instead of rebuilt when nothing relevant moved.
func (a *Axes) ensureTransforms() {
	if a == nil || a.transData != nil {
		return
	}
	a.axesBbox = transform.NewBbox(geom.Rect{})
	a.transAxes = transform.NewBboxTransformTo(a.axesBbox)
	a.transData = transform.NewCachedTransform(func() transform.T {
		dataToAxes := a.curDataToAxes
		if dataToAxes == nil {
			return a.transAxes
		}
		return transform.Chain{A: dataToAxes, B: a.transAxes}
	}, a.axesBbox.Node(), &a.dataNode)
}

// updateAxesBbox points the persistent transform graph at the current axes pixel
// rectangle, invalidating downstream transforms only when the rectangle changed.
func (a *Axes) updateAxesBbox(clip geom.Rect) {
	a.ensureTransforms()
	a.axesBbox.Set(clip)
}

// refreshDataTransform records the current data->axes transform and invalidates
// the cached transData only when that leg actually changed.
//
// The invalidation stage matches the kind of leg, so a split-aware consumer
// (transform.TransformedPath) refreshes the right sub-cache:
//
//   - Affine data leg (linear rectilinear axes): the change is detected by
//     comparing the flattened affine, so an unchanged graph reuses the cache and
//     only an actual change fires transform.InvalidAffine (refresh the trailing
//     affine, keep the cached non-affine vertex pass).
//   - Non-affine leg (log/symlog scales, polar/geo/3D projections with mutable
//     state): the leg is treated as changed every draw and fires
//     transform.InvalidNonAffine so the vertex pass is re-run. Firing only
//     InvalidAffine here would let a consumer reuse a stale projection on a
//     log-domain/limits change.
//
// A non-affine leg is still re-projected every draw (no resize win yet); a cheap
// non-affine-leg change check is a deferred follow-up.
func (a *Axes) refreshDataTransform(dataToAxes transform.T) {
	a.ensureTransforms()
	a.curDataToAxes = dataToAxes

	aff, ok := transform.AsAffine(dataToAxes)
	if !ok {
		// Non-affine leg: re-run the vertex pass. InvalidNonAffine also refreshes
		// the trailing affine in TransformedPath.revalidate.
		a.dataAffine = aff
		a.dataAffineOK = ok
		a.dataSnapSet = true
		a.dataNode.Invalidate(transform.InvalidNonAffine)
		return
	}
	if !a.dataSnapSet || !a.dataAffineOK || aff != a.dataAffine {
		a.dataAffine = aff
		a.dataAffineOK = ok
		a.dataSnapSet = true
		a.dataNode.Invalidate(transform.InvalidAffine)
	}
}

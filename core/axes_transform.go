package core

import (
	"reflect"

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
//   - Non-affine leg (log/symlog scales, polar/geo/skew projections): the change
//     is detected by comparing the leg structurally against the previous draw's
//     leg, so an unchanged leg reuses the cached non-affine vertex pass and only
//     an actual change fires transform.InvalidNonAffine. A resize that only moves
//     the axes bbox therefore refreshes the trailing affine alone (via the
//     separate axesBbox node), reusing the projection. Firing InvalidNonAffine
//     unconditionally would re-project every draw; firing only InvalidAffine
//     would let a consumer reuse a stale projection on a log-domain/limits change.
func (a *Axes) refreshDataTransform(dataToAxes transform.T) {
	a.ensureTransforms()
	prevLeg := a.curDataToAxes
	a.curDataToAxes = dataToAxes

	aff, ok := transform.AsAffine(dataToAxes)
	if !ok {
		// Non-affine leg. Re-run the vertex pass only when the leg actually
		// changed; InvalidNonAffine also refreshes the trailing affine in
		// TransformedPath.revalidate. Legs are rebuilt fresh each draw, so this is
		// a structural (value) comparison rather than pointer identity.
		changed := !a.dataSnapSet || a.dataAffineOK || !legsEqual(prevLeg, dataToAxes)
		a.dataAffine = aff
		a.dataAffineOK = ok
		a.dataSnapSet = true
		if changed {
			a.dataNode.Invalidate(transform.InvalidNonAffine)
		}
		return
	}
	if !a.dataSnapSet || !a.dataAffineOK || aff != a.dataAffine {
		a.dataAffine = aff
		a.dataAffineOK = ok
		a.dataSnapSet = true
		a.dataNode.Invalidate(transform.InvalidAffine)
	}
}

// legsEqual reports whether two data->axes legs are structurally identical, so an
// unchanged non-affine leg can reuse its cached projection across draws.
//
// Legs are value transforms rebuilt fresh each draw (e.g. a SeparableT of log
// scales, a polar/skew data transform), so this is a deep value comparison rather
// than pointer identity. reflect.DeepEqual is used because some legs embed
// non-comparable scales (FuncScale holds funcs) that would panic under ==; those
// compare unequal here, conservatively forcing a re-projection.
func legsEqual(a, b transform.T) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return reflect.DeepEqual(a, b)
}

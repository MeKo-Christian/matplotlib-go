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
// For an affine data leg (linear rectilinear axes) the change is detected by
// comparing the flattened affine, so an unchanged graph reuses the cache. For a
// non-affine leg (log/symlog scales, polar/geo/3D projections with mutable
// state) the leg is treated as changed every draw, preserving the previous
// rebuild-every-draw behavior for those cases.
func (a *Axes) refreshDataTransform(dataToAxes transform.T) {
	a.ensureTransforms()
	a.curDataToAxes = dataToAxes

	aff, ok := transform.AsAffine(dataToAxes)
	if !ok || !a.dataSnapSet || !a.dataAffineOK || aff != a.dataAffine {
		a.dataAffine = aff
		a.dataAffineOK = ok
		a.dataSnapSet = true
		a.dataNode.Invalidate(transform.InvalidAffine)
	}
}

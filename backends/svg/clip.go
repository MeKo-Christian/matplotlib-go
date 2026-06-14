package svg

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends/internal/mixedraster"
	"github.com/cwbudde/matplotlib-go/geom"
)

func (r *Renderer) ClipRect(rect geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipRect(rect)
		return
	}
	clipRect := normalizeRect(rect)
	if r.clipRect == nil {
		r.clipRect = &clipRect
		return
	}

	intersected := r.clipRect.Intersect(clipRect)
	r.clipRect = &intersected
}

func (r *Renderer) ClipPath(p geom.Path) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipPath(p)
		return
	}
	r.clipPath(affinePath(p, r.deviceFlip()), "", p)
}

func (r *Renderer) ClipPathTransformed(p geom.Path, transform geom.Affine) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipPath(mixedraster.ApplyAffine(p, transform))
		return
	}
	// Keep the path in its untransformed form and carry the affine on the
	// <clipPath> def, composing the device y-flip so the clip region lands in
	// SVG device space: clip = (flip ∘ transform)(p).
	r.clipPath(p, matrixTransform(r.deviceFlip().Mul(transform)), mixedraster.ApplyAffine(p, transform))
}

func (r *Renderer) clipPath(p geom.Path, transform string, rasterPath geom.Path) {
	if !p.Validate() {
		return
	}
	d := buildPathData(p)
	if d == "" {
		return
	}
	id := r.registerPathClip(d, transform)
	r.clipPathStack = append(r.clipPathStack, id)
	r.clipPaths = append(r.clipPaths, mixedraster.ClonePath(rasterPath))
}

func (r *Renderer) currentClipIDs() []string {
	count := len(r.clipPathStack)
	if r.clipRect != nil {
		count++
	}
	if count == 0 {
		return nil
	}

	ids := make([]string, 0, count)
	if r.clipRect != nil {
		ids = append(ids, r.registerClip(*r.clipRect))
	}
	ids = append(ids, r.clipPathStack...)
	return ids
}

func (r *Renderer) registerClip(rect geom.Rect) string {
	key := clipKey(rect)
	if id, ok := r.clipDefs[key]; ok {
		return id
	}

	r.clipIDCounter++
	id := r.defID("clip", key, r.clipIDCounter)
	r.clipDefs[key] = id
	rectCopy := rect
	r.clipOrder = append(r.clipOrder, clipDef{id: id, rect: &rectCopy})
	return id
}

func (r *Renderer) registerPathClip(d, transform string) string {
	key := d
	if transform != "" {
		key += "\x00" + transform
	}
	if id, ok := r.clipPathDefs[key]; ok {
		return id
	}

	r.clipIDCounter++
	id := r.defID("clip", key, r.clipIDCounter)
	r.clipPathDefs[key] = id
	r.clipOrder = append(r.clipOrder, clipDef{id: id, path: d, transform: transform})
	return id
}

func clipKey(rect geom.Rect) string {
	q := normalizeRect(rect)
	return fmt.Sprintf(
		"%s,%s,%s,%s",
		formatFloat(q.Min.X),
		formatFloat(q.Min.Y),
		formatFloat(q.Max.X),
		formatFloat(q.Max.Y),
	)
}

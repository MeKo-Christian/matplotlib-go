package transform

import "github.com/cwbudde/matplotlib-go/geom"

// Bbox is a mutable bounding box that participates in the transform
// invalidation graph.
//
// It mirrors matplotlib's mutable Bbox: bbox-linked transforms register as
// dependents, and mutating the rectangle (Set) invalidates those transforms so
// their cached affine recomputes lazily on next use rather than being rebuilt
// from scratch every draw. The canonical store is a geom.Rect (comparable), so
// the invalidate-on-change gate is a plain inequality.
type Bbox struct {
	TransformNode
	rect geom.Rect
}

// NewBbox creates a mutable bbox node with the given rectangle.
func NewBbox(rect geom.Rect) *Bbox {
	return &Bbox{rect: rect}
}

// Rect returns the current rectangle.
func (b *Bbox) Rect() geom.Rect {
	if b == nil {
		return geom.Rect{}
	}
	return b.rect
}

// Set replaces the rectangle, invalidating dependents only when it changes.
func (b *Bbox) Set(rect geom.Rect) {
	if b == nil || b.rect == rect {
		return
	}
	b.rect = rect
	b.Invalidate(InvalidAffine)
}

// Node returns the underlying invalidation node so callers can wire additional
// dependents (e.g. a composite transData cache) to bbox changes.
func (b *Bbox) Node() *TransformNode {
	if b == nil {
		return nil
	}
	return &b.TransformNode
}

// bboxLinked is the shared implementation for bbox-linked separable transforms.
// It lazily rebuilds a cached SeparableT whenever its node is invalidated (or on
// first use), mirroring CachedTransform.Current but specialized to the separable
// result so downstream Separable fast-paths keep working unchanged.
type bboxLinked struct {
	TransformNode
	build  func() SeparableT
	cached SeparableT
	built  bool
}

func (t *bboxLinked) current() SeparableT {
	if !t.built || t.Invalid() != InvalidNone {
		t.cached = t.build()
		t.built = true
		t.ClearInvalid()
	}
	return t.cached
}

func (t *bboxLinked) Apply(p geom.Pt) geom.Pt { return t.current().Apply(p) }

func (t *bboxLinked) Invert(p geom.Pt) (geom.Pt, bool) { return t.current().Invert(p) }

func (t *bboxLinked) XAxis() AxisTransform { return t.current().XAxis() }

func (t *bboxLinked) YAxis() AxisTransform { return t.current().YAxis() }

// BboxTransformTo maps the unit square [0,1]x[0,1] onto a (live) bbox.
//
// It is the live counterpart of NewUnitRectTransform/NewDisplayRectTransform and
// matches matplotlib's BboxTransformTo: point (u,v) maps to
// (minX + u*width, minY + v*height). When the bbox is resized via Bbox.Set the
// cached transform is invalidated and recomputed on next use.
type BboxTransformTo struct{ bboxLinked }

// NewBboxTransformTo builds a transform from the unit square to box.
func NewBboxTransformTo(box *Bbox) *BboxTransformTo {
	t := &BboxTransformTo{}
	t.build = func() SeparableT { return NewUnitRectTransform(box.Rect()) }
	box.AddDependent(&t.TransformNode)
	return t
}

// BboxTransformFrom maps a (live) bbox onto the unit square [0,1]x[0,1].
//
// It is the inverse mapping of BboxTransformTo and matches matplotlib's
// BboxTransformFrom.
type BboxTransformFrom struct{ bboxLinked }

// NewBboxTransformFrom builds a transform from box to the unit square.
func NewBboxTransformFrom(box *Bbox) *BboxTransformFrom {
	t := &BboxTransformFrom{}
	t.build = func() SeparableT {
		return NewRectTransform(box.Rect(), unitRect())
	}
	box.AddDependent(&t.TransformNode)
	return t
}

// BboxTransform maps one (live) bbox onto another, tracking both for changes.
//
// It matches matplotlib's BboxTransform(boxin, boxout).
type BboxTransform struct{ bboxLinked }

// NewBboxTransform builds a transform mapping the in bbox onto the out bbox.
func NewBboxTransform(in, out *Bbox) *BboxTransform {
	t := &BboxTransform{}
	t.build = func() SeparableT {
		return NewRectTransform(in.Rect(), out.Rect())
	}
	in.AddDependent(&t.TransformNode)
	out.AddDependent(&t.TransformNode)
	return t
}

func unitRect() geom.Rect {
	return geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}}
}

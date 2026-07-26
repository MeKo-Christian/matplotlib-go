package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

type DrawContext struct {
	// DataToPixel maps data coordinates to pixels.
	DataToPixel Transform2D
	// Axes is the owning axes for axes-local drawing.
	Axes *Axes
	// Projection configures data->axes mapping and non-Cartesian behavior.
	Projection Projection
	// Styling configuration in effect.
	RC style.RC
	// Clip is the axes pixel rectangle.
	Clip geom.Rect
	// FigureRect is the figure display rectangle in pixels.
	FigureRect geom.Rect
	// DrawOptions selects which artists participate in this draw pass. The
	// zero value matches Matplotlib's default Artist.draw_wrapper behavior:
	// animated artists are skipped, and the animation engine flips this
	// during overlay/background passes.
	DrawOptions DrawOptions
}

type AnimatedFilter uint8

const (
	// AnimatedFilterExcludeAnimated draws non-animated artists only. This is
	// the default and matches Matplotlib's per-frame figure draw.
	AnimatedFilterExcludeAnimated AnimatedFilter = iota
	// AnimatedFilterOnlyAnimated draws animated artists only. Used by the
	// animation engine to redraw the animated overlay on top of a saved
	// background region.
	AnimatedFilterOnlyAnimated
	// AnimatedFilterAll draws every artist regardless of animated state.
	// Useful for one-shot saves where blit semantics do not apply.
	AnimatedFilterAll
)

type DrawOptions struct {
	AnimatedFilter AnimatedFilter

	// Transparent forces axes patch backgrounds to be transparent at save time
	// (savefig.transparent). The figure background is handled by the save path.
	Transparent bool
	// FigureBackground, when non-nil, paints a full-viewport background fill
	// before any content (used for savefig.facecolor on backends that cannot
	// re-clear their surface).
	FigureBackground optional.Value[render.Color]
	// FigureEdge, when non-nil, strokes a border around the full viewport after
	// the background fill (savefig.edgecolor).
	FigureEdge optional.Value[render.Color]
	// FigureEdgeWidth is the figure border stroke width in device pixels.
	FigureEdgeWidth float64
}

type Transform2D struct {
	XScale      transform.Scale
	YScale      transform.Scale
	DataToAxes  transform.T
	AxesToPixel transform.T
	// composed is the persistent, cache-backed data->pixel transform (Phase 11
	// live bbox-linked transforms). When non-nil it is the authoritative
	// composition used by Apply/Invert/TransData; it is numerically identical to
	// chaining DataToAxes with AxesToPixel, which the separable-decomposition
	// fast-paths continue to rebuild from the component fields.
	composed transform.T
}

func (t *Transform2D) Apply(p geom.Pt) geom.Pt {
	tr := t.transData()
	if tr == nil {
		return p
	}
	return tr.Apply(p)
}

func (t *Transform2D) Invert(p geom.Pt) (geom.Pt, bool) {
	tr := t.transData()
	if tr == nil {
		return p, true
	}
	return tr.Invert(p)
}

func (t *Transform2D) transData() transform.T {
	if t == nil {
		return nil
	}

	if t.composed != nil {
		return t.composed
	}

	dataToAxes := t.DataToAxes
	if dataToAxes == nil {
		dataToAxes = transform.NewScaleTransform(t.XScale, t.YScale)
	}

	switch {
	case dataToAxes == nil:
		return t.AxesToPixel
	case t.AxesToPixel == nil:
		return dataToAxes
	default:
		return transform.Chain{A: dataToAxes, B: t.AxesToPixel}
	}
}

type CoordinateSpace uint8

const (
	CoordData CoordinateSpace = iota
	CoordAxes
	CoordFigure
)

type CoordinateSpec struct {
	X CoordinateSpace
	Y CoordinateSpace
}

func Coords(space CoordinateSpace) CoordinateSpec {
	return CoordinateSpec{X: space, Y: space}
}

func BlendCoords(xSpace, ySpace CoordinateSpace) CoordinateSpec {
	return CoordinateSpec{X: xSpace, Y: ySpace}
}

func (ctx *DrawContext) TransData() transform.T {
	if ctx == nil {
		return nil
	}
	return ctx.DataToPixel.transData()
}

// dataTransformDeps returns the invalidation nodes a data-coords
// transform.TransformedPath must depend on so it refreshes when the axes graph
// moves: the axes pixel bbox (the affine leg, via Bbox.Set on resize/pan/zoom)
// and the data leg node (refreshed by refreshDataTransform). It returns nil when
// the persistent axes graph is not initialized, in which case callers fall back
// to the uncached per-vertex path. Mirrors the wiring in Axes.ensureTransforms.
func (ctx *DrawContext) dataTransformDeps() []*transform.TransformNode {
	if ctx == nil || ctx.Axes == nil || ctx.Axes.transData == nil {
		return nil
	}
	return []*transform.TransformNode{ctx.Axes.axesBbox.Node(), &ctx.Axes.dataNode}
}

// dataLegIsNonAffine reports whether the axes data->axes leg currently has a
// non-affine remainder. It reads the flag refreshDataTransform recorded, so it
// avoids re-running an AsAffine traversal per artist (or per collection element).
//
// The flag describes the data leg specifically, so callers must only trust it
// when the artist's resolved transform IS the data transform (i.e.
// artistUsesDataCoords is true). It is the parity gate for the cached display
// path: a fully-affine leg has no projection to cache and collapsing scale+bbox
// into one matrix would diverge from the direct two-step chain by a last ULP,
// so those keep the direct per-vertex path.
func (ctx *DrawContext) dataLegIsNonAffine() bool {
	if ctx == nil || ctx.Axes == nil || ctx.Axes.transData == nil {
		return false
	}
	return ctx.Axes.dataSnapSet && !ctx.Axes.dataAffineOK
}

func (ctx *DrawContext) TransProjection() transform.T {
	if ctx == nil {
		return nil
	}
	if ctx.DataToPixel.DataToAxes != nil {
		return ctx.DataToPixel.DataToAxes
	}
	return transform.NewScaleTransform(ctx.DataToPixel.XScale, ctx.DataToPixel.YScale)
}

func (ctx *DrawContext) TransAxes() transform.T {
	if ctx == nil {
		return nil
	}
	if ctx.DataToPixel.AxesToPixel != nil {
		return ctx.DataToPixel.AxesToPixel
	}
	if rect, ok := unitSquareBounds(nil, ctx.Clip); ok {
		return transform.NewDisplayRectTransform(rect)
	}
	return nil
}

func (ctx *DrawContext) TransFigure() transform.T {
	if ctx == nil {
		return nil
	}
	rect := ctx.FigureRect
	if rect == (geom.Rect{}) {
		rect = ctx.Clip
	}
	if rect == (geom.Rect{}) {
		return nil
	}
	return transform.NewDisplayRectTransform(rect)
}

func (ctx *DrawContext) TransformFor(spec CoordinateSpec) transform.T {
	if spec.X == spec.Y {
		return ctx.transformForSpace(spec.X)
	}

	xTrans, okX := ctx.separableTransformForSpace(spec.X)
	yTrans, okY := ctx.separableTransformForSpace(spec.Y)
	if !okX || !okY {
		return nil
	}
	return transform.Blend(xTrans, yTrans)
}

// AffineTransformFor returns the coordinate transform as an affine matrix when
// the active transform graph is linear. Nonlinear projections or scales return
// false, matching Matplotlib's split between transformed paths and affine tails.
func (ctx *DrawContext) AffineTransformFor(spec CoordinateSpec) (geom.Affine, bool) {
	return transform.AsAffine(ctx.TransformFor(spec))
}

func (ctx *DrawContext) transformForSpace(space CoordinateSpace) transform.T {
	switch space {
	case CoordAxes:
		return ctx.TransAxes()
	case CoordFigure:
		return ctx.TransFigure()
	default:
		return ctx.TransData()
	}
}

func (ctx *DrawContext) separableTransformForSpace(space CoordinateSpace) (transform.Separable, bool) {
	switch space {
	case CoordAxes:
		return ctx.separableAxesTransform()
	case CoordFigure:
		tr := ctx.TransFigure()
		sep, ok := tr.(transform.Separable)
		return sep, ok
	default:
		return ctx.separableDataTransform()
	}
}

func (ctx *DrawContext) separableAxesTransform() (transform.Separable, bool) {
	if ctx == nil {
		return transform.SeparableT{}, false
	}
	if sep, ok := ctx.DataToPixel.AxesToPixel.(transform.Separable); ok {
		return sep, true
	}
	rect, ok := unitSquareBounds(ctx.DataToPixel.AxesToPixel, ctx.Clip)
	if !ok {
		return transform.SeparableT{}, false
	}
	return transform.NewDisplayRectTransform(rect), true
}

func (ctx *DrawContext) separableDataTransform() (transform.Separable, bool) {
	axesTrans, ok := ctx.separableAxesTransform()
	if !ok {
		return transform.SeparableT{}, false
	}

	if ctx != nil && ctx.DataToPixel.DataToAxes != nil {
		sep, ok := ctx.DataToPixel.DataToAxes.(transform.Separable)
		if !ok {
			return transform.SeparableT{}, false
		}
		return transform.ChainSeparable(sep, axesTrans), true
	}

	return transform.ChainSeparable(
		transform.NewScaleTransform(ctx.DataToPixel.XScale, ctx.DataToPixel.YScale),
		axesTrans,
	), true
}

func unitSquareBounds(tr transform.T, fallback geom.Rect) (geom.Rect, bool) {
	if tr == nil {
		if fallback == (geom.Rect{}) {
			return geom.Rect{}, false
		}
		return fallback, true
	}

	corners := []geom.Pt{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}

	rect := geom.Rect{Min: tr.Apply(corners[0]), Max: tr.Apply(corners[0])}
	for _, corner := range corners[1:] {
		pt := tr.Apply(corner)
		if pt.X < rect.Min.X {
			rect.Min.X = pt.X
		}
		if pt.Y < rect.Min.Y {
			rect.Min.Y = pt.Y
		}
		if pt.X > rect.Max.X {
			rect.Max.X = pt.X
		}
		if pt.Y > rect.Max.Y {
			rect.Max.Y = pt.Y
		}
	}
	return rect, true
}

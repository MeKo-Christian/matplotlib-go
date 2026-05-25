package transform

import "github.com/cwbudde/matplotlib-go/internal/geom"

// AxisTransform maps a single scalar coordinate and can invert it.
type AxisTransform interface {
	Forward(v float64) float64
	Inverse(v float64) (float64, bool)
}

// Separable is a 2D transform whose x/y components can be reasoned about independently.
type Separable interface {
	T
	XAxis() AxisTransform
	YAxis() AxisTransform
}

type identityAxis struct{}

func (identityAxis) Forward(v float64) float64         { return v }
func (identityAxis) Inverse(v float64) (float64, bool) { return v, true }

// ScaleAxis adapts a Scale into an axis transform.
type ScaleAxis struct{ Scale Scale }

func (a ScaleAxis) Forward(v float64) float64 {
	if a.Scale == nil {
		return v
	}
	return a.Scale.Fwd(v)
}

func (a ScaleAxis) Inverse(v float64) (float64, bool) {
	if a.Scale == nil {
		return v, true
	}
	return a.Scale.Inv(v)
}

// LinearAxis maps a scalar interval linearly into another interval.
type LinearAxis struct {
	Scale  float64
	Offset float64
}

// NewLinearAxis creates an affine scalar transform from [srcMin, srcMax] to [dstMin, dstMax].
func NewLinearAxis(srcMin, srcMax, dstMin, dstMax float64) LinearAxis {
	den := srcMax - srcMin
	if den == 0 {
		return LinearAxis{Offset: dstMin}
	}
	scale := (dstMax - dstMin) / den
	return LinearAxis{
		Scale:  scale,
		Offset: dstMin - srcMin*scale,
	}
}

func (a LinearAxis) Forward(v float64) float64 { return a.Scale*v + a.Offset }

func (a LinearAxis) Inverse(v float64) (float64, bool) {
	if a.Scale == 0 {
		return 0, false
	}
	return (v - a.Offset) / a.Scale, true
}

// OffsetAxis adds a fixed device-space offset after the base transform.
type OffsetAxis struct {
	Base  AxisTransform
	Delta float64
}

func (a OffsetAxis) Forward(v float64) float64 {
	return axisOrIdentity(a.Base).Forward(v) + a.Delta
}

func (a OffsetAxis) Inverse(v float64) (float64, bool) {
	return axisOrIdentity(a.Base).Inverse(v - a.Delta)
}

// ComposedAxis composes two scalar transforms as B(A(v)).
type ComposedAxis struct {
	A AxisTransform
	B AxisTransform
}

func (a ComposedAxis) Forward(v float64) float64 {
	return axisOrIdentity(a.B).Forward(axisOrIdentity(a.A).Forward(v))
}

func (a ComposedAxis) Inverse(v float64) (float64, bool) {
	invB, ok := axisOrIdentity(a.B).Inverse(v)
	if !ok {
		return 0, false
	}
	return axisOrIdentity(a.A).Inverse(invB)
}

// SeparableT is a separable 2D transform made of independent x/y scalar transforms.
type SeparableT struct {
	X AxisTransform
	Y AxisTransform
}

// NewSeparable creates a new separable transform from x/y axis transforms.
func NewSeparable(x, y AxisTransform) SeparableT {
	return SeparableT{X: x, Y: y}
}

func (t SeparableT) Apply(p geom.Pt) geom.Pt {
	return geom.Pt{
		X: axisOrIdentity(t.X).Forward(p.X),
		Y: axisOrIdentity(t.Y).Forward(p.Y),
	}
}

func (t SeparableT) Invert(p geom.Pt) (geom.Pt, bool) {
	x, ok := axisOrIdentity(t.X).Inverse(p.X)
	if !ok {
		return geom.Pt{}, false
	}
	y, ok := axisOrIdentity(t.Y).Inverse(p.Y)
	if !ok {
		return geom.Pt{}, false
	}
	return geom.Pt{X: x, Y: y}, true
}

func (t SeparableT) XAxis() AxisTransform { return axisOrIdentity(t.X) }
func (t SeparableT) YAxis() AxisTransform { return axisOrIdentity(t.Y) }

// NewScaleTransform creates a separable transform from data space into axes-fraction space.
func NewScaleTransform(xs, ys Scale) SeparableT {
	return NewSeparable(ScaleAxis{Scale: xs}, ScaleAxis{Scale: ys})
}

// NewRectTransform creates a transform mapping the source rectangle into the destination rectangle.
func NewRectTransform(src, dst geom.Rect) SeparableT {
	return NewSeparable(
		NewLinearAxis(src.Min.X, src.Max.X, dst.Min.X, dst.Max.X),
		NewLinearAxis(src.Min.Y, src.Max.Y, dst.Min.Y, dst.Max.Y),
	)
}

// NewUnitRectTransform maps axes/figure fraction coordinates into the destination rectangle.
func NewUnitRectTransform(dst geom.Rect) SeparableT {
	return NewRectTransform(
		geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}},
		dst,
	)
}

// NewDisplayRectTransform maps Matplotlib-style fraction coordinates into
// display-space pixels. Display space is y-up with the origin at the
// bottom-left, exactly like Matplotlib: fraction 0 lands at dst.Min.Y (bottom)
// and fraction 1 at dst.Max.Y (top). Backends own the device y-inversion at
// rasterization (see docs/adr/0003-display-coordinate-contract.md).
func NewDisplayRectTransform(dst geom.Rect) SeparableT {
	return NewSeparable(
		NewLinearAxis(0, 1, dst.Min.X, dst.Max.X),
		NewLinearAxis(0, 1, dst.Min.Y, dst.Max.Y),
	)
}

// ChainSeparable composes two separable transforms.
func ChainSeparable(a, b Separable) SeparableT {
	return NewSeparable(
		ComposedAxis{A: a.XAxis(), B: b.XAxis()},
		ComposedAxis{A: a.YAxis(), B: b.YAxis()},
	)
}

// Blend combines the x component of one transform with the y component of another.
func Blend(xTrans, yTrans Separable) SeparableT {
	return NewSeparable(xTrans.XAxis(), yTrans.YAxis())
}

// NewOffset creates a transform that applies a device-space translation after the base transform.
func NewOffset(base T, delta geom.Pt) OffsetT {
	return OffsetT{Base: base, Delta: delta}
}

// Frozen returns an immutable snapshot of the known Go transform graph.
//
// Value-style transforms are already immutable; dynamic cached transforms are
// resolved to their current transform and separable/offset graphs are copied
// recursively. Unknown transform implementations are returned as-is.
func Frozen(t T) T {
	switch v := t.(type) {
	case nil:
		return nil
	case AffineT:
		return v
	case SeparableT:
		return SeparableT{X: frozenAxis(v.X), Y: frozenAxis(v.Y)}
	case OffsetT:
		return OffsetT{Base: Frozen(v.Base), Delta: v.Delta}
	case *CachedTransform:
		return Frozen(v.Current())
	default:
		return t
	}
}

func frozenAxis(axis AxisTransform) AxisTransform {
	switch v := axis.(type) {
	case nil:
		return identityAxis{}
	case identityAxis:
		return v
	case LinearAxis:
		return v
	case ScaleAxis:
		return ScaleAxis{Scale: frozenScale(v.Scale)}
	case OffsetAxis:
		return OffsetAxis{Base: frozenAxis(v.Base), Delta: v.Delta}
	case ComposedAxis:
		return ComposedAxis{A: frozenAxis(v.A), B: frozenAxis(v.B)}
	default:
		return axis
	}
}

func frozenScale(scale Scale) Scale {
	if scale == nil {
		return nil
	}
	if setter, ok := scale.(DomainSetter); ok {
		minVal, maxVal := scale.Domain()
		return setter.WithDomain(minVal, maxVal)
	}
	return scale
}

// OffsetT applies a fixed offset after the base transform.
type OffsetT struct {
	Base  T
	Delta geom.Pt
}

func (t OffsetT) Apply(p geom.Pt) geom.Pt {
	if t.Base != nil {
		p = t.Base.Apply(p)
	}
	return geom.Pt{X: p.X + t.Delta.X, Y: p.Y + t.Delta.Y}
}

func (t OffsetT) Invert(p geom.Pt) (geom.Pt, bool) {
	shifted := geom.Pt{X: p.X - t.Delta.X, Y: p.Y - t.Delta.Y}
	if t.Base == nil {
		return shifted, true
	}
	return t.Base.Invert(shifted)
}

func axisOrIdentity(a AxisTransform) AxisTransform {
	if a == nil {
		return identityAxis{}
	}
	return a
}

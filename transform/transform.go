package transform

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

// T is a 2D transform that maps points and can be inverted.
type T interface {
	Apply(p geom.Pt) geom.Pt
	Invert(p geom.Pt) (geom.Pt, bool)
}

// AffineT wraps a geom.Affine to satisfy T.
type AffineT struct{ M geom.Affine }

func NewAffine(m geom.Affine) AffineT { return AffineT{M: m} }

func (a AffineT) Apply(p geom.Pt) geom.Pt { return a.M.Apply(p) }

func (a AffineT) Invert(p geom.Pt) (geom.Pt, bool) {
	inv, ok := a.M.Invert()
	if !ok {
		return geom.Pt{}, false
	}
	return inv.Apply(p), true
}

// AffineProvider is the capability interface a transform implements to declare
// its exact affine representation, mirroring Matplotlib's
// Transform.get_affine()/is_affine pair. A transform that can be represented
// exactly by a single affine matrix returns (matrix, true); one that cannot
// (e.g. a nonlinear projection) returns (_, false).
//
// AsAffine consults this interface for any transform outside the package's
// built-in set, so a third-party T can participate in affine flattening
// instead of being treated as opaquely non-affine.
type AffineProvider interface {
	AsAffine() (geom.Affine, bool)
}

// AsAffine returns t as an affine matrix when the transform graph is known to
// be affine. Nonlinear scales return false; unknown transform implementations
// are consulted via the AffineProvider interface and otherwise return false.
func AsAffine(t T) (geom.Affine, bool) {
	switch v := Frozen(t).(type) {
	case nil:
		return geom.Identity(), true
	case AffineT:
		return v.M, true
	case SeparableT:
		xScale, xOffset, ok := affineAxis(v.XAxis())
		if !ok {
			return geom.Affine{}, false
		}
		yScale, yOffset, ok := affineAxis(v.YAxis())
		if !ok {
			return geom.Affine{}, false
		}
		return geom.Affine{A: xScale, D: yScale, E: xOffset, F: yOffset}, true
	case Chain:
		a, ok := AsAffine(v.A)
		if !ok {
			return geom.Affine{}, false
		}
		b, ok := AsAffine(v.B)
		if !ok {
			return geom.Affine{}, false
		}
		return b.Mul(a), true
	case OffsetT:
		base, ok := AsAffine(v.Base)
		if !ok {
			return geom.Affine{}, false
		}
		return geom.Affine{A: 1, D: 1, E: v.Delta.X, F: v.Delta.Y}.Mul(base), true
	default:
		if ap, ok := v.(AffineProvider); ok {
			return ap.AsAffine()
		}
		return geom.Affine{}, false
	}
}

// splitAffine decomposes t into a non-affine remainder and the maximal trailing
// affine such that t.Apply(p) == trailing.Apply(nonAffine.Apply(p)). A nil
// nonAffine means the points pass through unchanged; fullyAffine reports
// nonAffine == nil (the whole transform reduces to a single affine matrix).
//
// This mirrors Matplotlib's get_affine()/transform_non_affine() split: the
// expensive non-affine projection can be cached while the cheap trailing affine
// is re-applied each draw. The recursion reuses Frozen and AsAffine, so it
// extracts exactly the same affine matrices the affine fast-path already trusts.
func splitAffine(t T) (nonAffine T, trailing geom.Affine, fullyAffine bool) {
	switch v := Frozen(t).(type) {
	case nil:
		return nil, geom.Identity(), true
	case AffineT:
		return nil, v.M, true
	case SeparableT:
		if m, ok := AsAffine(v); ok {
			return nil, m, true
		}
		return v, geom.Identity(), false
	case OffsetT:
		baseNonAffine, baseAffine, baseFull := splitAffine(v.Base)
		offset := geom.Affine{A: 1, D: 1, E: v.Delta.X, F: v.Delta.Y}
		return baseNonAffine, offset.Mul(baseAffine), baseFull
	case Chain:
		bNonAffine, bAffine, bFull := splitAffine(v.B)
		if bFull {
			aNonAffine, aAffine, aFull := splitAffine(v.A)
			return aNonAffine, bAffine.Mul(aAffine), aFull
		}
		// B has a non-affine remainder: the maximal trailing affine is B's
		// trailing affine and the non-affine part is A followed by B's remainder.
		return Chain{A: v.A, B: bNonAffine}, bAffine, false
	default:
		if ap, ok := v.(AffineProvider); ok {
			if m, ok := ap.AsAffine(); ok {
				return nil, m, true
			}
		}
		return v, geom.Identity(), false
	}
}

func affineAxis(axis AxisTransform) (scale, offset float64, ok bool) {
	switch v := axisOrIdentity(axis).(type) {
	case identityAxis:
		return 1, 0, true
	case LinearAxis:
		return v.Scale, v.Offset, true
	case ScaleAxis:
		if v.Scale == nil {
			return 1, 0, true
		}
		linear, ok := v.Scale.(Linear)
		if !ok {
			return 0, 0, false
		}
		axis := NewLinearAxis(linear.Min, linear.Max, 0, 1)
		return axis.Scale, axis.Offset, true
	case OffsetAxis:
		scale, offset, ok := affineAxis(v.Base)
		if !ok {
			return 0, 0, false
		}
		return scale, offset + v.Delta, true
	case ComposedAxis:
		aScale, aOffset, ok := affineAxis(v.A)
		if !ok {
			return 0, 0, false
		}
		bScale, bOffset, ok := affineAxis(v.B)
		if !ok {
			return 0, 0, false
		}
		return bScale * aScale, bScale*aOffset + bOffset, true
	default:
		return 0, 0, false
	}
}

// Scale maps a scalar domain to unit space [0..1] and back.
type Scale interface {
	Fwd(x float64) float64
	Inv(u float64) (float64, bool)
	Domain() (min, max float64)
}

// DomainSetter can clone a scale with a different data domain.
type DomainSetter interface {
	WithDomain(min, max float64) Scale
}

// Linear maps [Min,Max] linearly to [0,1].
type Linear struct{ Min, Max float64 }

func NewLinear(minVal, maxVal float64) Linear { return Linear{Min: minVal, Max: maxVal} }

func (s Linear) Domain() (float64, float64) { return s.Min, s.Max }

func (s Linear) WithDomain(min, max float64) Scale {
	s.Min = min
	s.Max = max
	return s
}

func (s Linear) Fwd(x float64) float64 {
	den := s.Max - s.Min
	if den == 0 { // degenerate domain
		return 0
	}
	return (x - s.Min) / den
}

func (s Linear) Inv(u float64) (float64, bool) {
	den := s.Max - s.Min
	if den == 0 {
		return s.Min, false
	}
	return s.Min + u*den, true
}

// Log maps (Min,Max], Min>0, Base>1 to [0,1] using log with the given base.
type Log struct {
	Min, Max, Base float64
	NonPositive    NonPositiveMode
}

func NewLog(minVal, maxVal, base float64) Log {
	return Log{Min: minVal, Max: maxVal, Base: base, NonPositive: NonPositiveMask}
}

func (s Log) Domain() (float64, float64) { return s.Min, s.Max }

func (s Log) WithDomain(min, max float64) Scale {
	s.Min, s.Max = normalizeLogDomain(min, max, s.Base)
	return s
}

func (s Log) valid() bool {
	if s.Base <= 1 {
		return false
	}
	if s.Min <= 0 || s.Max <= 0 {
		return false
	}
	if s.Min == s.Max {
		return false
	}
	return true
}

func (s Log) Fwd(x float64) float64 {
	if !s.valid() {
		return 0
	}
	if x <= 0 { // outside domain
		if s.NonPositive != NonPositiveClip {
			return math.NaN()
		}
		x = logClipFloor(s.Min, s.Max, s.Base)
	}
	lb := math.Log(s.Base)
	lo := math.Log(s.Min) / lb
	hi := math.Log(s.Max) / lb
	vx := math.Log(x) / lb
	return (vx - lo) / (hi - lo)
}

func (s Log) Inv(u float64) (float64, bool) {
	if !s.valid() {
		return s.Min, false
	}
	lb := math.Log(s.Base)
	lo := math.Log(s.Min) / lb
	hi := math.Log(s.Max) / lb
	vx := lo + u*(hi-lo)
	x := math.Pow(s.Base, vx)
	if x <= 0 {
		return 0, false
	}
	return x, true
}

// Chain composes two transforms: Apply(p) = B(A(p))
type Chain struct{ A, B T }

func (c Chain) Apply(p geom.Pt) geom.Pt { return c.B.Apply(c.A.Apply(p)) }

func (c Chain) Invert(p geom.Pt) (geom.Pt, bool) {
	// Inverse: A^{-1}(B^{-1}(p)) if both exist
	pb, ok := c.B.Invert(p)
	if !ok {
		return geom.Pt{}, false
	}
	pa, ok := c.A.Invert(pb)
	if !ok {
		return geom.Pt{}, false
	}
	return pa, true
}

// Axes2D composes per-axis scales with an axes->pixel affine transform.
type Axes2D struct {
	X           Scale
	Y           Scale
	AxesToPixel AffineT
}

// NewAxes2D creates a transform mapping data (x,y) -> pixel coordinates.
func NewAxes2D(xs, ys Scale, axesToPixel AffineT) Axes2D {
	return Axes2D{X: xs, Y: ys, AxesToPixel: axesToPixel}
}

//nolint:gocritic // Axes2D is stored by value throughout draw state; keeping value receivers avoids API churn.
func (t Axes2D) Apply(p geom.Pt) geom.Pt {
	u := t.X.Fwd(p.X)
	v := t.Y.Fwd(p.Y)
	return t.AxesToPixel.Apply(geom.Pt{X: u, Y: v})
}

//nolint:gocritic // Axes2D is stored by value throughout draw state; keeping value receivers avoids API churn.
func (t Axes2D) Invert(p geom.Pt) (geom.Pt, bool) {
	up, ok := t.AxesToPixel.Invert(p)
	if !ok {
		return geom.Pt{}, false
	}
	x, okx := t.X.Inv(up.X)
	y, oky := t.Y.Inv(up.Y)
	if !okx || !oky {
		return geom.Pt{}, false
	}
	return geom.Pt{X: x, Y: y}, true
}

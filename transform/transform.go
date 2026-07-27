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

// NewAffine2D returns an identity affine transform suitable for fluent
// rotation and skew composition.
func NewAffine2D() AffineT { return NewAffine(geom.Identity()) }

func (a AffineT) Apply(p geom.Pt) geom.Pt { return a.M.Apply(p) }

func (a AffineT) Invert(p geom.Pt) (geom.Pt, bool) {
	inv, ok := a.M.Invert()
	if !ok {
		return geom.Pt{}, false
	}
	return inv.Apply(p), true
}

// Rotate returns a copy with a counter-clockwise rotation, in radians,
// left-composed with the current transform.
func (a AffineT) Rotate(theta float64) AffineT {
	sin, cos := math.Sincos(theta)
	return NewAffine(geom.Affine{A: cos, B: sin, C: -sin, D: cos}.Mul(a.M))
}

// RotateDegrees is the degree-based form of Rotate.
func (a AffineT) RotateDegrees(degrees float64) AffineT {
	return a.Rotate(degrees * math.Pi / 180)
}

// RotateAround returns a copy rotated by theta radians around center.
func (a AffineT) RotateAround(center geom.Pt, theta float64) AffineT {
	toOrigin := geom.Affine{A: 1, D: 1, E: -center.X, F: -center.Y}
	fromOrigin := geom.Affine{A: 1, D: 1, E: center.X, F: center.Y}
	sin, cos := math.Sincos(theta)
	rotation := geom.Affine{A: cos, B: sin, C: -sin, D: cos}
	return NewAffine(fromOrigin.Mul(rotation).Mul(toOrigin).Mul(a.M))
}

// RotateAroundDegrees is the degree-based form of RotateAround.
func (a AffineT) RotateAroundDegrees(center geom.Pt, degrees float64) AffineT {
	return a.RotateAround(center, degrees*math.Pi/180)
}

// Skew returns a copy with x- and y-axis shear angles, in radians,
// left-composed with the current transform.
func (a AffineT) Skew(xShear, yShear float64) AffineT {
	shear := geom.Affine{
		A: 1,
		B: math.Tan(yShear),
		C: math.Tan(xShear),
		D: 1,
	}
	return NewAffine(shear.Mul(a.M))
}

// SkewDegrees is the degree-based form of Skew.
func (a AffineT) SkewDegrees(xShear, yShear float64) AffineT {
	return a.Skew(xShear*math.Pi/180, yShear*math.Pi/180)
}

// ScaledTranslation translates by an offset after transforming that offset
// through another transform. It mirrors Matplotlib's ScaledTranslation and is
// useful for offsets expressed in points, inches, or another scaled space.
type ScaledTranslation struct {
	TransformNode
	translation    geom.Pt
	scaleTransform T
}

// NewScaledTranslation creates a live scaled translation. Changes propagated
// by a dynamic scale transform invalidate this transform's dependents.
func NewScaledTranslation(translation geom.Pt, scaleTransform T) *ScaledTranslation {
	t := &ScaledTranslation{
		translation:    translation,
		scaleTransform: scaleTransform,
	}
	if node := dependencyNode(scaleTransform); node != nil {
		node.AddDependent(&t.TransformNode)
	}
	return t
}

func (t *ScaledTranslation) delta() geom.Pt {
	if t == nil {
		return geom.Pt{}
	}
	if t.scaleTransform == nil {
		return t.translation
	}
	return t.scaleTransform.Apply(t.translation)
}

// Apply applies the scaled translation.
func (t *ScaledTranslation) Apply(p geom.Pt) geom.Pt {
	delta := t.delta()
	return geom.Pt{X: p.X + delta.X, Y: p.Y + delta.Y}
}

// Invert removes the scaled translation.
func (t *ScaledTranslation) Invert(p geom.Pt) (geom.Pt, bool) {
	delta := t.delta()
	return geom.Pt{X: p.X - delta.X, Y: p.Y - delta.Y}, true
}

// AsAffine returns the current pure-translation matrix.
func (t *ScaledTranslation) AsAffine() (geom.Affine, bool) {
	delta := t.delta()
	return geom.Affine{A: 1, D: 1, E: delta.X, F: delta.Y}, true
}

// TransformWrapper is a replaceable transform proxy that keeps a stable node
// in a transform graph while its child changes.
type TransformWrapper struct {
	TransformNode
	child T
}

// NewTransformWrapper creates a replaceable proxy around child.
func NewTransformWrapper(child T) *TransformWrapper {
	w := &TransformWrapper{}
	w.setChild(child, false)
	return w
}

// Child returns the currently wrapped transform.
func (w *TransformWrapper) Child() T {
	if w == nil {
		return nil
	}
	return w.child
}

// Set replaces the wrapped transform and invalidates downstream dependents.
// All T implementations are two-dimensional, so Matplotlib's dimension
// compatibility condition is enforced by the interface itself.
func (w *TransformWrapper) Set(child T) {
	if w == nil {
		return
	}
	w.setChild(child, true)
}

func (w *TransformWrapper) setChild(child T, invalidate bool) {
	if node := dependencyNode(w.child); node != nil {
		node.RemoveDependent(&w.TransformNode)
	}
	w.child = child
	if node := dependencyNode(child); node != nil {
		node.AddDependent(&w.TransformNode)
	}
	if invalidate {
		w.Invalidate(InvalidAll)
	}
}

// Apply delegates to the current child; a nil child is the identity.
func (w *TransformWrapper) Apply(p geom.Pt) geom.Pt {
	if w == nil || w.child == nil {
		return p
	}
	return w.child.Apply(p)
}

// Invert delegates to the current child; a nil child is the identity.
func (w *TransformWrapper) Invert(p geom.Pt) (geom.Pt, bool) {
	if w == nil || w.child == nil {
		return p, true
	}
	return w.child.Invert(p)
}

// AsAffine delegates affine extraction to the current child.
func (w *TransformWrapper) AsAffine() (geom.Affine, bool) {
	if w == nil || w.child == nil {
		return geom.Identity(), true
	}
	return AsAffine(w.child)
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

// AffineScale is the capability interface a Scale implements to declare that its
// forward map is affine, i.e. that it carries Domain() linearly onto [0, 1].
// affineScaleDomain consults it for any Scale that is not the built-in Linear,
// so a wrapper such as an inverted axis can participate in affine flattening
// instead of forcing the whole transform graph onto the staged path.
//
// Declaring affineness rather than returning coefficients is deliberate: the
// flattened axis is rebuilt from the endpoints with NewLinearAxis, which is the
// same arithmetic as Matplotlib's BboxTransformFrom(viewLim). A wrapper that
// returned its own coefficients would be algebraically equal but not bit-equal
// (an inverted axis computing -scale, 1-offset instead of mapping the swapped
// domain), and that last ULP is exactly what the snapped/truncated raster
// paths turn into a whole pixel.
type AffineScale interface {
	Scale
	IsAffineScale() bool
}

// IsAffineScale reports whether s carries its domain linearly onto [0, 1], so a
// Scale wrapper can answer the AffineScale question for the scale it wraps
// without restating the built-in cases. A nil scale is the identity, which is
// affine.
func IsAffineScale(s Scale) bool {
	_, _, ok := affineScaleDomain(s)
	return ok
}

// affineScaleDomain reports the domain a Scale maps linearly onto [0, 1], or
// ok=false when the scale is nonlinear (log, symlog, function scales).
func affineScaleDomain(s Scale) (srcMin, srcMax float64, ok bool) {
	switch v := s.(type) {
	case nil:
		return 0, 1, true
	case Linear:
		return v.Min, v.Max, true
	case AffineScale:
		if !v.IsAffineScale() {
			return 0, 0, false
		}
		srcMin, srcMax = v.Domain()
		return srcMin, srcMax, true
	default:
		return 0, 0, false
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
		srcMin, srcMax, ok := affineScaleDomain(v.Scale)
		if !ok {
			return 0, 0, false
		}
		axis := NewLinearAxis(srcMin, srcMax, 0, 1)
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
	// Matplotlib's LogScale defaults to nonpositive='clip' (scale.py), so
	// non-positive data lands far below the axis rather than being masked away.
	return Log{Min: minVal, Max: maxVal, Base: base, NonPositive: NonPositiveClip}
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

// logClipSentinel mirrors matplotlib's LogTransform clip value: non-positive
// inputs have their log-space output pinned to -1000 (scale.py) so, once
// normalized, the point falls far below the axis and is viewport-clipped.
const logClipSentinel = -1000.0

func (s Log) Fwd(x float64) float64 {
	if !s.valid() {
		return 0
	}
	lb := math.Log(s.Base)
	lo := math.Log(s.Min) / lb
	hi := math.Log(s.Max) / lb
	var vx float64
	if x <= 0 { // outside domain
		if s.NonPositive != NonPositiveClip {
			return math.NaN()
		}
		vx = logClipSentinel
	} else {
		vx = math.Log(x) / lb
	}
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
	if a, ok := t.linearAffine(); ok {
		return geom.Pt{X: a.ax*p.X + a.ex, Y: a.ay*p.Y + a.ey}
	}
	u := t.X.Fwd(p.X)
	v := t.Y.Fwd(p.Y)
	return t.AxesToPixel.Apply(geom.Pt{X: u, Y: v})
}

// axesLinearAffine is the data->pixel mapping of an unrotated linear axes,
// collapsed to one multiply-add per coordinate.
type axesLinearAffine struct{ ax, ex, ay, ey float64 }

// linearAffine reproduces Matplotlib's composition of transLimits with
// transAxes, which multiplies the two matrices *before* mapping any point:
// `x*(w/den) + (w*(-min/den) + x0)`. Evaluating the two stages separately, as
// `w*((x-min)/den) + x0`, is the same map in exact arithmetic but not in
// float64 — it lands units_categories' bar edge on 267.49999999999994 where
// Matplotlib has exactly 267.5, and the AGG path snapper (`floor(v+0.5)`) then
// puts the edge one pixel left. The fast path is restricted to linear scales
// and an unrotated axes affine; anything else keeps the staged evaluation.
func (t Axes2D) linearAffine() (axesLinearAffine, bool) { //nolint:gocritic // see Apply.
	xs, ok := t.X.(Linear)
	if !ok {
		return axesLinearAffine{}, false
	}
	ys, ok := t.Y.(Linear)
	if !ok {
		return axesLinearAffine{}, false
	}
	m := t.AxesToPixel.M
	if m.B != 0 || m.C != 0 {
		return axesLinearAffine{}, false
	}
	xden := xs.Max - xs.Min
	yden := ys.Max - ys.Min
	if xden == 0 || yden == 0 {
		return axesLinearAffine{}, false
	}
	xScale := 1.0 / xden
	yScale := 1.0 / yden
	return axesLinearAffine{
		ax: m.A * xScale,
		ex: m.A*(-xs.Min*xScale) + m.E,
		ay: m.D * yScale,
		ey: m.D*(-ys.Min*yScale) + m.F,
	}, true
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

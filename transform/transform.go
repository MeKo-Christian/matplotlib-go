package transform

import (
    "math"
    "matplotlib-go/internal/geom"
)

// T is a 2D transform that maps points and can be inverted.
type T interface {
    Apply(p geom.Pt) geom.Pt
    Invert(p geom.Pt) (geom.Pt, bool)
}

// AffineT wraps a geom.Affine to satisfy T.
type AffineT struct{ M geom.Affine }

func NewAffine(M geom.Affine) AffineT { return AffineT{M: M} }

func (a AffineT) Apply(p geom.Pt) geom.Pt { return a.M.Apply(p) }

func (a AffineT) Invert(p geom.Pt) (geom.Pt, bool) {
    inv, ok := a.M.Invert()
    if !ok { return geom.Pt{}, false }
    return inv.Apply(p), true
}

// Scale maps a scalar domain to unit space [0..1] and back.
type Scale interface {
    Fwd(x float64) float64
    Inv(u float64) (float64, bool)
    Domain() (min, max float64)
}

// Linear maps [Min,Max] linearly to [0,1].
type Linear struct{ Min, Max float64 }

func NewLinear(min, max float64) Linear { return Linear{Min: min, Max: max} }

func (s Linear) Domain() (float64, float64) { return s.Min, s.Max }

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
type Log struct{ Min, Max, Base float64 }

func NewLog(min, max, base float64) Log { return Log{Min: min, Max: max, Base: base} }

func (s Log) Domain() (float64, float64) { return s.Min, s.Max }

func (s Log) valid() bool {
    if s.Base <= 1 { return false }
    if s.Min <= 0 || s.Max <= 0 { return false }
    if s.Min == s.Max { return false }
    return true
}

func (s Log) Fwd(x float64) float64 {
    if !s.valid() { return 0 }
    if x <= 0 { // outside domain
        return math.NaN()
    }
    lb := math.Log(s.Base)
    lo := math.Log(s.Min) / lb
    hi := math.Log(s.Max) / lb
    vx := math.Log(x) / lb
    return (vx - lo) / (hi - lo)
}

func (s Log) Inv(u float64) (float64, bool) {
    if !s.valid() { return s.Min, false }
    lb := math.Log(s.Base)
    lo := math.Log(s.Min) / lb
    hi := math.Log(s.Max) / lb
    vx := lo + u*(hi-lo)
    x := math.Pow(s.Base, vx)
    if x <= 0 { return 0, false }
    return x, true
}

// Chain composes two transforms: Apply(p) = B(A(p))
type Chain struct{ A, B T }

func (c Chain) Apply(p geom.Pt) geom.Pt { return c.B.Apply(c.A.Apply(p)) }

func (c Chain) Invert(p geom.Pt) (geom.Pt, bool) {
    // Inverse: A^{-1}(B^{-1}(p)) if both exist
    pb, ok := c.B.Invert(p)
    if !ok { return geom.Pt{}, false }
    pa, ok := c.A.Invert(pb)
    if !ok { return geom.Pt{}, false }
    return pa, true
}

// Axes2D composes per-axis scales with an axes->pixel affine transform.
type Axes2D struct {
    X Scale
    Y Scale
    AxesToPixel AffineT
}

// NewAxes2D creates a transform mapping data (x,y) -> pixel coordinates.
func NewAxes2D(xs, ys Scale, axesToPixel AffineT) Axes2D {
    return Axes2D{X: xs, Y: ys, AxesToPixel: axesToPixel}
}

func (t Axes2D) Apply(p geom.Pt) geom.Pt {
    u := t.X.Fwd(p.X)
    v := t.Y.Fwd(p.Y)
    return t.AxesToPixel.Apply(geom.Pt{X: u, Y: v})
}

func (t Axes2D) Invert(p geom.Pt) (geom.Pt, bool) {
    up, ok := t.AxesToPixel.Invert(p)
    if !ok { return geom.Pt{}, false }
    x, okx := t.X.Inv(up.X)
    y, oky := t.Y.Inv(up.Y)
    if !okx || !oky { return geom.Pt{}, false }
    return geom.Pt{X: x, Y: y}, true
}


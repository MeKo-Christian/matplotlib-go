package geom

// F64 is the canonical float type used across geometry.
type F64 = float64

// Pt represents a 2D point.
type Pt struct{ X, Y F64 }

// Rect is an axis-aligned rectangle with Max-exclusive semantics.
// That is, a point p is inside r iff Min.X <= p.X < Max.X and Min.Y <= p.Y < Max.Y.
type Rect struct{ Min, Max Pt }

// W returns the width (Max.X - Min.X).
func (r Rect) W() F64 { return r.Max.X - r.Min.X }

// H returns the height (Max.Y - Min.Y).
func (r Rect) H() F64 { return r.Max.Y - r.Min.Y }

// Inflate expands (or contracts if negative) the rectangle by dx,dy on all sides.
func (r Rect) Inflate(dx, dy F64) Rect {
    return Rect{
        Min: Pt{r.Min.X - dx, r.Min.Y - dy},
        Max: Pt{r.Max.X + dx, r.Max.Y + dy},
    }
}

// Contains returns true if point p lies within r using Max-exclusive semantics.
func (r Rect) Contains(p Pt) bool {
    return p.X >= r.Min.X && p.X < r.Max.X && p.Y >= r.Min.Y && p.Y < r.Max.Y
}

// Intersect returns the intersection of r and b with Max-exclusive semantics.
func (r Rect) Intersect(b Rect) Rect {
    min := Pt{X: maxf(r.Min.X, b.Min.X), Y: maxf(r.Min.Y, b.Min.Y)}
    max := Pt{X: minf(r.Max.X, b.Max.X), Y: minf(r.Max.Y, b.Max.Y)}
    // If empty, collapse to empty at boundary (Min >= Max per axis)
    if max.X < min.X {
        max.X = min.X
    }
    if max.Y < min.Y {
        max.Y = min.Y
    }
    return Rect{Min: min, Max: max}
}

func maxf(a, b F64) F64 { if a > b { return a }; return b }
func minf(a, b F64) F64 { if a < b { return a }; return b }

// Cmd is a path verb.
type Cmd uint8

const (
    MoveTo Cmd = iota
    LineTo
    QuadTo
    CubicTo
    ClosePath
)

// Path holds a compact path representation. For each command, V stores the
// associated control/endpoint points as follows:
//   MoveTo: 1 point (new current position)
//   LineTo: 1 point (endpoint)
//   QuadTo: 2 points (control, endpoint)
//   CubicTo: 3 points (control1, control2, endpoint)
//   ClosePath: 0 points
type Path struct{
    V []Pt
    C []Cmd
}

// Clear resets the path to empty slices.
func (p *Path) Clear() { p.V = p.V[:0]; p.C = p.C[:0] }

// MoveTo appends a MoveTo command.
func (p *Path) MoveTo(to Pt) { p.C = append(p.C, MoveTo); p.V = append(p.V, to) }

// LineTo appends a LineTo command.
func (p *Path) LineTo(to Pt) { p.C = append(p.C, LineTo); p.V = append(p.V, to) }

// QuadTo appends a quadratic curve with control and endpoint.
func (p *Path) QuadTo(ctrl, to Pt) { p.C = append(p.C, QuadTo); p.V = append(p.V, ctrl, to) }

// CubicTo appends a cubic curve with two controls and an endpoint.
func (p *Path) CubicTo(c1, c2, to Pt) { p.C = append(p.C, CubicTo); p.V = append(p.V, c1, c2, to) }

// Close closes the current subpath.
func (p *Path) Close() { p.C = append(p.C, ClosePath) }

// Validate checks internal consistency between commands and vertices.
// It returns false if the number of vertices does not match expectations.
func (p *Path) Validate() bool {
    need := 0
    for _, c := range p.C {
        switch c {
        case MoveTo, LineTo:
            need += 1
        case QuadTo:
            need += 2
        case CubicTo:
            need += 3
        case ClosePath:
            // no vertices
        default:
            return false
        }
    }
    return need == len(p.V)
}

// Affine is a 2x3 matrix representing a 2D affine transform.
// Mapping: (x', y') = (A*x + C*y + E, B*x + D*y + F)
type Affine struct{ A, B, C, D, E, F F64 }

// Identity returns the identity transform.
func Identity() Affine { return Affine{A: 1, D: 1} }

// Mul composes this transform with n, returning m∘n (apply n, then this).
func (m Affine) Mul(n Affine) Affine {
    // 2x3 matrix multiply with implicit last column [0 0 1]^T
    // Linear part Lm * Ln
    a := m.A*n.A + m.C*n.B
    b := m.B*n.A + m.D*n.B
    c := m.A*n.C + m.C*n.D
    d := m.B*n.C + m.D*n.D
    e := m.A*n.E + m.C*n.F + m.E
    f := m.B*n.E + m.D*n.F + m.F
    return Affine{A: a, B: b, C: c, D: d, E: e, F: f}
}

// Apply applies the transform to a point.
func (m Affine) Apply(p Pt) Pt {
    return Pt{X: m.A*p.X + m.C*p.Y + m.E, Y: m.B*p.X + m.D*p.Y + m.F}
}

// Invert returns the inverse transform, if it exists.
func (m Affine) Invert() (Affine, bool) {
    det := m.A*m.D - m.C*m.B
    if det == 0 {
        return Affine{}, false
    }
    invA := m.D / det
    invB := -m.B / det
    invC := -m.C / det
    invD := m.A / det
    // Inverse translation: -L^{-1} * t
    invE := -(invA*m.E + invC*m.F)
    invF := -(invB*m.E + invD*m.F)
    return Affine{A: invA, B: invB, C: invC, D: invD, E: invE, F: invF}, true
}


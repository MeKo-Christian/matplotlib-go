package geom

// F64 is the canonical float type used across geometry.
type F64 = float64

// Pt represents a 2D point.
type Pt struct{ X, Y F64 }

// Rect is an axis-aligned rectangle with Max-exclusive semantics.
// That is, a point p is inside r iff Min.X <= p.X < Max.X and Min.Y <= p.Y < Max.Y.
type Rect struct{ Min, Max Pt }

// RectFromPoints returns the bounding rectangle for points.
func RectFromPoints(points ...Pt) (Rect, bool) {
	if len(points) == 0 {
		return Rect{}, false
	}
	r := Rect{Min: points[0], Max: points[0]}
	for _, p := range points[1:] {
		if p.X < r.Min.X {
			r.Min.X = p.X
		}
		if p.Y < r.Min.Y {
			r.Min.Y = p.Y
		}
		if p.X > r.Max.X {
			r.Max.X = p.X
		}
		if p.Y > r.Max.Y {
			r.Max.Y = p.Y
		}
	}
	return r, true
}

// W returns the width (Max.X - Min.X).
func (r Rect) W() F64 { return r.Max.X - r.Min.X }

// H returns the height (Max.Y - Min.Y).
func (r Rect) H() F64 { return r.Max.Y - r.Min.Y }

// Empty reports whether the rectangle has no positive area.
func (r Rect) Empty() bool { return r.W() <= 0 || r.H() <= 0 }

// Inflate expands (or contracts if negative) the rectangle by dx,dy on all sides.
func (r Rect) Inflate(dx, dy F64) Rect {
	return Rect{
		Min: Pt{r.Min.X - dx, r.Min.Y - dy},
		Max: Pt{r.Max.X + dx, r.Max.Y + dy},
	}
}

// Padded expands or contracts the rectangle equally on all sides.
func (r Rect) Padded(pad F64) Rect { return r.Inflate(pad, pad) }

// Expanded scales the rectangle about its center.
func (r Rect) Expanded(xScale, yScale F64) Rect {
	cx := (r.Min.X + r.Max.X) / 2
	cy := (r.Min.Y + r.Max.Y) / 2
	hw := r.W() * xScale / 2
	hh := r.H() * yScale / 2
	return Rect{Min: Pt{cx - hw, cy - hh}, Max: Pt{cx + hw, cy + hh}}
}

// Translated moves the rectangle by dx,dy.
func (r Rect) Translated(dx, dy F64) Rect {
	return Rect{
		Min: Pt{r.Min.X + dx, r.Min.Y + dy},
		Max: Pt{r.Max.X + dx, r.Max.Y + dy},
	}
}

// Contains returns true if point p lies within r using Max-exclusive semantics.
func (r Rect) Contains(p Pt) bool {
	return p.X >= r.Min.X && p.X < r.Max.X && p.Y >= r.Min.Y && p.Y < r.Max.Y
}

// ContainsInclusive returns true if point p lies inside r including Max edges.
func (r Rect) ContainsInclusive(p Pt) bool {
	return p.X >= r.Min.X && p.X <= r.Max.X && p.Y >= r.Min.Y && p.Y <= r.Max.Y
}

// Intersect returns the intersection of r and b with Max-exclusive semantics.
func (r Rect) Intersect(b Rect) Rect {
	minPt := Pt{X: maxf(r.Min.X, b.Min.X), Y: maxf(r.Min.Y, b.Min.Y)}
	maxPt := Pt{X: minf(r.Max.X, b.Max.X), Y: minf(r.Max.Y, b.Max.Y)}
	// If empty, collapse to empty at boundary (Min >= Max per axis)
	if maxPt.X < minPt.X {
		maxPt.X = minPt.X
	}
	if maxPt.Y < minPt.Y {
		maxPt.Y = minPt.Y
	}
	return Rect{Min: minPt, Max: maxPt}
}

// Union returns the smallest rectangle covering r and b.
func (r Rect) Union(b Rect) Rect {
	if r.Empty() {
		return b
	}
	if b.Empty() {
		return r
	}
	return Rect{
		Min: Pt{X: minf(r.Min.X, b.Min.X), Y: minf(r.Min.Y, b.Min.Y)},
		Max: Pt{X: maxf(r.Max.X, b.Max.X), Y: maxf(r.Max.Y, b.Max.Y)},
	}
}

// UnionRects returns the smallest non-empty rectangle covering rects.
func UnionRects(rects ...Rect) (Rect, bool) {
	var out Rect
	have := false
	for _, r := range rects {
		if r.Empty() {
			continue
		}
		if !have {
			out = r
			have = true
			continue
		}
		out = out.Union(r)
	}
	return out, have
}

// Transformed returns the axis-aligned bounds of r after applying m.
func (r Rect) Transformed(m Affine) Rect {
	out, _ := RectFromPoints(
		m.Apply(r.Min),
		m.Apply(Pt{X: r.Max.X, Y: r.Min.Y}),
		m.Apply(r.Max),
		m.Apply(Pt{X: r.Min.X, Y: r.Max.Y}),
	)
	return out
}

// InverseTransformed returns the axis-aligned bounds after applying m's inverse.
func (r Rect) InverseTransformed(m Affine) (Rect, bool) {
	inv, ok := m.Invert()
	if !ok {
		return Rect{}, false
	}
	return r.Transformed(inv), true
}

func maxf(a, b F64) F64 {
	if a > b {
		return a
	}
	return b
}

func minf(a, b F64) F64 {
	if a < b {
		return a
	}
	return b
}

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
//
//	MoveTo: 1 point (new current position)
//	LineTo: 1 point (endpoint)
//	QuadTo: 2 points (control, endpoint)
//	CubicTo: 3 points (control1, control2, endpoint)
//	ClosePath: 0 points
type Path struct {
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

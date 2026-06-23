package geom

import "math"

// Path generators ported from Matplotlib's lib/matplotlib/path.py (3.10.9).
//
// These are the canonical builders for circles, arcs, wedges, ellipses, and
// regular polygons/stars. Core artists (patches, markers, pie/polar, hatches)
// should use these instead of rebuilding the geometry ad hoc, so there is a
// single faithful source for each shape.

const (
	// BezierCircleKappa is the optimal control-point offset for approximating a
	// quarter circle with a single cubic Bézier, (4/3)*tan(pi/8). It is the
	// constant used by the 4-cubic ellipse approximation (EllipseBezier).
	BezierCircleKappa = 0.5522847498307936

	// bezierCircleMagic is Matplotlib's Lancaster constant for the 8-cubic unit
	// circle approximation (Path.circle / Path.unit_circle). See
	// https://www.tinaja.com/glib/ellipse4.pdf.
	bezierCircleMagic = 0.2652031
)

// unitCircleVertices returns the 25 unique vertices of Matplotlib's 8-cubic
// unit circle (the 26th, CLOSEPOLY vertex, is implied by Close). The layout is
// MOVETO followed by eight CURVE4 segments.
func unitCircleVertices() [25]Pt {
	const m = bezierCircleMagic
	s := math.Sqrt(0.5)
	m45 := s * m
	return [25]Pt{
		{0, -1},

		{m, -1},
		{s - m45, -s - m45},
		{s, -s},

		{s + m45, -s + m45},
		{1, -m},
		{1, 0},

		{1, m},
		{s + m45, s - m45},
		{s, s},

		{s - m45, s + m45},
		{m, 1},
		{0, 1},

		{-m, 1},
		{-s + m45, s + m45},
		{-s, s},

		{-s - m45, s - m45},
		{-1, m},
		{-1, 0},

		{-1, -m},
		{-s - m45, -s + m45},
		{-s, -s},

		{-s + m45, -s - m45},
		{-m, -1},
		{0, -1},
	}
}

// Circle returns a Path approximating the circle of the given center and radius
// using eight cubic Bézier segments, matching Matplotlib's Path.circle.
func Circle(center Pt, radius float64) Path {
	v := unitCircleVertices()
	scale := func(p Pt) Pt {
		return Pt{X: p.X*radius + center.X, Y: p.Y*radius + center.Y}
	}
	path := Path{}
	path.MoveTo(scale(v[0]))
	for k := range 8 {
		i := k*3 + 1
		path.CubicTo(scale(v[i]), scale(v[i+1]), scale(v[i+2]))
	}
	path.Close()
	return path
}

// UnitCircle returns a Path of the unit circle (center 0, radius 1), matching
// Matplotlib's Path.unit_circle. It is the eight-cubic Lancaster approximation.
func UnitCircle() Path { return Circle(Pt{}, 1) }

// UnitCircleRightHalf returns a Path of the right half of the unit circle using
// four cubic Bézier segments, matching Matplotlib's Path.unit_circle_righthalf.
func UnitCircleRightHalf() Path {
	const m = bezierCircleMagic
	s := math.Sqrt(0.5)
	m45 := s * m
	path := Path{}
	path.MoveTo(Pt{0, -1})
	path.CubicTo(Pt{m, -1}, Pt{s - m45, -s - m45}, Pt{s, -s})
	path.CubicTo(Pt{s + m45, -s + m45}, Pt{1, -m}, Pt{1, 0})
	path.CubicTo(Pt{1, m}, Pt{s + m45, s - m45}, Pt{s, s})
	path.CubicTo(Pt{s - m45, s + m45}, Pt{m, 1}, Pt{0, 1})
	path.Close()
	return path
}

// EllipseBezier returns a Path approximating the axis-aligned ellipse of the
// given center and semi-axes (rx, ry) using four cubic Bézier segments with the
// kappa control-point offset. It returns an empty Path if rx or ry is not
// positive.
func EllipseBezier(center Pt, rx, ry float64) Path {
	if rx <= 0 || ry <= 0 {
		return Path{}
	}
	kx := rx * BezierCircleKappa
	ky := ry * BezierCircleKappa
	path := Path{}
	path.MoveTo(Pt{X: center.X + rx, Y: center.Y})
	path.CubicTo(
		Pt{X: center.X + rx, Y: center.Y + ky},
		Pt{X: center.X + kx, Y: center.Y + ry},
		Pt{X: center.X, Y: center.Y + ry},
	)
	path.CubicTo(
		Pt{X: center.X - kx, Y: center.Y + ry},
		Pt{X: center.X - rx, Y: center.Y + ky},
		Pt{X: center.X - rx, Y: center.Y},
	)
	path.CubicTo(
		Pt{X: center.X - rx, Y: center.Y - ky},
		Pt{X: center.X - kx, Y: center.Y - ry},
		Pt{X: center.X, Y: center.Y - ry},
	)
	path.CubicTo(
		Pt{X: center.X + kx, Y: center.Y - ry},
		Pt{X: center.X + rx, Y: center.Y - ky},
		Pt{X: center.X + rx, Y: center.Y},
	)
	path.Close()
	return path
}

// Arc returns a Path for the unit-circle arc from theta1 to theta2 (degrees),
// matching Matplotlib's Path.arc. theta2 is unwrapped to the shortest arc
// within 360 degrees. If n <= 0 the number of cubic segments is chosen
// automatically (one per 90 degrees, rounded up to a power of two).
func Arc(theta1, theta2 float64, n int) Path { return arc(theta1, theta2, n, false) }

// Wedge returns a Path for the unit-circle wedge (pie slice) from theta1 to
// theta2 (degrees), matching Matplotlib's Path.wedge. It is Arc with line
// segments from the origin to the arc endpoints and a closing segment.
func Wedge(theta1, theta2 float64, n int) Path { return arc(theta1, theta2, n, true) }

func arc(theta1, theta2 float64, n int, isWedge bool) Path {
	const halfPi = math.Pi * 0.5

	eta1 := theta1
	eta2 := theta2 - 360*math.Floor((theta2-theta1)/360)
	// Keep a full 2pi range from collapsing to 0 from floating-point error,
	// without expanding an existing 0 range.
	if theta2 != theta1 && eta2 <= eta1 {
		eta2 += 360
	}
	eta1 *= math.Pi / 180
	eta2 *= math.Pi / 180

	if n <= 0 {
		n = int(math.Pow(2, math.Ceil((eta2-eta1)/halfPi)))
	}
	if n < 1 {
		n = 1
	}
	deta := (eta2 - eta1) / float64(n)
	t := math.Tan(0.5 * deta)
	alpha := math.Sin(deta) * (math.Sqrt(4+3*t*t) - 1) / 3

	cos1, sin1 := math.Cos(eta1), math.Sin(eta1)
	path := Path{}
	if isWedge {
		path.MoveTo(Pt{X: 0, Y: 0})
		path.LineTo(Pt{X: cos1, Y: sin1})
	} else {
		path.MoveTo(Pt{X: cos1, Y: sin1})
	}
	for i := range n {
		a := eta1 + float64(i)*deta
		b := a + deta
		cosA, sinA := math.Cos(a), math.Sin(a)
		cosB, sinB := math.Cos(b), math.Sin(b)
		path.CubicTo(
			Pt{X: cosA - alpha*sinA, Y: sinA + alpha*cosA},
			Pt{X: cosB + alpha*sinB, Y: sinB - alpha*cosB},
			Pt{X: cosB, Y: sinB},
		)
	}
	if isWedge {
		path.LineTo(Pt{X: 0, Y: 0})
		path.Close()
	}
	return path
}

// UnitRectangle returns a Path of the unit rectangle from (0, 0) to (1, 1),
// matching Matplotlib's Path.unit_rectangle.
func UnitRectangle() Path {
	path := Path{}
	path.MoveTo(Pt{0, 0})
	path.LineTo(Pt{1, 0})
	path.LineTo(Pt{1, 1})
	path.LineTo(Pt{0, 1})
	path.Close()
	return path
}

// UnitRegularPolygon returns a Path for a regular polygon with numVertices
// vertices, inscribed in the unit circle and centered at the origin, with the
// first vertex pointing up. It matches Matplotlib's Path.unit_regular_polygon.
// It returns an empty Path for numVertices < 3.
func UnitRegularPolygon(numVertices int) Path {
	if numVertices < 3 {
		return Path{}
	}
	step := 2 * math.Pi / float64(numVertices)
	path := Path{}
	for i := range numVertices {
		theta := step*float64(i) + math.Pi/2
		p := Pt{X: math.Cos(theta), Y: math.Sin(theta)}
		if i == 0 {
			path.MoveTo(p)
		} else {
			path.LineTo(p)
		}
	}
	path.Close()
	return path
}

// UnitRegularStar returns a Path for a regular star with numVertices points and
// outer radius 1, centered at the origin, with the first point pointing up.
// innerCircle is the ratio of the inner radius to the outer radius. It matches
// Matplotlib's Path.unit_regular_star. It returns an empty Path for
// numVertices < 3.
func UnitRegularStar(numVertices int, innerCircle float64) Path {
	if numVertices < 3 {
		return Path{}
	}
	ns2 := numVertices * 2
	step := 2 * math.Pi / float64(ns2)
	path := Path{}
	for i := range ns2 {
		r := 1.0
		if i%2 == 1 {
			r = innerCircle
		}
		theta := step*float64(i) + math.Pi/2
		p := Pt{X: r * math.Cos(theta), Y: r * math.Sin(theta)}
		if i == 0 {
			path.MoveTo(p)
		} else {
			path.LineTo(p)
		}
	}
	path.Close()
	return path
}

// UnitRegularAsterisk returns a Path for a regular asterisk with numVertices
// points, matching Matplotlib's Path.unit_regular_asterisk (a star whose inner
// radius collapses to the center).
func UnitRegularAsterisk(numVertices int) Path {
	return UnitRegularStar(numVertices, 0.0)
}

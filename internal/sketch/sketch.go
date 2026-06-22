// Package sketch implements Matplotlib's sketch/xkcd path filter: a
// deterministic sinusoidal "hand-drawn" perturbation applied to path vertices.
//
// It is a faithful port of Matplotlib's C++ Sketch converter
// (third_party/matplotlib/src/path_converters.h) together with the
// vpgen_segmentator it depends on. The algorithm operates in y-up display/pixel
// space: scale and length are in pixels, randomness is a unitless factor.
//
// The package is intentionally free of any render dependency so every backend
// can call it; callers pass the three scalar parameters directly.
package sketch

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

// dblEpsilon mirrors std::numeric_limits<double>::epsilon() used by Matplotlib
// to guard the derived constants against divide-by-zero.
const dblEpsilon = 2.220446049250313e-16

// Active reports whether sketch parameters will perturb a path. Matplotlib
// treats scale==0 as "no sketch"; length/randomness only scale the wiggle.
func Active(scale, length, randomness float64) bool { return scale != 0 }

// rng is Matplotlib's RandomNumberGenerator: a linear congruential generator
// with the Microsoft Visual C++ constants and an implicit 2^32 modulus.
type rng struct{ seed uint32 }

func (r *rng) double() float64 {
	r.seed = 214013*r.seed + 2531011
	return float64(r.seed) / 4294967296.0 // 2^32
}

// Apply returns path perturbed with Matplotlib's sketch/xkcd filter. The input
// is treated as y-up display-space coordinates. scale==0 (or an empty path)
// returns a clone unchanged. Curves are flattened to line segments and the
// result is a MoveTo/LineTo polyline, matching Matplotlib's segmented output.
func Apply(path geom.Path, scale, length, randomness float64) geom.Path {
	if scale == 0 || len(path.C) == 0 {
		return path.Clone()
	}

	var pScale, logRandomness float64
	if length > dblEpsilon && randomness > dblEpsilon {
		pScale = 2 * math.Pi / (length * randomness)
	}
	if randomness > dblEpsilon {
		logRandomness = 2 * math.Log(randomness)
	}

	subs := flatten(path)

	var out geom.Path
	r := rng{seed: 0}
	seg := newSegmentator()

	// Sketch state. Per Matplotlib, the RNG is seeded once per path (here, once
	// per Apply call); m_p and m_has_last reset on every MoveTo but the RNG
	// stream continues across subpaths.
	var p, lastX, lastY float64
	hasLast := false

	// emit runs one segmentator output vertex through the Sketch state machine
	// and appends the (possibly perturbed) result to out.
	emit := func(x, y float64, move bool) {
		if move {
			hasLast = false
			p = 0
		}
		if hasLast {
			p += math.Exp(r.double() * logRandomness)
			den := lastX - x
			num := lastY - y
			l := num*num + den*den
			lastX, lastY = x, y
			if l != 0 {
				l = math.Sqrt(l)
				roverlen := math.Sin(p*pScale) * scale / l
				x += roverlen * num
				y -= roverlen * den
			}
		} else {
			lastX, lastY = x, y
		}
		hasLast = true
		if move {
			out.MoveTo(geom.Pt{X: x, Y: y})
		} else {
			out.LineTo(geom.Pt{X: x, Y: y})
		}
	}

	drain := func() {
		for {
			x, y, cmd := seg.vertex()
			if cmd == segStop {
				return
			}
			emit(x, y, cmd == segMove)
		}
	}

	for _, sp := range subs {
		if len(sp.pts) == 0 {
			continue
		}
		seg.moveTo(sp.pts[0])
		drain()
		for _, pt := range sp.pts[1:] {
			seg.lineTo(pt)
			drain()
		}
		if sp.closed {
			// The closing edge is subdivided and wiggled like any other edge.
			seg.lineTo(sp.pts[0])
			drain()
			out.Close()
		}
	}

	return out
}

// subpath is a flattened polyline: curves expanded to points, lines preserved
// as single segments so the segmentator controls sampling density.
type subpath struct {
	pts    []geom.Pt
	closed bool
}

func flatten(path geom.Path) []subpath {
	var subs []subpath
	cur := -1
	vi := 0
	var start, prev geom.Pt

	ensure := func(pt geom.Pt) {
		subs = append(subs, subpath{pts: []geom.Pt{pt}})
		cur = len(subs) - 1
		start, prev = pt, pt
	}

	for _, c := range path.C {
		switch c {
		case geom.MoveTo:
			ensure(path.V[vi])
			vi++
		case geom.LineTo:
			to := path.V[vi]
			vi++
			if cur < 0 {
				ensure(to)
				continue
			}
			subs[cur].pts = append(subs[cur].pts, to)
			prev = to
		case geom.QuadTo:
			ctrl, to := path.V[vi], path.V[vi+1]
			vi += 2
			if cur < 0 {
				ensure(to)
				continue
			}
			n := curveSegments(dist(prev, ctrl) + dist(ctrl, to))
			for i := 1; i <= n; i++ {
				t := float64(i) / float64(n)
				subs[cur].pts = append(subs[cur].pts, quadPoint(prev, ctrl, to, t))
			}
			prev = to
		case geom.CubicTo:
			c1, c2, to := path.V[vi], path.V[vi+1], path.V[vi+2]
			vi += 3
			if cur < 0 {
				ensure(to)
				continue
			}
			n := curveSegments(dist(prev, c1) + dist(c1, c2) + dist(c2, to))
			for i := 1; i <= n; i++ {
				t := float64(i) / float64(n)
				subs[cur].pts = append(subs[cur].pts, cubicPoint(prev, c1, c2, to, t))
			}
			prev = to
		case geom.ClosePath:
			if cur >= 0 {
				subs[cur].closed = true
				prev = start
			}
		}
	}
	return subs
}

// curveSegments picks a flattening density from a control-polygon length
// estimate. The segmentator re-samples to ~1px afterwards, so this only needs
// to keep the polyline approximation sub-pixel; ~1 segment per 3px, clamped.
func curveSegments(estLen float64) int {
	n := min(int(estLen/3)+1, 200)
	return max(n, 8)
}

func dist(a, b geom.Pt) float64 { return math.Hypot(b.X-a.X, b.Y-a.Y) }

func quadPoint(p0, p1, p2 geom.Pt, t float64) geom.Pt {
	u := 1 - t
	a, b, c := u*u, 2*u*t, t*t
	return geom.Pt{X: a*p0.X + b*p1.X + c*p2.X, Y: a*p0.Y + b*p1.Y + c*p2.Y}
}

func cubicPoint(p0, p1, p2, p3 geom.Pt, t float64) geom.Pt {
	u := 1 - t
	a, b, c, d := u*u*u, 3*u*u*t, 3*u*t*t, t*t*t
	return geom.Pt{
		X: a*p0.X + b*p1.X + c*p2.X + d*p3.X,
		Y: a*p0.Y + b*p1.Y + c*p2.Y + d*p3.Y,
	}
}

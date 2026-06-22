package agg

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/sketch"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) preparePathForPaint(path geom.Path, paint *render.Paint) (geom.Path, bool) {
	if pathHasNonFiniteVertices(path) {
		path = removeNonFinitePathVertices(path)
	}
	if len(path.C) == 0 || !path.Validate() {
		return geom.Path{}, false
	}
	if r.pathOutsideVisibleArea(path, paint) {
		return geom.Path{}, false
	}
	// This pipeline runs in y-down device space, mirroring Matplotlib's
	// draw_path converter chain (snap → simplify → curve → sketch), which all
	// execute AFTER the (1,-1)+height device flip is folded into the transform.
	if shouldSnapPath(path, paint) {
		path = snapPath(path, paint)
	}
	if paint.Simplify && paint.SimplifyThreshold > 0 {
		path = simplifyLinePath(path, paint.SimplifyThreshold)
	}
	// The sketch/xkcd wiggle is applied last, in device space, so its sinusoidal
	// perpendicular displacement has the same sign as Matplotlib's reference.
	if sketch.Active(paint.Sketch.Scale, paint.Sketch.Length, paint.Sketch.Randomness) {
		path = sketch.Apply(path, paint.Sketch.Scale, paint.Sketch.Length, paint.Sketch.Randomness)
	}
	return path, len(path.C) > 0
}

func pathHasNonFiniteVertices(path geom.Path) bool {
	for _, pt := range path.V {
		if !finitePt(pt) {
			return true
		}
	}
	return false
}

func removeNonFinitePathVertices(path geom.Path) geom.Path {
	out := geom.Path{}
	vi := 0
	haveCurrent := false
	needMove := true

	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return out
			}
			to := path.V[vi]
			vi++
			if !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			out.MoveTo(to)
			haveCurrent = true
			needMove = false
		case geom.LineTo:
			if vi >= len(path.V) {
				return out
			}
			to := path.V[vi]
			vi++
			if !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			if !haveCurrent || needMove {
				out.MoveTo(to)
			} else {
				out.LineTo(to)
			}
			haveCurrent = true
			needMove = false
		case geom.QuadTo:
			if vi+1 >= len(path.V) {
				return out
			}
			ctrl, to := path.V[vi], path.V[vi+1]
			vi += 2
			if !finitePt(ctrl) || !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			if !haveCurrent || needMove {
				out.MoveTo(to)
			} else {
				out.QuadTo(ctrl, to)
			}
			haveCurrent = true
			needMove = false
		case geom.CubicTo:
			if vi+2 >= len(path.V) {
				return out
			}
			c1, c2, to := path.V[vi], path.V[vi+1], path.V[vi+2]
			vi += 3
			if !finitePt(c1) || !finitePt(c2) || !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			if !haveCurrent || needMove {
				out.MoveTo(to)
			} else {
				out.CubicTo(c1, c2, to)
			}
			haveCurrent = true
			needMove = false
		case geom.ClosePath:
			if haveCurrent && !needMove {
				out.Close()
			}
			haveCurrent = false
			needMove = true
		}
	}
	return out
}

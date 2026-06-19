package agg

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func shouldSnapPath(path geom.Path, paint *render.Paint) bool {
	switch paint.Snap {
	case render.SnapOn:
		return true
	case render.SnapOff:
		return false
	case render.SnapAuto:
	default:
		return false
	}
	if len(path.V) > 1024 {
		return false
	}
	vi := 0
	var last geom.Pt
	haveLast := false
	lineCount := 0
	curveCount := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return false
			}
			last = path.V[vi]
			vi++
			haveLast = true
		case geom.LineTo:
			if vi >= len(path.V) {
				return false
			}
			to := path.V[vi]
			vi++
			if haveLast && math.Abs(last.X-to.X) >= 1e-4 && math.Abs(last.Y-to.Y) >= 1e-4 {
				return false
			}
			lineCount++
			last = to
			haveLast = true
		case geom.QuadTo:
			if vi+1 >= len(path.V) || !haveLast {
				return false
			}
			ctrl := path.V[vi]
			to := path.V[vi+1]
			vi += 2
			if !isAxisAlignedCornerQuad(last, ctrl, to) {
				return false
			}
			curveCount++
			last = to
			haveLast = true
		case geom.CubicTo:
			if vi+2 >= len(path.V) || !haveLast {
				return false
			}
			c1 := path.V[vi]
			c2 := path.V[vi+1]
			to := path.V[vi+2]
			vi += 3
			if !isAxisAlignedCornerCubic(last, c1, c2, to) {
				return false
			}
			curveCount++
			last = to
			haveLast = true
		case geom.ClosePath:
			haveLast = false
		}
	}
	return curveCount == 0 || (lineCount >= 4 && curveCount >= 4)
}

func isAxisAlignedCornerQuad(from, ctrl, to geom.Pt) bool {
	const eps = 1e-4
	return (math.Abs(ctrl.Y-from.Y) < eps && math.Abs(ctrl.X-to.X) < eps) ||
		(math.Abs(ctrl.X-from.X) < eps && math.Abs(ctrl.Y-to.Y) < eps)
}

func isAxisAlignedCornerCubic(from, c1, c2, to geom.Pt) bool {
	const eps = 1e-4
	return (math.Abs(c1.Y-from.Y) < eps && math.Abs(c2.X-to.X) < eps) ||
		(math.Abs(c1.X-from.X) < eps && math.Abs(c2.Y-to.Y) < eps)
}

func snapPath(path geom.Path, paint *render.Paint) geom.Path {
	out := clonePath(path)
	snapValue := 0.0
	strokeWidth := 0.0
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		strokeWidth = paint.LineWidth
	}
	if int(math.Round(strokeWidth))%2 != 0 {
		snapValue = 0.5
	}
	for i, pt := range out.V {
		out.V[i] = geom.Pt{
			X: math.Floor(pt.X+0.5) + snapValue,
			Y: math.Floor(pt.Y+0.5) + snapValue,
		}
	}
	return out
}

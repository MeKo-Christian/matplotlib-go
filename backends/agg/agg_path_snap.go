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
			last = to
			haveLast = true
		case geom.QuadTo:
			return false
		case geom.CubicTo:
			return false
		case geom.ClosePath:
			haveLast = false
		}
	}
	return true
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
			X: snapPathCoordinate(pt.X) + snapValue,
			Y: snapPathCoordinate(pt.Y) + snapValue,
		}
	}
	return out
}

func snapPathCoordinate(v float64) float64 {
	return math.Floor(v + 0.5 + 1e-9)
}

package agg

import (
	"fmt"
	"math"
	"os"

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
		if os.Getenv("MPLGO_SNAP_PROBE") != "" {
			fmt.Printf("SNAP %.17g %.17g -> %.17g %.17g\n", pt.X, pt.Y, snapPathCoordinate(pt.X)+snapValue, snapPathCoordinate(pt.Y)+snapValue)
		}
		out.V[i] = geom.Pt{
			X: snapPathCoordinate(pt.X) + snapValue,
			Y: snapPathCoordinate(pt.Y) + snapValue,
		}
	}
	return out
}

// snapPathCoordinate mirrors PathSnapper::vertex in matplotlib's
// path_converters.h, which is `floor(v + 0.5)` with no tolerance. The epsilon
// this used to add pushed coordinates that land just below .5 — matshow_basic's
// tick line arrives at 211.49999999999997, carrying the same transform noise
// matplotlib has — up to the next pixel, drawing them one pixel low.
func snapPathCoordinate(v float64) float64 {
	return math.Floor(v + 0.5)
}

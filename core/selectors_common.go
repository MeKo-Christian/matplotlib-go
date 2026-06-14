package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

func polygonSelectorPathFromData(points []geom.Pt, ctx *DrawContext) geom.Path {
	path := geom.Path{}
	if len(points) == 0 || ctx == nil {
		return path
	}
	for i, pt := range points {
		d := ctx.DataToPixel.Apply(pt)
		if i == 0 {
			path.MoveTo(d)
			continue
		}
		path.LineTo(d)
	}
	return path
}

func normalizedRect(a, b geom.Pt) (geom.Pt, geom.Pt) {
	return geom.Pt{X: math.Min(a.X, b.X), Y: math.Min(a.Y, b.Y)}, geom.Pt{X: math.Max(a.X, b.X), Y: math.Max(a.Y, b.Y)}
}

func axesContains(axes []*Axes, target *Axes) bool {
	for _, ax := range axes {
		if ax == target {
			return true
		}
	}
	return false
}

func dedupeAxes(axes []*Axes) []*Axes {
	seen := map[*Axes]bool{}
	out := make([]*Axes, 0, len(axes))
	for _, ax := range axes {
		if ax == nil {
			continue
		}
		if seen[ax] {
			continue
		}
		seen[ax] = true
		out = append(out, ax)
	}
	return out
}

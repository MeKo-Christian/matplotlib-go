package vectorhatch

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
)

const tileSide = 72.0

// ShapePath is one vector hatch shape in a pattern tile.
type ShapePath struct {
	Path   geom.Path
	Filled bool
}

// ShapePaths returns tile-local paths for Matplotlib's shape hatch glyphs.
func ShapePaths(hatch string, spacing float64) []ShapePath {
	if spacing <= 0 {
		spacing = 8
	}
	var out []ShapePath
	appendCircles := func(count int, size float64, filled bool) {
		if count <= 0 {
			return
		}
		step := math.Max(2, spacing/float64(count))
		radius := math.Max(0.5, step*size)
		out = append(out, shapeGrid(step, filled, func(center geom.Pt) geom.Path {
			return circlePath(center, radius)
		})...)
	}
	appendStars := func(count int) {
		if count <= 0 {
			return
		}
		step := math.Max(2, spacing/float64(count))
		outer := math.Max(0.75, step/3)
		inner := outer * 0.5
		out = append(out, shapeGrid(step, true, func(center geom.Pt) geom.Path {
			return starPath(center, outer, inner)
		})...)
	}

	appendCircles(strings.Count(hatch, "o"), 0.20, false)
	appendCircles(strings.Count(hatch, "O"), 0.35, false)
	appendCircles(strings.Count(hatch, "."), 0.10, true)
	appendStars(strings.Count(hatch, "*"))
	return out
}

func shapeGrid(spacing float64, filled bool, build func(geom.Pt) geom.Path) []ShapePath {
	if spacing <= 0 {
		return nil
	}
	var out []ShapePath
	row := 0
	for y := spacing / 2; y <= tileSide+1e-9; y += spacing {
		offset := 0.0
		if row%2 == 1 {
			offset = spacing / 2
		}
		for x := spacing/2 + offset; x <= tileSide+1e-9; x += spacing {
			out = append(out, ShapePath{
				Path:   build(geom.Pt{X: x, Y: y}),
				Filled: filled,
			})
		}
		row++
	}
	return out
}

func circlePath(center geom.Pt, radius float64) geom.Path {
	k := radius * 0.5522847498307936
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: center.X + radius, Y: center.Y})
	path.CubicTo(
		geom.Pt{X: center.X + radius, Y: center.Y + k},
		geom.Pt{X: center.X + k, Y: center.Y + radius},
		geom.Pt{X: center.X, Y: center.Y + radius},
	)
	path.CubicTo(
		geom.Pt{X: center.X - k, Y: center.Y + radius},
		geom.Pt{X: center.X - radius, Y: center.Y + k},
		geom.Pt{X: center.X - radius, Y: center.Y},
	)
	path.CubicTo(
		geom.Pt{X: center.X - radius, Y: center.Y - k},
		geom.Pt{X: center.X - k, Y: center.Y - radius},
		geom.Pt{X: center.X, Y: center.Y - radius},
	)
	path.CubicTo(
		geom.Pt{X: center.X + k, Y: center.Y - radius},
		geom.Pt{X: center.X + radius, Y: center.Y - k},
		geom.Pt{X: center.X + radius, Y: center.Y},
	)
	path.Close()
	return path
}

func starPath(center geom.Pt, outer, inner float64) geom.Path {
	path := geom.Path{}
	for i := 0; i < 10; i++ {
		radius := outer
		if i%2 == 1 {
			radius = inner
		}
		angle := -math.Pi/2 + float64(i)*math.Pi/5
		pt := geom.Pt{
			X: center.X + radius*math.Cos(angle),
			Y: center.Y + radius*math.Sin(angle),
		}
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}

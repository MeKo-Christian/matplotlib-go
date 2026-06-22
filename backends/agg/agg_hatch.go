package agg

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// SupportsNativeHatch reports that AGG consumes render.Paint hatch metadata
// directly while rasterizing a path.
func (r *Renderer) SupportsNativeHatch() bool { return true }

func (r *Renderer) drawNativeHatch(clipPath geom.Path, paint *render.Paint) {
	if paint == nil || paint.Hatch == "" {
		return
	}
	color := colorWithForcedAlpha(paint.HatchColor, paint)
	if color.A <= 0 {
		return
	}
	bounds, ok := pathBounds(clipPath)
	if !ok {
		return
	}
	counts := hatchCounts(paint.Hatch)
	if len(counts) == 0 {
		return
	}

	oldPaths := r.clipPaths
	r.clipPaths = append(clonePaths(oldPaths), clonePath(clipPath))
	defer func() {
		r.clipPaths = oldPaths
	}()

	for pattern, count := range counts {
		spacing := math.Max(2, render.DefaultHatchSpacing/float64(count))
		if paint.HatchSpacing > 0 {
			spacing = math.Max(2, paint.HatchSpacing/float64(count))
		}
		if pattern == '/' || pattern == '\\' || pattern == 'x' || pattern == 'X' {
			dpi := float64(r.resolution)
			if dpi <= 0 {
				dpi = 72
			}
			// Matplotlib hatch.py creates diagonal hatches in a DPI-sized unit
			// tile; only about half of the generated unit lines cross any given
			// scanline, so this display-space phase spacing matches the rendered
			// AGG density for repeated diagonal hatch characters.
			spacing = math.Max(2, dpi/(3*float64(count)))
		}
		hatchPaint := render.Paint{
			Stroke:    color,
			LineWidth: paint.HatchLineWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapSquare,
			// Hatch strokes are always anti-aliased, matching Matplotlib's AGG
			// hatch rendering. This is independent of the filled shape's own
			// antialiasing (e.g. contourf bands draw with antialiased=False but
			// their hatch lines remain anti-aliased).
			Antialias: render.AntialiasOn,
			Snap:      render.SnapOff,
		}
		if hatchPatternIsFilled(pattern) {
			hatchPaint.Fill = color
		}
		if hatchPaint.LineWidth <= 0 {
			hatchPaint.LineWidth = 1
		}
		for _, hatchPath := range hatchPatternPaths(pattern, bounds, spacing) {
			if len(hatchPath.C) == 0 {
				continue
			}
			r.pathDevice(hatchPath, &hatchPaint)
		}
	}
}

func hatchCounts(pattern string) map[rune]int {
	counts := make(map[rune]int)
	for _, ch := range pattern {
		switch ch {
		case '|', '-', '/', '\\', '+', 'x', 'X', 'o', 'O', '.', '*':
			counts[ch]++
		}
	}
	return counts
}

func hatchPatternPaths(pattern rune, bounds geom.Rect, spacing float64) []geom.Path {
	switch pattern {
	case '|':
		return []geom.Path{verticalHatchPath(bounds, spacing)}
	case '-':
		return []geom.Path{horizontalHatchPath(bounds, spacing)}
	case '/':
		return []geom.Path{slashHatchPath(bounds, spacing)}
	case '\\':
		return []geom.Path{backslashHatchPath(bounds, spacing)}
	case '+':
		return []geom.Path{
			verticalHatchPath(bounds, spacing),
			horizontalHatchPath(bounds, spacing),
		}
	case 'x', 'X':
		return []geom.Path{
			slashHatchPath(bounds, spacing),
			backslashHatchPath(bounds, spacing),
		}
	case 'o':
		return circleHatchPaths(bounds, spacing, 0.20, true)
	case 'O':
		return circleHatchPaths(bounds, spacing, 0.35, true)
	case '.':
		return circleHatchPaths(bounds, spacing, 0.10, false)
	case '*':
		return starHatchPaths(bounds, spacing)
	default:
		return nil
	}
}

func hatchPatternIsFilled(pattern rune) bool {
	return pattern == 'o' || pattern == 'O' || pattern == '.' || pattern == '*'
}

func verticalHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	minX := math.Floor(bounds.Min.X/spacing)*spacing - spacing
	maxX := bounds.Max.X + spacing
	for x := minX; x <= maxX; x += spacing {
		path.MoveTo(geom.Pt{X: x, Y: bounds.Min.Y - spacing})
		path.LineTo(geom.Pt{X: x, Y: bounds.Max.Y + spacing})
	}
	return path
}

func circleHatchPaths(bounds geom.Rect, spacing, radiusFactor float64, ring bool) []geom.Path {
	if bounds.W() <= 0 || bounds.H() <= 0 || spacing <= 0 {
		return nil
	}
	radius := math.Max(0.5, spacing*radiusFactor)
	return tiledShapeHatchPaths(bounds, spacing, radius, func(center geom.Pt) geom.Path {
		return circleHatchPath(center, radius, ring)
	})
}

func starHatchPaths(bounds geom.Rect, spacing float64) []geom.Path {
	if bounds.W() <= 0 || bounds.H() <= 0 || spacing <= 0 {
		return nil
	}
	outer := math.Max(0.75, spacing/3)
	inner := outer * 0.5
	return tiledShapeHatchPaths(bounds, spacing, outer, func(center geom.Pt) geom.Path {
		return starHatchPath(center, outer, inner)
	})
}

func tiledShapeHatchPaths(bounds geom.Rect, spacing, radius float64, build func(geom.Pt) geom.Path) []geom.Path {
	yStart := math.Floor((bounds.Min.Y-radius)/spacing) * spacing
	yEnd := bounds.Max.Y + radius
	var paths []geom.Path
	for y := yStart; y <= yEnd+1e-9; y += spacing {
		row := int(math.Round(y / spacing))
		xOffset := 0.0
		if row%2 != 0 {
			xOffset = spacing / 2
		}
		xStart := math.Floor((bounds.Min.X-radius-xOffset)/spacing)*spacing + xOffset
		for x := xStart; x <= bounds.Max.X+radius+1e-9; x += spacing {
			paths = append(paths, build(geom.Pt{X: x, Y: y}))
		}
	}
	return paths
}

func circleHatchPath(center geom.Pt, radius float64, ring bool) geom.Path {
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
	if ring {
		inner := radius * 0.9
		ik := inner * 0.5522847498307936
		path.MoveTo(geom.Pt{X: center.X + inner, Y: center.Y})
		path.CubicTo(
			geom.Pt{X: center.X + inner, Y: center.Y - ik},
			geom.Pt{X: center.X + ik, Y: center.Y - inner},
			geom.Pt{X: center.X, Y: center.Y - inner},
		)
		path.CubicTo(
			geom.Pt{X: center.X - ik, Y: center.Y - inner},
			geom.Pt{X: center.X - inner, Y: center.Y - ik},
			geom.Pt{X: center.X - inner, Y: center.Y},
		)
		path.CubicTo(
			geom.Pt{X: center.X - inner, Y: center.Y + ik},
			geom.Pt{X: center.X - ik, Y: center.Y + inner},
			geom.Pt{X: center.X, Y: center.Y + inner},
		)
		path.CubicTo(
			geom.Pt{X: center.X + ik, Y: center.Y + inner},
			geom.Pt{X: center.X + inner, Y: center.Y + ik},
			geom.Pt{X: center.X + inner, Y: center.Y},
		)
		path.Close()
	}
	return path
}

func starHatchPath(center geom.Pt, outer, inner float64) geom.Path {
	points := make([]geom.Pt, 0, 10)
	for i := 0; i < 10; i++ {
		radius := outer
		if i%2 == 1 {
			radius = inner
		}
		angle := -math.Pi/2 + float64(i)*math.Pi/5
		points = append(points, geom.Pt{
			X: center.X + radius*math.Cos(angle),
			Y: center.Y + radius*math.Sin(angle),
		})
	}
	path := geom.Path{}
	for i, pt := range points {
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}

func horizontalHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	minY := math.Floor(bounds.Min.Y/spacing)*spacing - spacing
	maxY := bounds.Max.Y + spacing
	for y := minY; y <= maxY; y += spacing {
		path.MoveTo(geom.Pt{X: bounds.Min.X - spacing, Y: y})
		path.LineTo(geom.Pt{X: bounds.Max.X + spacing, Y: y})
	}
	return path
}

func slashHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	span := bounds.W() + bounds.H() + 2*spacing
	x0 := bounds.Min.X - span
	x1 := bounds.Max.X + span
	// Matplotlib AGG draws the unit-square hatch into an origin-anchored
	// DPI-sized tile and repeats that tile over the clipped patch.
	minC := bounds.Min.X + bounds.Min.Y - span
	maxC := bounds.Max.X + bounds.Max.Y + span
	start := math.Floor(minC/spacing) * spacing
	for c := start; c <= maxC; c += spacing {
		path.MoveTo(geom.Pt{X: x0, Y: c - x0})
		path.LineTo(geom.Pt{X: x1, Y: c - x1})
	}
	return path
}

func backslashHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	span := bounds.W() + bounds.H() + 2*spacing
	x0 := bounds.Min.X - span
	x1 := bounds.Max.X + span
	minC := bounds.Min.Y - bounds.Max.X - span
	maxC := bounds.Max.Y - bounds.Min.X + span
	start := math.Floor(minC/spacing) * spacing
	for c := start; c <= maxC; c += spacing {
		path.MoveTo(geom.Pt{X: x0, Y: x0 + c})
		path.LineTo(geom.Pt{X: x1, Y: x1 + c})
	}
	return path
}

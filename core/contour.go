package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func contourLineColor(level float64, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64, fallback render.Color) render.Color {
	if opt.Color != nil {
		color := *opt.Color
		color.A *= alpha
		return color
	}
	if len(opt.Colors) > 0 {
		idx := indexOfLevel(levels, level)
		color := opt.Colors[idx%len(opt.Colors)]
		color.A *= alpha
		return color
	}
	if opt.Colormap != nil {
		return mapping.Color(level, alpha)
	}
	fallback.A *= alpha
	return fallback
}

func containsPoint(points []geom.Pt, point geom.Pt) bool {
	for _, existing := range points {
		if sameContourPoint(existing, point) {
			return true
		}
	}
	return false
}

func sameContourPoint(a, b geom.Pt) bool {
	return math.Abs(a.X-b.X) <= 1e-9 && math.Abs(a.Y-b.Y) <= 1e-9
}

func interpolatePoint(a, b geom.Pt, t float64) geom.Pt {
	return geom.Pt{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
	}
}

func indexOfLevel(levels []float64, level float64) int {
	for i, candidate := range levels {
		if math.Abs(candidate-level) <= 1e-12 {
			return i
		}
	}
	return 0
}

func valueOrDefaultFloat(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}

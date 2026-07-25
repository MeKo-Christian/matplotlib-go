package plot3d

import (
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/contourgeom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const defaultPatchZ = 1.0

func scaleGridDashes(dashes []float64, width float64) []float64 {
	if len(dashes) == 0 || width <= 0 {
		return dashes
	}
	scaled := make([]float64, len(dashes))
	for i, dash := range dashes {
		scaled[i] = dash * width
	}
	return scaled
}

//nolint:gocritic // RC is intentionally copied so rendering observes one stable style snapshot.
func pointsToPixels(rc style.RC, points float64) float64 {
	dpi := rc.DPI
	if dpi <= 0 {
		dpi = style.CurrentDefaults().DPI
		if dpi <= 0 {
			dpi = 96
		}
	}
	return points * dpi / 72
}

func colorCycleAt(ax *core.Axes, index int) render.Color {
	if ax == nil {
		return render.Color{A: 1}
	}
	palette := ax.ResolvedRC().Palette()
	if len(palette) == 0 {
		return render.Color{A: 1}
	}
	index %= len(palette)
	if index < 0 {
		index += len(palette)
	}
	return palette[index]
}

//nolint:gocritic // MarkerStyle is a public value-semantic option shared with core.
func markerLineOnly(marker core.MarkerStyle) bool {
	if marker.Tuple != nil {
		return marker.Tuple.Style == core.MarkerTupleAsterisk
	}
	switch marker.Type {
	case core.MarkerPlus, core.MarkerCross, core.MarkerTriDown, core.MarkerTriUp,
		core.MarkerTriLeft, core.MarkerTriRight, core.MarkerVLine, core.MarkerHLine,
		core.MarkerTickLeft, core.MarkerTickRight, core.MarkerTickUp, core.MarkerTickDown:
		return true
	default:
		return false
	}
}

func validErrorValues(values []float64, count int) bool {
	if len(values) != 0 && len(values) != 1 && len(values) != count {
		return false
	}
	for _, value := range values {
		if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}

func contourLevels(values, explicit []float64, levelCount int, filled bool) []float64 {
	return contourgeom.Levels(values, explicit, levelCount, filled)
}

//nolint:gocritic // Triangulation is a value-semantic geometry input shared with core.
func contourPolylines(tri core.Triangulation, values, levels []float64) ([][]geom.Pt, []float64) {
	geometryTri := contourGeometryTriangulation(tri)
	return contourgeom.Polylines(&geometryTri, values, levels)
}

func contourGridPolylines(x, y []float64, data [][]float64, levels []float64) ([][]geom.Pt, []float64) {
	return contourgeom.GridPolylines(x, y, data, levels)
}

func contourCellBandPolygons(points [4]geom.Pt, values [4]float64, low, high float64) [][]geom.Pt {
	return contourgeom.CellBandPolygons(points, values, low, high)
}

func triangleBandPolygon(points [3]geom.Pt, values [3]float64, low, high float64) []geom.Pt {
	return contourgeom.TriangleBandPolygon(points, values, low, high)
}

//nolint:gocritic // The adapter builds detached contour geometry from a triangulation value.
func contourGeometryTriangulation(tri core.Triangulation) contourgeom.Triangulation {
	return contourgeom.Triangulation{
		X:         tri.X,
		Y:         tri.Y,
		Triangles: tri.Triangles,
		Mask:      tri.Mask,
	}
}

func triangleFinite(values []float64, triangle [3]int) bool {
	for _, index := range triangle {
		if index < 0 || index >= len(values) || math.IsNaN(values[index]) || math.IsInf(values[index], 0) {
			return false
		}
	}
	return true
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func resolveErrorRange(symmetric, lower, upper []float64, index int) (float64, float64) {
	low := resolveError(symmetric, index)
	high := low
	if len(lower) > 0 {
		low = resolveError(lower, index)
	}
	if len(upper) > 0 {
		high = resolveError(upper, index)
	}
	return low, high
}

func resolveError(values []float64, index int) float64 {
	switch {
	case len(values) == 0:
		return 0
	case len(values) == 1:
		return math.Abs(values[0])
	case index < len(values):
		return math.Abs(values[index])
	default:
		return 0
	}
}

func cloneRenderColors(colors []render.Color) []render.Color {
	return append([]render.Color(nil), colors...)
}

func finiteRange(values []float64) (float64, float64) {
	minValue := math.Inf(1)
	maxValue := math.Inf(-1)
	for _, value := range values {
		if !isFinite(value) {
			continue
		}
		minValue = math.Min(minValue, value)
		maxValue = math.Max(maxValue, value)
	}
	if math.IsInf(minValue, 1) || math.IsInf(maxValue, -1) {
		return 0, 1
	}
	if minValue == maxValue {
		return minValue, minValue + 1
	}
	return minValue, maxValue
}

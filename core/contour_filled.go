package core

import (
	"sort"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// contourExtendedLevels returns the level boundaries used for filled-band
// geometry, inserting sentinel extremes for the requested extend mode so the
// under/over bands are generated. The public levels are left untouched; the
// colormap norm still spans the real first/last level, so the sentinel band
// midpoints map to the colormap's under/over colors.
func contourExtendedLevels(levels []float64, extend string) []float64 {
	extendMin := extend == "min" || extend == "both"
	extendMax := extend == "max" || extend == "both"
	if !extendMin && !extendMax {
		return levels
	}
	out := make([]float64, 0, len(levels)+2)
	if extendMin {
		out = append(out, -1e250)
	}
	out = append(out, levels...)
	if extendMax {
		out = append(out, 1e250)
	}
	return out
}

func contourBandPolygons(tri Triangulation, values, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color, []string) {
	polygons, colors, hatches, _ := contourBandPolygonsBands(tri, values, levels, opt, mapping, alpha)
	return polygons, colors, hatches
}

func contourBandPolygonsBands(tri Triangulation, values, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color, []string, []int) {
	polygons := [][]geom.Pt{}
	colors := []render.Color{}
	hatches := []string{}
	bands := []int{}
	for levelIdx := 0; levelIdx+1 < len(levels); levelIdx++ {
		low := levels[levelIdx]
		high := levels[levelIdx+1]
		color := contourBandColor(low, high, levelIdx, opt, mapping, alpha)
		hatch := contourBandHatch(opt.Hatches, levelIdx)
		for triIdx, triangle := range tri.Triangles {
			if tri.Masked(triIdx) {
				continue
			}
			polygon := triangleBandPolygon(
				[3]geom.Pt{tri.Point(triangle[0]), tri.Point(triangle[1]), tri.Point(triangle[2])},
				[3]float64{values[triangle[0]], values[triangle[1]], values[triangle[2]]},
				low,
				high,
			)
			if len(polygon) < 3 {
				continue
			}
			polygons = append(polygons, polygon)
			colors = append(colors, color)
			hatches = append(hatches, hatch)
			bands = append(bands, levelIdx)
		}
	}
	return polygons, colors, hatches, bands
}

func contourGridBandPolygons(x, y []float64, data [][]float64, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color, []string) {
	polygons, colors, hatches, _ := contourGridBandPolygonsBands(x, y, data, levels, opt, mapping, alpha)
	return polygons, colors, hatches
}

func contourGridBandPolygonsBands(x, y []float64, data [][]float64, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color, []string, []int) {
	rows := len(data)
	if rows < 2 || len(x) < 2 || len(y) < 2 || len(levels) < 2 {
		return nil, nil, nil, nil
	}
	cols := len(data[0])
	if cols < 2 || len(x) < cols || len(y) < rows {
		return nil, nil, nil, nil
	}
	for row := 1; row < rows; row++ {
		if len(data[row]) != cols {
			return nil, nil, nil, nil
		}
	}

	polygons := [][]geom.Pt{}
	colors := []render.Color{}
	hatches := []string{}
	bands := []int{}
	for levelIdx := 0; levelIdx+1 < len(levels); levelIdx++ {
		low := levels[levelIdx]
		high := levels[levelIdx+1]
		color := contourBandColor(low, high, levelIdx, opt, mapping, alpha)
		hatch := contourBandHatch(opt.Hatches, levelIdx)
		for row := 0; row+1 < rows; row++ {
			for col := 0; col+1 < cols; col++ {
				cellPolygons := contourCellBandPolygonsCornerMask(
					[4]geom.Pt{
						{X: x[col], Y: y[row]},
						{X: x[col+1], Y: y[row]},
						{X: x[col+1], Y: y[row+1]},
						{X: x[col], Y: y[row+1]},
					},
					[4]float64{
						data[row][col],
						data[row][col+1],
						data[row+1][col+1],
						data[row+1][col],
					},
					low,
					high,
					opt.CornerMask != nil && *opt.CornerMask,
				)
				for _, polygon := range cellPolygons {
					if len(polygon) < 3 {
						continue
					}
					polygons = append(polygons, polygon)
					colors = append(colors, color)
					hatches = append(hatches, hatch)
					bands = append(bands, levelIdx)
				}
			}
		}
	}
	return polygons, colors, hatches, bands
}

func contourCellBandPolygonsCornerMask(points [4]geom.Pt, values [4]float64, low, high float64, cornerMask bool) [][]geom.Pt {
	finite := 0
	var trianglePoints [3]geom.Pt
	var triangleValues [3]float64
	for i, value := range values {
		if !isFinite(value) {
			continue
		}
		if finite < len(trianglePoints) {
			trianglePoints[finite] = points[i]
			triangleValues[finite] = value
		}
		finite++
	}
	if finite == 4 {
		return contourCellBandPolygons(points, values, low, high)
	}
	if !cornerMask || finite != 3 {
		return nil
	}
	polygon := triangleBandPolygon(trianglePoints, triangleValues, low, high)
	if len(polygon) < 3 {
		return nil
	}
	return [][]geom.Pt{polygon}
}

// contourBandCompoundPaths groups per-cell band polygons into one compound path
// per band so hatch patterns tile continuously across the whole filled region
// (matching Matplotlib, which emits a single path per contourf level). It
// returns the compound paths with their per-band face color and hatch. Bands
// are emitted in ascending order; empty bands are skipped.
func contourBandCompoundPaths(polygons [][]geom.Pt, colors []render.Color, hatches []string, bands []int) ([]geom.Path, []render.Color, []string) {
	if len(polygons) == 0 || len(bands) != len(polygons) {
		return nil, nil, nil
	}
	order := []int{}
	byBand := map[int][]int{}
	for i, band := range bands {
		if _, seen := byBand[band]; !seen {
			order = append(order, band)
		}
		byBand[band] = append(byBand[band], i)
	}
	sort.Ints(order)

	paths := make([]geom.Path, 0, len(order))
	faceColors := make([]render.Color, 0, len(order))
	bandHatches := make([]string, 0, len(order))
	for _, band := range order {
		indices := byBand[band]
		var compound geom.Path
		for _, idx := range indices {
			compound = appendPolygonSubpath(compound, polygons[idx])
		}
		if len(compound.C) == 0 {
			continue
		}
		paths = append(paths, compound)
		faceColors = append(faceColors, colors[indices[0]])
		bandHatches = append(bandHatches, hatches[indices[0]])
	}
	return paths, faceColors, bandHatches
}

// appendPolygonSubpath appends a closed polygon as a subpath of dst.
func appendPolygonSubpath(dst geom.Path, polygon []geom.Pt) geom.Path {
	sub := polygonPath(polygon, true)
	dst.C = append(dst.C, sub.C...)
	dst.V = append(dst.V, sub.V...)
	return dst
}

// contourBandHatch returns the hatch for band levelIdx, cycling the hatch list
// (Matplotlib cycles hatches per filled region). An empty list yields "".
func contourBandHatch(hatches []string, levelIdx int) string {
	if len(hatches) == 0 {
		return ""
	}
	return hatches[levelIdx%len(hatches)]
}

// applyContourHatchStyle sets the default hatch color and width on a filled
// contour collection when hatches are present, mirroring Matplotlib's
// rcParams["hatch.color"]/["hatch.linewidth"]. The defaults come from rc.Hatch
// (linewidth in points converted to device pixels). It is a no-op when no hatch
// pattern is set, so unhatched contourf keeps the fast batched draw path.
func applyContourHatchStyle(fills *PolyCollection, hatches []string, rc *style.RC) {
	if fills == nil || rc == nil || !contourHatchesUsed(hatches) {
		return
	}
	fills.HatchColor = rc.Hatch.Color
	fills.HatchWidth = rc.Hatch.LineWidth // points; converted at the collection hatch sink
}

func contourHatchesUsed(hatches []string) bool {
	for _, h := range hatches {
		if h != "" {
			return true
		}
	}
	return false
}

// contourFillGeometry chooses the filled-contour geometry representation. When
// hatches are present it merges each band's per-cell polygons into a single
// compound path so the hatch tiles continuously (Matplotlib's one-path-per-band
// model); otherwise it keeps the per-cell polygons for the existing fast,
// seam-free solid-fill path. Exactly one of paths/polys is non-empty.
func contourFillGeometry(polygons [][]geom.Pt, faceColors []render.Color, hatches []string, bands []int) (paths []geom.Path, polys [][]geom.Pt, colors []render.Color, hatchOut []string) {
	if contourHatchesUsed(hatches) {
		paths, colors, hatchOut = contourBandCompoundPaths(polygons, faceColors, hatches, bands)
		return paths, nil, colors, hatchOut
	}
	return nil, polygons, faceColors, hatches
}

func contourCellBandPolygons(points [4]geom.Pt, values [4]float64, low, high float64) [][]geom.Pt {
	if polygons := contourSaddleBandPolygons(points, values, low, high); len(polygons) > 0 {
		return polygons
	}
	polygon := contourCellBandPolygon(points, values, low, high)
	if len(polygon) < 3 {
		return nil
	}
	return [][]geom.Pt{polygon}
}

func contourCellBandPolygon(points [4]geom.Pt, values [4]float64, low, high float64) []geom.Pt {
	polygon := []contourVertex{
		{Point: points[0], Value: values[0]},
		{Point: points[1], Value: values[1]},
		{Point: points[2], Value: values[2]},
		{Point: points[3], Value: values[3]},
	}
	polygon = clipContourPolygonMin(polygon, low)
	if len(polygon) < 3 {
		return nil
	}
	polygon = clipContourPolygonMax(polygon, high)
	if len(polygon) < 3 {
		return nil
	}
	out := make([]geom.Pt, len(polygon))
	for i, vertex := range polygon {
		out[i] = vertex.Point
	}
	out = rotateContourPolygonToMatplotlibStart(out)
	if contourPolygonHasConsecutiveDuplicate(out) && !contourPolygonClosed(out) {
		out = append(out, out[0])
	}
	return out
}

func contourSaddleBandPolygons(points [4]geom.Pt, values [4]float64, low, high float64) [][]geom.Pt {
	inBand := [4]bool{}
	for i, value := range values {
		if !isFinite(value) {
			return nil
		}
		inBand[i] = value >= low && value <= high
	}
	if inBand[0] != inBand[2] || inBand[1] != inBand[3] || inBand[0] == inBand[1] {
		return nil
	}

	polygons := [][]geom.Pt{}
	for i, inside := range inBand {
		if !inside {
			continue
		}
		prev := (i + 3) % 4
		next := (i + 1) % 4
		if !contourBandOutsideSameSide(values[prev], values[next], low, high) {
			return nil
		}
		nextPoint, nextOK := contourBandBoundaryIntersection(points[i], values[i], points[next], values[next], low, high)
		prevPoint, prevOK := contourBandBoundaryIntersection(points[i], values[i], points[prev], values[prev], low, high)
		if !nextOK || !prevOK {
			continue
		}
		polygon := rotateContourPolygonToMatplotlibStart([]geom.Pt{points[i], nextPoint, prevPoint})
		if len(polygon) > 0 {
			polygon = append(polygon, polygon[0])
		}
		polygons = append(polygons, polygon)
	}
	return polygons
}

func contourBandOutsideSameSide(a, b, low, high float64) bool {
	return (a < low && b < low) || (a > high && b > high)
}

func contourBandBoundaryIntersection(insidePoint geom.Pt, insideValue float64, outsidePoint geom.Pt, outsideValue float64, low, high float64) (geom.Pt, bool) {
	if !isFinite(insideValue) || !isFinite(outsideValue) || insideValue == outsideValue {
		return geom.Pt{}, false
	}
	threshold := low
	if outsideValue > high {
		threshold = high
	}
	if outsideValue >= low && outsideValue <= high {
		return geom.Pt{}, false
	}
	t := (threshold - insideValue) / (outsideValue - insideValue)
	if t < 0 || t > 1 {
		return geom.Pt{}, false
	}
	return interpolatePoint(insidePoint, outsidePoint, t), true
}

type contourVertex struct {
	Point geom.Pt
	Value float64
}

func triangleBandPolygon(points [3]geom.Pt, values [3]float64, low, high float64) []geom.Pt {
	polygon := []contourVertex{
		{Point: points[0], Value: values[0]},
		{Point: points[1], Value: values[1]},
		{Point: points[2], Value: values[2]},
	}
	polygon = clipContourPolygonMin(polygon, low)
	if len(polygon) < 3 {
		return nil
	}
	polygon = clipContourPolygonMax(polygon, high)
	if len(polygon) < 3 {
		return nil
	}
	out := make([]geom.Pt, len(polygon))
	for i, vertex := range polygon {
		out[i] = vertex.Point
	}
	return out
}

func rotateContourPolygonToMatplotlibStart(points []geom.Pt) []geom.Pt {
	if len(points) == 0 {
		return points
	}
	start := 0
	for i := 1; i < len(points); i++ {
		if points[i].X > points[start].X ||
			(points[i].X == points[start].X && points[i].Y < points[start].Y) {
			start = i
		}
	}
	if start == 0 {
		return points
	}
	out := make([]geom.Pt, 0, len(points))
	out = append(out, points[start:]...)
	out = append(out, points[:start]...)
	return out
}

func contourPolygonHasConsecutiveDuplicate(points []geom.Pt) bool {
	for i := 1; i < len(points); i++ {
		if points[i] == points[i-1] {
			return true
		}
	}
	return false
}

func contourPolygonClosed(points []geom.Pt) bool {
	return len(points) > 1 && points[0] == points[len(points)-1]
}

func clipContourPolygonMin(polygon []contourVertex, threshold float64) []contourVertex {
	return clipContourPolygon(polygon, func(value float64) bool {
		return value >= threshold
	}, threshold)
}

func clipContourPolygonMax(polygon []contourVertex, threshold float64) []contourVertex {
	return clipContourPolygon(polygon, func(value float64) bool {
		return value <= threshold
	}, threshold)
}

func clipContourPolygon(polygon []contourVertex, inside func(float64) bool, threshold float64) []contourVertex {
	if len(polygon) == 0 {
		return nil
	}
	out := make([]contourVertex, 0, len(polygon)+2)
	prev := polygon[len(polygon)-1]
	prevInside := inside(prev.Value)
	for _, curr := range polygon {
		currInside := inside(curr.Value)
		if currInside != prevInside && curr.Value != prev.Value {
			t := (threshold - prev.Value) / (curr.Value - prev.Value)
			out = append(out, contourVertex{
				Point: interpolatePoint(prev.Point, curr.Point, t),
				Value: threshold,
			})
		}
		if currInside {
			out = append(out, curr)
		}
		prev = curr
		prevInside = currInside
	}
	return out
}

func contourBandColor(low, high float64, idx int, opt ContourOptions, mapping ScalarMapInfo, alpha float64) render.Color {
	if len(opt.Colors) > 0 {
		color := opt.Colors[idx%len(opt.Colors)]
		color.A *= alpha
		return color
	}
	if opt.Color != nil {
		color := *opt.Color
		color.A *= alpha
		return color
	}
	return mapping.Color((low+high)*0.5, alpha)
}

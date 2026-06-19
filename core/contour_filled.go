package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func contourBandPolygons(tri Triangulation, values, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color) {
	polygons := [][]geom.Pt{}
	colors := []render.Color{}
	for levelIdx := 0; levelIdx+1 < len(levels); levelIdx++ {
		low := levels[levelIdx]
		high := levels[levelIdx+1]
		color := contourBandColor(low, high, levelIdx, opt, mapping, alpha)
		for triIdx, triangle := range tri.Triangles {
			if tri.masked(triIdx) {
				continue
			}
			polygon := triangleBandPolygon(
				[3]geom.Pt{tri.point(triangle[0]), tri.point(triangle[1]), tri.point(triangle[2])},
				[3]float64{values[triangle[0]], values[triangle[1]], values[triangle[2]]},
				low,
				high,
			)
			if len(polygon) < 3 {
				continue
			}
			polygons = append(polygons, polygon)
			colors = append(colors, color)
		}
	}
	return polygons, colors
}

func contourGridBandPolygons(x, y []float64, data [][]float64, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color) {
	rows := len(data)
	if rows < 2 || len(x) < 2 || len(y) < 2 || len(levels) < 2 {
		return nil, nil
	}
	cols := len(data[0])
	if cols < 2 || len(x) < cols || len(y) < rows {
		return nil, nil
	}
	for row := 1; row < rows; row++ {
		if len(data[row]) != cols {
			return nil, nil
		}
	}

	polygons := [][]geom.Pt{}
	colors := []render.Color{}
	for levelIdx := 0; levelIdx+1 < len(levels); levelIdx++ {
		low := levels[levelIdx]
		high := levels[levelIdx+1]
		color := contourBandColor(low, high, levelIdx, opt, mapping, alpha)
		for row := 0; row+1 < rows; row++ {
			for col := 0; col+1 < cols; col++ {
				cellPolygons := contourCellBandPolygons(
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
				)
				for _, polygon := range cellPolygons {
					if len(polygon) < 3 {
						continue
					}
					polygons = append(polygons, polygon)
					colors = append(colors, color)
				}
			}
		}
	}
	return polygons, colors
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

package core

import (
	"math"
	"sort"
)

func contourGridTriangulation(data [][]float64, opts []ContourOptions) (Triangulation, []float64, bool) {
	xCoords, yCoords, values, ok := contourGridCoordsValues(data, opts)
	if !ok {
		return Triangulation{}, nil, false
	}
	rows := len(data)
	cols := len(data[0])

	xPoints := make([]float64, 0, rows*cols)
	yPoints := make([]float64, 0, rows*cols)
	for yi := 0; yi < rows; yi++ {
		for xi := 0; xi < cols; xi++ {
			xPoints = append(xPoints, xCoords[xi])
			yPoints = append(yPoints, yCoords[yi])
		}
	}

	triangles := make([][3]int, 0, (rows-1)*(cols-1)*2)
	mask := make([]bool, 0, (rows-1)*(cols-1)*2)
	index := func(row, col int) int { return row*cols + col }
	for yi := 0; yi+1 < rows; yi++ {
		for xi := 0; xi+1 < cols; xi++ {
			p00 := index(yi, xi)
			p10 := index(yi, xi+1)
			p01 := index(yi+1, xi)
			p11 := index(yi+1, xi+1)

			triangles = append(triangles, [3]int{p00, p10, p11})
			mask = append(mask, !triangleFinite(values, [3]int{p00, p10, p11}))
			triangles = append(triangles, [3]int{p00, p11, p01})
			mask = append(mask, !triangleFinite(values, [3]int{p00, p11, p01}))
		}
	}

	return Triangulation{
		X:         xPoints,
		Y:         yPoints,
		Triangles: triangles,
		Mask:      mask,
	}, values, true
}

func contourGridCoordsValues(data [][]float64, opts []ContourOptions) ([]float64, []float64, []float64, bool) {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok || rows < 2 || cols < 2 {
		return nil, nil, nil, false
	}

	var opt ContourOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	xCoords := resolvedContourCoords(cols, opt.X, opt.XEdges)
	yCoords := resolvedContourCoords(rows, opt.Y, opt.YEdges)
	if len(xCoords) != cols || len(yCoords) != rows {
		return nil, nil, nil, false
	}

	values := make([]float64, 0, rows*cols)
	for yi := 0; yi < rows; yi++ {
		if len(data[yi]) != cols {
			return nil, nil, nil, false
		}
		values = append(values, data[yi]...)
	}
	return xCoords, yCoords, values, true
}

func triangleFinite(values []float64, tri [3]int) bool {
	return isFinite(values[tri[0]]) && isFinite(values[tri[1]]) && isFinite(values[tri[2]])
}

func resolvedContourCoords(size int, coords, edges []float64) []float64 {
	switch {
	case len(coords) == size:
		return append([]float64(nil), coords...)
	case len(edges) == size:
		return append([]float64(nil), edges...)
	case len(edges) == size+1:
		out := make([]float64, size)
		for i := 0; i < size; i++ {
			out[i] = (edges[i] + edges[i+1]) * 0.5
		}
		return out
	default:
		out := make([]float64, size)
		for i := range out {
			out[i] = float64(i)
		}
		return out
	}
}

func contourLevels(values, explicit []float64, levelCount int, filled bool) []float64 {
	if len(explicit) > 0 {
		levels := make([]float64, 0, len(explicit))
		for _, level := range explicit {
			if isFinite(level) {
				levels = append(levels, level)
			}
		}
		sort.Float64s(levels)
		return dedupeFloat64(levels)
	}

	if levelCount <= 0 {
		levelCount = 7
	}
	if filled && levelCount < 2 {
		levelCount = 2
	}

	minValue, maxValue := finiteRange(values)
	if !isFinite(minValue) || !isFinite(maxValue) {
		return nil
	}
	if minValue == maxValue {
		if filled {
			return []float64{minValue, minValue + 1}
		}
		return []float64{minValue}
	}

	levels := contourLocatorLevels(minValue, maxValue, levelCount, filled)
	if len(levels) > 0 {
		return levels
	}

	levels = make([]float64, levelCount)
	step := (maxValue - minValue) / float64(levelCount-1)
	for i := range levels {
		levels[i] = minValue + float64(i)*step
	}
	return levels
}

func contourLocatorLevels(minValue, maxValue float64, levelCount int, filled bool) []float64 {
	// Match matplotlib's _ensure_locator_exists: MaxNLocator(N + 1, min_n_ticks=1).
	// For levels=N the locator is asked for N+1 intervals so the resulting "nice" step
	// is roughly (zmax-zmin)/(N+1) — matching ContourSet._autolev's tick layout.
	levels := (MaxNLocator{
		N:     levelCount + 1,
		Steps: []float64{1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10},
	}).Ticks(minValue, maxValue, 0)
	if len(levels) == 0 {
		return nil
	}

	out := levels[:0]
	for _, level := range levels {
		if !isFinite(level) {
			continue
		}
		out = append(out, level)
	}
	return dedupeFloat64(out)
}

func dedupeFloat64(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := []float64{values[0]}
	for _, value := range values[1:] {
		if math.Abs(value-out[len(out)-1]) <= 1e-12 {
			continue
		}
		out = append(out, value)
	}
	return out
}

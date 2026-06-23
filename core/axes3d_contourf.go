package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Contourf projects a structured z grid and emits filled contour bands.
func (a *Axes3D) Contourf(x, y []float64, z [][]float64, opts ...PlotOptions) *PolyCollection {
	opt := firstPlotOptions(opts)
	colorOverride := opt.Color != nil
	// matplotlib Axes3D.contourf forwards kwargs to Axes.contourf unchanged:
	// filled bands are opaque unless the caller passes alpha.
	alpha := 1.0
	label := ""
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	label = opt.Label
	limitsChanged := a.observe3DContourf(x, y, z, opt)

	paths, colors, scalarValues, zorder, mapping := a.projectedContourFillData(x, y, z, alpha, opt)
	if len(paths) == 0 {
		return nil
	}
	cmap := mapping.Colormap
	norm := mapping.Norm
	vMin := mapping.VMin
	vMax := mapping.VMax
	if colorOverride {
		scalarValues = nil
	}
	if colorOverride {
		cmap = ""
		norm = nil
		vMin = 0
		vMax = 0
	}

	collection := &PolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords:       Coords(CoordData),
				Label:        label,
				Alpha:        1,
				Colormap:     cmap,
				Norm:         norm,
				VMin:         vMin,
				VMax:         vMax,
				ScalarValues: scalarValues,
				z:            zorder,
				// matplotlib ContourSet: antialiased=None with filled=True
				// resolves to False, so band fills render with hard edges.
				Antialias: render.AntialiasOff,
			},
			Paths:      paths,
			FaceColors: colors,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			paths, colors, scalarValues, zorder, mapping := a.projectedContourFillData(x, y, z, alpha, opt)
			collection.Polygons = nil
			collection.Paths = paths
			collection.FaceColors = colors
			if colorOverride {
				collection.Colormap = ""
				collection.Norm = nil
				collection.VMin = 0
				collection.VMax = 0
				collection.ScalarValues = nil
			} else {
				collection.Colormap = mapping.Colormap
				collection.Norm = mapping.Norm
				collection.VMin = mapping.VMin
				collection.VMax = mapping.VMax
				collection.ScalarValues = scalarValues
			}
			collection.z = zorder
		}
	}, limitsChanged)
	return collection
}

// TriContourf projects filled contour bands over an explicit triangulated 3D mesh.
func (a *Axes3D) TriContourf(tri Triangulation, z []float64, opts ...PlotOptions) *PolyCollection {
	if a == nil || len(tri.X) == 0 {
		return nil
	}
	if err := tri.Validate(); err != nil || len(z) != len(tri.X) {
		return nil
	}
	var ok bool
	tri, ok = tri.EnsureTriangles()
	if !ok {
		return nil
	}

	opt := firstPlotOptions(opts)
	colorOverride := opt.Color != nil
	// matplotlib Axes3D.tricontourf forwards kwargs to Axes.tricontourf
	// unchanged: filled bands are opaque unless the caller passes alpha.
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	limitsChanged := a.observe3DTriContourf(tri, z, opt)

	paths, colors, scalarValues, zorder, mapping := a.projectedTriContourFillData(tri, z, alpha, opt)
	if len(paths) == 0 {
		return nil
	}
	cmap := mapping.Colormap
	norm := mapping.Norm
	vMin := mapping.VMin
	vMax := mapping.VMax
	if colorOverride {
		scalarValues = nil
	}
	if colorOverride {
		cmap = ""
		norm = nil
		vMin = 0
		vMax = 0
	}

	collection := &PolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords:       Coords(CoordData),
				Label:        opt.Label,
				Alpha:        1,
				Colormap:     cmap,
				Norm:         norm,
				VMin:         vMin,
				VMax:         vMax,
				ScalarValues: scalarValues,
				z:            zorder,
				// matplotlib ContourSet: antialiased=None with filled=True
				// resolves to False, so band fills render with hard edges.
				Antialias: render.AntialiasOff,
			},
			Paths:      paths,
			FaceColors: colors,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			paths, colors, scalarValues, zorder, mapping := a.projectedTriContourFillData(tri, z, alpha, opt)
			collection.Polygons = nil
			collection.Paths = paths
			collection.FaceColors = colors
			if colorOverride {
				collection.Colormap = ""
				collection.Norm = nil
				collection.VMin = 0
				collection.VMax = 0
				collection.ScalarValues = nil
			} else {
				collection.Colormap = mapping.Colormap
				collection.Norm = mapping.Norm
				collection.VMin = mapping.VMin
				collection.VMax = mapping.VMax
				collection.ScalarValues = scalarValues
			}
			collection.z = zorder
		}
	}, limitsChanged)
	return collection
}

func contourScalarMap(values, levels []float64, opt PlotOptions) ScalarMapInfo {
	mapping := resolvePlotScalarMap(values, opt)
	if len(levels) >= 2 && opt.VMin == nil && opt.VMax == nil {
		mapping.VMin = levels[0]
		mapping.VMax = levels[len(levels)-1]
		if mapping.Norm == nil {
			mapping.Norm = Normalize{VMin: mapping.VMin, VMax: mapping.VMax}
		}
	}
	return mapping
}

func contourLayerValues(levels []float64, mapping ScalarMapInfo) []float64 {
	if len(levels) < 2 {
		return nil
	}
	values := make([]float64, len(levels)-1)
	logScale := mapping.Norm != nil && mapping.Norm.NormName() == "log"
	for i := range values {
		low := levels[i]
		high := levels[i+1]
		if logScale && low > 0 && high > 0 {
			values[i] = math.Sqrt(low) * math.Sqrt(high)
		} else {
			values[i] = 0.5 * (low + high)
		}
	}
	return values
}

func (a *Axes3D) projectedContourFillData(x, y []float64, z [][]float64, alpha float64, opt PlotOptions) ([]geom.Path, []render.Color, []float64, float64, ScalarMapInfo) {
	rows, cols, ok := validate3DGridContourInput(x, y, z)
	if !ok {
		return nil, nil, nil, defaultPatchZ, ScalarMapInfo{}
	}
	values := flattenGridValues(z)
	levels := contourLevels(values, opt.Levels, opt.LevelCount, true)
	if len(levels) < 2 {
		return nil, nil, nil, defaultPatchZ, ScalarMapInfo{}
	}
	zdir := normalized3DDir(opt.ZDir)
	mapping := contourScalarMap(values, levels, opt)
	layerValues := contourLayerValues(levels, mapping)
	collectionDepth := math.Inf(1)
	paths := make([]geom.Path, 0, len(levels)-1)
	colors := make([]render.Color, 0, len(levels)-1)

	var tri Triangulation
	var rotatedValues []float64
	if zdir != "z" {
		var ok bool
		tri, rotatedValues, ok = rotatedContourTriangulation(x[:cols], y[:rows], z, zdir)
		if !ok {
			return nil, nil, nil, defaultPatchZ, ScalarMapInfo{}
		}
	}

	for levelIdx := 0; levelIdx+1 < len(levels); levelIdx++ {
		low := levels[levelIdx]
		high := levels[levelIdx+1]
		bandLevel := layerValues[levelIdx]
		planeLevel := contourPlaneLevel(bandLevel, opt.Offset)
		var rawPolygons [][]geom.Pt
		if zdir == "z" {
			rawPolygons = contourGridBandPolygonsForLevel(x[:cols], y[:rows], z, low, high)
		} else {
			rawPolygons = contourTriBandPolygons(tri, rotatedValues, low, high)
		}
		if len(rawPolygons) == 0 {
			continue
		}

		projectedPolygons := make([][]geom.Pt, 0, len(rawPolygons))
		for _, polygon := range rawPolygons {
			if len(polygon) < 3 {
				continue
			}
			rawPolygon3D := contourPolyline3D(polygon, planeLevel, zdir)
			if opt.AxLimClip && !a.polygonWithin3DViewLimits(rawPolygon3D) {
				continue
			}
			projected := make([]geom.Pt, len(rawPolygon3D))
			for i, point3D := range rawPolygon3D {
				projectedPt, zDepth := a.projectPointDepth(point3D[0], point3D[1], point3D[2])
				projected[i] = projectedPt
				if zDepth < collectionDepth {
					collectionDepth = zDepth
				}
			}
			projectedPolygons = append(projectedPolygons, projected)
		}
		if len(projectedPolygons) == 0 {
			continue
		}
		path := contourBoundaryPath(projectedPolygons)
		if len(path.C) == 0 {
			path = contourPolygonsPath(projectedPolygons)
		}
		if len(path.C) == 0 {
			continue
		}

		color := mapping.Color(bandLevel, alpha)
		if opt.Color != nil {
			color = *opt.Color
			color.A *= alpha
		}
		paths = append(paths, path)
		colors = append(colors, color)
	}
	if len(paths) == 0 {
		return nil, nil, nil, defaultPatchZ, ScalarMapInfo{}
	}
	return paths, colors, layerValues, computed3DCollectionZ(collectionDepth), mapping
}

func (a *Axes3D) projectedTriContourFillData(tri Triangulation, z []float64, alpha float64, opt PlotOptions) ([]geom.Path, []render.Color, []float64, float64, ScalarMapInfo) {
	if a == nil {
		return nil, nil, nil, defaultPatchZ, ScalarMapInfo{}
	}
	zdir := normalized3DDir(opt.ZDir)
	rotatedTri, rotatedValues, ok := rotatedTriangulation3D(tri, z, zdir)
	if !ok {
		return nil, nil, nil, defaultPatchZ, ScalarMapInfo{}
	}
	levels := contourLevels(rotatedValues, opt.Levels, opt.LevelCount, true)
	if len(levels) < 2 {
		return nil, nil, nil, defaultPatchZ, ScalarMapInfo{}
	}
	mapping := contourScalarMap(rotatedValues, levels, opt)
	layerValues := contourLayerValues(levels, mapping)
	collectionDepth := math.Inf(1)
	paths := make([]geom.Path, 0, len(levels)-1)
	colors := make([]render.Color, 0, len(levels)-1)

	for levelIdx := 0; levelIdx+1 < len(levels); levelIdx++ {
		low := levels[levelIdx]
		high := levels[levelIdx+1]
		bandLevel := layerValues[levelIdx]
		planeLevel := contourPlaneLevel(bandLevel, opt.Offset)
		rawPolygons := contourTriBandPolygons(rotatedTri, rotatedValues, low, high)
		if len(rawPolygons) == 0 {
			continue
		}

		projectedPolygons := make([][]geom.Pt, 0, len(rawPolygons))
		for _, polygon := range rawPolygons {
			if len(polygon) < 3 {
				continue
			}
			rawPolygon3D := contourPolyline3D(polygon, planeLevel, zdir)
			if opt.AxLimClip && !a.polygonWithin3DViewLimits(rawPolygon3D) {
				continue
			}
			projected := make([]geom.Pt, len(rawPolygon3D))
			for i, point3D := range rawPolygon3D {
				projectedPt, zDepth := a.projectPointDepth(point3D[0], point3D[1], point3D[2])
				projected[i] = projectedPt
				if zDepth < collectionDepth {
					collectionDepth = zDepth
				}
			}
			projectedPolygons = append(projectedPolygons, projected)
		}
		if len(projectedPolygons) == 0 {
			continue
		}
		path := contourBoundaryPath(projectedPolygons)
		if len(path.C) == 0 {
			path = contourPolygonsPath(projectedPolygons)
		}
		if len(path.C) == 0 {
			continue
		}

		color := mapping.Color(bandLevel, alpha)
		if opt.Color != nil {
			color = *opt.Color
			color.A *= alpha
		}
		paths = append(paths, path)
		colors = append(colors, color)
	}
	if len(paths) == 0 {
		return nil, nil, nil, defaultPatchZ, ScalarMapInfo{}
	}
	return paths, colors, layerValues, computed3DCollectionZ(collectionDepth), mapping
}

func validate3DGridContourInput(x, y []float64, z [][]float64) (rows, cols int, ok bool) {
	if len(z) == 0 {
		return 0, 0, false
	}
	rows = len(z)
	cols = len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return 0, 0, false
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return 0, 0, false
		}
	}
	return rows, cols, true
}

func contourGridBandPolygonsForLevel(x, y []float64, data [][]float64, low, high float64) [][]geom.Pt {
	rows := len(data)
	if rows < 2 || len(x) < 2 || len(y) < 2 {
		return nil
	}
	cols := len(data[0])
	polygons := make([][]geom.Pt, 0)
	for row := 0; row+1 < rows; row++ {
		for col := 0; col+1 < cols; col++ {
			polygons = append(polygons, contourCellBandPolygons(
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
			)...)
		}
	}
	return polygons
}

func contourTriBandPolygons(tri Triangulation, values []float64, low, high float64) [][]geom.Pt {
	polygons := make([][]geom.Pt, 0)
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
		if len(polygon) >= 3 {
			polygons = append(polygons, polygon)
		}
	}
	return polygons
}

func rotatedContourTriangulation(x, y []float64, z [][]float64, zdir string) (Triangulation, []float64, bool) {
	rows, cols, ok := validate3DGridContourInput(x, y, z)
	if !ok {
		return Triangulation{}, nil, false
	}
	pointsX := make([]float64, 0, rows*cols)
	pointsY := make([]float64, 0, rows*cols)
	values := make([]float64, 0, rows*cols)
	triangles := make([][3]int, 0, (rows-1)*(cols-1)*2)
	mask := make([]bool, 0, (rows-1)*(cols-1)*2)
	index := func(row, col int) int { return row*cols + col }

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			p := rotate3DPointAxes(x[col], y[row], z[row][col], zdir)
			pointsX = append(pointsX, p[0])
			pointsY = append(pointsY, p[1])
			values = append(values, p[2])
		}
	}
	for row := 0; row+1 < rows; row++ {
		for col := 0; col+1 < cols; col++ {
			p00 := index(row, col)
			p10 := index(row, col+1)
			p01 := index(row+1, col)
			p11 := index(row+1, col+1)
			t0 := [3]int{p00, p10, p11}
			t1 := [3]int{p00, p11, p01}
			triangles = append(triangles, t0, t1)
			mask = append(mask, !triangleFinite(values, t0), !triangleFinite(values, t1))
		}
	}
	return Triangulation{X: pointsX, Y: pointsY, Triangles: triangles, Mask: mask}, values, true
}

func rotatedTriangulation3D(tri Triangulation, z []float64, zdir string) (Triangulation, []float64, bool) {
	if err := tri.Validate(); err != nil || len(z) != len(tri.X) {
		return Triangulation{}, nil, false
	}
	rotated := Triangulation{
		X:         make([]float64, len(tri.X)),
		Y:         make([]float64, len(tri.Y)),
		Triangles: append([][3]int(nil), tri.Triangles...),
		Mask:      append([]bool(nil), tri.Mask...),
	}
	values := make([]float64, len(z))
	for i := range z {
		p := rotate3DPointAxes(tri.X[i], tri.Y[i], z[i], zdir)
		rotated.X[i] = p[0]
		rotated.Y[i] = p[1]
		values[i] = p[2]
	}
	return rotated, values, true
}

func (a *Axes3D) projectedContourFloorPolygons(x, y []float64, z [][]float64, alpha float64, colorOverride *render.Color, levelCount int, explicitLevels []float64, offset *float64) ([][]geom.Pt, []render.Color, float64) {
	if a == nil || len(z) == 0 {
		return nil, nil, defaultPatchZ
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return nil, nil, defaultPatchZ
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return nil, nil, defaultPatchZ
		}
	}

	zMin, zMax := zGridRange(z)
	if zMin == zMax {
		zMin -= 0.5
		zMax += 0.5
	}
	floorZ := zMin - 0.2*(zMax-zMin)
	if offset != nil && isFinite(*offset) {
		floorZ = *offset
	}

	values := flattenGridValues(z)
	levels := contourLevels(values, explicitLevels, levelCount, true)
	if len(levels) < 2 {
		return nil, nil, defaultPatchZ
	}

	mapping := resolveScalarMapValues(values, "viridis", nil, nil)
	mapping.VMin = levels[0]
	mapping.VMax = levels[len(levels)-1]
	opt := ContourOptions{}
	if colorOverride != nil {
		opt.Color = colorOverride
	}
	rawPolygons, colors, _ := contourGridBandPolygons(x[:cols], y[:rows], z, levels, opt, mapping, alpha)
	if len(rawPolygons) == 0 {
		return nil, nil, defaultPatchZ
	}

	polygons := make([][]geom.Pt, 0, len(rawPolygons))
	projectedColors := make([]render.Color, 0, len(colors))
	collectionDepth := math.Inf(1)
	for i, polygon := range rawPolygons {
		if len(polygon) < 3 {
			continue
		}
		projected := make([]geom.Pt, len(polygon))
		for j, pt := range polygon {
			projectedPt, zDepth := a.projectPointDepth(pt.X, pt.Y, floorZ)
			projected[j] = projectedPt
			if zDepth < collectionDepth {
				collectionDepth = zDepth
			}
		}
		polygons = append(polygons, projected)
		if i < len(colors) {
			projectedColors = append(projectedColors, colors[i])
		} else if colorOverride != nil {
			color := *colorOverride
			color.A *= alpha
			projectedColors = append(projectedColors, color)
		} else {
			projectedColors = append(projectedColors, mapping.Color(0, 1))
		}
	}
	return polygons, projectedColors, computed3DCollectionZ(collectionDepth)
}

func zGridRange(z [][]float64) (float64, float64) {
	minVal, maxVal := 0.0, 0.0
	first := true
	for _, row := range z {
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				continue
			}
			if first {
				minVal, maxVal = v, v
				first = false
				continue
			}
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
	}
	return minVal, maxVal
}

func compoundContourPaths(polygons [][]geom.Pt, colors []render.Color) ([]geom.Path, []render.Color) {
	paths := make([]geom.Path, 0)
	pathColors := make([]render.Color, 0)
	var current [][]geom.Pt
	var currentColor render.Color
	haveCurrent := false
	flush := func() {
		if !haveCurrent || len(current) == 0 {
			return
		}
		path := contourBoundaryPath(current)
		if len(path.C) == 0 {
			path = contourPolygonsPath(current)
		}
		if len(path.C) > 0 {
			paths = append(paths, path)
			pathColors = append(pathColors, currentColor)
		}
		current = nil
		haveCurrent = false
	}
	for i, polygon := range polygons {
		if len(polygon) < 3 {
			continue
		}
		color := render.Color{}
		if i < len(colors) {
			color = colors[i]
		}
		if haveCurrent && color != currentColor {
			flush()
		}
		if !haveCurrent {
			currentColor = color
			haveCurrent = true
		}
		current = append(current, polygon)
	}
	flush()
	return paths, pathColors
}

type contourPointKey struct {
	x int64
	y int64
}

type contourEdgeKey struct {
	from contourPointKey
	to   contourPointKey
}

type contourBoundaryEdge struct {
	id      int
	from    geom.Pt
	to      geom.Pt
	fromKey contourPointKey
	toKey   contourPointKey
}

func contourPolygonsPath(polygons [][]geom.Pt) geom.Path {
	var path geom.Path
	for _, polygon := range polygons {
		if len(polygon) < 3 {
			continue
		}
		for i, pt := range polygon {
			if i == 0 {
				path.MoveTo(pt)
			} else {
				path.LineTo(pt)
			}
		}
		path.Close()
	}
	return path
}

func contourBoundaryPath(polygons [][]geom.Pt) geom.Path {
	edgesByKey := map[contourEdgeKey]contourBoundaryEdge{}
	ordered := make([]contourEdgeKey, 0)
	nextID := 0
	for _, polygon := range polygons {
		if len(polygon) < 3 {
			continue
		}
		for i, from := range polygon {
			to := polygon[(i+1)%len(polygon)]
			fromKey := contourPathPointKey(from)
			toKey := contourPathPointKey(to)
			if fromKey == toKey {
				continue
			}
			key := contourEdgeKey{from: fromKey, to: toKey}
			reverse := contourEdgeKey{from: toKey, to: fromKey}
			if _, ok := edgesByKey[reverse]; ok {
				delete(edgesByKey, reverse)
				continue
			}
			edgesByKey[key] = contourBoundaryEdge{
				id:      nextID,
				from:    from,
				to:      to,
				fromKey: fromKey,
				toKey:   toKey,
			}
			ordered = append(ordered, key)
			nextID++
		}
	}
	if len(edgesByKey) == 0 {
		return geom.Path{}
	}

	adjacent := map[contourPointKey][]contourBoundaryEdge{}
	for _, edge := range edgesByKey {
		adjacent[edge.fromKey] = append(adjacent[edge.fromKey], edge)
	}

	used := map[int]bool{}
	var path geom.Path
	for _, key := range ordered {
		edge, ok := edgesByKey[key]
		if !ok || used[edge.id] {
			continue
		}
		loop := []geom.Pt{edge.from, edge.to}
		used[edge.id] = true
		startKey := edge.fromKey
		currentKey := edge.toKey
		closed := currentKey == startKey
		for !closed {
			next, ok := nextUnusedContourBoundaryEdge(adjacent[currentKey], used)
			if !ok {
				break
			}
			loop = append(loop, next.to)
			used[next.id] = true
			currentKey = next.toKey
			closed = currentKey == startKey
		}
		if !closed || len(loop) < 4 {
			continue
		}
		if contourPathPointKey(loop[len(loop)-1]) == contourPathPointKey(loop[0]) {
			loop = loop[:len(loop)-1]
		}
		if len(loop) < 3 {
			continue
		}
		for i, pt := range loop {
			if i == 0 {
				path.MoveTo(pt)
			} else {
				path.LineTo(pt)
			}
		}
		path.Close()
	}
	return path
}

func nextUnusedContourBoundaryEdge(edges []contourBoundaryEdge, used map[int]bool) (contourBoundaryEdge, bool) {
	for _, edge := range edges {
		if !used[edge.id] {
			return edge, true
		}
	}
	return contourBoundaryEdge{}, false
}

func contourPathPointKey(pt geom.Pt) contourPointKey {
	const scale = 1e9
	return contourPointKey{
		x: int64(math.Round(pt.X * scale)),
		y: int64(math.Round(pt.Y * scale)),
	}
}

func flattenGridValues(z [][]float64) []float64 {
	values := make([]float64, 0)
	for _, row := range z {
		values = append(values, row...)
	}
	return values
}

func (a *Axes3D) projectedContourFillPolygons(x, y []float64, z [][]float64, opt ContourOptions, levelCount int) ([][]geom.Pt, []render.Color) {
	tri, values, ok := a.projectedContourTriangulation(x, y, z)
	if !ok {
		return nil, nil
	}
	if levelCount <= 0 {
		levelCount = 7
	}

	levels := contourLevels(values, nil, levelCount, true)
	if len(levels) < 2 {
		return nil, nil
	}

	mapping := resolveScalarMapValues(values, "", nil, nil)
	mapping.VMin = levels[0]
	mapping.VMax = levels[len(levels)-1]
	polygons, polygonColors, _ := contourBandPolygons(tri, values, levels, opt, mapping, 1.0)
	if len(polygons) == 0 {
		return nil, nil
	}

	filteredPolygons := make([][]geom.Pt, 0, len(polygons))
	filteredColors := make([]render.Color, 0, len(polygonColors))
	for i, polygon := range polygons {
		if len(polygon) < 3 {
			continue
		}
		if i < len(polygonColors) {
			filteredColors = append(filteredColors, polygonColors[i])
			filteredPolygons = append(filteredPolygons, polygon)
		}
	}
	return filteredPolygons, filteredColors
}

func (a *Axes3D) projectedContourTriangulation(x, y []float64, z [][]float64) (Triangulation, []float64, bool) {
	if a == nil || len(z) == 0 {
		return Triangulation{}, nil, false
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return Triangulation{}, nil, false
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return Triangulation{}, nil, false
		}
	}

	points := make([]geom.Pt, 0, rows*cols)
	values := make([]float64, 0, rows*cols)
	triangles := make([][3]int, 0, (rows-1)*(cols-1)*2)
	mask := make([]bool, 0, (rows-1)*(cols-1)*2)
	index := func(row, col int) int { return row*cols + col }

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			p := a.ProjectPoint(x[col], y[row], z[row][col])
			points = append(points, p)
			values = append(values, z[row][col])
		}
	}
	for row := 0; row+1 < rows; row++ {
		for col := 0; col+1 < cols; col++ {
			p00 := index(row, col)
			p10 := index(row, col+1)
			p01 := index(row+1, col)
			p11 := index(row+1, col+1)
			triangles = append(triangles, [3]int{p00, p10, p11})
			mask = append(mask, !triangleFinite(values, [3]int{p00, p10, p11}))
			triangles = append(triangles, [3]int{p00, p11, p01})
			mask = append(mask, !triangleFinite(values, [3]int{p00, p11, p01}))
		}
	}

	tri := Triangulation{
		X:         make([]float64, len(points)),
		Y:         make([]float64, len(points)),
		Triangles: triangles,
		Mask:      mask,
	}
	for i, pt := range points {
		tri.X[i] = pt.X
		tri.Y[i] = pt.Y
	}
	return tri, values, true
}

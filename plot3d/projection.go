package plot3d

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type projectedScatterPoint struct {
	point geom.Pt
	depth float64
	index int
}

//nolint:gocritic // ScatterOptions is an immutable snapshot retained by redraw closures.
func reprojectScatter3D(scatter *core.Scatter2D, points []projectedScatterPoint, opt core.ScatterOptions) {
	if scatter == nil {
		return
	}
	sort.SliceStable(points, func(i, j int) bool {
		return points[i].depth > points[j].depth
	})
	scatter.XY = scatter.XY[:0]
	for _, point := range points {
		scatter.XY = append(scatter.XY, point.point)
	}

	if len(opt.Sizes) > 1 {
		scatter.Sizes = reorderScatterFloat64s(opt.Sizes, points)
	}
	if len(opt.ScalarValues) > 0 {
		scatter.ScalarValues = reorderScatterFloat64s(opt.ScalarValues, points)
		scatter.Colors = depthShadedScatterColors(scatterScalarColors(scatter), points)
	} else {
		scatter.ScalarValues = nil
		scatter.Colors = depthShadedScatterColors(scatterPointColors(scatter.Color, opt.Colors, points), points)
	}
	if scatter.EdgeColorsFace {
		scatter.EdgeColors = cloneRenderColors(scatter.Colors)
	} else {
		scatter.EdgeColors = depthShadedScatterColors(scatterPointColors(scatter.EdgeColor, opt.EdgeColors, points), points)
	}
}

//nolint:gocritic // ScatterOptions remains value-semantic while projected slices are replaced.
func scatterOptionsForProjected(opt core.ScatterOptions, points []projectedScatterPoint) core.ScatterOptions {
	filtered := opt
	if len(opt.Colors) > 1 {
		filtered.Colors = reorderScatterColors(opt.Colors, points)
	}
	if len(opt.EdgeColors) > 1 {
		filtered.EdgeColors = reorderScatterColors(opt.EdgeColors, points)
	}
	if len(opt.ScalarValues) > 1 {
		filtered.ScalarValues = reorderScatterFloat64s(opt.ScalarValues, points)
	}
	if len(opt.Sizes) > 1 {
		filtered.Sizes = reorderScatterFloat64s(opt.Sizes, points)
	}
	return filtered
}

func reprojectLine3D(line *core.Line2D, points []geom.Pt) {
	if line == nil {
		return
	}
	line.XY = append(line.XY[:0], points...)
}

func (a *Axes3D) projectedScatterData(x, y, z []float64, axlimClip ...bool) []projectedScatterPoint {
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if len(z) < n {
		n = len(z)
	}
	clip := len(axlimClip) > 0 && axlimClip[0]
	points := make([]projectedScatterPoint, 0, n)
	for i := 0; i < n; i++ {
		if !isFinite3D(x[i], y[i], z[i]) {
			continue
		}
		if clip && !a.pointWithin3DViewLimits(vec3{x[i], y[i], z[i]}) {
			continue
		}
		point, depth := a.projectPointDepth(x[i], y[i], z[i])
		points = append(points, projectedScatterPoint{point: point, depth: depth, index: i})
	}
	return points
}

func scatterScalarColors(scatter *core.Scatter2D) []render.Color {
	if scatter == nil || len(scatter.ScalarValues) == 0 {
		return nil
	}
	mapping := scatter.ScalarMap().Resolved()
	alpha := scatter.EffectiveAlpha(scatter.Alpha)
	colors := make([]render.Color, len(scatter.ScalarValues))
	for i, value := range scatter.ScalarValues {
		colors[i] = mapping.Color(value, alpha)
	}
	return colors
}

func scatterPointColors(fallback render.Color, source []render.Color, points []projectedScatterPoint) []render.Color {
	if len(points) == 0 {
		return nil
	}
	colors := make([]render.Color, len(points))
	if len(source) == 1 {
		for i := range colors {
			colors[i] = source[0]
		}
		return colors
	}
	if len(source) > 1 {
		for i, point := range points {
			if point.index < len(source) {
				colors[i] = source[point.index]
			} else {
				colors[i] = fallback
			}
		}
		return colors
	}
	for i := range colors {
		colors[i] = fallback
	}
	return colors
}

func reorderScatterFloat64s(values []float64, points []projectedScatterPoint) []float64 {
	if len(values) == 0 || len(points) == 0 {
		return nil
	}
	reordered := make([]float64, len(points))
	if len(values) == 1 {
		for i := range reordered {
			reordered[i] = values[0]
		}
		return reordered
	}
	for i, point := range points {
		if point.index < len(values) {
			reordered[i] = values[point.index]
		}
	}
	return reordered
}

func reorderScatterColors(values []render.Color, points []projectedScatterPoint) []render.Color {
	if len(values) == 0 || len(points) == 0 {
		return nil
	}
	reordered := make([]render.Color, len(points))
	if len(values) == 1 {
		for i := range reordered {
			reordered[i] = values[0]
		}
		return reordered
	}
	for i, point := range points {
		if point.index < len(values) {
			reordered[i] = values[point.index]
		}
	}
	return reordered
}

func depthShadedScatterColors(colors []render.Color, points []projectedScatterPoint) []render.Color {
	if len(points) == 0 {
		return nil
	}
	minZ, maxZ := points[0].depth, points[0].depth
	for _, point := range points[1:] {
		if point.depth < minZ {
			minZ = point.depth
		}
		if point.depth > maxZ {
			maxZ = point.depth
		}
	}
	shadedColors := make([]render.Color, len(points))
	for i, point := range points {
		saturation := 1.0
		if maxZ != minZ {
			saturation = 1 - ((point.depth-minZ)/(maxZ-minZ))*0.7
		}
		shaded := render.Color{}
		if i < len(colors) {
			shaded = colors[i]
		}
		shadedColors[i] = shaded.WithAlphaMultiplier(saturation)
	}
	return shadedColors
}

//nolint:gocritic // Triangulation and PlotOptions remain value snapshots throughout projection.
func (a *Axes3D) projectTriangulationFaces(tri core.Triangulation, z []float64, baseColor render.Color, opt core.PlotOptions) ([][]geom.Pt, []render.Color, []float64, float64, core.ScalarMapInfo) {
	type triFace struct {
		polygon []geom.Pt
		value   float64
		color   render.Color
		depth   float64
	}
	faces := make([]triFace, 0, len(tri.Triangles))
	values := make([]float64, 0, len(tri.Triangles))
	collectionDepth := math.Inf(1)
	for triIdx, t := range tri.Triangles {
		if tri.Masked(triIdx) {
			continue
		}
		polygon := make([]geom.Pt, 0, 3)
		points := [3]vec3{}
		depth := 0.0
		value := 0.0
		valid := true
		for i, idx := range t {
			if idx < 0 || idx >= len(tri.X) || idx >= len(tri.Y) || idx >= len(z) || !isFinite3D(tri.X[idx], tri.Y[idx], z[idx]) {
				valid = false
				break
			}
			points[i] = vec3{tri.X[idx], tri.Y[idx], z[idx]}
			value += z[idx]
			pt, zDepth := a.projectPointDepth(tri.X[idx], tri.Y[idx], z[idx])
			polygon = append(polygon, pt)
			depth += zDepth
			if zDepth < collectionDepth {
				collectionDepth = zDepth
			}
		}
		if valid {
			if opt.AxLimClip && !a.polygonWithin3DViewLimits(points[:]) {
				continue
			}
			value /= 3
			normal := points[0].sub(points[1]).cross(points[1].sub(points[2]))
			faces = append(faces, triFace{
				polygon: polygon,
				value:   value,
				color:   shade3DFaceColor(baseColor, normal),
				depth:   depth / 3,
			})
			values = append(values, value)
		}
	}
	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].depth > faces[j].depth
	})
	useMapping := opt.Colormap.OrZero() != ""
	mapping := core.ScalarMapInfo{}
	if useMapping {
		mapping = resolvePlotScalarMap(values, opt)
	}
	shade := trisurfShadeEnabled(opt, useMapping)
	polygons := make([][]geom.Pt, len(faces))
	colors := make([]render.Color, len(faces))
	scalarValues := []float64(nil)
	if useMapping {
		scalarValues = make([]float64, len(faces))
	}
	for i, face := range faces {
		polygons[i] = face.polygon
		if useMapping {
			colors[i] = mapping.Color(face.value, baseColor.A)
			scalarValues[i] = face.value
		} else {
			if shade {
				colors[i] = face.color
			} else {
				colors[i] = baseColor
			}
		}
	}
	return polygons, colors, scalarValues, computed3DCollectionZ(collectionDepth), mapping
}

func (a *Axes3D) pointWithin3DViewLimits(point vec3) bool {
	if a == nil || !isFinite3D(point[0], point[1], point[2]) {
		return false
	}
	mins, maxs := a.projectionLimits()
	for i := range 3 {
		lo, hi := mins[i], maxs[i]
		if hi < lo {
			lo, hi = hi, lo
		}
		if point[i] < lo || point[i] > hi {
			return false
		}
	}
	return true
}

func (a *Axes3D) polygonWithin3DViewLimits(polygon []vec3) bool {
	if len(polygon) == 0 {
		return false
	}
	for _, point := range polygon {
		if !a.pointWithin3DViewLimits(point) {
			return false
		}
	}
	return true
}

func (a *Axes3D) clip3DPolylineRuns(polyline []vec3) [][]vec3 {
	if len(polyline) == 0 {
		return nil
	}
	runs := make([][]vec3, 0, 1)
	current := make([]vec3, 0, len(polyline))
	flush := func() {
		if len(current) < 2 {
			current = current[:0]
			return
		}
		run := make([]vec3, len(current))
		copy(run, current)
		runs = append(runs, run)
		current = current[:0]
	}
	for _, point := range polyline {
		if a.pointWithin3DViewLimits(point) {
			current = append(current, point)
			continue
		}
		flush()
	}
	flush()
	return runs
}

func (a *Axes3D) projectBar3DSegments(x, y, z, dx, dy, dz []float64, axlimClip ...bool) [][]geom.Pt {
	n := minLen(x, y, z, dx, dy, dz)
	segments := make([][]geom.Pt, 0, n*8)
	clip := len(axlimClip) > 0 && axlimClip[0]
	for i := 0; i < n; i++ {
		x0 := x[i]
		x1 := x[i] + dx[i]
		y0 := y[i]
		y1 := y[i] + dy[i]
		bottom := z[i]
		top := z[i] + dz[i]
		if x1 < x0 {
			x0, x1 = x1, x0
		}
		if y1 < y0 {
			y0, y1 = y1, y0
		}
		if top < bottom {
			bottom, top = top, bottom
		}

		corners := [8]vec3{
			{x0, y0, bottom},
			{x1, y0, bottom},
			{x1, y1, bottom},
			{x0, y1, bottom},
			{x0, y0, top},
			{x1, y0, top},
			{x1, y1, top},
			{x0, y1, top},
		}
		raw := [][]vec3{
			{corners[4], corners[5]},
			{corners[5], corners[6]},
			{corners[6], corners[7]},
			{corners[7], corners[4]},
			{corners[4], corners[0]},
			{corners[5], corners[1]},
			{corners[6], corners[2]},
			{corners[7], corners[3]},
		}
		projected, _ := a.project3DLineSegments(raw, clip)
		segments = append(segments, projected...)
	}
	return segments
}

func (a *Axes3D) projectBar3DShadedFaces(x, y, z, dx, dy, dz []float64, baseColors []render.Color, axlimClip ...bool) ([][]geom.Pt, []render.Color) {
	type face struct {
		polygon []geom.Pt
		color   render.Color
		depth   float64
	}
	n := minLen(x, y, z, dx, dy, dz)
	faces := make([]face, 0, n*6)
	clip := len(axlimClip) > 0 && axlimClip[0]
	for i := 0; i < n; i++ {
		x0 := x[i]
		x1 := x[i] + dx[i]
		y0 := y[i]
		y1 := y[i] + dy[i]
		z0 := z[i]
		z1 := z[i] + dz[i]
		if x1 < x0 {
			x0, x1 = x1, x0
		}
		if y1 < y0 {
			y0, y1 = y1, y0
		}
		if z1 < z0 {
			z0, z1 = z1, z0
		}
		corners := [8][3]float64{
			{x0, y0, z0},
			{x1, y0, z0},
			{x1, y1, z0},
			{x0, y1, z0},
			{x0, y0, z1},
			{x1, y0, z1},
			{x1, y1, z1},
			{x0, y1, z1},
		}
		faceIndices := [][4]int{
			{0, 3, 2, 1},
			{4, 5, 6, 7},
			{0, 1, 5, 4},
			{2, 3, 7, 6},
			{3, 0, 4, 7},
			{1, 2, 6, 5},
		}
		normals := []vec3{
			{0, 0, -1},
			{0, 0, 1},
			{0, -1, 0},
			{0, 1, 0},
			{-1, 0, 0},
			{1, 0, 0},
		}
		for faceIdx, indices := range faceIndices {
			polygon := make([]geom.Pt, 0, len(indices))
			polygon3D := make([]vec3, 0, len(indices))
			depth := 0.0
			for _, idx := range indices {
				c := corners[idx]
				polygon3D = append(polygon3D, vec3{c[0], c[1], c[2]})
				pt, zDepth := a.projectPointDepth(c[0], c[1], c[2])
				polygon = append(polygon, pt)
				depth += zDepth
			}
			if clip && !a.polygonWithin3DViewLimits(polygon3D) {
				continue
			}
			baseColor := render.Color{R: 1, G: 1, B: 1, A: 1}
			colorIdx := i*6 + faceIdx
			if colorIdx < len(baseColors) {
				baseColor = baseColors[colorIdx]
			}
			faces = append(faces, face{
				polygon: polygon,
				color:   shade3DFaceColor(baseColor, normals[faceIdx]),
				depth:   depth / float64(len(indices)),
			})
		}
	}
	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].depth > faces[j].depth
	})
	polygons := make([][]geom.Pt, len(faces))
	colors := make([]render.Color, len(faces))
	for i, face := range faces {
		polygons[i] = face.polygon
		colors[i] = face.color
	}
	return polygons, colors
}

func shade3DFaceColor(color render.Color, normal vec3) render.Color {
	// Matplotlib art3d._shade_colors maps the light-source dot product from
	// [-1, 1] to [0.3, 1.0] and preserves alpha.
	az := (90.0 - 225.0) * math.Pi / 180
	alt := 19.4712 * math.Pi / 180
	light := vec3{
		math.Cos(az) * math.Cos(alt),
		math.Sin(az) * math.Cos(alt),
		math.Sin(alt),
	}
	shade := 0.65 + 0.35*normal.unit().dot(light)
	if math.IsNaN(shade) {
		shade = 0.65
	}
	shade = math.Max(0.3, math.Min(1, shade))
	color.R *= shade
	color.G *= shade
	color.B *= shade
	return color
}

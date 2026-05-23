package core

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type projectedScatterPoint struct {
	point geom.Pt
	depth float64
}

func reprojectScatter3D(scatter *Scatter2D, points []projectedScatterPoint) {
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
	scatter.Colors = depthShadedScatterColors(scatter.Color, points)
	scatter.EdgeColors = depthShadedScatterColors(scatter.EdgeColor, points)
}

func reprojectLine3D(line *Line2D, points []geom.Pt) {
	if line == nil {
		return
	}
	line.XY = append(line.XY[:0], points...)
}

func (a *Axes3D) projectedScatterData(x, y, z []float64) []projectedScatterPoint {
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if len(z) < n {
		n = len(z)
	}
	points := make([]projectedScatterPoint, 0, n)
	for i := 0; i < n; i++ {
		if !isFinite3D(x[i], y[i], z[i]) {
			continue
		}
		point, depth := a.projectPointDepth(x[i], y[i], z[i])
		points = append(points, projectedScatterPoint{point: point, depth: depth})
	}
	return points
}

func depthShadedScatterColors(color render.Color, points []projectedScatterPoint) []render.Color {
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
	colors := make([]render.Color, len(points))
	for i, point := range points {
		saturation := 1.0
		if maxZ != minZ {
			saturation = 1 - ((point.depth-minZ)/(maxZ-minZ))*0.7
		}
		shaded := color
		shaded.A *= saturation
		colors[i] = shaded
	}
	return colors
}

func (a *Axes3D) projectTriangulationFaces(tri Triangulation, z []float64, baseColor render.Color, opt PlotOptions) ([][]geom.Pt, []render.Color, float64, ScalarMapInfo) {
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
		if tri.masked(triIdx) {
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
	useMapping := opt.Colormap != nil && *opt.Colormap != ""
	mapping := ScalarMapInfo{}
	if useMapping {
		mapping = resolvePlotScalarMap(values, opt)
	}
	shade := trisurfShadeEnabled(opt, useMapping)
	polygons := make([][]geom.Pt, len(faces))
	colors := make([]render.Color, len(faces))
	for i, face := range faces {
		polygons[i] = face.polygon
		if useMapping {
			colors[i] = mapping.Color(face.value, baseColor.A)
		} else {
			if shade {
				colors[i] = face.color
			} else {
				colors[i] = baseColor
			}
		}
	}
	return polygons, colors, computed3DCollectionZ(collectionDepth), mapping
}

func (a *Axes3D) pointWithin3DViewLimits(point vec3) bool {
	if a == nil || !isFinite3D(point[0], point[1], point[2]) {
		return false
	}
	mins, maxs := a.projectionLimits()
	return point[0] >= mins[0] && point[0] <= maxs[0] &&
		point[1] >= mins[1] && point[1] <= maxs[1] &&
		point[2] >= mins[2] && point[2] <= maxs[2]
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

func (a *Axes3D) projectBar3DSegments(x, y, z, dx, dy, dz []float64) [][]geom.Pt {
	n := minLen(x, y, z, dx, dy, dz)
	segments := make([][]geom.Pt, 0, n*8)
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

		p00 := a.ProjectPoint(x0, y0, top)
		p10 := a.ProjectPoint(x1, y0, top)
		p11 := a.ProjectPoint(x1, y1, top)
		p01 := a.ProjectPoint(x0, y1, top)
		q00 := a.ProjectPoint(x0, y0, bottom)
		q10 := a.ProjectPoint(x1, y0, bottom)
		q11 := a.ProjectPoint(x1, y1, bottom)
		q01 := a.ProjectPoint(x0, y1, bottom)

		segments = append(segments,
			[]geom.Pt{p00, p10},
			[]geom.Pt{p10, p11},
			[]geom.Pt{p11, p01},
			[]geom.Pt{p01, p00},
			[]geom.Pt{p00, q00},
			[]geom.Pt{p10, q10},
			[]geom.Pt{p11, q11},
			[]geom.Pt{p01, q01},
		)
	}
	return segments
}

func (a *Axes3D) projectBar3DFaces(x, y, z, dx, dy, dz []float64) [][]geom.Pt {
	polygons, _ := a.projectBar3DShadedFaces(x, y, z, dx, dy, dz, render.Color{R: 1, G: 1, B: 1, A: 1})
	return polygons
}

func (a *Axes3D) projectBar3DShadedFaces(x, y, z, dx, dy, dz []float64, baseColor render.Color) ([][]geom.Pt, []render.Color) {
	type face struct {
		polygon []geom.Pt
		color   render.Color
		depth   float64
	}
	n := minLen(x, y, z, dx, dy, dz)
	faces := make([]face, 0, n*6)
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
			{0, 1, 2, 3},
			{4, 5, 6, 7},
			{0, 1, 5, 4},
			{1, 2, 6, 5},
			{2, 3, 7, 6},
			{3, 0, 4, 7},
		}
		normals := []vec3{
			{0, 0, -1},
			{0, 0, 1},
			{0, -1, 0},
			{1, 0, 0},
			{0, 1, 0},
			{-1, 0, 0},
		}
		for faceIdx, indices := range faceIndices {
			polygon := make([]geom.Pt, 0, len(indices))
			depth := 0.0
			for _, idx := range indices {
				c := corners[idx]
				pt, zDepth := a.projectPointDepth(c[0], c[1], c[2])
				polygon = append(polygon, pt)
				depth += zDepth
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

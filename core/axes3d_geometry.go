package core

import (
	"math"
	"sort"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func normalized3DDir(dir string) string {
	dir = strings.ToLower(dir)
	dir = strings.TrimPrefix(dir, "-")
	switch dir {
	case "x", "y", "z":
		return dir
	default:
		return "z"
	}
}

func juggle3DPoint(x, y, z float64, zdir string) vec3 {
	return juggle3DPointSigned(x, y, z, normalized3DDir(zdir))
}

func juggle3DPointSigned(x, y, z float64, zdir string) vec3 {
	switch strings.ToLower(zdir) {
	case "x":
		return vec3{z, x, y}
	case "y":
		return vec3{x, z, y}
	case "-x", "-y", "-z":
		return rotate3DPointAxes(x, y, z, zdir)
	default:
		return vec3{x, y, z}
	}
}

func rotate3DPointAxes(x, y, z float64, zdir string) vec3 {
	switch strings.ToLower(zdir) {
	case "x", "-y":
		return vec3{y, z, x}
	case "-x", "y":
		return vec3{z, x, y}
	default:
		return vec3{x, y, z}
	}
}

func (a *Axes3D) observe3DStemData(x, y, z []float64, bottom float64, orientation string) bool {
	n := minLen(x, y, z)
	changed := false
	for i := 0; i < n; i++ {
		start, end := stem3DLineEndpoints(x[i], y[i], z[i], bottom, orientation)
		if a.observe3DPoint(start[0], start[1], start[2]) {
			changed = true
		}
		if a.observe3DPoint(end[0], end[1], end[2]) {
			changed = true
		}
	}
	return changed
}

func stem3DLineEndpoints(x, y, z, bottom float64, orientation string) (vec3, vec3) {
	switch normalized3DDir(orientation) {
	case "x":
		return vec3{bottom, y, z}, vec3{x, y, z}
	case "y":
		return vec3{x, bottom, z}, vec3{x, y, z}
	default:
		return vec3{x, y, bottom}, vec3{x, y, z}
	}
}

func (a *Axes3D) projectStem3DGeometry(x, y, z []float64, bottom float64, orientation string, axlimClip bool) ([][]geom.Pt, []geom.Pt, []geom.Pt, float64) {
	n := minLen(x, y, z)
	segments := make([][]geom.Pt, 0, n)
	baseline := make([]geom.Pt, 0, n)
	offsets := make([]geom.Pt, 0, n)
	depth := math.Inf(1)
	for i := 0; i < n; i++ {
		start, end := stem3DLineEndpoints(x[i], y[i], z[i], bottom, orientation)
		if axlimClip && (!a.pointWithin3DViewLimits(start) || !a.pointWithin3DViewLimits(end)) {
			continue
		}
		startPt, startDepth := a.projectPointDepth(start[0], start[1], start[2])
		endPt, endDepth := a.projectPointDepth(end[0], end[1], end[2])
		segments = append(segments, []geom.Pt{startPt, endPt})
		baseline = append(baseline, startPt)
		offsets = append(offsets, endPt)
		if startDepth < depth {
			depth = startDepth
		}
		if endDepth < depth {
			depth = endDepth
		}
	}
	return segments, baseline, offsets, computed3DCollectionZ(depth)
}

func (a *Axes3D) projectQuiver3DSegments(x, y, z, u, v, w []float64, opt Quiver3DOptions) ([][]geom.Pt, float64) {
	return a.project3DLineSegments(quiver3DRawSegments(x, y, z, u, v, w, opt), opt.AxLimClip)
}

func (a *Axes3D) observeQuiver3DData(x, y, z, u, v, w []float64, opt Quiver3DOptions) bool {
	n := minLen(x, y, z, u, v, w)
	changed := false
	for i := range n {
		if a.observe3DPoint(x[i], y[i], z[i]) {
			changed = true
		}
	}
	return changed
}

func quiver3DRawSegments(x, y, z, u, v, w []float64, opt Quiver3DOptions) [][]vec3 {
	n := minLen(x, y, z, u, v, w)
	length := 1.0
	if opt.Length != nil {
		length = *opt.Length
	}
	arrowRatio := 0.3
	if opt.ArrowLengthRatio != nil {
		arrowRatio = *opt.ArrowLengthRatio
	}
	pivot := strings.ToLower(opt.Pivot)
	if pivot != "middle" && pivot != "tip" {
		pivot = "tail"
	}

	segments := make([][]vec3, 0, n*3)
	for i := 0; i < n; i++ {
		if !isFinite3D(x[i], y[i], z[i]) || !isFinite3D(u[i], v[i], w[i]) {
			continue
		}
		vec := vec3{u[i], v[i], w[i]}
		if opt.Normalize {
			vec = vec.unit()
		}
		if vec.norm() == 0 {
			continue
		}
		start, end := quiver3DShaft(vec3{x[i], y[i], z[i]}, vec, length, pivot)
		segments = append(segments, []vec3{start, end})
		headLen := length * arrowRatio
		for _, headDir := range quiver3DHeadDirections(vec) {
			segments = append(segments, []vec3{start, start.sub(headDir.scale(headLen))})
		}
	}
	return segments
}

func quiver3DShaft(anchor, vec vec3, length float64, pivot string) (vec3, vec3) {
	shaftStart := 0.0
	shaftEnd := length
	switch pivot {
	case "tail":
		shaftStart -= length
		shaftEnd -= length
	case "middle":
		shaftStart -= length / 2
		shaftEnd -= length / 2
	}
	return anchor.sub(vec.scale(shaftStart)), anchor.sub(vec.scale(shaftEnd))
}

func quiver3DHeadDirections(vec vec3) [2]vec3 {
	axisNorm := math.Hypot(vec[0], vec[1])
	axis := vec3{0, 1, 0}
	if axisNorm != 0 {
		axis = vec3{vec[1] / axisNorm, -vec[0] / axisNorm, 0}
	}
	angle := 15 * math.Pi / 180
	return [2]vec3{
		rotate3DVector(vec, axis, angle),
		rotate3DVector(vec, axis, -angle),
	}
}

func rotate3DVector(vec, axis vec3, angle float64) vec3 {
	axis = axis.unit()
	c := math.Cos(angle)
	s := math.Sin(angle)
	return vec.scale(c).add(axis.cross(vec).scale(s)).add(axis.scale(axis.dot(vec) * (1 - c)))
}

type projected3DPolygon struct {
	polygon []geom.Pt
	depth   float64
	color   render.Color
}

// resolveFillBetween3DMode ports the mode='auto' branch of
// matplotlib Axes3D.fill_between: 'polygon' when all points lie on one 3D
// plane, 'quad' otherwise.
func resolveFillBetween3DMode(x1, y1, z1, x2, y2, z2 []float64, mode FillBetween3DMode) FillBetween3DMode {
	if mode == "" || mode == FillBetween3DModeAuto {
		if fillBetween3DCoplanar(x1, y1, z1, x2, y2, z2) {
			return FillBetween3DModePolygon
		}
		return FillBetween3DModeQuad
	}
	return mode
}

// polygon3DNormal ports art3d._generate_normals for a single polygon: pick
// three points equally spaced around the polygon and take
// cross(p[0]-p[n/3], p[n/3]-p[2n/3]).
func polygon3DNormal(ps []vec3) vec3 {
	n := len(ps)
	if n < 3 {
		return vec3{}
	}
	i2, i3 := n/3, 2*n/3
	v1 := ps[0].sub(ps[i2])
	v2 := ps[i2].sub(ps[i3])
	return v1.cross(v2)
}

func fillBetween3DRawPolygons(x1, y1, z1, x2, y2, z2 []float64, mode FillBetween3DMode) [][]vec3 {
	n := minLen(x1, y1, z1, x2, y2, z2)
	if n < 2 {
		return nil
	}
	mode = resolveFillBetween3DMode(x1[:n], y1[:n], z1[:n], x2[:n], y2[:n], z2[:n], mode)
	if mode == FillBetween3DModePolygon {
		polygon := make([]vec3, 0, n*2)
		for i := 0; i < n; i++ {
			polygon = append(polygon, vec3{x1[i], y1[i], z1[i]})
		}
		for i := n - 1; i >= 0; i-- {
			polygon = append(polygon, vec3{x2[i], y2[i], z2[i]})
		}
		return [][]vec3{polygon}
	}

	polygons := make([][]vec3, 0, n-1)
	for i := 0; i+1 < n; i++ {
		polygons = append(polygons, []vec3{
			{x1[i], y1[i], z1[i]},
			{x1[i+1], y1[i+1], z1[i+1]},
			{x2[i+1], y2[i+1], z2[i+1]},
			{x2[i], y2[i], z2[i]},
		})
	}
	return polygons
}

func fillBetween3DCoplanar(x1, y1, z1, x2, y2, z2 []float64) bool {
	points := make([]vec3, 0, len(x1)*2)
	for i := range x1 {
		points = append(points, vec3{x1[i], y1[i], z1[i]})
	}
	for i := range x2 {
		points = append(points, vec3{x2[i], y2[i], z2[i]})
	}
	if len(points) < 4 {
		return true
	}
	p0 := points[0]
	normal := vec3{}
	for i := 1; i+1 < len(points); i++ {
		normal = points[i].sub(p0).cross(points[i+1].sub(p0))
		if normal.norm() > 1e-12 {
			break
		}
	}
	if normal.norm() <= 1e-12 {
		return true
	}
	for _, p := range points[1:] {
		if math.Abs(p.sub(p0).dot(normal)) > 1e-12 {
			return false
		}
	}
	return true
}

func (a *Axes3D) observe3DPlaneBars(x, heights []float64, width, baseline float64, baselines []float64, z float64, zs []float64, zdir string) bool {
	changed := false
	for _, polygon := range planeBar3DRawPolygons(x, heights, width, baseline, baselines, z, zs, zdir) {
		for _, p := range polygon {
			if a.observe3DPoint(p[0], p[1], p[2]) {
				changed = true
			}
		}
	}
	return changed
}

func (a *Axes3D) project3DPlaneBars(x, heights []float64, width, baseline float64, baselines []float64, z float64, zs []float64, zdir string) ([][]geom.Pt, float64) {
	return a.projectSorted3DPolygons(planeBar3DRawPolygons(x, heights, width, baseline, baselines, z, zs, zdir))
}

func planeBar3DRawPolygons(x, heights []float64, width, baseline float64, baselines []float64, z float64, zs []float64, zdir string) [][]vec3 {
	n := minLen(x, heights)
	polygons := make([][]vec3, 0, n)
	for i := 0; i < n; i++ {
		base := baseline
		if len(baselines) == 1 {
			base = baselines[0]
		} else if i < len(baselines) {
			base = baselines[i]
		}
		zi := z
		if len(zs) == 1 {
			zi = zs[0]
		} else if i < len(zs) {
			zi = zs[i]
		}
		left := x[i] - width*0.5
		right := x[i] + width*0.5
		top := base + heights[i]
		polygons = append(polygons, []vec3{
			juggle3DPoint(left, base, zi, zdir),
			juggle3DPoint(left, top, zi, zdir),
			juggle3DPoint(right, top, zi, zdir),
			juggle3DPoint(right, base, zi, zdir),
		})
	}
	return polygons
}

func (a *Axes3D) observe3DErrorBarData(x, y, z, xErr, yErr, zErr []float64, opt ErrorBar3DOptions) bool {
	changed := false
	for _, segment := range errorBar3DRawSegments(x, y, z, xErr, yErr, zErr, opt) {
		for _, p := range segment {
			if a.observe3DPoint(p[0], p[1], p[2]) {
				changed = true
			}
		}
	}
	return changed
}

func (a *Axes3D) projectErrorBar3DSegments(x, y, z, xErr, yErr, zErr []float64, opt ErrorBar3DOptions) ([][]geom.Pt, float64) {
	return a.project3DLineSegments(errorBar3DRawSegments(x, y, z, xErr, yErr, zErr, opt), opt.AxLimClip)
}

func errorBar3DRawSegments(x, y, z, xErr, yErr, zErr []float64, opt ErrorBar3DOptions) [][]vec3 {
	n := minLen(x, y, z)
	segments := make([][]vec3, 0, n*15)
	caps := make([][]vec3, 0, n*12)
	capHalf := 0.0
	if opt.CapSize != nil && *opt.CapSize > 0 {
		capHalf = *opt.CapSize * 0.5
	}
	appendAxis := func(center vec3, axis int, low, high float64) {
		if low <= 0 && high <= 0 {
			return
		}
		start := center
		end := center
		start[axis] -= low
		end[axis] += high
		segments = append(segments, []vec3{start, end})
		if capHalf > 0 {
			if low > 0 {
				caps = append(caps, errorBar3DCapSegments(start, axis, capHalf)...)
			}
			if high > 0 {
				caps = append(caps, errorBar3DCapSegments(end, axis, capHalf)...)
			}
		}
	}
	for i := 0; i < n; i++ {
		center := vec3{x[i], y[i], z[i]}
		xLow, xHigh := resolveErrorRange(xErr, opt.XErrLower, opt.XErrUpper, i)
		yLow, yHigh := resolveErrorRange(yErr, opt.YErrLower, opt.YErrUpper, i)
		zLow, zHigh := resolveErrorRange(zErr, opt.ZErrLower, opt.ZErrUpper, i)
		appendAxis(center, 0, xLow, xHigh)
		appendAxis(center, 1, yLow, yHigh)
		appendAxis(center, 2, zLow, zHigh)
	}
	segments = append(segments, caps...)
	return segments
}

func errorBar3DCapSegments(center vec3, axis int, half float64) [][]vec3 {
	first := (axis + 1) % 3
	second := (axis + 2) % 3
	a0, a1 := center, center
	a0[first] -= half
	a1[first] += half
	b0, b1 := center, center
	b0[second] -= half
	b1[second] += half
	return [][]vec3{{a0, a1}, {b0, b1}}
}

func (a *Axes3D) projectSorted3DPolygons(raw [][]vec3, axlimClip ...bool) ([][]geom.Pt, float64) {
	polygons, _, z := a.projectSorted3DPolygonsWithColors(raw, nil, axlimClip...)
	return polygons, z
}

// projectSorted3DPolygonsWithColors projects raw polygons and depth-sorts
// them painter-style, keeping each optional per-polygon color paired with its
// polygon through the sort.
func (a *Axes3D) projectSorted3DPolygonsWithColors(raw [][]vec3, colors []render.Color, axlimClip ...bool) ([][]geom.Pt, []render.Color, float64) {
	projected := make([]projected3DPolygon, 0, len(raw))
	collectionDepth := math.Inf(1)
	clip := len(axlimClip) > 0 && axlimClip[0]
	for index, polygon3D := range raw {
		if len(polygon3D) < 3 {
			continue
		}
		if clip && !a.polygonWithin3DViewLimits(polygon3D) {
			continue
		}
		polygon := make([]geom.Pt, 0, len(polygon3D))
		depth := 0.0
		valid := true
		for _, p := range polygon3D {
			if !isFinite3D(p[0], p[1], p[2]) {
				valid = false
				break
			}
			pt, zDepth := a.projectPointDepth(p[0], p[1], p[2])
			polygon = append(polygon, pt)
			depth += zDepth
			if zDepth < collectionDepth {
				collectionDepth = zDepth
			}
		}
		if valid {
			item := projected3DPolygon{polygon: polygon, depth: depth / float64(len(polygon3D))}
			if index < len(colors) {
				item.color = colors[index]
			}
			projected = append(projected, item)
		}
	}
	sort.SliceStable(projected, func(i, j int) bool {
		return projected[i].depth > projected[j].depth
	})
	polygons := make([][]geom.Pt, len(projected))
	for i, item := range projected {
		polygons[i] = item.polygon
	}
	var sortedColors []render.Color
	if colors != nil {
		sortedColors = make([]render.Color, len(projected))
		for i, item := range projected {
			sortedColors[i] = item.color
		}
	}
	return polygons, sortedColors, computed3DCollectionZ(collectionDepth)
}

func (a *Axes3D) projectSorted3DLineSegments(raw [][]vec3) ([][]geom.Pt, float64) {
	type projectedSegment struct {
		segment []geom.Pt
		depth   float64
	}
	projected := make([]projectedSegment, 0, len(raw))
	collectionDepth := math.Inf(1)
	for _, segment3D := range raw {
		if len(segment3D) < 2 {
			continue
		}
		segment := make([]geom.Pt, 0, len(segment3D))
		depth := 0.0
		valid := true
		for _, p := range segment3D {
			if !isFinite3D(p[0], p[1], p[2]) {
				valid = false
				break
			}
			pt, zDepth := a.projectPointDepth(p[0], p[1], p[2])
			segment = append(segment, pt)
			depth += zDepth
			if zDepth < collectionDepth {
				collectionDepth = zDepth
			}
		}
		if valid {
			projected = append(projected, projectedSegment{segment: segment, depth: depth / float64(len(segment3D))})
		}
	}
	sort.SliceStable(projected, func(i, j int) bool {
		return projected[i].depth > projected[j].depth
	})
	segments := make([][]geom.Pt, len(projected))
	for i, item := range projected {
		segments[i] = item.segment
	}
	return segments, computed3DCollectionZ(collectionDepth)
}

func (a *Axes3D) project3DLineSegments(raw [][]vec3, axlimClip ...bool) ([][]geom.Pt, float64) {
	segments := make([][]geom.Pt, 0, len(raw))
	collectionDepth := math.Inf(1)
	clip := len(axlimClip) > 0 && axlimClip[0]
	for _, segment3D := range raw {
		if len(segment3D) < 2 {
			continue
		}
		runs := [][]vec3{segment3D}
		if clip {
			runs = a.clip3DPolylineRuns(segment3D)
		}
		for _, run3D := range runs {
			segment, minDepth, ok := a.project3DLineSegment(run3D)
			if !ok {
				continue
			}
			if minDepth < collectionDepth {
				collectionDepth = minDepth
			}
			segments = append(segments, segment)
		}
	}
	return segments, computed3DCollectionZ(collectionDepth)
}

func (a *Axes3D) project3DLineSegment(segment3D []vec3) ([]geom.Pt, float64, bool) {
	if len(segment3D) < 2 {
		return nil, math.Inf(1), false
	}
	segment := make([]geom.Pt, 0, len(segment3D))
	minDepth := math.Inf(1)
	for _, p := range segment3D {
		if !isFinite3D(p[0], p[1], p[2]) {
			return nil, math.Inf(1), false
		}
		pt, zDepth := a.projectPointDepth(p[0], p[1], p[2])
		segment = append(segment, pt)
		if zDepth < minDepth {
			minDepth = zDepth
		}
	}
	return segment, minDepth, true
}

func (a *Axes3D) projectedData(x, y, z []float64, axlimClip ...bool) []geom.Pt {
	if a == nil || a.Axes == nil {
		return nil
	}

	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if len(z) < n {
		n = len(z)
	}
	if n <= 0 {
		return nil
	}

	clip := len(axlimClip) > 0 && axlimClip[0]
	pts := make([]geom.Pt, 0, n)
	for i := 0; i < n; i++ {
		if !isFinite3D(x[i], y[i], z[i]) {
			continue
		}
		if clip && !a.pointWithin3DViewLimits(vec3{x[i], y[i], z[i]}) {
			continue
		}
		pts = append(pts, a.ProjectPoint(x[i], y[i], z[i]))
	}
	return pts
}

func (a *Axes3D) projectWireframeSegments(x, y []float64, z [][]float64, opts ...PlotOptions) [][]geom.Pt {
	if a == nil || len(z) == 0 {
		return nil
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return nil
	}
	for i := 1; i < rows; i++ {
		if len(z[i]) != cols {
			return nil
		}
	}

	rowIndices, colIndices := wireframeSampleIndices(rows, cols, firstPlotOptions(opts))
	opt := firstPlotOptions(opts)
	segments := make([][]geom.Pt, 0, len(rowIndices)+len(colIndices))
	for _, row := range rowIndices {
		line3D := make([]vec3, 0, cols)
		for col := 0; col < cols; col++ {
			line3D = append(line3D, vec3{x[col], y[row], z[row][col]})
		}
		lines, _ := a.project3DLineSegments([][]vec3{line3D}, opt.AxLimClip)
		for _, line := range lines {
			if len(line) > 1 {
				segments = append(segments, line)
			}
		}
	}
	for _, col := range colIndices {
		line3D := make([]vec3, 0, rows)
		for row := 0; row < rows; row++ {
			line3D = append(line3D, vec3{x[col], y[row], z[row][col]})
		}
		lines, _ := a.project3DLineSegments([][]vec3{line3D}, opt.AxLimClip)
		for _, line := range lines {
			if len(line) > 1 {
				segments = append(segments, line)
			}
		}
	}
	return segments
}

func firstPlotOptions(opts []PlotOptions) PlotOptions {
	if len(opts) == 0 {
		return PlotOptions{}
	}
	return opts[0]
}

func wireframeSampleIndices(rows, cols int, opt PlotOptions) ([]int, []int) {
	hasStride := opt.RStride != nil || opt.CStride != nil
	rstride, cstride := 1, 1
	if opt.RStride != nil {
		rstride = *opt.RStride
	}
	if opt.CStride != nil {
		cstride = *opt.CStride
	}
	if !hasStride {
		rcount, ccount := default3DSurfaceCount, default3DSurfaceCount
		if opt.RCount != nil {
			rcount = *opt.RCount
		}
		if opt.CCount != nil {
			ccount = *opt.CCount
		}
		rstride = samplingStrideFromCount(rows, rcount)
		cstride = samplingStrideFromCount(cols, ccount)
	}
	return steppedSampleIndices(rows, rstride), steppedSampleIndices(cols, cstride)
}

func samplingStrideFromCount(length, count int) int {
	if count <= 0 {
		return 0
	}
	stride := int(math.Ceil(float64(length) / float64(count)))
	if stride < 1 {
		return 1
	}
	return stride
}

func steppedSampleIndices(length, stride int) []int {
	if length <= 0 || stride <= 0 {
		return nil
	}
	if (length-1)%stride == 0 {
		indices := make([]int, 0, (length+stride-1)/stride)
		for i := 0; i < length; i += stride {
			indices = append(indices, i)
		}
		return indices
	}
	indices := make([]int, 0, (length+stride-1)/stride+1)
	for i := 0; i < length+stride; i += stride {
		if i >= length {
			indices = append(indices, length-1)
			break
		}
		indices = append(indices, i)
	}
	return indices
}

func project3DPoint(x, y, z, elevationDeg, azimuthDeg, distance float64) geom.Pt {
	return project3DPointWithLimits(x, y, z, elevationDeg, azimuthDeg, distance, vec3{0, 0, 0}, vec3{1, 1, 1})
}

func project3DPointWithLimits(
	x, y, z, elevationDeg, azimuthDeg, distance float64,
	mins, maxs vec3,
	state ...projected3DState,
) geom.Pt {
	s := projected3DStateOrDefault(state...)
	if distance <= 0 {
		az := azimuthDeg * math.Pi / 180
		v := elevationDeg * math.Pi / 180

		x2 := x*math.Cos(az) - y*math.Sin(az)
		y2 := x*math.Sin(az) + y*math.Cos(az)
		return geom.Pt{X: x2, Y: y2*math.Cos(v) - z*math.Sin(v)}
	}

	m := default3DProjectionMatrix(elevationDeg, azimuthDeg, distance, mins, maxs, s)
	tx, ty, _ := transform3DPoint(m, x, y, z)
	return geom.Pt{X: tx, Y: ty}
}

func default3DProjectionMatrix(
	elevationDeg, azimuthDeg, distance float64,
	mins, maxs vec3,
	state ...projected3DState,
) mat4 {
	s := projected3DStateOrDefault(state...)
	aspect := s.boxAspect
	world := worldTransformation(mins[0], maxs[0], mins[1], maxs[1], mins[2], maxs[2], aspect)
	center := vec3{0.5 * aspect[0], 0.5 * aspect[1], 0.5 * aspect[2]}

	elev := elevationDeg * math.Pi / 180
	az := azimuthDeg * math.Pi / 180
	viewDir := vec3{
		math.Cos(elev) * math.Cos(az),
		math.Cos(elev) * math.Sin(az),
		math.Sin(elev),
	}
	eye := center.add(viewDir.scale(distance))
	u, v, w := viewAxes(eye, center, elevationDeg, s.verticalAxis, s.rollDeg)
	viewEye := eye
	if !math.IsInf(s.focalLength, 1) {
		viewEye = center.add(viewDir.scale(distance * s.focalLength))
	}
	view := viewTransformation(u, v, w, viewEye)
	proj := perspectiveTransformation(-distance, distance, s.focalLength)
	if math.IsInf(s.focalLength, 1) {
		proj = orthographicTransformation(-distance, distance)
	}
	return proj.mul(view.mul(world))
}

func default3DBoxAspect() vec3 {
	aspect := vec3{4, 4, 3}
	scale := default3DBoxAspectScale * default3DBoxAspectZoom25 / aspect.norm()
	return aspect.scale(scale)
}

func worldTransformation(xmin, xmax, ymin, ymax, zmin, zmax float64, aspect vec3) mat4 {
	dx := (xmax - xmin) / aspect[0]
	dy := (ymax - ymin) / aspect[1]
	dz := (zmax - zmin) / aspect[2]
	return mat4{
		{1 / dx, 0, 0, -xmin / dx},
		{0, 1 / dy, 0, -ymin / dy},
		{0, 0, 1 / dz, -zmin / dz},
		{0, 0, 0, 1},
	}
}

func viewTransformation(u, v, w, eye vec3) mat4 {
	rot := mat4{
		{u[0], u[1], u[2], 0},
		{v[0], v[1], v[2], 0},
		{w[0], w[1], w[2], 0},
		{0, 0, 0, 1},
	}
	translate := mat4{
		{1, 0, 0, -eye[0]},
		{0, 1, 0, -eye[1]},
		{0, 0, 1, -eye[2]},
		{0, 0, 0, 1},
	}
	return rot.mul(translate)
}

func perspectiveTransformation(zfront, zback, focalLength float64) mat4 {
	b := (zfront + zback) / (zfront - zback)
	c := -2 * (zfront * zback) / (zfront - zback)
	return mat4{
		{focalLength, 0, 0, 0},
		{0, focalLength, 0, 0},
		{0, 0, b, c},
		{0, 0, -1, 0},
	}
}

func orthographicTransformation(zfront, zback float64) mat4 {
	a := -(zfront + zback)
	b := -(zfront - zback)
	return mat4{
		{2, 0, 0, 0},
		{0, 2, 0, 0},
		{0, 0, -2, 0},
		{0, 0, a, b},
	}
}

func transform3DPoint(m mat4, x, y, z float64) (float64, float64, float64) {
	vec := [4]float64{x, y, z, 1}
	var out [4]float64
	for row := range 4 {
		for col := range 4 {
			out[row] += m[row][col] * vec[col]
		}
	}
	if out[3] == 0 {
		return out[0], out[1], out[2]
	}
	return out[0] / out[3], out[1] / out[3], out[2] / out[3]
}

type vec3 [3]float64

func (v vec3) add(other vec3) vec3 {
	return vec3{v[0] + other[0], v[1] + other[1], v[2] + other[2]}
}

func (v vec3) sub(other vec3) vec3 {
	return vec3{v[0] - other[0], v[1] - other[1], v[2] - other[2]}
}

func (v vec3) scale(s float64) vec3 {
	return vec3{v[0] * s, v[1] * s, v[2] * s}
}

func (v vec3) norm() float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

func (v vec3) unit() vec3 {
	n := v.norm()
	if n == 0 {
		return vec3{}
	}
	return v.scale(1 / n)
}

func (v vec3) cross(other vec3) vec3 {
	return vec3{
		v[1]*other[2] - v[2]*other[1],
		v[2]*other[0] - v[0]*other[2],
		v[0]*other[1] - v[1]*other[0],
	}
}

func (v vec3) dot(other vec3) float64 {
	return v[0]*other[0] + v[1]*other[1] + v[2]*other[2]
}

type mat4 [4][4]float64

func (m mat4) mul(other mat4) mat4 {
	var out mat4
	for row := range 4 {
		for col := range 4 {
			for k := range 4 {
				out[row][col] += m[row][k] * other[k][col]
			}
		}
	}
	return out
}

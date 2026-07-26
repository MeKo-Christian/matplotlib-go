package plot3d

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// SetView updates the 3D viewing angles in degrees.
func (a *Axes3D) SetView(elevationDeg, azimuthDeg float64) {
	if a == nil {
		return
	}
	a.elevationDeg = elevationDeg
	a.azimuthDeg = azimuthDeg
	a.reproject3DArtists()
}

// SetViewInit sets the full 3D view parameters in degrees.
func (a *Axes3D) SetViewInit(elevationDeg, azimuthDeg, rollDeg float64, verticalAxis string) error {
	if a == nil {
		return nil
	}
	axis, err := parse3DVerticalAxis(verticalAxis)
	if err != nil {
		return err
	}
	a.elevationDeg = elevationDeg
	a.azimuthDeg = azimuthDeg
	a.rollDeg = rollDeg
	a.verticalAxis = axis
	a.reproject3DArtists()
	return nil
}

// SetDistance sets the perspective distance used by the 3D projection.
// Non-positive values disable perspective scaling.
func (a *Axes3D) SetDistance(distance float64) {
	if a == nil {
		return
	}
	a.distance = distance
	a.reproject3DArtists()
}

// SetProjectionType sets the 3D projection type, matching Matplotlib's
// set_proj_type for the supported static-rendering modes "persp" and "ortho".
func (a *Axes3D) SetProjectionType(projType string, focalLength ...float64) error {
	if a == nil {
		return nil
	}
	focal := math.NaN()
	if len(focalLength) > 0 {
		focal = focalLength[0]
	}
	switch projType {
	case "persp":
		if math.IsNaN(focal) {
			focal = default3DFocalLength
		}
		if focal <= 0 || !isFinite(focal) {
			return fmt.Errorf("focal_length = %v must be greater than 0", focal)
		}
	case "ortho":
		if !math.IsNaN(focal) && !math.IsInf(focal, 1) {
			return fmt.Errorf("focal_length = %v must be +Inf for proj_type = %s", focal, projType)
		}
		focal = math.Inf(1)
	default:
		return fmt.Errorf("invalid projection type %q", projType)
	}
	a.focalLength = focal
	a.reproject3DArtists()
	return nil
}

// ProjectionType reports the current 3D projection type.
func (a *Axes3D) ProjectionType() string {
	if a == nil {
		return ""
	}
	if math.IsInf(a.focalLength, 1) {
		return "ortho"
	}
	return "persp"
}

// SetDefaults sets standard Matplotlib-like defaults for elevation, azimuth,
// and perspective distance.
func (a *Axes3D) SetDefaults() {
	if a == nil {
		return
	}
	a.elevationDeg = default3DElevationDeg
	a.azimuthDeg = default3DAzimuthDeg
	a.showXLabels = true
	a.showYLabels = true
	a.showZLabels = true
	a.distance = default3DDistance
	a.rollDeg = default3DRollDeg
	a.verticalAxis = default3DVerticalAxis
	a.boxAspect = default3DBoxAspect()
	a.focalLength = default3DFocalLength
	a.reproject3DArtists()
}

// SetXLim fixes the 3D x-axis view limits used for projection and clipping.
func (a *Axes3D) SetXLim(minVal, maxVal float64) {
	a.setViewLimit3D(0, minVal, maxVal)
}

// SetYLim fixes the 3D y-axis view limits used for projection and clipping.
func (a *Axes3D) SetYLim(minVal, maxVal float64) {
	a.setViewLimit3D(1, minVal, maxVal)
}

// SetZLim fixes the 3D z-axis view limits used for projection and clipping.
func (a *Axes3D) SetZLim(minVal, maxVal float64) {
	a.setViewLimit3D(2, minVal, maxVal)
}

func (a *Axes3D) setViewLimit3D(axis int, minVal, maxVal float64) {
	if a == nil || axis < 0 || axis >= len(a.viewSet) || !isFinite(minVal) || !isFinite(maxVal) {
		return
	}
	a.viewMin[axis] = minVal
	a.viewMax[axis] = maxVal
	a.viewSet[axis] = true
	a.reproject3DArtists()
}

// SetBoxAspect3D sets the 3D box physical aspect ratios.
func (a *Axes3D) SetBoxAspect3D(aspect [3]float64, zoom ...float64) error {
	if a == nil {
		return nil
	}
	zoomFactor := 1.0
	if len(zoom) > 0 {
		zoomFactor = zoom[0]
	}
	if zoomFactor <= 0 {
		return fmt.Errorf("zoom = %v must be > 0", zoomFactor)
	}
	boxAspect := vec3{aspect[0], aspect[1], aspect[2]}
	if !isFinite3D(boxAspect[0], boxAspect[1], boxAspect[2]) {
		return fmt.Errorf("box aspect %v must be finite", boxAspect)
	}
	norm := boxAspect.norm()
	if norm == 0 {
		return fmt.Errorf("box aspect %v must not be zero", boxAspect)
	}
	scale := default3DBoxAspectScale * default3DBoxAspectZoom25 * zoomFactor / norm
	boxAspect = boxAspect.scale(scale)
	a.boxAspect = rollToVertical(boxAspect, a.verticalAxis, true)
	a.reproject3DArtists()
	return nil
}

func (a *Axes3D) SetZLabel(label string) {
	if a == nil {
		return
	}
	a.zLabel = label
}

// SetShowXTickLabels controls whether x-axis tick labels are drawn on the 3D frame.
func (a *Axes3D) SetShowXTickLabels(show bool) {
	if a == nil {
		return
	}
	a.showXLabels = show
	a.reproject3DArtists()
}

// SetShowYTickLabels controls whether y-axis tick labels are drawn on the 3D frame.
func (a *Axes3D) SetShowYTickLabels(show bool) {
	if a == nil {
		return
	}
	a.showYLabels = show
	a.reproject3DArtists()
}

// SetShowZTickLabels controls whether z-axis tick labels are drawn on the 3D frame.
func (a *Axes3D) SetShowZTickLabels(show bool) {
	if a == nil {
		return
	}
	a.showZLabels = show
	a.reproject3DArtists()
}

func (a *Axes3D) ZLabel() string {
	if a == nil {
		return ""
	}
	return a.zLabel
}

// View reports the current 3D orientation state.
func (a *Axes3D) View() (elevationDeg, azimuthDeg, distance float64) {
	if a == nil {
		return 0, 0, 0
	}
	return a.elevationDeg, a.azimuthDeg, a.distance
}

// ProjectPoint projects a single 3D point into this Axes3D data space.
func (a *Axes3D) ProjectPoint(x, y, z float64) geom.Pt {
	if a == nil {
		return geom.Pt{}
	}
	mins, maxs := a.projectionLimits()
	return project3DPointWithLimits(x, y, z, a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs, a.projectionState())
}

func (a *Axes3D) projectPointDepth(x, y, z float64) (geom.Pt, float64) {
	if a == nil {
		return geom.Pt{}, 0
	}
	if a.distance <= 0 {
		return a.ProjectPoint(x, y, z), z
	}
	mins, maxs := a.projectionLimits()
	m := default3DProjectionMatrix(a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs, a.projectionState())
	tx, ty, tz := transform3DPoint(m, x, y, z)
	return geom.Pt{X: tx, Y: ty}, tz
}

func (a *Axes3D) projectionLimits() (vec3, vec3) {
	if a == nil {
		return vec3{0, 0, 0}, vec3{1, 1, 1}
	}
	mins := vec3{0, 0, 0}
	maxs := vec3{1, 1, 1}
	if a.hasData {
		mins = a.dataMin
		maxs = a.dataMax
	}
	for i := range 3 {
		if a.viewSet[i] {
			mins[i] = a.viewMin[i]
			maxs[i] = a.viewMax[i]
			continue
		}
		if a.hasData && mins[i] == maxs[i] {
			mins[i] -= 0.5
			maxs[i] += 0.5
		}
		if a.hasData {
			margin := (maxs[i] - mins[i]) * a.default3DDataMargin(i)
			mins[i] -= margin
			maxs[i] += margin
		}
		viewMargin := (maxs[i] - mins[i]) * default3DWorldViewMargin
		mins[i] -= viewMargin
		maxs[i] += viewMargin
	}
	return mins, maxs
}

func (a *Axes3D) default3DDataMargin(axis int) float64 {
	if axis == 2 {
		if a != nil {
			return a.zMargin
		}
		return 0
	}
	return 0.05
}

func (a *Axes3D) ensure3DZMargin(margin float64) bool {
	if a == nil || margin <= a.zMargin {
		return false
	}
	a.zMargin = margin
	return true
}

func (a *Axes3D) observe3DPoint(x, y, z float64) bool {
	if a == nil || !isFinite3D(x, y, z) {
		return false
	}
	p := vec3{x, y, z}
	if !a.hasData {
		a.dataMin = p
		a.dataMax = p
		a.hasData = true
		return true
	}
	changed := false
	for i := range 3 {
		if p[i] < a.dataMin[i] {
			a.dataMin[i] = p[i]
			changed = true
		}
		if p[i] > a.dataMax[i] {
			a.dataMax[i] = p[i]
			changed = true
		}
	}
	return changed
}

func (a *Axes3D) observe3DData(x, y, z []float64) bool {
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if len(z) < n {
		n = len(z)
	}
	changed := false
	for i := 0; i < n; i++ {
		if a.observe3DPoint(x[i], y[i], z[i]) {
			changed = true
		}
	}
	return changed
}

func (a *Axes3D) observe3DGrid(x, y []float64, z [][]float64) bool {
	if a == nil || len(z) == 0 {
		return false
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return false
	}
	changed := false
	for row := 0; row < rows; row++ {
		if len(z[row]) != cols {
			return changed
		}
		for col := 0; col < cols; col++ {
			if a.observe3DPoint(x[col], y[row], z[row][col]) {
				changed = true
			}
		}
	}
	return changed
}

//nolint:gocritic // PlotOptions is an immutable snapshot retained by redraw closures.
func (a *Axes3D) observe3DContourf(x, y []float64, z [][]float64, opt core.PlotOptions) bool {
	if a == nil || len(z) == 0 {
		return false
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return false
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return false
		}
	}

	levels := contourLevels(flattenGridValues(z), opt.Levels, opt.LevelCount, true)
	if len(levels) < 2 {
		return a.observe3DGrid(x, y, z)
	}
	midpoints := make([]float64, 0, len(levels)-1)
	for i := 0; i+1 < len(levels); i++ {
		midpoints = append(midpoints, levels[i]+(levels[i+1]-levels[i])*0.5)
	}

	minX, maxX := finiteRange(x[:cols])
	minY, maxY := finiteRange(y[:rows])
	minZ, maxZ := zGridRange(z)
	minLevel, maxLevel := finiteRange(midpoints)
	if !isFinite(minX) || !isFinite(maxX) || !isFinite(minY) || !isFinite(maxY) || !isFinite(minZ) || !isFinite(maxZ) || !isFinite(minLevel) || !isFinite(maxLevel) {
		return false
	}

	minPoint := vec3{minX, minY, minZ}
	maxPoint := vec3{maxX, maxY, maxZ}
	switch normalized3DDir(opt.ZDir) {
	case "x":
		minPoint[0], maxPoint[0] = minLevel, maxLevel
	case "y":
		minPoint[1], maxPoint[1] = minLevel, maxLevel
	default:
		minPoint[2], maxPoint[2] = minLevel, maxLevel
	}

	changed := a.observe3DPoint(minPoint[0], minPoint[1], minPoint[2])
	if a.observe3DPoint(maxPoint[0], maxPoint[1], maxPoint[2]) {
		changed = true
	}
	return changed
}

//nolint:gocritic // Triangulation and PlotOptions remain value snapshots throughout observation.
func (a *Axes3D) observe3DTriContourf(tri core.Triangulation, z []float64, opt core.PlotOptions) bool {
	if a == nil || len(tri.X) == 0 || len(tri.X) != len(tri.Y) || len(z) != len(tri.X) {
		return false
	}
	zdir := normalized3DDir(opt.ZDir)
	rotatedValues := z
	if zdir != "z" {
		rotatedValues = make([]float64, len(z))
		for i := range z {
			p := rotate3DPointAxes(tri.X[i], tri.Y[i], z[i], zdir)
			rotatedValues[i] = p[2]
		}
	}
	levels := contourLevels(rotatedValues, opt.Levels, opt.LevelCount, true)
	if len(levels) < 2 {
		return a.observe3DTriangulation(tri, z)
	}
	midpoints := make([]float64, 0, len(levels)-1)
	for i := 0; i+1 < len(levels); i++ {
		midpoints = append(midpoints, levels[i]+(levels[i+1]-levels[i])*0.5)
	}

	minX, maxX := finiteRange(tri.X)
	minY, maxY := finiteRange(tri.Y)
	minZ, maxZ := finiteRange(z)
	minLevel, maxLevel := finiteRange(midpoints)
	if !isFinite(minX) || !isFinite(maxX) || !isFinite(minY) || !isFinite(maxY) || !isFinite(minZ) || !isFinite(maxZ) || !isFinite(minLevel) || !isFinite(maxLevel) {
		return false
	}

	minPoint := vec3{minX, minY, minZ}
	maxPoint := vec3{maxX, maxY, maxZ}
	switch zdir {
	case "x":
		minPoint[0], maxPoint[0] = minLevel, maxLevel
	case "y":
		minPoint[1], maxPoint[1] = minLevel, maxLevel
	default:
		minPoint[2], maxPoint[2] = minLevel, maxLevel
	}

	changed := a.observe3DPoint(minPoint[0], minPoint[1], minPoint[2])
	if a.observe3DPoint(maxPoint[0], maxPoint[1], maxPoint[2]) {
		changed = true
	}
	return changed
}

//nolint:gocritic // Triangulation is a value-semantic geometry input shared with core.
func (a *Axes3D) observe3DTriangulation(tri core.Triangulation, z []float64) bool {
	n := len(tri.X)
	if len(tri.Y) < n {
		n = len(tri.Y)
	}
	if len(z) < n {
		n = len(z)
	}
	changed := false
	for i := 0; i < n; i++ {
		if a.observe3DPoint(tri.X[i], tri.Y[i], z[i]) {
			changed = true
		}
	}
	return changed
}

func (a *Axes3D) observe3DBarData(x, y, z, dx, dy, dz []float64) bool {
	n := minLen(x, y, z, dx, dy, dz)
	changed := false
	for i := 0; i < n; i++ {
		x0, x1 := x[i], x[i]+dx[i]
		y0, y1 := y[i], y[i]+dy[i]
		z0, z1 := z[i], z[i]+dz[i]
		if x1 < x0 {
			x0, x1 = x1, x0
		}
		if y1 < y0 {
			y0, y1 = y1, y0
		}
		if z1 < z0 {
			z0, z1 = z1, z0
		}
		if a.observe3DPoint(x0, y0, z0) {
			changed = true
		}
		if a.observe3DPoint(x1, y1, z1) {
			changed = true
		}
	}
	return changed
}

func (a *Axes3D) add3DReprojector(reproject func(), limitsChanged bool) {
	if a == nil || reproject == nil {
		return
	}
	a.reprojectors = append(a.reprojectors, reproject)
	if limitsChanged {
		a.reproject3DArtists()
	}
}

func (a *Axes3D) reproject3DArtists() {
	if a == nil {
		return
	}
	a.InvalidateArtistOrder()
	for _, reproject := range a.reprojectors {
		reproject()
	}
}

func isFinite3D(x, y, z float64) bool {
	return !math.IsNaN(x) && !math.IsNaN(y) && !math.IsNaN(z) &&
		!math.IsInf(x, 0) && !math.IsInf(y, 0) && !math.IsInf(z, 0)
}

func minLen(slices ...[]float64) int {
	if len(slices) == 0 {
		return 0
	}
	n := len(slices[0])
	for _, s := range slices[1:] {
		if len(s) < n {
			n = len(s)
		}
	}
	return n
}

func repeatColor(color render.Color, n int) []render.Color {
	if n <= 0 {
		return nil
	}
	colors := make([]render.Color, n)
	for i := range colors {
		colors[i] = color
	}
	return colors
}

func faceColorAtIndex(colors []render.Color, idx int) render.Color {
	if len(colors) == 0 {
		return render.Color{}
	}
	if len(colors) == 1 {
		return colors[0]
	}
	if idx < len(colors) {
		return colors[idx]
	}
	return colors[len(colors)-1]
}

//nolint:gocritic // Shade selection reads a value-semantic PlotOptions snapshot.
func surfaceShadeEnabled(opt core.PlotOptions, useMapping bool) bool {
	if shade, ok := opt.Shade.Get(); ok {
		return shade
	}
	return !useMapping
}

//nolint:gocritic // Edge-color resolution reads a value-semantic PlotOptions snapshot.
func surfaceEdgeColors(faceColors []render.Color, opt core.PlotOptions) []render.Color {
	if opt.EdgeColor.IsSet() || len(opt.FaceColors) == 0 {
		return nil
	}
	edges := make([]render.Color, len(faceColors))
	copy(edges, faceColors)
	return edges
}

//nolint:gocritic // Shade selection reads a value-semantic PlotOptions snapshot.
func trisurfShadeEnabled(opt core.PlotOptions, useMapping bool) bool {
	if shade, ok := opt.Shade.Get(); ok {
		return shade
	}
	return !useMapping
}

//nolint:gocritic // Scalar-map resolution reads a value-semantic PlotOptions snapshot.
func resolvePlotScalarMap(values []float64, opt core.PlotOptions) core.ScalarMapInfo {
	cfg := opt.ScalarMapConfig()
	mapping, err := core.ResolveScalarMapValues(values, cfg)
	if err != nil {
		return core.ScalarMapInfo{Colormap: cfg.Colormap}.Resolved()
	}
	return mapping
}

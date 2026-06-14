package core

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Streamplot adds a streamline set over a rectilinear vector grid.
func (a *Axes) Streamplot(x, y []float64, u, v [][]float64, opts ...StreamplotOptions) *StreamplotSet {
	if a == nil {
		return nil
	}
	if !vectorStrictlyIncreasing(x) || !vectorStrictlyIncreasing(y) {
		return nil
	}
	if !sameGridShape(u, len(y), len(x)) || !sameGridShape(v, len(y), len(x)) {
		return nil
	}

	opt := StreamplotOptions{
		Density:              1,
		IntegrationDirection: streamDirectionBoth,
	}
	if len(opts) > 0 {
		opt = opts[0]
	}
	densityX, densityY := resolvedStreamDensity(opt)
	if densityX <= 0 || densityY <= 0 {
		return nil
	}

	grid := streamplotGrid{
		x: append([]float64(nil), x...),
		y: append([]float64(nil), y...),
		u: cloneMatrix(u),
		v: cloneMatrix(v),
	}
	var mapping ScalarMapInfo
	hasScalarColors := len(opt.CGrid) > 0
	if hasScalarColors {
		if !sameGridShape(opt.CGrid, len(y), len(x)) {
			return nil
		}
		var err error
		mapping, err = ResolveScalarMapGrid(opt.CGrid, ScalarMapConfig{
			Colormap: scalarColormap(opt.Colormap),
			Norm:     opt.Norm,
			VMin:     opt.VMin,
			VMax:     opt.VMax,
		})
		if err != nil {
			return nil
		}
	}
	trajectories := computeStreamTrajectories(grid, opt, densityX, densityY)
	if len(trajectories) == 0 {
		return nil
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	lineWidth := optionFloat(opt.LineWidth, 1.5)
	z := optionFloat(opt.ZOrder, 1)

	lines := &LineCollection{
		Collection: Collection{
			Coords:   Coords(CoordData),
			Colormap: mapping.Colormap,
			Norm:     mapping.Norm,
			VMin:     mapping.VMin,
			VMax:     mapping.VMax,
			z:        z,
		},
		Color:     color,
		LineWidth: lineWidth,
		Segments:  make([][]geom.Pt, 0, len(trajectories)),
	}
	for _, trajectory := range trajectories {
		lines.Segments = append(lines.Segments, trajectory.points)
		if hasScalarColors {
			value, ok := streamScalarAverage(grid, opt.CGrid, trajectory.points)
			if !ok {
				value = math.NaN()
			}
			lines.Colors = append(lines.Colors, mapping.Color(value, 1))
		}
	}
	if hasScalarColors {
		lines.Color = render.Color{}
	}

	arrowColor := color
	if opt.ArrowColor != nil {
		arrowColor = *opt.ArrowColor
	}
	arrowSize := optionFloat(opt.ArrowSize, 1)
	arrowCount := optionInt(opt.ArrowCount, 1)
	if arrowCount < 0 {
		return nil
	}

	arrowAnchors, arrowU, arrowV := sampleStreamArrows(trajectories, arrowCount)
	arrowScalars := []float64(nil)
	if hasScalarColors {
		arrowScalars = make([]float64, len(arrowAnchors))
		for i, anchor := range arrowAnchors {
			value, ok := interpolateStreamScalar(grid, opt.CGrid, anchor)
			if !ok {
				value = math.NaN()
			}
			arrowScalars[i] = value
		}
	}
	arrows := &Quiver{
		Anchors:        arrowAnchors,
		U:              arrowU,
		V:              arrowV,
		ScalarColors:   arrowScalars,
		Color:          arrowColor,
		EdgeColor:      render.Color{},
		EdgeWidth:      0,
		Alpha:          1,
		Pivot:          vectorPivotMiddle,
		Angles:         quiverAnglesXY,
		Units:          "dots",
		ScaleUnits:     "dots",
		HeadWidth:      3,
		HeadLength:     5,
		HeadAxisLength: 4.5,
		MinShaft:       1,
		MinLength:      1,
		Colormap:       mapping.Colormap,
		Norm:           mapping.Norm,
		VMin:           mapping.VMin,
		VMax:           mapping.VMax,
		z:              z,
		forceLengthPx:  12 * arrowSize,
	}
	if lineWidth > 0 {
		arrows.Width = math.Max(1, lineWidth*1.8)
	}

	set := &StreamplotSet{
		Lines:  lines,
		Arrows: arrows,
		Label:  opt.Label,
		z:      z,
	}
	a.Add(set)
	return set
}

// Draw renders the streamline line collection and its arrows.
func (s *StreamplotSet) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil {
		return
	}
	if s.Lines != nil {
		s.Lines.Draw(r, ctx)
	}
	if s.Arrows != nil {
		s.Arrows.Draw(r, ctx)
	}
}

// Bounds returns the line extents for autoscaling.
func (s *StreamplotSet) Bounds(ctx *DrawContext) geom.Rect {
	if s == nil || s.Lines == nil {
		return geom.Rect{}
	}
	return s.Lines.Bounds(ctx)
}

// Z returns the draw order.
func (s *StreamplotSet) Z() float64 {
	if s == nil {
		return 0
	}
	return s.z
}

// ScalarMap exposes scalar color mapping when streamplot arrows or lines carry one.
func (s *StreamplotSet) ScalarMap() ScalarMapInfo {
	if s == nil {
		return ScalarMapInfo{}
	}
	if s.Arrows != nil {
		if mapping := s.Arrows.ScalarMap(); scalarMapConfigured(mapping) {
			return mapping
		}
	}
	if s.Lines != nil && len(s.Lines.Colors) > 0 {
		if mapping := s.Lines.ScalarMap(); scalarMapConfigured(mapping) {
			return mapping
		}
	}
	return ScalarMapInfo{}
}

func (s *StreamplotSet) legendEntry() (legendEntry, bool) {
	if s == nil || s.Label == "" || s.Lines == nil {
		return legendEntry{}, false
	}
	return legendEntryFromLine(s.Label, s.Lines.Color, s.Lines.LineWidth, nil), true
}

func computeStreamTrajectories(grid streamplotGrid, opt StreamplotOptions, densityX, densityY float64) []streamTrajectory {
	mask := newStreamplotMask(grid.x[0], grid.x[len(grid.x)-1], grid.y[0], grid.y[len(grid.y)-1], densityX, densityY)
	minLength := optionFloat(opt.MinLength, 0.08)
	maxLength := optionFloat(opt.MaxLength, 4.0)
	if minLength <= 0 || maxLength <= 0 {
		return nil
	}
	broken := optionBool(opt.BrokenStreamlines, true)
	direction := normalizeStreamDirection(opt.IntegrationDirection)
	step := 0.35 * math.Min(minSpacing(grid.x), minSpacing(grid.y))
	if step <= 0 || !isFinite(step) {
		return nil
	}

	starts := opt.StartPoints
	if len(starts) == 0 {
		starts = mask.seedPoints()
	}

	trajectories := make([]streamTrajectory, 0, len(starts))
	for _, start := range starts {
		if !pointInsideGrid(start, grid) {
			continue
		}
		traj, used := integrateStream(grid, mask, start, step, minLength, maxLength, direction, broken)
		if len(traj.points) < 2 {
			continue
		}
		for idx := range used {
			mask.used[idx] = true
		}
		trajectories = append(trajectories, traj)
	}
	return trajectories
}

func integrateStream(grid streamplotGrid, mask *streamplotMask, start geom.Pt, step, minLength, maxLength float64, direction string, broken bool) (streamTrajectory, map[int]struct{}) {
	switch direction {
	case streamDirectionBackward:
		points, used := streamDirection(grid, mask, start, step, maxLength, -1, broken)
		if normalizedPathLength(points, grid) < minLength {
			return streamTrajectory{}, nil
		}
		return streamTrajectory{points: points}, used
	case streamDirectionForward:
		points, used := streamDirection(grid, mask, start, step, maxLength, 1, broken)
		if normalizedPathLength(points, grid) < minLength {
			return streamTrajectory{}, nil
		}
		return streamTrajectory{points: points}, used
	default:
		backward, usedBack := streamDirection(grid, mask, start, step, maxLength*0.5, -1, broken)
		forward, usedForward := streamDirection(grid, mask, start, step, maxLength*0.5, 1, broken)
		if len(backward) == 0 && len(forward) == 0 {
			return streamTrajectory{}, nil
		}
		points := make([]geom.Pt, 0, len(backward)+len(forward))
		for i := len(backward) - 1; i >= 0; i-- {
			points = append(points, backward[i])
		}
		if len(forward) > 0 {
			points = append(points, forward[1:]...)
		}
		if normalizedPathLength(points, grid) < minLength {
			return streamTrajectory{}, nil
		}
		used := map[int]struct{}{}
		for idx := range usedBack {
			used[idx] = struct{}{}
		}
		for idx := range usedForward {
			used[idx] = struct{}{}
		}
		return streamTrajectory{points: points}, used
	}
}

func streamDirection(grid streamplotGrid, mask *streamplotMask, start geom.Pt, step, maxLength float64, sign float64, broken bool) ([]geom.Pt, map[int]struct{}) {
	if !pointInsideGrid(start, grid) {
		return nil, nil
	}
	points := []geom.Pt{start}
	used := map[int]struct{}{mask.index(start): {}}
	total := 0.0
	current := start

	for total < maxLength {
		next, ok := streamStep(grid, current, step*sign)
		if !ok || !pointInsideGrid(next, grid) {
			break
		}
		idx := mask.index(next)
		if broken {
			if mask.used[idx] {
				break
			}
			if _, ok := used[idx]; ok {
				break
			}
		}
		total += normalizedSegmentLength(current, next, grid)
		if total > maxLength {
			break
		}
		points = append(points, next)
		used[idx] = struct{}{}
		current = next
	}
	return points, used
}

func streamStep(grid streamplotGrid, point geom.Pt, step float64) (geom.Pt, bool) {
	u1, v1, ok := interpolateStreamVector(grid, point)
	if !ok {
		return geom.Pt{}, false
	}
	mag1 := math.Hypot(u1, v1)
	if mag1 == 0 {
		return geom.Pt{}, false
	}
	mid := geom.Pt{
		X: point.X + (u1/mag1)*(step*0.5),
		Y: point.Y + (v1/mag1)*(step*0.5),
	}
	u2, v2, ok := interpolateStreamVector(grid, mid)
	if !ok {
		return geom.Pt{}, false
	}
	mag2 := math.Hypot(u2, v2)
	if mag2 == 0 {
		return geom.Pt{}, false
	}
	return geom.Pt{
		X: point.X + (u2/mag2)*step,
		Y: point.Y + (v2/mag2)*step,
	}, true
}

func interpolateStreamVector(grid streamplotGrid, point geom.Pt) (float64, float64, bool) {
	if !pointInsideGrid(point, grid) {
		return 0, 0, false
	}
	ix := locateInterval(grid.x, point.X)
	iy := locateInterval(grid.y, point.Y)
	if ix < 0 || iy < 0 {
		return 0, 0, false
	}
	x0, x1 := grid.x[ix], grid.x[ix+1]
	y0, y1 := grid.y[iy], grid.y[iy+1]
	tx := 0.0
	if x1 != x0 {
		tx = (point.X - x0) / (x1 - x0)
	}
	ty := 0.0
	if y1 != y0 {
		ty = (point.Y - y0) / (y1 - y0)
	}
	u00, u10 := grid.u[iy][ix], grid.u[iy][ix+1]
	u01, u11 := grid.u[iy+1][ix], grid.u[iy+1][ix+1]
	v00, v10 := grid.v[iy][ix], grid.v[iy][ix+1]
	v01, v11 := grid.v[iy+1][ix], grid.v[iy+1][ix+1]
	if !isFinite(u00) || !isFinite(u10) || !isFinite(u01) || !isFinite(u11) ||
		!isFinite(v00) || !isFinite(v10) || !isFinite(v01) || !isFinite(v11) {
		return 0, 0, false
	}
	u0 := u00*(1-tx) + u10*tx
	u1 := u01*(1-tx) + u11*tx
	v0 := v00*(1-tx) + v10*tx
	v1 := v01*(1-tx) + v11*tx
	return u0*(1-ty) + u1*ty, v0*(1-ty) + v1*ty, true
}

func streamScalarAverage(grid streamplotGrid, scalars [][]float64, points []geom.Pt) (float64, bool) {
	sum := 0.0
	count := 0
	for _, point := range points {
		value, ok := interpolateStreamScalar(grid, scalars, point)
		if !ok {
			continue
		}
		sum += value
		count++
	}
	if count == 0 {
		return 0, false
	}
	return sum / float64(count), true
}

func interpolateStreamScalar(grid streamplotGrid, scalars [][]float64, point geom.Pt) (float64, bool) {
	if !pointInsideGrid(point, grid) || !sameGridShape(scalars, len(grid.y), len(grid.x)) {
		return 0, false
	}
	ix := locateInterval(grid.x, point.X)
	iy := locateInterval(grid.y, point.Y)
	if ix < 0 || iy < 0 {
		return 0, false
	}
	x0, x1 := grid.x[ix], grid.x[ix+1]
	y0, y1 := grid.y[iy], grid.y[iy+1]
	tx := 0.0
	if x1 != x0 {
		tx = (point.X - x0) / (x1 - x0)
	}
	ty := 0.0
	if y1 != y0 {
		ty = (point.Y - y0) / (y1 - y0)
	}
	c00, c10 := scalars[iy][ix], scalars[iy][ix+1]
	c01, c11 := scalars[iy+1][ix], scalars[iy+1][ix+1]
	if !isFinite(c00) || !isFinite(c10) || !isFinite(c01) || !isFinite(c11) {
		return 0, false
	}
	c0 := c00*(1-tx) + c10*tx
	c1 := c01*(1-tx) + c11*tx
	return c0*(1-ty) + c1*ty, true
}

func pointInsideGrid(point geom.Pt, grid streamplotGrid) bool {
	return point.X >= grid.x[0] && point.X <= grid.x[len(grid.x)-1] &&
		point.Y >= grid.y[0] && point.Y <= grid.y[len(grid.y)-1]
}

func locateInterval(values []float64, target float64) int {
	if len(values) < 2 || target < values[0] || target > values[len(values)-1] {
		return -1
	}
	if target == values[len(values)-1] {
		return len(values) - 2
	}
	idx := sort.Search(len(values)-1, func(i int) bool { return values[i+1] > target })
	if idx < 0 || idx >= len(values)-1 {
		return -1
	}
	return idx
}

func normalizedPathLength(points []geom.Pt, grid streamplotGrid) float64 {
	total := 0.0
	for i := 1; i < len(points); i++ {
		total += normalizedSegmentLength(points[i-1], points[i], grid)
	}
	return total
}

func normalizedSegmentLength(a, b geom.Pt, grid streamplotGrid) float64 {
	xspan := grid.x[len(grid.x)-1] - grid.x[0]
	yspan := grid.y[len(grid.y)-1] - grid.y[0]
	if xspan == 0 {
		xspan = 1
	}
	if yspan == 0 {
		yspan = 1
	}
	return math.Hypot((b.X-a.X)/xspan, (b.Y-a.Y)/yspan)
}

func sampleStreamArrows(trajectories []streamTrajectory, count int) ([]geom.Pt, []float64, []float64) {
	if count <= 0 {
		return nil, nil, nil
	}
	var anchors []geom.Pt
	var u []float64
	var v []float64
	for _, trajectory := range trajectories {
		if len(trajectory.points) < 2 {
			continue
		}
		cum := make([]float64, len(trajectory.points))
		for i := 1; i < len(trajectory.points); i++ {
			cum[i] = cum[i-1] + math.Hypot(
				trajectory.points[i].X-trajectory.points[i-1].X,
				trajectory.points[i].Y-trajectory.points[i-1].Y,
			)
		}
		total := cum[len(cum)-1]
		if total == 0 {
			continue
		}
		for arrow := 0; arrow < count; arrow++ {
			target := total * float64(arrow+1) / float64(count+1)
			idx := sort.Search(len(cum), func(i int) bool { return cum[i] >= target })
			if idx <= 0 || idx >= len(cum) {
				continue
			}
			prev := trajectory.points[idx-1]
			next := trajectory.points[idx]
			segment := cum[idx] - cum[idx-1]
			t := 0.0
			if segment > 0 {
				t = (target - cum[idx-1]) / segment
			}
			point := geom.Pt{
				X: prev.X + (next.X-prev.X)*t,
				Y: prev.Y + (next.Y-prev.Y)*t,
			}
			anchors = append(anchors, point)
			u = append(u, next.X-prev.X)
			v = append(v, next.Y-prev.Y)
		}
	}
	return anchors, u, v
}

func newStreamplotMask(xmin, xmax, ymin, ymax, densityX, densityY float64) *streamplotMask {
	nx := int(math.Round(30 * densityX))
	ny := int(math.Round(30 * densityY))
	if nx < 1 {
		nx = 1
	}
	if ny < 1 {
		ny = 1
	}
	return &streamplotMask{
		nx:    nx,
		ny:    ny,
		used:  make([]bool, nx*ny),
		xmin:  xmin,
		xspan: xmax - xmin,
		ymin:  ymin,
		yspan: ymax - ymin,
	}
}

func (m *streamplotMask) seedPoints() []geom.Pt {
	points := make([]geom.Pt, 0, m.nx*m.ny)
	for yi := 0; yi < m.ny; yi++ {
		for xi := 0; xi < m.nx; xi++ {
			x := m.xmin + (float64(xi)+0.5)/float64(m.nx)*m.xspan
			y := m.ymin + (float64(yi)+0.5)/float64(m.ny)*m.yspan
			points = append(points, geom.Pt{X: x, Y: y})
		}
	}
	return points
}

func (m *streamplotMask) index(point geom.Pt) int {
	xn := clamp01((point.X - m.xmin) / maxF(m.xspan, 1e-9))
	yn := clamp01((point.Y - m.ymin) / maxF(m.yspan, 1e-9))
	xi := int(math.Min(float64(m.nx-1), math.Floor(xn*float64(m.nx))))
	yi := int(math.Min(float64(m.ny-1), math.Floor(yn*float64(m.ny))))
	return yi*m.nx + xi
}

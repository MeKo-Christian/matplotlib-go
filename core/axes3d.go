package core

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	default3DAzimuthDeg      = -60
	default3DElevationDeg    = 30
	default3DDistance        = 10
	default3DFocalLength     = 1
	default3DRollDeg         = 0
	default3DVerticalAxis    = 2
	default3DComputedZ       = 2.5
	default3DSurfaceCount    = 50
	default3DBoxAspectScale  = 1.8294640721620434
	default3DBoxAspectZoom25 = 25.0 / 24.0
	default3DViewMin         = -0.095
	default3DViewMax         = 0.09
)

// Axes3D represents an Axes with basic 3D projection helpers.
//
// The underlying artist model is still 2D (`*Axes`) with pre-projected
// 3D coordinates converted into data-space 2D points before drawing.
type Axes3D struct {
	*Axes
	azimuthDeg   float64
	elevationDeg float64
	rollDeg      float64
	verticalAxis int
	zLabel       string
	showXLabels  bool
	showYLabels  bool
	showZLabels  bool
	distance     float64
	boxAspect    vec3
	hasData      bool
	dataMin      vec3
	dataMax      vec3
	viewMin      vec3
	viewMax      vec3
	viewSet      [3]bool
	zMargin      float64
	reprojectors []func()
}

type axes3DFrame struct {
	axes *Axes3D
}

type projected3DState struct {
	rollDeg      float64
	verticalAxis int
	boxAspect    vec3
}

func (a *Axes3D) projectionState() projected3DState {
	if a == nil {
		return projected3DState{
			rollDeg:      default3DRollDeg,
			verticalAxis: default3DVerticalAxis,
			boxAspect:    default3DBoxAspect(),
		}
	}
	return projected3DState{
		rollDeg:      a.rollDeg,
		verticalAxis: a.verticalAxis,
		boxAspect:    a.boxAspect,
	}
}

func (a *Axes3D) project3DPointWithState(x, y, z float64, mins, maxs vec3) geom.Pt {
	state := a.projectionState()
	if a == nil {
		return project3DPointWithLimits(x, y, z, default3DElevationDeg, default3DAzimuthDeg, default3DDistance, mins, maxs, state)
	}
	return project3DPointWithLimits(x, y, z, a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs, state)
}

func projected3DStateOrDefault(state ...projected3DState) projected3DState {
	if len(state) > 0 {
		s := state[0]
		if s.verticalAxis < 0 || s.verticalAxis > 2 {
			s.verticalAxis = default3DVerticalAxis
		}
		if s.boxAspect[0] == 0 && s.boxAspect[1] == 0 && s.boxAspect[2] == 0 {
			s.boxAspect = default3DBoxAspect()
		}
		return s
	}
	return projected3DState{
		rollDeg:      default3DRollDeg,
		verticalAxis: default3DVerticalAxis,
		boxAspect:    default3DBoxAspect(),
	}
}

func parse3DVerticalAxis(axis string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(axis)) {
	case "x", "xaxis":
		return 0, nil
	case "y", "yaxis":
		return 1, nil
	case "z", "zaxis":
		return 2, nil
	}
	return default3DVerticalAxis, fmt.Errorf("invalid vertical axis %q", axis)
}

func normalize3DAngleDeg(angle float64) float64 {
	a := math.Mod(angle, 360)
	if a <= -180 {
		a += 360
	}
	if a > 180 {
		a -= 360
	}
	return a
}

func rollToVertical(v vec3, axis int, reverse bool) vec3 {
	shift := axis - 2
	if reverse {
		shift *= -1
	}
	if shift == 0 {
		return v
	}
	shift = ((shift % 3) + 3) % 3
	var out vec3
	for i := range 3 {
		out[i] = v[(i-shift+3)%3]
	}
	return out
}

func viewAxes(eye, center vec3, elevationDeg float64, verticalAxis int, rollDeg float64) (vec3, vec3, vec3) {
	w := eye.sub(center).unit()
	vertical := vec3{}
	if verticalAxis >= 0 && verticalAxis < 3 {
		vertical[verticalAxis] = 1
	}
	elevRad := normalize3DAngleDeg(elevationDeg) * math.Pi / 180
	if math.Abs(elevRad) > math.Pi/2 {
		vertical[verticalAxis] = -vertical[verticalAxis]
	}
	u := vertical.cross(w).unit()
	v := w.cross(u)
	if rollDeg == 0 {
		return u, v, w
	}
	rollRad := normalize3DAngleDeg(rollDeg) * math.Pi / 180
	u = rotateVecAroundAxis(u, w, -rollRad)
	v = rotateVecAroundAxis(v, w, -rollRad)
	return u, v, w
}

func rotateVecAroundAxis(v, axis vec3, angle float64) vec3 {
	axisNorm := axis.norm()
	if axisNorm == 0 {
		return v
	}
	if angle == 0 {
		return v
	}
	axis = axis.scale(1 / axisNorm)
	c := math.Cos(angle)
	s := math.Sin(angle)
	t := 2 * math.Sin(angle/2) * math.Sin(angle/2)
	vx, vy, vz := v[0], v[1], v[2]
	ux, uy, uz := axis[0], axis[1], axis[2]
	return vec3{
		(t*ux*ux+c)*vx + (t*ux*uy-uz*s)*vy + (t*ux*uz+uy*s)*vz,
		(t*uy*ux+uz*s)*vx + (t*uy*uy+c)*vy + (t*uy*uz-ux*s)*vz,
		(t*uz*ux-uy*s)*vx + (t*uz*uy+ux*s)*vy + (t*uz*uz+c)*vz,
	}
}

func (f *axes3DFrame) Draw(r render.Renderer, ctx *DrawContext) {
	if f == nil || f.axes == nil || r == nil || ctx == nil {
		return
	}
	mins, maxs := f.axes.projectionLimits()
	if !f.axes.hasData {
		mins, maxs = vec3{0, 0, 0}, vec3{1, 1, 1}
	}
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	gridLineWidth := 0.8
	axisLineWidth := 0.8
	gridColor := render.Color{R: 0.70, G: 0.70, B: 0.70, A: 1}
	axisColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	if ctx.RC.GridLineWidth > 0 {
		gridLineWidth = ctx.RC.GridLineWidth
	}
	if ctx.RC.AxisLineWidth > 0 {
		axisLineWidth = ctx.RC.AxisLineWidth
	}
	if ctx.RC.GridColor.A > 0 {
		gridColor = ctx.RC.GridColor
	}
	if ctx.RC.AxesEdgeColor.A > 0 {
		axisColor = ctx.RC.AxesEdgeColor
	}

	panes := f.axes.activePanePolygonsProjected(frameMins, frameMaxs, mins, maxs)
	(&PolyCollection{
		Polygons: panes,
		PatchCollection: PatchCollection{
			Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
			FaceColors: axes3DPaneFaceColors(),
			EdgeColor:  render.Color{A: 0},
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}).Draw(r, ctx)

	segments := f.axes.frameSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs)
	(&LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
		Segments:   segments,
		Color:      gridColor,
		LineWidth:  gridLineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}).Draw(r, ctx)

	axisLines := f.axes.axisLineSegmentsProjected(frameMins, frameMaxs, mins, maxs)
	(&LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
		Segments:   axisLines,
		Color:      axisColor,
		LineWidth:  axisLineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}).Draw(r, ctx)

	tickSegments := f.axes.axisTickSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs)
	(&LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
		Segments:   tickSegments,
		Color:      axisColor,
		LineWidth:  axisLineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}).Draw(r, ctx)

	if textRen, ok := r.(render.TextDrawer); ok {
		f.axes.draw3DTickLabels(textRen, r, ctx, frameMins, frameMaxs, mins, maxs)
		f.axes.draw3DAxisLabels(textRen, r, ctx, frameMins, frameMaxs)
	}
}

func (f *axes3DFrame) Z() float64 { return -1000 }

func (f *axes3DFrame) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func axes3DPaneFaceColors() []render.Color {
	return []render.Color{
		{R: 0.95, G: 0.95, B: 0.95, A: 0.5},
		{R: 0.90, G: 0.90, B: 0.90, A: 0.5},
		{R: 0.925, G: 0.925, B: 0.925, A: 0.5},
	}
}

func (f *axes3DFrame) DrawOverlay(r render.Renderer, ctx *DrawContext) {
}

// NewAxes3D wraps an existing axes and configures 3D default view settings.
func NewAxes3D(ax *Axes) *Axes3D {
	if ax == nil {
		return nil
	}
	axes := &Axes3D{
		Axes:         ax,
		azimuthDeg:   default3DAzimuthDeg,
		elevationDeg: default3DElevationDeg,
		rollDeg:      default3DRollDeg,
		verticalAxis: default3DVerticalAxis,
		showXLabels:  true,
		showYLabels:  true,
		showZLabels:  true,
		distance:     default3DDistance,
		boxAspect:    default3DBoxAspect(),
	}
	ax.Add(&axes3DFrame{axes: axes})
	return axes
}

// Plot3D projects x/y/z values and draws a line through projected points.
func (a *Axes3D) Plot3D(x, y, z []float64, opts ...PlotOptions) *Line2D {
	limitsChanged := a.observe3DData(x, y, z)
	projected := a.projectedData(x, y, z)
	if len(projected) == 0 {
		return nil
	}

	x2, y2 := make([]float64, len(projected)), make([]float64, len(projected))
	for i, p := range projected {
		x2[i] = p.X
		y2[i] = p.Y
	}

	if len(opts) > 0 {
		line := a.Plot(x2, y2, opts[0])
		a.add3DReprojector(func() {
			reprojectLine3D(line, a.projectedData(x, y, z))
		}, limitsChanged)
		return line
	}
	line := a.Plot(x2, y2)
	a.add3DReprojector(func() {
		reprojectLine3D(line, a.projectedData(x, y, z))
	}, limitsChanged)
	return line
}

// Scatter3D projects x/y/z values and draws markers through projected points.
func (a *Axes3D) Scatter3D(x, y, z []float64, opts ...ScatterOptions) *Scatter2D {
	limitsChanged := a.observe3DData(x, y, z)
	if a.ensure3DZMargin(0.05) {
		limitsChanged = true
	}
	projected := a.projectedData(x, y, z)
	if len(projected) == 0 {
		return nil
	}

	x2, y2 := make([]float64, len(projected)), make([]float64, len(projected))
	for i, p := range projected {
		x2[i] = p.X
		y2[i] = p.Y
	}

	if len(opts) > 0 {
		scatter := a.Scatter(x2, y2, opts[0])
		reprojectScatter3D(scatter, a.projectedScatterData(x, y, z))
		scatter.z = a.points3DCollectionZ(x, y, z)
		a.add3DReprojector(func() {
			reprojectScatter3D(scatter, a.projectedScatterData(x, y, z))
			scatter.z = a.points3DCollectionZ(x, y, z)
		}, limitsChanged)
		return scatter
	}
	scatter := a.Scatter(x2, y2)
	reprojectScatter3D(scatter, a.projectedScatterData(x, y, z))
	scatter.z = a.points3DCollectionZ(x, y, z)
	a.add3DReprojector(func() {
		reprojectScatter3D(scatter, a.projectedScatterData(x, y, z))
		scatter.z = a.points3DCollectionZ(x, y, z)
	}, limitsChanged)
	return scatter
}

// PlotSurface creates a placeholder mesh-like line-strip by plotting the input
// sequence as connected edges.
func (a *Axes3D) PlotSurface(x, y, z []float64, opts ...PlotOptions) *Line2D {
	return a.Plot3D(x, y, z, opts...)
}

// Text3D projects a point and renders arbitrary text at the projected location.
func (a *Axes3D) Text3D(x, y, z float64, text string, opts ...TextOptions) *Text {
	if a == nil || a.Axes == nil {
		return nil
	}
	limitsChanged := a.observe3DPoint(x, y, z)
	p := a.ProjectPoint(x, y, z)
	txt := a.Text(p.X, p.Y, text, opts...)
	a.add3DReprojector(func() {
		if txt != nil {
			txt.Position = a.ProjectPoint(x, y, z)
		}
	}, limitsChanged)
	return txt
}

// Stem3DOptions configures Axes3D.Stem3D.
type Stem3DOptions struct {
	Color           *render.Color
	LineWidth       *float64
	Marker          *MarkerType
	MarkerPath      *geom.Path
	MarkerSize      *float64
	Bottom          *float64
	Orientation     string
	BaselineColor   *render.Color
	BaselineWidth   *float64
	MarkerEdgeColor *render.Color
	MarkerEdgeWidth *float64
	Label           string
	Alpha           *float64
}

// FillBetween3DMode controls the polygon construction for 3D fill bands.
type FillBetween3DMode string

const (
	FillBetween3DModeAuto    FillBetween3DMode = "auto"
	FillBetween3DModeQuad    FillBetween3DMode = "quad"
	FillBetween3DModePolygon FillBetween3DMode = "polygon"
)

// FillBetween3DOptions configures Axes3D.FillBetween3D.
type FillBetween3DOptions struct {
	Color     *render.Color
	EdgeColor *render.Color
	EdgeWidth *float64
	Alpha     *float64
	Label     string
	Mode      FillBetween3DMode
}

// Quiver3DOptions configures Axes3D.Quiver.
type Quiver3DOptions struct {
	Color            *render.Color
	LineWidth        *float64
	Alpha            *float64
	Label            string
	Length           *float64
	ArrowLengthRatio *float64
	Pivot            string
	Normalize        bool
	AxLimClip        bool
}

// ErrorBar3DOptions configures Axes3D.ErrorBar3D.
type ErrorBar3DOptions struct {
	Color     *render.Color
	LineWidth *float64
	CapSize   *float64
	Alpha     *float64
	Label     string

	XErrLower []float64
	XErrUpper []float64
	YErrLower []float64
	YErrUpper []float64
	ZErrLower []float64
	ZErrUpper []float64
}

// Stem3D renders Matplotlib-style 3D stem lines, head markers, and a baseline.
func (a *Axes3D) Stem3D(x, y, z []float64, opts ...Stem3DOptions) *StemContainer {
	if a == nil || a.Axes == nil {
		return nil
	}
	n := minLen(x, y, z)
	if n <= 0 {
		return nil
	}

	var opt Stem3DOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	orientation := normalized3DDir(opt.Orientation)
	bottom := 0.0
	if opt.Bottom != nil {
		bottom = *opt.Bottom
	}

	limitsChanged := a.observe3DStemData(x[:n], y[:n], z[:n], bottom, orientation)
	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	color.A *= alpha

	lineWidth := 1.5
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}
	markerType := MarkerCircle
	if opt.Marker != nil {
		markerType = *opt.Marker
	}
	markerSize := 6.0
	if opt.MarkerSize != nil {
		markerSize = *opt.MarkerSize
	}
	markerEdgeColor := color
	if opt.MarkerEdgeColor != nil {
		markerEdgeColor = *opt.MarkerEdgeColor
		markerEdgeColor.A *= alpha
	}
	markerEdgeWidth := lineWidth * 0.8
	if opt.MarkerEdgeWidth != nil {
		markerEdgeWidth = *opt.MarkerEdgeWidth
	}
	baselineColor := color
	if opt.BaselineColor != nil {
		baselineColor = *opt.BaselineColor
		baselineColor.A *= alpha
	}
	baselineWidth := lineWidth
	if opt.BaselineWidth != nil {
		baselineWidth = *opt.BaselineWidth
	}
	markerPath := geom.Path{}
	if opt.MarkerPath != nil {
		markerPath = *opt.MarkerPath
	}
	if len(markerPath.C) == 0 {
		scatter := Scatter2D{Marker: markerType}
		markerPath = scatter.markerPrototypePath()
	}

	segments, baseline, offsets, zorder := a.projectStem3DGeometry(x[:n], y[:n], z[:n], bottom, orientation)
	stems := &LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Label: opt.Label, Alpha: 1, z: zorder},
		Segments:   segments,
		Color:      color,
		LineWidth:  lineWidth,
		LineJoin:   render.JoinRound,
		LineCap:    render.CapRound,
	}
	markers := &PathCollection{
		Collection:    Collection{Coords: Coords(CoordData), Label: opt.Label, Alpha: 1, z: zorder + 0.05},
		Path:          markerPath,
		Offsets:       offsets,
		Size:          markerSize * stemMarkerScale,
		PathInDisplay: true,
		FaceColor:     color,
		EdgeColor:     markerEdgeColor,
		EdgeWidth:     markerEdgeWidth,
		LineOnly:      markerType == MarkerPlus || markerType == MarkerCross,
	}
	baselineArtist := &Line2D{
		XY:    baseline,
		W:     baselineWidth,
		Col:   baselineColor,
		Label: "",
		z:     zorder - 0.05,
	}

	a.AddCollection(stems)
	a.AddCollection(markers)
	a.Add(baselineArtist)
	a.add3DReprojector(func() {
		segments, baseline, offsets, zorder := a.projectStem3DGeometry(x[:n], y[:n], z[:n], bottom, orientation)
		stems.Segments = segments
		stems.z = zorder
		markers.Offsets = offsets
		markers.z = zorder + 0.05
		baselineArtist.XY = baseline
		baselineArtist.z = zorder - 0.05
	}, limitsChanged)

	return &StemContainer{
		MarkerCollection: markers,
		StemLines:        stems,
		Baseline:         baselineArtist,
		Label:            opt.Label,
	}
}

// Stem is the Matplotlib-compatible 3D stem entry point.
func (a *Axes3D) Stem(x, y, z []float64, opts ...Stem3DOptions) *StemContainer {
	return a.Stem3D(x, y, z, opts...)
}

// FillBetween3D fills bands between two 3D curves.
func (a *Axes3D) FillBetween3D(x1, y1, z1, x2, y2, z2 []float64, opts ...FillBetween3DOptions) *PolyCollection {
	if a == nil || a.Axes == nil {
		return nil
	}
	n := minLen(x1, y1, z1, x2, y2, z2)
	if n < 2 {
		return nil
	}

	var opt FillBetween3DOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	limitsChanged := a.observe3DData(x1[:n], y1[:n], z1[:n])
	if a.observe3DData(x2[:n], y2[:n], z2[:n]) {
		limitsChanged = true
	}

	color := a.NextPatchColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	color.A *= alpha
	edgeColor := render.Color{}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
		edgeColor.A *= alpha
	}
	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	polygons, zorder := a.projectFillBetween3DPolygons(x1[:n], y1[:n], z1[:n], x2[:n], y2[:n], z2[:n], opt.Mode)
	if len(polygons) == 0 {
		return nil
	}
	colors := repeatColor(color, len(polygons))
	collection := &PolyCollection{
		Polygons: polygons,
		PatchCollection: PatchCollection{
			Collection: Collection{Coords: Coords(CoordData), Label: opt.Label, Alpha: 1, z: zorder},
			FaceColors: colors,
			EdgeColor:  edgeColor,
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		polygons, zorder := a.projectFillBetween3DPolygons(x1[:n], y1[:n], z1[:n], x2[:n], y2[:n], z2[:n], opt.Mode)
		collection.Polygons = polygons
		collection.FaceColors = repeatColor(color, len(polygons))
		collection.z = zorder
	}, limitsChanged)
	return collection
}

// FillBetween is the Matplotlib-compatible 3D fill_between entry point.
func (a *Axes3D) FillBetween(x1, y1, z1, x2, y2, z2 []float64, opts ...FillBetween3DOptions) *PolyCollection {
	return a.FillBetween3D(x1, y1, z1, x2, y2, z2, opts...)
}

// Quiver plots a 3D vector field as projected shafts and arrowheads.
func (a *Axes3D) Quiver(x, y, z, u, v, w []float64, opts ...Quiver3DOptions) *LineCollection {
	if a == nil || a.Axes == nil {
		return nil
	}
	n := minLen(x, y, z, u, v, w)
	if n <= 0 {
		return nil
	}

	var opt Quiver3DOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	limitsChanged := a.observeQuiver3DData(x[:n], y[:n], z[:n], u[:n], v[:n], w[:n], opt)
	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	color.A *= alpha
	lineWidth := 1.5
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}

	segments, zorder := a.projectQuiver3DSegments(x[:n], y[:n], z[:n], u[:n], v[:n], w[:n], opt)
	if len(segments) == 0 {
		return nil
	}
	collection := &LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Label: opt.Label, Alpha: 1, z: zorder},
		Segments:   segments,
		Color:      color,
		LineWidth:  lineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		segments, zorder := a.projectQuiver3DSegments(x[:n], y[:n], z[:n], u[:n], v[:n], w[:n], opt)
		collection.Segments = segments
		collection.z = zorder
	}, limitsChanged)
	return collection
}

// Quiver3D is an explicit alias for Quiver.
func (a *Axes3D) Quiver3D(x, y, z, u, v, w []float64, opts ...Quiver3DOptions) *LineCollection {
	return a.Quiver(x, y, z, u, v, w, opts...)
}

// ErrorBar3D plots projected x/y/z error ranges.
func (a *Axes3D) ErrorBar3D(x, y, z, xErr, yErr, zErr []float64, opts ...ErrorBar3DOptions) *LineCollection {
	if a == nil || a.Axes == nil {
		return nil
	}
	n := minLen(x, y, z)
	if n <= 0 {
		return nil
	}
	var opt ErrorBar3DOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if !validErrorValues(xErr, n) || !validErrorValues(yErr, n) || !validErrorValues(zErr, n) ||
		!validErrorValues(opt.XErrLower, n) || !validErrorValues(opt.XErrUpper, n) ||
		!validErrorValues(opt.YErrLower, n) || !validErrorValues(opt.YErrUpper, n) ||
		!validErrorValues(opt.ZErrLower, n) || !validErrorValues(opt.ZErrUpper, n) {
		return nil
	}

	limitsChanged := a.observe3DErrorBarData(x[:n], y[:n], z[:n], xErr, yErr, zErr, opt)
	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	color.A *= alpha
	lineWidth := 1.0
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}

	segments, zorder := a.projectErrorBar3DSegments(x[:n], y[:n], z[:n], xErr, yErr, zErr, opt)
	if len(segments) == 0 {
		return nil
	}
	collection := &LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Label: opt.Label, Alpha: 1, z: zorder},
		Segments:   segments,
		Color:      color,
		LineWidth:  lineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		segments, zorder := a.projectErrorBar3DSegments(x[:n], y[:n], z[:n], xErr, yErr, zErr, opt)
		collection.Segments = segments
		collection.z = zorder
	}, limitsChanged)
	return collection
}

// ErrorBar is the Matplotlib-compatible 3D errorbar entry point.
func (a *Axes3D) ErrorBar(x, y, z, xErr, yErr, zErr []float64, opts ...ErrorBar3DOptions) *LineCollection {
	return a.ErrorBar3D(x, y, z, xErr, yErr, zErr, opts...)
}

// PlotSurfaceGrid creates a filled surface from a structured z grid.
func (a *Axes3D) PlotSurfaceGrid(x, y []float64, z [][]float64, opts ...PlotOptions) *PolyCollection {
	return a.Surface(x, y, z, opts...)
}

// Wireframe draws a structured wireframe as line segments.
func (a *Axes3D) Wireframe(x, y []float64, z [][]float64, opts ...PlotOptions) *LineCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	segments := a.projectWireframeSegments(x, y, z, opts...)
	if len(segments) == 0 {
		return nil
	}

	color := a.NextColor()
	lineWidth := 2.0
	alpha := 1.0
	label := ""
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Color != nil {
			color = *opt.Color
		}
		if opt.LineWidth != nil {
			lineWidth = *opt.LineWidth
		}
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		label = opt.Label
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  label,
			Alpha:  alpha,
			z:      a.grid3DCollectionZ(x, y, z),
		},
		Segments:  segments,
		Color:     color,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			collection.Segments = a.projectWireframeSegments(x, y, z, opts...)
			collection.z = a.grid3DCollectionZ(x, y, z)
		}
	}, limitsChanged)
	return collection
}

// Contour projects a structured z grid and emits a placeholder wireframe contour.
func (a *Axes3D) Contour(x, y []float64, z [][]float64, opts ...PlotOptions) *LineCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	opt := firstPlotOptions(opts)
	segments, segmentLevels, levels, values, zorder := a.projectedContourLineData(x, y, z, opt)
	if len(segments) == 0 {
		return nil
	}

	color := a.NextColor()
	lineWidth := 1.0
	alpha := 1.0
	label := ""
	colorOverride := false
	if opt.Color != nil {
		color = *opt.Color
		colorOverride = true
	}
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	label = opt.Label

	mapping := ScalarMapInfo{}
	colors := []render.Color(nil)
	collectionAlpha := alpha
	if !colorOverride {
		mapping = contourScalarMap(values, levels, opt)
		colors = make([]render.Color, len(segmentLevels))
		for i, level := range segmentLevels {
			colors[i] = mapping.Color(level, alpha)
		}
		collectionAlpha = 1
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords:   Coords(CoordData),
			Label:    label,
			Alpha:    collectionAlpha,
			z:        zorder,
			Colormap: mapping.Colormap,
			Norm:     mapping.Norm,
			VMin:     mapping.VMin,
			VMax:     mapping.VMax,
		},
		Segments:  segments,
		Color:     color,
		Colors:    colors,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			segments, segmentLevels, levels, values, zorder := a.projectedContourLineData(x, y, z, opt)
			collection.Segments = segments
			if !colorOverride {
				mapping := contourScalarMap(values, levels, opt)
				colors := make([]render.Color, len(segmentLevels))
				for i, level := range segmentLevels {
					colors[i] = mapping.Color(level, alpha)
				}
				collection.Colors = colors
				collection.Colormap = mapping.Colormap
				collection.Norm = mapping.Norm
				collection.VMin = mapping.VMin
				collection.VMax = mapping.VMax
			} else {
				collection.Colormap = ""
				collection.Norm = nil
				collection.VMin = 0
				collection.VMax = 0
				collection.Colors = nil
			}
			collection.z = zorder
		}
	}, limitsChanged)
	return collection
}

// Contourf projects a structured z grid and emits filled contour bands.
func (a *Axes3D) Contourf(x, y []float64, z [][]float64, opts ...PlotOptions) *PolyCollection {
	opt := firstPlotOptions(opts)
	colorOverride := opt.Color != nil
	alpha := 0.45
	label := ""
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	label = opt.Label
	limitsChanged := a.observe3DContourf(x, y, z, opt)

	paths, colors, zorder, mapping := a.projectedContourFillData(x, y, z, alpha, opt)
	if len(paths) == 0 {
		return nil
	}
	cmap := mapping.Colormap
	norm := mapping.Norm
	vMin := mapping.VMin
	vMax := mapping.VMax
	if colorOverride {
		cmap = ""
		norm = nil
		vMin = 0
		vMax = 0
	}

	collection := &PolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords:   Coords(CoordData),
				Label:    label,
				Alpha:    1,
				Colormap: cmap,
				Norm:     norm,
				VMin:     vMin,
				VMax:     vMax,
				z:        zorder,
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
			paths, colors, zorder, mapping := a.projectedContourFillData(x, y, z, alpha, opt)
			collection.Polygons = nil
			collection.Paths = paths
			collection.FaceColors = colors
			if colorOverride {
				collection.Colormap = ""
				collection.Norm = nil
				collection.VMin = 0
				collection.VMax = 0
			} else {
				collection.Colormap = mapping.Colormap
				collection.Norm = mapping.Norm
				collection.VMin = mapping.VMin
				collection.VMax = mapping.VMax
			}
			collection.z = zorder
		}
	}, limitsChanged)
	return collection
}

// Surface draws a structured surface as projected, z-sorted quadrilateral faces.
func (a *Axes3D) Surface(x, y []float64, z [][]float64, opts ...PlotOptions) *PolyCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	polygons, faceColors, zorder, mapping := a.projectSurfacePolygons(x, y, z, opts...)
	if len(polygons) == 0 {
		return nil
	}

	alpha := 0.85
	label := ""
	edgeWidth := 1.0
	edgeColor := render.Color{A: 0}
	antialias := render.AntialiasDefault
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		if opt.EdgeWidth != nil && *opt.EdgeWidth >= 0 {
			edgeWidth = *opt.EdgeWidth
		} else if opt.LineWidth != nil && *opt.LineWidth >= 0 {
			edgeWidth = *opt.LineWidth
		}
		if opt.EdgeColor != nil {
			edgeColor = *opt.EdgeColor
		}
		if opt.Antialiased != nil && !*opt.Antialiased {
			antialias = render.AntialiasOff
		}
		label = opt.Label
	}
	for i := range faceColors {
		faceColors[i].A *= alpha
	}
	edgeColor.A *= alpha
	edgeColors := surfaceEdgeColors(faceColors, firstPlotOptions(opts))

	collection := &PolyCollection{
		Polygons: polygons,
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords:    Coords(CoordData),
				Label:     label,
				Alpha:     1,
				Antialias: antialias,
				Colormap:  mapping.Colormap,
				Norm:      mapping.Norm,
				VMin:      mapping.VMin,
				VMax:      mapping.VMax,
				z:         zorder,
			},
			FaceColors: faceColors,
			EdgeColor:  edgeColor,
			EdgeColors: edgeColors,
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			polygons, faceColors, zorder, mapping := a.projectSurfacePolygons(x, y, z, opts...)
			for i := range faceColors {
				faceColors[i].A *= alpha
			}
			collection.Polygons = polygons
			collection.FaceColors = faceColors
			collection.EdgeColors = surfaceEdgeColors(faceColors, firstPlotOptions(opts))
			collection.Colormap = mapping.Colormap
			collection.Norm = mapping.Norm
			collection.VMin = mapping.VMin
			collection.VMax = mapping.VMax
			collection.z = zorder
		}
	}, limitsChanged)
	return collection
}

func (a *Axes3D) projectedContourSegments(x, y []float64, z [][]float64, levelCount int) [][]geom.Pt {
	segments, _, _, _, _ := a.projectedContourLineData(x, y, z, PlotOptions{LevelCount: levelCount})
	return segments
}

func (a *Axes3D) projectedContourLineData(x, y []float64, z [][]float64, opt PlotOptions) ([][]geom.Pt, []float64, []float64, []float64, float64) {
	if a == nil {
		return nil, nil, nil, nil, defaultPatchZ
	}
	zdir := normalized3DDir(opt.ZDir)
	rawLines, rawLevels, levels, values, ok := a.contourLines3D(x, y, z, opt, zdir)
	if !ok || len(rawLines) == 0 {
		return nil, nil, nil, nil, defaultPatchZ
	}
	segments := make([][]geom.Pt, 0, len(rawLines))
	segmentLevels := make([]float64, 0, len(rawLines))
	depth := math.Inf(1)
	for i, polyline := range rawLines {
		if len(polyline) < 2 {
			continue
		}
		level := rawLevels[i]
		planeLevel := contourPlaneLevel(level, opt.Offset)
		runs := [][]vec3{contourPolyline3D(polyline, planeLevel, zdir)}
		if opt.AxLimClip {
			runs = a.clip3DPolylineRuns(runs[0])
		}
		for _, run := range runs {
			if len(run) < 2 {
				continue
			}
			projected := make([]geom.Pt, len(run))
			for j, point3D := range run {
				var zDepth float64
				projected[j], zDepth = a.projectPointDepth(point3D[0], point3D[1], point3D[2])
				if zDepth < depth {
					depth = zDepth
				}
			}
			segments = append(segments, projected)
			segmentLevels = append(segmentLevels, level)
		}
	}
	return segments, segmentLevels, levels, values, computed3DCollectionZ(depth)
}

type surfaceFace struct {
	polygon []geom.Pt
	value   float64
	color   render.Color
	normal  vec3
	depth   float64
}

func contourPlaneLevel(level float64, offset *float64) float64 {
	if offset != nil && isFinite(*offset) {
		return *offset
	}
	return level
}

func contourPolyline3D(polyline []geom.Pt, planeLevel float64, zdir string) []vec3 {
	points := make([]vec3, len(polyline))
	for i, pt := range polyline {
		points[i] = juggle3DPointSigned(pt.X, pt.Y, planeLevel, "-"+zdir)
	}
	return points
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

func (a *Axes3D) contourLines3D(x, y []float64, z [][]float64, opt PlotOptions, zdir string) ([][]geom.Pt, []float64, []float64, []float64, bool) {
	rows, cols, ok := validate3DGridContourInput(x, y, z)
	if !ok {
		return nil, nil, nil, nil, false
	}
	values := flattenGridValues(z)
	levels := contourLevels(values, opt.Levels, opt.LevelCount, false)
	if len(levels) == 0 {
		return nil, nil, nil, nil, false
	}
	if zdir == "z" {
		lines, lineLevels := contourGridPolylines(x[:cols], y[:rows], z, levels)
		return lines, lineLevels, levels, values, true
	}

	tri, rotatedValues, ok := rotatedContourTriangulation(x[:cols], y[:rows], z, zdir)
	if !ok {
		return nil, nil, nil, nil, false
	}
	lines, lineLevels := contourPolylines(tri, rotatedValues, levels)
	return lines, lineLevels, levels, rotatedValues, true
}

func (a *Axes3D) projectedContourFillData(x, y []float64, z [][]float64, alpha float64, opt PlotOptions) ([]geom.Path, []render.Color, float64, ScalarMapInfo) {
	rows, cols, ok := validate3DGridContourInput(x, y, z)
	if !ok {
		return nil, nil, defaultPatchZ, ScalarMapInfo{}
	}
	values := flattenGridValues(z)
	levels := contourLevels(values, opt.Levels, opt.LevelCount, true)
	if len(levels) < 2 {
		return nil, nil, defaultPatchZ, ScalarMapInfo{}
	}
	zdir := normalized3DDir(opt.ZDir)
	mapping := contourScalarMap(values, levels, opt)
	collectionDepth := math.Inf(1)
	paths := make([]geom.Path, 0, len(levels)-1)
	colors := make([]render.Color, 0, len(levels)-1)

	var tri Triangulation
	var rotatedValues []float64
	if zdir != "z" {
		var ok bool
		tri, rotatedValues, ok = rotatedContourTriangulation(x[:cols], y[:rows], z, zdir)
		if !ok {
			return nil, nil, defaultPatchZ, ScalarMapInfo{}
		}
	}

	for levelIdx := 0; levelIdx+1 < len(levels); levelIdx++ {
		low := levels[levelIdx]
		high := levels[levelIdx+1]
		bandLevel := 0.5 * (low + high)
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
		return nil, nil, defaultPatchZ, ScalarMapInfo{}
	}
	return paths, colors, computed3DCollectionZ(collectionDepth), mapping
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
		if tri.masked(triIdx) {
			continue
		}
		polygon := triangleBandPolygon(
			[3]geom.Pt{tri.point(triangle[0]), tri.point(triangle[1]), tri.point(triangle[2])},
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

func (a *Axes3D) projectSurfacePolygons(x, y []float64, z [][]float64, opts ...PlotOptions) ([][]geom.Pt, []render.Color, float64, ScalarMapInfo) {
	if a == nil || len(z) == 0 {
		return nil, nil, 0, ScalarMapInfo{}
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return nil, nil, 0, ScalarMapInfo{}
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return nil, nil, 0, ScalarMapInfo{}
		}
	}

	opt := firstPlotOptions(opts)
	faces := make([]surfaceFace, 0, (rows-1)*(cols-1))
	values := make([]float64, 0, (rows-1)*(cols-1))
	collectionDepth := math.Inf(1)
	rowIndices, colIndices := surfaceSampleIndices(rows, cols, opt)
	defaultColor := a.NextColor()
	useMapping := opt.Colormap != nil && *opt.Colormap != ""
	useExplicitFaceColors := len(opt.FaceColors) > 0
	shade := surfaceShadeEnabled(opt, useMapping)
	for rowIdx := 0; rowIdx+1 < len(rowIndices); rowIdx++ {
		row0 := rowIndices[rowIdx]
		row1 := rowIndices[rowIdx+1]
		for colIdx := 0; colIdx+1 < len(colIndices); colIdx++ {
			col0 := colIndices[colIdx]
			col1 := colIndices[colIdx+1]
			corners := [4]vec3{
				{x[col0], y[row0], z[row0][col0]},
				{x[col1], y[row0], z[row0][col1]},
				{x[col1], y[row1], z[row1][col1]},
				{x[col0], y[row1], z[row1][col0]},
			}
			normal := corners[0].sub(corners[1]).cross(corners[1].sub(corners[2]))
			polygon := make([]geom.Pt, 0, 2*(row1-row0)+2*(col1-col0))
			rawPolygon := make([]vec3, 0, 2*(row1-row0)+2*(col1-col0))
			value := 0.0
			depth := 0.0
			valid := true
			count := 0
			surfacePatchPerimeter(row0, row1, col0, col1, func(row, col int) {
				if !valid {
					return
				}
				zVal := z[row][col]
				if !isFinite3D(x[col], y[row], zVal) {
					valid = false
					return
				}
				rawPolygon = append(rawPolygon, vec3{x[col], y[row], zVal})
				pt, zDepth := a.projectPointDepth(x[col], y[row], zVal)
				polygon = append(polygon, pt)
				value += zVal
				depth += zDepth
				if zDepth < collectionDepth {
					collectionDepth = zDepth
				}
				count++
			})
			if !valid {
				continue
			}
			if opt.AxLimClip && !a.polygonWithin3DViewLimits(rawPolygon) {
				continue
			}
			value /= float64(count)
			depth /= float64(count)
			baseColor := defaultColor
			switch {
			case useExplicitFaceColors:
				baseColor = faceColorAtIndex(opt.FaceColors, len(faces))
			case opt.Color != nil:
				baseColor = *opt.Color
			}
			if shade && !useMapping {
				baseColor = shade3DFaceColor(baseColor, normal)
			}
			faces = append(faces, surfaceFace{
				polygon: polygon,
				value:   value,
				color:   baseColor,
				normal:  normal,
				depth:   depth,
			})
			values = append(values, value)
		}
	}
	if len(faces) == 0 {
		return nil, nil, 0, ScalarMapInfo{}
	}

	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].depth > faces[j].depth
	})

	mapping := ScalarMapInfo{}
	if useMapping {
		mapping = resolvePlotScalarMap(values, opt)
	}
	polygons := make([][]geom.Pt, len(faces))
	colors := make([]render.Color, len(faces))
	for i, face := range faces {
		polygons[i] = face.polygon
		if useMapping {
			colors[i] = mapping.Color(face.value, 1)
		} else {
			colors[i] = face.color
		}
	}
	return polygons, colors, computed3DCollectionZ(collectionDepth), mapping
}

func surfaceGridSampleIndices(length, count int) []int {
	if length <= 0 {
		return nil
	}
	if count <= 0 {
		count = default3DSurfaceCount
	}
	stride := int(math.Ceil(float64(length) / float64(count)))
	if stride < 1 {
		stride = 1
	}

	indices := make([]int, 0, (length+stride-1)/stride+1)
	if (length-1)%stride == 0 {
		for i := 0; i < length; i += stride {
			indices = append(indices, i)
		}
		return indices
	}

	for i := 0; i < length-1; i += stride {
		indices = append(indices, i)
	}
	return append(indices, length-1)
}

func surfaceSampleIndices(rows, cols int, opt PlotOptions) ([]int, []int) {
	hasStride := opt.RStride != nil || opt.CStride != nil
	if hasStride {
		rstride, cstride := 10, 10
		if opt.RStride != nil {
			rstride = *opt.RStride
		}
		if opt.CStride != nil {
			cstride = *opt.CStride
		}
		return steppedSampleIndices(rows, rstride), steppedSampleIndices(cols, cstride)
	}
	rcount, ccount := default3DSurfaceCount, default3DSurfaceCount
	if opt.RCount != nil {
		rcount = *opt.RCount
	}
	if opt.CCount != nil {
		ccount = *opt.CCount
	}
	return surfaceGridSampleIndices(rows, rcount), surfaceGridSampleIndices(cols, ccount)
}

func surfacePatchPerimeter(row0, row1, col0, col1 int, emit func(row, col int)) {
	for col := col0; col < col1; col++ {
		emit(row0, col)
	}
	for row := row0; row < row1; row++ {
		emit(row, col1)
	}
	for col := col1; col > col0; col-- {
		emit(row1, col)
	}
	for row := row1; row > row0; row-- {
		emit(row, col0)
	}
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
	rawPolygons, colors := contourGridBandPolygons(x[:cols], y[:rows], z, levels, opt, mapping, alpha)
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
	polygons, polygonColors := contourBandPolygons(tri, values, levels, opt, mapping, 1.0)
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

// Trisurf projects a triangulated unstructured surface mesh as filled polygons.
func (a *Axes3D) Trisurf(tri Triangulation, z []float64, opts ...PlotOptions) *PolyCollection {
	if a == nil || len(tri.X) == 0 {
		return nil
	}
	if err := tri.Validate(); err != nil || len(z) != len(tri.X) {
		return nil
	}
	limitsChanged := a.observe3DTriangulation(tri, z)

	color := a.NextColor()
	lineWidth := 1.0
	alpha := 1.0
	label := ""
	edgeColor := render.Color{A: 0}
	antialias := render.AntialiasDefault
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Color != nil {
			color = *opt.Color
		}
		if opt.EdgeWidth != nil {
			lineWidth = *opt.EdgeWidth
		} else if opt.LineWidth != nil {
			lineWidth = *opt.LineWidth
		}
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		if opt.EdgeColor != nil {
			edgeColor = *opt.EdgeColor
			edgeColor.A *= alpha
		}
		if opt.Antialiased != nil && !*opt.Antialiased {
			antialias = render.AntialiasOff
		}
		label = opt.Label
	}

	faceColor := color
	faceColor.A *= alpha
	faces, faceColors, faceZ, mapping := a.projectTriangulationFaces(tri, z, faceColor, firstPlotOptions(opts))
	if len(faces) == 0 {
		return nil
	}
	collection := &PolyCollection{
		Polygons: faces,
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords:    Coords(CoordData),
				Label:     label,
				Alpha:     1,
				Antialias: antialias,
				Colormap:  mapping.Colormap,
				Norm:      mapping.Norm,
				VMin:      mapping.VMin,
				VMax:      mapping.VMax,
				z:         faceZ,
			},
			FaceColors: faceColors,
			EdgeColor:  edgeColor,
			EdgeWidth:  lineWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			faces, faceColors, faceZ, mapping := a.projectTriangulationFaces(tri, z, faceColor, firstPlotOptions(opts))
			collection.Polygons = faces
			collection.FaceColors = faceColors
			collection.Colormap = mapping.Colormap
			collection.Norm = mapping.Norm
			collection.VMin = mapping.VMin
			collection.VMax = mapping.VMax
			collection.z = faceZ
		}
	}, limitsChanged)
	return collection
}

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

func (a *Axes3D) frameSegments(mins, maxs vec3) [][]geom.Pt {
	return a.frameSegmentsProjected(mins, maxs, mins, maxs, mins, maxs)
}

func (a *Axes3D) frameSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs vec3) [][]geom.Pt {
	return a.frameGridSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs)
}

func (a *Axes3D) activePanePolygons(mins, maxs vec3) [][]geom.Pt {
	return a.activePanePolygonsProjected(mins, maxs, mins, maxs)
}

func (a *Axes3D) activePanePolygonsProjected(mins, maxs, projMins, projMaxs vec3) [][]geom.Pt {
	planes := [6][4][3]int{
		{{0, 0, 0}, {0, 1, 0}, {0, 1, 1}, {0, 0, 1}},
		{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}},
		{{0, 1, 0}, {1, 1, 0}, {1, 1, 1}, {0, 1, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}},
		{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}},
	}
	project := func(corner [3]int) geom.Pt {
		x := mins[0]
		if corner[0] == 1 {
			x = maxs[0]
		}
		y := mins[1]
		if corner[1] == 1 {
			y = maxs[1]
		}
		z := mins[2]
		if corner[2] == 1 {
			z = maxs[2]
		}
		return a.project3DPointWithState(x, y, z, projMins, projMaxs)
	}

	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	panes := make([][]geom.Pt, 0, 3)
	for axis := range 3 {
		planeIndex := 2 * axis
		if highs[axis] {
			planeIndex++
		}
		plane := planes[planeIndex]
		polygon := make([]geom.Pt, len(plane))
		for i, corner := range plane {
			polygon[i] = project(corner)
		}
		panes = append(panes, polygon)
	}
	return panes
}

func (a *Axes3D) frameGridSegments(mins, maxs vec3) [][]geom.Pt {
	return a.frameGridSegmentsProjected(mins, maxs, mins, maxs, mins, maxs)
}

func (a *Axes3D) axisLineSegmentsProjected(mins, maxs, projMins, projMaxs vec3) [][]geom.Pt {
	project := func(p vec3) geom.Pt {
		return a.project3DPointWithState(p[0], p[1], p[2], projMins, projMaxs)
	}
	pairs := a.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)
	segments := make([][]geom.Pt, 0, len(pairs))
	for _, pair := range pairs {
		segments = append(segments, []geom.Pt{project(pair[0]), project(pair[1])})
	}
	return segments
}

func (a *Axes3D) axisTickSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs vec3) [][]geom.Pt {
	project := func(p vec3) geom.Pt {
		return a.project3DPointWithState(p[0], p[1], p[2], projMins, projMaxs)
	}
	pairs := a.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)
	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	tickDirs := [3]int{1, 0, 0}
	segments := make([][]geom.Pt, 0, 24)
	for axis, pair := range pairs {
		tickDir := tickDirs[axis]
		tickDelta := (tickMaxs[tickDir] - tickMins[tickDir]) / 12
		if !highs[tickDir] {
			tickDelta = -tickDelta
		}
		outward := pair[0][tickDir] + 0.1*tickDelta
		inward := pair[0][tickDir] - 0.2*tickDelta
		for _, tick := range frameAxisTicks(tickMins[axis], tickMaxs[axis]) {
			p0 := pair[0]
			p1 := pair[0]
			p0[axis] = tick
			p1[axis] = tick
			p0[tickDir] = outward
			p1[tickDir] = inward
			segments = append(segments, []geom.Pt{project(p0), project(p1)})
		}
	}
	return segments
}

func (a *Axes3D) axisLineEdgePointPairs(mins, maxs, projMins, projMaxs vec3) [][2]vec3 {
	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	minmax := vec3{}
	maxmin := vec3{}
	for i := range 3 {
		if highs[i] {
			minmax[i] = maxs[i]
			maxmin[i] = mins[i]
		} else {
			minmax[i] = mins[i]
			maxmin[i] = maxs[i]
		}
	}

	juggled := [3][3]int{
		{1, 0, 2},
		{0, 1, 2},
		{0, 2, 1},
	}
	pairs := make([][2]vec3, 0, 3)
	for axis := range 3 {
		p0 := minmax
		p0[juggled[axis][0]] = maxmin[juggled[axis][0]]
		p1 := p0
		p1[juggled[axis][1]] = maxmin[juggled[axis][1]]
		pairs = append(pairs, [2]vec3{p0, p1})
	}
	return pairs
}

func (a *Axes3D) frameGridSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs vec3) [][]geom.Pt {
	project := func(p vec3) geom.Pt {
		return a.project3DPointWithState(p[0], p[1], p[2], projMins, projMaxs)
	}
	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	minmax := vec3{}
	maxmin := vec3{}
	for i := range 3 {
		if highs[i] {
			minmax[i] = maxs[i]
			maxmin[i] = mins[i]
		} else {
			minmax[i] = mins[i]
			maxmin[i] = maxs[i]
		}
	}

	segments := make([][]geom.Pt, 0, 18)
	limits := [][2]float64{
		{tickMins[0], tickMaxs[0]},
		{tickMins[1], tickMaxs[1]},
		{tickMins[2], tickMaxs[2]},
	}
	for index, lim := range limits {
		for _, tick := range frameAxisTicks(lim[0], lim[1]) {
			p0 := minmax
			p1 := minmax
			p2 := minmax
			p0[index] = tick
			p1[index] = tick
			p2[index] = tick
			first := (index + 1) % 3
			second := (index + 2) % 3
			p0[first] = maxmin[first]
			p2[second] = maxmin[second]
			segments = append(segments, []geom.Pt{project(p0), project(p1), project(p2)})
		}
	}
	return segments
}

func (a *Axes3D) activePaneHighs(mins, maxs vec3) [3]bool {
	return a.activePaneHighsProjected(mins, maxs, mins, maxs)
}

func (a *Axes3D) activePaneHighsProjected(mins, maxs, projMins, projMaxs vec3) [3]bool {
	planes := [6][4]int{
		{0, 3, 7, 4},
		{1, 2, 6, 5},
		{0, 1, 5, 4},
		{3, 2, 6, 7},
		{0, 1, 2, 3},
		{4, 5, 6, 7},
	}
	corners := [8]vec3{
		{mins[0], mins[1], mins[2]},
		{maxs[0], mins[1], mins[2]},
		{maxs[0], maxs[1], mins[2]},
		{mins[0], maxs[1], mins[2]},
		{mins[0], mins[1], maxs[2]},
		{maxs[0], mins[1], maxs[2]},
		{maxs[0], maxs[1], maxs[2]},
		{mins[0], maxs[1], maxs[2]},
	}
	depths := [8]float64{}
	for i, corner := range corners {
		depths[i] = a.projectPointDepthWithProjectionLimits(corner[0], corner[1], corner[2], projMins, projMaxs)
	}

	means0 := [3]float64{}
	means1 := [3]float64{}
	highs := [3]bool{}
	equals := [3]bool{}
	equalCount := 0
	for axis := range 3 {
		means0[axis] = meanPlaneDepth(depths, planes[2*axis])
		means1[axis] = meanPlaneDepth(depths, planes[2*axis+1])
		highs[axis] = means0[axis] < means1[axis]
		if math.Abs(means0[axis]-means1[axis]) <= math.Nextafter(1, 2)-1 {
			equals[axis] = true
			equalCount++
		}
	}
	if equalCount == 2 {
		vertical := -1
		for i := range equals {
			if !equals[i] {
				vertical = i
				break
			}
		}
		switch vertical {
		case 2:
			highs[0], highs[1] = true, true
		case 1:
			highs[0], highs[2] = true, false
		case 0:
			highs[1], highs[2] = false, false
		}
	}
	return highs
}

func meanPlaneDepth(depths [8]float64, plane [4]int) float64 {
	return (depths[plane[0]] + depths[plane[1]] + depths[plane[2]] + depths[plane[3]]) / 4
}

func computed3DCollectionZ(projectedDepth float64) float64 {
	if math.IsNaN(projectedDepth) || math.IsInf(projectedDepth, 0) {
		return defaultPatchZ
	}
	return default3DComputedZ - projectedDepth
}

func (a *Axes3D) points3DCollectionZ(x, y, z []float64) float64 {
	n := minLen(x, y, z)
	depth := math.Inf(1)
	for i := 0; i < n; i++ {
		if !isFinite3D(x[i], y[i], z[i]) {
			continue
		}
		_, zDepth := a.projectPointDepth(x[i], y[i], z[i])
		if zDepth < depth {
			depth = zDepth
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) grid3DCollectionZ(x, y []float64, z [][]float64) float64 {
	if a == nil || len(z) == 0 {
		return defaultPatchZ
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return defaultPatchZ
	}
	depth := math.Inf(1)
	for row := 0; row < rows; row++ {
		if len(z[row]) != cols {
			return computed3DCollectionZ(depth)
		}
		for col := 0; col < cols; col++ {
			if !isFinite3D(x[col], y[row], z[row][col]) {
				continue
			}
			_, zDepth := a.projectPointDepth(x[col], y[row], z[row][col])
			if zDepth < depth {
				depth = zDepth
			}
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) triangulation3DCollectionZ(tri Triangulation, z []float64) float64 {
	if a == nil {
		return defaultPatchZ
	}
	depth := math.Inf(1)
	for triIdx, t := range tri.Triangles {
		if tri.masked(triIdx) {
			continue
		}
		for _, idx := range t {
			if idx < 0 || idx >= len(tri.X) || idx >= len(tri.Y) || idx >= len(z) || !isFinite3D(tri.X[idx], tri.Y[idx], z[idx]) {
				continue
			}
			_, zDepth := a.projectPointDepth(tri.X[idx], tri.Y[idx], z[idx])
			if zDepth < depth {
				depth = zDepth
			}
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) bar3DCollectionZ(x, y, z, dx, dy, dz []float64) float64 {
	n := minLen(x, y, z, dx, dy, dz)
	depth := math.Inf(1)
	for i := 0; i < n; i++ {
		x0, x1 := x[i], x[i]+dx[i]
		y0, y1 := y[i], y[i]+dy[i]
		z0, z1 := z[i], z[i]+dz[i]
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
		for _, corner := range corners {
			if !isFinite3D(corner[0], corner[1], corner[2]) {
				continue
			}
			_, zDepth := a.projectPointDepth(corner[0], corner[1], corner[2])
			if zDepth < depth {
				depth = zDepth
			}
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) projectPointDepthWithLimits(x, y, z float64, mins, maxs vec3) (geom.Pt, float64) {
	if a == nil {
		return geom.Pt{}, 0
	}
	state := a.projectionState()
	if a.distance <= 0 {
		return project3DPointWithLimits(x, y, z, a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs, state), z
	}
	m := default3DProjectionMatrix(a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs, state)
	tx, ty, tz := transform3DPoint(m, x, y, z)
	return geom.Pt{X: tx, Y: ty}, tz
}

func (a *Axes3D) projectPointDepthWithProjectionLimits(x, y, z float64, mins, maxs vec3) float64 {
	state := a.projectionState()
	if a.distance <= 0 {
		return z
	}
	m := default3DProjectionMatrix(a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs, state)
	_, _, tz := transform3DPoint(m, x, y, z)
	return tz
}

func (a *Axes3D) draw3DTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, mins, maxs, tickMins, tickMaxs vec3) {
	fontSize := ctx.RC.TickLabelSize("x")
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	centers, deltas := axes3DLabelCentersDeltas(ctx, tickMins, tickMaxs)
	labelDeltas := vec3{}
	for i := range 3 {
		labelDeltas[i] = (defaultTickPadPt + 8) * deltas[i]
	}
	axisLines := a.axisLineEdgePointPairs(mins, maxs, tickMins, tickMaxs)
	tickDirs := [3]int{1, 0, 0}

	if a.showXLabels {
		xTicks := frameAxisTicks(tickMins[0], tickMaxs[0])
		for i, tick := range xTicks {
			pos := axisLines[0][0]
			pos[0] = tick
			pos[tickDirs[0]] = axisLines[0][0][tickDirs[0]]
			anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 0), tickMins, tickMaxs)
			draw3DTextAtAnchorAligned(textRen, r, ctx, format3DTick(tick, i, xTicks), anchor, fontSize, textColor, textLayoutVAlignTop)
		}
	}
	if a.showYLabels {
		yTicks := frameAxisTicks(tickMins[1], tickMaxs[1])
		for i, tick := range yTicks {
			pos := axisLines[1][0]
			pos[1] = tick
			pos[tickDirs[1]] = axisLines[1][0][tickDirs[1]]
			anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 1), tickMins, tickMaxs)
			draw3DTextAtAnchorAligned(textRen, r, ctx, format3DTick(tick, i, yTicks), anchor, fontSize, textColor, textLayoutVAlignTop)
		}
	}
	zTicks := frameAxisTicks(tickMins[2], tickMaxs[2])
	if a.showZLabels {
		for i, tick := range zTicks {
			pos := axisLines[2][0]
			pos[2] = tick
			pos[tickDirs[2]] = axisLines[2][0][tickDirs[2]]
			anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 2), tickMins, tickMaxs)
			draw3DTextAtAnchorAligned(textRen, r, ctx, format3DTick(tick, i, zTicks), anchor, fontSize, textColor, textLayoutVAlignTop)
		}
	}
}

func (a *Axes3D) draw3DAxisLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, mins, maxs vec3) {
	fontSize := axisLabelFontSize(ctx)
	textColor := ctx.RC.DefaultAxesLabelColor()
	projMins, projMaxs := a.projectionLimits()
	centers, deltas := axes3DLabelCentersDeltas(ctx, projMins, projMaxs)
	labelDeltas := vec3{}
	for i := range 3 {
		labelDeltas[i] = (4 + 21) * deltas[i]
	}
	axisLines := a.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)
	if a.XLabel != "" {
		pos := midpoint3D(axisLines[0][0], axisLines[0][1])
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 0), projMins, projMaxs)
		draw3DTextAtAnchor(textRen, r, ctx, a.XLabel, anchor, fontSize, textColor)
	}
	if a.YLabel != "" {
		pos := midpoint3D(axisLines[1][0], axisLines[1][1])
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 1), projMins, projMaxs)
		draw3DTextAtAnchor(textRen, r, ctx, a.YLabel, anchor, fontSize, textColor)
	}
	if a.zLabel != "" {
		pos := midpoint3D(axisLines[2][0], axisLines[2][1])
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 2), projMins, projMaxs)
		draw3DTextAtAnchor(textRen, r, ctx, a.zLabel, anchor, fontSize, textColor)
	}
}

func axes3DLabelCentersDeltas(ctx *DrawContext, mins, maxs vec3) (vec3, vec3) {
	centers := vec3{}
	deltas := vec3{}
	dpi := 100.0
	clipWidth := 1.0
	clipHeight := 1.0
	if ctx != nil {
		if ctx.RC.DPI > 0 {
			dpi = ctx.RC.DPI
		}
		if ctx.Clip.W() > 0 {
			clipWidth = ctx.Clip.W()
		}
		if ctx.Clip.H() > 0 {
			clipHeight = ctx.Clip.H()
		}
	}
	deltasPerPoint := 48 / (72 * (clipWidth + clipHeight) / dpi)
	// matplotlib uses scale = 1/12 * 24/25 = 0.08 (not 1/12) to match automargin behavior
	const scale = 0.08
	for i := range 3 {
		centers[i] = (mins[i] + maxs[i]) / 2
		deltas[i] = (maxs[i] - mins[i]) * scale * deltasPerPoint
	}
	return centers, deltas
}

func move3DLabelFromCenter(pos, centers, deltas vec3, axis int) vec3 {
	for i := range 3 {
		if i == axis {
			continue
		}
		if pos[i] < centers[i] {
			pos[i] -= deltas[i]
		} else {
			pos[i] += deltas[i]
		}
	}
	return pos
}

func midpoint3D(a, b vec3) vec3 {
	return vec3{(a[0] + b[0]) / 2, (a[1] + b[1]) / 2, (a[2] + b[2]) / 2}
}

func (a *Axes3D) project3DLabelAnchor(ctx *DrawContext, pos, projMins, projMaxs vec3) geom.Pt {
	projected := a.project3DPointWithState(pos[0], pos[1], pos[2], projMins, projMaxs)
	return ctx.TransformFor(Coords(CoordData)).Apply(projected)
}

func draw3DTextAtAnchor(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, label string, anchor geom.Pt, fontSize float64, textColor render.Color) {
	draw3DTextAtAnchorAligned(textRen, r, ctx, label, anchor, fontSize, textColor, textLayoutVAlignCenter)
}

func draw3DTextAtAnchorAligned(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, label string, anchor geom.Pt, fontSize float64, textColor render.Color, vAlign textLayoutVerticalAlign) {
	if label == "" {
		return
	}
	layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	origin := alignedSingleLineOrigin(anchor, layout, TextAlignCenter, vAlign)
	drawDisplayText(textRen, label, origin, fontSize, textColor, ctx.RC.FontKey, ctx.RC.UseTeX)
}

func frameAxisTicks(minVal, maxVal float64) []float64 {
	ticks := AutoLocator{}.Ticks(minVal, maxVal, 9)
	if len(ticks) == 0 {
		return nil
	}
	eps := 1e-12 * math.Max(1, math.Abs(maxVal-minVal))
	filtered := ticks[:0]
	for _, tick := range ticks {
		if tick >= minVal-eps && tick <= maxVal+eps {
			filtered = append(filtered, tick)
		}
	}
	return filtered
}

func format3DTick(v float64, i int, ticks []float64) string {
	step := 0.0
	if len(ticks) > 1 {
		if i+1 < len(ticks) {
			step = math.Abs(ticks[i+1] - ticks[i])
		} else {
			step = math.Abs(ticks[i] - ticks[i-1])
		}
	}
	return strings.ReplaceAll(formatScalarTickLabel(ScalarFormatter{Prec: 3}, v, step), "-", "\u2212")
}

// Bar3DPlaneOptions configures Axes3D.Bar, the projected 2D bar variant.
type Bar3DPlaneOptions struct {
	Color     *render.Color
	Width     *float64
	EdgeColor *render.Color
	EdgeWidth *float64
	Alpha     *float64
	Baseline  *float64
	Baselines []float64
	Z         *float64
	Zs        []float64
	ZDir      string
	Label     string
}

// Bar projects 2D bars into the plane orthogonal to ZDir, matching mplot3d's
// 2D bar compatibility path rather than the cuboid Bar3D helper.
func (a *Axes3D) Bar(x, heights []float64, opts ...Bar3DPlaneOptions) *PolyCollection {
	if a == nil || a.Axes == nil {
		return nil
	}
	n := minLen(x, heights)
	if n <= 0 {
		return nil
	}

	var opt Bar3DPlaneOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	width := 0.8
	if opt.Width != nil {
		width = *opt.Width
	}
	baseline := 0.0
	if opt.Baseline != nil {
		baseline = *opt.Baseline
	}
	z := 0.0
	if opt.Z != nil {
		z = *opt.Z
	}
	zdir := normalized3DDir(opt.ZDir)

	limitsChanged := a.observe3DPlaneBars(x[:n], heights[:n], width, baseline, opt.Baselines, z, opt.Zs, zdir)
	color := a.NextPatchColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	color.A *= alpha
	edgeColor := render.Color{}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
		edgeColor.A *= alpha
	}
	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	polygons, zorder := a.project3DPlaneBars(x[:n], heights[:n], width, baseline, opt.Baselines, z, opt.Zs, zdir)
	if len(polygons) == 0 {
		return nil
	}
	collection := &PolyCollection{
		Polygons: polygons,
		PatchCollection: PatchCollection{
			Collection: Collection{Coords: Coords(CoordData), Label: opt.Label, Alpha: 1, z: zorder},
			FaceColors: repeatColor(color, len(polygons)),
			EdgeColor:  edgeColor,
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		polygons, zorder := a.project3DPlaneBars(x[:n], heights[:n], width, baseline, opt.Baselines, z, opt.Zs, zdir)
		collection.Polygons = polygons
		collection.FaceColors = repeatColor(color, len(polygons))
		collection.z = zorder
	}, limitsChanged)
	return collection
}

// Bar3DOptions configures projected wireframe bars.
type Bar3DOptions struct {
	Color     *render.Color
	LineWidth *float64
	Alpha     *float64
	Label     string
}

// VoxelOptions configures boolean-grid voxel rendering.
type VoxelOptions struct {
	FaceColor  *render.Color
	FaceColors map[[3]int]render.Color
	EdgeColor  *render.Color
	EdgeColors map[[3]int]render.Color
	Alpha      *float64
	Shade      *bool
	Label      string
}

// Bar3D draws a simple projected wireframe column for each x/y/z sample.
func (a *Axes3D) Bar3D(x, y, z, dx, dy, dz []float64, opts ...Bar3DOptions) *LineCollection {
	n := minLen(x, y, z, dx, dy, dz)
	if n <= 0 || a == nil {
		return nil
	}
	limitsChanged := a.observe3DBarData(x, y, z, dx, dy, dz)

	color := a.NextColor()
	lineWidth := 1.0
	alpha := 1.0
	edgeAlpha := 0.0
	label := ""
	if len(opts) > 0 {
		o := opts[0]
		if o.Color != nil {
			color = *o.Color
		}
		if o.LineWidth != nil {
			lineWidth = *o.LineWidth
			edgeAlpha = alpha
		}
		if o.Alpha != nil && *o.Alpha >= 0 && *o.Alpha <= 1 {
			alpha = *o.Alpha
			if o.LineWidth != nil {
				edgeAlpha = alpha
			}
		}
		label = o.Label
	}

	faceColor := color
	if len(opts) > 0 && opts[0].Alpha != nil {
		faceColor.A *= alpha
	}
	faces, faceColors := a.projectBar3DShadedFaces(x, y, z, dx, dy, dz, faceColor)
	barZ := a.bar3DCollectionZ(x, y, z, dx, dy, dz)
	if len(faces) > 0 {
		faceCollection := &PolyCollection{
			Polygons: faces,
			PatchCollection: PatchCollection{
				Collection: Collection{Coords: Coords(CoordData), Alpha: 1, z: barZ},
				FaceColors: faceColors,
				EdgeColor:  render.Color{A: 0},
				LineJoin:   render.JoinMiter,
				LineCap:    render.CapButt,
			},
		}
		a.Add(faceCollection)
		a.add3DReprojector(func() {
			if faceCollection != nil {
				faces, faceColors := a.projectBar3DShadedFaces(x, y, z, dx, dy, dz, faceColor)
				faceCollection.Polygons = faces
				faceCollection.FaceColors = faceColors
				faceCollection.z = a.bar3DCollectionZ(x, y, z, dx, dy, dz)
			}
		}, limitsChanged)
	}

	segments := a.projectBar3DSegments(x, y, z, dx, dy, dz)

	collection := &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  label,
			Alpha:  edgeAlpha,
			z:      barZ,
		},
		Segments:  segments,
		Color:     color,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			collection.Segments = a.projectBar3DSegments(x, y, z, dx, dy, dz)
			collection.z = a.bar3DCollectionZ(x, y, z, dx, dy, dz)
		}
	}, limitsChanged)
	return collection
}

// Voxels renders a boolean occupancy grid as per-voxel face collections with
// internal-face culling, matching Matplotlib's voxel artist model.
func (a *Axes3D) Voxels(filled [][][]bool, opts ...VoxelOptions) map[[3]int]*PolyCollection {
	if a == nil || a.Axes == nil {
		return nil
	}
	_, _, _, ok := voxelGridShape(filled)
	if !ok {
		return nil
	}

	var opt VoxelOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	alpha := 1.0
	if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
		alpha = *opt.Alpha
	}
	// Resolve the default face color once here so the reprojector closure
	// does not re-call NextPatchColor() and advance the color cycle.
	if opt.FaceColor == nil {
		faceColor := a.NextPatchColor()
		opt.FaceColor = &faceColor
	}
	// Default edge width matches matplotlib's patch.linewidth (1.0) when an
	// edge color is configured; without a positive width the edges are invisible.
	edgeWidth := 0.0
	if opt.EdgeColor != nil && opt.EdgeColor.A > 0 {
		edgeWidth = 1.0
	}
	limitsChanged := a.observe3DVoxels(filled)
	projected := a.projectVoxelCollections(filled, opt, alpha)
	if len(projected) == 0 {
		return map[[3]int]*PolyCollection{}
	}

	collections := make(map[[3]int]*PolyCollection, len(projected))
	for coord, voxel := range projected {
		collection := &PolyCollection{
			Polygons: voxel.polygons,
			PatchCollection: PatchCollection{
				Collection: Collection{Coords: Coords(CoordData), Label: opt.Label, Alpha: 1, z: voxel.zorder},
				FaceColors: voxel.faceColors,
				EdgeColor:  voxel.edgeColor,
				EdgeWidth:  edgeWidth,
				LineJoin:   render.JoinMiter,
				LineCap:    render.CapButt,
			},
		}
		a.Add(collection)
		collections[coord] = collection
	}
	a.add3DReprojector(func() {
		refreshed := a.projectVoxelCollections(filled, opt, alpha)
		for coord, collection := range collections {
			voxel, ok := refreshed[coord]
			if !ok {
				continue
			}
			collection.Polygons = voxel.polygons
			collection.FaceColors = voxel.faceColors
			collection.EdgeColor = voxel.edgeColor
			collection.EdgeWidth = edgeWidth
			collection.z = voxel.zorder
		}
	}, limitsChanged)
	return collections
}

// Voxel projects unstructured rectangular prisms as wireframe voxels.
func (a *Axes3D) Voxel(x, y, z, dx, dy, dz []float64, opts ...PlotOptions) *LineCollection {
	if len(opts) == 0 {
		return a.Bar3D(x, y, z, dx, dy, dz)
	}

	o := opts[0]
	voxelOpts := make([]Bar3DOptions, 1)
	if o.Color != nil {
		voxelOpts[0].Color = o.Color
	}
	if o.LineWidth != nil {
		voxelOpts[0].LineWidth = o.LineWidth
	}
	if o.Alpha != nil {
		voxelOpts[0].Alpha = o.Alpha
	}
	voxelOpts[0].Label = o.Label
	return a.Bar3D(x, y, z, dx, dy, dz, voxelOpts...)
}

type projectedVoxelCollection struct {
	polygons   [][]geom.Pt
	faceColors []render.Color
	edgeColor  render.Color
	zorder     float64
}

func voxelGridShape(filled [][][]bool) (int, int, int, bool) {
	if len(filled) == 0 || len(filled[0]) == 0 || len(filled[0][0]) == 0 {
		return 0, 0, 0, false
	}
	nx, ny, nz := len(filled), len(filled[0]), len(filled[0][0])
	for i := 0; i < nx; i++ {
		if len(filled[i]) != ny {
			return 0, 0, 0, false
		}
		for j := 0; j < ny; j++ {
			if len(filled[i][j]) != nz {
				return 0, 0, 0, false
			}
		}
	}
	return nx, ny, nz, true
}

func (a *Axes3D) observe3DVoxels(filled [][][]bool) bool {
	nx, ny, nz, ok := voxelGridShape(filled)
	if !ok {
		return false
	}
	changed := false
	for i := 0; i < nx; i++ {
		for j := 0; j < ny; j++ {
			for k := 0; k < nz; k++ {
				if !filled[i][j][k] {
					continue
				}
				if a.observe3DPoint(float64(i), float64(j), float64(k)) {
					changed = true
				}
				if a.observe3DPoint(float64(i+1), float64(j+1), float64(k+1)) {
					changed = true
				}
			}
		}
	}
	return changed
}

func (a *Axes3D) projectVoxelCollections(filled [][][]bool, opt VoxelOptions, alpha float64) map[[3]int]projectedVoxelCollection {
	nx, ny, nz, ok := voxelGridShape(filled)
	if !ok {
		return nil
	}
	defaultFaceColor := a.NextPatchColor()
	if opt.FaceColor != nil {
		defaultFaceColor = *opt.FaceColor
	}
	defaultFaceColor.A *= alpha
	defaultEdgeColor := render.Color{}
	if opt.EdgeColor != nil {
		defaultEdgeColor = *opt.EdgeColor
		defaultEdgeColor.A *= alpha
	}
	shade := true
	if opt.Shade != nil {
		shade = *opt.Shade
	}

	type voxelFace struct {
		polygon []geom.Pt
		color   render.Color
		depth   float64
	}
	projected := map[[3]int]projectedVoxelCollection{}
	for i := 0; i < nx; i++ {
		for j := 0; j < ny; j++ {
			for k := 0; k < nz; k++ {
				if !filled[i][j][k] {
					continue
				}
				coord := [3]int{i, j, k}
				faceColor := defaultFaceColor
				if color, ok := opt.FaceColors[coord]; ok {
					faceColor = color
					faceColor.A *= alpha
				}
				edgeColor := defaultEdgeColor
				if color, ok := opt.EdgeColors[coord]; ok {
					edgeColor = color
					edgeColor.A *= alpha
				}

				faces := make([]voxelFace, 0, 6)
				for _, raw := range voxelVisibleFaces(filled, i, j, k) {
					polygon := make([]geom.Pt, len(raw.polygon))
					depth := 0.0
					for idx, point := range raw.polygon {
						projectedPt, zDepth := a.projectPointDepth(point[0], point[1], point[2])
						polygon[idx] = projectedPt
						depth += zDepth
					}
					color := faceColor
					if shade {
						color = shade3DFaceColor(color, raw.normal)
					}
					faces = append(faces, voxelFace{
						polygon: polygon,
						color:   color,
						depth:   depth / float64(len(raw.polygon)),
					})
				}
				if len(faces) == 0 {
					continue
				}
				sort.SliceStable(faces, func(aIdx, bIdx int) bool {
					return faces[aIdx].depth > faces[bIdx].depth
				})
				polygons := make([][]geom.Pt, len(faces))
				colors := make([]render.Color, len(faces))
				minDepth := math.Inf(1)
				for idx, face := range faces {
					polygons[idx] = face.polygon
					colors[idx] = face.color
					if face.depth < minDepth {
						minDepth = face.depth
					}
				}
				projected[coord] = projectedVoxelCollection{
					polygons:   polygons,
					faceColors: colors,
					edgeColor:  edgeColor,
					zorder:     computed3DCollectionZ(minDepth),
				}
			}
		}
	}
	return projected
}

type voxelRawFace struct {
	polygon []vec3
	normal  vec3
}

func voxelVisibleFaces(filled [][][]bool, i, j, k int) []voxelRawFace {
	nx, ny, nz, _ := voxelGridShape(filled)
	visible := make([]voxelRawFace, 0, 6)
	neighbors := []struct {
		delta  [3]int
		normal vec3
		face   []vec3
	}{
		{
			delta:  [3]int{-1, 0, 0},
			normal: vec3{-1, 0, 0},
			face: []vec3{
				{float64(i), float64(j), float64(k)},
				{float64(i), float64(j + 1), float64(k)},
				{float64(i), float64(j + 1), float64(k + 1)},
				{float64(i), float64(j), float64(k + 1)},
			},
		},
		{
			delta:  [3]int{1, 0, 0},
			normal: vec3{1, 0, 0},
			face: []vec3{
				{float64(i + 1), float64(j), float64(k)},
				{float64(i + 1), float64(j), float64(k + 1)},
				{float64(i + 1), float64(j + 1), float64(k + 1)},
				{float64(i + 1), float64(j + 1), float64(k)},
			},
		},
		{
			delta:  [3]int{0, -1, 0},
			normal: vec3{0, -1, 0},
			face: []vec3{
				{float64(i), float64(j), float64(k)},
				{float64(i), float64(j), float64(k + 1)},
				{float64(i + 1), float64(j), float64(k + 1)},
				{float64(i + 1), float64(j), float64(k)},
			},
		},
		{
			delta:  [3]int{0, 1, 0},
			normal: vec3{0, 1, 0},
			face: []vec3{
				{float64(i), float64(j + 1), float64(k)},
				{float64(i + 1), float64(j + 1), float64(k)},
				{float64(i + 1), float64(j + 1), float64(k + 1)},
				{float64(i), float64(j + 1), float64(k + 1)},
			},
		},
		{
			delta:  [3]int{0, 0, -1},
			normal: vec3{0, 0, -1},
			face: []vec3{
				{float64(i), float64(j), float64(k)},
				{float64(i + 1), float64(j), float64(k)},
				{float64(i + 1), float64(j + 1), float64(k)},
				{float64(i), float64(j + 1), float64(k)},
			},
		},
		{
			delta:  [3]int{0, 0, 1},
			normal: vec3{0, 0, 1},
			face: []vec3{
				{float64(i), float64(j), float64(k + 1)},
				{float64(i), float64(j + 1), float64(k + 1)},
				{float64(i + 1), float64(j + 1), float64(k + 1)},
				{float64(i + 1), float64(j), float64(k + 1)},
			},
		},
	}
	for _, neighbor := range neighbors {
		ni := i + neighbor.delta[0]
		nj := j + neighbor.delta[1]
		nk := k + neighbor.delta[2]
		if ni >= 0 && ni < nx && nj >= 0 && nj < ny && nk >= 0 && nk < nz && filled[ni][nj][nk] {
			continue
		}
		visible = append(visible, voxelRawFace{polygon: neighbor.face, normal: neighbor.normal})
	}
	return visible
}

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
	if maxVal < minVal {
		minVal, maxVal = maxVal, minVal
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
}

// SetShowYTickLabels controls whether y-axis tick labels are drawn on the 3D frame.
func (a *Axes3D) SetShowYTickLabels(show bool) {
	if a == nil {
		return
	}
	a.showYLabels = show
}

// SetShowZTickLabels controls whether z-axis tick labels are drawn on the 3D frame.
func (a *Axes3D) SetShowZTickLabels(show bool) {
	if a == nil {
		return
	}
	a.showZLabels = show
	a.reproject3DArtists()
}

func (a *Axes3D) GetZLabel() string {
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
		if !a.hasData {
			continue
		}
		if mins[i] == maxs[i] {
			mins[i] -= 0.5
			maxs[i] += 0.5
		}
		margin := (maxs[i] - mins[i]) * a.default3DDataMargin(i)
		mins[i] -= margin
		maxs[i] += margin
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

func axes3DFrameLimits(mins, maxs vec3) (vec3, vec3) {
	for i := range 3 {
		delta := (maxs[i] - mins[i]) / 12
		mins[i] -= 0.25 * delta
		maxs[i] += 0.25 * delta
	}
	return mins, maxs
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

func (a *Axes3D) observe3DContourf(x, y []float64, z [][]float64, opt PlotOptions) bool {
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

func (a *Axes3D) observe3DTriangulation(tri Triangulation, z []float64) bool {
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
	a.zsorted = false
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

func surfaceShadeEnabled(opt PlotOptions, useMapping bool) bool {
	if opt.Shade != nil {
		return *opt.Shade
	}
	return !useMapping
}

func surfaceEdgeColors(faceColors []render.Color, opt PlotOptions) []render.Color {
	if opt.EdgeColor != nil || len(opt.FaceColors) == 0 {
		return nil
	}
	edges := make([]render.Color, len(faceColors))
	copy(edges, faceColors)
	return edges
}

func trisurfShadeEnabled(opt PlotOptions, useMapping bool) bool {
	if opt.Shade != nil {
		return *opt.Shade
	}
	return !useMapping
}

func resolvePlotScalarMap(values []float64, opt PlotOptions) ScalarMapInfo {
	cmap := ""
	if opt.Colormap != nil {
		cmap = *opt.Colormap
	}
	mapping, err := ResolveScalarMapValues(values, ScalarMapConfig{
		Colormap: cmap,
		Norm:     opt.Norm,
		VMin:     opt.VMin,
		VMax:     opt.VMax,
	})
	if err != nil {
		return ScalarMapInfo{Colormap: resolvedColormapName(cmap)}.Resolved()
	}
	return mapping
}

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

func (a *Axes3D) projectStem3DGeometry(x, y, z []float64, bottom float64, orientation string) ([][]geom.Pt, []geom.Pt, []geom.Pt, float64) {
	n := minLen(x, y, z)
	segments := make([][]geom.Pt, 0, n)
	baseline := make([]geom.Pt, 0, n)
	offsets := make([]geom.Pt, 0, n)
	depth := math.Inf(1)
	for i := 0; i < n; i++ {
		start, end := stem3DLineEndpoints(x[i], y[i], z[i], bottom, orientation)
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
	return a.project3DLineSegments(quiver3DRawSegments(x, y, z, u, v, w, opt))
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
}

func (a *Axes3D) projectFillBetween3DPolygons(x1, y1, z1, x2, y2, z2 []float64, mode FillBetween3DMode) ([][]geom.Pt, float64) {
	raw := fillBetween3DRawPolygons(x1, y1, z1, x2, y2, z2, mode)
	return a.projectSorted3DPolygons(raw)
}

func fillBetween3DRawPolygons(x1, y1, z1, x2, y2, z2 []float64, mode FillBetween3DMode) [][]vec3 {
	n := minLen(x1, y1, z1, x2, y2, z2)
	if n < 2 {
		return nil
	}
	if mode == "" || mode == FillBetween3DModeAuto {
		if fillBetween3DCoplanar(x1[:n], y1[:n], z1[:n], x2[:n], y2[:n], z2[:n]) {
			mode = FillBetween3DModePolygon
		} else {
			mode = FillBetween3DModeQuad
		}
	}
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
	return a.project3DLineSegments(errorBar3DRawSegments(x, y, z, xErr, yErr, zErr, opt))
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

func (a *Axes3D) projectSorted3DPolygons(raw [][]vec3) ([][]geom.Pt, float64) {
	projected := make([]projected3DPolygon, 0, len(raw))
	collectionDepth := math.Inf(1)
	for _, polygon3D := range raw {
		if len(polygon3D) < 3 {
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
			projected = append(projected, projected3DPolygon{polygon: polygon, depth: depth / float64(len(polygon3D))})
		}
	}
	sort.SliceStable(projected, func(i, j int) bool {
		return projected[i].depth > projected[j].depth
	})
	polygons := make([][]geom.Pt, len(projected))
	for i, item := range projected {
		polygons[i] = item.polygon
	}
	return polygons, computed3DCollectionZ(collectionDepth)
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

func (a *Axes3D) project3DLineSegments(raw [][]vec3) ([][]geom.Pt, float64) {
	segments := make([][]geom.Pt, 0, len(raw))
	collectionDepth := math.Inf(1)
	for _, segment3D := range raw {
		if len(segment3D) < 2 {
			continue
		}
		segment := make([]geom.Pt, 0, len(segment3D))
		valid := true
		for _, p := range segment3D {
			if !isFinite3D(p[0], p[1], p[2]) {
				valid = false
				break
			}
			pt, zDepth := a.projectPointDepth(p[0], p[1], p[2])
			segment = append(segment, pt)
			if zDepth < collectionDepth {
				collectionDepth = zDepth
			}
		}
		if valid {
			segments = append(segments, segment)
		}
	}
	return segments, computed3DCollectionZ(collectionDepth)
}

func (a *Axes3D) projectedData(x, y, z []float64) []geom.Pt {
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

	pts := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		pts[i] = a.ProjectPoint(x[i], y[i], z[i])
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
	segments := make([][]geom.Pt, 0, len(rowIndices)+len(colIndices))
	for _, row := range rowIndices {
		line := make([]geom.Pt, 0, cols)
		for col := 0; col < cols; col++ {
			line = append(line, a.ProjectPoint(x[col], y[row], z[row][col]))
		}
		if len(line) > 1 {
			segments = append(segments, line)
		}
	}
	for _, col := range colIndices {
		line := make([]geom.Pt, 0, rows)
		for row := 0; row < rows; row++ {
			line = append(line, a.ProjectPoint(x[col], y[row], z[row][col]))
		}
		if len(line) > 1 {
			segments = append(segments, line)
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
	view := viewTransformation(u, v, w, eye)
	proj := perspectiveTransformation(-distance, distance, default3DFocalLength)
	return proj.mul(view.mul(world))
}

func default3DBoxAspect() vec3 {
	aspect := vec3{4, 4, 3}
	scale := default3DBoxAspectScale / aspect.norm()
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

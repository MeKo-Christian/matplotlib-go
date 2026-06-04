package core

import (
	"fmt"
	"math"
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
	default3DScatterSize     = 20.0
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
	opt := firstPlotOptions(opts)
	projected := a.projectedData(x, y, z, opt.AxLimClip)
	if len(projected) == 0 {
		return nil
	}

	x2, y2 := make([]float64, len(projected)), make([]float64, len(projected))
	for i, p := range projected {
		x2[i] = p.X
		y2[i] = p.Y
	}

	if len(opts) > 0 {
		line := a.Plot(x2, y2, opt)
		a.add3DReprojector(func() {
			reprojectLine3D(line, a.projectedData(x, y, z, opt.AxLimClip))
		}, limitsChanged)
		return line
	}
	line := a.Plot(x2, y2)
	a.add3DReprojector(func() {
		reprojectLine3D(line, a.projectedData(x, y, z, opt.AxLimClip))
	}, limitsChanged)
	return line
}

// Scatter3D projects x/y/z values and draws markers through projected points.
func (a *Axes3D) Scatter3D(x, y, z []float64, opts ...ScatterOptions) *Scatter2D {
	limitsChanged := a.observe3DData(x, y, z)
	if a.ensure3DZMargin(0.05) {
		limitsChanged = true
	}
	opt := ScatterOptions{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.Size == nil {
		size := default3DScatterSize
		opt.Size = &size
	}
	projected := a.projectedData(x, y, z, opt.AxLimClip)
	if len(projected) == 0 {
		return nil
	}

	x2, y2 := make([]float64, len(projected)), make([]float64, len(projected))
	for i, p := range projected {
		x2[i] = p.X
		y2[i] = p.Y
	}

	if len(opts) > 0 {
		scatter := a.Scatter(x2, y2, opt)
		reprojectScatter3D(scatter, a.projectedScatterData(x, y, z, opt.AxLimClip))
		scatter.z = a.points3DCollectionZ(x, y, z)
		a.add3DReprojector(func() {
			reprojectScatter3D(scatter, a.projectedScatterData(x, y, z, opt.AxLimClip))
			scatter.z = a.points3DCollectionZ(x, y, z)
		}, limitsChanged)
		return scatter
	}
	scatter := a.Scatter(x2, y2, opt)
	reprojectScatter3D(scatter, a.projectedScatterData(x, y, z, opt.AxLimClip))
	scatter.z = a.points3DCollectionZ(x, y, z)
	a.add3DReprojector(func() {
		reprojectScatter3D(scatter, a.projectedScatterData(x, y, z, opt.AxLimClip))
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
	AxLimClip       bool
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
	AxLimClip bool

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
	baselineColor := a.colorCycleAt(3)
	if opt.BaselineColor != nil {
		baselineColor = *opt.BaselineColor
		baselineColor.A *= alpha
	} else {
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

	segments, baseline, offsets, zorder := a.projectStem3DGeometry(x[:n], y[:n], z[:n], bottom, orientation, opt.AxLimClip)
	stems := &LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Label: opt.Label, Alpha: 1, z: zorder},
		Segments:   segments,
		Color:      color,
		LineWidth:  lineWidth,
		LineJoin:   render.JoinRound,
		LineCap:    render.CapButt,
	}
	markers := &PathCollection{
		Collection:    Collection{Coords: Coords(CoordData), Label: opt.Label, Alpha: 1, z: zorder + 0.05},
		Path:          markerPath,
		Offsets:       offsets,
		Size:          pointsToPixels(a.resolvedRC(), markerSize),
		PathInDisplay: true,
		FaceColor:     color,
		EdgeColor:     markerEdgeColor,
		EdgeWidth:     markerEdgeWidth,
		LineOnly:      markerLineOnly(NewMarkerStyle(markerType)),
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
		segments, baseline, offsets, zorder := a.projectStem3DGeometry(x[:n], y[:n], z[:n], bottom, orientation, opt.AxLimClip)
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

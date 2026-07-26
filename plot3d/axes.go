package plot3d

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/optional"
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
	default3DWorldViewMargin = 1.0 / 48.0
	default3DScatterSize     = 20.0
)

// Axes3D represents an Axes with basic 3D projection helpers.
//
// The underlying artist model is still 2D (`*Axes`) with pre-projected
// 3D coordinates converted into data-space 2D points before drawing.
type Axes3D struct {
	*core.Axes
	azimuthDeg   float64
	elevationDeg float64
	rollDeg      float64
	verticalAxis int
	zLabel       string
	showXLabels  bool
	showYLabels  bool
	showZLabels  bool
	distance     float64
	focalLength  float64
	boxAspect    vec3
	hasData      bool
	dataMin      vec3
	dataMax      vec3
	viewMin      vec3
	viewMax      vec3
	viewSet      [3]bool
	zMargin      float64
	reprojectors []func()

	// One-shot guards for the misleadingly-named placeholder helpers so the
	// nudge toward the real APIs is emitted once per axes, not per call.
	plotSurfaceWarned bool
	voxelWarned       bool
}

type axes3DFrame struct {
	axes *Axes3D
}

type projected3DState struct {
	rollDeg      float64
	verticalAxis int
	boxAspect    vec3
	focalLength  float64
}

func (a *Axes3D) projectionState() projected3DState {
	if a == nil {
		return projected3DState{
			rollDeg:      default3DRollDeg,
			verticalAxis: default3DVerticalAxis,
			boxAspect:    default3DBoxAspect(),
			focalLength:  default3DFocalLength,
		}
	}
	return projected3DState{
		rollDeg:      a.rollDeg,
		verticalAxis: a.verticalAxis,
		boxAspect:    a.boxAspect,
		focalLength:  a.focalLength,
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
		if s.focalLength == 0 {
			s.focalLength = default3DFocalLength
		}
		return s
	}
	return projected3DState{
		rollDeg:      default3DRollDeg,
		verticalAxis: default3DVerticalAxis,
		boxAspect:    default3DBoxAspect(),
		focalLength:  default3DFocalLength,
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

func (f *axes3DFrame) Draw(r render.Renderer, ctx *core.DrawContext) {
	if f == nil || f.axes == nil || r == nil || ctx == nil {
		return
	}
	mins, maxs := f.axes.projectionLimits()
	frameMins, frameMaxs := mins, maxs
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
	(&core.PolyCollection{
		Polygons: panes,
		PatchCollection: core.PatchCollection{
			Collection: core.Collection{Coords: core.Coords(core.CoordData), Alpha: 1},
			FaceColors: axes3DPaneFaceColors(),
			EdgeColor:  render.Color{A: 0},
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}).Draw(r, ctx)

	nbins := axes3DTickBins(ctx)
	segments := f.axes.frameSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs, nbins)
	(&core.LineCollection{
		Collection: core.Collection{Coords: core.Coords(core.CoordData), Alpha: 1},
		Segments:   segments,
		Color:      gridColor,
		LineWidth:  gridLineWidth,
		Dashes:     scaleGridDashes(ctx.RC.GridDashes, gridLineWidth),
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}).Draw(r, ctx)

	axisLines := f.axes.axisLineSegmentsProjected(frameMins, frameMaxs, mins, maxs)
	(&core.LineCollection{
		Collection: core.Collection{Coords: core.Coords(core.CoordData), Alpha: 1},
		Segments:   axisLines,
		Color:      axisColor,
		LineWidth:  axisLineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}).Draw(r, ctx)

	tickSegments := f.axes.axisTickSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs, nbins)
	(&core.LineCollection{
		Collection: core.Collection{Coords: core.Coords(core.CoordData), Alpha: 1},
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

func (f *axes3DFrame) Bounds(*core.DrawContext) geom.Rect { return geom.Rect{} }

func axes3DPaneFaceColors() []render.Color {
	return []render.Color{
		{R: 0.95, G: 0.95, B: 0.95, A: 0.5},
		{R: 0.90, G: 0.90, B: 0.90, A: 0.5},
		{R: 0.925, G: 0.925, B: 0.925, A: 0.5},
	}
}

func (f *axes3DFrame) DrawOverlay(r render.Renderer, ctx *core.DrawContext) {
}

// NewAxes wraps an existing axes and configures 3D default view settings.
func NewAxes(ax *core.Axes) *Axes3D {
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
		focalLength:  default3DFocalLength,
		boxAspect:    default3DBoxAspect(),
	}
	ax.Add(&axes3DFrame{axes: axes})
	return axes
}

// Plot3D projects x/y/z values and draws a line through projected points.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Plot3D(x, y, z []float64, opt core.PlotOptions) *core.Line2D {
	limitsChanged := a.observe3DData(x, y, z)
	projected := a.projectedData(x, y, z, opt.AxLimClip)
	if len(projected) == 0 {
		return nil
	}

	x2, y2 := make([]float64, len(projected)), make([]float64, len(projected))
	for i, p := range projected {
		x2[i] = p.X
		y2[i] = p.Y
	}

	line, err := a.Plot(x2, y2, opt)
	if err != nil {
		return nil
	}
	a.add3DReprojector(func() {
		reprojectLine3D(line, a.projectedData(x, y, z, opt.AxLimClip))
	}, limitsChanged)
	return line
}

// Scatter3D projects x/y/z values and draws markers through projected points.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Scatter3D(x, y, z []float64, opt core.ScatterOptions) *core.Scatter2D {
	limitsChanged := a.observe3DData(x, y, z)
	if a.ensure3DZMargin(0.05) {
		limitsChanged = true
	}
	if !opt.Size.IsSet() {
		size := default3DScatterSize
		opt.Size = optional.Of(size)
	}
	projected := a.projectedScatterData(x, y, z, opt.AxLimClip)
	if len(projected) == 0 {
		return nil
	}

	x2, y2 := make([]float64, len(projected)), make([]float64, len(projected))
	for i, p := range projected {
		x2[i] = p.point.X
		y2[i] = p.point.Y
	}
	projectedOpt := scatterOptionsForProjected(opt, projected)

	scatter, err := a.Scatter(x2, y2, projectedOpt)
	if err != nil {
		return nil
	}
	reprojectScatter3D(scatter, a.projectedScatterData(x, y, z, opt.AxLimClip), opt)
	scatter.SetZ(a.points3DCollectionZ(x, y, z))
	a.add3DReprojector(func() {
		reprojectScatter3D(scatter, a.projectedScatterData(x, y, z, opt.AxLimClip), opt)
		scatter.SetZ(a.points3DCollectionZ(x, y, z))
	}, limitsChanged)
	return scatter
}

// PlotSurface draws the input sequence as a connected 3D line strip. Despite
// the name it does NOT render a Matplotlib-style surface: it takes flat 1D
// slices and delegates to [Axes3D.Plot3D]. For an actual surface from a 2D
// height grid use [Axes3D.Surface] (or [Axes3D.Trisurf] for an unstructured
// triangulation). The name is retained for backwards compatibility.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) PlotSurface(x, y, z []float64, opt core.PlotOptions) *core.Line2D {
	if a != nil && !a.plotSurfaceWarned {
		a.plotSurfaceWarned = true
		diag.Warnf("Axes3D.PlotSurface draws a line strip, not a surface; use Axes3D.Surface(x, y, z) for a real surface")
	}
	return a.Plot3D(x, y, z, opt)
}

// Text3D projects a point and renders arbitrary text at the projected location.
//
//nolint:gocritic // TextOptions is forwarded unchanged to the axes method.
func (a *Axes3D) Text3D(x, y, z float64, text string, opt core.TextOptions) *core.Text {
	if a == nil || a.Axes == nil {
		return nil
	}
	p := a.ProjectPoint(x, y, z)
	txt := a.Text(p.X, p.Y, text, opt)
	a.add3DReprojector(func() {
		if txt != nil {
			txt.Position = a.ProjectPoint(x, y, z)
		}
	}, false)
	return txt
}

// Stem3DOptions configures Axes3D.Stem3D.
type Stem3DOptions struct {
	Color           optional.Value[render.Color]
	LineWidth       optional.Value[float64]
	Marker          optional.Value[core.MarkerType]
	MarkerPath      optional.Value[geom.Path]
	MarkerSize      optional.Value[float64]
	Bottom          optional.Value[float64]
	Orientation     string
	BaselineColor   optional.Value[render.Color]
	BaselineWidth   optional.Value[float64]
	MarkerEdgeColor optional.Value[render.Color]
	MarkerEdgeWidth optional.Value[float64]
	Label           string
	Alpha           optional.Value[float64]
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
	Color     optional.Value[render.Color]
	EdgeColor optional.Value[render.Color]
	EdgeWidth optional.Value[float64]
	Alpha     optional.Value[float64]
	Label     string
	Mode      FillBetween3DMode
	// Shade mirrors matplotlib's fill_between shade parameter: nil defaults
	// to true in 'quad' mode and false in 'polygon' mode.
	Shade     optional.Value[bool]
	AxLimClip bool
}

// Quiver3DOptions configures Axes3D.Quiver.
type Quiver3DOptions struct {
	Color            optional.Value[render.Color]
	LineWidth        optional.Value[float64]
	Alpha            optional.Value[float64]
	Label            string
	Length           optional.Value[float64]
	ArrowLengthRatio optional.Value[float64]
	Pivot            string
	Normalize        bool
	AxLimClip        bool
}

// ErrorBar3DOptions configures Axes3D.ErrorBar3D.
type ErrorBar3DOptions struct {
	Color     optional.Value[render.Color]
	LineWidth optional.Value[float64]
	CapSize   optional.Value[float64]
	Alpha     optional.Value[float64]
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
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Stem3D(x, y, z []float64, opt Stem3DOptions) *core.StemContainer {
	if a == nil || a.Axes == nil {
		return nil
	}
	n := minLen(x, y, z)
	if n <= 0 {
		return nil
	}

	orientation := normalized3DDir(opt.Orientation)
	bottom := 0.0
	if v, ok := opt.Bottom.Get(); ok {
		bottom = v
	}

	limitsChanged := a.observe3DStemData(x[:n], y[:n], z[:n], bottom, orientation)
	color := a.NextColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}
	alpha := 1.0
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && opt.Alpha.OrZero() <= 1 {
		alpha = opt.Alpha.OrZero()
	}
	color = color.WithAlphaMultiplier(alpha)

	lineWidth := 1.5 // points; converted at the collection/line Paint sink
	if v, ok := opt.LineWidth.Get(); ok {
		lineWidth = v
	}
	markerType := core.MarkerCircle
	if v, ok := opt.Marker.Get(); ok {
		markerType = v
	}
	markerSize := 6.0
	if v, ok := opt.MarkerSize.Get(); ok {
		markerSize = v
	}
	markerEdgeColor := color
	if opt.MarkerEdgeColor.IsSet() {
		markerEdgeColor = opt.MarkerEdgeColor.OrZero()
		markerEdgeColor = markerEdgeColor.WithAlphaMultiplier(alpha)
	}
	markerEdgeWidth := 1.0 // points; converted at the collection Paint sink
	if v, ok := opt.MarkerEdgeWidth.Get(); ok {
		markerEdgeWidth = v
	}
	baselineColor := colorCycleAt(a.Axes, 3)
	if opt.BaselineColor.IsSet() {
		baselineColor = opt.BaselineColor.OrZero()
		baselineColor = baselineColor.WithAlphaMultiplier(alpha)
	} else {
		baselineColor = baselineColor.WithAlphaMultiplier(alpha)
	}
	baselineWidth := lineWidth
	if v, ok := opt.BaselineWidth.Get(); ok {
		baselineWidth = v
	}
	markerPath := geom.Path{}
	if v, ok := opt.MarkerPath.Get(); ok {
		markerPath = v
	}
	if len(markerPath.C) == 0 {
		scatter := core.Scatter2D{Marker: markerType}
		markerPath = scatter.PrototypePath()
	}

	segments, baseline, offsets, zorder := a.projectStem3DGeometry(x[:n], y[:n], z[:n], bottom, orientation, opt.AxLimClip)
	stems := &core.LineCollection{
		Collection: core.Collection{Coords: core.Coords(core.CoordData), Label: opt.Label, Alpha: 1},
		Segments:   segments,
		Color:      color,
		LineWidth:  lineWidth,
		LineJoin:   render.JoinRound,
		LineCap:    render.CapButt,
	}
	markers := &core.PathCollection{
		Collection:    core.Collection{Coords: core.Coords(core.CoordData), Label: opt.Label, Alpha: 1},
		Path:          markerPath,
		Offsets:       offsets,
		Size:          pointsToPixels(a.ResolvedRC(), markerSize),
		PathInDisplay: true,
		FaceColor:     color,
		EdgeColor:     markerEdgeColor,
		EdgeWidth:     markerEdgeWidth,
		LineOnly:      markerLineOnly(core.NewMarkerStyle(markerType)),
	}
	baselineArtist := &core.Line2D{
		XY:    baseline,
		W:     baselineWidth,
		Col:   baselineColor,
		Label: "",
	}
	stems.SetZ(zorder)
	markers.SetZ(zorder + 0.05)
	baselineArtist.SetZ(zorder - 0.05)

	a.AddCollection(stems)
	a.AddCollection(markers)
	a.Add(baselineArtist)
	a.add3DReprojector(func() {
		segments, baseline, offsets, zorder := a.projectStem3DGeometry(x[:n], y[:n], z[:n], bottom, orientation, opt.AxLimClip)
		stems.Segments = segments
		stems.SetZ(zorder)
		markers.Offsets = offsets
		markers.SetZ(zorder + 0.05)
		baselineArtist.XY = baseline
		baselineArtist.SetZ(zorder - 0.05)
	}, limitsChanged)

	return &core.StemContainer{
		MarkerCollection: markers,
		StemLines:        stems,
		Baseline:         baselineArtist,
		Label:            opt.Label,
	}
}

// Stem is the Matplotlib-compatible 3D stem entry point.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Stem(x, y, z []float64, opt Stem3DOptions) *core.StemContainer {
	return a.Stem3D(x, y, z, opt)
}

// FillBetween3D fills bands between two 3D curves.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) FillBetween3D(x1, y1, z1, x2, y2, z2 []float64, opt FillBetween3DOptions) *core.PolyCollection {
	if a == nil || a.Axes == nil {
		return nil
	}
	n := minLen(x1, y1, z1, x2, y2, z2)
	if n < 2 {
		return nil
	}

	limitsChanged := a.observe3DData(x1[:n], y1[:n], z1[:n])
	if a.observe3DData(x2[:n], y2[:n], z2[:n]) {
		limitsChanged = true
	}

	color := a.NextPatchColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}
	alpha := 1.0
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && opt.Alpha.OrZero() <= 1 {
		alpha = opt.Alpha.OrZero()
	}
	color = color.WithAlphaMultiplier(alpha)
	edgeColor := render.Color{}
	if opt.EdgeColor.IsSet() {
		edgeColor = opt.EdgeColor.OrZero()
		edgeColor = edgeColor.WithAlphaMultiplier(alpha)
	}
	edgeWidth := 0.0
	if v, ok := opt.EdgeWidth.Get(); ok {
		edgeWidth = v
	}

	// matplotlib fill_between: shade defaults to true in 'quad' mode and
	// false in 'polygon' mode; shaded quads get per-face lightsource-scaled
	// copies of the base color (Poly3DCollection(shade=True) ->
	// art3d._generate_normals + _shade_colors).
	mode := resolveFillBetween3DMode(x1[:n], y1[:n], z1[:n], x2[:n], y2[:n], z2[:n], opt.Mode)
	shade := mode == FillBetween3DModeQuad
	if v, ok := opt.Shade.Get(); ok {
		shade = v
	}
	raw := fillBetween3DRawPolygons(x1[:n], y1[:n], z1[:n], x2[:n], y2[:n], z2[:n], mode)
	faceColors := make([]render.Color, len(raw))
	for i, polygon := range raw {
		if shade {
			faceColors[i] = shade3DFaceColor(color, polygon3DNormal(polygon))
		} else {
			faceColors[i] = color
		}
	}

	project := func() ([][]geom.Pt, []render.Color, float64) {
		return a.projectSorted3DPolygonsWithColors(raw, faceColors, opt.AxLimClip)
	}
	polygons, colors, zorder := project()
	if len(polygons) == 0 {
		return nil
	}
	collection := &core.PolyCollection{
		Polygons: polygons,
		PatchCollection: core.PatchCollection{
			Collection: core.Collection{Coords: core.Coords(core.CoordData), Label: opt.Label, Alpha: 1},
			FaceColors: colors,
			EdgeColor:  edgeColor,
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	collection.SetZ(zorder)
	a.Add(collection)
	a.add3DReprojector(func() {
		polygons, colors, zorder := project()
		collection.Polygons = polygons
		collection.FaceColors = colors
		collection.SetZ(zorder)
	}, limitsChanged)
	return collection
}

// FillBetween is the Matplotlib-compatible 3D fill_between entry point.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) FillBetween(x1, y1, z1, x2, y2, z2 []float64, opt FillBetween3DOptions) *core.PolyCollection {
	return a.FillBetween3D(x1, y1, z1, x2, y2, z2, opt)
}

// Quiver plots a 3D vector field as projected shafts and arrowheads.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Quiver(x, y, z, u, v, w []float64, opt Quiver3DOptions) *core.LineCollection {
	if a == nil || a.Axes == nil {
		return nil
	}
	n := minLen(x, y, z, u, v, w)
	if n <= 0 {
		return nil
	}

	limitsChanged := a.observeQuiver3DData(x[:n], y[:n], z[:n], u[:n], v[:n], w[:n], opt)
	color := a.NextColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}
	alpha := 1.0
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && opt.Alpha.OrZero() <= 1 {
		alpha = opt.Alpha.OrZero()
	}
	color = color.WithAlphaMultiplier(alpha)
	lineWidth := 1.5 // points; converted at the collection Paint sink
	if v, ok := opt.LineWidth.Get(); ok {
		lineWidth = v
	}

	segments, zorder := a.projectQuiver3DSegments(x[:n], y[:n], z[:n], u[:n], v[:n], w[:n], opt)
	if len(segments) == 0 {
		return nil
	}
	collection := &core.LineCollection{
		Collection: core.Collection{Coords: core.Coords(core.CoordData), Label: opt.Label, Alpha: 1},
		Segments:   segments,
		Color:      color,
		LineWidth:  lineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}
	collection.SetZ(zorder)
	a.Add(collection)
	a.add3DReprojector(func() {
		segments, zorder := a.projectQuiver3DSegments(x[:n], y[:n], z[:n], u[:n], v[:n], w[:n], opt)
		collection.Segments = segments
		collection.SetZ(zorder)
	}, limitsChanged)
	return collection
}

// Quiver3D is an explicit alias for Quiver.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Quiver3D(x, y, z, u, v, w []float64, opt Quiver3DOptions) *core.LineCollection {
	return a.Quiver(x, y, z, u, v, w, opt)
}

// ErrorBar3D plots projected x/y/z error ranges.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) ErrorBar3D(x, y, z, xErr, yErr, zErr []float64, opt ErrorBar3DOptions) *core.LineCollection {
	if a == nil || a.Axes == nil {
		return nil
	}
	n := minLen(x, y, z)
	if n <= 0 {
		return nil
	}
	if !validErrorValues(xErr, n) || !validErrorValues(yErr, n) || !validErrorValues(zErr, n) ||
		!validErrorValues(opt.XErrLower, n) || !validErrorValues(opt.XErrUpper, n) ||
		!validErrorValues(opt.YErrLower, n) || !validErrorValues(opt.YErrUpper, n) ||
		!validErrorValues(opt.ZErrLower, n) || !validErrorValues(opt.ZErrUpper, n) {
		return nil
	}

	limitsChanged := a.observe3DErrorBarData(x[:n], y[:n], z[:n], xErr, yErr, zErr, opt)
	color := a.NextColor()
	if v, ok := opt.Color.Get(); ok {
		color = v
	}
	alpha := 1.0
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && opt.Alpha.OrZero() <= 1 {
		alpha = opt.Alpha.OrZero()
	}
	color = color.WithAlphaMultiplier(alpha)
	lineWidth := 1.0
	if v, ok := opt.LineWidth.Get(); ok {
		lineWidth = v
	}

	segments, zorder := a.projectErrorBar3DSegments(x[:n], y[:n], z[:n], xErr, yErr, zErr, opt)
	if len(segments) == 0 {
		return nil
	}
	collection := &core.LineCollection{
		Collection: core.Collection{Coords: core.Coords(core.CoordData), Label: opt.Label, Alpha: 1},
		Segments:   segments,
		Color:      color,
		LineWidth:  lineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}
	collection.SetZ(zorder)
	a.Add(collection)
	a.add3DReprojector(func() {
		segments, zorder := a.projectErrorBar3DSegments(x[:n], y[:n], z[:n], xErr, yErr, zErr, opt)
		collection.Segments = segments
		collection.SetZ(zorder)
	}, limitsChanged)
	return collection
}

// ErrorBar is the Matplotlib-compatible 3D errorbar entry point.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) ErrorBar(x, y, z, xErr, yErr, zErr []float64, opt ErrorBar3DOptions) *core.LineCollection {
	return a.ErrorBar3D(x, y, z, xErr, yErr, zErr, opt)
}

// PlotSurfaceGrid creates a filled surface from a structured z grid.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) PlotSurfaceGrid(x, y []float64, z [][]float64, opt core.PlotOptions) *core.PolyCollection {
	return a.Surface(x, y, z, opt)
}

// Wireframe draws a structured wireframe as line segments.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes3D) Wireframe(x, y []float64, z [][]float64, opt core.PlotOptions) *core.LineCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	segments := a.projectWireframeSegments(x, y, z, opt)
	if len(segments) == 0 {
		return nil
	}

	// NextColor runs even when the caller supplied a color, so the property
	// cycle advances once per wireframe either way.
	color := opt.Color.Or(a.NextColor())
	alpha := 1.0
	if v, ok := opt.Alpha.Get(); ok && v >= 0 && v <= 1 {
		alpha = v
	}

	collection := &core.LineCollection{
		Collection: core.Collection{
			Coords: core.Coords(core.CoordData),
			Label:  opt.Label,
			Alpha:  alpha,
		},
		Segments: segments,
		Color:    color,
		// points (matplotlib lines.linewidth default); converted at the collection Paint sink
		LineWidth: opt.LineWidth.Or(1.5),
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	collection.SetZ(a.grid3DCollectionZ(x, y, z))
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			collection.Segments = a.projectWireframeSegments(x, y, z, opt)
			collection.SetZ(a.grid3DCollectionZ(x, y, z))
		}
	}, limitsChanged)
	return collection
}

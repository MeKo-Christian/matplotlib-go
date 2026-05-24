package core

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

// Artist is anything that can draw itself with a z-order and optional bounds.
type Artist interface {
	Draw(r render.Renderer, ctx *DrawContext)
	Z() float64
	Bounds(ctx *DrawContext) geom.Rect
}

// WidgetArtist draws in the axes-local widget layer. Widget-layer artists are
// rendered and picked above regular axes artists regardless of data z-order.
type WidgetArtist interface {
	Artist
	WidgetLayer()
}

const (
	defaultPatchZ = 1.0
	defaultLineZ  = 2.0
)

func zOrDefault(z, fallback float64) float64 {
	if z != 0 {
		return z
	}
	return fallback
}

// StickyEdgeArtist can opt an autoscaled data edge out of margin expansion.
// This mirrors Matplotlib's sticky_edges behavior used by bars and images:
// if a margin would cross a sticky value that sits on a data bound, the bound
// is clamped to that value.
type StickyEdgeArtist interface {
	StickyEdges() (x []float64, y []float64)
}

// OverlayArtist is an optional extension for artists that need to render
// outside the axes clip, such as annotations with labels offset from the plot.
type OverlayArtist interface {
	DrawOverlay(r render.Renderer, ctx *DrawContext)
}

// ArtistFunc adapts a function to an Artist.
type ArtistFunc func(r render.Renderer, ctx *DrawContext)

func (f ArtistFunc) Draw(r render.Renderer, ctx *DrawContext) { f(r, ctx) }
func (f ArtistFunc) Z() float64                               { return 0 }
func (f ArtistFunc) Bounds(_ *DrawContext) geom.Rect          { return geom.Rect{} }

// DrawContext carries per-draw state like transforms and style.
type DrawContext struct {
	// DataToPixel maps data coordinates to pixels.
	DataToPixel Transform2D
	// Axes is the owning axes for axes-local drawing.
	Axes *Axes
	// Projection configures data->axes mapping and non-Cartesian behavior.
	Projection Projection
	// Styling configuration in effect.
	RC style.RC
	// Clip is the axes pixel rectangle.
	Clip geom.Rect
	// FigureRect is the figure display rectangle in pixels.
	FigureRect geom.Rect
	// DrawOptions selects which artists participate in this draw pass. The
	// zero value matches Matplotlib's default Artist.draw_wrapper behavior:
	// animated artists are skipped, and the animation engine flips this
	// during overlay/background passes.
	DrawOptions DrawOptions
}

// AnimatedFilter selects which artists are drawn during a figure draw pass.
type AnimatedFilter uint8

const (
	// AnimatedFilterExcludeAnimated draws non-animated artists only. This is
	// the default and matches Matplotlib's per-frame figure draw.
	AnimatedFilterExcludeAnimated AnimatedFilter = iota
	// AnimatedFilterOnlyAnimated draws animated artists only. Used by the
	// animation engine to redraw the animated overlay on top of a saved
	// background region.
	AnimatedFilterOnlyAnimated
	// AnimatedFilterAll draws every artist regardless of animated state.
	// Useful for one-shot saves where blit semantics do not apply.
	AnimatedFilterAll
)

// DrawOptions controls what a figure draw pass renders. The zero value matches
// Matplotlib's default Artist.draw_wrapper behavior: animated artists are
// skipped so that the animation engine can manage them via a blit-style
// overlay path.
type DrawOptions struct {
	AnimatedFilter AnimatedFilter
}

// Transform2D wires x/y scales with an axes->pixel affine transform.
type Transform2D struct {
	XScale      transform.Scale
	YScale      transform.Scale
	DataToAxes  transform.T
	AxesToPixel transform.T
}

// Apply transforms a data-space point to pixel coordinates.
func (t *Transform2D) Apply(p geom.Pt) geom.Pt {
	tr := t.transData()
	if tr == nil {
		return p
	}
	return tr.Apply(p)
}

// Invert transforms a pixel-space point back into data coordinates.
func (t *Transform2D) Invert(p geom.Pt) (geom.Pt, bool) {
	tr := t.transData()
	if tr == nil {
		return p, true
	}
	return tr.Invert(p)
}

func (t *Transform2D) transData() transform.T {
	if t == nil {
		return nil
	}

	dataToAxes := t.DataToAxes
	if dataToAxes == nil {
		dataToAxes = transform.NewScaleTransform(t.XScale, t.YScale)
	}

	switch {
	case dataToAxes == nil:
		return t.AxesToPixel
	case t.AxesToPixel == nil:
		return dataToAxes
	default:
		return transform.Chain{A: dataToAxes, B: t.AxesToPixel}
	}
}

// CoordinateSpace identifies a Matplotlib-style coordinate system.
type CoordinateSpace uint8

const (
	CoordData CoordinateSpace = iota
	CoordAxes
	CoordFigure
)

// CoordinateSpec identifies the x/y coordinate spaces used by a point.
type CoordinateSpec struct {
	X CoordinateSpace
	Y CoordinateSpace
}

// Coords uses the same coordinate space for x and y.
func Coords(space CoordinateSpace) CoordinateSpec {
	return CoordinateSpec{X: space, Y: space}
}

// BlendCoords uses separate coordinate spaces for x and y.
func BlendCoords(xSpace, ySpace CoordinateSpace) CoordinateSpec {
	return CoordinateSpec{X: xSpace, Y: ySpace}
}

// TransData returns the Matplotlib-style data->display transform.
func (ctx *DrawContext) TransData() transform.T {
	if ctx == nil {
		return nil
	}
	return ctx.DataToPixel.transData()
}

// TransProjection returns the projection-specific data->axes transform stage
// before the final axes->display mapping is applied.
func (ctx *DrawContext) TransProjection() transform.T {
	if ctx == nil {
		return nil
	}
	if ctx.DataToPixel.DataToAxes != nil {
		return ctx.DataToPixel.DataToAxes
	}
	return transform.NewScaleTransform(ctx.DataToPixel.XScale, ctx.DataToPixel.YScale)
}

// TransAxes returns the Matplotlib-style axes-fraction->display transform.
func (ctx *DrawContext) TransAxes() transform.T {
	if ctx == nil {
		return nil
	}
	if ctx.DataToPixel.AxesToPixel != nil {
		return ctx.DataToPixel.AxesToPixel
	}
	if rect, ok := unitSquareBounds(nil, ctx.Clip); ok {
		return transform.NewDisplayRectTransform(rect)
	}
	return nil
}

// TransFigure returns the Matplotlib-style figure-fraction->display transform.
func (ctx *DrawContext) TransFigure() transform.T {
	if ctx == nil {
		return nil
	}
	rect := ctx.FigureRect
	if rect == (geom.Rect{}) {
		rect = ctx.Clip
	}
	if rect == (geom.Rect{}) {
		return nil
	}
	return transform.NewDisplayRectTransform(rect)
}

// TransformFor resolves a coordinate specification into a display transform.
func (ctx *DrawContext) TransformFor(spec CoordinateSpec) transform.T {
	if spec.X == spec.Y {
		return ctx.transformForSpace(spec.X)
	}

	xTrans, okX := ctx.separableTransformForSpace(spec.X)
	yTrans, okY := ctx.separableTransformForSpace(spec.Y)
	if !okX || !okY {
		return nil
	}
	return transform.Blend(xTrans, yTrans)
}

func (ctx *DrawContext) transformForSpace(space CoordinateSpace) transform.T {
	switch space {
	case CoordAxes:
		return ctx.TransAxes()
	case CoordFigure:
		return ctx.TransFigure()
	default:
		return ctx.TransData()
	}
}

func (ctx *DrawContext) separableTransformForSpace(space CoordinateSpace) (transform.Separable, bool) {
	switch space {
	case CoordAxes:
		return ctx.separableAxesTransform()
	case CoordFigure:
		tr := ctx.TransFigure()
		sep, ok := tr.(transform.Separable)
		return sep, ok
	default:
		return ctx.separableDataTransform()
	}
}

func (ctx *DrawContext) separableAxesTransform() (transform.Separable, bool) {
	if ctx == nil {
		return transform.SeparableT{}, false
	}
	if sep, ok := ctx.DataToPixel.AxesToPixel.(transform.Separable); ok {
		return sep, true
	}
	rect, ok := unitSquareBounds(ctx.DataToPixel.AxesToPixel, ctx.Clip)
	if !ok {
		return transform.SeparableT{}, false
	}
	return transform.NewDisplayRectTransform(rect), true
}

func (ctx *DrawContext) separableDataTransform() (transform.Separable, bool) {
	axesTrans, ok := ctx.separableAxesTransform()
	if !ok {
		return transform.SeparableT{}, false
	}

	if ctx != nil && ctx.DataToPixel.DataToAxes != nil {
		sep, ok := ctx.DataToPixel.DataToAxes.(transform.Separable)
		if !ok {
			return transform.SeparableT{}, false
		}
		return transform.ChainSeparable(sep, axesTrans), true
	}

	return transform.ChainSeparable(
		transform.NewScaleTransform(ctx.DataToPixel.XScale, ctx.DataToPixel.YScale),
		axesTrans,
	), true
}

func unitSquareBounds(tr transform.T, fallback geom.Rect) (geom.Rect, bool) {
	if tr == nil {
		if fallback == (geom.Rect{}) {
			return geom.Rect{}, false
		}
		return fallback, true
	}

	corners := []geom.Pt{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}

	rect := geom.Rect{Min: tr.Apply(corners[0]), Max: tr.Apply(corners[0])}
	for _, corner := range corners[1:] {
		pt := tr.Apply(corner)
		if pt.X < rect.Min.X {
			rect.Min.X = pt.X
		}
		if pt.Y < rect.Min.Y {
			rect.Min.Y = pt.Y
		}
		if pt.X > rect.Max.X {
			rect.Max.X = pt.X
		}
		if pt.Y > rect.Max.Y {
			rect.Max.Y = pt.Y
		}
	}
	return rect, true
}

// Figure is the root of the Artist tree. It contains Axes children.
type Figure struct {
	SizePx    geom.Pt
	RC        style.RC
	Children  []*Axes
	Artists   []Artist
	zsorted   bool
	SupTitle  string
	SupXLabel string
	SupYLabel string

	layoutEngine LayoutEngine
}

// NewFigure creates a new figure with pixel dimensions and optional style overrides.
func NewFigure(w, h int, opts ...style.Option) *Figure {
	rc := style.Apply(style.CurrentDefaults(), opts...)
	return &Figure{
		SizePx:       geom.Pt{X: float64(w), Y: float64(h)},
		RC:           rc,
		Children:     nil,
		Artists:      nil,
		SupTitle:     "",
		SupXLabel:    "",
		SupYLabel:    "",
		layoutEngine: LayoutEngineNone,
	}
}

// Axes represents an axes region inside a figure.
type Axes struct {
	RectFraction  geom.Rect // [0..1] fraction in figure coords
	RC            *style.RC // nil => inherit figure RC
	XScale        transform.Scale
	YScale        transform.Scale
	projection    Projection
	Artists       []Artist
	zsorted       bool
	WidgetArtists []Artist
	widgetZsorted bool

	// Axis control
	XAxis      *Axis // bottom x-axis
	YAxis      *Axis // left y-axis
	XAxisTop   *Axis // optional top x-axis
	YAxisRight *Axis // optional right y-axis
	ExtraAxes  []*Axis
	ShowFrame  bool // draw top and right border lines when no explicit top/right axis exists

	// Text labels
	Title  string // title above the plot
	XLabel string // x-axis label below ticks
	YLabel string // y-axis label left of ticks

	// Color cycling for multiple series. Matplotlib keeps separate cycles for
	// line artists and shape/fill artists, so scatter markers do not advance
	// the line cycle.
	ColorCycle      *color.ColorCycle
	PatchColorCycle *color.ColorCycle

	aspectMode  string
	aspectValue float64
	boxAspect   float64
	xLabelSide  AxisSide
	yLabelSide  AxisSide

	shareX *Axes
	shareY *Axes
	figure *Figure

	xLimitsManual bool
	yLimitsManual bool

	xUnits *axisUnitsState
	yUnits *axisUnitsState

	subplotSpec    *SubplotSpec
	axesLocator    AxesLocator
	positionManual bool

	childAxes []*Axes

	colorbarParent   *Axes
	colorbarWidth    float64
	colorbarPadding  float64
	colorbarAspect   float64
	colorbarBase     geom.Rect
	colorbarExtend   string
	colorbarLocation string
	colorbarTicks    []float64
	colorbarBounds   []float64

	coordFormatter CoordFormatter
}

// TickParams controls axis tick visibility and styling.
type TickParams struct {
	Reset         bool
	Axis          string
	Which         string
	Color         *render.Color
	Length        *float64
	Width         *float64
	Direction     *string
	ShowTicks     *bool
	ShowLabels    *bool
	Top           *bool
	Bottom        *bool
	Left          *bool
	Right         *bool
	LabelTop      *bool
	LabelBottom   *bool
	LabelLeft     *bool
	LabelRight    *bool
	LabelRotation *float64
	LabelSize     *float64
	LabelPad      *float64
	LabelHAlign   *TextAlign
	LabelVAlign   *TextVerticalAlign
	GridVisible   *bool
	GridColor     *render.Color
	GridAlpha     *float64
	GridLineWidth *float64
	GridDashes    []float64
}

// LocatorParams controls the target tick density for automatic locators.
type LocatorParams struct {
	Axis       string
	MajorCount int
	MinorCount int
}

// AddAxes appends an Axes to the Figure. If opts are provided, the Axes gets its
// own RC copy; otherwise it inherits from the Figure.
func (f *Figure) AddAxes(r geom.Rect, opts ...style.Option) *Axes {
	proj, _ := lookupProjection("rectilinear")
	return f.addAxesWithProjection(r, proj, opts...)
}

// AddAxesProjection appends an Axes with the requested named projection.
func (f *Figure) AddAxesProjection(r geom.Rect, projection string, opts ...style.Option) (*Axes, error) {
	proj, err := lookupProjection(projection)
	if err != nil {
		return nil, err
	}
	return f.addAxesWithProjection(r, proj, opts...), nil
}

// AddPolarAxes appends an Axes configured with the built-in polar projection.
func (f *Figure) AddPolarAxes(r geom.Rect, opts ...style.Option) *Axes {
	ax, err := f.AddAxesProjection(r, "polar", opts...)
	if err != nil {
		return nil
	}
	return ax
}

// AddRadarAxes appends an Axes configured with the built-in radar projection.
// Labels, when provided, define the spoke count and angular tick labels.
func (f *Figure) AddRadarAxes(r geom.Rect, labels []string, opts ...style.Option) (*Axes, error) {
	if len(labels) > 0 && len(labels) < 3 {
		return nil, fmt.Errorf("radar axes require at least 3 labels")
	}
	ax, err := f.AddAxesProjection(r, "radar", opts...)
	if err != nil {
		return nil, err
	}
	if len(labels) > 0 {
		if err := ax.SetRadarLabels(labels); err != nil {
			return nil, err
		}
	}
	return ax, nil
}

// AddSkewXAxes appends an Axes configured with the built-in skewx projection,
// commonly used as the basis for Skew-T style meteorological plots.
func (f *Figure) AddSkewXAxes(r geom.Rect, opts ...style.Option) (*Axes, error) {
	return f.AddAxesProjection(r, "skewx", opts...)
}

// AddAxes3D appends an Axes configured with the built-in 3D projection.
func (f *Figure) AddAxes3D(r geom.Rect, opts ...style.Option) (*Axes3D, error) {
	ax, err := f.AddAxesProjection(r, "3d", opts...)
	if err != nil {
		return nil, err
	}
	return NewAxes3D(ax), nil
}

func (f *Figure) addAxesWithProjection(r geom.Rect, proj Projection, opts ...style.Option) *Axes {
	var rc *style.RC
	effective := f.RC
	if len(opts) > 0 {
		v := style.Apply(f.RC, opts...)
		rc = &v
		effective = v
	}
	ax := &Axes{
		RectFraction:    r,
		RC:              rc,
		XScale:          transform.NewLinear(0, 1),
		YScale:          transform.NewLinear(0, 1),
		projection:      cloneProjection(proj),
		XAxis:           NewXAxis(),
		YAxis:           NewYAxis(),
		ShowFrame:       true,
		ColorCycle:      color.NewColorCycle(effective.Palette()),
		PatchColorCycle: color.NewColorCycle(effective.Palette()),
		aspectMode:      "auto",
		aspectValue:     1,
		xLabelSide:      AxisBottom,
		yLabelSide:      AxisLeft,
		figure:          f,
	}
	if ax.projection == nil {
		ax.projection, _ = lookupProjection("rectilinear")
	}
	ax.projection.ConfigureAxes(ax)
	ax.applyStyleDefaults(effective)
	ax.addDefaultGrids(effective)
	f.Children = append(f.Children, ax)
	return ax
}

// Add registers a figure-level Artist.
func (f *Figure) Add(art Artist) {
	if f == nil {
		return
	}
	f.Artists = append(f.Artists, art)
	f.zsorted = false
}

// Add registers an Artist with the Axes.
func (a *Axes) Add(art Artist) {
	if a == nil {
		return
	}
	if _, ok := art.(WidgetArtist); ok {
		a.AddWidget(art)
		return
	}
	a.Artists = append(a.Artists, art)
	a.zsorted = false
}

// AddWidget registers an Artist with the axes-local widget layer.
func (a *Axes) AddWidget(art Artist) {
	if a == nil {
		return
	}
	a.WidgetArtists = append(a.WidgetArtists, art)
	a.widgetZsorted = false
}

// SetPosition sets the active axes rectangle in figure-normalized
// coordinates. For subplot-backed axes this mirrors Matplotlib's ability to
// keep the subplot spec while overriding the drawn position.
func (a *Axes) SetPosition(r geom.Rect) {
	if a == nil {
		return
	}
	a.RectFraction = r
	a.positionManual = true
}

// ProjectionName reports the normalized name of the axes projection.
func (a *Axes) ProjectionName() string {
	if a == nil || a.projection == nil {
		return "rectilinear"
	}
	return a.projection.Name()
}

// SetThetaZeroLocation changes the display angle used for theta=0 on polar
// axes. Supported values include the compass points E, NE, N, NW, W, SW, S,
// and SE, plus the full direction names.
func (a *Axes) SetThetaZeroLocation(location string, offsetDeg ...float64) error {
	proj, ok := polarProjectionForAxes(a)
	if !ok {
		return fmt.Errorf("theta zero location requires polar axes")
	}

	base, ok := polarCompassAngle(location)
	if !ok {
		return fmt.Errorf("unsupported theta zero location %q", location)
	}
	if len(offsetDeg) > 0 {
		base += offsetDeg[0] * math.Pi / 180
	}
	proj.thetaOffset = normalizePolarAngle(base)
	return nil
}

// SetThetaDirection changes whether theta increases clockwise or
// counterclockwise on polar axes.
func (a *Axes) SetThetaDirection(direction string) error {
	proj, ok := polarProjectionForAxes(a)
	if !ok {
		return fmt.Errorf("theta direction requires polar axes")
	}

	switch strings.ToLower(strings.TrimSpace(direction)) {
	case "counterclockwise", "anticlockwise", "ccw", "1", "+1", "positive":
		proj.thetaDirection = 1
	case "clockwise", "cw", "-1", "negative":
		proj.thetaDirection = -1
	default:
		return fmt.Errorf("unsupported theta direction %q", direction)
	}
	return nil
}

// SetRadialLabelPosition changes the angle used for radial ticks, labels, and
// the radial spine on polar axes.
func (a *Axes) SetRadialLabelPosition(angleDeg float64) error {
	proj, ok := polarProjectionForAxes(a)
	if !ok {
		return fmt.Errorf("radial label position requires polar axes")
	}
	proj.radialLabelAngle = normalizePolarAngle(angleDeg * math.Pi / 180)
	return nil
}

// SetRadarVariableCount changes the number of equally spaced spokes on a radar
// projection and resets spoke labels to numeric defaults.
func (a *Axes) SetRadarVariableCount(n int) error {
	proj, ok := radarProjectionForAxes(a)
	if !ok {
		return fmt.Errorf("radar variable count requires radar axes")
	}
	if n < 3 {
		return fmt.Errorf("radar axes require at least 3 variables")
	}
	proj.radarVariables = n
	proj.radarLabels = nil
	configureRadarThetaAxis(a, proj)
	return nil
}

// SetRadarLabels changes the spoke labels on a radar projection. The label
// count defines the number of spokes.
func (a *Axes) SetRadarLabels(labels []string) error {
	proj, ok := radarProjectionForAxes(a)
	if !ok {
		return fmt.Errorf("radar labels require radar axes")
	}
	if len(labels) < 3 {
		return fmt.Errorf("radar axes require at least 3 labels")
	}
	proj.radarVariables = len(labels)
	proj.radarLabels = append([]string(nil), labels...)
	configureRadarThetaAxis(a, proj)
	return nil
}

// SetSkewXAngle changes the skew angle used by skewx projection axes.
func (a *Axes) SetSkewXAngle(angleDeg float64) error {
	proj, ok := skewXProjectionForAxes(a)
	if !ok {
		return fmt.Errorf("skewx angle requires skewx axes")
	}
	if math.IsNaN(angleDeg) || math.IsInf(angleDeg, 0) {
		return fmt.Errorf("skewx angle must be finite")
	}
	cos := math.Cos(angleDeg * math.Pi / 180)
	if math.Abs(cos) < 1e-6 {
		return fmt.Errorf("skewx angle %g is too close to vertical", angleDeg)
	}
	proj.angleDeg = angleDeg
	return nil
}

// SetXLim sets the x-axis limits.
func (a *Axes) SetXLim(minVal, maxVal float64) {
	target := a.xScaleRoot()
	target.XScale = replaceScaleDomain(target.XScale, minVal, maxVal)
	target.xLimitsManual = true
	target.refreshUnitAxis(true)
}

// SetYLim sets the y-axis limits.
func (a *Axes) SetYLim(minVal, maxVal float64) {
	target := a.yScaleRoot()
	target.YScale = replaceScaleDomain(target.YScale, minVal, maxVal)
	target.yLimitsManual = true
	target.refreshUnitAxis(false)
}

// SetXScale replaces the x-axis scale while preserving the current view limits.
func (a *Axes) SetXScale(name string, opts ...transform.ScaleOption) error {
	return a.setScale(true, name, opts...)
}

// SetYScale replaces the y-axis scale while preserving the current view limits.
func (a *Axes) SetYScale(name string, opts ...transform.ScaleOption) error {
	return a.setScale(false, name, opts...)
}

// SetXLimLog sets the x-axis to logarithmic scale with given limits.
func (a *Axes) SetXLimLog(minVal, maxVal, base float64) {
	target := a.xScaleRoot()
	if state := target.unitState(true); state != nil && !state.scaleCompatible("log") {
		return
	}
	target.XScale = transform.NewLog(minVal, maxVal, base)
	target.xLimitsManual = true
	configureScaleAxes(target.XAxis, target.XAxisTop, "log", transform.ResolveScaleOptions(
		transform.WithScaleDomain(minVal, maxVal),
		transform.WithScaleBase(base),
	))
}

// SetYLimLog sets the y-axis to logarithmic scale with given limits.
func (a *Axes) SetYLimLog(minVal, maxVal, base float64) {
	target := a.yScaleRoot()
	if state := target.unitState(false); state != nil && !state.scaleCompatible("log") {
		return
	}
	target.YScale = transform.NewLog(minVal, maxVal, base)
	target.yLimitsManual = true
	configureScaleAxes(target.YAxis, target.YAxisRight, "log", transform.ResolveScaleOptions(
		transform.WithScaleDomain(minVal, maxVal),
		transform.WithScaleBase(base),
	))
}

// InvertX reverses the x-axis direction while preserving the underlying scale type.
func (a *Axes) InvertX() {
	target := a.xScaleRoot()
	if target == nil || target.XScale == nil {
		return
	}
	target.XScale = toggleInvertedScale(target.XScale)
}

// InvertY reverses the y-axis direction while preserving the underlying scale type.
func (a *Axes) InvertY() {
	target := a.yScaleRoot()
	if target == nil || target.YScale == nil {
		return
	}
	target.YScale = toggleInvertedScale(target.YScale)
}

// XInverted reports whether the effective x-axis direction is reversed.
func (a *Axes) XInverted() bool {
	if a == nil {
		return false
	}
	return scaleDomainDescending(a.effectiveXScale())
}

// YInverted reports whether the effective y-axis direction is reversed.
func (a *Axes) YInverted() bool {
	if a == nil {
		return false
	}
	return scaleDomainDescending(a.effectiveYScale())
}

// AutoScale computes axis limits from the data bounds of all artists,
// adding the specified margin fraction on each side (e.g. 0.05 = 5%).
// A margin of 0 fits exactly to the data. If no artists have non-zero bounds,
// limits remain unchanged.
func (a *Axes) AutoScale(margin float64) {
	a.autoScaleAxis(true, margin, false)
	a.autoScaleAxis(false, margin, false)
}

func (a *Axes) autoScaleIfEnabled(margin float64) {
	if a == nil || a.ProjectionName() != "rectilinear" {
		return
	}
	a.autoScaleAxis(true, margin, true)
	a.autoScaleAxis(false, margin, true)
}

func (a *Axes) autoScaleAxis(isX bool, margin float64, respectManual bool) {
	if a == nil {
		return
	}

	target := a.xScaleRoot()
	if !isX {
		target = a.yScaleRoot()
	}
	if target == nil {
		return
	}
	if respectManual {
		if isX && target.xLimitsManual {
			return
		}
		if !isX && target.yLimitsManual {
			return
		}
	}

	var minVal, maxVal float64
	var stickies []float64
	first := true

	for _, ax := range a.autoscaleAxesForTarget(isX, target) {
		for _, art := range ax.Artists {
			b := art.Bounds(nil)
			if b.W() == 0 && b.H() == 0 && b.Min.X == 0 && b.Min.Y == 0 {
				continue // skip zero-bounds artists (grids, etc.)
			}
			lo, hi := b.Min.X, b.Max.X
			if !isX {
				lo, hi = b.Min.Y, b.Max.Y
			}
			if math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) {
				continue
			}
			if sticky, ok := art.(StickyEdgeArtist); ok {
				xSticky, ySticky := sticky.StickyEdges()
				if isX {
					stickies = appendFinite(stickies, xSticky...)
				} else {
					stickies = appendFinite(stickies, ySticky...)
				}
			}
			if first {
				minVal, maxVal = lo, hi
				first = false
				continue
			}
			if lo < minVal {
				minVal = lo
			}
			if hi > maxVal {
				maxVal = hi
			}
		}
	}
	if first {
		return // no data artists
	}

	span := maxVal - minVal
	if span == 0 {
		span = 1 // avoid zero-span
	}
	lowerSticky, upperSticky := stickyBounds(minVal, maxVal, stickies)
	minVal -= span * margin
	maxVal += span * margin
	if !math.IsNaN(lowerSticky) && minVal < lowerSticky {
		minVal = lowerSticky
	}
	if !math.IsNaN(upperSticky) && maxVal > upperSticky {
		maxVal = upperSticky
	}

	if isX {
		target.XScale = replaceScaleDomain(target.XScale, minVal, maxVal)
		target.refreshUnitAxis(true)
	} else {
		target.YScale = replaceScaleDomain(target.YScale, minVal, maxVal)
		target.refreshUnitAxis(false)
	}
}

func (a *Axes) autoscaleAxesForTarget(isX bool, target *Axes) []*Axes {
	if a == nil {
		return nil
	}
	fig := a.figure
	if fig == nil || target == nil {
		return []*Axes{a}
	}

	axes := make([]*Axes, 0, len(fig.Children))
	for _, candidate := range fig.Children {
		if candidate == nil {
			continue
		}
		root := candidate.xScaleRoot()
		if !isX {
			root = candidate.yScaleRoot()
		}
		if root == target {
			axes = append(axes, candidate)
		}
	}
	if len(axes) == 0 {
		return []*Axes{a}
	}
	return axes
}

func appendFinite(dst []float64, values ...float64) []float64 {
	for _, value := range values {
		if !math.IsNaN(value) && !math.IsInf(value, 0) {
			dst = append(dst, value)
		}
	}
	return dst
}

func stickyBounds(minVal, maxVal float64, stickies []float64) (float64, float64) {
	if len(stickies) == 0 {
		return math.NaN(), math.NaN()
	}
	sort.Float64s(stickies)

	tol := 1e-5 * math.Abs(maxVal-minVal)
	lower := math.NaN()
	upper := math.NaN()
	for _, sticky := range stickies {
		if sticky < minVal+tol {
			lower = sticky
		}
		if math.IsNaN(upper) && sticky > maxVal-tol {
			upper = sticky
		}
	}
	return lower, upper
}

// AddGrid adds grid lines for the specified axis.
func (a *Axes) AddGrid(axis AxisSide) *Grid {
	grid := NewGrid(axis)
	rc := a.resolvedRC()
	grid.Color = rc.GridColor
	grid.LineWidth = rc.GridLineWidth
	grid.MinorColor = rc.MinorGridColor
	grid.MinorLineWidth = rc.MinorGridLineWidth
	grid.Dashes = styleCloneDashes(rc.GridDashes)
	grid.MinorDashes = styleCloneDashes(rc.MinorGridDashes)
	if isPolarProjection(a.projection) {
		grid.z = 1.5
	}
	a.Add(grid)
	return grid
}

// AddXGrid adds vertical grid lines based on x-axis ticks.
func (a *Axes) AddXGrid() *Grid {
	return a.AddGrid(AxisBottom)
}

// AddYGrid adds horizontal grid lines based on y-axis ticks.
func (a *Axes) AddYGrid() *Grid {
	return a.AddGrid(AxisLeft)
}

// SetTitle sets the title displayed above the plot.
func (a *Axes) SetTitle(title string) { a.Title = title }

// SetSuptitle sets the figure-level title.
func (f *Figure) SetSuptitle(title string) {
	if f != nil {
		f.SupTitle = title
	}
}

// SetSupTitle is an alias for SetSuptitle.
func (f *Figure) SetSupTitle(title string) { f.SetSuptitle(title) }

// SetSupxlabel sets the figure-level x label.
func (f *Figure) SetSupxlabel(label string) {
	if f != nil {
		f.SupXLabel = label
	}
}

// SetSupXLabel is an alias for SetSupxlabel.
func (f *Figure) SetSupXLabel(label string) { f.SetSupxlabel(label) }

// SetSupylabel sets the figure-level y label.
func (f *Figure) SetSupylabel(label string) {
	if f != nil {
		f.SupYLabel = label
	}
}

// SetSupYLabel is an alias for SetSupylabel.
func (f *Figure) SetSupYLabel(label string) { f.SetSupylabel(label) }

// SetXLabel sets the label displayed below the x-axis.
func (a *Axes) SetXLabel(label string) { a.XLabel = label }

// SetYLabel sets the label displayed left of the y-axis.
func (a *Axes) SetYLabel(label string) { a.YLabel = label }

// SetXLabelPosition controls whether the x-axis label is placed above or below the axes.
func (a *Axes) SetXLabelPosition(position string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "", "bottom":
		a.xLabelSide = AxisBottom
	case "top":
		a.xLabelSide = AxisTop
	default:
		return fmt.Errorf("unsupported x label position %q", position)
	}
	return nil
}

// SetYLabelPosition controls whether the y-axis label is placed left or right of the axes.
func (a *Axes) SetYLabelPosition(position string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "", "left":
		a.yLabelSide = AxisLeft
	case "right":
		a.yLabelSide = AxisRight
	default:
		return fmt.Errorf("unsupported y label position %q", position)
	}
	return nil
}

// SetAspect configures the data aspect ratio for the axes box.
// Supported values are "auto", "equal", and "ratio" (which requires one numeric value).
func (a *Axes) SetAspect(mode string, value ...float64) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "auto":
		a.aspectMode = "auto"
		a.aspectValue = 1
	case "equal":
		a.aspectMode = "equal"
		a.aspectValue = 1
	case "ratio":
		if len(value) == 0 || value[0] <= 0 || math.IsNaN(value[0]) || math.IsInf(value[0], 0) {
			return fmt.Errorf("ratio aspect requires a positive finite value")
		}
		a.aspectMode = "ratio"
		a.aspectValue = value[0]
	default:
		return fmt.Errorf("unsupported aspect mode %q", mode)
	}
	return nil
}

// SetAxisEqual is a convenience helper for a 1:1 data aspect ratio.
func (a *Axes) SetAxisEqual() {
	_ = a.SetAspect("equal")
}

// SetBoxAspect constrains the physical height/width ratio of the axes box.
func (a *Axes) SetBoxAspect(aspect float64) error {
	if a == nil {
		return nil
	}
	if aspect <= 0 || math.IsNaN(aspect) || math.IsInf(aspect, 0) {
		return fmt.Errorf("box aspect must be a positive finite value")
	}
	a.boxAspect = aspect
	return nil
}

// ClearBoxAspect removes any physical box-aspect constraint.
func (a *Axes) ClearBoxAspect() {
	if a == nil {
		return
	}
	a.boxAspect = 0
}

// NextColor returns the next color in the axes color cycle.
func (a *Axes) NextColor() render.Color {
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	return a.ColorCycle.Next()
}

// PeekColor returns the current color without advancing the cycle.
func (a *Axes) PeekColor() render.Color {
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	return a.ColorCycle.Peek()
}

func (a *Axes) colorCycleAt(index int) render.Color {
	palette := a.resolvedRC().Palette()
	if len(palette) == 0 {
		return render.Color{A: 1}
	}
	idx := index % len(palette)
	if idx < 0 {
		idx += len(palette)
	}
	return palette[idx]
}

// NextPatchColor returns the next color in the shape/fill cycle.
func (a *Axes) NextPatchColor() render.Color {
	if a.PatchColorCycle == nil {
		a.PatchColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	return a.PatchColorCycle.Next()
}

// PeekPatchColor returns the current shape/fill color without advancing.
func (a *Axes) PeekPatchColor() render.Color {
	if a.PatchColorCycle == nil {
		a.PatchColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	return a.PatchColorCycle.Peek()
}

// ResetColorCycle resets the color cycle to the first color.
func (a *Axes) ResetColorCycle() {
	if a.ColorCycle != nil {
		a.ColorCycle.Reset()
	}
	if a.PatchColorCycle != nil {
		a.PatchColorCycle.Reset()
	}
}

// TopAxis returns the explicit top x-axis, creating it on first use.
func (a *Axes) TopAxis() *Axis {
	if a == nil {
		return nil
	}
	if a.XAxisTop == nil {
		a.XAxisTop = cloneAxisForSide(a.XAxis, AxisTop)
	}
	return a.XAxisTop
}

// RightAxis returns the explicit right y-axis, creating it on first use.
func (a *Axes) RightAxis() *Axis {
	if a == nil {
		return nil
	}
	if a.YAxisRight == nil {
		a.YAxisRight = cloneAxisForSide(a.YAxis, AxisRight)
	}
	return a.YAxisRight
}

// HideTopAxis removes the explicit top x-axis.
func (a *Axes) HideTopAxis() {
	if a == nil {
		return
	}
	a.XAxisTop = nil
}

// HideRightAxis removes the explicit right y-axis.
func (a *Axes) HideRightAxis() {
	if a == nil {
		return
	}
	a.YAxisRight = nil
}

// SetXAxisSide repositions the primary x-axis to the requested side.
func (a *Axes) SetXAxisSide(side AxisSide) error {
	if a == nil {
		return nil
	}
	if side != AxisBottom && side != AxisTop {
		return fmt.Errorf("x-axis side must be bottom or top")
	}
	if a.XAxis == nil {
		a.XAxis = cloneAxisForSide(nil, side)
	} else {
		a.XAxis.Side = side
	}
	if side == AxisTop {
		a.XAxisTop = nil
	}
	return nil
}

// SetYAxisSide repositions the primary y-axis to the requested side.
func (a *Axes) SetYAxisSide(side AxisSide) error {
	if a == nil {
		return nil
	}
	if side != AxisLeft && side != AxisRight {
		return fmt.Errorf("y-axis side must be left or right")
	}
	if a.YAxis == nil {
		a.YAxis = cloneAxisForSide(nil, side)
	} else {
		a.YAxis.Side = side
	}
	if side == AxisRight {
		a.YAxisRight = nil
	}
	return nil
}

// MoveXAxisToTop is a convenience helper that repositions the primary x-axis.
func (a *Axes) MoveXAxisToTop() error { return a.SetXAxisSide(AxisTop) }

// MoveXAxisToBottom is a convenience helper that repositions the primary x-axis.
func (a *Axes) MoveXAxisToBottom() error { return a.SetXAxisSide(AxisBottom) }

// MoveYAxisToLeft is a convenience helper that repositions the primary y-axis.
func (a *Axes) MoveYAxisToLeft() error { return a.SetYAxisSide(AxisLeft) }

// MoveYAxisToRight is a convenience helper that repositions the primary y-axis.
func (a *Axes) MoveYAxisToRight() error { return a.SetYAxisSide(AxisRight) }

// SetXTickLabelPosition controls whether x tick labels appear on the bottom,
// top, both, or neither side.
func (a *Axes) SetXTickLabelPosition(position string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "", "bottom":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = true
		}
		if a.XAxisTop != nil {
			a.XAxisTop.ShowLabels = false
		}
	case "top":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = false
		}
		a.TopAxis().ShowLabels = true
	case "both":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = true
		}
		a.TopAxis().ShowLabels = true
	case "none":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = false
		}
		if a.XAxisTop != nil {
			a.XAxisTop.ShowLabels = false
		}
	default:
		return fmt.Errorf("unsupported x tick label position %q", position)
	}
	return nil
}

// SetYTickLabelPosition controls whether y tick labels appear on the left,
// right, both, or neither side.
func (a *Axes) SetYTickLabelPosition(position string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "", "left":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = true
		}
		if a.YAxisRight != nil {
			a.YAxisRight.ShowLabels = false
		}
	case "right":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = false
		}
		a.RightAxis().ShowLabels = true
	case "both":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = true
		}
		a.RightAxis().ShowLabels = true
	case "none":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = false
		}
		if a.YAxisRight != nil {
			a.YAxisRight.ShowLabels = false
		}
	default:
		return fmt.Errorf("unsupported y tick label position %q", position)
	}
	return nil
}

// MinorticksOn enables default minor locators on the requested axis selection.
func (a *Axes) MinorticksOn(axis string) error {
	if err := validateAxisSpec(axis); err != nil {
		return err
	}
	for _, ax := range a.axesForSpec(axis) {
		enableMinorTicks(ax)
	}
	return nil
}

// MinorticksOff disables minor locators on the requested axis selection.
func (a *Axes) MinorticksOff(axis string) error {
	if err := validateAxisSpec(axis); err != nil {
		return err
	}
	for _, ax := range a.axesForSpec(axis) {
		if ax != nil {
			ax.MinorLocator = nil
		}
	}
	return nil
}

// LocatorParams updates the target major/minor tick density for the selected axes.
func (a *Axes) LocatorParams(params LocatorParams) error {
	if err := validateAxisSpec(params.Axis); err != nil {
		return err
	}
	for _, axis := range a.axesForSpec(params.Axis) {
		if axis == nil {
			continue
		}
		if params.MajorCount > 0 {
			axis.MajorTickCount = params.MajorCount
		}
		if params.MinorCount > 0 {
			axis.MinorTickCount = params.MinorCount
		}
	}
	return nil
}

// TickParams applies visibility/styling updates to the selected axis ticks.
func (a *Axes) TickParams(params TickParams) error {
	if err := validateAxisSpec(params.Axis); err != nil {
		return err
	}
	which := normalizeTickWhich(params.Which)
	if which == "" {
		return fmt.Errorf("unsupported tick selection %q", params.Which)
	}

	for _, axis := range a.axesForSpec(params.Axis) {
		if axis == nil {
			continue
		}
		if params.Reset {
			resetAxisTickParams(axis)
		}
		if params.Color != nil {
			tickColor := *params.Color
			labelColor := *params.Color
			axis.TickColor = &tickColor
			axis.TickLabelColor = &labelColor
		}
		if params.Width != nil {
			switch which {
			case "major":
				axis.TickLineWidth = *params.Width
			case "minor":
				axis.MinorTickLineWidth = *params.Width
			case "both":
				axis.TickLineWidth = *params.Width
				axis.MinorTickLineWidth = *params.Width
			}
		}
		if params.ShowTicks != nil {
			axis.ShowTicks = *params.ShowTicks
		}
		if params.Direction != nil {
			if err := axis.SetTickDirection(*params.Direction); err != nil {
				return err
			}
		}
		if params.ShowLabels != nil {
			switch which {
			case "major":
				axis.ShowLabels = *params.ShowLabels
			case "minor":
				axis.ShowMinorLabels = *params.ShowLabels
			case "both":
				axis.ShowLabels = *params.ShowLabels
				axis.ShowMinorLabels = *params.ShowLabels
			}
		}
		if params.Length != nil {
			switch which {
			case "major":
				axis.TickSize = *params.Length
			case "minor":
				axis.MinorTickSize = *params.Length
			case "both":
				axis.TickSize = *params.Length
				axis.MinorTickSize = *params.Length
			}
		}
		switch which {
		case "major":
			applyTickLabelParams(&axis.MajorLabelStyle, params)
		case "minor":
			applyTickLabelParams(&axis.MinorLabelStyle, params)
		case "both":
			applyTickLabelParams(&axis.MajorLabelStyle, params)
			applyTickLabelParams(&axis.MinorLabelStyle, params)
		}
	}
	a.applyTickSideParams(params)
	a.applyTickGridParams(params, which)
	return nil
}

func resetAxisTickParams(axis *Axis) {
	if axis == nil {
		return
	}
	var defaults *Axis
	switch axis.Side {
	case AxisLeft, AxisRight:
		defaults = NewYAxis()
	default:
		defaults = NewXAxis()
	}
	axis.TickColor = nil
	axis.TickLabelColor = nil
	axis.TickLineCap = defaults.TickLineCap
	axis.TickLineJoin = defaults.TickLineJoin
	axis.TickLineWidth = defaults.TickLineWidth
	axis.MinorTickLineWidth = defaults.MinorTickLineWidth
	axis.TickSize = defaults.TickSize
	axis.MinorTickSize = defaults.MinorTickSize
	axis.MajorTickCount = defaults.MajorTickCount
	axis.MinorTickCount = defaults.MinorTickCount
	axis.TickDirection = defaults.TickDirection
	axis.ShowTicks = defaults.ShowTicks
	axis.ShowLabels = defaults.ShowLabels
	axis.ShowMinorLabels = defaults.ShowMinorLabels
	axis.MajorLabelStyle = defaults.MajorLabelStyle
	axis.MinorLabelStyle = defaults.MinorLabelStyle
}

func (a *Axes) applyTickSideParams(params TickParams) {
	if a == nil {
		return
	}
	if axisAllowsX(params.Axis) {
		if params.Bottom != nil {
			if *params.Bottom && a.XAxis == nil {
				a.XAxis = NewXAxis()
			}
			if a.XAxis != nil {
				a.XAxis.ShowTicks = *params.Bottom
			}
		}
		if params.Top != nil {
			if *params.Top {
				a.TopAxis().ShowTicks = true
			} else if a.XAxisTop != nil {
				a.XAxisTop.ShowTicks = false
			}
		}
		if params.LabelBottom != nil {
			if *params.LabelBottom && a.XAxis == nil {
				a.XAxis = NewXAxis()
			}
			if a.XAxis != nil {
				a.XAxis.ShowLabels = *params.LabelBottom
			}
		}
		if params.LabelTop != nil {
			if *params.LabelTop {
				a.TopAxis().ShowLabels = true
			} else if a.XAxisTop != nil {
				a.XAxisTop.ShowLabels = false
			}
		}
	}
	if axisAllowsY(params.Axis) {
		if params.Left != nil {
			if *params.Left && a.YAxis == nil {
				a.YAxis = NewYAxis()
			}
			if a.YAxis != nil {
				a.YAxis.ShowTicks = *params.Left
			}
		}
		if params.Right != nil {
			if *params.Right {
				a.RightAxis().ShowTicks = true
			} else if a.YAxisRight != nil {
				a.YAxisRight.ShowTicks = false
			}
		}
		if params.LabelLeft != nil {
			if *params.LabelLeft && a.YAxis == nil {
				a.YAxis = NewYAxis()
			}
			if a.YAxis != nil {
				a.YAxis.ShowLabels = *params.LabelLeft
			}
		}
		if params.LabelRight != nil {
			if *params.LabelRight {
				a.RightAxis().ShowLabels = true
			} else if a.YAxisRight != nil {
				a.YAxisRight.ShowLabels = false
			}
		}
	}
}

func (a *Axes) applyTickGridParams(params TickParams, which string) {
	if a == nil || (params.GridVisible == nil && params.GridColor == nil && params.GridAlpha == nil && params.GridLineWidth == nil && params.GridDashes == nil) {
		return
	}
	for _, artist := range a.Artists {
		grid, ok := artist.(*Grid)
		if !ok || !gridMatchesAxisSpec(grid, params.Axis) {
			continue
		}
		if params.GridVisible != nil {
			switch which {
			case "major":
				grid.Major = *params.GridVisible
			case "minor":
				grid.Minor = *params.GridVisible
			case "both":
				grid.Major = *params.GridVisible
				grid.Minor = *params.GridVisible
			}
		}
		if params.GridColor != nil {
			switch which {
			case "major":
				grid.Color = *params.GridColor
			case "minor":
				grid.MinorColor = *params.GridColor
			case "both":
				grid.Color = *params.GridColor
				grid.MinorColor = *params.GridColor
			}
		}
		if params.GridAlpha != nil {
			alpha := *params.GridAlpha
			if alpha < 0 {
				alpha = 0
			}
			if alpha > 1 {
				alpha = 1
			}
			switch which {
			case "major":
				grid.Alpha = alpha
			case "minor":
				grid.MinorColor.A = alpha
			case "both":
				grid.Alpha = alpha
				grid.MinorColor.A = alpha
			}
		}
		if params.GridLineWidth != nil {
			switch which {
			case "major":
				grid.LineWidth = *params.GridLineWidth
			case "minor":
				grid.MinorLineWidth = *params.GridLineWidth
			case "both":
				grid.LineWidth = *params.GridLineWidth
				grid.MinorLineWidth = *params.GridLineWidth
			}
		}
		if params.GridDashes != nil {
			dashes := styleCloneDashes(params.GridDashes)
			switch which {
			case "major":
				grid.Dashes = dashes
			case "minor":
				grid.MinorDashes = dashes
			case "both":
				grid.Dashes = styleCloneDashes(dashes)
				grid.MinorDashes = styleCloneDashes(dashes)
			}
		}
	}
}

func gridMatchesAxisSpec(grid *Grid, spec string) bool {
	if grid == nil {
		return false
	}
	switch normalizeAxisSpec(spec) {
	case "both":
		return true
	case "x", "bottom", "top":
		return grid.Axis == AxisBottom || grid.Axis == AxisTop
	case "y", "left", "right":
		return grid.Axis == AxisLeft || grid.Axis == AxisRight
	default:
		return false
	}
}

// SetAxisLineStyle applies cap/join/dash styling to the selected axes.
func (a *Axes) SetAxisLineStyle(spec string, lineCap render.LineCap, join render.LineJoin, dashes ...float64) error {
	if err := validateAxisSpec(spec); err != nil {
		return err
	}
	for _, axis := range a.axesForSpec(spec) {
		if axis == nil {
			continue
		}
		axis.SetLineStyle(lineCap, join, dashes...)
	}
	return nil
}

// TwinX creates an overlay axes sharing the x-scale with an independent y-scale on the right.
func (a *Axes) TwinX() *Axes {
	twin := a.newOverlayAxes()
	if twin == nil {
		return nil
	}
	twin.shareX = a.xScaleRoot()
	if twin.XAxis != nil {
		twin.XAxis.ShowSpine = false
		twin.XAxis.ShowTicks = false
		twin.XAxis.ShowLabels = false
	}
	if twin.YAxis != nil {
		twin.YAxis.ShowSpine = false
		twin.YAxis.ShowTicks = false
		twin.YAxis.ShowLabels = false
	}
	twin.ShowFrame = false
	twin.RightAxis()
	return twin
}

// TwinY creates an overlay axes sharing the y-scale with an independent x-scale on the top.
func (a *Axes) TwinY() *Axes {
	twin := a.newOverlayAxes()
	if twin == nil {
		return nil
	}
	twin.shareY = a.yScaleRoot()
	if twin.YAxis != nil {
		twin.YAxis.ShowSpine = false
		twin.YAxis.ShowTicks = false
		twin.YAxis.ShowLabels = false
	}
	if twin.XAxis != nil {
		twin.XAxis.ShowSpine = false
		twin.XAxis.ShowTicks = false
		twin.XAxis.ShowLabels = false
	}
	twin.ShowFrame = false
	twin.TopAxis()
	return twin
}

// SecondaryXAxis creates an overlay axes that displays transformed x ticks on the requested side.
func (a *Axes) SecondaryXAxis(side AxisSide, forward func(float64) float64, inverse func(float64) (float64, bool)) (*Axes, error) {
	if side != AxisTop && side != AxisBottom {
		return nil, fmt.Errorf("secondary x-axis side must be top or bottom")
	}
	return a.newSecondaryAxes(true, side, forward, inverse)
}

// SecondaryYAxis creates an overlay axes that displays transformed y ticks on the requested side.
func (a *Axes) SecondaryYAxis(side AxisSide, forward func(float64) float64, inverse func(float64) (float64, bool)) (*Axes, error) {
	if side != AxisLeft && side != AxisRight {
		return nil, fmt.Errorf("secondary y-axis side must be left or right")
	}
	return a.newSecondaryAxes(false, side, forward, inverse)
}

// layout computes the pixel rectangle for this Axes inside the Figure.
func (a *Axes) layout(f *Figure) (pixelRect geom.Rect) {
	// Figure fractions use a bottom-left origin, but display pixels use a
	// top-left origin. Flip Y so subplot rows land in the expected order.
	minPt := geom.Pt{X: f.SizePx.X * a.RectFraction.Min.X, Y: f.SizePx.Y * (1 - a.RectFraction.Max.Y)}
	maxPt := geom.Pt{X: f.SizePx.X * a.RectFraction.Max.X, Y: f.SizePx.Y * (1 - a.RectFraction.Min.Y)}
	return geom.Rect{Min: minPt, Max: maxPt}
}

func (a *Axes) xScaleRoot() *Axes {
	if a == nil {
		return nil
	}
	cur := a
	for cur.shareX != nil {
		cur = cur.shareX
		if cur == nil {
			return a
		}
	}
	return cur
}

func (a *Axes) yScaleRoot() *Axes {
	if a == nil {
		return nil
	}
	cur := a
	for cur.shareY != nil {
		cur = cur.shareY
		if cur == nil {
			return a
		}
	}
	return cur
}

func (a *Axes) effectiveXAxis() *Axis {
	return a.XAxis
}

func (a *Axes) effectiveYAxis() *Axis {
	return a.YAxis
}

func (a *Axes) effectiveTopAxis() *Axis {
	if a == nil {
		return nil
	}
	return a.XAxisTop
}

func (a *Axes) effectiveRightAxis() *Axis {
	if a == nil {
		return nil
	}
	return a.YAxisRight
}

func (a *Axes) effectiveXScale() transform.Scale {
	if a.shareX != nil {
		return a.shareX.effectiveXScale()
	}
	return a.XScale
}

func (a *Axes) effectiveYScale() transform.Scale {
	if a.shareY != nil {
		return a.shareY.effectiveYScale()
	}
	return a.YScale
}

func (a *Axes) setScale(isX bool, name string, opts ...transform.ScaleOption) error {
	var target *Axes
	var current transform.Scale
	var primary, secondary *Axis
	var units *axisUnitsState

	if isX {
		target = a.xScaleRoot()
		current = target.XScale
		primary = target.XAxis
		secondary = target.XAxisTop
		units = target.xUnits
	} else {
		target = a.yScaleRoot()
		current = target.YScale
		primary = target.YAxis
		secondary = target.YAxisRight
		units = target.yUnits
	}

	if units != nil && !units.scaleCompatible(name) {
		return fmt.Errorf("%s units require a linear axis scale", units.name())
	}

	minVal, maxVal := currentScaleDomain(current)
	cfg := transform.ResolveScaleOptions(append([]transform.ScaleOption{
		transform.WithScaleDomain(minVal, maxVal),
	}, opts...)...)
	scale, err := transform.NewScaleWithOptions(name, cfg)
	if err != nil {
		return err
	}

	if isX {
		target.XScale = scale
	} else {
		target.YScale = scale
	}
	configureScaleAxes(primary, secondary, name, cfg)
	configureChildScaleAxes(target, isX, name, cfg)
	target.refreshUnitAxis(isX)
	return nil
}

// effectiveRC resolves the RC for this axes, inheriting from the Figure if needed.
func (a *Axes) effectiveRC(f *Figure) style.RC {
	if a.RC != nil {
		return *a.RC
	}
	if a.figure != nil {
		return a.figure.RC
	}
	if f == nil {
		return style.CurrentDefaults()
	}
	return f.RC
}

func (a *Axes) resolvedRC() style.RC {
	if a == nil {
		return style.CurrentDefaults()
	}
	return a.effectiveRC(a.figure)
}

func currentScaleDomain(s transform.Scale) (float64, float64) {
	if s == nil {
		return 0, 1
	}
	return s.Domain()
}

func replaceScaleDomain(s transform.Scale, minVal, maxVal float64) transform.Scale {
	switch v := s.(type) {
	case nil:
		minVal, maxVal = nonsingularLinearDomain(minVal, maxVal)
		return transform.NewLinear(minVal, maxVal)
	case transform.Linear:
		minVal, maxVal = nonsingularLinearDomain(minVal, maxVal)
		return v.WithDomain(minVal, maxVal)
	case transform.DomainSetter:
		return v.WithDomain(minVal, maxVal)
	case invertedScale:
		return replaceScaleDomain(v.base, minVal, maxVal)
	default:
		minVal, maxVal = nonsingularLinearDomain(minVal, maxVal)
		return transform.NewLinear(minVal, maxVal)
	}
}

func nonsingularLinearDomain(minVal, maxVal float64) (float64, float64) {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) || math.IsInf(minVal, 0) || math.IsInf(maxVal, 0) {
		return minVal, maxVal
	}
	if minVal != maxVal {
		return minVal, maxVal
	}
	expand := math.Abs(minVal) * 0.05
	if expand == 0 {
		expand = 0.05
	}
	return minVal - expand, maxVal + expand
}

func configureScaleAxes(primary, secondary *Axis, scaleName string, cfg transform.ScaleOptions) {
	configureScaleAxis(primary, scaleName, cfg)
	configureScaleAxis(secondary, scaleName, cfg)
}

func configureChildScaleAxes(root *Axes, isX bool, scaleName string, cfg transform.ScaleOptions) {
	if root == nil {
		return
	}
	for _, child := range root.childAxes {
		if child == nil {
			continue
		}
		if isX {
			if child.shareX == root || childLinkedSecondaryScale(child.XScale, root, true) {
				configureScaleAxes(child.XAxis, child.XAxisTop, scaleName, cfg)
			}
			continue
		}
		if child.shareY == root || childLinkedSecondaryScale(child.YScale, root, false) {
			configureScaleAxes(child.YAxis, child.YAxisRight, scaleName, cfg)
		}
	}
}

func childLinkedSecondaryScale(scale transform.Scale, root *Axes, isX bool) bool {
	linked, ok := scale.(linkedSecondaryScale)
	if !ok || linked.parent == nil || linked.isX != isX {
		return false
	}
	if isX {
		return linked.parent.xScaleRoot() == root
	}
	return linked.parent.yScaleRoot() == root
}

func configureScaleAxis(axis *Axis, scaleName string, cfg transform.ScaleOptions) {
	if axis == nil {
		return
	}

	switch strings.ToLower(scaleName) {
	case "log", "functionlog":
		axis.Locator = LogLocator{Base: cfg.Base, Minor: false}
		axis.Formatter = LogFormatterMathText{Base: cfg.Base, SciNotation: true}
		if len(cfg.Subs) > 0 {
			axis.MinorLocator = LogLocator{Base: cfg.Base, Minor: true, Subs: cfg.Subs}
		} else {
			axis.MinorLocator = nil
		}
	case "symlog":
		axis.Locator = SymLogLocator{Base: cfg.Base, LinThresh: cfg.LinThresh}
		axis.Formatter = LogFormatterMathText{Base: cfg.Base, SciNotation: true}
		axis.MinorLocator = SymLogLocator{Base: cfg.Base, LinThresh: cfg.LinThresh, Subs: cfg.Subs}
	case "asinh":
		axis.Locator = AsinhLocator{LinearWidth: cfg.LinearWidth, Base: cfg.Base}
		if cfg.Base > 1 {
			axis.Formatter = LogFormatterMathText{Base: cfg.Base, SciNotation: true}
		} else {
			axis.Formatter = StrMethodFormatter{Template: "{x:.3g}"}
		}
		axis.MinorLocator = AsinhLocator{LinearWidth: cfg.LinearWidth, Base: cfg.Base, Subs: cfg.Subs}
	case "logit":
		axis.Locator = LogitLocator{}
		axis.Formatter = LogitFormatter{}
		axis.MinorLocator = LogitLocator{Minor: true}
		axis.MinorFormatter = LogitFormatter{Minor: true}
	default:
		axis.Locator = LinearLocator{}
		axis.Formatter = ScalarFormatter{Prec: 3}
		axis.MinorLocator = nil
	}
}

func (a *Axes) applyStyleDefaults(rc style.RC) {
	if a == nil {
		return
	}
	if a.XAxis != nil {
		a.XAxis.Color = rc.XTickColor
		a.XAxis.LineWidth = rc.AxisLineWidth
	}
	if a.YAxis != nil {
		a.YAxis.Color = rc.YTickColor
		a.YAxis.LineWidth = rc.AxisLineWidth
	}
	if a.XAxisTop != nil {
		a.XAxisTop.Color = rc.XTickColor
		a.XAxisTop.LineWidth = rc.AxisLineWidth
	}
	if a.YAxisRight != nil {
		a.YAxisRight.Color = rc.YTickColor
		a.YAxisRight.LineWidth = rc.AxisLineWidth
	}
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewColorCycle(rc.Palette())
	}
	if a.PatchColorCycle == nil {
		a.PatchColorCycle = color.NewColorCycle(rc.Palette())
	}
}

func (a *Axes) addDefaultGrids(rc style.RC) {
	if a == nil || !rc.GridVisible {
		return
	}

	addGrid := func(axis AxisSide) {
		grid := a.AddGrid(axis)
		grid.Dashes = styleCloneDashes(rc.GridDashes)
		grid.MinorDashes = styleCloneDashes(rc.MinorGridDashes)
		switch strings.ToLower(strings.TrimSpace(rc.GridWhich)) {
		case "minor":
			grid.Major = false
			grid.Minor = true
		case "both":
			grid.Major = true
			grid.Minor = true
		default:
			grid.Major = true
			grid.Minor = false
		}
	}

	switch strings.ToLower(strings.TrimSpace(rc.GridAxis)) {
	case "x":
		addGrid(AxisBottom)
	case "y":
		addGrid(AxisLeft)
	default:
		addGrid(AxisBottom)
		addGrid(AxisLeft)
	}
}

func styleCloneDashes(dashes []float64) []float64 {
	if len(dashes) == 0 {
		return nil
	}
	cloned := make([]float64, len(dashes))
	copy(cloned, dashes)
	return cloned
}

// DrawFigure performs a traversal and draws the figure into the renderer
// using the zero DrawOptions. Animated artists are skipped, matching
// Matplotlib's default per-frame draw behavior.
func DrawFigure(fig *Figure, r render.Renderer) {
	DrawFigureWithOptions(fig, r, DrawOptions{})
}

// DrawFigureWithOptions performs a traversal and draws the figure into the
// renderer using the supplied DrawOptions. The animation engine flips the
// AnimatedFilter to drive background and overlay passes through this entry
// point.
func DrawFigureWithOptions(fig *Figure, r render.Renderer, opts DrawOptions) {
	vp := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: fig.SizePx.X, Y: fig.SizePx.Y}}
	_ = r.Begin(vp)
	defer r.End()
	setRendererResolution(r, fig.RC.DPI)

	prepareFigureLayout(fig, r, vp)
	syncAxesLocators(fig, r)
	syncColorbarAxes(fig)
	syncAxesLocators(fig, r)
	alignment := computeFigureTextAlignment(fig, r, vp)

	for _, ax := range fig.Children {
		px := ax.adjustedLayout(fig)
		xAxis := ax.effectiveXAxis()
		yAxis := ax.effectiveYAxis()
		topAxis := ax.effectiveTopAxis()
		rightAxis := ax.effectiveRightAxis()

		// Build DrawContext with composed transform
		ctx := newAxesDrawContext(ax, fig, vp, px)
		ctx.DrawOptions = opts
		setRendererResolution(r, ctx.RC.DPI)

		// In an animated-overlay pass we only redraw the animated artists on
		// top of a previously captured background, so skip backgrounds,
		// chrome, axis ticks, spines, and labels.
		if opts.AnimatedFilter == AnimatedFilterOnlyAnimated {
			r.Save()
			r.ClipRect(px)
			if framePath, ok := projectionFramePath(ctx.Projection, px); ok {
				r.ClipPath(framePath)
			}
			for _, art := range ax.Artists {
				drawArtist(r, ctx, art)
			}
			for _, art := range ax.WidgetArtists {
				drawArtist(r, ctx, art)
			}
			r.Restore()
			for _, art := range ax.Artists {
				if overlay, ok := art.(OverlayArtist); ok {
					drawOverlayArtist(r, ctx, art, overlay)
				}
			}
			for _, art := range ax.WidgetArtists {
				if overlay, ok := art.(OverlayArtist); ok {
					drawOverlayArtist(r, ctx, art, overlay)
				}
			}
			continue
		}

		if ctx.RC.AxesBackground != fig.RC.FigureBackground() {
			backgroundPath := pixelRectPath(px)
			if framePath, ok := projectionFramePath(ctx.Projection, px); ok {
				backgroundPath = framePath
			}
			r.Path(backgroundPath, &render.Paint{
				Fill: ctx.RC.AxesBackground,
			})
		}

		// Draw only clipped content (data and grids) while the axes clip is active.
		r.Save()
		r.ClipRect(px)
		if framePath, ok := projectionFramePath(ctx.Projection, px); ok {
			r.ClipPath(framePath)
		}

		if !ax.zsorted {
			sortArtists(ax.Artists)
			ax.zsorted = true
		}
		for _, art := range ax.Artists {
			drawArtist(r, ctx, art)
		}
		if !ax.widgetZsorted {
			sortArtists(ax.WidgetArtists)
			ax.widgetZsorted = true
		}
		for _, art := range ax.WidgetArtists {
			drawArtist(r, ctx, art)
		}

		r.Restore()

		for _, art := range ax.Artists {
			if overlay, ok := art.(OverlayArtist); ok {
				drawOverlayArtist(r, ctx, art, overlay)
			}
		}
		for _, art := range ax.WidgetArtists {
			if overlay, ok := art.(OverlayArtist); ok {
				drawOverlayArtist(r, ctx, art, overlay)
			}
		}

		// Matplotlib draws Axis objects (ticks and tick labels, zorder 1.5)
		// before Spine artists (zorder 2.5). This matters at endpoint ticks:
		// the spine overpaints the tick cap by a single coverage level.
		if xAxis != nil {
			xAxis.DrawTicks(r, ctx)
			xAxis.DrawTickLabels(r, ctx)
		}
		if yAxis != nil {
			yAxis.DrawTicks(r, ctx)
			yAxis.DrawTickLabels(r, ctx)
		}
		if topAxis != nil {
			topAxis.DrawTicks(r, ctx)
			topAxis.DrawTickLabels(r, ctx)
		}
		if rightAxis != nil {
			rightAxis.DrawTicks(r, ctx)
			rightAxis.DrawTickLabels(r, ctx)
		}
		for _, extraAxis := range ax.ExtraAxes {
			if extraAxis != nil {
				extraAxis.DrawTicks(r, ctx)
				extraAxis.DrawTickLabels(r, ctx)
			}
		}

		// Draw spines outside the clip so they can straddle the axes edge the
		// same way Matplotlib does.
		if xAxis != nil {
			xAxis.Draw(r, ctx)
		}
		if yAxis != nil {
			yAxis.Draw(r, ctx)
		}
		if topAxis != nil {
			topAxis.Draw(r, ctx)
		}
		if rightAxis != nil {
			rightAxis.Draw(r, ctx)
		}
		for _, extraAxis := range ax.ExtraAxes {
			if extraAxis != nil {
				extraAxis.Draw(r, ctx)
			}
		}
		if ax.ShowFrame {
			ref := xAxis
			if ref == nil {
				ref = topAxis
			}
			if ref == nil {
				ref = yAxis
			}
			if ref == nil {
				ref = rightAxis
			}
			DrawFrame(r, ctx, ref, topAxis == nil, rightAxis == nil)
		}

		// Draw axes text labels outside the clip rect.
		drawAxesLabels(ax, r, ctx, px, alignment)
	}

	setRendererResolution(r, fig.RC.DPI)
	drawFigureArtistsWithOptions(fig, r, vp, opts)
	if opts.AnimatedFilter != AnimatedFilterOnlyAnimated {
		drawFigureLabels(fig, r, vp)
	}
}

func setRendererResolution(r render.Renderer, dpi float64) {
	if dpi <= 0 {
		return
	}
	if setter, ok := r.(render.DPIAware); ok {
		setter.SetResolution(uint(math.Round(dpi)))
	}
}

// drawAxesLabels renders title, xlabel, and ylabel outside the clipped axes area.
func drawAxesLabels(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, alignment figureTextAlignment) {
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	titleColor := ctx.RC.DefaultAxesTitleColor()
	labelColor := ctx.RC.DefaultAxesLabelColor()
	titleSize := titleFontSize(ctx)
	labelSize := axisLabelFontSize(ctx)

	// Title: centered above the plot
	if ax.Title != "" {
		layout := measureSingleLineTextLayout(r, ax.Title, titleSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		drawDisplayText(
			textRen,
			ax.Title,
			alignedSingleLineOrigin(titleAnchorPoint(ax, r, ctx, px, alignment), layout, TextAlignCenter, textLayoutVAlignBaseline),
			titleSize,
			titleColor,
			ctx.RC.FontKey,
			ctx.RC.UseTeX,
		)
	}

	// XLabel: centered relative to the selected axis side and padded from the
	// union of the spine and visible tick-label bounds, matching Matplotlib's
	// default label placement model.
	if ax.XLabel != "" && ax.ProjectionName() != "3d" {
		side := ax.effectiveXLabelSide()
		layout := measureSingleLineTextLayout(r, ax.XLabel, labelSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		anchor, vAlign := xLabelAnchorPoint(ax, r, ctx, px, side, alignment)
		drawDisplayText(
			textRen,
			ax.XLabel,
			alignedSingleLineOrigin(anchor, layout, TextAlignCenter, vAlign),
			labelSize,
			labelColor,
			ctx.RC.FontKey,
			ctx.RC.UseTeX,
		)
	}

	// YLabel: vertical text if supported, else horizontal fallback
	if ax.YLabel != "" && ax.ProjectionName() != "3d" {
		side := ax.effectiveYLabelSide()
		anchor := yLabelAnchorPoint(ax, r, ctx, px, side, alignment)
		if side == AxisLeft {
			anchor.X -= ctx.RC.AxisLineWidth
			anchor.Y += ctx.RC.AxisLineWidth
		}
		angle := math.Pi / 2
		if side == AxisRight {
			angle = -math.Pi / 2
		}
		switch ren := r.(type) {
		case render.RotatedTextDrawer:
			drawDisplayTextRotated(ren, ax.YLabel, anchor, labelSize, angle, labelColor, ctx.RC.FontKey, ctx.RC.UseTeX)
		case render.VerticalTextDrawer:
			if angle < 0 {
				layout := measureSingleLineTextLayout(r, ax.YLabel, labelSize, ctx.RC.FontKey, ctx.RC.UseTeX)
				drawDisplayText(
					textRen,
					ax.YLabel,
					alignedSingleLineOrigin(geom.Pt{X: anchor.X, Y: px.Min.Y + px.H()/2}, layout, TextAlignCenter, textLayoutVAlignCenter),
					labelSize,
					labelColor,
					ctx.RC.FontKey,
					ctx.RC.UseTeX,
				)
			} else {
				drawDisplayTextVertical(ren, ax.YLabel, geom.Pt{X: anchor.X, Y: anchor.Y}, labelSize, labelColor, ctx.RC.FontKey)
			}
		default:
			layout := measureSingleLineTextLayout(r, ax.YLabel, labelSize, ctx.RC.FontKey, ctx.RC.UseTeX)
			drawDisplayText(
				textRen,
				ax.YLabel,
				alignedSingleLineOrigin(geom.Pt{X: anchor.X, Y: px.Min.Y + px.H()/2}, layout, TextAlignCenter, textLayoutVAlignCenter),
				labelSize,
				labelColor,
				ctx.RC.FontKey,
				ctx.RC.UseTeX,
			)
		}
	}
}

func titleFontSize(ctx *DrawContext) float64 {
	if ctx == nil {
		return 12
	}
	return ctx.RC.TitleSize()
}

func axisLabelFontSize(ctx *DrawContext) float64 {
	if ctx == nil {
		return 8
	}
	return ctx.RC.AxisLabelSize()
}

func figureLabelFontSize(ctx *DrawContext) float64 {
	if ctx == nil {
		return 12
	}
	return ctx.RC.TitleSize()
}

func titleAnchorPoint(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, alignment figureTextAlignment) geom.Pt {
	titlePadPx := pointsToPixels(ctx.RC, 6)
	topExtent := titleTopExtent(ax, r, ctx, px)
	if aligned, ok := alignment.titleExtents[alignmentKey(AxisTop, spinePixelY(AxisTop, px))]; ok {
		topExtent = aligned
	}
	return geom.Pt{
		X: ctx.TransAxes().Apply(geom.Pt{X: 0.5, Y: 1}).X,
		Y: topExtent - titlePadPx,
	}
}

func xLabelAnchorPoint(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, side AxisSide, alignment figureTextAlignment) (geom.Pt, textLayoutVerticalAlign) {
	anchor := ctx.TransAxes().Apply(geom.Pt{X: 0.5, Y: 0})

	xAxis := ax.axisForXLabelSide(side)
	if side == AxisTop {
		topExtent := spinePixelY(AxisTop, px)
		if xAxis != nil {
			if tickBounds, ok := axisTickLabelBounds(xAxis, r, ctx); ok {
				topExtent = math.Min(topExtent, tickBounds.Min.Y)
			}
		}
		if aligned, ok := alignment.xLabelExtents[alignmentKey(side, spinePixelY(side, px))]; ok {
			topExtent = aligned
		}
		anchor.Y = topExtent - axisLabelPadPx(ctx)
		return anchor, textLayoutVAlignBaseline
	}

	bottomExtent := spinePixelY(AxisBottom, px)
	if xAxis != nil {
		if tickBounds, ok := axisTickLabelBounds(xAxis, r, ctx); ok {
			bottomExtent = math.Max(bottomExtent, tickBounds.Max.Y)
		}
	}
	if aligned, ok := alignment.xLabelExtents[alignmentKey(side, spinePixelY(side, px))]; ok {
		bottomExtent = aligned
	}
	anchor.Y = bottomExtent + axisLabelPadPx(ctx)
	return anchor, textLayoutVAlignTop
}

func yLabelAnchorPoint(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, side AxisSide, alignment figureTextAlignment) geom.Pt {
	anchor := ctx.TransAxes().Apply(geom.Pt{X: 0, Y: 0.5})
	anchor.X = spinePixelX(AxisLeft, px) - axisLabelPadPx(ctx)

	yAxis := ax.axisForYLabelSide(side)
	if side == AxisRight {
		spineX := spinePixelX(AxisRight, px)
		rightExtent := spineX
		if tickBounds, ok := axisTickLabelBounds(yAxis, r, ctx); ok {
			rightExtent = math.Max(rightExtent, tickBounds.Max.X)
		}
		if aligned, ok := alignment.yLabelExtents[alignmentKey(side, spinePixelX(side, px))]; ok {
			rightExtent = aligned
		}
		anchor.X = rightExtent + axisLabelPadPx(ctx)
		return anchor
	}

	spineX := spinePixelX(AxisLeft, px)
	leftExtent := spineX
	if tickBounds, ok := axisTickLabelBounds(yAxis, r, ctx); ok {
		leftExtent = math.Min(leftExtent, tickBounds.Min.X)
	}
	if aligned, ok := alignment.yLabelExtents[alignmentKey(side, spinePixelX(side, px))]; ok {
		leftExtent = aligned
	}
	anchor.X = leftExtent - axisLabelPadPx(ctx)
	return anchor
}

func axisLabelPadPx(ctx *DrawContext) float64 {
	if ctx == nil {
		return pointsToPixels(style.CurrentDefaults(), 4)
	}
	return pointsToPixels(ctx.RC, 4)
}

func (a *Axes) effectiveXLabelSide() AxisSide {
	if a == nil {
		return AxisBottom
	}
	if a.xLabelSide == AxisTop {
		return AxisTop
	}
	return AxisBottom
}

func (a *Axes) effectiveYLabelSide() AxisSide {
	if a == nil {
		return AxisLeft
	}
	if a.yLabelSide == AxisRight {
		return AxisRight
	}
	return AxisLeft
}

func (a *Axes) axisForXLabelSide(side AxisSide) *Axis {
	if a == nil {
		return nil
	}
	if side == AxisTop {
		return a.XAxisTop
	}
	if a.XAxis != nil {
		return a.XAxis
	}
	return a.XAxisTop
}

func (a *Axes) axisForYLabelSide(side AxisSide) *Axis {
	if a == nil {
		return nil
	}
	if side == AxisRight {
		if a.YAxisRight != nil {
			return a.YAxisRight
		}
		return nil
	}
	if a.YAxis != nil {
		return a.YAxis
	}
	return a.YAxisRight
}

func spinePixelX(side AxisSide, px geom.Rect) float64 {
	p1, _ := spinePixelEndpoints(side, px)
	return p1.X
}

func spinePixelY(side AxisSide, px geom.Rect) float64 {
	p1, _ := spinePixelEndpoints(side, px)
	return p1.Y
}

func (a *Axes) adjustedLayout(f *Figure) geom.Rect {
	px := a.layout(f)
	if a.colorbarParent != nil {
		return a.adjustedColorbarLayout(f, px)
	}
	target := 0.0
	if a.boxAspect > 0 {
		target = a.boxAspect
	} else {
		switch a.aspectMode {
		case "equal":
			target = a.dataAspectTarget(1)
		case "ratio":
			target = a.dataAspectTarget(a.aspectValue)
		}
	}
	if target <= 0 || math.IsNaN(target) || math.IsInf(target, 0) {
		return px
	}
	return rectWithAspect(px, target)
}

func (a *Axes) adjustedColorbarLayout(f *Figure, px geom.Rect) geom.Rect {
	if a == nil || f == nil || f.SizePx.X <= 0 || f.SizePx.Y <= 0 {
		return px
	}
	if colorbarIsHorizontal(a.colorbarLocation) {
		return px
	}
	width := a.colorbarWidth
	if width > 0 {
		width *= f.SizePx.X
	} else {
		aspect := resolvedColorbarAspect(a.colorbarAspect)
		if aspect > 0 {
			aspect *= colorbarExtensionShrink(a.colorbarExtend)
			width = px.H() / aspect
		}
	}
	if width <= 0 || width >= px.W() {
		return px
	}
	px.Max.X = px.Min.X + width
	return px
}

func (a *Axes) dataAspectTarget(aspect float64) float64 {
	if a == nil || aspect <= 0 {
		return 0
	}
	xMin, xMax := currentScaleDomain(a.effectiveXScale())
	yMin, yMax := currentScaleDomain(a.effectiveYScale())
	xSpan := math.Abs(xMax - xMin)
	ySpan := math.Abs(yMax - yMin)
	if xSpan == 0 || ySpan == 0 {
		return 0
	}
	return aspect * ySpan / xSpan
}

func rectWithAspect(r geom.Rect, target float64) geom.Rect {
	if target <= 0 {
		return r
	}
	cur := r.H() / r.W()
	switch {
	case cur > target:
		newH := r.W() * target
		pad := (r.H() - newH) / 2
		r.Min.Y += pad
		r.Max.Y -= pad
	case cur < target:
		newW := r.H() / target
		pad := (r.W() - newW) / 2
		r.Min.X += pad
		r.Max.X -= pad
	}
	return r
}

func cloneAxisForSide(src *Axis, side AxisSide) *Axis {
	var axis Axis
	if src != nil {
		axis = *src
	} else {
		switch side {
		case AxisBottom, AxisTop:
			axis = *NewXAxis()
		case AxisLeft, AxisRight:
			axis = *NewYAxis()
		}
	}
	axis.Side = side
	axis.ShowSpine = true
	axis.ShowTicks = true
	axis.ShowLabels = true
	return &axis
}

func (a *Axes) newOverlayAxes() *Axes {
	if a == nil || a.figure == nil {
		return nil
	}
	overlay := a.figure.addAxesWithProjection(a.RectFraction, cloneProjection(a.projection))
	overlay.RC = a.RC
	overlay.aspectMode = a.aspectMode
	overlay.aspectValue = a.aspectValue
	overlay.boxAspect = a.boxAspect
	a.childAxes = append(a.childAxes, overlay)
	return overlay
}

func (a *Axes) newSecondaryAxes(isX bool, side AxisSide, forward func(float64) float64, inverse func(float64) (float64, bool)) (*Axes, error) {
	if a == nil || a.figure == nil {
		return nil, fmt.Errorf("secondary axes require a figure-backed axes")
	}
	if forward == nil || inverse == nil {
		return nil, fmt.Errorf("secondary axes require forward and inverse functions")
	}
	overlay := a.newOverlayAxes()
	if overlay == nil {
		return nil, fmt.Errorf("could not create overlay axes")
	}
	overlay.ShowFrame = false

	if isX {
		overlay.XScale = linkedSecondaryScale{parent: a, isX: true, forward: forward, inverse: inverse}
		overlay.shareY = a.yScaleRoot()
		if overlay.YAxis != nil {
			overlay.YAxis.ShowSpine = false
			overlay.YAxis.ShowTicks = false
			overlay.YAxis.ShowLabels = false
		}
		if overlay.XAxis != nil {
			overlay.XAxis.ShowSpine = false
			overlay.XAxis.ShowTicks = false
			overlay.XAxis.ShowLabels = false
		}
		var secondaryAxis *Axis
		if side == AxisTop {
			secondaryAxis = overlay.TopAxis()
		} else {
			overlay.XAxis = cloneAxisForSide(a.XAxis, AxisBottom)
			secondaryAxis = overlay.XAxis
		}
		if secondaryAxis != nil {
			secondaryAxis.ShowSpine = false
		}
	} else {
		overlay.YScale = linkedSecondaryScale{parent: a, isX: false, forward: forward, inverse: inverse}
		overlay.shareX = a.xScaleRoot()
		if overlay.XAxis != nil {
			overlay.XAxis.ShowSpine = false
			overlay.XAxis.ShowTicks = false
			overlay.XAxis.ShowLabels = false
		}
		if overlay.YAxis != nil {
			overlay.YAxis.ShowSpine = false
			overlay.YAxis.ShowTicks = false
			overlay.YAxis.ShowLabels = false
		}
		var secondaryAxis *Axis
		if side == AxisRight {
			secondaryAxis = overlay.RightAxis()
		} else {
			overlay.YAxis = cloneAxisForSide(a.YAxis, AxisLeft)
			secondaryAxis = overlay.YAxis
		}
		if secondaryAxis != nil {
			secondaryAxis.ShowSpine = false
		}
	}
	return overlay, nil
}

type linkedSecondaryScale struct {
	parent  *Axes
	isX     bool
	forward func(float64) float64
	inverse func(float64) (float64, bool)
}

func (s linkedSecondaryScale) primaryScale() transform.Scale {
	if s.parent == nil {
		return nil
	}
	if s.isX {
		return s.parent.effectiveXScale()
	}
	return s.parent.effectiveYScale()
}

func (s linkedSecondaryScale) Domain() (float64, float64) {
	base := s.primaryScale()
	if base == nil || s.forward == nil {
		return 0, 1
	}
	minVal, maxVal := base.Domain()
	return s.forward(minVal), s.forward(maxVal)
}

func (s linkedSecondaryScale) Fwd(x float64) float64 {
	base := s.primaryScale()
	if base == nil || s.inverse == nil {
		return 0
	}
	primary, ok := s.inverse(x)
	if !ok {
		return math.NaN()
	}
	return base.Fwd(primary)
}

func (s linkedSecondaryScale) Inv(u float64) (float64, bool) {
	base := s.primaryScale()
	if base == nil || s.forward == nil {
		return 0, false
	}
	primary, ok := base.Inv(u)
	if !ok {
		return 0, false
	}
	return s.forward(primary), true
}

func validateAxisSpec(spec string) error {
	switch normalizeAxisSpec(spec) {
	case "", "both", "x", "y", "bottom", "top", "left", "right":
		return nil
	default:
		return fmt.Errorf("unsupported axis selection %q", spec)
	}
}

func normalizeAxisSpec(spec string) string {
	spec = strings.ToLower(strings.TrimSpace(spec))
	if spec == "" {
		return "both"
	}
	return spec
}

func axisAllowsX(spec string) bool {
	switch normalizeAxisSpec(spec) {
	case "both", "x", "bottom", "top":
		return true
	default:
		return false
	}
}

func axisAllowsY(spec string) bool {
	switch normalizeAxisSpec(spec) {
	case "both", "y", "left", "right":
		return true
	default:
		return false
	}
}

func normalizeTickWhich(which string) string {
	switch strings.ToLower(strings.TrimSpace(which)) {
	case "", "both":
		return "both"
	case "major":
		return "major"
	case "minor":
		return "minor"
	default:
		return ""
	}
}

func applyTickLabelParams(style *TickLabelStyle, params TickParams) {
	if style == nil {
		return
	}
	*style = normalizeTickLabelStyle(*style)
	if params.LabelRotation != nil {
		style.Rotation = *params.LabelRotation
	}
	if params.LabelSize != nil {
		style.FontSize = *params.LabelSize
	}
	if params.LabelPad != nil {
		style.Pad = *params.LabelPad
	}
	if params.LabelHAlign != nil {
		style.HAlign = *params.LabelHAlign
		style.AutoAlign = false
	}
	if params.LabelVAlign != nil {
		style.VAlign = *params.LabelVAlign
		style.AutoAlign = false
	}
}

func (a *Axes) axesForSpec(spec string) []*Axis {
	switch normalizeAxisSpec(spec) {
	case "x":
		return []*Axis{a.XAxis, a.XAxisTop}
	case "y":
		return []*Axis{a.YAxis, a.YAxisRight}
	case "bottom":
		return []*Axis{a.XAxis}
	case "top":
		return []*Axis{a.XAxisTop}
	case "left":
		return []*Axis{a.YAxis}
	case "right":
		return []*Axis{a.YAxisRight}
	default:
		return []*Axis{a.XAxis, a.XAxisTop, a.YAxis, a.YAxisRight}
	}
}

func enableMinorTicks(axis *Axis) {
	if axis == nil || axis.MinorLocator != nil {
		return
	}
	switch loc := axis.Locator.(type) {
	case LogLocator:
		axis.MinorLocator = LogLocator{Base: loc.Base, Minor: true, Subs: loc.Subs}
	case AutoLocator:
		axis.MinorLocator = AutoMinorLocator{Major: loc}
	case MaxNLocator:
		axis.MinorLocator = AutoMinorLocator{Major: loc}
	case MultipleLocator:
		axis.MinorLocator = AutoMinorLocator{Major: loc}
	default:
		axis.MinorLocator = MinorLinearLocator{}
	}
}

type invertedScale struct {
	base transform.Scale
}

func (s invertedScale) Fwd(x float64) float64 {
	return 1 - s.base.Fwd(x)
}

func (s invertedScale) Inv(u float64) (float64, bool) {
	return s.base.Inv(1 - u)
}

func (s invertedScale) Domain() (float64, float64) {
	maxVal, minVal := s.base.Domain()
	return minVal, maxVal
}

func toggleInvertedScale(s transform.Scale) transform.Scale {
	if s == nil {
		return nil
	}
	if inv, ok := s.(invertedScale); ok {
		return inv.base
	}
	return invertedScale{base: s}
}

func scaleDomainDescending(s transform.Scale) bool {
	if s == nil {
		return false
	}
	minVal, maxVal := s.Domain()
	return minVal > maxVal
}

package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

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
	XAxis        *Axis // bottom x-axis
	YAxis        *Axis // left y-axis
	XAxisTop     *Axis // optional top x-axis
	YAxisRight   *Axis // optional right y-axis
	ExtraAxes    []*Axis
	ShowFrame    bool // draw top and right border lines when no explicit top/right axis exists
	PatchVisible bool // draw the axes background patch

	// Text labels
	Title  string // title above the plot
	XLabel string // x-axis label below ticks
	YLabel string // y-axis label left of ticks

	// Color cycling for multiple series. Matplotlib keeps separate cycles for
	// line artists and shape/fill artists, so scatter markers do not advance
	// the line cycle.
	ColorCycle      *color.ColorCycle
	PatchColorCycle *color.ColorCycle

	aspectMode   string
	aspectValue  float64
	boxAspect    float64
	adjustable   string // "" => "box"; "datalim" shifts data limits instead of the display rect
	anchor       string // "" => "C"; cardinal anchor for aspect-shrunk axes
	axisBelowSet bool
	axisBelowZ   float64
	xLabelSide   AxisSide
	yLabelSide   AxisSide

	// Per-axes autoscale margins (Phase 7). nil => fall back to
	// defaultAutoScaleMargin (0.05); a non-nil value (including 0) is honored.
	xMargin       *float64
	yMargin       *float64
	autolimitMode string // "" => "data"; "round_numbers" snaps limits to round values

	// label_outer / shared-axes suppression (Phase 7). Hide redundant tick
	// labels (and offset text) / axis labels on inner shared subplots without
	// mutating the possibly-shared Axis artist.
	hideXTickLabels     bool
	hideYTickLabels     bool
	hideTopTickLabels   bool
	hideRightTickLabels bool
	hideXLabel          bool
	hideYLabel          bool

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

	colorbarParent        *Axes
	colorbarWidth         float64
	colorbarPadding       float64
	colorbarAspect        float64
	colorbarBase          geom.Rect
	colorbarExtend        string
	colorbarLocation      string
	colorbarTicks         []float64
	colorbarBounds        []float64
	colorbarMinorTicks    bool
	colorbarExtendFracMin float64
	colorbarExtendFracMax float64

	coordFormatter CoordFormatter

	// Persistent transform graph (Phase 11: live bbox-linked transforms).
	// axesBbox holds the axes pixel rectangle; transAxes maps the unit square
	// onto it via transform.BboxTransformTo. transData composes the data->axes
	// leg with transAxes and is cached across draws: resizing the axes (or
	// changing limits/scales) invalidates these nodes rather than rebuilding the
	// whole graph. Lazily initialized by ensureTransforms.
	axesBbox      *transform.Bbox
	transAxes     *transform.BboxTransformTo
	transData     *transform.CachedTransform
	dataNode      transform.TransformNode
	curDataToAxes transform.T
	dataAffine    geom.Affine
	dataAffineOK  bool
	dataSnapSet   bool
}

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

func (a *Axes) AddWidget(art Artist) {
	if a == nil {
		return
	}
	a.WidgetArtists = append(a.WidgetArtists, art)
	a.widgetZsorted = false
}

func (a *Axes) SetPosition(r geom.Rect) {
	if a == nil {
		return
	}
	a.RectFraction = r
	a.positionManual = true
}

func (a *Axes) ProjectionName() string {
	if a == nil || a.projection == nil {
		return "rectilinear"
	}
	return a.projection.Name()
}

// SetFrameOn mirrors matplotlib's Axes.set_frame_on: turning the frame off
// removes every spine from the draw list (axes/_base.py draw:
// `if not (self.axison and self._frameon): artists.remove(spine)`), in
// addition to the top/right frame edges.
func (a *Axes) SetFrameOn(on bool) {
	if a == nil {
		return
	}
	a.ShowFrame = on
	for _, axis := range []*Axis{a.XAxis, a.YAxis, a.XAxisTop, a.YAxisRight} {
		if axis != nil {
			axis.ShowSpine = on
		}
	}
}

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

// SetRadialLabelPosition mirrors matplotlib's set_rlabel_position: the angle
// is in theta-data degrees; the display position additionally follows the
// theta zero location and direction.
func (a *Axes) SetRadialLabelPosition(angleDeg float64) error {
	proj, ok := polarProjectionForAxes(a)
	if !ok {
		return fmt.Errorf("radial label position requires polar axes")
	}
	proj.radialLabelAngle = normalizePolarAngle(angleDeg * math.Pi / 180)
	return nil
}

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

func (a *Axes) NextColor() render.Color {
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	return a.ColorCycle.Next()
}

func (a *Axes) PeekColor() render.Color {
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	return a.ColorCycle.Peek()
}

// LineProps carries one step of the axes.prop_cycle for a line artist. Color is
// always present; the linestyle/marker/linewidth fields are set only when the
// active prop cycle carries that property, so a caller applies them solely where
// the user left the corresponding option unset. With a color-only prop cycle
// (the default), the Has* flags are false and behavior matches NextColor.
type LineProps struct {
	Color        render.Color
	LineStyle    string
	HasLineStyle bool
	Marker       string
	HasMarker    bool
	LineWidth    float64
	HasLineWidth bool
}

// NextLineProps advances the shared color cycle and returns its color together
// with any non-color properties (linestyle/marker/linewidth) carried by the
// axes.prop_cycle for the same step. It is the line-artist counterpart of
// NextColor; the two share the one ColorCycle index, so mixing them keeps the
// cycle in lock-step.
func (a *Axes) NextLineProps() LineProps {
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	idx := a.ColorCycle.Index()
	props := LineProps{Color: a.ColorCycle.Next()}
	a.fillCycleProps(&props, idx)
	return props
}

// fillCycleProps populates the non-color fields of props from the prop cycle
// entry at index idx, when an axes.prop_cycle with extra properties is active.
func (a *Axes) fillCycleProps(props *LineProps, idx int) {
	pc := a.resolvedRC().PropCycle
	if pc == nil || pc.Len() == 0 {
		return
	}
	row := pc.Row(idx)
	if v, ok := row["linestyle"].(string); ok {
		props.LineStyle, props.HasLineStyle = v, true
	}
	if v, ok := row["marker"].(string); ok {
		props.Marker, props.HasMarker = v, true
	}
	if v, ok := row["linewidth"].(float64); ok {
		props.LineWidth, props.HasLineWidth = v, true
	}
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

func (a *Axes) NextPatchColor() render.Color {
	if a.PatchColorCycle == nil {
		a.PatchColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	return a.PatchColorCycle.Next()
}

func (a *Axes) PeekPatchColor() render.Color {
	if a.PatchColorCycle == nil {
		a.PatchColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	return a.PatchColorCycle.Peek()
}

func (a *Axes) ResetColorCycle() {
	if a.ColorCycle != nil {
		a.ColorCycle.Reset()
	}
	if a.PatchColorCycle != nil {
		a.PatchColorCycle.Reset()
	}
}

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

func styleCloneDashes(dashes []float64) []float64 {
	if len(dashes) == 0 {
		return nil
	}
	cloned := make([]float64, len(dashes))
	copy(cloned, dashes)
	return cloned
}

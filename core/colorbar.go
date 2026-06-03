package core

import (
	"math"
	"strings"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

// ColorbarOptions configures figure-level colorbar placement.
type ColorbarOptions struct {
	Width       float64
	Padding     float64
	Aspect      float64
	Shrink      float64
	Anchor      *geom.Pt
	Label       string
	Colormap    *string
	VMin        *float64
	VMax        *float64
	Extend      string
	Location    string
	Orientation string
	Ticks       []float64
	Boundaries  []float64
	Values      []float64
	Spacing     string
	DrawEdges   bool
	ExtendRect  bool
}

// Colorbar renders a vertical gradient keyed to a scalar colormap.
type Colorbar struct {
	Mapping     ScalarMapInfo
	Mappable    ScalarMappable
	Colormap    string
	Extend      string
	Orientation string
	Boundaries  []float64
	Values      []float64
	Spacing     string
	DrawEdges   bool
	ExtendRect  bool
	Alpha       float64
	BorderColor render.Color
	BorderWidth float64
	z           float64
}

const (
	defaultColorbarFraction          = 0.15
	defaultColorbarPadding           = 0.05
	defaultHorizontalColorbarPadding = 0.15
	defaultColorbarAspect            = 20.0
)

// AddColorbar creates a dedicated axes to the right of a plot and populates it
// with a colorbar derived from a scalar-mappable artist.
func (f *Figure) AddColorbar(parent *Axes, mappable ScalarMappable, opts ...ColorbarOptions) *Axes {
	if f == nil || parent == nil || mappable == nil {
		return nil
	}

	cfg := ColorbarOptions{}
	if len(opts) > 0 {
		cfg = opts[0]
	}
	cfg.Aspect = resolvedColorbarAspect(cfg.Aspect)
	location := normalizeColorbarLocation(cfg.Location, cfg.Orientation)
	extend := normalizeColorbarExtend(cfg.Extend)

	mapping := mappable.ScalarMap().Resolved()
	cmapOverride := ""
	if cfg.Colormap != nil && *cfg.Colormap != "" {
		cmapOverride = *cfg.Colormap
		mapping.Colormap = cmapOverride
	}
	vmin := mapping.VMin
	if cfg.VMin != nil {
		vmin = *cfg.VMin
	}
	vmax := mapping.VMax
	if cfg.VMax != nil {
		vmax = *cfg.VMax
	}
	if vmin == vmax {
		vmax = vmin + 1
	}
	mapping.VMin = vmin
	mapping.VMax = vmax
	boundaries := colorbarOptionBoundaries(cfg.Values, cfg.Boundaries)
	if len(boundaries) >= 2 {
		if cfg.VMin == nil {
			mapping.VMin = boundaries[0]
		}
		if cfg.VMax == nil {
			mapping.VMax = boundaries[len(boundaries)-1]
		}
	}
	if _, ok := mapping.Norm.(Normalize); ok && mapping.Norm != nil {
		mapping.Norm = Normalize{VMin: mapping.VMin, VMax: mapping.VMax}
	}

	base := colorbarBaseRect(parent)
	thickness := resolvedColorbarThickness(f, base, cfg.Width, cfg.Aspect, location)
	slotThickness := resolvedColorbarSlotThickness(base, cfg.Width, location)
	padding := resolvedColorbarLayoutPadding(f, base, cfg.Padding, location)
	useResolvedSlot := colorbarUsesResolvedSlot(f, parent)
	if useResolvedSlot {
		padding = resolvedColorbarPadding(base, cfg.Padding, location)
		slotThickness = resolvedColorbarSlotThickness(base, cfg.Width, location)
	}
	parentRect, rect := colorbarPlacementRect(f, base, thickness, slotThickness, padding, location, useResolvedSlot)
	if !useResolvedSlot {
		parent.RectFraction = parentRect
	}
	rect = insetColorbarRectForExtensions(f, rect, extend, location)
	if rect.Min.X >= rect.Max.X {
		return nil
	}
	if rect.Min.Y >= rect.Max.Y {
		return nil
	}
	rect = applyColorbarShrinkAnchor(rect, cfg.Shrink, cfg.Anchor, location)

	ax := f.AddAxes(rect)
	ax.colorbarParent = parent
	ax.colorbarWidth = cfg.Width
	ax.colorbarPadding = cfg.Padding
	ax.colorbarAspect = cfg.Aspect
	ax.colorbarBase = base
	ax.colorbarExtend = extend
	ax.colorbarLocation = location
	ax.colorbarTicks = cloneFloat64s(cfg.Ticks)
	ax.colorbarBounds = cloneFloat64s(boundaries)
	ax.ShowFrame = false
	configureColorbarAxes(ax, location, cfg.Label)
	configureColorbarScale(ax, mapping, location, cfg.Ticks, boundaries, extend)

	ax.Add(&Colorbar{
		Mapping:     mapping,
		Mappable:    mappable,
		Colormap:    cmapOverride,
		Extend:      extend,
		Orientation: colorbarOrientation(location),
		Boundaries:  cloneFloat64s(boundaries),
		Values:      cloneFloat64s(cfg.Values),
		Spacing:     normalizeColorbarSpacing(cfg.Spacing),
		DrawEdges:   cfg.DrawEdges,
		ExtendRect:  cfg.ExtendRect,
		Alpha:       1,
		BorderColor: f.RC.AxesEdgeColor,
		BorderWidth: f.RC.AxisLineWidth,
		z:           -10,
	})

	return ax
}

func syncColorbarMapping(ax *Axes) {
	if ax == nil {
		return
	}
	for _, art := range ax.Artists {
		cb, ok := art.(*Colorbar)
		if !ok || cb == nil {
			continue
		}
		mapping := cb.currentMapping()
		cb.Mapping = mapping
		configureColorbarScale(ax, mapping, ax.colorbarLocation, ax.colorbarTicks, ax.colorbarBounds, ax.colorbarExtend)
	}
}

func insetColorbarRectForExtensions(fig *Figure, rect geom.Rect, extend, location string) geom.Rect {
	extend = normalizeColorbarExtend(extend)
	if extend == "neither" || rect.W() <= 0 || rect.H() <= 0 {
		return rect
	}
	extensionCount := colorbarExtensionCount(extend)
	if colorbarIsHorizontal(location) {
		widthFrac := rect.W() * 0.05 / (1 + 0.05*extensionCount)
		if extend == "min" || extend == "both" {
			rect.Min.X += widthFrac
		}
		if extend == "max" || extend == "both" {
			rect.Max.X -= widthFrac
		}
		if rect.Min.X >= rect.Max.X {
			return geom.Rect{}
		}
		return rect
	}

	heightFrac := rect.H() * 0.05 / (1 + 0.05*extensionCount)
	if extend == "min" || extend == "both" {
		rect.Min.Y += heightFrac
	}
	if extend == "max" || extend == "both" {
		rect.Max.Y -= heightFrac
	}
	if rect.Min.Y >= rect.Max.Y {
		return geom.Rect{}
	}
	return rect
}

func colorbarExtensionShrink(extend string) float64 {
	extensionCount := colorbarExtensionCount(extend)
	if extensionCount <= 0 {
		return 1
	}
	return 1 / (1 + 0.05*extensionCount)
}

func colorbarExtensionCount(extend string) float64 {
	extend = normalizeColorbarExtend(extend)
	extensionCount := 0.0
	if extend == "min" || extend == "both" {
		extensionCount++
	}
	if extend == "max" || extend == "both" {
		extensionCount++
	}
	return extensionCount
}

func normalizeColorbarLocation(location, orientation string) string {
	loc := strings.ToLower(strings.TrimSpace(location))
	switch loc {
	case "left", "right", "top", "bottom":
		return loc
	}
	orient := strings.ToLower(strings.TrimSpace(orientation))
	switch orient {
	case "horizontal":
		return "bottom"
	case "vertical":
		return "right"
	default:
		return "right"
	}
}

func colorbarOptionBoundaries(values, boundaries []float64) []float64 {
	if len(boundaries) >= 2 {
		return cloneFloat64s(boundaries)
	}
	if len(values) < 2 {
		return nil
	}
	out := make([]float64, len(values)+1)
	for i := 0; i+1 < len(values); i++ {
		out[i+1] = (values[i] + values[i+1]) * 0.5
	}
	out[0] = 2*out[1] - out[2]
	last := len(out) - 1
	out[last] = 2*out[last-1] - out[last-2]
	return out
}

func colorbarInteriorBoundaries(boundaries []float64, extend string) []float64 {
	out := cloneFloat64s(boundaries)
	if len(out) < 2 {
		return out
	}
	switch normalizeColorbarExtend(extend) {
	case "min":
		if len(out) > 2 {
			out = out[1:]
		}
	case "max":
		if len(out) > 2 {
			out = out[:len(out)-1]
		}
	case "both":
		if len(out) > 3 {
			out = out[1 : len(out)-1]
		}
	}
	if len(out) < 2 {
		return cloneFloat64s(boundaries)
	}
	return out
}

func colorbarInteriorValues(values []float64, boundaries []float64, extend string) []float64 {
	out := cloneFloat64s(values)
	if len(out) != len(boundaries)-1 {
		return out
	}
	switch normalizeColorbarExtend(extend) {
	case "min":
		if len(out) > 1 {
			out = out[1:]
		}
	case "max":
		if len(out) > 1 {
			out = out[:len(out)-1]
		}
	case "both":
		if len(out) > 2 {
			out = out[1 : len(out)-1]
		}
	}
	return out
}

func normalizeColorbarSpacing(spacing string) string {
	switch strings.ToLower(strings.TrimSpace(spacing)) {
	case "proportional":
		return "proportional"
	default:
		return "uniform"
	}
}

func colorbarOrientation(location string) string {
	if colorbarIsHorizontal(location) {
		return "horizontal"
	}
	return "vertical"
}

func colorbarIsHorizontal(location string) bool {
	location = strings.ToLower(strings.TrimSpace(location))
	return location == "bottom" || location == "top"
}

func resolvedColorbarPadding(base geom.Rect, padding float64, location ...string) float64 {
	horizontal := len(location) > 0 && colorbarIsHorizontal(location[0])
	if padding > 0 {
		if horizontal {
			return base.H() * padding
		}
		return base.W() * padding
	}
	if horizontal {
		return base.H() * defaultHorizontalColorbarPadding
	}
	return base.W() * defaultColorbarPadding
}

func resolvedColorbarLayoutPadding(fig *Figure, base geom.Rect, padding float64, location ...string) float64 {
	resolved := resolvedColorbarPadding(base, padding, location...)
	if padding > 0 || fig == nil || fig.layoutEngine != LayoutEngineConstrained || fig.SizePx.X <= 0 {
		return resolved
	}
	if len(location) > 0 && colorbarIsHorizontal(location[0]) {
		if fig.SizePx.Y <= 0 {
			return resolved
		}
		return resolved + layoutPadPx(fig, LayoutEngineConstrained)/fig.SizePx.Y
	}
	return resolved + layoutPadPx(fig, LayoutEngineConstrained)/fig.SizePx.X
}

func resolvedColorbarAspect(aspect float64) float64 {
	if aspect > 0 {
		return aspect
	}
	return defaultColorbarAspect
}

func resolvedColorbarWidth(fig *Figure, base geom.Rect, width, aspect float64) float64 {
	return resolvedColorbarThickness(fig, base, width, aspect, "right")
}

func resolvedColorbarThickness(fig *Figure, base geom.Rect, width, aspect float64, location string) float64 {
	if width > 0 {
		return width
	}
	if colorbarIsHorizontal(location) {
		fractionHeight := base.H() * defaultColorbarFraction
		if fig == nil || fig.SizePx.X <= 0 || fig.SizePx.Y <= 0 || aspect <= 0 {
			return fractionHeight
		}
		aspectHeight := base.W() * fig.SizePx.X / (aspect * fig.SizePx.Y)
		if aspectHeight <= 0 {
			return fractionHeight
		}
		return math.Min(fractionHeight, aspectHeight)
	}
	fractionWidth := base.W() * defaultColorbarFraction
	if fig == nil || fig.SizePx.X <= 0 || fig.SizePx.Y <= 0 || aspect <= 0 {
		return fractionWidth
	}
	aspectWidth := base.H() * fig.SizePx.Y / (aspect * fig.SizePx.X)
	if aspectWidth <= 0 {
		return fractionWidth
	}
	return math.Min(fractionWidth, aspectWidth)
}

func resolvedColorbarSlotWidth(base geom.Rect, width float64) float64 {
	return resolvedColorbarSlotThickness(base, width, "right")
}

func resolvedColorbarSlotThickness(base geom.Rect, width float64, location string) float64 {
	if width > 0 {
		return width
	}
	if colorbarIsHorizontal(location) {
		return base.H() * defaultColorbarFraction
	}
	return base.W() * defaultColorbarFraction
}

func constrainedColorbarSlotOffset(fig *Figure, base geom.Rect) float64 {
	if fig == nil || fig.SizePx.X <= 0 {
		return 0
	}
	baseWidthPx := base.W() * fig.SizePx.X
	// Matplotlib's constrained layout separates the parent axes and the colorbar
	// by the parent spine's line width in addition to the pad/space, so the
	// drawn gap is measured edge-to-edge rather than frame-to-frame. Mirror that
	// here so the colorbar lands at the same offset reserved below.
	// AxisLineWidth is stored in pixels, but the rasterized spine is snapped to
	// device pixels before the colorbar is placed against it.
	lineWidthPx := math.Round(fig.RC.AxisLineWidth)
	return (constrainedLayoutPadPx(fig) + 0.5*constrainedLayoutDefaultSpacePx(baseWidthPx, 1) + lineWidthPx) / fig.SizePx.X
}

func colorbarParentRect(base geom.Rect, width, padding float64, useResolvedSlot bool) geom.Rect {
	parent, _ := colorbarPlacementRect(nil, base, width, resolvedColorbarSlotWidth(base, width), padding, "right", useResolvedSlot)
	return parent
}

func colorbarSlotLeft(base geom.Rect, width float64, useResolvedSlot bool) float64 {
	if !useResolvedSlot {
		return base.Max.X - base.W()*defaultColorbarFraction
	}
	return base.Max.X - width
}

func colorbarPlacementRect(fig *Figure, base geom.Rect, thickness, slotThickness, padding float64, location string, useResolvedSlot bool) (geom.Rect, geom.Rect) {
	parent := base
	rect := base
	if padding < 0 {
		padding = 0
	}
	switch location {
	case "left":
		if !useResolvedSlot {
			left := base.Min.X + slotThickness + padding
			if left > base.Max.X {
				left = base.Max.X
			}
			parent.Min.X = left
		}
		slotRight := math.Min(base.Min.X+slotThickness, base.Max.X)
		rect.Min.X = math.Max(slotRight-thickness, base.Min.X)
		rect.Max.X = slotRight
	case "top":
		if !useResolvedSlot {
			top := base.Max.Y - slotThickness - padding
			if top < base.Min.Y {
				top = base.Min.Y
			}
			parent.Max.Y = top
		}
		slotBottom := math.Max(base.Max.Y-slotThickness, base.Min.Y)
		rect.Min.Y = slotBottom
		rect.Max.Y = math.Min(slotBottom+thickness, base.Max.Y)
	case "bottom":
		if !useResolvedSlot {
			bottom := base.Min.Y + slotThickness + padding
			if bottom > base.Max.Y {
				bottom = base.Max.Y
			}
			parent.Min.Y = bottom
		}
		slotTop := math.Min(base.Min.Y+slotThickness, base.Max.Y)
		rect.Min.Y = math.Max(slotTop-thickness, base.Min.Y)
		rect.Max.Y = slotTop
	default:
		if useResolvedSlot {
			slotLeft := base.Max.X + padding + constrainedColorbarSlotOffset(fig, base)
			if slotLeft+slotThickness > 1 {
				slotThickness = math.Max(thickness, 1-slotLeft)
			}
			rect.Min.X = slotLeft
			rect.Max.X = slotLeft + slotThickness
			return parent, rect
		}
		right := colorbarSlotLeft(base, thickness, useResolvedSlot) - padding
		if right > base.Min.X {
			parent.Max.X = right
		}
		slotLeft := colorbarSlotLeft(base, thickness, useResolvedSlot)
		rect.Min.X = slotLeft
		rect.Max.X = math.Min(slotLeft+thickness, base.Max.X)
	}
	return parent, rect
}

func applyColorbarShrinkAnchor(rect geom.Rect, shrink float64, anchor *geom.Pt, location string) geom.Rect {
	if rect.W() <= 0 || rect.H() <= 0 {
		return rect
	}
	if shrink <= 0 || shrink >= 1 {
		return rect
	}
	if colorbarIsHorizontal(location) {
		width := rect.W() * shrink
		ax := 0.5
		if anchor != nil {
			ax = clamp01(anchor.X)
		}
		left := rect.Min.X + (rect.W()-width)*ax
		rect.Min.X = left
		rect.Max.X = left + width
		return rect
	}
	height := rect.H() * shrink
	ay := 0.5
	if anchor != nil {
		ay = clamp01(anchor.Y)
	}
	bottom := rect.Min.Y + (rect.H()-height)*ay
	rect.Min.Y = bottom
	rect.Max.Y = bottom + height
	return rect
}

func colorbarUsesResolvedSlot(fig *Figure, parent *Axes) bool {
	return fig != nil && fig.layoutEngine == LayoutEngineConstrained && parent != nil && parent.subplotSpec != nil
}

func colorbarBaseRect(parent *Axes) geom.Rect {
	if parent == nil {
		return geom.Rect{}
	}
	if parent.subplotSpec != nil {
		return parent.subplotSpec.Rect()
	}
	return parent.RectFraction
}

func configureColorbarScale(ax *Axes, mapping ScalarMapInfo, location string, ticks, boundaries []float64, extend string) {
	if ax == nil {
		return
	}
	vmin, vmax := mapping.VMin, mapping.VMax
	if colorbarIsHorizontal(location) {
		configureHorizontalColorbarScale(ax, mapping, location, ticks, boundaries, extend)
		return
	}
	target := verticalColorbarAxis(ax, location)
	if len(boundaries) >= 2 {
		inside := colorbarInteriorBoundaries(boundaries, extend)
		ax.SetYLim(inside[0], inside[len(inside)-1])
		if target != nil {
			target.Locator = FixedLocator{TicksList: cloneFloat64s(boundaries)}
			target.Formatter = ScalarFormatter{Prec: 6}
		}
		applyExplicitColorbarTicks(ax, location, ticks)
		return
	}
	switch norm := mapping.Norm.(type) {
	case LogNorm:
		base := 10.0
		if vmin > 0 && vmax > 0 {
			ax.SetYLimLog(vmin, vmax, base)
		} else {
			ax.SetYLim(vmin, vmax)
		}
		if target != nil {
			target.Locator = LogLocator{Base: base}
			target.Formatter = LogFormatterMathText{Base: base, SciNotation: true}
		}
	case AsinhNorm:
		if isFinite(vmin) && isFinite(vmax) && vmin != vmax {
			linearWidth := asinhNormLinearWidth(norm.LinearWidth)
			ax.YScale = transform.NewAsinh(vmin, vmax, linearWidth)
			ax.yLimitsManual = true
			configureScaleAxes(ax.YAxis, ax.YAxisRight, "asinh", transform.ResolveScaleOptions(
				transform.WithScaleDomain(vmin, vmax),
				transform.WithScaleBase(10),
				transform.WithScaleLinearWidth(linearWidth),
			))
		} else {
			ax.SetYLim(vmin, vmax)
		}
	case BoundaryNorm:
		ax.SetYLim(vmin, vmax)
		if target != nil {
			target.Locator = FixedLocator{TicksList: append([]float64(nil), norm.Boundaries...)}
			target.Formatter = ScalarFormatter{Prec: 6}
		}
	default:
		if isNonlinearColorbarNorm(mapping.Norm) && isFinite(vmin) && isFinite(vmax) && vmin != vmax {
			norm := mapping.Norm
			ax.YScale = transform.NewFuncScale(vmin, vmax, norm.Map, norm.Inverse)
			ax.yLimitsManual = true
			configureScaleAxes(ax.YAxis, ax.YAxisRight, "function", transform.ResolveScaleOptions())
			applyExplicitColorbarTicks(ax, location, ticks)
			return
		}
		ax.SetYLim(vmin, vmax)
	}
	applyExplicitColorbarTicks(ax, location, ticks)
}

func verticalColorbarAxis(ax *Axes, location string) *Axis {
	if ax == nil {
		return nil
	}
	if location == "left" {
		return ax.YAxis
	}
	return ax.RightAxis()
}

func configureHorizontalColorbarScale(ax *Axes, mapping ScalarMapInfo, location string, ticks, boundaries []float64, extend string) {
	vmin, vmax := mapping.VMin, mapping.VMax
	target := ax.XAxis
	if location == "top" {
		target = ax.TopAxis()
	}
	if len(boundaries) >= 2 {
		inside := colorbarInteriorBoundaries(boundaries, extend)
		ax.SetXLim(inside[0], inside[len(inside)-1])
		if target != nil {
			target.Locator = FixedLocator{TicksList: cloneFloat64s(boundaries)}
			target.Formatter = ScalarFormatter{Prec: 6}
		}
		applyExplicitColorbarTicks(ax, location, ticks)
		return
	}
	switch norm := mapping.Norm.(type) {
	case LogNorm:
		base := 10.0
		if vmin > 0 && vmax > 0 {
			ax.SetXLimLog(vmin, vmax, base)
		} else {
			ax.SetXLim(vmin, vmax)
		}
		if target != nil {
			target.Locator = LogLocator{Base: base}
			target.Formatter = LogFormatterMathText{Base: base, SciNotation: true}
		}
	case AsinhNorm:
		if isFinite(vmin) && isFinite(vmax) && vmin != vmax {
			linearWidth := asinhNormLinearWidth(norm.LinearWidth)
			ax.XScale = transform.NewAsinh(vmin, vmax, linearWidth)
			ax.xLimitsManual = true
			configureScaleAxes(ax.XAxis, ax.XAxisTop, "asinh", transform.ResolveScaleOptions(
				transform.WithScaleDomain(vmin, vmax),
				transform.WithScaleBase(10),
				transform.WithScaleLinearWidth(linearWidth),
			))
		} else {
			ax.SetXLim(vmin, vmax)
		}
	case BoundaryNorm:
		ax.SetXLim(vmin, vmax)
		if target != nil {
			target.Locator = FixedLocator{TicksList: append([]float64(nil), norm.Boundaries...)}
			target.Formatter = ScalarFormatter{Prec: 6}
		}
	default:
		if isNonlinearColorbarNorm(mapping.Norm) && isFinite(vmin) && isFinite(vmax) && vmin != vmax {
			norm := mapping.Norm
			ax.XScale = transform.NewFuncScale(vmin, vmax, norm.Map, norm.Inverse)
			ax.xLimitsManual = true
			configureScaleAxes(ax.XAxis, ax.XAxisTop, "function", transform.ResolveScaleOptions())
			applyExplicitColorbarTicks(ax, location, ticks)
			return
		}
		ax.SetXLim(vmin, vmax)
	}
	applyExplicitColorbarTicks(ax, location, ticks)
}

func applyExplicitColorbarTicks(ax *Axes, location string, ticks []float64) {
	if ax == nil || len(ticks) == 0 {
		return
	}
	locator := FixedLocator{TicksList: cloneFloat64s(ticks)}
	formatter := ScalarFormatter{Prec: 6}
	switch location {
	case "left":
		if ax.YAxis != nil {
			ax.YAxis.Locator = locator
			ax.YAxis.Formatter = formatter
		}
	case "top":
		top := ax.TopAxis()
		top.Locator = locator
		top.Formatter = formatter
	case "bottom":
		if ax.XAxis != nil {
			ax.XAxis.Locator = locator
			ax.XAxis.Formatter = formatter
		}
	default:
		right := ax.RightAxis()
		right.Locator = locator
		right.Formatter = formatter
	}
}

func configureColorbarAxes(ax *Axes, location, label string) {
	if ax == nil {
		return
	}
	if ax.XAxis != nil {
		ax.XAxis.ShowSpine = false
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
		ax.XAxis.MinorLocator = nil
	}
	if ax.XAxisTop != nil {
		ax.XAxisTop.ShowSpine = false
		ax.XAxisTop.ShowTicks = false
		ax.XAxisTop.ShowLabels = false
		ax.XAxisTop.MinorLocator = nil
	}
	if ax.YAxis != nil {
		ax.YAxis.ShowSpine = false
		ax.YAxis.ShowTicks = false
		ax.YAxis.ShowLabels = false
		ax.YAxis.MinorLocator = nil
	}
	if ax.YAxisRight != nil {
		ax.YAxisRight.ShowSpine = false
		ax.YAxisRight.ShowTicks = false
		ax.YAxisRight.ShowLabels = false
		ax.YAxisRight.MinorLocator = nil
	}

	switch location {
	case "left":
		if ax.YAxis != nil {
			ax.YAxis.ShowTicks = true
			ax.YAxis.ShowLabels = true
		}
		_ = ax.SetYLabelPosition("left")
		if label != "" {
			ax.SetYLabel(label)
		}
	case "top":
		top := ax.TopAxis()
		top.ShowSpine = false
		top.ShowTicks = true
		top.ShowLabels = true
		top.MinorLocator = nil
		_ = ax.SetXLabelPosition("top")
		if label != "" {
			ax.SetXLabel(label)
		}
	case "bottom":
		if ax.XAxis != nil {
			ax.XAxis.ShowTicks = true
			ax.XAxis.ShowLabels = true
		}
		_ = ax.SetXLabelPosition("bottom")
		if label != "" {
			ax.SetXLabel(label)
		}
	default:
		right := ax.RightAxis()
		right.ShowSpine = true
		right.ShowTicks = true
		right.ShowLabels = true
		right.MinorLocator = nil
		_ = ax.SetYLabelPosition("right")
		if label != "" {
			ax.SetYLabel(label)
		}
	}
}

func (c *Colorbar) currentMapping() ScalarMapInfo {
	if c == nil {
		return ScalarMapInfo{}.Resolved()
	}
	mapping := c.Mapping
	if c.Mappable != nil {
		mapping = c.Mappable.ScalarMap()
	}
	mapping = mapping.Resolved()
	if c.Colormap != "" {
		mapping.Colormap = c.Colormap
	}
	return mapping
}

func isNonlinearColorbarNorm(norm ScalarNormalizer) bool {
	switch norm.(type) {
	case nil, Normalize, NoNorm:
		return false
	default:
		return norm != nil
	}
}

func normalizeColorbarExtend(extend string) string {
	switch extend {
	case "min", "max", "both":
		return extend
	default:
		return "neither"
	}
}

type colorbarExtensionPath struct {
	Path      geom.Path
	OverRange bool
}

func colorbarExtensionPaths(clip geom.Rect, extend, orientation string, extendRect bool) []colorbarExtensionPath {
	extend = normalizeColorbarExtend(extend)
	if extend == "neither" || clip.W() <= 0 || clip.H() <= 0 {
		return nil
	}
	out := make([]colorbarExtensionPath, 0, 2)
	if orientation == "horizontal" {
		width := clip.W() * 0.05
		if extend == "min" || extend == "both" {
			verts := []geom.Pt{
				{X: clip.Min.X, Y: clip.Min.Y},
				{X: clip.Min.X, Y: clip.Max.Y},
				{X: clip.Min.X - width, Y: (clip.Min.Y + clip.Max.Y) * 0.5},
			}
			if extendRect {
				verts = []geom.Pt{
					{X: clip.Min.X, Y: clip.Min.Y},
					{X: clip.Min.X, Y: clip.Max.Y},
					{X: clip.Min.X - width, Y: clip.Max.Y},
					{X: clip.Min.X - width, Y: clip.Min.Y},
				}
			}
			out = append(out, colorbarExtensionPath{
				OverRange: false,
				Path: geom.Path{
					V: verts,
					C: closedPolygonCmds(len(verts)),
				},
			})
		}
		if extend == "max" || extend == "both" {
			verts := []geom.Pt{
				{X: clip.Max.X, Y: clip.Min.Y},
				{X: clip.Max.X + width, Y: (clip.Min.Y + clip.Max.Y) * 0.5},
				{X: clip.Max.X, Y: clip.Max.Y},
			}
			if extendRect {
				verts = []geom.Pt{
					{X: clip.Max.X, Y: clip.Min.Y},
					{X: clip.Max.X + width, Y: clip.Min.Y},
					{X: clip.Max.X + width, Y: clip.Max.Y},
					{X: clip.Max.X, Y: clip.Max.Y},
				}
			}
			out = append(out, colorbarExtensionPath{
				OverRange: true,
				Path: geom.Path{
					V: verts,
					C: closedPolygonCmds(len(verts)),
				},
			})
		}
		return out
	}

	height := clip.H() * 0.05
	if extend == "min" || extend == "both" {
		verts := []geom.Pt{
			{X: clip.Min.X, Y: clip.Min.Y},
			{X: (clip.Min.X + clip.Max.X) * 0.5, Y: clip.Min.Y - height},
			{X: clip.Max.X, Y: clip.Min.Y},
		}
		if extendRect {
			verts = []geom.Pt{
				{X: clip.Min.X, Y: clip.Min.Y - height},
				{X: clip.Max.X, Y: clip.Min.Y - height},
				{X: clip.Max.X, Y: clip.Min.Y},
				{X: clip.Min.X, Y: clip.Min.Y},
			}
		}
		out = append(out, colorbarExtensionPath{
			OverRange: false,
			Path: geom.Path{
				V: verts,
				C: closedPolygonCmds(len(verts)),
			},
		})
	}
	if extend == "max" || extend == "both" {
		verts := []geom.Pt{
			{X: clip.Min.X, Y: clip.Max.Y},
			{X: clip.Max.X, Y: clip.Max.Y},
			{X: (clip.Min.X + clip.Max.X) * 0.5, Y: clip.Max.Y + height},
		}
		if extendRect {
			verts = []geom.Pt{
				{X: clip.Min.X, Y: clip.Max.Y},
				{X: clip.Max.X, Y: clip.Max.Y},
				{X: clip.Max.X, Y: clip.Max.Y + height},
				{X: clip.Min.X, Y: clip.Max.Y + height},
			}
		}
		out = append(out, colorbarExtensionPath{
			OverRange: true,
			Path: geom.Path{
				V: verts,
				C: closedPolygonCmds(len(verts)),
			},
		})
	}
	return out
}

func closedPolygonCmds(n int) []geom.Cmd {
	if n <= 0 {
		return nil
	}
	cmds := make([]geom.Cmd, n)
	for i := range cmds {
		if i == 0 {
			cmds[i] = geom.MoveTo
		} else {
			cmds[i] = geom.LineTo
		}
	}
	return append(cmds, geom.ClosePath)
}

// Draw renders a gradient across the colorbar axes.
func (c *Colorbar) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil {
		return
	}

	const gradientHeight = 256

	mapping := c.currentMapping()
	c.Mapping = mapping
	cmap := matcolor.GetColormap(mapping.Colormap)
	alpha := c.Alpha
	if alpha <= 0 {
		alpha = 1
	}
	orientation := c.normalizedOrientation()

	if boundaries, values, ok := c.boundaryData(mapping); ok {
		for i := 0; i+1 < len(boundaries); i++ {
			rect := colorbarBoundaryCellRectAt(ctx.Clip, boundaries, i, c.Spacing, orientation)
			path := snappedFillRectPath(rect)
			if len(path.C) == 0 {
				continue
			}
			value := values[i]
			col := mapping.Color(value, alpha)
			r.Path(path, &render.Paint{
				Fill:      col,
				LineJoin:  render.JoinMiter,
				LineCap:   render.CapButt,
				Antialias: render.AntialiasDefault,
			})
		}
		if c.DrawEdges {
			drawColorbarBoundaryDividers(r, ctx.Clip, boundaries, c.Spacing, orientation, c.BorderColor, c.BorderWidth)
		}
	} else {
		for i := 0; i < gradientHeight; i++ {
			t := (float64(i) + 0.5) / float64(gradientHeight)
			col := cmap.AtValue(t)
			col.A *= alpha

			path := snappedFillRectPath(colorbarCellRect(ctx.Clip, i, gradientHeight, orientation))
			if len(path.C) == 0 {
				continue
			}
			r.Path(path, &render.Paint{
				Fill:      col,
				LineJoin:  render.JoinMiter,
				LineCap:   render.CapButt,
				Antialias: render.AntialiasDefault,
			})
		}
	}

	if normalizeColorbarExtend(c.Extend) == "neither" {
		outlinePath := pixelRectPath(ctx.Clip)
		if snapped := snappedStrokeRectPath(ctx.Clip); len(snapped.C) > 0 {
			outlinePath = snapped
		}
		r.Path(outlinePath, &render.Paint{
			Stroke:    c.BorderColor,
			LineWidth: c.BorderWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
		})
	}
}

func (c *Colorbar) boundaryData(mapping ScalarMapInfo) ([]float64, []float64, bool) {
	if c == nil {
		return nil, nil, false
	}
	boundaries := cloneFloat64s(c.Boundaries)
	if len(boundaries) < 2 {
		if norm, ok := mapping.Norm.(BoundaryNorm); ok && len(norm.Boundaries) >= 2 {
			boundaries = cloneFloat64s(norm.Boundaries)
		}
	}
	if len(boundaries) < 2 {
		return nil, nil, false
	}
	values := cloneFloat64s(c.Values)
	if len(values) != len(boundaries)-1 {
		values = make([]float64, len(boundaries)-1)
		for i := range values {
			values[i] = (boundaries[i] + boundaries[i+1]) * 0.5
		}
	}
	return colorbarInteriorBoundaries(boundaries, c.Extend), colorbarInteriorValues(values, boundaries, c.Extend), true
}

func (c *Colorbar) boundaryExtensionValue(mapping ScalarMapInfo, overRange bool) (float64, bool) {
	if c == nil {
		return 0, false
	}
	extend := normalizeColorbarExtend(c.Extend)
	if extend == "neither" {
		return 0, false
	}
	boundaries := cloneFloat64s(c.Boundaries)
	if len(boundaries) < 2 {
		if norm, ok := mapping.Norm.(BoundaryNorm); ok && len(norm.Boundaries) >= 2 {
			boundaries = cloneFloat64s(norm.Boundaries)
		}
	}
	values := cloneFloat64s(c.Values)
	if len(boundaries) < 2 || len(values) != len(boundaries)-1 {
		return 0, false
	}
	if !overRange && (extend == "min" || extend == "both") {
		return values[0], true
	}
	if overRange && (extend == "max" || extend == "both") {
		return values[len(values)-1], true
	}
	return 0, false
}

func drawColorbarBoundaryDividers(r render.Renderer, clip geom.Rect, boundaries []float64, spacing, orientation string, color render.Color, width float64) {
	if r == nil || len(boundaries) < 3 || clip.W() <= 0 || clip.H() <= 0 {
		return
	}
	for i := 1; i+1 < len(boundaries); i++ {
		var path geom.Path
		if orientation == "horizontal" {
			x := colorbarBoundaryCoord(clip.Min.X, clip.Max.X, boundaries, i, spacing)
			x = math.Floor(x) + 0.5
			path = geom.Path{
				V: []geom.Pt{{X: x, Y: clip.Min.Y}, {X: x, Y: clip.Max.Y}},
				C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			}
		} else {
			y := colorbarBoundaryCoord(clip.Max.Y, clip.Min.Y, boundaries, i, spacing)
			y = math.Floor(y) + 0.5
			path = geom.Path{
				V: []geom.Pt{{X: clip.Min.X, Y: y}, {X: clip.Max.X, Y: y}},
				C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			}
		}
		r.Path(path, &render.Paint{
			Stroke:    color,
			LineWidth: width,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
		})
	}
}

func colorbarBoundaryCoord(start, end float64, boundaries []float64, index int, spacing string) float64 {
	if len(boundaries) < 2 {
		return start
	}
	t := float64(index) / float64(len(boundaries)-1)
	if normalizeColorbarSpacing(spacing) == "proportional" {
		span := boundaries[len(boundaries)-1] - boundaries[0]
		if span != 0 {
			t = (boundaries[index] - boundaries[0]) / span
		}
	}
	return start + (end-start)*t
}

func colorbarBoundaryCellRectAt(clip geom.Rect, boundaries []float64, index int, spacing, orientation string) geom.Rect {
	if index < 0 || index+1 >= len(boundaries) {
		return geom.Rect{}
	}
	if normalizeColorbarSpacing(spacing) == "proportional" {
		return colorbarBoundaryCellRect(clip, boundaries[index], boundaries[index+1], boundaries[0], boundaries[len(boundaries)-1], orientation)
	}
	return colorbarCellRect(clip, index, len(boundaries)-1, orientation)
}

// DrawOverlay renders colorbar extension patches outside the axes clip.
func (c *Colorbar) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil {
		return
	}
	if normalizeColorbarExtend(c.Extend) == "neither" {
		return
	}

	mapping := c.currentMapping()
	c.Mapping = mapping
	cmap := matcolor.GetColormap(mapping.Colormap)
	alpha := c.Alpha
	if alpha <= 0 {
		alpha = 1
	}

	orientation := c.normalizedOrientation()
	for _, ext := range colorbarExtensionPaths(ctx.Clip, c.Extend, orientation, c.ExtendRect) {
		col := render.Color{}
		if value, ok := c.boundaryExtensionValue(mapping, ext.OverRange); ok {
			col = mapping.Color(value, alpha)
		} else {
			t := -1.0
			if ext.OverRange {
				t = 2
			}
			col = cmap.AtValue(t)
			col.A *= alpha
		}
		r.Path(ext.Path, &render.Paint{
			Fill:      col,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Antialias: render.AntialiasOff,
		})
	}

	outline := colorbarExtendedOutlinePath(ctx.Clip, c.Extend, orientation, c.ExtendRect)
	if len(outline.C) > 0 {
		r.Path(outline, &render.Paint{
			Stroke:    c.BorderColor,
			LineWidth: c.BorderWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
		})
	}
}

func colorbarExtendedOutlinePath(clip geom.Rect, extend, orientation string, extendRect bool) geom.Path {
	extend = normalizeColorbarExtend(extend)
	if extend == "neither" || clip.W() <= 0 || clip.H() <= 0 {
		return geom.Path{}
	}
	if extendRect {
		outlineRect := clip
		if orientation == "horizontal" {
			width := clip.W() * 0.05
			if extend == "min" || extend == "both" {
				outlineRect.Min.X -= width
			}
			if extend == "max" || extend == "both" {
				outlineRect.Max.X += width
			}
		} else {
			height := clip.H() * 0.05
			if extend == "min" || extend == "both" {
				outlineRect.Min.Y -= height
			}
			if extend == "max" || extend == "both" {
				outlineRect.Max.Y += height
			}
		}
		return snappedStrokeRectPath(outlineRect)
	}
	if orientation == "horizontal" {
		width := clip.W() * 0.05
		left := clip.Min.X
		leftTip := left
		if extend == "min" || extend == "both" {
			leftTip -= width
		}
		right := clip.Max.X
		rightTip := right
		if extend == "max" || extend == "both" {
			rightTip += width
		}
		midY := (clip.Min.Y + clip.Max.Y) * 0.5
		verts := []geom.Pt{{X: left, Y: clip.Max.Y}}
		if extend == "min" || extend == "both" {
			verts = append(verts, geom.Pt{X: leftTip, Y: midY})
		}
		verts = append(verts, geom.Pt{X: left, Y: clip.Min.Y}, geom.Pt{X: right, Y: clip.Min.Y})
		if extend == "max" || extend == "both" {
			verts = append(verts, geom.Pt{X: rightTip, Y: midY})
		}
		verts = append(verts, geom.Pt{X: right, Y: clip.Max.Y})
		return geom.Path{V: verts, C: closedPolygonCmds(len(verts))}
	}

	height := clip.H() * 0.05
	bottom := clip.Min.Y
	bottomTip := bottom
	if extend == "min" || extend == "both" {
		bottomTip -= height
	}
	top := clip.Max.Y
	topTip := top
	if extend == "max" || extend == "both" {
		topTip += height
	}
	midX := (clip.Min.X + clip.Max.X) * 0.5
	verts := []geom.Pt{{X: clip.Min.X, Y: bottom}}
	if extend == "min" || extend == "both" {
		verts = append(verts, geom.Pt{X: midX, Y: bottomTip})
	}
	verts = append(verts, geom.Pt{X: clip.Max.X, Y: bottom}, geom.Pt{X: clip.Max.X, Y: top})
	if extend == "max" || extend == "both" {
		verts = append(verts, geom.Pt{X: midX, Y: topTip})
	}
	verts = append(verts, geom.Pt{X: clip.Min.X, Y: top})
	return geom.Path{V: verts, C: closedPolygonCmds(len(verts))}
}

func colorbarCellRect(clip geom.Rect, index, count int, orientation string) geom.Rect {
	if count <= 0 {
		return geom.Rect{}
	}
	if orientation == "horizontal" {
		x0 := clip.Min.X + clip.W()*float64(index)/float64(count)
		x1 := clip.Min.X + clip.W()*float64(index+1)/float64(count)
		return geom.Rect{
			Min: geom.Pt{X: x0, Y: clip.Min.Y},
			Max: geom.Pt{X: x1, Y: clip.Max.Y},
		}
	}
	// Display space is y-up: index 0 (the lowest scalar value) belongs at the
	// bottom of the bar (clip.Min.Y) and the last index at the top (clip.Max.Y),
	// matching the tick labels the axis system now lays out y-up.
	y0 := clip.Min.Y + clip.H()*float64(index)/float64(count)
	y1 := clip.Min.Y + clip.H()*float64(index+1)/float64(count)
	return geom.Rect{
		Min: geom.Pt{X: clip.Min.X, Y: y0},
		Max: geom.Pt{X: clip.Max.X, Y: y1},
	}
}

func colorbarBoundaryCellRect(clip geom.Rect, low, high, vmin, vmax float64, orientation string) geom.Rect {
	span := vmax - vmin
	if span == 0 {
		return geom.Rect{}
	}
	if orientation == "horizontal" {
		x0 := clip.Min.X + clip.W()*((low-vmin)/span)
		x1 := clip.Min.X + clip.W()*((high-vmin)/span)
		return geom.Rect{
			Min: geom.Pt{X: x0, Y: clip.Min.Y},
			Max: geom.Pt{X: x1, Y: clip.Max.Y},
		}
	}
	// y-up: the low boundary maps to the bottom (clip.Min.Y), the high boundary
	// to the top (clip.Max.Y).
	y0 := clip.Min.Y + clip.H()*((low-vmin)/span)
	y1 := clip.Min.Y + clip.H()*((high-vmin)/span)
	return geom.Rect{
		Min: geom.Pt{X: clip.Min.X, Y: y0},
		Max: geom.Pt{X: clip.Max.X, Y: y1},
	}
}

func (c *Colorbar) normalizedOrientation() string {
	if c != nil && strings.ToLower(strings.TrimSpace(c.Orientation)) == "horizontal" {
		return "horizontal"
	}
	return "vertical"
}

// Bounds returns an empty rect so colorbars do not affect autoscaling.
func (c *Colorbar) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the colorbar z-order.
func (c *Colorbar) Z() float64 { return c.z }

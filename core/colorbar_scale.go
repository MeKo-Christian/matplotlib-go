package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

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

func colorbarInteriorBoundaries(boundaries []float64, extend ColorbarExtend) []float64 {
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

func colorbarInteriorValues(values, boundaries []float64, extend ColorbarExtend) []float64 {
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

// colorbarAxisOps bundles the orientation-specific operations so the per-norm
// locator wiring can be written once for both vertical and horizontal colorbars.
type colorbarAxisOps struct {
	setLim    func(min, max float64)
	setLimLog func(min, max, base float64)
	setScale  func(s transform.Scale)
	setManual func()
	primary   *Axis
	secondary *Axis
	target    *Axis
}

//nolint:gocritic // The value is an immutable snapshot read by the callee.
func configureColorbarScale(ax *Axes, mapping ScalarMapInfo, location ColorbarLocation, ticks, boundaries []float64, extend ColorbarExtend) {
	if ax == nil {
		return
	}
	if colorbarIsHorizontal(location) {
		configureHorizontalColorbarScale(ax, &mapping, location, ticks, boundaries, extend)
		return
	}
	ops := colorbarAxisOps{
		setLim:    ax.SetYLim,
		setLimLog: ax.SetYLimLog,
		setScale:  func(s transform.Scale) { ax.YScale = s },
		setManual: func() { ax.yLimitsManual = true },
		primary:   ax.YAxis,
		secondary: ax.YAxisRight,
		target:    verticalColorbarAxis(ax, location),
	}
	applyColorbarNormScale(ax, ops, &mapping, location, ticks, boundaries, extend)
}

func configureHorizontalColorbarScale(ax *Axes, mapping *ScalarMapInfo, location ColorbarLocation, ticks, boundaries []float64, extend ColorbarExtend) {
	target := ax.XAxis
	if location == "top" {
		target = ax.TopAxis()
	}
	ops := colorbarAxisOps{
		setLim:    ax.SetXLim,
		setLimLog: ax.SetXLimLog,
		setScale:  func(s transform.Scale) { ax.XScale = s },
		setManual: func() { ax.xLimitsManual = true },
		primary:   ax.XAxis,
		secondary: ax.XAxisTop,
		target:    target,
	}
	applyColorbarNormScale(ax, ops, mapping, location, ticks, boundaries, extend)
}

// applyColorbarNormScale selects the colorbar axis scale, major/minor locators,
// and formatter based on the norm type, mirroring matplotlib's
// Colorbar._reset_locator_formatter_scale / _get_ticker_locator_formatter.
func applyColorbarNormScale(ax *Axes, ops colorbarAxisOps, mapping *ScalarMapInfo, location ColorbarLocation, ticks, boundaries []float64, extend ColorbarExtend) {
	vmin, vmax := mapping.VMin, mapping.VMax
	target := ops.target
	if ax != nil {
		rc := ax.resolvedRC()
		defer applyRCFormatterDefaultsToAxis(target, &rc)
	}

	if len(boundaries) >= 2 {
		inside := colorbarInteriorBoundaries(boundaries, extend)
		ops.setLim(inside[0], inside[len(inside)-1])
		if target != nil {
			target.Locator = ticker.FixedLocator{TicksList: cloneFloat64s(boundaries)}
			target.Formatter = ticker.ScalarFormatter{Prec: 6}
			if ax.colorbarMinorTicks {
				target.MinorLocator = ticker.FixedLocator{TicksList: cloneFloat64s(boundaries)}
			}
		}
		applyExplicitColorbarTicks(ax, location, ticks)
		finalizeColorbarMinorTicks(ax, ops)
		return
	}

	switch norm := mapping.Norm.(type) {
	case LogNorm:
		base := 10.0
		if vmin > 0 && vmax > 0 {
			ops.setLimLog(vmin, vmax, base)
		} else {
			ops.setLim(vmin, vmax)
		}
		if target != nil {
			target.Locator = ticker.LogLocator{Base: base}
			target.Formatter = ticker.LogFormatterMathText{Base: base, SciNotation: true}
		}
	case SymLogNorm:
		if isFinite(vmin) && isFinite(vmax) && vmin != vmax {
			base := norm.Base
			if base <= 1 {
				base = 10
			}
			linThresh := norm.LinThresh
			if linThresh <= 0 {
				linThresh = 1
			}
			linScale := norm.LinScale
			if linScale == 0 {
				linScale = 1
			}
			ops.setScale(transform.NewSymLog(vmin, vmax, base, linThresh, linScale))
			ops.setManual()
			configureScaleAxes(ops.primary, ops.secondary, "symlog", transform.ResolveScaleOptions(
				transform.WithScaleDomain(vmin, vmax),
				transform.WithScaleBase(base),
				transform.WithScaleLinThresh(linThresh),
			))
		} else {
			ops.setLim(vmin, vmax)
		}
	case AsinhNorm:
		if isFinite(vmin) && isFinite(vmax) && vmin != vmax {
			linearWidth := asinhNormLinearWidth(norm.LinearWidth)
			ops.setScale(transform.NewAsinh(vmin, vmax, linearWidth))
			ops.setManual()
			configureScaleAxes(ops.primary, ops.secondary, "asinh", transform.ResolveScaleOptions(
				transform.WithScaleDomain(vmin, vmax),
				transform.WithScaleBase(10),
				transform.WithScaleLinearWidth(linearWidth),
			))
		} else {
			ops.setLim(vmin, vmax)
		}
	case BoundaryNorm:
		ops.setLim(vmin, vmax)
		if target != nil {
			target.Locator = ticker.FixedLocator{TicksList: append([]float64(nil), norm.Boundaries...)}
			target.Formatter = ticker.ScalarFormatter{Prec: 6}
			if ax.colorbarMinorTicks {
				target.MinorLocator = ticker.FixedLocator{TicksList: append([]float64(nil), norm.Boundaries...)}
			}
		}
	case NoNorm:
		ops.setLim(vmin, vmax)
		if target != nil {
			base := 1 + math.Floor(float64(colorbarNoNormValueCount(ax, boundaries))/10.0)
			if base < 1 {
				base = 1
			}
			target.Locator = ticker.IndexLocator{Base: base, Offset: 0.5}
			target.Formatter = ticker.ScalarFormatter{Prec: 6}
		}
	case PowerNorm, TwoSlopeNorm, CenteredNorm:
		if isFinite(vmin) && isFinite(vmax) && vmin != vmax {
			n := mapping.Norm
			ops.setScale(transform.NewFuncScale(vmin, vmax, n.Map, n.Inverse))
			ops.setManual()
			configureScaleAxes(ops.primary, ops.secondary, "function", transform.ResolveScaleOptions())
			// matplotlib's function-scale default major locator is AutoLocator
			// (nice 1/2/2.5/5 ticks in data space), not LinearLocator.
			if target != nil {
				target.Locator = ticker.AutoLocator{}
				target.Formatter = ticker.ScalarFormatter{Prec: 6}
			}
		} else {
			ops.setLim(vmin, vmax)
		}
	default:
		if isNonlinearColorbarNorm(mapping.Norm) && isFinite(vmin) && isFinite(vmax) && vmin != vmax {
			n := mapping.Norm
			ops.setScale(transform.NewFuncScale(vmin, vmax, n.Map, n.Inverse))
			ops.setManual()
			configureScaleAxes(ops.primary, ops.secondary, "function", transform.ResolveScaleOptions())
			applyExplicitColorbarTicks(ax, location, ticks)
			finalizeColorbarMinorTicks(ax, ops)
			return
		}
		ops.setLim(vmin, vmax)
	}
	applyExplicitColorbarTicks(ax, location, ticks)
	finalizeColorbarMinorTicks(ax, ops)
}

// colorbarNoNormValueCount mirrors len(self._values) for a NoNorm colorbar,
// used to size the IndexLocator base (1 + int(N/10)).
func colorbarNoNormValueCount(ax *Axes, boundaries []float64) int {
	if len(boundaries) >= 2 {
		return len(boundaries) - 1
	}
	if ax != nil && len(ax.colorbarBounds) >= 2 {
		return len(ax.colorbarBounds) - 1
	}
	return 0
}

// finalizeColorbarMinorTicks handles opt-in colorbar minor ticks for scales that
// do not supply their own minor locator. Log/symlog/asinh scales install a minor
// locator that matplotlib shows by default, so those are left untouched. Linear
// and function scales have no default minor locator; ColorbarOptions.MinorTicks
// adds an AutoMinorLocator there, mirroring matplotlib's minorticks_on(). With
// MinorTicks off (the default) the scale-supplied minor locator is preserved.
func finalizeColorbarMinorTicks(ax *Axes, ops colorbarAxisOps) {
	if ops.target == nil || ax == nil || !ax.colorbarMinorTicks {
		return
	}
	if ops.target.MinorLocator == nil {
		ops.target.MinorLocator = ticker.AutoMinorLocator{}
	}
}

func verticalColorbarAxis(ax *Axes, location ColorbarLocation) *Axis {
	if ax == nil {
		return nil
	}
	if location == "left" {
		return ax.YAxis
	}
	return ax.RightAxis()
}

func applyExplicitColorbarTicks(ax *Axes, location ColorbarLocation, ticks []float64) {
	if ax == nil || len(ticks) == 0 {
		return
	}
	locator := ticker.FixedLocator{TicksList: cloneFloat64s(ticks)}
	formatter := ticker.ScalarFormatter{Prec: 6}
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

func configureColorbarAxes(ax *Axes, location ColorbarLocation, label string) {
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
		right.ShowSpine = false
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

func normalizeColorbarExtend(extend ColorbarExtend) ColorbarExtend {
	switch extend {
	case "min", "max", "both":
		return extend
	default:
		return "neither"
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

func colorbarBoundaryCellRectAt(clip geom.Rect, boundaries []float64, index int, spacing string, orientation PlotOrientation) geom.Rect {
	if index < 0 || index+1 >= len(boundaries) {
		return geom.Rect{}
	}
	if normalizeColorbarSpacing(spacing) == "proportional" {
		return colorbarBoundaryCellRect(clip, boundaries[index], boundaries[index+1], boundaries[0], boundaries[len(boundaries)-1], orientation)
	}
	return colorbarCellRect(clip, index, len(boundaries)-1, orientation)
}

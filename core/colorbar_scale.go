package core

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
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

func colorbarBoundaryCellRectAt(clip geom.Rect, boundaries []float64, index int, spacing, orientation string) geom.Rect {
	if index < 0 || index+1 >= len(boundaries) {
		return geom.Rect{}
	}
	if normalizeColorbarSpacing(spacing) == "proportional" {
		return colorbarBoundaryCellRect(clip, boundaries[index], boundaries[index+1], boundaries[0], boundaries[len(boundaries)-1], orientation)
	}
	return colorbarCellRect(clip, index, len(boundaries)-1, orientation)
}

package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
)

const (
	defaultColorbarFraction          = 0.15
	defaultColorbarPadding           = 0.05
	defaultHorizontalColorbarPadding = 0.15
	defaultColorbarAspect            = 20.0
)

func insetColorbarRectForExtensions(fig *Figure, rect geom.Rect, extend ColorbarExtend, location ColorbarLocation, fracMin, fracMax float64) geom.Rect {
	extend = normalizeColorbarExtend(extend)
	if extend == "neither" || rect.W() <= 0 || rect.H() <= 0 {
		return rect
	}
	hasMin := extend == "min" || extend == "both"
	hasMax := extend == "max" || extend == "both"
	total := colorbarExtendTotal(hasMin, hasMax, fracMin, fracMax)
	if colorbarIsHorizontal(location) {
		body := rect.W() / total
		if hasMin {
			rect.Min.X += fracMin * body
		}
		if hasMax {
			rect.Max.X -= fracMax * body
		}
		if rect.Min.X >= rect.Max.X {
			return geom.Rect{}
		}
		return rect
	}

	body := rect.H() / total
	if hasMin {
		rect.Min.Y += fracMin * body
	}
	if hasMax {
		rect.Max.Y -= fracMax * body
	}
	if rect.Min.Y >= rect.Max.Y {
		return geom.Rect{}
	}
	return rect
}

func colorbarExtensionShrink(extend ColorbarExtend, fracMin, fracMax float64) float64 {
	extend = normalizeColorbarExtend(extend)
	hasMin := extend == "min" || extend == "both"
	hasMax := extend == "max" || extend == "both"
	total := colorbarExtendTotal(hasMin, hasMax, fracMin, fracMax)
	if total <= 1 {
		return 1
	}
	return 1 / total
}

// colorbarExtendTotal returns the body-plus-extensions multiplier so a slot of
// length L holds a body of L/total with min/max extensions of frac*body each.
func colorbarExtendTotal(hasMin, hasMax bool, fracMin, fracMax float64) float64 {
	total := 1.0
	if hasMin {
		total += fracMin
	}
	if hasMax {
		total += fracMax
	}
	return total
}

// colorbarAutoExtendLengths mirrors matplotlib extendfrac='auto': uniform
// spacing yields 1/(N-1) on both sides; proportional spacing sizes each side to
// the first/last interior interval as a fraction of the interior span.
func colorbarAutoExtendLengths(interior []float64, spacing string) (float64, float64) {
	n := len(interior)
	if n < 2 {
		return 0.05, 0.05
	}
	uniform := 1.0 / float64(n-1)
	if normalizeColorbarSpacing(spacing) == "uniform" {
		return uniform, uniform
	}
	span := interior[n-1] - interior[0]
	if span == 0 {
		return uniform, uniform
	}
	return (interior[1] - interior[0]) / span, (interior[n-1] - interior[n-2]) / span
}

// colorbarExtendLengths resolves the per-side extension fractions, mirroring
// colorbar.py _get_extension_lengths(frac, automin, automax, default=0.05).
func colorbarExtendLengths(frac []float64, auto bool, automin, automax float64) (float64, float64) {
	const def = 0.05
	switch {
	case auto:
		return automin, automax
	case len(frac) >= 2:
		return frac[0], frac[1]
	case len(frac) == 1:
		return frac[0], frac[0]
	default:
		return def, def
	}
}

func normalizeColorbarLocation(location ColorbarLocation, orientation PlotOrientation) ColorbarLocation {
	loc := ColorbarLocation(strings.ToLower(strings.TrimSpace(string(location))))
	switch loc {
	case "left", "right", "top", "bottom":
		return loc
	}
	orient := strings.ToLower(strings.TrimSpace(string(orientation)))
	switch orient {
	case "horizontal":
		return "bottom"
	case "vertical":
		return "right"
	default:
		return "right"
	}
}

func colorbarOrientation(location ColorbarLocation) PlotOrientation {
	if colorbarIsHorizontal(location) {
		return "horizontal"
	}
	return "vertical"
}

func colorbarIsHorizontal(location ColorbarLocation) bool {
	location = ColorbarLocation(strings.ToLower(strings.TrimSpace(string(location))))
	return location == "bottom" || location == "top"
}

func resolvedColorbarPadding(base geom.Rect, padding float64, location ...ColorbarLocation) float64 {
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

func resolvedColorbarLayoutPadding(fig *Figure, base geom.Rect, padding float64, location ...ColorbarLocation) float64 {
	resolved := resolvedColorbarPadding(base, padding, location...)
	if padding > 0 || fig == nil || fig.layoutEngine != LayoutEngineConstrained || fig.SizePx.X <= 0 {
		return resolved
	}
	if len(location) > 0 && colorbarIsHorizontal(location[0]) {
		if fig.SizePx.Y <= 0 {
			return resolved
		}
		return resolved + constrainedLayoutPadPx(fig)/fig.SizePx.Y
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

func resolvedColorbarThickness(fig *Figure, base geom.Rect, width, aspect float64, location ColorbarLocation) float64 {
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

func resolvedColorbarSlotThickness(base geom.Rect, width float64, location ColorbarLocation) float64 {
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
	// AxisLineWidth is stored in points; convert to device pixels (the rasterized
	// spine is snapped to device pixels before the colorbar is placed against it).
	lineWidthPx := math.Round(pointsToPixels(fig.RC, fig.RC.AxisLineWidth))
	return (layoutPadPx(fig, LayoutEngineConstrained) + 0.5*constrainedLayoutDefaultSpacePx(fig, baseWidthPx, 1, true) + lineWidthPx) / fig.SizePx.X
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

func colorbarPlacementRect(fig *Figure, base geom.Rect, thickness, slotThickness, padding float64, location ColorbarLocation, useResolvedSlot bool) (geom.Rect, geom.Rect) {
	return colorbarPlacementRectWithSlotOffset(fig, base, thickness, slotThickness, padding, location, useResolvedSlot, math.NaN())
}

func colorbarPlacementRectWithSlotOffset(fig *Figure, base geom.Rect, thickness, slotThickness, padding float64, location ColorbarLocation, useResolvedSlot bool, slotOffset float64) (geom.Rect, geom.Rect) {
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
			if !math.IsNaN(slotOffset) {
				slotLeft = base.Max.X + padding + slotOffset
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

func applyColorbarShrinkAnchor(rect geom.Rect, shrink float64, anchor *geom.Pt, location ColorbarLocation) geom.Rect {
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

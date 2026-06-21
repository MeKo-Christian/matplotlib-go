package core

import (
	"math"
	"sort"
)

func (a *Axes) AutoScale(margin float64) {
	a.autoScaleAxis(true, margin, true)
	a.autoScaleAxis(false, margin, true)
	a.applyAspectDatalim()
}

func (a *Axes) autoScaleIfEnabled(margin float64) {
	if a == nil || a.ProjectionName() != "rectilinear" {
		return
	}
	a.autoScaleAxis(true, margin, true)
	a.autoScaleAxis(false, margin, true)
	a.applyAspectDatalim()
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
	effMargin := a.effectiveMargin(isX, margin)
	lowerSticky, upperSticky := stickyBounds(minVal, maxVal, stickies)
	minVal -= span * effMargin
	maxVal += span * effMargin
	if !math.IsNaN(lowerSticky) && minVal < lowerSticky {
		minVal = lowerSticky
	}
	if !math.IsNaN(upperSticky) && maxVal > upperSticky {
		maxVal = upperSticky
	}

	if a.autolimitMode == "round_numbers" {
		minVal, maxVal = a.roundNumberLimits(isX, minVal, maxVal)
	}

	if isX {
		target.XScale = replaceScaleDomain(target.XScale, minVal, maxVal)
		target.refreshUnitAxis(true)
	} else {
		target.YScale = replaceScaleDomain(target.YScale, minVal, maxVal)
		target.refreshUnitAxis(false)
	}
}

// effectiveMargin returns the per-axes autoscale margin, falling back to the
// supplied default when no explicit margin is set.
func (a *Axes) effectiveMargin(isX bool, fallback float64) float64 {
	if isX {
		if a.xMargin != nil {
			return *a.xMargin
		}
	} else if a.yMargin != nil {
		return *a.yMargin
	}
	return fallback
}

// roundNumberLimits expands [minVal, maxVal] to the extreme major ticks of the
// axis locator, mirroring Matplotlib's autolimit_mode='round_numbers'
// (Locator.view_limits returns the first and last raw tick).
func (a *Axes) roundNumberLimits(isX bool, minVal, maxVal float64) (float64, float64) {
	var loc Locator
	if isX {
		if a.XAxis != nil {
			loc = a.XAxis.Locator
		}
	} else if a.YAxis != nil {
		loc = a.YAxis.Locator
	}
	if loc == nil || minVal >= maxVal {
		return minVal, maxVal
	}
	ticks := loc.Ticks(minVal, maxVal, 0)
	if len(ticks) < 2 {
		return minVal, maxVal
	}
	lo, hi := ticks[0], ticks[len(ticks)-1]
	if math.IsNaN(lo) || math.IsNaN(hi) || lo >= hi {
		return minVal, maxVal
	}
	return lo, hi
}

// SetXMargin sets the autoscale padding fraction for the x-axis (Matplotlib
// Axes.set_xmargin). A value of 0 removes padding; the default is 0.05.
func (a *Axes) SetXMargin(m float64) {
	if a == nil || math.IsNaN(m) || math.IsInf(m, 0) || m <= -0.5 {
		return
	}
	v := m
	a.xMargin = &v
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
}

// SetYMargin sets the autoscale padding fraction for the y-axis (Matplotlib
// Axes.set_ymargin).
func (a *Axes) SetYMargin(m float64) {
	if a == nil || math.IsNaN(m) || math.IsInf(m, 0) || m <= -0.5 {
		return
	}
	v := m
	a.yMargin = &v
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
}

// Margins sets both axis autoscale padding fractions (Matplotlib Axes.margins).
func (a *Axes) Margins(x, y float64) {
	a.SetXMargin(x)
	a.SetYMargin(y)
}

// SetAutolimitMode selects how autoscaled limits are finalized: "data" (the
// default, tight to the padded data range) or "round_numbers" (snapped to the
// locator's round tick values). Mirrors rcParam axes.autolimit_mode.
func (a *Axes) SetAutolimitMode(mode string) {
	if a == nil {
		return
	}
	switch mode {
	case "", "data":
		a.autolimitMode = ""
	case "round_numbers":
		a.autolimitMode = "round_numbers"
	default:
		return
	}
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
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

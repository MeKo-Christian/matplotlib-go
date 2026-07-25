package core

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

type autoscaleBoundsInfo struct {
	bounds       geom.Rect
	hasData      bool
	minPositiveX float64
	minPositiveY float64
}

type autoscaleBoundsProvider interface {
	autoscaleBounds(*DrawContext) autoscaleBoundsInfo
}

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
	minPositive := math.Inf(1)
	var stickies []float64
	first := true

	for _, ax := range a.autoscaleAxesForTarget(isX, target) {
		for _, art := range ax.Artists {
			info := artistAutoscaleBounds(art)
			if !info.hasData {
				continue
			}
			b := info.bounds
			lo, hi := b.Min.X, b.Max.X
			artMinPositive := info.minPositiveX
			if !isX {
				lo, hi = b.Min.Y, b.Max.Y
				artMinPositive = info.minPositiveY
			}
			if math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) {
				continue
			}
			if artMinPositive > 0 && artMinPositive < minPositive {
				minPositive = artMinPositive
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

	scale := target.XScale
	if !isX {
		scale = target.YScale
	}
	minVal, maxVal = autoscaleNonsingularDomain(scale, minVal, maxVal, minPositive)
	effMargin := a.effectiveMargin(isX, margin)
	if isLogScale(scale) {
		stickies = positiveFinite(stickies)
	}
	lowerSticky, upperSticky := stickyBounds(minVal, maxVal, stickies)
	minVal, maxVal = transformedMarginDomain(scale, minVal, maxVal, effMargin)
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

func artistAutoscaleBounds(art Artist) autoscaleBoundsInfo {
	if provider, ok := art.(autoscaleBoundsProvider); ok {
		return provider.autoscaleBounds(nil)
	}
	bounds := art.Bounds(nil)
	// A zero rectangle is historically how non-data artists report no bounds.
	// Data artists that can legitimately occupy only the origin implement
	// autoscaleBoundsProvider so that presence is explicit instead.
	if bounds == (geom.Rect{}) {
		return autoscaleBoundsInfo{}
	}
	return autoscaleBoundsInfo{
		bounds:       bounds,
		hasData:      true,
		minPositiveX: minimumPositive(bounds.Min.X, bounds.Max.X),
		minPositiveY: minimumPositive(bounds.Min.Y, bounds.Max.Y),
	}
}

func pointAutoscaleBounds(points []geom.Pt) autoscaleBoundsInfo {
	info := autoscaleBoundsInfo{
		minPositiveX: math.Inf(1),
		minPositiveY: math.Inf(1),
	}
	for _, point := range points {
		if !finitePoint(point) {
			continue
		}
		if !info.hasData {
			info.bounds = geom.Rect{Min: point, Max: point}
			info.hasData = true
		} else {
			info.bounds = expandRect(info.bounds, point)
		}
		if point.X > 0 && point.X < info.minPositiveX {
			info.minPositiveX = point.X
		}
		if point.Y > 0 && point.Y < info.minPositiveY {
			info.minPositiveY = point.Y
		}
	}
	return info
}

// autoscaleNonsingularDomain mirrors Locator.nonsingular. Log axes use
// LogLocator's positive-domain and adjacent-decade behavior; the remaining
// scales use transforms.nonsingular(expander=.05).
func autoscaleNonsingularDomain(
	scale transform.Scale,
	minVal, maxVal, minPositive float64,
) (float64, float64) {
	if logScale, ok := underlyingLogScale(scale); ok {
		if minVal > maxVal {
			minVal, maxVal = maxVal, minVal
		}
		switch {
		case !isFiniteNumber(minVal) || !isFiniteNumber(maxVal), maxVal <= 0:
			return 1, logScale.Base
		case minVal <= 0:
			if !isFiniteNumber(minPositive) {
				minPositive = 1e-300
			}
			minVal = minPositive
		}
		if minVal == maxVal {
			return decadeLess(minVal, logScale.Base), decadeGreater(maxVal, logScale.Base)
		}
		return minVal, maxVal
	}
	return nonsingularAutoscaleDomain(minVal, maxVal)
}

func nonsingularAutoscaleDomain(minVal, maxVal float64) (float64, float64) {
	const (
		expander              = 0.05
		tiny                  = 1e-15
		smallestNormalFloat64 = 0x1p-1022
	)
	if !isFiniteNumber(minVal) || !isFiniteNumber(maxVal) {
		return -expander, expander
	}
	if maxVal < minVal {
		minVal, maxVal = maxVal, minVal
	}
	maxAbs := math.Max(math.Abs(minVal), math.Abs(maxVal))
	if maxAbs < (1e6/tiny)*smallestNormalFloat64 {
		return -expander, expander
	}
	if maxVal-minVal <= maxAbs*tiny {
		if minVal == 0 && maxVal == 0 {
			return -expander, expander
		}
		minVal -= expander * math.Abs(minVal)
		maxVal += expander * math.Abs(maxVal)
	}
	return minVal, maxVal
}

func transformedMarginDomain(
	scale transform.Scale,
	minVal, maxVal, margin float64,
) (float64, float64) {
	marginScale := replaceScaleDomain(scale, minVal, maxVal)
	minTransformed := marginScale.Fwd(minVal)
	maxTransformed := marginScale.Fwd(maxVal)
	delta := (maxTransformed - minTransformed) * margin
	if !isFiniteNumber(delta) {
		return minVal, maxVal
	}
	lower, lowerOK := marginScale.Inv(minTransformed - delta)
	upper, upperOK := marginScale.Inv(maxTransformed + delta)
	if !lowerOK || !upperOK || !isFiniteNumber(lower) || !isFiniteNumber(upper) {
		return minVal, maxVal
	}
	return lower, upper
}

func underlyingLogScale(scale transform.Scale) (transform.Log, bool) {
	switch value := scale.(type) {
	case transform.Log:
		return value, true
	case invertedScale:
		return underlyingLogScale(value.base)
	default:
		return transform.Log{}, false
	}
}

func isLogScale(scale transform.Scale) bool {
	_, ok := underlyingLogScale(scale)
	return ok
}

func decadeLess(value, base float64) float64 {
	less := math.Pow(base, math.Floor(math.Log(value)/math.Log(base)))
	if less == value {
		less /= base
	}
	return less
}

func decadeGreater(value, base float64) float64 {
	greater := math.Pow(base, math.Ceil(math.Log(value)/math.Log(base)))
	if greater == value {
		greater *= base
	}
	return greater
}

func minimumPositive(values ...float64) float64 {
	minimum := math.Inf(1)
	for _, value := range values {
		if value > 0 && value < minimum {
			minimum = value
		}
	}
	return minimum
}

func positiveFinite(values []float64) []float64 {
	out := values[:0]
	for _, value := range values {
		if value > 0 && isFiniteNumber(value) {
			out = append(out, value)
		}
	}
	return out
}

func isFiniteNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
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
	var loc ticker.Locator
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

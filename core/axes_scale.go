package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/transform"
)

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
		axis.Formatter = LogFormatterMathText{
			Base:              cfg.Base,
			SciNotation:       true,
			UseMinorThreshold: true,
			MinorThresholds:   [2]float64{1, 0.4},
		}
		if len(cfg.Subs) > 0 {
			axis.MinorLocator = LogLocator{Base: cfg.Base, Minor: true, Subs: cfg.Subs}
		} else {
			axis.MinorLocator = LogLocator{Base: cfg.Base, Minor: true, SubsMode: "auto"}
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

package core

import "github.com/cwbudde/matplotlib-go/style"

func scalarFormatterWithRC(formatter ScalarFormatter, rc *style.RC) ScalarFormatter {
	if rc == nil {
		return formatter
	}
	cfg := rc.Axes.Formatter
	formatter.PowerLimits = cfg.Limits
	formatter.UsePowerLimits = true
	formatter.DisableOffset = !cfg.UseOffset
	formatter.OffsetThreshold = cfg.OffsetThreshold
	formatter.offsetThresholdSet = true
	formatter.UseLocale = cfg.UseLocale
	formatter.UseMathText = cfg.UseMathText
	return formatter
}

func applyRCFormatterDefaultsToAxis(axis *Axis, rc *style.RC) {
	if axis == nil || rc == nil {
		return
	}
	axis.Formatter = formatterWithRC(axis.Formatter, rc)
	axis.MinorFormatter = formatterWithRC(axis.MinorFormatter, rc)
}

func formatterWithRC(formatter Formatter, rc *style.RC) Formatter {
	switch current := formatter.(type) {
	case ScalarFormatter:
		return scalarFormatterWithRC(current, rc)
	case LogFormatterMathText:
		current.MinExponent = rc.Axes.Formatter.MinExponent
		return current
	default:
		return formatter
	}
}

func (a *Axes) applyRCFormatterDefaults(rc *style.RC) {
	if a == nil || rc == nil {
		return
	}
	for _, axis := range []*Axis{a.XAxis, a.YAxis, a.XAxisTop, a.YAxisRight} {
		applyRCFormatterDefaultsToAxis(axis, rc)
	}
	for _, axis := range a.ExtraAxes {
		applyRCFormatterDefaultsToAxis(axis, rc)
	}
}

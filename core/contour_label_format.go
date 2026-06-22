package core

import (
	"fmt"
	"math"
)

// contourLabelFmtFormatter resolves a contour label's text from the typed
// Matplotlib clabel `fmt` variants. It ports ContourLabeler.get_text
// (contour.py): a dict maps a level to a string (falling back to "%1.3f" for
// missing keys), a format string is applied with %-formatting, and otherwise a
// Formatter (which already covers the callable case via FuncFormatter) is used.
type contourLabelFmtFormatter struct {
	dict     map[float64]string
	format   string
	fallback Formatter
}

// Format implements the Formatter interface.
func (f contourLabelFmtFormatter) Format(x float64) string {
	if f.dict != nil {
		if s, ok := lookupLevelString(f.dict, x); ok {
			return s
		}
		return fmt.Sprintf("%1.3f", x)
	}
	if f.format != "" {
		return fmt.Sprintf(f.format, x)
	}
	if f.fallback != nil {
		return f.fallback.Format(x)
	}
	return ScalarFormatter{Prec: 3}.Format(x)
}

// newContourLabelFormatter selects the label formatter for a clabel call,
// honoring (in order) FormatDict, FormatString, an explicit Formatter, the
// contour set's existing formatter, and finally a default ScalarFormatter.
func newContourLabelFormatter(opt ClabelOptions, fallback Formatter) Formatter {
	if len(opt.FormatDict) > 0 {
		return contourLabelFmtFormatter{dict: opt.FormatDict, fallback: fallback}
	}
	if opt.FormatString != "" {
		return contourLabelFmtFormatter{format: opt.FormatString, fallback: fallback}
	}
	if opt.Formatter != nil {
		return opt.Formatter
	}
	if fallback != nil {
		return fallback
	}
	return ScalarFormatter{Prec: 3}
}

// lookupLevelString finds the dict entry for a level using a tolerant match so
// that float rounding between level generation and the user-provided keys does
// not drop a label.
func lookupLevelString(dict map[float64]string, x float64) (string, bool) {
	if s, ok := dict[x]; ok {
		return s, true
	}
	for k, v := range dict {
		if math.Abs(k-x) <= 1e-9*math.Max(1, math.Abs(k)) {
			return v, true
		}
	}
	return "", false
}

// applyRightSideUp constrains a label angle to [-π/2, π/2] when rightSideUp is
// set (Matplotlib's default), keeping text upright; otherwise the raw angle is
// returned unchanged.
func applyRightSideUp(angle float64, rightSideUp bool) float64 {
	if rightSideUp {
		return normalizeLabelAngle(angle)
	}
	return angle
}

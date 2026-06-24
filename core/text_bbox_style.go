package core

import (
	"fmt"
	"strconv"
	"strings"
)

// boxStyleSpec holds the parsed result of a Matplotlib boxstyle string.
type boxStyleSpec struct {
	style        BoxStyle
	pad          float64
	hasPad       bool
	roundingSize float64
	hasRounding  bool
	toothSize    float64
	hasTooth     bool
}

// boxStyleNames maps Matplotlib boxstyle names to the BoxStyle enum. These are
// the names registered on matplotlib.patches.BoxStyle.
var boxStyleNames = map[string]BoxStyle{
	"square":     BoxStyleSquare,
	"circle":     BoxStyleCircle,
	"ellipse":    BoxStyleEllipse,
	"larrow":     BoxStyleLArrow,
	"rarrow":     BoxStyleRArrow,
	"darrow":     BoxStyleDArrow,
	"round":      BoxStyleRound,
	"round4":     BoxStyleRound4,
	"sawtooth":   BoxStyleSawtooth,
	"roundtooth": BoxStyleRoundtooth,
}

// parseBoxStyleSpec parses a Matplotlib boxstyle spec such as
// "round,pad=0.3,rounding_size=0.2". It mirrors matplotlib.patches._Style.__new__:
// whitespace is stripped, the comma-separated tokens are split, the first token is
// the lowercased style name, and the remaining tokens are key=float pairs. The
// recognized keys are pad (all styles), rounding_size (round/round4), and
// tooth_size (sawtooth/roundtooth); unknown keys are reported as errors, matching
// Matplotlib's strict argument handling.
func parseBoxStyleSpec(spec string) (boxStyleSpec, error) {
	var out boxStyleSpec
	cleaned := strings.ReplaceAll(spec, " ", "")
	parts := strings.Split(cleaned, ",")
	name := strings.ToLower(parts[0])
	style, ok := boxStyleNames[name]
	if !ok {
		return out, fmt.Errorf("unknown boxstyle: %q", spec)
	}
	out.style = style

	for _, tok := range parts[1:] {
		if tok == "" {
			continue
		}
		kv := strings.SplitN(tok, "=", 2)
		if len(kv) != 2 {
			return out, fmt.Errorf("incorrect boxstyle argument: %q", spec)
		}
		val, err := strconv.ParseFloat(kv[1], 64)
		if err != nil {
			return out, fmt.Errorf("incorrect boxstyle argument: %q", spec)
		}
		switch kv[0] {
		case "pad":
			out.pad, out.hasPad = val, true
		case "rounding_size":
			out.roundingSize, out.hasRounding = val, true
		case "tooth_size":
			out.toothSize, out.hasTooth = val, true
		default:
			return out, fmt.Errorf("unknown boxstyle argument %q in %q", kv[0], spec)
		}
	}
	return out, nil
}

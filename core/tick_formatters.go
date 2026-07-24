package core

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/style"
)

const maxEngineeringExp = 30

func scalarUsesScientific(x float64) bool {
	ax := math.Abs(x)
	return (ax >= 1e6) || (ax > 0 && ax <= 1e-4)
}

func scalarFormatterUsesScientific(f ScalarFormatter, x float64) bool {
	if f.DisableScientific {
		return false
	}
	if !f.UsePowerLimits {
		return scalarUsesScientific(x)
	}
	if x == 0 || math.IsNaN(x) || math.IsInf(x, 0) {
		return false
	}
	ax := math.Abs(x)
	exp := int(math.Floor(math.Log10(ax)))
	return exp <= f.PowerLimits[0] || exp >= f.PowerLimits[1]
}

func scalarStepPrecision(step float64) int {
	step = math.Abs(step)
	if step == 0 || math.IsNaN(step) || math.IsInf(step, 0) {
		return 0
	}

	pow10 := 1.0
	for prec := 0; prec <= 12; prec++ {
		scaled := step * pow10
		if approx(scaled, math.Round(scaled), 1e-9*math.Max(1, math.Abs(scaled))) {
			return prec
		}
		pow10 *= 10
	}
	return 6
}

func formatScalarTickLabel(f ScalarFormatter, x, step float64) string {
	if scalarFormatterUsesScientific(f, x) {
		return f.Format(x)
	}

	prec := scalarStepPrecision(step)
	if prec < 0 {
		return f.Format(x)
	}

	if approx(x, 0, 1e-12*math.Max(1, math.Abs(step))) {
		x = 0
	}

	label := scalarFixMinus(scalarFormatFixed(f, x, prec))
	if f.UseMathText {
		return scalarMathTextLabel(f, label)
	}
	return label
}

// ScalarFormatter formats numbers with fixed precision and trims trailing zeros.
// The single-value Format method uses scientific notation if |x| >= 1e6 or
// (0 < |x| <= 1e-4), unless custom power limits or scientific suppression are
// configured.
//
// When formatting a *sequence* of ticks (the axis path), the formatter mirrors
// Matplotlib's ScalarFormatter: it factors a shared additive offset and/or a
// ×10ⁿ order-of-magnitude out of the ticks into axis offset text (see
// OffsetText) and renders the ticks with uniform fixed precision. Both
// behaviours are enabled by default; set DisableOffset to suppress the additive
// offset and DisableScientific to suppress the order-of-magnitude multiplier.
type ScalarFormatter struct {
	Prec int

	// PowerLimits follows Matplotlib's inclusive scientific-notation
	// thresholds when UsePowerLimits is true: exponents <= min or >= max use
	// scientific notation. When UsePowerLimits is false the Matplotlib default
	// limits (-5, 5) drive the tick-sequence order-of-magnitude selection.
	PowerLimits    [2]int
	UsePowerLimits bool

	DisableScientific bool
	UseMathText       bool
	// UseLocale formats decimal and grouping separators according to LC_ALL,
	// LC_NUMERIC, or LANG, matching axes.formatter.use_locale.
	UseLocale bool

	// DisableOffset suppresses the shared additive offset in the tick-sequence
	// path (Matplotlib's useOffset=False).
	DisableOffset bool
	// OffsetThreshold mirrors Matplotlib's axes.formatter.offset_threshold:
	// the minimum number of leading digits an offset must save before it is
	// applied. Zero selects the Matplotlib default (4).
	OffsetThreshold int
	// offsetThresholdSet distinguishes an rc-configured zero/negative
	// threshold from ScalarFormatter's zero-value default.
	offsetThresholdSet bool
}

// scalarTickContext holds the offset / order-of-magnitude / precision that
// Matplotlib's ScalarFormatter.set_locs derives from a tick sequence.
type scalarTickContext struct {
	offset  float64
	oom     int
	sigfigs int
	valid   bool
}

func scalarEffectivePowerLimits(f ScalarFormatter) (int, int) {
	if f.UsePowerLimits {
		return f.PowerLimits[0], f.PowerLimits[1]
	}
	return -5, 6 // Matplotlib axes.formatter.limits default.
}

func scalarOffsetThreshold(f ScalarFormatter) int {
	if f.offsetThresholdSet || f.OffsetThreshold > 0 {
		return f.OffsetThreshold
	}
	return 4 // Matplotlib axes.formatter.offset_threshold default.
}

func finiteTicksCopy(ticks []float64) []float64 {
	out := make([]float64, 0, len(ticks))
	for _, t := range ticks {
		if !math.IsNaN(t) && !math.IsInf(t, 0) {
			out = append(out, t)
		}
	}
	return out
}

// newScalarTickContext ports ScalarFormatter.set_locs: compute the shared
// offset, the order of magnitude, and the fixed precision for a tick sequence.
// The ticks passed are the already-visible major (or minor) ticks.
func newScalarTickContext(f ScalarFormatter, ticks []float64) scalarTickContext {
	locs := finiteTicksCopy(ticks)
	if len(locs) == 0 {
		return scalarTickContext{}
	}
	ctx := scalarTickContext{valid: true}
	if !f.DisableOffset {
		ctx.offset = scalarComputeOffset(f, locs)
	}
	ctx.oom = scalarOrderOfMagnitude(f, locs, ctx.offset)
	ctx.sigfigs = scalarSetFormat(locs, ctx.offset, ctx.oom)
	return ctx
}

// scalarComputeOffset ports ScalarFormatter._compute_offset.
func scalarComputeOffset(f ScalarFormatter, locs []float64) float64 {
	lmin, lmax := locs[0], locs[0]
	for _, v := range locs {
		if v < lmin {
			lmin = v
		}
		if v > lmax {
			lmax = v
		}
	}
	// Only use an offset when every tick has the same (non-zero-straddling) sign.
	if lmin == lmax || (lmin <= 0 && 0 <= lmax) {
		return 0
	}
	absMin, absMax := math.Abs(lmin), math.Abs(lmax)
	if absMin > absMax {
		absMin, absMax = absMax, absMin
	}
	sign := math.Copysign(1, lmin)

	oomMax := math.Ceil(math.Log10(absMax))
	// Smallest power of ten at which floor(absMin/10^oom) != floor(absMax/10^oom).
	oom := oomMax
	for o := oomMax; ; o-- {
		if floorDiv(absMin, o) != floorDiv(absMax, o) {
			oom = o + 1
			break
		}
		if o < oomMax-50 { // numerical safety net
			oom = o
			break
		}
	}
	if (absMax-absMin)/math.Pow10(int(oom)) <= 1e-2 {
		for o := oomMax; ; o-- {
			if floorDiv(absMax, o)-floorDiv(absMin, o) > 1 {
				oom = o + 1
				break
			}
			if o < oomMax-50 {
				oom = o
				break
			}
		}
	}
	n := scalarOffsetThreshold(f) - 1
	if floorDiv(absMax, oom) >= math.Pow10(n) {
		return sign * floorDiv(absMax, oom) * math.Pow10(int(oom))
	}
	return 0
}

// floorDiv returns floor(value / 10**oom), matching Python's // operator used
// in ScalarFormatter._compute_offset.
func floorDiv(value, oom float64) float64 {
	return math.Floor(value / math.Pow10(int(oom)))
}

// scalarOrderOfMagnitude ports ScalarFormatter._set_order_of_magnitude.
func scalarOrderOfMagnitude(f ScalarFormatter, locs []float64, offset float64) int {
	if f.DisableScientific {
		return 0
	}
	lo, hi := scalarEffectivePowerLimits(f)
	if lo == hi && lo != 0 {
		return lo
	}
	var oom int
	if offset != 0 {
		minVal, maxVal := locs[0], locs[0]
		for _, v := range locs {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
		span := maxVal - minVal
		if span <= 0 {
			return 0
		}
		oom = int(math.Floor(math.Log10(span)))
	} else {
		var val float64
		for _, v := range locs {
			if a := math.Abs(v); a > val {
				val = a
			}
		}
		if val == 0 {
			return 0
		}
		oom = int(math.Floor(math.Log10(val)))
	}
	switch {
	case oom <= lo:
		return oom
	case oom >= hi:
		return oom
	default:
		return 0
	}
}

// scalarSetFormat ports ScalarFormatter._set_format, returning the number of
// fractional digits used to render every (scaled) tick.
func scalarSetFormat(locs []float64, offset float64, oom int) int {
	scale := math.Pow10(oom)
	scaled := make([]float64, len(locs))
	for i, v := range locs {
		scaled[i] = (v - offset) / scale
	}
	minVal, maxVal := scaled[0], scaled[0]
	maxAbs := 0.0
	for _, v := range scaled {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
		if a := math.Abs(v); a > maxAbs {
			maxAbs = a
		}
	}
	locRange := maxVal - minVal
	if locRange == 0 {
		locRange = maxAbs
	}
	if locRange == 0 {
		locRange = 1
	}
	locRangeOOM := int(math.Floor(math.Log10(locRange)))
	sigfigs := 3 - locRangeOOM
	if sigfigs < 0 {
		sigfigs = 0
	}
	thresh := 1e-3 * math.Pow10(locRangeOOM)
	for sigfigs >= 0 {
		maxErr := 0.0
		for _, v := range scaled {
			if e := math.Abs(v - roundDecimals(v, sigfigs)); e > maxErr {
				maxErr = e
			}
		}
		if maxErr < thresh {
			sigfigs--
		} else {
			break
		}
	}
	return sigfigs + 1
}

// roundDecimals rounds to n decimal places using banker's rounding to match
// numpy.round used in ScalarFormatter._set_format.
func roundDecimals(x float64, n int) float64 {
	scale := math.Pow10(n)
	return math.RoundToEven(x*scale) / scale
}

// formatScalarTickLabelCtx renders a single tick using a precomputed context,
// porting ScalarFormatter.__call__.
func formatScalarTickLabelCtx(f ScalarFormatter, x float64, ctx scalarTickContext) string {
	if !ctx.valid {
		return f.Format(x)
	}
	xp := (x - ctx.offset) / math.Pow10(ctx.oom)
	if math.Abs(xp) < 1e-8 {
		xp = 0
	}
	s := scalarFixMinus(scalarFormatFixed(f, xp, ctx.sigfigs))
	if f.UseMathText {
		return scalarMathTextLabel(f, s)
	}
	return s
}

// OffsetText ports ScalarFormatter.get_offset: the shared offset / ×10ⁿ text
// rendered alongside the axis. Returns "" when no offset or magnitude applies.
func (f ScalarFormatter) OffsetText(ticks []float64) string {
	ctx := newScalarTickContext(f, ticks)
	if !ctx.valid || (ctx.offset == 0 && ctx.oom == 0) {
		return ""
	}
	offsetStr := ""
	if ctx.offset != 0 {
		offsetStr = scalarFormatData(f, ctx.offset)
		if ctx.offset > 0 {
			offsetStr = "+" + offsetStr
		}
	}
	sciStr := ""
	if ctx.oom != 0 {
		if f.UseMathText {
			sciStr = scalarFormatData(f, math.Pow10(ctx.oom))
		} else {
			sciStr = fmt.Sprintf("1e%d", ctx.oom)
		}
	}
	if f.UseMathText {
		if sciStr != "" {
			sciStr = `\times\mathdefault{` + sciStr + `}`
		}
		return scalarFixMinus("$" + sciStr + `\mathdefault{` + offsetStr + `}$`)
	}
	return scalarFixMinus(sciStr + offsetStr)
}

// scalarFormatData ports ScalarFormatter.format_data.
func scalarFormatData(f ScalarFormatter, value float64) string {
	e := int(math.Floor(math.Log10(math.Abs(value))))
	s := roundDecimals(value/math.Pow10(e), 10)
	var significand string
	if s == math.Trunc(s) {
		significand = scalarFormatFixed(f, s, 0)
	} else {
		significand = scalarFormatGeneral(f, s, 10)
	}
	if e == 0 {
		return significand
	}
	exponent := scalarFormatInteger(f, e)
	if f.UseMathText {
		significand = scalarMathTextNumber(f, significand)
		exponent = scalarMathTextNumber(f, exponent)
		expStr := "10^{" + exponent + "}"
		if s == 1 {
			return expStr
		}
		return significand + ` \times ` + expStr
	}
	return significand + "e" + exponent
}

func (f ScalarFormatter) Format(x float64) string {
	if math.IsNaN(x) {
		return "NaN"
	}
	if math.IsInf(x, 1) {
		return "+Inf"
	}
	if math.IsInf(x, -1) {
		return "-Inf"
	}
	p := f.Prec
	if p < 0 {
		p = 0
	}
	if scalarFormatterUsesScientific(f, x) {
		return formatScalarScientific(f, x, p)
	}
	s := scalarTrimFixed(scalarFormatFixed(f, x, p), p)
	return scalarFixMinus(s)
}

func formatScalarScientific(f ScalarFormatter, x float64, prec int) string {
	if x == 0 {
		if f.UseMathText {
			return `$\mathdefault{0}$`
		}
		return "0"
	}
	sign := ""
	if x < 0 {
		sign = "-"
		x = -x
	}
	exp := int(math.Floor(math.Log10(x)))
	mantissa := x / math.Pow10(exp)
	if approx(mantissa, 10, 1e-12) {
		mantissa = 1
		exp++
	}
	m := scalarTrimFixed(scalarFormatFixed(f, mantissa, prec), prec)
	if f.UseMathText {
		m = scalarMathTextNumber(f, m)
		if m == "1" {
			return scalarFixMinus(fmt.Sprintf(`$\mathdefault{%s10^{%s}}$`, sign, scalarFormatInteger(f, exp)))
		}
		return scalarFixMinus(fmt.Sprintf(`$\mathdefault{%s%s\times10^{%s}}$`, sign, m, scalarFormatInteger(f, exp)))
	}
	expLabel := scalarFormatInteger(f, exp)
	if exp > 0 {
		expLabel = "+" + expLabel
	}
	return scalarFixMinus(fmt.Sprintf("%s%se%s", sign, m, expLabel))
}

// scalarFixMinus swaps ASCII hyphens for U+2212 MINUS SIGN, matplotlib's
// fix_minus. The axes.unicode_minus rcParam (default true) turns it off.
func scalarFixMinus(s string) string {
	if !style.CurrentDefaults().Axes.UnicodeMinus {
		return s
	}
	return strings.ReplaceAll(s, "-", "−")
}

// FixedFormatter returns labels by tick index.
type FixedFormatter struct {
	Labels []string
}

func (f FixedFormatter) Format(float64) string { return "" }

func (f FixedFormatter) FormatTick(_ float64, index int, _ []float64) string {
	if index < 0 || index >= len(f.Labels) {
		return ""
	}
	return f.Labels[index]
}

// NullFormatter suppresses tick labels entirely.
type NullFormatter struct{}

func (NullFormatter) Format(float64) string { return "" }

// FuncFormatter adapts a function into a Formatter.
type FuncFormatter func(float64) string

func (f FuncFormatter) Format(x float64) string {
	if f == nil {
		return ""
	}
	return f(x)
}

// FormatStrFormatter uses fmt.Sprintf formatting.
type FormatStrFormatter struct {
	Pattern string
}

func (f FormatStrFormatter) Format(x float64) string {
	if f.Pattern == "" {
		return ""
	}
	return fmt.Sprintf(f.Pattern, x)
}

// StrMethodFormatter implements a small subset of Matplotlib's "{x:.2f}" style.
type StrMethodFormatter struct {
	Template string
}

func (f StrMethodFormatter) Format(x float64) string {
	if f.Template == "" {
		return ""
	}

	out := f.Template
	for {
		start := strings.Index(out, "{x")
		if start < 0 {
			return out
		}
		end := strings.IndexByte(out[start:], '}')
		if end < 0 {
			return out
		}
		end += start

		spec := out[start+2 : end]
		repl := formatStrMethodValue(x, spec)
		out = out[:start] + repl + out[end+1:]
	}
}

// EngFormatter formats values with SI engineering prefixes.
type EngFormatter struct {
	Unit            string
	Places          int
	PlacesSet       bool
	Sep             string
	SepSet          bool
	UseUnicodeMicro bool
	UseMathText     bool
}

func (f EngFormatter) Format(x float64) string {
	sep := f.Sep
	if sep == "" && !f.SepSet {
		sep = " "
	}
	fixedPlaces := f.PlacesSet || f.Places > 0
	if x == 0 {
		value := "0"
		if fixedPlaces {
			value = strconv.FormatFloat(0, 'f', f.Places, 64)
		}
		return f.formatEngineeringValue(value, sep, "")
	}
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return (ScalarFormatter{Prec: 6}).Format(x)
	}

	absX := math.Abs(x)
	exp := int(math.Floor(math.Log10(absX)/3.0) * 3)
	if exp > maxEngineeringExp {
		exp = maxEngineeringExp
	}
	if exp < -30 {
		exp = -30
	}

	prefix := engineeringPrefix(exp)
	scaled := x / math.Pow(10, float64(exp))
	if f.UseUnicodeMicro && exp == -6 {
		prefix = "µ"
	}
	if fixedPlaces && math.Abs(parseFormattedFloat(strconv.FormatFloat(scaled, 'f', f.Places, 64))) >= 1000 && exp < maxEngineeringExp {
		scaled /= 1000
		exp += 3
		prefix = engineeringPrefix(exp)
		if f.UseUnicodeMicro && exp == -6 {
			prefix = "µ"
		}
	}

	var value string
	if fixedPlaces {
		value = strconv.FormatFloat(scaled, 'f', f.Places, 64)
	} else {
		value = strconv.FormatFloat(scaled, 'g', 6, 64)
	}
	return f.formatEngineeringValue(value, sep, prefix)
}

// FormatEng is an alias for Format, matching Matplotlib's format_eng helper.
func (f EngFormatter) FormatEng(x float64) string { return f.Format(x) }

func (f EngFormatter) formatEngineeringValue(value, sep, prefix string) string {
	suffix := ""
	if prefix != "" || f.Unit != "" {
		suffix = sep + prefix + f.Unit
	}
	if f.UseMathText {
		return scalarFixMinus(`$\mathdefault{` + value + `}$` + suffix)
	}
	return scalarFixMinus(value + suffix)
}

// PercentFormatter formats values as percentages of XMax.
type PercentFormatter struct {
	XMax         float64
	Decimals     int
	DecimalsSet  bool
	DisplayRange float64
	Symbol       string
	NoSymbol     bool
	UseTeX       bool
	IsLaTeX      bool
}

func (f PercentFormatter) Format(x float64) string {
	xMax := f.XMax
	if xMax == 0 {
		xMax = 100
	}
	symbol := f.Symbol
	if f.NoSymbol {
		symbol = ""
	} else if symbol == "" {
		symbol = "%"
	}
	if f.UseTeX && !f.IsLaTeX {
		symbol = escapeTeXSymbol(symbol)
	}
	decimals := f.Decimals
	if decimals < 0 || (!f.DecimalsSet && decimals == 0) {
		decimals = percentAutoDecimals((f.DisplayRange / xMax) * 100)
	}
	if decimals < 0 {
		decimals = 0
	}
	return scalarFixMinus(strconv.FormatFloat((x/xMax)*100, 'f', decimals, 64) + symbol)
}

func escapeTeXSymbol(symbol string) string {
	replacer := strings.NewReplacer(
		"\\", `\textbackslash{}`,
		"%", `\%`,
		"$", `\$`,
		"#", `\#`,
		"_", `\_`,
		"{", `\{`,
		"}", `\}`,
		"&", `\&`,
		"~", `\textasciitilde{}`,
		"^", `\textasciicircum{}`,
	)
	return replacer.Replace(symbol)
}

// LogFormatter formats tick labels on a log axis. For Base==10, exact
// decades use Matplotlib-style powers such as 10³. Otherwise it falls back to
// ScalarFormatter.
type LogFormatter struct {
	Base float64

	LabelOnlyBase     bool
	MinorThresholds   [2]float64
	UseMinorThreshold bool
}

func (f LogFormatter) Format(x float64) string {
	if f.LabelOnlyBase && !logTickIsDecade(x, logFormatterBase(f.Base)) {
		return ""
	}
	if logFormatterBase(f.Base) == 10 {
		if x <= 0 {
			return ""
		}
		k := math.Round(math.Log10(x))
		pow := math.Pow(10, k)
		m := x / pow
		// Tolerate small rounding
		if approx(m, 1, 1e-12) {
			return "10" + superscriptInt(int(k))
		}
	}
	// Fallback
	return (ScalarFormatter{Prec: 6}).Format(x)
}

func (f LogFormatter) FormatTick(x float64, _ int, ticks []float64) string {
	if !logFormatterShouldLabel(x, ticks, f.Base, f.LabelOnlyBase, f.UseMinorThreshold, f.MinorThresholds) {
		return ""
	}
	return f.Format(x)
}

// LogFormatterExponent formats log ticks as exponents in the selected base.
type LogFormatterExponent struct {
	Base float64

	LabelOnlyBase     bool
	MinorThresholds   [2]float64
	UseMinorThreshold bool
}

func (f LogFormatterExponent) Format(x float64) string {
	if x == 0 {
		return "0"
	}
	base := f.Base
	if base <= 1 || math.IsNaN(base) || math.IsInf(base, 0) {
		base = 10
	}
	if f.LabelOnlyBase && !logTickIsDecade(x, base) {
		return ""
	}
	if x < 0 {
		return ""
	}
	exponent := math.Log(x) / math.Log(base)
	if approx(exponent, math.Round(exponent), 1e-10) {
		exponent = math.Round(exponent)
	}
	return (ScalarFormatter{Prec: 6}).Format(exponent)
}

func (f LogFormatterExponent) FormatTick(x float64, _ int, ticks []float64) string {
	if !logFormatterShouldLabel(x, ticks, f.Base, f.LabelOnlyBase, f.UseMinorThreshold, f.MinorThresholds) {
		return ""
	}
	return f.Format(x)
}

// LogFormatterMathText formats log ticks as MathText base/exponent labels.
type LogFormatterMathText struct {
	Base        float64
	SciNotation bool
	// MinExponent displays values with smaller absolute exponents as plain
	// numbers, mirroring axes.formatter.min_exponent.
	MinExponent int

	LabelOnlyBase     bool
	MinorThresholds   [2]float64
	UseMinorThreshold bool
}

func (f LogFormatterMathText) Format(x float64) string {
	if x == 0 {
		return `$\mathdefault{0}$`
	}
	base := f.Base
	if base <= 1 || math.IsNaN(base) || math.IsInf(base, 0) {
		base = 10
	}
	if f.LabelOnlyBase && !logTickIsDecade(x, base) {
		return ""
	}
	sign := ""
	if x < 0 {
		sign = "-"
		x = -x
	}
	exponent := math.Log(x) / math.Log(base)
	isDecade := approx(exponent, math.Round(exponent), 1e-10)
	if isDecade {
		exponent = math.Round(exponent)
	}
	if math.Abs(exponent) < float64(f.MinExponent) {
		return fmt.Sprintf(`$\mathdefault{%s%g}$`, sign, x)
	}
	baseLabel := formatLogBase(base)
	if f.SciNotation && !isDecade {
		floorExp := math.Floor(exponent)
		coeff := math.Pow(base, exponent-floorExp)
		if approx(coeff, math.Round(coeff), 1e-10) {
			coeff = math.Round(coeff)
		}
		return fmt.Sprintf(`$\mathdefault{%s%g\times%s^{%d}}$`, sign, coeff, baseLabel, int(floorExp))
	}
	if !isDecade {
		return fmt.Sprintf(`$\mathdefault{%s%s^{%.2f}}$`, sign, baseLabel, exponent)
	}
	return fmt.Sprintf(`$\mathdefault{%s%s^{%d}}$`, sign, baseLabel, int(exponent))
}

func (f LogFormatterMathText) FormatTick(x float64, _ int, ticks []float64) string {
	if !logFormatterShouldLabel(x, ticks, f.Base, f.LabelOnlyBase, f.UseMinorThreshold, f.MinorThresholds) {
		return ""
	}
	return f.Format(x)
}

func formatLogBase(base float64) string {
	if approx(base, math.Round(base), 1e-12) {
		return strconv.FormatInt(int64(math.Round(base)), 10)
	}
	return strconv.FormatFloat(base, 'g', -1, 64)
}

func logFormatterBase(base float64) float64 {
	if base <= 1 || math.IsNaN(base) || math.IsInf(base, 0) {
		return 10
	}
	return base
}

func logTickIsDecade(x, base float64) bool {
	if x == 0 || math.IsNaN(x) || math.IsInf(x, 0) {
		return false
	}
	exponent := math.Log(math.Abs(x)) / math.Log(base)
	return approx(exponent, math.Round(exponent), 1e-10)
}

func logFormatterShouldLabel(x float64, ticks []float64, base float64, labelOnlyBase, useMinorThreshold bool, thresholds [2]float64) bool {
	base = logFormatterBase(base)
	isDecade := logTickIsDecade(x, base)
	if labelOnlyBase && !isDecade {
		return false
	}
	if !useMinorThreshold || isDecade {
		return true
	}
	lo, hi, ok := positiveLogTickRange(ticks, base)
	if !ok {
		return true
	}
	numDecades := hi - lo
	decadeTicks := countDecadeTicks(ticks, base)
	if decadeTicks > int(thresholds[0]) {
		return false
	}
	if numDecades > thresholds[1] {
		return logTickMantissaInSparseSubset(x, base)
	}
	return true
}

func positiveLogTickRange(ticks []float64, base float64) (float64, float64, bool) {
	have := false
	lo, hi := 0.0, 0.0
	for _, tick := range ticks {
		if tick <= 0 || math.IsNaN(tick) || math.IsInf(tick, 0) {
			continue
		}
		v := math.Log(tick) / math.Log(base)
		if !have || v < lo {
			lo = v
		}
		if !have || v > hi {
			hi = v
		}
		have = true
	}
	return lo, hi, have
}

func countDecadeTicks(ticks []float64, base float64) int {
	n := 0
	for _, tick := range ticks {
		if logTickIsDecade(tick, base) {
			n++
		}
	}
	return n
}

func logTickMantissaInSparseSubset(x, base float64) bool {
	if x <= 0 {
		return false
	}
	exp := math.Floor(math.Log(x) / math.Log(base))
	mantissa := x / math.Pow(base, exp)
	if approx(base, 10, 1e-12) {
		for _, allowed := range []float64{1, 2, 3, 4, 6, 10} {
			if approx(mantissa, allowed, 1e-10) {
				return true
			}
		}
		return false
	}
	rounded := math.Round(mantissa)
	return rounded >= 1 && rounded <= base && approx(mantissa, rounded, 1e-10)
}

func superscriptInt(v int) string {
	if v == 0 {
		return "⁰"
	}
	if v < 0 {
		return "⁻" + superscriptDigits(-v)
	}
	return superscriptDigits(v)
}

func superscriptDigits(v int) string {
	const digits = "⁰¹²³⁴⁵⁶⁷⁸⁹"
	if v == 0 {
		return "⁰"
	}
	buf := make([]string, 0, 4)
	for v > 0 {
		d := v % 10
		buf = append(buf, string([]rune(digits)[d]))
		v /= 10
	}
	for left, right := 0, len(buf)-1; left < right; left, right = left+1, right-1 {
		buf[left], buf[right] = buf[right], buf[left]
	}
	return strings.Join(buf, "")
}

// LogitFormatter formats probability ticks for logit axes.
type LogitFormatter struct {
	OneHalf     string
	UseOverline bool
	Minor       bool
}

func (f LogitFormatter) Format(x float64) string {
	if f.Minor || x <= 0 || x >= 1 || math.IsNaN(x) {
		return ""
	}
	label := ""
	if approx(2*x, 1, 1e-12) {
		if f.OneHalf != "" {
			label = f.OneHalf
		} else {
			label = `\frac{1}{2}`
		}
	} else if x < 0.5 && isPowerOfTen(x) {
		label = fmt.Sprintf("10^{%d}", int(math.Round(math.Log10(x))))
	} else if x > 0.5 && isPowerOfTen(1-x) {
		baseLabel := fmt.Sprintf("10^{%d}", int(math.Round(math.Log10(1-x))))
		if f.UseOverline {
			label = `\overline{` + baseLabel + `}`
		} else {
			label = "1-" + baseLabel
		}
	} else if x < 0.1 {
		label = logitFormatValue(x, true)
	} else if x > 0.9 {
		baseLabel := logitFormatValue(1-x, true)
		if f.UseOverline {
			label = `\overline{` + baseLabel + `}`
		} else {
			label = "1-" + baseLabel
		}
	} else {
		label = logitFormatValue(x, false)
	}
	return `$\mathdefault{` + label + `}$`
}

func logitFormatValue(x float64, sciNotation bool) string {
	if !sciNotation {
		return strconv.FormatFloat(x, 'g', -1, 64)
	}
	exp := int(math.Floor(math.Log10(x)))
	mantissa := x * math.Pow10(-exp)
	if approx(mantissa, math.Round(mantissa), 1e-12) {
		mantissa = math.Round(mantissa)
	}
	mantissaLabel := strconv.FormatFloat(mantissa, 'g', -1, 64)
	return fmt.Sprintf(`%s\cdot10^{%d}`, mantissaLabel, exp)
}

func formatStrMethodValue(x float64, spec string) string {
	spec = strings.TrimPrefix(spec, ":")
	if spec == "" {
		return (ScalarFormatter{Prec: 6}).Format(x)
	}

	verb := spec[len(spec)-1]
	precision := -1
	if dot := strings.IndexByte(spec, '.'); dot >= 0 && dot+1 < len(spec) {
		num := spec[dot+1 : len(spec)-1]
		if p, err := strconv.Atoi(num); err == nil {
			precision = p
		}
	}

	switch verb {
	case 'f', 'F', 'e', 'E', 'g', 'G':
		return strconv.FormatFloat(x, byte(verb), precision, 64)
	case '%':
		if precision < 0 {
			precision = 0
		}
		return strconv.FormatFloat(x*100, 'f', precision, 64) + "%"
	default:
		return (ScalarFormatter{Prec: 6}).Format(x)
	}
}

func parseFormattedFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return math.NaN()
	}
	return v
}

func percentAutoDecimals(displayRange float64) int {
	if displayRange <= 0 || math.IsNaN(displayRange) || math.IsInf(displayRange, 0) {
		return 0
	}
	decimals := int(math.Ceil(2.0 - math.Log10(2.0*displayRange)))
	if decimals > 5 {
		return 5
	}
	if decimals < 0 {
		return 0
	}
	return decimals
}

func engineeringPrefix(exp int) string {
	switch exp {
	case -30:
		return "q"
	case -27:
		return "r"
	case -24:
		return "y"
	case -21:
		return "z"
	case -18:
		return "a"
	case -15:
		return "f"
	case -12:
		return "p"
	case -9:
		return "n"
	case -6:
		return "µ"
	case -3:
		return "m"
	case 0:
		return ""
	case 3:
		return "k"
	case 6:
		return "M"
	case 9:
		return "G"
	case 12:
		return "T"
	case 15:
		return "P"
	case 18:
		return "E"
	case 21:
		return "Z"
	case 24:
		return "Y"
	case 27:
		return "R"
	case 30:
		return "Q"
	default:
		return ""
	}
}

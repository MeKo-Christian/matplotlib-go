package core

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Locator computes tick positions for a numeric range.
type Locator interface {
	Ticks(minVal, maxVal float64, targetCount int) []float64
}

// Formatter converts numeric tick values to strings.
type Formatter interface {
	Format(x float64) string
}

// IndexedFormatter can tailor labels to a tick's position in the sequence.
type IndexedFormatter interface {
	FormatTick(x float64, index int, ticks []float64) string
}

// LinearLocator places ticks at nice multiples of 1,2,2.5,5×10^k.
type LinearLocator struct{}

// Ticks returns a strictly increasing slice of ticks that cover [min,max]
// using the smallest step from {1,2,2.5,5,10}×10^k that does not exceed
// the requested tick density.
func (LinearLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	if targetCount <= 0 {
		targetCount = 1
	}
	if math.IsNaN(minVal) || math.IsNaN(maxVal) {
		return nil
	}
	if minVal == maxVal {
		return []float64{minVal}
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	span := maxVal - minVal
	raw := span / float64(targetCount)
	if raw <= 0 || math.IsInf(raw, 0) || math.IsNaN(raw) {
		return []float64{minVal, maxVal}
	}
	// Determine exponent of 10 for raw step.
	exp := math.Floor(math.Log10(raw))
	base := math.Pow(10, exp)
	candidates := []float64{1 * base, 2 * base, 2.5 * base, 5 * base, 10 * base}
	step := candidates[len(candidates)-1]
	for _, c := range candidates {
		if c >= raw {
			step = c
			break
		}
	}
	// Align start/end to cover [min,max]
	start := math.Floor(minVal/step) * step
	end := math.Ceil(maxVal/step) * step
	// Generate ticks
	// Guard against pathological loops
	nmax := int(2*float64(targetCount) + 20)
	var ticks []float64
	for v, i := start, 0; v <= end+0.5*step && i < nmax; v, i = v+step, i+1 {
		// Avoid negative zero
		if v == 0 {
			v = 0
		}
		ticks = append(ticks, v)
	}
	// Ensure strictly increasing and within coverage
	// Remove potential duplicates due to floating rounding
	out := make([]float64, 0, len(ticks))
	var last float64
	for i, v := range ticks {
		if i == 0 || v > last {
			out = append(out, v)
			last = v
		}
	}
	return out
}

// FixedLocator returns a predefined set of tick positions.
type FixedLocator struct {
	TicksList []float64
}

func (l FixedLocator) Ticks(minVal, maxVal float64, _ int) []float64 {
	if len(l.TicksList) == 0 {
		return nil
	}
	ticks := append([]float64(nil), l.TicksList...)
	sort.Float64s(ticks)
	return dedupeTicks(ticks)
}

// NullLocator suppresses ticks entirely.
type NullLocator struct{}

func (NullLocator) Ticks(float64, float64, int) []float64 { return nil }

// MultipleLocator places ticks at integer multiples of Base, optionally offset.
type MultipleLocator struct {
	Base   float64
	Offset float64
}

func (l MultipleLocator) Ticks(minVal, maxVal float64, _ int) []float64 {
	if l.Base <= 0 || math.IsNaN(l.Base) || math.IsInf(l.Base, 0) {
		return nil
	}
	if math.IsNaN(minVal) || math.IsNaN(maxVal) {
		return nil
	}
	if minVal == maxVal {
		return []float64{minVal}
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	startN := math.Ceil((minVal - l.Offset) / l.Base)
	endN := math.Floor((maxVal - l.Offset) / l.Base)
	if endN < startN {
		return nil
	}

	nmax := int(endN-startN) + 2
	ticks := make([]float64, 0, nmax)
	for n := startN; n <= endN; n++ {
		v := l.Offset + n*l.Base
		if approx(v, 0, 1e-12*math.Max(1, math.Abs(l.Base))) {
			v = 0
		}
		ticks = append(ticks, v)
	}
	return dedupeTicks(ticks)
}

// MaxNLocator places up to N+1 nice ticks across the view limits.
type MaxNLocator struct {
	N       int
	Integer bool
	Steps   []float64
}

func (l MaxNLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) {
		return nil
	}
	if minVal == maxVal {
		return []float64{minVal}
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	maxIntervals := l.N
	if maxIntervals <= 0 {
		maxIntervals = targetCount
	}
	if maxIntervals <= 0 {
		maxIntervals = 6
	}

	span := maxVal - minVal
	raw := span / float64(maxIntervals)
	if raw <= 0 || math.IsInf(raw, 0) || math.IsNaN(raw) {
		return []float64{minVal, maxVal}
	}

	step := niceStepCeil(raw, l.normalizedSteps())
	if l.Integer && step < 1 {
		step = 1
	}

	ticks := generateBoundedTicks(minVal, maxVal, step)
	if l.Integer {
		filtered := ticks[:0]
		for _, tick := range ticks {
			if approx(tick, math.Round(tick), 1e-9) {
				filtered = append(filtered, math.Round(tick))
			}
		}
		ticks = filtered
	}
	return dedupeTicks(ticks)
}

func (l MaxNLocator) normalizedSteps() []float64 {
	if len(l.Steps) == 0 {
		return []float64{1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10}
	}
	out := make([]float64, 0, len(l.Steps))
	for _, step := range l.Steps {
		if step > 0 && !math.IsNaN(step) && !math.IsInf(step, 0) {
			out = append(out, step)
		}
	}
	if len(out) == 0 {
		return []float64{1, 2, 2.5, 5, 10}
	}
	sort.Float64s(out)
	return dedupeTicks(out)
}

// AutoLocator is a MaxNLocator tuned for general linear axes.
type AutoLocator struct {
	MaxNLocator
}

func (l AutoLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	loc := l.MaxNLocator
	if loc.N <= 0 {
		loc.N = targetCount
	}
	if loc.N <= 0 {
		loc.N = 9
	}
	if len(loc.Steps) == 0 {
		loc.Steps = []float64{1, 2, 2.5, 5, 10}
	}
	return loc.Ticks(minVal, maxVal, targetCount)
}

func scalarUsesScientific(x float64) bool {
	ax := math.Abs(x)
	return (ax >= 1e6) || (ax > 0 && ax <= 1e-4)
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
	if scalarUsesScientific(x) {
		return f.Format(x)
	}

	prec := scalarStepPrecision(step)
	if prec < 0 {
		return f.Format(x)
	}

	if approx(x, 0, 1e-12*math.Max(1, math.Abs(step))) {
		x = 0
	}

	return scalarFixMinus(strconv.FormatFloat(x, 'f', prec, 64))
}

// MinorLinearLocator subdivides the intervals between major ticks.
// N is the number of subdivisions per major interval (e.g. N=5 gives 4 minor ticks
// between each pair of major ticks). If N <= 1, defaults to 5.
type MinorLinearLocator struct {
	N int // subdivisions per major interval
}

func (m MinorLinearLocator) Ticks(minVal, maxVal float64, _ int) []float64 {
	n := m.N
	if n <= 1 {
		n = 5
	}
	// Get major ticks to subdivide between them
	majors := (LinearLocator{}).Ticks(minVal, maxVal, 6)
	if len(majors) < 2 {
		return nil
	}

	step := (majors[1] - majors[0]) / float64(n)
	if step <= 0 {
		return nil
	}

	// Generate minor ticks across the full range, excluding major positions
	start := majors[0]
	end := majors[len(majors)-1]
	var ticks []float64
	for v := start; v <= end+step*0.5; v += step {
		// Skip if this coincides with a major tick
		isMajor := false
		for _, mj := range majors {
			if math.Abs(v-mj) < step*0.01 {
				isMajor = true
				break
			}
		}
		if !isMajor && v >= minVal && v <= maxVal {
			ticks = append(ticks, v)
		}
	}
	return ticks
}

// AutoMinorLocator subdivides automatically chosen major intervals.
type AutoMinorLocator struct {
	N     int
	Major Locator
}

func (l AutoMinorLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	n := l.N
	if n <= 1 {
		n = 5
	}

	major := l.Major
	if major == nil {
		major = AutoLocator{}
	}
	majors := major.Ticks(minVal, maxVal, targetCount)
	if len(majors) < 2 {
		return nil
	}

	ticks := make([]float64, 0, len(majors)*(n-1))
	for i := 0; i < len(majors)-1; i++ {
		start := majors[i]
		end := majors[i+1]
		step := (end - start) / float64(n)
		if step <= 0 || math.IsNaN(step) || math.IsInf(step, 0) {
			continue
		}
		for j := 1; j < n; j++ {
			v := start + step*float64(j)
			if v >= minVal && v <= maxVal {
				ticks = append(ticks, v)
			}
		}
	}
	sort.Float64s(ticks)
	return dedupeTicks(ticks)
}

// LogLocator produces logarithmic ticks for positive domains. Major ticks
// at Base^k within [min,max]. If Minor is true, places minor ticks at
// 2×Base^k and 5×Base^k where they lie within [min,max].
type LogLocator struct {
	Base  float64
	Minor bool
	Subs  []float64
}

func (l LogLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	base := l.Base
	if base <= 1 {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	if minVal <= 0 || maxVal <= 0 {
		return nil
	}
	// Find exponent range
	lb := math.Log(base)
	kmin := math.Ceil(math.Log(minVal) / lb)
	kmax := math.Floor(math.Log(maxVal)/lb + 1e-10) // Add small epsilon to handle floating point precision
	var ticks []float64
	// Majors
	for k := kmin; k <= kmax; k++ {
		v := math.Pow(base, k)
		if v >= minVal && v <= maxVal {
			ticks = append(ticks, v)
		}
		if l.Minor {
			for _, sub := range l.minorMultipliers() {
				mv := sub * math.Pow(base, k)
				if mv > v && mv < math.Pow(base, k+1) && mv >= minVal && mv <= maxVal {
					ticks = append(ticks, mv)
				}
			}
		}
	}
	sort.Float64s(ticks)
	// Deduplicate
	out := ticks[:0]
	var last float64
	first := true
	for _, v := range ticks {
		if first || v > last {
			out = append(out, v)
			last = v
			first = false
		}
	}
	return out
}

func (l LogLocator) minorMultipliers() []float64 {
	subs := l.Subs
	if len(subs) == 0 {
		subs = []float64{2, 5}
	}

	out := make([]float64, 0, len(subs))
	for _, sub := range subs {
		if sub <= 1 || sub >= l.Base {
			continue
		}
		out = append(out, sub)
	}
	sort.Float64s(out)

	deduped := out[:0]
	var last float64
	first := true
	for _, sub := range out {
		if first || !approx(sub, last, 1e-12) {
			deduped = append(deduped, sub)
			last = sub
			first = false
		}
	}
	return deduped
}

// ScalarFormatter formats numbers with fixed precision and trims trailing zeros.
// Uses scientific notation if |x| >= 1e6 or (0 < |x| <= 1e-4).
type ScalarFormatter struct{ Prec int }

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
	var s string
	if scalarUsesScientific(x) {
		s = strconv.FormatFloat(x, 'e', p, 64)
		// normalize exponent: remove leading zeros in e+00X
		if i := strings.LastIndexByte(s, 'e'); i >= 0 && i+2 < len(s) {
			sign := s[i+1]
			exp := strings.TrimLeft(s[i+2:], "0")
			if exp == "" {
				exp = "0"
			}
			s = s[:i+2] + string(sign) + exp
		}
	} else {
		s = strconv.FormatFloat(x, 'f', p, 64)
	}
	// Trim trailing zeros and possible dot
	if strings.ContainsAny(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return scalarFixMinus(s)
}

func scalarFixMinus(s string) string {
	return strings.ReplaceAll(s, "-", "\u2212")
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
	Unit   string
	Places int
	Sep    string
}

func (f EngFormatter) Format(x float64) string {
	if x == 0 {
		return "0" + f.Sep + f.Unit
	}
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return (ScalarFormatter{Prec: 6}).Format(x)
	}

	sep := f.Sep
	absX := math.Abs(x)
	exp := int(math.Floor(math.Log10(absX)/3.0) * 3)
	if exp > 24 {
		exp = 24
	}
	if exp < -24 {
		exp = -24
	}

	prefix := engineeringPrefix(exp)
	scaled := x / math.Pow(10, float64(exp))
	if f.Places >= 0 {
		return strconv.FormatFloat(scaled, 'f', f.Places, 64) + sep + prefix + f.Unit
	}
	return (ScalarFormatter{Prec: 6}).Format(scaled) + sep + prefix + f.Unit
}

// PercentFormatter formats values as percentages of XMax.
type PercentFormatter struct {
	XMax     float64
	Decimals int
	Symbol   string
}

func (f PercentFormatter) Format(x float64) string {
	xMax := f.XMax
	if xMax == 0 {
		xMax = 1
	}
	symbol := f.Symbol
	if symbol == "" {
		symbol = "%"
	}
	decimals := f.Decimals
	if decimals < 0 {
		decimals = 0
	}
	return strconv.FormatFloat((x/xMax)*100, 'f', decimals, 64) + symbol
}

// LogFormatter formats tick labels on a log axis. For Base==10, exact
// decades use Matplotlib-style powers such as 10³. Otherwise it falls back to
// ScalarFormatter.
type LogFormatter struct{ Base float64 }

func (f LogFormatter) Format(x float64) string {
	if f.Base == 10 {
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

func approx(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

func formatTickLabel(formatter Formatter, x float64, index int, ticks []float64) string {
	if formatter == nil {
		return ""
	}
	if indexed, ok := formatter.(IndexedFormatter); ok {
		return indexed.FormatTick(x, index, ticks)
	}
	return formatter.Format(x)
}

func dedupeTicks(ticks []float64) []float64 {
	if len(ticks) == 0 {
		return nil
	}
	out := ticks[:0]
	var last float64
	first := true
	for _, tick := range ticks {
		if first || !approx(tick, last, 1e-12*math.Max(1, math.Abs(tick))) {
			out = append(out, tick)
			last = tick
			first = false
		}
	}
	return out
}

func niceStepCeil(raw float64, steps []float64) float64 {
	exp := math.Floor(math.Log10(raw))
	base := math.Pow(10, exp)
	scaled := raw / base
	for _, step := range steps {
		if scaled <= step {
			return step * base
		}
	}
	return steps[len(steps)-1] * base
}

func generateBoundedTicks(minVal, maxVal, step float64) []float64 {
	if step <= 0 || math.IsNaN(step) || math.IsInf(step, 0) {
		return nil
	}
	start := math.Floor(minVal/step) * step
	end := math.Ceil(maxVal/step) * step
	nmax := int(math.Ceil((end-start)/step)) + 3
	ticks := make([]float64, 0, nmax)
	for v, i := start, 0; v <= end+0.5*step && i < nmax; v, i = v+step, i+1 {
		if approx(v, 0, 1e-12*math.Max(1, math.Abs(step))) {
			v = 0
		}
		ticks = append(ticks, v)
	}
	return ticks
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

func engineeringPrefix(exp int) string {
	switch exp {
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
		return "u"
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
	default:
		return ""
	}
}

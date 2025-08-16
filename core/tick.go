package core

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// Locator computes tick positions for a numeric range.
type Locator interface {
	Ticks(min, max float64, targetCount int) []float64
}

// Formatter converts numeric tick values to strings.
type Formatter interface {
	Format(x float64) string
}

// LinearLocator places ticks at nice multiples of 1,2,5×10^k.
type LinearLocator struct{}

// Ticks returns a strictly increasing slice of ticks that cover [min,max]
// using a step chosen from {1,2,5}×10^k close to span/targetCount.
func (LinearLocator) Ticks(min, max float64, targetCount int) []float64 {
	if targetCount <= 0 {
		targetCount = 1
	}
	if math.IsNaN(min) || math.IsNaN(max) {
		return nil
	}
	if min == max {
		return []float64{min}
	}
	if min > max {
		min, max = max, min
	}
	span := max - min
	raw := span / float64(targetCount)
	if raw <= 0 || math.IsInf(raw, 0) || math.IsNaN(raw) {
		return []float64{min, max}
	}
	// Determine exponent of 10 for raw step.
	exp := math.Floor(math.Log10(raw))
	base := math.Pow(10, exp)
	candidates := []float64{1 * base, 2 * base, 5 * base}
	step := candidates[0]
	best := math.Abs(candidates[0] - raw)
	for _, c := range candidates[1:] {
		if d := math.Abs(c - raw); d < best {
			best = d
			step = c
		}
	}
	// Align start/end to cover [min,max]
	start := math.Floor(min/step) * step
	end := math.Ceil(max/step) * step
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

// LogLocator produces logarithmic ticks for positive domains. Major ticks
// at Base^k within [min,max]. If Minor is true, places minor ticks at
// 2×Base^k and 5×Base^k where they lie within [min,max].
type LogLocator struct {
	Base  float64
	Minor bool
}

func (l LogLocator) Ticks(min, max float64, targetCount int) []float64 {
	base := l.Base
	if base <= 1 {
		return nil
	}
	if min > max {
		min, max = max, min
	}
	if min <= 0 || max <= 0 {
		return nil
	}
	// Find exponent range
	lb := math.Log(base)
	kmin := math.Ceil(math.Log(min) / lb)
	kmax := math.Floor(math.Log(max) / lb)
	var ticks []float64
	// Majors
	for k := kmin; k <= kmax; k++ {
		v := math.Pow(base, k)
		if v >= min && v <= max {
			ticks = append(ticks, v)
		}
		if l.Minor {
			// Minors at 2,5 per decade (common convention)
			m2 := 2 * math.Pow(base, k)
			m5 := 5 * math.Pow(base, k)
			if m2 > v && m2 < math.Pow(base, k+1) && m2 >= min && m2 <= max {
				ticks = append(ticks, m2)
			}
			if m5 > v && m5 < math.Pow(base, k+1) && m5 >= min && m5 <= max {
				ticks = append(ticks, m5)
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
	ax := math.Abs(x)
	var s string
	if (ax >= 1e6) || (ax > 0 && ax <= 1e-4) {
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
	return s
}

// LogFormatter formats tick labels on a log axis. For Base==10 it prefers
// forms like 1e3, 2e3, 5e3 when values are exact multiples. Otherwise it
// falls back to ScalarFormatter.
type LogFormatter struct{ Base float64 }

func (f LogFormatter) Format(x float64) string {
	if f.Base == 10 {
		if x <= 0 {
			return ""
		}
		k := math.Floor(math.Log10(x))
		pow := math.Pow(10, k)
		m := x / pow
		// Tolerate small rounding
		if approx(m, 1, 1e-12) {
			return "1e" + strconv.FormatFloat(k, 'f', 0, 64)
		}
		if approx(m, 2, 1e-12) {
			return "2e" + strconv.FormatFloat(k, 'f', 0, 64)
		}
		if approx(m, 5, 1e-12) {
			return "5e" + strconv.FormatFloat(k, 'f', 0, 64)
		}
	}
	// Fallback
	return (ScalarFormatter{Prec: 6}).Format(x)
}

func approx(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

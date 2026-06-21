package core

import (
	"math"
	"sort"
)

// Locator computes tick positions for a numeric range.
type Locator interface {
	Ticks(minVal, maxVal float64, targetCount int) []float64
}

func locatorWithMajorContext(locator, major Locator) Locator {
	if locator == nil {
		return nil
	}
	auto, ok := locator.(AutoMinorLocator)
	if !ok || auto.Major != nil {
		return locator
	}
	auto.Major = major
	return auto
}

// Formatter converts numeric tick values to strings.
type Formatter interface {
	Format(x float64) string
}

// IndexedFormatter can tailor labels to a tick's position in the sequence.
type IndexedFormatter interface {
	FormatTick(x float64, index int, ticks []float64) string
}

// OffsetFormatter can provide Matplotlib-style axis offset text for a full tick
// sequence, such as ConciseDateFormatter's shared date context.
type OffsetFormatter interface {
	OffsetText(ticks []float64) string
}

// tickLabelFormatter returns a per-tick label function for a fixed tick
// sequence. Any ScalarFormatter context is computed once up front, so formatting
// a whole sequence in a loop stays O(n) instead of recomputing the shared
// context (offset, order of magnitude, precision) for every tick.
func tickLabelFormatter(formatter Formatter, ticks []float64) func(tick float64, index int) string {
	if scalarFormatter, ok := formatter.(ScalarFormatter); ok && len(ticks) >= 1 {
		ctx := newScalarTickContext(scalarFormatter, ticks)
		return func(tick float64, _ int) string {
			return formatScalarTickLabelCtx(scalarFormatter, tick, ctx)
		}
	}
	return func(tick float64, index int) string {
		return formatTickLabel(formatter, tick, index, ticks)
	}
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

func isPowerOfTen(x float64) bool {
	if x <= 0 || math.IsNaN(x) || math.IsInf(x, 0) {
		return false
	}
	return approx(math.Log10(x), math.Round(math.Log10(x)), 1e-10)
}

func approx(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
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

func dedupeTicksByStep(ticks []float64, step float64) []float64 {
	if len(ticks) == 0 {
		return nil
	}
	tol := 1e-9 * math.Abs(step)
	if tol == 0 || math.IsNaN(tol) || math.IsInf(tol, 0) {
		return dedupeTicks(ticks)
	}
	out := ticks[:0]
	var last float64
	first := true
	for _, tick := range ticks {
		if first || !approx(tick, last, tol) {
			out = append(out, tick)
			last = tick
			first = false
		}
	}
	return out
}

func dedupeTicksSorted(ticks []float64) []float64 {
	if len(ticks) == 0 {
		return nil
	}
	out := append([]float64(nil), ticks...)
	sort.Float64s(out)
	return dedupeTicks(out)
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
	zeroTol := 1e-12 * math.Abs(step)
	for v, i := start, 0; v <= end+0.5*step && i < nmax; v, i = v+step, i+1 {
		if approx(v, 0, zeroTol) {
			v = 0
		}
		ticks = append(ticks, v)
	}
	return ticks
}

func tickMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

package core

import (
	"math"
	"sort"
	"strings"
)

// LinearLocator places ticks at nice multiples of 1,2,2.5,5×10^k by default.
//
// If NumTicks is positive, it mirrors Matplotlib's LinearLocator by returning
// exactly NumTicks evenly spaced ticks across the view interval. If Presets
// contains an exact [min,max] key, that preset wins.
type LinearLocator struct {
	NumTicks int
	Presets  map[[2]float64][]float64
}

// Ticks returns a strictly increasing slice of ticks that cover [min,max]
// using the smallest step from {1,2,2.5,5,10}×10^k that does not exceed
// the requested tick density.
func (l LinearLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	if ticks, ok := l.Presets[[2]float64{minVal, maxVal}]; ok {
		return append([]float64(nil), ticks...)
	}
	if l.NumTicks > 0 {
		switch l.NumTicks {
		case 1:
			if math.IsNaN(minVal) || math.IsNaN(maxVal) {
				return nil
			}
			return []float64{minVal}
		default:
			return linearSpacedTicks(minVal, maxVal, l.NumTicks)
		}
	}
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

func linearSpacedTicks(minVal, maxVal float64, count int) []float64 {
	if count <= 0 || math.IsNaN(minVal) || math.IsNaN(maxVal) {
		return nil
	}
	if count == 1 || minVal == maxVal {
		return []float64{minVal}
	}
	ticks := make([]float64, count)
	step := (maxVal - minVal) / float64(count-1)
	for i := range ticks {
		v := minVal + step*float64(i)
		if approx(v, 0, 1e-12*math.Max(1, math.Abs(step))) {
			v = 0
		}
		ticks[i] = v
	}
	ticks[count-1] = maxVal
	return ticks
}

// IndexLocator places ticks every Base units starting at Offset. It mirrors
// Matplotlib's IndexLocator for index-style plots.
type IndexLocator struct {
	Base   float64
	Offset float64
}

func (l IndexLocator) Ticks(minVal, maxVal float64, _ int) []float64 {
	if l.Base <= 0 || math.IsNaN(l.Base) || math.IsInf(l.Base, 0) {
		return nil
	}
	if math.IsNaN(minVal) || math.IsNaN(maxVal) {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	start := minVal + l.Offset
	tol := 1e-12 * math.Max(1, math.Abs(maxVal))
	nmax := int(math.Ceil((maxVal-start)/l.Base)) + 2
	if nmax < 1 {
		return nil
	}
	ticks := make([]float64, 0, nmax)
	for v, i := start, 0; v <= maxVal+tol && i < nmax; v, i = v+l.Base, i+1 {
		if approx(v, 0, 1e-12*math.Max(1, math.Abs(l.Base))) {
			v = 0
		}
		ticks = append(ticks, v)
	}
	return dedupeTicks(ticks)
}

// FixedLocator returns a predefined set of tick positions.
type FixedLocator struct {
	TicksList []float64
	Nbins     int
}

func (l FixedLocator) Ticks(minVal, maxVal float64, _ int) []float64 {
	if len(l.TicksList) == 0 {
		return nil
	}
	ticks := append([]float64(nil), l.TicksList...)
	sort.Float64s(ticks)
	ticks = dedupeTicks(ticks)
	if l.Nbins <= 0 {
		return ticks
	}
	nbins := l.Nbins
	if nbins < 2 {
		nbins = 2
	}
	step := int(math.Ceil(float64(len(ticks)) / float64(nbins)))
	if step < 1 {
		step = 1
	}
	best := everyNthTick(ticks, 0, step)
	for offset := 1; offset < step; offset++ {
		candidate := everyNthTick(ticks, offset, step)
		if len(candidate) == 0 {
			continue
		}
		if minAbs(candidate) < minAbs(best) {
			best = candidate
		}
	}
	return append([]float64(nil), best...)
}

func everyNthTick(ticks []float64, offset, step int) []float64 {
	if step <= 0 || offset < 0 || offset >= len(ticks) {
		return nil
	}
	out := make([]float64, 0, (len(ticks)-offset+step-1)/step)
	for i := offset; i < len(ticks); i += step {
		out = append(out, ticks[i])
	}
	return out
}

func minAbs(xs []float64) float64 {
	if len(xs) == 0 {
		return math.Inf(1)
	}
	minVal := math.Abs(xs[0])
	for _, x := range xs[1:] {
		if ax := math.Abs(x); ax < minVal {
			minVal = ax
		}
	}
	return minVal
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
	N         int
	Integer   bool
	Steps     []float64
	Symmetric bool
	Prune     string
	MinTicks  int
}

func (l MaxNLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) || math.IsInf(minVal, 0) || math.IsInf(maxVal, 0) {
		return nil
	}
	if minVal == maxVal {
		expand := math.Max(1, math.Abs(minVal))*1e-13 + 1e-14
		minVal -= expand
		maxVal += expand
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	if l.Symmetric {
		bound := math.Max(math.Abs(minVal), math.Abs(maxVal))
		minVal, maxVal = -bound, bound
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
	integerCount := math.Floor(maxVal) - math.Ceil(minVal) + 1
	integerMode := l.Integer && integerCount >= float64(l.minTicks())
	if integerMode && step < 1 {
		step = 1
	}

	ticks := generateBoundedTicks(minVal, maxVal, step)
	if integerMode {
		filtered := ticks[:0]
		for _, tick := range ticks {
			if approx(tick, math.Round(tick), 1e-9) {
				filtered = append(filtered, math.Round(tick))
			}
		}
		ticks = filtered
	}
	return l.pruneTicks(dedupeTicksByStep(ticks, step))
}

func (l MaxNLocator) normalizedSteps() []float64 {
	if len(l.Steps) == 0 {
		return []float64{1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10}
	}
	out := make([]float64, 0, len(l.Steps))
	for _, step := range l.Steps {
		if step >= 1 && step <= 10 && !math.IsNaN(step) && !math.IsInf(step, 0) {
			out = append(out, step)
		}
	}
	if len(out) == 0 {
		return []float64{1, 2, 2.5, 5, 10}
	}
	sort.Float64s(out)
	out = dedupeTicks(out)
	if out[0] != 1 {
		out = append([]float64{1}, out...)
	}
	if out[len(out)-1] != 10 {
		out = append(out, 10)
	}
	return out
}

func (l MaxNLocator) minTicks() int {
	if l.MinTicks > 0 {
		return l.MinTicks
	}
	return 2
}

func (l MaxNLocator) pruneTicks(ticks []float64) []float64 {
	switch strings.ToLower(strings.TrimSpace(l.Prune)) {
	case "lower":
		if len(ticks) > 0 {
			return ticks[1:]
		}
	case "upper":
		if len(ticks) > 0 {
			return ticks[:len(ticks)-1]
		}
	case "both":
		if len(ticks) > 2 {
			return ticks[1 : len(ticks)-1]
		}
		return nil
	}
	return ticks
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
	Base     float64
	Minor    bool
	Subs     []float64
	SubsMode string
	NumTicks int
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
	nDecades := int(kmax-kmin) + 1
	numTicks := l.NumTicks
	if numTicks <= 0 {
		numTicks = targetCount
	}
	if numTicks <= 0 {
		numTicks = 9
	}
	if numTicks < 2 {
		numTicks = 2
	}
	stride := 1
	if nDecades > 0 {
		stride = nDecades/(numTicks+1) + 1
	}
	var ticks []float64
	multipliers := l.minorMultipliers(nDecades)
	includeMultipliers := l.Minor || l.SubsMode != "" || len(l.Subs) > 0
	if l.Minor && includeMultipliers && len(multipliers) == 0 {
		return nil
	}
	if includeMultipliers && len(multipliers) > 0 && stride > 1 && nDecades > 1 {
		return nil
	}
	includeMajors := strings.ToLower(strings.TrimSpace(l.SubsMode)) != "auto"
	// Majors
	for k := kmin; k <= kmax; k += float64(stride) {
		v := math.Pow(base, k)
		if includeMajors && v >= minVal && v <= maxVal {
			ticks = append(ticks, v)
		}
		if includeMultipliers {
			for _, sub := range multipliers {
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

func (l LogLocator) minorMultipliers(nDecades int) []float64 {
	subs := l.Subs
	if len(subs) == 0 {
		switch strings.ToLower(strings.TrimSpace(l.SubsMode)) {
		case "auto", "":
			if l.Minor && (nDecades >= 10 || l.Base < 3) {
				return nil
			}
			if l.SubsMode == "" && l.Minor {
				subs = []float64{2, 5}
			} else {
				subs = logAutoSubs(l.Base, false)
			}
		case "all":
			if nDecades >= 10 || l.Base < 3 {
				subs = []float64{1}
			} else {
				subs = logAutoSubs(l.Base, true)
			}
		default:
			if l.Minor {
				subs = []float64{2, 5}
			} else {
				return nil
			}
		}
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

func logAutoSubs(base float64, includeOne bool) []float64 {
	if base <= 1 || math.IsNaN(base) || math.IsInf(base, 0) {
		return nil
	}
	start := 2.0
	if includeOne {
		start = 1
	}
	limit := math.Ceil(base)
	out := make([]float64, 0, int(math.Max(0, limit-start)))
	for sub := start; sub < limit && sub < base; sub++ {
		out = append(out, sub)
	}
	return out
}

// SymLogLocator places ticks linearly around zero and logarithmically outside
// the linear threshold, matching Matplotlib's SymmetricalLogLocator.
type SymLogLocator struct {
	Base      float64
	LinThresh float64
	Subs      []float64
	NumTicks  int
}

func (l SymLogLocator) Ticks(minVal, maxVal float64, _ int) []float64 {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	base := l.Base
	if base <= 1 || math.IsNaN(base) || math.IsInf(base, 0) {
		base = 10
	}
	linThresh := l.LinThresh
	if linThresh <= 0 || math.IsNaN(linThresh) || math.IsInf(linThresh, 0) {
		linThresh = 1
	}
	if minVal >= -linThresh && maxVal <= linThresh {
		return dedupeTicksSorted([]float64{minVal, 0, maxVal})
	}

	hasLower := minVal < -linThresh
	hasUpper := maxVal > linThresh
	hasLinear := (hasLower && maxVal > -linThresh) || (hasUpper && minVal < linThresh)

	lowerStart, lowerEnd := 0, 0
	if hasLower {
		upperLimit := math.Min(-linThresh, maxVal)
		lowerStart, lowerEnd = symlogExponentRange(math.Abs(upperLimit), math.Abs(minVal)+1, base)
	}
	upperStart, upperEnd := 0, 0
	if hasUpper {
		lowerLimit := math.Max(linThresh, minVal)
		upperStart, upperEnd = symlogExponentRange(lowerLimit, maxVal+1, base)
	}

	totalTicks := (lowerEnd - lowerStart) + (upperEnd - upperStart)
	if hasLinear {
		totalTicks++
	}
	numTicks := l.NumTicks
	if numTicks <= 1 {
		numTicks = 15
	}
	stride := totalTicks / (numTicks - 1)
	if stride < 1 {
		stride = 1
	}

	decades := make([]float64, 0, totalTicks)
	if hasLower {
		for exp := lowerEnd - 1; exp >= lowerStart; exp -= stride {
			decades = append(decades, -math.Pow(base, float64(exp)))
		}
	}
	if hasLinear {
		decades = append(decades, 0)
	}
	if hasUpper {
		for exp := upperStart; exp < upperEnd; exp += stride {
			decades = append(decades, math.Pow(base, float64(exp)))
		}
	}

	subs := l.Subs
	if len(subs) == 0 {
		subs = []float64{1}
	}
	ticks := make([]float64, 0, len(decades)*len(subs))
	for _, decade := range decades {
		if decade == 0 {
			ticks = append(ticks, 0)
			continue
		}
		for _, sub := range subs {
			if sub <= 0 || math.IsNaN(sub) || math.IsInf(sub, 0) {
				continue
			}
			tick := sub * decade
			if tick >= minVal && tick <= maxVal {
				ticks = append(ticks, tick)
			}
		}
	}
	return dedupeTicksSorted(ticks)
}

func symlogExponentRange(lo, hi, base float64) (int, int) {
	if lo <= 0 || hi <= 0 || hi < lo {
		return 0, 0
	}
	logBase := math.Log(base)
	start := int(math.Floor(math.Log(lo) / logBase))
	end := int(math.Ceil(math.Log(hi) / logBase))
	return start, end
}

// AsinhLocator places ticks approximately evenly on an inverse-sinh scale.
type AsinhLocator struct {
	LinearWidth float64
	NumTicks    int
	SymThresh   float64
	Base        float64
	Subs        []float64
}

func (l AsinhLocator) Ticks(minVal, maxVal float64, _ int) []float64 {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	if minVal*maxVal < 0 {
		symThresh := l.SymThresh
		if symThresh <= 0 {
			symThresh = 0.2
		}
		if minVal != 0 && math.Abs(1+maxVal/minVal) < symThresh {
			bound := math.Max(math.Abs(minVal), math.Abs(maxVal))
			minVal, maxVal = -bound, bound
		}
	}

	linearWidth := l.LinearWidth
	if linearWidth <= 0 || math.IsNaN(linearWidth) || math.IsInf(linearWidth, 0) {
		linearWidth = 1
	}
	numTicks := l.NumTicks
	if numTicks <= 0 {
		numTicks = 11
	}

	yMin := linearWidth * math.Asinh(minVal/linearWidth)
	yMax := linearWidth * math.Asinh(maxVal/linearWidth)
	if yMin == yMax {
		return linearSpacedTicks(minVal, maxVal, numTicks)
	}
	ys := linearSpacedTicks(yMin, yMax, numTicks)
	if yMin*yMax < 0 {
		kept := ys[:0]
		for _, y := range ys {
			if math.Abs(y/(yMax-yMin)) > 0.5/float64(numTicks) {
				kept = append(kept, y)
			}
		}
		ys = append(kept, 0)
	}

	xs := make([]float64, 0, len(ys))
	for _, y := range ys {
		xs = append(xs, linearWidth*math.Sinh(y/linearWidth))
	}

	ticks := asinhRoundedTicks(xs, l.Base, l.Subs)
	if len(ticks) >= 2 {
		return ticks
	}
	return linearSpacedTicks(minVal, maxVal, numTicks)
}

func asinhRoundedTicks(xs []float64, base float64, subs []float64) []float64 {
	if base == 0 {
		base = 10
	}
	ticks := make([]float64, 0, len(xs)*tickMaxInt(1, len(subs)))
	if base > 1 {
		logBase := math.Log(base)
		for _, x := range xs {
			pow := 0.0
			switch {
			case x > 0:
				pow = math.Pow(base, math.Floor(math.Log(x)/logBase))
			case x < 0:
				pow = -math.Pow(base, math.Floor(math.Log(-x)/logBase))
			}
			if len(subs) == 0 {
				ticks = append(ticks, pow)
				continue
			}
			for _, sub := range subs {
				if sub <= 0 || math.IsNaN(sub) || math.IsInf(sub, 0) {
					continue
				}
				ticks = append(ticks, pow*sub)
			}
		}
		return dedupeTicksSorted(ticks)
	}

	for _, x := range xs {
		pow := 1.0
		if x != 0 {
			pow = math.Pow(10, math.Floor(math.Log10(math.Abs(x))))
		}
		ticks = append(ticks, pow*math.Round(x/pow))
	}
	return dedupeTicksSorted(ticks)
}

// LogitLocator places ticks on probability axes in Matplotlib's logit pattern:
// ..., 1e-2, 1e-1, 1/2, 1-1e-1, 1-1e-2, ...
type LogitLocator struct {
	Minor bool
	Nbins int
}

func (l LogitLocator) Ticks(minVal, maxVal float64, _ int) []float64 {
	minVal, maxVal = logitNonsingular(minVal, maxVal)
	nbins := l.Nbins
	if nbins <= 0 {
		nbins = 9
	}

	bInf := logitLowerIdealIndex(minVal)
	bSup := logitUpperIdealIndex(maxVal)
	numIdeal := bSup - bInf - 1
	if numIdeal >= 2 {
		if numIdeal > nbins {
			factor := int(math.Ceil(float64(numIdeal) / float64(nbins)))
			ticks := make([]float64, 0, numIdeal/factor+2)
			for b := bInf; b <= bSup; b++ {
				isMajor := b%factor == 0
				if l.Minor == isMajor {
					continue
				}
				ticks = append(ticks, logitIdealTick(b))
			}
			return visibleTicks(dedupeTicksSorted(ticks), minVal, maxVal)
		}
		if l.Minor {
			return visibleTicks(logitMinorTicks(bInf, bSup), minVal, maxVal)
		}
		ticks := make([]float64, 0, bSup-bInf+1)
		for b := bInf; b <= bSup; b++ {
			ticks = append(ticks, logitIdealTick(b))
		}
		return visibleTicks(dedupeTicksSorted(ticks), minVal, maxVal)
	}
	if l.Minor {
		return nil
	}
	return (MaxNLocator{N: nbins, Steps: []float64{1, 2, 5, 10}}).Ticks(minVal, maxVal, nbins)
}

func logitNonsingular(minVal, maxVal float64) (float64, float64) {
	const minPos = 1e-7
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	if math.IsNaN(minVal) || math.IsNaN(maxVal) || math.IsInf(minVal, 0) || math.IsInf(maxVal, 0) || maxVal <= 0 || minVal >= 1 {
		return minPos, 1 - minPos
	}
	if minVal <= 0 {
		minVal = minPos
	}
	if maxVal >= 1 {
		maxVal = 1 - minPos
	}
	if minVal == maxVal {
		minVal, maxVal = 0.1*minVal, 1-0.1*minVal
	}
	return minVal, maxVal
}

func logitLowerIdealIndex(v float64) int {
	switch {
	case v < 0.5:
		return int(math.Floor(math.Log10(v)))
	case v < 0.9:
		return 0
	default:
		return -int(math.Ceil(math.Log10(1 - v)))
	}
}

func logitUpperIdealIndex(v float64) int {
	switch {
	case v <= 0.5:
		return int(math.Ceil(math.Log10(v)))
	case v <= 0.9:
		return 1
	default:
		return -int(math.Floor(math.Log10(1 - v)))
	}
}

func logitIdealTick(index int) float64 {
	switch {
	case index < 0:
		return math.Pow10(index)
	case index > 0:
		return 1 - math.Pow10(-index)
	default:
		return 0.5
	}
}

func logitMinorTicks(bInf, bSup int) []float64 {
	ticks := make([]float64, 0, 8*tickMaxInt(0, bSup-bInf))
	for b := bInf; b < bSup; b++ {
		switch {
		case b < -1:
			base := math.Pow10(b)
			for n := 2; n < 10; n++ {
				ticks = append(ticks, float64(n)*base)
			}
		case b == -1:
			for n := 2; n < 5; n++ {
				ticks = append(ticks, float64(n)/10)
			}
		case b == 0:
			for n := 6; n < 9; n++ {
				ticks = append(ticks, float64(n)/10)
			}
		default:
			base := math.Pow10(-b - 1)
			for n := 9; n >= 2; n-- {
				ticks = append(ticks, 1-float64(n)*base)
			}
		}
	}
	return dedupeTicksSorted(ticks)
}

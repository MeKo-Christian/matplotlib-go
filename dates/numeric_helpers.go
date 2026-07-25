package dates

import "math"

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

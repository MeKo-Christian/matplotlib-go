package core

import (
	"math"
	"math/rand"
	"strings"
	"testing"
)

func strictlyIncreasing(xs []float64) bool {
	for i := 1; i < len(xs); i++ {
		if !(xs[i] > xs[i-1]) {
			return false
		}
	}
	return true
}

func TestLinearLocator_BasicRanges(t *testing.T) {
	cases := [][2]float64{{-1, 1}, {0, 1e-9}, {1, 1e6}, {-1e6, -1}, {2, 2}}
	targets := []int{3, 5, 7}
	for _, c := range cases {
		for _, n := range targets {
			ticks := (LinearLocator{}).Ticks(c[0], c[1], n)
			if len(ticks) == 0 {
				t.Fatalf("no ticks for range %+v", c)
			}
			if !strictlyIncreasing(ticks) {
				t.Fatalf("ticks not strictly increasing: %+v", ticks)
			}
			min, max := c[0], c[1]
			if min > max {
				min, max = max, min
			}
			if ticks[0] > min+1e-12 {
				t.Fatalf("first tick %v > min %v", ticks[0], min)
			}
			if ticks[len(ticks)-1] < max-1e-12 {
				t.Fatalf("last tick %v < max %v", ticks[len(ticks)-1], max)
			}
			// Do not assert exact count band here; coverage and monotonicity suffice.
		}
	}
}

func TestLinearLocator_Property(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 200; i++ {
		a := r.Float64()*2e6 - 1e6
		b := a + (r.Float64()*2e6 + 1e-9)
		n := 2 + int(r.Float64()*8)
		ticks := (LinearLocator{}).Ticks(a, b, n)
		if !strictlyIncreasing(ticks) {
			t.Fatalf("not increasing: %+v", ticks)
		}
		// Coverage
		min, max := a, b
		if min > max {
			min, max = max, min
		}
		if ticks[0] > min+1e-9 {
			t.Fatalf("first > min: %v > %v", ticks[0], min)
		}
		if ticks[len(ticks)-1] < max-1e-9 {
			t.Fatalf("last < max: %v < %v", ticks[len(ticks)-1], max)
		}
	}
}

func TestLogLocator_MajorsMonotone(t *testing.T) {
	bases := []float64{2, 10}
	for _, b := range bases {
		l := LogLocator{Base: b}
		ticks := l.Ticks(1, 1e6, 0)
		if len(ticks) == 0 {
			t.Fatalf("no log ticks for base %v", b)
		}
		if !strictlyIncreasing(ticks) {
			t.Fatalf("log ticks not increasing: %+v", ticks)
		}
		// All ticks should be within [min,max]
		if ticks[0] < 1-1e-12 || ticks[len(ticks)-1] > 1e6+1e-12 {
			t.Fatalf("log ticks out of range: first=%v last=%v", ticks[0], ticks[len(ticks)-1])
		}
	}
}

func TestLogLocator_MinorsBetweenMajors(t *testing.T) {
	l := LogLocator{Base: 10, Minor: true}
	ticks := l.Ticks(1, 1e3, 0)
	if !strictlyIncreasing(ticks) {
		t.Fatalf("log ticks not increasing: %+v", ticks)
	}
	// Must contain the canonical set within [1,1000]
	want := []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000}
	// Build a map for quick lookup with tolerance
	has := func(v float64) bool {
		for _, t := range ticks {
			if math.Abs(t-v) <= 1e-12 {
				return true
			}
		}
		return false
	}
	for _, v := range want {
		if !has(v) {
			t.Fatalf("missing expected tick %v in %+v", v, ticks)
		}
	}
}

func TestScalarFormatter_TrimAndScientific(t *testing.T) {
	f := ScalarFormatter{Prec: 6}
	if got := f.Format(1.0); got != "1" {
		t.Fatalf("Format(1.0)=%q", got)
	}
	if got := f.Format(1.230000); got != "1.23" {
		t.Fatalf("trim zeros: %q", got)
	}
	if got := f.Format(1234567); !strings.Contains(got, "e") {
		t.Fatalf("expected scientific for large: %q", got)
	}
	if got := f.Format(0.0000123); !strings.Contains(got, "e") {
		t.Fatalf("expected scientific for small: %q", got)
	}
}

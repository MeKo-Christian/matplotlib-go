package core

import (
	"math"
	"testing"
)

func approxEqual(a, b, tol float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return math.Abs(a-b) <= tol
}

// newSqrtFuncNorm builds a FuncNorm with a sqrt forward / square inverse over
// [vmin, vmax], the canonical Matplotlib FuncNorm example.
func newSqrtFuncNorm(vmin, vmax float64, clip bool) FuncNorm {
	return FuncNorm{
		Forward: math.Sqrt,
		Reverse: func(v float64) float64 { return v * v },
		VMin:    vmin,
		VMax:    vmax,
		Clip:    clip,
	}
}

func TestFuncNormMapMatchesTransformedNormalization(t *testing.T) {
	n := newSqrtFuncNorm(0, 100, false)
	// forward(vmin)=0, forward(vmax)=10. value=25 -> sqrt=5 -> (5-0)/10 = 0.5.
	if got := n.Map(25); !approxEqual(got, 0.5, 1e-12) {
		t.Fatalf("Map(25) = %v, want 0.5", got)
	}
	if got := n.Map(0); !approxEqual(got, 0, 1e-12) {
		t.Fatalf("Map(0) = %v, want 0", got)
	}
	if got := n.Map(100); !approxEqual(got, 1, 1e-12) {
		t.Fatalf("Map(100) = %v, want 1", got)
	}
}

func TestFuncNormMapOutsideRangeWithoutClip(t *testing.T) {
	n := newSqrtFuncNorm(1, 100, false)
	// Without clip, values outside the range still transform (>1 or <0).
	if got := n.Map(400); !(got > 1) {
		t.Fatalf("Map(400) = %v, want > 1 (no clip)", got)
	}
}

func TestFuncNormMapClampsWithClip(t *testing.T) {
	n := newSqrtFuncNorm(1, 100, true)
	if got := n.Map(400); !approxEqual(got, 1, 1e-12) {
		t.Fatalf("Map(400) clipped = %v, want 1", got)
	}
	if got := n.Map(0); !approxEqual(got, 0, 1e-12) {
		t.Fatalf("Map(0) clipped = %v, want 0", got)
	}
}

func TestFuncNormInverseRoundTrips(t *testing.T) {
	n := newSqrtFuncNorm(0, 100, false)
	for _, v := range []float64{0, 12.5, 25, 64, 100} {
		mapped := n.Map(v)
		inv, ok := n.Inverse(mapped)
		if !ok {
			t.Fatalf("Inverse(%v) reported not invertible", mapped)
		}
		if !approxEqual(inv, v, 1e-9) {
			t.Fatalf("Inverse(Map(%v)) = %v, want %v", v, inv, v)
		}
	}
}

func TestFuncNormAutoscaleUsesTransformDomainFiniteValues(t *testing.T) {
	// vmin/vmax unset (NaN) -> autoscale from finite data in the transform
	// domain. Negative values are outside the sqrt domain (forward = NaN) and
	// must be ignored, like Matplotlib's autoscale_None.
	n := FuncNorm{Forward: math.Sqrt, Reverse: func(v float64) float64 { return v * v }, VMin: math.NaN(), VMax: math.NaN()}
	scaled := n.Autoscale([]float64{-5, 4, 9, 16, math.NaN()}).(FuncNorm)
	if !approxEqual(scaled.VMin, 4, 1e-12) {
		t.Fatalf("autoscale VMin = %v, want 4", scaled.VMin)
	}
	if !approxEqual(scaled.VMax, 16, 1e-12) {
		t.Fatalf("autoscale VMax = %v, want 16", scaled.VMax)
	}
}

func TestFuncNormAutoscaleHonorsExplicitLimits(t *testing.T) {
	n := newSqrtFuncNorm(2, 50, false)
	scaled := n.Autoscale([]float64{1, 2, 3, 100}).(FuncNorm)
	if scaled.VMin != 2 || scaled.VMax != 50 {
		t.Fatalf("autoscale overrode explicit limits: %v..%v", scaled.VMin, scaled.VMax)
	}
}

func TestFuncNormValidate(t *testing.T) {
	if err := (FuncNorm{VMin: 0, VMax: 1}).Validate(); err == nil {
		t.Fatal("expected error when forward/inverse are nil")
	}
	if err := newSqrtFuncNorm(0, 1, false).Validate(); err != nil {
		t.Fatalf("valid FuncNorm rejected: %v", err)
	}
	if err := newSqrtFuncNorm(5, 1, false).Validate(); err == nil {
		t.Fatal("expected error when vmin > vmax")
	}
}

func TestFuncNormNonFiniteInputs(t *testing.T) {
	n := newSqrtFuncNorm(0, 100, false)
	if got := n.Map(math.NaN()); !math.IsNaN(got) {
		t.Fatalf("Map(NaN) = %v, want NaN", got)
	}
	if got := n.Map(math.Inf(1)); !math.IsNaN(got) {
		t.Fatalf("Map(+Inf) = %v, want NaN", got)
	}
}

func TestFuncNormNormNameAndRange(t *testing.T) {
	n := newSqrtFuncNorm(2, 8, false)
	if n.NormName() != "function" {
		t.Fatalf("NormName() = %q, want \"function\"", n.NormName())
	}
	lo, hi := n.Range()
	if lo != 2 || hi != 8 {
		t.Fatalf("Range() = %v..%v, want 2..8", lo, hi)
	}
}

func TestFuncNormSatisfiesScalarNormalizer(t *testing.T) {
	var _ ScalarNormalizer = FuncNorm{}
}

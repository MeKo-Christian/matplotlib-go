package transform

import (
	"math"
	"math/rand"
	"testing"
)

func TestScaleRegistryBuiltins(t *testing.T) {
	names := ScaleNames()
	want := []string{"asinh", "func", "function", "functionlog", "linear", "log", "logit", "symlog"}
	for _, name := range want {
		found := false
		for _, got := range names {
			if got == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("built-in scale %q not registered: %v", name, names)
		}
	}
}

func TestNewScale_LogNormalizesDefaultDomain(t *testing.T) {
	scale, err := NewScale("log")
	if err != nil {
		t.Fatalf("NewScale(log): %v", err)
	}

	logScale, ok := scale.(Log)
	if !ok {
		t.Fatalf("log scale type = %T, want transform.Log", scale)
	}

	minVal, maxVal := logScale.Domain()
	if minVal <= 0 || maxVal <= 0 || minVal == maxVal {
		t.Fatalf("normalized log domain = (%v, %v), want positive non-degenerate range", minVal, maxVal)
	}
}

func TestSymLogScale_RoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	s := NewSymLog(-250, 500, 10, 2, 1.5)
	for i := 0; i < 200; i++ {
		x := -250 + r.Float64()*750
		u := s.Fwd(x)
		xr, ok := s.Inv(u)
		if !ok {
			t.Fatalf("symlog inverse failed for x=%v", x)
		}
		if !approx(x, xr, 1e-9*(1+math.Abs(x))) {
			t.Fatalf("symlog roundtrip mismatch: x=%v xr=%v", x, xr)
		}
	}
}

func TestSymLogTransformMatchesMatplotlibLinscaleAdjustment(t *testing.T) {
	s := NewSymLog(-1000, 1000, 10, 2, 1.5)
	adjusted := 1.5 / (1 - math.Pow(10, -1))
	cases := []struct {
		x    float64
		want float64
	}{
		{x: -20, want: -2 * (adjusted + 1)},
		{x: -2, want: -2 * adjusted},
		{x: 0, want: 0},
		{x: 2, want: 2 * adjusted},
		{x: 20, want: 2 * (adjusted + 1)},
	}

	for _, tc := range cases {
		got, ok := s.transform(tc.x)
		if !ok {
			t.Fatalf("transform(%v) failed", tc.x)
		}
		if !approx(got, tc.want, 1e-12) {
			t.Fatalf("transform(%v) = %v, want Matplotlib %v", tc.x, got, tc.want)
		}
		inv, ok := s.inverse(got)
		if !ok {
			t.Fatalf("inverse(%v) failed", got)
		}
		if !approx(inv, tc.x, 1e-12*(1+math.Abs(tc.x))) {
			t.Fatalf("inverse(transform(%v)) = %v", tc.x, inv)
		}
	}
}

func TestAsinhScale_RoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(12))
	s := NewAsinh(-25, 40, 0.5)
	for i := 0; i < 200; i++ {
		x := -25 + r.Float64()*65
		u := s.Fwd(x)
		xr, ok := s.Inv(u)
		if !ok {
			t.Fatalf("asinh inverse failed for x=%v", x)
		}
		if !approx(x, xr, 1e-9*(1+math.Abs(x))) {
			t.Fatalf("asinh roundtrip mismatch: x=%v xr=%v", x, xr)
		}
	}
}

func TestLogitScale_RoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(13))
	s := NewLogit(0.02, 0.98, NonPositiveMask, 1e-6)
	for i := 0; i < 200; i++ {
		x := 0.02 + r.Float64()*0.96
		u := s.Fwd(x)
		xr, ok := s.Inv(u)
		if !ok {
			t.Fatalf("logit inverse failed for x=%v", x)
		}
		if !approx(x, xr, 1e-9*(1+math.Abs(x))) {
			t.Fatalf("logit roundtrip mismatch: x=%v xr=%v", x, xr)
		}
	}
}

func TestLogScale_NonPositiveClip(t *testing.T) {
	s := Log{Min: 1, Max: 100, Base: 10, NonPositive: NonPositiveClip}
	// Non-positive input pins the log-space output to the -1000 sentinel
	// (matplotlib scale.py), so the normalized coordinate is finite but lands
	// far below the [0,1] axis span and is viewport-clipped.
	got := s.Fwd(-5)
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("clipped log forward should stay finite, got %v", got)
	}
	if got >= 0 {
		t.Fatalf("clipped log forward = %v, want far below 0 (off-axis)", got)
	}
	// mask mode masks the same input to NaN instead.
	mask := Log{Min: 1, Max: 100, Base: 10, NonPositive: NonPositiveMask}
	if masked := mask.Fwd(-5); !math.IsNaN(masked) {
		t.Fatalf("masked log forward should be NaN, got %v", masked)
	}
}

func TestLog_DefaultNonPositiveIsClip(t *testing.T) {
	// Matplotlib's LogScale defaults to nonpositive='clip'. Both the direct
	// constructor and the registry factory (no explicit WithScaleNonPositive)
	// must clip, not mask.
	if got := NewLog(1, 100, 10).Fwd(-5); math.IsNaN(got) {
		t.Fatalf("NewLog default forward of non-positive = NaN, want finite clip")
	}
	scale, err := NewScale("log", WithScaleDomain(1, 100), WithScaleBase(10))
	if err != nil {
		t.Fatalf("NewScale(log): %v", err)
	}
	logScale, ok := scale.(Log)
	if !ok {
		t.Fatalf("scale type = %T, want Log", scale)
	}
	if logScale.NonPositive != NonPositiveClip {
		t.Fatalf("default log NonPositive = %q, want %q", logScale.NonPositive, NonPositiveClip)
	}
	if got := logScale.Fwd(-5); math.IsNaN(got) {
		t.Fatalf("default log factory forward of non-positive = NaN, want finite clip")
	}
}

func TestLogScale_NormalizesNonPositiveDomain(t *testing.T) {
	scale, err := NewScale(
		"log",
		WithScaleDomain(-5, 100),
		WithScaleBase(10),
		WithScaleNonPositive(NonPositiveClip),
	)
	if err != nil {
		t.Fatalf("NewScale(log): %v", err)
	}
	logScale, ok := scale.(Log)
	if !ok {
		t.Fatalf("scale type = %T, want Log", scale)
	}
	minVal, maxVal := logScale.Domain()
	if minVal <= 0 || maxVal <= minVal {
		t.Fatalf("normalized log domain = (%v, %v), want positive increasing domain", minVal, maxVal)
	}
	if got := logScale.Fwd(10); math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("normalized log forward should stay finite, got %v", got)
	}
}

func TestLogitScale_NonPositiveHandling(t *testing.T) {
	mask := NewLogit(0.01, 0.99, NonPositiveMask, 1e-6)
	if got := mask.Fwd(-0.5); !math.IsNaN(got) {
		t.Fatalf("masked logit forward should be NaN, got %v", got)
	}

	clip := NewLogit(0.01, 0.99, NonPositiveClip, 1e-6)
	// Below 0 clips the logit output to -1000, above 1 to +1000 (matplotlib
	// scale.py), so normalized coordinates land far below/above the axis.
	if got := clip.Fwd(-0.5); math.IsNaN(got) || math.IsInf(got, 0) || got >= 0 {
		t.Fatalf("clipped logit forward below 0 = %v, want finite and off the bottom", got)
	}
	if got := clip.Fwd(1.5); math.IsNaN(got) || math.IsInf(got, 0) || got <= 1 {
		t.Fatalf("clipped logit forward above 1 = %v, want finite and off the top", got)
	}
}

func TestFunctionScale(t *testing.T) {
	scale, err := NewScale(
		"function",
		WithScaleDomain(-3, 3),
		WithScaleFunctions(
			func(x float64) float64 { return x * x * x },
			func(y float64) (float64, bool) { return math.Cbrt(y), true },
		),
	)
	if err != nil {
		t.Fatalf("NewScale(function): %v", err)
	}

	for _, x := range []float64{-3, -1, -0.25, 0, 0.25, 1, 3} {
		u := scale.Fwd(x)
		xr, ok := scale.Inv(u)
		if !ok {
			t.Fatalf("function inverse failed for x=%v", x)
		}
		if !approx(x, xr, 1e-9*(1+math.Abs(x))) {
			t.Fatalf("function roundtrip mismatch: x=%v xr=%v", x, xr)
		}
	}
}

func TestFunctionLogScale(t *testing.T) {
	scale, err := NewScale(
		"functionlog",
		WithScaleDomain(1, 100),
		WithScaleBase(10),
		WithScaleFunctions(
			func(x float64) float64 { return x * x },
			func(y float64) (float64, bool) { return math.Sqrt(y), true },
		),
	)
	if err != nil {
		t.Fatalf("NewScale(functionlog): %v", err)
	}
	if _, ok := scale.(FuncLogScale); !ok {
		t.Fatalf("functionlog scale type = %T, want FuncLogScale", scale)
	}

	for _, x := range []float64{1, 2, 10, 100} {
		u := scale.Fwd(x)
		xr, ok := scale.Inv(u)
		if !ok {
			t.Fatalf("functionlog inverse failed for x=%v", x)
		}
		if !approx(x, xr, 1e-9*(1+math.Abs(x))) {
			t.Fatalf("functionlog roundtrip mismatch: x=%v xr=%v", x, xr)
		}
	}
}

func TestScaleRegistryRegister(t *testing.T) {
	r := NewScaleRegistry()
	if err := r.Register("custom", func(opts ScaleOptions) (Scale, error) {
		return NewLinear(opts.DomainMin, opts.DomainMax), nil
	}); err != nil {
		t.Fatalf("Register(custom): %v", err)
	}

	scale, err := r.New("custom", WithScaleDomain(3, 7))
	if err != nil {
		t.Fatalf("New(custom): %v", err)
	}

	minVal, maxVal := scale.Domain()
	if minVal != 3 || maxVal != 7 {
		t.Fatalf("custom scale domain = (%v, %v), want (3, 7)", minVal, maxVal)
	}

	if err := r.Register("custom", func(opts ScaleOptions) (Scale, error) {
		return NewLinear(opts.DomainMin, opts.DomainMax), nil
	}); err == nil {
		t.Fatal("duplicate Register(custom) should fail")
	}
}

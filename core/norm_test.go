package core

import (
	"math"
	"testing"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestNormalizeMapsLinearlyAndHonorsClip(t *testing.T) {
	norm := Normalize{VMin: -1, VMax: 1}
	if got, want := norm.Map(-2), -0.5; !floatApprox(got, want, 1e-12) {
		t.Fatalf("Normalize.Map(-2) = %v, want %v", got, want)
	}
	if got, want := norm.Map(0), 0.5; !floatApprox(got, want, 1e-12) {
		t.Fatalf("Normalize.Map(0) = %v, want %v", got, want)
	}
	if got, want := norm.Map(2), 1.5; !floatApprox(got, want, 1e-12) {
		t.Fatalf("Normalize.Map(2) = %v, want %v", got, want)
	}

	clipped := Normalize{VMin: -1, VMax: 1, Clip: true}
	if got := clipped.Map(-2); got != 0 {
		t.Fatalf("clipped under value = %v, want 0", got)
	}
	if got := clipped.Map(2); got != 1 {
		t.Fatalf("clipped over value = %v, want 1", got)
	}
}

func TestNormalizeAutoscaleAndValidation(t *testing.T) {
	norm := (Normalize{VMin: math.NaN(), VMax: math.NaN()}).Autoscale([]float64{math.NaN(), 2, -4, 10}).(Normalize)
	if norm.VMin != -4 || norm.VMax != 10 {
		t.Fatalf("autoscaled range = %v..%v, want -4..10", norm.VMin, norm.VMax)
	}
	if got := (Normalize{VMin: 3, VMax: 3}).Map(4); got != 0 {
		t.Fatalf("equal-range Normalize.Map = %v, want 0", got)
	}
	if err := (Normalize{VMin: 2, VMax: 1}).Validate(); err == nil {
		t.Fatal("expected reversed range validation error")
	}
}

func TestLogNormMapsPositiveDomain(t *testing.T) {
	norm := LogNorm{VMin: 1, VMax: 100}
	if got := norm.Map(1); !floatApprox(got, 0, 1e-12) {
		t.Fatalf("LogNorm.Map(1) = %v, want 0", got)
	}
	if got := norm.Map(10); !floatApprox(got, 0.5, 1e-12) {
		t.Fatalf("LogNorm.Map(10) = %v, want 0.5", got)
	}
	if got := norm.Map(100); !floatApprox(got, 1, 1e-12) {
		t.Fatalf("LogNorm.Map(100) = %v, want 1", got)
	}
	if got := norm.Map(0); !math.IsNaN(got) {
		t.Fatalf("LogNorm.Map(0) = %v, want NaN", got)
	}
}

func TestDivergingAndNonlinearNorms(t *testing.T) {
	twoSlope := TwoSlopeNorm{VMin: -4, VCenter: 0, VMax: 2}
	if got := twoSlope.Map(-2); !floatApprox(got, 0.25, 1e-12) {
		t.Fatalf("TwoSlopeNorm.Map(-2) = %v, want 0.25", got)
	}
	if got := twoSlope.Map(1); !floatApprox(got, 0.75, 1e-12) {
		t.Fatalf("TwoSlopeNorm.Map(1) = %v, want 0.75", got)
	}

	centered := CenteredNorm{VCenter: 0, HalfRange: 4}
	if got := centered.Map(-2); !floatApprox(got, 0.25, 1e-12) {
		t.Fatalf("CenteredNorm.Map(-2) = %v, want 0.25", got)
	}
	if got := centered.Map(4); !floatApprox(got, 1, 1e-12) {
		t.Fatalf("CenteredNorm.Map(4) = %v, want 1", got)
	}

	power := PowerNorm{Gamma: 2, VMin: 0, VMax: 2}
	if got := power.Map(1); !floatApprox(got, 0.25, 1e-12) {
		t.Fatalf("PowerNorm.Map(1) = %v, want 0.25", got)
	}
	if got := power.Map(-1); !floatApprox(got, -0.5, 1e-12) {
		t.Fatalf("PowerNorm.Map(-1) = %v, want -0.5", got)
	}
}

func TestSymLogNormMapsSymmetricallyAroundZero(t *testing.T) {
	norm := SymLogNorm{VMin: -100, VMax: 100, LinThresh: 1, LinScale: 1, Base: 10}
	if got := norm.Map(-100); !floatApprox(got, 0, 1e-12) {
		t.Fatalf("SymLogNorm.Map(-100) = %v, want 0", got)
	}
	if got := norm.Map(0); !floatApprox(got, 0.5, 1e-12) {
		t.Fatalf("SymLogNorm.Map(0) = %v, want 0.5", got)
	}
	if got := norm.Map(100); !floatApprox(got, 1, 1e-12) {
		t.Fatalf("SymLogNorm.Map(100) = %v, want 1", got)
	}
}

func TestAsinhNormMapsSmoothSymmetricDomain(t *testing.T) {
	norm := AsinhNorm{LinearWidth: 2, VMin: -10, VMax: 10}
	if got := norm.Map(-10); !floatApprox(got, 0, 1e-12) {
		t.Fatalf("AsinhNorm.Map(-10) = %v, want 0", got)
	}
	if got := norm.Map(0); !floatApprox(got, 0.5, 1e-12) {
		t.Fatalf("AsinhNorm.Map(0) = %v, want 0.5", got)
	}
	if got := norm.Map(10); !floatApprox(got, 1, 1e-12) {
		t.Fatalf("AsinhNorm.Map(10) = %v, want 1", got)
	}

	value, ok := norm.Inverse(0.5)
	if !ok {
		t.Fatal("AsinhNorm.Inverse(0.5) returned !ok")
	}
	if !floatApprox(value, 0, 1e-12) {
		t.Fatalf("AsinhNorm.Inverse(0.5) = %v, want 0", value)
	}
}

func TestAsinhNormAutoscaleClipAndValidation(t *testing.T) {
	norm := (AsinhNorm{LinearWidth: 0.5, VMin: math.NaN(), VMax: math.NaN()}).Autoscale([]float64{math.NaN(), -4, 9}).(AsinhNorm)
	if norm.VMin != -4 || norm.VMax != 9 {
		t.Fatalf("AsinhNorm autoscaled range = %v..%v, want -4..9", norm.VMin, norm.VMax)
	}
	if got := (AsinhNorm{VMin: -1, VMax: 1, Clip: true}).Map(10); got != 1 {
		t.Fatalf("clipped AsinhNorm over value = %v, want 1", got)
	}
	if err := (AsinhNorm{LinearWidth: -1, VMin: -1, VMax: 1}).Validate(); err == nil {
		t.Fatal("expected linear_width validation error")
	}
}

func TestCenteredAndNonlinearNormsMatchUpstreamAudit(t *testing.T) {
	centered := CenteredNorm{VCenter: 0, HalfRange: 4}
	centeredCases := map[float64]float64{-2: 0.25, 0: 0.5, 4: 1, 8: 1.5}
	for value, want := range centeredCases {
		if got := centered.Map(value); !floatApprox(got, want, 1e-12) {
			t.Fatalf("CenteredNorm.Map(%v) = %v, want %v", value, got, want)
		}
	}
	if got := (CenteredNorm{VCenter: 0, HalfRange: 4, Clip: true}).Map(8); got != 1 {
		t.Fatalf("clipped CenteredNorm.Map(8) = %v, want 1", got)
	}

	twoSlope := TwoSlopeNorm{VMin: -4, VCenter: 0, VMax: 2}
	twoSlopeCases := map[float64]float64{-4: 0, -2: 0.25, 0: 0.5, 1: 0.75, 2: 1}
	for value, want := range twoSlopeCases {
		if got := twoSlope.Map(value); !floatApprox(got, want, 1e-12) {
			t.Fatalf("TwoSlopeNorm.Map(%v) = %v, want %v", value, got, want)
		}
	}
	for value, want := range map[float64]float64{0.25: -2, 0.5: 0, 0.75: 1} {
		got, ok := twoSlope.Inverse(value)
		if !ok || !floatApprox(got, want, 1e-12) {
			t.Fatalf("TwoSlopeNorm.Inverse(%v) = %v ok=%v, want %v ok=true", value, got, ok, want)
		}
	}
	// Out-of-range values extrapolate to ±inf so the colormap routes them to the
	// under/over colors, matching np.interp(..., left=-inf, right=inf) in
	// matplotlib's TwoSlopeNorm.__call__/inverse.
	if got := twoSlope.Map(-5); !math.IsInf(got, -1) {
		t.Fatalf("TwoSlopeNorm.Map(below vmin) = %v, want -inf", got)
	}
	if got := twoSlope.Map(3); !math.IsInf(got, 1) {
		t.Fatalf("TwoSlopeNorm.Map(above vmax) = %v, want +inf", got)
	}
	if got, ok := twoSlope.Inverse(-0.1); !ok || !math.IsInf(got, -1) {
		t.Fatalf("TwoSlopeNorm.Inverse(<0) = %v ok=%v, want -inf ok=true", got, ok)
	}
	if got, ok := twoSlope.Inverse(1.1); !ok || !math.IsInf(got, 1) {
		t.Fatalf("TwoSlopeNorm.Inverse(>1) = %v ok=%v, want +inf ok=true", got, ok)
	}
	autoscaled := (TwoSlopeNorm{VCenter: 0, VMin: math.NaN(), VMax: math.NaN()}).Autoscale([]float64{2, 4}).(TwoSlopeNorm)
	if autoscaled.VMin != -4 || autoscaled.VMax != 4 {
		t.Fatalf("TwoSlopeNorm autoscaled range = %v..%v, want -4..4", autoscaled.VMin, autoscaled.VMax)
	}
	for _, norm := range []TwoSlopeNorm{
		{VMin: 0, VCenter: 0, VMax: 1},
		{VMin: -1, VCenter: 1, VMax: 1},
	} {
		if err := norm.Validate(); err == nil {
			t.Fatalf("TwoSlopeNorm{%v,%v,%v}.Validate() = nil, want ascending-order error", norm.VMin, norm.VCenter, norm.VMax)
		}
	}

	power := PowerNorm{Gamma: 2, VMin: 0, VMax: 2}
	if got := power.Map(-1); !floatApprox(got, -0.5, 1e-12) {
		t.Fatalf("PowerNorm.Map(-1) = %v, want -0.5", got)
	}
	if got := power.Map(4); !floatApprox(got, 4, 1e-12) {
		t.Fatalf("PowerNorm.Map(4) = %v, want 4", got)
	}
	if got := (PowerNorm{Gamma: 2, VMin: 0, VMax: 2, Clip: true}).Map(4); got != 1 {
		t.Fatalf("clipped PowerNorm.Map(4) = %v, want 1", got)
	}

	for _, tc := range []struct {
		name  string
		norm  ScalarNormalizer
		value float64
	}{
		{"symlog", SymLogNorm{VMin: -100, VMax: 100, LinThresh: 1, LinScale: 1, Base: 10}, 0.75},
		{"asinh", AsinhNorm{LinearWidth: 2, VMin: -10, VMax: 10}, 0.75},
	} {
		mappedInverse, ok := tc.norm.Inverse(tc.value)
		if !ok {
			t.Fatalf("%s inverse(%v) returned !ok", tc.name, tc.value)
		}
		if got := tc.norm.Map(mappedInverse); !floatApprox(got, tc.value, 1e-12) {
			t.Fatalf("%s Map(Inverse(%v)) = %v, want %v", tc.name, tc.value, got, tc.value)
		}
	}
}

func TestBoundaryNormReturnsDiscreteColorIndexes(t *testing.T) {
	norm := BoundaryNorm{Boundaries: []float64{0, 10, 20}, NColors: 5}
	if got := norm.Index(-1); got != -1 {
		t.Fatalf("BoundaryNorm.Index(-1) = %d, want -1", got)
	}
	if got := norm.Index(5); got != 0 {
		t.Fatalf("BoundaryNorm.Index(5) = %d, want 0", got)
	}
	if got := norm.Index(15); got != 4 {
		t.Fatalf("BoundaryNorm.Index(15) = %d, want 4", got)
	}
	if got := norm.Index(20); got != 5 {
		t.Fatalf("BoundaryNorm.Index(20) = %d, want 5", got)
	}
	if got := norm.Map(15); !floatApprox(got, 1, 1e-12) {
		t.Fatalf("BoundaryNorm.Map(15) = %v, want 1", got)
	}
}

func TestBoundaryNormMatchesUpstreamIndexAudit(t *testing.T) {
	tests := []struct {
		name   string
		source string
		norm   BoundaryNorm
		values map[float64]int
	}{
		{
			name:   "base bins use under and over sentinels",
			source: "colors.py:BoundaryNorm.__call__: under=-1 over=ncolors when clip is false",
			norm:   BoundaryNorm{Boundaries: []float64{0, 1, 2, 3}, NColors: 3},
			values: map[float64]int{-0.1: -1, 0: 0, 0.5: 0, 1: 1, 2.9: 2, 3: 3},
		},
		{
			name:   "clip clamps out of range values",
			source: "colors.py:BoundaryNorm.__call__: clip clamps to first and last color",
			norm:   BoundaryNorm{Boundaries: []float64{0, 10, 20}, NColors: 5, Clip: true},
			values: map[float64]int{-1: 0, 0: 0, 5: 0, 15: 4, 20: 4, 25: 4},
		},
		{
			name:   "more colors than regions stretches indexes",
			source: "colors.py:BoundaryNorm.__call__: stretch region indexes over color range",
			norm:   BoundaryNorm{Boundaries: []float64{0, 10, 20}, NColors: 5},
			values: map[float64]int{-1: -1, 5: 0, 15: 4, 20: 5},
		},
		{
			name:   "extend min offsets interior bins",
			source: "colors.py:BoundaryNorm.__init__: extend min increments offset",
			norm:   BoundaryNorm{Boundaries: []float64{0, 1, 2}, NColors: 3, Extend: "min"},
			values: map[float64]int{-0.1: -1, 0.5: 1, 1.5: 2, 2: 3},
		},
		{
			name:   "extend both keeps under and over sentinels",
			source: "colors.py:BoundaryNorm.__init__: extend both adds min and max regions",
			norm:   BoundaryNorm{Boundaries: []float64{0, 1, 2}, NColors: 4, Extend: "both"},
			values: map[float64]int{-0.1: -1, 0.5: 1, 1.5: 2, 2: 4},
		},
	}

	for _, tc := range tests {
		if tc.source == "" {
			t.Fatalf("%s missing upstream source", tc.name)
		}
		if err := tc.norm.Validate(); err != nil {
			t.Fatalf("%s Validate() = %v", tc.name, err)
		}
		for value, want := range tc.values {
			if got := tc.norm.Index(value); got != want {
				t.Fatalf("%s BoundaryNorm.Index(%v) = %d, want %d", tc.name, value, got, want)
			}
		}
	}

	if err := (BoundaryNorm{Boundaries: []float64{0, 1, 2}, NColors: 3, Clip: true, Extend: "both"}).Validate(); err == nil {
		t.Fatal("clip=true with extend should fail validation")
	}
	if err := (BoundaryNorm{Boundaries: []float64{0, 1, 2}, NColors: 2, Extend: "both"}).Validate(); err == nil {
		t.Fatal("ncolors smaller than extended region count should fail validation")
	}
	if _, ok := (BoundaryNorm{Boundaries: []float64{0, 1}, NColors: 1}).Inverse(0); ok {
		t.Fatal("BoundaryNorm.Inverse returned ok=true, want non-invertible")
	}
}

func TestResolveScalarMapRejectsNormWithVMinVMax(t *testing.T) {
	vmin := 0.0
	_, err := ResolveScalarMapValues([]float64{1, 2, 3}, ScalarMapConfig{
		Norm: Normalize{VMin: 1, VMax: 3},
		VMin: &vmin,
	})
	if err == nil {
		t.Fatal("expected norm/vmin conflict validation error")
	}
}

func TestPlotOptionsScalarMapConfig(t *testing.T) {
	name := "plasma"
	vmin, vmax := -2.0, 5.0
	norm := PowerNorm{Gamma: 2, VMin: vmin, VMax: vmax}
	options := PlotOptions{
		Colormap: &name,
		Norm:     norm,
		VMin:     &vmin,
		VMax:     &vmax,
	}

	got := options.ScalarMapConfig()
	if got.Colormap != name || got.VMin != &vmin || got.VMax != &vmax {
		t.Fatalf("ScalarMapConfig() = %+v, want colormap and limit pointers from PlotOptions", got)
	}
	if gotNorm, ok := got.Norm.(PowerNorm); !ok || gotNorm != norm {
		t.Fatalf("ScalarMapConfig().Norm = %#v, want %#v", got.Norm, norm)
	}
}

func TestScalarMapInfoRoutesBadUnderAndOverColorsThroughColormap(t *testing.T) {
	cmapName := "scalar-map-test-bounds"
	bad := render.Color{R: 1, A: 1}
	under := render.Color{G: 1, A: 1}
	over := render.Color{B: 1, A: 1}
	matcolor.RegisterColormap(cmapName, matcolor.NewColormap(cmapName, []matcolor.ColorStop{
		{Pos: 0, Color: render.Color{A: 1}},
		{Pos: 1, Color: render.Color{R: 1, G: 1, B: 1, A: 1}},
	}).WithBad(bad).WithUnder(under).WithOver(over))

	mapping := ScalarMapInfo{
		Colormap: cmapName,
		Norm:     Normalize{VMin: 0, VMax: 1},
	}.Resolved()
	if got := mapping.Color(math.NaN(), 1); got != bad {
		t.Fatalf("bad color = %+v, want %+v", got, bad)
	}
	if got := mapping.Color(-1, 1); got != under {
		t.Fatalf("under color = %+v, want %+v", got, under)
	}
	if got := mapping.Color(2, 1); got != over {
		t.Fatalf("over color = %+v, want %+v", got, over)
	}
}

func TestScalarMapInfoResolvedCachesColormapLookup(t *testing.T) {
	mapping := ScalarMapInfo{
		Colormap: "viridis_r",
		Norm:     Normalize{VMin: 0, VMax: 1},
	}.Resolved()

	if got, want := mapping.resolvedColormapName, "viridis_r"; got != want {
		t.Fatalf("resolvedColormapName = %q, want %q", got, want)
	}
	if got, want := mapping.resolvedColormap.Name(), "viridis_r"; got != want {
		t.Fatalf("resolvedColormap.Name() = %q, want %q", got, want)
	}

	allocs := testing.AllocsPerRun(1000, func() {
		_ = mapping.Color(0.25, 1)
	})
	if allocs != 0 {
		t.Fatalf("pre-resolved ScalarMapInfo.Color allocs/run = %v, want 0", allocs)
	}
}

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

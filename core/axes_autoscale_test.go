package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestAutoScaleSingleOriginPointMatchesMatplotlib(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())

	ax.Plot([]float64{0}, []float64{0})

	assertScaleDomain(
		t,
		ax.XScale,
		-0.05500000000000001,
		0.05500000000000001,
		"single origin x",
	)
	assertScaleDomain(
		t,
		ax.YScale,
		-0.05500000000000001,
		0.05500000000000001,
		"single origin y",
	)
}

func TestAutoScaleScatterAtOriginHasExplicitDataBounds(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.Add(&Scatter2D{XY: []geom.Pt{{}}})

	ax.AutoScale(defaultAutoScaleMargin)

	assertScaleDomain(t, ax.XScale, -0.05500000000000001, 0.05500000000000001, "scatter x")
	assertScaleDomain(t, ax.YScale, -0.05500000000000001, 0.05500000000000001, "scatter y")
}

func TestAutoScaleLogMarginsMatchMatplotlib(t *testing.T) {
	tests := []struct {
		name    string
		x       []float64
		wantMin float64
		wantMax float64
	}{
		{
			name:    "positive range",
			x:       []float64{1, 100},
			wantMin: 0.7943282347242815,
			wantMax: 125.89254117941675,
		},
		{
			name:    "mixed values use smallest positive",
			x:       []float64{-10, 0.01, 100},
			wantMin: 0.00630957344480193,
			wantMax: 158.48931924611142,
		},
		{
			name:    "single point expands to adjacent decades",
			x:       []float64{10},
			wantMin: 0.7943282347242815,
			wantMax: 125.89254117941675,
		},
		{
			name:    "nonpositive data uses default log domain",
			x:       []float64{-10, -1},
			wantMin: 0.8912509381337456,
			wantMax: 11.220184543019636,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fig := NewFigure(800, 600)
			ax := fig.AddAxes(unitRect())
			if err := ax.SetXScale("log"); err != nil {
				t.Fatalf("SetXScale(log): %v", err)
			}
			y := make([]float64, len(test.x))
			for i := range y {
				y[i] = float64(i)
			}

			ax.Plot(test.x, y)

			assertScaleDomain(t, ax.XScale, test.wantMin, test.wantMax, test.name)
		})
	}
}

func TestAutoScaleSymLogMarginIsAppliedInTransformSpace(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	if err := ax.SetXScale(
		"symlog",
		transform.WithScaleBase(10),
		transform.WithScaleLinThresh(1),
		transform.WithScaleLinearScale(1),
	); err != nil {
		t.Fatalf("SetXScale(symlog): %v", err)
	}

	ax.Plot([]float64{-100, 10}, []float64{0, 1})

	assertScaleDomain(
		t,
		ax.XScale,
		-182.43623925784647,
		18.243623925784647,
		"symlog transform-space margin",
	)
}

func TestAutoScaleSingleNonzeroPointUsesLocatorNonsingularExpansion(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())

	ax.Plot([]float64{2}, []float64{2})

	assertScaleDomain(t, ax.XScale, 1.89, 2.1100000000000003, "single nonzero x")
	assertScaleDomain(t, ax.YScale, 1.89, 2.1100000000000003, "single nonzero y")
}

func TestAutoScaleTinyPointUsesMatplotlibFloatNormalThreshold(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())

	ax.Plot([]float64{1e-300}, []float64{1e-300})

	assertScaleDomain(t, ax.XScale, -0.05500000000000001, 0.05500000000000001, "tiny x")
	assertScaleDomain(t, ax.YScale, -0.05500000000000001, 0.05500000000000001, "tiny y")
}

func assertScaleDomain(
	t *testing.T,
	scale transform.Scale,
	wantMin, wantMax float64,
	context string,
) {
	t.Helper()
	gotMin, gotMax := scale.Domain()
	const tolerance = 1e-12
	if math.Abs(gotMin-wantMin) > tolerance || math.Abs(gotMax-wantMax) > tolerance {
		t.Fatalf(
			"%s domain = (%0.17g, %0.17g), want (%0.17g, %0.17g)",
			context,
			gotMin,
			gotMax,
			wantMin,
			wantMax,
		)
	}
}

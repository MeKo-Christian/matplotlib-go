package colorbar_composition

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestConstrainedColorbarLayoutMatchesMatplotlib3109(t *testing.T) {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("agg.New: %v", err)
	}
	core.DrawFigure(fig, r)

	if len(fig.Children) < 2 {
		t.Fatalf("expected parent axes and colorbar, got %d axes", len(fig.Children))
	}
	parent := fig.Children[0]
	colorbar := fig.Children[1]

	assertRectApprox(t, "parent display", parent.DisplayRect(), geom.Rect{
		Min: geom.Pt{X: 51.06977777777778, Y: 47.44477777777778},
		Max: geom.Pt{X: 846.4725805555556, Y: 673.4996666666666},
	}, 1e-6)
	assertRectApprox(t, "colorbar original slot", fractionToPixel(colorbar.RectFraction), geom.Rect{
		Min: geom.Pt{X: 899.4302206944444, Y: 47.44477777777778},
		Max: geom.Pt{X: 1018.740641111111, Y: 673.4996666666666},
	}, 1e-6)
	assertRectApprox(t, "colorbar active display", colorbar.DisplayRect(), geom.Rect{
		Min: geom.Pt{X: 899.4302206944444, Y: 47.44477777777778},
		Max: geom.Pt{X: 930.7329651388889, Y: 673.4996666666666},
	}, 1e-6)
}

func fractionToPixel(r geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: r.Min.X * Width, Y: r.Min.Y * Height},
		Max: geom.Pt{X: r.Max.X * Width, Y: r.Max.Y * Height},
	}
}

func assertRectApprox(t *testing.T, label string, got, want geom.Rect, tol float64) {
	t.Helper()
	if !approx(got.Min.X, want.Min.X, tol) ||
		!approx(got.Min.Y, want.Min.Y, tol) ||
		!approx(got.Max.X, want.Max.X, tol) ||
		!approx(got.Max.Y, want.Max.Y, tol) {
		t.Fatalf("%s = %+v, want %+v", label, got, want)
	}
}

func approx(got, want, tol float64) bool {
	return math.Abs(got-want) <= tol
}

package colorbar_horizontal_ticks

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func TestPlotUsesDirectAxesGeometry(t *testing.T) {
	fig := Plot()
	if len(fig.Children) < 2 {
		t.Fatalf("figure axes = %d, want plot axes plus colorbar", len(fig.Children))
	}

	got := fig.Children[0].RectFraction
	want := geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.376},
		Max: geom.Pt{X: 0.90, Y: 0.88},
	}
	if !rectApprox(got, want, 1e-12) {
		t.Fatalf("plot axes rect = %+v, want direct-add-axes colorbar rect %+v", got, want)
	}
}

func rectApprox(got, want geom.Rect, tol float64) bool {
	return floatApprox(got.Min.X, want.Min.X, tol) &&
		floatApprox(got.Min.Y, want.Min.Y, tol) &&
		floatApprox(got.Max.X, want.Max.X, tol) &&
		floatApprox(got.Max.Y, want.Max.Y, tol)
}

func floatApprox(got, want, tol float64) bool {
	if got > want {
		return got-want <= tol
	}
	return want-got <= tol
}

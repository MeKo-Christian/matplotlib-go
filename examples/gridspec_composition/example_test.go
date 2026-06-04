package gridspec_composition

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestSubFigureInsetMatchesMatplotlib3109(t *testing.T) {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("agg.New: %v", err)
	}
	core.DrawFigure(fig, r)

	if len(fig.Children) < 4 {
		t.Fatalf("expected four axes, got %d", len(fig.Children))
	}
	assertRectClose(t, fig.Children[3].DisplayRect(), geom.Rect{
		Min: geom.Pt{X: 680, Y: 35.2},
		Max: geom.Pt{X: 928, Y: 281.6},
	}, 1e-9)
}

func assertRectClose(t *testing.T, got, want geom.Rect, tol float64) {
	t.Helper()
	if math.Abs(got.Min.X-want.Min.X) > tol ||
		math.Abs(got.Min.Y-want.Min.Y) > tol ||
		math.Abs(got.Max.X-want.Max.X) > tol ||
		math.Abs(got.Max.Y-want.Max.Y) > tol {
		t.Fatalf("rect = %+v, want %+v", got, want)
	}
}

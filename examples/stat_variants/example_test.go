package stat_variants

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

func TestPlotPlacesStatGridsBelowPatches(t *testing.T) {
	fig := Plot()
	seen := 0
	for _, ax := range fig.Children {
		for _, art := range ax.Artists {
			grid, ok := art.(*core.Grid)
			if !ok {
				continue
			}
			seen++
			if grid.Z() >= 1.0 {
				t.Fatalf("grid zorder = %v, want below default patch zorder 1.0", grid.Z())
			}
		}
	}
	if seen != 4 {
		t.Fatalf("stat grid count = %d, want 4", seen)
	}
}

func TestPlotUsesMatplotlibFixtureAxesRects(t *testing.T) {
	fig := Plot()
	want := []geom.Rect{
		{Min: geom.Pt{X: 0.08, Y: 0.585}, Max: geom.Pt{X: 0.475, Y: 0.93}},
		{Min: geom.Pt{X: 0.575, Y: 0.585}, Max: geom.Pt{X: 0.97, Y: 0.93}},
		{Min: geom.Pt{X: 0.08, Y: 0.10}, Max: geom.Pt{X: 0.475, Y: 0.445}},
		{Min: geom.Pt{X: 0.575, Y: 0.10}, Max: geom.Pt{X: 0.97, Y: 0.445}},
	}
	if got := len(fig.Children); got != len(want) {
		t.Fatalf("axes count = %d, want %d", got, len(want))
	}
	for i, wantRect := range want {
		if !rectApprox(fig.Children[i].RectFraction, wantRect, 1e-12) {
			t.Fatalf("axes %d rect = %+v, want %+v", i, fig.Children[i].RectFraction, wantRect)
		}
	}
}

func rectApprox(got, want geom.Rect, tol float64) bool {
	return math.Abs(got.Min.X-want.Min.X) <= tol &&
		math.Abs(got.Min.Y-want.Min.Y) <= tol &&
		math.Abs(got.Max.X-want.Max.X) <= tol &&
		math.Abs(got.Max.Y-want.Max.Y) <= tol
}

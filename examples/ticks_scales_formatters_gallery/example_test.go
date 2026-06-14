package ticks_scales_formatters_gallery

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestCategoryInsetUsesMatplotlibFigureFractionRect(t *testing.T) {
	fig := Plot()
	if len(fig.Children) <= 5 {
		t.Fatalf("gallery axes count = %d, want category inset at index 5", len(fig.Children))
	}
	got := fig.Children[5].RectFraction
	want := geom.Rect{
		Min: geom.Pt{X: 0.30, Y: 0.16},
		Max: geom.Pt{X: 0.43, Y: 0.30},
	}
	if got != want {
		t.Fatalf("category inset rect = %+v, want %+v", got, want)
	}
}

func TestCustomUnitPanelMatchesMatplotlibDomainAndFormatter(t *testing.T) {
	fig := Plot()
	if len(fig.Children) <= 6 {
		t.Fatalf("gallery axes count = %d, want custom-unit panel at index 6", len(fig.Children))
	}
	ax := fig.Children[6]
	xMin, xMax := ax.XScale.Domain()
	if xMin != 3 || xMax != 44 {
		t.Fatalf("custom-unit xlim = (%v, %v), want (3, 44)", xMin, xMax)
	}
	if got := ax.XAxis.Formatter.Format(21.1); got != "21.1 km" {
		t.Fatalf("custom-unit x formatter = %q, want %q", got, "21.1 km")
	}
}

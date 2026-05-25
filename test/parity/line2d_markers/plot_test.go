package line2d_markers

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
)

func TestPlotUsesReferenceLegendLocation(t *testing.T) {
	fig := Plot()
	if len(fig.Children) == 0 {
		t.Fatal("figure has no axes")
	}

	legend := findLegend(fig.Children[0])
	if legend == nil {
		t.Fatal("plot should include a legend")
	}
	if legend.Location != core.LegendUpperLeft {
		t.Fatalf("legend location = %v, want upper left", legend.Location)
	}
}

func findLegend(ax *core.Axes) *core.Legend {
	if ax == nil {
		return nil
	}
	for _, art := range ax.Artists {
		legend, ok := art.(*core.Legend)
		if ok {
			return legend
		}
	}
	return nil
}

package stat_variants

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
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

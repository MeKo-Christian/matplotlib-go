package axes_grid1_showcase

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
)

func TestPlotUsesAxesCoordinateTextForTileLabels(t *testing.T) {
	fig := Plot()
	if fig == nil || len(fig.Children) == 0 {
		t.Fatal("Plot returned no axes")
	}

	var found bool
	for _, art := range fig.Children[0].Artists {
		text, ok := art.(*core.Text)
		if ok && text.Content == "image grid" && text.Coords == core.Coords(core.CoordAxes) {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("first ImageGrid tile should use axes-coordinate Text for the image grid label")
	}
}

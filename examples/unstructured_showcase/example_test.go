package unstructured_showcase

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
)

func TestPlotUsesTextArtistsForSourceTextCalls(t *testing.T) {
	fig := Plot()
	if fig == nil || len(fig.Children) == 0 {
		t.Fatal("Plot returned no axes")
	}

	var axesText bool
	for _, art := range fig.Children[0].Artists {
		text, ok := art.(*core.Text)
		if ok && text.Content == "explicit triangular mesh" && text.Coords == core.Coords(core.CoordAxes) {
			axesText = true
			break
		}
	}
	if !axesText {
		t.Fatal("mesh label should be an axes-coordinate Text artist")
	}

	var figureText bool
	for _, art := range fig.Artists {
		text, ok := art.(*core.Text)
		if ok && text.Content == "unstructured gallery family\ntriangulation, tripcolor, tricontour" && text.Coords == core.Coords(core.CoordFigure) {
			figureText = true
			break
		}
	}
	if !figureText {
		t.Fatal("figure note should be a figure-coordinate Text artist")
	}
}

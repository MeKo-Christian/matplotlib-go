package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestArtistLabelAccessorsCoverCommonLabeledArtists(t *testing.T) {
	artists := []Artist{
		&Line2D{},
		&Scatter2D{},
		&PathCollection{},
		&LineCollection{},
		&PatchCollection{},
		&Rectangle{},
		&Bar2D{},
		&Fill2D{},
		&Hist2D{},
		&BoxPlot2D{},
		&Image2D{},
		&ErrorBar{},
		&Stairs2D{},
		&Quiver{},
		&Barbs{},
	}

	for _, art := range artists {
		SetArtistLabel(art, "shared-label")
		if got := ArtistLabel(art); got != "shared-label" {
			t.Fatalf("%T ArtistLabel = %q, want shared-label", art, got)
		}
	}
}

func TestLegendCollectionUsesSharedArtistLabels(t *testing.T) {
	artists := []Artist{
		&Line2D{Label: "visible", XY: []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}}, Col: render.Color{A: 1}, W: 1},
		&Line2D{Label: "_hidden", XY: []geom.Pt{{X: 0, Y: 1}, {X: 1, Y: 0}}, Col: render.Color{A: 1}, W: 1},
	}

	entries := collectLegendEntries(artists)
	if len(entries) != 1 {
		t.Fatalf("legend entries = %d, want 1", len(entries))
	}
	if entries[0].Label != "visible" {
		t.Fatalf("legend label = %q, want visible", entries[0].Label)
	}
}

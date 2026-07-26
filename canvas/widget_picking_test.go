package canvas

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/widgets"
)

func TestPickPrefersWidgetLayerOverLaterHighZDataArtist(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	button := widgets.NewButton(ax, "Run", widgets.ButtonOptions{})
	ax.Add(widgetPickDataArtist{})

	hits := Pick(fig, geom.Pt{X: 60, Y: 40})
	if len(hits) == 0 {
		t.Fatal("expected pick hits")
	}
	if hits[0].Artist != button {
		t.Fatalf("top pick = %T, want button widget", hits[0].Artist)
	}
}

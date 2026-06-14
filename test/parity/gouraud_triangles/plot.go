package gouraud_triangles

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func Render() image.Image {
	fig := core.NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("RendererAgg Gouraud triangles")
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 3.2)
	ax.XAxis.Locator = core.MultipleLocator{Base: 0.5}
	ax.YAxis.Locator = core.MultipleLocator{Base: 0.5}

	ax.Add(&common.GouraudFixtureArtist{
		Points: []geom.Pt{
			{X: 0.35, Y: 0.35},
			{X: 1.80, Y: 0.30},
			{X: 3.40, Y: 0.55},
			{X: 0.80, Y: 1.70},
			{X: 2.20, Y: 2.70},
			{X: 3.55, Y: 1.75},
		},
		Triangles: [][3]int{{0, 1, 3}, {1, 4, 3}, {1, 2, 4}, {2, 5, 4}},
		Values:    []float64{0.05, 0.38, 0.82, 0.62, 1.00, 0.28},
	})

	return common.RenderFixtureFigure(fig, 980, 620)
}

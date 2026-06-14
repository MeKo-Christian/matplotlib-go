package spy_marker

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.22, Y: 0.14}, Max: geom.Pt{X: 0.78, Y: 0.9}})
	ax.SetTitle("Spy Marker")
	color := render.Color{R: 0.16, G: 0.38, B: 0.72, A: 1}
	marker := core.MarkerSquare
	ax.Spy(common.SparseFixtureData(14, 14), core.SpyOptions{
		Precision:  0.1,
		Marker:     &marker,
		MarkerSize: 8,
		Color:      &color,
	})

	return common.RenderImageFixture(fig, 640, 360)
}

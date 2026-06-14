package clip_path_batch

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func Render() image.Image {
	fig := core.NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("RendererAgg clip path batch")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 5.4)
	ax.XAxis.Locator = core.MultipleLocator{Base: 1}
	ax.YAxis.Locator = core.MultipleLocator{Base: 1}
	ax.AddYGrid()
	ax.Add(&common.ClipPathBatchFixtureArtist{})

	return common.RenderFixtureFigure(fig, 980, 620)
}

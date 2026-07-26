package spy_image

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Render() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.22, Y: 0.14}, Max: geom.Pt{X: 0.78, Y: 0.9}})
	ax.SetTitle("Spy Image")
	useImage := true
	ax.Spy(common.SparseFixtureData(14, 14), core.SpyOptions{
		Precision: 0.1,
		UseImage:  optional.Of(useImage),
	})

	return common.RenderImageFixture(fig, 640, 360)
}

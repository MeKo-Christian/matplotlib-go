package colorbar_boundary_values

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

func Plot() *core.Figure {
	fig, ax := common.ColorNormFixtureFigure("Boundary Colorbar Values")
	cmap := "viridis"
	xmin, xmax := 0.0, 4.0
	ymin, ymax := 0.0, 3.0
	img := ax.Image([][]float64{
		{-0.7, -0.2, 0.2, 0.8},
		{-0.5, 0.0, 0.6, 1.2},
		{-0.1, 0.3, 0.9, 1.4},
	}, core.ImageOptions{
		Colormap: optional.Of(cmap),
		Norm:     core.Normalize{VMin: -0.5, VMax: 1.2},
		XMin:     optional.Of(xmin),
		XMax:     optional.Of(xmax),
		YMin:     optional.Of(ymin),
		YMax:     optional.Of(ymax),
		Origin:   core.ImageOriginLower,
	})
	if img != nil {
		fig.AddColorbar(ax, img, core.ColorbarOptions{
			Label:      "bands",
			Extend:     "both",
			ExtendRect: true,
			Boundaries: []float64{-0.5, -0.1, 0.4, 1.2},
			Values:     []float64{-0.35, 0.15, 0.8},
			Spacing:    "uniform",
			DrawEdges:  true,
			Ticks:      []float64{-0.5, -0.1, 0.4, 1.2},
		})
	}
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 3)
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), 640, 360)
}

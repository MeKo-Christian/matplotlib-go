package imshow_rgb

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
)

// rgbGradient builds an (ny,nx,3) image whose red ramps left→right and green
// ramps top→bottom, with a constant blue. Channel values are exact fractions
// so the float→byte truncation matches matplotlib.
func rgbGradient(ny, nx int) [][][]float64 {
	out := make([][][]float64, ny)
	for row := range out {
		out[row] = make([][]float64, nx)
		for col := range out[row] {
			out[row][col] = []float64{
				float64(col) / float64(nx-1),
				float64(row) / float64(ny-1),
				0.30,
			}
		}
	}
	return out
}

// rgbaBlocks builds an (ny,nx,4) solid-red image whose alpha ramps
// left→right, exercising the RGBA alpha-bypass path over a white background.
func rgbaBlocks(ny, nx int) [][][]float64 {
	out := make([][][]float64, ny)
	for row := range out {
		out[row] = make([][]float64, nx)
		for col := range out[row] {
			out[row][col] = []float64{
				0.85, 0.10, 0.10,
				float64(col) / float64(nx-1),
			}
		}
	}
	return out
}

func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)

	nearest := "nearest"

	axRGB := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.12}, Max: geom.Pt{X: 0.47, Y: 0.90}})
	axRGB.SetTitle("RGB")
	_, _ = axRGB.ImShowRGB(rgbGradient(8, 8), core.ImShowRGBOptions{
		Interpolation: optional.Of(nearest),
		Aspect:        core.AspectAuto,
	})

	axRGBA := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.55, Y: 0.12}, Max: geom.Pt{X: 0.95, Y: 0.90}})
	axRGBA.SetTitle("RGBA")
	_, _ = axRGBA.ImShowRGB(rgbaBlocks(8, 8), core.ImShowRGBOptions{
		Interpolation: optional.Of(nearest),
		Aspect:        core.AspectAuto,
	})

	return fig
}

func Render() image.Image {
	fig := Plot()
	return common.RenderImageFixture(fig, 640, 360)
}

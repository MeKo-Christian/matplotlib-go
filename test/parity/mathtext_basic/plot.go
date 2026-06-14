package mathtext_basic

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
)

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.14},
		Max: geom.Pt{X: 0.92, Y: 0.88},
	})

	const n = 120
	x := make([]float64, n)
	y := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1) * 4 * math.Pi
		x[i] = t
		y[i] = math.Sin(t) * math.Exp(-0.08*t)
	}

	lineWidth := 2.0
	blue := render.Color{R: 31.0 / 255.0, G: 119.0 / 255.0, B: 180.0 / 255.0, A: 1}
	ax.Plot(x, y, core.PlotOptions{Color: &blue, LineWidth: &lineWidth})
	ax.SetTitle(`MathText $\alpha^2 + \beta_i$`)
	ax.SetXLabel(`phase $\theta$`)
	ax.SetYLabel(`amplitude $\frac{1}{\sqrt{2}}$`)
	ax.Text(0.98, 0.92, `$x_{\mathrm{max}}$`, core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		HAlign:   core.TextAlignRight,
		VAlign:   core.TextVAlignTop,
		FontSize: 12,
	})
	ax.Annotate(`$\Delta y \approx \frac{1}{2}$`, 3.2, 0.35, core.AnnotationOptions{
		OffsetX:    common.ReferencePointsToPixels(34),
		OffsetY:    common.ReferencePointsToPixels(-26),
		FontSize:   12,
		ArrowColor: render.Color{R: 0, G: 0, B: 0, A: 1},
		ArrowWidth: 1,
	})
	ax.Text(0.03, 0.93, `$\omega_n = 2\pi f_n$`, core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignTop,
		FontSize: 11,
		BBox: &core.TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1},
			Padding:   0.3,
		},
	})
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

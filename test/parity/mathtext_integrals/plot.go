package mathtext_integrals

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

const (
	Width  = 640
	Height = 360
)

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.08},
		Max: geom.Pt{X: 0.95, Y: 0.92},
	})
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)
	hideAxes(ax)

	ax.Text(0.50, 0.80, `$\int_0^\infty e^{-x}\,dx = 1$`, centerText(24))
	ax.Text(0.34, 0.50, `$\sum_{i=1}^{n} i^2$`, centerText(26))
	ax.Text(0.66, 0.50, `$\prod_{k=1}^{m} k$`, centerText(26))
	ax.Text(0.50, 0.16, `$\lim_{x\to 0} \frac{\sin x}{x} = 1$`, centerText(23))
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

func centerText(size float64) core.TextOptions {
	return core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignMiddle,
		FontSize: size,
	}
}

func hideAxes(ax *core.Axes) {
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false
}

package mathtext_matrices

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

	ax.Text(0.30, 0.64, `$\genfrac{(}{)}{0}{0}{a\quad b}{c\quad d}$`, centerText(25))
	ax.Text(0.70, 0.64, `$\genfrac{[}{]}{0}{0}{1\quad 0}{0\quad 1}$`, centerText(25))
	ax.Text(0.30, 0.30, `$\genfrac{(}{)}{0}{0}{x}{y}$`, centerText(24))
	ax.Text(0.70, 0.30, `$\left\langle\genfrac{}{}{0}{0}{u}{v}\right\rangle$`, centerText(24))
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

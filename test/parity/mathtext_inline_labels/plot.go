package mathtext_inline_labels

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
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.92, Y: 0.88},
	})

	const n = 90
	x := make([]float64, n)
	y1 := make([]float64, n)
	y2 := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1) * 5
		x[i] = t
		y1[i] = 0.55 + 0.35*math.Sin(1.5*t)
		y2[i] = 0.48 + 0.28*math.Cos(1.5*t+0.45)
	}

	lineWidth := 2.0
	blue := render.Color{R: 31.0 / 255.0, G: 119.0 / 255.0, B: 180.0 / 255.0, A: 1}
	orange := render.Color{R: 1, G: 127.0 / 255.0, B: 14.0 / 255.0, A: 1}
	_, _ = ax.Plot(x, y1, core.PlotOptions{Color: &blue, LineWidth: &lineWidth, Label: `state $x_i(t)$`})
	_, _ = ax.Plot(x, y2, core.PlotOptions{Color: &orange, LineWidth: &lineWidth, Label: `state $y_i(t)$`})
	ax.SetTitle(`Inline labels: $\omega_n$ response`)
	ax.SetXLabel(`time $t$`)
	ax.SetYLabel(`state $x_i(t)$`)
	ax.AddLegend()
	ax.Text(0.03, 0.88, `peak $\alpha_i^2$`, core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignTop,
		FontSize: 12,
	})
	ax.Text(0.97, 0.08, `ratio $\frac{a}{b}$`, core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		HAlign:   core.TextAlignRight,
		VAlign:   core.TextVAlignBottom,
		FontSize: 12,
	})
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

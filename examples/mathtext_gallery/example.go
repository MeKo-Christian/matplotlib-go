// Package mathtext_gallery is a user-facing gallery for MathText layout.
package mathtext_gallery

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 900
	Height = 560
	DPI    = 100
)

var mathTextPanels = []struct {
	Title string
	Exprs []string
}{
	{
		Title: "Fractions and Roots",
		Exprs: []string{
			`$\frac{1}{2} + \frac{a+b}{c+d}$`,
			`$\sqrt{x^2 + y^2}$`,
		},
	},
	{
		Title: "Operators",
		Exprs: []string{
			`$\int_0^\infty e^{-x}\,dx = 1$`,
			`$\sum_{i=1}^{n} i^2$`,
		},
	},
	{
		Title: "Fences and Matrices",
		Exprs: []string{
			`$\left[\frac{1}{1+x}\right]$`,
			`$\genfrac{(}{)}{0}{0}{a\quad b}{c\quad d}$`,
		},
	},
	{
		Title: "Inline Labels",
		Exprs: []string{
			`phase $\alpha_i^2$ peak`,
			`ratio $\frac{a}{b}$`,
		},
	},
}

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	fig.Text(0.05, 0.94, "MathText Gallery", core.TextOptions{FontSize: 13})

	for i, panel := range mathTextPanels {
		col := i % 2
		row := i / 2
		x0 := 0.07 + float64(col)*0.46
		y0 := 0.52 - float64(row)*0.42
		ax := fig.AddAxes(geom.Rect{
			Min: geom.Pt{X: x0, Y: y0},
			Max: geom.Pt{X: x0 + 0.39, Y: y0 + 0.31},
		})
		ax.SetTitle(panel.Title)
		ax.SetXLim(0, 1)
		ax.SetYLim(0, 1)
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
		ax.YAxis.ShowTicks = false
		ax.YAxis.ShowLabels = false

		ax.Text(0.50, 0.64, panel.Exprs[0], centerMathText(18))
		ax.Text(0.50, 0.34, panel.Exprs[1], centerMathText(17))
	}

	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func centerMathText(size float64) core.TextOptions {
	return core.TextOptions{
		Coords:   core.Coords(core.CoordData),
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignMiddle,
		FontSize: size,
	}
}

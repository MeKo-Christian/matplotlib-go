package named_colors

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 640
	Height = 360
)

// Plot builds a compact fixture covering Matplotlib named-color resolution.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.18},
		Max: geom.Pt{X: 0.90, Y: 0.88},
	})
	ax.SetTitle("Named Colors")
	ax.SetXLabel("color spec")
	ax.SetYLabel("value")
	ax.SetXLim(0, 8)
	ax.SetYLim(0, 8)

	specs := []any{
		"#66c2a5",
		"0.35",
		"tab:orange",
		"rebeccapurple",
		"xkcd:cloudy blue",
		"C3",
		"(0.15, 0.45, 0.65, 1)",
	}
	labels := []string{"hex", "gray", "tab", "css", "xkcd", "C3", "tuple"}
	colors := make([]render.Color, 0, len(specs))
	for _, spec := range specs {
		col, err := matcolor.ToRGBA(spec)
		if err != nil {
			panic(err)
		}
		colors = append(colors, col)
	}

	x := []float64{1, 2, 3, 4, 5, 6, 7}
	width := 0.68
	edgeColor := render.Color{R: 0.12, G: 0.12, B: 0.12, A: 1}
	edgeWidth := 1.0
	_, _ = ax.Bar(x, []float64{2.4, 3.4, 4.3, 5.2, 6.1, 4.8, 3.7}, core.BarOptions{
		Width:     &width,
		Colors:    colors,
		EdgeColor: &edgeColor,
		EdgeWidth: &edgeWidth,
	})
	ax.XAxis.Locator = ticker.FixedLocator{TicksList: x}
	ax.XAxis.Formatter = ticker.FixedFormatter{Labels: labels}
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 2, 4, 6, 8}}
	return fig
}

// Render returns the AGG-rendered parity image.
func Render() image.Image {
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(Plot(), r)
	return r.Image()
}

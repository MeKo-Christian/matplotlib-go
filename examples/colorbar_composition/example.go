// Package colorbar_composition is a showcase example that combines an
// imshow heatmap with an attached colorbar managed by ConstrainedLayout.
package colorbar_composition

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 1000
	Height = 700
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	fig.ConstrainedLayout()
	ax := fig.AddSubplot(1, 1, 1)

	rows, cols := 80, 120
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		for col := range data[row] {
			x := (float64(col)/float64(cols-1))*4 - 2
			y := (float64(row)/float64(rows-1))*4 - 2
			r := math.Hypot(x, y)
			data[row][col] = math.Sin(3*r) * math.Exp(-0.6*r)
		}
	}

	cmap := "inferno"
	extent := [4]float64{0, float64(cols), 0, float64(rows)}
	im := ax.ImShow(data, core.ImShowOptions{
		Colormap: optional.Of(cmap),
		Origin:   optional.Of(core.ImageOriginLower),
		Extent:   optional.Of(extent),
		Aspect:   optional.Of(core.AspectAuto),
	})

	ax.SetTitle("Heatmap with Colorbar")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(0, float64(cols))
	ax.SetYLim(0, float64(rows))
	yTicks := make([]float64, 0, rows/20+1)
	for tick := 0; tick <= rows; tick += 20 {
		yTicks = append(yTicks, float64(tick))
	}
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: yTicks}

	gridColor := render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	for _, grid := range []*core.Grid{ax.AddXGrid(), ax.AddYGrid()} {
		grid.Color = gridColor
		grid.LineWidth = 0.5
	}

	cbar := fig.AddColorbar(ax, im, core.ColorbarOptions{})
	cbar.SetYLabel("Intensity")

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
	return r.Image()
}

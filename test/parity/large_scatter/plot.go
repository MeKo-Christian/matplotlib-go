package large_scatter

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := core.NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.09, Y: 0.13}, Max: geom.Pt{X: 0.95, Y: 0.88}})
	ax.SetTitle("RendererAgg marker batch")
	ax.SetXLim(-0.5, 14.5)
	ax.SetYLim(-1.5, 11.5)
	ax.AddYGrid()

	points := make([]geom.Pt, 0, 180)
	sizes := make([]float64, 0, 180)
	colors := make([]render.Color, 0, 180)
	edges := make([]render.Color, 0, 180)
	for i := 0; i < 180; i++ {
		x := float64(i%15) + 0.24*math.Sin(float64(i)*0.73)
		y := float64((i*7)%12) + 0.28*math.Cos(float64(i)*0.41)
		points = append(points, geom.Pt{X: x, Y: y})
		radius := 4.0 + float64((i*11)%9)
		sizes = append(sizes, core.ScatterAreaFromRadius(radius, fig.RC.DPI))
		t := float64(i%30) / 29.0
		colors = append(colors, render.Color{R: 0.12 + 0.70*t, G: 0.58 - 0.25*t, B: 0.88 - 0.56*t, A: 0.72})
		edges = append(edges, render.Color{R: 0.08, G: 0.10 + 0.28*t, B: 0.18, A: 0.95})
	}
	ax.Add(&core.Scatter2D{
		XY:         points,
		Sizes:      sizes,
		Colors:     colors,
		EdgeColors: edges,
		EdgeWidth:  0.75,
		Marker:     core.MarkerCircle,
		Label:      "batched markers",
	})
	ax.AddLegend()

	return common.RenderFixtureFigure(fig, 980, 620)
}

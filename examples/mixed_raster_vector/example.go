package mixed_raster_vector

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 640
	DPI    = 100
)

// Plot builds a mixed raster/vector output showcase.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddPolarAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.12}, Max: geom.Pt{X: 0.88, Y: 0.88}})
	ax.SetTitle("Mixed Raster / Vector")
	ax.SetYLim(0, 1.0)

	thetaGrid := ax.AddGrid(core.AxisBottom)
	thetaGrid.Color = render.Color{R: 0.82, G: 0.84, B: 0.88, A: 1}
	thetaGrid.LineWidth = 0.8
	radiusGrid := ax.AddGrid(core.AxisLeft)
	radiusGrid.Color = render.Color{R: 0.84, G: 0.86, B: 0.90, A: 0.9}
	radiusGrid.LineWidth = 0.8

	const n = 240
	points := make([]geom.Pt, 0, n)
	sizes := make([]float64, 0, n)
	colors := make([]render.Color, 0, n)
	edges := make([]render.Color, 0, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1)
		theta := 8.0*math.Pi*t + 0.18*math.Sin(11*t)
		radius := 0.12 + 0.98*t + 0.08*math.Sin(7*theta)
		points = append(points, geom.Pt{X: theta, Y: radius})
		sizes = append(sizes, core.ScatterAreaFromRadius(3.2+2.4*math.Sin(math.Pi*t)*math.Sin(math.Pi*t), fig.RC.DPI))
		colors = append(colors, render.Color{R: 0.08 + 0.70*t, G: 0.30 + 0.45*(1-t), B: 0.86 - 0.48*t, A: 0.56})
		edges = append(edges, render.Color{R: 0.02, G: 0.08, B: 0.18, A: 0.42})
	}
	cloud := &core.Scatter2D{
		XY:         points,
		Sizes:      sizes,
		Colors:     colors,
		EdgeColors: edges,
		EdgeWidth:  0.45,
		Marker:     core.MarkerCircle,
		Label:      "raster cloud",
	}
	cloud.SetRasterization(render.Rasterization{Mode: render.RasterizeAlways})
	ax.Add(cloud)

	lineTheta := make([]float64, 180)
	lineRadius := make([]float64, 180)
	for i := range lineTheta {
		theta := 2 * math.Pi * float64(i) / float64(len(lineTheta)-1)
		lineTheta[i] = theta
		lineRadius[i] = 0.58 + 0.16*math.Cos(5*theta)
	}
	lineColor := render.Color{R: 0.08, G: 0.16, B: 0.30, A: 1}
	lineWidth := 1.8
	ax.Plot(lineTheta, lineRadius, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
		Label:     "vector line",
	})
	ax.AddLegend()
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}

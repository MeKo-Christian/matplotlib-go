package legend_layout_matrix

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 420
	DPI    = 100
)

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.90}})
	ax.SetTitle("Legend Layout Matrix")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 5)

	blue := render.Color{R: 0.12, G: 0.31, B: 0.68, A: 1}
	orange := render.Color{R: 0.86, G: 0.43, B: 0.16, A: 1}
	green := render.Color{R: 0.21, G: 0.58, B: 0.27, A: 1}
	red := render.Color{R: 0.72, G: 0.20, B: 0.18, A: 1}
	purple := render.Color{R: 0.47, G: 0.31, B: 0.70, A: 1}
	lineWidth := 2.0
	markerSize := 5.0
	circle := core.MarkerCircle
	square := core.MarkerSquare

	ax.Plot([]float64{0.4, 1.4, 2.4, 3.4, 4.4, 5.4}, []float64{1.0, 1.7, 1.4, 2.2, 2.0, 2.7}, core.PlotOptions{
		Color:     &blue,
		LineWidth: &lineWidth,
		Label:     "line",
	})
	ax.Scatter([]float64{0.7, 1.7, 2.7, 3.7, 4.7}, []float64{3.3, 3.8, 3.1, 3.6, 3.2}, core.ScatterOptions{
		Color:     &orange,
		EdgeColor: &red,
		Marker:    &circle,
		Label:     "scatter",
	})
	ax.ErrorBar([]float64{1.0, 2.3, 3.6, 4.9}, []float64{0.85, 1.15, 0.95, 1.25}, nil, []float64{0.22, 0.18, 0.25, 0.20}, core.ErrorBarOptions{
		Color:      &green,
		LineWidth:  &lineWidth,
		CapSize:    &markerSize,
		Marker:     &square,
		MarkerSize: &markerSize,
		Label:      "errorbar",
	})
	handlerLine := ax.Plot([]float64{0.5, 1.6, 2.7, 3.8, 4.9}, []float64{4.5, 4.1, 4.35, 4.0, 4.25}, core.PlotOptions{
		Color:     &purple,
		LineWidth: &lineWidth,
		Label:     "handler patch",
	})

	legend := ax.AddLegend()
	legend.Location = core.LegendUpperLeft
	legend.Title = "Collected"
	legend.NumColumns = 2
	legend.ColumnSpacing = 28
	legend.MarkerScale = 1.8
	legend.ScatterPoints = 3
	legend.SetHandler(handlerLine, core.LegendEntryOptions{
		Sample:    core.LegendSamplePatch,
		FaceColor: render.Color{R: 0.74, G: 0.52, B: 0.83, A: 0.85},
		EdgeColor: purple,
		EdgeWidth: 1.2,
		Hatch:     "/",
	})

	proxyLegend := ax.AddLegend()
	proxyLegend.Location = core.LegendLowerRight
	proxyLegend.FrameOn = false
	proxyLegend.AddEntry("proxy patch", core.LegendEntryOptions{
		Sample:    core.LegendSamplePatch,
		FaceColor: render.Color{R: 0.93, G: 0.77, B: 0.33, A: 0.92},
		EdgeColor: render.Color{R: 0.45, G: 0.30, B: 0.08, A: 1},
		EdgeWidth: 1.2,
		Hatch:     "xx",
	})

	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

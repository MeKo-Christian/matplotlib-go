package annotation_composition

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1040
	Height = 720
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(1040, 720)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.14},
		Max: geom.Pt{X: 0.90, Y: 0.88},
	})
	ax.SetTitle("Text and Arrow Annotations")
	ax.SetXLabel("phase")
	ax.SetYLabel("response")
	gridColor := render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	for _, grid := range []*core.Grid{ax.AddXGrid(), ax.AddYGrid()} {
		grid.Color = gridColor
		grid.LineWidth = 0.5
	}

	x := make([]float64, 240)
	y := make([]float64, 240)
	for i := range x {
		xv := 6 * math.Pi * float64(i) / float64(len(x)-1)
		x[i] = xv
		y[i] = math.Sin(xv)*math.Exp(-0.015*xv) + 0.2*math.Cos(0.5*xv)
	}
	_, _ = ax.Plot(x, y, core.PlotOptions{Label: "signal"})
	ax.SetXLim(0, 6*math.Pi)
	ax.SetYLim(-1.2, 1.2)
	ax.AddLegend()

	peakX := math.Pi / 2
	peakY := math.Sin(peakX)*math.Exp(-0.015*peakX) + 0.2*math.Cos(0.5*peakX)
	ax.Annotate("Peak\n= 0.42", peakX, peakY, core.AnnotationOptions{
		OffsetX:    optional.Of(48.0),
		OffsetY:    optional.Of(-42.0),
		FontSize:   12,
		ArrowColor: render.Color{R: 0, G: 0, B: 0, A: 1},
		ArrowWidth: optional.Of(1.0),
	})
	ax.Text(0.20, 0.90, "m∫T  φ x =  λ/4", core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		FontSize: 12,
	})
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

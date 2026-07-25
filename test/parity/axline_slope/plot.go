// Package axline_slope is a focused parity fixture for Axes.AxLineSlope.
package axline_slope

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds a single slope-defined infinite line clipped to the axes.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 6)
	ax.SetTitle("Slope-defined axline")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	color := render.Color{R: 31.0 / 255, G: 119.0 / 255, B: 180.0 / 255, A: 1}
	lineWidth := 1.5
	ax.AxLineSlope(
		geom.Pt{X: 2, Y: 3},
		0.75,
		core.ReferenceLineOptions{Color: &color, LineWidth: &lineWidth},
	)
	return fig
}

// Render is the AGG-rendered parity image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

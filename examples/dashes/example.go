// Package dashes is a showcase example that draws four parallel horizontal
// lines, each with a different dash pattern, demonstrating Line2D's dash
// support.
package dashes

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

// Plot builds the showcase figure. It is backend-agnostic: callers choose
// how to render it (AGG, GoBasic, SVG, ...).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Dash Patterns")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 5)

	specs := []struct {
		y       float64
		pattern []float64
		color   render.Color
	}{
		{4, nil, render.Color{R: 0, G: 0, B: 0, A: 1}},
		{3, []float64{10, 4}, render.Color{R: 0.8, G: 0, B: 0, A: 1}},
		{2, []float64{6, 2, 2, 2}, render.Color{R: 0, G: 0.6, B: 0, A: 1}},
		{1, []float64{2, 2}, render.Color{R: 0, G: 0, B: 0.8, A: 1}},
	}

	for _, spec := range specs {
		line := &core.Line2D{
			XY: []geom.Pt{
				{X: 1, Y: spec.y},
				{X: 9, Y: spec.y},
			},
			W:   3.0,
			Col: spec.color,
		}
		if len(spec.pattern) > 0 {
			line.SetDashes(spec.pattern...)
		}
		ax.Add(line)
	}

	return fig
}

// Render is the AGG-rendered showcase image. Used by the parity registry
// (test/parity/dashes), the web demo, and CLI exporters.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

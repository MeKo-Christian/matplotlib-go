// Package linecollection_linestyle is a parity fixture exercising LineCollection
// string linestyles ("solid"/"dashed"/"dashdot"/"dotted") converted to dash
// patterns, here through Axes.HLines.
package linecollection_linestyle

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Shared with test/parity/linecollection_linestyle/plot.py.
var (
	ys         = []float64{1, 2, 3, 4}
	lineStyles = []string{"solid", "dashed", "dashdot", "dotted"}
)

// Plot builds the figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetTitle("Line Styles")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 5)

	ax.HLines(ys, []float64{0.5}, []float64{9.5}, core.LineCollectionOptions{
		Color:      optional.Of(render.Color{R: 0, G: 0, B: 0, A: 1}),
		LineWidth:  optional.Of(2.0),
		LineStyles: lineStyles,
		LineCap:    render.CapButt,
	})
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
	return r.Image()
}

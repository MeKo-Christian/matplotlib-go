package joins_caps

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

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Line Joins and Caps")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 6)

	joinPath := []geom.Pt{
		{X: 1, Y: 5}, {X: 3, Y: 5}, {X: 3, Y: 3}, {X: 5, Y: 3},
	}
	miterLine := &core.Line2D{
		XY:  joinPath,
		W:   8.0,
		Col: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
	}
	ax.Add(miterLine)

	capPath := []geom.Pt{
		{X: 7, Y: 5}, {X: 9, Y: 5},
	}
	capLine := &core.Line2D{
		XY:  capPath,
		W:   8.0,
		Col: render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1},
	}
	ax.Add(capLine)
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

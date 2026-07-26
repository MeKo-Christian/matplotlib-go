package quiver_autoscale

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
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
//
// It draws a quiver field with NO explicit scale, exercising matplotlib's
// default autoscale (quiver.py:673-681): scale = 1.8 * amean * sn / span with
// sn = max(10, sqrt(N)) and span = 1 for the default units="width". The field
// is the deterministic cos(0.5x)/sin(0.5y) grid so the Go and matplotlib sides
// produce identical arrow data. An explicit color is set on both sides because
// matplotlib's default quiver color is not the property cycle.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})
	ax.SetTitle("Quiver Autoscale (default scale)")
	ax.SetXLim(-0.5, 8.5)
	ax.SetYLim(-0.5, 6.5)

	var qx, qy, qu, qv []float64
	for iy := 0; iy <= 6; iy++ {
		y := float64(iy)
		for ix := 0; ix <= 8; ix++ {
			x := float64(ix)
			qx = append(qx, x)
			qy = append(qy, y)
			qu = append(qu, math.Cos(0.5*x))
			qv = append(qv, math.Sin(0.5*y))
		}
	}

	blue := render.Color{R: 0.15, G: 0.35, B: 0.65, A: 1}
	ax.Quiver(qx, qy, qu, qv, core.QuiverOptions{Color: optional.Of(blue)})
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

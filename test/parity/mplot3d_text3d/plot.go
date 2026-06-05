package mplot3d_text3d

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 560
	DPI    = 100
)

type label struct {
	x, y, z float64
	s       string
}

// Plot places flat (zdir=None) text labels at 3D positions. Explicit axis
// limits are set in both Go and the matplotlib reference so the projection is
// identical regardless of whether text artists participate in autoscaling
// (Go's Text3D expands the data limits; matplotlib text does not).
func Plot() *core.Figure {
	fig := core.NewFigure(720, 560)
	ax, err := fig.AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		panic(err)
	}

	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	ax.SetZLim(0, 10)

	labels := []label{
		{1, 1, 1, "alpha"},
		{8, 2, 3, "beta"},
		{3, 7, 5, "gamma"},
		{6, 6, 8, "delta"},
		{2, 4, 9, "epsilon"},
		{9, 9, 2, "zeta"},
	}
	for _, l := range labels {
		ax.Text3D(l.x, l.y, l.z, l.s)
	}

	ax.SetXLabel("X")
	ax.SetYLabel("Y")
	ax.SetZLabel("Z")
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
	return r.GetImage()
}

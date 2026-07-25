package mplot3d_text3d

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/plot3d"
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
// limits mirror the matplotlib reference and keep the text-only example on the
// same 0..10 view volume as the upstream 3D text/pathpatch gallery pattern.
func Plot() *core.Figure {
	fig := core.NewFigure(720, 560)
	ax, err := plot3d.AddAxes(fig, geom.Rect{
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
	return r.Image()
}

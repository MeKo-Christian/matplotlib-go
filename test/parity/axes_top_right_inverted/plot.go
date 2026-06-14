package axes_top_right_inverted

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
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
		Min: geom.Pt{X: 0.1, Y: 0.12},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Top/Right Axes + Inversion")
	ax.SetXLabel("Bottom X")
	ax.SetYLabel("Left Y")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)

	top := ax.TopAxis()
	top.MinorLocator = nil
	right := ax.RightAxis()
	right.MinorLocator = nil

	line := &core.Line2D{
		XY: []geom.Pt{
			{X: 1, Y: 2},
			{X: 3, Y: 4},
			{X: 6, Y: 6.5},
			{X: 8.5, Y: 8},
		},
		W:   2.2,
		Col: render.Color{R: 0.15, G: 0.35, B: 0.75, A: 1},
	}
	ax.Add(line)

	scatter := &core.Scatter2D{
		XY: []geom.Pt{
			{X: 2, Y: 8},
			{X: 5, Y: 5},
			{X: 8, Y: 2},
		},
		Size:      core.ScatterAreaFromRadius(9.0, style.Default.DPI),
		Color:     render.Color{R: 0.85, G: 0.35, B: 0.20, A: 0.9},
		EdgeColor: render.Color{R: 0.45, G: 0.15, B: 0.05, A: 1},
		EdgeWidth: 1.0,
		Marker:    core.MarkerDiamond,
		Alpha:     1.0,
	}
	ax.Add(scatter)

	ax.InvertX()
	ax.InvertY()
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

// Package named_colors_gallery is a user-facing swatch gallery for Matplotlib
// color-name compatibility.
package named_colors_gallery

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 900
	Height = 520
	DPI    = 100
)

type swatch struct {
	Label string
	Spec  any
}

var swatches = []swatch{
	{Label: "b", Spec: "b"},
	{Label: "tab:orange", Spec: "tab:orange"},
	{Label: "tab:green", Spec: "tab:green"},
	{Label: "rebeccapurple", Spec: "rebeccapurple"},
	{Label: "mediumseagreen", Spec: "mediumseagreen"},
	{Label: "goldenrod", Spec: "goldenrod"},
	{Label: "xkcd:cloudy blue", Spec: "xkcd:cloudy blue"},
	{Label: "xkcd:burnt orange", Spec: "xkcd:burnt orange"},
	{Label: "0.25", Spec: "0.25"},
	{Label: "#66c2a5", Spec: "#66c2a5"},
	{Label: "C3", Spec: "C3"},
	{Label: "rgba tuple", Spec: []float64{0.15, 0.45, 0.65, 1}},
}

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.05, Y: 0.08}, Max: geom.Pt{X: 0.95, Y: 0.88}})
	ax.SetTitle("Named Color Swatches")
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 3)
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false

	for i, sw := range swatches {
		col, err := matcolor.ToRGBA(sw.Spec)
		if err != nil {
			panic(err)
		}
		x := float64(i % 4)
		y := float64(2 - i/4)
		ax.AddPatch(&core.Rectangle{
			Patch: core.Patch{
				FaceColor: col,
				EdgeColor: render.Color{R: 0.15, G: 0.15, B: 0.15, A: 1},
				EdgeWidth: 0.8,
			},
			XY:     geom.Pt{X: x + 0.12, Y: y + 0.28},
			Width:  0.76,
			Height: 0.42,
		})
		ax.Text(x+0.50, y+0.14, sw.Label, core.TextOptions{
			HAlign:   core.TextAlignCenter,
			VAlign:   core.TextVAlignMiddle,
			FontSize: 9,
		})
	}

	fig.Text(0.05, 0.94, "CSS4, Tableau, xkcd, grayscale, cycle, hex, and tuple specs", core.TextOptions{
		FontSize: 10,
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
	return r.GetImage()
}

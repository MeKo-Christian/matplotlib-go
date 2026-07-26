package layout_bbox_helpers

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 420
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(720, 420)

	leftRect := geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.19}, Max: geom.Pt{X: 0.45, Y: 0.78}}
	rightRect := geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.34}, Max: geom.Pt{X: 0.88, Y: 0.82}}
	unionRect, _ := geom.UnionRects(leftRect, rightRect)
	paddedRect := unionRect.Padded(0.035)
	anchoredRect, _ := paddedRect.Anchored(geom.Pt{X: 0.24, Y: 0.18}, "lower right")

	left := fig.AddAxes(leftRect)
	configureAxes(left, "left bbox", render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1})
	right := fig.AddAxes(rightRect)
	configureAxes(right, "right bbox", render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1})

	anchored := fig.AddAxes(anchoredRect)
	configureAxes(anchored, "anchored", render.Color{R: 0.84, G: 0.15, B: 0.16, A: 1})

	fig.Add(&core.Rectangle{
		Patch: core.Patch{
			FaceColor: optional.Of(render.Color{R: 0.95, G: 0.74, B: 0.20, A: 0.08}),
			EdgeColor: optional.Of(render.Color{R: 0.42, G: 0.34, B: 0.12, A: 1}),
			EdgeWidth: optional.Of(1.4),
			Dashes:    []float64{6, 4},
			DashUnits: core.DashUnitsMatplotlib,
		},
		XY:     paddedRect.Min,
		Width:  paddedRect.W(),
		Height: paddedRect.H(),
		Coords: core.Coords(core.CoordFigure),
	})
	return fig
}

func configureAxes(ax *core.Axes, title string, col render.Color) {
	ax.SetTitle(title)
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: 0.10, Y: 0.20},
			{X: 0.38, Y: 0.76},
			{X: 0.66, Y: 0.35},
			{X: 0.90, Y: 0.82},
		},
		W:   1.8,
		Col: col,
	})
	ax.Text(0.08, 0.08, title, core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		FontSize: 9,
		Color:    col,
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignBottom,
	})
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

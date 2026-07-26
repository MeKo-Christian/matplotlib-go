package path_clipped_transformed

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

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.90, Y: 0.84},
	})
	ax.SetTitle("Clipped Transformed Path")
	ax.SetXLabel("X")
	ax.SetYLabel("Y")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 6)
	gridColor := render.Color{R: 0.82, G: 0.82, B: 0.82, A: 1}
	for _, grid := range []*core.Grid{ax.AddXGrid(), ax.AddYGrid()} {
		grid.Color = gridColor
		grid.LineWidth = 0.5
	}

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: -1.6, Y: 0.8})
	path.CubicTo(
		geom.Pt{X: 1.0, Y: 7.2},
		geom.Pt{X: 4.2, Y: -1.2},
		geom.Pt{X: 6.3, Y: 4.8},
	)
	path.CubicTo(
		geom.Pt{X: 8.1, Y: 9.0},
		geom.Pt{X: 12.0, Y: 4.1},
		geom.Pt{X: 10.8, Y: -0.9},
	)
	path.LineTo(geom.Pt{X: 8.0, Y: 1.1})
	path.CubicTo(
		geom.Pt{X: 5.8, Y: 2.2},
		geom.Pt{X: 3.8, Y: 1.2},
		geom.Pt{X: 1.2, Y: 3.2},
	)
	path.Close()

	ax.AddPatch(&core.PathPatch{
		Patch: core.Patch{
			FaceColor: optional.Of(render.Color{R: 0.12, G: 0.47, B: 0.71, A: 0.38}),
			EdgeColor: optional.Of(render.Color{R: 0.05, G: 0.20, B: 0.36, A: 1}),
			EdgeWidth: optional.Of(1.7),
		},
		Path:   path,
		Coords: core.Coords(core.CoordData),
	})
	ax.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 10, Y: 6},
		},
		W:   1.0,
		Col: render.Color{R: 0.84, G: 0.15, B: 0.16, A: 0.78},
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
	return r.Image()
}

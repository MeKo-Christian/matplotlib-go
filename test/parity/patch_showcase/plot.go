package patch_showcase

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 930
	Height = 340
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(930, 340)

	left := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.05, Y: 0.16}, Max: geom.Pt{X: 0.31, Y: 0.88}})
	left.SetTitle("Patch Primitives")
	left.SetXLim(0, 6)
	left.SetYLim(0, 4)
	left.AddPatch(&core.Rectangle{
		Patch: core.Patch{
			FaceColor: optional.Of(render.Color{R: 0.95, G: 0.70, B: 0.23, A: 0.86}),
			EdgeColor: optional.Of(render.Color{R: 0.48, G: 0.27, B: 0.08, A: 1}),
			EdgeWidth: optional.Of(1.1),
			Hatch:     "/",
		},
		XY:     geom.Pt{X: 0.6, Y: 0.7},
		Width:  1.5,
		Height: 1.0,
	})
	left.AddPatch(&core.Circle{
		Patch: core.Patch{
			FaceColor: optional.Of(render.Color{R: 0.22, G: 0.57, B: 0.82, A: 0.82}),
			EdgeColor: optional.Of(render.Color{R: 0.11, G: 0.29, B: 0.44, A: 1}),
			EdgeWidth: optional.Of(1.0),
		},
		Center: geom.Pt{X: 3.0, Y: 1.25},
		Radius: 0.56,
	})
	left.AddPatch(&core.Ellipse{
		Patch: core.Patch{
			FaceColor: optional.Of(render.Color{R: 0.23, G: 0.72, B: 0.51, A: 0.80}),
			EdgeColor: optional.Of(render.Color{R: 0.10, G: 0.36, B: 0.24, A: 1}),
			EdgeWidth: optional.Of(1.0),
		},
		Center: geom.Pt{X: 4.8, Y: 2.75},
		Width:  1.55,
		Height: 0.95,
		Angle:  28,
	})
	left.AddPatch(&core.Polygon{
		Patch: core.Patch{
			FaceColor: optional.Of(render.Color{R: 0.84, G: 0.34, B: 0.34, A: 0.82}),
			EdgeColor: optional.Of(render.Color{R: 0.48, G: 0.14, B: 0.14, A: 1}),
			EdgeWidth: optional.Of(1.0),
		},
		XY: []geom.Pt{
			{X: 2.15, Y: 3.2},
			{X: 2.85, Y: 2.25},
			{X: 1.35, Y: 2.45},
		},
	})

	middle := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.37, Y: 0.16}, Max: geom.Pt{X: 0.63, Y: 0.88}})
	middle.SetTitle("Fancy Arrow + Path")
	middle.SetXLim(0, 6)
	middle.SetYLim(0, 4)
	middle.AddPatch(&core.FancyArrow{
		Patch: core.Patch{
			FaceColor: optional.Of(render.Color{R: 0.91, G: 0.42, B: 0.22, A: 0.88}),
			EdgeColor: optional.Of(render.Color{R: 0.58, G: 0.22, B: 0.10, A: 1}),
			EdgeWidth: optional.Of(1.0),
		},
		XY:                 geom.Pt{X: 0.9, Y: 3.2},
		DX:                 2.2,
		DY:                 -1.0,
		Width:              0.18,
		HeadWidth:          0.62,
		HeadLength:         0.62,
		LengthIncludesHead: true,
	})
	star := geom.Path{}
	star.MoveTo(geom.Pt{X: 4.15, Y: 0.95})
	star.LineTo(geom.Pt{X: 4.45, Y: 1.70})
	star.LineTo(geom.Pt{X: 5.22, Y: 1.75})
	star.LineTo(geom.Pt{X: 4.63, Y: 2.22})
	star.LineTo(geom.Pt{X: 4.84, Y: 2.96})
	star.LineTo(geom.Pt{X: 4.15, Y: 2.54})
	star.LineTo(geom.Pt{X: 3.46, Y: 2.96})
	star.LineTo(geom.Pt{X: 3.67, Y: 2.22})
	star.LineTo(geom.Pt{X: 3.08, Y: 1.75})
	star.LineTo(geom.Pt{X: 3.85, Y: 1.70})
	star.Close()
	middle.AddPatch(&core.PathPatch{
		Patch: core.Patch{
			FaceColor: optional.Of(render.Color{R: 0.76, G: 0.76, B: 0.86, A: 0.72}),
			EdgeColor: optional.Of(render.Color{R: 0.29, G: 0.29, B: 0.45, A: 1}),
			EdgeWidth: optional.Of(1.0),
			Hatch:     "x",
		},
		Path: star,
	})

	right := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.69, Y: 0.16}, Max: geom.Pt{X: 0.95, Y: 0.88}})
	right.SetTitle("Fancy Boxes")
	right.SetXLim(0, 6)
	right.SetYLim(0, 4)
	right.AddPatch(&core.FancyBboxPatch{
		Patch: core.Patch{
			FaceColor: optional.Of(render.Color{R: 0.29, G: 0.67, B: 0.78, A: 0.28}),
			EdgeColor: optional.Of(render.Color{R: 0.10, G: 0.37, B: 0.45, A: 1}),
			EdgeWidth: optional.Of(1.0),
			Hatch:     "/",
		},
		XY:           geom.Pt{X: 0.9, Y: 0.8},
		Width:        2.1,
		Height:       1.25,
		Pad:          0.14,
		BoxStyle:     core.BoxStyleRound,
		RoundingSize: 0.24,
	})
	right.AddPatch(&core.FancyBboxPatch{
		Patch: core.Patch{
			FaceColor: optional.Of(render.Color{R: 0.96, G: 0.87, B: 0.60, A: 0.82}),
			EdgeColor: optional.Of(render.Color{R: 0.50, G: 0.39, B: 0.12, A: 1}),
			EdgeWidth: optional.Of(1.0),
		},
		XY:       geom.Pt{X: 3.35, Y: 1.55},
		Width:    1.65,
		Height:   1.05,
		Pad:      0.10,
		BoxStyle: core.BoxStyleSquare,
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

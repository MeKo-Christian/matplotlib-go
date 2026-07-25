package transform_coordinates

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
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
		Min: geom.Pt{X: 0.13, Y: 0.16},
		Max: geom.Pt{X: 0.90, Y: 0.84},
	})
	ax.SetTitle("Transform Coordinates")
	ax.SetXLabel("X")
	ax.SetYLabel("Y")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	gridColor := render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	for _, grid := range []*core.Grid{ax.AddXGrid(), ax.AddYGrid()} {
		grid.Color = gridColor
		grid.LineWidth = 0.5
	}

	lineColor := render.Color{R: 0.14, G: 0.37, B: 0.74, A: 1}
	pointColor := render.Color{R: 0.88, G: 0.42, B: 0.16, A: 0.92}
	textColor := render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1}

	ax.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: 1.0, Y: 1.5},
			{X: 2.5, Y: 3.2},
			{X: 4.5, Y: 5.6},
			{X: 7.0, Y: 6.4},
			{X: 8.8, Y: 8.2},
		},
		W:   2.2,
		Col: lineColor,
	})
	ax.Add(&core.Scatter2D{
		XY: []geom.Pt{
			{X: 2.5, Y: 3.2},
			{X: 7.0, Y: 6.4},
			{X: 8.8, Y: 8.2},
		},
		Size:      core.ScatterAreaFromRadius(8.0, style.Default.DPI),
		Color:     pointColor,
		EdgeColor: render.Color{R: 0.45, G: 0.18, B: 0.05, A: 1},
		EdgeWidth: 1.0,
		Marker:    core.MarkerDiamond,
		Alpha:     1.0,
	})

	ax.Text(1.3, 1.1, "data", core.TextOptions{
		FontSize: 11,
		Color:    textColor,
		Coords:   core.Coords(core.CoordData),
	})
	ax.Text(0.03, 0.97, "axes", core.TextOptions{
		FontSize: 11,
		Color:    textColor,
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignTop,
		Coords:   core.Coords(core.CoordAxes),
	})
	fig.Text(0.07, 0.08, "figure", core.TextOptions{
		FontSize: 11,
		Color:    textColor,
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignBottom,
	})
	ax.Text(0.50, 0.22, "blend", core.TextOptions{
		FontSize: 11,
		Color:    textColor,
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignBottom,
		Coords:   core.BlendCoords(core.CoordFigure, core.CoordAxes),
	})
	ax.Annotate("axes note", 0.82, 0.78, core.AnnotationOptions{
		Coords:     core.Coords(core.CoordAxes),
		OffsetX:    optional.Of(-48.0),
		OffsetY:    optional.Of(-26.0),
		FontSize:   10,
		Color:      textColor,
		ArrowColor: textColor,
		ArrowWidth: optional.Of(1.25),
		HAlign:     core.TextAlignRight,
		VAlign:     core.TextVAlignTop,
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

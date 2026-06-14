package transform_annotation_modes

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 420
	DPI    = 100
)

// Plot builds an annotation coordinate-mode parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.15}, Max: geom.Pt{X: 0.88, Y: 0.82}})
	ax.SetTitle("Annotation Coordinate Modes")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	common.AddReferenceXYGrid(ax)

	lineColor := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	ax.Plot([]float64{1, 3, 5, 7, 9}, []float64{1.5, 4, 3, 7, 8.5}, core.PlotOptions{Color: &lineColor, LineWidth: &width})
	textColor := render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1}

	ax.Annotate("data", 3, 4, core.AnnotationOptions{
		Coords:     core.Coords(core.CoordData),
		OffsetX:    34,
		OffsetY:    -30,
		FontSize:   10,
		Color:      textColor,
		ArrowColor: textColor,
		ArrowWidth: 1.1,
	})
	ax.Annotate("axes", 0.78, 0.74, core.AnnotationOptions{
		Coords:     core.Coords(core.CoordAxes),
		OffsetX:    -46,
		OffsetY:    28,
		FontSize:   10,
		Color:      textColor,
		ArrowColor: textColor,
		ArrowWidth: 1.1,
		HAlign:     core.TextAlignRight,
	})
	ax.Annotate("figure", 0.72, 0.24, core.AnnotationOptions{
		Coords:     core.Coords(core.CoordFigure),
		OffsetX:    42,
		OffsetY:    24,
		FontSize:   10,
		Color:      textColor,
		ArrowColor: textColor,
		ArrowWidth: 1.1,
	})
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

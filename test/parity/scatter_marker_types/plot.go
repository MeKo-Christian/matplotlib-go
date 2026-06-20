package scatter_marker_types

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
		Min: geom.Pt{X: 0.06, Y: 0.08},
		Max: geom.Pt{X: 0.98, Y: 0.92},
	})
	ax.SetTitle("Marker Style Grid")
	ax.SetXLim(0.5, 7.5)
	ax.SetYLim(0.5, 6.5)

	styles := []core.MarkerStyle{
		core.NewMarkerStyle(core.MarkerPoint),
		core.NewMarkerStyle(core.MarkerPixel),
		core.NewMarkerStyle(core.MarkerCircle),
		core.NewMarkerStyle(core.MarkerTriangleDown),
		core.NewMarkerStyle(core.MarkerTriangleUp),
		core.NewMarkerStyle(core.MarkerTriangleLeft),
		core.NewMarkerStyle(core.MarkerTriangleRight),
		core.NewMarkerStyle(core.MarkerTriDown),
		core.NewMarkerStyle(core.MarkerTriUp),
		core.NewMarkerStyle(core.MarkerTriLeft),
		core.NewMarkerStyle(core.MarkerTriRight),
		core.NewMarkerStyle(core.MarkerOctagon),
		core.NewMarkerStyle(core.MarkerSquare),
		core.NewMarkerStyle(core.MarkerPentagon),
		core.NewMarkerStyle(core.MarkerFilledPlus),
		core.NewMarkerStyle(core.MarkerStar),
		core.NewMarkerStyle(core.MarkerHexagon1),
		core.NewMarkerStyle(core.MarkerHexagon2),
		core.NewMarkerStyle(core.MarkerPlus),
		core.NewMarkerStyle(core.MarkerCross),
		core.NewMarkerStyle(core.MarkerFilledX),
		core.NewMarkerStyle(core.MarkerDiamond),
		core.NewMarkerStyle(core.MarkerThinDiamond),
		core.NewMarkerStyle(core.MarkerVLine),
		core.NewMarkerStyle(core.MarkerHLine),
		core.NewMarkerStyle(core.MarkerTickLeft),
		core.NewMarkerStyle(core.MarkerTickRight),
		core.NewMarkerStyle(core.MarkerTickUp),
		core.NewMarkerStyle(core.MarkerTickDown),
		core.NewMarkerStyle(core.MarkerCaretLeft),
		core.NewMarkerStyle(core.MarkerCaretRight),
		core.NewMarkerStyle(core.MarkerCaretUp),
		core.NewMarkerStyle(core.MarkerCaretDown),
		core.NewMarkerStyle(core.MarkerCaretLeftBase),
		core.NewMarkerStyle(core.MarkerCaretRightBase),
		core.NewMarkerStyle(core.MarkerCaretUpBase),
		core.NewMarkerStyle(core.MarkerCaretDownBase),
		core.NewTupleMarkerStyle(5, core.MarkerTuplePolygon, 18),
		core.NewTupleMarkerStyle(5, core.MarkerTupleStar, 18),
		core.NewTupleMarkerStyle(6, core.MarkerTupleAsterisk, 0),
		core.NewMathTextMarkerStyle("$f$"),
		{Type: core.MarkerCircle, FillStyle: core.MarkerFillNone},
	}

	palette := style.Default.Palette()
	edge := render.Color{R: 0.05, G: 0.05, B: 0.05, A: 1}
	edgeWidth := 1.2
	for i, markerStyle := range styles {
		x := float64(i%7) + 1
		y := float64(6 - i/7)
		color := palette[i%len(palette)]
		scatter := &core.Scatter2D{
			XY:          []geom.Pt{{X: x, Y: y}},
			Size:        core.ScatterAreaFromRadius(8.0, style.Default.DPI),
			Color:       color,
			EdgeColor:   edge,
			EdgeWidth:   edgeWidth,
			MarkerStyle: markerStyle,
			Alpha:       1.0,
		}
		ax.Add(scatter)
	}
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

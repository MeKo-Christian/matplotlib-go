package line2d_markers

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
)

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.09, Y: 0.14},
		Max: geom.Pt{X: 0.94, Y: 0.88},
	})
	ax.SetTitle("Line2D Markers")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 6)

	buttCap := render.CapButt
	filled := core.MarkerCircle
	edgeAuto := core.AutoMarkerColor()
	ax.Plot([]float64{0.7, 2.0, 3.3}, []float64{5.35, 5.65, 5.35}, core.PlotOptions{
		Color:           colorPtr(render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}),
		LineWidth:       floatPtr(1.5),
		LineCap:         &buttCap,
		Marker:          &filled,
		MarkerSize:      floatPtr(9),
		MarkerFaceColor: colorPtr(render.Color{R: 1.00, G: 0.50, B: 0.05, A: 0.55}),
		MarkerEdgeSpec:  &edgeAuto,
		MarkerEdgeWidth: floatPtr(1.2),
		Label:           "filled auto edge",
	})

	unfilled := core.MarkerSquare
	faceNone := core.NoMarkerColor()
	ax.Plot([]float64{4.2, 5.5, 6.8}, []float64{5.35, 5.65, 5.35}, core.PlotOptions{
		Color:           colorPtr(render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1}),
		LineWidth:       floatPtr(1.5),
		LineCap:         &buttCap,
		Marker:          &unfilled,
		MarkerSize:      floatPtr(9),
		MarkerFaceSpec:  &faceNone,
		MarkerEdgeColor: colorPtr(render.Color{R: 0.08, G: 0.28, B: 0.08, A: 1}),
		MarkerEdgeWidth: floatPtr(1.4),
		Label:           "face none",
	})

	lineOnly := core.MarkerPlus
	ax.Plot([]float64{7.6, 8.7, 9.8}, []float64{5.35, 5.65, 5.35}, core.PlotOptions{
		Color:      colorPtr(render.Color{R: 0.84, G: 0.15, B: 0.16, A: 1}),
		LineWidth:  floatPtr(1.5),
		LineCap:    &buttCap,
		Marker:     &lineOnly,
		MarkerSize: floatPtr(11),
		Label:      "line-only",
	})

	custom := geom.Path{}
	custom.MoveTo(geom.Pt{X: 0, Y: -0.55})
	custom.LineTo(geom.Pt{X: 0.48, Y: 0.36})
	custom.LineTo(geom.Pt{X: -0.48, Y: 0.36})
	custom.Close()
	ax.Plot([]float64{0.8, 2.1, 3.4}, []float64{3.85, 4.2, 3.85}, core.PlotOptions{
		Color:           colorPtr(render.Color{R: 0.58, G: 0.40, B: 0.74, A: 1}),
		LineWidth:       floatPtr(1.4),
		LineCap:         &buttCap,
		MarkerPath:      &custom,
		MarkerSize:      floatPtr(10),
		MarkerFaceColor: colorPtr(render.Color{R: 0.58, G: 0.40, B: 0.74, A: 0.55}),
		MarkerEdgeColor: colorPtr(render.Color{R: 0.25, G: 0.12, B: 0.40, A: 1}),
		Label:           "custom path",
	})

	tuple := core.NewTupleMarkerStyle(5, core.MarkerTupleStar, 18)
	ax.Plot([]float64{4.3, 5.5, 6.7}, []float64{3.85, 4.2, 3.85}, core.PlotOptions{
		Color:           colorPtr(render.Color{R: 0.55, G: 0.34, B: 0.29, A: 1}),
		LineWidth:       floatPtr(1.4),
		LineCap:         &buttCap,
		MarkerStyle:     &tuple,
		MarkerSize:      floatPtr(11),
		MarkerFaceColor: colorPtr(render.Color{R: 0.55, G: 0.34, B: 0.29, A: 0.65}),
		MarkerEdgeColor: colorPtr(render.Color{R: 0.25, G: 0.12, B: 0.08, A: 1}),
		Label:           "tuple star",
	})

	mathMarker := core.NewMathTextMarkerStyle("$f$")
	ax.Plot([]float64{7.4, 8.6, 9.8}, []float64{3.85, 4.2, 3.85}, core.PlotOptions{
		Color:           colorPtr(render.Color{R: 0.50, G: 0.50, B: 0.50, A: 1}),
		LineWidth:       floatPtr(1.4),
		LineCap:         &buttCap,
		MarkerStyle:     &mathMarker,
		MarkerSize:      floatPtr(12),
		MarkerFaceColor: colorPtr(render.Color{R: 0.89, G: 0.47, B: 0.76, A: 0.65}),
		MarkerEdgeColor: colorPtr(render.Color{R: 0.35, G: 0.18, B: 0.32, A: 1}),
		Label:           "mathtext",
	})

	halfSpecs := []struct {
		x     float64
		fill  core.MarkerFillStyle
		label string
	}{
		{1.4, core.MarkerFillLeft, "left"},
		{3.4, core.MarkerFillRight, "right"},
		{5.4, core.MarkerFillTop, "top"},
		{7.4, core.MarkerFillBottom, "bottom"},
	}
	altFace := core.ExplicitMarkerColor(render.Color{R: 0.95, G: 0.78, B: 0.18, A: 1})
	for _, spec := range halfSpecs {
		style := core.NewMarkerStyle(core.MarkerCircle)
		style.FillStyle = spec.fill
		ax.Plot([]float64{spec.x - 0.4, spec.x, spec.x + 0.4}, []float64{2.0, 2.32, 2.0}, core.PlotOptions{
			Color:           colorPtr(render.Color{R: 0.09, G: 0.75, B: 0.81, A: 1}),
			LineWidth:       floatPtr(1.2),
			LineCap:         &buttCap,
			MarkerStyle:     &style,
			MarkerSize:      floatPtr(12),
			MarkerFaceColor: colorPtr(render.Color{R: 0.09, G: 0.75, B: 0.81, A: 1}),
			MarkerFaceAlt:   &altFace,
			MarkerEdgeColor: colorPtr(render.Color{R: 0.05, G: 0.25, B: 0.28, A: 1}),
			Label:           "half " + spec.label,
		})
	}

	legend := ax.AddLegend()
	legend.Location = core.LegendUpperLeft
	return fig
}

func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

func colorPtr(c render.Color) *render.Color { return &c }

func floatPtr(v float64) *float64 { return &v }

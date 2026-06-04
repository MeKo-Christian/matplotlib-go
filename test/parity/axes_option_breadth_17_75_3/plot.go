package axes_option_breadth_17_75_3

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 840
	Height = 540
)

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	scatterAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.58}, Max: geom.Pt{X: 0.46, Y: 0.92}})
	scatterAx.SetTitle("Scatter options")
	scatterAx.SetXLim(0, 6)
	scatterAx.SetYLim(0, 5)
	edgeColor := render.Color{R: 0.08, G: 0.08, B: 0.08, A: 1}
	edgeWidth := 1.5
	scatterAx.Scatter(
		[]float64{0.8, 1.8, 2.8, 3.8, 4.8},
		[]float64{1.2, 3.2, 2.2, 3.8, 1.8},
		core.ScatterOptions{
			ScalarValues: []float64{-1.0, -0.2, 0.35, 0.8, 1.2},
			Colormap:     "viridis",
			Sizes: []float64{
				core.ScatterAreaFromRadius(7, 100),
				core.ScatterAreaFromRadius(10, 100),
				core.ScatterAreaFromRadius(13, 100),
				core.ScatterAreaFromRadius(16, 100),
				core.ScatterAreaFromRadius(19, 100),
			},
			EdgeColor: &edgeColor,
			EdgeWidth: &edgeWidth,
		},
	)

	barAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.58}, Max: geom.Pt{X: 0.96, Y: 0.92}})
	barAx.SetTitle("Bar options")
	barAx.SetXLim(0, 4.2)
	barAx.SetYLim(0, 6.2)
	align := core.BarAlignEdge
	widths := []float64{0.42, 0.58, 0.36}
	barEdge := render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1}
	barEdgeWidth := 1.0
	base := barAx.Bar([]float64{0.6, 1.7, 2.9}, []float64{2.0, 3.1, 1.6}, core.BarOptions{
		Widths: widths,
		Colors: []render.Color{
			{R: 0.12, G: 0.47, B: 0.71, A: 1},
			{R: 1.00, G: 0.50, B: 0.05, A: 1},
			{R: 0.17, G: 0.63, B: 0.17, A: 1},
		},
		EdgeColor: &barEdge,
		EdgeWidth: &barEdgeWidth,
		Align:     &align,
	})
	barAx.BarLabel(base, nil, core.BarLabelOptions{Format: "%.1f", Padding: 2, FontSize: 8})
	top := barAx.Bar([]float64{0.6, 1.7, 2.9}, []float64{1.3, 1.0, 1.8}, core.BarOptions{
		Widths:    widths,
		Baselines: []float64{2.0, 3.1, 1.6},
		Color:     colorPtr(render.Color{R: 0.58, G: 0.40, B: 0.74, A: 1}),
		EdgeColor: &barEdge,
		EdgeWidth: &barEdgeWidth,
		Align:     &align,
	})
	barAx.BarLabel(top, nil, core.BarLabelOptions{Format: "%.1f", Padding: 2, FontSize: 8})

	fillAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.10}, Max: geom.Pt{X: 0.46, Y: 0.44}})
	fillAx.SetTitle("Fill options")
	fillAx.SetXLim(0, 5)
	fillAx.SetYLim(-1.4, 1.4)
	step := core.FillStepPost
	fillAlpha := 0.55
	fillAx.FillBetween(
		[]float64{0, 1, 2, 3, 4, 5},
		[]float64{-0.6, 0.8, 0.2, 1.0, -0.2, 0.7},
		[]float64{0.4, -0.4, 0.6, -0.3, 0.5, -0.5},
		core.FillOptions{
			Where:       []bool{false, true, true, false, true, true},
			Interpolate: true,
			Step:        step,
			Color:       colorPtr(render.Color{R: 0.84, G: 0.15, B: 0.16, A: 1}),
			Alpha:       &fillAlpha,
			EdgeColor:   colorPtr(render.Color{R: 0.50, G: 0.05, B: 0.06, A: 1}),
			EdgeWidth:   floatPtr(1.0),
		},
	)

	errorAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.10}, Max: geom.Pt{X: 0.96, Y: 0.44}})
	errorAx.SetTitle("Errorbar options")
	errorAx.SetXLim(0, 6)
	errorAx.SetYLim(0, 5)
	marker := core.MarkerCircle
	markerSize := 4.5
	capSize := 4.0
	errorColor := render.Color{R: 0.09, G: 0.75, B: 0.81, A: 1}
	errorAx.ErrorBar(
		[]float64{0.8, 1.6, 2.4, 3.2, 4.0, 4.8},
		[]float64{1.0, 2.0, 1.5, 3.3, 2.6, 4.0},
		[]float64{0.18, 0.24, 0.20, 0.30, 0.22, 0.26},
		[]float64{0.35, 0.42, 0.25, 0.50, 0.38, 0.45},
		core.ErrorBarOptions{
			Color:           &errorColor,
			CapSize:         &capSize,
			Marker:          &marker,
			MarkerSize:      &markerSize,
			ErrorEvery:      2,
			ErrorEveryStart: 1,
		},
	)

	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

func colorPtr(v render.Color) *render.Color { return &v }

func floatPtr(v float64) *float64 { return &v }

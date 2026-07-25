// Package widgets_gallery is the parity-test wrapper for the widgets_gallery showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/widgets_gallery.
package widgets_gallery

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/widgets"
)

const (
	Width  = 1100
	Height = 760
)

// Plot returns the parity figure with Matplotlib-compatible widget visuals.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))

	mainAx := fig.AddAxes(axesRect(0.06, 0.36, 0.64, 0.56))
	mainAx.SetTitle("Widgets and selectors")
	mainAx.SetXLabel("phase")
	mainAx.SetYLabel("amplitude")
	mainAx.SetXLim(0, 2*math.Pi)
	mainAx.SetYLim(-1.35, 1.35)
	gridColor := render.Color{R: 0.85, G: 0.85, B: 0.85, A: 1}
	gridWidth := 0.8
	yGrid := mainAx.AddYGrid()
	yGrid.Color = gridColor
	yGrid.LineWidth = gridWidth

	x := make([]float64, 220)
	signal := make([]float64, len(x))
	modulation := make([]float64, len(x))
	for i := range x {
		x[i] = 2 * math.Pi * float64(i) / float64(len(x)-1)
		signal[i] = math.Sin(x[i])
		modulation[i] = 0.55 * math.Cos(1.7*x[i])
	}
	blue := render.Color{R: 0.12, G: 0.34, B: 0.68, A: 1}
	orange := render.Color{R: 0.84, G: 0.35, B: 0.18, A: 1}
	lwSignal := 2.2
	lwMod := 1.7
	mainAx.Plot(x, signal, core.PlotOptions{Color: &blue, LineWidth: &lwSignal, Label: "signal"})
	mainAx.Plot(x, modulation, core.PlotOptions{Color: &orange, LineWidth: &lwMod, Label: "modulation"})
	mainAx.AddLegend()

	spanAlpha := 0.18
	mainAx.AxVSpan(0.7, 1.35, core.VSpanOptions{Color: &blue, Alpha: &spanAlpha})
	mainAx.AddPatch(&core.Rectangle{
		Patch: core.Patch{FaceColor: colorWithAlpha(blue, 0.20), EdgeColor: colorWithAlpha(blue, 0.20), EdgeWidth: 1},
		XY:    geom.Pt{X: 2.25, Y: -0.95}, Width: 0.8, Height: 0.7,
	})
	mainAx.AddPatch(&core.Ellipse{
		Patch:  core.Patch{FaceColor: colorWithAlpha(orange, 0.22), EdgeColor: colorWithAlpha(orange, 0.22), EdgeWidth: 1},
		Center: geom.Pt{X: 4.0, Y: 0.625}, Width: 0.9, Height: 0.75,
	})
	green := render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1}
	mainAx.AddPatch(&core.Polygon{
		Patch: core.Patch{FaceColor: colorWithAlpha(green, 0.22), EdgeColor: green, EdgeWidth: 1},
		XY: []geom.Pt{
			{X: 4.95, Y: -0.9},
			{X: 5.75, Y: -0.75},
			{X: 5.45, Y: -0.15},
		},
	})
	lasso := []geom.Pt{
		{X: 1.2, Y: 0.75},
		{X: 1.45, Y: 1.00},
		{X: 1.85, Y: 0.92},
		{X: 2.05, Y: 0.63},
		{X: 1.75, Y: 0.42},
		{X: 1.34, Y: 0.50},
	}
	mainAx.Add(&core.Line2D{XY: lasso, Col: render.Color{R: 0.58, G: 0.40, B: 0.74, A: 1}, W: 1.4})
	refColor := render.Color{R: 0.2, G: 0.2, B: 0.2, A: 1}
	refAlpha := 0.75
	refWidth := 1.0
	mainAx.AxVLine(2.8, core.VLineOptions{Color: &refColor, LineWidth: &refWidth, Alpha: &refAlpha})
	mainAx.AxHLine(0.35, core.HLineOptions{Color: &refColor, LineWidth: &refWidth, Alpha: &refAlpha})

	auxAx := fig.AddAxes(axesRect(0.76, 0.36, 0.18, 0.56))
	auxAx.XAxis = mainAx.XAxis
	auxAx.SetTitle("shared cursor")
	auxAx.SetXLim(0, 2*math.Pi)
	auxAx.SetYLim(-1, 1)
	auxGrid := auxAx.AddYGrid()
	auxGrid.Color = gridColor
	auxGrid.LineWidth = gridWidth
	auxAx.Plot(x, modulation, core.PlotOptions{Color: &orange, LineWidth: &lwMod})
	auxAx.AxVLine(4.45, core.VLineOptions{Color: &refColor, LineWidth: &refWidth, Alpha: &refAlpha})
	auxAx.AxHLine(0.0, core.HLineOptions{Color: &refColor, LineWidth: &refWidth, Alpha: &refAlpha})

	widgets.NewButton(widgets.NewAxes(fig, axesRect(0.06, 0.23, 0.16, 0.07)), "Apply")
	widgets.NewSlider(widgets.NewAxes(fig, axesRect(0.28, 0.23, 0.26, 0.07)), "gain", 0, 1, 0.68)
	widgets.NewRangeSlider(widgets.NewAxes(fig, axesRect(0.59, 0.23, 0.23, 0.07)), "window", 0, 1, 0.22, 0.78)
	widgets.NewTextBox(widgets.NewAxes(fig, axesRect(0.86, 0.23, 0.10, 0.07)), "label", "phase scan")
	widgets.NewCheckButtons(widgets.NewAxes(fig, axesRect(0.06, 0.07, 0.36, 0.12)), []string{"signal", "modulation", "grid"}, []bool{true, true, false})
	widgets.NewRadioButtons(widgets.NewAxes(fig, axesRect(0.55, 0.07, 0.28, 0.12)), []string{"blue", "amber", "mono"}, 1)

	return fig
}

// Render returns the parity image with Matplotlib-compatible widget visuals.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

func axesRect(left, bottom, width, height float64) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: left, Y: bottom},
		Max: geom.Pt{X: left + width, Y: bottom + height},
	}
}

func colorWithAlpha(color render.Color, alpha float64) render.Color {
	color.A = alpha
	return color
}

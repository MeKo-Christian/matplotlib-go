// Package ticks_scales_formatters_gallery is a user-facing gallery for tick,
// scale, formatter, date, category, and custom-unit behavior.
package ticks_scales_formatters_gallery

import (
	"fmt"
	"image"
	"time"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/dates"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

const (
	Width  = 1320
	Height = 900
	DPI    = 100
)

var (
	blue   = render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	orange = render.Color{R: 1.00, G: 0.50, B: 0.05, A: 1}
	green  = render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1}
	purple = render.Color{R: 0.58, G: 0.40, B: 0.74, A: 1}
	brown  = render.Color{R: 0.55, G: 0.34, B: 0.29, A: 1}
	gray   = render.Color{R: 0.50, G: 0.50, B: 0.50, A: 1}
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	common.RegisterTestDistanceUnits()

	fig := core.NewFigure(Width, Height)
	fig.Text(0.05, 0.955, "Ticks, Scales, and Formatters", core.TextOptions{FontSize: 15})
	fig.Text(0.05, 0.928, "locators, scale defaults, formatter families, dates, categories, and custom units", core.TextOptions{FontSize: 11})

	addLocatorPanel(fig, panelRect(0, 0))
	addLogPanel(fig, panelRect(0, 1))
	addScalePanel(fig, panelRect(1, 0))
	addFormatterPanel(fig, panelRect(1, 1))
	addDateCategoryPanel(fig, panelRect(2, 0))
	addCustomUnitPanel(fig, panelRect(2, 1))

	return fig
}

func panelRect(row, col int) geom.Rect {
	x0 := 0.07 + float64(col)*0.46
	y0 := 0.66 - float64(row)*0.29
	return geom.Rect{
		Min: geom.Pt{X: x0, Y: y0},
		Max: geom.Pt{X: x0 + 0.38, Y: y0 + 0.18},
	}
}

func configureAxes(ax *core.Axes, title, xlabel, ylabel string) {
	ax.SetTitle(title)
	ax.SetXLabel(xlabel)
	ax.SetYLabel(ylabel)
	common.AddReferenceYGrid(ax)
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
}

func addLocatorPanel(fig *core.Figure, rect geom.Rect) {
	ax := fig.AddAxes(rect)
	configureAxes(ax, "Major and Minor Locators", "MultipleLocator + minor ticks", "score")
	_, _ = ax.Plot(
		[]float64{0, 1.5, 3, 4.5, 6},
		[]float64{0.12, 0.38, 0.58, 0.74, 0.90},
		core.PlotOptions{Color: &blue, LineWidth: ptr(2.0)},
	)
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 1)
	ax.XAxis.Locator = ticker.MultipleLocator{Base: 1.5}
	ax.XAxis.MinorLocator = ticker.AutoMinorLocator{N: 3}
	xGrid := ax.AddXGrid()
	xGrid.Minor = true
	xGrid.MinorLineWidth = 0.35
	xGrid.MinorColor = render.Color{R: 0.72, G: 0.76, B: 0.80, A: 0.55}
	xGrid.MinorDashes = []float64{1, 2}
}

func addLogPanel(fig *core.Figure, rect geom.Rect) {
	ax := fig.AddAxes(rect)
	configureAxes(ax, "Log Scale and Minor Grid", "base-10 log", "score")
	_, _ = ax.Plot(
		[]float64{1, 3, 10, 30, 100, 300, 1000},
		[]float64{0.10, 0.22, 0.38, 0.55, 0.70, 0.82, 0.91},
		core.PlotOptions{Color: &orange, LineWidth: ptr(2.0)},
	)
	_ = ax.SetXScale("log", transform.WithScaleBase(10))
	ax.SetXLim(1, 1000)
	ax.SetYLim(0, 1)
	ax.XAxis.MinorLocator = ticker.LogLocator{Base: 10, SubsMode: "auto"}
	ax.XAxis.Formatter = ticker.LogFormatterMathText{Base: 10}
	xGrid := ax.AddXGrid()
	xGrid.Minor = true
	xGrid.MinorLineWidth = 0.35
	xGrid.MinorColor = render.Color{R: 0.72, G: 0.76, B: 0.80, A: 0.55}
	xGrid.MinorDashes = []float64{1, 2}
}

func addScalePanel(fig *core.Figure, rect geom.Rect) {
	ax := fig.AddAxes(rect)
	configureAxes(ax, "Signed Scale Defaults", "symlog with signed markers", "response")
	x := []float64{-1000, -100, -10, -1, 0, 1, 10, 100, 1000}
	_, _ = ax.Plot(x, []float64{0.12, 0.21, 0.32, 0.44, 0.51, 0.59, 0.70, 0.82, 0.91}, core.PlotOptions{Color: &green, LineWidth: ptr(2.0)})
	ax.Scatter(x, []float64{0.15, 0.24, 0.35, 0.47, 0.54, 0.62, 0.73, 0.85, 0.94}, core.ScatterOptions{Color: &purple, Size: ptr(core.ScatterAreaFromRadius(4.5, DPI))})
	_ = ax.SetXScale("symlog", transform.WithScaleBase(10), transform.WithScaleLinThresh(1))
	ax.SetXLim(-1000, 1000)
	ax.SetYLim(0, 1)
}

func addFormatterPanel(fig *core.Figure, rect geom.Rect) {
	ax := fig.AddAxes(rect)
	configureAxes(ax, "Formatter Families", "position", "formatted values")
	x := []float64{0, 1, 2, 3, 4}
	_, _ = ax.Plot(x, []float64{0.15, 0.32, 0.48, 0.66, 0.86}, core.PlotOptions{Color: &purple, LineWidth: ptr(2.0)})
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 1)
	ax.XAxis.Locator = ticker.FixedLocator{TicksList: x}
	ax.XAxis.Formatter = ticker.FixedFormatter{Labels: []string{"0", "1 kHz", "25%", "1.2e3", "custom"}}
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0.15, 0.50, 0.85}}
	ax.YAxis.Formatter = ticker.FuncFormatter(func(v float64) string {
		return fmt.Sprintf("y=%0.2f", v)
	})
}

func addDateCategoryPanel(fig *core.Figure, rect geom.Rect) {
	ax := fig.AddAxes(rect)
	configureAxes(ax, "Date and Category Formatters", "date axis with category labels", "requests")
	dateValues := []time.Time{
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC),
	}
	if _, err := ax.Plot(dateValues, []float64{0.08, 0.38, 0.30, 0.48, 0.42}, core.PlotOptions{Color: &brown, LineWidth: ptr(2.0)}); err != nil {
		panic(err)
	}
	ax.XAxis.Locator = dates.DayLocator{ByMonthDay: []int{1, 7, 14, 21}, Location: time.UTC}
	ax.XAxis.Formatter = dates.DateFormatter{Layout: "02 Jan", Location: time.UTC}
	ax.SetYLim(0, 1)
	ax.AutoScale(0.04)

	right := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.30, Y: 0.16},
		Max: geom.Pt{X: 0.43, Y: 0.30},
	})
	right.SetTitle("Categories")
	right.XAxis.ShowLabels = false
	right.YAxis.ShowLabels = true
	if _, err := right.BarUnits([]string{"draft", "review", "ship"}, []float64{0.35, 0.75, 0.55}, core.BarOptions{Color: &gray, Width: ptr(0.72)}); err != nil {
		panic(err)
	}
	right.SetYLim(0, 1)
}

func addCustomUnitPanel(fig *core.Figure, rect geom.Rect) {
	ax := fig.AddAxes(rect)
	configureAxes(ax, "Custom Unit Converter", "distance", "pace")
	distances := []common.TestDistanceKM{5, 10, 21.1, 30, 42.2}
	pace := []float64{0.75, 0.69, 0.58, 0.52, 0.60}
	if _, err := ax.Plot(distances, pace, core.PlotOptions{Color: &blue, LineWidth: ptr(2.0)}); err != nil {
		panic(err)
	}
	if _, err := ax.Scatter(distances, pace, core.ScatterOptions{Color: &green, EdgeColor: &blue, Size: ptr(core.ScatterAreaFromRadius(5.0, DPI))}); err != nil {
		panic(err)
	}
	ax.SetXLim(3, 44)
	ax.SetYLim(0, 1)
	ax.XAxis.Formatter = ticker.FuncFormatter(func(v float64) string {
		return fmt.Sprintf("%g km", v)
	})
	ax.YAxis.Formatter = ticker.PercentFormatter{XMax: 1, DisplayRange: 1}
}

func ptr(v float64) *float64 { return &v }

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

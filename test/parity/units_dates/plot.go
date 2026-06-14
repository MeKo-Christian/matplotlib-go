package units_dates

import (
	"image"
	"time"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 380
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(720, 380)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.18}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("Date Units")
	ax.SetXLabel("Date")
	ax.SetYLabel("Requests")
	common.AddReferenceYGrid(ax)

	dates := []time.Time{
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC),
	}
	lower := []float64{6, 7, 5, 8, 7}
	upper := []float64{10, 15, 13, 18, 16}
	fillColor := render.Color{R: 0.85, G: 0.91, B: 0.96, A: 1}
	lineColor := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	lineWidth := 2.0
	if _, err := ax.FillBetweenUnits(dates, lower, upper, core.FillOptions{Color: &fillColor}); err != nil {
		panic(err)
	}
	if _, err := ax.PlotUnits(dates, []float64{8, 12, 9, 15, 13}, core.PlotOptions{Color: &lineColor, LineWidth: &lineWidth}); err != nil {
		panic(err)
	}
	ax.XAxis.Locator = core.DayLocator{ByMonthDay: []int{5, 12, 19}, Location: time.UTC}
	ax.XAxis.Formatter = core.DateFormatter{Layout: "02 Jan", Location: time.UTC}
	ax.AutoScale(0.06)
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, Width, Height)
}

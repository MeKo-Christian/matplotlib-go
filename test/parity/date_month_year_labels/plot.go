package date_month_year_labels

import (
	"image"
	"time"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 360
	DPI    = 100
)

// Plot builds a month/year date-label parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.18}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("Monthly + Yearly Date Labels")
	ax.SetXLabel("month")
	ax.SetYLabel("index")
	common.AddReferenceYGrid(ax)

	dates := []time.Time{
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	y := []float64{2, 4, 3, 7}
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	if _, err := ax.PlotUnits(dates, y, core.PlotOptions{Color: &color, LineWidth: &width}); err != nil {
		panic(err)
	}
	ax.SetXLim(common.ReferenceDateNumber(dates[0]), common.ReferenceDateNumber(dates[len(dates)-1]))
	ax.SetYLim(0, 8)
	ax.XAxis.Locator = core.MonthLocator{ByMonth: []time.Month{time.January, time.July}, Location: time.UTC}
	ax.XAxis.Formatter = core.DateFormatter{Layout: "Jan 2006", Location: time.UTC}
	ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 4, 8}}
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

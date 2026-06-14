package date_concise_intraday_labels

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

// Plot builds an intraday concise-date formatter parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.18}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("Concise Intraday Dates")
	ax.SetXLabel("time")
	ax.SetYLabel("requests")
	common.AddReferenceYGrid(ax)

	start := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	times := []time.Time{
		start,
		start.Add(6 * time.Hour),
		start.Add(12 * time.Hour),
		start.Add(18 * time.Hour),
	}
	y := []float64{8, 11, 9, 14}
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	if _, err := ax.PlotUnits(times, y, core.PlotOptions{Color: &color, LineWidth: &width}); err != nil {
		panic(err)
	}
	ax.SetXLim(common.ReferenceDateNumber(times[0]), common.ReferenceDateNumber(times[len(times)-1]))
	ax.SetYLim(0, 16)
	ax.XAxis.Locator = core.HourLocator{ByHour: []int{0, 6, 12, 18}, Location: time.UTC}
	ax.XAxis.Formatter = core.ConciseDateFormatter{Location: time.UTC}
	ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 8, 16}}
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

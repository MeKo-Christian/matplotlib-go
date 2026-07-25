package units_overview

import (
	"image"
	"time"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const (
	Width  = 1200
	Height = 420
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	common.RegisterTestDistanceUnits()

	fig := core.NewFigure(1200, 420)

	dateAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.06, Y: 0.18},
		Max: geom.Pt{X: 0.32, Y: 0.86},
	})
	dateAx.SetTitle("Dates")
	dateAx.SetYLabel("Requests")
	common.AddReferenceYGrid(dateAx)
	_, err := dateAx.Plot(
		[]time.Time{
			time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 7, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC),
		},
		[]float64{12, 18, 9, 21},
		core.PlotOptions{
			Color:     &render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1},
			LineWidth: common.FloatPtr(2.0),
		},
	)
	if err != nil {
		panic(err)
	}
	dateAx.AutoScale(0.05)

	categoryAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.38, Y: 0.18},
		Max: geom.Pt{X: 0.64, Y: 0.86},
	})
	categoryAx.SetTitle("Categories")
	categoryAx.SetYLabel("Count")
	common.AddReferenceYGrid(categoryAx)
	_, err = categoryAx.BarUnits(
		[]string{"alpha", "beta", "gamma", "delta"},
		[]float64{4, 9, 6, 7},
		core.BarOptions{
			Color:     &render.Color{R: 1.0, G: 0.50, B: 0.05, A: 1},
			EdgeColor: &render.Color{R: 0.60, G: 0.30, B: 0.03, A: 1},
			EdgeWidth: common.FloatPtr(1.0),
		},
	)
	if err != nil {
		panic(err)
	}
	categoryAx.AutoScale(0.10)

	unitAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.70, Y: 0.18},
		Max: geom.Pt{X: 0.96, Y: 0.86},
	})
	unitAx.SetTitle("Custom Units")
	unitAx.SetXLabel("Distance")
	unitAx.SetYLabel("Pace")
	common.AddReferenceXYGrid(unitAx)
	_, err = unitAx.ScatterUnits(
		[]common.TestDistanceKM{5, 10, 21.1, 42.2},
		[]float64{6.4, 5.8, 5.2, 5.5},
		core.ScatterOptions{
			Color:     &render.Color{R: 0.17, G: 0.63, B: 0.17, A: 0.92},
			EdgeColor: &render.Color{R: 0.09, G: 0.36, B: 0.09, A: 1},
			EdgeWidth: common.FloatPtr(1.0),
			Size:      common.FloatPtr(core.ScatterAreaFromRadius(8.0, style.Default.DPI)),
		},
	)
	if err != nil {
		panic(err)
	}
	unitAx.AutoScale(0.08)
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

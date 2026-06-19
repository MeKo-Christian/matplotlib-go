package spectrum_variants

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 900
	Height = 640
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(900, 640)
	x := spectrumSignal()
	width := 1.8

	magAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.72}, Max: geom.Pt{X: 0.96, Y: 0.93}})
	magAx.SetTitle("Magnitude Spectrum")
	magAx.MagnitudeSpectrum(x, core.SignalSpectrumOptions{
		Fs:     64,
		Window: "none",
		Scale:  core.SignalSpectrumScaleDB,
		PlotOptions: core.PlotOptions{
			Color:     &render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1},
			LineWidth: &width,
		},
	})
	magAx.AddYGrid()

	angleAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.41}, Max: geom.Pt{X: 0.96, Y: 0.62}})
	angleAx.SetTitle("Angle Spectrum")
	angleAx.AngleSpectrum(x, core.SignalSpectrumOptions{
		Fs:     64,
		Fc:     4,
		Window: "none",
		Sides:  core.SignalSpectrumSidesTwoSided,
		PlotOptions: core.PlotOptions{
			Color:     &render.Color{R: 1.00, G: 0.50, B: 0.05, A: 1},
			LineWidth: &width,
		},
	})
	angleAx.AddYGrid()

	phaseAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.10}, Max: geom.Pt{X: 0.96, Y: 0.31}})
	phaseAx.SetTitle("Phase Spectrum")
	phaseAx.PhaseSpectrum(x, core.SignalSpectrumOptions{
		Fs:     64,
		Window: "none",
		Sides:  core.SignalSpectrumSidesOneSided,
		PlotOptions: core.PlotOptions{
			Color:     &render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1},
			LineWidth: &width,
		},
	})
	phaseAx.AddYGrid()
	return fig
}

func spectrumSignal() []float64 {
	x := make([]float64, 128)
	for i := range x {
		t := float64(i) / 64.0
		x[i] = math.Sin(2*math.Pi*5.25*t) + 0.35*math.Cos(2*math.Pi*12.4*t+0.4)
	}
	return x
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

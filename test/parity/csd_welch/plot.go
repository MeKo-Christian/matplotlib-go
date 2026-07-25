package csd_welch

import (
	"image"
	"math"

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
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.125, Y: 0.11}, Max: geom.Pt{X: 0.9, Y: 0.88}})
	ax.SetTitle("Welch Cross-Spectral Density")
	x, y := signals()
	width := 1.5
	ax.CSD(x, y, core.SignalSpectrumOptions{
		Fs: 64, Fc: 2, NFFT: 64, NOverlap: 32, PadTo: 128,
		Window: "hann", Detrend: core.SignalDetrendMean,
		PlotOptions: core.PlotOptions{Color: &render.Color{R: 0.84, G: 0.15, B: 0.16, A: 1}, LineWidth: &width},
	})
	return fig
}

func signals() ([]float64, []float64) {
	x := make([]float64, 256)
	y := make([]float64, 256)
	for i := range x {
		t := float64(i) / 64
		x[i] = math.Sin(2*math.Pi*8*t) + 0.35*math.Sin(2*math.Pi*15*t+0.2) + 0.1*math.Cos(2*math.Pi*3*t)
		y[i] = 0.8*math.Sin(2*math.Pi*8*t+0.45) + 0.25*math.Sin(2*math.Pi*15*t-0.3) + 0.12*math.Cos(2*math.Pi*20*t)
	}
	return x, y
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

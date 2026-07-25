package specgram_psd

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
	ax.SetTitle("PSD Spectrogram")
	ax.SetXLabel("Time")
	ax.SetYLabel("Frequency")
	cmap := "viridis"
	vmin, vmax := -55.0, -5.0
	ax.Specgram(signal(), core.SpecgramOptions{
		Fs: 64, NFFT: 64, NOverlap: 48, PadTo: 128,
		Window: "hann", Colormap: &cmap, VMin: &vmin, VMax: &vmax,
	})
	return fig
}

func signal() []float64 {
	x := make([]float64, 384)
	for i := range x {
		t := float64(i) / 64
		phase := 2 * math.Pi * (3*t + 0.75*t*t)
		x[i] = 0.8*math.Sin(phase) + 0.25*math.Sin(2*math.Pi*18*t)
	}
	return x
}

func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

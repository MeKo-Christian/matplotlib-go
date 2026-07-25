// Package image_variants_gallery collects image interpolation, alpha, matshow,
// and spy variants into a focused user-facing gallery.
package image_variants_gallery

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1080
	Height = 720
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	drawInterpolationPanel(fig, 0, "nearest")
	drawInterpolationPanel(fig, 1, "bilinear")
	drawInterpolationPanel(fig, 2, "bicubic")
	drawAlphaPanel(fig)
	drawMatshowPanel(fig)
	drawSpyPanels(fig)
	fig.Text(0.05, 0.96, "Image Variants Gallery", core.TextOptions{FontSize: 13})
	return fig
}

func drawInterpolationPanel(fig *core.Figure, index int, mode string) {
	const (
		left = 0.06
		gap  = 0.035
		w    = 0.27
	)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: left + float64(index)*(w+gap), Y: 0.58},
		Max: geom.Pt{X: left + float64(index)*(w+gap) + w, Y: 0.88},
	})
	ax.SetTitle(mode)
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	cmap := "viridis"
	vmin, vmax := 0.0, 1.0
	extent := [4]float64{0, 16, 0, 16}
	ax.ImShow(imageData(16), core.ImShowOptions{
		Colormap:      &cmap,
		VMin:          &vmin,
		VMax:          &vmax,
		Origin:        core.ImageOriginLower,
		Extent:        &extent,
		Aspect:        "auto",
		Interpolation: &mode,
	})
}

func drawAlphaPanel(fig *core.Figure) {
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.06, Y: 0.12}, Max: geom.Pt{X: 0.31, Y: 0.43}})
	ax.SetTitle("alpha overlay")
	ax.SetXLim(0, 18)
	ax.SetYLim(0, 18)
	baseCmap := "Greys"
	overlayCmap := "magma"
	alpha := 0.58
	extent := [4]float64{0, 18, 0, 18}
	ax.ImShow(checkerData(18), core.ImShowOptions{Colormap: &baseCmap, Origin: core.ImageOriginLower, Extent: &extent, Aspect: "auto"})
	ax.ImShow(radialData(18), core.ImShowOptions{Colormap: &overlayCmap, Alpha: &alpha, Origin: core.ImageOriginLower, Extent: &extent, Aspect: "auto"})
}

func drawMatshowPanel(fig *core.Figure) {
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.39, Y: 0.12}, Max: geom.Pt{X: 0.61, Y: 0.43}})
	ax.SetTitle("matshow")
	cmap := "plasma"
	ax.MatShow(matshowData(), core.MatShowOptions{Colormap: &cmap})
}

func drawSpyPanels(fig *core.Figure) {
	markerAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.68, Y: 0.12}, Max: geom.Pt{X: 0.80, Y: 0.43}})
	markerAx.SetTitle("spy markers")
	marker := core.MarkerSquare
	color := render.Color{R: 0.12, G: 0.38, B: 0.70, A: 1}
	markerAx.Spy(sparseData(18), core.SpyOptions{Precision: 0.1, Marker: &marker, MarkerSize: 7, Color: &color})

	imageAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.84, Y: 0.12}, Max: geom.Pt{X: 0.96, Y: 0.43}})
	imageAx.SetTitle("spy image")
	useImage := true
	imageAx.Spy(sparseData(18), core.SpyOptions{Precision: 0.1, UseImage: &useImage})
}

func imageData(n int) [][]float64 {
	data := make([][]float64, n)
	for y := range data {
		data[y] = make([]float64, n)
		for x := range data[y] {
			checker := 0.0
			if (x+y)%2 == 0 {
				checker = 0.35
			}
			wave := 0.25 * math.Sin(float64(x)*0.95) * math.Cos(float64(y)*0.7)
			gradient := 0.35 * float64(x+y) / float64(2*(n-1))
			data[y][x] = clampUnit(0.20 + checker + wave + gradient)
		}
	}
	return data
}

func checkerData(n int) [][]float64 {
	data := make([][]float64, n)
	for y := range data {
		data[y] = make([]float64, n)
		for x := range data[y] {
			if (x/3+y/3)%2 == 0 {
				data[y][x] = 0.25
			} else {
				data[y][x] = 0.85
			}
		}
	}
	return data
}

func radialData(n int) [][]float64 {
	data := make([][]float64, n)
	for y := range data {
		data[y] = make([]float64, n)
		yy := 2*float64(y)/float64(n-1) - 1
		for x := range data[y] {
			xx := 2*float64(x)/float64(n-1) - 1
			data[y][x] = math.Exp(-3.5*(xx*xx+yy*yy)) + 0.2*math.Sin(10*xx)
		}
	}
	return data
}

func matshowData() [][]float64 {
	return [][]float64{
		{0.2, 0.3, 0.5, 0.8, 0.6},
		{0.1, 0.6, 0.7, 0.4, 0.3},
		{0.9, 0.8, 0.2, 0.3, 0.5},
		{0.4, 0.2, 0.6, 0.9, 0.7},
	}
}

func sparseData(n int) [][]float64 {
	data := make([][]float64, n)
	for y := range data {
		data[y] = make([]float64, n)
		for x := range data[y] {
			if x == y || x+y == n-1 || (x+2*y)%7 == 0 || (2*x+y)%11 == 0 {
				data[y][x] = 1
			}
		}
	}
	return data
}

func clampUnit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
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

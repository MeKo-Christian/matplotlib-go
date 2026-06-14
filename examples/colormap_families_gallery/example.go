// Package colormap_families_gallery shows representative colormap families as
// compact scalar strips.
package colormap_families_gallery

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
	Height = 520
	DPI    = 100
)

type cmapRow struct {
	Title string
	Name  string
	Data  [][]float64
	VMin  float64
	VMax  float64
}

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	rows := []cmapRow{
		{Title: "sequential: viridis", Name: "viridis", Data: gradientStrip(3, 80, false), VMin: 0, VMax: 1},
		{Title: "sequential reversed: viridis_r", Name: "viridis_r", Data: gradientStrip(3, 80, false), VMin: 0, VMax: 1},
		{Title: "perceptual: plasma", Name: "plasma", Data: gradientStrip(3, 80, false), VMin: 0, VMax: 1},
		{Title: "diverging: RdBu", Name: "RdBu", Data: divergingStrip(3, 80), VMin: -1, VMax: 1},
		{Title: "qualitative: tab10", Name: "tab10", Data: categoricalStrip(3, 80, 10), VMin: 0, VMax: 9},
		{Title: "cyclic: twilight", Name: "twilight", Data: cyclicStrip(3, 80), VMin: 0, VMax: 1},
	}

	const (
		left   = 0.35
		right  = 0.95
		top    = 0.88
		height = 0.09
		gap    = 0.045
	)
	for i, row := range rows {
		y1 := top - float64(i)*(height+gap)
		ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: left, Y: y1 - height}, Max: geom.Pt{X: right, Y: y1}})
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
		ax.YAxis.ShowTicks = false
		ax.YAxis.ShowLabels = false
		extent := [4]float64{0, float64(len(row.Data[0])), 0, float64(len(row.Data))}
		nearest := "nearest"
		cmap := row.Name
		vmin, vmax := row.VMin, row.VMax
		ax.ImShow(row.Data, core.ImShowOptions{
			Colormap:      &cmap,
			VMin:          &vmin,
			VMax:          &vmax,
			Origin:        core.ImageOriginLower,
			Extent:        &extent,
			Aspect:        "auto",
			Interpolation: &nearest,
		})
		fig.Text(0.06, y1-height/2, row.Title, core.TextOptions{
			HAlign:   core.TextAlignLeft,
			VAlign:   core.TextVAlignMiddle,
			FontSize: 10,
		})
	}
	fig.Text(0.06, 0.94, "Colormap Family Gallery", core.TextOptions{FontSize: 13})
	return fig
}

func gradientStrip(rows, cols int, reverse bool) [][]float64 {
	data := make([][]float64, rows)
	for y := range data {
		data[y] = make([]float64, cols)
		for x := range data[y] {
			v := float64(x) / float64(cols-1)
			if reverse {
				v = 1 - v
			}
			data[y][x] = v
		}
	}
	return data
}

func divergingStrip(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for y := range data {
		data[y] = make([]float64, cols)
		for x := range data[y] {
			data[y][x] = 2*float64(x)/float64(cols-1) - 1
		}
	}
	return data
}

func categoricalStrip(rows, cols, n int) [][]float64 {
	data := make([][]float64, rows)
	for y := range data {
		data[y] = make([]float64, cols)
		for x := range data[y] {
			data[y][x] = math.Floor(float64(x) * float64(n) / float64(cols))
		}
	}
	return data
}

func cyclicStrip(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for y := range data {
		data[y] = make([]float64, cols)
		for x := range data[y] {
			phase := float64(x) / float64(cols-1)
			data[y][x] = math.Mod(phase+0.08*math.Sin(phase*2*math.Pi), 1)
		}
	}
	return data
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

package main

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})

	const n = 64
	data := make([][]float64, n)
	for j := range n {
		row := make([]float64, n)
		for i := range n {
			row[i] = float64(i*j) / float64(n*n)
		}
		data[j] = row
	}

	cmap := "viridis"
	ax.SetTitle("ImShow with Extent + Bilinear Interpolation")
	ax.ImShow(data, core.ImShowOptions{
		Colormap:      &cmap,
		Aspect:        "equal",
		Extent:        &[4]float64{-2, 2, -1, 1},
		Interpolation: ptr("bilinear"),
	})

	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      640,
		Height:     360,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if createErr != nil {
		fmt.Printf("error creating renderer: %v\n", createErr)
		return
	}

	if err := core.SavePNG(fig, r, "imshow_extent.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}
	fmt.Println("saved imshow_extent.png")
}

func ptr[T any](v T) *T {
	return &v
}

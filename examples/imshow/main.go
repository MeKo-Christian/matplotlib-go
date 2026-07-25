package main

import (
	"fmt"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
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
		Colormap:      optional.Of(cmap),
		Aspect:        optional.Of("equal"),
		Extent:        optional.Of([4]float64{-2, 2, -1, 1}),
		Interpolation: optional.Of("bilinear"),
	})

	if err := fig.Save("imshow_extent.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}
	fmt.Println("saved imshow_extent.png")
}

func ptr[T any](v T) *T {
	return &v
}

package main

import (
	"flag"
	"fmt"
	"math"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

const (
	width  = 760
	height = 460
)

func main() {
	output := flag.String("out", "mesh_gouraud.png", "output PNG file")
	flag.Parse()

	fig := core.NewFigure(width, height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.14},
		Max: geom.Pt{X: 0.88, Y: 0.90},
	})
	ax.SetTitle("Gouraud PColorMesh")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	x := linspace(-3.0, 3.0, 9)
	y := linspace(-2.2, 2.2, 7)
	z := smoothField(x, y)
	cmap := "viridis"
	vmin, vmax := -0.85, 0.85
	mesh := ax.PColorMesh(z, core.MeshOptions{
		XEdges:   x,
		YEdges:   y,
		Shading:  core.MeshShadingGouraud,
		Colormap: &cmap,
		VMin:     &vmin,
		VMax:     &vmax,
		Label:    "gouraud mesh",
	})
	if mesh == nil {
		fmt.Println("error creating Gouraud mesh")
		return
	}
	ax.SetXLim(x[0], x[len(x)-1])
	ax.SetYLim(y[0], y[len(y)-1])
	ax.AddXGrid()
	ax.AddYGrid()
	fig.AddColorbar(ax, mesh, core.ColorbarOptions{Label: "value"})

	if err := fig.Save(*output); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}
	fmt.Printf("saved %s\n", *output)
}

func linspace(start, stop float64, n int) []float64 {
	out := make([]float64, n)
	if n == 1 {
		out[0] = start
		return out
	}
	step := (stop - start) / float64(n-1)
	for i := range out {
		out[i] = start + float64(i)*step
	}
	return out
}

func smoothField(x, y []float64) [][]float64 {
	values := make([][]float64, len(y))
	for yi, yy := range y {
		values[yi] = make([]float64, len(x))
		for xi, xx := range x {
			r := math.Hypot(xx*0.82, yy*1.10)
			values[yi][xi] = math.Sin(1.8*r)*math.Exp(-0.12*r*r) + 0.18*math.Cos(1.6*xx-0.7*yy)
		}
	}
	return values
}

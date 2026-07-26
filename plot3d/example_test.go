package plot3d_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/plot3d"
)

func ExampleAddAxes() {
	fig := core.NewFigure(640, 480)
	ax, err := plot3d.AddAxes(fig, geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	if err != nil {
		panic(err)
	}
	ax.Plot3D(
		[]float64{0, 1, 2},
		[]float64{0, 1, 0},
		[]float64{0, 0.5, 1},
		core.PlotOptions{},
	)

	fmt.Println(ax.ProjectionName())
	// Output:
	// 3d
}

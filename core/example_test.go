package core_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Example builds a figure with one Axes and a line artist, then renders it
// through the no-op NullRenderer to exercise the draw traversal.
func Example() {
	fig := core.NewFigure(640, 480)

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.95, Y: 0.95},
	})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 1)

	ax.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 5, Y: 0.8},
			{X: 10, Y: 0.3},
		},
		W:   2.0,
		Col: render.Color{R: 0, G: 0, B: 0, A: 1},
	})

	core.DrawFigure(fig, &render.NullRenderer{})

	fmt.Printf("figure %dx%d with %d axes\n",
		int(fig.SizePx.X), int(fig.SizePx.Y), len(fig.Children))

	// Output:
	// figure 640x480 with 1 axes
}

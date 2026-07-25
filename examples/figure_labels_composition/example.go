package figure_labels_composition

import (
	"fmt"
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1100
	Height = 720
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(1100, 720)
	fig.ConstrainedLayout()
	grid := fig.Subplots(2, 2)
	fig.SetSuptitle("Shared-Axis Figure Labels")
	fig.SetSupXLabel("time [s]")
	fig.SetSupYLabel("amplitude")

	textBox := &core.TextBBoxOptions{
		FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor: render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
	}

	for row := range grid {
		for col, ax := range grid[row] {
			x := make([]float64, 180)
			y := make([]float64, 180)
			for i := range x {
				xv := 2 * math.Pi * float64(i) / float64(len(x)-1)
				x[i] = xv
				y[i] = math.Sin(xv+float64(row)*0.5) * (1 + float64(col)*0.2)
			}
			ax.Plot(x, y, core.PlotOptions{
				LineWidth: common.FloatPtr(1.5),
				Label:     fmt.Sprintf("series %d", row*2+col+1),
			})
			ax.SetTitle(fmt.Sprintf("Panel %d", row*2+col+1))
			ax.SetXLabel("local x")
			ax.SetYLabel("local y")
			ax.SetXLim(0, 2*math.Pi)
			ax.SetYLim(-1.6, 1.6)
			common.AddReferenceXYGrid(ax)
		}
	}

	grid[0][0].Text(0.02, 0.92, "upper-left\nnote", core.TextOptions{
		Coords: core.Coords(core.CoordAxes),
		VAlign: core.TextVAlignTop,
		BBox:   textBox,
	})
	grid[1][1].Text(0.98, 0.08, "lower-right", core.TextOptions{
		Coords: core.Coords(core.CoordAxes),
		HAlign: core.TextAlignRight,
		VAlign: core.TextVAlignBottom,
		BBox:   textBox,
	})
	fig.Text(0.985, 0.94, "Figure note", core.TextOptions{
		HAlign: core.TextAlignRight,
		VAlign: core.TextVAlignTop,
		BBox:   textBox,
	})
	legend := fig.AddLegend()
	legend.SetLocator(core.BBoxToAnchorLocator{
		X:        0.99,
		Y:        0.90,
		Location: core.LegendUpperRight,
	})
	return fig
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

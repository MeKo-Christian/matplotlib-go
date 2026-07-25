package locator_maxn_edge_labels

import (
	"fmt"
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 720
	Height = 540
	DPI    = 100
)

// Plot builds a MaxNLocator edge-behavior parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	axes := []*core.Axes{
		fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.70}, Max: geom.Pt{X: 0.92, Y: 0.92}}),
		fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.40}, Max: geom.Pt{X: 0.92, Y: 0.62}}),
		fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.10}, Max: geom.Pt{X: 0.92, Y: 0.32}}),
	}
	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0

	axes[0].SetTitle("MaxNLocator Degenerate View")
	axes[0].SetXLabel("expanded around 2")
	axes[0].SetYLabel("value")
	axes[0].Plot([]float64{2, 2}, []float64{0.2, 0.9}, core.PlotOptions{Color: &color, LineWidth: &width})
	axes[0].SetXLim(2, 2)
	axes[0].SetYLim(0, 1)
	axes[0].XAxis.Locator = ticker.MaxNLocator{N: 2, Steps: []float64{1, 2, 2.5, 5, 10}}

	axes[1].SetTitle("MaxNLocator Prune Both")
	axes[1].SetXLabel("pruned range")
	axes[1].SetYLabel("value")
	axes[1].Plot([]float64{-3, -1, 1, 3, 5, 7}, []float64{0.16, 0.28, 0.44, 0.62, 0.78, 0.90}, core.PlotOptions{Color: &color, LineWidth: &width})
	axes[1].SetXLim(-3, 7)
	axes[1].SetYLim(0, 1)
	axes[1].XAxis.Locator = ticker.MaxNLocator{N: 5, Prune: "both"}

	axes[2].SetTitle("MaxNLocator Large Offset")
	axes[2].SetXLabel("1e6 + offset")
	axes[2].SetYLabel("value")
	axes[2].Plot(
		[]float64{1_000_000, 1_000_001, 1_000_002, 1_000_003, 1_000_004},
		[]float64{0.14, 0.30, 0.50, 0.72, 0.88},
		core.PlotOptions{Color: &color, LineWidth: &width},
	)
	axes[2].SetXLim(1_000_000, 1_000_004)
	axes[2].SetYLim(0, 1)
	axes[2].XAxis.Locator = ticker.MaxNLocator{N: 4}
	axes[2].XAxis.Formatter = ticker.FuncFormatter(func(v float64) string {
		return fmt.Sprintf("+%.0f", v-1_000_000)
	})

	for _, ax := range axes {
		common.AddReferenceYGrid(ax)
		ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 0.5, 1.0}}
	}
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

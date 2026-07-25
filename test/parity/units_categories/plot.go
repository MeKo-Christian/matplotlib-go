package units_categories

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 760
	Height = 360
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(760, 360)
	left := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.20}, Max: geom.Pt{X: 0.47, Y: 0.86}})
	left.SetTitle("Categorical X")
	left.SetYLabel("Count")
	common.AddReferenceYGrid(left)
	orange := render.Color{R: 1.0, G: 0.50, B: 0.05, A: 1}
	edgeOrange := render.Color{R: 0.60, G: 0.30, B: 0.03, A: 1}
	barEdgeWidth := 1.0
	barWidth := 0.8
	_, err := left.Bar([]string{"draft", "review", "ship", "watch"}, []float64{3, 8, 6, 4}, core.BarOptions{
		Color:     &orange,
		EdgeColor: &edgeOrange,
		EdgeWidth: &barEdgeWidth,
		Width:     &barWidth,
	})
	if err != nil {
		panic(err)
	}
	left.AutoScale(0.10)

	right := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.58, Y: 0.20}, Max: geom.Pt{X: 0.94, Y: 0.86}})
	right.SetTitle("Categorical Y")
	right.SetXLabel("Hours")
	right.SetAxisBelow(true)
	xGrid := right.AddGrid(core.AxisBottom)
	xGrid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	xGrid.LineWidth = 0.5
	green := render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1}
	edgeGreen := render.Color{R: 0.09, G: 0.36, B: 0.09, A: 1}
	orientation := core.BarHorizontal
	_, err = right.BarH([]string{"north", "south", "east"}, []float64{4, 7, 5}, core.BarOptions{
		Color:       &green,
		EdgeColor:   &edgeGreen,
		EdgeWidth:   &barEdgeWidth,
		Width:       &barWidth,
		Orientation: &orientation,
	})
	if err != nil {
		panic(err)
	}
	right.AutoScale(0.10)
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, Width, Height)
}

package plot_variants

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 840
	Height = 620
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(840, 620)

	stepAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.585}, Max: geom.Pt{X: 0.475, Y: 0.93}})
	stepAx.SetTitle("Step + Stairs")
	stepAx.SetXLim(0, 6)
	stepAx.SetYLim(0, 5.2)
	stepAx.SetAxisBelow(true)
	stepAx.AddYGrid()
	stepWhere := core.StepWherePost
	stepAx.Step(
		[]float64{0.6, 1.4, 2.2, 3.0, 3.8, 4.6, 5.4},
		[]float64{1.1, 2.5, 1.7, 3.4, 2.9, 4.1, 3.6},
		core.StepOptions{
			Where:     &stepWhere,
			Color:     &render.Color{R: 0.15, G: 0.39, B: 0.78, A: 1},
			LineWidth: common.FloatPtr(2.0),
		},
	)
	fillTrue := true
	stairsBaseline := 0.35
	stepAx.Stairs(
		[]float64{0.9, 1.7, 1.4, 2.6, 1.8, 2.2},
		[]float64{0.4, 1.1, 2.0, 2.9, 3.7, 4.6, 5.5},
		core.StairsOptions{
			Fill:      &fillTrue,
			Baseline:  &stairsBaseline,
			Color:     &render.Color{R: 0.91, G: 0.49, B: 0.20, A: 0.72},
			EdgeColor: &render.Color{R: 0.58, G: 0.26, B: 0.08, A: 1},
			LineWidth: common.FloatPtr(1.5),
		},
	)

	fillAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.575, Y: 0.585}, Max: geom.Pt{X: 0.97, Y: 0.93}})
	fillAx.SetTitle("FillBetweenX + Refs")
	fillAx.SetXLim(0, 7)
	fillAx.SetYLim(0, 6)
	fillAx.SetAxisBelow(true)
	fillAx.AddXGrid()
	_, _ = fillAx.FillBetweenX(
		[]float64{0.4, 1.2, 2.0, 2.8, 3.6, 4.4, 5.2},
		[]float64{1.3, 2.1, 1.7, 2.8, 2.2, 3.1, 2.6},
		[]float64{3.4, 4.1, 4.8, 5.1, 5.6, 6.0, 6.3},
		core.FillOptions{
			Color:     &render.Color{R: 0.24, G: 0.68, B: 0.54, A: 0.72},
			EdgeColor: &render.Color{R: 0.12, G: 0.38, B: 0.28, A: 1},
			EdgeWidth: common.FloatPtr(1.2),
		},
	)
	fillAx.AxVSpan(2.2, 3.1, core.VSpanOptions{
		Color: &render.Color{R: 0.92, G: 0.75, B: 0.18, A: 1},
		Alpha: common.FloatPtr(0.20),
	})
	fillAx.AxHLine(4.0, core.HLineOptions{
		Color:     &render.Color{R: 0.52, G: 0.18, B: 0.18, A: 1},
		LineWidth: common.FloatPtr(1.2),
		Dashes:    []float64{4 * 36.0 / DPI, 3 * 36.0 / DPI},
	})
	fillAx.AxVLine(5.3, core.VLineOptions{
		Color:     &render.Color{R: 0.18, G: 0.22, B: 0.55, A: 1},
		LineWidth: common.FloatPtr(1.2),
		Dashes:    []float64{2 * 36.0 / DPI, 2 * 36.0 / DPI},
	})
	fillAx.AxLine(
		geom.Pt{X: 0.9, Y: 0.3},
		geom.Pt{X: 6.4, Y: 5.6},
		core.ReferenceLineOptions{
			Color:     &render.Color{R: 0.22, G: 0.22, B: 0.22, A: 1},
			LineWidth: common.FloatPtr(1.1),
		},
	)

	brokenAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.10}, Max: geom.Pt{X: 0.475, Y: 0.445}})
	brokenAx.SetTitle("broken_barh")
	brokenAx.SetXLim(0, 10)
	brokenAx.SetYLim(0, 4.4)
	brokenAx.SetAxisBelow(true)
	brokenAx.AddXGrid()
	brokenAx.BrokenBarH(
		[][2]float64{{0.8, 1.6}, {3.1, 2.2}, {6.5, 1.3}},
		[2]float64{0.7, 0.9},
		core.BarOptions{Color: &render.Color{R: 0.21, G: 0.51, B: 0.76, A: 1}},
	)
	brokenAx.BrokenBarH(
		[][2]float64{{1.6, 1.0}, {4.0, 1.4}, {7.1, 1.7}},
		[2]float64{2.1, 0.9},
		core.BarOptions{Color: &render.Color{R: 0.86, G: 0.38, B: 0.16, A: 1}},
	)
	for _, label := range []struct {
		x, y float64
		text string
	}{
		{1.6, 1.15, "prep"},
		{4.2, 1.15, "run"},
		{7.15, 1.15, "cool"},
		{2.1, 2.55, "IO"},
		{4.7, 2.55, "fit"},
		{7.95, 2.55, "ship"},
	} {
		brokenAx.Text(label.x, label.y, label.text, core.TextOptions{
			HAlign:   core.TextAlignCenter,
			VAlign:   core.TextVAlignMiddle,
			FontSize: 10,
			Color:    render.Color{R: 1, G: 1, B: 1, A: 1},
		})
	}

	stackAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.575, Y: 0.10}, Max: geom.Pt{X: 0.97, Y: 0.445}})
	stackAx.SetTitle("Stacked Bars + Labels")
	stackAx.SetXLim(0.4, 4.6)
	stackAx.SetYLim(0, 7.6)
	stackAx.SetAxisBelow(true)
	stackAx.AddYGrid()
	x := []float64{1, 2, 3, 4}
	base := []float64{0, 0, 0, 0}
	seriesA := []float64{1.4, 2.2, 1.8, 2.5}
	seriesB := []float64{2.1, 1.6, 2.4, 1.7}
	bottom, _ := stackAx.Bar(x, seriesA, core.BarOptions{
		Baselines: base,
		Color:     &render.Color{R: 0.16, G: 0.59, B: 0.49, A: 1},
	})
	top, _ := stackAx.Bar(x, seriesB, core.BarOptions{
		Baselines: seriesA,
		Color:     &render.Color{R: 0.88, G: 0.47, B: 0.16, A: 1},
	})
	stackAx.BarLabel(bottom, []string{"A1", "A2", "A3", "A4"}, core.BarLabelOptions{
		Position: "center",
		Color:    render.Color{R: 1, G: 1, B: 1, A: 1},
		FontSize: 10,
	})
	stackAx.BarLabel(top, nil, core.BarLabelOptions{
		Format:   "%.1f",
		Color:    render.Color{R: 0.20, G: 0.20, B: 0.20, A: 1},
		FontSize: 10,
		Padding:  4,
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

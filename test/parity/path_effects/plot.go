package path_effects

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const (
	Width  = 640
	Height = 360
)

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height, style.WithFont("DejaVu Sans", 12))

	textAx := panel(fig, 0.06, 0.55, 0.47, 0.93)
	textAx.Text(0.5, 0.52, "Path FX", core.TextOptions{
		FontSize: 26,
		Color:    render.Color{R: 0.08, G: 0.18, B: 0.34, A: 1},
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignMiddle,
		PathEffects: []render.PathEffect{
			render.SimplePatchShadowPathEffect(geom.Pt{X: 4, Y: -5}, render.Color{R: 0.02, G: 0.03, B: 0.04, A: 0.75}, 0.55, 0.25),
			render.NormalPathEffect(),
		},
	})

	lineAx := panel(fig, 0.53, 0.55, 0.94, 0.93)
	line := lineAx.Plot(pathEffectLineX(), pathEffectLineY(), core.PlotOptions{
		Color:     colorPtr(render.Color{R: 0.08, G: 0.34, B: 0.66, A: 1}),
		LineWidth: floatPtr(3),
		LineCap:   lineCapPtr(render.CapButt),
	})
	line.PathEffects = render.WithStrokePathEffects(render.Color{R: 1, G: 1, B: 1, A: 0.96}, 10, geom.Pt{})

	scatterAx := panel(fig, 0.06, 0.08, 0.47, 0.47)
	size := 560.0
	scatter := scatterAx.Scatter(
		[]float64{0.22, 0.50, 0.78, 0.68},
		[]float64{0.68, 0.34, 0.70, 0.30},
		core.ScatterOptions{
			Size:      &size,
			Color:     colorPtr(render.Color{R: 0.89, G: 0.22, B: 0.24, A: 1}),
			EdgeColor: colorPtr(render.Color{R: 0.04, G: 0.06, B: 0.08, A: 1}),
			EdgeWidth: floatPtr(1.4),
		},
	)
	scatter.Colors = []render.Color{
		{R: 0.89, G: 0.22, B: 0.24, A: 1},
		{R: 0.10, G: 0.55, B: 0.38, A: 1},
		{R: 0.13, G: 0.35, B: 0.72, A: 1},
		{R: 0.94, G: 0.64, B: 0.18, A: 1},
	}
	scatter.PathEffects = []render.PathEffect{
		render.SimplePatchShadowPathEffect(geom.Pt{X: 4, Y: -5}, render.Color{R: 0.02, G: 0.03, B: 0.04, A: 0.70}, 0.5, 0.3),
		render.NormalPathEffect(),
	}

	polygonAx := panel(fig, 0.53, 0.08, 0.94, 0.47)
	polygonAx.AddPatch(&core.Polygon{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.94, G: 0.77, B: 0.28, A: 1},
			EdgeColor: render.Color{R: 0.07, G: 0.20, B: 0.38, A: 1},
			EdgeWidth: 2.2,
			LineJoin:  render.JoinRound,
			PathEffects: []render.PathEffect{
				render.SimplePatchShadowPathEffect(geom.Pt{X: 5, Y: -6}, render.Color{R: 0.02, G: 0.03, B: 0.04, A: 0.70}, 0.45, 0.35),
				render.PathPatchPathEffect(render.Color{R: 0.95, G: 0.92, B: 0.82, A: 0.75}, render.Color{R: 0.83, G: 0.20, B: 0.19, A: 1}, 5.5, geom.Pt{}),
				render.NormalPathEffect(),
			},
		},
		XY: []geom.Pt{
			{X: 0.12, Y: 0.20},
			{X: 0.32, Y: 0.82},
			{X: 0.68, Y: 0.76},
			{X: 0.90, Y: 0.34},
			{X: 0.60, Y: 0.12},
		},
	})

	return fig
}

func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

func panel(fig *core.Figure, x0, y0, x1, y1 float64) *core.Axes {
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: x0, Y: y0},
		Max: geom.Pt{X: x1, Y: y1},
	})
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.XAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.ShowFrame = false
	return ax
}

func pathEffectLineX() []float64 {
	x := make([]float64, 49)
	for i := range x {
		x[i] = 0.08 + 0.84*float64(i)/48
	}
	return x
}

func pathEffectLineY() []float64 {
	y := make([]float64, 49)
	for i := range y {
		t := float64(i) / 48
		y[i] = 0.50 + 0.24*math.Sin(t*math.Pi*2.35)
	}
	return y
}

func colorPtr(c render.Color) *render.Color { return &c }

func floatPtr(v float64) *float64 { return &v }

func lineCapPtr(v render.LineCap) *render.LineCap { return &v }

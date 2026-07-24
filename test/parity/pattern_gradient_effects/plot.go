package pattern_gradient_effects

import (
	"image"
	"math"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
)

func init() {
	matcolor.RegisterColormap("pattern-gradient-linear", matcolor.NewLinearSegmentedColormap("pattern-gradient-linear", []matcolor.ColorStop{
		{Pos: 0, Color: render.Color{R: 0.90, G: 0.16, B: 0.18, A: 1}},
		{Pos: 0.5, Color: render.Color{R: 0.96, G: 0.78, B: 0.20, A: 1}},
		{Pos: 1, Color: render.Color{R: 0.10, G: 0.30, B: 0.78, A: 1}},
	}))
	matcolor.RegisterColormap("pattern-gradient-radial", matcolor.NewLinearSegmentedColormap("pattern-gradient-radial", []matcolor.ColorStop{
		{Pos: 0, Color: render.Color{R: 0.98, G: 0.98, B: 0.86, A: 1}},
		{Pos: 1, Color: render.Color{R: 0.08, G: 0.50, B: 0.36, A: 1}},
	}))
}

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, Width)
	ax.SetYLim(Height, 0)
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.XAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.ShowFrame = false

	linearGradient(ax)
	radialGradient(ax)
	pattern(ax)
	patchEffects(ax)

	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

func linearGradient(ax *core.Axes) {
	const width = 164
	left := linspace(0, 0.5, width/2)
	right := linspace(0.5, 1, width-width/2)
	row := append(left, right...)
	img := repeatRow(row, 100)
	interp := "bilinear"
	cmap := "pattern-gradient-linear"
	x0, x1, y0, y1 := 42.0, 205.0, 142.0, 42.0
	ax.Image(img, core.ImageOptions{
		XMin:          &x0,
		XMax:          &x1,
		YMin:          &y0,
		YMax:          &y1,
		Colormap:      &cmap,
		Interpolation: &interp,
	})
	frame := &core.Rectangle{
		XY:     geom.Pt{X: 42, Y: 42},
		Width:  163,
		Height: 100,
		Patch: core.Patch{
			EdgeColor: render.Color{R: 0.06, G: 0.08, B: 0.10, A: 1},
			EdgeWidth: 1.8,
		},
	}
	frame.SetFaceColor(render.Color{})
	ax.AddPatch(frame)
}

func radialGradient(ax *core.Axes) {
	const (
		w = 164
		h = 100
	)
	img := make([][]float64, h)
	for y := range img {
		img[y] = make([]float64, w)
		for x := range img[y] {
			dist := math.Sqrt(math.Pow((float64(x)-w/2.0)/82, 2) + math.Pow((float64(y)-h/2.0)/82, 2))
			img[y][x] = clamp01(dist)
		}
	}
	interp := "bilinear"
	cmap := "pattern-gradient-radial"
	x0, x1, y0, y1 := 235.0, 398.0, 142.0, 42.0
	ax.Image(img, core.ImageOptions{
		XMin:          &x0,
		XMax:          &x1,
		YMin:          &y0,
		YMax:          &y1,
		Colormap:      &cmap,
		Interpolation: &interp,
	})
	frame := &core.Rectangle{
		XY:     geom.Pt{X: 235, Y: 42},
		Width:  163,
		Height: 100,
		Patch: core.Patch{
			EdgeColor: render.Color{R: 0.06, G: 0.08, B: 0.10, A: 1},
			EdgeWidth: 1.8,
		},
	}
	frame.SetFaceColor(render.Color{})
	ax.AddPatch(frame)
}

func pattern(ax *core.Axes) {
	ax.AddPatch(&core.Polygon{
		XY: []geom.Pt{
			{X: 455, Y: 44},
			{X: 594, Y: 54},
			{X: 570, Y: 145},
			{X: 430, Y: 128},
		},
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.93, G: 0.94, B: 0.98, A: 1},
			EdgeColor: render.Color{R: 0.06, G: 0.08, B: 0.10, A: 1},
			EdgeWidth: 1.8,
			Hatch:     "///",
		},
	})
}

func patchEffects(ax *core.Axes) {
	ax.AddPatch(&core.Rectangle{
		XY:     geom.Pt{X: 86, Y: 210},
		Width:  142,
		Height: 88,
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.12, G: 0.56, B: 0.40, A: 1},
			EdgeColor: render.Color{R: 0.02, G: 0.09, B: 0.16, A: 1},
			EdgeWidth: 2.2,
			LineJoin:  render.JoinRound,
			PathEffects: []render.PathEffect{
				render.StrokePathEffect(render.Color{R: 1, G: 0.92, B: 0.58, A: 0.95}, 9, geom.Pt{X: 4, Y: 4}),
				render.NormalPathEffect(),
			},
		},
	})

	ax.AddPatch(&core.Rectangle{
		XY:     geom.Pt{X: 362, Y: 210},
		Width:  144,
		Height: 88,
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.10, G: 0.28, B: 0.74, A: 1},
			PathEffects: []render.PathEffect{
				render.SimplePatchShadowPathEffect(geom.Pt{X: 8, Y: -8}, render.Color{R: 0.08, G: 0.12, B: 0.24, A: 0.45}, 0.45, 0.35),
				render.NormalPathEffect(),
			},
		},
	})
}

func linspace(start, stop float64, n int) []float64 {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return []float64{start}
	}
	out := make([]float64, n)
	for i := range out {
		out[i] = start + (stop-start)*float64(i)/float64(n-1)
	}
	return out
}

func repeatRow(row []float64, n int) [][]float64 {
	img := make([][]float64, n)
	for i := range img {
		img[i] = append([]float64(nil), row...)
	}
	return img
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

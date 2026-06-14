package fill_between

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Fill Between Curves")
	ax.SetXLim(0, 6.28)
	ax.SetYLim(-1.5, 1.5)

	n := 50
	x := make([]float64, n)
	y1 := make([]float64, n)
	y2 := make([]float64, n)
	for i := 0; i < n; i++ {
		t := 6.28 * float64(i) / float64(n-1)
		x[i] = t
		y1[i] = math.Sin(t)
		y2[i] = 0.8 * math.Cos(t)
	}

	fill := core.FillBetween(x, y1, y2, render.Color{R: 0.8, G: 0.3, B: 0.3, A: 0.6})
	fill.EdgeColor = render.Color{R: 0.5, G: 0.1, B: 0.1, A: 1.0}
	fill.EdgeWidth = 1.5
	ax.Add(fill)

	sineLine := &core.Line2D{XY: make([]geom.Pt, n), W: 2.0, Col: render.Color{R: 1, G: 0, B: 0, A: 1}}
	cosLine := &core.Line2D{XY: make([]geom.Pt, n), W: 2.0, Col: render.Color{R: 0, G: 0, B: 1, A: 1}}
	for i := 0; i < n; i++ {
		sineLine.XY[i] = geom.Pt{X: x[i], Y: y1[i]}
		cosLine.XY[i] = geom.Pt{X: x[i], Y: y2[i]}
	}
	ax.Add(sineLine)
	ax.Add(cosLine)
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
	return r.GetImage()
}

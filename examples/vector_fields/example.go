package vector_fields

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 919
	Height = 620
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(919, 620)
	axes := map[string]*core.Axes{
		"quiver": fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.58}, Max: geom.Pt{X: 0.47, Y: 0.92}}),
		"barbs":  fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.58}, Max: geom.Pt{X: 0.97, Y: 0.92}}),
		"stream": fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.10}, Max: geom.Pt{X: 0.47, Y: 0.44}}),
		"xy":     fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.10}, Max: geom.Pt{X: 0.97, Y: 0.44}}),
	}
	addVectorGrid := func(ax *core.Axes) {
		xGrid := ax.AddXGrid()
		yGrid := ax.AddYGrid()
		for _, grid := range []*core.Grid{xGrid, yGrid} {
			grid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
			grid.LineWidth = 0.5
		}
	}

	quiverAx := axes["quiver"]
	quiverAx.SetTitle("Quiver + Key")
	quiverAx.SetXLim(0, 6)
	quiverAx.SetYLim(0, 5)
	addVectorGrid(quiverAx)
	var qx, qy, qu, qv []float64
	for row := 0; row < 4; row++ {
		for col := 0; col < 5; col++ {
			x := 0.8 + float64(col)*1.0
			y := 0.8 + float64(row)*0.95
			qx = append(qx, x)
			qy = append(qy, y)
			qu = append(qu, 0.55+0.08*math.Sin(y*0.9))
			qv = append(qv, 0.22*math.Cos(x*0.8))
		}
	}
	scaleWidth := 10.0
	widthDots := 2.2
	quiver := quiverAx.Quiver(qx, qy, qu, qv, core.QuiverOptions{
		Color:      optional.Of(render.Color{R: 0.14, G: 0.42, B: 0.73, A: 1}),
		Scale:      optional.Of(scaleWidth),
		ScaleUnits: "width",
		Units:      "dots",
		Width:      optional.Of(widthDots),
	})
	if quiver != nil {
		quiverAx.QuiverKey(quiver, 0.78, 0.12, 0.5, "0.5", core.QuiverKeyOptions{
			Coords:   core.Coords(core.CoordAxes),
			LabelPos: "E",
		})
	}

	barbAx := axes["barbs"]
	barbAx.SetTitle("Barbs")
	barbAx.SetXLim(0, 6)
	barbAx.SetYLim(0, 5)
	addVectorGrid(barbAx)
	var bx, by, bu, bv []float64
	for row := 0; row < 4; row++ {
		for col := 0; col < 5; col++ {
			x := 0.9 + float64(col)*0.95
			y := 0.8 + float64(row)*0.95
			bx = append(bx, x)
			by = append(by, y)
			bu = append(bu, 14+5*math.Sin(y*0.8))
			bv = append(bv, 8*math.Cos(x*0.7))
		}
	}
	barbLen := 6.0
	barbLineWidth := 1.0
	barbAx.Barbs(bx, by, bu, bv, core.BarbsOptions{
		BarbColor: optional.Of(render.Color{R: 0.47, G: 0.23, B: 0.12, A: 1}),
		FlagColor: optional.Of(render.Color{R: 0.86, G: 0.52, B: 0.24, A: 1}),
		LineWidth: optional.Of(barbLineWidth),
		Length:    optional.Of(barbLen),
	})

	streamAx := axes["stream"]
	streamAx.SetTitle("Streamplot")
	streamAx.SetXLim(0, 6)
	streamAx.SetYLim(0, 5)
	addVectorGrid(streamAx)
	sx := []float64{0, 1, 2, 3, 4, 5, 6}
	sy := []float64{0, 1, 2, 3, 4, 5}
	su := make([][]float64, len(sy))
	sv := make([][]float64, len(sy))
	for yi, y := range sy {
		su[yi] = make([]float64, len(sx))
		sv[yi] = make([]float64, len(sx))
		for xi, x := range sx {
			su[yi][xi] = 1.0 + 0.12*math.Cos(y*0.7)
			sv[yi][xi] = 0.35*math.Sin((x-3)*0.8) - 0.10*(y-2.5)
		}
	}
	streamFalse := false
	streamLineWidth := 1.5
	streamArrowSize := 1.0
	streamAx.Streamplot(sx, sy, su, sv, core.StreamplotOptions{
		StartPoints:          []geom.Pt{{X: 0.4, Y: 0.8}, {X: 0.4, Y: 2.2}, {X: 0.4, Y: 3.6}},
		BrokenStreamlines:    optional.Of(streamFalse),
		IntegrationDirection: "forward",
		ArrowSize:            optional.Of(streamArrowSize),
		LineWidth:            optional.Of(streamLineWidth),
		Color:                optional.Of(render.Color{R: 0.13, G: 0.53, B: 0.39, A: 1}),
	})

	xyAx := axes["xy"]
	xyAx.SetTitle("Quiver XY")
	xyAx.SetXLim(0, 6)
	xyAx.SetYLim(0, 5)
	addVectorGrid(xyAx)
	xg := []float64{0.8, 1.8, 2.8, 3.8, 4.8}
	yg := []float64{0.8, 1.8, 2.8, 3.8}
	ugu := make([][]float64, len(yg))
	ugv := make([][]float64, len(yg))
	for yi, y := range yg {
		ugu[yi] = make([]float64, len(xg))
		ugv[yi] = make([]float64, len(xg))
		for xi, x := range xg {
			ugu[yi][xi] = -(y - 2.4) * 0.35
			ugv[yi][xi] = (x - 2.8) * 0.35
		}
	}
	xyScale := 9.0
	xyWidth := 1.9
	xyAx.QuiverGrid(xg, yg, ugu, ugv, core.QuiverOptions{
		Color:      optional.Of(render.Color{R: 0.74, G: 0.23, B: 0.27, A: 1}),
		Pivot:      "middle",
		Angles:     "xy",
		Scale:      optional.Of(xyScale),
		ScaleUnits: "width",
		Units:      "dots",
		Width:      optional.Of(xyWidth),
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

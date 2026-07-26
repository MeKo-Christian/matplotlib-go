package mesh_contour_tri

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 980
	Height = 620
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(980, 620)

	meshAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.57}, Max: geom.Pt{X: 0.46, Y: 0.93}})
	meshAx.SetTitle("PColorMesh")
	meshAx.SetXLim(0, 4)
	meshAx.SetYLim(0, 3)
	meshEdgeWidth := 0.8
	meshEdgeColor := render.Color{R: 0.95, G: 0.95, B: 0.95, A: 1}
	meshAx.PColorMesh([][]float64{
		{0.2, 0.6, 0.3, 0.9},
		{0.4, 0.8, 0.5, 0.7},
		{0.1, 0.3, 0.9, 0.6},
	}, core.MeshOptions{
		XEdges:    []float64{0, 1, 2, 3, 4},
		YEdges:    []float64{0, 1, 2, 3},
		EdgeColor: optional.Of(meshEdgeColor),
		EdgeWidth: optional.Of(meshEdgeWidth),
	})

	contourAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.57}, Max: geom.Pt{X: 0.96, Y: 0.93}})
	contourAx.SetTitle("Contour + Contourf")
	contourAx.SetXLim(0, 4)
	contourAx.SetYLim(0, 4)
	contourData := [][]float64{
		{0.0, 0.4, 0.8, 0.4, 0.0},
		{0.2, 0.8, 1.3, 0.8, 0.2},
		{0.3, 1.0, 1.7, 1.0, 0.3},
		{0.2, 0.8, 1.3, 0.8, 0.2},
		{0.0, 0.4, 0.8, 0.4, 0.0},
	}
	contourLevels := []float64{0.2, 0.6, 1.0, 1.4, 1.8}
	// Keep the catalog pair explicit: the Matplotlib reference script passes
	// linewidths=1.0 for both contour and tricontour. The library default
	// correctly follows contour.linewidth -> lines.linewidth (1.5).
	contourLineWidth := 1.0
	contourAx.Contourf(contourData, core.ContourOptions{
		Levels: contourLevels,
	})
	contourAx.Contour(contourData, core.ContourOptions{
		Levels:     []float64{0.4, 0.8, 1.2, 1.6},
		LabelLines: true,
		Color:      optional.Of(render.Color{R: 0.18, G: 0.18, B: 0.18, A: 1}),
		LineWidth:  optional.Of(contourLineWidth),
	})

	histAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.10}, Max: geom.Pt{X: 0.46, Y: 0.46}})
	histAx.SetTitle("Hist2D")
	histAx.SetXLim(0, 4)
	histAx.SetYLim(0, 4)
	histAx.Hist2D(
		[]float64{0.4, 0.7, 1.1, 1.4, 1.8, 2.1, 2.3, 2.6, 2.9, 3.2, 3.4, 3.6},
		[]float64{0.6, 1.0, 1.2, 1.6, 1.4, 2.0, 2.3, 2.1, 2.8, 3.0, 3.2, 3.4},
		core.Hist2DOptions{
			XBinEdges: []float64{0, 1, 2, 3, 4},
			YBinEdges: []float64{0, 1, 2, 3, 4},
			EdgeColor: optional.Of(meshEdgeColor),
			EdgeWidth: optional.Of(meshEdgeWidth),
		},
	)

	triAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.10}, Max: geom.Pt{X: 0.96, Y: 0.46}})
	triAx.SetTitle("Triangulation")
	triAx.SetXLim(0, 4)
	triAx.SetYLim(0, 4)
	tri := core.Triangulation{
		X:         []float64{0.4, 1.6, 3.0, 0.8, 2.1, 3.5},
		Y:         []float64{0.5, 0.4, 0.7, 2.2, 2.8, 2.1},
		Triangles: [][3]int{{0, 1, 3}, {1, 4, 3}, {1, 2, 4}, {2, 5, 4}},
	}
	triAx.TriColor(tri, []float64{0.2, 0.8, 1.0, 1.5, 1.1, 0.6}, core.TriColorOptions{})
	triLineWidth := 1.0
	triAx.TriPlot(tri, core.TriPlotOptions{
		Color:     optional.Of(render.Color{R: 0.15, G: 0.15, B: 0.15, A: 1}),
		LineWidth: optional.Of(triLineWidth),
	})
	triAx.TriContour(tri, []float64{0.2, 0.8, 1.0, 1.5, 1.1, 0.6}, core.ContourOptions{
		Levels:    []float64{0.7, 1.1},
		Color:     optional.Of(render.Color{R: 0.98, G: 0.98, B: 0.98, A: 1}),
		LineWidth: optional.Of(contourLineWidth),
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

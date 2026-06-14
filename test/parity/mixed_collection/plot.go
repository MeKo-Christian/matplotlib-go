package mixed_collection

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
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
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("RendererAgg mixed path collection")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 6)
	ax.XAxis.Locator = core.MultipleLocator{Base: 2}
	ax.YAxis.Locator = core.MultipleLocator{Base: 1}

	paths := []geom.Path{
		common.FixtureRectPath(0.8, 0.7, 1.4, 1.2),
		common.FixtureTrianglePath(2.9, 1.0, 0.9),
		common.FixtureDiamondPath(4.6, 1.2, 0.7),
		common.FixtureStarPath(6.4, 1.0, 0.75),
		common.FixtureRectPath(7.8, 0.7, 1.1, 1.7),
		common.FixtureTrianglePath(1.7, 4.0, 1.1),
		common.FixtureDiamondPath(3.8, 4.1, 0.8),
		common.FixtureStarPath(5.8, 4.0, 0.85),
		common.FixtureRectPath(7.5, 3.3, 1.5, 1.1),
	}
	faces := []render.Color{
		{R: 0.13, G: 0.47, B: 0.70, A: 0.65},
		{R: 1.00, G: 0.50, B: 0.05, A: 0.72},
		{R: 0.17, G: 0.63, B: 0.17, A: 0.70},
		{R: 0.84, G: 0.15, B: 0.16, A: 0.62},
		{R: 0.58, G: 0.40, B: 0.74, A: 0.70},
		{R: 0.55, G: 0.34, B: 0.29, A: 0.66},
		{R: 0.89, G: 0.47, B: 0.76, A: 0.66},
		{R: 0.50, G: 0.50, B: 0.50, A: 0.70},
		{R: 0.74, G: 0.74, B: 0.13, A: 0.70},
	}
	edges := []render.Color{
		{R: 0.02, G: 0.14, B: 0.23, A: 1},
		{R: 0.46, G: 0.21, B: 0.02, A: 1},
		{R: 0.02, G: 0.30, B: 0.06, A: 1},
		{R: 0.45, G: 0.04, B: 0.05, A: 1},
		{R: 0.28, G: 0.17, B: 0.42, A: 1},
		{R: 0.31, G: 0.17, B: 0.14, A: 1},
		{R: 0.44, G: 0.19, B: 0.37, A: 1},
		{R: 0.20, G: 0.20, B: 0.20, A: 1},
		{R: 0.36, G: 0.36, B: 0.04, A: 1},
	}
	widths := []float64{1.1, 1.6, 1.0, 1.8, 1.2, 1.4, 1.0, 1.6, 1.2}
	ax.AddCollection(&core.PatchCollection{
		Collection: core.Collection{Label: "mixed collection"},
		Paths:      paths,
		FaceColors: faces,
		EdgeColors: edges,
		EdgeWidths: widths,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	})
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, Width, Height)
}

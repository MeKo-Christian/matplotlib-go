package mplot3d_terrain

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/plot3d"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 900
	Height = 640
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(900, 640)
	ax, err := plot3d.AddAxes(fig, geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.08}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	if err != nil {
		panic(err)
	}
	ax.SetTitle("3D Surface + Filled Contours")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetView(35, -60)
	x, y, z := common.SinusoidalTerrain(90, 70)
	zeroWidth := 0.0
	viridis := "viridis"
	surfaceAlpha := 0.85
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	contourWidth := 0.6
	contourLevels := 8
	contourOffset := common.GridMin(z) - 0.2
	contourfAlpha := 0.45
	orange := render.Color{R: 1.0, G: 0.4980392156862745, B: 0.054901960784313725, A: 1}
	triAlpha := 0.7
	ax.PlotSurfaceGrid(x, y, z, core.PlotOptions{Colormap: optional.Of(viridis), LineWidth: optional.Of(zeroWidth), Alpha: optional.Of(surfaceAlpha)})
	ax.Plot3D([]float64{0, 0.9, 0.9, 0, 0}, []float64{0, 0, 0.9, 0.9, 0}, []float64{-0.2, -0.2, -0.2, -0.2, -0.2}, core.PlotOptions{Color: optional.Of(black)})
	ax.Scatter3D([]float64{0.2, 0.5, 0.8}, []float64{0.2, 0.5, 0.8}, []float64{0.3, 0.35, 0.2}, core.ScatterOptions{})
	ax.Contour(x, y, z, core.PlotOptions{Color: optional.Of(black), LineWidth: optional.Of(contourWidth), LevelCount: contourLevels})
	ax.Contourf(x, y, z, core.PlotOptions{Colormap: optional.Of(viridis), LevelCount: contourLevels, Offset: optional.Of(contourOffset), Alpha: optional.Of(contourfAlpha)})
	tri := core.Triangulation{X: []float64{0, 0.5, 1}, Y: []float64{0, 0, 0.4}, Triangles: [][3]int{{0, 1, 2}}}
	ax.Trisurf(tri, []float64{0.1, 0.4, 0.9}, core.PlotOptions{Color: optional.Of(orange), Alpha: optional.Of(triAlpha)})
	ax.Text(0.70, 0.62, "3D demo", core.TextOptions{Coords: core.Coords(core.CoordAxes)})
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return common.RenderFixtureFigure(fig, Width, Height)
}

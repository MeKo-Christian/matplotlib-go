// Package colorbar_variants_gallery shows common norm and extension colorbar
// combinations as a compact user-facing gallery.
package colorbar_variants_gallery

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1040
	Height = 720
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	registerExtensionColormap()
	fig := core.NewFigure(Width, Height)

	addImageNorm(fig, geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.60}, Max: geom.Pt{X: 0.34, Y: 0.88}}, "LogNorm", "magma", logData(6, 8), core.LogNorm{VMin: 1, VMax: 1000}, "log value")
	addImageNorm(fig, geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.60}, Max: geom.Pt{X: 0.84, Y: 0.88}}, "TwoSlopeNorm", "RdBu", twoSlopeData(6, 8), core.TwoSlopeNorm{VMin: -3, VCenter: 0, VMax: 6}, "anomaly")
	addBoundaryPanel(fig)
	addExtensionsPanel(fig)

	fig.Text(0.06, 0.95, "Colorbar Norms and Extensions", core.TextOptions{FontSize: 13})
	return fig
}

func addImageNorm(fig *core.Figure, rect geom.Rect, title, cmap string, data [][]float64, norm core.ScalarNormalizer, label string) {
	ax := fig.AddAxes(rect)
	ax.SetTitle(title)
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	extent := [4]float64{0, float64(len(data[0])), 0, float64(len(data))}
	img := ax.ImShow(data, core.ImShowOptions{
		Colormap: &cmap,
		Norm:     norm,
		Origin:   core.ImageOriginLower,
		Extent:   &extent,
		Aspect:   "auto",
	})
	if img != nil {
		fig.AddColorbar(ax, img, core.ColorbarOptions{Label: label, Padding: 0.03})
	}
}

func addBoundaryPanel(fig *core.Figure) {
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.16}, Max: geom.Pt{X: 0.34, Y: 0.44}})
	ax.SetTitle("BoundaryNorm")
	cmap := "viridis"
	mesh := ax.PColorMesh([][]float64{
		{0.2, 0.8, 1.2, 1.8},
		{2.2, 2.8, 3.2, 3.8},
		{0.5, 1.5, 2.5, 3.5},
	}, core.MeshOptions{
		XEdges:   []float64{0, 1, 2, 3, 4},
		YEdges:   []float64{0, 1, 2, 3},
		Colormap: &cmap,
		Norm:     core.BoundaryNorm{Boundaries: []float64{0, 1, 2, 3, 4}, NColors: 256},
	})
	if mesh != nil {
		fig.AddColorbar(ax, mesh, core.ColorbarOptions{
			Label:     "band",
			Padding:   0.03,
			DrawEdges: true,
			Ticks:     []float64{0, 1, 2, 3, 4},
		})
	}
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 3)
}

func addExtensionsPanel(fig *core.Figure) {
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.16}, Max: geom.Pt{X: 0.84, Y: 0.44}})
	ax.SetTitle("extensions")
	cmap := "phase18 extension"
	vmin, vmax := 0.0, 1.0
	mesh := ax.PColorMesh([][]float64{
		{-0.35, 0.15, 0.35},
		{0.55, 0.85, 1.35},
	}, core.MeshOptions{
		XEdges:   []float64{0, 1, 2, 3},
		YEdges:   []float64{0, 1, 2},
		Colormap: &cmap,
		VMin:     &vmin,
		VMax:     &vmax,
	})
	if mesh != nil {
		fig.AddColorbar(ax, mesh, core.ColorbarOptions{Label: "extended", Extend: "both", Padding: 0.03})
	}
	ax.SetXLim(0, 3)
	ax.SetYLim(0, 2)
}

func registerExtensionColormap() {
	under := render.Color{R: 0.08, G: 0.16, B: 0.72, A: 1}
	over := render.Color{R: 0.78, G: 0.12, B: 0.08, A: 1}
	matcolor.RegisterColormap("phase18 extension", matcolor.GetColormap("viridis").Copy("phase18 extension").WithUnder(under).WithOver(over))
}

func logData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for y := range data {
		data[y] = make([]float64, cols)
		for x := range data[y] {
			data[y][x] = math.Pow(10, 3*float64(x+y)/float64(rows+cols-2))
		}
	}
	return data
}

func twoSlopeData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for y := range data {
		data[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range data[y] {
			xx := float64(x) / float64(cols-1)
			data[y][x] = 6*xx - 3 + 1.5*(yy-0.5)
		}
	}
	return data
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

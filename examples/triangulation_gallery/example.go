// Package triangulation_gallery demonstrates unstructured triangulation and
// masked mesh variants in one catalog-backed showcase.
package triangulation_gallery

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1320
	Height = 760
	DPI    = 100
)

const maskedMeshCmap = "triangulation gallery mask"

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	registerMaskedMeshColormap()

	fig := core.NewFigure(Width, Height)
	tri, values := sampleTriangulation()

	meshAx := fig.AddAxes(panel(0, 0))
	configureTriAxes(meshAx, "Triplot")
	meshColor := render.Color{R: 0.18, G: 0.24, B: 0.34, A: 1}
	meshWidth := 1.25
	meshAx.TriPlot(tri, core.TriPlotOptions{
		Color:     optional.Of(meshColor),
		LineWidth: optional.Of(meshWidth),
		Label:     "triplot",
	})

	colorAx := fig.AddAxes(panel(0, 1))
	configureTriAxes(colorAx, "Tripcolor + Tricontour")
	cmap := "viridis"
	edgeColor := render.Color{R: 1, G: 1, B: 1, A: 1}
	edgeWidth := 0.55
	colorAx.TriColor(tri, values, core.TriColorOptions{
		Colormap:  optional.Of(cmap),
		EdgeColor: optional.Of(edgeColor),
		EdgeWidth: optional.Of(edgeWidth),
		Label:     "tripcolor",
	})
	contourColor := render.Color{R: 0.07, G: 0.10, B: 0.16, A: 0.95}
	contourWidth := 1.05
	colorAx.TriContour(tri, values, core.ContourOptions{
		Color:      optional.Of(contourColor),
		LineWidth:  optional.Of(contourWidth),
		LevelCount: 6,
		LabelLines: true,
		LabelColor: optional.Of(contourColor),
	})

	fillAx := fig.AddAxes(panel(1, 0))
	configureTriAxes(fillAx, "Tricontourf")
	fillMap := "plasma"
	fillAx.TriContourf(tri, values, core.ContourOptions{
		Colormap:   optional.Of(fillMap),
		LevelCount: 7,
		Label:      "tricontourf",
	})
	highlight := render.Color{R: 1, G: 1, B: 1, A: 0.88}
	highlightWidth := 0.9
	fillAx.TriContour(tri, values, core.ContourOptions{
		Color:      optional.Of(highlight),
		LineWidth:  optional.Of(highlightWidth),
		LevelCount: 7,
	})

	maskAx := fig.AddAxes(panel(1, 1))
	configureMeshAxes(maskAx, "Masked PColorMesh")
	maskAx.PColorMesh(maskedMeshData(), core.MeshOptions{
		XEdges:    []float64{0, 1, 2, 3, 4, 5},
		YEdges:    []float64{0, 1, 2, 3, 4},
		Mask:      maskedMeshMask(),
		Colormap:  optional.Of(maskedMeshCmap),
		EdgeColor: optional.Of(edgeColor),
		EdgeWidth: optional.Of(edgeWidth),
		Label:     "masked mesh",
	})

	fig.Text(0.98, 0.975, "triangulation gallery\ntriplot, tripcolor, tricontour, tricontourf, masked mesh", core.TextOptions{
		HAlign:   core.TextAlignRight,
		VAlign:   core.TextVAlignTop,
		FontSize: 11,
		BBox: optional.Of(core.TextBBoxOptions{
			FaceColor:    render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor:    render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1},
			Padding:      0.35 * 11 * DPI / 72,
			CornerRadius: 0.35 * 11 * DPI / 72,
		}),
	})

	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	return renderFixtureFigure(fig, Width, Height)
}

func panel(row, col int) geom.Rect {
	const (
		left   = 0.065
		right  = 0.955
		bottom = 0.11
		top    = 0.91
		hgap   = 0.085
		vgap   = 0.12
	)
	w := (right - left - hgap) / 2
	h := (top - bottom - vgap) / 2
	x0 := left + float64(col)*(w+hgap)
	y1 := top - float64(row)*(h+vgap)
	return geom.Rect{
		Min: geom.Pt{X: x0, Y: y1 - h},
		Max: geom.Pt{X: x0 + w, Y: y1},
	}
}

func configureTriAxes(ax *core.Axes, title string) {
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(-0.1, 3.1)
	ax.SetYLim(-0.15, 2.65)
	_ = ax.SetAspect("equal")
}

func configureMeshAxes(ax *core.Axes, title string) {
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(0, 5)
	ax.SetYLim(0, 4)
	_ = ax.SetAspect("equal")
}

func sampleTriangulation() (core.Triangulation, []float64) {
	tri := core.Triangulation{
		X: []float64{0.0, 0.85, 1.75, 2.85, 0.2, 1.1, 2.1, 0.55, 1.55, 2.55},
		Y: []float64{0.0, 0.2, 0.05, 0.3, 1.0, 1.15, 1.25, 2.15, 2.3, 2.05},
		Triangles: [][3]int{
			{0, 1, 4},
			{1, 5, 4},
			{1, 2, 5},
			{2, 6, 5},
			{2, 3, 6},
			{4, 5, 7},
			{5, 8, 7},
			{5, 6, 8},
			{6, 9, 8},
		},
	}

	values := make([]float64, len(tri.X))
	for i := range values {
		values[i] = math.Sin(tri.X[i]*1.4) + 0.7*math.Cos((tri.Y[i]+0.15)*2.1)
	}
	return tri, values
}

func maskedMeshData() [][]float64 {
	return [][]float64{
		{0.15, 0.35, 0.62, 0.48, 0.82},
		{0.30, 0.58, 0.76, 0.52, 0.68},
		{0.46, 0.72, 0.55, 0.28, 0.41},
		{0.22, 0.50, 0.88, 0.65, 0.34},
	}
}

func maskedMeshMask() [][]bool {
	return [][]bool{
		{false, true, false, false, false},
		{false, false, false, true, false},
		{true, false, false, false, false},
		{false, false, true, false, false},
	}
}

func registerMaskedMeshColormap() {
	bad := render.Color{R: 0.62, G: 0.62, B: 0.62, A: 0.78}
	matcolor.RegisterColormap(maskedMeshCmap, matcolor.LookupColormap("viridis").WithBad(bad))
}

func ptr[T any](v T) *T {
	return &v
}

func renderFixtureFigure(fig *core.Figure, width, height int) image.Image {
	r, err := agg.New(width, height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

// Package unstructured_showcase is a showcase example. The body lived previously in
// examples/unstructured/showcase; this is its self-contained home.
package unstructured_showcase

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
	Width  = 1320
	Height = 520
	DPI    = 100
)

// UnstructuredShowcase builds the same plot as
// test/matplotlib_ref/plots/unstructured_showcase.py.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	tri, values := sampleTriangulation()

	// Mirror the Python example's bbox=dict(boxstyle="round,pad=...",
	// facecolor="white", edgecolor=(0.75,0.75,0.75,1.0)) styling. Padding
	// and CornerRadius are in pixels here; matplotlib's pad is in fontsize
	// units (points), so multiply by DPI/72 to convert. matplotlib's
	// "round" boxstyle defaults rounding_size to the pad value.
	bboxBorder := render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1}
	bboxFace := render.Color{R: 1, G: 1, B: 1, A: 1}
	axTextPadPx := 0.3 * 10.0 * float64(DPI) / 72.0
	figTextPadPx := 0.35 * 11.0 * float64(DPI) / 72.0

	meshAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.16},
		Max: geom.Pt{X: 0.31, Y: 0.88},
	})
	configureAxes(meshAx, "Triangulation")
	meshColor := render.Color{R: 0.18, G: 0.24, B: 0.34, A: 1}
	meshWidth := 1.35
	meshAx.TriPlot(tri, core.TriPlotOptions{
		Color:     optional.Of(meshColor),
		LineWidth: optional.Of(meshWidth),
		Label:     "triplot",
	})
	meshAx.Text(0.98, 0.02, "explicit triangular mesh", core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		HAlign:   core.TextAlignRight,
		VAlign:   core.TextVAlignBottom,
		FontSize: 10,
		BBox: optional.Of(core.TextBBoxOptions{
			FaceColor:    bboxFace,
			EdgeColor:    bboxBorder,
			Padding:      axTextPadPx,
			CornerRadius: axTextPadPx,
		}),
	})

	colorAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.37, Y: 0.16},
		Max: geom.Pt{X: 0.63, Y: 0.88},
	})
	configureAxes(colorAx, "Tripcolor + Tricontour")
	cmap := "viridis"
	edgeColor := render.Color{R: 1, G: 1, B: 1, A: 1}
	edgeWidth := 0.6
	colorAx.TriColor(tri, values, core.TriColorOptions{
		Colormap:  optional.Of(cmap),
		EdgeColor: optional.Of(edgeColor),
		EdgeWidth: optional.Of(edgeWidth),
		Label:     "tripcolor",
	})
	contourColor := render.Color{R: 0.08, G: 0.12, B: 0.18, A: 0.95}
	contourWidth := 1.15
	contourSet := colorAx.TriContour(tri, values, core.ContourOptions{
		Color:      optional.Of(contourColor),
		LineWidth:  optional.Of(contourWidth),
		LevelCount: 6,
	})
	colorAx.Clabel(contourSet, core.ClabelOptions{
		Inline:       optional.Of(true),
		FormatString: "%.3g",
		FontSize:     optional.Of(10.0),
		Colors:       []render.Color{contourColor},
	})

	fillAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.69, Y: 0.16},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})
	configureAxes(fillAx, "Filled Tricontour")
	fillMap := "plasma"
	fillAx.TriContourf(tri, values, core.ContourOptions{
		Colormap:   optional.Of(fillMap),
		LevelCount: 7,
		Label:      "tricontourf",
	})
	highlight := render.Color{R: 1, G: 1, B: 1, A: 0.88}
	highlightWidth := 0.95
	fillAx.TriContour(tri, values, core.ContourOptions{
		Color:      optional.Of(highlight),
		LineWidth:  optional.Of(highlightWidth),
		LevelCount: 7,
	})

	fig.Text(0.98, 0.98, "unstructured gallery family\ntriangulation, tripcolor, tricontour", core.TextOptions{
		HAlign:   core.TextAlignRight,
		VAlign:   core.TextVAlignTop,
		FontSize: 11,
		BBox: optional.Of(core.TextBBoxOptions{
			FaceColor:    bboxFace,
			EdgeColor:    bboxBorder,
			Padding:      figTextPadPx,
			CornerRadius: figTextPadPx,
		}),
	})

	return fig
}

func configureAxes(ax *core.Axes, title string) {
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(-0.1, 3.1)
	ax.SetYLim(-0.15, 2.65)
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

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(DPI)
	core.DrawFigure(fig, r)
	return r.Image()
}

// Package arrays_showcase is a showcase example. The body lived previously in
// examples/arrays/showcase; this is its self-contained home.
package arrays_showcase

import (
	"fmt"
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1240
	Height = 620
	DPI    = 100
)

// ArraysShowcase builds the same plot as
// test/matplotlib_ref/plots/arrays_showcase.py.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	drawAnnotatedHeatmap(fig)
	drawMeshAndContour(fig)
	drawSpyMatrix(fig)
	fig.Text(0.98, 0.98, "arrays gallery family\nheatmap, quad mesh, sparse matrix", core.TextOptions{
		HAlign:   core.TextAlignRight,
		VAlign:   core.TextVAlignTop,
		FontSize: 11,
		BBox:     textBox(11, 0.35),
	})

	return fig
}

func axesRect(x0, y0, x1, y1 float64) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: x0, Y: y0},
		Max: geom.Pt{X: x1, Y: y1},
	}
}

func ptr[T any](v T) *T {
	return &v
}

func arange(n int) []float64 {
	values := make([]float64, n)
	for i := range values {
		values[i] = float64(i)
	}
	return values
}

func textBox(fontSize, pad float64) *core.TextBBoxOptions {
	padding := pad * fontSize * DPI / 72.0
	return &core.TextBBoxOptions{
		FaceColor:    render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor:    render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1},
		LineWidth:    1,
		Padding:      padding,
		CornerRadius: padding,
	}
}

func annotatedData() [][]float64 {
	return [][]float64{
		{0.12, 0.28, 0.46, 0.64, 0.82},
		{0.18, 0.34, 0.58, 0.74, 0.88},
		{0.24, 0.42, 0.63, 0.79, 0.91},
		{0.16, 0.38, 0.61, 0.83, 0.97},
	}
}

func waveGrid(rows, cols int, phase float64) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			values[y][x] = 0.55 + 0.25*math.Sin((xx*2.3+phase)*math.Pi) + 0.20*math.Cos((yy*2.8-phase*0.4)*math.Pi)
		}
	}
	return values
}

func sparsePattern(rows, cols int) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		for x := range cols {
			if x == y || x+y == cols-1 || (x+2*y)%7 == 0 || (2*x+y)%11 == 0 {
				values[y][x] = 1
			}
		}
	}
	return values
}

func useMatrixTopAxis(ax *core.Axes) {
	if ax.XAxis != nil {
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
	}
	top := ax.TopAxis()
	top.ShowSpine = true
	top.ShowTicks = true
	top.ShowLabels = true
	_ = ax.SetXLabelPosition("top")
}

func setMatrixTicks(ax *core.Axes, rows, cols int) {
	xTicks := arange(cols)
	yTicks := arange(rows)
	for _, axis := range []*core.Axis{ax.XAxis, ax.XAxisTop} {
		if axis != nil {
			axis.Locator = core.FixedLocator{TicksList: xTicks}
			axis.Formatter = core.ScalarFormatter{Prec: 0}
		}
	}
	if ax.YAxis != nil {
		ax.YAxis.Locator = core.FixedLocator{TicksList: yTicks}
		ax.YAxis.Formatter = core.ScalarFormatter{Prec: 0}
	}
}

func drawAnnotatedHeatmap(fig *core.Figure) {
	data := annotatedData()
	rows, cols := len(data), len(data[0])

	ax := fig.AddAxes(axesRect(0.05, 0.14, 0.31, 0.88))
	ax.SetTitle("Annotated Heatmap")
	ax.SetXLabel("column")
	ax.SetYLabel("row")

	ax.Image(data, core.ImageOptions{
		Colormap: ptr("viridis"),
		XMin:     ptr(-0.5),
		XMax:     ptr(float64(cols) - 0.5),
		YMin:     ptr(-0.5),
		YMax:     ptr(float64(rows) - 0.5),
		Origin:   core.ImageOriginUpper,
	})
	ax.SetXLim(-0.5, float64(cols)-0.5)
	ax.SetYLim(float64(rows)-0.5, -0.5)
	_ = ax.SetAspect("equal")
	setMatrixTicks(ax, rows, cols)
	useMatrixTopAxis(ax)

	minValue, maxValue := matrixRange(data)
	threshold := (minValue + maxValue) / 2.0
	for row := range rows {
		for col := range cols {
			textColor := render.Color{R: 0.12, G: 0.12, B: 0.14, A: 1}
			if data[row][col] >= threshold {
				textColor = render.Color{R: 1, G: 1, B: 1, A: 1}
			}
			ax.Text(float64(col), float64(row), fmt.Sprintf("%.2f", data[row][col]), core.TextOptions{
				HAlign:   core.TextAlignCenter,
				VAlign:   core.TextVAlignMiddle,
				FontSize: 10,
				Color:    textColor,
			})
		}
	}
}

func drawMeshAndContour(fig *core.Figure) {
	rows, cols := 8, 10
	data := waveGrid(rows, cols, 0.35)

	ax := fig.AddAxes(axesRect(0.37, 0.14, 0.63, 0.88))
	ax.SetTitle("PColorMesh + Contour")
	ax.SetXLabel("x bin")
	ax.SetYLabel("y bin")

	ax.PColorMesh(data, core.MeshOptions{
		XEdges:    arange(cols + 1),
		YEdges:    arange(rows + 1),
		Colormap:  ptr("plasma"),
		EdgeColor: ptr(render.Color{R: 1, G: 1, B: 1, A: 1}),
		EdgeWidth: ptr(0.65),
		Label:     "pcolormesh",
	})

	contourColor := ptr(render.Color{R: 0.14, G: 0.10, B: 0.16, A: 0.95})
	ax.Contour(data, core.ContourOptions{
		Color:          contourColor,
		LineWidth:      ptr(1.1),
		LevelCount:     6,
		X:              arange(cols),
		Y:              arange(rows),
		LabelLines:     true,
		LabelFormatter: core.FormatStrFormatter{Pattern: "%.3g"},
		LabelColor:     contourColor,
		LabelFontSize:  ptr(10.0),
	})
	ax.SetXLim(0, float64(cols))
	ax.SetYLim(0, float64(rows))
}

func drawSpyMatrix(fig *core.Figure) {
	ax := fig.AddAxes(axesRect(0.69, 0.14, 0.95, 0.88))
	ax.SetTitle("Spy")
	ax.SetXLabel("column")
	ax.SetYLabel("row")

	ax.Spy(sparsePattern(18, 18), core.SpyOptions{
		Precision:  0.1,
		Marker:     ptr(core.MarkerSquare),
		MarkerSize: 10,
		Color:      ptr(render.Color{R: 0.16, G: 0.38, B: 0.72, A: 1}),
		Label:      "spy",
	})
	ax.Text(0.98, 0.02, "sparse structure view", core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		HAlign:   core.TextAlignRight,
		VAlign:   core.TextVAlignBottom,
		FontSize: 10,
		BBox:     textBox(10, 0.3),
	})
}

func matrixRange(data [][]float64) (float64, float64) {
	minValue := math.Inf(1)
	maxValue := math.Inf(-1)
	for _, row := range data {
		for _, value := range row {
			minValue = math.Min(minValue, value)
			maxValue = math.Max(maxValue, value)
		}
	}
	return minValue, maxValue
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
	return r.GetImage()
}

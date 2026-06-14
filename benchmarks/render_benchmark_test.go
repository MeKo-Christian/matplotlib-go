package benchmarks

import (
	"image"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

var benchmarkImageSink image.Image
var benchmarkColorSink render.Color

func BenchmarkCatalogRender(b *testing.B) {
	cases := []string{
		"basic_line",
		"lines_markers_gallery",
		"scatter_gallery",
		"image_variants_gallery",
		"mesh_contour_tri",
		"triangulation_gallery",
		"mplot3d_gallery",
		"text_layout_gallery",
		"mathtext_gallery",
		"widgets_gallery",
	}

	for _, id := range cases {
		id := id
		b.Run(id, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				img, _, err := parity.Render(id)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkImageSink = img
			}
		})
	}
}

func BenchmarkLargeScatter100KDraw(b *testing.B) {
	fig := largeScatterFigure(100_000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkImageSink = drawAGG(fig)
	}
}

func BenchmarkLargeScatter100KRedrawReuseRenderer(b *testing.B) {
	fig := largeScatterFigure(100_000)
	r := newAGGRenderer(fig)
	bg := fig.RC.FigureBackground()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Clear(bg)
		core.DrawFigure(fig, r)
		benchmarkImageSink = r.ImageView()
	}
}

func BenchmarkLargeScatter100KBuildAndDraw(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchmarkImageSink = drawAGG(largeScatterFigure(100_000))
	}
}

func BenchmarkScalarMappedImageColors(b *testing.B) {
	data := scalarGrid(128, 128)
	mapping, err := core.ResolveScalarMapGrid(data, core.ScalarMapConfig{Colormap: "viridis_r"})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sink render.Color
		for _, row := range data {
			for _, value := range row {
				sink = mapping.Color(value, 1)
			}
		}
		benchmarkColorSink = sink
	}
}

func BenchmarkScalarMappedScatterColors(b *testing.B) {
	values := scalarValues(50_000)
	mapping, err := core.ResolveScalarMapValues(values, core.ScalarMapConfig{Colormap: "plasma_r"})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sink render.Color
		for _, value := range values {
			sink = mapping.Color(value, 0.72)
		}
		benchmarkColorSink = sink
	}
}

func BenchmarkScalarMappedQuadMeshColors(b *testing.B) {
	data := scalarGrid(96, 96)
	mapping, err := core.ResolveScalarMapGrid(data, core.ScalarMapConfig{Colormap: "magma_r"})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sink render.Color
		for y := 0; y+1 < len(data); y++ {
			for x := 0; x+1 < len(data[y]); x++ {
				value := (data[y][x] + data[y][x+1] + data[y+1][x] + data[y+1][x+1]) * 0.25
				sink = mapping.Color(value, 1)
			}
		}
		benchmarkColorSink = sink
	}
}

func benchmarkRepeatedDraw(b *testing.B, fig *core.Figure) {
	b.Helper()
	r := newAGGRenderer(fig)
	bg := fig.RC.FigureBackground()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Clear(bg)
		core.DrawFigure(fig, r)
		benchmarkImageSink = r.ImageView()
	}
}

func largeScatterFigure(n int) *core.Figure {
	fig := core.NewFigure(980, 620, style.WithDPI(100))
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.08, Y: 0.12},
		Max: geom.Pt{X: 0.96, Y: 0.90},
	})
	ax.SetTitle("100k point scatter")
	ax.SetXLim(-1, 101)
	ax.SetYLim(-1, 101)
	ax.AddYGrid()

	points := make([]geom.Pt, 0, n)
	sizes := make([]float64, 0, n)
	colors := make([]render.Color, 0, n)
	for i := 0; i < n; i++ {
		x := math.Mod(float64(i)*0.61803398875, 100)
		y := math.Mod(float64(i)*0.41421356237, 100)
		points = append(points, geom.Pt{
			X: x + 0.18*math.Sin(float64(i)*0.013),
			Y: y + 0.18*math.Cos(float64(i)*0.017),
		})
		sizes = append(sizes, core.ScatterAreaFromRadius(2.2+float64(i%5)*0.25, fig.RC.DPI))
		t := float64(i%256) / 255
		colors = append(colors, render.Color{
			R: 0.12 + 0.72*t,
			G: 0.42 + 0.20*math.Sin(float64(i)*0.01),
			B: 0.86 - 0.52*t,
			A: 0.72,
		})
	}

	ax.Add(&core.Scatter2D{
		XY:        points,
		Sizes:     sizes,
		Colors:    colors,
		EdgeColor: render.Color{A: 0},
		Marker:    core.MarkerCircle,
	})
	return fig
}

func scalarMappedImageFigure(rows, cols int) *core.Figure {
	fig := core.NewFigure(720, 520, style.WithDPI(100))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.12}, Max: geom.Pt{X: 0.92, Y: 0.90}})
	data := scalarGrid(rows, cols)
	cmap := "viridis_r"
	ax.Image(data, core.ImageOptions{Colormap: &cmap})
	return fig
}

func scalarMappedScatterFigure(n int) *core.Figure {
	fig := core.NewFigure(900, 560, style.WithDPI(100))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.12}, Max: geom.Pt{X: 0.94, Y: 0.90}})
	ax.SetXLim(-2, 102)
	ax.SetYLim(-2, 102)
	x := make([]float64, 0, n)
	y := make([]float64, 0, n)
	scalars := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		xv := math.Mod(float64(i)*0.61803398875, 100)
		yv := math.Mod(float64(i)*0.41421356237, 100)
		x = append(x, xv+0.1*math.Sin(float64(i)*0.011))
		y = append(y, yv+0.1*math.Cos(float64(i)*0.013))
		scalars = append(scalars, math.Sin(float64(i)*0.015)+math.Cos(float64(i)*0.009))
	}
	size := core.ScatterAreaFromRadius(1.8, fig.RC.DPI)
	cmap := "plasma_r"
	ax.Scatter(x, y, core.ScatterOptions{
		ScalarValues: scalars,
		Colormap:     cmap,
		Size:         &size,
		EdgeColor:    &render.Color{A: 0},
	})
	return fig
}

func scalarMappedQuadMeshFigure(rows, cols int) *core.Figure {
	fig := core.NewFigure(720, 520, style.WithDPI(100))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.12}, Max: geom.Pt{X: 0.92, Y: 0.90}})
	data := scalarGrid(rows, cols)
	cmap := "magma_r"
	ax.PColorMesh(data, core.MeshOptions{Colormap: &cmap})
	return fig
}

func scalarGrid(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for y := range data {
		row := make([]float64, cols)
		for x := range row {
			row[x] = math.Sin(float64(x)*0.05) + math.Cos(float64(y)*0.04) + 0.2*math.Sin(float64(x+y)*0.02)
		}
		data[y] = row
	}
	return data
}

func scalarValues(n int) []float64 {
	values := make([]float64, n)
	for i := range values {
		values[i] = math.Sin(float64(i)*0.015) + math.Cos(float64(i)*0.009)
	}
	return values
}

func drawAGG(fig *core.Figure) image.Image {
	r := newAGGRenderer(fig)
	core.DrawFigure(fig, r)
	return r.ImageView()
}

func newAGGRenderer(fig *core.Figure) *agg.Renderer {
	r, err := agg.New(int(fig.SizePx.X), int(fig.SizePx.Y), fig.RC.FigureBackground())
	if err != nil {
		panic(err)
	}
	r.SetResolution(uint(math.Round(fig.RC.DPI)))
	return r
}

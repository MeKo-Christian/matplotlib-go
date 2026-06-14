package benchmarks

import (
	"image"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

var benchmarkImageSink image.Image

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

func BenchmarkLargeScatter100KBuildAndDraw(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchmarkImageSink = drawAGG(largeScatterFigure(100_000))
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

func drawAGG(fig *core.Figure) image.Image {
	r, err := agg.New(int(fig.SizePx.X), int(fig.SizePx.Y), render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(uint(math.Round(fig.RC.DPI)))
	core.DrawFigure(fig, r)
	return r.GetImage()
}

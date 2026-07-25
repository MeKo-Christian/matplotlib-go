package main

import (
	"flag"
	"log"
	"math"
	"os"
	"path/filepath"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
)

type themedExample struct {
	name  string
	title string
	file  string
}

func main() {
	outputDir := flag.String("output-dir", "examples/styling", "directory for rendered theme examples")
	flag.Parse()
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		log.Fatalf("create output directory: %v", err)
	}

	examples := []themedExample{
		{name: "default", title: "Default Theme", file: "default_theme.png"},
		{name: "ggplot", title: "GGPlot Theme", file: "ggplot_theme.png"},
		{name: "publication", title: "Publication Theme", file: "publication_theme.png"},
	}

	for _, example := range examples {
		example.file = filepath.Join(*outputDir, example.file)
		if err := renderThemeExample(example); err != nil {
			log.Fatalf("render %s: %v", example.name, err)
		}
		log.Printf("saved %s", example.file)
	}
}

func renderThemeExample(example themedExample) error {
	// Each image uses the same plot body so only theme defaults differ.
	theme := style.MustTheme(example.name)
	fig := core.NewFigure(900, 520, style.WithTheme(theme))
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.14},
		Max: geom.Pt{X: 0.94, Y: 0.88},
	})

	x := linspace(0, 10, 220)
	sinWave := make([]float64, len(x))
	cosWave := make([]float64, len(x))
	envelope := make([]float64, len(x))
	for i, xv := range x {
		sinWave[i] = math.Sin(xv)
		cosWave[i] = 0.7 * math.Cos(xv*0.7)
		envelope[i] = 0.18 + 0.12*math.Sin(xv*0.4)
	}

	ax.SetTitle(example.title)
	ax.SetXLabel("Time")
	ax.SetYLabel("Signal")
	ax.SetXLim(0, 10)
	ax.SetYLim(-1.4, 1.4)

	xGrid := ax.AddXGrid()
	xGrid.Minor = true // Mirrors Matplotlib's grid(True, which="both").
	yGrid := ax.AddYGrid()

	_, _ = ax.FillBetweenPlot(x, shift(sinWave, envelope), shift(sinWave, negate(envelope)), core.FillOptions{
		Label: "Band",
	})
	_, _ = ax.Plot(x, sinWave, core.PlotOptions{Label: "sin(x)"})
	_, _ = ax.Plot(x, cosWave, core.PlotOptions{Label: "0.7 cos(0.7x)"})
	ax.Scatter([]float64{1.5, 4.8, 8.2}, []float64{0.99, -0.73, 0.94}, core.ScatterOptions{
		Label: "Samples",
	})

	ax.Annotate("peak", 1.5, 0.99, core.AnnotationOptions{})
	legend := ax.AddLegend()
	legend.Location = core.LegendUpperLeft
	_ = yGrid

	return fig.Save(example.file)
}

func linspace(start, end float64, n int) []float64 {
	if n <= 1 {
		return []float64{start}
	}
	values := make([]float64, n)
	step := (end - start) / float64(n-1)
	for i := range values {
		values[i] = start + float64(i)*step
	}
	return values
}

func shift(values, delta []float64) []float64 {
	out := make([]float64, len(values))
	for i, value := range values {
		offset := 0.0
		if i < len(delta) {
			offset = delta[i]
		}
		out[i] = value + offset
	}
	return out
}

func negate(values []float64) []float64 {
	out := make([]float64, len(values))
	for i, value := range values {
		out[i] = -value
	}
	return out
}

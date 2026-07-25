package main

import (
	"flag"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/pyplot"
	"github.com/cwbudde/matplotlib-go/style"
)

func main() {
	outputDir := flag.String("output-dir", "examples/styling/rc", "directory for rendered rc examples")
	flag.Parse()
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		log.Fatalf("create output directory: %v", err)
	}

	// Runtime rc changes apply to subsequently created pyplot figures.
	pyplot.RCDefaults()
	if err := pyplot.RC("figure", style.Params{
		"dpi":       "144",
		"facecolor": "#faf7f0",
	}); err != nil {
		log.Fatalf("configure figure rc: %v", err)
	}
	if err := pyplot.RC("axes", style.Params{
		"facecolor":  "#fffdf8",
		"edgecolor":  "#3d342c",
		"labelcolor": "#2c241d",
	}); err != nil {
		log.Fatalf("configure axes rc: %v", err)
	}
	if err := pyplot.RC("grid", style.Params{
		"color":     "#d6cbbd",
		"linewidth": "0.75",
	}); err != nil {
		log.Fatalf("configure grid rc: %v", err)
	}

	x := linspace(0, 2*math.Pi, 240)
	y := make([]float64, len(x))
	for i, xv := range x {
		y[i] = math.Sin(xv)
	}

	ax := pyplot.GCA()
	ax.SetTitle("Runtime rc defaults")
	ax.SetXLabel("x")
	ax.SetYLabel("sin(x)")
	ax.AddXGrid()
	ax.AddYGrid()
	if _, err := pyplot.Plot(x, y, core.PlotOptions{Label: "sin(x)"}); err != nil {
		log.Fatalf("plot rc defaults: %v", err)
	}
	pyplot.Legend()
	if err := pyplot.Savefig(filepath.Join(*outputDir, "rc_defaults.png")); err != nil {
		log.Fatalf("save rc_defaults.png: %v", err)
	}

	// RCContext temporarily overrides defaults, matching matplotlib.pyplot.rc_context.
	restore, err := pyplot.RCContext(style.Params{
		"figure.facecolor": "#202733",
		"text.color":       "#f7f3ea",
		"axes.facecolor":   "#273142",
		"axes.edgecolor":   "#d8c7a1",
		"grid.color":       "#607087",
		"lines.color":      "#ffb347",
	})
	if err != nil {
		log.Fatalf("push rc context: %v", err)
	}

	pyplot.Figure()
	ax = pyplot.GCA()
	ax.SetTitle("Temporary rc_context override")
	ax.SetXLabel("x")
	ax.SetYLabel("sin(x)")
	ax.AddXGrid()
	ax.AddYGrid()
	if _, err := pyplot.Plot(x, y, core.PlotOptions{Label: "sin(x)"}); err != nil {
		restore()
		log.Fatalf("plot rc context: %v", err)
	}
	pyplot.Legend()
	if err := pyplot.Savefig(filepath.Join(*outputDir, "rc_context.png")); err != nil {
		restore()
		log.Fatalf("save rc_context.png: %v", err)
	}
	restore()
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

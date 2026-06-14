package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	// Command line flags
	var (
		backendFlag = flag.String("backend", string(backends.AutoBackend), "Rendering backend (auto, agg, gobasic, skia)")
		outputFlag  = flag.String("output", "out.png", "Output filename")
		widthFlag   = flag.Int("width", 640, "Image width")
		heightFlag  = flag.Int("height", 360, "Image height")
		listFlag    = flag.Bool("list", false, "List available backends")
	)
	flag.Parse()

	// List backends if requested
	if *listFlag {
		fmt.Println("Available backends:")
		for _, backend := range backends.Available() {
			info, _ := backends.DefaultRegistry.Get(backend)
			fmt.Printf("  %s - %s\n", backend, info.Description)
		}
		return
	}

	fig := core.NewFigure(*widthFlag, *heightFlag)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})

	ax.SetTitle("Basic Line")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 1)

	// Keep the plotted samples identical to the Python counterpart; only backend
	// selection and capability reporting are Go-specific.
	line := &core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 0.2},
			{X: 3, Y: 0.9},
			{X: 6, Y: 0.4},
			{X: 10, Y: 0.8},
		},
		W:   2.0,
		Col: render.Color{R: 0, G: 0, B: 0, A: 1}, // black line
	}

	ax.Add(line)

	// Create renderer using the registry helper
	config := backends.Config{
		Width:      *widthFlag,
		Height:     *heightFlag,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1}, // white
		DPI:        100,
	}

	renderer, backend, err := backends.NewRenderer(*backendFlag, config, nil)
	if err != nil {
		fmt.Printf("Error creating renderer: %v\n", err)
		os.Exit(1)
	}

	// Render and save
	fmt.Printf("Rendering with %s backend...\n", backend)
	err = core.SavePNG(fig, renderer, *outputFlag)
	if err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created %s using %s backend!\n", *outputFlag, backend)

	// Show backend capabilities used
	fmt.Printf("Backend capabilities: ")
	info, _ := backends.DefaultRegistry.Get(backend)
	for i, cap := range info.Capabilities {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(string(cap))
	}
	fmt.Println()
}

package main

import (
	"flag"
	"fmt"
	"os"

	"matplotlib-go/backends"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/transform"
	_ "matplotlib-go/backends/gobasic" // Register GoBasic
	_ "matplotlib-go/backends/skia"    // Register Skia (stub)
)

func main() {
	// Command line flags
	var (
		backendFlag = flag.String("backend", "gobasic", "Rendering backend (gobasic, skia)")
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

	// Validate backend
	backend := backends.Backend(*backendFlag)
	available := backends.Available()
	found := false
	for _, b := range available {
		if b == backend {
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Error: Backend '%s' not available\n", backend)
		fmt.Println("Available backends:")
		for _, b := range available {
			fmt.Printf("  %s\n", b)
		}
		os.Exit(1)
	}

	// Create figure
	fig := core.NewFigure(*widthFlag, *heightFlag)

	// Add axes
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 1)

	// Create a line with sample data
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

	// Add line to axes
	ax.Add(line)

	// Create renderer using backend factory
	config := backends.Config{
		Width:      *widthFlag,
		Height:     *heightFlag,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1}, // white
		DPI:        72.0,
	}

	renderer, err := backends.Create(backend, config)
	if err != nil {
		fmt.Printf("Error creating %s renderer: %v\n", backend, err)
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
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
	var (
		backendFlag    = flag.String("backend", "", "Rendering backend (leave empty for auto-selection)")
		outputFlag     = flag.String("output", "backend-demo.png", "Output filename")
		widthFlag      = flag.Int("width", 800, "Image width")
		heightFlag     = flag.Int("height", 600, "Image height")
		useCaseFlag    = flag.String("usecase", "basic", "Use case (basic, publication, interactive, scientific)")
		listFlag       = flag.Bool("list", false, "List available backends and exit")
		capabilitiesFlag = flag.Bool("capabilities", false, "Show backend capabilities matrix and exit")
	)
	flag.Parse()

	// List backends if requested
	if *listFlag {
		fmt.Println("Available backends:")
		for _, backend := range backends.Available() {
			info, _ := backends.DefaultRegistry.Get(backend)
			status := "✓ Available"
			if !info.Available {
				status = "✗ Not Available"
			}
			fmt.Printf("  %-10s - %s [%s]\n", backend, info.Description, status)
		}
		return
	}

	// Show capabilities matrix if requested
	if *capabilitiesFlag {
		fmt.Println("Backend Capabilities:")
		fmt.Print(backends.CapabilityMatrix())
		return
	}

	// Select backend
	var backend backends.Backend
	var err error

	if *backendFlag == "" {
		// Auto-select based on use case
		backend, err = backends.GetRecommendedBackend(*useCaseFlag)
		if err != nil {
			fmt.Printf("Error selecting backend for use case '%s': %v\n", *useCaseFlag, err)
			fmt.Println("Available backends:")
			for _, b := range backends.Available() {
				fmt.Printf("  %s\n", b)
			}
			os.Exit(1)
		}
		fmt.Printf("Auto-selected %s backend for %s use case\n", backend, *useCaseFlag)
	} else {
		backend = backends.Backend(*backendFlag)
		available := false
		for _, b := range backends.Available() {
			if b == backend {
				available = true
				break
			}
		}
		if !available {
			fmt.Printf("Backend '%s' is not available\n", backend)
			fmt.Println("Available backends:")
			for _, b := range backends.Available() {
				fmt.Printf("  %s\n", b)
			}
			os.Exit(1)
		}
	}

	// Create demo figure
	fig := createDemoFigure(*widthFlag, *heightFlag)

	// Create renderer
	config := backends.Config{
		Width:      *widthFlag,
		Height:     *heightFlag,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1}, // white background
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

	fmt.Printf("✓ Successfully created %s using %s backend\n", *outputFlag, backend)

	// Show backend info
	info, _ := backends.DefaultRegistry.Get(backend)
	fmt.Printf("Backend: %s - %s\n", info.Name, info.Description)
	fmt.Printf("Capabilities: ")
	for i, cap := range info.Capabilities {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(string(cap))
	}
	fmt.Println()
}

func createDemoFigure(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)

	// Add axes with margins
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-1, 1)

	// Add multiple lines to showcase backend capabilities
	colors := []render.Color{
		{R: 0.2, G: 0.4, B: 0.8, A: 1}, // blue
		{R: 0.8, G: 0.2, B: 0.2, A: 1}, // red
		{R: 0.2, G: 0.8, B: 0.2, A: 1}, // green
	}

	for i, color := range colors {
		points := make([]geom.Pt, 50)
		for j := range points {
			x := float64(j) * 10.0 / 49.0
			y := 0.8 * float64(i-1) * (0.5 + 0.5*sin(x*2.0+float64(i)))
			points[j] = geom.Pt{X: x, Y: y}
		}

		line := &core.Line2D{
			XY:  points,
			W:   2.0 + float64(i),
			Col: color,
		}

		ax.Add(line)
	}

	return fig
}

// Simple sin approximation (since we're not importing math)
func sin(x float64) float64 {
	// Taylor series approximation for sin(x) - good enough for demo
	x = x - float64(int(x/(2*3.14159)))*2*3.14159 // normalize to [-2π, 2π]
	if x < 0 {
		x = -x
		return -(x - x*x*x/6 + x*x*x*x*x/120)
	}
	return x - x*x*x/6 + x*x*x*x*x/120
}
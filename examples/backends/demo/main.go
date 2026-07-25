package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func main() {
	var (
		backendFlag      = flag.String("backend", "", "Rendering backend (leave empty for auto-selection)")
		outputFlag       = flag.String("output", "backend-demo.png", "Output filename")
		widthFlag        = flag.Int("width", 800, "Image width")
		heightFlag       = flag.Int("height", 600, "Image height")
		useCaseFlag      = flag.String("usecase", "basic", "Use case (basic, publication, interactive, scientific)")
		listFlag         = flag.Bool("list", false, "List available backends and exit")
		capabilitiesFlag = flag.Bool("capabilities", false, "Show backend capabilities matrix and exit")
	)
	flag.Parse()

	// Registry-only mode: useful for comparing which Go backends are linked in.
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

	// Matplotlib exposes this differently; the Python counterpart prints its active backend.
	if *capabilitiesFlag {
		fmt.Println("Backend Capabilities:")
		fmt.Print(backends.CapabilityMatrix())
		return
	}

	var backend backends.Backend
	var err error

	if *backendFlag == "" {
		// Auto-select based on the same use-case names listed by backends/info.
		backend, err = backends.RecommendedBackend(*useCaseFlag)
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

	fig := createDemoFigure(*widthFlag, *heightFlag)

	config := backends.Config{
		Width:      *widthFlag,
		Height:     *heightFlag,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100.0,
	}

	renderer, err := backends.Create(backend, config)
	if err != nil {
		fmt.Printf("Error creating %s renderer: %v\n", backend, err)
		os.Exit(1)
	}

	fmt.Printf("Rendering with %s backend...\n", backend)
	err = core.SavePNG(fig, renderer, *outputFlag)
	if err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Successfully created %s using %s backend\n", *outputFlag, backend)

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

	// Keep the plot body close to the Matplotlib counterpart: two trigonometric lines.
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-1, 1)
	ax.SetTitle("Matplotlib-Go backend demo")
	ax.AddXGrid()
	ax.AddYGrid()

	x := make([]float64, 200)
	sinY := make([]float64, len(x))
	cosY := make([]float64, len(x))
	for i := range x {
		x[i] = float64(i) * 10.0 / float64(len(x)-1)
		sinY[i] = math.Sin(x[i])
		cosY[i] = math.Cos(x[i])
	}
	_, _ = ax.Plot(x, sinY, core.PlotOptions{Label: "sin(x)"})
	_, _ = ax.Plot(x, cosY, core.PlotOptions{Label: "cos(x)"})
	ax.AddLegend()

	return fig
}

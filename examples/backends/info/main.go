package main

import (
	"fmt"
	"os"

	"matplotlib-go/backends"
	"matplotlib-go/render"
	_ "matplotlib-go/backends/gobasic" // Register GoBasic
	_ "matplotlib-go/backends/skia"    // Register Skia (stub)
)

func main() {
	fmt.Println("Matplotlib-Go Backend Information")
	fmt.Println("=================================")
	fmt.Println()

	// Show all available backends
	available := backends.Available()
	fmt.Printf("Available backends: %d\n", len(available))
	for _, backend := range available {
		info, _ := backends.DefaultRegistry.Get(backend)
		fmt.Printf("  - %s: %s\n", backend, info.Description)
	}
	fmt.Println()

	// Show capability matrix
	fmt.Println("Backend Capabilities:")
	fmt.Println(backends.CapabilityMatrix())

	// Show recommended backends for common use cases
	fmt.Println("Recommended Backends by Use Case:")
	fmt.Println("---------------------------------")
	
	useCases := []string{"basic", "publication", "interactive", "scientific"}
	for _, useCase := range useCases {
		backend, err := backends.GetRecommendedBackend(useCase)
		if err != nil {
			fmt.Printf("  %s: No suitable backend found\n", useCase)
		} else {
			fmt.Printf("  %s: %s\n", useCase, backend)
		}
	}

	// Test backend creation
	fmt.Println()
	fmt.Println("Backend Creation Test:")
	fmt.Println("----------------------")

	config := backends.SimpleConfig(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})
	
	for _, backend := range available {
		renderer, err := backends.Create(backend, config)
		if err != nil {
			fmt.Printf("  %s: FAILED - %v\n", backend, err)
		} else if renderer == nil {
			fmt.Printf("  %s: FAILED - nil renderer\n", backend)
		} else {
			fmt.Printf("  %s: OK\n", backend)
		}
	}

	// Exit with appropriate code
	if len(available) == 0 {
		fmt.Println("\nWarning: No backends available!")
		os.Exit(1)
	}
}
package backends_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all" // register rendering backends
	"github.com/cwbudde/matplotlib-go/render"
)

// Example selects the best available backend and creates a renderer from it
// without naming a specific backend.
func Example() {
	backend, err := backends.BestBackend(nil)
	if err != nil {
		fmt.Println("no backend available:", err)
		return
	}

	white := render.Color{R: 1, G: 1, B: 1, A: 1}
	r, err := backends.Create(backend, backends.SimpleConfig(640, 480, white))

	fmt.Println("renderer ready:", err == nil && r != nil)

	// Output:
	// renderer ready: true
}

// Example_requireCapabilities asks for a backend that can shape text and hint
// fonts, then confirms the selected backend advertises those capabilities.
func Example_requireCapabilities() {
	required := []backends.Capability{
		backends.TextShaping,
		backends.FontHinting,
	}

	backend, err := backends.BestBackend(required)
	if err != nil {
		fmt.Println("no capable backend:", err)
		return
	}

	ok := backends.HasCapability(backend, backends.TextShaping) &&
		backends.HasCapability(backend, backends.FontHinting)
	fmt.Println("backend satisfies text requirements:", ok)

	// Output:
	// backend satisfies text requirements: true
}

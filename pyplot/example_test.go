package pyplot_test

import (
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/cwbudde/matplotlib-go/backends/all" // register rendering backends
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/pyplot"
)

// Example builds a plot with the stateful pyplot convenience layer and saves
// it to a PNG file, mirroring the typical Matplotlib pyplot workflow.
func Example() {
	pyplot.FigureSized(640, 480)
	if _, err := pyplot.Plot(
		[]float64{0, 1, 2, 3, 4},
		[]float64{0, 1, 4, 9, 16},
		core.PlotOptions{},
	); err != nil {
		fmt.Println("plot failed:", err)
		return
	}
	pyplot.Title("y = x^2")
	pyplot.XLabel("x")
	pyplot.YLabel("y")

	dir, err := os.MkdirTemp("", "pyplot-example")
	if err != nil {
		fmt.Println("temp dir failed:", err)
		return
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "parabola.png")
	if err := pyplot.Savefig(path); err != nil {
		fmt.Println("savefig failed:", err)
		return
	}

	info, err := os.Stat(path)
	fmt.Println("wrote non-empty PNG:", err == nil && info.Size() > 0)

	// Output:
	// wrote non-empty PNG: true
}

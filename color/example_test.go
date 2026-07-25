package color_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/color"
)

// Example looks up a Matplotlib-compatible colormap by name and samples it at
// the midpoint of its normalized [0,1] range.
func Example() {
	cmap := color.LookupColormap("viridis")
	mid := cmap.At(0.5)

	fmt.Printf("%s midpoint: R=%.3f G=%.3f B=%.3f\n",
		cmap.Name(), mid.R, mid.G, mid.B)

	// Output:
	// viridis midpoint: R=0.125 G=0.565 B=0.549
}

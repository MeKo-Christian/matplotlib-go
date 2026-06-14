package svg_test

import (
	"fmt"

	_ "github.com/cwbudde/matplotlib-go/backends/svg"
)

func Example() {
	fmt.Println("SVG backend registered")

	// Output:
	// SVG backend registered
}

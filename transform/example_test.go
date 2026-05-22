package transform_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/transform"
)

// Example shows how a linear axis maps data values onto a display range and
// how the mapping is inverted.
func Example() {
	// Map the data domain [0, 10] onto the pixel range [100, 500].
	axis := transform.NewLinearAxis(0, 10, 100, 500)

	// Forward: data value -> display position.
	fmt.Printf("data 2.5 -> display %.0f\n", axis.Forward(2.5))

	// Inverse: display position -> data value.
	if v, ok := axis.Inverse(300); ok {
		fmt.Printf("display 300 -> data %.1f\n", v)
	}

	// Output:
	// data 2.5 -> display 200
	// display 300 -> data 5.0
}

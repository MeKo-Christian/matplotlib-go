package style_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/style"
)

// Example applies a set of functional rc options to the current defaults,
// producing a derived RC without mutating global state.
func Example() {
	base := style.CurrentDefaults()

	rc := style.Apply(base,
		style.WithDPI(150),
		style.WithLineWidth(2.5),
	)

	fmt.Printf("base DPI=%.0f  derived DPI=%.0f lw=%.1f\n",
		base.DPI, rc.DPI, rc.LineWidth)

	// Output:
	// base DPI=100  derived DPI=150 lw=2.5
}

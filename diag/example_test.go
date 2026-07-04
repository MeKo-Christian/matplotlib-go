package diag_test

import (
	"github.com/cwbudde/matplotlib-go/diag"
)

// ExampleSetHandler silences every matplotlib-go diagnostic warning until the
// returned restore func is called.
func ExampleSetHandler() {
	restore := diag.SetHandler(nil)

	// ... use matplotlib-go without any warning output ...

	restore()
	// Output:
}

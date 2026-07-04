package diag_test

import (
	"github.com/cwbudde/matplotlib-go/diag"
)

// ExampleSetHandler silences every matplotlib-go diagnostic warning for the
// lifetime of the program (or until the returned restore func is called).
func ExampleSetHandler() {
	restore := diag.SetHandler(nil)
	defer restore()

	// ... use matplotlib-go without any warning output ...

	// Output:
}

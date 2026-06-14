package ps_test

import (
	"fmt"

	_ "github.com/cwbudde/matplotlib-go/backends/ps"
)

func Example() {
	fmt.Println("PostScript backend registered")

	// Output:
	// PostScript backend registered
}

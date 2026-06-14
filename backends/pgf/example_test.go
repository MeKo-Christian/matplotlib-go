package pgf_test

import (
	"fmt"

	_ "github.com/cwbudde/matplotlib-go/backends/pgf"
)

func Example() {
	fmt.Println("PGF backend registered")

	// Output:
	// PGF backend registered
}

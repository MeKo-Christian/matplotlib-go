package agg_test

import (
	"fmt"

	_ "github.com/cwbudde/matplotlib-go/backends/agg"
)

func Example() {
	fmt.Println("AGG backend registered")

	// Output:
	// AGG backend registered
}

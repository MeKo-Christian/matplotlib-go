package gobasic_test

import (
	"fmt"

	_ "github.com/cwbudde/matplotlib-go/backends/gobasic"
)

func Example() {
	fmt.Println("GoBasic backend registered")

	// Output:
	// GoBasic backend registered
}

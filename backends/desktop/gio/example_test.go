package gio_test

import (
	"fmt"

	_ "github.com/cwbudde/matplotlib-go/backends/desktop/gio"
)

func Example() {
	fmt.Println("Gio desktop backend registered")

	// Output:
	// Gio desktop backend registered
}

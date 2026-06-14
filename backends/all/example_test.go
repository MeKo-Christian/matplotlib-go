package all_test

import (
	"fmt"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
)

func Example() {
	fmt.Println("built-in backends registered")

	// Output:
	// built-in backends registered
}

package pdf_test

import (
	"fmt"

	_ "github.com/cwbudde/matplotlib-go/backends/pdf"
)

func Example() {
	fmt.Println("PDF backend registered")

	// Output:
	// PDF backend registered
}

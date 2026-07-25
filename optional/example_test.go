package optional_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/optional"
)

// ExampleValue_Or shows the distinction that plain and pointer option fields
// could not make: an omitted option falls back to its default, while an
// explicit zero is honored.
func ExampleValue_Or() {
	type imageOptions struct {
		Alpha optional.Value[float64]
	}

	defaulted := imageOptions{}
	opaque := imageOptions{Alpha: optional.Of(1.0)}
	transparent := imageOptions{Alpha: optional.Of(0.0)}

	for _, opt := range []imageOptions{defaulted, opaque, transparent} {
		fmt.Println(opt.Alpha.Or(1))
	}
	// Output:
	// 1
	// 1
	// 0
}

func ExampleValue_Get() {
	set := optional.Of("bilinear")
	if filter, ok := set.Get(); ok {
		fmt.Println("filter:", filter)
	}

	var unset optional.Value[string]
	if _, ok := unset.Get(); !ok {
		fmt.Println("filter: renderer default")
	}
	// Output:
	// filter: bilinear
	// filter: renderer default
}

package tri_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/tri"
)

// ExampleNew builds a Delaunay triangulation of the unit square and reports
// the number of triangles produced.
func ExampleNew() {
	x := []float64{0, 1, 1, 0}
	y := []float64{0, 0, 1, 1}

	t, err := tri.New(x, y)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("triangles:", len(t.Triangles))
	// Output: triangles: 2
}

package geom_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/geom"
)

func ExampleRect_Union() {
	a := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 2, Y: 1}}
	b := geom.Rect{Min: geom.Pt{X: 1, Y: -1}, Max: geom.Pt{X: 3, Y: 2}}

	union := a.Union(b)
	fmt.Println(union.Min, union.Max)
	// Output: {0 -1} {3 2}
}

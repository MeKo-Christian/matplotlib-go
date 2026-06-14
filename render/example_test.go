package render_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Example_color shows the normalized sRGBA color contract: components live in
// [0,1] and Premultiply scales the color channels by alpha for compositing.
func Example_color() {
	c := render.Color{R: 1, G: 0.5, B: 0, A: 0.5}
	p := c.Premultiply()

	fmt.Printf("straight:      R=%.2f G=%.2f B=%.2f A=%.2f\n", c.R, c.G, c.B, c.A)
	fmt.Printf("premultiplied: R=%.2f G=%.2f B=%.2f A=%.2f\n", p.R, p.G, p.B, p.A)

	// Output:
	// straight:      R=1.00 G=0.50 B=0.00 A=0.50
	// premultiplied: R=0.50 G=0.25 B=0.00 A=0.50
}

// Example_nullRenderer uses NullRenderer to validate that a draw pass keeps
// the state stack balanced without producing any output.
func Example_nullRenderer() {
	var r render.NullRenderer
	vp := geom.Rect{Max: geom.Pt{X: 100, Y: 100}}

	_ = r.Begin(vp)
	r.Save()
	r.Restore()
	err := r.End()

	fmt.Println("draw pass balanced:", err == nil)

	// Output:
	// draw pass balanced: true
}

package canvas_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/geom"
)

// Example connects a handler to a Dispatcher and emits a mouse-press event to
// it, demonstrating the interactive event-delivery path.
func Example() {
	fig := &canvas.Figure{}
	var d canvas.Dispatcher

	d.Connect(canvas.EventMousePress, func(e canvas.Event) error {
		fmt.Printf("%s at (%.0f, %.0f)\n", e.Type, e.Position.X, e.Position.Y)
		return nil
	})

	press := canvas.NewMouseEvent(
		canvas.EventMousePress, fig, geom.Pt{X: 120, Y: 80}, canvas.MouseButton(1),
	)
	_ = d.Emit(press.Event)

	// Output:
	// mouse_press at (120, 80)
}

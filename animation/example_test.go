package animation_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/animation"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
)

type exampleCanvas struct {
	fig *core.Figure
}

func (c *exampleCanvas) Figure() *core.Figure { return c.fig }
func (c *exampleCanvas) Draw() error          { return nil }
func (c *exampleCanvas) Resize(_, _ int) error {
	return nil
}
func (c *exampleCanvas) Connect(canvas.EventType, canvas.Handler) canvas.ConnectionID {
	return 0
}
func (c *exampleCanvas) Disconnect(canvas.ConnectionID) {}
func (c *exampleCanvas) Close() error                   { return nil }

func Example() {
	cnv := &exampleCanvas{fig: core.NewFigure(320, 240)}
	anim, err := animation.NewFuncAnimation(
		animation.Config{Canvas: cnv, Frames: 2},
		func(frame int) ([]core.Artist, error) {
			fmt.Printf("frame %d\n", frame)
			return nil, nil
		},
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, _ = anim.Step()
	_, _ = anim.Step()

	// Output:
	// frame 0
	// frame 1
}

package widgets_test

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/widgets"
)

func ExampleNewButton() {
	fig := core.NewFigure(320, 200)
	ax := widgets.NewAxes(fig, geom.Rect{
		Min: geom.Pt{X: 0.2, Y: 0.4},
		Max: geom.Pt{X: 0.8, Y: 0.6},
	})
	button := widgets.NewButton(ax, "Apply")
	button.OnClicked(func(*widgets.Button) {
		fmt.Println("applied")
	})

	button.Click()
	// Output:
	// applied
}

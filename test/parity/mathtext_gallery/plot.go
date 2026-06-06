package mathtext_gallery

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	showcase "github.com/cwbudde/matplotlib-go/examples/mathtext_gallery"
)

func Render() image.Image {
	return showcase.Render()
}

func Plot() *core.Figure {
	return showcase.Plot()
}

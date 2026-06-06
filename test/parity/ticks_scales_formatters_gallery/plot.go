// Package ticks_scales_formatters_gallery is the parity-test wrapper for the
// ticks_scales_formatters_gallery showcase. The canonical rendering body lives
// in github.com/cwbudde/matplotlib-go/examples/ticks_scales_formatters_gallery.
package ticks_scales_formatters_gallery

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	showcase "github.com/cwbudde/matplotlib-go/examples/ticks_scales_formatters_gallery"
)

func Render() image.Image {
	return showcase.Render()
}

func Plot() *core.Figure {
	return showcase.Plot()
}

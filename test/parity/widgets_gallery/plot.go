// Package widgets_gallery is the parity-test wrapper for the widgets_gallery showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/widgets_gallery.
package widgets_gallery

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	showcase "github.com/cwbudde/matplotlib-go/examples/widgets_gallery"
	"github.com/cwbudde/matplotlib-go/style"
)

// Plot returns the parity figure with Matplotlib-compatible widget visuals.
func Plot() *core.Figure {
	return showcase.Plot(style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
}

// Render returns the parity image with Matplotlib-compatible widget visuals.
func Render() image.Image {
	return showcase.Render(style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
}

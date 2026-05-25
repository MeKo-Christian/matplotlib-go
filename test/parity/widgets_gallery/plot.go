// Package widgets_gallery is the parity-test wrapper for the widgets_gallery showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/widgets_gallery.
package widgets_gallery

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/widgets_gallery"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

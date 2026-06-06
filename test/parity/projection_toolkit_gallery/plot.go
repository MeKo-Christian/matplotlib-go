// Package projection_toolkit_gallery is the parity-test wrapper for the
// projection_toolkit_gallery showcase. The canonical rendering body lives in
// github.com/cwbudde/matplotlib-go/examples/projection_toolkit_gallery.
package projection_toolkit_gallery

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	showcase "github.com/cwbudde/matplotlib-go/examples/projection_toolkit_gallery"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

// Plot returns the backend-agnostic showcase figure.
func Plot() *core.Figure {
	return showcase.Plot()
}

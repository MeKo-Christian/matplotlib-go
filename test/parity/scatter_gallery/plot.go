// Package scatter_gallery is the parity-test wrapper for the scatter_gallery
// showcase. The canonical rendering body lives in
// github.com/cwbudde/matplotlib-go/examples/scatter_gallery; this file imports
// it so the parity registry and golden tests share that single source of truth.
package scatter_gallery

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/scatter_gallery"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

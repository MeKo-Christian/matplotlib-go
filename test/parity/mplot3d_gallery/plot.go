// Package mplot3d_gallery is the parity-test wrapper for the mplot3d_gallery showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/mplot3d_gallery.
package mplot3d_gallery

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	showcase "github.com/cwbudde/matplotlib-go/examples/mplot3d_gallery"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

// Plot returns the backend-agnostic showcase figure.
func Plot() *core.Figure {
	return showcase.Plot()
}

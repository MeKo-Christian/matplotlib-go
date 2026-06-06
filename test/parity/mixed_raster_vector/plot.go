// Package mixed_raster_vector is the parity-test wrapper for the mixed_raster_vector showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/mixed_raster_vector.
package mixed_raster_vector

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	showcase "github.com/cwbudde/matplotlib-go/examples/mixed_raster_vector"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

// Plot returns the backend-agnostic showcase figure.
func Plot() *core.Figure {
	return showcase.Plot()
}

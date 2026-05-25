// Package animation_gallery is the parity-test wrapper for the animation gallery showcase.
package animation_gallery

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/animation_gallery"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

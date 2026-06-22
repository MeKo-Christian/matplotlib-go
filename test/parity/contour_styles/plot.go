// Package contour_styles is the parity-test wrapper for the contour_styles showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/contour_styles;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package contour_styles

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/contour_styles"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

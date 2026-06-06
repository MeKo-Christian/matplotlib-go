// Package histogram_variants is the parity-test wrapper for the histogram_variants showcase. The
// canonical rendering body lives in
// github.com/cwbudde/matplotlib-go/examples/histogram_variants; this file imports it so the
// parity registry and golden tests share that single source of truth.
package histogram_variants

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/histogram_variants"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

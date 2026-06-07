// Package animation_scatter_frame is the parity-test wrapper for the
// deterministic animated-scatter frame fixture. The canonical rendering body
// lives in github.com/cwbudde/matplotlib-go/examples/animation_gallery; this
// file imports it so the parity registry and golden tests share that single
// source of truth.
package animation_scatter_frame

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/animation_gallery"
)

// Render returns the animated-scatter fixture image at the golden frame.
func Render() image.Image {
	return showcase.RenderScatterFrame()
}

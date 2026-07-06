// Package quiver_autoscale is the parity-test wrapper for the quiver_autoscale showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/quiver_autoscale;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package quiver_autoscale

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/quiver_autoscale"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

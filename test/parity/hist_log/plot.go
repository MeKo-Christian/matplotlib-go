// Package hist_log is the parity-test wrapper for the hist_log showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/hist_log;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package hist_log

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/hist_log"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

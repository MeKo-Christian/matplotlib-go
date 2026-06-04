// Package basic_line_labels is the parity-test wrapper for the
// basic_line_labels showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/basic_line_labels;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package basic_line_labels

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/basic_line_labels"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}

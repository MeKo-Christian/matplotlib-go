//go:build !skia || js

package webdemo

import "github.com/cwbudde/matplotlib-go/render"

func optionalBackendDescriptors() []BackendDescriptor {
	return nil
}

func newOptionalRasterRenderer(_ string, _ int, _ int, _ render.Color) (rasterRenderer, bool, error) {
	return nil, false, nil
}

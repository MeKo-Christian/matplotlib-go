package core

import (
	"image"

	"github.com/cwbudde/matplotlib-go/render"
)

const offsetImageDefaultInterpolation = "antialiased"

type offsetBoxImage struct {
	render.Image
	interpolation string
}

func offsetImageWithMatplotlibDefaults(img render.Image) render.Image {
	if img == nil || img.Interpolation() != "" {
		return img
	}
	return offsetBoxImage{Image: img, interpolation: offsetImageDefaultInterpolation}
}

func (i offsetBoxImage) Interpolation() string {
	if i.interpolation != "" {
		return i.interpolation
	}
	if i.Image == nil {
		return ""
	}
	return i.Image.Interpolation()
}

func (i offsetBoxImage) RGBA() *image.RGBA {
	if rgba, ok := i.Image.(render.RGBAImage); ok {
		return rgba.RGBA()
	}
	return nil
}

func (i offsetBoxImage) Alpha() float64 {
	if alpha, ok := i.Image.(render.ImageAlpha); ok {
		return alpha.Alpha()
	}
	return 1
}

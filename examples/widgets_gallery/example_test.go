package widgets_gallery

import (
	"image"
	"image/color"
	"testing"
)

func TestRenderProducesNonBlankImage(t *testing.T) {
	img := Render()
	if !hasNonWhitePixel(img) {
		t.Fatal("widgets gallery render is blank")
	}
}

func hasNonWhitePixel(img image.Image) bool {
	if img == nil {
		return false
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if color.RGBAModel.Convert(img.At(x, y)).(color.RGBA) != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
				return true
			}
		}
	}
	return false
}

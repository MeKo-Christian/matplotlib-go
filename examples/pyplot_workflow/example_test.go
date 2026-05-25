package pyplot_workflow

import (
	"image"
	"image/color"
	"testing"
)

func TestPlotBuildsPyplotWorkflow(t *testing.T) {
	fig := Plot()
	if fig == nil {
		t.Fatal("Plot returned nil")
	}
	if got := len(fig.Children); got < 5 {
		t.Fatalf("figure axes = %d, want four subplots plus colorbar", got)
	}
}

func TestRenderProducesNonBlankImage(t *testing.T) {
	img := Render()
	if !hasNonWhitePixel(img) {
		t.Fatal("pyplot workflow render is blank")
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

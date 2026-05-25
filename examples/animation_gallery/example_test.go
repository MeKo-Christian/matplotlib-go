package animation_gallery

import (
	"errors"
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/animation"
)

func TestRenderProducesNonBlankImage(t *testing.T) {
	img := Render()
	if !hasNonWhitePixel(img) {
		t.Fatal("animation gallery render is blank")
	}
}

func TestFuncAnimationDemoStepsAndReportsUnsupportedSave(t *testing.T) {
	demo, err := NewFuncAnimationDemo()
	if err != nil {
		t.Fatalf("NewFuncAnimationDemo: %v", err)
	}
	if demo.Animation.Config().Frames != 12 {
		t.Fatalf("func animation frames = %d, want 12", demo.Animation.Config().Frames)
	}
	if _, err := demo.Animation.Step(); err != nil {
		t.Fatalf("func animation step: %v", err)
	}
	if demo.Animation.TotalFramesDrawn() != 1 {
		t.Fatalf("func animation drawn frames = %d, want 1", demo.Animation.TotalFramesDrawn())
	}
	if err := demo.Animation.Save("out.gif"); !errors.Is(err, animation.ErrWriterUnsupported) {
		t.Fatalf("func animation Save error = %v, want ErrWriterUnsupported", err)
	}
}

func TestArtistAnimationDemoSteps(t *testing.T) {
	demo, err := NewArtistAnimationDemo()
	if err != nil {
		t.Fatalf("NewArtistAnimationDemo: %v", err)
	}
	if len(demo.Artists) != 3 {
		t.Fatalf("artist animation artists = %d, want 3", len(demo.Artists))
	}
	for i := 0; i < 3; i++ {
		if _, err := demo.Animation.Step(); err != nil {
			t.Fatalf("artist animation step %d: %v", i, err)
		}
	}
	if demo.Animation.TotalFramesDrawn() != 3 {
		t.Fatalf("artist animation drawn frames = %d, want 3", demo.Animation.TotalFramesDrawn())
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

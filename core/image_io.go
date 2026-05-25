package core

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/cwbudde/matplotlib-go/render"
)

// ErrImageIOUnsupported marks image IO operations outside the supported subset.
var ErrImageIOUnsupported = errors.New("core: image IO unsupported")

// ImRead decodes an image file into renderer-facing RGBA image data.
func ImRead(path string) (*render.ImageData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoded, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return render.NewImageData(toRGBA(decoded)), nil
}

// ImSave writes image data to path. PNG output is supported for RGBA-backed images.
func ImSave(path string, img render.Image) error {
	if strings.ToLower(filepath.Ext(path)) != ".png" {
		return fmt.Errorf("%w: only png output is supported", ErrImageIOUnsupported)
	}
	rgbaSource, ok := img.(render.RGBAImage)
	if !ok || rgbaSource.RGBA() == nil {
		return fmt.Errorf("%w: imsave requires rgba image data", ErrImageIOUnsupported)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	if err := png.Encode(file, toNRGBA(rgbaSource.RGBA())); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func toRGBA(src image.Image) *image.RGBA {
	if src == nil {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}
	if rgba, ok := src.(*image.RGBA); ok {
		return rgba
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	if nrgba, ok := src.(*image.NRGBA); ok {
		for y := 0; y < bounds.Dy(); y++ {
			for x := 0; x < bounds.Dx(); x++ {
				dst.SetRGBA(x, y, color.RGBA(nrgba.NRGBAAt(bounds.Min.X+x, bounds.Min.Y+y)))
			}
		}
		return dst
	}
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			c := color.NRGBAModel.Convert(src.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.NRGBA)
			dst.SetRGBA(x, y, color.RGBA(c))
		}
	}
	return dst
}

func toNRGBA(src *image.RGBA) *image.NRGBA {
	if src == nil {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}
	bounds := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			c := src.RGBAAt(bounds.Min.X+x, bounds.Min.Y+y)
			dst.SetNRGBA(x, y, color.NRGBA(c))
		}
	}
	return dst
}

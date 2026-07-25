package core

import (
	"errors"
	"image"
	"image/png"
	"math"
	"os"

	"github.com/cwbudde/matplotlib-go/render"
)

// savePNGTight writes a tight-bbox cropped PNG (savefig.bbox="tight"). It scans
// the rendered raster for the bounding box of content that differs from the
// figure background, pads it by savefig.pad_inches, and encodes the crop.
//
// The AGG buffer holds straight (non-premultiplied) alpha, so the RGBA bytes are
// reinterpreted as NRGBA for both scanning and encoding, mirroring the backend's
// own SavePNG.
func savePNGTight(r render.Renderer, eff *Figure, resolved *resolvedSaveFigure, path string) error {
	provider, ok := r.(render.RGBAExporter)
	if !ok {
		return errors.New("savefig: bbox=tight is only supported for raster (PNG) output")
	}
	img := provider.Image()
	if img == nil {
		return errors.New("savefig: renderer returned no image for tight bbox")
	}
	src := &image.NRGBA{Pix: img.Pix, Stride: img.Stride, Rect: img.Rect}

	bg := eff.RC.FigureBackground()
	crop := tightContentBounds(src, bg, resolved.transparent)
	if crop.Empty() {
		crop = src.Rect
	} else {
		pad := int(math.Round(resolved.padInches * eff.RC.DPI))
		if pad > 0 {
			crop.Min.X -= pad
			crop.Min.Y -= pad
			crop.Max.X += pad
			crop.Max.Y += pad
		}
		crop = crop.Intersect(src.Rect)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, src.SubImage(crop))
}

// tightContentBounds returns the bounding rectangle of pixels that differ from
// the background. For transparent output, any non-zero alpha counts as content.
// An empty rectangle means nothing was drawn.
func tightContentBounds(img *image.NRGBA, bg render.Color, transparent bool) image.Rectangle {
	b := img.Bounds()
	bgR, bgG, bgB, bgA := channelByte(bg.R), channelByte(bg.G), channelByte(bg.B), channelByte(bg.A)

	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y
	found := false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		row := img.PixOffset(b.Min.X, y)
		for x := b.Min.X; x < b.Max.X; x++ {
			i := row + (x-b.Min.X)*4
			rr, gg, bb, aa := img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3]
			var content bool
			if transparent {
				content = aa != 0
			} else {
				content = rr != bgR || gg != bgG || bb != bgB || aa != bgA
			}
			if !content {
				continue
			}
			found = true
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x+1 > maxX {
				maxX = x + 1
			}
			if y+1 > maxY {
				maxY = y + 1
			}
		}
	}
	if !found {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX, maxY)
}

// channelByte converts a normalized sRGBA channel (0..1) to an 8-bit value.
func channelByte(value float64) uint8 {
	switch {
	case value <= 0:
		return 0
	case value >= 1:
		return 0xFF
	default:
		return uint8(value*255 + 0.5)
	}
}

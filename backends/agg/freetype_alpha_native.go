//go:build cgo && !purego

package agg

/*
#cgo !systemfreetype CFLAGS: -I${SRCDIR}/../../third_party/freetype/prefix/include/freetype2
#cgo !systemfreetype LDFLAGS: -L${SRCDIR}/../../third_party/freetype/prefix/lib -lfreetype -lm
#cgo systemfreetype pkg-config: freetype2
#include <ft2build.h>
#include FT_FREETYPE_H

static int mpl_go_alpha_pixel_mode_gray(void) {
	return FT_PIXEL_MODE_GRAY;
}

static int mpl_go_alpha_pixel_mode_mono(void) {
	return FT_PIXEL_MODE_MONO;
}
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/cwbudde/matplotlib-go/render"
)

func orAlphaMask(dst, src *image.Alpha, x, y int) {
	for sy := src.Bounds().Min.Y; sy < src.Bounds().Max.Y; sy++ {
		dy := y + sy
		if dy < dst.Bounds().Min.Y || dy >= dst.Bounds().Max.Y {
			continue
		}
		for sx := src.Bounds().Min.X; sx < src.Bounds().Max.X; sx++ {
			dx := x + sx
			if dx < dst.Bounds().Min.X || dx >= dst.Bounds().Max.X {
				continue
			}
			dstOff := dst.PixOffset(dx, dy)
			srcOff := src.PixOffset(sx, sy)
			dst.Pix[dstOff] |= src.Pix[srcOff]
		}
	}
}

func (r *Renderer) blendAlphaMask(mask *image.Alpha, dstX, dstY int, textColor render.Color) bool {
	if r.ctx == nil || r.ctx.image == nil || mask == nil {
		return false
	}

	paint := renderColorToRGBA(textColor)
	data := r.ctx.image.Data
	stride := r.ctx.image.Stride()
	width, height := r.ctx.image.Width(), r.ctx.image.Height()
	if stride <= 0 || len(data) == 0 {
		return false
	}

	drew := false
	for my := mask.Bounds().Min.Y; my < mask.Bounds().Max.Y; my++ {
		y := dstY + my
		if y < 0 || y >= height {
			continue
		}
		for mx := mask.Bounds().Min.X; mx < mask.Bounds().Max.X; mx++ {
			x := dstX + mx
			if x < 0 || x >= width {
				continue
			}
			cover := uint32(mask.Pix[mask.PixOffset(mx, my)])
			if cover == 0 {
				continue
			}
			srcA := uint32(paint.A) * cover / 255
			if srcA == 0 {
				continue
			}
			off := y*stride + x*4
			invA := 255 - srcA
			data[off+0] = uint8((uint32(paint.R)*srcA + uint32(data[off+0])*invA + 127) / 255)
			data[off+1] = uint8((uint32(paint.G)*srcA + uint32(data[off+1])*invA + 127) / 255)
			data[off+2] = uint8((uint32(paint.B)*srcA + uint32(data[off+2])*invA + 127) / 255)
			data[off+3] = uint8(srcA + uint32(data[off+3])*invA/255)
			drew = true
		}
	}
	return drew
}

func fillAlphaRect(mask *image.Alpha, x0, y0, x1, y1 int) {
	if mask == nil {
		return
	}
	bounds := mask.Bounds()
	rect := image.Rect(x0, y0, x1, y1).Intersect(bounds)
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		row := mask.PixOffset(rect.Min.X, y)
		for x := rect.Min.X; x < rect.Max.X; x++ {
			mask.Pix[row+x-rect.Min.X] = 0xff
		}
	}
}

func overAlphaMask(dst, src *image.Alpha, x0, y0 int) {
	if dst == nil || src == nil {
		return
	}
	dstBounds := dst.Bounds()
	srcBounds := src.Bounds()
	for sy := srcBounds.Min.Y; sy < srcBounds.Max.Y; sy++ {
		dy := y0 + sy - srcBounds.Min.Y
		if dy < dstBounds.Min.Y || dy >= dstBounds.Max.Y {
			continue
		}
		for sx := srcBounds.Min.X; sx < srcBounds.Max.X; sx++ {
			dx := x0 + sx - srcBounds.Min.X
			if dx < dstBounds.Min.X || dx >= dstBounds.Max.X {
				continue
			}
			srcA := uint32(src.Pix[src.PixOffset(sx, sy)])
			if srcA == 0 {
				continue
			}
			di := dst.PixOffset(dx, dy)
			dstA := uint32(dst.Pix[di])
			dst.Pix[di] = uint8(srcA + dstA*(255-srcA)/255)
		}
	}
}

// fillDeviceRect blends a solid color over the half-open device rect
// [x0,x1) x [y0,y1) (matplotlib FT2Image.draw_rect_filled is inclusive; callers
// pass the +1-adjusted half-open bounds).
func (r *Renderer) fillDeviceRect(x0, y0, x1, y1 int, textColor render.Color) {
	if x1 <= x0 || y1 <= y0 {
		return
	}
	mask := image.NewAlpha(image.Rect(0, 0, x1-x0, y1-y0))
	for i := range mask.Pix {
		mask.Pix[i] = 0xff
	}
	r.blendAlphaMask(mask, x0, y0, textColor)
}

func freetypeBitmapMask(bitmap C.FT_Bitmap) (*image.Alpha, bool) {
	width := int(bitmap.width)
	height := int(bitmap.rows)
	pitch := int(bitmap.pitch)
	if width <= 0 || height <= 0 || pitch == 0 || bitmap.buffer == nil {
		return nil, false
	}

	mask := image.NewAlpha(image.Rect(0, 0, width, height))
	bufferLen := absInt(pitch) * height
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bitmap.buffer)), bufferLen)
	pixelMode := int(bitmap.pixel_mode)

	for row := 0; row < height; row++ {
		srcRow := row * pitch
		if pitch < 0 {
			srcRow = (height - 1 - row) * -pitch
		}
		dstRow := row * mask.Stride
		switch pixelMode {
		case int(C.mpl_go_alpha_pixel_mode_gray()):
			copy(mask.Pix[dstRow:dstRow+width], buffer[srcRow:srcRow+width])
		case int(C.mpl_go_alpha_pixel_mode_mono()):
			for col := 0; col < width; col++ {
				byteIndex := srcRow + col/8
				bit := byte(1 << uint(7-col%8))
				if buffer[byteIndex]&bit != 0 {
					mask.Pix[dstRow+col] = 0xff
				}
			}
		default:
			return nil, false
		}
	}
	return mask, true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

package core

import (
	"fmt"
	"image"
	"image/color"

	"github.com/cwbudde/matplotlib-go/internal/diag"
)

// rgbArrayKind classifies a 3D imshow input array by its channel count.
type rgbArrayKind int

const (
	rgbArrayInvalid rgbArrayKind = iota
	rgbArrayScalar               // (M,N,1): squeeze to scalar, apply colormap
	rgbArrayRGB                  // (M,N,3): opaque alpha added
	rgbArrayRGBA                 // (M,N,4): alpha preserved
)

// normalizeRGBArray converts an (M,N,3) or (M,N,4) float array (channel values
// in [0,1]) into an *image.RGBA, faithfully porting matplotlib's
// _ImageBase._normalize_image_array (image.py): out-of-range values are clipped
// to [0,1] with a warning. Rows must be rectangular and every pixel must carry
// the same channel count.
//
// A (M,N,1) array is reported as rgbArrayScalar so the caller can route it to
// the scalar colormap path; in that case the returned image is nil.
func normalizeRGBArray(data [][][]float64) (*image.RGBA, rgbArrayKind, error) {
	rows := len(data)
	if rows == 0 {
		return nil, rgbArrayInvalid, fmt.Errorf("core: imshow RGB array is empty")
	}
	cols := len(data[0])
	if cols == 0 {
		return nil, rgbArrayInvalid, fmt.Errorf("core: imshow RGB array has empty rows")
	}
	channels := len(data[0][0])
	switch channels {
	case 1:
		return nil, rgbArrayScalar, nil
	case 3, 4:
	default:
		return nil, rgbArrayInvalid, fmt.Errorf("core: imshow RGB array must have 3 or 4 channels, got %d", channels)
	}

	// Validate rectangular shape up front so partial conversions never occur.
	for y := range rows {
		if len(data[y]) != cols {
			return nil, rgbArrayInvalid, fmt.Errorf("core: imshow RGB array row %d has %d columns, want %d", y, len(data[y]), cols)
		}
		for x := range cols {
			if len(data[y][x]) != channels {
				return nil, rgbArrayInvalid, fmt.Errorf("core: imshow RGB array pixel (%d,%d) has %d channels, want %d", y, x, len(data[y][x]), channels)
			}
		}
	}

	clipped := false
	clip := func(v float64) float64 {
		if v < 0 {
			clipped = true
			return 0
		}
		if v > 1 {
			clipped = true
			return 1
		}
		return v
	}

	img := image.NewRGBA(image.Rect(0, 0, cols, rows))
	for y := range rows {
		for x := range cols {
			px := data[y][x]
			a := 1.0
			if channels == 4 {
				a = clip(px[3])
			}
			// Store straight (non-premultiplied) RGBA: the AGG backend
			// premultiplies by the effective alpha at draw time
			// (backends/agg/agg_clip.go), so premultiplying here would double up.
			img.SetRGBA(x, y, color.RGBA{
				R: floatChannelToByte(clip(px[0])),
				G: floatChannelToByte(clip(px[1])),
				B: floatChannelToByte(clip(px[2])),
				A: floatChannelToByte(a),
			})
		}
	}

	if clipped {
		diag.Warnf("imshow: clipping input data to the valid range for RGB(A) images ([0..1] for floats)")
	}
	return img, rgbArrayKindForChannels(channels), nil
}

func rgbArrayKindForChannels(channels int) rgbArrayKind {
	if channels == 4 {
		return rgbArrayRGBA
	}
	return rgbArrayRGB
}

func floatChannelToByte(v float64) uint8 {
	// Truncate toward zero over [0,1] → [0,255], matching matplotlib's
	// (x * 255).astype(uint8) byte conversion (colorizer._pass_image_data).
	// v is already clipped to [0,1].
	return uint8(v * 255)
}

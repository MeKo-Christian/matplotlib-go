// Package imageinterp holds the image-interpolation policy shared by the
// backends, so AGG and SVG cannot disagree about when a raster is drawn
// unfiltered. It is internal: this is a backend detail, not public API.
package imageinterp

import (
	"math"
	"strings"
)

// Nearest and Hanning are the two concrete filters Matplotlib's
// "antialiased"/"auto" policy can resolve to.
const (
	Nearest = "nearest"
	Hanning = "hanning"
)

// IsAdaptive reports whether name selects Matplotlib's
// scale-dependent policy rather than a fixed filter. The empty name is *not*
// adaptive: it means "renderer default", which every backend treats as
// nearest-neighbour.
func IsAdaptive(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "auto", "antialiased":
		return true
	default:
		return false
	}
}

// Resolve resolves an interpolation name against the source and
// destination extents, mirroring Matplotlib's _ImageBase._make_image: an
// "antialiased" image is drawn with nearest-neighbour when it is upsampled by
// an integer factor or by more than 3x in both directions, and with a hanning
// filter otherwise. Fixed filter names pass through untouched, so callers can
// resolve unconditionally.
//
// Non-positive extents leave the name alone: there is nothing to compare, and
// guessing would silently change how the image is drawn.
func Resolve(name string, srcW, srcH, dstW, dstH float64) string {
	if !IsAdaptive(name) {
		return name
	}

	if srcW <= 0 || srcH <= 0 || dstW <= 0 || dstH <= 0 {
		return Hanning
	}

	if upsampledWithoutResampling(dstW, srcW) && upsampledWithoutResampling(dstH, srcH) {
		return Nearest
	}

	return Hanning
}

// IsNearest reports whether name draws pixels unfiltered. The
// empty name counts: it defers to the renderer default, which is nearest.
func IsNearest(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "none", Nearest:
		return true
	default:
		return false
	}
}

func upsampledWithoutResampling(dst, src float64) bool {
	return dst > 3*src || floatAlmostEqual(dst, src) || floatAlmostEqual(dst, 2*src)
}

func floatAlmostEqual(a, b float64) bool {
	const eps = 1e-9

	return math.Abs(a-b) <= eps*math.Max(1.0, math.Max(math.Abs(a), math.Abs(b)))
}

package agg

import (
	"math"
	"strings"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/imageinterp"
	"github.com/cwbudde/matplotlib-go/render"
)

// parseInterpolationName maps a matplotlib-style interpolation string to an
// agg_go ImageFilter. The bool reports whether the name was recognised; the
// returned filter falls back to NoFilter for unrecognised inputs (caller may
// log a warning or surface the failure as the application sees fit). The
// second return value reports whether interpolation was requested as a
// Matplotlib “auto“/“antialiased“ policy and therefore needs source-vs-
// destination scale inspection before a concrete filter can be chosen.
//
// Recognised names (case-insensitive, surrounding whitespace trimmed):
//   - ""        — same as "none" / "nearest"; no filtering
//   - "none"    — no filtering (nearest-neighbour)
//   - "nearest" — alias for "none"
//   - "bilinear", "bicubic", "spline16", "spline36", "hanning", "hamming",
//     "hermite", "kaiser", "quadric", "catrom", "gaussian", "bessel",
//     "mitchell", "sinc", "lanczos", and "blackman" — direct mappings
//     to agg_go filters.
//   - "auto", "antialiased" — scale-dependent fallback to nearest/hanning:
//     nearest for integer and high upscales, otherwise hanning.
func parseInterpolationName(name string) (agglib.ImageFilter, bool, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "none", "nearest":
		return agglib.NoFilter, false, true
	case "auto", "antialiased":
		return agglib.FilterHanning, true, true
	case "bilinear":
		return agglib.Bilinear, false, true
	case "bicubic":
		return agglib.Bicubic, false, true
	case "hanning":
		return agglib.Hanning, false, true
	case "hamming":
		return agglib.FilterHamming, false, true
	case "hermite":
		return agglib.Hermite, false, true
	case "kaiser":
		// Kaiser is not a public first-class filter in this agg_go version.
		// Use the closest available pre-defined family member.
		return agglib.Blackman, false, true
	case "quadric":
		return agglib.Quadric, false, true
	case "catrom":
		return agglib.Catrom, false, true
	case "gaussian":
		return agglib.FilterGaussian, false, true
	case "bessel":
		return agglib.FilterBessel, false, true
	case "mitchell":
		return agglib.FilterMitchell, false, true
	case "sinc":
		return agglib.FilterSinc, false, true
	case "lanczos":
		return agglib.FilterLanczos, false, true
	case "spline16":
		return agglib.Spline16, false, true
	case "spline36":
		return agglib.Spline36, false, true
	case "blackman":
		return agglib.Blackman, false, true
	default:
		return agglib.NoFilter, false, false
	}
}

// resolveInterpolationName resolves auto/antialiased into a concrete filter by
// applying Matplotlib's scale-based fallback behavior. It returns false when the
// interpolation string was unknown.
func resolveInterpolationName(name string, srcW, srcH, dstW, dstH float64) (agglib.ImageFilter, bool) {
	filter, adaptive, ok := parseInterpolationName(name)
	if !ok {
		return agglib.NoFilter, false
	}
	if !adaptive {
		return filter, true
	}
	if shouldUseNearestForAutoResample(srcW, srcH, dstW, dstH) {
		return agglib.NoFilter, true
	}
	return filter, true
}

// shouldUseNearestForAutoResample defers to the shared policy in imageinterp so
// the AGG and SVG backends cannot disagree about when an "antialiased" image is
// drawn unfiltered.
func shouldUseNearestForAutoResample(srcW, srcH, dstW, dstH float64) bool {
	if srcW <= 0 || srcH <= 0 || dstW <= 0 || dstH <= 0 {
		return false
	}

	return imageinterp.Resolve("antialiased", srcW, srcH, dstW, dstH) == imageinterp.Nearest
}

func floatAlmostEqual(a, b float64) bool {
	const eps = 1e-9

	return math.Abs(a-b) <= eps*math.Max(1.0, math.Max(math.Abs(a), math.Abs(b)))
}

// resampleForFilter selects an appropriate ImageResample mode for a given
// filter. NoFilter implies NoResample (nearest-neighbour); every other filter
// implies ResampleAlways so the filter actually fires on transformed images.
func resampleForFilter(filter agglib.ImageFilter) agglib.ImageResample {
	if filter == agglib.NoFilter {
		return agglib.NoResample
	}
	return agglib.ResampleAlways
}

// applyInterpolation configures the active AGG surface's filter and resample
// state from img.Interpolation(). Empty / unrecognised names fall back to
// NoFilter (nearest-neighbour). Caller is responsible for save/restore via
// GetImageFilter/GetImageResample around the call.
func applyInterpolation(s *aggSurface, img render.Image, dstW, dstH float64) {
	filter := agglib.NoFilter
	srcW, srcH := float64(0), float64(0)
	interpolation := ""
	if img != nil {
		if w, h := img.Size(); w > 0 && h > 0 {
			srcW = float64(w)
			srcH = float64(h)
		}
		interpolation = img.Interpolation()
	}

	if f, ok := resolveInterpolationName(interpolation, srcW, srcH, dstW, dstH); ok {
		filter = f
	}
	s.SetImageFilter(filter)
	s.SetImageResample(resampleForFilter(filter))
}

package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/style"
)

// lineStyleToDashes maps a Matplotlib line-style spec to a renderer dash
// pattern. The returned values are already scaled by lineWidth, matching
// Matplotlib's scale_dashes (lines.py), because LineCollection.Draw forwards
// dash patterns to the backend unscaled. A nil result means a solid stroke.
//
// Base patterns mirror the Matplotlib rcParams defaults:
//
//	dashed:  3.7, 1.6
//	dashdot: 6.4, 1.6, 1, 1.6
//	dotted:  1, 1.65
//
// Shared by the contour styling path and LineCollection's string linestyle
// support.
func lineStyleToDashes(spec string, lineWidth float64) []float64 {
	return lineStyleToDashesRC(spec, lineWidth, lineWidth, &style.Default.Lines)
}

func lineStyleToDashesRC(spec string, lineWidth, pointPx float64, rc *style.LinesRC) []float64 {
	if rc == nil {
		rc = &style.Default.Lines
	}
	var base []float64
	switch strings.ToLower(strings.TrimSpace(spec)) {
	case "", "-", "solid", "none":
		return nil
	case "--", "dashed":
		base = rc.DashedPattern
	case "-.", "dashdot":
		base = rc.DashDotPattern
	case ":", "dotted":
		base = rc.DottedPattern
	default:
		return nil
	}
	scale := pointPx
	if rc.ScaleDashes {
		scale = lineWidth
	}
	if scale <= 0 {
		scale = 1
	}
	out := make([]float64, len(base))
	for i, v := range base {
		out[i] = v * scale
	}
	return out
}

// contourMonochrome reports whether a contour is drawn in a single color, which
// is the condition under which Matplotlib applies negative_linestyles to
// negative levels (ContourSet._process_colors sets monochrome from
// cmap.monochrome; a single explicit color produces a monochrome ListedColormap,
// while a multi-color colormap such as the viridis default does not).
func contourMonochrome(opt ContourOptions) bool {
	return opt.Color.IsSet() || len(opt.Colors) == 1
}

// resolveContourLineStyles returns one line-style name per level, porting
// Matplotlib's ContourSet._process_linestyles (contour.py).
//
//   - An explicit LineStyles list is cycled (ceil-padded) and truncated to the
//     level count; a single-element list applies that style to every level.
//   - With no LineStyles and a monochrome contour, negative levels (below a
//     tiny negative epsilon) receive NegativeLineStyles (default "dashed");
//     all other levels are "solid".
//   - Otherwise every level is "solid".
func resolveContourLineStyles(levels []float64, opt ContourOptions, monochrome bool) []string {
	n := len(levels)
	if n == 0 {
		return nil
	}
	styles := make([]string, n)

	if len(opt.LineStyles) > 0 {
		src := opt.LineStyles
		for i := range styles {
			styles[i] = src[i%len(src)]
		}
		return styles
	}

	for i := range styles {
		styles[i] = "solid"
	}
	if !monochrome {
		return styles
	}

	negative := "dashed"
	if style := opt.NegativeLineStyles.OrZero(); style != "" {
		negative = style
	}
	zmin, zmax := levels[0], levels[0]
	for _, lev := range levels {
		zmin = math.Min(zmin, lev)
		zmax = math.Max(zmax, lev)
	}
	eps := -(zmax - zmin) * 1e-15
	for i, lev := range levels {
		if lev < eps {
			styles[i] = negative
		}
	}
	return styles
}

// contourLineDashPatterns builds the per-polyline dash patterns for a contour
// line collection from the resolved per-level styles.
func contourLineDashPatterns(
	polylineLevels, levels []float64,
	styles []string,
	lineWidth float64,
	linesRC *style.LinesRC,
) [][]float64 {
	if len(styles) == 0 {
		return nil
	}
	dashes := make([][]float64, len(polylineLevels))
	any := false
	for i, level := range polylineLevels {
		li := contourLevelIndex(levels, level)
		style := "solid"
		if li >= 0 && li < len(styles) {
			style = styles[li]
		}
		dashes[i] = lineStyleToDashesRC(style, lineWidth, 1, linesRC)
		if dashes[i] != nil {
			any = true
		}
	}
	if !any {
		return nil
	}
	return dashes
}

// contourLevelIndex returns the index of level within levels (tolerant match),
// or -1 when absent.
func contourLevelIndex(levels []float64, level float64) int {
	for i, existing := range levels {
		if math.Abs(existing-level) <= 1e-12 {
			return i
		}
	}
	return -1
}

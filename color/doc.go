// Package color provides colormaps, color cycles, and color utilities used to
// map scalar data and series indices onto rendered colors.
//
// A [Colormap] maps a normalized position in [0,1] to a [render.Color] via
// [Colormap.At], and maps an arbitrary data value through a normalization via
// [Colormap.AtValue]. The Matplotlib-compatible colormaps (viridis, plasma,
// coolwarm, and others) are available by name through [GetColormap];
// [DefaultColormap] returns viridis. Custom colormaps are built with
// [NewColormap] and made available globally with [RegisterColormap].
// [Colormap.Reversed] produces the reversed variant, and [Colormap.WithBad],
// [Colormap.WithUnder], and [Colormap.WithOver] set the colors used for
// masked, below-range, and above-range values.
//
// A [ColorCycle] supplies the sequence of colors assigned to successive plot
// series; [NewDefaultColorCycle] returns the Matplotlib tab10 cycle.
package color

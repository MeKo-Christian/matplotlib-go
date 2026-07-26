package pyplot

import "github.com/cwbudde/matplotlib-go/core"

// Contour delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Contour(data [][]float64, opt core.ContourOptions) *core.ContourSet {
	return GCA().Contour(data, opt)
}

// Clabel delegates contour labeling to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Clabel(cs *core.ContourSet, opt core.ClabelOptions) []core.ContourLabel {
	return GCA().Clabel(cs, opt)
}

// Contourf delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Contourf(data [][]float64, opt core.ContourOptions) *core.ContourSet {
	return GCA().Contourf(data, opt)
}

// TriPlot delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func TriPlot(tri core.Triangulation, opt core.TriPlotOptions) *core.LineCollection {
	return GCA().TriPlot(tri, opt)
}

// TriColor delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func TriColor(tri core.Triangulation, values []float64, opt core.TriColorOptions) *core.PolyCollection {
	return GCA().TriColor(tri, values, opt)
}

// TriContour delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func TriContour(tri core.Triangulation, values []float64, opt core.ContourOptions) *core.ContourSet {
	return GCA().TriContour(tri, values, opt)
}

// TriContourf delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func TriContourf(tri core.Triangulation, values []float64, opt core.ContourOptions) *core.ContourSet {
	return GCA().TriContourf(tri, values, opt)
}

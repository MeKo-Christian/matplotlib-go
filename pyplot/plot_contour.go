package pyplot

import "github.com/cwbudde/matplotlib-go/core"

// Contour delegates to the current axes.
func Contour(data [][]float64, opts ...core.ContourOptions) *core.ContourSet {
	return GCA().Contour(data, opts...)
}

// Clabel delegates contour labeling to the current axes.
func Clabel(cs *core.ContourSet, opts ...core.ClabelOptions) []core.ContourLabel {
	return GCA().Clabel(cs, opts...)
}

// Contourf delegates to the current axes.
func Contourf(data [][]float64, opts ...core.ContourOptions) *core.ContourSet {
	return GCA().Contourf(data, opts...)
}

// TriPlot delegates to the current axes.
func TriPlot(tri core.Triangulation, opts ...core.TriPlotOptions) *core.LineCollection {
	return GCA().TriPlot(tri, opts...)
}

// TriColor delegates to the current axes.
func TriColor(tri core.Triangulation, values []float64, opts ...core.TriColorOptions) *core.PolyCollection {
	return GCA().TriColor(tri, values, opts...)
}

// TriContour delegates to the current axes.
func TriContour(tri core.Triangulation, values []float64, opts ...core.ContourOptions) *core.ContourSet {
	return GCA().TriContour(tri, values, opts...)
}

// TriContourf delegates to the current axes.
func TriContourf(tri core.Triangulation, values []float64, opts ...core.ContourOptions) *core.ContourSet {
	return GCA().TriContourf(tri, values, opts...)
}

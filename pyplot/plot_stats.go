package pyplot

import (
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

// Bar delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Bar(x, heights any, opt core.BarOptions) (*core.Bar2D, error) {
	return GCA().Bar(x, heights, opt)
}

// BarH delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func BarH(y, widths any, opt core.BarOptions) (*core.Bar2D, error) {
	return GCA().BarH(y, widths, opt)
}

// BrokenBarH delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func BrokenBarH(xRanges [][2]float64, yRange [2]float64, opt core.BarOptions) *core.Bar2D {
	return GCA().BrokenBarH(xRanges, yRange, opt)
}

// BarLabel delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func BarLabel(bar *core.Bar2D, labels []string, opt core.BarLabelOptions) []*core.Text {
	return GCA().BarLabel(bar, labels, opt)
}

// FillBetween delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func FillBetween(x, y1, y2 any, opt core.FillOptions) (*core.Fill2D, error) {
	return GCA().FillBetween(x, y1, y2, opt)
}

// FillBetweenX delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func FillBetweenX(y, x1, x2 []float64, opt core.FillOptions) (*core.Fill2D, error) {
	return GCA().FillBetweenX(y, x1, x2, opt)
}

// Fill delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Fill(x, y []float64, opt core.FillOptions) *core.PolyCollection {
	return GCA().Fill(x, y, opt)
}

// Hist delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Hist(data []float64, opt core.HistOptions) (*core.Hist2D, error) {
	return GCA().Hist(data, opt)
}

// BoxPlot delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func BoxPlot(data []float64, opt core.BoxPlotOptions) *core.BoxPlot2D {
	return GCA().BoxPlot(data, opt)
}

// Bxp delegates precomputed boxplot statistics to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Bxp(stats []core.BxpStat, opt core.BxpOptions) *core.BxpContainer {
	return GCA().Bxp(stats, opt)
}

// StackPlot delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func StackPlot(x []float64, ys [][]float64, opt core.StackPlotOptions) []*core.Fill2D {
	return GCA().StackPlot(x, ys, opt)
}

// ECDF delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func ECDF(data []float64, opt core.ECDFOptions) *core.Line2D {
	return GCA().ECDF(data, opt)
}

// ErrorBar delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func ErrorBar(x, y, xErr, yErr []float64, opt core.ErrorBarOptions) (*core.ErrorBar, error) {
	return GCA().ErrorBar(x, y, xErr, yErr, opt)
}

// Arrow adds a FancyArrow artist to the current axes.
//
//nolint:gocritic // The arrow prototype is an immutable snapshot of the caller's options.
func Arrow(x, y, dx, dy float64, arrow core.Arrow) *core.Arrow {
	arrow.XY = geom.Pt{X: x, Y: y}
	arrow.DX = dx
	arrow.DY = dy
	if arrow.Coords == (core.CoordinateSpec{}) {
		arrow.Coords = core.Coords(core.CoordData)
	}
	GCA().Add(&arrow)
	return &arrow
}

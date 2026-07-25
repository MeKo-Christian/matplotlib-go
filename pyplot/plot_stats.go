package pyplot

import (
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

// Bar delegates to the current axes.
func Bar(x, heights any, opts ...core.BarOptions) (*core.Bar2D, error) {
	return GCA().Bar(x, heights, opts...)
}

// BarH delegates to the current axes.
func BarH(y, widths any, opts ...core.BarOptions) (*core.Bar2D, error) {
	return GCA().BarH(y, widths, opts...)
}

// BrokenBarH delegates to the current axes.
func BrokenBarH(xRanges [][2]float64, yRange [2]float64, opts ...core.BarOptions) *core.Bar2D {
	return GCA().BrokenBarH(xRanges, yRange, opts...)
}

// BarLabel delegates to the current axes.
func BarLabel(bar *core.Bar2D, labels []string, opts ...core.BarLabelOptions) []*core.Text {
	return GCA().BarLabel(bar, labels, opts...)
}

// FillBetween delegates to the current axes.
func FillBetween(x, y1, y2 any, opts ...core.FillOptions) (*core.Fill2D, error) {
	return GCA().FillBetween(x, y1, y2, opts...)
}

// FillBetweenX delegates to the current axes.
func FillBetweenX(y, x1, x2 []float64, opts ...core.FillOptions) (*core.Fill2D, error) {
	return GCA().FillBetweenX(y, x1, x2, opts...)
}

// Fill delegates to the current axes.
func Fill(x, y []float64, opts ...core.FillOptions) *core.PolyCollection {
	return GCA().Fill(x, y, opts...)
}

// Hist delegates to the current axes.
func Hist(data []float64, opts ...core.HistOptions) (*core.Hist2D, error) {
	return GCA().Hist(data, opts...)
}

// BoxPlot delegates to the current axes.
func BoxPlot(data []float64, opts ...core.BoxPlotOptions) *core.BoxPlot2D {
	return GCA().BoxPlot(data, opts...)
}

// Bxp delegates precomputed boxplot statistics to the current axes.
func Bxp(stats []core.BxpStat, opts ...core.BxpOptions) *core.BxpContainer {
	return GCA().Bxp(stats, opts...)
}

// StackPlot delegates to the current axes.
func StackPlot(x []float64, ys [][]float64, opts ...core.StackPlotOptions) []*core.Fill2D {
	return GCA().StackPlot(x, ys, opts...)
}

// ECDF delegates to the current axes.
func ECDF(data []float64, opts ...core.ECDFOptions) *core.Line2D {
	return GCA().ECDF(data, opts...)
}

// ErrorBar delegates to the current axes.
func ErrorBar(x, y, xErr, yErr []float64, opts ...core.ErrorBarOptions) (*core.ErrorBar, error) {
	return GCA().ErrorBar(x, y, xErr, yErr, opts...)
}

// Arrow adds a FancyArrow artist to the current axes.
func Arrow(x, y, dx, dy float64, opts ...core.Arrow) *core.Arrow {
	arrow := core.Arrow{
		XY:     geom.Pt{X: x, Y: y},
		DX:     dx,
		DY:     dy,
		Coords: core.Coords(core.CoordData),
	}
	if len(opts) > 0 {
		arrow = opts[0]
		arrow.XY = geom.Pt{X: x, Y: y}
		arrow.DX = dx
		arrow.DY = dy
		if arrow.Coords == (core.CoordinateSpec{}) {
			arrow.Coords = core.Coords(core.CoordData)
		}
	}
	GCA().Add(&arrow)
	return &arrow
}

package pyplot

import (
	"time"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

// Plot delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Plot(x, y any, opt core.PlotOptions) (*core.Line2D, error) {
	return GCA().Plot(x, y, opt)
}

// PlotDate delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func PlotDate(x []time.Time, y []float64, opt core.PlotOptions) (*core.Line2D, error) {
	return GCA().PlotDate(x, y, opt)
}

// SemilogX delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func SemilogX(x, y []float64, opt core.PlotOptions) *core.Line2D {
	return GCA().SemilogX(x, y, opt)
}

// SemilogY delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func SemilogY(x, y []float64, opt core.PlotOptions) *core.Line2D {
	return GCA().SemilogY(x, y, opt)
}

// LogLog delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func LogLog(x, y []float64, opt core.PlotOptions) *core.Line2D {
	return GCA().LogLog(x, y, opt)
}

// Step delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Step(x, y []float64, opt core.StepOptions) *core.Line2D {
	return GCA().Step(x, y, opt)
}

// Stairs delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Stairs(values, edges []float64, opt core.StairsOptions) *core.Stairs2D {
	return GCA().Stairs(values, edges, opt)
}

// Scatter delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Scatter(x, y any, opt core.ScatterOptions) (*core.Scatter2D, error) {
	return GCA().Scatter(x, y, opt)
}

// Plot3D delegates to the current 3D axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Plot3D(x, y, z []float64, opt core.PlotOptions) *core.Line2D {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Plot3D(x, y, z, opt)
}

// Scatter3D delegates to the current 3D axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Scatter3D(x, y, z []float64, opt core.ScatterOptions) *core.Scatter2D {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Scatter3D(x, y, z, opt)
}

// Wireframe delegates to the current 3D axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Wireframe(x, y []float64, z [][]float64, opt core.PlotOptions) *core.LineCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Wireframe(x, y, z, opt)
}

// Surface delegates to the current 3D axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Surface(x, y []float64, z [][]float64, opt core.PlotOptions) *core.PolyCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Surface(x, y, z, opt)
}

// Voxel delegates to the current 3D axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Voxel(x, y, z, dx, dy, dz []float64, opt core.PlotOptions) *core.LineCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Voxel(x, y, z, dx, dy, dz, opt)
}

// Trisurf delegates to the current 3D axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Trisurf(tri core.Triangulation, z []float64, opt core.PlotOptions) *core.PolyCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Trisurf(tri, z, opt)
}

// Contour3D delegates to the current 3D axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Contour3D(x, y []float64, z [][]float64, opt core.PlotOptions) *core.LineCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Contour(x, y, z, opt)
}

// Contourf3D delegates to the current 3D axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Contourf3D(x, y []float64, z [][]float64, opt core.PlotOptions) *core.PolyCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Contourf(x, y, z, opt)
}

// Text3D delegates to the current 3D axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Text3D(x, y, z float64, text string, opt core.TextOptions) *core.Text {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Text3D(x, y, z, text, opt)
}

// Text delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Text(x, y float64, text string, opt core.TextOptions) *core.Text {
	return GCA().Text(x, y, text, opt)
}

// Annotate delegates to the current axes.
//
//nolint:gocritic // AnnotationOptions is forwarded unchanged to the axes method.
func Annotate(text string, x, y float64, opt core.AnnotationOptions) *core.Annotation {
	return GCA().Annotate(text, x, y, opt)
}

// AxHLine delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func AxHLine(y float64, opt core.HLineOptions) *core.Segment2D {
	return GCA().AxHLine(y, opt)
}

// AxVLine delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func AxVLine(x float64, opt core.VLineOptions) *core.Segment2D {
	return GCA().AxVLine(x, opt)
}

// AxLine delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func AxLine(p1, p2 geom.Pt, opt core.ReferenceLineOptions) *core.InfiniteLine2D {
	return GCA().AxLine(p1, p2, opt)
}

// AxLineSlope delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func AxLineSlope(point geom.Pt, slope float64, opt core.ReferenceLineOptions) *core.InfiniteLine2D {
	return GCA().AxLineSlope(point, slope, opt)
}

// AxHSpan delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func AxHSpan(yMin, yMax float64, opt core.HSpanOptions) *core.Span2D {
	return GCA().AxHSpan(yMin, yMax, opt)
}

// AxVSpan delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func AxVSpan(xMin, xMax float64, opt core.VSpanOptions) *core.Span2D {
	return GCA().AxVSpan(xMin, xMax, opt)
}

// HLines adds horizontal line segments to the current axes.
//
//nolint:gocritic // LineCollectionOptions is forwarded unchanged to the axes method.
func HLines(y, xMin, xMax []float64, opt core.LineCollectionOptions) *core.LineCollection {
	return GCA().HLines(y, xMin, xMax, opt)
}

// VLines adds vertical line segments to the current axes.
//
//nolint:gocritic // LineCollectionOptions is forwarded unchanged to the axes method.
func VLines(x, yMin, yMax []float64, opt core.LineCollectionOptions) *core.LineCollection {
	return GCA().VLines(x, yMin, yMax, opt)
}

// Stem delegates to the current axes.
//
//nolint:gocritic // StemOptions is forwarded unchanged to the axes method.
func Stem(x, y []float64, opt core.StemOptions) *core.StemContainer {
	return GCA().Stem(x, y, opt)
}

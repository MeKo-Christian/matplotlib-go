package pyplot

import (
	"time"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

// Plot delegates to the current axes.
func Plot(x, y any, opts ...core.PlotOptions) (*core.Line2D, error) {
	return GCA().Plot(x, y, opts...)
}

// PlotDate delegates to the current axes.
func PlotDate(x []time.Time, y []float64, opts ...core.PlotOptions) (*core.Line2D, error) {
	return GCA().PlotDate(x, y, opts...)
}

// SemilogX delegates to the current axes.
func SemilogX(x, y []float64, opts ...core.PlotOptions) *core.Line2D {
	return GCA().SemilogX(x, y, opts...)
}

// SemilogY delegates to the current axes.
func SemilogY(x, y []float64, opts ...core.PlotOptions) *core.Line2D {
	return GCA().SemilogY(x, y, opts...)
}

// LogLog delegates to the current axes.
func LogLog(x, y []float64, opts ...core.PlotOptions) *core.Line2D {
	return GCA().LogLog(x, y, opts...)
}

// Step delegates to the current axes.
func Step(x, y []float64, opts ...core.StepOptions) *core.Line2D {
	return GCA().Step(x, y, opts...)
}

// Stairs delegates to the current axes.
func Stairs(values, edges []float64, opts ...core.StairsOptions) *core.Stairs2D {
	return GCA().Stairs(values, edges, opts...)
}

// Scatter delegates to the current axes.
func Scatter(x, y any, opts ...core.ScatterOptions) (*core.Scatter2D, error) {
	return GCA().Scatter(x, y, opts...)
}

// Plot3D delegates to the current 3D axes.
func Plot3D(x, y, z []float64, opts ...core.PlotOptions) *core.Line2D {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Plot3D(x, y, z, opts...)
}

// Scatter3D delegates to the current 3D axes.
func Scatter3D(x, y, z []float64, opts ...core.ScatterOptions) *core.Scatter2D {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Scatter3D(x, y, z, opts...)
}

// Wireframe delegates to the current 3D axes.
func Wireframe(x, y []float64, z [][]float64, opts ...core.PlotOptions) *core.LineCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Wireframe(x, y, z, opts...)
}

// Surface delegates to the current 3D axes.
func Surface(x, y []float64, z [][]float64, opts ...core.PlotOptions) *core.PolyCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Surface(x, y, z, opts...)
}

// Voxel delegates to the current 3D axes.
func Voxel(x, y, z, dx, dy, dz []float64, opts ...core.PlotOptions) *core.LineCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Voxel(x, y, z, dx, dy, dz, opts...)
}

// Trisurf delegates to the current 3D axes.
func Trisurf(tri core.Triangulation, z []float64, opts ...core.PlotOptions) *core.PolyCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Trisurf(tri, z, opts...)
}

// Contour3D delegates to the current 3D axes.
func Contour3D(x, y []float64, z [][]float64, opts ...core.PlotOptions) *core.LineCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Contour(x, y, z, opts...)
}

// Contourf3D delegates to the current 3D axes.
func Contourf3D(x, y []float64, z [][]float64, opts ...core.PlotOptions) *core.PolyCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Contourf(x, y, z, opts...)
}

// Text3D delegates to the current 3D axes.
func Text3D(x, y, z float64, text string, opts ...core.TextOptions) *core.Text {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Text3D(x, y, z, text, opts...)
}

// Text delegates to the current axes.
func Text(x, y float64, text string, opts ...core.TextOptions) *core.Text {
	return GCA().Text(x, y, text, opts...)
}

// Annotate delegates to the current axes.
func Annotate(text string, x, y float64, opts ...core.AnnotationOptions) *core.Annotation {
	return GCA().Annotate(text, x, y, opts...)
}

// AxHLine delegates to the current axes.
func AxHLine(y float64, opts ...core.HLineOptions) *core.Segment2D {
	return GCA().AxHLine(y, opts...)
}

// AxVLine delegates to the current axes.
func AxVLine(x float64, opts ...core.VLineOptions) *core.Segment2D {
	return GCA().AxVLine(x, opts...)
}

// AxLine delegates to the current axes.
func AxLine(p1, p2 geom.Pt, opts ...core.ReferenceLineOptions) *core.InfiniteLine2D {
	return GCA().AxLine(p1, p2, opts...)
}

// AxLineSlope delegates to the current axes.
func AxLineSlope(point geom.Pt, slope float64, opts ...core.ReferenceLineOptions) *core.InfiniteLine2D {
	return GCA().AxLineSlope(point, slope, opts...)
}

// AxHSpan delegates to the current axes.
func AxHSpan(yMin, yMax float64, opts ...core.HSpanOptions) *core.Span2D {
	return GCA().AxHSpan(yMin, yMax, opts...)
}

// AxVSpan delegates to the current axes.
func AxVSpan(xMin, xMax float64, opts ...core.VSpanOptions) *core.Span2D {
	return GCA().AxVSpan(xMin, xMax, opts...)
}

// HLines adds horizontal line segments to the current axes.
func HLines(y, xMin, xMax []float64, opts ...core.LineCollection) *core.LineCollection {
	return GCA().HLines(y, xMin, xMax, opts...)
}

// VLines adds vertical line segments to the current axes.
func VLines(x, yMin, yMax []float64, opts ...core.LineCollection) *core.LineCollection {
	return GCA().VLines(x, yMin, yMax, opts...)
}

// Stem delegates to the current axes.
func Stem(x, y []float64, opts ...core.StemOptions) *core.StemContainer {
	return GCA().Stem(x, y, opts...)
}

func addLineCollection(segments [][]geom.Pt, opts ...core.LineCollection) *core.LineCollection {
	ax := GCA()
	collection := core.LineCollection{
		Collection: core.Collection{
			Coords: core.Coords(core.CoordData),
			Alpha:  1,
		},
		Segments:  segments,
		Color:     ax.NextColor(),
		LineWidth: 1,
	}
	if len(opts) > 0 {
		collection = opts[0]
		collection.Segments = segments
		if collection.Coords == (core.CoordinateSpec{}) {
			collection.Coords = core.Coords(core.CoordData)
		}
		if collection.Alpha == 0 {
			collection.Alpha = 1
		}
		if collection.Color.A == 0 && len(collection.Colors) == 0 {
			collection.Color = ax.NextColor()
		}
		if collection.LineWidth == 0 && len(collection.LineWidths) == 0 {
			collection.LineWidth = 1
		}
	}
	ax.Add(&collection)
	return &collection
}

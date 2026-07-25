package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/tri"
)

// Triangulation stores an unstructured triangular mesh. It is an alias for
// [tri.Triangulation]; the reusable triangulation toolkit (Delaunay,
// point-location, interpolation, refinement, analysis) lives in the tri
// package.
type Triangulation = tri.Triangulation

// NewTriangulation returns a Delaunay triangulation for the supplied points.
func NewTriangulation(x, y []float64) (Triangulation, error) {
	return tri.New(x, y)
}

// TriPlotOptions configures triplot rendering.
type TriPlotOptions struct {
	Color     *render.Color
	LineWidth *float64
	Alpha     *float64
	Label     string
}

// TriColorOptions configures tripcolor rendering.
type TriColorOptions struct {
	Colormap  *string
	Norm      ScalarNormalizer
	VMin      *float64
	VMax      *float64
	Alpha     *float64
	EdgeColor *render.Color
	EdgeWidth *float64
	Label     string
}

// TriPlot draws the unique edges of the supplied triangulation.
func (a *Axes) TriPlot(tri Triangulation, opts ...TriPlotOptions) *LineCollection {
	if err := tri.Validate(); err != nil || len(tri.Triangles) == 0 {
		return nil
	}

	var opt TriPlotOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	alpha := meshAlpha(opt.Alpha)
	color = color.WithAlphaMultiplier(alpha)

	width := 1.0
	if opt.LineWidth != nil {
		width = *opt.LineWidth
	}

	edges := tri.Edges()
	segments := make([][]geom.Pt, 0, len(edges))
	for _, edge := range edges {
		segments = append(segments, []geom.Pt{tri.Point(edge[0]), tri.Point(edge[1])})
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  opt.Label,
			Alpha:  1,
		},
		Segments:  segments,
		Color:     color,
		LineWidth: width,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	return collection
}

// TriColor draws per-triangle colored polygons over a triangulation. Values
// may be provided per triangle or per point; point values are averaged onto
// each triangle.
func (a *Axes) TriColor(tri Triangulation, values []float64, opts ...TriColorOptions) *PolyCollection {
	if err := tri.Validate(); err != nil || len(tri.Triangles) == 0 {
		return nil
	}

	var opt TriColorOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	triangleValues, ok := triColorValues(tri, values)
	if !ok {
		return nil
	}

	cmap := ""
	if opt.Colormap != nil {
		cmap = *opt.Colormap
	}
	mapping, err := ResolveScalarMapValues(triangleValues, ScalarMapConfig{
		Colormap: cmap,
		Norm:     opt.Norm,
		VMin:     opt.VMin,
		VMax:     opt.VMax,
	})
	if err != nil {
		return nil
	}
	alpha := meshAlpha(opt.Alpha)

	edgeColor := render.Color{}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}
	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	polygons := make([][]geom.Pt, 0, len(tri.Triangles))
	faceColors := make([]render.Color, 0, len(tri.Triangles))
	edgeColors := make([]render.Color, 0, len(tri.Triangles))
	for triIdx, triangle := range tri.Triangles {
		if tri.Masked(triIdx) {
			continue
		}
		polygons = append(polygons, []geom.Pt{
			tri.Point(triangle[0]),
			tri.Point(triangle[1]),
			tri.Point(triangle[2]),
		})
		value := triangleValues[triIdx]
		if !isFinite(value) {
			faceColors = append(faceColors, render.Color{})
			edgeColors = append(edgeColors, render.Color{})
			continue
		}
		faceColors = append(faceColors, mapping.Color(value, alpha))
		edgeColors = append(edgeColors, edgeColor)
	}

	collection := &PolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords:   Coords(CoordData),
				Label:    opt.Label,
				Alpha:    1,
				Colormap: mapping.Colormap,
				Norm:     mapping.Norm,
				VMin:     mapping.VMin,
				VMax:     mapping.VMax,
			},
			FaceColors: faceColors,
			EdgeColors: edgeColors,
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
		Polygons: polygons,
	}
	a.Add(collection)
	return collection
}

func triColorValues(tri Triangulation, values []float64) ([]float64, bool) {
	switch len(values) {
	case len(tri.Triangles):
		out := make([]float64, len(values))
		copy(out, values)
		return out, true
	case len(tri.X):
		out := make([]float64, len(tri.Triangles))
		for i, triangle := range tri.Triangles {
			out[i] = meshValueAverage([]float64{
				values[triangle[0]],
				values[triangle[1]],
				values[triangle[2]],
			})
		}
		return out, true
	default:
		return nil, false
	}
}

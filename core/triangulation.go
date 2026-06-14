package core

import (
	"fmt"
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Triangulation stores an unstructured triangular mesh.
type Triangulation struct {
	X         []float64
	Y         []float64
	Triangles [][3]int
	Mask      []bool
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

// Validate verifies that the triangulation references valid point indices.
func (t Triangulation) Validate() error {
	if len(t.X) == 0 || len(t.Y) == 0 {
		return fmt.Errorf("triangulation requires coordinates")
	}
	if len(t.X) != len(t.Y) {
		return fmt.Errorf("triangulation X/Y lengths differ")
	}
	for triIdx, tri := range t.Triangles {
		for _, idx := range tri {
			if idx < 0 || idx >= len(t.X) {
				return fmt.Errorf("triangle %d references point %d outside 0..%d", triIdx, idx, len(t.X)-1)
			}
		}
	}
	if len(t.Mask) > 0 && len(t.Mask) != len(t.Triangles) {
		return fmt.Errorf("triangulation mask length %d does not match triangles %d", len(t.Mask), len(t.Triangles))
	}
	return nil
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
	color.A *= alpha

	width := 1.0
	if opt.LineWidth != nil {
		width = *opt.LineWidth
	}

	edgeSet := map[[2]int]struct{}{}
	segments := make([][]geom.Pt, 0, len(tri.Triangles)*3)
	for triIdx, triangle := range tri.Triangles {
		if tri.masked(triIdx) {
			continue
		}
		edges := [][2]int{
			sortedEdge(triangle[0], triangle[1]),
			sortedEdge(triangle[1], triangle[2]),
			sortedEdge(triangle[2], triangle[0]),
		}
		for _, edge := range edges {
			if _, exists := edgeSet[edge]; exists {
				continue
			}
			edgeSet[edge] = struct{}{}
			segments = append(segments, []geom.Pt{tri.point(edge[0]), tri.point(edge[1])})
		}
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
		if tri.masked(triIdx) {
			continue
		}
		polygons = append(polygons, []geom.Pt{
			tri.point(triangle[0]),
			tri.point(triangle[1]),
			tri.point(triangle[2]),
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

func sortedEdge(a, b int) [2]int {
	if a < b {
		return [2]int{a, b}
	}
	return [2]int{b, a}
}

func (t Triangulation) point(idx int) geom.Pt {
	return geom.Pt{X: t.X[idx], Y: t.Y[idx]}
}

func (t Triangulation) masked(triIdx int) bool {
	return len(t.Mask) > 0 && triIdx < len(t.Mask) && t.Mask[triIdx]
}

func autoTriangulate(t Triangulation) (Triangulation, bool) {
	if len(t.Triangles) > 0 {
		return t, true
	}
	if len(t.Mask) > 0 || len(t.X) != len(t.Y) || len(t.X) < 3 {
		return Triangulation{}, false
	}

	triangles, ok := delaunayTriangles(t.X, t.Y)
	if !ok || len(triangles) == 0 {
		return Triangulation{}, false
	}
	t.Triangles = triangles
	return t, true
}

type delaunayTriangle struct {
	a int
	b int
	c int
}

func delaunayTriangles(x, y []float64) ([][3]int, bool) {
	n := len(x)
	minX, maxX := x[0], x[0]
	minY, maxY := y[0], y[0]
	for i := 1; i < n; i++ {
		minX = math.Min(minX, x[i])
		maxX = math.Max(maxX, x[i])
		minY = math.Min(minY, y[i])
		maxY = math.Max(maxY, y[i])
	}
	dx := maxX - minX
	dy := maxY - minY
	delta := math.Max(dx, dy)
	if delta == 0 {
		return nil, false
	}
	midX := (minX + maxX) / 2
	midY := (minY + maxY) / 2

	px := append(append([]float64(nil), x...), midX-20*delta, midX, midX+20*delta)
	py := append(append([]float64(nil), y...), midY-delta, midY+20*delta, midY-delta)
	superA, superB, superC := n, n+1, n+2
	super, ok := orientedDelaunayTriangle(superA, superB, superC, px, py)
	if !ok {
		return nil, false
	}
	triangles := []delaunayTriangle{super}

	for p := 0; p < n; p++ {
		bad := make([]bool, len(triangles))
		boundary := make(map[[2]int]int)
		for i, tri := range triangles {
			if pointInCircumcircle(px[p], py[p], tri, px, py) {
				bad[i] = true
				boundary[sortedEdge(tri.a, tri.b)]++
				boundary[sortedEdge(tri.b, tri.c)]++
				boundary[sortedEdge(tri.c, tri.a)]++
			}
		}

		kept := triangles[:0]
		for i, tri := range triangles {
			if !bad[i] {
				kept = append(kept, tri)
			}
		}
		triangles = kept

		for edge, count := range boundary {
			if count != 1 {
				continue
			}
			tri, ok := orientedDelaunayTriangle(edge[0], edge[1], p, px, py)
			if ok {
				triangles = append(triangles, tri)
			}
		}
	}

	out := make([][3]int, 0, len(triangles))
	for _, tri := range triangles {
		if tri.a >= n || tri.b >= n || tri.c >= n {
			continue
		}
		out = append(out, [3]int{tri.a, tri.b, tri.c})
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		for k := 0; k < 3; k++ {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return false
	})
	return out, len(out) > 0
}

func orientedDelaunayTriangle(a, b, c int, x, y []float64) (delaunayTriangle, bool) {
	area := (x[b]-x[a])*(y[c]-y[a]) - (y[b]-y[a])*(x[c]-x[a])
	if math.Abs(area) < 1e-14 {
		return delaunayTriangle{}, false
	}
	if area < 0 {
		a, b = b, a
	}
	return delaunayTriangle{a: a, b: b, c: c}, true
}

func pointInCircumcircle(px, py float64, tri delaunayTriangle, x, y []float64) bool {
	ax := x[tri.a] - px
	ay := y[tri.a] - py
	bx := x[tri.b] - px
	by := y[tri.b] - py
	cx := x[tri.c] - px
	cy := y[tri.c] - py

	det := (ax*ax+ay*ay)*(bx*cy-cx*by) -
		(bx*bx+by*by)*(ax*cy-cx*ay) +
		(cx*cx+cy*cy)*(ax*by-bx*ay)
	return det > 1e-12
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

// Package tri provides an unstructured triangular-mesh toolkit: a
// [Triangulation] data structure with Delaunay construction, edge/neighbor
// derivation and plane coefficients, point location ([TriFinder]), field
// interpolation ([LinearTriInterpolator], [CubicTriInterpolator]), uniform
// mesh refinement ([UniformTriRefiner]) and mesh analysis ([TriAnalyzer]).
//
// It mirrors the public surface of matplotlib's matplotlib.tri module. One
// deliberate deviation: the Delaunay triangulation is computed with a pure-Go
// Bowyer–Watson algorithm rather than matplotlib's Qhull backend, so the mesh
// connectivity is not guaranteed byte-for-byte identical to matplotlib for
// cocircular inputs. Downstream rendering and interpolation match within
// tolerance.
package tri

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/geom"
)

// Triangulation stores an unstructured triangular mesh. Triangles are wound
// anticlockwise (matching matplotlib); Mask, when present, flags triangles to
// be treated as absent.
type Triangulation struct {
	X         []float64
	Y         []float64
	Triangles [][3]int
	Mask      []bool
}

// New returns a Delaunay triangulation for the supplied points.
func New(x, y []float64) (Triangulation, error) {
	t, ok := Triangulation{
		X: append([]float64(nil), x...),
		Y: append([]float64(nil), y...),
	}.EnsureTriangles()
	if !ok {
		return Triangulation{}, fmt.Errorf("could not triangulate %d point(s)", min(len(x), len(y)))
	}
	return t, nil
}

// EnsureTriangles returns a triangulation guaranteed to have a triangle list:
// if Triangles is already populated it is returned unchanged, otherwise a
// Delaunay triangulation of the points is computed. It reports false when no
// triangulation could be produced (too few points, mismatched coordinates, a
// pre-set mask with no triangles, or degenerate/collinear input).
func (t Triangulation) EnsureTriangles() (Triangulation, bool) {
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

// Point returns the coordinates of point idx.
func (t Triangulation) Point(idx int) geom.Pt {
	return geom.Pt{X: t.X[idx], Y: t.Y[idx]}
}

// Masked reports whether triangle triIdx is masked out.
func (t Triangulation) Masked(triIdx int) bool {
	return len(t.Mask) > 0 && triIdx < len(t.Mask) && t.Mask[triIdx]
}

// Edges returns the unique undirected edges of all non-masked triangles, each
// as an ordered (min,max) index pair.
func (t Triangulation) Edges() [][2]int {
	edgeSet := make(map[[2]int]struct{}, len(t.Triangles)*3)
	edges := make([][2]int, 0, len(t.Triangles)*3)
	for triIdx, triangle := range t.Triangles {
		if t.Masked(triIdx) {
			continue
		}
		for _, e := range [3][2]int{
			sortedEdge(triangle[0], triangle[1]),
			sortedEdge(triangle[1], triangle[2]),
			sortedEdge(triangle[2], triangle[0]),
		} {
			if _, ok := edgeSet[e]; ok {
				continue
			}
			edgeSet[e] = struct{}{}
			edges = append(edges, e)
		}
	}
	return edges
}

// Neighbors returns, for every triangle, the indices of the three neighbouring
// triangles. Following matplotlib, neighbors[t][j] is the triangle adjacent to
// the edge from vertex j to vertex (j+1)%3, or -1 if that edge lies on the mesh
// boundary. Neighbours are computed over all triangles regardless of mask.
//
// This assumes consistent (anticlockwise) winding, as matplotlib does: the
// directed edge (u,v) of one triangle is shared as (v,u) by its neighbour.
func (t Triangulation) Neighbors() [][3]int {
	type loc struct{ tri, edge int }
	directed := make(map[[2]int]loc, len(t.Triangles)*3)
	for triIdx, triangle := range t.Triangles {
		for j := 0; j < 3; j++ {
			directed[[2]int{triangle[j], triangle[(j+1)%3]}] = loc{tri: triIdx, edge: j}
		}
	}
	neighbors := make([][3]int, len(t.Triangles))
	for triIdx, triangle := range t.Triangles {
		neighbors[triIdx] = [3]int{-1, -1, -1}
		for j := 0; j < 3; j++ {
			u, v := triangle[j], triangle[(j+1)%3]
			if l, ok := directed[[2]int{v, u}]; ok {
				neighbors[triIdx][j] = l.tri
			}
		}
	}
	return neighbors
}

// PlaneCoefficients returns, for each triangle, the coefficients (a, b, c) of
// the plane z = a*x + b*y + c fitted through its three vertices' z values.
// Masked or degenerate triangles yield zero coefficients. It mirrors
// matplotlib's Triangulation.calculate_plane_coefficients.
func (t Triangulation) PlaneCoefficients(z []float64) ([][3]float64, error) {
	if len(z) != len(t.X) {
		return nil, fmt.Errorf("plane coefficients require one z per point: got %d, want %d", len(z), len(t.X))
	}
	out := make([][3]float64, len(t.Triangles))
	for triIdx, tri := range t.Triangles {
		if t.Masked(triIdx) {
			continue
		}
		x0, y0, z0 := t.X[tri[0]], t.Y[tri[0]], z[tri[0]]
		x1, y1, z1 := t.X[tri[1]], t.Y[tri[1]], z[tri[1]]
		x2, y2, z2 := t.X[tri[2]], t.Y[tri[2]], z[tri[2]]
		// Solve for (a,b) from the two edge vectors, then c.
		e1x, e1y, e1z := x1-x0, y1-y0, z1-z0
		e2x, e2y, e2z := x2-x0, y2-y0, z2-z0
		det := e1x*e2y - e2x*e1y
		if det == 0 {
			continue
		}
		a := (e1z*e2y - e2z*e1y) / det
		b := (e1x*e2z - e2x*e1z) / det
		c := z0 - a*x0 - b*y0
		out[triIdx] = [3]float64{a, b, c}
	}
	return out, nil
}

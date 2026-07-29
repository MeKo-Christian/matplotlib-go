package core

import (
	"sort"

	"github.com/cwbudde/matplotlib-go/geom"
)

// This file ports Matplotlib's TriContourGenerator (src/tri/_tri.cpp) for
// unfilled triangular-mesh contours. Faithfulness here is not cosmetic: inline
// contour labels are placed at a *vertex index* along each contour line
// (ContourLabeler.locate_label reduces to "the middle vertex" whenever a line
// is shorter than the label width, which is the common case for coarse
// meshes), so a line that traces the same geometry from a different start
// vertex or in the opposite direction gets its label at a different point.
// Reproducing Matplotlib's traversal order, start point and winding is
// therefore what makes label placement match.
//
// The filled case (create_filled_contour/follow_boundary) is deliberately not
// ported; contourBandPolygons still serves tricontourf.

// triEdge identifies the edge of a triangle running from the triangle's local
// vertex edge to vertex (edge+1)%3, mirroring Matplotlib's TriEdge.
type triEdge struct {
	tri  int
	edge int
}

// triContourMesh caches the connectivity Matplotlib's Triangulation derives
// lazily (neighbours and boundaries) so a whole level set is traced from one
// pass over the mesh.
type triContourMesh struct {
	tri       Triangulation
	values    []float64
	neighbors [][3]int
	// boundaries lists each closed boundary loop as an ordered run of
	// triEdges, matching Matplotlib's Triangulation::_boundaries.
	boundaries [][]triEdge
}

// triContourPolylines traces the contour lines of an unstructured mesh, one
// polyline per contour line, in Matplotlib's emission order: for each level,
// the open lines seeded from mesh boundaries first, then the closed interior
// loops in triangle-index order.
//
//nolint:gocritic // Triangulation is a read-only value type throughout the contour API.
func triContourPolylines(tr Triangulation, values, levels []float64) ([][]geom.Pt, []float64) {
	mesh, ok := newTriContourMesh(tr, values)
	if !ok {
		return nil, nil
	}

	var polylines [][]geom.Pt
	var polylineLevels []float64
	for _, level := range levels {
		visited := make([]bool, len(tr.Triangles))
		for _, line := range mesh.boundaryLines(level, visited) {
			if len(line) < 2 {
				continue
			}
			polylines = append(polylines, line)
			polylineLevels = append(polylineLevels, level)
		}
		for _, line := range mesh.interiorLines(level, visited) {
			if len(line) < 2 {
				continue
			}
			polylines = append(polylines, line)
			polylineLevels = append(polylineLevels, level)
		}
	}
	return polylines, polylineLevels
}

//nolint:gocritic // Triangulation is a read-only value type throughout the contour API.
func newTriContourMesh(tr Triangulation, values []float64) (*triContourMesh, bool) {
	if len(tr.Triangles) == 0 || len(values) < len(tr.X) || len(tr.X) != len(tr.Y) {
		return nil, false
	}
	for _, triangle := range tr.Triangles {
		for _, idx := range triangle {
			if idx < 0 || idx >= len(values) {
				return nil, false
			}
		}
	}

	mesh := &triContourMesh{tri: tr, values: values}
	mesh.neighbors = triContourNeighbors(tr)
	mesh.calculateBoundaries()
	return mesh, true
}

// triContourNeighbors mirrors Triangulation::calculate_neighbors: masked
// triangles take no part, so their edges read as mesh boundaries. This differs
// from tri.Triangulation.Neighbors, which deliberately ignores the mask.
//
//nolint:gocritic // Triangulation is a read-only value type throughout the contour API.
func triContourNeighbors(tr Triangulation) [][3]int {
	type loc struct{ tri, edge int }
	directed := make(map[[2]int]loc, len(tr.Triangles)*3)
	for triIdx, triangle := range tr.Triangles {
		if tr.Masked(triIdx) {
			continue
		}
		for j := 0; j < 3; j++ {
			directed[[2]int{triangle[j], triangle[(j+1)%3]}] = loc{tri: triIdx, edge: j}
		}
	}

	neighbors := make([][3]int, len(tr.Triangles))
	for triIdx, triangle := range tr.Triangles {
		neighbors[triIdx] = [3]int{-1, -1, -1}
		if tr.Masked(triIdx) {
			continue
		}
		for j := 0; j < 3; j++ {
			u, v := triangle[j], triangle[(j+1)%3]
			if l, ok := directed[[2]int{v, u}]; ok {
				neighbors[triIdx][j] = l.tri
			}
		}
	}
	return neighbors
}

// calculateBoundaries ports Triangulation::calculate_boundaries. Matplotlib
// seeds each boundary loop from the smallest remaining (tri, edge) pair — the
// begin() of a std::set<TriEdge> — so the loops come out in a deterministic
// order that the contour traversal below depends on.
func (m *triContourMesh) calculateBoundaries() {
	candidates := make([]triEdge, 0, len(m.tri.Triangles))
	remaining := make(map[triEdge]struct{}, len(m.tri.Triangles))
	for triIdx := range m.tri.Triangles {
		if m.tri.Masked(triIdx) {
			continue
		}
		for edge := 0; edge < 3; edge++ {
			if m.neighbors[triIdx][edge] == -1 {
				e := triEdge{tri: triIdx, edge: edge}
				candidates = append(candidates, e)
				remaining[e] = struct{}{}
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].tri != candidates[j].tri {
			return candidates[i].tri < candidates[j].tri
		}
		return candidates[i].edge < candidates[j].edge
	})

	for _, seed := range candidates {
		if _, ok := remaining[seed]; !ok {
			continue
		}
		boundary := []triEdge{}
		current := seed
		for {
			boundary = append(boundary, current)
			delete(remaining, current)

			// Move to the next edge of the current triangle, then walk
			// around its start point until an edge without a neighbour is
			// reached: that is the next boundary edge.
			triIdx := current.tri
			edge := (current.edge + 1) % 3
			point := m.trianglePointIndex(triIdx, edge)
			for m.neighbors[triIdx][edge] != -1 {
				triIdx = m.neighbors[triIdx][edge]
				next, ok := m.edgeInTriangle(triIdx, point)
				if !ok {
					break
				}
				edge = next
			}

			current = triEdge{tri: triIdx, edge: edge}
			if current == boundary[0] {
				break
			}
			if _, ok := remaining[current]; !ok {
				break
			}
		}
		m.boundaries = append(m.boundaries, boundary)
	}
}

// edgeInTriangle returns the local edge of triIdx that starts at point.
func (m *triContourMesh) edgeInTriangle(triIdx, point int) (int, bool) {
	for edge := 0; edge < 3; edge++ {
		if m.tri.Triangles[triIdx][edge] == point {
			return edge, true
		}
	}
	return 0, false
}

func (m *triContourMesh) trianglePointIndex(triIdx, edge int) int {
	return m.tri.Triangles[triIdx][edge%3]
}

func (m *triContourMesh) z(point int) float64 {
	return m.values[point]
}

// boundaryLines ports TriContourGenerator::find_boundary_lines: walk the
// boundaries in order and, wherever the level is crossed downwards, follow the
// line through the interior until it reaches the boundary again.
func (m *triContourMesh) boundaryLines(level float64, visited []bool) [][]geom.Pt {
	var lines [][]geom.Pt
	for _, boundary := range m.boundaries {
		startAbove, endAbove := false, false
		for i, e := range boundary {
			if i == 0 {
				startAbove = m.z(m.trianglePointIndex(e.tri, e.edge)) >= level
			} else {
				startAbove = endAbove
			}
			endAbove = m.z(m.trianglePointIndex(e.tri, e.edge+1)) >= level
			if startAbove && !endAbove {
				cursor := e
				lines = append(lines, m.followInterior(&cursor, true, level, visited))
			}
		}
	}
	return lines
}

// interiorLines ports TriContourGenerator::find_interior_lines: scan triangles
// in index order and seed a closed loop from the first unvisited triangle the
// level passes through, closing the loop by repeating its first point.
func (m *triContourMesh) interiorLines(level float64, visited []bool) [][]geom.Pt {
	var lines [][]geom.Pt
	for triIdx := range m.tri.Triangles {
		if visited[triIdx] || m.tri.Masked(triIdx) {
			continue
		}
		visited[triIdx] = true

		edge, ok := m.exitEdge(triIdx, level)
		if !ok {
			continue
		}
		cursor, ok := m.neighborEdge(triIdx, edge)
		if !ok {
			continue
		}
		line := m.followInterior(&cursor, false, level, visited)
		if len(line) > 0 {
			line = append(line, line[0])
		}
		lines = append(lines, line)
	}
	return lines
}

// followInterior ports TriContourGenerator::follow_interior.
func (m *triContourMesh) followInterior(cursor *triEdge, endOnBoundary bool, level float64, visited []bool) []geom.Pt {
	line := []geom.Pt{m.edgeInterp(cursor.tri, cursor.edge, level)}
	for endOnBoundary || !visited[cursor.tri] {
		edge, ok := m.exitEdge(cursor.tri, level)
		if !ok {
			break
		}
		cursor.edge = edge
		visited[cursor.tri] = true
		line = append(line, m.edgeInterp(cursor.tri, edge, level))

		next, ok := m.neighborEdge(cursor.tri, edge)
		if endOnBoundary && !ok {
			break
		}
		if !ok {
			break
		}
		*cursor = next
	}
	return line
}

// neighborEdge returns the same mesh edge as seen from the adjacent triangle,
// mirroring Triangulation::get_neighbor_edge.
func (m *triContourMesh) neighborEdge(triIdx, edge int) (triEdge, bool) {
	neighbor := m.neighbors[triIdx][edge]
	if neighbor == -1 {
		return triEdge{}, false
	}
	point := m.trianglePointIndex(triIdx, edge+1)
	next, ok := m.edgeInTriangle(neighbor, point)
	if !ok {
		return triEdge{}, false
	}
	return triEdge{tri: neighbor, edge: next}, true
}

// exitEdge ports TriContourGenerator::get_exit_edge for the unfilled case
// (on_upper is always false without filled bands).
func (m *triContourMesh) exitEdge(triIdx int, level float64) (int, bool) {
	config := 0
	for j := 0; j < 3; j++ {
		if m.z(m.trianglePointIndex(triIdx, j)) >= level {
			config |= 1 << j
		}
	}
	switch config {
	case 1, 3:
		return 2, true
	case 2, 6:
		return 0, true
	case 4, 5:
		return 1, true
	default: // 0 and 7: the contour does not pass through this triangle.
		return -1, false
	}
}

// edgeInterp ports TriContourGenerator::edge_interp/interp. The fraction is
// written exactly as Matplotlib writes it; the algebraically equivalent
// (level-z1)/(z2-z1) form rounds differently.
func (m *triContourMesh) edgeInterp(triIdx, edge int, level float64) geom.Pt {
	p1 := m.trianglePointIndex(triIdx, edge)
	p2 := m.trianglePointIndex(triIdx, edge+1)
	z1, z2 := m.z(p1), m.z(p2)
	fraction := (z2 - level) / (z2 - z1)
	a, b := m.tri.Point(p1), m.tri.Point(p2)
	return geom.Pt{
		X: a.X*fraction + b.X*(1-fraction),
		Y: a.Y*fraction + b.Y*(1-fraction),
	}
}

// Package qhull computes 2-D Delaunay triangulations, aiming to reproduce the
// connectivity that matplotlib's Qhull backend (qhull 8.0.2, options
// "d Qt Qbb Qc Qz") produces. For points in general position the Delaunay
// triangulation is unique, so any correct construction matches Qhull; the
// engine's additional job is to match Qhull's diagonal choice on cocircular
// inputs, where the triangulation is non-unique.
//
// This file holds a robust baseline construction (exact-predicate incremental
// insertion). Qhull-faithful cocircular tie-breaking is built on top of it.
package qhull

import (
	"fmt"
	"sort"
)

// Delaunay returns the Delaunay triangulation of the points (x, y) as a list of
// triangles (anticlockwise vertex indices) and, for each triangle, its three
// neighbours. neighbors[i][j] is the triangle adjacent across the edge from
// vertex j to vertex (j+1)%3, or -1 on the convex hull boundary.
func Delaunay(x, y []float64) (triangles, neighbors [][3]int, err error) {
	if len(x) != len(y) {
		return nil, nil, fmt.Errorf("qhull: x and y length mismatch (%d vs %d)", len(x), len(y))
	}
	if len(x) < 3 {
		return nil, nil, fmt.Errorf("qhull: need at least 3 points, got %d", len(x))
	}

	tris, ok := bowyerWatson(x, y)
	if !ok {
		return nil, nil, fmt.Errorf("qhull: degenerate input (all points collinear?)")
	}
	neighbors = computeNeighbors(tris)
	return tris, neighbors, nil
}

// edge is an oriented half-edge used to derive triangle adjacency.
type triEdge struct{ u, v int }

// computeNeighbors derives the matplotlib-style neighbour array from a list of
// consistently anticlockwise-wound triangles: the directed edge (u,v) of one
// triangle is shared as (v,u) by its neighbour.
func computeNeighbors(tris [][3]int) [][3]int {
	loc := make(map[triEdge][2]int, len(tris)*3) // edge -> (triIdx, edgeIdx)
	for i, tr := range tris {
		for j := 0; j < 3; j++ {
			loc[triEdge{tr[j], tr[(j+1)%3]}] = [2]int{i, j}
		}
	}
	nbrs := make([][3]int, len(tris))
	for i, tr := range tris {
		nbrs[i] = [3]int{-1, -1, -1}
		for j := 0; j < 3; j++ {
			if l, ok := loc[triEdge{tr[(j+1)%3], tr[j]}]; ok {
				nbrs[i][j] = l[0]
			}
		}
	}
	return nbrs
}

// bwTri is an anticlockwise-wound triangle during incremental insertion.
type bwTri struct{ a, b, c int }

// bowyerWatson builds the Delaunay triangulation with the incremental
// Bowyer–Watson algorithm using exact predicates, so the resulting triangle set
// is the true Delaunay triangulation. Cocircular ties are resolved by treating
// "on the circumcircle" as outside (inCircle > 0 strictly), a deterministic
// rule; matching Qhull's specific cocircular diagonal is handled separately.
func bowyerWatson(x, y []float64) ([][3]int, bool) {
	n := len(x)
	minX, maxX, minY, maxY := x[0], x[0], y[0], y[0]
	for i := 1; i < n; i++ {
		if x[i] < minX {
			minX = x[i]
		}
		if x[i] > maxX {
			maxX = x[i]
		}
		if y[i] < minY {
			minY = y[i]
		}
		if y[i] > maxY {
			maxY = y[i]
		}
	}
	dx, dy := maxX-minX, maxY-minY
	delta := dx
	if dy > delta {
		delta = dy
	}
	if delta == 0 {
		return nil, false
	}
	midX, midY := (minX+maxX)/2, (minY+maxY)/2

	// Super-triangle vertices appended after the real points. The margin is
	// large because the in-circle test is exact (math/big.Rat): a bigger
	// enclosing triangle is always strictly safer, never less accurate, and
	// avoids dropping thin hull triangles whose circumcircle would otherwise
	// reach a too-close super vertex.
	const k = 100000.0
	px := append(append([]float64(nil), x...), midX-k*delta, midX, midX+k*delta)
	py := append(append([]float64(nil), y...), midY-delta, midY+k*delta, midY-delta)

	super, ok := orientTri(n, n+1, n+2, px, py)
	if !ok {
		return nil, false
	}
	tris := []bwTri{super}

	for p := 0; p < n; p++ {
		bad := make([]bool, len(tris))
		boundary := make(map[[2]int]int)
		for i, tr := range tris {
			if inCircle(px[tr.a], py[tr.a], px[tr.b], py[tr.b], px[tr.c], py[tr.c], px[p], py[p]) > 0 {
				bad[i] = true
				boundary[sortEdge(tr.a, tr.b)]++
				boundary[sortEdge(tr.b, tr.c)]++
				boundary[sortEdge(tr.c, tr.a)]++
			}
		}
		kept := tris[:0]
		for i, tr := range tris {
			if !bad[i] {
				kept = append(kept, tr)
			}
		}
		tris = kept
		for edge, count := range boundary {
			if count != 1 {
				continue
			}
			if tr, ok := orientTri(edge[0], edge[1], p, px, py); ok {
				tris = append(tris, tr)
			}
		}
	}

	out := make([][3]int, 0, len(tris))
	for _, tr := range tris {
		if tr.a >= n || tr.b >= n || tr.c >= n {
			continue
		}
		out = append(out, [3]int{tr.a, tr.b, tr.c})
	}
	sort.Slice(out, func(i, j int) bool {
		for k := 0; k < 3; k++ {
			if out[i][k] != out[j][k] {
				return out[i][k] < out[j][k]
			}
		}
		return false
	})
	return out, len(out) > 0
}

// orientTri returns the triangle (a,b,c) re-wound anticlockwise, or false if the
// three points are collinear.
func orientTri(a, b, c int, x, y []float64) (bwTri, bool) {
	s := orient2d(x[a], y[a], x[b], y[b], x[c], y[c])
	if s == 0 {
		return bwTri{}, false
	}
	if s < 0 {
		a, b = b, a
	}
	return bwTri{a, b, c}, true
}

func sortEdge(a, b int) [2]int {
	if a < b {
		return [2]int{a, b}
	}
	return [2]int{b, a}
}

package qhull

import (
	"fmt"
	"math"
	"sort"
)

// This file closes the cocircular tie-breaking that the exact-predicate engine
// (delaunay.go) leaves arbitrary. matplotlib's Qhull triangulates each Delaunay
// cell by fanning it from the cell's last-created vertex: qh_triangulate_facet
// picks apex = SETfirst_(facet->vertices), and Qhull keeps each facet's vertex
// set inverse-sorted by vertex->id (poly2_r.c:25), so the apex is the highest-id
// = last-created vertex (vertex->id = qh->vertex_id++ at creation). Hence the
// diagonal of a cocircular cell is fixed entirely by the vertex CREATION ORDER.
//
// delaunayFromOrder reproduces that choice: it takes the exact (but arbitrarily
// split) Delaunay triangulation, merges adjacent triangles whose shared edge is
// cocircular into convex polygon cells, and re-fans each cell from its
// last-created vertex. The creation order itself is computed by the qh_buildhull
// port (see build.go/buildhull.go); this function is the geometry that consumes
// it. Feeding it Qhull's true creation order reproduces Qhull's triangulation
// bit-for-bit (validated against the differential corpus).

// delaunayFromOrder returns the Delaunay triangulation whose cocircular diagonals
// match the vertex creation order. order is a permutation of [0,n): order[k] is
// the id of the point Qhull created k-th, so a larger index means a later (higher
// vertex->id) creation. Triangles are wound anticlockwise; the triangle list and
// neighbours follow the same conventions as Delaunay.
func delaunayFromOrder(x, y []float64, order []int) (triangles, neighbors [][3]int, err error) {
	n := len(x)
	if len(y) != n {
		return nil, nil, fmt.Errorf("qhull: x and y length mismatch (%d vs %d)", n, len(y))
	}
	rank, err := creationRank(order, n)
	if err != nil {
		return nil, nil, err
	}

	base, baseNbrs, err := Delaunay(x, y)
	if err != nil {
		return nil, nil, err
	}

	cells := groupCocircularCells(project(x, y), base, baseNbrs)
	out := fanCells(x, y, base, cells, rank)
	return out, computeNeighbors(out), nil
}

// fanCells re-triangulates every cell as a fan from its last-created vertex and
// returns the triangles in the deterministic (lexicographic) order Delaunay uses.
func fanCells(x, y []float64, base [][3]int, cells [][]int, rank []int) [][3]int {
	out := make([][3]int, 0, len(base))
	for _, cell := range cells {
		out = append(out, fanCell(x, y, base, cell, rank)...)
	}
	sortTriangles(out)
	return out
}

// DelaunayMatched returns the Delaunay triangulation with matplotlib/Qhull's
// cocircular diagonal choice resolved from the computed vertex creation order
// (buildHullOrderRidge + the per-cell fan). It is layered on the robust exact-predicate
// Delaunay and degrades gracefully:
//
//   - General-position inputs have no cocircular cells, so the exact triangulation
//     is already canonical and is returned directly (the order computation, which
//     would be wasted, is skipped — this is also the fast path for large inputs).
//   - When cocircular cells exist, the build order is computed and each cell is
//     fanned from its last-created vertex, reproducing Qhull's diagonal.
//   - If the order computation bails (an unported hull degeneracy), the exact
//     triangulation — itself a valid Delaunay triangulation — is returned.
//
// Triangles are wound anticlockwise and returned in deterministic order, with the
// same neighbour conventions as Delaunay.
func DelaunayMatched(x, y []float64) (triangles, neighbors [][3]int, err error) {
	base, baseNbrs, err := Delaunay(x, y)
	if err != nil {
		return nil, nil, err
	}
	cells := groupCocircularCells(project(x, y), base, baseNbrs)
	if !hasMultiCell(cells) {
		return base, baseNbrs, nil // general position: exact triangulation is canonical
	}
	order, ok := buildHullOrderRidge(project(x, y))
	if !ok {
		return base, baseNbrs, nil // fallback: a valid Delaunay triangulation
	}
	rank, err := creationRank(order, len(x))
	if err != nil {
		return base, baseNbrs, nil
	}
	out := fanCells(x, y, base, cells, rank)
	return out, computeNeighbors(out), nil
}

// hasMultiCell reports whether any cell merges more than one base triangle, i.e.
// whether the input actually has a cocircular diagonal to resolve.
func hasMultiCell(cells [][]int) bool {
	for _, c := range cells {
		if len(c) > 1 {
			return true
		}
	}
	return false
}

// creationRank inverts the creation-order permutation: rank[p] is the position of
// point p in order (its creation index). It validates that order is a permutation
// of [0,n).
func creationRank(order []int, n int) ([]int, error) {
	if len(order) != n {
		return nil, fmt.Errorf("qhull: creation order has %d entries, want %d", len(order), n)
	}
	rank := make([]int, n)
	seen := make([]bool, n)
	for k, p := range order {
		if p < 0 || p >= n || seen[p] {
			return nil, fmt.Errorf("qhull: creation order is not a permutation of [0,%d)", n)
		}
		seen[p] = true
		rank[p] = k
	}
	return rank, nil
}

// groupCocircularCells partitions the base triangles into cells: maximal groups
// of triangles connected through interior edges whose four points Qhull treats as
// cocircular (the lifted facets are coplanar within Qhull's premerge tolerance, so
// the edge is an arbitrary Delaunay diagonal). A triangle with no such interior
// edge forms a singleton cell. Each returned cell is a list of base triangle
// indices.
func groupCocircularCells(q *qstate, tris, nbrs [][3]int) [][]int {
	uf := newUnionFind(len(tris))
	for i, tr := range tris {
		for j := range 3 {
			m := nbrs[i][j]
			if m <= i { // visit each undirected adjacency once
				continue
			}
			if cocircularSharedEdge(q, tr, tris[m]) {
				uf.union(i, m)
			}
		}
	}
	groups := map[int][]int{}
	for i := range tris {
		r := uf.find(i)
		groups[r] = append(groups[r], i)
	}
	// Deterministic order: by smallest member triangle index.
	roots := make([]int, 0, len(groups))
	for r := range groups {
		roots = append(roots, r)
	}
	sort.Ints(roots)
	cells := make([][]int, 0, len(groups))
	for _, r := range roots {
		cells = append(cells, groups[r])
	}
	return cells
}

// cocircularRelTol is the threshold on the scale-invariant normalized in-circle
// determinant (cocircularResidual, in [0,1]) below which four points are treated
// as cocircular — i.e. the shared edge is an arbitrary Delaunay diagonal that
// Qhull's premerge would flatten. Qhull merges lifted facets that are coplanar
// within roundoff (premerge_centrum = 2*DISTround, ~1e-15 relative); the
// equivalent input-space statement is "cocircular within roundoff". The constant
// sits in the wide gap the corpus exhibits between constructed-cocircular inputs
// (a few ULPs, ~1e-16 — exactly 0 for integer grids) and genuinely distinct
// points (>= ~1e-5), comfortably above roundoff and far below real geometric
// separation, so the grouping is insensitive to its exact value.
const cocircularRelTol = 1e-12

// cocircularSharedEdge reports whether triangles t1 and t2 (which share an edge)
// bound a cell Qhull would merge: their four distinct vertices are cocircular
// within roundoff. The test uses cocircularResidual rather than an absolute
// lifted-distance test, which mis-scales on ill-conditioned (thin) base triangles
// that arise inside larger cocircular cells (e.g. regular polygons).
func cocircularSharedEdge(q *qstate, t1, t2 [3]int) bool {
	w2 := oppositeVertex(t1, t2)
	if w2 < 0 {
		return false
	}
	return cocircularResidual(q, t1[0], t1[1], t1[2], w2) <= cocircularRelTol
}

// cocircularResidual returns |det| / Σ|terms| for the in-circle determinant of the
// four lifted points (a, b, c, d), a dimensionless value in [0,1] that is zero iff
// the points are exactly cocircular and grows with the relative deviation. Using
// the paraboloid-lifted, Qbb-scaled coordinates from project keeps it in Qhull's
// arithmetic frame; the permanent-style denominator makes it scale- and
// conditioning-invariant.
func cocircularResidual(q *qstate, a, b, c, d int) float64 {
	pa, pb, pc, pd := q.pts[a], q.pts[b], q.pts[c], q.pts[d]
	adx, ady := pa[0]-pd[0], pa[1]-pd[1]
	bdx, bdy := pb[0]-pd[0], pb[1]-pd[1]
	cdx, cdy := pc[0]-pd[0], pc[1]-pd[1]
	ad := adx*adx + ady*ady
	bd := bdx*bdx + bdy*bdy
	cd := cdx*cdx + cdy*cdy
	det := adx*(bdy*cd-cdy*bd) - ady*(bdx*cd-cdx*bd) + ad*(bdx*cdy-cdx*bdy)
	perm := math.Abs(adx)*(math.Abs(bdy)*cd+math.Abs(cdy)*bd) +
		math.Abs(ady)*(math.Abs(bdx)*cd+math.Abs(cdx)*bd) +
		math.Abs(ad)*(math.Abs(bdx)*math.Abs(cdy)+math.Abs(cdx)*math.Abs(bdy))
	if perm == 0 {
		return math.Inf(1)
	}
	return math.Abs(det) / perm
}

// oppositeVertex returns the vertex of t2 that is not shared with t1, or -1 if
// the triangles do not share exactly one such vertex.
func oppositeVertex(t1, t2 [3]int) int {
	opp := -1
	for _, v := range t2 {
		if v != t1[0] && v != t1[1] && v != t1[2] {
			if opp != -1 {
				return -1
			}
			opp = v
		}
	}
	return opp
}

// fanCell re-triangulates one cocircular cell as a fan from its last-created
// vertex. The cell's boundary edges are those appearing in exactly one of its
// triangles; the apex is the cell vertex with the highest creation rank; each
// boundary edge not incident to the apex forms one anticlockwise triangle with
// it. A singleton cell reproduces its own triangle.
func fanCell(x, y []float64, tris [][3]int, cell, rank []int) [][3]int {
	boundary := cellBoundaryEdges(tris, cell)

	apex := -1
	bestRank := -1
	for _, ti := range cell {
		for _, v := range tris[ti] {
			if rank[v] > bestRank {
				bestRank, apex = rank[v], v
			}
		}
	}

	out := make([][3]int, 0, len(boundary))
	for _, e := range boundary {
		if e[0] == apex || e[1] == apex {
			continue
		}
		out = append(out, windCCW(apex, e[0], e[1], x, y))
	}
	return out
}

// cellBoundaryEdges returns the undirected edges that bound a cell: edges used by
// exactly one of the cell's triangles (interior diagonals are used by two).
func cellBoundaryEdges(tris [][3]int, cell []int) [][2]int {
	count := map[[2]int]int{}
	for _, ti := range cell {
		tr := tris[ti]
		for j := range 3 {
			count[sortEdge(tr[j], tr[(j+1)%3])]++
		}
	}
	edges := make([][2]int, 0, len(count))
	for e, c := range count {
		if c == 1 {
			edges = append(edges, e)
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i][0] != edges[j][0] {
			return edges[i][0] < edges[j][0]
		}
		return edges[i][1] < edges[j][1]
	})
	return edges
}

// windCCW returns the triangle (a,b,c) re-ordered anticlockwise on the real
// coordinates.
func windCCW(a, b, c int, x, y []float64) [3]int {
	if orient2d(x[a], y[a], x[b], y[b], x[c], y[c]) < 0 {
		b, c = c, b
	}
	return [3]int{a, b, c}
}

// sortTriangles orders a triangle list lexicographically, matching the
// deterministic order Delaunay produces.
func sortTriangles(tris [][3]int) {
	sort.Slice(tris, func(i, j int) bool {
		for k := range 3 {
			if tris[i][k] != tris[j][k] {
				return tris[i][k] < tris[j][k]
			}
		}
		return false
	})
}

// unionFind is a minimal disjoint-set structure for grouping triangles.
type unionFind struct{ parent []int }

func newUnionFind(n int) *unionFind {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &unionFind{parent: p}
}

func (u *unionFind) find(i int) int {
	for u.parent[i] != i {
		u.parent[i] = u.parent[u.parent[i]]
		i = u.parent[i]
	}
	return i
}

func (u *unionFind) union(a, b int) {
	ra, rb := u.find(a), u.find(b)
	if ra != rb {
		u.parent[rb] = ra
	}
}

package qhull

import (
	"math"
	"slices"
	"sort"
)

// This file ports the incremental part of Qhull's (8.0.2) 2-D Delaunay pipeline:
// qh_initialhull (build the initial simplex from the qh_maxsimplex vertices),
// qh_partitionall (assign points to facet outside sets), the qh_buildhull loop
// (qh_nextfurthest + qh_addpoint: qh_findhorizon, qh_makenewfacets,
// qh_partitionvisible), and — crucially for cocircular inputs — qh_premerge's
// coplanarhorizon merge (qh_mergecycle). The build is faithful to Qhull's float64
// arithmetic and, above all, to its data-structure ORDERING, because the order in
// which points become vertices is the vertex creation order that fixes every
// cocircular cell's fan apex (see fanfromorder.go).
//
// Representation note: rather than Qhull's explicit ridge graph, a facet is a SET
// of (lifted) vertices; its boundary polygon and edges are derived on demand by
// angle-sorting the vertices in the facet plane about the outward normal, and
// neighbour links are rebuilt globally by matching opposite directed edges. This
// is exact for convex-hull facets (always convex polygons) and makes the
// coplanarhorizon merge a one-line vertex-set union, while preserving the list
// ordering, outside-set assignment, and creation order Qhull produces.
//
// References: third_party/qhull-8.0.2/src/libqhull_r/{libqhull_r.c,poly_r.c,
// poly2_r.c,merge_r.c,geom_r.c}.

// hullFacet is a facet of the 3-D (lifted) hull. It is simplicial (3 vertices)
// when created and may grow non-simplicial through coplanarhorizon merges. verts
// holds its lifted-point indices (a set; boundary order is derived via
// orderedVerts). nbr[k] is the neighbour across the k-th edge of the ordered
// boundary; it is rebuilt by rebuildNeighbors. outside holds point indices the
// facet is "above" (outside of), furthest last.
type hullFacet struct {
	verts    []int
	nbr      []int // neighbour facet id across ordered edge k; -1 if none
	normal   [3]float64
	offset   float64
	outside  []int
	furthest float64 // distance of the last (furthest) outside point
	id       int

	toporient bool
	visible   bool
	newfacet  bool

	coplanarhorizon bool // apex of the in-progress add is coplanar with this facet
	visitid         int  // qh_findhorizon BFS marker
	prev, next      int  // doubly-linked facet list (facet ids); sentinels at both ends
}

// hull is the incremental builder state, mirroring the subset of qhT the 2-D
// Delaunay build touches.
type hull struct {
	q        *qstate
	facets   []*hullFacet // indexed by facet id (append-only; visible ones flagged)
	headID   int          // facet_list head sentinel id
	tailID   int          // facet_tail sentinel id
	nextID   int          // facet_next (id) for the buildhull walk
	createID int          // next vertex creation id to assign

	// order is the vertex creation order being built: order[k] = point id created
	// k-th (real points only; the Qz infinity point is skipped).
	order []int
	vid   []int // vid[pt] = creation id of point pt's vertex, or -1 if not a vertex

	interior    [3]float64 // qh.interior_point (centroid of the initial simplex)
	minVisible  float64    // qh.MINvisible
	maxCoplanar float64    // qh.MAXcoplanar
	minOutside  float64    // qh.MINoutside
	visit       int        // qh.visit_id counter
	ok          bool       // cleared on an unsupported degeneracy (Gaussian fallback)
}

// buildHullOrder runs the incremental hull on the projected points and returns
// the vertex creation order (real point ids, in ascending creation id). It is the
// computed analogue of the captured creation_order.json. It reports false if the
// build cannot proceed (e.g. a degenerate facet needing the Gaussian fallback,
// not yet ported).
func buildHullOrder(q *qstate) ([]int, bool) {
	h := &hull{q: q, createID: 0}
	h.minVisible = 2 * q.distRound // premerge_centrum (hull_dim<=3, merging)
	h.maxCoplanar = h.minVisible
	h.minOutside = 2 * h.minVisible
	if !h.initialHull() {
		return nil, false
	}
	h.partitionAll()
	if !h.buildLoop() {
		return nil, false
	}
	return h.order, true
}

// delaunayComputed is the fully self-contained (no Qhull, no captured fixture)
// Qhull-faithful engine: it computes the vertex creation order with the
// incremental hull (including the coplanarhorizon merge), then fans each
// cocircular cell from its last-created vertex, reproducing matplotlib's diagonal
// choice.
func delaunayComputed(x, y []float64) (triangles, neighbors [][3]int, err error) {
	order, ok := buildHullOrder(project(x, y))
	if !ok {
		return nil, nil, errDegenerateBuild
	}
	return delaunayFromOrder(x, y, order)
}

// errDegenerateBuild is returned when the incremental hull hits a degeneracy that
// needs the unported Gaussian-elimination fallback.
var errDegenerateBuild = errBuild("qhull: degenerate facet (Gaussian fallback unported)")

type errBuild string

func (e errBuild) Error() string { return string(e) }

// recordVertex appends a newly-created real vertex to the creation order. The Qz
// infinity point (index q.n) is a hull vertex but not an input site, so it is
// recorded as consuming an id slot but omitted from the returned order.
func (h *hull) recordVertex(pt int) {
	if h.vid == nil {
		h.vid = make([]int, h.q.n+1)
		for i := range h.vid {
			h.vid[i] = -1
		}
	}
	h.vid[pt] = h.createID
	if pt != h.q.n {
		h.order = append(h.order, pt)
	}
	h.createID++
}

// ---- facet list helpers -------------------------------------------------

// newFacet allocates a facet with the given vertex set.
func (h *hull) newFacet(verts []int, toporient bool) *hullFacet {
	f := &hullFacet{
		verts:     append([]int(nil), verts...),
		id:        len(h.facets),
		toporient: toporient,
		prev:      -1,
		next:      -1,
	}
	h.facets = append(h.facets, f)
	return f
}

// appendFacet links f at the tail of the facet list (before the tail sentinel).
func (h *hull) appendFacet(f *hullFacet) {
	tail := h.facets[h.tailID]
	prev := tail.prev
	h.facets[prev].next = f.id
	f.prev = prev
	f.next = h.tailID
	tail.prev = f.id
}

// prependFacet links f at the head of the facet list (just after the head
// sentinel) and makes it facet_next (qh_prependfacet before qh.facet_next).
func (h *hull) prependFacet(f *hullFacet) {
	head := h.facets[h.headID]
	next := head.next
	head.next = f.id
	f.prev = h.headID
	f.next = next
	h.facets[next].prev = f.id
	h.nextID = f.id
}

// removeFacet unlinks f from the facet list, advancing facet_next past it.
func (h *hull) removeFacet(f *hullFacet) {
	if h.nextID == f.id {
		h.nextID = f.next
	}
	h.facets[f.prev].next = f.next
	h.facets[f.next].prev = f.prev
}

// liveFacets calls fn for each non-visible facet in facet-list order.
func (h *hull) liveFacets(fn func(*hullFacet)) {
	for id := h.facets[h.headID].next; id != h.tailID; id = h.facets[id].next {
		if !h.facets[id].visible {
			fn(h.facets[id])
		}
	}
}

// furthestNext mirrors qh_furthestnext: it moves the facet holding the globally
// furthest outside point to the front of the list and makes it facet_next. Qhull
// calls this once in qh_initbuild (after partitioning) to seed the build with the
// overall furthest point; with PICKfurthest off, the subsequent qh_buildhull walk
// is plain facet-list order. Ties go to the first such facet in list order
// (strict > below), matching Qhull.
func (h *hull) furthestNext() {
	bestID := -1
	bestDist := -math.MaxFloat64
	for id := h.facets[h.headID].next; id != h.tailID; id = h.facets[id].next {
		f := h.facets[id]
		if f.visible || len(f.outside) == 0 {
			continue
		}
		if f.furthest > bestDist {
			bestDist, bestID = f.furthest, id
		}
	}
	if bestID >= 0 {
		f := h.facets[bestID]
		h.removeFacet(f)
		h.prependFacet(f)
	}
}

// ---- geometry primitives (qh_distplane / qh_setfacetplane, hull_dim 3) ---

// distplane mirrors qh_distplane: signed distance of the lifted point pt to the
// facet's hyperplane (positive = outside/above).
func (h *hull) distplane(pt int, f *hullFacet) float64 {
	p := h.q.pts[pt]
	return f.offset + p[0]*f.normal[0] + p[1]*f.normal[1] + p[2]*f.normal[2]
}

// setFacetPlane mirrors qh_setfacetplane via qh_sethyperplane_det (hull_dim 3):
// the normal is the cofactor determinant of three vertex coordinate differences,
// normalised to unit length and oriented by toporient; offset = -point0·normal. It
// reports false if the determinant normal is degenerate (nearzero) — the Gaussian
// fallback is not yet ported, so such inputs bail out of the build. Only the first
// three vertices are used; for a merged (coplanar) facet the plane is unchanged
// and setFacetPlane is not recomputed.
func (h *hull) setFacetPlane(f *hullFacet) bool {
	r0, r1, r2 := h.q.pts[f.verts[0]], h.q.pts[f.verts[1]], h.q.pts[f.verts[2]]
	dx1, dy1, dz1 := r1[0]-r0[0], r1[1]-r0[1], r1[2]-r0[2]
	dx2, dy2, dz2 := r2[0]-r0[0], r2[1]-r0[1], r2[2]-r0[2]
	n := [3]float64{
		det2(dy2, dz2, dy1, dz1),
		det2(dx1, dz1, dx2, dz2),
		det2(dx2, dy2, dx1, dy1),
	}
	norm := math.Sqrt(n[0]*n[0] + n[1]*n[1] + n[2]*n[2])
	if !(norm > 0) {
		return false
	}
	if !f.toporient {
		norm = -norm
	}
	n[0] /= norm
	n[1] /= norm
	n[2] /= norm
	f.normal = n
	f.offset = -(r0[0]*n[0] + r0[1]*n[1] + r0[2]*n[2])
	// nearzero check: any of the three base vertices far from its own plane.
	for _, v := range f.verts[:3] {
		if v == f.verts[0] {
			continue
		}
		p := h.q.pts[v]
		d := f.offset + p[0]*n[0] + p[1]*n[1] + p[2]*n[2]
		if d > h.q.distRound || d < -h.q.distRound {
			return false // would need qh_sethyperplane_gauss
		}
	}
	return true
}

// orientOutside flips a facet so the interior point is below it (qh_orientoutside).
func (h *hull) orientOutside(f *hullFacet) {
	if h.interior[0]*f.normal[0]+h.interior[1]*f.normal[1]+h.interior[2]*f.normal[2]+f.offset > h.q.distRound {
		f.toporient = !f.toporient
		f.normal[0], f.normal[1], f.normal[2] = -f.normal[0], -f.normal[1], -f.normal[2]
		f.offset = -f.offset
	}
}

// orderedVerts returns f's vertices ordered anticlockwise about its outward
// normal, i.e. the facet's boundary polygon. For a convex-hull facet (always a
// convex polygon) angle-sorting in the facet plane reproduces the boundary cycle,
// so consecutive pairs are the facet's edges and shared edges appear reversed in
// the adjoining facet (used by rebuildNeighbors and the cone build).
func (h *hull) orderedVerts(f *hullFacet) []int {
	vs := f.verts
	if len(vs) <= 1 {
		return vs
	}
	n := f.normal
	// in-plane basis e1,e2 with e1×e2 ∝ n
	e1 := basisPerp(n)
	e2 := [3]float64{
		n[1]*e1[2] - n[2]*e1[1],
		n[2]*e1[0] - n[0]*e1[2],
		n[0]*e1[1] - n[1]*e1[0],
	}
	var c [3]float64
	for _, v := range vs {
		p := h.q.pts[v]
		c[0] += p[0]
		c[1] += p[1]
		c[2] += p[2]
	}
	inv := 1 / float64(len(vs))
	c[0] *= inv
	c[1] *= inv
	c[2] *= inv
	out := append([]int(nil), vs...)
	ang := make(map[int]float64, len(vs))
	for _, v := range vs {
		p := h.q.pts[v]
		dx, dy, dz := p[0]-c[0], p[1]-c[1], p[2]-c[2]
		a := dx*e1[0] + dy*e1[1] + dz*e1[2]
		b := dx*e2[0] + dy*e2[1] + dz*e2[2]
		ang[v] = math.Atan2(b, a)
	}
	sort.Slice(out, func(i, j int) bool { return ang[out[i]] < ang[out[j]] })
	return out
}

// basisPerp returns a unit vector perpendicular to n, choosing the axis least
// aligned with n for numerical stability.
func basisPerp(n [3]float64) [3]float64 {
	ax, ay, az := math.Abs(n[0]), math.Abs(n[1]), math.Abs(n[2])
	var t [3]float64
	switch {
	case ax <= ay && ax <= az:
		t = [3]float64{1, 0, 0}
	case ay <= az:
		t = [3]float64{0, 1, 0}
	default:
		t = [3]float64{0, 0, 1}
	}
	// e1 = normalize(t - (t·n)n)
	d := t[0]*n[0] + t[1]*n[1] + t[2]*n[2]
	e := [3]float64{t[0] - d*n[0], t[1] - d*n[1], t[2] - d*n[2]}
	l := math.Sqrt(e[0]*e[0] + e[1]*e[1] + e[2]*e[2])
	if l == 0 {
		return [3]float64{1, 0, 0}
	}
	return [3]float64{e[0] / l, e[1] / l, e[2] / l}
}

// rebuildNeighbors recomputes every live facet's nbr slice by matching opposite
// directed boundary edges across the whole hull (a closed orientable manifold, so
// each internal edge (a,b) has exactly one reverse (b,a)).
func (h *hull) rebuildNeighbors() {
	type es struct{ f, k int }
	edge := map[[2]int]es{}
	h.liveFacets(func(f *hullFacet) {
		ov := h.orderedVerts(f)
		f.verts = ov
		if cap(f.nbr) < len(ov) {
			f.nbr = make([]int, len(ov))
		} else {
			f.nbr = f.nbr[:len(ov)]
		}
		for k := range f.nbr {
			f.nbr[k] = -1
		}
	})
	h.liveFacets(func(f *hullFacet) {
		n := len(f.verts)
		for k := range n {
			a, b := f.verts[k], f.verts[(k+1)%n]
			if o, ok := edge[[2]int{b, a}]; ok {
				f.nbr[k] = o.f
				h.facets[o.f].nbr[o.k] = f.id
				delete(edge, [2]int{b, a})
			} else {
				edge[[2]int{a, b}] = es{f.id, k}
			}
		}
	})
}

// ---- initial simplex (qh_createsimplex + qh_initialhull) -----------------

// initialHull builds the initial simplex from qh_maxsimplex and orients its
// facets so the interior point lies below each. It reports false on degeneracy.
func (h *hull) initialHull() bool {
	simplex := h.q.maxsimplex() // creation order [p0,p1,p2,p3]
	if len(simplex) != hullDim+1 {
		return false
	}
	head := h.newFacet(nil, false)
	tail := h.newFacet(nil, false)
	h.headID, h.tailID = head.id, tail.id
	head.next, tail.prev = tail.id, head.id
	head.prev, tail.next = -1, -1

	// Record the four simplex vertices in creation order.
	for _, p := range simplex {
		h.recordVertex(p)
	}
	// Vertices in descending creation id (Qhull's qh_initialvertices order).
	vset := [4]int{simplex[3], simplex[2], simplex[1], simplex[0]}

	// interior point = centroid of the simplex vertices.
	for _, p := range simplex {
		h.interior[0] += h.q.pts[p][0]
		h.interior[1] += h.q.pts[p][1]
		h.interior[2] += h.q.pts[p][2]
	}
	h.interior[0] /= float64(hullDim + 1)
	h.interior[1] /= float64(hullDim + 1)
	h.interior[2] /= float64(hullDim + 1)

	// One facet per omitted vertex, alternating toporient (qh_createsimplex).
	toporient := true
	simplexFacets := make([]*hullFacet, 0, hullDim+1)
	for i := range hullDim + 1 {
		v := make([]int, 0, hullDim)
		for j := range hullDim + 1 {
			if j != i {
				v = append(v, vset[j])
			}
		}
		f := h.newFacet(v, toporient)
		h.appendFacet(f)
		simplexFacets = append(simplexFacets, f)
		toporient = !toporient
	}
	for _, f := range simplexFacets {
		if !h.setFacetPlane(f) {
			return false
		}
	}
	for _, f := range simplexFacets {
		h.orientOutside(f)
	}
	h.rebuildNeighbors()
	h.nextID = h.facets[h.headID].next
	return true
}

// ---- partition (qh_partitionall) ----------------------------------------

// partitionAll mirrors qh_partitionall for the (!BESToutside, MERGING,
// KEEPcoplanar) Delaunay config. Block 1 walks the facet list and assigns each
// not-yet-claimed point to the FIRST facet it is clearly outside of (distance >=
// distoutside), keeping that facet's running-furthest point deferred to the end of
// its outside set; points not clearly outside any facet fall through. Block 2
// re-partitions those leftovers with the best-facet search (qh_partitionpoint via
// qh_findbestnew, which with no new facets reduces to the best live facet).
func (h *hull) partitionAll() {
	inSimplex := make([]bool, h.q.n+1)
	h.liveFacets(func(f *hullFacet) {
		for _, v := range f.verts {
			inSimplex[v] = true
		}
	})
	pointset := make([]int, 0, h.q.n)
	for p := 0; p <= h.q.n; p++ {
		if !inSimplex[p] {
			pointset = append(pointset, p)
		}
	}
	// distoutside = 2*MINoutside at partition time (USEfindbestnew false,
	// max_outside 0; MERGING true).
	distoutside := 2 * h.minOutside

	for id := h.facets[h.headID].next; id != h.tailID; id = h.facets[id].next {
		f := h.facets[id]
		if f.visible {
			continue
		}
		leftover := pointset[:0]
		bestpoint := -1
		var bestdist float64
		for _, p := range pointset {
			d := h.distplane(p, f)
			if d < distoutside {
				leftover = append(leftover, p)
				continue
			}
			switch {
			case bestpoint == -1:
				bestpoint, bestdist = p, d
			case d > bestdist:
				f.outside = append(f.outside, bestpoint)
				bestpoint, bestdist = p, d
			default:
				f.outside = append(f.outside, p)
			}
		}
		if bestpoint != -1 {
			f.outside = append(f.outside, bestpoint)
			f.furthest = bestdist
		}
		pointset = leftover
	}

	for _, p := range pointset {
		h.partitionPoint(p)
	}
}

// findBest returns the live facet that the point is furthest above, and that
// distance (qh_findbest's converged best for this config).
func (h *hull) findBest(pt int) (*hullFacet, float64) {
	var best *hullFacet
	bestDist := -math.MaxFloat64
	h.liveFacets(func(f *hullFacet) {
		if d := h.distplane(pt, f); d > bestDist {
			bestDist, best = d, f
		}
	})
	return best, bestDist
}

// partitionPoint mirrors qh_partitionpoint over the live facet list (block 2 of
// qh_partitionall): it finds the best facet, and if the point is outside it
// (dist >= MINoutside) appends it keeping the furthest last, then — like Qhull —
// moves a freshly-occupied facet to the tail so it is processed after facet_next.
func (h *hull) partitionPoint(p int) {
	f, dist := h.findBest(p)
	if f == nil || dist < h.minOutside {
		return // coplanar/inside: dropped by the clean build
	}
	wasEmpty := len(f.outside) == 0
	h.addOutside(f, p, dist)
	if wasEmpty && h.nextID != f.id {
		h.removeFacet(f)
		h.appendFacet(f)
	}
}

// addOutside inserts p into f's outside set keeping the furthest point last
// (qh_partitionpoint's append / append-2nd-last rule).
func (h *hull) addOutside(f *hullFacet, p int, dist float64) {
	if len(f.outside) == 0 || dist > f.furthest {
		f.outside = append(f.outside, p)
		f.furthest = dist
		return
	}
	n := len(f.outside)
	f.outside = append(f.outside, 0)
	f.outside[n] = f.outside[n-1]
	f.outside[n-1] = p
}

// ---- build loop (qh_buildhull + qh_nextfurthest + qh_addpoint) -----------

// buildLoop drains facet outside sets in Qhull's facet_next order, adding each
// furthest point. It reports false on an unsupported degeneracy.
func (h *hull) buildLoop() bool {
	h.ok = true
	h.nextID = h.facets[h.headID].next // qh_buildhull: facet_next = facet_list
	h.furthestNext()                   // qh_initbuild: seed with the overall furthest facet
	for {
		f := h.nextFurthest()
		if f == nil {
			break
		}
		furthest := f.outside[len(f.outside)-1]
		f.outside = f.outside[:len(f.outside)-1]
		if !h.addPoint(furthest, f) {
			return false
		}
		if !h.ok {
			return false
		}
	}
	return true
}

// nextFurthest walks the facet list from facet_next and returns the first live
// facet with a non-empty outside set, without advancing facet_next past it
// (qh_nextfurthest on the default no-NARROWhull/RANDOMoutside path).
func (h *hull) nextFurthest() *hullFacet {
	for h.nextID != h.tailID {
		f := h.facets[h.nextID]
		if f.visible || len(f.outside) == 0 {
			h.nextID = f.next
			continue
		}
		return f
	}
	return nil
}

// ---- addPoint (qh_addpoint: horizon, cone, premerge, partition, delete) ---

// addPoint inserts furthest into the hull. It marks the visible facets and the
// coplanar horizon facets, builds a cone of new facets from the apex over the
// horizon edges (folding coplanar-horizon edges into their horizon facet instead
// — the qh_premerge coplanarhorizon merge), redistributes the orphaned outside
// points, and deletes the visible facets. The apex is the single new vertex, so
// the sequence of addPoint calls is the vertex creation order.
func (h *hull) addPoint(furthest int, seed *hullFacet) bool {
	h.rebuildNeighbors() // ensure boundary/adjacency current for the horizon flood
	visible := h.findHorizon(furthest, seed)

	apex := furthest
	h.recordVertex(apex)

	newFacets := make([]*hullFacet, 0, len(visible)+2)
	var mergedOrder []*hullFacet // merged horizon facets, first-encounter (samecycle) order
	mergedSeen := map[int]bool{}

	for _, vf := range visible {
		n := len(vf.verts)
		for _, k := range h.edgeOrder(vf) {
			nb := h.facets[vf.nbr[k]]
			if nb.visible {
				continue
			}
			u, v := vf.verts[k], vf.verts[(k+1)%n]
			if nb.coplanarhorizon {
				// qh_mergecycle: fold this cone facet into the coplanar horizon
				// facet (apex joins it; it keeps its plane, moves to the tail and
				// becomes a new facet). The horizon facet is recorded once, in the
				// order its samecycle is first encountered = newfacet_list order.
				if !slices.Contains(nb.verts, apex) {
					nb.verts = append(nb.verts, apex)
				}
				if !mergedSeen[nb.id] {
					mergedSeen[nb.id] = true
					mergedOrder = append(mergedOrder, nb)
				}
				continue
			}
			nf := h.makeConeFacet(apex, u, v)
			if nf == nil {
				return false
			}
			newFacets = append(newFacets, nf)
		}
	}

	// Move each merged horizon facet to the tail and flag it new (qh_mergecycle_facets).
	for _, mf := range mergedOrder {
		mf.coplanarhorizon = false
		mf.newfacet = true
		h.removeFacet(mf)
		h.appendFacet(mf)
		newFacets = append(newFacets, mf)
	}
	for _, nb := range h.allFacets() {
		nb.coplanarhorizon = false
	}

	if len(newFacets) == 0 {
		return false
	}

	h.partitionVisible(visible, newFacets)

	for _, vf := range visible {
		h.removeFacet(vf)
	}
	h.rebuildNeighbors()
	return h.ok
}

// edgeOrder returns the indices into f's ordered boundary edges in the order
// Qhull visits them when building cone facets and flooding the horizon. Qhull's
// qh_makenew_simplicial / qh_findhorizon iterate a simplicial facet's neighbours
// in vertex order, i.e. by the OPPOSITE (third) vertex's creation id descending;
// reproducing that order makes cone-facet ids — and so the furthest-point
// tie-breaking among symmetric facets — match Qhull. Non-simplicial facets keep
// their geometric boundary order.
func (h *hull) edgeOrder(f *hullFacet) []int {
	n := len(f.verts)
	ks := make([]int, n)
	for i := range ks {
		ks[i] = i
	}
	if n == 3 {
		opp := func(k int) int { return h.vid[f.verts[(k+2)%3]] }
		sort.SliceStable(ks, func(i, j int) bool { return opp(ks[i]) > opp(ks[j]) })
	}
	return ks
}

// allFacets returns every live (non-sentinel, non-visible) facet.
func (h *hull) allFacets() []*hullFacet {
	var out []*hullFacet
	h.liveFacets(func(f *hullFacet) { out = append(out, f) })
	return out
}

// findHorizon marks every facet the point is above (dist > MINvisible), starting
// from the seed and flooding across neighbours; non-visible neighbours coplanar
// with the point (dist within [-MAXcoplanar, MINvisible]) are flagged
// coplanarhorizon (qh_findhorizon). Returns the visible facets in visible-list
// order.
func (h *hull) findHorizon(pt int, seed *hullFacet) []*hullFacet {
	h.visit++
	visit := h.visit
	seed.visible = true
	seed.visitid = visit
	visible := []*hullFacet{seed}
	for i := 0; i < len(visible); i++ {
		vf := visible[i]
		for _, k := range h.edgeOrder(vf) {
			nb := h.facets[vf.nbr[k]]
			if nb.visitid == visit {
				continue
			}
			nb.visitid = visit
			d := h.distplane(pt, nb)
			switch {
			case d > h.minVisible:
				nb.visible = true
				visible = append(visible, nb)
			case d >= -h.maxCoplanar:
				nb.coplanarhorizon = true
			}
		}
	}
	return visible
}

// makeConeFacet creates a new triangular facet {apex, u, v}, sets its plane, and
// orients it outward. Returns nil on a degeneracy needing the Gaussian fallback.
func (h *hull) makeConeFacet(apex, u, v int) *hullFacet {
	nf := h.newFacet([]int{apex, u, v}, false)
	nf.newfacet = true
	if !h.setFacetPlane(nf) {
		h.ok = false
		return nil
	}
	h.orientOutside(nf)
	h.appendFacet(nf)
	return nf
}

// findBestNew mirrors qh_findbestnew over the new facets: it scans them in
// newfacet_list order and returns the FIRST facet the point is clearly outside of
// (dist >= distoutside), else the overall best. Early-return on clearly-outside —
// not a global maximum — is what reproduces Qhull's outside-set assignment order.
func (h *hull) findBestNew(pt int, cand []*hullFacet) (*hullFacet, float64) {
	distoutside := 2 * h.minOutside
	var best *hullFacet
	bestDist := -math.MaxFloat64
	for _, f := range cand {
		if f.visible {
			continue
		}
		if d := h.distplane(pt, f); d > bestDist {
			best, bestDist = f, d
			if d >= distoutside {
				break
			}
		}
	}
	return best, bestDist
}

// partitionVisible redistributes the outside points of the deleted visible facets
// onto the new cone/merged facets (qh_partitionvisible), keeping each new facet's
// furthest point last.
func (h *hull) partitionVisible(visible, newFacets []*hullFacet) {
	for _, vf := range visible {
		for _, p := range vf.outside {
			f, dist := h.findBestNew(p, newFacets)
			if f != nil && dist >= h.minOutside {
				h.addOutside(f, p, dist)
			}
		}
		vf.outside = nil
	}
}

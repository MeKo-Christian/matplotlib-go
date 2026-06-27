package qhull

import "math"

// This file ports the incremental part of Qhull's (8.0.2) 2-D Delaunay pipeline:
// qh_initialhull (build the initial simplex from the qh_maxsimplex vertices),
// qh_partitionall (assign points to facet outside sets), and the qh_buildhull
// loop (qh_nextfurthest + qh_addpoint: qh_findbest, qh_findhorizon,
// qh_makenewfacets, qh_partitionvisible) — faithful to Qhull's float64 arithmetic
// and, crucially, to its data-structure ORDERING, because the order points are
// added is the vertex creation order that fixes every cocircular cell's fan apex
// (see fanfromorder.go). This is the clean (no-premerge) hull; premerge's effect
// on the build order for fully-cocircular inputs is layered on separately.
//
// References: third_party/qhull-8.0.2/src/libqhull_r/{libqhull_r.c,poly2_r.c,
// geom_r.c}.

// hullFacet is a triangular facet of the 3-D (lifted) hull. vertices are point
// indices into qstate.pts, kept inverse-sorted by creation id (Qhull's
// invariant), so vertices[0] is the highest-id vertex. neighbors[k] is the facet
// across the edge opposite vertices[k]. outside holds point indices this facet is
// "above" (outside of), furthest last.
type hullFacet struct {
	vertices   [3]int
	neighbors  [3]int // facet ids; -1 if none yet
	normal     [3]float64
	offset     float64
	outside    []int
	furthest   float64 // distance of the last (furthest) outside point
	id         int
	toporient  bool
	visible    bool
	newfacet   bool
	visitid    int // qh_findhorizon BFS marker
	prev, next int // doubly-linked facet list (facet ids); sentinels at both ends
}

// hull is the incremental builder state, mirroring the subset of qhT the 2-D
// Delaunay build touches.
type hull struct {
	q        *qstate
	facets   []*hullFacet // indexed by facet id (append-only; visible ones flagged)
	headID   int          // facet_list head sentinel id
	tailID   int          // facet_tail sentinel id
	nextID   int          // facet_next (id) for the buildhull walk
	createID int          // next vertex creation id to assign (post-simplex)

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
// incremental hull, then fans each cocircular cell from its last-created vertex.
// It reproduces matplotlib's diagonal choice for general position and every
// cocircular case whose order is determined by the clean build; cases that need
// premerge's build-order reordering are not yet closed (see PLAN.md Phase 12).
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

// newFacet allocates a facet with the given descending-id vertex triple.
func (h *hull) newFacet(v [3]int, toporient bool) *hullFacet {
	f := &hullFacet{
		vertices:  v,
		neighbors: [3]int{-1, -1, -1},
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

// ---- geometry primitives (qh_distplane / qh_setfacetplane, hull_dim 3) ---

// distplane mirrors qh_distplane: signed distance of the lifted point pt to the
// facet's hyperplane (positive = outside/above).
func (h *hull) distplane(pt int, f *hullFacet) float64 {
	p := h.q.pts[pt]
	return f.offset + p[0]*f.normal[0] + p[1]*f.normal[1] + p[2]*f.normal[2]
}

// setFacetPlane mirrors qh_setfacetplane via qh_sethyperplane_det (hull_dim 3):
// the normal is the cofactor determinant of the vertex coordinate differences,
// normalised to unit length and oriented by toporient; offset = -point0·normal.
// It reports false if the determinant normal is degenerate (nearzero) — the
// Gaussian fallback is not yet ported, so such inputs bail out of the build.
func (h *hull) setFacetPlane(f *hullFacet) bool {
	r0, r1, r2 := h.q.pts[f.vertices[0]], h.q.pts[f.vertices[1]], h.q.pts[f.vertices[2]]
	// dX/dY/dZ(p,0) = r_p - r_0
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
	// nearzero check: any vertex far from its own plane => degenerate normal.
	for _, v := range f.vertices {
		if v == f.vertices[0] {
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

// ---- initial simplex (qh_createsimplex + qh_initialhull) -----------------

// initialHull builds the initial simplex from qh_maxsimplex and orients its
// facets so the interior point lies below each. It reports false on degeneracy.
func (h *hull) initialHull() bool {
	simplex := h.q.maxsimplex() // creation order [p0,p1,p2,p3]
	if len(simplex) != hullDim+1 {
		return false
	}
	// Sentinels bracket the facet list.
	head := h.newFacet([3]int{-1, -1, -1}, false)
	tail := h.newFacet([3]int{-1, -1, -1}, false)
	h.headID, h.tailID = head.id, tail.id
	head.next, tail.prev = tail.id, head.id
	head.prev, tail.next = -1, -1

	// Vertices appended front-first => descending creation id. simplex is ascending
	// id, so the vertex set is its reverse. Record the four simplex vertices in
	// creation order.
	for _, p := range simplex {
		h.recordVertex(p)
	}
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

	// One facet per omitted vertex, alternating toporient, vertices kept in
	// descending-id order.
	toporient := true
	simplexFacets := make([]*hullFacet, 0, hullDim+1)
	for i := range hullDim + 1 {
		var v [3]int
		k := 0
		for j := range hullDim + 1 {
			if j != i {
				v[k] = vset[j]
				k++
			}
		}
		f := h.newFacet(v, toporient)
		h.appendFacet(f)
		simplexFacets = append(simplexFacets, f)
		toporient = !toporient
	}
	// Fully-connected neighbours: createsimplex sets each facet's neighbours to all
	// the others, in facet order, then we re-key them to the opposite-vertex slot.
	for _, f := range simplexFacets {
		for _, g := range simplexFacets {
			if g.id == f.id {
				continue
			}
			h.setNeighborAcrossSharedEdge(f, g)
		}
	}
	// Planes + orientation (interior below all).
	for _, f := range simplexFacets {
		if !h.setFacetPlane(f) {
			return false
		}
	}
	for _, f := range simplexFacets {
		h.orientOutside(f)
	}
	h.nextID = h.facets[h.headID].next
	return true
}

// setNeighborAcrossSharedEdge records g as f's neighbour across the edge opposite
// the one vertex of f not shared with g (used for the simplex, where every pair of
// facets shares exactly hullDim-1 vertices).
func (h *hull) setNeighborAcrossSharedEdge(f, g *hullFacet) {
	for k := range 3 {
		if !contains3(g.vertices, f.vertices[k]) {
			f.neighbors[k] = g.id
			return
		}
	}
}

// ---- partition (qh_partitionall) ----------------------------------------

// partitionAll mirrors qh_partitionall for the (!BESToutside, MERGING,
// KEEPcoplanar) Delaunay config. Block 1 walks the facet list and assigns each
// not-yet-claimed point to the FIRST facet it is clearly outside of (distance >=
// distoutside), keeping that facet's running-furthest point deferred to the end of
// its outside set; points not clearly outside any facet fall through. Block 2
// re-partitions those leftovers with the best-facet search (qh_partitionpoint via
// qh_findbestnew, which with no new facets reduces to the best live facet). This
// greedy first-facet order — not a global furthest search — is what fixes the
// vertex creation order for cocircular inputs.
func (h *hull) partitionAll() {
	inSimplex := make([]bool, h.q.n+1)
	h.liveFacets(func(f *hullFacet) {
		for _, v := range f.vertices {
			inSimplex[v] = true
		}
	})
	// pointset: non-simplex real points in ascending id order (the infinity point
	// q.n is a simplex vertex, so it is excluded here).
	pointset := make([]int, 0, h.q.n)
	for p := 0; p <= h.q.n; p++ {
		if !inSimplex[p] {
			pointset = append(pointset, p)
		}
	}
	// distoutside = (USEfindbestnew?2:1) * max((MERGING?2:1)*MINoutside, max_outside).
	// At partition time no merges have happened (USEfindbestnew false) and
	// max_outside is 0, so distoutside = 2*MINoutside.
	distoutside := 2 * h.minOutside

	// Block 1: greedy first-facet assignment.
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
				f.outside = append(f.outside, bestpoint) // old best loses the last slot
				bestpoint, bestdist = p, d
			default:
				f.outside = append(f.outside, p)
			}
		}
		if bestpoint != -1 {
			f.outside = append(f.outside, bestpoint) // furthest last
			f.furthest = bestdist
		}
		pointset = leftover
	}

	// Block 2: leftover points re-partitioned by best-facet search.
	for _, p := range pointset {
		h.partitionPoint(p)
	}
}

// partitionPoint mirrors qh_partitionpoint over the LIVE facet list (the
// no-new-facets case used by qh_partitionall block 2): it finds the best facet for
// the point and, if the point is outside it (dist >= MINoutside), appends it
// keeping the furthest last, then — like Qhull — moves a freshly-occupied facet to
// the tail so it is processed after qh.facet_next.
func (h *hull) partitionPoint(p int) {
	f, dist := h.findBest(p)
	if f == nil || dist < h.minOutside {
		return // coplanar/inside: dropped by the clean build (no premerge)
	}
	wasEmpty := len(f.outside) == 0
	h.addOutside(f, p, dist)
	if wasEmpty && h.nextID != f.id {
		h.removeFacet(f)
		h.appendFacet(f)
	}
}

// findBest returns the live facet that the point is furthest above, and that
// distance. Qhull's qh_findbest is a directed search; the furthest facet it
// converges to equals the global maximum used here.
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

// findBestNew returns the new facet (one of cand) the point is furthest above.
func (h *hull) findBestNew(pt int, cand []*hullFacet) (*hullFacet, float64) {
	var best *hullFacet
	bestDist := -math.MaxFloat64
	for _, f := range cand {
		if f.visible {
			continue
		}
		if d := h.distplane(pt, f); d > bestDist {
			bestDist, best = d, f
		}
	}
	return best, bestDist
}

// addOutside inserts p into f's outside set keeping the furthest point last
// (qh_partitionpoint's append / append-2nd-last rule).
func (h *hull) addOutside(f *hullFacet, p int, dist float64) {
	if len(f.outside) == 0 || dist > f.furthest {
		f.outside = append(f.outside, p)
		f.furthest = dist
		return
	}
	// insert just before the current furthest (the last element)
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
// facet with a non-empty outside set (whose last element is its furthest point),
// without advancing facet_next past it — mirroring qh_nextfurthest on the default
// (no NARROWhull/RANDOMoutside) path.
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

// contains3 reports whether v appears in the triple.
func contains3(t [3]int, v int) bool { return t[0] == v || t[1] == v || t[2] == v }

// ---- addPoint (qh_addpoint: horizon, cone, match, partition, delete) ------

// addPoint inserts furthest (which lies above the seed facet) into the hull. It
// marks the visible facets, builds a cone of new facets from the apex to the
// horizon, links them, redistributes the orphaned outside points, and deletes the
// visible facets. The apex is the single new vertex, so the sequence of addPoint
// calls is the post-simplex vertex creation order.
func (h *hull) addPoint(furthest int, seed *hullFacet) bool {
	visible := h.findHorizon(furthest, seed)

	apex := furthest
	h.recordVertex(apex)

	// Cone: one new facet per horizon edge of each visible facet, in visible-list
	// then neighbour-slot order (qh_makenewfacets + qh_makenew_simplicial).
	newFacets := make([]*hullFacet, 0, len(visible)+2)
	for _, vf := range visible {
		for k := range 3 {
			nb := h.facets[vf.neighbors[k]]
			if nb.visible {
				continue
			}
			u, v := vf.vertices[(k+1)%3], vf.vertices[(k+2)%3]
			nf := h.makeConeFacet(apex, u, v, nb)
			if nf == nil {
				return false
			}
			// re-point the horizon neighbour's slot (that referenced vf) at nf
			for kk := range 3 {
				if nb.neighbors[kk] == vf.id {
					nb.neighbors[kk] = nf.id
					break
				}
			}
			newFacets = append(newFacets, nf)
		}
	}
	if len(newFacets) == 0 {
		return false
	}
	h.matchNewFacets(newFacets)

	h.partitionVisible(visible, newFacets)

	for _, vf := range visible {
		h.removeFacet(vf)
	}
	return h.ok
}

// findHorizon marks every facet the point is above (dist > MINvisible), starting
// from the seed and flooding across neighbours, and returns the visible facets in
// Qhull's visible-list order (qh_findhorizon).
func (h *hull) findHorizon(pt int, seed *hullFacet) []*hullFacet {
	h.visit++
	visit := h.visit
	seed.visible = true
	seed.visitid = visit
	visible := []*hullFacet{seed}
	for i := 0; i < len(visible); i++ {
		vf := visible[i]
		for k := range 3 {
			nb := h.facets[vf.neighbors[k]]
			if nb.visitid == visit {
				continue
			}
			nb.visitid = visit
			if h.distplane(pt, nb) > h.minVisible {
				nb.visible = true
				visible = append(visible, nb)
			}
		}
	}
	return visible
}

// makeConeFacet creates a new facet {apex, u, v} (vertices descending by creation
// id) whose edge opposite the apex faces the horizon neighbour nb, sets its plane,
// and orients it outward. Returns nil on a degeneracy that needs the unported
// Gaussian fallback.
func (h *hull) makeConeFacet(apex, u, v int, nb *hullFacet) *hullFacet {
	hi, lo := u, v
	if h.vid[lo] > h.vid[hi] {
		hi, lo = lo, hi
	}
	nf := h.newFacet([3]int{apex, hi, lo}, false)
	nf.newfacet = true
	nf.neighbors[0] = nb.id // opposite apex == edge (hi,lo) == shared with horizon
	if !h.setFacetPlane(nf) {
		h.ok = false
		return nil
	}
	h.orientOutside(nf)
	h.appendFacet(nf)
	return nf
}

// matchNewFacets links the cone facets to each other across their apex edges: two
// new facets sharing the apex edge (apex,w) are neighbours across the slot
// opposite the other (non-apex, non-w) vertex (qh_matchnewfacets, simplicial).
func (h *hull) matchNewFacets(newFacets []*hullFacet) {
	type slot struct {
		f *hullFacet
		k int
	}
	pending := map[int]slot{} // key: the shared non-apex vertex w
	for _, nf := range newFacets {
		// vertices = [apex, hi, lo]; apex edge (apex,hi) is opposite lo (slot 2),
		// apex edge (apex,lo) is opposite hi (slot 1).
		for _, e := range [2]struct {
			w, k int
		}{{nf.vertices[1], 2}, {nf.vertices[2], 1}} {
			if other, ok := pending[e.w]; ok {
				nf.neighbors[e.k] = other.f.id
				other.f.neighbors[other.k] = nf.id
				delete(pending, e.w)
			} else {
				pending[e.w] = slot{nf, e.k}
			}
		}
	}
}

// partitionVisible redistributes the outside points of the deleted visible facets
// onto the new cone facets (qh_partitionvisible), keeping each new facet's
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

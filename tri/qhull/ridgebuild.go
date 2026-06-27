package qhull

import (
	"math"
	"os"
	"strings"
)

// This file is the faithful re-port of Qhull's (8.0.2) 2-D Delaunay incremental
// build that supersedes the vertex-set model in buildhull.go. The decisive
// difference: a facet here keeps Qhull's own data layout — a vertex set kept
// INVERSE-SORTED by creation id, and a PARALLEL neighbour array (nbr[i] is the
// facet across the ridge opposite verts[i]) maintained incrementally — instead of
// re-deriving boundary edges and adjacency geometrically every step. Reproducing
// Qhull's neighbour iteration order (qh_findhorizon / qh_makenew_simplicial both
// walk visible->neighbors in that parallel, inverse-id order) is what makes the
// computed vertex CREATION ORDER match Qhull bit-for-bit, which in turn fixes every
// cocircular cell's fan apex (see fanfromorder.go) — the whole point of Phase 12.
//
// References: third_party/qhull-8.0.2/src/libqhull_r/{libqhull_r.c,poly_r.c,
// poly2_r.c,geom_r.c}. Key correspondences are cited inline.

// rfacet mirrors the subset of facetT the 2-D Delaunay build touches. For a
// simplicial facet (the steady state) verts has hull_dim entries kept descending
// by creation id, and nbr is parallel: nbr[i] is the neighbour across the ridge
// formed by every vertex except verts[i]. A facet turns non-simplicial only
// through the coplanarhorizon merge (ridges then becomes authoritative).
type rfacet struct {
	id    int
	verts []int // vertices, descending creation id (qh keeps SETfirst = newest)
	nbr   []int // parallel to verts: neighbour across ridge opposite verts[i]; -1 none

	// Non-simplicial facets (post-merge) carry an explicit, ordered ridge list:
	// redge[i] is the i-th boundary edge (2 vertices, descending id), rnbr[i] the
	// facet across it, and rtop[i] whether THIS facet is the ridge's top (the
	// orientation qh_makenew_nonsimplicial reads to orient new cone facets), in
	// Qhull ridge-creation order. Simplicial facets leave these nil and use the
	// parallel nbr array (redge[i] is implicitly verts\verts[i]).
	redge [][2]int
	rnbr  []int
	rtop  []bool

	// matRidge records the edges of a STILL-SIMPLICIAL facet that have a
	// materialised ridge (Qhull's visible->ridges populated as a side-effect of
	// qh_makeridges on a neighbour: poly2/merge_r.c:2136 appends the shared ridge to
	// BOTH facets). When such a facet becomes visible, qh_makenewfacets runs
	// qh_makenew_nonsimplicial over these ridges and its last cone (newfacet2) is the
	// replacement, even though the facet is still flagged simplicial. Edges are
	// stored as the cone edge would be, in appended (ridge-creation) order.
	matRidge [][2]int

	normal [3]float64
	offset float64

	outside  []int
	furthest float64

	simplicial bool
	toporient  bool
	upper      bool // upperdelaunay (faces away from the paraboloid)
	visible    bool
	newfacet   bool
	mergehz    bool // apex coplanar with this (horizon) facet → qh_mergecycle

	visitid    int
	prev, next int
}

// rhull is the incremental builder state (the slice of qhT the build reads/writes).
type rhull struct {
	q              *qstate
	facets         []*rfacet
	headID, tailID int
	nextID         int // qh.facet_next
	createID       int // next vertex creation id

	order []int
	vid   []int // vid[pt] = creation id, or -1

	interior    [3]float64
	minVisible  float64
	maxCoplanar float64
	minOutside  float64
	maxOutside  float64 // qh.max_outside (running, for qh_DISToutside)
	visit       int
	ok          bool

	replace   map[int]int // visible facet id → its replacement new facet id
	newListID int         // qh.newfacet_list head (current addPoint's first new facet)

	samecycle    map[int][]int // horizon facet id → cone facets to merge into it
	mergeHorizon []int         // horizon facet ids in first-encounter order
	mergeMarked  []int         // facets flagged mergehz this addPoint (to clear)
	mergeEnabled bool          // opt-in coplanar-horizon merge (WIP, Stage 3c.6)
}

// buildHullOrderRidge is the faithful analogue of buildHullOrder: it runs the
// incremental hull and returns the real-point vertex creation order. It reports
// false on a degeneracy that would need the unported Gaussian fallback.
func buildHullOrderRidge(q *qstate) ([]int, bool) {
	h := &rhull{q: q, replace: map[int]int{}}
	// The coplanar-horizon merge (linkSamecycle/premerge/mergeInto + explicit
	// redge/rnbr/rtop ridge lists, non-simplicial-horizon geometric orientation, and
	// a full-scan fallback for the post-merge directed search) is on by default and
	// strictly better than the simplicial-only build (33/34 vs 24/34 cocircular
	// exact-order; general stays 27/27). The replacement model is a FAITHFUL port
	// (verified against a QHATTACH oracle in third_party/qhull-8.0.2, gitignored,
	// dumping every REPL/CONE/MAKERIDGES/MERGEFACET/NSRIDGE), not a heuristic. Four
	// pieces:
	//  1. qh_makenewfacets' newfacet2 LEAK (poly2_r.c:2490-2515): newfacet/newfacet2
	//     reset only at function entry, not per visible facet, so a ridgeless facet
	//     inherits the most-recent prior facet's newfacet2 — see makeNewFacets' runNF2.
	//  2. qh_makeridges propagation (merge_r.c:2133+2136): each NEW ridge is appended
	//     to BOTH facets, so a still-simplicial neighbour of a merge gains a ridge
	//     (visible->ridges non-empty) — see propagateRidge, called per new ridge in
	//     makeRidges (covers both the horizon's and the cone's edges, per-merge).
	//  3. qh_mergesimplex ridge order (merge_r.c:3753) via qh_setdel SWAP-REMOVE
	//     (qset_r.c:346): the merged facet's ridge sequence — hence its last cone /
	//     newfacet2 — depends on swap-remove, see swapRemoveRidge in mergeInto.
	//  4. qh_makeridges lists a facet's PRE-EXISTING materialised ridges FIRST
	//     (merge_r.c:2090-2091, only missing boundary ridges appended) — see makeRidges
	//     honouring matRidge order.
	// Remaining 1 (grid5x4): a cross-addPoint merge-history divergence — Qhull's
	// horizon hf[15,4,0] pre-carries ridge [15,0] from an EARLIER addPoint's merge of
	// the facet across [15,0] (f11), which this engine's merge history does not
	// reproduce (its matRidge is empty at merge time), so the first cone differs. The
	// merge ORDER within a premerge already matches qh_mergecycle_all (facet-list /
	// cone-creation order) for single-cone samecycles; the gap is which facets merged
	// at earlier addPoints (coplanarity-decision fidelity), a deeper layer. 33/34 beats
	// the shipped vertex-set engine (buildhull.go, 31/34). QHULL_MERGE=0 falls back to
	// the simplicial engine. See PLAN.md Stage 3c.6.
	h.mergeEnabled = os.Getenv("QHULL_MERGE") != "0"
	h.minVisible = 2 * q.distRound // premerge_centrum, hull_dim<=3 with merging
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

// ---- vertex creation order ----------------------------------------------

func (h *rhull) recordVertex(pt int) {
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

// ---- facet list management (qh_appendfacet / qh_removefacet) -------------

func (h *rhull) newFacet(verts []int) *rfacet {
	f := &rfacet{
		id:         len(h.facets),
		verts:      append([]int(nil), verts...),
		simplicial: true,
		prev:       -1,
		next:       -1,
	}
	f.nbr = make([]int, len(verts))
	for i := range f.nbr {
		f.nbr[i] = -1
	}
	h.facets = append(h.facets, f)
	return f
}

func (h *rhull) appendFacet(f *rfacet) {
	tail := h.facets[h.tailID]
	if h.nextID == h.tailID {
		h.nextID = f.id // qh_appendfacet: facet_next at tail → new facet
	}
	prev := tail.prev
	h.facets[prev].next = f.id
	f.prev = prev
	f.next = h.tailID
	tail.prev = f.id
}

func (h *rhull) removeFacet(f *rfacet) {
	if h.nextID == f.id {
		h.nextID = f.next
	}
	h.facets[f.prev].next = f.next
	h.facets[f.next].prev = f.prev
}

// liveFacets calls fn for each non-visible facet in list order.
func (h *rhull) liveFacets(fn func(*rfacet)) {
	for id := h.facets[h.headID].next; id != h.tailID; id = h.facets[id].next {
		if !h.facets[id].visible {
			fn(h.facets[id])
		}
	}
}

// ---- geometry (qh_distplane / qh_setfacetplane, hull_dim 3) --------------

func (h *rhull) distplane(pt int, f *rfacet) float64 {
	p := h.q.pts[pt]
	return f.offset + p[0]*f.normal[0] + p[1]*f.normal[1] + p[2]*f.normal[2]
}

// setFacetPlane mirrors qh_setfacetplane → qh_sethyperplane_det (hull_dim 3): the
// raw normal is the cofactor determinant of the first three vertices' differences,
// normalised, and — crucially — its sign is set by the facet's combinatorial
// toporient (qh_normalize2 negates the normal when !toporient), NOT by a geometric
// interior-below test. Reproducing toporient exactly is what makes facets that
// contain the Qz infinity point (the upper-delaunay "ceiling") get the same
// orientation as Qhull, so a real outside point has the same dist sign and the
// qh_findbestnew scan reaches the same facet. offset = -verts[0]·normal. Reports
// false on a near-degenerate norm (Gaussian fallback unported). upperdelaunay is
// then classified by the sign of the last normal coordinate.
func (h *rhull) setFacetPlane(f *rfacet) bool {
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
	if !f.toporient { // qh_normalize2: negate normal when !toporient
		norm = -norm
	}
	n[0] /= norm
	n[1] /= norm
	n[2] /= norm
	f.normal = n
	f.offset = -(r0[0]*n[0] + r0[1]*n[1] + r0[2]*n[2])
	// qh_setfacetplane: upperdelaunay if last normal coord ≳ 0 (Qz: the lower hull
	// is the Delaunay triangulation; facets with the infinity point face up).
	f.upper = n[2] > -h.q.distRound
	return true
}

// ---- initial simplex (qh_initialvertices + qh_createsimplex) -------------

func (h *rhull) initialHull() bool {
	simplex := h.q.maxsimplex() // creation (append) order [p0,p1,p2,p3]
	if len(simplex) != hullDim+1 {
		return false
	}
	head := h.newFacet(nil)
	tail := h.newFacet(nil)
	h.headID, h.tailID = head.id, tail.id
	head.next, tail.prev = tail.id, head.id
	head.prev, tail.next = -1, -1

	for _, p := range simplex {
		h.recordVertex(p)
	}
	// qh_initialvertices: vertices inserted at position 0 → descending creation id.
	vset := []int{simplex[3], simplex[2], simplex[1], simplex[0]}

	for _, p := range simplex {
		h.interior[0] += h.q.pts[p][0]
		h.interior[1] += h.q.pts[p][1]
		h.interior[2] += h.q.pts[p][2]
	}
	h.interior[0] /= float64(hullDim + 1)
	h.interior[1] /= float64(hullDim + 1)
	h.interior[2] /= float64(hullDim + 1)

	// qh_createsimplex: facet_i omits vset[i]; toporient alternates from True;
	// neighbours = the other facets in list order, parallel to the (inverse-id)
	// vertex order.
	sf := make([]*rfacet, 0, hullDim+1)
	toporient := true
	for i := range hullDim + 1 {
		v := make([]int, 0, hullDim)
		for j := range hullDim + 1 {
			if j != i {
				v = append(v, vset[j])
			}
		}
		f := h.newFacet(v)
		f.toporient = toporient
		toporient = !toporient
		h.appendFacet(f)
		sf = append(sf, f)
	}
	for i, f := range sf {
		k := 0
		for j := range sf {
			if j != i {
				f.nbr[k] = sf[j].id
				k++
			}
		}
	}
	for _, f := range sf {
		if !h.setFacetPlane(f) {
			return false
		}
	}
	// qh_initialhull: if the first facet's oriented normal points toward the
	// interior point, the whole simplex is inward — flip every facet's toporient
	// together (keeping orientations consistent) and recompute the planes.
	if h.distplane2(h.interior, sf[0]) > h.q.distRound {
		for _, f := range sf {
			f.toporient = !f.toporient
			if !h.setFacetPlane(f) {
				return false
			}
		}
	}
	h.nextID = h.facets[h.headID].next
	return true
}

// distplane2 is qh_distplane for an explicit point (here the interior point).
func (h *rhull) distplane2(p [3]float64, f *rfacet) float64 {
	return f.offset + p[0]*f.normal[0] + p[1]*f.normal[1] + p[2]*f.normal[2]
}

// ---- partition (qh_partitionall + qh_partitionpoint) --------------------

func (h *rhull) partitionAll() {
	inSimplex := make([]bool, h.q.n+1)
	h.liveFacets(func(f *rfacet) {
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

func (h *rhull) findBestAll(pt int) (*rfacet, float64) {
	var best *rfacet
	bestDist := -math.MaxFloat64
	h.liveFacets(func(f *rfacet) {
		if d := h.distplane(pt, f); d > bestDist {
			bestDist, best = d, f
		}
	})
	return best, bestDist
}

func (h *rhull) partitionPoint(p int) {
	f, dist := h.findBestAll(p)
	if f == nil || dist < h.minOutside {
		return
	}
	wasEmpty := len(f.outside) == 0
	h.addOutside(f, p, dist)
	if wasEmpty && h.nextID != f.id {
		h.removeFacet(f)
		h.appendFacet(f)
	}
}

func (h *rhull) addOutside(f *rfacet, p int, dist float64) {
	// NB: qh.max_outside is NOT advanced for ordinary outside points
	// (qh_partitionpoint); only qh_partitioncoplanar / qh_findbesthorizon grow it.
	// So during a general-position build qh.max_outside ≈ 0 and qh_DISToutside stays
	// tiny, which is why qh_findbestnew returns the FIRST clearly-outside facet in
	// scan order rather than the globally furthest one.
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

// ---- build loop (qh_buildhull + qh_nextfurthest) ------------------------

func (h *rhull) buildLoop() bool {
	h.ok = true
	h.nextID = h.facets[h.headID].next
	h.furthestNext() // qh_initbuild seeding (PICKfurthest off → one-time)
	for {
		f := h.nextFurthest()
		if f == nil {
			break
		}
		furthest := f.outside[len(f.outside)-1]
		f.outside = f.outside[:len(f.outside)-1]
		if !h.addPoint(furthest, f) || !h.ok {
			return false
		}
	}
	// The coplanar-horizon merge (qh_mergecycle) is implemented but still drops
	// the coplanar interior points of a cocircular cell — it needs the
	// qh_partitioncoplanar promotion layer to keep them pickable (PLAN.md Stage
	// 3c.6). Until then a dropped point means the merge produced an invalid order,
	// so bail and let the caller fall back to the exact triangulation.
	if len(h.order) != h.q.n {
		return false
	}
	return true
}

func (h *rhull) furthestNext() {
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
		// qh_furthestnext prepends; emulate by inserting just after head.
		head := h.facets[h.headID]
		next := head.next
		head.next = f.id
		f.prev = h.headID
		f.next = next
		h.facets[next].prev = f.id
		h.nextID = f.id
	}
}

func (h *rhull) nextFurthest() *rfacet {
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

// ---- addPoint (qh_addpoint) ---------------------------------------------

func (h *rhull) addPoint(furthest int, seed *rfacet) bool {
	h.traceStep(furthest, seed)
	visible := h.findHorizon(furthest, seed)

	apex := furthest
	h.recordVertex(apex)

	newFacets, ok := h.makeNewFacets(apex, visible)
	if !ok {
		return false
	}
	if len(newFacets) == 0 {
		return false
	}
	h.matchNewFacets(newFacets)
	newFacets = h.premerge(newFacets) // qh_premerge → qh_mergecycle (coplanar horizon)
	if len(newFacets) == 0 {
		return false
	}
	h.newListID = newFacets[0].id // qh.newfacet_list head (first surviving new facet)
	h.partitionVisible(visible, newFacets)
	for _, vf := range visible {
		h.removeFacet(vf)
	}
	// qh_resetlists: clear newfacet on every facet from newfacet_list to the tail
	// (cone facets plus any horizon facet partitionPointInto moved there and
	// re-flagged). facet_next is NOT reset.
	for id := h.newListID; id != h.tailID; {
		f := h.facets[id]
		next := f.next
		f.newfacet = false
		id = next
	}
	// coplanarhorizon is per-addPoint state (qh_findhorizon re-marks each time):
	// clear it so a stale flag never triggers a merge in a later step.
	for _, id := range h.mergeMarked {
		h.facets[id].mergehz = false
	}
	h.mergeMarked = h.mergeMarked[:0]
	return h.ok
}

// findHorizon floods the visible facets from the seed (qh_findhorizon): each
// visible facet is moved to the tail of facet_list in BFS order, and non-visible
// neighbours coplanar with the apex are flagged mergehz (the coplanarhorizon
// merge). Neighbours are walked in stored (parallel, inverse-id) order.
func (h *rhull) findHorizon(pt int, seed *rfacet) []*rfacet {
	h.visit++
	visit := h.visit
	h.removeFacet(seed)
	h.appendFacet(seed)
	seed.visible = true
	seed.visitid = visit
	visible := []*rfacet{seed}
	for i := 0; i < len(visible); i++ {
		vf := visible[i]
		for _, r := range h.boundary(vf) {
			if r.nbr < 0 {
				continue
			}
			nb := h.facets[r.nbr]
			if nb.visitid == visit {
				continue
			}
			nb.visitid = visit
			d := h.distplane(pt, nb)
			switch {
			case d > h.minVisible:
				h.removeFacet(nb)
				h.appendFacet(nb)
				nb.visible = true
				visible = append(visible, nb)
			case d >= -h.maxCoplanar && h.mergeEnabled:
				nb.mergehz = true
				h.mergeMarked = append(h.mergeMarked, nb.id)
			}
		}
	}
	return visible
}

// ridgeRef is one boundary ridge of a facet during iteration: its two vertices
// (descending creation id), the facet across it, and the slot index into the
// facet's neighbour array.
type ridgeRef struct {
	edge [2]int
	nbr  int
	slot int
	top  bool // whether f is this ridge's top (qh_makeridges orientation)
}

// boundary returns f's boundary ridges in Qhull iteration order: parallel
// (inverse-id vertex) order for a simplicial facet — ridge k is the edge opposite
// verts[k] — and the explicit ordered ridge list for a merged (non-simplicial) one.
// top mirrors qh_makeridges: a simplicial facet is its ridge k's top iff
// toporient ^ (k odd).
func (h *rhull) boundary(f *rfacet) []ridgeRef {
	if !f.simplicial {
		out := make([]ridgeRef, len(f.rnbr))
		for i := range f.rnbr {
			out[i] = ridgeRef{f.redge[i], f.rnbr[i], i, f.rtop[i]}
		}
		return out
	}
	n := len(f.verts)
	out := make([]ridgeRef, n)
	for k := range n {
		var e [2]int
		j := 0
		for m := range n {
			if m != k {
				e[j] = f.verts[m]
				j++
			}
		}
		out[k] = ridgeRef{e, f.nbr[k], k, f.toporient != (k&0x1 == 1)}
	}
	return out
}

// findSlot returns the boundary slot of facet f whose neighbour is nbID, or -1.
func (h *rhull) findSlot(f *rfacet, nbID int) int {
	arr := f.nbr
	if !f.simplicial {
		arr = f.rnbr
	}
	for s, x := range arr {
		if x == nbID {
			return s
		}
	}
	return -1
}

// setNbrSlot relinks slot s of facet f to neighbour nbID.
func (h *rhull) setNbrSlot(f *rfacet, s, nbID int) {
	if f.simplicial {
		f.nbr[s] = nbID
	} else {
		f.rnbr[s] = nbID
	}
}

// makeNewFacets builds the cone of new facets from the apex over the horizon
// (qh_makenewfacets → qh_makenew_simplicial). For each visible facet in BFS order,
// each neighbour that is a (non-visible) horizon facet yields one new facet
// {apex} ∪ ridge, appended in that exact order; the horizon facet's slot that
// pointed to the visible facet is relinked to the new facet. Non-simplicial
// visible facets (post-merge) go through the ridge walk.
func (h *rhull) makeNewFacets(apex int, visible []*rfacet) ([]*rfacet, bool) {
	newFacets := make([]*rfacet, 0, len(visible)+2)
	// Faithful qh_makenewfacets replacement (poly2_r.c:2490-2515). newfacet/newfacet2
	// are declared ONCE and reset to NULL only at function entry — NOT inside the
	// per-visible-facet loop — so a ridgeless facet inherits the most recent prior
	// facet's newfacet2 ("if (newfacet2) newfacet= newfacet2"). We reproduce that
	// leak with a running newfacet2 (runNF2): for each visible facet, newfacet2 is
	// recomputed (to its last nonsimplicial ridge cone, possibly none) iff it has
	// ridges, and otherwise persists; visible->f.replace = newfacet2 if set, else the
	// facet's own last simplicial cone. A facet "has ridges" if it is non-simplicial
	// (all edges) or carries propagated matRidge edges (a still-simplicial neighbour
	// of an earlier merge). General position: no ridges ever ⇒ each facet's own last
	// cone (27/27 preserved).
	runNF2 := -1
	for _, vf := range visible {
		lastRidge, lastCone := -1, -1
		vfHasRidges := !vf.simplicial || len(vf.matRidge) > 0
		for _, r := range h.boundary(vf) {
			if r.nbr < 0 {
				continue
			}
			nb := h.facets[r.nbr]
			if nb.visible {
				continue
			}
			// new facet = apex prepended to the shared ridge vertices (both
			// inverse-id sorted, apex highest ⇒ stays inverse-id sorted).
			nf := h.newFacet([]int{apex, r.edge[0], r.edge[1]})
			nf.newfacet = true
			horizonskip := h.findSlot(nb, vf.id)
			if vf.simplicial {
				// qh_makenew_simplicial toporient parity (from the horizon facet).
				if nb.toporient {
					nf.toporient = horizonskip&0x1 == 1
				} else {
					nf.toporient = horizonskip&0x1 == 0
				}
			} else {
				// qh_makenew_nonsimplicial: orient from the ridge (top == visible).
				nf.toporient = r.top
			}
			if !h.setFacetPlane(nf) {
				h.ok = false
				return nil, false
			}
			if !nb.simplicial {
				// horizonskip parity (qh_makenew_simplicial) indexes the horizon's
				// neighbour list, but a merged horizon's neighbour order does not yet
				// match Qhull's qh_mergesimplex order, so the parity is unreliable.
				// Non-simplicial horizons occur only in cocircular merge cases (never
				// in general position), so orient these cone facets geometrically
				// (interior below) without touching the proven simplicial path.
				if h.distplane2(h.interior, nf) > 0 {
					nf.toporient = !nf.toporient
					nf.normal[0], nf.normal[1], nf.normal[2] = -nf.normal[0], -nf.normal[1], -nf.normal[2]
					nf.offset = -nf.offset
					nf.upper = nf.normal[2] > -h.q.distRound
				}
			}
			h.appendFacet(nf)
			nf.nbr[0] = nb.id                    // neighbour opposite apex = horizon facet
			h.setNbrSlot(nb, horizonskip, nf.id) // relink horizon's slot (was vf)
			if nb.mergehz {
				nf.mergehz = true
				h.linkSamecycle(nb, nf)
			}
			// qh_makenew_nonsimplicial builds a cone for each materialised ridge (a
			// non-simplicial facet's full boundary, or a simplicial facet's matRidge
			// edges) and returns the last; qh_makenew_simplicial returns the last cone
			// overall. Track both for this facet.
			if !vf.simplicial || edgeInList(r.edge, vf.matRidge) {
				lastRidge = nf.id
			}
			lastCone = nf.id
			newFacets = append(newFacets, nf)
		}
		// newfacet2 is recomputed iff vf has ridges (to lastRidge, maybe -1); replace
		// = newfacet2 when set, else this facet's own last (simplicial) cone.
		if vfHasRidges {
			runNF2 = lastRidge
		}
		if runNF2 >= 0 {
			h.replace[vf.id] = runNF2
		} else if lastCone >= 0 {
			h.replace[vf.id] = lastCone
		}
	}
	return newFacets, true
}

// linkSamecycle records a coplanar-horizon cone facet for the merge into its
// horizon facet (qh_makenew_simplicial's samecycle/mergehorizon flagging). Cone
// facets are grouped by horizon facet in first-encounter order.
func (h *rhull) linkSamecycle(horizon, cone *rfacet) {
	if h.samecycle == nil {
		h.samecycle = map[int][]int{}
	}
	if _, seen := h.samecycle[horizon.id]; !seen {
		h.mergeHorizon = append(h.mergeHorizon, horizon.id)
	}
	h.samecycle[horizon.id] = append(h.samecycle[horizon.id], cone.id)
}

// makeRidges materialises a simplicial facet's implicit ridges into the explicit
// ordered redge/rnbr lists (qh_makeridges). Pre-existing materialised ridges
// (f.matRidge, propagated from an earlier neighbour merge) come FIRST in their
// recorded order: qh_makeridges (merge_r.c:2090-2091) marks neighbours that already
// have a ridge as "seen" and only creates+APPENDS the missing boundary ridges. This
// ordering is load-bearing — combined with the swap-remove in mergeInto it sets the
// merged facet's final ridge sequence, hence qh_makenew_nonsimplicial's last cone.
func (h *rhull) makeRidges(f *rfacet) {
	if !f.simplicial {
		return
	}
	b := h.boundary(f)
	used := make([]bool, len(b))
	f.redge = make([][2]int, 0, len(b))
	f.rnbr = make([]int, 0, len(b))
	f.rtop = make([]bool, 0, len(b))
	add := func(r ridgeRef) {
		f.redge = append(f.redge, r.edge)
		f.rnbr = append(f.rnbr, r.nbr)
		f.rtop = append(f.rtop, r.top)
	}
	for _, m := range f.matRidge { // pre-existing ridges first, in recorded order
		for i, r := range b {
			if !used[i] && sameEdge(r.edge, m) {
				add(r)
				used[i] = true
				break
			}
		}
	}
	for i, r := range b { // then the remaining (newly materialised) boundary ridges
		if !used[i] {
			add(r)
			// qh_makeridges appends each NEW ridge to BOTH facets (merge_r.c:2133+2136):
			// propagate it onto the still-simplicial neighbour across it, so when that
			// neighbour later turns visible/merges it carries a materialised ridge.
			h.propagateRidge(r.nbr, r.edge)
		}
	}
	f.simplicial = false
	f.matRidge = nil
}

// premerge folds every coplanar-horizon cone facet into its horizon facet
// (qh_premerge → qh_mergecycle): the horizon absorbs the apex, becomes
// non-simplicial, keeps its plane, and moves to the tail as a new facet; the cone
// facets are deleted. It returns the surviving new-facet set (non-merged cone
// facets followed by the merged horizons) for partitioning.
func (h *rhull) premerge(coneFacets []*rfacet) []*rfacet {
	if len(h.mergeHorizon) == 0 {
		return coneFacets
	}
	merged := map[int]bool{}
	for _, hid := range h.mergeHorizon {
		cones := h.samecycle[hid]
		h.mergeInto(h.facets[hid], cones)
		for _, c := range cones {
			merged[c] = true
		}
	}
	out := make([]*rfacet, 0, len(coneFacets))
	for _, nf := range coneFacets {
		if !merged[nf.id] {
			out = append(out, nf)
		}
	}
	for _, hid := range h.mergeHorizon {
		out = append(out, h.facets[hid])
	}
	h.samecycle = nil
	h.mergeHorizon = h.mergeHorizon[:0]
	return out
}

// mergeInto merges the samecycle cone facets into horizon facet H
// (qh_mergecycle). H's ridges are materialised; the ridge each cone shares with H
// is dropped; each cone's two apex ridges are appended to H (retargeting the
// sibling's neighbour slot to H), unless the sibling is itself merging into H, in
// which case the ridge is internal and dropped. The apex is prepended to H's
// vertices, interior vertices (now bordering only H) are removed, H keeps its
// plane and moves to the tail as a non-simplicial new facet, and the cone facets
// are deleted.

func (h *rhull) mergeInto(hf *rfacet, cones []int) {
	apex := h.facets[cones[0]].verts[0]
	// qh_mergefacet materialises facet1 (the cone) BEFORE facet2 (the horizon):
	// merge_r.c:3464-3465 calls qh_makeridges(facet1) then qh_makeridges(facet2). So
	// each cone's makeRidges runs first, propagating its base edge onto hf (still
	// simplicial), and hf's own makeRidges then lists that shared edge FIRST (its
	// pre-existing matRidge) — which the swap-remove below turns into the correct
	// final ridge order. Doing hf first would miss this and mis-order the last cone.
	for _, cID := range cones {
		h.makeRidges(h.facets[cID])
	}
	h.makeRidges(hf)
	// Faithful qh_mergesimplex replay (merge_r.c:3753): each cone (facet1) merges
	// into hf (facet2) in samecycle order. For each of the cone's three ridges (in
	// skip/boundary order — base [e_hi,e_lo], then [apex,e_lo], then [apex,e_hi]):
	// a ridge already shared with hf (the base, or an apex edge whose sibling cone
	// already merged) is DELETED via qh_delridge_merge — a swap-remove (qset_r.c:346
	// moves the last element into the slot); every other ridge is APPENDED to hf.
	// Reproducing the swap-remove order is what makes the merged facet's ridge list —
	// and hence qh_makenew_nonsimplicial's last cone (newfacet2) — match Qhull.
	for _, cID := range cones {
		c := h.facets[cID]
		for i := range c.redge {
			edge, sib, top := c.redge[i], c.rnbr[i], c.rtop[i]
			if h.swapRemoveRidge(hf, edge) {
				continue // ridge shared with hf (base, or interior cone-cone) → delete
			}
			hf.redge = append(hf.redge, edge)
			hf.rnbr = append(hf.rnbr, sib)
			hf.rtop = append(hf.rtop, top)
			if sib >= 0 {
				if sl := h.findSlot(h.facets[sib], c.id); sl >= 0 {
					h.setNbrSlot(h.facets[sib], sl, hf.id)
				}
			}
		}
		h.removeFacet(c)
		c.visible = true
		h.replace[c.id] = hf.id // qh_getreplacement chains merged cone → horizon
	}
	hf.verts = append([]int{apex}, hf.verts...)
	onBoundary := map[int]bool{}
	for _, e := range hf.redge {
		onBoundary[e[0]] = true
		onBoundary[e[1]] = true
	}
	kept := hf.verts[:0]
	for _, v := range hf.verts {
		if onBoundary[v] {
			kept = append(kept, v)
		}
	}
	hf.verts = kept
	h.removeFacet(hf)
	h.appendFacet(hf)
	hf.newfacet = true
	hf.simplicial = false
}

// propagateRidge mirrors qh_makeridges appending a newly materialised ridge to the
// facet on the other side (merge_r.c:2136): a still-simplicial neighbour records the
// shared edge in matRidge, so when it later turns visible qh_makenewfacets takes the
// nonsimplicial path (replacement = the cone over this edge, newfacet2) and when it
// later becomes a horizon makeRidges lists this ridge first.
func (h *rhull) propagateRidge(nbID int, edge [2]int) {
	if nbID < 0 {
		return
	}
	nb := h.facets[nbID]
	if nb.visible || !nb.simplicial {
		return
	}
	if !edgeInList(edge, nb.matRidge) {
		nb.matRidge = append(nb.matRidge, edge)
	}
}

// swapRemoveRidge removes the ridge with the given edge from hf's parallel ridge
// lists using Qhull's qh_setdel swap-remove (the last element fills the gap), and
// reports whether a ridge was found. The swap-remove order is load-bearing: it
// determines the merged facet's final ridge sequence (qh_makenew_nonsimplicial reads
// the last ridge as newfacet2).
func (h *rhull) swapRemoveRidge(hf *rfacet, edge [2]int) bool {
	for i := range hf.redge {
		if !sameEdge(hf.redge[i], edge) {
			continue
		}
		last := len(hf.redge) - 1
		hf.redge[i] = hf.redge[last]
		hf.rnbr[i] = hf.rnbr[last]
		hf.rtop[i] = hf.rtop[last]
		hf.redge = hf.redge[:last]
		hf.rnbr = hf.rnbr[:last]
		hf.rtop = hf.rtop[:last]
		return true
	}
	return false
}

// sameEdge reports whether two boundary edges are the same unordered vertex pair.
func sameEdge(a, b [2]int) bool {
	return (a[0] == b[0] && a[1] == b[1]) || (a[0] == b[1] && a[1] == b[0])
}

// edgeInList reports whether e (as an unordered pair) is already in list.
func edgeInList(e [2]int, list [][2]int) bool {
	for _, x := range list {
		if sameEdge(e, x) {
			return true
		}
	}
	return false
}

// matchNewFacets links the cone's sibling facets across their shared {apex, x}
// ridges (qh_matchnewfacets). Each new simplicial facet {apex, hi, lo} has two
// open ridges — {apex, lo} opposite verts[1]=hi, and {apex, hi} opposite
// verts[2]=lo — and each such edge is shared by exactly two cone facets.
func (h *rhull) matchNewFacets(newFacets []*rfacet) {
	type slot struct{ f, k int }
	open := map[int]slot{}
	for _, nf := range newFacets {
		if !nf.simplicial {
			continue
		}
		for k := 1; k < len(nf.verts); k++ {
			other := nf.verts[len(nf.verts)-k] // k=1→verts[2], k=2→verts[1]
			if o, ok := open[other]; ok {
				nf.nbr[k] = o.f
				h.facets[o.f].nbr[o.k] = nf.id
				delete(open, other)
			} else {
				open[other] = slot{nf.id, k}
			}
		}
	}
}

// findBest mirrors qh_findbest for the qh_partitionvisible call with the
// non-merge configuration (qh_USEfindbestnew = Ztotmerge>50 is false for these
// inputs, so qh_partitionpoint uses qh_findbest, NOT qh_findbestnew): a directed
// greedy walk over the NEW facets starting at startfacet (the replacement). At
// each facet it tests new-facet neighbours; the moment one is clearly outside and
// lower (dist ≥ MINoutside, !upperdelaunay — noupper is set, so upper facets never
// trigger the early return) it is returned; otherwise it switches to the first
// strictly-further lower neighbour and walks on. Upper facets are only adopted as
// a fallback bestfacet when no lower one is available. This is why an outside
// point reaches the correct lower facet instead of a coincidentally-closer upper
// (infinity) facet. Returns the best facet and its distance (the caller drops it
// if dist < MINoutside, i.e. coplanar/inside).
func (h *rhull) findBest(pt, startID int) (*rfacet, float64) {
	h.visit++
	visit := h.visit
	start := h.facets[startID]
	d := h.distplane(pt, start)
	// noupper = !qh_NOupper = False, so a clearly-outside facet of EITHER kind
	// triggers the early return; the bestfacet running value, however, is only set
	// to a lower facet (an upper one is adopted only as a last resort during the
	// directed walk's switch test).
	if d >= h.minOutside { // clearly outside (upper or lower)
		return start, d
	}
	bestDist := d
	var best *rfacet
	if !start.upper {
		best = start
	}
	start.visitid = visit
	facet := start
	for facet != nil {
		var next *rfacet
		for _, r := range h.boundary(facet) { // boundary handles merged (rnbr) facets
			nbID := r.nbr
			if nbID < 0 {
				continue
			}
			nb := h.facets[nbID]
			if !nb.newfacet || nb.visitid == visit {
				continue
			}
			nb.visitid = visit
			d := h.distplane(pt, nb)
			if d > bestDist {
				if d >= h.minOutside {
					return nb, d
				}
				if !nb.upper {
					best, bestDist, next = nb, d, nb
					break
				} else if best == nil {
					bestDist, next = d, nb
					break
				}
			}
		}
		facet = next
	}
	// testhorizon: qh_findbest always finishes with qh_findbesthorizon (unless it
	// early-returned a clearly-outside facet above), which broadens the search to
	// the bestfacet's full neighbourhood — including OLD (horizon) facets — so a
	// point that is not outside any new cone facet but is still outside a surviving
	// horizon facet is assigned there instead of being dropped.
	if best == nil {
		best = start
	}
	return h.findBestHorizon(pt, best, bestDist)
}

// findBestHorizon mirrors qh_findbesthorizon (non-checkmax, noupper): from the
// directed search's best facet it explores the facet graph (new AND old facets),
// moving to any neighbour strictly further from the point and exploring every
// neighbour within qh_SEARCHdist of the current best, returning the furthest lower
// facet found. With qh.max_outside ≈ 0 the search distance is tiny, so this is a
// hill-climb to the locally furthest facet.
func (h *rhull) findBestHorizon(pt int, start *rfacet, bestDist float64) (*rfacet, float64) {
	h.visit++
	visit := h.visit
	searchdist := 2 * math.Max(2*h.minOutside, h.maxOutside) // qh_SEARCHdist
	minsearch := bestDist - searchdist
	best := start
	start.visitid = visit
	queue := []*rfacet{start}
	for len(queue) > 0 {
		f := queue[0]
		queue = queue[1:]
		for _, r := range h.boundary(f) { // boundary handles merged (rnbr) facets
			nbID := r.nbr
			if nbID < 0 {
				continue
			}
			nb := h.facets[nbID]
			if nb.visitid == visit || nb.visible {
				continue
			}
			nb.visitid = visit
			d := h.distplane(pt, nb)
			if d > bestDist {
				// eligibility: !upper || (!noupper && dist≥MINoutside); noupper=False.
				if !nb.upper || d >= h.minOutside {
					minsearch = d - searchdist
					best, bestDist = nb, d
				}
			}
			if d >= minsearch {
				queue = append(queue, nb)
			}
		}
	}
	return best, bestDist
}

// partitionVisible redistributes each visible facet's orphaned outside points onto
// the new facets (qh_partitionvisible), starting each point's directed best search
// from the visible facet's replacement (qh_getreplacement). qh.max_outside stays at
// roundoff throughout these builds, so qh_findbestnew (used on merge steps) and
// qh_findbest agree — both return the replacement when the point is clearly outside
// it — so the only thing that distinguishes a merge step is the replacement choice.
func (h *rhull) partitionVisible(visible, newFacets []*rfacet) {
	for _, vf := range visible {
		start, ok := h.replace[vf.id]
		if !ok {
			start = newFacets[0].id
		}
		// qh_getreplacement: chain through replaced/merged (now-visible) facets.
		for h.facets[start].visible {
			next, ok := h.replace[start]
			if !ok || next == start {
				start = newFacets[0].id
				break
			}
			start = next
		}
		for _, p := range vf.outside {
			h.partitionPointInto(p, start)
		}
		vf.outside = nil
	}
}

// partitionPointInto mirrors qh_partitionpoint for the build (qh_findbest start):
// it finds the best facet for p, adds it keeping the furthest last, and — when the
// facet becomes newly outside (its outside set was empty) — applies Qhull's
// facet_next bookkeeping so the newly-outside facet is processed in the right
// order: a new cone facet resets facet_next to the head of the new facets (only if
// facet_next is itself a new facet); an old (horizon) facet is moved to the tail
// and re-flagged new. This ordering is what defers the Qz infinity point's cone
// (and points that land on it) past the lower-facet picks.
func (h *rhull) partitionPointInto(p, startID int) {
	f, dist := h.findBest(p, startID)
	if (f == nil || dist < h.minOutside) && h.mergeEnabled {
		// After a merge the directed walk may miss the facet a point is clearly
		// outside of; fall back to a full scan so it is not spuriously dropped.
		if f2, d2 := h.findBestAll(p); f2 != nil && d2 >= h.minOutside {
			f, dist = f2, d2
		}
	}
	if f == nil || dist < h.minOutside {
		return // coplanar/inside (qh_partitioncoplanar; dropped here)
	}
	isnewoutside := len(f.outside) == 0
	h.addOutside(f, p, dist)
	if isnewoutside && h.nextID != f.id {
		if f.newfacet {
			if h.nextID != h.tailID && h.facets[h.nextID].newfacet {
				h.nextID = h.newListID
			}
		} else {
			h.removeFacet(f)
			h.appendFacet(f)
			if h.newListID != h.tailID {
				f.newfacet = true
			}
		}
	}
}

// ---- per-step trace (diff against third_party/qhull-8.0.2 stepdump) ------

// traceStep dumps the lower facets (those whose outward normal faces the
// paraboloid, normal[2] < 0) in facet-list order with their outside sets, in the
// same shape as stepdump.c's QHSTEP dump, when QHULL_TRACE matches the case. Pure
// dev scaffolding; removed once 61/61 is locked.
func (h *rhull) traceStep(furthest int, seed *rfacet) {
	want := os.Getenv("QHULL_TRACE")
	if want == "" {
		return
	}
	var b strings.Builder
	b.WriteString("STEP pick p")
	b.WriteString(itoa(furthest))
	b.WriteString(" from f")
	b.WriteString(itoa(seed.id))
	b.WriteByte('\n')
	for id := h.facets[h.headID].next; id != h.tailID; id = h.facets[id].next {
		f := h.facets[id]
		if f.visible || f.upper {
			continue
		}
		b.WriteString("  f")
		b.WriteString(itoa(f.id))
		b.WriteString(" verts[")
		for _, v := range f.verts {
			b.WriteString(itoa(v))
			b.WriteByte(',')
		}
		b.WriteString("] outside[")
		for _, p := range f.outside {
			b.WriteString(itoa(p))
			b.WriteByte(',')
		}
		b.WriteString("]\n")
	}
	os.Stderr.WriteString(b.String())
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

package qhull

import "testing"

// computedCocircularRatchet is the minimum number of cocircular corpus cases the
// fully self-contained engine (incremental hull order + cocircular fan, no Qhull
// and no captured fixture) must reproduce. The faithful build — greedy first-facet
// partition (qh_partitionall), one-time furthest-facet seeding (qh_furthestnext),
// the non-simplicial coplanarhorizon merge (qh_premerge → qh_mergecycle), the
// first-clearly-outside qh_findbestnew, and Qhull's vertex-id cone/horizon
// iteration order — closes 31/34. The remaining 3 (grid5x4, rings_1.0_2.0_6,
// rings_1.0_2.0_8) diverge only in the last 2-3 insertions: tie-breaking among
// nearly-equidistant cocircular points in highly symmetric multi-ring/large-grid
// configs, which needs finer merge-to-tail ordering than the vertex-set facet
// model reproduces (PLAN.md Phase 12, Stage 3b). Bump as cases close; never lower.
const computedCocircularRatchet = 31

// TestDelaunayComputed exercises delaunayComputed end-to-end against the
// differential corpus. General position has a unique Delaunay, so it is a hard
// gate (the creation order is irrelevant there); cocircular cases are gated by the
// ratchet until the premerge build-order port lands.
func TestDelaunayComputed(t *testing.T) {
	c := loadCorpus(t)
	byCat := map[string]*struct{ pass, total int }{}
	var built, degenerate int
	for _, tc := range c.Cases {
		st := byCat[tc.Category]
		if st == nil {
			st = &struct{ pass, total int }{}
			byCat[tc.Category] = st
		}
		st.total++

		got, nbr, err := delaunayComputed(tc.X, tc.Y)
		if err != nil {
			degenerate++
			continue
		}
		built++
		if len(got) == len(tc.Triangles) &&
			sameTriangleSet(got, tc.Triangles) &&
			sameNeighborGraph(got, nbr, tc.Triangles, tc.Neighbors) {
			st.pass++
		}
	}
	t.Logf("built %d/%d cases (%d degenerate)", built, built+degenerate, degenerate)
	if degenerate != 0 {
		t.Errorf("incremental hull hit %d degenerate cases (Gaussian fallback unported)", degenerate)
	}

	if gen := byCat["general"]; gen != nil {
		t.Logf("category general   : %d/%d", gen.pass, gen.total)
		if gen.pass != gen.total {
			t.Errorf("general position must match Qhull exactly: %d/%d", gen.pass, gen.total)
		}
	}
	if co := byCat["cocircular"]; co != nil {
		t.Logf("category cocircular: %d/%d (ratchet %d, target %d)", co.pass, co.total, computedCocircularRatchet, co.total)
		if co.pass < computedCocircularRatchet {
			t.Errorf("cocircular regressed below ratchet: %d/%d < %d", co.pass, co.total, computedCocircularRatchet)
		}
	}
}

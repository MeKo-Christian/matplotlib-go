package qhull

import "testing"

// computedCocircularRatchet is the minimum number of cocircular corpus cases the
// fully self-contained engine (incremental hull order + cocircular fan, no Qhull
// and no captured fixture) must reproduce. delaunayComputed now drives the faithful
// ridge-graph engine (buildHullOrderRidge): Qhull's own facet layout (inverse-id
// vertex set + parallel neighbour array) with the faithful coplanarhorizon merge
// (newfacet2 leak + per-merge qh_makeridges propagation + qh_mergesimplex
// swap-remove + pre-existing-ridge ordering) closes 33/34. The remaining 1
// (grid5x4) diverges only in the intermediate ridge order of one merged quad
// between its creation and the Qz infinity-point processing — a deep, cosmetic-only
// fidelity layer (PLAN.md Phase 12, Stage 3c.6f). Bump as cases close; never lower.
const computedCocircularRatchet = 33

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

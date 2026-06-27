package qhull

import (
	"os"
	"testing"
)

// TestComputedOrderMatchesGroundTruth is the tight feedback loop for the
// Qhull-faithful build-order port (PLAN.md Phase 12, Stage 3c). The fan model is
// proven (TestFanFromGroundTruthOrder: 34/34 + 27/27 from Qhull's CAPTURED order),
// so the only thing standing between the computed engine and full parity is
// computing the vertex creation order bit-for-bit. This test compares the computed
// buildHullOrder against the captured creation_order.json, exactly, per case.
//
// It is intentionally NOT a hard gate during the port: it logs the per-category
// exact-order match counts so progress is measurable. Set QHULL_ORDER_STRICT=1 to
// make divergences fail (used once 61/61 is reached, to lock it in).
//
// Note on general position: a unique Delaunay means a divergent build order still
// yields the correct triangle SET, so general-position exact-order match can lag
// without affecting TestDelaunayComputed. Cocircular cases are where exact order is
// load-bearing — those must reach 34/34 here to close the phase.
func TestComputedOrderMatchesGroundTruth(t *testing.T) {
	corp := loadCorpus(t)
	orders := loadCreationOrders(t)
	strict := os.Getenv("QHULL_ORDER_STRICT") == "1"

	byCat := map[string]*struct{ pass, total int }{}
	var diverged []string
	for _, tc := range corp.Cases {
		st := byCat[tc.Category]
		if st == nil {
			st = &struct{ pass, total int }{}
			byCat[tc.Category] = st
		}
		st.total++

		want, ok := orders.Order[tc.Name]
		if !ok {
			t.Fatalf("%s: no captured creation order", tc.Name)
		}
		got, built := buildHullOrder(project(tc.X, tc.Y))
		if built && sameIntSlice(got, want) {
			st.pass++
		} else {
			diverged = append(diverged, tc.Name)
		}
	}

	for _, cat := range []string{"general", "cocircular"} {
		if st := byCat[cat]; st != nil {
			t.Logf("category %-10s: %d/%d exact-order match", cat, st.pass, st.total)
		}
	}
	if len(diverged) > 0 {
		t.Logf("order divergences (%d): %v", len(diverged), diverged)
		if strict {
			t.Errorf("computed build order diverged from Qhull in %d cases", len(diverged))
		}
	}
}

// TestComputedOrderRidge measures the faithful ridge-graph engine
// (buildHullOrderRidge) against the captured ground truth, the same way as the
// vertex-set engine above. This is the development gate for Stage 3c; it logs
// per-category exact-order match counts and the divergent cases.
func TestComputedOrderRidge(t *testing.T) {
	corp := loadCorpus(t)
	orders := loadCreationOrders(t)

	byCat := map[string]*struct{ pass, total int }{}
	var diverged []string
	for _, tc := range corp.Cases {
		st := byCat[tc.Category]
		if st == nil {
			st = &struct{ pass, total int }{}
			byCat[tc.Category] = st
		}
		st.total++
		want := orders.Order[tc.Name]
		got, built := buildHullOrderRidge(project(tc.X, tc.Y))
		if built && sameIntSlice(got, want) {
			st.pass++
		} else {
			diverged = append(diverged, tc.Name)
		}
	}
	for _, cat := range []string{"general", "cocircular"} {
		if st := byCat[cat]; st != nil {
			t.Logf("ridge %-10s: %d/%d exact-order match", cat, st.pass, st.total)
		}
	}
	if len(diverged) > 0 {
		t.Logf("ridge divergences (%d): %v", len(diverged), diverged)
	}
	// General position has a unique build with no merges; the faithful engine must
	// reproduce its creation order exactly. Hard-gate it to prevent regression
	// while the cocircular merge path (Stage 3c.6) is completed.
	if st := byCat["general"]; st != nil && st.pass != st.total {
		t.Errorf("ridge general exact-order regressed: %d/%d (must be %d)", st.pass, st.total, st.total)
	}
}

// sameIntSlice reports whether a and b are element-wise equal.
func sameIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

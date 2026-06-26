package qhull

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// creationOrders is the ground-truth vertex creation order captured from the
// vendored Qhull 8.0.2 via the introspect tool (testdata/gen_creation_order.py):
// order[name][k] is the input point id that Qhull created k-th.
type creationOrders struct {
	QhullVersion string           `json:"qhull_version"`
	Order        map[string][]int `json:"order"`
}

func loadCreationOrders(t *testing.T) creationOrders {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "creation_order.json"))
	if err != nil {
		t.Fatalf("read creation_order.json: %v", err)
	}
	var c creationOrders
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("parse creation_order.json: %v", err)
	}
	return c
}

// TestFanFromGroundTruthOrder proves the geometric model: given Qhull's true
// vertex creation order, fanning each cocircular cell from its last-created
// vertex reproduces Qhull's triangulation exactly. This isolates all remaining
// Phase-12 work to "compute the creation order" — if this passes 34/34 + 27/27,
// the diagonal-selection geometry is correct and only the qh_buildhull port
// remains. It is a hard gate for BOTH categories (uses captured ground truth).
func TestFanFromGroundTruthOrder(t *testing.T) {
	corp := loadCorpus(t)
	orders := loadCreationOrders(t)

	byCat := map[string]*struct{ pass, total int }{}
	var fails []string
	for _, tc := range corp.Cases {
		st := byCat[tc.Category]
		if st == nil {
			st = &struct{ pass, total int }{}
			byCat[tc.Category] = st
		}
		st.total++

		order, ok := orders.Order[tc.Name]
		if !ok {
			t.Fatalf("%s: no captured creation order", tc.Name)
		}
		gotTris, gotNbrs, err := delaunayFromOrder(tc.X, tc.Y, order)
		match := err == nil &&
			len(gotTris) == len(tc.Triangles) &&
			sameTriangleSet(gotTris, tc.Triangles) &&
			sameNeighborGraph(gotTris, gotNbrs, tc.Triangles, tc.Neighbors)
		if match {
			st.pass++
		} else {
			fails = append(fails, tc.Name)
		}
	}

	for _, cat := range []string{"general", "cocircular"} {
		if st := byCat[cat]; st != nil {
			t.Logf("category %-10s: %d/%d match", cat, st.pass, st.total)
			if st.pass != st.total {
				t.Errorf("%s: fan-from-ground-truth-order must match Qhull exactly: %d/%d",
					cat, st.pass, st.total)
			}
		}
	}
	if len(fails) > 0 {
		t.Errorf("failing cases: %v", fails)
	}
}

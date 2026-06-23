package qhull

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// corpusCase is one differential-test fixture: original (x,y) plus the
// triangles and neighbors that matplotlib's Qhull backend produced for it.
type corpusCase struct {
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	X         []float64 `json:"x"`
	Y         []float64 `json:"y"`
	Triangles [][3]int  `json:"triangles"`
	Neighbors [][3]int  `json:"neighbors"`
}

type corpus struct {
	QhullVersion string       `json:"qhull_version"`
	Cases        []corpusCase `json:"cases"`
}

func loadCorpus(t *testing.T) corpus {
	t.Helper()
	path := filepath.Join("testdata", "corpus.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}
	var c corpus
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("parse corpus: %v", err)
	}
	if len(c.Cases) == 0 {
		t.Fatal("empty corpus")
	}
	return c
}

// triKey canonicalizes a triangle to its sorted vertex tuple, so two
// triangulations can be compared as sets independent of array order and of
// per-triangle vertex rotation/orientation.
func triKey(tri [3]int) [3]int {
	v := tri
	sort.Ints(v[:])
	return v
}

// triangleSet returns the multiset of canonical triangle keys.
func triangleSet(tris [][3]int) map[[3]int]int {
	m := make(map[[3]int]int, len(tris))
	for _, tr := range tris {
		m[triKey(tr)]++
	}
	return m
}

func sameTriangleSet(a, b [][3]int) bool {
	sa, sb := triangleSet(a), triangleSet(b)
	if len(sa) != len(sb) {
		return false
	}
	for k, n := range sa {
		if sb[k] != n {
			return false
		}
	}
	return true
}

// neighborGraph returns the set of undirected adjacencies {triKey,triKey}
// between triangles that share an edge, keyed canonically. This captures the
// connectivity Qhull reports in its neighbors array without depending on
// triangle ordering.
func neighborGraph(tris, nbrs [][3]int) map[[2][3]int]struct{} {
	g := make(map[[2][3]int]struct{})
	for i, row := range nbrs {
		for _, nb := range row {
			if nb < 0 {
				continue
			}
			a, b := triKey(tris[i]), triKey(tris[nb])
			pair := [2][3]int{a, b}
			if greaterKey(a, b) {
				pair = [2][3]int{b, a}
			}
			g[pair] = struct{}{}
		}
	}
	return g
}

func greaterKey(a, b [3]int) bool {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func sameNeighborGraph(aTris, aNbrs, bTris, bNbrs [][3]int) bool {
	ga := neighborGraph(aTris, aNbrs)
	gb := neighborGraph(bTris, bNbrs)
	if len(ga) != len(gb) {
		return false
	}
	for k := range ga {
		if _, ok := gb[k]; !ok {
			return false
		}
	}
	return true
}

// cocircularRatchet is the minimum number of cocircular corpus cases whose
// connectivity must match Qhull. It is a progress ratchet for the in-progress
// faithful qh_buildhull port: bump it as the port closes cases, never lower it.
// The target is the full cocircular count (all cases), reached once the hull
// construction order and qh_triangulate_facet fan are ported.
const cocircularRatchet = 6

// TestDelaunayConnectivityVsQhull is the primary differential gate: for every
// corpus case the Go engine must reproduce Qhull's triangle set and neighbor
// graph (compared as sets/graphs, independent of Qhull's internal array order
// and per-row vertex rotation). General-position inputs have a unique Delaunay,
// so they are a hard gate (all must match). Cocircular inputs are non-unique;
// matching Qhull's specific diagonal needs the construction-order-dependent fan,
// gated by a ratchet until the hull port lands.
func TestDelaunayConnectivityVsQhull(t *testing.T) {
	c := loadCorpus(t)
	type stat struct {
		pass, total int
		fails       []string
	}
	byCat := map[string]*stat{}

	for _, tc := range c.Cases {
		st := byCat[tc.Category]
		if st == nil {
			st = &stat{}
			byCat[tc.Category] = st
		}
		st.total++

		gotTris, gotNbrs, err := Delaunay(tc.X, tc.Y)
		ok := err == nil &&
			len(gotTris) == len(tc.Triangles) &&
			sameTriangleSet(gotTris, tc.Triangles) &&
			sameNeighborGraph(gotTris, gotNbrs, tc.Triangles, tc.Neighbors)
		if ok {
			st.pass++
		} else {
			st.fails = append(st.fails, tc.Name)
		}
	}

	gen := byCat["general"]
	if gen != nil {
		t.Logf("category general   : %d/%d match", gen.pass, gen.total)
		if gen.pass != gen.total {
			t.Errorf("general position must match Qhull exactly: %d/%d (failing: %v)",
				gen.pass, gen.total, gen.fails)
		}
	}

	if co := byCat["cocircular"]; co != nil {
		t.Logf("category cocircular: %d/%d match (ratchet %d, target %d); remaining: %v",
			co.pass, co.total, cocircularRatchet, co.total, co.fails)
		if co.pass < cocircularRatchet {
			t.Errorf("cocircular regressed below ratchet: %d/%d < %d",
				co.pass, co.total, cocircularRatchet)
		}
	}
}

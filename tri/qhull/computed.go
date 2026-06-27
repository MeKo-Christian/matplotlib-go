package qhull

// delaunayComputed is the fully self-contained (no Qhull, no captured fixture)
// Qhull-faithful engine: it computes the vertex creation order with the faithful
// ridge-graph incremental hull (buildHullOrderRidge, including the coplanarhorizon
// merge), then fans each cocircular cell from its last-created vertex, reproducing
// matplotlib's diagonal choice.
func delaunayComputed(x, y []float64) (triangles, neighbors [][3]int, err error) {
	order, ok := buildHullOrderRidge(project(x, y))
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

package tri

import (
	"fmt"
	"math"
)

// CubicTriInterpolator performs C1-smooth cubic interpolation over a
// triangulation using a reduced Hsieh–Clough–Tocher (HCT) finite element. It
// is a faithful port of matplotlib's CubicTriInterpolator (_triinterpolate.py),
// including the three gradient-estimation strategies.
type CubicTriInterpolator struct {
	finder TriFinder

	triangles [][3]int     // compressed (unmasked) triangles, renumbered nodes
	triRenum  []int        // original triangle index -> compressed index (or -1)
	pts       [][2]float64 // scaled node coordinates
	trisPts   [][3][2]float64
	eccs      [][3]float64
	dof       [][9]float64 // per-triangle 9 dofs

	unitX, unitY float64
}

// CubicKind selects the gradient-estimation strategy.
type CubicKind int

const (
	// CubicMinE minimises a bending energy (matplotlib default, 'min_E').
	CubicMinE CubicKind = iota
	// CubicGeom uses a fast geometric weighted average ('geom').
	CubicGeom
	// CubicUser uses caller-supplied nodal gradients ('user').
	CubicUser
)

// NewCubicTriInterpolator builds a cubic interpolator for field z. For
// CubicUser, dzdx and dzdy (nodal gradients, one per node) must be supplied;
// otherwise pass nil.
func NewCubicTriInterpolator(t Triangulation, z []float64, kind CubicKind, dzdx, dzdy []float64) (*CubicTriInterpolator, error) {
	if len(z) != len(t.X) {
		return nil, fmt.Errorf("interpolator requires one z per node: got %d, want %d", len(z), len(t.X))
	}
	if kind == CubicUser && (len(dzdx) != len(t.X) || len(dzdy) != len(t.X)) {
		return nil, fmt.Errorf("CubicUser requires dzdx and dzdy of length %d", len(t.X))
	}

	comp, cx, cy, triRenum, nodeRenum := t.compress()
	if len(comp) == 0 {
		return nil, fmt.Errorf("triangulation has no unmasked triangles")
	}

	// Remap z (and user gradients) onto compressed node numbering.
	nNodes := len(cx)
	cz := make([]float64, nNodes)
	var cdx, cdy []float64
	if kind == CubicUser {
		cdx = make([]float64, nNodes)
		cdy = make([]float64, nNodes)
	}
	for old, nw := range nodeRenum {
		if nw < 0 {
			continue
		}
		cz[nw] = z[old]
		if kind == CubicUser {
			cdx[nw] = dzdx[old]
			cdy[nw] = dzdy[old]
		}
	}

	unitX := ptp(cx)
	unitY := ptp(cy)
	if unitX == 0 {
		unitX = 1
	}
	if unitY == 0 {
		unitY = 1
	}
	pts := make([][2]float64, nNodes)
	for i := range pts {
		pts[i] = [2]float64{cx[i] / unitX, cy[i] / unitY}
	}

	ci := &CubicTriInterpolator{
		finder:    t.TriFinder(),
		triangles: comp,
		triRenum:  triRenum,
		pts:       pts,
		unitX:     unitX,
		unitY:     unitY,
	}
	ci.trisPts = make([][3][2]float64, len(comp))
	ci.eccs = make([][3]float64, len(comp))
	for i, tr := range comp {
		ci.trisPts[i] = [3][2]float64{pts[tr[0]], pts[tr[1]], pts[tr[2]]}
		ci.eccs[i] = triEccentricities(ci.trisPts[i])
	}

	// Estimate nodal gradients then assemble per-triangle dofs.
	var dz [][2]float64
	switch kind {
	case CubicUser:
		dz = make([][2]float64, nNodes)
		for i := range dz {
			dz[i] = [2]float64{cdx[i] * unitX, cdy[i] * unitY}
		}
	case CubicGeom:
		dz = ci.computeGeomDz(cz)
	default: // CubicMinE
		dz = ci.computeMinEDz(cz)
	}
	ci.dof = ci.computeDofFromDf(cz, dz)
	return ci, nil
}

// Interpolate returns z at (x, y), or ok=false outside the unmasked mesh.
func (ci *CubicTriInterpolator) Interpolate(x, y float64) (float64, bool) {
	tri, sx, sy, ok := ci.locate(x, y)
	if !ok {
		return 0, false
	}
	alpha := ci.alphaVec(tri, sx, sy)
	return hctFunctionValue(alpha, ci.eccs[tri], ci.dof[tri]), true
}

// Gradient returns (dz/dx, dz/dy) at (x, y), or ok=false outside the mesh.
func (ci *CubicTriInterpolator) Gradient(x, y float64) (float64, float64, bool) {
	tri, sx, sy, ok := ci.locate(x, y)
	if !ok {
		return 0, 0, false
	}
	alpha := ci.alphaVec(tri, sx, sy)
	j := jacobian(ci.trisPts[tri])
	d := hctFunctionDerivatives(alpha, j, ci.eccs[tri], ci.dof[tri])
	// Undo coordinate scaling on the derivatives.
	return d[0] / ci.unitX, d[1] / ci.unitY, true
}

// interpolateAt evaluates within a specific original triangle, bypassing the
// point locator. Used by RefineField for boundary nodes. Implements
// triIndexInterpolator.
func (ci *CubicTriInterpolator) interpolateAt(origTri int, x, y float64) (float64, bool) {
	if origTri < 0 || origTri >= len(ci.triRenum) {
		return 0, false
	}
	tri := ci.triRenum[origTri]
	if tri < 0 {
		return 0, false
	}
	alpha := ci.alphaVec(tri, x/ci.unitX, y/ci.unitY)
	return hctFunctionValue(alpha, ci.eccs[tri], ci.dof[tri]), true
}

// locate maps (x,y) to a compressed triangle index and scaled coordinates.
func (ci *CubicTriInterpolator) locate(x, y float64) (tri int, sx, sy float64, ok bool) {
	orig := ci.finder.Find(x, y)
	if orig < 0 || orig >= len(ci.triRenum) {
		return 0, 0, 0, false
	}
	tri = ci.triRenum[orig]
	if tri < 0 {
		return 0, 0, 0, false
	}
	return tri, x / ci.unitX, y / ci.unitY, true
}

// --- per-triangle geometry ---

func (ci *CubicTriInterpolator) alphaVec(tri int, x, y float64) [3]float64 {
	p := ci.trisPts[tri]
	a := [2]float64{p[1][0] - p[0][0], p[1][1] - p[0][1]}
	b := [2]float64{p[2][0] - p[0][0], p[2][1] - p[0][1]}
	// ab = [[a0,a1],[b0,b1]]; metric = ab @ ab^T.
	metric := [2][2]float64{
		{a[0]*a[0] + a[1]*a[1], a[0]*b[0] + a[1]*b[1]},
		{b[0]*a[0] + b[1]*a[1], b[0]*b[0] + b[1]*b[1]},
	}
	mInv := pseudoInv22Sym(metric)
	om := [2]float64{x - p[0][0], y - p[0][1]}
	// Covar = ab @ om.
	covar := [2]float64{a[0]*om[0] + a[1]*om[1], b[0]*om[0] + b[1]*om[1]}
	ksi0 := mInv[0][0]*covar[0] + mInv[0][1]*covar[1]
	ksi1 := mInv[1][0]*covar[0] + mInv[1][1]*covar[1]
	return [3]float64{1 - ksi0 - ksi1, ksi0, ksi1}
}

func jacobian(p [3][2]float64) [2][2]float64 {
	a := [2]float64{p[1][0] - p[0][0], p[1][1] - p[0][1]}
	b := [2]float64{p[2][0] - p[0][0], p[2][1] - p[0][1]}
	return [2][2]float64{{a[0], a[1]}, {b[0], b[1]}}
}

func triEccentricities(p [3][2]float64) [3]float64 {
	a := [2]float64{p[2][0] - p[1][0], p[2][1] - p[1][1]}
	b := [2]float64{p[0][0] - p[2][0], p[0][1] - p[2][1]}
	c := [2]float64{p[1][0] - p[0][0], p[1][1] - p[0][1]}
	dotA := a[0]*a[0] + a[1]*a[1]
	dotB := b[0]*b[0] + b[1]*b[1]
	dotC := c[0]*c[0] + c[1]*c[1]
	return [3]float64{(dotC - dotB) / dotA, (dotA - dotC) / dotB, (dotB - dotA) / dotC}
}

// --- compression ---

// compress returns the unmasked triangles with nodes renumbered to the set of
// used points, plus the original->compressed triangle map and the
// original->compressed node map (-1 for dropped entries). Mirrors matplotlib's
// TriAnalyzer._get_compressed_triangulation.
func (t Triangulation) compress() (tris [][3]int, x, y []float64, triRenum, nodeRenum []int) {
	used := make([]bool, len(t.X))
	triRenum = make([]int, len(t.Triangles))
	tris = make([][3]int, 0, len(t.Triangles))
	for i, tr := range t.Triangles {
		if t.Masked(i) {
			triRenum[i] = -1
			continue
		}
		triRenum[i] = len(tris)
		tris = append(tris, tr)
		used[tr[0]] = true
		used[tr[1]] = true
		used[tr[2]] = true
	}
	nodeRenum = make([]int, len(t.X))
	newIdx := 0
	for i := range used {
		if used[i] {
			nodeRenum[i] = newIdx
			x = append(x, t.X[i])
			y = append(y, t.Y[i])
			newIdx++
		} else {
			nodeRenum[i] = -1
		}
	}
	for k := range tris {
		tris[k] = [3]int{nodeRenum[tris[k][0]], nodeRenum[tris[k][1]], nodeRenum[tris[k][2]]}
	}
	return tris, x, y, triRenum, nodeRenum
}

func ptp(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	lo, hi := v[0], v[0]
	for _, x := range v[1:] {
		lo = math.Min(lo, x)
		hi = math.Max(hi, x)
	}
	return hi - lo
}

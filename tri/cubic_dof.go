package tri

import "math"

// computeDofFromDf assembles per-triangle 9-dof vectors from nodal values cz
// and nodal gradients dz (get_dof_vec / compute_dof_from_df).
func (ci *CubicTriInterpolator) computeDofFromDf(cz []float64, dz [][2]float64) [][9]float64 {
	out := make([][9]float64, len(ci.triangles))
	for t, tr := range ci.triangles {
		j := jacobian(ci.trisPts[t])
		j1 := mat2x2Mul(hctJ0toJ1, j)
		j2 := mat2x2Mul(hctJ0toJ2, j)
		dz0, dz1, dz2 := dz[tr[0]], dz[tr[1]], dz[tr[2]]
		col0 := mat2Vec(j, dz0)
		col1 := mat2Vec(j1, dz1)
		col2 := mat2Vec(j2, dz2)
		var dof [9]float64
		dof[0], dof[3], dof[6] = cz[tr[0]], cz[tr[1]], cz[tr[2]]
		dof[1], dof[4], dof[7] = col0[0], col1[0], col2[0]
		dof[2], dof[5], dof[8] = col0[1], col1[1], col2[1]
		out[t] = dof
	}
	return out
}

func mat2Vec(m [2][2]float64, v [2]float64) [2]float64 {
	return [2]float64{m[0][0]*v[0] + m[0][1]*v[1], m[1][0]*v[0] + m[1][1]*v[1]}
}

// computeGeomDz estimates nodal gradients as an angle-weighted average of the
// per-triangle linear gradients (_DOF_estimator_geom).
func (ci *CubicTriInterpolator) computeGeomDz(cz []float64) [][2]float64 {
	nNodes := len(ci.pts)
	weights := ci.geomWeights()
	grads := ci.geomGrads(cz)

	wSum := make([]float64, nNodes)
	dfxSum := make([]float64, nNodes)
	dfySum := make([]float64, nNodes)
	for t, tr := range ci.triangles {
		for apex := 0; apex < 3; apex++ {
			node := tr[apex]
			w := weights[t][apex]
			wSum[node] += w
			dfxSum[node] += w * grads[t][0]
			dfySum[node] += w * grads[t][1]
		}
	}
	dz := make([][2]float64, nNodes)
	for i := 0; i < nNodes; i++ {
		if wSum[i] != 0 {
			dz[i] = [2]float64{dfxSum[i] / wSum[i], dfySum[i] / wSum[i]}
		}
	}
	return dz
}

// geomWeights builds per-triangle, per-apex weights proportional to the apex
// angle (compute_geom_weights).
func (ci *CubicTriInterpolator) geomWeights() [][3]float64 {
	out := make([][3]float64, len(ci.triangles))
	for t := range ci.triangles {
		p := ci.trisPts[t]
		for ipt := 0; ipt < 3; ipt++ {
			p0 := p[ipt%3]
			p1 := p[(ipt+1)%3]
			p2 := p[(ipt+2)%3] // (ipt-1) mod 3
			alpha1 := math.Atan2(p1[1]-p0[1], p1[0]-p0[0])
			alpha2 := math.Atan2(p2[1]-p0[1], p2[0]-p0[0])
			angle := math.Abs(pyMod1((alpha2 - alpha1) / math.Pi))
			out[t][ipt] = 0.5 - math.Abs(angle-0.5)
		}
	}
	return out
}

// pyMod1 returns the Python-style a % 1 (always in [0,1)).
func pyMod1(a float64) float64 { return a - math.Floor(a) }

// geomGrads computes the per-triangle gradient of the linear interpolant
// (compute_geom_grads).
func (ci *CubicTriInterpolator) geomGrads(cz []float64) [][2]float64 {
	out := make([][2]float64, len(ci.triangles))
	for t, tr := range ci.triangles {
		p := ci.trisPts[t]
		// dM = [[dM1x, dM2x],[dM1y, dM2y]].
		dM := [2][2]float64{
			{p[1][0] - p[0][0], p[2][0] - p[0][0]},
			{p[1][1] - p[0][1], p[2][1] - p[0][1]},
		}
		dMInv := safeInv22(dM)
		dZ1 := cz[tr[1]] - cz[tr[0]]
		dZ2 := cz[tr[2]] - cz[tr[0]]
		out[t] = [2]float64{
			dZ1*dMInv[0][0] + dZ2*dMInv[1][0],
			dZ1*dMInv[0][1] + dZ2*dMInv[1][1],
		}
	}
	return out
}

// computeMinEDz estimates nodal gradients by minimising bending energy via a
// Jacobi-preconditioned conjugate-gradient solve (_DOF_estimator_min_E),
// falling back to the geometric estimate if the solver does not converge.
func (ci *CubicTriInterpolator) computeMinEDz(cz []float64) [][2]float64 {
	nNodes := len(ci.pts)
	geom := ci.computeGeomDz(cz)
	uf0 := make([]float64, 2*nNodes)
	for i := 0; i < nNodes; i++ {
		uf0[2*i] = geom[i][0]
		uf0[2*i+1] = geom[i][1]
	}

	rows, cols, vals, ff := ci.buildKffAndFf(cz)
	nDof := 2 * nNodes
	a := newSparseCOO(vals, rows, cols, nDof)
	a.compress()

	uf, err := conjugateGradient(a, ff, uf0, 1e-10, 1000)
	// Compare against the initial guess; keep whichever has the smaller residual.
	res0 := vecNorm(vecSub(a.dot(uf0), ff))
	if res0 < err {
		uf = uf0
	}

	dz := make([][2]float64, nNodes)
	for i := 0; i < nNodes; i++ {
		dz[i] = [2]float64{uf[2*i], uf[2*i+1]}
	}
	return dz
}

// buildKffAndFf assembles the global stiffness matrix (free dofs) in COO form
// and the force vector (get_Kff_and_Ff).
func (ci *CubicTriInterpolator) buildKffAndFf(cz []float64) (rows, cols []int, vals, ff []float64) {
	nDof := 2 * len(ci.pts)
	ff = make([]float64, nDof)
	fDof := [6]int{1, 2, 4, 5, 7, 8}
	cDof := [3]int{0, 3, 6}

	for t, tr := range ci.triangles {
		j := jacobian(ci.trisPts[t])
		k := hctBendingMatrix(j, ci.eccs[t])

		// Global dof index for each of the 9 local dofs (-1 for condensed z).
		gdof := [9]int{-1, 2 * tr[0], 2*tr[0] + 1, -1, 2 * tr[1], 2*tr[1] + 1, -1, 2 * tr[2], 2*tr[2] + 1}

		// Kff entries.
		for _, i := range fDof {
			for _, jj := range fDof {
				rows = append(rows, gdof[i])
				cols = append(cols, gdof[jj])
				vals = append(vals, k[i][jj])
			}
		}
		// Ff = -(Kfc @ Uc) scattered to global free dofs.
		uc := [3]float64{cz[tr[0]], cz[tr[1]], cz[tr[2]]}
		for _, i := range fDof {
			var s float64
			for ci2, c := range cDof {
				s += k[i][c] * uc[ci2]
			}
			ff[gdof[i]] += -s
		}
	}
	return rows, cols, vals, ff
}

// --- sparse matrix (COO) and PCG solver ---

type sparseCOO struct {
	n          int
	rows, cols []int
	vals       []float64
}

func newSparseCOO(vals []float64, rows, cols []int, n int) *sparseCOO {
	return &sparseCOO{n: n, rows: append([]int(nil), rows...), cols: append([]int(nil), cols...), vals: append([]float64(nil), vals...)}
}

// compress sums duplicate (row,col) entries.
func (s *sparseCOO) compress() {
	acc := make(map[[2]int]float64, len(s.vals))
	for i, v := range s.vals {
		key := [2]int{s.rows[i], s.cols[i]}
		acc[key] += v
	}
	s.rows = s.rows[:0]
	s.cols = s.cols[:0]
	s.vals = s.vals[:0]
	for key, v := range acc {
		s.rows = append(s.rows, key[0])
		s.cols = append(s.cols, key[1])
		s.vals = append(s.vals, v)
	}
}

func (s *sparseCOO) dot(v []float64) []float64 {
	out := make([]float64, s.n)
	for i, val := range s.vals {
		out[s.rows[i]] += val * v[s.cols[i]]
	}
	return out
}

func (s *sparseCOO) diag() []float64 {
	d := make([]float64, s.n)
	for i := range s.vals {
		if s.rows[i] == s.cols[i] {
			d[s.rows[i]] += s.vals[i]
		}
	}
	return d
}

// conjugateGradient solves A x = b with a Jacobi preconditioner, following the
// algorithm in matplotlib's _cg. Returns the solution and the residual norm.
func conjugateGradient(a *sparseCOO, b, x0 []float64, tol float64, maxiter int) ([]float64, float64) {
	n := len(b)
	bNorm := vecNorm(b)
	kvec := a.diag()
	for i := range kvec {
		kvec[i] = math.Max(kvec[i], 1e-6)
	}

	x := append([]float64(nil), x0...)
	r := vecSub(b, a.dot(x))
	w := vecDivElem(r, kvec)
	p := make([]float64, n)
	beta := 0.0
	rho := dotVec(r, w)
	k := 0

	for math.Sqrt(math.Abs(rho)) > tol*bNorm && k < maxiter {
		for i := 0; i < n; i++ {
			p[i] = w[i] + beta*p[i]
		}
		z := a.dot(p)
		alpha := rho / dotVec(p, z)
		for i := 0; i < n; i++ {
			r[i] -= alpha * z[i]
		}
		w = vecDivElem(r, kvec)
		rhoOld := rho
		rho = dotVec(r, w)
		for i := 0; i < n; i++ {
			x[i] += alpha * p[i]
		}
		beta = rho / rhoOld
		k++
	}
	err := vecNorm(vecSub(a.dot(x), b))
	return x, err
}

func vecSub(a, b []float64) []float64 {
	out := make([]float64, len(a))
	for i := range a {
		out[i] = a[i] - b[i]
	}
	return out
}

func vecDivElem(a, b []float64) []float64 {
	out := make([]float64, len(a))
	for i := range a {
		out[i] = a[i] / b[i]
	}
	return out
}

func dotVec(a, b []float64) float64 {
	var s float64
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

func vecNorm(a []float64) float64 { return math.Sqrt(dotVec(a, a)) }

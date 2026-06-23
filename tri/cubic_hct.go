package tri

import "math"

// Reduced HCT element shape-function matrices (Bernadou & Hassan), ported from
// matplotlib's _ReducedHCT_Element. M, M0, M1, M2 are 9x10; they map the
// monomial vector V (and its derivatives) to the 9 element shape functions,
// modulated by triangle eccentricities.
var (
	hctM = [9][10]float64{
		{0.00, 0.00, 0.00, 4.50, 4.50, 0.00, 0.00, 0.00, 0.00, 0.00},
		{-0.25, 0.00, 0.00, 0.50, 1.25, 0.00, 0.00, 0.00, 0.00, 0.00},
		{-0.25, 0.00, 0.00, 1.25, 0.50, 0.00, 0.00, 0.00, 0.00, 0.00},
		{0.50, 1.00, 0.00, -1.50, 0.00, 3.00, 3.00, 0.00, 0.00, 3.00},
		{0.00, 0.00, 0.00, -0.25, 0.25, 0.00, 1.00, 0.00, 0.00, 0.50},
		{0.25, 0.00, 0.00, -0.50, -0.25, 1.00, 0.00, 0.00, 0.00, 1.00},
		{0.50, 0.00, 1.00, 0.00, -1.50, 0.00, 0.00, 3.00, 3.00, 3.00},
		{0.25, 0.00, 0.00, -0.25, -0.50, 0.00, 0.00, 0.00, 1.00, 1.00},
		{0.00, 0.00, 0.00, 0.25, -0.25, 0.00, 0.00, 1.00, 0.00, 0.50},
	}
	hctM0 = [9][10]float64{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{-1.00, 0, 0, 1.50, 1.50, 0, 0, 0, 0, -3.00},
		{-0.50, 0, 0, 0.75, 0.75, 0, 0, 0, 0, -1.50},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{1.00, 0, 0, -1.50, -1.50, 0, 0, 0, 0, 3.00},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0.50, 0, 0, -0.75, -0.75, 0, 0, 0, 0, 1.50},
	}
	hctM1 = [9][10]float64{
		{-0.50, 0, 0, 1.50, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{-0.25, 0, 0, 0.75, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0.50, 0, 0, -1.50, 0, 0, 0, 0, 0, 0},
		{0.25, 0, 0, -0.75, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	hctM2 = [9][10]float64{
		{0.50, 0, 0, 0, -1.50, 0, 0, 0, 0, 0},
		{0.25, 0, 0, 0, -0.75, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{-0.50, 0, 0, 0, 1.50, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{-0.25, 0, 0, 0, 0.75, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	hctRotateDV  = [6][2]float64{{1, 0}, {0, 1}, {0, 1}, {-1, -1}, {-1, -1}, {1, 0}}
	hctRotateD2V = [9][3]float64{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
		{0, 1, 0},
		{1, 1, 1},
		{0, -2, -1},
		{1, 1, 1},
		{1, 0, 0},
		{-2, 0, -1},
	}

	hctGaussPts = [9][3]float64{
		{13. / 18, 4. / 18, 1. / 18},
		{4. / 18, 13. / 18, 1. / 18},
		{7. / 18, 7. / 18, 4. / 18},
		{1. / 18, 13. / 18, 4. / 18},
		{1. / 18, 4. / 18, 13. / 18},
		{4. / 18, 7. / 18, 7. / 18},
		{4. / 18, 1. / 18, 13. / 18},
		{13. / 18, 1. / 18, 4. / 18},
		{7. / 18, 4. / 18, 7. / 18},
	}
	hctEMat   = [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 2}}
	hctJ0toJ1 = [2][2]float64{{-1, 1}, {-1, 0}}
	hctJ0toJ2 = [2][2]float64{{0, -1}, {1, -1}}
)

func argmin3(a [3]float64) int {
	idx := 0
	if a[1] < a[idx] {
		idx = 1
	}
	if a[2] < a[idx] {
		idx = 2
	}
	return idx
}

// monomialV builds the 10-vector V used by the shape functions.
func monomialV(x, y, z float64) [10]float64 {
	xs, ys, zs := x*x, y*y, z*z
	return [10]float64{xs * x, ys * y, zs * z, xs * z, xs * y, ys * x, ys * z, zs * y, zs * x, x * y * z}
}

// matVec9x10 returns M·v for M (9x10) and v (10).
func matVec9x10(m [9][10]float64, v [10]float64) [9]float64 {
	var out [9]float64
	for i := 0; i < 9; i++ {
		var s float64
		for k := 0; k < 10; k++ {
			s += m[i][k] * v[k]
		}
		out[i] = s
	}
	return out
}

// combineShape computes M·V + e0·M0·V + e1·M1·V + e2·M2·V for a vector V (the
// rolled eccentricities e).
func combineShape(v [10]float64, e [3]float64) [9]float64 {
	base := matVec9x10(hctM, v)
	m0 := matVec9x10(hctM0, v)
	m1 := matVec9x10(hctM1, v)
	m2 := matVec9x10(hctM2, v)
	var out [9]float64
	for i := 0; i < 9; i++ {
		out[i] = base[i] + e[0]*m0[i] + e[1]*m1[i] + e[2]*m2[i]
	}
	return out
}

func roll3(a [3]float64, subtri int) [3]float64 {
	return [3]float64{a[subtri%3], a[(subtri+1)%3], a[(subtri+2)%3]}
}

// roll9 returns s with s[i] = p[(i-shift) mod 9].
func roll9(p [9]float64, shift int) [9]float64 {
	var out [9]float64
	for i := 0; i < 9; i++ {
		out[i] = p[((i-shift)%9+9)%9]
	}
	return out
}

// hctFunctionValue evaluates the interpolated field value (get_function_values).
func hctFunctionValue(alpha, ecc [3]float64, dofs [9]float64) float64 {
	subtri := argmin3(alpha)
	ksi := roll3(alpha, subtri)
	e := roll3(ecc, subtri)
	v := monomialV(ksi[0], ksi[1], ksi[2])
	prod := combineShape(v, e)
	s := roll9(prod, 3*subtri)
	var out float64
	for i := 0; i < 9; i++ {
		out += dofs[i] * s[i]
	}
	return out
}

// hctFunctionDerivatives evaluates [dz/dx, dz/dy] in (scaled) global coords.
func hctFunctionDerivatives(alpha [3]float64, j [2][2]float64, ecc [3]float64, dofs [9]float64) [2]float64 {
	subtri := argmin3(alpha)
	ksi := roll3(alpha, subtri)
	e := roll3(ecc, subtri)
	x, y, z := ksi[0], ksi[1], ksi[2]
	xs, ys, zs := x*x, y*y, z*z
	// dV is 10x2.
	dV := [10][2]float64{
		{-3 * xs, -3 * xs},
		{3 * ys, 0},
		{0, 3 * zs},
		{-2 * x * z, -2*x*z + xs},
		{-2*x*y + xs, -2 * x * y},
		{2*x*y - ys, -ys},
		{2 * y * z, ys},
		{zs, 2 * y * z},
		{-zs, 2*x*z - zs},
		{x*z - y*z, x*y - y*z},
	}
	// dV = dV @ R where R = rotate_dV[2*subtri:2*subtri+2] (2x2).
	r0 := hctRotateDV[2*subtri]
	r1 := hctRotateDV[2*subtri+1]
	var dVr [10][2]float64
	for i := 0; i < 10; i++ {
		dVr[i][0] = dV[i][0]*r0[0] + dV[i][1]*r1[0]
		dVr[i][1] = dV[i][0]*r0[1] + dV[i][1]*r1[1]
	}
	// prod (9x2) = M@dVr + e·Mi@dVr, column by column.
	col0 := combineShapeCol(dVr, e, 0)
	col1 := combineShapeCol(dVr, e, 1)
	// dsdksi = roll rows by 3*subtri; dfdksi = dofs · dsdksi (1x2).
	s0 := roll9(col0, 3*subtri)
	s1 := roll9(col1, 3*subtri)
	var dfdksi [2]float64
	for i := 0; i < 9; i++ {
		dfdksi[0] += dofs[i] * s0[i]
		dfdksi[1] += dofs[i] * s1[i]
	}
	jInv := safeInv22(j)
	// dfdx = J_inv @ dfdksi.
	return [2]float64{
		jInv[0][0]*dfdksi[0] + jInv[0][1]*dfdksi[1],
		jInv[1][0]*dfdksi[0] + jInv[1][1]*dfdksi[1],
	}
}

// combineShapeCol applies combineShape to one column of a 10x2/10x3 matrix.
func combineShapeCol(dV [10][2]float64, e [3]float64, col int) [9]float64 {
	var v [10]float64
	for i := 0; i < 10; i++ {
		v[i] = dV[i][col]
	}
	return combineShape(v, e)
}

// hctD2Sdksi2 returns the 9x3 Hessian of shape functions (get_d2Sidksij2).
func hctD2Sdksi2(alpha, ecc [3]float64) [9][3]float64 {
	subtri := argmin3(alpha)
	ksi := roll3(alpha, subtri)
	e := roll3(ecc, subtri)
	x, y, z := ksi[0], ksi[1], ksi[2]
	d2V := [10][3]float64{
		{6 * x, 6 * x, 6 * x},
		{6 * y, 0, 0},
		{0, 6 * z, 0},
		{2 * z, 2*z - 4*x, 2*z - 2*x},
		{2*y - 4*x, 2 * y, 2*y - 2*x},
		{2*x - 4*y, 0, -2 * y},
		{2 * z, 0, 2 * y},
		{0, 2 * y, 2 * z},
		{0, 2*x - 4*z, -2 * z},
		{-2 * z, -2 * y, x - y - z},
	}
	// d2V = d2V @ R, R = rotate_d2V[3*subtri:3*subtri+3] (3x3).
	r0 := hctRotateD2V[3*subtri]
	r1 := hctRotateD2V[3*subtri+1]
	r2 := hctRotateD2V[3*subtri+2]
	var d2Vr [10][3]float64
	for i := 0; i < 10; i++ {
		for c := 0; c < 3; c++ {
			d2Vr[i][c] = d2V[i][0]*r0[c] + d2V[i][1]*r1[c] + d2V[i][2]*r2[c]
		}
	}
	var prod [9][3]float64
	for c := 0; c < 3; c++ {
		var v [10]float64
		for i := 0; i < 10; i++ {
			v[i] = d2Vr[i][c]
		}
		col := combineShape(v, e)
		s := roll9(col, 3*subtri)
		for i := 0; i < 9; i++ {
			prod[i][c] = s[i]
		}
	}
	return prod
}

// hctHrotFromJ returns the 3x3 Hessian-rotation matrix and the triangle area.
func hctHrotFromJ(j [2][2]float64) (hrot [3][3]float64, area float64) {
	jInv := safeInv22(j)
	ji00, ji11 := jInv[0][0], jInv[1][1]
	ji10, ji01 := jInv[1][0], jInv[0][1]
	hrot = [3][3]float64{
		{ji00 * ji00, ji10 * ji10, ji00 * ji10},
		{ji01 * ji01, ji11 * ji11, ji01 * ji11},
		{2 * ji00 * ji01, 2 * ji11 * ji10, ji00*ji11 + ji10*ji01},
	}
	area = 0.5 * (j[0][0]*j[1][1] - j[0][1]*j[1][0])
	return hrot, area
}

// hctBendingMatrix returns the 9x9 element stiffness matrix in nodal coords.
func hctBendingMatrix(j [2][2]float64, ecc [3]float64) [9][9]float64 {
	j1 := mat2x2Mul(hctJ0toJ1, j)
	j2 := mat2x2Mul(hctJ0toJ2, j)
	// DOF_rot 9x9.
	var dofRot [9][9]float64
	dofRot[0][0] = 1
	dofRot[3][3] = 1
	dofRot[6][6] = 1
	put2x2(&dofRot, 1, 1, j)
	put2x2(&dofRot, 4, 4, j1)
	put2x2(&dofRot, 7, 7, j2)

	hrot, area := hctHrotFromJ(j)

	var k [9][9]float64
	for ig := 0; ig < 9; ig++ {
		alpha := hctGaussPts[ig]
		weight := 1.0 / 9.0
		d2 := hctD2Sdksi2(alpha, ecc) // 9x3
		// d2x = d2 @ hrot (9x3).
		var d2x [9][3]float64
		for i := 0; i < 9; i++ {
			for c := 0; c < 3; c++ {
				d2x[i][c] = d2[i][0]*hrot[0][c] + d2[i][1]*hrot[1][c] + d2[i][2]*hrot[2][c]
			}
		}
		// k += weight * (d2x @ E @ d2x^T).
		// tmp = d2x @ E (9x3).
		var tmp [9][3]float64
		for i := 0; i < 9; i++ {
			for c := 0; c < 3; c++ {
				tmp[i][c] = d2x[i][0]*hctEMat[0][c] + d2x[i][1]*hctEMat[1][c] + d2x[i][2]*hctEMat[2][c]
			}
		}
		for i := 0; i < 9; i++ {
			for jj := 0; jj < 9; jj++ {
				var s float64
				for c := 0; c < 3; c++ {
					s += tmp[i][c] * d2x[jj][c]
				}
				k[i][jj] += weight * s
			}
		}
	}
	// k = DOF_rot^T @ k @ DOF_rot, then scale by area.
	k = mat9Mul(mat9Mul(mat9T(dofRot), k), dofRot)
	for i := 0; i < 9; i++ {
		for jj := 0; jj < 9; jj++ {
			k[i][jj] *= area
		}
	}
	return k
}

// --- small matrix helpers ---

func mat2x2Mul(a, b [2][2]float64) [2][2]float64 {
	return [2][2]float64{
		{a[0][0]*b[0][0] + a[0][1]*b[1][0], a[0][0]*b[0][1] + a[0][1]*b[1][1]},
		{a[1][0]*b[0][0] + a[1][1]*b[1][0], a[1][0]*b[0][1] + a[1][1]*b[1][1]},
	}
}

func put2x2(m *[9][9]float64, r, c int, b [2][2]float64) {
	m[r][c] = b[0][0]
	m[r][c+1] = b[0][1]
	m[r+1][c] = b[1][0]
	m[r+1][c+1] = b[1][1]
}

func mat9Mul(a, b [9][9]float64) [9][9]float64 {
	var out [9][9]float64
	for i := 0; i < 9; i++ {
		for k := 0; k < 9; k++ {
			if a[i][k] == 0 {
				continue
			}
			for jj := 0; jj < 9; jj++ {
				out[i][jj] += a[i][k] * b[k][jj]
			}
		}
	}
	return out
}

func mat9T(a [9][9]float64) [9][9]float64 {
	var out [9][9]float64
	for i := 0; i < 9; i++ {
		for jj := 0; jj < 9; jj++ {
			out[i][jj] = a[jj][i]
		}
	}
	return out
}

// safeInv22 inverts a 2x2 matrix, returning the zero matrix when rank-deficient.
func safeInv22(m [2][2]float64) [2][2]float64 {
	prod1 := m[0][0] * m[1][1]
	delta := prod1 - m[0][1]*m[1][0]
	var di float64
	if math.Abs(delta) > 1e-8*math.Abs(prod1) {
		di = 1 / delta
	}
	return [2][2]float64{
		{m[1][1] * di, -m[0][1] * di},
		{-m[1][0] * di, m[0][0] * di},
	}
}

// pseudoInv22Sym returns the Moore–Penrose pseudo-inverse of a symmetric 2x2
// matrix, handling rank-1 and rank-0 cases (matplotlib _pseudo_inv22sym).
func pseudoInv22Sym(m [2][2]float64) [2][2]float64 {
	prod1 := m[0][0] * m[1][1]
	delta := prod1 - m[0][1]*m[1][0]
	if math.Abs(delta) > 1e-8*math.Abs(prod1) {
		return [2][2]float64{
			{m[1][1] / delta, -m[0][1] / delta},
			{-m[1][0] / delta, m[0][0] / delta},
		}
	}
	tr := m[0][0] + m[1][1]
	var sq float64
	if math.Abs(tr) >= 1e-8 {
		sq = 1 / (tr * tr)
	}
	return [2][2]float64{
		{m[0][0] * sq, m[0][1] * sq},
		{m[1][0] * sq, m[1][1] * sq},
	}
}

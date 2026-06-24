package qhull

import "math"

// This file ports the foundational, deterministic stages of Qhull's
// (qhull 8.0.2) 2-D Delaunay pipeline, faithful to its float64 arithmetic so the
// downstream point-creation order matches Qhull. The creation order ultimately
// fixes the cocircular diagonal choice (each Delaunay cell is fanned from its
// last-created vertex). Stages here: input projection to the paraboloid with the
// Qz "infinity" point, Qbb last-coordinate scaling, the qh_maxmin tolerance/
// extreme-point pass, and the qh_maxsimplex initial simplex. This is validated
// (build_test.go) but not yet wired into Delaunay: the incremental qh_buildhull
// loop that would consume it — and the cocircular coplanar-promotion needed for
// byte-for-byte parity — are deferred (PLAN.md "Phase 12"). The shipped Delaunay
// uses the robust exact-predicate engine in delaunay.go.
//
// Reference: third_party/qhull-8.0.2/src/libqhull_r/{geom2_r.c,poly2_r.c}.

// realEpsilon mirrors Qhull's REALepsilon (DBL_EPSILON).
const realEpsilon = 2.2204460492503131e-16

// qhRatioMaxsimplex mirrors qh_RATIOmaxsimplex (user_r.h).
const qhRatioMaxsimplex = 1.0e-3

// Our lifted hull_dim is always 3, which is below qh_INITIALmax (8), so the
// plain qh_maxsimplex path applies (no min/max-coordinate seeding loop).

// qstate holds the projected point set and the global tolerances/extreme-point
// data Qhull derives, for a 2-D Delaunay (lifted hull_dim = 3).
type qstate struct {
	// pts holds the projected 3-D points: the n input points (mean-subtracted,
	// lifted to the paraboloid and Qbb-scaled in the last coordinate) followed
	// by the single Qz infinity point at index n.
	pts [][3]float64
	n   int // number of real input points (infinity point is at index n)

	maxAbs    float64    // qh.MAXabs_coord
	maxWidth  float64    // qh.MAXwidth
	maxSum    float64    // qh.MAXsumcoord
	distRound float64    // qh.DISTround
	nearZero  [3]float64 // qh.NEARzero[k]

	minLast, maxLast float64 // qh.MINlastcoord / MAXlastcoord (pre-scaling)

	// maxpoints is the qh_maxmin extreme-point set: [min,max] point index per
	// dimension, in that append order (duplicates kept, as Qhull does).
	maxpoints []int
}

const hullDim = 3 // 2-D Delaunay lifts to a 3-D hull.

// project mirrors qh_projectinput with DELAUNAY+ATinfinity: it mean-subtracts
// (as matplotlib's wrapper does), lifts each point to the paraboloid, appends
// the Qz infinity point (xy centroid, last coord 1.1*maxboloid), then runs
// qh_maxmin and Qbb qh_scalelast. It returns the populated qstate.
func project(x, y []float64) *qstate {
	n := len(x)
	// matplotlib subtracts the coordinate mean before calling Qhull.
	var xm, ym float64
	for i := 0; i < n; i++ {
		xm += x[i]
		ym += y[i]
	}
	xm /= float64(n)
	ym /= float64(n)

	q := &qstate{n: n, pts: make([][3]float64, n+1)}
	var infX, infY, maxboloid float64
	for i := 0; i < n; i++ {
		px := x[i] - xm
		py := y[i] - ym
		paraboloid := px*px + py*py
		q.pts[i] = [3]float64{px, py, paraboloid}
		infX += px
		infY += py
		if paraboloid > maxboloid {
			maxboloid = paraboloid
		}
	}
	// Qz infinity point.
	q.pts[n] = [3]float64{infX / float64(n), infY / float64(n), maxboloid * 1.1}

	q.maxmin()
	q.scalelast()
	q.distRound = distround(hullDim, q.maxAbs, q.maxSum)
	return q
}

// maxmin mirrors qh_maxmin (SCALElast set): it finds the per-dimension extreme
// points, and accumulates MAXabs_coord, MAXwidth, MAXsumcoord and NEARzero. The
// last (paraboloid) coordinate is treated specially for SCALElast: its
// contribution to MAXabs/MAXwidth uses the running MAXabs_coord, but its extreme
// points and MIN/MAXlastcoord are still taken from the raw values.
func (q *qstate) maxmin() {
	np := q.n + 1
	q.maxWidth = -math.MaxFloat64
	q.maxpoints = q.maxpoints[:0]
	for k := 0; k < hullDim; k++ {
		minIdx, maxIdx := 0, 0
		for i := 0; i < np; i++ {
			if q.pts[maxIdx][k] < q.pts[i][k] {
				maxIdx = i
			} else if q.pts[minIdx][k] > q.pts[i][k] {
				minIdx = i
			}
		}
		if k == hullDim-1 {
			q.minLast = q.pts[minIdx][k]
			q.maxLast = q.pts[maxIdx][k]
		}
		var maxcoord float64
		if k == hullDim-1 { // SCALElast
			maxcoord = q.maxAbs
		} else {
			maxcoord = math.Max(q.pts[maxIdx][k], -q.pts[minIdx][k])
			if w := q.pts[maxIdx][k] - q.pts[minIdx][k]; w > q.maxWidth {
				q.maxWidth = w
			}
		}
		if maxcoord > q.maxAbs {
			q.maxAbs = maxcoord
		}
		q.maxSum += maxcoord
		q.maxpoints = append(q.maxpoints, minIdx, maxIdx)
		q.nearZero[k] = 80 * q.maxSum * realEpsilon
	}
}

// scalelast mirrors qh_scalelast (Qbb): scale the last coordinate from
// [minLast, maxLast] to [0, maxAbs].
func (q *qstate) scalelast() {
	low, high, newhigh := q.minLast, q.maxLast, q.maxAbs
	scale := newhigh / (high - low)
	shift := -low * scale
	for i := range q.pts {
		q.pts[i][2] = q.pts[i][2]*scale + shift
	}
}

// distround mirrors qh_distround.
func distround(dimension int, maxabs, maxsumabs float64) float64 {
	maxdistsum := math.Sqrt(float64(dimension)) * maxabs
	if maxsumabs < maxdistsum {
		maxdistsum = maxsumabs
	}
	return realEpsilon * (float64(dimension)*maxdistsum*1.01 + maxabs)
}

// det2 and det3 mirror Qhull's det2_/det3_ macros (geom_r.h).
func det2(a1, a2, b1, b2 float64) float64 { return a1*b2 - a2*b1 }

func det3(a1, a2, a3, b1, b2, b3, c1, c2, c3 float64) float64 {
	return a1*det2(b2, b3, c2, c3) - a2*det2(b1, b3, c1, c3) + a3*det2(b1, b2, c1, c2)
}

// detsimplex mirrors qh_detsimplex+qh_determinant for dim 2 or 3: the signed
// determinant of the simplex (apex, base...) using the first dim coordinates,
// with the nearzero flag from NEARzero.
func (q *qstate) detsimplex(apex int, base []int, dim int) (det float64, nearzero bool) {
	a := q.pts[apex]
	switch dim {
	case 2:
		r0, r1 := q.pts[base[0]], q.pts[base[1]]
		det = det2(r0[0]-a[0], r0[1]-a[1], r1[0]-a[0], r1[1]-a[1])
		nearzero = math.Abs(det) < 10*q.nearZero[1]
	case 3:
		r0, r1, r2 := q.pts[base[0]], q.pts[base[1]], q.pts[base[2]]
		det = det3(
			r0[0]-a[0], r0[1]-a[1], r0[2]-a[2],
			r1[0]-a[0], r1[1]-a[1], r1[2]-a[2],
			r2[0]-a[0], r2[1]-a[1], r2[2]-a[2],
		)
		nearzero = math.Abs(det) < 10*q.nearZero[2]
	}
	return det, nearzero
}

// maxsimplex mirrors qh_maxsimplex for hull_dim < qh_INITIALmax: it builds a
// (hullDim+1)-point initial simplex by greedily maximizing the determinant,
// preferring the maxpoints extreme set and falling back to all points. The
// returned point indices are in Qhull's append order, which fixes the first
// vertex ids (creation order).
func (q *qstate) maxsimplex() []int {
	simplex := make([]int, 0, hullDim+1)
	var maxdet float64

	// Seed with the min/max first-coordinate points among maxpoints.
	maxcoord, mincoord := -math.MaxFloat64, math.MaxFloat64
	maxx, minx := -1, -1
	for _, p := range q.maxpoints {
		if q.pts[p][0] > maxcoord {
			maxcoord = q.pts[p][0]
			maxx = p
		}
		if q.pts[p][0] < mincoord {
			mincoord = q.pts[p][0]
			minx = p
		}
	}
	maxdet = maxcoord - mincoord
	simplex = setUnique(simplex, minx)
	if len(simplex) < 2 {
		simplex = setUnique(simplex, maxx)
	}

	for i := len(simplex); i < hullDim+1; i++ {
		prevdet := maxdet
		maxpoint := -1
		maxdet = -1.0
		maxnearzero := false
		for _, p := range q.maxpoints {
			if setIn(simplex, p) || p == maxpoint {
				continue
			}
			det, nz := q.detsimplex(p, simplex, i)
			if det = math.Abs(det); det > maxdet {
				maxdet, maxpoint, maxnearzero = det, p, nz
			}
		}
		targetdet := prevdet * q.maxWidth
		mindet := 10 * qhRatioMaxsimplex * targetdet
		maybeFalseNarrow := maxdet > 0.0 && maxdet/targetdet < qhRatioMaxsimplex
		if maxpoint < 0 || maxnearzero || maybeFalseNarrow {
			for p := 0; p <= q.n; p++ {
				if setIn(q.maxpoints, p) || setIn(simplex, p) {
					continue
				}
				det, nz := q.detsimplex(p, simplex, i)
				if det = math.Abs(det); det > maxdet {
					maxdet, maxpoint, maxnearzero = det, p, nz
					if !maxnearzero && maxdet > mindet {
						break
					}
				}
			}
		}
		simplex = append(simplex, maxpoint)
	}
	return simplex
}

func setIn(set []int, v int) bool {
	for _, e := range set {
		if e == v {
			return true
		}
	}
	return false
}

func setUnique(set []int, v int) []int {
	if v < 0 || setIn(set, v) {
		return set
	}
	return append(set, v)
}

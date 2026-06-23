package qhull

import "math/big"

// Exact geometric predicates over math/big.Rat. IEEE-754 doubles are exactly
// representable as rationals, so these return the mathematically correct sign of
// the orientation and in-circle determinants with no epsilon. They yield the
// true (unique, for general-position inputs) Delaunay triangulation, which is
// the same connectivity Qhull produces for those inputs.
//
// These are used by the baseline incremental construction; the Qhull-faithful
// cocircular tie-breaking is layered on top (see delaunay.go).

func bigRat(v float64) *big.Rat {
	r := new(big.Rat)
	r.SetFloat64(v)
	return r
}

// orient2d returns the sign (-1, 0, +1) of the signed area of (a,b,c); +1 when
// the points wind anticlockwise.
func orient2d(ax, ay, bx, by, cx, cy float64) int {
	bax := new(big.Rat).Sub(bigRat(bx), bigRat(ax))
	bay := new(big.Rat).Sub(bigRat(by), bigRat(ay))
	cax := new(big.Rat).Sub(bigRat(cx), bigRat(ax))
	cay := new(big.Rat).Sub(bigRat(cy), bigRat(ay))
	left := new(big.Rat).Mul(bax, cay)
	right := new(big.Rat).Mul(bay, cax)
	return left.Sub(left, right).Sign()
}

// inCircle returns the sign (-1, 0, +1) of the in-circle determinant for the
// anticlockwise triangle (a,b,c). +1 means d is strictly inside the
// circumcircle; 0 means the four points are cocircular.
func inCircle(ax, ay, bx, by, cx, cy, dx, dy float64) int {
	adx := new(big.Rat).Sub(bigRat(ax), bigRat(dx))
	ady := new(big.Rat).Sub(bigRat(ay), bigRat(dy))
	bdx := new(big.Rat).Sub(bigRat(bx), bigRat(dx))
	bdy := new(big.Rat).Sub(bigRat(by), bigRat(dy))
	cdx := new(big.Rat).Sub(bigRat(cx), bigRat(dx))
	cdy := new(big.Rat).Sub(bigRat(cy), bigRat(dy))

	ad2 := new(big.Rat).Add(new(big.Rat).Mul(adx, adx), new(big.Rat).Mul(ady, ady))
	bd2 := new(big.Rat).Add(new(big.Rat).Mul(bdx, bdx), new(big.Rat).Mul(bdy, bdy))
	cd2 := new(big.Rat).Add(new(big.Rat).Mul(cdx, cdx), new(big.Rat).Mul(cdy, cdy))

	// | adx ady ad2 |
	// | bdx bdy bd2 |
	// | cdx cdy cd2 |
	t1 := new(big.Rat).Mul(adx, new(big.Rat).Sub(new(big.Rat).Mul(bdy, cd2), new(big.Rat).Mul(cdy, bd2)))
	t2 := new(big.Rat).Mul(ady, new(big.Rat).Sub(new(big.Rat).Mul(bdx, cd2), new(big.Rat).Mul(cdx, bd2)))
	t3 := new(big.Rat).Mul(ad2, new(big.Rat).Sub(new(big.Rat).Mul(bdx, cdy), new(big.Rat).Mul(cdx, bdy)))
	det := t1.Sub(t1, t2).Add(t1, t3)
	return det.Sign()
}

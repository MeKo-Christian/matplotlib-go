package tri

import "math"

// TriAnalyzer provides tools for analysing and improving a triangular mesh,
// mirroring matplotlib's TriAnalyzer.
type TriAnalyzer struct {
	tri Triangulation
}

// NewTriAnalyzer returns an analyzer for t.
func NewTriAnalyzer(t Triangulation) *TriAnalyzer {
	return &TriAnalyzer{tri: t}
}

// ScaleFactors returns factors (kx, ky) that rescale the triangulation so its
// unmasked triangles fit exactly inside a unit square.
func (a *TriAnalyzer) ScaleFactors() (kx, ky float64) {
	used := a.usedNodes()
	var xs, ys []float64
	for i, ok := range used {
		if ok {
			xs = append(xs, a.tri.X[i])
			ys = append(ys, a.tri.Y[i])
		}
	}
	return 1 / ptp(xs), 1 / ptp(ys)
}

func (a *TriAnalyzer) usedNodes() []bool {
	used := make([]bool, len(a.tri.X))
	for i, tr := range a.tri.Triangles {
		if a.tri.Masked(i) {
			continue
		}
		used[tr[0]] = true
		used[tr[1]] = true
		used[tr[2]] = true
	}
	return used
}

// CircleRatios returns, per triangle, the ratio of incircle to circumcircle
// radius — a flatness measure that is 0.5 for equilateral triangles and near 0
// for slivers. Masked triangles yield NaN. When rescale is true the mesh is
// first scaled to a unit square (the default in matplotlib).
func (a *TriAnalyzer) CircleRatios(rescale bool) []float64 {
	kx, ky := 1.0, 1.0
	if rescale {
		kx, ky = a.ScaleFactors()
	}
	out := make([]float64, len(a.tri.Triangles))
	for i, tr := range a.tri.Triangles {
		if a.tri.Masked(i) {
			out[i] = math.NaN()
			continue
		}
		x0, y0 := a.tri.X[tr[0]]*kx, a.tri.Y[tr[0]]*ky
		x1, y1 := a.tri.X[tr[1]]*kx, a.tri.Y[tr[1]]*ky
		x2, y2 := a.tri.X[tr[2]]*kx, a.tri.Y[tr[2]]*ky
		la := math.Hypot(x1-x0, y1-y0)
		lb := math.Hypot(x2-x1, y2-y1)
		lc := math.Hypot(x0-x2, y0-y2)
		s := (la + lb + lc) * 0.5
		prod := s * (la + lb - s) * (la + lc - s) * (lb + lc - s)
		abc := la * lb * lc
		var circum float64
		if prod <= 0 {
			circum = math.Inf(1)
		} else {
			circum = abc / (4 * math.Sqrt(prod))
		}
		inR := abc / (4 * circum * s)
		out[i] = inR / circum
	}
	return out
}

// FlatTriMask returns a mask that removes excessively flat border triangles
// (circle ratio < minCircleRatio) from the triangulation. Triangles are masked
// iteratively, only when they touch the current mesh border, so no interior
// holes are created. Initially masked triangles remain masked. Mirrors
// matplotlib's TriAnalyzer.get_flat_tri_mask.
func (a *TriAnalyzer) FlatTriMask(minCircleRatio float64, rescale bool) []bool {
	ntri := len(a.tri.Triangles)
	ratios := a.CircleRatios(rescale)
	maskBad := make([]bool, ntri)
	for i := range ratios {
		maskBad[i] = ratios[i] < minCircleRatio // NaN comparisons are false
	}

	current := make([]bool, ntri)
	if len(a.tri.Mask) == ntri {
		copy(current, a.tri.Mask)
	}

	valid := a.tri.Neighbors() // mutable copy

	for {
		nadd := 0
		// Wavefront: unmasked triangles with at least one border (-1) neighbour.
		added := make([]bool, ntri)
		for i := 0; i < ntri; i++ {
			if current[i] {
				continue
			}
			border := false
			for _, nb := range valid[i] {
				if nb == -1 {
					border = true
					break
				}
			}
			if border && maskBad[i] {
				added[i] = true
			}
		}
		for i := 0; i < ntri; i++ {
			if added[i] {
				current[i] = true
				nadd++
			}
		}
		if nadd == 0 {
			break
		}
		// Update neighbour table: newly masked triangles no longer connect.
		for i := 0; i < ntri; i++ {
			if added[i] {
				valid[i] = [3]int{-1, -1, -1}
			}
		}
		for i := 0; i < ntri; i++ {
			for j := 0; j < 3; j++ {
				if nb := valid[i][j]; nb != -1 && current[nb] {
					valid[i][j] = -1
				}
			}
		}
	}
	return current
}

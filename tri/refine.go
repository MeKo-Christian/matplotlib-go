package tri

// UniformTriRefiner refines a triangulation by recursively splitting each
// triangle into four child triangles on its edge midpoints. It mirrors
// matplotlib's UniformTriRefiner.
type UniformTriRefiner struct {
	tri Triangulation
}

// NewUniformTriRefiner returns a refiner for t.
func NewUniformTriRefiner(t Triangulation) *UniformTriRefiner {
	return &UniformTriRefiner{tri: t}
}

// RefineTriangulation returns a uniformly refined triangulation: each triangle
// is divided into 4**subdiv child triangles. When returnTriIndex is true, the
// second result maps every refined node to the index of an original
// (ancestor) triangle that contains it (-1 if none); otherwise it is nil.
func (r *UniformTriRefiner) RefineTriangulation(subdiv int, returnTriIndex bool) (Triangulation, []int) {
	refi := r.tri
	ancestors := make([]int, len(refi.Triangles))
	for i := range ancestors {
		ancestors[i] = i
	}
	for n := 0; n < subdiv; n++ {
		refi, ancestors = refineOnce(refi, ancestors)
	}

	if !returnTriIndex {
		return refi, nil
	}

	foundIndex := make([]int, len(refi.X))
	for i := range foundIndex {
		foundIndex[i] = -1
	}
	// Assign masked ancestors first, then overwrite with unmasked ones so that
	// a refined node prefers an unmasked containing triangle (matplotlib).
	assign := func(wantMasked bool) {
		for childIdx, child := range refi.Triangles {
			anc := ancestors[childIdx]
			if r.tri.Masked(anc) != wantMasked {
				continue
			}
			for _, node := range child {
				foundIndex[node] = anc
			}
		}
	}
	if len(r.tri.Mask) > 0 {
		assign(true)
	}
	assign(false)
	return refi, foundIndex
}

// RefineField refines the triangulation and interpolates the field z onto the
// refined nodes. If interp is nil, a CubicTriInterpolator (min_E) is used.
func (r *UniformTriRefiner) RefineField(z []float64, interp TriInterpolator, subdiv int) (Triangulation, []float64, error) {
	if interp == nil {
		ci, err := NewCubicTriInterpolator(r.tri, z, CubicMinE, nil, nil)
		if err != nil {
			return Triangulation{}, nil, err
		}
		interp = ci
	}
	refi, foundIndex := r.RefineTriangulation(subdiv, true)
	hinter, hasHint := interp.(triIndexInterpolator)
	refiZ := make([]float64, len(refi.X))
	for i := range refi.X {
		var v float64
		var ok bool
		// Prefer the ancestor-triangle hint so nodes that fall exactly on a mesh
		// boundary still interpolate (matplotlib passes tri_index for this).
		if hasHint && foundIndex[i] >= 0 {
			v, ok = hinter.interpolateAt(foundIndex[i], refi.X[i], refi.Y[i])
		}
		if !ok {
			v, ok = interp.Interpolate(refi.X[i], refi.Y[i])
		}
		if ok {
			refiZ[i] = v
		}
	}
	return refi, refiZ, nil
}

// refineOnce splits every triangle into four on its edge midpoints, sharing
// midpoints between adjacent triangles. Masks and ancestor indices are
// replicated to the four children. Mirrors _refine_triangulation_once.
func refineOnce(t Triangulation, ancestors []int) (Triangulation, []int) {
	refiX := append([]float64(nil), t.X...)
	refiY := append([]float64(nil), t.Y...)
	midIdx := make(map[[2]int]int, len(t.Triangles)*3)
	midpoint := func(a, b int) int {
		key := sortedEdge(a, b)
		if idx, ok := midIdx[key]; ok {
			return idx
		}
		idx := len(refiX)
		refiX = append(refiX, (t.X[a]+t.X[b])/2)
		refiY = append(refiY, (t.Y[a]+t.Y[b])/2)
		midIdx[key] = idx
		return idx
	}

	childTris := make([][3]int, 0, len(t.Triangles)*4)
	for _, tr := range t.Triangles {
		v0, v1, v2 := tr[0], tr[1], tr[2]
		m0 := midpoint(v0, v1)
		m1 := midpoint(v1, v2)
		m2 := midpoint(v2, v0)
		childTris = append(
			childTris,
			[3]int{v0, m0, m2},
			[3]int{v1, m1, m0},
			[3]int{v2, m2, m1},
			[3]int{m0, m1, m2},
		)
	}

	child := Triangulation{X: refiX, Y: refiY, Triangles: childTris}
	if len(t.Mask) > 0 {
		child.Mask = make([]bool, len(childTris))
		for i, m := range t.Mask {
			for k := 0; k < 4; k++ {
				child.Mask[4*i+k] = m
			}
		}
	}
	var childAncestors []int
	if ancestors != nil {
		childAncestors = make([]int, len(childTris))
		for i, a := range ancestors {
			for k := 0; k < 4; k++ {
				childAncestors[4*i+k] = a
			}
		}
	}
	return child, childAncestors
}

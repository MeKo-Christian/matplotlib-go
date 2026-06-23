package tri

import "fmt"

// TriInterpolator interpolates a scalar field defined at the nodes of a
// triangulation. It mirrors matplotlib's matplotlib.tri.TriInterpolator API.
type TriInterpolator interface {
	// Interpolate returns the interpolated value at (x, y). ok is false when
	// the point lies outside the triangulation.
	Interpolate(x, y float64) (value float64, ok bool)
	// Gradient returns the field gradient (dz/dx, dz/dy) at (x, y). ok is false
	// when the point lies outside the triangulation.
	Gradient(x, y float64) (dzdx, dzdy float64, ok bool)
}

// triIndexInterpolator is an optional capability: interpolate within a known
// original triangle, bypassing point location. RefineField uses it for refined
// nodes that fall exactly on a mesh boundary.
type triIndexInterpolator interface {
	interpolateAt(origTri int, x, y float64) (float64, bool)
}

// LinearTriInterpolator performs linear (C0) interpolation: within each
// triangle the field is the plane through its three vertex values. The
// gradient is therefore constant per triangle. It matches matplotlib's
// LinearTriInterpolator.
type LinearTriInterpolator struct {
	tri    Triangulation
	finder TriFinder
	plane  [][3]float64
}

// NewLinearTriInterpolator builds a linear interpolator for the field z (one
// value per triangulation node).
func NewLinearTriInterpolator(t Triangulation, z []float64) (*LinearTriInterpolator, error) {
	if len(z) != len(t.X) {
		return nil, fmt.Errorf("interpolator requires one z per node: got %d, want %d", len(z), len(t.X))
	}
	plane, err := t.PlaneCoefficients(z)
	if err != nil {
		return nil, err
	}
	return &LinearTriInterpolator{tri: t, finder: t.TriFinder(), plane: plane}, nil
}

// Interpolate returns z at (x, y), or ok=false outside the mesh.
func (li *LinearTriInterpolator) Interpolate(x, y float64) (float64, bool) {
	i := li.finder.Find(x, y)
	if i < 0 || i >= len(li.plane) {
		return 0, false
	}
	c := li.plane[i]
	return c[0]*x + c[1]*y + c[2], true
}

// interpolateAt evaluates within a specific original triangle, bypassing the
// point locator. Used by RefineField for nodes that fall exactly on mesh
// boundaries. It implements triIndexInterpolator.
func (li *LinearTriInterpolator) interpolateAt(origTri int, x, y float64) (float64, bool) {
	if origTri < 0 || origTri >= len(li.plane) {
		return 0, false
	}
	c := li.plane[origTri]
	return c[0]*x + c[1]*y + c[2], true
}

// Gradient returns (dz/dx, dz/dy) at (x, y), constant within each triangle.
func (li *LinearTriInterpolator) Gradient(x, y float64) (float64, float64, bool) {
	i := li.finder.Find(x, y)
	if i < 0 || i >= len(li.plane) {
		return 0, 0, false
	}
	c := li.plane[i]
	return c[0], c[1], true
}

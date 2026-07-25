// Package contourgeom exposes shared contour geometry to repository-internal
// consumers without making the algorithms part of core's public API.
package contourgeom

import "github.com/cwbudde/matplotlib-go/geom"

// Triangulation is the geometry-only contour view of a triangulation.
type Triangulation struct {
	X         []float64
	Y         []float64
	Triangles [][3]int
	Mask      []bool
}

// Provider supplies the canonical contour algorithms owned by core.
type Provider struct {
	Levels              func(values, explicit []float64, levelCount int, filled bool) []float64
	Polylines           func(tri *Triangulation, values, levels []float64) ([][]geom.Pt, []float64)
	GridPolylines       func(x, y []float64, data [][]float64, levels []float64) ([][]geom.Pt, []float64)
	CellBandPolygons    func(points [4]geom.Pt, values [4]float64, low, high float64) [][]geom.Pt
	TriangleBandPolygon func(points [3]geom.Pt, values [3]float64, low, high float64) []geom.Pt
}

var provider Provider

// Register installs the canonical contour geometry implementation.
func Register(p Provider) {
	provider = p
}

// Levels resolves explicit or automatically located contour levels.
func Levels(values, explicit []float64, levelCount int, filled bool) []float64 {
	return provider.Levels(values, explicit, levelCount, filled)
}

// Polylines builds triangulated contour polylines.
func Polylines(tri *Triangulation, values, levels []float64) ([][]geom.Pt, []float64) {
	return provider.Polylines(tri, values, levels)
}

// GridPolylines builds structured-grid contour polylines.
func GridPolylines(x, y []float64, data [][]float64, levels []float64) ([][]geom.Pt, []float64) {
	return provider.GridPolylines(x, y, data, levels)
}

// CellBandPolygons clips one structured cell to a contour band.
func CellBandPolygons(points [4]geom.Pt, values [4]float64, low, high float64) [][]geom.Pt {
	return provider.CellBandPolygons(points, values, low, high)
}

// TriangleBandPolygon clips one triangle to a contour band.
func TriangleBandPolygon(points [3]geom.Pt, values [3]float64, low, high float64) []geom.Pt {
	return provider.TriangleBandPolygon(points, values, low, high)
}

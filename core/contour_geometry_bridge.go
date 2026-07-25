package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/contourgeom"
)

func init() {
	contourgeom.Register(contourgeom.Provider{
		Levels: contourLevels,
		Polylines: func(tri *contourgeom.Triangulation, values, levels []float64) ([][]geom.Pt, []float64) {
			return contourPolylines(Triangulation{
				X:         tri.X,
				Y:         tri.Y,
				Triangles: tri.Triangles,
				Mask:      tri.Mask,
			}, values, levels)
		},
		GridPolylines:       contourGridPolylines,
		CellBandPolygons:    contourCellBandPolygons,
		TriangleBandPolygon: triangleBandPolygon,
	})
}

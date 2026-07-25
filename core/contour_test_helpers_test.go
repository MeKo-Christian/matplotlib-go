package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

func pointsEqual(got, want []geom.Pt, tol float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if math.Abs(got[i].X-want[i].X) > tol || math.Abs(got[i].Y-want[i].Y) > tol {
			return false
		}
	}
	return true
}

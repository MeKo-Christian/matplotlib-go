package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func rectFromPoints(p1, p2 geom.Pt) (geom.Rect, bool) {
	minX := math.Min(p1.X, p2.X)
	maxX := math.Max(p1.X, p2.X)
	minY := math.Min(p1.Y, p2.Y)
	maxY := math.Max(p1.Y, p2.Y)
	if maxX <= minX || maxY <= minY {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}, true
}

func snappedFillRectPath(rect geom.Rect) geom.Path {
	return snappedRectPath(rect, false)
}

func snappedStrokeRectPath(rect geom.Rect) geom.Path {
	return snappedRectPath(rect, true)
}

func snappedRectPath(rect geom.Rect, centerOnPixels bool) geom.Path {
	snapped, ok := snapPixelRect(rect, centerOnPixels)
	if !ok {
		return geom.Path{}
	}
	return pixelRectPath(snapped)
}

func snapPixelRect(rect geom.Rect, centerOnPixels bool) (geom.Rect, bool) {
	minX, maxX, okX := snapRectAxis(rect.Min.X, rect.Max.X, centerOnPixels, 1)
	// Display space is y-up; the backend flips at the device boundary. The
	// half-pixel centering offset must point the opposite way for Y so stroked
	// rectangle edges land on the same device pixel after the flip.
	minY, maxY, okY := snapRectAxis(rect.Min.Y, rect.Max.Y, centerOnPixels, -1)
	if !okX || !okY {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}, true
}

func snapRectAxis(minVal, maxVal float64, centerOnPixels bool, offsetSign float64) (float64, float64, bool) {
	if maxVal <= minVal {
		return 0, 0, false
	}

	snap := math.Round
	offset := 0.0
	if centerOnPixels {
		offset = 0.5 * offsetSign
	}

	minSnap := snap(minVal) + offset
	maxSnap := snap(maxVal) + offset
	if maxSnap <= minSnap {
		maxSnap = minSnap + 1
	}
	return minSnap, maxSnap, true
}

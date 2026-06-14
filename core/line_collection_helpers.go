package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// HLines adds horizontal data-space line segments, broadcasting single-value
// extents in the same spirit as Matplotlib's Axes.hlines.
func (a *Axes) HLines(y, xMin, xMax []float64, opts ...LineCollection) *LineCollection {
	if a == nil || !lineCollectionInputLengthsValid(len(y), len(xMin), len(xMax)) {
		return nil
	}
	segments := make([][]geom.Pt, len(y))
	for i := range y {
		segments[i] = []geom.Pt{
			{X: lineCollectionValueAt(xMin, i), Y: y[i]},
			{X: lineCollectionValueAt(xMax, i), Y: y[i]},
		}
	}
	return a.lineCollectionFromSegments(segments, opts...)
}

// VLines adds vertical data-space line segments, broadcasting single-value
// extents in the same spirit as Matplotlib's Axes.vlines.
func (a *Axes) VLines(x, yMin, yMax []float64, opts ...LineCollection) *LineCollection {
	if a == nil || !lineCollectionInputLengthsValid(len(x), len(yMin), len(yMax)) {
		return nil
	}
	segments := make([][]geom.Pt, len(x))
	for i := range x {
		segments[i] = []geom.Pt{
			{X: x[i], Y: lineCollectionValueAt(yMin, i)},
			{X: x[i], Y: lineCollectionValueAt(yMax, i)},
		}
	}
	return a.lineCollectionFromSegments(segments, opts...)
}

func (a *Axes) lineCollectionFromSegments(segments [][]geom.Pt, opts ...LineCollection) *LineCollection {
	collection := LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Alpha:  1,
		},
		Segments:  segments,
		Color:     a.NextColor(),
		LineWidth: 1,
		LineCap:   render.CapButt,
	}
	if len(opts) > 0 {
		collection = opts[0]
		collection.Segments = segments
		if collection.Coords == (CoordinateSpec{}) {
			collection.Coords = Coords(CoordData)
		}
		if collection.Alpha == 0 {
			collection.Alpha = 1
		}
		if collection.Color.A == 0 && len(collection.Colors) == 0 {
			collection.Color = a.NextColor()
		}
		if collection.LineWidth == 0 && len(collection.LineWidths) == 0 {
			collection.LineWidth = 1
		}
	}
	a.AddCollection(&collection)
	return &collection
}

func lineCollectionInputLengthsValid(primary, minLen, maxLen int) bool {
	return primary > 0 && (minLen == 1 || minLen == primary) && (maxLen == 1 || maxLen == primary)
}

func lineCollectionValueAt(values []float64, i int) float64 {
	if len(values) == 1 {
		return values[0]
	}
	return values[i]
}

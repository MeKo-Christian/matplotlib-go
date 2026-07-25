package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

// LineCollectionOptions configures Axes.HLines and Axes.VLines.
//
// It mirrors the keyword arguments Matplotlib's hlines/vlines accept, plus the
// LineCollection kwargs they forward. Segment endpoints come from the call's
// positional arguments, so unlike the LineCollection artist this type cannot
// describe the geometry — building a collection by hand and passing it to
// [Axes.AddCollection] remains available for everything this type omits.
type LineCollectionOptions struct {
	// Coords selects the coordinate space of the segment endpoints. The zero
	// value is data coordinates.
	Coords CoordinateSpec
	// Label is the legend entry. Empty adds none.
	Label string
	// Alpha is the collection opacity in [0,1]. Unset is fully opaque.
	Alpha optional.Value[float64]
	// Color is the color shared by every segment. Unset takes the next color
	// from the axes property cycle, unless Colors is non-empty.
	Color optional.Value[render.Color]
	// Colors assigns a color per segment and overrides Color.
	Colors []render.Color
	// LineWidth is the width in pixels shared by every segment. Unset is 1,
	// unless LineWidths is non-empty.
	LineWidth optional.Value[float64]
	// LineWidths assigns a width per segment and overrides LineWidth.
	LineWidths []float64
	// LineStyle and LineStyles accept Matplotlib linestyle strings ("-", "--",
	// "-.", ":", or the named "solid"/"dashed"/"dashdot"/"dotted") and are
	// converted to width-scaled dash patterns at draw time. Explicit numeric
	// Dashes/DashPatterns take precedence; a per-segment LineStyles entry
	// overrides the scalar LineStyle.
	LineStyle    string
	LineStyles   []string
	Dashes       []float64
	DashPatterns [][]float64
	LineCap      render.LineCap
	LineJoin     render.LineJoin
	Antialias    render.AntialiasMode
	PathEffects  []render.PathEffect
}

// HLines adds horizontal data-space line segments, broadcasting single-value
// extents in the same spirit as Matplotlib's Axes.hlines.
//
//nolint:gocritic // LineCollectionOptions is an immutable snapshot of the caller's options.
func (a *Axes) HLines(y, xMin, xMax []float64, opt LineCollectionOptions) *LineCollection {
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
	return a.lineCollectionFromSegments(segments, opt)
}

// VLines adds vertical data-space line segments, broadcasting single-value
// extents in the same spirit as Matplotlib's Axes.vlines.
//
//nolint:gocritic // LineCollectionOptions is an immutable snapshot of the caller's options.
func (a *Axes) VLines(x, yMin, yMax []float64, opt LineCollectionOptions) *LineCollection {
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
	return a.lineCollectionFromSegments(segments, opt)
}

//nolint:gocritic // LineCollectionOptions is read-only here; the artist is built from a copy.
func (a *Axes) lineCollectionFromSegments(segments [][]geom.Pt, opt LineCollectionOptions) *LineCollection {
	// The scalar Color and LineWidth only fall back to their defaults when the
	// per-segment slice that would override them is empty, so a caller that
	// supplies Colors does not advance the property cycle.
	color, colorSet := opt.Color.Get()
	if !colorSet && len(opt.Colors) == 0 {
		color = a.NextColor()
	}
	lineWidth, lineWidthSet := opt.LineWidth.Get()
	if !lineWidthSet && len(opt.LineWidths) == 0 {
		lineWidth = 1
	}

	collection := LineCollection{
		Collection: Collection{
			Coords:      opt.Coords,
			Label:       opt.Label,
			Alpha:       opt.Alpha.Or(1),
			Antialias:   opt.Antialias,
			PathEffects: opt.PathEffects,
		},
		Segments:     segments,
		Colors:       opt.Colors,
		Color:        color,
		LineWidths:   opt.LineWidths,
		LineWidth:    lineWidth,
		DashPatterns: opt.DashPatterns,
		Dashes:       opt.Dashes,
		LineStyle:    opt.LineStyle,
		LineStyles:   opt.LineStyles,
		LineJoin:     opt.LineJoin,
		LineCap:      opt.LineCap,
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

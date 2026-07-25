package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/optarg"
	"github.com/cwbudde/matplotlib-go/render"
)

// EventOrientation controls whether eventplot lines are drawn vertically or horizontally.
type EventOrientation uint8

const (
	EventOrientationVertical EventOrientation = iota
	EventOrientationHorizontal
)

// EventPlotOptions configures Axes.Eventplot.
type EventPlotOptions struct {
	Orientation EventOrientation
	LineOffsets []float64
	LineLengths []float64
	Colors      []render.Color
	LineWidth   float64
	Alpha       float64
	Label       string
}

// EventCollection is a thin LineCollection wrapper used for eventplot.
type EventCollection struct {
	LineCollection
	Positions   [][]float64
	Orientation EventOrientation
	LineOffsets []float64
	LineLengths []float64
}

// Eventplot adds a collection of short event markers to the axes.
func (a *Axes) Eventplot(positions [][]float64, opts ...EventPlotOptions) *EventCollection {
	if a == nil || len(positions) == 0 {
		return nil
	}
	cfg := EventPlotOptions{
		Orientation: EventOrientationVertical,
		LineWidth:   1.5,
		Alpha:       1,
	}
	if supplied, ok := optarg.Optional("eventplot", opts); ok {
		cfg = supplied
		if cfg.LineWidth <= 0 {
			cfg.LineWidth = 1.5
		}
		if cfg.Alpha <= 0 {
			cfg.Alpha = 1
		}
	}

	offsets := cfg.LineOffsets
	if len(offsets) == 0 {
		offsets = make([]float64, len(positions))
		for i := range offsets {
			offsets[i] = float64(i + 1)
		}
	}
	lengths := cfg.LineLengths
	if len(lengths) == 0 {
		lengths = make([]float64, len(positions))
		for i := range lengths {
			lengths[i] = 0.8
		}
	}

	segments := make([][]geom.Pt, 0)
	colors := make([]render.Color, 0)
	for i, group := range positions {
		offset := floatAt(offsets, i, float64(i+1))
		length := math.Abs(floatAt(lengths, i, 0.8))
		half := length / 2
		col := colorAt(a.NextColor(), cfg.Colors, i)
		col = col.WithAlphaMultiplier(clampOneToOne(cfg.Alpha))
		for _, value := range group {
			if !isFinite(value) {
				continue
			}
			segment := []geom.Pt{
				{X: value, Y: offset - half},
				{X: value, Y: offset + half},
			}
			if cfg.Orientation == EventOrientationHorizontal {
				segment = []geom.Pt{
					{X: offset - half, Y: value},
					{X: offset + half, Y: value},
				}
			}
			segments = append(segments, segment)
			colors = append(colors, col)
		}
	}
	if len(segments) == 0 {
		return nil
	}

	collection := &EventCollection{
		LineCollection: LineCollection{
			Collection: Collection{Label: cfg.Label, Alpha: 1, z: 2},
			Segments:   segments,
			Colors:     colors,
			LineWidth:  cfg.LineWidth,
			LineCap:    render.CapButt,
			LineJoin:   render.JoinRound,
		},
		Positions:   positions,
		Orientation: cfg.Orientation,
		LineOffsets: append([]float64(nil), offsets...),
		LineLengths: append([]float64(nil), lengths...),
	}
	a.AddCollection(collection)
	return collection
}

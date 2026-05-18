package core

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// LineDrawStyle controls how consecutive data points are connected.
type LineDrawStyle uint8

const (
	LineDrawStyleDefault LineDrawStyle = iota
	LineDrawStyleStepsPre
	LineDrawStyleStepsMid
	LineDrawStyleStepsPost
)

// DashUnits describes the unit system used by Line2D.Dashes.
type DashUnits uint8

const (
	// DashUnitsPixels interprets Dashes as renderer pixel lengths. This keeps
	// existing direct field assignments and backend-facing paint unchanged.
	DashUnitsPixels DashUnits = iota
	// DashUnitsMatplotlib interprets Dashes like Matplotlib Line2D.set_dashes:
	// input lengths are in points and, with the default scale_dashes=True, are
	// scaled by the line width before rasterization.
	DashUnitsMatplotlib
)

// Line2D is a minimal polyline artist (stroke only).
type Line2D struct {
	XY          []geom.Pt    // data space points
	W           float64      // stroke width (px for now)
	Col         render.Color // stroke color
	Dashes      []float64    // dash pattern (on/off pairs)
	DashUnits   DashUnits    // unit system for Dashes
	PathEffects []render.PathEffect
	DrawStyle   LineDrawStyle // optional step-style connection mode
	Label       string        // series label for legend
	z           float64       // z-order
}

// SetDashes sets the dash sequence using Matplotlib Line2D.set_dashes units.
func (l *Line2D) SetDashes(seq ...float64) {
	if len(seq) == 0 {
		l.Dashes = nil
		l.DashUnits = DashUnitsPixels
		return
	}
	l.Dashes = append([]float64(nil), seq...)
	l.DashUnits = DashUnitsMatplotlib
}

// Draw renders the line by transforming points to pixel space and drawing a path.
func (l *Line2D) Draw(r render.Renderer, ctx *DrawContext) {
	points := l.pathPoints()
	if len(points) == 0 {
		return // nothing to draw
	}

	p := geom.Path{}
	for i, v := range points {
		q := (&ctx.DataToPixel).Apply(v)
		if i == 0 {
			p.C = append(p.C, geom.MoveTo)
		} else {
			p.C = append(p.C, geom.LineTo)
		}
		p.V = append(p.V, q)
	}

	paint := render.Paint{
		LineWidth:   l.W,
		LineJoin:    render.JoinRound, // Default to round joins
		LineCap:     render.CapButt,   // Default to butt caps
		MiterLimit:  10.0,             // Standard miter limit
		Stroke:      l.Col,
		Dashes:      lineDashesForPaint(l.Dashes, l.W, l.DashUnits),
		PathEffects: append([]render.PathEffect(nil), l.PathEffects...),
		Snap:        render.SnapAuto,
		Simplify:    ctx != nil && ctx.RC.PathSimplify,
	}
	if ctx != nil {
		paint.SimplifyThreshold = ctx.RC.PathSimplifyThreshold
		paint.MaxChunkVertices = ctx.RC.AggPathChunkSize
	}
	r.Path(p, &paint)
}

func lineDashesForPaint(dashes []float64, lineWidth float64, units DashUnits) []float64 {
	if len(dashes) == 0 {
		return nil
	}
	out := append([]float64(nil), dashes...)
	if units == DashUnitsMatplotlib {
		for i := range out {
			out[i] *= lineWidth
		}
	}
	return out
}

// Z returns the z-order for sorting.
func (l *Line2D) Z() float64 {
	return zOrDefault(l.z, defaultLineZ)
}

// Bounds returns the bounding box of all points in data space.
func (l *Line2D) Bounds(*DrawContext) geom.Rect {
	if len(l.XY) == 0 {
		return geom.Rect{}
	}
	r := geom.Rect{Min: l.XY[0], Max: l.XY[0]}
	for _, pt := range l.XY[1:] {
		if pt.X < r.Min.X {
			r.Min.X = pt.X
		}
		if pt.Y < r.Min.Y {
			r.Min.Y = pt.Y
		}
		if pt.X > r.Max.X {
			r.Max.X = pt.X
		}
		if pt.Y > r.Max.Y {
			r.Max.Y = pt.Y
		}
	}
	return r
}

func (l *Line2D) pathPoints() []geom.Pt {
	if len(l.XY) < 2 {
		return l.XY
	}

	switch l.DrawStyle {
	case LineDrawStyleStepsPre:
		out := make([]geom.Pt, 0, 2*len(l.XY)-1)
		out = append(out, l.XY[0])
		for i := 1; i < len(l.XY); i++ {
			out = append(out,
				geom.Pt{X: l.XY[i-1].X, Y: l.XY[i].Y},
				l.XY[i],
			)
		}
		return out
	case LineDrawStyleStepsMid:
		out := make([]geom.Pt, 0, 3*len(l.XY)-2)
		out = append(out, l.XY[0])
		for i := 1; i < len(l.XY); i++ {
			midX := (l.XY[i-1].X + l.XY[i].X) / 2
			out = append(out,
				geom.Pt{X: midX, Y: l.XY[i-1].Y},
				geom.Pt{X: midX, Y: l.XY[i].Y},
				l.XY[i],
			)
		}
		return out
	case LineDrawStyleStepsPost:
		out := make([]geom.Pt, 0, 2*len(l.XY)-1)
		out = append(out, l.XY[0])
		for i := 1; i < len(l.XY); i++ {
			out = append(out,
				geom.Pt{X: l.XY[i].X, Y: l.XY[i-1].Y},
				l.XY[i],
			)
		}
		return out
	default:
		return l.XY
	}
}

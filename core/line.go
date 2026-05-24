package core

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
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

// Line2D is a polyline artist with optional Matplotlib-style data markers.
type Line2D struct {
	ArtistRasterization
	XY              []geom.Pt    // data space points
	W               float64      // stroke width (px for now)
	Col             render.Color // stroke color
	Dashes          []float64    // dash pattern (on/off pairs)
	DashUnits       DashUnits    // unit system for Dashes
	PathEffects     []render.PathEffect
	DrawStyle       LineDrawStyle // optional step-style connection mode
	Marker          MarkerType    // optional data marker
	MarkerSet       bool          // true when Marker should be drawn
	MarkerStyle     MarkerStyle   // optional rich marker style
	MarkerPath      geom.Path     // optional custom marker path in normalized marker space
	MarkerSize      float64       // marker size in points, 0 uses Matplotlib's 6 pt default
	MarkerFaceColor render.Color  // marker fill, 0 alpha falls back to line color
	MarkerEdgeColor render.Color  // marker edge, 0 alpha falls back to line color
	MarkerEdgeWidth float64       // marker edge width in pixels, 0 uses 1 px
	MarkEvery       int           // optional every-N marker subset; <=1 draws every point
	Label           string        // series label for legend
	z               float64       // z-order
	pickRadius      float64       // pick tolerance in pixels (0 = default)
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
	tr := artistTransformFor(ctx, l, Coords(CoordData))
	for i, v := range points {
		q := v
		if tr != nil {
			q = tr.Apply(v)
		}
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
		Stroke:      l.ApplyArtistAlpha(l.Col),
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
	l.drawMarkers(r, ctx)
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
	if len(l.XY) == 0 || !artistUsesDataCoords(l, Coords(CoordData)) {
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

func (l *Line2D) drawMarkers(r render.Renderer, ctx *DrawContext) {
	if l == nil || r == nil || ctx == nil || !l.hasMarkers() {
		return
	}
	points := l.markerPoints()
	if len(points) == 0 {
		return
	}
	markerPath := l.markerPrototypePath(r, ctx)
	if len(markerPath.C) == 0 {
		return
	}
	markerSize := l.resolvedMarkerSize(ctx)
	if markerSize <= 0 {
		return
	}

	markers := &PathCollection{
		Collection: Collection{
			ArtistRasterization: l.ArtistRasterization,
			Coords:              Coords(CoordData),
			Alpha:               1,
			PathEffects:         cloneRenderPathEffects(l.PathEffects),
		},
		Path:          markerPath,
		Offsets:       points,
		Size:          markerSize,
		PathInDisplay: true,
		FaceColor:     l.resolvedMarkerFaceColor(),
		EdgeColor:     l.resolvedMarkerEdgeColor(),
		EdgeWidth:     l.resolvedMarkerEdgeWidth(),
		LineJoin:      (&Scatter2D{Marker: l.Marker, MarkerStyle: l.MarkerStyle, MarkerPath: l.MarkerPath}).markerLineJoin(),
		LineJoinSet:   true,
		LineCap:       render.CapButt,
		LineCapSet:    true,
		LineOnly:      markerLineOnly(l.resolvedMarkerStyle()),
	}
	markers.Draw(r, ctx)
}

func (l *Line2D) hasMarkers() bool {
	if l == nil {
		return false
	}
	return l.MarkerSet || len(l.MarkerPath.C) > 0 || l.MarkerStyle.Type != 0 || l.MarkerStyle.FillStyle != 0 ||
		l.MarkerStyle.Tuple != nil || l.MarkerStyle.MathText != "" || len(l.MarkerStyle.Path.C) > 0
}

func (l *Line2D) markerPoints() []geom.Pt {
	if l == nil || len(l.XY) == 0 {
		return nil
	}
	if l.MarkEvery <= 1 {
		return append([]geom.Pt(nil), l.XY...)
	}
	out := make([]geom.Pt, 0, (len(l.XY)+l.MarkEvery-1)/l.MarkEvery)
	for i, pt := range l.XY {
		if i%l.MarkEvery == 0 {
			out = append(out, pt)
		}
	}
	return out
}

func (l *Line2D) resolvedMarkerStyle() MarkerStyle {
	if l == nil {
		return MarkerStyle{}
	}
	if l.MarkerStyle.Tuple != nil || l.MarkerStyle.MathText != "" || len(l.MarkerStyle.Path.C) > 0 || l.MarkerStyle.Type != 0 || l.MarkerStyle.FillStyle != 0 {
		style := l.MarkerStyle
		if style.FillStyle == 0 {
			style.FillStyle = MarkerFillFull
		}
		return style
	}
	return MarkerStyle{Type: l.Marker, FillStyle: MarkerFillFull}
}

func (l *Line2D) markerPrototypePath(r render.Renderer, ctx *DrawContext) geom.Path {
	if l == nil {
		return geom.Path{}
	}
	scatter := Scatter2D{
		Marker:      l.Marker,
		MarkerStyle: l.resolvedMarkerStyle(),
		MarkerPath:  l.MarkerPath,
	}
	return scatter.markerPrototypePathForContext(r, ctx)
}

func (l *Line2D) resolvedMarkerSize(ctx *DrawContext) float64 {
	size := 6.0
	if l != nil && l.MarkerSize > 0 {
		size = l.MarkerSize
	}
	rc := style.Default
	if ctx != nil {
		rc = ctx.RC
	}
	return 0.5 * pointsToPixels(rc, size)
}

func (l *Line2D) resolvedMarkerFaceColor() render.Color {
	if l == nil {
		return render.Color{}
	}
	color := l.MarkerFaceColor
	if color.A <= 0 {
		color = l.Col
	}
	return l.ApplyArtistAlpha(color)
}

func (l *Line2D) resolvedMarkerEdgeColor() render.Color {
	if l == nil {
		return render.Color{}
	}
	color := l.MarkerEdgeColor
	if color.A <= 0 {
		color = l.Col
	}
	return l.ApplyArtistAlpha(color)
}

func (l *Line2D) resolvedMarkerEdgeWidth() float64 {
	if l == nil || l.MarkerEdgeWidth <= 0 {
		return 1
	}
	return l.MarkerEdgeWidth
}

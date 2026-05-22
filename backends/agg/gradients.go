package agg

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// SupportsGradientFill reports that the AGG backend renders Paint.FillGradient
// natively via Agg2D's linear and radial gradient span generators.
//
// The current AGG bridge supports two-stop linear and radial gradients. For
// gradients with more than two stops, the first and last stops drive the
// linear gradient endpoints; the radial path additionally uses the middle stop
// when exactly three stops are supplied via the multi-stop variant.
func (r *Renderer) SupportsGradientFill() bool { return true }

// SupportsPatternFill reports that the AGG backend consumes Paint.FillPattern
// by replaying the pattern tile through AGG path drawing under the destination
// path clip.
func (r *Renderer) SupportsPatternFill() bool { return true }

// applyGradientFill configures the AGG fill state for the gradient described
// by paint.FillGradient and returns true. The caller is responsible for
// resetting the fill color to a solid value afterwards if it issues further
// fill operations with a different source.
func (r *Renderer) applyGradientFill(paint *render.Paint) bool {
	g := &paint.FillGradient
	if g.Kind == render.GradientNone || len(g.Stops) == 0 {
		return false
	}

	stops := g.Stops
	first := stops[0].Color
	last := stops[len(stops)-1].Color
	first = colorWithForcedAlpha(first, paint)
	last = colorWithForcedAlpha(last, paint)

	switch g.Kind {
	case render.LinearGradient:
		r.ctx.SetFillLinearGradient(
			g.Start.X, g.Start.Y,
			g.End.X, g.End.Y,
			renderColorToAGG(first), renderColorToAGG(last),
			1.0,
		)
		return true
	case render.RadialGradient:
		if len(stops) >= 3 {
			mid := colorWithForcedAlpha(stops[len(stops)/2].Color, paint)
			r.ctx.SetFillRadialGradientMultiStop(
				g.Center.X, g.Center.Y, g.Radius,
				renderColorToAGG(first),
				renderColorToAGG(mid),
				renderColorToAGG(last),
			)
			return true
		}
		r.ctx.SetFillRadialGradient(
			g.Center.X, g.Center.Y, g.Radius,
			renderColorToAGG(first), renderColorToAGG(last),
			1.0,
		)
		return true
	}
	return false
}

func (r *Renderer) drawPatternFill(clipPath geom.Path, paint *render.Paint) bool {
	if paint == nil {
		return false
	}
	pattern := paint.FillPattern
	if pattern.ID == "" && len(pattern.Path.V) == 0 {
		return false
	}

	cell := pattern.Cell
	cellW := cell.W()
	cellH := cell.H()
	if cellW <= 0 || cellH <= 0 {
		if bounds, ok := pathBounds(pattern.Path); ok {
			cell = bounds
			cellW = cell.W()
			cellH = cell.H()
		}
	}
	if cellW <= 0 {
		cellW = 16
	}
	if cellH <= 0 {
		cellH = 16
	}

	bounds, ok := pathBounds(clipPath)
	if !ok {
		return false
	}
	oldPaths := r.clipPaths
	r.clipPaths = append(clonePaths(oldPaths), clonePath(clipPath))
	defer func() {
		r.clipPaths = oldPaths
	}()

	startX := cell.Min.X + math.Floor((bounds.Min.X-cell.Min.X)/cellW-1)*cellW
	endX := bounds.Max.X + cellW
	startY := cell.Min.Y + math.Floor((bounds.Min.Y-cell.Min.Y)/cellH-1)*cellH
	endY := bounds.Max.Y + cellH

	bg := colorWithForcedAlpha(pattern.Background, paint)
	fg := colorWithForcedAlpha(pattern.Foreground, paint)
	for y := startY; y <= endY; y += cellH {
		for x := startX; x <= endX; x += cellW {
			if bg.A > 0 {
				r.Path(patternCellRect(x, y, cellW, cellH), &render.Paint{
					Fill:      bg,
					Antialias: paint.Antialias,
					Snap:      paint.Snap,
				})
			}
			if len(pattern.Path.C) == 0 || fg.A <= 0 {
				continue
			}
			tilePath := patternTilePath(pattern, geom.Pt{X: x, Y: y})
			tilePaint := render.Paint{
				Antialias: paint.Antialias,
				Snap:      paint.Snap,
			}
			if pattern.LineWidth > 0 {
				tilePaint.Stroke = fg
				tilePaint.LineWidth = pattern.LineWidth
				tilePaint.LineJoin = paint.LineJoin
				tilePaint.LineCap = paint.LineCap
			} else {
				tilePaint.Fill = fg
			}
			r.Path(tilePath, &tilePaint)
		}
	}
	return true
}

func patternCellRect(x, y, w, h float64) geom.Path {
	var path geom.Path
	path.MoveTo(geom.Pt{X: x, Y: y})
	path.LineTo(geom.Pt{X: x + w, Y: y})
	path.LineTo(geom.Pt{X: x + w, Y: y + h})
	path.LineTo(geom.Pt{X: x, Y: y + h})
	path.Close()
	return path
}

func patternTilePath(pattern render.PatternFill, offset geom.Pt) geom.Path {
	out := clonePath(pattern.Path)
	transform := geom.Affine{A: 1, D: 1, E: offset.X, F: offset.Y}
	if pattern.HasTransform {
		transform = transform.Mul(pattern.Transform)
	}
	for i, pt := range out.V {
		out.V[i] = transform.Apply(pt)
	}
	return out
}

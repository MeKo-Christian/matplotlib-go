package pgf

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// SupportsGradientFill reports that the PGF backend renders Paint.FillGradient
// natively via \pgfdeclare{horizontal,radial}shading + \pgfuseshading inside a
// path clip. Matplotlib's own PGF backend rasterizes gradients (mixed mode);
// this port keeps them fully vector. Implements render.GradientFiller.
func (r *Renderer) SupportsGradientFill() bool { return true }

// SupportsPatternFill reports that the PGF backend renders Paint.FillPattern
// natively by tiling the pattern cell inside a clipped pgfscope, mirroring the
// existing native hatch path. Implements render.PatternFiller.
func (r *Renderer) SupportsPatternFill() bool { return true }

func hasGradientFill(paint *render.Paint) bool {
	return paint != nil &&
		paint.FillGradient.Kind != render.GradientNone &&
		len(paint.FillGradient.Stops) > 0
}

func hasPatternFill(paint *render.Paint) bool {
	return paint != nil &&
		(paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0)
}

// writeGradientFill clips to the path and paints the gradient with a declared
// TikZ shading placed (and, for linear gradients, rotated) over the path's
// bounding box. PGF/TeX is natively y-up display space, so gradient geometry —
// already in y-up display coordinates — is emitted without a device flip.
func (r *Renderer) writeGradientFill(w *strings.Builder, path geom.Path, paint *render.Paint) bool {
	if !hasGradientFill(paint) {
		return false
	}
	g := paint.FillGradient
	g.Stops = normalizeGradientStops(g.Stops)

	switch g.Kind {
	case render.LinearGradient:
		return r.writeLinearGradientFill(w, path, &g)
	case render.RadialGradient:
		bounds, ok := path.Bounds()
		if !ok {
			return false
		}
		return r.writeRadialGradientFill(w, path, &g, bounds)
	default:
		return false
	}
}

// writeLinearGradientFill fills the path with a linear gradient using pgf's
// \pgfshadepath, which rescales a 0–100bp shading to cover the current path's
// bounding box, rotates it by the gradient angle, clips to the path and draws
// it directly (no TeX box-model placement, so it never overflows or clips at
// page edges). The gradient is fit to the path's extent along the Start→End
// direction; matplotlib has no PGF gradient reference, so this bbox-fit
// interpretation is the natural one.
func (r *Renderer) writeLinearGradientFill(w *strings.Builder, path geom.Path, g *render.GradientFill) bool {
	start := transformedGradientPoint(g.Start, g)
	end := transformedGradientPoint(g.End, g)
	dx := end.X - start.X
	dy := end.Y - start.Y
	if dx == 0 && dy == 0 {
		return r.writeSolidGradientFill(w, path, sampleGradientColor(g.Stops, 0.5))
	}
	angle := math.Atan2(dy, dx) * 180 / math.Pi

	// \pgfshadepath maps the path's bounding box onto the MIDDLE half of the
	// declared 0..100bp shading: the path's lower-left corner lands at 25bp and
	// its upper-right at 75bp (see the pgf manual for \pgfshadepath). So the
	// gradient body must occupy 25..75bp, with the terminal colors padding
	// 0..25bp and 75..100bp (the Extend behavior outside the gradient).
	var spec strings.Builder
	emit := func(pos float64, c render.Color) {
		if spec.Len() > 0 {
			spec.WriteString("; ")
		}
		fmt.Fprintf(&spec, "color(%sbp)=(%s)", shortFloat(pos), r.colorName(c))
	}
	emit(0, sampleGradientColor(g.Stops, 0))
	for _, s := range g.Stops {
		emit(25+clamp01(s.Offset)*50, s.Color)
	}
	emit(100, sampleGradientColor(g.Stops, 1))

	name := r.nextShadingName()
	fmt.Fprintf(w, "\\pgfdeclarehorizontalshading{%s}{100bp}{%s}\n", name, spec.String())
	if !writePathOps(w, path) {
		return false
	}
	fmt.Fprintf(w, "\\pgfshadepath{%s}{%s}\n", name, shortFloat(angle))
	// \pgfshadepath leaves the path current; discard it so a following stroke
	// pass starts from a clean path.
	w.WriteString("\\pgfusepath{}\n")
	return true
}

func (r *Renderer) writeRadialGradientFill(w *strings.Builder, path geom.Path, g *render.GradientFill, bounds geom.Rect) bool {
	center := transformedGradientPoint(g.Center, g)
	focal := center
	if g.Focal != (geom.Pt{}) {
		focal = transformedGradientPoint(g.Focal, g)
	}
	radius := transformedGradientRadius(g.Radius, g)
	if radius <= 0 {
		return r.writeSolidGradientFill(w, path, sampleGradientColor(g.Stops, 1))
	}

	// The declared radii must reach the farthest bbox corner so the shading box
	// covers the whole clipped region; beyond the last real stop the terminal
	// color is held constant (Extend behavior).
	maxDist := 0.0
	for _, c := range rectCorners(bounds) {
		maxDist = math.Max(maxDist, math.Hypot(c.X-center.X, c.Y-center.Y))
	}
	if maxDist <= 0 {
		return r.writeSolidGradientFill(w, path, sampleGradientColor(g.Stops, 1))
	}

	var spec strings.Builder
	for i, s := range g.Stops {
		if i > 0 {
			spec.WriteString("; ")
		}
		fmt.Fprintf(&spec, "color(%spt)=(%s)", shortFloat(s.Offset*radius), r.colorName(s.Color))
	}
	if maxDist > radius {
		fmt.Fprintf(&spec, "; color(%spt)=(%s)", shortFloat(maxDist), r.colorName(g.Stops[len(g.Stops)-1].Color))
	}

	name := r.nextShadingName()
	fmt.Fprintf(w, "\\pgfdeclareradialshading{%s}{\\pgfpoint{%spt}{%spt}}{%s}\n",
		name, shortFloat(focal.X-center.X), shortFloat(focal.Y-center.Y), spec.String())

	w.WriteString("\\pgfscope\n")
	if !writePathOps(w, path) {
		w.WriteString("\\endpgfscope\n")
		return false
	}
	w.WriteString("\\pgfusepath{clip}\n")
	fmt.Fprintf(w, "\\pgftext[at=\\pgfpoint{%spt}{%spt}]{\\pgfuseshading{%s}}\n",
		shortFloat(center.X), shortFloat(center.Y), name)
	w.WriteString("\\endpgfscope\n")
	return true
}

// writeSolidGradientFill is the degenerate fallback (zero-length axis or
// collapsed bounds): fill the path with a single representative stop color.
func (r *Renderer) writeSolidGradientFill(w *strings.Builder, path geom.Path, c render.Color) bool {
	if !writePathOps(w, path) {
		return false
	}
	writeFillOpacity(w, c.A)
	writeStrokeOpacity(w, 1)
	writeFillColor(w, r.colorName(c))
	w.WriteString("\\pgfusepath{fill}\n")
	return true
}

func (r *Renderer) writePatternFill(w *strings.Builder, path geom.Path, paint *render.Paint) bool {
	if !hasPatternFill(paint) {
		return false
	}
	pattern := paint.FillPattern

	cell := pattern.Cell
	cellW := cell.W()
	cellH := cell.H()
	if cellW <= 0 || cellH <= 0 {
		if b, ok := pattern.Path.Bounds(); ok {
			cell = b
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

	bounds, ok := path.Bounds()
	if !ok {
		return false
	}

	w.WriteString("\\pgfscope\n")
	if !writePathOps(w, path) {
		w.WriteString("\\endpgfscope\n")
		return false
	}
	w.WriteString("\\pgfusepath{clip}\n")

	startX := cell.Min.X + math.Floor((bounds.Min.X-cell.Min.X)/cellW-1)*cellW
	endX := bounds.Max.X + cellW
	startY := cell.Min.Y + math.Floor((bounds.Min.Y-cell.Min.Y)/cellH-1)*cellH
	endY := bounds.Max.Y + cellH

	bg := pattern.Background
	fg := pattern.Foreground
	for y := startY; y <= endY; y += cellH {
		for x := startX; x <= endX; x += cellW {
			if bg.A > 0 {
				if writePathOps(w, patternCellRect(x, y, cellW, cellH)) {
					writeFillOpacity(w, bg.A)
					writeStrokeOpacity(w, 1)
					writeFillColor(w, r.colorName(bg))
					w.WriteString("\\pgfusepath{fill}\n")
				}
			}
			if len(pattern.Path.C) == 0 || fg.A <= 0 {
				continue
			}
			tile := patternTilePath(&pattern, geom.Pt{X: x, Y: y})
			if !writePathOps(w, tile) {
				continue
			}
			if pattern.LineWidth > 0 {
				writeFillOpacity(w, 1)
				writeStrokeOpacity(w, fg.A)
				writeStrokeColor(w, r.colorName(fg))
				fmt.Fprintf(w, "\\pgfsetlinewidth{%spt}\n", shortFloat(pattern.LineWidth))
				w.WriteString("\\pgfusepath{stroke}\n")
			} else {
				writeFillOpacity(w, fg.A)
				writeStrokeOpacity(w, 1)
				writeFillColor(w, r.colorName(fg))
				w.WriteString("\\pgfusepath{fill}\n")
			}
		}
	}
	w.WriteString("\\endpgfscope\n")
	return true
}

func (r *Renderer) nextShadingName() string {
	r.shadingCounter++
	return fmt.Sprintf("mplgpgfshading%d", r.shadingCounter)
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

func patternTilePath(pattern *render.PatternFill, offset geom.Pt) geom.Path {
	transform := geom.Affine{A: 1, D: 1, E: offset.X, F: offset.Y}
	if pattern.HasTransform {
		transform = transform.Mul(pattern.Transform)
	}
	return pattern.Path.Transformed(transform)
}

func rectCorners(r geom.Rect) [4]geom.Pt {
	return [4]geom.Pt{
		{X: r.Min.X, Y: r.Min.Y},
		{X: r.Max.X, Y: r.Min.Y},
		{X: r.Max.X, Y: r.Max.Y},
		{X: r.Min.X, Y: r.Max.Y},
	}
}

func transformedGradientPoint(p geom.Pt, gradient *render.GradientFill) geom.Pt {
	if !gradient.HasTransform {
		return p
	}
	return gradient.Transform.Apply(p)
}

func transformedGradientRadius(radius float64, gradient *render.GradientFill) float64 {
	if radius == 0 || !gradient.HasTransform {
		return radius
	}
	xScale := math.Hypot(gradient.Transform.A, gradient.Transform.B)
	yScale := math.Hypot(gradient.Transform.C, gradient.Transform.D)
	return radius * math.Max(xScale, yScale)
}

// sampleGradientColor returns the gradient color at offset, clamping to the
// terminal stop colors for offsets outside the [first,last] stop range.
func sampleGradientColor(stops []render.GradientStop, offset float64) render.Color {
	if len(stops) == 0 {
		return render.Color{A: 1}
	}
	if offset <= stops[0].Offset {
		return stops[0].Color
	}
	last := stops[len(stops)-1]
	if offset >= last.Offset {
		return last.Color
	}
	for i := 1; i < len(stops); i++ {
		if offset > stops[i].Offset {
			continue
		}
		lo := stops[i-1]
		hi := stops[i]
		span := hi.Offset - lo.Offset
		if span <= 0 {
			return hi.Color
		}
		t := (offset - lo.Offset) / span
		return lerpColor(lo.Color, hi.Color, t)
	}
	return last.Color
}

func lerpColor(a, b render.Color, t float64) render.Color {
	return render.Color{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
		A: a.A + (b.A-a.A)*t,
	}
}

func normalizeGradientStops(in []render.GradientStop) []render.GradientStop {
	if len(in) == 0 {
		return nil
	}
	out := append([]render.GradientStop(nil), in...)
	for i := range out {
		out[i].Offset = clamp01(out[i].Offset)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Offset < out[j].Offset
	})
	if len(out) == 1 {
		return out
	}
	for i := 1; i < len(out); i++ {
		if out[i].Offset <= out[i-1].Offset {
			out[i].Offset = math.Nextafter(out[i-1].Offset, 1)
		}
	}
	return out
}

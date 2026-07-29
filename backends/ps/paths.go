package ps

import (
	"fmt"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/sketch"
	"github.com/cwbudde/matplotlib-go/render"
)

// Path draws a path using the provided paint.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if rr := r.activeRaster(); rr != nil {
		rr.Path(p, paint)
		return
	}
	if !r.began || paint == nil {
		return
	}
	if render.DrawPathWithEffects(r, p, paint, r.Path) {
		return
	}
	// Sketch/xkcd perturbation in y-up display space (PostScript is natively y-up).
	if eff := render.EffectiveSketch(paint.Sketch, r.defaultSketch); render.SketchActive(eff) {
		p = sketch.Apply(p, eff.Scale, eff.Length, eff.Randomness)
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasGradient := !hasHatch && hasGradientFill(paint)
	hasPattern := !hasHatch && !hasGradient && hasPatternFill(paint)
	hasFill := paint.Fill.A > 0 || hasHatch || hasGradient || hasPattern
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasStroke {
		return
	}

	// Gradient and pattern fills build (and clip to) their own path, so they are
	// handled before the shared writePathOps pre-build below. An optional stroke
	// outline is re-emitted afterwards, mirroring the AGG/PDF fill+stroke order.
	if hasGradient || hasPattern {
		if hasGradient {
			r.writeGradientFill(&r.content, p, paint)
		} else {
			r.writePatternFill(&r.content, p, paint)
		}
		if hasStroke {
			if !writePathOps(&r.content, p) {
				return
			}
			writeStrokeColor(&r.content, paint.Stroke)
			writeLineState(&r.content, paint)
			r.content.WriteString("stroke\n")
		}
		return
	}

	if !writePathOps(&r.content, p) {
		return
	}

	switch {
	case hasHatch:
		if paint.Fill.A > 0 {
			writeFillColor(&r.content, paint.Fill)
			r.content.WriteString("gsave fill grestore\n")
		}
		r.writeHatchFill(p, paint)
		if hasStroke {
			if !writePathOps(&r.content, p) {
				return
			}
			writeStrokeColor(&r.content, paint.Stroke)
			writeLineState(&r.content, paint)
			r.content.WriteString("stroke\n")
		}
	case hasFill && hasStroke:
		writeFillColor(&r.content, paint.Fill)
		r.content.WriteString("gsave fill grestore\n")
		writeStrokeColor(&r.content, paint.Stroke)
		writeLineState(&r.content, paint)
		r.content.WriteString("stroke\n")
	case hasFill:
		writeFillColor(&r.content, paint.Fill)
		r.content.WriteString("fill\n")
	case hasStroke:
		writeStrokeColor(&r.content, paint.Stroke)
		writeLineState(&r.content, paint)
		r.content.WriteString("stroke\n")
	}
}

// DrawPathWithEffects applies renderer-neutral path effect passes.
func (r *Renderer) DrawPathWithEffects(p geom.Path, paint *render.Paint) bool {
	if rr := r.activeRaster(); rr != nil {
		if effects, ok := rr.(render.PathEffectDrawer); ok {
			return effects.DrawPathWithEffects(p, paint)
		}
		rr.Path(p, paint)
		return true
	}
	return render.DrawPathWithEffects(r, p, paint, r.Path)
}

func writePathOps(w *strings.Builder, p geom.Path) bool {
	if !p.Validate() || len(p.C) == 0 {
		return false
	}
	w.WriteString("newpath\n")
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			vi++
			fmt.Fprintf(w, "%s %s moveto\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.LineTo:
			pt := p.V[vi]
			vi++
			fmt.Fprintf(w, "%s %s lineto\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.QuadTo:
			if vi == 0 {
				vi += 2
				continue
			}
			prev := lastEndpoint(p, vi)
			ctrl := p.V[vi]
			end := p.V[vi+1]
			vi += 2
			c1 := geom.Pt{
				X: prev.X + (2.0/3.0)*(ctrl.X-prev.X),
				Y: prev.Y + (2.0/3.0)*(ctrl.Y-prev.Y),
			}
			c2 := geom.Pt{
				X: end.X + (2.0/3.0)*(ctrl.X-end.X),
				Y: end.Y + (2.0/3.0)*(ctrl.Y-end.Y),
			}
			fmt.Fprintf(
				w, "%s %s %s %s %s %s curveto\n",
				shortFloat(c1.X), shortFloat(c1.Y),
				shortFloat(c2.X), shortFloat(c2.Y),
				shortFloat(end.X), shortFloat(end.Y),
			)
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			end := p.V[vi+2]
			vi += 3
			fmt.Fprintf(
				w, "%s %s %s %s %s %s curveto\n",
				shortFloat(c1.X), shortFloat(c1.Y),
				shortFloat(c2.X), shortFloat(c2.Y),
				shortFloat(end.X), shortFloat(end.Y),
			)
		case geom.ClosePath:
			w.WriteString("closepath\n")
		}
	}
	return true
}

func lastEndpoint(p geom.Path, vi int) geom.Pt {
	consumed := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo, geom.LineTo:
			consumed++
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.QuadTo:
			consumed += 2
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.CubicTo:
			consumed += 3
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.ClosePath:
		}
	}
	return geom.Pt{}
}

func writePathPaintOps(w *strings.Builder, path geom.Path, paint *render.Paint) bool {
	if paint == nil {
		return false
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasFill := paint.Fill.A > 0 || hasHatch
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	switch {
	case hasHatch:
		if paint.Fill.A > 0 {
			writeFillColor(w, paint.Fill)
			w.WriteString("gsave fill grestore\n")
		}
		writeStrokeColor(w, paint.HatchColor)
		lineWidth := paint.HatchLineWidth
		if lineWidth <= 0 {
			lineWidth = 1
		}
		fmt.Fprintf(w, "%s setlinewidth\n", shortFloat(lineWidth))
		w.WriteString("gsave clip newpath\n")
		for _, line := range hatchPatternLines(paint.Hatch, paint.HatchSpacing) {
			fmt.Fprintf(
				w, "newpath %s %s moveto %s %s lineto\nstroke\n",
				shortFloat(line[0].X), shortFloat(line[0].Y),
				shortFloat(line[1].X), shortFloat(line[1].Y),
			)
		}
		writeHatchShapeOps(w, paint.Hatch, paint.HatchSpacing)
		w.WriteString("grestore\n")
		if hasStroke {
			if !writePathOps(w, path) {
				return false
			}
			writeStrokeColor(w, paint.Stroke)
			writeLineState(w, paint)
			w.WriteString("stroke\n")
		}
	case hasFill && hasStroke:
		writeFillColor(w, paint.Fill)
		w.WriteString("gsave fill grestore\n")
		writeStrokeColor(w, paint.Stroke)
		writeLineState(w, paint)
		w.WriteString("stroke\n")
	case hasFill:
		writeFillColor(w, paint.Fill)
		w.WriteString("fill\n")
	case hasStroke:
		writeStrokeColor(w, paint.Stroke)
		writeLineState(w, paint)
		w.WriteString("stroke\n")
	default:
		return false
	}
	return true
}

func writeLineState(w *strings.Builder, paint *render.Paint) {
	lineWidth := paint.LineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	fmt.Fprintf(w, "%s setlinewidth\n", shortFloat(lineWidth))
	fmt.Fprintf(w, "%d setlinejoin\n", lineJoin(paint.LineJoin))
	fmt.Fprintf(w, "%d setlinecap\n", lineCap(paint.LineCap))
	if paint.MiterLimit > 0 {
		fmt.Fprintf(w, "%s setmiterlimit\n", shortFloat(paint.MiterLimit))
	}
	if len(paint.Dashes) > 0 {
		w.WriteByte('[')
		for i, d := range paint.Dashes {
			if i > 0 {
				w.WriteByte(' ')
			}
			w.WriteString(shortFloat(d))
		}
		fmt.Fprintf(w, "] %s setdash\n", shortFloat(paint.DashOffset))
	} else {
		w.WriteString("[] 0 setdash\n")
	}
}

func paintVisible(paint *render.Paint) bool {
	if paint == nil {
		return false
	}
	return paint.Fill.A > 0 ||
		(paint.Hatch != "" && paint.HatchColor.A > 0) ||
		(paint.Stroke.A > 0 && paint.LineWidth > 0)
}

package pgf

import (
	"fmt"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Path draws a vector path.
func (r *Renderer) Path(path geom.Path, paint *render.Paint) {
	if rr := r.activeRaster(); rr != nil {
		rr.Path(path, paint)
		return
	}
	if !r.began || paint == nil {
		return
	}
	if render.DrawPathWithEffects(r, path, paint, r.Path) {
		return
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasFill := paint.Fill.A > 0 || hasHatch
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasStroke {
		return
	}

	r.writePathPaintOps(&r.content, path, paint)
}

// DrawPathWithEffects applies renderer-neutral path effect passes.
func (r *Renderer) DrawPathWithEffects(path geom.Path, paint *render.Paint) bool {
	if rr := r.activeRaster(); rr != nil {
		if effects, ok := rr.(render.PathEffectDrawer); ok {
			return effects.DrawPathWithEffects(path, paint)
		}
		rr.Path(path, paint)
		return true
	}
	return render.DrawPathWithEffects(r, path, paint, r.Path)
}

func writePathOps(w *strings.Builder, path geom.Path) bool {
	if !path.Validate() || len(path.C) == 0 {
		return false
	}
	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			pt := path.V[vi]
			vi++
			fmt.Fprintf(w, "\\pgfpathmoveto{\\pgfpoint{%spt}{%spt}}\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.LineTo:
			pt := path.V[vi]
			vi++
			fmt.Fprintf(w, "\\pgfpathlineto{\\pgfpoint{%spt}{%spt}}\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.QuadTo:
			if vi == 0 {
				vi += 2
				continue
			}
			prev := lastEndpoint(path, vi)
			ctrl := path.V[vi]
			end := path.V[vi+1]
			vi += 2
			c1 := geom.Pt{X: prev.X + (2.0/3.0)*(ctrl.X-prev.X), Y: prev.Y + (2.0/3.0)*(ctrl.Y-prev.Y)}
			c2 := geom.Pt{X: end.X + (2.0/3.0)*(ctrl.X-end.X), Y: end.Y + (2.0/3.0)*(ctrl.Y-end.Y)}
			writeCurve(w, c1, c2, end)
		case geom.CubicTo:
			c1 := path.V[vi]
			c2 := path.V[vi+1]
			end := path.V[vi+2]
			vi += 3
			writeCurve(w, c1, c2, end)
		case geom.ClosePath:
			w.WriteString("\\pgfpathclose\n")
		}
	}
	return true
}

func (r *Renderer) writePathPaintOps(w *strings.Builder, path geom.Path, paint *render.Paint) bool {
	if paint == nil {
		return false
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasFill := paint.Fill.A > 0 || hasHatch
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	switch {
	case hasHatch:
		if paint.Fill.A > 0 {
			if !writePathOps(w, path) {
				return false
			}
			writeFillOpacity(w, paint.Fill.A)
			writeStrokeOpacity(w, 1)
			writeFillColor(w, r.colorName(paint.Fill))
			w.WriteString("\\pgfusepath{fill}\n")
		}
		r.writeHatchFillTo(w, path, paint)
		if hasStroke {
			if !writePathOps(w, path) {
				return false
			}
			writeFillOpacity(w, 1)
			writeStrokeOpacity(w, paint.Stroke.A)
			writeStrokeColor(w, r.colorName(paint.Stroke))
			writeLineState(w, paint)
			w.WriteString("\\pgfusepath{stroke}\n")
		}
	case hasFill && hasStroke:
		if !writePathOps(w, path) {
			return false
		}
		writeFillOpacity(w, paint.Fill.A)
		writeStrokeOpacity(w, paint.Stroke.A)
		writeFillColor(w, r.colorName(paint.Fill))
		writeStrokeColor(w, r.colorName(paint.Stroke))
		writeLineState(w, paint)
		w.WriteString("\\pgfusepath{fill,stroke}\n")
	case hasFill:
		if !writePathOps(w, path) {
			return false
		}
		writeFillOpacity(w, paint.Fill.A)
		writeStrokeOpacity(w, 1)
		writeFillColor(w, r.colorName(paint.Fill))
		w.WriteString("\\pgfusepath{fill}\n")
	case hasStroke:
		if !writePathOps(w, path) {
			return false
		}
		writeFillOpacity(w, 1)
		writeStrokeOpacity(w, paint.Stroke.A)
		writeStrokeColor(w, r.colorName(paint.Stroke))
		writeLineState(w, paint)
		w.WriteString("\\pgfusepath{stroke}\n")
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
	fmt.Fprintf(w, "\\pgfsetlinewidth{%spt}\n", shortFloat(lineWidth))
	if len(paint.Dashes) > 0 {
		w.WriteString("\\pgfsetdash{")
		for i, d := range paint.Dashes {
			if i > 0 {
				w.WriteString(",")
			}
			fmt.Fprintf(w, "{%spt}", shortFloat(d))
		}
		w.WriteString("}{0pt}\n")
	} else {
		w.WriteString("\\pgfsetdash{}{0pt}\n")
	}
}

func writeCurve(w *strings.Builder, c1, c2, end geom.Pt) {
	fmt.Fprintf(w, "\\pgfpathcurveto{\\pgfpoint{%spt}{%spt}}{\\pgfpoint{%spt}{%spt}}{\\pgfpoint{%spt}{%spt}}\n",
		shortFloat(c1.X), shortFloat(c1.Y),
		shortFloat(c2.X), shortFloat(c2.Y),
		shortFloat(end.X), shortFloat(end.Y))
}

func lastEndpoint(path geom.Path, vi int) geom.Pt {
	consumed := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo, geom.LineTo:
			consumed++
			if consumed == vi {
				return path.V[consumed-1]
			}
		case geom.QuadTo:
			consumed += 2
			if consumed == vi {
				return path.V[consumed-1]
			}
		case geom.CubicTo:
			consumed += 3
			if consumed == vi {
				return path.V[consumed-1]
			}
		case geom.ClosePath:
		}
	}
	return geom.Pt{}
}

func paintVisible(paint *render.Paint) bool {
	if paint == nil {
		return false
	}
	return paint.Fill.A > 0 ||
		(paint.Hatch != "" && paint.HatchColor.A > 0) ||
		(paint.Stroke.A > 0 && paint.LineWidth > 0)
}

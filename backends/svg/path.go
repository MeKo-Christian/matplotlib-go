package svg

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if rr := r.activeRaster(); rr != nil {
		rr.Path(p, paint)
		return
	}
	if !p.Validate() || paint == nil {
		return
	}
	if render.DrawPathWithEffects(r, p, paint, r.Path) {
		return
	}

	d := buildPathData(affinePath(p, r.deviceFlip()))
	if d == "" {
		return
	}

	hasGradient := paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0
	hasPattern := paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0
	hasFill := paint.Fill.A > 0 || hasGradient || hasPattern
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasHatch && !hasStroke {
		return
	}

	var b strings.Builder
	b.WriteString(`<path`)
	writeAttr(&b, "d", d)
	writeForcedOpacity(&b, *paint)

	switch {
	case hasHatch:
		writeAttr(&b, "fill", "url(#"+r.registerHatch(*paint)+")")
	case hasGradient:
		writeAttr(&b, "fill", "url(#"+r.registerGradient(&paint.FillGradient)+")")
	case hasPattern:
		writeAttr(&b, "fill", "url(#"+r.registerPatternFill(&paint.FillPattern)+")")
	case paint.Fill.A > 0:
		writeColorAttrs(&b, "fill", paint.Fill, forcedOpacity(*paint))
	default:
		writeAttr(&b, "fill", "none")
	}

	writeStrokeAttrs(&b, *paint)

	b.WriteString(" />")

	r.nodes = append(r.nodes, svgNode{
		content:   b.String(),
		clipIDs:   r.currentClipIDs(),
		filterIDs: r.currentFilterIDs(),
	})
}

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

func (r *Renderer) DrawPathEffectFilter(path geom.Path, paint render.Paint, effect render.PathEffect, draw func(geom.Path, *render.Paint)) bool {
	if r == nil || draw == nil {
		return false
	}
	id, ok := r.registerPathEffectFilter(effect)
	if !ok {
		return false
	}
	r.filterStack = append(r.filterStack, id)
	draw(path, &paint)
	r.filterStack = r.filterStack[:len(r.filterStack)-1]
	return true
}

func (r *Renderer) SupportsPathEffectFilter(effect render.PathEffect) bool {
	_, ok := normalizePathEffectFilter(effect)
	return ok
}

func (r *Renderer) currentFilterIDs() []string {
	if len(r.filterStack) == 0 {
		return nil
	}
	out := make([]string, len(r.filterStack))
	copy(out, r.filterStack)
	return out
}

func (r *Renderer) registerPathEffectFilter(effect render.PathEffect) (string, bool) {
	name, ok := normalizePathEffectFilter(effect)
	if !ok {
		return "", false
	}
	radius := effect.FilterRadius
	if radius <= 0 {
		radius = 1
	}
	key := filterKey(name, radius)
	if id, ok := r.filterDefs[key]; ok {
		return id, true
	}

	r.filterIDCounter++
	id := r.defID("filter", key, r.filterIDCounter)
	r.filterDefs[key] = id
	r.filterOrder = append(r.filterOrder, filterDef{id: id, name: name, radius: radius})
	return id, true
}

func normalizePathEffectFilter(effect render.PathEffect) (string, bool) {
	name := strings.ToLower(strings.TrimSpace(effect.Filter))
	switch name {
	case "blur", "gaussian", "gaussian-blur", "shadow":
		return "blur", true
	default:
		return "", false
	}
}

func buildPathData(p geom.Path) string {
	if len(p.C) == 0 {
		return ""
	}

	var b strings.Builder
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(p.V) {
				return ""
			}
			pt := quantizePt(p.V[vi])
			vi++
			b.WriteString("M ")
			b.WriteString(formatFloat(pt.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(pt.Y))
		case geom.LineTo:
			if vi >= len(p.V) {
				return ""
			}
			pt := quantizePt(p.V[vi])
			vi++
			b.WriteString(" L ")
			b.WriteString(formatFloat(pt.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(pt.Y))
		case geom.QuadTo:
			if vi+1 >= len(p.V) {
				return ""
			}
			ctrl := quantizePt(p.V[vi])
			to := quantizePt(p.V[vi+1])
			vi += 2
			b.WriteString(" Q ")
			b.WriteString(formatFloat(ctrl.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(ctrl.Y))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.Y))
		case geom.CubicTo:
			if vi+2 >= len(p.V) {
				return ""
			}
			c1 := quantizePt(p.V[vi])
			c2 := quantizePt(p.V[vi+1])
			to := quantizePt(p.V[vi+2])
			vi += 3
			b.WriteString(" C ")
			b.WriteString(formatFloat(c1.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(c1.Y))
			b.WriteString(" ")
			b.WriteString(formatFloat(c2.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(c2.Y))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.Y))
		case geom.ClosePath:
			b.WriteString(" Z")
		default:
			return ""
		}
	}

	d := b.String()
	return strings.TrimSpace(d)
}

func dashedArray(dashes []float64) string {
	if len(dashes) < 2 {
		return ""
	}

	var b strings.Builder
	for i := 0; i < len(dashes)-1; i += 2 {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(formatFloat(dashes[i]))
		b.WriteString(",")
		b.WriteString(formatFloat(dashes[i+1]))
	}

	return b.String()
}

func mapLineJoin(v render.LineJoin) string {
	switch v {
	case render.JoinRound:
		return "round"
	case render.JoinBevel:
		return "bevel"
	default:
		return "miter"
	}
}

func mapLineCap(v render.LineCap) string {
	switch v {
	case render.CapButt:
		return "butt"
	case render.CapRound:
		return "round"
	case render.CapSquare:
		return "square"
	default:
		return "butt"
	}
}

func writeColorAttrs(b *strings.Builder, attr string, c render.Color, forced bool) {
	colorValue, alpha := colorToStyle(c)
	writeAttr(b, attr, colorValue)
	if !forced && alpha < 1 {
		writeFloatAttr(b, attr+"-opacity", alpha)
	}
}

func writeStrokeAttrs(b *strings.Builder, paint render.Paint) {
	if paint.Stroke.A <= 0 || paint.LineWidth <= 0 {
		writeAttr(b, "stroke", "none")
		return
	}

	writeColorAttrs(b, "stroke", paint.Stroke, forcedOpacity(paint))
	writeFloatAttr(b, "stroke-width", paint.LineWidth)
	writeAttr(b, "stroke-linejoin", mapLineJoin(paint.LineJoin))
	writeAttr(b, "stroke-linecap", mapLineCap(paint.LineCap))
	if paint.MiterLimit > 0 {
		writeFloatAttr(b, "stroke-miterlimit", paint.MiterLimit)
	}
	if len(paint.Dashes) >= 2 {
		writeAttr(b, "stroke-dasharray", dashedArray(paint.Dashes))
	}
}

func forcedOpacity(paint render.Paint) bool {
	return paint.ForceAlpha && clamp01(clampFloat(paint.Alpha)) < 1
}

func writeForcedOpacity(b *strings.Builder, paint render.Paint) {
	if forcedOpacity(paint) {
		writeFloatAttr(b, "opacity", clamp01(clampFloat(paint.Alpha)))
	}
}

func filterKey(name string, radius float64) string {
	return strings.Join([]string{name, formatFloat(radius)}, "\x00")
}

func writeFilterDef(b *strings.Builder, filter filterDef) {
	b.WriteString(`    <filter id="`)
	b.WriteString(filter.id)
	b.WriteString(`" x="-20%" y="-20%" width="140%" height="140%">`)
	switch filter.name {
	case "blur":
		b.WriteString(`<feGaussianBlur`)
		writeFloatAttr(b, "stdDeviation", filter.radius)
		b.WriteString(` />`)
	default:
		b.WriteString(`<feComposite operator="over" />`)
	}
	b.WriteString("</filter>\n")
}

package svg

import (
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// gradientDef captures one renderer-neutral gradient as an SVG <defs> entry.
//
// The struct stores both linear and radial geometry to keep emission tabular;
// the kind field discriminates which subset of fields is valid.
type gradientDef struct {
	id           string
	kind         render.GradientKind
	x1, y1       float64
	x2, y2       float64
	cx, cy, rr   float64
	fx, fy       float64
	hasFocal     bool
	stops        []render.GradientStop
	transform    geom.Affine
	hasTransform bool
}

// patternFillDef captures one renderer-neutral pattern fill as an SVG <defs>
// entry. The path is pre-serialized to its `d` attribute so emission can reuse
// the same compact float formatting as solid paths.
type patternFillDef struct {
	id           string
	cell         geom.Rect
	pathData     string
	foreground   render.Color
	background   render.Color
	lineWidth    float64
	transform    geom.Affine
	hasTransform bool
}

// SupportsGradientFill reports that the SVG backend renders Paint.FillGradient
// natively via <linearGradient>/<radialGradient> defs.
func (r *Renderer) SupportsGradientFill() bool { return true }

// SupportsPatternFill reports that the SVG backend renders Paint.FillPattern
// natively via <pattern> defs.
func (r *Renderer) SupportsPatternFill() bool { return true }

func (r *Renderer) registerGradient(g *render.GradientFill) string {
	stops := normalizeStops(g.Stops)
	key := gradientKey(g, stops)
	if id, ok := r.gradientDefs[key]; ok {
		return id
	}

	r.gradientIDCounter++
	id := r.defID("gradient", key, r.gradientIDCounter)
	r.gradientDefs[key] = id

	def := gradientDef{
		id:           id,
		kind:         g.Kind,
		stops:        stops,
		transform:    g.Transform,
		hasTransform: g.HasTransform,
	}
	switch g.Kind {
	case render.LinearGradient:
		def.x1 = g.Start.X
		def.y1 = g.Start.Y
		def.x2 = g.End.X
		def.y2 = g.End.Y
	case render.RadialGradient:
		def.cx = g.Center.X
		def.cy = g.Center.Y
		def.rr = g.Radius
		if g.Focal != (geom.Pt{}) || g.Focal == g.Center {
			def.fx = g.Focal.X
			def.fy = g.Focal.Y
			def.hasFocal = true
		}
	}
	r.gradientOrder = append(r.gradientOrder, def)
	return id
}

func (r *Renderer) registerPatternFill(p *render.PatternFill) string {
	pathData := buildPathData(p.Path)
	key := patternFillKey(p, pathData)
	if id, ok := r.patternFillDefs[key]; ok {
		return id
	}

	r.patternFillIDCounter++
	id := r.defID("pattern", key, r.patternFillIDCounter)
	r.patternFillDefs[key] = id

	r.patternFillOrder = append(r.patternFillOrder, patternFillDef{
		id:           id,
		cell:         p.Cell,
		pathData:     pathData,
		foreground:   p.Foreground,
		background:   p.Background,
		lineWidth:    p.LineWidth,
		transform:    p.Transform,
		hasTransform: p.HasTransform,
	})
	return id
}

func normalizeStops(in []render.GradientStop) []render.GradientStop {
	if len(in) == 0 {
		return nil
	}
	out := make([]render.GradientStop, len(in))
	copy(out, in)
	return out
}

func gradientKey(g *render.GradientFill, stops []render.GradientStop) string {
	parts := make([]string, 0, 8+len(stops)*5)
	parts = append(
		parts,
		strconv.Itoa(int(g.Kind)),
		formatFloat(g.Start.X), formatFloat(g.Start.Y),
		formatFloat(g.End.X), formatFloat(g.End.Y),
		formatFloat(g.Center.X), formatFloat(g.Center.Y),
		formatFloat(g.Focal.X), formatFloat(g.Focal.Y),
		formatFloat(g.Radius),
		strconv.FormatBool(g.HasTransform),
	)
	if g.HasTransform {
		parts = append(
			parts,
			formatFloat(g.Transform.A), formatFloat(g.Transform.B),
			formatFloat(g.Transform.C), formatFloat(g.Transform.D),
			formatFloat(g.Transform.E), formatFloat(g.Transform.F),
		)
	}
	for _, s := range stops {
		parts = append(
			parts,
			formatFloat(s.Offset),
			formatFloat(s.Color.R), formatFloat(s.Color.G),
			formatFloat(s.Color.B), formatFloat(s.Color.A),
		)
	}
	return strings.Join(parts, "\x00")
}

func patternFillKey(p *render.PatternFill, pathData string) string {
	parts := []string{
		p.ID,
		pathData,
		formatFloat(p.Cell.Min.X), formatFloat(p.Cell.Min.Y),
		formatFloat(p.Cell.Max.X), formatFloat(p.Cell.Max.Y),
		formatFloat(p.Foreground.R), formatFloat(p.Foreground.G),
		formatFloat(p.Foreground.B), formatFloat(p.Foreground.A),
		formatFloat(p.Background.R), formatFloat(p.Background.G),
		formatFloat(p.Background.B), formatFloat(p.Background.A),
		formatFloat(p.LineWidth),
		strconv.FormatBool(p.HasTransform),
	}
	if p.HasTransform {
		parts = append(
			parts,
			formatFloat(p.Transform.A), formatFloat(p.Transform.B),
			formatFloat(p.Transform.C), formatFloat(p.Transform.D),
			formatFloat(p.Transform.E), formatFloat(p.Transform.F),
		)
	}
	return strings.Join(parts, "\x00")
}

func writeGradientDef(b *strings.Builder, g *gradientDef) {
	switch g.kind {
	case render.LinearGradient:
		b.WriteString(`    <linearGradient id="`)
		b.WriteString(g.id)
		b.WriteString(`" gradientUnits="userSpaceOnUse"`)
		writeFloatAttr(b, "x1", g.x1)
		writeFloatAttr(b, "y1", g.y1)
		writeFloatAttr(b, "x2", g.x2)
		writeFloatAttr(b, "y2", g.y2)
		if g.hasTransform {
			writeAttr(b, "gradientTransform", matrixTransform(g.transform))
		}
		b.WriteString(">")
		writeGradientStops(b, g.stops)
		b.WriteString("</linearGradient>\n")
	case render.RadialGradient:
		b.WriteString(`    <radialGradient id="`)
		b.WriteString(g.id)
		b.WriteString(`" gradientUnits="userSpaceOnUse"`)
		writeFloatAttr(b, "cx", g.cx)
		writeFloatAttr(b, "cy", g.cy)
		writeFloatAttr(b, "r", g.rr)
		if g.hasFocal {
			writeFloatAttr(b, "fx", g.fx)
			writeFloatAttr(b, "fy", g.fy)
		}
		if g.hasTransform {
			writeAttr(b, "gradientTransform", matrixTransform(g.transform))
		}
		b.WriteString(">")
		writeGradientStops(b, g.stops)
		b.WriteString("</radialGradient>\n")
	}
}

func writeGradientStops(b *strings.Builder, stops []render.GradientStop) {
	for _, s := range stops {
		b.WriteString(`<stop`)
		writeFloatAttr(b, "offset", clamp01(s.Offset))
		writeAttr(b, "stop-color", colorToHex(s.Color))
		alpha := clamp01(s.Color.A)
		if alpha < 1 {
			writeFloatAttr(b, "stop-opacity", alpha)
		}
		b.WriteString(` />`)
	}
}

func writePatternFillDef(b *strings.Builder, p *patternFillDef) {
	w := p.cell.W()
	h := p.cell.H()
	if w <= 0 {
		w = 16
	}
	if h <= 0 {
		h = 16
	}

	b.WriteString(`    <pattern id="`)
	b.WriteString(p.id)
	b.WriteString(`" patternUnits="userSpaceOnUse"`)
	writeFloatAttr(b, "x", p.cell.Min.X)
	writeFloatAttr(b, "y", p.cell.Min.Y)
	writeFloatAttr(b, "width", w)
	writeFloatAttr(b, "height", h)
	if p.hasTransform {
		writeAttr(b, "patternTransform", matrixTransform(p.transform))
	}
	b.WriteString(">")
	if p.background.A > 0 {
		b.WriteString(`<rect x="0" y="0"`)
		writeFloatAttr(b, "width", w)
		writeFloatAttr(b, "height", h)
		writeColorAttrs(b, "fill", p.background, false)
		b.WriteString(` />`)
	}
	if p.pathData != "" && p.foreground.A > 0 {
		b.WriteString(`<path`)
		writeAttr(b, "d", p.pathData)
		if p.lineWidth > 0 {
			writeAttr(b, "fill", "none")
			writeColorAttrs(b, "stroke", p.foreground, false)
			writeFloatAttr(b, "stroke-width", p.lineWidth)
			writeAttr(b, "stroke-linecap", "butt")
		} else {
			writeColorAttrs(b, "fill", p.foreground, false)
			writeAttr(b, "stroke", "none")
		}
		b.WriteString(` />`)
	}
	b.WriteString("</pattern>\n")
}

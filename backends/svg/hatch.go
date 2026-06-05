package svg

import (
	"math"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/backends/internal/vectorhatch"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) SupportsNativeHatch() bool { return true }

func (r *Renderer) registerHatch(paint render.Paint) string {
	lineWidth := paint.HatchLineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	spacing := paint.HatchSpacing
	if spacing <= 0 {
		spacing = 8
	}
	forced := forcedOpacity(paint)
	key := hatchKey(paint.Hatch, paint.Fill, paint.HatchColor, lineWidth, spacing, forced)
	if id, ok := r.hatchDefs[key]; ok {
		return id
	}

	r.hatchIDCounter++
	id := r.defID("hatch", key, r.hatchIDCounter)
	r.hatchDefs[key] = id
	r.hatchOrder = append(r.hatchOrder, hatchDef{
		id:        id,
		hatch:     paint.Hatch,
		faceColor: paint.Fill,
		lineColor: paint.HatchColor,
		lineWidth: lineWidth,
		spacing:   spacing,
		forced:    forced,
	})
	return id
}

func hatchKey(hatch string, face, line render.Color, width, spacing float64, forced bool) string {
	return strings.Join([]string{
		hatch,
		formatFloat(face.R),
		formatFloat(face.G),
		formatFloat(face.B),
		formatFloat(face.A),
		formatFloat(line.R),
		formatFloat(line.G),
		formatFloat(line.B),
		formatFloat(line.A),
		formatFloat(width),
		formatFloat(spacing),
		strconv.FormatBool(forced),
	}, "\x00")
}

func writeHatchDef(b *strings.Builder, hatch hatchDef) {
	b.WriteString(`    <pattern id="`)
	b.WriteString(hatch.id)
	b.WriteString(`" patternUnits="userSpaceOnUse" width="72" height="72">`)
	if hatch.faceColor.A > 0 {
		b.WriteString(`<rect x="0" y="0" width="72" height="72"`)
		writeColorAttrs(b, "fill", hatch.faceColor, hatch.forced)
		b.WriteString(` />`)
	}
	if hatch.lineColor.A > 0 {
		d := hatchPathData(hatch.hatch, hatch.spacing)
		if d != "" {
			b.WriteString(`<path`)
			writeAttr(b, "d", d)
			writeAttr(b, "fill", "none")
			writeColorAttrs(b, "stroke", hatch.lineColor, hatch.forced)
			writeFloatAttr(b, "stroke-width", hatch.lineWidth)
			writeAttr(b, "stroke-linecap", "butt")
			b.WriteString(` />`)
		}
		writeHatchShapeDefs(b, hatch)
	}
	b.WriteString("</pattern>\n")
}

func writeHatchShapeDefs(b *strings.Builder, hatch hatchDef) {
	for _, shape := range vectorhatch.ShapePaths(hatch.hatch, hatch.spacing) {
		d := buildPathData(shape.Path)
		if d == "" {
			continue
		}
		b.WriteString(`<path`)
		writeAttr(b, "d", d)
		if shape.Filled {
			writeColorAttrs(b, "fill", hatch.lineColor, hatch.forced)
			writeAttr(b, "stroke", "none")
		} else {
			writeAttr(b, "fill", "none")
			writeColorAttrs(b, "stroke", hatch.lineColor, hatch.forced)
			writeFloatAttr(b, "stroke-width", hatch.lineWidth)
			writeAttr(b, "stroke-linecap", "butt")
		}
		b.WriteString(` />`)
	}
}

func hatchPathData(hatch string, spacing float64) string {
	if spacing <= 0 {
		spacing = 8
	}
	var b strings.Builder
	writeHatchLines := func(count int, draw func(float64)) {
		if count <= 0 {
			return
		}
		step := math.Max(2, spacing/float64(count))
		for v := -72.0; v <= 144; v += step {
			draw(v)
		}
	}
	line := func(x1, y1, x2, y2 float64) {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("M ")
		b.WriteString(formatFloat(x1))
		b.WriteByte(' ')
		b.WriteString(formatFloat(y1))
		b.WriteString(" L ")
		b.WriteString(formatFloat(x2))
		b.WriteByte(' ')
		b.WriteString(formatFloat(y2))
	}

	verticalCount := strings.Count(hatch, "|") + strings.Count(hatch, "+")
	horizontalCount := strings.Count(hatch, "-") + strings.Count(hatch, "+")
	slashCount := strings.Count(hatch, "/") + strings.Count(hatch, "x") + strings.Count(hatch, "X")
	backslashCount := strings.Count(hatch, `\`) + strings.Count(hatch, "x") + strings.Count(hatch, "X")

	writeHatchLines(verticalCount, func(x float64) { line(x, 0, x, 72) })
	writeHatchLines(horizontalCount, func(y float64) { line(0, y, 72, y) })
	writeHatchLines(slashCount, func(x float64) { line(x, 72, x+72, 0) })
	writeHatchLines(backslashCount, func(x float64) { line(x, 0, x+72, 72) })
	return b.String()
}

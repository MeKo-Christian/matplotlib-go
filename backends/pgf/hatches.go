package pgf

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/backends/internal/vectorhatch"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// SupportsNativeHatch reports that the PGF backend consumes hatch metadata
// directly in Path.
func (r *Renderer) SupportsNativeHatch() bool { return true }

func (r *Renderer) writeHatchFill(path geom.Path, paint *render.Paint) {
	r.writeHatchFillTo(&r.content, path, paint)
}

func (r *Renderer) writeHatchFillTo(w *strings.Builder, path geom.Path, paint *render.Paint) {
	if paint == nil || paint.Hatch == "" || paint.HatchColor.A <= 0 {
		return
	}
	w.WriteString("\\pgfscope\n")
	if !writePathOps(w, path) {
		w.WriteString("\\endpgfscope\n")
		return
	}
	w.WriteString("\\pgfusepath{clip}\n")
	writeFillOpacity(w, 1)
	writeStrokeOpacity(w, paint.HatchColor.A)
	writeStrokeColor(w, r.colorName(paint.HatchColor))
	lineWidth := paint.HatchLineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	fmt.Fprintf(w, "\\pgfsetlinewidth{%spt}\n", shortFloat(lineWidth))
	for _, line := range hatchPatternLines(paint.Hatch, paint.HatchSpacing) {
		fmt.Fprintf(w, "\\pgfpathmoveto{\\pgfpoint{%spt}{%spt}}\n", shortFloat(line[0].X), shortFloat(line[0].Y))
		fmt.Fprintf(w, "\\pgfpathlineto{\\pgfpoint{%spt}{%spt}}\n", shortFloat(line[1].X), shortFloat(line[1].Y))
		w.WriteString("\\pgfusepath{stroke}\n")
	}
	writeHatchShapeOps(w, paint.Hatch, paint.HatchSpacing)
	w.WriteString("\\endpgfscope\n")
}

func writeHatchShapeOps(w *strings.Builder, hatch string, spacing float64) {
	for _, shape := range vectorhatch.ShapePaths(hatch, spacing) {
		if !writePathOps(w, shape.Path) {
			continue
		}
		if shape.Filled {
			w.WriteString("\\pgfusepath{fill}\n")
		} else {
			w.WriteString("\\pgfusepath{stroke}\n")
		}
	}
}

func hatchPatternLines(hatch string, spacing float64) [][2]geom.Pt {
	if spacing <= 0 {
		spacing = 8
	}
	lines := make([][2]geom.Pt, 0)
	writeHatchLines := func(count int, draw func(float64)) {
		if count <= 0 {
			return
		}
		step := math.Max(2, spacing/float64(count))
		for v := -72.0; v <= 144; v += step {
			draw(v)
		}
	}
	add := func(x1, y1, x2, y2 float64) {
		lines = append(lines, [2]geom.Pt{{X: x1, Y: y1}, {X: x2, Y: y2}})
	}
	verticalCount := strings.Count(hatch, "|") + strings.Count(hatch, "+")
	horizontalCount := strings.Count(hatch, "-") + strings.Count(hatch, "+")
	slashCount := strings.Count(hatch, "/") + strings.Count(hatch, "x") + strings.Count(hatch, "X")
	backslashCount := strings.Count(hatch, `\`) + strings.Count(hatch, "x") + strings.Count(hatch, "X")

	writeHatchLines(verticalCount, func(x float64) { add(x, 0, x, 72) })
	writeHatchLines(horizontalCount, func(y float64) { add(0, y, 72, y) })
	writeHatchLines(slashCount, func(x float64) { add(x, 72, x+72, 0) })
	writeHatchLines(backslashCount, func(x float64) { add(x, 0, x+72, 72) })
	return lines
}

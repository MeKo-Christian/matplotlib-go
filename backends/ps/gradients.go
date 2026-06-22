package ps

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// SupportsGradientFill reports that the PostScript backend renders
// Paint.FillGradient natively via Level-3 axial/radial shading dictionaries
// painted with the shfill operator. Implements render.GradientFiller.
func (r *Renderer) SupportsGradientFill() bool { return true }

// SupportsPatternFill reports that the PostScript backend renders
// Paint.FillPattern natively by tiling the pattern cell inside a path clip.
// Implements render.PatternFiller.
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

// writeGradientFill clips to the path and paints the gradient with shfill.
//
// PostScript is natively y-up display space (matching the matplotlib-go
// coordinate frame), so the gradient endpoints — already in y-up display
// coordinates — are emitted directly without a device flip.
func (r *Renderer) writeGradientFill(w *strings.Builder, path geom.Path, paint *render.Paint) bool {
	if !hasGradientFill(paint) {
		return false
	}
	w.WriteString("gsave\n")
	if !writePathOps(w, path) {
		w.WriteString("grestore\n")
		return false
	}
	w.WriteString("clip\nnewpath\n")
	w.WriteString(psShadingDictionary(&paint.FillGradient))
	w.WriteString(" shfill\ngrestore\n")
	return true
}

// writePatternFill clips to the path and tiles the pattern cell across the
// path's bounding box. PostScript has no first-class colored fill+stroke
// pattern with a renderer-neutral background, so the cell is materialized the
// same way the AGG backend does: a background rectangle (when opaque) plus the
// foreground geometry per cell, all confined by the clip.
func (r *Renderer) writePatternFill(w *strings.Builder, path geom.Path, paint *render.Paint) bool {
	if !hasPatternFill(paint) {
		return false
	}
	pattern := paint.FillPattern

	cell := pattern.Cell
	cellW := cell.W()
	cellH := cell.H()
	if cellW <= 0 || cellH <= 0 {
		if bounds, ok := pattern.Path.Bounds(); ok {
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

	bounds, ok := path.Bounds()
	if !ok {
		return false
	}
	if !writePathOps(w, path) {
		return false
	}
	w.WriteString("gsave\nclip\nnewpath\n")

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
					writeFillColor(w, bg)
					w.WriteString("fill\n")
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
				writeStrokeColor(w, fg)
				fmt.Fprintf(w, "%s setlinewidth\n", shortFloat(pattern.LineWidth))
				w.WriteString("stroke\n")
			} else {
				writeFillColor(w, fg)
				w.WriteString("fill\n")
			}
		}
	}
	w.WriteString("grestore\n")
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

func patternTilePath(pattern *render.PatternFill, offset geom.Pt) geom.Path {
	transform := geom.Affine{A: 1, D: 1, E: offset.X, F: offset.Y}
	if pattern.HasTransform {
		transform = transform.Mul(pattern.Transform)
	}
	return pattern.Path.Transformed(transform)
}

// psShadingDictionary builds a PostScript Level-3 shading dictionary. The dict
// is byte-compatible with the PDF shading dictionaries (PostScript and PDF
// share the type-2/3 shading and type-2/3 function formats); only the painting
// operator differs (shfill consumes the inline dict, PDF's sh a named one).
func psShadingDictionary(gradient *render.GradientFill) string {
	stops := normalizeGradientStops(gradient.Stops)
	switch gradient.Kind {
	case render.LinearGradient:
		start := transformedGradientPoint(gradient.Start, gradient)
		end := transformedGradientPoint(gradient.End, gradient)
		return fmt.Sprintf(
			"<< /ShadingType 2 /ColorSpace /DeviceRGB /Coords [%s %s %s %s] /Domain [0 1] /Function %s /Extend [true true] >>",
			shortFloat(start.X), shortFloat(start.Y),
			shortFloat(end.X), shortFloat(end.Y),
			gradientFunctionDictionary(stops),
		)
	case render.RadialGradient:
		center := transformedGradientPoint(gradient.Center, gradient)
		focal := center
		if gradient.Focal != (geom.Pt{}) {
			focal = transformedGradientPoint(gradient.Focal, gradient)
		}
		radius := transformedGradientRadius(gradient.Radius, gradient)
		return fmt.Sprintf(
			"<< /ShadingType 3 /ColorSpace /DeviceRGB /Coords [%s %s 0 %s %s %s] /Domain [0 1] /Function %s /Extend [true true] >>",
			shortFloat(center.X), shortFloat(center.Y),
			shortFloat(focal.X), shortFloat(focal.Y), shortFloat(radius),
			gradientFunctionDictionary(stops),
		)
	default:
		return fmt.Sprintf(
			"<< /ShadingType 2 /ColorSpace /DeviceRGB /Coords [0 0 1 0] /Domain [0 1] /Function %s /Extend [true true] >>",
			gradientFunctionDictionary(stops),
		)
	}
}

func transformedGradientPoint(p geom.Pt, gradient *render.GradientFill) geom.Pt {
	if !gradient.HasTransform {
		return p
	}
	return gradient.Transform.Apply(p)
}

func transformedGradientRadius(radius float64, gradient *render.GradientFill) float64 {
	if radius == 0 {
		return 0
	}
	if !gradient.HasTransform {
		return radius
	}
	xScale := math.Hypot(gradient.Transform.A, gradient.Transform.B)
	yScale := math.Hypot(gradient.Transform.C, gradient.Transform.D)
	return radius * math.Max(xScale, yScale)
}

func gradientFunctionDictionary(stops []render.GradientStop) string {
	stops = normalizeGradientStops(stops)
	if len(stops) == 0 {
		stops = []render.GradientStop{
			{Offset: 0, Color: render.Color{A: 1}},
			{Offset: 1, Color: render.Color{A: 1}},
		}
	}
	if len(stops) == 1 {
		stops = []render.GradientStop{
			{Offset: 0, Color: stops[0].Color},
			{Offset: 1, Color: stops[0].Color},
		}
	}
	if len(stops) == 2 {
		return type2FunctionDictionary(stops[0].Color, stops[1].Color)
	}

	var functions strings.Builder
	var bounds strings.Builder
	var encode strings.Builder
	for i := 0; i < len(stops)-1; i++ {
		if i > 0 {
			functions.WriteByte(' ')
			encode.WriteByte(' ')
		}
		functions.WriteString(type2FunctionDictionary(stops[i].Color, stops[i+1].Color))
		encode.WriteString("0 1")
	}
	for i := 1; i < len(stops)-1; i++ {
		if i > 1 {
			bounds.WriteByte(' ')
		}
		bounds.WriteString(shortFloat(stops[i].Offset))
	}
	return fmt.Sprintf(
		"<< /FunctionType 3 /Domain [0 1] /Functions [%s] /Bounds [%s] /Encode [%s] >>",
		functions.String(), bounds.String(), encode.String(),
	)
}

func type2FunctionDictionary(c0, c1 render.Color) string {
	return fmt.Sprintf(
		"<< /FunctionType 2 /Domain [0 1] /C0 %s /C1 %s /N 1 >>",
		psColorArray(c0), psColorArray(c1),
	)
}

func psColorArray(c render.Color) string {
	return fmt.Sprintf(
		"[%s %s %s]",
		shortFloat(clamp01(c.R)),
		shortFloat(clamp01(c.G)),
		shortFloat(clamp01(c.B)),
	)
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

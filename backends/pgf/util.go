package pgf

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func normalizedAffine(affine geom.Affine) geom.Affine {
	if affine == (geom.Affine{}) {
		return geom.Identity()
	}
	return affine
}

func cloneRectPtr(rect *geom.Rect) *geom.Rect {
	if rect == nil {
		return nil
	}
	cloned := *rect
	return &cloned
}

func normalizeRect(rect geom.Rect) geom.Rect {
	minX, maxX := rect.Min.X, rect.Max.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := rect.Min.Y, rect.Max.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}
}

func writeTransform(w *strings.Builder, transform geom.Affine) {
	fmt.Fprintf(
		w, "\\pgftransformcm{%s}{%s}{%s}{%s}{\\pgfpoint{%spt}{%spt}}\n",
		shortFloat(transform.A),
		shortFloat(transform.B),
		shortFloat(transform.C),
		shortFloat(transform.D),
		shortFloat(transform.E),
		shortFloat(transform.F),
	)
}

func (r *Renderer) colorName(c render.Color) string {
	c = render.Color{R: clamp01(c.R), G: clamp01(c.G), B: clamp01(c.B), A: clamp01(c.A)}
	key := strings.Join([]string{shortFloat(c.R), shortFloat(c.G), shortFloat(c.B)}, ",")
	if name, ok := r.colorNames[key]; ok {
		return name
	}
	name := fmt.Sprintf("mplgpgfcolor%d", len(r.colorNames)+1)
	r.colorNames[key] = name
	// Color definitions are collected separately and injected at the pgfpicture's
	// outermost group (see Begin/End). Emitting \definecolor inline would scope it
	// to whatever \pgfscope happens to be active, so later references to the same
	// cached name would hit an undefined color once that scope closes.
	fmt.Fprintf(&r.colorDefs, "\\definecolor{%s}{rgb}{%s,%s,%s}\n", name, shortFloat(c.R), shortFloat(c.G), shortFloat(c.B))
	return name
}

func writeFillColor(w *strings.Builder, colorName string) {
	fmt.Fprintf(w, "\\pgfsetfillcolor{%s}\n", colorName)
}

func writeStrokeColor(w *strings.Builder, colorName string) {
	fmt.Fprintf(w, "\\pgfsetstrokecolor{%s}\n", colorName)
}

func writeFillOpacity(w *strings.Builder, alpha float64) {
	fmt.Fprintf(w, "\\pgfsetfillopacity{%s}\n", shortFloat(clamp01(alpha)))
}

func writeStrokeOpacity(w *strings.Builder, alpha float64) {
	fmt.Fprintf(w, "\\pgfsetstrokeopacity{%s}\n", shortFloat(clamp01(alpha)))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func shortFloat(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0"
	}
	if v == 0 {
		return "0"
	}
	s := strconv.FormatFloat(v, 'f', 6, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	if s == "" || s == "-0" {
		return "0"
	}
	return s
}

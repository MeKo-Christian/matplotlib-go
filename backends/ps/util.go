package ps

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

func textPathAffine(origin geom.Pt) geom.Affine {
	// render.TextPath emits y-up display outlines, so positioning is a plain
	// translation to the baseline origin (no baseline reflection).
	return translateAffine(origin)
}

func rotationAffine(angle float64, pivot geom.Pt) geom.Affine {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return translateAffine(pivot).
		Mul(geom.Affine{A: cos, B: sin, C: -sin, D: cos}).
		Mul(translateAffine(geom.Pt{X: -pivot.X, Y: -pivot.Y}))
}

func translateAffine(p geom.Pt) geom.Affine {
	return geom.Affine{A: 1, D: 1, E: p.X, F: p.Y}
}

func affinePath(path geom.Path, affine geom.Affine) geom.Path {
	if len(path.V) == 0 {
		return path
	}
	out := geom.Path{
		V: make([]geom.Pt, len(path.V)),
		C: append([]geom.Cmd(nil), path.C...),
	}
	for i, pt := range path.V {
		out.V[i] = affine.Apply(pt)
	}
	return out
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

func writeStrokeColor(w *strings.Builder, c render.Color) {
	writeColor(w, c)
}

func writeFillColor(w *strings.Builder, c render.Color) {
	writeColor(w, c)
}

func writeColor(w *strings.Builder, c render.Color) {
	fmt.Fprintf(w, "%s %s %s setrgbcolor\n", shortFloat(clamp01(c.R)), shortFloat(clamp01(c.G)), shortFloat(clamp01(c.B)))
}

func lineJoin(join render.LineJoin) int {
	switch join {
	case render.JoinRound:
		return 1
	case render.JoinBevel:
		return 2
	default:
		return 0
	}
}

func lineCap(lineCap render.LineCap) int {
	switch lineCap {
	case render.CapRound:
		return 1
	case render.CapSquare:
		return 2
	default:
		return 0
	}
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

package svg

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func matrixTransform(a geom.Affine) string {
	return "matrix(" + shortFloat(a.A) + " " + shortFloat(a.B) + " " +
		shortFloat(a.C) + " " + shortFloat(a.D) + " " +
		shortFloat(a.E) + " " + shortFloat(a.F) + ")"
}

func colorToHex(c render.Color) string {
	return fmt.Sprintf(
		"rgb(%d,%d,%d)",
		toByte(c.R),
		toByte(c.G),
		toByte(c.B),
	)
}

func colorToStyle(c render.Color) (string, float64) {
	return colorToHex(c), clamp01(c.A)
}

func normalizeSVGOptions(opts render.SVGOptions) render.SVGOptions {
	if opts.FontPolicy != render.SVGFontPolicyPath {
		opts.FontPolicy = render.SVGFontPolicyNone
	}
	if len(opts.Metadata) > 0 {
		metadata := make(map[string]string, len(opts.Metadata))
		for k, v := range opts.Metadata {
			metadata[k] = v
		}
		opts.Metadata = metadata
	}
	return opts
}

func (r *Renderer) defID(prefix, content string, sequence int) string {
	if r != nil && r.options.HashSalt != "" {
		sum := sha256.Sum256([]byte(r.options.HashSalt + content))
		return prefix + hex.EncodeToString(sum[:])[:10]
	}
	return prefix + strconv.Itoa(sequence)
}

func toByte(v float64) uint8 {
	v = clamp01(v)
	return uint8(v*255 + 0.5)
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

func quantize(v float64) float64 {
	return math.Round(v/quantizationGrid) * quantizationGrid
}

func quantizePt(p geom.Pt) geom.Pt {
	return geom.Pt{X: quantize(p.X), Y: quantize(p.Y)}
}

func normalizeRect(rect geom.Rect) geom.Rect {
	minX := rect.Min.X
	minY := rect.Min.Y
	maxX := rect.Max.X
	maxY := rect.Max.Y

	if maxX < minX {
		minX, maxX = maxX, minX
	}
	if maxY < minY {
		minY, maxY = maxY, minY
	}

	return geom.Rect{
		Min: geom.Pt{X: quantize(minX), Y: quantize(minY)},
		Max: geom.Pt{X: quantize(maxX), Y: quantize(maxY)},
	}
}

func writeAttr(b *strings.Builder, name, value string) {
	b.WriteString(" ")
	b.WriteString(name)
	b.WriteByte('=')
	b.WriteString(strconv.Quote(value))
}

func writeFloatAttr(b *strings.Builder, name string, value float64) {
	b.WriteString(" ")
	b.WriteString(name)
	b.WriteString("=\"")
	b.WriteString(formatFloat(value))
	b.WriteString("\"")
}

func formatFloat(v float64) string {
	return shortFloat(v)
}

func shortFloat(v float64) string {
	s := strconv.FormatFloat(clampFloat(v), 'f', 6, 64)
	if i := strings.IndexByte(s, '.'); i >= 0 {
		end := len(s)
		for end > i && s[end-1] == '0' {
			end--
		}
		if end > 0 && s[end-1] == '.' {
			end--
		}
		s = s[:end]
	}
	if s == "-0" || s == "" {
		s = "0"
	}
	return s
}

func rotateTransform(angleDeg, cx, cy float64) string {
	return matrixTransform(rotationAffine(angleDeg, cx, cy))
}

func rotationAffine(angleDeg, cx, cy float64) geom.Affine {
	rad := angleDeg * math.Pi / 180
	cos := math.Cos(rad)
	sin := math.Sin(rad)
	return geom.Affine{
		A: cos,
		B: sin,
		C: -sin,
		D: cos,
		E: cx - cos*cx + sin*cy,
		F: cy - sin*cx - cos*cy,
	}
}

func (r *Renderer) deviceFlip() geom.Affine {
	return geom.Affine{A: 1, D: -1, F: r.height}
}

func (r *Renderer) flipY(y float64) float64 {
	return r.height - y
}

func affinePath(path geom.Path, affine geom.Affine) geom.Path {
	if len(path.V) == 0 {
		return path
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, pt := range path.V {
		out.V[i] = affine.Apply(pt)
	}
	return out
}

func clampFloat(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}

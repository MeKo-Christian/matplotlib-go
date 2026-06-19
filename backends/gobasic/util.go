package gobasic

import (
	"encoding/binary"
	"hash/fnv"
	"image/color"
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// quantizationEpsilon is the precision limit for float values to ensure determinism.
// All floating point coordinates and measurements are snapped to this precision.
const (
	quantizationEpsilon = 1e-6
	defaultFontHeight   = 13.0
)

// quantize snaps a float64 value to quantizationEpsilon precision to eliminate
// tiny differences that could lead to cross-platform rendering variations.
func quantize(v float64) float64 {
	return math.Round(v/quantizationEpsilon) * quantizationEpsilon
}

func scalePointSize(size float64, dpi uint) float64 {
	if dpi == 0 {
		dpi = 72
	}
	return size * float64(dpi) / 72.0
}

// quantizePt quantizes both X and Y coordinates of a point.
func quantizePt(p geom.Pt) geom.Pt {
	return geom.Pt{
		X: quantize(p.X),
		Y: quantize(p.Y),
	}
}

// quantizePath quantizes all vertices in a path for deterministic rendering.
func quantizePath(p geom.Path) geom.Path {
	result := geom.Path{
		C: make([]geom.Cmd, len(p.C)),
		V: make([]geom.Pt, len(p.V)),
	}

	copy(result.C, p.C)
	for i, v := range p.V {
		result.V[i] = quantizePt(v)
	}

	return result
}

func clonePaths(paths []geom.Path) []geom.Path {
	if len(paths) == 0 {
		return nil
	}
	out := make([]geom.Path, len(paths))
	for i, path := range paths {
		out[i] = clonePath(path)
	}
	return out
}

func clonePath(path geom.Path) geom.Path {
	return geom.Path{
		V: append([]geom.Pt(nil), path.V...),
		C: append([]geom.Cmd(nil), path.C...),
	}
}

func hashPath(path geom.Path) uint64 {
	h := fnv.New64a()
	var buf [8]byte
	for _, cmd := range path.C {
		_, _ = h.Write([]byte{byte(cmd)})
	}
	for _, pt := range path.V {
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(quantize(pt.X)))
		_, _ = h.Write(buf[:])
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(quantize(pt.Y)))
		_, _ = h.Write(buf[:])
	}
	return h.Sum64()
}

func renderColorToRGBA(c render.Color) color.RGBA {
	return color.RGBA{
		R: toByte(c.R),
		G: toByte(c.G),
		B: toByte(c.B),
		A: toByte(c.A),
	}
}

func toByte(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

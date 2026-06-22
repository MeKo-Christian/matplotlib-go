package agg

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// gouraudDilation matches Matplotlib's AGG backend, which dilates each Gouraud
// triangle by 0.5 subpixels for numerically stable, seam-free antialiased
// rasterization (span_gen.triangle(..., 0.5) in _backend_agg.h).
const gouraudDilation = 0.5

// drawGouraudTriangle renders one interpolated-color triangle. When antialiased
// is true it routes through AGG's span_gouraud_rgba generator (the Matplotlib
// path); otherwise it falls back to the binary point-sampled rasterizer for
// crisp, coverage-free cells.
func (r *Renderer) drawGouraudTriangle(tri *render.GouraudTriangle, antialiased bool) {
	if tri == nil {
		return
	}
	if antialiased {
		r.drawGouraudTriangleAA(tri)
		return
	}
	img := r.ctx.image
	if img == nil || img.Width() <= 0 || img.Height() <= 0 {
		return
	}

	minX := int(math.Floor(math.Min(tri.P[0].X, math.Min(tri.P[1].X, tri.P[2].X))))
	maxX := int(math.Ceil(math.Max(tri.P[0].X, math.Max(tri.P[1].X, tri.P[2].X))))
	minY := int(math.Floor(math.Min(tri.P[0].Y, math.Min(tri.P[1].Y, tri.P[2].Y))))
	maxY := int(math.Ceil(math.Max(tri.P[0].Y, math.Max(tri.P[1].Y, tri.P[2].Y))))

	clipMinX, clipMinY := 0, 0
	clipMaxX, clipMaxY := img.Width()-1, img.Height()-1
	if r.clipRect != nil {
		clipMinX = maxInt(clipMinX, int(math.Floor(r.clipRect.Min.X)))
		clipMinY = maxInt(clipMinY, int(math.Floor(r.clipRect.Min.Y)))
		clipMaxX = minInt(clipMaxX, int(math.Ceil(r.clipRect.Max.X))-1)
		clipMaxY = minInt(clipMaxY, int(math.Ceil(r.clipRect.Max.Y))-1)
	}
	minX = maxInt(minX, clipMinX)
	minY = maxInt(minY, clipMinY)
	maxX = minInt(maxX, clipMaxX)
	maxY = minInt(maxY, clipMaxY)
	if minX > maxX || minY > maxY {
		return
	}

	area := edgeFunction(tri.P[0], tri.P[1], tri.P[2])
	if area == 0 || math.IsNaN(area) || math.IsInf(area, 0) {
		return
	}

	stride := img.Stride()
	if stride <= 0 {
		return
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			p := geom.Pt{X: float64(x) + 0.5, Y: float64(y) + 0.5}
			w0 := edgeFunction(tri.P[1], tri.P[2], p) / area
			w1 := edgeFunction(tri.P[2], tri.P[0], p) / area
			w2 := edgeFunction(tri.P[0], tri.P[1], p) / area
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			src := interpolateColor(tri.Color[0], tri.Color[1], tri.Color[2], w0, w1, w2)
			if src.A <= 0 {
				continue
			}
			off := y*stride + x*4
			if off < 0 || off+3 >= len(img.Data) {
				continue
			}
			blendPixelRGBA(img.Data[off:off+4], src)
		}
	}
}

// drawGouraudTriangleAA renders an antialiased Gouraud triangle through the AGG
// painter's span_gouraud_rgba generator, compositing into the shared buffer
// (r.ctx may be a clip-mask temp surface). Vertices are already in device space
// and clipping is established by the caller (applyClipRect / withClipPathMask).
func (r *Renderer) drawGouraudTriangleAA(tri *render.GouraudTriangle) {
	if r.ctx == nil || r.ctx.image == nil {
		return
	}
	area := edgeFunction(tri.P[0], tri.P[1], tri.P[2])
	if area == 0 || math.IsNaN(area) || math.IsInf(area, 0) {
		return
	}
	r.ctx.GouraudTriangle(
		tri.P[0].X, tri.P[0].Y,
		tri.P[1].X, tri.P[1].Y,
		tri.P[2].X, tri.P[2].Y,
		renderColorToAGG(tri.Color[0]),
		renderColorToAGG(tri.Color[1]),
		renderColorToAGG(tri.Color[2]),
		gouraudDilation,
	)
}

func edgeFunction(a, b, c geom.Pt) float64 {
	return (c.X-a.X)*(b.Y-a.Y) - (c.Y-a.Y)*(b.X-a.X)
}

func interpolateColor(c0, c1, c2 render.Color, w0, w1, w2 float64) render.Color {
	return render.Color{
		R: c0.R*w0 + c1.R*w1 + c2.R*w2,
		G: c0.G*w0 + c1.G*w1 + c2.G*w2,
		B: c0.B*w0 + c1.B*w1 + c2.B*w2,
		A: c0.A*w0 + c1.A*w1 + c2.A*w2,
	}
}

func blendPixelRGBA(dst []uint8, src render.Color) {
	sa := uint32(math.Round(clamp01(src.A) * 255))
	if sa == 0 {
		return
	}
	sr := uint8(math.Round(clamp01(src.R) * 255))
	sg := uint8(math.Round(clamp01(src.G) * 255))
	sb := uint8(math.Round(clamp01(src.B) * 255))
	if sa >= 255 {
		dst[0] = sr
		dst[1] = sg
		dst[2] = sb
		dst[3] = 255
		return
	}

	da := uint32(dst[3])
	combinedA := ((sa + da) << 8) - sa*da
	if combinedA == 0 {
		dst[0], dst[1], dst[2], dst[3] = 0, 0, 0, 0
		return
	}
	dst[3] = uint8(combinedA >> 8)
	dst[0] = fixedBlendChannel(dst[0], sr, uint8(da), uint8(sa), combinedA)
	dst[1] = fixedBlendChannel(dst[1], sg, uint8(da), uint8(sa), combinedA)
	dst[2] = fixedBlendChannel(dst[2], sb, uint8(da), uint8(sa), combinedA)
}

func fixedBlendChannel(dst, src, dstA, srcA uint8, combinedA uint32) uint8 {
	dstPremul := int64(uint32(dst) * uint32(dstA))
	numerator := ((int64(src) << 8) - dstPremul) * int64(srcA)
	numerator += dstPremul << 8
	return uint8(numerator / int64(combinedA))
}

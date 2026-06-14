package gobasic

import "github.com/cwbudde/matplotlib-go/geom"

// The core "display space" is y-UP (origin bottom-left, +Y up), matching
// matplotlib. The gobasic raster buffer (image.RGBA) is y-DOWN (origin
// top-left, +Y down). This backend therefore OWNS the device y-inversion: it
// converts y-up display coordinates to its native y-down buffer at the public
// method boundary, mirroring matplotlib's src/_backend_agg.h::draw_path which
// appends scaling(1,-1); translation(0,height) to the path transform.
//
// All internal rasterization (fillPath, drawStroke, clip masks, blendPixel)
// keeps operating in device space exactly as before; only the inputs are
// flipped at the boundary so ordinary content rasterizes byte-identically.

// deviceHeight returns the y-down buffer height used for the device flip.
func (r *Renderer) deviceHeight() float64 {
	if r.dst == nil {
		return 0
	}
	return float64(r.dst.Bounds().Dy())
}

// devY converts a y-up display Y coordinate to the y-down device Y.
func (r *Renderer) devY(y float64) float64 { return r.deviceHeight() - y }

// devPt converts a y-up display point to a y-down device point.
func (r *Renderer) devPt(p geom.Pt) geom.Pt {
	return geom.Pt{X: p.X, Y: r.deviceHeight() - p.Y}
}

// devRect converts a y-up display rectangle to a y-down device rectangle,
// swapping Min.Y/Max.Y so the result keeps Min.Y <= Max.Y.
func (r *Renderer) devRect(rc geom.Rect) geom.Rect {
	h := r.deviceHeight()
	return geom.Rect{
		Min: geom.Pt{X: rc.Min.X, Y: h - rc.Max.Y},
		Max: geom.Pt{X: rc.Max.X, Y: h - rc.Min.Y},
	}
}

// devPath returns a copy of p with every vertex flipped to device space.
func (r *Renderer) devPath(p geom.Path) geom.Path {
	out := p
	out.V = make([]geom.Pt, len(p.V))
	h := r.deviceHeight()
	for i, v := range p.V {
		out.V[i] = geom.Pt{X: v.X, Y: h - v.Y}
	}
	return out
}

// deviceFlipAffine returns the affine F mapping a y-up display point to the
// y-down device buffer: (x, y) -> (x, H - y).
func (r *Renderer) deviceFlipAffine() geom.Affine {
	return geom.Affine{A: 1, B: 0, C: 0, D: -1, E: 0, F: geom.F64(r.deviceHeight())}
}

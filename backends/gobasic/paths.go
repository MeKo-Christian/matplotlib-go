package gobasic

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/internal/sketch"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/vector"
)

// Path draws a path with the given paint style. The incoming path is in y-up
// display coordinates; it is flipped to the y-down device buffer once here, then
// the device-space pipeline runs unchanged.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	dp := r.devPath(p)
	// Sketch/xkcd perturbation runs in y-down device space (after the flip),
	// matching Matplotlib's draw_path converter chain, which folds the
	// (1,-1)+height device flip into the transform before sketching. Applying it
	// in y-up space would negate the perpendicular wiggle.
	if paint != nil {
		if eff := render.EffectiveSketch(paint.Sketch, r.defaultSketch); render.SketchActive(eff) {
			dp = sketch.Apply(dp, eff.Scale, eff.Length, eff.Randomness)
		}
	}
	r.pathDevice(dp, paint)
}

// pathDevice draws a path that is already in y-down device coordinates.
func (r *Renderer) pathDevice(p geom.Path, paint *render.Paint) {
	if paint == nil {
		return
	}
	if render.DrawPathWithEffects(r, p, paint, r.pathDevice) {
		return
	}
	if !p.Validate() {
		return // Invalid path
	}

	// Quantize path coordinates for deterministic rendering
	p = quantizePath(p)

	// Carry the full paint state forward, then quantize the numeric stroke
	// parameters in place. Reconstructing a paint from a hand-picked subset of
	// fields silently dropped Alpha/CompositeMode/Antialias/Snap/FillPattern/
	// FillGradient; copying the struct preserves them for any downstream pass
	// that consults them. Dashes are reallocated before mutation so the caller's
	// slice is never modified.
	quantized := *paint
	quantized.LineWidth = quantize(paint.LineWidth)
	quantized.MiterLimit = quantize(paint.MiterLimit)
	if len(paint.Dashes) > 0 {
		dashes := make([]float64, len(paint.Dashes))
		for i, dash := range paint.Dashes {
			dashes[i] = quantize(dash)
		}
		quantized.Dashes = dashes
	}
	quantizedPaint := &quantized

	// gobasic rasterizes solid fills and strokes only (it advertises neither
	// GradientFill nor PatternFill). Warn once if a paint reaches it carrying a
	// fill feature it cannot honor, so the omission is a signal rather than a
	// silent blank fill.
	warnUnsupportedGobasicPaint(paint)

	// Fill first if requested
	if quantizedPaint.Fill.A > 0 {
		r.fillPath(p, quantizedPaint.Fill)
	}

	// Then stroke if requested
	if quantizedPaint.Stroke.A > 0 && quantizedPaint.LineWidth > 0 {
		r.drawStroke(p, quantizedPaint)
	}
}

// DrawPathWithEffects applies renderer-neutral path effect passes.
func (r *Renderer) DrawPathWithEffects(p geom.Path, paint *render.Paint) bool {
	return render.DrawPathWithEffects(r, r.devPath(p), paint, r.pathDevice)
}

var (
	gobasicGradientWarnOnce  sync.Once
	gobasicPatternWarnOnce   sync.Once
	gobasicCompositeWarnOnce sync.Once
)

// warnUnsupportedGobasicPaint emits a one-shot diagnostic when a paint carries a
// fill feature the pure-Go renderer cannot honor. gobasic does not advertise
// GradientFill or PatternFill, and only the SourceOver composite model, so these
// are dropped during rasterization; the warning makes that visible instead of
// rendering a silently wrong (or blank) fill.
func warnUnsupportedGobasicPaint(paint *render.Paint) {
	if paint == nil {
		return
	}
	if paint.FillGradient.Kind != render.GradientNone {
		gobasicGradientWarnOnce.Do(func() {
			diag.Warnf("gobasic: gradient fills are not supported; drawing solid fill instead")
		})
	}
	if paint.FillPattern.ID != "" || len(paint.FillPattern.Path.C) > 0 {
		gobasicPatternWarnOnce.Do(func() {
			diag.Warnf("gobasic: pattern fills are not supported; drawing solid fill instead")
		})
	}
	if paint.CompositeMode != render.CompositeSourceOver {
		gobasicCompositeWarnOnce.Do(func() {
			diag.Warnf("gobasic: composite mode %d is not supported; using source-over", paint.CompositeMode)
		})
	}
}

// fillPath fills a path with the given color.
func (r *Renderer) fillPath(p geom.Path, fillColor render.Color) {
	var clipBounds image.Rectangle
	var rasterBounds image.Rectangle
	var offsetX, offsetY float64
	if r.clipRect == nil && len(r.clipPaths) == 0 {
		clipBounds = r.dst.Bounds()
		rasterBounds = clipBounds
	} else {
		pathBounds, ok := pathPixelBounds(p)
		if !ok {
			return
		}
		clipBounds = r.dst.Bounds()
		if r.clipRect != nil {
			clipBounds = clipBounds.Intersect(image.Rect(
				int(math.Floor(r.clipRect.Min.X)),
				int(math.Floor(r.clipRect.Min.Y)),
				int(math.Ceil(r.clipRect.Max.X)),
				int(math.Ceil(r.clipRect.Max.Y)),
			))
		}
		clipBounds = clipBounds.Intersect(pathBounds)
		if clipBounds.Empty() {
			return
		}
		rasterBounds = image.Rect(0, 0, clipBounds.Dx(), clipBounds.Dy())
		offsetX = float64(clipBounds.Min.X)
		offsetY = float64(clipBounds.Min.Y)
	}

	// Reset and rebuild path for filling.
	r.rasterizer.Reset(rasterBounds.Dx(), rasterBounds.Dy())

	vi := 0 // vertex index

	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			// Apply explicit rounding to ensure consistent float32 conversion
			r.rasterizer.MoveTo(float32(math.Round((pt.X-offsetX)*1e6)/1e6), float32(math.Round((pt.Y-offsetY)*1e6)/1e6))
			vi++
		case geom.LineTo:
			pt := p.V[vi]
			r.rasterizer.LineTo(float32(math.Round((pt.X-offsetX)*1e6)/1e6), float32(math.Round((pt.Y-offsetY)*1e6)/1e6))
			vi++
		case geom.QuadTo:
			ctrl := p.V[vi]
			to := p.V[vi+1]
			r.rasterizer.QuadTo(
				float32(math.Round((ctrl.X-offsetX)*1e6)/1e6), float32(math.Round((ctrl.Y-offsetY)*1e6)/1e6),
				float32(math.Round((to.X-offsetX)*1e6)/1e6), float32(math.Round((to.Y-offsetY)*1e6)/1e6),
			)
			vi += 2
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			to := p.V[vi+2]
			r.rasterizer.CubeTo(
				float32(math.Round((c1.X-offsetX)*1e6)/1e6), float32(math.Round((c1.Y-offsetY)*1e6)/1e6),
				float32(math.Round((c2.X-offsetX)*1e6)/1e6), float32(math.Round((c2.Y-offsetY)*1e6)/1e6),
				float32(math.Round((to.X-offsetX)*1e6)/1e6), float32(math.Round((to.Y-offsetY)*1e6)/1e6),
			)
			vi += 3
		case geom.ClosePath:
			r.rasterizer.ClosePath()
		}
	}

	// Draw the filled path using premultiplied alpha
	red, green, blue, alpha := fillColor.ToPremultipliedRGBA()
	c := color.RGBA{R: red, G: green, B: blue, A: alpha}

	if r.clipRect == nil && len(r.clipPaths) == 0 {
		r.rasterizer.Draw(r.dst, rasterBounds, image.NewUniform(c), image.Point{})
		return
	}

	if len(r.clipPaths) == 0 {
		// Use a zero-origin mask for the clipped path, then draw that mask directly
		// into the matching destination rectangle.
		r.rasterizer.Draw(r.dst, clipBounds, image.NewUniform(c), image.Point{})
		return
	}

	mask := image.NewAlpha(rasterBounds)
	r.rasterizer.Draw(mask, rasterBounds, image.NewUniform(color.Alpha{A: 255}), image.Point{})
	for y := clipBounds.Min.Y; y < clipBounds.Max.Y; y++ {
		for x := clipBounds.Min.X; x < clipBounds.Max.X; x++ {
			local := mask.PixOffset(x-clipBounds.Min.X, y-clipBounds.Min.Y)
			if local < 0 || local >= len(mask.Pix) {
				continue
			}
			coverage := mask.Pix[local]
			if coverage == 0 {
				continue
			}
			clipA := r.clipMaskAlphaAt(x, y)
			if clipA == 0 {
				continue
			}
			src := renderColorToRGBA(fillColor)
			src.A = uint8(uint32(src.A) * uint32(coverage) * uint32(clipA) / (255 * 255))
			if src.A == 0 {
				continue
			}
			r.blendPixelNoClip(x, y, src)
		}
	}
}

func pathPixelBounds(p geom.Path) (image.Rectangle, bool) {
	if len(p.V) == 0 {
		return image.Rectangle{}, false
	}

	minX, maxX := p.V[0].X, p.V[0].X
	minY, maxY := p.V[0].Y, p.V[0].Y
	for _, pt := range p.V[1:] {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}

	return image.Rect(
		int(math.Floor(minX))-1,
		int(math.Floor(minY))-1,
		int(math.Ceil(maxX))+1,
		int(math.Ceil(maxY))+1,
	), true
}

// drawStroke handles stroke drawing for paths using proper stroke geometry.
func (r *Renderer) drawStroke(p geom.Path, paint *render.Paint) {
	// Convert stroke to filled path with proper joins, caps, and dashes
	strokePath := strokeToPath(p, paint)
	if len(strokePath.C) == 0 {
		return // No stroke geometry generated
	}

	// Fill the stroke geometry with the stroke color
	r.fillPath(strokePath, paint.Stroke)
}

func appendPathToRasterizer(ras *vector.Rasterizer, p geom.Path, offsetX, offsetY float64) {
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			ras.MoveTo(rasterCoord(pt.X-offsetX), rasterCoord(pt.Y-offsetY))
			vi++
		case geom.LineTo:
			pt := p.V[vi]
			ras.LineTo(rasterCoord(pt.X-offsetX), rasterCoord(pt.Y-offsetY))
			vi++
		case geom.QuadTo:
			ctrl := p.V[vi]
			to := p.V[vi+1]
			ras.QuadTo(
				rasterCoord(ctrl.X-offsetX), rasterCoord(ctrl.Y-offsetY),
				rasterCoord(to.X-offsetX), rasterCoord(to.Y-offsetY),
			)
			vi += 2
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			to := p.V[vi+2]
			ras.CubeTo(
				rasterCoord(c1.X-offsetX), rasterCoord(c1.Y-offsetY),
				rasterCoord(c2.X-offsetX), rasterCoord(c2.Y-offsetY),
				rasterCoord(to.X-offsetX), rasterCoord(to.Y-offsetY),
			)
			vi += 3
		case geom.ClosePath:
			ras.ClosePath()
		}
	}
}

func rasterCoord(v float64) float32 {
	return float32(math.Round(v*1e6) / 1e6)
}

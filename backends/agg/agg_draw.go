package agg

import (
	"image"
	"math"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Path draws a path with the given paint style. The incoming path is in y-up
// display coordinates; it is flipped to the y-down device buffer once here, then
// the device-space pipeline runs unchanged.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	r.pathDevice(r.devPath(p), paint)
}

// pathDevice draws a path that is already in y-down device coordinates.
func (r *Renderer) pathDevice(p geom.Path, paint *render.Paint) {
	if render.DrawPathWithEffects(r, p, paint, r.pathDevice) {
		return
	}
	if r.hasClipPath() {
		bounds, haveBounds := pathDrawBounds(p, paint)
		r.withClipPathMask(bounds, haveBounds, func() {
			r.drawPathDirect(p, paint)
		})
		return
	}
	r.drawPathDirect(p, paint)
}

// DrawPathWithEffects applies renderer-neutral path effect passes.
func (r *Renderer) DrawPathWithEffects(p geom.Path, paint *render.Paint) bool {
	return render.DrawPathWithEffects(r, r.devPath(p), paint, r.pathDevice)
}

func (r *Renderer) drawPathDirect(p geom.Path, paint *render.Paint) {
	if !p.Validate() || paint == nil {
		return
	}
	paintCopy := *paint
	paint = &paintCopy
	applyForcedAlpha(paint)
	p, ok := r.preparePathForPaint(p, paint)
	if !ok {
		return
	}
	restoreAA := r.applyAntialiasMode(paint.Antialias)
	defer restoreAA()

	// Fill first if requested. Gradient fills take precedence over solid
	// fills; the gradient endpoint colors carry their own alpha so a missing
	// solid Fill.A is fine for gradient-only paints.
	hasGradient := paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0
	hasPattern := !hasGradient && paint.Hatch == "" && (paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0)
	if hasGradient {
		if r.applyGradientFill(paint) {
			r.buildPath(p)
			r.ctx.Fill()
			// Restore the solid fill source so downstream draws (hatch
			// overlay, stroke, marker batches, etc.) are not accidentally
			// painted through the gradient span generator.
			r.ctx.SetFillColor(renderColorToAGG(colorWithForcedAlpha(paint.Fill, paint)))
		}
	} else if hasPattern {
		r.drawPatternFill(p, paint)
	} else if paint.Fill.A > 0 {
		r.buildPath(p)
		r.ctx.SetFillColor(renderColorToAGG(colorWithForcedAlpha(paint.Fill, paint)))
		r.ctx.Fill()
	}

	if paint.Hatch != "" {
		r.drawNativeHatch(p, paint)
	}

	// Then stroke if requested
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		r.ctx.SetStrokeColor(renderColorToAGG(colorWithForcedAlpha(paint.Stroke, paint)))
		r.ctx.SetStrokeWidth(paint.LineWidth)

		// Map line join
		switch paint.LineJoin {
		case render.JoinMiter:
			r.ctx.SetLineJoin(agglib.JoinMiter)
		case render.JoinRound:
			r.ctx.SetLineJoin(agglib.JoinRound)
		case render.JoinBevel:
			r.ctx.SetLineJoin(agglib.JoinBevel)
		}

		// Map line cap
		switch paint.LineCap {
		case render.CapButt:
			r.ctx.SetLineCap(agglib.CapButt)
		case render.CapRound:
			r.ctx.SetLineCap(agglib.CapRound)
		case render.CapSquare:
			r.ctx.SetLineCap(agglib.CapSquare)
		}

		// Set miter limit
		if paint.MiterLimit > 0 {
			r.ctx.SetMiterLimit(paint.MiterLimit)
		}

		// Handle dashes
		r.ctx.ClearDashes()
		if len(paint.Dashes) >= 2 {
			r.ctx.SetDashPattern(paint.Dashes)
		}

		for _, path := range chunkStrokePath(p, paint.MaxChunkVertices) {
			r.buildPath(path)
			r.ctx.Stroke()
		}

		// Clean up dashes
		if len(paint.Dashes) >= 2 {
			r.ctx.ClearDashes()
		}
	}
}

// buildPath converts a geom.Path into AGG path commands on the current context.
func (r *Renderer) buildPath(p geom.Path) {
	r.ctx.BeginPath()

	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(p.V) {
				return
			}
			pt := p.V[vi]
			r.ctx.MoveTo(pt.X, pt.Y)
			vi++
		case geom.LineTo:
			if vi >= len(p.V) {
				return
			}
			pt := p.V[vi]
			r.ctx.LineTo(pt.X, pt.Y)
			vi++
		case geom.QuadTo:
			if vi+1 >= len(p.V) {
				return
			}
			ctrl := p.V[vi]
			to := p.V[vi+1]
			r.ctx.QuadricCurveTo(ctrl.X, ctrl.Y, to.X, to.Y)
			vi += 2
		case geom.CubicTo:
			if vi+2 >= len(p.V) {
				return
			}
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			to := p.V[vi+2]
			r.ctx.CubicCurveTo(c1.X, c1.Y, c2.X, c2.Y, to.X, to.Y)
			vi += 3
		case geom.ClosePath:
			r.ctx.ClosePath()
		}
	}
}

func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	// dst arrives in y-up display space; flip to the y-down device buffer.
	// drawImageDirect places image row 0 at dst.Min.Y (device top), keeping the
	// image upright.
	dst = r.devRect(dst)
	if r.hasClipPath() {
		bounds, haveBounds := imageDrawBounds(dst)
		r.withClipPathMaskPremultiplied(bounds, haveBounds, func() {
			r.drawImageDirect(img, dst)
		})
		return
	}
	r.drawImageDirect(img, dst)
}

// DrawBboxImage draws a Matplotlib BboxImage/OffsetImage-style image. The bbox
// is the logical display bbox in y-up coordinates; the raster image is ceiled to
// integer pixels and then placed using RendererAgg's integer draw_image rules.
func (r *Renderer) DrawBboxImage(img render.Image, bbox geom.Rect) bool {
	if img == nil || r == nil || r.ctx == nil {
		return false
	}
	dst := r.devRect(bbox)
	if r.hasClipPath() {
		bounds, haveBounds := imageDrawBounds(dst)
		r.withClipPathMaskPremultiplied(bounds, haveBounds, func() {
			r.drawBboxImageDirect(img, dst)
		})
		return true
	}
	return r.drawBboxImageDirect(img, dst)
}

func (r *Renderer) drawImageDirect(img render.Image, dst geom.Rect) {
	aggImg, ok := renderImageToAGG(img)
	if !ok {
		return
	}

	agg := r.ctx
	prevBlendMode := agg.GetBlendMode()
	prevFilter := agg.GetImageFilter()
	prevResample := agg.GetImageResample()
	agg.SetBlendMode(agglib.BlendSrcOver)
	defer func() {
		agg.SetBlendMode(prevBlendMode)
		agg.SetImageFilter(prevFilter)
		agg.SetImageResample(prevResample)
	}()

	x := dst.Min.X
	y := dst.Min.Y
	w := dst.W()
	h := dst.H()
	if w < 0 {
		x += w
		w = -w
	}
	if h < 0 {
		y += h
		h = -h
	}
	if w <= 0 || h <= 0 {
		return
	}
	if nearestImg, nx, ny, nw, nh, ok := nearestScaledImageForDirectDraw(img, x, y, w, h); ok {
		agg.SetImageFilter(agglib.NoFilter)
		agg.SetImageResample(agglib.NoResample)
		_ = agg.DrawImageScaled(nearestImg, nx, ny, nw, nh)
		return
	}
	applyInterpolation(agg, img, w, h)

	_ = agg.DrawImageScaled(aggImg, x, y, w, h)
}

func (r *Renderer) drawBboxImageDirect(img render.Image, dst geom.Rect) bool {
	x := dst.Min.X
	y := dst.Min.Y
	w := dst.W()
	h := dst.H()
	if w < 0 {
		x += w
		w = -w
	}
	if h < 0 {
		y += h
		h = -h
	}
	if w <= 0 || h <= 0 {
		return false
	}
	if nearestImg, nx, ny, nw, nh, ok := nearestScaledBboxImageForDirectDraw(img, x, y, w, h); ok {
		agg := r.ctx
		prevBlendMode := agg.GetBlendMode()
		prevFilter := agg.GetImageFilter()
		prevResample := agg.GetImageResample()
		agg.SetBlendMode(agglib.BlendSrcOver)
		agg.SetImageFilter(agglib.NoFilter)
		agg.SetImageResample(agglib.NoResample)
		defer func() {
			agg.SetBlendMode(prevBlendMode)
			agg.SetImageFilter(prevFilter)
			agg.SetImageResample(prevResample)
		}()
		_ = agg.DrawImageScaled(nearestImg, nx, ny, nw, nh)
		return true
	}
	r.drawImageDirect(img, dst)
	return true
}

func nearestScaledImageForDirectDraw(img render.Image, x, y, w, h float64) (*agglib.Image, float64, float64, float64, float64, bool) {
	if img == nil || w <= 0 || h <= 0 {
		return nil, 0, 0, 0, 0, false
	}
	srcW, srcH := img.Size()
	if srcW <= 0 || srcH <= 0 {
		return nil, 0, 0, 0, 0, false
	}
	filter, ok := resolveInterpolationName(img.Interpolation(), float64(srcW), float64(srcH), w, h)
	if !ok || filter != agglib.NoFilter {
		return nil, 0, 0, 0, 0, false
	}
	drawX := math.Round(x)
	roundedY := math.Round(y)
	dstW := int(math.Ceil(w))
	dstH := int(math.Ceil(h))
	if dstW <= 0 || dstH <= 0 {
		return nil, 0, 0, 0, 0, false
	}
	if dstW == srcW && dstH == srcH && floatAlmostEqual(drawX, x) && floatAlmostEqual(roundedY, y) {
		return nil, 0, 0, 0, 0, false
	}
	drawY := roundedY
	rgbaImage, ok := img.(render.RGBAImage)
	if !ok {
		return nil, 0, 0, 0, 0, false
	}
	scaled := scaleRGBANearest(rgbaImage.RGBA(), dstW, dstH)
	if scaled == nil {
		return nil, 0, 0, 0, 0, false
	}
	data := render.NewImageData(scaled)
	data.SetInterpolation("nearest")
	if alpha, ok := img.(render.ImageAlpha); ok {
		data.SetAlpha(alpha.Alpha())
	}
	aggImg, ok := renderImageToAGG(data)
	if !ok {
		return nil, 0, 0, 0, 0, false
	}
	return aggImg, drawX, drawY, float64(dstW), float64(dstH), true
}

func nearestScaledBboxImageForDirectDraw(img render.Image, x, y, w, h float64) (*agglib.Image, float64, float64, float64, float64, bool) {
	aggImg, _, _, nw, nh, ok := nearestScaledImageForDirectDraw(img, x, y, w, h)
	if !ok {
		return nil, 0, 0, 0, 0, false
	}
	drawX := math.Round(x)
	drawY := math.Trunc(y + h - nh)
	return aggImg, drawX, drawY, nw, nh, true
}

func scaleRGBANearest(src *image.RGBA, dstW, dstH int) *image.RGBA {
	if src == nil || dstW <= 0 || dstH <= 0 {
		return nil
	}
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return nil
	}
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	scaleX := float64(dstW) / float64(srcW)
	scaleY := float64(dstH) / float64(srcH)
	for y := 0; y < dstH; y++ {
		srcY := int(math.Round((float64(y)+0.5)/scaleY - 0.5))
		if srcY < 0 {
			srcY = 0
		}
		if srcY >= srcH {
			srcY = srcH - 1
		}
		srcRow := src.Pix[src.PixOffset(bounds.Min.X, bounds.Min.Y+srcY):]
		for x := 0; x < dstW; x++ {
			srcX := int(math.Round((float64(x)+0.5)/scaleX - 0.5))
			if srcX < 0 {
				srcX = 0
			}
			if srcX >= srcW {
				srcX = srcW - 1
			}
			copy(dst.Pix[dst.PixOffset(x, y):dst.PixOffset(x, y)+4], srcRow[srcX*4:srcX*4+4])
		}
	}
	return dst
}

// ImageTransformed draws an image using the provided affine transformation.
// Used by core.Image2D when rotation is requested.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, affine geom.Affine) {
	// affine maps image space -> y-up display. Compose the device y-flip
	// F(y -> H - y) so the final affine maps image space -> y-down device. With
	// Mul being m∘n (apply n then m), F.Mul(affine) applies affine then F, i.e.
	// deviceAffine(p) = F(affine(p)). Image rows stay upright.
	affine = r.deviceFlipAffine().Mul(affine)
	if r.hasClipPath() {
		bounds, haveBounds := transformedImageDrawBounds(img, affine)
		r.withClipPathMaskPremultiplied(bounds, haveBounds, func() {
			r.drawImageTransformedDirect(img, affine)
		})
		return
	}
	r.drawImageTransformedDirect(img, affine)
}

// deviceFlipAffine returns the affine F that maps a y-up display point to the
// y-down device buffer: (x, y) -> (x, H - y).
func (r *Renderer) deviceFlipAffine() geom.Affine {
	return geom.Affine{A: 1, B: 0, C: 0, D: -1, E: 0, F: float64(r.height)}
}

func (r *Renderer) drawImageTransformedDirect(img render.Image, affine geom.Affine) {
	aggImg, ok := renderImageToAGG(img)
	if !ok {
		return
	}

	agg := r.ctx
	prevBlendMode := agg.GetBlendMode()
	prevFilter := agg.GetImageFilter()
	prevResample := agg.GetImageResample()
	affineDispX, affineDispY := imageTransformDisplaySpan(img, affine)
	agg.SetBlendMode(agglib.BlendSrcOver)
	applyInterpolation(agg, img, affineDispX, affineDispY)
	defer func() {
		agg.SetBlendMode(prevBlendMode)
		agg.SetImageFilter(prevFilter)
		agg.SetImageResample(prevResample)
	}()

	transform := agglib.NewTransformationsFromValues(
		affine.A,
		affine.B,
		affine.C,
		affine.D,
		affine.E,
		affine.F,
	)
	_ = agg.DrawImageTransformed(aggImg, transform)
}

func extractImageAlpha(img render.Image) float64 {
	if img == nil {
		return 1
	}
	imageAlpha, ok := img.(render.ImageAlpha)
	if !ok {
		return 1
	}
	return clamp01(imageAlpha.Alpha())
}

// DrawMarkers renders one marker path at many display-space offsets.
func (r *Renderer) DrawMarkers(batch render.MarkerBatch) bool {
	if len(batch.Marker.C) == 0 || len(batch.Items) == 0 {
		return false
	}
	for i := range batch.Items {
		item := &batch.Items[i]
		markerPaint := item.Paint
		markerPaint.Snap = render.SnapAuto
		// Build the marker at its y-up display offset, then flip to device.
		path := transformMarkerPath(batch.Marker, item.Transform, item.Offset)
		if len(path.C) == 0 {
			continue
		}
		path = r.devPath(path)
		// Snapping does not commute with the y-flip, so apply pixel-centering in
		// device space against the device-space offset to stay net-neutral.
		if !shouldSnapPath(batch.Marker, &markerPaint) {
			offDev := r.devPt(item.Offset)
			snapped := geom.Pt{
				X: math.Floor(offDev.X+0.5) + 0.5,
				Y: math.Floor(offDev.Y+0.5) + 0.5,
			}
			delta := geom.Pt{X: snapped.X - offDev.X, Y: snapped.Y - offDev.Y}
			for vi := range path.V {
				path.V[vi].X += delta.X
				path.V[vi].Y += delta.Y
			}
		}
		paint := markerPaint
		if !item.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		// path is already in device space; call pathDevice to avoid double-flip.
		r.pathDevice(path, &paint)
	}
	return true
}

// DrawPathCollection renders a display-space path collection.
func (r *Renderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	if len(batch.Items) == 0 {
		return false
	}
	for i := range batch.Items {
		item := &batch.Items[i]
		if len(item.Path.C) == 0 {
			continue
		}
		paint := item.Paint
		if item.Hatch != "" {
			paint.Hatch = item.Hatch
			paint.HatchColor = item.HatchColor
			paint.HatchLineWidth = item.HatchWidth
			paint.HatchSpacing = item.HatchSpacing
		}
		if !item.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		r.Path(item.Path, &paint)
	}
	return true
}

// DrawQuadMesh renders pcolor/pcolormesh-style quadrilateral cells.
func (r *Renderer) DrawQuadMesh(batch render.QuadMeshBatch) bool {
	if len(batch.Cells) == 0 {
		return false
	}
	for i := range batch.Cells {
		cell := &batch.Cells[i]
		path := geom.Path{}
		path.MoveTo(cell.Quad[0])
		path.LineTo(cell.Quad[1])
		path.LineTo(cell.Quad[2])
		path.LineTo(cell.Quad[3])
		path.Close()
		paint := render.Paint{
			Fill:         cell.Face,
			Stroke:       cell.Edge,
			LineWidth:    cell.LineWidth,
			LineJoin:     render.JoinMiter,
			LineCap:      render.CapButt,
			Dashes:       append([]float64(nil), cell.Dashes...),
			Hatch:        cell.Hatch,
			HatchColor:   cell.HatchColor,
			HatchSpacing: cell.HatchSpacing,
			Antialias:    render.AntialiasDefault,
			Snap:         cell.Snap,
		}
		if cell.HatchWidth > 0 {
			paint.HatchLineWidth = cell.HatchWidth
		}
		if !cell.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		if paint.LineWidth <= 0 || paint.Stroke.A <= 0 {
			paint.Stroke = render.Color{}
			paint.LineWidth = 0
		}
		if paint.Fill.A <= 0 {
			paint.Fill = render.Color{}
		}
		r.Path(path, &paint)
	}
	return true
}

// DrawGouraudTriangles renders interpolated-color triangles directly into the
// AGG surface buffer.
func (r *Renderer) DrawGouraudTriangles(batch render.GouraudTriangleBatch) bool {
	if len(batch.Triangles) == 0 || r.ctx == nil || r.ctx.image == nil {
		return false
	}
	// drawGouraudTriangle writes directly into the device buffer, so flip each
	// triangle's vertices to device space first. The clip-mask bounds are then
	// derived from the flipped triangles to match.
	devBatch := render.GouraudTriangleBatch{
		Triangles: make([]render.GouraudTriangle, len(batch.Triangles)),
	}
	for i := range batch.Triangles {
		tri := batch.Triangles[i]
		for j := range tri.P {
			tri.P[j] = r.devPt(tri.P[j])
		}
		devBatch.Triangles[i] = tri
	}
	draw := func() {
		for i := range devBatch.Triangles {
			r.drawGouraudTriangle(&devBatch.Triangles[i])
		}
	}
	if r.hasClipPath() {
		bounds, haveBounds := gouraudTriangleBatchBounds(devBatch)
		r.withClipPathMask(bounds, haveBounds, draw)
	} else {
		draw()
	}
	return true
}

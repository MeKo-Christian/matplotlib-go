package agg

import (
	"math"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Path draws a path with the given paint style.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if render.DrawPathWithEffects(r, p, paint, r.Path) {
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
	return render.DrawPathWithEffects(r, p, paint, r.Path)
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
	if r.hasClipPath() {
		bounds, haveBounds := imageDrawBounds(dst)
		r.withClipPathMask(bounds, haveBounds, func() {
			r.drawImageDirect(img, dst)
		})
		return
	}
	r.drawImageDirect(img, dst)
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
	applyInterpolation(agg, img, w, h)

	_ = agg.DrawImageScaled(aggImg, x, y, w, h)
}

// ImageTransformed draws an image using the provided affine transformation.
// Used by core.Image2D when rotation is requested.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, affine geom.Affine) {
	if r.hasClipPath() {
		bounds, haveBounds := transformedImageDrawBounds(img, affine)
		r.withClipPathMask(bounds, haveBounds, func() {
			r.drawImageTransformedDirect(img, affine)
		})
		return
	}
	r.drawImageTransformedDirect(img, affine)
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
		offset := item.Offset
		markerPaint := item.Paint
		markerPaint.Snap = render.SnapAuto
		if !shouldSnapPath(batch.Marker, &markerPaint) {
			offset.X = math.Floor(offset.X+0.5) + 0.5
			offset.Y = math.Floor(offset.Y+0.5) + 0.5
		}
		path := transformMarkerPath(batch.Marker, item.Transform, offset)
		if len(path.C) == 0 {
			continue
		}
		paint := markerPaint
		if !item.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		r.Path(path, &paint)
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
			Snap:         render.SnapOn,
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
	draw := func() {
		for i := range batch.Triangles {
			r.drawGouraudTriangle(&batch.Triangles[i])
		}
	}
	if r.hasClipPath() {
		bounds, haveBounds := gouraudTriangleBatchBounds(batch)
		r.withClipPathMask(bounds, haveBounds, draw)
	} else {
		draw()
	}
	return true
}

package gobasic

import (
	"image"
	"image/color"
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// DrawImage draws an image within the destination rectangle.
func (r *Renderer) DrawImage(img render.Image, dst geom.Rect) {
	if img == nil {
		return
	}

	src := asRGBAImage(img)
	if src == nil {
		return
	}

	// dst arrives in y-up display space; flip to the y-down device buffer.
	// drawBitmapScaledWithAlpha places image row 0 at the device-top row
	// (dst.Min.Y after the flip), keeping the image upright.
	dst = r.devRect(dst)

	// Destination rectangle in integer coordinates.
	minX := int(math.Floor(dst.Min.X))
	minY := int(math.Floor(dst.Min.Y))
	maxX := int(math.Ceil(dst.Max.X))
	maxY := int(math.Ceil(dst.Max.Y))
	if maxX <= minX || maxY <= minY {
		return
	}

	r.drawBitmapScaledWithAlpha(src, minX, minY, maxX-minX, maxY-minY, imageAlphaMultiplier(img))
}

// ImageTransformed draws a raster image through an arbitrary affine transform.
// The affine maps source image pixels into display coordinates.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, transform geom.Affine) {
	if img == nil {
		return
	}
	src := asRGBAImage(img)
	if src == nil {
		return
	}
	// transform maps image space -> y-up display. Compose the device y-flip so
	// the final affine maps image space -> y-down device. Mul is m∘n (apply n
	// then m), so deviceFlipAffine().Mul(transform)(p) = F(transform(p)); image
	// rows stay upright. This matches core/image.go's imageTransform convention.
	transform = r.deviceFlipAffine().Mul(transform)
	r.drawBitmapTransformed(src, transform, imageAlphaMultiplier(img))
}

func (r *Renderer) drawBitmapScaled(src *image.RGBA, dstX, dstY, dstW, dstH int) {
	r.drawBitmapScaledWithAlpha(src, dstX, dstY, dstW, dstH, 1)
}

func (r *Renderer) drawBitmapScaledWithAlpha(src *image.RGBA, dstX, dstY, dstW, dstH int, alpha float64) {
	if src == nil || dstW <= 0 || dstH <= 0 {
		return
	}
	if alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}

	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	if srcW <= 0 || srcH <= 0 {
		return
	}

	dst, ok := r.drawTargetRect(dstX, dstY, dstX+dstW, dstY+dstH)
	if !ok {
		return
	}

	srcMin := src.Bounds().Min
	for y := dst.Min.Y; y < dst.Max.Y; y++ {
		sy := nearestScaledSourceIndex(y, dstY, dstH, srcH)

		srcIdxBase := src.PixOffset(srcMin.X, srcMin.Y+sy)
		srcRow := src.Pix[srcIdxBase : srcIdxBase+srcW*4]
		for x := dst.Min.X; x < dst.Max.X; x++ {
			sx := nearestScaledSourceIndex(x, dstX, dstW, srcW)

			srcOffset := sx * 4
			srcColor := color.RGBA{
				R: srcRow[srcOffset],
				G: srcRow[srcOffset+1],
				B: srcRow[srcOffset+2],
				A: srcRow[srcOffset+3],
			}
			if alpha < 1 {
				srcColor.A = uint8(math.Round(float64(srcColor.A) * alpha))
			}
			r.blendPixel(x, y, srcColor)
		}
	}
}

func nearestScaledSourceIndex(dstIndex, dstOrigin, dstSize, srcSize int) int {
	if dstSize <= 0 || srcSize <= 0 {
		return 0
	}
	rel := dstIndex - dstOrigin
	idx := int(math.Round((float64(rel)+0.5)*float64(srcSize)/float64(dstSize) - 0.5))
	if idx < 0 {
		return 0
	}
	if idx >= srcSize {
		return srcSize - 1
	}
	return idx
}

func (r *Renderer) drawBitmapTransformed(src *image.RGBA, transform geom.Affine, alpha float64) {
	if src == nil || alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}

	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	if srcW <= 0 || srcH <= 0 {
		return
	}
	inv, ok := transform.Invert()
	if !ok {
		return
	}

	corners := []geom.Pt{
		transform.Apply(geom.Pt{X: 0, Y: 0}),
		transform.Apply(geom.Pt{X: float64(srcW), Y: 0}),
		transform.Apply(geom.Pt{X: float64(srcW), Y: float64(srcH)}),
		transform.Apply(geom.Pt{X: 0, Y: float64(srcH)}),
	}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, corner := range corners {
		if corner.X < minX {
			minX = corner.X
		}
		if corner.X > maxX {
			maxX = corner.X
		}
		if corner.Y < minY {
			minY = corner.Y
		}
		if corner.Y > maxY {
			maxY = corner.Y
		}
	}
	if math.IsNaN(minX) || math.IsNaN(minY) || math.IsNaN(maxX) || math.IsNaN(maxY) ||
		math.IsInf(minX, 0) || math.IsInf(minY, 0) || math.IsInf(maxX, 0) || math.IsInf(maxY, 0) {
		return
	}

	dst, ok := r.drawTargetRect(
		int(math.Floor(minX)),
		int(math.Floor(minY)),
		int(math.Ceil(maxX)),
		int(math.Ceil(maxY)),
	)
	if !ok {
		return
	}

	srcMin := src.Bounds().Min
	for y := dst.Min.Y; y < dst.Max.Y; y++ {
		for x := dst.Min.X; x < dst.Max.X; x++ {
			sp := inv.Apply(geom.Pt{X: float64(x) + 0.5, Y: float64(y) + 0.5})
			sx := int(math.Floor(sp.X))
			sy := int(math.Floor(sp.Y))
			if sx < 0 || sy < 0 || sx >= srcW || sy >= srcH {
				continue
			}
			p := src.PixOffset(srcMin.X+sx, srcMin.Y+sy)
			srcColor := color.RGBA{
				R: src.Pix[p],
				G: src.Pix[p+1],
				B: src.Pix[p+2],
				A: src.Pix[p+3],
			}
			if alpha < 1 {
				srcColor.A = uint8(math.Round(float64(srcColor.A) * alpha))
			}
			r.blendPixel(x, y, srcColor)
		}
	}
}

func (r *Renderer) drawBitmapRotated(src *image.RGBA, anchor, pivot geom.Pt, angle float64) {
	if src == nil {
		return
	}

	srcW := float64(src.Bounds().Dx())
	srcH := float64(src.Bounds().Dy())
	if srcW <= 0 || srcH <= 0 {
		return
	}

	cos, sin := math.Cos(angle), math.Sin(angle)

	corners := [4]struct{ x, y float64 }{
		{-pivot.X, -pivot.Y},
		{srcW - pivot.X, -pivot.Y},
		{srcW - pivot.X, srcH - pivot.Y},
		{-pivot.X, srcH - pivot.Y},
	}

	minX := math.Inf(1)
	maxX := math.Inf(-1)
	minY := math.Inf(1)
	maxY := math.Inf(-1)
	for _, corner := range corners {
		rx := corner.x*cos - corner.y*sin
		ry := corner.x*sin + corner.y*cos
		if rx < minX {
			minX = rx
		}
		if rx > maxX {
			maxX = rx
		}
		if ry < minY {
			minY = ry
		}
		if ry > maxY {
			maxY = ry
		}
	}

	boundsW := int(math.Ceil(maxX - minX))
	boundsH := int(math.Ceil(maxY - minY))
	if boundsW <= 0 || boundsH <= 0 {
		return
	}

	minXInt := int(math.Floor(anchor.X + minX))
	minYInt := int(math.Floor(anchor.Y + minY))
	drawBounds, ok := r.drawTargetRect(minXInt, minYInt, minXInt+boundsW, minYInt+boundsH)
	if !ok {
		return
	}

	srcMin := src.Bounds().Min
	for y := drawBounds.Min.Y; y < drawBounds.Max.Y; y++ {
		for x := drawBounds.Min.X; x < drawBounds.Max.X; x++ {
			localX := float64(x) - anchor.X
			localY := float64(y) - anchor.Y

			// Inverse rotation from destination to source coordinates.
			sxF := localX*cos + localY*sin
			syF := -localX*sin + localY*cos

			srcX := int(math.Round(sxF + pivot.X - 0.5))
			srcY := int(math.Round(syF + pivot.Y - 0.5))
			if srcX < 0 || srcY < 0 || srcX >= int(srcW) || srcY >= int(srcH) {
				continue
			}

			p := src.PixOffset(srcMin.X+srcX, srcMin.Y+srcY)
			r.blendPixel(x, y, color.RGBA{
				R: src.Pix[p],
				G: src.Pix[p+1],
				B: src.Pix[p+2],
				A: src.Pix[p+3],
			})
		}
	}
}

func asRGBAImage(img render.Image) *image.RGBA {
	rgbaImage, ok := img.(interface {
		RGBA() *image.RGBA
	})
	if ok {
		return rgbaImage.RGBA()
	}
	return nil
}

func imageAlphaMultiplier(img render.Image) float64 {
	alphaImage, ok := img.(render.ImageAlpha)
	if !ok {
		return 1
	}
	alpha := alphaImage.Alpha()
	if alpha <= 0 {
		return 0
	}
	if alpha >= 1 {
		return 1
	}
	return alpha
}

func scaleImageNearest(src *image.RGBA, scaleX, scaleY float64) *image.RGBA {
	if src == nil || scaleX <= 0 || scaleY <= 0 {
		return nil
	}

	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	dstW := int(math.Ceil(float64(srcW) * scaleX))
	dstH := int(math.Ceil(float64(srcH) * scaleY))
	if dstW <= 0 || dstH <= 0 {
		return nil
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		srcY := int(math.Round((float64(y)+0.5)/scaleY - 0.5))
		if srcY < 0 {
			srcY = 0
		}
		if srcY >= srcH {
			srcY = srcH - 1
		}
		srcRow := src.Pix[src.PixOffset(src.Bounds().Min.X, src.Bounds().Min.Y+srcY):]
		for x := 0; x < dstW; x++ {
			srcX := int(math.Round((float64(x)+0.5)/scaleX - 0.5))
			if srcX < 0 {
				srcX = 0
			}
			if srcX >= srcW {
				srcX = srcW - 1
			}

			srcOffset := srcRow[srcX*4:]
			dstOffset := dst.PixOffset(x, y)
			dst.Pix[dstOffset] = srcOffset[0]
			dst.Pix[dstOffset+1] = srcOffset[1]
			dst.Pix[dstOffset+2] = srcOffset[2]
			dst.Pix[dstOffset+3] = srcOffset[3]
		}
	}

	return dst
}

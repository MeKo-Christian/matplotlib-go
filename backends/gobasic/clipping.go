package gobasic

import (
	"image"
	"image/color"
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"golang.org/x/image/vector"
)

func (r *Renderer) drawTargetRect(minX, minY, maxX, maxY int) (image.Rectangle, bool) {
	if minX >= maxX || minY >= maxY {
		return image.Rectangle{}, false
	}

	bounds := image.Rect(minX, minY, maxX, maxY).Intersect(r.dst.Bounds())
	if bounds.Empty() {
		return image.Rectangle{}, false
	}

	if r.clipRect != nil {
		clipBounds := image.Rect(
			int(math.Floor(r.clipRect.Min.X)),
			int(math.Floor(r.clipRect.Min.Y)),
			int(math.Ceil(r.clipRect.Max.X)),
			int(math.Ceil(r.clipRect.Max.Y)),
		)
		bounds = bounds.Intersect(clipBounds)
	}

	if bounds.Empty() {
		return image.Rectangle{}, false
	}

	return bounds, true
}

func (r *Renderer) clipMaskAlphaAt(x, y int) uint8 {
	if len(r.clipPaths) == 0 {
		return 255
	}
	if x < 0 || y < 0 || x >= r.dst.Bounds().Dx() || y >= r.dst.Bounds().Dy() {
		return 0
	}
	alpha := 255
	for _, path := range r.clipPaths {
		mask := r.clipMaskForPath(path)
		if mask == nil {
			return 0
		}
		i := mask.PixOffset(x, y)
		if i < 0 || i >= len(mask.Pix) {
			return 0
		}
		alpha = alpha * int(mask.Pix[i]) / 255
		if alpha == 0 {
			return 0
		}
	}
	return uint8(alpha)
}

func (r *Renderer) clipMaskForPath(path geom.Path) *image.Alpha {
	if len(path.C) == 0 || !path.Validate() || r.dst == nil {
		return nil
	}
	b := r.dst.Bounds()
	key := clipMaskKey{
		width:  b.Dx(),
		height: b.Dy(),
		hash:   hashPath(path),
	}
	if r.clipMaskMap == nil {
		r.clipMaskMap = make(map[clipMaskKey]*image.Alpha)
	}
	if mask, ok := r.clipMaskMap[key]; ok {
		return mask
	}

	mask := image.NewAlpha(b)
	ras := vector.NewRasterizer(b.Dx(), b.Dy())
	appendPathToRasterizer(ras, path, 0, 0)
	ras.Draw(mask, b, image.NewUniform(color.Alpha{A: 255}), image.Point{})
	r.clipMaskMap[key] = mask
	return mask
}

package agg

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func finitePt(pt geom.Pt) bool {
	return !math.IsNaN(pt.X) && !math.IsInf(pt.X, 0) && !math.IsNaN(pt.Y) && !math.IsInf(pt.Y, 0)
}

func (r *Renderer) pathOutsideVisibleArea(path geom.Path, paint *render.Paint) bool {
	bounds, ok := pathBounds(path)
	if !ok {
		return true
	}
	visible := geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: float64(r.width), Y: float64(r.height)}}
	if r.viewport != (geom.Rect{}) {
		visible = visible.Intersect(r.viewport)
	}
	if r.clipRect != nil {
		visible = visible.Intersect(*r.clipRect)
	}
	if visible.W() <= 0 || visible.H() <= 0 {
		return true
	}

	pad := 1.0
	if paint != nil && paint.Stroke.A > 0 && paint.LineWidth > 0 {
		pad += paint.LineWidth / 2
	}
	return !rectsOverlap(bounds.Inflate(pad, pad), visible)
}

func pathBounds(path geom.Path) (geom.Rect, bool) {
	if len(path.V) == 0 {
		return geom.Rect{}, false
	}
	var bounds geom.Rect
	ok := false
	for _, pt := range path.V {
		if !finitePt(pt) {
			continue
		}
		if !ok {
			bounds = geom.Rect{Min: pt, Max: pt}
			ok = true
			continue
		}
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}
	return bounds, ok
}

func pathDrawBounds(path geom.Path, paint *render.Paint) (geom.Rect, bool) {
	if paint == nil {
		return geom.Rect{}, false
	}
	bounds, ok := pathBounds(path)
	if !ok {
		return geom.Rect{}, false
	}
	pad := 2.0
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		pad += paint.LineWidth / 2
	}
	if paint.Hatch != "" && paint.HatchLineWidth > 0 {
		pad = math.Max(pad, paint.HatchLineWidth/2+2)
	}
	return bounds.Inflate(pad, pad), true
}

func imageDrawBounds(dst geom.Rect) (geom.Rect, bool) {
	minX := math.Min(dst.Min.X, dst.Max.X)
	minY := math.Min(dst.Min.Y, dst.Max.Y)
	maxX := math.Max(dst.Min.X, dst.Max.X)
	maxY := math.Max(dst.Min.Y, dst.Max.Y)
	if minX == maxX || minY == maxY {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}.Inflate(2, 2), true
}

func transformedImageDrawBounds(img render.Image, affine geom.Affine) (geom.Rect, bool) {
	if img == nil {
		return geom.Rect{}, false
	}
	w, h := img.Size()
	if w <= 0 || h <= 0 {
		return geom.Rect{}, false
	}
	return pointsBounds([]geom.Pt{
		affine.Apply(geom.Pt{}),
		affine.Apply(geom.Pt{X: float64(w)}),
		affine.Apply(geom.Pt{Y: float64(h)}),
		affine.Apply(geom.Pt{X: float64(w), Y: float64(h)}),
	}, 2)
}

func imageTransformDisplaySpan(img render.Image, affine geom.Affine) (float64, float64) {
	w, h := img.Size()
	if w <= 0 || h <= 0 {
		return 0, 0
	}

	bounds, ok := pointsBounds([]geom.Pt{
		affine.Apply(geom.Pt{}),
		affine.Apply(geom.Pt{X: float64(w)}),
		affine.Apply(geom.Pt{Y: float64(h)}),
		affine.Apply(geom.Pt{X: float64(w), Y: float64(h)}),
	}, 0)
	if !ok {
		return 0, 0
	}
	return bounds.W(), bounds.H()
}

func gouraudTriangleBatchBounds(batch render.GouraudTriangleBatch) (geom.Rect, bool) {
	points := make([]geom.Pt, 0, len(batch.Triangles)*3)
	for i := range batch.Triangles {
		points = append(points, batch.Triangles[i].P[:]...)
	}
	return pointsBounds(points, 1)
}

func pointsBounds(points []geom.Pt, pad float64) (geom.Rect, bool) {
	var bounds geom.Rect
	ok := false
	for _, pt := range points {
		if !finitePt(pt) {
			continue
		}
		if !ok {
			bounds = geom.Rect{Min: pt, Max: pt}
			ok = true
			continue
		}
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}
	if !ok {
		return geom.Rect{}, false
	}
	return bounds.Inflate(pad, pad), true
}

func rectsOverlap(a, b geom.Rect) bool {
	return a.Max.X >= b.Min.X && b.Max.X >= a.Min.X && a.Max.Y >= b.Min.Y && b.Max.Y >= a.Min.Y
}

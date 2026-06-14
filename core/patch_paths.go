package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

func linspace(start, end float64, count int) []float64 {
	if count <= 0 {
		return nil
	}
	if count == 1 {
		return []float64{start}
	}
	step := (end - start) / float64(count-1)
	out := make([]float64, count)
	for i := range out {
		out[i] = start + float64(i)*step
	}
	return out
}

func ellipseBezierPath(center geom.Pt, rx, ry float64) geom.Path {
	if rx <= 0 || ry <= 0 {
		return geom.Path{}
	}
	kx := rx * patchArcKappa
	ky := ry * patchArcKappa
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: center.X + rx, Y: center.Y})
	path.CubicTo(
		geom.Pt{X: center.X + rx, Y: center.Y + ky},
		geom.Pt{X: center.X + kx, Y: center.Y + ry},
		geom.Pt{X: center.X, Y: center.Y + ry},
	)
	path.CubicTo(
		geom.Pt{X: center.X - kx, Y: center.Y + ry},
		geom.Pt{X: center.X - rx, Y: center.Y + ky},
		geom.Pt{X: center.X - rx, Y: center.Y},
	)
	path.CubicTo(
		geom.Pt{X: center.X - rx, Y: center.Y - ky},
		geom.Pt{X: center.X - kx, Y: center.Y - ry},
		geom.Pt{X: center.X, Y: center.Y - ry},
	)
	path.CubicTo(
		geom.Pt{X: center.X + kx, Y: center.Y - ry},
		geom.Pt{X: center.X + rx, Y: center.Y - ky},
		geom.Pt{X: center.X + rx, Y: center.Y},
	)
	path.Close()
	return path
}

func approxPtCore(a, b geom.Pt, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol && math.Abs(a.Y-b.Y) <= tol
}

func buildDisplayPath(ctx *DrawContext, coords CoordinateSpec, local geom.Path, localToCoords geom.Affine) geom.Path {
	return buildArtistDisplayPath(ctx, nil, coords, local, localToCoords)
}

func buildArtistDisplayPath(ctx *DrawContext, art any, fallback CoordinateSpec, local geom.Path, localToCoords geom.Affine) geom.Path {
	path := applyAffinePath(local, localToCoords)
	if ctx == nil {
		return path
	}
	tr := artistTransformFor(ctx, art, fallback)
	if tr == nil {
		return path
	}
	return applyTransformPath(path, tr)
}

func isDataCoords(spec CoordinateSpec) bool {
	return spec.X == CoordData && spec.Y == CoordData
}

func patchAffine(origin geom.Pt, angleDeg float64) geom.Affine {
	rad := angleDeg * math.Pi / 180
	cosA := math.Cos(rad)
	sinA := math.Sin(rad)
	return geom.Affine{
		A: cosA,
		B: sinA,
		C: -sinA,
		D: cosA,
		E: origin.X,
		F: origin.Y,
	}
}

func translateAffine(offset geom.Pt) geom.Affine {
	return geom.Affine{A: 1, D: 1, E: offset.X, F: offset.Y}
}

func patchRectPath(rect geom.Rect) geom.Path {
	rect = normalizeRect(rect)
	return polygonPath([]geom.Pt{
		rect.Min,
		{X: rect.Max.X, Y: rect.Min.Y},
		rect.Max,
		{X: rect.Min.X, Y: rect.Max.Y},
	}, true)
}

func rectanglePath(width, height float64) geom.Path {
	return patchRectPath(geom.Rect{
		Min: geom.Pt{X: math.Min(0, width), Y: math.Min(0, height)},
		Max: geom.Pt{X: math.Max(0, width), Y: math.Max(0, height)},
	})
}

func roundedRectPath(rect geom.Rect, radius float64) geom.Path {
	rect = normalizeRect(rect)
	if rect.W() == 0 || rect.H() == 0 {
		return geom.Path{}
	}

	maxRadius := math.Min(rect.W(), rect.H()) / 2
	if radius <= 0 {
		return patchRectPath(rect)
	}
	if radius > maxRadius {
		radius = maxRadius
	}

	left, top := rect.Min.X, rect.Min.Y
	right, bottom := rect.Max.X, rect.Max.Y
	k := radius * patchArcKappa

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: left + radius, Y: top})
	path.LineTo(geom.Pt{X: right - radius, Y: top})
	path.CubicTo(
		geom.Pt{X: right - radius + k, Y: top},
		geom.Pt{X: right, Y: top + radius - k},
		geom.Pt{X: right, Y: top + radius},
	)
	path.LineTo(geom.Pt{X: right, Y: bottom - radius})
	path.CubicTo(
		geom.Pt{X: right, Y: bottom - radius + k},
		geom.Pt{X: right - radius + k, Y: bottom},
		geom.Pt{X: right - radius, Y: bottom},
	)
	path.LineTo(geom.Pt{X: left + radius, Y: bottom})
	path.CubicTo(
		geom.Pt{X: left + radius - k, Y: bottom},
		geom.Pt{X: left, Y: bottom - radius + k},
		geom.Pt{X: left, Y: bottom - radius},
	)
	path.LineTo(geom.Pt{X: left, Y: top + radius})
	path.CubicTo(
		geom.Pt{X: left, Y: top + radius - k},
		geom.Pt{X: left + radius - k, Y: top},
		geom.Pt{X: left + radius, Y: top},
	)
	path.Close()
	return path
}

func matplotlibRoundBoxPath(rect geom.Rect, radius float64) geom.Path {
	rect = normalizeRect(rect)
	if rect.W() == 0 || rect.H() == 0 {
		return geom.Path{}
	}
	maxRadius := math.Min(rect.W(), rect.H()) / 2
	if radius <= 0 {
		return patchRectPath(rect)
	}
	if radius > maxRadius {
		radius = maxRadius
	}

	left, bottom := rect.Min.X, rect.Min.Y
	right, top := rect.Max.X, rect.Max.Y
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: left + radius, Y: bottom})
	path.LineTo(geom.Pt{X: right - radius, Y: bottom})
	path.QuadTo(geom.Pt{X: right, Y: bottom}, geom.Pt{X: right, Y: bottom + radius})
	path.LineTo(geom.Pt{X: right, Y: top - radius})
	path.QuadTo(geom.Pt{X: right, Y: top}, geom.Pt{X: right - radius, Y: top})
	path.LineTo(geom.Pt{X: left + radius, Y: top})
	path.QuadTo(geom.Pt{X: left, Y: top}, geom.Pt{X: left, Y: top - radius})
	path.LineTo(geom.Pt{X: left, Y: bottom + radius})
	path.QuadTo(geom.Pt{X: left, Y: bottom}, geom.Pt{X: left + radius, Y: bottom})
	path.Close()
	return path
}

func ellipsePath(width, height float64) geom.Path {
	rx := math.Abs(width) / 2
	ry := math.Abs(height) / 2
	if rx == 0 || ry == 0 {
		return geom.Path{}
	}

	points := make([]geom.Pt, 0, patchCircleSegments)
	for i := 0; i < patchCircleSegments; i++ {
		angle := 2 * math.Pi * float64(i) / patchCircleSegments
		points = append(points, geom.Pt{
			X: rx * math.Cos(angle),
			Y: ry * math.Sin(angle),
		})
	}
	return polygonPath(points, true)
}

func polygonPath(points []geom.Pt, close bool) geom.Path {
	if len(points) == 0 {
		return geom.Path{}
	}
	path := geom.Path{}
	for i, pt := range points {
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	if close {
		path.Close()
	}
	return path
}

func applyAffinePath(path geom.Path, affine geom.Affine) geom.Path {
	if len(path.C) == 0 {
		return geom.Path{}
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

func applyTransformPath(path geom.Path, tr transform.T) geom.Path {
	if len(path.C) == 0 || tr == nil {
		return path
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, pt := range path.V {
		out.V[i] = tr.Apply(pt)
	}
	return out
}

func pathBounds(path geom.Path) (geom.Rect, bool) {
	if len(path.V) == 0 {
		return geom.Rect{}, false
	}
	bounds := geom.Rect{Min: path.V[0], Max: path.V[0]}
	for _, pt := range path.V[1:] {
		bounds = expandRect(bounds, pt)
	}
	return bounds, true
}

func normalizeRect(rect geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: math.Min(rect.Min.X, rect.Max.X), Y: math.Min(rect.Min.Y, rect.Max.Y)},
		Max: geom.Pt{X: math.Max(rect.Min.X, rect.Max.X), Y: math.Max(rect.Min.Y, rect.Max.Y)},
	}
}

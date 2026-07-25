package widgets

import (
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
)

const widgetCircleSegments = 48

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func boolPtr(value bool) *bool {
	return &value
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func clampFloat(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func resolvedFontSize(size float64, ctx *core.DrawContext) float64 {
	if size > 0 {
		return size
	}
	if ctx != nil && ctx.RC.FontSize > 0 {
		return ctx.RC.FontSize
	}
	return 12
}

func pointsToPixels(rc *style.RC, points float64) float64 {
	dpi := rc.DPI
	if dpi <= 0 {
		dpi = style.CurrentDefaults().DPI
		if dpi <= 0 {
			dpi = 96
		}
	}
	return points * dpi / 72.0
}

func rectCenter(rect geom.Rect) geom.Pt {
	return geom.Pt{
		X: rect.Min.X + rect.W()/2,
		Y: rect.Min.Y + rect.H()/2,
	}
}

func pixelLinePath(p1, p2 geom.Pt) geom.Path {
	return geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{p1, p2},
	}
}

func roundedRectPath(rect geom.Rect, radius float64) geom.Path {
	rect = normalizeRect(rect)
	if rect.W() == 0 || rect.H() == 0 {
		return geom.Path{}
	}
	if radius <= 0 {
		return rectanglePath(rect)
	}
	radius = math.Min(radius, math.Min(rect.W(), rect.H())/2)

	left, bottom := rect.Min.X, rect.Min.Y
	right, top := rect.Max.X, rect.Max.Y
	k := radius * geom.BezierCircleKappa

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: left + radius, Y: bottom})
	path.LineTo(geom.Pt{X: right - radius, Y: bottom})
	path.CubicTo(
		geom.Pt{X: right - radius + k, Y: bottom},
		geom.Pt{X: right, Y: bottom + radius - k},
		geom.Pt{X: right, Y: bottom + radius},
	)
	path.LineTo(geom.Pt{X: right, Y: top - radius})
	path.CubicTo(
		geom.Pt{X: right, Y: top - radius + k},
		geom.Pt{X: right - radius + k, Y: top},
		geom.Pt{X: right - radius, Y: top},
	)
	path.LineTo(geom.Pt{X: left + radius, Y: top})
	path.CubicTo(
		geom.Pt{X: left + radius - k, Y: top},
		geom.Pt{X: left, Y: top - radius + k},
		geom.Pt{X: left, Y: top - radius},
	)
	path.LineTo(geom.Pt{X: left, Y: bottom + radius})
	path.CubicTo(
		geom.Pt{X: left, Y: bottom + radius - k},
		geom.Pt{X: left + radius - k, Y: bottom},
		geom.Pt{X: left + radius, Y: bottom},
	)
	path.Close()
	return path
}

func rectanglePath(rect geom.Rect) geom.Path {
	rect = normalizeRect(rect)
	path := geom.Path{}
	path.MoveTo(rect.Min)
	path.LineTo(geom.Pt{X: rect.Max.X, Y: rect.Min.Y})
	path.LineTo(rect.Max)
	path.LineTo(geom.Pt{X: rect.Min.X, Y: rect.Max.Y})
	path.Close()
	return path
}

func normalizeRect(rect geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: math.Min(rect.Min.X, rect.Max.X), Y: math.Min(rect.Min.Y, rect.Max.Y)},
		Max: geom.Pt{X: math.Max(rect.Min.X, rect.Max.X), Y: math.Max(rect.Min.Y, rect.Max.Y)},
	}
}

func ellipsePath(width, height float64) geom.Path {
	rx := math.Abs(width) / 2
	ry := math.Abs(height) / 2
	if rx == 0 || ry == 0 {
		return geom.Path{}
	}
	path := geom.Path{}
	for i := 0; i < widgetCircleSegments; i++ {
		angle := 2 * math.Pi * float64(i) / widgetCircleSegments
		point := geom.Pt{X: rx * math.Cos(angle), Y: ry * math.Sin(angle)}
		if i == 0 {
			path.MoveTo(point)
		} else {
			path.LineTo(point)
		}
	}
	path.Close()
	return path
}

func patchAffine(origin geom.Pt, angleDegrees float64) geom.Affine {
	radians := angleDegrees * math.Pi / 180
	cosAngle := math.Cos(radians)
	sinAngle := math.Sin(radians)
	return geom.Affine{
		A: cosAngle,
		B: sinAngle,
		C: -sinAngle,
		D: cosAngle,
		E: origin.X,
		F: origin.Y,
	}
}

func translateAffine(offset geom.Pt) geom.Affine {
	return geom.Affine{A: 1, D: 1, E: offset.X, F: offset.Y}
}

func applyAffinePath(path geom.Path, affine geom.Affine) geom.Path {
	if len(path.C) == 0 {
		return geom.Path{}
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, point := range path.V {
		out.V[i] = affine.Apply(point)
	}
	return out
}

func pathBounds(path geom.Path) (geom.Rect, bool) {
	if len(path.V) == 0 {
		return geom.Rect{}, false
	}
	bounds := geom.Rect{Min: path.V[0], Max: path.V[0]}
	for _, point := range path.V[1:] {
		bounds.Min.X = math.Min(bounds.Min.X, point.X)
		bounds.Min.Y = math.Min(bounds.Min.Y, point.Y)
		bounds.Max.X = math.Max(bounds.Max.X, point.X)
		bounds.Max.Y = math.Max(bounds.Max.Y, point.Y)
	}
	return bounds, true
}

func unionRect(a, b geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: math.Min(a.Min.X, b.Min.X), Y: math.Min(a.Min.Y, b.Min.Y)},
		Max: geom.Pt{X: math.Max(a.Max.X, b.Max.X), Y: math.Max(a.Max.Y, b.Max.Y)},
	}
}

func distancePointToSegment(a, b, point geom.Pt) float64 {
	delta := geom.Pt{X: b.X - a.X, Y: b.Y - a.Y}
	lengthSquared := delta.X*delta.X + delta.Y*delta.Y
	if lengthSquared == 0 {
		return math.Hypot(point.X-a.X, point.Y-a.Y)
	}
	t := ((point.X-a.X)*delta.X + (point.Y-a.Y)*delta.Y) / lengthSquared
	t = clampFloat(t, 0, 1)
	nearest := geom.Pt{X: a.X + t*delta.X, Y: a.Y + t*delta.Y}
	return math.Hypot(point.X-nearest.X, point.Y-nearest.Y)
}

func pixelRectPath(rect geom.Rect) geom.Path {
	path := geom.Path{}
	path.MoveTo(rect.Min)
	path.LineTo(geom.Pt{X: rect.Max.X, Y: rect.Min.Y})
	path.LineTo(rect.Max)
	path.LineTo(geom.Pt{X: rect.Min.X, Y: rect.Max.Y})
	path.Close()
	return path
}

func pointInPolygon(point geom.Pt, polygon []geom.Pt) bool {
	if len(polygon) < 3 {
		return false
	}
	inside := false
	previous := len(polygon) - 1
	for i := range polygon {
		currentPoint := polygon[i]
		previousPoint := polygon[previous]
		if pointOnSegment(point, previousPoint, currentPoint) {
			return true
		}
		if (currentPoint.Y > point.Y) != (previousPoint.Y > point.Y) {
			x := (previousPoint.X-currentPoint.X)*(point.Y-currentPoint.Y)/(previousPoint.Y-currentPoint.Y) + currentPoint.X
			if point.X < x {
				inside = !inside
			}
		}
		previous = i
	}
	return inside
}

func pointOnSegment(point, a, b geom.Pt) bool {
	const epsilon = 1e-9
	cross := (point.Y-a.Y)*(b.X-a.X) - (point.X-a.X)*(b.Y-a.Y)
	if math.Abs(cross) > epsilon {
		return false
	}
	dot := (point.X-a.X)*(b.X-a.X) + (point.Y-a.Y)*(b.Y-a.Y)
	if dot < -epsilon {
		return false
	}
	lengthSquared := (b.X-a.X)*(b.X-a.X) + (b.Y-a.Y)*(b.Y-a.Y)
	return dot <= lengthSquared+epsilon
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

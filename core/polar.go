package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

const (
	polarCircleSegments          = 128
	polarArcMinSegments          = 12
	defaultPolarRadialLabelAngle = math.Pi / 8
)

type polarDataTransform struct {
	theta          transform.Scale
	r              transform.Scale
	thetaOffset    float64
	thetaDirection float64
}

func (t polarDataTransform) Apply(p geom.Pt) geom.Pt {
	uTheta := 0.0
	if t.theta != nil {
		uTheta = t.theta.Fwd(p.X)
	}
	uRadius := 0.0
	if t.r != nil {
		uRadius = t.r.Fwd(p.Y)
	}

	angle := polarTransformAngle(t.thetaOffset, t.thetaDirection, uTheta)
	return geom.Pt{
		X: 0.5 + 0.5*uRadius*math.Cos(angle),
		Y: 0.5 + 0.5*uRadius*math.Sin(angle),
	}
}

func (t polarDataTransform) Invert(p geom.Pt) (geom.Pt, bool) {
	dx := p.X - 0.5
	dy := p.Y - 0.5
	uRadius := 2 * math.Hypot(dx, dy)
	angle := math.Atan2(dy, dx)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	uTheta := polarInvertAngle(t.thetaOffset, t.thetaDirection, angle)

	theta := 0.0
	if t.theta != nil {
		var ok bool
		theta, ok = t.theta.Inv(uTheta)
		if !ok {
			return geom.Pt{}, false
		}
	}

	radius := 0.0
	if t.r != nil {
		var ok bool
		radius, ok = t.r.Inv(uRadius)
		if !ok {
			return geom.Pt{}, false
		}
	}

	return geom.Pt{X: theta, Y: radius}, true
}

func polarCenterAndRadius(clip geom.Rect) (geom.Pt, float64) {
	size := math.Min(clip.W(), clip.H())
	return geom.Pt{
		X: clip.Min.X + clip.W()/2,
		Y: clip.Min.Y + clip.H()/2,
	}, size / 2
}

func polarProjectionFramePath(proj Projection, clip geom.Rect) geom.Path {
	center, radius := polarCenterAndRadius(clip)
	if sides := radarFrameSidesForProjection(proj); sides >= 3 {
		return polarPolygonFramePath(proj, center, radius, sides)
	}
	return polarCirclePath(center, radius)
}

func polarCirclePath(center geom.Pt, radius float64) geom.Path {
	if radius <= 0 {
		return geom.Path{}
	}
	return polarArcPath(center, radius, 0, 2*math.Pi, polarCircleSegments, true)
}

func polarPolygonFramePath(proj Projection, center geom.Pt, radius float64, sides int) geom.Path {
	if radius <= 0 || sides < 3 {
		return geom.Path{}
	}
	path := geom.Path{}
	for i := range sides {
		angle := polarFrameAngle(proj, i, sides)
		pt := polarPixelPoint(center, radius, angle)
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}

func polarArcPath(center geom.Pt, radius, start, end float64, segments int, closePath bool) geom.Path {
	if radius <= 0 || segments <= 0 {
		return geom.Path{}
	}
	span := end - start
	if closePath && math.Abs(span) >= 2*math.Pi {
		span = 2 * math.Pi
	}

	path := matplotlibArcPath(center, radius, start*180/math.Pi, (start+span)*180/math.Pi)
	if closePath {
		path.Close()
	}
	return path
}

func polarPixelPoint(center geom.Pt, radius, angle float64) geom.Pt {
	return geom.Pt{
		X: center.X + radius*math.Cos(angle),
		Y: center.Y + radius*math.Sin(angle),
	}
}

func polarAngleForTheta(proj Projection, scale transform.Scale, theta float64) float64 {
	offset, direction, _ := polarProjectionState(proj)
	if scale == nil {
		return normalizePolarAngle(offset)
	}
	return polarTransformAngle(offset, direction, scale.Fwd(theta))
}

func polarThetaTicks(axis *Axis, scale transform.Scale, minor bool) []float64 {
	if axis == nil || scale == nil {
		return nil
	}
	domainMin, domainMax := scale.Domain()

	var locator ticker.Locator
	if minor {
		locator = axis.MinorLocator
	} else {
		locator = axis.Locator
	}
	if locator == nil {
		return nil
	}

	count := axis.majorTickTargetCount()
	if minor {
		count = axis.minorTickTargetCount()
	}
	ticks := visibleTicks(locator.Ticks(domainMin, domainMax, count), domainMin, domainMax)
	if len(ticks) == 0 {
		return nil
	}

	span := math.Abs(domainMax - domainMin)
	if approx(span, 2*math.Pi, 1e-9*math.Max(1, span)) && len(ticks) > 1 {
		last := ticks[len(ticks)-1]
		if approx(last, domainMax, 1e-9*math.Max(1, span)) {
			ticks = ticks[:len(ticks)-1]
		}
	}
	return ticks
}

func polarRadialTicks(axis *Axis, scale transform.Scale, minor bool) []float64 {
	if axis == nil || scale == nil {
		return nil
	}
	domainMin, domainMax := scale.Domain()

	var locator ticker.Locator
	if minor {
		locator = axis.MinorLocator
	} else {
		locator = axis.Locator
	}
	if locator == nil {
		return nil
	}

	count := axis.majorTickTargetCount()
	if minor {
		count = axis.minorTickTargetCount()
	}
	ticks := visibleTicks(locator.Ticks(domainMin, domainMax, count), domainMin, domainMax)
	out := make([]float64, 0, len(ticks))
	for _, tick := range ticks {
		u := scale.Fwd(tick)
		if u <= 0 || u > 1 {
			continue
		}
		out = append(out, tick)
	}
	return out
}

func polarIsFullCircle(scale transform.Scale) bool {
	if scale == nil {
		return true
	}
	minVal, maxVal := scale.Domain()
	span := math.Abs(maxVal - minVal)
	return approx(span, 2*math.Pi, 1e-9*math.Max(1, span))
}

func polarTickPaint(color render.Color, width float64, dashes []float64) render.Paint {
	return render.Paint{
		LineWidth: width,
		Stroke:    color,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
		Dashes:    dashes,
	}
}

// polarRadialTickLabelAlignments ports matplotlib's
// PolarAxes.get_yaxis_text1_transform (projections/polar.py): for a
// full-circle theta range the radial tick labels anchor at the projected
// (rlabel_position, r) point with ha='left', va='bottom' and no pad shift.
func polarRadialTickLabelAlignments(fullCircle bool, angle float64) (TextAlign, textLayoutVerticalAlign) {
	if fullCircle {
		return TextAlignLeft, textLayoutVAlignBottom
	}
	return polarTickLabelAlignments(angle)
}

func polarTickLabelAlignments(angle float64) (TextAlign, textLayoutVerticalAlign) {
	angle = math.Mod(angle, 2*math.Pi)
	if angle < 0 {
		angle += 2 * math.Pi
	}

	cosA := math.Cos(angle)
	sinA := math.Sin(angle)

	hAlign := TextAlignCenter
	switch {
	case cosA > 0.25:
		hAlign = TextAlignLeft
	case cosA < -0.25:
		hAlign = TextAlignRight
	}

	vAlign := textLayoutVAlignCenter
	switch {
	case sinA > 0.25:
		vAlign = textLayoutVAlignBottom
	case sinA < -0.25:
		vAlign = textLayoutVAlignTop
	}

	return hAlign, vAlign
}

func polarAxesContainsDisplayPoint(clip geom.Rect, p geom.Pt) bool {
	center, radius := polarCenterAndRadius(clip)
	return math.Hypot(p.X-center.X, p.Y-center.Y) <= radius
}

func polarProjectionContainsDisplayPoint(proj Projection, clip geom.Rect, p geom.Pt) bool {
	if sides := radarFrameSidesForProjection(proj); sides >= 3 {
		path := polarProjectionFramePath(proj, clip)
		return pointInPolygon(p, path.V)
	}
	return polarAxesContainsDisplayPoint(clip, p)
}

func pointInPolygon(p geom.Pt, polygon []geom.Pt) bool {
	if len(polygon) < 3 {
		return false
	}
	inside := false
	j := len(polygon) - 1
	for i := range polygon {
		pi := polygon[i]
		pj := polygon[j]
		if pointOnSegment(p, pj, pi) {
			return true
		}
		if (pi.Y > p.Y) != (pj.Y > p.Y) {
			x := (pj.X-pi.X)*(p.Y-pi.Y)/(pj.Y-pi.Y) + pi.X
			if p.X < x {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}

func pointOnSegment(p, a, b geom.Pt) bool {
	const eps = 1e-9
	cross := (p.Y-a.Y)*(b.X-a.X) - (p.X-a.X)*(b.Y-a.Y)
	if math.Abs(cross) > eps {
		return false
	}
	dot := (p.X-a.X)*(b.X-a.X) + (p.Y-a.Y)*(b.Y-a.Y)
	if dot < -eps {
		return false
	}
	lengthSq := (b.X-a.X)*(b.X-a.X) + (b.Y-a.Y)*(b.Y-a.Y)
	return dot <= lengthSq+eps
}

func polarTransformAngle(offset, direction, uTheta float64) float64 {
	if direction == 0 {
		direction = 1
	}
	return normalizePolarAngle(offset + direction*2*math.Pi*uTheta)
}

func polarInvertAngle(offset, direction, angle float64) float64 {
	if direction == 0 {
		direction = 1
	}
	return normalizePolarFraction(direction * (angle - offset) / (2 * math.Pi))
}

func normalizePolarAngle(angle float64) float64 {
	angle = math.Mod(angle, 2*math.Pi)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	return angle
}

func normalizePolarFraction(v float64) float64 {
	v = math.Mod(v, 1)
	if v < 0 {
		v += 1
	}
	if approx(v, 1, 1e-12) {
		return 0
	}
	return v
}

func polarProjectionForAxes(ax *Axes) (*polarProjection, bool) {
	if ax == nil {
		return nil, false
	}
	return polarProjectionFor(ax.projection)
}

func radarProjectionForAxes(ax *Axes) (*polarProjection, bool) {
	proj, ok := polarProjectionForAxes(ax)
	return proj, ok && proj.isRadar()
}

func polarProjectionFor(proj Projection) (*polarProjection, bool) {
	p, ok := proj.(*polarProjection)
	return p, ok && p != nil
}

func radarFrameSidesForProjection(proj Projection) int {
	p, ok := polarProjectionFor(proj)
	if !ok || !p.isRadar() {
		return 0
	}
	return p.radarVariableCount()
}

func polarFrameAngle(proj Projection, i, sides int) float64 {
	if sides <= 0 {
		return 0
	}
	offset, direction, _ := polarProjectionState(proj)
	return polarTransformAngle(offset, direction, float64(i)/float64(sides))
}

func polarProjectionState(proj Projection) (offset, direction, radialLabelAngle float64) {
	offset = 0
	direction = 1
	radialLabelAngle = defaultPolarRadialLabelAngle
	if p, ok := polarProjectionFor(proj); ok {
		offset = p.thetaOffset
		if p.thetaDirection != 0 {
			direction = p.thetaDirection
		}
		if p.radialLabelAngle != 0 {
			radialLabelAngle = p.radialLabelAngle
		}
	}
	return offset, direction, radialLabelAngle
}

// polarRadialLabelAngleForProjection returns the display-space angle of the
// radial tick labels. matplotlib stores rlabel_position in theta-data space
// and maps it through the theta direction and offset
// (polar.py RadialTick.update_position:
// `angle = rlabel_position * direction + offset`).
func polarRadialLabelAngleForProjection(proj Projection) float64 {
	offset, direction, rlabel := polarProjectionState(proj)
	return normalizePolarAngle(offset + direction*rlabel)
}

func polarCompassAngle(location string) (float64, bool) {
	switch strings.ToLower(strings.TrimSpace(location)) {
	case "e", "east":
		return 0, true
	case "ne", "northeast", "north-east":
		return math.Pi / 4, true
	case "n", "north":
		return math.Pi / 2, true
	case "nw", "northwest", "north-west":
		return 3 * math.Pi / 4, true
	case "w", "west":
		return math.Pi, true
	case "sw", "southwest", "south-west":
		return 5 * math.Pi / 4, true
	case "s", "south":
		return 3 * math.Pi / 2, true
	case "se", "southeast", "south-east":
		return 7 * math.Pi / 4, true
	default:
		return 0, false
	}
}

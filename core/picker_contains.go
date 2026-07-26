package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

// pickRadiusFor returns either the artist's overridden pick radius or the
// default tolerance. The artist itself is unused but kept on the receiver for
// clarity at call sites.
//
// The Line2D PickRadius accessor is defined separately in this file.

// transformForPick resolves the per-artist display transform from a coords
// specification. Data-only artists can also use the raw DataToPixel transform
// via [dataToPixel].
func transformForPick(ctx *DrawContext, coords CoordinateSpec) transform.T {
	if ctx == nil {
		return nil
	}
	return ctx.TransformFor(coords)
}

// transformPath maps a path through a transform and returns the resulting
// pixel-space path. Curve commands are flattened to their endpoints — adequate
// for hit-testing against the approximate polygon outline used everywhere in
// this code base (circles, ellipses, rounded boxes already sample many
// vertices).
func transformPath(path geom.Path, tr transform.T) []geom.Pt {
	if len(path.V) == 0 {
		return nil
	}
	out := make([]geom.Pt, 0, len(path.V))
	for _, pt := range path.V {
		if tr != nil {
			out = append(out, tr.Apply(pt))
		} else {
			out = append(out, pt)
		}
	}
	return out
}

// PickRadius returns the cursor tolerance, in pixels, the line uses for hit
// testing. Override it by setting the unexported pickRadius field via
// SetPickRadius. The default of 0 means [DefaultPickRadius] applies.
func (l *Line2D) PickRadius() float64 {
	if l == nil {
		return DefaultPickRadius
	}
	if l.pickRadius > 0 {
		return l.pickRadius
	}
	return DefaultPickRadius
}

// SetPickRadius sets the cursor tolerance in pixels. Use a non-positive value
// to restore the default.
func (l *Line2D) SetPickRadius(r float64) {
	if l == nil {
		return
	}
	if r < 0 {
		r = 0
	}
	l.pickRadius = r
}

// Contains reports whether p (in figure pixels) lies within the line's pick
// tolerance of any segment of the polyline.
func (l *Line2D) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if l == nil || ctx == nil {
		return false, PickInfo{}
	}
	points := l.pathPoints()
	if len(points) == 0 {
		return false, PickInfo{}
	}
	tol := l.PickRadius()
	if w := l.W; w > 0 {
		tol = math.Max(tol, w/2)
	}

	pxPoints := make([]geom.Pt, len(points))
	for i, pt := range points {
		pxPoints[i] = (&ctx.DataToPixel).Apply(pt)
	}

	bestDist := math.Inf(1)
	bestIdx := -1
	if len(pxPoints) == 1 {
		bestDist = math.Hypot(p.X-pxPoints[0].X, p.Y-pxPoints[0].Y)
		bestIdx = 0
	}
	for i := 1; i < len(pxPoints); i++ {
		d := distancePointToSegment(pxPoints[i-1], pxPoints[i], p)
		if d < bestDist {
			bestDist = d
			bestIdx = i - 1
		}
	}
	if bestDist <= tol {
		return true, PickInfo{Index: bestIdx, Distance: bestDist}
	}
	return false, PickInfo{}
}

// containsPath is the shared hit-test for closed patch artists. It transforms
// the artist's local path into pixel space and tests inclusion using the
// even-odd rule. Stroke-only patches still match when the cursor is near the
// edge within the patch's stroke half-width.
func containsPath(local geom.Path, affine geom.Affine, coords CoordinateSpec, p geom.Pt, ctx *DrawContext, stroke float64) bool {
	if ctx == nil || len(local.C) == 0 {
		return false
	}
	tr := transformForPick(ctx, coords)
	displayPath := applyAffinePath(local, affine)
	pxPoints := transformPath(displayPath, tr)
	if len(pxPoints) < 2 {
		return false
	}
	if pointInPolygon(p, pxPoints) {
		return true
	}
	if stroke > 0 {
		tol := math.Max(DefaultPickRadius, stroke/2)
		// closing edge: from last to first
		if d := distancePointToSegment(pxPoints[len(pxPoints)-1], pxPoints[0], p); d <= tol {
			return true
		}
		for i := 1; i < len(pxPoints); i++ {
			if d := distancePointToSegment(pxPoints[i-1], pxPoints[i], p); d <= tol {
				return true
			}
		}
	}
	return false
}

// Contains reports whether p (in figure pixels) lies inside the rectangle
// (including its rotated extent).
func (r *Rectangle) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if r == nil {
		return false, PickInfo{}
	}
	local := rectanglePath(r.Width, r.Height)
	if containsPath(local, patchAffine(r.XY, r.Angle), r.Coords, p, ctx, r.EdgeWidth.OrZero()) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}

// Contains reports whether p (in figure pixels) lies inside the circle.
func (c *Circle) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if c == nil || c.Radius <= 0 {
		return false, PickInfo{}
	}
	local := ellipsePath(c.Radius*2, c.Radius*2)
	if containsPath(local, translateAffine(c.Center), c.Coords, p, ctx, c.EdgeWidth.OrZero()) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}

// Contains reports whether p (in figure pixels) lies inside the ellipse.
func (e *Ellipse) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if e == nil || e.Width == 0 || e.Height == 0 {
		return false, PickInfo{}
	}
	local := ellipsePath(e.Width, e.Height)
	if containsPath(local, patchAffine(e.Center, e.Angle), e.Coords, p, ctx, e.EdgeWidth.OrZero()) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}

// Contains reports whether p (in figure pixels) lies inside the polygon. Open
// polygons fall back to a polyline distance test.
func (poly *Polygon) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if poly == nil || len(poly.XY) < 2 {
		return false, PickInfo{}
	}
	if !poly.Open {
		local := polygonPath(poly.XY, true)
		if containsPath(local, geom.Identity(), poly.Coords, p, ctx, poly.EdgeWidth.OrZero()) {
			return true, PickInfo{}
		}
		return false, PickInfo{}
	}
	tr := transformForPick(ctx, poly.Coords)
	pxPoints := transformPath(polygonPath(poly.XY, false), tr)
	tol := math.Max(DefaultPickRadius, poly.EdgeWidth.OrZero()/2)
	for i := 1; i < len(pxPoints); i++ {
		if distancePointToSegment(pxPoints[i-1], pxPoints[i], p) <= tol {
			return true, PickInfo{Index: i - 1}
		}
	}
	return false, PickInfo{}
}

// Contains reports whether p (in figure pixels) lies inside the path patch.
func (pp *PathPatch) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if pp == nil || len(pp.Path.C) == 0 {
		return false, PickInfo{}
	}
	if containsPath(pp.Path, geom.Identity(), pp.Coords, p, ctx, pp.EdgeWidth.OrZero()) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}

// Contains reports whether p (in figure pixels) lies inside the fancy bbox.
func (b *FancyBboxPatch) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if b == nil {
		return false, PickInfo{}
	}
	if containsPath(b.localPath(), translateAffine(b.XY), b.Coords, p, ctx, b.EdgeWidth.OrZero()) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}

// Contains reports whether p (in figure pixels) lies inside the arrow's
// filled outline.
func (a *FancyArrow) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if a == nil || ctx == nil {
		return false, PickInfo{}
	}
	local := a.localPath()
	if len(local.C) == 0 {
		return false, PickInfo{}
	}
	if containsPath(local, geom.Identity(), a.Coords, p, ctx, a.EdgeWidth.OrZero()) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}

// Contains reports whether p (in figure pixels) lies inside the image's
// rectangular extent.
func (i *Image2D) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if i == nil || ctx == nil {
		return false, PickInfo{}
	}
	bounds := i.Bounds(ctx)
	if bounds == (geom.Rect{}) {
		return false, PickInfo{}
	}
	corners := [4]geom.Pt{
		bounds.Min,
		{X: bounds.Max.X, Y: bounds.Min.Y},
		bounds.Max,
		{X: bounds.Min.X, Y: bounds.Max.Y},
	}
	pxCorners := make([]geom.Pt, 4)
	for k, c := range corners {
		pxCorners[k] = (&ctx.DataToPixel).Apply(c)
	}
	if pointInPolygon(p, pxCorners) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}

// Contains reports whether p (in figure pixels) lies within the bounding box
// of the text. Glyph metrics come from the shared sfnt shaper
// (render.MeasureTextMetrics) using the resolved font, matching Matplotlib's
// exact-bbox behavior rather than the old FontSize×rune-count heuristic. If the
// font cannot be shaped (font-less environments) MeasureTextMetrics falls back
// to a proportional estimate, so this never panics.
func (t *Text) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if t == nil || ctx == nil || t.Content == "" {
		return false, PickInfo{}
	}
	tr := transformForPick(ctx, t.Coords)
	if tr == nil {
		return false, PickInfo{}
	}
	anchor := tr.Apply(t.Position)
	anchor.X += t.OffsetX
	anchor.Y += t.OffsetY

	fontSize := t.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}
	fontKey := resolvedTextFontKey(t.FontKey, t.FontProperties, ctx)

	// Measure each line with the real font; the bounding box width is the widest
	// line advance and the height stacks the per-line ascent+descent.
	lines := strings.Split(t.Content, "\n")
	var width, ascent, descent float64
	for i, line := range lines {
		m := render.MeasureTextMetrics(line, fontSize, fontKey)
		if m.W > width {
			width = m.W
		}
		if i == 0 {
			ascent = m.Ascent
		}
		if i == len(lines)-1 {
			descent = m.Descent
		}
		if i > 0 {
			// Interior line break advances by a full line height.
			ascent += m.Ascent + m.Descent
		}
	}
	height := ascent + descent
	if width <= 0 {
		width = fontSize
	}
	if height <= 0 {
		height = fontSize
		ascent, descent = fontSize*0.8, fontSize*0.2
	}

	var minX, maxX, minY, maxY float64
	switch t.HAlign {
	case TextAlignRight:
		minX, maxX = anchor.X-width, anchor.X
	case TextAlignCenter:
		minX, maxX = anchor.X-width/2, anchor.X+width/2
	default: // TextAlignLeft
		minX, maxX = anchor.X, anchor.X+width
	}
	switch t.VAlign {
	case TextVAlignTop:
		minY, maxY = anchor.Y, anchor.Y+height
	case TextVAlignMiddle:
		minY, maxY = anchor.Y-height/2, anchor.Y+height/2
	case TextVAlignBottom:
		minY, maxY = anchor.Y-height, anchor.Y
	default: // TextVAlignBaseline — split the box at the first-line baseline
		// (display space is y-down, so ascent extends to smaller Y).
		minY, maxY = anchor.Y-ascent, anchor.Y+descent
	}
	if p.X >= minX && p.X <= maxX && p.Y >= minY && p.Y <= maxY {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}

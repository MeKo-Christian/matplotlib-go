package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func (a *Annotation) drawAnnotationOverlay(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil {
		return
	}

	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	if displayTextIsEmpty(a.Content) {
		return
	}

	fontSize := resolvedFontSize(a.FontSize, ctx)
	fontKey := resolvedTextFontKey(a.FontKey, a.FontProperties, ctx)
	parseMath := parseMathEnabled(a.ParseMath)
	target := transformedPoint(ctx, a.Coords, a.Point, 0, 0)
	if annotationPointClipped(a.AnnotationClip, a.Coords, target, ctx.Clip) {
		return
	}
	anchor := a.textAnchor(ctx)
	lines := strings.Split(a.Content, "\n")
	if len(lines) > 1 {
		a.drawMultilineAnnotation(r, textRen, ctx, target, anchor, fontSize, fontKey, parseMath, lines)
		return
	}
	layout := measureSingleLineTextLayoutParseMath(r, a.Content, fontSize, fontKey, parseMath, ctx.RC.UseTeX)
	origin := alignedSingleLineOrigin(anchor, layout, a.HAlign, layoutVerticalAlign(a.VAlign, false))
	box, ok := textInkRect(origin, layout)
	if !ok {
		box = geom.Rect{
			Min: geom.Pt{X: origin.X, Y: origin.Y - layout.Ascent},
			Max: geom.Pt{X: origin.X + layout.Width, Y: origin.Y + layout.Descent},
		}
	}
	if bbox, ok := textBBoxRect(origin, layout, a.BBox, ctx, fontSize); ok {
		box = bbox
	}
	a.drawArrowFromBox(r, ctx, box, target)

	textColor := a.ApplyArtistAlpha(resolvedTextColor(a.Color, ctx))
	if a.Angle != 0 {
		if rotated, ok := r.(render.RotatedTextDrawer); ok {
			angle := a.Angle * math.Pi / 180
			vAlign := layoutVerticalAlign(a.VAlign, false)
			drawOrigin := tickLabelDrawOriginFromP(anchor, layout, a.HAlign, vAlign, angle, false)
			rotAnchor := rotatedTextBackendAnchorFromP(anchor, layout, a.HAlign, vAlign, angle, false)
			drawTextBBoxRotated(r, anchor, drawOrigin, layout, a.BBox, ctx, fontSize, a.Angle)
			drawDisplayTextRotatedParseMath(rotated, a.Content, rotAnchor, fontSize, angle, textColor, fontKey, parseMath, ctx.RC.UseTeX)
			return
		}
	}
	drawTextBBox(r, origin, layout, a.BBox, ctx, fontSize)
	drawDisplayTextParseMath(textRen, a.Content, origin, fontSize, textColor, fontKey, parseMath, ctx.RC.UseTeX)
}

func (a *Annotation) drawMultilineAnnotation(r render.Renderer, textRen render.TextDrawer, ctx *DrawContext, target, anchor geom.Pt, fontSize float64, fontKey string, parseMath bool, lines []string) {
	rect, ok := multilineTextBlockRect(r, ctx, anchor, fontSize, fontKey, parseMath, lines, a.Linespacing, a.HAlign, a.VAlign)
	if !ok {
		return
	}
	box := rect
	if a.BBox != nil {
		cfg := resolvedTextBBoxOptions(*a.BBox, ctx, fontSize)
		box.Min.X -= cfg.Padding
		box.Min.Y -= cfg.Padding
		box.Max.X += cfg.Padding
		box.Max.Y += cfg.Padding
	}
	a.drawArrowFromBox(r, ctx, box, target)
	text := &Text{
		ArtistRasterization: a.ArtistRasterization,
		Content:             strings.Join(lines, "\n"),
		HAlign:              a.HAlign,
		VAlign:              a.VAlign,
		Angle:               a.Angle,
		BBox:                a.BBox,
		Linespacing:         a.Linespacing,
		Color:               a.Color,
		ParseMath:           a.ParseMath,
	}
	text.drawMultilineText(r, textRen, ctx, anchor, fontSize, fontKey, parseMath, lines)
}

func multilineTextBlockRect(r render.Renderer, ctx *DrawContext, anchor geom.Pt, fontSize float64, fontKey string, parseMath bool, lines []string, linespacing float64, hAlign TextAlign, vAlign TextVerticalAlign) (geom.Rect, bool) {
	block, ok := measureMultilineTextBlock(r, ctx, anchor, fontSize, fontKey, parseMath, ctx.RC.UseTeX, lines, linespacing, hAlign, layoutVerticalAlign(vAlign, false))
	if !ok {
		return geom.Rect{}, false
	}
	return block.Rect, true
}

func (a *Annotation) textAnchor(ctx *DrawContext) geom.Pt {
	offset := a.offsetPixels(ctx)
	if a != nil && a.TextPosition != nil {
		return transformedPoint(ctx, a.TextCoords, *a.TextPosition, offset.X, offset.Y)
	}
	if a == nil {
		return geom.Pt{}
	}
	return transformedPoint(ctx, a.Coords, a.Point, offset.X, offset.Y)
}

func (a *Annotation) offsetPixels(ctx *DrawContext) geom.Pt {
	if a == nil {
		return geom.Pt{}
	}
	offset := geom.Pt{X: a.OffsetX, Y: a.OffsetY}
	if a.OffsetUnits == AnnotationOffsetPoints && ctx != nil {
		offset.X = pointsToPixels(ctx.RC, offset.X)
		offset.Y = pointsToPixels(ctx.RC, offset.Y)
	}
	return offset
}

func (a *Annotation) drawArrow(r render.Renderer, ctx *DrawContext, start, target geom.Pt) {
	path := a.ConnectionStyle.connect(start, target, arrowShrinkPixels(ctx, 2), arrowShrinkPixels(ctx, 2))
	a.drawArrowPath(r, ctx, path)
}

func (a *Annotation) drawArrowFromBox(r render.Renderer, ctx *DrawContext, box geom.Rect, target geom.Pt) {
	clipBox := box
	if a.BBox == nil {
		pad := arrowPlainTextPatchPadding(ctx)
		clipBox.Min.X -= pad
		clipBox.Min.Y -= pad
		clipBox.Max.X += pad
		clipBox.Max.Y += pad
	}
	a.drawArrowFromPatchBox(r, ctx, box, clipBox, target)
}

func (a *Annotation) drawArrowFromPatchBox(r render.Renderer, ctx *DrawContext, relposBox, clipBox geom.Rect, target geom.Pt) {
	start := rectCenter(relposBox)
	path := a.ConnectionStyle.connect(start, target, 0, 0)
	path = clipConnectionPathToRect(path, clipBox, true)
	path = shrinkPathEndpoints(path, arrowShrinkPixels(ctx, 2), arrowShrinkPixels(ctx, 2))
	a.drawArrowPath(r, ctx, path)
}

func (a *Annotation) drawArrowPath(r render.Renderer, ctx *DrawContext, path geom.Path) {
	if a.ArrowWidth <= 0 || a.ArrowHeadSize <= 0 {
		return
	}
	color := a.ApplyArtistAlpha(resolvedArrowColor(a.ArrowColor, a.Color, ctx))
	patch := &FancyArrowPatch{
		Patch: Patch{
			FaceColor: color,
			EdgeColor: color,
			EdgeWidth: a.ArrowWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		},
		ArrowStyle:      a.ArrowStyle,
		ConnectionStyle: a.ConnectionStyle,
		MutationScale:   a.ArrowHeadSize,
	}
	for _, part := range patch.displayParts(ctx, path) {
		if len(part.path.C) == 0 {
			continue
		}
		if part.fillable {
			patch.drawStyledPath(r, &ctx.RC, part.path, geom.Path{})
		} else {
			patch.drawStyledPath(r, &ctx.RC, geom.Path{}, part.path)
		}
	}
}

func arrowPlainTextPatchPadding(ctx *DrawContext) float64 {
	if ctx == nil {
		return 2
	}
	return pointsToPixels(ctx.RC, 4) / 2
}

func clipConnectionPathToRect(path geom.Path, rect geom.Rect, start bool) geom.Path {
	if len(path.V) < 2 {
		return path
	}
	polygon := pixelRectPath(rect).Interpolated(8).V
	if len(polygon) < 3 {
		return path
	}
	endpoint := pathEnd(path)
	if start {
		endpoint = pathStart(path)
	}
	if !pointInPolygon(endpoint, polygon) {
		return path
	}
	if clipped, ok := clipQuadraticConnectionPathToPolygon(path, polygon, start); ok {
		return clipped
	}
	boundary, ok := connectionPatchBoundaryPoint(path, polygon, start)
	if !ok {
		return path
	}
	out := path
	out.V = append([]geom.Pt(nil), path.V...)
	if start {
		out.V[0] = boundary
	} else {
		out.V[len(out.V)-1] = boundary
	}
	return out
}

func resolvedArrowColor(arrow, text render.Color, ctx *DrawContext) render.Color {
	if arrow == (render.Color{}) {
		return resolvedTextColor(text, ctx)
	}
	if arrow.A == 0 && (arrow.R != 0 || arrow.G != 0 || arrow.B != 0) {
		arrow.A = 1
	}
	return arrow
}

func annotationPointClipped(annotationClip *bool, coords CoordinateSpec, target geom.Pt, clip geom.Rect) bool {
	if annotationClip != nil {
		return *annotationClip && !clip.ContainsInclusive(target)
	}
	return coords == Coords(CoordData) && !clip.ContainsInclusive(target)
}

func transformedPoint(ctx *DrawContext, spec CoordinateSpec, p geom.Pt, dxPx, dyPx float64) geom.Pt {
	if ctx == nil {
		return p
	}

	base := ctx.TransformFor(spec)
	if base == nil {
		return p
	}
	if dxPx != 0 || dyPx != 0 {
		return transform.NewOffset(base, geom.Pt{X: dxPx, Y: dyPx}).Apply(p)
	}
	return base.Apply(p)
}

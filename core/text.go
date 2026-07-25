package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func (t *Text) drawText(r render.Renderer, ctx *DrawContext) {
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	if displayTextIsEmpty(t.Content) {
		return
	}

	fontSize := resolvedFontSize(t.FontSize, ctx)
	fontKey := resolvedTextFontKey(t.FontKey, t.FontProperties, ctx)
	parseMath := parseMathEnabled(t.ParseMath)
	anchor := transformedPoint(ctx, t.Coords, t.Position, t.OffsetX, t.OffsetY)
	wrapWidth := t.WrapWidth
	if wrapWidth <= 0 && t.Wrap && !ctx.RC.UseTeX {
		wrapWidth = textAutoWrapWidth(ctx, anchor, t.HAlign, t.Angle)
	}
	lines := wrappedTextLines(r, t.Content, fontSize, fontKey, parseMath, ctx.RC.UseTeX, wrapWidth)
	if len(lines) > 1 {
		t.drawMultilineText(r, textRen, ctx, anchor, fontSize, fontKey, parseMath, lines)
		return
	}
	content := t.Content
	if len(lines) == 1 {
		content = lines[0]
	}
	layout := measureSingleLineTextLayoutParseMath(r, content, fontSize, fontKey, parseMath, ctx.RC.UseTeX)
	hAlign, vAlign := textRotationLayoutAlignments(t.HAlign, t.VAlign, t.Angle, t.RotationMode)
	origin := alignedSingleLineOrigin(anchor, layout, hAlign, vAlign)
	textColor := t.ApplyArtistAlpha(resolvedTextColor(t.Color, ctx))
	if t.Angle != 0 {
		if rotated, ok := r.(render.RotatedTextDrawer); ok {
			angle := t.Angle * math.Pi / 180
			anchorMode := t.RotationMode == TextRotationModeAnchor
			drawOrigin := tickLabelDrawOriginFromP(anchor, layout, hAlign, vAlign, angle, anchorMode)
			rotAnchor := rotatedTextBackendAnchorFromP(anchor, layout, hAlign, vAlign, angle, anchorMode)
			drawTextBBoxRotated(r, anchor, drawOrigin, layout, t.BBox, ctx, fontSize, t.Angle)
			if len(t.PathEffects) > 0 && drawTextPathEffects(r, ctx.RC, content, drawOrigin, rotAnchor, fontSize, angle, textColor, fontKey, ctx.RC.UseTeX, t.PathEffects) {
				return
			}
			drawDisplayTextRotatedParseMath(rotated, content, rotAnchor, fontSize, angle, textColor, fontKey, parseMath, ctx.RC.UseTeX)
			return
		}
	}
	drawTextBBox(r, origin, layout, t.BBox, ctx, fontSize)
	if len(t.PathEffects) > 0 && drawTextPathEffects(r, ctx.RC, content, origin, origin, fontSize, 0, textColor, fontKey, ctx.RC.UseTeX, t.PathEffects) {
		return
	}
	drawDisplayTextParseMath(textRen, content, origin, fontSize, textColor, fontKey, parseMath, ctx.RC.UseTeX)
}

func (t *Text) drawMultilineText(r render.Renderer, textRen render.TextDrawer, ctx *DrawContext, anchor geom.Pt, fontSize float64, fontKey string, parseMath bool, lines []string) {
	hAlign, vAlign := textRotationLayoutAlignments(t.HAlign, t.VAlign, t.Angle, t.RotationMode)
	textColor := t.ApplyArtistAlpha(resolvedTextColor(t.Color, ctx))
	block, ok := measureMultilineTextBlock(r, ctx, anchor, fontSize, fontKey, parseMath, ctx.RC.UseTeX, lines, t.Linespacing, hAlign, vAlign)
	if !ok {
		return
	}
	lineAlign := hAlign
	if t.MultiAlignment != nil {
		lineAlign = *t.MultiAlignment
	}

	if t.Angle != 0 {
		if rotated, ok := r.(render.RotatedTextDrawer); ok {
			angle := t.Angle * math.Pi / 180
			rotatedBlock, ok := measureRotatedMultilineTextBlock(r, fontSize, fontKey, parseMath, ctx.RC.UseTeX, lines, t.Linespacing, hAlign, vAlign, lineAlign, angle, t.RotationMode)
			if ok {
				if t.BBox != nil {
					drawMultilineTextBBoxRotatedMatplotlib(r, anchor, rotatedBlock, t.BBox, ctx, fontSize, t.Angle)
				}
				for i, line := range lines {
					if line == "" {
						continue
					}
					origin := geom.Pt{
						X: anchor.X + rotatedBlock.LineOffsets[i].X,
						Y: anchor.Y + rotatedBlock.LineOffsets[i].Y,
					}
					rotAnchor := rotatedTextBackendAnchorForOrigin(origin, rotatedBlock.Layouts[i], angle)
					if len(t.PathEffects) > 0 && drawTextPathEffects(r, ctx.RC, line, origin, rotAnchor, fontSize, angle, textColor, fontKey, ctx.RC.UseTeX, t.PathEffects) {
						continue
					}
					drawDisplayTextRotatedParseMath(rotated, line, rotAnchor, fontSize, angle, textColor, fontKey, parseMath, ctx.RC.UseTeX)
				}
				return
			}
		}
	}

	if t.BBox != nil {
		if t.Angle != 0 {
			if _, ok := r.(render.RotatedTextDrawer); ok {
				drawMultilineTextBBoxRotated(r, block.Rect, t.BBox, ctx, fontSize, geom.Pt{
					X: block.Rect.Min.X + block.Width/2,
					Y: block.Rect.Min.Y + block.Height,
				}, t.Angle)
			} else {
				drawMultilineTextBBox(r, block.Rect, t.BBox, ctx, fontSize)
			}
		} else {
			drawMultilineTextBBox(r, block.Rect, t.BBox, ctx, fontSize)
		}
	}

	for i, line := range lines {
		if line == "" {
			continue
		}
		origin := geom.Pt{
			X: block.Rect.Min.X,
			Y: block.BaselineYs[i],
		}
		if block.Layouts[i].Width < block.Width {
			switch lineAlign {
			case TextAlignCenter:
				origin.X += (block.Width - block.Layouts[i].Width) / 2
			case TextAlignRight:
				origin.X += block.Width - block.Layouts[i].Width
			}
		}
		if t.Angle != 0 {
			if rotated, ok := r.(render.RotatedTextDrawer); ok {
				angle := t.Angle * math.Pi / 180
				rotAnchor := textRotationAnchor(origin, block.Layouts[i], lineAlign, textLayoutVAlignBaseline, angle, t.RotationMode)
				if len(t.PathEffects) > 0 && drawTextPathEffects(r, ctx.RC, line, origin, rotAnchor, fontSize, angle, textColor, fontKey, ctx.RC.UseTeX, t.PathEffects) {
					continue
				}
				drawDisplayTextRotatedParseMath(rotated, line, rotAnchor, fontSize, angle, textColor, fontKey, parseMath, ctx.RC.UseTeX)
				continue
			}
		}
		if len(t.PathEffects) > 0 && drawTextPathEffects(r, ctx.RC, line, origin, origin, fontSize, 0, textColor, fontKey, ctx.RC.UseTeX, t.PathEffects) {
			continue
		}
		drawDisplayTextParseMath(textRen, line, origin, fontSize, textColor, fontKey, parseMath, ctx.RC.UseTeX)
	}
}

func drawTextPathEffects(r render.Renderer, rc style.RC, text string, origin, pivot geom.Pt, size, angle float64, textColor render.Color, fontKey string, useTeX bool, effects []render.PathEffect) bool {
	if r == nil || len(effects) == 0 || useTeX {
		return false
	}

	paths, ok := displayTextPaths(r, text, origin, size, fontKey)
	if !ok {
		return false
	}
	if angle != 0 {
		cos := math.Cos(angle)
		sin := math.Sin(angle)
		affine := translateAffine(pivot).
			Mul(geom.Affine{A: cos, B: sin, C: -sin, D: cos}).
			Mul(translateAffine(geom.Pt{X: -pivot.X, Y: -pivot.Y}))
		for i := range paths {
			paths[i] = applyAffinePath(paths[i], affine)
		}
	}

	paint := render.Paint{
		Fill:        textColor,
		PathEffects: devicePathEffects(rc, effects),
	}
	for _, path := range paths {
		r.Path(path, &paint)
	}
	return true
}

func displayTextPaths(r render.Renderer, text string, origin geom.Pt, size float64, fontKey string) ([]geom.Path, bool) {
	if layout, ok := layoutDisplayText(r, text, size, fontKey); ok {
		return mathTextLayoutPaths(r, layout, origin, fontKey)
	}
	display := normalizeDisplayText(text)
	if display == "" {
		return nil, false
	}
	if pather, ok := r.(render.TextPather); ok {
		if path, ok := pather.TextPath(display, origin, size, fontKey); ok {
			return []geom.Path{path}, true
		}
	}
	if path, ok := render.TextPath(display, origin, size, fontKey); ok {
		return []geom.Path{path}, true
	}
	return nil, false
}

func resolvedTextColor(c render.Color, ctx *DrawContext) render.Color {
	if c == (render.Color{}) {
		if ctx != nil {
			return ctx.RC.DefaultTextColor()
		}
		return render.Color{R: 0, G: 0, B: 0, A: 1}
	}
	if c.A == 0 && (c.R != 0 || c.G != 0 || c.B != 0) {
		c.A = 1
	}
	return c
}

func resolvedFontSize(size float64, ctx *DrawContext) float64 {
	if size > 0 {
		return size
	}
	if ctx != nil && ctx.RC.FontSize > 0 {
		return ctx.RC.FontSize
	}
	return 12
}

// DrawAlignedText draws a single line of text aligned around an anchor point.
// It applies the font, math-text, TeX, and color settings from ctx.
func DrawAlignedText(r render.Renderer, ctx *DrawContext, anchor geom.Pt, text string, size float64, color render.Color, hAlign TextAlign, vAlign TextVerticalAlign) {
	DrawAlignedTextWithFont(r, ctx, anchor, text, size, color, hAlign, vAlign, "")
}

// DrawAlignedTextWithFont draws aligned single-line text with an explicit font
// key. An empty font key inherits the draw context font.
func DrawAlignedTextWithFont(r render.Renderer, ctx *DrawContext, anchor geom.Pt, text string, size float64, color render.Color, hAlign TextAlign, vAlign TextVerticalAlign, fontKey string) {
	if r == nil || ctx == nil || displayTextIsEmpty(text) {
		return
	}
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}
	fontSize := resolvedFontSize(size, ctx)
	if fontKey == "" {
		fontKey = ctx.RC.FontKey
	}
	layout := measureSingleLineTextLayout(r, text, fontSize, fontKey, ctx.RC.UseTeX)
	origin := alignedSingleLineOrigin(anchor, layout, hAlign, layoutVerticalAlign(vAlign, false))
	drawDisplayText(textRen, text, origin, fontSize, resolvedTextColor(color, ctx), fontKey, ctx.RC.UseTeX)
}

// FontKeyWithWeight returns fontKey with the requested numeric weight.
func FontKeyWithWeight(fontKey string, weight int) string {
	return fontKeyWithWeight(fontKey, weight)
}

func resolvedTextFontKey(fontKey string, props *render.FontProperties, ctx *DrawContext) string {
	fontKey = strings.TrimSpace(fontKey)
	if fontKey != "" {
		return fontKey
	}
	if props != nil {
		return render.FontPropertiesKey(*props)
	}
	if ctx != nil {
		return ctx.RC.FontKey
	}
	return ""
}

func parseMathEnabled(parseMath *bool) bool {
	return parseMath == nil || *parseMath
}

func cloneFontProperties(props *render.FontProperties) *render.FontProperties {
	if props == nil {
		return nil
	}
	cloned := *props
	cloned.Families = append([]string(nil), props.Families...)
	cloned.Features = append([]render.TextFeature(nil), props.Features...)
	return &cloned
}

func cloneBool(v *bool) *bool {
	if v == nil {
		return nil
	}
	cloned := *v
	return &cloned
}

func cloneTextBBoxOptions(opt *TextBBoxOptions) *TextBBoxOptions {
	if opt == nil {
		return nil
	}
	cloned := *opt
	return &cloned
}

func cloneTextAlign(align *TextAlign) *TextAlign {
	if align == nil {
		return nil
	}
	cloned := *align
	return &cloned
}

func resolvedTextLinespacing(linespacing float64) float64 {
	if linespacing > 0 {
		return linespacing
	}
	return 1.2
}

func nearestPointOnRect(rect geom.Rect, pt geom.Pt) geom.Pt {
	return geom.Pt{
		X: clampFloat(pt.X, rect.Min.X, rect.Max.X),
		Y: clampFloat(pt.Y, rect.Min.Y, rect.Max.Y),
	}
}

func clampFloat(v, minVal, maxVal float64) float64 {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

func pixelLinePath(p1, p2 geom.Pt) geom.Path {
	return geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{p1, p2},
	}
}

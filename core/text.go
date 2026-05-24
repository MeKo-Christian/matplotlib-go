package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

// TextAlign controls horizontal text anchoring.
type TextAlign uint8

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
)

// TextVerticalAlign controls vertical text anchoring.
type TextVerticalAlign uint8

const (
	TextVAlignBaseline TextVerticalAlign = iota
	TextVAlignBottom
	TextVAlignMiddle
	TextVAlignTop
)

// TextRotationMode controls how alignment interacts with text rotation.
type TextRotationMode string

const (
	TextRotationModeDefault TextRotationMode = ""
	TextRotationModeAnchor  TextRotationMode = "anchor"
	TextRotationModeXTick   TextRotationMode = "xtick"
	TextRotationModeYTick   TextRotationMode = "ytick"
)

// TextOptions configures a Text artist.
type TextOptions struct {
	FontSize float64
	Color    render.Color
	HAlign   TextAlign
	VAlign   TextVerticalAlign
	Angle    float64
	// RotationModeAnchor aligns the unrotated text first, then rotates it around
	// the text box anchor. The zero value keeps Matplotlib's default rotated-bbox
	// alignment behavior.
	RotationMode TextRotationMode
	Coords       CoordinateSpec
	OffsetX      float64
	OffsetY      float64
	// WrapWidth wraps text to this maximum display-pixel width when positive.
	WrapWidth float64
	// MultiAlignment controls per-line alignment within multiline or wrapped
	// text. Nil follows HAlign, matching Matplotlib's multialignment=None.
	MultiAlignment *TextAlign
	ClipOn         *bool
	BBox           *TextBBoxOptions
	FontKey        string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties
	ParseMath      *bool
	PathEffects    []render.PathEffect
}

// TextBBoxOptions configures a rectangular background behind text.
type TextBBoxOptions struct {
	FaceColor    render.Color
	EdgeColor    render.Color
	LineWidth    float64
	Padding      float64
	CornerRadius float64
}

// AnnotationOptions configures an Annotation artist.
type AnnotationOptions struct {
	Coords          CoordinateSpec
	OffsetX         float64
	OffsetY         float64
	FontSize        float64
	Color           render.Color
	ArrowColor      render.Color
	ArrowWidth      float64
	ArrowHeadSize   float64
	ArrowStyle      ArrowStyle
	ConnectionStyle ConnectionStyle
	HAlign          TextAlign
	VAlign          TextVerticalAlign
	FontKey         string
	BBox            *TextBBoxOptions
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties
	ParseMath      *bool
	AnnotationClip *bool
}

// Text renders arbitrary text at a data-space position.
type Text struct {
	ArtistRasterization
	Position geom.Pt
	Content  string

	FontSize float64
	Color    render.Color
	HAlign   TextAlign
	VAlign   TextVerticalAlign
	Angle    float64
	// RotationModeAnchor aligns the unrotated text first, then rotates it around
	// the text box anchor. The zero value keeps Matplotlib's default rotated-bbox
	// alignment behavior.
	RotationMode TextRotationMode
	Coords       CoordinateSpec
	OffsetX      float64
	OffsetY      float64
	// WrapWidth wraps text to this maximum display-pixel width when positive.
	WrapWidth float64
	// MultiAlignment controls per-line alignment within multiline or wrapped
	// text. Nil follows HAlign, matching Matplotlib's multialignment=None.
	MultiAlignment *TextAlign
	ClipOn         bool
	BBox           *TextBBoxOptions
	FontKey        string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties
	ParseMath      *bool
	PathEffects    []render.PathEffect
	z              float64
}

// Annotation renders text offset from a data point with an arrow.
type Annotation struct {
	ArtistRasterization
	Point   geom.Pt
	Content string

	OffsetX         float64
	OffsetY         float64
	FontSize        float64
	Color           render.Color
	ArrowColor      render.Color
	ArrowWidth      float64
	ArrowHeadSize   float64
	ArrowStyle      ArrowStyle
	ConnectionStyle ConnectionStyle
	HAlign          TextAlign
	VAlign          TextVerticalAlign
	Coords          CoordinateSpec
	FontKey         string
	BBox            *TextBBoxOptions
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties
	ParseMath      *bool
	AnnotationClip *bool
	z              float64
}

// Text adds arbitrary text positioned in data coordinates.
func (a *Axes) Text(x, y float64, text string, opts ...TextOptions) *Text {
	opt := TextOptions{
		HAlign: TextAlignLeft,
		VAlign: TextVAlignBaseline,
	}
	if len(opts) > 0 {
		opt = opts[0]
	}
	clipOn := true
	if opt.ClipOn != nil {
		clipOn = *opt.ClipOn
	}

	artist := &Text{
		Position:       geom.Pt{X: x, Y: y},
		Content:        text,
		FontSize:       opt.FontSize,
		Color:          opt.Color,
		HAlign:         opt.HAlign,
		VAlign:         opt.VAlign,
		Angle:          opt.Angle,
		RotationMode:   opt.RotationMode,
		Coords:         opt.Coords,
		OffsetX:        opt.OffsetX,
		OffsetY:        opt.OffsetY,
		WrapWidth:      opt.WrapWidth,
		MultiAlignment: cloneTextAlign(opt.MultiAlignment),
		ClipOn:         clipOn,
		BBox:           cloneTextBBoxOptions(opt.BBox),
		FontKey:        opt.FontKey,
		FontProperties: cloneFontProperties(opt.FontProperties),
		ParseMath:      cloneBool(opt.ParseMath),
		PathEffects:    cloneRenderPathEffects(opt.PathEffects),
		z:              500,
	}
	a.Add(artist)
	return artist
}

// Text adds arbitrary text positioned in figure-fraction coordinates.
func (f *Figure) Text(x, y float64, text string, opts ...TextOptions) *Text {
	if f == nil {
		return nil
	}
	opt := TextOptions{
		HAlign: TextAlignLeft,
		VAlign: TextVAlignBaseline,
		Coords: Coords(CoordFigure),
	}
	if len(opts) > 0 {
		opt = opts[0]
		opt.Coords = Coords(CoordFigure)
	}
	clipOn := true
	if opt.ClipOn != nil {
		clipOn = *opt.ClipOn
	}

	artist := &Text{
		Position:       geom.Pt{X: x, Y: y},
		Content:        text,
		FontSize:       opt.FontSize,
		Color:          opt.Color,
		HAlign:         opt.HAlign,
		VAlign:         opt.VAlign,
		Angle:          opt.Angle,
		RotationMode:   opt.RotationMode,
		Coords:         opt.Coords,
		OffsetX:        opt.OffsetX,
		OffsetY:        opt.OffsetY,
		WrapWidth:      opt.WrapWidth,
		MultiAlignment: cloneTextAlign(opt.MultiAlignment),
		ClipOn:         clipOn,
		BBox:           cloneTextBBoxOptions(opt.BBox),
		FontKey:        opt.FontKey,
		FontProperties: cloneFontProperties(opt.FontProperties),
		ParseMath:      cloneBool(opt.ParseMath),
		PathEffects:    cloneRenderPathEffects(opt.PathEffects),
		z:              500,
	}
	f.Artists = append(f.Artists, artist)
	f.zsorted = false
	return artist
}

// Annotate adds an arrow annotation pointing to a data-space point.
func (a *Axes) Annotate(text string, x, y float64, opts ...AnnotationOptions) *Annotation {
	opt := AnnotationOptions{
		OffsetX:       28,
		OffsetY:       -20,
		ArrowWidth:    1.25,
		ArrowHeadSize: 8,
	}
	defaultArrowStyle, _ := ArrowStyleFromString("-|>")
	defaultArrowStyle.HeadWidth = 0.36
	defaultConnectionStyle, _ := ConnectionStyleFromString("arc3")
	if len(opts) > 0 {
		opt = opts[0]
		if opt.OffsetX == 0 && opt.OffsetY == 0 {
			opt.OffsetX = 28
			opt.OffsetY = -20
		}
		if opt.ArrowWidth <= 0 {
			opt.ArrowWidth = 1.25
		}
		if opt.ArrowHeadSize <= 0 {
			opt.ArrowHeadSize = 8
		}
	}
	if opt.ArrowStyle.Name == "" {
		opt.ArrowStyle = defaultArrowStyle
	}
	if opt.ConnectionStyle.Name == "" {
		opt.ConnectionStyle = defaultConnectionStyle
	}

	artist := &Annotation{
		Point:           geom.Pt{X: x, Y: y},
		Content:         text,
		OffsetX:         opt.OffsetX,
		OffsetY:         opt.OffsetY,
		FontSize:        opt.FontSize,
		Color:           opt.Color,
		ArrowColor:      opt.ArrowColor,
		ArrowWidth:      opt.ArrowWidth,
		ArrowHeadSize:   opt.ArrowHeadSize,
		ArrowStyle:      opt.ArrowStyle,
		ConnectionStyle: opt.ConnectionStyle,
		HAlign:          annotationHAlign(opt),
		VAlign:          annotationVAlign(opt),
		Coords:          opt.Coords,
		FontKey:         opt.FontKey,
		BBox:            cloneTextBBoxOptions(opt.BBox),
		FontProperties:  cloneFontProperties(opt.FontProperties),
		ParseMath:       cloneBool(opt.ParseMath),
		AnnotationClip:  cloneBool(opt.AnnotationClip),
		z:               900,
	}
	a.Add(artist)
	return artist
}

// Draw renders text inside the axes clip.
func (t *Text) Draw(r render.Renderer, ctx *DrawContext) {
	if t == nil || ctx == nil {
		return
	}
	if !t.ClipOn {
		return
	}
	t.drawText(r, ctx)
}

// DrawOverlay renders unclipped text after the axes clip has been removed.
func (t *Text) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if t == nil || t.ClipOn {
		return
	}
	t.drawText(r, ctx)
}

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
	lines := wrappedTextLines(r, t.Content, fontSize, fontKey, parseMath, ctx.RC.UseTeX, t.WrapWidth)
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
			rotAnchor := textRotationAnchor(origin, layout, hAlign, vAlign, angle, t.RotationMode)
			drawTextBBoxRotated(r, origin, layout, t.BBox, ctx, fontSize, rotAnchor, t.Angle)
			if len(t.PathEffects) > 0 && drawTextPathEffects(r, content, origin, rotAnchor, fontSize, angle, textColor, fontKey, ctx.RC.UseTeX, t.PathEffects) {
				return
			}
			drawDisplayTextRotatedParseMath(rotated, content, rotAnchor, fontSize, angle, textColor, fontKey, parseMath, ctx.RC.UseTeX)
			return
		}
	}
	drawTextBBox(r, origin, layout, t.BBox, ctx, fontSize)
	if len(t.PathEffects) > 0 && drawTextPathEffects(r, content, origin, origin, fontSize, 0, textColor, fontKey, ctx.RC.UseTeX, t.PathEffects) {
		return
	}
	drawDisplayTextParseMath(textRen, content, origin, fontSize, textColor, fontKey, parseMath, ctx.RC.UseTeX)
}

func (t *Text) drawMultilineText(r render.Renderer, textRen render.TextDrawer, ctx *DrawContext, anchor geom.Pt, fontSize float64, fontKey string, parseMath bool, lines []string) {
	layouts := make([]singleLineTextLayout, len(lines))
	maxWidth := 0.0
	for i, line := range lines {
		layouts[i] = measureSingleLineTextLayoutParseMath(r, line, fontSize, fontKey, parseMath, ctx.RC.UseTeX)
		maxWidth = math.Max(maxWidth, layouts[i].Width)
	}

	lineHeight := pointsToPixels(ctx.RC, fontSize)
	lineGap := lineHeight * 0.2
	lineAdvance := lineHeight + lineGap
	blockHeight := lineHeight
	if len(lines) > 1 {
		blockHeight += lineAdvance * float64(len(lines)-1)
	}

	left := anchor.X
	hAlign, vAlign := textRotationLayoutAlignments(t.HAlign, t.VAlign, t.Angle, t.RotationMode)
	switch hAlign {
	case TextAlignCenter:
		left -= maxWidth / 2
	case TextAlignRight:
		left -= maxWidth
	}

	top := anchor.Y
	switch vAlign {
	case textLayoutVAlignCenter:
		top -= blockHeight / 2
	case textLayoutVAlignBottom:
		top -= blockHeight
	case textLayoutVAlignBaseline:
		top -= layouts[0].Ascent
	}

	if t.BBox != nil {
		rect := geom.Rect{
			Min: geom.Pt{X: left, Y: top},
			Max: geom.Pt{X: left + maxWidth, Y: top + blockHeight},
		}
		if t.Angle != 0 {
			if _, ok := r.(render.RotatedTextDrawer); ok {
				drawMultilineTextBBoxRotated(r, rect, t.BBox, ctx, fontSize, geom.Pt{
					X: left + maxWidth/2,
					Y: top + blockHeight,
				}, t.Angle)
			} else {
				drawMultilineTextBBox(r, rect, t.BBox, ctx, fontSize)
			}
		} else {
			drawMultilineTextBBox(r, rect, t.BBox, ctx, fontSize)
		}
	}

	textColor := t.ApplyArtistAlpha(resolvedTextColor(t.Color, ctx))
	lineAlign := hAlign
	if t.MultiAlignment != nil {
		lineAlign = *t.MultiAlignment
	}
	for i, line := range lines {
		if line == "" {
			continue
		}
		origin := geom.Pt{
			X: left,
			Y: top + lineAdvance*float64(i) + layouts[i].Ascent,
		}
		if layouts[i].Width < maxWidth {
			switch lineAlign {
			case TextAlignCenter:
				origin.X += (maxWidth - layouts[i].Width) / 2
			case TextAlignRight:
				origin.X += maxWidth - layouts[i].Width
			}
		}
		if t.Angle != 0 {
			if rotated, ok := r.(render.RotatedTextDrawer); ok {
				angle := t.Angle * math.Pi / 180
				rotAnchor := textRotationAnchor(origin, layouts[i], lineAlign, textLayoutVAlignBaseline, angle, t.RotationMode)
				drawDisplayTextRotatedParseMath(rotated, line, rotAnchor, fontSize, angle, textColor, fontKey, parseMath, ctx.RC.UseTeX)
				continue
			}
		}
		drawDisplayTextParseMath(textRen, line, origin, fontSize, textColor, fontKey, parseMath, ctx.RC.UseTeX)
	}
}

func textRotationLayoutAlignments(hAlign TextAlign, vAlign TextVerticalAlign, angleDeg float64, mode TextRotationMode) (TextAlign, textLayoutVerticalAlign) {
	layoutVAlign := layoutVerticalAlign(vAlign, false)
	switch mode {
	case TextRotationModeXTick:
		return xTickRotationHAlign(angleDeg, vAlign), layoutVAlign
	case TextRotationModeYTick:
		return hAlign, yTickRotationVAlign(angleDeg, hAlign)
	default:
		return hAlign, layoutVAlign
	}
}

func xTickRotationHAlign(angleDeg float64, vAlign TextVerticalAlign) TextAlign {
	angle := normalizedTextRotationAngle(angleDeg)
	anchorAtBottom := vAlign == TextVAlignBottom
	if angle <= 10 || (85 <= angle && angle <= 95) || 350 <= angle || (170 <= angle && angle <= 190) || (265 <= angle && angle <= 275) {
		return TextAlignCenter
	}
	if (10 < angle && angle < 85) || (190 < angle && angle < 265) {
		if anchorAtBottom {
			return TextAlignLeft
		}
		return TextAlignRight
	}
	if anchorAtBottom {
		return TextAlignRight
	}
	return TextAlignLeft
}

func yTickRotationVAlign(angleDeg float64, hAlign TextAlign) textLayoutVerticalAlign {
	angle := normalizedTextRotationAngle(angleDeg)
	anchorAtLeft := hAlign == TextAlignLeft
	if angle <= 10 || 350 <= angle || (170 <= angle && angle <= 190) || (80 <= angle && angle <= 100) || (260 <= angle && angle <= 280) {
		return textLayoutVAlignCenter
	}
	if (190 < angle && angle < 260) || (10 < angle && angle < 80) {
		if anchorAtLeft {
			return textLayoutVAlignBaseline
		}
		return textLayoutVAlignTop
	}
	if anchorAtLeft {
		return textLayoutVAlignTop
	}
	return textLayoutVAlignBaseline
}

func normalizedTextRotationAngle(angleDeg float64) float64 {
	angle := math.Mod(angleDeg, 360)
	if angle < 0 {
		angle += 360
	}
	return angle
}

func textRotationAnchor(origin geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign, angle float64, mode TextRotationMode) geom.Pt {
	if mode == TextRotationModeAnchor {
		pivot := tickLabelBottomCenterOffset(layout)
		return geom.Pt{
			X: origin.X + pivot.X,
			Y: origin.Y + pivot.Y,
		}
	}
	return tickLabelRotationAnchor(origin, layout, hAlign, vAlign, angle)
}

func wrappedTextLines(r render.Renderer, text string, fontSize float64, fontKey string, parseMath, useTeX bool, maxWidth float64) []string {
	paragraphs := strings.Split(text, "\n")
	if maxWidth <= 0 {
		return paragraphs
	}
	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		current := words[0]
		for _, word := range words[1:] {
			candidate := current + " " + word
			if measureSingleLineTextLayoutParseMath(r, candidate, fontSize, fontKey, parseMath, useTeX).Width <= maxWidth {
				current = candidate
				continue
			}
			lines = append(lines, current)
			current = word
		}
		lines = append(lines, current)
	}
	return lines
}

func drawTextPathEffects(r render.Renderer, text string, origin, pivot geom.Pt, size, angle float64, textColor render.Color, fontKey string, useTeX bool, effects []render.PathEffect) bool {
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
		PathEffects: cloneRenderPathEffects(effects),
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

// Bounds returns an empty rect so labels do not affect autoscaling.
func (t *Text) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the text z-order.
func (t *Text) Z() float64 { return t.z }

// Draw is a no-op because annotations render outside the axes clip via DrawOverlay.
func (a *Annotation) Draw(render.Renderer, *DrawContext) {}

// DrawOverlay renders the full annotation without the axes clip applied.
func (a *Annotation) DrawOverlay(r render.Renderer, ctx *DrawContext) {
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
	anchor := transformedPoint(ctx, a.Coords, a.Point, a.OffsetX, a.OffsetY)
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
	start := nearestPointOnRect(box, target)

	a.drawArrow(r, ctx, start, target)
	drawTextBBox(r, origin, layout, a.BBox, ctx, fontSize)

	textColor := a.ApplyArtistAlpha(resolvedTextColor(a.Color, ctx))
	drawDisplayTextParseMath(textRen, a.Content, origin, fontSize, textColor, fontKey, parseMath, ctx.RC.UseTeX)
}

// Bounds returns an empty rect so annotations do not affect autoscaling.
func (a *Annotation) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the annotation z-order.
func (a *Annotation) Z() float64 { return a.z }

func (a *Annotation) drawArrow(r render.Renderer, ctx *DrawContext, start, target geom.Pt) {
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
		MutationScale:   a.ArrowHeadSize / 0.4,
	}
	path := a.ConnectionStyle.connect(start, target, 0, 0)
	for _, part := range patch.displayParts(ctx, path) {
		if len(part.path.C) == 0 {
			continue
		}
		if part.fillable {
			patch.drawStyledPath(r, part.path, geom.Path{})
		} else {
			patch.drawStyledPath(r, geom.Path{}, part.path)
		}
	}
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

func resolvedArrowColor(arrow, text render.Color, ctx *DrawContext) render.Color {
	if arrow == (render.Color{}) {
		return resolvedTextColor(text, ctx)
	}
	if arrow.A == 0 && (arrow.R != 0 || arrow.G != 0 || arrow.B != 0) {
		arrow.A = 1
	}
	return arrow
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

func annotationPointClipped(annotationClip *bool, coords CoordinateSpec, target geom.Pt, clip geom.Rect) bool {
	if annotationClip != nil {
		return *annotationClip && !clip.ContainsInclusive(target)
	}
	return coords == Coords(CoordData) && !clip.ContainsInclusive(target)
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

func drawTextBBox(r render.Renderer, origin geom.Pt, layout singleLineTextLayout, opt *TextBBoxOptions, ctx *DrawContext, fontSize float64) {
	rect, ok := textBBoxRect(origin, layout, opt, ctx, fontSize)
	if !ok {
		return
	}
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)

	path := pixelRectPath(rect)
	if cfg.CornerRadius > 0 {
		path = roundedRectPath(rect, cfg.CornerRadius)
	}
	r.Path(path, &render.Paint{
		Fill:      cfg.FaceColor,
		Stroke:    cfg.EdgeColor,
		LineWidth: cfg.LineWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func drawTextBBoxRotated(r render.Renderer, origin geom.Pt, layout singleLineTextLayout, opt *TextBBoxOptions, ctx *DrawContext, fontSize float64, pivot geom.Pt, angleDeg float64) {
	rect, ok := textBBoxRect(origin, layout, opt, ctx, fontSize)
	if !ok {
		return
	}
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)

	path := pixelRectPath(rect)
	if cfg.CornerRadius > 0 {
		path = roundedRectPath(rect, cfg.CornerRadius)
	}
	path = rotatePathAround(path, pivot, angleDeg)
	r.Path(path, &render.Paint{
		Fill:      cfg.FaceColor,
		Stroke:    cfg.EdgeColor,
		LineWidth: cfg.LineWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func textBBoxRect(origin geom.Pt, layout singleLineTextLayout, opt *TextBBoxOptions, ctx *DrawContext, fontSize float64) (geom.Rect, bool) {
	if opt == nil {
		return geom.Rect{}, false
	}
	rect, ok := textInkRect(origin, layout)
	if !ok {
		return geom.Rect{}, false
	}
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)
	rect.Min.X -= cfg.Padding
	rect.Min.Y -= cfg.Padding
	rect.Max.X += cfg.Padding
	rect.Max.Y += cfg.Padding
	return rect, true
}

func drawMultilineTextBBox(r render.Renderer, rect geom.Rect, opt *TextBBoxOptions, ctx *DrawContext, fontSize float64) {
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)
	rect.Min.X -= cfg.Padding
	rect.Min.Y -= cfg.Padding
	rect.Max.X += cfg.Padding
	rect.Max.Y += cfg.Padding

	path := pixelRectPath(rect)
	if cfg.CornerRadius > 0 {
		path = roundedRectPath(rect, cfg.CornerRadius)
	}
	r.Path(path, &render.Paint{
		Fill:      cfg.FaceColor,
		Stroke:    cfg.EdgeColor,
		LineWidth: cfg.LineWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func drawMultilineTextBBoxRotated(r render.Renderer, rect geom.Rect, opt *TextBBoxOptions, ctx *DrawContext, fontSize float64, pivot geom.Pt, angleDeg float64) {
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)
	rect.Min.X -= cfg.Padding
	rect.Min.Y -= cfg.Padding
	rect.Max.X += cfg.Padding
	rect.Max.Y += cfg.Padding

	path := pixelRectPath(rect)
	if cfg.CornerRadius > 0 {
		path = roundedRectPath(rect, cfg.CornerRadius)
	}
	path = rotatePathAround(path, pivot, angleDeg)
	r.Path(path, &render.Paint{
		Fill:      cfg.FaceColor,
		Stroke:    cfg.EdgeColor,
		LineWidth: cfg.LineWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func resolvedTextBBoxOptions(opt TextBBoxOptions, ctx *DrawContext, fontSize float64) TextBBoxOptions {
	if opt.FaceColor == (render.Color{}) {
		opt.FaceColor = render.Color{R: 1, G: 1, B: 1, A: 1}
	} else {
		opt.FaceColor = resolvedTextBBoxColor(opt.FaceColor)
	}
	if opt.EdgeColor == (render.Color{}) {
		opt.EdgeColor = render.Color{R: 0, G: 0, B: 0, A: 1}
	} else {
		opt.EdgeColor = resolvedTextBBoxColor(opt.EdgeColor)
	}
	if opt.LineWidth <= 0 {
		opt.LineWidth = 1
		if ctx != nil {
			opt.LineWidth = pointsToPixels(ctx.RC, 1)
		}
	}
	if opt.Padding <= 0 {
		opt.Padding = 4
		if ctx != nil {
			opt.Padding = pointsToPixels(ctx.RC, 0.4*fontSize)
		}
	}
	return opt
}

func resolvedTextBBoxColor(c render.Color) render.Color {
	if c.A == 0 && (c.R != 0 || c.G != 0 || c.B != 0) {
		c.A = 1
	}
	return c
}

func annotationHAlign(opt AnnotationOptions) TextAlign {
	if opt.HAlign != TextAlignLeft {
		return opt.HAlign
	}
	if opt.OffsetX < 0 {
		return TextAlignRight
	}
	return TextAlignLeft
}

func annotationVAlign(opt AnnotationOptions) TextVerticalAlign {
	if opt.VAlign != TextVAlignBaseline {
		return opt.VAlign
	}
	if opt.OffsetY > 0 {
		return TextVAlignTop
	}
	if opt.OffsetY < 0 {
		return TextVAlignBottom
	}
	return TextVAlignMiddle
}

func alignedTextOrigin(anchor geom.Pt, metrics render.TextMetrics, hAlign TextAlign, vAlign TextVerticalAlign) geom.Pt {
	origin := geom.Pt{X: anchor.X, Y: anchor.Y}

	switch hAlign {
	case TextAlignCenter:
		origin.X -= metrics.W / 2
	case TextAlignRight:
		origin.X -= metrics.W
	}

	switch vAlign {
	case TextVAlignTop:
		origin.Y += metrics.Ascent
	case TextVAlignMiddle:
		origin.Y += (metrics.Ascent - metrics.Descent) / 2
	case TextVAlignBottom:
		origin.Y -= metrics.Descent
	}

	return origin
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

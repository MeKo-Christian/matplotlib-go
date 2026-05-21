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

// TextOptions configures a Text artist.
type TextOptions struct {
	FontSize    float64
	Color       render.Color
	HAlign      TextAlign
	VAlign      TextVerticalAlign
	Angle       float64
	Coords      CoordinateSpec
	OffsetX     float64
	OffsetY     float64
	ClipOn      *bool
	BBox        *TextBBoxOptions
	PathEffects []render.PathEffect
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
	Coords        CoordinateSpec
	OffsetX       float64
	OffsetY       float64
	FontSize      float64
	Color         render.Color
	ArrowColor    render.Color
	ArrowWidth    float64
	ArrowHeadSize float64
	HAlign        TextAlign
	VAlign        TextVerticalAlign
}

// Text renders arbitrary text at a data-space position.
type Text struct {
	ArtistRasterization
	Position geom.Pt
	Content  string

	FontSize    float64
	Color       render.Color
	HAlign      TextAlign
	VAlign      TextVerticalAlign
	Angle       float64
	Coords      CoordinateSpec
	OffsetX     float64
	OffsetY     float64
	ClipOn      bool
	BBox        *TextBBoxOptions
	PathEffects []render.PathEffect
	z           float64
}

// Annotation renders text offset from a data point with an arrow.
type Annotation struct {
	ArtistRasterization
	Point   geom.Pt
	Content string

	OffsetX       float64
	OffsetY       float64
	FontSize      float64
	Color         render.Color
	ArrowColor    render.Color
	ArrowWidth    float64
	ArrowHeadSize float64
	HAlign        TextAlign
	VAlign        TextVerticalAlign
	Coords        CoordinateSpec
	z             float64
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
		Position:    geom.Pt{X: x, Y: y},
		Content:     text,
		FontSize:    opt.FontSize,
		Color:       opt.Color,
		HAlign:      opt.HAlign,
		VAlign:      opt.VAlign,
		Angle:       opt.Angle,
		Coords:      opt.Coords,
		OffsetX:     opt.OffsetX,
		OffsetY:     opt.OffsetY,
		ClipOn:      clipOn,
		BBox:        cloneTextBBoxOptions(opt.BBox),
		PathEffects: cloneRenderPathEffects(opt.PathEffects),
		z:           500,
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
		Position:    geom.Pt{X: x, Y: y},
		Content:     text,
		FontSize:    opt.FontSize,
		Color:       opt.Color,
		HAlign:      opt.HAlign,
		VAlign:      opt.VAlign,
		Angle:       opt.Angle,
		Coords:      opt.Coords,
		OffsetX:     opt.OffsetX,
		OffsetY:     opt.OffsetY,
		ClipOn:      clipOn,
		BBox:        cloneTextBBoxOptions(opt.BBox),
		PathEffects: cloneRenderPathEffects(opt.PathEffects),
		z:           500,
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

	artist := &Annotation{
		Point:         geom.Pt{X: x, Y: y},
		Content:       text,
		OffsetX:       opt.OffsetX,
		OffsetY:       opt.OffsetY,
		FontSize:      opt.FontSize,
		Color:         opt.Color,
		ArrowColor:    opt.ArrowColor,
		ArrowWidth:    opt.ArrowWidth,
		ArrowHeadSize: opt.ArrowHeadSize,
		HAlign:        annotationHAlign(opt),
		VAlign:        annotationVAlign(opt),
		Coords:        opt.Coords,
		z:             900,
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
	anchor := transformedPoint(ctx, t.Coords, t.Position, t.OffsetX, t.OffsetY)
	if strings.Contains(t.Content, "\n") {
		t.drawMultilineText(r, textRen, ctx, anchor, fontSize)
		return
	}
	layout := measureSingleLineTextLayout(r, t.Content, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	origin := alignedSingleLineOrigin(anchor, layout, t.HAlign, layoutVerticalAlign(t.VAlign, false))
	drawTextBBox(r, origin, layout, t.BBox, ctx, fontSize)
	if t.Angle != 0 {
		if rotated, ok := r.(render.RotatedTextDrawer); ok {
			angle := t.Angle * math.Pi / 180
			rotAnchor := tickLabelRotationAnchor(origin, layout, t.HAlign, layoutVerticalAlign(t.VAlign, false), angle)
			if len(t.PathEffects) > 0 && drawTextPathEffects(r, t.Content, origin, rotAnchor, fontSize, angle, resolvedTextColor(t.Color, ctx), ctx.RC.FontKey, ctx.RC.UseTeX, t.PathEffects) {
				return
			}
			drawDisplayTextRotated(rotated, t.Content, rotAnchor, fontSize, angle, resolvedTextColor(t.Color, ctx), ctx.RC.FontKey, ctx.RC.UseTeX)
			return
		}
	}
	if len(t.PathEffects) > 0 && drawTextPathEffects(r, t.Content, origin, origin, fontSize, 0, resolvedTextColor(t.Color, ctx), ctx.RC.FontKey, ctx.RC.UseTeX, t.PathEffects) {
		return
	}
	drawDisplayText(textRen, t.Content, origin, fontSize, resolvedTextColor(t.Color, ctx), ctx.RC.FontKey, ctx.RC.UseTeX)
}

func (t *Text) drawMultilineText(r render.Renderer, textRen render.TextDrawer, ctx *DrawContext, anchor geom.Pt, fontSize float64) {
	lines := strings.Split(t.Content, "\n")
	layouts := make([]singleLineTextLayout, len(lines))
	maxWidth := 0.0
	for i, line := range lines {
		layouts[i] = measureSingleLineTextLayout(r, line, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
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
	switch t.HAlign {
	case TextAlignCenter:
		left -= maxWidth / 2
	case TextAlignRight:
		left -= maxWidth
	}

	top := anchor.Y
	switch layoutVerticalAlign(t.VAlign, false) {
	case textLayoutVAlignCenter:
		top -= blockHeight / 2
	case textLayoutVAlignBottom:
		top -= blockHeight
	case textLayoutVAlignBaseline:
		top -= layouts[0].Ascent
	}

	if t.BBox != nil {
		drawMultilineTextBBox(r, geom.Rect{
			Min: geom.Pt{X: left, Y: top},
			Max: geom.Pt{X: left + maxWidth, Y: top + blockHeight},
		}, t.BBox, ctx, fontSize)
	}

	textColor := resolvedTextColor(t.Color, ctx)
	for i, line := range lines {
		if line == "" {
			continue
		}
		origin := geom.Pt{
			X: left,
			Y: top + lineAdvance*float64(i) + layouts[i].Ascent,
		}
		if layouts[i].Width < maxWidth {
			switch t.HAlign {
			case TextAlignCenter:
				origin.X += (maxWidth - layouts[i].Width) / 2
			case TextAlignRight:
				origin.X += maxWidth - layouts[i].Width
			}
		}
		drawDisplayText(textRen, line, origin, fontSize, textColor, ctx.RC.FontKey, ctx.RC.UseTeX)
	}
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
	target := transformedPoint(ctx, a.Coords, a.Point, 0, 0)
	anchor := transformedPoint(ctx, a.Coords, a.Point, a.OffsetX, a.OffsetY)
	layout := measureSingleLineTextLayout(r, a.Content, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	origin := alignedSingleLineOrigin(anchor, layout, a.HAlign, layoutVerticalAlign(a.VAlign, false))
	box, ok := textInkRect(origin, layout)
	if !ok {
		box = geom.Rect{
			Min: geom.Pt{X: origin.X, Y: origin.Y - layout.Ascent},
			Max: geom.Pt{X: origin.X + layout.Width, Y: origin.Y + layout.Descent},
		}
	}
	start := nearestPointOnRect(box, target)

	linePaint := render.Paint{
		Stroke:    defaultArrowColor(a.ArrowColor, a.Color),
		LineWidth: a.ArrowWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	r.Path(pixelLinePath(start, target), &linePaint)

	head := annotationHeadPath(start, target, a.ArrowHeadSize)
	if len(head.C) > 0 {
		r.Path(head, &render.Paint{
			Fill:      defaultArrowColor(a.ArrowColor, a.Color),
			Stroke:    resolvedArrowColor(a.ArrowColor, a.Color, ctx),
			LineWidth: a.ArrowWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		})
	}

	drawDisplayText(textRen, a.Content, origin, fontSize, resolvedTextColor(a.Color, ctx), ctx.RC.FontKey, ctx.RC.UseTeX)
}

// Bounds returns an empty rect so annotations do not affect autoscaling.
func (a *Annotation) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the annotation z-order.
func (a *Annotation) Z() float64 { return a.z }

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

func defaultArrowColor(arrow, text render.Color) render.Color {
	return resolvedArrowColor(arrow, text, nil)
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

func cloneTextBBoxOptions(opt *TextBBoxOptions) *TextBBoxOptions {
	if opt == nil {
		return nil
	}
	cloned := *opt
	return &cloned
}

func drawTextBBox(r render.Renderer, origin geom.Pt, layout singleLineTextLayout, opt *TextBBoxOptions, ctx *DrawContext, fontSize float64) {
	if opt == nil {
		return
	}
	rect, ok := textInkRect(origin, layout)
	if !ok {
		return
	}
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

func annotationHeadPath(start, tip geom.Pt, size float64) geom.Path {
	dx := tip.X - start.X
	dy := tip.Y - start.Y
	length := math.Hypot(dx, dy)
	if length == 0 || size <= 0 {
		return geom.Path{}
	}

	ux := dx / length
	uy := dy / length
	base := geom.Pt{X: tip.X - ux*size, Y: tip.Y - uy*size}
	perpX := -uy * size * 0.45
	perpY := ux * size * 0.45

	return geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{
			tip,
			{X: base.X + perpX, Y: base.Y + perpY},
			{X: base.X - perpX, Y: base.Y - perpY},
		},
	}
}

func pixelLinePath(p1, p2 geom.Pt) geom.Path {
	return geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{p1, p2},
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

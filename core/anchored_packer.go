package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// PackOrientation controls whether an AnchoredPacker lays children out in a
// horizontal or vertical row.
type PackOrientation uint8

const (
	// PackHorizontal lays children out left-to-right.
	PackHorizontal PackOrientation = iota
	// PackVertical lays children out top-to-bottom.
	PackVertical
)

// PackAlignment controls child alignment along the packer's cross axis.
type PackAlignment uint8

const (
	// PackAlignDefault keeps Matplotlib's packer default alignment.
	PackAlignDefault PackAlignment = iota
	// PackAlignStart aligns children to the top or left edge.
	PackAlignStart
	// PackAlignCenter centers children along the cross axis.
	PackAlignCenter
	// PackAlignEnd aligns children to the bottom or right edge.
	PackAlignEnd
)

// AnchoredPackerOptions configures an anchored horizontal or vertical packer.
type AnchoredPackerOptions struct {
	Location LegendLocation
	Locator  AnchoredBoxLocator
	Padding  float64
	Inset    float64
	Sep      float64
	FrameOn  *bool
	Align    PackAlignment

	BackgroundColor render.Color
	BorderColor     render.Color
	BorderWidth     float64
	TextColor       render.Color
	FontSize        float64
	FontKey         string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties
}

// PackedTextOptions configures a text child added to an AnchoredPacker.
type PackedTextOptions struct {
	TextColor render.Color
	FontSize  float64
	FontKey   string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties
}

// AnchoredPacker is an anchored HPacker/VPacker-style container for fixed-size
// drawing areas, text children, and image children.
type AnchoredPacker struct {
	Orientation PackOrientation

	Location LegendLocation
	Locator  AnchoredBoxLocator
	Padding  float64
	Inset    float64
	Sep      float64
	FrameOn  bool
	Align    PackAlignment

	BackgroundColor render.Color
	BorderColor     render.Color
	BorderWidth     float64
	TextColor       render.Color
	FontSize        float64
	FontKey         string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties

	children []packedBoxChild
	z        float64
}

type packedBoxChild interface {
	size(render.Renderer, *DrawContext, *AnchoredPacker) (geom.Pt, bool)
	draw(render.Renderer, *DrawContext, geom.Rect, *AnchoredPacker)
}

// AddAnchoredPacker appends an anchored horizontal or vertical packer inside
// an axes.
func (a *Axes) AddAnchoredPacker(orientation PackOrientation, opts ...AnchoredPackerOptions) *AnchoredPacker {
	rc := style.CurrentDefaults()
	if a != nil {
		rc = a.resolvedRC()
	}
	packer := newAnchoredPacker(orientation, rc, opts...)
	if a != nil {
		a.Add(packer)
	}
	return packer
}

func newAnchoredPacker(orientation PackOrientation, rc style.RC, opts ...AnchoredPackerOptions) *AnchoredPacker {
	cfg := AnchoredPackerOptions{
		Location:        LegendUpperRight,
		Padding:         -1,
		Inset:           -1,
		Sep:             -1,
		Align:           PackAlignCenter,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		TextColor:       rc.LegendTextColor,
		BorderWidth:     1,
	}
	frameOn := true
	cfg.FrameOn = &frameOn
	if len(opts) > 0 {
		opt := opts[0]
		cfg.Location = opt.Location
		cfg.Locator = opt.Locator
		if opt.Padding >= 0 {
			cfg.Padding = opt.Padding
		}
		if opt.Inset >= 0 {
			cfg.Inset = opt.Inset
		}
		if opt.Sep >= 0 {
			cfg.Sep = opt.Sep
		}
		if opt.FrameOn != nil {
			cfg.FrameOn = cloneBool(opt.FrameOn)
		}
		if opt.Align != PackAlignDefault {
			cfg.Align = opt.Align
		}
		if opt.BackgroundColor != (render.Color{}) {
			cfg.BackgroundColor = opt.BackgroundColor
		}
		if opt.BorderColor != (render.Color{}) {
			cfg.BorderColor = opt.BorderColor
		}
		if opt.BorderWidth > 0 {
			cfg.BorderWidth = opt.BorderWidth
		}
		if opt.TextColor != (render.Color{}) {
			cfg.TextColor = opt.TextColor
		}
		if opt.FontSize > 0 {
			cfg.FontSize = opt.FontSize
		}
		cfg.FontKey = opt.FontKey
		cfg.FontProperties = cloneFontProperties(opt.FontProperties)
	}
	return &AnchoredPacker{
		Orientation:     orientation,
		Location:        cfg.Location,
		Locator:         cfg.Locator,
		Padding:         cfg.Padding,
		Inset:           cfg.Inset,
		Sep:             cfg.Sep,
		FrameOn:         cfg.FrameOn == nil || *cfg.FrameOn,
		Align:           cfg.Align,
		BackgroundColor: cfg.BackgroundColor,
		BorderColor:     cfg.BorderColor,
		BorderWidth:     cfg.BorderWidth,
		TextColor:       cfg.TextColor,
		FontSize:        cfg.FontSize,
		FontKey:         cfg.FontKey,
		FontProperties:  cloneFontProperties(cfg.FontProperties),
		z:               950,
	}
}

// AddDrawingArea appends a fixed-size DrawingArea-style child and returns it so
// local paths can be added.
func (a *AnchoredPacker) AddDrawingArea(width, height float64) *PackedDrawingArea {
	if a == nil {
		return nil
	}
	child := &PackedDrawingArea{Width: width, Height: height}
	a.children = append(a.children, child)
	return child
}

// AddText appends a text-area child.
func (a *AnchoredPacker) AddText(text string, opts ...PackedTextOptions) *AnchoredPacker {
	if a == nil {
		return nil
	}
	child := &packedText{Content: text}
	if len(opts) > 0 {
		opt := opts[0]
		child.TextColor = opt.TextColor
		child.FontSize = opt.FontSize
		child.FontKey = opt.FontKey
		child.FontProperties = cloneFontProperties(opt.FontProperties)
	}
	a.children = append(a.children, child)
	return a
}

// AddImage appends an OffsetImage-style child. Non-positive zoom values use 1.
func (a *AnchoredPacker) AddImage(img render.Image, zoom float64) *AnchoredPacker {
	if a == nil {
		return nil
	}
	a.children = append(a.children, &packedImage{Image: img, Zoom: zoom})
	return a
}

// Draw renders the packed anchored box.
func (a *AnchoredPacker) Draw(r render.Renderer, ctx *DrawContext) {
	if a == nil || r == nil || ctx == nil {
		return
	}
	layout := a.layout(r, ctx)
	if layout.empty {
		return
	}
	if a.FrameOn {
		r.Path(pixelRectPath(layout.frame), &render.Paint{
			Fill:      a.BackgroundColor,
			Stroke:    a.BorderColor,
			LineWidth: pointsToPixels(ctx.RC, a.BorderWidth),
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Snap:      render.SnapAuto,
		})
	}
	for i, child := range a.children {
		child.draw(r, ctx, layout.children[i], a)
	}
}

// Bounds returns an empty rect so anchored packers do not affect autoscaling.
func (a *AnchoredPacker) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the anchored packer z-order.
func (a *AnchoredPacker) Z() float64 { return a.z }

func (a *AnchoredPacker) boxRect(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	layout := a.layout(r, ctx)
	if layout.empty {
		return geom.Rect{}, false
	}
	return layout.frame, true
}

type anchoredPackerLayout struct {
	frame    geom.Rect
	content  geom.Rect
	children []geom.Rect
	empty    bool
}

func (a *AnchoredPacker) layout(r render.Renderer, ctx *DrawContext) anchoredPackerLayout {
	if a == nil || len(a.children) == 0 {
		return anchoredPackerLayout{empty: true}
	}
	sizes := make([]geom.Pt, 0, len(a.children))
	for _, child := range a.children {
		size, ok := child.size(r, ctx, a)
		if !ok || size.X < 0 || size.Y < 0 {
			return anchoredPackerLayout{empty: true}
		}
		sizes = append(sizes, size)
	}
	contentSize := a.packedSize(sizes, ctx)
	padding := a.resolvedPadding(ctx)
	frame := resolveAnchoredBoxRect(a.Locator, ctx.Clip, contentSize.X+padding*2, contentSize.Y+padding*2, a.Location, a.resolvedInset(ctx))
	content := geom.Rect{
		Min: geom.Pt{X: frame.Min.X + padding, Y: frame.Min.Y + padding},
		Max: geom.Pt{X: frame.Max.X - padding, Y: frame.Max.Y - padding},
	}
	return anchoredPackerLayout{
		frame:    frame,
		content:  content,
		children: a.childRects(content, sizes, ctx),
	}
}

func (a *AnchoredPacker) packedSize(sizes []geom.Pt, ctx *DrawContext) geom.Pt {
	if len(sizes) == 0 {
		return geom.Pt{}
	}
	sep := a.resolvedSep(ctx)
	var out geom.Pt
	for i, size := range sizes {
		if i > 0 {
			if a.Orientation == PackVertical {
				out.Y += sep
			} else {
				out.X += sep
			}
		}
		if a.Orientation == PackVertical {
			out.Y += size.Y
			if size.X > out.X {
				out.X = size.X
			}
		} else {
			out.X += size.X
			if size.Y > out.Y {
				out.Y = size.Y
			}
		}
	}
	return out
}

func (a *AnchoredPacker) childRects(content geom.Rect, sizes []geom.Pt, ctx *DrawContext) []geom.Rect {
	rects := make([]geom.Rect, 0, len(sizes))
	sep := a.resolvedSep(ctx)
	cursorX := content.Min.X
	// Display space is y-up: vertical packing runs top-to-bottom, so the first
	// child starts at the top edge (content.Max.Y) and the cursor moves downward.
	cursorY := content.Max.Y
	for _, size := range sizes {
		var min geom.Pt
		if a.Orientation == PackVertical {
			// Cross axis is X: Start=left, End=right (no y-flip).
			min = geom.Pt{X: alignedPackMin(content.Min.X, content.W(), size.X, a.Align), Y: cursorY - size.Y}
			cursorY -= size.Y + sep
		} else {
			// Cross axis is Y under y-up: Start=top, End=bottom (flipped).
			min = geom.Pt{X: cursorX, Y: alignedPackMinY(content.Min.Y, content.H(), size.Y, a.Align)}
			cursorX += size.X + sep
		}
		rects = append(rects, geom.Rect{
			Min: min,
			Max: geom.Pt{X: min.X + size.X, Y: min.Y + size.Y},
		})
	}
	return rects
}

// alignedPackMin positions a child along an axis where Start is the low edge
// (left, or bottom under y-up).
func alignedPackMin(start, span, childSpan float64, align PackAlignment) float64 {
	switch align {
	case PackAlignStart:
		return start
	case PackAlignEnd:
		return start + span - childSpan
	case PackAlignCenter, PackAlignDefault:
		return start + (span-childSpan)/2
	default:
		return start + (span-childSpan)/2
	}
}

// alignedPackMinY positions a child on the vertical cross axis where, under the
// y-up contract, Start means the top (high Y) and End means the bottom.
func alignedPackMinY(start, span, childSpan float64, align PackAlignment) float64 {
	switch align {
	case PackAlignEnd:
		return start
	case PackAlignCenter, PackAlignDefault:
		return start + (span-childSpan)/2
	case PackAlignStart:
		return start + span - childSpan
	default:
		return start + (span-childSpan)/2
	}
}

func (a *AnchoredPacker) resolvedPadding(ctx *DrawContext) float64 {
	if a != nil && a.Padding >= 0 {
		return a.Padding
	}
	return pointsToPixels(ctx.RC, 0.1*ctx.RC.LegendSize())
}

func (a *AnchoredPacker) resolvedInset(ctx *DrawContext) float64 {
	if a != nil && a.Inset >= 0 {
		return a.Inset
	}
	return pointsToPixels(ctx.RC, 0.1*ctx.RC.LegendSize())
}

func (a *AnchoredPacker) resolvedSep(ctx *DrawContext) float64 {
	if a != nil && a.Sep >= 0 {
		return a.Sep
	}
	return pointsToPixels(ctx.RC, 0.2*ctx.RC.LegendSize())
}

// PackedDrawingArea is a fixed-size child whose paths use Matplotlib
// DrawingArea coordinates: (0,0) is the lower-left corner.
type PackedDrawingArea struct {
	Width  float64
	Height float64

	paths []anchoredDrawingAreaPath
}

// AddPath appends a local display-space path to the drawing-area child.
func (a *PackedDrawingArea) AddPath(path geom.Path, paint render.Paint) *PackedDrawingArea {
	if a == nil {
		return nil
	}
	a.paths = append(a.paths, anchoredDrawingAreaPath{
		path:  clonePath(path),
		paint: clonePaint(paint),
	})
	return a
}

func (a *PackedDrawingArea) size(_ render.Renderer, ctx *DrawContext, _ *AnchoredPacker) (geom.Pt, bool) {
	if a == nil || a.Width < 0 || a.Height < 0 {
		return geom.Pt{}, false
	}
	scale := drawingAreaScale(ctx)
	return geom.Pt{X: a.Width * scale, Y: a.Height * scale}, true
}

func (a *PackedDrawingArea) draw(r render.Renderer, ctx *DrawContext, box geom.Rect, _ *AnchoredPacker) {
	if a == nil {
		return
	}
	scale := drawingAreaScale(ctx)
	for _, child := range a.paths {
		paint := child.paint
		r.Path(localDrawingAreaPath(child.path, box, scale), &paint)
	}
}

type packedImage struct {
	Image render.Image
	Zoom  float64
}

func (i *packedImage) size(_ render.Renderer, ctx *DrawContext, _ *AnchoredPacker) (geom.Pt, bool) {
	if i == nil || i.Image == nil {
		return geom.Pt{}, false
	}
	width, height := i.Image.Size()
	zoom := i.resolvedZoom()
	scale := drawingAreaScale(ctx)
	return geom.Pt{X: float64(width) * zoom * scale, Y: float64(height) * zoom * scale}, true
}

func (i *packedImage) draw(r render.Renderer, _ *DrawContext, box geom.Rect, _ *AnchoredPacker) {
	if i == nil || i.Image == nil {
		return
	}
	img := offsetImageWithMatplotlibDefaults(i.Image)
	if drawer, ok := r.(render.BboxImageDrawer); ok && drawer.DrawBboxImage(img, box) {
		return
	}
	r.DrawImage(img, box)
}

func (i *packedImage) resolvedZoom() float64 {
	if i != nil && i.Zoom > 0 {
		return i.Zoom
	}
	return 1
}

func drawingAreaScale(ctx *DrawContext) float64 {
	if ctx == nil {
		return 1
	}
	scale := pointsToPixels(ctx.RC, 1)
	if scale <= 0 {
		return 1
	}
	return scale
}

type packedText struct {
	Content string

	TextColor render.Color
	FontSize  float64
	FontKey   string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties
}

func (t *packedText) size(r render.Renderer, ctx *DrawContext, parent *AnchoredPacker) (geom.Pt, bool) {
	if t == nil || displayTextIsEmpty(t.Content) {
		return geom.Pt{}, true
	}
	if _, ok := r.(render.TextDrawer); !ok {
		return geom.Pt{}, false
	}
	layout := t.layout(r, ctx, parent)
	return geom.Pt{X: layout.Width, Y: layout.Height}, true
}

func (t *packedText) draw(r render.Renderer, ctx *DrawContext, box geom.Rect, parent *AnchoredPacker) {
	if t == nil || displayTextIsEmpty(t.Content) {
		return
	}
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}
	layout := t.layout(r, ctx, parent)
	origin := alignedSingleLineOrigin(rectCenter(box), layout, TextAlignCenter, textLayoutVAlignCenter)
	drawDisplayText(textRen, t.Content, origin, t.resolvedFontSize(ctx, parent), t.resolvedTextColor(ctx, parent), t.resolvedFontKey(ctx, parent), ctx.RC.UseTeX)
}

func (t *packedText) layout(r render.Renderer, ctx *DrawContext, parent *AnchoredPacker) singleLineTextLayout {
	return measureSingleLineTextLayout(r, t.Content, t.resolvedFontSize(ctx, parent), t.resolvedFontKey(ctx, parent), ctx.RC.UseTeX)
}

func (t *packedText) resolvedFontSize(ctx *DrawContext, parent *AnchoredPacker) float64 {
	if t != nil && t.FontSize > 0 {
		return t.FontSize
	}
	if parent != nil && parent.FontSize > 0 {
		return parent.FontSize
	}
	return resolvedFontSize(0, ctx)
}

func (t *packedText) resolvedFontKey(ctx *DrawContext, parent *AnchoredPacker) string {
	if t != nil && (t.FontKey != "" || t.FontProperties != nil) {
		return resolvedTextFontKey(t.FontKey, t.FontProperties, ctx)
	}
	if parent != nil && (parent.FontKey != "" || parent.FontProperties != nil) {
		return resolvedTextFontKey(parent.FontKey, parent.FontProperties, ctx)
	}
	return ctx.RC.FontKey
}

func (t *packedText) resolvedTextColor(ctx *DrawContext, parent *AnchoredPacker) render.Color {
	if t != nil && t.TextColor != (render.Color{}) {
		return t.TextColor
	}
	if parent != nil && parent.TextColor != (render.Color{}) {
		return parent.TextColor
	}
	return resolvedTextColor(render.Color{}, ctx)
}

package core

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// AnchoredTextOptions configures an anchored annotation box.
type AnchoredTextOptions struct {
	Location        LegendLocation
	Locator         AnchoredBoxLocator
	Padding         float64
	Inset           float64
	RowGap          float64
	BoxPadding      float64
	CornerRadius    float64
	BackgroundColor render.Color
	BorderColor     render.Color
	TextColor       render.Color
	BorderWidth     float64
	FontSize        float64
	// TextAlign controls horizontal alignment of multi-line text inside the
	// box; mirrors matplotlib's text(..., ha=...) parameter. Defaults to
	// TextAlignLeft.
	TextAlign TextAlign
}

// AnchoredTextBox draws a boxed block of text anchored to an axes or figure corner.
type AnchoredTextBox struct {
	Content string

	Location        LegendLocation
	Locator         AnchoredBoxLocator
	Padding         float64
	Inset           float64
	RowGap          float64
	BoxPadding      float64
	CornerRadius    float64
	BackgroundColor render.Color
	BorderColor     render.Color
	TextColor       render.Color
	BorderWidth     float64
	FontSize        float64
	TextAlign       TextAlign
	z               float64
}

// AddAnchoredText appends an anchored text box inside an axes.
func (a *Axes) AddAnchoredText(text string, opts ...AnchoredTextOptions) *AnchoredTextBox {
	box := newAnchoredTextBox(text, a.resolvedRC(), opts...)
	a.Add(box)
	return box
}

// AddAnchoredText appends a figure-level anchored text box.
func (f *Figure) AddAnchoredText(text string, opts ...AnchoredTextOptions) *AnchoredTextBox {
	rc := style.CurrentDefaults()
	if f != nil {
		rc = f.RC
	}
	box := newAnchoredTextBox(text, rc, opts...)
	if f != nil {
		f.Add(box)
	}
	return box
}

func newAnchoredTextBox(text string, rc style.RC, opts ...AnchoredTextOptions) *AnchoredTextBox {
	cfg := AnchoredTextOptions{
		Location:        LegendUpperLeft,
		Padding:         -1,
		Inset:           -1,
		RowGap:          -1,
		CornerRadius:    0,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		TextColor:       rc.LegendTextColor,
		BorderWidth:     1,
	}
	if len(opts) > 0 {
		opt := opts[0]
		cfg.Location = opt.Location
		cfg.Locator = opt.Locator
		if opt.Padding > 0 {
			cfg.Padding = opt.Padding
		}
		if opt.Inset > 0 {
			cfg.Inset = opt.Inset
		}
		if opt.RowGap > 0 {
			cfg.RowGap = opt.RowGap
		}
		if opt.BoxPadding > 0 {
			cfg.BoxPadding = opt.BoxPadding
		}
		if opt.CornerRadius > 0 {
			cfg.CornerRadius = opt.CornerRadius
		}
		if opt.BackgroundColor != (render.Color{}) {
			cfg.BackgroundColor = opt.BackgroundColor
		}
		if opt.BorderColor != (render.Color{}) {
			cfg.BorderColor = opt.BorderColor
		}
		if opt.TextColor != (render.Color{}) {
			cfg.TextColor = opt.TextColor
		}
		if opt.BorderWidth > 0 {
			cfg.BorderWidth = opt.BorderWidth
		}
		if opt.FontSize > 0 {
			cfg.FontSize = opt.FontSize
		}
		cfg.TextAlign = opt.TextAlign
	}

	return &AnchoredTextBox{
		Content:         text,
		Location:        cfg.Location,
		Locator:         cfg.Locator,
		Padding:         cfg.Padding,
		Inset:           cfg.Inset,
		RowGap:          cfg.RowGap,
		BoxPadding:      cfg.BoxPadding,
		CornerRadius:    cfg.CornerRadius,
		BackgroundColor: cfg.BackgroundColor,
		BorderColor:     cfg.BorderColor,
		TextColor:       cfg.TextColor,
		BorderWidth:     cfg.BorderWidth,
		FontSize:        cfg.FontSize,
		TextAlign:       cfg.TextAlign,
		z:               950,
	}
}

// Draw renders the anchored text box.
func (a *AnchoredTextBox) Draw(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil {
		return
	}
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	lines := strings.Split(a.Content, "\n")
	if len(lines) == 0 {
		return
	}

	layouts := make([]singleLineTextLayout, len(lines))
	fontSize := resolvedFontSize(a.FontSize, ctx)
	for i, line := range lines {
		layouts[i] = measureSingleLineTextLayout(r, line, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	}

	boxLayout := a.layout(r, ctx, layouts, fontSize)
	boxPath := snappedPixelRectPath(boxLayout.patchBox)
	if a.CornerRadius > 0 {
		boxPath = roundedRectPath(boxLayout.patchBox, a.CornerRadius)
	}
	r.Path(boxPath, &render.Paint{
		Fill:      a.BackgroundColor,
		Stroke:    a.BorderColor,
		LineWidth: a.BorderWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})

	y := boxLayout.contentBox.Max.Y - boxLayout.padding
	leftX := boxLayout.contentBox.Min.X + boxLayout.padding
	rightX := boxLayout.contentBox.Max.X - boxLayout.padding
	for i, line := range lines {
		layout := layouts[i]
		if line == "" {
			y -= boxLayout.lineAdvance
			continue
		}
		var anchorX float64
		switch a.TextAlign {
		case TextAlignRight:
			anchorX = rightX
		case TextAlignCenter:
			anchorX = (leftX + rightX) / 2
		default:
			anchorX = leftX
		}
		drawDisplayText(
			textRen,
			line,
			alignedSingleLineOrigin(geom.Pt{X: anchorX, Y: y}, layout, a.TextAlign, textLayoutVAlignTop),
			fontSize,
			resolvedTextColor(a.TextColor, ctx),
			ctx.RC.FontKey,
			ctx.RC.UseTeX,
		)
		y -= boxLayout.lineAdvance
	}
}

// Bounds returns an empty rect so anchored text boxes do not affect autoscaling.
func (a *AnchoredTextBox) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the anchored text box z-order.
func (a *AnchoredTextBox) Z() float64 { return a.z }

// SetLocator overrides the anchored-box placement strategy for this text box.
func (a *AnchoredTextBox) SetLocator(locator AnchoredBoxLocator) {
	if a == nil {
		return
	}
	a.Locator = locator
}

func (a *AnchoredTextBox) boxRect(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if a == nil || r == nil || ctx == nil {
		return geom.Rect{}, false
	}

	lines := strings.Split(a.Content, "\n")
	if len(lines) == 0 {
		return geom.Rect{}, false
	}

	fontSize := resolvedFontSize(a.FontSize, ctx)
	layouts := make([]singleLineTextLayout, len(lines))
	for i, line := range lines {
		layouts[i] = measureSingleLineTextLayout(r, line, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	}

	return a.layout(r, ctx, layouts, fontSize).patchBox, true
}

type anchoredTextLayout struct {
	contentBox  geom.Rect
	patchBox    geom.Rect
	padding     float64
	lineAdvance float64
}

func (a *AnchoredTextBox) layout(_ render.Renderer, ctx *DrawContext, layouts []singleLineTextLayout, fontSize float64) anchoredTextLayout {
	padding := a.resolvedPadding(fontSize, ctx)
	inset := a.resolvedInset(fontSize, ctx)
	boxPadding := a.resolvedBoxPadding()
	maxWidth := 0.0
	for _, layout := range layouts {
		if layout.Width > maxWidth {
			maxWidth = layout.Width
		}
	}
	contentHeight := 0.0
	lineHeight := anchoredTextLineHeight(layouts, fontSize, ctx)
	lineAdvance := lineHeight + a.resolvedRowGapForLineHeight(lineHeight, fontSize, ctx)
	if len(layouts) > 0 {
		contentHeight = lineHeight + lineAdvance*float64(len(layouts)-1)
	}
	contentBox := resolveAnchoredBoxRect(a.Locator, ctx.Clip, maxWidth+padding*2, contentHeight+padding*2, a.Location, inset)
	return anchoredTextLayout{
		contentBox:  contentBox,
		patchBox:    expandAnchoredRect(contentBox, boxPadding),
		padding:     padding,
		lineAdvance: lineAdvance,
	}
}

func (a *AnchoredTextBox) resolvedPadding(fontSize float64, ctx *DrawContext) float64 {
	if a != nil && a.Padding >= 0 {
		return a.Padding
	}
	return pointsToPixels(ctx.RC, 0.4*fontSize)
}

func (a *AnchoredTextBox) resolvedInset(fontSize float64, ctx *DrawContext) float64 {
	if a != nil && a.Inset >= 0 {
		return a.Inset
	}
	return pointsToPixels(ctx.RC, 0.5*fontSize)
}

func (a *AnchoredTextBox) resolvedRowGap(fontSize float64, ctx *DrawContext) float64 {
	if a != nil && a.RowGap >= 0 {
		return a.RowGap
	}
	return 0.2 * a.lineHeight(fontSize, ctx)
}

func (a *AnchoredTextBox) resolvedRowGapForLineHeight(lineHeight, fontSize float64, ctx *DrawContext) float64 {
	if a != nil && a.RowGap >= 0 {
		return a.RowGap
	}
	if lineHeight <= 0 {
		lineHeight = a.lineHeight(fontSize, ctx)
	}
	return 0.2 * lineHeight
}

func (a *AnchoredTextBox) resolvedBoxPadding() float64 {
	if a == nil || a.BoxPadding <= 0 {
		return 0
	}
	return a.BoxPadding
}

func (a *AnchoredTextBox) lineHeight(fontSize float64, ctx *DrawContext) float64 {
	return pointsToPixels(ctx.RC, fontSize)
}

func (a *AnchoredTextBox) lineAdvance(fontSize float64, ctx *DrawContext) float64 {
	return a.lineHeight(fontSize, ctx) + a.resolvedRowGap(fontSize, ctx)
}

func anchoredTextLineHeight(layouts []singleLineTextLayout, fontSize float64, ctx *DrawContext) float64 {
	height := 0.0
	for _, layout := range layouts {
		lineHeight := layout.RunAscent + layout.RunDescent
		if lineHeight <= 0 {
			lineHeight = layout.Height
		}
		if lineHeight > height {
			height = lineHeight
		}
	}
	if height <= 0 && ctx != nil {
		height = pointsToPixels(ctx.RC, fontSize)
	}
	return height
}

func expandAnchoredRect(r geom.Rect, pad float64) geom.Rect {
	if pad <= 0 {
		return r
	}
	return geom.Rect{
		Min: geom.Pt{X: r.Min.X - pad, Y: r.Min.Y - pad},
		Max: geom.Pt{X: r.Max.X + pad, Y: r.Max.Y + pad},
	}
}

func anchoredBoxRect(clip geom.Rect, width, height float64, location LegendLocation, inset float64) geom.Rect {
	var minX, minY float64
	switch location {
	case LegendUpperLeft:
		minX = clip.Min.X + inset
		minY = clip.Max.Y - inset - height
	case LegendLowerRight:
		minX = clip.Max.X - inset - width
		minY = clip.Min.Y + inset
	case LegendLowerLeft:
		minX = clip.Min.X + inset
		minY = clip.Min.Y + inset
	case LegendRight, LegendCenterRight:
		minX = clip.Max.X - inset - width
		minY = clip.Min.Y + (clip.H()-height)/2
	case LegendCenterLeft:
		minX = clip.Min.X + inset
		minY = clip.Min.Y + (clip.H()-height)/2
	case LegendLowerCenter:
		minX = clip.Min.X + (clip.W()-width)/2
		minY = clip.Min.Y + inset
	case LegendUpperCenter:
		minX = clip.Min.X + (clip.W()-width)/2
		minY = clip.Max.Y - inset - height
	case LegendCenter:
		minX = clip.Min.X + (clip.W()-width)/2
		minY = clip.Min.Y + (clip.H()-height)/2
	default:
		minX = clip.Max.X - inset - width
		minY = clip.Max.Y - inset - height
	}
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: minX + width, Y: minY + height},
	}
}

package core

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// AnchoredTextOptions configures an anchored annotation box.
type AnchoredTextOptions struct {
	Location LegendLocation
	Locator  AnchoredBoxLocator
	Padding  float64
	Inset    float64
	// Deprecated: RowGap no longer affects line spacing. Multi-line text now
	// follows matplotlib's Text._get_layout (linespacing 1.2 × font line
	// height); the field is retained only for API compatibility.
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

	Location LegendLocation
	Locator  AnchoredBoxLocator
	Padding  float64
	Inset    float64
	// Deprecated: RowGap no longer affects line spacing (see AnchoredTextOptions.RowGap).
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
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) AddAnchoredText(text string, opt AnchoredTextOptions) *AnchoredTextBox {
	box := newAnchoredTextBox(text, a.resolvedRC(), opt)
	a.Add(box)
	return box
}

// AddAnchoredText appends a figure-level anchored text box.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (f *Figure) AddAnchoredText(text string, opt AnchoredTextOptions) *AnchoredTextBox {
	rc := style.CurrentDefaults()
	if f != nil {
		rc = f.RC
	}
	box := newAnchoredTextBox(text, rc, opt)
	if f != nil {
		f.Add(box)
	}
	return box
}

//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func newAnchoredTextBox(text string, rc style.RC, opt AnchoredTextOptions) *AnchoredTextBox {
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

// Draw is a no-op because anchored offset boxes render outside the axes clip.
func (a *AnchoredTextBox) Draw(render.Renderer, *DrawContext) {}

// DrawOverlay renders the anchored text box without the axes clip applied.
func (a *AnchoredTextBox) DrawOverlay(r render.Renderer, ctx *DrawContext) {
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

	fontSize := resolvedFontSize(a.FontSize, ctx)
	boxLayout := a.layout(r, ctx, lines, fontSize)
	boxPath := pixelRectPath(boxLayout.patchBox)
	snap := render.SnapAuto
	if a.CornerRadius > 0 {
		boxPath = roundedRectPath(boxLayout.patchBox, a.CornerRadius)
		snap = render.SnapOff
	}
	r.Path(boxPath, &render.Paint{
		Fill:      a.BackgroundColor,
		Stroke:    a.BorderColor,
		LineWidth: pointsToPixels(ctx.RC, a.BorderWidth),
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
		Snap:      snap,
	})

	// Draw each text line at its matplotlib baseline (Text._get_layout port via
	// measureMultilineTextBlock). block.baselineYs are relative to the inner top
	// edge of the content box (y=0 at top, negative downward in display space).
	top := boxLayout.contentBox.Max.Y - boxLayout.padding
	leftX := boxLayout.contentBox.Min.X + boxLayout.padding
	rightX := boxLayout.contentBox.Max.X - boxLayout.padding
	for i, line := range lines {
		if line == "" {
			continue
		}
		layout := boxLayout.block.Layouts[i]
		var anchorX float64
		switch a.TextAlign {
		case TextAlignRight:
			anchorX = rightX
		case TextAlignCenter:
			anchorX = (leftX + rightX) / 2
		default:
			anchorX = leftX
		}
		baselineY := top + boxLayout.block.BaselineYs[i]
		drawDisplayText(
			textRen,
			line,
			alignedSingleLineOrigin(geom.Pt{X: anchorX, Y: baselineY}, layout, a.TextAlign, textLayoutVAlignBaseline),
			fontSize,
			resolvedTextColor(a.TextColor, ctx),
			ctx.RC.FontKey,
			ctx.RC.UseTeX,
		)
	}
}

// Bounds returns an empty rect so anchored text boxes do not affect autoscaling.
func (a *AnchoredTextBox) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the anchored text box z-order.
func (a *AnchoredTextBox) Z() float64 { return a.z }

func (a *AnchoredTextBox) boxRect(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if a == nil || r == nil || ctx == nil {
		return geom.Rect{}, false
	}

	lines := strings.Split(a.Content, "\n")
	if len(lines) == 0 {
		return geom.Rect{}, false
	}

	fontSize := resolvedFontSize(a.FontSize, ctx)
	return a.layout(r, ctx, lines, fontSize).patchBox, true
}

type anchoredTextLayout struct {
	contentBox geom.Rect
	patchBox   geom.Rect
	padding    float64
	block      multilineTextBlockLayout
}

func (a *AnchoredTextBox) layout(r render.Renderer, ctx *DrawContext, lines []string, fontSize float64) anchoredTextLayout {
	padding := a.resolvedPadding(fontSize, ctx)
	inset := a.resolvedInset(fontSize, ctx)
	boxPadding := a.resolvedBoxPadding()

	// Measure the multi-line block with matplotlib's Text._get_layout semantics
	// (linespacing 1.2 × font line height). The block is measured at the origin
	// with top alignment so block.BaselineYs are offsets from the inner top edge.
	block, _ := measureMultilineTextBlock(
		r, ctx, geom.Pt{}, fontSize, ctx.RC.FontKey,
		true, ctx.RC.UseTeX, lines, 0, a.TextAlign, textLayoutVAlignTop,
	)

	contentBox := resolveAnchoredBoxRect(a.Locator, ctx.Clip, block.Width+padding*2, block.Height+padding*2, a.Location, inset)
	return anchoredTextLayout{
		contentBox: contentBox,
		patchBox:   expandAnchoredRect(contentBox, boxPadding),
		padding:    padding,
		block:      block,
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

func (a *AnchoredTextBox) resolvedBoxPadding() float64 {
	if a == nil || a.BoxPadding <= 0 {
		return 0
	}
	return a.BoxPadding
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

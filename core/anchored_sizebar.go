package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// AnchoredSizeBarOptions configures an anchored scale bar.
type AnchoredSizeBarOptions struct {
	Location LegendLocation
	Locator  AnchoredBoxLocator
	Coords   CoordinateSpec

	Padding float64
	Inset   float64
	Sep     float64

	SizeVertical float64
	FillBar      *bool
	LabelTop     bool
	FrameOn      *bool

	Color           render.Color
	BackgroundColor render.Color
	BorderColor     render.Color
	BorderWidth     float64
	LineWidth       float64
	FontSize        float64
	FontKey         string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties
}

// AnchoredSizeBar draws a Matplotlib-style anchored scale bar with a label.
type AnchoredSizeBar struct {
	Size  float64
	Label string

	Location LegendLocation
	Locator  AnchoredBoxLocator
	Coords   CoordinateSpec

	Padding float64
	Inset   float64
	Sep     float64

	SizeVertical float64
	FillBar      bool
	LabelTop     bool
	FrameOn      bool

	Color           render.Color
	BackgroundColor render.Color
	BorderColor     render.Color
	BorderWidth     float64
	LineWidth       float64
	FontSize        float64
	FontKey         string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties
	z              float64
}

// AddAnchoredSizeBar appends an anchored scale bar to the axes.
func (a *Axes) AddAnchoredSizeBar(size float64, label string, opts ...AnchoredSizeBarOptions) *AnchoredSizeBar {
	rc := style.CurrentDefaults()
	if a != nil {
		rc = a.resolvedRC()
	}
	bar := newAnchoredSizeBar(size, label, rc, opts...)
	if a != nil {
		a.Add(bar)
	}
	return bar
}

func newAnchoredSizeBar(size float64, label string, rc style.RC, opts ...AnchoredSizeBarOptions) *AnchoredSizeBar {
	cfg := AnchoredSizeBarOptions{
		Location:        LegendLowerRight,
		Coords:          Coords(CoordData),
		Padding:         -1,
		Inset:           -1,
		Sep:             -1,
		Color:           render.Color{R: 0, G: 0, B: 0, A: 1},
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		BorderWidth:     1,
		LineWidth:       1.5,
	}
	frameOn := true
	cfg.FrameOn = &frameOn
	if len(opts) > 0 {
		opt := opts[0]
		cfg.Location = opt.Location
		cfg.Locator = opt.Locator
		if opt.Coords != (CoordinateSpec{}) {
			cfg.Coords = opt.Coords
		}
		if opt.Padding >= 0 {
			cfg.Padding = opt.Padding
		}
		if opt.Inset >= 0 {
			cfg.Inset = opt.Inset
		}
		if opt.Sep >= 0 {
			cfg.Sep = opt.Sep
		}
		cfg.SizeVertical = opt.SizeVertical
		cfg.FillBar = cloneBool(opt.FillBar)
		cfg.LabelTop = opt.LabelTop
		if opt.FrameOn != nil {
			cfg.FrameOn = cloneBool(opt.FrameOn)
		}
		if opt.Color != (render.Color{}) {
			cfg.Color = opt.Color
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
		if opt.LineWidth > 0 {
			cfg.LineWidth = opt.LineWidth
		}
		if opt.FontSize > 0 {
			cfg.FontSize = opt.FontSize
		}
		cfg.FontKey = opt.FontKey
		cfg.FontProperties = cloneFontProperties(opt.FontProperties)
	}
	fillBar := cfg.SizeVertical > 0
	if cfg.FillBar != nil {
		fillBar = *cfg.FillBar
	}
	return &AnchoredSizeBar{
		Size:            size,
		Label:           label,
		Location:        cfg.Location,
		Locator:         cfg.Locator,
		Coords:          cfg.Coords,
		Padding:         cfg.Padding,
		Inset:           cfg.Inset,
		Sep:             cfg.Sep,
		SizeVertical:    cfg.SizeVertical,
		FillBar:         fillBar,
		LabelTop:        cfg.LabelTop,
		FrameOn:         cfg.FrameOn == nil || *cfg.FrameOn,
		Color:           cfg.Color,
		BackgroundColor: cfg.BackgroundColor,
		BorderColor:     cfg.BorderColor,
		BorderWidth:     cfg.BorderWidth,
		LineWidth:       cfg.LineWidth,
		FontSize:        cfg.FontSize,
		FontKey:         cfg.FontKey,
		FontProperties:  cloneFontProperties(cfg.FontProperties),
		z:               950,
	}
}

func (a *AnchoredSizeBar) Draw(r render.Renderer, ctx *DrawContext) {
	if a == nil || r == nil || ctx == nil {
		return
	}
	textRen, ok := r.(render.TextDrawer)
	if !ok {
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
			LineWidth: a.BorderWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Snap:      render.SnapAuto,
		})
	}
	if layout.bar.H() > 0 {
		paint := &render.Paint{
			Stroke:    a.Color,
			LineWidth: a.LineWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Snap:      render.SnapAuto,
		}
		if a.FillBar {
			paint.Fill = a.Color
		}
		r.Path(pixelRectPath(layout.bar), paint)
	} else {
		r.Path(geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{
				{X: layout.bar.Min.X, Y: layout.bar.Min.Y},
				{X: layout.bar.Max.X, Y: layout.bar.Min.Y},
			},
		}, &render.Paint{
			Stroke:    a.Color,
			LineWidth: a.LineWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
		})
	}
	if a.Label != "" {
		drawDisplayText(textRen, a.Label, layout.labelOrigin, layout.fontSize, a.Color, layout.fontKey, ctx.RC.UseTeX)
	}
}

// Bounds returns an empty rect so anchored size bars do not affect autoscaling.
func (a *AnchoredSizeBar) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the anchored size bar z-order.
func (a *AnchoredSizeBar) Z() float64 { return a.z }

func (a *AnchoredSizeBar) boxRect(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	layout := a.layout(r, ctx)
	if layout.empty {
		return geom.Rect{}, false
	}
	return layout.frame, true
}

type anchoredSizeBarLayout struct {
	frame       geom.Rect
	content     geom.Rect
	bar         geom.Rect
	labelOrigin geom.Pt
	fontSize    float64
	fontKey     string
	empty       bool
}

func (a *AnchoredSizeBar) layout(r render.Renderer, ctx *DrawContext) anchoredSizeBarLayout {
	if a == nil || r == nil || ctx == nil {
		return anchoredSizeBarLayout{empty: true}
	}
	fontSize := resolvedFontSize(a.FontSize, ctx)
	fontKey := resolvedTextFontKey(a.FontKey, a.FontProperties, ctx)
	labelLayout := measureSingleLineTextLayout(r, a.Label, fontSize, fontKey, ctx.RC.UseTeX)
	barW, barH := a.displaySize(ctx)
	if barW <= 0 && labelLayout.Width <= 0 {
		return anchoredSizeBarLayout{empty: true}
	}
	if barH < 0 {
		barH = -barH
	}
	padding := a.resolvedPadding(fontSize, ctx)
	sep := a.resolvedSep(ctx)
	labelH := labelLayout.RunAscent + labelLayout.RunDescent
	if labelH <= 0 {
		labelH = labelLayout.Height
	}
	contentW := math.Max(barW, labelLayout.Width)
	contentH := barH
	if a.Label != "" {
		contentH += sep + labelH
	}
	content := resolveAnchoredBoxRect(a.Locator, ctx.Clip, contentW+padding*2, contentH+padding*2, a.Location, a.resolvedInset(fontSize, ctx))
	inner := geom.Rect{
		Min: geom.Pt{X: content.Min.X + padding, Y: content.Min.Y + padding},
		Max: geom.Pt{X: content.Max.X - padding, Y: content.Max.Y - padding},
	}
	barX := inner.Min.X + (inner.W()-barW)/2
	labelX := inner.Min.X + inner.W()/2
	var barY, labelCenterY float64
	if a.LabelTop {
		barY = inner.Min.Y
		labelCenterY = inner.Min.Y + barH + sep + labelH/2
	} else {
		labelCenterY = inner.Min.Y + labelH/2
		barY = inner.Min.Y + labelH + sep
	}
	bar := geom.Rect{
		Min: geom.Pt{X: barX, Y: barY},
		Max: geom.Pt{X: barX + barW, Y: barY + barH},
	}
	labelOrigin := alignedSingleLineOrigin(geom.Pt{X: labelX, Y: labelCenterY}, labelLayout, TextAlignCenter, textLayoutVAlignCenter)
	return anchoredSizeBarLayout{
		frame:       content,
		content:     inner,
		bar:         bar,
		labelOrigin: labelOrigin,
		fontSize:    fontSize,
		fontKey:     fontKey,
	}
}

func (a *AnchoredSizeBar) displaySize(ctx *DrawContext) (float64, float64) {
	tr := ctx.TransformFor(a.Coords)
	if tr == nil {
		return a.Size, a.SizeVertical
	}
	origin := tr.Apply(geom.Pt{})
	size := tr.Apply(geom.Pt{X: a.Size, Y: 0})
	vertical := tr.Apply(geom.Pt{X: 0, Y: a.SizeVertical})
	return math.Abs(size.X - origin.X), math.Abs(vertical.Y - origin.Y)
}

func (a *AnchoredSizeBar) resolvedPadding(fontSize float64, ctx *DrawContext) float64 {
	if a != nil && a.Padding >= 0 {
		return a.Padding
	}
	return pointsToPixels(ctx.RC, 0.1*fontSize)
}

func (a *AnchoredSizeBar) resolvedInset(fontSize float64, ctx *DrawContext) float64 {
	if a != nil && a.Inset >= 0 {
		return a.Inset
	}
	return pointsToPixels(ctx.RC, 0.1*fontSize)
}

func (a *AnchoredSizeBar) resolvedSep(ctx *DrawContext) float64 {
	if a != nil && a.Sep >= 0 {
		return a.Sep
	}
	return pointsToPixels(ctx.RC, 2)
}

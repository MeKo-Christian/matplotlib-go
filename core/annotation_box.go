package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/optarg"
	"github.com/cwbudde/matplotlib-go/render"
)

// AnnotationBboxOptions configures an AnnotationBbox text-area artist.
type AnnotationBboxOptions struct {
	XYCoords     CoordinateSpec
	BoxCoords    CoordinateSpec
	BoxPosition  *geom.Pt
	BoxAlignment *geom.Pt
	FrameOn      *bool
	Padding      float64

	Image     render.Image
	ImageZoom float64

	FaceColor render.Color
	EdgeColor render.Color
	LineWidth float64
	TextColor render.Color
	FontSize  float64
	FontKey   string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties

	Arrow           bool
	ArrowColor      render.Color
	ArrowWidth      float64
	ArrowHeadSize   float64
	ArrowStyle      ArrowStyle
	ConnectionStyle ConnectionStyle
	AnnotationClip  *bool
}

// AnnotationBbox renders a framed text box tied to an annotated point.
type AnnotationBbox struct {
	ArtistRasterization
	Point   geom.Pt
	Content string

	XYCoords     CoordinateSpec
	BoxCoords    CoordinateSpec
	BoxPosition  *geom.Pt
	BoxAlignment geom.Pt
	FrameOn      bool
	Padding      float64

	Image     render.Image
	ImageZoom float64

	FaceColor render.Color
	EdgeColor render.Color
	LineWidth float64
	TextColor render.Color
	FontSize  float64
	FontKey   string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties *render.FontProperties

	Arrow           bool
	ArrowColor      render.Color
	ArrowWidth      float64
	ArrowHeadSize   float64
	ArrowStyle      ArrowStyle
	ConnectionStyle ConnectionStyle
	AnnotationClip  *bool
	z               float64
}

// AnnotationBbox adds a framed text box tied to an annotated point.
func (a *Axes) AnnotationBbox(text string, x, y float64, opts ...AnnotationBboxOptions) *AnnotationBbox {
	if a == nil {
		return nil
	}
	cfg := AnnotationBboxOptions{
		XYCoords:  Coords(CoordData),
		BoxCoords: Coords(CoordData),
		FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 1.0, // points; converted at the Paint sink
	}
	frameOn := true
	cfg.FrameOn = &frameOn
	align := geom.Pt{X: 0.5, Y: 0.5}
	cfg.BoxAlignment = &align

	defaultArrowStyle, _ := ArrowStyleFromString("-|>")
	defaultArrowStyle.HeadWidth = 0.36
	defaultConnectionStyle, _ := ConnectionStyleFromString("arc3")
	cfg.ArrowStyle = defaultArrowStyle
	cfg.ConnectionStyle = defaultConnectionStyle
	cfg.ArrowWidth = 1.0 // points; converted at the arrow patch sink
	cfg.ArrowHeadSize = 8

	if opt, ok := optarg.Optional("annotation bbox", opts); ok {
		cfg.XYCoords = opt.XYCoords
		cfg.BoxCoords = opt.BoxCoords
		cfg.BoxPosition = clonePoint(opt.BoxPosition)
		if opt.BoxAlignment != nil {
			cfg.BoxAlignment = clonePoint(opt.BoxAlignment)
		}
		if opt.FrameOn != nil {
			cfg.FrameOn = cloneBool(opt.FrameOn)
		}
		cfg.Padding = opt.Padding
		cfg.Image = opt.Image
		cfg.ImageZoom = opt.ImageZoom
		if opt.FaceColor != (render.Color{}) {
			cfg.FaceColor = opt.FaceColor
		}
		if opt.EdgeColor != (render.Color{}) {
			cfg.EdgeColor = opt.EdgeColor
		}
		if opt.LineWidth > 0 {
			cfg.LineWidth = opt.LineWidth
		}
		cfg.TextColor = opt.TextColor
		cfg.FontSize = opt.FontSize
		cfg.FontKey = opt.FontKey
		cfg.FontProperties = cloneFontProperties(opt.FontProperties)
		cfg.Arrow = opt.Arrow
		cfg.ArrowColor = opt.ArrowColor
		if opt.ArrowWidth > 0 {
			cfg.ArrowWidth = opt.ArrowWidth
		}
		if opt.ArrowHeadSize > 0 {
			cfg.ArrowHeadSize = opt.ArrowHeadSize
		}
		if opt.ArrowStyle.Name != "" {
			cfg.ArrowStyle = opt.ArrowStyle
		}
		if opt.ConnectionStyle.Name != "" {
			cfg.ConnectionStyle = opt.ConnectionStyle
		}
		cfg.AnnotationClip = cloneBool(opt.AnnotationClip)
	}

	if cfg.BoxAlignment == nil {
		cfg.BoxAlignment = &align
	}
	artist := &AnnotationBbox{
		Point:           geom.Pt{X: x, Y: y},
		Content:         text,
		XYCoords:        cfg.XYCoords,
		BoxCoords:       cfg.BoxCoords,
		BoxPosition:     clonePoint(cfg.BoxPosition),
		BoxAlignment:    *cfg.BoxAlignment,
		FrameOn:         cfg.FrameOn == nil || *cfg.FrameOn,
		Padding:         cfg.Padding,
		Image:           cfg.Image,
		ImageZoom:       cfg.ImageZoom,
		FaceColor:       cfg.FaceColor,
		EdgeColor:       cfg.EdgeColor,
		LineWidth:       cfg.LineWidth,
		TextColor:       cfg.TextColor,
		FontSize:        cfg.FontSize,
		FontKey:         cfg.FontKey,
		FontProperties:  cloneFontProperties(cfg.FontProperties),
		Arrow:           cfg.Arrow,
		ArrowColor:      cfg.ArrowColor,
		ArrowWidth:      cfg.ArrowWidth,
		ArrowHeadSize:   cfg.ArrowHeadSize,
		ArrowStyle:      cfg.ArrowStyle,
		ConnectionStyle: cfg.ConnectionStyle,
		AnnotationClip:  cloneBool(cfg.AnnotationClip),
		z:               900,
	}
	a.Add(artist)
	return artist
}

// Draw is a no-op because annotation boxes render outside the axes clip.
func (a *AnnotationBbox) Draw(render.Renderer, *DrawContext) {}

// DrawOverlay renders the annotation box without the axes clip applied.
func (a *AnnotationBbox) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil || (a.Image == nil && displayTextIsEmpty(a.Content)) {
		return
	}

	target := transformedPoint(ctx, a.XYCoords, a.Point, 0, 0)
	if annotationPointClipped(a.AnnotationClip, a.XYCoords, target, ctx.Clip) {
		return
	}

	fontSize := resolvedFontSize(a.FontSize, ctx)
	fontKey := resolvedTextFontKey(a.FontKey, a.FontProperties, ctx)
	var layout singleLineTextLayout
	if a.Image == nil {
		var ok bool
		if _, ok = r.(render.TextDrawer); !ok {
			return
		}
		layout = measureSingleLineTextLayout(r, a.Content, fontSize, fontKey, ctx.RC.UseTeX)
	}
	boxAnchor := a.boxAnchor(ctx)
	contentBox := annotationBoxRect(boxAnchor, a.contentSize(layout, ctx), a.BoxAlignment)
	box := expandAnchoredRect(contentBox, a.resolvedPadding(fontSize, ctx))

	if a.Arrow {
		a.drawArrowFromBox(r, ctx, box, target)
	}
	if a.FrameOn {
		r.Path(pixelRectPath(box), &render.Paint{
			Fill:      a.FaceColor,
			Stroke:    a.EdgeColor,
			LineWidth: pointsToPixels(ctx.RC, a.LineWidth),
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Snap:      render.SnapAuto,
		})
	}

	if a.Image != nil {
		img := offsetImageWithMatplotlibDefaults(a.Image)
		if drawer, ok := r.(render.BboxImageDrawer); ok && drawer.DrawBboxImage(img, contentBox) {
			return
		}
		r.DrawImage(img, contentBox)
		return
	}
	textRen := r.(render.TextDrawer)
	origin := alignedSingleLineOrigin(rectCenter(contentBox), layout, TextAlignCenter, textLayoutVAlignCenter)
	drawDisplayText(textRen, a.Content, origin, fontSize, resolvedTextColor(a.TextColor, ctx), fontKey, ctx.RC.UseTeX)
}

// Bounds returns an empty rect so annotation boxes do not affect autoscaling.
func (a *AnnotationBbox) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the annotation-box z-order.
func (a *AnnotationBbox) Z() float64 { return a.z }

func (a *AnnotationBbox) boxAnchor(ctx *DrawContext) geom.Pt {
	if a != nil && a.BoxPosition != nil {
		return transformedPoint(ctx, a.BoxCoords, *a.BoxPosition, 0, 0)
	}
	if a == nil {
		return geom.Pt{}
	}
	return transformedPoint(ctx, a.XYCoords, a.Point, 0, 0)
}

func (a *AnnotationBbox) resolvedPadding(fontSize float64, ctx *DrawContext) float64 {
	if a != nil && a.Padding > 0 {
		return a.Padding
	}
	return pointsToPixels(ctx.RC, 0.4*fontSize)
}

func (a *AnnotationBbox) contentSize(layout singleLineTextLayout, ctx *DrawContext) geom.Pt {
	if a != nil && a.Image != nil {
		width, height := a.Image.Size()
		zoom := a.ImageZoom
		if zoom <= 0 {
			zoom = 1
		}
		scale := drawingAreaScale(ctx)
		return geom.Pt{X: float64(width) * zoom * scale, Y: float64(height) * zoom * scale}
	}
	return annotationTextBoxSize(layout)
}

func (a *AnnotationBbox) drawArrow(r render.Renderer, ctx *DrawContext, start, target geom.Pt) {
	annotation := Annotation{
		Color:           a.TextColor,
		ArrowColor:      a.ArrowColor,
		ArrowWidth:      a.ArrowWidth,
		ArrowHeadSize:   a.ArrowHeadSize,
		ArrowStyle:      a.ArrowStyle,
		ConnectionStyle: a.ConnectionStyle,
	}
	annotation.drawArrow(r, ctx, start, target)
}

func (a *AnnotationBbox) drawArrowFromBox(r render.Renderer, ctx *DrawContext, box geom.Rect, target geom.Pt) {
	annotation := Annotation{
		Color:           a.TextColor,
		ArrowColor:      a.ArrowColor,
		ArrowWidth:      a.ArrowWidth,
		ArrowHeadSize:   a.ArrowHeadSize,
		ArrowStyle:      a.ArrowStyle,
		ConnectionStyle: a.ConnectionStyle,
	}
	annotation.drawArrowFromPatchBox(r, ctx, box, box, target)
}

func annotationTextBoxSize(layout singleLineTextLayout) geom.Pt {
	height := layout.RunAscent + layout.RunDescent
	if height <= 0 {
		height = layout.Height
	}
	return geom.Pt{X: layout.Width, Y: height}
}

func annotationBoxRect(anchor, size, alignment geom.Pt) geom.Rect {
	min := geom.Pt{
		X: anchor.X - size.X*alignment.X,
		Y: anchor.Y - size.Y*(1-alignment.Y),
	}
	return geom.Rect{
		Min: min,
		Max: geom.Pt{X: min.X + size.X, Y: min.Y + size.Y},
	}
}

func rectCenter(rect geom.Rect) geom.Pt {
	return geom.Pt{
		X: (rect.Min.X + rect.Max.X) / 2,
		Y: (rect.Min.Y + rect.Max.Y) / 2,
	}
}

func clonePoint(pt *geom.Pt) *geom.Pt {
	if pt == nil {
		return nil
	}
	v := *pt
	return &v
}

package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
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
	TextVAlignCenterBaseline
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
	// Wrap computes a display-pixel wrap width from the figure box when
	// WrapWidth is not set.
	Wrap bool
	// MultiAlignment controls per-line alignment within multiline or wrapped
	// text. Nil follows HAlign, matching Matplotlib's multialignment=None.
	MultiAlignment *TextAlign
	// Linespacing controls multiline baseline advance as a multiple of the font
	// size in display pixels. Zero uses Matplotlib's normal 1.2 spacing.
	Linespacing float64
	ClipOn      *bool
	BBox        *TextBBoxOptions
	FontKey     string
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

// AnnotationOffsetUnits controls how AnnotationOptions.OffsetX/Y are converted.
type AnnotationOffsetUnits uint8

const (
	// AnnotationOffsetPixels interprets OffsetX/Y as display pixels.
	AnnotationOffsetPixels AnnotationOffsetUnits = iota
	// AnnotationOffsetPoints interprets OffsetX/Y as typographic points.
	AnnotationOffsetPoints
)

// AnnotationOptions configures an Annotation artist.
type AnnotationOptions struct {
	Coords CoordinateSpec
	// TextPosition sets an explicit annotation text anchor, matching
	// Matplotlib's xytext. When nil, the text anchor is the annotated point plus
	// OffsetX/OffsetY in display pixels.
	TextPosition *geom.Pt
	// TextCoords controls TextPosition's coordinate space. The zero value is
	// data coordinates.
	TextCoords CoordinateSpec
	OffsetX    float64
	OffsetY    float64
	// OffsetUnits controls whether OffsetX/Y are display pixels or typographic
	// points. The zero value preserves the historical pixel behavior.
	OffsetUnits     AnnotationOffsetUnits
	FontSize        float64
	Color           render.Color
	ArrowColor      render.Color
	ArrowWidth      float64
	ArrowHeadSize   float64
	ArrowStyle      ArrowStyle
	ConnectionStyle ConnectionStyle
	HAlign          TextAlign
	VAlign          TextVerticalAlign
	Angle           float64
	FontKey         string
	BBox            *TextBBoxOptions
	Linespacing     float64
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
	// Wrap computes a display-pixel wrap width from the figure box when
	// WrapWidth is not set.
	Wrap bool
	// MultiAlignment controls per-line alignment within multiline or wrapped
	// text. Nil follows HAlign, matching Matplotlib's multialignment=None.
	MultiAlignment *TextAlign
	// Linespacing controls multiline baseline advance as a multiple of the font
	// size in display pixels. Zero uses Matplotlib's normal 1.2 spacing.
	Linespacing float64
	ClipOn      bool
	BBox        *TextBBoxOptions
	FontKey     string
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

	TextPosition    *geom.Pt
	TextCoords      CoordinateSpec
	OffsetX         float64
	OffsetY         float64
	OffsetUnits     AnnotationOffsetUnits
	FontSize        float64
	Color           render.Color
	ArrowColor      render.Color
	ArrowWidth      float64
	ArrowHeadSize   float64
	ArrowStyle      ArrowStyle
	ConnectionStyle ConnectionStyle
	HAlign          TextAlign
	VAlign          TextVerticalAlign
	Angle           float64
	Coords          CoordinateSpec
	FontKey         string
	BBox            *TextBBoxOptions
	Linespacing     float64
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
	clipOn := false
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
		Wrap:           opt.Wrap,
		MultiAlignment: cloneTextAlign(opt.MultiAlignment),
		Linespacing:    opt.Linespacing,
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
		Wrap:           opt.Wrap,
		MultiAlignment: cloneTextAlign(opt.MultiAlignment),
		Linespacing:    opt.Linespacing,
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
		TextPosition:    clonePoint(opt.TextPosition),
		TextCoords:      opt.TextCoords,
		OffsetX:         opt.OffsetX,
		OffsetY:         opt.OffsetY,
		OffsetUnits:     opt.OffsetUnits,
		FontSize:        opt.FontSize,
		Color:           opt.Color,
		ArrowColor:      opt.ArrowColor,
		ArrowWidth:      opt.ArrowWidth,
		ArrowHeadSize:   opt.ArrowHeadSize,
		ArrowStyle:      opt.ArrowStyle,
		ConnectionStyle: opt.ConnectionStyle,
		HAlign:          annotationHAlign(opt),
		VAlign:          annotationVAlign(opt),
		Angle:           opt.Angle,
		Coords:          opt.Coords,
		FontKey:         opt.FontKey,
		BBox:            cloneTextBBoxOptions(opt.BBox),
		Linespacing:     opt.Linespacing,
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

// Bounds returns an empty rect so labels do not affect autoscaling.
func (t *Text) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the text z-order.
func (t *Text) Z() float64 { return t.z }

// Draw is a no-op because annotations render outside the axes clip via DrawOverlay.
func (a *Annotation) Draw(render.Renderer, *DrawContext) {}

// DrawOverlay renders the full annotation without the axes clip applied.
func (a *Annotation) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	a.drawAnnotationOverlay(r, ctx)
}

// Bounds returns an empty rect so annotations do not affect autoscaling.
func (a *Annotation) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the annotation z-order.
func (a *Annotation) Z() float64 { return a.z }

func annotationHAlign(opt AnnotationOptions) TextAlign {
	return opt.HAlign
}

func annotationVAlign(opt AnnotationOptions) TextVerticalAlign {
	return opt.VAlign
}

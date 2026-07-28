package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
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
	// text. Unset follows HAlign, matching Matplotlib's multialignment=None.
	MultiAlignment optional.Value[TextAlign]
	// Linespacing controls multiline baseline advance as a multiple of the font
	// size in display pixels. Zero uses Matplotlib's normal 1.2 spacing.
	Linespacing float64
	// ClipOn clips the text to the axes. Unset keeps each entry point's own
	// default: Axes.Text does not clip, Figure.Text does.
	ClipOn optional.Value[bool]
	// BBox draws a styled background box behind the text. Unset draws none.
	BBox    optional.Value[TextBBoxOptions]
	FontKey string
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties optional.Value[render.FontProperties]
	ParseMath      optional.Value[bool]
	PathEffects    []render.PathEffect
}

// TextBBoxOptions configures a styled background box behind text.
//
// The box geometry mirrors Matplotlib's Text bbox, which is drawn with a
// FancyBboxPatch. Three ways to pick a shape are supported, in priority order:
//
//   - Style: a Matplotlib boxstyle spec string, e.g. "round,pad=0.3" or
//     "sawtooth,pad=0.5,tooth_size=0.1". When non-empty it populates BoxStyle,
//     Padding, RoundingSize, and ToothSize (parsed internally).
//   - BoxStyle (+ the per-style params below): the typed enum alternative.
//   - CornerRadius: a backward-compatible shortcut for a square box with rounded
//     corners (only consulted when Style is empty and BoxStyle is BoxStyleSquare).
type TextBBoxOptions struct {
	FaceColor    render.Color
	EdgeColor    render.Color
	LineWidth    float64
	Padding      float64
	CornerRadius float64
	// Style is a Matplotlib boxstyle spec such as "round,pad=0.3,rounding_size=0.2".
	// When set it overrides BoxStyle/RoundingSize/ToothSize and (if it carries a
	// pad) Padding.
	Style string
	// BoxStyle selects a FancyBboxPatch shape (square, round, sawtooth, arrows, …).
	BoxStyle BoxStyle
	// RoundingSize is the corner radius for BoxStyleRound / BoxStyleRound4.
	RoundingSize float64
	// ToothSize is the tooth amplitude for BoxStyleSawtooth / BoxStyleRoundtooth.
	ToothSize float64
	// ArrowHeadWidth and ArrowHeadAngle tune the arrow boxstyles (Go extension;
	// Matplotlib's larrow/rarrow/darrow take only pad).
	ArrowHeadWidth float64
	ArrowHeadAngle float64
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
	// Matplotlib's xytext. Unset anchors the text at the annotated point plus
	// OffsetX/OffsetY.
	TextPosition optional.Value[geom.Pt]
	// TextCoords controls TextPosition's coordinate space. The zero value is
	// data coordinates.
	TextCoords CoordinateSpec
	// OffsetX and OffsetY shift the text from the annotated point. Unset uses
	// Matplotlib's (28, -20) default offset; an explicit 0 means no shift on
	// that axis.
	OffsetX optional.Value[float64]
	OffsetY optional.Value[float64]
	// OffsetUnits controls whether OffsetX/Y are display pixels or typographic
	// points. The zero value preserves the historical pixel behavior.
	OffsetUnits AnnotationOffsetUnits
	FontSize    float64
	Color       render.Color
	ArrowColor  render.Color
	// ArrowWidth is the arrow line width in points. Unset uses 1.25.
	ArrowWidth optional.Value[float64]
	// ArrowHeadSize is the arrow head length in points. Unset uses 8.
	ArrowHeadSize optional.Value[float64]
	// ArrowStyle and ConnectionStyle default to Matplotlib's "-|>" and "arc3".
	ArrowStyle      ArrowStyle
	ConnectionStyle ConnectionStyle
	HAlign          TextAlign
	VAlign          TextVerticalAlign
	Angle           float64
	FontKey         string
	BBox            optional.Value[TextBBoxOptions]
	Linespacing     float64
	// FontProperties is a structured alternative to FontKey. FontKey wins when
	// both are set.
	FontProperties optional.Value[render.FontProperties]
	// ParseMath enables mathtext parsing. Unset follows the rc default.
	ParseMath optional.Value[bool]
	// AnnotationClip clips the annotation to the axes. Unset follows
	// Matplotlib's behavior of clipping only when the text is in data
	// coordinates.
	AnnotationClip optional.Value[bool]
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
//
//nolint:gocritic // TextOptions is an immutable snapshot of the caller's options.
func (a *Axes) Text(x, y float64, text string, opt TextOptions) *Text {
	// TextAlignLeft and TextVAlignBaseline are the zero values, so the zero
	// TextOptions already carries Matplotlib's default alignment.
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
		MultiAlignment: opt.MultiAlignment.Ptr(),
		Linespacing:    opt.Linespacing,
		ClipOn:         opt.ClipOn.Or(false),
		BBox:           opt.BBox.Ptr(),
		FontKey:        opt.FontKey,
		FontProperties: cloneFontProperties(opt.FontProperties.Ptr()),
		ParseMath:      opt.ParseMath.Ptr(),
		PathEffects:    cloneRenderPathEffects(opt.PathEffects),
		z:              500,
	}
	a.Add(artist)
	return artist
}

// Text adds arbitrary text positioned in figure-fraction coordinates.
//
//nolint:gocritic // TextOptions is an immutable snapshot of the caller's options.
func (f *Figure) Text(x, y float64, text string, opt TextOptions) *Text {
	if f == nil {
		return nil
	}
	// Figure text is always positioned in figure-fraction coordinates.
	opt.Coords = Coords(CoordFigure)

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
		MultiAlignment: opt.MultiAlignment.Ptr(),
		Linespacing:    opt.Linespacing,
		ClipOn:         opt.ClipOn.Or(true),
		BBox:           opt.BBox.Ptr(),
		FontKey:        opt.FontKey,
		FontProperties: cloneFontProperties(opt.FontProperties.Ptr()),
		ParseMath:      opt.ParseMath.Ptr(),
		PathEffects:    cloneRenderPathEffects(opt.PathEffects),
		z:              500,
	}
	f.Artists = append(f.Artists, artist)
	f.zsorted = false
	return artist
}

// Annotate adds an arrow annotation pointing to a data-space point.
//
//nolint:gocritic // AnnotationOptions is an immutable snapshot of the caller's options.
func (a *Axes) Annotate(text string, x, y float64, opt AnnotationOptions) *Annotation {
	arrowStyle := opt.ArrowStyle
	if arrowStyle.Name == "" {
		arrowStyle, _ = ArrowStyleFromString("-|>")
		arrowStyle.HeadWidth = 0.36
	}
	connectionStyle := opt.ConnectionStyle
	if connectionStyle.Name == "" {
		connectionStyle, _ = ConnectionStyleFromString("arc3")
	}

	artist := &Annotation{
		Point:           geom.Pt{X: x, Y: y},
		Content:         text,
		TextPosition:    opt.TextPosition.Ptr(),
		TextCoords:      opt.TextCoords,
		OffsetX:         opt.OffsetX.Or(28),
		OffsetY:         opt.OffsetY.Or(-20),
		OffsetUnits:     opt.OffsetUnits,
		FontSize:        opt.FontSize,
		Color:           opt.Color,
		ArrowColor:      opt.ArrowColor,
		ArrowWidth:      opt.ArrowWidth.Or(1.25),
		ArrowHeadSize:   opt.ArrowHeadSize.Or(annotationDefaultHeadSize(a, opt.FontSize)),
		ArrowStyle:      arrowStyle,
		ConnectionStyle: connectionStyle,
		HAlign:          annotationHAlign(opt),
		VAlign:          annotationVAlign(opt),
		Angle:           opt.Angle,
		Coords:          opt.Coords,
		FontKey:         opt.FontKey,
		BBox:            opt.BBox.Ptr(),
		Linespacing:     opt.Linespacing,
		FontProperties:  cloneFontPropertiesValue(opt.FontProperties).Ptr(),
		ParseMath:       opt.ParseMath.Ptr(),
		AnnotationClip:  opt.AnnotationClip.Ptr(),
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

// annotationDefaultHeadSize is Matplotlib's default arrow mutation scale, which
// is the annotation text's own font size rather than a constant
// (Annotation.update_positions: ms = arrowprops.get("mutation_scale",
// self.get_size())).
func annotationDefaultHeadSize(a *Axes, fontSize float64) float64 {
	if fontSize > 0 {
		return fontSize
	}
	if a != nil {
		if rc := a.resolvedRC(); rc.FontSize > 0 {
			return rc.FontSize
		}
	}
	return 12
}

func annotationHAlign(opt AnnotationOptions) TextAlign {
	return opt.HAlign
}

func annotationVAlign(opt AnnotationOptions) TextVerticalAlign {
	return opt.VAlign
}

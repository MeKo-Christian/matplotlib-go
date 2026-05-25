package render

import (
	"errors"
	"image"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// DefaultHatchSpacing matches Matplotlib's default hatch density of six
// pattern rows or lines in a 100 DPI tile.
const DefaultHatchSpacing = 100.0 / 6.0

// Paint configures drawing style for paths.
type Paint struct {
	LineWidth         float64
	LineJoin          LineJoin
	LineCap           LineCap
	MiterLimit        float64
	Stroke            Color
	Fill              Color
	Dashes            []float64 // on/off pairs, in user space units
	Antialias         AntialiasMode
	Snap              SnapMode // path snapping policy; zero preserves existing unsnapped behavior
	Simplify          bool     // simplify line-only paths before rasterization
	SimplifyThreshold float64  // simplification tolerance in display pixels
	MaxChunkVertices  int      // split very large stroke-only line paths; <=0 uses backend default
	Hatch             string
	HatchColor        Color
	HatchLineWidth    float64
	HatchSpacing      float64
	FillPattern       PatternFill
	FillGradient      GradientFill
	PathEffects       []PathEffect
	CompositeMode     CompositeMode
	Rasterization     Rasterization
	Sketch            SketchParams
	ForceAlpha        bool
	Alpha             float64
	ClipPathTransform geom.Affine
	HasClipPathTrans  bool
}

// CompositeMode identifies the Porter-Duff/blend operation used for a draw.
//
// SourceOver is the zero value because it is the default model used by the
// existing raster and vector backends.
type CompositeMode uint8

const (
	CompositeSourceOver CompositeMode = iota
	CompositeSource
	CompositeMultiply
	CompositeScreen
	CompositeOverlay
	CompositeDarken
	CompositeLighten
)

// PatternFill describes a renderer-neutral tiled path pattern.
//
// Backends that cannot consume it natively may fall back to normal solid fill
// handling or a core-side pattern expansion path.
type PatternFill struct {
	ID            string
	Cell          geom.Rect
	Path          geom.Path
	Foreground    Color
	Background    Color
	LineWidth     float64
	Transform     geom.Affine
	HasTransform  bool
	CompositeMode CompositeMode
}

// GradientKind identifies a renderer-neutral gradient geometry.
type GradientKind uint8

const (
	GradientNone GradientKind = iota
	LinearGradient
	RadialGradient
)

// GradientStop is one color stop in a gradient fill.
type GradientStop struct {
	Offset float64
	Color  Color
}

// GradientFill describes a renderer-neutral linear or radial gradient.
type GradientFill struct {
	Kind         GradientKind
	Start        geom.Pt
	End          geom.Pt
	Center       geom.Pt
	Focal        geom.Pt
	Radius       float64
	Stops        []GradientStop
	Transform    geom.Affine
	HasTransform bool
}

// PathEffectKind identifies a post-stroke/post-fill path rendering pass.
type PathEffectKind uint8

const (
	PathEffectNormal PathEffectKind = iota
	PathEffectStroke
	PathEffectShadow
	PathEffectPathPatch
	PathEffectTickedStroke
	PathEffectFilter
)

// PathEffect describes one renderer-neutral path effect pass.
type PathEffect struct {
	Kind          PathEffectKind
	Offset        geom.Pt
	Stroke        Color
	Fill          Color
	LineWidth     float64
	LineJoin      LineJoin
	LineCap       LineCap
	MiterLimit    float64
	Dashes        []float64
	CompositeMode CompositeMode
	Filter        string
	FilterRadius  float64
	ShadowAlpha   float64
	ShadowRho     float64
	TickSpacing   float64
	TickAngle     float64
	TickLength    float64
}

// NormalPathEffect draws the original path unchanged.
func NormalPathEffect() PathEffect {
	return PathEffect{Kind: PathEffectNormal}
}

// StrokePathEffect draws the path with an alternate stroke. Offset is in
// display pixels, matching the display-space paths sent to renderers.
func StrokePathEffect(stroke Color, linewidth float64, offset geom.Pt) PathEffect {
	return PathEffect{
		Kind:      PathEffectStroke,
		Offset:    offset,
		Stroke:    stroke,
		LineWidth: linewidth,
	}
}

// WithStrokePathEffects returns Stroke followed by Normal, matching
// Matplotlib's patheffects.withStroke helper.
func WithStrokePathEffects(stroke Color, linewidth float64, offset geom.Pt) []PathEffect {
	return []PathEffect{
		StrokePathEffect(stroke, linewidth, offset),
		NormalPathEffect(),
	}
}

// SimplePatchShadowPathEffect draws a filled shadow patch. If shadow has zero
// alpha, the original fill color is darkened by rho and drawn at alpha.
func SimplePatchShadowPathEffect(offset geom.Pt, shadow Color, alpha, rho float64) PathEffect {
	if alpha <= 0 {
		alpha = 0.3
	}
	if rho <= 0 {
		rho = 0.3
	}
	if shadow.A > 0 {
		shadow.A *= alpha
	}
	return PathEffect{
		Kind:        PathEffectShadow,
		Offset:      offset,
		Fill:        shadow,
		ShadowAlpha: alpha,
		ShadowRho:   rho,
	}
}

// SimpleLineShadowPathEffect draws a stroked shadow line. If shadow has zero
// alpha, the original stroke color is darkened by rho and drawn at alpha.
func SimpleLineShadowPathEffect(offset geom.Pt, shadow Color, alpha, rho float64) PathEffect {
	if alpha <= 0 {
		alpha = 0.3
	}
	if rho <= 0 {
		rho = 0.3
	}
	if shadow.A > 0 {
		shadow.A *= alpha
	}
	return PathEffect{
		Kind:        PathEffectShadow,
		Offset:      offset,
		Stroke:      shadow,
		ShadowAlpha: alpha,
		ShadowRho:   rho,
	}
}

// PathPatchPathEffect draws the original path with an alternate patch paint.
func PathPatchPathEffect(fill, stroke Color, linewidth float64, offset geom.Pt) PathEffect {
	return PathEffect{
		Kind:      PathEffectPathPatch,
		Offset:    offset,
		Fill:      fill,
		Stroke:    stroke,
		LineWidth: linewidth,
	}
}

// TickedStrokePathEffect draws short tick marks along the path.
func TickedStrokePathEffect(stroke Color, linewidth, spacing, angle, length float64, offset geom.Pt) PathEffect {
	return PathEffect{
		Kind:        PathEffectTickedStroke,
		Offset:      offset,
		Stroke:      stroke,
		LineWidth:   linewidth,
		TickSpacing: spacing,
		TickAngle:   angle,
		TickLength:  length,
	}
}

// FilterPathEffect draws a repaint pass into an offscreen filter surface when
// the backend supports render.FilterRenderer. Supported filter names are
// "none", "identity", "blur", "gaussian", "gaussian-blur", and "shadow".
func FilterPathEffect(fill, stroke Color, linewidth float64, filter string, radius float64, offset geom.Pt) PathEffect {
	return PathEffect{
		Kind:         PathEffectFilter,
		Offset:       offset,
		Fill:         fill,
		Stroke:       stroke,
		LineWidth:    linewidth,
		Filter:       filter,
		FilterRadius: radius,
	}
}

// RasterizationMode controls artist-level mixed raster/vector intent.
type RasterizationMode uint8

const (
	RasterizeDefault RasterizationMode = iota
	RasterizeNever
	RasterizeAlways
	RasterizeAuto
)

// Rasterization carries mixed raster/vector output policy for a draw group.
type Rasterization struct {
	Mode RasterizationMode
	DPI  float64
}

// AntialiasMode controls path antialiasing behavior.
type AntialiasMode uint8

const (
	AntialiasDefault AntialiasMode = iota
	AntialiasOn
	AntialiasOff
)

// SnapMode controls whether path vertices are aligned to device pixels.
type SnapMode uint8

const (
	// SnapOff disables path snapping.
	SnapOff SnapMode = iota
	// SnapAuto snaps simple horizontal/vertical paths, matching Matplotlib's
	// default path snap mode when callers opt into it.
	SnapAuto
	// SnapOn forces snapping for all path vertices.
	SnapOn
)

// SketchParams describes Matplotlib-style sketch/jitter rendering.
//
// Backends may ignore this until they implement a sketch pass.
type SketchParams struct {
	Scale      float64
	Length     float64
	Randomness float64
}

// LineJoin controls how path joins are rendered.
type LineJoin uint8

const (
	JoinMiter LineJoin = iota
	JoinRound
	JoinBevel
)

// LineCap controls how path endpoints are rendered.
type LineCap uint8

const (
	CapButt LineCap = iota
	CapRound
	CapSquare
)

// Color is a normalized sRGBA tuple in [0..1].
//
// In this codebase, callers generally provide plotting-style colors directly
// (for example Matplotlib face/edge tuples), so these channels are interpreted
// as display-encoded sRGB values with straight alpha unless a backend states
// otherwise.
type Color struct{ R, G, B, A float64 }

// Premultiply returns a color with RGB components premultiplied by alpha.
//
// The operation is applied to the stored channel values as-is. It does not
// perform any transfer-curve conversion.
func (c Color) Premultiply() Color {
	return Color{
		R: c.R * c.A,
		G: c.G * c.A,
		B: c.B * c.A,
		A: c.A,
	}
}

// ToPremultipliedRGBA converts normalized sRGBA values to 8-bit premultiplied
// RGBA suitable for raster backends in this repository.
func (c Color) ToPremultipliedRGBA() (uint8, uint8, uint8, uint8) {
	premul := c.Premultiply()
	return uint8(premul.R*255 + 0.5),
		uint8(premul.G*255 + 0.5),
		uint8(premul.B*255 + 0.5),
		uint8(premul.A*255 + 0.5)
}

// Glyph represents a single shaped glyph.
type Glyph struct {
	ID      uint32
	Advance float64
	Offset  geom.Pt
}

// GlyphRun represents a run of glyphs to render at a baseline origin.
type GlyphRun struct {
	Glyphs  []Glyph
	Origin  geom.Pt
	Size    float64
	FontKey string
}

// TextMetrics provides basic text measurements.
type TextMetrics struct{ W, H, Ascent, Descent float64 }

// FontHeightMetrics describes font-wide vertical line metrics for a text run.
// These are distinct from the actual ink extents of a specific string.
type FontHeightMetrics struct{ Ascent, Descent, LineGap float64 }

// TextBounds describes the rendered ink bounds of a string relative to the
// baseline origin used for DrawText.
type TextBounds struct{ X, Y, W, H float64 }

// Image is a minimal interface for raster images passed to renderers.
type Image interface {
	Size() (w, h int)
	// Interpolation returns the resampling filter name (e.g. "bilinear",
	// "bicubic"). Empty string means renderer default (typically nearest).
	Interpolation() string
}

// ImageAlpha exposes an optional per-image scalar alpha multiplier.
//
// The value is interpreted in [0,1] and applied by renderers as a
// pre-composite multiplier to RGBA image output. Keeping this as optional
// avoids forcing all image types to implement alpha handling when they have no
// caller-level multiplier.
type ImageAlpha interface {
	Alpha() float64
}

// RGBAImage is an optional renderer-facing image interface that exposes direct
// RGBA pixel access. Renderers may use this for efficient conversion and
// image-space transforms.
type RGBAImage interface {
	Image
	RGBA() *image.RGBA
}

// JPEGImage is an optional renderer-facing image interface that exposes an
// already-encoded JPEG payload. Vector backends can embed this directly using
// PDF DCTDecode or equivalent mechanisms without round-tripping through RGBA.
type JPEGImage interface {
	Image
	JPEGData() []byte
}

// ImageData is a concrete raster image implementation that satisfies both Image
// and RGBAImage. It is the primary raster type used by heatmaps and image
// artists in this package.
type ImageData struct {
	rgba          *image.RGBA
	interpolation string
	alpha         float64
	hasAlpha      bool
}

// NewImageData creates a new ImageData from an RGBA source image.
// A nil image results in an empty image.
func NewImageData(img *image.RGBA) *ImageData {
	if img == nil {
		return &ImageData{rgba: image.NewRGBA(image.Rect(0, 0, 0, 0))}
	}
	return &ImageData{
		rgba:     img,
		alpha:    1,
		hasAlpha: true,
	}
}

// Size returns the raster dimensions in pixels.
func (i *ImageData) Size() (w, h int) {
	if i == nil || i.rgba == nil {
		return 0, 0
	}
	return i.rgba.Bounds().Dx(), i.rgba.Bounds().Dy()
}

// RGBA returns the backing RGBA image.
func (i *ImageData) RGBA() *image.RGBA {
	if i == nil {
		return nil
	}
	return i.rgba
}

// Interpolation returns the resampling filter name (empty = renderer default).
func (i *ImageData) Interpolation() string {
	if i == nil {
		return ""
	}
	return i.interpolation
}

// SetInterpolation sets the resampling filter name. Empty resets to default.
func (i *ImageData) SetInterpolation(name string) {
	if i == nil {
		return
	}
	i.interpolation = name
}

// Alpha returns the image-level scalar alpha multiplier in [0, 1].
func (i *ImageData) Alpha() float64 {
	if i == nil || !i.hasAlpha {
		return 1
	}
	return i.alpha
}

// SetAlpha sets the image-level scalar alpha multiplier in [0, 1].
func (i *ImageData) SetAlpha(alpha float64) {
	if i == nil {
		return
	}
	i.alpha = clamp01(alpha)
	i.hasAlpha = true
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Renderer defines the core drawing verbs.
type Renderer interface {
	Begin(viewport geom.Rect) error
	End() error

	// State stack
	Save()
	Restore()

	// Clipping
	ClipRect(r geom.Rect)
	ClipPath(p geom.Path)

	// Drawing
	Path(p geom.Path, paint *Paint)
	Image(img Image, dst geom.Rect)
	GlyphRun(run GlyphRun, color Color)
	MeasureText(text string, size float64, fontKey string) TextMetrics
}

// NullRenderer is a no-op renderer used for traversal/tests.
type NullRenderer struct {
	began  bool
	stack  int
	cstack int
}

var _ Renderer = (*NullRenderer)(nil)

// Begin starts a drawing session for the given viewport.
func (n *NullRenderer) Begin(_ geom.Rect) error {
	if n.began {
		return errors.New("Begin called twice")
	}
	n.began = true
	return nil
}

// End ends a drawing session.
func (n *NullRenderer) End() error {
	if !n.began {
		return errors.New("End called before Begin")
	}
	n.began = false
	n.stack = 0
	n.cstack = 0
	return nil
}

// Save pushes state.
func (n *NullRenderer) Save() { n.stack++ }

// Restore pops state; underflow is clamped to zero.
func (n *NullRenderer) Restore() {
	if n.stack > 0 {
		n.stack--
	}
}

// ClipRect pushes a rectangular clip.
func (n *NullRenderer) ClipRect(_ geom.Rect) { n.cstack++ }

// ClipPath pushes a path clip.
func (n *NullRenderer) ClipPath(_ geom.Path) { n.cstack++ }

// Path draws a path using the provided paint; no-op here.
func (n *NullRenderer) Path(_ geom.Path, _ *Paint) {}

// Image draws an image in the destination rectangle; no-op here.
func (n *NullRenderer) Image(_ Image, _ geom.Rect) {}

// GlyphRun draws a run of glyphs with the given color; no-op here.
func (n *NullRenderer) GlyphRun(_ GlyphRun, _ Color) {}

// MeasureText returns zero metrics in the null renderer.
func (n *NullRenderer) MeasureText(_ string, _ float64, _ string) TextMetrics { return TextMetrics{} }

// depth returns the current state stack depth (for tests).
func (n *NullRenderer) depth() int { return n.stack }

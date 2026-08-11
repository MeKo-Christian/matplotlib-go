package render

import (
	"image"
	"strings"
	"time"

	"github.com/cwbudde/matplotlib-go/geom"
)

// DPIAware is implemented by renderers that adapt text/layout behavior to DPI.
type DPIAware interface {
	SetResolution(dpi uint)
}

// SketchAware is implemented by renderers that honor a default sketch/xkcd path
// perturbation. The default applies to every path whose paint does not carry an
// explicit (per-artist) Sketch; a zero-value SketchParams disables it.
type SketchAware interface {
	SetDefaultSketch(params SketchParams)
}

// URLMarker is implemented by vector backends that emit clickable or
// identifiable output. It carries the active hyperlink target (url) and element
// id (gid) that apply to subsequent draws, mirroring matplotlib's
// GraphicsContext.get_url/get_gid. Higher-level code (see core.drawArtist) sets
// these around an artist's draw and restores the previous values afterwards;
// backends stamp the current values onto each emitted element (SVG wraps
// elements in <a xlink:href> and stamps id=; PDF emits /Link annotations).
type URLMarker interface {
	SetURL(url string)
	URL() string
	SetGID(gid string)
	GID() string
}

// TransparentClearer is implemented by raster renderers that can reset their
// buffer pixels to a fully transparent state mid-frame, without disturbing the
// current drawing/clip stack. Matplotlib renders the figure background as a
// real, non-antialiased figure patch composited over a transparent RGBA buffer
// (not an opaque clear); reproducing that lets a sketch/xkcd wiggle perforate
// the canvas border exactly as the reference does.
type TransparentClearer interface {
	ClearTransparent()
}

// SketchActive reports whether sketch parameters will perturb a path. Matplotlib
// treats scale==0 as "no sketch".
func SketchActive(p SketchParams) bool { return p.Scale != 0 }

// EffectiveSketch resolves the sketch parameters for a draw: an explicit
// per-paint value wins, otherwise the renderer's default applies.
func EffectiveSketch(paint, def SketchParams) SketchParams {
	if paint != (SketchParams{}) {
		return paint
	}
	return def
}

// DPIProvider is implemented by renderers that expose their current DPI, letting
// mathtext layout compute matplotlib's exact fontsize*dpi/72 thickness.
type DPIProvider interface {
	Resolution() uint
}

// TextDrawer is implemented by renderers that support direct text drawing.
type TextDrawer interface {
	DrawText(text string, origin geom.Pt, size float64, textColor Color)
}

// FontTextDrawer is implemented by renderers that can draw text with an
// explicit font key instead of relying on prior measurement calls to configure
// renderer-local font state.
type FontTextDrawer interface {
	TextDrawer
	DrawTextWithFont(text string, origin geom.Pt, size float64, textColor Color, fontKey string)
}

// RotatedTextDrawer is implemented by renderers that support rotated text.
type RotatedTextDrawer interface {
	TextDrawer
	// DrawTextRotated matches Matplotlib's default y-axis label anchoring:
	// the point is the bottom-center anchor of the unrotated text box, and the
	// text is then rotated around that anchor.
	DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor Color)
}

// FontRotatedTextDrawer is implemented by renderers that can draw rotated text
// with an explicit font key.
type FontRotatedTextDrawer interface {
	RotatedTextDrawer
	DrawTextRotatedWithFont(text string, anchor geom.Pt, size, angle float64, textColor Color, fontKey string)
}

// TextBounder is implemented by renderers that can report the actual ink bounds
// of text relative to the baseline origin used for DrawText.
type TextBounder interface {
	MeasureTextBounds(text string, size float64, fontKey string) (TextBounds, bool)
}

// MathGlyphMetric carries matplotlib `_mathtext.TruetypeFonts._get_info` metrics
// for a single glyph, used to lay out and rasterize mathtext pixel-exactly.
// All values are in device pixels relative to the glyph baseline (y-up): Iceberg
// = horiBearingY/64 (the TeX "height"), Ymax/Ymin the ink bbox top/bottom,
// Advance the UNHINTED linearHoriAdvance, KernToPrev the kerning to the previous
// glyph in the run.
type MathGlyphMetric struct {
	Advance    float64
	Iceberg    float64
	Height     float64 // glyph.metrics.height/64 (full ink height)
	Xmin       float64
	Xmax       float64
	Ymin       float64
	Ymax       float64
	KernToPrev float64
}

// MathGlyphMeasurer is implemented by renderers that can return matplotlib's
// exact per-glyph mathtext metrics (`_get_info`). Returns false when the
// pixel-exact (FreeType) path is unavailable (e.g. purego/WASM), in which case
// the mathtext layout falls back to whole-run MeasureText.
type MathGlyphMeasurer interface {
	MeasureMathGlyphRun(text string, size float64, fontKey string) ([]MathGlyphMetric, bool)
}

// MathGlyphPlacement is one positioned glyph in a flattened mathtext expression,
// in matplotlib `ship` coordinates (baseline-relative, y-up): Ox/Oy are the glyph
// origin (Oy is the baseline; the glyph ink rises to Oy+Iceberg).
type MathGlyphPlacement struct {
	Text     string
	FontSize float64
	FontKey  string
	Ox       float64
	Oy       float64
}

// MathRectPlacement is one filled rule (fraction bar, radical vinculum) in ship
// coordinates (baseline-relative, y-up), X1<X2, Y1<Y2.
type MathRectPlacement struct {
	X1, Y1, X2, Y2 float64
}

// MathTextImageDrawer is implemented by renderers that can rasterize a flattened
// mathtext expression with matplotlib's `_mathtext.Output.to_raster` pixel
// placement (per-glyph integer blitting + the `draw_rect_filled` rule formula,
// sharing one bounding box). anchor is the expression baseline origin in display
// space. Returns false when the pixel-exact path is unavailable.
type MathTextImageDrawer interface {
	// boxAscent/boxDescent are the expression's layout ascent/descent (matplotlib
	// box.height/box.depth), needed to reproduce to_raster's image height and the
	// backend's round(baseline+descent)+1 placement.
	DrawMathTextImage(glyphs []MathGlyphPlacement, rects []MathRectPlacement, anchor geom.Pt, boxAscent, boxDescent float64, textColor Color) bool
}

// RotatedMathTextImageDrawer is implemented by raster backends that can rotate
// the flattened mathtext image the same way Matplotlib's AGG draw_text_image
// path does. origin is the expression baseline-left origin in display space.
type RotatedMathTextImageDrawer interface {
	DrawMathTextImageRotated(glyphs []MathGlyphPlacement, rects []MathRectPlacement, origin geom.Pt, boxAscent, boxDescent, angle float64, textColor Color) bool
}

// TextFontMetricer is implemented by renderers that can report font-wide line
// metrics separately from the ink bounds of a particular string.
type TextFontMetricer interface {
	MeasureFontHeights(size float64, fontKey string) (FontHeightMetrics, bool)
}

// TextPather is implemented by renderers that can convert text to vector paths.
type TextPather interface {
	TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool)
}

// TeXMetricer is implemented by renderers that can measure text through a
// TeX/LaTeX pipeline when text.usetex is enabled.
type TeXMetricer interface {
	MeasureTeX(text string, size float64, fontKey string) (TextMetrics, bool)
}

// TeXDrawer is implemented by renderers that can draw text through a
// TeX/LaTeX pipeline when text.usetex is enabled. Returning false asks core to
// fall back to the normal text path.
type TeXDrawer interface {
	DrawTeX(text string, origin geom.Pt, size float64, textColor Color, fontKey string) bool
}

// RotatedTeXDrawer is implemented by renderers that can draw TeX output with a
// rotation around the same anchor used by RotatedTextDrawer.
type RotatedTeXDrawer interface {
	TeXDrawer
	DrawTeXRotated(text string, anchor geom.Pt, size, angle float64, textColor Color, fontKey string) bool
}

// VerticalTextDrawer is implemented by renderers that support vertical text.
type VerticalTextDrawer interface {
	TextDrawer
	DrawTextVertical(text string, center geom.Pt, size float64, textColor Color)
}

// FontVerticalTextDrawer is implemented by renderers that can draw vertical
// text with an explicit font key.
type FontVerticalTextDrawer interface {
	VerticalTextDrawer
	DrawTextVerticalWithFont(text string, center geom.Pt, size float64, textColor Color, fontKey string)
}

// ImageTransformer is implemented by renderers that support affine image transforms.
type ImageTransformer interface {
	ImageTransformed(img Image, dst geom.Rect, transform geom.Affine)
}

// BboxImageDrawer draws an image whose output size is determined by a display
// bbox, matching Matplotlib's BboxImage/OffsetImage integer placement rules.
type BboxImageDrawer interface {
	DrawBboxImage(img Image, bbox geom.Rect) bool
}

// ClipPathTransformer is implemented by renderers that can apply an affine
// transform directly to a path-based clip definition.
type ClipPathTransformer interface {
	ClipPathTransformed(path geom.Path, transform geom.Affine)
}

// RGBAExporter is implemented by raster renderers that expose direct RGBA
// buffer access in display pixel order.
type RGBAExporter interface {
	Image() *image.RGBA
}

// BufferRegion holds copied pixels and their destination rectangle in renderer
// (device) coordinates. It is the renderer-neutral equivalent of Matplotlib's
// BufferRegion used by copy_from_bbox / restore_region; Rect matches matplotlib's
// BufferRegion.get_extents (already y-flipped into the device buffer).
type BufferRegion struct {
	Image *image.RGBA
	Rect  geom.Rect
}

// BufferRegioner is implemented by renderers that support blitting-style
// region copy and restore.
//
// CopyFromBBox's bbox and RestoreRegion's bbox/offset are all expressed in y-up
// display space; the renderer flips them into its device buffer. For
// RestoreRegion, bbox is an absolute display rectangle selecting which portion of
// the captured region to restore (nil restores all of it), and offset is a
// display-space translation delta (+offset.Y moves the restored content up).
type BufferRegioner interface {
	CopyFromBBox(bbox geom.Rect) *BufferRegion
	RestoreRegion(region *BufferRegion, bbox *geom.Rect, offset geom.Pt)
}

// FilterRenderer is implemented by renderers that support drawing into a
// temporary offscreen surface and compositing a post-processed result.
type FilterRenderer interface {
	StartFilter()
	StopFilter(postProcess func(img *image.RGBA, dpi float64) (*image.RGBA, geom.Pt))
}

// PathEffectFilterDrawer is implemented by vector renderers that can apply a
// filter path effect without rasterizing the filtered pass.
type PathEffectFilterDrawer interface {
	SupportsPathEffectFilter(effect PathEffect) bool
	DrawPathEffectFilter(path geom.Path, paint Paint, effect PathEffect, draw func(geom.Path, *Paint)) bool
}

// PatternFiller is implemented by renderers that consume Paint.FillPattern
// natively instead of relying on renderer-neutral fallback expansion.
type PatternFiller interface {
	SupportsPatternFill() bool
}

// GradientFiller is implemented by renderers that consume Paint.FillGradient
// natively instead of relying on renderer-neutral fallback expansion.
type GradientFiller interface {
	SupportsGradientFill() bool
}

// PathEffectDrawer is implemented by renderers that can apply path effects as
// native post-stroke/post-fill rendering passes.
type PathEffectDrawer interface {
	DrawPathWithEffects(path geom.Path, paint *Paint) bool
}

// RasterizationController is implemented by vector-capable renderers that can
// bracket a draw group into a raster sub-output for mixed raster/vector files.
type RasterizationController interface {
	StartRasterized(options Rasterization) bool
	StopRasterized() bool
}

// MarkerItem is one positioned marker instance in a repeated marker batch.
// Transform is applied to Marker first, then Offset is added in display space.
type MarkerItem struct {
	Offset      geom.Pt
	Transform   geom.Affine
	Paint       Paint
	Snap        SnapMode
	SnapSet     bool
	Antialiased bool
}

// MarkerBatch describes one marker path rendered at many display-space
// positions.
type MarkerBatch struct {
	Marker geom.Path
	Items  []MarkerItem
}

// MarkerDrawer is implemented by renderers with a native repeated-marker path.
type MarkerDrawer interface {
	DrawMarkers(batch MarkerBatch) bool
}

// PathCollectionItem is one display-space path with its paint state.
type PathCollectionItem struct {
	Path         geom.Path
	Paint        Paint
	Hatch        string
	HatchColor   Color
	HatchWidth   float64
	HatchSpacing float64
	Antialiased  bool
}

// PathCollectionBatch describes many display-space paths rendered as one
// collection operation.
type PathCollectionBatch struct {
	Items []PathCollectionItem
}

// PathCollectionDrawer is implemented by renderers with a native collection
// path.
type PathCollectionDrawer interface {
	DrawPathCollection(batch PathCollectionBatch) bool
}

// NativeHatcher is implemented by renderers that consume Paint hatch metadata
// directly during Path rendering.
type NativeHatcher interface {
	SupportsNativeHatch() bool
}

// QuadMeshCell is one display-space quadrilateral cell.
type QuadMeshCell struct {
	Quad         [4]geom.Pt
	Face         Color
	Edge         Color
	LineWidth    float64
	Dashes       []float64
	Hatch        string
	HatchColor   Color
	HatchWidth   float64
	HatchSpacing float64
	Antialiased  bool
	Snap         SnapMode
}

// QuadMeshBatch describes pcolor/pcolormesh-style quadrilateral cells.
type QuadMeshBatch struct {
	Cells []QuadMeshCell
}

// QuadMeshDrawer is implemented by renderers with a native quad mesh path.
type QuadMeshDrawer interface {
	DrawQuadMesh(batch QuadMeshBatch) bool
}

// GouraudTriangle describes one triangle with per-vertex display-space colors.
type GouraudTriangle struct {
	P     [3]geom.Pt
	Color [3]Color
}

// GouraudTriangleBatch describes interpolated-color triangles.
type GouraudTriangleBatch struct {
	Triangles   []GouraudTriangle
	Antialiased bool
}

// GouraudTriangleDrawer is implemented by renderers with native Gouraud
// triangle shading.
type GouraudTriangleDrawer interface {
	DrawGouraudTriangles(batch GouraudTriangleBatch) bool
}

// PNGExporter is implemented by renderers that can export their output to PNG.
type PNGExporter interface {
	SavePNG(path string) error
}

// SaveOption is the shared public option surface for extension-driven save
// APIs. Format-specific option functions such as WithSVGMetadata,
// WithPDFMetadata, WithPSFontPolicy, and WithPGFPreamble implement this
// interface directly, so callers can pass typed options through SaveFig and
// registry save dispatch without backend-name conditionals.
type SaveOption interface {
	applySaveOptions(*SaveOptions)
}

// SaveOptions is a resolved bag of per-format save options.
type SaveOptions struct {
	SVG SVGOptions
	PDF PDFOptions
	PS  PSOptions
	PGF PGFOptions
	// Figure holds format-agnostic save-time options (savefig.* rcParams).
	Figure FigureOptions

	hasSVG    bool
	hasPDF    bool
	hasPS     bool
	hasPGF    bool
	hasFigure bool
}

// FigureOptions holds format-agnostic, save-time figure options mirroring
// Matplotlib's savefig.* rcParams. The Has* flags record which values were set
// explicitly by the caller so they can override rcParams defaults.
type FigureOptions struct {
	// DPI overrides the render resolution at save time (savefig.dpi); 0 means
	// "use the figure DPI".
	DPI float64
	// HasDPI reports whether DPI was set explicitly.
	HasDPI bool
	// Facecolor overrides the figure face color (savefig.facecolor).
	Facecolor Color
	// HasFacecolor reports whether Facecolor was set explicitly.
	HasFacecolor bool
	// Edgecolor overrides the figure edge color (savefig.edgecolor).
	Edgecolor Color
	// HasEdgecolor reports whether Edgecolor was set explicitly.
	HasEdgecolor bool
	// Transparent renders the figure (and axes) backgrounds transparent
	// (savefig.transparent).
	Transparent bool
	// HasTransparent reports whether Transparent was set explicitly.
	HasTransparent bool
	// BboxInches selects the saved bounding box (savefig.bbox): "" (standard) or
	// "tight".
	BboxInches string
	// HasBbox reports whether BboxInches was set explicitly.
	HasBbox bool
	// PadInches is the padding around a tight bbox, in inches (savefig.pad_inches).
	PadInches float64
	// HasPad reports whether PadInches was set explicitly.
	HasPad bool
	// Format overrides the output format independent of the path extension
	// (savefig.format).
	Format string
}

// FigureOption configures format-agnostic save-time options.
type FigureOption func(*FigureOptions)

func (f FigureOption) applySaveOptions(dst *SaveOptions) {
	if f == nil {
		return
	}
	f(&dst.Figure)
	dst.hasFigure = true
}

// WithSaveDPI overrides the render resolution at save time (savefig.dpi).
func WithSaveDPI(dpi float64) FigureOption {
	return func(opts *FigureOptions) {
		opts.DPI = dpi
		opts.HasDPI = true
	}
}

// WithSaveFacecolor overrides the figure face color at save time
// (savefig.facecolor).
func WithSaveFacecolor(c Color) FigureOption {
	return func(opts *FigureOptions) {
		opts.Facecolor = c
		opts.HasFacecolor = true
	}
}

// WithSaveEdgecolor overrides the figure edge color at save time
// (savefig.edgecolor).
func WithSaveEdgecolor(c Color) FigureOption {
	return func(opts *FigureOptions) {
		opts.Edgecolor = c
		opts.HasEdgecolor = true
	}
}

// WithSaveTransparent renders transparent figure/axes backgrounds at save time
// (savefig.transparent).
func WithSaveTransparent(transparent bool) FigureOption {
	return func(opts *FigureOptions) {
		opts.Transparent = transparent
		opts.HasTransparent = true
	}
}

// WithSaveBboxInches selects the saved bounding box (savefig.bbox): "" for the
// standard box or "tight" to crop to content.
func WithSaveBboxInches(bbox string) FigureOption {
	return func(opts *FigureOptions) {
		opts.BboxInches = bbox
		opts.HasBbox = true
	}
}

// WithSavePadInches sets the padding around a tight bbox in inches
// (savefig.pad_inches).
func WithSavePadInches(pad float64) FigureOption {
	return func(opts *FigureOptions) {
		opts.PadInches = pad
		opts.HasPad = true
	}
}

// WithSaveFormat overrides the output format independent of the path extension
// (savefig.format).
func WithSaveFormat(format string) FigureOption {
	return func(opts *FigureOptions) {
		opts.Format = format
	}
}

// ResolveSaveOptions applies shared save options onto deterministic per-format
// defaults.
func ResolveSaveOptions(opts ...SaveOption) SaveOptions {
	cfg := SaveOptions{
		SVG: DefaultSVGOptions(),
		PDF: DefaultPDFOptions(),
		PS:  DefaultPSOptions(),
		PGF: DefaultPGFOptions(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt.applySaveOptions(&cfg)
		}
	}
	cfg.SVG.Metadata = cloneSVGMetadata(cfg.SVG.Metadata)
	cfg.PDF.Metadata = clonePDFMetadata(cfg.PDF.Metadata)
	cfg.PGF.Metadata = clonePGFMetadata(cfg.PGF.Metadata)
	cfg.PGF.Preamble = cloneStrings(cfg.PGF.Preamble)
	return cfg
}

// HasSVGOptions reports whether the caller supplied any SVG-specific option.
func (opts SaveOptions) HasSVGOptions() bool { return opts.hasSVG }

// HasPDFOptions reports whether the caller supplied any PDF-specific option.
func (opts SaveOptions) HasPDFOptions() bool { return opts.hasPDF }

// HasPSOptions reports whether the caller supplied any PostScript-specific option.
func (opts SaveOptions) HasPSOptions() bool { return opts.hasPS }

// HasPGFOptions reports whether the caller supplied any PGF-specific option.
func (opts SaveOptions) HasPGFOptions() bool { return opts.hasPGF }

// ValidateForExtension rejects format-specific options that do not match the
// output extension. This keeps unsupported options from being silently ignored.
func (opts SaveOptions) ValidateForExtension(ext string) error {
	normalized := strings.ToLower(strings.TrimSpace(ext))
	if normalized != "" && normalized[0] != '.' {
		normalized = "." + normalized
	}

	var allowed string
	switch normalized {
	case ".svg":
		allowed = "svg"
	case ".pdf":
		allowed = "pdf"
	case ".ps", ".eps":
		allowed = "ps"
	case ".pgf":
		allowed = "pgf"
	case ".png":
		allowed = "png"
	}

	if opts.hasSVG && allowed != "svg" {
		return unsupportedSaveOptionError(normalized, "SVG")
	}
	if opts.hasPDF && allowed != "pdf" {
		return unsupportedSaveOptionError(normalized, "PDF")
	}
	if opts.hasPS && allowed != "ps" {
		return unsupportedSaveOptionError(normalized, "PostScript")
	}
	if opts.hasPGF && allowed != "pgf" {
		return unsupportedSaveOptionError(normalized, "PGF")
	}
	return nil
}

func unsupportedSaveOptionError(ext, name string) error {
	if ext == "" {
		ext = "<none>"
	}
	return &SaveOptionError{Extension: ext, OptionSet: name}
}

// SaveOptionError reports a mismatched save option and target extension.
type SaveOptionError struct {
	Extension string
	OptionSet string
}

func (e *SaveOptionError) Error() string {
	return e.OptionSet + " options are not supported for " + e.Extension + " output"
}

func (opts SaveOptions) applySaveOptions(dst *SaveOptions) {
	if opts.hasSVG {
		dst.SVG = opts.SVG
		dst.hasSVG = true
	}
	if opts.hasPDF {
		dst.PDF = opts.PDF
		dst.hasPDF = true
	}
	if opts.hasPS {
		dst.PS = opts.PS
		dst.hasPS = true
	}
	if opts.hasPGF {
		dst.PGF = opts.PGF
		dst.hasPGF = true
	}
	if opts.hasFigure {
		dst.Figure = opts.Figure
		dst.hasFigure = true
	}
}

// BackgroundClearer is implemented by renderers that can replace their page or
// surface background before drawing. The save pipeline uses it to apply
// savefig.facecolor and savefig.transparent.
type BackgroundClearer interface {
	Clear(Color)
}

// Resizer is implemented by raster renderers that can reallocate their surface
// to a new pixel size. The save pipeline uses it to apply savefig.dpi on raster
// output (where the canvas grows with resolution).
type Resizer interface {
	Resize(width, height int) error
}

// VectorPageSizer is implemented by vector renderers whose physical page can
// have fractional PostScript-point dimensions (1 point = 1/72 inch).
type VectorPageSizer interface {
	SetPageSize(widthPoints, heightPoints float64)
}

// SVGExporter is implemented by renderers that can export their output to SVG.
type SVGExporter interface {
	SaveSVG(path string) error
}

// PDFExporter is implemented by renderers that can export their output to PDF.
type PDFExporter interface {
	SavePDF(path string) error
}

// PSExporter is implemented by renderers that can export their output to
// PostScript or EPS.
type PSExporter interface {
	SavePS(path string) error
}

// PSFontPolicy controls whether PostScript text is emitted as filled glyph
// outlines or through a built-in Base14 font.
type PSFontPolicy string

const (
	// PSFontPolicyPath converts text to filled glyph outlines through the
	// shared font manager. This mirrors the deterministic PDF default and
	// avoids claiming font embedding/searchability that Level-2 PS cannot
	// provide without a Type 3/Type 42 embedding layer.
	PSFontPolicyPath PSFontPolicy = "path"
	// PSFontPolicyBase14 emits direct PostScript text through Helvetica.
	// Non-ASCII runes are replaced because no custom font is embedded.
	PSFontPolicyBase14 PSFontPolicy = "base14"
)

// PSOptions carries PostScript-specific renderer behavior knobs.
type PSOptions struct {
	FontPolicy PSFontPolicy
}

// PSOption mutates PSOptions.
type PSOption func(*PSOptions)

func (opt PSOption) applySaveOptions(opts *SaveOptions) {
	opts.hasPS = true
	if opt != nil {
		opt(&opts.PS)
	}
}

// DefaultPSOptions returns deterministic PostScript defaults.
func DefaultPSOptions() PSOptions {
	return PSOptions{FontPolicy: PSFontPolicyPath}
}

// ResolvePSOptions applies opts onto deterministic PS defaults.
func ResolvePSOptions(opts ...PSOption) PSOptions {
	cfg := DefaultPSOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// WithPSFontPolicy configures glyph-path versus Base14 direct text output.
func WithPSFontPolicy(policy PSFontPolicy) PSOption {
	return func(cfg *PSOptions) {
		switch policy {
		case PSFontPolicyBase14:
			cfg.FontPolicy = PSFontPolicyBase14
		default:
			cfg.FontPolicy = PSFontPolicyPath
		}
	}
}

// PSOptionSetter is implemented by renderers whose draw-time behavior depends
// on PSOptions, such as the text output policy.
type PSOptionSetter interface {
	SetPSOptions(opts PSOptions)
}

// PGFExporter is implemented by renderers that can export their output to
// PGF/TikZ for inclusion in LaTeX documents.
type PGFExporter interface {
	SavePGF(path string) error
}

// PGFCommentPolicy controls whether generator comments are retained.
type PGFCommentPolicy string

const (
	PGFCommentPolicyKeep  PGFCommentPolicy = "keep"
	PGFCommentPolicyStrip PGFCommentPolicy = "strip"
)

// PGFVerificationMode reserves a shared option for future TeX verification
// checks while keeping the public save API stable.
type PGFVerificationMode string

const (
	PGFVerificationModeNone   PGFVerificationMode = "none"
	PGFVerificationModeStrict PGFVerificationMode = "strict"
)

// PGFOptions carries PGF/TikZ-specific export behavior knobs.
type PGFOptions struct {
	Metadata         map[string]string
	Preamble         []string
	CommentPolicy    PGFCommentPolicy
	VerificationMode PGFVerificationMode
}

// PGFOption mutates PGFOptions.
type PGFOption func(*PGFOptions)

func (opt PGFOption) applySaveOptions(opts *SaveOptions) {
	opts.hasPGF = true
	if opt != nil {
		opt(&opts.PGF)
	}
}

// DefaultPGFOptions returns deterministic PGF defaults.
func DefaultPGFOptions() PGFOptions {
	return PGFOptions{
		CommentPolicy:    PGFCommentPolicyKeep,
		VerificationMode: PGFVerificationModeNone,
	}
}

// ResolvePGFOptions applies opts onto deterministic PGF defaults.
func ResolvePGFOptions(opts ...PGFOption) PGFOptions {
	cfg := DefaultPGFOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	cfg.Metadata = clonePGFMetadata(cfg.Metadata)
	cfg.Preamble = cloneStrings(cfg.Preamble)
	return cfg
}

// WithPGFMetadata adds deterministic metadata comments to the PGF document.
func WithPGFMetadata(metadata map[string]string) PGFOption {
	return func(cfg *PGFOptions) {
		cfg.Metadata = clonePGFMetadata(metadata)
	}
}

// WithPGFPreamble records TeX preamble lines needed by the PGF snippet.
func WithPGFPreamble(lines ...string) PGFOption {
	return func(cfg *PGFOptions) {
		cfg.Preamble = cloneStrings(lines)
	}
}

// WithPGFCommentPolicy configures whether PGF generator comments are emitted.
func WithPGFCommentPolicy(policy PGFCommentPolicy) PGFOption {
	return func(cfg *PGFOptions) {
		switch policy {
		case PGFCommentPolicyStrip:
			cfg.CommentPolicy = PGFCommentPolicyStrip
		default:
			cfg.CommentPolicy = PGFCommentPolicyKeep
		}
	}
}

// WithPGFVerificationMode records the desired future TeX verification policy.
func WithPGFVerificationMode(mode PGFVerificationMode) PGFOption {
	return func(cfg *PGFOptions) {
		switch mode {
		case PGFVerificationModeStrict:
			cfg.VerificationMode = PGFVerificationModeStrict
		default:
			cfg.VerificationMode = PGFVerificationModeNone
		}
	}
}

// PGFOptionExporter is implemented by PGF renderers that accept resolved
// options at export time.
type PGFOptionExporter interface {
	SavePGFWithOptions(path string, opts PGFOptions) error
}

// PGFOptionSetter is implemented by renderers whose draw-time behavior depends
// on PGFOptions, such as generator comment output.
type PGFOptionSetter interface {
	SetPGFOptions(opts PGFOptions)
}

// PDFFontPolicy controls whether PDF text is emitted using embedded fonts or
// converted to filled glyph paths.
type PDFFontPolicy string

const (
	// PDFFontPolicyPath converts text to filled glyph outlines. This is the
	// default for the initial PDF backend implementation because it does not
	// require embedded font subsetting.
	PDFFontPolicyPath PDFFontPolicy = "path"
	// PDFFontPolicyEmbed embeds subsetted fonts and emits real PDF text
	// objects.
	PDFFontPolicyEmbed PDFFontPolicy = "embed"
)

// PDFOptions carries PDF-specific export and renderer behavior knobs.
type PDFOptions struct {
	FontPolicy PDFFontPolicy
	// Metadata is written into the PDF /Info dictionary. Empty metadata maps
	// produce a deterministic empty /Info entry.
	Metadata map[string]string
	// CreationDate optionally overrides the document creation date. When zero,
	// the renderer reads SOURCE_DATE_EPOCH for reproducible output and falls
	// back to the current time when neither is set.
	CreationDate time.Time
}

// PDFOption mutates PDFOptions.
type PDFOption func(*PDFOptions)

func (opt PDFOption) applySaveOptions(opts *SaveOptions) {
	opts.hasPDF = true
	if opt != nil {
		opt(&opts.PDF)
	}
}

// DefaultPDFOptions returns deterministic PDF defaults.
func DefaultPDFOptions() PDFOptions {
	return PDFOptions{FontPolicy: PDFFontPolicyPath}
}

// ResolvePDFOptions applies opts onto deterministic PDF defaults.
func ResolvePDFOptions(opts ...PDFOption) PDFOptions {
	cfg := DefaultPDFOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	cfg.Metadata = clonePDFMetadata(cfg.Metadata)
	return cfg
}

// WithPDFFontPolicy configures embedded versus path-based text output.
func WithPDFFontPolicy(policy PDFFontPolicy) PDFOption {
	return func(cfg *PDFOptions) {
		switch policy {
		case PDFFontPolicyEmbed:
			cfg.FontPolicy = PDFFontPolicyEmbed
		default:
			cfg.FontPolicy = PDFFontPolicyPath
		}
	}
}

// WithPDFMetadata adds deterministic metadata entries to the PDF /Info
// dictionary.
func WithPDFMetadata(metadata map[string]string) PDFOption {
	return func(cfg *PDFOptions) {
		cfg.Metadata = clonePDFMetadata(metadata)
	}
}

// WithPDFCreationDate overrides the PDF creation date for reproducible output.
func WithPDFCreationDate(t time.Time) PDFOption {
	return func(cfg *PDFOptions) {
		cfg.CreationDate = t
	}
}

// PDFOptionExporter is implemented by PDF renderers that accept resolved
// options at export time.
type PDFOptionExporter interface {
	SavePDFWithOptions(path string, opts PDFOptions) error
}

// PDFOptionSetter is implemented by renderers whose draw-time behavior depends
// on PDFOptions, such as text-as-path policy and metadata.
type PDFOptionSetter interface {
	SetPDFOptions(opts PDFOptions)
}

func clonePDFMetadata(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// SVGFontPolicy controls whether SVG text is emitted as native <text> nodes
// or converted to filled glyph paths.
type SVGFontPolicy string

const (
	SVGFontPolicyNone SVGFontPolicy = "none"
	SVGFontPolicyPath SVGFontPolicy = "path"
)

// SVGOptions carries SVG-specific export and renderer behavior knobs.
type SVGOptions struct {
	FontPolicy SVGFontPolicy
	Metadata   map[string]string
	HashSalt   string
}

// SVGOption mutates SVGOptions.
type SVGOption func(*SVGOptions)

func (opt SVGOption) applySaveOptions(opts *SaveOptions) {
	opts.hasSVG = true
	if opt != nil {
		opt(&opts.SVG)
	}
}

// DefaultSVGOptions returns deterministic SVG defaults.
func DefaultSVGOptions() SVGOptions {
	return SVGOptions{FontPolicy: SVGFontPolicyNone}
}

// ResolveSVGOptions applies opts onto deterministic SVG defaults.
func ResolveSVGOptions(opts ...SVGOption) SVGOptions {
	cfg := DefaultSVGOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	cfg.Metadata = cloneSVGMetadata(cfg.Metadata)
	return cfg
}

// WithSVGFontPolicy configures native text versus glyph-path text output.
func WithSVGFontPolicy(policy SVGFontPolicy) SVGOption {
	return func(cfg *SVGOptions) {
		switch policy {
		case SVGFontPolicyPath:
			cfg.FontPolicy = SVGFontPolicyPath
		default:
			cfg.FontPolicy = SVGFontPolicyNone
		}
	}
}

// WithSVGMetadata adds deterministic metadata entries to the SVG document.
func WithSVGMetadata(metadata map[string]string) SVGOption {
	return func(cfg *SVGOptions) {
		cfg.Metadata = cloneSVGMetadata(metadata)
	}
}

// WithSVGHashSalt enables content-hashed SVG IDs using the provided salt.
func WithSVGHashSalt(salt string) SVGOption {
	return func(cfg *SVGOptions) {
		cfg.HashSalt = salt
	}
}

// SVGOptionExporter is implemented by SVG renderers that accept resolved
// options at export time.
type SVGOptionExporter interface {
	SaveSVGWithOptions(path string, opts SVGOptions) error
}

// SVGOptionSetter is implemented by renderers whose draw-time behavior depends
// on SVGOptions, such as text-as-path and content-hashed IDs.
type SVGOptionSetter interface {
	SetSVGOptions(opts SVGOptions)
}

func cloneSVGMetadata(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func clonePGFMetadata(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

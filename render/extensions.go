package render

import (
	"image"
	"time"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// DPIAware is implemented by renderers that adapt text/layout behavior to DPI.
type DPIAware interface {
	SetResolution(dpi uint)
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

// ClipPathTransformer is implemented by renderers that can apply an affine
// transform directly to a path-based clip definition.
type ClipPathTransformer interface {
	ClipPathTransformed(path geom.Path, transform geom.Affine)
}

// RGBAExporter is implemented by raster renderers that expose direct RGBA
// buffer access in display pixel order.
type RGBAExporter interface {
	GetImage() *image.RGBA
}

// BufferRegion holds copied pixels and their destination rectangle in renderer
// coordinates. It is the renderer-neutral equivalent of Matplotlib's
// BufferRegion used by copy_from_bbox / restore_region.
type BufferRegion struct {
	Image *image.RGBA
	Rect  geom.Rect
}

// BufferRegioner is implemented by renderers that support blitting-style
// region copy and restore.
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

// PGFExporter is implemented by renderers that can export their output to
// PGF/TikZ for inclusion in LaTeX documents.
type PGFExporter interface {
	SavePGF(path string) error
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

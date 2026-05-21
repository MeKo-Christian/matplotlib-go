package pgf

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/backends/internal/mixedraster"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const defaultFontHeight = 13.0

type state struct {
	inContent bool
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

// Renderer implements render.Renderer by emitting PGF/TikZ commands.
type Renderer struct {
	width      int
	height     int
	background render.Color
	resolution uint

	began      bool
	viewport   geom.Rect
	content    strings.Builder
	document   []byte
	stack      []state
	clipRect   *geom.Rect
	clipPaths  []geom.Path
	raster     *mixedraster.Session
	colorNames map[string]string
	pathIDs    map[string]string
	pgfOpts    render.PGFOptions
}

var (
	_ render.Renderer                = (*Renderer)(nil)
	_ render.PGFExporter             = (*Renderer)(nil)
	_ render.DPIAware                = (*Renderer)(nil)
	_ render.ImageTransformer        = (*Renderer)(nil)
	_ render.MarkerDrawer            = (*Renderer)(nil)
	_ render.PathCollectionDrawer    = (*Renderer)(nil)
	_ render.FontTextDrawer          = (*Renderer)(nil)
	_ render.FontRotatedTextDrawer   = (*Renderer)(nil)
	_ render.NativeHatcher           = (*Renderer)(nil)
	_ render.RasterizationController = (*Renderer)(nil)
	_ render.PGFOptionExporter       = (*Renderer)(nil)
	_ render.PGFOptionSetter         = (*Renderer)(nil)
)

// New creates a PGF renderer with a point-sized canvas.
func New(width, height int, background render.Color) (*Renderer, error) {
	if width <= 0 {
		width = 640
	}
	if height <= 0 {
		height = 480
	}
	if background == (render.Color{}) {
		background = render.Color{R: 1, G: 1, B: 1, A: 1}
	}
	return &Renderer{
		width:      width,
		height:     height,
		background: background,
		resolution: 72,
		pgfOpts:    render.DefaultPGFOptions(),
	}, nil
}

// SetPGFOptions implements render.PGFOptionSetter.
func (r *Renderer) SetPGFOptions(opts render.PGFOptions) {
	r.pgfOpts = normalizePGFOptions(opts)
}

// SetResolution implements render.DPIAware.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi == 0 {
		dpi = 72
	}
	r.resolution = dpi
}

// Begin starts a drawing session.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("pgf: Begin called twice")
	}
	r.began = true
	r.viewport = viewport
	r.content.Reset()
	r.document = nil
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipPaths = nil
	r.raster = nil
	r.colorNames = map[string]string{}
	r.pathIDs = map[string]string{}

	r.writeDocumentComments(&r.content)
	r.content.WriteString("\\begingroup\n")
	r.content.WriteString("\\begin{pgfpicture}\n")
	fmt.Fprintf(&r.content, "\\pgfpathrectangle{\\pgfpoint{0pt}{0pt}}{\\pgfpoint{%spt}{%spt}}\n",
		shortFloat(float64(r.width)), shortFloat(float64(r.height)))
	r.content.WriteString("\\pgfusepath{use as bounding box}\n")
	fmt.Fprintf(&r.content, "\\pgftransformcm{1}{0}{0}{-1}{\\pgfpoint{0pt}{%spt}}\n", shortFloat(float64(r.height)))
	if r.background.A > 0 {
		writeFillOpacity(&r.content, r.background.A)
		writeFillColor(&r.content, r.colorName(r.background))
		fmt.Fprintf(&r.content, "\\pgfpathrectangle{\\pgfpoint{0pt}{0pt}}{\\pgfpoint{%spt}{%spt}}\n",
			shortFloat(float64(r.width)), shortFloat(float64(r.height)))
		r.content.WriteString("\\pgfusepath{fill}\n")
	}
	return nil
}

// StartRasterized begins a transparent offscreen raster group for mixed output.
func (r *Renderer) StartRasterized(options render.Rasterization) bool {
	if r == nil || !r.began || r.raster != nil {
		return false
	}
	session, ok := mixedraster.Start(r.width, r.height, r.viewport, options, r.resolution, r.clipRect, r.clipPaths)
	if !ok {
		return false
	}
	r.raster = session
	return true
}

// StopRasterized embeds the active raster group as self-contained PGF pixels.
func (r *Renderer) StopRasterized() bool {
	if r == nil || r.raster == nil {
		return false
	}
	session := r.raster
	r.raster = nil
	img, rect, ok := session.Stop()
	if !ok {
		return false
	}
	r.Image(img, rect)
	return true
}

func (r *Renderer) activeRaster() render.Renderer {
	if r == nil || r.raster == nil {
		return nil
	}
	return r.raster.Renderer()
}

// End finalizes the PGF document.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("pgf: End called before Begin")
	}
	r.began = false
	r.content.WriteString("\\end{pgfpicture}\n")
	r.content.WriteString("\\endgroup\n")
	r.document = []byte(r.content.String())
	return nil
}

// Save pushes graphics state.
func (r *Renderer) Save() {
	if rr := r.activeRaster(); rr != nil {
		rr.Save()
		return
	}
	r.stack = append(r.stack, state{
		inContent: r.began,
		clipRect:  cloneRectPtr(r.clipRect),
		clipPaths: mixedraster.ClonePaths(r.clipPaths),
	})
	if r.began {
		r.content.WriteString("\\pgfscope\n")
	}
}

// Restore pops graphics state.
func (r *Renderer) Restore() {
	if rr := r.activeRaster(); rr != nil {
		rr.Restore()
		return
	}
	if len(r.stack) == 0 {
		return
	}
	top := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	r.clipRect = top.clipRect
	r.clipPaths = mixedraster.ClonePaths(top.clipPaths)
	if top.inContent && r.began {
		r.content.WriteString("\\endpgfscope\n")
	}
}

// ClipRect installs a rectangular clip.
func (r *Renderer) ClipRect(rect geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipRect(rect)
		return
	}
	if !r.began {
		return
	}
	normalized := normalizeRect(rect)
	if r.clipRect == nil {
		r.clipRect = &normalized
	} else {
		intersected := r.clipRect.Intersect(normalized)
		r.clipRect = &intersected
	}
	fmt.Fprintf(&r.content, "\\pgfpathrectangle{\\pgfpoint{%spt}{%spt}}{\\pgfpoint{%spt}{%spt}}\n",
		shortFloat(rect.Min.X), shortFloat(rect.Min.Y), shortFloat(rect.W()), shortFloat(rect.H()))
	r.content.WriteString("\\pgfusepath{clip}\n")
}

// ClipPath installs an arbitrary path clip.
func (r *Renderer) ClipPath(path geom.Path) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipPath(path)
		return
	}
	if !r.began {
		return
	}
	if !writePathOps(&r.content, path) {
		return
	}
	r.clipPaths = append(r.clipPaths, mixedraster.ClonePath(path))
	r.content.WriteString("\\pgfusepath{clip}\n")
}

// Path draws a vector path.
func (r *Renderer) Path(path geom.Path, paint *render.Paint) {
	if rr := r.activeRaster(); rr != nil {
		rr.Path(path, paint)
		return
	}
	if !r.began || paint == nil {
		return
	}
	if render.DrawPathWithEffects(r, path, paint, r.Path) {
		return
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasFill := paint.Fill.A > 0 || hasHatch
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasStroke {
		return
	}

	r.writePathPaintOps(&r.content, path, paint)
}

// DrawPathWithEffects applies renderer-neutral path effect passes.
func (r *Renderer) DrawPathWithEffects(path geom.Path, paint *render.Paint) bool {
	if rr := r.activeRaster(); rr != nil {
		if effects, ok := rr.(render.PathEffectDrawer); ok {
			return effects.DrawPathWithEffects(path, paint)
		}
		rr.Path(path, paint)
		return true
	}
	return render.DrawPathWithEffects(r, path, paint, r.Path)
}

// SupportsNativeHatch reports that the PGF backend consumes hatch metadata
// directly in Path.
func (r *Renderer) SupportsNativeHatch() bool { return true }

func (r *Renderer) writeHatchFill(path geom.Path, paint *render.Paint) {
	r.writeHatchFillTo(&r.content, path, paint)
}

func (r *Renderer) writeHatchFillTo(w *strings.Builder, path geom.Path, paint *render.Paint) {
	if paint == nil || paint.Hatch == "" || paint.HatchColor.A <= 0 {
		return
	}
	w.WriteString("\\pgfscope\n")
	if !writePathOps(w, path) {
		w.WriteString("\\endpgfscope\n")
		return
	}
	w.WriteString("\\pgfusepath{clip}\n")
	writeFillOpacity(w, 1)
	writeStrokeOpacity(w, paint.HatchColor.A)
	writeStrokeColor(w, r.colorName(paint.HatchColor))
	lineWidth := paint.HatchLineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	fmt.Fprintf(w, "\\pgfsetlinewidth{%spt}\n", shortFloat(lineWidth))
	for _, line := range hatchPatternLines(paint.Hatch, paint.HatchSpacing) {
		fmt.Fprintf(w, "\\pgfpathmoveto{\\pgfpoint{%spt}{%spt}}\n", shortFloat(line[0].X), shortFloat(line[0].Y))
		fmt.Fprintf(w, "\\pgfpathlineto{\\pgfpoint{%spt}{%spt}}\n", shortFloat(line[1].X), shortFloat(line[1].Y))
		w.WriteString("\\pgfusepath{stroke}\n")
	}
	w.WriteString("\\endpgfscope\n")
}

// DrawMarkers renders one marker path at many display-space offsets using a
// reusable PGF macro for identical marker geometry and paint.
func (r *Renderer) DrawMarkers(batch render.MarkerBatch) bool {
	if rr := r.activeRaster(); rr != nil {
		if markers, ok := rr.(render.MarkerDrawer); ok {
			return markers.DrawMarkers(batch)
		}
		return false
	}
	if !r.began || len(batch.Marker.C) == 0 || len(batch.Items) == 0 || !batch.Marker.Validate() {
		return false
	}
	emitted := false
	for i := range batch.Items {
		item := batch.Items[i]
		if !paintVisible(&item.Paint) {
			continue
		}
		name := r.registerPathMacro("M", batch.Marker, &item.Paint)
		if name == "" {
			continue
		}
		t := normalizedAffine(item.Transform)
		t.E += item.Offset.X
		t.F += item.Offset.Y
		r.content.WriteString("\\pgfscope\n")
		writeTransform(&r.content, t)
		fmt.Fprintf(&r.content, "\\csname %s\\endcsname\n", name)
		r.content.WriteString("\\endpgfscope\n")
		emitted = true
	}
	return emitted
}

// DrawPathCollection renders display-space paths through reusable PGF macros
// keyed by path geometry and paint.
func (r *Renderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	if rr := r.activeRaster(); rr != nil {
		if paths, ok := rr.(render.PathCollectionDrawer); ok {
			return paths.DrawPathCollection(batch)
		}
		return false
	}
	if !r.began || len(batch.Items) == 0 {
		return false
	}
	emitted := false
	for i := range batch.Items {
		item := batch.Items[i]
		if !item.Path.Validate() || len(item.Path.C) == 0 {
			continue
		}
		paint := item.Paint
		if item.Hatch != "" {
			paint.Hatch = item.Hatch
			paint.HatchColor = item.HatchColor
			paint.HatchLineWidth = item.HatchWidth
			paint.HatchSpacing = item.HatchSpacing
		}
		if !paintVisible(&paint) {
			continue
		}
		name := r.registerPathMacro("P", item.Path, &paint)
		if name == "" {
			continue
		}
		fmt.Fprintf(&r.content, "\\csname %s\\endcsname\n", name)
		emitted = true
	}
	return emitted
}

func (r *Renderer) registerPathMacro(prefix string, path geom.Path, paint *render.Paint) string {
	key := pathMacroKey(prefix, path, paint)
	if name, ok := r.pathIDs[key]; ok {
		return name
	}
	var body strings.Builder
	if !r.writePathPaintOps(&body, path, paint) {
		return ""
	}
	name := fmt.Sprintf("mplgpgf%s%d", prefix, len(r.pathIDs)+1)
	r.pathIDs[key] = name
	fmt.Fprintf(&r.content, "\\expandafter\\def\\csname %s\\endcsname{%%\n%s}\n", name, body.String())
	return name
}

// Image draws an RGBA raster image as deterministic pure-PGF pixel rectangles.
// This keeps .pgf output self-contained at the cost of large output for dense
// images; callers that need compact publication files should prefer PDF/SVG.
func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.Image(img, dst)
		return
	}
	if !r.began || img == nil || dst.W() <= 0 || dst.H() <= 0 {
		return
	}
	width, height := img.Size()
	if width <= 0 || height <= 0 {
		return
	}
	r.writeImagePixels(img, geom.Affine{
		A: dst.W() / float64(width),
		D: dst.H() / float64(height),
		E: dst.Min.X,
		F: dst.Min.Y,
	})
}

// ImageTransformed draws a raster image through an arbitrary affine transform.
// The affine maps source image pixels into display coordinates.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, transform geom.Affine) {
	if rr := r.activeRaster(); rr != nil {
		if tr, ok := rr.(render.ImageTransformer); ok {
			tr.ImageTransformed(img, geom.Rect{}, transform)
		}
		return
	}
	if !r.began || img == nil {
		return
	}
	r.writeImagePixels(img, transform)
}

func (r *Renderer) writeImagePixels(img render.Image, transform geom.Affine) {
	rgbaSource, ok := img.(render.RGBAImage)
	if !ok || rgbaSource.RGBA() == nil {
		return
	}
	src := rgbaSource.RGBA()
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return
	}
	alphaMul := imageAlphaMultiplier(img)
	r.content.WriteString("\\pgfscope\n")
	writeTransform(&r.content, normalizedAffine(transform))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := src.RGBAAt(bounds.Min.X+x, bounds.Min.Y+y)
			alpha := clamp01(float64(c.A) / 255.0 * alphaMul)
			if alpha <= 0 {
				continue
			}
			writeFillOpacity(&r.content, alpha)
			writeFillColor(&r.content, r.colorName(render.Color{
				R: float64(c.R) / 255.0,
				G: float64(c.G) / 255.0,
				B: float64(c.B) / 255.0,
				A: 1,
			}))
			fmt.Fprintf(&r.content, "\\pgfpathrectangle{\\pgfpoint{%spt}{%spt}}{\\pgfpoint{1pt}{1pt}}\n",
				shortFloat(float64(x)), shortFloat(float64(y)))
			r.content.WriteString("\\pgfusepath{fill}\n")
		}
	}
	r.content.WriteString("\\endpgfscope\n")
}

// GlyphRun draws glyph IDs as fallback text when possible.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if !r.began || len(run.Glyphs) == 0 || run.Size <= 0 || textColor.A <= 0 {
		return
	}
	penX := run.Origin.X
	penY := run.Origin.Y
	for _, glyph := range run.Glyphs {
		if glyph.ID != 0 {
			r.DrawTextWithFont(string(rune(glyph.ID)), geom.Pt{X: penX + glyph.Offset.X, Y: penY + glyph.Offset.Y}, run.Size, textColor, run.FontKey)
		}
		advance := glyph.Advance
		if advance == 0 && glyph.ID != 0 {
			advance = r.MeasureText(string(rune(glyph.ID)), run.Size, run.FontKey).W
		}
		penX += advance
	}
}

// MeasureText returns deterministic approximate text metrics for layout.
func (r *Renderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	if size <= 0 {
		size = defaultFontHeight
	}
	width := 0.6 * size * float64(len([]rune(text)))
	ascent := 0.8 * size
	descent := 0.2 * size
	return render.TextMetrics{W: width, H: ascent + descent, Ascent: ascent, Descent: descent}
}

// DrawText draws text at a baseline origin.
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	r.DrawTextWithFont(text, origin, size, textColor, "")
}

// DrawTextWithFont draws text with an explicit font key. Font selection is
// intentionally left to LaTeX in this generator-only slice.
func (r *Renderer) DrawTextWithFont(text string, origin geom.Pt, size float64, textColor render.Color, _ string) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, origin, size, textColor)
		}
		return
	}
	r.writeText(text, origin, size, 0, textColor)
}

// DrawTextRotated draws text rotated around its anchor.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	r.DrawTextRotatedWithFont(text, anchor, size, angle, textColor, "")
}

// DrawTextRotatedWithFont draws rotated text with an explicit font key.
func (r *Renderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, size, angle float64, textColor render.Color, _ string) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.RotatedTextDrawer); ok {
			textRen.DrawTextRotated(text, anchor, size, angle, textColor)
		} else if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, anchor, size, textColor)
		}
		return
	}
	r.writeText(text, anchor, size, angle, textColor)
}

func (r *Renderer) writeText(text string, origin geom.Pt, size, angle float64, textColor render.Color) {
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 {
		return
	}
	colorName := r.colorName(textColor)
	writeFillOpacity(&r.content, textColor.A)
	writeStrokeOpacity(&r.content, textColor.A)
	fmt.Fprintf(&r.content, "\\pgfsetstrokecolor{%s}\n\\pgfsetfillcolor{%s}\n", colorName, colorName)
	rotate := ""
	if angle != 0 && !math.IsNaN(angle) && !math.IsInf(angle, 0) {
		rotate = ",rotate=" + shortFloat(angle)
	}
	fmt.Fprintf(&r.content, "\\pgftext[left,base%s,at=\\pgfpoint{%spt}{%spt}]{\\fontsize{%spt}{%spt}\\selectfont %s}\n",
		rotate, shortFloat(origin.X), shortFloat(origin.Y), shortFloat(size), shortFloat(size*1.2), escapeTeXText(text))
}

// SavePGF writes the finalized PGF document to path.
func (r *Renderer) SavePGF(path string) error {
	if len(r.document) == 0 {
		return errors.New("pgf: SavePGF called before End")
	}
	return os.WriteFile(path, r.document, 0o644)
}

// SavePGFWithOptions writes the finalized PGF document to path using opts for
// export-time options.
func (r *Renderer) SavePGFWithOptions(path string, opts render.PGFOptions) error {
	if len(r.document) == 0 {
		return errors.New("pgf: SavePGFWithOptions called before End")
	}
	r.SetPGFOptions(opts)
	return os.WriteFile(path, []byte(r.documentWithOptions(r.pgfOpts)), 0o644)
}

func (r *Renderer) documentWithOptions(opts render.PGFOptions) string {
	body := stripLeadingPGFComments(string(r.document))
	var b strings.Builder
	saved := r.pgfOpts
	r.pgfOpts = opts
	r.writeDocumentComments(&b)
	r.pgfOpts = saved
	b.WriteString(body)
	return b.String()
}

func stripLeadingPGFComments(doc string) string {
	for strings.HasPrefix(doc, "%") {
		lineEnd := strings.IndexByte(doc, '\n')
		if lineEnd < 0 {
			return ""
		}
		doc = doc[lineEnd+1:]
	}
	return doc
}

func (r *Renderer) writeDocumentComments(w *strings.Builder) {
	opts := normalizePGFOptions(r.pgfOpts)
	if opts.CommentPolicy == render.PGFCommentPolicyStrip {
		return
	}
	w.WriteString("% Generated by matplotlib-go\n")
	if len(opts.Metadata) > 0 {
		keys := make([]string, 0, len(opts.Metadata))
		for key := range opts.Metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(w, "%% metadata %s: %s\n", escapePGFComment(key), escapePGFComment(opts.Metadata[key]))
		}
	}
	for _, line := range opts.Preamble {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fmt.Fprintf(w, "%% preamble: %s\n", escapePGFComment(line))
	}
	if opts.VerificationMode != render.PGFVerificationModeNone {
		fmt.Fprintf(w, "%% verification: %s\n", opts.VerificationMode)
	}
}

func normalizePGFOptions(opts render.PGFOptions) render.PGFOptions {
	if opts.CommentPolicy == "" {
		opts.CommentPolicy = render.PGFCommentPolicyKeep
	}
	if opts.VerificationMode == "" {
		opts.VerificationMode = render.PGFVerificationModeNone
	}
	if len(opts.Metadata) > 0 {
		metadata := make(map[string]string, len(opts.Metadata))
		for k, v := range opts.Metadata {
			metadata[k] = v
		}
		opts.Metadata = metadata
	}
	if len(opts.Preamble) > 0 {
		preamble := make([]string, len(opts.Preamble))
		copy(preamble, opts.Preamble)
		opts.Preamble = preamble
	}
	return opts
}

func escapePGFComment(s string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(s)
}

func writePathOps(w *strings.Builder, path geom.Path) bool {
	if !path.Validate() || len(path.C) == 0 {
		return false
	}
	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			pt := path.V[vi]
			vi++
			fmt.Fprintf(w, "\\pgfpathmoveto{\\pgfpoint{%spt}{%spt}}\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.LineTo:
			pt := path.V[vi]
			vi++
			fmt.Fprintf(w, "\\pgfpathlineto{\\pgfpoint{%spt}{%spt}}\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.QuadTo:
			if vi == 0 {
				vi += 2
				continue
			}
			prev := lastEndpoint(path, vi)
			ctrl := path.V[vi]
			end := path.V[vi+1]
			vi += 2
			c1 := geom.Pt{X: prev.X + (2.0/3.0)*(ctrl.X-prev.X), Y: prev.Y + (2.0/3.0)*(ctrl.Y-prev.Y)}
			c2 := geom.Pt{X: end.X + (2.0/3.0)*(ctrl.X-end.X), Y: end.Y + (2.0/3.0)*(ctrl.Y-end.Y)}
			writeCurve(w, c1, c2, end)
		case geom.CubicTo:
			c1 := path.V[vi]
			c2 := path.V[vi+1]
			end := path.V[vi+2]
			vi += 3
			writeCurve(w, c1, c2, end)
		case geom.ClosePath:
			w.WriteString("\\pgfpathclose\n")
		}
	}
	return true
}

func (r *Renderer) writePathPaintOps(w *strings.Builder, path geom.Path, paint *render.Paint) bool {
	if paint == nil {
		return false
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasFill := paint.Fill.A > 0 || hasHatch
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	switch {
	case hasHatch:
		if paint.Fill.A > 0 {
			if !writePathOps(w, path) {
				return false
			}
			writeFillOpacity(w, paint.Fill.A)
			writeStrokeOpacity(w, 1)
			writeFillColor(w, r.colorName(paint.Fill))
			w.WriteString("\\pgfusepath{fill}\n")
		}
		r.writeHatchFillTo(w, path, paint)
		if hasStroke {
			if !writePathOps(w, path) {
				return false
			}
			writeFillOpacity(w, 1)
			writeStrokeOpacity(w, paint.Stroke.A)
			writeStrokeColor(w, r.colorName(paint.Stroke))
			writeLineState(w, paint)
			w.WriteString("\\pgfusepath{stroke}\n")
		}
	case hasFill && hasStroke:
		if !writePathOps(w, path) {
			return false
		}
		writeFillOpacity(w, paint.Fill.A)
		writeStrokeOpacity(w, paint.Stroke.A)
		writeFillColor(w, r.colorName(paint.Fill))
		writeStrokeColor(w, r.colorName(paint.Stroke))
		writeLineState(w, paint)
		w.WriteString("\\pgfusepath{fill,stroke}\n")
	case hasFill:
		if !writePathOps(w, path) {
			return false
		}
		writeFillOpacity(w, paint.Fill.A)
		writeStrokeOpacity(w, 1)
		writeFillColor(w, r.colorName(paint.Fill))
		w.WriteString("\\pgfusepath{fill}\n")
	case hasStroke:
		if !writePathOps(w, path) {
			return false
		}
		writeFillOpacity(w, 1)
		writeStrokeOpacity(w, paint.Stroke.A)
		writeStrokeColor(w, r.colorName(paint.Stroke))
		writeLineState(w, paint)
		w.WriteString("\\pgfusepath{stroke}\n")
	default:
		return false
	}
	return true
}

func writeLineState(w *strings.Builder, paint *render.Paint) {
	lineWidth := paint.LineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	fmt.Fprintf(w, "\\pgfsetlinewidth{%spt}\n", shortFloat(lineWidth))
	if len(paint.Dashes) > 0 {
		w.WriteString("\\pgfsetdash{")
		for i, d := range paint.Dashes {
			if i > 0 {
				w.WriteString(",")
			}
			fmt.Fprintf(w, "{%spt}", shortFloat(d))
		}
		w.WriteString("}{0pt}\n")
	} else {
		w.WriteString("\\pgfsetdash{}{0pt}\n")
	}
}

func writeCurve(w *strings.Builder, c1, c2, end geom.Pt) {
	fmt.Fprintf(w, "\\pgfpathcurveto{\\pgfpoint{%spt}{%spt}}{\\pgfpoint{%spt}{%spt}}{\\pgfpoint{%spt}{%spt}}\n",
		shortFloat(c1.X), shortFloat(c1.Y),
		shortFloat(c2.X), shortFloat(c2.Y),
		shortFloat(end.X), shortFloat(end.Y))
}

func lastEndpoint(path geom.Path, vi int) geom.Pt {
	consumed := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo, geom.LineTo:
			consumed++
			if consumed == vi {
				return path.V[consumed-1]
			}
		case geom.QuadTo:
			consumed += 2
			if consumed == vi {
				return path.V[consumed-1]
			}
		case geom.CubicTo:
			consumed += 3
			if consumed == vi {
				return path.V[consumed-1]
			}
		case geom.ClosePath:
		}
	}
	return geom.Pt{}
}

func paintVisible(paint *render.Paint) bool {
	if paint == nil {
		return false
	}
	return paint.Fill.A > 0 ||
		(paint.Hatch != "" && paint.HatchColor.A > 0) ||
		(paint.Stroke.A > 0 && paint.LineWidth > 0)
}

func normalizedAffine(affine geom.Affine) geom.Affine {
	if affine == (geom.Affine{}) {
		return geom.Identity()
	}
	return affine
}

func cloneRectPtr(rect *geom.Rect) *geom.Rect {
	if rect == nil {
		return nil
	}
	cloned := *rect
	return &cloned
}

func normalizeRect(rect geom.Rect) geom.Rect {
	minX, maxX := rect.Min.X, rect.Max.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := rect.Min.Y, rect.Max.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}
}

func writeTransform(w *strings.Builder, transform geom.Affine) {
	fmt.Fprintf(w, "\\pgftransformcm{%s}{%s}{%s}{%s}{\\pgfpoint{%spt}{%spt}}\n",
		shortFloat(transform.A),
		shortFloat(transform.B),
		shortFloat(transform.C),
		shortFloat(transform.D),
		shortFloat(transform.E),
		shortFloat(transform.F),
	)
}

func pathMacroKey(prefix string, path geom.Path, paint *render.Paint) string {
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteByte('|')
	for _, cmd := range path.C {
		fmt.Fprintf(&b, "%d;", cmd)
	}
	b.WriteByte('|')
	for _, pt := range path.V {
		b.WriteString(shortFloat(pt.X))
		b.WriteByte(',')
		b.WriteString(shortFloat(pt.Y))
		b.WriteByte(';')
	}
	b.WriteByte('|')
	writePaintKey(&b, paint)
	return b.String()
}

func writePaintKey(b *strings.Builder, paint *render.Paint) {
	if paint == nil {
		return
	}
	writeColorKey(b, paint.Fill)
	b.WriteByte('|')
	writeColorKey(b, paint.Stroke)
	b.WriteByte('|')
	b.WriteString(shortFloat(paint.LineWidth))
	b.WriteByte('|')
	b.WriteString(paint.Hatch)
	b.WriteByte('|')
	writeColorKey(b, paint.HatchColor)
	b.WriteByte('|')
	b.WriteString(shortFloat(paint.HatchLineWidth))
	b.WriteByte('|')
	b.WriteString(shortFloat(paint.HatchSpacing))
}

func writeColorKey(b *strings.Builder, c render.Color) {
	b.WriteString(shortFloat(c.R))
	b.WriteByte(',')
	b.WriteString(shortFloat(c.G))
	b.WriteByte(',')
	b.WriteString(shortFloat(c.B))
	b.WriteByte(',')
	b.WriteString(shortFloat(c.A))
}

func imageAlphaMultiplier(img render.Image) float64 {
	if alphaImage, ok := img.(render.ImageAlpha); ok {
		return clamp01(alphaImage.Alpha())
	}
	return 1
}

func (r *Renderer) colorName(c render.Color) string {
	c = render.Color{R: clamp01(c.R), G: clamp01(c.G), B: clamp01(c.B), A: clamp01(c.A)}
	key := strings.Join([]string{shortFloat(c.R), shortFloat(c.G), shortFloat(c.B)}, ",")
	if name, ok := r.colorNames[key]; ok {
		return name
	}
	name := fmt.Sprintf("mplgpgfcolor%d", len(r.colorNames)+1)
	r.colorNames[key] = name
	fmt.Fprintf(&r.content, "\\definecolor{%s}{rgb}{%s,%s,%s}\n", name, shortFloat(c.R), shortFloat(c.G), shortFloat(c.B))
	return name
}

func writeFillColor(w *strings.Builder, colorName string) {
	fmt.Fprintf(w, "\\pgfsetfillcolor{%s}\n", colorName)
}

func writeStrokeColor(w *strings.Builder, colorName string) {
	fmt.Fprintf(w, "\\pgfsetstrokecolor{%s}\n", colorName)
}

func writeFillOpacity(w *strings.Builder, alpha float64) {
	fmt.Fprintf(w, "\\pgfsetfillopacity{%s}\n", shortFloat(clamp01(alpha)))
}

func writeStrokeOpacity(w *strings.Builder, alpha float64) {
	fmt.Fprintf(w, "\\pgfsetstrokeopacity{%s}\n", shortFloat(clamp01(alpha)))
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

func shortFloat(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0"
	}
	if v == 0 {
		return "0"
	}
	s := strconv.FormatFloat(v, 'f', 6, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	if s == "" || s == "-0" {
		return "0"
	}
	return s
}

func escapeTeXText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\textbackslash{}`)
		case '{', '}':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '#', '$', '%', '&', '_':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '^':
			b.WriteString(`\textasciicircum{}`)
		case '~':
			b.WriteString(`\textasciitilde{}`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func hatchPatternLines(hatch string, spacing float64) [][2]geom.Pt {
	if spacing <= 0 {
		spacing = 8
	}
	lines := make([][2]geom.Pt, 0)
	writeHatchLines := func(count int, draw func(float64)) {
		if count <= 0 {
			return
		}
		step := math.Max(2, spacing/float64(count))
		for v := -72.0; v <= 144; v += step {
			draw(v)
		}
	}
	add := func(x1, y1, x2, y2 float64) {
		lines = append(lines, [2]geom.Pt{{X: x1, Y: y1}, {X: x2, Y: y2}})
	}
	verticalCount := strings.Count(hatch, "|") + strings.Count(hatch, "+")
	horizontalCount := strings.Count(hatch, "-") + strings.Count(hatch, "+")
	slashCount := strings.Count(hatch, "/") + strings.Count(hatch, "x") + strings.Count(hatch, "X")
	backslashCount := strings.Count(hatch, `\`) + strings.Count(hatch, "x") + strings.Count(hatch, "X")

	writeHatchLines(verticalCount, func(x float64) { add(x, 0, x, 72) })
	writeHatchLines(horizontalCount, func(y float64) { add(0, y, 72, y) })
	writeHatchLines(slashCount, func(x float64) { add(x, 72, x+72, 0) })
	writeHatchLines(backslashCount, func(x float64) { add(x, 0, x+72, 72) })
	return lines
}

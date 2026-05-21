package ps

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
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

// Renderer implements render.Renderer by emitting deterministic Level-2
// PostScript content.
type Renderer struct {
	width      int
	height     int
	background render.Color
	resolution uint

	began       bool
	viewport    geom.Rect
	content     strings.Builder
	document    []byte
	stack       []state
	clipRect    *geom.Rect
	clipPaths   []geom.Path
	raster      *mixedraster.Session
	markerIDs   map[string]string
	imageIDs    map[string]string
	psOpts      render.PSOptions
	lastFontKey string
}

var (
	_ render.Renderer                = (*Renderer)(nil)
	_ render.PSExporter              = (*Renderer)(nil)
	_ render.DPIAware                = (*Renderer)(nil)
	_ render.ImageTransformer        = (*Renderer)(nil)
	_ render.MarkerDrawer            = (*Renderer)(nil)
	_ render.NativeHatcher           = (*Renderer)(nil)
	_ render.PathCollectionDrawer    = (*Renderer)(nil)
	_ render.RasterizationController = (*Renderer)(nil)
	_ render.TextPather              = (*Renderer)(nil)
	_ render.FontTextDrawer          = (*Renderer)(nil)
	_ render.FontRotatedTextDrawer   = (*Renderer)(nil)
	_ render.FontVerticalTextDrawer  = (*Renderer)(nil)
	_ render.PSOptionSetter          = (*Renderer)(nil)
)

// New creates a PostScript renderer with a point-sized page.
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
		psOpts:     render.DefaultPSOptions(),
	}, nil
}

// SetResolution implements render.DPIAware.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi == 0 {
		dpi = 72
	}
	r.resolution = dpi
}

// SetPSOptions implements render.PSOptionSetter.
func (r *Renderer) SetPSOptions(opts render.PSOptions) {
	r.psOpts = opts
}

// Begin starts a drawing session for the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("ps: Begin called twice")
	}
	r.began = true
	r.viewport = viewport
	r.content.Reset()
	r.document = nil
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipPaths = nil
	r.raster = nil
	r.markerIDs = map[string]string{}
	r.imageIDs = map[string]string{}
	r.lastFontKey = ""

	r.content.WriteString("gsave\n")
	fmt.Fprintf(&r.content, "0 %s translate\n", shortFloat(float64(r.height)))
	r.content.WriteString("1 -1 scale\n")
	if r.background.A > 0 {
		writeFillColor(&r.content, r.background)
		fmt.Fprintf(&r.content, "newpath 0 0 moveto %s 0 lineto %s %s lineto 0 %s lineto closepath fill\n",
			shortFloat(float64(r.width)),
			shortFloat(float64(r.width)),
			shortFloat(float64(r.height)),
			shortFloat(float64(r.height)),
		)
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

// StopRasterized embeds the active raster group as a PostScript colorimage.
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

// End finalizes the current drawing session.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("ps: End called before Begin")
	}
	r.began = false
	r.content.WriteString("grestore\n")
	r.document = buildDocument(r.width, r.height, r.content.String(), false)
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
		r.content.WriteString("gsave\n")
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
		r.content.WriteString("grestore\n")
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
	fmt.Fprintf(&r.content, "newpath %s %s moveto %s %s lineto %s %s lineto %s %s lineto closepath clip newpath\n",
		shortFloat(rect.Min.X), shortFloat(rect.Min.Y),
		shortFloat(rect.Max.X), shortFloat(rect.Min.Y),
		shortFloat(rect.Max.X), shortFloat(rect.Max.Y),
		shortFloat(rect.Min.X), shortFloat(rect.Max.Y),
	)
}

// ClipPath installs an arbitrary path clip.
func (r *Renderer) ClipPath(p geom.Path) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipPath(p)
		return
	}
	if !r.began {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}
	r.clipPaths = append(r.clipPaths, mixedraster.ClonePath(p))
	r.content.WriteString("clip newpath\n")
}

// Path draws a path using the provided paint.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if rr := r.activeRaster(); rr != nil {
		rr.Path(p, paint)
		return
	}
	if !r.began || paint == nil {
		return
	}
	if render.DrawPathWithEffects(r, p, paint, r.Path) {
		return
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasFill := paint.Fill.A > 0 || hasHatch
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasStroke {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}

	switch {
	case hasHatch:
		if paint.Fill.A > 0 {
			writeFillColor(&r.content, paint.Fill)
			r.content.WriteString("gsave fill grestore\n")
		}
		r.writeHatchFill(p, paint)
		if hasStroke {
			if !writePathOps(&r.content, p) {
				return
			}
			writeStrokeColor(&r.content, paint.Stroke)
			writeLineState(&r.content, paint)
			r.content.WriteString("stroke\n")
		}
	case hasFill && hasStroke:
		writeFillColor(&r.content, paint.Fill)
		r.content.WriteString("gsave fill grestore\n")
		writeStrokeColor(&r.content, paint.Stroke)
		writeLineState(&r.content, paint)
		r.content.WriteString("stroke\n")
	case hasFill:
		writeFillColor(&r.content, paint.Fill)
		r.content.WriteString("fill\n")
	case hasStroke:
		writeStrokeColor(&r.content, paint.Stroke)
		writeLineState(&r.content, paint)
		r.content.WriteString("stroke\n")
	}
}

// DrawPathWithEffects applies renderer-neutral path effect passes.
func (r *Renderer) DrawPathWithEffects(p geom.Path, paint *render.Paint) bool {
	if rr := r.activeRaster(); rr != nil {
		if effects, ok := rr.(render.PathEffectDrawer); ok {
			return effects.DrawPathWithEffects(p, paint)
		}
		rr.Path(p, paint)
		return true
	}
	return render.DrawPathWithEffects(r, p, paint, r.Path)
}

// SupportsNativeHatch reports that the PS backend consumes hatch metadata
// directly in Path.
func (r *Renderer) SupportsNativeHatch() bool { return true }

// DrawMarkers renders one marker path at many display-space offsets using a
// reusable PostScript procedure for identical marker geometry and paint.
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
		marker := affinePath(batch.Marker, normalizedAffine(item.Transform))
		if !marker.Validate() || len(marker.C) == 0 || !paintVisible(&item.Paint) {
			continue
		}
		name := r.registerMarkerProcedure(marker, &item.Paint)
		if name == "" {
			continue
		}
		fmt.Fprintf(&r.content, "gsave\n%s %s translate\n%s\ngrestore\n",
			shortFloat(item.Offset.X),
			shortFloat(item.Offset.Y),
			name,
		)
		emitted = true
	}
	return emitted
}

// DrawPathCollection renders display-space paths through reusable PostScript
// procedures keyed by path geometry and paint.
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
		name := r.registerPathProcedure("P", item.Path, &paint)
		if name == "" {
			continue
		}
		fmt.Fprintf(&r.content, "%s\n", name)
		emitted = true
	}
	return emitted
}

func (r *Renderer) registerMarkerProcedure(marker geom.Path, paint *render.Paint) string {
	return r.registerPathProcedure("M", marker, paint)
}

func (r *Renderer) registerPathProcedure(prefix string, path geom.Path, paint *render.Paint) string {
	key := pathProcedureKey(prefix, path, paint)
	if name, ok := r.markerIDs[key]; ok {
		return name
	}
	var body strings.Builder
	if !writePathOps(&body, path) || !writePathPaintOps(&body, path, paint) {
		return ""
	}
	name := fmt.Sprintf("%s%d", prefix, len(r.markerIDs)+1)
	r.markerIDs[key] = name
	fmt.Fprintf(&r.content, "/%s {\n%s} bind def\n", name, body.String())
	return name
}

func (r *Renderer) writeHatchFill(p geom.Path, paint *render.Paint) {
	if paint == nil || paint.Hatch == "" || paint.HatchColor.A <= 0 {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}
	r.content.WriteString("gsave clip newpath\n")
	writeStrokeColor(&r.content, paint.HatchColor)
	lineWidth := paint.HatchLineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	fmt.Fprintf(&r.content, "%s setlinewidth\n", shortFloat(lineWidth))
	r.content.WriteString("0 setlinecap\n")
	for _, line := range hatchPatternLines(paint.Hatch, paint.HatchSpacing) {
		fmt.Fprintf(&r.content, "newpath %s %s moveto %s %s lineto\nstroke\n",
			shortFloat(line[0].X), shortFloat(line[0].Y),
			shortFloat(line[1].X), shortFloat(line[1].Y),
		)
	}
	r.content.WriteString("grestore\n")
}

// Image draws an RGBA raster image into the destination rectangle using a
// Level-2 colorimage operator. PostScript has no native alpha channel, so
// translucent image pixels are pre-composited over white for this first slice.
func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.Image(img, dst)
		return
	}
	if !r.began || img == nil || dst.W() <= 0 || dst.H() <= 0 {
		return
	}
	rgb, width, height, ok := encodePSImageRGB(img)
	if !ok {
		return
	}
	r.writeImageWithMatrix(rgb, width, height, geom.Affine{
		A: dst.W(),
		D: dst.H(),
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
	width, height := img.Size()
	if width <= 0 || height <= 0 {
		return
	}
	rgb, _, _, ok := encodePSImageRGB(img)
	if !ok {
		return
	}
	r.writeImageWithMatrix(rgb, width, height, geom.Affine{
		A: transform.A * float64(width),
		B: transform.B * float64(width),
		C: transform.C * float64(height),
		D: transform.D * float64(height),
		E: transform.E,
		F: transform.F,
	})
}

func (r *Renderer) writeImageWithMatrix(rgb string, width, height int, matrix geom.Affine) {
	name := r.registerImageProcedure(rgb, width, height)
	if name == "" {
		return
	}
	fmt.Fprintf(&r.content, "gsave\n[%s %s %s %s %s %s] concat\n%s\ngrestore\n",
		shortFloat(matrix.A),
		shortFloat(matrix.B),
		shortFloat(matrix.C),
		shortFloat(matrix.D),
		shortFloat(matrix.E),
		shortFloat(matrix.F),
		name,
	)
}

func (r *Renderer) registerImageProcedure(rgb string, width, height int) string {
	if width <= 0 || height <= 0 || rgb == "" {
		return ""
	}
	key := fmt.Sprintf("%dx%d:%s", width, height, rgb)
	if name, ok := r.imageIDs[key]; ok {
		return name
	}
	name := fmt.Sprintf("I%d", len(r.imageIDs)+1)
	r.imageIDs[key] = name
	fmt.Fprintf(&r.content, "/%s {\n/DeviceRGB setcolorspace\n", name)
	fmt.Fprintf(&r.content, "%d %d 8 [%d 0 0 -%d 0 %d]\n", width, height, width, height, height)
	fmt.Fprintf(&r.content, "{<%s>} false 3 colorimage\n} bind def\n", rgb)
	return name
}

// GlyphRun draws a shaped run using glyph IDs as Unicode scalar values when
// possible. Core text paths normally call DrawText directly.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if !r.began || len(run.Glyphs) == 0 || run.Size <= 0 || textColor.A <= 0 {
		return
	}
	penX := run.Origin.X
	penY := run.Origin.Y
	for _, glyph := range run.Glyphs {
		if glyph.ID != 0 {
			r.DrawTextWithFont(string(rune(glyph.ID)), geom.Pt{
				X: penX + glyph.Offset.X,
				Y: penY + glyph.Offset.Y,
			}, run.Size, textColor, run.FontKey)
		}
		advance := glyph.Advance
		if advance == 0 && glyph.ID != 0 {
			advance = r.MeasureText(string(rune(glyph.ID)), run.Size, run.FontKey).W
		}
		penX += advance
	}
}

// MeasureText returns deterministic approximate text metrics for layout.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	if size <= 0 {
		size = defaultFontHeight
	}
	width := 0.6 * size * float64(len([]rune(text)))
	ascent := 0.8 * size
	descent := 0.2 * size
	return render.TextMetrics{W: width, H: ascent + descent, Ascent: ascent, Descent: descent}
}

// TextPath converts text to vector glyph outlines through the shared font
// manager. PS path text is drawn through a local Y reflection so glyphs remain
// upright inside the backend's top-left display-coordinate transform.
func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}
	return render.TextPath(text, origin, size, fontKey)
}

// DrawText draws text using a standard PostScript base font.
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	r.DrawTextWithFont(text, origin, size, textColor, r.lastFontKey)
}

// DrawTextWithFont draws text with an explicit font key using the configured
// PS font policy. The default mirrors PDF's deterministic glyph-path output;
// PSFontPolicyBase14 keeps simple searchable Helvetica text with no embedding.
func (r *Renderer) DrawTextWithFont(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, origin, size, textColor)
		}
		return
	}
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 {
		return
	}
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	if r.psOpts.FontPolicy != render.PSFontPolicyBase14 && r.drawTextPath(text, size, textColor, fontKey, textPathAffine(origin)) {
		return
	}
	r.writeTextAt(text, origin, size, 0, textColor)
}

// DrawTextRotated draws text around the supplied anchor point.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	r.DrawTextRotatedWithFont(text, anchor, size, angle, textColor, r.lastFontKey)
}

// DrawTextRotatedWithFont draws rotated text with an explicit font key.
func (r *Renderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.RotatedTextDrawer); ok {
			textRen.DrawTextRotated(text, anchor, size, angle, textColor)
		} else if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, anchor, size, textColor)
		}
		return
	}
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 {
		return
	}
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	if r.psOpts.FontPolicy != render.PSFontPolicyBase14 && !math.IsNaN(angle) && !math.IsInf(angle, 0) {
		metrics := r.MeasureText(text, size, fontKey)
		origin := geom.Pt{X: anchor.X - metrics.W/2, Y: anchor.Y - metrics.Descent}
		transform := rotationAffine(angle, anchor).Mul(textPathAffine(origin))
		if r.drawTextPath(text, size, textColor, fontKey, transform) {
			return
		}
	}
	r.writeTextAt(text, anchor, size, angle, textColor)
}

// DrawTextVertical draws vertical text centered on the supplied point.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	r.DrawTextVerticalWithFont(text, center, size, textColor, r.lastFontKey)
}

// DrawTextVerticalWithFont draws vertical text with an explicit font key.
func (r *Renderer) DrawTextVerticalWithFont(text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.VerticalTextDrawer); ok {
			textRen.DrawTextVertical(text, center, size, textColor)
		} else if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, center, size, textColor)
		}
		return
	}
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 {
		return
	}
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	runes := []rune(text)
	lineHeight := r.MeasureText("M", size, fontKey).H
	startY := center.Y - lineHeight*float64(len(runes)-1)/2
	for i, ch := range runes {
		s := string(ch)
		metrics := r.MeasureText(s, size, fontKey)
		origin := geom.Pt{X: center.X - metrics.W/2, Y: startY + float64(i)*lineHeight}
		if r.psOpts.FontPolicy != render.PSFontPolicyBase14 && r.drawTextPath(s, size, textColor, fontKey, textPathAffine(origin)) {
			continue
		}
		r.writeTextAt(s, origin, size, 0, textColor)
	}
}

func (r *Renderer) drawTextPath(text string, size float64, textColor render.Color, fontKey string, transform geom.Affine) bool {
	path, ok := render.TextPath(text, geom.Pt{}, size, fontKey)
	if !ok || !path.Validate() || len(path.C) == 0 {
		return false
	}
	r.Path(affinePath(path, transform), &render.Paint{Fill: textColor})
	return true
}

func (r *Renderer) writeTextAt(text string, origin geom.Pt, size, angle float64, textColor render.Color) {
	r.content.WriteString("gsave\n")
	writeFillColor(&r.content, textColor)
	fmt.Fprintf(&r.content, "%s %s translate\n", shortFloat(origin.X), shortFloat(origin.Y))
	if angle != 0 && !math.IsNaN(angle) && !math.IsInf(angle, 0) {
		fmt.Fprintf(&r.content, "%s rotate\n", shortFloat(angle*180/math.Pi))
	}
	r.content.WriteString("1 -1 scale\n")
	fmt.Fprintf(&r.content, "/Helvetica findfont %s scalefont setfont\n", shortFloat(size))
	fmt.Fprintf(&r.content, "0 0 moveto (%s) show\n", escapePSString(text))
	r.content.WriteString("grestore\n")
}

// SavePS writes the finalized document to path.
func (r *Renderer) SavePS(path string) error {
	if len(r.document) == 0 {
		return errors.New("ps: SavePS called before End")
	}
	ext := strings.ToLower(filepath.Ext(path))
	data := r.document
	if ext == ".eps" {
		data = buildDocument(r.width, r.height, r.content.String(), true)
	}
	return os.WriteFile(path, data, 0o644)
}

func buildDocument(width, height int, content string, eps bool) []byte {
	var b strings.Builder
	if eps {
		b.WriteString("%!PS-Adobe-3.0 EPSF-3.0\n")
	} else {
		b.WriteString("%!PS-Adobe-3.0\n")
	}
	b.WriteString("%%Creator: matplotlib-go\n")
	fmt.Fprintf(&b, "%%%%BoundingBox: 0 0 %d %d\n", width, height)
	b.WriteString("%%LanguageLevel: 2\n")
	if !eps {
		b.WriteString("%%Pages: 1\n")
	}
	b.WriteString("%%EndComments\n")
	if !eps {
		b.WriteString("%%Page: 1 1\n")
	}
	b.WriteString(content)
	if !eps {
		b.WriteString("showpage\n")
	}
	b.WriteString("%%EOF\n")
	return []byte(b.String())
}

func writePathOps(w *strings.Builder, p geom.Path) bool {
	if !p.Validate() || len(p.C) == 0 {
		return false
	}
	w.WriteString("newpath\n")
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			vi++
			fmt.Fprintf(w, "%s %s moveto\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.LineTo:
			pt := p.V[vi]
			vi++
			fmt.Fprintf(w, "%s %s lineto\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.QuadTo:
			if vi == 0 {
				vi += 2
				continue
			}
			prev := lastEndpoint(p, vi)
			ctrl := p.V[vi]
			end := p.V[vi+1]
			vi += 2
			c1 := geom.Pt{
				X: prev.X + (2.0/3.0)*(ctrl.X-prev.X),
				Y: prev.Y + (2.0/3.0)*(ctrl.Y-prev.Y),
			}
			c2 := geom.Pt{
				X: end.X + (2.0/3.0)*(ctrl.X-end.X),
				Y: end.Y + (2.0/3.0)*(ctrl.Y-end.Y),
			}
			fmt.Fprintf(w, "%s %s %s %s %s %s curveto\n",
				shortFloat(c1.X), shortFloat(c1.Y),
				shortFloat(c2.X), shortFloat(c2.Y),
				shortFloat(end.X), shortFloat(end.Y),
			)
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			end := p.V[vi+2]
			vi += 3
			fmt.Fprintf(w, "%s %s %s %s %s %s curveto\n",
				shortFloat(c1.X), shortFloat(c1.Y),
				shortFloat(c2.X), shortFloat(c2.Y),
				shortFloat(end.X), shortFloat(end.Y),
			)
		case geom.ClosePath:
			w.WriteString("closepath\n")
		}
	}
	return true
}

func lastEndpoint(p geom.Path, vi int) geom.Pt {
	consumed := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo, geom.LineTo:
			consumed++
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.QuadTo:
			consumed += 2
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.CubicTo:
			consumed += 3
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.ClosePath:
		}
	}
	return geom.Pt{}
}

func writePathPaintOps(w *strings.Builder, path geom.Path, paint *render.Paint) bool {
	if paint == nil {
		return false
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasFill := paint.Fill.A > 0 || hasHatch
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	switch {
	case hasHatch:
		if paint.Fill.A > 0 {
			writeFillColor(w, paint.Fill)
			w.WriteString("gsave fill grestore\n")
		}
		writeStrokeColor(w, paint.HatchColor)
		lineWidth := paint.HatchLineWidth
		if lineWidth <= 0 {
			lineWidth = 1
		}
		fmt.Fprintf(w, "%s setlinewidth\n", shortFloat(lineWidth))
		w.WriteString("gsave clip newpath\n")
		for _, line := range hatchPatternLines(paint.Hatch, paint.HatchSpacing) {
			fmt.Fprintf(w, "newpath %s %s moveto %s %s lineto\nstroke\n",
				shortFloat(line[0].X), shortFloat(line[0].Y),
				shortFloat(line[1].X), shortFloat(line[1].Y),
			)
		}
		w.WriteString("grestore\n")
		if hasStroke {
			if !writePathOps(w, path) {
				return false
			}
			writeStrokeColor(w, paint.Stroke)
			writeLineState(w, paint)
			w.WriteString("stroke\n")
		}
	case hasFill && hasStroke:
		writeFillColor(w, paint.Fill)
		w.WriteString("gsave fill grestore\n")
		writeStrokeColor(w, paint.Stroke)
		writeLineState(w, paint)
		w.WriteString("stroke\n")
	case hasFill:
		writeFillColor(w, paint.Fill)
		w.WriteString("fill\n")
	case hasStroke:
		writeStrokeColor(w, paint.Stroke)
		writeLineState(w, paint)
		w.WriteString("stroke\n")
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
	fmt.Fprintf(w, "%s setlinewidth\n", shortFloat(lineWidth))
	fmt.Fprintf(w, "%d setlinejoin\n", lineJoin(paint.LineJoin))
	fmt.Fprintf(w, "%d setlinecap\n", lineCap(paint.LineCap))
	if paint.MiterLimit > 0 {
		fmt.Fprintf(w, "%s setmiterlimit\n", shortFloat(paint.MiterLimit))
	}
	if len(paint.Dashes) > 0 {
		w.WriteByte('[')
		for i, d := range paint.Dashes {
			if i > 0 {
				w.WriteByte(' ')
			}
			w.WriteString(shortFloat(d))
		}
		w.WriteString("] 0 setdash\n")
	} else {
		w.WriteString("[] 0 setdash\n")
	}
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

func textPathAffine(origin geom.Pt) geom.Affine {
	return translateAffine(origin).Mul(geom.Affine{A: 1, D: -1})
}

func rotationAffine(angle float64, pivot geom.Pt) geom.Affine {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return translateAffine(pivot).
		Mul(geom.Affine{A: cos, B: sin, C: -sin, D: cos}).
		Mul(translateAffine(geom.Pt{X: -pivot.X, Y: -pivot.Y}))
}

func translateAffine(p geom.Pt) geom.Affine {
	return geom.Affine{A: 1, D: 1, E: p.X, F: p.Y}
}

func affinePath(path geom.Path, affine geom.Affine) geom.Path {
	if len(path.V) == 0 {
		return path
	}
	out := geom.Path{
		V: make([]geom.Pt, len(path.V)),
		C: append([]geom.Cmd(nil), path.C...),
	}
	for i, pt := range path.V {
		out.V[i] = affine.Apply(pt)
	}
	return out
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

func pathProcedureKey(prefix string, path geom.Path, paint *render.Paint) string {
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

func writeStrokeColor(w *strings.Builder, c render.Color) {
	writeColor(w, c)
}

func writeFillColor(w *strings.Builder, c render.Color) {
	writeColor(w, c)
}

func writeColor(w *strings.Builder, c render.Color) {
	fmt.Fprintf(w, "%s %s %s setrgbcolor\n", shortFloat(clamp01(c.R)), shortFloat(clamp01(c.G)), shortFloat(clamp01(c.B)))
}

func lineJoin(join render.LineJoin) int {
	switch join {
	case render.JoinRound:
		return 1
	case render.JoinBevel:
		return 2
	default:
		return 0
	}
}

func lineCap(lineCap render.LineCap) int {
	switch lineCap {
	case render.CapRound:
		return 1
	case render.CapSquare:
		return 2
	default:
		return 0
	}
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

func escapePSString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '(', ')':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 || r > 0x7e {
				b.WriteByte('?')
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

func imageAlphaMultiplier(img render.Image) float64 {
	if alphaImage, ok := img.(render.ImageAlpha); ok {
		return clamp01(alphaImage.Alpha())
	}
	return 1
}

func encodePSImageRGB(img render.Image) (string, int, int, bool) {
	rgbaSource, ok := img.(render.RGBAImage)
	if !ok || rgbaSource.RGBA() == nil {
		return "", 0, 0, false
	}
	src := rgbaSource.RGBA()
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return "", 0, 0, false
	}
	alphaMul := imageAlphaMultiplier(img)
	var b strings.Builder
	b.Grow(width * height * 6)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)
			r, g, blue := compositeImagePixelOverWhite(c.R, c.G, c.B, c.A, alphaMul)
			fmt.Fprintf(&b, "%02x%02x%02x", r, g, blue)
		}
	}
	return b.String(), width, height, true
}

func compositeImagePixelOverWhite(red, green, blue, alpha uint8, alphaMul float64) (uint8, uint8, uint8) {
	a := clamp01(float64(alpha) / 255.0 * alphaMul)
	return uint8(float64(red)*a + 255*(1-a) + 0.5),
		uint8(float64(green)*a + 255*(1-a) + 0.5),
		uint8(float64(blue)*a + 255*(1-a) + 0.5)
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

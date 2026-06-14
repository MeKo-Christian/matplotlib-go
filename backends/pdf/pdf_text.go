package pdf

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"github.com/cwbudde/matplotlib-go/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/font/sfnt"
)

// GlyphRun draws shaped glyphs as filled outlines. GlyphRun only carries glyph
// IDs, so this remains a practical fallback for simple code-point-shaped runs.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if len(run.Glyphs) == 0 {
		return
	}
	if run.FontKey != "" {
		r.lastFontKey = run.FontKey
	}
	size := run.Size
	if size <= 0 {
		size = 12
	}
	penX := run.Origin.X
	penY := run.Origin.Y
	for _, glyph := range run.Glyphs {
		if glyph.ID != 0 {
			text := string(rune(glyph.ID))
			origin := geom.Pt{X: penX + glyph.Offset.X, Y: penY + glyph.Offset.Y}
			r.DrawTextWithFont(text, origin, size, textColor, r.lastFontKey)
		}
		advance := glyph.Advance
		if advance <= 0 && glyph.ID != 0 {
			advance = r.MeasureText(string(rune(glyph.ID)), size, r.lastFontKey).W
		}
		penX += advance
	}
}

// MeasureText returns rough metrics so layout code does not divide by zero.
// A future revision will plumb in the shared font manager.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" {
		return render.TextMetrics{}
	}
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	if size <= 0 {
		size = defaultFontHeight
	}
	// Crude width estimate; consistent across backends that lack a real font
	// shaper. Refined once the shared font pipeline is wired up.
	width := size * 0.5 * float64(len(text))
	return render.TextMetrics{
		W:       width,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

// TextPath converts text to vector glyph outlines through the shared font
// manager. This backs Matplotlib-style text-as-path PDF output.
func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}
	return render.TextPath(text, origin, size, fontKey)
}

// DrawText renders text using the active PDF font policy and the most recently
// resolved font key when one has been primed by MeasureText or DrawTextWithFont.
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	r.DrawTextWithFont(text, origin, size, textColor, r.lastFontKey)
}

// DrawTextWithFont renders text using the active PDF font policy and an
// explicit font key.
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
	if r.pdfOpts.FontPolicy == render.PDFFontPolicyEmbed && r.drawEmbeddedText(text, origin, size, textColor, fontKey, geom.Affine{A: 1, D: 1}) {
		return
	}
	path, ok := r.TextPath(text, origin, size, fontKey)
	if !ok {
		return
	}
	// render.TextPath already emits y-up display outlines, matching the y-up PDF
	// page; no baseline reflection is needed.
	r.Path(path, &render.Paint{Fill: textColor})
}

// DrawTextRotated renders text as outlines around Matplotlib's bottom-center
// rotated-text anchor.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	r.DrawTextRotatedWithFont(text, anchor, size, angle, textColor, r.lastFontKey)
}

// DrawTextRotatedWithFont renders rotated text as filled glyph paths.
func (r *Renderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.RotatedTextDrawer); ok {
			textRen.DrawTextRotated(text, anchor, size, angle, textColor)
		} else if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, anchor, size, textColor)
		}
		return
	}
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return
	}
	metrics := r.MeasureText(text, size, fontKey)
	if metrics.W <= 0 || metrics.H <= 0 {
		return
	}
	origin := geom.Pt{
		X: anchor.X - metrics.W/2,
		Y: anchor.Y - metrics.Descent,
	}
	if r.pdfOpts.FontPolicy == render.PDFFontPolicyEmbed && r.drawEmbeddedText(text, origin, size, textColor, fontKey, rotationAffine(angle, anchor)) {
		return
	}
	path, ok := r.TextPath(text, origin, size, fontKey)
	if !ok {
		return
	}
	// render.TextPath already emits y-up display outlines; rotate about the
	// anchor directly.
	path = affinePath(path, rotationAffine(angle, anchor))
	r.Path(path, &render.Paint{Fill: textColor})
}

// DrawTextVertical renders one character per line as filled glyph paths.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	r.DrawTextVerticalWithFont(text, center, size, textColor, r.lastFontKey)
}

// DrawTextVerticalWithFont renders vertical text with an explicit font key.
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
	runes := []rune(text)
	lineMetrics := r.MeasureText("M", size, fontKey)
	lineH := lineMetrics.H
	if lineH <= 0 {
		lineH = size
	}
	totalH := lineH * float64(len(runes))
	y := center.Y - totalH/2 + lineMetrics.Ascent
	for _, ch := range runes {
		s := string(ch)
		chMetrics := r.MeasureText(s, size, fontKey)
		if chMetrics.W <= 0 || chMetrics.H <= 0 {
			y += lineH
			continue
		}
		origin := geom.Pt{X: center.X - chMetrics.W/2, Y: y}
		r.DrawTextWithFont(s, origin, size, textColor, fontKey)
		y += lineH
	}
}

// LastTeXError returns the most recent TeX pipeline error recorded by MeasureTeX
// or DrawTeX. A nil value means the last TeX operation succeeded.
func (r *Renderer) LastTeXError() error {
	if r == nil {
		return nil
	}
	return r.texErr
}

// MeasureTeX measures a TeX string by rendering it through the external
// latex+dvipng cache and converting tight PNG pixel dimensions to PDF points.
func (r *Renderer) MeasureTeX(text string, size float64, fontKey string) (render.TextMetrics, bool) {
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok {
		return render.TextMetrics{}, false
	}
	return r.texMetricsToPoints(result.Metrics), true
}

// DrawTeX embeds a TeX-rendered PNG as a PDF image XObject.
func (r *Renderer) DrawTeX(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) bool {
	if rr := r.activeRaster(); rr != nil {
		if texRen, ok := rr.(render.TeXDrawer); ok {
			return texRen.DrawTeX(text, origin, size, textColor, fontKey)
		}
		if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, origin, size, textColor)
			return true
		}
		return false
	}
	if !r.began {
		return false
	}
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return false
	}
	metrics := r.texMetricsToPoints(result.Metrics)
	topLeft := geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent}
	r.Image(render.NewImageData(img), geom.Rect{
		Min: topLeft,
		Max: geom.Pt{X: topLeft.X + metrics.W, Y: topLeft.Y + metrics.H},
	})
	return true
}

// DrawTeXRotated embeds a TeX-rendered PNG and rotates it around the
// Matplotlib-style text rotation anchor.
func (r *Renderer) DrawTeXRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) bool {
	if rr := r.activeRaster(); rr != nil {
		if texRen, ok := rr.(render.RotatedTeXDrawer); ok {
			return texRen.DrawTeXRotated(text, anchor, size, angle, textColor, fontKey)
		}
		if textRen, ok := rr.(render.RotatedTextDrawer); ok {
			textRen.DrawTextRotated(text, anchor, size, angle, textColor)
			return true
		}
		return false
	}
	if !r.began || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return false
	}
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return false
	}
	metrics := r.texMetricsToPoints(result.Metrics)
	origin := geom.Pt{X: anchor.X - metrics.W/2, Y: anchor.Y - metrics.Descent}
	topLeft := geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent}
	scale := r.texPointScale()
	transform := rotationAffine(angle, anchor).Mul(geom.Affine{A: scale, D: scale, E: topLeft.X, F: topLeft.Y})
	r.ImageTransformed(render.NewImageData(img), geom.Rect{}, transform)
	return true
}

func (r *Renderer) renderTeX(text string, size float64, fontKey string) (tex.RenderResult, bool) {
	if r == nil || text == "" || size <= 0 {
		return tex.RenderResult{}, false
	}
	if r.texManager == nil {
		r.texManager = tex.NewManager(tex.ManagerConfig{})
	}
	result, err := r.texManager.Render(text, size, r.resolution, fontKey)
	if err != nil {
		r.texErr = err
		return tex.RenderResult{}, false
	}
	r.texErr = nil
	return result, true
}

func (r *Renderer) texMetricsToPoints(metrics render.TextMetrics) render.TextMetrics {
	scale := r.texPointScale()
	return render.TextMetrics{
		W:       metrics.W * scale,
		H:       metrics.H * scale,
		Ascent:  metrics.Ascent * scale,
		Descent: metrics.Descent * scale,
	}
}

func (r *Renderer) texPointScale() float64 {
	dpi := r.resolution
	if dpi == 0 {
		dpi = 72
	}
	return 72 / float64(dpi)
}

func colorizeTeXImage(src *image.RGBA, c render.Color) *image.RGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	red := uint8(clamp01(c.R)*255 + 0.5)
	green := uint8(clamp01(c.G)*255 + 0.5)
	blue := uint8(clamp01(c.B)*255 + 0.5)
	alphaScale := clamp01(c.A)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, a16 := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			alpha := uint8(float64(a16>>8)*alphaScale + 0.5)
			dst.SetRGBA(x, y, color.RGBA{R: red, G: green, B: blue, A: alpha})
		}
	}
	return dst
}

func (r *Renderer) drawEmbeddedText(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string, transform geom.Affine) bool {
	shaped, ok := render.ShapeText(text, origin, size, render.TextShapingOptions{FontKey: fontKey})
	if !ok || len(shaped.Runs) == 0 {
		return false
	}
	writeFillColor(&r.content, textColor)
	r.writeAlphaState(&render.Paint{Fill: textColor})
	for _, run := range shaped.Runs {
		if len(run.Glyphs) == 0 {
			continue
		}
		fontName, cids, ok := r.registerEmbeddedTextRun(run)
		if !ok || len(cids) == 0 {
			return false
		}
		start := transform.Apply(run.Glyphs[0].Origin)
		a, b, c, d := transform.A, transform.B, transform.C, transform.D
		if a == 0 && b == 0 && c == 0 && d == 0 {
			a, d = 1, 1
		}
		fmt.Fprintf(
			&r.content, "BT\n/%s %s Tf\n%s %s %s %s %s %s Tm\n%s Tj\nET\n",
			escapeName(fontName),
			shortFloat(size),
			shortFloat(a),
			shortFloat(b),
			shortFloat(c),
			shortFloat(d),
			shortFloat(start.X),
			shortFloat(start.Y),
			pdfCIDHexString(cids),
		)
	}
	return true
}

func (r *Renderer) registerEmbeddedTextRun(run render.ShapedRun) (string, []uint16, bool) {
	font, ok := r.registerEmbeddedFont(run.Face)
	if !ok {
		return "", nil, false
	}
	cids := make([]uint16, 0, len(run.Glyphs))
	for _, glyph := range run.Glyphs {
		cid := font.cidByGID[glyph.GlyphIndex]
		if cid == 0 {
			cid = uint16(len(font.cidByGID) + 1)
			font.cidByGID[glyph.GlyphIndex] = cid
			font.gidByCID[cid] = glyph.GlyphIndex
		}
		if glyph.Rune != 0 {
			font.runeByCID[cid] = glyph.Rune
		}
		cids = append(cids, cid)
	}
	return font.name, cids, true
}

func (r *Renderer) registerEmbeddedFont(face render.FontFace) (*pdfEmbeddedFont, bool) {
	if r.fontIDs == nil {
		r.fontIDs = map[string]string{}
	}
	key := pdfFontFaceKey(face)
	if key == "" {
		return nil, false
	}
	if name, ok := r.fontIDs[key]; ok {
		for i := range r.fonts {
			if r.fonts[i].name == name {
				return &r.fonts[i], true
			}
		}
	}
	data, ok := pdfFontFaceData(face)
	if !ok {
		return nil, false
	}
	parsed, err := sfnt.Parse(data)
	if err != nil {
		return nil, false
	}
	baseName := pdfFontPostScriptName(parsed, face)
	name := fmt.Sprintf("F%d", len(r.fonts)+1)
	r.fontIDs[key] = name
	r.fonts = append(r.fonts, pdfEmbeddedFont{
		name:      name,
		face:      face,
		data:      data,
		baseName:  baseName,
		cidByGID:  map[sfnt.GlyphIndex]uint16{},
		gidByCID:  map[uint16]sfnt.GlyphIndex{},
		runeByCID: map[uint16]rune{},
	})
	return &r.fonts[len(r.fonts)-1], true
}

// SavePDF writes the buffered document to path.
func (r *Renderer) SavePDF(path string) error {
	if len(r.document) == 0 {
		return errors.New("pdf: SavePDF called before End")
	}
	return os.WriteFile(path, r.document, 0o644)
}

// SavePDFWithOptions writes the buffered document to path using opts to
// override any setter-supplied options for this single call.
func (r *Renderer) SavePDFWithOptions(path string, opts render.PDFOptions) error {
	if !r.began && len(r.document) == 0 {
		return errors.New("pdf: SavePDFWithOptions called before End")
	}
	doc, err := buildDocument(r.width, r.height, r.content.Bytes(), r.images, r.hatchPatterns, r.fillPatterns, r.shadings, r.forms, r.alphaStates, r.fonts, opts)
	if err != nil {
		return err
	}
	return os.WriteFile(path, doc, 0o644)
}

// Bytes returns the serialized PDF document. Returns an error if End has not
// been called since the last Begin.
func (r *Renderer) Bytes() ([]byte, error) {
	if len(r.document) == 0 {
		return nil, errors.New("pdf: document is empty; call End first")
	}
	return r.document, nil
}

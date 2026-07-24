package svg

import (
	"encoding/base64"
	"encoding/xml"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/render"
)

var svgGlyphResolveWarnOnce sync.Once

func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if len(run.Glyphs) == 0 {
		return
	}

	if run.FontKey != "" {
		r.lastFontKey = run.FontKey
	}

	penX := run.Origin.X
	penY := run.Origin.Y

	size := run.Size
	if size <= 0 {
		size = 12
	}

	for _, glyph := range run.Glyphs {
		if glyph.ID == 0 {
			if glyph.Advance > 0 {
				penX += glyph.Advance
			}
			continue
		}

		// render.Glyph.ID is a font glyph index; reverse the cmap to a rune
		// rather than casting it directly.
		ch, ok := render.GlyphIDToRune(run.FontKey, glyph.ID)
		if !ok {
			svgGlyphResolveWarnOnce.Do(func() {
				diag.Warnf("svg: could not resolve glyph index %d to a rune via the font cmap; skipping", glyph.ID)
			})
			if glyph.Advance > 0 {
				penX += glyph.Advance
			}
			continue
		}

		r.DrawText(string(ch), geom.Pt{X: penX + glyph.Offset.X, Y: penY + glyph.Offset.Y}, size, textColor)

		advance := glyph.Advance
		if advance <= 0 {
			advance = r.MeasureText(string(ch), size, run.FontKey).W
		}
		penX += advance
	}
}

var (
	_ render.TextBounder      = (*Renderer)(nil)
	_ render.TextFontMetricer = (*Renderer)(nil)
)

// MeasureText returns single-line metrics from the shared pure-Go font shaper,
// the same stack TextPath uses to emit glyphs, so anchoring matches the drawn
// outlines (and the AGG raster backend) instead of the old 7x13 bitmap estimate.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" || size <= 0 {
		return render.TextMetrics{}
	}
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	return render.MeasureTextMetrics(text, size, fontKey)
}

// MeasureTextBounds reports the ink bounds of text relative to the baseline
// origin used for DrawText (render.TextBounder).
func (r *Renderer) MeasureTextBounds(text string, size float64, fontKey string) (render.TextBounds, bool) {
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	return render.MeasureTextInkBounds(text, size, fontKey)
}

// MeasureFontHeights reports font-wide vertical metrics (render.TextFontMetricer).
func (r *Renderer) MeasureFontHeights(size float64, fontKey string) (render.FontHeightMetrics, bool) {
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	return render.MeasureFontHeightMetrics(size, fontKey)
}

func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, origin, size, textColor)
		}
		return
	}
	if text == "" || size <= 0 {
		return
	}

	r.renderTextNode(text, origin.X, origin.Y, size, textColor, "", geom.Affine{}, false)
}

func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.RotatedTextDrawer); ok {
			textRen.DrawTextRotated(text, anchor, size, angle, textColor)
		} else if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, anchor, size, textColor)
		}
		return
	}
	if text == "" || size <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return
	}

	metrics := r.MeasureText(text, size, "")
	if metrics.W <= 0 || metrics.H <= 0 {
		return
	}

	origin := geom.Pt{
		X: anchor.X - metrics.W/2,
		Y: anchor.Y - metrics.Descent,
	}
	affine := rotationAffine(-angle*180/math.Pi, anchor.X, anchor.Y)
	r.renderTextNode(text, origin.X, origin.Y, size, textColor, matrixTransform(affine), affine, true)
}

func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.VerticalTextDrawer); ok {
			textRen.DrawTextVertical(text, center, size, textColor)
		} else if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, center, size, textColor)
		}
		return
	}
	if text == "" || size <= 0 {
		return
	}

	runes := []rune(text)
	lineMetrics := r.MeasureText("M", size, "")
	lineH := lineMetrics.H
	if lineH <= 0 {
		lineH = size
	}

	totalH := lineH * float64(len(runes))
	y := center.Y - totalH/2 + lineMetrics.Ascent

	for _, ch := range runes {
		s := string(ch)
		chMetrics := r.MeasureText(s, size, "")
		if chMetrics.W <= 0 || chMetrics.H <= 0 {
			continue
		}

		x := center.X - chMetrics.W/2
		r.renderTextNode(s, x, y, size, textColor, "", geom.Affine{}, false)
		y += lineH
	}
}

func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if fontKey == "" {
		fontKey = r.lastFontKey
	}
	return render.TextPath(text, origin, size, fontKey)
}

func (r *Renderer) renderTextNode(text string, x, y, size float64, textColor render.Color, transform string, affine geom.Affine, hasAffine bool) {
	if text == "" || size <= 0 {
		return
	}
	if r.options.FontPolicy == render.SVGFontPolicyPath {
		r.renderTextPathNode(text, geom.Pt{X: x, Y: y}, size, textColor, affine, hasAffine)
		return
	}

	var content strings.Builder
	content.WriteString(`<text`)
	writeFloatAttr(&content, "x", x)
	writeFloatAttr(&content, "y", r.flipY(y))
	writeFloatAttr(&content, "font-size", size)
	writeAttr(&content, "font-family", r.svgFontFamily(r.lastFontKey))
	props := render.ParseFontProperties(r.lastFontKey)
	if props.Style != "" && props.Style != render.FontStyleNormal {
		writeAttr(&content, "font-style", string(props.Style))
	}
	if props.Weight > 0 && props.Weight != 400 {
		writeAttr(&content, "font-weight", strconv.Itoa(props.Weight))
	}
	if props.Stretch != "" && props.Stretch != "normal" {
		writeAttr(&content, "font-stretch", props.Stretch)
	}
	if props.Variant != "" && props.Variant != "normal" {
		writeAttr(&content, "font-variant", props.Variant)
	}
	writeAttr(&content, "fill", colorToHex(textColor))
	alpha := clamp01(textColor.A)
	if alpha < 1 {
		writeFloatAttr(&content, "fill-opacity", alpha)
	}
	// Position is emitted in device space (y flipped). Any rotation affine is
	// expressed in y-up display space, so conjugate it by the device flip
	// (flip ∘ affine ∘ flip) to apply the same rotation about the flipped
	// anchor while keeping glyphs upright (the flip has no net vertical scale
	// after conjugation). Mirrors Matplotlib positioning text at height-y with
	// rotate(-angle) rather than flipping the glyph bitmap.
	if hasAffine {
		svgAffine := r.deviceFlip().Mul(affine).Mul(r.deviceFlip())
		writeAttr(&content, "transform", matrixTransform(svgAffine))
	} else if transform != "" {
		writeAttr(&content, "transform", transform)
	}
	content.WriteString(">")
	content.WriteString(escapeText(text))
	content.WriteString("</text>")

	r.nodes = append(r.nodes, r.newNode(content.String()))
}

func (r *Renderer) renderTextPathNode(text string, origin geom.Pt, size float64, textColor render.Color, affine geom.Affine, hasAffine bool) {
	path, ok := r.TextPath(text, origin, size, r.lastFontKey)
	if !ok {
		return
	}
	if hasAffine {
		path = affinePath(path, affine)
	}
	r.Path(path, &render.Paint{Fill: textColor})
}

func fontFamily(key string) string {
	return render.CSSFontFamily(key)
}

func (r *Renderer) svgFontFamily(key string) string {
	if family := r.registerFontFace(key); family != "" {
		return family
	}
	return fontFamily(key)
}

func (r *Renderer) registerFontFace(key string) string {
	path := strings.TrimSpace(key)
	if path == "" || !isFontFile(path) {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return ""
	}
	if r.fontFaces == nil {
		r.fontFaces = map[string]fontFaceDef{}
	}
	if face, ok := r.fontFaces[path]; ok {
		return face.family
	}
	face := fontFaceDef{
		family: "mplgo-font-" + strconv.Itoa(len(r.fontFaces)+1),
		data:   base64.StdEncoding.EncodeToString(data),
		mime:   fontMIME(path),
		format: fontFormat(path),
	}
	r.fontFaces[path] = face
	r.fontFaceOrder = append(r.fontFaceOrder, face)
	return face.family
}

func isFontFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ttf", ".otf", ".ttc", ".dfont":
		return true
	default:
		return false
	}
}

func fontMIME(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".otf":
		return "font/otf"
	default:
		return "font/ttf"
	}
}

func fontFormat(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".otf":
		return "opentype"
	default:
		return "truetype"
	}
}

func escapeText(text string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(text))
	return b.String()
}

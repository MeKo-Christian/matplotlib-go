package ps

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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

var (
	_ render.TextBounder      = (*Renderer)(nil)
	_ render.TextFontMetricer = (*Renderer)(nil)
)

// MeasureText returns single-line metrics from the shared pure-Go font shaper,
// the same stack TextPath uses to emit glyphs, so anchoring matches the drawn
// outlines.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	if size <= 0 {
		size = defaultFontHeight
	}
	return render.MeasureTextMetrics(text, size, fontKey)
}

// MeasureTextBounds reports the ink bounds of text relative to the baseline
// origin used for DrawText (render.TextBounder).
func (r *Renderer) MeasureTextBounds(text string, size float64, fontKey string) (render.TextBounds, bool) {
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	if size <= 0 {
		size = defaultFontHeight
	}
	return render.MeasureTextInkBounds(text, size, fontKey)
}

// MeasureFontHeights reports font-wide vertical metrics (render.TextFontMetricer).
func (r *Renderer) MeasureFontHeights(size float64, fontKey string) (render.FontHeightMetrics, bool) {
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	if size <= 0 {
		size = defaultFontHeight
	}
	return render.MeasureFontHeightMetrics(size, fontKey)
}

// TextPath converts text to vector glyph outlines through the shared font
// manager. render.TextPath emits y-up display outlines that the y-up PostScript
// page consumes directly, so no baseline reflection is applied.
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
	fmt.Fprintf(&r.content, "/Helvetica findfont %s scalefont setfont\n", shortFloat(size))
	fmt.Fprintf(&r.content, "0 0 moveto (%s) show\n", escapePSString(text))
	r.content.WriteString("grestore\n")
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

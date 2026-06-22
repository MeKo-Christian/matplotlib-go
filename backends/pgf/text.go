package pgf

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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

// DrawTextVertical draws vertical (stacked-glyph) text centered on the
// supplied point, matching the PS/PDF backends. Implements
// render.VerticalTextDrawer.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	r.DrawTextVerticalWithFont(text, center, size, textColor, "")
}

// DrawTextVerticalWithFont draws vertical text with an explicit font key.
// Font selection is left to LaTeX in this generator-only slice, so the key is
// used only for metric lookup. Implements render.FontVerticalTextDrawer.
func (r *Renderer) DrawTextVerticalWithFont(text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
	if rr := r.activeRaster(); rr != nil {
		switch textRen := rr.(type) {
		case render.VerticalTextDrawer:
			textRen.DrawTextVertical(text, center, size, textColor)
		case render.TextDrawer:
			textRen.DrawText(text, center, size, textColor)
		}
		return
	}
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 {
		return
	}
	runes := []rune(text)
	lineHeight := r.MeasureText("M", size, fontKey).H
	startY := center.Y - lineHeight*float64(len(runes)-1)/2
	for i, ch := range runes {
		s := string(ch)
		metrics := r.MeasureText(s, size, fontKey)
		origin := geom.Pt{X: center.X - metrics.W/2, Y: startY + float64(i)*lineHeight}
		r.writeText(s, origin, size, 0, textColor)
	}
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

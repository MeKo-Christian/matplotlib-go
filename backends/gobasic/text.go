package gobasic

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// GlyphRun renders glyph IDs as code points where available.
// The mapping is a practical fallback for renderers that expose only glyph IDs.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if len(run.Glyphs) == 0 {
		return
	}
	penX := run.Origin.X + run.Glyphs[0].Offset.X
	penY := run.Origin.Y + run.Glyphs[0].Offset.Y
	size := run.Size
	if size <= 0 {
		size = 12
	}

	for _, glyph := range run.Glyphs {
		advance := glyph.Advance
		ch := rune(glyph.ID)
		if ch > 0 {
			_ = r.MeasureText(string(ch), size, run.FontKey)
			r.DrawText(string(ch), geom.Pt{
				X: penX + glyph.Offset.X,
				Y: penY + glyph.Offset.Y,
			}, size, textColor)

			if advance <= 0 {
				advance = r.MeasureText(string(ch), size, run.FontKey).W
			}
		}
		penX += glyph.Offset.X + advance
	}
}

// MeasureText returns approximate text metrics for layout.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" || size <= 0 {
		return render.TextMetrics{}
	}
	if fontKey == "" {
		fontKey = "DejaVuSans"
	}
	r.lastFontKey = fontKey

	return measureText(text, size, fontKey, r.resolution)
}

// TextPath converts text to a vector path using the shared font manager.
func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if fontKey == "" {
		fontKey = r.lastFontKey
	}
	return render.TextPath(text, origin, size, fontKey)
}

// DrawText renders text at the requested origin.
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	if text == "" || size <= 0 {
		return
	}

	metrics := r.MeasureText(text, size, r.lastFontKey)
	if metrics.W <= 0 || metrics.H <= 0 {
		return
	}

	src := r.renderTextBitmap(text, size, textColor, r.lastFontKey)
	if src == nil {
		return
	}

	// origin.Y is the y-up display baseline; flip it to the device baseline and
	// place the upright bitmap with its top ascent above that baseline.
	x := int(math.Round(origin.X))
	y := int(math.Round(r.devY(origin.Y) - metrics.Ascent))
	r.drawBitmapScaled(src, x, y, src.Bounds().Dx(), src.Bounds().Dy())
}

// DrawTextRotated renders text using Matplotlib-like anchor rotation. The
// anchor is the bottom-center of the unrotated text box.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	if text == "" || size <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return
	}

	src := r.renderTextBitmap(text, size, textColor, r.lastFontKey)
	if src == nil {
		return
	}

	const epsilon = 1e-12
	pivotX := float64(src.Bounds().Dx()) / 2
	pivotY := float64(src.Bounds().Dy())
	// anchor is the bottom-center of the unrotated box in y-up display space;
	// flip it to the device buffer. The bitmap pivot (bottom-center) is placed at
	// the device anchor. The rotation sign stays -angle so a positive display
	// (CCW) rotation still renders CCW visually, matching the pre-y-flip behavior.
	devAnchor := r.devPt(anchor)
	if math.Abs(angle) <= epsilon {
		x := int(math.Round(devAnchor.X - pivotX))
		y := int(math.Round(devAnchor.Y - pivotY))
		r.drawBitmapScaled(src, x, y, src.Bounds().Dx(), src.Bounds().Dy())
		return
	}

	r.drawBitmapRotated(src, devAnchor, geom.Pt{X: pivotX, Y: pivotY}, -angle)
}

// DrawTextVertical renders one character per line.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	runes := []rune(text)
	if len(runes) == 0 || size <= 0 {
		return
	}

	lineHeight := r.MeasureText("M", size, r.lastFontKey).H
	if lineHeight <= 0 {
		return
	}

	totalHeight := lineHeight * float64(len(runes))
	// DrawText flips each baseline to the device buffer, so the per-character
	// layout is expressed in y-up display space: start at the top of the stack
	// (highest Y) and step downward by lineHeight per character.
	y := center.Y + totalHeight/2
	for i, ch := range runes {
		chMetrics := r.MeasureText(string(ch), size, r.lastFontKey)
		if chMetrics.W <= 0 || chMetrics.H <= 0 {
			continue
		}

		x := center.X - chMetrics.W/2
		r.DrawText(string(ch), geom.Pt{
			X: x,
			Y: y - float64(i)*lineHeight - chMetrics.Ascent,
		}, size, textColor)
	}
}

func (r *Renderer) renderTextBitmap(text string, size float64, textColor render.Color, fontKey string) *image.RGBA {
	if text == "" || size <= 0 {
		return nil
	}
	if fontKey == "" {
		fontKey = "DejaVuSans"
	}
	return renderTextBitmap(text, size, textColor, fontKey, r.resolution)
}

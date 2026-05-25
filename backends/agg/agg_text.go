package agg

import (
	"image"
	"image/png"
	"math"
	"os"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// GlyphRun draws a run of glyphs.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if len(run.Glyphs) == 0 || run.Size <= 0 || r.ctx == nil {
		return
	}

	if run.FontKey != "" {
		r.lastFontKey = run.FontKey
	}

	font := r.configureTextFont(run.Size, r.lastFontKey)
	if font.backend != textBackendRaster {
		return
	}

	if r.drawGlyphRunAsPaths(font.face, run, textColor) {
		return
	}

	_ = r.drawGlyphRunLegacy(run, textColor, r.fontPixelSize(font.size))
}

func (r *Renderer) drawGlyphRunAsPaths(face render.FontFace, run render.GlyphRun, textColor render.Color) bool {
	if r.ctx == nil {
		return false
	}

	fontSize := r.fontPixelSize(run.Size)
	if fontSize <= 0 {
		return false
	}

	fontData, err := loadSFNTFontFace(face)
	if err != nil {
		return false
	}

	ppem := fixed.Int26_6(math.Round(fontSize * 64))
	if ppem <= 0 {
		return false
	}

	var (
		path  geom.Path
		buf   sfnt.Buffer
		drawn bool
	)
	for _, glyph := range run.Glyphs {
		if glyph.ID == 0 {
			continue
		}
		segments, err := fontData.LoadGlyph(&buf, sfnt.GlyphIndex(glyph.ID), ppem, nil)
		if err != nil || len(segments) == 0 {
			continue
		}
		appendGlyphSegments(&path, segments, geom.Pt{
			X: run.Origin.X + glyph.Offset.X,
			Y: run.Origin.Y + glyph.Offset.Y,
		})
		drawn = true
	}
	if !drawn {
		return false
	}

	r.Path(path, &render.Paint{Fill: textColor})
	return true
}

func (r *Renderer) drawGlyphRunLegacy(run render.GlyphRun, textColor render.Color, size float64) bool {
	if size <= 0 {
		return false
	}

	fontPath := r.resolveTextFontPath(run.FontKey)
	if fontPath == "" {
		return false
	}

	drawn := false
	for _, glyph := range run.Glyphs {
		if glyph.ID == 0 {
			continue
		}
		origin := geom.Pt{
			X: run.Origin.X + glyph.Offset.X,
			Y: run.Origin.Y + glyph.Offset.Y,
		}
		if r.drawTextPathFallback(string(rune(glyph.ID)), origin, size, textColor, fontPath) {
			drawn = true
		}
	}
	return drawn
}

func appendGlyphSegments(path *geom.Path, segments sfnt.Segments, origin geom.Pt) {
	for _, segment := range segments {
		switch segment.Op {
		case sfnt.SegmentOpMoveTo:
			path.MoveTo(sfntPoint(segment.Args[0], origin))
		case sfnt.SegmentOpLineTo:
			path.LineTo(sfntPoint(segment.Args[0], origin))
		case sfnt.SegmentOpQuadTo:
			path.QuadTo(
				sfntPoint(segment.Args[0], origin),
				sfntPoint(segment.Args[1], origin),
			)
		case sfnt.SegmentOpCubeTo:
			path.CubicTo(
				sfntPoint(segment.Args[0], origin),
				sfntPoint(segment.Args[1], origin),
				sfntPoint(segment.Args[2], origin),
			)
		}
	}
}

func sfntPoint(p fixed.Point26_6, origin geom.Pt) geom.Pt {
	return geom.Pt{
		X: origin.X + sfntFixedToFloat(p.X),
		Y: origin.Y + sfntFixedToFloat(p.Y),
	}
}

func sfntFixedToFloat(v fixed.Int26_6) float64 {
	return float64(v) / 64.0
}

// MeasureText measures text dimensions using the active font engine.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" || size <= 0 {
		return render.TextMetrics{}
	}

	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}

	font := r.configureTextFont(size, fontKey)

	var (
		w       float64
		ascent  float64
		descent float64
	)
	switch font.backend {
	case textBackendRaster:
		if metrics, ok := r.measureRasterText(text, font.face, font.size); ok {
			return metrics
		}
		sizePx := r.fontPixelSize(font.size)
		if x, y, bw, h, ok := measureTextPathBounds(text, sizePx, font.fontPath); ok {
			w = math.Max(w, x+bw)
			ascent = math.Max(0, -y)
			descent = math.Max(0, y+h)
		} else {
			return render.TextMetrics{}
		}
	case textBackendGSV:
		sizePx := r.fontPixelSize(font.size)
		w = measureLocalGSVTextWidth(text, sizePx)
		if _, y, _, h, ok := measureLocalGSVTextBounds(text, sizePx); ok {
			ascent = math.Max(0, -y)
			descent = math.Max(0, y+h)
		} else {
			ascent = sizePx
			descent = 0
		}
	default:
		return render.TextMetrics{}
	}

	h := ascent + descent
	if h <= 0 {
		h = font.size
	}
	if ascent <= 0 {
		ascent = h
	}
	if descent < 0 {
		descent = 0
	}

	return render.TextMetrics{
		W:       w,
		H:       h,
		Ascent:  ascent,
		Descent: descent,
	}
}

// MeasureTextBounds reports the actual ink bounds of text relative to the
// baseline origin used for DrawText.
func (r *Renderer) MeasureTextBounds(text string, size float64, fontKey string) (render.TextBounds, bool) {
	if text == "" || size <= 0 {
		return render.TextBounds{}, false
	}

	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}

	font := r.configureTextFont(size, fontKey)
	if font.backend == textBackendGSV {
		sizePx := r.fontPixelSize(font.size)
		x, y, w, h, ok := measureLocalGSVTextBounds(text, sizePx)
		if !ok {
			return render.TextBounds{}, false
		}
		return render.TextBounds{X: x, Y: y, W: w, H: h}, true
	}
	if font.backend != textBackendRaster {
		return render.TextBounds{}, false
	}
	sizePx := r.fontPixelSize(font.size)
	fontKey = fontReference(font.face)
	if shaped, ok := render.ShapeText(text, geom.Pt{}, sizePx, render.TextShapingOptions{FontKey: fontKey}); ok {
		if shapedTextMatchesRuneSequence(text, shaped) {
			if bounds, ok := r.measureNativeFreetypeTextBounds(text, font.face, font.size, matplotlibTextHintingFactor); ok {
				return bounds, true
			}
		}
		return shaped.Bounds, true
	}
	if x, y, w, h, ok := measureTextPathBounds(text, sizePx, font.fontPath); ok {
		return render.TextBounds{X: x, Y: y, W: w, H: h}, true
	}
	return render.TextBounds{}, false
}

// MeasureFontHeights reports font-wide ascent, descent, and line-gap values
// for the current raster text face, distinct from a particular string's ink
// bounds.
func (r *Renderer) MeasureFontHeights(size float64, fontKey string) (render.FontHeightMetrics, bool) {
	if size <= 0 {
		return render.FontHeightMetrics{}, false
	}

	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}

	font := r.configureTextFont(size, fontKey)
	if font.backend == textBackendGSV {
		sizePx := r.fontPixelSize(font.size)
		if _, y, _, h, ok := measureLocalGSVTextBounds("lp", sizePx); ok {
			return render.FontHeightMetrics{
				Ascent:  math.Max(0, -y),
				Descent: math.Max(0, y+h),
			}, true
		}
		return render.FontHeightMetrics{}, false
	}
	if font.backend != textBackendRaster {
		return render.FontHeightMetrics{}, false
	}

	if metrics, ok := r.measureNativeFreetypeFontHeights(font.face, font.size, matplotlibTextHintingFactor); ok {
		return metrics, true
	}

	if metrics, ok := rasterFontHeightMetrics(font.face, font.size, r.resolution); ok {
		return render.FontHeightMetrics{
			Ascent:  metrics.ascent,
			Descent: metrics.descent,
			LineGap: metrics.lineGap,
		}, true
	}

	if err := r.ctx.ConfigureTextFont(font.fontPath, font.size, r.resolution); err != nil {
		sizePx := r.fontPixelSize(font.size)
		if _, y, _, h, ok := measureTextPathBounds("lp", sizePx, font.fontPath); ok {
			return render.FontHeightMetrics{
				Ascent:  math.Max(0, -y),
				Descent: math.Max(0, y+h),
			}, true
		}
		return render.FontHeightMetrics{}, false
	}

	metrics, ok := r.ctx.fontHeightMetrics()
	if !ok {
		return render.FontHeightMetrics{}, false
	}
	return render.FontHeightMetrics{
		Ascent:  metrics.ascent,
		Descent: metrics.descent,
		LineGap: metrics.lineGap,
	}, true
}

func rasterFontHeightMetrics(face render.FontFace, size float64, resolution uint) (fontHeightMetrics, bool) {
	if fontReference(face) == "" || size <= 0 {
		return fontHeightMetrics{}, false
	}
	dpi := float64(resolution)
	if dpi <= 0 {
		dpi = 72
	}
	resource, err := loadSFNTFontFace(face)
	if err != nil {
		return fontHeightMetrics{}, false
	}
	scale, ok := sfntMetricScale(resource.data, size, dpi)
	if !ok {
		return fontHeightMetrics{}, false
	}
	ascent, descent, lineGap, ok := sfntTableHeightMetrics(resource.data, scale)
	if !ok {
		return fontHeightMetrics{}, false
	}
	return fontHeightMetrics{
		ascent:  ascent,
		descent: descent,
		lineGap: lineGap,
	}, true
}

// TextPath converts text to a vector path using the renderer's resolved font.
func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if text == "" || size <= 0 {
		return geom.Path{}, false
	}
	if fontKey == "" {
		fontKey = r.lastFontKey
	}
	font := r.configureTextFont(size, fontKey)
	if fontReference(font.face) == "" {
		return geom.Path{}, false
	}
	return render.TextPath(text, origin, r.fontPixelSize(size), fontReference(font.face))
}

// DrawText renders text at the given position with the specified size and color.
// This is a helper method (not part of the Renderer interface).
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	r.drawTextWithFontContext(text, origin, size, textColor, r.lastFontKey)
}

// DrawTextWithFont renders text with an explicit font key, avoiding the legacy
// MeasureText-before-DrawText state handoff.
func (r *Renderer) DrawTextWithFont(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	r.withTemporaryFontKey(func() {
		r.drawTextWithFontContext(text, origin, size, textColor, fontKey)
	})
}

func (r *Renderer) drawTextWithFontContext(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	if r.hasClipPath() {
		bounds, haveBounds := r.textDrawBounds(text, origin, size, fontKey)
		r.withClipPathMask(bounds, haveBounds, func() {
			r.drawTextDirect(text, origin, size, textColor, fontKey)
		})
		return
	}
	r.drawTextDirect(text, origin, size, textColor, fontKey)
}

func (r *Renderer) drawTextDirect(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	if text == "" || size <= 0 {
		return
	}

	font := r.configureTextFont(size, fontKey)

	switch font.backend {
	case textBackendRaster:
		// drawRasterText flips the origin to device internally; drawTextPathFallback
		// routes through r.Path which flips itself. Both receive the display origin.
		if r.drawRasterText(text, font.face, origin, font.size, textColor) {
			return
		}
		sizePx := r.fontPixelSize(font.size)
		if r.drawTextPathFallback(text, origin, sizePx, textColor, font.fontPath) {
			return
		}
		// drawEmergencyGSVText draws directly through the painter (device space).
		if r.drawEmergencyGSVText(text, r.devPt(origin), font.size, textColor) {
			return
		}
		return
	case textBackendGSV:
		_ = r.drawEmergencyGSVText(text, r.devPt(origin), font.size, textColor)
	default:
		_ = r.drawEmergencyGSVText(text, r.devPt(origin), size, textColor)
		return
	}
}

// DrawTextRotated renders text using Matplotlib-like anchor rotation. The
// anchor is the bottom-center of the unrotated text box.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	r.drawTextRotatedWithFontContext(text, anchor, size, angle, textColor, r.lastFontKey)
}

// DrawTextRotatedWithFont renders rotated text with an explicit font key.
func (r *Renderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	r.withTemporaryFontKey(func() {
		r.drawTextRotatedWithFontContext(text, anchor, size, angle, textColor, fontKey)
	})
}

func (r *Renderer) drawTextRotatedWithFontContext(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	if r.hasClipPath() {
		bounds, haveBounds := r.rotatedTextDrawBounds(text, anchor, size, angle, fontKey)
		r.withClipPathMask(bounds, haveBounds, func() {
			r.drawTextRotatedDirect(text, anchor, size, angle, textColor, fontKey)
		})
		return
	}
	r.drawTextRotatedDirect(text, anchor, size, angle, textColor, fontKey)
}

func (r *Renderer) drawTextRotatedDirect(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	if text == "" || size <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return
	}

	metrics := r.MeasureText(text, size, fontKey)
	if metrics.W <= 0 || metrics.H <= 0 {
		return
	}

	bounds, haveBounds := r.MeasureTextBounds(text, size, fontKey)
	origin := rotatedTextOrigin(anchor, metrics, bounds, haveBounds)
	font := r.configureTextFont(size, fontKey)

	// The painter transform operates in y-down device space. The geometry it
	// receives is flipped to device (path fallback flips via r.Path; the direct
	// outline/GSV draws below use a device origin). Rotate about the DEVICE
	// anchor; the rotation sign stays -angle so a positive display-space (CCW)
	// rotation still renders CCW visually, matching the pre-y-flip behavior.
	devAnchor := r.devPt(anchor)
	devOrigin := r.devPt(origin)

	r.ctx.PushTransform()
	defer r.ctx.PopTransform()

	r.ctx.Translate(-devAnchor.X, -devAnchor.Y)
	r.ctx.Rotate(-angle)
	r.ctx.Translate(devAnchor.X, devAnchor.Y)

	if fontReference(font.face) != "" {
		sizePx := r.fontPixelSize(font.size)
		fontKey := fontReference(font.face)
		// drawTextPathFallback routes through r.Path, which flips the display
		// origin to device itself, so pass the un-flipped display origin here.
		if r.drawTextPathFallback(text, origin, sizePx, textColor, fontKey) {
			return
		}
		if font.fontPath != "" {
			if face, err := r.configureOutlineFont(font.fontPath, sizePx); err == nil {
				r.ctx.SetFillColor(renderColorToAGG(textColor))
				r.ctx.SetStrokeColor(renderColorToAGG(textColor))
				// drawTrueTypeOutlineText draws directly through the painter, so
				// it needs the device-space origin.
				if drawTrueTypeOutlineText(r.ctx, face, devOrigin.X, devOrigin.Y, text) {
					return
				}
			}
		}
	}

	if font.backend == textBackendGSV {
		// drawEmergencyGSVText draws directly through the painter (device space).
		_ = r.drawEmergencyGSVText(text, devOrigin, font.size, textColor)
		return
	}
	_ = r.drawEmergencyGSVText(text, devOrigin, size, textColor)
}

func (r *Renderer) fontPixelSize(size float64) float64 {
	dpi := float64(r.resolution)
	if dpi <= 0 {
		dpi = 72
	}
	return size * dpi / 72.0
}

func (r *Renderer) drawTextPathFallback(text string, origin geom.Pt, size float64, textColor render.Color, fontPath string) bool {
	if fontPath == "" {
		return false
	}
	path, ok := render.TextPath(text, origin, size, fontPath)
	if !ok {
		return false
	}
	r.Path(path, &render.Paint{
		Fill: textColor,
	})
	return true
}

func (r *Renderer) drawEmergencyGSVText(text string, origin geom.Pt, size float64, textColor render.Color) bool {
	if r == nil || r.ctx == nil || !r.emergencyTextFallback || text == "" || size <= 0 {
		return false
	}
	lineWidth := r.ctx.painter.GetLineWidth()
	lineCap := r.ctx.painter.GetLineCap()
	lineJoin := r.ctx.painter.GetLineJoin()
	lineColor := r.ctx.painter.GetLineColor()
	defer func() {
		r.ctx.painter.LineWidth(lineWidth)
		r.ctx.painter.LineCap(lineCap)
		r.ctx.painter.LineJoin(lineJoin)
		r.ctx.painter.LineColor(lineColor)
	}()

	sizePx := r.fontPixelSize(size)
	r.ctx.SetStrokeColor(renderColorToAGG(textColor))
	r.ctx.SetStrokeWidth(math.Max(1, sizePx*0.08))
	r.ctx.SetLineCap(agglib.CapRound)
	r.ctx.SetLineJoin(agglib.JoinRound)
	if !appendLocalGSVText(r.ctx, origin.X, origin.Y, sizePx, text) {
		return false
	}
	r.ctx.Stroke()
	r.markEmergencyTextFallback()
	return true
}

func rotatedTextOrigin(anchor geom.Pt, metrics render.TextMetrics, bounds render.TextBounds, haveBounds bool) geom.Pt {
	return geom.Pt{
		X: anchor.X - metrics.W/2,
		Y: anchor.Y - metrics.Descent,
	}
}

func (r *Renderer) textDrawBounds(text string, origin geom.Pt, size float64, fontKey string) (geom.Rect, bool) {
	if text == "" || size <= 0 {
		return geom.Rect{}, false
	}
	if bounds, ok := r.MeasureTextBounds(text, size, fontKey); ok && bounds.W > 0 && bounds.H > 0 {
		return geom.Rect{
			Min: geom.Pt{X: origin.X + bounds.X, Y: origin.Y + bounds.Y},
			Max: geom.Pt{X: origin.X + bounds.X + bounds.W, Y: origin.Y + bounds.Y + bounds.H},
		}.Inflate(2, 2), true
	}
	metrics := r.MeasureText(text, size, fontKey)
	if metrics.W <= 0 || metrics.H <= 0 {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent},
		Max: geom.Pt{X: origin.X + metrics.W, Y: origin.Y + metrics.Descent},
	}.Inflate(2, 2), true
}

func (r *Renderer) rotatedTextDrawBounds(text string, anchor geom.Pt, size, angle float64, fontKey string) (geom.Rect, bool) {
	if text == "" || size <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return geom.Rect{}, false
	}
	metrics := r.MeasureText(text, size, fontKey)
	if metrics.W <= 0 || metrics.H <= 0 {
		return geom.Rect{}, false
	}
	bounds, haveBounds := r.MeasureTextBounds(text, size, fontKey)
	origin := rotatedTextOrigin(anchor, metrics, bounds, haveBounds)
	var rect geom.Rect
	if haveBounds && bounds.W > 0 && bounds.H > 0 {
		rect = geom.Rect{
			Min: geom.Pt{X: origin.X + bounds.X, Y: origin.Y + bounds.Y},
			Max: geom.Pt{X: origin.X + bounds.X + bounds.W, Y: origin.Y + bounds.Y + bounds.H},
		}
	} else {
		rect = geom.Rect{
			Min: geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent},
			Max: geom.Pt{X: origin.X + metrics.W, Y: origin.Y + metrics.Descent},
		}
	}
	return rotatedRectBounds(rect, anchor, -angle)
}

func rotatedRectBounds(rect geom.Rect, anchor geom.Pt, angle float64) (geom.Rect, bool) {
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	rotate := func(pt geom.Pt) geom.Pt {
		dx := pt.X - anchor.X
		dy := pt.Y - anchor.Y
		return geom.Pt{
			X: anchor.X + dx*cosA - dy*sinA,
			Y: anchor.Y + dx*sinA + dy*cosA,
		}
	}
	return pointsBounds([]geom.Pt{
		rotate(rect.Min),
		rotate(geom.Pt{X: rect.Max.X, Y: rect.Min.Y}),
		rotate(rect.Max),
		rotate(geom.Pt{X: rect.Min.X, Y: rect.Max.Y}),
	}, 3)
}

// DrawTextVertical renders text vertically (one character per line, top to bottom).
// This is used for ylabel rendering where true rotation is not available.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	if text == "" || size <= 0 {
		return
	}
	r.drawTextVerticalWithFontContext(text, center, size, textColor, r.lastFontKey)
}

// DrawTextVerticalWithFont renders vertical text with an explicit font key.
func (r *Renderer) DrawTextVerticalWithFont(text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
	if text == "" || size <= 0 {
		return
	}
	r.withTemporaryFontKey(func() {
		r.drawTextVerticalWithFontContext(text, center, size, textColor, fontKey)
	})
}

func (r *Renderer) drawTextVerticalWithFontContext(text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
	if r.hasClipPath() {
		metrics := r.MeasureText(text, size, fontKey)
		if metrics.W <= 0 || metrics.H <= 0 {
			return
		}
		// drawTextDirect flips the baseline origin to device itself, so the
		// painter transform must rotate about the DEVICE center. The angle sign
		// is negated (Rotate(+π/2) = Rotate(-(-π/2))) to match drawTextRotatedDirect's
		// device-space convention, keeping the visual orientation unchanged.
		devCenter := r.devPt(center)
		r.ctx.PushTransform()
		defer r.ctx.PopTransform()
		r.ctx.Translate(devCenter.X, devCenter.Y)
		r.ctx.Rotate(math.Pi / 2)
		r.ctx.Translate(-devCenter.X, -devCenter.Y)
		r.drawTextDirect(text, geom.Pt{X: center.X + metrics.Descent, Y: center.Y}, size, textColor, fontKey)
		return
	}
	r.drawTextRotatedDirect(text, center, size, -math.Pi/2, textColor, fontKey)
}

func (r *Renderer) withTemporaryFontKey(draw func()) {
	previous := r.lastFontKey
	defer func() {
		r.lastFontKey = previous
	}()
	draw()
}

// GetImage returns the rendered image as a standard Go image.RGBA.
func (r *Renderer) GetImage() *image.RGBA {
	return r.ctx.GetImage().ToGoImage()
}

// SavePNG saves the rendered image to a PNG file.
//
// The AGG render buffer holds straight (non-premultiplied) alpha, so GetImage
// returns straight RGBA bytes in an image.RGBA (whose channels are
// premultiplied by Go's contract). Encoding that directly makes png.Encode
// apply an unclamped premultiplied->straight conversion that overflows for any
// channel where value>alpha, corrupting semi-transparent fills over a
// transparent surface. Reinterpret the identical byte layout as NRGBA so the
// straight values are written verbatim.
func (r *Renderer) SavePNG(path string) error {
	img := r.GetImage()
	nrgba := &image.NRGBA{Pix: img.Pix, Stride: img.Stride, Rect: img.Rect}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, nrgba)
}

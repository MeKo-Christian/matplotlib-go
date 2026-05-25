package agg

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"sync"
	"unicode/utf8"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var (
	rasterFontCacheMu sync.RWMutex
	rasterFontCache   = map[string]*opentype.Font{}
)

const matplotlibTextHintingFactor = 8

func loadRasterFont(face render.FontFace) (*opentype.Font, error) {
	key := rasterFontCacheKey(face)
	if key == "" {
		return nil, errors.New("agg: raster font face has no path or data")
	}
	rasterFontCacheMu.RLock()
	if cached, ok := rasterFontCache[key]; ok {
		rasterFontCacheMu.RUnlock()
		return cached, nil
	}
	rasterFontCacheMu.RUnlock()

	fontData := face.Data
	if face.Path != "" {
		var err error
		fontData, err = os.ReadFile(face.Path)
		if err != nil {
			return nil, err
		}
	}

	parsed, err := opentype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	rasterFontCacheMu.Lock()
	rasterFontCache[key] = parsed
	rasterFontCacheMu.Unlock()
	return parsed, nil
}

func (r *Renderer) openRasterFace(face render.FontFace, size float64) (font.Face, error) {
	parsed, err := loadRasterFont(face)
	if err != nil {
		return nil, err
	}

	dpi := float64(r.resolution)
	if dpi <= 0 {
		dpi = 72
	}

	return opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

func (r *Renderer) measureRasterText(text string, face render.FontFace, size float64) (render.TextMetrics, bool) {
	fontKey := fontReference(face)
	if text == "" || fontKey == "" || size <= 0 {
		return render.TextMetrics{}, false
	}

	fontFace, err := r.openRasterFace(face, size)
	if err != nil {
		return render.TextMetrics{}, false
	}
	defer func() { _ = fontFace.Close() }()

	shaped, ok := render.ShapeText(text, geom.Pt{}, r.fontPixelSize(size), render.TextShapingOptions{FontKey: fontKey})
	if !ok {
		return render.TextMetrics{}, false
	}
	advance := shaped.Advance.X
	bounds := shaped.Bounds
	runeEquivalent := shapedTextMatchesRuneSequence(text, shaped)
	if runeEquivalent {
		if nativeMetrics, ok := r.measureNativeFreetypeText(text, face, size, matplotlibTextHintingFactor); ok {
			advance = nativeMetrics.W
		}
		if nativeBounds, ok := r.measureNativeFreetypeTextBounds(text, face, size, matplotlibTextHintingFactor); ok {
			bounds = nativeBounds
		}
	}

	metrics := fontFace.Metrics()
	ascent := quantize(float64(metrics.Ascent.Ceil()))
	descent := quantize(float64(metrics.Descent.Ceil()))
	height := quantize(float64(metrics.Height.Ceil()))
	if fontHeights, ok := r.measureNativeFreetypeFontHeights(face, size, matplotlibTextHintingFactor); ok {
		ascent = math.Max(fontHeights.Ascent, math.Max(0, -bounds.Y))
		descent = math.Max(fontHeights.Descent, math.Max(0, bounds.Y+bounds.H))
		height = ascent + descent
	} else if fontHeights, ok := rasterFontHeightMetrics(face, size, r.resolution); ok {
		ascent = math.Max(fontHeights.ascent, math.Max(0, -bounds.Y))
		descent = math.Max(fontHeights.descent, math.Max(0, bounds.Y+bounds.H))
		height = ascent + descent
	}
	return render.TextMetrics{
		W:       quantize(advance),
		H:       height,
		Ascent:  ascent,
		Descent: descent,
	}, true
}

func (r *Renderer) drawRasterText(text string, face render.FontFace, origin geom.Pt, size float64, textColor render.Color) bool {
	fontKey := fontReference(face)
	if text == "" || fontKey == "" || size <= 0 {
		return false
	}

	// origin arrives in y-up display space. Flip the baseline to the y-down
	// device buffer (matplotlib flipy()==True: glyph positions flip, the glyph
	// bitmaps stay upright). The upright bitmap rasterization below is unchanged.
	origin = r.devPt(origin)

	shaped, ok := render.ShapeText(text, geom.Pt{}, r.fontPixelSize(size), render.TextShapingOptions{FontKey: fontKey})
	if !ok {
		return false
	}
	if shapedTextMatchesRuneSequence(text, shaped) && r.drawNativeFreetypeRunText(text, face, origin, size, textColor, matplotlibTextHintingFactor) {
		return true
	}
	if !shapedGlyphsDrawableByRune(shaped) {
		return false
	}

	primaryFace, err := r.openRasterFace(face, size)
	if err != nil {
		return false
	}
	defer func() { _ = primaryFace.Close() }()

	faces := map[string]font.Face{rasterFontCacheKey(face): primaryFace}
	defer closeRasterFaces(faces, rasterFontCacheKey(face))

	uniform := image.NewUniform(renderColorToRGBA(textColor))
	drewGlyph := false
	for _, glyph := range shaped.Glyphs {
		glyphFace := glyph.Face
		if rasterFontCacheKey(glyphFace) == "" {
			glyphFace = face
		}
		face, err := r.rasterFaceForFace(faces, glyphFace, size)
		if err != nil {
			return false
		}
		dot := fixed.Point26_6{
			X: fixed.Int26_6(math.Round((origin.X + glyph.Origin.X) * 64.0)),
			Y: fixed.Int26_6(math.Round(origin.Y * 64.0)),
		}
		dr, mask, maskp, _, ok := face.Glyph(dot, glyph.Rune)
		if !ok || dr.Empty() {
			continue
		}

		src := image.NewRGBA(image.Rect(0, 0, dr.Dx(), dr.Dy()))
		draw.DrawMask(src, src.Bounds(), uniform, image.Point{}, mask, maskp, draw.Over)
		img, err := agglib.NewImageFromStandardImage(src)
		if err != nil {
			return false
		}
		if err := r.ctx.DrawImageScaled(img, float64(dr.Min.X), float64(dr.Min.Y), float64(dr.Dx()), float64(dr.Dy())); err != nil {
			return false
		}
		drewGlyph = true
	}

	return drewGlyph || len(shaped.Glyphs) == 0 || shaped.Advance.X > 0
}

func shapedTextMatchesRuneSequence(text string, shaped render.ShapedText) bool {
	if utf8.RuneCountInString(text) != len(shaped.Glyphs) {
		return false
	}
	i := 0
	for _, r := range text {
		if shaped.Glyphs[i].Rune != r {
			return false
		}
		i++
	}
	return true
}

func shapedGlyphsDrawableByRune(shaped render.ShapedText) bool {
	resources := map[string]*sfntFontResource{}
	buffers := map[string]*sfnt.Buffer{}
	for _, glyph := range shaped.Glyphs {
		if glyph.Rune == 0 || glyph.GlyphIndex == 0 {
			return false
		}
		key := sfntFontCacheKey(glyph.Face)
		if key == "" {
			return false
		}
		resource := resources[key]
		if resource == nil {
			var err error
			resource, err = loadSFNTFontFace(glyph.Face)
			if err != nil {
				return false
			}
			resources[key] = resource
		}
		buf := buffers[key]
		if buf == nil {
			buf = &sfnt.Buffer{}
			buffers[key] = buf
		}
		index, err := resource.font.GlyphIndex(buf, glyph.Rune)
		if err != nil || index != glyph.GlyphIndex {
			return false
		}
	}
	return true
}

func (r *Renderer) rasterFaceForFace(faces map[string]font.Face, fontFace render.FontFace, size float64) (font.Face, error) {
	key := rasterFontCacheKey(fontFace)
	if face := faces[key]; face != nil {
		return face, nil
	}
	face, err := r.openRasterFace(fontFace, size)
	if err != nil {
		return nil, err
	}
	faces[key] = face
	return face, nil
}

func closeRasterFaces(faces map[string]font.Face, keepPath string) {
	for fontPath, face := range faces {
		if fontPath != keepPath && face != nil {
			_ = face.Close()
		}
	}
}

func rasterFontCacheKey(face render.FontFace) string {
	if face.Path != "" {
		return "path:" + filepath.Clean(face.Path)
	}
	if face.Family != "" && len(face.Data) > 0 {
		return "embedded:" + face.Family
	}
	return ""
}

func renderColorToRGBA(c render.Color) color.RGBA {
	return color.RGBA{
		R: uint8(math.Round(clamp01(c.R) * 255)),
		G: uint8(math.Round(clamp01(c.G) * 255)),
		B: uint8(math.Round(clamp01(c.B) * 255)),
		A: uint8(math.Round(clamp01(c.A) * 255)),
	}
}

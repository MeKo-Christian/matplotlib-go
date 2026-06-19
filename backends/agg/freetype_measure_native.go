//go:build cgo && !purego

package agg

/*
#cgo !systemfreetype CFLAGS: -I${SRCDIR}/../../third_party/freetype/prefix/include/freetype2
#cgo !systemfreetype LDFLAGS: -L${SRCDIR}/../../third_party/freetype/prefix/lib -lfreetype -lm
#cgo systemfreetype pkg-config: freetype2
#include <stdlib.h>
#include <ft2build.h>
#include FT_FREETYPE_H
#include FT_GLYPH_H

static FT_Int32 mpl_go_measure_force_autohint_load_flags(void) {
	return FT_LOAD_DEFAULT | FT_LOAD_FORCE_AUTOHINT;
}

static int mpl_go_measure_has_kerning(FT_Face face) {
	return FT_HAS_KERNING(face);
}
*/
import "C"

import (
	"math"
	"sync"
	"unsafe"

	"github.com/cwbudde/matplotlib-go/render"
)

type nativeFreetypeMeasureKey struct {
	fontPath      string
	text          string
	size          float64
	dpi           uint
	hintingFactor int
}

type nativeFreetypeMeasureCacheEntry struct {
	bounds  render.TextBounds
	metrics render.TextMetrics
}

var (
	nativeFreetypeMeasureCacheMu sync.RWMutex
	nativeFreetypeMeasureCache   = map[nativeFreetypeMeasureKey]nativeFreetypeMeasureCacheEntry{}
)

func (r *Renderer) measureNativeFreetypeText(text string, face render.FontFace, size float64, hintingFactor int) (render.TextMetrics, bool) {
	bounds, metrics, ok := r.measureNativeFreetypeTextRun(text, face, size, hintingFactor)
	if !ok {
		return render.TextMetrics{}, false
	}
	descent := math.Max(0, bounds.Y+bounds.H)
	return render.TextMetrics{
		W:       metrics.W,
		H:       bounds.H,
		Ascent:  math.Max(0, bounds.H-descent),
		Descent: descent,
	}, true
}

func (r *Renderer) measureNativeFreetypeTextBounds(text string, face render.FontFace, size float64, hintingFactor int) (render.TextBounds, bool) {
	bounds, _, ok := r.measureNativeFreetypeTextRun(text, face, size, hintingFactor)
	return bounds, ok
}

func (r *Renderer) measureNativeFreetypeFontHeights(face render.FontFace, size float64, hintingFactor int) (render.FontHeightMetrics, bool) {
	bounds, _, ok := r.measureNativeFreetypeTextRun("lp", face, size, hintingFactor)
	if !ok {
		return render.FontHeightMetrics{}, false
	}
	return render.FontHeightMetrics{
		Ascent:  math.Max(0, -bounds.Y),
		Descent: math.Max(0, bounds.Y+bounds.H),
	}, true
}

func (r *Renderer) measureNativeFreetypeTextRun(text string, face render.FontFace, size float64, hintingFactor int) (render.TextBounds, render.TextMetrics, bool) {
	if text == "" || face.Path == "" || size <= 0 {
		return render.TextBounds{}, render.TextMetrics{}, false
	}
	dpi := r.resolution
	if dpi == 0 {
		dpi = 72
	}
	if hintingFactor <= 0 {
		hintingFactor = 1
	}
	key := nativeFreetypeMeasureKey{
		fontPath:      face.Path,
		text:          text,
		size:          size,
		dpi:           dpi,
		hintingFactor: hintingFactor,
	}
	nativeFreetypeMeasureCacheMu.RLock()
	if cached, ok := nativeFreetypeMeasureCache[key]; ok {
		nativeFreetypeMeasureCacheMu.RUnlock()
		return cached.bounds, cached.metrics, true
	}
	nativeFreetypeMeasureCacheMu.RUnlock()

	var bounds render.TextBounds
	var metrics render.TextMetrics
	ok := withNativeFreetypeRun(face.Path, text, size, dpi, hintingFactor, func(run nativeFreetypeRun) bool {
		bounds = render.TextBounds{
			X: float64(run.bbox.xMin) / 64.0,
			Y: -float64(run.bbox.yMax) / 64.0,
			W: float64(run.bbox.xMax-run.bbox.xMin) / 64.0,
			H: float64(run.bbox.yMax-run.bbox.yMin) / 64.0,
		}
		metrics = render.TextMetrics{W: run.advance}
		return true
	})
	if ok {
		nativeFreetypeMeasureCacheMu.Lock()
		nativeFreetypeMeasureCache[key] = nativeFreetypeMeasureCacheEntry{bounds: bounds, metrics: metrics}
		nativeFreetypeMeasureCacheMu.Unlock()
	}
	return bounds, metrics, ok
}

// measureNativeFreetypeGlyphRun returns matplotlib `_get_info` metrics for every
// glyph in text, used for pixel-exact mathtext layout/rasterization. It mirrors
// matplotlib FT2Font.set_size exactly: horizontal resolution dpi*hintingFactor +
// a face transform xx=1/hintingFactor, FORCE_AUTOHINT load. linearHoriAdvance is
// unhinted (the wrapper divides it by hintingFactor — ft2font_wrapper.cpp:313),
// horiBearingY/height are vertical (no hf), and the bbox is the transformed glyph
// CBox (1x). See render.MathGlyphMetric.
func (r *Renderer) measureNativeFreetypeGlyphRun(text string, fontPath string, size float64, hintingFactor int) ([]render.MathGlyphMetric, bool) {
	if text == "" || fontPath == "" || size <= 0 {
		return nil, false
	}
	if hintingFactor <= 0 {
		hintingFactor = 1
	}
	dpi := r.resolution
	if dpi == 0 {
		dpi = 72
	}

	var library C.FT_Library
	if C.FT_Init_FreeType(&library) != 0 {
		return nil, false
	}
	defer C.FT_Done_FreeType(library)

	path := C.CString(fontPath)
	defer C.free(unsafe.Pointer(path))

	var ftFace C.FT_Face
	if C.FT_New_Face(library, path, 0, &ftFace) != 0 {
		return nil, false
	}
	defer C.FT_Done_Face(ftFace)

	charSize := C.FT_F26Dot6(math.Round(size * 64.0))
	if C.FT_Set_Char_Size(ftFace, charSize, 0, C.FT_UInt(dpi*uint(hintingFactor)), C.FT_UInt(dpi)) != 0 {
		return nil, false
	}
	matrix := C.FT_Matrix{
		xx: C.FT_Fixed(math.Round(65536.0 / float64(hintingFactor))),
		yy: 0x10000,
	}
	C.FT_Set_Transform(ftFace, &matrix, nil)
	loadFlags := C.mpl_go_measure_force_autohint_load_flags()
	hf := float64(hintingFactor)

	runes := []rune(text)
	out := make([]render.MathGlyphMetric, 0, len(runes))
	var previousGlyph C.FT_UInt
	for _, rr := range runes {
		glyphIndex := C.FT_Get_Char_Index(ftFace, C.FT_ULong(rr))
		if glyphIndex == 0 {
			return nil, false
		}
		kern := 0.0
		if previousGlyph != 0 && C.mpl_go_measure_has_kerning(ftFace) != 0 {
			var kerning C.FT_Vector
			if C.FT_Get_Kerning(ftFace, previousGlyph, glyphIndex, C.FT_KERNING_DEFAULT, &kerning) == 0 {
				kern = float64(kerning.x) / hf / 64.0
			}
		}
		if C.FT_Load_Glyph(ftFace, glyphIndex, loadFlags) != 0 {
			return nil, false
		}
		slot := ftFace.glyph
		var glyph C.FT_Glyph
		if C.FT_Get_Glyph(slot, &glyph) != 0 {
			return nil, false
		}
		var cbox C.FT_BBox
		C.FT_Glyph_Get_CBox(glyph, C.FT_GLYPH_BBOX_SUBPIXELS, &cbox)
		C.FT_Done_Glyph(glyph)

		out = append(out, render.MathGlyphMetric{
			Advance:    float64(slot.linearHoriAdvance) / hf / 65536.0,
			Iceberg:    float64(slot.metrics.horiBearingY) / 64.0,
			Height:     float64(slot.metrics.height) / 64.0,
			Xmin:       float64(cbox.xMin) / 64.0,
			Xmax:       float64(cbox.xMax) / 64.0,
			Ymin:       float64(cbox.yMin) / 64.0,
			Ymax:       float64(cbox.yMax) / 64.0,
			KernToPrev: kern,
		})
		previousGlyph = glyphIndex
	}
	return out, true
}

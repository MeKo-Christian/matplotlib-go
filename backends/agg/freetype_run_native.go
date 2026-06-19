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

static FT_Int32 mpl_go_run_force_autohint_load_flags(void) {
	return FT_LOAD_DEFAULT | FT_LOAD_FORCE_AUTOHINT;
}

static int mpl_go_run_has_kerning(FT_Face face) {
	return FT_HAS_KERNING(face);
}

static FT_Bitmap *mpl_go_run_bitmap_glyph_bitmap(FT_Glyph glyph) {
	return &((FT_BitmapGlyph)glyph)->bitmap;
}

static FT_Int mpl_go_run_bitmap_glyph_left(FT_Glyph glyph) {
	return ((FT_BitmapGlyph)glyph)->left;
}

static FT_Int mpl_go_run_bitmap_glyph_top(FT_Glyph glyph) {
	return ((FT_BitmapGlyph)glyph)->top;
}
*/
import "C"

import (
	"image"
	"math"
	"unsafe"
)

type nativeFreetypeRun struct {
	glyphs  []C.FT_Glyph
	bbox    C.FT_BBox
	advance float64
}

func withNativeFreetypeRun(fontPath, text string, size float64, dpi uint, hintingFactor int, fn func(nativeFreetypeRun) bool) bool {
	if fontPath == "" || text == "" || size <= 0 || fn == nil {
		return false
	}
	if hintingFactor <= 0 {
		hintingFactor = 1
	}

	var library C.FT_Library
	if C.FT_Init_FreeType(&library) != 0 {
		return false
	}
	defer C.FT_Done_FreeType(library)

	path := C.CString(fontPath)
	defer C.free(unsafe.Pointer(path))

	var ftFace C.FT_Face
	if C.FT_New_Face(library, path, 0, &ftFace) != 0 {
		return false
	}
	defer C.FT_Done_Face(ftFace)

	charSize := C.FT_F26Dot6(math.Round(size * 64.0))
	// Match Matplotlib's legacy text.hinting_factor trick: hint on a denser
	// horizontal grid, then shrink the outline back after loading.
	if C.FT_Set_Char_Size(ftFace, charSize, 0, C.FT_UInt(dpi*uint(hintingFactor)), C.FT_UInt(dpi)) != 0 {
		return false
	}
	matrix := C.FT_Matrix{
		xx: C.FT_Fixed(math.Round(65536.0 / float64(hintingFactor))),
		yy: 0x10000,
	}
	C.FT_Set_Transform(ftFace, &matrix, nil)

	loadFlags := C.mpl_go_run_force_autohint_load_flags()
	glyphs := make([]C.FT_Glyph, 0, len([]rune(text)))
	defer func() { freeNativeGlyphs(glyphs) }()
	var bbox C.FT_BBox
	bbox.xMin, bbox.yMin = 32000, 32000
	bbox.xMax, bbox.yMax = -32000, -32000
	var pen C.FT_Vector
	var previousGlyph C.FT_UInt
	haveGlyph := false
	haveBox := false

	for _, rr := range text {
		glyphIndex := C.FT_Get_Char_Index(ftFace, C.FT_ULong(rr))
		if glyphIndex == 0 {
			return false
		}
		if previousGlyph != 0 && C.mpl_go_run_has_kerning(ftFace) != 0 {
			var kerning C.FT_Vector
			if C.FT_Get_Kerning(ftFace, previousGlyph, glyphIndex, C.FT_KERNING_DEFAULT, &kerning) == 0 {
				pen.x += C.FT_Pos(int(kerning.x) / hintingFactor)
				pen.y += C.FT_Pos(int(kerning.y) / hintingFactor)
			}
		}
		if C.FT_Load_Glyph(ftFace, glyphIndex, loadFlags) != 0 {
			return false
		}
		var glyph C.FT_Glyph
		if C.FT_Get_Glyph(ftFace.glyph, &glyph) != 0 {
			return false
		}
		C.FT_Glyph_Transform(glyph, nil, &pen)

		var glyphBox C.FT_BBox
		C.FT_Glyph_Get_CBox(glyph, C.FT_GLYPH_BBOX_SUBPIXELS, &glyphBox)
		if !haveBox {
			bbox = glyphBox
			haveBox = true
		} else {
			if glyphBox.xMin < bbox.xMin {
				bbox.xMin = glyphBox.xMin
			}
			if glyphBox.yMin < bbox.yMin {
				bbox.yMin = glyphBox.yMin
			}
			if glyphBox.xMax > bbox.xMax {
				bbox.xMax = glyphBox.xMax
			}
			if glyphBox.yMax > bbox.yMax {
				bbox.yMax = glyphBox.yMax
			}
		}
		if glyphBox.xMin < glyphBox.xMax && glyphBox.yMin < glyphBox.yMax {
			glyphs = append(glyphs, glyph)
		} else {
			C.FT_Done_Glyph(glyph)
		}

		pen.x += ftFace.glyph.advance.x
		pen.y += ftFace.glyph.advance.y
		previousGlyph = glyphIndex
		haveGlyph = true
	}
	if !haveGlyph {
		return false
	}
	if !haveBox {
		bbox.xMin, bbox.yMin, bbox.xMax, bbox.yMax = 0, 0, 0, 0
	}

	advance := float64(pen.x) / 64.0
	return fn(nativeFreetypeRun{
		glyphs:  glyphs,
		bbox:    bbox,
		advance: advance,
	})
}

func nativeFreetypeRunMask(run nativeFreetypeRun) (*image.Alpha, bool) {
	maskWidth := int((run.bbox.xMax-run.bbox.xMin)/64) + 2
	maskHeight := int((run.bbox.yMax-run.bbox.yMin)/64) + 2
	if maskWidth <= 0 || maskHeight <= 0 {
		return nil, false
	}
	mask := image.NewAlpha(image.Rect(0, 0, maskWidth, maskHeight))

	for i := range run.glyphs {
		if C.FT_Glyph_To_Bitmap(&run.glyphs[i], C.FT_RENDER_MODE_NORMAL, nil, 1) != 0 {
			return nil, false
		}
		bitmap := C.mpl_go_run_bitmap_glyph_bitmap(run.glyphs[i])
		glyphMask, ok := freetypeBitmapMask(*bitmap)
		if !ok {
			continue
		}

		x := int(float64(C.mpl_go_run_bitmap_glyph_left(run.glyphs[i])) - float64(run.bbox.xMin)/64.0)
		y := int(float64(run.bbox.yMax)/64.0) - int(C.mpl_go_run_bitmap_glyph_top(run.glyphs[i])) + 1
		orAlphaMask(mask, glyphMask, x, y)
	}
	return mask, true
}

func freeNativeGlyphs(glyphs []C.FT_Glyph) {
	for _, glyph := range glyphs {
		if glyph != nil {
			C.FT_Done_Glyph(glyph)
		}
	}
}

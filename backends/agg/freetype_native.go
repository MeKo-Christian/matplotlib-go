//go:build cgo && !purego

package agg

/*
// Default build: link the system FreeType via pkg-config.
#cgo !freetype261 pkg-config: freetype2
// Parity build (-tags freetype261): link the vendored static FreeType 2.6.1
// (matplotlib's pinned version) built by third_party/freetype/build.sh. Using
// tag-conditional cgo flags (not PKG_CONFIG_PATH) makes the FreeType version a
// function of the build tag, so go's build cache never serves a system-FreeType
// agg object to a parity build or vice versa. ${SRCDIR} keeps it relocatable;
// only libfreetype.a lives in the prefix, so -lfreetype links statically.
#cgo freetype261 CFLAGS: -I${SRCDIR}/../../third_party/freetype/prefix/include/freetype2
#cgo freetype261 LDFLAGS: -L${SRCDIR}/../../third_party/freetype/prefix/lib -lfreetype -lm
#include <stdlib.h>
#include <ft2build.h>
#include FT_FREETYPE_H
#include FT_GLYPH_H

static FT_Int32 mpl_go_force_autohint_load_flags(void) {
	return FT_LOAD_DEFAULT | FT_LOAD_FORCE_AUTOHINT;
}

static int mpl_go_pixel_mode_gray(void) {
	return FT_PIXEL_MODE_GRAY;
}

static int mpl_go_pixel_mode_mono(void) {
	return FT_PIXEL_MODE_MONO;
}

static int mpl_go_has_kerning(FT_Face face) {
	return FT_HAS_KERNING(face);
}

static void mpl_go_freetype_version(FT_Library library, FT_Int *major, FT_Int *minor, FT_Int *patch) {
	FT_Library_Version(library, major, minor, patch);
}

static FT_Bitmap *mpl_go_bitmap_glyph_bitmap(FT_Glyph glyph) {
	return &((FT_BitmapGlyph)glyph)->bitmap;
}

static FT_Int mpl_go_bitmap_glyph_left(FT_Glyph glyph) {
	return ((FT_BitmapGlyph)glyph)->left;
}

static FT_Int mpl_go_bitmap_glyph_top(FT_Glyph glyph) {
	return ((FT_BitmapGlyph)glyph)->top;
}
*/
import "C"

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"unsafe"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type nativeFreetypeRun struct {
	glyphs  []C.FT_Glyph
	bbox    C.FT_BBox
	advance float64
}

func (r *Renderer) drawNativeFreetypeText(text string, face render.FontFace, origin geom.Pt, size float64, textColor render.Color) bool {
	if r.ctx == nil || text == "" || face.Path == "" || size <= 0 {
		return false
	}

	var library C.FT_Library
	if C.FT_Init_FreeType(&library) != 0 {
		return false
	}
	defer C.FT_Done_FreeType(library)

	path := C.CString(face.Path)
	defer C.free(unsafe.Pointer(path))

	var ftFace C.FT_Face
	if C.FT_New_Face(library, path, 0, &ftFace) != 0 {
		return false
	}
	defer C.FT_Done_Face(ftFace)

	dpi := r.resolution
	if dpi == 0 {
		dpi = 72
	}
	charSize := C.FT_F26Dot6(math.Round(size * 64.0))
	if C.FT_Set_Char_Size(ftFace, 0, charSize, C.FT_UInt(dpi), C.FT_UInt(dpi)) != 0 {
		return false
	}

	var matrix C.FT_Matrix
	matrix.xx = 0x10000
	matrix.yy = 0x10000

	penX := C.FT_Pos(0)
	baselineY := float64(r.height) - origin.Y
	loadFlags := C.mpl_go_force_autohint_load_flags()
	paint := renderColorToRGBA(textColor)
	uniform := image.NewUniform(paint)
	var previousGlyph C.FT_UInt
	drewGlyph := false

	for _, rr := range text {
		glyphIndex := C.FT_Get_Char_Index(ftFace, C.FT_ULong(rr))
		if glyphIndex == 0 {
			return false
		}
		if previousGlyph != 0 && C.mpl_go_has_kerning(ftFace) != 0 {
			var kerning C.FT_Vector
			if C.FT_Get_Kerning(ftFace, previousGlyph, glyphIndex, C.FT_KERNING_DEFAULT, &kerning) == 0 {
				penX += kerning.x
			}
		}
		delta := C.FT_Vector{
			x: C.FT_Pos(math.Round((origin.X + float64(penX)/64.0) * 64.0)),
			y: C.FT_Pos(math.Round(baselineY * 64.0)),
		}
		C.FT_Set_Transform(ftFace, &matrix, &delta)
		if C.FT_Load_Glyph(ftFace, glyphIndex, loadFlags) != 0 {
			return false
		}

		slot := ftFace.glyph
		if C.FT_Render_Glyph(slot, C.FT_RENDER_MODE_NORMAL) != 0 {
			return false
		}
		bitmap := slot.bitmap
		width := int(bitmap.width)
		height := int(bitmap.rows)
		if width > 0 && height > 0 && bitmap.buffer != nil {
			mask, ok := freetypeBitmapMask(bitmap)
			if !ok {
				return false
			}
			src := image.NewRGBA(mask.Bounds())
			draw.DrawMask(src, src.Bounds(), uniform, image.Point{}, mask, image.Point{}, draw.Over)
			img, err := agglib.NewImageFromStandardImage(src)
			if err != nil {
				return false
			}
			x := float64(slot.bitmap_left)
			y := float64(r.height - int(slot.bitmap_top))
			if err := r.ctx.DrawImageScaled(img, x, y, float64(width), float64(height)); err != nil {
				return false
			}
			drewGlyph = true
		}

		penX += slot.advance.x
		previousGlyph = glyphIndex
	}

	return drewGlyph
}

func (r *Renderer) drawNativeFreetypeRunText(text string, face render.FontFace, origin geom.Pt, size float64, textColor render.Color, hintingFactor int) bool {
	if r.ctx == nil || text == "" || face.Path == "" || size <= 0 {
		return false
	}

	dpi := r.resolution
	if dpi == 0 {
		dpi = 72
	}
	return withNativeFreetypeRun(face.Path, text, size, dpi, hintingFactor, func(run nativeFreetypeRun) bool {
		maskWidth := int((run.bbox.xMax-run.bbox.xMin)/64) + 2
		maskHeight := int((run.bbox.yMax-run.bbox.yMin)/64) + 2
		if maskWidth <= 0 || maskHeight <= 0 {
			return false
		}
		mask := image.NewAlpha(image.Rect(0, 0, maskWidth, maskHeight))

		for i := range run.glyphs {
			if C.FT_Glyph_To_Bitmap(&run.glyphs[i], C.FT_RENDER_MODE_NORMAL, nil, 1) != 0 {
				return false
			}
			bitmap := C.mpl_go_bitmap_glyph_bitmap(run.glyphs[i])
			glyphMask, ok := freetypeBitmapMask(*bitmap)
			if !ok {
				continue
			}

			x := int(float64(C.mpl_go_bitmap_glyph_left(run.glyphs[i])) - float64(run.bbox.xMin)/64.0)
			y := int(float64(run.bbox.yMax)/64.0) - int(C.mpl_go_bitmap_glyph_top(run.glyphs[i])) + 1
			orAlphaMask(mask, glyphMask, x, y)
		}

		descent := -float64(run.bbox.yMin) / 64.0
		dstX := math.Round(origin.X + float64(run.bbox.xMin)/64.0)
		bottomY := math.Round(origin.Y+descent) + 1
		dstY := bottomY - float64(maskHeight)
		return r.blendAlphaMask(mask, int(dstX), int(dstY), textColor)
	})
}

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
	return bounds, metrics, ok
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

	loadFlags := C.mpl_go_force_autohint_load_flags()
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
		if previousGlyph != 0 && C.mpl_go_has_kerning(ftFace) != 0 {
			var kerning C.FT_Vector
			if C.FT_Get_Kerning(ftFace, previousGlyph, glyphIndex, C.FT_KERNING_DEFAULT, &kerning) == 0 {
				pen.x += kerning.x
				pen.y += kerning.y
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
		C.FT_Glyph_Transform(glyph, &matrix, nil)

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

	advance := float64(pen.x) * float64(matrix.xx) / 65536.0 / 64.0
	return fn(nativeFreetypeRun{
		glyphs:  glyphs,
		bbox:    bbox,
		advance: advance,
	})
}

func freeNativeGlyphs(glyphs []C.FT_Glyph) {
	for _, glyph := range glyphs {
		if glyph != nil {
			C.FT_Done_Glyph(glyph)
		}
	}
}

func orAlphaMask(dst, src *image.Alpha, x, y int) {
	for sy := src.Bounds().Min.Y; sy < src.Bounds().Max.Y; sy++ {
		dy := y + sy
		if dy < dst.Bounds().Min.Y || dy >= dst.Bounds().Max.Y {
			continue
		}
		for sx := src.Bounds().Min.X; sx < src.Bounds().Max.X; sx++ {
			dx := x + sx
			if dx < dst.Bounds().Min.X || dx >= dst.Bounds().Max.X {
				continue
			}
			dstOff := dst.PixOffset(dx, dy)
			srcOff := src.PixOffset(sx, sy)
			dst.Pix[dstOff] |= src.Pix[srcOff]
		}
	}
}

func (r *Renderer) blendAlphaMask(mask *image.Alpha, dstX, dstY int, textColor render.Color) bool {
	if r.ctx == nil || r.ctx.image == nil || mask == nil {
		return false
	}

	paint := renderColorToRGBA(textColor)
	data := r.ctx.image.Data
	stride := r.ctx.image.Stride()
	width, height := r.ctx.image.Width(), r.ctx.image.Height()
	if stride <= 0 || len(data) == 0 {
		return false
	}

	drew := false
	for my := mask.Bounds().Min.Y; my < mask.Bounds().Max.Y; my++ {
		y := dstY + my
		if y < 0 || y >= height {
			continue
		}
		for mx := mask.Bounds().Min.X; mx < mask.Bounds().Max.X; mx++ {
			x := dstX + mx
			if x < 0 || x >= width {
				continue
			}
			cover := uint32(mask.Pix[mask.PixOffset(mx, my)])
			if cover == 0 {
				continue
			}
			srcA := uint32(paint.A) * cover / 255
			if srcA == 0 {
				continue
			}
			off := y*stride + x*4
			invA := 255 - srcA
			data[off+0] = uint8((uint32(paint.R)*srcA + uint32(data[off+0])*invA + 127) / 255)
			data[off+1] = uint8((uint32(paint.G)*srcA + uint32(data[off+1])*invA + 127) / 255)
			data[off+2] = uint8((uint32(paint.B)*srcA + uint32(data[off+2])*invA + 127) / 255)
			data[off+3] = uint8(srcA + uint32(data[off+3])*invA/255)
			drew = true
		}
	}
	return drew
}

func nativeFreetypeVersion() string {
	var library C.FT_Library
	if C.FT_Init_FreeType(&library) != 0 {
		return ""
	}
	defer C.FT_Done_FreeType(library)

	var major, minor, patch C.FT_Int
	C.mpl_go_freetype_version(library, &major, &minor, &patch)
	return fmt.Sprintf("%d.%d.%d", int(major), int(minor), int(patch))
}

func freetypeBitmapMask(bitmap C.FT_Bitmap) (*image.Alpha, bool) {
	width := int(bitmap.width)
	height := int(bitmap.rows)
	pitch := int(bitmap.pitch)
	if width <= 0 || height <= 0 || pitch == 0 || bitmap.buffer == nil {
		return nil, false
	}

	mask := image.NewAlpha(image.Rect(0, 0, width, height))
	bufferLen := absInt(pitch) * height
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bitmap.buffer)), bufferLen)
	pixelMode := int(bitmap.pixel_mode)

	for row := 0; row < height; row++ {
		srcRow := row * pitch
		if pitch < 0 {
			srcRow = (height - 1 - row) * -pitch
		}
		dstRow := row * mask.Stride
		switch pixelMode {
		case int(C.mpl_go_pixel_mode_gray()):
			copy(mask.Pix[dstRow:dstRow+width], buffer[srcRow:srcRow+width])
		case int(C.mpl_go_pixel_mode_mono()):
			for col := 0; col < width; col++ {
				byteIndex := srcRow + col/8
				bit := byte(1 << uint(7-col%8))
				if buffer[byteIndex]&bit != 0 {
					mask.Pix[dstRow+col] = 0xff
				}
			}
		default:
			return nil, false
		}
	}
	return mask, true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

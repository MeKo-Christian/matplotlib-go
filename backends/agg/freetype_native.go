//go:build cgo && !purego

package agg

/*
// Default build: statically link the vendored FreeType 2.6.1 — matplotlib's
// pinned version, the one used to generate every reference image. It is built
// by `just freetype261-build` (third_party/freetype/build.sh) into the
// gitignored prefix. Pinning the FreeType version is what makes the AGG text
// rasterization byte-match the matplotlib references (the autohinter changed
// between 2.6.1 and current system FreeType, ~20 RMSE on dense text). ${SRCDIR}
// keeps the paths relocatable; only libfreetype.a ships in the prefix, so
// -lfreetype links statically.
//
// Compile fallback (-tags systemfreetype): link the system FreeType via
// pkg-config for environments without the vendored prefix (IDEs, quick vet).
// This is NOT parity-exact — golden/reference tests are expected to diverge —
// and exists only so the cgo packages compile without building FreeType 2.6.1.
#cgo !systemfreetype CFLAGS: -I${SRCDIR}/../../third_party/freetype/prefix/include/freetype2
#cgo !systemfreetype LDFLAGS: -L${SRCDIR}/../../third_party/freetype/prefix/lib -lfreetype -lm
#cgo systemfreetype pkg-config: freetype2
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
	loadFlags := C.mpl_go_force_autohint_load_flags()
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
		if previousGlyph != 0 && C.mpl_go_has_kerning(ftFace) != 0 {
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

// DrawMathTextImage rasterizes a flattened mathtext expression with matplotlib's
// _mathtext.Output.to_raster pixel placement: one shared bounding box, per-glyph
// integer blitting (int(ox)+bitmap_left, int(oy-iceberg)) and the draw_rect_filled
// rule formula, then the whole image anchored at round(.)+1. Coordinates from core
// are layout space (y-down, baseline = 0; Oy/rect.Y negative above baseline).
func (r *Renderer) DrawMathTextImage(glyphs []render.MathGlyphPlacement, rects []render.MathRectPlacement, anchor geom.Pt, textColor render.Color) bool {
	if r.ctx == nil || len(glyphs)+len(rects) == 0 {
		return false
	}
	hf := matplotlibTextHintingFactor

	type placed struct {
		g    nativeMathGlyph
		ox   float64
		oy   float64
		have bool
	}
	rendered := make([]placed, len(glyphs))

	// Bounding box over glyph ink, rects, and the origin (0), with the -1/+1
	// border, in layout y-down space. Glyph ink top = oy - ymax, bottom = oy - ymin.
	xmin, ymin := 0.0, 0.0
	xmax, ymax := 0.0, 0.0
	for i, gp := range glyphs {
		runes := []rune(gp.Text)
		if len(runes) != 1 {
			continue
		}
		var fontPath string
		r.withTemporaryFontKey(func() {
			font := r.configureTextFont(gp.FontSize, gp.FontKey)
			if font.backend == textBackendRaster {
				fontPath = font.face.Path
			}
		})
		if fontPath == "" {
			return false
		}
		g, ok := r.nativeFreetypeMathGlyph(runes[0], fontPath, gp.FontSize, hf)
		if !ok {
			return false
		}
		rendered[i] = placed{g: g, ox: gp.Ox, oy: gp.Oy, have: true}
		xmin = math.Min(xmin, gp.Ox+g.xmin)
		xmax = math.Max(xmax, gp.Ox+g.xmax)
		ymin = math.Min(ymin, gp.Oy-g.ymax)
		ymax = math.Max(ymax, gp.Oy-g.ymin)
	}
	for _, rc := range rects {
		xmin = math.Min(xmin, rc.X1)
		xmax = math.Max(xmax, rc.X2)
		ymin = math.Min(ymin, rc.Y1)
		ymax = math.Max(ymax, rc.Y2)
	}
	xmin -= 1
	ymin -= 1
	xmax += 1
	ymax += 1
	_ = xmax
	_ = ymax

	// Device placement of the math image. origin (display) X is the expression
	// left; its device baseline is height-origin.Y. matplotlib places the image
	// at round(text_x), round(.)+1; the -xmin/-ymin borders are folded into the
	// per-element int() below.
	baselineDev := float64(r.height) - anchor.Y
	imageLeftDev := math.Round(anchor.X)
	imageTopDev := math.Round(baselineDev+ymin) + 1

	for i := range rendered {
		p := &rendered[i]
		if !p.have || p.g.mask == nil {
			continue
		}
		gx := int(imageLeftDev) + int(p.ox-xmin) + p.g.bitmapLeft
		gy := int(imageTopDev) + int((p.oy-ymin)-p.g.iceberg)
		r.blendAlphaMask(p.g.mask, gx, gy, textColor)
	}

	for _, rc := range rects {
		y1 := rc.Y1 - ymin
		y2 := rc.Y2 - ymin
		height := int(y2-y1) - 1
		if height < 0 {
			height = 0
		}
		var yy int
		if height == 0 {
			yy = int((y1+y2)/2 - 0.5)
		} else {
			yy = int(y1)
		}
		x0 := int(imageLeftDev) + int(rc.X1-xmin)
		x1 := int(imageLeftDev) + int(math.Ceil(rc.X2-xmin))
		top := int(imageTopDev) + yy
		r.fillDeviceRect(x0, top, x1+1, top+height+1, textColor)
	}
	return true
}

// fillDeviceRect blends a solid color over the half-open device rect
// [x0,x1) x [y0,y1) (matplotlib FT2Image.draw_rect_filled is inclusive; callers
// pass the +1-adjusted half-open bounds).
func (r *Renderer) fillDeviceRect(x0, y0, x1, y1 int, textColor render.Color) {
	if x1 <= x0 || y1 <= y0 {
		return
	}
	mask := image.NewAlpha(image.Rect(0, 0, x1-x0, y1-y0))
	for i := range mask.Pix {
		mask.Pix[i] = 0xff
	}
	r.blendAlphaMask(mask, x0, y0, textColor)
}

// nativeMathGlyph bundles one mathtext glyph's matplotlib `_get_info` metrics
// (device pixels, baseline-relative, y-up) with its rendered hinted bitmap, so
// DrawMathTextImage can position it exactly like _mathtext.Output.to_raster.
type nativeMathGlyph struct {
	iceberg               float64
	xmin, xmax            float64
	ymin, ymax            float64
	mask                  *image.Alpha
	bitmapLeft, bitmapTop int
}

// nativeFreetypeMathGlyph loads, measures and renders a single glyph with the
// matplotlib FT2Font setup (dpi*hf horizontal grid + xx=1/hf transform,
// FORCE_AUTOHINT) and returns its _get_info metrics plus the hinted bitmap. The
// CBox is taken before FT_Render_Glyph (from the transformed outline, 1x).
func (r *Renderer) nativeFreetypeMathGlyph(rr rune, fontPath string, size float64, hintingFactor int) (nativeMathGlyph, bool) {
	if fontPath == "" || size <= 0 {
		return nativeMathGlyph{}, false
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
		return nativeMathGlyph{}, false
	}
	defer C.FT_Done_FreeType(library)

	path := C.CString(fontPath)
	defer C.free(unsafe.Pointer(path))

	var ftFace C.FT_Face
	if C.FT_New_Face(library, path, 0, &ftFace) != 0 {
		return nativeMathGlyph{}, false
	}
	defer C.FT_Done_Face(ftFace)

	charSize := C.FT_F26Dot6(math.Round(size * 64.0))
	if C.FT_Set_Char_Size(ftFace, charSize, 0, C.FT_UInt(dpi*uint(hintingFactor)), C.FT_UInt(dpi)) != 0 {
		return nativeMathGlyph{}, false
	}
	matrix := C.FT_Matrix{
		xx: C.FT_Fixed(math.Round(65536.0 / float64(hintingFactor))),
		yy: 0x10000,
	}
	C.FT_Set_Transform(ftFace, &matrix, nil)

	glyphIndex := C.FT_Get_Char_Index(ftFace, C.FT_ULong(rr))
	if glyphIndex == 0 {
		return nativeMathGlyph{}, false
	}
	if C.FT_Load_Glyph(ftFace, glyphIndex, C.mpl_go_force_autohint_load_flags()) != 0 {
		return nativeMathGlyph{}, false
	}
	slot := ftFace.glyph

	var glyph C.FT_Glyph
	if C.FT_Get_Glyph(slot, &glyph) != 0 {
		return nativeMathGlyph{}, false
	}
	var cbox C.FT_BBox
	C.FT_Glyph_Get_CBox(glyph, C.FT_GLYPH_BBOX_SUBPIXELS, &cbox)
	C.FT_Done_Glyph(glyph)

	out := nativeMathGlyph{
		iceberg: float64(slot.metrics.horiBearingY) / 64.0,
		xmin:    float64(cbox.xMin) / 64.0,
		xmax:    float64(cbox.xMax) / 64.0,
		ymin:    float64(cbox.yMin) / 64.0,
		ymax:    float64(cbox.yMax) / 64.0,
	}

	if C.FT_Render_Glyph(slot, C.FT_RENDER_MODE_NORMAL) == 0 {
		if mask, ok := freetypeBitmapMask(slot.bitmap); ok {
			out.mask = mask
			out.bitmapLeft = int(slot.bitmap_left)
			out.bitmapTop = int(slot.bitmap_top)
		}
	}
	return out, true
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

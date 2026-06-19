//go:build cgo && !purego

package agg

/*
#cgo !systemfreetype CFLAGS: -I${SRCDIR}/../../third_party/freetype/prefix/include/freetype2
#cgo !systemfreetype LDFLAGS: -L${SRCDIR}/../../third_party/freetype/prefix/lib -lfreetype -lm
#cgo systemfreetype pkg-config: freetype2
#include <stdlib.h>
#include <ft2build.h>
#include FT_FREETYPE_H

static FT_Int32 mpl_go_text_force_autohint_load_flags(void) {
	return FT_LOAD_DEFAULT | FT_LOAD_FORCE_AUTOHINT;
}

static int mpl_go_text_has_kerning(FT_Face face) {
	return FT_HAS_KERNING(face);
}
*/
import "C"

import (
	"image"
	"image/draw"
	"math"
	"unsafe"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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
	loadFlags := C.mpl_go_text_force_autohint_load_flags()
	paint := renderColorToRGBA(textColor)
	uniform := image.NewUniform(paint)
	var previousGlyph C.FT_UInt
	drewGlyph := false

	for _, rr := range text {
		glyphIndex := C.FT_Get_Char_Index(ftFace, C.FT_ULong(rr))
		if glyphIndex == 0 {
			return false
		}
		if previousGlyph != 0 && C.mpl_go_text_has_kerning(ftFace) != 0 {
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

// Matplotlib's Python AGG backend calls Python round() before draw_text_image,
// which is ties-to-even, not Go's half-away-from-zero math.Round.
func pythonRound(v float64) float64 {
	half := math.Floor(v) + 0.5
	if math.Abs(v-half) < 1e-9 {
		v = half
	}
	return math.RoundToEven(v)
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
		mask, ok := nativeFreetypeRunMask(run)
		if !ok {
			return false
		}
		maskHeight := mask.Bounds().Dy()
		descent := -float64(run.bbox.yMin) / 64.0
		dstX := pythonRound(origin.X + float64(run.bbox.xMin)/64.0)
		bottomY := pythonRound(origin.Y+descent) + 1
		dstY := bottomY - float64(maskHeight)
		return r.blendAlphaMask(mask, int(dstX), int(dstY), textColor)
	})
}

func (r *Renderer) drawNativeFreetypeRunTextRotated(text string, face render.FontFace, origin geom.Pt, size, angle float64, textColor render.Color, hintingFactor int) bool {
	if r.ctx == nil || text == "" || face.Path == "" || size <= 0 {
		return false
	}

	dpi := r.resolution
	if dpi == 0 {
		dpi = 72
	}
	return withNativeFreetypeRun(face.Path, text, size, dpi, hintingFactor, func(run nativeFreetypeRun) bool {
		mask, ok := nativeFreetypeRunMask(run)
		if !ok {
			return false
		}
		src := image.NewRGBA(mask.Bounds())
		draw.DrawMask(src, src.Bounds(), image.NewUniform(renderColorToRGBA(textColor)), image.Point{}, mask, image.Point{}, draw.Over)
		img, err := agglib.NewImageFromStandardImage(src)
		if err != nil {
			return false
		}

		origin = r.devPt(origin)

		descent := -float64(run.bbox.yMin) / 64.0
		sinT := math.Sin(angle)
		cosT := math.Cos(angle)
		x := pythonRound(origin.X + float64(run.bbox.xMin)/64.0 + descent*sinT)
		y := pythonRound(origin.Y+descent*cosT) + 1
		height := float64(mask.Bounds().Dy())
		transform := agglib.NewTransformationsFromValues(
			cosT,
			-sinT,
			sinT,
			cosT,
			x-height*sinT,
			y-height*cosT,
		)

		prevBlendMode := r.ctx.GetBlendMode()
		prevFilter := r.ctx.GetImageFilter()
		prevResample := r.ctx.GetImageResample()
		defer func() {
			r.ctx.SetBlendMode(prevBlendMode)
			r.ctx.SetImageFilter(prevFilter)
			r.ctx.SetImageResample(prevResample)
		}()
		r.ctx.SetBlendMode(agglib.BlendSrcOver)
		r.ctx.SetImageFilter(agglib.Spline36)
		r.ctx.SetImageResample(resampleForFilter(agglib.Spline36))

		err = r.ctx.DrawImageTransformed(img, transform)
		return err == nil
	})
}

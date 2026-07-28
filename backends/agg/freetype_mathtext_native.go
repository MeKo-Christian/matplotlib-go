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

static FT_Int32 mpl_go_math_force_autohint_load_flags(void) {
	return FT_LOAD_DEFAULT | FT_LOAD_FORCE_AUTOHINT;
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

// mathShipOffsetV is matplotlib's `off_v = oy + box.height` from _mathtext.ship,
// where oy is the raster shift (-ymin) and box.height the expression ascent.
//
// ship adds this to each glyph's accumulated cur_v as a single term, and the
// association is not free: `cur_v + (boxAscent - ymin)` and
// `(boxAscent + cur_v) - ymin` differ in the last ULP, and to_raster blits with
// int(), which truncates rather than rounds — so one ULP becomes one pixel
// whenever the sum lands on an integer. \dot{x} at 26 pt is the case that
// showed it: matplotlib reaches int(0.99999999999999645) = 0 for the dot where
// the port reached exactly 1.0 and drew it a pixel low.
func mathShipOffsetV(boxAscent, ymin float64) float64 {
	return boxAscent - ymin
}

// DrawMathTextImage rasterizes a flattened mathtext expression with matplotlib's
// _mathtext.Output.to_raster pixel placement: one shared bounding box, per-glyph
// integer blitting (int(ox)+bitmap_left, int(oy-iceberg)) and the draw_rect_filled
// rule formula, then the whole image anchored at round(.)+1. Coordinates from core
// are layout space (y-down, baseline = 0; Oy/rect.Y negative above baseline).
func (r *Renderer) DrawMathTextImage(glyphs []render.MathGlyphPlacement, rects []render.MathRectPlacement, anchor geom.Pt, boxAscent, boxDescent float64, textColor render.Color) bool {
	if r.ctx == nil || len(glyphs)+len(rects) == 0 {
		return false
	}
	hf := matplotlibTextHintingFactor

	rendered := make([]placedNativeMathGlyph, len(glyphs))

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
		rendered[i] = placedNativeMathGlyph{g: g, ox: gp.Ox, oy: gp.Oy, have: true}
		shipY := boxAscent + gp.Oy
		xmin = math.Min(xmin, gp.Ox+g.xmin)
		xmax = math.Max(xmax, gp.Ox+g.xmax)
		ymin = math.Min(ymin, shipY-g.ymax)
		ymax = math.Max(ymax, shipY-g.ymin)
	}
	for _, rc := range rects {
		xmin = math.Min(xmin, rc.X1)
		xmax = math.Max(xmax, rc.X2)
		ymin = math.Min(ymin, boxAscent+rc.Y1)
		ymax = math.Max(ymax, boxAscent+rc.Y2)
	}
	xmin -= 1
	ymin -= 1
	xmax += 1
	ymax += 1
	_ = xmax
	_ = ymax

	// Device placement of the math image, matching matplotlib RendererAgg.draw_mathtext
	// + draw_text_image. The renderer receives the device baseline (height-anchor.Y)
	// and blits the image with its BOTTOM-LEFT at (round(x), round(y+descent)+1), so
	// the top is round(B+descent)+1-imageHeight. descent and imageHeight come from
	// Output.to_raster: d = (ymax-ymin)-box.height, h = (ymax-ymin)-box.depth,
	// imageHeight = ceil(h+max(d,0)). The -xmin border is already folded into the
	// per-glyph int(Ox-xmin) below; do NOT add it to the image left.
	baselineDev := float64(r.height) - anchor.Y
	totalH := ymax - ymin
	parseDescent := totalH - boxAscent
	parseH := totalH - boxDescent
	imageHeight := math.Ceil(parseH + math.Max(parseDescent, 0))
	imageLeftDev := pythonRound(anchor.X)
	imageTopDev := pythonRound(baselineDev+parseDescent) + 1 - imageHeight
	offV := mathShipOffsetV(boxAscent, ymin)

	for i := range rendered {
		p := &rendered[i]
		if !p.have || p.g.mask == nil {
			continue
		}
		gx := int(imageLeftDev) + int(p.ox-xmin) + p.g.bitmapLeft
		gy := int(imageTopDev) + int((p.oy+offV)-p.g.iceberg)
		r.blendAlphaMask(p.g.mask, gx, gy, textColor)
	}

	for _, rc := range rects {
		y1 := rc.Y1 + offV
		y2 := rc.Y2 + offV
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

// DrawMathTextImageRotated follows matplotlib's RendererAgg.draw_mathtext:
// rasterize the complete mathtext output first, round the baseline/descent
// adjusted bottom-left, then rotate the image through draw_text_image semantics.
func (r *Renderer) DrawMathTextImageRotated(glyphs []render.MathGlyphPlacement, rects []render.MathRectPlacement, origin geom.Pt, boxAscent, boxDescent, angle float64, textColor render.Color) bool {
	if r.ctx == nil || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return false
	}
	img, ok := r.mathTextAlphaImage(glyphs, rects, boxAscent, boxDescent)
	if !ok {
		return false
	}
	src := image.NewRGBA(img.mask.Bounds())
	draw.DrawMask(src, src.Bounds(), image.NewUniform(renderColorToRGBA(textColor)), image.Point{}, img.mask, image.Point{}, draw.Over)
	aggImg, err := agglib.NewImageFromStandardImage(src)
	if err != nil {
		return false
	}

	baselineDev := float64(r.height) - origin.Y
	sinT := math.Sin(angle)
	cosT := math.Cos(angle)
	imageHeight := float64(img.mask.Bounds().Dy())
	// RendererAgg.draw_mathtext: round(x + ox + descent*sin), round(y - oy +
	// descent*cos), then draw_text_image at y+1. draw_text_image's rotated branch
	// maps the image through translate(0,-h) · rotate(-angle) · translate(x,y),
	// which is exactly the transform below.
	x := pythonRound(origin.X + img.descent*sinT)
	y := pythonRound(baselineDev+img.descent*cosT) + 1
	transform := agglib.NewTransformationsFromValues(
		cosT,
		-sinT,
		sinT,
		cosT,
		x-imageHeight*sinT,
		y-imageHeight*cosT,
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

	return r.ctx.DrawImageTransformed(aggImg, transform) == nil
}

type mathTextAlphaImage struct {
	mask    *image.Alpha
	descent float64
}

type placedNativeMathGlyph struct {
	g    nativeMathGlyph
	ox   float64
	oy   float64
	have bool
}

func (r *Renderer) mathTextAlphaImage(glyphs []render.MathGlyphPlacement, rects []render.MathRectPlacement, boxAscent, boxDescent float64) (mathTextAlphaImage, bool) {
	if len(glyphs)+len(rects) == 0 {
		return mathTextAlphaImage{}, false
	}

	rendered := make([]placedNativeMathGlyph, len(glyphs))

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
			return mathTextAlphaImage{}, false
		}
		g, ok := r.nativeFreetypeMathGlyph(runes[0], fontPath, gp.FontSize, matplotlibTextHintingFactor)
		if !ok {
			return mathTextAlphaImage{}, false
		}
		rendered[i] = placedNativeMathGlyph{g: g, ox: gp.Ox, oy: gp.Oy, have: true}
		shipY := boxAscent + gp.Oy
		xmin = math.Min(xmin, gp.Ox+g.xmin)
		xmax = math.Max(xmax, gp.Ox+g.xmax)
		ymin = math.Min(ymin, shipY-g.ymax)
		ymax = math.Max(ymax, shipY-g.ymin)
	}
	for _, rc := range rects {
		xmin = math.Min(xmin, rc.X1)
		xmax = math.Max(xmax, rc.X2)
		ymin = math.Min(ymin, boxAscent+rc.Y1)
		ymax = math.Max(ymax, boxAscent+rc.Y2)
	}
	xmin -= 1
	ymin -= 1
	xmax += 1
	ymax += 1

	totalH := ymax - ymin
	parseDescent := totalH - boxAscent
	parseH := totalH - boxDescent
	imageHeight := int(math.Ceil(parseH + math.Max(parseDescent, 0)))
	imageWidth := int(math.Ceil(xmax - xmin))
	if imageWidth <= 0 || imageHeight <= 0 {
		return mathTextAlphaImage{}, false
	}
	mask := image.NewAlpha(image.Rect(0, 0, imageWidth, imageHeight))
	offV := mathShipOffsetV(boxAscent, ymin)

	for i := range rendered {
		p := &rendered[i]
		if !p.have || p.g.mask == nil {
			continue
		}
		gx := int(p.ox-xmin) + p.g.bitmapLeft
		gy := int((p.oy + offV) - p.g.iceberg)
		overAlphaMask(mask, p.g.mask, gx, gy)
	}

	for _, rc := range rects {
		y1 := rc.Y1 + offV
		y2 := rc.Y2 + offV
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
		x0 := int(rc.X1 - xmin)
		x1 := int(math.Ceil(rc.X2 - xmin))
		fillAlphaRect(mask, x0, yy, x1+1, yy+height+1)
	}

	return mathTextAlphaImage{mask: mask, descent: parseDescent}, true
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
	if C.FT_Load_Glyph(ftFace, glyphIndex, C.mpl_go_math_force_autohint_load_flags()) != 0 {
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

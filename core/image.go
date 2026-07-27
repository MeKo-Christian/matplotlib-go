package core

import (
	"image"
	"image/color"
	"math"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/render"
)

// rotatedImageFallbackWarnOnce guards a single diagnostic for the case where a
// renderer can rotate neither via rasterizeTransformed nor ImageTransformer.
var rotatedImageFallbackWarnOnce sync.Once

// Draw renders the rasterized image through the renderer.
func (i *Image2D) Draw(r render.Renderer, ctx *DrawContext) {
	if i == nil || r == nil {
		return
	}

	dst := i.destinationRect(ctx)
	if dst.W() <= 0 || dst.H() <= 0 {
		return
	}
	raster, drawDst, ok := i.rasterizeForDestination(ctx, dst)
	if !ok {
		return
	}
	angleRad := i.AngleDeg * math.Pi / 180
	if angleRad == 0 {
		r.DrawImage(raster, drawDst)
		return
	}

	anchor := i.rotationAnchor(ctx, dst)
	if transformed, transformedDst, ok := i.rasterizeTransformed(ctx, anchor, angleRad); ok {
		r.DrawImage(transformed, transformedDst)
		return
	}

	if tr, ok := r.(render.ImageTransformer); ok {
		transform := imageTransform(drawDst, raster, anchor, angleRad)
		tr.ImageTransformed(raster, drawDst, transform)
		return
	}

	// Fallback: the renderer supports neither rasterizeTransformed nor
	// ImageTransformer, so the rotation cannot be honored. Warn once (rather
	// than silently mis-rendering) and draw the image axis-aligned. The AGG
	// backend implements both paths, so this never fires there.
	rotatedImageFallbackWarnOnce.Do(func() {
		diag.Warnf("image rotation (%.3g deg) not supported by renderer %T; drawing axis-aligned", i.AngleDeg, r)
	})
	r.DrawImage(raster, drawDst)
}

func matplotlibImageDrawRect(dst geom.Rect) geom.Rect {
	// Matplotlib's Image._make_image(..., round_to_pixel_border=True) rounds
	// the resampled image buffer up to whole output pixels and keeps the
	// original left/bottom anchor. Display space is y-up, so the bottom anchor is
	// dst.Min.Y: preserve left/bottom and extend right/up by the ceiled size.
	width := math.Ceil(math.Abs(dst.W()))
	height := math.Ceil(math.Abs(dst.H()))
	if width <= 0 || height <= 0 {
		return dst
	}
	return geom.Rect{
		Min: geom.Pt{
			X: dst.Min.X,
			Y: dst.Min.Y,
		},
		Max: geom.Pt{
			X: dst.Min.X + width,
			Y: dst.Min.Y + height,
		},
	}
}

func (i *Image2D) rasterize() (render.Image, bool) {
	if i != nil && i.rgba != nil {
		return i.rasterizeRGBA()
	}
	return i.rasterizeToSize(0, 0)
}

// rasterizeRGBA wraps pre-colored RGBA pixels for the renderer, applying the
// origin flip and per-image alpha. The backend handles any resampling to the
// destination rect, so true-color images skip the scalar data-stage prefilter.
func (i *Image2D) rasterizeRGBA() (render.Image, bool) {
	if i == nil || i.rgba == nil {
		return nil, false
	}
	b := i.rgba.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return nil, false
	}
	out := i.rgba
	if i.Origin == ImageOriginLower {
		// Bitmap row 0 is drawn at the top edge; origin 'lower' wants array row
		// 0 at the bottom, so flip vertically (mirrors the scalar pixelY flip).
		out = flipRGBAVertical(i.rgba)
	}
	data := render.NewImageData(out)
	data.SetInterpolation(i.Interpolation)
	data.SetAlpha(clampOneToOne(i.Alpha))
	return data, true
}

func flipRGBAVertical(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			dst.SetRGBA(x, h-1-y, src.RGBAAt(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

// rasterizeForDestination resamples the image and reports the pixel rect it is
// to be drawn into.
//
// Matplotlib resamples an image over its bounding box intersected with the clip
// box, not over the whole image (Image._make_image starts with
// Bbox.intersection(out_bbox, clip_bbox) and derives both the output buffer size
// and the affine from that intersection). For an image that fits inside its axes
// the two are the same rect and this is the plain destination. For one that
// overhangs, resampling the whole image and clipping the result afterwards
// samples a different set of source cells: the buffer-rounding compensation is
// computed from a different width, and the pixel grid starts at a different
// device coordinate, so pixel centres land elsewhere in source space.
func (i *Image2D) rasterizeForDestination(ctx *DrawContext, dst geom.Rect) (render.Image, geom.Rect, bool) {
	if i == nil || i.rgba != nil || i.AngleDeg != 0 {
		raster, ok := i.rasterize()
		return raster, matplotlibImageDrawRect(dst), ok
	}
	rows, cols := scalarImageDimensions(i.Data)
	if rows == 0 || cols == 0 {
		return nil, geom.Rect{}, false
	}

	clipped := dst
	if ctx != nil && ctx.Clip.W() > 0 && ctx.Clip.H() > 0 {
		clipped = dst.Intersect(ctx.Clip)
		if clipped.W() <= 0 || clipped.H() <= 0 {
			return nil, geom.Rect{}, false
		}
	}
	drawDst := matplotlibImageDrawRect(clipped)

	width := int(math.Round(math.Abs(drawDst.W())))
	height := int(math.Round(math.Abs(drawDst.H())))
	if width <= cols || height <= rows ||
		!shouldRasterizeScalarDataStage(i.Interpolation, width, height, cols, rows) {
		raster, ok := i.rasterize()
		return raster, matplotlibImageDrawRect(dst), ok
	}

	affine := i.matplotlibResampleAffine(dst, clipped, cols, rows, width, height)
	raster, ok := i.rasterizeToSizeAt(width, height, affine)
	return raster, drawDst, ok
}

// resampleScale is the affine Matplotlib hands its image resampler, reduced to
// one scale and one offset per axis. scale is device pixels per source cell;
// offset is the source-space coordinate the output buffer starts at, which is
// non-zero exactly when the clip box cuts into the image.
type resampleScale struct {
	x, y   float64
	ox, oy float64

	// columns is the source column for each destination column, precomputed
	// once per raster because AGG derives it by stepping an integer DDA along
	// the scanline rather than by mapping each pixel independently. A nil
	// slice means "map each column arithmetically", which is what the callers
	// that do not model the affine want.
	columns []int
}

// column is the source column for a destination column, from the precomputed
// span if there is one and by the plain per-pixel map otherwise.
func (s resampleScale) column(index, targetWidth, sourceCols int) int {
	if index >= 0 && index < len(s.columns) {
		return s.columns[index]
	}
	return scaledNearestIndexAt(index, targetWidth, sourceCols, s.x, s.ox)
}

// matplotlibResampleAffine reproduces that affine in the order Matplotlib
// composes it (Image._make_image).
//
// The naive ratio targetSize/sourceSize is the same number in exact arithmetic
// but not in float64, and the difference decides cell boundaries: the nearest
// rule in nearestSourceIndex snaps a coordinate that is within 1/512 of a cell
// edge onto the edge, so a last-place difference in the scale moves a whole
// source row or column. Matplotlib scales the exact device extent by the source
// count and then, only when the clipped extent is not already whole pixels,
// stretches it to fill the rounded-up buffer. Reproducing that order inherits
// the same rounding noise.
//
// base is the image's exact device rect and clipped that rect intersected with
// the clip box. The offsets are the clip's inset expressed in source cells.
//
// bufWidth and bufHeight are the output buffer's size in pixels. The height is
// needed only for origin "upper": Matplotlib flips that case in array space
// ahead of the affine, so its offset is measured from the opposite edge.
func (i *Image2D) matplotlibResampleAffine(base, clipped geom.Rect, cols, rows, bufWidth, bufHeight int) resampleScale {
	baseW := math.Abs(base.W())
	baseH := math.Abs(base.H())
	clipW := math.Abs(clipped.W())
	clipH := math.Abs(clipped.H())
	if baseW <= 0 || baseH <= 0 || clipW <= 0 || clipH <= 0 || cols <= 0 || rows <= 0 {
		return resampleScale{}
	}
	sx := baseW / float64(cols)
	sy := baseH / float64(rows)
	outW := math.Ceil(clipW)
	outH := math.Ceil(clipH)
	if clipW != outW || clipH != outH {
		sx *= 1 + (outW-clipW)/clipW
		sy *= 1 + (outH-clipH)/clipH
	}

	// Device space here is y-down, so the image's bottom edge is base.Max.Y.
	ox := (clipped.Min.X - base.Min.X) / (baseW / float64(cols))
	oy := (base.Max.Y - clipped.Max.Y) / (baseH / float64(rows))
	if i == nil || i.Origin != ImageOriginLower {
		oy = float64(rows) - float64(bufHeight)/sy - oy
	}
	return resampleScale{
		x: sx, y: sy, ox: ox, oy: oy,
		columns: aggSpanColumns(bufWidth, cols, sx, ox),
	}
}

// aggSpanColumns is the source column AGG's nearest-neighbour image filter
// reads for each destination column of one scanline.
//
// AGG does not map each pixel independently. span_interpolator_linear::begin
// transforms only the two ends of the span, rounds both to 1/256 of a source
// cell, and walks between them with dda2_line_interpolator — integer arithmetic
// whose accumulated remainder decides where each cell boundary falls. Mapping
// per pixel instead agrees with it over most of a span but not at every
// boundary, and one boundary is one whole mis-drawn column.
//
// The y axis needs no equivalent: an axis-aligned image has the same source row
// at both ends of a scanline, so the interpolator's y is constant and reduces
// to the single rounding in nearestSourceIndex.
func aggSpanColumns(width, sourceCols int, scale, offset float64) []int {
	if width <= 0 || sourceCols <= 0 || scale <= 0 {
		return nil
	}
	const subpixelScale = 1 << imageSubpixelShift
	begin := aggIround((0.5/scale + offset) * subpixelScale)
	end := aggIround((float64(width)+0.5)/scale*subpixelScale + offset*subpixelScale)

	out := make([]int, width)
	dda := newDDA2(begin, end, width)
	for x := range out {
		index := dda.value >> imageSubpixelShift
		if index < 0 {
			index = 0
		}
		if index >= sourceCols {
			index = sourceCols - 1
		}
		out[x] = index
		dda.next()
	}
	return out
}

// dda2 is agg::dda2_line_interpolator in its forward-adjusted form.
type dda2 struct {
	count, left, rem, mod, value int
}

func newDDA2(from, to, count int) dda2 {
	if count <= 0 {
		count = 1
	}
	delta := to - from
	d := dda2{
		count: count,
		left:  delta / count, // Go and C both truncate toward zero here.
		rem:   delta % count,
		value: from,
	}
	d.mod = d.rem
	if d.mod <= 0 {
		d.mod += count
		d.rem += count
		d.left--
	}
	d.mod -= count
	return d
}

func (d *dda2) next() {
	d.mod += d.rem
	d.value += d.left
	if d.mod > 0 {
		d.mod -= d.count
		d.value++
	}
}

// aggIround is agg::iround: round half away from zero.
func aggIround(v float64) int {
	if v < 0 {
		return int(v - 0.5)
	}
	return int(v + 0.5)
}

func (i *Image2D) rasterizeToSize(targetWidth, targetHeight int) (render.Image, bool) {
	return i.rasterizeToSizeAt(targetWidth, targetHeight, resampleScale{})
}

// rasterizeToSizeAt resamples to targetWidth x targetHeight. A zero scale falls
// back to the exact targetSize/sourceSize ratio.
func (i *Image2D) rasterizeToSizeAt(targetWidth, targetHeight int, scale resampleScale) (render.Image, bool) {
	rows := len(i.Data)
	if rows == 0 {
		return nil, false
	}

	cols := 0
	for _, row := range i.Data {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return nil, false
	}
	if targetWidth <= 0 {
		targetWidth = cols
	}
	if targetHeight <= 0 {
		targetHeight = rows
	}

	mapping := i.ScalarMap().Resolved()
	img := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	if targetWidth != cols || targetHeight != rows {
		for y := 0; y < targetHeight; y++ {
			for x := 0; x < targetWidth; x++ {
				v, ok := i.scaledScalarValueAt(y, x, targetHeight, targetWidth, rows, cols, scale)
				if !ok {
					continue
				}
				img.Set(x, y, toRGBAColor(mapping.Color(v, 1)))
			}
		}
		data := render.NewImageData(img)
		data.SetInterpolation(i.interpolationAfterScalarRasterize(targetWidth, targetHeight, cols, rows))
		data.SetAlpha(clampOneToOne(i.Alpha))
		return data, true
	}
	for row, values := range i.Data {
		for col := 0; col < cols; col++ {
			if col >= len(values) {
				continue
			}
			v := values[col]
			pixelY := row
			if i.Origin == ImageOriginLower {
				pixelY = rows - 1 - row
			}

			img.Set(col, pixelY, toRGBAColor(mapping.Color(v, 1)))
		}
	}

	data := render.NewImageData(img)
	data.SetInterpolation(i.Interpolation)
	data.SetAlpha(clampOneToOne(i.Alpha))
	return data, true
}

func (i *Image2D) rasterizeTransformed(ctx *DrawContext, anchor geom.Pt, angle float64) (render.Image, geom.Rect, bool) {
	if i == nil || ctx == nil || i.rgba != nil {
		// True-color images carry no scalar Data to resample here; let Draw fall
		// back to the renderer's ImageTransformer using the wrapped bitmap.
		return nil, geom.Rect{}, false
	}
	rows, cols := scalarImageDimensions(i.Data)
	if rows == 0 || cols == 0 {
		return nil, geom.Rect{}, false
	}
	filter, ok := scalarImageFilterForInterpolation(i.Interpolation, cols, rows, cols, rows)
	if !ok {
		switch normalizedImageInterpolation(i.Interpolation) {
		case "", "nearest", "none":
			filter = scalarImageFilter{}
		default:
			return nil, geom.Rect{}, false
		}
	}

	srcToDisplay, ok := i.sourceToDisplayAffine(ctx, anchor, angle, cols, rows)
	if !ok {
		return nil, geom.Rect{}, false
	}
	inv, ok := srcToDisplay.Invert()
	if !ok {
		return nil, geom.Rect{}, false
	}

	bounds := transformedSourceBounds(srcToDisplay, float64(cols), float64(rows))
	if !ctx.Clip.Empty() {
		bounds = bounds.Intersect(ctx.Clip)
	}
	if bounds.Empty() {
		return nil, geom.Rect{}, false
	}

	outW := int(math.Ceil(bounds.W()))
	outH := int(math.Ceil(bounds.H()))
	if outW <= 0 || outH <= 0 {
		return nil, geom.Rect{}, false
	}
	outRect := geom.Rect{
		Min: geom.Pt{X: math.Floor(bounds.Min.X), Y: math.Ceil(bounds.Min.Y)},
	}
	outRect.Max = geom.Pt{X: outRect.Min.X + float64(outW), Y: outRect.Min.Y + float64(outH)}

	mapping := i.ScalarMap().Resolved()
	img := image.NewRGBA(image.Rect(0, 0, outW, outH))
	for y := 0; y < outH; y++ {
		displayY := outRect.Max.Y - (float64(y) + 0.5)
		for x := 0; x < outW; x++ {
			displayX := outRect.Min.X + float64(x) + 0.5
			src := inv.Apply(geom.Pt{X: displayX, Y: displayY})
			if !transformedSourceInRange(src, rows, cols) {
				continue
			}

			rowCoord := src.Y - 0.5
			if i.Origin == ImageOriginUpper {
				rowCoord = float64(rows) - src.Y - 0.5
			}
			colCoord := src.X - 0.5
			v, ok := transformedScalarSample(i.Data, rowCoord, colCoord, filter)
			if !ok {
				continue
			}
			c := mapping.Color(v, 1)
			c = c.WithAlphaMultiplier(clampOneToOne(i.Alpha))
			img.Set(x, y, toRGBAColor(c))
		}
	}

	data := render.NewImageData(img)
	data.SetInterpolation("nearest")
	data.SetAlpha(1)
	return data, outRect, true
}

func (i *Image2D) sourceToDisplayAffine(ctx *DrawContext, anchor geom.Pt, angle float64, cols, rows int) (geom.Affine, bool) {
	if ctx == nil || cols <= 0 || rows <= 0 {
		return geom.Affine{}, false
	}
	y0 := i.YMin
	y1 := i.YMax
	if i.Origin == ImageOriginUpper {
		y0, y1 = i.YMax, i.YMin
	}
	p00 := ctx.DataToPixel.Apply(geom.Pt{X: i.XMin, Y: y0})
	p10 := ctx.DataToPixel.Apply(geom.Pt{X: i.XMax, Y: y0})
	p01 := ctx.DataToPixel.Apply(geom.Pt{X: i.XMin, Y: y1})

	rot := displayRotation(anchor, angle)
	p00 = rot.Apply(p00)
	p10 = rot.Apply(p10)
	p01 = rot.Apply(p01)

	return geom.Affine{
		A: (p10.X - p00.X) / float64(cols),
		B: (p10.Y - p00.Y) / float64(cols),
		C: (p01.X - p00.X) / float64(rows),
		D: (p01.Y - p00.Y) / float64(rows),
		E: p00.X,
		F: p00.Y,
	}, true
}

func displayRotation(anchor geom.Pt, angle float64) geom.Affine {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return geom.Affine{
		A: cos,
		B: sin,
		C: -sin,
		D: cos,
		E: anchor.X - cos*anchor.X + sin*anchor.Y,
		F: anchor.Y - sin*anchor.X - cos*anchor.Y,
	}
}

func transformedSourceBounds(t geom.Affine, width, height float64) geom.Rect {
	points := []geom.Pt{
		t.Apply(geom.Pt{}),
		t.Apply(geom.Pt{X: width}),
		t.Apply(geom.Pt{Y: height}),
		t.Apply(geom.Pt{X: width, Y: height}),
	}
	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, p := range points[1:] {
		minX = minF(minX, p.X)
		maxX = maxF(maxX, p.X)
		minY = minF(minY, p.Y)
		maxY = maxF(maxY, p.Y)
	}
	return geom.Rect{Min: geom.Pt{X: minX, Y: minY}, Max: geom.Pt{X: maxX, Y: maxY}}
}

func transformedScalarSample(data [][]float64, rowCoord, colCoord float64, filter scalarImageFilter) (float64, bool) {
	if filter.weight == nil || filter.radius <= 0 {
		return scalarAt(data, int(math.Floor(rowCoord+0.5)), int(math.Floor(colCoord+0.5)))
	}
	return filteredScalarSample(data, rowCoord, colCoord, filter)
}

func transformedSourceInRange(src geom.Pt, rows, cols int) bool {
	if rows <= 0 || cols <= 0 {
		return false
	}
	const negativeEdgeTolerance = 1.0 / 200.0
	const positiveEdgeTolerance = 1.0 / 58.0
	return src.X >= -negativeEdgeTolerance &&
		src.X <= float64(cols)+positiveEdgeTolerance &&
		src.Y >= -negativeEdgeTolerance &&
		src.Y <= float64(rows)+positiveEdgeTolerance
}

func (i *Image2D) scaledScalarValue(dstY, dstX, targetHeight, targetWidth, rows, cols int) (float64, bool) {
	return i.scaledScalarValueAt(dstY, dstX, targetHeight, targetWidth, rows, cols, resampleScale{})
}

func (i *Image2D) scaledScalarValueAt(dstY, dstX, targetHeight, targetWidth, rows, cols int, scale resampleScale) (float64, bool) {
	// The raster is built top-down, so with origin "lower" the topmost
	// destination row holds the last data row. Reflect the destination row and
	// then map, rather than mapping and reflecting the index: under the boundary
	// rule in scaledNearestIndex those are different answers wherever a source
	// coordinate snaps up onto a cell edge, and only this order reproduces the
	// affine Matplotlib hands its resampler. For the filtered branch the two
	// orders agree, so sharing this row keeps that path's output unchanged.
	srcRow := dstY
	if i != nil && i.Origin == ImageOriginLower {
		srcRow = targetHeight - 1 - dstY
	}
	if i != nil {
		if filter, ok := scalarImageFilterForInterpolation(i.Interpolation, targetWidth, targetHeight, cols, rows); ok {
			rowCoord := imageSourceCenter(srcRow, targetHeight, rows, scale.y, scale.oy) - 0.5
			colCoord := imageSourceCenter(dstX, targetWidth, cols, scale.x, scale.ox) - 0.5
			return filteredScalarSample(i.Data, rowCoord, colCoord, filter)
		}
	}
	row := scaledNearestIndexAt(srcRow, targetHeight, rows, scale.y, scale.oy)
	col := scale.column(dstX, targetWidth, cols)
	return scalarAt(i.Data, row, col)
}

func (i *Image2D) interpolationAfterScalarRasterize(targetWidth, targetHeight, sourceWidth, sourceHeight int) string {
	if i == nil {
		return ""
	}
	if _, ok := scalarImageFilterForInterpolation(i.Interpolation, targetWidth, targetHeight, sourceWidth, sourceHeight); ok {
		return "nearest"
	}
	return i.Interpolation
}

func scalarImageDimensions(data [][]float64) (int, int) {
	rows := len(data)
	cols := 0
	for _, row := range data {
		if len(row) > cols {
			cols = len(row)
		}
	}
	return rows, cols
}

func shouldRasterizeScalarDataStage(interpolation string, targetWidth, targetHeight, sourceWidth, sourceHeight int) bool {
	switch normalizedImageInterpolation(interpolation) {
	case "", "nearest", "none":
		return true
	}
	if targetWidth < 3*sourceWidth || targetHeight < 3*sourceHeight {
		return false
	}
	_, ok := scalarImageFilterForInterpolation(interpolation, targetWidth, targetHeight, sourceWidth, sourceHeight)
	return ok
}

func normalizedImageInterpolation(interpolation string) string {
	return strings.ToLower(strings.TrimSpace(interpolation))
}

func scaledNearestIndex(index, targetSize, sourceSize int) int {
	return scaledNearestIndexAt(index, targetSize, sourceSize, 0, 0)
}

func scaledNearestIndexAt(index, targetSize, sourceSize int, scale, offset float64) int {
	if targetSize <= 0 || sourceSize <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= targetSize {
		return sourceSize - 1
	}
	scaled := nearestSourceIndex(imageSourceCenter(index, targetSize, sourceSize, scale, offset))
	if scaled < 0 {
		return 0
	}
	if scaled >= sourceSize {
		return sourceSize - 1
	}
	return scaled
}

// imageSubpixelShift mirrors AGG's image_subpixel_shift: the span interpolator
// carries source coordinates as fixed point with this many fractional bits.
const imageSubpixelShift = 8

// nearestSourceIndex maps a source-space coordinate to a source cell the way
// agg::span_image_filter_*_nn does, which is how Matplotlib resamples an image
// with interpolation="nearest".
//
// AGG quantizes the coordinate to 1/256 of a cell with round-half-away-from-zero
// (agg_basics.h iround) and then takes the cell with an arithmetic right shift,
// which is a floor. Plain floor(u) is a different rule: a coordinate landing
// within 1/512 of a cell below a boundary snaps up onto the boundary and
// therefore belongs to the cell on the right, where floor keeps it on the left.
// A whole source column or row is drawn from the wrong cell whenever a boundary
// happens to fall in that gap.
//
// Note the caller models the coordinate as (dest+0.5)*src/dst, with no
// translation term. Matplotlib's affine also carries a sub-pixel translation
// when the image does not exactly fill its clipped bbox, plus the scale
// correction it applies after rounding the output buffer up to whole pixels,
// and it interpolates the fixed-point coordinate across a scanline by DDA
// rather than transforming every pixel. None of that is modelled here.
func nearestSourceIndex(u float64) int {
	scaled := u * (1 << imageSubpixelShift)
	if scaled < 0 {
		scaled -= 0.5
	} else {
		scaled += 0.5
	}
	// int(scaled) truncates toward zero and the shift floors, exactly as the C++
	// does. The two disagree for negative coordinates, so keep both steps rather
	// than folding them into one floor.
	return int(scaled) >> imageSubpixelShift
}

// imageSourceCenter is the source-space coordinate of a destination pixel's
// centre. scale is device pixels per source cell and offset the source-space
// coordinate the output buffer starts at; a non-positive scale falls back to the
// exact targetSize/sourceSize ratio, which is the same mapping for an image the
// clip box does not cut.
func imageSourceCenter(index, targetSize, sourceSize int, scale, offset float64) float64 {
	if scale > 0 {
		return (float64(index)+0.5)/scale + offset
	}
	return (float64(index) + 0.5) * float64(sourceSize) / float64(targetSize)
}

func bilinearScalarSample(data [][]float64, rowCoord, colCoord float64) (float64, bool) {
	return filteredScalarSample(data, rowCoord, colCoord, scalarImageFilter{
		radius: 1,
		weight: func(x float64) float64 { return 1 - x },
	})
}

type scalarImageFilter struct {
	radius float64
	weight func(float64) float64
}

func scalarImageFilterForInterpolation(interpolation string, targetWidth, targetHeight, sourceWidth, sourceHeight int) (scalarImageFilter, bool) {
	switch normalizedImageInterpolation(interpolation) {
	case "bilinear":
		return scalarImageFilter{radius: 1, weight: func(x float64) float64 { return 1 - x }}, true
	case "hanning":
		return scalarImageFilter{radius: 1, weight: func(x float64) float64 { return 0.5 + 0.5*math.Cos(math.Pi*x) }}, true
	case "hamming":
		return scalarImageFilter{radius: 1, weight: func(x float64) float64 { return 0.54 + 0.46*math.Cos(math.Pi*x) }}, true
	case "hermite":
		return scalarImageFilter{radius: 1, weight: func(x float64) float64 { return (2*x-3)*x*x + 1 }}, true
	case "quadric":
		return scalarImageFilter{radius: 1.5, weight: func(x float64) float64 {
			if x < 0.5 {
				return 0.75 - x*x
			}
			if x < 1.5 {
				t := x - 1.5
				return 0.5 * t * t
			}
			return 0
		}}, true
	case "bicubic":
		return scalarImageFilter{radius: 2, weight: cubicBSplineWeight}, true
	case "kaiser":
		k := newKaiserScalarFilter(6.33)
		return scalarImageFilter{radius: 1, weight: k.weight}, true
	case "catrom":
		return scalarImageFilter{radius: 2, weight: func(x float64) float64 {
			if x < 1 {
				return 0.5 * (2 + x*x*(-5+x*3))
			}
			if x < 2 {
				return 0.5 * (4 + x*(-8+x*(5-x)))
			}
			return 0
		}}, true
	case "mitchell":
		return newMitchellScalarFilter(1.0/3.0, 1.0/3.0), true
	case "spline16":
		return scalarImageFilter{radius: 2, weight: func(x float64) float64 {
			if x < 1 {
				return ((x-9.0/5.0)*x-1.0/5.0)*x + 1
			}
			return ((-1.0/3.0*(x-1)+4.0/5.0)*(x-1) - 7.0/15.0) * (x - 1)
		}}, true
	case "spline36":
		return scalarImageFilter{radius: 3, weight: func(x float64) float64 {
			if x < 1 {
				return ((13.0/11.0*x-453.0/209.0)*x-3.0/209.0)*x + 1
			}
			if x < 2 {
				return ((-6.0/11.0*(x-1)+270.0/209.0)*(x-1) - 156.0/209.0) * (x - 1)
			}
			return ((1.0/11.0*(x-2)-45.0/209.0)*(x-2) + 26.0/209.0) * (x - 2)
		}}, true
	case "gaussian":
		return scalarImageFilter{radius: 2, weight: func(x float64) float64 {
			return math.Exp(-2*x*x) * math.Sqrt(2/math.Pi)
		}}, true
	case "bessel":
		return scalarImageFilter{radius: 3.2383, weight: func(x float64) float64 {
			if x == 0 {
				return math.Pi / 4
			}
			return math.J1(math.Pi*x) / (2 * x)
		}}, true
	case "sinc":
		return newWindowedSincScalarFilter(4.0, nil), true
	case "lanczos":
		return newWindowedSincScalarFilter(4.0, func(x, radius float64) float64 {
			return sincWeight(x / radius)
		}), true
	case "blackman":
		return newWindowedSincScalarFilter(4.0, func(x, radius float64) float64 {
			xr := math.Pi * x / radius
			return 0.42 + 0.5*math.Cos(xr) + 0.08*math.Cos(2*xr)
		}), true
	case "auto", "antialiased":
		if shouldUseNearestForScalarAuto(targetWidth, targetHeight, sourceWidth, sourceHeight) {
			return scalarImageFilter{}, false
		}
		return scalarImageFilter{radius: 1, weight: func(x float64) float64 { return 0.5 + 0.5*math.Cos(math.Pi*x) }}, true
	default:
		return scalarImageFilter{}, false
	}
}

func shouldUseNearestForScalarAuto(targetWidth, targetHeight, sourceWidth, sourceHeight int) bool {
	return (targetWidth > 3*sourceWidth || targetWidth == sourceWidth || targetWidth == 2*sourceWidth) &&
		(targetHeight > 3*sourceHeight || targetHeight == sourceHeight || targetHeight == 2*sourceHeight)
}

func cubicBSplineWeight(x float64) float64 {
	pow3 := func(v float64) float64 {
		if v <= 0 {
			return 0
		}
		return v * v * v
	}
	return (pow3(x+2) - 4*pow3(x+1) + 6*pow3(x) - 4*pow3(x-1)) / 6
}

type kaiserScalarFilter struct {
	a   float64
	i0a float64
}

func newKaiserScalarFilter(a float64) kaiserScalarFilter {
	k := kaiserScalarFilter{a: a}
	k.i0a = 1 / besselI0(a)
	return k
}

func (k kaiserScalarFilter) weight(x float64) float64 {
	return besselI0(k.a*math.Sqrt(1-x*x)) * k.i0a
}

func besselI0(x float64) float64 {
	const epsilon = 1e-12
	sum := 1.0
	y := x * x / 4
	t := y
	for i := 2; t > epsilon; i++ {
		sum += t
		t *= y / (float64(i) * float64(i))
	}
	return sum
}

func newMitchellScalarFilter(b, c float64) scalarImageFilter {
	p0 := (6 - 2*b) / 6
	p2 := (-18 + 12*b + 6*c) / 6
	p3 := (12 - 9*b - 6*c) / 6
	q0 := (8*b + 24*c) / 6
	q1 := (-12*b - 48*c) / 6
	q2 := (6*b + 30*c) / 6
	q3 := (-b - 6*c) / 6
	return scalarImageFilter{radius: 2, weight: func(x float64) float64 {
		if x < 1 {
			return p0 + x*x*(p2+x*p3)
		}
		if x < 2 {
			return q0 + x*(q1+x*(q2+x*q3))
		}
		return 0
	}}
}

func newWindowedSincScalarFilter(radius float64, window func(float64, float64) float64) scalarImageFilter {
	if radius < 2 {
		radius = 2
	}
	return scalarImageFilter{radius: radius, weight: func(x float64) float64 {
		if x == 0 {
			return 1
		}
		if x > radius {
			return 0
		}
		w := 1.0
		if window != nil {
			w = window(x, radius)
		}
		return sincWeight(x) * w
	}}
}

func sincWeight(x float64) float64 {
	if x == 0 {
		return 1
	}
	x *= math.Pi
	return math.Sin(x) / x
}

func filteredScalarSample(data [][]float64, rowCoord, colCoord float64, filter scalarImageFilter) (float64, bool) {
	rows := len(data)
	if rows == 0 {
		return 0, false
	}
	cols := 0
	for _, row := range data {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return 0, false
	}

	rowCoord = clampFloat(rowCoord, 0, float64(rows-1))
	colCoord = clampFloat(colCoord, 0, float64(cols-1))
	if filter.weight == nil || filter.radius <= 0 {
		return 0, false
	}

	rowStart := int(math.Ceil(rowCoord - filter.radius))
	rowEnd := int(math.Floor(rowCoord + filter.radius))
	colStart := int(math.Ceil(colCoord - filter.radius))
	colEnd := int(math.Floor(colCoord + filter.radius))

	sum := 0.0
	weightSum := 0.0
	for row := rowStart; row <= rowEnd; row++ {
		wy := filter.weight(math.Abs(rowCoord - float64(row)))
		if wy == 0 {
			continue
		}
		clampedRow := minInt(maxInt(row, 0), rows-1)
		for col := colStart; col <= colEnd; col++ {
			wx := filter.weight(math.Abs(colCoord - float64(col)))
			if wx == 0 {
				continue
			}
			v, ok := scalarAt(data, clampedRow, minInt(maxInt(col, 0), cols-1))
			if !ok {
				continue
			}
			weight := wy * wx
			sum += v * weight
			weightSum += weight
		}
	}
	if weightSum == 0 {
		return 0, false
	}
	return sum / weightSum, true
}

func scalarAt(data [][]float64, row, col int) (float64, bool) {
	if row < 0 || row >= len(data) || col < 0 || col >= len(data[row]) {
		return 0, false
	}
	v := data[row][col]
	if !isFinite(v) {
		return 0, false
	}
	return v, true
}

func (i *Image2D) destinationRect(ctx *DrawContext) geom.Rect {
	if ctx == nil {
		return geom.Rect{}
	}
	p1 := ctx.DataToPixel.Apply(geom.Pt{X: i.XMin, Y: i.YMin})
	p2 := ctx.DataToPixel.Apply(geom.Pt{X: i.XMax, Y: i.YMax})
	minX := minF(p1.X, p2.X)
	maxX := maxF(p1.X, p2.X)
	minY := minF(p1.Y, p2.Y)
	maxY := maxF(p1.Y, p2.Y)
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}
}

func (i *Image2D) rotationAnchor(ctx *DrawContext, dst geom.Rect) geom.Pt {
	switch i.RotateAt {
	case ImageAnchorTopLeft:
		return dst.Min
	case ImageAnchorTopRight:
		return geom.Pt{X: dst.Max.X, Y: dst.Min.Y}
	case ImageAnchorBottomLeft:
		return geom.Pt{X: dst.Min.X, Y: dst.Max.Y}
	case ImageAnchorBottomRight:
		return dst.Max
	case ImageAnchorCustom:
		if ctx != nil {
			return ctx.DataToPixel.Apply(geom.Pt{X: i.RotateX, Y: i.RotateY})
		}
		fallthrough
	case ImageAnchorCenter:
		fallthrough
	default:
		return geom.Pt{
			X: (dst.Min.X + dst.Max.X) * 0.5,
			Y: (dst.Min.Y + dst.Max.Y) * 0.5,
		}
	}
}

func imageTransform(dst geom.Rect, raster render.Image, anchor geom.Pt, angle float64) geom.Affine {
	srcW, srcH := raster.Size()
	if srcW <= 0 || srcH <= 0 {
		return geom.Identity()
	}

	sx := dst.W() / float64(srcW)
	sy := dst.H() / float64(srcH)
	// Display space is y-up. The raster's row 0 is its top row (origin handling
	// already baked into the bitmap), so it must map to the top edge dst.Max.Y
	// and rows advance downward to dst.Min.Y. This keeps the rotated path's
	// orientation identical to the axis-aligned drawImageDirect path at angle 0.
	scale := geom.Affine{
		A: sx,
		D: -sy,
		E: dst.Min.X,
		F: dst.Max.Y,
	}

	cos := math.Cos(angle)
	sin := math.Sin(angle)
	rot := geom.Affine{
		A: cos,
		B: sin,
		C: -sin,
		D: cos,
		E: anchor.X,
		F: anchor.Y,
	}
	negAnchor := geom.Affine{
		A: 1,
		D: 1,
		E: -anchor.X,
		F: -anchor.Y,
	}

	around := rot.Mul(negAnchor)
	return around.Mul(scale)
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func toRGBAColor(c render.Color) color.Color {
	red := toByte(c.R)
	green := toByte(c.G)
	blue := toByte(c.B)
	alpha := toByte(c.A)
	return color.RGBA{R: red, G: green, B: blue, A: alpha}
}

func toByte(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v * 255)
}

func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func dataRange(data [][]float64) (min float64, max float64) {
	first := true
	for _, row := range data {
		for _, v := range row {
			if !isFinite(v) {
				continue
			}
			if first {
				min = v
				max = v
				first = false
				continue
			}
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
	}
	if first {
		return 0, 1
	}
	return min, max
}

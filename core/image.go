package core

import (
	"image"
	"image/color"
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Draw renders the rasterized image through the renderer.
func (i *Image2D) Draw(r render.Renderer, ctx *DrawContext) {
	if i == nil || r == nil {
		return
	}

	dst := i.destinationRect(ctx)
	if dst.W() <= 0 || dst.H() <= 0 {
		return
	}
	dst = matplotlibImageDrawRect(dst)

	raster, ok := i.rasterizeForRect(dst)
	if !ok {
		return
	}

	angleRad := i.AngleDeg * math.Pi / 180
	if angleRad == 0 {
		r.Image(raster, dst)
		return
	}

	if tr, ok := r.(render.ImageTransformer); ok {
		anchor := i.rotationAnchor(ctx, dst)
		transform := imageTransform(dst, raster, anchor, angleRad)
		tr.ImageTransformed(raster, dst, transform)
		return
	}

	// Fallback: ignore rotation and render axis-aligned image.
	r.Image(raster, dst)
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
	return i.rasterizeToSize(0, 0)
}

func (i *Image2D) rasterizeForRect(dst geom.Rect) (render.Image, bool) {
	if i == nil || i.AngleDeg != 0 {
		return i.rasterize()
	}
	width := int(math.Round(math.Abs(dst.W())))
	height := int(math.Round(math.Abs(dst.H())))
	if width <= len(i.Data[0]) || height <= len(i.Data) {
		return i.rasterize()
	}
	switch i.Interpolation {
	case "", "nearest", "none", "bilinear":
	default:
		return i.rasterize()
	}
	return i.rasterizeToSize(width, height)
}

func (i *Image2D) rasterizeToSize(targetWidth, targetHeight int) (render.Image, bool) {
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
				v, ok := i.scaledScalarValue(y, x, targetHeight, targetWidth, rows, cols)
				if !ok {
					continue
				}
				img.Set(x, y, toRGBAColor(mapping.Color(v, 1)))
			}
		}
		data := render.NewImageData(img)
		if i.Interpolation == "bilinear" {
			data.SetInterpolation("nearest")
		} else {
			data.SetInterpolation(i.Interpolation)
		}
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

func (i *Image2D) scaledScalarValue(dstY, dstX, targetHeight, targetWidth, rows, cols int) (float64, bool) {
	if i != nil && i.Interpolation == "bilinear" {
		rowCoord := imageSourceCoord(dstY, targetHeight, rows)
		if i.Origin == ImageOriginLower {
			rowCoord = float64(rows-1) - rowCoord
		}
		colCoord := imageSourceCoord(dstX, targetWidth, cols)
		return bilinearScalarSample(i.Data, rowCoord, colCoord)
	}
	row := scaledNearestIndex(dstY, targetHeight, rows)
	col := scaledNearestIndex(dstX, targetWidth, cols)
	if i != nil && i.Origin == ImageOriginLower {
		row = rows - 1 - row
	}
	return scalarAt(i.Data, row, col)
}

func scaledNearestIndex(index, targetSize, sourceSize int) int {
	if targetSize <= 0 || sourceSize <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= targetSize {
		return sourceSize - 1
	}
	scaled := int(math.Floor((float64(index) + 0.5) * float64(sourceSize) / float64(targetSize)))
	if scaled < 0 {
		return 0
	}
	if scaled >= sourceSize {
		return sourceSize - 1
	}
	return scaled
}

func imageSourceCoord(index, targetSize, sourceSize int) float64 {
	if targetSize <= 0 || sourceSize <= 0 {
		return 0
	}
	return (float64(index)+0.5)*float64(sourceSize)/float64(targetSize) - 0.5
}

func bilinearScalarSample(data [][]float64, rowCoord, colCoord float64) (float64, bool) {
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
	row0 := int(math.Floor(rowCoord))
	col0 := int(math.Floor(colCoord))
	row1 := minInt(row0+1, rows-1)
	col1 := minInt(col0+1, cols-1)
	wy := rowCoord - float64(row0)
	wx := colCoord - float64(col0)

	v00, ok00 := scalarAt(data, row0, col0)
	v10, ok10 := scalarAt(data, row0, col1)
	v01, ok01 := scalarAt(data, row1, col0)
	v11, ok11 := scalarAt(data, row1, col1)
	if !ok00 || !ok10 || !ok01 || !ok11 {
		return 0, false
	}
	top := v00*(1-wx) + v10*wx
	bottom := v01*(1-wx) + v11*wx
	return top*(1-wy) + bottom*wy, true
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

package agg

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"runtime"
	"sync"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) hasClipPath() bool {
	return len(r.clipPaths) > 0 && r.ctx != nil && r.ctx.image != nil
}

type pixelRegion struct {
	minX int
	minY int
	maxX int
	maxY int
}

func (r *Renderer) withClipPathMask(bounds geom.Rect, haveBounds bool, draw func()) {
	r.withClipPathMaskComposite(bounds, haveBounds, false, draw)
}

func (r *Renderer) withClipPathMaskPremultiplied(bounds geom.Rect, haveBounds bool, draw func()) {
	r.withClipPathMaskComposite(bounds, haveBounds, true, draw)
}

func (r *Renderer) withClipPathMaskComposite(bounds geom.Rect, haveBounds bool, premultiplied bool, draw func()) {
	paths := clonePaths(r.clipPaths)
	if len(paths) == 0 || r.ctx == nil || r.ctx.image == nil {
		draw()
		return
	}

	region, ok := r.clipCompositeRegion(bounds, haveBounds)
	if !ok {
		return
	}
	masks, ok := r.clipMasksForPaths(paths)
	if !ok {
		return
	}

	target := r.ctx
	temp := r.clipTempSurface()
	clearImageRegion(temp.image, region)

	oldPaths := r.clipPaths
	r.clipDepth++
	r.ctx = temp
	r.clipPaths = nil
	r.ctx.ClipBox(float64(region.minX), float64(region.minY), float64(region.maxX), float64(region.maxY))
	draw()
	r.clipDepth--
	r.clipPaths = oldPaths
	r.ctx = target
	r.applyClipRect()

	r.compositeClipSurface(temp.image, masks, region, premultiplied)
}

func (r *Renderer) clipTempSurface() *aggSurface {
	if r.clipDepth > 0 {
		return newAggSurface(r.width, r.height)
	}
	if r.clipScratch == nil || r.clipScratch.image == nil || r.clipScratch.image.Width() != r.width || r.clipScratch.image.Height() != r.height {
		r.clipScratch = newAggSurface(r.width, r.height)
	}
	return r.clipScratch
}

func (r *Renderer) clipCompositeRegion(bounds geom.Rect, haveBounds bool) (pixelRegion, bool) {
	region := pixelRegion{
		minX: 0,
		minY: 0,
		maxX: r.width,
		maxY: r.height,
	}
	if r.clipRect != nil {
		minX, minY, maxX, maxY := quantizedClipBox(*r.clipRect)
		region.minX = maxInt(region.minX, minX)
		region.minY = maxInt(region.minY, minY)
		region.maxX = minInt(region.maxX, maxX)
		region.maxY = minInt(region.maxY, maxY)
	}
	if haveBounds {
		region.minX = maxInt(region.minX, int(math.Floor(bounds.Min.X)))
		region.minY = maxInt(region.minY, int(math.Floor(bounds.Min.Y)))
		region.maxX = minInt(region.maxX, int(math.Ceil(bounds.Max.X)))
		region.maxY = minInt(region.maxY, int(math.Ceil(bounds.Max.Y)))
	}
	region.minX = maxInt(region.minX, 0)
	region.minY = maxInt(region.minY, 0)
	region.maxX = minInt(region.maxX, r.width)
	region.maxY = minInt(region.maxY, r.height)
	return region, region.minX < region.maxX && region.minY < region.maxY
}

func (r *Renderer) clipMasksForPaths(paths []geom.Path) ([][]uint8, bool) {
	masks := make([][]uint8, 0, len(paths))
	for _, path := range paths {
		mask := r.clipMaskForPath(path)
		if len(mask) == 0 {
			return nil, false
		}
		masks = append(masks, mask)
	}
	return masks, true
}

func clearImageRegion(img *agglib.Image, region pixelRegion) {
	if img == nil {
		return
	}
	stride := img.Stride()
	if stride <= 0 {
		return
	}
	clearRows := func(rows pixelRegion) {
		for y := rows.minY; y < rows.maxY; y++ {
			start := y*stride + rows.minX*4
			end := y*stride + rows.maxX*4
			if start < 0 {
				start = 0
			}
			if end > len(img.Data) {
				end = len(img.Data)
			}
			clear(img.Data[start:end])
		}
	}
	runRowWorkers(region, clearRows)
}

const minParallelClipPixels = 65536

func runRowWorkers(region pixelRegion, fn func(pixelRegion)) {
	ranges := parallelRowRanges(region, parallelWorkersForRegion(region))
	if len(ranges) <= 1 {
		fn(region)
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(ranges))
	for _, rows := range ranges {
		rows := rows
		go func() {
			defer wg.Done()
			fn(rows)
		}()
	}
	wg.Wait()
}

func parallelWorkersForRegion(region pixelRegion) int {
	rows := region.maxY - region.minY
	cols := region.maxX - region.minX
	if rows <= 0 || cols <= 0 || rows*cols < minParallelClipPixels {
		return 1
	}
	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		return 1
	}
	if workers > rows {
		workers = rows
	}
	return workers
}

func parallelRowRanges(region pixelRegion, workers int) []pixelRegion {
	rows := region.maxY - region.minY
	if rows <= 0 {
		return nil
	}
	if workers <= 1 {
		return []pixelRegion{region}
	}
	if workers > rows {
		workers = rows
	}
	ranges := make([]pixelRegion, 0, workers)
	base := rows / workers
	extra := rows % workers
	start := region.minY
	for i := 0; i < workers; i++ {
		n := base
		if i < extra {
			n++
		}
		next := start + n
		part := region
		part.minY = start
		part.maxY = next
		ranges = append(ranges, part)
		start = next
	}
	return ranges
}

func (r *Renderer) compositeClipSurface(src *agglib.Image, masks [][]uint8, region pixelRegion, premultiplied bool) {
	dst := r.ctx.image
	if src == nil || dst == nil {
		return
	}
	if src.Width() != dst.Width() || src.Height() != dst.Height() {
		return
	}
	if region.minX >= region.maxX || region.minY >= region.maxY {
		return
	}

	srcStride := src.Stride()
	dstStride := dst.Stride()
	width := r.width
	compositeRows := func(rows pixelRegion) {
		for y := rows.minY; y < rows.maxY; y++ {
			for x := rows.minX; x < rows.maxX; x++ {
				maskA := clipMaskAlpha(masks, width, x, y)
				if maskA == 0 {
					continue
				}
				srcOff := y*srcStride + x*4
				dstOff := y*dstStride + x*4
				if srcOff < 0 || srcOff+3 >= len(src.Data) || dstOff < 0 || dstOff+3 >= len(dst.Data) {
					continue
				}
				sa := src.Data[srcOff+3]
				if sa == 0 {
					continue
				}
				sr := src.Data[srcOff]
				sg := src.Data[srcOff+1]
				sb := src.Data[srcOff+2]
				if premultiplied {
					sr = unpremultiplyAlphaByte(sr, sa)
					sg = unpremultiplyAlphaByte(sg, sa)
					sb = unpremultiplyAlphaByte(sb, sa)
				}
				blendPixelRGBA(dst.Data[dstOff:dstOff+4], render.Color{
					R: float64(sr) / 255,
					G: float64(sg) / 255,
					B: float64(sb) / 255,
					A: (float64(sa) / 255) * (float64(maskA) / 255),
				})
			}
		}
	}
	runRowWorkers(region, compositeRows)
}

func unpremultiplyAlphaByte(channel, alpha uint8) uint8 {
	if alpha == 0 {
		return 0
	}
	if alpha == 255 || channel == 0 {
		return channel
	}
	v := int(channel) * 255 / int(alpha)
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func clipMaskAlpha(masks [][]uint8, width, x, y int) uint8 {
	alpha := 255
	i := y*width + x
	for _, mask := range masks {
		if len(mask) == 0 {
			return 0
		}
		if i < 0 || i >= len(mask) {
			return 0
		}
		alpha = alpha * int(mask[i]) / 255
		if alpha == 0 {
			return 0
		}
	}
	return uint8(alpha)
}

func (r *Renderer) clipMaskForPath(path geom.Path) []uint8 {
	if len(path.C) == 0 || !path.Validate() || r.width <= 0 || r.height <= 0 {
		return nil
	}
	if r.clipMaskMap == nil {
		r.clipMaskMap = make(map[clipMaskKey][]uint8)
	}
	key := clipMaskKey{
		width:  r.width,
		height: r.height,
		hash:   hashPath(path),
	}
	if mask, ok := r.clipMaskMap[key]; ok {
		return mask
	}

	surface := newAggSurface(r.width, r.height)
	surface.Clear(agglib.NewColor(0, 0, 0, 0))
	oldCtx := r.ctx
	r.ctx = surface
	r.ctx.ClipBox(0, 0, float64(r.width), float64(r.height))
	r.buildPath(path)
	r.ctx.SetFillColor(agglib.NewColor(255, 255, 255, 255))
	r.ctx.Fill()
	r.ctx = oldCtx

	img := surface.image
	mask := make([]uint8, r.width*r.height)
	stride := img.Stride()
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			srcOff := y*stride + x*4 + 3
			if srcOff >= 0 && srcOff < len(img.Data) {
				mask[y*r.width+x] = img.Data[srcOff]
			}
		}
	}
	r.clipMaskMap[key] = mask
	return mask
}

func hashPath(path geom.Path) uint64 {
	h := fnv.New64a()
	var buf [8]byte
	for _, cmd := range path.C {
		_, _ = h.Write([]byte{byte(cmd)})
	}
	for _, pt := range path.V {
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(quantize(pt.X)))
		_, _ = h.Write(buf[:])
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(quantize(pt.Y)))
		_, _ = h.Write(buf[:])
	}
	return h.Sum64()
}

func clonePaths(paths []geom.Path) []geom.Path {
	if len(paths) == 0 {
		return nil
	}
	out := make([]geom.Path, len(paths))
	for i, path := range paths {
		out[i] = clonePath(path)
	}
	return out
}

func clonePath(path geom.Path) geom.Path {
	return geom.Path{
		V: append([]geom.Pt(nil), path.V...),
		C: append([]geom.Cmd(nil), path.C...),
	}
}

func (r *Renderer) applyClipRect() {
	if r.clipRect != nil {
		minX, minY, maxX, maxY := quantizedClipBox(*r.clipRect)
		r.ctx.ClipBox(float64(minX), float64(minY), float64(maxX), float64(maxY))
		return
	}
	r.ctx.ClipBox(0, 0, float64(r.width), float64(r.height))
}

func quantizedClipBox(rect geom.Rect) (minX, minY, maxX, maxY int) {
	// Matplotlib 3.10.9's RendererAgg::set_clipbox converts each display-space
	// clip edge with int(floor(edge+0.5)). The rect is already in device
	// coordinates here, so apply the same half-up integer quantization directly.
	return int(math.Floor(rect.Min.X + 0.5)),
		int(math.Floor(rect.Min.Y + 0.5)),
		int(math.Floor(rect.Max.X + 0.5)),
		int(math.Floor(rect.Max.Y + 0.5))
}

// renderImageToAGG converts a renderer image into an AGG image type.
func renderImageToAGG(img render.Image) (*agglib.Image, bool) {
	if img == nil {
		return nil, false
	}

	rgbaImage, ok := img.(render.RGBAImage)
	if !ok {
		return nil, false
	}

	rgba := rgbaImage.RGBA()
	if rgba == nil || rgba.Bounds().Dx() <= 0 || rgba.Bounds().Dy() <= 0 {
		return nil, false
	}
	bounds := rgba.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	stride := width * 4
	data := make([]uint8, height*stride)
	alpha := extractImageAlpha(img)
	for y := 0; y < height; y++ {
		srcOff := rgba.PixOffset(bounds.Min.X, bounds.Min.Y+y)
		dstOff := y * stride
		copy(data[dstOff:dstOff+stride], rgba.Pix[srcOff:srcOff+stride])
		for x := 0; x < width; x++ {
			off := dstOff + x*4
			effectiveAlpha := float64(data[off+3]) * alpha / 255
			if effectiveAlpha >= 1 {
				continue
			}
			data[off+0] = uint8(float64(data[off+0]) * effectiveAlpha)
			data[off+1] = uint8(float64(data[off+1]) * effectiveAlpha)
			data[off+2] = uint8(float64(data[off+2]) * effectiveAlpha)
			data[off+3] = uint8(effectiveAlpha * 255)
		}
	}
	return agglib.NewImage(data, width, height, stride), true
}

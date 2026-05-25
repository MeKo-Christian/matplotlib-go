package gobasic

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/vector"
)

// quantizationEpsilon is the precision limit for float values to ensure determinism.
// All floating point coordinates and measurements are snapped to this precision.
const (
	quantizationEpsilon = 1e-6
	defaultFontHeight   = 13.0
)

// quantize snaps a float64 value to quantizationEpsilon precision to eliminate
// tiny differences that could lead to cross-platform rendering variations.
func quantize(v float64) float64 {
	return math.Round(v/quantizationEpsilon) * quantizationEpsilon
}

func scalePointSize(size float64, dpi uint) float64 {
	if dpi == 0 {
		dpi = 72
	}
	return size * float64(dpi) / 72.0
}

// quantizePt quantizes both X and Y coordinates of a point.
func quantizePt(p geom.Pt) geom.Pt {
	return geom.Pt{
		X: quantize(p.X),
		Y: quantize(p.Y),
	}
}

// quantizePath quantizes all vertices in a path for deterministic rendering.
func quantizePath(p geom.Path) geom.Path {
	result := geom.Path{
		C: make([]geom.Cmd, len(p.C)),
		V: make([]geom.Pt, len(p.V)),
	}

	copy(result.C, p.C)
	for i, v := range p.V {
		result.V[i] = quantizePt(v)
	}

	return result
}

// state represents a saved graphics state.
type state struct {
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

// Renderer implements render.Renderer using pure Go dependencies.
type Renderer struct {
	dst         *image.RGBA
	viewport    geom.Rect
	began       bool
	stack       []state
	clipRect    *geom.Rect
	clipPaths   []geom.Path
	clipMaskMap map[clipMaskKey]*image.Alpha
	rasterizer  *vector.Rasterizer
	lastFontKey string
	resolution  uint
}

type clipMaskKey struct {
	width  int
	height int
	hash   uint64
}

var (
	_ render.Renderer           = (*Renderer)(nil)
	_ render.DPIAware           = (*Renderer)(nil)
	_ render.ImageTransformer   = (*Renderer)(nil)
	_ render.TextDrawer         = (*Renderer)(nil)
	_ render.TextPather         = (*Renderer)(nil)
	_ render.RotatedTextDrawer  = (*Renderer)(nil)
	_ render.VerticalTextDrawer = (*Renderer)(nil)
	_ render.PNGExporter        = (*Renderer)(nil)
)

// New creates a new GoBasic renderer with the specified dimensions and background color.
func New(w, h int, bg render.Color) *Renderer {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill with background color using premultiplied alpha
	red, green, blue, alpha := bg.ToPremultipliedRGBA()
	bgColor := color.RGBA{R: red, G: green, B: blue, A: alpha}

	// Fill the entire image with background color
	for y := 0; y < h; y++ {
		row := dst.PixOffset(0, y)
		for x := 0; x < w; x++ {
			i := row + x*4
			dst.Pix[i] = bgColor.R
			dst.Pix[i+1] = bgColor.G
			dst.Pix[i+2] = bgColor.B
			dst.Pix[i+3] = bgColor.A
		}
	}

	return &Renderer{
		dst:         dst,
		rasterizer:  vector.NewRasterizer(w, h),
		resolution:  72,
		clipMaskMap: make(map[clipMaskKey]*image.Alpha),
	}
}

// Begin starts a drawing session with the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("Begin called twice")
	}
	r.began = true
	r.viewport = viewport
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipPaths = r.clipPaths[:0]
	return nil
}

// End finishes the drawing session.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("End called before Begin")
	}
	r.began = false
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipPaths = r.clipPaths[:0]
	return nil
}

// Save pushes the current graphics state onto the stack.
func (r *Renderer) Save() {
	var clipCopy *geom.Rect
	if r.clipRect != nil {
		rectCopy := *r.clipRect
		clipCopy = &rectCopy
	}
	r.stack = append(r.stack, state{
		clipRect:  clipCopy,
		clipPaths: clonePaths(r.clipPaths),
	})
}

// Restore pops the graphics state from the stack.
func (r *Renderer) Restore() {
	if len(r.stack) == 0 {
		return // No state to restore
	}

	// Pop the last state
	s := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]

	// Restore state
	r.clipRect = s.clipRect
	r.clipPaths = clonePaths(s.clipPaths)
}

// ClipRect sets a rectangular clip region. The incoming rect is in y-up display
// space; it is flipped to the y-down device buffer so it matches the device-space
// pixel comparisons in fillPath/drawTargetRect.
func (r *Renderer) ClipRect(rect geom.Rect) {
	rect = r.devRect(rect)
	if r.clipRect == nil {
		r.clipRect = &rect
	} else {
		// Intersect with existing clip
		intersected := r.clipRect.Intersect(rect)
		r.clipRect = &intersected
	}
}

// ClipPath adds a path-based clip region to the current graphics state. The
// incoming path is in y-up display space; it is flipped to device space so the
// rasterized clip mask aligns with device pixels.
func (r *Renderer) ClipPath(p geom.Path) {
	if len(p.C) == 0 || !p.Validate() {
		return
	}
	r.clipPaths = append(r.clipPaths, clonePath(r.devPath(p)))
}

// Path draws a path with the given paint style. The incoming path is in y-up
// display coordinates; it is flipped to the y-down device buffer once here, then
// the device-space pipeline runs unchanged.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	r.pathDevice(r.devPath(p), paint)
}

// pathDevice draws a path that is already in y-down device coordinates.
func (r *Renderer) pathDevice(p geom.Path, paint *render.Paint) {
	if paint == nil {
		return
	}
	if render.DrawPathWithEffects(r, p, paint, r.pathDevice) {
		return
	}
	if !p.Validate() {
		return // Invalid path
	}

	// Quantize path coordinates for deterministic rendering
	p = quantizePath(p)

	// Quantize paint parameters for consistency
	quantizedPaint := &render.Paint{
		LineWidth:  quantize(paint.LineWidth),
		LineJoin:   paint.LineJoin,
		LineCap:    paint.LineCap,
		MiterLimit: quantize(paint.MiterLimit),
		Stroke:     paint.Stroke,
		Fill:       paint.Fill,
		Dashes:     make([]float64, len(paint.Dashes)),
	}

	// Quantize dash pattern
	for i, dash := range paint.Dashes {
		quantizedPaint.Dashes[i] = quantize(dash)
	}

	// Fill first if requested
	if quantizedPaint.Fill.A > 0 {
		r.fillPath(p, quantizedPaint.Fill)
	}

	// Then stroke if requested
	if quantizedPaint.Stroke.A > 0 && quantizedPaint.LineWidth > 0 {
		r.drawStroke(p, quantizedPaint)
	}
}

// DrawPathWithEffects applies renderer-neutral path effect passes.
func (r *Renderer) DrawPathWithEffects(p geom.Path, paint *render.Paint) bool {
	return render.DrawPathWithEffects(r, r.devPath(p), paint, r.pathDevice)
}

// fillPath fills a path with the given color.
func (r *Renderer) fillPath(p geom.Path, fillColor render.Color) {
	var clipBounds image.Rectangle
	var rasterBounds image.Rectangle
	var offsetX, offsetY float64
	if r.clipRect == nil && len(r.clipPaths) == 0 {
		clipBounds = r.dst.Bounds()
		rasterBounds = clipBounds
	} else {
		pathBounds, ok := pathPixelBounds(p)
		if !ok {
			return
		}
		clipBounds = r.dst.Bounds()
		if r.clipRect != nil {
			clipBounds = clipBounds.Intersect(image.Rect(
				int(math.Floor(r.clipRect.Min.X)),
				int(math.Floor(r.clipRect.Min.Y)),
				int(math.Ceil(r.clipRect.Max.X)),
				int(math.Ceil(r.clipRect.Max.Y)),
			))
		}
		clipBounds = clipBounds.Intersect(pathBounds)
		if clipBounds.Empty() {
			return
		}
		rasterBounds = image.Rect(0, 0, clipBounds.Dx(), clipBounds.Dy())
		offsetX = float64(clipBounds.Min.X)
		offsetY = float64(clipBounds.Min.Y)
	}

	// Reset and rebuild path for filling.
	r.rasterizer.Reset(rasterBounds.Dx(), rasterBounds.Dy())

	vi := 0 // vertex index

	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			// Apply explicit rounding to ensure consistent float32 conversion
			r.rasterizer.MoveTo(float32(math.Round((pt.X-offsetX)*1e6)/1e6), float32(math.Round((pt.Y-offsetY)*1e6)/1e6))
			vi++
		case geom.LineTo:
			pt := p.V[vi]
			r.rasterizer.LineTo(float32(math.Round((pt.X-offsetX)*1e6)/1e6), float32(math.Round((pt.Y-offsetY)*1e6)/1e6))
			vi++
		case geom.QuadTo:
			ctrl := p.V[vi]
			to := p.V[vi+1]
			r.rasterizer.QuadTo(
				float32(math.Round((ctrl.X-offsetX)*1e6)/1e6), float32(math.Round((ctrl.Y-offsetY)*1e6)/1e6),
				float32(math.Round((to.X-offsetX)*1e6)/1e6), float32(math.Round((to.Y-offsetY)*1e6)/1e6),
			)
			vi += 2
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			to := p.V[vi+2]
			r.rasterizer.CubeTo(
				float32(math.Round((c1.X-offsetX)*1e6)/1e6), float32(math.Round((c1.Y-offsetY)*1e6)/1e6),
				float32(math.Round((c2.X-offsetX)*1e6)/1e6), float32(math.Round((c2.Y-offsetY)*1e6)/1e6),
				float32(math.Round((to.X-offsetX)*1e6)/1e6), float32(math.Round((to.Y-offsetY)*1e6)/1e6),
			)
			vi += 3
		case geom.ClosePath:
			r.rasterizer.ClosePath()
		}
	}

	// Draw the filled path using premultiplied alpha
	red, green, blue, alpha := fillColor.ToPremultipliedRGBA()
	c := color.RGBA{R: red, G: green, B: blue, A: alpha}

	if r.clipRect == nil && len(r.clipPaths) == 0 {
		r.rasterizer.Draw(r.dst, rasterBounds, image.NewUniform(c), image.Point{})
		return
	}

	if len(r.clipPaths) == 0 {
		// Use a zero-origin mask for the clipped path, then draw that mask directly
		// into the matching destination rectangle.
		r.rasterizer.Draw(r.dst, clipBounds, image.NewUniform(c), image.Point{})
		return
	}

	mask := image.NewAlpha(rasterBounds)
	r.rasterizer.Draw(mask, rasterBounds, image.NewUniform(color.Alpha{A: 255}), image.Point{})
	for y := clipBounds.Min.Y; y < clipBounds.Max.Y; y++ {
		for x := clipBounds.Min.X; x < clipBounds.Max.X; x++ {
			local := mask.PixOffset(x-clipBounds.Min.X, y-clipBounds.Min.Y)
			if local < 0 || local >= len(mask.Pix) {
				continue
			}
			coverage := mask.Pix[local]
			if coverage == 0 {
				continue
			}
			clipA := r.clipMaskAlphaAt(x, y)
			if clipA == 0 {
				continue
			}
			src := renderColorToRGBA(fillColor)
			src.A = uint8(uint32(src.A) * uint32(coverage) * uint32(clipA) / (255 * 255))
			if src.A == 0 {
				continue
			}
			r.blendPixelNoClip(x, y, src)
		}
	}
}

func pathPixelBounds(p geom.Path) (image.Rectangle, bool) {
	if len(p.V) == 0 {
		return image.Rectangle{}, false
	}

	minX, maxX := p.V[0].X, p.V[0].X
	minY, maxY := p.V[0].Y, p.V[0].Y
	for _, pt := range p.V[1:] {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}

	return image.Rect(
		int(math.Floor(minX))-1,
		int(math.Floor(minY))-1,
		int(math.Ceil(maxX))+1,
		int(math.Ceil(maxY))+1,
	), true
}

// drawStroke handles stroke drawing for paths using proper stroke geometry.
func (r *Renderer) drawStroke(p geom.Path, paint *render.Paint) {
	// Convert stroke to filled path with proper joins, caps, and dashes
	strokePath := strokeToPath(p, paint)
	if len(strokePath.C) == 0 {
		return // No stroke geometry generated
	}

	// Fill the stroke geometry with the stroke color
	r.fillPath(strokePath, paint.Stroke)
}

// Image draws an image within the destination rectangle.
func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	if img == nil {
		return
	}

	src := asRGBAImage(img)
	if src == nil {
		return
	}

	// dst arrives in y-up display space; flip to the y-down device buffer.
	// drawBitmapScaledWithAlpha places image row 0 at the device-top row
	// (dst.Min.Y after the flip), keeping the image upright.
	dst = r.devRect(dst)

	// Destination rectangle in integer coordinates.
	minX := int(math.Floor(dst.Min.X))
	minY := int(math.Floor(dst.Min.Y))
	maxX := int(math.Ceil(dst.Max.X))
	maxY := int(math.Ceil(dst.Max.Y))
	if maxX <= minX || maxY <= minY {
		return
	}

	r.drawBitmapScaledWithAlpha(src, minX, minY, maxX-minX, maxY-minY, imageAlphaMultiplier(img))
}

// ImageTransformed draws a raster image through an arbitrary affine transform.
// The affine maps source image pixels into display coordinates.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, transform geom.Affine) {
	if img == nil {
		return
	}
	src := asRGBAImage(img)
	if src == nil {
		return
	}
	// transform maps image space -> y-up display. Compose the device y-flip so
	// the final affine maps image space -> y-down device. Mul is m∘n (apply n
	// then m), so deviceFlipAffine().Mul(transform)(p) = F(transform(p)); image
	// rows stay upright. This matches core/image.go's imageTransform convention.
	transform = r.deviceFlipAffine().Mul(transform)
	r.drawBitmapTransformed(src, transform, imageAlphaMultiplier(img))
}

// GlyphRun renders glyph IDs as code points where available.
// The mapping is a practical fallback for renderers that expose only glyph IDs.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if len(run.Glyphs) == 0 {
		return
	}
	penX := run.Origin.X + run.Glyphs[0].Offset.X
	penY := run.Origin.Y + run.Glyphs[0].Offset.Y
	size := run.Size
	if size <= 0 {
		size = 12
	}

	for _, glyph := range run.Glyphs {
		advance := glyph.Advance
		ch := rune(glyph.ID)
		if ch > 0 {
			_ = r.MeasureText(string(ch), size, run.FontKey)
			r.DrawText(string(ch), geom.Pt{
				X: penX + glyph.Offset.X,
				Y: penY + glyph.Offset.Y,
			}, size, textColor)

			if advance <= 0 {
				advance = r.MeasureText(string(ch), size, run.FontKey).W
			}
		}
		penX += glyph.Offset.X + advance
	}
}

// MeasureText returns approximate text metrics for layout.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" || size <= 0 {
		return render.TextMetrics{}
	}
	if fontKey == "" {
		fontKey = "DejaVuSans"
	}
	r.lastFontKey = fontKey

	return measureText(text, size, fontKey, r.resolution)
}

// TextPath converts text to a vector path using the shared font manager.
func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if fontKey == "" {
		fontKey = r.lastFontKey
	}
	return render.TextPath(text, origin, size, fontKey)
}

// GetImage returns the underlying image.RGBA for PNG export.
func (r *Renderer) GetImage() *image.RGBA {
	return r.dst
}

// SavePNG saves the rendered image to a PNG file.
func (r *Renderer) SavePNG(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, r.dst)
}

// DrawText renders text at the requested origin.
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	if text == "" || size <= 0 {
		return
	}

	metrics := r.MeasureText(text, size, r.lastFontKey)
	if metrics.W <= 0 || metrics.H <= 0 {
		return
	}

	src := r.renderTextBitmap(text, size, textColor, r.lastFontKey)
	if src == nil {
		return
	}

	// origin.Y is the y-up display baseline; flip it to the device baseline and
	// place the upright bitmap with its top ascent above that baseline.
	x := int(math.Round(origin.X))
	y := int(math.Round(r.devY(origin.Y) - metrics.Ascent))
	r.drawBitmapScaled(src, x, y, src.Bounds().Dx(), src.Bounds().Dy())
}

// DrawTextRotated renders text using Matplotlib-like anchor rotation. The
// anchor is the bottom-center of the unrotated text box.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	if text == "" || size <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return
	}

	src := r.renderTextBitmap(text, size, textColor, r.lastFontKey)
	if src == nil {
		return
	}

	const epsilon = 1e-12
	pivotX := float64(src.Bounds().Dx()) / 2
	pivotY := float64(src.Bounds().Dy())
	// anchor is the bottom-center of the unrotated box in y-up display space;
	// flip it to the device buffer. The bitmap pivot (bottom-center) is placed at
	// the device anchor. The rotation sign stays -angle so a positive display
	// (CCW) rotation still renders CCW visually, matching the pre-y-flip behavior.
	devAnchor := r.devPt(anchor)
	if math.Abs(angle) <= epsilon {
		x := int(math.Round(devAnchor.X - pivotX))
		y := int(math.Round(devAnchor.Y - pivotY))
		r.drawBitmapScaled(src, x, y, src.Bounds().Dx(), src.Bounds().Dy())
		return
	}

	r.drawBitmapRotated(src, devAnchor, geom.Pt{X: pivotX, Y: pivotY}, -angle)
}

// DrawTextVertical renders one character per line.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	runes := []rune(text)
	if len(runes) == 0 || size <= 0 {
		return
	}

	lineHeight := r.MeasureText("M", size, r.lastFontKey).H
	if lineHeight <= 0 {
		return
	}

	totalHeight := lineHeight * float64(len(runes))
	// DrawText flips each baseline to the device buffer, so the per-character
	// layout is expressed in y-up display space: start at the top of the stack
	// (highest Y) and step downward by lineHeight per character.
	y := center.Y + totalHeight/2
	for i, ch := range runes {
		chMetrics := r.MeasureText(string(ch), size, r.lastFontKey)
		if chMetrics.W <= 0 || chMetrics.H <= 0 {
			continue
		}

		x := center.X - chMetrics.W/2
		r.DrawText(string(ch), geom.Pt{
			X: x,
			Y: y - float64(i)*lineHeight - chMetrics.Ascent,
		}, size, textColor)
	}
}

// SetResolution sets the text rendering resolution used for point-sized fonts.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi > 0 {
		r.resolution = dpi
	}
}

func (r *Renderer) renderTextBitmap(text string, size float64, textColor render.Color, fontKey string) *image.RGBA {
	if text == "" || size <= 0 {
		return nil
	}
	if fontKey == "" {
		fontKey = "DejaVuSans"
	}
	return renderTextBitmap(text, size, textColor, fontKey, r.resolution)
}

func (r *Renderer) drawBitmapScaled(src *image.RGBA, dstX, dstY, dstW, dstH int) {
	r.drawBitmapScaledWithAlpha(src, dstX, dstY, dstW, dstH, 1)
}

func (r *Renderer) drawBitmapScaledWithAlpha(src *image.RGBA, dstX, dstY, dstW, dstH int, alpha float64) {
	if src == nil || dstW <= 0 || dstH <= 0 {
		return
	}
	if alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}

	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	if srcW <= 0 || srcH <= 0 {
		return
	}

	dst, ok := r.drawTargetRect(dstX, dstY, dstX+dstW, dstY+dstH)
	if !ok {
		return
	}

	srcMin := src.Bounds().Min
	for y := dst.Min.Y; y < dst.Max.Y; y++ {
		sy := int(math.Floor(float64(y-dstY) * float64(srcH) / float64(dstH)))
		if sy < 0 {
			sy = 0
		}
		if sy >= srcH {
			sy = srcH - 1
		}

		srcIdxBase := src.PixOffset(srcMin.X, srcMin.Y+sy)
		srcRow := src.Pix[srcIdxBase : srcIdxBase+srcW*4]
		for x := dst.Min.X; x < dst.Max.X; x++ {
			sx := int(math.Floor(float64(x-dstX) * float64(srcW) / float64(dstW)))
			if sx < 0 {
				sx = 0
			}
			if sx >= srcW {
				sx = srcW - 1
			}

			srcOffset := sx * 4
			srcColor := color.RGBA{
				R: srcRow[srcOffset],
				G: srcRow[srcOffset+1],
				B: srcRow[srcOffset+2],
				A: srcRow[srcOffset+3],
			}
			if alpha < 1 {
				srcColor.A = uint8(math.Round(float64(srcColor.A) * alpha))
			}
			r.blendPixel(x, y, srcColor)
		}
	}
}

func (r *Renderer) drawBitmapTransformed(src *image.RGBA, transform geom.Affine, alpha float64) {
	if src == nil || alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}

	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	if srcW <= 0 || srcH <= 0 {
		return
	}
	inv, ok := transform.Invert()
	if !ok {
		return
	}

	corners := []geom.Pt{
		transform.Apply(geom.Pt{X: 0, Y: 0}),
		transform.Apply(geom.Pt{X: float64(srcW), Y: 0}),
		transform.Apply(geom.Pt{X: float64(srcW), Y: float64(srcH)}),
		transform.Apply(geom.Pt{X: 0, Y: float64(srcH)}),
	}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, corner := range corners {
		if corner.X < minX {
			minX = corner.X
		}
		if corner.X > maxX {
			maxX = corner.X
		}
		if corner.Y < minY {
			minY = corner.Y
		}
		if corner.Y > maxY {
			maxY = corner.Y
		}
	}
	if math.IsNaN(minX) || math.IsNaN(minY) || math.IsNaN(maxX) || math.IsNaN(maxY) ||
		math.IsInf(minX, 0) || math.IsInf(minY, 0) || math.IsInf(maxX, 0) || math.IsInf(maxY, 0) {
		return
	}

	dst, ok := r.drawTargetRect(
		int(math.Floor(minX)),
		int(math.Floor(minY)),
		int(math.Ceil(maxX)),
		int(math.Ceil(maxY)),
	)
	if !ok {
		return
	}

	srcMin := src.Bounds().Min
	for y := dst.Min.Y; y < dst.Max.Y; y++ {
		for x := dst.Min.X; x < dst.Max.X; x++ {
			sp := inv.Apply(geom.Pt{X: float64(x) + 0.5, Y: float64(y) + 0.5})
			sx := int(math.Floor(sp.X))
			sy := int(math.Floor(sp.Y))
			if sx < 0 || sy < 0 || sx >= srcW || sy >= srcH {
				continue
			}
			p := src.PixOffset(srcMin.X+sx, srcMin.Y+sy)
			srcColor := color.RGBA{
				R: src.Pix[p],
				G: src.Pix[p+1],
				B: src.Pix[p+2],
				A: src.Pix[p+3],
			}
			if alpha < 1 {
				srcColor.A = uint8(math.Round(float64(srcColor.A) * alpha))
			}
			r.blendPixel(x, y, srcColor)
		}
	}
}

func (r *Renderer) drawBitmapRotated(src *image.RGBA, anchor, pivot geom.Pt, angle float64) {
	if src == nil {
		return
	}

	srcW := float64(src.Bounds().Dx())
	srcH := float64(src.Bounds().Dy())
	if srcW <= 0 || srcH <= 0 {
		return
	}

	cos, sin := math.Cos(angle), math.Sin(angle)

	corners := [4]struct{ x, y float64 }{
		{-pivot.X, -pivot.Y},
		{srcW - pivot.X, -pivot.Y},
		{srcW - pivot.X, srcH - pivot.Y},
		{-pivot.X, srcH - pivot.Y},
	}

	minX := math.Inf(1)
	maxX := math.Inf(-1)
	minY := math.Inf(1)
	maxY := math.Inf(-1)
	for _, corner := range corners {
		rx := corner.x*cos - corner.y*sin
		ry := corner.x*sin + corner.y*cos
		if rx < minX {
			minX = rx
		}
		if rx > maxX {
			maxX = rx
		}
		if ry < minY {
			minY = ry
		}
		if ry > maxY {
			maxY = ry
		}
	}

	boundsW := int(math.Ceil(maxX - minX))
	boundsH := int(math.Ceil(maxY - minY))
	if boundsW <= 0 || boundsH <= 0 {
		return
	}

	minXInt := int(math.Floor(anchor.X + minX))
	minYInt := int(math.Floor(anchor.Y + minY))
	drawBounds, ok := r.drawTargetRect(minXInt, minYInt, minXInt+boundsW, minYInt+boundsH)
	if !ok {
		return
	}

	srcMin := src.Bounds().Min
	for y := drawBounds.Min.Y; y < drawBounds.Max.Y; y++ {
		for x := drawBounds.Min.X; x < drawBounds.Max.X; x++ {
			localX := float64(x) - anchor.X
			localY := float64(y) - anchor.Y

			// Inverse rotation from destination to source coordinates.
			sxF := localX*cos + localY*sin
			syF := -localX*sin + localY*cos

			srcX := int(math.Round(sxF + pivot.X - 0.5))
			srcY := int(math.Round(syF + pivot.Y - 0.5))
			if srcX < 0 || srcY < 0 || srcX >= int(srcW) || srcY >= int(srcH) {
				continue
			}

			p := src.PixOffset(srcMin.X+srcX, srcMin.Y+srcY)
			r.blendPixel(x, y, color.RGBA{
				R: src.Pix[p],
				G: src.Pix[p+1],
				B: src.Pix[p+2],
				A: src.Pix[p+3],
			})
		}
	}
}

func (r *Renderer) drawTargetRect(minX, minY, maxX, maxY int) (image.Rectangle, bool) {
	if minX >= maxX || minY >= maxY {
		return image.Rectangle{}, false
	}

	bounds := image.Rect(minX, minY, maxX, maxY).Intersect(r.dst.Bounds())
	if bounds.Empty() {
		return image.Rectangle{}, false
	}

	if r.clipRect != nil {
		clipBounds := image.Rect(
			int(math.Floor(r.clipRect.Min.X)),
			int(math.Floor(r.clipRect.Min.Y)),
			int(math.Ceil(r.clipRect.Max.X)),
			int(math.Ceil(r.clipRect.Max.Y)),
		)
		bounds = bounds.Intersect(clipBounds)
	}

	if bounds.Empty() {
		return image.Rectangle{}, false
	}

	return bounds, true
}

func (r *Renderer) blendPixel(x, y int, src color.RGBA) {
	if src.A == 0 {
		return
	}
	if len(r.clipPaths) > 0 {
		clipA := r.clipMaskAlphaAt(x, y)
		if clipA == 0 {
			return
		}
		src.A = uint8(uint16(src.A) * uint16(clipA) / 255)
		if src.A == 0 {
			return
		}
	}
	r.blendPixelNoClip(x, y, src)
}

func (r *Renderer) blendPixelNoClip(x, y int, src color.RGBA) {
	i := r.dst.PixOffset(x, y)
	dr := uint32(r.dst.Pix[i])
	dg := uint32(r.dst.Pix[i+1])
	db := uint32(r.dst.Pix[i+2])
	da := uint32(r.dst.Pix[i+3])

	sr := uint32(src.R)
	sg := uint32(src.G)
	sb := uint32(src.B)
	sa := uint32(src.A)

	if sa == 255 {
		r.dst.Pix[i] = src.R
		r.dst.Pix[i+1] = src.G
		r.dst.Pix[i+2] = src.B
		r.dst.Pix[i+3] = src.A
		return
	}

	outA := sa + ((255 - sa) * da / 255)
	if outA == 0 {
		r.dst.Pix[i] = 0
		r.dst.Pix[i+1] = 0
		r.dst.Pix[i+2] = 0
		r.dst.Pix[i+3] = 0
		return
	}

	r.dst.Pix[i] = uint8((sr*sa + dr*(255-sa)*da/255) / outA)
	r.dst.Pix[i+1] = uint8((sg*sa + dg*(255-sa)*da/255) / outA)
	r.dst.Pix[i+2] = uint8((sb*sa + db*(255-sa)*da/255) / outA)
	r.dst.Pix[i+3] = uint8(outA)
}

func (r *Renderer) clipMaskAlphaAt(x, y int) uint8 {
	if len(r.clipPaths) == 0 {
		return 255
	}
	if x < 0 || y < 0 || x >= r.dst.Bounds().Dx() || y >= r.dst.Bounds().Dy() {
		return 0
	}
	alpha := 255
	for _, path := range r.clipPaths {
		mask := r.clipMaskForPath(path)
		if mask == nil {
			return 0
		}
		i := mask.PixOffset(x, y)
		if i < 0 || i >= len(mask.Pix) {
			return 0
		}
		alpha = alpha * int(mask.Pix[i]) / 255
		if alpha == 0 {
			return 0
		}
	}
	return uint8(alpha)
}

func (r *Renderer) clipMaskForPath(path geom.Path) *image.Alpha {
	if len(path.C) == 0 || !path.Validate() || r.dst == nil {
		return nil
	}
	b := r.dst.Bounds()
	key := clipMaskKey{
		width:  b.Dx(),
		height: b.Dy(),
		hash:   hashPath(path),
	}
	if r.clipMaskMap == nil {
		r.clipMaskMap = make(map[clipMaskKey]*image.Alpha)
	}
	if mask, ok := r.clipMaskMap[key]; ok {
		return mask
	}

	mask := image.NewAlpha(b)
	ras := vector.NewRasterizer(b.Dx(), b.Dy())
	appendPathToRasterizer(ras, path, 0, 0)
	ras.Draw(mask, b, image.NewUniform(color.Alpha{A: 255}), image.Point{})
	r.clipMaskMap[key] = mask
	return mask
}

func appendPathToRasterizer(ras *vector.Rasterizer, p geom.Path, offsetX, offsetY float64) {
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			ras.MoveTo(rasterCoord(pt.X-offsetX), rasterCoord(pt.Y-offsetY))
			vi++
		case geom.LineTo:
			pt := p.V[vi]
			ras.LineTo(rasterCoord(pt.X-offsetX), rasterCoord(pt.Y-offsetY))
			vi++
		case geom.QuadTo:
			ctrl := p.V[vi]
			to := p.V[vi+1]
			ras.QuadTo(
				rasterCoord(ctrl.X-offsetX), rasterCoord(ctrl.Y-offsetY),
				rasterCoord(to.X-offsetX), rasterCoord(to.Y-offsetY),
			)
			vi += 2
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			to := p.V[vi+2]
			ras.CubeTo(
				rasterCoord(c1.X-offsetX), rasterCoord(c1.Y-offsetY),
				rasterCoord(c2.X-offsetX), rasterCoord(c2.Y-offsetY),
				rasterCoord(to.X-offsetX), rasterCoord(to.Y-offsetY),
			)
			vi += 3
		case geom.ClosePath:
			ras.ClosePath()
		}
	}
}

func rasterCoord(v float64) float32 {
	return float32(math.Round(v*1e6) / 1e6)
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

func asRGBAImage(img render.Image) *image.RGBA {
	rgbaImage, ok := img.(interface {
		RGBA() *image.RGBA
	})
	if ok {
		return rgbaImage.RGBA()
	}
	return nil
}

func imageAlphaMultiplier(img render.Image) float64 {
	alphaImage, ok := img.(render.ImageAlpha)
	if !ok {
		return 1
	}
	alpha := alphaImage.Alpha()
	if alpha <= 0 {
		return 0
	}
	if alpha >= 1 {
		return 1
	}
	return alpha
}

func renderColorToRGBA(c render.Color) color.RGBA {
	return color.RGBA{
		R: toByte(c.R),
		G: toByte(c.G),
		B: toByte(c.B),
		A: toByte(c.A),
	}
}

func toByte(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

func scaleImageNearest(src *image.RGBA, scaleX, scaleY float64) *image.RGBA {
	if src == nil || scaleX <= 0 || scaleY <= 0 {
		return nil
	}

	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	dstW := int(math.Ceil(float64(srcW) * scaleX))
	dstH := int(math.Ceil(float64(srcH) * scaleY))
	if dstW <= 0 || dstH <= 0 {
		return nil
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		srcY := int(math.Round((float64(y)+0.5)/scaleY - 0.5))
		if srcY < 0 {
			srcY = 0
		}
		if srcY >= srcH {
			srcY = srcH - 1
		}
		srcRow := src.Pix[src.PixOffset(src.Bounds().Min.X, src.Bounds().Min.Y+srcY):]
		for x := 0; x < dstW; x++ {
			srcX := int(math.Round((float64(x)+0.5)/scaleX - 0.5))
			if srcX < 0 {
				srcX = 0
			}
			if srcX >= srcW {
				srcX = srcW - 1
			}

			srcOffset := srcRow[srcX*4:]
			dstOffset := dst.PixOffset(x, y)
			dst.Pix[dstOffset] = srcOffset[0]
			dst.Pix[dstOffset+1] = srcOffset[1]
			dst.Pix[dstOffset+2] = srcOffset[2]
			dst.Pix[dstOffset+3] = srcOffset[3]
		}
	}

	return dst
}

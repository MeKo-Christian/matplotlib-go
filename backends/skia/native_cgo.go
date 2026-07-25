//go:build skia && skiacgo

// Package skia native cgo bridge.
//
// This file is compiled only under `-tags "skia skiacgo"`. It binds the narrow
// C ABI in skia_cwrap.h (implemented by skia_cwrap.cpp, compiled by cgo from
// this package directory) and provides a surfaceBridge backed by a real Skia
// raster surface. The Skia include/library locations are supplied at build time
// through CGO_CXXFLAGS / CGO_LDFLAGS (see the `skia-cgo-*` just recipes) so they
// stay out of the source tree; only the C++ standard library and the wrapper's
// own flags live in the cgo directives below.
package skia

// #cgo CXXFLAGS: -std=c++17
// #cgo LDFLAGS: -lstdc++
// #include <stdlib.h>
// #include "skia_cwrap.h"
import "C"

import (
	"image"
	"image/color"
	"runtime"
	"sort"
	"strings"
	"unsafe"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

var (
	_ surfaceBridge     = (*nativeSurfaceBridge)(nil)
	_ nativeBatchBridge = (*nativeSurfaceBridge)(nil)
	_ nativeHatchBridge = (*nativeSurfaceBridge)(nil)
)

// selectSurfaceBridge returns the native Skia bridge under the skiacgo build
// tag. The CPU bridge is retained inside it as the fallback for paths the
// native wrapper does not yet handle (transformed/pattern fills).
func selectSurfaceBridge(width, height int, mode RenderMode) surfaceBridge {
	return selectSurfaceBridgeWithSamples(width, height, mode, 1)
}

func selectSurfaceBridgeWithSamples(width, height int, mode RenderMode, sampleCount int) surfaceBridge {
	if mode == "" {
		mode = ModeCPU
	}
	bridge := &nativeSurfaceBridge{
		width:       width,
		height:      height,
		mode:        mode,
		sampleCount: sampleCount,
		cpu:         newCPUSurfaceBridge(width, height, mode),
	}
	if mode == ModeGPU {
		bridge.gpu = newNativeGPUSurface(width, height, sampleCount)
		if bridge.gpu != nil {
			runtime.SetFinalizer(bridge, (*nativeSurfaceBridge).finalize)
		}
	}
	return bridge
}

// nativeSkiaVersion returns the linked Skia milestone string. Exposed for the
// build/guard tests so they can confirm the native library is actually linked.
func nativeSkiaVersion() string {
	return C.GoString(C.mgsk_version())
}

type nativeSurfaceBridge struct {
	width  int
	height int
	mode   RenderMode
	cpu    surfaceBridge
	gpu    *nativeSurface

	sampleCount int
}

func (b *nativeSurfaceBridge) Info() BridgeInfo {
	accelerated := b != nil && b.gpu != nil && b.gpu.isGPU()
	description := "native Skia raster surface via cgo C-ABI wrapper"
	if b != nil && b.mode == ModeGPU {
		if accelerated {
			description = "native Skia Ganesh GPU surface via EGL/OpenGL with deterministic CPU readback"
		} else {
			description = "Skia GPU mode unavailable at runtime; native CPU raster fallback"
		}
	}
	return BridgeInfo{
		Binding:         BindingExternalCAPI,
		Mode:            b.mode,
		NativeSurface:   true,
		Accelerated:     accelerated,
		SupportsShaders: true,
		Description:     description,
	}
}

func (b *nativeSurfaceBridge) SupportsShaders() bool { return true }

func (b *nativeSurfaceBridge) FlushGPU() bool {
	return b != nil && b.gpu != nil && b.gpu.flushGPU()
}

func (b *nativeSurfaceBridge) finalize() {
	if b == nil || b.gpu == nil {
		return
	}
	b.gpu.delete()
	b.gpu = nil
}

func (b *nativeSurfaceBridge) acquireSurface(w, h int) (*nativeSurface, func()) {
	if b != nil && b.gpu != nil && b.gpu.w == w && b.gpu.h == h {
		b.gpu.clear()
		return b.gpu, func() {}
	}
	if b != nil && b.mode == ModeGPU {
		if surf := newNativeGPUSurface(w, h, b.sampleCount); surf != nil {
			surf.clear()
			return surf, surf.delete
		}
	}
	surf := newNativeSurface(w, h)
	if surf == nil {
		return nil, func() {}
	}
	return surf, surf.delete
}

// DrawPathFill rasterizes gradient fills through a native Skia surface and
// composites the result over dst. Pattern fills and transformed gradients are
// delegated to the CPU bridge, which already handles them deterministically.
func (b *nativeSurfaceBridge) DrawPathFill(dst *image.RGBA, path geom.Path, paint render.Paint, state bridgeDrawState) bool {
	if dst == nil || !path.Validate() {
		return false
	}
	grad := paint.FillGradient
	if grad.Kind == render.GradientNone || len(grad.Stops) == 0 || grad.HasTransform {
		// Not a (plain) gradient fill — let the CPU bridge handle pattern fills
		// and transformed gradients.
		return b.cpu.DrawPathFill(dst, path, paint, state)
	}

	bounds := shaderDrawBounds(dst.Bounds(), path, state.clipRect)
	if bounds.Empty() {
		return true
	}
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()

	surf, release := b.acquireSurface(w, h)
	if surf == nil {
		return b.cpu.DrawPathFill(dst, path, paint, state)
	}
	defer release()

	verbs, coords := pathToVerbsCoords(path)
	cp, freePaint := newGradientCPaint(grad, float64(h))
	defer freePaint()

	C.mgsk_draw_path(surf.ptr,
		bytePtr(verbs), C.int(len(verbs)),
		floatPtr(coords), C.int(len(coords)),
		cp)

	rendered := surf.readImage()
	clipMasks := rasterizeClipMasks(w, h, state.clipPaths)
	compositeNativeOver(dst, rendered, bounds, clipMasks)
	return true
}

// drawMarkersNative renders a marker batch through one native Skia surface and
// composites it over dst. It handles solid fill/stroke markers with a uniform
// edge width; batches with gradient/pattern/hatch paints or non-uniform edge
// widths return false so the caller falls back to the per-item CPU loop.
func (b *nativeSurfaceBridge) drawMarkersNative(dst *image.RGBA, batch render.MarkerBatch, state bridgeDrawState, height float64) bool {
	if dst == nil || len(batch.Marker.C) == 0 || len(batch.Items) == 0 {
		return false
	}
	lineWidth := batch.Items[0].Paint.LineWidth
	antialias := true
	matrices := make([]float32, 0, len(batch.Items)*6)
	fillColors := make([]float32, 0, len(batch.Items)*4)
	strokeColors := make([]float32, 0, len(batch.Items)*4)
	for i := range batch.Items {
		item := &batch.Items[i]
		p := item.Paint
		if p.FillGradient.Kind != render.GradientNone || hasPatternFill(p.FillPattern) || p.Hatch != "" {
			return false
		}
		if p.Stroke.A > 0 && p.LineWidth > 0 && p.LineWidth != lineWidth {
			return false
		}
		if !item.Antialiased || p.Antialias == render.AntialiasOff {
			antialias = false
		}
		t := item.Transform
		// Compose item affine + offset, then flip to device space (y -> H - y):
		//   Xdev =  A*x + C*y + (E+offX)
		//   Ydev = -B*x - D*y + (H - F - offY)
		matrices = append(
			matrices,
			float32(t.A), float32(t.C), float32(t.E+item.Offset.X),
			float32(-t.B), float32(-t.D), float32(height-t.F-item.Offset.Y),
		)
		fillColors = appendColor(fillColors, p.Fill)
		strokeColors = appendColor(strokeColors, p.Stroke)
	}

	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	flippedState := bridgeDrawState{
		clipRect:  flipRectPtrY(state.clipRect, height),
		clipPaths: flipPathsY(state.clipPaths, height),
	}
	surf, release := b.acquireSurface(w, h)
	if surf == nil {
		return false
	}
	defer release()

	verbs, coords := pathToVerbsCoords(batch.Marker)
	C.mgsk_draw_markers(
		surf.ptr,
		bytePtr(verbs), C.int(len(verbs)),
		floatPtr(coords), C.int(len(coords)),
		floatPtr(matrices),
		floatPtr(fillColors),
		floatPtr(strokeColors),
		C.float(lineWidth),
		boolToCInt(antialias),
		C.int(len(batch.Items)),
	)

	rendered := surf.readImage()
	bounds := dst.Bounds()
	if flippedState.clipRect != nil {
		bounds = bounds.Intersect(rectToImage(*flippedState.clipRect))
	}
	clipMasks := rasterizeClipMasks(w, h, flippedState.clipPaths)
	compositeNativeOver(dst, rendered, bounds, clipMasks)
	return true
}

// drawPathCollectionNative renders solid fill/stroke path-collection items
// through one native Skia surface and composites the result over dst. It
// declines richer paint features so DrawPathCollection can keep using the
// existing renderer-neutral fallbacks for those items.
func (b *nativeSurfaceBridge) drawPathCollectionNative(dst *image.RGBA, batch render.PathCollectionBatch, state bridgeDrawState) bool {
	if dst == nil || len(batch.Items) == 0 {
		return false
	}
	for i := range batch.Items {
		item := &batch.Items[i]
		if len(item.Path.C) == 0 {
			continue
		}
		if !item.Path.Validate() || !canDrawPathCollectionItemNative(item) {
			return false
		}
	}

	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	height := float64(h)
	flippedState := bridgeDrawState{
		clipRect:  flipRectPtrY(state.clipRect, height),
		clipPaths: flipPathsY(state.clipPaths, height),
	}
	surf, release := b.acquireSurface(w, h)
	if surf == nil {
		return false
	}
	defer release()

	drew := false
	for i := range batch.Items {
		item := &batch.Items[i]
		if len(item.Path.C) == 0 {
			continue
		}
		paint := item.Paint
		applyForcedAlpha(&paint)
		if !item.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		cp := newSolidCPaint(paint)
		verbs, coords := pathToVerbsCoords(flipPathY(item.Path, height))
		C.mgsk_draw_path(surf.ptr,
			bytePtr(verbs), C.int(len(verbs)),
			floatPtr(coords), C.int(len(coords)),
			&cp)
		drew = true
	}
	if !drew {
		return true
	}

	rendered := surf.readImage()
	bounds := dst.Bounds()
	if flippedState.clipRect != nil {
		bounds = bounds.Intersect(rectToImage(*flippedState.clipRect))
	}
	clipMasks := rasterizeClipMasks(w, h, flippedState.clipPaths)
	compositeNativeOver(dst, rendered, bounds, clipMasks)
	return true
}

// drawQuadMeshNative renders face-only quadrilateral cells as two native
// SkVertices triangles per cell. Cells with edges, dashes, or hatches are
// declined so DrawQuadMesh can preserve the existing CPU fallback behavior.
func (b *nativeSurfaceBridge) drawQuadMeshNative(dst *image.RGBA, batch render.QuadMeshBatch, state bridgeDrawState) bool {
	if dst == nil || len(batch.Cells) == 0 {
		return false
	}
	positions := make([]float32, 0, len(batch.Cells)*12)
	colors := make([]float32, 0, len(batch.Cells)*24)
	antialias := true
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	height := float64(h)
	for i := range batch.Cells {
		cell := &batch.Cells[i]
		if !canDrawQuadMeshCellNative(cell) {
			return false
		}
		if cell.Face.A <= 0 {
			continue
		}
		if !cell.Antialiased {
			antialias = false
		}
		p0 := flipPointY(cell.Quad[0], height)
		p1 := flipPointY(cell.Quad[1], height)
		p2 := flipPointY(cell.Quad[2], height)
		p3 := flipPointY(cell.Quad[3], height)
		for _, p := range []geom.Pt{p0, p1, p2, p0, p2, p3} {
			positions = append(positions, float32(p.X), float32(p.Y))
			colors = appendColor(colors, cell.Face)
		}
	}
	if len(positions) == 0 {
		return true
	}

	flippedState := bridgeDrawState{
		clipRect:  flipRectPtrY(state.clipRect, height),
		clipPaths: flipPathsY(state.clipPaths, height),
	}
	surf, release := b.acquireSurface(w, h)
	if surf == nil {
		return false
	}
	defer release()

	C.mgsk_draw_vertices(
		surf.ptr,
		floatPtr(positions),
		floatPtr(colors),
		C.int(len(positions)/6),
		boolToCInt(antialias),
	)

	rendered := surf.readImage()
	bounds := dst.Bounds()
	if flippedState.clipRect != nil {
		bounds = bounds.Intersect(rectToImage(*flippedState.clipRect))
	}
	clipMasks := rasterizeClipMasks(w, h, flippedState.clipPaths)
	compositeNativeOver(dst, rendered, bounds, clipMasks)
	return true
}

// drawImageTransformedNative renders an RGBA-backed image through native Skia.
// The affine maps source image pixels into y-up display coordinates; this method
// composes the device y-flip before passing the matrix to Skia's y-down canvas.
func (b *nativeSurfaceBridge) drawImageTransformedNative(dst *image.RGBA, img render.Image, transform geom.Affine, state bridgeDrawState) bool {
	if dst == nil || img == nil {
		return false
	}
	rgba, ok := img.(render.RGBAImage)
	if !ok {
		return false
	}
	src := rgba.RGBA()
	if src == nil {
		return false
	}
	srcBounds := src.Bounds()
	srcW, srcH := srcBounds.Dx(), srcBounds.Dy()
	if srcW <= 0 || srcH <= 0 || src.Stride < srcW*4 || len(src.Pix) == 0 {
		return false
	}

	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	height := float64(h)
	deviceTransform := geom.Affine{A: 1, D: -1, F: geom.F64(height)}.Mul(transform)
	matrix := []float32{
		float32(deviceTransform.A), float32(deviceTransform.C), float32(deviceTransform.E),
		float32(deviceTransform.B), float32(deviceTransform.D), float32(deviceTransform.F),
	}
	alpha := imageAlphaMultiplier(img)
	if alpha <= 0 {
		return true
	}

	surf, release := b.acquireSurface(w, h)
	if surf == nil {
		return false
	}
	defer release()

	pixOffset := src.PixOffset(srcBounds.Min.X, srcBounds.Min.Y)
	if pixOffset < 0 || pixOffset >= len(src.Pix) {
		return false
	}
	C.mgsk_draw_image(
		surf.ptr,
		(*C.uint8_t)(unsafe.Pointer(&src.Pix[pixOffset])),
		C.int(srcW),
		C.int(srcH),
		C.int(src.Stride),
		floatPtr(matrix),
		C.float(alpha),
		C.int(skiaImageSampling(img.Interpolation())),
	)

	flippedState := bridgeDrawState{
		clipRect:  flipRectPtrY(state.clipRect, height),
		clipPaths: flipPathsY(state.clipPaths, height),
	}
	rendered := surf.readImage()
	bounds := dst.Bounds()
	if flippedState.clipRect != nil {
		bounds = bounds.Intersect(rectToImage(*flippedState.clipRect))
	}
	clipMasks := rasterizeClipMasks(w, h, flippedState.clipPaths)
	compositeNativeOver(dst, rendered, bounds, clipMasks)
	return true
}

// drawHatchPathNative fills path-local hatch pixels via a repeated SkShader on
// a native Skia surface, then composites them back into the CPU image.
func (b *nativeSurfaceBridge) drawHatchPathNative(dst *image.RGBA, path geom.Path, paint render.Paint, state bridgeDrawState) bool {
	if dst == nil || paint.Hatch == "" || !path.Validate() {
		return false
	}
	paintCopy := paint
	applyForcedAlpha(&paintCopy)
	color := paintCopy.HatchColor
	if color.A <= 0 {
		return true
	}
	lineWidth := paintCopy.HatchLineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	spacing := paintCopy.HatchSpacing
	if spacing <= 0 {
		spacing = render.DefaultHatchSpacing
	}

	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	height := float64(h)
	surf, release := b.acquireSurface(w, h)
	if surf == nil {
		return false
	}
	defer release()

	flippedPath := flipPathY(path, height)
	verbs, coords := pathToVerbsCoords(flippedPath)
	hatch := C.CString(paintCopy.Hatch)
	defer C.free(unsafe.Pointer(hatch))
	C.mgsk_draw_hatch_path(
		surf.ptr,
		bytePtr(verbs), C.int(len(verbs)),
		floatPtr(coords), C.int(len(coords)),
		hatch, C.int(len(paintCopy.Hatch)),
		C.float(color.R), C.float(color.G), C.float(color.B), C.float(color.A),
		C.float(lineWidth),
		C.float(spacing),
		boolToCInt(paintCopy.Antialias != render.AntialiasOff),
	)

	flippedState := bridgeDrawState{
		clipRect:  flipRectPtrY(state.clipRect, height),
		clipPaths: flipPathsY(state.clipPaths, height),
	}
	rendered := surf.readImage()
	bounds := dst.Bounds()
	if flippedState.clipRect != nil {
		bounds = bounds.Intersect(rectToImage(*flippedState.clipRect))
	}
	clipMasks := rasterizeClipMasks(w, h, flippedState.clipPaths)
	compositeNativeOver(dst, rendered, bounds, clipMasks)
	return true
}

// drawGouraudNative renders interpolated-color triangles via SkVertices through
// a native Skia surface and composites the result over dst.
func (b *nativeSurfaceBridge) drawGouraudNative(dst *image.RGBA, batch render.GouraudTriangleBatch, state bridgeDrawState) bool {
	if dst == nil || len(batch.Triangles) == 0 {
		return false
	}
	positions := make([]float32, 0, len(batch.Triangles)*6)
	colors := make([]float32, 0, len(batch.Triangles)*12)
	for i := range batch.Triangles {
		tri := &batch.Triangles[i]
		for v := 0; v < 3; v++ {
			positions = append(positions, float32(tri.P[v].X), float32(tri.P[v].Y))
			colors = appendColor(colors, tri.Color[v])
		}
	}

	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	surf, release := b.acquireSurface(w, h)
	if surf == nil {
		return false
	}
	defer release()

	C.mgsk_draw_vertices(
		surf.ptr,
		floatPtr(positions),
		floatPtr(colors),
		C.int(len(batch.Triangles)),
		boolToCInt(true),
	)

	rendered := surf.readImage()
	bounds := dst.Bounds()
	if state.clipRect != nil {
		bounds = bounds.Intersect(rectToImage(*state.clipRect))
	}
	clipMasks := rasterizeClipMasks(w, h, state.clipPaths)
	compositeNativeOver(dst, rendered, bounds, clipMasks)
	return true
}

// --- native surface handle -------------------------------------------------

type nativeSurface struct {
	ptr *C.MgSkSurface
	w   int
	h   int
}

func newNativeSurface(w, h int) *nativeSurface {
	if w <= 0 || h <= 0 {
		return nil
	}
	ptr := C.mgsk_surface_new(C.int(w), C.int(h))
	if ptr == nil {
		return nil
	}
	return &nativeSurface{ptr: ptr, w: w, h: h}
}

func newNativeGPUSurface(w, h, sampleCount int) *nativeSurface {
	if w <= 0 || h <= 0 {
		return nil
	}
	if sampleCount < 0 {
		sampleCount = 0
	}
	ptr := C.mgsk_surface_new_gpu(C.int(w), C.int(h), C.int(sampleCount))
	if ptr == nil {
		return nil
	}
	return &nativeSurface{ptr: ptr, w: w, h: h}
}

func (s *nativeSurface) delete() {
	if s != nil && s.ptr != nil {
		C.mgsk_surface_delete(s.ptr)
		s.ptr = nil
	}
}

func (s *nativeSurface) clear() {
	if s == nil || s.ptr == nil {
		return
	}
	C.mgsk_surface_clear(s.ptr, 0, 0, 0, 0)
}

func (s *nativeSurface) flushGPU() bool {
	return s != nil && s.ptr != nil && C.mgsk_surface_flush_gpu(s.ptr) != 0
}

func (s *nativeSurface) isGPU() bool {
	return s != nil && s.ptr != nil && C.mgsk_surface_is_gpu(s.ptr) != 0
}

func (s *nativeSurface) readImage() *image.RGBA {
	buf := make([]byte, s.h*s.w*4)
	C.mgsk_surface_read_pixels(s.ptr, (*C.uint8_t)(unsafe.Pointer(&buf[0])), C.int(s.w*4))
	return &image.RGBA{Pix: buf, Stride: s.w * 4, Rect: image.Rect(0, 0, s.w, s.h)}
}

// --- marshaling helpers ----------------------------------------------------

// newGradientCPaint builds a C-heap MgSkPaint describing a gradient fill, with
// its stop arrays also in C memory so the struct passed to C contains no Go
// pointers (cgo pointer-passing rule). Gradient endpoints are flipped from
// y-up display space into the y-down device space the caller already used for
// the path. The returned func frees all allocations.
func newGradientCPaint(grad render.GradientFill, height float64) (*C.MgSkPaint, func()) {
	cp := (*C.MgSkPaint)(C.malloc(C.size_t(unsafe.Sizeof(C.MgSkPaint{}))))
	*cp = C.MgSkPaint{}
	cp.antialias = C.int(1)

	stops := append([]render.GradientStop(nil), grad.Stops...)
	sort.SliceStable(stops, func(i, j int) bool { return stops[i].Offset < stops[j].Offset })
	n := len(stops)

	offs := (*C.float)(C.malloc(C.size_t(n) * C.size_t(unsafe.Sizeof(C.float(0)))))
	cols := (*C.float)(C.malloc(C.size_t(n*4) * C.size_t(unsafe.Sizeof(C.float(0)))))
	offSlice := unsafe.Slice((*float32)(unsafe.Pointer(offs)), n)
	colSlice := unsafe.Slice((*float32)(unsafe.Pointer(cols)), n*4)
	for i, s := range stops {
		offSlice[i] = float32(clamp01(s.Offset))
		colSlice[i*4+0] = float32(s.Color.R)
		colSlice[i*4+1] = float32(s.Color.G)
		colSlice[i*4+2] = float32(s.Color.B)
		colSlice[i*4+3] = float32(s.Color.A)
	}
	cp.grad_nstops = C.int(n)
	cp.grad_offsets = offs
	cp.grad_colors = cols

	switch grad.Kind {
	case render.LinearGradient:
		cp.grad_kind = C.int(C.MGSK_GRAD_LINEAR)
		cp.gx0 = C.float(grad.Start.X)
		cp.gy0 = C.float(height - grad.Start.Y)
		cp.gx1 = C.float(grad.End.X)
		cp.gy1 = C.float(height - grad.End.Y)
	case render.RadialGradient:
		cp.grad_kind = C.int(C.MGSK_GRAD_RADIAL)
		cp.gx0 = C.float(grad.Center.X)
		cp.gy0 = C.float(height - grad.Center.Y)
		cp.gr = C.float(grad.Radius)
	}

	free := func() {
		C.free(unsafe.Pointer(offs))
		C.free(unsafe.Pointer(cols))
		C.free(unsafe.Pointer(cp))
	}
	return cp, free
}

func newSolidCPaint(p render.Paint) C.MgSkPaint {
	cp := C.MgSkPaint{}
	cp.fill_r = C.float(p.Fill.R)
	cp.fill_g = C.float(p.Fill.G)
	cp.fill_b = C.float(p.Fill.B)
	cp.fill_a = C.float(p.Fill.A)
	cp.stroke_r = C.float(p.Stroke.R)
	cp.stroke_g = C.float(p.Stroke.G)
	cp.stroke_b = C.float(p.Stroke.B)
	cp.stroke_a = C.float(p.Stroke.A)
	cp.line_width = C.float(p.LineWidth)
	cp.antialias = boolToCInt(p.Antialias != render.AntialiasOff)
	return cp
}

func canDrawPathCollectionItemNative(item *render.PathCollectionItem) bool {
	if item == nil {
		return false
	}
	p := item.Paint
	return p.FillGradient.Kind == render.GradientNone &&
		!hasPatternFill(p.FillPattern) &&
		p.Hatch == "" &&
		item.Hatch == "" &&
		len(p.PathEffects) == 0 &&
		len(p.Dashes) == 0 &&
		p.CompositeMode == render.CompositeSourceOver &&
		p.Rasterization.Mode == render.RasterizeDefault &&
		p.LineJoin == render.JoinMiter &&
		p.LineCap == render.CapButt &&
		!p.HasClipPathTrans &&
		p.Sketch.Scale == 0 &&
		p.Sketch.Length == 0 &&
		p.Sketch.Randomness == 0
}

func canDrawQuadMeshCellNative(cell *render.QuadMeshCell) bool {
	return cell != nil &&
		cell.Edge.A <= 0 &&
		cell.LineWidth <= 0 &&
		len(cell.Dashes) == 0 &&
		cell.Hatch == "" &&
		cell.HatchColor.A <= 0 &&
		cell.HatchWidth <= 0 &&
		cell.Snap == render.SnapOff
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

func skiaImageSampling(name string) int {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "bilinear", "hanning", "hamming", "hermite", "kaiser", "quadric", "gaussian", "bessel", "sinc", "lanczos", "blackman", "auto", "antialiased":
		return 1
	case "bicubic", "spline16", "spline36", "catrom", "mitchell":
		return 2
	default:
		return 0
	}
}

func appendColor(dst []float32, c render.Color) []float32 {
	return append(dst, float32(c.R), float32(c.G), float32(c.B), float32(c.A))
}

func flipPointY(p geom.Pt, height float64) geom.Pt {
	return geom.Pt{X: p.X, Y: height - p.Y}
}

func pathToVerbsCoords(p geom.Path) ([]uint8, []float32) {
	verbs := make([]uint8, 0, len(p.C))
	coords := make([]float32, 0, len(p.V)*2)
	vi := 0
	add := func(pt geom.Pt) { coords = append(coords, float32(pt.X), float32(pt.Y)) }
	for _, c := range p.C {
		switch c {
		case geom.MoveTo:
			verbs = append(verbs, C.MGSK_VERB_MOVE)
			add(p.V[vi])
			vi++
		case geom.LineTo:
			verbs = append(verbs, C.MGSK_VERB_LINE)
			add(p.V[vi])
			vi++
		case geom.QuadTo:
			verbs = append(verbs, C.MGSK_VERB_QUAD)
			add(p.V[vi])
			add(p.V[vi+1])
			vi += 2
		case geom.CubicTo:
			verbs = append(verbs, C.MGSK_VERB_CUBIC)
			add(p.V[vi])
			add(p.V[vi+1])
			add(p.V[vi+2])
			vi += 3
		case geom.ClosePath:
			verbs = append(verbs, C.MGSK_VERB_CLOSE)
		}
	}
	return verbs, coords
}

func bytePtr(b []uint8) *C.uint8_t {
	if len(b) == 0 {
		return nil
	}
	return (*C.uint8_t)(unsafe.Pointer(&b[0]))
}

func floatPtr(f []float32) *C.float {
	if len(f) == 0 {
		return nil
	}
	return (*C.float)(unsafe.Pointer(&f[0]))
}

func boolToCInt(v bool) C.int {
	if v {
		return C.int(1)
	}
	return C.int(0)
}

func rectToImage(r geom.Rect) image.Rectangle {
	return image.Rect(int(r.Min.X), int(r.Min.Y), int(r.Max.X), int(r.Max.Y))
}

func rasterizeClipMasks(w, h int, clipPaths []geom.Path) []*image.Alpha {
	if len(clipPaths) == 0 {
		return nil
	}
	masks := make([]*image.Alpha, 0, len(clipPaths))
	for _, clip := range clipPaths {
		if mask := rasterizeMask(w, h, clip); mask != nil {
			masks = append(masks, mask)
		}
	}
	return masks
}

// compositeNativeOver source-over composites the straight-alpha `rendered`
// surface onto dst within bounds, attenuating by any clip masks. It mirrors the
// per-pixel compositing the CPU bridge performs, so native and CPU outputs
// align.
func compositeNativeOver(dst, rendered *image.RGBA, bounds image.Rectangle, clipMasks []*image.Alpha) {
	bounds = bounds.Intersect(dst.Bounds()).Intersect(rendered.Bounds())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			i := rendered.PixOffset(x, y)
			a := rendered.Pix[i+3]
			if a == 0 {
				continue
			}
			src := color.RGBA{R: rendered.Pix[i], G: rendered.Pix[i+1], B: rendered.Pix[i+2], A: a}
			for _, mask := range clipMasks {
				src.A = uint8(uint32(src.A) * uint32(alphaAt(mask, x, y)) / 255)
				if src.A == 0 {
					break
				}
			}
			if src.A == 0 {
				continue
			}
			blendPixel(dst, x, y, src)
		}
	}
}

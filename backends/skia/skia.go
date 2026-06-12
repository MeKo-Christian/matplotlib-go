//go:build skia

package skia

import (
	"errors"
	"image"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Renderer implements the Skia backend's CPU raster contract.
//
// The current skia-tagged implementation uses the shared Go raster surface as a
// compatibility layer while the external Skia C ABI is still being wired. This
// keeps backend registration, save dispatch, and renderer contract tests
// functional without claiming GPU acceleration or Skia-native optional paths.
type Renderer struct {
	*gobasic.Renderer
	width       int
	height      int
	viewport    geom.Rect
	resolution  uint
	background  render.Color
	bridge      surfaceBridge
	stack       []trackedState
	filterStack []filterState
	clipRect    *geom.Rect
	clipPaths   []geom.Path
	useGPU      bool
	sampleCount int
	colorType   string
}

type filterState struct {
	renderer  *gobasic.Renderer
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

var (
	_ render.Renderer                 = (*Renderer)(nil)
	_ render.DPIAware                 = (*Renderer)(nil)
	_ render.FilterRenderer           = (*Renderer)(nil)
	_ render.TextDrawer               = (*Renderer)(nil)
	_ render.TextPather               = (*Renderer)(nil)
	_ render.RotatedTextDrawer        = (*Renderer)(nil)
	_ render.VerticalTextDrawer       = (*Renderer)(nil)
	_ render.RGBAExporter             = (*Renderer)(nil)
	_ render.PNGExporter              = (*Renderer)(nil)
	_ render.PatternFiller            = (*Renderer)(nil)
	_ render.GradientFiller           = (*Renderer)(nil)
	_ render.CapabilityBridgeReporter = (*Renderer)(nil)
	_ render.RendererModeReporter     = (*Renderer)(nil)
)

// New creates a new Skia renderer with the given configuration.
func New(config backends.Config) (*Renderer, error) {
	skiaConfig, ok := config.Options.(backends.SkiaConfig)
	if !ok {
		// Use defaults if no Skia-specific config provided
		skiaConfig = backends.SkiaConfig{
			UseGPU:      false,
			SampleCount: 1,
			ColorType:   "RGBA8888",
		}
	}
	if skiaConfig.UseGPU && !gpuBuildEnabled {
		return nil, errors.New("skia backend GPU mode requires the skiagpu build tag")
	}
	if config.Width <= 0 || config.Height <= 0 {
		return nil, errors.New("skia backend requires positive width and height")
	}
	if skiaConfig.SampleCount <= 0 {
		skiaConfig.SampleCount = 1
	}
	if skiaConfig.ColorType == "" {
		skiaConfig.ColorType = "RGBA8888"
	}

	cpu := gobasic.New(config.Width, config.Height, config.Background)
	if config.DPI > 0 {
		cpu.SetResolution(uint(config.DPI))
	}
	resolution := uint(72)
	if config.DPI > 0 {
		resolution = uint(config.DPI)
	}
	// Under the skiagpu build tag a GPU-mode request selects the GPU render mode,
	// but the surface remains the deterministic CPU readback bridge until a native
	// SkSurface::MakeRenderTarget path exists. This keeps golden tests reproducible
	// while the GPU plumbing is scaffolded.
	useGPU := skiaConfig.UseGPU && gpuBuildEnabled
	mode := ModeCPU
	if useGPU {
		mode = ModeGPU
	}
	return &Renderer{
		Renderer:    cpu,
		width:       config.Width,
		height:      config.Height,
		resolution:  resolution,
		background:  config.Background,
		bridge:      selectSurfaceBridge(config.Width, config.Height, mode),
		useGPU:      useGPU,
		sampleCount: skiaConfig.SampleCount,
		colorType:   skiaConfig.ColorType,
	}, nil
}

// SetResolution implements render.DPIAware while keeping Skia-side filter
// surfaces in sync with the embedded CPU renderer.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi == 0 {
		dpi = 72
	}
	if r == nil {
		return
	}
	r.resolution = dpi
	if r.Renderer != nil {
		r.Renderer.SetResolution(dpi)
	}
}

// Begin starts a drawing session and resets Skia-side tracked graphics state.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r == nil || r.Renderer == nil {
		return errors.New("skia renderer is not initialized")
	}
	if err := r.Renderer.Begin(viewport); err != nil {
		return err
	}
	r.viewport = viewport
	r.stack = r.stack[:0]
	r.filterStack = r.filterStack[:0]
	r.clipRect = nil
	r.clipPaths = r.clipPaths[:0]
	return nil
}

// End finishes a drawing session and clears Skia-side tracked graphics state.
func (r *Renderer) End() error {
	if r == nil || r.Renderer == nil {
		return errors.New("skia renderer is not initialized")
	}
	if err := r.Renderer.End(); err != nil {
		return err
	}
	r.stack = r.stack[:0]
	r.filterStack = r.filterStack[:0]
	r.clipRect = nil
	r.clipPaths = r.clipPaths[:0]
	return nil
}

// Save pushes the current graphics state.
func (r *Renderer) Save() {
	if r == nil || r.Renderer == nil {
		return
	}
	r.Renderer.Save()
	var clipCopy *geom.Rect
	if r.clipRect != nil {
		copied := *r.clipRect
		clipCopy = &copied
	}
	r.stack = append(r.stack, trackedState{
		clipRect:  clipCopy,
		clipPaths: clonePaths(r.clipPaths),
	})
}

// Restore pops the current graphics state.
func (r *Renderer) Restore() {
	if r == nil || r.Renderer == nil {
		return
	}
	r.Renderer.Restore()
	if len(r.stack) == 0 {
		r.clipRect = nil
		r.clipPaths = r.clipPaths[:0]
		return
	}
	state := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	r.clipRect = state.clipRect
	r.clipPaths = clonePaths(state.clipPaths)
}

// ClipRect intersects the current rectangular clip.
func (r *Renderer) ClipRect(rect geom.Rect) {
	if r == nil || r.Renderer == nil {
		return
	}
	r.Renderer.ClipRect(rect)
	if r.clipRect == nil {
		copied := rect
		r.clipRect = &copied
		return
	}
	intersected := r.clipRect.Intersect(rect)
	r.clipRect = &intersected
}

// ClipPath adds a path clip to the current graphics state.
func (r *Renderer) ClipPath(p geom.Path) {
	if r == nil || r.Renderer == nil {
		return
	}
	r.Renderer.ClipPath(p)
	if len(p.C) == 0 || !p.Validate() {
		return
	}
	r.clipPaths = append(r.clipPaths, clonePath(p))
}

// Path draws a path. Shader fills are routed through the Skia bridge; all other
// path behavior falls back to the shared CPU raster renderer.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if r == nil || r.Renderer == nil || paint == nil {
		return
	}
	if render.DrawPathWithEffects(r, p, paint, r.Path) {
		return
	}
	if r.drawShaderFill(p, paint) {
		return
	}
	if r.drawHatchPrecedencePath(p, paint) {
		return
	}
	if r.drawNativeHatchPath(p, paint) {
		return
	}
	r.Renderer.Path(p, paint)
}

// DrawPathWithEffects applies renderer-neutral path effect passes.
func (r *Renderer) DrawPathWithEffects(p geom.Path, paint *render.Paint) bool {
	return render.DrawPathWithEffects(r, p, paint, r.Path)
}

// StartFilter redirects subsequent drawing to a transparent offscreen CPU
// surface so renderer-neutral filter passes can be post-processed before
// compositing back into the Skia output buffer.
func (r *Renderer) StartFilter() {
	if r == nil || r.Renderer == nil {
		return
	}
	r.filterStack = append(r.filterStack, filterState{
		renderer:  r.Renderer,
		clipRect:  cloneRectPtr(r.clipRect),
		clipPaths: clonePaths(r.clipPaths),
	})
	filter := gobasic.New(r.width, r.height, render.Color{})
	filter.SetResolution(r.resolution)
	if err := filter.Begin(r.viewport); err != nil {
		r.filterStack = r.filterStack[:len(r.filterStack)-1]
		return
	}
	if r.clipRect != nil {
		filter.ClipRect(*r.clipRect)
	}
	for _, clipPath := range r.clipPaths {
		filter.ClipPath(clipPath)
	}
	r.Renderer = filter
}

// StopFilter restores the previous CPU surface and composites the processed
// offscreen pass back into it.
func (r *Renderer) StopFilter(postProcess func(*image.RGBA, float64) (*image.RGBA, geom.Pt)) {
	if r == nil || r.Renderer == nil || len(r.filterStack) == 0 {
		return
	}
	filtered := r.Renderer.GetImage()
	_ = r.Renderer.End()
	state := r.filterStack[len(r.filterStack)-1]
	r.filterStack = r.filterStack[:len(r.filterStack)-1]
	r.Renderer = state.renderer
	r.clipRect = state.clipRect
	r.clipPaths = clonePaths(state.clipPaths)
	if r.Renderer == nil || filtered == nil {
		return
	}
	offset := geom.Pt{}
	if postProcess != nil {
		filtered, offset = postProcess(filtered, float64(r.resolution))
	}
	if filtered == nil || filtered.Bounds().Dx() <= 0 || filtered.Bounds().Dy() <= 0 {
		return
	}
	draw := &image.RGBA{
		Pix:    append([]uint8(nil), filtered.Pix...),
		Stride: filtered.Stride,
		Rect:   filtered.Rect,
	}
	r.Renderer.Image(render.NewImageData(draw), geom.Rect{
		Min: offset,
		Max: geom.Pt{
			X: offset.X + float64(filtered.Bounds().Dx()),
			Y: offset.Y + float64(filtered.Bounds().Dy()),
		},
	})
}

// SupportsPatternFill reports whether the active bridge consumes pattern fills.
func (r *Renderer) SupportsPatternFill() bool {
	return r != nil && r.bridge != nil && r.bridge.SupportsShaders()
}

// SupportsGradientFill reports whether the active bridge consumes gradient fills.
func (r *Renderer) SupportsGradientFill() bool {
	return r != nil && r.bridge != nil && r.bridge.SupportsShaders()
}

// BridgeInfo returns the current Skia bridge mode and feature surface.
func (r *Renderer) BridgeInfo() BridgeInfo {
	if r == nil || r.bridge == nil {
		return BridgeInfo{}
	}
	return r.bridge.Info()
}

// IsCapabilityBridged reports whether the named backends.Capability is
// satisfied through the CPU surface bridge rather than a truly batch-native
// Skia code path.
//
// With the pure-Go CPU bridge the optional batch interfaces and transformed
// images are satisfied by compatibility loops or the embedded CPU renderer, and
// hatch metadata is consumed via the renderer-neutral DrawHatchFallback helper,
// so those report as bridged. When a native Skia surface bridge is active (the
// skiacgo build links a real Skia library), markers, path collections, quad
// meshes, and transformed RGBA images route through native Skia rasterization;
// hatching still loops through the CPU path until its native entrypoint lands,
// so it stays bridged.
func (r *Renderer) IsCapabilityBridged(name string) bool {
	native := r != nil && r.bridge != nil && r.bridge.Info().NativeSurface
	switch name {
	case "imagetransform", "markerbatch", "pathcollectionbatch", "quadmeshbatch":
		return !native
	case "nativehatcher":
		return true
	default:
		return false
	}
}

// RendererModeLabel reports the active Skia render mode for capability reports.
func (r *Renderer) RendererModeLabel() string {
	info := r.BridgeInfo()
	if info.Mode == "" {
		return ""
	}
	return string(info.Mode)
}

// GetSurface returns the underlying Skia surface for advanced operations.
func (r *Renderer) GetSurface() interface{} {
	// No Skia-native surface exists until the external C ABI lands.
	return nil
}

// GetImage returns the CPU raster output buffer.
func (r *Renderer) GetImage() *image.RGBA {
	if r == nil || r.Renderer == nil {
		return nil
	}
	return r.Renderer.GetImage()
}

// FlushGPU flushes pending GPU operations (if using GPU backend).
func (r *Renderer) FlushGPU() {
	if !r.useGPU {
		return
	}
	// TODO: Call GrDirectContext::flushAndSubmit()
}

// GPU returns true if this renderer is using GPU acceleration.
func (r *Renderer) GPU() bool {
	return r.useGPU
}

// SampleCount returns the MSAA sample count.
func (r *Renderer) SampleCount() int {
	return r.sampleCount
}

// ColorType returns the configured Skia color type name.
func (r *Renderer) ColorType() string {
	return r.colorType
}

func buildTagAvailable() bool {
	return true
}

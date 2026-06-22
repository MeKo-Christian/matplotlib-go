// Package agg implements the render.Renderer interface using the AGG (Anti-Grain Geometry)
// rendering library via github.com/cwbudde/agg_go. AGG provides high-quality
// anti-aliased 2D rendering with sub-pixel accuracy.
package agg

import (
	"errors"
	"image"
	"math"
	"os"
	"path/filepath"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
)

func localMatplotlibDejaVuSansPath() string {
	const rel = "third_party/matplotlib/lib/matplotlib/mpl-data/fonts/ttf/DejaVuSans.ttf"
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		path := filepath.Join(wd, rel)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return ""
		}
		wd = parent
	}
}

func defaultTextFontFace() render.FontFace {
	if path := localMatplotlibDejaVuSansPath(); path != "" {
		return render.FontFace{Path: path, Family: "DejaVu Sans", Style: render.FontStyleNormal, Weight: 400}
	}
	face, _ := render.DefaultFontManager().FindFont(render.ParseFontProperties("DejaVu Sans"))
	return face
}

// Renderer implements render.Renderer using the AGG rendering backend.
type Renderer struct {
	ctx                       *aggSurface
	width                     int
	height                    int
	resolution                uint
	began                     bool
	viewport                  geom.Rect
	stack                     []state
	clipRect                  *geom.Rect
	clipPaths                 []geom.Path
	clipMaskMap               map[clipMaskKey][]uint8
	clipScratch               *aggSurface
	clipDepth                 int
	filterStack               []filterState
	defaultFontFace           render.FontFace
	fontPath                  string // path to default TrueType font when one is available on disk
	fallback                  bool   // true if emergency GSV text fallback was used
	emergencyTextFallback     bool
	emergencyTextFallbackUsed bool
	lastFontKey               string
	outlineText               *agglib.FreeTypeOutlineText
	texManager                *tex.Manager
	texErr                    error
	markerScratch             geom.Path
	defaultSketch             render.SketchParams
}

// SetDefaultSketch sets the sketch/xkcd perturbation applied to paths whose
// paint does not carry its own. Implements render.SketchAware.
func (r *Renderer) SetDefaultSketch(params render.SketchParams) { r.defaultSketch = params }

// state represents a saved graphics state.
type state struct {
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

// filterState captures the AGG rendering context required to restore state after
// a temporary filter pass.
type filterState struct {
	ctx       *aggSurface
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

// BufferRegion is kept as a package-local alias for existing AGG callers; the
// shared optional renderer contract lives in render.BufferRegion.
type BufferRegion = render.BufferRegion

type clipMaskKey struct {
	width  int
	height int
	hash   uint64
}

var (
	_ render.Renderer               = (*Renderer)(nil)
	_ render.DPIAware               = (*Renderer)(nil)
	_ render.TextDrawer             = (*Renderer)(nil)
	_ render.FontTextDrawer         = (*Renderer)(nil)
	_ render.RotatedTextDrawer      = (*Renderer)(nil)
	_ render.FontRotatedTextDrawer  = (*Renderer)(nil)
	_ render.VerticalTextDrawer     = (*Renderer)(nil)
	_ render.FontVerticalTextDrawer = (*Renderer)(nil)
	_ render.TextBounder            = (*Renderer)(nil)
	_ render.TextFontMetricer       = (*Renderer)(nil)
	_ render.TextPather             = (*Renderer)(nil)
	_ render.TeXMetricer            = (*Renderer)(nil)
	_ render.TeXDrawer              = (*Renderer)(nil)
	_ render.RotatedTeXDrawer       = (*Renderer)(nil)
	_ render.ImageTransformer       = (*Renderer)(nil)
	_ render.RGBAExporter           = (*Renderer)(nil)
	_ render.BufferRegioner         = (*Renderer)(nil)
	_ render.FilterRenderer         = (*Renderer)(nil)
	_ render.MarkerDrawer           = (*Renderer)(nil)
	_ render.PathCollectionDrawer   = (*Renderer)(nil)
	_ render.QuadMeshDrawer         = (*Renderer)(nil)
	_ render.GouraudTriangleDrawer  = (*Renderer)(nil)
	_ render.NativeHatcher          = (*Renderer)(nil)
	_ render.GradientFiller         = (*Renderer)(nil)
	_ render.PatternFiller          = (*Renderer)(nil)
	_ render.PNGExporter            = (*Renderer)(nil)
)

// New creates a new AGG renderer with the specified dimensions and background color.
// Returns an error if width or height are not positive.
func New(w, h int, bg render.Color) (*Renderer, error) {
	if w <= 0 || h <= 0 {
		return nil, errors.New("agg: width and height must be positive")
	}

	ctx := newAggSurface(w, h)

	// Clear with background color
	bgColor := renderColorToAGG(bg)
	ctx.Clear(bgColor)

	r := &Renderer{
		ctx:             ctx,
		width:           w,
		height:          h,
		resolution:      72,
		clipMaskMap:     make(map[clipMaskKey][]uint8),
		defaultFontFace: defaultTextFontFace(),
		texManager:      tex.NewManager(tex.ManagerConfig{}),
	}

	// Prefer DejaVu Sans (the same default font Matplotlib ships with) through
	// an explicit font resource. Path-backed resources can also prime the
	// legacy AGG text context; embedded resources are handled by the raster
	// pipeline without writing a temporary font file.
	r.fontPath = r.defaultFontFace.Path
	_ = r.ctx.ConfigureTextFont(r.fontPath, 12, r.resolution)

	return r, nil
}

// SetEmergencyTextFallback enables or disables the legacy GSV vector text
// fallback. It is disabled by default because it is a diagnostic/emergency path,
// not a Matplotlib parity renderer.
func (r *Renderer) SetEmergencyTextFallback(enabled bool) {
	if r == nil {
		return
	}
	r.emergencyTextFallback = enabled
}

// EmergencyTextFallbackUsed reports whether the explicit GSV emergency fallback
// has been used by this renderer.
func (r *Renderer) EmergencyTextFallbackUsed() bool {
	return r != nil && r.emergencyTextFallbackUsed
}

// NativeFreetypeVersion returns the linked FreeType library version (e.g.
// "2.13.2") when the cgo-backed FreeType pipeline is available, or "" under
// purego builds where the native path is stubbed out. Useful for tests and
// diagnostics that should only run when real FreeType metrics are in play.
func NativeFreetypeVersion() string {
	return nativeFreetypeVersion()
}

// SetResolution sets the font rendering resolution used for text metrics and glyph sizing.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi > 0 {
		r.resolution = dpi
	}
	r.ctx.SetResolution(r.resolution)
}

// Resolution reports the current font rendering DPI (render.DPIProvider).
func (r *Renderer) Resolution() uint {
	return r.resolution
}

// Clear resets the reusable renderer surface to c and clears transient drawing
// state left by the previous frame. It is intended for repeated redraw loops
// that keep one renderer for a stable canvas size.
func (r *Renderer) Clear(c render.Color) {
	if r == nil || r.ctx == nil {
		return
	}
	r.began = false
	r.viewport = geom.Rect{}
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipPaths = r.clipPaths[:0]
	r.filterStack = nil
	r.clipDepth = 0
	r.ctx.ClipBox(0, 0, float64(r.width), float64(r.height))
	r.ctx.Clear(renderColorToAGG(c))
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
	r.filterStack = nil
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipPaths = r.clipPaths[:0]
	return nil
}

// StartFilter saves the active surface and begins a new temporary filter surface.
func (r *Renderer) StartFilter() {
	if r == nil || r.ctx == nil {
		return
	}
	var clipRect *geom.Rect
	if r.clipRect != nil {
		rect := *r.clipRect
		clipRect = &rect
	}
	r.filterStack = append(r.filterStack, filterState{
		ctx:       r.ctx,
		clipRect:  clipRect,
		clipPaths: clonePaths(r.clipPaths),
	})
	r.ctx = newAggSurface(r.width, r.height)
	r.clipRect = nil
	r.clipPaths = nil
	r.applyClipRect()
}

// StopFilter restores the previous surface and optionally processes the temporary
// filter surface before compositing it back.
func (r *Renderer) StopFilter(postProcess func(img *image.RGBA, dpi float64) (*image.RGBA, geom.Pt)) {
	if r == nil || len(r.filterStack) == 0 || r.ctx == nil {
		return
	}

	state := r.filterStack[len(r.filterStack)-1]
	r.filterStack = r.filterStack[:len(r.filterStack)-1]

	filtered := r.ctx.GetImage().ToGoImage()
	r.ctx = state.ctx
	r.clipRect = state.clipRect
	r.clipPaths = state.clipPaths
	r.applyClipRect()
	if state.ctx == nil {
		return
	}

	postProcessed := filtered
	offset := geom.Pt{}
	if postProcess != nil {
		postProcessed, offset = postProcess(filtered, float64(r.resolution))
	}
	if postProcessed == nil || postProcessed.Bounds().Dx() <= 0 || postProcessed.Bounds().Dy() <= 0 {
		return
	}

	draw := &image.RGBA{
		Pix:    append([]uint8(nil), postProcessed.Pix...),
		Stride: postProcessed.Stride,
		Rect:   postProcessed.Rect,
	}
	r.drawImageDirect(render.NewImageData(draw), geom.Rect{
		Min: offset,
		Max: geom.Pt{
			X: offset.X + float64(postProcessed.Bounds().Dx()),
			Y: offset.Y + float64(postProcessed.Bounds().Dy()),
		},
	})
}

// CopyFromBBox captures a pixel region from the active surface.
func (r *Renderer) CopyFromBBox(bbox geom.Rect) *BufferRegion {
	if r == nil || r.ctx == nil || r.ctx.image == nil {
		return nil
	}
	if bbox.W() <= 0 || bbox.H() <= 0 || r.width <= 0 || r.height <= 0 {
		return nil
	}

	// bbox arrives in y-up display space; flip to the y-down device buffer so
	// the captured rows match matplotlib's copy_from_bbox (height - y).
	bbox = r.devRect(bbox)

	minX := int(math.Floor(bbox.Min.X))
	minY := int(math.Floor(bbox.Min.Y))
	maxX := int(math.Ceil(bbox.Max.X))
	maxY := int(math.Ceil(bbox.Max.Y))

	minX = maxInt(minX, 0)
	minY = maxInt(minY, 0)
	maxX = minInt(maxX, r.width)
	maxY = minInt(maxY, r.height)
	if minX < 0 || minY < 0 || maxX < 0 || maxY < 0 || minX >= maxX || minY >= maxY {
		return nil
	}

	width := maxX - minX
	height := maxY - minY
	if width <= 0 || height <= 0 {
		return nil
	}

	src := r.ctx.GetImage()
	if src == nil {
		return nil
	}

	out := image.NewRGBA(image.Rect(0, 0, width, height))
	srcStride := src.Stride()
	dstStride := out.Stride
	for y := 0; y < height; y++ {
		srcOff := (minY+y)*srcStride + minX*4
		dstOff := y * dstStride
		copy(out.Pix[dstOff:dstOff+width*4], src.Data[srcOff:srcOff+width*4])
	}

	return &BufferRegion{
		Image: out,
		Rect: geom.Rect{
			Min: geom.Pt{X: float64(minX), Y: float64(minY)},
			Max: geom.Pt{X: float64(maxX), Y: float64(maxY)},
		},
	}
}

// RestoreRegion composits a previously captured buffer region back onto the
// current surface. A nil bbox restores the full region.
//
// TODO(y-flip): region.Rect is captured in y-down device space by CopyFromBbox,
// and the crop/offset arithmetic below operates in device/image-local space. If
// callers ever pass bbox/offset in y-up display coordinates they must be flipped
// here. This path is not exercised by the static PNG goldens, so it is left
// device-space for now.
func (r *Renderer) RestoreRegion(region *BufferRegion, bbox *geom.Rect, offset geom.Pt) {
	if r == nil || region == nil || region.Image == nil || r.ctx == nil {
		return
	}
	if r.width <= 0 || r.height <= 0 || region.Image.Bounds().Dx() <= 0 || region.Image.Bounds().Dy() <= 0 {
		return
	}

	minX, minY := 0, 0
	maxX, maxY := region.Image.Bounds().Dx(), region.Image.Bounds().Dy()
	if bbox != nil && bbox.W() > 0 && bbox.H() > 0 {
		minX = int(math.Floor(bbox.Min.X))
		minY = int(math.Floor(bbox.Min.Y))
		maxX = int(math.Ceil(bbox.Max.X))
		maxY = int(math.Ceil(bbox.Max.Y))
		minX = maxInt(minX, 0)
		minY = maxInt(minY, 0)
		maxX = minInt(maxX, region.Image.Bounds().Dx())
		maxY = minInt(maxY, region.Image.Bounds().Dy())
		if minX >= maxX || minY >= maxY {
			return
		}
	}

	width := maxX - minX
	height := maxY - minY
	if width <= 0 || height <= 0 {
		return
	}

	cropped := image.NewRGBA(image.Rect(0, 0, width, height))
	src := region.Image
	srcStride := src.Stride
	dstStride := cropped.Stride
	for y := 0; y < height; y++ {
		srcBase := (minY+y)*srcStride + minX*4
		dstBase := y * dstStride
		copy(cropped.Pix[dstBase:dstBase+width*4], src.Pix[srcBase:srcBase+width*4])
	}

	drawX := region.Rect.Min.X + float64(minX) + offset.X
	drawY := region.Rect.Min.Y + float64(minY) + offset.Y
	r.drawImageDirect(render.NewImageData(cropped), geom.Rect{
		Min: geom.Pt{X: drawX, Y: drawY},
		Max: geom.Pt{X: drawX + float64(width), Y: drawY + float64(height)},
	})
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
	r.ctx.PushTransform()
}

// Restore pops the graphics state from the stack.
func (r *Renderer) Restore() {
	if len(r.stack) == 0 {
		return
	}
	s := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	r.clipRect = s.clipRect
	r.clipPaths = clonePaths(s.clipPaths)
	r.ctx.PopTransform()
	r.applyClipRect()
}

// ClipRect sets a rectangular clip region.
func (r *Renderer) ClipRect(rect geom.Rect) {
	// rect arrives in y-up display space; flip to the y-down device buffer so
	// the intersect/applyClipRect logic stays in device space, unchanged.
	rect = r.devRect(rect)
	if r.clipRect == nil {
		r.clipRect = &rect
	} else {
		intersected := r.clipRect.Intersect(rect)
		r.clipRect = &intersected
	}
	r.applyClipRect()
}

// ClipPath adds a path-based clip region to the current graphics state.
func (r *Renderer) ClipPath(p geom.Path) {
	if len(p.C) == 0 || !p.Validate() {
		return
	}
	// Store the clip path in y-down device space so it matches the device-space
	// geometry handed to the clip-mask pipeline.
	r.clipPaths = append(r.clipPaths, clonePath(r.devPath(p)))
}

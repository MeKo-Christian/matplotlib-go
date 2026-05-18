// Package agg implements the render.Renderer interface using the AGG (Anti-Grain Geometry)
// rendering library via github.com/cwbudde/agg_go. AGG provides high-quality
// anti-aliased 2D rendering with sub-pixel accuracy.
package agg

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
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
}

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
	r.clipPaths = append(r.clipPaths, clonePath(p))
}

// Path draws a path with the given paint style.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if render.DrawPathWithEffects(r, p, paint, r.Path) {
		return
	}
	if r.hasClipPath() {
		bounds, haveBounds := pathDrawBounds(p, paint)
		r.withClipPathMask(bounds, haveBounds, func() {
			r.drawPathDirect(p, paint)
		})
		return
	}
	r.drawPathDirect(p, paint)
}

// DrawPathWithEffects applies renderer-neutral path effect passes.
func (r *Renderer) DrawPathWithEffects(p geom.Path, paint *render.Paint) bool {
	return render.DrawPathWithEffects(r, p, paint, r.Path)
}

func (r *Renderer) drawPathDirect(p geom.Path, paint *render.Paint) {
	if !p.Validate() || paint == nil {
		return
	}
	paintCopy := *paint
	paint = &paintCopy
	applyForcedAlpha(paint)
	p, ok := r.preparePathForPaint(p, paint)
	if !ok {
		return
	}
	restoreAA := r.applyAntialiasMode(paint.Antialias)
	defer restoreAA()

	// Fill first if requested. Gradient fills take precedence over solid
	// fills; the gradient endpoint colors carry their own alpha so a missing
	// solid Fill.A is fine for gradient-only paints.
	hasGradient := paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0
	if hasGradient {
		if r.applyGradientFill(paint) {
			r.buildPath(p)
			r.ctx.Fill()
			// Restore the solid fill source so downstream draws (hatch
			// overlay, stroke, marker batches, etc.) are not accidentally
			// painted through the gradient span generator.
			r.ctx.SetFillColor(renderColorToAGG(colorWithForcedAlpha(paint.Fill, paint)))
		}
	} else if paint.Fill.A > 0 {
		r.buildPath(p)
		r.ctx.SetFillColor(renderColorToAGG(colorWithForcedAlpha(paint.Fill, paint)))
		r.ctx.Fill()
	}

	if paint.Hatch != "" {
		r.drawNativeHatch(p, paint)
	}

	// Then stroke if requested
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		r.ctx.SetStrokeColor(renderColorToAGG(colorWithForcedAlpha(paint.Stroke, paint)))
		r.ctx.SetStrokeWidth(paint.LineWidth)

		// Map line join
		switch paint.LineJoin {
		case render.JoinMiter:
			r.ctx.SetLineJoin(agglib.JoinMiter)
		case render.JoinRound:
			r.ctx.SetLineJoin(agglib.JoinRound)
		case render.JoinBevel:
			r.ctx.SetLineJoin(agglib.JoinBevel)
		}

		// Map line cap
		switch paint.LineCap {
		case render.CapButt:
			r.ctx.SetLineCap(agglib.CapButt)
		case render.CapRound:
			r.ctx.SetLineCap(agglib.CapRound)
		case render.CapSquare:
			r.ctx.SetLineCap(agglib.CapSquare)
		}

		// Set miter limit
		if paint.MiterLimit > 0 {
			r.ctx.SetMiterLimit(paint.MiterLimit)
		}

		// Handle dashes
		r.ctx.ClearDashes()
		if len(paint.Dashes) >= 2 {
			r.ctx.SetDashPattern(paint.Dashes)
		}

		for _, path := range chunkStrokePath(p, paint.MaxChunkVertices) {
			r.buildPath(path)
			r.ctx.Stroke()
		}

		// Clean up dashes
		if len(paint.Dashes) >= 2 {
			r.ctx.ClearDashes()
		}
	}
}

// buildPath converts a geom.Path into AGG path commands on the current context.
func (r *Renderer) buildPath(p geom.Path) {
	r.ctx.BeginPath()

	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(p.V) {
				return
			}
			pt := p.V[vi]
			r.ctx.MoveTo(pt.X, pt.Y)
			vi++
		case geom.LineTo:
			if vi >= len(p.V) {
				return
			}
			pt := p.V[vi]
			r.ctx.LineTo(pt.X, pt.Y)
			vi++
		case geom.QuadTo:
			if vi+1 >= len(p.V) {
				return
			}
			ctrl := p.V[vi]
			to := p.V[vi+1]
			r.ctx.QuadricCurveTo(ctrl.X, ctrl.Y, to.X, to.Y)
			vi += 2
		case geom.CubicTo:
			if vi+2 >= len(p.V) {
				return
			}
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			to := p.V[vi+2]
			r.ctx.CubicCurveTo(c1.X, c1.Y, c2.X, c2.Y, to.X, to.Y)
			vi += 3
		case geom.ClosePath:
			r.ctx.ClosePath()
		}
	}
}

func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	if r.hasClipPath() {
		bounds, haveBounds := imageDrawBounds(dst)
		r.withClipPathMask(bounds, haveBounds, func() {
			r.drawImageDirect(img, dst)
		})
		return
	}
	r.drawImageDirect(img, dst)
}

func (r *Renderer) drawImageDirect(img render.Image, dst geom.Rect) {
	aggImg, ok := renderImageToAGG(img)
	if !ok {
		return
	}

	agg := r.ctx
	prevBlendMode := agg.GetBlendMode()
	prevFilter := agg.GetImageFilter()
	prevResample := agg.GetImageResample()
	agg.SetBlendMode(agglib.BlendSrcOver)
	defer func() {
		agg.SetBlendMode(prevBlendMode)
		agg.SetImageFilter(prevFilter)
		agg.SetImageResample(prevResample)
	}()

	x := dst.Min.X
	y := dst.Min.Y
	w := dst.W()
	h := dst.H()
	if w < 0 {
		x += w
		w = -w
	}
	if h < 0 {
		y += h
		h = -h
	}
	if w <= 0 || h <= 0 {
		return
	}
	applyInterpolation(agg, img, w, h)

	_ = agg.DrawImageScaled(aggImg, x, y, w, h)
}

// ImageTransformed draws an image using the provided affine transformation.
// Used by core.Image2D when rotation is requested.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, affine geom.Affine) {
	if r.hasClipPath() {
		bounds, haveBounds := transformedImageDrawBounds(img, affine)
		r.withClipPathMask(bounds, haveBounds, func() {
			r.drawImageTransformedDirect(img, affine)
		})
		return
	}
	r.drawImageTransformedDirect(img, affine)
}

func (r *Renderer) drawImageTransformedDirect(img render.Image, affine geom.Affine) {
	aggImg, ok := renderImageToAGG(img)
	if !ok {
		return
	}

	agg := r.ctx
	prevBlendMode := agg.GetBlendMode()
	prevFilter := agg.GetImageFilter()
	prevResample := agg.GetImageResample()
	affineDispX, affineDispY := imageTransformDisplaySpan(img, affine)
	agg.SetBlendMode(agglib.BlendSrcOver)
	applyInterpolation(agg, img, affineDispX, affineDispY)
	defer func() {
		agg.SetBlendMode(prevBlendMode)
		agg.SetImageFilter(prevFilter)
		agg.SetImageResample(prevResample)
	}()

	transform := agglib.NewTransformationsFromValues(
		affine.A,
		affine.B,
		affine.C,
		affine.D,
		affine.E,
		affine.F,
	)
	_ = agg.DrawImageTransformed(aggImg, transform)
}

func extractImageAlpha(img render.Image) float64 {
	if img == nil {
		return 1
	}
	imageAlpha, ok := img.(render.ImageAlpha)
	if !ok {
		return 1
	}
	return clamp01(imageAlpha.Alpha())
}

// DrawMarkers renders one marker path at many display-space offsets.
func (r *Renderer) DrawMarkers(batch render.MarkerBatch) bool {
	if len(batch.Marker.C) == 0 || len(batch.Items) == 0 {
		return false
	}
	for i := range batch.Items {
		item := &batch.Items[i]
		offset := item.Offset
		markerPaint := item.Paint
		markerPaint.Snap = render.SnapAuto
		if !shouldSnapPath(batch.Marker, &markerPaint) {
			offset.X = math.Floor(offset.X+0.5) + 0.5
			offset.Y = math.Floor(offset.Y+0.5) + 0.5
		}
		path := transformMarkerPath(batch.Marker, item.Transform, offset)
		if len(path.C) == 0 {
			continue
		}
		paint := markerPaint
		if !item.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		r.Path(path, &paint)
	}
	return true
}

// DrawPathCollection renders a display-space path collection.
func (r *Renderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	if len(batch.Items) == 0 {
		return false
	}
	for i := range batch.Items {
		item := &batch.Items[i]
		if len(item.Path.C) == 0 {
			continue
		}
		paint := item.Paint
		if item.Hatch != "" {
			paint.Hatch = item.Hatch
			paint.HatchColor = item.HatchColor
			paint.HatchLineWidth = item.HatchWidth
			paint.HatchSpacing = item.HatchSpacing
		}
		if !item.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		r.Path(item.Path, &paint)
	}
	return true
}

// DrawQuadMesh renders pcolor/pcolormesh-style quadrilateral cells.
func (r *Renderer) DrawQuadMesh(batch render.QuadMeshBatch) bool {
	if len(batch.Cells) == 0 {
		return false
	}
	for i := range batch.Cells {
		cell := &batch.Cells[i]
		path := geom.Path{}
		path.MoveTo(cell.Quad[0])
		path.LineTo(cell.Quad[1])
		path.LineTo(cell.Quad[2])
		path.LineTo(cell.Quad[3])
		path.Close()
		paint := render.Paint{
			Fill:         cell.Face,
			Stroke:       cell.Edge,
			LineWidth:    cell.LineWidth,
			LineJoin:     render.JoinMiter,
			LineCap:      render.CapButt,
			Dashes:       append([]float64(nil), cell.Dashes...),
			Hatch:        cell.Hatch,
			HatchColor:   cell.HatchColor,
			HatchSpacing: cell.HatchSpacing,
			Antialias:    render.AntialiasDefault,
			Snap:         render.SnapOn,
		}
		if cell.HatchWidth > 0 {
			paint.HatchLineWidth = cell.HatchWidth
		}
		if !cell.Antialiased {
			paint.Antialias = render.AntialiasOff
		}
		if paint.LineWidth <= 0 || paint.Stroke.A <= 0 {
			paint.Stroke = render.Color{}
			paint.LineWidth = 0
		}
		if paint.Fill.A <= 0 {
			paint.Fill = render.Color{}
		}
		r.Path(path, &paint)
	}
	return true
}

// DrawGouraudTriangles renders interpolated-color triangles directly into the
// AGG surface buffer.
func (r *Renderer) DrawGouraudTriangles(batch render.GouraudTriangleBatch) bool {
	if len(batch.Triangles) == 0 || r.ctx == nil || r.ctx.image == nil {
		return false
	}
	draw := func() {
		for i := range batch.Triangles {
			r.drawGouraudTriangle(&batch.Triangles[i])
		}
	}
	if r.hasClipPath() {
		bounds, haveBounds := gouraudTriangleBatchBounds(batch)
		r.withClipPathMask(bounds, haveBounds, draw)
	} else {
		draw()
	}
	return true
}

// GlyphRun draws a run of glyphs.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if len(run.Glyphs) == 0 || run.Size <= 0 || r.ctx == nil {
		return
	}

	if run.FontKey != "" {
		r.lastFontKey = run.FontKey
	}

	font := r.configureTextFont(run.Size, r.lastFontKey)
	if font.backend != textBackendRaster {
		return
	}

	if r.drawGlyphRunAsPaths(font.face, run, textColor) {
		return
	}

	_ = r.drawGlyphRunLegacy(run, textColor, r.fontPixelSize(font.size))
}

func (r *Renderer) drawGlyphRunAsPaths(face render.FontFace, run render.GlyphRun, textColor render.Color) bool {
	if r.ctx == nil {
		return false
	}

	fontSize := r.fontPixelSize(run.Size)
	if fontSize <= 0 {
		return false
	}

	fontData, err := loadSFNTFontFace(face)
	if err != nil {
		return false
	}

	ppem := fixed.Int26_6(math.Round(fontSize * 64))
	if ppem <= 0 {
		return false
	}

	var (
		path  geom.Path
		buf   sfnt.Buffer
		drawn bool
	)
	for _, glyph := range run.Glyphs {
		if glyph.ID == 0 {
			continue
		}
		segments, err := fontData.LoadGlyph(&buf, sfnt.GlyphIndex(glyph.ID), ppem, nil)
		if err != nil || len(segments) == 0 {
			continue
		}
		appendGlyphSegments(&path, segments, geom.Pt{
			X: run.Origin.X + glyph.Offset.X,
			Y: run.Origin.Y + glyph.Offset.Y,
		})
		drawn = true
	}
	if !drawn {
		return false
	}

	r.Path(path, &render.Paint{Fill: textColor})
	return true
}

func (r *Renderer) drawGlyphRunLegacy(run render.GlyphRun, textColor render.Color, size float64) bool {
	if size <= 0 {
		return false
	}

	fontPath := r.resolveTextFontPath(run.FontKey)
	if fontPath == "" {
		return false
	}

	drawn := false
	for _, glyph := range run.Glyphs {
		if glyph.ID == 0 {
			continue
		}
		origin := geom.Pt{
			X: run.Origin.X + glyph.Offset.X,
			Y: run.Origin.Y + glyph.Offset.Y,
		}
		if r.drawTextPathFallback(string(rune(glyph.ID)), origin, size, textColor, fontPath) {
			drawn = true
		}
	}
	return drawn
}

func appendGlyphSegments(path *geom.Path, segments sfnt.Segments, origin geom.Pt) {
	for _, segment := range segments {
		switch segment.Op {
		case sfnt.SegmentOpMoveTo:
			path.MoveTo(sfntPoint(segment.Args[0], origin))
		case sfnt.SegmentOpLineTo:
			path.LineTo(sfntPoint(segment.Args[0], origin))
		case sfnt.SegmentOpQuadTo:
			path.QuadTo(
				sfntPoint(segment.Args[0], origin),
				sfntPoint(segment.Args[1], origin),
			)
		case sfnt.SegmentOpCubeTo:
			path.CubicTo(
				sfntPoint(segment.Args[0], origin),
				sfntPoint(segment.Args[1], origin),
				sfntPoint(segment.Args[2], origin),
			)
		}
	}
}

func sfntPoint(p fixed.Point26_6, origin geom.Pt) geom.Pt {
	return geom.Pt{
		X: origin.X + sfntFixedToFloat(p.X),
		Y: origin.Y + sfntFixedToFloat(p.Y),
	}
}

func sfntFixedToFloat(v fixed.Int26_6) float64 {
	return float64(v) / 64.0
}

// MeasureText measures text dimensions using the active font engine.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" || size <= 0 {
		return render.TextMetrics{}
	}

	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}

	font := r.configureTextFont(size, fontKey)

	var (
		w       float64
		ascent  float64
		descent float64
	)
	switch font.backend {
	case textBackendRaster:
		if metrics, ok := r.measureRasterText(text, font.face, font.size); ok {
			return metrics
		}
		sizePx := r.fontPixelSize(font.size)
		if x, y, bw, h, ok := measureTextPathBounds(text, sizePx, font.fontPath); ok {
			w = math.Max(w, x+bw)
			ascent = math.Max(0, -y)
			descent = math.Max(0, y+h)
		} else {
			return render.TextMetrics{}
		}
	case textBackendGSV:
		sizePx := r.fontPixelSize(font.size)
		w = measureLocalGSVTextWidth(text, sizePx)
		if _, y, _, h, ok := measureLocalGSVTextBounds(text, sizePx); ok {
			ascent = math.Max(0, -y)
			descent = math.Max(0, y+h)
		} else {
			ascent = sizePx
			descent = 0
		}
	default:
		return render.TextMetrics{}
	}

	h := ascent + descent
	if h <= 0 {
		h = font.size
	}
	if ascent <= 0 {
		ascent = h
	}
	if descent < 0 {
		descent = 0
	}

	return render.TextMetrics{
		W:       w,
		H:       h,
		Ascent:  ascent,
		Descent: descent,
	}
}

// MeasureTextBounds reports the actual ink bounds of text relative to the
// baseline origin used for DrawText.
func (r *Renderer) MeasureTextBounds(text string, size float64, fontKey string) (render.TextBounds, bool) {
	if text == "" || size <= 0 {
		return render.TextBounds{}, false
	}

	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}

	font := r.configureTextFont(size, fontKey)
	if font.backend == textBackendGSV {
		sizePx := r.fontPixelSize(font.size)
		x, y, w, h, ok := measureLocalGSVTextBounds(text, sizePx)
		if !ok {
			return render.TextBounds{}, false
		}
		return render.TextBounds{X: x, Y: y, W: w, H: h}, true
	}
	if font.backend != textBackendRaster {
		return render.TextBounds{}, false
	}
	sizePx := r.fontPixelSize(font.size)
	fontKey = fontReference(font.face)
	if shaped, ok := render.ShapeText(text, geom.Pt{}, sizePx, render.TextShapingOptions{FontKey: fontKey}); ok {
		if shapedTextMatchesRuneSequence(text, shaped) {
			if bounds, ok := r.measureNativeFreetypeTextBounds(text, font.face, font.size, matplotlibTextHintingFactor); ok {
				return bounds, true
			}
		}
		return shaped.Bounds, true
	}
	if x, y, w, h, ok := measureTextPathBounds(text, sizePx, font.fontPath); ok {
		return render.TextBounds{X: x, Y: y, W: w, H: h}, true
	}
	return render.TextBounds{}, false
}

// MeasureFontHeights reports font-wide ascent, descent, and line-gap values
// for the current raster text face, distinct from a particular string's ink
// bounds.
func (r *Renderer) MeasureFontHeights(size float64, fontKey string) (render.FontHeightMetrics, bool) {
	if size <= 0 {
		return render.FontHeightMetrics{}, false
	}

	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}

	font := r.configureTextFont(size, fontKey)
	if font.backend == textBackendGSV {
		sizePx := r.fontPixelSize(font.size)
		if _, y, _, h, ok := measureLocalGSVTextBounds("lp", sizePx); ok {
			return render.FontHeightMetrics{
				Ascent:  math.Max(0, -y),
				Descent: math.Max(0, y+h),
			}, true
		}
		return render.FontHeightMetrics{}, false
	}
	if font.backend != textBackendRaster {
		return render.FontHeightMetrics{}, false
	}

	if metrics, ok := r.measureNativeFreetypeFontHeights(font.face, font.size, matplotlibTextHintingFactor); ok {
		return metrics, true
	}

	if metrics, ok := rasterFontHeightMetrics(font.face, font.size, r.resolution); ok {
		return render.FontHeightMetrics{
			Ascent:  metrics.ascent,
			Descent: metrics.descent,
			LineGap: metrics.lineGap,
		}, true
	}

	if err := r.ctx.ConfigureTextFont(font.fontPath, font.size, r.resolution); err != nil {
		sizePx := r.fontPixelSize(font.size)
		if _, y, _, h, ok := measureTextPathBounds("lp", sizePx, font.fontPath); ok {
			return render.FontHeightMetrics{
				Ascent:  math.Max(0, -y),
				Descent: math.Max(0, y+h),
			}, true
		}
		return render.FontHeightMetrics{}, false
	}

	metrics, ok := r.ctx.fontHeightMetrics()
	if !ok {
		return render.FontHeightMetrics{}, false
	}
	return render.FontHeightMetrics{
		Ascent:  metrics.ascent,
		Descent: metrics.descent,
		LineGap: metrics.lineGap,
	}, true
}

func rasterFontHeightMetrics(face render.FontFace, size float64, resolution uint) (fontHeightMetrics, bool) {
	if fontReference(face) == "" || size <= 0 {
		return fontHeightMetrics{}, false
	}
	dpi := float64(resolution)
	if dpi <= 0 {
		dpi = 72
	}
	resource, err := loadSFNTFontFace(face)
	if err != nil {
		return fontHeightMetrics{}, false
	}
	scale, ok := sfntMetricScale(resource.data, size, dpi)
	if !ok {
		return fontHeightMetrics{}, false
	}
	ascent, descent, lineGap, ok := sfntTableHeightMetrics(resource.data, scale)
	if !ok {
		return fontHeightMetrics{}, false
	}
	return fontHeightMetrics{
		ascent:  ascent,
		descent: descent,
		lineGap: lineGap,
	}, true
}

// TextPath converts text to a vector path using the renderer's resolved font.
func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if text == "" || size <= 0 {
		return geom.Path{}, false
	}
	if fontKey == "" {
		fontKey = r.lastFontKey
	}
	font := r.configureTextFont(size, fontKey)
	if fontReference(font.face) == "" {
		return geom.Path{}, false
	}
	return render.TextPath(text, origin, r.fontPixelSize(size), fontReference(font.face))
}

// DrawText renders text at the given position with the specified size and color.
// This is a helper method (not part of the Renderer interface).
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	r.drawTextWithFontContext(text, origin, size, textColor, r.lastFontKey)
}

// DrawTextWithFont renders text with an explicit font key, avoiding the legacy
// MeasureText-before-DrawText state handoff.
func (r *Renderer) DrawTextWithFont(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	r.withTemporaryFontKey(func() {
		r.drawTextWithFontContext(text, origin, size, textColor, fontKey)
	})
}

func (r *Renderer) drawTextWithFontContext(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	if r.hasClipPath() {
		bounds, haveBounds := r.textDrawBounds(text, origin, size, fontKey)
		r.withClipPathMask(bounds, haveBounds, func() {
			r.drawTextDirect(text, origin, size, textColor, fontKey)
		})
		return
	}
	r.drawTextDirect(text, origin, size, textColor, fontKey)
}

func (r *Renderer) drawTextDirect(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	if text == "" || size <= 0 {
		return
	}

	font := r.configureTextFont(size, fontKey)

	switch font.backend {
	case textBackendRaster:
		if r.drawRasterText(text, font.face, origin, font.size, textColor) {
			return
		}
		sizePx := r.fontPixelSize(font.size)
		if r.drawTextPathFallback(text, origin, sizePx, textColor, font.fontPath) {
			return
		}
		if r.drawEmergencyGSVText(text, origin, font.size, textColor) {
			return
		}
		return
	case textBackendGSV:
		_ = r.drawEmergencyGSVText(text, origin, font.size, textColor)
	default:
		_ = r.drawEmergencyGSVText(text, origin, size, textColor)
		return
	}
}

// DrawTextRotated renders text using Matplotlib-like anchor rotation. The
// anchor is the bottom-center of the unrotated text box.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	r.drawTextRotatedWithFontContext(text, anchor, size, angle, textColor, r.lastFontKey)
}

// DrawTextRotatedWithFont renders rotated text with an explicit font key.
func (r *Renderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	r.withTemporaryFontKey(func() {
		r.drawTextRotatedWithFontContext(text, anchor, size, angle, textColor, fontKey)
	})
}

func (r *Renderer) drawTextRotatedWithFontContext(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	if r.hasClipPath() {
		bounds, haveBounds := r.rotatedTextDrawBounds(text, anchor, size, angle, fontKey)
		r.withClipPathMask(bounds, haveBounds, func() {
			r.drawTextRotatedDirect(text, anchor, size, angle, textColor, fontKey)
		})
		return
	}
	r.drawTextRotatedDirect(text, anchor, size, angle, textColor, fontKey)
}

func (r *Renderer) drawTextRotatedDirect(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	if text == "" || size <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return
	}

	metrics := r.MeasureText(text, size, fontKey)
	if metrics.W <= 0 || metrics.H <= 0 {
		return
	}

	bounds, haveBounds := r.MeasureTextBounds(text, size, fontKey)
	origin := rotatedTextOrigin(anchor, metrics, bounds, haveBounds)
	font := r.configureTextFont(size, fontKey)

	r.ctx.PushTransform()
	defer r.ctx.PopTransform()

	r.ctx.Translate(-anchor.X, -anchor.Y)
	r.ctx.Rotate(-angle)
	r.ctx.Translate(anchor.X, anchor.Y)

	if fontReference(font.face) != "" {
		sizePx := r.fontPixelSize(font.size)
		fontKey := fontReference(font.face)
		if r.drawTextPathFallback(text, origin, sizePx, textColor, fontKey) {
			return
		}
		if font.fontPath != "" {
			if face, err := r.configureOutlineFont(font.fontPath, sizePx); err == nil {
				r.ctx.SetFillColor(renderColorToAGG(textColor))
				r.ctx.SetStrokeColor(renderColorToAGG(textColor))
				if drawTrueTypeOutlineText(r.ctx, face, origin.X, origin.Y, text) {
					return
				}
			}
		}
	}

	if font.backend == textBackendGSV {
		_ = r.drawEmergencyGSVText(text, origin, font.size, textColor)
		return
	}
	_ = r.drawEmergencyGSVText(text, origin, size, textColor)
}

func (r *Renderer) fontPixelSize(size float64) float64 {
	dpi := float64(r.resolution)
	if dpi <= 0 {
		dpi = 72
	}
	return size * dpi / 72.0
}

func (r *Renderer) drawTextPathFallback(text string, origin geom.Pt, size float64, textColor render.Color, fontPath string) bool {
	if fontPath == "" {
		return false
	}
	path, ok := render.TextPath(text, origin, size, fontPath)
	if !ok {
		return false
	}
	r.Path(path, &render.Paint{
		Fill: textColor,
	})
	return true
}

func (r *Renderer) drawEmergencyGSVText(text string, origin geom.Pt, size float64, textColor render.Color) bool {
	if r == nil || r.ctx == nil || !r.emergencyTextFallback || text == "" || size <= 0 {
		return false
	}
	lineWidth := r.ctx.painter.GetLineWidth()
	lineCap := r.ctx.painter.GetLineCap()
	lineJoin := r.ctx.painter.GetLineJoin()
	lineColor := r.ctx.painter.GetLineColor()
	defer func() {
		r.ctx.painter.LineWidth(lineWidth)
		r.ctx.painter.LineCap(lineCap)
		r.ctx.painter.LineJoin(lineJoin)
		r.ctx.painter.LineColor(lineColor)
	}()

	sizePx := r.fontPixelSize(size)
	r.ctx.SetStrokeColor(renderColorToAGG(textColor))
	r.ctx.SetStrokeWidth(math.Max(1, sizePx*0.08))
	r.ctx.SetLineCap(agglib.CapRound)
	r.ctx.SetLineJoin(agglib.JoinRound)
	if !appendLocalGSVText(r.ctx, origin.X, origin.Y, sizePx, text) {
		return false
	}
	r.ctx.Stroke()
	r.markEmergencyTextFallback()
	return true
}

func rotatedTextOrigin(anchor geom.Pt, metrics render.TextMetrics, bounds render.TextBounds, haveBounds bool) geom.Pt {
	if haveBounds && bounds.W > 0 && bounds.H > 0 {
		return geom.Pt{
			X: anchor.X - (bounds.X + bounds.W/2),
			Y: anchor.Y - (bounds.Y + bounds.H),
		}
	}

	return geom.Pt{
		X: anchor.X - metrics.W/2,
		Y: anchor.Y - metrics.Descent,
	}
}

func (r *Renderer) textDrawBounds(text string, origin geom.Pt, size float64, fontKey string) (geom.Rect, bool) {
	if text == "" || size <= 0 {
		return geom.Rect{}, false
	}
	if bounds, ok := r.MeasureTextBounds(text, size, fontKey); ok && bounds.W > 0 && bounds.H > 0 {
		return geom.Rect{
			Min: geom.Pt{X: origin.X + bounds.X, Y: origin.Y + bounds.Y},
			Max: geom.Pt{X: origin.X + bounds.X + bounds.W, Y: origin.Y + bounds.Y + bounds.H},
		}.Inflate(2, 2), true
	}
	metrics := r.MeasureText(text, size, fontKey)
	if metrics.W <= 0 || metrics.H <= 0 {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent},
		Max: geom.Pt{X: origin.X + metrics.W, Y: origin.Y + metrics.Descent},
	}.Inflate(2, 2), true
}

func (r *Renderer) rotatedTextDrawBounds(text string, anchor geom.Pt, size, angle float64, fontKey string) (geom.Rect, bool) {
	if text == "" || size <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return geom.Rect{}, false
	}
	metrics := r.MeasureText(text, size, fontKey)
	if metrics.W <= 0 || metrics.H <= 0 {
		return geom.Rect{}, false
	}
	bounds, haveBounds := r.MeasureTextBounds(text, size, fontKey)
	origin := rotatedTextOrigin(anchor, metrics, bounds, haveBounds)
	var rect geom.Rect
	if haveBounds && bounds.W > 0 && bounds.H > 0 {
		rect = geom.Rect{
			Min: geom.Pt{X: origin.X + bounds.X, Y: origin.Y + bounds.Y},
			Max: geom.Pt{X: origin.X + bounds.X + bounds.W, Y: origin.Y + bounds.Y + bounds.H},
		}
	} else {
		rect = geom.Rect{
			Min: geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent},
			Max: geom.Pt{X: origin.X + metrics.W, Y: origin.Y + metrics.Descent},
		}
	}
	return rotatedRectBounds(rect, anchor, -angle)
}

func rotatedRectBounds(rect geom.Rect, anchor geom.Pt, angle float64) (geom.Rect, bool) {
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	rotate := func(pt geom.Pt) geom.Pt {
		dx := pt.X - anchor.X
		dy := pt.Y - anchor.Y
		return geom.Pt{
			X: anchor.X + dx*cosA - dy*sinA,
			Y: anchor.Y + dx*sinA + dy*cosA,
		}
	}
	return pointsBounds([]geom.Pt{
		rotate(rect.Min),
		rotate(geom.Pt{X: rect.Max.X, Y: rect.Min.Y}),
		rotate(rect.Max),
		rotate(geom.Pt{X: rect.Min.X, Y: rect.Max.Y}),
	}, 3)
}

// DrawTextVertical renders text vertically (one character per line, top to bottom).
// This is used for ylabel rendering where true rotation is not available.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	if text == "" || size <= 0 {
		return
	}
	r.drawTextVerticalWithFontContext(text, center, size, textColor, r.lastFontKey)
}

// DrawTextVerticalWithFont renders vertical text with an explicit font key.
func (r *Renderer) DrawTextVerticalWithFont(text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
	if text == "" || size <= 0 {
		return
	}
	r.withTemporaryFontKey(func() {
		r.drawTextVerticalWithFontContext(text, center, size, textColor, fontKey)
	})
}

func (r *Renderer) drawTextVerticalWithFontContext(text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
	if r.hasClipPath() {
		metrics := r.MeasureText(text, size, fontKey)
		if metrics.W <= 0 || metrics.H <= 0 {
			return
		}
		r.ctx.PushTransform()
		defer r.ctx.PopTransform()
		r.ctx.Translate(center.X, center.Y)
		r.ctx.Rotate(-math.Pi / 2)
		r.ctx.Translate(-center.X, -center.Y)
		r.drawTextDirect(text, geom.Pt{X: center.X + metrics.Descent, Y: center.Y}, size, textColor, fontKey)
		return
	}
	r.drawTextRotatedDirect(text, center, size, -math.Pi/2, textColor, fontKey)
}

func (r *Renderer) withTemporaryFontKey(draw func()) {
	previous := r.lastFontKey
	defer func() {
		r.lastFontKey = previous
	}()
	draw()
}

// GetImage returns the rendered image as a standard Go image.RGBA.
func (r *Renderer) GetImage() *image.RGBA {
	return r.ctx.GetImage().ToGoImage()
}

// SavePNG saves the rendered image to a PNG file.
func (r *Renderer) SavePNG(path string) error {
	img := r.GetImage()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func transformMarkerPath(path geom.Path, affine geom.Affine, offset geom.Pt) geom.Path {
	if len(path.C) == 0 {
		return geom.Path{}
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, pt := range path.V {
		pt = affine.Apply(pt)
		out.V[i] = geom.Pt{X: pt.X + offset.X, Y: pt.Y + offset.Y}
	}
	return out
}

func (r *Renderer) drawGouraudTriangle(tri *render.GouraudTriangle) {
	if tri == nil {
		return
	}
	img := r.ctx.image
	if img == nil || img.Width() <= 0 || img.Height() <= 0 {
		return
	}

	minX := int(math.Floor(math.Min(tri.P[0].X, math.Min(tri.P[1].X, tri.P[2].X))))
	maxX := int(math.Ceil(math.Max(tri.P[0].X, math.Max(tri.P[1].X, tri.P[2].X))))
	minY := int(math.Floor(math.Min(tri.P[0].Y, math.Min(tri.P[1].Y, tri.P[2].Y))))
	maxY := int(math.Ceil(math.Max(tri.P[0].Y, math.Max(tri.P[1].Y, tri.P[2].Y))))

	clipMinX, clipMinY := 0, 0
	clipMaxX, clipMaxY := img.Width()-1, img.Height()-1
	if r.clipRect != nil {
		clipMinX = maxInt(clipMinX, int(math.Floor(r.clipRect.Min.X)))
		clipMinY = maxInt(clipMinY, int(math.Floor(r.clipRect.Min.Y)))
		clipMaxX = minInt(clipMaxX, int(math.Ceil(r.clipRect.Max.X))-1)
		clipMaxY = minInt(clipMaxY, int(math.Ceil(r.clipRect.Max.Y))-1)
	}
	minX = maxInt(minX, clipMinX)
	minY = maxInt(minY, clipMinY)
	maxX = minInt(maxX, clipMaxX)
	maxY = minInt(maxY, clipMaxY)
	if minX > maxX || minY > maxY {
		return
	}

	area := edgeFunction(tri.P[0], tri.P[1], tri.P[2])
	if area == 0 || math.IsNaN(area) || math.IsInf(area, 0) {
		return
	}

	stride := img.Stride()
	if stride <= 0 {
		return
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			p := geom.Pt{X: float64(x) + 0.5, Y: float64(y) + 0.5}
			w0 := edgeFunction(tri.P[1], tri.P[2], p) / area
			w1 := edgeFunction(tri.P[2], tri.P[0], p) / area
			w2 := edgeFunction(tri.P[0], tri.P[1], p) / area
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			src := interpolateColor(tri.Color[0], tri.Color[1], tri.Color[2], w0, w1, w2)
			if src.A <= 0 {
				continue
			}
			off := y*stride + x*4
			if off < 0 || off+3 >= len(img.Data) {
				continue
			}
			blendPixelRGBA(img.Data[off:off+4], src)
		}
	}
}

func edgeFunction(a, b, c geom.Pt) float64 {
	return (c.X-a.X)*(b.Y-a.Y) - (c.Y-a.Y)*(b.X-a.X)
}

func interpolateColor(c0, c1, c2 render.Color, w0, w1, w2 float64) render.Color {
	return render.Color{
		R: c0.R*w0 + c1.R*w1 + c2.R*w2,
		G: c0.G*w0 + c1.G*w1 + c2.G*w2,
		B: c0.B*w0 + c1.B*w1 + c2.B*w2,
		A: c0.A*w0 + c1.A*w1 + c2.A*w2,
	}
}

func blendPixelRGBA(dst []uint8, src render.Color) {
	sa := uint32(math.Round(clamp01(src.A) * 255))
	if sa == 0 {
		return
	}
	sr := uint8(math.Round(clamp01(src.R) * 255))
	sg := uint8(math.Round(clamp01(src.G) * 255))
	sb := uint8(math.Round(clamp01(src.B) * 255))
	if sa >= 255 {
		dst[0] = sr
		dst[1] = sg
		dst[2] = sb
		dst[3] = 255
		return
	}

	da := uint32(dst[3])
	combinedA := ((sa + da) << 8) - sa*da
	if combinedA == 0 {
		dst[0], dst[1], dst[2], dst[3] = 0, 0, 0, 0
		return
	}
	dst[3] = uint8(combinedA >> 8)
	dst[0] = fixedBlendChannel(dst[0], sr, uint8(da), uint8(sa), combinedA)
	dst[1] = fixedBlendChannel(dst[1], sg, uint8(da), uint8(sa), combinedA)
	dst[2] = fixedBlendChannel(dst[2], sb, uint8(da), uint8(sa), combinedA)
}

func fixedBlendChannel(dst, src, dstA, srcA uint8, combinedA uint32) uint8 {
	dstPremul := int64(uint32(dst) * uint32(dstA))
	numerator := ((int64(src) << 8) - dstPremul) * int64(srcA)
	numerator += dstPremul << 8
	return uint8(numerator / int64(combinedA))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// renderColorToAGG converts a normalized render.Color to AGG's 8-bit SRGBA
// color type without applying any transfer-curve conversion.
func renderColorToAGG(c render.Color) agglib.Color {
	return agglib.NewColor(
		uint8(math.Round(clamp01(c.R)*255)),
		uint8(math.Round(clamp01(c.G)*255)),
		uint8(math.Round(clamp01(c.B)*255)),
		uint8(math.Round(clamp01(c.A)*255)),
	)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// quantize snaps values for cache keys and text metrics. Path rasterization
// itself uses explicit snapping/simplification policy instead of this grid.
const quantizationGrid = 1e-6

func quantize(v float64) float64 {
	return math.Round(v/quantizationGrid) * quantizationGrid
}

func quantizePt(p geom.Pt) geom.Pt {
	return geom.Pt{X: quantize(p.X), Y: quantize(p.Y)}
}

// SupportsNativeHatch reports that AGG consumes render.Paint hatch metadata
// directly while rasterizing a path.
func (r *Renderer) SupportsNativeHatch() bool { return true }

func applyForcedAlpha(paint *render.Paint) {
	if paint == nil || !paint.ForceAlpha {
		return
	}
	alpha := clamp01(paint.Alpha)
	if paint.Stroke.A > 0 {
		paint.Stroke.A = alpha
	}
	if paint.Fill.A > 0 {
		paint.Fill.A = alpha
	}
	if paint.HatchColor.A > 0 {
		paint.HatchColor.A = alpha
	}
}

func colorWithForcedAlpha(c render.Color, paint *render.Paint) render.Color {
	if paint != nil && paint.ForceAlpha && c.A > 0 {
		c.A = clamp01(paint.Alpha)
	}
	return c
}

func (r *Renderer) applyAntialiasMode(mode render.AntialiasMode) func() {
	if r.ctx == nil {
		return func() {}
	}
	prev := r.ctx.GetAntiAliasGamma()
	switch mode {
	case render.AntialiasOn:
		r.ctx.SetAntiAliasGamma(1.0)
	case render.AntialiasOff:
		// AGG exposes antialiasing through the rasterizer gamma curve rather
		// than a boolean switch. A low gamma sharply suppresses partial
		// coverage and gives callers an aliased-style path when requested.
		r.ctx.SetAntiAliasGamma(0.1)
	default:
		return func() {}
	}
	return func() {
		r.ctx.SetAntiAliasGamma(prev)
	}
}

const defaultPathChunkVertices = 32768

func (r *Renderer) preparePathForPaint(path geom.Path, paint *render.Paint) (geom.Path, bool) {
	path = removeNonFinitePathVertices(path)
	if len(path.C) == 0 || !path.Validate() {
		return geom.Path{}, false
	}
	if r.pathOutsideVisibleArea(path, paint) {
		return geom.Path{}, false
	}
	if shouldSnapPath(path, paint) {
		path = snapPath(path, paint)
	}
	if paint.Simplify && paint.SimplifyThreshold > 0 {
		path = simplifyLinePath(path, paint.SimplifyThreshold)
	}
	return path, len(path.C) > 0
}

func removeNonFinitePathVertices(path geom.Path) geom.Path {
	out := geom.Path{}
	vi := 0
	haveCurrent := false
	needMove := true

	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return out
			}
			to := path.V[vi]
			vi++
			if !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			out.MoveTo(to)
			haveCurrent = true
			needMove = false
		case geom.LineTo:
			if vi >= len(path.V) {
				return out
			}
			to := path.V[vi]
			vi++
			if !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			if !haveCurrent || needMove {
				out.MoveTo(to)
			} else {
				out.LineTo(to)
			}
			haveCurrent = true
			needMove = false
		case geom.QuadTo:
			if vi+1 >= len(path.V) {
				return out
			}
			ctrl, to := path.V[vi], path.V[vi+1]
			vi += 2
			if !finitePt(ctrl) || !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			if !haveCurrent || needMove {
				out.MoveTo(to)
			} else {
				out.QuadTo(ctrl, to)
			}
			haveCurrent = true
			needMove = false
		case geom.CubicTo:
			if vi+2 >= len(path.V) {
				return out
			}
			c1, c2, to := path.V[vi], path.V[vi+1], path.V[vi+2]
			vi += 3
			if !finitePt(c1) || !finitePt(c2) || !finitePt(to) {
				haveCurrent = false
				needMove = true
				continue
			}
			if !haveCurrent || needMove {
				out.MoveTo(to)
			} else {
				out.CubicTo(c1, c2, to)
			}
			haveCurrent = true
			needMove = false
		case geom.ClosePath:
			if haveCurrent && !needMove {
				out.Close()
			}
			haveCurrent = false
			needMove = true
		}
	}
	return out
}

func finitePt(pt geom.Pt) bool {
	return !math.IsNaN(pt.X) && !math.IsInf(pt.X, 0) && !math.IsNaN(pt.Y) && !math.IsInf(pt.Y, 0)
}

func (r *Renderer) pathOutsideVisibleArea(path geom.Path, paint *render.Paint) bool {
	bounds, ok := pathBounds(path)
	if !ok {
		return true
	}
	visible := geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: float64(r.width), Y: float64(r.height)}}
	if r.viewport != (geom.Rect{}) {
		visible = visible.Intersect(r.viewport)
	}
	if r.clipRect != nil {
		visible = visible.Intersect(*r.clipRect)
	}
	if visible.W() <= 0 || visible.H() <= 0 {
		return true
	}

	pad := 1.0
	if paint != nil && paint.Stroke.A > 0 && paint.LineWidth > 0 {
		pad += paint.LineWidth / 2
	}
	return !rectsOverlap(bounds.Inflate(pad, pad), visible)
}

func pathBounds(path geom.Path) (geom.Rect, bool) {
	if len(path.V) == 0 {
		return geom.Rect{}, false
	}
	var bounds geom.Rect
	ok := false
	for _, pt := range path.V {
		if !finitePt(pt) {
			continue
		}
		if !ok {
			bounds = geom.Rect{Min: pt, Max: pt}
			ok = true
			continue
		}
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}
	return bounds, ok
}

func pathDrawBounds(path geom.Path, paint *render.Paint) (geom.Rect, bool) {
	if paint == nil {
		return geom.Rect{}, false
	}
	bounds, ok := pathBounds(path)
	if !ok {
		return geom.Rect{}, false
	}
	pad := 2.0
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		pad += paint.LineWidth / 2
	}
	if paint.Hatch != "" && paint.HatchLineWidth > 0 {
		pad = math.Max(pad, paint.HatchLineWidth/2+2)
	}
	return bounds.Inflate(pad, pad), true
}

func imageDrawBounds(dst geom.Rect) (geom.Rect, bool) {
	minX := math.Min(dst.Min.X, dst.Max.X)
	minY := math.Min(dst.Min.Y, dst.Max.Y)
	maxX := math.Max(dst.Min.X, dst.Max.X)
	maxY := math.Max(dst.Min.Y, dst.Max.Y)
	if minX == maxX || minY == maxY {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}.Inflate(2, 2), true
}

func transformedImageDrawBounds(img render.Image, affine geom.Affine) (geom.Rect, bool) {
	if img == nil {
		return geom.Rect{}, false
	}
	w, h := img.Size()
	if w <= 0 || h <= 0 {
		return geom.Rect{}, false
	}
	return pointsBounds([]geom.Pt{
		affine.Apply(geom.Pt{}),
		affine.Apply(geom.Pt{X: float64(w)}),
		affine.Apply(geom.Pt{Y: float64(h)}),
		affine.Apply(geom.Pt{X: float64(w), Y: float64(h)}),
	}, 2)
}

func imageTransformDisplaySpan(img render.Image, affine geom.Affine) (float64, float64) {
	w, h := img.Size()
	if w <= 0 || h <= 0 {
		return 0, 0
	}

	bounds, ok := pointsBounds([]geom.Pt{
		affine.Apply(geom.Pt{}),
		affine.Apply(geom.Pt{X: float64(w)}),
		affine.Apply(geom.Pt{Y: float64(h)}),
		affine.Apply(geom.Pt{X: float64(w), Y: float64(h)}),
	}, 0)
	if !ok {
		return 0, 0
	}
	return bounds.W(), bounds.H()
}

func gouraudTriangleBatchBounds(batch render.GouraudTriangleBatch) (geom.Rect, bool) {
	points := make([]geom.Pt, 0, len(batch.Triangles)*3)
	for i := range batch.Triangles {
		points = append(points, batch.Triangles[i].P[:]...)
	}
	return pointsBounds(points, 1)
}

func pointsBounds(points []geom.Pt, pad float64) (geom.Rect, bool) {
	var bounds geom.Rect
	ok := false
	for _, pt := range points {
		if !finitePt(pt) {
			continue
		}
		if !ok {
			bounds = geom.Rect{Min: pt, Max: pt}
			ok = true
			continue
		}
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}
	if !ok {
		return geom.Rect{}, false
	}
	return bounds.Inflate(pad, pad), true
}

func rectsOverlap(a, b geom.Rect) bool {
	return a.Max.X >= b.Min.X && b.Max.X >= a.Min.X && a.Max.Y >= b.Min.Y && b.Max.Y >= a.Min.Y
}

func shouldSnapPath(path geom.Path, paint *render.Paint) bool {
	switch paint.Snap {
	case render.SnapOn:
		return true
	case render.SnapOff:
		return false
	case render.SnapAuto:
	default:
		return false
	}
	if len(path.V) > 1024 {
		return false
	}
	vi := 0
	var last geom.Pt
	haveLast := false
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return false
			}
			last = path.V[vi]
			vi++
			haveLast = true
		case geom.LineTo:
			if vi >= len(path.V) {
				return false
			}
			to := path.V[vi]
			vi++
			if haveLast && math.Abs(last.X-to.X) >= 1e-4 && math.Abs(last.Y-to.Y) >= 1e-4 {
				return false
			}
			last = to
			haveLast = true
		case geom.QuadTo, geom.CubicTo:
			return false
		case geom.ClosePath:
			haveLast = false
		}
	}
	return true
}

func snapPath(path geom.Path, paint *render.Paint) geom.Path {
	out := clonePath(path)
	snapValue := 0.0
	strokeWidth := 0.0
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		strokeWidth = paint.LineWidth
	}
	if int(math.Round(strokeWidth))%2 != 0 {
		snapValue = 0.5
	}
	for i, pt := range out.V {
		out.V[i] = geom.Pt{
			X: math.Floor(pt.X+0.5) + snapValue,
			Y: math.Floor(pt.Y+0.5) + snapValue,
		}
	}
	return out
}

func simplifyLinePath(path geom.Path, threshold float64) geom.Path {
	if threshold <= 0 || pathHasCurvesOrClose(path) {
		return path
	}
	out := geom.Path{}
	var current []geom.Pt
	flush := func() {
		if len(current) == 0 {
			return
		}
		points := simplifyPolyline(current, threshold)
		if len(points) > 0 {
			out.MoveTo(points[0])
			for _, pt := range points[1:] {
				out.LineTo(pt)
			}
		}
		current = current[:0]
	}

	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			flush()
			current = append(current, path.V[vi])
			vi++
		case geom.LineTo:
			current = append(current, path.V[vi])
			vi++
		}
	}
	flush()
	return out
}

func pathHasCurvesOrClose(path geom.Path) bool {
	for _, cmd := range path.C {
		if cmd == geom.QuadTo || cmd == geom.CubicTo || cmd == geom.ClosePath {
			return true
		}
	}
	return false
}

func simplifyPolyline(points []geom.Pt, threshold float64) []geom.Pt {
	if len(points) <= 2 {
		return append([]geom.Pt(nil), points...)
	}
	keep := make([]bool, len(points))
	keep[0] = true
	keep[len(points)-1] = true
	simplifyPolylineRange(points, threshold*threshold, 0, len(points)-1, keep)
	out := make([]geom.Pt, 0, len(points))
	for i, pt := range points {
		if keep[i] {
			out = append(out, pt)
		}
	}
	return out
}

func simplifyPolylineRange(points []geom.Pt, threshold2 float64, first, last int, keep []bool) {
	if last <= first+1 {
		return
	}
	maxDist2 := -1.0
	maxIndex := -1
	for i := first + 1; i < last; i++ {
		dist2 := pointSegmentDistanceSquared(points[i], points[first], points[last])
		if dist2 > maxDist2 {
			maxDist2 = dist2
			maxIndex = i
		}
	}
	if maxDist2 > threshold2 && maxIndex >= 0 {
		keep[maxIndex] = true
		simplifyPolylineRange(points, threshold2, first, maxIndex, keep)
		simplifyPolylineRange(points, threshold2, maxIndex, last, keep)
	}
}

func pointSegmentDistanceSquared(p, a, b geom.Pt) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if dx == 0 && dy == 0 {
		return squaredDistance(p, a)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	proj := geom.Pt{X: a.X + t*dx, Y: a.Y + t*dy}
	return squaredDistance(p, proj)
}

func squaredDistance(a, b geom.Pt) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}

func chunkStrokePath(path geom.Path, maxVertices int) []geom.Path {
	if maxVertices <= 0 {
		maxVertices = defaultPathChunkVertices
	}
	if len(path.V) <= maxVertices || pathHasCurvesOrClose(path) {
		return []geom.Path{path}
	}

	chunks := make([]geom.Path, 0, len(path.V)/maxVertices+1)
	vi := 0
	var current geom.Path
	currentVertices := 0
	haveCurrent := false
	var last geom.Pt

	flush := func() {
		if len(current.C) > 1 {
			chunks = append(chunks, current)
		}
		current = geom.Path{}
		currentVertices = 0
		haveCurrent = false
	}

	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				flush()
				return chunks
			}
			if currentVertices >= maxVertices {
				flush()
			}
			last = path.V[vi]
			vi++
			current.MoveTo(last)
			currentVertices++
			haveCurrent = true
		case geom.LineTo:
			if vi >= len(path.V) {
				flush()
				return chunks
			}
			to := path.V[vi]
			vi++
			if !haveCurrent {
				current.MoveTo(to)
				currentVertices++
			} else if currentVertices >= maxVertices {
				flush()
				current.MoveTo(last)
				currentVertices++
			}
			current.LineTo(to)
			currentVertices++
			last = to
			haveCurrent = true
		}
	}
	flush()
	if len(chunks) == 0 {
		return []geom.Path{path}
	}
	return chunks
}

func (r *Renderer) drawNativeHatch(clipPath geom.Path, paint *render.Paint) {
	if paint == nil || paint.Hatch == "" {
		return
	}
	color := colorWithForcedAlpha(paint.HatchColor, paint)
	if color.A <= 0 {
		return
	}
	bounds, ok := pathBounds(clipPath)
	if !ok {
		return
	}
	counts := hatchCounts(paint.Hatch)
	if len(counts) == 0 {
		return
	}

	oldPaths := r.clipPaths
	r.clipPaths = append(clonePaths(oldPaths), clonePath(clipPath))
	defer func() {
		r.clipPaths = oldPaths
	}()

	for pattern, count := range counts {
		spacing := math.Max(2, 32/float64(count))
		if paint.HatchSpacing > 0 {
			spacing = math.Max(2, paint.HatchSpacing/float64(count))
		}
		hatchPaint := render.Paint{
			Stroke:    color,
			LineWidth: paint.HatchLineWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
			Antialias: paint.Antialias,
			Snap:      render.SnapOff,
		}
		if hatchPaint.LineWidth <= 0 {
			hatchPaint.LineWidth = 1
		}
		for _, hatchPath := range hatchPatternPaths(pattern, bounds, spacing) {
			if len(hatchPath.C) == 0 {
				continue
			}
			r.Path(hatchPath, &hatchPaint)
		}
	}
}

func hatchCounts(pattern string) map[rune]int {
	counts := make(map[rune]int)
	for _, ch := range pattern {
		switch ch {
		case '|', '-', '/', '\\', '+', 'x', 'X':
			counts[ch]++
		}
	}
	return counts
}

func hatchPatternPaths(pattern rune, bounds geom.Rect, spacing float64) []geom.Path {
	switch pattern {
	case '|':
		return []geom.Path{verticalHatchPath(bounds, spacing)}
	case '-':
		return []geom.Path{horizontalHatchPath(bounds, spacing)}
	case '/':
		return []geom.Path{slashHatchPath(bounds, spacing)}
	case '\\':
		return []geom.Path{backslashHatchPath(bounds, spacing)}
	case '+':
		return []geom.Path{
			verticalHatchPath(bounds, spacing),
			horizontalHatchPath(bounds, spacing),
		}
	case 'x', 'X':
		return []geom.Path{
			slashHatchPath(bounds, spacing),
			backslashHatchPath(bounds, spacing),
		}
	default:
		return nil
	}
}

func verticalHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	minX := math.Floor(bounds.Min.X/spacing)*spacing - spacing
	maxX := bounds.Max.X + spacing
	for x := minX; x <= maxX; x += spacing {
		path.MoveTo(geom.Pt{X: x, Y: bounds.Min.Y - spacing})
		path.LineTo(geom.Pt{X: x, Y: bounds.Max.Y + spacing})
	}
	return path
}

func horizontalHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	minY := math.Floor(bounds.Min.Y/spacing)*spacing - spacing
	maxY := bounds.Max.Y + spacing
	for y := minY; y <= maxY; y += spacing {
		path.MoveTo(geom.Pt{X: bounds.Min.X - spacing, Y: y})
		path.LineTo(geom.Pt{X: bounds.Max.X + spacing, Y: y})
	}
	return path
}

func slashHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	width := bounds.W()
	height := bounds.H()
	extent := width + height + 2*spacing
	start := bounds.Min.X - height - spacing
	end := bounds.Max.X + spacing
	for x := start; x <= end; x += spacing {
		path.MoveTo(geom.Pt{X: x, Y: bounds.Max.Y + spacing})
		path.LineTo(geom.Pt{X: x + extent, Y: bounds.Min.Y - spacing})
	}
	return path
}

func backslashHatchPath(bounds geom.Rect, spacing float64) geom.Path {
	var path geom.Path
	width := bounds.W()
	height := bounds.H()
	extent := width + height + 2*spacing
	start := bounds.Min.X - height - spacing
	end := bounds.Max.X + spacing
	for x := start; x <= end; x += spacing {
		path.MoveTo(geom.Pt{X: x, Y: bounds.Min.Y - spacing})
		path.LineTo(geom.Pt{X: x + extent, Y: bounds.Max.Y + spacing})
	}
	return path
}

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

	r.compositeClipSurface(temp.image, masks, region)
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
		region.minX = maxInt(region.minX, int(math.Floor(r.clipRect.Min.X)))
		region.minY = maxInt(region.minY, int(math.Floor(r.clipRect.Min.Y)))
		region.maxX = minInt(region.maxX, int(math.Ceil(r.clipRect.Max.X)))
		region.maxY = minInt(region.maxY, int(math.Ceil(r.clipRect.Max.Y)))
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

func (r *Renderer) compositeClipSurface(src *agglib.Image, masks [][]uint8, region pixelRegion) {
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
				blendPixelRGBA(dst.Data[dstOff:dstOff+4], render.Color{
					R: float64(src.Data[srcOff]) / 255,
					G: float64(src.Data[srcOff+1]) / 255,
					B: float64(src.Data[srcOff+2]) / 255,
					A: (float64(sa) / 255) * (float64(maskA) / 255),
				})
			}
		}
	}
	runRowWorkers(region, compositeRows)
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
		minX := math.Floor(r.clipRect.Min.X)
		minY := math.Floor(r.clipRect.Min.Y)
		maxX := math.Ceil(r.clipRect.Max.X)
		maxY := math.Ceil(r.clipRect.Max.Y)
		r.ctx.ClipBox(minX, minY, maxX, maxY)
		return
	}
	r.ctx.ClipBox(0, 0, float64(r.width), float64(r.height))
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

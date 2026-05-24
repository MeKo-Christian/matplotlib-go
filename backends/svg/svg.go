package svg

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cwbudde/matplotlib-go/backends/internal/mixedraster"
	"github.com/cwbudde/matplotlib-go/backends/internal/vectorhatch"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

const (
	quantizationGrid  = 1e-6
	defaultFontHeight = 13.0
)

type state struct {
	clipRect      *geom.Rect
	clipPathDepth int
}

type svgNode struct {
	content string
	// clipIDs lists active clip-path defs in outer-to-inner order. An empty
	// slice means the node is unclipped. The serializer wraps the content in
	// nested <g clip-path="url(#…)"> groups, opening outer-most first.
	clipIDs []string
	// filterIDs lists active SVG filter defs in outer-to-inner order.
	filterIDs []string
}

// clipDef describes one <clipPath> entry in the SVG <defs> block. Exactly one
// of rect or path is populated. Storing both in a single ordered slice keeps
// emission in registration order so def IDs and document order stay aligned.
type clipDef struct {
	id        string
	rect      *geom.Rect
	path      string
	transform string
}

// markerDef caches one marker-path's `d` attribute for `<defs><path id="…"/>`
// reuse via `<use href="#id">`. Marker batches with identical geometry share
// a single def regardless of how the markers are colored.
type markerDef struct {
	id   string
	data string
}

// pathCollectionDef caches a collection item's display-space path for reuse
// via <use>. Collection paths are already positioned in display coordinates.
type pathCollectionDef struct {
	id   string
	data string
}

type hatchDef struct {
	id        string
	hatch     string
	faceColor render.Color
	lineColor render.Color
	lineWidth float64
	spacing   float64
	forced    bool
}

type fontFaceDef struct {
	family string
	data   string
	mime   string
	format string
}

type filterDef struct {
	id     string
	name   string
	radius float64
}

// Renderer implements render.Renderer using SVG path/text recording.
type Renderer struct {
	width      int
	height     int
	viewport   geom.Rect
	began      bool
	stack      []state
	clipRect   *geom.Rect
	resolution uint
	background render.Color

	nodes           []svgNode
	clipDefs        map[string]string // rect-key → clipDef ID
	clipPathDefs    map[string]string // path data → clipDef ID
	clipOrder       []clipDef         // registration-order defs (rects and paths interleaved)
	clipPathStack   []string          // active path-clip IDs in outer-to-inner order
	clipPaths       []geom.Path       // active path clips in display coordinates for mixed raster replay
	clipIDCounter   int
	markerDefs      map[string]string // marker path data → markerDef ID
	markerOrder     []markerDef       // registration-order marker defs
	markerIDCounter int
	pathDefs        map[string]string // collection path data → pathCollectionDef ID
	pathOrder       []pathCollectionDef
	pathIDCounter   int
	hatchDefs       map[string]string // hatch paint metadata → hatchDef ID
	hatchOrder      []hatchDef
	hatchIDCounter  int

	gradientDefs         map[string]string // gradient key → gradientDef ID
	gradientOrder        []gradientDef     // registration-order gradient defs
	gradientIDCounter    int
	patternFillDefs      map[string]string // pattern key → patternFillDef ID
	patternFillOrder     []patternFillDef  // registration-order pattern defs
	patternFillIDCounter int

	fontFaces     map[string]fontFaceDef
	fontFaceOrder []fontFaceDef

	filterDefs      map[string]string // filter metadata → filterDef ID
	filterOrder     []filterDef
	filterIDCounter int
	filterStack     []string

	lastFontKey string
	texManager  *tex.Manager
	texErr      error
	options     render.SVGOptions
	raster      *mixedraster.Session
}

var (
	_ render.Renderer                = (*Renderer)(nil)
	_ render.DPIAware                = (*Renderer)(nil)
	_ render.TextDrawer              = (*Renderer)(nil)
	_ render.RotatedTextDrawer       = (*Renderer)(nil)
	_ render.VerticalTextDrawer      = (*Renderer)(nil)
	_ render.TextPather              = (*Renderer)(nil)
	_ render.TeXMetricer             = (*Renderer)(nil)
	_ render.TeXDrawer               = (*Renderer)(nil)
	_ render.RotatedTeXDrawer        = (*Renderer)(nil)
	_ render.ImageTransformer        = (*Renderer)(nil)
	_ render.ClipPathTransformer     = (*Renderer)(nil)
	_ render.MarkerDrawer            = (*Renderer)(nil)
	_ render.PathCollectionDrawer    = (*Renderer)(nil)
	_ render.NativeHatcher           = (*Renderer)(nil)
	_ render.GradientFiller          = (*Renderer)(nil)
	_ render.PatternFiller           = (*Renderer)(nil)
	_ render.PathEffectFilterDrawer  = (*Renderer)(nil)
	_ render.RasterizationController = (*Renderer)(nil)
	_ render.SVGExporter             = (*Renderer)(nil)
)

// New creates a new SVG renderer with the specified dimensions and background color.
func New(w, h int, bg render.Color) (*Renderer, error) {
	if w <= 0 || h <= 0 {
		return nil, errors.New("svg: width and height must be positive")
	}

	return &Renderer{
		width:           w,
		height:          h,
		background:      bg,
		resolution:      72,
		clipDefs:        map[string]string{},
		clipPathDefs:    map[string]string{},
		markerDefs:      map[string]string{},
		pathDefs:        map[string]string{},
		hatchDefs:       map[string]string{},
		gradientDefs:    map[string]string{},
		patternFillDefs: map[string]string{},
		fontFaces:       map[string]fontFaceDef{},
		filterDefs:      map[string]string{},
		texManager:      tex.NewManager(tex.ManagerConfig{}),
		options:         render.DefaultSVGOptions(),
	}, nil
}

// SetSVGOptions configures SVG-specific draw and serialization behavior.
func (r *Renderer) SetSVGOptions(opts render.SVGOptions) {
	if r == nil {
		return
	}
	r.options = normalizeSVGOptions(opts)
}

// Begin starts a drawing session with the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("Begin called twice")
	}

	r.began = true
	r.viewport = viewport
	r.nodes = nil
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipDefs = map[string]string{}
	r.clipPathDefs = map[string]string{}
	r.clipOrder = nil
	r.clipPathStack = nil
	r.clipPaths = nil
	r.clipIDCounter = 0
	r.markerDefs = map[string]string{}
	r.markerOrder = nil
	r.markerIDCounter = 0
	r.pathDefs = map[string]string{}
	r.pathOrder = nil
	r.pathIDCounter = 0
	r.hatchDefs = map[string]string{}
	r.hatchOrder = nil
	r.hatchIDCounter = 0
	r.fontFaces = map[string]fontFaceDef{}
	r.fontFaceOrder = nil
	r.filterDefs = map[string]string{}
	r.filterOrder = nil
	r.filterIDCounter = 0
	r.filterStack = nil
	r.lastFontKey = ""
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
	return nil
}

// StartRasterized begins a transparent offscreen raster group for mixed output.
func (r *Renderer) StartRasterized(options render.Rasterization) bool {
	if r == nil || !r.began || r.raster != nil {
		return false
	}
	session, ok := mixedraster.Start(r.width, r.height, r.viewport, options, r.resolution, r.clipRect, r.clipPaths)
	if !ok {
		return false
	}
	r.raster = session
	return true
}

// StopRasterized embeds the active raster group as an SVG image.
func (r *Renderer) StopRasterized() bool {
	if r == nil || r.raster == nil {
		return false
	}
	session := r.raster
	r.raster = nil
	img, rect, ok := session.Stop()
	if !ok {
		return false
	}
	r.Image(img, rect)
	return true
}

func (r *Renderer) activeRaster() render.Renderer {
	if r == nil || r.raster == nil {
		return nil
	}
	return r.raster.Renderer()
}

// Save pushes the current graphics state onto the stack.
func (r *Renderer) Save() {
	if rr := r.activeRaster(); rr != nil {
		rr.Save()
		return
	}
	var clipCopy *geom.Rect
	if r.clipRect != nil {
		copyRect := *r.clipRect
		clipCopy = &copyRect
	}
	r.stack = append(r.stack, state{
		clipRect:      clipCopy,
		clipPathDepth: len(r.clipPathStack),
	})
}

// Restore pops the graphics state from the stack.
func (r *Renderer) Restore() {
	if rr := r.activeRaster(); rr != nil {
		rr.Restore()
		return
	}
	if len(r.stack) == 0 {
		return
	}

	s := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	r.clipRect = s.clipRect
	if s.clipPathDepth < len(r.clipPathStack) {
		r.clipPathStack = r.clipPathStack[:s.clipPathDepth]
	}
	if s.clipPathDepth < len(r.clipPaths) {
		r.clipPaths = r.clipPaths[:s.clipPathDepth]
	}
}

// ClipRect sets a rectangular clip region.
func (r *Renderer) ClipRect(rect geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipRect(rect)
		return
	}
	clipRect := normalizeRect(rect)
	if r.clipRect == nil {
		r.clipRect = &clipRect
		return
	}

	intersected := r.clipRect.Intersect(clipRect)
	r.clipRect = &intersected
}

// ClipPath pushes a path-based clip region onto the active clip stack. Each
// active path clip becomes its own <clipPath> def and the rendered content is
// wrapped in nested <g clip-path="…"> groups (outer-most first) so SVG's
// natural clip-on-clip composition applies.
func (r *Renderer) ClipPath(p geom.Path) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipPath(p)
		return
	}
	r.clipPath(p, "", p)
}

// ClipPathTransformed pushes a path-based clip region with an affine transform
// attached to the stored SVG <clipPath> definition. The active content is still
// wrapped in clip groups, matching Matplotlib's approach for transformed text
// and images where applying clip-path directly to the element would transform
// the clip again.
func (r *Renderer) ClipPathTransformed(p geom.Path, transform geom.Affine) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipPath(mixedraster.ApplyAffine(p, transform))
		return
	}
	r.clipPath(p, matrixTransform(transform), mixedraster.ApplyAffine(p, transform))
}

func (r *Renderer) clipPath(p geom.Path, transform string, rasterPath geom.Path) {
	if !p.Validate() {
		return
	}
	d := buildPathData(p)
	if d == "" {
		return
	}
	id := r.registerPathClip(d, transform)
	r.clipPathStack = append(r.clipPathStack, id)
	r.clipPaths = append(r.clipPaths, mixedraster.ClonePath(rasterPath))
}

// Path draws a path with the given paint style.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if rr := r.activeRaster(); rr != nil {
		rr.Path(p, paint)
		return
	}
	if !p.Validate() || paint == nil {
		return
	}
	if render.DrawPathWithEffects(r, p, paint, r.Path) {
		return
	}

	d := buildPathData(p)
	if d == "" {
		return
	}

	hasGradient := paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0
	hasPattern := paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0
	hasFill := paint.Fill.A > 0 || hasGradient || hasPattern
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasHatch && !hasStroke {
		return
	}

	var b strings.Builder
	b.WriteString(`<path`)
	writeAttr(&b, "d", d)
	writeForcedOpacity(&b, *paint)

	switch {
	case hasHatch:
		writeAttr(&b, "fill", "url(#"+r.registerHatch(*paint)+")")
	case hasGradient:
		writeAttr(&b, "fill", "url(#"+r.registerGradient(&paint.FillGradient)+")")
	case hasPattern:
		writeAttr(&b, "fill", "url(#"+r.registerPatternFill(&paint.FillPattern)+")")
	case paint.Fill.A > 0:
		writeColorAttrs(&b, "fill", paint.Fill, forcedOpacity(*paint))
	default:
		writeAttr(&b, "fill", "none")
	}

	writeStrokeAttrs(&b, *paint)

	b.WriteString(" />")

	r.nodes = append(r.nodes, svgNode{
		content:   b.String(),
		clipIDs:   r.currentClipIDs(),
		filterIDs: r.currentFilterIDs(),
	})
}

// DrawPathWithEffects applies renderer-neutral path effect passes.
func (r *Renderer) DrawPathWithEffects(p geom.Path, paint *render.Paint) bool {
	if rr := r.activeRaster(); rr != nil {
		if effects, ok := rr.(render.PathEffectDrawer); ok {
			return effects.DrawPathWithEffects(p, paint)
		}
		rr.Path(p, paint)
		return true
	}
	return render.DrawPathWithEffects(r, p, paint, r.Path)
}

// DrawPathEffectFilter renders supported filter effects as native SVG filters.
func (r *Renderer) DrawPathEffectFilter(path geom.Path, paint render.Paint, effect render.PathEffect, draw func(geom.Path, *render.Paint)) bool {
	if r == nil || draw == nil {
		return false
	}
	id, ok := r.registerPathEffectFilter(effect)
	if !ok {
		return false
	}
	r.filterStack = append(r.filterStack, id)
	draw(path, &paint)
	r.filterStack = r.filterStack[:len(r.filterStack)-1]
	return true
}

// SupportsPathEffectFilter reports whether SVG can render the filter effect
// without falling back to mixed raster output.
func (r *Renderer) SupportsPathEffectFilter(effect render.PathEffect) bool {
	_, ok := normalizePathEffectFilter(effect)
	return ok
}

// Image draws an image within the destination rectangle.
func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.Image(img, dst)
		return
	}
	rgba := asRGBAImage(img)
	if rgba == nil {
		return
	}
	r.renderImageNode(rgba, dst, "")
}

// ImageTransformed draws an image with an arbitrary affine transform applied
// to the destination rectangle. The transform is emitted as an SVG
// matrix(a b c d e f) so viewers reproduce the placement and skew/rotation
// exactly, without rasterizing the image first.
func (r *Renderer) ImageTransformed(img render.Image, dst geom.Rect, transform geom.Affine) {
	if rr := r.activeRaster(); rr != nil {
		if tr, ok := rr.(render.ImageTransformer); ok {
			tr.ImageTransformed(img, dst, transform)
		} else {
			rr.Image(img, dst)
		}
		return
	}
	rgba := asRGBAImage(img)
	if rgba == nil {
		return
	}
	r.renderImageNode(rgba, dst, matrixTransform(transform))
}

// matrixTransform formats a geom.Affine as an SVG matrix(a b c d e f) string.
// The convention matches SVG's: (x', y') = (a*x + c*y + e, b*x + d*y + f).
func matrixTransform(a geom.Affine) string {
	return "matrix(" + shortFloat(a.A) + " " + shortFloat(a.B) + " " +
		shortFloat(a.C) + " " + shortFloat(a.D) + " " +
		shortFloat(a.E) + " " + shortFloat(a.F) + ")"
}

func (r *Renderer) renderImageNode(rgba *image.RGBA, dst geom.Rect, transform string) {
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

	encoded, err := encodeImage(rgba)
	if err != nil {
		return
	}

	uri := "data:image/png;base64," + encoded

	var b strings.Builder
	b.WriteString(`<image x="`)
	b.WriteString(formatFloat(x))
	b.WriteString(`" y="`)
	b.WriteString(formatFloat(y))
	b.WriteString(`" width="`)
	b.WriteString(formatFloat(w))
	b.WriteString(`" height="`)
	b.WriteString(formatFloat(h))
	b.WriteString(`" preserveAspectRatio="none"`)
	b.WriteString(` href="`)
	b.WriteString(uri)
	b.WriteString(`" xlink:href="`)
	b.WriteString(uri)
	b.WriteString(`"`)
	if transform != "" {
		writeAttr(&b, "transform", transform)
	}
	b.WriteString(` />`)

	r.nodes = append(r.nodes, svgNode{
		content:   b.String(),
		clipIDs:   r.currentClipIDs(),
		filterIDs: r.currentFilterIDs(),
	})
}

// DrawMarkers renders a single marker geometry at many display-space positions
// using SVG's `<defs><path id="…"/></defs>` + `<use href="#…">` idiom. The
// marker path is registered once per unique geometry (so a 1000-point scatter
// with a circular marker emits one <path> def and 1000 short <use> tags), and
// each `<use>` carries the per-item matrix transform plus paint attributes.
func (r *Renderer) DrawMarkers(batch render.MarkerBatch) bool {
	if rr := r.activeRaster(); rr != nil {
		if markers, ok := rr.(render.MarkerDrawer); ok {
			return markers.DrawMarkers(batch)
		}
		return false
	}
	if len(batch.Marker.C) == 0 || len(batch.Items) == 0 {
		return false
	}
	if !batch.Marker.Validate() {
		return false
	}
	d := buildPathData(batch.Marker)
	if d == "" {
		return false
	}
	markerID := r.registerMarker(d)

	var b strings.Builder
	emitted := 0
	for i := range batch.Items {
		item := &batch.Items[i]
		paint := item.Paint
		hasFill := paint.Fill.A > 0
		hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
		if !hasFill && !hasStroke {
			continue
		}

		// Combined transform: apply per-item Transform first, then translate by Offset.
		t := geom.Affine{
			A: item.Transform.A,
			B: item.Transform.B,
			C: item.Transform.C,
			D: item.Transform.D,
			E: item.Transform.E + item.Offset.X,
			F: item.Transform.F + item.Offset.Y,
		}

		b.WriteString(`<use href="#`)
		b.WriteString(markerID)
		b.WriteString(`" xlink:href="#`)
		b.WriteString(markerID)
		b.WriteString(`"`)
		writeAttr(&b, "transform", matrixTransform(t))
		writeForcedOpacity(&b, paint)

		if hasFill {
			writeColorAttrs(&b, "fill", paint.Fill, forcedOpacity(paint))
		} else {
			writeAttr(&b, "fill", "none")
		}

		writeStrokeAttrs(&b, paint)

		b.WriteString(" />")
		emitted++
	}

	if emitted == 0 {
		return true
	}

	r.nodes = append(r.nodes, svgNode{
		content:   b.String(),
		clipIDs:   r.currentClipIDs(),
		filterIDs: r.currentFilterIDs(),
	})
	return true
}

func (r *Renderer) registerMarker(d string) string {
	if id, ok := r.markerDefs[d]; ok {
		return id
	}

	r.markerIDCounter++
	id := r.defID("marker", d, r.markerIDCounter)
	r.markerDefs[d] = id
	r.markerOrder = append(r.markerOrder, markerDef{id: id, data: d})
	return id
}

// DrawPathCollection renders display-space paths using SVG defs plus use
// elements. Identical path geometry is registered once and reused with
// per-item paint attributes.
func (r *Renderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	if rr := r.activeRaster(); rr != nil {
		if paths, ok := rr.(render.PathCollectionDrawer); ok {
			return paths.DrawPathCollection(batch)
		}
		return false
	}
	if len(batch.Items) == 0 {
		return false
	}

	var b strings.Builder
	emitted := 0
	for i := range batch.Items {
		item := &batch.Items[i]
		if !item.Path.Validate() {
			continue
		}
		d := buildPathData(item.Path)
		if d == "" {
			continue
		}

		paint := item.Paint
		if item.Hatch != "" {
			paint.Hatch = item.Hatch
			paint.HatchColor = item.HatchColor
			paint.HatchLineWidth = item.HatchWidth
			paint.HatchSpacing = item.HatchSpacing
		}
		hasFill := paint.Fill.A > 0
		hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
		hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
		if !hasFill && !hasHatch && !hasStroke {
			continue
		}

		pathID := r.registerCollectionPath(d)
		b.WriteString(`<use href="#`)
		b.WriteString(pathID)
		b.WriteString(`" xlink:href="#`)
		b.WriteString(pathID)
		b.WriteString(`"`)
		writeForcedOpacity(&b, paint)
		if hasHatch {
			writeAttr(&b, "fill", "url(#"+r.registerHatch(paint)+")")
		} else if hasFill {
			writeColorAttrs(&b, "fill", paint.Fill, forcedOpacity(paint))
		} else {
			writeAttr(&b, "fill", "none")
		}
		writeStrokeAttrs(&b, paint)
		b.WriteString(" />")
		emitted++
	}

	if emitted == 0 {
		return true
	}

	r.nodes = append(r.nodes, svgNode{
		content:   b.String(),
		clipIDs:   r.currentClipIDs(),
		filterIDs: r.currentFilterIDs(),
	})
	return true
}

func (r *Renderer) registerCollectionPath(d string) string {
	if id, ok := r.pathDefs[d]; ok {
		return id
	}

	r.pathIDCounter++
	id := r.defID("pathcoll", d, r.pathIDCounter)
	r.pathDefs[d] = id
	r.pathOrder = append(r.pathOrder, pathCollectionDef{id: id, data: d})
	return id
}

func (r *Renderer) SupportsNativeHatch() bool { return true }

func (r *Renderer) registerHatch(paint render.Paint) string {
	lineWidth := paint.HatchLineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	spacing := paint.HatchSpacing
	if spacing <= 0 {
		spacing = 8
	}
	forced := forcedOpacity(paint)
	key := hatchKey(paint.Hatch, paint.Fill, paint.HatchColor, lineWidth, spacing, forced)
	if id, ok := r.hatchDefs[key]; ok {
		return id
	}

	r.hatchIDCounter++
	id := r.defID("hatch", key, r.hatchIDCounter)
	r.hatchDefs[key] = id
	r.hatchOrder = append(r.hatchOrder, hatchDef{
		id:        id,
		hatch:     paint.Hatch,
		faceColor: paint.Fill,
		lineColor: paint.HatchColor,
		lineWidth: lineWidth,
		spacing:   spacing,
		forced:    forced,
	})
	return id
}

// GlyphRun draws a run of glyph IDs as characters where available.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if len(run.Glyphs) == 0 {
		return
	}

	if run.FontKey != "" {
		r.lastFontKey = run.FontKey
	}

	penX := run.Origin.X
	penY := run.Origin.Y

	size := run.Size
	if size <= 0 {
		size = 12
	}

	for _, glyph := range run.Glyphs {
		if glyph.ID == 0 {
			if glyph.Advance > 0 {
				penX += glyph.Advance
			}
			continue
		}

		r.DrawText(string(rune(glyph.ID)), geom.Pt{X: penX + glyph.Offset.X, Y: penY + glyph.Offset.Y}, size, textColor)

		advance := glyph.Advance
		if advance <= 0 {
			advance = r.MeasureText(string(rune(glyph.ID)), size, run.FontKey).W
		}
		penX += advance
	}
}

// MeasureText returns text metrics based on a built-in monospace-compatible font.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" || size <= 0 {
		return render.TextMetrics{}
	}
	if fontKey != "" {
		r.lastFontKey = fontKey
	}

	scale := size / defaultFontHeight
	if scale <= 0 {
		return render.TextMetrics{}
	}

	face := basicfont.Face7x13
	width := float64(font.MeasureString(face, text).Ceil())
	height := float64(face.Metrics().Height.Ceil())
	ascent := float64(face.Metrics().Ascent.Ceil())
	desc := float64(face.Metrics().Descent.Ceil())

	if width <= 0 || height <= 0 {
		return render.TextMetrics{}
	}

	return render.TextMetrics{
		W:       quantize(width * scale),
		H:       quantize(height * scale),
		Ascent:  quantize(ascent * scale),
		Descent: quantize(desc * scale),
	}
}

// DrawText renders text using an SVG <text> element.
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, origin, size, textColor)
		}
		return
	}
	if text == "" || size <= 0 {
		return
	}

	r.renderTextNode(text, origin.X, origin.Y, size, textColor, "", geom.Affine{}, false)
}

// DrawTextRotated renders text using Matplotlib-like anchor rotation. The
// anchor is the bottom-center of the unrotated text box.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.RotatedTextDrawer); ok {
			textRen.DrawTextRotated(text, anchor, size, angle, textColor)
		} else if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, anchor, size, textColor)
		}
		return
	}
	if text == "" || size <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return
	}

	metrics := r.MeasureText(text, size, "")
	if metrics.W <= 0 || metrics.H <= 0 {
		return
	}

	origin := geom.Pt{
		X: anchor.X - metrics.W/2,
		Y: anchor.Y - metrics.Descent,
	}
	affine := rotationAffine(-angle*180/math.Pi, anchor.X, anchor.Y)
	r.renderTextNode(text, origin.X, origin.Y, size, textColor, matrixTransform(affine), affine, true)
}

// DrawTextVertical renders one character per line.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	if rr := r.activeRaster(); rr != nil {
		if textRen, ok := rr.(render.VerticalTextDrawer); ok {
			textRen.DrawTextVertical(text, center, size, textColor)
		} else if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, center, size, textColor)
		}
		return
	}
	if text == "" || size <= 0 {
		return
	}

	runes := []rune(text)
	lineMetrics := r.MeasureText("M", size, "")
	lineH := lineMetrics.H
	if lineH <= 0 {
		lineH = size
	}

	totalH := lineH * float64(len(runes))
	y := center.Y - totalH/2 + lineMetrics.Ascent

	for _, ch := range runes {
		s := string(ch)
		chMetrics := r.MeasureText(s, size, "")
		if chMetrics.W <= 0 || chMetrics.H <= 0 {
			continue
		}

		x := center.X - chMetrics.W/2
		r.renderTextNode(s, x, y, size, textColor, "", geom.Affine{}, false)
		y += lineH
	}
}

// TextPath converts text to a vector path using the shared font manager.
func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if fontKey == "" {
		fontKey = r.lastFontKey
	}
	return render.TextPath(text, origin, size, fontKey)
}

// LastTeXError returns the most recent TeX pipeline error recorded by MeasureTeX
// or DrawTeX. A nil value means the last TeX operation succeeded.
func (r *Renderer) LastTeXError() error {
	if r == nil {
		return nil
	}
	return r.texErr
}

// MeasureTeX measures a TeX string by rendering it through the external
// latex+dvipng cache and using the resulting tight PNG dimensions.
func (r *Renderer) MeasureTeX(text string, size float64, fontKey string) (render.TextMetrics, bool) {
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok {
		return render.TextMetrics{}, false
	}
	return result.Metrics, true
}

// DrawTeX embeds a TeX-rendered PNG as an SVG image element.
func (r *Renderer) DrawTeX(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) bool {
	if rr := r.activeRaster(); rr != nil {
		if texRen, ok := rr.(render.TeXDrawer); ok {
			return texRen.DrawTeX(text, origin, size, textColor, fontKey)
		}
		if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, origin, size, textColor)
			return true
		}
		return false
	}
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return false
	}
	topLeft := geom.Pt{X: origin.X, Y: origin.Y - result.Metrics.Ascent}
	r.renderImageNode(img, geom.Rect{
		Min: topLeft,
		Max: geom.Pt{X: topLeft.X + float64(img.Bounds().Dx()), Y: topLeft.Y + float64(img.Bounds().Dy())},
	}, "")
	return true
}

// DrawTeXRotated embeds a TeX-rendered PNG and rotates it around the
// Matplotlib-style text rotation anchor.
func (r *Renderer) DrawTeXRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) bool {
	if rr := r.activeRaster(); rr != nil {
		if texRen, ok := rr.(render.RotatedTeXDrawer); ok {
			return texRen.DrawTeXRotated(text, anchor, size, angle, textColor, fontKey)
		}
		if textRen, ok := rr.(render.RotatedTextDrawer); ok {
			textRen.DrawTextRotated(text, anchor, size, angle, textColor)
			return true
		}
		return false
	}
	if math.IsNaN(angle) || math.IsInf(angle, 0) {
		return false
	}
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return false
	}

	metrics := result.Metrics
	origin := geom.Pt{X: anchor.X - metrics.W/2, Y: anchor.Y - metrics.Descent}
	topLeft := geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent}
	transform := rotateTransform(-angle*180/math.Pi, anchor.X, anchor.Y)
	r.renderImageNode(img, geom.Rect{
		Min: topLeft,
		Max: geom.Pt{X: topLeft.X + float64(img.Bounds().Dx()), Y: topLeft.Y + float64(img.Bounds().Dy())},
	}, transform)
	return true
}

func (r *Renderer) renderTeX(text string, size float64, fontKey string) (tex.RenderResult, bool) {
	if r == nil || text == "" || size <= 0 {
		return tex.RenderResult{}, false
	}
	if r.texManager == nil {
		r.texManager = tex.NewManager(tex.ManagerConfig{})
	}
	result, err := r.texManager.Render(text, size, r.resolution, fontKey)
	if err != nil {
		r.texErr = err
		return tex.RenderResult{}, false
	}
	r.texErr = nil
	return result, true
}

// SetResolution sets raster-free text metric scale basis.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi > 0 {
		r.resolution = dpi
	}
}

// SaveSVG writes all recorded content into an SVG document.
func (r *Renderer) SaveSVG(path string) error {
	return r.SaveSVGWithOptions(path, r.options)
}

// SaveSVGWithOptions writes all recorded content into an SVG document using
// the provided serialization options.
func (r *Renderer) SaveSVGWithOptions(path string, opts render.SVGOptions) error {
	if path == "" {
		return errors.New("svg: path is required")
	}
	r.SetSVGOptions(opts)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(r.renderSVG())
	return err
}

func (r *Renderer) renderSVG() string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("\n")
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"`+"\n"+`width="%s" height="%s" viewBox="0 0 %d %d"`+"\n"+`preserveAspectRatio="xMidYMid meet">`+"\n",
		formatFloat(float64(r.width)),
		formatFloat(float64(r.height)),
		r.width,
		r.height)
	writeMetadata(&b, r.options)

	if len(r.clipOrder) > 0 || len(r.markerOrder) > 0 || len(r.pathOrder) > 0 || len(r.hatchOrder) > 0 || len(r.gradientOrder) > 0 || len(r.patternFillOrder) > 0 || len(r.fontFaceOrder) > 0 || len(r.filterOrder) > 0 {
		b.WriteString("  <defs>\n")
		if len(r.fontFaceOrder) > 0 {
			b.WriteString("    <style type=\"text/css\"><![CDATA[\n")
			for _, face := range r.fontFaceOrder {
				b.WriteString("      @font-face { font-family: \"")
				b.WriteString(face.family)
				b.WriteString("\"; src: url(\"data:")
				b.WriteString(face.mime)
				b.WriteString(";base64,")
				b.WriteString(face.data)
				b.WriteString("\") format(\"")
				b.WriteString(face.format)
				b.WriteString("\"); }\n")
			}
			b.WriteString("    ]]></style>\n")
		}
		for _, filter := range r.filterOrder {
			writeFilterDef(&b, filter)
		}
		for _, clip := range r.clipOrder {
			b.WriteString("    <clipPath id=\"" + clip.id + "\" clipPathUnits=\"userSpaceOnUse\">")
			if clip.rect != nil {
				w := clip.rect.W()
				h := clip.rect.H()
				b.WriteString("<rect x=\"")
				b.WriteString(formatFloat(clip.rect.Min.X))
				b.WriteString(`" y="`)
				b.WriteString(formatFloat(clip.rect.Min.Y))
				b.WriteString(`" width="`)
				b.WriteString(formatFloat(w))
				b.WriteString(`" height="`)
				b.WriteString(formatFloat(h))
				b.WriteString(`" />`)
			} else {
				b.WriteString(`<path d="`)
				b.WriteString(clip.path)
				if clip.transform != "" {
					b.WriteString(`" transform="`)
					b.WriteString(clip.transform)
				}
				b.WriteString(`" />`)
			}
			b.WriteString("</clipPath>\n")
		}
		for _, hatch := range r.hatchOrder {
			writeHatchDef(&b, hatch)
		}
		for i := range r.gradientOrder {
			writeGradientDef(&b, &r.gradientOrder[i])
		}
		for i := range r.patternFillOrder {
			writePatternFillDef(&b, &r.patternFillOrder[i])
		}
		for _, m := range r.markerOrder {
			b.WriteString(`    <path id="`)
			b.WriteString(m.id)
			b.WriteString(`" d="`)
			b.WriteString(m.data)
			b.WriteString(`" />` + "\n")
		}
		for _, p := range r.pathOrder {
			b.WriteString(`    <path id="`)
			b.WriteString(p.id)
			b.WriteString(`" d="`)
			b.WriteString(p.data)
			b.WriteString(`" />` + "\n")
		}
		b.WriteString("  </defs>\n")
	}

	bgColor, bgAlpha := colorToStyle(r.background)
	b.WriteString("  <rect x=\"0\" y=\"0\" width=\"100%\" height=\"100%\" ")
	if bgAlpha <= 0 {
		b.WriteString(`fill="none" />`)
		b.WriteString("\n")
	} else {
		b.WriteString(`fill="`)
		b.WriteString(bgColor)
		b.WriteString(`"`)
		if bgAlpha < 1 {
			b.WriteString(` fill-opacity="`)
			b.WriteString(formatFloat(bgAlpha))
			b.WriteString(`"`)
		}
		b.WriteString(" />\n")
	}

	for _, node := range r.nodes {
		if len(node.clipIDs) == 0 && len(node.filterIDs) == 0 {
			b.WriteString("  ")
			b.WriteString(node.content)
			b.WriteString("\n")
			continue
		}
		b.WriteString("  ")
		for _, id := range node.clipIDs {
			b.WriteString("<g clip-path=\"url(#")
			b.WriteString(id)
			b.WriteString(")\">")
		}
		for _, id := range node.filterIDs {
			b.WriteString("<g filter=\"url(#")
			b.WriteString(id)
			b.WriteString(")\">")
		}
		b.WriteString(node.content)
		for range node.filterIDs {
			b.WriteString("</g>")
		}
		for range node.clipIDs {
			b.WriteString("</g>")
		}
		b.WriteString("\n")
	}

	b.WriteString("</svg>\n")
	return b.String()
}

func (r *Renderer) renderTextNode(text string, x, y, size float64, textColor render.Color, transform string, affine geom.Affine, hasAffine bool) {
	if text == "" || size <= 0 {
		return
	}
	if r.options.FontPolicy == render.SVGFontPolicyPath {
		r.renderTextPathNode(text, geom.Pt{X: x, Y: y}, size, textColor, affine, hasAffine)
		return
	}

	var content strings.Builder
	content.WriteString(`<text`)
	writeFloatAttr(&content, "x", x)
	writeFloatAttr(&content, "y", y)
	writeFloatAttr(&content, "font-size", size)
	writeAttr(&content, "font-family", r.svgFontFamily(r.lastFontKey))
	writeAttr(&content, "fill", colorToHex(textColor))
	alpha := clamp01(textColor.A)
	if alpha < 1 {
		writeFloatAttr(&content, "fill-opacity", alpha)
	}
	if transform != "" {
		writeAttr(&content, "transform", transform)
	}
	content.WriteString(">")
	content.WriteString(escapeText(text))
	content.WriteString("</text>")

	r.nodes = append(r.nodes, svgNode{
		content:   content.String(),
		clipIDs:   r.currentClipIDs(),
		filterIDs: r.currentFilterIDs(),
	})
}

func (r *Renderer) renderTextPathNode(text string, origin geom.Pt, size float64, textColor render.Color, affine geom.Affine, hasAffine bool) {
	path, ok := r.TextPath(text, origin, size, r.lastFontKey)
	if !ok {
		return
	}
	if hasAffine {
		path = affinePath(path, affine)
	}
	r.Path(path, &render.Paint{Fill: textColor})
}

// currentClipIDs returns the active clip-path chain in outer-to-inner order.
// Returns nil when no clip is active.
func (r *Renderer) currentClipIDs() []string {
	count := len(r.clipPathStack)
	if r.clipRect != nil {
		count++
	}
	if count == 0 {
		return nil
	}

	ids := make([]string, 0, count)
	if r.clipRect != nil {
		ids = append(ids, r.registerClip(*r.clipRect))
	}
	ids = append(ids, r.clipPathStack...)
	return ids
}

func (r *Renderer) currentFilterIDs() []string {
	if len(r.filterStack) == 0 {
		return nil
	}
	out := make([]string, len(r.filterStack))
	copy(out, r.filterStack)
	return out
}

func (r *Renderer) registerClip(rect geom.Rect) string {
	key := clipKey(rect)
	if id, ok := r.clipDefs[key]; ok {
		return id
	}

	r.clipIDCounter++
	id := r.defID("clip", key, r.clipIDCounter)
	r.clipDefs[key] = id
	rectCopy := rect
	r.clipOrder = append(r.clipOrder, clipDef{id: id, rect: &rectCopy})
	return id
}

func (r *Renderer) registerPathEffectFilter(effect render.PathEffect) (string, bool) {
	name, ok := normalizePathEffectFilter(effect)
	if !ok {
		return "", false
	}
	radius := effect.FilterRadius
	if radius <= 0 {
		radius = 1
	}
	key := filterKey(name, radius)
	if id, ok := r.filterDefs[key]; ok {
		return id, true
	}

	r.filterIDCounter++
	id := r.defID("filter", key, r.filterIDCounter)
	r.filterDefs[key] = id
	r.filterOrder = append(r.filterOrder, filterDef{id: id, name: name, radius: radius})
	return id, true
}

func normalizePathEffectFilter(effect render.PathEffect) (string, bool) {
	name := strings.ToLower(strings.TrimSpace(effect.Filter))
	switch name {
	case "blur", "gaussian", "gaussian-blur", "shadow":
		return "blur", true
	default:
		return "", false
	}
}

func (r *Renderer) registerPathClip(d, transform string) string {
	key := d
	if transform != "" {
		key += "\x00" + transform
	}
	if id, ok := r.clipPathDefs[key]; ok {
		return id
	}

	r.clipIDCounter++
	id := r.defID("clip", key, r.clipIDCounter)
	r.clipPathDefs[key] = id
	r.clipOrder = append(r.clipOrder, clipDef{id: id, path: d, transform: transform})
	return id
}

func clipKey(rect geom.Rect) string {
	q := normalizeRect(rect)
	return fmt.Sprintf("%s,%s,%s,%s",
		formatFloat(q.Min.X),
		formatFloat(q.Min.Y),
		formatFloat(q.Max.X),
		formatFloat(q.Max.Y),
	)
}

func encodeImage(img *image.RGBA) (string, error) {
	if img == nil {
		return "", errors.New("svg: image is nil")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func asRGBAImage(img render.Image) *image.RGBA {
	rgbaImage, ok := img.(interface {
		RGBA() *image.RGBA
	})
	if !ok {
		return nil
	}

	return rgbaImage.RGBA()
}

func colorizeTeXImage(src *image.RGBA, c render.Color) *image.RGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	r := uint8(clamp01(c.R)*255 + 0.5)
	g := uint8(clamp01(c.G)*255 + 0.5)
	b := uint8(clamp01(c.B)*255 + 0.5)
	alphaScale := clamp01(c.A)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, a16 := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			a := uint8(float64(a16>>8)*alphaScale + 0.5)
			dst.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	return dst
}

func buildPathData(p geom.Path) string {
	if len(p.C) == 0 {
		return ""
	}

	var b strings.Builder
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(p.V) {
				return ""
			}
			pt := quantizePt(p.V[vi])
			vi++
			b.WriteString("M ")
			b.WriteString(formatFloat(pt.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(pt.Y))
		case geom.LineTo:
			if vi >= len(p.V) {
				return ""
			}
			pt := quantizePt(p.V[vi])
			vi++
			b.WriteString(" L ")
			b.WriteString(formatFloat(pt.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(pt.Y))
		case geom.QuadTo:
			if vi+1 >= len(p.V) {
				return ""
			}
			ctrl := quantizePt(p.V[vi])
			to := quantizePt(p.V[vi+1])
			vi += 2
			b.WriteString(" Q ")
			b.WriteString(formatFloat(ctrl.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(ctrl.Y))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.Y))
		case geom.CubicTo:
			if vi+2 >= len(p.V) {
				return ""
			}
			c1 := quantizePt(p.V[vi])
			c2 := quantizePt(p.V[vi+1])
			to := quantizePt(p.V[vi+2])
			vi += 3
			b.WriteString(" C ")
			b.WriteString(formatFloat(c1.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(c1.Y))
			b.WriteString(" ")
			b.WriteString(formatFloat(c2.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(c2.Y))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.Y))
		case geom.ClosePath:
			b.WriteString(" Z")
		default:
			return ""
		}
	}

	d := b.String()
	return strings.TrimSpace(d)
}

func dashedArray(dashes []float64) string {
	if len(dashes) < 2 {
		return ""
	}

	var b strings.Builder
	for i := 0; i < len(dashes)-1; i += 2 {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(formatFloat(dashes[i]))
		b.WriteString(",")
		b.WriteString(formatFloat(dashes[i+1]))
	}

	return b.String()
}

func mapLineJoin(v render.LineJoin) string {
	switch v {
	case render.JoinRound:
		return "round"
	case render.JoinBevel:
		return "bevel"
	default:
		return "miter"
	}
}

func mapLineCap(v render.LineCap) string {
	switch v {
	case render.CapButt:
		return "butt"
	case render.CapRound:
		return "round"
	case render.CapSquare:
		return "square"
	default:
		return "butt"
	}
}

func colorToHex(c render.Color) string {
	return fmt.Sprintf("rgb(%d,%d,%d)",
		toByte(c.R),
		toByte(c.G),
		toByte(c.B),
	)
}

func colorToStyle(c render.Color) (string, float64) {
	return colorToHex(c), clamp01(c.A)
}

func normalizeSVGOptions(opts render.SVGOptions) render.SVGOptions {
	if opts.FontPolicy != render.SVGFontPolicyPath {
		opts.FontPolicy = render.SVGFontPolicyNone
	}
	if len(opts.Metadata) > 0 {
		metadata := make(map[string]string, len(opts.Metadata))
		for k, v := range opts.Metadata {
			metadata[k] = v
		}
		opts.Metadata = metadata
	}
	return opts
}

func (r *Renderer) defID(prefix, content string, sequence int) string {
	if r != nil && r.options.HashSalt != "" {
		sum := sha256.Sum256([]byte(r.options.HashSalt + content))
		return prefix + hex.EncodeToString(sum[:])[:10]
	}
	return prefix + strconv.Itoa(sequence)
}

func writeMetadata(b *strings.Builder, opts render.SVGOptions) {
	metadata := normalizeSVGOptions(opts).Metadata
	if epoch := strings.TrimSpace(os.Getenv("SOURCE_DATE_EPOCH")); epoch != "" {
		if _, ok := metadata["Date"]; !ok {
			if ts, err := strconv.ParseInt(epoch, 10, 64); err == nil {
				if metadata == nil {
					metadata = map[string]string{}
				}
				metadata["Date"] = time.Unix(ts, 0).UTC().Format(time.RFC3339)
			}
		}
	}
	if len(metadata) == 0 {
		b.WriteString("  <metadata></metadata>\n")
		return
	}
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	b.WriteString("  <metadata>\n")
	for _, key := range keys {
		b.WriteString("    <meta")
		writeAttr(b, "name", key)
		writeAttr(b, "content", metadata[key])
		b.WriteString(" />\n")
	}
	b.WriteString("  </metadata>\n")
}

func writeColorAttrs(b *strings.Builder, attr string, c render.Color, forced bool) {
	colorValue, alpha := colorToStyle(c)
	writeAttr(b, attr, colorValue)
	if !forced && alpha < 1 {
		writeFloatAttr(b, attr+"-opacity", alpha)
	}
}

func writeStrokeAttrs(b *strings.Builder, paint render.Paint) {
	if paint.Stroke.A <= 0 || paint.LineWidth <= 0 {
		writeAttr(b, "stroke", "none")
		return
	}

	writeColorAttrs(b, "stroke", paint.Stroke, forcedOpacity(paint))
	writeFloatAttr(b, "stroke-width", paint.LineWidth)
	writeAttr(b, "stroke-linejoin", mapLineJoin(paint.LineJoin))
	writeAttr(b, "stroke-linecap", mapLineCap(paint.LineCap))
	if paint.MiterLimit > 0 {
		writeFloatAttr(b, "stroke-miterlimit", paint.MiterLimit)
	}
	if len(paint.Dashes) >= 2 {
		writeAttr(b, "stroke-dasharray", dashedArray(paint.Dashes))
	}
}

func forcedOpacity(paint render.Paint) bool {
	return paint.ForceAlpha && clamp01(clampFloat(paint.Alpha)) < 1
}

func writeForcedOpacity(b *strings.Builder, paint render.Paint) {
	if forcedOpacity(paint) {
		writeFloatAttr(b, "opacity", clamp01(clampFloat(paint.Alpha)))
	}
}

func hatchKey(hatch string, face, line render.Color, width, spacing float64, forced bool) string {
	return strings.Join([]string{
		hatch,
		formatFloat(face.R),
		formatFloat(face.G),
		formatFloat(face.B),
		formatFloat(face.A),
		formatFloat(line.R),
		formatFloat(line.G),
		formatFloat(line.B),
		formatFloat(line.A),
		formatFloat(width),
		formatFloat(spacing),
		strconv.FormatBool(forced),
	}, "\x00")
}

func writeHatchDef(b *strings.Builder, hatch hatchDef) {
	b.WriteString(`    <pattern id="`)
	b.WriteString(hatch.id)
	b.WriteString(`" patternUnits="userSpaceOnUse" width="72" height="72">`)
	if hatch.faceColor.A > 0 {
		b.WriteString(`<rect x="0" y="0" width="72" height="72"`)
		writeColorAttrs(b, "fill", hatch.faceColor, hatch.forced)
		b.WriteString(` />`)
	}
	if hatch.lineColor.A > 0 {
		d := hatchPathData(hatch.hatch, hatch.spacing)
		if d != "" {
			b.WriteString(`<path`)
			writeAttr(b, "d", d)
			writeAttr(b, "fill", "none")
			writeColorAttrs(b, "stroke", hatch.lineColor, hatch.forced)
			writeFloatAttr(b, "stroke-width", hatch.lineWidth)
			writeAttr(b, "stroke-linecap", "butt")
			b.WriteString(` />`)
		}
		writeHatchShapeDefs(b, hatch)
	}
	b.WriteString("</pattern>\n")
}

func writeHatchShapeDefs(b *strings.Builder, hatch hatchDef) {
	for _, shape := range vectorhatch.ShapePaths(hatch.hatch, hatch.spacing) {
		d := buildPathData(shape.Path)
		if d == "" {
			continue
		}
		b.WriteString(`<path`)
		writeAttr(b, "d", d)
		if shape.Filled {
			writeColorAttrs(b, "fill", hatch.lineColor, hatch.forced)
			writeAttr(b, "stroke", "none")
		} else {
			writeAttr(b, "fill", "none")
			writeColorAttrs(b, "stroke", hatch.lineColor, hatch.forced)
			writeFloatAttr(b, "stroke-width", hatch.lineWidth)
			writeAttr(b, "stroke-linecap", "butt")
		}
		b.WriteString(` />`)
	}
}

func filterKey(name string, radius float64) string {
	return strings.Join([]string{name, formatFloat(radius)}, "\x00")
}

func writeFilterDef(b *strings.Builder, filter filterDef) {
	b.WriteString(`    <filter id="`)
	b.WriteString(filter.id)
	b.WriteString(`" x="-20%" y="-20%" width="140%" height="140%">`)
	switch filter.name {
	case "blur":
		b.WriteString(`<feGaussianBlur`)
		writeFloatAttr(b, "stdDeviation", filter.radius)
		b.WriteString(` />`)
	default:
		b.WriteString(`<feComposite operator="over" />`)
	}
	b.WriteString("</filter>\n")
}

func hatchPathData(hatch string, spacing float64) string {
	if spacing <= 0 {
		spacing = 8
	}
	var b strings.Builder
	writeHatchLines := func(count int, draw func(float64)) {
		if count <= 0 {
			return
		}
		step := math.Max(2, spacing/float64(count))
		for v := -72.0; v <= 144; v += step {
			draw(v)
		}
	}
	line := func(x1, y1, x2, y2 float64) {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("M ")
		b.WriteString(formatFloat(x1))
		b.WriteByte(' ')
		b.WriteString(formatFloat(y1))
		b.WriteString(" L ")
		b.WriteString(formatFloat(x2))
		b.WriteByte(' ')
		b.WriteString(formatFloat(y2))
	}

	verticalCount := strings.Count(hatch, "|") + strings.Count(hatch, "+")
	horizontalCount := strings.Count(hatch, "-") + strings.Count(hatch, "+")
	slashCount := strings.Count(hatch, "/") + strings.Count(hatch, "x") + strings.Count(hatch, "X")
	backslashCount := strings.Count(hatch, `\`) + strings.Count(hatch, "x") + strings.Count(hatch, "X")

	writeHatchLines(verticalCount, func(x float64) { line(x, 0, x, 72) })
	writeHatchLines(horizontalCount, func(y float64) { line(0, y, 72, y) })
	writeHatchLines(slashCount, func(x float64) { line(x, 72, x+72, 0) })
	writeHatchLines(backslashCount, func(x float64) { line(x, 0, x+72, 72) })
	return b.String()
}

func toByte(v float64) uint8 {
	v = clamp01(v)
	return uint8(v*255 + 0.5)
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

func quantize(v float64) float64 {
	return math.Round(v/quantizationGrid) * quantizationGrid
}

func quantizePt(p geom.Pt) geom.Pt {
	return geom.Pt{X: quantize(p.X), Y: quantize(p.Y)}
}

func normalizeRect(rect geom.Rect) geom.Rect {
	minX := rect.Min.X
	minY := rect.Min.Y
	maxX := rect.Max.X
	maxY := rect.Max.Y

	if maxX < minX {
		minX, maxX = maxX, minX
	}
	if maxY < minY {
		minY, maxY = maxY, minY
	}

	return geom.Rect{
		Min: geom.Pt{X: quantize(minX), Y: quantize(minY)},
		Max: geom.Pt{X: quantize(maxX), Y: quantize(maxY)},
	}
}

func writeAttr(b *strings.Builder, name, value string) {
	b.WriteString(" ")
	b.WriteString(name)
	b.WriteByte('=')
	b.WriteString(strconv.Quote(value))
}

func writeFloatAttr(b *strings.Builder, name string, value float64) {
	b.WriteString(" ")
	b.WriteString(name)
	b.WriteString("=\"")
	b.WriteString(formatFloat(value))
	b.WriteString("\"")
}

func formatFloat(v float64) string {
	return shortFloat(v)
}

// shortFloat formats v with up to 6 decimal digits, mirroring matplotlib's
// _short_float_fmt: trailing zeros (and a trailing decimal point) are stripped,
// negative zero is normalized to "0", and NaN/Inf are clamped to "0". The
// output stays in fixed (non-exponent) notation so SVG number attributes remain
// portable across viewers.
func shortFloat(v float64) string {
	s := strconv.FormatFloat(clampFloat(v), 'f', 6, 64)
	if i := strings.IndexByte(s, '.'); i >= 0 {
		end := len(s)
		for end > i && s[end-1] == '0' {
			end--
		}
		if end > 0 && s[end-1] == '.' {
			end--
		}
		s = s[:end]
	}
	if s == "-0" || s == "" {
		s = "0"
	}
	return s
}

// rotateTransform returns a canonical SVG matrix transform for rotation around
// (cx, cy), using the same compact float formatting as other transforms.
func rotateTransform(angleDeg, cx, cy float64) string {
	return matrixTransform(rotationAffine(angleDeg, cx, cy))
}

func rotationAffine(angleDeg, cx, cy float64) geom.Affine {
	rad := angleDeg * math.Pi / 180
	cos := math.Cos(rad)
	sin := math.Sin(rad)
	return geom.Affine{
		A: cos,
		B: sin,
		C: -sin,
		D: cos,
		E: cx - cos*cx + sin*cy,
		F: cy - sin*cx - cos*cy,
	}
}

func affinePath(path geom.Path, affine geom.Affine) geom.Path {
	if len(path.V) == 0 {
		return path
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, pt := range path.V {
		out.V[i] = affine.Apply(pt)
	}
	return out
}

func clampFloat(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}

func fontFamily(key string) string {
	return render.CSSFontFamily(key)
}

func (r *Renderer) svgFontFamily(key string) string {
	if family := r.registerFontFace(key); family != "" {
		return family
	}
	return fontFamily(key)
}

func (r *Renderer) registerFontFace(key string) string {
	path := strings.TrimSpace(key)
	if path == "" || !isFontFile(path) {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return ""
	}
	if r.fontFaces == nil {
		r.fontFaces = map[string]fontFaceDef{}
	}
	if face, ok := r.fontFaces[path]; ok {
		return face.family
	}
	face := fontFaceDef{
		family: "mplgo-font-" + strconv.Itoa(len(r.fontFaces)+1),
		data:   base64.StdEncoding.EncodeToString(data),
		mime:   fontMIME(path),
		format: fontFormat(path),
	}
	r.fontFaces[path] = face
	r.fontFaceOrder = append(r.fontFaceOrder, face)
	return face.family
}

func isFontFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ttf", ".otf", ".ttc", ".dfont":
		return true
	default:
		return false
	}
}

func fontMIME(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".otf":
		return "font/otf"
	default:
		return "font/ttf"
	}
}

func fontFormat(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".otf":
		return "opentype"
	default:
		return "truetype"
	}
}

func escapeText(text string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(text))
	return b.String()
}

package pdf

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends/internal/mixedraster"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/font/sfnt"
)

const (
	// defaultFontHeight matches the SVG backend default. It is the fallback
	// ascent the MeasureText path returns when no text drawer is wired up.
	defaultFontHeight = 13.0
)

// state captures the parts of renderer state that change with Save/Restore.
type state struct {
	// inContent reports whether the saved state lies inside a content stream.
	// Save before any draw call is permitted and behaves like the initial
	// identity transform.
	inContent bool
	clipRect  *geom.Rect
	clipPaths []geom.Path
}

type pdfImage struct {
	name     string
	width    int
	height   int
	colors   int
	rgb      []byte
	alpha    []byte
	hasAlpha bool
	filter   string
}

type pdfImageObject struct {
	pdfImage
	objectID int
	smaskID  int
}

type pdfHatchPattern struct {
	name      string
	hatch     string
	faceColor render.Color
	lineColor render.Color
	lineWidth float64
	spacing   float64
}

type pdfHatchPatternObject struct {
	pdfHatchPattern
	objectID int
}

type pdfFillPattern struct {
	name    string
	pattern render.PatternFill
}

type pdfFillPatternObject struct {
	pdfFillPattern
	objectID int
}

type pdfShading struct {
	name     string
	gradient render.GradientFill
}

type pdfShadingObject struct {
	pdfShading
	objectID int
}

type pdfFormXObject struct {
	name              string
	path              geom.Path
	paintOp           string
	bbox              geom.Rect
	lineJoin          render.LineJoin
	lineCap           render.LineCap
	content           []byte
	hasContent        bool
	transparencyGroup bool
}

type pdfFormXObjectObject struct {
	pdfFormXObject
	objectID int
}

type pdfAlphaState struct {
	name        string
	strokeAlpha float64
	fillAlpha   float64
}

type pdfEmbeddedFont struct {
	name      string
	face      render.FontFace
	data      []byte
	baseName  string
	cidByGID  map[sfnt.GlyphIndex]uint16
	gidByCID  map[uint16]sfnt.GlyphIndex
	runeByCID map[uint16]rune
}

type pdfEmbeddedFontObject struct {
	pdfEmbeddedFont
	type0ID      int
	cidFontID    int
	descriptorID int
	fontFileID   int
	cidToGIDID   int
	widthsID     int
	toUnicodeID  int
}

// Renderer implements render.Renderer by emitting a PDF document.
//
// The renderer buffers a single content stream. Calling End() finalizes the
// document into memory; SavePDF then flushes the in-memory PDF bytes to disk.
// The buffer is reusable: callers can Begin/End again to overwrite the
// previous document.
type Renderer struct {
	width      int
	height     int
	viewport   geom.Rect
	background render.Color
	resolution uint

	began     bool
	stack     []state
	clipRect  *geom.Rect
	clipPaths []geom.Path

	// content is the page content stream under construction.
	content        bytes.Buffer
	images         []pdfImage
	imageIDs       map[string]string
	hatchPatterns  []pdfHatchPattern
	hatchIDs       map[string]string
	fillPatterns   []pdfFillPattern
	fillPatternIDs map[string]string
	shadings       []pdfShading
	shadingIDs     map[string]string
	forms          []pdfFormXObject
	formIDs        map[string]string
	alphaStates    []pdfAlphaState
	alphaIDs       map[string]string
	fonts          []pdfEmbeddedFont
	fontIDs        map[string]string
	// document is the fully serialized PDF bytes ready for write.
	document []byte

	// pdfOpts carries setter-supplied options. SavePDFWithOptions overrides
	// fields directly for that single call.
	pdfOpts render.PDFOptions

	lastFontKey string
	texManager  *tex.Manager
	texErr      error
	raster      *mixedraster.Session
}

// Compile-time interface assertions.
var (
	_ render.Renderer                = (*Renderer)(nil)
	_ render.PNGExporter             = nil // explicitly not implemented
	_ render.PDFExporter             = (*Renderer)(nil)
	_ render.DPIAware                = (*Renderer)(nil)
	_ render.ImageTransformer        = (*Renderer)(nil)
	_ render.TextPather              = (*Renderer)(nil)
	_ render.FontTextDrawer          = (*Renderer)(nil)
	_ render.FontRotatedTextDrawer   = (*Renderer)(nil)
	_ render.FontVerticalTextDrawer  = (*Renderer)(nil)
	_ render.TeXMetricer             = (*Renderer)(nil)
	_ render.TeXDrawer               = (*Renderer)(nil)
	_ render.RotatedTeXDrawer        = (*Renderer)(nil)
	_ render.NativeHatcher           = (*Renderer)(nil)
	_ render.PatternFiller           = (*Renderer)(nil)
	_ render.GradientFiller          = (*Renderer)(nil)
	_ render.PathEffectFilterDrawer  = (*Renderer)(nil)
	_ render.MarkerDrawer            = (*Renderer)(nil)
	_ render.PathCollectionDrawer    = (*Renderer)(nil)
	_ render.RasterizationController = (*Renderer)(nil)
	_ render.PDFOptionExporter       = (*Renderer)(nil)
	_ render.PDFOptionSetter         = (*Renderer)(nil)
)

// New constructs a PDF renderer that produces a single-page document of the
// given width and height in points (1 point = 1/72 inch).
func New(width, height int, background render.Color) (*Renderer, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("pdf: invalid size %dx%d", width, height)
	}
	r := &Renderer{
		width:      width,
		height:     height,
		background: background,
		resolution: 72,
		pdfOpts:    render.DefaultPDFOptions(),
		texManager: tex.NewManager(tex.ManagerConfig{}),
	}
	return r, nil
}

// SetResolution implements render.DPIAware. PDF coordinates are always in
// points, so the renderer keeps DPI as informational state for callers that
// want to scale rasterized output later.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi == 0 {
		dpi = 72
	}
	r.resolution = dpi
}

// SetPDFOptions implements render.PDFOptionSetter.
func (r *Renderer) SetPDFOptions(opts render.PDFOptions) {
	r.pdfOpts = opts
}

// Begin starts a drawing session for the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("pdf: Begin called twice")
	}
	r.began = true
	r.viewport = viewport
	r.content.Reset()
	r.document = nil
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipPaths = nil
	r.raster = nil
	r.images = r.images[:0]
	r.imageIDs = map[string]string{}
	r.hatchPatterns = r.hatchPatterns[:0]
	r.hatchIDs = map[string]string{}
	r.fillPatterns = r.fillPatterns[:0]
	r.fillPatternIDs = map[string]string{}
	r.shadings = r.shadings[:0]
	r.shadingIDs = map[string]string{}
	r.forms = r.forms[:0]
	r.formIDs = map[string]string{}
	r.alphaStates = r.alphaStates[:0]
	r.alphaIDs = map[string]string{}
	r.fonts = r.fonts[:0]
	r.fontIDs = map[string]string{}
	r.lastFontKey = ""

	// PDF's coordinate origin is bottom-left with +Y up, which matches the
	// matplotlib-go y-up display space exactly (mirroring Matplotlib's PDF
	// backend, whose flipy() is False). No device flip is emitted; draws use
	// display coordinates directly.

	if r.background.A > 0 {
		// Paint the page background as a filled rectangle covering the page.
		writeFillColor(&r.content, r.background)
		fmt.Fprintf(
			&r.content, "0 0 %s %s re f\n",
			shortFloat(float64(r.width)),
			shortFloat(float64(r.height)),
		)
	}
	return nil
}

// End finalizes the current drawing session.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("pdf: End called before Begin")
	}
	r.began = false
	doc, err := buildDocument(r.width, r.height, r.content.Bytes(), r.images, r.hatchPatterns, r.fillPatterns, r.shadings, r.forms, r.alphaStates, r.fonts, r.pdfOpts)
	if err != nil {
		return err
	}
	r.document = doc
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

// StopRasterized embeds the active raster group as a PDF image XObject.
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

// Save pushes graphics state.
func (r *Renderer) Save() {
	if rr := r.activeRaster(); rr != nil {
		rr.Save()
		return
	}
	r.stack = append(r.stack, state{
		inContent: r.began,
		clipRect:  cloneRectPtr(r.clipRect),
		clipPaths: mixedraster.ClonePaths(r.clipPaths),
	})
	if r.began {
		r.content.WriteString("q\n")
	}
}

// Restore pops graphics state.
func (r *Renderer) Restore() {
	if rr := r.activeRaster(); rr != nil {
		rr.Restore()
		return
	}
	if len(r.stack) == 0 {
		return
	}
	top := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	r.clipRect = top.clipRect
	r.clipPaths = mixedraster.ClonePaths(top.clipPaths)
	if top.inContent && r.began {
		r.content.WriteString("Q\n")
	}
}

// ClipRect installs a rectangular clip.
func (r *Renderer) ClipRect(rect geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipRect(rect)
		return
	}
	if !r.began {
		return
	}
	normalized := normalizeRect(rect)
	if r.clipRect == nil {
		r.clipRect = &normalized
	} else {
		intersected := r.clipRect.Intersect(normalized)
		r.clipRect = &intersected
	}
	fmt.Fprintf(
		&r.content, "%s %s %s %s re W n\n",
		shortFloat(rect.Min.X),
		shortFloat(rect.Min.Y),
		shortFloat(rect.W()),
		shortFloat(rect.H()),
	)
}

// ClipPath installs an arbitrary path clip.
func (r *Renderer) ClipPath(p geom.Path) {
	if rr := r.activeRaster(); rr != nil {
		rr.ClipPath(p)
		return
	}
	if !r.began {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}
	r.clipPaths = append(r.clipPaths, mixedraster.ClonePath(p))
	r.content.WriteString("W n\n")
}

// SupportsNativeHatch reports that the PDF backend emits hatch fills as
// native PDF tiling pattern resources.
func (r *Renderer) SupportsNativeHatch() bool { return true }

// SupportsPatternFill reports that the PDF backend renders Paint.FillPattern
// natively through colored tiling pattern resources.
func (r *Renderer) SupportsPatternFill() bool { return true }

// SupportsGradientFill reports that the PDF backend renders Paint.FillGradient
// natively through axial and radial shading resources.
func (r *Renderer) SupportsGradientFill() bool { return true }

// Path draws a path using the provided paint.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if rr := r.activeRaster(); rr != nil {
		rr.Path(p, paint)
		return
	}
	if !r.began || paint == nil {
		return
	}
	if render.DrawPathWithEffects(r, p, paint, r.Path) {
		return
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasGradient := !hasHatch && paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0
	hasPattern := !hasHatch && !hasGradient && (paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0)
	hasFill := paint.Fill.A > 0 || hasHatch || hasGradient || hasPattern
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasStroke {
		return
	}
	if hasGradient {
		r.writeGradientFill(p, paint)
		if hasStroke && writePathOps(&r.content, p) {
			r.writeAlphaState(paint)
			writeStrokeColor(&r.content, paint.Stroke)
			writeLineState(&r.content, paint)
			r.content.WriteString("S\n")
		}
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}

	if hasFill {
		r.writeAlphaState(paint)
		if hasHatch {
			writePatternFill(&r.content, r.registerHatchPattern(*paint))
		} else if hasPattern {
			writePatternFill(&r.content, r.registerFillPattern(paint.FillPattern))
		} else {
			writeFillColor(&r.content, paint.Fill)
		}
	}
	if hasStroke {
		if !hasFill {
			r.writeAlphaState(paint)
		}
		writeStrokeColor(&r.content, paint.Stroke)
		writeLineState(&r.content, paint)
	}

	switch {
	case hasFill && hasStroke:
		r.content.WriteString("B\n")
	case hasFill:
		r.content.WriteString("f\n")
	case hasStroke:
		r.content.WriteString("S\n")
	}
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

// SupportsPathEffectFilter reports whether PDF can render the filter effect
// through backend-local PDF resources instead of core mixed-raster fallback.
func (r *Renderer) SupportsPathEffectFilter(effect render.PathEffect) bool {
	return isPDFIdentityPathEffectFilter(effect) || isPDFBlurPathEffectFilter(effect)
}

// DrawPathEffectFilter captures supported filter passes into PDF-native
// resources: identity filters use transparency-group Form XObjects, while blur
// filters repaint the pass to an isolated soft-mask image XObject.
func (r *Renderer) DrawPathEffectFilter(path geom.Path, paint render.Paint, effect render.PathEffect, _ func(geom.Path, *render.Paint)) bool {
	if r == nil || !r.began || !r.SupportsPathEffectFilter(effect) {
		return false
	}
	if isPDFBlurPathEffectFilter(effect) {
		return r.drawBlurredPathEffectFilter(path, paint, effect)
	}
	name, ok := r.registerPathEffectForm(path, &paint)
	if !ok {
		return false
	}
	fmt.Fprintf(&r.content, "q\n/%s Do\nQ\n", escapeName(name))
	return true
}

func (r *Renderer) drawBlurredPathEffectFilter(path geom.Path, paint render.Paint, effect render.PathEffect) bool {
	dpi := float64(r.resolution)
	if dpi <= 0 {
		dpi = 72
	}
	session, ok := mixedraster.Start(
		r.width,
		r.height,
		r.viewport,
		render.Rasterization{Mode: render.RasterizeAuto, DPI: dpi},
		r.resolution,
		r.clipRect,
		r.clipPaths,
	)
	if !ok {
		return false
	}
	session.Renderer().Path(path, &paint)
	img, rect, ok := session.Stop()
	if !ok || img == nil || img.RGBA() == nil {
		return false
	}
	filterEffect := effect
	filterEffect.FilterRadius *= dpi / 72.0
	filtered, _ := render.ApplyPathEffectFilter(img.RGBA(), filterEffect, dpi)
	if filtered == nil {
		return false
	}
	r.Image(render.NewImageData(filtered), rect)
	return true
}

func (r *Renderer) writeAlphaState(paint *render.Paint) {
	if paint == nil {
		return
	}
	strokeAlpha := 1.0
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		strokeAlpha = clamp01(paint.Stroke.A)
	}
	fillAlpha := 1.0
	if paint.Fill.A > 0 || (paint.Hatch != "" && paint.HatchColor.A > 0) || (paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0) || (paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0) {
		fillAlpha = clamp01(paint.Fill.A)
		if paint.Hatch != "" && paint.HatchColor.A > 0 {
			fillAlpha = clamp01(paint.HatchColor.A)
		} else if paint.FillGradient.Kind != render.GradientNone && len(paint.FillGradient.Stops) > 0 {
			fillAlpha = gradientAlpha(paint.FillGradient.Stops)
		} else if paint.FillPattern.ID != "" || len(paint.FillPattern.Path.V) > 0 {
			fillAlpha = patternAlpha(paint.FillPattern)
		}
	}
	if strokeAlpha >= 1 && fillAlpha >= 1 {
		return
	}
	name := r.registerAlphaState(strokeAlpha, fillAlpha)
	fmt.Fprintf(&r.content, "/%s gs\n", name)
}

func (r *Renderer) registerAlphaState(strokeAlpha, fillAlpha float64) string {
	if r.alphaIDs == nil {
		r.alphaIDs = map[string]string{}
	}
	strokeAlpha = clamp01(strokeAlpha)
	fillAlpha = clamp01(fillAlpha)
	key := shortFloat(strokeAlpha) + "\x00" + shortFloat(fillAlpha)
	if id, ok := r.alphaIDs[key]; ok {
		return id
	}
	id := fmt.Sprintf("A%d", len(r.alphaStates)+1)
	r.alphaIDs[key] = id
	r.alphaStates = append(r.alphaStates, pdfAlphaState{
		name:        id,
		strokeAlpha: strokeAlpha,
		fillAlpha:   fillAlpha,
	})
	return id
}

func (r *Renderer) registerHatchPattern(paint render.Paint) string {
	if r.hatchIDs == nil {
		r.hatchIDs = map[string]string{}
	}
	lineWidth := paint.HatchLineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	spacing := paint.HatchSpacing
	if spacing <= 0 {
		spacing = 8
	}
	key := hatchPatternKey(paint.Hatch, paint.Fill, paint.HatchColor, lineWidth, spacing)
	if id, ok := r.hatchIDs[key]; ok {
		return id
	}
	id := fmt.Sprintf("Pa%d", len(r.hatchPatterns)+1)
	r.hatchIDs[key] = id
	r.hatchPatterns = append(r.hatchPatterns, pdfHatchPattern{
		name:      id,
		hatch:     paint.Hatch,
		faceColor: paint.Fill,
		lineColor: paint.HatchColor,
		lineWidth: lineWidth,
		spacing:   spacing,
	})
	return id
}

func (r *Renderer) registerFillPattern(pattern render.PatternFill) string {
	if r.fillPatternIDs == nil {
		r.fillPatternIDs = map[string]string{}
	}
	pattern.Path = clonePath(pattern.Path)
	key := fillPatternKey(pattern)
	if id, ok := r.fillPatternIDs[key]; ok {
		return id
	}
	id := fmt.Sprintf("Pf%d", len(r.fillPatterns)+1)
	r.fillPatternIDs[key] = id
	r.fillPatterns = append(r.fillPatterns, pdfFillPattern{
		name:    id,
		pattern: pattern,
	})
	return id
}

func (r *Renderer) writeGradientFill(p geom.Path, paint *render.Paint) {
	if paint == nil || paint.FillGradient.Kind == render.GradientNone || len(paint.FillGradient.Stops) == 0 {
		return
	}
	name := r.registerShading(paint.FillGradient)
	r.content.WriteString("q\n")
	if !writePathOps(&r.content, p) {
		r.content.WriteString("Q\n")
		return
	}
	r.content.WriteString("W\nn\n")
	r.writeAlphaState(paint)
	fmt.Fprintf(&r.content, "/%s sh\nQ\n", escapeName(name))
}

func (r *Renderer) registerShading(gradient render.GradientFill) string {
	if r.shadingIDs == nil {
		r.shadingIDs = map[string]string{}
	}
	gradient.Stops = normalizeGradientStops(gradient.Stops)
	key := shadingKey(gradient)
	if id, ok := r.shadingIDs[key]; ok {
		return id
	}
	id := fmt.Sprintf("Sh%d", len(r.shadings)+1)
	r.shadingIDs[key] = id
	r.shadings = append(r.shadings, pdfShading{
		name:     id,
		gradient: gradient,
	})
	return id
}

// DrawMarkers renders one marker path at many display-space offsets using a
// reusable Form XObject for the marker geometry.
func (r *Renderer) DrawMarkers(batch render.MarkerBatch) bool {
	if rr := r.activeRaster(); rr != nil {
		if markers, ok := rr.(render.MarkerDrawer); ok {
			return markers.DrawMarkers(batch)
		}
		return false
	}
	if !r.began || len(batch.Marker.C) == 0 || len(batch.Items) == 0 || !batch.Marker.Validate() {
		return false
	}
	emitted := false
	for i := range batch.Items {
		item := batch.Items[i]
		marker := affinePath(batch.Marker, item.Transform)
		if !marker.Validate() || len(marker.C) == 0 {
			continue
		}
		paint := item.Paint
		paintOp := paintOperator(&paint)
		if paintOp == "" {
			continue
		}
		name := r.registerFormXObject("M", marker, paintOp, &paint)
		r.writePaintState(&paint)
		fmt.Fprintf(
			&r.content, "q\n1 0 0 1 %s %s cm\n/%s Do\nQ\n",
			shortFloat(item.Offset.X),
			shortFloat(item.Offset.Y),
			name,
		)
		emitted = true
	}
	return emitted
}

// DrawPathCollection renders display-space paths through Form XObject
// templates with per-item paint state applied at invocation time.
func (r *Renderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	if rr := r.activeRaster(); rr != nil {
		if paths, ok := rr.(render.PathCollectionDrawer); ok {
			return paths.DrawPathCollection(batch)
		}
		return false
	}
	if !r.began || len(batch.Items) == 0 {
		return false
	}
	emitted := false
	for i := range batch.Items {
		item := batch.Items[i]
		if !item.Path.Validate() || len(item.Path.C) == 0 {
			continue
		}
		paint := item.Paint
		if item.Hatch != "" {
			paint.Hatch = item.Hatch
			paint.HatchColor = item.HatchColor
			paint.HatchLineWidth = item.HatchWidth
			paint.HatchSpacing = item.HatchSpacing
		}
		paintOp := paintOperator(&paint)
		if paintOp == "" {
			continue
		}
		name := r.registerFormXObject("P", item.Path, paintOp, &paint)
		r.writePaintState(&paint)
		fmt.Fprintf(&r.content, "/%s Do\n", name)
		emitted = true
	}
	return emitted
}

func (r *Renderer) registerFormXObject(prefix string, path geom.Path, paintOp string, paint *render.Paint) string {
	if r.formIDs == nil {
		r.formIDs = map[string]string{}
	}
	key := formXObjectKey(prefix, path, paintOp, paint)
	if id, ok := r.formIDs[key]; ok {
		return id
	}
	name := fmt.Sprintf("%s%d", prefix, len(r.forms)+1)
	bbox, ok := pathBounds(path)
	if !ok {
		bbox = geom.Rect{}
	}
	padding := formPadding(paint)
	bbox = bbox.Inflate(padding, padding)
	r.formIDs[key] = name
	r.forms = append(r.forms, pdfFormXObject{
		name:     name,
		path:     clonePath(path),
		paintOp:  paintOp,
		bbox:     bbox,
		lineJoin: paint.LineJoin,
		lineCap:  paint.LineCap,
	})
	return name
}

func (r *Renderer) registerPathEffectForm(path geom.Path, paint *render.Paint) (string, bool) {
	if paint == nil || !path.Validate() || len(path.C) == 0 {
		return "", false
	}
	if r.formIDs == nil {
		r.formIDs = map[string]string{}
	}
	content, ok := r.capturePathEffectFormContent(path, paint)
	if !ok || len(content) == 0 {
		return "", false
	}
	bbox, ok := pathBounds(path)
	if !ok {
		return "", false
	}
	padding := formPadding(paint)
	bbox = bbox.Inflate(padding, padding)
	key := pathEffectFormKey(content, bbox)
	if id, ok := r.formIDs[key]; ok {
		return id, true
	}
	name := fmt.Sprintf("E%d", len(r.forms)+1)
	r.formIDs[key] = name
	r.forms = append(r.forms, pdfFormXObject{
		name:              name,
		bbox:              bbox,
		content:           append([]byte(nil), content...),
		hasContent:        true,
		transparencyGroup: true,
	})
	return name, true
}

func (r *Renderer) capturePathEffectFormContent(path geom.Path, paint *render.Paint) ([]byte, bool) {
	outer := r.content
	r.content = bytes.Buffer{}
	r.Path(path, paint)
	content := append([]byte(nil), r.content.Bytes()...)
	r.content = outer
	return content, len(content) > 0
}

// Image draws a raster image into the destination rectangle as a PDF image
// XObject. RGBA images with alpha get a grayscale soft mask.
func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.Image(img, dst)
		return
	}
	if !r.began || img == nil || dst.W() <= 0 || dst.H() <= 0 {
		return
	}
	matrix := geom.Affine{A: dst.W(), D: dst.H(), E: dst.Min.X, F: dst.Min.Y}
	r.drawImageWithMatrix(img, matrix)
}

// ImageTransformed draws a raster image through an arbitrary affine transform.
// The affine maps source image pixels into display coordinates; PDF image
// XObjects paint a unit square, so the current transformation matrix includes
// the source image dimensions.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, transform geom.Affine) {
	if rr := r.activeRaster(); rr != nil {
		if tr, ok := rr.(render.ImageTransformer); ok {
			tr.ImageTransformed(img, geom.Rect{}, transform)
		}
		return
	}
	if !r.began || img == nil {
		return
	}
	w, h := img.Size()
	if w <= 0 || h <= 0 {
		return
	}
	matrix := geom.Affine{
		A: transform.A * float64(w),
		B: transform.B * float64(w),
		C: transform.C * float64(h),
		D: transform.D * float64(h),
		E: transform.E,
		F: transform.F,
	}
	r.drawImageWithMatrix(img, matrix)
}

func (r *Renderer) drawImageWithMatrix(img render.Image, matrix geom.Affine) {
	if jpegSource, ok := img.(render.JPEGImage); ok {
		pdfImg, ok := encodePDFJPEGImage("", jpegSource)
		if !ok {
			return
		}
		name := r.registerImage(pdfImg)
		r.writeImageInvocation(matrix, name)
		return
	}
	rgbaSource, ok := img.(render.RGBAImage)
	if !ok || rgbaSource.RGBA() == nil {
		return
	}
	pdfImg, ok := encodePDFImage(
		"",
		rgbaSource.RGBA(),
		imageAlphaMultiplier(img),
	)
	if !ok {
		return
	}
	name := r.registerImage(pdfImg)
	r.writeImageInvocation(matrix, name)
}

func (r *Renderer) writeImageInvocation(matrix geom.Affine, name string) {
	fmt.Fprintf(
		&r.content, "q\n%s %s %s %s %s %s cm\n/%s Do\nQ\n",
		shortFloat(matrix.A),
		shortFloat(matrix.B),
		shortFloat(matrix.C),
		shortFloat(matrix.D),
		shortFloat(matrix.E),
		shortFloat(matrix.F),
		name,
	)
}

func (r *Renderer) registerImage(img pdfImage) string {
	if r.imageIDs == nil {
		r.imageIDs = map[string]string{}
	}
	key := imageKey(img)
	if name, ok := r.imageIDs[key]; ok {
		return name
	}
	img.name = fmt.Sprintf("Im%d", len(r.images)+1)
	r.imageIDs[key] = img.name
	r.images = append(r.images, img)
	return img.name
}

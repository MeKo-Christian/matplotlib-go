package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

const (
	autoRasterizeScatterPointThreshold = 1000
	autoRasterizeContourPathThreshold  = 200
)

// ArtistRasterization stores artist-level mixed raster/vector output intent.
//
// It is embedded by concrete artists that expose Matplotlib-style
// SetRasterized behavior without making rasterization mandatory on every
// custom Artist implementation.
type ArtistRasterization struct {
	rasterization render.Rasterization
	hidden        bool
	animated      bool
	alpha         float64
	alphaSet      bool
	inLayout      bool
	inLayoutSet   bool
	clipOn        bool
	clipOnSet     bool
	clipRect      geom.Rect
	hasClipRect   bool
	clipPath      geom.Path
	hasClipPath   bool
	clipPathSpec  CoordinateSpec
	hasClipSpec   bool
	clipPathTrans geom.Affine
	hasClipTrans  bool
	transform     transform.T
	hasTransform  bool
	transformSpec CoordinateSpec
	hasTransSpec  bool
	stale         bool
}

// ArtistClip stores explicit clipping metadata for an artist. ClipOn defaults
// to true on zero-value ArtistRasterization, matching Matplotlib's Artist
// metadata defaults.
type ArtistClip struct {
	ClipOn            bool
	ClipRect          geom.Rect
	HasClipRect       bool
	ClipPath          geom.Path
	HasClipPath       bool
	ClipPathCoords    CoordinateSpec
	HasClipPathCoords bool
	ClipPathTransform geom.Affine
	HasClipPathTrans  bool
}

// ArtistTransform stores explicit transform metadata for path-like artists.
// Transform, when set, maps the artist's coordinates directly into display
// space. Otherwise Coords selects one of the DrawContext coordinate transforms.
type ArtistTransform struct {
	Transform    transform.T
	HasTransform bool
	Coords       CoordinateSpec
	HasCoords    bool
}

// SetVisible toggles whether the artist participates in drawing. The zero
// value is visible, matching Matplotlib's default Artist state.
func (a *ArtistRasterization) SetVisible(visible bool) {
	if a == nil {
		return
	}
	a.hidden = !visible
	a.stale = true
}

// Visible reports whether the artist should be drawn.
func (a *ArtistRasterization) Visible() bool {
	return a == nil || !a.hidden
}

// SetAnimated marks the artist as animated. Animated artists are skipped by
// the default figure draw path and drawn only when the figure is rendered
// with DrawOptions.IncludeAnimated true (typically by the animation engine
// during a blit-style overlay pass). The zero value is non-animated, matching
// Matplotlib's default Artist state.
func (a *ArtistRasterization) SetAnimated(animated bool) {
	if a == nil {
		return
	}
	a.animated = animated
	a.stale = true
}

// Animated reports whether the artist is in animated mode.
func (a *ArtistRasterization) Animated() bool {
	return a != nil && a.animated
}

// SetAlpha stores an artist-level alpha multiplier in the inclusive range
// [0,1]. Concrete artists combine this with their own color/alpha fields.
func (a *ArtistRasterization) SetAlpha(alpha float64) {
	if a == nil {
		return
	}
	a.alpha = clampOneToOne(alpha)
	a.alphaSet = true
	a.stale = true
}

// ClearAlpha removes the artist-level alpha override.
func (a *ArtistRasterization) ClearAlpha() {
	if a == nil {
		return
	}
	a.alpha = 0
	a.alphaSet = false
	a.stale = true
}

// ArtistAlpha reports the artist-level alpha multiplier and whether it is set.
func (a *ArtistRasterization) ArtistAlpha() (float64, bool) {
	if a == nil || !a.alphaSet {
		return 1, false
	}
	return a.alpha, true
}

// EffectiveAlpha combines a local Matplotlib-style alpha value with the
// artist-level alpha multiplier. A non-positive local alpha means "unset" for
// existing Go artists and therefore behaves as 1 before the multiplier.
func (a *ArtistRasterization) EffectiveAlpha(local float64) float64 {
	alpha := 1.0
	if local > 0 && local <= 1 {
		alpha = local
	}
	if a != nil && a.alphaSet {
		alpha *= a.alpha
	}
	return clampOneToOne(alpha)
}

// ApplyArtistAlpha multiplies a color's existing alpha by the artist-level
// alpha override when one is set.
func (a *ArtistRasterization) ApplyArtistAlpha(color render.Color) render.Color {
	if a != nil && a.alphaSet {
		color.A *= a.alpha
	}
	color.A = clampOneToOne(color.A)
	return color
}

// SetInLayout records whether layout engines should account for this artist.
func (a *ArtistRasterization) SetInLayout(inLayout bool) {
	if a == nil {
		return
	}
	a.inLayout = inLayout
	a.inLayoutSet = true
	a.stale = true
}

// InLayout reports whether layout engines should account for this artist. The
// zero value is true, matching Matplotlib's default Artist.in_layout behavior.
func (a *ArtistRasterization) InLayout() bool {
	return a == nil || !a.inLayoutSet || a.inLayout
}

// SetClipOn toggles whether explicit artist-level clipping is applied. The
// zero value is true.
func (a *ArtistRasterization) SetClipOn(clipOn bool) {
	if a == nil {
		return
	}
	a.clipOn = clipOn
	a.clipOnSet = true
	a.stale = true
}

// ClipOn reports whether explicit artist-level clipping is enabled.
func (a *ArtistRasterization) ClipOn() bool {
	return a == nil || !a.clipOnSet || a.clipOn
}

// SetClipRect stores an explicit rectangular clip in display coordinates.
func (a *ArtistRasterization) SetClipRect(rect geom.Rect) {
	if a == nil {
		return
	}
	a.clipRect = rect
	a.hasClipRect = true
	a.stale = true
}

// ClearClipRect removes the explicit rectangular clip.
func (a *ArtistRasterization) ClearClipRect() {
	if a == nil {
		return
	}
	a.clipRect = geom.Rect{}
	a.hasClipRect = false
	a.stale = true
}

// SetClipPath stores an explicit path clip in display coordinates.
func (a *ArtistRasterization) SetClipPath(path geom.Path) {
	if a == nil {
		return
	}
	a.clipPath = cloneArtistClipPath(path)
	a.hasClipPath = true
	a.clipPathSpec = CoordinateSpec{}
	a.hasClipSpec = false
	a.stale = true
}

// SetClipPathCoords stores an explicit path clip in the given coordinate
// system. The coordinate transform is resolved at draw time.
func (a *ArtistRasterization) SetClipPathCoords(path geom.Path, coords CoordinateSpec) {
	if a == nil {
		return
	}
	a.clipPath = cloneArtistClipPath(path)
	a.hasClipPath = true
	a.clipPathSpec = coords
	a.hasClipSpec = true
	a.clipPathTrans = geom.Affine{}
	a.hasClipTrans = false
	a.stale = true
}

// ClearClipPath removes the explicit path clip and any clip-path transform.
func (a *ArtistRasterization) ClearClipPath() {
	if a == nil {
		return
	}
	a.clipPath = geom.Path{}
	a.hasClipPath = false
	a.clipPathSpec = CoordinateSpec{}
	a.hasClipSpec = false
	a.clipPathTrans = geom.Affine{}
	a.hasClipTrans = false
	a.stale = true
}

// SetClipPathTransform stores an affine transform for the explicit path clip.
func (a *ArtistRasterization) SetClipPathTransform(transform geom.Affine) {
	if a == nil {
		return
	}
	a.clipPathTrans = transform
	a.hasClipTrans = true
	a.clipPathSpec = CoordinateSpec{}
	a.hasClipSpec = false
	a.stale = true
}

// ClearClipPathTransform removes the explicit path-clip transform.
func (a *ArtistRasterization) ClearClipPathTransform() {
	if a == nil {
		return
	}
	a.clipPathTrans = geom.Affine{}
	a.hasClipTrans = false
	a.stale = true
}

// ArtistClip reports the explicit clipping metadata for this artist.
func (a *ArtistRasterization) ArtistClip() ArtistClip {
	if a == nil {
		return ArtistClip{ClipOn: true}
	}
	return ArtistClip{
		ClipOn:            a.ClipOn(),
		ClipRect:          a.clipRect,
		HasClipRect:       a.hasClipRect,
		ClipPath:          cloneArtistClipPath(a.clipPath),
		HasClipPath:       a.hasClipPath,
		ClipPathCoords:    a.clipPathSpec,
		HasClipPathCoords: a.hasClipSpec,
		ClipPathTransform: a.clipPathTrans,
		HasClipPathTrans:  a.hasClipTrans,
	}
}

// SetTransform stores an explicit coordinate-to-display transform for the
// artist. Passing nil clears the explicit transform.
func (a *ArtistRasterization) SetTransform(tr transform.T) {
	if a == nil {
		return
	}
	a.transform = tr
	a.hasTransform = tr != nil
	a.stale = true
}

// ClearTransform removes the explicit coordinate-to-display transform.
func (a *ArtistRasterization) ClearTransform() {
	if a == nil {
		return
	}
	a.transform = nil
	a.hasTransform = false
	a.stale = true
}

// SetTransformCoords stores the coordinate system used by this artist when no
// explicit transform is configured.
func (a *ArtistRasterization) SetTransformCoords(spec CoordinateSpec) {
	if a == nil {
		return
	}
	a.transformSpec = spec
	a.hasTransSpec = true
	a.stale = true
}

// ClearTransformCoords removes the coordinate-system override.
func (a *ArtistRasterization) ClearTransformCoords() {
	if a == nil {
		return
	}
	a.transformSpec = CoordinateSpec{}
	a.hasTransSpec = false
	a.stale = true
}

// ArtistTransform reports the explicit transform metadata for this artist.
func (a *ArtistRasterization) ArtistTransform() ArtistTransform {
	if a == nil {
		return ArtistTransform{}
	}
	return ArtistTransform{
		Transform:    a.transform,
		HasTransform: a.hasTransform,
		Coords:       a.transformSpec,
		HasCoords:    a.hasTransSpec,
	}
}

// SetStale updates the artist stale flag.
func (a *ArtistRasterization) SetStale(stale bool) {
	if a == nil {
		return
	}
	a.stale = stale
}

// Stale reports whether artist metadata changed since it was last cleared.
func (a *ArtistRasterization) Stale() bool {
	return a != nil && a.stale
}

// SetRasterized toggles rasterized output for this artist when a vector
// renderer supports mixed raster/vector embedding.
func (a *ArtistRasterization) SetRasterized(enabled bool) {
	if a == nil {
		return
	}
	if enabled {
		a.rasterization.Mode = render.RasterizeAlways
		return
	}
	a.rasterization.Mode = render.RasterizeNever
}

// SetRasterization sets the complete rasterization policy for this artist.
func (a *ArtistRasterization) SetRasterization(options render.Rasterization) {
	if a == nil {
		return
	}
	a.rasterization = options
}

// Rasterization returns the configured mixed output policy.
func (a *ArtistRasterization) Rasterization() render.Rasterization {
	if a == nil {
		return render.Rasterization{}
	}
	return a.rasterization
}

type rasterizationProvider interface {
	Rasterization() render.Rasterization
}

type pathEffectRasterizationProvider interface {
	rasterizationPathEffects() []render.PathEffect
}

type visibilityProvider interface {
	Visible() bool
}

type animatedProvider interface {
	Animated() bool
}

type clipProvider interface {
	ArtistClip() ArtistClip
}

type artistTransformProvider interface {
	ArtistTransform() ArtistTransform
}

func artistTransformFor(ctx *DrawContext, art any, fallback CoordinateSpec) transform.T {
	if provider, ok := art.(artistTransformProvider); ok {
		options := provider.ArtistTransform()
		if options.HasTransform {
			return options.Transform
		}
		if options.HasCoords {
			if ctx == nil {
				return nil
			}
			return ctx.TransformFor(options.Coords)
		}
	}
	if ctx == nil {
		return nil
	}
	return ctx.TransformFor(fallback)
}

func artistUsesDataCoords(art any, fallback CoordinateSpec) bool {
	if provider, ok := art.(artistTransformProvider); ok {
		options := provider.ArtistTransform()
		if options.HasTransform {
			return false
		}
		if options.HasCoords {
			return isDataCoords(options.Coords)
		}
	}
	return isDataCoords(fallback)
}

func drawArtist(r render.Renderer, ctx *DrawContext, art Artist) {
	if art == nil {
		return
	}
	if provider, ok := art.(visibilityProvider); ok && !provider.Visible() {
		return
	}
	if !drawSelectByAnimated(ctx, art) {
		return
	}
	draw := func() {
		drawRasterizedArtist(r, ctx, art)
	}
	drawWithArtistClip(r, ctx, art, draw)
}

// drawSelectByAnimated returns whether art should be drawn under the current
// animation filter. The default (zero DrawOptions) draws non-animated artists
// and skips animated ones, matching Matplotlib's Artist.draw_wrapper behavior.
// During an animation overlay pass the engine flips DrawContext.DrawOptions to
// include animated artists; during a background-snapshot pass it explicitly
// excludes them.
func drawSelectByAnimated(ctx *DrawContext, art Artist) bool {
	if ctx == nil {
		return true
	}
	mode := ctx.DrawOptions.AnimatedFilter
	if mode == AnimatedFilterAll {
		return true
	}
	provider, ok := art.(animatedProvider)
	isAnimated := ok && provider.Animated()
	switch mode {
	case AnimatedFilterExcludeAnimated:
		return !isAnimated
	case AnimatedFilterOnlyAnimated:
		return isAnimated
	default:
		return !isAnimated
	}
}

func drawRasterizedArtist(r render.Renderer, ctx *DrawContext, art Artist) {
	options, ok := artistRasterizationOptions(r, art, ctx)
	if !ok {
		art.Draw(r, ctx)
		return
	}
	controller, ok := r.(render.RasterizationController)
	if !ok || !controller.StartRasterized(options) {
		art.Draw(r, ctx)
		return
	}
	defer controller.StopRasterized()
	art.Draw(r, ctx)
}

func drawOverlayArtist(r render.Renderer, ctx *DrawContext, art Artist, overlay OverlayArtist) {
	if overlay == nil {
		return
	}
	if provider, ok := art.(visibilityProvider); ok && !provider.Visible() {
		return
	}
	if !drawSelectByAnimated(ctx, art) {
		return
	}
	draw := func() {
		drawRasterizedOverlayArtist(r, ctx, art, overlay)
	}
	drawWithArtistClip(r, ctx, art, draw)
}

func drawRasterizedOverlayArtist(r render.Renderer, ctx *DrawContext, art Artist, overlay OverlayArtist) {
	options, ok := explicitArtistRasterizationOptions(art, ctx)
	if !ok {
		overlay.DrawOverlay(r, ctx)
		return
	}
	controller, ok := r.(render.RasterizationController)
	if !ok || !controller.StartRasterized(options) {
		overlay.DrawOverlay(r, ctx)
		return
	}
	defer controller.StopRasterized()
	overlay.DrawOverlay(r, ctx)
}

func drawWithArtistClip(r render.Renderer, ctx *DrawContext, art Artist, draw func()) {
	clip, ok := artistClipOptions(art)
	if !ok {
		draw()
		return
	}
	r.Save()
	if clip.HasClipRect {
		r.ClipRect(clip.ClipRect)
	}
	if clip.HasClipPath {
		applyArtistClipPath(r, ctx, clip)
	}
	draw()
	r.Restore()
}

func applyArtistClipPath(r render.Renderer, ctx *DrawContext, clip ArtistClip) {
	switch {
	case clip.HasClipPathCoords:
		if ctx == nil {
			r.ClipPath(clip.ClipPath)
			return
		}
		if affine, ok := ctx.AffineTransformFor(clip.ClipPathCoords); ok {
			clipPathTransformed(r, clip.ClipPath, affine)
			return
		}
		if tr := ctx.TransformFor(clip.ClipPathCoords); tr != nil {
			r.ClipPath(applyTransformPath(clip.ClipPath, tr))
			return
		}
		r.ClipPath(clip.ClipPath)
	case clip.HasClipPathTrans:
		clipPathTransformed(r, clip.ClipPath, clip.ClipPathTransform)
	default:
		r.ClipPath(clip.ClipPath)
	}
}

func clipPathTransformed(r render.Renderer, path geom.Path, affine geom.Affine) {
	if transformer, ok := r.(render.ClipPathTransformer); ok {
		transformer.ClipPathTransformed(path, affine)
		return
	}
	r.ClipPath(applyAffinePath(path, affine))
}

func artistClipOptions(art Artist) (ArtistClip, bool) {
	provider, ok := art.(clipProvider)
	if !ok {
		return ArtistClip{}, false
	}
	clip := provider.ArtistClip()
	if !clip.ClipOn || (!clip.HasClipRect && !clip.HasClipPath) {
		return ArtistClip{}, false
	}
	return clip, true
}

func cloneArtistClipPath(path geom.Path) geom.Path {
	if len(path.C) == 0 && len(path.V) == 0 {
		return geom.Path{}
	}
	return geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: append([]geom.Pt(nil), path.V...),
	}
}

func explicitArtistRasterizationOptions(art Artist, ctx *DrawContext) (render.Rasterization, bool) {
	if provider, ok := art.(rasterizationProvider); ok {
		options := provider.Rasterization()
		switch options.Mode {
		case render.RasterizeAlways, render.RasterizeAuto:
			return withRasterizationDPI(options, ctx), true
		}
	}
	return render.Rasterization{}, false
}

func artistRasterizationOptions(r render.Renderer, art Artist, ctx *DrawContext) (render.Rasterization, bool) {
	if options, ok := explicitArtistRasterizationOptions(art, ctx); ok {
		return options, true
	}
	if provider, ok := art.(rasterizationProvider); ok && provider.Rasterization().Mode == render.RasterizeNever {
		return render.Rasterization{}, false
	}

	if !shouldAutoRasterizeArtist(r, art) {
		return render.Rasterization{}, false
	}
	return withRasterizationDPI(render.Rasterization{Mode: render.RasterizeAuto}, ctx), true
}

func withRasterizationDPI(options render.Rasterization, ctx *DrawContext) render.Rasterization {
	if options.DPI <= 0 && ctx != nil && ctx.RC.DPI > 0 {
		options.DPI = ctx.RC.DPI
	}
	return options
}

func shouldAutoRasterizeArtist(r render.Renderer, art Artist) bool {
	if _, ok := r.(render.RasterizationController); !ok {
		return false
	}
	if denseScatterArtist(art) {
		return true
	}
	if denseContourArtist(art) {
		return true
	}
	if hasUnsupportedFilterPathEffect(r, art) {
		return true
	}
	return false
}

func denseScatterArtist(art Artist) bool {
	scatter, ok := art.(*Scatter2D)
	return ok && scatter != nil && len(scatter.XY) >= autoRasterizeScatterPointThreshold
}

func denseContourArtist(art Artist) bool {
	contour, ok := art.(*ContourSet)
	if !ok || contour == nil {
		return false
	}
	return contourPathCount(contour) >= autoRasterizeContourPathThreshold
}

func contourPathCount(contour *ContourSet) int {
	if contour == nil {
		return 0
	}
	count := 0
	if contour.Lines != nil {
		count += len(contour.Lines.Segments)
	}
	if contour.Fills != nil {
		count += len(contour.Fills.Polygons)
		count += len(contour.Fills.Paths)
	}
	return count
}

func hasUnsupportedFilterPathEffect(r render.Renderer, art Artist) bool {
	if _, ok := r.(render.FilterRenderer); ok {
		return false
	}
	provider, ok := art.(pathEffectRasterizationProvider)
	if !ok {
		return false
	}
	nativeFilter, hasNativeFilter := r.(render.PathEffectFilterDrawer)
	for _, effect := range provider.rasterizationPathEffects() {
		if effect.Kind != render.PathEffectFilter {
			continue
		}
		if !hasNativeFilter || !nativeFilter.SupportsPathEffectFilter(effect) {
			return true
		}
	}
	return false
}

func (l *Line2D) rasterizationPathEffects() []render.PathEffect {
	if l == nil {
		return nil
	}
	return l.PathEffects
}

func (s *Scatter2D) rasterizationPathEffects() []render.PathEffect {
	if s == nil {
		return nil
	}
	return s.PathEffects
}

func (c *Collection) rasterizationPathEffects() []render.PathEffect {
	if c == nil {
		return nil
	}
	return c.PathEffects
}

func (p *Patch) rasterizationPathEffects() []render.PathEffect {
	if p == nil {
		return nil
	}
	return p.PathEffects
}

package core

import "github.com/cwbudde/matplotlib-go/render"

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
	alpha         float64
	alphaSet      bool
	inLayout      bool
	inLayoutSet   bool
	stale         bool
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

func drawArtist(r render.Renderer, ctx *DrawContext, art Artist) {
	if art == nil {
		return
	}
	if provider, ok := art.(visibilityProvider); ok && !provider.Visible() {
		return
	}
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

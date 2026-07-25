package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const (
	patchCircleSegments = 48
	// patchArcKappa aliases geom.BezierCircleKappa so the 4-cubic circle
	// approximation has a single source of truth in the geom package.
	patchArcKappa = geom.BezierCircleKappa
)

// BoxStyle controls the shape used by FancyBboxPatch.
type BoxStyle uint8

const (
	BoxStyleSquare BoxStyle = iota
	BoxStyleRound
	BoxStyleCircle
	BoxStyleEllipse
	BoxStyleRArrow
	BoxStyleLArrow
	BoxStyleDArrow
	BoxStyleRound4
	BoxStyleSawtooth
	BoxStyleRoundtooth
)

// Patch stores the common face/edge styling shared by patch-like artists.
//
// In Go this acts as the reusable embedded base for concrete patch artists
// rather than a directly-instantiable drawable type.
type Patch struct {
	ArtistRasterization
	FaceColor render.Color
	EdgeColor render.Color
	EdgeWidth float64
	// The Set flags distinguish explicit transparent/zero values from omitted
	// values that inherit patch.* rcParams.
	faceColorSet bool
	edgeColorSet bool
	edgeWidthSet bool
	Alpha        float64
	Dashes       []float64
	DashUnits    DashUnits
	Label        string

	Hatch        string
	HatchColor   render.Color
	HatchWidth   float64
	HatchSpacing float64
	PathEffects  []render.PathEffect

	LineJoin render.LineJoin
	LineCap  render.LineCap
	// Antialias overrides patch.antialiased when non-default.
	Antialias render.AntialiasMode
	// Sketch is a per-artist sketch/xkcd override; the zero value inherits the
	// figure default.
	Sketch render.SketchParams
	z      float64

	// pathCache holds the persistent display-path projection cache (Phase 13).
	// Every artist embedding Patch reuses its non-affine projection across
	// affine-only redraws; the zero value is an empty cache filled on first draw.
	pathCache  displayPathCache
	rcDefaults *style.RC
}

// SetFaceColor explicitly sets the patch fill. A transparent color remains
// transparent even when patch.facecolor is visible.
func (p *Patch) SetFaceColor(color render.Color) {
	if p == nil {
		return
	}
	p.FaceColor = color
	p.faceColorSet = true
}

// SetEdgeColor explicitly sets the patch edge. A transparent color suppresses
// the edge even when patch.force_edgecolor is enabled.
func (p *Patch) SetEdgeColor(color render.Color) {
	if p == nil {
		return
	}
	p.EdgeColor = color
	p.edgeColorSet = true
}

// SetEdgeWidth explicitly sets the patch edge width. Zero suppresses the edge
// instead of inheriting patch.linewidth.
func (p *Patch) SetEdgeWidth(width float64) {
	if p == nil {
		return
	}
	p.EdgeWidth = width
	p.edgeWidthSet = true
}

func (p *Patch) patchBase() *Patch { return p }

// displayPathCacheSlot exposes the embedded display-path cache so the shared
// buildArtistDisplayPath can reuse this artist's non-affine projection across
// draws. Promoted to every type embedding Patch.
func (p *Patch) displayPathCacheSlot() *displayPathCache {
	if p == nil {
		return nil
	}
	return &p.pathCache
}

// AddPatch mirrors Matplotlib's patch-oriented API on top of the generic Add.
func (a *Axes) AddPatch(art Artist) {
	if a != nil && art != nil {
		if provider, ok := art.(interface{ patchBase() *Patch }); ok {
			rc := a.resolvedRC()
			provider.patchBase().rcDefaults = &rc
		}
		a.Add(art)
	}
}

// Z returns the patch z-order for sorting.
func (p *Patch) Z() float64 {
	if p == nil {
		return 0
	}
	return zOrDefault(p.z, defaultPatchZ)
}

func (p *Patch) legendEntry() (legendEntry, bool) {
	if p == nil || p.Label == "" {
		return legendEntry{}, false
	}
	rc := style.Default
	if p.rcDefaults != nil {
		rc = *p.rcDefaults
	}
	return legendEntryFromPatchStyle(
		p.Label,
		p.resolvedFaceColorForRC(&rc),
		p.resolvedEdgeColorForRC(&rc),
		p.resolvedEdgeWidth(&rc),
		p.Hatch,
		p.resolvedHatchColor(),
		p.resolvedHatchWidth(),
	), true
}

func (p *Patch) resolvedFaceColor() render.Color {
	if p == nil {
		return render.Color{}
	}
	return patchAlphaColor(p.FaceColor, p.Alpha)
}

func (p *Patch) resolvedFaceColorForRC(rc *style.RC) render.Color {
	color := p.resolvedFaceColor()
	if p != nil && !p.faceColorSet && p.FaceColor == (render.Color{}) && rc != nil {
		color = patchAlphaColor(rc.DefaultPatchFaceColor(), p.Alpha)
	}
	return color
}

func (p *Patch) resolvedEdgeColor() render.Color {
	if p == nil {
		return render.Color{}
	}
	return patchAlphaColor(p.EdgeColor, p.Alpha)
}

func (p *Patch) resolvedEdgeColorForRC(rc *style.RC) render.Color {
	color := p.resolvedEdgeColor()
	if p != nil && !p.edgeColorSet && p.EdgeColor == (render.Color{}) &&
		rc != nil && rc.Patch.ForceEdgeColor {
		color = patchAlphaColor(rc.Patch.EdgeColor, p.Alpha)
	}
	return color
}

func (p *Patch) resolvedHatchColor() render.Color {
	if p == nil {
		return render.Color{}
	}

	color := p.HatchColor
	if color.A <= 0 {
		color = p.EdgeColor
	}
	if color.A <= 0 {
		color = p.FaceColor
	}
	if color.A <= 0 {
		color = render.Color{R: 0, G: 0, B: 0, A: 1}
	}
	return patchAlphaColor(color, p.Alpha)
}

// resolvedHatchWidth returns the hatch line width in points (matplotlib
// hatch.linewidth defaults to 1.0 pt). Callers convert to device pixels.
func (p *Patch) resolvedHatchWidth() float64 {
	if p == nil || p.HatchWidth <= 0 {
		return 1.0
	}
	return p.HatchWidth
}

func (p *Patch) resolvedHatchSpacing() float64 {
	if p == nil || p.HatchSpacing <= 0 {
		return render.DefaultHatchSpacing
	}
	return p.HatchSpacing
}

func patchAlphaColor(color render.Color, alpha float64) render.Color {
	if alpha > 0 && alpha <= 1 {
		color = color.WithAlphaMultiplier(alpha)
	}
	if color.A < 0 {
		color.A = 0
	}
	if color.A > 1 {
		color.A = 1
	}
	return color
}

func (p *Patch) resolvedAntialias(rc *style.RC) render.AntialiasMode {
	if p != nil && p.Antialias != render.AntialiasDefault {
		return p.Antialias
	}
	if rc != nil && !rc.Patch.Antialiased {
		return render.AntialiasOff
	}
	return render.AntialiasOn
}

func (p *Patch) resolvedEdgeWidth(rc *style.RC) float64 {
	if p != nil && (p.edgeWidthSet || p.EdgeWidth > 0) {
		return p.EdgeWidth
	}
	if rc != nil {
		return rc.Patch.LineWidth
	}
	return 0
}

func (p *Patch) strokePaint(rc *style.RC, color render.Color, edgeWidth float64) render.Paint {
	widthPx := pointsToPixels(*rc, edgeWidth)
	return render.Paint{
		Stroke:      color,
		LineWidth:   widthPx,
		LineJoin:    p.LineJoin,
		LineCap:     p.LineCap,
		Dashes:      patchDashesForPaint(p.Dashes, widthPx, p.DashUnits),
		PathEffects: devicePathEffects(*rc, p.PathEffects),
		Snap:        render.SnapAuto,
		Sketch:      p.Sketch,
		Antialias:   p.resolvedAntialias(rc),
	}
}

func patchDashesForPaint(dashes []float64, edgeWidth float64, units DashUnits) []float64 {
	return lineDashesForPaint(dashes, edgeWidth, units)
}

func (p *Patch) drawStyledPath(r render.Renderer, rc *style.RC, fillPath, strokePath geom.Path) {
	if p == nil || r == nil {
		return
	}
	if len(fillPath.C) == 0 && len(strokePath.C) == 0 {
		return
	}

	faceColor := p.resolvedFaceColorForRC(rc)
	edgeColor := p.resolvedEdgeColorForRC(rc)
	edgeWidth := p.resolvedEdgeWidth(rc)
	hasEdge := edgeWidth > 0 && edgeColor.A > 0
	edgeWidthPx := pointsToPixels(*rc, edgeWidth)
	nativeHatch := false
	if hatcher, ok := r.(render.NativeHatcher); ok {
		nativeHatch = hatcher.SupportsNativeHatch()
	}

	if len(fillPath.C) > 0 {
		paint := render.Paint{
			Fill: faceColor, Snap: render.SnapAuto, Sketch: p.Sketch,
			Antialias: p.resolvedAntialias(rc),
		}
		paint.PathEffects = devicePathEffects(*rc, p.PathEffects)
		if nativeHatch && p.Hatch != "" {
			paint.Hatch = p.Hatch
			paint.HatchColor = p.resolvedHatchColor()
			paint.HatchLineWidth = pointsToPixels(*rc, p.resolvedHatchWidth())
			paint.HatchSpacing = p.resolvedHatchSpacing()
		}
		combinedStroke := len(strokePath.C) == 0 && hasEdge
		if combinedStroke {
			if p.Hatch == "" {
				paint = p.strokePaint(rc, edgeColor, edgeWidth)
				paint.Fill = faceColor
			} else if nativeHatch {
				paint.Stroke = edgeColor
				paint.LineWidth = edgeWidthPx
				paint.LineJoin = p.LineJoin
				paint.LineCap = p.LineCap
				paint.Dashes = patchDashesForPaint(p.Dashes, edgeWidthPx, p.DashUnits)
			}
		}
		if faceColor.A > 0 || combinedStroke || (nativeHatch && p.Hatch != "") {
			r.Path(fillPath, &paint)
		}
	}

	if len(fillPath.C) > 0 && p.Hatch != "" {
		if !nativeHatch {
			p.drawHatch(r, rc, fillPath)
		}
		if !nativeHatch && len(strokePath.C) == 0 && hasEdge {
			paint := p.strokePaint(rc, edgeColor, edgeWidth)
			r.Path(fillPath, &paint)
		}
	}

	if len(strokePath.C) > 0 && hasEdge {
		paint := p.strokePaint(rc, edgeColor, edgeWidth)
		r.Path(strokePath, &paint)
	}
}

func (p *Patch) drawHatch(r render.Renderer, rc *style.RC, clipPath geom.Path) {
	render.DrawHatchFallback(r, clipPath, render.Paint{
		Hatch:          p.Hatch,
		HatchColor:     p.resolvedHatchColor(),
		HatchLineWidth: pointsToPixels(*rc, p.resolvedHatchWidth()),
		HatchSpacing:   p.resolvedHatchSpacing(),
		Antialias:      p.resolvedAntialias(rc),
	})
}

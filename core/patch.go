package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
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
	Alpha     float64
	Dashes    []float64
	DashUnits DashUnits
	Label     string

	Hatch        string
	HatchColor   render.Color
	HatchWidth   float64
	HatchSpacing float64
	PathEffects  []render.PathEffect

	LineJoin render.LineJoin
	LineCap  render.LineCap
	// Sketch is a per-artist sketch/xkcd override; the zero value inherits the
	// figure default.
	Sketch render.SketchParams
	z      float64

	// pathCache holds the persistent display-path projection cache (Phase 13).
	// Every artist embedding Patch reuses its non-affine projection across
	// affine-only redraws; the zero value is an empty cache filled on first draw.
	pathCache displayPathCache
}

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
	return legendEntryFromPatchStyle(
		p.Label,
		p.resolvedFaceColor(),
		p.resolvedEdgeColor(),
		p.EdgeWidth,
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

func (p *Patch) resolvedEdgeColor() render.Color {
	if p == nil {
		return render.Color{}
	}
	return patchAlphaColor(p.EdgeColor, p.Alpha)
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

func (p *Patch) resolvedHatchWidth() float64 {
	if p == nil || p.HatchWidth <= 0 {
		return 100.0 / 72.0
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
		color.A *= alpha
	}
	if color.A < 0 {
		color.A = 0
	}
	if color.A > 1 {
		color.A = 1
	}
	return color
}

func (p *Patch) strokePaint(color render.Color) render.Paint {
	return render.Paint{
		Stroke:      color,
		LineWidth:   p.EdgeWidth,
		LineJoin:    p.LineJoin,
		LineCap:     p.LineCap,
		Dashes:      patchDashesForPaint(p.Dashes, p.EdgeWidth, p.DashUnits),
		PathEffects: cloneRenderPathEffects(p.PathEffects),
		Snap:        render.SnapAuto,
		Sketch:      p.Sketch,
	}
}

func patchDashesForPaint(dashes []float64, edgeWidth float64, units DashUnits) []float64 {
	return lineDashesForPaint(dashes, edgeWidth, units)
}

func (p *Patch) drawStyledPath(r render.Renderer, fillPath, strokePath geom.Path) {
	if p == nil || r == nil {
		return
	}
	if len(fillPath.C) == 0 && len(strokePath.C) == 0 {
		return
	}

	faceColor := p.resolvedFaceColor()
	edgeColor := p.resolvedEdgeColor()
	hasEdge := p.EdgeWidth > 0 && edgeColor.A > 0
	nativeHatch := false
	if hatcher, ok := r.(render.NativeHatcher); ok {
		nativeHatch = hatcher.SupportsNativeHatch()
	}

	if len(fillPath.C) > 0 {
		paint := render.Paint{Fill: faceColor, Snap: render.SnapAuto, Sketch: p.Sketch}
		paint.PathEffects = cloneRenderPathEffects(p.PathEffects)
		if nativeHatch && p.Hatch != "" {
			paint.Hatch = p.Hatch
			paint.HatchColor = p.resolvedHatchColor()
			paint.HatchLineWidth = p.resolvedHatchWidth()
			paint.HatchSpacing = p.resolvedHatchSpacing()
		}
		combinedStroke := len(strokePath.C) == 0 && hasEdge
		if combinedStroke {
			if p.Hatch == "" {
				paint = p.strokePaint(edgeColor)
				paint.Fill = faceColor
			} else if nativeHatch {
				paint.Stroke = edgeColor
				paint.LineWidth = p.EdgeWidth
				paint.LineJoin = p.LineJoin
				paint.LineCap = p.LineCap
				paint.Dashes = patchDashesForPaint(p.Dashes, p.EdgeWidth, p.DashUnits)
			}
		}
		if faceColor.A > 0 || combinedStroke || (nativeHatch && p.Hatch != "") {
			r.Path(fillPath, &paint)
		}
	}

	if len(fillPath.C) > 0 && p.Hatch != "" {
		if !nativeHatch {
			p.drawHatch(r, fillPath)
		}
		if !nativeHatch && len(strokePath.C) == 0 && hasEdge {
			paint := p.strokePaint(edgeColor)
			r.Path(fillPath, &paint)
		}
	}

	if len(strokePath.C) > 0 && hasEdge {
		paint := p.strokePaint(edgeColor)
		r.Path(strokePath, &paint)
	}
}

func (p *Patch) drawHatch(r render.Renderer, clipPath geom.Path) {
	render.DrawHatchFallback(r, clipPath, render.Paint{
		Hatch:          p.Hatch,
		HatchColor:     p.resolvedHatchColor(),
		HatchLineWidth: p.resolvedHatchWidth(),
		HatchSpacing:   p.resolvedHatchSpacing(),
	})
}

package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

func (a *Axes) SetXLim(minVal, maxVal float64) {
	target := a.xScaleRoot()
	target.XScale = replaceScaleDomain(target.XScale, minVal, maxVal)
	target.xLimitsManual = true
	target.refreshUnitAxis(true)
}

func (a *Axes) SetYLim(minVal, maxVal float64) {
	target := a.yScaleRoot()
	target.YScale = replaceScaleDomain(target.YScale, minVal, maxVal)
	target.yLimitsManual = true
	target.refreshUnitAxis(false)
}

func (a *Axes) SetXScale(name string, opts ...transform.ScaleOption) error {
	return a.setScale(true, name, opts...)
}

func (a *Axes) SetYScale(name string, opts ...transform.ScaleOption) error {
	return a.setScale(false, name, opts...)
}

func (a *Axes) SetXLimLog(minVal, maxVal, base float64) {
	target := a.xScaleRoot()
	if state := target.unitState(true); state != nil && !state.scaleCompatible("log") {
		return
	}
	target.XScale = transform.NewLog(minVal, maxVal, base)
	target.xLimitsManual = true
	configureScaleAxes(target.XAxis, target.XAxisTop, "log", transform.ResolveScaleOptions(
		transform.WithScaleDomain(minVal, maxVal),
		transform.WithScaleBase(base),
	))
}

func (a *Axes) SetYLimLog(minVal, maxVal, base float64) {
	target := a.yScaleRoot()
	if state := target.unitState(false); state != nil && !state.scaleCompatible("log") {
		return
	}
	target.YScale = transform.NewLog(minVal, maxVal, base)
	target.yLimitsManual = true
	configureScaleAxes(target.YAxis, target.YAxisRight, "log", transform.ResolveScaleOptions(
		transform.WithScaleDomain(minVal, maxVal),
		transform.WithScaleBase(base),
	))
}

func (a *Axes) InvertX() {
	target := a.xScaleRoot()
	if target == nil || target.XScale == nil {
		return
	}
	target.XScale = toggleInvertedScale(target.XScale)
}

func (a *Axes) InvertY() {
	target := a.yScaleRoot()
	if target == nil || target.YScale == nil {
		return
	}
	target.YScale = toggleInvertedScale(target.YScale)
}

func (a *Axes) XInverted() bool {
	if a == nil {
		return false
	}
	return scaleDomainDescending(a.effectiveXScale())
}

func (a *Axes) YInverted() bool {
	if a == nil {
		return false
	}
	return scaleDomainDescending(a.effectiveYScale())
}

func (a *Axes) SetAspect(mode string, value ...float64) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "auto":
		a.aspectMode = "auto"
		a.aspectValue = 1
	case "equal":
		a.aspectMode = "equal"
		a.aspectValue = 1
	case "ratio":
		if len(value) == 0 || value[0] <= 0 || math.IsNaN(value[0]) || math.IsInf(value[0], 0) {
			return fmt.Errorf("ratio aspect requires a positive finite value")
		}
		a.aspectMode = "ratio"
		a.aspectValue = value[0]
	default:
		return fmt.Errorf("unsupported aspect mode %q", mode)
	}
	return nil
}

// SetAdjustable selects how an aspect constraint is satisfied: "box" (the
// default — shrink the axes rectangle) or "datalim" (keep the box and expand the
// data limits). Mirrors Matplotlib Axes.set_adjustable; "datalim" is rejected on
// shared axes, matching upstream.
func (a *Axes) SetAdjustable(adjustable string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(adjustable)) {
	case "", "box":
		a.adjustable = ""
	case "datalim":
		if a.shareX != nil || a.shareY != nil {
			return fmt.Errorf("adjustable 'datalim' is incompatible with shared axes")
		}
		a.adjustable = "datalim"
	default:
		return fmt.Errorf("unsupported adjustable %q", adjustable)
	}
	a.autoScaleIfEnabled(defaultAutoScaleMargin)
	return nil
}

// SetAnchor sets how an aspect-shrunk axes is positioned within its original
// rectangle, using Matplotlib's cardinal anchor names ("C", "N", "NE", "E",
// "SE", "S", "SW", "W", "NW").
func (a *Axes) SetAnchor(anchor string) error {
	if a == nil {
		return nil
	}
	name := strings.ToUpper(strings.TrimSpace(anchor))
	if name == "" {
		name = "C"
	}
	if _, ok := anchorCoefs[name]; !ok {
		return fmt.Errorf("unsupported anchor %q", anchor)
	}
	a.anchor = name
	return nil
}

// anchorCoefs maps Matplotlib anchor names to the (cx, cy) fraction of the
// leftover space placed below/left of the shrunk box (Bbox.coefs).
var anchorCoefs = map[string][2]float64{
	"C":  {0.5, 0.5},
	"SW": {0, 0},
	"S":  {0.5, 0},
	"SE": {1, 0},
	"E":  {1, 0.5},
	"NE": {1, 1},
	"N":  {0.5, 1},
	"NW": {0, 1},
	"W":  {0, 0.5},
}

func (a *Axes) anchorCoef() (float64, float64) {
	if c, ok := anchorCoefs[a.anchor]; ok {
		return c[0], c[1]
	}
	return 0.5, 0.5 // default center anchor
}

func (a *Axes) SetAxisEqual() {
	_ = a.SetAspect("equal")
}

func (a *Axes) SetBoxAspect(aspect float64) error {
	if a == nil {
		return nil
	}
	if aspect <= 0 || math.IsNaN(aspect) || math.IsInf(aspect, 0) {
		return fmt.Errorf("box aspect must be a positive finite value")
	}
	a.boxAspect = aspect
	return nil
}

func (a *Axes) ClearBoxAspect() {
	if a == nil {
		return
	}
	a.boxAspect = 0
}

func (a *Axes) layout(f *Figure) (pixelRect geom.Rect) {
	// Display space is y-up with a bottom-left origin (Matplotlib convention).
	// Figure fractions map directly to display pixels without a Y flip; the
	// device y-inversion is owned by the backend at rasterization.
	minPt := geom.Pt{X: f.SizePx.X * a.RectFraction.Min.X, Y: f.SizePx.Y * a.RectFraction.Min.Y}
	maxPt := geom.Pt{X: f.SizePx.X * a.RectFraction.Max.X, Y: f.SizePx.Y * a.RectFraction.Max.Y}
	return geom.Rect{Min: minPt, Max: maxPt}
}

func (a *Axes) adjustedLayout(f *Figure) geom.Rect {
	px := a.layout(f)
	if a.colorbarParent != nil {
		return a.adjustedColorbarLayout(f, px)
	}
	// adjustable='datalim' keeps the full axes box; the aspect is satisfied by
	// expanding data limits during autoscale (see applyAspectDatalim).
	if a.adjustable == "datalim" && a.boxAspect <= 0 {
		return px
	}
	target := 0.0
	if a.boxAspect > 0 {
		target = a.boxAspect
	} else {
		switch a.aspectMode {
		case "equal":
			target = a.dataAspectTarget(1)
		case "ratio":
			target = a.dataAspectTarget(a.aspectValue)
		}
	}
	if target <= 0 || math.IsNaN(target) || math.IsInf(target, 0) {
		return px
	}
	cx, cy := a.anchorCoef()
	return rectWithAspectInFigureFraction(a.RectFraction, target, f, cx, cy)
}

func (a *Axes) adjustedColorbarLayout(f *Figure, px geom.Rect) geom.Rect {
	if a == nil || f == nil || f.SizePx.X <= 0 || f.SizePx.Y <= 0 {
		return px
	}
	if colorbarIsHorizontal(a.colorbarLocation) {
		return px
	}
	width := a.colorbarWidth
	if width > 0 {
		width *= f.SizePx.X
	} else {
		aspect := resolvedColorbarAspect(a.colorbarAspect)
		if aspect > 0 {
			aspect *= colorbarExtensionShrink(a.colorbarExtend)
			width = px.H() / aspect
		}
	}
	if width <= 0 || width >= px.W() {
		return px
	}
	px.Max.X = px.Min.X + width
	return px
}

// applyAspectDatalim enforces an equal/ratio aspect with adjustable='datalim'
// by expanding (never shrinking) the data limits around their center so the
// visible pixels-per-data scale matches on both axes. It is idempotent: once the
// ratio is satisfied within tolerance it makes no further change. Mirrors the
// datalim branch of Matplotlib's Axes.apply_aspect.
func (a *Axes) applyAspectDatalim() {
	if a == nil || a.adjustable != "datalim" || a.figure == nil || a.boxAspect > 0 {
		return
	}
	var aspect float64
	switch a.aspectMode {
	case "equal":
		aspect = 1
	case "ratio":
		aspect = a.aspectValue
	default:
		return
	}
	if aspect <= 0 {
		return
	}
	f := a.figure
	boxWpx := a.RectFraction.W() * f.SizePx.X
	boxHpx := a.RectFraction.H() * f.SizePx.Y
	if boxWpx <= 0 || boxHpx <= 0 {
		return
	}
	dataRatio := (boxHpx / boxWpx) / aspect

	xMin, xMax := currentScaleDomain(a.effectiveXScale())
	yMin, yMax := currentScaleDomain(a.effectiveYScale())
	xsize := math.Abs(xMax - xMin)
	ysize := math.Abs(yMax - yMin)
	if xsize <= 0 || ysize <= 0 {
		return
	}
	yExpander := dataRatio*xsize/ysize - 1.0
	if math.Abs(yExpander) < 0.005 {
		return
	}
	if yExpander > 0 {
		target := dataRatio * xsize
		yc := 0.5 * (yMin + yMax)
		a.YScale = replaceScaleDomain(a.YScale, yc-target/2, yc+target/2)
		a.refreshUnitAxis(false)
	} else {
		target := ysize / dataRatio
		xc := 0.5 * (xMin + xMax)
		a.XScale = replaceScaleDomain(a.XScale, xc-target/2, xc+target/2)
		a.refreshUnitAxis(true)
	}
}

func (a *Axes) dataAspectTarget(aspect float64) float64 {
	if a == nil || aspect <= 0 {
		return 0
	}
	xMin, xMax := currentScaleDomain(a.effectiveXScale())
	yMin, yMax := currentScaleDomain(a.effectiveYScale())
	xSpan := math.Abs(xMax - xMin)
	ySpan := math.Abs(yMax - yMin)
	if xSpan == 0 || ySpan == 0 {
		return 0
	}
	return aspect * ySpan / xSpan
}

// rectWithAspect shrinks r to the target height/width ratio and positions the
// shrunk box using the (cx, cy) anchor coefficients (0.5, 0.5 == centered).
func rectWithAspect(r geom.Rect, target, cx, cy float64) geom.Rect {
	if target <= 0 {
		return r
	}
	cur := r.H() / r.W()
	switch {
	case cur > target:
		newH := r.W() * target
		pad := r.H() - newH
		r.Min.Y += pad * cy
		r.Max.Y = r.Min.Y + newH
	case cur < target:
		newW := r.H() / target
		pad := r.W() - newW
		r.Min.X += pad * cx
		r.Max.X = r.Min.X + newW
	}
	return r
}

func rectWithAspectInFigureFraction(r geom.Rect, boxAspect float64, f *Figure, cx, cy float64) geom.Rect {
	if f == nil || f.SizePx.X <= 0 || f.SizePx.Y <= 0 || boxAspect <= 0 {
		return rectWithAspect(r, boxAspect, cx, cy)
	}
	figAspect := f.SizePx.Y / f.SizePx.X
	if figAspect <= 0 || math.IsNaN(figAspect) || math.IsInf(figAspect, 0) {
		return rectWithAspect(r, boxAspect, cx, cy)
	}
	frac := rectWithAspect(r, boxAspect/figAspect, cx, cy)
	return geom.Rect{
		Min: geom.Pt{
			X: f.SizePx.X * frac.Min.X,
			Y: f.SizePx.Y * frac.Min.Y,
		},
		Max: geom.Pt{
			X: f.SizePx.X * frac.Max.X,
			Y: f.SizePx.Y * frac.Max.Y,
		},
	}
}

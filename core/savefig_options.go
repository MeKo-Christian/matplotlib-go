package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

// resolvedSaveFigure carries the save-time figure overrides after layering
// per-call render.FigureOptions over the figure's savefig.* rcParams.
type resolvedSaveFigure struct {
	dpi          float64 // 0 means "use the figure DPI"
	facecolor    render.Color
	hasFacecolor bool
	edgecolor    render.Color
	hasEdgecolor bool
	transparent  bool
	bboxTight    bool
	padInches    float64
	format       string // normalized, no leading dot; "" means infer from path
}

// parseSavefigColor resolves a savefig color string. The "auto"/"" sentinels
// report ok=false so the caller falls back to the figure background.
func parseSavefigColor(value string) (render.Color, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.EqualFold(trimmed, "auto") {
		return render.Color{}, false
	}
	c, err := color.ToRGBA(trimmed)
	if err != nil {
		return render.Color{}, false
	}
	return c, true
}

// resolveSaveFigureOptions layers per-call FigureOptions over fig.RC.Savefig.
func resolveSaveFigureOptions(fig *Figure, figOpts *render.FigureOptions) resolvedSaveFigure {
	sf := fig.RC.Savefig
	out := resolvedSaveFigure{
		dpi:         sf.Dpi,
		transparent: sf.Transparent,
		bboxTight:   strings.EqualFold(strings.TrimSpace(sf.BboxInches), "tight"),
		padInches:   sf.PadInches,
		format:      strings.TrimSpace(strings.ToLower(sf.Format)),
	}
	if c, ok := parseSavefigColor(sf.Facecolor); ok {
		out.facecolor, out.hasFacecolor = c, true
	}
	if c, ok := parseSavefigColor(sf.Edgecolor); ok {
		out.edgecolor, out.hasEdgecolor = c, true
	}

	if figOpts.HasDPI {
		out.dpi = figOpts.DPI
	}
	if figOpts.HasFacecolor {
		out.facecolor, out.hasFacecolor = figOpts.Facecolor, true
	}
	if figOpts.HasEdgecolor {
		out.edgecolor, out.hasEdgecolor = figOpts.Edgecolor, true
	}
	if figOpts.HasTransparent {
		out.transparent = figOpts.Transparent
	}
	if figOpts.HasBbox {
		out.bboxTight = strings.EqualFold(strings.TrimSpace(figOpts.BboxInches), "tight")
	}
	if figOpts.HasPad {
		out.padInches = figOpts.PadInches
	}
	if f := strings.TrimSpace(strings.ToLower(figOpts.Format)); f != "" {
		out.format = f
	}
	return out
}

// prepareSaveFigure applies the resolved savefig.* overrides to the renderer and
// returns an effective figure (a shallow copy with adjusted RC) plus the
// DrawOptions to pass to DrawFigureWithOptions. The caller's figure is never
// mutated. It also reports the resolved options so format/bbox handling can use
// them.
func prepareSaveFigure(fig *Figure, r render.Renderer, figOpts *render.FigureOptions) (*Figure, DrawOptions, resolvedSaveFigure) {
	resolved := resolveSaveFigureOptions(fig, figOpts)
	eff := *fig

	resized := false
	if resolved.dpi > 0 && !almostEqualSaveDPI(resolved.dpi, fig.RC.DPI) {
		eff.RC.DPI = resolved.dpi
		if fig.RC.DPI > 0 {
			scale := resolved.dpi / fig.RC.DPI
			eff.SizePx = geom.Pt{
				X: math.Round(fig.SizePx.X * scale),
				Y: math.Round(fig.SizePx.Y * scale),
			}
			if resizer, ok := r.(render.Resizer); ok {
				if err := resizer.Resize(int(eff.SizePx.X), int(eff.SizePx.Y)); err == nil {
					resized = true
				}
			}
		}
	} else if resolved.dpi > 0 {
		eff.RC.DPI = resolved.dpi
	}

	var effBg render.Color
	switch {
	case resolved.transparent || !fig.RC.Figure.FrameOn:
		effBg = render.Color{}
	case resolved.hasFacecolor:
		effBg = resolved.facecolor
	default:
		effBg = fig.RC.FigureBackground()
	}
	eff.RC.Background = [4]float64{effBg.R, effBg.G, effBg.B, effBg.A}

	drawOpts := DrawOptions{Transparent: resolved.transparent || !fig.RC.Figure.FrameOn}
	if clearer, ok := r.(render.BackgroundClearer); ok && (resolved.transparent || !fig.RC.Figure.FrameOn || resolved.hasFacecolor || resized) {
		// A resized surface starts transparent, so it must be re-cleared to the
		// effective background even when no face color override was requested.
		clearer.Clear(effBg)
	} else if resolved.hasFacecolor && !resolved.transparent {
		bg := effBg
		drawOpts.FigureBackground = optional.Of(bg)
	}
	if resolved.hasEdgecolor && resolved.edgecolor.A > 0 {
		ec := resolved.edgecolor
		drawOpts.FigureEdge = optional.Of(ec)
		drawOpts.FigureEdgeWidth = pointsToPixels(eff.RC, 1.0)
	}

	return &eff, drawOpts, resolved
}

// prepareVectorSaveFigure expresses the effective figure in PDF points. Vector
// page geometry is tied to figure inches, not to either the figure's raster DPI
// or savefig.dpi. The latter is returned separately for rasterized sub-artists.
func prepareVectorSaveFigure(fig *Figure, r render.Renderer, figOpts *render.FigureOptions) (*Figure, DrawOptions, resolvedSaveFigure, float64) {
	resolved := resolveSaveFigureOptions(fig, figOpts)
	eff := *fig
	renderDPI := resolved.dpi
	if renderDPI <= 0 {
		renderDPI = fig.RC.DPI
	}
	if renderDPI <= 0 {
		renderDPI = 72
	}
	if fig.RC.DPI > 0 {
		eff.SizePx = geom.Pt{
			X: fig.SizePx.X * 72 / fig.RC.DPI,
			Y: fig.SizePx.Y * 72 / fig.RC.DPI,
		}
	}
	eff.RC.DPI = 72
	if sizer, ok := r.(render.VectorPageSizer); ok {
		sizer.SetPageSize(eff.SizePx.X, eff.SizePx.Y)
	}

	var effBg render.Color
	switch {
	case resolved.transparent || !fig.RC.Figure.FrameOn:
		effBg = render.Color{}
	case resolved.hasFacecolor:
		effBg = resolved.facecolor
	default:
		effBg = fig.RC.FigureBackground()
	}
	eff.RC.Background = [4]float64{effBg.R, effBg.G, effBg.B, effBg.A}

	drawOpts := DrawOptions{
		Transparent: resolved.transparent || !fig.RC.Figure.FrameOn,
		rasterDPI:   renderDPI,
	}
	if clearer, ok := r.(render.BackgroundClearer); ok && (resolved.transparent || !fig.RC.Figure.FrameOn || resolved.hasFacecolor) {
		clearer.Clear(effBg)
	} else if resolved.hasFacecolor && !resolved.transparent {
		drawOpts.FigureBackground = optional.Of(effBg)
	}
	if resolved.hasEdgecolor && resolved.edgecolor.A > 0 {
		drawOpts.FigureEdge = optional.Of(resolved.edgecolor)
		drawOpts.FigureEdgeWidth = pointsToPixels(eff.RC, 1)
	}

	return &eff, drawOpts, resolved, renderDPI
}

func almostEqualSaveDPI(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// rejectTightBboxForVector returns an error when a tight bounding box is
// requested for a vector format, which the port does not yet support (it would
// require analytical artist window extents rather than a raster scan).
func rejectTightBboxForVector(bboxTight bool, format string) error {
	if bboxTight {
		return fmt.Errorf("savefig: bbox=tight is not supported for %s output", format)
	}
	return nil
}

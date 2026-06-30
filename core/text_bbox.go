package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/diag"
	"github.com/cwbudde/matplotlib-go/render"
)

// textBBoxStyledPath builds the bbox path for an unpadded text box rect in local
// coordinates. The returned path already includes padding, which the chosen box
// style applies outward. cfg must be the resolved options from
// resolvedTextBBoxOptions. The second return is the pixel-snap mode to use.
func textBBoxStyledPath(localRect geom.Rect, cfg TextBBoxOptions) (geom.Path, render.SnapMode) {
	// Backward-compatible CornerRadius shortcut: a square box with rounded corners.
	if cfg.Style == "" && cfg.BoxStyle == BoxStyleSquare && cfg.CornerRadius > 0 {
		padded := geom.Rect{
			Min: geom.Pt{X: localRect.Min.X - cfg.Padding, Y: localRect.Min.Y - cfg.Padding},
			Max: geom.Pt{X: localRect.Max.X + cfg.Padding, Y: localRect.Max.Y + cfg.Padding},
		}
		return roundedRectPath(padded, cfg.CornerRadius), render.SnapOff
	}
	fb := FancyBboxPatch{
		BoxStyle:       cfg.BoxStyle,
		Pad:            cfg.Padding,
		RoundingSize:   cfg.RoundingSize,
		ToothSize:      cfg.ToothSize,
		ArrowHeadWidth: cfg.ArrowHeadWidth,
		ArrowHeadAngle: cfg.ArrowHeadAngle,
		MutationSize:   1,
	}
	path := fb.boxStylePath(localRect.Min.X, localRect.Min.Y, localRect.W(), localRect.H(), 1)
	snap := render.SnapAuto
	if cfg.BoxStyle != BoxStyleSquare {
		snap = render.SnapOff
	}
	return path, snap
}

func drawTextBBox(r render.Renderer, origin geom.Pt, layout singleLineTextLayout, opt *TextBBoxOptions, ctx *DrawContext, fontSize float64) {
	if opt == nil {
		return
	}
	localRect, ok := textLineBoxRect(origin, layout)
	if !ok {
		return
	}
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)

	path, snap := textBBoxStyledPath(localRect, cfg)
	r.Path(path, &render.Paint{
		Fill:      cfg.FaceColor,
		Stroke:    cfg.EdgeColor,
		LineWidth: pointsToPixels(ctx.RC, cfg.LineWidth),
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
		Snap:      snap,
	})
}

func drawTextBBoxRotated(r render.Renderer, anchor, drawOrigin geom.Pt, layout singleLineTextLayout, opt *TextBBoxOptions, ctx *DrawContext, fontSize, angleDeg float64) {
	path, ok := rotatedTextBBoxPath(anchor, drawOrigin, layout, opt, ctx, fontSize, angleDeg)
	if !ok {
		return
	}
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)

	r.Path(path, &render.Paint{
		Fill:      cfg.FaceColor,
		Stroke:    cfg.EdgeColor,
		LineWidth: pointsToPixels(ctx.RC, cfg.LineWidth),
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func rotatedTextBBoxPath(anchor, drawOrigin geom.Pt, layout singleLineTextLayout, opt *TextBBoxOptions, ctx *DrawContext, fontSize, angleDeg float64) (geom.Path, bool) {
	if opt == nil || layout.Width <= 0 || layout.Height <= 0 {
		return geom.Path{}, false
	}
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)
	angle := angleDeg * math.Pi / 180
	rotate := func(x, y, a float64) (float64, float64) {
		cosT := math.Cos(a)
		sinT := math.Sin(a)
		return x*cosT - y*sinT, x*sinT + y*cosT
	}

	// Port matplotlib.text._get_textbox for the single-line case. Both the
	// layout offsets and renderer geometry use Matplotlib's y-up display
	// coordinates here.
	lineX := drawOrigin.X - anchor.X
	lineY := drawOrigin.Y - anchor.Y
	x1, y1 := rotate(lineX, lineY, -angle)
	y1 -= layout.Descent
	x2 := x1 + layout.Width
	y2 := y1 + layout.Height
	xBox := math.Min(x1, x2)
	yBox := math.Min(y1, y2)
	wBox := math.Abs(x2 - x1)
	hBox := math.Abs(y2 - y1)
	xBox, yBox = rotate(xBox, yBox, angle)

	localRect := geom.Rect{Max: geom.Pt{X: wBox, Y: hBox}}
	path, _ := textBBoxStyledPath(localRect, cfg)
	for i := range path.V {
		x, y := rotate(path.V[i].X, path.V[i].Y, angle)
		path.V[i] = geom.Pt{X: anchor.X + xBox + x, Y: anchor.Y + yBox + y}
	}
	return path, true
}

func textBBoxRect(origin geom.Pt, layout singleLineTextLayout, opt *TextBBoxOptions, ctx *DrawContext, fontSize float64) (geom.Rect, bool) {
	if opt == nil {
		return geom.Rect{}, false
	}
	rect, ok := textLineBoxRect(origin, layout)
	if !ok {
		return geom.Rect{}, false
	}
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)
	rect.Min.X -= cfg.Padding
	rect.Min.Y -= cfg.Padding
	rect.Max.X += cfg.Padding
	rect.Max.Y += cfg.Padding
	return rect, true
}

func textLineBoxRect(origin geom.Pt, layout singleLineTextLayout) (geom.Rect, bool) {
	if layout.Width <= 0 || layout.Height <= 0 {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - layout.Descent},
		Max: geom.Pt{X: origin.X + layout.Width, Y: origin.Y + layout.Ascent},
	}, true
}

func drawMultilineTextBBox(r render.Renderer, rect geom.Rect, opt *TextBBoxOptions, ctx *DrawContext, fontSize float64) {
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)
	path, snap := textBBoxStyledPath(rect, cfg)
	r.Path(path, &render.Paint{
		Fill:      cfg.FaceColor,
		Stroke:    cfg.EdgeColor,
		LineWidth: pointsToPixels(ctx.RC, cfg.LineWidth),
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
		Snap:      snap,
	})
}

func drawMultilineTextBBoxRotated(r render.Renderer, rect geom.Rect, opt *TextBBoxOptions, ctx *DrawContext, fontSize float64, pivot geom.Pt, angleDeg float64) {
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)
	path, _ := textBBoxStyledPath(rect, cfg)
	path = rotatePathAround(path, pivot, -angleDeg)
	r.Path(path, &render.Paint{
		Fill:      cfg.FaceColor,
		Stroke:    cfg.EdgeColor,
		LineWidth: pointsToPixels(ctx.RC, cfg.LineWidth),
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func drawMultilineTextBBoxRotatedMatplotlib(r render.Renderer, anchor geom.Pt, block rotatedMultilineTextBlockLayout, opt *TextBBoxOptions, ctx *DrawContext, fontSize, angleDeg float64) {
	if opt == nil || block.TextBoxWidth <= 0 || block.TextBoxHeight <= 0 {
		return
	}
	cfg := resolvedTextBBoxOptions(*opt, ctx, fontSize)
	angle := angleDeg * math.Pi / 180
	localRect := geom.Rect{Max: geom.Pt{X: block.TextBoxWidth, Y: block.TextBoxHeight}}
	path, _ := textBBoxStyledPath(localRect, cfg)
	for i := range path.V {
		rot := rotateTextLayoutPoint(path.V[i], angle)
		path.V[i] = geom.Pt{
			X: anchor.X + block.TextBoxX + rot.X,
			Y: anchor.Y + block.TextBoxY + rot.Y,
		}
	}
	r.Path(path, &render.Paint{
		Fill:      cfg.FaceColor,
		Stroke:    cfg.EdgeColor,
		LineWidth: pointsToPixels(ctx.RC, cfg.LineWidth),
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func resolvedTextBBoxOptions(opt TextBBoxOptions, ctx *DrawContext, fontSize float64) TextBBoxOptions {
	if opt.Style != "" {
		if spec, err := parseBoxStyleSpec(opt.Style); err != nil {
			// The draw path can't return an error; fall back to a plain square
			// box (matplotlib's default boxstyle) and surface the problem.
			diag.Warnf("text bbox: %v; using square box", err)
			opt.BoxStyle = BoxStyleSquare
		} else {
			opt.BoxStyle = spec.style
			if spec.hasPad {
				opt.Padding = spec.pad
			}
			// Matplotlib scales rounding_size/tooth_size by the mutation scale
			// (the font size in pixels), exactly like pad. Convert the spec
			// fractions to pixels so the FancyBboxPatch geometry (driven here with
			// MutationSize=1 and pixel pad) matches the reference.
			if spec.hasRounding {
				opt.RoundingSize = boxStyleFractionToPixels(ctx, spec.roundingSize, fontSize)
			}
			if spec.hasTooth {
				opt.ToothSize = boxStyleFractionToPixels(ctx, spec.toothSize, fontSize)
			}
		}
	}
	if opt.FaceColor == (render.Color{}) {
		opt.FaceColor = render.Color{R: 1, G: 1, B: 1, A: 1}
	} else {
		opt.FaceColor = resolvedTextBBoxColor(opt.FaceColor)
	}
	if opt.EdgeColor == (render.Color{}) {
		opt.EdgeColor = render.Color{R: 0, G: 0, B: 0, A: 1}
	} else {
		opt.EdgeColor = resolvedTextBBoxColor(opt.EdgeColor)
	}
	if opt.LineWidth <= 0 {
		// Stored in points (matplotlib bbox patch linewidth default 1 pt);
		// converted to device pixels at the Paint sinks.
		opt.LineWidth = 1
	}
	if opt.Padding <= 0 {
		opt.Padding = 4
		if ctx != nil {
			opt.Padding = pointsToPixels(ctx.RC, 0.4*fontSize)
		}
	} else if opt.Padding < 1 && ctx != nil {
		opt.Padding = pointsToPixels(ctx.RC, opt.Padding*fontSize)
	}
	return opt
}

// boxStyleFractionToPixels converts a Matplotlib boxstyle fraction (of the font
// size) into display pixels, mirroring the fractional-padding scaling above.
func boxStyleFractionToPixels(ctx *DrawContext, frac, fontSize float64) float64 {
	if ctx == nil {
		return frac
	}
	return pointsToPixels(ctx.RC, frac*fontSize)
}

func resolvedTextBBoxColor(c render.Color) render.Color {
	if c.A == 0 && (c.R != 0 || c.G != 0 || c.B != 0) {
		c.A = 1
	}
	return c
}

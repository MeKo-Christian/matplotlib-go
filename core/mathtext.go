package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	mt "github.com/cwbudde/matplotlib-go/internal/mathtext"
	"github.com/cwbudde/matplotlib-go/render"
)

// MathTextLayoutRun is one text draw in a laid-out MathText expression.
type MathTextLayoutRun = mt.MathTextLayoutRun

// MathTextLayoutRule is a filled rule, such as a fraction bar or root vinculum.
type MathTextLayoutRule = mt.MathTextLayoutRule

// MathTextLayout is a lightweight layout tree flattened into draw runs and
// rules. Offsets and rectangles are relative to the expression baseline.
type MathTextLayout = mt.MathTextLayout

type mathTextMeasurer struct {
	r render.Renderer
}

func (m mathTextMeasurer) MeasureText(text string, size float64, fontKey string) mt.Metrics {
	metrics := m.r.MeasureText(text, size, fontKey)
	ascent := metrics.Ascent
	descent := metrics.Descent
	boundsY := 0.0
	boundsH := 0.0
	if bounder, ok := m.r.(render.TextBounder); ok {
		if bounds, ok := bounder.MeasureTextBounds(text, size, fontKey); ok && bounds.W > 0 && bounds.H > 0 {
			ascent = math.Max(0, -bounds.Y)
			descent = math.Max(0, bounds.Y+bounds.H)
			boundsY = bounds.Y
			boundsH = bounds.H
		}
	}
	return mt.Metrics{
		W:       metrics.W,
		H:       ascent + descent,
		Ascent:  ascent,
		Descent: descent,
		BoundsY: boundsY,
		BoundsH: boundsH,
	}
}

// DPI implements mt.DPIMeasurer so the layout can use matplotlib's exact
// fontsize*dpi/72 thickness. Returns 0 when the renderer doesn't expose DPI.
func (m mathTextMeasurer) DPI() float64 {
	if p, ok := m.r.(render.DPIProvider); ok {
		return float64(p.Resolution())
	}
	return 0
}

// GlyphRun implements mt.GlyphMeasurer by delegating to the renderer's
// render.MathGlyphMeasurer capability (matplotlib `_get_info`). Returns false
// when the renderer lacks the pixel-exact path (purego/WASM), so the layout
// falls back to whole-run MeasureText.
func (m mathTextMeasurer) GlyphRun(text string, size float64, fontKey string) ([]mt.GlyphInfo, bool) {
	measurer, ok := m.r.(render.MathGlyphMeasurer)
	if !ok {
		return nil, false
	}
	metrics, ok := measurer.MeasureMathGlyphRun(text, size, fontKey)
	if !ok || len(metrics) == 0 {
		return nil, false
	}
	out := make([]mt.GlyphInfo, len(metrics))
	for i, gm := range metrics {
		out[i] = mt.GlyphInfo{
			Advance:    gm.Advance,
			Iceberg:    gm.Iceberg,
			Height:     gm.Height,
			Xmin:       gm.Xmin,
			Xmax:       gm.Xmax,
			Ymin:       gm.Ymin,
			Ymax:       gm.Ymax,
			KernToPrev: gm.KernToPrev,
		}
	}
	return out, true
}

// mathLayoutImageMetrics computes matplotlib's RasterParse (get_text_width_height_descent)
// metrics for a laid-out MathText expression: the ink-image bbox (xmax-xmin,
// ymax-ymin) with the -1/+1 border, exactly as _mathtext.Output.to_raster. The
// Agg backend ALIGNS mathtext by these image metrics — NOT the advance/box
// width — so a centered fraction (whose box.width includes a trailing Hbox(2t)
// that the ink omits) would otherwise land ~1px off. Returns the image width and
// the image ascent/descent: w = xmax-xmin, ascent = (ymax-ymin) - box.depth,
// descent = (ymax-ymin) - box.height (so DrawMathTextImage's parseDescent =
// totalH - boxAscent equals this descent — the two stay consistent). ok is false
// when the renderer lacks the pixel-exact GlyphRun path (purego/vector), where
// matplotlib uses the (box-based) to_vector metrics instead.
func mathLayoutImageMetrics(r render.Renderer, layout MathTextLayout, fontKey string) (w, ascent, descent float64, ok bool) {
	measurer := mathTextMeasurer{r: r}
	xmin, ymin, xmax, ymax := 0.0, 0.0, 0.0, 0.0
	sawGlyph := false
	for _, run := range layout.Runs {
		infos, ok := measurer.GlyphRun(run.Text, run.FontSize, resolveRunFontKey(run, fontKey))
		if !ok || len(infos) == 0 {
			return 0, 0, 0, false
		}
		info := infos[0]
		shipY := layout.Ascent + run.Offset.Y
		xmin = math.Min(xmin, run.Offset.X+info.Xmin)
		xmax = math.Max(xmax, run.Offset.X+info.Xmax)
		ymin = math.Min(ymin, shipY-info.Ymax)
		ymax = math.Max(ymax, shipY-info.Ymin)
		sawGlyph = true
	}
	for _, rule := range layout.Rules {
		xmin = math.Min(xmin, rule.Rect.Min.X)
		xmax = math.Max(xmax, rule.Rect.Max.X)
		ymin = math.Min(ymin, layout.Ascent+rule.Rect.Min.Y)
		ymax = math.Max(ymax, layout.Ascent+rule.Rect.Max.Y)
	}
	if !sawGlyph && len(layout.Rules) == 0 {
		return 0, 0, 0, false
	}
	xmin, ymin, xmax, ymax = xmin-1, ymin-1, xmax+1, ymax+1
	totalH := ymax - ymin
	return xmax - xmin, totalH - layout.Descent, totalH - layout.Ascent, true
}

type mathTextFontResolver struct{}

func (mathTextFontResolver) ResolveMathFontKey(base string, request mt.FontRequest) string {
	props := render.ParseFontProperties(base)
	if len(request.Families) > 0 {
		props.File = ""
		props.Families = append([]string(nil), request.Families...)
	} else if families := mathFontFamilyFallbacks(props.MathFontFamily); len(families) > 0 {
		props.File = ""
		props.Families = families
	}
	if request.Style != "" {
		props.Style = render.FontStyle(request.Style)
	}
	if request.Weight > 0 {
		props.Weight = request.Weight
	}
	if face, ok := render.DefaultFontManager().FindFont(props); ok && face.Path != "" {
		return face.Path
	}
	if len(props.Families) > 0 {
		return strings.Join(props.Families, ", ")
	}
	if props.File != "" {
		return props.File
	}
	return base
}

func mathFontFamilyFallbacks(family string) []string {
	switch strings.ToLower(strings.TrimSpace(family)) {
	case "":
		return nil
	case "dejavusans", "dejavu sans":
		return []string{"DejaVu Sans"}
	case "dejavuserif", "dejavu serif":
		return []string{"DejaVu Serif"}
	case "cm", "computer modern":
		return []string{"cmmi10", "cmr10", "Computer Modern Roman"}
	case "stix":
		return []string{"STIXGeneral", "STIXSizeOneSym", "DejaVu Serif"}
	case "stixsans":
		return []string{"STIXNonUnicode", "DejaVu Sans"}
	case "custom":
		return nil
	default:
		return []string{family}
	}
}

func mathTextOptions() mt.Options {
	return mt.Options{
		FontResolver: mathTextFontResolver{},
		Cache:        mt.DefaultCache(),
	}
}

func normalizeDisplayText(text string) string {
	return mt.NormalizeDisplay(text)
}

func fullMathExpression(text string) (string, bool) {
	return mt.FullExpression(text)
}

func displayTextIsEmpty(text string) bool {
	return mt.DisplayTextIsEmpty(text)
}

func displayTextForMathParsing(text string, parseMath bool) string {
	if parseMath {
		return normalizeDisplayText(text)
	}
	return strings.ReplaceAll(text, `\$`, "$")
}

// LayoutMathText parses and lays out one MathText expression without requiring
// dollar delimiters.
func LayoutMathText(r render.Renderer, expr string, size float64, fontKey string) (MathTextLayout, bool) {
	return mt.LayoutMathText(mathTextMeasurer{r: r}, expr, size, fontKey, mathTextOptions())
}

func layoutDisplayText(r render.Renderer, text string, size float64, fontKey string) (MathTextLayout, bool) {
	return mt.LayoutDisplay(mathTextMeasurer{r: r}, text, size, fontKey, mathTextOptions())
}

func drawDisplayText(textRen render.TextDrawer, text string, origin geom.Pt, size float64, textColor render.Color, fontKey string, useTeX ...bool) {
	drawDisplayTextParseMath(textRen, text, origin, size, textColor, fontKey, true, useTeX...)
}

func drawDisplayTextParseMath(textRen render.TextDrawer, text string, origin geom.Pt, size float64, textColor render.Color, fontKey string, parseMath bool, useTeX ...bool) {
	if texEnabled(useTeX) {
		if texRen, ok := textRen.(render.TeXDrawer); ok && texRen.DrawTeX(text, origin, size, textColor, fontKey) {
			return
		}
	}

	if parseMath {
		if ren, ok := textRen.(render.Renderer); ok {
			if layout, ok := layoutDisplayText(ren, text, size, fontKey); ok {
				drawMathTextLayout(ren, textRen, layout, origin, textColor, fontKey)
				return
			}
		}
	}
	display := displayTextForMathParsing(text, parseMath)
	if display == "" {
		return
	}
	drawTextWithFontContext(textRen, display, origin, size, textColor, fontKey)
}

func drawDisplayTextRotated(textRen render.RotatedTextDrawer, text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string, useTeX ...bool) {
	drawDisplayTextRotatedParseMath(textRen, text, anchor, size, angle, textColor, fontKey, true, useTeX...)
}

func drawDisplayTextRotatedParseMath(textRen render.RotatedTextDrawer, text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string, parseMath bool, useTeX ...bool) {
	if texEnabled(useTeX) {
		if texRen, ok := textRen.(render.RotatedTeXDrawer); ok && texRen.DrawTeXRotated(text, anchor, size, angle, textColor, fontKey) {
			return
		}
	}

	if parseMath {
		if ren, ok := textRen.(render.Renderer); ok {
			if layout, ok := layoutDisplayText(ren, text, size, fontKey); ok {
				if drawMathTextLayoutRotated(ren, layout, anchor, angle, textColor, fontKey) {
					return
				}
			}
		}
	}
	display := displayTextForMathParsing(text, parseMath)
	if display == "" {
		return
	}
	drawTextRotatedWithFontContext(textRen, display, anchor, size, angle, textColor, fontKey)
}

func drawDisplayTextVertical(textRen render.VerticalTextDrawer, text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
	if expr, ok := fullMathExpression(text); ok {
		if ren, ok := textRen.(render.Renderer); ok {
			if layout, ok := LayoutMathText(ren, expr, size, fontKey); ok {
				if drawMathTextLayoutVertical(ren, layout, center, textColor, fontKey) {
					return
				}
			}
		}
	}
	display := normalizeDisplayText(text)
	if display == "" {
		return
	}
	drawTextVerticalWithFontContext(textRen, display, center, size, textColor, fontKey)
}

func drawTextWithFontContext(textRen render.TextDrawer, text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	if fontRen, ok := textRen.(render.FontTextDrawer); ok {
		fontRen.DrawTextWithFont(text, origin, size, textColor, fontKey)
		return
	}
	primeTextFont(textRen, text, size, fontKey)
	textRen.DrawText(text, origin, size, textColor)
}

func drawTextRotatedWithFontContext(textRen render.RotatedTextDrawer, text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	if fontRen, ok := textRen.(render.FontRotatedTextDrawer); ok {
		fontRen.DrawTextRotatedWithFont(text, anchor, size, angle, textColor, fontKey)
		return
	}
	primeTextFont(textRen, text, size, fontKey)
	textRen.DrawTextRotated(text, anchor, size, angle, textColor)
}

func drawTextVerticalWithFontContext(textRen render.VerticalTextDrawer, text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
	if fontRen, ok := textRen.(render.FontVerticalTextDrawer); ok {
		fontRen.DrawTextVerticalWithFont(text, center, size, textColor, fontKey)
		return
	}
	primeTextFont(textRen, text, size, fontKey)
	textRen.DrawTextVertical(text, center, size, textColor)
}

func primeTextFont(textRen render.TextDrawer, sample string, size float64, fontKey string) {
	if fontKey == "" {
		return
	}
	if ren, ok := textRen.(render.Renderer); ok {
		_ = ren.MeasureText(sample, size, fontKey)
	}
}

// mathRunDevicePoint converts a math layout run offset (layout space, y-down:
// negative Y sits above the baseline) into the renderer's y-up device space by
// negating Y. The glyph outline pipeline (render/text_path.go sfntPoint) builds
// ascenders at larger Y, so the layout offsets must be flipped to match.
func mathRunDevicePoint(origin geom.Pt, offset geom.Pt) geom.Pt {
	return geom.Pt{X: origin.X + offset.X, Y: origin.Y - offset.Y}
}

// mathRuleDeviceRect converts a math layout rule rect (layout space, y-down)
// into y-up device space. Negating Y inverts min/max ordering, so the corners
// are swapped to keep the rect well-formed (Min.Y <= Max.Y).
func mathRuleDeviceRect(origin geom.Pt, rule geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: origin.X + rule.Min.X, Y: origin.Y - rule.Max.Y},
		Max: geom.Pt{X: origin.X + rule.Max.X, Y: origin.Y - rule.Min.Y},
	}
}

// rasterizeMathRule ports matplotlib's _mathtext.Output.to_raster rectangle
// rasterization for the Agg backend. Given a device-space rule rect (y-down,
// Min.Y = top), it reproduces matplotlib exactly:
//
//	height = max(int(y2 - y1) - 1, 0)
//	if height == 0: y = int((y1+y2)/2 - 0.5)   # 1px bar centered on the rule
//	else:           y = int(y1)
//	draw_rect_filled(int(x1), y, ceil(x2), y+height)   # inclusive both corners
//
// FT2Image.draw_rect_filled fills [x0..x1] and [y0..y1] inclusive, so the
// returned half-open rect adds +1 on the max corners. This is what places the
// fraction bar 1-2px lower than the old grow+snap path, matching the reference.
func rasterizeMathRule(r geom.Rect) geom.Rect {
	x0 := math.Trunc(r.Min.X)
	x1 := math.Ceil(r.Max.X)
	y1, y2 := r.Min.Y, r.Max.Y
	height := math.Trunc(y2-y1) - 1
	if height < 0 {
		height = 0
	}
	var y float64
	if height == 0 {
		center := (y1 + y2) / 2
		y = math.Trunc(center - 0.5)
	} else {
		y = math.Trunc(y1)
	}
	return geom.Rect{
		Min: geom.Pt{X: x0, Y: y},
		Max: geom.Pt{X: x1 + 1, Y: y + height + 1},
	}
}

func drawMathTextLayout(r render.Renderer, textRen render.TextDrawer, layout MathTextLayout, origin geom.Pt, textColor render.Color, fontKey string) {
	// matplotlib's Agg backend draws mathtext rules as box paths with snapping
	// enabled (gc.set_snap default None -> SNAP_AUTO), growing sub-pixel bars to
	// 1px first; PathSnapper then quantizes each corner to floor(v+0.5). The Go
	// snapPath is a faithful port, so route raster rules through it via SnapAuto.
	// Vector backends (no RGBA export) draw exact, unsnapped rects like
	// matplotlib's vector backends.
	_, rasterBackend := r.(render.RGBAExporter)

	// Pixel-exact path: on a raster backend that implements matplotlib's
	// to_raster glyph+rect blitting (cgo FreeType), draw the whole expression
	// through it. Purego/WASM and vector backends fall through to the per-run
	// subpixel path below.
	if rasterBackend {
		if imgDrawer, ok := textRen.(render.MathTextImageDrawer); ok {
			glyphs := make([]render.MathGlyphPlacement, 0, len(layout.Runs))
			for _, run := range layout.Runs {
				glyphs = append(glyphs, render.MathGlyphPlacement{
					Text:     run.Text,
					FontSize: run.FontSize,
					FontKey:  resolveRunFontKey(run, fontKey),
					Ox:       run.Offset.X,
					Oy:       run.Offset.Y,
				})
			}
			rects := make([]render.MathRectPlacement, 0, len(layout.Rules))
			for _, rule := range layout.Rules {
				rects = append(rects, render.MathRectPlacement{
					X1: rule.Rect.Min.X, Y1: rule.Rect.Min.Y,
					X2: rule.Rect.Max.X, Y2: rule.Rect.Max.Y,
				})
			}
			if imgDrawer.DrawMathTextImage(glyphs, rects, origin, layout.Ascent, layout.Descent, textColor) {
				return
			}
		}
	}

	for _, rule := range layout.Rules {
		rect := mathRuleDeviceRect(origin, rule.Rect)
		paint := render.Paint{Fill: textColor}
		if rasterBackend {
			// Faithful port of matplotlib _mathtext.Output.to_raster's rect
			// rasterization (which then blits via FT2Image.draw_rect_filled).
			// This integer-aligns the bar exactly as matplotlib does, so no
			// PathSnapper pass is needed.
			rect = rasterizeMathRule(rect)
		}
		r.Path(pixelRectPath(rect), &paint)
	}
	for _, run := range layout.Runs {
		runFontKey := resolveRunFontKey(run, fontKey)
		drawTextWithFontContext(textRen, run.Text, mathRunDevicePoint(origin, run.Offset), run.FontSize, textColor, runFontKey)
	}
}

func drawMathTextLayoutRotated(r render.Renderer, layout MathTextLayout, anchor geom.Pt, angle float64, textColor render.Color, fontKey string) bool {
	if math.IsNaN(angle) || math.IsInf(angle, 0) {
		return false
	}
	origin := geom.Pt{
		X: anchor.X - layout.Width/2,
		Y: anchor.Y - layout.Descent,
	}
	return drawMathTextLayoutPathTransformed(r, layout, origin, anchor, angle, textColor, fontKey)
}

func drawMathTextLayoutVertical(r render.Renderer, layout MathTextLayout, center geom.Pt, textColor render.Color, fontKey string) bool {
	origin := alignedSingleLineOrigin(center, singleLineTextLayout{
		TextLineLayout: render.TextLineLayout{
			Width:   layout.Width,
			Ascent:  layout.Ascent,
			Descent: layout.Descent,
			Height:  layout.Height,
		},
	}, TextAlignCenter, textLayoutVAlignCenter)
	return drawMathTextLayoutPathTransformed(r, layout, origin, center, math.Pi/2, textColor, fontKey)
}

func drawMathTextLayoutPathTransformed(r render.Renderer, layout MathTextLayout, origin geom.Pt, pivot geom.Pt, angle float64, textColor render.Color, fontKey string) bool {
	paths, ok := mathTextLayoutPaths(r, layout, origin, fontKey)
	if !ok {
		return false
	}
	if angle == 0 {
		for _, path := range paths {
			r.Path(path, &render.Paint{Fill: textColor})
		}
		return true
	}

	cos := math.Cos(angle)
	sin := math.Sin(angle)
	affine := translateAffine(pivot).
		Mul(geom.Affine{A: cos, B: sin, C: -sin, D: cos}).
		Mul(translateAffine(geom.Pt{X: -pivot.X, Y: -pivot.Y}))
	for _, path := range paths {
		r.Path(applyAffinePath(path, affine), &render.Paint{Fill: textColor})
	}
	return true
}

func mathTextLayoutPaths(r render.Renderer, layout MathTextLayout, origin geom.Pt, fontKey string) ([]geom.Path, bool) {
	paths := make([]geom.Path, 0, len(layout.Rules)+len(layout.Runs))
	for _, rule := range layout.Rules {
		rect := mathRuleDeviceRect(origin, rule.Rect)
		paths = append(paths, pixelRectPath(rect))
	}
	for _, run := range layout.Runs {
		runFontKey := resolveRunFontKey(run, fontKey)
		runPath, ok := mathTextRunPath(r, run.Text, mathRunDevicePoint(origin, run.Offset), run.FontSize, runFontKey)
		if !ok {
			return nil, false
		}
		paths = append(paths, runPath)
	}
	return paths, true
}

func mathTextRunPath(r render.Renderer, text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if pather, ok := r.(render.TextPather); ok {
		if path, ok := pather.TextPath(text, origin, size, fontKey); ok {
			return path, true
		}
	}
	return render.TextPath(text, origin, size, fontKey)
}

func resolveRunFontKey(run MathTextLayoutRun, fallback string) string {
	if strings.TrimSpace(run.FontKey) != "" {
		return run.FontKey
	}
	return fallback
}

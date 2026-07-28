package core

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	mt "github.com/cwbudde/mathtext"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
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
	if infos, ok := m.GlyphRun(text, size, fontKey); ok {
		advance := 0.0
		for _, info := range infos {
			advance += info.KernToPrev + info.Advance
		}
		if advance > 0 {
			metrics.W = advance
		}
	}
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
	applyMatplotlibDecimalKern(out, []rune(text), size, m.DPI())
	return out, true
}

func applyMatplotlibDecimalKern(infos []mt.GlyphInfo, runes []rune, size, dpi float64) {
	if len(infos) != len(runes) || len(infos) < 3 {
		return
	}
	if dpi <= 0 {
		dpi = 100
	}
	// Matplotlib's math parser compacts decimal literals: in "1.2", the digit
	// after the decimal point is shifted left by 0.075 pt at 72 dpi, scaled to
	// the expression DPI. This matches MathTextParser("path") glyph positions
	// for numeric mantissas used by ScalarFormatter.
	adjust := math.Round(size*dpi/100.0*0.075*4.0) / 4.0
	if adjust == 0 {
		return
	}
	for i := 2; i < len(runes); i++ {
		if runes[i-1] == '.' && unicode.IsDigit(runes[i-2]) && unicode.IsDigit(runes[i]) {
			infos[i].KernToPrev -= adjust
		}
	}
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

type mathTextFontResolver struct {
	rc style.MathtextRC
}

func (r mathTextFontResolver) ResolveMathFontKey(base string, request mt.FontRequest) string {
	props := render.ParseFontProperties(base)
	if strings.TrimSpace(props.MathFontFamily) == "" {
		// Default the math font set from rcParams["mathtext.fontset"] when the
		// font key does not already encode one (matplotlib's mathtext.fontset).
		props.MathFontFamily = r.rc.Fontset
	}
	if len(request.Families) > 0 {
		props.File = ""
		props.Families = mathFontRequestFamilies(props.MathFontFamily, request.Families)
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

// mathTextProfileResolver adapts mathtext's semantic fontset profile to the
// renderer's concrete font keys. Besides per-glyph resolution it deliberately
// forwards Constants and SizedAlternatives: the mathtext layout engine detects
// those optional capabilities on the GlyphResolver itself.
type mathTextProfileResolver struct {
	profile *mt.MathFontProfile
}

func (r mathTextProfileResolver) ResolveMathGlyph(base string, request mt.GlyphRequest) mt.GlyphResolution {
	resolved := r.profile.ResolveMathGlyph(base, request)
	resolved.FontKey = resolveMathTextFontPattern(resolved.FontKey)
	return resolved
}

func (r mathTextProfileResolver) ResolveMathGlyphCandidates(
	base string,
	request mt.GlyphRequest,
) []mt.GlyphResolution {
	resolved := r.profile.ResolveMathGlyphCandidates(base, request)
	for i := range resolved {
		resolved[i].FontKey = resolveMathTextFontPattern(resolved[i].FontKey)
	}
	return resolved
}

func (r mathTextProfileResolver) Constants() mt.MathFontConstants {
	return r.profile.Constants()
}

func (r mathTextProfileResolver) SizedAlternatives(symbol string) []mt.GlyphResolution {
	resolved := r.profile.SizedAlternatives(symbol)
	for i := range resolved {
		resolved[i].FontKey = resolveMathTextFontPattern(resolved[i].FontKey)
	}
	return resolved
}

func resolveMathTextFontPattern(pattern string) string {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return ""
	}
	if strings.HasPrefix(pattern, "fontprops:") {
		props := render.ParseFontProperties(pattern)
		if face, ok := render.DefaultFontManager().FindFont(props); ok && face.Path != "" {
			return filepath.Clean(face.Path)
		}
		return pattern
	}
	props := parseMathTextFontPattern(pattern)
	if face, ok := render.DefaultFontManager().FindFont(props); ok && face.Path != "" {
		return filepath.Clean(face.Path)
	}
	if key := render.FontPropertiesKey(props); key != "" {
		return key
	}
	return pattern
}

func parseMathTextFontPattern(pattern string) render.FontProperties {
	if props := render.ParseFontProperties(pattern); props.File != "" {
		return props
	}
	parts := strings.Split(pattern, ":")
	props := render.ParseFontProperties(strings.TrimSpace(parts[0]))
	for _, raw := range parts[1:] {
		part := strings.ToLower(strings.TrimSpace(raw))
		key, value, hasValue := strings.Cut(part, "=")
		if hasValue {
			part = strings.TrimSpace(value)
		}
		switch key {
		case "style", "slant":
			switch part {
			case "italic":
				props.Style = render.FontStyleItalic
			case "oblique":
				props.Style = render.FontStyleOblique
			case "normal", "roman":
				props.Style = render.FontStyleNormal
			}
		case "weight":
			props.Weight = parseMathTextFontWeight(part, props.Weight)
		default:
			switch part {
			case "italic":
				props.Style = render.FontStyleItalic
			case "oblique":
				props.Style = render.FontStyleOblique
			case "bold":
				props.Weight = 700
			case "normal", "regular":
				// A bare "normal" is both normal posture and regular weight.
				props.Style = render.FontStyleNormal
				props.Weight = 400
			}
		}
	}
	return props
}

func parseMathTextFontWeight(value string, fallback int) int {
	switch value {
	case "bold":
		return 700
	case "normal", "regular":
		return 400
	}
	if weight, err := strconv.Atoi(value); err == nil && weight > 0 {
		return weight
	}
	return fallback
}

func mathFontRequestFamilies(mathFontFamily string, requested []string) []string {
	families := append([]string(nil), requested...)
	if len(families) == 0 || !isDejaVuSerifRomanRequest(families) {
		return families
	}
	switch strings.ToLower(strings.TrimSpace(mathFontFamily)) {
	case "", "dejavusans", "dejavu sans":
		return []string{"DejaVu Sans", "sans-serif"}
	case "dejavuserif", "dejavu serif":
		return []string{"DejaVu Serif", "serif"}
	case "cm", "computer modern":
		return []string{"cmr10", "Computer Modern Roman"}
	case "stix", "stixsans":
		return []string{"STIXGeneral"}
	default:
		return families
	}
}

func isDejaVuSerifRomanRequest(families []string) bool {
	first := strings.ToLower(strings.TrimSpace(families[0]))
	return first == "dejavu serif" || first == "dejavuserif"
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

func mathTextOptions(r render.Renderer, fontKey, text string) mt.Options {
	rc := style.CurrentDefaults().Mathtext
	useProfile := rc != style.Default.Mathtext || defaultMathTextNeedsGlyphProfile(text)
	if explicit := strings.TrimSpace(render.ParseFontProperties(fontKey).MathFontFamily); explicit != "" {
		switch mt.MathFontSet(explicit) {
		case mt.MathFontSetDejaVuSans, mt.MathFontSetDejaVuSerif, mt.MathFontSetCM,
			mt.MathFontSetSTIX, mt.MathFontSetSTIXSans, mt.MathFontSetCustom:
			useProfile = useProfile || rc.Fontset != explicit
			rc.Fontset = explicit
		}
	}
	opts := mt.Options{
		FontResolver:   mathTextFontResolver{rc: rc},
		Cache:          mt.DefaultCache(),
		MeasurementKey: mathTextRendererMeasurementKey(r, rc),
	}
	// The original resolver is pixel-exact for Matplotlib's unchanged default
	// DejaVu Sans profile. The semantic profile is required only once an rc
	// setting (or an explicit math-font family) opts into alternate font maps.
	if !useProfile {
		return opts
	}
	profile := mt.NewMathFontProfile(mt.MathFontSet(rc.Fontset))
	profile.Fallback = mt.MathFontSet(rc.Fallback)
	profile.Fonts = map[mt.FontClass]string{
		mt.FontClassBold:         rc.BF,
		mt.FontClassBoldItalic:   rc.BFit,
		mt.FontClassCalligraphic: rc.Cal,
		mt.FontClassItalic:       rc.It,
		mt.FontClassRoman:        rc.RM,
		mt.FontClassSans:         rc.SF,
		mt.FontClassTypewriter:   rc.TT,
	}
	opts.GlyphResolver = mathTextProfileResolver{profile: profile}
	opts.DefaultFontClass = mt.FontClass(rc.Default)
	return opts
}

// defaultMathTextNeedsGlyphProfile identifies the virtual alphabets that need
// the fontset's glyph substitution table. Plain default DejaVu MathText is
// laid out through FontResolver alone: that long-standing path is pixel-exact
// with Matplotlib's DejaVu Sans output, including fractions and sized radicals.
func defaultMathTextNeedsGlyphProfile(text string) bool {
	for _, command := range [...]string{`\mathbb`, `\mathcal`, `\mathfrak`, `\mathscr`} {
		if strings.Contains(text, command) {
			return true
		}
	}
	return false
}

func mathTextRendererMeasurementKey(r render.Renderer, rc style.MathtextRC) string {
	value := reflect.ValueOf(r)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		// A value renderer has no stable instance identity. Keep parsed-node
		// caching, but leave final-layout caching disabled rather than sharing
		// metrics across potentially different renderer configurations.
		return ""
	}
	dpi := 0.0
	if provider, ok := r.(render.DPIProvider); ok {
		dpi = float64(provider.Resolution())
	}
	return fmt.Sprintf("%T@%x@%g\x1f%s", r, value.Pointer(), dpi, mathTextMeasurementKey(rc))
}

func mathTextMeasurementKey(rc style.MathtextRC) string {
	return fmt.Sprintf("mplgo-mathtext-v1\x1f%s\x1f%s\x1f%s\x1f%s\x1f%s\x1f%s\x1f%s\x1f%s\x1f%s\x1f%s",
		rc.Fontset, rc.Default, rc.Fallback, rc.BF, rc.BFit, rc.Cal, rc.It, rc.RM, rc.SF, rc.TT)
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
	return mt.LayoutMathText(mathTextMeasurer{r: r}, expr, size, fontKey, mathTextOptions(r, fontKey, expr))
}

func layoutDisplayText(r render.Renderer, text string, size float64, fontKey string) (MathTextLayout, bool) {
	return mt.LayoutDisplay(mathTextMeasurer{r: r}, text, size, fontKey, mathTextOptions(r, fontKey, text))
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
			glyphs, rects := mathTextImagePlacements(layout, fontKey)
			if imgDrawer.DrawMathTextImage(glyphs, rects, origin, layout.Ascent, layout.Descent, textColor) {
				return
			}
		}
	}

	for _, rule := range layout.Rules {
		rect := mathRuleDeviceRect(origin, mathTextRectToGeom(rule.Rect))
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
		drawTextWithFontContext(textRen, run.Text, mathRunDevicePoint(origin, mathTextPtToGeom(run.Offset)), run.FontSize, textColor, runFontKey)
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
	if _, rasterBackend := r.(render.RGBAExporter); rasterBackend {
		if imgDrawer, ok := r.(render.RotatedMathTextImageDrawer); ok {
			glyphs, rects := mathTextImagePlacements(layout, fontKey)
			cosT := math.Cos(angle)
			sinT := math.Sin(angle)
			drawOrigin := geom.Pt{
				X: anchor.X - (layout.Width/2*cosT - layout.Descent*sinT),
				Y: anchor.Y - (layout.Width/2*sinT + layout.Descent*cosT),
			}
			if os.Getenv("MATHDBG") != "" {
				fmt.Printf("MATHDBG rot anchor=%v W=%v A=%v D=%v origin=%v angle=%v\n", anchor, layout.Width, layout.Ascent, layout.Descent, drawOrigin, angle)
				for _, g := range glyphs {
					fmt.Printf("MATHDBG glyph %q size=%v ox=%v oy=%v\n", g.Text, g.FontSize, g.Ox, g.Oy)
				}
				for _, rc := range rects {
					fmt.Printf("MATHDBG rect %v %v %v %v\n", rc.X1, rc.Y1, rc.X2, rc.Y2)
				}
			}
			if imgDrawer.DrawMathTextImageRotated(glyphs, rects, drawOrigin, layout.Ascent, layout.Descent, angle, textColor) {
				return true
			}
		}
	}
	return drawMathTextLayoutPathTransformed(r, layout, origin, anchor, angle, textColor, fontKey)
}

func mathTextImagePlacements(layout MathTextLayout, fontKey string) ([]render.MathGlyphPlacement, []render.MathRectPlacement) {
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
	return glyphs, rects
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
		rect := mathRuleDeviceRect(origin, mathTextRectToGeom(rule.Rect))
		paths = append(paths, pixelRectPath(rect))
	}
	for _, run := range layout.Runs {
		runFontKey := resolveRunFontKey(run, fontKey)
		runPath, ok := mathTextRunPath(r, run.Text, mathRunDevicePoint(origin, mathTextPtToGeom(run.Offset)), run.FontSize, runFontKey)
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

func mathTextPtToGeom(pt mt.Pt) geom.Pt {
	return geom.Pt{X: pt.X, Y: pt.Y}
}

func mathTextRectToGeom(rect mt.Rect) geom.Rect {
	return geom.Rect{
		Min: mathTextPtToGeom(rect.Min),
		Max: mathTextPtToGeom(rect.Max),
	}
}

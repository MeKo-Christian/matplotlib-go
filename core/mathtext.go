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
		if expr, ok := fullMathExpression(text); ok {
			if ren, ok := textRen.(render.Renderer); ok {
				if layout, ok := LayoutMathText(ren, expr, size, fontKey); ok {
					if drawMathTextLayoutRotated(ren, layout, anchor, angle, textColor, fontKey) {
						return
					}
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

func drawMathTextLayout(r render.Renderer, textRen render.TextDrawer, layout MathTextLayout, origin geom.Pt, textColor render.Color, fontKey string) {
	for _, rule := range layout.Rules {
		rect := mathRuleDeviceRect(origin, rule.Rect)
		r.Path(pixelRectPath(rect), &render.Paint{Fill: textColor})
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

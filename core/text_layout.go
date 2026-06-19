package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type textLayoutVerticalAlign uint8

const (
	textLayoutVAlignTop textLayoutVerticalAlign = iota
	textLayoutVAlignBottom
	textLayoutVAlignCenter
	textLayoutVAlignBaseline
	textLayoutVAlignCenterBaseline
)

type singleLineTextLayout struct {
	render.TextLineLayout
	MathLayout *MathTextLayout
}

func measureSingleLineTextLayout(r render.Renderer, text string, size float64, fontKey string, useTeX ...bool) singleLineTextLayout {
	return measureSingleLineTextLayoutParseMath(r, text, size, fontKey, true, useTeX...)
}

func measureSingleLineTextLayoutParseMath(r render.Renderer, text string, size float64, fontKey string, parseMath bool, useTeX ...bool) singleLineTextLayout {
	if texEnabled(useTeX) {
		if metricer, ok := r.(render.TeXMetricer); ok {
			if metrics, ok := metricer.MeasureTeX(text, size, fontKey); ok {
				return singleLineTextLayout{
					TextLineLayout: render.TextLineLayout{
						Width:   metrics.W,
						Ascent:  metrics.Ascent,
						Descent: metrics.Descent,
						Height:  metrics.H,
					},
				}
			}
		}
	}

	if parseMath {
		if layout, ok := layoutDisplayText(r, text, size, fontKey); ok {
			width, ascent, descent := layout.Width, layout.Ascent, layout.Descent
			height := layout.Height
			// On the Agg raster backend, matplotlib aligns mathtext by the
			// ink-image bbox (get_text_width_height_descent → to_raster), not the
			// advance box. Override the metrics so centered/right-aligned math
			// anchors to the same pixel as matplotlib. Vector/purego keep the box
			// metrics (matplotlib's to_vector path).
			if _, isRaster := r.(render.RGBAExporter); isRaster {
				if w, a, d, ok := mathLayoutImageMetrics(r, layout, fontKey); ok {
					width, ascent, descent, height = w, a, d, a+d
				}
			}
			lp := r.MeasureText("lp", size, fontKey)
			if lp.H > height {
				height = lp.H
			}
			if lp.Descent > descent {
				descent = lp.Descent
			}
			ascent = height - descent
			if ascent < 0 {
				ascent = 0
			}
			return singleLineTextLayout{
				TextLineLayout: render.TextLineLayout{
					Width:   width,
					Ascent:  ascent,
					Descent: descent,
					Height:  height,
				},
				MathLayout: &layout,
			}
		}
	}

	display := displayTextForMathParsing(text, parseMath)
	return singleLineTextLayout{
		TextLineLayout: render.MeasureTextLineLayout(r, display, size, fontKey),
	}
}

func texEnabled(useTeX []bool) bool {
	return len(useTeX) > 0 && useTeX[0]
}

func textBaselineOffset(layout singleLineTextLayout, align textLayoutVerticalAlign) float64 {
	switch align {
	case textLayoutVAlignTop:
		return -layout.Ascent
	case textLayoutVAlignBottom:
		return layout.Descent
	case textLayoutVAlignCenter:
		return -(layout.Ascent - layout.Descent) / 2
	case textLayoutVAlignCenterBaseline:
		return -layout.Ascent / 2
	default:
		return 0
	}
}

func textHorizontalOriginOffset(layout singleLineTextLayout, align TextAlign) float64 {
	switch align {
	case TextAlignLeft:
		return 0
	case TextAlignRight:
		return layout.Width
	default:
		return layout.Width / 2
	}
}

func alignedSingleLineOrigin(anchor geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign) geom.Pt {
	return geom.Pt{
		X: anchor.X - textHorizontalOriginOffset(layout, hAlign),
		Y: anchor.Y + textBaselineOffset(layout, vAlign),
	}
}

func layoutVerticalAlign(vAlign TextVerticalAlign, preferCenterBaseline bool) textLayoutVerticalAlign {
	switch vAlign {
	case TextVAlignTop:
		return textLayoutVAlignTop
	case TextVAlignBottom:
		return textLayoutVAlignBottom
	case TextVAlignBaseline:
		return textLayoutVAlignBaseline
	case TextVAlignCenterBaseline:
		return textLayoutVAlignCenterBaseline
	case TextVAlignMiddle:
		if preferCenterBaseline {
			return textLayoutVAlignCenterBaseline
		}
		return textLayoutVAlignCenter
	default:
		return textLayoutVAlignBaseline
	}
}

type multilineTextBlockLayout struct {
	Layouts      []singleLineTextLayout
	BaselineYs   []float64
	Rect         geom.Rect
	Width        float64
	Height       float64
	LineAscents  []float64
	LineDescents []float64
}

func measureMultilineTextBlock(r render.Renderer, ctx *DrawContext, anchor geom.Pt, fontSize float64, fontKey string, parseMath, useTeX bool, lines []string, linespacing float64, hAlign TextAlign, vAlign textLayoutVerticalAlign) (multilineTextBlockLayout, bool) {
	if len(lines) == 0 {
		return multilineTextBlockLayout{}, false
	}

	block := multilineTextBlockLayout{
		Layouts:      make([]singleLineTextLayout, len(lines)),
		BaselineYs:   make([]float64, len(lines)),
		LineAscents:  make([]float64, len(lines)),
		LineDescents: make([]float64, len(lines)),
	}
	for i, line := range lines {
		block.Layouts[i] = measureMultilineLineLayout(r, line, fontSize, fontKey, parseMath, useTeX)
		block.Width = math.Max(block.Width, block.Layouts[i].Width)
	}

	lpLayout := measureMultilineLineLayout(r, "lp", fontSize, fontKey, false, useTeX)
	lpHeight, lpDescent := multilineMatplotlibHeightDescent(lpLayout)
	lpAscent := math.Max(0, lpHeight-lpDescent)
	spacing := resolvedTextLinespacing(linespacing)
	baselineOffsets := make([]float64, len(lines))
	thisY := 0.0
	for i, layout := range block.Layouts {
		height, descent := multilineMatplotlibHeightDescent(layout)
		height = math.Max(height, lpHeight)
		descent = math.Max(descent, lpDescent)
		ascent := math.Max(0, height-descent)
		block.LineAscents[i] = ascent
		block.LineDescents[i] = descent

		if i == 0 {
			thisY = -ascent
		} else {
			thisY -= math.Max(lpAscent*spacing, ascent*spacing)
		}
		baselineOffsets[i] = -thisY
		thisY -= descent
	}
	block.Height = -thisY

	left := anchor.X
	switch hAlign {
	case TextAlignCenter:
		left -= block.Width / 2
	case TextAlignRight:
		left -= block.Width
	}

	top := anchor.Y
	switch vAlign {
	case textLayoutVAlignCenter:
		top += block.Height / 2
	case textLayoutVAlignBottom:
		top += block.Height
	case textLayoutVAlignBaseline:
		top += baselineOffsets[len(baselineOffsets)-1]
	case textLayoutVAlignCenterBaseline:
		top += block.LineAscents[0] / 2
	}
	for i, offset := range baselineOffsets {
		block.BaselineYs[i] = top - offset
	}
	block.Rect = geom.Rect{
		Min: geom.Pt{X: left, Y: top - block.Height},
		Max: geom.Pt{X: left + block.Width, Y: top},
	}
	return block, true
}

type rotatedMultilineTextBlockLayout struct {
	Layouts       []singleLineTextLayout
	LineOffsets   []geom.Pt
	TextBoxX      float64
	TextBoxY      float64
	TextBoxWidth  float64
	TextBoxHeight float64
}

func measureRotatedMultilineTextBlock(r render.Renderer, fontSize float64, fontKey string, parseMath, useTeX bool, lines []string, linespacing float64, hAlign TextAlign, vAlign textLayoutVerticalAlign, lineAlign TextAlign, angle float64, mode TextRotationMode) (rotatedMultilineTextBlockLayout, bool) {
	if len(lines) == 0 {
		return rotatedMultilineTextBlockLayout{}, false
	}
	layout := rotatedMultilineTextBlockLayout{
		Layouts:     make([]singleLineTextLayout, len(lines)),
		LineOffsets: make([]geom.Pt, len(lines)),
	}
	widths := make([]float64, len(lines))
	heights := make([]float64, len(lines))
	xs := make([]float64, len(lines))
	ys := make([]float64, len(lines))

	lpLayout := measureMultilineLineLayout(r, "lp", fontSize, fontKey, false, useTeX)
	lpHeight, lpDescent := multilineMatplotlibHeightDescent(lpLayout)
	minDY := (lpHeight - lpDescent) * resolvedTextLinespacing(linespacing)
	spacing := resolvedTextLinespacing(linespacing)

	width := 0.0
	thisY := 0.0
	descent := 0.0
	baseline := 0.0
	for i, line := range lines {
		lineLayout := measureMultilineLineLayout(r, line, fontSize, fontKey, parseMath, useTeX)
		lineHeight, lineDescent := multilineMatplotlibHeightDescent(lineLayout)
		lineHeight = math.Max(lineHeight, lpHeight)
		lineDescent = math.Max(lineDescent, lpDescent)

		layout.Layouts[i] = lineLayout
		widths[i] = lineLayout.Width
		heights[i] = lineHeight
		width = math.Max(width, lineLayout.Width)

		baseline = (lineHeight - lineDescent) - thisY
		if i == 0 {
			thisY = -(lineHeight - lineDescent)
		} else {
			thisY -= math.Max(minDY, (lineHeight-lineDescent)*spacing)
		}
		xs[i] = 0
		ys[i] = thisY
		thisY -= lineDescent
		descent = lineDescent
	}

	xmin, xmax := 0.0, width
	ymax := 0.0
	ymin := ys[len(ys)-1] - descent
	corners := []geom.Pt{
		{X: xmin, Y: ymin},
		{X: xmin, Y: ymax},
		{X: xmax, Y: ymax},
		{X: xmax, Y: ymin},
	}
	minRotX, maxRotX := math.Inf(1), math.Inf(-1)
	minRotY, maxRotY := math.Inf(1), math.Inf(-1)
	for _, corner := range corners {
		rot := rotateTextLayoutPoint(corner, angle)
		minRotX = math.Min(minRotX, rot.X)
		maxRotX = math.Max(maxRotX, rot.X)
		minRotY = math.Min(minRotY, rot.Y)
		maxRotY = math.Max(maxRotY, rot.Y)
	}

	var offsetX, offsetY float64
	if mode == TextRotationModeAnchor {
		offsetX = textLayoutHorizontalOffset(xmin, xmax, hAlign)
		offsetY = textLayoutVerticalOffset(ymin, ymax, ymax-ymin, descent, baseline, vAlign, true)
		rot := rotateTextLayoutPoint(geom.Pt{X: offsetX, Y: offsetY}, angle)
		offsetX, offsetY = rot.X, rot.Y
	} else {
		offsetX = textLayoutHorizontalOffset(minRotX, maxRotX, hAlign)
		offsetY = textLayoutVerticalOffset(minRotY, maxRotY, maxRotY-minRotY, descent, baseline, vAlign, false)
	}

	for i := range lines {
		lineX := xs[i]
		switch lineAlign {
		case TextAlignCenter:
			lineX += width/2 - widths[i]/2
		case TextAlignRight:
			lineX += width - widths[i]
		}
		rot := rotateTextLayoutPoint(geom.Pt{X: lineX, Y: ys[i]}, angle)
		layout.LineOffsets[i] = geom.Pt{X: rot.X - offsetX, Y: rot.Y - offsetY}
	}

	var projectedXs, projectedYs []float64
	for i := range lines {
		linePt := rotateTextLayoutPoint(layout.LineOffsets[i], -angle)
		y1 := linePt.Y - descent
		x2 := linePt.X + widths[i]
		y2 := y1 + heights[i]
		projectedXs = append(projectedXs, linePt.X, x2)
		projectedYs = append(projectedYs, y1, y2)
	}
	xtBox, ytBox := minFloat64(projectedXs), minFloat64(projectedYs)
	layout.TextBoxWidth = maxFloat64(projectedXs) - xtBox
	layout.TextBoxHeight = maxFloat64(projectedYs) - ytBox
	boxOrigin := rotateTextLayoutPoint(geom.Pt{X: xtBox, Y: ytBox}, angle)
	layout.TextBoxX = boxOrigin.X
	layout.TextBoxY = boxOrigin.Y
	return layout, true
}

func textLayoutHorizontalOffset(minX, maxX float64, align TextAlign) float64 {
	switch align {
	case TextAlignCenter:
		return (minX + maxX) / 2
	case TextAlignRight:
		return maxX
	default:
		return minX
	}
}

func textLayoutVerticalOffset(minY, maxY, height, descent, baseline float64, align textLayoutVerticalAlign, anchorMode bool) float64 {
	switch align {
	case textLayoutVAlignCenter:
		return (minY + maxY) / 2
	case textLayoutVAlignTop:
		return maxY
	case textLayoutVAlignBaseline:
		if anchorMode {
			return maxY - baseline
		}
		return minY + descent
	case textLayoutVAlignCenterBaseline:
		if anchorMode {
			return maxY - baseline/2
		}
		return minY + height - baseline/2
	default:
		return minY
	}
}

func rotateTextLayoutPoint(p geom.Pt, angle float64) geom.Pt {
	cosT := math.Cos(angle)
	sinT := math.Sin(angle)
	return geom.Pt{
		X: p.X*cosT - p.Y*sinT,
		Y: p.X*sinT + p.Y*cosT,
	}
}

func rotatedTextBackendAnchorForOrigin(origin geom.Pt, layout singleLineTextLayout, angle float64) geom.Pt {
	cosT := math.Cos(angle)
	sinT := math.Sin(angle)
	return geom.Pt{
		X: origin.X + (layout.Width/2*cosT - layout.Descent*sinT),
		Y: origin.Y + (layout.Width/2*sinT + layout.Descent*cosT),
	}
}

func minFloat64(values []float64) float64 {
	min := math.Inf(1)
	for _, v := range values {
		min = math.Min(min, v)
	}
	return min
}

func maxFloat64(values []float64) float64 {
	max := math.Inf(-1)
	for _, v := range values {
		max = math.Max(max, v)
	}
	return max
}

func measureMultilineLineLayout(r render.Renderer, line string, fontSize float64, fontKey string, parseMath, useTeX bool) singleLineTextLayout {
	if line != "" {
		return measureSingleLineTextLayoutParseMath(r, line, fontSize, fontKey, parseMath, useTeX)
	}
	layout := measureSingleLineTextLayoutParseMath(r, "lp", fontSize, fontKey, false, useTeX)
	layout.Width = 0
	return layout
}

func multilineMatplotlibHeightDescent(layout singleLineTextLayout) (float64, float64) {
	height := layout.Height
	descent := layout.Descent
	if height <= 0 {
		height = layout.Ascent + layout.Descent
	}
	if height <= 0 {
		height = layout.RunAscent + layout.RunDescent
		descent = layout.RunDescent
	}
	if height <= 0 {
		return 0, 0
	}
	if descent < 0 {
		descent = 0
	}
	if descent > height {
		descent = height
	}
	return height, descent
}

func textRotationLayoutAlignments(hAlign TextAlign, vAlign TextVerticalAlign, angleDeg float64, mode TextRotationMode) (TextAlign, textLayoutVerticalAlign) {
	layoutVAlign := layoutVerticalAlign(vAlign, false)
	switch mode {
	case TextRotationModeXTick:
		return xTickRotationHAlign(angleDeg, vAlign), layoutVAlign
	case TextRotationModeYTick:
		return hAlign, yTickRotationVAlign(angleDeg, hAlign)
	default:
		return hAlign, layoutVAlign
	}
}

func xTickRotationHAlign(angleDeg float64, vAlign TextVerticalAlign) TextAlign {
	angle := normalizedTextRotationAngle(angleDeg)
	anchorAtBottom := vAlign == TextVAlignBottom
	if angle <= 10 || (85 <= angle && angle <= 95) || 350 <= angle || (170 <= angle && angle <= 190) || (265 <= angle && angle <= 275) {
		return TextAlignCenter
	}
	if (10 < angle && angle < 85) || (190 < angle && angle < 265) {
		if anchorAtBottom {
			return TextAlignLeft
		}
		return TextAlignRight
	}
	if anchorAtBottom {
		return TextAlignRight
	}
	return TextAlignLeft
}

func yTickRotationVAlign(angleDeg float64, hAlign TextAlign) textLayoutVerticalAlign {
	angle := normalizedTextRotationAngle(angleDeg)
	anchorAtLeft := hAlign == TextAlignLeft
	if angle <= 10 || 350 <= angle || (170 <= angle && angle <= 190) || (80 <= angle && angle <= 100) || (260 <= angle && angle <= 280) {
		return textLayoutVAlignCenter
	}
	if (190 < angle && angle < 260) || (10 < angle && angle < 80) {
		if anchorAtLeft {
			return textLayoutVAlignBaseline
		}
		return textLayoutVAlignTop
	}
	if anchorAtLeft {
		return textLayoutVAlignTop
	}
	return textLayoutVAlignBaseline
}

func normalizedTextRotationAngle(angleDeg float64) float64 {
	angle := math.Mod(angleDeg, 360)
	if angle < 0 {
		angle += 360
	}
	return angle
}

func textAutoWrapWidth(ctx *DrawContext, anchor geom.Pt, hAlign TextAlign, angleDeg float64) float64 {
	if ctx == nil {
		return 0
	}
	figureBox := ctx.FigureRect
	if figureBox.W() <= 0 || figureBox.H() <= 0 {
		figureBox = ctx.Clip
	}
	if figureBox.W() <= 0 || figureBox.H() <= 0 {
		return 0
	}
	angle := normalizedTextRotationAngle(angleDeg)
	left := textDistanceToBox(angle, anchor, figureBox)
	right := textDistanceToBox(math.Mod(180+angle, 360), anchor, figureBox)
	switch hAlign {
	case TextAlignLeft:
		return left
	case TextAlignRight:
		return right
	default:
		return 2 * math.Min(left, right)
	}
}

func textDistanceToBox(rotation float64, anchor geom.Pt, box geom.Rect) float64 {
	const epsilon = 1e-12
	cosDeg := func(deg float64) float64 {
		v := math.Cos(deg * math.Pi / 180)
		if math.Abs(v) < epsilon {
			if v < 0 {
				return -epsilon
			}
			return epsilon
		}
		return v
	}
	var h1, h2 float64
	switch {
	case rotation > 270:
		quad := rotation - 270
		h1 = (anchor.Y - box.Min.Y) / cosDeg(quad)
		h2 = (box.Max.X - anchor.X) / cosDeg(90-quad)
	case rotation > 180:
		quad := rotation - 180
		h1 = (anchor.X - box.Min.X) / cosDeg(quad)
		h2 = (anchor.Y - box.Min.Y) / cosDeg(90-quad)
	case rotation > 90:
		quad := rotation - 90
		h1 = (box.Max.Y - anchor.Y) / cosDeg(quad)
		h2 = (anchor.X - box.Min.X) / cosDeg(90-quad)
	default:
		h1 = (box.Max.X - anchor.X) / cosDeg(rotation)
		h2 = (box.Max.Y - anchor.Y) / cosDeg(90-rotation)
	}
	return math.Min(h1, h2)
}

func textRotationAnchor(origin geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign, angle float64, mode TextRotationMode) geom.Pt {
	// Recover matplotlib's text anchor point P from the unrotated draw origin, then
	// map matplotlib's baseline-left draw origin to the backend rotation pivot.
	p := geom.Pt{
		X: origin.X + textHorizontalOriginOffset(layout, hAlign),
		Y: origin.Y - textBaselineOffset(layout, vAlign),
	}
	return rotatedTextBackendAnchorFromP(p, layout, hAlign, vAlign, angle, mode == TextRotationModeAnchor)
}

func wrappedTextLines(r render.Renderer, text string, fontSize float64, fontKey string, parseMath, useTeX bool, maxWidth float64) []string {
	paragraphs := strings.Split(text, "\n")
	if maxWidth <= 0 {
		return paragraphs
	}
	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		words := strings.Split(paragraph, " ")
		for len(words) > 0 {
			if len(words) == 1 {
				lines = append(lines, words[0])
				break
			}
			for i := 2; i <= len(words); i++ {
				candidate := strings.Join(words[:i], " ")
				width := math.Ceil(measureSingleLineTextLayoutParseMath(r, candidate, fontSize, fontKey, parseMath, useTeX).Width)
				if width > maxWidth {
					lines = append(lines, strings.Join(words[:i-1], " "))
					words = words[i-1:]
					break
				}
				if i == len(words) {
					lines = append(lines, candidate)
					words = nil
					break
				}
			}
		}
	}
	return lines
}

func alignedTextOrigin(anchor geom.Pt, metrics render.TextMetrics, hAlign TextAlign, vAlign TextVerticalAlign) geom.Pt {
	origin := geom.Pt{X: anchor.X, Y: anchor.Y}

	switch hAlign {
	case TextAlignCenter:
		origin.X -= metrics.W / 2
	case TextAlignRight:
		origin.X -= metrics.W
	}

	switch vAlign {
	case TextVAlignTop:
		origin.Y -= metrics.Ascent
	case TextVAlignMiddle:
		origin.Y -= (metrics.Ascent - metrics.Descent) / 2
	case TextVAlignBottom:
		origin.Y += metrics.Descent
	case TextVAlignCenterBaseline:
		origin.Y -= metrics.Ascent / 2
	}

	return origin
}

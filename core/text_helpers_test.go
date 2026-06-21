package core

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func matplotlibMultilineOffsetsForTest(r render.Renderer, lines []string, fontSize float64, fontKey string, linespacing float64, hAlign TextAlign, vAlign TextVerticalAlign, angle float64) []geom.Pt {
	widths := make([]float64, len(lines))
	xs := make([]float64, len(lines))
	ys := make([]float64, len(lines))
	lp := measureSingleLineTextLayoutParseMath(r, "lp", fontSize, fontKey, false, false)
	lpH, lpD := multilineMatplotlibHeightDescent(lp)
	minDY := (lpH - lpD) * resolvedTextLinespacing(linespacing)
	thisY := 0.0
	descent := 0.0
	baseline := 0.0
	width := 0.0
	for i, line := range lines {
		layout := measureMultilineLineLayout(r, line, fontSize, fontKey, true, false)
		w := layout.Width
		h, d := multilineMatplotlibHeightDescent(layout)
		h = math.Max(h, lpH)
		d = math.Max(d, lpD)
		widths[i] = w
		width = math.Max(width, w)
		baseline = (h - d) - thisY
		if i == 0 {
			thisY = -(h - d)
		} else {
			thisY -= math.Max(minDY, (h-d)*resolvedTextLinespacing(linespacing))
		}
		xs[i] = 0
		ys[i] = thisY
		thisY -= d
		descent = d
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
		rot := rotatePointForTest(corner, angle)
		minRotX = math.Min(minRotX, rot.X)
		maxRotX = math.Max(maxRotX, rot.X)
		minRotY = math.Min(minRotY, rot.Y)
		maxRotY = math.Max(maxRotY, rot.Y)
	}
	offsetX := minRotX
	switch hAlign {
	case TextAlignCenter:
		offsetX = (minRotX + maxRotX) / 2
	case TextAlignRight:
		offsetX = maxRotX
	}
	offsetY := minRotY
	switch vAlign {
	case TextVAlignMiddle:
		offsetY = (minRotY + maxRotY) / 2
	case TextVAlignTop:
		offsetY = maxRotY
	case TextVAlignBaseline:
		offsetY = minRotY + descent
	case TextVAlignCenterBaseline:
		offsetY = minRotY + (maxRotY - minRotY) - baseline/2
	}
	out := make([]geom.Pt, len(lines))
	for i := range lines {
		lineX := xs[i]
		switch hAlign {
		case TextAlignCenter:
			lineX += width/2 - widths[i]/2
		case TextAlignRight:
			lineX += width - widths[i]
		}
		rot := rotatePointForTest(geom.Pt{X: lineX, Y: ys[i]}, angle)
		out[i] = geom.Pt{X: rot.X - offsetX, Y: rot.Y - offsetY}
	}
	return out
}

func rotatedTextBaselineOriginForTest(anchor geom.Pt, layout singleLineTextLayout, angle float64) geom.Pt {
	cosT := math.Cos(angle)
	sinT := math.Sin(angle)
	return geom.Pt{
		X: anchor.X - (layout.Width/2*cosT - layout.Descent*sinT),
		Y: anchor.Y - (layout.Width/2*sinT + layout.Descent*cosT),
	}
}

func rotatePointForTest(p geom.Pt, angle float64) geom.Pt {
	cosT := math.Cos(angle)
	sinT := math.Sin(angle)
	return geom.Pt{X: p.X*cosT - p.Y*sinT, Y: p.X*sinT + p.Y*cosT}
}

func matplotlibRotatedTextBBoxPathForTest(anchor, drawOrigin geom.Pt, layout singleLineTextLayout, opt *TextBBoxOptions, ctx *DrawContext, fontSize, angleDeg float64) (geom.Path, bool) {
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

	local := pixelRectPath(geom.Rect{
		Min: geom.Pt{X: -cfg.Padding, Y: -cfg.Padding},
		Max: geom.Pt{X: wBox + cfg.Padding, Y: hBox + cfg.Padding},
	})
	for i := range local.V {
		x, y := rotate(local.V[i].X, local.V[i].Y, angle)
		local.V[i] = geom.Pt{X: anchor.X + xBox + x, Y: anchor.Y + yBox + y}
	}
	return local, true
}

type textRecordingRenderer struct {
	render.NullRenderer
	pathCount  int
	pathPaints []render.Paint
	pathCalls  []recordedPathCall
	texts      []string
	textColors []render.Color
	textSizes  []float64
	origins    []geom.Pt
	imageDsts  []geom.Rect
}

func (r *textRecordingRenderer) Path(p geom.Path, paint *render.Paint) {
	r.pathCount++
	call := recordedPathCall{path: p}
	if paint != nil {
		call.paint = *paint
		r.pathPaints = append(r.pathPaints, call.paint)
	}
	r.pathCalls = append(r.pathCalls, call)
}

func (r *textRecordingRenderer) Image(_ render.Image, dst geom.Rect) {
	r.imageDsts = append(r.imageDsts, dst)
}

func (r *textRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

type mathInkBoundsRenderer struct {
	textRecordingRenderer
}

type mathRasterMetricRenderer struct {
	render.NullRenderer
}

type mathRasterLogitWidthRenderer struct {
	render.NullRenderer
}

type inlineMathLineMetricRenderer struct {
	render.NullRenderer
}

func (inlineMathLineMetricRenderer) GetImage() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, 1, 1))
}

func (inlineMathLineMetricRenderer) MeasureText(text string, _ float64, _ string) render.TextMetrics {
	switch text {
	case "lp":
		return render.TextMetrics{W: 10, H: 14, Ascent: 11, Descent: 3}
	case "time ":
		return render.TextMetrics{W: 30, H: 15, Ascent: 13, Descent: 2}
	default:
		return render.TextMetrics{W: 5, H: 15, Ascent: 13, Descent: 2}
	}
}

func (inlineMathLineMetricRenderer) MeasureMathGlyphRun(text string, _ float64, _ string) ([]render.MathGlyphMetric, bool) {
	if text != "t" {
		return nil, false
	}
	return []render.MathGlyphMetric{{
		Advance: 5,
		Iceberg: 13,
		Height:  15,
		Xmin:    0,
		Xmax:    5,
		Ymin:    -2,
		Ymax:    13,
	}}, true
}

func (mathRasterMetricRenderer) GetImage() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, 1, 1))
}

func (mathRasterMetricRenderer) MeasureMathGlyphRun(text string, size float64, _ string) ([]render.MathGlyphMetric, bool) {
	metric := render.MathGlyphMetric{}
	switch text {
	case "√":
		if math.Abs(size-12.6433) < 0.01 {
			metric = render.MathGlyphMetric{Xmin: 1.9688, Xmax: 19.0625, Ymin: -5.3125, Ymax: 27.9375}
		} else {
			metric = render.MathGlyphMetric{Xmin: 2.0938, Xmax: 20.2812, Ymin: -5.6094, Ymax: 29.4844}
		}
	case "x":
		metric = render.MathGlyphMetric{Xmin: -0.8281, Xmax: 19.2031, Ymin: 0, Ymax: 18}
	case "+":
		metric = render.MathGlyphMetric{Xmin: 3.3750, Xmax: 23.3750, Ymin: 0, Ymax: 20}
	case "1":
		metric = render.MathGlyphMetric{Xmin: 3.5625, Xmax: 17.5000, Ymin: 0, Ymax: 23}
	case "3":
		metric = render.MathGlyphMetric{Xmin: 1.1875, Xmax: 8.6250, Ymin: 0, Ymax: 12}
	case "y":
		metric = render.MathGlyphMetric{Xmin: -0.7969, Xmax: 19.2969, Ymin: -7, Ymax: 18}
	default:
		return nil, false
	}
	metric.Iceberg = metric.Ymax
	return []render.MathGlyphMetric{metric}, true
}

func (mathRasterLogitWidthRenderer) GetImage() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, 1, 1))
}

func (mathRasterLogitWidthRenderer) MeasureText(text string, _ float64, _ string) render.TextMetrics {
	if text == "m" {
		return render.TextMetrics{W: 13.5, H: 10, Ascent: 8, Descent: 2}
	}
	return render.TextMetrics{W: 8, H: 10, Ascent: 8, Descent: 2}
}

func (mathRasterLogitWidthRenderer) MeasureMathGlyphRun(text string, size float64, _ string) ([]render.MathGlyphMetric, bool) {
	metric := render.MathGlyphMetric{}
	switch {
	case text == "1" && math.Abs(size-10) < 0.01:
		metric = render.MathGlyphMetric{Advance: 8.82769775390625, Xmin: 1.5625, Xmax: 7.625, Ymin: 0, Ymax: 10}
	case text == "0" && math.Abs(size-10) < 0.01:
		metric = render.MathGlyphMetric{Advance: 8.82769775390625, Xmin: 0.875, Xmax: 7.8125, Ymin: 0, Ymax: 10}
	case text == "−" && math.Abs(size-10) < 0.01:
		metric = render.MathGlyphMetric{Advance: 11.625732421875, Xmin: 1.46875, Xmax: 10.15625, Ymin: 4, Ymax: 6}
	case text == "−" && math.Abs(size-7) < 0.01:
		metric = render.MathGlyphMetric{Advance: 8.16943359375, Xmin: 1.015625, Xmax: 7.125, Ymin: 2, Ymax: 4}
	case text == "1" && math.Abs(size-7) < 0.01:
		metric = render.MathGlyphMetric{Advance: 6.2032470703125, Xmin: 1.046875, Xmax: 5.25, Ymin: 0, Ymax: 7}
	case text == "m" && math.Abs(size-10) < 0.01:
		metric = render.MathGlyphMetric{Advance: 13.51593017578125, Xmin: 0.484375, Xmax: 12.5625, Ymin: 0, Ymax: 8}
	case text == "x" && math.Abs(size-10) < 0.01:
		metric = render.MathGlyphMetric{Advance: 8.211181640625, Xmin: 0.40625, Xmax: 7.765625, Ymin: 0, Ymax: 8}
	default:
		return nil, false
	}
	metric.Iceberg = metric.Ymax
	return []render.MathGlyphMetric{metric}, true
}

func (r *mathInkBoundsRenderer) MeasureTextBounds(text string, size float64, _ string) (render.TextBounds, bool) {
	if text == "" || size <= 0 {
		return render.TextBounds{}, false
	}
	return render.TextBounds{
		X: 0,
		Y: -size * 0.55,
		W: float64(len(text)) * size * 0.5,
		H: size * 0.70,
	}, true
}

type fontMetricTextRecordingRenderer struct {
	textRecordingRenderer
	fontHeights render.FontHeightMetrics
}

func (r *fontMetricTextRecordingRenderer) MeasureFontHeights(float64, string) (render.FontHeightMetrics, bool) {
	return r.fontHeights, true
}

func (r *textRecordingRenderer) DrawText(text string, origin geom.Pt, size float64, col render.Color) {
	r.texts = append(r.texts, text)
	r.textColors = append(r.textColors, col)
	r.textSizes = append(r.textSizes, size)
	r.origins = append(r.origins, origin)
}

func hasPathPaint(paints []render.Paint, fill, stroke render.Color, lineWidth float64) bool {
	for _, paint := range paints {
		if paint.Fill == fill && paint.Stroke == stroke && approx(paint.LineWidth, lineWidth, 1e-12) {
			return true
		}
	}
	return false
}

func hasPaintAlpha(paints []render.Paint, alpha float64) bool {
	for _, paint := range paints {
		if approx(paint.Fill.A, alpha, 1e-12) || approx(paint.Stroke.A, alpha, 1e-12) {
			return true
		}
	}
	return false
}

type recordedFontTextCall struct {
	text    string
	anchor  geom.Pt
	fontKey string
}

type fontAwareTextRecordingRenderer struct {
	textRecordingRenderer
	fontTextCalls     []recordedFontTextCall
	fontRotatedCalls  []recordedFontTextCall
	fontVerticalCalls []recordedFontTextCall
	verticalCalls     []string
}

func (r *fontAwareTextRecordingRenderer) DrawTextWithFont(text string, _ geom.Pt, _ float64, _ render.Color, fontKey string) {
	r.fontTextCalls = append(r.fontTextCalls, recordedFontTextCall{text: text, fontKey: fontKey})
}

func (r *fontAwareTextRecordingRenderer) DrawTextRotated(text string, anchor geom.Pt, _ float64, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, anchor)
}

func (r *fontAwareTextRecordingRenderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, _ float64, _ float64, _ render.Color, fontKey string) {
	r.fontRotatedCalls = append(r.fontRotatedCalls, recordedFontTextCall{text: text, anchor: anchor, fontKey: fontKey})
}

func (r *fontAwareTextRecordingRenderer) DrawTextVertical(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.verticalCalls = append(r.verticalCalls, text)
}

func (r *fontAwareTextRecordingRenderer) DrawTextVerticalWithFont(text string, _ geom.Pt, _ float64, _ render.Color, fontKey string) {
	r.fontVerticalCalls = append(r.fontVerticalCalls, recordedFontTextCall{text: text, fontKey: fontKey})
}

type verticalMathTextRecordingRenderer struct {
	textRecordingRenderer
	verticalTexts []string
	textPathCalls []string
}

func (r *verticalMathTextRecordingRenderer) DrawTextVertical(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.verticalTexts = append(r.verticalTexts, text)
}

func (r *verticalMathTextRecordingRenderer) TextPath(text string, origin geom.Pt, _ float64, _ string) (geom.Path, bool) {
	r.textPathCalls = append(r.textPathCalls, text)
	return patchRectPath(geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - 4},
		Max: geom.Pt{X: origin.X + 4, Y: origin.Y},
	}), true
}

type rotatedMathTextRecordingRenderer struct {
	textRecordingRenderer
	rotatedTexts  []string
	textPathCalls []string
}

func (r *rotatedMathTextRecordingRenderer) DrawTextRotated(text string, _ geom.Pt, _ float64, _ float64, _ render.Color) {
	r.rotatedTexts = append(r.rotatedTexts, text)
}

func (r *rotatedMathTextRecordingRenderer) TextPath(text string, origin geom.Pt, _ float64, _ string) (geom.Path, bool) {
	r.textPathCalls = append(r.textPathCalls, text)
	return patchRectPath(geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - 4},
		Max: geom.Pt{X: origin.X + 4, Y: origin.Y},
	}), true
}

type texRecordingRenderer struct {
	textRecordingRenderer
	texDraws        []string
	texRotatedDraws []string
}

func (r *texRecordingRenderer) MeasureTeX(text string, size float64, fontKey string) (render.TextMetrics, bool) {
	return render.TextMetrics{W: 123, H: 22, Ascent: 17, Descent: 5}, true
}

func (r *texRecordingRenderer) DrawTeX(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) bool {
	r.texDraws = append(r.texDraws, text)
	return true
}

func (r *texRecordingRenderer) DrawTeXRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) bool {
	r.texRotatedDraws = append(r.texRotatedDraws, text)
	return true
}

func (r *texRecordingRenderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, anchor)
}

func containsTextString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsFontTextCall(calls []recordedFontTextCall, want string) bool {
	for _, call := range calls {
		if call.text == want {
			return true
		}
	}
	return false
}

func containsPathPointForTextTest(calls []recordedPathCall, want geom.Pt) bool {
	for _, call := range calls {
		if pathHasPointForTextTest(call.path, want) {
			return true
		}
	}
	return false
}

func containsPathPointNearForTextTest(calls []recordedPathCall, want geom.Pt, tol float64) bool {
	for _, call := range calls {
		if pathHasPointNearForTextTest(call.path, want, tol) {
			return true
		}
	}
	return false
}

func pathHasPointForTextTest(path geom.Path, want geom.Pt) bool {
	for _, pt := range path.V {
		if approx(pt.X, want.X, 1e-9) && approx(pt.Y, want.Y, 1e-9) {
			return true
		}
	}
	return false
}

func pathHasPointNearForTextTest(path geom.Path, want geom.Pt, tol float64) bool {
	for _, pt := range path.V {
		if math.Hypot(pt.X-want.X, pt.Y-want.Y) <= tol {
			return true
		}
	}
	return false
}

func approxRect(got, want geom.Rect, tol float64) bool {
	return approx(got.Min.X, want.Min.X, tol) &&
		approx(got.Min.Y, want.Min.Y, tol) &&
		approx(got.Max.X, want.Max.X, tol) &&
		approx(got.Max.Y, want.Max.Y, tol)
}

func containsMathRunText(runs []MathTextLayoutRun, text string) bool {
	for _, run := range runs {
		if run.Text == text {
			return true
		}
	}
	return false
}

func containsMathRun(runs []MathTextLayoutRun, text string, size float64) bool {
	for _, run := range runs {
		if run.Text == text && almostEqualFloat(run.FontSize, size) {
			return true
		}
	}
	return false
}

func almostEqualFloat(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 1e-9
}

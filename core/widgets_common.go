package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// WidgetCallbackID identifies a registered interactive widget callback.
type WidgetCallbackID int64

type widgetCallbackRegistry[T any] struct {
	next  WidgetCallbackID
	order []WidgetCallbackID

	callbacks map[WidgetCallbackID]T
}

type widgetVisualDefaults struct {
	ButtonFace             render.Color
	ButtonEdge             render.Color
	ButtonText             render.Color
	ButtonPad              float64
	ButtonRadius           float64
	ButtonLineWidth        float64
	ButtonHoverFace        render.Color
	ButtonHoverBlend       float64
	ButtonPressedBlend     float64
	PanelFace              render.Color
	PanelEdge              render.Color
	PanelPad               float64
	PanelRadius            float64
	PanelLineWidth         float64
	Track                  render.Color
	Fill                   render.Color
	Handle                 render.Color
	HandleEdge             render.Color
	Text                   render.Color
	Check                  render.Color
	RadioDot               render.Color
	TextBoxFace            render.Color
	TextBoxEdge            render.Color
	SliderPanelPad         float64
	SliderRadius           float64
	SliderLabelX           widgetStyleCoordinate
	SliderLabelY           widgetStyleCoordinate
	SliderLabelAlign       TextAlign
	SliderValueX           widgetStyleCoordinate
	SliderValueY           widgetStyleCoordinate
	SliderValueAlign       TextAlign
	SliderTextVAlign       textLayoutVerticalAlign
	SliderTrackXPad        float64
	SliderTrackYMin        float64
	SliderTrackYMax        float64
	SliderTrackRadius      float64
	SliderHandleSize       float64
	SliderHandleLine       float64
	SliderInitColor        render.Color
	SliderInitLine         float64
	RangeTupleText         bool
	TextBoxPanelPad        float64
	TextBoxLabelX          widgetStyleCoordinate
	TextBoxLabelY          widgetStyleCoordinate
	TextBoxLabelAlign      TextAlign
	TextBoxLabelVAlign     textLayoutVerticalAlign
	TextBoxInputXPad       float64
	TextBoxInputYMin       float64
	TextBoxInputYMax       float64
	TextBoxTextX           widgetStyleCoordinate
	TextBoxTextY           widgetStyleCoordinate
	TextBoxTextAlign       TextAlign
	TextBoxTextVAlign      textLayoutVerticalAlign
	TextBoxCaretYMin       widgetStyleCoordinate
	TextBoxCaretYMax       widgetStyleCoordinate
	TextBoxRadius          float64
	TextBoxLineWidth       float64
	TextBoxActiveEdgeBlend float64
	CheckBoxFace           render.Color
	CheckBoxMaxSize        float64
	CheckBoxScale          float64
	CheckBoxXPad           float64
	CheckLabelGap          float64
	CheckBoxRadius         float64
	CheckBoxLineWidth      float64
	CheckMarkWidth         float64
	CheckMarkStyle         widgetCheckMarkStyle
	RadioCenterXPad        float64
	RadioOuterSize         float64
	RadioInnerSize         float64
	RadioLineWidth         float64
	RadioLabelGap          float64
	RadioInactiveFace      render.Color
	RadioActiveOuter       bool
}

type widgetStyleCoordinate struct {
	Value    float64
	Fraction bool
}

type widgetCheckMarkStyle uint8

const (
	widgetCheckMarkTick widgetCheckMarkStyle = iota
	widgetCheckMarkX
)

func widgetPixelCoord(value float64) widgetStyleCoordinate {
	return widgetStyleCoordinate{Value: value}
}

func widgetFractionCoord(value float64) widgetStyleCoordinate {
	return widgetStyleCoordinate{Value: value, Fraction: true}
}

func widgetDefaultsForAxes(a *Axes) widgetVisualDefaults {
	if a == nil {
		return widgetDefaultsForRC(style.Default)
	}
	return widgetDefaultsForRC(a.resolvedRC())
}

func widgetDefaultsForRC(rc style.RC) widgetVisualDefaults {
	switch rc.WidgetVisualStyle {
	case style.WidgetVisualMatplotlib:
		return widgetVisualDefaults{
			ButtonFace:             render.Color{R: 0.85, G: 0.85, B: 0.85, A: 1},
			ButtonEdge:             render.Color{R: 0, G: 0, B: 0, A: 1},
			ButtonText:             render.Color{R: 0, G: 0, B: 0, A: 1},
			ButtonPad:              0,
			ButtonRadius:           0,
			ButtonLineWidth:        1,
			ButtonHoverFace:        render.Color{R: 0.95, G: 0.95, B: 0.95, A: 1},
			ButtonHoverBlend:       0,
			ButtonPressedBlend:     0.12,
			PanelFace:              render.Color{R: 1, G: 1, B: 1, A: 1},
			PanelEdge:              render.Color{R: 0, G: 0, B: 0, A: 1},
			PanelPad:               0,
			PanelRadius:            0,
			PanelLineWidth:         1,
			Track:                  render.Color{R: 211.0 / 255.0, G: 211.0 / 255.0, B: 211.0 / 255.0, A: 1},
			Fill:                   render.Color{R: 31.0 / 255.0, G: 119.0 / 255.0, B: 180.0 / 255.0, A: 1},
			Handle:                 render.Color{R: 1, G: 1, B: 1, A: 1},
			HandleEdge:             render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1},
			Text:                   render.Color{R: 0, G: 0, B: 0, A: 1},
			Check:                  render.Color{R: 0, G: 0, B: 0, A: 1},
			RadioDot:               render.Color{R: 0, G: 0, B: 1, A: 1},
			TextBoxFace:            render.Color{R: 0.95, G: 0.95, B: 0.95, A: 1},
			TextBoxEdge:            render.Color{R: 0, G: 0, B: 0, A: 1},
			SliderPanelPad:         0,
			SliderRadius:           0,
			SliderLabelX:           widgetFractionCoord(-0.02),
			SliderLabelY:           widgetFractionCoord(0.5),
			SliderLabelAlign:       TextAlignRight,
			SliderValueX:           widgetFractionCoord(1.02),
			SliderValueY:           widgetFractionCoord(0.5),
			SliderValueAlign:       TextAlignLeft,
			SliderTextVAlign:       textLayoutVAlignCenter,
			SliderTrackXPad:        0,
			SliderTrackYMin:        0.25,
			SliderTrackYMax:        0.75,
			SliderTrackRadius:      0,
			SliderHandleSize:       10,
			SliderHandleLine:       1,
			SliderInitColor:        render.Color{R: 1, G: 0, B: 0, A: 1},
			SliderInitLine:         1,
			RangeTupleText:         true,
			TextBoxPanelPad:        0,
			TextBoxLabelX:          widgetFractionCoord(-0.01),
			TextBoxLabelY:          widgetFractionCoord(0.5),
			TextBoxLabelAlign:      TextAlignRight,
			TextBoxLabelVAlign:     textLayoutVAlignCenter,
			TextBoxInputXPad:       0,
			TextBoxInputYMin:       0,
			TextBoxInputYMax:       1,
			TextBoxTextX:           widgetFractionCoord(0.05),
			TextBoxTextY:           widgetFractionCoord(0.5),
			TextBoxTextAlign:       TextAlignLeft,
			TextBoxTextVAlign:      textLayoutVAlignCenter,
			TextBoxCaretYMin:       widgetFractionCoord(0.32),
			TextBoxCaretYMax:       widgetFractionCoord(0.68),
			TextBoxRadius:          0,
			TextBoxLineWidth:       1,
			TextBoxActiveEdgeBlend: 0,
			CheckBoxFace:           render.Color{A: 0},
			CheckBoxMaxSize:        8,
			CheckBoxScale:          0.50,
			CheckBoxXPad:           0.15,
			CheckLabelGap:          0.25,
			CheckBoxRadius:         0,
			CheckBoxLineWidth:      1,
			CheckMarkWidth:         1,
			CheckMarkStyle:         widgetCheckMarkX,
			RadioCenterXPad:        0.15,
			RadioOuterSize:         8,
			RadioInnerSize:         0,
			RadioLineWidth:         1,
			RadioLabelGap:          0.25,
			RadioInactiveFace:      render.Color{A: 0},
			RadioActiveOuter:       true,
		}
	default:
		return widgetVisualDefaults{
			ButtonFace:             render.Color{R: 0.94, G: 0.95, B: 0.97, A: 1},
			ButtonEdge:             render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
			ButtonText:             render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
			ButtonPad:              6,
			ButtonRadius:           10,
			ButtonLineWidth:        1.25,
			ButtonHoverFace:        render.Color{A: 0},
			ButtonHoverBlend:       0.06,
			ButtonPressedBlend:     0.12,
			PanelFace:              render.Color{R: 0.96, G: 0.97, B: 0.98, A: 1},
			PanelEdge:              render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
			PanelPad:               4,
			PanelRadius:            12,
			PanelLineWidth:         1.1,
			Track:                  render.Color{R: 0.83, G: 0.85, B: 0.89, A: 1},
			Fill:                   render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
			Handle:                 render.Color{R: 0.09, G: 0.18, B: 0.34, A: 1},
			HandleEdge:             render.Color{R: 0.272, G: 0.344, B: 0.472, A: 1},
			Text:                   render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
			Check:                  render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
			RadioDot:               render.Color{R: 0.85, G: 0.32, B: 0.17, A: 1},
			TextBoxFace:            render.Color{R: 1, G: 1, B: 1, A: 1},
			TextBoxEdge:            render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
			SliderPanelPad:         4,
			SliderRadius:           12,
			SliderLabelX:           widgetPixelCoord(14),
			SliderLabelY:           widgetPixelCoord(22),
			SliderLabelAlign:       TextAlignLeft,
			SliderValueX:           widgetPixelCoord(-14),
			SliderValueY:           widgetPixelCoord(22),
			SliderValueAlign:       TextAlignRight,
			SliderTextVAlign:       textLayoutVAlignTop,
			SliderTrackXPad:        14,
			SliderTrackYMin:        -26,
			SliderTrackYMax:        -14,
			SliderTrackRadius:      -1,
			SliderHandleSize:       1.9,
			SliderHandleLine:       1,
			SliderInitColor:        render.Color{A: 0},
			SliderInitLine:         0,
			RangeTupleText:         false,
			TextBoxPanelPad:        4,
			TextBoxLabelX:          widgetPixelCoord(4),
			TextBoxLabelY:          widgetPixelCoord(20),
			TextBoxLabelAlign:      TextAlignLeft,
			TextBoxLabelVAlign:     textLayoutVAlignTop,
			TextBoxInputXPad:       4,
			TextBoxInputYMin:       30,
			TextBoxInputYMax:       -8,
			TextBoxTextX:           widgetPixelCoord(12),
			TextBoxTextY:           widgetFractionCoord(0.5),
			TextBoxTextAlign:       TextAlignLeft,
			TextBoxTextVAlign:      textLayoutVAlignCenter,
			TextBoxCaretYMin:       widgetPixelCoord(8),
			TextBoxCaretYMax:       widgetPixelCoord(-8),
			TextBoxRadius:          8,
			TextBoxLineWidth:       1.2,
			TextBoxActiveEdgeBlend: 0.65,
			CheckBoxFace:           render.Color{R: 1, G: 1, B: 1, A: 1},
			CheckBoxMaxSize:        16,
			CheckBoxScale:          0.42,
			CheckBoxXPad:           14,
			CheckLabelGap:          10,
			CheckBoxRadius:         3,
			CheckBoxLineWidth:      1,
			CheckMarkWidth:         2,
			CheckMarkStyle:         widgetCheckMarkTick,
			RadioCenterXPad:        24,
			RadioOuterSize:         16,
			RadioInnerSize:         8,
			RadioLineWidth:         1,
			RadioLabelGap:          16,
			RadioInactiveFace:      render.Color{R: 1, G: 1, B: 1, A: 1},
			RadioActiveOuter:       false,
		}
	}
}

func (r *widgetCallbackRegistry[T]) add(cb T) WidgetCallbackID {
	if r == nil {
		return 0
	}
	if r.next == 0 {
		r.next = 1
	} else {
		r.next++
	}
	if r.callbacks == nil {
		r.callbacks = make(map[WidgetCallbackID]T)
	}
	id := r.next
	r.callbacks[id] = cb
	r.order = append(r.order, id)
	return id
}

func (r *widgetCallbackRegistry[T]) remove(id WidgetCallbackID) {
	if r == nil || id == 0 || r.callbacks == nil {
		return
	}
	delete(r.callbacks, id)
	for i, existing := range r.order {
		if existing == id {
			r.order = append(r.order[:i], r.order[i+1:]...)
			return
		}
	}
}

func (r *widgetCallbackRegistry[T]) each(fn func(T)) {
	if r == nil || fn == nil {
		return
	}
	for _, id := range r.order {
		cb, ok := r.callbacks[id]
		if !ok {
			continue
		}
		fn(cb)
	}
}

// AddWidgetAxes appends an axes prepared for widget controls.
func (f *Figure) AddWidgetAxes(r geom.Rect, opts ...style.Option) *Axes {
	if f == nil {
		return nil
	}
	ax := f.AddAxes(r, opts...)
	prepareWidgetAxes(ax)
	return ax
}

// AddWidgetAxes appends widget axes inside the subfigure rectangle.
func (sf *SubFigure) AddWidgetAxes(r geom.Rect, opts ...style.Option) *Axes {
	if sf == nil || sf.figure == nil {
		return nil
	}
	ax := sf.figure.AddWidgetAxes(composeRect(sf.RectFraction, r), opts...)
	return ax
}

// AddWidgetAxes creates widget axes covering this subplot span.
func (spec SubplotSpec) AddWidgetAxes(opts ...SubplotAxesOption) *Axes {
	ax := spec.AddAxes(opts...)
	prepareWidgetAxes(ax)
	return ax
}

func prepareWidgetAxes(a *Axes) {
	if a == nil {
		return
	}
	if a.XAxis != nil {
		a.XAxis.ShowSpine = false
		a.XAxis.ShowTicks = false
		a.XAxis.ShowLabels = false
	}
	if a.YAxis != nil {
		a.YAxis.ShowSpine = false
		a.YAxis.ShowTicks = false
		a.YAxis.ShowLabels = false
	}
	if a.XAxisTop != nil {
		a.XAxisTop.ShowSpine = false
		a.XAxisTop.ShowTicks = false
		a.XAxisTop.ShowLabels = false
	}
	if a.YAxisRight != nil {
		a.YAxisRight.ShowSpine = false
		a.YAxisRight.ShowTicks = false
		a.YAxisRight.ShowLabels = false
	}
	for _, axis := range a.ExtraAxes {
		if axis == nil {
			continue
		}
		axis.ShowSpine = false
		axis.ShowTicks = false
		axis.ShowLabels = false
	}
	a.ShowFrame = false
	a.SetXLim(0, 1)
	a.SetYLim(0, 1)
}

func widgetStyledPanelRect(rect geom.Rect, pad float64) geom.Rect {
	return insetRect(rect, pad)
}

func widgetStyledSliderTrack(panel geom.Rect, defaults widgetVisualDefaults) geom.Rect {
	yMin := widgetStyleCoord(panel.Min.Y, panel.Max.Y, defaults.SliderTrackYMin)
	yMax := widgetStyleCoord(panel.Min.Y, panel.Max.Y, defaults.SliderTrackYMax)
	if yMin > yMax {
		yMin, yMax = yMax, yMin
	}
	return geom.Rect{
		Min: geom.Pt{X: panel.Min.X + defaults.SliderTrackXPad, Y: yMin},
		Max: geom.Pt{X: panel.Max.X - defaults.SliderTrackXPad, Y: yMax},
	}
}

func widgetSliderTrackForContext(ctx *DrawContext) geom.Rect {
	if ctx == nil {
		return geom.Rect{}
	}
	defaults := widgetDefaultsForRC(ctx.RC)
	panel := widgetStyledPanelRect(ctx.Clip, defaults.SliderPanelPad)
	return widgetStyledSliderTrack(panel, defaults)
}

func widgetSliderTrackRadius(track geom.Rect, defaults widgetVisualDefaults) float64 {
	if defaults.SliderTrackRadius >= 0 {
		return defaults.SliderTrackRadius
	}
	return track.H() / 2
}

func widgetTextBoxInputRect(panel geom.Rect, defaults widgetVisualDefaults) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{
			X: panel.Min.X + defaults.TextBoxInputXPad,
			Y: widgetStyleCoord(panel.Min.Y, panel.Max.Y, defaults.TextBoxInputYMin),
		},
		Max: geom.Pt{
			X: panel.Max.X - defaults.TextBoxInputXPad,
			Y: widgetStyleCoord(panel.Min.Y, panel.Max.Y, defaults.TextBoxInputYMax),
		},
	}
}

func widgetTextBoxTextAnchor(input geom.Rect, defaults widgetVisualDefaults) geom.Pt {
	return geom.Pt{
		X: widgetResolvedCoord(input.Min.X, input.Max.X, defaults.TextBoxTextX),
		Y: widgetResolvedCoord(input.Min.Y, input.Max.Y, defaults.TextBoxTextY),
	}
}

func widgetButtonRowCenterY(panel geom.Rect, index, count int, sourceLayoutSentinel float64) float64 {
	if count <= 0 {
		return panel.Min.Y
	}
	if sourceLayoutSentinel >= 0 && sourceLayoutSentinel <= 1 {
		fraction := 1 - float64(index+1)/float64(count+1)
		return panel.Min.Y + panel.H()*fraction
	}
	rowHeight := panel.H() / float64(count)
	return panel.Max.Y - rowHeight*float64(index) - rowHeight/2
}

func widgetStyleCoord(minV, maxV, value float64) float64 {
	if value >= 0 && value <= 1 {
		return minV + (maxV-minV)*value
	}
	if value < 0 {
		return maxV + value
	}
	return minV + value
}

func widgetResolvedCoord(minV, maxV float64, coord widgetStyleCoordinate) float64 {
	if coord.Fraction {
		return minV + (maxV-minV)*coord.Value
	}
	return widgetStyleCoord(minV, maxV, coord.Value)
}

func drawWidgetPanel(r render.Renderer, rect geom.Rect, fill, edge render.Color, width, radius float64) {
	if rect.W() <= 0 || rect.H() <= 0 {
		return
	}
	path := roundedRectPath(rect, math.Min(radius, math.Min(rect.W(), rect.H())/2))
	paint := render.Paint{Fill: fill}
	if edge.A > 0 && width > 0 {
		paint.Stroke = edge
		paint.LineWidth = width
		paint.LineJoin = render.JoinRound
		paint.LineCap = render.CapRound
		if radius <= 0 {
			paint.Snap = render.SnapAuto
		}
	}
	r.Path(path, &paint)
}

func drawCenteredWidgetText(r render.Renderer, ctx *DrawContext, center geom.Pt, text string, size float64, color render.Color) {
	drawWidgetText(r, ctx, center, text, size, color, TextAlignCenter, textLayoutVAlignCenter)
}

func drawWidgetText(r render.Renderer, ctx *DrawContext, anchor geom.Pt, text string, size float64, color render.Color, hAlign TextAlign, vAlign textLayoutVerticalAlign) {
	textRen, ok := r.(render.TextDrawer)
	if !ok || displayTextIsEmpty(text) {
		return
	}
	fontSize := resolvedFontSize(size, ctx)
	layout := measureSingleLineTextLayout(r, text, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	origin := alignedSingleLineOrigin(anchor, layout, hAlign, vAlign)
	drawDisplayText(textRen, text, origin, fontSize, resolvedTextColor(color, ctx), ctx.RC.FontKey, ctx.RC.UseTeX)
}

func insetRect(rect geom.Rect, pad float64) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: rect.Min.X + pad, Y: rect.Min.Y + pad},
		Max: geom.Pt{X: rect.Max.X - pad, Y: rect.Max.Y - pad},
	}
}

func clampInt(v, minVal, maxVal int) int {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

func mixColor(a, b render.Color, t float64) render.Color {
	t = clampFloat(t, 0, 1)
	return render.Color{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
		A: a.A + (b.A-a.A)*t,
	}
}

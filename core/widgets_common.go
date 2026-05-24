package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// WidgetCallbackID identifies a registered interactive widget callback.
type WidgetCallbackID int64

type widgetCallbackRegistry[T any] struct {
	next      WidgetCallbackID
	callbacks map[WidgetCallbackID]T
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
	return id
}

func (r *widgetCallbackRegistry[T]) remove(id WidgetCallbackID) {
	if r == nil || id == 0 || r.callbacks == nil {
		return
	}
	delete(r.callbacks, id)
}

func (r *widgetCallbackRegistry[T]) each(fn func(T)) {
	if r == nil || fn == nil {
		return
	}
	for _, cb := range r.callbacks {
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

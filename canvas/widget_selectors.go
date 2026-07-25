package canvas

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/widgets"
)

type selectorDragKind uint8

type selectorDragState struct {
	kind selectorDragKind

	axes    *Axes
	span    *widgets.SpanSelector
	rect    *widgets.RectangleSelector
	ellipse *widgets.EllipseSelector
	polygon *widgets.PolygonSelector
	lasso   *widgets.LassoSelector

	pressData geom.Pt

	spanStart float64
	spanEnd   float64
	spanMove  bool

	rectStartMin geom.Pt
	rectStartMax geom.Pt
	rectMove     bool

	polygonStartPoints []geom.Pt
	polygonVertexIdx   int
	polygonPendingIdx  int

	square bool
	center bool
}

func (w *WidgetInteraction) findInactiveSelector(axes *Axes) any {
	if axes == nil {
		return nil
	}
	artists := axesInteractionArtists(axes)
	for i := len(artists) - 1; i >= 0; i-- {
		switch sel := artists[i].(type) {
		case *widgets.SpanSelector:
			if !sel.Active {
				return sel
			}
		case *widgets.RectangleSelector:
			if !sel.Active {
				return sel
			}
		case *widgets.EllipseSelector:
			if !sel.Active {
				return sel
			}
		case *widgets.PolygonSelector:
			if !sel.Active {
				return sel
			}
		case *widgets.LassoSelector:
			if !sel.Active {
				return sel
			}
		}
	}
	return nil
}

func (w *WidgetInteraction) clearSelectorDragLocked() {
	switch {
	case w.draggingSelector.kind == selectorDragLasso && w.draggingSelector.lasso != nil:
		w.draggingSelector.lasso.Clear()
	case w.draggingSelector.kind == selectorDragNone:
	}
	w.draggingSelector = selectorDragState{}
}

func (w *WidgetInteraction) updateDraggingSelectorFromMouseLocked(mouse *MouseEvent) bool {
	state := w.draggingSelector
	data, ok := w.pixelToDataLocked(mouse, state.axes)
	if !ok {
		return false
	}
	changed := false

	switch state.kind {
	case selectorDragSpan:
		if state.span == nil {
			return false
		}
		cursor := data.X
		if state.span.Orientation == "vertical" {
			cursor = data.Y
		}
		pressCursor := state.spanStart
		if state.span.Orientation == "vertical" {
			pressCursor = state.pressData.Y
		}
		if state.spanMove {
			delta := cursor - pressCursor
			changed = state.span.SetSpan(state.spanStart+delta, state.spanEnd+delta)
		} else {
			changed = state.span.SetSpan(pressCursor, cursor)
		}
	case selectorDragRectangle:
		if state.rect == nil {
			return false
		}
		delta := geom.Pt{
			X: data.X - state.pressData.X,
			Y: data.Y - state.pressData.Y,
		}
		if state.rectMove {
			minPoint := state.rectStartMin
			maxPoint := state.rectStartMax
			minPoint = geom.Pt{X: minPoint.X + delta.X, Y: minPoint.Y + delta.Y}
			maxPoint = geom.Pt{X: maxPoint.X + delta.X, Y: maxPoint.Y + delta.Y}
			changed = state.rect.SetBounds(minPoint, maxPoint)
			break
		} else {
			minPoint, maxPoint := selectorBoundsFromDrag(state.pressData, data, state.square, state.center)
			changed = state.rect.SetBounds(minPoint, maxPoint)
		}
	case selectorDragEllipse:
		if state.ellipse == nil {
			return false
		}
		delta := geom.Pt{
			X: data.X - state.pressData.X,
			Y: data.Y - state.pressData.Y,
		}
		if state.rectMove {
			minPoint := state.rectStartMin
			maxPoint := state.rectStartMax
			minPoint = geom.Pt{X: minPoint.X + delta.X, Y: minPoint.Y + delta.Y}
			maxPoint = geom.Pt{X: maxPoint.X + delta.X, Y: maxPoint.Y + delta.Y}
			changed = state.ellipse.SetBounds(minPoint, maxPoint)
			break
		} else {
			minPoint, maxPoint := selectorBoundsFromDrag(state.pressData, data, state.square, state.center)
			changed = state.ellipse.SetBounds(minPoint, maxPoint)
		}
	case selectorDragPolygonDrawing:
		if state.polygon == nil {
			return false
		}
		changed = state.polygon.SetPoint(state.polygonPendingIdx, data)
	case selectorDragPolygonMoveAll:
		if state.polygon == nil || len(state.polygonStartPoints) != len(state.polygon.Points) {
			return false
		}
		delta := geom.Pt{
			X: data.X - state.pressData.X,
			Y: data.Y - state.pressData.Y,
		}
		for i, point := range state.polygonStartPoints {
			if state.polygon.SetPoint(i, geom.Pt{X: point.X + delta.X, Y: point.Y + delta.Y}) {
				changed = true
			}
		}
	case selectorDragPolygonMoveVertex:
		if state.polygon == nil || state.polygonVertexIdx < 0 || state.polygonVertexIdx >= len(state.polygonStartPoints) {
			return false
		}
		base := state.polygonStartPoints[state.polygonVertexIdx]
		changed = state.polygon.SetPoint(state.polygonVertexIdx, geom.Pt{
			X: base.X + (data.X - state.pressData.X),
			Y: base.Y + (data.Y - state.pressData.Y),
		})
	case selectorDragLasso:
		if state.lasso == nil {
			return false
		}
		changed = state.lasso.AddPoint(data)
	}
	return changed
}

func (w *WidgetInteraction) finalizeDraggingSelectorLocked() bool {
	state := w.draggingSelector
	changed := false

	switch state.kind {
	case selectorDragSpan:
		if state.span != nil && state.span.Active {
			state.span.TriggerOnSelect()
			changed = true
		}
	case selectorDragRectangle:
		if state.rect != nil && state.rect.Active {
			state.rect.TriggerOnSelect()
			changed = true
		}
	case selectorDragEllipse:
		if state.ellipse != nil && state.ellipse.Active {
			state.ellipse.TriggerOnSelect()
			changed = true
		}
	case selectorDragPolygonDrawing, selectorDragPolygonMoveAll, selectorDragPolygonMoveVertex:
		if state.polygon != nil && state.polygon.Active && state.polygon.Closed {
			state.polygon.TriggerOnSelect()
			changed = true
		}
	case selectorDragLasso:
		if state.lasso != nil && state.lasso.Finish() {
			changed = true
		}
	}
	w.draggingSelector = selectorDragState{}
	return changed
}

func (w *WidgetInteraction) handleSelectorKey(modifiers Modifier, key string) bool {
	if w.focusedSelector == nil {
		return false
	}
	axes := w.axesForSelectorLocked(w.focusedSelector)
	command := strings.ToLower(key)
	switch keyState := w.focusedSelector.(type) {
	case *widgets.SpanSelector:
		return w.handleSpanSelectorKey(modifiers, command, keyState, axes)
	case *widgets.RectangleSelector:
		return w.handleRectangleSelectorKey(modifiers, command, keyState, axes)
	case *widgets.EllipseSelector:
		return w.handleEllipseSelectorKey(modifiers, command, keyState, axes)
	case *widgets.PolygonSelector:
		return w.handlePolygonSelectorKey(modifiers, command, keyState, axes)
	case *widgets.LassoSelector:
		return w.handleLassoSelectorKey(command, keyState)
	default:
		return false
	}
}

func (w *WidgetInteraction) handleSpanSelectorKey(modifiers Modifier, key string, selector *widgets.SpanSelector, axes *Axes) bool {
	if selector == nil || !selector.Active {
		return false
	}
	if key == "escape" {
		return selector.Clear()
	}
	if axes == nil || w.figure == nil {
		return false
	}
	ctx := core.AxesDrawContext(axes, w.figure)
	if ctx == nil || ctx.DataToPixel.XScale == nil || ctx.DataToPixel.YScale == nil {
		return false
	}
	xMin, xMax := ctx.DataToPixel.XScale.Domain()
	yMin, yMax := ctx.DataToPixel.YScale.Domain()
	stepX := (xMax - xMin) / 20
	stepY := (yMax - yMin) / 20
	if stepX == 0 {
		stepX = 1
	}
	if stepY == 0 {
		stepY = 1
	}
	delta := 0.0
	switch key {
	case "left":
		delta = -stepX
	case "right":
		delta = stepX
	case "up":
		delta = stepY
	case "down":
		delta = -stepY
	default:
		return false
	}
	if selector.Orientation == "vertical" {
		delta = -delta
	}
	if modifiers&ModifierControl != 0 {
		delta *= 10
	}
	changed := selector.Move(delta)
	if changed {
		selector.TriggerOnSelect()
	}
	return changed
}

func (w *WidgetInteraction) handleRectangleSelectorKey(modifiers Modifier, key string, selector *widgets.RectangleSelector, axes *Axes) bool {
	if selector == nil {
		return false
	}
	if key == "escape" {
		return selector.Clear()
	}
	if !selector.Active {
		return false
	}
	delta, ok := w.selectorMoveDeltaForKey(modifiers, key, axes)
	if !ok {
		return false
	}
	changed := selector.MoveBy(delta)
	if changed {
		selector.TriggerOnSelect()
	}
	return changed
}

func (w *WidgetInteraction) handleEllipseSelectorKey(modifiers Modifier, key string, selector *widgets.EllipseSelector, axes *Axes) bool {
	if selector == nil {
		return false
	}
	if key == "escape" {
		return selector.Clear()
	}
	if !selector.Active {
		return false
	}
	delta, ok := w.selectorMoveDeltaForKey(modifiers, key, axes)
	if !ok {
		return false
	}
	changed := selector.MoveBy(delta)
	if changed {
		selector.TriggerOnSelect()
	}
	return changed
}

func (w *WidgetInteraction) selectorMoveDeltaForKey(modifiers Modifier, key string, axes *Axes) (geom.Pt, bool) {
	if axes == nil || w.figure == nil {
		return geom.Pt{}, false
	}
	if key == "escape" {
		return geom.Pt{}, false
	}
	if key != "left" && key != "right" && key != "up" && key != "down" {
		return geom.Pt{}, false
	}
	ctx := core.AxesDrawContext(axes, w.figure)
	if ctx == nil || ctx.DataToPixel.XScale == nil || ctx.DataToPixel.YScale == nil {
		return geom.Pt{}, false
	}
	xMin, xMax := ctx.DataToPixel.XScale.Domain()
	yMin, yMax := ctx.DataToPixel.YScale.Domain()
	delta := geom.Pt{
		X: (xMax - xMin) / 20,
		Y: (yMax - yMin) / 20,
	}
	if delta.X == 0 {
		delta.X = 1
	}
	if delta.Y == 0 {
		delta.Y = 1
	}
	switch key {
	case "left":
		delta.Y = 0
		delta.X = -delta.X
	case "right":
		delta.Y = 0
	case "up":
		delta.X = 0
	case "down":
		delta.X = 0
		delta.Y = -delta.Y
	}
	if modifiers&ModifierControl != 0 {
		delta.X *= 10
		delta.Y *= 10
	}
	return delta, true
}

func (w *WidgetInteraction) handlePolygonSelectorKey(modifiers Modifier, key string, selector *widgets.PolygonSelector, axes *Axes) bool {
	if selector == nil {
		return false
	}
	if key == "escape" {
		return selector.Clear()
	}
	if !selector.Active || !selector.Closed || axes == nil || w.figure == nil {
		return false
	}
	ctx := core.AxesDrawContext(axes, w.figure)
	if ctx == nil || ctx.DataToPixel.XScale == nil || ctx.DataToPixel.YScale == nil {
		return false
	}
	xMin, xMax := ctx.DataToPixel.XScale.Domain()
	yMin, yMax := ctx.DataToPixel.YScale.Domain()
	step := geom.Pt{
		X: (xMax - xMin) / 20,
		Y: (yMax - yMin) / 20,
	}
	delta := geom.Pt{}
	switch key {
	case "left":
		delta.X = -step.X
	case "right":
		delta.X = step.X
	case "up":
		delta.Y = step.Y
	case "down":
		delta.Y = -step.Y
	default:
		return false
	}
	if modifiers&ModifierControl != 0 {
		delta.X *= 10
		delta.Y *= 10
	}
	changed := false
	for i, pt := range selector.Points {
		if selector.SetPoint(i, geom.Pt{X: pt.X + delta.X, Y: pt.Y + delta.Y}) {
			changed = true
		}
	}
	if changed {
		selector.TriggerOnSelect()
	}
	return changed
}

func (w *WidgetInteraction) handleLassoSelectorKey(key string, selector *widgets.LassoSelector) bool {
	if selector == nil {
		return false
	}
	if key == "escape" {
		return selector.Clear()
	}
	return false
}

func (w *WidgetInteraction) axesForSelectorLocked(selector any) *Axes {
	if selector == nil {
		return nil
	}
	fig := w.figure
	if fig == nil {
		return nil
	}
	for _, axes := range fig.Children {
		if axes == nil {
			continue
		}
		for _, artist := range axesInteractionArtists(axes) {
			if artist == selector {
				return axes
			}
		}
	}
	return nil
}

func selectorBoundsFromDrag(
	press, current geom.Pt,
	square bool,
	center bool,
) (geom.Pt, geom.Pt) {
	minPoint := press
	maxPoint := current
	if center {
		delta := geom.Pt{
			X: current.X - press.X,
			Y: current.Y - press.Y,
		}
		minPoint = geom.Pt{X: press.X - delta.X, Y: press.Y - delta.Y}
		maxPoint = geom.Pt{X: press.X + delta.X, Y: press.Y + delta.Y}
	}
	if square {
		sideX := current.X - press.X
		sideY := current.Y - press.Y
		side := math.Abs(sideX)
		if math.Abs(sideY) > side {
			side = math.Abs(sideY)
		}
		if side == 0 {
			return geom.Pt{X: press.X, Y: press.Y}, geom.Pt{X: press.X, Y: press.Y}
		}
		sx := sideX
		if sx == 0 {
			sx = 1
		}
		if sx > 0 {
			sx = side
		} else {
			sx = -side
		}
		sy := sideY
		if sy == 0 {
			sy = 1
		}
		if sy > 0 {
			sy = side
		} else {
			sy = -side
		}
		maxPoint = geom.Pt{
			X: press.X + sx,
			Y: press.Y + sy,
		}
		if center {
			minPoint = geom.Pt{
				X: press.X - sx,
				Y: press.Y - sy,
			}
		}
	}
	return normalizedMinMax(minPoint, maxPoint)
}

func normalizedMinMax(minPoint, maxPoint geom.Pt) (geom.Pt, geom.Pt) {
	if minPoint.X <= maxPoint.X && minPoint.Y <= maxPoint.Y {
		return minPoint, maxPoint
	}
	return geom.Pt{X: math.Min(minPoint.X, maxPoint.X), Y: math.Min(minPoint.Y, maxPoint.Y)}, geom.Pt{X: math.Max(minPoint.X, maxPoint.X), Y: math.Max(minPoint.Y, maxPoint.Y)}
}

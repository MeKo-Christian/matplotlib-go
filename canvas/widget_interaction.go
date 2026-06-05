package canvas

import (
	"sync"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

type WidgetInteraction struct {
	mu sync.Mutex

	figure *Figure
	draw   func() error

	dispatch *Dispatcher
	connects []ConnectionID

	hoveredButton *core.Button
	pressedButton *core.Button
	focusedButton *core.Button

	draggingSlider            *core.Slider
	draggingSliderAxes        *Axes
	draggingRangeSlider       *core.RangeSlider
	draggingRangeSliderAxes   *Axes
	draggingRangeSliderHandle int

	draggingSelector selectorDragState
	focusedSelector  any

	focusedText              *core.TextBox
	focusedSlider            *core.Slider
	focusedRangeSlider       *core.RangeSlider
	focusedRangeSliderHandle int
	focusedCheck             *core.CheckButtons
	focusedCheckIndex        int
	focusedRadio             *core.RadioButtons
	focusedRadioIndex        int
}

const (
	selectorDragNone selectorDragKind = iota
	selectorDragSpan
	selectorDragRectangle
	selectorDragEllipse
	selectorDragPolygonDrawing
	selectorDragPolygonMoveVertex
	selectorDragPolygonMoveAll
	selectorDragLasso
)

func NewWidgetInteraction(fig *Figure, draw func() error) *WidgetInteraction {
	if fig == nil {
		return nil
	}
	return &WidgetInteraction{figure: fig, draw: draw}
}

func (w *WidgetInteraction) Attach(dispatcher *Dispatcher) {
	if w == nil || dispatcher == nil {
		return
	}
	w.Detach()

	w.mu.Lock()
	w.dispatch = dispatcher
	w.connects = append(
		w.connects,
		dispatcher.Connect(EventMousePress, func(ev Event) error {
			return w.handleMousePress(MouseEvent{Event: ev})
		}),
		dispatcher.Connect(EventMouseMove, func(ev Event) error {
			return w.handleMouseMove(MouseEvent{Event: ev})
		}),
		dispatcher.Connect(EventMouseRelease, func(ev Event) error {
			return w.handleMouseRelease(MouseEvent{Event: ev})
		}),
		dispatcher.Connect(EventFigureLeave, func(ev Event) error {
			return w.handleMouseLeave(MouseEvent{Event: ev})
		}),
		dispatcher.Connect(EventKeyPress, func(ev Event) error {
			return w.handleKeyPress(KeyEvent{Event: ev})
		}),
	)
	w.mu.Unlock()
}

func (w *WidgetInteraction) Detach() {
	if w == nil {
		return
	}
	w.mu.Lock()
	dispatcher := w.dispatch
	ids := w.connects
	w.connects = nil
	w.dispatch = nil
	w.mu.Unlock()

	for _, id := range ids {
		dispatcher.Disconnect(id)
	}
}

func (w *WidgetInteraction) handleMousePress(mouse MouseEvent) error {
	if w == nil {
		return nil
	}
	hit := w.pickWidget(mouse.Event)
	axes := w.resolveAxesForEvent(mouse)
	if axes == nil {
		axes = hit.axes
	}

	w.mu.Lock()
	changed := false

	if mouse.Button == MouseButtonLeft && hit.widget == nil && axes != nil {
		hit.widget = w.findInactiveSelector(axes)
	}

	if w.pressedButton != nil {
		w.pressedButton.Pressed = false
		w.pressedButton = nil
		changed = true
	}
	if w.draggingSlider != nil {
		w.draggingSlider.Dragging = false
		w.draggingSlider = nil
		w.draggingSliderAxes = nil
		changed = true
	}
	if w.draggingSelector.kind != selectorDragNone {
		w.clearSelectorDragLocked()
		changed = true
	}

	nextHovered := widgetButton(hit.widget)
	if w.hoveredButton != nextHovered {
		if w.hoveredButton != nil {
			w.hoveredButton.Hovered = false
		}
		w.hoveredButton = nextHovered
		if nextHovered != nil {
			nextHovered.Hovered = true
		}
		changed = true
	}

	if mouse.Button != MouseButtonLeft {
		if w.blurFocusedTextLocked() {
			changed = true
		}
		w.focusedSelector = nil
		w.focusedButton = nil
		w.focusedSlider = nil
		w.focusedRangeSlider = nil
		w.focusedRangeSliderHandle = 0
		w.focusedCheck = nil
		w.focusedCheckIndex = 0
		w.focusedRadio = nil
		w.focusedRadioIndex = 0
		w.focusedText = nil
		w.mu.Unlock()
		if changed {
			return w.callDraw()
		}
		return nil
	}

	w.focusedButton = nil
	w.focusedSlider = nil
	w.focusedRangeSlider = nil
	w.focusedRangeSliderHandle = 0
	w.focusedCheck = nil
	w.focusedCheckIndex = 0
	w.focusedRadio = nil
	w.focusedRadioIndex = 0
	w.focusedText = nil
	w.focusedSelector = nil

	switch widget := hit.widget.(type) {
	case *core.Button:
		w.focusedButton = widget
		if widget.Enabled {
			w.pressedButton = widget
			widget.Pressed = true
			widget.Hovered = true
			changed = true
		}
		if w.blurFocusedTextLocked() {
			changed = true
		}
	case *core.Slider:
		w.focusedSlider = widget
		if widget.Enabled {
			w.draggingSlider = widget
			w.draggingSliderAxes = axes
			widget.Dragging = true
			before := widget.Value
			w.setSliderValueFromPointLocked(widget, axes, mouse.Position)
			if widget.Value != before {
				changed = true
			}
		}
		if w.blurFocusedTextLocked() {
			changed = true
		}
	case *core.RangeSlider:
		w.focusedRangeSlider = widget
		w.focusedRangeSliderHandle = clampInt(hit.info.Index, 0, 1)
		if widget.Enabled {
			w.draggingRangeSlider = widget
			w.draggingRangeSliderAxes = axes
			w.draggingRangeSliderHandle = w.focusedRangeSliderHandle
			widget.Dragging = true
			beforeLow, beforeHigh := widget.Low, widget.High
			w.setRangeSliderValueFromPointLocked(widget, axes, mouse.Position, w.draggingRangeSliderHandle)
			if widget.Low != beforeLow || widget.High != beforeHigh {
				changed = true
			}
		}
		if w.blurFocusedTextLocked() {
			changed = true
		}
	case *core.CheckButtons:
		w.focusedCheck = widget
		w.focusedCheckIndex = clampInt(hit.info.Index, 0, len(widget.Labels)-1)
		if w.focusedCheckIndex < 0 {
			w.focusedCheckIndex = 0
		}
		w.focusedRadio = nil
		w.focusedRadioIndex = 0
		if len(widget.Labels) == 0 {
			break
		}
		before := widget.Values[w.focusedCheckIndex]
		widget.Toggle(w.focusedCheckIndex)
		if widget.Values[w.focusedCheckIndex] != before {
			changed = true
		}
		if w.blurFocusedTextLocked() {
			changed = true
		}
	case *core.RadioButtons:
		w.focusedRadio = widget
		w.focusedRadioIndex = clampInt(hit.info.Index, 0, len(widget.Labels)-1)
		if w.focusedRadioIndex < 0 {
			w.focusedRadioIndex = 0
		}
		w.focusedCheck = nil
		w.focusedCheckIndex = 0
		if len(widget.Labels) == 0 {
			break
		}
		if before := widget.Active; before != w.focusedRadioIndex {
			widget.SetActive(w.focusedRadioIndex)
			changed = true
		}
		if w.blurFocusedTextLocked() {
			changed = true
		}
	case *core.TextBox:
		w.focusedText = widget
		if !widget.Active {
			w.blurFocusedTextLocked()
			widget.Active = true
			changed = true
		}
		if axes != nil {
			index := cursorIndexFromTextPoint(widget, axes, w.figure, mouse.Position)
			beforeA, beforeB := widget.Selection()
			widget.SetSelection(index, index)
			if beforeA != index || beforeB != index {
				changed = true
			}
		}
	case *core.SpanSelector:
		w.focusedSelector = widget
		if data, ok := w.pixelToDataLocked(mouse, axes); ok {
			selectorData := data.X
			if widget.Orientation == "vertical" {
				selectorData = data.Y
			}
			w.draggingSelector = selectorDragState{
				kind:      selectorDragSpan,
				axes:      axes,
				span:      widget,
				pressData: data,
			}
			w.draggingSelector.spanStart = selectorData
			w.draggingSelector.spanEnd = selectorData
			if widget.Active {
				w.draggingSelector.spanMove = true
				w.draggingSelector.spanStart = widget.Start
				w.draggingSelector.spanEnd = widget.End
			} else if widget.SetSpan(selectorData, selectorData) {
				changed = true
			}
			changed = true
		}
	case *core.RectangleSelector:
		w.focusedSelector = widget
		if data, ok := w.pixelToDataLocked(mouse, axes); ok {
			w.draggingSelector = selectorDragState{
				kind:         selectorDragRectangle,
				axes:         axes,
				rect:         widget,
				pressData:    data,
				rectStartMin: widget.Min,
				rectStartMax: widget.Max,
				rectMove:     widget.Active,
				square:       mouse.Modifiers&ModifierShift != 0,
				center:       mouse.Modifiers&ModifierControl != 0,
			}
			if !widget.Active && widget.SetBounds(data, data) {
				changed = true
			}
			changed = true
		}
	case *core.EllipseSelector:
		w.focusedSelector = widget
		if data, ok := w.pixelToDataLocked(mouse, axes); ok {
			w.draggingSelector = selectorDragState{
				kind:         selectorDragEllipse,
				axes:         axes,
				ellipse:      widget,
				pressData:    data,
				rectStartMin: widget.Min,
				rectStartMax: widget.Max,
				rectMove:     widget.Active,
				square:       mouse.Modifiers&ModifierShift != 0,
				center:       mouse.Modifiers&ModifierControl != 0,
			}
			if !widget.Active && widget.SetBounds(data, data) {
				changed = true
			}
			changed = true
		}
	case *core.PolygonSelector:
		w.focusedSelector = widget
		if data, ok := w.pixelToDataLocked(mouse, axes); ok {
			shift := mouse.Modifiers&ModifierShift != 0
			control := mouse.Modifiers&ModifierControl != 0
			ctrlOrShift := control || shift
			state := selectorDragState{
				kind:      selectorDragPolygonDrawing,
				axes:      axes,
				polygon:   widget,
				pressData: data,
			}
			switch {
			case widget.Active && !widget.Closed && ctrlOrShift && control && hit.info.Index >= 0:
				state.kind = selectorDragPolygonMoveVertex
				state.polygonVertexIdx = hit.info.Index
				state.polygonStartPoints = append([]geom.Pt(nil), widget.Points...)
			case widget.Active && !widget.Closed && shift:
				state.kind = selectorDragPolygonMoveAll
				state.polygonStartPoints = append([]geom.Pt(nil), widget.Points...)
			case widget.Active && widget.Closed && (mouse.Modifiers&ModifierShift != 0):
				state.kind = selectorDragPolygonMoveAll
				state.polygonStartPoints = append([]geom.Pt(nil), widget.Points...)
			case widget.Active && widget.Closed && (mouse.Modifiers&ModifierControl != 0) && hit.info.Index >= 0:
				state.kind = selectorDragPolygonMoveVertex
				state.polygonVertexIdx = hit.info.Index
				state.polygonStartPoints = append([]geom.Pt(nil), widget.Points...)
			case widget.Active && !widget.Closed && hit.info.Index >= 0 && hit.info.Index == 0 && len(widget.Points) > 2:
				if widget.Close() {
					changed = true
				}
				if widget.Closed {
					widget.TriggerOnSelect()
				}
				state.kind = selectorDragNone
			case widget.Active && !widget.Closed:
				if widget.AppendPoint(data) {
					state.kind = selectorDragPolygonDrawing
					state.polygonPendingIdx = len(widget.Points) - 1
					state.polygonStartPoints = append([]geom.Pt(nil), widget.Points...)
					changed = true
				}
			default:
				if !widget.Active {
					widget.Clear()
				}
				if widget.AppendPoint(data) {
					state.kind = selectorDragPolygonDrawing
					state.polygonPendingIdx = len(widget.Points) - 1
					state.polygonStartPoints = append([]geom.Pt(nil), widget.Points...)
					changed = true
				}
			}
			if state.kind != selectorDragNone {
				w.draggingSelector = state
				changed = true
			}
		}
	case *core.LassoSelector:
		w.focusedSelector = widget
		if data, ok := w.pixelToDataLocked(mouse, axes); ok {
			w.draggingSelector = selectorDragState{
				kind:      selectorDragLasso,
				axes:      axes,
				lasso:     widget,
				pressData: data,
			}
			widget.Clear()
			if widget.Begin(data) {
				changed = true
			}
			changed = true
		}
	default:
		if w.blurFocusedTextLocked() {
			changed = true
		}
		w.focusedSelector = nil
	}
	w.mu.Unlock()
	if changed {
		return w.callDraw()
	}
	return nil
}

func (w *WidgetInteraction) handleMouseMove(mouse MouseEvent) error {
	if w == nil {
		return nil
	}
	hit := w.pickWidget(mouse.Event)
	axes := w.resolveAxesForEvent(mouse)
	if axes == nil {
		axes = hit.axes
	}

	w.mu.Lock()
	changed := false
	nextHovered := widgetButton(hit.widget)
	if w.hoveredButton != nextHovered {
		if w.hoveredButton != nil {
			w.hoveredButton.Hovered = false
		}
		w.hoveredButton = nextHovered
		if nextHovered != nil {
			nextHovered.Hovered = true
		}
		changed = true
	}

	if w.draggingSlider != nil {
		before := w.draggingSlider.Value
		w.setSliderValueFromPointLocked(w.draggingSlider, w.draggingSliderAxes, mouse.Position)
		if w.draggingSlider.Value != before {
			changed = true
		}
	}
	if w.draggingRangeSlider != nil {
		beforeLow, beforeHigh := w.draggingRangeSlider.Low, w.draggingRangeSlider.High
		w.setRangeSliderValueFromPointLocked(w.draggingRangeSlider, w.draggingRangeSliderAxes, mouse.Position, w.draggingRangeSliderHandle)
		if w.draggingRangeSlider.Low != beforeLow || w.draggingRangeSlider.High != beforeHigh {
			changed = true
		}
	}

	if w.draggingSelector.kind != selectorDragNone {
		if w.updateDraggingSelectorFromMouseLocked(mouse) {
			changed = true
		}
	}
	if w.updateCursorFromMouseLocked(mouse, axes) {
		changed = true
	}
	w.mu.Unlock()
	if changed {
		return w.callDraw()
	}
	return nil
}

func (w *WidgetInteraction) handleMouseRelease(mouse MouseEvent) error {
	if w == nil {
		return nil
	}
	hit := w.pickWidget(mouse.Event)
	axes := w.resolveAxesForEvent(mouse)
	if axes == nil {
		axes = hit.axes
	}

	w.mu.Lock()
	changed := false
	clickButton := (*core.Button)(nil)
	nextHovered := widgetButton(hit.widget)

	if w.pressedButton != nil {
		if w.pressedButton == nextHovered {
			clickButton = w.pressedButton
		}
		w.pressedButton.Pressed = false
		w.pressedButton = nil
		changed = true
	}

	if w.draggingSlider != nil {
		w.draggingSlider.Dragging = false
		w.draggingSlider = nil
		w.draggingSliderAxes = nil
		changed = true
	}
	if w.draggingRangeSlider != nil {
		w.draggingRangeSlider.Dragging = false
		w.draggingRangeSlider = nil
		w.draggingRangeSliderAxes = nil
		w.draggingRangeSliderHandle = 0
		changed = true
	}
	if w.draggingSelector.kind != selectorDragNone {
		if w.finalizeDraggingSelectorLocked() {
			changed = true
		}
	}
	if w.updateCursorFromMouseLocked(mouse, axes) {
		changed = true
	}

	if w.hoveredButton != nextHovered {
		if w.hoveredButton != nil {
			w.hoveredButton.Hovered = false
		}
		w.hoveredButton = nextHovered
		if nextHovered != nil {
			nextHovered.Hovered = true
		}
		changed = true
	}
	w.mu.Unlock()

	if clickButton != nil {
		clickButton.Click()
	}
	if changed {
		return w.callDraw()
	}
	return nil
}

func (w *WidgetInteraction) handleKeyPress(ev KeyEvent) error {
	if w == nil {
		return nil
	}
	key := normalizeWidgetKey(ev.Key)
	if key == "" {
		return nil
	}

	w.mu.Lock()
	if w.focusedText != nil && w.focusedText.Active {
		w.mu.Unlock()
		draw := w.handleTextKey(w.focusedText, ev, key)
		if draw {
			return w.callDraw()
		}
		return nil
	}
	if w.focusedButton != nil {
		button := w.focusedButton
		w.mu.Unlock()
		if key == "enter" || key == "space" {
			button.Click()
			return w.callDraw()
		}
		return nil
	}
	if w.focusedSlider != nil {
		w.mu.Unlock()
		draw := w.handleSliderKey(w.focusedSlider, ev, key)
		if draw {
			return w.callDraw()
		}
		return nil
	}
	if w.focusedRangeSlider != nil {
		w.mu.Unlock()
		draw := w.handleRangeSliderKey(w.focusedRangeSlider, w.focusedRangeSliderHandle, ev, key)
		if draw {
			return w.callDraw()
		}
		return nil
	}
	if w.focusedCheck != nil {
		w.mu.Unlock()
		draw := w.handleCheckKey(w.focusedCheck, w.focusedCheckIndex, ev, key)
		if draw {
			return w.callDraw()
		}
		return nil
	}
	if w.focusedRadio != nil {
		w.mu.Unlock()
		draw := w.handleRadioKey(w.focusedRadio, w.focusedRadioIndex, key)
		if draw {
			return w.callDraw()
		}
		return nil
	}
	if w.focusedSelector != nil {
		draw := w.handleSelectorKey(ev, key)
		w.mu.Unlock()
		if draw {
			return w.callDraw()
		}
		return nil
	}
	w.mu.Unlock()
	return nil
}

func (w *WidgetInteraction) handleMouseLeave(mouse MouseEvent) error {
	if w == nil {
		return nil
	}

	w.mu.Lock()
	changed := false

	if w.pressedButton != nil {
		w.pressedButton.Pressed = false
		w.pressedButton = nil
		changed = true
	}
	if w.hoveredButton != nil {
		w.hoveredButton.Hovered = false
		w.hoveredButton = nil
		changed = true
	}
	if w.draggingSlider != nil {
		w.draggingSlider.Dragging = false
		w.draggingSlider = nil
		w.draggingSliderAxes = nil
		changed = true
	}
	if w.draggingRangeSlider != nil {
		w.draggingRangeSlider.Dragging = false
		w.draggingRangeSlider = nil
		w.draggingRangeSliderAxes = nil
		w.draggingRangeSliderHandle = 0
		changed = true
	}
	if w.draggingSelector.kind != selectorDragNone {
		w.clearSelectorDragLocked()
		changed = true
	}
	if w.blurFocusedTextLocked() {
		changed = true
	}
	w.focusedSelector = nil
	if w.updateCursorFromMouseLocked(mouse, nil) {
		changed = true
	}

	w.mu.Unlock()
	if changed {
		return w.callDraw()
	}
	return nil
}

func (w *WidgetInteraction) resolveAxesForEvent(mouse MouseEvent) *Axes {
	axes := mouse.Axes
	if axes != nil {
		return axes
	}
	fig := mouse.Figure
	if fig == nil {
		fig = w.figure
	}
	if fig == nil {
		return nil
	}
	resolved, _, ok := ResolveEventTarget(fig, mouse.Position)
	if !ok {
		return nil
	}
	return resolved
}

func (w *WidgetInteraction) pixelToDataLocked(mouse MouseEvent, axes *Axes) (geom.Pt, bool) {
	if axes == nil {
		axes = w.resolveAxesForEvent(mouse)
	}
	if axes == nil {
		return geom.Pt{}, false
	}
	if mouse.HasDataPosition {
		return mouse.DataPosition, true
	}
	return axes.PixelToData(mouse.Position)
}

func (w *WidgetInteraction) updateCursorFromMouseLocked(mouse MouseEvent, axes *Axes) bool {
	if w.figure == nil && mouse.Figure == nil {
		return false
	}
	fig := w.figure
	if fig == nil {
		fig = mouse.Figure
	}
	if axes == nil {
		axes = w.resolveAxesForEvent(mouse)
	}

	changed := false
	seenMulti := map[*core.MultiCursor]bool{}
	for _, axis := range fig.Children {
		if axis == nil {
			continue
		}
		for _, art := range axesInteractionArtists(axis) {
			switch selector := art.(type) {
			case *core.Cursor:
				if axes == axis {
					data, ok := w.pixelToDataLocked(mouse, axis)
					if !ok {
						if selector.Hide() {
							changed = true
						}
						break
					}
					changed = selector.SetData(data.X, data.Y) || changed
				} else if selector.Hide() {
					changed = true
				}
			case *core.MultiCursor:
				if seenMulti[selector] {
					continue
				}
				seenMulti[selector] = true
				if axes != nil && axesInAxesList(selector.Axes, axes) {
					changed = selector.SetFigurePoint(mouse.Position) || changed
				} else {
					changed = selector.Hide() || changed
				}
			}
		}
	}
	return changed
}

func (w *WidgetInteraction) callDraw() error {
	w.mu.Lock()
	draw := w.draw
	w.mu.Unlock()
	if draw == nil {
		return nil
	}
	return draw()
}

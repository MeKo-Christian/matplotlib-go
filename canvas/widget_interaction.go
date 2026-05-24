package canvas

import (
	"math"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

// WidgetInteraction wires interactive behavior onto widget artists.
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

type selectorDragKind uint8

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

type selectorDragState struct {
	kind selectorDragKind

	axes    *Axes
	span    *core.SpanSelector
	rect    *core.RectangleSelector
	ellipse *core.EllipseSelector
	polygon *core.PolygonSelector
	lasso   *core.LassoSelector

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

var (
	clipboardMu   sync.Mutex
	clipboardText string
)

// NewWidgetInteraction creates a widget interaction helper for a figure.
func NewWidgetInteraction(fig *Figure, draw func() error) *WidgetInteraction {
	if fig == nil {
		return nil
	}
	return &WidgetInteraction{figure: fig, draw: draw}
}

// Attach installs widget event handlers on the dispatcher.
func (w *WidgetInteraction) Attach(dispatcher *Dispatcher) {
	if w == nil || dispatcher == nil {
		return
	}
	w.Detach()

	w.mu.Lock()
	w.dispatch = dispatcher
	w.connects = append(w.connects,
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

// Detach removes all widget handlers from the dispatcher.
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

func (w *WidgetInteraction) findInactiveSelector(axes *Axes) any {
	if axes == nil {
		return nil
	}
	artists := axesInteractionArtists(axes)
	for i := len(artists) - 1; i >= 0; i-- {
		switch sel := artists[i].(type) {
		case *core.SpanSelector:
			if !sel.Active {
				return sel
			}
		case *core.RectangleSelector:
			if !sel.Active {
				return sel
			}
		case *core.EllipseSelector:
			if !sel.Active {
				return sel
			}
		case *core.PolygonSelector:
			if !sel.Active {
				return sel
			}
		case *core.LassoSelector:
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

func (w *WidgetInteraction) updateDraggingSelectorFromMouseLocked(mouse MouseEvent) bool {
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
			min := state.rectStartMin
			max := state.rectStartMax
			min = geom.Pt{X: min.X + delta.X, Y: min.Y + delta.Y}
			max = geom.Pt{X: max.X + delta.X, Y: max.Y + delta.Y}
			changed = state.rect.SetBounds(min, max)
			break
		} else {
			min, max := selectorBoundsFromDrag(state.pressData, data, state.square, state.center)
			changed = state.rect.SetBounds(min, max)
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
			min := state.rectStartMin
			max := state.rectStartMax
			min = geom.Pt{X: min.X + delta.X, Y: min.Y + delta.Y}
			max = geom.Pt{X: max.X + delta.X, Y: max.Y + delta.Y}
			changed = state.ellipse.SetBounds(min, max)
			break
		} else {
			min, max := selectorBoundsFromDrag(state.pressData, data, state.square, state.center)
			changed = state.ellipse.SetBounds(min, max)
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

func (w *WidgetInteraction) handleSelectorKey(ev KeyEvent, key string) bool {
	if w.focusedSelector == nil {
		return false
	}
	axes := w.axesForSelectorLocked(w.focusedSelector)
	command := strings.ToLower(key)
	switch keyState := w.focusedSelector.(type) {
	case *core.SpanSelector:
		return w.handleSpanSelectorKey(ev, command, keyState, axes)
	case *core.RectangleSelector:
		return w.handleRectangleSelectorKey(ev, command, keyState, axes)
	case *core.EllipseSelector:
		return w.handleEllipseSelectorKey(ev, command, keyState, axes)
	case *core.PolygonSelector:
		return w.handlePolygonSelectorKey(ev, command, keyState, axes)
	case *core.LassoSelector:
		return w.handleLassoSelectorKey(ev, command, keyState)
	default:
		return false
	}
}

func (w *WidgetInteraction) handleSpanSelectorKey(ev KeyEvent, key string, selector *core.SpanSelector, axes *Axes) bool {
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
	if ev.Modifiers&ModifierControl != 0 {
		delta *= 10
	}
	changed := selector.Move(delta)
	if changed {
		selector.TriggerOnSelect()
	}
	return changed
}

func (w *WidgetInteraction) handleRectangleSelectorKey(ev KeyEvent, key string, selector *core.RectangleSelector, axes *Axes) bool {
	if selector == nil {
		return false
	}
	if key == "escape" {
		return selector.Clear()
	}
	if !selector.Active {
		return false
	}
	delta, ok := w.selectorMoveDeltaForKey(ev, key, axes)
	if !ok {
		return false
	}
	changed := selector.MoveBy(delta)
	if changed {
		selector.TriggerOnSelect()
	}
	return changed
}

func (w *WidgetInteraction) handleEllipseSelectorKey(ev KeyEvent, key string, selector *core.EllipseSelector, axes *Axes) bool {
	if selector == nil {
		return false
	}
	if key == "escape" {
		return selector.Clear()
	}
	if !selector.Active {
		return false
	}
	delta, ok := w.selectorMoveDeltaForKey(ev, key, axes)
	if !ok {
		return false
	}
	changed := selector.MoveBy(delta)
	if changed {
		selector.TriggerOnSelect()
	}
	return changed
}

func (w *WidgetInteraction) selectorMoveDeltaForKey(ev KeyEvent, key string, axes *Axes) (geom.Pt, bool) {
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
		delta.Y = delta.Y
	case "down":
		delta.X = 0
		delta.Y = -delta.Y
	}
	if ev.Modifiers&ModifierControl != 0 {
		delta.X *= 10
		delta.Y *= 10
	}
	return delta, true
}

func (w *WidgetInteraction) handlePolygonSelectorKey(ev KeyEvent, key string, selector *core.PolygonSelector, axes *Axes) bool {
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
	if ev.Modifiers&ModifierControl != 0 {
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

func (w *WidgetInteraction) handleLassoSelectorKey(_ KeyEvent, key string, selector *core.LassoSelector) bool {
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

func axesInteractionArtists(axes *core.Axes) []core.Artist {
	if axes == nil {
		return nil
	}
	if len(axes.WidgetArtists) == 0 {
		return axes.Artists
	}
	out := make([]core.Artist, 0, len(axes.Artists)+len(axes.WidgetArtists))
	out = append(out, axes.Artists...)
	out = append(out, axes.WidgetArtists...)
	return out
}

func selectorBoundsFromDrag(
	press, current geom.Pt,
	square bool,
	center bool,
) (geom.Pt, geom.Pt) {
	min := press
	max := current
	if center {
		delta := geom.Pt{
			X: current.X - press.X,
			Y: current.Y - press.Y,
		}
		min = geom.Pt{X: press.X - delta.X, Y: press.Y - delta.Y}
		max = geom.Pt{X: press.X + delta.X, Y: press.Y + delta.Y}
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
		max = geom.Pt{
			X: press.X + sx,
			Y: press.Y + sy,
		}
		if center {
			min = geom.Pt{
				X: press.X - sx,
				Y: press.Y - sy,
			}
		}
	}
	return normalizedMinMax(min, max)
}

func normalizedMinMax(min, max geom.Pt) (geom.Pt, geom.Pt) {
	if min.X <= max.X && min.Y <= max.Y {
		return min, max
	}
	return geom.Pt{X: math.Min(min.X, max.X), Y: math.Min(min.Y, max.Y)}, geom.Pt{X: math.Max(min.X, max.X), Y: math.Max(min.Y, max.Y)}
}

func (w *WidgetInteraction) handleTextKey(tb *core.TextBox, ev KeyEvent, key string) bool {
	if tb == nil {
		return false
	}

	raw := ev.Key
	if raw == "" {
		return false
	}

	command := strings.ToLower(key)
	shift := ev.Modifiers&ModifierShift != 0
	ctrlOrMeta := ev.Modifiers&(ModifierControl|ModifierMeta) != 0
	word := ev.Modifiers&(ModifierControl|ModifierMeta) != 0
	before := tb.Value
	beforeSelStart, beforeSelEnd := tb.Selection()

	switch command {
	case "left":
		tb.MoveCaretLeft(word, shift)
	case "right":
		tb.MoveCaretRight(word, shift)
	case "home":
		tb.MoveCaretToStart(shift)
	case "end":
		tb.MoveCaretToEnd(shift)
	case "backspace":
		tb.Backspace()
	case "delete", "del":
		tb.Delete()
	case "enter":
		tb.Submit()
		return false
	case "escape":
		tb.Cancel()
		return false
	case "a":
		if ctrlOrMeta {
			tb.SelectAll()
			start, end := tb.Selection()
			return beforeSelStart != start || beforeSelEnd != end
		}
		tb.InsertText(raw)
	case "c":
		if !ctrlOrMeta {
			tb.InsertText(raw)
			break
		}
		text := tb.SelectedText()
		if text == "" {
			return false
		}
		setClipboard(text)
	case "x":
		if !ctrlOrMeta {
			tb.InsertText(raw)
			break
		}
		text := tb.SelectedText()
		if text == "" {
			return false
		}
		setClipboard(text)
		tb.Delete()
	case "v":
		if !ctrlOrMeta {
			tb.InsertText(raw)
			break
		}
		tb.InsertText(getClipboard())
		return tb.Value != before
	case " ":
		tb.InsertText(" ")
	default:
		if ctrlOrMeta {
			return false
		}
		if len(raw) == 1 {
			tb.InsertText(raw)
			return tb.Value != before
		}
		return false
	}
	return tb.Value != before
}

func (w *WidgetInteraction) handleSliderKey(slider *core.Slider, ev KeyEvent, key string) bool {
	if slider == nil {
		return false
	}
	step := slider.Step
	if step <= 0 {
		step = 1
	}
	delta := 0.0
	switch key {
	case "left", "down":
		delta = -step
	case "right", "up":
		delta = step
	default:
		return false
	}
	if ev.Modifiers&ModifierControl != 0 {
		delta *= 10
	}
	before := slider.Value
	slider.SetValue(slider.Value + delta)
	return slider.Value != before
}

func (w *WidgetInteraction) handleRangeSliderKey(slider *core.RangeSlider, handle int, ev KeyEvent, key string) bool {
	if slider == nil {
		return false
	}
	step := slider.Step
	if step <= 0 {
		step = 1
	}
	delta := 0.0
	switch key {
	case "left", "down":
		delta = -step
	case "right", "up":
		delta = step
	default:
		return false
	}
	if ev.Modifiers&ModifierControl != 0 {
		delta *= 10
	}
	beforeLow, beforeHigh := slider.Low, slider.High
	if handle <= 0 {
		slider.SetLow(slider.Low + delta)
	} else {
		slider.SetHigh(slider.High + delta)
	}
	return slider.Low != beforeLow || slider.High != beforeHigh
}

func (w *WidgetInteraction) handleCheckKey(checks *core.CheckButtons, focusedIndex int, ev KeyEvent, key string) bool {
	if checks == nil || len(checks.Labels) == 0 {
		return false
	}
	command := strings.ToLower(key)
	focusedIndex = clampInt(focusedIndex, 0, len(checks.Labels)-1)
	switch command {
	case "up", "left":
		if focusedIndex > 0 {
			focusedIndex--
		} else {
			focusedIndex = len(checks.Labels) - 1
		}
		w.focusedCheckIndex = focusedIndex
		return true
	case "down", "right":
		if focusedIndex < len(checks.Labels)-1 {
			focusedIndex++
		} else {
			focusedIndex = 0
		}
		w.focusedCheckIndex = focusedIndex
		return true
	case "space", "enter":
		_ = ev
		if focusedIndex < 0 {
			focusedIndex = 0
		}
		before := checks.Values[focusedIndex]
		checks.Toggle(focusedIndex)
		w.focusedCheckIndex = focusedIndex
		return checks.Values[focusedIndex] != before
	}
	return false
}

func (w *WidgetInteraction) handleRadioKey(radios *core.RadioButtons, focusedIndex int, key string) bool {
	if radios == nil || len(radios.Labels) == 0 {
		return false
	}
	focusedIndex = clampInt(focusedIndex, 0, len(radios.Labels)-1)

	switch strings.ToLower(key) {
	case "up", "left":
		if focusedIndex > 0 {
			focusedIndex--
		} else {
			focusedIndex = len(radios.Labels) - 1
		}
		w.focusedRadioIndex = focusedIndex
		radios.SetActive(focusedIndex)
		return true
	case "down", "right":
		if focusedIndex < len(radios.Labels)-1 {
			focusedIndex++
		} else {
			focusedIndex = 0
		}
		w.focusedRadioIndex = focusedIndex
		radios.SetActive(focusedIndex)
		return true
	case "space", "enter":
		w.focusedRadioIndex = focusedIndex
		radios.SetActive(focusedIndex)
		return true
	}
	return false
}

func (w *WidgetInteraction) setSliderValueFromPointLocked(slider *core.Slider, ax *Axes, position geom.Pt) {
	if slider == nil || ax == nil {
		return
	}
	ctx := core.AxesDrawContext(ax, w.figure)
	if ctx == nil {
		return
	}
	panel := widgetInsetRect(ctx.Clip, 4)
	if panel.W() <= 0 {
		return
	}
	track := geom.Rect{
		Min: geom.Pt{X: panel.Min.X + 14, Y: panel.Max.Y - 26},
		Max: geom.Pt{X: panel.Max.X - 14, Y: panel.Max.Y - 14},
	}
	if track.W() <= 0 {
		return
	}
	v := slider.Min + (slider.Max-slider.Min)*((position.X-track.Min.X)/track.W())
	slider.SetValue(v)
}

func (w *WidgetInteraction) setRangeSliderValueFromPointLocked(slider *core.RangeSlider, ax *Axes, position geom.Pt, handle int) {
	if slider == nil || ax == nil {
		return
	}
	ctx := core.AxesDrawContext(ax, w.figure)
	if ctx == nil {
		return
	}
	panel := widgetInsetRect(ctx.Clip, 4)
	if panel.W() <= 0 {
		return
	}
	track := geom.Rect{
		Min: geom.Pt{X: panel.Min.X + 14, Y: panel.Max.Y - 26},
		Max: geom.Pt{X: panel.Max.X - 14, Y: panel.Max.Y - 14},
	}
	if track.W() <= 0 {
		return
	}
	v := slider.Min + (slider.Max-slider.Min)*((position.X-track.Min.X)/track.W())
	if handle <= 0 {
		slider.SetLow(v)
		return
	}
	slider.SetHigh(v)
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

func (w *WidgetInteraction) blurFocusedTextLocked() bool {
	if w.focusedText == nil || !w.focusedText.Active {
		return false
	}
	w.focusedText.Active = false
	w.focusedText = nil
	return true
}

type widgetPick struct {
	axes   *Axes
	widget any
	info   core.PickInfo
}

func (w *WidgetInteraction) pickWidget(ev Event) widgetPick {
	fig := ev.Figure
	if fig == nil {
		fig = w.figure
	}
	if fig == nil {
		return widgetPick{}
	}
	hits := Pick(fig, ev.Position)
	for _, hit := range hits {
		switch hit.Artist.(type) {
		case *core.Button, *core.Slider, *core.RangeSlider, *core.CheckButtons, *core.RadioButtons, *core.TextBox,
			*core.SpanSelector, *core.RectangleSelector, *core.EllipseSelector, *core.PolygonSelector, *core.LassoSelector,
			*core.Cursor, *core.MultiCursor:
			return widgetPick{
				axes:   hit.Axes,
				widget: hit.Artist,
				info:   hit.Info,
			}
		}
	}
	return widgetPick{}
}

func axesInAxesList(axesList []*Axes, axes *Axes) bool {
	for _, candidate := range axesList {
		if candidate == axes {
			return true
		}
	}
	return false
}

func widgetButton(v any) *core.Button {
	widget, ok := v.(*core.Button)
	if ok && widget.Enabled {
		return widget
	}
	return nil
}

func cursorIndexFromTextPoint(tb *core.TextBox, axes *Axes, fig *Figure, point geom.Pt) int {
	if tb == nil || axes == nil || fig == nil {
		return 0
	}
	ctx := core.AxesDrawContext(axes, fig)
	panel := widgetInsetRect(ctx.Clip, 4)
	input := geom.Rect{
		Min: geom.Pt{X: panel.Min.X + 4, Y: panel.Min.Y + 30},
		Max: geom.Pt{X: panel.Max.X - 4, Y: panel.Max.Y - 8},
	}
	if input.W() <= 0 {
		return 0
	}
	size := widgetFontSize(tb.FontSize, ctx)
	cell := size * 0.42
	if cell <= 0 {
		cell = 1
	}
	index := int(math.Round((point.X - (input.Min.X + 12)) / cell))
	if index < 0 {
		return 0
	}
	runes := []rune(tb.Value)
	if index > len(runes) {
		return len(runes)
	}
	return index
}

func normalizeWidgetKey(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "arrowleft", "left":
		return "left"
	case "arrowright", "right":
		return "right"
	case "arrowup", "up":
		return "up"
	case "arrowdown", "down":
		return "down"
	case "pageup":
		return "up"
	case "pagedown":
		return "down"
	case "home":
		return "home"
	case "end":
		return "end"
	case "escape":
		return "escape"
	case "return":
		return "enter"
	case "spacebar":
		return "space"
	default:
		return key
	}
}

func setClipboard(value string) {
	clipboardMu.Lock()
	defer clipboardMu.Unlock()
	clipboardText = value
}

func getClipboard() string {
	clipboardMu.Lock()
	defer clipboardMu.Unlock()
	return clipboardText
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func widgetInsetRect(rect geom.Rect, pad float64) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: rect.Min.X + pad, Y: rect.Min.Y + pad},
		Max: geom.Pt{X: rect.Max.X - pad, Y: rect.Max.Y - pad},
	}
}

func widgetFontSize(size float64, ctx *core.DrawContext) float64 {
	if size > 0 {
		return size
	}
	if ctx != nil && ctx.RC.FontSize > 0 {
		return ctx.RC.FontSize
	}
	return 12
}

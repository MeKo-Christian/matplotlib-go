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

	draggingSlider     *core.Slider
	draggingSliderAxes *Axes

	focusedText       *core.TextBox
	focusedSlider     *core.Slider
	focusedCheck      *core.CheckButtons
	focusedCheckIndex int
	focusedRadio      *core.RadioButtons
	focusedRadioIndex int
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

	w.mu.Lock()
	changed := false

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
		w.focusedSlider = nil
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

	w.focusedSlider = nil
	w.focusedCheck = nil
	w.focusedCheckIndex = 0
	w.focusedRadio = nil
	w.focusedRadioIndex = 0
	w.focusedText = nil

	switch widget := hit.widget.(type) {
	case *core.Button:
		if widget.Enabled {
			w.pressedButton = widget
			widget.Pressed = true
			widget.Hovered = true
			changed = true
		}
		if w.blurFocusedTextLocked() {
			changed = true
		}
		w.focusedText = nil
		// click callback waits until release.
	case *core.Slider:
		w.focusedText = nil
		w.focusedSlider = widget
		if widget.Enabled {
			w.draggingSlider = widget
			w.draggingSliderAxes = hit.axes
			widget.Dragging = true
			before := widget.Value
			w.setSliderValueFromPointLocked(widget, hit.axes, mouse.Position)
			changed = changed || widget.Value != before
		}
	case *core.CheckButtons:
		w.focusedText = nil
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
		changed = changed || widget.Values[w.focusedCheckIndex] != before
		changed = true
		if w.blurFocusedTextLocked() {
			changed = true
		}
	case *core.RadioButtons:
		w.focusedText = nil
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
	case *core.TextBox:
		w.focusedText = widget
		w.focusedSlider = nil
		w.focusedCheck = nil
		w.focusedCheckIndex = 0
		w.focusedRadio = nil
		w.focusedRadioIndex = 0
		if !widget.Active {
			w.blurFocusedTextLocked()
			widget.Active = true
			changed = true
		}
		if hit.axes != nil {
			index := cursorIndexFromTextPoint(widget, hit.axes, w.figure, mouse.Position)
			beforeA, beforeB := widget.Selection()
			widget.SetSelection(index, index)
			if beforeA != index || beforeB != index {
				changed = true
			}
		}
	default:
		if w.blurFocusedTextLocked() {
			changed = true
		}
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
	if w.focusedSlider != nil {
		w.mu.Unlock()
		draw := w.handleSliderKey(w.focusedSlider, ev, key)
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
	w.mu.Unlock()
	return nil
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
		case *core.Button, *core.Slider, *core.CheckButtons, *core.RadioButtons, *core.TextBox:
			return widgetPick{
				axes:   hit.Axes,
				widget: hit.Artist,
				info:   hit.Info,
			}
		}
	}
	return widgetPick{}
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

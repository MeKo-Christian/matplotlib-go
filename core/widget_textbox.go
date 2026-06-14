package core

import (
	"math"
	"unicode"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// TextBoxSubmitCallback receives the text after submit/cancel events.
type TextBoxSubmitCallback func(*TextBox, string)

// TextBoxChangeCallback receives the new text after each value change.
type TextBoxChangeCallback func(*TextBox, string)

// TextBoxOptions configures a TextBox widget artist.
type TextBoxOptions struct {
	FaceColor   render.Color
	EdgeColor   render.Color
	TextColor   render.Color
	Placeholder string
	FontSize    float64
	Active      *bool
}

// TextBox draws a static text-entry control.
type TextBox struct {
	Label       string
	Value       string
	Placeholder string
	FaceColor   render.Color
	EdgeColor   render.Color
	TextColor   render.Color
	FontSize    float64
	Active      bool
	caret       int
	selection   [2]int

	onSubmit widgetCallbackRegistry[TextBoxSubmitCallback]
	onCancel widgetCallbackRegistry[TextBoxSubmitCallback]
	onChange widgetCallbackRegistry[TextBoxChangeCallback]

	z float64
}

// TextBox adds a text-box widget artist to the axes.
func (a *Axes) TextBox(label, value string, opts ...TextBoxOptions) *TextBox {
	if a == nil {
		return nil
	}
	defaults := widgetDefaultsForAxes(a)
	cfg := TextBoxOptions{
		FaceColor: defaults.TextBoxFace,
		EdgeColor: defaults.TextBoxEdge,
		TextColor: defaults.Text,
	}
	if len(opts) > 0 {
		cfg = mergeTextBoxOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	w := &TextBox{
		Label:       label,
		Value:       value,
		Placeholder: cfg.Placeholder,
		FaceColor:   cfg.FaceColor,
		EdgeColor:   cfg.EdgeColor,
		TextColor:   cfg.TextColor,
		FontSize:    cfg.FontSize,
		selection:   [2]int{0, 0},
		Active:      boolValue(cfg.Active, false),
		z:           1200,
	}
	if value != "" {
		w.caret = len([]rune(value))
		w.selection[0] = w.caret
		w.selection[1] = w.caret
	}
	a.AddWidget(w)
	return w
}

func (t *TextBox) OnSubmit(cb TextBoxSubmitCallback) WidgetCallbackID {
	if t == nil || any(cb) == nil {
		return 0
	}
	return t.onSubmit.add(cb)
}

func (t *TextBox) OnCancel(cb TextBoxSubmitCallback) WidgetCallbackID {
	if t == nil || any(cb) == nil {
		return 0
	}
	return t.onCancel.add(cb)
}

func (t *TextBox) OnChange(cb TextBoxChangeCallback) WidgetCallbackID {
	if t == nil || any(cb) == nil {
		return 0
	}
	return t.onChange.add(cb)
}

func (t *TextBox) RemoveOnSubmit(id WidgetCallbackID) {
	if t == nil {
		return
	}
	t.onSubmit.remove(id)
}

func (t *TextBox) RemoveOnCancel(id WidgetCallbackID) {
	if t == nil {
		return
	}
	t.onCancel.remove(id)
}

func (t *TextBox) RemoveOnChange(id WidgetCallbackID) {
	if t == nil {
		return
	}
	t.onChange.remove(id)
}

func (t *TextBox) triggerSubmit() {
	if t == nil {
		return
	}
	t.onSubmit.each(func(cb TextBoxSubmitCallback) { cb(t, t.Value) })
}

func (t *TextBox) triggerCancel() {
	if t == nil {
		return
	}
	t.onCancel.each(func(cb TextBoxSubmitCallback) { cb(t, t.Value) })
}

func (t *TextBox) triggerChange() {
	if t == nil {
		return
	}
	t.onChange.each(func(cb TextBoxChangeCallback) { cb(t, t.Value) })
}

// Submit emits submit callbacks with the current text.
func (t *TextBox) Submit() {
	if t == nil {
		return
	}
	t.triggerSubmit()
}

// Cancel emits cancel callbacks with the current text.
func (t *TextBox) Cancel() {
	if t == nil {
		return
	}
	t.triggerCancel()
}

// Activate marks the text box as focused.
func (t *TextBox) Activate(active bool) {
	if t == nil {
		return
	}
	t.Active = active
}

// SetValue updates the text content and places the caret at the end.
func (t *TextBox) SetValue(value string) {
	if t == nil {
		return
	}
	if t.Value == value {
		return
	}
	t.Value = value
	count := runeCount(value)
	t.caret = count
	t.selection = [2]int{count, count}
	t.triggerChange()
}

// SetCaret sets the insertion point and collapses the selection.
func (t *TextBox) SetCaret(index int) {
	if t == nil {
		return
	}
	index = clampTextIndex(t.Value, index)
	t.caret = index
	t.selection = [2]int{index, index}
}

// SetSelection sets the highlighted range in rune offsets. Any order is
// normalized so the first index is less than or equal to the second.
func (t *TextBox) SetSelection(start, end int) {
	if t == nil {
		return
	}
	start = clampTextIndex(t.Value, start)
	end = clampTextIndex(t.Value, end)
	if start > end {
		start, end = end, start
	}
	t.selection = [2]int{start, end}
	t.caret = end
}

// SelectAll marks the entire value as selected.
func (t *TextBox) SelectAll() {
	if t == nil {
		return
	}
	count := runeCount(t.Value)
	t.selection = [2]int{0, count}
	t.caret = count
}

// MoveCaretLeft moves the caret left by one rune or one word.
func (t *TextBox) MoveCaretLeft(word, extend bool) {
	if t == nil {
		return
	}
	anchor := t.caret
	if !extend {
		anchor = -1
	}
	n := runeCount(t.Value)
	if n == 0 {
		return
	}
	next := t.caret
	if word {
		next = moveTextCaretWordLeft([]rune(t.Value), t.caret)
	} else if next > 0 {
		next--
	}
	t.SetCaret(next)
	if extend {
		start := t.caret
		if anchor >= 0 {
			start = anchor
		}
		t.selection = [2]int{start, t.caret}
		if t.selection[0] > t.selection[1] {
			t.selection[0], t.selection[1] = t.selection[1], t.selection[0]
		}
	}
}

// MoveCaretRight moves the caret right by one rune or one word.
func (t *TextBox) MoveCaretRight(word, extend bool) {
	if t == nil {
		return
	}
	anchor := t.caret
	if !extend {
		anchor = -1
	}
	n := runeCount(t.Value)
	next := t.caret
	if word {
		next = moveTextCaretWordRight([]rune(t.Value), t.caret)
	} else if next < n {
		next++
	}
	t.SetCaret(next)
	if extend {
		start := t.caret
		if anchor >= 0 {
			start = anchor
		}
		t.selection = [2]int{start, t.caret}
		if t.selection[0] > t.selection[1] {
			t.selection[0], t.selection[1] = t.selection[1], t.selection[0]
		}
	}
}

// MoveCaretToStart moves the caret to the beginning of the text.
func (t *TextBox) MoveCaretToStart(extend bool) {
	if t == nil {
		return
	}
	anchor := t.caret
	if !extend {
		anchor = 0
	}
	t.SetCaret(0)
	if extend {
		t.selection = [2]int{anchor, 0}
		if t.selection[0] > t.selection[1] {
			t.selection[0], t.selection[1] = t.selection[1], t.selection[0]
		}
	}
}

// MoveCaretToEnd moves the caret to the end of the text.
func (t *TextBox) MoveCaretToEnd(extend bool) {
	if t == nil {
		return
	}
	count := runeCount(t.Value)
	anchor := t.caret
	if !extend {
		anchor = count
	}
	t.SetCaret(count)
	if extend {
		t.selection = [2]int{anchor, t.caret}
		if t.selection[0] > t.selection[1] {
			t.selection[0], t.selection[1] = t.selection[1], t.selection[0]
		}
	}
}

// InsertText replaces the current selection (or caret position) with text.
func (t *TextBox) InsertText(value string) {
	if t == nil {
		return
	}
	if value == "" {
		return
	}
	text := []rune(t.Value)
	insert := []rune(value)
	start, end := t.selection[0], t.selection[1]
	if start > end {
		start, end = end, start
	}
	next := make([]rune, 0, len(text)-maxInt(end-start, 0)+len(insert))
	next = append(next, text[:start]...)
	next = append(next, insert...)
	next = append(next, text[end:]...)
	t.Value = string(next)
	n := start + len(insert)
	t.caret = n
	t.selection = [2]int{n, n}
	t.triggerChange()
}

// Backspace deletes the previous selection or rune.
func (t *TextBox) Backspace() {
	if t == nil {
		return
	}
	start, end := t.selection[0], t.selection[1]
	if start > end {
		start, end = end, start
	}
	text := []rune(t.Value)
	if start != end {
		text = append(text[:start], text[end:]...)
		t.Value = string(text)
		t.caret = start
		t.selection = [2]int{start, start}
		t.triggerChange()
		return
	}
	if t.caret == 0 {
		return
	}
	text = append(text[:t.caret-1], text[t.caret:]...)
	t.Value = string(text)
	t.caret--
	t.selection = [2]int{t.caret, t.caret}
	t.triggerChange()
}

// Delete removes the current selection or the rune after the caret.
func (t *TextBox) Delete() {
	if t == nil {
		return
	}
	start, end := t.selection[0], t.selection[1]
	if start > end {
		start, end = end, start
	}
	text := []rune(t.Value)
	if start != end {
		text = append(text[:start], text[end:]...)
		t.Value = string(text)
		t.caret = start
		t.selection = [2]int{start, start}
		t.triggerChange()
		return
	}
	if t.caret >= len(text) {
		return
	}
	text = append(text[:t.caret], text[t.caret+1:]...)
	t.Value = string(text)
	t.selection = [2]int{t.caret, t.caret}
	t.triggerChange()
}

// SelectedText returns the highlighted value as a substring.
func (t *TextBox) SelectedText() string {
	if t == nil {
		return ""
	}
	start, end := t.selection[0], t.selection[1]
	if start > end {
		start, end = end, start
	}
	text := []rune(t.Value)
	if start < 0 || end < 0 || start >= len(text) || end > len(text) || start >= end {
		return ""
	}
	return string(text[start:end])
}

func (t *TextBox) Selection() (int, int) {
	if t == nil {
		return 0, 0
	}
	return t.selection[0], t.selection[1]
}

func (t *TextBox) Caret() int {
	if t == nil {
		return 0
	}
	return t.caret
}

func (t *TextBox) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if t == nil || ctx == nil {
		return false, PickInfo{}
	}
	return t.Bounds(ctx).Contains(p), PickInfo{}
}

func (t *TextBox) Draw(r render.Renderer, ctx *DrawContext) {
	if t == nil || r == nil || ctx == nil {
		return
	}
	defaults := widgetDefaultsForRC(ctx.RC)
	panel := widgetStyledPanelRect(ctx.Clip, defaults.TextBoxPanelPad)
	drawWidgetPanel(r, panel, render.Color{A: 0}, render.Color{A: 0}, 0, 0)
	fontSize := resolvedFontSize(t.FontSize, ctx)
	labelAnchor := geom.Pt{
		X: widgetResolvedCoord(panel.Min.X, panel.Max.X, defaults.TextBoxLabelX),
		Y: widgetResolvedCoord(panel.Min.Y, panel.Max.Y, defaults.TextBoxLabelY),
	}
	drawWidgetText(r, ctx, labelAnchor, t.Label, fontSize, t.TextColor, defaults.TextBoxLabelAlign, defaults.TextBoxLabelVAlign)

	input := widgetTextBoxInputRect(panel, defaults)
	edge := t.EdgeColor
	if t.Active && defaults.TextBoxActiveEdgeBlend > 0 {
		edge = mixColor(edge, render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1}, defaults.TextBoxActiveEdgeBlend)
	}
	drawWidgetPanel(r, input, t.FaceColor, edge, defaults.TextBoxLineWidth, defaults.TextBoxRadius)

	display := t.Value
	displayColor := t.TextColor
	if display == "" {
		display = t.Placeholder
		displayColor = mixColor(t.TextColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.45)
	}
	textAnchor := widgetTextBoxTextAnchor(input, defaults)
	drawWidgetText(r, ctx, textAnchor, display, fontSize, displayColor, defaults.TextBoxTextAlign, defaults.TextBoxTextVAlign)
	if t.Active {
		caretIndex := clampInt(t.caret, 0, len([]rune(t.Value)))
		caretX := textAnchor.X + fontSize*0.42*float64(caretIndex)
		if caretX > input.Max.X {
			caretX = input.Max.X
		}
		if caretX < input.Min.X {
			caretX = input.Min.X
		}
		r.Path(pixelLinePath(
			geom.Pt{X: caretX, Y: widgetResolvedCoord(input.Min.Y, input.Max.Y, defaults.TextBoxCaretYMin)},
			geom.Pt{X: caretX, Y: widgetResolvedCoord(input.Min.Y, input.Max.Y, defaults.TextBoxCaretYMax)},
		), &render.Paint{
			Stroke:    edge,
			LineWidth: 1.2,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		})
	}
}

func (t *TextBox) Bounds(ctx *DrawContext) geom.Rect {
	if t == nil || ctx == nil {
		return geom.Rect{}
	}
	defaults := widgetDefaultsForRC(ctx.RC)
	return widgetStyledPanelRect(ctx.Clip, defaults.TextBoxPanelPad)
}

// CaretForPoint maps a figure-pixel point to a text insertion index using the
// same visual-style text anchor geometry used for drawing.
func (t *TextBox) CaretForPoint(p geom.Pt, ctx *DrawContext) int {
	if t == nil || ctx == nil {
		return 0
	}
	defaults := widgetDefaultsForRC(ctx.RC)
	panel := widgetStyledPanelRect(ctx.Clip, defaults.TextBoxPanelPad)
	input := widgetTextBoxInputRect(panel, defaults)
	if input.W() <= 0 {
		return 0
	}
	fontSize := resolvedFontSize(t.FontSize, ctx)
	cell := fontSize * 0.42
	if cell <= 0 {
		cell = 1
	}
	textAnchor := widgetTextBoxTextAnchor(input, defaults)
	index := int(math.Round((p.X - textAnchor.X) / cell))
	return clampTextIndex(t.Value, index)
}

func (t *TextBox) Z() float64   { return t.z }
func (t *TextBox) WidgetLayer() {}

func clampTextIndex(s string, index int) int {
	max := runeCount(s)
	if index < 0 {
		return 0
	}
	if index > max {
		return max
	}
	return index
}

func runeCount(s string) int {
	return len([]rune(s))
}

func isTextWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func moveTextCaretWordLeft(runes []rune, caret int) int {
	if caret <= 0 {
		return 0
	}
	i := caret
	for i > 0 && isTextWhitespace(runes[i-1]) {
		i--
	}
	for i > 0 && !isTextWhitespace(runes[i-1]) {
		i--
	}
	return i
}

func moveTextCaretWordRight(runes []rune, caret int) int {
	i := caret
	for i < len(runes) && isTextWhitespace(runes[i]) {
		i++
	}
	for i < len(runes) && !isTextWhitespace(runes[i]) {
		i++
	}
	return i
}

func mergeTextBoxOptions(base, override TextBoxOptions) TextBoxOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.Placeholder != "" {
		base.Placeholder = override.Placeholder
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	if override.Active != nil {
		base.Active = override.Active
	}
	return base
}

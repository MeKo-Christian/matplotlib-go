package core

import (
	"fmt"
	"math"
	"strings"
	"unicode"

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

// ButtonCallback receives an active button widget.
type ButtonCallback func(*Button)

// SliderCallback receives the active slider and updated value.
type SliderCallback func(*Slider, float64)

// RangeSliderCallback receives the active range slider and updated low/high values.
type RangeSliderCallback func(*RangeSlider, float64, float64)

// CheckButtonsCallback receives the active check button, changed index and
// state after the change.
type CheckButtonsCallback func(*CheckButtons, int, bool)

// RadioButtonsCallback receives the active radio widget and selected index.
type RadioButtonsCallback func(*RadioButtons, int)

// TextBoxSubmitCallback receives the text after submit/cancel events.
type TextBoxSubmitCallback func(*TextBox, string)

// TextBoxChangeCallback receives the new text after each value change.
type TextBoxChangeCallback func(*TextBox, string)

// ButtonOptions configures a Button widget artist.
type ButtonOptions struct {
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	Pressed   *bool
	Disabled  *bool
	FontSize  float64
}

// SliderOptions configures a Slider widget artist.
type SliderOptions struct {
	FaceColor   render.Color
	TrackColor  render.Color
	FillColor   render.Color
	HandleColor render.Color
	TextColor   render.Color
	ValueStep   *float64
	ValueFormat *string
	FontSize    float64
}

// RangeSliderOptions configures a RangeSlider widget artist.
type RangeSliderOptions = SliderOptions

// CheckButtonsOptions configures a CheckButtons widget artist.
type CheckButtonsOptions struct {
	FaceColor  render.Color
	EdgeColor  render.Color
	TextColor  render.Color
	CheckColor render.Color
	FontSize   float64
}

// RadioButtonsOptions configures a RadioButtons widget artist.
type RadioButtonsOptions struct {
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	DotColor  render.Color
	FontSize  float64
}

// TextBoxOptions configures a TextBox widget artist.
type TextBoxOptions struct {
	FaceColor   render.Color
	EdgeColor   render.Color
	TextColor   render.Color
	Placeholder string
	FontSize    float64
	Active      *bool
}

// Button draws a static button-style control inside its owning axes.
type Button struct {
	Label     string
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	Pressed   bool
	Enabled   bool
	FontSize  float64
	Hovered   bool

	onClicked widgetCallbackRegistry[ButtonCallback]

	z float64
}

// Slider draws a static slider control inside its owning axes.
type Slider struct {
	Label       string
	Min         float64
	Max         float64
	Enabled     bool
	Value       float64
	FaceColor   render.Color
	TrackColor  render.Color
	FillColor   render.Color
	HandleColor render.Color
	TextColor   render.Color
	Step        float64
	ValueFormat string
	FontSize    float64
	Dragging    bool

	onChanged widgetCallbackRegistry[SliderCallback]

	z float64
}

// RangeSlider draws a two-handle slider control inside its owning axes.
type RangeSlider struct {
	Label       string
	Min         float64
	Max         float64
	Enabled     bool
	Low         float64
	High        float64
	FaceColor   render.Color
	TrackColor  render.Color
	FillColor   render.Color
	HandleColor render.Color
	TextColor   render.Color
	Step        float64
	ValueFormat string
	FontSize    float64
	Dragging    bool

	onChanged widgetCallbackRegistry[RangeSliderCallback]

	z float64
}

// CheckButtons draws a static checklist-style control.
type CheckButtons struct {
	Labels     []string
	Values     []bool
	FaceColor  render.Color
	EdgeColor  render.Color
	TextColor  render.Color
	CheckColor render.Color
	FontSize   float64
	focus      int

	onChanged widgetCallbackRegistry[CheckButtonsCallback]

	z float64
}

// RadioButtons draws a static radio-button control.
type RadioButtons struct {
	Labels    []string
	Active    int
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	DotColor  render.Color
	FontSize  float64
	focus     int

	onChanged widgetCallbackRegistry[RadioButtonsCallback]

	z float64
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

// Button adds a button widget artist to the axes.
func (a *Axes) Button(label string, opts ...ButtonOptions) *Button {
	if a == nil {
		return nil
	}
	cfg := ButtonOptions{
		FaceColor: render.Color{R: 0.94, G: 0.95, B: 0.97, A: 1},
		EdgeColor: render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
		TextColor: render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeButtonOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	enabled := true
	if cfg.Disabled != nil {
		enabled = !*cfg.Disabled
	}
	w := &Button{
		Label:     label,
		FaceColor: cfg.FaceColor,
		EdgeColor: cfg.EdgeColor,
		TextColor: cfg.TextColor,
		Pressed:   boolValue(cfg.Pressed, false),
		Enabled:   enabled,
		FontSize:  cfg.FontSize,
		z:         1200,
	}
	a.AddWidget(w)
	return w
}

// Slider adds a slider widget artist to the axes.
func (a *Axes) Slider(label string, min, max, value float64, opts ...SliderOptions) *Slider {
	if a == nil {
		return nil
	}
	cfg := SliderOptions{
		FaceColor:   render.Color{R: 0.96, G: 0.97, B: 0.98, A: 1},
		TrackColor:  render.Color{R: 0.83, G: 0.85, B: 0.89, A: 1},
		FillColor:   render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		HandleColor: render.Color{R: 0.09, G: 0.18, B: 0.34, A: 1},
		TextColor:   render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeSliderOptions(cfg, opts[0])
	}
	step := (max - min) / 100
	if cfg.ValueStep != nil && *cfg.ValueStep > 0 {
		step = *cfg.ValueStep
	}
	if step == 0 {
		step = 1
	}
	valueFormat := "%.2f"
	if cfg.ValueFormat != nil && strings.TrimSpace(*cfg.ValueFormat) != "" {
		valueFormat = *cfg.ValueFormat
	}
	prepareWidgetAxes(a)
	w := &Slider{
		Label:       label,
		Min:         min,
		Max:         max,
		Enabled:     true,
		Value:       clampSliderValue(min, max, value),
		FaceColor:   cfg.FaceColor,
		TrackColor:  cfg.TrackColor,
		FillColor:   cfg.FillColor,
		HandleColor: cfg.HandleColor,
		TextColor:   cfg.TextColor,
		Step:        math.Abs(step),
		ValueFormat: valueFormat,
		FontSize:    cfg.FontSize,
		z:           1200,
	}
	a.AddWidget(w)
	return w
}

// RangeSlider adds a two-handle range slider widget artist to the axes.
func (a *Axes) RangeSlider(label string, min, max, low, high float64, opts ...RangeSliderOptions) *RangeSlider {
	if a == nil {
		return nil
	}
	cfg := RangeSliderOptions{
		FaceColor:   render.Color{R: 0.96, G: 0.97, B: 0.98, A: 1},
		TrackColor:  render.Color{R: 0.83, G: 0.85, B: 0.89, A: 1},
		FillColor:   render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		HandleColor: render.Color{R: 0.09, G: 0.18, B: 0.34, A: 1},
		TextColor:   render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeSliderOptions(cfg, opts[0])
	}
	step := (max - min) / 100
	if cfg.ValueStep != nil && *cfg.ValueStep > 0 {
		step = *cfg.ValueStep
	}
	if step == 0 {
		step = 1
	}
	valueFormat := "%.2f"
	if cfg.ValueFormat != nil && strings.TrimSpace(*cfg.ValueFormat) != "" {
		valueFormat = *cfg.ValueFormat
	}
	low = clampSliderValue(min, max, low)
	high = clampSliderValue(min, max, high)
	if low > high {
		low, high = high, low
	}
	prepareWidgetAxes(a)
	w := &RangeSlider{
		Label:       label,
		Min:         min,
		Max:         max,
		Enabled:     true,
		Low:         low,
		High:        high,
		FaceColor:   cfg.FaceColor,
		TrackColor:  cfg.TrackColor,
		FillColor:   cfg.FillColor,
		HandleColor: cfg.HandleColor,
		TextColor:   cfg.TextColor,
		Step:        math.Abs(step),
		ValueFormat: valueFormat,
		FontSize:    cfg.FontSize,
		z:           1200,
	}
	a.AddWidget(w)
	return w
}

// CheckButtons adds a check-button widget artist to the axes.
func (a *Axes) CheckButtons(labels []string, active []bool, opts ...CheckButtonsOptions) *CheckButtons {
	if a == nil || len(labels) == 0 {
		return nil
	}
	cfg := CheckButtonsOptions{
		FaceColor:  render.Color{R: 0.96, G: 0.97, B: 0.98, A: 1},
		EdgeColor:  render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
		TextColor:  render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
		CheckColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeCheckButtonsOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	values := make([]bool, len(labels))
	copy(values, active)
	w := &CheckButtons{
		Labels:     append([]string(nil), labels...),
		Values:     values,
		FaceColor:  cfg.FaceColor,
		EdgeColor:  cfg.EdgeColor,
		TextColor:  cfg.TextColor,
		CheckColor: cfg.CheckColor,
		focus:      -1,
		FontSize:   cfg.FontSize,
		z:          1200,
	}
	a.AddWidget(w)
	return w
}

// RadioButtons adds a radio-button widget artist to the axes.
func (a *Axes) RadioButtons(labels []string, active int, opts ...RadioButtonsOptions) *RadioButtons {
	if a == nil || len(labels) == 0 {
		return nil
	}
	cfg := RadioButtonsOptions{
		FaceColor: render.Color{R: 0.96, G: 0.97, B: 0.98, A: 1},
		EdgeColor: render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
		TextColor: render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
		DotColor:  render.Color{R: 0.85, G: 0.32, B: 0.17, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeRadioButtonsOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	w := &RadioButtons{
		Labels:    append([]string(nil), labels...),
		Active:    clampInt(active, 0, len(labels)-1),
		FaceColor: cfg.FaceColor,
		EdgeColor: cfg.EdgeColor,
		TextColor: cfg.TextColor,
		DotColor:  cfg.DotColor,
		focus:     clampInt(active, 0, len(labels)-1),
		FontSize:  cfg.FontSize,
		z:         1200,
	}
	a.AddWidget(w)
	return w
}

// TextBox adds a text-box widget artist to the axes.
func (a *Axes) TextBox(label, value string, opts ...TextBoxOptions) *TextBox {
	if a == nil {
		return nil
	}
	cfg := TextBoxOptions{
		FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor: render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
		TextColor: render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
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

func (b *Button) OnClicked(cb ButtonCallback) WidgetCallbackID {
	if b == nil || any(cb) == nil {
		return 0
	}
	return b.onClicked.add(cb)
}

func (b *Button) RemoveOnClicked(id WidgetCallbackID) {
	if b == nil {
		return
	}
	b.onClicked.remove(id)
}

func (b *Button) triggerOnClicked() {
	if b == nil {
		return
	}
	b.onClicked.each(func(cb ButtonCallback) { cb(b) })
}

// Click triggers all registered click callbacks.
func (b *Button) Click() {
	if b == nil || !b.Enabled {
		return
	}
	b.triggerOnClicked()
}

func (s *Slider) OnChanged(cb SliderCallback) WidgetCallbackID {
	if s == nil || any(cb) == nil {
		return 0
	}
	return s.onChanged.add(cb)
}

func (s *Slider) RemoveOnChanged(id WidgetCallbackID) {
	if s == nil {
		return
	}
	s.onChanged.remove(id)
}

func (s *Slider) triggerOnChanged() {
	if s == nil {
		return
	}
	s.onChanged.each(func(cb SliderCallback) { cb(s, s.Value) })
}

// SetValue updates the slider value and emits an on-changed event when the
// value changes.
func (s *Slider) SetValue(value float64) {
	if s == nil {
		return
	}
	clamped := normalizeSliderValue(s.Min, s.Max, s.Step, value)
	if clamped == s.Value {
		return
	}
	s.Value = clamped
	s.triggerOnChanged()
}

func (s *RangeSlider) OnChanged(cb RangeSliderCallback) WidgetCallbackID {
	if s == nil || any(cb) == nil {
		return 0
	}
	return s.onChanged.add(cb)
}

func (s *RangeSlider) RemoveOnChanged(id WidgetCallbackID) {
	if s == nil {
		return
	}
	s.onChanged.remove(id)
}

func (s *RangeSlider) triggerOnChanged() {
	if s == nil {
		return
	}
	s.onChanged.each(func(cb RangeSliderCallback) { cb(s, s.Low, s.High) })
}

// SetRange updates the selected interval and emits an on-changed event when
// either endpoint changes.
func (s *RangeSlider) SetRange(low, high float64) {
	if s == nil {
		return
	}
	low = normalizeSliderValue(s.Min, s.Max, s.Step, low)
	high = normalizeSliderValue(s.Min, s.Max, s.Step, high)
	if low > high {
		low, high = high, low
	}
	if low == s.Low && high == s.High {
		return
	}
	s.Low = low
	s.High = high
	s.triggerOnChanged()
}

// SetLow updates the lower range endpoint without crossing the high endpoint.
func (s *RangeSlider) SetLow(low float64) {
	if s == nil {
		return
	}
	low = math.Min(low, s.High)
	s.SetRange(low, s.High)
}

// SetHigh updates the upper range endpoint without crossing the low endpoint.
func (s *RangeSlider) SetHigh(high float64) {
	if s == nil {
		return
	}
	high = math.Max(high, s.Low)
	s.SetRange(s.Low, high)
}

func (c *CheckButtons) OnChanged(cb CheckButtonsCallback) WidgetCallbackID {
	if c == nil || any(cb) == nil {
		return 0
	}
	return c.onChanged.add(cb)
}

func (c *CheckButtons) RemoveOnChanged(id WidgetCallbackID) {
	if c == nil {
		return
	}
	c.onChanged.remove(id)
}

func (c *CheckButtons) triggerOnChanged(index int, value bool) {
	if c == nil {
		return
	}
	c.onChanged.each(func(cb CheckButtonsCallback) { cb(c, index, value) })
}

// SetValue updates one check-button state and emits an on-change event when
// it mutates.
func (c *CheckButtons) SetValue(index int, checked bool) {
	if c == nil {
		return
	}
	if index < 0 || index >= len(c.Values) {
		return
	}
	if c.Values[index] == checked {
		return
	}
	c.Values[index] = checked
	c.triggerOnChanged(index, checked)
}

// Toggle flips one check-button state.
func (c *CheckButtons) Toggle(index int) {
	if c == nil || index < 0 || index >= len(c.Values) {
		return
	}
	c.SetValue(index, !c.Values[index])
}

func (r *RadioButtons) OnChanged(cb RadioButtonsCallback) WidgetCallbackID {
	if r == nil || any(cb) == nil {
		return 0
	}
	return r.onChanged.add(cb)
}

func (r *RadioButtons) RemoveOnChanged(id WidgetCallbackID) {
	if r == nil {
		return
	}
	r.onChanged.remove(id)
}

func (r *RadioButtons) triggerOnChanged(active int) {
	if r == nil {
		return
	}
	r.onChanged.each(func(cb RadioButtonsCallback) { cb(r, active) })
}

// SetActive updates the selected radio index and emits an on-change event when
// the value mutates.
func (r *RadioButtons) SetActive(index int) {
	if r == nil || len(r.Labels) == 0 {
		return
	}
	index = clampInt(index, 0, len(r.Labels)-1)
	if r.Active == index {
		return
	}
	r.Active = index
	r.triggerOnChanged(index)
}

// Next decrements or increments the active radio index by one and emits a
// change event when the index changes.
func (r *RadioButtons) Next(delta int) {
	if r == nil || len(r.Labels) == 0 {
		return
	}
	if delta == 0 {
		return
	}
	r.SetActive(r.Active + delta)
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

func clampWidgetIndex(index, max int) int {
	if max <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= max {
		return max - 1
	}
	return index
}

func (b *Button) Draw(r render.Renderer, ctx *DrawContext) {
	if b == nil || r == nil || ctx == nil {
		return
	}
	bounds := insetRect(ctx.Clip, 6)
	fill := b.FaceColor
	if !b.Enabled {
		fill = mixColor(fill, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.45)
		edge := mixColor(b.EdgeColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.6)
		drawWidgetPanel(r, bounds, fill, edge, 1.25, 10)
		drawCenteredWidgetText(r, ctx, geom.Pt{
			X: bounds.Min.X + bounds.W()/2,
			Y: bounds.Min.Y + bounds.H()/2,
		}, b.Label, b.FontSize, mixColor(b.TextColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.35))
		return
	}
	if b.Pressed {
		fill = mixColor(fill, render.Color{R: 0, G: 0, B: 0, A: 1}, 0.12)
	}
	if b.Hovered && !b.Pressed {
		fill = mixColor(fill, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.06)
	}
	drawWidgetPanel(r, bounds, fill, b.EdgeColor, 1.25, 10)
	labelColor := b.TextColor
	if b.Pressed {
		labelColor = mixColor(b.TextColor, render.Color{R: 0, G: 0, B: 0, A: 1}, 0.3)
	}
	drawCenteredWidgetText(r, ctx, geom.Pt{
		X: bounds.Min.X + bounds.W()/2,
		Y: bounds.Min.Y + bounds.H()/2,
	}, b.Label, b.FontSize, labelColor)
}

func (b *Button) Bounds(ctx *DrawContext) geom.Rect {
	if b == nil || ctx == nil {
		return geom.Rect{}
	}
	return insetRect(ctx.Clip, 6)
}

func (b *Button) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if b == nil || ctx == nil {
		return false, PickInfo{}
	}
	if !b.Enabled {
		return false, PickInfo{}
	}
	if b.Bounds(ctx).Contains(p) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}
func (b *Button) Z() float64   { return b.z }
func (b *Button) WidgetLayer() {}

func (s *Slider) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || r == nil || ctx == nil {
		return
	}
	panel := insetRect(ctx.Clip, 4)
	drawWidgetPanel(r, panel, s.FaceColor, render.Color{A: 0}, 0, 12)
	textColor := s.TextColor
	fontSize := resolvedFontSize(s.FontSize, ctx)
	drawWidgetText(r, ctx, geom.Pt{X: panel.Min.X + 14, Y: panel.Min.Y + 22}, s.Label, fontSize, textColor, TextAlignLeft, textLayoutVAlignTop)
	drawWidgetText(r, ctx, geom.Pt{X: panel.Max.X - 14, Y: panel.Min.Y + 22}, sliderDisplayValue(s), fontSize, textColor, TextAlignRight, textLayoutVAlignTop)

	track := geom.Rect{
		Min: geom.Pt{X: panel.Min.X + 14, Y: panel.Max.Y - 26},
		Max: geom.Pt{X: panel.Max.X - 14, Y: panel.Max.Y - 14},
	}
	drawWidgetPanel(r, track, s.TrackColor, render.Color{A: 0}, 0, track.H()/2)
	fraction := sliderFraction(s.Min, s.Max, s.Value)
	fill := track
	fill.Max.X = fill.Min.X + track.W()*fraction
	drawWidgetPanel(r, fill, s.FillColor, render.Color{A: 0}, 0, fill.H()/2)
	handleX := track.Min.X + track.W()*fraction
	handle := ellipsePath(track.H()*1.9, track.H()*1.9)
	handlePath := applyAffinePath(handle, patchAffine(geom.Pt{X: handleX, Y: track.Min.Y + track.H()/2}, 0))
	r.Path(handlePath, &render.Paint{
		Fill:      s.HandleColor,
		Stroke:    mixColor(s.HandleColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.2),
		LineWidth: 1,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	})
}

func (s *RangeSlider) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || r == nil || ctx == nil {
		return
	}
	panel := insetRect(ctx.Clip, 4)
	drawWidgetPanel(r, panel, s.FaceColor, render.Color{A: 0}, 0, 12)
	textColor := s.TextColor
	fontSize := resolvedFontSize(s.FontSize, ctx)
	drawWidgetText(r, ctx, geom.Pt{X: panel.Min.X + 14, Y: panel.Min.Y + 22}, s.Label, fontSize, textColor, TextAlignLeft, textLayoutVAlignTop)
	drawWidgetText(r, ctx, geom.Pt{X: panel.Max.X - 14, Y: panel.Min.Y + 22}, rangeSliderDisplayValue(s), fontSize, textColor, TextAlignRight, textLayoutVAlignTop)

	track := widgetSliderTrack(panel)
	drawWidgetPanel(r, track, s.TrackColor, render.Color{A: 0}, 0, track.H()/2)
	lowFraction := sliderFraction(s.Min, s.Max, s.Low)
	highFraction := sliderFraction(s.Min, s.Max, s.High)
	fill := track
	fill.Min.X = track.Min.X + track.W()*lowFraction
	fill.Max.X = track.Min.X + track.W()*highFraction
	drawWidgetPanel(r, fill, s.FillColor, render.Color{A: 0}, 0, fill.H()/2)

	for _, fraction := range []float64{lowFraction, highFraction} {
		handleX := track.Min.X + track.W()*fraction
		handle := ellipsePath(track.H()*1.9, track.H()*1.9)
		handlePath := applyAffinePath(handle, patchAffine(geom.Pt{X: handleX, Y: track.Min.Y + track.H()/2}, 0))
		r.Path(handlePath, &render.Paint{
			Fill:      s.HandleColor,
			Stroke:    mixColor(s.HandleColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.2),
			LineWidth: 1,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		})
	}
}

func (s *Slider) Bounds(ctx *DrawContext) geom.Rect {
	if s == nil || ctx == nil {
		return geom.Rect{}
	}
	return insetRect(ctx.Clip, 4)
}

func (s *Slider) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if s == nil || ctx == nil {
		return false, PickInfo{}
	}
	if s.Bounds(ctx).Contains(p) {
		return true, PickInfo{}
	}
	return false, PickInfo{}
}
func (s *Slider) Z() float64   { return s.z }
func (s *Slider) WidgetLayer() {}

func (s *RangeSlider) Bounds(ctx *DrawContext) geom.Rect {
	if s == nil || ctx == nil {
		return geom.Rect{}
	}
	return insetRect(ctx.Clip, 4)
}

func (s *RangeSlider) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if s == nil || ctx == nil {
		return false, PickInfo{}
	}
	bounds := s.Bounds(ctx)
	if !bounds.Contains(p) {
		return false, PickInfo{}
	}
	track := widgetSliderTrack(bounds)
	if track.W() <= 0 {
		return true, PickInfo{}
	}
	lowX := track.Min.X + track.W()*sliderFraction(s.Min, s.Max, s.Low)
	highX := track.Min.X + track.W()*sliderFraction(s.Min, s.Max, s.High)
	if math.Abs(p.X-lowX) <= math.Abs(p.X-highX) {
		return true, PickInfo{Index: 0}
	}
	return true, PickInfo{Index: 1}
}

func (s *RangeSlider) Z() float64   { return s.z }
func (s *RangeSlider) WidgetLayer() {}

func (c *CheckButtons) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if c == nil || ctx == nil {
		return false, PickInfo{}
	}
	if len(c.Labels) == 0 {
		return false, PickInfo{}
	}
	panel := c.Bounds(ctx)
	if !panel.Contains(p) {
		return false, PickInfo{}
	}
	rowHeight := panel.H() / float64(len(c.Labels))
	if rowHeight <= 0 {
		return false, PickInfo{}
	}
	row := int((p.Y - panel.Min.Y) / rowHeight)
	if row < 0 || row >= len(c.Labels) {
		return false, PickInfo{}
	}
	return true, PickInfo{Index: row}
}

func (rdo *RadioButtons) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if rdo == nil || ctx == nil {
		return false, PickInfo{}
	}
	if len(rdo.Labels) == 0 {
		return false, PickInfo{}
	}
	panel := rdo.Bounds(ctx)
	if !panel.Contains(p) {
		return false, PickInfo{}
	}
	rowHeight := panel.H() / float64(len(rdo.Labels))
	if rowHeight <= 0 {
		return false, PickInfo{}
	}
	row := int((p.Y - panel.Min.Y) / rowHeight)
	if row < 0 || row >= len(rdo.Labels) {
		return false, PickInfo{}
	}
	return true, PickInfo{Index: row}
}

func (t *TextBox) Contains(p geom.Pt, ctx *DrawContext) (bool, PickInfo) {
	if t == nil || ctx == nil {
		return false, PickInfo{}
	}
	return t.Bounds(ctx).Contains(p), PickInfo{}
}

func (c *CheckButtons) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || r == nil || ctx == nil {
		return
	}
	panel := insetRect(ctx.Clip, 4)
	drawWidgetPanel(r, panel, c.FaceColor, c.EdgeColor, 1.1, 12)
	if len(c.Labels) == 0 {
		return
	}
	rowHeight := panel.H() / float64(len(c.Labels))
	fontSize := resolvedFontSize(c.FontSize, ctx)
	for i, label := range c.Labels {
		rowMinY := panel.Min.Y + rowHeight*float64(i)
		rowMaxY := rowMinY + rowHeight
		boxSize := math.Min(16, rowHeight*0.42)
		box := geom.Rect{
			Min: geom.Pt{X: panel.Min.X + 14, Y: rowMinY + (rowHeight-boxSize)/2},
			Max: geom.Pt{X: panel.Min.X + 14 + boxSize, Y: rowMaxY - (rowHeight-boxSize)/2},
		}
		drawWidgetPanel(r, box, render.Color{R: 1, G: 1, B: 1, A: 1}, c.EdgeColor, 1, 3)
		if i < len(c.Values) && c.Values[i] {
			path := geom.Path{}
			path.MoveTo(geom.Pt{X: box.Min.X + box.W()*0.18, Y: box.Min.Y + box.H()*0.56})
			path.LineTo(geom.Pt{X: box.Min.X + box.W()*0.42, Y: box.Max.Y - box.H()*0.20})
			path.LineTo(geom.Pt{X: box.Max.X - box.W()*0.16, Y: box.Min.Y + box.H()*0.22})
			r.Path(path, &render.Paint{
				Stroke:    c.CheckColor,
				LineWidth: 2,
				LineJoin:  render.JoinRound,
				LineCap:   render.CapRound,
			})
		}
		drawWidgetText(r, ctx, geom.Pt{X: box.Max.X + 10, Y: rowMinY + rowHeight/2}, label, fontSize, c.TextColor, TextAlignLeft, textLayoutVAlignCenter)
	}
}

func (c *CheckButtons) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil || ctx == nil {
		return geom.Rect{}
	}
	return insetRect(ctx.Clip, 4)
}
func (c *CheckButtons) Z() float64   { return c.z }
func (c *CheckButtons) WidgetLayer() {}

func (rdo *RadioButtons) Draw(r render.Renderer, ctx *DrawContext) {
	if rdo == nil || r == nil || ctx == nil {
		return
	}
	panel := insetRect(ctx.Clip, 4)
	drawWidgetPanel(r, panel, rdo.FaceColor, rdo.EdgeColor, 1.1, 12)
	if len(rdo.Labels) == 0 {
		return
	}
	rowHeight := panel.H() / float64(len(rdo.Labels))
	fontSize := resolvedFontSize(rdo.FontSize, ctx)
	for i, label := range rdo.Labels {
		center := geom.Pt{X: panel.Min.X + 24, Y: panel.Min.Y + rowHeight*float64(i) + rowHeight/2}
		outer := ellipsePath(16, 16)
		outerPath := applyAffinePath(outer, patchAffine(center, 0))
		r.Path(outerPath, &render.Paint{
			Fill:      render.Color{R: 1, G: 1, B: 1, A: 1},
			Stroke:    rdo.EdgeColor,
			LineWidth: 1,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		})
		if i == clampInt(rdo.Active, 0, len(rdo.Labels)-1) {
			inner := ellipsePath(8, 8)
			innerPath := applyAffinePath(inner, patchAffine(center, 0))
			r.Path(innerPath, &render.Paint{Fill: rdo.DotColor})
		}
		drawWidgetText(r, ctx, geom.Pt{X: center.X + 16, Y: center.Y}, label, fontSize, rdo.TextColor, TextAlignLeft, textLayoutVAlignCenter)
	}
}

func (rdo *RadioButtons) Bounds(ctx *DrawContext) geom.Rect {
	if rdo == nil || ctx == nil {
		return geom.Rect{}
	}
	return insetRect(ctx.Clip, 4)
}
func (rdo *RadioButtons) Z() float64   { return rdo.z }
func (rdo *RadioButtons) WidgetLayer() {}

func (t *TextBox) Draw(r render.Renderer, ctx *DrawContext) {
	if t == nil || r == nil || ctx == nil {
		return
	}
	panel := insetRect(ctx.Clip, 4)
	drawWidgetPanel(r, panel, render.Color{A: 0}, render.Color{A: 0}, 0, 0)
	fontSize := resolvedFontSize(t.FontSize, ctx)
	drawWidgetText(r, ctx, geom.Pt{X: panel.Min.X + 4, Y: panel.Min.Y + 20}, t.Label, fontSize, t.TextColor, TextAlignLeft, textLayoutVAlignTop)

	input := geom.Rect{
		Min: geom.Pt{X: panel.Min.X + 4, Y: panel.Min.Y + 30},
		Max: geom.Pt{X: panel.Max.X - 4, Y: panel.Max.Y - 8},
	}
	edge := t.EdgeColor
	if t.Active {
		edge = mixColor(edge, render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1}, 0.65)
	}
	drawWidgetPanel(r, input, t.FaceColor, edge, 1.2, 8)

	display := t.Value
	displayColor := t.TextColor
	if display == "" {
		display = t.Placeholder
		displayColor = mixColor(t.TextColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.45)
	}
	drawWidgetText(r, ctx, geom.Pt{X: input.Min.X + 12, Y: input.Min.Y + input.H()/2}, display, fontSize, displayColor, TextAlignLeft, textLayoutVAlignCenter)
	if t.Active {
		caretIndex := clampInt(t.caret, 0, len([]rune(t.Value)))
		caretX := input.Min.X + 12 + fontSize*0.42*float64(caretIndex)
		if caretX > input.Max.X-12 {
			caretX = input.Max.X - 12
		}
		if caretX < input.Min.X+12 {
			caretX = input.Min.X + 12
		}
		r.Path(pixelLinePath(
			geom.Pt{X: caretX, Y: input.Min.Y + 8},
			geom.Pt{X: caretX, Y: input.Max.Y - 8},
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
	panel := insetRect(ctx.Clip, 4)
	return geom.Rect{
		Min: geom.Pt{X: panel.Min.X, Y: panel.Min.Y + 26},
		Max: geom.Pt{X: panel.Max.X, Y: panel.Max.Y},
	}
}
func (t *TextBox) Z() float64   { return t.z }
func (t *TextBox) WidgetLayer() {}

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

func sliderFraction(min, max, value float64) float64 {
	if max <= min {
		return 0
	}
	return clampFloat((value-min)/(max-min), 0, 1)
}

func widgetSliderTrack(panel geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: panel.Min.X + 14, Y: panel.Max.Y - 26},
		Max: geom.Pt{X: panel.Max.X - 14, Y: panel.Max.Y - 14},
	}
}

func sliderDisplayValue(s *Slider) string {
	if s == nil {
		return ""
	}
	format := strings.TrimSpace(s.ValueFormat)
	if format == "" {
		format = "%.2f"
	}
	formatted := func() string {
		defer func() {
			if recover() != nil {
				format = "%.2f"
			}
		}()
		return fmt.Sprintf(format, s.Value)
	}()
	if strings.HasPrefix(formatted, "%!") {
		return fmt.Sprintf("%.2f", s.Value)
	}
	return formatted
}

func rangeSliderDisplayValue(s *RangeSlider) string {
	if s == nil {
		return ""
	}
	format := strings.TrimSpace(s.ValueFormat)
	if format == "" {
		format = "%.2f"
	}
	formatOne := func(value float64) (out string) {
		defer func() {
			if recover() != nil {
				out = fmt.Sprintf("%.2f", value)
			}
		}()
		out = fmt.Sprintf(format, value)
		if strings.HasPrefix(out, "%!") {
			out = fmt.Sprintf("%.2f", value)
		}
		return out
	}
	return formatOne(s.Low) + " - " + formatOne(s.High)
}

func clampSliderValue(min, max, value float64) float64 {
	if max <= min {
		return min
	}
	return math.Max(min, math.Min(max, value))
}

func normalizeSliderValue(min, max, step, value float64) float64 {
	clamped := clampSliderValue(min, max, value)
	if max <= min {
		return clamped
	}
	if step == 0 {
		return clamped
	}
	step = math.Abs(step)
	if !isFinite(step) || step == 0 {
		return clamped
	}
	based := (clamped - min) / step
	return min + math.Round(based)*step
}

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

func mergeButtonOptions(base, override ButtonOptions) ButtonOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.Pressed != nil {
		base.Pressed = override.Pressed
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	return base
}

func mergeSliderOptions(base, override SliderOptions) SliderOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.TrackColor != (render.Color{}) {
		base.TrackColor = override.TrackColor
	}
	if override.FillColor != (render.Color{}) {
		base.FillColor = override.FillColor
	}
	if override.HandleColor != (render.Color{}) {
		base.HandleColor = override.HandleColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	if override.ValueStep != nil {
		base.ValueStep = override.ValueStep
	}
	if override.ValueFormat != nil {
		base.ValueFormat = override.ValueFormat
	}
	return base
}

func mergeCheckButtonsOptions(base, override CheckButtonsOptions) CheckButtonsOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.CheckColor != (render.Color{}) {
		base.CheckColor = override.CheckColor
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	return base
}

func mergeRadioButtonsOptions(base, override RadioButtonsOptions) RadioButtonsOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.DotColor != (render.Color{}) {
		base.DotColor = override.DotColor
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	return base
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

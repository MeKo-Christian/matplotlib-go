package canvas

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/widgets"
)

func TestWidgetInteractionCheckRadioAndTextBoxAcrossVisualStyles(t *testing.T) {
	tests := []struct {
		name         string
		opt          style.Option
		textClickX   float64
		wantText     string
		checkRow1Y   float64
		radioRow1Y   float64
		wantCheckRow int
		wantRadioRow int
	}{
		{
			name:         "go",
			opt:          style.WithWidgetVisualStyle(style.WidgetVisualGo),
			textClickX:   28.4,
			wantText:     "abZcd",
			checkRow1Y:   40,
			radioRow1Y:   40,
			wantCheckRow: 1,
			wantRadioRow: 1,
		},
		{
			name:         "matplotlib",
			opt:          style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib),
			textClickX:   14.4,
			wantText:     "abZcd",
			checkRow1Y:   40,
			radioRow1Y:   40,
			wantCheckRow: 1,
			wantRadioRow: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fig := core.NewFigure(120, 80, tt.opt)
			axChecks := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
			checks := widgets.NewCheckButtons(axChecks, []string{"A", "B", "C"}, []bool{false, false, false}, widgets.CheckButtonsOptions{})

			var dispatcherChecks Dispatcher
			wiChecks := NewWidgetInteraction(fig, func() error { return nil })
			wiChecks.Attach(&dispatcherChecks)
			defer wiChecks.Detach()

			checkPoint := geom.Pt{X: 20, Y: tt.checkRow1Y}
			if err := dispatcherChecks.Emit(Event{Type: EventMousePress, Figure: fig, Axes: axChecks, Position: checkPoint, Button: MouseButtonLeft}); err != nil {
				t.Fatalf("check press: %v", err)
			}
			if !checks.Values[tt.wantCheckRow] {
				t.Fatalf("check row %d = false, want true; values=%v", tt.wantCheckRow, checks.Values)
			}

			figRadio := core.NewFigure(120, 80, tt.opt)
			axRadio := figRadio.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
			radios := widgets.NewRadioButtons(axRadio, []string{"x", "y", "z"}, 0, widgets.RadioButtonsOptions{})

			var dispatcherRadio Dispatcher
			wiRadio := NewWidgetInteraction(figRadio, func() error { return nil })
			wiRadio.Attach(&dispatcherRadio)
			defer wiRadio.Detach()

			radioPoint := geom.Pt{X: 20, Y: tt.radioRow1Y}
			if err := dispatcherRadio.Emit(Event{Type: EventMousePress, Figure: figRadio, Axes: axRadio, Position: radioPoint, Button: MouseButtonLeft}); err != nil {
				t.Fatalf("radio press: %v", err)
			}
			if radios.Active() != tt.wantRadioRow {
				t.Fatalf("radio active = %d, want %d", radios.Active(), tt.wantRadioRow)
			}

			figText := core.NewFigure(120, 80, tt.opt)
			axText := figText.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
			text := widgets.NewTextBox(axText, "Query", "abcd", widgets.TextBoxOptions{})

			var dispatcherText Dispatcher
			wiText := NewWidgetInteraction(figText, func() error { return nil })
			wiText.Attach(&dispatcherText)
			defer wiText.Detach()

			textPoint := geom.Pt{X: tt.textClickX, Y: 40}
			if err := dispatcherText.Emit(Event{Type: EventMousePress, Figure: figText, Axes: axText, Position: textPoint, Button: MouseButtonLeft}); err != nil {
				t.Fatalf("text press: %v", err)
			}
			if !text.Active {
				t.Fatal("text box should become active on press")
			}
			if err := dispatcherText.Emit(Event{Type: EventKeyPress, Figure: figText, Axes: axText, Key: "Z"}); err != nil {
				t.Fatalf("text insert: %v", err)
			}
			if text.Value() != tt.wantText {
				t.Fatalf("text value = %q, want %q", text.Value(), tt.wantText)
			}
		})
	}
}

func TestWidgetInteractionCheckAndRadioKeyboardNavigation(t *testing.T) {
	fig := core.NewFigure(120, 80)
	axChecks := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 0.45}})
	axRadio := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0.55}, Max: geom.Pt{X: 1, Y: 1}})

	checks := widgets.NewCheckButtons(axChecks, []string{"A", "B", "C"}, []bool{false, false, false}, widgets.CheckButtonsOptions{})
	radios := widgets.NewRadioButtons(axRadio, []string{"x", "y", "z"}, 0, widgets.RadioButtonsOptions{})
	var checkEvents []string
	var radioEvents []int

	checks.OnChanged(func(_ *widgets.CheckButtons, index int, checked bool) {
		if checked {
			checkEvents = append(checkEvents, "on")
		} else {
			checkEvents = append(checkEvents, "off")
		}
	})
	radios.OnChanged(func(_ *widgets.RadioButtons, active int) {
		radioEvents = append(radioEvents, active)
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	checkRow0 := geom.Pt{X: 20, Y: 28}
	radioRow0 := geom.Pt{X: 20, Y: 72}

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: axChecks, Position: checkRow0, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("check press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: axChecks, Position: checkRow0, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("check release: %v", err)
	}
	if !checks.Values[0] {
		t.Fatalf("check row 0 should be true after click")
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: axChecks, Key: "down"}); err != nil {
		t.Fatalf("check down: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: axChecks, Key: "space"}); err != nil {
		t.Fatalf("check space: %v", err)
	}
	if !checks.Values[1] {
		t.Fatalf("check row 1 should be true after keyboard toggle")
	}
	if len(checkEvents) < 2 {
		t.Fatalf("expected at least two check events, got %d", len(checkEvents))
	}

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: axRadio, Position: radioRow0, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("radio press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: axRadio, Position: radioRow0, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("radio release: %v", err)
	}
	if radios.Active() != 0 {
		t.Fatalf("radio should activate row 0 on press, got %d", radios.Active())
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: axRadio, Key: "right"}); err != nil {
		t.Fatalf("radio right key: %v", err)
	}
	if radios.Active() != 1 {
		t.Fatalf("radio should navigate right to row 1, got %d", radios.Active())
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: axRadio, Key: "enter"}); err != nil {
		t.Fatalf("radio enter key: %v", err)
	}
	if len(radioEvents) == 0 {
		t.Fatal("expected radio onChanged events")
	}
}

func TestWidgetInteractionDisabledCheckAndRadioIgnoreInput(t *testing.T) {
	fig := core.NewFigure(120, 80)
	axChecks := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 0.45}})
	axRadio := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0.55}, Max: geom.Pt{X: 1, Y: 1}})

	disabled := true
	checks := widgets.NewCheckButtons(axChecks, []string{"A", "B"}, []bool{false, false}, widgets.CheckButtonsOptions{Disabled: optional.Of(disabled)})
	radios := widgets.NewRadioButtons(axRadio, []string{"x", "y"}, 0, widgets.RadioButtonsOptions{Disabled: optional.Of(disabled)})

	var checkEvents int
	var radioEvents int
	checks.OnChanged(func(*widgets.CheckButtons, int, bool) {
		checkEvents++
	})
	radios.OnChanged(func(*widgets.RadioButtons, int) {
		radioEvents++
	})

	var draws int
	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error {
		draws++
		return nil
	})
	wi.Attach(&dispatcher)
	defer wi.Detach()

	checkRow0 := geom.Pt{X: 20, Y: 28}
	radioRow0 := geom.Pt{X: 20, Y: 72}

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: axChecks, Position: checkRow0, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("disabled check press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: axChecks, Position: checkRow0, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("disabled check release: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: axChecks, Key: "space"}); err != nil {
		t.Fatalf("disabled check space: %v", err)
	}
	if checks.Values[0] || checks.Values[1] {
		t.Fatalf("disabled checks changed to %v, want all false", checks.Values)
	}

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: axRadio, Position: radioRow0, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("disabled radio press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: axRadio, Position: radioRow0, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("disabled radio release: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: axRadio, Key: "right"}); err != nil {
		t.Fatalf("disabled radio right: %v", err)
	}
	if radios.Active() != 0 {
		t.Fatalf("disabled radio active = %d, want 0", radios.Active())
	}
	if checkEvents != 0 || radioEvents != 0 {
		t.Fatalf("disabled widget callbacks = %d/%d, want 0/0", checkEvents, radioEvents)
	}
	if draws != 0 {
		t.Fatalf("disabled check/radio requested %d draws, want 0", draws)
	}
}

func TestWidgetInteractionTextBoxEditing(t *testing.T) {
	setClipboard("")
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	tb := widgets.NewTextBox(ax, "Query", "hello", widgets.TextBoxOptions{})

	var submitCalls []string
	var cancelCalls []string
	tb.OnSubmit(func(_ *widgets.TextBox, value string) {
		submitCalls = append(submitCalls, value)
	})
	tb.OnCancel(func(_ *widgets.TextBox, value string) {
		cancelCalls = append(cancelCalls, value)
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	focusPoint := geom.Pt{X: 70, Y: 50}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: focusPoint, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("text focus press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: focusPoint, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("text focus release: %v", err)
	}
	if !tb.Active {
		t.Fatal("text box should become active on click")
	}

	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Key: "A"}); err != nil {
		t.Fatalf("insert uppercase A: %v", err)
	}
	if tb.Value() != "helloA" {
		t.Fatalf("text value = %q, want helloA", tb.Value())
	}

	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Key: "a", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl+a: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Key: "c", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl+c: %v", err)
	}
	if got := getClipboard(); got != "helloA" {
		t.Fatalf("clipboard = %q, want helloA", got)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Key: "x", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl+x: %v", err)
	}
	if tb.Value() != "" {
		t.Fatalf("text value after cut = %q, want empty", tb.Value())
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Key: "v", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl+v: %v", err)
	}
	if tb.Value() != "helloA" {
		t.Fatalf("text value after paste = %q, want helloA", tb.Value())
	}

	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Key: "enter"}); err != nil {
		t.Fatalf("enter: %v", err)
	}
	if len(submitCalls) != 1 || submitCalls[0] != "helloA" {
		t.Fatalf("submit calls = %v, want [helloA]", submitCalls)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Key: "escape"}); err != nil {
		t.Fatalf("escape: %v", err)
	}
	if len(cancelCalls) != 1 || cancelCalls[0] != "helloA" {
		t.Fatalf("cancel calls = %v, want [helloA]", cancelCalls)
	}
}

func TestWidgetInteractionNonWidgetEventsContinueToUsers(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 0.5, Y: 1}})
	_ = widgets.NewButton(ax, "Run", widgets.ButtonOptions{})

	var received int
	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()
	_ = dispatcher.Connect(EventMousePress, func(ev Event) error {
		received++
		return nil
	})

	if err := dispatcher.Emit(Event{
		Type:     EventMousePress,
		Figure:   fig,
		Position: geom.Pt{X: 90, Y: 40},
		Button:   MouseButtonLeft,
	}); err != nil {
		t.Fatalf("non-widget press: %v", err)
	}
	if received != 1 {
		t.Fatalf("mouse press handlers received %d events, want 1", received)
	}
}

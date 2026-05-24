package canvas

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func TestWidgetInteractionButtonClick(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})

	button := ax.Button("Run")
	clicks := 0
	button.OnClicked(func(*core.Button) {
		clicks++
	})

	var drawCalls int
	wi := NewWidgetInteraction(fig, func() error {
		drawCalls++
		return nil
	})
	var dispatcher Dispatcher
	wi.Attach(&dispatcher)
	defer wi.Detach()

	point := geom.Pt{X: 60, Y: 40}
	if err := dispatcher.Emit(Event{
		Type:     EventMousePress,
		Figure:   fig,
		Axes:     ax,
		Position: point,
		Button:   MouseButtonLeft,
	}); err != nil {
		t.Fatalf("press: %v", err)
	}
	if !button.Pressed || !button.Hovered {
		t.Fatalf("button after press: pressed=%v hovered=%v", button.Pressed, button.Hovered)
	}
	if err := dispatcher.Emit(Event{
		Type:     EventMouseRelease,
		Figure:   fig,
		Axes:     ax,
		Position: point,
		Button:   MouseButtonLeft,
	}); err != nil {
		t.Fatalf("release: %v", err)
	}
	if button.Pressed {
		t.Fatal("button should clear pressed after release")
	}
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}

	button.Enabled = false
	drawCalls = 0
	if err := dispatcher.Emit(Event{
		Type:     EventMousePress,
		Figure:   fig,
		Axes:     ax,
		Position: point,
		Button:   MouseButtonLeft,
	}); err != nil {
		t.Fatalf("disabled press: %v", err)
	}
	if button.Pressed {
		t.Fatal("disabled button should not enter pressed state")
	}
	if err := dispatcher.Emit(Event{
		Type:     EventMouseRelease,
		Figure:   fig,
		Axes:     ax,
		Position: point,
		Button:   MouseButtonLeft,
	}); err != nil {
		t.Fatalf("disabled release: %v", err)
	}
	if clicks != 1 {
		t.Fatalf("disabled click should not fire callback: got %d", clicks)
	}
	if drawCalls == 0 {
		t.Fatal("disabled click should still redraw after state refresh")
	}
}

func TestWidgetInteractionSliderDragAndNudge(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})

	slider := ax.Slider("gain", 0, 10, 5)
	var values []float64
	slider.OnChanged(func(_ *core.Slider, value float64) {
		values = append(values, value)
	})

	var draws int
	wi := NewWidgetInteraction(fig, func() error {
		draws++
		return nil
	})
	var dispatcher Dispatcher
	wi.Attach(&dispatcher)
	defer wi.Detach()

	pressPoint := geom.Pt{X: 90, Y: 60}
	movePoint := geom.Pt{X: 18, Y: 60}

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: pressPoint, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("slider press: %v", err)
	}
	if slider.Value != 8.6 {
		t.Fatalf("slider value after press = %v, want 8.6", slider.Value)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: movePoint, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("slider drag: %v", err)
	}
	if slider.Value != 0 {
		t.Fatalf("slider value after drag = %v, want 0", slider.Value)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: movePoint, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("slider release: %v", err)
	}
	if slider.Dragging {
		t.Fatal("slider should stop dragging after release")
	}

	if len(values) == 0 {
		t.Fatal("expected slider onChanged callback")
	}
	if values[len(values)-1] != 0 {
		t.Fatalf("last slider value = %v, want 0", values[len(values)-1])
	}

	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "right"}); err != nil {
		t.Fatalf("slider right key: %v", err)
	}
	if slider.Value != 0.1 {
		t.Fatalf("slider value after right = %v, want 0.1", slider.Value)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "right", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("slider ctrl+right key: %v", err)
	}
	if slider.Value != 1.1 {
		t.Fatalf("slider value after ctrl+right = %v, want 1.1", slider.Value)
	}
	if draws == 0 {
		t.Fatal("slider interactions should request draws")
	}
}

func TestWidgetInteractionCheckAndRadioKeyboardNavigation(t *testing.T) {
	fig := core.NewFigure(120, 80)
	axChecks := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 0.45}})
	axRadio := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0.55}, Max: geom.Pt{X: 1, Y: 1}})

	checks := axChecks.CheckButtons([]string{"A", "B", "C"}, []bool{false, false, false})
	radios := axRadio.RadioButtons([]string{"x", "y", "z"}, 0)
	var checkEvents []string
	var radioEvents []int

	checks.OnChanged(func(_ *core.CheckButtons, index int, checked bool) {
		if checked {
			checkEvents = append(checkEvents, "on")
		} else {
			checkEvents = append(checkEvents, "off")
		}
	})
	radios.OnChanged(func(_ *core.RadioButtons, active int) {
		radioEvents = append(radioEvents, active)
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: axChecks, Position: geom.Pt{X: 20, Y: 50}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("check press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: axChecks, Position: geom.Pt{X: 20, Y: 50}, Button: MouseButtonLeft}); err != nil {
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

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: axRadio, Position: geom.Pt{X: 20, Y: 10}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("radio press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: axRadio, Position: geom.Pt{X: 20, Y: 10}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("radio release: %v", err)
	}
	if radios.Active != 0 {
		t.Fatalf("radio should activate row 0 on press, got %d", radios.Active)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: axRadio, Key: "right"}); err != nil {
		t.Fatalf("radio right key: %v", err)
	}
	if radios.Active != 1 {
		t.Fatalf("radio should navigate right to row 1, got %d", radios.Active)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: axRadio, Key: "enter"}); err != nil {
		t.Fatalf("radio enter key: %v", err)
	}
	if len(radioEvents) == 0 {
		t.Fatal("expected radio onChanged events")
	}
}

func TestWidgetInteractionTextBoxEditing(t *testing.T) {
	setClipboard("")
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	tb := ax.TextBox("Query", "hello")

	var submitCalls []string
	var cancelCalls []string
	tb.OnSubmit(func(_ *core.TextBox, value string) {
		submitCalls = append(submitCalls, value)
	})
	tb.OnCancel(func(_ *core.TextBox, value string) {
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
	if tb.Value != "helloA" {
		t.Fatalf("text value = %q, want helloA", tb.Value)
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
	if tb.Value != "" {
		t.Fatalf("text value after cut = %q, want empty", tb.Value)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Key: "v", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl+v: %v", err)
	}
	if tb.Value != "helloA" {
		t.Fatalf("text value after paste = %q, want helloA", tb.Value)
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
	_ = ax.Button("Run")

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

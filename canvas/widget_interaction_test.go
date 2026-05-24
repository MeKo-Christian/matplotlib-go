package canvas

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func assertCloseEnough(t *testing.T, got, want float64) {
	t.Helper()
	const eps = 1e-9
	if math.Abs(got-want) > eps {
		t.Fatalf("value = %v, want %v", got, want)
	}
}

func sortedPair(a, b float64) (float64, float64) {
	if a <= b {
		return a, b
	}
	return b, a
}

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

func TestWidgetInteractionSpanSelectorMouseAndKeyboard(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	span := ax.SpanSelector("horizontal")

	var got [][2]float64
	span.OnSelect(func(_ *core.SpanSelector, min, max float64) {
		got = append(got, [2]float64{min, max})
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	startPx := geom.Pt{X: 40, Y: 50}
	endPx := geom.Pt{X: 100, Y: 50}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: startPx, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("span press: %v", err)
	}
	if !span.Active {
		t.Fatal("span should become active on drag start")
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: endPx, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("span drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: endPx, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("span release: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("span select callbacks = %d, want 1", len(got))
	}
	dStart, _ := ax.PixelToData(startPx)
	dEnd, _ := ax.PixelToData(endPx)
	wantMin, wantMax := sortedPair(dStart.X, dEnd.X)
	assertCloseEnough(t, got[0][0], wantMin)
	assertCloseEnough(t, got[0][1], wantMax)

	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "left"}); err != nil {
		t.Fatalf("span left key: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "right", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("span right ctrl key: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("span callback count = %d, want 3", len(got))
	}
	expectedAfter := wantMin - 5 + 50
	assertCloseEnough(t, span.Start, expectedAfter)
	assertCloseEnough(t, span.End, wantMax-5+50)
}

func TestWidgetInteractionRectangleSelectorMouseAndKeyboard(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	rect := ax.RectangleSelector()

	var rectBounds [][2]float64
	rect.OnSelect(func(_ *core.RectangleSelector, bounds geom.Rect) {
		rectBounds = append(rectBounds, [2]float64{
			bounds.W(),
			bounds.H(),
		})
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	press := geom.Pt{X: 30, Y: 30}
	move := geom.Pt{X: 120, Y: 90}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: press, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("rectangle press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: move, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("rectangle drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: move, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("rectangle release: %v", err)
	}
	if !rect.Active {
		t.Fatal("rectangle should be active after drag")
	}
	if len(rectBounds) != 1 {
		t.Fatalf("rectangle select callbacks = %d, want 1", len(rectBounds))
	}
	if rectBounds[0][0] != rectBounds[0][1] {
		t.Fatalf("rectangle shift drag expected square bounds, got w=%g h=%g", rectBounds[0][0], rectBounds[0][1])
	}

	beforeMin, beforeMax := rect.Min, rect.Max
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "left"}); err != nil {
		t.Fatalf("rectangle left key: %v", err)
	}
	assertCloseEnough(t, rect.Min.X-beforeMin.X, -5)
	assertCloseEnough(t, rect.Max.X-beforeMax.X, -5)
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "right", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("rectangle ctrl-right key: %v", err)
	}
	assertCloseEnough(t, rect.Min.X-(beforeMin.X+45), 0)
	assertCloseEnough(t, rect.Max.X-(beforeMax.X+45), 0)
}

func TestWidgetInteractionRectangleSelectorModifierMouseCreate(t *testing.T) {
	fig := core.NewFigure(200, 120)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 200)
	ax.SetYLim(0, 120)
	rect := ax.RectangleSelector()

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	startCenter := geom.Pt{X: 120, Y: 30}
	endCenter := geom.Pt{X: 160, Y: 70}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: startCenter, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: endCenter, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl move: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: endCenter, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("ctrl release: %v", err)
	}

	startData, ok := ax.PixelToData(startCenter)
	if !ok {
		t.Fatal("ctrl start pixelToData failed")
	}
	endData, ok := ax.PixelToData(endCenter)
	if !ok {
		t.Fatal("ctrl end pixelToData failed")
	}

	assertCloseEnough(t, rect.Min.X, 2*startData.X-endData.X)
	assertCloseEnough(t, rect.Max.X, endData.X)
	assertCloseEnough(t, rect.Min.Y, 2*startData.Y-endData.Y)
	assertCloseEnough(t, rect.Max.Y, endData.Y)

	startSquare := geom.Pt{X: 40, Y: 80}
	endSquare := geom.Pt{X: 100, Y: 60}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: startSquare, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("shift press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: endSquare, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("shift move: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: endSquare, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("shift release: %v", err)
	}

	if math.Abs((rect.Max.X-rect.Min.X)-(rect.Max.Y-rect.Min.Y)) > 1e-9 {
		t.Fatalf("rectangle with shift should be square, got width=%g height=%g", rect.Max.X-rect.Min.X, rect.Max.Y-rect.Min.Y)
	}

	startSquareFromCenter := geom.Pt{X: 150, Y: 40}
	endSquareFromCenter := geom.Pt{X: 170, Y: 100}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: startSquareFromCenter, Button: MouseButtonLeft, Modifiers: ModifierShift | ModifierControl}); err != nil {
		t.Fatalf("shift+ctrl press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: endSquareFromCenter, Button: MouseButtonLeft, Modifiers: ModifierShift | ModifierControl}); err != nil {
		t.Fatalf("shift+ctrl move: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: endSquareFromCenter, Button: MouseButtonLeft, Modifiers: ModifierShift | ModifierControl}); err != nil {
		t.Fatalf("shift+ctrl release: %v", err)
	}

	startDC, ok := ax.PixelToData(startSquareFromCenter)
	if !ok {
		t.Fatal("shift+ctrl start pixelToData failed")
	}
	if math.Abs((rect.Min.X+rect.Max.X)/2-startDC.X) > 1e-9 {
		t.Fatalf("rectangle with shift+ctrl should stay centered on press, got center x %g want %g", (rect.Min.X+rect.Max.X)/2, startDC.X)
	}
	if math.Abs((rect.Max.X-rect.Min.X)-(rect.Max.Y-rect.Min.Y)) > 1e-9 {
		t.Fatalf("rectangle with shift+ctrl should remain square, got width=%g height=%g", rect.Max.X-rect.Min.X, rect.Max.Y-rect.Min.Y)
	}
}

func TestWidgetInteractionPolygonSelectorEdit(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	polygon := ax.PolygonSelector()

	var onSelectCount int
	polygon.OnSelect(func(*core.PolygonSelector, []geom.Pt) {
		onSelectCount++
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	p1 := geom.Pt{X: 30, Y: 20}
	p2 := geom.Pt{X: 130, Y: 20}
	p3 := geom.Pt{X: 80, Y: 80}
	p1Data, ok := ax.PixelToData(p1)
	if !ok {
		t.Fatal("pixelToData failed for first polygon point")
	}
	p2Data, ok := ax.PixelToData(p2)
	if !ok {
		t.Fatal("pixelToData failed for second polygon point")
	}
	p3Data, ok := ax.PixelToData(p3)
	if !ok {
		t.Fatal("pixelToData failed for third polygon point")
	}
	if !polygon.AppendPoint(p1Data) || !polygon.AppendPoint(p2Data) || !polygon.AppendPoint(p3Data) {
		t.Fatal("manual polygon point append should succeed")
	}
	if !polygon.Close() {
		t.Fatal("polygon close should succeed")
	}
	if !polygon.Closed {
		t.Fatal("polygon should be closed")
	}
	points := []geom.Pt{p1Data, p2Data, p3Data}
	if onSelectCount == 0 {
		polygon.TriggerOnSelect()
		if onSelectCount != 1 {
			t.Fatal("expected polygon onSelect after close")
		}
	}

	// Shift+drag should move all vertices together.
	center := geom.Pt{X: 80, Y: 40}
	beforeData0 := polygon.Points[0]
	beforeMoveData, _ := ax.PixelToData(center)
	afterMoveData, _ := ax.PixelToData(geom.Pt{X: 90, Y: 40})
	delta := geom.Pt{
		X: afterMoveData.X - beforeMoveData.X,
		Y: afterMoveData.Y - beforeMoveData.Y,
	}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: center, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("polygon move press: %v", err)
	}
	translated := geom.Pt{X: 90, Y: 40}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: translated, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("polygon move drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: translated, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("polygon move release: %v", err)
	}
	if len(polygon.Points) != len(points) {
		t.Fatalf("polygon points = %d, want %d", len(polygon.Points), len(points))
	}
	assertCloseEnough(t, polygon.Points[0].X-beforeData0.X-delta.X, 0)
	assertCloseEnough(t, polygon.Points[0].Y-beforeData0.Y-delta.Y, 0)
}

func TestWidgetInteractionPolygonSelectorPreCompleteMoveModes(t *testing.T) {
	fig := core.NewFigure(200, 120)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 200)
	ax.SetYLim(0, 120)

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	addPoint := func(pt geom.Pt) {
		if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: pt, Button: MouseButtonLeft}); err != nil {
			t.Fatalf("add press %v: %v", pt, err)
		}
		if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: pt, Button: MouseButtonLeft}); err != nil {
			t.Fatalf("add release %v: %v", pt, err)
		}
	}

	poly := ax.PolygonSelector()

	addPoint(geom.Pt{X: 50, Y: 70})
	addPoint(geom.Pt{X: 150, Y: 70})

	// Move all vertices before completion (shift).
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: geom.Pt{X: 100, Y: 100}, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("polygon shift press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: geom.Pt{X: 100, Y: 120}, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("polygon shift move: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: geom.Pt{X: 100, Y: 120}, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
		t.Fatalf("polygon shift release: %v", err)
	}

	addPoint(geom.Pt{X: 50, Y: 100})
	addPoint(geom.Pt{X: 50, Y: 70})

	if got, ok := poly.Points[0], true; !ok {
		_ = got
	}
	assertCloseEnough(t, poly.Points[0].Y, 90)
	assertCloseEnough(t, poly.Points[1].Y, 90)

	// Move a vertex before completion (control).
	poly2 := ax.PolygonSelector()
	addPoint = func(pt geom.Pt) {
		if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: pt, Button: MouseButtonLeft}); err != nil {
			t.Fatalf("add2 press %v: %v", pt, err)
		}
		if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: pt, Button: MouseButtonLeft}); err != nil {
			t.Fatalf("add2 release %v: %v", pt, err)
		}
	}
	addPoint(geom.Pt{X: 20, Y: 20})
	addPoint(geom.Pt{X: 120, Y: 20})

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: geom.Pt{X: 20, Y: 20}, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("polygon control press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: geom.Pt{X: 30, Y: 20}, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("polygon control move: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: geom.Pt{X: 30, Y: 20}, Button: MouseButtonLeft, Modifiers: ModifierControl}); err != nil {
		t.Fatalf("polygon control release: %v", err)
	}

	addPoint(geom.Pt{X: 20, Y: 120})
	addPoint(geom.Pt{X: 30, Y: 20})

	if poly2.Points[0].X != 30 {
		t.Fatalf("polygon control move should shift first vertex x, got %g", poly2.Points[0].X)
	}
}

func TestWidgetInteractionEllipseSelectorMouse(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	ellipse := ax.EllipseSelector()

	var got float64
	var selected bool
	ellipse.OnSelect(func(_ *core.EllipseSelector, bounds geom.Rect) {
		got = bounds.W()
		selected = true
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	press := geom.Pt{X: 80, Y: 40}
	move := geom.Pt{X: 120, Y: 80}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: press, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("ellipse press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: move, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("ellipse drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: move, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("ellipse release: %v", err)
	}
	if !ellipse.Active || !selected {
		t.Fatal("ellipse should become active and selected after drag")
	}
	if got <= 0 {
		t.Fatalf("expected positive ellipse width on select, got %g", got)
	}
}

func TestWidgetInteractionLassoSelectorMouse(t *testing.T) {
	fig := core.NewFigure(160, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 100)
	lasso := ax.LassoSelector()

	var got int
	lasso.OnSelect(func(_ *core.LassoSelector, points []geom.Pt) {
		got = len(points)
	})

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	points := []geom.Pt{{X: 20, Y: 20}, {X: 60, Y: 20}, {X: 80, Y: 70}}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: points[0], Button: MouseButtonLeft}); err != nil {
		t.Fatalf("lasso press: %v", err)
	}
	for _, point := range points[1:] {
		if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: point, Button: MouseButtonLeft}); err != nil {
			t.Fatalf("lasso move: %v", err)
		}
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: points[len(points)-1], Button: MouseButtonLeft}); err != nil {
		t.Fatalf("lasso release: %v", err)
	}
	if !lasso.Active {
		t.Fatal("lasso should activate on release")
	}
	if got < len(points) {
		t.Fatalf("lasso onSelect points = %d, want >=%d", got, len(points))
	}
}

func TestWidgetInteractionCursorHover(t *testing.T) {
	fig := core.NewFigure(200, 100)
	left := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 0.5, Y: 1}})
	right := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.5, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	leftCursor := left.Cursor()
	multiCursor := left.MultiCursor(right)

	var dispatcher Dispatcher
	wi := NewWidgetInteraction(fig, func() error { return nil })
	wi.Attach(&dispatcher)
	defer wi.Detach()

	insideLeft := geom.Pt{X: 40, Y: 50}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Position: insideLeft}); err != nil {
		t.Fatalf("move left: %v", err)
	}
	if _, _, ok := leftCursor.Position(); !ok {
		t.Fatal("left cursor should be visible inside left axis")
	}
	if _, _, ok := multiCursor.Position(); !ok {
		t.Fatal("multi cursor should be visible inside left axis")
	}

	insideRight := geom.Pt{X: 140, Y: 50}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Position: insideRight}); err != nil {
		t.Fatalf("move right: %v", err)
	}
	if _, _, ok := leftCursor.Position(); ok {
		t.Fatal("left cursor should hide outside left axis")
	}
	if _, _, ok := multiCursor.Position(); !ok {
		t.Fatal("multi cursor should remain visible in right axis")
	}

	outside := geom.Pt{X: 250, Y: 50}
	if err := dispatcher.Emit(Event{Type: EventFigureLeave, Figure: fig, Position: outside}); err != nil {
		t.Fatalf("figure leave: %v", err)
	}
	if _, _, ok := multiCursor.Position(); ok {
		t.Fatal("multi cursor should hide on figure leave")
	}
}

package canvas

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/widgets"
)

func TestWidgetInteractionAcrossVisualStyles(t *testing.T) {
	styles := []struct {
		name string
		opt  style.Option
	}{
		{"go", style.WithWidgetVisualStyle(style.WidgetVisualGo)},
		{"matplotlib", style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib)},
	}

	for _, st := range styles {
		t.Run(st.name, func(t *testing.T) {
			// Active-state semantics + disabled states: a button click fires
			// exactly one callback while enabled and none while disabled. Each
			// widget uses its own full-axes figure so hit-testing matches the
			// proven coordinates from the single-style interaction tests.
			figB := core.NewFigure(120, 80, st.opt)
			axB := figB.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
			button := widgets.NewButton(axB, "Run", widgets.ButtonOptions{})
			clicks := 0
			button.OnClicked(func(*widgets.Button) { clicks++ })

			var dispatcherB Dispatcher
			wiB := NewWidgetInteraction(figB, func() error { return nil })
			wiB.Attach(&dispatcherB)
			defer wiB.Detach()

			pressRelease := func(d *Dispatcher, fig *core.Figure, ax *core.Axes, p geom.Pt, mods Modifier) {
				t.Helper()
				if err := d.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: p, Button: MouseButtonLeft, Modifiers: mods}); err != nil {
					t.Fatalf("press: %v", err)
				}
				if err := d.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: p, Button: MouseButtonLeft, Modifiers: mods}); err != nil {
					t.Fatalf("release: %v", err)
				}
			}

			buttonCenter := geom.Pt{X: 60, Y: 40}
			pressRelease(&dispatcherB, figB, axB, buttonCenter, 0)
			if clicks != 1 {
				t.Fatalf("enabled button clicks = %d, want 1", clicks)
			}
			if button.Pressed {
				t.Fatal("button should clear pressed after release")
			}
			button.Enabled = false
			pressRelease(&dispatcherB, figB, axB, buttonCenter, 0)
			if clicks != 1 {
				t.Fatalf("disabled button clicks = %d, want 1 (unchanged)", clicks)
			}

			// Handle behavior + keyboard modifiers: a drag focuses the slider, then
			// keyboard nudges apply deterministic deltas (step and ctrl-boosted
			// step) independent of the style's handle geometry.
			figS := core.NewFigure(120, 80, st.opt)
			axS := figS.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
			slider := widgets.NewSlider(axS, "gain", 0, 10, 5, widgets.SliderOptions{})
			var sliderChanges int
			slider.OnChanged(func(*widgets.Slider, float64) { sliderChanges++ })

			var dispatcherS Dispatcher
			wiS := NewWidgetInteraction(figS, func() error { return nil })
			wiS.Attach(&dispatcherS)
			defer wiS.Detach()

			pressRelease(&dispatcherS, figS, axS, geom.Pt{X: 90, Y: 60}, 0)
			if sliderChanges == 0 {
				t.Fatal("slider press should fire OnChanged")
			}
			base := slider.Value()
			if err := dispatcherS.Emit(Event{Type: EventKeyPress, Figure: figS, Axes: axS, Key: "right"}); err != nil {
				t.Fatalf("slider right: %v", err)
			}
			assertCloseEnough(t, slider.Value()-base, 0.1)
			afterStep := slider.Value()
			if err := dispatcherS.Emit(Event{Type: EventKeyPress, Figure: figS, Axes: axS, Key: "right", Modifiers: ModifierControl}); err != nil {
				t.Fatalf("slider ctrl+right: %v", err)
			}
			assertCloseEnough(t, slider.Value()-afterStep, 1.0)

			// Disabled state ignores keyboard nudges.
			slider.Enabled = false
			frozen := slider.Value()
			if err := dispatcherS.Emit(Event{Type: EventKeyPress, Figure: figS, Axes: axS, Key: "right"}); err != nil {
				t.Fatalf("disabled slider right: %v", err)
			}
			assertCloseEnough(t, slider.Value()-frozen, 0)

			// Handle/modifier geometry: a shift-constrained rectangle drag must
			// produce a square selection in both styles (data-space, geometry
			// independent of widget chrome).
			figR := core.NewFigure(160, 100, st.opt)
			axR := figR.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
			axR.SetXLim(0, 100)
			axR.SetYLim(0, 100)
			rect := widgets.NewRectangleSelector(axR, widgets.RectangleSelectorOptions{})

			var dispatcherR Dispatcher
			wiR := NewWidgetInteraction(figR, func() error { return nil })
			wiR.Attach(&dispatcherR)
			defer wiR.Detach()

			press := geom.Pt{X: 30, Y: 30}
			move := geom.Pt{X: 120, Y: 90}
			if err := dispatcherR.Emit(Event{Type: EventMousePress, Figure: figR, Axes: axR, Position: press, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
				t.Fatalf("rect press: %v", err)
			}
			if err := dispatcherR.Emit(Event{Type: EventMouseMove, Figure: figR, Axes: axR, Position: move, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
				t.Fatalf("rect move: %v", err)
			}
			if err := dispatcherR.Emit(Event{Type: EventMouseRelease, Figure: figR, Axes: axR, Position: move, Button: MouseButtonLeft, Modifiers: ModifierShift}); err != nil {
				t.Fatalf("rect release: %v", err)
			}
			if !rect.Active {
				t.Fatal("rectangle selector should be active after drag")
			}
			if math.Abs((rect.Max.X-rect.Min.X)-(rect.Max.Y-rect.Min.Y)) > 1e-9 {
				t.Fatalf("shift drag should be square, got w=%g h=%g", rect.Max.X-rect.Min.X, rect.Max.Y-rect.Min.Y)
			}
		})
	}
}

func TestWidgetInteractionSliderDragUsesVisualStyleGeometry(t *testing.T) {
	tests := []struct {
		name      string
		opt       style.Option
		press     geom.Pt
		wantValue float64
	}{
		{
			name:      "go",
			opt:       style.WithWidgetVisualStyle(style.WidgetVisualGo),
			press:     geom.Pt{X: 90, Y: 40},
			wantValue: 8.6,
		},
		{
			name:      "matplotlib",
			opt:       style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib),
			press:     geom.Pt{X: 90, Y: 40},
			wantValue: 7.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fig := core.NewFigure(120, 80, tt.opt)
			ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
			slider := widgets.NewSlider(ax, "gain", 0, 10, 5, widgets.SliderOptions{})

			var dispatcher Dispatcher
			wi := NewWidgetInteraction(fig, func() error { return nil })
			wi.Attach(&dispatcher)
			defer wi.Detach()

			if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: tt.press, Button: MouseButtonLeft}); err != nil {
				t.Fatalf("slider press: %v", err)
			}
			assertCloseEnough(t, slider.Value(), tt.wantValue)
		})
	}
}

func TestWidgetInteractionRangeSliderHandleSelectionUsesVisualStyleGeometry(t *testing.T) {
	tests := []struct {
		name     string
		opt      style.Option
		press    geom.Pt
		wantLow  float64
		wantHigh float64
	}{
		{
			name:     "go",
			opt:      style.WithWidgetVisualStyle(style.WidgetVisualGo),
			press:    geom.Pt{X: 30, Y: 40},
			wantLow:  1.4,
			wantHigh: 8,
		},
		{
			name:     "matplotlib",
			opt:      style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib),
			press:    geom.Pt{X: 30, Y: 40},
			wantLow:  2.5,
			wantHigh: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fig := core.NewFigure(120, 80, tt.opt)
			ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
			slider := widgets.NewRangeSlider(ax, "window", 0, 10, 2, 8, widgets.RangeSliderOptions{})

			var dispatcher Dispatcher
			wi := NewWidgetInteraction(fig, func() error { return nil })
			wi.Attach(&dispatcher)
			defer wi.Detach()

			if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: tt.press, Button: MouseButtonLeft}); err != nil {
				t.Fatalf("range slider press: %v", err)
			}
			assertCloseEnough(t, slider.Low(), tt.wantLow)
			assertCloseEnough(t, slider.High(), tt.wantHigh)
		})
	}
}

func TestWidgetInteractionButtonClick(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})

	button := widgets.NewButton(ax, "Run", widgets.ButtonOptions{})
	clicks := 0
	button.OnClicked(func(*widgets.Button) {
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

func TestWidgetInteractionCanUseDrawIdleCallback(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	button := widgets.NewButton(ax, "Run", widgets.ButtonOptions{})

	var drawIdleCalls int
	wi := NewWidgetInteraction(fig, func() error {
		drawIdleCalls++
		return nil
	})
	var dispatcher Dispatcher
	wi.Attach(&dispatcher)
	defer wi.Detach()

	point := geom.Pt{X: 60, Y: 40}
	if err := dispatcher.Emit(Event{
		Type:     EventMouseMove,
		Figure:   fig,
		Position: point,
	}); err != nil {
		t.Fatalf("hover: %v", err)
	}
	if !button.Hovered {
		t.Fatal("button should enter hovered state")
	}
	if drawIdleCalls != 1 {
		t.Fatalf("draw-idle callback calls = %d, want 1", drawIdleCalls)
	}
}

func TestWidgetInteractionButtonKeyboardActivation(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})

	button := widgets.NewButton(ax, "Run", widgets.ButtonOptions{})
	clicks := 0
	button.OnClicked(func(*widgets.Button) {
		clicks++
	})

	wi := NewWidgetInteraction(fig, func() error { return nil })
	var dispatcher Dispatcher
	wi.Attach(&dispatcher)
	defer wi.Detach()

	point := geom.Pt{X: 60, Y: 40}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: point, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("button focus press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: point, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("button focus release: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "enter"}); err != nil {
		t.Fatalf("button enter: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "space"}); err != nil {
		t.Fatalf("button space: %v", err)
	}
	if clicks != 3 {
		t.Fatalf("clicks = %d, want mouse click plus enter and space", clicks)
	}
}

func TestWidgetInteractionSliderDragAndNudge(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})

	slider := widgets.NewSlider(ax, "gain", 0, 10, 5, widgets.SliderOptions{})
	var values []float64
	slider.OnChanged(func(_ *widgets.Slider, value float64) {
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
	if slider.Value() != 8.6 {
		t.Fatalf("slider value after press = %v, want 8.6", slider.Value())
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: movePoint, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("slider drag: %v", err)
	}
	if slider.Value() != 0 {
		t.Fatalf("slider value after drag = %v, want 0", slider.Value())
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
	if slider.Value() != 0.1 {
		t.Fatalf("slider value after right = %v, want 0.1", slider.Value())
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "right", Modifiers: ModifierControl}); err != nil {
		t.Fatalf("slider ctrl+right key: %v", err)
	}
	if slider.Value() != 1.1 {
		t.Fatalf("slider value after ctrl+right = %v, want 1.1", slider.Value())
	}
	if draws == 0 {
		t.Fatal("slider interactions should request draws")
	}
}

func TestWidgetInteractionRangeSliderDragAndNudge(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})

	slider := widgets.NewRangeSlider(ax, "window", 0, 10, 2, 8, widgets.RangeSliderOptions{})
	var ranges [][2]float64
	slider.OnChanged(func(_ *widgets.RangeSlider, low, high float64) {
		ranges = append(ranges, [2]float64{low, high})
	})

	var draws int
	wi := NewWidgetInteraction(fig, func() error {
		draws++
		return nil
	})
	var dispatcher Dispatcher
	wi.Attach(&dispatcher)
	defer wi.Detach()

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: geom.Pt{X: 85, Y: 60}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("range slider high press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: geom.Pt{X: 110, Y: 60}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("range slider high drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: geom.Pt{X: 110, Y: 60}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("range slider high release: %v", err)
	}
	if slider.High() != 10 {
		t.Fatalf("range slider high = %v, want 10", slider.High())
	}

	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: ax, Key: "left"}); err != nil {
		t.Fatalf("range slider left key: %v", err)
	}
	if slider.High() != 9.9 {
		t.Fatalf("range slider high after left = %v, want 9.9", slider.High())
	}
	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: ax, Position: geom.Pt{X: 25, Y: 60}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("range slider low press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseMove, Figure: fig, Axes: ax, Position: geom.Pt{X: 14, Y: 60}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("range slider low drag: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: ax, Position: geom.Pt{X: 14, Y: 60}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("range slider low release: %v", err)
	}
	if slider.Low() != 0 {
		t.Fatalf("range slider low = %v, want 0", slider.Low())
	}
	if len(ranges) == 0 {
		t.Fatal("expected range slider callbacks")
	}
	if draws == 0 {
		t.Fatal("range slider interactions should request draws")
	}
}

func TestWidgetInteractionDisabledSlidersIgnoreKeyboardNudges(t *testing.T) {
	fig := core.NewFigure(120, 80)
	axSlider := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 0.45}})
	axRange := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0.55}, Max: geom.Pt{X: 1, Y: 1}})

	slider := widgets.NewSlider(axSlider, "gain", 0, 10, 5, widgets.SliderOptions{})
	slider.Enabled = false
	rangeSlider := widgets.NewRangeSlider(axRange, "window", 0, 10, 2, 8, widgets.RangeSliderOptions{})
	rangeSlider.Enabled = false

	var sliderEvents int
	var rangeEvents int
	slider.OnChanged(func(_ *widgets.Slider, _ float64) {
		sliderEvents++
	})
	rangeSlider.OnChanged(func(_ *widgets.RangeSlider, _, _ float64) {
		rangeEvents++
	})

	var draws int
	wi := NewWidgetInteraction(fig, func() error {
		draws++
		return nil
	})
	var dispatcher Dispatcher
	wi.Attach(&dispatcher)
	defer wi.Detach()

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: axSlider, Position: geom.Pt{X: 90, Y: 30}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("disabled slider press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: axSlider, Position: geom.Pt{X: 90, Y: 30}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("disabled slider release: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: axSlider, Key: "right"}); err != nil {
		t.Fatalf("disabled slider right key: %v", err)
	}
	if slider.Value() != 5 {
		t.Fatalf("disabled slider value = %v, want unchanged 5", slider.Value())
	}

	if err := dispatcher.Emit(Event{Type: EventMousePress, Figure: fig, Axes: axRange, Position: geom.Pt{X: 85, Y: 60}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("disabled range slider press: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventMouseRelease, Figure: fig, Axes: axRange, Position: geom.Pt{X: 85, Y: 60}, Button: MouseButtonLeft}); err != nil {
		t.Fatalf("disabled range slider release: %v", err)
	}
	if err := dispatcher.Emit(Event{Type: EventKeyPress, Figure: fig, Axes: axRange, Key: "left"}); err != nil {
		t.Fatalf("disabled range slider left key: %v", err)
	}
	if rangeSlider.Low() != 2 || rangeSlider.High() != 8 {
		t.Fatalf("disabled range slider = [%v, %v], want unchanged [2, 8]", rangeSlider.Low(), rangeSlider.High())
	}
	if sliderEvents != 0 || rangeEvents != 0 {
		t.Fatalf("disabled slider callbacks = %d/%d, want 0/0", sliderEvents, rangeEvents)
	}
	if draws != 0 {
		t.Fatalf("disabled slider interactions requested %d draws, want 0", draws)
	}
}

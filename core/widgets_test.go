package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestWidgetConstructorsPrepareAxesAndStoreState(t *testing.T) {
	fig := NewFigure(800, 600)
	axButton := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.3, Y: 0.2}})
	axSlider := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.25}, Max: geom.Pt{X: 0.5, Y: 0.38}})
	axRange := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.58}, Max: geom.Pt{X: 0.5, Y: 0.70}})
	axCheck := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.55, Y: 0.1}, Max: geom.Pt{X: 0.8, Y: 0.32}})
	axRadio := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.82, Y: 0.1}, Max: geom.Pt{X: 0.95, Y: 0.32}})
	axText := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.42}, Max: geom.Pt{X: 0.55, Y: 0.56}})

	pressed := true
	active := true
	button := axButton.Button("Run", ButtonOptions{Pressed: &pressed})
	slider := axSlider.Slider("gain", 0, 10, 12)
	rangeSlider := axRange.RangeSlider("window", 0, 10, 8, 2)
	checks := axCheck.CheckButtons([]string{"A", "B"}, []bool{true, false})
	radios := axRadio.RadioButtons([]string{"x", "y", "z"}, 5)
	text := axText.TextBox("Query", "", TextBoxOptions{Placeholder: "type...", Active: &active})

	if button == nil || slider == nil || rangeSlider == nil || checks == nil || radios == nil || text == nil {
		t.Fatal("expected widget constructors to return artists")
	}
	if !button.Pressed {
		t.Fatal("button should store pressed state")
	}
	disabled := true
	disabledButton := axButton.Button("Stop", ButtonOptions{Disabled: &disabled})
	if disabledButton.Enabled {
		t.Fatal("button should honor Disabled option")
	}
	if slider.Value != 10 {
		t.Fatalf("slider value = %v, want clamped max 10", slider.Value)
	}
	if rangeSlider.Low != 2 || rangeSlider.High != 8 {
		t.Fatalf("range slider range = [%v, %v], want sorted [2, 8]", rangeSlider.Low, rangeSlider.High)
	}
	if radios.Active != 2 {
		t.Fatalf("radio active index = %d, want 2", radios.Active)
	}
	if text.Placeholder != "type..." || !text.Active {
		t.Fatal("text box should store placeholder and active state")
	}
	for _, ax := range []*Axes{axButton, axSlider, axRange, axCheck, axRadio, axText} {
		if ax.XAxis.ShowTicks || ax.YAxis.ShowTicks || ax.XAxis.ShowLabels || ax.YAxis.ShowLabels {
			t.Fatal("widget constructors should hide axis ticks and labels")
		}
	}
}

func TestSliderConstructorsSnapInitialValuesToStep(t *testing.T) {
	fig := NewFigure(800, 600)
	axSlider := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.8, Y: 0.2}})
	axRange := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.3}, Max: geom.Pt{X: 0.8, Y: 0.4}})

	step := 0.5
	slider := axSlider.Slider("gain", 0, 10, 5.26, SliderOptions{ValueStep: &step})
	if slider.Value != 5.5 {
		t.Fatalf("slider initial value = %v, want snapped 5.5", slider.Value)
	}

	rangeSlider := axRange.RangeSlider("window", 0, 10, 7.76, 2.24, RangeSliderOptions{ValueStep: &step})
	if rangeSlider.Low != 2 || rangeSlider.High != 8 {
		t.Fatalf("range slider initial range = [%v, %v], want snapped [2, 8]", rangeSlider.Low, rangeSlider.High)
	}
}

type widgetLayerRecordingRenderer struct {
	render.NullRenderer
	events []string
}

func (r *widgetLayerRecordingRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.events = append(r.events, "widget")
}

type widgetLayerDataArtist struct {
	events *[]string
	z      float64
}

func (a widgetLayerDataArtist) Draw(_ render.Renderer, _ *DrawContext) {
	*a.events = append(*a.events, "data")
}

func (a widgetLayerDataArtist) Z() float64 { return a.z }

func (a widgetLayerDataArtist) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func TestWidgetLayerDrawsAboveDataArtistZOrder(t *testing.T) {
	fig := NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	button := ax.Button("Run")
	button.z = -100
	rec := &widgetLayerRecordingRenderer{}
	ax.Add(widgetLayerDataArtist{events: &rec.events, z: 10000})

	DrawFigure(fig, rec)

	if len(rec.events) == 0 {
		t.Fatal("expected draw events")
	}
	if got := rec.events[len(rec.events)-1]; got != "widget" {
		t.Fatalf("topmost draw event = %q, want widget", got)
	}
}

func TestAddWidgetAxesHelpersPrepareAndCompose(t *testing.T) {
	fig := NewFigure(800, 600)
	fig.ConstrainedLayout()

	root := fig.AddWidgetAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.3, Y: 0.2},
	})
	if root == nil {
		t.Fatal("AddWidgetAxes returned nil")
	}
	assertWidgetAxesPrepared(t, root)

	sub := fig.AddSubFigure(geom.Rect{
		Min: geom.Pt{X: 0.5, Y: 0.2},
		Max: geom.Pt{X: 0.9, Y: 0.8},
	})
	subAx := sub.AddWidgetAxes(geom.Rect{
		Min: geom.Pt{X: 0.25, Y: 0.25},
		Max: geom.Pt{X: 0.75, Y: 0.75},
	})
	if subAx == nil {
		t.Fatal("SubFigure.AddWidgetAxes returned nil")
	}
	if subAx.RectFraction.Min.X != 0.6 || subAx.RectFraction.Max.X != 0.8 {
		t.Fatalf("subfigure widget axes rect = %+v, want composed x [0.6, 0.8]", subAx.RectFraction)
	}
	assertWidgetAxesPrepared(t, subAx)

	gs := fig.GridSpec(2, 1, WithGridSpecHeightRatios(8, 1))
	plot := gs.Cell(0, 0).AddAxes()
	widget := gs.Cell(1, 0).AddWidgetAxes()
	if plot == nil || widget == nil {
		t.Fatal("GridSpec widget composition returned nil axes")
	}
	if widget.subplotSpec == nil {
		t.Fatal("subplot-backed widget axes should keep subplot spec for constrained layout")
	}
	assertWidgetAxesPrepared(t, widget)
	DrawFigure(fig, &render.NullRenderer{})
	if widget.RectFraction.H() <= 0 || plot.RectFraction.H() <= 0 {
		t.Fatalf("constrained layout produced empty axes: plot=%+v widget=%+v", plot.RectFraction, widget.RectFraction)
	}
}

func assertWidgetAxesPrepared(t *testing.T, ax *Axes) {
	t.Helper()
	if ax.XAxis.ShowTicks || ax.YAxis.ShowTicks || ax.XAxis.ShowLabels || ax.YAxis.ShowLabels {
		t.Fatal("widget axes should hide ticks and labels")
	}
	if ax.XAxis.ShowSpine || ax.YAxis.ShowSpine || ax.ShowFrame {
		t.Fatal("widget axes should hide spines and frame")
	}
}

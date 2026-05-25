package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
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
	disabledChecks := axCheck.CheckButtons([]string{"C"}, []bool{false}, CheckButtonsOptions{Disabled: &disabled})
	if disabledChecks.Enabled {
		t.Fatal("check buttons should honor Disabled option")
	}
	disabledRadios := axRadio.RadioButtons([]string{"m", "n"}, 0, RadioButtonsOptions{Disabled: &disabled})
	if disabledRadios.Enabled {
		t.Fatal("radio buttons should honor Disabled option")
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

func TestWidgetConstructorsUseConfiguredVisualStyle(t *testing.T) {
	goFig := NewFigure(800, 600)
	goAx := goFig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.2}})
	goButton := goAx.Button("Apply")
	goSlider := goAx.Slider("gain", 0, 1, 0.5)

	mplFig := NewFigure(800, 600, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	mplAx := mplFig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.2}})
	mplButton := mplAx.Button("Apply")
	mplSlider := mplAx.Slider("gain", 0, 1, 0.5)
	mplRange := mplAx.RangeSlider("window", 0, 1, 0.25, 0.75)
	mplText := mplAx.TextBox("label", "phase scan")
	mplChecks := mplAx.CheckButtons([]string{"signal"}, []bool{true})
	mplRadios := mplAx.RadioButtons([]string{"blue", "amber"}, 1)

	if goButton.FaceColor == mplButton.FaceColor {
		t.Fatal("Matplotlib widget visual style should not replace the Go default style")
	}
	if goSlider.TrackColor == mplSlider.TrackColor {
		t.Fatal("Matplotlib slider visual style should differ from the Go default style")
	}
	assertColorEqual(t, mplButton.FaceColor, render.Color{R: 0.85, G: 0.85, B: 0.85, A: 1}, "Matplotlib button face")
	assertColorEqual(t, mplSlider.TrackColor, render.Color{R: 211.0 / 255.0, G: 211.0 / 255.0, B: 211.0 / 255.0, A: 1}, "Matplotlib slider track")
	assertColorEqual(t, mplSlider.FillColor, render.Color{R: 31.0 / 255.0, G: 119.0 / 255.0, B: 180.0 / 255.0, A: 1}, "Matplotlib slider fill")
	assertColorEqual(t, mplSlider.HandleColor, render.Color{R: 1, G: 1, B: 1, A: 1}, "Matplotlib slider handle")
	assertColorEqual(t, mplRange.TrackColor, mplSlider.TrackColor, "Matplotlib range slider track")
	assertColorEqual(t, mplText.FaceColor, render.Color{R: 0.95, G: 0.95, B: 0.95, A: 1}, "Matplotlib text box face")
	assertColorEqual(t, mplChecks.CheckColor, render.Color{R: 0, G: 0, B: 0, A: 1}, "Matplotlib check mark")
	assertColorEqual(t, mplRadios.DotColor, render.Color{R: 0, G: 0, B: 1, A: 1}, "Matplotlib radio active dot")
}

func assertColorEqual(t *testing.T, got, want render.Color, label string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s color = %+v, want %+v", label, got, want)
	}
}

func TestWidgetVisualStyleProvidesGeometryDefaults(t *testing.T) {
	goDefaults := widgetDefaultsForRC(style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualGo)))
	mplDefaults := widgetDefaultsForRC(style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib)))

	if goDefaults.ButtonRadius <= 0 {
		t.Fatalf("Go button radius = %v, want rounded native button chrome", goDefaults.ButtonRadius)
	}
	if mplDefaults.ButtonRadius != 0 {
		t.Fatalf("Matplotlib button radius = %v, want square panel chrome", mplDefaults.ButtonRadius)
	}
	if goDefaults.SliderPanelPad == mplDefaults.SliderPanelPad {
		t.Fatalf("slider panel padding should differ by visual style: go=%v mpl=%v", goDefaults.SliderPanelPad, mplDefaults.SliderPanelPad)
	}
	if mplDefaults.SliderTrackYMin != 0.25 || mplDefaults.SliderTrackYMax != 0.75 {
		t.Fatalf("Matplotlib slider track fractions = [%v, %v], want [0.25, 0.75]", mplDefaults.SliderTrackYMin, mplDefaults.SliderTrackYMax)
	}
	if mplDefaults.RadioOuterSize >= goDefaults.RadioOuterSize {
		t.Fatalf("Matplotlib radio size = %v, want smaller than Go native %v", mplDefaults.RadioOuterSize, goDefaults.RadioOuterSize)
	}
}

func TestStyledWidgetGeometryUsesVisualPolicy(t *testing.T) {
	clip := geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 210, Y: 80}}
	mplDefaults := widgetDefaultsForRC(style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib)))

	panel := widgetStyledPanelRect(clip, mplDefaults.SliderPanelPad)
	if panel != clip {
		t.Fatalf("Matplotlib slider panel = %+v, want full clip %+v", panel, clip)
	}
	track := widgetStyledSliderTrack(panel, mplDefaults)
	wantTrack := geom.Rect{Min: geom.Pt{X: 10, Y: 35}, Max: geom.Pt{X: 210, Y: 65}}
	if track != wantTrack {
		t.Fatalf("Matplotlib slider track = %+v, want %+v", track, wantTrack)
	}

	goDefaults := widgetDefaultsForRC(style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualGo)))
	goPanel := widgetStyledPanelRect(clip, goDefaults.SliderPanelPad)
	goTrack := widgetStyledSliderTrack(goPanel, goDefaults)
	if goTrack == track {
		t.Fatalf("Go and Matplotlib slider tracks should differ: %+v", track)
	}
}

func TestWidgetCallbackRegistryPreservesRegistrationOrder(t *testing.T) {
	var got []int
	var ordered widgetCallbackRegistry[func()]
	for i := 0; i < 24; i++ {
		value := i
		id := ordered.add(func() {
			got = append(got, value)
		})
		if i%5 == 0 {
			ordered.remove(id)
		}
	}
	ordered.each(func(cb func()) { cb() })

	want := []int{1, 2, 3, 4, 6, 7, 8, 9, 11, 12, 13, 14, 16, 17, 18, 19, 21, 22, 23}
	if len(got) != len(want) {
		t.Fatalf("callback order length = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("callback order = %v, want %v", got, want)
		}
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

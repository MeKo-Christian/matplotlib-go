package widgets

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestWidgetConstructorsPrepareAxesAndStoreState(t *testing.T) {
	fig := core.NewFigure(800, 600)
	axButton := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.3, Y: 0.2}})
	axSlider := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.25}, Max: geom.Pt{X: 0.5, Y: 0.38}})
	axRange := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.58}, Max: geom.Pt{X: 0.5, Y: 0.70}})
	axCheck := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.55, Y: 0.1}, Max: geom.Pt{X: 0.8, Y: 0.32}})
	axRadio := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.82, Y: 0.1}, Max: geom.Pt{X: 0.95, Y: 0.32}})
	axText := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.42}, Max: geom.Pt{X: 0.55, Y: 0.56}})

	pressed := true
	active := true
	button := NewButton(axButton, "Run", ButtonOptions{Pressed: &pressed})
	slider := NewSlider(axSlider, "gain", 0, 10, 12)
	rangeSlider := NewRangeSlider(axRange, "window", 0, 10, 8, 2)
	checks := NewCheckButtons(axCheck, []string{"A", "B"}, []bool{true, false})
	radios := NewRadioButtons(axRadio, []string{"x", "y", "z"}, 5)
	text := NewTextBox(axText, "Query", "", TextBoxOptions{Placeholder: "type...", Active: &active})

	if button == nil || slider == nil || rangeSlider == nil || checks == nil || radios == nil || text == nil {
		t.Fatal("expected widget constructors to return artists")
	}
	if !button.Pressed {
		t.Fatal("button should store pressed state")
	}
	disabled := true
	disabledButton := NewButton(axButton, "Stop", ButtonOptions{Disabled: &disabled})
	if disabledButton.Enabled {
		t.Fatal("button should honor Disabled option")
	}
	disabledChecks := NewCheckButtons(axCheck, []string{"C"}, []bool{false}, CheckButtonsOptions{Disabled: &disabled})
	if disabledChecks.Enabled {
		t.Fatal("check buttons should honor Disabled option")
	}
	disabledRadios := NewRadioButtons(axRadio, []string{"m", "n"}, 0, RadioButtonsOptions{Disabled: &disabled})
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
	for _, ax := range []*core.Axes{axButton, axSlider, axRange, axCheck, axRadio, axText} {
		if ax.XAxis.ShowTicks || ax.YAxis.ShowTicks || ax.XAxis.ShowLabels || ax.YAxis.ShowLabels {
			t.Fatal("widget constructors should hide axis ticks and labels")
		}
	}
}

func TestSliderConstructorsSnapInitialValuesToStep(t *testing.T) {
	fig := core.NewFigure(800, 600)
	axSlider := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.8, Y: 0.2}})
	axRange := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.3}, Max: geom.Pt{X: 0.8, Y: 0.4}})

	step := 0.5
	slider := NewSlider(axSlider, "gain", 0, 10, 5.26, SliderOptions{ValueStep: &step})
	if slider.Value != 5.5 {
		t.Fatalf("slider initial value = %v, want snapped 5.5", slider.Value)
	}

	rangeSlider := NewRangeSlider(axRange, "window", 0, 10, 7.76, 2.24, RangeSliderOptions{ValueStep: &step})
	if rangeSlider.Low != 2 || rangeSlider.High != 8 {
		t.Fatalf("range slider initial range = [%v, %v], want snapped [2, 8]", rangeSlider.Low, rangeSlider.High)
	}
}

func TestWidgetConstructorsUseConfiguredVisualStyle(t *testing.T) {
	goFig := core.NewFigure(800, 600)
	goAx := goFig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.2}})
	goButton := NewButton(goAx, "Apply")
	goSlider := NewSlider(goAx, "gain", 0, 1, 0.5)

	mplFig := core.NewFigure(800, 600, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	mplAx := mplFig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.2}})
	mplButton := NewButton(mplAx, "Apply")
	mplSlider := NewSlider(mplAx, "gain", 0, 1, 0.5)
	mplRange := NewRangeSlider(mplAx, "window", 0, 1, 0.25, 0.75)
	mplText := NewTextBox(mplAx, "label", "phase scan")
	mplChecks := NewCheckButtons(mplAx, []string{"signal"}, []bool{true})
	mplRadios := NewRadioButtons(mplAx, []string{"blue", "amber"}, 1)

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

func approx(a, b, tolerance float64) bool {
	if a > b {
		return a-b <= tolerance
	}
	return b-a <= tolerance
}

func approxRect(a, b geom.Rect, tolerance float64) bool {
	return approx(a.Min.X, b.Min.X, tolerance) &&
		approx(a.Min.Y, b.Min.Y, tolerance) &&
		approx(a.Max.X, b.Max.X, tolerance) &&
		approx(a.Max.Y, b.Max.Y, tolerance)
}

func TestWidgetVisualStyleProvidesGeometryDefaults(t *testing.T) {
	goRC := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualGo))
	mplRC := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	goDefaults := widgetDefaultsForRC(&goRC)
	mplDefaults := widgetDefaultsForRC(&mplRC)

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
	if !approx(mplDefaults.SliderHandleSize, pointsToPixels(&mplRC, 10), 1e-12) {
		t.Fatalf("Matplotlib slider handle size = %v, want 10 pt in pixels", mplDefaults.SliderHandleSize)
	}
	if !approx(mplDefaults.SliderHandleLine, pointsToPixels(&mplRC, 1), 1e-12) {
		t.Fatalf("Matplotlib slider handle edge width = %v, want 1 pt in pixels", mplDefaults.SliderHandleLine)
	}
	if !approx(mplDefaults.SliderInitLine, pointsToPixels(&mplRC, 1), 1e-12) {
		t.Fatalf("Matplotlib slider init line width = %v, want 1 pt in pixels", mplDefaults.SliderInitLine)
	}
	if !approx(mplDefaults.CheckBoxMaxSize, pointsToPixels(&mplRC, mplRC.FontSize/2), 1e-12) {
		t.Fatalf("Matplotlib checkbox marker size = %v, want half default font size in pixels", mplDefaults.CheckBoxMaxSize)
	}
	if !approx(mplDefaults.CheckBoxLineWidth, pointsToPixels(&mplRC, 1), 1e-12) {
		t.Fatalf("Matplotlib checkbox frame width = %v, want 1 pt in pixels", mplDefaults.CheckBoxLineWidth)
	}
	if !approx(mplDefaults.CheckMarkWidth, pointsToPixels(&mplRC, 1), 1e-12) {
		t.Fatalf("Matplotlib check mark width = %v, want 1 pt in pixels", mplDefaults.CheckMarkWidth)
	}
	if !approx(mplDefaults.RadioOuterSize, pointsToPixels(&mplRC, mplRC.FontSize/2), 1e-12) {
		t.Fatalf("Matplotlib radio marker size = %v, want half default font size in pixels", mplDefaults.RadioOuterSize)
	}
	if !approx(mplDefaults.RadioLineWidth, pointsToPixels(&mplRC, 1), 1e-12) {
		t.Fatalf("Matplotlib radio edge width = %v, want 1 pt in pixels", mplDefaults.RadioLineWidth)
	}
	if mplDefaults.RadioOuterSize >= goDefaults.RadioOuterSize {
		t.Fatalf("Matplotlib radio size = %v, want smaller than Go native %v", mplDefaults.RadioOuterSize, goDefaults.RadioOuterSize)
	}
}

func TestStyledWidgetGeometryUsesVisualPolicy(t *testing.T) {
	clip := geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 210, Y: 80}}
	mplRC := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	mplDefaults := widgetDefaultsForRC(&mplRC)

	panel := widgetStyledPanelRect(clip, mplDefaults.SliderPanelPad)
	if panel != clip {
		t.Fatalf("Matplotlib slider panel = %+v, want full clip %+v", panel, clip)
	}
	track := widgetStyledSliderTrack(panel, &mplDefaults)
	wantTrack := geom.Rect{Min: geom.Pt{X: 10, Y: 35}, Max: geom.Pt{X: 210, Y: 65}}
	if track != wantTrack {
		t.Fatalf("Matplotlib slider track = %+v, want %+v", track, wantTrack)
	}

	goRC := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualGo))
	goDefaults := widgetDefaultsForRC(&goRC)
	goPanel := widgetStyledPanelRect(clip, goDefaults.SliderPanelPad)
	goTrack := widgetStyledSliderTrack(goPanel, &goDefaults)
	if goTrack == track {
		t.Fatalf("Go and Matplotlib slider tracks should differ: %+v", track)
	}
}

func TestWidgetHelperGeometryIsRendererNeutral(t *testing.T) {
	clip := geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 210, Y: 80}}
	goRC := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualGo))
	mplRC := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	goDefaults := widgetDefaultsForRC(&goRC)
	mplDefaults := widgetDefaultsForRC(&mplRC)

	goPanel := widgetStyledPanelRect(clip, goDefaults.SliderPanelPad)
	if !approxRect(goPanel, geom.Rect{Min: geom.Pt{X: 14, Y: 24}, Max: geom.Pt{X: 206, Y: 76}}, 1e-9) {
		t.Fatalf("Go styled panel = %+v, want inset by native panel padding", goPanel)
	}
	goTrack := widgetStyledSliderTrack(goPanel, &goDefaults)
	if !approxRect(goTrack, geom.Rect{Min: geom.Pt{X: 28, Y: 50}, Max: geom.Pt{X: 192, Y: 62}}, 1e-9) {
		t.Fatalf("Go slider track = %+v, want native track geometry", goTrack)
	}
	if got, want := widgetSliderTrackRadius(goTrack, &goDefaults), goTrack.H()/2; got != want {
		t.Fatalf("Go slider track radius = %v, want half track height %v", got, want)
	}

	mplPanel := widgetStyledPanelRect(clip, mplDefaults.SliderPanelPad)
	if mplPanel != clip {
		t.Fatalf("Matplotlib styled panel = %+v, want full clip %+v", mplPanel, clip)
	}
	mplTrack := widgetStyledSliderTrack(mplPanel, &mplDefaults)
	if !approxRect(mplTrack, geom.Rect{Min: geom.Pt{X: 10, Y: 35}, Max: geom.Pt{X: 210, Y: 65}}, 1e-9) {
		t.Fatalf("Matplotlib slider track = %+v, want source fraction geometry", mplTrack)
	}
	if got := widgetSliderTrackRadius(mplTrack, &mplDefaults); got != 0 {
		t.Fatalf("Matplotlib slider track radius = %v, want rectangular track", got)
	}
}

func TestWidgetRowAndCoordinateHelpersAreRendererNeutral(t *testing.T) {
	panel := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 120, Y: 90}}

	if got, want := widgetButtonRowCenterY(panel, 0, 3, 0.15), 67.5; got != want {
		t.Fatalf("source-style row 0 center = %v, want %v", got, want)
	}
	if got, want := widgetButtonRowCenterY(panel, 1, 3, 0.15), 45.0; got != want {
		t.Fatalf("source-style row 1 center = %v, want %v", got, want)
	}
	if got, want := widgetButtonRowCenterY(panel, 0, 3, 14), 75.0; got != want {
		t.Fatalf("native row 0 center = %v, want %v", got, want)
	}
	if got, want := widgetStyleCoord(panel.Min.X, panel.Max.X, 0.25), 30.0; got != want {
		t.Fatalf("fraction style coord = %v, want %v", got, want)
	}
	if got, want := widgetStyleCoord(panel.Min.X, panel.Max.X, -8), 112.0; got != want {
		t.Fatalf("negative pixel style coord = %v, want %v", got, want)
	}
	if got, want := widgetResolvedCoord(panel.Min.X, panel.Max.X, widgetFractionCoord(-0.02)), -2.4; got != want {
		t.Fatalf("resolved fractional coord = %v, want %v", got, want)
	}
}

func TestMatplotlibSliderTrackUsesRectangularPatch(t *testing.T) {
	rc := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	ctx := &core.DrawContext{
		RC:   rc,
		Clip: geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 210, Y: 80}},
	}
	slider := &Slider{
		Label:       "gain",
		Min:         0,
		Max:         1,
		Value:       0.5,
		Enabled:     true,
		FaceColor:   render.Color{R: 1, G: 1, B: 1, A: 1},
		TrackColor:  render.Color{R: 211.0 / 255.0, G: 211.0 / 255.0, B: 211.0 / 255.0, A: 1},
		FillColor:   render.Color{R: 31.0 / 255.0, G: 119.0 / 255.0, B: 180.0 / 255.0, A: 1},
		HandleColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		TextColor:   render.Color{A: 1},
		ValueFormat: "%.2f",
	}
	rec := &widgetChromeRecordingRenderer{}

	slider.Draw(rec, ctx)

	track, ok := rec.pathWithFill(slider.TrackColor)
	if !ok {
		t.Fatal("Matplotlib slider track path not recorded")
	}
	if hasCurveCommand(track.path) {
		t.Fatalf("Matplotlib slider track should be rectangular, got curved path commands %v", track.path.C)
	}
	if bounds, ok := track.path.Bounds(); !ok || !approxRect(bounds, geom.Rect{Min: geom.Pt{X: 10, Y: 35}, Max: geom.Pt{X: 210, Y: 65}}, 1e-9) {
		t.Fatalf("Matplotlib slider track bounds = %+v, want x full axes and y [0.25,0.75]", bounds)
	}
}

func TestMatplotlibWidgetSquarePanelUsesSnapAuto(t *testing.T) {
	rec := &widgetChromeRecordingRenderer{}
	rect := geom.Rect{Min: geom.Pt{X: 10.2, Y: 20.3}, Max: geom.Pt{X: 110.2, Y: 50.3}}

	drawWidgetPanel(rec, rect, render.Color{R: 1, G: 1, B: 1, A: 1}, render.Color{A: 1}, 1, 0)

	if len(rec.paths) != 1 {
		t.Fatalf("widget panel paths = %d, want 1", len(rec.paths))
	}
	call := rec.paths[0]
	if call.paint.Snap != render.SnapAuto {
		t.Fatalf("square widget panel snap = %v, want Matplotlib SnapAuto", call.paint.Snap)
	}
	if bounds, ok := call.path.Bounds(); !ok || !approxRect(bounds, rect, 1e-9) {
		t.Fatalf("square widget panel bounds = %+v, want true patch bounds %+v", bounds, rect)
	}
}

func TestMatplotlibCheckButtonsUseSourceFractionLayout(t *testing.T) {
	rc := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	ctx := &core.DrawContext{
		RC:   rc,
		Clip: geom.Rect{Min: geom.Pt{X: 66, Y: 53.2}, Max: geom.Pt{X: 462, Y: 144.4}},
	}
	checks := &CheckButtons{
		Labels:     []string{"signal", "modulation", "grid"},
		Values:     []bool{true, true, false},
		Enabled:    true,
		FaceColor:  render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor:  render.Color{A: 1},
		TextColor:  render.Color{A: 1},
		CheckColor: render.Color{A: 1},
	}
	rec := &widgetChromeRecordingRenderer{}

	checks.Draw(rec, ctx)

	signal := rec.textCall("signal")
	wantLabelX := ctx.Clip.Min.X + ctx.Clip.W()*0.25
	if signal.origin.X < wantLabelX-1e-9 || signal.origin.X > wantLabelX+1e-9 {
		t.Fatalf("Matplotlib CheckButtons label X = %.2f, want %.2f", signal.origin.X, wantLabelX)
	}
	marker, ok := rec.smallStrokePathNear(ctx.Clip.Min.X+ctx.Clip.W()*0.15, ctx.Clip.Min.Y+ctx.Clip.H()*0.75)
	if !ok {
		t.Fatalf("Matplotlib CheckButtons marker not near source position x=.15 y=.75; paths=%+v", rec.paths)
	}
	if bounds, _ := marker.path.Bounds(); rectCenter(bounds).X < ctx.Clip.Min.X+ctx.Clip.W()*0.15-1 || rectCenter(bounds).X > ctx.Clip.Min.X+ctx.Clip.W()*0.15+1 {
		t.Fatalf("Matplotlib CheckButtons marker center X = %.2f, want near %.2f", rectCenter(bounds).X, ctx.Clip.Min.X+ctx.Clip.W()*0.15)
	}
}

func TestMatplotlibRadioButtonsUseSourceFractionLayout(t *testing.T) {
	rc := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	ctx := &core.DrawContext{
		RC:   rc,
		Clip: geom.Rect{Min: geom.Pt{X: 605, Y: 53.2}, Max: geom.Pt{X: 913, Y: 144.4}},
	}
	radios := &RadioButtons{
		Labels:    []string{"blue", "amber", "mono"},
		Active:    1,
		Enabled:   true,
		FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor: render.Color{A: 1},
		TextColor: render.Color{A: 1},
		DotColor:  render.Color{B: 1, A: 1},
	}
	rec := &widgetChromeRecordingRenderer{}

	radios.Draw(rec, ctx)

	blue := rec.textCall("blue")
	wantLabelX := ctx.Clip.Min.X + ctx.Clip.W()*0.25
	if blue.origin.X < wantLabelX-1e-9 || blue.origin.X > wantLabelX+1e-9 {
		t.Fatalf("Matplotlib RadioButtons label X = %.2f, want %.2f", blue.origin.X, wantLabelX)
	}
	marker, ok := rec.smallStrokePathNear(ctx.Clip.Min.X+ctx.Clip.W()*0.15, ctx.Clip.Min.Y+ctx.Clip.H()*0.75)
	if !ok {
		t.Fatalf("Matplotlib RadioButtons marker not near source position x=.15 y=.75; paths=%+v", rec.paths)
	}
	if bounds, _ := marker.path.Bounds(); rectCenter(bounds).Y < ctx.Clip.Min.Y+ctx.Clip.H()*0.75-1 || rectCenter(bounds).Y > ctx.Clip.Min.Y+ctx.Clip.H()*0.75+1 {
		t.Fatalf("Matplotlib RadioButtons marker center Y = %.2f, want near %.2f", rectCenter(bounds).Y, ctx.Clip.Min.Y+ctx.Clip.H()*0.75)
	}
}

func TestMatplotlibSliderTextAnchorsMatchAxesLayout(t *testing.T) {
	rc := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	ctx := &core.DrawContext{
		RC:   rc,
		Clip: geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 110, Y: 60}},
	}
	slider := &Slider{
		Label:       "gain",
		Min:         0,
		Max:         1,
		Value:       0.5,
		Enabled:     true,
		FaceColor:   render.Color{R: 1, G: 1, B: 1, A: 1},
		TrackColor:  render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1},
		FillColor:   render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1},
		HandleColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		TextColor:   render.Color{A: 1},
		ValueFormat: "%.2f",
	}
	rec := &widgetChromeRecordingRenderer{}

	slider.Draw(rec, ctx)

	label := rec.textCall("gain")
	value := rec.textCall("0.50")
	if label.origin.X >= ctx.Clip.Min.X {
		t.Fatalf("Matplotlib slider label origin X = %.2f, want left of axes min %.2f", label.origin.X, ctx.Clip.Min.X)
	}
	if value.origin.X <= ctx.Clip.Max.X {
		t.Fatalf("Matplotlib slider value origin X = %.2f, want right of axes max %.2f", value.origin.X, ctx.Clip.Max.X)
	}
	if label.origin.Y < ctx.Clip.Min.Y || label.origin.Y > ctx.Clip.Max.Y {
		t.Fatalf("Matplotlib slider label origin Y = %.2f, want vertically centered in axes %+v", label.origin.Y, ctx.Clip)
	}
}

func TestGoSliderTextAnchorsStayInsidePanel(t *testing.T) {
	rc := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualGo))
	ctx := &core.DrawContext{
		RC:   rc,
		Clip: geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 110, Y: 60}},
	}
	slider := &Slider{
		Label:       "gain",
		Min:         0,
		Max:         1,
		Value:       0.5,
		Enabled:     true,
		FaceColor:   render.Color{R: 1, G: 1, B: 1, A: 1},
		TrackColor:  render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1},
		FillColor:   render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1},
		HandleColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		TextColor:   render.Color{A: 1},
		ValueFormat: "%.2f",
	}
	rec := &widgetChromeRecordingRenderer{}

	slider.Draw(rec, ctx)

	label := rec.textCall("gain")
	value := rec.textCall("0.50")
	if label.origin.X <= ctx.Clip.Min.X || value.origin.X >= ctx.Clip.Max.X {
		t.Fatalf("Go slider text should remain inside panel, got label X %.2f value X %.2f clip %+v", label.origin.X, value.origin.X, ctx.Clip)
	}
}

func TestMatplotlibTextBoxTextAnchorsMatchAxesLayout(t *testing.T) {
	rc := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	ctx := &core.DrawContext{
		RC:   rc,
		Clip: geom.Rect{Min: geom.Pt{X: 20, Y: 30}, Max: geom.Pt{X: 120, Y: 70}},
	}
	textBox := &TextBox{
		Label:     "label",
		Value:     "phase",
		Active:    true,
		FaceColor: render.Color{R: 0.95, G: 0.95, B: 0.95, A: 1},
		EdgeColor: render.Color{A: 1},
		TextColor: render.Color{A: 1},
	}
	rec := &widgetChromeRecordingRenderer{}

	textBox.Draw(rec, ctx)

	label := rec.textCall("label")
	value := rec.textCall("phase")
	if label.origin.X >= ctx.Clip.Min.X {
		t.Fatalf("Matplotlib textbox label origin X = %.2f, want outside left of %.2f", label.origin.X, ctx.Clip.Min.X)
	}
	wantValueX := ctx.Clip.Min.X + ctx.Clip.W()*0.05
	if value.origin.X < wantValueX-2 || value.origin.X > wantValueX+2 {
		t.Fatalf("Matplotlib textbox value origin X = %.2f, want near %.2f", value.origin.X, wantValueX)
	}
}

func TestMatplotlibButtonHoverUsesPolicyFaceColor(t *testing.T) {
	rc := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	ctx := &core.DrawContext{
		RC:   rc,
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 40}},
	}
	button := &Button{
		Label:     "Apply",
		FaceColor: render.Color{R: 0.85, G: 0.85, B: 0.85, A: 1},
		EdgeColor: render.Color{A: 1},
		TextColor: render.Color{A: 1},
		Enabled:   true,
		Hovered:   true,
	}
	rec := &widgetChromeRecordingRenderer{}

	button.Draw(rec, ctx)

	if len(rec.paths) == 0 {
		t.Fatal("button draw emitted no paths")
	}
	want := render.Color{R: 0.95, G: 0.95, B: 0.95, A: 1}
	if got := rec.paths[0].paint.Fill; got != want {
		t.Fatalf("Matplotlib button hover fill = %+v, want %+v", got, want)
	}
}

func TestMatplotlibCheckAndRadioUseFilledActiveMarkers(t *testing.T) {
	rc := style.Apply(style.Default, style.WithWidgetVisualStyle(style.WidgetVisualMatplotlib))
	ctx := &core.DrawContext{
		RC:   rc,
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 120, Y: 90}},
	}

	checks := &CheckButtons{
		Labels:     []string{"signal", "grid"},
		Values:     []bool{true, false},
		Enabled:    true,
		FaceColor:  render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor:  render.Color{A: 1},
		TextColor:  render.Color{A: 1},
		CheckColor: render.Color{A: 1},
	}
	checkRec := &widgetChromeRecordingRenderer{}
	checks.Draw(checkRec, ctx)
	if checkRec.smallWhiteFilledPaths() != 0 {
		t.Fatalf("Matplotlib check buttons should use unfilled square frames, got %d small white filled paths", checkRec.smallWhiteFilledPaths())
	}
	if checkRec.smallStrokeOnlyLineMarkers() == 0 {
		t.Fatal("Matplotlib check buttons should draw an active x marker")
	}

	radios := &RadioButtons{
		Labels:    []string{"blue", "amber"},
		Active:    1,
		Enabled:   true,
		FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor: render.Color{A: 1},
		TextColor: render.Color{A: 1},
		DotColor:  render.Color{B: 1, A: 1},
	}
	radioRec := &widgetChromeRecordingRenderer{}
	radios.Draw(radioRec, ctx)
	if radioRec.smallWhiteFilledPaths() != 0 {
		t.Fatalf("Matplotlib radio buttons should use unfilled inactive markers, got %d small white filled paths", radioRec.smallWhiteFilledPaths())
	}
	if radioRec.smallFilledPaths(radios.DotColor) != 1 {
		t.Fatalf("Matplotlib radio buttons should draw one active filled marker, got %d", radioRec.smallFilledPaths(radios.DotColor))
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

type widgetChromeTextCall struct {
	text   string
	origin geom.Pt
	size   float64
}

type widgetChromePathCall struct {
	path  geom.Path
	paint render.Paint
}

type widgetChromeRecordingRenderer struct {
	render.NullRenderer
	texts []widgetChromeTextCall
	paths []widgetChromePathCall
}

func (r *widgetChromeRecordingRenderer) DrawText(text string, origin geom.Pt, size float64, _ render.Color) {
	r.texts = append(r.texts, widgetChromeTextCall{text: text, origin: origin, size: size})
}

func (r *widgetChromeRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{W: float64(len([]rune(text))) * size * 0.5, H: size}
}

func (r *widgetChromeRecordingRenderer) Path(path geom.Path, paint *render.Paint) {
	call := widgetChromePathCall{path: path}
	if paint != nil {
		call.paint = *paint
	}
	r.paths = append(r.paths, call)
}

func (r *widgetChromeRecordingRenderer) textCall(text string) widgetChromeTextCall {
	for _, call := range r.texts {
		if call.text == text {
			return call
		}
	}
	return widgetChromeTextCall{}
}

func (r *widgetChromeRecordingRenderer) pathWithFill(color render.Color) (widgetChromePathCall, bool) {
	for i := range r.paths {
		call := &r.paths[i]
		if call.paint.Fill == color {
			return *call, true
		}
	}
	return widgetChromePathCall{}, false
}

func (r *widgetChromeRecordingRenderer) smallStrokePathNear(x, y float64) (widgetChromePathCall, bool) {
	for i := range r.paths {
		call := &r.paths[i]
		bounds, ok := call.path.Bounds()
		if !ok || bounds.W() > 24 || bounds.H() > 24 || call.paint.Stroke.A == 0 {
			continue
		}
		center := rectCenter(bounds)
		if center.X >= x-4 && center.X <= x+4 && center.Y >= y-4 && center.Y <= y+4 {
			return *call, true
		}
	}
	return widgetChromePathCall{}, false
}

func hasCurveCommand(path geom.Path) bool {
	for _, cmd := range path.C {
		if cmd == geom.QuadTo || cmd == geom.CubicTo {
			return true
		}
	}
	return false
}

func (r *widgetChromeRecordingRenderer) smallWhiteFilledPaths() int {
	return r.smallFilledPaths(render.Color{R: 1, G: 1, B: 1, A: 1})
}

func (r *widgetChromeRecordingRenderer) smallFilledPaths(color render.Color) int {
	n := 0
	for i := range r.paths {
		call := &r.paths[i]
		bounds, ok := call.path.Bounds()
		if !ok || bounds.W() > 24 || bounds.H() > 24 {
			continue
		}
		if call.paint.Fill == color {
			n++
		}
	}
	return n
}

func (r *widgetChromeRecordingRenderer) smallStrokeOnlyLineMarkers() int {
	n := 0
	for i := range r.paths {
		call := &r.paths[i]
		bounds, ok := call.path.Bounds()
		if !ok || bounds.W() > 24 || bounds.H() > 24 {
			continue
		}
		if call.paint.Fill.A == 0 && call.paint.Stroke.A > 0 && len(call.path.C) >= 4 {
			n++
		}
	}
	return n
}

type widgetLayerDataArtist struct {
	events *[]string
	z      float64
}

func (a widgetLayerDataArtist) Draw(_ render.Renderer, _ *core.DrawContext) {
	*a.events = append(*a.events, "data")
}

func (a widgetLayerDataArtist) Z() float64 { return a.z }

func (a widgetLayerDataArtist) Bounds(*core.DrawContext) geom.Rect { return geom.Rect{} }

func TestWidgetLayerDrawsAboveDataArtistZOrder(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	button := NewButton(ax, "Run")
	button.z = -100
	rec := &widgetLayerRecordingRenderer{}
	ax.Add(widgetLayerDataArtist{events: &rec.events, z: 10000})

	core.DrawFigure(fig, rec)

	if len(rec.events) == 0 {
		t.Fatal("expected draw events")
	}
	if got := rec.events[len(rec.events)-1]; got != "widget" {
		t.Fatalf("topmost draw event = %q, want widget", got)
	}
}

func TestAddWidgetAxesHelpersPrepareAndCompose(t *testing.T) {
	fig := core.NewFigure(800, 600)
	fig.ConstrainedLayout()

	root := NewAxes(fig, geom.Rect{
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
	subAx := NewSubFigureAxes(sub, geom.Rect{
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

	gs := fig.GridSpec(2, 1, core.WithGridSpecHeightRatios(8, 1))
	plot := gs.Cell(0, 0).AddAxes()
	widget := NewSubplotAxes(gs.Cell(1, 0))
	if plot == nil || widget == nil {
		t.Fatal("GridSpec widget composition returned nil axes")
	}
	assertWidgetAxesPrepared(t, widget)
	core.DrawFigure(fig, &render.NullRenderer{})
	if widget.RectFraction.H() <= 0 || plot.RectFraction.H() <= 0 {
		t.Fatalf("constrained layout produced empty axes: plot=%+v widget=%+v", plot.RectFraction, widget.RectFraction)
	}
}

func assertWidgetAxesPrepared(t *testing.T, ax *core.Axes) {
	t.Helper()
	if ax.XAxis.ShowTicks || ax.YAxis.ShowTicks || ax.XAxis.ShowLabels || ax.YAxis.ShowLabels {
		t.Fatal("widget axes should hide ticks and labels")
	}
	if ax.XAxis.ShowSpine || ax.YAxis.ShowSpine || ax.ShowFrame {
		t.Fatal("widget axes should hide spines and frame")
	}
}

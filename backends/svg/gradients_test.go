package svg

import (
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestCapabilityAdvertisesPatternAndGradient(t *testing.T) {
	r := mustNewRenderer(t)
	if !r.SupportsGradientFill() {
		t.Fatal("SVG renderer should advertise SupportsGradientFill")
	}
	if !r.SupportsPatternFill() {
		t.Fatal("SVG renderer should advertise SupportsPatternFill")
	}
}

func TestPathWithLinearGradientEmitsDef(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 10, Y: 10})
		path.LineTo(geom.Pt{X: 100, Y: 10})
		path.LineTo(geom.Pt{X: 100, Y: 50})
		path.LineTo(geom.Pt{X: 10, Y: 50})
		path.Close()

		r.Path(path, &render.Paint{
			FillGradient: render.GradientFill{
				Kind:  render.LinearGradient,
				Start: geom.Pt{X: 10, Y: 30},
				End:   geom.Pt{X: 100, Y: 30},
				Stops: []render.GradientStop{
					{Offset: 0, Color: render.Color{R: 1, A: 1}},
					{Offset: 1, Color: render.Color{B: 1, A: 1}},
				},
			},
		})
	})

	for _, want := range []string{
		`<linearGradient id="gradient1" gradientUnits="userSpaceOnUse" x1="10" y1="30" x2="100" y2="30">`,
		`<stop offset="0" stop-color="rgb(255,0,0)" />`,
		`<stop offset="1" stop-color="rgb(0,0,255)" />`,
		`fill="url(#gradient1)"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected gradient fragment %q in:\n%s", want, content)
		}
	}
}

func TestPathWithRadialGradientEmitsDef(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 60, Y: 60})
		path.LineTo(geom.Pt{X: 100, Y: 60})
		path.LineTo(geom.Pt{X: 100, Y: 100})
		path.Close()

		r.Path(path, &render.Paint{
			FillGradient: render.GradientFill{
				Kind:   render.RadialGradient,
				Center: geom.Pt{X: 80, Y: 80},
				Radius: 30,
				Stops: []render.GradientStop{
					{Offset: 0, Color: render.Color{R: 1, G: 1, B: 1, A: 1}},
					{Offset: 1, Color: render.Color{A: 1}},
				},
			},
		})
	})

	for _, want := range []string{
		`<radialGradient id="gradient1" gradientUnits="userSpaceOnUse" cx="80" cy="80" r="30">`,
		`<stop offset="0" stop-color="rgb(255,255,255)" />`,
		`<stop offset="1" stop-color="rgb(0,0,0)" />`,
		`fill="url(#gradient1)"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected radial gradient fragment %q in:\n%s", want, content)
		}
	}
}

func TestGradientDefsAreReusedAcrossPaths(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		grad := render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 0, Y: 0},
			End:   geom.Pt{X: 50, Y: 50},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{G: 1, A: 1}},
			},
		}
		var path geom.Path
		path.MoveTo(geom.Pt{X: 0, Y: 0})
		path.LineTo(geom.Pt{X: 50, Y: 0})
		path.LineTo(geom.Pt{X: 50, Y: 50})
		path.Close()

		r.Path(path, &render.Paint{FillGradient: grad})
		r.Path(path, &render.Paint{FillGradient: grad})
	})

	if got := strings.Count(content, `<linearGradient id="gradient`); got != 1 {
		t.Fatalf("identical gradients should share one def, got %d in:\n%s", got, content)
	}
	if got := strings.Count(content, `fill="url(#gradient1)"`); got != 2 {
		t.Fatalf("both paths should reference shared gradient def, got %d in:\n%s", got, content)
	}
}

func TestPathWithPatternFillEmitsDef(t *testing.T) {
	var dot geom.Path
	dot.MoveTo(geom.Pt{X: 4, Y: 4})
	dot.LineTo(geom.Pt{X: 12, Y: 4})
	dot.LineTo(geom.Pt{X: 12, Y: 12})
	dot.LineTo(geom.Pt{X: 4, Y: 12})
	dot.Close()

	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 20, Y: 20})
		path.LineTo(geom.Pt{X: 120, Y: 20})
		path.LineTo(geom.Pt{X: 120, Y: 80})
		path.LineTo(geom.Pt{X: 20, Y: 80})
		path.Close()

		r.Path(path, &render.Paint{
			FillPattern: render.PatternFill{
				ID:         "dots",
				Cell:       geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 16, Y: 16}},
				Path:       dot,
				Foreground: render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1},
				Background: render.Color{R: 1, G: 1, B: 1, A: 1},
			},
		})
	})

	for _, want := range []string{
		`<pattern id="pattern1" patternUnits="userSpaceOnUse" x="0" y="0" width="16" height="16">`,
		`<rect x="0" y="0" width="16" height="16" fill="rgb(255,255,255)" />`,
		`fill="rgb(51,102,204)"`,
		`fill="url(#pattern1)"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected pattern fragment %q in:\n%s", want, content)
		}
	}
}

func TestGradientStopOpacityEmittedWhenAlphaPartial(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 0, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 10})
		path.Close()

		r.Path(path, &render.Paint{
			FillGradient: render.GradientFill{
				Kind:  render.LinearGradient,
				Start: geom.Pt{X: 0, Y: 0},
				End:   geom.Pt{X: 10, Y: 0},
				Stops: []render.GradientStop{
					{Offset: 0, Color: render.Color{R: 1, A: 0.25}},
					{Offset: 1, Color: render.Color{B: 1, A: 1}},
				},
			},
		})
	})

	if !strings.Contains(content, `stop-opacity="0.25"`) {
		t.Fatalf("expected partial stop-opacity in:\n%s", content)
	}
}

func TestHatchTakesPrecedenceOverGradient(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 0, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 10})
		path.Close()

		r.Path(path, &render.Paint{
			Hatch:          "/",
			HatchColor:     render.Color{A: 1},
			HatchLineWidth: 1,
			FillGradient: render.GradientFill{
				Kind:  render.LinearGradient,
				Start: geom.Pt{X: 0, Y: 0},
				End:   geom.Pt{X: 10, Y: 0},
				Stops: []render.GradientStop{
					{Offset: 0, Color: render.Color{R: 1, A: 1}},
					{Offset: 1, Color: render.Color{B: 1, A: 1}},
				},
			},
		})
	})

	if !strings.Contains(content, `fill="url(#hatch1)"`) {
		t.Fatalf("hatch should win against gradient fill in:\n%s", content)
	}
	if strings.Contains(content, `fill="url(#gradient`) {
		t.Fatalf("gradient ref should be skipped when hatch is set:\n%s", content)
	}
}

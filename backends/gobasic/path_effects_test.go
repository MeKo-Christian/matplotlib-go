package gobasic

import (
	"image"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestGoBasicPathEffectsRenderSemanticPasses(t *testing.T) {
	t.Run("normal draws original path", func(t *testing.T) {
		img := renderPathEffects([]render.PathEffect{render.NormalPathEffect()})
		if !isDark(img.RGBAAt(30, 30)) {
			t.Fatalf("normal path effect did not draw original fill at center: %+v", img.RGBAAt(30, 30))
		}
	})

	t.Run("stroke then normal draws offset stroke behind fill", func(t *testing.T) {
		img := renderPathEffects(render.WithStrokePathEffects(
			render.Color{R: 1, G: 0, B: 0, A: 1},
			10,
			geom.Pt{X: 18, Y: 0},
		))
		if got := img.RGBAAt(66, 30); got.R < 180 || got.G > 220 || got.B > 220 {
			t.Fatalf("stroke path effect did not draw red offset stroke, got %+v", got)
		}
		if !isDark(img.RGBAAt(30, 30)) {
			t.Fatalf("normal pass did not draw original fill over stroke: %+v", img.RGBAAt(30, 30))
		}
	})

	t.Run("path patch draws alternate offset fill", func(t *testing.T) {
		img := renderPathEffects([]render.PathEffect{
			render.PathPatchPathEffect(
				render.Color{R: 0, G: 1, B: 0, A: 1},
				render.Color{},
				0,
				geom.Pt{X: 18, Y: 0},
			),
			render.NormalPathEffect(),
		})
		if got := img.RGBAAt(51, 30); got.G < 180 || got.R > 80 || got.B > 80 {
			t.Fatalf("path-patch effect did not draw green offset fill, got %+v", got)
		}
	})

	t.Run("patch shadow derives visible offset shadow", func(t *testing.T) {
		img := renderPathEffects([]render.PathEffect{
			render.SimplePatchShadowPathEffect(geom.Pt{X: 18, Y: 0}, render.Color{}, 0.8, 0.2),
			render.NormalPathEffect(),
		})
		got := img.RGBAAt(51, 30)
		if got.R > 120 || got.G > 120 || got.B > 120 {
			t.Fatalf("patch shadow did not draw a dark offset fill, got %+v", got)
		}
	})

	t.Run("line shadow derives visible offset stroke", func(t *testing.T) {
		img := renderLinePathEffects([]render.PathEffect{
			render.SimpleLineShadowPathEffect(geom.Pt{X: 0, Y: 10}, render.Color{}, 0.8, 0.2),
			render.NormalPathEffect(),
		})
		if got := img.RGBAAt(45, 40); got.R > 120 || got.G > 120 || got.B > 120 {
			t.Fatalf("line shadow did not draw a dark offset stroke, got %+v", got)
		}
	})

	t.Run("ticked stroke builds tick geometry", func(t *testing.T) {
		img := renderLinePathEffects([]render.PathEffect{
			render.TickedStrokePathEffect(render.Color{B: 1, A: 1}, 2, 10, 90, 1, geom.Pt{}),
		})
		if got := countNonBackgroundPixels(img, image.Rect(10, 20, 85, 45), semanticWhite); got < 20 {
			t.Fatalf("ticked stroke drew too few pixels: %d", got)
		}
	})

	t.Run("filter falls back to deterministic repaint", func(t *testing.T) {
		r := New(80, 60, render.Color{R: 1, G: 1, B: 1, A: 1})
		if _, ok := any(r).(render.FilterRenderer); ok {
			t.Fatal("GoBasic should not expose offscreen filter support")
		}
		if _, ok := any(r).(render.PathEffectFilterDrawer); ok {
			t.Fatal("GoBasic should not expose native path-effect filters")
		}

		img := renderPathEffects([]render.PathEffect{
			render.FilterPathEffect(render.Color{R: 0, G: 0, B: 1, A: 1}, render.Color{}, 0, "blur", 4, geom.Pt{X: 18, Y: 0}),
			render.NormalPathEffect(),
		})
		if got := img.RGBAAt(51, 30); got.B < 180 || got.R > 80 || got.G > 80 {
			t.Fatalf("filter fallback did not repaint offset blue fill, got %+v", got)
		}
	})
}

func TestGoBasicArtistPathEffectsRouteThroughRendererNeutralPipeline(t *testing.T) {
	t.Run("line", func(t *testing.T) {
		r := newStartedSemanticRenderer(t, 100, 70)
		defer r.End()

		line := &core.Line2D{
			XY:          []geom.Pt{{X: 15, Y: 35}, {X: 85, Y: 35}},
			Col:         render.Color{R: 0, G: 0, B: 0, A: 1},
			W:           3,
			PathEffects: render.WithStrokePathEffects(render.Color{R: 1, G: 0, B: 0, A: 1}, 11, geom.Pt{X: 0, Y: 12}),
		}
		line.Draw(r, &core.DrawContext{})
		if got := r.Image().RGBAAt(50, 47); got.R < 180 || got.G > 90 || got.B > 90 {
			t.Fatalf("line path effect did not draw red offset stroke, got %+v", got)
		}
	})

	t.Run("patch", func(t *testing.T) {
		r := newStartedSemanticRenderer(t, 100, 70)
		defer r.End()

		patch := &core.Rectangle{
			Patch: core.Patch{
				FaceColor: render.Color{R: 0, G: 0, B: 0, A: 1},
				PathEffects: []render.PathEffect{
					render.PathPatchPathEffect(render.Color{G: 1, A: 1}, render.Color{}, 0, geom.Pt{X: 20, Y: 0}),
					render.NormalPathEffect(),
				},
			},
			XY:     geom.Pt{X: 20, Y: 20},
			Width:  25,
			Height: 25,
		}
		patch.Draw(r, &core.DrawContext{})
		if got := r.Image().RGBAAt(55, 30); got.G < 180 || got.R > 90 || got.B > 90 {
			t.Fatalf("patch path effect did not draw green offset fill, got %+v", got)
		}
	})

	t.Run("path collection", func(t *testing.T) {
		r := newStartedSemanticRenderer(t, 100, 70)
		defer r.End()

		coll := &core.PathCollection{
			Collection: core.Collection{
				Coords: core.Coords(core.CoordData),
				Alpha:  1,
				PathEffects: []render.PathEffect{
					render.PathPatchPathEffect(render.Color{B: 1, A: 1}, render.Color{}, 0, geom.Pt{X: 20, Y: 0}),
					render.NormalPathEffect(),
				},
			},
			Path:      rectPath(-8, -8, 16, 16),
			Offsets:   []geom.Pt{{X: 35, Y: 35}},
			Size:      1,
			FaceColor: render.Color{R: 0, G: 0, B: 0, A: 1},
		}
		coll.Draw(r, &core.DrawContext{})
		if got := r.Image().RGBAAt(55, 35); got.B < 180 || got.R > 90 || got.G > 90 {
			t.Fatalf("path collection path effect did not draw blue offset fill, got %+v", got)
		}
	})

	t.Run("text", func(t *testing.T) {
		r := newStartedSemanticRenderer(t, 160, 80)
		defer r.End()

		text := &core.Text{
			Content:  "Go",
			Position: geom.Pt{X: 30, Y: 45},
			FontSize: 28,
			Color:    render.Color{R: 0, G: 0, B: 0, A: 1},
			ClipOn:   true,
			PathEffects: []render.PathEffect{
				render.StrokePathEffect(render.Color{R: 1, G: 0, B: 0, A: 1}, 4, geom.Pt{X: 4, Y: 0}),
				render.NormalPathEffect(),
			},
		}
		text.Draw(r, &core.DrawContext{})
		if !imageHasNonBackgroundPixel(r.Image(), semanticWhite) {
			t.Fatal("text path effects rendered a blank image")
		}
	})
}

func renderPathEffects(effects []render.PathEffect) *image.RGBA {
	r := New(80, 60, render.Color{R: 1, G: 1, B: 1, A: 1})
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 80, Y: 60}})
	r.Path(rectPath(20, 20, 25, 25), &render.Paint{
		Fill:        render.Color{R: 0, G: 0, B: 0, A: 1},
		PathEffects: effects,
	})
	_ = r.End()
	return r.Image()
}

func renderLinePathEffects(effects []render.PathEffect) *image.RGBA {
	r := New(100, 70, render.Color{R: 1, G: 1, B: 1, A: 1})
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 100, Y: 70}})
	var p geom.Path
	p.MoveTo(geom.Pt{X: 15, Y: 30})
	p.LineTo(geom.Pt{X: 85, Y: 30})
	r.Path(p, &render.Paint{
		Stroke:      render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth:   3,
		LineCap:     render.CapButt,
		PathEffects: effects,
	})
	_ = r.End()
	return r.Image()
}

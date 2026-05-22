package gobasic

import (
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestPathPaintStateSemantics(t *testing.T) {
	t.Run("alpha fill blends over background", func(t *testing.T) {
		r := newStartedSemanticRenderer(t, 80, 60)
		defer r.End()

		r.Path(rectPath(10, 10, 60, 40), &render.Paint{
			Fill:      render.Color{R: 1, G: 0, B: 0, A: 0.5},
			Antialias: render.AntialiasOff,
		})

		got := r.GetImage().RGBAAt(30, 25)
		if got.R < 240 || got.G < 105 || got.G > 155 || got.B < 105 || got.B > 155 {
			t.Fatalf("expected semi-transparent red over white at center, got %+v", got)
		}
	})

	t.Run("dashes leave gaps", func(t *testing.T) {
		r := newStartedSemanticRenderer(t, 100, 50)
		defer r.End()

		var p geom.Path
		p.MoveTo(geom.Pt{X: 10, Y: 25})
		p.LineTo(geom.Pt{X: 90, Y: 25})
		r.Path(p, &render.Paint{
			Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
			LineWidth: 6,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Dashes:    []float64{10, 10},
		})

		if !isDark(r.GetImage().RGBAAt(15, 25)) {
			t.Fatalf("expected dash segment near x=15 to be dark, got %+v", r.GetImage().RGBAAt(15, 25))
		}
		if got := r.GetImage().RGBAAt(25, 25); got != semanticWhite {
			t.Fatalf("expected dash gap near x=25 to remain white, got %+v", got)
		}
	})

	t.Run("line caps affect endpoints", func(t *testing.T) {
		butt := renderLineCap(render.CapButt)
		square := renderLineCap(render.CapSquare)

		if got := butt.RGBAAt(16, 25); got != semanticWhite {
			t.Fatalf("butt cap should not extend before start point, got %+v", got)
		}
		if !isDark(square.RGBAAt(16, 25)) {
			t.Fatalf("square cap should extend before start point, got %+v", square.RGBAAt(16, 25))
		}
	})

	t.Run("line joins affect corner shape", func(t *testing.T) {
		bevel := renderLineJoin(render.JoinBevel)
		miter := renderLineJoin(render.JoinMiter)

		bevelPixels := countNonBackgroundPixels(bevel, image.Rect(42, 4, 59, 22), semanticWhite)
		miterPixels := countNonBackgroundPixels(miter, image.Rect(42, 4, 59, 22), semanticWhite)
		if miterPixels <= bevelPixels {
			t.Fatalf("expected miter join to occupy more outer-corner pixels than bevel, got miter=%d bevel=%d", miterPixels, bevelPixels)
		}
	})
}

func TestClipRectStackIntersectsAndRestoresSemantically(t *testing.T) {
	r := newStartedSemanticRenderer(t, 100, 70)
	defer r.End()

	r.Save()
	r.ClipRect(geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 90, Y: 60}})
	r.Save()
	r.ClipRect(geom.Rect{Min: geom.Pt{X: 40, Y: 10}, Max: geom.Pt{X: 90, Y: 60}})
	r.Path(rectPath(0, 0, 100, 70), &render.Paint{
		Fill: render.Color{R: 1, G: 0, B: 0, A: 1},
	})
	r.Restore()
	r.Path(rectPath(0, 0, 100, 70), &render.Paint{
		Fill: render.Color{R: 0, G: 0, B: 1, A: 1},
	})
	r.Restore()
	r.Path(rectPath(0, 0, 8, 8), &render.Paint{
		Fill: render.Color{R: 0, G: 1, B: 0, A: 1},
	})

	img := r.GetImage()
	if got := img.RGBAAt(20, 30); got.B <= 200 || got.R >= 80 {
		t.Fatalf("outer clip should be restored after nested clip, got %+v", got)
	}
	if got := img.RGBAAt(95, 30); got != semanticWhite {
		t.Fatalf("outer clip should still reject pixels outside it, got %+v", got)
	}
	if got := img.RGBAAt(4, 4); got.G <= 200 || got.R >= 80 || got.B >= 80 {
		t.Fatalf("final restore should remove all clip rects, got %+v", got)
	}
}

func TestTextMetricsContractsAndLimitations(t *testing.T) {
	r := New(160, 80, render.Color{R: 1, G: 1, B: 1, A: 1})
	r.SetResolution(96)

	metrics := r.MeasureText("fallback font", 12, "definitely-missing-font")
	if metrics.W <= 0 || metrics.H <= 0 {
		t.Fatalf("missing font fallback should still produce basic metrics, got %+v", metrics)
	}

	path, ok := r.TextPath("Hi", geom.Pt{X: 10, Y: 40}, 12, "definitely-missing-font")
	if !ok || len(path.C) == 0 {
		t.Fatalf("missing font fallback should still produce a text path, ok=%v path=%+v", ok, path)
	}

	if _, ok := any(r).(render.TextBounder); ok {
		t.Fatal("GoBasic should not claim text ink bounds support")
	}
	if _, ok := any(r).(render.TextFontMetricer); ok {
		t.Fatal("GoBasic should not claim font-wide metrics support")
	}
}

func TestRotatedImageUsesGoBasicImageTransform(t *testing.T) {
	r := newStartedSemanticRenderer(t, 80, 60)
	defer r.End()

	if _, ok := any(r).(render.ImageTransformer); !ok {
		t.Fatal("GoBasic should expose image transforms for mixed-raster replay")
	}

	img := &core.Image2D{
		Data:     [][]float64{{0, 1}, {1, 0}},
		Colormap: "viridis",
		VMin:     0,
		VMax:     1,
		Alpha:    1,
		XMin:     10,
		XMax:     70,
		YMin:     10,
		YMax:     50,
		AngleDeg: 35,
	}
	img.Draw(r, &core.DrawContext{})

	if !imageHasNonBackgroundPixel(r.GetImage(), semanticWhite) {
		t.Fatal("rotated image transform should draw pixels")
	}
}

func TestCollectionFallbackRoutingRendersWithGoBasic(t *testing.T) {
	t.Run("interfaces remain renderer-neutral fallbacks", func(t *testing.T) {
		r := New(120, 90, render.Color{R: 1, G: 1, B: 1, A: 1})
		if _, ok := any(r).(render.MarkerDrawer); ok {
			t.Fatal("GoBasic should not expose native marker batch drawing")
		}
		if _, ok := any(r).(render.PathCollectionDrawer); ok {
			t.Fatal("GoBasic should not expose native path collection drawing")
		}
		if _, ok := any(r).(render.QuadMeshDrawer); ok {
			t.Fatal("GoBasic should not expose native quad mesh drawing")
		}
		if _, ok := any(r).(render.GouraudTriangleDrawer); ok {
			t.Fatal("GoBasic should not expose native Gouraud triangle drawing")
		}
		if hatcher, ok := any(r).(render.NativeHatcher); ok && hatcher.SupportsNativeHatch() {
			t.Fatal("GoBasic should route hatches through the renderer-neutral fallback")
		}
	})

	t.Run("marker collection fallback draws paths", func(t *testing.T) {
		r := newStartedSemanticRenderer(t, 120, 90)
		defer r.End()

		coll := &core.PathCollection{
			Collection: core.Collection{Coords: core.Coords(core.CoordData), Alpha: 1},
			Path:       rectPath(-1, -1, 2, 2),
			Offsets:    []geom.Pt{{X: 40, Y: 40}, {X: 80, Y: 40}},
			Size:       10,
			FaceColor:  render.Color{R: 0, G: 0, B: 1, A: 1},
		}
		coll.Draw(r, &core.DrawContext{})

		if !imageHasNonBackgroundPixel(r.GetImage(), semanticWhite) {
			t.Fatal("marker fallback rendered a blank image")
		}
	})

	t.Run("hatched patch collection fallback draws hatch paths", func(t *testing.T) {
		r := newStartedSemanticRenderer(t, 120, 90)
		defer r.End()

		coll := &core.PatchCollection{
			Collection: core.Collection{Coords: core.Coords(core.CoordData), Alpha: 1},
			Paths:      []geom.Path{rectPath(25, 20, 70, 50)},
			Hatch:      "/",
			HatchColor: render.Color{R: 0, G: 0, B: 0, A: 1},
			HatchWidth: 1,
		}
		coll.Draw(r, &core.DrawContext{})

		if !imageHasNonBackgroundPixel(r.GetImage(), semanticWhite) {
			t.Fatal("hatched patch fallback rendered a blank image")
		}
	})

	t.Run("flat and Gouraud quad meshes fall back to patch drawing", func(t *testing.T) {
		for _, shading := range []core.MeshShading{core.MeshShadingFlat, core.MeshShadingGouraud} {
			r := newStartedSemanticRenderer(t, 120, 90)
			mesh := &core.QuadMesh{
				PatchCollection: core.PatchCollection{
					Collection: core.Collection{Coords: core.Coords(core.CoordData), Alpha: 1},
					FaceColors: []render.Color{{R: 1, G: 0, B: 0, A: 1}},
				},
				XEdges:  []float64{20, 90},
				YEdges:  []float64{15, 70},
				Shading: shading,
				Values:  [][]float64{{0, 1}, {1, 0}},
			}
			mesh.Draw(r, &core.DrawContext{})
			if err := r.End(); err != nil {
				t.Fatalf("End failed for shading %s: %v", shading, err)
			}
			if !imageHasNonBackgroundPixel(r.GetImage(), semanticWhite) {
				t.Fatalf("%s quad mesh fallback rendered a blank image", shading)
			}
		}
	})
}

func newStartedSemanticRenderer(t *testing.T, w, h int) *Renderer {
	t.Helper()

	r := New(w, h, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: float64(w), Y: float64(h)}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	return r
}

func renderLineCap(cap render.LineCap) *image.RGBA {
	r := New(80, 50, render.Color{R: 1, G: 1, B: 1, A: 1})
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 80, Y: 50}})
	var p geom.Path
	p.MoveTo(geom.Pt{X: 20, Y: 25})
	p.LineTo(geom.Pt{X: 50, Y: 25})
	r.Path(p, &render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 10,
		LineJoin:  render.JoinMiter,
		LineCap:   cap,
	})
	_ = r.End()
	return r.GetImage()
}

func renderLineJoin(join render.LineJoin) *image.RGBA {
	r := New(100, 70, render.Color{R: 1, G: 1, B: 1, A: 1})
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 100, Y: 70}})
	var p geom.Path
	p.MoveTo(geom.Pt{X: 20, Y: 55})
	p.LineTo(geom.Pt{X: 50, Y: 15})
	p.LineTo(geom.Pt{X: 80, Y: 55})
	r.Path(p, &render.Paint{
		Stroke:     render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth:  18,
		LineJoin:   join,
		LineCap:    render.CapButt,
		MiterLimit: 10,
	})
	_ = r.End()
	return r.GetImage()
}

func rectPath(x, y, w, h float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: x, Y: y})
	p.LineTo(geom.Pt{X: x + w, Y: y})
	p.LineTo(geom.Pt{X: x + w, Y: y + h})
	p.LineTo(geom.Pt{X: x, Y: y + h})
	p.Close()
	return p
}

func imageHasNonBackgroundPixel(img image.Image, bg color.RGBA) bool {
	if img == nil {
		return false
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if color.RGBAModel.Convert(img.At(x, y)).(color.RGBA) != bg {
				return true
			}
		}
	}
	return false
}

func countNonBackgroundPixels(img image.Image, rect image.Rectangle, bg color.RGBA) int {
	if img == nil {
		return 0
	}
	bounds := rect.Intersect(img.Bounds())
	count := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if color.RGBAModel.Convert(img.At(x, y)).(color.RGBA) != bg {
				count++
			}
		}
	}
	return count
}

func isDark(c color.RGBA) bool {
	return c.R < 80 && c.G < 80 && c.B < 80
}

var semanticWhite = color.RGBA{R: 255, G: 255, B: 255, A: 255}

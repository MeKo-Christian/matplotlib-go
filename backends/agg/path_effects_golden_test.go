package agg

import (
	"flag"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
)

var updatePathEffectsGoldens = flag.Bool("update-path-effects-golden", false, "update AGG path-effect golden fixtures")

type pathEffectGoldenFixture struct {
	name   string
	render func(*testing.T) image.Image
}

func TestPathEffectGoldens(t *testing.T) {
	for _, fixture := range pathEffectGoldenFixtures() {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			got := fixture.render(t)
			if *updatePathEffectsGoldens {
				goldenPath := filepath.Join("testdata", pathEffectGoldenWriteDir(), fixture.name+".png")
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
					t.Fatalf("create golden dir: %v", err)
				}
				if err := imagecmp.SavePNG(got, goldenPath); err != nil {
					t.Fatalf("update golden %s: %v", goldenPath, err)
				}
				t.Skip("updated AGG path-effect golden")
			}

			goldenPath := pathEffectGoldenReadPath(fixture.name)
			want, err := imagecmp.LoadPNG(goldenPath)
			if err != nil {
				t.Fatalf("load golden %s: %v\n(rerun with -update-path-effects-golden)", goldenPath, err)
			}
			diff, err := imagecmp.ComparePNG(got, want, 1)
			if err != nil {
				t.Fatalf("compare golden %s: %v", goldenPath, err)
			}
			if !diff.Identical {
				artifacts := filepath.Join("testdata", "_artifacts", "path_effects_golden")
				if err := os.MkdirAll(artifacts, 0o755); err != nil {
					t.Fatalf("create artifact dir: %v", err)
				}
				if err := imagecmp.SavePNG(got, filepath.Join(artifacts, fixture.name+"_got.png")); err != nil {
					t.Fatalf("save got artifact: %v", err)
				}
				if err := imagecmp.SaveDiffImage(got, want, 1, filepath.Join(artifacts, fixture.name+"_diff.png")); err != nil {
					t.Fatalf("save diff artifact: %v", err)
				}
				t.Fatalf("golden mismatch: MaxDiff=%d MeanAbs=%.2f PSNR=%.2f", diff.MaxDiff, diff.MeanAbs, diff.PSNR)
			}
		})
	}
}

func pathEffectGoldenFixtures() []pathEffectGoldenFixture {
	return []pathEffectGoldenFixture{
		{name: "text_drop_shadow", render: renderTextDropShadowFixture},
		{name: "line_halo", render: renderLineHaloFixture},
		{name: "scatter_marker_shadow", render: renderScatterMarkerShadowFixture},
		{name: "polygon_effect_stack", render: renderPolygonEffectStackFixture},
	}
}

func renderTextDropShadowFixture(t *testing.T) image.Image {
	t.Helper()
	fig := core.NewFigure(
		320,
		220,
		style.WithBackground(0.96, 0.97, 0.98, 1),
		style.WithFont("DejaVu Sans", 12),
	)
	fig.Text(0.5, 0.52, "Path FX", core.TextOptions{
		FontSize: 34,
		Color:    render.Color{R: 0.08, G: 0.18, B: 0.34, A: 1},
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignMiddle,
		PathEffects: []render.PathEffect{
			render.SimplePatchShadowPathEffect(geom.Pt{X: 5, Y: 6}, render.Color{R: 0.02, G: 0.03, B: 0.04, A: 0.9}, 0.55, 0.25),
			render.NormalPathEffect(),
		},
	})
	return renderPathEffectFigure(t, fig)
}

func renderLineHaloFixture(t *testing.T) image.Image {
	t.Helper()
	return renderPathEffectCanvas(t, drawLineHalo)
}

func renderScatterMarkerShadowFixture(t *testing.T) image.Image {
	t.Helper()
	return renderPathEffectCanvas(t, drawScatterMarkerShadow)
}

func renderPolygonEffectStackFixture(t *testing.T) image.Image {
	t.Helper()
	return renderPathEffectCanvas(t, drawPolygonEffectStack)
}

func renderPathEffectFigure(t *testing.T, fig *core.Figure) image.Image {
	t.Helper()
	r := mustNew(t, int(fig.SizePx.X), int(fig.SizePx.Y))
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderPathEffectCanvas(t *testing.T, draw func(*Renderer)) image.Image {
	t.Helper()
	r := mustNew(t, 320, 220)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 320, Y: 220}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Path(rectPath(0, 0, 320, 220), &render.Paint{
		Fill: render.Color{R: 0.96, G: 0.97, B: 0.98, A: 1},
	})
	draw(r)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	return r.GetImage()
}

func drawLineHalo(r *Renderer) {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 32, Y: 155})
	for i := 1; i <= 48; i++ {
		t := float64(i) / 48
		x := 32 + 256*t
		y := 112 + 36*math.Sin(t*math.Pi*2.35)
		p.LineTo(geom.Pt{X: x, Y: y})
	}
	r.Path(p, &render.Paint{
		Stroke:    render.Color{R: 0.08, G: 0.34, B: 0.66, A: 1},
		LineWidth: 3,
		LineCap:   render.CapRound,
		LineJoin:  render.JoinRound,
		PathEffects: render.WithStrokePathEffects(
			render.Color{R: 1, G: 1, B: 1, A: 0.96},
			11,
			geom.Pt{},
		),
	})
}

func drawScatterMarkerShadow(r *Renderer) {
	colors := []render.Color{
		{R: 0.89, G: 0.22, B: 0.24, A: 1},
		{R: 0.10, G: 0.55, B: 0.38, A: 1},
		{R: 0.13, G: 0.35, B: 0.72, A: 1},
		{R: 0.94, G: 0.64, B: 0.18, A: 1},
	}
	for i, c := range []geom.Pt{
		{X: 78, Y: 82},
		{X: 138, Y: 136},
		{X: 202, Y: 78},
		{X: 250, Y: 144},
	} {
		r.Path(circlePath(c, 22+float64(i%2)*6, 36), &render.Paint{
			Fill:      colors[i],
			Stroke:    render.Color{R: 0.04, G: 0.06, B: 0.08, A: 1},
			LineWidth: 1.4,
			PathEffects: []render.PathEffect{
				render.SimplePatchShadowPathEffect(geom.Pt{X: 5, Y: 7}, render.Color{R: 0.02, G: 0.03, B: 0.04, A: 0.7}, 0.5, 0.3),
				render.NormalPathEffect(),
			},
		})
	}
}

func drawPolygonEffectStack(r *Renderer) {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 56, Y: 156})
	p.LineTo(geom.Pt{X: 112, Y: 48})
	p.LineTo(geom.Pt{X: 206, Y: 58})
	p.LineTo(geom.Pt{X: 270, Y: 136})
	p.LineTo(geom.Pt{X: 188, Y: 178})
	p.Close()

	r.Path(p, &render.Paint{
		Fill:      render.Color{R: 0.94, G: 0.77, B: 0.28, A: 1},
		Stroke:    render.Color{R: 0.07, G: 0.20, B: 0.38, A: 1},
		LineWidth: 2.2,
		LineJoin:  render.JoinRound,
		PathEffects: []render.PathEffect{
			render.SimplePatchShadowPathEffect(geom.Pt{X: 7, Y: 8}, render.Color{R: 0.02, G: 0.03, B: 0.04, A: 0.7}, 0.45, 0.35),
			render.PathPatchPathEffect(render.Color{R: 0.95, G: 0.92, B: 0.82, A: 0.75}, render.Color{R: 0.83, G: 0.20, B: 0.19, A: 1}, 6, geom.Pt{}),
			render.NormalPathEffect(),
		},
	})
}

func circlePath(center geom.Pt, radius float64, steps int) geom.Path {
	var p geom.Path
	for i := 0; i <= steps; i++ {
		a := 2 * math.Pi * float64(i) / float64(steps)
		pt := geom.Pt{X: center.X + radius*math.Cos(a), Y: center.Y + radius*math.Sin(a)}
		if i == 0 {
			p.MoveTo(pt)
			continue
		}
		p.LineTo(pt)
	}
	p.Close()
	return p
}

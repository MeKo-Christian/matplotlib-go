package agg

import (
	"flag"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
)

var updatePatternGradientGoldens = flag.Bool("update-pattern-gradient-golden", false, "update AGG pattern/gradient golden fixtures")

type patternGradientGoldenFixture struct {
	name string
	draw func(*Renderer)
}

func TestPatternGradientGoldens(t *testing.T) {
	for _, fixture := range patternGradientGoldenFixtures() {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			got := renderPatternGradientFixture(t, fixture.draw)
			goldenPath := filepath.Join("testdata", "pattern_gradient_golden", fixture.name+".png")
			if *updatePatternGradientGoldens {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
					t.Fatalf("create golden dir: %v", err)
				}
				if err := imagecmp.SavePNG(got, goldenPath); err != nil {
					t.Fatalf("update golden %s: %v", goldenPath, err)
				}
				t.Skip("updated AGG pattern/gradient golden")
			}

			want, err := imagecmp.LoadPNG(goldenPath)
			if err != nil {
				t.Fatalf("load golden %s: %v\n(rerun with -update-pattern-gradient-golden)", goldenPath, err)
			}
			diff, err := imagecmp.ComparePNG(got, want, 1)
			if err != nil {
				t.Fatalf("compare golden %s: %v", goldenPath, err)
			}
			if !diff.Identical {
				artifacts := filepath.Join("testdata", "_artifacts", "pattern_gradient_golden")
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

func patternGradientGoldenFixtures() []patternGradientGoldenFixture {
	return []patternGradientGoldenFixture{
		{name: "gradient_fill_bar", draw: drawGradientFillBars},
		{name: "radial_gradient_pie_wedge", draw: drawRadialGradientPieWedge},
		{name: "pattern_fill_polygon", draw: drawPatternFillPolygon},
		{name: "gradient_streamline_plot", draw: drawGradientStreamlines},
	}
}

func renderPatternGradientFixture(t *testing.T, draw func(*Renderer)) image.Image {
	t.Helper()
	r := mustNew(t, 320, 220)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 320, Y: 220}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	draw(r)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	return r.Image()
}

func drawGradientFillBars(r *Renderer) {
	for i := 0; i < 5; i++ {
		x0 := 34 + float64(i)*52
		y0 := 170 - float64(i)*24
		rect := rectPath(x0, y0, x0+34, 188)
		r.Path(rect, &render.Paint{
			FillGradient: render.GradientFill{
				Kind:  render.LinearGradient,
				Start: geom.Pt{X: x0, Y: y0},
				End:   geom.Pt{X: x0 + 34, Y: 188},
				Stops: []render.GradientStop{
					{Offset: 0, Color: render.Color{R: 0.10, G: 0.46, B: 0.82, A: 1}},
					{Offset: 1, Color: render.Color{R: 0.95, G: 0.68, B: 0.18, A: 1}},
				},
			},
			Stroke:    render.Color{R: 0.10, G: 0.12, B: 0.16, A: 1},
			LineWidth: 1.2,
		})
	}
}

func drawRadialGradientPieWedge(r *Renderer) {
	wedge := wedgePath(160, 116, 74, -32*math.Pi/180, 246*math.Pi/180, 64)
	r.Path(wedge, &render.Paint{
		FillGradient: render.GradientFill{
			Kind:   render.RadialGradient,
			Center: geom.Pt{X: 160, Y: 116},
			Radius: 74,
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1.00, G: 0.96, B: 0.58, A: 1}},
				{Offset: 0.55, Color: render.Color{R: 0.97, G: 0.42, B: 0.24, A: 1}},
				{Offset: 1, Color: render.Color{R: 0.38, G: 0.11, B: 0.44, A: 1}},
			},
		},
		Stroke:    render.Color{R: 0.16, G: 0.08, B: 0.22, A: 1},
		LineWidth: 1.5,
	})
}

func drawPatternFillPolygon(r *Renderer) {
	var tile geom.Path
	tile.MoveTo(geom.Pt{X: 0, Y: 16})
	tile.LineTo(geom.Pt{X: 16, Y: 0})
	tile.MoveTo(geom.Pt{X: -4, Y: 4})
	tile.LineTo(geom.Pt{X: 4, Y: -4})
	tile.MoveTo(geom.Pt{X: 12, Y: 20})
	tile.LineTo(geom.Pt{X: 20, Y: 12})

	polygon := geom.Path{}
	polygon.MoveTo(geom.Pt{X: 56, Y: 164})
	polygon.LineTo(geom.Pt{X: 112, Y: 42})
	polygon.LineTo(geom.Pt{X: 224, Y: 56})
	polygon.LineTo(geom.Pt{X: 272, Y: 152})
	polygon.LineTo(geom.Pt{X: 182, Y: 188})
	polygon.Close()
	r.Path(polygon, &render.Paint{
		FillPattern: render.PatternFill{
			ID:         "diagonal-stripe",
			Cell:       geom.Rect{Max: geom.Pt{X: 16, Y: 16}},
			Path:       tile,
			Foreground: render.Color{R: 0.06, G: 0.26, B: 0.58, A: 1},
			Background: render.Color{R: 0.83, G: 0.91, B: 0.96, A: 1},
			LineWidth:  2,
		},
		Stroke:    render.Color{R: 0.04, G: 0.10, B: 0.18, A: 1},
		LineWidth: 1.6,
	})
}

func drawGradientStreamlines(r *Renderer) {
	for i := 0; i < 9; i++ {
		y := 42 + float64(i)*16
		path := ribbonPath(y, float64(i)*0.45)
		r.Path(path, &render.Paint{
			FillGradient: render.GradientFill{
				Kind:  render.LinearGradient,
				Start: geom.Pt{X: 30, Y: y},
				End:   geom.Pt{X: 290, Y: y},
				Stops: []render.GradientStop{
					{Offset: 0, Color: render.Color{R: 0.08, G: 0.32, B: 0.64, A: 0.92}},
					{Offset: 0.5, Color: render.Color{R: 0.10, G: 0.62, B: 0.46, A: 0.92}},
					{Offset: 1, Color: render.Color{R: 0.92, G: 0.70, B: 0.18, A: 0.92}},
				},
			},
		})
	}
}

func rectPath(x0, y0, x1, y1 float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: x0, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y1})
	p.LineTo(geom.Pt{X: x0, Y: y1})
	p.Close()
	return p
}

func wedgePath(cx, cy, radius, start, end float64, steps int) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: cx, Y: cy})
	for i := 0; i <= steps; i++ {
		t := start + (end-start)*float64(i)/float64(steps)
		pt := geom.Pt{X: cx + radius*math.Cos(t), Y: cy + radius*math.Sin(t)}
		if i == 0 {
			p.LineTo(pt)
		} else {
			p.LineTo(pt)
		}
	}
	p.Close()
	return p
}

func ribbonPath(y, phase float64) geom.Path {
	const (
		x0    = 28.0
		x1    = 292.0
		steps = 36
	)
	top := make([]geom.Pt, 0, steps+1)
	bottom := make([]geom.Pt, 0, steps+1)
	for i := 0; i <= steps; i++ {
		t := float64(i) / steps
		x := x0 + (x1-x0)*t
		center := y + 5*math.Sin(t*math.Pi*2+phase)
		half := 3.8 + 1.2*math.Cos(t*math.Pi*3+phase)
		top = append(top, geom.Pt{X: x, Y: center - half})
		bottom = append(bottom, geom.Pt{X: x, Y: center + half})
	}
	var p geom.Path
	p.MoveTo(top[0])
	for i := 1; i < len(top); i++ {
		p.LineTo(top[i])
	}
	for i := len(bottom) - 1; i >= 0; i-- {
		p.LineTo(bottom[i])
	}
	p.Close()
	return p
}

package pgf

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func renderPGFDocument(t *testing.T, draw func(*Renderer)) []byte {
	t.Helper()
	r, err := New(120, 80, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 120, Y: 80}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	draw(r)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	return r.document
}

func testRectPath() geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 40})
	p.LineTo(geom.Pt{X: 10, Y: 40})
	p.Close()
	return p
}

func TestDocumentationRecordsTextMetricsPolicy(t *testing.T) {
	doc, err := os.ReadFile("doc.go")
	if err != nil {
		t.Fatalf("ReadFile doc.go: %v", err)
	}
	text := strings.ReplaceAll(string(doc), "\n// ", " ")
	for _, want := range []string{
		"deterministic approximations",
		"exact TeX/font metrics are delegated to LaTeX",
		"does not implement TeX metric extraction",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("PGF package docs should document %q limitation:\n%s", want, doc)
		}
	}
}

func TestPathEffectFilterFallbackPolicyIsDocumented(t *testing.T) {
	r, err := New(120, 80, render.Color{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := any(r).(render.FilterRenderer); ok {
		t.Fatal("PGF should not expose offscreen filter support")
	}
	if _, ok := any(r).(render.PathEffectFilterDrawer); ok {
		t.Fatal("PGF should not expose native filtered path-effect support")
	}
	doc, err := os.ReadFile("doc.go")
	if err != nil {
		t.Fatalf("ReadFile doc.go: %v", err)
	}
	text := strings.Join(strings.Fields(strings.ReplaceAll(string(doc), "\n// ", " ")), " ")
	for _, want := range []string{
		"Filter path effects",
		"renderer-neutral fallback",
		"does not expose native filtered path-effect support",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("PGF package docs should document %q fallback:\n%s", want, doc)
		}
	}
}

func TestPathAlphaEmitsPGFOpacityCommands(t *testing.T) {
	doc := renderPGFDocument(t, func(r *Renderer) {
		r.Path(testRectPath(), &render.Paint{
			Fill:      render.Color{R: 1, A: 0.25},
			Stroke:    render.Color{B: 1, A: 0.5},
			LineWidth: 2,
		})
		r.Path(testRectPath(), &render.Paint{
			Stroke:    render.Color{A: 1},
			LineWidth: 1,
		})
	})
	for _, want := range [][]byte{
		[]byte(`\pgfsetfillopacity{0.25}`),
		[]byte(`\pgfsetstrokeopacity{0.5}`),
		[]byte(`\pgfsetfillopacity{1}`),
		[]byte(`\pgfsetstrokeopacity{1}`),
	} {
		if !bytes.Contains(doc, want) {
			t.Fatalf("missing %q in\n%s", want, doc)
		}
	}
}

func TestPathWithHatchEmitsClippedHatchLines(t *testing.T) {
	r, err := New(120, 80, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	hatcher, ok := any(r).(render.NativeHatcher)
	if !ok {
		t.Fatal("PGF renderer should implement render.NativeHatcher")
	}
	if !hatcher.SupportsNativeHatch() {
		t.Fatal("SupportsNativeHatch returned false")
	}

	doc := renderPGFDocument(t, func(r *Renderer) {
		r.Path(testRectPath(), &render.Paint{
			Fill:           render.Color{G: 1, A: 0.3},
			Hatch:          "x",
			HatchColor:     render.Color{R: 0.2, G: 0.3, B: 0.4, A: 0.75},
			HatchLineWidth: 1.5,
			HatchSpacing:   12,
		})
	})
	for _, want := range [][]byte{
		[]byte(`\pgfusepath{clip}`),
		[]byte(`\pgfsetlinewidth{1.5pt}`),
		[]byte(`\pgfpathlineto`),
		[]byte(`\pgfusepath{stroke}`),
	} {
		if !bytes.Contains(doc, want) {
			t.Fatalf("missing %q in\n%s", want, doc)
		}
	}
}

func TestPathWithShapeHatchEmitsClippedShapeGeometry(t *testing.T) {
	doc := renderPGFDocument(t, func(r *Renderer) {
		r.Path(testRectPath(), &render.Paint{
			Hatch:          "oO.*",
			HatchColor:     render.Color{A: 1},
			HatchLineWidth: 1,
			HatchSpacing:   12,
		})
	})
	for _, want := range [][]byte{
		[]byte(`\pgfusepath{clip}`),
		[]byte(`\pgfpathcurveto`),
		[]byte(`\pgfusepath{fill}`),
		[]byte(`\pgfusepath{stroke}`),
	} {
		if !bytes.Contains(doc, want) {
			t.Fatalf("missing shape hatch fragment %q in\n%s", want, doc)
		}
	}
}

func TestImageEmitsSelfContainedPixelRectangles(t *testing.T) {
	doc := renderPGFDocument(t, func(r *Renderer) {
		img := image.NewRGBA(image.Rect(0, 0, 2, 1))
		img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
		img.SetRGBA(1, 0, color.RGBA{G: 128, A: 128})
		data := render.NewImageData(img)
		data.SetAlpha(0.5)

		r.DrawImage(data, geom.Rect{
			Min: geom.Pt{X: 10, Y: 20},
			Max: geom.Pt{X: 14, Y: 22},
		})
	})
	for _, want := range [][]byte{
		[]byte(`\pgftransformcm{2}{0}{0}{-2}{\pgfpoint{10pt}{22pt}}`),
		[]byte(`\pgfpathrectangle{\pgfpoint{0pt}{0pt}}{\pgfpoint{1pt}{1pt}}`),
		[]byte(`\pgfpathrectangle{\pgfpoint{1pt}{0pt}}{\pgfpoint{1pt}{1pt}}`),
		[]byte(`\pgfsetfillopacity{0.5}`),
		[]byte(`\pgfsetfillopacity{0.25098}`),
	} {
		if !bytes.Contains(doc, want) {
			t.Fatalf("missing %q in\n%s", want, doc)
		}
	}
}

func TestImageTransformedEmitsAffinePixelScope(t *testing.T) {
	doc := renderPGFDocument(t, func(r *Renderer) {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		img.SetRGBA(0, 0, color.RGBA{B: 255, A: 255})
		r.ImageTransformed(render.NewImageData(img), geom.Rect{}, geom.Affine{
			A: 3, B: 0.5, C: -0.25, D: 4, E: 7, F: 11,
		})
	})
	if !bytes.Contains(doc, []byte(`\pgftransformcm{3}{0.5}{-0.25}{4}{\pgfpoint{7pt}{11pt}}`)) {
		t.Fatalf("missing transformed image matrix in\n%s", doc)
	}
	if !bytes.Contains(doc, []byte(`\pgfpathrectangle{\pgfpoint{0pt}{0pt}}{\pgfpoint{1pt}{1pt}}`)) {
		t.Fatalf("missing transformed image pixel in\n%s", doc)
	}
}

func TestRasterizedArtistEmbedsPixelsAndKeepsVectorText(t *testing.T) {
	fig := core.NewFigure(120, 80)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.2, Y: 0.2}, Max: geom.Pt{X: 0.8, Y: 0.75}})
	line := ax.Plot([]float64{0, 0.5, 1}, []float64{0, 1, 0})
	line.SetRasterized(true)
	ax.SetTitle("Vector title")

	r, err := New(120, 80, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	core.DrawFigure(fig, r)
	doc := r.document

	if !bytes.Contains(doc, []byte(`\pgfpathrectangle{\pgfpoint{`)) {
		t.Fatalf("rasterized artist did not emit PGF pixel rectangles:\n%s", doc)
	}
	if !bytes.Contains(doc, []byte(`Vector title`)) {
		t.Fatalf("surrounding title text was not preserved as PGF text:\n%s", doc)
	}
}

func TestDrawMarkersEmitsReusableMacro(t *testing.T) {
	r, err := New(120, 80, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := any(r).(render.MarkerDrawer); !ok {
		t.Fatal("PGF renderer should implement render.MarkerDrawer")
	}

	doc := renderPGFDocument(t, func(r *Renderer) {
		ok := r.DrawMarkers(render.MarkerBatch{
			Marker: testRectPath(),
			Items: []render.MarkerItem{
				{
					Offset: geom.Pt{X: 10, Y: 20},
					Paint:  render.Paint{Fill: render.Color{R: 1, A: 1}},
				},
				{
					Offset:    geom.Pt{X: 30, Y: 40},
					Transform: geom.Affine{A: 2, D: 2},
					Paint:     render.Paint{Fill: render.Color{R: 1, A: 1}},
				},
			},
		})
		if !ok {
			t.Fatal("DrawMarkers returned false")
		}
	})
	if got := strings.Count(string(doc), `\expandafter\def\csname mplgpgfM1\endcsname`); got != 1 {
		t.Fatalf("expected one marker macro definition, got %d in\n%s", got, doc)
	}
	for _, want := range [][]byte{
		[]byte(`\pgftransformcm{1}{0}{0}{1}{\pgfpoint{10pt}{20pt}}`),
		[]byte(`\pgftransformcm{2}{0}{0}{2}{\pgfpoint{30pt}{40pt}}`),
		[]byte(`\csname mplgpgfM1\endcsname`),
	} {
		if !bytes.Contains(doc, want) {
			t.Fatalf("missing %q in\n%s", want, doc)
		}
	}
}

func TestDrawPathCollectionEmitsReusableMacro(t *testing.T) {
	r, err := New(120, 80, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := any(r).(render.PathCollectionDrawer); !ok {
		t.Fatal("PGF renderer should implement render.PathCollectionDrawer")
	}

	doc := renderPGFDocument(t, func(r *Renderer) {
		ok := r.DrawPathCollection(render.PathCollectionBatch{
			Items: []render.PathCollectionItem{
				{
					Path:  testRectPath(),
					Paint: render.Paint{Stroke: render.Color{B: 1, A: 1}, LineWidth: 2},
				},
				{
					Path:  testRectPath(),
					Paint: render.Paint{Stroke: render.Color{B: 1, A: 1}, LineWidth: 2},
				},
			},
		})
		if !ok {
			t.Fatal("DrawPathCollection returned false")
		}
	})
	if got := strings.Count(string(doc), `\expandafter\def\csname mplgpgfP1\endcsname`); got != 1 {
		t.Fatalf("expected one path collection macro definition, got %d in\n%s", got, doc)
	}
	if got := strings.Count(string(doc), `\csname mplgpgfP1\endcsname`); got != 3 {
		t.Fatalf("expected one definition plus two path collection uses, got %d in\n%s", got, doc)
	}
}

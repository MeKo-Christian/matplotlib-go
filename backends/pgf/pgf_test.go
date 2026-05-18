package pgf

import (
	"bytes"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
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

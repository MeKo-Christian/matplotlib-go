package svg

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type sizeOnlyImage struct {
	w int
	h int
}

func (i sizeOnlyImage) Size() (w, h int)      { return i.w, i.h }
func (i sizeOnlyImage) Interpolation() string { return "" }

func mustNewRenderer(t *testing.T) *Renderer {
	t.Helper()
	r, err := New(180, 120, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	return r
}

func renderSVGDocument(t *testing.T, draw func(*Renderer)) string {
	t.Helper()

	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	draw(r)

	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	return r.renderSVG()
}

func circleMarkerPath() geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: -1, Y: 0})
	p.LineTo(geom.Pt{X: 1, Y: 0})
	p.LineTo(geom.Pt{X: 0, Y: 1})
	p.Close()
	return p
}

func writeSVGFakeCommand(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake command: %v", err)
	}
	return path
}

func writeSVGTestPNG(t *testing.T, path string, c color.RGBA) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode fixture PNG: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write fixture PNG: %v", err)
	}
}

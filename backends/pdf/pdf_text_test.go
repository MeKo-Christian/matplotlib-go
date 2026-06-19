package pdf

import (
	"bytes"
	"image/color"
	"math"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestRendererImplementsTextAsPathInterfaces(t *testing.T) {
	r := newTestRenderer(t)
	if _, ok := any(r).(render.TextPather); !ok {
		t.Fatal("PDF renderer should implement render.TextPather")
	}
	if _, ok := any(r).(render.FontTextDrawer); !ok {
		t.Fatal("PDF renderer should implement render.FontTextDrawer")
	}
	if _, ok := any(r).(render.FontRotatedTextDrawer); !ok {
		t.Fatal("PDF renderer should implement render.FontRotatedTextDrawer")
	}
}

func TestRendererTextPathUsesSharedFontOutlines(t *testing.T) {
	r := newTestRenderer(t)
	path, ok := r.TextPath("Ag", geom.Pt{X: 10, Y: 30}, 14, "DejaVu Sans")
	if !ok {
		t.Fatal("TextPath returned !ok")
	}
	if !path.Validate() {
		t.Fatalf("TextPath returned invalid path: commands=%d vertices=%d", len(path.C), len(path.V))
	}
	if len(path.C) == 0 {
		t.Fatal("TextPath returned an empty outline")
	}
}

func TestDrawTextWithFontEmitsFilledGlyphPath(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})

	r.DrawTextWithFont("A", geom.Pt{X: 20, Y: 40}, 16, render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}, "DejaVu Sans")

	raw := r.content.String()
	if !strings.Contains(raw, "0.1 0.2 0.3 rg") {
		t.Fatalf("expected text fill color in content stream, got %q", raw)
	}
	if !strings.Contains(raw, " m\n") || !strings.Contains(raw, "f\n") {
		t.Fatalf("expected glyph outline path filled in content stream, got %q", raw)
	}
}

func TestDrawTextWithEmbeddedFontEmitsType0FontResource(t *testing.T) {
	r := newTestRenderer(t)
	r.SetPDFOptions(render.ResolvePDFOptions(render.WithPDFFontPolicy(render.PDFFontPolicyEmbed)))
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})

	r.DrawTextWithFont("AB", geom.Pt{X: 20, Y: 40}, 16, render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}, "DejaVu Sans")

	raw := r.content.String()
	for _, want := range []string{"BT\n", "/F1 16 Tf", "<00010002> Tj", "ET\n"} {
		if !strings.Contains(raw, want) {
			t.Fatalf("embedded text content missing %q:\n%s", want, raw)
		}
	}
	if strings.Contains(raw, " m\n") {
		t.Fatalf("embedded text should not emit glyph outline paths:\n%s", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if !bytes.Contains(data, []byte("/Font << /F1")) {
		t.Fatalf("page resources should reference embedded font F1:\n%s", data)
	}
	for _, want := range []string{
		"/Subtype /Type0",
		"/Encoding /Identity-H",
		"/DescendantFonts [",
		"/Subtype /CIDFontType2",
		"/CIDToGIDMap",
		"/FontFile2",
		"/ToUnicode",
	} {
		if !bytes.Contains(data, []byte(want)) {
			t.Fatalf("PDF document missing embedded font marker %q:\n%s", want, data)
		}
	}
}

func TestDrawTeXEmbedsCachedPNGImage(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell commands are POSIX-only")
	}
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture.png")
	writePDFTestPNG(t, fixture, color.RGBA{A: 255})
	latex := writePDFFakeCommand(t, dir, "latex", `#!/bin/sh
touch file.dvi
`)
	dvipng := writePDFFakeCommand(t, dir, "dvipng", `#!/bin/sh
out=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    shift
    out="$1"
  fi
  shift
done
cp "$FAKE_TEX_PNG" "$out"
`)
	t.Setenv("FAKE_TEX_PNG", fixture)

	r := newTestRenderer(t)
	r.texManager = tex.NewManager(tex.ManagerConfig{
		CacheDir:      filepath.Join(dir, "cache"),
		LaTeXCommand:  latex,
		DVIPNGCommand: dvipng,
	})
	if _, ok := any(r).(render.TeXMetricer); !ok {
		t.Fatal("PDF renderer should implement render.TeXMetricer")
	}
	if _, ok := any(r).(render.TeXDrawer); !ok {
		t.Fatal("PDF renderer should implement render.TeXDrawer")
	}
	if _, ok := any(r).(render.RotatedTeXDrawer); !ok {
		t.Fatal("PDF renderer should implement render.RotatedTeXDrawer")
	}

	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	metrics, ok := r.MeasureTeX(`signal $\alpha$`, 12, "DejaVu Sans")
	if !ok || metrics.W != 2 || metrics.H != 2 {
		t.Fatalf("MeasureTeX = %+v, %v; want 2x2 metrics and ok", metrics, ok)
	}
	if !r.DrawTeX(`signal $\alpha$`, geom.Pt{X: 8, Y: 10}, 12, render.Color{R: 1, A: 1}, "DejaVu Sans") {
		t.Fatal("DrawTeX returned false")
	}
	if !r.DrawTeXRotated(`x`, geom.Pt{X: 20, Y: 30}, 12, math.Pi/2, render.Color{B: 1, A: 1}, "DejaVu Sans") {
		t.Fatal("DrawTeXRotated returned false")
	}
	raw := r.content.String()
	if !strings.Contains(raw, "/Im1 Do") || !strings.Contains(raw, "/Im2 Do") || !strings.Contains(raw, "0 2 -2 0") {
		t.Fatalf("expected normal and rotated TeX image invocations, got %q", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if got := pdfDocumentObjectCountContaining(doc, "/Subtype /Image"); got < 2 {
		t.Fatalf("expected TeX image XObjects and soft masks, got %d image objects; objects: %#v", got, doc.Objects)
	}
}

package pdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/internal/pdfcompare"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
)

func newTestRenderer(t *testing.T) *Renderer {
	t.Helper()
	r, err := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return r
}

func TestShortFloat(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{1.5, "1.5"},
		{1.5000001, "1.5"},
		{math.Copysign(0, -1), "0"},
		{-1.25, "-1.25"},
		{123.456, "123.456"},
	}
	for _, c := range cases {
		got := shortFloat(c.in)
		if got != c.want {
			t.Errorf("shortFloat(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRendererProducesPDFHeaderAndEOF(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-1.7\n")) {
		t.Errorf("missing PDF-1.7 header; got prefix %q", data[:min(len(data), 16)])
	}
	if !bytes.HasSuffix(data, []byte("%%EOF\n")) {
		tail := data
		if len(tail) > 64 {
			tail = tail[len(tail)-64:]
		}
		t.Errorf("missing %%%%EOF trailer; got tail %q", tail)
	}
}

func TestRendererBeginTwiceFails(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("first Begin: %v", err)
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err == nil {
		t.Errorf("second Begin should fail")
	}
}

func TestRendererEndBeforeBeginFails(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.End(); err == nil {
		t.Errorf("End before Begin should fail")
	}
}

func TestRendererSaveRestoreEmitsBracketedQ(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.Save()
	r.Save()
	r.Restore()
	r.Restore()
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	// The content stream is flate-compressed, but the buffer is still around
	// while End ran. Re-build a probe renderer to inspect the raw content.
	probe := newTestRenderer(t)
	_ = probe.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	probe.Save()
	probe.Save()
	probe.Restore()
	probe.Restore()
	raw := probe.content.String()
	if strings.Count(raw, "q\n") != 2 {
		t.Errorf("expected 2 q lines, got %d in %q", strings.Count(raw, "q\n"), raw)
	}
	if strings.Count(raw, "Q\n") != 2 {
		t.Errorf("expected 2 Q lines, got %d in %q", strings.Count(raw, "Q\n"), raw)
	}
}

func TestRendererPathFillsAndStrokes(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 50})
	p.Close()
	r.Path(p, &render.Paint{
		Fill:      render.Color{R: 1, G: 0, B: 0, A: 1},
		Stroke:    render.Color{R: 0, G: 0, B: 1, A: 1},
		LineWidth: 1,
	})
	raw := r.content.String()
	if !strings.Contains(raw, "10 10 m") {
		t.Errorf("missing MoveTo in %q", raw)
	}
	if !strings.Contains(raw, "h") {
		t.Errorf("missing close-path in %q", raw)
	}
	if !strings.Contains(raw, "B\n") {
		t.Errorf("expected fill+stroke operator B in %q", raw)
	}
	if !strings.Contains(raw, "1 0 0 rg") {
		t.Errorf("expected red fill color in %q", raw)
	}
	if !strings.Contains(raw, "0 0 1 RG") {
		t.Errorf("expected blue stroke color in %q", raw)
	}
}

func TestRendererPathAlphaEmitsExtGState(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.Path(pdfTestRectPath(10, 10, 50, 40), &render.Paint{
		Fill:      render.Color{R: 1, A: 0.25},
		Stroke:    render.Color{B: 1, A: 0.5},
		LineWidth: 2,
	})
	raw := r.content.String()
	if !strings.Contains(raw, "/A1 gs") {
		t.Fatalf("expected content stream to select alpha ExtGState, got %q", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if !pdfDocumentBodyContains(doc, "/ExtGState << /A1") {
		t.Fatalf("page resources should reference ExtGState A1; objects: %#v", doc.Objects)
	}
	resourceBody := pdfDocumentObjectBodyContaining(doc, "/ExtGState << /A1")
	for _, want := range []string{"/Type /ExtGState", "/CA 0.5", "/ca 0.25"} {
		if !strings.Contains(resourceBody, want) {
			t.Fatalf("ExtGState resource missing %q:\n%s", want, resourceBody)
		}
	}
}

func TestRendererNativeHatchEmitsTilingPattern(t *testing.T) {
	r := newTestRenderer(t)
	hatcher, ok := any(r).(render.NativeHatcher)
	if !ok {
		t.Fatal("PDF renderer should implement render.NativeHatcher")
	}
	if !hatcher.SupportsNativeHatch() {
		t.Fatal("PDF renderer should report native hatch support")
	}

	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.Path(pdfTestRectPath(10, 10, 60, 50), &render.Paint{
		Fill:           render.Color{R: 0.9, G: 0.8, B: 0.7, A: 1},
		Hatch:          "/",
		HatchColor:     render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1},
		HatchLineWidth: 1.5,
		HatchSpacing:   8,
	})
	raw := r.content.String()
	if !strings.Contains(raw, "/Pattern cs") || !strings.Contains(raw, "/Pa1 scn") {
		t.Fatalf("expected page content to select hatch pattern, got %q", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	doc, err := pdfcompare.Parse(data)
	if err != nil {
		t.Fatalf("pdfcompare.Parse: %v", err)
	}
	if !pdfDocumentBodyContains(doc, "/Pattern << /Pa1") {
		t.Fatalf("page resources should reference hatch pattern Pa1; objects: %#v", doc.Objects)
	}
	patternBody := pdfDocumentObjectBodyContaining(doc, "/PatternType 1")
	for _, want := range []string{
		"/Type /Pattern",
		"/PaintType 1",
		"/TilingType 1",
		"/XStep 72",
		"/YStep 72",
		"0.9 0.8 0.7 rg 0 0 72 72 re f",
		"0.1 0.2 0.3 RG",
		"1.5 w",
		" S",
	} {
		if !strings.Contains(patternBody, want) {
			t.Fatalf("hatch pattern object missing %q:\n%s", want, patternBody)
		}
	}
}

func TestDrawMarkersEmitsReusableFormXObject(t *testing.T) {
	r := newTestRenderer(t)
	drawer, ok := any(r).(render.MarkerDrawer)
	if !ok {
		t.Fatal("PDF renderer should implement render.MarkerDrawer")
	}
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})

	ok = drawer.DrawMarkers(render.MarkerBatch{
		Marker: pdfTestTrianglePath(),
		Items: []render.MarkerItem{
			{
				Offset: geom.Pt{X: 20, Y: 30},
				Paint: render.Paint{
					Fill:      render.Color{R: 1, A: 1},
					Stroke:    render.Color{A: 1},
					LineWidth: 2,
				},
				Antialiased: true,
			},
			{
				Offset: geom.Pt{X: 40, Y: 50},
				Paint: render.Paint{
					Fill:      render.Color{G: 1, A: 1},
					Stroke:    render.Color{A: 1},
					LineWidth: 2,
				},
				Antialiased: true,
			},
		},
	})
	if !ok {
		t.Fatal("DrawMarkers returned false")
	}
	if got := strings.Count(r.content.String(), "/M1 Do"); got != 2 {
		t.Fatalf("expected two marker form invocations, got %d in %q", got, r.content.String())
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if !pdfDocumentBodyContains(doc, "/XObject << /M1") {
		t.Fatalf("page resources should reference marker form M1; objects: %#v", doc.Objects)
	}
	formBody := pdfDocumentObjectBodyContaining(doc, "/Subtype /Form")
	for _, want := range []string{"/Type /XObject", "/Subtype /Form", "/BBox", " m ", " l ", "B"} {
		if !strings.Contains(formBody, want) {
			t.Fatalf("marker form object missing %q:\n%s", want, formBody)
		}
	}
}

func TestDrawPathCollectionEmitsFormXObjects(t *testing.T) {
	r := newTestRenderer(t)
	drawer, ok := any(r).(render.PathCollectionDrawer)
	if !ok {
		t.Fatal("PDF renderer should implement render.PathCollectionDrawer")
	}
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})

	ok = drawer.DrawPathCollection(render.PathCollectionBatch{Items: []render.PathCollectionItem{
		{
			Path: pdfTestRectPath(10, 10, 30, 25),
			Paint: render.Paint{
				Fill: render.Color{B: 1, A: 1},
			},
			Antialiased: true,
		},
		{
			Path: pdfTestRectPath(40, 10, 65, 25),
			Paint: render.Paint{
				Stroke:    render.Color{R: 1, A: 1},
				LineWidth: 1,
			},
			Antialiased: true,
		},
	}})
	if !ok {
		t.Fatal("DrawPathCollection returned false")
	}
	if !strings.Contains(r.content.String(), "/P1 Do") || !strings.Contains(r.content.String(), "/P2 Do") {
		t.Fatalf("expected path collection form invocations, got %q", r.content.String())
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if !pdfDocumentBodyContains(doc, "/XObject << /P1") || !pdfDocumentBodyContains(doc, "/P2") {
		t.Fatalf("page resources should reference path collection forms; objects: %#v", doc.Objects)
	}
	if got := pdfDocumentObjectCountContaining(doc, "/Subtype /Form"); got < 2 {
		t.Fatalf("expected at least two form XObjects, got %d; objects: %#v", got, doc.Objects)
	}
}

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

func TestImageEmitsXObjectResourceAndDrawOperator(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(1, 0, color.RGBA{B: 255, A: 255})

	r.Image(render.NewImageData(img), geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 50, Y: 40}})

	raw := r.content.String()
	if !strings.Contains(raw, "/Im1 Do") {
		t.Fatalf("expected image draw operator in content stream, got %q", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	for _, want := range []string{"/XObject << /Im1", "/Subtype /Image", "/ColorSpace /DeviceRGB", "/Filter /FlateDecode"} {
		if !bytes.Contains(data, []byte(want)) {
			t.Fatalf("serialized PDF missing %q:\n%s", want, data)
		}
	}
}

func TestFlateImageEmitsPNGDecodeParms(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})

	r.Image(render.NewImageData(img), geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 50, Y: 40}})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	body := pdfDocumentObjectBodyContaining(doc, "/Subtype /Image")
	for _, want := range []string{
		"/Filter /FlateDecode",
		"/DecodeParms << /Predictor 10 /Colors 3 /Columns 2 /BitsPerComponent 8 >>",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("image object missing %q:\n%s", want, body)
		}
	}
}

func TestJPEGImageEmitsDCTDecodeXObject(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := jpegTestImage{
		w:    2,
		h:    1,
		data: []byte{0xff, 0xd8, 0xff, 0xd9},
	}

	r.Image(img, geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 50, Y: 40}})

	if !strings.Contains(r.content.String(), "/Im1 Do") {
		t.Fatalf("expected JPEG image draw operator, got %q", r.content.String())
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	for _, want := range []string{"/Subtype /Image", "/ColorSpace /DeviceRGB", "/Filter /DCTDecode"} {
		if !bytes.Contains(data, []byte(want)) {
			t.Fatalf("serialized PDF missing %q:\n%s", want, data)
		}
	}
}

func TestImageReusesXObjectForRepeatedImageData(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	data := render.NewImageData(img)

	r.Image(data, geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 30, Y: 40}})
	r.Image(data, geom.Rect{Min: geom.Pt{X: 40, Y: 20}, Max: geom.Pt{X: 60, Y: 40}})

	if got := strings.Count(r.content.String(), "/Im1 Do"); got != 2 {
		t.Fatalf("expected both draws to invoke reused Im1 XObject, got %d in %q", got, r.content.String())
	}
	if strings.Contains(r.content.String(), "/Im2 Do") {
		t.Fatalf("did not expect duplicate image XObject invocation in %q", r.content.String())
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if got := pdfDocumentObjectCountContaining(doc, "/Subtype /Image"); got != 1 {
		t.Fatalf("expected one image XObject for repeated image data, got %d; objects: %#v", got, doc.Objects)
	}
}

func TestImageWithAlphaEmitsSoftMask(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 20, G: 40, B: 60, A: 128})

	r.Image(render.NewImageData(img), geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 30, Y: 40}})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	for _, want := range []string{"/SMask", "/ColorSpace /DeviceGray"} {
		if !bytes.Contains(data, []byte(want)) {
			t.Fatalf("serialized PDF missing %q:\n%s", want, data)
		}
	}
}

func TestImageTransformedEmitsAffineImageMatrix(t *testing.T) {
	r := newTestRenderer(t)
	if _, ok := any(r).(render.ImageTransformer); !ok {
		t.Fatal("PDF renderer should implement render.ImageTransformer")
	}
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})

	r.ImageTransformed(render.NewImageData(img), geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 30, Y: 50}}, geom.Affine{
		A: 2,
		B: 0.5,
		C: -1,
		D: 3,
		E: 7,
		F: 11,
	})

	raw := r.content.String()
	if !strings.Contains(raw, "4 1 -3 9 7 11 cm") || !strings.Contains(raw, "/Im1 Do") {
		t.Fatalf("expected transformed image matrix and XObject invocation, got %q", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if !pdfDocumentBodyContains(doc, "/XObject << /Im1") {
		t.Fatalf("page resources should reference transformed image Im1; objects: %#v", doc.Objects)
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

func TestRendererClipRectEmitsRectangleClip(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.ClipRect(geom.Rect{Min: geom.Pt{X: 5, Y: 5}, Max: geom.Pt{X: 95, Y: 95}})
	raw := r.content.String()
	if !strings.Contains(raw, "5 5 90 90 re W n") {
		t.Errorf("expected rectangle clip operator in %q", raw)
	}
}

func TestRendererClipPathEmitsClipOperators(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: 10, Y: 0})
	p.LineTo(geom.Pt{X: 10, Y: 10})
	p.Close()
	r.ClipPath(p)
	raw := r.content.String()
	if !strings.Contains(raw, "W n\n") {
		t.Errorf("expected clip operator W n in %q", raw)
	}
}

func TestSavePDFWritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")

	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 50})
	r.Path(p, &render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 2,
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if err := r.SavePDF(path); err != nil {
		t.Fatalf("SavePDF: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-1.7\n")) {
		t.Errorf("missing PDF header in %q", data[:min(len(data), 16)])
	}
	if !bytes.Contains(data, []byte("startxref")) {
		t.Errorf("missing startxref")
	}
}

func TestSavePDFBeforeEndFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.pdf")

	r := newTestRenderer(t)
	if err := r.SavePDF(path); err == nil {
		t.Errorf("SavePDF before End should fail")
	}
}

func TestSerializationDeterministic(t *testing.T) {
	build := func() []byte {
		r := newTestRenderer(t)
		_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
		var p geom.Path
		p.MoveTo(geom.Pt{X: 10, Y: 10})
		p.LineTo(geom.Pt{X: 90, Y: 10})
		p.LineTo(geom.Pt{X: 90, Y: 50})
		p.Close()
		r.Path(p, &render.Paint{
			Fill:      render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1},
			Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
			LineWidth: 1,
		})
		_ = r.End()
		out, err := r.Bytes()
		if err != nil {
			t.Fatalf("Bytes: %v", err)
		}
		cp := make([]byte, len(out))
		copy(cp, out)
		return cp
	}
	a := build()
	b := build()
	if !bytes.Equal(a, b) {
		t.Errorf("PDF output is not deterministic; len(a)=%d len(b)=%d", len(a), len(b))
	}
}

func TestGeneratedPDFStructuralCompareIgnoresXRefOffsetNoise(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 50})
	r.Path(p, &render.Paint{Stroke: render.Color{A: 1}, LineWidth: 1})
	r.DrawTextWithFont("A", geom.Pt{X: 20, Y: 70}, 12, render.Color{A: 1}, "DejaVu Sans")
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	noisy := rewriteXRefOffsetsForTest(data)
	if bytes.Equal(data, noisy) {
		t.Fatal("test setup failed: xref rewrite did not change the PDF bytes")
	}

	diff, err := pdfcompare.ParseAndDiff(data, noisy)
	if err != nil {
		t.Fatalf("ParseAndDiff: %v", err)
	}
	if diff != "" {
		t.Fatalf("xref offset noise should not produce a structural diff, got: %s", diff)
	}
}

func TestQuadraticPromotedToCubic(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.QuadTo(geom.Pt{X: 10, Y: 20}, geom.Pt{X: 20, Y: 0})
	r.Path(p, &render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 1,
	})
	raw := r.content.String()
	// We promoted Quad to Cubic, so we should see a `c` curve operator but no
	// stray Quad-style operator.
	if !strings.Contains(raw, " c\n") {
		t.Errorf("expected cubic-curve operator c in %q", raw)
	}
}

func rewriteXRefOffsetsForTest(data []byte) []byte {
	out := append([]byte(nil), data...)
	xrefStart := bytes.Index(out, []byte("xref\n"))
	trailerStart := bytes.Index(out, []byte("trailer\n"))
	if xrefStart < 0 || trailerStart < 0 || trailerStart <= xrefStart {
		return out
	}
	for i := xrefStart; i+20 <= trailerStart; i++ {
		if (i == xrefStart || out[i-1] == '\n') && tenDigits(out[i:i+10]) && out[i+10] == ' ' {
			copy(out[i:i+10], []byte("9999999999"))
		}
	}
	startXRef := bytes.Index(out, []byte("startxref\n"))
	if startXRef >= 0 {
		valueStart := startXRef + len("startxref\n")
		valueEnd := valueStart
		for valueEnd < len(out) && out[valueEnd] >= '0' && out[valueEnd] <= '9' {
			out[valueEnd] = '1'
			valueEnd++
		}
	}
	return out
}

func tenDigits(b []byte) bool {
	if len(b) != 10 {
		return false
	}
	for _, c := range b {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func pdfTestRectPath(x0, y0, x1, y1 float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: x0, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y0})
	p.LineTo(geom.Pt{X: x1, Y: y1})
	p.LineTo(geom.Pt{X: x0, Y: y1})
	p.Close()
	return p
}

func pdfTestTrianglePath() geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: -4})
	p.LineTo(geom.Pt{X: 4, Y: 4})
	p.LineTo(geom.Pt{X: -4, Y: 4})
	p.Close()
	return p
}

func mustParsePDF(t *testing.T, r *Renderer) *pdfcompare.Document {
	t.Helper()
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	doc, err := pdfcompare.Parse(data)
	if err != nil {
		t.Fatalf("pdfcompare.Parse: %v", err)
	}
	return doc
}

func pdfDocumentBodyContains(doc *pdfcompare.Document, needle string) bool {
	return pdfDocumentObjectBodyContaining(doc, needle) != ""
}

func pdfDocumentObjectCountContaining(doc *pdfcompare.Document, needle string) int {
	if doc == nil {
		return 0
	}
	count := 0
	for _, obj := range doc.Objects {
		if strings.Contains(obj.Body, needle) {
			count++
		}
	}
	return count
}

func pdfDocumentObjectBodyContaining(doc *pdfcompare.Document, needle string) string {
	if doc == nil {
		return ""
	}
	for _, obj := range doc.Objects {
		if strings.Contains(obj.Body, needle) {
			return obj.Body
		}
	}
	return ""
}

func writePDFFakeCommand(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake command: %v", err)
	}
	return path
}

func writePDFTestPNG(t *testing.T, path string, c color.RGBA) {
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

type jpegTestImage struct {
	w, h int
	data []byte
}

func (j jpegTestImage) Size() (int, int)      { return j.w, j.h }
func (j jpegTestImage) Interpolation() string { return "" }
func (j jpegTestImage) JPEGData() []byte      { return append([]byte(nil), j.data...) }

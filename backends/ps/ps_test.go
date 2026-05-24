package ps

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
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

func TestImageEmitsPostScriptColorImage(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 0xff, A: 0xff})
	img.SetRGBA(1, 0, color.RGBA{G: 0xff, A: 0xff})

	r.Image(render.NewImageData(img), geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 50, Y: 40},
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if !bytes.Contains(r.document, []byte("2 1 8 [2 0 0 -1 0 1]")) {
		t.Fatalf("missing colorimage geometry in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("colorimage")) {
		t.Fatalf("missing colorimage operator in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("<ff000000ff00>")) {
		t.Fatalf("missing deterministic RGB image payload in\n%s", r.document)
	}
}

func TestPathEffectFilterFallbackPolicyIsDocumented(t *testing.T) {
	r := newTestRenderer(t)
	if _, ok := any(r).(render.FilterRenderer); ok {
		t.Fatal("PS should not expose offscreen filter support")
	}
	if _, ok := any(r).(render.PathEffectFilterDrawer); ok {
		t.Fatal("PS should not expose native filtered path-effect support")
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
			t.Fatalf("PS package docs should document %q fallback:\n%s", want, doc)
		}
	}
}

func TestMathTextDrawsRunsAndRules(t *testing.T) {
	fig := core.NewFigure(200, 100)
	fig.Text(0.5, 0.5, `$\frac{1}{2}+\alpha_i$`, core.TextOptions{
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignMiddle,
		FontSize: 18,
	})

	r := newTestRenderer(t)
	core.DrawFigure(fig, r)
	content := string(r.document)
	if strings.Count(content, "fill\n") < 5 {
		t.Fatalf("MathText should emit multiple filled PostScript paths, got\n%s", content)
	}
	if !strings.Contains(content, "closepath\n0 0 0 setrgbcolor\nfill\n") {
		t.Fatalf("MathText fraction rule should emit a filled path, got\n%s", content)
	}
	if strings.Contains(content, `$\\frac`) {
		t.Fatalf("MathText should be laid out before PS emission, got\n%s", content)
	}
}

func TestImageTransformedEmitsConcatMatrix(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	transformer, ok := any(r).(render.ImageTransformer)
	if !ok {
		t.Fatal("PS renderer should implement render.ImageTransformer")
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{B: 0xff, A: 0xff})
	img.SetRGBA(1, 0, color.RGBA{R: 0xff, G: 0xff, A: 0xff})

	transformer.ImageTransformed(render.NewImageData(img), geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 12, Y: 21},
	}, geom.Affine{
		A: 2, B: 0.5,
		C: -0.25, D: 3,
		E: 10, F: 20,
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if !bytes.Contains(r.document, []byte("[4 1 -0.25 3 10 20] concat")) {
		t.Fatalf("missing transformed image concat matrix in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("<0000ffffff00>")) {
		t.Fatalf("missing transformed image payload in\n%s", r.document)
	}
}

func TestRasterizedArtistEmbedsImageAndKeepsVectorText(t *testing.T) {
	fig := core.NewFigure(200, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.2, Y: 0.2}, Max: geom.Pt{X: 0.8, Y: 0.75}})
	line := ax.Plot([]float64{0, 0.5, 1}, []float64{0, 1, 0})
	line.SetRasterized(true)
	ax.SetTitle("Vector title")

	r := newTestRenderer(t)
	r.SetPSOptions(render.ResolvePSOptions(render.WithPSFontPolicy(render.PSFontPolicyBase14)))
	core.DrawFigure(fig, r)

	if !bytes.Contains(r.document, []byte("colorimage")) {
		t.Fatalf("rasterized artist did not emit a PostScript image:\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("(Vector title) show")) {
		t.Fatalf("surrounding title text was not preserved as vector text:\n%s", r.document)
	}
}

func TestRepeatedImagesReusePostScriptProcedure(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 0xff, A: 0xff})
	img.SetRGBA(1, 0, color.RGBA{G: 0xff, A: 0xff})
	data := render.NewImageData(img)

	r.Image(data, geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 50, Y: 40}})
	r.Image(data, geom.Rect{Min: geom.Pt{X: 60, Y: 20}, Max: geom.Pt{X: 100, Y: 40}})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if got := bytes.Count(r.document, []byte("/I1 {")); got != 1 {
		t.Fatalf("expected one image procedure definition, got %d in\n%s", got, r.document)
	}
	if got := bytes.Count(r.document, []byte("<ff000000ff00>")); got != 1 {
		t.Fatalf("expected one reused image payload, got %d in\n%s", got, r.document)
	}
	if got := bytes.Count(r.document, []byte("\nI1\ngrestore")); got != 2 {
		t.Fatalf("expected two image procedure invocations, got %d in\n%s", got, r.document)
	}
}

func TestPathWithHatchEmitsClippedHatchLines(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	hatcher, ok := any(r).(render.NativeHatcher)
	if !ok {
		t.Fatal("PS renderer should implement render.NativeHatcher")
	}
	if !hatcher.SupportsNativeHatch() {
		t.Fatal("SupportsNativeHatch returned false")
	}

	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 50})
	p.Close()
	r.Path(p, &render.Paint{
		Fill:           render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1},
		Hatch:          "/",
		HatchColor:     render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1},
		HatchLineWidth: 1.5,
		HatchSpacing:   8,
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if !bytes.Contains(r.document, []byte("clip newpath")) {
		t.Fatalf("missing hatch clip in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("0.1 0.2 0.3 setrgbcolor")) {
		t.Fatalf("missing hatch color in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("1.5 setlinewidth")) {
		t.Fatalf("missing hatch linewidth in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("lineto\nstroke")) {
		t.Fatalf("missing hatch stroke lines in\n%s", r.document)
	}
}

func TestPathWithShapeHatchEmitsClippedShapeGeometry(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 50})
	p.Close()
	r.Path(p, &render.Paint{
		Hatch:          "oO.*",
		HatchColor:     render.Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   12,
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	for _, want := range [][]byte{
		[]byte("curveto"),
		[]byte("closepath"),
		[]byte("fill"),
		[]byte("stroke"),
	} {
		if !bytes.Contains(r.document, want) {
			t.Fatalf("missing shape hatch fragment %q in\n%s", want, r.document)
		}
	}
}

func TestPartialAlphaVectorPaintIsDocumentedOpaque(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 50})
	p.Close()
	r.Path(p, &render.Paint{
		Fill:      render.Color{R: 0.2, G: 0.4, B: 0.6, A: 0.25},
		Stroke:    render.Color{R: 0.8, G: 0.1, B: 0.1, A: 0.5},
		LineWidth: 2,
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if !bytes.Contains(r.document, []byte("0.2 0.4 0.6 setrgbcolor")) {
		t.Fatalf("partial-alpha fill should be emitted as opaque RGB in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("0.8 0.1 0.1 setrgbcolor")) {
		t.Fatalf("partial-alpha stroke should be emitted as opaque RGB in\n%s", r.document)
	}
	if bytes.Contains(r.document, []byte("opacity")) || bytes.Contains(r.document, []byte("alpha")) {
		t.Fatalf("Level-2 PS should not claim native alpha state in\n%s", r.document)
	}
}

func TestDefaultTextPolicyEmitsGlyphPaths(t *testing.T) {
	r := newTestRenderer(t)
	if _, ok := any(r).(render.TextPather); !ok {
		t.Fatal("PS renderer should implement render.TextPather")
	}
	if _, ok := any(r).(render.PSOptionSetter); !ok {
		t.Fatal("PS renderer should implement render.PSOptionSetter")
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	r.DrawTextWithFont("Ag", geom.Pt{X: 20, Y: 50}, 18, render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}, "DejaVu Sans")

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if bytes.Contains(r.document, []byte("(Ag) show")) || bytes.Contains(r.document, []byte("/Helvetica findfont")) {
		t.Fatalf("default PS text policy should not emit direct Base14 text in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("0.1 0.2 0.3 setrgbcolor")) ||
		!bytes.Contains(r.document, []byte("fill")) {
		t.Fatalf("glyph-path text should emit filled colored paths in\n%s", r.document)
	}
}

func TestBase14TextPolicyEmitsDirectHelveticaText(t *testing.T) {
	r := newTestRenderer(t)
	r.SetPSOptions(render.ResolvePSOptions(render.WithPSFontPolicy(render.PSFontPolicyBase14)))
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	r.DrawTextRotatedWithFont(`A()`, geom.Pt{X: 20, Y: 50}, 18, math.Pi/2, render.Color{A: 1}, "DejaVu Sans")

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if !bytes.Contains(r.document, []byte("/Helvetica findfont 18 scalefont setfont")) {
		t.Fatalf("Base14 policy should use Helvetica in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte(`(A\(\)) show`)) {
		t.Fatalf("Base14 policy should escape direct text in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("90 rotate")) {
		t.Fatalf("Base14 rotated text should convert radians to PS degrees in\n%s", r.document)
	}
}

func TestDrawMarkersEmitsReusableProcedure(t *testing.T) {
	r := newTestRenderer(t)
	drawer, ok := any(r).(render.MarkerDrawer)
	if !ok {
		t.Fatal("PS renderer should implement render.MarkerDrawer")
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	var marker geom.Path
	marker.MoveTo(geom.Pt{X: 0, Y: 0})
	marker.LineTo(geom.Pt{X: 5, Y: 10})
	marker.LineTo(geom.Pt{X: 10, Y: 0})
	marker.Close()

	ok = drawer.DrawMarkers(render.MarkerBatch{
		Marker: marker,
		Items: []render.MarkerItem{
			{
				Offset: geom.Pt{X: 20, Y: 30},
				Paint:  render.Paint{Fill: render.Color{R: 1, A: 1}},
			},
			{
				Offset: geom.Pt{X: 40, Y: 50},
				Paint:  render.Paint{Fill: render.Color{R: 1, A: 1}},
			},
		},
	})
	if !ok {
		t.Fatal("DrawMarkers returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if got := bytes.Count(r.document, []byte("/M1 {")); got != 1 {
		t.Fatalf("expected one marker procedure definition, got %d in\n%s", got, r.document)
	}
	if !bytes.Contains(r.document, []byte("20 30 translate\nM1")) ||
		!bytes.Contains(r.document, []byte("40 50 translate\nM1")) {
		t.Fatalf("missing translated marker invocations in\n%s", r.document)
	}
	if got := bytes.Count(r.document, []byte("\nM1\ngrestore")); got != 2 {
		t.Fatalf("expected two marker procedure invocations, got %d in\n%s", got, r.document)
	}
}

func TestDrawPathCollectionEmitsReusableProcedure(t *testing.T) {
	r := newTestRenderer(t)
	drawer, ok := any(r).(render.PathCollectionDrawer)
	if !ok {
		t.Fatal("PS renderer should implement render.PathCollectionDrawer")
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	var path geom.Path
	path.MoveTo(geom.Pt{X: 10, Y: 10})
	path.LineTo(geom.Pt{X: 20, Y: 10})
	path.LineTo(geom.Pt{X: 20, Y: 20})
	path.Close()

	ok = drawer.DrawPathCollection(render.PathCollectionBatch{
		Items: []render.PathCollectionItem{
			{
				Path:  path,
				Paint: render.Paint{Fill: render.Color{G: 1, A: 1}},
			},
			{
				Path:  path,
				Paint: render.Paint{Fill: render.Color{G: 1, A: 1}},
			},
		},
	})
	if !ok {
		t.Fatal("DrawPathCollection returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if got := bytes.Count(r.document, []byte("/P1 {")); got != 1 {
		t.Fatalf("expected one path collection procedure definition, got %d in\n%s", got, r.document)
	}
	if got := bytes.Count(r.document, []byte("P1\n")); got != 2 {
		t.Fatalf("expected two path collection procedure invocations, got %d in\n%s", got, r.document)
	}
}

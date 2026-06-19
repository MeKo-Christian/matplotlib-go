package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/pdfcompare"
	"github.com/cwbudde/matplotlib-go/render"
)

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

func TestPathEffectIdentityFilterEmitsTransparencyGroup(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})

	var p geom.Path
	p.MoveTo(geom.Pt{X: 20, Y: 60})
	p.LineTo(geom.Pt{X: 180, Y: 60})
	r.Path(p, &render.Paint{
		Stroke:    render.Color{B: 1, A: 1},
		LineWidth: 2,
		PathEffects: []render.PathEffect{
			render.FilterPathEffect(
				render.Color{},
				render.Color{R: 1, A: 1},
				8,
				"identity",
				0,
				geom.Pt{X: 2, Y: 2},
			),
			render.NormalPathEffect(),
		},
	})

	raw := r.content.String()
	if !strings.Contains(raw, "/E1 Do") {
		t.Fatalf("expected page content to invoke path-effect Form XObject, got %q", raw)
	}
	if strings.Contains(raw, "/Im") {
		t.Fatalf("identity filter path effect should not rasterize to an image XObject, got %q", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if !pdfDocumentBodyContains(doc, "/XObject << /E1") {
		t.Fatalf("page resources should reference path-effect form E1; objects: %#v", doc.Objects)
	}
	formBody := pdfDocumentObjectBodyContaining(doc, "/Subtype /Form")
	for _, want := range []string{
		"/Subtype /Form",
		"/Group << /S /Transparency",
		"1 0 0 RG",
		"8 w",
		"S",
	} {
		if !strings.Contains(formBody, want) {
			t.Fatalf("path-effect form object missing %q:\n%s", want, formBody)
		}
	}
}

func TestPathEffectBlurFilterSoftMaskPolicyIsDocumented(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(".", "doc.go"))
	if err != nil {
		t.Fatalf("read doc.go: %v", err)
	}
	doc := strings.Join(strings.Fields(strings.ReplaceAll(string(data), "\n// ", " ")), " ")
	for _, want := range []string{
		"identity path-effect filters",
		"blurred path-effect filters",
		"soft-mask image XObjects",
		"without routing through core mixed-raster fallback",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("pdf doc.go missing path-effect blur policy phrase %q:\n%s", want, doc)
		}
	}
}

func TestPathEffectBlurFilterEmitsSoftMaskImage(t *testing.T) {
	r := newTestRenderer(t)
	blur := render.FilterPathEffect(
		render.Color{R: 1, A: 1},
		render.Color{},
		0,
		"blur",
		4,
		geom.Pt{X: 2, Y: 2},
	)
	if !r.SupportsPathEffectFilter(blur) {
		t.Fatal("PDF should report native support for blurred path-effect soft masks")
	}
	identity := render.FilterPathEffect(
		render.Color{R: 1, A: 1},
		render.Color{},
		0,
		"identity",
		0,
		geom.Pt{X: 2, Y: 2},
	)
	if !r.SupportsPathEffectFilter(identity) {
		t.Fatal("PDF should keep identity path-effect filters in native transparency groups")
	}

	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.Path(pdfTestRectPath(40, 30, 50, 30), &render.Paint{
		Fill: render.Color{B: 1, A: 1},
		PathEffects: []render.PathEffect{
			blur,
			render.NormalPathEffect(),
		},
	})

	raw := r.content.String()
	if strings.Contains(raw, "StartRasterized") {
		t.Fatalf("blurred path effect should not route through core mixed-raster fallback, got %q", raw)
	}
	if !strings.Contains(raw, "/Im1 Do") {
		t.Fatalf("expected blurred filter pass to invoke an image XObject, got %q", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	imageBody := pdfDocumentObjectBodyContaining(doc, "/Subtype /Image")
	for _, want := range []string{"/Subtype /Image", "/SMask"} {
		if !strings.Contains(imageBody, want) {
			t.Fatalf("blurred filter image object missing %q:\n%s", want, imageBody)
		}
	}
	if !pdfDocumentBodyContains(doc, "/ColorSpace /DeviceGray") {
		t.Fatalf("blurred filter image should emit a grayscale soft-mask object; objects: %#v", doc.Objects)
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

func TestRendererNativeShapeHatchEmitsPatternGeometry(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.Path(pdfTestRectPath(10, 10, 60, 50), &render.Paint{
		Hatch:          "oO.*",
		HatchColor:     render.Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   12,
	})
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
	patternBody := pdfDocumentObjectBodyContaining(doc, "/PatternType 1")
	for _, want := range []string{" c ", " h ", " S", " f "} {
		if !strings.Contains(patternBody, want) {
			t.Fatalf("shape hatch pattern object missing %q:\n%s", want, patternBody)
		}
	}
}

func TestRendererPatternFillEmitsTilingPattern(t *testing.T) {
	r := newTestRenderer(t)
	filler, ok := any(r).(render.PatternFiller)
	if !ok {
		t.Fatal("PDF renderer should implement render.PatternFiller")
	}
	if !filler.SupportsPatternFill() {
		t.Fatal("PDF renderer should report native pattern fill support")
	}

	var tile geom.Path
	tile.MoveTo(geom.Pt{X: 4, Y: 4})
	tile.LineTo(geom.Pt{X: 12, Y: 4})
	tile.LineTo(geom.Pt{X: 12, Y: 12})
	tile.LineTo(geom.Pt{X: 4, Y: 12})
	tile.Close()

	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.Path(pdfTestRectPath(20, 20, 90, 70), &render.Paint{
		FillPattern: render.PatternFill{
			ID:         "dots",
			Cell:       geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 16, Y: 16}},
			Path:       tile,
			Foreground: render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1},
			Background: render.Color{R: 1, G: 1, B: 1, A: 1},
			LineWidth:  1.25,
		},
	})

	raw := r.content.String()
	if !strings.Contains(raw, "/Pattern cs") || !strings.Contains(raw, "/Pf1 scn") {
		t.Fatalf("expected page content to select pattern fill, got %q", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if !pdfDocumentBodyContains(doc, "/Pattern << /Pf1") {
		t.Fatalf("page resources should reference fill pattern Pf1; objects: %#v", doc.Objects)
	}
	patternBody := pdfDocumentObjectBodyContaining(doc, "/PatternType 1")
	for _, want := range []string{
		"/Type /Pattern",
		"/PaintType 1",
		"/BBox [ 0 0 16 16 ]",
		"/XStep 16",
		"/YStep 16",
		"1 1 1 rg 0 0 16 16 re f",
		"0.2 0.4 0.8 rg",
		"1.25 w",
		"f",
	} {
		if !strings.Contains(patternBody, want) {
			t.Fatalf("fill pattern object missing %q:\n%s", want, patternBody)
		}
	}
}

func TestRendererLinearGradientEmitsAxialShading(t *testing.T) {
	r := newTestRenderer(t)
	filler, ok := any(r).(render.GradientFiller)
	if !ok {
		t.Fatal("PDF renderer should implement render.GradientFiller")
	}
	if !filler.SupportsGradientFill() {
		t.Fatal("PDF renderer should report native gradient fill support")
	}

	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.Path(pdfTestRectPath(10, 10, 60, 50), &render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 10, Y: 10},
			End:   geom.Pt{X: 60, Y: 10},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{B: 1, A: 1}},
			},
		},
	})

	raw := r.content.String()
	for _, want := range []string{"W\nn\n", "/Sh1 sh"} {
		if !strings.Contains(raw, want) {
			t.Fatalf("linear gradient content missing %q:\n%s", want, raw)
		}
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if !pdfDocumentBodyContains(doc, "/Shading << /Sh1") {
		t.Fatalf("page resources should reference shading Sh1; objects: %#v", doc.Objects)
	}
	shadingBody := pdfDocumentObjectBodyContaining(doc, "/ShadingType 2")
	for _, want := range []string{
		"/ShadingType 2",
		"/ColorSpace /DeviceRGB",
		"/Coords [ 10 10 60 10 ]",
		"/C0 [ 1 0 0 ]",
		"/C1 [ 0 0 1 ]",
	} {
		if !strings.Contains(shadingBody, want) {
			t.Fatalf("linear shading object missing %q:\n%s", want, shadingBody)
		}
	}
}

func TestRendererRadialGradientEmitsRadialShadingAndStroke(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	r.Path(pdfTestRectPath(20, 15, 70, 65), &render.Paint{
		Stroke:    render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1},
		LineWidth: 2,
		FillGradient: render.GradientFill{
			Kind:   render.RadialGradient,
			Center: geom.Pt{X: 45, Y: 40},
			Radius: 30,
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, G: 1, A: 1}},
				{Offset: 1, Color: render.Color{A: 1}},
			},
		},
	})

	raw := r.content.String()
	for _, want := range []string{"/Sh1 sh", "0.1 0.2 0.3 RG", "2 w", "S\n"} {
		if !strings.Contains(raw, want) {
			t.Fatalf("radial gradient content missing %q:\n%s", want, raw)
		}
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	shadingBody := pdfDocumentObjectBodyContaining(doc, "/ShadingType 3")
	for _, want := range []string{
		"/ShadingType 3",
		"/Coords [ 45 40 0 45 40 30 ]",
		"/C0 [ 1 1 0 ]",
		"/C1 [ 0 0 0 ]",
	} {
		if !strings.Contains(shadingBody, want) {
			t.Fatalf("radial shading object missing %q:\n%s", want, shadingBody)
		}
	}
}

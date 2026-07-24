package svg

import (
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"codeberg.org/go-fonts/dejavu/dejavusans"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestFontPolicyPathDrawsTextAsPath(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.SetSVGOptions(render.ResolveSVGOptions(render.WithSVGFontPolicy(render.SVGFontPolicyPath)))
		r.DrawText("A", geom.Pt{X: 10, Y: 30}, 14, render.Color{A: 1})
	})

	if strings.Contains(content, "<text") {
		t.Fatalf("path font policy should not emit native text nodes, got %q", content)
	}
	if !strings.Contains(content, `<path d="`) || !strings.Contains(content, `fill="rgb(0,0,0)"`) {
		t.Fatalf("path font policy should emit filled glyph paths, got %q", content)
	}
}

func TestRenderSVGDrawsMathTextRunsAndRules(t *testing.T) {
	fig := core.NewFigure(180, 120)
	fig.Text(0.5, 0.5, `$\frac{1}{2}+\alpha_i$`, core.TextOptions{
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignMiddle,
		FontSize: 18,
	})

	r := mustNewRenderer(t)
	core.DrawFigure(fig, r)
	content := r.renderSVG()
	if strings.Count(content, "<text") < 4 {
		t.Fatalf("MathText should emit multiple SVG text runs, got %q", content)
	}
	if !strings.Contains(content, `<path d="`) || !strings.Contains(content, `fill="rgb(0,0,0)"`) {
		t.Fatalf("MathText fraction rule should emit a filled SVG path, got %q", content)
	}
	if strings.Contains(content, `$\frac`) {
		t.Fatalf("MathText should be laid out before SVG emission, got %q", content)
	}
}

func TestDrawTeXEmbedsCachedPNG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell commands are POSIX-only")
	}
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture.png")
	writeSVGTestPNG(t, fixture, color.RGBA{A: 255})
	latex := writeSVGFakeCommand(t, dir, "latex", `#!/bin/sh
touch file.dvi
`)
	dvipng := writeSVGFakeCommand(t, dir, "dvipng", `#!/bin/sh
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

	r := mustNewRenderer(t)
	r.texManager = tex.NewManager(tex.ManagerConfig{
		CacheDir:      filepath.Join(dir, "cache"),
		LaTeXCommand:  latex,
		DVIPNGCommand: dvipng,
	})

	if metrics, ok := r.MeasureTeX(`signal $\alpha$`, 12, "DejaVu Sans"); !ok || metrics.W != 2 || metrics.H != 2 {
		t.Fatalf("MeasureTeX = %+v, %v; want 2x2 metrics and ok", metrics, ok)
	}
	if !r.DrawTeX(`signal $\alpha$`, geom.Pt{X: 8, Y: 10}, 12, render.Color{R: 1, A: 1}, "DejaVu Sans") {
		t.Fatal("DrawTeX returned false")
	}

	content := r.renderSVG()
	if !strings.Contains(content, `<image x="8" y="8" width="2" height="2"`) {
		t.Fatalf("TeX image should be placed from baseline-adjusted origin, got %q", content)
	}
	if !strings.Contains(content, `data:image/png;base64,`) {
		t.Fatalf("TeX image should be embedded as a PNG data URI, got %q", content)
	}
	if strings.Contains(content, `>signal`) {
		t.Fatalf("TeX text should not fall back to native SVG text, got %q", content)
	}
}

func TestDrawTeXRotatedEmbedsCachedPNGWithTransform(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell commands are POSIX-only")
	}
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture.png")
	writeSVGTestPNG(t, fixture, color.RGBA{A: 255})
	latex := writeSVGFakeCommand(t, dir, "latex", `#!/bin/sh
touch file.dvi
`)
	dvipng := writeSVGFakeCommand(t, dir, "dvipng", `#!/bin/sh
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

	r := mustNewRenderer(t)
	r.texManager = tex.NewManager(tex.ManagerConfig{
		CacheDir:      filepath.Join(dir, "cache"),
		LaTeXCommand:  latex,
		DVIPNGCommand: dvipng,
	})

	if !r.DrawTeXRotated(`x`, geom.Pt{X: 20, Y: 30}, 12, math.Pi/2, render.Color{B: 1, A: 1}, "DejaVu Sans") {
		t.Fatal("DrawTeXRotated returned false")
	}

	content := r.renderSVG()
	if !strings.Contains(content, `<image`) || !strings.Contains(content, `transform="matrix(0 -1 1 0 -10 50)"`) {
		t.Fatalf("rotated TeX should embed an image with anchor rotation, got %q", content)
	}
}

func TestDrawTextSupportsNegativeCoordinates(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	r.DrawText("neg", geom.Pt{X: -15, Y: 30}, 12, render.Color{R: 0})
	r.End()

	tmp, err := os.CreateTemp("", "matplotlib-go-svg-negative-*.svg")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmpPath) })

	if err := r.SaveSVG(tmpPath); err != nil {
		t.Fatalf("SaveSVG failed: %v", err)
	}
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "x=\"-15\"") {
		t.Fatalf("expected preserved negative x coordinate, got %q", content)
	}
}

func TestGlyphRunUsesFontKeyAndOffsetsWithoutAccumulatingOffset(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.GlyphRun(render.GlyphRun{
			Glyphs: []render.Glyph{
				{ID: 'A', Advance: 8, Offset: geom.Pt{X: 2, Y: 3}},
				{ID: 'B', Advance: 7, Offset: geom.Pt{X: 1, Y: -1}},
			},
			Origin:  geom.Pt{X: 10, Y: 20},
			Size:    12,
			FontKey: "sans-serif",
		}, render.Color{A: 1})
	})

	if !strings.Contains(content, `font-family="DejaVu Sans, Arial, sans-serif"`) {
		t.Fatalf("glyph run should honor sans-serif font selection, got %q", content)
	}
	if !strings.Contains(content, `<text x="12" y="97"`) {
		t.Fatalf("first glyph should render at origin plus its offset, got %q", content)
	}
	if !strings.Contains(content, `<text x="19" y="101"`) {
		t.Fatalf("second glyph should advance from origin without accumulating prior offsets, got %q", content)
	}
}

func TestBeginResetsLastFontKey(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}

	if err := r.Begin(viewport); err != nil {
		t.Fatalf("first Begin failed: %v", err)
	}
	r.MeasureText("sample", 12, "monospace")
	r.DrawText("first", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("first End failed: %v", err)
	}

	if err := r.Begin(viewport); err != nil {
		t.Fatalf("second Begin failed: %v", err)
	}
	r.DrawText("second", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("second End failed: %v", err)
	}

	content := r.renderSVG()
	if strings.Contains(content, `font-family="DejaVu Sans Mono, monospace"`) {
		t.Fatalf("plain text should not inherit font family from a previous drawing session, got %q", content)
	}
	if !strings.Contains(content, `font-family="DejaVu Sans, Arial, sans-serif"`) {
		t.Fatalf("plain text should fall back to default sans font family, got %q", content)
	}
}

func TestTextEscapingAndRotationSerialization(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.DrawTextRotated(`A<&"B`, geom.Pt{X: 90, Y: 70}, 12, 0.5, render.Color{R: 0.2, G: 0.4, B: 0.6, A: 0.75})
	})

	for _, want := range []string{
		`transform="matrix(0.877583 0.479426 -0.479426 0.877583 34.988846 -37.027427)"`,
		`fill="rgb(51,102,153)"`,
		`fill-opacity="0.75"`,
		`A&lt;&amp;&#34;B`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected rotated text attribute %q in %q", want, content)
		}
	}
}

func TestDrawTextVerticalEmitsOneNodePerRune(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.DrawTextVertical("AB", geom.Pt{X: 90, Y: 60}, 12, render.Color{A: 1})
	})

	if strings.Count(content, "<text") != 2 {
		t.Fatalf("expected one text node per rune for vertical text, got %q", content)
	}
	if !strings.Contains(content, ">A</text>") || !strings.Contains(content, ">B</text>") {
		t.Fatalf("expected both vertical glyph nodes in %q", content)
	}
}

func TestFontFamilyVariants(t *testing.T) {
	tests := map[string]string{
		"serif":       "DejaVu Serif, serif",
		"sans-serif":  "DejaVu Sans, Arial, sans-serif",
		"monospace":   "DejaVu Sans Mono, monospace",
		"mono_space":  "DejaVu Sans Mono, monospace",
		"custom-font": "DejaVu Sans, Arial, sans-serif",
	}

	for key, want := range tests {
		if got := fontFamily(key); got != want {
			t.Fatalf("unexpected font family for %q: want %q, got %q", key, want, got)
		}
	}
}

func TestNativeTextEmitsStructuredFontProperties(t *testing.T) {
	fontKey := render.FontPropertiesKey(render.FontProperties{
		Families: []string{"First Face", "Second Face", "serif"},
		Style:    render.FontStyleItalic,
		Weight:   700,
		Stretch:  "condensed",
		Variant:  "small-caps",
	})
	content := renderSVGDocument(t, func(r *Renderer) {
		r.MeasureText("styled", 14, fontKey)
		r.DrawText("styled", geom.Pt{X: 10, Y: 30}, 14, render.Color{A: 1})
	})
	for _, attr := range []string{
		`font-family="First Face, Second Face, serif"`,
		`font-style="italic"`,
		`font-weight="700"`,
		`font-stretch="condensed"`,
		`font-variant="small-caps"`,
	} {
		if !strings.Contains(content, attr) {
			t.Errorf("SVG text missing %s: %s", attr, content)
		}
	}
}

func TestDrawTextEmbedsDirectFontFile(t *testing.T) {
	dir := t.TempDir()
	fontPath := filepath.Join(dir, "DejaVuSans.ttf")
	if err := os.WriteFile(fontPath, dejavusans.TTF, 0o644); err != nil {
		t.Fatalf("write test font: %v", err)
	}

	content := renderSVGDocument(t, func(r *Renderer) {
		r.MeasureText("embedded", 12, fontPath)
		r.DrawText("embedded", geom.Pt{X: 20, Y: 30}, 12, render.Color{A: 1})
	})

	if !strings.Contains(content, "@font-face") {
		t.Fatalf("expected embedded @font-face, got %q", content)
	}
	if !strings.Contains(content, "data:font/ttf;base64,") {
		t.Fatalf("expected embedded font data URI, got %q", content)
	}
	if !strings.Contains(content, `font-family="mplgo-font-1"`) {
		t.Fatalf("expected text node to use embedded font family, got %q", content)
	}
}

func TestDrawTextAndRotationGuardsSkipInvalidInput(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.DrawText("", geom.Pt{X: 10, Y: 10}, 12, render.Color{A: 1})
		r.DrawText("zero", geom.Pt{X: 10, Y: 10}, 0, render.Color{A: 1})
		r.DrawTextRotated("nan", geom.Pt{X: 10, Y: 10}, 12, math.NaN(), render.Color{A: 1})
		r.DrawTextRotated("inf", geom.Pt{X: 10, Y: 10}, 12, math.Inf(1), render.Color{A: 1})
		r.DrawTextRotated("zero", geom.Pt{X: 10, Y: 10}, 0, 1, render.Color{A: 1})
		r.DrawTextVertical("", geom.Pt{X: 10, Y: 10}, 12, render.Color{A: 1})
		r.DrawTextVertical("zero", geom.Pt{X: 10, Y: 10}, 0, render.Color{A: 1})
	})

	if strings.Contains(content, "<text") {
		t.Fatalf("invalid text inputs should not emit text nodes, got %q", content)
	}
}

func TestGlyphRunSkipsMissingGlyphsAndFallsBackToMeasuredAdvance(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	expectedAdvance := r.MeasureText("A", 12, "monospace").W
	r.nodes = nil

	r.GlyphRun(render.GlyphRun{
		Glyphs: []render.Glyph{
			{ID: 0, Advance: 5},
			{ID: 'A', Advance: 0},
			{ID: 'B', Advance: 4},
		},
		Origin:  geom.Pt{X: 10, Y: 20},
		Size:    12,
		FontKey: "monospace",
	}, render.Color{A: 1})

	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	content := r.renderSVG()
	if strings.Count(content, "<text") != 2 {
		t.Fatalf("expected only visible glyphs to emit text nodes, got %q", content)
	}
	if !strings.Contains(content, `<text x="15" y="100"`) {
		t.Fatalf("expected skipped glyph advance to shift first visible glyph, got %q", content)
	}
	secondX := `x="` + formatFloat(15+expectedAdvance) + `"`
	if !strings.Contains(content, secondX) {
		t.Fatalf("expected measured advance fallback to place second glyph at %s in %q", secondX, content)
	}
	if !strings.Contains(content, `font-family="DejaVu Sans Mono, monospace"`) {
		t.Fatalf("glyph run should propagate font key to rendered glyphs, got %q", content)
	}
}

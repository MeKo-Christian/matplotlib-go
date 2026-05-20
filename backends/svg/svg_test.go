package svg

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"codeberg.org/go-fonts/dejavu/dejavusans"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
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

func TestNewInvalidDimensions(t *testing.T) {
	r, err := New(0, 10, render.Color{})
	if err == nil || r != nil {
		t.Fatal("expected error for non-positive dimensions")
	}
}

func TestSerializationDeterministic(t *testing.T) {
	draw := func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 10, Y: 10})
		path.LineTo(geom.Pt{X: 170, Y: 110})
		path.Close()
		r.Path(path, &render.Paint{
			Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
			Fill:      render.Color{R: 0.5, G: 0.25, B: 0.1, A: 0.5},
			LineWidth: 2,
		})
		r.DrawText("hello", geom.Pt{X: 20, Y: 30}, 14, render.Color{R: 0, G: 0, B: 0, A: 1})
		r.DrawTextRotated("rot", geom.Pt{X: 80, Y: 60}, 12, 0.5, render.Color{R: 0, G: 0, B: 0, A: 1})
	}

	first := renderSVGDocument(t, draw)
	second := renderSVGDocument(t, draw)

	if first != second {
		t.Fatalf("SVG output not deterministic across renders:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestSaveSVG(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	var path geom.Path
	path.MoveTo(geom.Pt{X: 10, Y: 10})
	path.LineTo(geom.Pt{X: 170, Y: 110})
	r.Path(path, &render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 2,
	})

	r.DrawText("line", geom.Pt{X: 20, Y: 30}, 14, render.Color{R: 0, G: 0, B: 0, A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	tmp, err := os.CreateTemp("", "matplotlib-go-svg-*.svg")
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

	if !strings.Contains(content, "<svg") || !strings.Contains(content, "</svg>") {
		t.Fatal("SVG output missing root element")
	}
	if !strings.Contains(content, "<path") {
		t.Fatal("SVG output missing path node")
	}
	if !strings.Contains(content, "<text") || !strings.Contains(content, ">line<") {
		t.Fatal("SVG output missing text node")
	}
}

func TestRasterizedArtistEmbedsImageWhileKeepingTextVector(t *testing.T) {
	fig := core.NewFigure(180, 120)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.2, Y: 0.2}, Max: geom.Pt{X: 0.8, Y: 0.75}})
	line := ax.Plot([]float64{0, 0.5, 1}, []float64{0, 1, 0})
	line.SetRasterized(true)
	ax.SetTitle("Vector title")

	r := mustNewRenderer(t)
	core.DrawFigure(fig, r)

	tmp, err := os.CreateTemp("", "matplotlib-go-mixed-svg-*.svg")
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
	if !strings.Contains(content, "<image") {
		t.Fatalf("rasterized artist did not emit an SVG image: %s", content)
	}
	if !strings.Contains(content, `<g clip-path="url(#`) || !strings.Contains(content, `><image`) {
		t.Fatalf("rasterized image should preserve the active axes clip: %s", content)
	}
	if !strings.Contains(content, "Vector title") {
		t.Fatalf("surrounding title text was not preserved as vector SVG text: %s", content)
	}
}

func TestRenderSVGEmitsDefaultEmptyMetadata(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.DrawText("plain", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1})
	})

	if !strings.Contains(content, "  <metadata></metadata>\n") {
		t.Fatalf("default SVG should contain an empty metadata block, got %q", content)
	}
	if strings.Contains(content, `name="Date"`) {
		t.Fatalf("default SVG metadata should not include a generated date, got %q", content)
	}
}

func TestRenderSVGMetadataOptionAndSourceDateEpoch(t *testing.T) {
	t.Setenv("SOURCE_DATE_EPOCH", "0")
	content := renderSVGDocument(t, func(r *Renderer) {
		r.SetSVGOptions(render.ResolveSVGOptions(
			render.WithSVGMetadata(map[string]string{
				"Title":  "Plot",
				"Author": "Codex",
			}),
		))
	})

	for _, want := range []string{
		`<meta name="Author" content="Codex" />`,
		`<meta name="Date" content="1970-01-01T00:00:00Z" />`,
		`<meta name="Title" content="Plot" />`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected metadata fragment %q in %q", want, content)
		}
	}
	if strings.Index(content, `name="Author"`) > strings.Index(content, `name="Title"`) {
		t.Fatalf("metadata entries should serialize in deterministic key order, got %q", content)
	}
}

func TestHashSaltUsesContentDerivedIDs(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.SetSVGOptions(render.ResolveSVGOptions(render.WithSVGHashSalt("stable")))
		r.ClipRect(geom.Rect{Min: geom.Pt{X: 1, Y: 2}, Max: geom.Pt{X: 11, Y: 12}})
		r.DrawText("clipped", geom.Pt{X: 5, Y: 6}, 12, render.Color{A: 1})
	})

	if strings.Contains(content, `id="clip1"`) || strings.Contains(content, `url(#clip1)`) {
		t.Fatalf("hash-salted SVG should not use sequential clip IDs, got %q", content)
	}
	re := regexp.MustCompile(`id="clip[0-9a-f]{10}"`)
	if !re.MatchString(content) {
		t.Fatalf("hash-salted clip ID should use a stable content hash, got %q", content)
	}
}

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

func TestSaveSVGPreservesClip(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	r.ClipRect(geom.Rect{
		Min: geom.Pt{X: 10, Y: 10},
		Max: geom.Pt{X: 50, Y: 50},
	})
	r.DrawText("clipped", geom.Pt{X: 20, Y: 20}, 12, render.Color{R: 1})
	r.End()

	tmp, err := os.CreateTemp("", "matplotlib-go-svg-clip-*.svg")
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
	if !strings.Contains(content, "<clipPath") {
		t.Fatal("SVG output should contain clipPath definitions")
	}
	if !strings.Contains(content, "clip-path=\"url(#") {
		t.Fatal("SVG output should apply clip-path to content")
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
	if !strings.Contains(content, `<text x="12" y="23"`) {
		t.Fatalf("first glyph should render at origin plus its offset, got %q", content)
	}
	if !strings.Contains(content, `<text x="19" y="19"`) {
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

func TestRenderSVGPreservesClipStackAcrossSaveRestore(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipRect(geom.Rect{
			Min: geom.Pt{X: 5, Y: 5},
			Max: geom.Pt{X: 50, Y: 60},
		})
		r.Save()
		r.ClipRect(geom.Rect{
			Min: geom.Pt{X: 10, Y: 20},
			Max: geom.Pt{X: 30, Y: 40},
		})
		r.DrawText("inner", geom.Pt{X: 15, Y: 25}, 12, render.Color{A: 1})
		r.Restore()
		r.DrawText("outer", geom.Pt{X: 15, Y: 25}, 12, render.Color{A: 1})
	})

	if strings.Count(content, "<clipPath") != 2 {
		t.Fatalf("expected two clip path definitions after nested clipping, got %q", content)
	}
	if !strings.Contains(content, `<rect x="5" y="5" width="45" height="55" />`) {
		t.Fatalf("missing outer clip rect in SVG defs: %q", content)
	}
	if !strings.Contains(content, `<rect x="10" y="20" width="20" height="20" />`) {
		t.Fatalf("missing intersected inner clip rect in SVG defs: %q", content)
	}

	re := regexp.MustCompile(`<g clip-path="url\(#(clip\d+)\)"><text[^>]*>inner</text></g>\s*<g clip-path="url\(#(clip\d+)\)"><text[^>]*>outer</text></g>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) != 3 {
		t.Fatalf("expected clipped inner and restored outer groups, got %q", content)
	}
	if matches[1] == matches[2] {
		t.Fatalf("expected restore to switch back to the outer clip, got %q", content)
	}
}

func TestPathSerializesStrokeFillOpacityAndDashes(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 10, Y: 15})
		path.LineTo(geom.Pt{X: 60, Y: 15})
		path.LineTo(geom.Pt{X: 60, Y: 45})
		path.Close()

		r.Path(path, &render.Paint{
			LineWidth:  2.5,
			LineJoin:   render.JoinRound,
			LineCap:    render.CapSquare,
			MiterLimit: 7,
			Stroke:     render.Color{R: 1, G: 0, B: 0, A: 0.25},
			Fill:       render.Color{G: 1, A: 0.5},
			Dashes:     []float64{4, 2, 1, 3},
		})
	})

	for _, want := range []string{
		`fill="rgb(0,255,0)"`,
		`fill-opacity="0.5"`,
		`stroke="rgb(255,0,0)"`,
		`stroke-opacity="0.25"`,
		`stroke-width="2.5"`,
		`stroke-linejoin="round"`,
		`stroke-linecap="square"`,
		`stroke-miterlimit="7"`,
		`stroke-dasharray="4,2,1,3"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected serialized path attribute %q in %q", want, content)
		}
	}
}

func circleMarkerPath() geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: -1, Y: 0})
	p.LineTo(geom.Pt{X: 1, Y: 0})
	p.LineTo(geom.Pt{X: 0, Y: 1})
	p.Close()
	return p
}

func TestDrawMarkersEmitsDefAndUseElements(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		ok := r.DrawMarkers(render.MarkerBatch{
			Marker: circleMarkerPath(),
			Items: []render.MarkerItem{
				{
					Offset:    geom.Pt{X: 10, Y: 20},
					Transform: geom.Identity(),
					Paint:     render.Paint{Fill: render.Color{R: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
				},
				{
					Offset:    geom.Pt{X: 50, Y: 60},
					Transform: geom.Identity(),
					Paint:     render.Paint{Fill: render.Color{B: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
				},
			},
		})
		if !ok {
			t.Fatal("DrawMarkers returned false")
		}
	})

	if !strings.Contains(content, `<path id="marker1" d="M -1 0 L 1 0 L 0 1 Z" />`) {
		t.Fatalf("expected marker path def, got %q", content)
	}
	if got := strings.Count(content, `<use href="#marker1"`); got != 2 {
		t.Fatalf("expected 2 use references, got %d in %q", got, content)
	}
	if !strings.Contains(content, `transform="matrix(1 0 0 1 10 20)"`) {
		t.Fatalf("expected first item transform, got %q", content)
	}
	if !strings.Contains(content, `transform="matrix(1 0 0 1 50 60)"`) {
		t.Fatalf("expected second item transform, got %q", content)
	}
	if !strings.Contains(content, `fill="rgb(255,0,0)"`) || !strings.Contains(content, `fill="rgb(0,0,255)"`) {
		t.Fatalf("expected per-item fills, got %q", content)
	}
}

func TestDrawMarkersDedupesIdenticalMarkerGeometry(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		batch := render.MarkerBatch{
			Marker: circleMarkerPath(),
			Items: []render.MarkerItem{
				{
					Offset: geom.Pt{X: 0, Y: 0}, Transform: geom.Identity(),
					Paint: render.Paint{Fill: render.Color{A: 1}},
				},
			},
		}
		r.DrawMarkers(batch)
		r.DrawMarkers(batch)
	})

	if got := strings.Count(content, `<path id="marker`); got != 1 {
		t.Fatalf("identical marker geometry should share one def, got %d defs in %q", got, content)
	}
	if got := strings.Count(content, `<use href="#marker1"`); got != 2 {
		t.Fatalf("expected 2 use references sharing marker1, got %d in %q", got, content)
	}
}

func TestDrawMarkersHonorsActiveClip(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipRect(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}})
		r.DrawMarkers(render.MarkerBatch{
			Marker: circleMarkerPath(),
			Items: []render.MarkerItem{
				{
					Offset: geom.Pt{X: 50, Y: 50}, Transform: geom.Identity(),
					Paint: render.Paint{Fill: render.Color{A: 1}},
				},
			},
		})
	})

	if !strings.Contains(content, `<g clip-path="url(#clip1)"><use href="#marker1"`) {
		t.Fatalf("markers should be wrapped in the active clip group, got %q", content)
	}
}

func TestDrawMarkersRejectsEmptyBatchOrInvalidMarker(t *testing.T) {
	r := mustNewRenderer(t)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 180, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	if ok := r.DrawMarkers(render.MarkerBatch{}); ok {
		t.Error("DrawMarkers should return false for empty batch")
	}
	if ok := r.DrawMarkers(render.MarkerBatch{
		Marker: circleMarkerPath(),
		Items:  nil,
	}); ok {
		t.Error("DrawMarkers should return false when Items is empty")
	}

	// Items with no fill and no stroke should produce no output but still
	// register the marker def (caller asked for a marker batch with that
	// geometry) — implementation returns true to signal it took ownership.
	ok := r.DrawMarkers(render.MarkerBatch{
		Marker: circleMarkerPath(),
		Items: []render.MarkerItem{
			{Offset: geom.Pt{}, Transform: geom.Identity(), Paint: render.Paint{}},
		},
	})
	if !ok {
		t.Error("DrawMarkers should return true even when all items are invisible")
	}

	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	content := r.renderSVG()
	if strings.Contains(content, "<use ") {
		t.Fatalf("invisible markers should not emit use elements, got %q", content)
	}
}

func TestDrawPathCollectionEmitsDefsAndUseElements(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		itemPath := circleMarkerPath()
		ok := r.DrawPathCollection(render.PathCollectionBatch{
			Items: []render.PathCollectionItem{
				{
					Path:  itemPath,
					Paint: render.Paint{Fill: render.Color{R: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
				},
				{
					Path:  itemPath,
					Paint: render.Paint{Fill: render.Color{B: 1, A: 1}, Stroke: render.Color{A: 1}, LineWidth: 1},
				},
			},
		})
		if !ok {
			t.Fatal("DrawPathCollection returned false")
		}
	})

	if !strings.Contains(content, `<path id="pathcoll1" d="M -1 0 L 1 0 L 0 1 Z" />`) {
		t.Fatalf("expected path collection path def, got %q", content)
	}
	if got := strings.Count(content, `<use href="#pathcoll1"`); got != 2 {
		t.Fatalf("expected 2 path collection use references, got %d in %q", got, content)
	}
	if !strings.Contains(content, `fill="rgb(255,0,0)"`) || !strings.Contains(content, `fill="rgb(0,0,255)"`) {
		t.Fatalf("expected per-item fills, got %q", content)
	}
}

func TestDrawPathCollectionHonorsActiveClipAndRejectsEmptyBatch(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		if ok := r.DrawPathCollection(render.PathCollectionBatch{}); ok {
			t.Fatal("DrawPathCollection should reject an empty batch")
		}
		r.ClipRect(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}})
		ok := r.DrawPathCollection(render.PathCollectionBatch{
			Items: []render.PathCollectionItem{
				{
					Path:  circleMarkerPath(),
					Paint: render.Paint{Fill: render.Color{A: 1}},
				},
			},
		})
		if !ok {
			t.Fatal("DrawPathCollection returned false")
		}
	})

	if !strings.Contains(content, `<g clip-path="url(#clip1)"><use href="#pathcoll1"`) {
		t.Fatalf("path collection should be wrapped in the active clip group, got %q", content)
	}
}

func TestPathWithHatchEmitsPatternFill(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 10, Y: 10})
		path.LineTo(geom.Pt{X: 70, Y: 10})
		path.LineTo(geom.Pt{X: 70, Y: 50})
		path.Close()
		r.Path(path, &render.Paint{
			Fill:           render.Color{G: 1, A: 0.5},
			Hatch:          "/",
			HatchColor:     render.Color{R: 1, A: 0.75},
			HatchLineWidth: 2,
			HatchSpacing:   12,
		})
	})

	for _, want := range []string{
		`<pattern id="hatch1" patternUnits="userSpaceOnUse" width="72" height="72">`,
		`<rect x="0" y="0" width="72" height="72" fill="rgb(0,255,0)" fill-opacity="0.5" />`,
		`stroke="rgb(255,0,0)"`,
		`stroke-opacity="0.75"`,
		`stroke-width="2"`,
		`fill="url(#hatch1)"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected hatch pattern fragment %q in %q", want, content)
		}
	}
}

func TestHatchPatternDefsAreReused(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 0, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 10})
		path.Close()
		paint := &render.Paint{
			Fill:           render.Color{A: 1},
			Hatch:          "x",
			HatchColor:     render.Color{A: 1},
			HatchLineWidth: 1,
		}
		r.Path(path, paint)
		r.Path(path, paint)
	})

	if got := strings.Count(content, `<pattern id="hatch`); got != 1 {
		t.Fatalf("identical hatch metadata should share one pattern def, got %d in %q", got, content)
	}
	if got := strings.Count(content, `fill="url(#hatch1)"`); got != 2 {
		t.Fatalf("both paths should reference shared hatch pattern, got %d refs in %q", got, content)
	}
}

func TestPathForcedAlphaUsesElementOpacity(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		var path geom.Path
		path.MoveTo(geom.Pt{X: 0, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 0})
		path.LineTo(geom.Pt{X: 10, Y: 10})
		path.Close()
		r.Path(path, &render.Paint{
			Fill:       render.Color{R: 1, A: 0.2},
			Stroke:     render.Color{B: 1, A: 0.2},
			LineWidth:  1,
			ForceAlpha: true,
			Alpha:      0.2,
		})
	})

	if !strings.Contains(content, `opacity="0.2"`) {
		t.Fatalf("forced-alpha path should carry element opacity, got %q", content)
	}
	if strings.Contains(content, `fill-opacity=`) || strings.Contains(content, `stroke-opacity=`) {
		t.Fatalf("forced-alpha path should not double-apply per-color opacity, got %q", content)
	}
}

func TestSupportsNativeHatch(t *testing.T) {
	r := mustNewRenderer(t)
	if !r.SupportsNativeHatch() {
		t.Fatal("SVG renderer should advertise native hatch consumption")
	}
}

func TestImageTransformedEmitsMatrixAttribute(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		img.SetRGBA(0, 0, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		// Affine: scale 1.5 in x, rotate-ish via b=0.5, translate (10, 20).
		r.ImageTransformed(render.NewImageData(img), geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 2, Y: 2},
		}, geom.Affine{A: 1.5, B: 0.5, C: -0.25, D: 1, E: 10, F: 20})
	})

	for _, want := range []string{
		`<image x="0" y="0" width="2" height="2"`,
		`href="data:image/png;base64,`,
		`transform="matrix(1.5 0.5 -0.25 1 10 20)"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected serialized transformed-image attribute %q in %q", want, content)
		}
	}
}

func TestImageTransformedHonorsClip(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipRect(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 50, Y: 50}})
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		r.ImageTransformed(render.NewImageData(img), geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 10, Y: 10},
		}, geom.Identity())
	})

	if !strings.Contains(content, `<g clip-path="url(#clip1)"><image`) {
		t.Fatalf("transformed image should respect active clip, got %q", content)
	}
}

func TestImageTransformedSkipsUnsupportedImage(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ImageTransformed(sizeOnlyImage{w: 10, h: 10}, geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 10, Y: 10},
		}, geom.Identity())
	})

	if strings.Contains(content, "<image") {
		t.Fatalf("unsupported image should not emit image node, got %q", content)
	}
}

func TestMatrixTransformFormat(t *testing.T) {
	got := matrixTransform(geom.Affine{A: 1, B: 0, C: 0, D: 1, E: 0, F: 0})
	if got != "matrix(1 0 0 1 0 0)" {
		t.Errorf("identity matrix transform = %q, want %q", got, "matrix(1 0 0 1 0 0)")
	}

	got = matrixTransform(geom.Affine{A: 0.5, B: 0.25, C: -0.5, D: 1.25, E: 10, F: -20})
	if got != "matrix(0.5 0.25 -0.5 1.25 10 -20)" {
		t.Errorf("matrix transform = %q, want compact form", got)
	}
}

func TestImageSerializesEmbeddedPNGAndNormalizesDestinationRect(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		img.SetRGBA(0, 0, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		r.Image(render.NewImageData(img), geom.Rect{
			Min: geom.Pt{X: 30, Y: 40},
			Max: geom.Pt{X: 10, Y: 20},
		})
	})

	for _, want := range []string{
		`<image x="10" y="20" width="20" height="20"`,
		`preserveAspectRatio="none"`,
		`href="data:image/png;base64,`,
		`xlink:href="data:image/png;base64,`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected serialized image attribute %q in %q", want, content)
		}
	}
}

func TestTextEscapingAndRotationSerialization(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.DrawTextRotated(`A<&"B`, geom.Pt{X: 90, Y: 70}, 12, 0.5, render.Color{R: 0.2, G: 0.4, B: 0.6, A: 0.75})
	})

	for _, want := range []string{
		`transform="matrix(0.877583 -0.479426 0.479426 0.877583 -22.542218 51.717519)"`,
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

func TestClipPathNestsInsideActiveRectClip(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipRect(geom.Rect{
			Min: geom.Pt{X: 10, Y: 15},
			Max: geom.Pt{X: 70, Y: 75},
		})

		var clipPath geom.Path
		clipPath.MoveTo(geom.Pt{X: 0, Y: 0})
		clipPath.LineTo(geom.Pt{X: 30, Y: 0})
		clipPath.LineTo(geom.Pt{X: 30, Y: 30})
		clipPath.Close()
		r.ClipPath(clipPath)

		r.DrawText("clipped", geom.Pt{X: 20, Y: 30}, 12, render.Color{A: 1})
	})

	if got := strings.Count(content, "<clipPath"); got != 2 {
		t.Fatalf("expected two clip defs (rect + path), got %d in %q", got, content)
	}
	if !strings.Contains(content, `<rect x="10" y="15" width="60" height="60" />`) {
		t.Fatalf("expected rect clip def, got %q", content)
	}
	if !strings.Contains(content, `<path d="M 0 0 L 30 0 L 30 30 Z" />`) {
		t.Fatalf("expected path clip def, got %q", content)
	}

	// Rect clip wraps the outer <g>; path clip nests inside.
	re := regexp.MustCompile(`<g clip-path="url\(#(clip\d+)\)"><g clip-path="url\(#(clip\d+)\)"><text[^>]*>clipped</text></g></g>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) != 3 {
		t.Fatalf("expected rect-clip outer group wrapping path-clip inner group, got %q", content)
	}
	if matches[1] == matches[2] {
		t.Fatalf("nested groups should reference distinct clip IDs, got %q", content)
	}
}

func TestClipPathDedupesIdenticalPathsAcrossNodes(t *testing.T) {
	makePath := func() geom.Path {
		var p geom.Path
		p.MoveTo(geom.Pt{X: 0, Y: 0})
		p.LineTo(geom.Pt{X: 40, Y: 0})
		p.LineTo(geom.Pt{X: 40, Y: 40})
		p.Close()
		return p
	}

	content := renderSVGDocument(t, func(r *Renderer) {
		r.Save()
		r.ClipPath(makePath())
		r.DrawText("first", geom.Pt{X: 10, Y: 10}, 12, render.Color{A: 1})
		r.Restore()

		r.Save()
		r.ClipPath(makePath())
		r.DrawText("second", geom.Pt{X: 10, Y: 30}, 12, render.Color{A: 1})
		r.Restore()
	})

	if got := strings.Count(content, "<clipPath"); got != 1 {
		t.Fatalf("identical path clips should share one def, got %d in %q", got, content)
	}
	if got := strings.Count(content, `clip-path="url(#clip1)"`); got != 2 {
		t.Fatalf("both nodes should reference the shared clip def, got %d refs in %q", got, content)
	}
}

func TestClipPathStackUnwindsOnRestore(t *testing.T) {
	makePath := func(offset float64) geom.Path {
		var p geom.Path
		p.MoveTo(geom.Pt{X: offset, Y: offset})
		p.LineTo(geom.Pt{X: offset + 20, Y: offset})
		p.LineTo(geom.Pt{X: offset + 20, Y: offset + 20})
		p.Close()
		return p
	}

	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipPath(makePath(0))
		r.Save()
		r.ClipPath(makePath(50))
		r.DrawText("nested", geom.Pt{X: 5, Y: 5}, 12, render.Color{A: 1})
		r.Restore()
		r.DrawText("popped", geom.Pt{X: 5, Y: 30}, 12, render.Color{A: 1})
	})

	// "nested" sees both clips; "popped" sees only the outer clip after Restore.
	nestedRe := regexp.MustCompile(`<g clip-path="url\(#clip1\)"><g clip-path="url\(#clip2\)"><text[^>]*>nested</text></g></g>`)
	if !nestedRe.MatchString(content) {
		t.Fatalf("expected nested element wrapped in both clips, got %q", content)
	}
	poppedRe := regexp.MustCompile(`<g clip-path="url\(#clip1\)"><text[^>]*>popped</text></g>`)
	if !poppedRe.MatchString(content) {
		t.Fatalf("expected popped element wrapped in only the outer clip, got %q", content)
	}
	if strings.Contains(content, `<g clip-path="url(#clip2)"><text x="5" y="30"`) {
		t.Fatalf("inner clip leaked past Restore, got %q", content)
	}
}

func TestClipPathRejectsEmptyPath(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipPath(geom.Path{}) // empty path → no clip, no def
		r.DrawText("unclipped", geom.Pt{X: 5, Y: 5}, 12, render.Color{A: 1})
	})

	if strings.Contains(content, "<clipPath") {
		t.Fatalf("empty path should not register clip defs, got %q", content)
	}
	if strings.Contains(content, "clip-path=") {
		t.Fatalf("empty path should not wrap content in clip groups, got %q", content)
	}
}

func TestClipPathTransformedEmitsPathTransformAndDedupesByTransform(t *testing.T) {
	makePath := func() geom.Path {
		var p geom.Path
		p.MoveTo(geom.Pt{X: 0, Y: 0})
		p.LineTo(geom.Pt{X: 20, Y: 0})
		p.LineTo(geom.Pt{X: 20, Y: 20})
		p.Close()
		return p
	}
	shift := geom.Affine{A: 1, D: 1, E: 10, F: 15}
	scale := geom.Affine{A: 2, D: 2}

	content := renderSVGDocument(t, func(r *Renderer) {
		r.Save()
		r.ClipPathTransformed(makePath(), shift)
		r.DrawText("first", geom.Pt{X: 5, Y: 5}, 12, render.Color{A: 1})
		r.Restore()

		r.Save()
		r.ClipPathTransformed(makePath(), shift)
		r.DrawText("second", geom.Pt{X: 5, Y: 20}, 12, render.Color{A: 1})
		r.Restore()

		r.Save()
		r.ClipPathTransformed(makePath(), scale)
		r.DrawText("third", geom.Pt{X: 5, Y: 35}, 12, render.Color{A: 1})
		r.Restore()
	})

	if got := strings.Count(content, "<clipPath"); got != 2 {
		t.Fatalf("path clips should dedupe by path plus transform, got %d defs in %q", got, content)
	}
	for _, want := range []string{
		`<path d="M 0 0 L 20 0 L 20 20 Z" transform="matrix(1 0 0 1 10 15)" />`,
		`<path d="M 0 0 L 20 0 L 20 20 Z" transform="matrix(2 0 0 2 0 0)" />`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected transformed clip def %q in %q", want, content)
		}
	}
	if got := strings.Count(content, `clip-path="url(#clip1)"`); got != 2 {
		t.Fatalf("first transform should share clip1 across two nodes, got %d refs in %q", got, content)
	}
	if got := strings.Count(content, `clip-path="url(#clip2)"`); got != 1 {
		t.Fatalf("second transform should use clip2 once, got %d refs in %q", got, content)
	}
}

func TestSetResolutionIgnoresZeroAndStoresPositiveDPI(t *testing.T) {
	r := mustNewRenderer(t)

	r.SetResolution(0)
	if got := r.resolution; got != 72 {
		t.Fatalf("zero DPI should be ignored, got %d", got)
	}

	r.SetResolution(144)
	if got := r.resolution; got != 144 {
		t.Fatalf("positive DPI should be stored, got %d", got)
	}
}

func TestBuildPathDataSupportsQuadraticCubicAndClose(t *testing.T) {
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.QuadTo, geom.CubicTo, geom.ClosePath},
		V: []geom.Pt{
			{X: 1, Y: 2},
			{X: 3, Y: 4},
			{X: 5, Y: 6},
			{X: 7, Y: 8},
			{X: 9, Y: 10},
			{X: 11, Y: 12},
		},
	}

	got := buildPathData(path)
	want := "M 1 2 Q 3 4 5 6 C 7 8 9 10 11 12 Z"
	if got != want {
		t.Fatalf("unexpected path data:\nwant %q\ngot  %q", want, got)
	}
}

func TestBuildPathDataRejectsMalformedCommands(t *testing.T) {
	tests := []geom.Path{
		{C: []geom.Cmd{geom.MoveTo}},
		{C: []geom.Cmd{geom.LineTo}},
		{C: []geom.Cmd{geom.QuadTo}, V: []geom.Pt{{X: 1, Y: 2}}},
		{C: []geom.Cmd{geom.CubicTo}, V: []geom.Pt{{X: 1, Y: 2}, {X: 3, Y: 4}}},
		{C: []geom.Cmd{geom.Cmd(99)}, V: []geom.Pt{{X: 1, Y: 2}}},
	}

	for _, path := range tests {
		if got := buildPathData(path); got != "" {
			t.Fatalf("malformed path should serialize to empty data, got %q for %+v", got, path)
		}
	}
}

func TestHelperFormattingBranches(t *testing.T) {
	if got := dashedArray([]float64{5}); got != "" {
		t.Fatalf("single dash segment should not emit dash array, got %q", got)
	}
	if got := dashedArray([]float64{5, 2, 9}); got != "5,2" {
		t.Fatalf("odd dash lists should ignore trailing value, got %q", got)
	}

	if got := mapLineJoin(render.JoinBevel); got != "bevel" {
		t.Fatalf("expected bevel join mapping, got %q", got)
	}
	if got := mapLineJoin(render.LineJoin(99)); got != "miter" {
		t.Fatalf("unknown join should fall back to miter, got %q", got)
	}
	if got := mapLineCap(render.CapRound); got != "round" {
		t.Fatalf("expected round cap mapping, got %q", got)
	}
	if got := mapLineCap(render.LineCap(99)); got != "butt" {
		t.Fatalf("unknown cap should fall back to butt, got %q", got)
	}

	if got := clamp01(-0.1); got != 0 {
		t.Fatalf("negative colors should clamp to 0, got %v", got)
	}
	if got := clamp01(1.1); got != 1 {
		t.Fatalf("oversaturated colors should clamp to 1, got %v", got)
	}
	if got := clampFloat(1.25); got != 1.25 {
		t.Fatalf("finite float should pass through, got %v", got)
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
	if !strings.Contains(content, `<text x="15" y="20"`) {
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

func TestImageSkipsUnsupportedImageAndDegenerateRect(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.Image(sizeOnlyImage{w: 10, h: 10}, geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 10, Y: 10},
		})
		r.Image(render.NewImageData(image.NewRGBA(image.Rect(0, 0, 1, 1))), geom.Rect{
			Min: geom.Pt{X: 10, Y: 10},
			Max: geom.Pt{X: 10, Y: 20},
		})
	})

	if strings.Contains(content, "<image") {
		t.Fatalf("unsupported images and degenerate rects should not emit image nodes, got %q", content)
	}
}

func TestSaveSVGErrorPaths(t *testing.T) {
	r := mustNewRenderer(t)

	if err := r.SaveSVG(""); err == nil {
		t.Fatal("empty output path should return an error")
	}

	dir := t.TempDir()
	if err := r.SaveSVG(filepath.Join(dir, ".")); err == nil {
		t.Fatal("writing SVG to a directory should return an error")
	}
}

func TestHelperImageBranches(t *testing.T) {
	if _, err := encodeImage(nil); err == nil {
		t.Fatal("encoding nil image should fail")
	}
	if got := asRGBAImage(sizeOnlyImage{w: 3, h: 4}); got != nil {
		t.Fatalf("non-RGBA image should not convert, got %#v", got)
	}

	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	converted := asRGBAImage(render.NewImageData(img))
	if converted == nil || converted.Bounds() != img.Bounds() {
		t.Fatalf("expected RGBA image conversion to preserve bounds, got %#v", converted)
	}
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

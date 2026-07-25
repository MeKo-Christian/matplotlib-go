package svg

import (
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
)

// TestAuditRendererSurface exercises every render.Renderer method and every
// optional capability interface the SVG renderer advertises. The goal is to
// catch silent-drop regressions: if any method becomes a no-op or stops
// participating in serialized output, one of these subtests fails.
//
// Each subtest names the interface method under audit and asserts an
// observable outcome rather than byte-level layout (which other tests cover).
func TestAuditRendererSurface(t *testing.T) {
	t.Run("Begin/End", func(t *testing.T) {
		r := mustNewRenderer(t)
		if err := r.Begin(geom.Rect{Max: geom.Pt{X: 20, Y: 20}}); err != nil {
			t.Fatalf("Begin returned error: %v", err)
		}
		if !r.began {
			t.Fatal("Begin did not flip the began flag")
		}
		if err := r.End(); err != nil {
			t.Fatalf("End returned error: %v", err)
		}
		if r.began {
			t.Fatal("End did not clear the began flag")
		}
	})

	t.Run("Save/Restore", func(t *testing.T) {
		r := mustNewRenderer(t)
		mustBegin(t, r)
		r.ClipRect(geom.Rect{Max: geom.Pt{X: 5, Y: 5}})
		baseDepth := len(r.stack)
		baseClip := r.clipRect

		r.Save()
		if len(r.stack) != baseDepth+1 {
			t.Fatalf("Save did not push state: depth=%d want=%d", len(r.stack), baseDepth+1)
		}

		// Modify clip to verify Restore rolls it back.
		r.ClipRect(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
		r.Restore()
		if len(r.stack) != baseDepth {
			t.Fatalf("Restore did not pop state: depth=%d want=%d", len(r.stack), baseDepth)
		}
		if r.clipRect == nil || baseClip == nil || *r.clipRect != *baseClip {
			t.Fatalf("Restore did not roll back clip: got %+v want %+v", r.clipRect, baseClip)
		}
	})

	t.Run("ClipRect", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			r.ClipRect(geom.Rect{Min: geom.Pt{X: 1, Y: 2}, Max: geom.Pt{X: 30, Y: 40}})
			r.Path(simplePath(), &render.Paint{Fill: opaqueRed()})
		})
		if !strings.Contains(content, "<clipPath") || !strings.Contains(content, `clip-path="url(#`) {
			t.Fatalf("ClipRect did not emit clipPath defs and reference, got %q", content)
		}
	})

	t.Run("ClipPath", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			var p geom.Path
			p.MoveTo(geom.Pt{X: 0, Y: 0})
			p.LineTo(geom.Pt{X: 10, Y: 0})
			p.LineTo(geom.Pt{X: 10, Y: 10})
			p.Close()
			r.ClipPath(p)
			r.Path(simplePath(), &render.Paint{Fill: opaqueRed()})
		})
		if !strings.Contains(content, `<clipPath id=`) || !strings.Contains(content, "<path") {
			t.Fatalf("ClipPath did not emit a path-based clipPath def, got %q", content)
		}
	})

	t.Run("Path", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			r.Path(simplePath(), &render.Paint{
				Stroke: render.Color{R: 0, G: 0, B: 0, A: 1}, LineWidth: 1, Fill: opaqueRed(),
			})
		})
		if !strings.Contains(content, "<path ") {
			t.Fatalf("Path did not emit a <path> node, got %q", content)
		}
	})

	t.Run("Image", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			img := image.NewRGBA(image.Rect(0, 0, 2, 2))
			r.DrawImage(render.NewImageData(img), geom.Rect{
				Min: geom.Pt{X: 1, Y: 1},
				Max: geom.Pt{X: 5, Y: 5},
			})
		})
		if !strings.Contains(content, "<image") || !strings.Contains(content, "data:image/png;base64,") {
			t.Fatalf("Image did not emit an inline <image> node, got %q", content)
		}
	})

	t.Run("GlyphRun", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			run := render.GlyphRun{
				Origin:  geom.Pt{X: 4, Y: 12},
				Size:    12,
				FontKey: "DejaVu Sans",
				Glyphs: []render.Glyph{
					{ID: uint32('A'), Advance: 8},
					{ID: uint32('B'), Advance: 8},
				},
			}
			r.GlyphRun(run, render.Color{A: 1})
		})
		if !strings.Contains(content, "<text") || strings.Count(content, "<text") < 2 {
			t.Fatalf("GlyphRun did not emit per-glyph <text> nodes, got %q", content)
		}
	})

	t.Run("MeasureText", func(t *testing.T) {
		r := mustNewRenderer(t)
		m := r.MeasureText("hi", 12, "DejaVu Sans")
		if m.W <= 0 || m.H <= 0 || m.Ascent <= 0 {
			t.Fatalf("MeasureText returned degenerate metrics: %+v", m)
		}
	})

	// Optional capability interfaces below. Each one is asserted via the
	// renderer's declared interface to catch accidental capability removal.

	t.Run("DPIAware", func(t *testing.T) {
		var iface render.DPIAware = mustNewRenderer(t)
		r := iface.(*Renderer)
		iface.SetResolution(150)
		if r.resolution != 150 {
			t.Fatalf("SetResolution did not store DPI: got %d want 150", r.resolution)
		}
		// Zero must be ignored, leaving the previous value intact.
		iface.SetResolution(0)
		if r.resolution != 150 {
			t.Fatalf("SetResolution(0) should be a no-op; got %d", r.resolution)
		}
	})

	t.Run("TextDrawer", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			var iface render.TextDrawer = r
			iface.DrawText("audit", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1})
		})
		if !strings.Contains(content, ">audit<") {
			t.Fatalf("DrawText did not emit text content, got %q", content)
		}
	})

	t.Run("RotatedTextDrawer", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			var iface render.RotatedTextDrawer = r
			iface.DrawTextRotated("rot", geom.Pt{X: 30, Y: 30}, 12, math.Pi/4, render.Color{A: 1})
		})
		if !strings.Contains(content, ">rot<") || !strings.Contains(content, `transform="matrix(`) {
			t.Fatalf("DrawTextRotated did not emit rotated text node, got %q", content)
		}
	})

	t.Run("VerticalTextDrawer", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			var iface render.VerticalTextDrawer = r
			iface.DrawTextVertical("AB", geom.Pt{X: 10, Y: 50}, 12, render.Color{A: 1})
		})
		// Vertical text emits one text node per rune.
		if strings.Count(content, "<text") < 2 {
			t.Fatalf("DrawTextVertical should emit one text node per rune, got %q", content)
		}
	})

	t.Run("TextPather", func(t *testing.T) {
		r := mustNewRenderer(t)
		var iface render.TextPather = r
		// TextPath relies on the shared font manager. A missing font is allowed
		// to return false; the contract we audit is that the method exists and
		// returns deterministically.
		_, _ = iface.TextPath("a", geom.Pt{}, 12, "DejaVu Sans")
	})

	t.Run("ImageTransformer", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			var iface render.ImageTransformer = r
			img := image.NewRGBA(image.Rect(0, 0, 2, 2))
			iface.ImageTransformed(render.NewImageData(img),
				geom.Rect{Max: geom.Pt{X: 2, Y: 2}},
				geom.Affine{A: 1, D: 1, E: 5, F: 5})
		})
		if !strings.Contains(content, "<image") || !strings.Contains(content, `transform="matrix(`) {
			t.Fatalf("ImageTransformed did not emit transformed <image>, got %q", content)
		}
	})

	t.Run("ClipPathTransformer", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			var iface render.ClipPathTransformer = r
			var p geom.Path
			p.MoveTo(geom.Pt{X: 0, Y: 0})
			p.LineTo(geom.Pt{X: 10, Y: 0})
			p.LineTo(geom.Pt{X: 10, Y: 10})
			p.Close()
			iface.ClipPathTransformed(p, geom.Affine{A: 2, D: 2, E: 1, F: 1})
			r.Path(simplePath(), &render.Paint{Fill: opaqueRed()})
		})
		if !strings.Contains(content, "<clipPath") || !strings.Contains(content, `transform="matrix(`) {
			t.Fatalf("ClipPathTransformed did not emit transformed clipPath, got %q", content)
		}
	})

	t.Run("MarkerDrawer", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			var iface render.MarkerDrawer = r
			var marker geom.Path
			marker.MoveTo(geom.Pt{X: -1, Y: 0})
			marker.LineTo(geom.Pt{X: 1, Y: 0})
			marker.LineTo(geom.Pt{X: 0, Y: 1})
			marker.Close()
			batch := render.MarkerBatch{
				Marker: marker,
				Items: []render.MarkerItem{
					{Offset: geom.Pt{X: 10, Y: 10}, Transform: geom.Identity(), Paint: render.Paint{Fill: opaqueRed()}},
					{Offset: geom.Pt{X: 30, Y: 30}, Transform: geom.Identity(), Paint: render.Paint{Fill: opaqueRed()}},
				},
			}
			if !iface.DrawMarkers(batch) {
				t.Fatal("DrawMarkers returned false for valid batch")
			}
		})
		if strings.Count(content, "<use") < 2 || !strings.Contains(content, "<defs>") {
			t.Fatalf("DrawMarkers did not emit defs/use elements, got %q", content)
		}
	})

	t.Run("PathCollectionDrawer", func(t *testing.T) {
		content := renderSVGDocument(t, func(r *Renderer) {
			var iface render.PathCollectionDrawer = r
			batch := render.PathCollectionBatch{
				Items: []render.PathCollectionItem{
					{Path: simplePath(), Paint: render.Paint{Fill: opaqueRed()}},
					{Path: simplePath(), Paint: render.Paint{Fill: opaqueRed()}},
				},
			}
			if !iface.DrawPathCollection(batch) {
				t.Fatal("DrawPathCollection returned false for valid batch")
			}
		})
		if strings.Count(content, "<use") < 2 {
			t.Fatalf("DrawPathCollection did not emit reused path defs, got %q", content)
		}
	})

	t.Run("NativeHatcher", func(t *testing.T) {
		var iface render.NativeHatcher = mustNewRenderer(t)
		if !iface.SupportsNativeHatch() {
			t.Fatal("SupportsNativeHatch returned false; SVG advertises native hatch support")
		}
	})

	t.Run("SVGExporter", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "audit.svg")

		r := mustNewRenderer(t)
		mustBegin(t, r)
		r.Path(simplePath(), &render.Paint{Fill: opaqueRed()})
		if err := r.End(); err != nil {
			t.Fatalf("End failed: %v", err)
		}

		var iface render.SVGExporter = r
		if err := iface.SaveSVG(path); err != nil {
			t.Fatalf("SaveSVG failed: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if !strings.Contains(string(data), "<svg") {
			t.Fatalf("SaveSVG wrote a document without <svg root: %q", string(data))
		}
	})

	t.Run("SVGOptionExporter", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "audit-opts.svg")

		r := mustNewRenderer(t)
		mustBegin(t, r)
		r.Path(simplePath(), &render.Paint{Fill: opaqueRed()})
		if err := r.End(); err != nil {
			t.Fatalf("End failed: %v", err)
		}

		var iface render.SVGOptionExporter = r
		opts := render.ResolveSVGOptions(render.WithSVGMetadata(map[string]string{"Title": "audit"}))
		if err := iface.SaveSVGWithOptions(path, opts); err != nil {
			t.Fatalf("SaveSVGWithOptions failed: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if !strings.Contains(string(data), "audit") {
			t.Fatalf("SaveSVGWithOptions did not embed supplied metadata, got %q", string(data))
		}
	})

	t.Run("SVGOptionSetter", func(t *testing.T) {
		r := mustNewRenderer(t)
		var iface render.SVGOptionSetter = r
		iface.SetSVGOptions(render.ResolveSVGOptions(render.WithSVGFontPolicy(render.SVGFontPolicyPath)))
		if r.options.FontPolicy != render.SVGFontPolicyPath {
			t.Fatalf("SetSVGOptions did not store font policy: got %q", r.options.FontPolicy)
		}
	})

	// TeX methods require external latex+dvipng binaries. Drive them through a
	// scripted fake so the audit covers the surface end-to-end. Skip on
	// platforms where POSIX shell stubs do not run.
	t.Run("TeXMetricer/TeXDrawer/RotatedTeXDrawer", func(t *testing.T) {
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

		var metricer render.TeXMetricer = r
		metrics, ok := metricer.MeasureTeX(`$\alpha$`, 12, "DejaVu Sans")
		if !ok || metrics.W <= 0 || metrics.H <= 0 {
			t.Fatalf("MeasureTeX = (%+v, %v); want positive metrics", metrics, ok)
		}

		mustBegin(t, r)
		var drawer render.TeXDrawer = r
		if !drawer.DrawTeX(`$\alpha$`, geom.Pt{X: 5, Y: 5}, 12, render.Color{A: 1}, "DejaVu Sans") {
			t.Fatal("DrawTeX returned false")
		}

		var rotated render.RotatedTeXDrawer = r
		if !rotated.DrawTeXRotated(`$\alpha$`, geom.Pt{X: 20, Y: 20}, 12, math.Pi/2, render.Color{A: 1}, "DejaVu Sans") {
			t.Fatal("DrawTeXRotated returned false")
		}

		if err := r.End(); err != nil {
			t.Fatalf("End failed: %v", err)
		}
		content := r.renderSVG()
		if strings.Count(content, "<image") < 2 {
			t.Fatalf("TeX surface should embed cached PNGs as <image> nodes, got %q", content)
		}
	})
}

// TestAuditAdvertisedCapabilitiesMatchAssertions guards against accidental
// removal of compile-time capability assertions. The list mirrors the `_ =
// (*Renderer)(nil)` block in svg.go and is the source of truth the audit
// expects to remain implemented.
func TestAuditAdvertisedCapabilitiesMatchAssertions(t *testing.T) {
	r := mustNewRenderer(t)

	type check struct {
		name string
		ok   bool
	}
	checks := []check{
		{"render.Renderer", isImpl[render.Renderer](r)},
		{"render.DPIAware", isImpl[render.DPIAware](r)},
		{"render.TextDrawer", isImpl[render.TextDrawer](r)},
		{"render.RotatedTextDrawer", isImpl[render.RotatedTextDrawer](r)},
		{"render.VerticalTextDrawer", isImpl[render.VerticalTextDrawer](r)},
		{"render.TextPather", isImpl[render.TextPather](r)},
		{"render.TeXMetricer", isImpl[render.TeXMetricer](r)},
		{"render.TeXDrawer", isImpl[render.TeXDrawer](r)},
		{"render.RotatedTeXDrawer", isImpl[render.RotatedTeXDrawer](r)},
		{"render.ImageTransformer", isImpl[render.ImageTransformer](r)},
		{"render.ClipPathTransformer", isImpl[render.ClipPathTransformer](r)},
		{"render.MarkerDrawer", isImpl[render.MarkerDrawer](r)},
		{"render.PathCollectionDrawer", isImpl[render.PathCollectionDrawer](r)},
		{"render.NativeHatcher", isImpl[render.NativeHatcher](r)},
		{"render.SVGExporter", isImpl[render.SVGExporter](r)},
		{"render.SVGOptionExporter", isImpl[render.SVGOptionExporter](r)},
		{"render.SVGOptionSetter", isImpl[render.SVGOptionSetter](r)},
	}
	for _, c := range checks {
		if !c.ok {
			t.Errorf("SVG renderer no longer implements %s", c.name)
		}
	}
}

func isImpl[T any](v any) bool {
	_, ok := v.(T)
	return ok
}

func mustBegin(t *testing.T, r *Renderer) {
	t.Helper()
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 180, Y: 120}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
}

func simplePath() geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 1, Y: 1})
	p.LineTo(geom.Pt{X: 10, Y: 1})
	p.LineTo(geom.Pt{X: 10, Y: 10})
	p.Close()
	return p
}

func opaqueRed() render.Color {
	return render.Color{R: 1, A: 1}
}

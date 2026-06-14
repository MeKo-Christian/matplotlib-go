package agg

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestTeXManagerBuildsAndCachesPNG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell commands are POSIX-only")
	}
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture.png")
	writeTestPNG(t, fixture, color.RGBA{A: 255})

	latexLog := filepath.Join(dir, "latex.log")
	dvipngLog := filepath.Join(dir, "dvipng.log")
	latex := writeFakeCommand(t, dir, "latex", `#!/bin/sh
echo latex >> "$LATEX_LOG"
test -f file.tex || exit 7
touch file.dvi
`)
	dvipng := writeFakeCommand(t, dir, "dvipng", `#!/bin/sh
echo dvipng >> "$DVIPNG_LOG"
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
	t.Setenv("LATEX_LOG", latexLog)
	t.Setenv("DVIPNG_LOG", dvipngLog)
	t.Setenv("FAKE_TEX_PNG", fixture)

	manager := tex.NewManager(tex.ManagerConfig{
		CacheDir:      filepath.Join(dir, "cache"),
		LaTeXCommand:  latex,
		DVIPNGCommand: dvipng,
	})

	first, err := manager.Render(`signal $\alpha$`, 12, 144, "DejaVu Sans")
	if err != nil {
		t.Fatalf("first render: %v", err)
	}
	second, err := manager.Render(`signal $\alpha$`, 12, 144, "DejaVu Sans")
	if err != nil {
		t.Fatalf("second render: %v", err)
	}
	if first.PNGPath != second.PNGPath {
		t.Fatalf("cache path changed between renders: %q != %q", first.PNGPath, second.PNGPath)
	}
	if first.Metrics.W != 2 || first.Metrics.H != 2 || first.Metrics.Ascent != 2 {
		t.Fatalf("metrics from cached PNG = %+v, want width/height/ascent of fixture", first.Metrics)
	}
	if got := lineCount(t, latexLog); got != 1 {
		t.Fatalf("latex command count = %d, want 1", got)
	}
	if got := lineCount(t, dvipngLog); got != 1 {
		t.Fatalf("dvipng command count = %d, want 1", got)
	}
}

func TestTeXManagerReportsSubprocessFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell commands are POSIX-only")
	}
	dir := t.TempDir()
	latex := writeFakeCommand(t, dir, "latex-fail", `#!/bin/sh
echo "latex exploded"
exit 42
`)

	manager := tex.NewManager(tex.ManagerConfig{
		CacheDir:      filepath.Join(dir, "cache"),
		LaTeXCommand:  latex,
		DVIPNGCommand: "unused",
	})

	_, err := manager.Render(`bad $\tex$`, 12, 100, "DejaVu Sans")
	if err == nil {
		t.Fatal("expected render error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "latex-fail") || !strings.Contains(msg, "bad") || !strings.Contains(msg, "latex exploded") {
		t.Fatalf("error message missing command, TeX text, or output: %v", err)
	}
}

func TestAggDrawTeXCompositesCachedPNG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell commands are POSIX-only")
	}
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture.png")
	writeTestPNG(t, fixture, color.RGBA{A: 255})
	latex := writeFakeCommand(t, dir, "latex", `#!/bin/sh
touch file.dvi
`)
	dvipng := writeFakeCommand(t, dir, "dvipng", `#!/bin/sh
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

	r, err := New(24, 24, render.Color{})
	if err != nil {
		t.Fatalf("New renderer: %v", err)
	}
	r.texManager = tex.NewManager(tex.ManagerConfig{
		CacheDir:      filepath.Join(dir, "cache"),
		LaTeXCommand:  latex,
		DVIPNGCommand: dvipng,
	})

	if !r.DrawTeX(`x`, geom.Pt{X: 8, Y: 10}, 12, render.Color{R: 1, A: 1}, "DejaVu Sans") {
		t.Fatal("DrawTeX returned false")
	}
	// Display space is y-up: baseline display y=10 maps to device y=H-10=14, so
	// the 2x2 image ascends into device rows 12-13.
	got := r.GetImage().RGBAAt(8, 12)
	if got.R < 200 || got.G != 0 || got.B != 0 || got.A < 200 {
		t.Fatalf("DrawTeX pixel = %+v, want opaque red text image", got)
	}
}

func TestAggDrawTeXAppliesAlphaAndClip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell commands are POSIX-only")
	}
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture.png")
	writeSizedTestPNG(t, fixture, 4, 4, color.RGBA{A: 255})
	latex := writeFakeCommand(t, dir, "latex", `#!/bin/sh
touch file.dvi
`)
	dvipng := writeFakeCommand(t, dir, "dvipng", `#!/bin/sh
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

	r, err := New(24, 24, render.Color{})
	if err != nil {
		t.Fatalf("New renderer: %v", err)
	}
	r.texManager = tex.NewManager(tex.ManagerConfig{
		CacheDir:      filepath.Join(dir, "cache"),
		LaTeXCommand:  latex,
		DVIPNGCommand: dvipng,
	})
	// Display space is y-up: baseline display y=12 maps to device y=H-12=12, so
	// the 4x4 image ascends into device rows 8-11 (cols 8-11). The clip rect in
	// display coords {(8,14)-(10,16)} device-flips to device cols[8,10] rows[8,10],
	// keeping the image's top-left while masking the right column.
	r.ClipRect(geom.Rect{Min: geom.Pt{X: 8, Y: 14}, Max: geom.Pt{X: 10, Y: 16}})

	if !r.DrawTeX(`x`, geom.Pt{X: 8, Y: 12}, 12, render.Color{G: 1, A: 0.5}, "DejaVu Sans") {
		t.Fatal("DrawTeX returned false")
	}
	inside := r.GetImage().RGBAAt(8, 8)
	if inside.G < 120 || inside.A < 120 || inside.A > 136 || inside.R != 0 || inside.B != 0 {
		t.Fatalf("clipped alpha TeX pixel = %+v, want half-alpha green", inside)
	}
	outside := r.GetImage().RGBAAt(11, 8)
	if outside != (color.RGBA{}) {
		t.Fatalf("TeX draw escaped clip: outside pixel = %+v", outside)
	}
}

func TestAggDrawTeXRotatedCompositesCachedPNG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell commands are POSIX-only")
	}
	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture.png")
	writeTestPNG(t, fixture, color.RGBA{A: 255})
	latex := writeFakeCommand(t, dir, "latex", `#!/bin/sh
touch file.dvi
`)
	dvipng := writeFakeCommand(t, dir, "dvipng", `#!/bin/sh
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

	r, err := New(32, 32, render.Color{})
	if err != nil {
		t.Fatalf("New renderer: %v", err)
	}
	r.texManager = tex.NewManager(tex.ManagerConfig{
		CacheDir:      filepath.Join(dir, "cache"),
		LaTeXCommand:  latex,
		DVIPNGCommand: dvipng,
	})

	if !r.DrawTeXRotated(`x`, geom.Pt{X: 16, Y: 16}, 12, 0.4, render.Color{B: 1, A: 1}, "DejaVu Sans") {
		t.Fatal("DrawTeXRotated returned false")
	}
	if got := nonZeroPixelCount(r.GetImage()); got == 0 {
		t.Fatal("DrawTeXRotated produced a blank image")
	}
}

func TestTeXSourcePreservesBoxesRulesAndFontFamily(t *testing.T) {
	source := tex.Source(`\fbox{\rule{1em}{0.4pt}}`, 12, "DejaVu Sans")
	for _, want := range []string{
		`\sffamily \fbox{\rule{1em}{0.4pt}}`,
		`\@ifpackageloaded{underscore}`,
		`\@ifpackageloaded{textcomp}`,
		`\fontsize{12}{15}%`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("TeX source missing %q:\n%s", want, source)
		}
	}
}

func writeFakeCommand(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake command: %v", err)
	}
	return path
}

func writeTestPNG(t *testing.T, path string, c color.RGBA) {
	t.Helper()
	writeSizedTestPNG(t, path, 2, 2, c)
}

func writeSizedTestPNG(t *testing.T, path string, w, h int, c color.RGBA) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
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

func lineCount(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	return bytes.Count(data, []byte{'\n'})
}

func nonZeroPixelCount(img *image.RGBA) int {
	if img == nil {
		return 0
	}
	n := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			if img.RGBAAt(x, y).A != 0 {
				n++
			}
		}
	}
	return n
}

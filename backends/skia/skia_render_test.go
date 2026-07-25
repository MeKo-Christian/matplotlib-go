//go:build skia

// External skia_test package: this file imports test/parity (→ backends/all →
// backends/skia), which would be an import cycle from an internal test. See the
// note in parity_test.go.
package skia_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/backends/skia"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

func TestSkiaTaggedRendererImplementsCPUBaseContract(t *testing.T) {
	r, err := skia.New(backends.Config{
		Width:      32,
		Height:     24,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		Options:    backends.SkiaConfig{UseGPU: false, SampleCount: 1, ColorType: "RGBA8888"},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if r.GPU() {
		t.Fatal("CPU config should not enable GPU mode")
	}
	if r.SampleCount() != 1 {
		t.Fatalf("SampleCount() = %d, want 1", r.SampleCount())
	}

	viewport := geom.Rect{Max: geom.Pt{X: 32, Y: 24}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	r.Save()
	r.ClipRect(geom.Rect{Min: geom.Pt{X: 4, Y: 4}, Max: geom.Pt{X: 28, Y: 20}})
	r.Path(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 32, Y: 0}, {X: 32, Y: 24}, {X: 0, Y: 24}},
	}, &render.Paint{
		Fill: render.Color{R: 0.1, G: 0.2, B: 0.8, A: 1},
	})
	r.Restore()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	img.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	img.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, A: 255})
	r.DrawImage(render.NewImageData(img), geom.Rect{Min: geom.Pt{X: 12, Y: 8}, Max: geom.Pt{X: 20, Y: 16}})

	if err := r.End(); err != nil {
		t.Fatalf("End() error = %v", err)
	}

	out := r.Image()
	if out == nil {
		t.Fatal("Image() returned nil")
	}
	if got := out.RGBAAt(1, 1); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("pixel outside clip = %#v, want unchanged white", got)
	}
	if got := out.RGBAAt(6, 6); got == (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("pixel inside clip stayed background: %#v", got)
	}
	if got := out.RGBAAt(14, 10); got == (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("image draw pixel stayed background: %#v", got)
	}

	path := filepath.Join(t.TempDir(), "skia.png")
	if err := r.SavePNG(path); err != nil {
		t.Fatalf("SavePNG() error = %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open saved PNG: %v", err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode saved PNG: %v", err)
	}
	if decoded.Bounds().Dx() != 32 || decoded.Bounds().Dy() != 24 {
		t.Fatalf("saved PNG bounds = %v, want 32x24", decoded.Bounds())
	}
}

func TestSkiaTaggedRendererDrawsMathText(t *testing.T) {
	fig := core.NewFigure(120, 80)
	fig.Text(0.5, 0.5, `$\frac{1}{2}+\alpha_i$`, core.TextOptions{
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignMiddle,
		FontSize: 16,
	})

	r, err := skia.New(backends.Config{
		Width:      120,
		Height:     80,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		Options:    backends.SkiaConfig{UseGPU: false},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	core.DrawFigure(fig, r)
	if got := nonWhitePixels(r.Image()); got == 0 {
		t.Fatal("MathText rendered through Skia-tagged backend produced a blank image")
	}
}

func TestSkiaTaggedRendererMatchesMathTextGoldens(t *testing.T) {
	for _, id := range []string{
		"mathtext_basic",
		"mathtext_fractions",
		"mathtext_integrals",
		"mathtext_matrices",
		"mathtext_inline_labels",
	} {
		id := id
		t.Run(id, func(t *testing.T) {
			fig, _, err := parity.Figure(id)
			if err != nil {
				t.Fatalf("parity figure %s: %v", id, err)
			}
			r, err := skia.New(backends.Config{
				Width:      int(fig.SizePx.X),
				Height:     int(fig.SizePx.Y),
				Background: render.Color{R: 1, G: 1, B: 1, A: 1},
				Options:    backends.SkiaConfig{UseGPU: false},
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			core.DrawFigure(fig, r)
			got := r.Image()
			if got == nil {
				t.Fatal("Skia renderer returned nil image")
			}
			wantPath := filepath.Join("..", "..", "testdata", "golden", id+".png")
			want, err := imagecmp.LoadPNG(wantPath)
			if err != nil {
				t.Fatalf("load AGG golden %s: %v", wantPath, err)
			}
			diff, err := imagecmp.ComparePNG(got, want, 255)
			if err != nil {
				t.Fatalf("compare Skia MathText against AGG golden: %v", err)
			}
			const (
				minPSNR    = 35.0
				maxMeanAbs = 3.0
			)
			if diff.PSNR < minPSNR || diff.MeanAbs > maxMeanAbs {
				artifactsDir := filepath.Join("..", "..", "testdata", "_artifacts", "skia_mathtext")
				if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
					t.Fatalf("create artifacts dir: %v", err)
				}
				if err := imagecmp.SavePNG(got, filepath.Join(artifactsDir, id+"_got.png")); err != nil {
					t.Fatalf("save Skia artifact: %v", err)
				}
				if err := imagecmp.SavePNG(want, filepath.Join(artifactsDir, id+"_agg_golden.png")); err != nil {
					t.Fatalf("save AGG golden artifact: %v", err)
				}
				if err := imagecmp.SaveDiffImage(got, want, 10, filepath.Join(artifactsDir, id+"_diff.png")); err != nil {
					t.Fatalf("save diff artifact: %v", err)
				}
				t.Fatalf("Skia MathText mismatch for %s: PSNR=%.2f (min %.2f), MeanAbs=%.2f (max %.2f), MaxDiff=%d",
					id, diff.PSNR, minPSNR, diff.MeanAbs, maxMeanAbs, diff.MaxDiff)
			}
		})
	}
}

func TestSkiaTaggedRegistryAdvertisesImplementedCPUCapabilities(t *testing.T) {
	info, ok := backends.DefaultRegistry.Get(backends.Skia)
	if !ok {
		t.Fatal("skia backend should be registered")
	}
	if !info.Available {
		t.Fatal("skia build-tag backend should be available")
	}
	if _, ok := info.SaveFormats[".png"]; !ok {
		t.Fatal("skia build-tag backend should advertise PNG save dispatch")
	}

	renderer, err := backends.Create(backends.Skia, backends.TestDefaultConfig(32, 24))
	if err != nil {
		t.Fatalf("Create(skia) error = %v", err)
	}
	if err := backends.VerifyRendererCapabilities(backends.Skia, renderer); err != nil {
		t.Fatalf("VerifyRendererCapabilities(skia) error = %v", err)
	}

	for _, cap := range []backends.Capability{
		backends.AntiAliasing,
		backends.PathClip,
		backends.DPIAware,
		backends.TextShaping,
		backends.TextPathing,
		backends.RotatedText,
		backends.VerticalText,
		backends.RGBABuffer,
		backends.PNGExport,
	} {
		if !backends.HasCapability(backends.Skia, cap) {
			t.Fatalf("skia backend should advertise %s", cap)
		}
	}
	if backends.HasCapability(backends.Skia, backends.GPUAccel) {
		t.Fatal("CPU skia backend should not advertise GPU acceleration")
	}

	if status := backends.RendererCapabilityStatus(backends.Skia, renderer, backends.GouraudTriangleBatch); status != backends.CapabilityNative {
		t.Fatalf("RendererCapabilityStatus(skia, %s) = %s, want %s", backends.GouraudTriangleBatch, status, backends.CapabilityNative)
	}

	// ImageTransform, NativeHatcher, MarkerBatch, PathCollectionBatch, and
	// QuadMeshBatch are native when a real Skia library is linked (skiacgo
	// build); the pure-Go build satisfies them through the CPU bridge.
	wantNativeBatch := backends.CapabilityBridged
	if sk, ok := renderer.(*skia.Renderer); ok && sk.BridgeInfo().NativeSurface {
		wantNativeBatch = backends.CapabilityNative
	}
	for _, cap := range []backends.Capability{
		backends.ImageTransform,
		backends.MarkerBatch,
		backends.NativeHatcher,
		backends.PathCollectionBatch,
		backends.QuadMeshBatch,
	} {
		if status := backends.RendererCapabilityStatus(backends.Skia, renderer, cap); status != wantNativeBatch {
			t.Fatalf("RendererCapabilityStatus(skia, %s) = %s, want %s", cap, status, wantNativeBatch)
		}
	}
}

func TestSkiaTaggedBackendComparisonReportLabelsCPUMode(t *testing.T) {
	report := backends.BackendComparisonReport(backends.TestDefaultConfig(32, 24))
	if !strings.Contains(report, "skia/cpu") {
		t.Fatalf("BackendComparisonReport did not label Skia CPU mode:\n%s", report)
	}
	r, err := skia.New(backends.TestDefaultConfig(32, 24))
	if err != nil {
		t.Fatalf("New CPU renderer: %v", err)
	}
	if status := backends.RendererCapabilityStatus(backends.Skia, r, backends.GPUAccel); status != backends.CapabilityUnsupported {
		t.Fatalf("CPU gpuaccel status = %q, want %q", status, backends.CapabilityUnsupported)
	}
}

func TestSkiaTaggedPathEffectFilterUsesOffscreenSurface(t *testing.T) {
	renderer, err := skia.New(backends.TestDefaultConfig(80, 60))
	if err != nil {
		t.Fatalf("New(skia) error = %v", err)
	}
	if _, ok := any(renderer).(render.FilterRenderer); !ok {
		t.Fatal("skia tagged renderer should expose render.FilterRenderer for filtered path effects")
	}
	if err := renderer.Begin(geom.Rect{Max: geom.Pt{X: 80, Y: 60}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	renderer.Path(skiaTestRectPath(25, 20, 20, 20), &render.Paint{
		Fill: render.Color{G: 1, A: 1},
		PathEffects: []render.PathEffect{
			render.FilterPathEffect(render.Color{R: 1, A: 1}, render.Color{}, 0, "blur", 4, geom.Pt{}),
			render.NormalPathEffect(),
		},
	})
	if err := renderer.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	img := renderer.Image()
	if got := img.RGBAAt(23, 30); got.R <= got.G {
		t.Fatalf("expected blurred red filter pass outside normal green fill, got %+v", got)
	}
	if got := img.RGBAAt(35, 30); got.G == 0 {
		t.Fatalf("expected normal green fill to remain after filter pass, got %+v", got)
	}
}

func nonWhitePixels(img *image.RGBA) int {
	if img == nil {
		return 0
	}
	count := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			if img.RGBAAt(x, y) != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
				count++
			}
		}
	}
	return count
}

func skiaTestRectPath(x, y, w, h float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: x, Y: y})
	p.LineTo(geom.Pt{X: x + w, Y: y})
	p.LineTo(geom.Pt{X: x + w, Y: y + h})
	p.LineTo(geom.Pt{X: x, Y: y + h})
	p.Close()
	return p
}

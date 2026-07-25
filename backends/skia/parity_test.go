//go:build skia && skiacgo

// This test lives in the external skia_test package on purpose: it imports
// test/parity, which transitively pulls in backends/all (→ backends/skia). An
// internal (package skia) test importing that chain forms an import cycle, so
// the skia-tagged parity tests must compile as an external test package.
package skia_test

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/backends/skia"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

// Default tolerances when a catalog case does not override them. The native
// Skia renderer still uses renderer-neutral fallbacks for text and some artist
// features, so complete figures do not necessarily pixel-match AGG exactly.
const (
	defaultSkiaParityMinPSNR    = 22.0
	defaultSkiaParityMaxMeanAbs = 22.0
)

// TestSkiaParityAgainstAGGGoldens iterates every catalog case opted into a
// SkiaParityFamily and compares the skiacgo native CPU output against the
// committed AGG golden under testdata/golden/. Per-case MinPSNR / MaxMeanAbs
// overrides on the catalog row take precedence over the package defaults; this
// matches the AGENTS.md convention that the catalog is the single source of
// truth for tolerances. Failures emit got / golden / diff artifacts under
// testdata/_artifacts/skia_parity/{id}/.
func TestSkiaParityAgainstAGGGoldens(t *testing.T) {
	cases := examplecatalog.Cases()
	any := false
	for _, c := range cases {
		if c.SkiaParityFamily == "" {
			continue
		}
		any = true
		c := c
		t.Run(c.ID, func(t *testing.T) {
			fig, _, err := parity.Figure(c.ID)
			if err != nil {
				t.Skipf("parity figure %s unavailable: %v", c.ID, err)
			}
			got := renderSkiaFigure(t, fig)

			wantPath := filepath.Join("..", "..", "testdata", "golden", c.ID+".png")
			want, err := imagecmp.LoadPNG(wantPath)
			if err != nil {
				t.Fatalf("load AGG golden %s: %v", wantPath, err)
			}
			diff, err := imagecmp.ComparePNG(got, want, 255)
			if err != nil {
				t.Fatalf("compare Skia vs AGG golden %s: %v", c.ID, err)
			}

			minPSNR, maxMeanAbs := skiaParityTolerances(c)
			if diff.PSNR >= minPSNR && diff.MeanAbs <= maxMeanAbs {
				return
			}

			artifactsDir := filepath.Join("..", "..", "testdata", "_artifacts", "skia_parity", c.ID)
			if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
				t.Fatalf("create artifacts dir %s: %v", artifactsDir, err)
			}
			if err := imagecmp.SavePNG(got, filepath.Join(artifactsDir, c.ID+"_got.png")); err != nil {
				t.Fatalf("save got artifact: %v", err)
			}
			if err := imagecmp.SavePNG(want, filepath.Join(artifactsDir, c.ID+"_agg_golden.png")); err != nil {
				t.Fatalf("save golden artifact: %v", err)
			}
			if err := imagecmp.SaveDiffImage(got, want, 10, filepath.Join(artifactsDir, c.ID+"_diff.png")); err != nil {
				t.Fatalf("save diff artifact: %v", err)
			}
			t.Fatalf("Skia vs AGG parity mismatch for %s (family=%s): PSNR=%.2f (min %.2f), MeanAbs=%.2f (max %.2f), MaxDiff=%d",
				c.ID, c.SkiaParityFamily, diff.PSNR, minPSNR, diff.MeanAbs, maxMeanAbs, diff.MaxDiff)
		})
	}
	if !any {
		t.Fatal("no catalog cases opted into SkiaParityFamily; check internal/examplecatalog/catalog.go")
	}
}

func renderSkiaFigure(t *testing.T, fig *core.Figure) image.Image {
	t.Helper()
	width := int(fig.SizePx.X)
	height := int(fig.SizePx.Y)
	r, err := skia.New(backends.Config{
		Width:      width,
		Height:     height,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		Options:    backends.SkiaConfig{UseGPU: false},
	})
	if err != nil {
		t.Fatalf("create Skia renderer: %v", err)
	}
	core.DrawFigure(fig, r)
	img := r.GetImage()
	if img == nil {
		t.Fatal("Skia renderer returned nil image")
	}
	return img
}

// skiaParityMaxMeanAbsOverride relaxes the per-case MeanAbs ceiling for cases
// whose catalog tolerance is tuned for AGG-vs-matplotlib_ref parity
// (TestReferenceCompare) but whose skia-CPU-vs-AGG comparison carries
// irreducible cross-rasterizer differences. Skia inherits renderer-neutral
// path-effect placement while native Skia edges differ from AGG antialiasing.
// pattern_gradient_effects: shader fills (gradient/radial/pattern) match AGG at
// PSNR ~44.5 after the device y-flip fix; the residual MeanAbs ~2.2 is the
// drop-shadow path-effect orientation, hatch line density, and edge AA — none of
// which are shader fills. 3.0 keeps the case a tight regression guard (a
// re-introduced fill y-flip blows well past it) without demanding AGG-exact
// path-effect/hatch parity the CPU bridge does not promise.
var skiaParityMaxMeanAbsOverride = map[string]float64{
	"errorbar_basic":           0.5,
	"pattern_gradient_effects": 3.0,
}

var skiaParityMinPSNROverride = map[string]float64{
	"errorbar_basic": 50.0,
}

func skiaParityTolerances(c examplecatalog.Case) (minPSNR, maxMeanAbs float64) {
	minPSNR = defaultSkiaParityMinPSNR
	maxMeanAbs = defaultSkiaParityMaxMeanAbs
	if c.MinPSNR > 0 {
		minPSNR = c.MinPSNR
	}
	if c.MaxMeanAbs > 0 {
		maxMeanAbs = c.MaxMeanAbs
	}
	if override, ok := skiaParityMaxMeanAbsOverride[c.ID]; ok {
		maxMeanAbs = override
	}
	if override, ok := skiaParityMinPSNROverride[c.ID]; ok {
		minPSNR = override
	}
	return minPSNR, maxMeanAbs
}

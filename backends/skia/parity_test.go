//go:build skia

package skia

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

// Default tolerances when a catalog case does not override them. The Skia CPU
// compatibility renderer wraps the GoBasic surface, so most catalog cases will
// not pixel-match AGG exactly. These defaults are intentionally generous to
// document that Skia parity is a regression guard against the CPU bridge
// silently drifting, not an exact-match contract.
const (
	defaultSkiaParityMinPSNR    = 22.0
	defaultSkiaParityMaxMeanAbs = 22.0
)

// TestSkiaParityAgainstAGGGoldens iterates every catalog case opted into a
// SkiaParityFamily and compares the Skia CPU renderer output against the
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
	r, err := New(backends.Config{
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

func skiaParityTolerances(c examplecatalog.Case) (minPSNR, maxMeanAbs float64) {
	minPSNR = defaultSkiaParityMinPSNR
	maxMeanAbs = defaultSkiaParityMaxMeanAbs
	if c.MinPSNR > 0 {
		minPSNR = c.MinPSNR
	}
	if c.MaxMeanAbs > 0 {
		maxMeanAbs = c.MaxMeanAbs
	}
	return minPSNR, maxMeanAbs
}

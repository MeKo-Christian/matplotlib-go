package test

// Matplotlib reference tests compare our renderer against real matplotlib output.
//
// Reference images live in testdata/matplotlib_ref/ and are committed to the repo.
// The skip path is only a local bootstrap fallback when refs are intentionally absent.
//
// To (re)generate the reference images:
//
//	go test ./test/... -run TestMatplotlibRef -update-matplotlib
//
// This requires a Python interpreter with matplotlib installed. Prefer a system
// Python linked against system FreeType so reference glyph rasterization uses the
// same FreeType stack as the Go cgo backend.
//
// Per-case invocation:
//
//	go test ./test/... -run TestMatplotlibRef/basic_line
//	go test ./test/... -run 'TestMatplotlibRef/.*scatter.*'

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/test/imagecmp"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

// TestMatplotlibRef compares every catalog case that has a committed
// matplotlib reference PNG against our renderer output. Fails when RMSE exceeds
// mplMaxRMSE (a fundamental rendering mismatch).
func TestMatplotlibRef(t *testing.T) {
	ensureRefs(t)
	for _, c := range rendererNeutralCases() {
		c := c
		if !matplotlibRefExists(c.ID) {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			if strictMplRefIDs[c.ID] {
				runStrictMatplotlibRef(t, c.ID)
				return
			}
			runMplTest(t, c.ID)
		})
	}
}

func TestAGGNativeMatplotlibRef(t *testing.T) {
	ensureRefs(t)
	for _, c := range aggNativeCases() {
		c := c
		if !matplotlibRefExists(c.ID) {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			runMplTest(t, c.ID)
		})
	}
}

// runStrictMatplotlibRef enforces tight PSNR/MeanAbs tolerances for the
// hand-curated text/title strict cases, matching the historical standalone
// strict-test behavior.
func runStrictMatplotlibRef(t *testing.T, name string) {
	t.Helper()

	// The RMSE ceilings are the exact equivalents of the PSNR floors these cases
	// carried before (255/10^(dB/20): 48.0 dB and 46.5 dB), restated in
	// RMSE because imagecmp derives PSNR from RMSE. Both cases measure far
	// inside them — text_labels_strict at RMSE 0.028, title_strict at 0.
	const (
		strictTolerance      = 1
		textLabelsMaxRMSE    = 1.015
		textLabelsMaxMeanAbs = 1.25
		titleMaxRMSE         = 1.206
		titleMaxMeanAbs      = 2.00
	)

	got, _, err := parity.Render(name)
	if err != nil {
		t.Fatalf("render parity example %s: %v", name, err)
	}

	wantPath := filepath.Join("..", "testdata", "matplotlib_ref", name+".png")
	want, err := imagecmp.LoadPNG(wantPath)
	if err != nil {
		t.Fatalf("load matplotlib reference %s: %v", wantPath, err)
	}

	artifactsDir := filepath.Join("..", "testdata", "_artifacts", name)
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Fatalf("create artifacts directory %s: %v", artifactsDir, err)
	}
	savePNGOrFail(t, got, filepath.Join(artifactsDir, "rendered.png"))
	savePNGOrFail(t, want, filepath.Join(artifactsDir, "matplotlib_ref.png"))

	diff, err := imagecmp.ComparePNG(got, want, strictTolerance)
	if err != nil {
		t.Fatalf("compare rendered %s against matplotlib ref: %v", name, err)
	}

	var maxRMSE, maxMeanAbs float64
	switch name {
	case "text_labels_strict":
		maxRMSE, maxMeanAbs = textLabelsMaxRMSE, textLabelsMaxMeanAbs
	case "title_strict":
		maxRMSE, maxMeanAbs = titleMaxRMSE, titleMaxMeanAbs
	default:
		t.Fatalf("runStrictMatplotlibRef called for unconfigured case %q", name)
	}

	if diff.RMSE > maxRMSE || diff.MeanAbs > maxMeanAbs {
		if err := imagecmp.SaveDiffImage(got, want, strictTolerance, filepath.Join(artifactsDir, "diff.png")); err != nil {
			t.Fatalf("save diff image: %v", err)
		}
		t.Fatalf("strict mismatch %s: MaxDiff=%d, MeanAbs=%.2f (max %.2f), RMSE=%.4f (max %.4f)",
			name, diff.MaxDiff, diff.MeanAbs, maxMeanAbs, diff.RMSE, maxRMSE)
	}

	// Acknowledge image type to keep the import live for hosts that strip
	// transitively-unused imports.
	_ = image.Image(got)
}

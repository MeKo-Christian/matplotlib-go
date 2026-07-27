package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/test/imagecmp"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

// === non-text residual diagnostics ====================================
//
// PLAN.md calls for AGG parity diagnostics on the remaining non-text
// residuals. These are opt-in dev probes: they render catalog cases, log the
// AGG-vs-Matplotlib residual metrics, and dump diff PNGs so the gaps can be
// localized. They never assert a ceiling, so the default suite is unaffected.

const (
	// nonTextResidualEnv gates the non-text residual probes.
	nonTextResidualEnv = "MPL_GO_RESIDUAL_DIAG"
	// nonTextResidualTolerance is the per-channel LSB tolerance for the residual
	// metrics, matching the reference-compare driver's 1-LSB tolerance.
	nonTextResidualTolerance = 1
	// nonTextResidualHighDiffThreshold flags a pixel as a high-diff residual when
	// any channel differs by more than this; used for bbox localization and the
	// diff-image highlight.
	nonTextResidualHighDiffThreshold = 32
)

type nonTextResidualCase struct {
	id       string
	category string
}

// TestNonTextResidualDiagnostics probes the remaining non-text AGG parity
// residuals across five categories: dense path collections, repeated
// translucent overlaps, image interpolation modes, hatch clipping, and mixed
// raster/vector fallbacks. It is gated by MPL_GO_RESIDUAL_DIAG, logs metrics,
// and writes diff artifacts under testdata/_artifacts/non_text_residuals/. It
// does not assert a residual ceiling — it is a localization probe, not a gate.
func TestNonTextResidualDiagnostics(t *testing.T) {
	if os.Getenv(nonTextResidualEnv) == "" {
		t.Skipf("set %s=1 to run non-text residual diagnostics", nonTextResidualEnv)
	}

	cases := []nonTextResidualCase{
		{id: "large_scatter", category: "dense-path-collections"},
		{id: "mixed_collection", category: "dense-path-collections"},
		{id: "scatter_gallery", category: "translucent-overlaps"},
		{id: "image_alpha", category: "translucent-overlaps"},
		{id: "imshow_interpolation_matrix", category: "image-interpolation"},
		{id: "imshow_bilinear", category: "image-interpolation"},
		{id: "imshow_bicubic", category: "image-interpolation"},
		{id: "patch_showcase", category: "hatch-clipping"},
		{id: "patch_style_matrix", category: "hatch-clipping"},
		{id: "mixed_raster_vector", category: "mixed-raster-vector"},
	}

	artifactsDir := writableArtifactsDir(t, filepath.Join("..", "testdata", "_artifacts", "non_text_residuals"))

	for _, tc := range cases {
		tc := tc
		t.Run(tc.category+"/"+tc.id, func(t *testing.T) {
			if !matplotlibRefExists(tc.id) {
				t.Skipf("no matplotlib reference for %s", tc.id)
			}

			got, _, err := parity.Render(tc.id)
			if err != nil {
				t.Skipf("render parity example %s: %v", tc.id, err)
			}
			want, err := imagecmp.LoadPNG(filepath.Join("..", "testdata", "matplotlib_ref", tc.id+".png"))
			if err != nil {
				t.Skipf("load matplotlib reference %s: %v", tc.id, err)
			}

			diff, err := imagecmp.ComparePNG(got, want, nonTextResidualTolerance)
			if err != nil {
				t.Skipf("compare %s: %v", tc.id, err)
			}

			// Reuse the alpha-residual summary to localize high-diff pixels over
			// the full rendered frame.
			summary := alphaResidualDiagnostics(got, want, got.Bounds(), nonTextResidualHighDiffThreshold)

			t.Logf(
				"category=%s id=%s MaxDiff=%d MeanAbs=%.3f RMSE=%.3f PSNR=%.2fdB highDiff=%d/%d (%.4f) bbox=%v",
				tc.category, tc.id,
				diff.MaxDiff, diff.MeanAbs, diff.RMSE, diff.PSNR,
				summary.highDiff, summary.total, summary.highDiffRatio(), summary.bbox,
			)

			savePNGOrFail(t, got, filepath.Join(artifactsDir, tc.id+"_rendered.png"))
			savePNGOrFail(t, want, filepath.Join(artifactsDir, tc.id+"_matplotlib_ref.png"))
			diffPath := filepath.Join(artifactsDir, tc.id+"_diff.png")
			if err := imagecmp.SaveDiffImage(got, want, nonTextResidualHighDiffThreshold, diffPath); err != nil {
				t.Fatalf("save diff image %s: %v", diffPath, err)
			}
		})
	}
}

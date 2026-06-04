package test

// Shared test helpers for golden, matplotlib reference, and reference-compare
// suites. The drivers in golden_test.go, matplotlib_ref_test.go, and
// reference_compare_test.go iterate examplecatalog.Cases() and call into the
// helpers below.

import (
	"flag"
	"fmt"
	"image"
	"math"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

// ----------------------------------------------------------------------------
// CLI flags
// ----------------------------------------------------------------------------

var (
	updateGolden     = flag.Bool("update-golden", false, "Update golden images instead of comparing")
	updateMatplotlib = flag.Bool("update-matplotlib", false,
		"Regenerate matplotlib reference images via `uv run` (or python3) before comparing")
)

// ----------------------------------------------------------------------------
// Constants and defaults
// ----------------------------------------------------------------------------

const (
	mplRefDir  = "../testdata/matplotlib_ref"
	mplMinPSNR = 10.0 // dB; below this indicates a fundamental rendering error

	referenceCompareTolerance  = 1
	referenceCompareMinPSNR    = 44.0
	referenceCompareMaxMeanAbs = 2.50

	optionalVisualTestsEnv = "RUN_OPTIONAL_VISUAL_TESTS"
)

// goldenReadPath returns the golden PNG path for id. The AGG backend links the
// vendored FreeType 2.6.1 (matplotlib's pinned version) by default, so a single
// canonical golden set in testdata/golden/ byte-matches both the live render and
// the matplotlib references — no per-FreeType-version split is needed.
func goldenReadPath(id string) string {
	return filepath.Join("..", "testdata", "golden", id+".png")
}

// goldenWriteDir returns the directory that -update-golden writes to.
func goldenWriteDir() string {
	return "golden"
}

// optionalVisualGoldenIDs lists catalog cases whose golden tests are gated by
// RUN_OPTIONAL_VISUAL_TESTS=true. Cases not in this set always run. The set
// reflects the historical gating that lived in golden_test.go and the small
// fixture/showcase files; it does not perfectly track the catalog's Optional
// flag (e.g. specialty_depth was gated despite being FixtureOnly, not Optional).
var optionalVisualGoldenIDs = map[string]bool{
	"boxplot_basic":           true,
	"specialty_depth":         true,
	"errorbar_basic":          true,
	"text_labels_strict":      true,
	"axes_top_right_inverted": true,
	"axes_control_surface":    true,
	"transform_coordinates":   true,
	"plot_variants":           true,
	"spectrum_variants":       true,
	"stat_variants":           true,
	"units_overview":          true,
	"units_dates":             true,
	"units_categories":        true,
	"units_custom_converter":  true,
	"patch_showcase":          true,
	"mesh_contour_tri":        true,
	"stem_plot":               true,
	"specialty_artists":       true,
	"vector_fields":           true,
	"geo_aitoff_axes":         true,
	"geo_hammer_axes":         true,
	"geo_lambert_axes":        true,
	"radar_basic":             true,
	"skewt_basic":             true,
	"mplot3d_basic":           true,
	"mplot3d_terrain":         true,
	"mplot3d_plot3d":          true,
	"mplot3d_scatter3d":       true,
	"mplot3d_surface3d":       true,
	"mplot3d_wire3d":          true,
	"mplot3d_trisurf3d":       true,
	"mplot3d_bar3d":           true,
	"mplot3d_voxels":          true,
	"mplot3d_quiver3d":        true,
	"mplot3d_errorbar3d":      true,
	"mplot3d_stem3d":          true,
	"mplot3d_fill_between3d":  true,
	"unstructured_showcase":   true,
	"arrays_showcase":         true,
	"axisartist_showcase":     true,
	"axes_grid1_showcase":     true,
}

// optionalVisualMplRefIDs gates matplotlib_ref tests that historically lived
// in test/text_strict_test.go and were always RUN_OPTIONAL_VISUAL_TESTS-gated.
var optionalVisualMplRefIDs = map[string]bool{
	"text_labels_strict": true,
	"title_strict":       true,
}

// ----------------------------------------------------------------------------
// Optional-test gate
// ----------------------------------------------------------------------------

func requireOptionalVisualTests(t *testing.T) {
	t.Helper()
	if os.Getenv(optionalVisualTestsEnv) == "true" {
		return
	}
	t.Skip("skipping optional visual parity test (set RUN_OPTIONAL_VISUAL_TESTS=true to run)")
}

// ----------------------------------------------------------------------------
// Golden image driver
// ----------------------------------------------------------------------------

// runGoldenTest renders the parity example for testName and compares it
// against the committed golden PNG. With -update-golden the golden image is
// rewritten and the test skipped.
func runGoldenTest(t *testing.T, testName string) {
	img, _, err := parity.Render(testName)
	if err != nil {
		t.Fatalf("Failed to render parity example %s: %v", testName, err)
	}

	if *updateGolden {
		writePath := "../testdata/" + goldenWriteDir() + "/" + testName + ".png"
		if err := imagecmp.SavePNG(img, writePath); err != nil {
			t.Fatalf("Failed to update golden image: %v", err)
		}
		t.Skip("Updated golden image")
		return
	}

	goldenPath := goldenReadPath(testName)
	want, err := imagecmp.LoadPNG(goldenPath)
	if err != nil {
		t.Fatalf("Failed to load golden image %s: %v", goldenPath, err)
	}

	diff, err := imagecmp.ComparePNG(img, want, 1) // ≤1 LSB tolerance
	if err != nil {
		t.Fatalf("Image comparison failed: %v", err)
	}

	if !diff.Identical {
		artifactsDir := "../testdata/_artifacts"
		if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
			t.Fatalf("Could not create artifacts directory %s: %v", artifactsDir, err)
		}
		gotPath := filepath.Join(artifactsDir, testName+"_got.png")
		if err := imagecmp.SavePNG(img, gotPath); err != nil {
			t.Fatalf("Could not save got image %s: %v", gotPath, err)
		}
		diffPath := filepath.Join(artifactsDir, testName+"_diff.png")
		if err := imagecmp.SaveDiffImage(img, want, 1, diffPath); err != nil {
			t.Fatalf("Could not save diff image %s: %v", diffPath, err)
		}
		t.Logf("Debug images saved to %s/", artifactsDir)
		t.Fatalf("Golden image mismatch: MaxDiff=%d, MeanAbs=%.2f, PSNR=%.2fdB",
			diff.MaxDiff, diff.MeanAbs, diff.PSNR)
	}

	t.Logf("Golden image match: MaxDiff=%d, MeanAbs=%.2f, PSNR=%.2fdB",
		diff.MaxDiff, diff.MeanAbs, diff.PSNR)
}

// goldenExists reports whether testdata/golden/{id}.png is present.
func goldenExists(id string) bool {
	_, err := os.Stat(filepath.Join("..", "testdata", "golden", id+".png"))
	return err == nil
}

// ----------------------------------------------------------------------------
// Matplotlib reference driver
// ----------------------------------------------------------------------------

var (
	mplOnce sync.Once
	mplErr  error
)

func matplotlibPythonPath() (string, error) {
	candidates := []string{}
	if env := os.Getenv("MATPLOTLIB_GO_PYTHON"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates, "/usr/bin/python3")
	if pyPath, err := exec.LookPath("python3"); err == nil {
		candidates = append(candidates, pyPath)
	}

	seen := map[string]bool{}
	for _, candidate := range candidates {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		cmd := exec.Command(candidate, "-c", "import matplotlib.ft2font")
		if err := cmd.Run(); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no Python interpreter with matplotlib found; set MATPLOTLIB_GO_PYTHON")
}

// ensureRefs regenerates reference images when -update-matplotlib is set,
// or skips the calling test when the committed ref directory is missing
// locally.
func ensureRefs(t *testing.T) {
	t.Helper()
	if !*updateMatplotlib {
		if _, err := os.Stat(mplRefDir); os.IsNotExist(err) {
			t.Skip("matplotlib refs not found – run with -update-matplotlib to generate")
		}
		return
	}

	mplOnce.Do(func() {
		script := filepath.Join("matplotlib_ref", "generate.py")
		outDir, _ := filepath.Abs(mplRefDir)

		pyPath, err := matplotlibPythonPath()
		if err != nil {
			mplErr = err
			return
		}
		cmd := exec.Command(pyPath, script, "--output-dir", outDir)
		if repoRoot, err := filepath.Abs(".."); err == nil {
			cmd.Env = append(os.Environ(), "PYTHONPATH="+prependEnvPath(repoRoot, os.Getenv("PYTHONPATH")))
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		mplErr = cmd.Run()
	})

	if mplErr != nil {
		t.Fatalf("matplotlib reference generation failed: %v", mplErr)
	}
}

func prependEnvPath(path, existing string) string {
	if existing == "" {
		return path
	}
	return path + string(os.PathListSeparator) + existing
}

// runMplTest renders the parity example, loads the matplotlib reference,
// writes artefacts for inspection, and fails only when PSNR drops below
// mplMinPSNR (a fundamental rendering error).
func runMplTest(t *testing.T, name string) {
	t.Helper()
	ensureRefs(t)

	got, _, err := parity.Render(name)
	if err != nil {
		t.Fatalf("render parity example %s: %v", name, err)
	}

	refPath := filepath.Join(mplRefDir, name+".png")
	want, err := imagecmp.LoadPNG(refPath)
	if err != nil {
		t.Fatalf("load matplotlib ref %s: %v", refPath, err)
	}

	diff, err := imagecmp.ComparePNG(got, want, 255)
	if err != nil {
		t.Fatalf("image comparison failed: %v", err)
	}

	t.Logf("PSNR=%.1f dB  MeanAbs=%.2f  MaxDiff=%d", diff.PSNR, diff.MeanAbs, diff.MaxDiff)

	artifactsDir := matplotlibArtifactsDir(t)
	if err := imagecmp.SavePNG(got, filepath.Join(artifactsDir, name+"_go.png")); err != nil {
		t.Fatalf("could not save rendered image: %v", err)
	}
	if err := imagecmp.SavePNG(want, filepath.Join(artifactsDir, name+"_mpl.png")); err != nil {
		t.Fatalf("could not save matplotlib image: %v", err)
	}
	if err := imagecmp.SaveDiffImage(got, want, 10, filepath.Join(artifactsDir, name+"_diff.png")); err != nil {
		t.Fatalf("could not save diff image: %v", err)
	}

	if !math.IsInf(diff.PSNR, 1) && diff.PSNR < mplMinPSNR {
		t.Errorf("PSNR %.1f dB < %.1f dB: likely fundamental rendering mismatch with matplotlib",
			diff.PSNR, mplMinPSNR)
	}
}

func matplotlibArtifactsDir(t *testing.T) string {
	t.Helper()
	return writableArtifactsDir(t, filepath.Join("..", "testdata", "_artifacts", "matplotlib_ref"))
}

// matplotlibRefExists reports whether testdata/matplotlib_ref/{id}.png is
// present.
func matplotlibRefExists(id string) bool {
	_, err := os.Stat(filepath.Join("..", "testdata", "matplotlib_ref", id+".png"))
	return err == nil
}

// ----------------------------------------------------------------------------
// Reference-compare driver
// ----------------------------------------------------------------------------

// runReferenceCompareTest cross-checks the golden + matplotlib reference for
// the given case. Tolerances come from the catalog row; zero values fall back
// to the package defaults above.
func runReferenceCompareTest(t *testing.T, c *examplecatalog.Case) {
	t.Helper()

	got, _, err := parity.Render(c.ID)
	if err != nil {
		t.Fatalf("render parity example %s: %v", c.ID, err)
	}

	goldenPath := goldenReadPath(c.ID)
	matplotlibPath := filepath.Join("..", "testdata", "matplotlib_ref", c.ID+".png")

	golden, err := imagecmp.LoadPNG(goldenPath)
	if err != nil {
		t.Fatalf("load golden image %s: %v", goldenPath, err)
	}

	matplotlibRef, err := imagecmp.LoadPNG(matplotlibPath)
	if err != nil {
		t.Fatalf("load matplotlib reference %s: %v", matplotlibPath, err)
	}

	artifactsDir := referenceCompareArtifactsDir(t)
	savePNGOrFail(t, got, filepath.Join(artifactsDir, c.ID+"_rendered.png"))
	savePNGOrFail(t, golden, filepath.Join(artifactsDir, c.ID+"_golden.png"))
	savePNGOrFail(t, matplotlibRef, filepath.Join(artifactsDir, c.ID+"_matplotlib_ref.png"))

	diff, err := imagecmp.ComparePNG(golden, matplotlibRef, referenceCompareTolerance)
	if err != nil {
		t.Fatalf("compare %s and %s: %v", goldenPath, matplotlibPath, err)
	}

	diffPath := filepath.Join(artifactsDir, c.ID+"_golden_vs_matplotlib_ref_diff.png")
	if err := imagecmp.SaveDiffImage(golden, matplotlibRef, referenceCompareTolerance, diffPath); err != nil {
		t.Fatalf("save diff image %s: %v", diffPath, err)
	}

	t.Logf("%s: MaxDiff=%d MeanAbs=%.2f RMSE=%.2f PSNR=%.2fdB",
		c.ID, diff.MaxDiff, diff.MeanAbs, diff.RMSE, diff.PSNR)

	minPSNR := referenceCompareMinPSNR
	if c.MinPSNR > 0 {
		minPSNR = c.MinPSNR
	}
	maxMeanAbs := referenceCompareMaxMeanAbs
	if c.MaxMeanAbs > 0 {
		maxMeanAbs = c.MaxMeanAbs
	}
	if diff.PSNR < minPSNR || diff.MeanAbs > maxMeanAbs || (c.MaxRMSE > 0 && diff.RMSE > c.MaxRMSE) {
		t.Fatalf("reference mismatch for %s: PSNR=%.2f (min %.2f), MeanAbs=%.2f (max %.2f), RMSE=%.2f (max %.2f)",
			c.ID, diff.PSNR, minPSNR, diff.MeanAbs, maxMeanAbs, diff.RMSE, c.MaxRMSE)
	}
}

func referenceCompareArtifactsDir(t *testing.T) string {
	t.Helper()
	return writableArtifactsDir(t, filepath.Join("..", "testdata", "_artifacts", "reference_compare"))
}

// ----------------------------------------------------------------------------
// Misc helpers
// ----------------------------------------------------------------------------

func savePNGOrFail(t *testing.T, img image.Image, path string) {
	t.Helper()
	if err := imagecmp.SavePNG(img, path); err != nil {
		t.Fatalf("save PNG %s: %v", path, err)
	}
}

// writableArtifactsDir creates artifactsDir if missing and falls back to t.TempDir
// when the location is not writable.
func writableArtifactsDir(t *testing.T, artifactsDir string) string {
	t.Helper()
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Logf("could not create artifacts directory %s: %v; using temp dir", artifactsDir, err)
		return t.TempDir()
	}
	probe, err := os.CreateTemp(artifactsDir, ".write-probe-*")
	if err != nil {
		t.Logf("artifacts directory %s is not writable: %v; using temp dir", artifactsDir, err)
		return t.TempDir()
	}
	probeName := probe.Name()
	if err := probe.Close(); err != nil {
		t.Logf("could not close write probe %s: %v; using temp dir", probeName, err)
		return t.TempDir()
	}
	if err := os.Remove(probeName); err != nil {
		t.Logf("could not remove write probe %s: %v", probeName, err)
	}
	return artifactsDir
}

// ----------------------------------------------------------------------------
// Data generators (used by parity examples and a few diagnostic tests)
// ----------------------------------------------------------------------------

// normalData generates a seeded normally-distributed sample using Box-Muller.
func normalData(seed1, seed2 uint64, n int, mean, stddev float64) []float64 {
	rng := rand.New(rand.NewPCG(seed1, seed2))
	data := make([]float64, n)
	for i := range data {
		u1 := rng.Float64()
		u2 := rng.Float64()
		data[i] = math.Sqrt(-2*math.Log(u1))*math.Cos(2*math.Pi*u2)*stddev + mean
	}
	return data
}

// ----------------------------------------------------------------------------
// Catalog convenience
// ----------------------------------------------------------------------------

// allCases returns every catalog case (cheap; the catalog is small).
func allCases() []examplecatalog.Case { return examplecatalog.Cases() }

func rendererNeutralCases() []examplecatalog.Case {
	all := allCases()
	out := make([]examplecatalog.Case, 0, len(all))
	for _, c := range all {
		if c.NativeBackend == "" {
			out = append(out, c)
		}
	}
	return out
}

func aggNativeCases() []examplecatalog.Case {
	all := allCases()
	out := make([]examplecatalog.Case, 0, len(all))
	for _, c := range all {
		if c.NativeBackend == "agg" {
			out = append(out, c)
		}
	}
	return out
}

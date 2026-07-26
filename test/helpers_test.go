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
	"strings"
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
	mplRefDir = "../testdata/matplotlib_ref"
	// mplMaxRMSE is the loose "did we render the right picture at all" floor,
	// the RMSE equivalent (255/10^(dB/20)) of the 10 dB PSNR floor this check
	// used before Phase 3.1. Above it, output has diverged fundamentally rather
	// than in placement or antialiasing.
	mplMaxRMSE = 80.6

	referenceCompareTolerance = 1
	// There is no PSNR floor here: imagecmp derives PSNR from RMSE, so a floor
	// would only restate a MaxRMSE ceiling. The former default of 44 dB was
	// equivalent to RMSE 1.609 and was unreachable for 65 cases; it only ever
	// passed because the PSNR accumulator overflowed. See Phase 3.1 in PLAN.md.
	referenceCompareMaxMeanAbs = 2.50
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

// strictMplRefIDs routes hand-curated text/title cases to the tight
// PSNR/MeanAbs strict check instead of the loose runMplTest floor.
//
// Historical note (Phase 18, 2026-07-01): these two plus 49 golden cases used
// to skip in default CI behind RUN_OPTIONAL_VISUAL_TESTS=true. The gate
// predated the vendored FreeType 2.6.1 default (glyph rasterization now
// byte-matches the references) and each render costs ~0.05 s, so every case
// now runs unconditionally — a live-render regression can no longer hide
// behind a skipped test.
var strictMplRefIDs = map[string]bool{
	"text_labels_strict": true,
	"title_strict":       true,
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

// Matplotlib pins these versions for every committed reference image. The AGG
// backend links the same FreeType, so the golden/reference corpus is only
// self-consistent when references are generated with this exact toolchain.
// Generating with any other matplotlib/FreeType silently rewrites the PNGs with
// different text metrics and hinting (e.g. 3.6.3/2.13.2 flips ~0.5% of pixels
// even on a bare line plot), so the version guard below refuses to proceed.
const (
	pinnedMatplotlibVersion = "3.10.9"
	pinnedFreeTypeVersion   = "2.6.1"
)

// pythonMatplotlibVersions reports the matplotlib and FreeType versions an
// interpreter would render with, or an error if it lacks a usable matplotlib.
func pythonMatplotlibVersions(py string) (mplVersion, freetypeVersion string, err error) {
	const probe = "import matplotlib, matplotlib.ft2font as f; print(matplotlib.__version__); print(f.__freetype_version__)"
	out, err := exec.Command(py, "-c", probe).Output()
	if err != nil {
		return "", "", err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return "", "", fmt.Errorf("unexpected version probe output: %q", string(out))
	}
	return strings.TrimSpace(lines[0]), strings.TrimSpace(lines[1]), nil
}

// matplotlibPythonPath finds an interpreter whose matplotlib/FreeType match the
// pinned reference toolchain. It prefers the PATH python3 over the system
// /usr/bin/python3 (which on some hosts ships an older matplotlib) and, crucially,
// refuses to return a version-mismatched interpreter — a wrong toolchain would
// corrupt every regenerated reference image rather than fail loudly.
func matplotlibPythonPath() (string, error) {
	candidates := []string{}
	if env := os.Getenv("MATPLOTLIB_GO_PYTHON"); env != "" {
		candidates = append(candidates, env)
	}
	if pyPath, err := exec.LookPath("python3"); err == nil {
		candidates = append(candidates, pyPath)
	}
	candidates = append(candidates, "/usr/bin/python3")

	seen := map[string]bool{}
	var report []string
	for _, candidate := range candidates {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		mplVer, ftVer, err := pythonMatplotlibVersions(candidate)
		if err != nil {
			report = append(report, fmt.Sprintf("  %s: no usable matplotlib (%v)", candidate, err))
			continue
		}
		if mplVer == pinnedMatplotlibVersion && ftVer == pinnedFreeTypeVersion {
			return candidate, nil
		}
		report = append(report, fmt.Sprintf("  %s: matplotlib %s / FreeType %s", candidate, mplVer, ftVer))
	}
	return "", fmt.Errorf(
		"no Python interpreter with the pinned matplotlib %s / FreeType %s found; "+
			"generating references with a different toolchain corrupts testdata/matplotlib_ref/*.png.\n"+
			"Checked:\n%s\n"+
			"Set MATPLOTLIB_GO_PYTHON to an interpreter with the pinned versions.",
		pinnedMatplotlibVersion, pinnedFreeTypeVersion, strings.Join(report, "\n"),
	)
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
// writes artefacts for inspection, and fails only when RMSE exceeds
// mplMaxRMSE (a fundamental rendering error).
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

	t.Logf("PSNR=%.1f dB  MeanAbs=%.2f  RMSE=%.2f  MaxDiff=%d", diff.PSNR, diff.MeanAbs, diff.RMSE, diff.MaxDiff)

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

	if diff.RMSE > mplMaxRMSE {
		t.Errorf("RMSE %.1f > %.1f: likely fundamental rendering mismatch with matplotlib",
			diff.RMSE, mplMaxRMSE)
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

	// The live render must byte-match the committed golden (≤1 LSB) so this
	// test alone catches a live regression — the golden-vs-reference compare
	// below only relates two committed files. Skipped under -update-golden,
	// where goldens are being rewritten in the same run.
	if !*updateGolden {
		liveDiff, err := imagecmp.ComparePNG(got, golden, 1)
		if err != nil {
			t.Fatalf("compare live render and golden %s: %v", goldenPath, err)
		}
		if !liveDiff.Identical {
			liveDiffPath := filepath.Join(artifactsDir, c.ID+"_rendered_vs_golden_diff.png")
			if err := imagecmp.SaveDiffImage(got, golden, 1, liveDiffPath); err != nil {
				t.Fatalf("save diff image %s: %v", liveDiffPath, err)
			}
			t.Fatalf("live render diverges from golden %s: MaxDiff=%d MeanAbs=%.2f RMSE=%.2f (diff: %s)",
				goldenPath, liveDiff.MaxDiff, liveDiff.MeanAbs, liveDiff.RMSE, liveDiffPath)
		}
	}

	diff, err := imagecmp.ComparePNG(golden, matplotlibRef, referenceCompareTolerance)
	if err != nil {
		t.Fatalf("compare %s and %s: %v", goldenPath, matplotlibPath, err)
	}

	diffPath := filepath.Join(artifactsDir, c.ID+"_golden_vs_matplotlib_ref_diff.png")
	if err := imagecmp.SaveDiffImage(golden, matplotlibRef, referenceCompareTolerance, diffPath); err != nil {
		t.Fatalf("save diff image %s: %v", diffPath, err)
	}

	t.Logf("%s: MaxDiff=%d MeanAbs=%.2f RMSE=%.2f PSNR=%.2fdB DiffPixels=%d Clusters=%d LargestCluster=%d",
		c.ID, diff.MaxDiff, diff.MeanAbs, diff.RMSE, diff.PSNR,
		diff.DiffPixels, diff.Clusters, diff.LargestCluster)

	maxMeanAbs := referenceCompareMaxMeanAbs
	if c.MaxMeanAbs > 0 {
		maxMeanAbs = c.MaxMeanAbs
	}
	if diff.MeanAbs > maxMeanAbs || (c.MaxRMSE > 0 && diff.RMSE > c.MaxRMSE) {
		t.Fatalf("reference mismatch for %s: MeanAbs=%.2f (max %.2f), RMSE=%.2f (max %.2f)",
			c.ID, diff.MeanAbs, maxMeanAbs, diff.RMSE, c.MaxRMSE)
	}

	// Residual-shape gate. The amplitude gates above are whole-image averages
	// and cannot see a small, dense, high-amplitude residual — a glyph drawn in
	// the wrong place. These bounds can.
	if c.MaxDiffPixels > 0 && diff.DiffPixels > c.MaxDiffPixels {
		t.Fatalf("reference residual spread for %s: %d differing pixels (max %d), largest cluster %d (diff: %s)",
			c.ID, diff.DiffPixels, c.MaxDiffPixels, diff.LargestCluster, diffPath)
	}
	if c.MaxLargestCluster > 0 && diff.LargestCluster > c.MaxLargestCluster {
		t.Fatalf("reference residual cluster for %s: largest cluster %d px (max %d) across %d differing pixels (diff: %s)",
			c.ID, diff.LargestCluster, c.MaxLargestCluster, diff.DiffPixels, diffPath)
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

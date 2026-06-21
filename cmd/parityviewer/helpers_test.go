package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRerenderHelpersAndEnv(t *testing.T) {
	if err := ensureRerenderSupported(false, "/repo", "/repo/testdata/golden"); err != nil {
		t.Fatalf("expected rerender to be allowed, got %v", err)
	}
	if err := ensureRerenderSupported(true, "/repo", "/repo/testdata/golden"); err == nil || !strings.Contains(err.Error(), "--parity-dir") {
		t.Fatalf("expected parity-mode rerender error, got %v", err)
	}
	if err := rerenderArtifact("/repo", "   "); err == nil || !strings.Contains(err.Error(), "missing name") {
		t.Fatalf("expected empty rerenderArtifact name to fail, got %v", err)
	}
	if !samePath("/repo/testdata/golden/../golden", "/repo/testdata/golden") {
		t.Fatal("samePath should treat cleaned equivalent paths as equal")
	}
	if got := testNameFromCaseName("axes_top_right_inverted"); got != "^TestGolden/axes_top_right_inverted$" {
		t.Fatalf("testNameFromCaseName returned %q", got)
	}

	t.Setenv("PORT", " 9090 ")
	if got := envOr("PORT", "8090"); got != "9090" {
		t.Fatalf("envOr should trim and return env value, got %q", got)
	}
	t.Setenv("EMPTY_VAR", "   ")
	if got := envOr("EMPTY_VAR", "fallback"); got != "fallback" {
		t.Fatalf("envOr should fall back for blank values, got %q", got)
	}
	if got := envOr("MISSING_VAR", "fallback"); got != "fallback" {
		t.Fatalf("envOr should fall back for missing values, got %q", got)
	}
}

func TestCompareImagesAndDiffHelpers(t *testing.T) {
	ref := image.NewRGBA(image.Rect(0, 0, 2, 1))
	act := image.NewRGBA(image.Rect(0, 0, 2, 1))
	ref.SetRGBA(0, 0, color.RGBA{R: 1, G: 2, B: 3, A: 10})
	act.SetRGBA(0, 0, color.RGBA{R: 1, G: 2, B: 3, A: 10})
	ref.SetRGBA(1, 0, color.RGBA{R: 10, G: 20, B: 30, A: 40})
	act.SetRGBA(1, 0, color.RGBA{R: 20, G: 30, B: 40, A: 40})

	stats := compareImages(ref, act)
	if stats.TotalPixels != 2 {
		t.Fatalf("TotalPixels = %d, want 2", stats.TotalPixels)
	}
	if stats.DiffPixels != 1 {
		t.Fatalf("DiffPixels = %d, want 1", stats.DiffPixels)
	}
	if stats.MaxDiff != 10 {
		t.Fatalf("MaxDiff = %d, want 10", stats.MaxDiff)
	}
	if math.Abs(stats.AvgDiff-3.75) > 1e-9 {
		t.Fatalf("AvgDiff = %v, want 3.75", stats.AvgDiff)
	}
	if math.Abs(stats.RMSE-math.Sqrt(37.5)) > 1e-9 {
		t.Fatalf("RMSE = %v, want %v", stats.RMSE, math.Sqrt(37.5))
	}
	if math.Abs(stats.DiffRatio-0.5) > 1e-9 {
		t.Fatalf("DiffRatio = %v, want 0.5", stats.DiffRatio)
	}

	raw := rawDiffImage(ref, act)
	if got := raw.RGBAAt(0, 0); got != (color.RGBA{R: 0, G: 0xaa, B: 0, A: 255}) {
		t.Fatalf("raw identical pixel = %+v", got)
	}
	if got := raw.RGBAAt(1, 0); got != (color.RGBA{R: 10, G: 10, B: 10, A: 255}) {
		t.Fatalf("raw diff pixel = %+v", got)
	}

	amp := amplifiedDiffImage(ref, act)
	if got := amp.RGBAAt(0, 0); got != (color.RGBA{R: 0, G: 0xaa, B: 0, A: 255}) {
		t.Fatalf("amplified identical pixel = %+v", got)
	}
	if got := amp.RGBAAt(1, 0); got != (color.RGBA{R: 10, G: 0, B: 0, A: 255}) {
		t.Fatalf("amplified diff pixel = %+v", got)
	}

	if r, g, b, a := rgbaAt(ref, 1, 0); r != 10 || g != 20 || b != 30 || a != 40 {
		t.Fatalf("rgbaAt returned %d,%d,%d,%d", r, g, b, a)
	}
	if r, g, b, a := rgbaAt(nil, 0, 0); r != 0 || g != 0 || b != 0 || a != 0 {
		t.Fatalf("rgbaAt(nil) returned %d,%d,%d,%d", r, g, b, a)
	}
	if got := unionBounds(image.Rect(2, 2, 4, 4), image.Rect(0, 1, 3, 3)); got != image.Rect(0, 1, 4, 4) {
		t.Fatalf("unionBounds = %v", got)
	}
	if got := max4(1, 9, 3, 7); got != 9 {
		t.Fatalf("max4 = %d, want 9", got)
	}
	if got := absDiff8(3, 10); got != 7 {
		t.Fatalf("absDiff8 = %d, want 7", got)
	}
	if got := sqDiff(3, 10); got != 49 {
		t.Fatalf("sqDiff = %v, want 49", got)
	}
	if got := clampAlpha(0); got != 255 {
		t.Fatalf("clampAlpha(0) = %d, want 255", got)
	}
	if got := clampAlpha(12); got != 12 {
		t.Fatalf("clampAlpha(12) = %d, want 12", got)
	}
	if got := min(3, 10); got != 3 {
		t.Fatalf("min = %d, want 3", got)
	}
	if got := max(3, 10); got != 10 {
		t.Fatalf("max = %d, want 10", got)
	}
}

func TestCompositeAndBadgeHelpers(t *testing.T) {
	src := image.NewRGBA(image.Rect(5, 7, 7, 8))
	src.SetRGBA(5, 7, color.RGBA{R: 255, G: 0, B: 0, A: 128})
	src.SetRGBA(6, 7, color.RGBA{R: 0, G: 255, B: 0, A: 255})

	composited := compositeOverSolid(src, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	if composited.Bounds() != image.Rect(0, 0, 2, 1) {
		t.Fatalf("compositeOverSolid bounds = %v", composited.Bounds())
	}
	if got := composited.RGBAAt(0, 0); got.A != 255 || got.R == 0 {
		t.Fatalf("unexpected blended pixel %+v", got)
	}
	if got := compositePixel(255, 0, 0, 128, color.RGBA{R: 10, G: 20, B: 30, A: 255}); got.A != 255 {
		t.Fatalf("compositePixel alpha = %d, want 255", got.A)
	}
	if compositeOverSolid(nil, color.RGBA{}) != nil {
		t.Fatal("compositeOverSolid(nil) should return nil")
	}

	if got := badgeClassRMSE(5); got != badgeClassOK {
		t.Fatalf("badgeClassRMSE(5) = %q", got)
	}
	if got := badgeClassRMSE(6); got != badgeClassWarn {
		t.Fatalf("badgeClassRMSE(6) = %q", got)
	}
	if got := badgeClassRMSE(21); got != badgeClassBad {
		t.Fatalf("badgeClassRMSE(21) = %q", got)
	}
	if got := badgeClassAvgDiff(2); got != badgeClassOK {
		t.Fatalf("badgeClassAvgDiff(2) = %q", got)
	}
	if got := badgeClassAvgDiff(3); got != badgeClassWarn {
		t.Fatalf("badgeClassAvgDiff(3) = %q", got)
	}
	if got := badgeClassAvgDiff(9); got != badgeClassBad {
		t.Fatalf("badgeClassAvgDiff(9) = %q", got)
	}
	if got := badgeClassMaxDiff(10); got != badgeClassOK {
		t.Fatalf("badgeClassMaxDiff(10) = %q", got)
	}
	if got := badgeClassMaxDiff(20); got != badgeClassWarn {
		t.Fatalf("badgeClassMaxDiff(20) = %q", got)
	}
	if got := badgeClassMaxDiff(41); got != badgeClassBad {
		t.Fatalf("badgeClassMaxDiff(41) = %q", got)
	}
	if got := badgeClassDiffRatio(0.01); got != badgeClassOK {
		t.Fatalf("badgeClassDiffRatio(0.01) = %q", got)
	}
	if got := badgeClassDiffRatio(0.02); got != badgeClassWarn {
		t.Fatalf("badgeClassDiffRatio(0.02) = %q", got)
	}
	if got := badgeClassDiffRatio(0.06); got != badgeClassBad {
		t.Fatalf("badgeClassDiffRatio(0.06) = %q", got)
	}
}

func TestLoadCasesFromDirectories(t *testing.T) {
	baseDir := t.TempDir()
	artifactDir := t.TempDir()

	writePNG(t, filepath.Join(baseDir, "zeta.png"), solidRGBA(2, 2, color.RGBA{R: 0, G: 0, B: 0, A: 255}))
	writePNG(t, filepath.Join(baseDir, "alpha.png"), solidRGBA(2, 2, color.RGBA{R: 50, G: 50, B: 50, A: 255}))
	writePNG(t, filepath.Join(baseDir, "missing.png"), solidRGBA(2, 2, color.RGBA{R: 5, G: 5, B: 5, A: 255}))

	writePNG(t, filepath.Join(artifactDir, "zeta.png"), solidRGBA(2, 2, color.RGBA{R: 0, G: 0, B: 0, A: 255}))
	writePNG(t, filepath.Join(artifactDir, "alpha.png"), solidRGBA(2, 2, color.RGBA{R: 200, G: 200, B: 200, A: 255}))

	result, err := loadCasesFromDirectories(baseDir, artifactDir, "", "")
	if err != nil {
		t.Fatalf("loadCasesFromDirectories failed: %v", err)
	}
	if result.ComparedCount != 2 {
		t.Fatalf("ComparedCount = %d, want 2", result.ComparedCount)
	}
	if result.SkippedCount != 1 {
		t.Fatalf("SkippedCount = %d, want 1", result.SkippedCount)
	}
	if len(result.Cases) != 2 {
		t.Fatalf("len(Cases) = %d, want 2", len(result.Cases))
	}
	if result.Cases[0].Name != "alpha" || result.Cases[1].Name != "zeta" {
		t.Fatalf("cases not sorted by descending RMSE then name: %+v", result.Cases)
	}
	if result.Cases[0].RefWidth != 2 || result.Cases[0].ActHeight != 2 {
		t.Fatalf("unexpected case dimensions: %+v", result.Cases[0])
	}
	if result.Cases[0].RefImageURL == "" || result.Cases[0].ActImageURL == "" || result.Cases[0].RawDiffURL == "" || result.Cases[0].AmpDiffURL == "" {
		t.Fatalf("expected lazy image URLs to be populated: %+v", result.Cases[0])
	}

	filtered, err := loadCasesFromDirectories(baseDir, artifactDir, "ta", "ze")
	if err != nil {
		t.Fatalf("filtered loadCasesFromDirectories failed: %v", err)
	}
	if filtered.ComparedCount != 1 || len(filtered.Cases) != 1 || filtered.Cases[0].Name != "zeta" {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
}

func solidRGBA(width, height int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

package test

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"matplotlib-go/test/imagecmp"
)

// TestArtifactGeneration verifies that our test infrastructure
// properly creates debug artifacts when tests fail.
// This test is normally skipped to avoid CI failures.
func TestArtifactGeneration(t *testing.T) {
	if os.Getenv("GENERATE_ARTIFACTS") != "true" {
		t.Skip("Skipping artifact generation test (set GENERATE_ARTIFACTS=true to run)")
	}

	// Create two different images to simulate a golden test failure
	img1 := createTestImage(100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255}) // Red
	img2 := createTestImage(100, 100, color.RGBA{R: 0, G: 255, B: 0, A: 255}) // Green

	// Compare them (this will fail)
	diff, err := imagecmp.ComparePNG(img1, img2, 1)
	if err != nil {
		t.Fatalf("ComparePNG failed: %v", err)
	}

	if diff.Identical {
		t.Fatal("Expected images to be different for artifact test")
	}

	// Generate artifacts like a real golden test would
	artifactsDir := "../_artifacts"
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Fatalf("Could not create artifacts directory: %v", err)
	}

	// Save the "got" image
	gotPath := filepath.Join(artifactsDir, "test_got.png")
	if err := imagecmp.SavePNG(img1, gotPath); err != nil {
		t.Fatalf("Could not save got image: %v", err)
	}

	// Save the "want" image
	wantPath := filepath.Join(artifactsDir, "test_want.png")
	if err := imagecmp.SavePNG(img2, wantPath); err != nil {
		t.Fatalf("Could not save want image: %v", err)
	}

	// Save the diff image
	diffPath := filepath.Join(artifactsDir, "test_diff.png")
	if err := imagecmp.SaveDiffImage(img1, img2, 1, diffPath); err != nil {
		t.Fatalf("Could not save diff image: %v", err)
	}

	t.Logf("Generated test artifacts in %s/", artifactsDir)
	t.Logf("MaxDiff=%d, MeanAbs=%.2f, PSNR=%.2fdB", diff.MaxDiff, diff.MeanAbs, diff.PSNR)
}

func createTestImage(width, height int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

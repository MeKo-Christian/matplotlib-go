package imagecmp

import (
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestComparePNG_Identical(t *testing.T) {
	// Create two identical 10x10 red images
	img1 := createSolidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img2 := createSolidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	result, err := ComparePNG(img1, img2, 0)
	if err != nil {
		t.Fatalf("ComparePNG failed: %v", err)
	}

	if !result.Identical {
		t.Error("Expected identical images to be reported as identical")
	}

	if result.MaxDiff != 0 {
		t.Errorf("Expected MaxDiff=0 for identical images, got %d", result.MaxDiff)
	}

	if result.MeanAbs != 0 {
		t.Errorf("Expected MeanAbs=0 for identical images, got %f", result.MeanAbs)
	}

	if !math.IsInf(result.PSNR, 1) {
		t.Errorf("Expected PSNR=+Inf for identical images, got %f", result.PSNR)
	}
}

func TestComparePNG_DifferentSizes(t *testing.T) {
	img1 := createSolidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img2 := createSolidImage(5, 5, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	_, err := ComparePNG(img1, img2, 0)
	if err == nil {
		t.Error("Expected error when comparing images of different sizes")
	}
}

func TestComparePNG_SinglePixelDiff(t *testing.T) {
	// Create two 10x10 images, second has one pixel different
	img1 := createSolidImage(10, 10, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	img2 := createSolidImage(10, 10, color.RGBA{R: 100, G: 100, B: 100, A: 255})

	// Modify one pixel in img2
	img2.Set(5, 5, color.RGBA{R: 105, G: 100, B: 100, A: 255}) // +5 difference in red

	result, err := ComparePNG(img1, img2, 1)
	if err != nil {
		t.Fatalf("ComparePNG failed: %v", err)
	}

	if result.MaxDiff != 5 {
		t.Errorf("Expected MaxDiff=5, got %d", result.MaxDiff)
	}

	if result.Identical {
		t.Error("Expected images with difference > tolerance to not be identical")
	}

	// Test with tolerance that should pass
	result, err = ComparePNG(img1, img2, 5)
	if err != nil {
		t.Fatalf("ComparePNG failed: %v", err)
	}

	if !result.Identical {
		t.Error("Expected images within tolerance to be considered identical")
	}
}

func TestComparePNG_GradientImages(t *testing.T) {
	// Create gradient images to test PSNR calculation
	img1 := createGradientImage(100, 100)
	img2 := createNoisyGradientImage(100, 100, 10) // Add noise level 10

	result, err := ComparePNG(img1, img2, 255) // High tolerance to focus on PSNR
	if err != nil {
		t.Fatalf("ComparePNG failed: %v", err)
	}

	// PSNR should be finite and reasonable (not infinite, not too low)
	if math.IsInf(result.PSNR, 1) || result.PSNR < 10 || result.PSNR > 50 {
		t.Errorf("Expected reasonable PSNR value, got %f", result.PSNR)
	}

	if result.MeanAbs <= 0 {
		t.Error("Expected non-zero mean absolute difference for different images")
	}
}

func TestLoadPNG_NonExistentFile(t *testing.T) {
	_, err := LoadPNG("nonexistent_file.png")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestSavePNG_RoundTrip(t *testing.T) {
	// Create test directory
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test.png")

	// Create and save image
	originalImg := createSolidImage(20, 20, color.RGBA{R: 128, G: 64, B: 192, A: 255})

	err := SavePNG(originalImg, testPath)
	if err != nil {
		t.Fatalf("SavePNG failed: %v", err)
	}

	// Load it back
	loadedImg, err := LoadPNG(testPath)
	if err != nil {
		t.Fatalf("LoadPNG failed: %v", err)
	}

	// Compare
	result, err := ComparePNG(originalImg, loadedImg, 1) // Allow for minor encoding differences
	if err != nil {
		t.Fatalf("ComparePNG failed: %v", err)
	}

	if !result.Identical {
		t.Errorf("Round-trip PNG save/load changed image: MaxDiff=%d", result.MaxDiff)
	}
}

func TestHashPNG_Deterministic(t *testing.T) {
	img := createSolidImage(10, 10, color.RGBA{R: 123, G: 45, B: 67, A: 255})

	// Hash multiple times
	hash1 := HashPNG(img)
	hash2 := HashPNG(img)

	if hash1 != hash2 {
		t.Error("HashPNG should be deterministic")
	}

	// Hash should be different for different images
	img2 := createSolidImage(10, 10, color.RGBA{R: 124, G: 45, B: 67, A: 255})
	hash3 := HashPNG(img2)

	if hash1 == hash3 {
		t.Error("Different images should have different hashes")
	}

	// Check hash format (should be 64 hex characters for SHA256)
	if len(hash1) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}
}

func TestSaveDiffImage(t *testing.T) {
	tempDir := t.TempDir()
	diffPath := filepath.Join(tempDir, "diff.png")

	// Create two slightly different images
	img1 := createSolidImage(10, 10, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	img2 := createSolidImage(10, 10, color.RGBA{R: 100, G: 100, B: 100, A: 255})

	// Modify a few pixels in img2
	img2.Set(2, 2, color.RGBA{R: 120, G: 100, B: 100, A: 255}) // Above threshold
	img2.Set(3, 3, color.RGBA{R: 101, G: 100, B: 100, A: 255}) // Below threshold

	err := SaveDiffImage(img1, img2, 5, diffPath)
	if err != nil {
		t.Fatalf("SaveDiffImage failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(diffPath); os.IsNotExist(err) {
		t.Error("Diff image file was not created")
	}

	// Load and verify the diff image
	diffImg, err := LoadPNG(diffPath)
	if err != nil {
		t.Fatalf("Failed to load diff image: %v", err)
	}

	// Check that the pixel with large difference is highlighted (red)
	diffColor := color.RGBAModel.Convert(diffImg.At(2, 2)).(color.RGBA)
	if diffColor.R != 255 || diffColor.G != 0 || diffColor.B != 0 {
		t.Errorf("Expected red highlight at (2,2), got %+v", diffColor)
	}

	// Check that the pixel with small difference is not highlighted
	sameColor := color.RGBAModel.Convert(diffImg.At(3, 3)).(color.RGBA)
	if sameColor.R == 255 && sameColor.G == 0 && sameColor.B == 0 {
		t.Error("Expected pixel within tolerance to not be highlighted red")
	}
}

// Helper functions for creating test images

func createSolidImage(width, height int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func createGradientImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a simple gradient
			intensity := uint8((x + y) * 255 / (width + height - 2))
			img.Set(x, y, color.RGBA{R: intensity, G: intensity, B: intensity, A: 255})
		}
	}
	return img
}

func createNoisyGradientImage(width, height int, noiseLevel uint8) *image.RGBA {
	img := createGradientImage(width, height)

	// Add some deterministic "noise" by modifying every other pixel
	for y := 0; y < height; y += 2 {
		for x := 0; x < width; x += 2 {
			original := img.RGBAAt(x, y)
			noisy := color.RGBA{
				R: clampAdd(original.R, noiseLevel),
				G: clampAdd(original.G, noiseLevel),
				B: clampAdd(original.B, noiseLevel),
				A: original.A,
			}
			img.Set(x, y, noisy)
		}
	}
	return img
}

func clampAdd(original, delta uint8) uint8 {
	if original > 255-delta {
		return 255
	}
	return original + delta
}

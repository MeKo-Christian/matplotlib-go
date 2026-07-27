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
	if result.RMSE != 0.25 {
		t.Errorf("Expected RMSE=0.25, got %f", result.RMSE)
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

// TestComparePNG_LargeDiffDoesNotOverflowPSNR guards the fix: the
// PSNR accumulator used to square uint8 differences in uint8 arithmetic, so any
// per-channel difference above 15 wrapped mod 256 and reported a PSNR far better
// than the actual error. A single black-vs-white pixel is the extreme case:
// 255 squared wraps to 1, understating the error by more than 40 dB.
func TestComparePNG_LargeDiffDoesNotOverflowPSNR(t *testing.T) {
	for _, tc := range []struct {
		name string
		diff uint8
	}{
		{"just below the overflow threshold", 15},
		{"just above the overflow threshold", 16},
		{"mid range", 200},
		{"full scale", 255},
	} {
		t.Run(tc.name, func(t *testing.T) {
			img1 := createSolidImage(10, 10, color.RGBA{A: 255})
			img2 := createSolidImage(10, 10, color.RGBA{A: 255})
			img2.Set(5, 5, color.RGBA{R: tc.diff, A: 255})

			result, err := ComparePNG(img1, img2, 1)
			if err != nil {
				t.Fatalf("ComparePNG failed: %v", err)
			}

			// One channel of one pixel out of 100 differs, and squared error is
			// averaged over the four channels: MSE = diff^2 / (4*100).
			wantRMSE := math.Sqrt(float64(tc.diff) * float64(tc.diff) / 400)
			if math.Abs(result.RMSE-wantRMSE) > 1e-9 {
				t.Errorf("RMSE = %v, want %v", result.RMSE, wantRMSE)
			}

			wantPSNR := 20 * math.Log10(255/wantRMSE)
			if math.Abs(result.PSNR-wantPSNR) > 1e-9 {
				t.Errorf("PSNR = %v dB, want %v dB", result.PSNR, wantPSNR)
			}
		})
	}
}

// TestComparePNG_PSNRIsDerivedFromRMSE pins the identity the catalog tolerances
// rely on: PSNR is a monotone restatement of RMSE, not an independent gate, so a
// MinPSNR floor can only ever duplicate a MaxRMSE ceiling.
func TestComparePNG_PSNRIsDerivedFromRMSE(t *testing.T) {
	img1 := createGradientImage(64, 64)
	img2 := createNoisyGradientImage(64, 64, 40)

	result, err := ComparePNG(img1, img2, 255)
	if err != nil {
		t.Fatalf("ComparePNG failed: %v", err)
	}
	if result.RMSE <= 0 {
		t.Fatalf("expected a non-zero RMSE, got %v", result.RMSE)
	}

	want := 20 * math.Log10(255/result.RMSE)
	if math.Abs(result.PSNR-want) > 1e-9 {
		t.Errorf("PSNR = %v dB, want 20*log10(255/RMSE) = %v dB", result.PSNR, want)
	}
}

// TestComparePNG_ResidualShape pins the localization metrics added in Phase
// 3.2: how many pixels differ, how many 8-connected clusters they form, and how
// large the biggest one is.
func TestComparePNG_ResidualShape(t *testing.T) {
	tests := []struct {
		name           string
		paint          func(img *image.RGBA)
		diffPixels     int
		clusters       int
		largestCluster int
	}{
		{
			name:  "identical images have no residual",
			paint: func(*image.RGBA) {},
		},
		{
			name:           "one pixel is one cluster",
			paint:          func(img *image.RGBA) { img.Set(5, 5, color.RGBA{R: 200, A: 255}) },
			diffPixels:     1,
			clusters:       1,
			largestCluster: 1,
		},
		{
			name: "diagonal neighbours join under 8-connectivity",
			paint: func(img *image.RGBA) {
				img.Set(2, 2, color.RGBA{R: 200, A: 255})
				img.Set(3, 3, color.RGBA{R: 200, A: 255})
			},
			diffPixels:     2,
			clusters:       1,
			largestCluster: 2,
		},
		{
			name: "separated pixels stay separate clusters",
			paint: func(img *image.RGBA) {
				img.Set(1, 1, color.RGBA{R: 200, A: 255})
				img.Set(8, 8, color.RGBA{R: 200, A: 255})
				img.Set(1, 8, color.RGBA{R: 200, A: 255})
			},
			diffPixels:     3,
			clusters:       3,
			largestCluster: 1,
		},
		{
			name: "largest cluster wins over the pixel count",
			paint: func(img *image.RGBA) {
				for y := 4; y < 8; y++ {
					for x := 4; x < 7; x++ {
						img.Set(x, y, color.RGBA{R: 200, A: 255})
					}
				}
				img.Set(0, 0, color.RGBA{R: 200, A: 255})
			},
			diffPixels:     13,
			clusters:       2,
			largestCluster: 12,
		},
		{
			name:  "differences within tolerance are not counted",
			paint: func(img *image.RGBA) { img.Set(5, 5, color.RGBA{R: 1, A: 255}) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			img1 := createSolidImage(10, 10, color.RGBA{A: 255})
			img2 := createSolidImage(10, 10, color.RGBA{A: 255})
			tc.paint(img2)

			result, err := ComparePNG(img1, img2, 1)
			if err != nil {
				t.Fatalf("ComparePNG failed: %v", err)
			}
			if result.DiffPixels != tc.diffPixels {
				t.Errorf("DiffPixels = %d, want %d", result.DiffPixels, tc.diffPixels)
			}
			if result.Clusters != tc.clusters {
				t.Errorf("Clusters = %d, want %d", result.Clusters, tc.clusters)
			}
			if result.LargestCluster != tc.largestCluster {
				t.Errorf("LargestCluster = %d, want %d", result.LargestCluster, tc.largestCluster)
			}
		})
	}
}

// TestComparePNG_ShapeSeesWhatRMSECannot is the reason exists: a
// small, dense, maximum-amplitude residual — a glyph in the wrong place —
// disappears into a whole-image average but is obvious as one cluster.
func TestComparePNG_ShapeSeesWhatRMSECannot(t *testing.T) {
	img1 := createSolidImage(640, 360, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	img2 := createSolidImage(640, 360, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	for y := 100; y < 106; y++ {
		for x := 200; x < 206; x++ {
			img2.Set(x, y, color.RGBA{A: 255})
		}
	}

	result, err := ComparePNG(img1, img2, 1)
	if err != nil {
		t.Fatalf("ComparePNG failed: %v", err)
	}
	// 2.8 is basic_line's real MaxRMSE, and 36 black pixels on a 640x360 frame
	// slip under it just as its misplaced tick label does.
	if result.RMSE > 2.8 {
		t.Fatalf("RMSE = %v, expected the misplaced block to hide under basic_line's bound", result.RMSE)
	}
	if result.LargestCluster != 36 || result.Clusters != 1 {
		t.Errorf("Clusters = %d, LargestCluster = %d, want 1 and 36", result.Clusters, result.LargestCluster)
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

# imagecmp - Image Comparison for Golden Tests

The `imagecmp` package provides utilities for comparing images in tests, specifically designed for golden image testing in matplotlib-go.

## Overview

Golden image testing is a technique where we generate reference images (golden images) and compare new rendering output against these references to detect visual regressions. This package provides pixel-perfect comparison with configurable tolerances.

## Core Functions

### `ComparePNG(got, want image.Image, tolerance uint8) (DiffResult, error)`

Compares two images pixel-by-pixel and returns detailed metrics:

- **MaxDiff**: Maximum per-channel difference found
- **MeanAbs**: Mean absolute difference across all channels
- **PSNR**: Peak Signal-to-Noise Ratio in dB
- **Identical**: True if all pixels are within tolerance

The tolerance parameter allows for minor encoding differences (typically 1 for ≤1 LSB tolerance).

### `LoadPNG(path string) (image.Image, error)`

Loads a PNG image from the filesystem for comparison.

### `SavePNG(img image.Image, path string) error`

Saves an image as a PNG file.

### `HashPNG(img image.Image) string`

Computes a SHA256 hash of the image's raw RGBA data for deterministic CI assertions.

### `SaveDiffImage(got, want image.Image, threshold uint8, outputPath string) error`

Creates a visual diff image highlighting pixels that differ by more than the threshold in bright red.

## Usage in Tests

```go
func TestMyRenderer_Golden(t *testing.T) {
    // Render your content
    img := renderMyContent()

    // Load golden reference
    want, err := imagecmp.LoadPNG("testdata/golden/my_content.png")
    if err != nil {
        t.Fatalf("Failed to load golden image: %v", err)
    }

    // Compare with 1 LSB tolerance
    diff, err := imagecmp.ComparePNG(img, want, 1)
    if err != nil {
        t.Fatalf("Comparison failed: %v", err)
    }

    if !diff.Identical {
        // Save debug artifacts
        imagecmp.SavePNG(img, "_artifacts/got.png")
        imagecmp.SaveDiffImage(img, want, 1, "_artifacts/diff.png")

        t.Fatalf("Golden mismatch: MaxDiff=%d, PSNR=%.2fdB",
            diff.MaxDiff, diff.PSNR)
    }
}
```

## Updating Golden Images

Use the `-update-golden` flag to regenerate reference images:

```bash
go test -update-golden ./test/
```

## Deterministic Testing

The package is designed for cross-platform deterministic testing:

- Uses SHA256 hashing for exact binary comparison
- Handles premultiplied alpha consistently
- Avoids floating-point precision issues in comparison logic
- PSNR calculation provides quality metrics for CI monitoring

## Debug Artifacts

When tests fail, the package can generate debug artifacts:

- `*_got.png`: The actual rendered output
- `*_want.png`: The expected golden reference
- `*_diff.png`: Visual diff highlighting changed pixels

These artifacts are automatically uploaded by CI on test failures to aid debugging.

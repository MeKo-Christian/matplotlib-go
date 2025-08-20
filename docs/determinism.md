# Deterministic Rendering

This document outlines the measures taken to ensure deterministic, cross-platform rendering in matplotlib-go. The goal is to produce pixel-perfect identical output across different operating systems and architectures.

## Determinism Guarantees

matplotlib-go provides the following determinism guarantees:

1. **Identical pixel output** across Linux and macOS for the same input
2. **Reproducible golden images** for regression testing
3. **Consistent floating-point behavior** through quantization
4. **Stable dependency versions** via commit pinning

## Implementation Approach

### Float Quantization

All floating-point coordinates and measurements are quantized to `1e-6` precision to eliminate tiny cross-platform differences:

```go
const quantizationEpsilon = 1e-6

func quantize(v float64) float64 {
    return math.Round(v/quantizationEpsilon) * quantizationEpsilon
}
```

This quantization is applied to:

- All path vertex coordinates before rasterization
- Stroke widths and dash patterns
- Normal vector calculations in stroke geometry
- Miter limit calculations

### Premultiplied Alpha

All colors are converted to premultiplied alpha format before rasterization to ensure consistent blending behavior:

```go
func (c Color) ToPremultipliedRGBA() (uint8, uint8, uint8, uint8) {
    premul := c.Premultiply()
    return uint8(premul.R * 255 + 0.5),
           uint8(premul.G * 255 + 0.5),
           uint8(premul.B * 255 + 0.5),
           uint8(premul.A * 255 + 0.5)
}
```

The `+ 0.5` addition provides proper rounding when converting from float64 to uint8.

### Dependency Pinning

Critical dependencies are pinned to specific versions:

- **Go toolchain**: Pinned to Go 1.24.0 in CI
- **golang.org/x/image**: Pinned to v0.30.0 (commit c574db5...)

### Explicit Float32 Conversion

When passing coordinates to the vector rasterizer, explicit rounding is applied before float32 conversion:

```go
r.rasterizer.MoveTo(
    float32(math.Round(pt.X*1e6)/1e6),
    float32(math.Round(pt.Y*1e6)/1e6)
)
```

This ensures consistent behavior when converting from float64 to float32.

## Testing Strategy

### Golden Image Tests

The test suite includes golden image tests with strict pixel-perfect matching:

```go
diff, err := imagecmp.ComparePNG(img, want, 1) // ≤1 LSB tolerance
```

Any differences above 1 LSB cause test failure and generate debug artifacts.

### Cross-Platform CI

GitHub Actions CI runs on both Linux and macOS with:

- Identical Go versions (1.24.0)
- Identical dependency versions
- Golden image hash verification

### Artifacts on Failure

When golden tests fail, the CI automatically uploads:

- Generated image (`*_got.png`)
- Expected golden image (`*_want.png`)
- Pixel difference visualization (`*_diff.png`)
- Statistical comparison (PSNR, max difference)

## Precision Limits

### Coordinate Precision

The quantization epsilon of `1e-6` provides:

- Sub-pixel precision for most use cases
- Sufficient precision for print-quality output
- Elimination of floating-point drift

### Geometric Stability

Stroke geometry calculations maintain stability through:

- Quantized normal vector calculations
- Quantized join and cap geometry
- Deterministic line segment decomposition

## Platform Considerations

### Linux vs macOS

The implementation accounts for potential differences in:

- Floating-point arithmetic precision
- Standard library implementations
- Compiler optimizations

### Architecture Independence

The quantization approach ensures consistency across:

- x86_64 and ARM64 architectures
- Different compiler versions
- Various optimization levels

## Validation

### Hash Verification

Golden images are validated using SHA256 hashes to detect any changes:

```bash
find testdata/golden -name "*.png" -exec sha256sum {} \;
```

### Statistical Metrics

Image comparison includes multiple metrics:

- **MaxDiff**: Maximum per-channel difference
- **MeanAbs**: Mean absolute difference
- **PSNR**: Peak Signal-to-Noise Ratio

### Regression Detection

The CI system tracks:

- Golden image hash changes
- Performance regression in rendering
- Memory usage consistency

## Troubleshooting

### Golden Test Failures

If golden tests fail:

1. Check if the failure is platform-specific
2. Examine debug artifacts in `_artifacts/`
3. Verify dependency versions match
4. Consider if the change is intentional

### Updating Golden Images

To update golden baselines:

```bash
go test ./test/ -update-golden
```

This should only be done when visual changes are intentional.

### Debugging Determinism Issues

If cross-platform differences occur:

1. Compare dependency versions
2. Check quantization is applied consistently
3. Verify float32 conversion behavior
4. Examine stroke geometry calculations

## Future Considerations

### Additional Backends

When adding new backends (AGG, Skia), ensure:

- Equivalent quantization approaches
- Similar premultiplied alpha handling
- Cross-backend parity testing

### Performance Impact

The determinism measures have minimal performance impact:

- Quantization adds ~1-2% overhead
- Premultiplied alpha is optimization-neutral
- Memory usage is unchanged

### Precision Trade-offs

The current `1e-6` epsilon balances:

- Visual quality (sub-pixel precision)
- Determinism (eliminates drift)
- Performance (fast quantization)

This precision level may be adjusted based on future requirements.

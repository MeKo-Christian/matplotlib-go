# Matplotlib-Go (working title)

A plotting library for Go inspired by Matplotlib.  
Renderer-agnostic at the core, with support for high-quality raster and vector backends (AGG, Skia, etc.).

---

## Vision

**North-star:**  
Deliver a Go-native, Matplotlib-like plotting system with:

- **Familiar model:** `Figure → Axes → Artists` hierarchy
- **Renderer independence:** consistent outputs across CPU raster (AGG), GPU (Skia), and vector formats (SVG/PDF)
- **Deterministic results:** identical plots across machines and CI, great for testing
- **Beautiful text:** robust font handling, shaping (via HarfBuzz), and precise metrics
- **Comprehensive export:** PNG, SVG, PDF (and more via backends)
- **Go-idiomatic API:** options-based configuration, no hidden global state; optional `pyplot` shim for scripting
- **Cross-platform interactivity:** pan/zoom, picking, animations, WASM/web backends

---

## Constraints & Principles

- **Backend-agnostic core:** all plot logic independent of rendering technology
- **Determinism:** golden image tests, locked fonts, stable outputs
- **Minimal global state:** figures and axes are explicit values, not hidden globals
- **Extensibility:** artists, colormaps, and backends are pluggable
- **Quality-first:** correctness, readability, and sharp rendering over premature optimization
- **Interoperability:** ability to export or consume simple plot specifications (for testing or migration)

---

## Endgame

When this repo is “done”, it should provide:

- A stable core API for 2D plotting (lines, scatter, images, text, legends, colorbars, etc.)
- Multiple renderers (AGG, Skia, SVG, PDF) with visual parity
- A gallery of reproducible, high-quality examples
- Deterministic test suite with image baselines
- Documentation and guides, including **“Matplotlib to Go”** migration notes

---

## Testing

This project uses golden image testing to ensure visual consistency across platforms and detect rendering regressions.

### Running Tests

```bash
# Run all tests
make test

# Run only golden image tests
go test ./test/

# Update golden images when making intentional changes
go test ./test/ -update-golden
```

### Golden Image Testing

Golden tests compare rendered output against reference images stored in `testdata/golden/`. When tests fail, debug artifacts are saved to `_artifacts/` and uploaded by CI:

- `*_got.png`: Actual rendered output
- `*_want.png`: Expected golden reference
- `*_diff.png`: Visual diff highlighting changes

The comparison uses pixel-perfect RGBA matching with configurable tolerance (typically ±1 LSB) and reports PSNR metrics for quality assessment.

---

🚀 _Plotting for Go, without compromise._

# Matplotlib-Go (working title)

A plotting library for Go inspired by Matplotlib.  
Renderer-agnostic at the core, with support for high-quality raster backends today and vector/GPU backends later.

---

## Vision

**North-star:**  
Deliver a Go-native, Matplotlib-like plotting system with:

- **Familiar model:** `Figure → Axes → Artists` hierarchy
- **Renderer independence:** consistent outputs across CPU raster today, with room for GPU and vector backends later
- **Deterministic results:** identical plots across machines and CI, great for testing
- **Beautiful text:** robust font handling, fallback fonts, and precise metrics
- **Comprehensive export:** PNG today, with SVG/PDF planned via future backends
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
- Multiple renderers (AGG, GoBasic, SVG/PDF, and opt-in Skia) with visual parity
- A gallery of reproducible, high-quality examples
- Deterministic test suite with image baselines
- Documentation and guides, including **“Matplotlib to Go”** migration notes

---

## Testing

This project uses golden image testing to ensure visual consistency across platforms and detect rendering regressions.

### Running Tests

```bash
# Run all tests
just test

# Run only golden image tests
go test -tags freetype ./test/

# Update golden images when making intentional changes
go test -tags freetype ./test/ -update-golden
```

### Golden Image Testing

Golden tests compare rendered output against reference images stored in `testdata/golden/`. Matplotlib parity and text-layout checks use committed reference images from `testdata/matplotlib_ref/`. When tests fail, debug artifacts are saved to `testdata/_artifacts/` and uploaded by CI:

- `*_got.png`: Actual rendered output
- `*_want.png`: Expected golden reference
- `*_diff.png`: Visual diff highlighting changes

The comparison uses pixel-perfect RGBA matching with configurable tolerance (typically ±1 LSB) and reports PSNR metrics for quality assessment.

To refresh the committed Matplotlib reference images intentionally:

```bash
go test -tags freetype ./test/... -run TestMpl -update-matplotlib
```

## Web Demo

The repository now includes a browser demo under [`web/`](web) backed by Go
compiled to WebAssembly from [`cmd/wasm`](cmd/wasm).

Build the web artifact locally with:

```bash
just web-build
```

`web/main.wasm` and `web/wasm_exec.js` are generated build outputs. Re-run
`just web-build` before serving `web/`, especially after changes under
[`cmd/wasm`](cmd/wasm) or the browser bridge, or the page may report an
incompatible WASM build.

Then serve the `web/` directory with any static file server, for example:

```bash
python3 -m http.server 8000 --directory web
```

The same demo catalog can also be rendered directly to PNGs without serving the
HTML page:

```bash
go run -tags freetype ./cmd/webdemoexport --output-dir testdata/_artifacts/webdemo/go
```

To compare those direct Go exports against Matplotlib references for the same
demo source, generate both image sets and open the parity viewer:

```bash
just web-parity-update
just parity-viewer-all
```

Use `just parity-viewer` for only the standard golden/reference suite, or
`just web-parity-viewer` for only the browser-demo suite.

The GitHub Actions workflow [`.github/workflows/deploy-wasm.yml`](.github/workflows/deploy-wasm.yml)
builds the same artifact and deploys it to GitHub Pages on pushes to `main`.

---

🚀 _Plotting for Go, without compromise._

# Examples Gallery

The curated gallery is driven by `internal/examplecatalog`. Rows with
`Showcase: true` are user-facing examples with an importable source package at
`examples/<id>/example.go`, a runnable `Plot() *core.Figure` snippet, and a
`Render() image.Image` helper for quick raster output.

## Browser Gallery

The browser gallery is the primary interactive entry point. It renders selected
showcase examples through Go compiled to WebAssembly and exposes the same source
files that the parity tests use.

```bash
just web-build
python3 -m http.server 8000 --directory web
```

Then open `http://localhost:8000`.

## CLI Gallery

Use `cmd/example` when you want deterministic files from the same showcase
catalog without serving the web app:

```bash
go run ./cmd/example -list
go run ./cmd/example -name basic_line -o basic_line.png
go run ./cmd/example -name colorbar_composition -format svg -o colorbar.svg
```

The CLI writes PNG, SVG, PDF, PS/EPS, and PGF through the registered backend
save dispatch.

## Anti-Gallery

These entries document deliberate differences from Matplotlib behavior. They
are not regressions; they are choices made to keep output deterministic,
portable, or Go-idiomatic.

| Area | Divergence | Reason |
| ---- | ---------- | ------ |
| Boxplot bootstrap | `BoxPlot2D.Bootstrap` is accepted for API parity, but confidence intervals use the deterministic fallback unless explicit intervals are provided. | Random bootstrap output is hostile to golden-image tests and reproducible examples. |
| PostScript alpha | Level-2 PS emits partially transparent vector artists as opaque RGB; RGBA images are composited over white. | Level-2 PostScript has no PDF-style alpha graphics state. |
| PostScript images | JPEG passthrough is not exposed in PS output. | `colorimage` consumes decoded sample data rather than PDF-style image XObjects. |
| PGF dense images | Raster images and mixed raster/vector groups are emitted as self-contained PGF pixel rectangles. | This keeps PGF output deterministic and sidecar-free, at the cost of compactness. |
| Example helpers | Some Go examples use helpers such as `AnnotatedHeatmap` and `Spy` instead of mirroring every Matplotlib call one-for-one. | The gallery favors readable Go API usage while parity tests keep the underlying behavior checked against Matplotlib references. |


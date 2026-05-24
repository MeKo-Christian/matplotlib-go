# Backend Selection Guide

matplotlib-go renders through a pluggable backend: the artist tree in `core`
calls a backend-agnostic `render.Renderer`, and a concrete backend turns those
calls into pixels or vector output. This guide explains **which backend to use
when** and shows the capability differences that drive the choice.

For the registry API and how to add a backend, see
[`docs/backends.md`](backends.md). For the renderer contract itself, see the
GoDoc for package `render`.

## TL;DR

| You want…                                        | Use         |
| ------------------------------------------------ | ----------- |
| A PNG that looks like Matplotlib                 | **AGG**     |
| A build with zero non-stdlib dependencies        | **GoBasic** |
| A scalable figure for the web or further editing | **SVG**     |
| A print-ready document with embedded fonts       | **PDF**     |
| EPS/PostScript for a legacy publishing pipeline  | **PS**      |
| A figure typeset by LaTeX with document fonts    | **PGF**     |
| Skia parity work / future GPU acceleration       | **Skia**    |

If you do not care, do nothing: the registry auto-selects a backend, and
`Savefig`/`SaveFig` pick one from the file extension.

## Choosing by output format

The simplest rule is to let the **file extension** choose:

```go
core.SaveFig(fig, renderer, "plot.png") // raster  -> AGG / GoBasic / Skia
core.SaveFig(fig, renderer, "plot.svg") // vector  -> SVG
core.SaveFig(fig, renderer, "plot.pdf") // vector  -> PDF
core.SaveFig(fig, renderer, "plot.eps") // vector  -> PS
core.SaveFig(fig, renderer, "plot.pgf") // vector  -> PGF
```

`pyplot.Savefig` and `backends.SelectBackendForExtension` apply the same
mapping. You only need to think about backends explicitly when two backends
can produce the same format (raster: AGG, GoBasic, Skia) or when you have a
hard constraint such as "pure Go only".

## The backends

### AGG — primary raster backend

Anti-Grain Geometry renderer with sub-pixel-accurate anti-aliasing, font
hinting, and transformed-image support. **This is the parity reference**: the
golden fixtures in `test/testdata/golden` are produced with AGG, so it is the
backend that most closely matches Matplotlib's own AGG output.

Use AGG for publication-quality PNGs and anywhere text fidelity matters.
It is available by default (the `agg_go` dependency is vendored as optional but
builds out of the box).

### GoBasic — pure-Go fallback

Pure-Go renderer built on `golang.org/x/image/vector`. No dependency beyond
`golang.org/x/image`. Anti-aliasing is good; text shaping and pattern/gradient
fills are basic or fall back to approximations.

Use GoBasic when the deployment forbids non-stdlib/cgo dependencies, for the
smallest possible build, or as the correctness fallback. It is the default
backend when nothing else is selected.

### SVG — primary vector backend

Pure-Go SVG generator with native `<text>` output, path recording, and
clip-path transforms. Deterministic byte output, so it is safe to diff in
tests and commit. Scales without resolution loss and is editable in vector
tools and browsers.

Use SVG for the web, for figures that will be re-styled downstream, and for
structurally testable vector output. SVG is the entry point for the WASM
browser gallery.

### PDF — publication vector backend

Pure-Go PDF backend with deterministic serialization and **embedded fonts**,
so a document renders identically anywhere. Supports pattern and gradient
fills and image transforms; metadata can be set via `render.WithPDFMetadata`.
Path effects stay vector for normal stroke/fill effects and identity filter
passes; blurred path-effect filters are emitted as backend-local image XObjects
with soft masks, keeping the routing behind the path-effect filter capability.

Use PDF for print-ready, self-contained documents and reports.

### PS — PostScript / EPS backend

Pure-Go Level-2 PostScript/EPS backend with deterministic serialization.

Use PS only when a legacy publishing pipeline requires EPS/PostScript;
otherwise prefer PDF.

### PGF — LaTeX-native backend

Generator-only PGF/TikZ backend: it emits TikZ source for `\input` into a
LaTeX document. It does **not** rasterize or anti-alias on its own — LaTeX does
the typesetting, so figure text uses the document's fonts and math.

Use PGF when a figure must match the surrounding LaTeX document exactly.
Matplotlib-Go keeps ordinary PGF saves dependency-free: draw-time text metrics
are deterministic approximations, while exact TeX/font metrics are delegated to
LaTeX at document compile time. PGF-specific save options are available through
the shared `render.SaveOption` surface, including `render.WithPGFMetadata`,
`render.WithPGFPreamble`, `render.WithPGFCommentPolicy`, and
`render.WithPGFVerificationMode`.

### Skia — opt-in accelerated raster backend

Build-tagged CPU raster renderer (`-tags skia`) with a Skia-local bridge
boundary; the external Skia C ABI and GPU mode are deferred. It supports text
shaping, pattern/gradient fills, and PNG export.

Use Skia for Skia-parity development and static PNG comparisons. It is **not**
registered unless you build with `-tags skia`, so it never appears in the
auto-selection of a default build.

## Capability matrix

Excerpt from `go run ./examples/backends/demo/main.go --capabilities`
(`✓` = supported, `~` = partial/fallback, blank = unsupported). Skia is omitted
because it requires `-tags skia`.

| Capability       | AGG | GoBasic | SVG | PDF | PS  | PGF |
| ---------------- | --- | ------- | --- | --- | --- | --- |
| Anti-aliasing    | ✓   | ✓       | ✓   | ✓   |     |     |
| Sub-pixel        | ✓   |         |     |     |     |     |
| Font hinting     | ✓   |         |     |     |     |     |
| Text shaping     | ✓   | ✓       | ✓   | ✓   | ✓   | ✓   |
| Pattern fill     | ✓   | ~       | ✓   | ✓   |     |     |
| Gradient fill    | ✓   | ~       | ✓   | ✓   |     |     |
| Image transform  | ✓   | ✓       | ✓   | ✓   | ✓   | ✓   |
| Clip-path xform  |     |         | ✓   |     |     |     |
| Vector output    |     |         | ✓   | ✓   | ✓   | ✓   |
| GPU acceleration |     |         |     |     |     |     |

Notes:

- "Anti-aliasing" for PS/PGF is blank because those formats delegate
  rasterization to the consumer (a PostScript interpreter or LaTeX).
- Query capabilities programmatically with `backends.HasCapability`,
  `backends.SupportsRendererCapability`, and
  `backends.VerifyRendererCapabilities` before relying on a feature.

## Selecting a backend in code

Prefer capability-driven selection over hard-coding a backend name:

```go
import (
    "github.com/cwbudde/matplotlib-go/backends"
    _ "github.com/cwbudde/matplotlib-go/backends/all" // register all backends
)

// Auto-select the best available backend (falls back to GoBasic).
backend, err := backends.GetBestBackend(nil)

// Require capabilities; selection fails if nothing satisfies them.
backend, err = backends.GetBestBackend([]backends.Capability{
    backends.TextShaping,
    backends.FontHinting,
})

// Honor the MATPLOTLIB_BACKEND environment variable, then fall back.
renderer, _, err := backends.NewRendererFromEnv(
    backends.SimpleConfig(640, 480, white), backends.TextCapabilities)
```

Override the auto-selection at run time without recompiling:

```sh
MATPLOTLIB_BACKEND=svg go run ./yourprogram
```

## Decision checklist

1. **Need a specific file format?** The extension decides — let `SaveFig`
   dispatch.
2. **Raster, and parity matters?** AGG.
3. **Raster, and no non-stdlib dependencies allowed?** GoBasic.
4. **Vector for the web or further editing?** SVG.
5. **Self-contained print document?** PDF (EPS only for legacy: PS).
6. **Figure embedded in a LaTeX document?** PGF.
7. **Working on Skia parity / GPU future?** Build with `-tags skia`.
8. **Unsure?** Pass `nil` to `GetBestBackend` and move on.

# Performance Profiling Notes

Last updated: 2026-06-14.

This note records the first Phase 4.2 profiling sweep. The benchmark harness is
in `benchmarks/render_benchmark_test.go`.

## Commands

```bash
just bench-render
BENCHTIME=10x just bench-render

just profile-render
CATALOG_BENCHTIME=1x SCATTER_BENCHTIME=1x just profile-render
```

Benchmark reports and profiles are written under `testdata/_artifacts/perf/`.
GitHub Actions also runs `.github/workflows/benchmark-report.yml` as a
report-only workflow and uploads that directory as the
`render-benchmark-report` artifact. The workflow is intentionally non-blocking
until baseline variance is known.

## Baseline Results

Machine: Linux amd64, Intel Core i7-1255U.

Selected catalog render benchmarks, `benchtime=10x`:

| Case                     |       ns/op |       B/op | allocs/op |
| ------------------------ | ----------: | ---------: | --------: |
| `basic_line`             |  18,118,878 |  2,803,944 |     4,852 |
| `lines_markers_gallery`  |  53,205,720 |  7,943,089 |    20,795 |
| `scatter_gallery`        |  64,560,622 |  6,931,000 |    19,348 |
| `image_variants_gallery` | 145,327,965 | 18,769,086 |   465,069 |
| `mesh_contour_tri`       |  77,675,508 | 10,107,532 |    33,912 |
| `triangulation_gallery`  | 135,610,916 | 15,872,982 |    50,518 |
| `mplot3d_gallery`        | 154,059,404 | 21,140,244 |   115,285 |
| `text_layout_gallery`    |  41,318,026 |  9,493,024 |    22,868 |
| `mathtext_gallery`       |  58,320,050 |  6,304,180 |    14,395 |
| `widgets_gallery`        |  49,179,709 | 10,869,548 |    26,804 |

100k scatter stress benchmark, `benchtime=5x`:

| Case                            |       ns/op |        B/op | allocs/op |
| ------------------------------- | ----------: | ----------: | --------: |
| `BenchmarkLargeScatter100KDraw` | 604,604,763 | 366,167,259 | 3,724,508 |

P2 renderer-reuse smoke benchmark after the AGG reuse path landed,
`benchtime=1x`:

| Case                                           |       ns/op |        B/op | allocs/op |
| ---------------------------------------------- | ----------: | ----------: | --------: |
| `BenchmarkLargeScatter100KDraw`                | 462,756,293 | 132,824,952 | 2,429,696 |
| `BenchmarkLargeScatter100KRedrawReuseRenderer` | 460,516,167 | 125,008,520 | 2,423,092 |

P2 scalar-mapping smoke benchmarks after the resolved-colormap cache landed,
`benchtime=1x`:

| Case                                  |     ns/op | B/op | allocs/op |
| ------------------------------------- | --------: | ---: | --------: |
| `BenchmarkScalarMappedImageColors`    | 2,405,271 |    0 |         0 |
| `BenchmarkScalarMappedScatterColors`  | 7,144,169 |    0 |         0 |
| `BenchmarkScalarMappedQuadMeshColors` | 1,178,998 |    0 |         0 |

The selected catalog cases are below the sub-second typical-plot target on this
machine. The 100k scatter stress case is also below one second, but allocation
volume is high enough to make GC pressure the main risk for repeated or
interactive rendering.

## Regression Budgets

These budgets are intentionally loose while CI is report-only. They make the
first P1 target explicit and should be tightened after several benchmark
artifacts establish normal variance.

| Benchmark                       | Time budget | Allocation budget |       Object budget |
| ------------------------------- | ----------: | ----------------: | ------------------: |
| `BenchmarkLargeScatter100KDraw` |   700 ms/op |         400 MB/op | 4,000,000 allocs/op |

## Memory Targets And Tuning Guide

These are v1.0 guidance targets for the AGG backend on the profiling machine,
not hard CI gates yet. Treat them as review thresholds when changing renderer,
collection, image, or text code.

| Scenario                  | Benchmark or proxy                                                                                               |                           v1.0 memory target | Notes                                                                                                                                                                                          |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------- | -------------------------------------------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Typical catalog plots     | `BenchmarkCatalogRender` rows such as `basic_line`, `scatter_gallery`, `mesh_contour_tri`, and `mplot3d_gallery` | <= 25 MB/op for ordinary gallery-scale plots | Current selected rows range from about 2.8 MB/op to 21.2 MB/op. Catalog additions that exceed this should document the reason, such as large raster inputs or intentionally dense text.        |
| 100k scatter stress       | `BenchmarkLargeScatter100KDraw`                                                                                  |           <= 150 MB/op and <= 2.6M allocs/op | The post-P1/P2 smoke row is about 133 MB/op one-shot and 125 MB/op with renderer reuse. Keep this below the loose regression budget and tighten after several CI artifacts establish variance. |
| Repeated redraw           | `BenchmarkLargeScatter100KRedrawReuseRenderer`                                                                   |         <= 130 MB/op with stable canvas size | Repeated redraw should reuse the AGG renderer and avoid owned image copies unless the caller needs to retain the frame. The current smoke row is about 125 MB/op.                              |
| Scalar-mapped color loops | `BenchmarkScalarMappedImageColors`, `BenchmarkScalarMappedScatterColors`, `BenchmarkScalarMappedQuadMeshColors`  |                0 B/op after scalar-map setup | Per-value mapping on a resolved scalar map should remain allocation-free.                                                                                                                      |

Practical tuning advice:

- Reuse renderers for long-running views. Keep one `agg.Renderer` per stable
  canvas size, call `agg.Renderer.Clear(fig.RC.FigureBackground())`, then draw
  the next frame into the same surface.
- Avoid `GetImage` when ownership is unnecessary. Prefer
  `agg.Renderer.ImageView` for PNG encoding, UI upload, or benchmark inspection
  where the caller consumes the frame before the next `Clear`.
- Batch markers through collection paths instead of many independent artists.
  `PathCollection` can route repeated marker prototypes through `MarkerBatch`,
  which avoids per-point full-path allocation in the backend fast path.
- Text-heavy tick labels are still the main CPU risk. Reduce tick density,
  avoid unnecessary redraws of unchanged axes, and profile text-heavy catalog
  additions with `just profile-render`.
- Backend selection matters. The default AGG backend is the parity backend and
  should be used for golden/reference-sensitive work. Simpler or future
  interactive backends may trade byte-for-byte parity for lower redraw overhead;
  measure with the same benchmark shape before switching.

## Hotspots

### 1. Text measurement through native FreeType

Catalog CPU profile:

- `runtime.cgocall`: 63.67% flat.
- `agg.withNativeFreetypeRun`: 62.24% cumulative.
- `FT_Load_Glyph`: 38.86% cumulative.
- `FT_New_Face`: 13.17% cumulative.
- `core.measureSingleLineTextLayoutParseMath`: 56.42% cumulative.
- Tick-label drawing alone (`Axis.DrawTickLabels`) accounts for 31.61%
  cumulative.

Interpretation: text-heavy typical plots spend most CPU in repeated FreeType
face creation and glyph loading for measurement/layout, especially tick labels.
The likely optimization target is caching native FreeType face/run metrics per
font/size/DPI/text/hinting tuple, or keeping an initialized face cache in the
AGG renderer.

### 2. Text shaping and font fallback allocations

Catalog allocation profile:

- `render.ShapeTextRuns`: 165.65 MB cumulative, 107.17 MB flat.
- `render.fontFaceSupportsRune`: 169.20 MB cumulative, 146.75 MB flat.
- `render.fontFaceCacheKey`: 53.01 MB flat.
- `core.measureSingleLineTextLayoutParseMath`: 300.11 MB cumulative.

Interpretation: text layout repeatedly builds shaping buffers, font cache keys,
and rune-support checks. Reusing `sfnt.Buffer`, memoizing support checks, and
avoiding repeated `filepath.Clean` string construction in hot loops should
reduce allocations.

### 3. Renderer surface and image extraction allocations

Catalog allocation profile:

- `image.NewRGBA`: 300.97 MB flat.
- `agg.newAggSurface`: 277.52 MB cumulative.
- `agg.GetImage` / `agg_go.Image.ToGoImage`: 278.02 MB cumulative.

Interpretation: each render allocates a fresh AGG surface and then copies it to
a Go image for benchmark parity. This is expected for one-shot exports, but
interactive and repeated rendering should reuse renderer/surface storage where
possible and avoid `GetImage` copies when the caller only needs to draw or save.

### 4. 100k scatter path-collection rendering

100k scatter CPU profile:

- `core.PathCollection.drawPathCollection`: 84.91% cumulative.
- `agg.Renderer.DrawPathCollection`: 73.72% cumulative.
- `agg.Renderer.drawPathDirect`: 71.78% cumulative.
- AGG rasterization/blending routines dominate the flat CPU after collection
  dispatch (`BlendPix`, rasterizer sorting/line/sweep).

100k scatter allocation profile:

- `core.PathCollection.drawPathCollection`: 1,998.72 MB cumulative, 95.61%.
- `geom.Path.CubicTo`: 568.69 MB flat.
- `core.applyAffinePath`: 473.68 MB flat.
- `agg.Renderer.devPath`: 222.59 MB flat.
- `agg_go ConvTransform.Vertex`: 197.50 MB flat.

Interpretation: the stress case repeatedly clones and transforms marker paths
per point, then flips/clones paths again in the AGG backend. The primary
optimization target is a true marker-batch path in AGG that transforms marker
vertices without allocating a full `geom.Path` per marker. A second target is
removing the extra `devPath` clone by folding y-flip into the native draw path
or reusing scratch buffers.

### 5. Scalar color mapping overhead

Catalog CPU profile:

- `core.ScalarMapInfo.Color`: 4.39% cumulative.
- `color.GetColormap`: 1.87% cumulative.

Interpretation: scalar-mapped image/collection paths repeatedly resolve
colormap metadata. Caching the resolved colormap and normalization state in
`ScalarMapInfo` users would be a small but visible win after the larger text
and path-collection issues.

## Recommended Optimization Order

1. Add native FreeType face/metric caching in AGG text measurement.
2. Memoize font fallback rune-support checks and reduce shaping-buffer churn.
3. Add or deepen AGG marker/path-collection fast paths for repeated marker
   prototypes, avoiding per-marker `geom.Path` allocation and duplicate
   y-flip clones.
4. Add renderer/surface reuse APIs for repeated interactive renders.
5. Cache resolved colormap/norm state for scalar-mapped artists.

## P1 Progress

Implemented on 2026-06-14:

- AGG native FreeType measurement now caches measured text-run bounds/metrics by
  font path, text, size, DPI, and hinting factor. This avoids repeated
  `FT_New_Face` / glyph-load measurement work for repeated tick labels and for
  `MeasureText` + `MeasureTextBounds` pairs.
- Text fallback now memoizes `fontFaceSupportsRune(face, rune)` by canonical
  face key and rune, and `ShapeTextRuns` reuses its local `sfnt.Buffer` across
  resolved runs.
- Display-space scatter markers now combine scale and translation into one
  affine path application. The focused `benchtime=1x` smoke row moved
  `BenchmarkLargeScatter100KDraw` to about 323 MB/op and 3.52M allocs/op on the
  profiling machine. Run a longer `BENCHTIME=10x just bench-render` sweep before
  tightening the regression budget.
- `PathCollection` now routes single-prototype display-space markers through
  `MarkerBatch` even when size, face color, edge color, and linewidth vary per
  item. AGG consumes that batch with renderer scratch storage, applying marker
  transform and y-flip in one pass and avoiding the extra `devPath` clone.
  A focused `benchtime=1x` smoke row for `BenchmarkLargeScatter100KDraw` moved
  again to about 524 ms/op, 135 MB/op, and 2.43M allocs/op. Longer benchmark
  sweeps should set the final guarded threshold.

## P2 Progress

Implemented on 2026-06-14:

- `agg.Renderer.Clear` resets the reusable AGG surface, renderer clip state, and
  active clip box before the next frame. Long-running apps can keep one renderer
  per stable canvas size, call `Clear(fig.RC.FigureBackground())`, then call
  `core.DrawFigure(fig, renderer)` for each redraw instead of constructing a
  fresh renderer.
- `agg.Renderer.ImageView` exposes a non-owning view over the AGG buffer for
  callers that only need to encode, copy to a UI surface, or benchmark the
  result. `GetImage` remains the owned-copy API.
- `SavePNG` and the render benchmark harness now use the zero-copy view where an
  owned Go image is unnecessary. The reused renderer benchmark is
  `BenchmarkLargeScatter100KRedrawReuseRenderer`.
- A focused `benchtime=1x` smoke run moved the 100k scatter redraw path from
  about 463 ms/op, 133 MB/op, and 2.43M allocs/op for one-shot rendering to
  about 461 ms/op, 125 MB/op, and 2.42M allocs/op with a reused renderer.

## P2 Scalar-Mapping Progress

Implemented on 2026-06-14:

- `ScalarMapInfo.Resolved` now caches the resolved colormap alongside the
  resolved colormap name and normalizer/range state. Per-value `Color` calls on
  a resolved mapping no longer call `color.GetColormap`, which avoids repeated
  name normalization and reversed-colormap reconstruction in image, scatter, and
  mesh mapping loops.
- Focused benchmarks now cover scalar-mapped image grids, scatter arrays, and
  quad-mesh cell averages:
  `BenchmarkScalarMappedImageColors`, `BenchmarkScalarMappedScatterColors`, and
  `BenchmarkScalarMappedQuadMeshColors`.
- A focused `benchtime=1x` smoke run reported zero allocations for all three
  per-value scalar mapping loops after setup.

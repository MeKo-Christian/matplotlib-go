# Large File Decomposition

This note records the baseline for the decomposition effort: split the
largest source and test files into focused units without changing behavior.

Run the audit with `just large-file-audit`:

```bash
just large-file-audit
```

The command reports tracked and untracked (but non-ignored) working-tree Go
files at or above 1000 lines and non-Go artifacts at or above 256 KiB. It
skips deleted paths plus ignored build outputs and local diagnostic artifacts,
so package moves can be audited before they are staged.

## Baseline Inventory

Captured from the tracked tree before decomposition work began.

### Large Go Files

| Lines | File                                               |
| ----: | -------------------------------------------------- |
|  4834 | pre-split 3D tests (now `plot3d/*_test.go`)        |
|  3037 | `internal/examplecatalog/public_surface_parity.go` |
|  3002 | `core/text_test.go`                                |
|  2171 | `core/axis_test.go`                                |
|  2049 | `core/contour.go`                                  |
|  1803 | `pyplot/pyplot.go`                                 |
|  1779 | `core/axis.go`                                     |
|  1614 | `pyplot/pyplot_test.go`                            |
|  1612 | `canvas/widget_interaction_test.go`                |
|  1601 | `cmd/parityviewer/main.go`                         |
|  1587 | `core/text.go`                                     |
|  1547 | pre-split 3D contour/surface code (now `plot3d/`)  |
|  1524 | `core/mesh_contour_test.go`                        |
|  1496 | `core/legend_test.go`                              |
|  1464 | `core/patch_test.go`                               |
|  1409 | `backends/svg/svg_test.go`                         |
|  1394 | `backends/agg/agg_test.go`                         |
|  1340 | `core/arrow_patch.go`                              |
|  1327 | `core/legend.go`                                   |
|  1311 | `core/plot.go`                                     |
|  1295 | `core/colorbar.go`                                 |
|  1280 | `backends/pdf/pdf_test.go`                         |
|  1242 | `core/scatter.go`                                  |
|  1174 | `backends/ps/ps.go`                                |
|  1139 | `core/colorbar_test.go`                            |
|  1139 | `backends/gobasic/gobasic.go`                      |
|  1129 | `color/named_colors_data.go`                       |
|  1122 | `core/collection_test.go`                          |
|  1106 | `backends/agg/freetype_native.go`                  |
|  1096 | `backends/agg/agg_paths.go`                        |
|  1051 | `test/diagnostics_test.go`                         |
|  1051 | `style/mplstyle.go`                                |
|  1048 | `backends/pgf/pgf.go`                              |

### Large Non-Go Artifacts

|     Size | File                                                      |
| -------: | --------------------------------------------------------- |
| 2136 KiB | `docs/matplotlib-parity-status.md`                        |
| 1340 KiB | `testdata/svg_golden/mathtext_basic.svg`                  |
|  440 KiB | `test/testdata/public_api/stable_public_api.json`         |
|  344 KiB | `testdata/matplotlib_ref/mplot3d_gallery.png`             |
|  324 KiB | `testdata/matplotlib_ref/projection_toolkit_gallery.png`  |
|  300 KiB | `testdata/matplotlib_ref/imshow_interpolation_matrix.png` |
|  296 KiB | `testdata/golden/mplot3d_gallery.png`                     |
|  292 KiB | `testdata/golden/projection_toolkit_gallery.png`          |
|  264 KiB | `testdata/svg_golden/mixed_raster_vector.svg`             |
|  264 KiB | `testdata/golden/imshow_interpolation_matrix.png`         |

## Decomposition Rules

- Prefer move-only splits by responsibility; do not combine behavior changes
  with file decomposition.
- Keep package boundaries and public API stable unless a later PLAN item
  explicitly authorizes an API change.
- Split tests by behavior family first, moving shared test helpers into
  dedicated helper files when multiple split files need them.
- Treat generated catalogs, color tables, and golden/reference fixtures as
  large-by-design unless the phase records a generator or sharding decision.

## Generated and Fixture Data Strategy L7 keeps the large generated/catalog artifacts intact and documents why

they are exceptions to the 1k-line source target:

| File                                                | Decision                   | Rationale                                                                                                                                                                                                                                                                                                                                                                                                                           | Drift guard                                                                                                                                                                                                                                                                                                     |
| --------------------------------------------------- | -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `internal/examplecatalog/public_surface_parity.go`  | Keep-large curated catalog | The file is hand-maintained parity classification data, not raw generated output. It cross-links upstream public-surface rows to local feature coverage, catalog cases, examples, implementation files, closure phases, and rationale notes. Splitting it by module would make duplicate/overlapping public-surface decisions harder to review and would separate related override/rule decisions from the exported lookup helpers. | `TestAllUpstreamPublicRowsAreClassified`, `TestPartialAndOmissionRowsHaveNotes`, and the landmark public-surface tests keep the catalog complete and documented against `test/testdata/parity_surface/upstream_public_surface.json`, which is generated by `internal/examplecatalog/extract_public_surface.py`. |
| `color/named_colors_data.go`                        | Keep-large generated table | The file is generated from `third_party/matplotlib/lib/matplotlib/_color_data.py` and embedded as Go string tables so runtime color lookup has no Python or `third_party` dependency. Artificially sharding CSS4/Tableau/xkcd rows would not improve review quality because the source of truth is upstream Matplotlib data.                                                                                                        | `TestNamedColorInventoryMatchesMatplotlibTables` compares the committed Go tables against the vendored Matplotlib `_color_data.py`; `TestNamedColorCatalogSizesMatchMatplotlib` locks the expected table sizes.                                                                                                 |
| `docs/plans/{prebreak-public-api,api-tiering}.json` | Keep-large API audit       | The pre-break API snapshot and its exhaustive keep/demote/delete classification are review evidence for the coordinated v1 API break. Sharding either file would make exact symbol coverage and provenance harder to verify.                                                                                                                                                                                                        | `TestPublicAPITieringMatchesPreBreakSnapshot` checks the preserved snapshot hash, row count, duplicate-free exact coverage, and landmark decisions; the adjacent generator has a deterministic `--check` mode.                                                                                                  |
| Golden/reference PNG, SVG, PDF, and JSON fixtures   | Keep intact                | Binary and serialized fixtures are review artifacts. Splitting or repacking them would add fixture plumbing without reducing behavior risk.                                                                                                                                                                                                                                                                                         | Existing golden/reference, SVG/PDF comparison, and public API artifact tests validate these files at their natural granularity.                                                                                                                                                                                 |

## Verification Snapshot

Refreshed at the close of the API rework, after the package moves (`plot3d`, `ticker`,
`dates`, `widgets`), the error/options conventions, the mutable-field cleanup,
and the scalar-map consolidation. The remaining large files are either
catalog/fixture data covered above or behavior families whose extra split would
currently add more cross-file coupling than review clarity.

```text
just large-file-audit

Large tracked or untracked Go files (>= 1000 lines)
   3039 internal/examplecatalog/public_surface_parity.go
   2848 style/mplstyle.go
   2169 core/plot.go
   1761 plot3d/wire_surface_test.go
   1385 core/colorbar_test.go
   1385 core/arrow_patch.go
   1319 style/style.go
   1219 core/scatter.go
   1140 ticker/locators.go
   1136 core/collection_test.go
   1129 color/named_colors_data.go
   1103 core/line.go
   1083 ticker/formatters.go
   1067 dates/date_tick.go
   1052 render/extensions.go
   1044 backends/webagg/webagg_test.go
   1037 core/signal_helpers.go
   1032 ticker/ticker_test.go
   1009 core/line_test.go

Large tracked or untracked non-Go artifacts (>= 256 KiB)
   2148K docs/matplotlib-parity-status.md
   1832K testdata/svg_golden/mathtext_basic.svg
    512K test/testdata/public_api/stable_public_api.json
    496K docs/plans/prebreak-public-api.json
    344K testdata/matplotlib_ref/mplot3d_gallery.png
    336K testdata/matplotlib_ref/projection_toolkit_gallery.png
    332K docs/plans/api-tiering.json
    300K testdata/matplotlib_ref/imshow_interpolation_matrix.png
    296K testdata/golden/mplot3d_gallery.png
    292K testdata/golden/projection_toolkit_gallery.png
    272K testdata/svg_golden/mixed_raster_vector.svg
    264K testdata/golden/imshow_interpolation_matrix.png
```

The file set is unchanged from the mid-Phase-2 snapshot; only line counts
moved. `core/plot.go` grew the most (1903 → 2169) because the options migration
moved every plot-family option struct from variadic pointer fields to a single
`optional.Value[T]` options value declared beside its entry point, and the
`stable_public_api.json` growth (500 KiB → 512 KiB) is the re-frozen surface
described in `docs/plans/mutable-fields.md`.

### Remaining Large Go Source Decisions

- `plot3d/wire_surface_test.go`: keep-large behavior family. The file now
  contains only the 3D wireframe, surface, trisurf, and voxel tests from the
  former monolithic 3D test file. These tests share 3D setup, projection
  assumptions, and rendered polygon ordering assertions; another split would
  separate closely related mplot3d surface-family cases. Drift guard:
  `go test ./plot3d -run 'TestAxes3D(Wireframe|Surface|TriSurf|Voxel)'` and the
  catalog mplot3d parity rows.
- `ticker/locators.go`, `ticker/formatters.go`, `ticker/ticker_test.go`, and
  `dates/date_tick.go`: keep-large migrated families for the `dates` package
  split. Locators and formatters each share normalization and tick-context
  contracts, while the test file covers their cross-family behavior. A later
  decomposition can split them by linear/log/date family after the pre-v1 API
  is re-frozen. Drift guard: `go test ./ticker ./dates`.
- `core/arrow_patch.go`: keep-large public API family. Arrow patches combine
  Matplotlib-compatible connection styles, arrow styles, mutation scaling,
  clipping, and the `FancyArrowPatch` artist. The helpers are tightly coupled to
  shared path construction semantics, while the patch source files around it
  already hold the simpler patch families. Drift guard:
  `go test ./core -run 'Test(Arrow|Connection|FancyArrowPatch)'`.
- `core/plot.go`: keep-large plotting primitive family. This file remains the
  central Line2D and plot-family implementation. It keeps style cycling,
  marker/stroke options, draw-time marker batching, autoscale behavior, and
  legend handle integration in one review unit; moving more pieces would create
  package-local helper traffic without reducing an algorithmic hotspot. Drift
  guard: `go test ./core -run 'Test(Line|Plot|Dashes|Marker)'` plus line/plot
  catalog parity rows.
- `core/scatter.go`: keep-large collection primitive family. Scatter owns the
  public scatter options, scalar-mappable integration, marker sizing/color
  resolution, legend sampling, picking, and draw dispatch. Those pieces share
  collection state and should move together if a future collection-layer
  refactor happens. Drift guard: `go test ./core -run 'TestScatter'` and
  scatter catalog parity rows.
- `core/colorbar_test.go`: keep-large test family. Production colorbar code is
  split into layout, scale, and draw files. The remaining test file is a compact
  integration suite that exercises those pieces together, including mutable
  mappable and formatter behavior that is less clear when split away from its
  shared fixtures. Drift guard: `go test ./core -run 'TestColorbar'` and
  colorbar catalog parity rows.
- `core/collection_test.go`: keep-large test family. Collection rendering tests
  intentionally stay together because they share recorder helpers and validate
  cross-collection contracts for path, patch, line, and scalar-mapped
  collections. Drift guard: `go test ./core -run 'Test.*Collection'` and
  collection catalog parity rows.
- `style/mplstyle.go`: keep-large parser family. The file is a self-contained
  Matplotlib `.mplstyle` parser and rc translation table. Splitting the parser
  from the key/value conversion tables would make review harder unless a future
  style-system refactor introduces a generated table or schema. Drift guard:
  `go test ./style`.
- `style/style.go`: keep-large configuration family. The bulk of the file is the
  `RC` struct itself plus its defaults and the `With*` option constructors that
  write single `RC` fields. Splitting the options away from the struct they
  mutate would force every field addition into two files without isolating any
  logic. Drift guard: `go test ./style`.
- `core/line.go`: keep-large artist family. `Line2D` now owns the value types
  the mutable-field pass folded into it — `DashPattern`, `MarkerColorSpec`, and
  `MarkEverySpec` — alongside the artist's accessors. Those types exist to give
  the artist's fields one writer each, so they stay next to the fields they
  describe. Drift guard: `go test ./core -run 'TestLine2D'`.
- `core/line_test.go`: keep-large test family. The file pairs with `core/line.go`
  and shares one `recordingRenderer` fixture across the dash, marker, gap-color,
  and bounds cases; splitting it would duplicate that fixture. Drift guard: the
  file is its own guard.
- `core/signal_helpers.go`: keep-large algorithm family. The spectral entry
  points (`Specgram`, `PSD`, `CSD`, `Cohere`, `XCorr`, `ACorr`) are thin wrappers
  over a shared FFT windowing, detrending, and segment-averaging core. The
  helpers are only correct as a set, and Matplotlib keeps the same grouping in
  `mlab.py`. Drift guard: `go test ./core -run 'Test(Specgram|PSD|CSD|Cohere|Corr)'`.
- `render/extensions.go`: keep-large contract file. This is the optional
  renderer-capability surface: the save-option value types, their per-format
  validation, and the backend-specific `SVGOption`/`PDFOption`/`PSOption`/
  `PGFOption` families. Backends type-assert against these interfaces, so keeping
  the contract in one file is what makes the capability set reviewable. Drift
  guard: `go test ./render ./backends/...`.
- `backends/webagg/webagg_test.go`: keep-large test family. Roughly the first
  150 lines are stub implementations of the event loop, timer, artist, and
  renderer interfaces that every case in the file needs; the cases themselves
  exercise one `Manager` lifecycle. Drift guard: the file is its own guard.

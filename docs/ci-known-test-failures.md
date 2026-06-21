# CI: known pre-existing test failures (tracked follow-up)

The CI workflows were repaired so the project builds and the `build`, `vet`,
`fmt`, and `lint` jobs are green (see `.github/actions/setup-cgo-deps` and the
`ci.yml` / `test-unit.yml` / `test-lint.yml` / `test-fmt.yml` workflows). Before
that fix CI had **never built** (no system cgo dependencies were installed), so
the test suite itself was never validated on CI.

Running the full suite now (`CGO_ENABLED=1 go test -tags freetype ./...`) reveals
pre-existing failures that fail on `main` as well — they are **not** introduced
by the formatter/layout/date or MathText work on this branch. They are recorded
here so the `unit` / `Test (Linux)` jobs can be brought green in a focused,
parity-aware follow-up rather than papered over.

As of 2026-06-21, 3 packages fail locally; CI additionally hits an
environment-specific FreeType failure.

## 1. `internal/examplecatalog` — doc currency

- `TestMatplotlibParityStatusDocIsCurrent` — `docs/matplotlib-parity-status.md`
  no longer matches what the generator produces. The `fix: formatting` commit
  (prettier) reflowed that file, and the generated form differs.

**Fix:** regenerate the parity-status doc and commit, or exclude it from
prettier in `treefmt.toml` (as `docs/matplotlib-migration-notes.md` already is)
so generated content and the formatter do not fight.

## 2. root package (`.`) — PLAN.md / doc assertions

`large_file_audit_test.go` and the performance-doc tests assert that `PLAN.md`
and the docs contain exact marker strings:

- `TestLargeFileAuditPlumbingIsDocumented`, `TestGeneratedDataStrategyIsDocumented`
- `TestAxis{Types,Polar,Ticks,TickLabels,Spine}SplitIsTracked`
- `TestContour{API,Filled,Labels,Levels,Lines}SplitIsTracked`
- `TestPerformanceP2{MemoryTargetsAndTuningGuide,RendererReuse,ScalarMappingCache}IsDocumented`

These look for literal text such as
`[x] **L1 — Add a repeatable large-file audit.**`. That string is absent from
`origin/main`'s `PLAN.md`, so the assertions already fail there; prettier
reflowing `PLAN.md` in the `fix: formatting` commit makes the mismatch worse.

**Fix:** reconcile `PLAN.md`/docs with the expected markers (or relax the tests
to be whitespace/format-insensitive), and exclude `PLAN.md` from prettier in
`treefmt.toml` so the markers are stable. There is a direct conflict today:
prettier-formatting `PLAN.md` (required for the green `fmt` job) reflows the
markers these tests match on.

## 3. `test` package — golden / reference drift

Pre-existing rendering drift (the Phase 7 close-out note already flagged some of
these as "unrelated branch drift"):

- `TestReferenceCompare/{line2d_markers, legend_layout_matrix, geo_lambert_axes,
radar_basic}` — marginally over their per-case `MaxRMSE` caps (e.g.
  line2d_markers RMSE 3.17 vs cap 3.0; radar_basic 0.53 vs 0.50).
- `TestSVGGolden/{bar_basic, errorbar_basic, hist_basic, mathtext_basic,
polar_axes, mixed_raster_vector}` — structural SVG diffs vs the committed
  `.svg` goldens.
- `TestPDFGolden/{bar_basic, basic_line, hist_basic, mathtext_basic,
mathtext_inline_labels, mesh_contour_tri, mixed_raster_vector, polar_axes,
text_labels_strict}` — PDF golden drift.

**Fix:** regenerate the stale SVG/PDF goldens (`-update-svg-golden`, the PDF
golden update flag) once the render is confirmed correct, and ratchet the four
`TestReferenceCompare` `MaxRMSE` caps to the actual values plus small headroom
(the catalog's established pattern). This is parity-sensitive: confirm the
current render is the intended one before re-baselining.

## 4. CI-only: native FreeType `MeasureTextBounds`

In the CI runner only (passes locally), `backends/agg`'s
`TestMeasureTextBoundsUsesMatplotlibPlainTextBounds`,
`TestFontManagerPrefersBundledMatplotlibDejaVuFont`, and related FreeType
text-metric tests fail with `native MeasureTextBounds(...) failed` — the native
call errors. `TestNativeFreetypeIsPinned261` passes, so the linked FreeType **is**
2.6.1; the divergence is in font resolution / HarfBuzz shaping on the runner.

**Fix:** needs in-CI iteration (add diagnostics to the failing path; check
fontconfig/`HOME`, the installed `libharfbuzz-dev` version, and whether the
agg_go text path resolves the bundled DejaVu under the runner). Consider pointing
`PKG_CONFIG_PATH` at `third_party/freetype/prefix/lib/pkgconfig` so agg_go also
uses the vendored 2.6.1 rather than the system FreeType installed for it.

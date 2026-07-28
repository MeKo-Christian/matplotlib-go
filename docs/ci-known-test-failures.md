# CI: known pre-existing test failures (tracked follow-up)

The CI workflows were repaired so the project builds and the `build`, `vet`,
`fmt`, and `lint` jobs are green (see `.github/actions/setup-cgo-deps` and the
`test.yml` orchestrator with its `test-build` / `test-unit` / `test-lint` /
`test-fmt` reusable workflows). Before
that fix CI had **never built** (no system cgo dependencies were installed), so
the test suite itself was never validated on CI.

## Status (2026-06-22)

The local suite (`RUN_OPTIONAL_VISUAL_TESTS=true CGO_ENABLED=1 go test -tags
freetype ./...`) is now **green**. The doc/PLAN-marker, catalog, golden, and
parity-tolerance failures previously tracked here (sections 1–3 below) were
resolved, along with a set of newer regressions introduced by the contour
styles / colorbar / RGB-imshow feature commits. The only remaining known
failure is the CI-only native-FreeType text-metric case (section 4), which
passes locally and needs in-CI iteration.

## 4. CI-only: native FreeType `MeasureTextBounds` (open)

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

# Matplotlib-Go Development Plan

This plan tracks the **remaining** work to bring `matplotlib-go` to a stable
v1.0 release. The foundation is built; what follows is the focused path to the
finish. The roadmap is cross-checked against the local upstream Matplotlib
snapshot in `third_party/matplotlib` so uncovered areas stay explicit instead of
sliding into a vague "future work" bucket.

---

# Plan Tracking

- `✅` = done and stable
- `🧪` = implemented but under hardening
- `⚪` = in progress
- `⚠️` = deferred / design decision required
- `[ ]` = not started

---

# Already Shipped (Foundation Complete)

The project is well past proof-of-concept. The following are implemented,
tested, and stable:

- **Core model & transforms** — `Figure → Axes → Artists` with stale/callback
  propagation and draw scheduling; explicit transform graph (`transData`,
  `transAxes`, `transFigure`, blended/offset transforms, projection layer).
- **Backends (4)** — AGG (primary anti-aliased raster, native batches, buffer
  regions, filters), GoBasic (pure-Go correctness fallback), SVG (deterministic
  vector with structural diff harness), Skia (opt-in CPU raster). Capability
  matrix and save dispatch are registry-driven.
- **Publication vectors** — PDF, PS/EPS, and PGF via the shared save pipeline.
- **Plot vocabulary** — full 2D basics, statistical, color-mapped, vector-field,
  specialty, patches/collections, and 3D (`mplot3d`) families, plus non-Cartesian
  projections (polar, radar, skewx, geo).
- **Layout / style / API** — subplots, GridSpec, mosaic, subfigures, twin/secondary
  axes; `tight_layout` / `constrained_layout`; insets, dividers, image grids,
  axisartist; `rcParams` / themes / `.mplstyle`; OO core API plus a `pyplot` layer.
- **MathText & usetex** — first-class across the active backends.
- **Renderer effects** — patterns, gradients, and path effects (renderer-neutral
  pipeline with backend-native fast paths).
- **Interactive runtime** — headless event model, draw-idle/blit redraw, Gio
  desktop and WebAgg/WASM frontends, widgets and selectors.
- **Animation** — `FuncAnimation` / `ArtistAnimation` with GIF/APNG/HTML writers
  (dependency-free) and optional ffmpeg MP4/WebM.
- **Parity infrastructure** — catalog-driven `TestGolden` / `TestMatplotlibRef` /
  `TestReferenceCompare` with per-case tolerances; public-surface inventory and
  classification map; catalog-backed CLI and browser galleries.

*Former Phases 1–6, 8A, 9–19 are complete; see git history for the detailed
per-phase implementation logs.*

---

# Phase 1: Backend Deepening (Skia Native + GPU)

**Goal:** finish the backend-specific Skia work. The historical blocker (no Skia
C-ABI binding) is **resolved**: a real Skia library is built locally and a narrow
C-ABI wrapper links it under `-tags "skia skiacgo"`. The remaining work is wiring
the rest of the native primitives through that wrapper and standing up real GPU
mode — no longer external-access-blocked, just unbuilt.

## Done

- [x] **AGG parity diagnostics for non-text residuals** (the one self-contained,
      never-Skia-blocked item): `TestNonTextResidualDiagnostics`
      (`test/diagnostics_test.go`, env-gated by `MPL_GO_RESIDUAL_DIAG`) logs
      per-case residual metrics across dense path collections, translucent
      overlaps, image-interpolation modes, hatch clipping, and mixed
      raster/vector, dumping diff PNGs under
      `testdata/_artifacts/non_text_residuals/`.
- [x] **C-ABI Skia wrapper + cgo bridge** (`backends/skia/skia_cwrap.{h,cpp}`,
      `native_cgo.go`; tag `skiacgo`). Links Skia milestone 151. Native today:
      gradient path fills (`SkShaders` gradients), marker batches
      (`SkCanvas`/`SkPath`), Gouraud triangles (`SkVertices`). Verified by
      `native_cgo_test.go`. `IsCapabilityBridged` flips `MarkerBatch` to native
      (`✓`) when the native surface is linked; default `-tags skia` stays the
      pure-Go CPU bridge.

## Build / test (carry-on entry points)

- Build the library once (outside the repo): a Skia checkout with `include/` and
  a built `out/Shared/libskia.so`. For combined use with text, build with
  `skia_use_freetype=false` (see caveat).
- `just build-skia-native` / `just test-skia-native` (override `SKIA_ROOT`, default
  `/mnt/projekte/Code/skia`). They pass `CGO_CXXFLAGS`/`CGO_LDFLAGS` and add an
  rpath so the test binary finds `libskia.so` without `LD_LIBRARY_PATH`.
- **FreeType caveat:** a Skia built without system FreeType statically bundles its
  own. Combining `skiacgo` with the `freetype` tag (agg's vendored FreeType 2.6.1)
  in one binary produces duplicate `FT_*` symbols and a runtime crash, so the
  native recipes deliberately omit `freetype`. Rebuild Skia with
  `skia_use_freetype=false` to run native Skia and agg text in one binary.

## Remaining work (each unblocked; wire the wrapper entrypoint)

- [x] **Native path collections.** Add a `drawPathCollectionNative` to the
      `nativeBatchBridge` interface + dispatch in `Renderer.DrawPathCollection`
      (`skia_native.go`); render all items into one native surface (loop
      `mgsk_draw_path`, or add a batched multi-path C entrypoint). Flip
      `pathcollectionbatch` in `IsCapabilityBridged`.
- [x] **Native quad meshes.** Add `drawQuadMeshNative`; emit two `SkVertices`
      triangles per cell with the face color (reuse `mgsk_draw_vertices`) or one
      `mgsk_draw_path` per cell. Flip `quadmeshbatch`.
- [x] **Native transformed images.** Add `mgsk_draw_image` (SkImage raster copy +
      `drawImageRect` with sampling) to the wrapper; implement
      `render.ImageTransformer` on the native path.
- [x] **Native hatching via tiled `SkShader`.** Add a tiled-shader hatch
      entrypoint to the wrapper and route `NativeHatcher` through it instead of
      `render.DrawHatchFallback`; flip `nativehatcher`.
- [ ] **Real GPU mode** (`SkSurface::MakeRenderTarget`) behind `skiagpu` +
      `skiacgo`: add `mgsk_surface_new_gpu` (Ganesh `GrDirectContext` over GL or
      Vulkan) + flush; wire `FlushGPU`; keep deterministic CPU readback for
      goldens. Requires a platform GPU context (GL/Vulkan dev libs). The
      CPU-readback scaffold (`gpu_scaffold_test.go`) already exists.
- [ ] **Per-mode reporting divergence.** Once GPU is real, split
      `ModeCapabilities` (`strategy.go`) so CPU and GPU capability sets differ and
      the live `BackendComparisonReport` shows per-mode native/fallback/unavailable.
- [ ] **Parity hardening.** Add `skiacgo` tests comparing native output to the AGG
      goldens within tolerance (markers/gradients/Gouraud first), and update each
      `NativePathRequirements` row in `strategy.go` from `StatusDeferred` to
      `StatusImplemented` as its primitive lands.

**Exit criterion:**

- [ ] Skia is a viable secondary raster backend: native batch primitives + real
      GPU acceleration available under `skiacgo`/`skiagpu`, with parity-checked
      output and truthful per-mode capability reporting. *(Wrapper + first native
      primitives done; remaining items above.)*

---

# Phase 2: Visual Parity Closure via Code Parity (RMSE ≤ 5)

**Goal:** every catalog case renders within `RMSE 5` of its Matplotlib
reference, or carries a documented, frozen tolerance exception — and the route
there is **code parity**: when a case diverges, find the upstream code path in
`third_party/matplotlib` (3.10.9) responsible for the difference and make the
Go implementation a faithful, idiomatic translation of it. Visual parity then
follows by construction instead of by tuning. The existing bans stand and are
the enforcement half of this principle: no example-source workarounds,
fixture-specific core branches, catalog-ID conditionals, or unexplained
empirical constants (`internal/examplecatalog.ValidationClusters`).

**Status (2026-06-12, after W1+W2, measured `testdata/golden/` vs
`testdata/matplotlib_ref/`, harness metric):** 165 paired cases; **139 at
RMSE ≤ 5**, strict-text cases at exactly 0. The **MathText family is closed**
(`mathtext_integrals` 0.00, `mathtext_matrices` 0.24, `mathtext_gallery` 2.28,
`mathtext_fractions` 4.01; only `mathtext_basic` 5.33 / `mathtext_inline_labels`
5.88 marginally above, and their diffs are general text placement). **The
mplot3d family is closed** (W1) and the **geo/polar/radar family is closed
except `geo_lambert_axes` 5.24 and the toolkit gallery** (W2). **26 cases
remain above RMSE 5**, covered by workstreams W2b–W5.

## Workstreams (ordered by residual; each names the upstream source to translate)

- [x] **W1 — mplot3d structural parity. DONE 2026-06-12** — whole family now
      ≤ 4.6 (`gallery` 22.7→3.4, `tricontourf3d` 13.0→1.7, `contourf3d`
      6.1→0.9, `fill_between3d` 5.3→2.6, `errorbar3d` 5.1→2.6, `terrain`
      4.9→1.0, `surface3d` 4.7→0.8). Four faithful ports, each cited from
      upstream:
      1. `Axes3D.Contourf`/`TriContourf` defaulted alpha to an empirical 0.45;
         upstream forwards kwargs unchanged → opaque (now 1.0).
      2. 3D tick count was a fixed `MaxNLocator{N: 9}`; ported
         `XAxis.get_tick_space` (all 3D axes inherit XAxis) +
         `MaxNLocator._raw_ticks` nbins='auto' clipping → tick density now
         adapts to axes width (`axes3DTickBins`), which closed the gallery and
         most of the family tail.
      3. `FillBetween3D` drew unshaded uniform quads; ported the
         shade-default (quad→true, polygon→false) +
         `art3d._generate_normals`/`_shade_colors` per-quad lightsource
         shading (colors now survive the painter depth sort).
      4. Filled contour bands now render `antialiased=False` like upstream
         `ContourSet`, and — cross-repo, in `../agg_go` — `AntialiasOff` is a
         true binary-coverage mode (`SetAntiAliased`, mirroring AGG
         `scanline_bin`: any touched cell → fully covered pixel) instead of
         the old gamma-0.1 approximation that suppressed partial coverage.
         This also dropped 2D filled-contour cases (`mesh_contour_tri`
         7.1→6.9, `unstructured_showcase` 5.5→4.2, `triangulation_gallery`
         4.3→3.3).
- [x] **W2 — geo/polar/radar projections. LARGELY DONE 2026-06-12** — five
      faithful ports landed: (1) `Axes.Fill` now creates patch-zorder-1
      artists like upstream `fill()` → Polygon (was 2, drew gridlines under
      fills); (2) full-circle radial tick labels anchor ha=left/va=bottom with
      no pad (`PolarAxes.get_yaxis_text1_transform`); (3) geo/polar gridline
      paints use `SnapAuto` so the backend's PathSnapper port (device-frame
      `floor(v+0.5)+0.5`, h/v-only paths) matches AGG — straight Mollweide
      parallels and radar spokes now land on matplotlib's pixel rows;
      (4) the default geo formatter is upstream's `ThetaFormatter` (degrees
      with degree sign, rounded to grid spacing) — fixtures mirror the Python
      references' plain-formatter overrides; (5) `rlabel_position` is stored
      in theta-data space and mapped through theta offset/direction at draw
      time (`RadialTick.update_position`), plus `Axes.SetFrameOn` porting
      `set_frame_on(False)`'s remove-all-spines semantics.
      Results: `radar_basic` 6.9→**0.35**, `geo_mollweide_axes` 5.9→**4.3**,
      `polar_axes` 4.8→4.75, aitoff/hammer 3.9→3.8, `skewt_basic` 3.6→3.1,
      `projection_toolkit_gallery` 20.2→**11.9**.
      Remaining (tracked under W2b below): `geo_lambert_axes` 5.24 (center
      "0" tick label + xlabel fringe), and the gallery's non-geo panels.
- [x] **W2b — gallery leftovers. DONE 2026-06-12** — closed the remaining
      projection/toolkit residuals without example-source workarounds. Final
      focused reference-compare metrics after regolding the affected projection
      cases: `geo_lambert_axes` **0.60**, `axisartist_showcase` **2.66**,
      `projection_toolkit_gallery` **3.93**; neighboring projection cases also
      stay under RMSE 5 (`polar_axes` 4.75, `geo_mollweide_axes` 4.24,
      `geo_aitoff_axes` 3.67, `geo_hammer_axes` 3.67, `radar_basic` 0.35,
      `skewt_basic` 2.87).
      - [x] **W2b.1 — Baseline and classify residuals.** Regenerate/inspect
            `TestReferenceCompare/{geo_lambert_axes,projection_toolkit_gallery,axisartist_showcase}`;
            use the committed diff artifacts under
            `testdata/_artifacts/reference_compare/` and, if needed, an
            env-gated probe in `test/diagnostics_test.go` to record the first
            divergent intermediate values. 2026-06-12 update: direct
            rendered-vs-reference metrics after the first fixes are
            `projection_toolkit_gallery` RMSE 11.29, `geo_lambert_axes` 5.22,
            and `axisartist_showcase` 5.68. The gallery residual is now
            concentrated in the bottom Skew-T/AxisArtist panels; the Lambert
            panel is already about RMSE 3.8 inside the gallery.
      - [x] **W2b.2 — Lambert label fringe.** Fix `geo_lambert_axes` 5.24 by
            translating the relevant upstream geo tick-label/xlabel placement
            from `third_party/matplotlib/lib/matplotlib/projections/geo.py`
            into `core/projection_lambert.go` / `core/geo.go`; expected residual
            is the center `"0"` tick label and xlabel fringe only. 2026-06-12
            investigation: GeoAxes transform, frame fallback, label anchor
            `(260, 46.4444)`, and Matplotlib's `draw_text` bitmap offset all
            matched upstream for the fixture. The proven root cause was a
            floating half-tie in AGG text-image placement: Go computed
            `227.25000000000006 + 1.25`, so round-to-even landed one pixel to
            the right where Matplotlib hit the exact half tie. `pythonRound`
            now normalizes tiny half-tie noise before round-to-even, dropping
            `geo_lambert_axes` to RMSE 0.60.
      - [x] **W2b.3 — Skew-T axes box placement.** Fix the Skew-T panel in
            `projection_toolkit_gallery` by comparing the local skewx projection
            path (`core/skew.go`, `core/axis.go`) against the upstream custom
            skew projection embedded in
            `test/matplotlib_ref/plots/projection_toolkit_gallery.py`; keep the
            gallery example unchanged except for any unavoidable API call cleanup.
            2026-06-12 progress: fixed the default log formatter's Matplotlib
            `minor_thresholds=(1, 0.4)` behavior and the Skew-X y-grid blended
            transform so pressure gridlines span the axes instead of being
            skew-clipped. Closure came from matching AGG-style device-space
            horizontal spine snapping and preserving Matplotlib `Line2D`
            snap-auto on reference-line artists; remaining line/grid differences
            are below the W2b threshold and consistent with renderer-level AA.
      - [x] **W2b.4 — Parasite right-axis placement.** Fix the AxisArtist/Twin
            panel residual in `projection_toolkit_gallery` and
            `axisartist_showcase` by porting host/parasite placement semantics
            from `third_party/matplotlib/lib/mpl_toolkits/axes_grid1/parasite_axes.py`
            and adjacent axisartist code into `core/parasite_axes.go` /
            `core/axis_artist.go`. 2026-06-12 investigation: host/twin limits,
            right-axis styling, text box, legend, and dense ImageGrid layout are
            visually/layout-equivalent to the Python references. The residual
            concentrated in dashed floating axes and legend strokes rather than
            parasite placement. Reference-line dashes now use Matplotlib
            linewidth scaling and snap-auto like `Line2D`, dropping
            `axisartist_showcase` to RMSE 2.66 and helping the gallery reach
            RMSE 3.93.
      - [x] **W2b.5 — Verify and ratchet.** Regold only after the core fixes,
            then run the focused reference compares plus neighboring projection
            cases (`geo_*`, `polar_axes`, `radar_basic`, `skewt_basic`); update
            per-case tolerances only downward or document any frozen exception.
            2026-06-12 verification:
            `go test ./backends/agg -run 'Test(PythonRoundTreatsFloatingHalfTiesLikeMatplotlibTextPlacement|BlendAlphaMaskAppliesTextAlphaAndClips)$'`,
            `go test ./core -run 'TestAxesAx(HLine_UsesBlendedCoordinates|HLine_ScalesDashesLikeMatplotlibLine2D|HLine_UsesMatplotlibLine2DSnap|Line_ClipsToCurrentView)$|TestSpinePixelEndpoints(RightBoundaryUsesMatplotlibPathSnapper|RightBoundaryRoundsPastHalfPixel|HorizontalBoundariesUseDeviceSpaceSnap)$'`,
            targeted `TestGolden` update for the projection cases, and
            `TestReferenceCompare/(geo_lambert_axes|projection_toolkit_gallery|axisartist_showcase|geo_mollweide_axes|geo_aitoff_axes|geo_hammer_axes|polar_axes|radar_basic|skewt_basic)$`.
- [ ] **W3 — ticks, scales, and inset placement.**
      `ticks_scales_formatters_gallery` 17.6, `date_concise_intraday_labels`
      5.8. Focus on the large, structural differences first: tick locations,
      axes rectangles, and text/formatter output. Small line/grid blending
      differences from the Go AGG port (`../agg_go`) are not W3 blockers unless
      they mask a larger geometry or style mismatch. Port targets:
      `lib/matplotlib/{ticker,scale,dates,category,units}.py`, plus
      `Figure.add_axes` / `Axes.set_position` placement semantics.
      2026-06-12 status: RMSE target reached without regolding:
      `ticks_scales_formatters_gallery` 4.759,
      `date_concise_intraday_labels` 4.995. Neighboring unit/date checks:
      `units_categories` 4.443, `units_custom_converter` 0.117,
      `units_dates` 0.862. `TestReferenceCompare` still reports the old
      golden-vs-reference RMSEs until W3.9 regolds and ratchets tolerances.
      - [x] **W3.1 — Baseline, visual triage, and probes.** Regenerate/inspect
            `TestReferenceCompare/{ticks_scales_formatters_gallery,date_concise_intraday_labels}`;
            compare the rendered PNGs visually against
            `testdata/matplotlib_ref/`, and classify each residual by panel:
            locator positions, formatter strings, axes placement, or renderer
            blending. Preserve the example sources. If visual inspection is
            insufficient, add only env-gated diagnostics in
            `test/diagnostics_test.go` that log major tick locations, minor tick
            locations, formatter inputs/outputs, inset rectangles in figure
            coordinates, and final display-space label boxes.
      - [x] **W3.2 — MultipleLocator and AutoMinorLocator positions.** Fix the
            "Major and Minor Locators" panel by comparing
            `core/tick_locators.go`, `core/axis.go`, and `core/axes_ticks.go`
            against `third_party/matplotlib/lib/matplotlib/ticker.py`
            (`MultipleLocator.tick_values`, `AutoMinorLocator.__call__`) and the
            axis view-limit handoff. Add focused unit coverage in
            `core/tick_test.go` for the `0..6`, base-1.5,
            `AutoMinorLocator(3)` case before touching goldens.
      - [x] **W3.3 — Log-scale locator and formatter handoff.** Fix the log panel
            by tracing `SetXScale("log")`, default major locator installation,
            explicit `LogLocator(base=10, subs="auto")` minor ticks, and
            `LogFormatterMathtext(base=10)` labels through
            `transform/scale_registry.go`, `core/axes_scale.go`,
            `core/axes_ticks.go`, `core/tick_locators.go`, and the formatter
            path. Port the relevant behavior from
            `third_party/matplotlib/lib/matplotlib/scale.py` and
            `third_party/matplotlib/lib/matplotlib/ticker.py`; verify with
            `locator_log_minor_threshold_labels`, `scale_log_variants`, and the
            W3 gallery.
      - [x] **W3.4 — Gallery date locator/formatter panel.** Fix the "Date and
            Category Formatters" main axes by tracing `DayLocator`,
            `DateFormatter("%d %b", tz=UTC)`, `Axes.margins(0.04)`, and date
            unit conversion through `core/date_tick.go`, `core/units.go`, and
            `core/axes.go`. Port the relevant behavior from
            `third_party/matplotlib/lib/matplotlib/dates.py`; confirm the
            `01 Feb`, `07 Feb`, `14 Feb`, `21 Feb` ticks land at upstream
            positions before evaluating pixel residue.
      - [x] **W3.5 — Concise intraday date labels.** Fix
            `date_concise_intraday_labels` by porting the shared locator-aware
            label-level selection, zero-format handling, and offset-string
            suppression from
            `third_party/matplotlib/lib/matplotlib/dates.py`
            (`ConciseDateFormatter`) into `core/date_tick.go`. Keep the public
            Go API value-style, but make the formatter observe the full tick
            sequence exactly like upstream.
      - [x] **W3.6 — Category inset axes placement.** Fix the embedded
            "Categories" axes rectangle in `ticks_scales_formatters_gallery` by
            comparing `Figure.AddAxes`, axes bounding-box handling, and any
            locator/position override path in `core/figure.go`, `core/axes.go`,
            `core/axes_locator.go`, and `internal/geom` with upstream
            `Figure.add_axes` / `Axes.set_position`. The target rectangle is the
            Python reference's figure-fraction `go_rect(0.30, 0.16, 0.43, 0.30)`;
            do not compensate by changing the example.
      - [x] **W3.7 — Category tick conversion inside the inset.** Once the inset
            rectangle is correct, verify the categorical bar positions and hidden
            x tick labels against
            `third_party/matplotlib/lib/matplotlib/category.py`. If the bars or
            ticks still differ after placement is fixed, port the category
            mapping/default-unit behavior through `core/units.go` and
            `core/axes.go`, with focused coverage in `core/units_test.go`.
      - [x] **W3.8 — Custom-unit formatter duplication/offset.** Fix the
            custom-unit panel by tracing `PlotUnits` / `ScatterUnits` /
            `AutoScale` through `core/units.go`, `core/axes.go`, and
            `internal/parityutil` test converters, then port the relevant
            conversion and default-unit behavior from
            `third_party/matplotlib/lib/matplotlib/units.py`. Confirm tick
            labels are produced once, at the same display positions as upstream,
            and that `units_custom_converter` still passes.
      - [ ] **W3.9 — Verification and tolerance ratchet.** After the core fixes,
            regold `ticks_scales_formatters_gallery` and
            `date_concise_intraday_labels`; run their focused
            `TestReferenceCompare` targets plus neighboring locator/unit/date
            cases (`locator_*`, `scale_*`, `units_*`, `date_*`,
            `ticks_styling_surface`). Update tolerances only downward unless a
            documented upstream-incompatible exception remains.
- [ ] **W4 — text layout: wrapping and rotated multiline.**
      `text_layout_gallery` 14.6. The diff isolates the residual to: the wrap
      point of `wrap=True` text (display-width logic in
      `Text._get_wrapped_text`), the rotated multiline block, and the
      `rotation_mode="anchor"` box. Port `lib/matplotlib/text.py`
      (`_get_layout`, `_get_wrapped_text`, rotation/anchor handling)
      faithfully; the unrotated alignment grid already matches.
- [ ] **W5 — the 5–7.3 band (≈ 21 cases).** `layout_bbox_helpers` 7.3,
      `plot_variants` 7.2, `axes_convenience_helpers` 7.2, `legend_layout_matrix`
      7.1, `line2d_markers` 7.0, `fill_stacked` 6.9, `specialty_depth` 6.9,
      `mesh_contour_tri` 6.9 (improved from 7.1 by the W1 antialiasing port;
      remaining residual is contour band geometry/labels), `mixed_raster_vector`
      6.4, `arrays_showcase` 6.3, `widgets_gallery` 6.2, `fill_basic` 6.1,
      `annotation_legend_offsetbox_gallery` 6.0, `clip_path_batch` 6.0,
      `mathtext_inline_labels` 5.9, `annotation_composition` 5.8,
      `fill_variants` 5.5, `mathtext_basic` 5.3,
      `specialty_artists` 5.3, `axes_option_breadth_17_75_3` 5.3
      (`unstructured_showcase` dropped to 4.2 via the W1 antialiasing port).
      Expect a few shared root causes (legend/offsetbox layout, fill-edge
      handling, label placement) rather than independent bugs: diagnose each
      diff first, group by root cause, then fix each group as an upstream port
      (`legend.py`, `offsetbox.py`, collection/fill paths).

## Method (code parity, per failing case)

1. Inspect the committed diff artifact
   (`testdata/_artifacts/reference_compare/{id}_golden_vs_matplotlib_ref_diff.png`;
   regenerate with `just parity-viewer-print` or
   `go test ./test -run TestReferenceCompare`). Classify the residual:
   geometry, placement, text, or anti-aliasing.
2. Locate the upstream code path in `third_party/matplotlib`. Instrument both
   sides to find the *first diverging intermediate value* — Python via
   `PYTHONPATH=. python3 test/matplotlib_ref/plots/<id>.py` with temporary
   prints, Go via an env-gated probe in `test/diagnostics_test.go`.
3. Make the Go side a faithful idiomatic translation of the upstream
   computation. Cite the upstream file/function in the commit message so every
   fix carries its provenance.
4. Regold, confirm the metric *and* the visual diff, and check neighboring
   cases for regressions (`TestReferenceCompare` full run).

## Tolerance ratchet (new)

The committed per-case tolerances are far looser than today's actuals (e.g.
`mathtext_gallery` MaxRMSE 55 vs actual 2.3, `widgets_gallery` 120 vs 6.2,
`text_annotation_matrix` 42 vs 4.7), so the suite currently cannot catch large
regressions on already-good cases.

- [ ] After each workstream lands, ratchet the affected rows down to
      ≈ actual + small headroom.
- [ ] End state: closed cases use the package defaults (no per-row override);
      overrides remain only on documented, frozen exceptions.

**Exit criteria:**

- [ ] `TestReferenceCompare` records no catalog case above `RMSE 5` except those
      with a documented, frozen tolerance exception.
- [ ] Every parity fix names the upstream matplotlib code it translates
      (commit-message provenance), and is validated against the visual diff,
      not just the metric delta.
- [ ] Catalog tolerances are ratcheted so a regression of any closed case
      fails CI.

---

# Phase 3: Parity Status Reporting

**Goal:** finish and keep `docs/matplotlib-parity-status.md` as the single
human-readable parity surface, generated from the machine inventories
(`internal/examplecatalog.PublicSurfaceParityRows` over the committed
`test/testdata/parity_surface/upstream_public_surface.json`, plus
`BrowserDemoCoverageRows` / `FeatureCoverageMatrix`).

The doc already exists with Feature Coverage, Browser Demo Coverage, Public
Surface Summary, Closure Owner Summary, and Open Public Surface Rows sections,
and the browser-side CI gates are in place (a `Showcase: true` row without a
browser accounting row, or a browser demo referencing a non-catalog family, both
fail CI). Remaining work is the upstream-family detail and its guard:

- [ ] One table per upstream feature family with columns: upstream API / registry
      item, Go status (`direct-equivalent` / `idiomatic-equivalent` / `partial` /
      `not-started` / `intentional-omission`), local API, parity fixture, user
      example, browser demo, and remaining work — generated, not hand-written.
- [ ] CI fails when an upstream public row or enumerable registry item is tracked
      but unclassified.
- [ ] Every `partial`, `not-started`, and `intentional-omission` row has a
      rationale and a next action.

**Exit criterion:**

- [ ] A developer can open `docs/matplotlib-parity-status.md` and see, per tracked
      upstream feature, whether it is ported and whether it has examples / a
      browser demo, with CI guarding completeness.

---

# Phase 4: Documentation, Performance, and v1.0 Release

**Goal:** make the project consumable by users who have not followed the
development thread, establish performance baselines, and tag a stable v1.0.
*(The Matplotlib migration guide, the backend-selection guide
`docs/backend-selection.md`, the showcase caption/snippet review, the
intentional-divergence "anti-gallery", and the README browser-gallery entry
point are already done.)*

### 4.1 API Documentation

- [ ] Package-level GoDoc passes for every public package, with a worked example
      per package.
- [ ] Hosted documentation site (pkg.go.dev plus a curated landing page on the
      existing GitHub Pages deployment).

### 4.2 Performance Pass

- [ ] Profiling sweep across the catalog: find hotspots that exceed the
      100k-point smoothness goal and the sub-second typical-plot goal.
- [ ] Reusable benchmark suite under `benchmarks/` with CI regression tracking.
- [ ] Documented memory-usage targets and a tuning guide for long-running apps.

### 4.3 Release Readiness

- [ ] Semantic-version policy decision and `CHANGELOG.md` baseline.
- [ ] Final golden / reference regeneration pass with per-case tolerances frozen
      for v1.0.
- [ ] Public API stability audit: rename or hide any symbol not intended for the
      v1.0 surface.
- [ ] CI gate: `just fmt && just lint && just test` plus catalog-driven parity
      checks all pass on the release branch.
- [ ] Tag v1.0.

**Exit criteria:**

- [ ] A new user can install the module, follow the docs, and reproduce every
      showcase plot.
- [ ] The public API surface is documented, audited, and frozen for v1.0.
- [ ] Performance and parity baselines are tracked in CI.

---

# Development Guidelines

## Backend Strategy

- **Primary raster backend:** AGG (`backends/agg/`) — anti-aliased, sub-pixel
  accurate, reference for parity fixtures.
- **AGG port ownership:** if a parity failure is caused by a fundamental
  rasterization, text, path, transform, or blending issue in the Go AGG port,
  fix `../agg_go` rather than adding compensating behavior in this repository.
- **Pure-Go fallback:** GoBasic (`backends/gobasic/`) — dependency-light
  correctness fallback.
- **Primary vector backend:** SVG (`backends/svg/`) — deterministic,
  browser-readable, structurally tested.
- **Publication vector backends:** PDF / PS / PGF.
- **Accelerated raster backend:** Skia (`backends/skia/`) — opt-in CPU and future
  GPU paths.

## Testing Strategy

- Catalog-driven parity tests (`internal/examplecatalog.Case` + `test/`).
- Golden image tests for raster backends, structural diff for vector backends.
- Property-based tests for data ranges and transforms.
- Visual regression against Matplotlib references with documented per-case
  tolerances.
- `go test ./...` runs the full suite; `go test ./test/ -run <id>` runs one
  parity case.

## API Design Principles

- Follow Matplotlib conventions where sensible; document and explain divergences.
- Use functional options for configuration; keep zero-value defaults useful.
- Keep the object-oriented core API first-class; offer `pyplot` as a
  migration-friendly convenience layer.
- Provide escape hatches (renderer access, raw paths) for advanced cases.

## Performance Goals

- Handle datasets up to 100k points smoothly.
- Sub-second rendering for typical plots.
- Memory-efficient for long-running applications and animation.

## Examples-Driven Development

- Every feature gets a working example tied to the catalog.
- Examples serve as integration tests and gallery content.
- Showcase examples appear in the WASM browser gallery and the README.
- Examples demonstrate real-world usage rather than minimal API smoke tests.

---

This roadmap reflects the work remaining to bring matplotlib-go to a stable,
documented v1.0 release. **Phase 1** hardens the backend matrix (mostly the
deferred external Skia binding); **Phase 2** closes the remaining visual parity
gap via code parity with upstream (chiefly mplot3d, projections, ticks/scales,
and text layout — MathText is closed); **Phase 3** finishes and guards the
parity status report; **Phase 4** delivers documentation, performance
baselines, and the v1.0 release.

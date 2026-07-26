# Matplotlib-Go Remaining Work

This file tracks only unresolved work required for a stable v1.0 release.
Completed implementation history is available in git.

## Execution Order

1. Phase 1 completes Skia GPU support and can proceed independently.
2. Phase 2 performs the breaking API and package rework.
3. Phase 3 runs visual QA against the final pre-v1.0 renderer/API state.
4. Phase 4 freezes, validates, and tags v1.0.

## Phase 1: Backend Deepening (Skia Native + GPU)

**Goal:** make Skia a parity-checked secondary raster backend with truthful CPU
and GPU capability reporting.

The native C-ABI bridge and CPU primitives already exist under
`-tags "skia skiacgo"`. Continue with `just build-skia-native` and
`just test-skia-native` (`SKIA_ROOT` defaults to `/mnt/projekte/Code/skia`).
Native recipes must omit the `freetype` tag unless Skia was built with
`skia_use_freetype=false`; otherwise duplicate FreeType symbols crash at runtime.

- [x] Add real GPU surfaces behind `skiagpu` + `skiacgo`: expose
      `mgsk_surface_new_gpu`, create a Ganesh `GrDirectContext` over GL or
      Vulkan, wire `FlushGPU`, and retain deterministic CPU readback for
      goldens. Reuse `gpu_scaffold_test.go`.
- [x] Split `ModeCapabilities` so CPU and GPU modes report distinct
      native/fallback/unavailable capabilities through
      `BackendComparisonReport`.
- [x] Add `skiacgo` parity tests against AGG goldens, beginning with markers,
      gradients, and Gouraud triangles; promote each verified
      `NativePathRequirements` row from `StatusDeferred` to
      `StatusImplemented`.

Completed 2026-07-25: the GPU tier uses a headless EGL/OpenGL Ganesh render
target with synchronized RGBA readback and native CPU fallback when a context is
unavailable. `just test-skia-native` and `just test-skia-gpu-native` cover the
native primitive/parity matrix and runtime capability split.

**Caveat found in Phase 3.1 (2026-07-26):** the parity half of the third item was
verified with a broken metric. `imagecmp`'s PSNR accumulator overflowed, so nine
Skia-vs-AGG comparisons that reported passing the harness's 22 dB floor were in
fact failing it — `line2d_markers` at RMSE 63, the MathText family at 21-30, and
`patch_showcase`, `mesh_contour_tri`, `pattern_gradient_effects` between. Those
bounds are now stated in RMSE and pinned to measured behavior so a further
regression fails, but they are not evidence of parity: the
`StatusDeferred`→`StatusImplemented` promotions need revalidation. Note also that
`gouraud_triangles`, `pcolormesh_gouraud`, and `mathtext_accents` skip in that
harness for want of a Go figure factory, so they were never measured at all.

**Done when:** native batch primitives and real GPU acceleration are available
under `skiacgo`/`skiagpu`, output is parity-checked, and per-mode capability
reporting reflects runtime behavior.

## Phase 2: Go-Idiomatic API Rework and `core/` Split

**Goal:** perform the single coordinated pre-v1.0 breaking pass: make the API
Go-idiomatic, split the `core/` package, and re-freeze the public surface.
Rendering must remain byte-identical to the pre-break golden baseline.

- [x] **2.1 Surface tiering.** Classify all 3,102 frozen symbols as keep,
      demote, or delete, deciding explicitly on Python-style introspection,
      `*Units` variants, and backend-only renderer extensions.
- [x] **2.2 Package split.** Move 3D axes and projection into `plot3d`, tick
      locators/formatters and date ticks into `ticker`/`dates`, and widgets and
      selectors into `widgets`, keeping builds, tests, and goldens green through
      each move.
- [x] **2.3 Idiomatic conventions.** One error convention (`(T, error)` for
      rejected input, `diag.Warnf` only for accepted degradations, all four
      `*Units` folds); one options model (exactly one options value per entry
      point, `optional.Value[T]` fields, typed enum constants); `Figure.Save`/
      `WriteTo`/`Image`; `GetX()` renamed to `X()`; one public writer per artist
      field; a documented concurrency contract; and shared alpha,
      option-unpacking, and scalar-map resolution paths.
- [x] **2.4 Re-freeze.** Regenerate the frozen API, the public-surface
      classifications, and the parity-status document; record migration notes
      for every break and the Phase 4 changelog fragment.

**Done when:** `core/` no longer owns plot3d/ticker/widgets; plot methods use
the chosen error and options conventions; no raw-string option enums remain;
the new figure save surface is used by examples; the API is re-frozen; and all
goldens remain byte-identical to the pre-break baseline.

Completed 2026-07-26. Every golden and Matplotlib-reference fixture stayed
byte-identical through the whole phase. The final freeze is 3,178 symbols across
29 packages in `test/testdata/public_api/stable_public_api.json`; that artifact
is the surface Phases 3 and 4 are measured against, and
`docs/plans/phase2-freeze-delta.md` reconciles every symbol of it against the
Phase 2.1 tiering decisions. The pass was a net deletion:
all 205 variadic option tails and all 408 pointer-to-primitive option fields are
gone, `internal/optarg` with them. Design records live in `docs/plans/`
(`phase2-public-api-tiering.md`, `phase2-warn-and-skip-inventory.md`,
`phase2-extra-option-rejection.md`, `phase2-options-model.md`,
`phase2-mutable-fields.md`, `phase4-changelog-draft.md`), user-facing breaks in
`docs/matplotlib-migration-notes.md`, the concurrency contract in
`docs/concurrency.md`, and the refreshed large-file inventory in
`docs/large-file-decomposition.md`.

### Phase 2 Follow-ups

Deliberately deferred out of Phase 2 because each changes behavior or scope that
the phase's byte-identical-rendering rule excluded. Neither blocked the "done
when" criteria above. Both are now closed.

- [x] Reconcile the plot3d/core scalar-map range divergence. Neither guard was
      right: Matplotlib's `ContourSet._process_colors` ends in
      `norm.autoscale_None(self.levels)`, so the mapping autoscales over the
      levels — per limit, and for line contours as well as filled ones.
      `core.ResolveContourScalarMap` now expresses exactly that and both
      packages call it; `plot3d/contourf.go`'s duplicated color-override block
      folded into `core.ContourFillScalarMap`. Two goldens moved
      (`mplot3d_terrain`, `mplot3d_tricontour3d`), both **closer** to their
      Matplotlib references (RMSE 1.159 → 1.015 and 2.500 → 2.443).
- [x] Reconcile the frozen-surface symbol delta against the tiering
      classification. Walked all 402 removed and 478 added names; the residue is
      empty. All 19 `delete` rows are gone and the one `demote` row landed in
      `backends`. The growth (3,102 → 3,178) is 73 symbols of field-to-name
      trades the audit only counts in one direction, 30 symbols of pre-existing
      API the baseline collector never parsed (build-tagged FFmpeg and native
      Skia), 19 helpers the package split forced open, and follow-up 1's two
      contour helpers. Three previously unrecorded findings: `core.WidgetArtist`
      and `Axes.AddWidget` were deleted while classified `keep` (now in the
      migration notes), `Renderer.Image` kept its symbol id while changing
      meaning, and the freeze is 3,178 rather than 3,176. Recorded in
      `docs/plans/phase2-freeze-delta.{md,json}` and enforced by
      `TestPublicAPIFreezeDeltaIsReconciled`, which pins both inputs by hash so
      the delta cannot grow silently before Phase 4 tags the surface.

## Phase 3: Visual QA and Tolerance Closure

**Goal:** inspect every case whose tolerance can hide a visible divergence,
record its disposition, and hand Phase 4 a defensible frozen tolerance set.
Run after Phase 2.

Phase 2's byte-identical-rendering rule has ended: Phase 3 fixes are expected to
move goldens, and Phase 4 regenerates them. Each moved golden must be justified
by a reference-compare improvement in the audit.

### 3.0 Audit (done 2026-07-26)

All 190 catalog cases were measured golden-vs-reference and classified by
residual _shape_, not just magnitude. Results in
`docs/plans/phase3-tolerance-audit.{md,json}`, regenerated by
`python3 docs/plans/generate_phase3_tolerance_audit.py` (standalone: reads only
committed PNGs plus the catalog, no cgo). This audit **replaces the per-case
queue this phase previously carried** — most of the cases it named turned out to
be near-identical to Matplotlib with a loose tolerance, while cases it never
mentioned (`basic_line`) carry real divergences. Findings:

- 21 cases are pixel-identical; 124 differ on under 1% of the frame.
- 14 cases have a residual that one integer pixel offset cancels — placement
  bugs, in four offset families.
- After 3.1's fold, 44 cases allow at least 3x their measured residual and 31
  allow under 1.15x. (Before it: 66 and 23, gated additionally by PSNR.)
- `imagecmp.ComparePNG` computed its PSNR sum as `float64(diffR*diffR)` in
  `uint8` arithmetic, so every square above 15 wrapped mod 256. 177 of 190 cases
  reported a corrupted, always-flattering PSNR, and every `MinPSNR` floor in the
  catalog had been calibrated against that broken number. Fixed in 3.1.
- Whole-image RMSE cannot bound a localized error: `basic_line`'s residual is
  one y tick label rendered a pixel low — 139 pixels reaching per-channel
  difference 249 — and still scores RMSE 2.49 against a 2.8 allowance.

- [x] **3.1 Fix the comparison metric before ratcheting anything against it.**
      Done 2026-07-26. `imagecmp.ComparePNG` now accumulates squared error once,
      in `float64`, and derives `PSNR = 20*log10(255/RMSE)`; the separate
      accumulator that squared `uint8` differences in `uint8` arithmetic is gone.
      `TestComparePNG_LargeDiffDoesNotOverflowPSNR` pins the boundary at 15/16
      LSB and `TestComparePNG_PSNRIsDerivedFromRMSE` pins the identity.

      Because PSNR is now provably a restatement of RMSE, `Case.MinPSNR` was
      removed along with all 47 row-level values, and every PSNR gate in the tree
      was restated in RMSE. No gate was allowed to loosen silently: of the 190
      cases, 65 floors never bound (`MaxRMSE` was tighter), 60 bound and were
      reachable and were folded into `MaxRMSE` at their exact equivalent
      (`255/10^(dB/20)`), and 65 were unreachable — those cases were only ever
      passing because the metric flattered them, so their floors were dropped for
      3.6 to re-derive from measurement. The reference-compare harness now gates
      on RMSE and MeanAbs alone; MeanAbs is kept because L1 and L2 residuals are
      genuinely independent.

      Four gates outside the catalog needed recalibration rather than
      translation, having been unreachable under correct math:
      `mathtext_vector_raster_test.go` (external PDF/PS/SVG rasterizers vs AGG
      goldens; measured RMSE 19.3-32.1, bound 40), `skia_render_test.go` MathText
      (measured 19.1-30.2, bound 35), and `backends/skia/parity_test.go`'s
      default plus nine per-case bounds. That last one is a finding worth
      flagging: nine Skia-vs-AGG comparisons had been failing the harness's 22 dB
      default all along, `line2d_markers` by a wide margin at RMSE 63. **Phase
      1's `NativePathRequirements` promotions from `StatusDeferred` to
      `StatusImplemented` were therefore signed off against a broken metric and
      need revalidation** — the new ceilings pin current behavior so a further
      regression still fails, but they are not evidence of parity.

      The strict text cases came through unchanged: `text_labels_strict` measures
      RMSE 0.028 and `title_strict` 0, both far inside their translated bounds.

- [ ] **3.2 Add a localization gate that RMSE cannot express.** Extend the
      reference-compare harness with a residual-shape bound — differing-pixel
      count and largest-cluster size are already computed by the audit
      generator; port them to `imagecmp` and give `examplecatalog.Case` the
      corresponding per-case fields. This is the gate that catches a wholly
      misplaced glyph, and the reason `basic_line` currently passes.
- [ ] **3.3 Fix the four shift families.** Each is a candidate shared root
      cause; confirm against `third_party/matplotlib` 3.10.9 and fix in the core
      library, not the fixtures.
  - `+0+1` (7 cases, text baseline rounding one pixel low; note only _some_
    labels in a figure shift, so this is a tie-break at `.5`, not a uniform
    offset): `basic_line`, `basic_line_labels`, `mathtext_gallery`,
    `formatter_engineering_labels`, `matshow_basic`, `mathtext_accents`,
    `specgram_psd`.
  - `-1+0` (4 cases, horizontal text advance/placement): `mathtext_fractions`,
    `formatter_scalar_scientific_labels`, `animation_subplots_frame`,
    `quad_mesh`.
  - `+1+0` (2 cases, image/colorbar edge column off by one — in
    `colormap_diverging` the entire residual is one 258-px-tall column at the
    axes edge): `colormap_diverging`, `image_heatmap`.
  - `+2-1` (1 case): `transform_coordinates`.
- [ ] **3.4 Investigate the localized non-shift divergences.** Cases whose
      residual is confined but not a clean offset, worst first: `line2d_markers`,
      `stat_variants`, `annotation_legend_offsetbox_gallery`, `imshow_clipped`
      (a single cluster covering the whole image area — a resampling offset in
      the clipped-image path), `axes_control_surface`, `text_annotation_matrix`,
      `legend_layout_matrix`, `transform_annotation_modes`, `patch_style_matrix`,
      `unstructured_showcase`, `mathtext_basic`, `colorbar_boundary_values`,
      `line2d_semantics`. Fix or record as an accepted difference.
- [ ] **3.5 Disposition the dense residuals.** These are broad and
      low-amplitude, where RMSE _is_ the right metric; the question is only
      whether the amplitude is defensible. `geo_aitoff_axes`,
      `geo_hammer_axes`, `geo_lambert_axes`, `geo_mollweide_axes` (6-8% of the
      frame, p99 amplitude ~13 — graticule/background antialiasing);
      `colormap_cyclic` (19% of the frame); `imshow_bilinear` (1,117 regularly
      spaced 5x5 clusters — an interpolation sample-center offset);
      `imshow_interpolation_matrix`, `imshow_transformed`, `clip_path_batch`,
      `pcolormesh_gouraud`, `skewt_basic`, `projection_toolkit_gallery`,
      `image_variants_gallery`, `ticks_styling_surface`, `widgets_gallery`.
- [ ] **3.6 Ratchet every tolerance and enforce the ratchet.** Set each case's
      bound from its post-fix measured value plus stated headroom, in both
      directions: tighten the 44 loose cases (worst offenders
      `linecollection_linestyle` 179x, `spy_image` 124x, `errorbar_capthick` 64x,
      `bar_basic_frame` 50x, `scatter_plotnonfinite` 37x, `dashes` 36x) and
      relieve the 31 cases now under 1.15x slack. Fragility there is not a flake
      risk — `TestReferenceCompare` relates two committed files — but it does
      mean any fix in 3.3-3.5 will trip a neighbouring case, so ratchet after
      those land, not before. Also re-derive the 65 bounds whose PSNR floor 3.1
      had to drop as unreachable, and the Skia ceilings 3.1 pinned to current
      behavior. Add a meta-test that regenerates the audit and fails when any
      tolerance drifts from measured-plus-headroom, so the set cannot loosen
      silently before Phase 4 freezes it.
- [ ] **3.7 Commit the per-case disposition table.** One row per case:
      `fixed`, `documented exception`, or `upstream difference`, carrying the
      measured residual it was signed off at. Fold it into the audit document so
      it regenerates with the numbers rather than drifting from them.

**Done when:** the comparison metric is correct, no catalog case has an
effectively disabled or flake-prone gate, every case has a written disposition,
and Phase 4 receives the ratcheted tolerance set plus its enforcing test.

## Phase 4: v1.0 Release

**Goal:** publish a reproducible, documented, semantically versioned v1.0 after
Phases 1–3.

- [ ] Decide and document the semantic-version policy; establish the
      `CHANGELOG.md` baseline and include Phase 2's breaking changes.
- [ ] After Phase 3, regenerate final goldens and Matplotlib references and
      freeze per-case tolerances.
- [ ] Re-run the post-Phase-2 public API freeze and confirm the exported
      surface, migration notes, and parity-status documentation agree.
- [ ] Make the release-branch gate fully green:
      `just fmt && just lint && just test`, plus catalog-driven parity checks.
      Resolve or reclassify every entry in
      `docs/ci-known-test-failures.md`.
- [ ] Verify a new user can install the module, follow the documentation, and
      reproduce every showcase plot.
- [ ] Confirm performance and parity baselines are active in CI.
- [ ] Tag v1.0.

## Global Gates

- Prefer core-library fixes over fixture tweaks when Matplotlib output differs.
- Use `third_party/matplotlib` 3.10.9 as the behavioral authority.
- Keep parity cases catalog-driven through `internal/examplecatalog.Case`.
- Run `just fmt`, `just lint`, and `just test` for each completed work unit.
- If a parity failure belongs to the Go AGG rasterizer, fix `../agg_go` rather
  than compensating in this repository.

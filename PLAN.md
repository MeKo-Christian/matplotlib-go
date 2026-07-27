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
byte-identical through the whole phase. The final freeze is 3,181 symbols across
29 packages in `test/testdata/public_api/stable_public_api.json`; that artifact
is the surface Phases 3 and 4 are measured against, and
`docs/plans/api-freeze-delta.md` reconciles every symbol of it against the
Phase 2.1 tiering decisions. The pass was a net deletion:
all 205 variadic option tails and all 408 pointer-to-primitive option fields are
gone, `internal/optarg` with them. Design records live in `docs/plans/`
(`api-tiering.md`, `warn-and-skip-inventory.md`,
`extra-option-rejection.md`, `options-model.md`,
`mutable-fields.md`, `changelog-draft.md`), user-facing breaks in
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
      classification. Walked all 402 removed and 480 added names; the residue is
      empty. All 19 `delete` rows are gone and the one `demote` row landed in
      `backends`. The growth (3,102 → 3,181) is 73 symbols of field-to-name
      trades the audit only counts in one direction, 30 symbols of pre-existing
      API the baseline collector never parsed (build-tagged FFmpeg and native
      Skia), 19 helpers the package split forced open, follow-up 1's two
      contour helpers, 3.3.4's `transform.AffineScale`/`IsAffineScale` pair, and
      3.3.5's `core.Figure.CanvasSize`. Three previously unrecorded findings: `core.WidgetArtist`
      and `Axes.AddWidget` were deleted while classified `keep` (now in the
      migration notes), `Renderer.Image` kept its symbol id while changing
      meaning, and the freeze is 3,181 rather than 3,176. Recorded in
      `docs/plans/api-freeze-delta.{md,json}` and enforced by
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
`docs/plans/tolerance-audit.{md,json}`, regenerated by
`python3 docs/plans/generate_tolerance_audit.py` (standalone: reads only
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

- [x] **3.2 Add a localization gate that RMSE cannot express.** Done
      2026-07-26. `imagecmp.DiffResult` now carries `DiffPixels`, `Clusters` and
      `LargestCluster`: the pixels whose largest per-channel difference exceeds
      the tolerance, the 8-connected components they form, and the size of the
      biggest one. Connectivity matches the audit generator's 3x3 structuring
      element, and the ported numbers reproduce it exactly — `basic_line`
      139/4/73, `specgram_psd` 435/4/353, `line2d_markers` 3385/167/2055,
      `imshow_clipped` 2298/1/2298, `colormap_diverging` 259/2/258, all
      identical to `tolerance-audit.json`.

      `examplecatalog.Case` gained `MaxDiffPixels` / `MaxLargestCluster` (zero =
      unbounded), gated in `runReferenceCompareTest` after the amplitude
      gates and reported in its per-case log line.
      `TestComparePNG_ShapeSeesWhatRMSECannot` pins the motivating case: 36
      black pixels on a 640x360 white frame score RMSE 2.76 — inside
      `basic_line`'s 2.8 allowance — while reading as one 36-px cluster.

      Bounds are populated for the 27 localized cases named in 3.3 and 3.4, at
      1.25x their measured values rounded up (`basic_line` 180/92,
      `colormap_diverging` 330/330, `line2d_markers` 4300/2600). The remaining
      163 cases are still unbounded on shape: their residuals are dense, where
      the largest cluster ranges to 42,872 px (`colormap_cyclic`) against a
      median of 87, so a single default would be meaningless — 3.6 sets them
      per case from post-fix measurements.
- [ ] **3.3 Fix the four shift families.** Each is a candidate shared root
      cause; confirm against `third_party/matplotlib` 3.10.9 and fix in the core
      library, not the fixtures. **7 of 14 fixed, and an eighth largely fixed,
  2026-07-27**; the audit's shift families are down from 14 cases to 7 and its
  pixel-identical count up from 21 to 27. First pass (`basic_line` and
      `basic_line_labels` are now pixel-identical to Matplotlib;
      `animation_subplots_frame` went from 178 differing pixels to 38, largest
      cluster 70 to 2). Two root causes, both the same mistake — a tolerance
      applied to a decision Matplotlib makes exactly:
  - **The rounding tolerances were removed.** `pythonRound` snapped anything
    within 1e-9 of `.5` onto the tie before rounding half-to-even, and
    `snapPathCoordinate` was `floor(v + 0.5 + 1e-9)` where matplotlib's
    `PathSnapper::vertex` is plain `floor(v + 0.5)`. Both are now exact. This is
    only correct because the port's text origins already match matplotlib
    bit-for-bit, including its float64 noise: `basic_line`'s "0.6" y tick label
    arrives at 149.49999999999997 in *both* — the locator emits `3*0.2 ==
    0.6000000000000001` in each — and Python rounds that down. The epsilon
    rounded it up, drawing that one label a pixel low while its siblings, which
    land on exact `.5`, stayed put. That is why only _some_ labels shifted.
  - **`transData` is now composed the way matplotlib composes it.**
    `core.Axes.ensureTransforms` flattens `transLimits x transAxes` into a single
    affine before mapping any vertex (`transform.AsAffine` on the chain, which is
    the same matrix product numpy performs), instead of evaluating the two legs
    in sequence. Staged evaluation is the same map in exact arithmetic but not in
    float64: `units_categories`' bar edge came out 267.49999999999994 staged
    against exactly 267.5 composed, and the exact snapper then moved that edge a
    pixel left. Removing the snap epsilon without this would have regressed
    `units_categories` from 104 differing pixels to 746.
  - All 15 goldens the two fixes moved improve against the matplotlib reference,
    none regress: `pattern_gradient_effects` 2796 px to 697,
    `annotation_legend_offsetbox_gallery` RMSE 3.51 to 1.43,
    `axes_control_surface` 3.38 to 0.69, `spectrum_variants` 1.31 to 0.18.
    `basic_line`, `basic_line_labels` and `animation_subplots_frame` had their
    `MaxDiffPixels`/`MaxLargestCluster` ratcheted onto the new measurements so a
    re-regression fails; the rest wait for 3.6. The PDF and SVG goldens were
    regenerated for the last-decimal coordinate change (they are Go-authored, not
    matplotlib references).
  Subtasks:
  - [x] **3.3.1 Remove the rounding tolerances** (`pythonRound`,
        `snapPathCoordinate`). Done 2026-07-27, described above.
  - [x] **3.3.2 Compose `transData` as a single affine.** Done 2026-07-27,
        described above. Required by 3.3.1: without it `units_categories`
        regresses from 104 differing pixels to 746.
  - [x] **3.3.3 The mathtext cases.** Done 2026-07-27. `mathtext_fractions`
        (84 differing pixels) and `mathtext_accents` (8) are now **pixel-identical**
        to matplotlib and `mathtext_gallery` went 210 to 64, its residual now a
        single cluster. Three root causes, all the same species as 3.3.1 — an
        expression that is right in exact arithmetic but not in float64, exposed
        because `_mathtext.Output.to_raster` blits with `int()`, which truncates
        rather than rounds, so one ULP becomes one whole pixel whenever a
        coordinate lands on an integer:
    - `mathtext/sqrt.go` laid the root index out **at** the shrunk size, where
      matplotlib measures at full size and scales the resulting box
      (`Char.shrink` multiplies an already-measured width). A glyph re-measured
      at 11.27pt is not 0.49x itself at 23pt — the trailing advance-minus-ink
      kern and the 26.6 char-size quantization do not scale linearly. This put
      every glyph after the index in `\sqrt[3]{x + 1} + \sqrt{y}` 0.035px left of
      matplotlib, which was invisible for all of them except the one `+` whose
      `int(ox)` crossed 112/113.
    - `mathtext/frac.go` summed the numerator/rule/denominator stack height in
      its own order; `Vlist.vpack` walks the children doing `x += d + p.height`,
      carrying the previous child's depth, and `vlist_out` then places the
      numerator at `(shift - height) + cnum.height`. Now accumulated in
      matplotlib's order and grouping.
    - `backends/agg/freetype_mathtext_native.go` computed the glyph row as
      `(boxAscent + oy) - ymin`, where `ship` forms `off_v = oy + box.height`
      **first** and adds it to the accumulated `cur_v` as one term. Factored into
      `mathShipOffsetV`.

      Note the two `formatter_*` cases the audit had grouped here are **not**
      shift cases at all — at zero tolerance they carry 527 and 188 clusters that
      no integer offset cancels. They belong in 3.4.

      Two of the three fixes are in `github.com/cwbudde/mathtext`, a tagged
      dependency rather than repo code; they shipped as **mathtext v0.5.1** and
      `go.mod` requires that version. No `replace` and no new build
      prerequisite — the build stays self-contained.
  - [x] **3.3.4 `matshow_basic`.** Done 2026-07-27. Now **pixel-identical** to
        matplotlib (family `identical`; 20 differing pixels to 0, RMSE 1.25 to
        0.038, max amplitude 241 to 1 — the whole remaining residual is uniform
        ±1 LSB colormap rounding over the image area). The guess above was wrong
        in its cause: the tick path does go through `transData`, but for this
        case `transData` was not the single affine 3.3.2 built. The y axis is
        inverted (`matshow` puts the origin at the top), and `core.invertedScale`
        wraps the linear scale rather than swapping its domain, so
        `transform.affineAxis` — which only accepted a concrete
        `transform.Linear` — reported the graph non-affine and
        `ensureTransforms` fell back to the staged `Chain`. Staged, the y=2 tick
        landed on exactly 148.5 where matplotlib reaches 148.50000000000003; the
        AGG marker blit truncates (`floor(v + 0.5)`), so that one ULP moved the
        tick a whole pixel. The other three ticks differ in the same ULP but do
        not straddle an integer, which is why only one was visible.

        Fixed by giving `Scale` an affine-flattening capability:
        `transform.AffineScale` / `transform.IsAffineScale`, consulted by
        `affineAxis` for any scale that is not the built-in `Linear`.
        `invertedScale` implements it. The flattened axis is rebuilt from
        `Domain()` with `NewLinearAxis`, which is matplotlib's
        `BboxTransformFrom(viewLim)` arithmetic — matplotlib has no inversion
        wrapper at all, `invert_yaxis()` just swaps `viewLim`, and
        `invertedScale.Domain()` already returns the swapped endpoints. Having
        the wrapper return its own coefficients instead (`-scale`, `1-offset`
        over the unswapped domain) would be algebraically equal but not
        bit-equal, which is the whole point here. An inverted *log* axis still
        reports non-affine and keeps the staged path.

        Two new exported symbols, so the frozen API audit was regenerated.
        Tolerances ratcheted 1.608/25/25 to 0.1/4/4; verified the ratchet bites
        by restoring the pre-fix golden (fails at MaxDiff=241). No other golden
        moved — `-update-golden` across all 190 cases changed only this one.
  - [x] **3.3.5 `specgram_psd` and `quad_mesh`.** Done 2026-07-27. The probe
        said they are **not** one family, so they were split: `quad_mesh` is
        fixed and pixel-identical, `specgram_psd` moved to 3.3.6.

        `quad_mesh` was never a mesh case. Its entire 70-px residual is the
        `"0"` x tick label — the glyph bitmap is byte-identical, drawn one pixel
        left. Root cause: **matplotlib's figure pixel size is derived, not
        given.** A figure asked for in pixels is `figsize=(px/dpi, px/dpi)`, and
        the pixel extent is the product back: `9.8*100 == 980.0000000000001`,
        not 980. `core.NewFigure` cast the integer directly, so `Axes.layout`
        (`SizePx * RectFraction`) put the axes left edge at exactly 98.0 where
        matplotlib has 98.00000000000001. The centred label origin then landed
        on exactly 94.5 against matplotlib's 94.50000000000001, and the AGG text
        blit rounds ties-to-even: 94 against 95. Same species as 3.3.1 and
        3.3.4 — a tie in the port that is one ULP past the tie in matplotlib.

        `NewFigure` now derives `SizePx` through `figureSizePx(w, h, rc.DPI)`.
        Blast radius is bounded by arithmetic: only sizes that are not exact
        multiples of the dpi are affected, and across the whole repo that is
        980x620, 980x720 and 930x340 — 640, 360 and 620 all round-trip exactly.
        `-update-golden` over all 190 cases moved exactly three, all **closer**
        to matplotlib: `quad_mesh` 70 differing pixels to **0** (RMSE 1.347 to
        0.497; family `shift -1+0` to `identical`), `mixed_collection` 538 to
        468 with its largest cluster 70 to 10 (RMSE 1.575 to 0.956 — the same
        tick label), `clip_path_batch` 56,880 to 56,810 (RMSE 1.430 to 0.691).
        The audit's pixel-identical count is up 26 to 27 and its shift families
        down 8 to 7. Tolerances ratcheted on all three; verified the
        `quad_mesh` ratchet bites by restoring the pre-fix golden.

        Two consequences worth recording. The derived size can sit an ULP
        *below* the requested integer (402 px at 100 dpi is 401.99999999999994),
        so truncating it to size a canvas allocates a pixel short; the new
        `Figure.CanvasSize` rounds instead and is the one place that decision
        lives. Matplotlib itself truncates here — `figsize=(4.02, 2)` at 100 dpi
        really does save a 401-px PNG — but no fixture depends on that and
        copying it would change committed image dimensions for no parity gain.
        `CanvasSize` is one new exported symbol, so the frozen API audit and the
        freeze-delta artifact were regenerated (3,180 to 3,181). The
        Go-authored `usetex_golden/basic.png` also moved, by a clean +1 px on
        every cluster: its figure is 220 wide and 220/100*100 is
        220.00000000000003, so the centred TeX box crossed the same tie.
  - [ ] **3.3.6 The image nearest-resample boundaries**: `colormap_diverging`,
        `image_heatmap`, `specgram_psd`.
        In `colormap_diverging` the entire residual is one 258-px-tall column at
        an interior cell boundary, not the axes edge. A first attempt failed and
        was reverted: reproducing AGG's 1/256 fixed-point span interpolation in
        `scaleRGBANearest` (`floor(round(u*256)/256)` rather than
        `round(u-0.5)`) left both cases untouched and regressed
        `colorbar_composition` from 2563 to 2821 differing pixels, so these two
        do not reach the raster nearest-neighbour path that was changed. Find
        the path they do take before trying again.

        `specgram_psd` joined this group from 3.3.5, where the probe showed it
        is the same defect rather than the mesh-edge case the roadmap had
        guessed. Its 435-px residual is 4 clusters: one 353-px horizontal row at
        y=281 spanning x[81:434] (`+0+1`) and two 1-px-wide vertical strips at
        x=481 (`+1+0`). All three sit on *source-cell boundaries of the
        spectrogram image*: at x=481 the port still emits the left cell's colour
        where matplotlib has already switched to the right cell's, and
        identically across y=281. Only 2 of the image's ~86 cell boundaries
        flip, which is what a tie in the dest→src index map looks like rather
        than a systematic off-by-one. (The earlier roadmap text called these
        "three 1-px-wide vertical strips … a mesh column edge"; the dominant
        cluster is in fact the horizontal row, and nothing here is a mesh edge.)
  - [ ] **3.3.7 `transform_coordinates`** (`+2-1`, the only two-axis shift).
  - [ ] **3.3.8 The remaining `mathtext_gallery` cluster** (64 px, one 10x11
        glyph at x[638:648] y[176:187], family `+0-1` after 3.3.3). It is the
        last shift residual in the mathtext family and did not respond to any of
        the three fixes above, so it has a fourth cause.
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

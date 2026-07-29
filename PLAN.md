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
      library, not the fixtures. **11 of 14 fixed, and a twelfth largely
      fixed, 2026-07-28**; the audit's shift families are down from 14 cases to 3
      and its pixel-identical count up from 21 to 29. First pass (`basic_line` and
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
    arrives at 149.49999999999997 in _both_ — the locator emits `3*0.2 ==
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

  - [x] **3.3.6 The image nearest-resample boundaries**: `colormap_diverging`,
        `image_heatmap`, `specgram_psd`. Done 2026-07-28. The first two are now
        **1 differing pixel** each (from 259 and 270); `specgram_psd` went 435 to
        354, and what is left of it is not this defect. `imshow_clipped` fell out
        of 3.4 and is now **pixel-identical** (2298 to 0, RMSE 2.938 to 0.051),
        and `lognorm_imshow` 1270 to 599.

        The reverted first attempt looked in the wrong function. These images
        never reach `backends/agg.scaleRGBANearest`: `Image2D.Draw` resamples
        the scalar grid **in `core`, in Go**, to the destination pixel size
        (`rasterizeForDestination`) and hands AGG an already-dest-sized raster
        tagged `"nearest"`, so AGG's own pre-scale short-circuits on the size
        match. Leave `scaleRGBANearest` alone here — it is a third, different
        rule serving true-colour images and downscales, and it is what regressed
        `colorbar_composition` last time.

        Three separate defects, all in `core/image.go`:

        **The column rule is a DDA, not arithmetic.** The port mapped each
        destination column independently with `floor(u)`. AGG does not map
        pixels independently at all: `span_interpolator_linear::begin`
        transforms only the two *ends* of the scanline, rounds each to 1/256 of a
        source cell with round-half-away-from-zero, and walks between them with
        `dda2_line_interpolator` — integer arithmetic whose accumulated remainder
        decides where each cell boundary lands. No per-pixel formula reproduces
        it. `floor(u)` is wrong on `colormap_diverging` and `image_heatmap`;
        `floor(round(u*256)/256)`, which is the per-pixel reading of the same
        C++ and what a first pass at this item shipped, is wrong on
        `asinh_norm_image`, `twoslope_norm_image` and `lognorm_imshow`. Both
        rules agree with the DDA over most of a span and diverge at boundaries,
        and one boundary is one whole mis-drawn column. Verified by recovering
        the column runs from matplotlib's own resampled buffer for six fixtures:
        the DDA matches all six, `floor` three, per-pixel rounding three.

        **The row rule is the per-pixel rounding.** An axis-aligned image has the
        same source row at both ends of a scanline, so the interpolator's y is
        constant and reduces to the single `floor(round(u*256)/256)`. Here plain
        `floor` really is wrong, and it is what the port used. Confirmed against
        the same buffers. Coupled to it: `origin="lower"` reflected the
        *resulting index* (`rows-1-fp(u)`) where matplotlib reflects in the frame
        it resamples (`fp(rows-u)`); under `floor` those agree, under the
        rounding rule they do not, and they disagree at exactly the coordinates
        the rule changes. The destination row is reflected first now.

        **The affine was missing its clip translation.** Matplotlib resamples
        over `Bbox.intersection(out_bbox, clip_bbox)`, deriving both the output
        buffer size and the affine from the *clipped* rect
        (`Image._make_image`). The port ceiled the unclipped rect, mapped from
        the image's own corner, and clipped the result afterwards — which samples
        a different set of source cells, because the buffer-rounding
        compensation is computed from a different width and the pixel grid
        starts at a different device coordinate. `rasterizeForDestination` now
        intersects with `ctx.Clip` first and `matplotlibResampleAffine` carries
        the per-axis offset. Checked against matplotlib's recorded affine for
        `imshow_clipped`: sx 128.25, ox 2.0, sy 43.3333, oy 1.0, all reproduced
        exactly. This is the only current fixture where the term is non-zero, and
        it takes the case to zero differing pixels.

        `-update-golden` moved 14 goldens across the two passes. Every one is at
        least as good as before 3.3.6 on both differing pixels and RMSE — there
        is no case that had to be traded away, and no tolerance ceiling was
        raised. Ratcheted: `imshow_clipped` (3.60 to 0.1), `lognorm_imshow` (4.4
        to 0.65), `colorbar_symlog_ticks` (1.6 to 0.65), `twoslope_norm_image`
        (1.1 to 0.75), `asinh_norm_image` (2.10 to 1.85),
        `image_variants_gallery` (2.7 to 1.55), plus the three named cases.

        `specgram_psd`'s surviving 353-px row at device y=281 (destination row
        239 of 278) is **not** any of the three defects above. Its raster was
        compared directly against matplotlib's own resampled buffer: under the
        old index reflection the two agree on every one of the 278 rows, under
        the corrected one they differ at exactly this row, and yet the corrected
        form is what every other origin-lower case needs. matplotlib's
        `specgram` stores its array flipped (`flipud` + `origin="upper"`) where
        the port kept the natural order and set `ImageOriginLower`, so the two
        reflected in opposite frames. Fixing it in `core.Axes.Specgram` rather
        than in the resampler was the right call, but calling it an
        "orientation asymmetry" was not: the two orientations are the same
        picture, and the row moved for the float reason 3.4.4 gives. Closed
        under 3.4.4.

        Two things worth recording for whoever reads this next. The scale
        composition in `matplotlibResampleAffine` — exact device extent over
        source count, then the stretch compensating for rounding the buffer up —
        is bit-for-bit the naive ratio on every current fixture and by itself
        changed no pixel; it is kept because it is the faithful composition and
        because the offset is derived alongside it. And the filtered
        (non-nearest) branch now shares the reflected destination row and the
        affine, which is a no-op for it today but stops the two paths drifting.

  - [x] **3.3.7 `transform_coordinates`.** Done 2026-07-28. 261 differing
        pixels to **24** (largest cluster 226 to 2, RMSE 2.34 to 0.14).
        `transform_annotation_modes` fell out of 3.4 alongside it (402 to 19,
        RMSE 2.68 to 0.06) and `annotation_composition` went RMSE 0.93 to 0.21.

        The `+2-1` family label is a whole-image classification and was
        misleading: the text is byte-identical and the residual was three
        unrelated things. Localizing it first is what made this cheap.

        **The arrow's `patchA` box was the glyph ink rect, not the layout rect.**
        Matplotlib passes `Text.get_window_extent` to the annotation's
        `FancyArrowPatch` (`Annotation.update_positions`): the arrow starts at
        that box's centre, is clipped by it grown by `points_to_pixels(4)/2`,
        then shrunk by 2 points at each end. `_get_layout` builds the box from
        the advance width and `max(h, lp_h)` over `max(d, lp_d)`, measured from
        the literal string `"lp"` — 11 above the baseline and 3 below for
        `"axes note"` at 10 pt, against a ~9-px ink height. `core/annotation.go`
        reproduced the whole clip/shrink structure faithfully but fed it
        `textInkRect`, so the shaft was tilted and its tail walked several
        pixels along the text. The layout rect is now built inline, from
        `layout.Width/Ascent/Descent` — which `render.MeasureTextLineLayout`
        already derives with matplotlib's own `"lp"` clamp. The multiline
        annotation path (`measureMultilineTextBlock`) had it right all along;
        only the single-line path diverged from its sibling. Verified against
        matplotlib's recorded geometry for this fixture: the shaft's tail and
        control point now match to 1e-3 px.

        **The arrow's default mutation scale was a constant.** `Axes.Annotate`
        defaulted `ArrowHeadSize` to 8; matplotlib's `mutation_scale` defaults
        to the annotation text's *own font size*
        (`ms = arrowprops.get("mutation_scale", self.get_size())`). Now
        `annotationDefaultHeadSize`, resolving through the axes rc when the
        options leave the size unset.

        Two fixtures had also dropped calls their reference scripts make, which
        no core change can compensate for. `test/parity/transform_coordinates`
        omitted `set_axisbelow(True)`, so the grid sat at the default z=1.5 —
        above `Scatter2D`'s patch z but below `Line2D`'s — and painted over the
        diamond marker exactly as the pixels showed; the port's default is
        matplotlib's and was never wrong. And `transform_coordinates`,
        `transform_annotation_modes` and `annotation_composition` all omitted
        `arrowstyle="->"`, taking the port's filled `-|>` default instead of the
        reference's open caret; that, not any geometry, was the last 10x10 blob
        at each arrow tip.

        `-update-golden` moved five goldens and every one improves on both
        differing pixels and RMSE. `annotation_composition` regressed on the
        first pass (RMSE 0.93 to 1.17) — the head-size fix made its *wrong*
        head bigger — and the arrowstyle fix is what turned it into a win; that
        intermediate state is why the "every moved golden must improve" check is
        run before ratcheting, not after. All five ratcheted; no ceiling raised.
        `mathtext_basic` carries PDF and SVG goldens, regenerated with
        `-update-pdf-golden -update-svg-golden`.

        **The arrow head consumed its line width in points.**
        `FancyArrowPatch.displayParts` converted the mutation scale to pixels
        but passed `EdgeWidth` straight through, where
        `_get_path_in_displaycoord` scales *both* by
        `dpi_cor = points_to_pixels(1)`. The line width only feeds
        `pad_projected = 0.5*lw/sin_t`, the amount the shaft is pulled back so a
        projected cap does not overshoot the head — so the error was a fixed
        fraction of a pixel at every arrow tip, invisible per-arrow and worth
        tens of pixels across a gallery. Found by pinning the shaft against
        matplotlib's recorded path: with this fixed, all three vertices of
        `transform_coordinates`' shaft match to **0.01 px**. Seven goldens
        improved, none regressed, including two more 3.4 cases —
        `patch_style_matrix` (RMSE 2.68 to 2.64) and `text_annotation_matrix`
        (2.85 to 2.76). `transform_coordinates` finished at **24 differing
        pixels, largest cluster 2, RMSE 0.14**; `transform_annotation_modes` at
        **19 / 3 / 0.06**.

        Left standing: `textInkRect`'s own fallback branch
        (`core/axis_ticklabels.go`) uses the y-up convention where its ink
        branch and every caller use y-down. It is unreachable under the AGG
        backend, so it was recorded rather than fixed blind here.

  - [x] **3.3.8 The remaining `mathtext_gallery` cluster.** Done 2026-07-28.
        Now **pixel-identical** to matplotlib (64 differing pixels to 0, RMSE
        0.964 to 0.008, max amplitude 255 to 1, family `shift +0-1` to
        `identical`). The audit's pixel-identical count is up to 29 and its
        shift families down to 3 — and none of those three is mathtext.

        The cluster was the upper limit `n` of `$\sum_{i=1}^{n} i^2$`, drawn one
        pixel high. **The fourth cause is that the over/under limit stack was
        computed in closed form.** `layoutMathLimits` had
        `y := -(base.Ascent + gap + super.Descent)` and derived the box ascent
        as `-y + super.Ascent`. Matplotlib builds
        `Vlist([HCentered(super), Vbox(0,vgap), HCentered(nucleus),
        Vbox(0,vgap), HCentered(sub)])`, takes its height from `Vlist.vpack`'s
        literal left-to-right walk (`x += d + p.height`, carrying the previous
        child's depth in `d`), sets
        `shift_amount = sub.height + vgap + nucleus.depth`, and derives every
        row baseline in `ship.vlist_out` as `cur_v = shift - vlist.height`
        followed by one `cur_v += p.height` per row. Algebraically identical,
        one ULP apart — the same species as 3.3.1/3.3.4/3.3.5 and the same
        defect `layoutMathFrac` had already fixed for fractions in 3.3.3;
        `layoutMathLimits` never got the same treatment.

        What makes this one worth recording is that **the tie is structural,
        not incidental**. `Output.to_raster` blits each glyph at
        `int(oy - iceberg)`, which truncates, and the top limit is by
        construction the topmost ink of the expression — so the raster bounding
        box is derived from that very glyph, and its `int()` argument is
        *exactly* 1.0 for any limit glyph with no ink above its iceberg. Every
        `\sum`, `\prod` and `\lim` sits on that tie and the closed form was a
        coin flip. At 17 pt matplotlib's walk reaches -29.427083333333336 for
        the `n` where the closed form reached -29.427083333333332, so `int()`
        gave 0 instead of 1; at 26 pt both land on 1.0, which is why
        `mathtext_integrals` never showed it and stays at 0 differing pixels
        through the fix. The nucleus baseline is no longer hardcoded `0` either
        — the accumulation does not cancel exactly (-3.55e-15 here) and
        matplotlib ships what it walks.

        The fix is in `github.com/cwbudde/mathtext`, shipped as **v0.5.2**;
        `go.mod` requires that version, with no `replace`. It carries a test
        that pins the exact float64 ship coordinates at 17 and 26 pt against a
        DejaVu Sans metrics fixture recorded from matplotlib 3.10.9 —
        comparisons are `==`, because a tolerance cannot see this defect.
        `layoutMathScript`'s regular sub/superscript branch is deliberately
        untouched: the `i^2` of this same expression already matches (its
        `int()` argument is 17.627, nowhere near a tie) and matplotlib's
        non-overunder branch has a genuinely different structure.

        `-update-golden` across all 190 cases moved exactly one golden,
        `mathtext_gallery`, and it improves on both differing pixels and RMSE.
        Tolerances ratcheted 2.6/80/80 to 0.1/4/4 — the same bound 3.3.4 gave
        `matshow_basic`, whose residual is the same uniform ±1 LSB noise;
        verified the ratchet bites by restoring the pre-fix golden (fails at
        MaxDiff=255, RMSE 0.96). The pixel gates are what bind here (measured
        0/0 against 4/4); at a measured RMSE of 0.008 the 0.1 ceiling reads as
        12.5x slack in the audit and is left for 3.6 to set with the rest.
        `mathtext_integrals` carries a Go-authored PDF golden, regenerated for
        the last-decimal coordinate change with `-update-pdf-golden`; its
        rendered output is unchanged.

- [ ] **3.4 Investigate the localized non-shift divergences.** Cases whose
      residual is confined but not a clean offset. Each has been localized to a
      named artist and given a subtask below; every one is now a hypothesis to
      confirm or refute, not a case to go looking for.
      (`transform_annotation_modes` left this list under 3.3.7 — 402 differing
      pixels to 19, RMSE 2.68 to 0.06. `mathtext_basic` left it under 3.4.1 —
      RMSE 2.37 to 0.14, max per-channel difference 254 to 16, i.e. nothing but
      resample antialiasing left.)

      **Two thresholds are quoted below and they are not interchangeable.** The
      audit table counts every pixel differing by more than 1 LSB, which is the
      right gate but buries structure under antialiasing haze. The localization
      here re-runs the same 8-connected clustering at an amplitude of 32, which
      keeps only pixels a viewer could see; `line2d_markers` is 3,385 pixels at
      the first threshold and 505 at the second. Subtasks quote **`>1 / >32`**.
      Where a per-cluster integer offset is given it is the `(dx, dy)` in
      [-2, 2] that minimizes the residual **of that cluster alone** — a local
      probe, unrelated to the audit's whole-frame `family` column.

      Sequencing: 3.4.2 went first, on the theory that marker path snapping was
      the only hypothesis here that plausibly explained residuals in several
      other cases (bar edges, patch borders, legend handles). **It killed that
      hypothesis** — marker snapping was already correct and the residual was
      two legend defects instead — so 3.4.5, 3.4.8 and 3.4.12 no longer have a
      shared cause waiting for them and must be measured on their own. 3.4.7
      went next as the cheapest, on the theory that it was an API gap with a
      known answer; **that premise was stale too** — the field it wanted already
      existed, and the case was a fixture omission plus a double-converted
      mutation scale. 3.4.9 is the deepest and should go last.

      **Check the Go fixture against its Python counterpart before blaming the
      core.** 3.4.3's dominant residual was neither a port bug nor a hypothesis
      worth testing: the Go fixture simply never passed a kwarg the Python
      fixture did, and the port's default was matplotlib's own correct default.
      Nothing in the harness compares the two sources, so a divergence like that
      is invisible until the pixels are localized to a named artist and the two
      files are read side by side. Do that first for every remaining subtask.

  - [x] **3.4.1 `mathtext_basic`: the rotated y-axis label.** Its whole residual
        was the `amplitude $\frac{1}{\sqrt{2}}$` y-label, drawn one device pixel
        off. Two independent defects, and each had been hiding the other.

        **Glue rounding happened before the shrink.** matplotlib rounds a
        stretched glue in `ship` (`cur_g = round(clamp(glue_set * cur_glue))`,
        `_mathtext.vlist_out`), which runs after every enclosing `Node.shrink()`
        has scaled `glue_set`. `layoutMathSqrt` rounded the fill glue between the
        vinculum and the radicand while building the box, so the `\frac`
        denominator shrink then scaled an already-rounded value: matplotlib
        shrinks 3.9653 to 2.7757 and rounds to 3, the port rounded to 4 and
        shrank to 2.8. The 0.2 px gap grew the `to_raster` bounding box by a
        whole row (image 91x28 where matplotlib has 91x29, parse descent
        11.1056 where matplotlib has 11.3056). Fixed in `mathtext` v0.5.3:
        `mathLayoutBox` carries the unrounded glue plus the span of runs/rules it
        positions, and `shrinkMathBox` re-rounds and patches the difference. Only
        vertical `Vlist` glue is tracked — the horizontal `HCentered` 'ss' glue
        in fractions and matrices is already rounded against shrunk boxes.

        **The rotated draw path inverted the anchor with the wrong metrics.**
        `rotatedTextBackendAnchorFromP` builds the anchor from
        `measureSingleLineTextLayout`, which on a raster backend reports
        matplotlib's `to_raster` ink-image metrics; `drawMathTextLayoutRotated`
        undid it with the advance box instead, losing half the width difference
        along the label's baseline and a whole descent difference across it. A
        `x += 2*sign(sin); y++` fudge in `DrawMathTextImageRotated`, attributed
        to an agg_go quarter-turn quirk, was compensating it. Neither is a quirk:
        `translate(0,-h) · rotate(-angle) · translate(x,y)` from
        `_backend_agg.h::draw_text_image` is reproduced exactly, and with the
        right metrics the port reaches matplotlib's own `round(x + descent*sin)`
        = 15 and `round(y + descent*cos) + 1` = 223 for this label. Both fudges
        removed.

        The two errors cancelled in `x - imageHeight`, which is why the label
        landed on matplotlib's columns before either was fixed and one pixel off
        after only the first. Result: 198 differing pixels to 65, none above
        amplitude 16, RMSE 2.37 to 0.14; tolerances 2.55/220/70 to 0.30/100/40.
        `-update-golden` over all 190 cases moved exactly one golden.
        `mathtext_basic` also carries Go-authored PDF and SVG goldens, which
        encode the moved radicand; regenerated with `-update-pdf-golden
        -update-svg-golden`.

        Not a bug: the label is clipped by the left figure edge. matplotlib
        clips it identically — the rotated image is 29 px thick and spans device
        x [-14, 15], so 14 columns fall outside the canvas, and both images carry
        the same 44 rows of ink in column 0. The fixture's
        `add_axes([0.10, 0.14, ...])` simply does not leave room for a
        two-level fraction in the y-label at 640x360.

  - [x] **3.4.2 `line2d_markers`: two legend defects, not marker snapping.**
        Done 2026-07-28. The entry this replaces localized the residual to the
        marker glyphs and "the left spine" and asked whether the port routes
        marker paths through `snapPath` at all. Both were wrong, and the
        localization was off by an artist in each case.

        **Marker snapping was never missing.** `Line2D` markers reach
        `agg.DrawMarkers`, which sets `SnapAuto` whenever `MarkerItem.SnapSet`
        is false, and the axis-aligned square passes `shouldSnapPath`. The
        in-plot square marker already agreed with matplotlib to the last LSB
        (`75` where matplotlib writes `76`). The left spine was likewise already
        byte-exact — it is at device x **58**, where both images read
        `[241 0 241]`; x=64 is the legend frame's left edge. So the "marker path
        snapping" hypothesis that 3.4.5, 3.4.8 and 3.4.12 were all deferring to
        does not exist. Every >32-amplitude pixel was inside the legend.

        **The legend's marker edge width was passed in points.**
        `drawMarkerSample` computes `markerEdgeWidthPx` and the half-fill branch
        used it, but the main branch passed the raw points value. One wrong unit
        produced both visible symptoms, because `snapPath` keys its half-pixel
        offset off `round(linewidth)%2` exactly as `PathSnapper` does: at 1.4 pt
        / DPI 100 the port stroked 1.4 px and rounded to 1 (odd → `snapValue`
        0.5), where matplotlib strokes 1.944 px and rounds to 2 (even → 0.0).
        Hence a 3-px soft band centred on x=78.5 against matplotlib's crisp 2-px
        band centred on 78.0. Coverage sums pin it: 0.201+1.0+0.201 = 1.402
        against 0.973+0.973 = 1.946. The half-fill entries, already correct,
        were the ones with 3-8 px clusters instead of 38-51.

        **The legend frame was not snapped.** `legend.py` builds the
        `legendPatch` with a literal `snap=True`, i.e. `SNAP_TRUE`, which snaps
        the rounded box *despite* its curves — `SNAP_AUTO` would reject them,
        and `SnapAuto` is what the port passed. The port's layout was already
        right: instrumenting matplotlib gives the identical pre-snap patch,
        device x `[64.5444, 223.4194]`, y `[96.2444, 309.8556]`. Applying
        `floor(v+0.5)+0.5` to `devPath`'s `(x, height - y)` reproduces
        matplotlib's measured edges exactly — 65.5 / 223.5 / 50.5 / 264.5 — and
        the port was rendering the raw unsnapped 64.57 / 223.4 / 50.09 / 263.82.

        Result: 3,385/505 px to 673/52, RMSE 3.227 to 1.594; the legend square
        handle now lands on matplotlib's columns and the frame's left edge is
        byte-identical. Tolerances 3.3/4300/2600 to 1.85/800/120. What is left
        is 2-px antialiasing specks at the ends of the legend handle lines.

        Both fixes are global to legends, so `-update-golden` moved 12 goldens
        plus `mixed_raster_vector`'s Go-authored PDF and SVG goldens (the SVG
        diff is the marker edge width, `0.45` to `0.625` = 0.45·100/72, which is
        the bug stated in one number). Every case with a matplotlib reference
        improved on differing-pixel count; `mathtext_inline_labels` (1,013 px to
        4), `animation_gallery` (1,142 to 73) and `annotation_composition` (506
        to 8) were almost entirely legend frame, and their ceilings were
        ratcheted with the rest. `large_scatter` is the one case whose RMSE rose
        (1.262 to 1.314) while its pixel count fell 2,573 to 1,773 and its
        largest cluster 806 to 144: its legend handle now has matplotlib's
        two-transition-pixel edge instead of one, which sharpens a pre-existing
        1-px handle placement offset of the 3.4.10 family rather than adding a
        new defect.

        Not fixed, and worth its own subtask: `markerSnapMode` /
        `markerSnapThreshold` in `core/scatter.go` is a marker-size threshold
        table with **no matplotlib counterpart** — matplotlib decides purely on
        geometry under `SNAP_AUTO`. It is also dead on AGG, because
        `drawLegendMarkerPath` never sets `MarkerItem.SnapSet` and `DrawMarkers`
        therefore overrides it with `SnapAuto`; it still steers the non-AGG
        fallback. Also left alone: the legend *shadow* paint still passes
        `SnapAuto`, since matplotlib draws the shadow through a separate
        `Shadow` artist and there is no evidence it snaps.

  - [x] **3.4.3 `line2d_semantics`: butt caps and a second dasher, not dash
        phase.** Done 2026-07-29. The entry this replaces called the residual a
        phase drift and asked where the dash offset is reset relative to the
        segment loop. **There is no such loop.** `agg_go`'s `vcgen_dash` is
        already phase-continuous: it accumulates `currDashStart` across every
        vertex of a polyline, and `ConvAdaptorVCGen` restarts the generator per
        **subpath**, exactly as AGG does. The old entry also mislocalized the
        artist — it described "a red dashed line", and this fixture has none;
        the red artist is solid, broken by `NaN` and `inf`. Two unrelated causes.

        **10 of the 16 clusters were line endpoints, and the Go fixture was
        wrong.** Each sat on a polyline end, ink in the port where matplotlib
        had background, with a width that scaled with linewidth — a projecting
        cap. The Python fixture passes `solid_capstyle="butt"` to all four
        `ax.plot` calls; the Go fixture passed no cap at all and took the rc
        default `SolidCap: CapSquare`, which is matplotlib's own correct
        `lines.solid_capstyle: projecting` default. So the port was right and
        the fixture silently disagreed with its Python counterpart. The red
        line's three subpaths give exactly six ends and all six were present,
        which is what identified it. `PlotOptions.LineCap` already expressed
        this, so there was no API gap; the dashed `gapcolor` line needed no
        change, since a dashed stroke already selects `CapButt`.

        **The other 6 were gapcolor gaps, dashed by a second, disagreeing
        dasher.** All six were single columns at dash boundaries of the
        `gapcolor` line — and the other six boundaries matched. matplotlib
        (`lines.py`) paints the gaps by stroking the *same path* a second time
        through the *same* `conv_dash`, with `_get_inverse_dash_pattern`: rotate
        the sequence (`dashes[-1:] + dashes[:-1]`) and advance the dash offset
        by `dashes[-1]` so the new first entry is skipped. The two passes then
        interlock by construction. The port instead built an explicit gap
        polyline in core (`dashGapPath`) and stroked it undashed, so its float
        boundaries drifted from `vcgen_dash`'s and left white slivers — at
        x=420 the port read `[134 183 147]`, green over white with no orange,
        where matplotlib read `[129 156 101]`.

        The port could not express matplotlib's approach because **dash offset
        did not exist anywhere in the codebase**: no phase field on
        `render.Paint`, PDF/PS/PGF hardcoding the phase operand to `0`, SVG
        never emitting `stroke-dashoffset`, and the AGG backend never calling
        `DashStart` even though `agg_go` has supported it all along. Added
        `render.Paint.DashOffset` (one new field on the frozen API) and plumbed
        it through agg, svg, pdf, ps, pgf and gobasic — the last one also being
        the Skia CPU fallback. `dashGapPath` is deleted; `inverseDashPattern`
        replaces it in ~12 lines. `backends/agg/` gained its first dash
        rendering test.

        Result: 300/138 px to 71/0, RMSE 2.775 to 0.171, max per-channel
        difference 13 — **nothing survives above antialiasing amplitude**, and
        the >32 cluster count went 16 to 0. Tolerances 2.6/380/23 to
        0.20/85/8. `-update-golden` moved exactly one golden: `GapColor` has a
        single fixture and every other paint carries offset 0.

        Three further dash defects surfaced while measuring this case. None of
        them moved a golden — no catalog case reaches any of them — but all
        three are now fixed rather than carried:

        - **Dash phase reset at AGG chunk boundaries.** `chunkStrokePath` issues
          one `Stroke()` per 32768-vertex chunk and `ConvAdaptorVCGen` restarts
          the generator per stroke, so a dashed polyline past that size snapped
          back to the start of the pattern at every boundary. This is the bug
          the old 3.4.3 entry was describing — real, just not what this case
          hit. Chunking is an implementation detail of the port, not of
          matplotlib, so it must be invisible: `chunkStrokePath` now returns a
          `pathChunk` carrying the arc length since the last *genuine* MoveTo,
          and the draw loop re-enters the pattern there. A chunk that opens on a
          MoveTo present in the source path still reports phase 0, because AGG
          is supposed to restart at a real subpath.
        - **`linestyle=(offset, (on, off))` was unsupported.** `DashPattern`
          gained `Offset`, in the same units as `Lengths` and scaled with them
          (`lines._scale_dashes` scales the offset by the line width too), plus
          `WithOffset`/`ScaledOffset` and the `Line2D.SetLineStyleDashes`
          setter; `PlotOptions.DashOffset` is the `ax.Plot` spelling. The phase
          is folded into one cycle at the point it reaches the paint, matching
          `_get_dash_pattern`'s `offset %= dsum`.
        - **Odd-length dash patterns were mishandled two different ways.** AGG's
          `SetDashPattern` dropped the unpaired trailing entry (`[10, 5, 20]`
          collapsed to a 15px `[10, 5]` cycle) and gobasic — which is also the
          Skia CPU fallback — rejected odd lengths outright and drew the line
          solid. An odd sequence alternates on/off forever, so its second pass
          inverts its first; matplotlib's `Dashes` caster walks it twice "in
          accordance with the pdf/ps/svg specs", and those specs cycle odd
          arrays that way of their own accord. `render.NormalizeDashes` does the
          doubling for the two backends that build explicit `(on, off)` pairs.
          SVG, PDF, PS and PGF emit the array as given and needed no change.

  - [x] **3.4.4 `specgram_psd`: one image row, and the 3.3.6 note is stale.**
        Done 2026-07-29. 354 differing pixels to **1**, RMSE 1.003 to **0.056**.
        The whole residual was device row 281 — one contiguous run x[81:434];
        the colour band boundary matplotlib puts at row 281 the port put at row
        282, with rows 44-280 and 282-319 byte-identical. What survives is a
        single pixel at the axes' bottom edge differing by 2.

        **The two orientations were never in disagreement about the picture.**
        matplotlib's `specgram` hands `imshow` a flipped array with
        `origin="upper"` (`Axes.specgram`: `Z = np.flipud(Z)`, then
        `origin='upper'`, extent left as `(xmin, xmax, freqs[0], freqs[-1])`);
        `Axes.Specgram` kept the natural row order and set `ImageOriginLower`.
        Those are the same image in exact arithmetic, which is why 275 of 276
        rows already matched and why the old 3.4 entry was right to reject
        "orientation error" as the diagnosis. They are not the same in float64:
        origin-lower reflects the *destination* row and uses `oy` as given,
        while origin-upper leaves the destination row alone and folds the flip
        into the offset (`oy = rows - bufHeight/sy - oy`). Different groupings
        of the same algebra, and `nearestSourceIndex` quantizes to 1/256 of a
        cell with `agg::iround` and floors — so a coordinate within 1/512 of a
        cell edge snaps the other way, moving one whole source row. The same
        `int()`-at-a-tie species as all of 3.3.

        Fixed by mirroring matplotlib in `core.Axes.Specgram`: reverse the row
        order of the array handed to `Image` and pass `ImageOriginUpper`, which
        puts the port on matplotlib's float path instead of an equivalent one.
        `SpecgramResult.Spectrum` stays in natural order, as matplotlib returns
        the unflipped `spec`. Nothing in the resampler changed, so no other case
        moved — the full suite is green and no audit row regressed.

        Tolerances ratcheted 0.035/1.05/400/400 to 0.01/0.1/8/4. The audit's
        `shift +0+1` classification was an artifact — any one-row residual is
        trivially cancelled by a one-row roll — and `specgram_psd` now reads
        `sparse-local`, verdict `ok`. Regenerating the audit also absorbed the
        drift from the three golden-updating commits since it was last written
        at `7c614276`, so its diff is much wider than this change.

  - [x] **3.4.5 `stat_variants`: two abutting bar edges, not one.** Done
        2026-07-29. 176 pixels above amplitude 32 to **0**, 933 differing pixels
        to 653, RMSE 2.37 to **0.28**, max per-channel difference 224 to 17.
        Snapping was not the defect — the entry this replaces read the residual
        as "the last bar's right edge" and asked about `Rectangle` path
        snapping; the localization was off by an artist and the snapper was
        already correct.

        **Matplotlib draws that boundary in two pixel columns and the port drew
        one.** Every >32 pixel was in columns 759 and 760, rows 415-557: the
        shared boundary at data `x = 5` between the two rightmost bars of the
        Stacked Multi-Hist axes. Above the shorter bar's top only one edge
        exists, and matplotlib puts it at column 759 where the port put 760;
        below it matplotlib has *both* columns dark and the port only 760.

        **A `Rectangle`'s bbox transform is folded into `transData` before any
        vertex is mapped.** Its vertices are `Path.unit_rectangle()` mapped by
        `BboxTransformTo(bbox) + transData`, and `Transform.get_affine()`
        collapses that pair into one matrix (a single `np.dot`). So with
        `a = 55.3`, `b = 482.99999999999994` the bar ending at `x = 5` reaches
        its right edge as `(a*w) + (a*x0 + b) = 759.4999999999999` while the bar
        starting there reaches its left edge as `a*x0 + b = 759.5`. Those
        straddle `PathSnapper::vertex`' exact `floor(v + 0.5)`, and at linewidth
        0.7 pt (0.972 px, `round → 1`, odd → `snapValue` 0.5) they land at 759.5
        and 760.5 — one crisp column each. The port computed both edges in data
        space (`x ± width/2`, exactly `5.0` either way) and applied
        `DataToPixel` to the corner points, so both became `a*5 + b` and
        collapsed onto one column. Same species as 3.3.2: identical in exact
        arithmetic, not in float64, because the legs were evaluated in sequence
        instead of composed. It is also why *only* this edge diverged — of the
        six bin boundaries only `x = 5` maps onto an exact `.5` device tie.

        Fixed by `rectPatchCorners` in `core/bar.go`, which builds the bbox
        affine `{A: w, D: h, E: x0, F: y0}`, composes it with the data affine
        when the data leg is affine, and maps the unit rect's corners;
        `Bar2D.createVerticalBarPath`/`createHorizontalBarPath` and `Hist2D`'s
        bar branch route through it, the latter spelling out matplotlib's
        `hist` → `bar(bins[:-1] + 0.5*totwidth, …, align='center')` round trip
        rather than shortcutting to `edges[i]`. Signed widths/heights are kept
        as matplotlib keeps them; the device rect is still normalized
        afterwards, so winding is unchanged.
        `TestAbuttingBarEdgesKeepMatplotlibsULPSplit` pins both device values.

        `-update-golden` over all 190 cases moved **exactly one** golden, as
        expected — only an edge landing on a `.5` tie can move.
        Tolerances ratcheted 0.35/3.4/1400/530 to 0.1/0.36/820/530.

        What remained was 653 pixels at amplitude ≤17, spread over the axes
        interiors and rows 256/451, attributed here to "±1-2 LSB rounding when
        the translucent fills composite" and handed to 3.5. That was the right
        mechanism but badly undersized: 3.5.1 traced it to a single blender
        formula mismatch worth 4.6M pixels suite-wide and fixed it, taking this
        case to 79 differing pixels and RMSE 0.067. Tolerances ratcheted again
        there, to 0.01/0.0838/99/68.

  - [x] **3.4.6 `patch_style_matrix`: the bracket bar was drawn at half width.**
        Done 2026-07-29. The entry this replaces localized the residual to the
        cap bars of the `|-|` arrow and guessed a points-vs-pixels units mistake
        of the kind 3.3.7 found. The localization was right and the diagnosis
        was wrong: the units were already correct and the error was a factor of
        two in the geometry.

        **`width` is a half-extent, not a full one.**
        `patches._Curve._get_bracket` passes the bracket width straight into
        `bezier.get_normal_points(x0, y0, cos_t, sin_t, width)`, whose last
        argument is the *distance* each of the two returned points sits off the
        line. So matplotlib's bar spans `2 * widthA * scaleA`, and `bracketPath`
        (`core/arrow_patch.go`) halved it instead. Confirmed against the vendored
        3.10.9 directly: `ArrowStyle("|-|").transmute` on a path through the
        origin at `mutation_size=15` returns bar endpoints at x = ±15, where the
        port produced ±7.5. The fixture's `mutation_scale=15` therefore lost
        7.5 px at each end of both bars — far more than the "about 4 px" the old
        entry quoted, which was an artifact of the local offset probe being
        clamped to [-2, 2].

        The bars are the only thing that moved. Result: 2,473/373 px to
        2,381/281 px, RMSE 2.640 to 1.717; every cluster above amplitude 32 is
        now 4 px or smaller (the largest was 13). Tolerances ratcheted
        2.90/3150/75 to 2.00/2450/70 — the RMSE and pixel gates both bite
        against the pre-fix golden (2.640 > 2.00, 2,473 > 2,450). What is left
        is 161 scattered specks at amplitude >32, none of them structural, in
        the hatch and wedge-arrow band at x[549:643] y[154:216]; the largest
        cluster at the audit's own >1 threshold is unchanged at 53 px.
        `TestArrowStyleBracketScaleOverridesMutationSize` encoded the halved
        bound and was corrected against matplotlib's output.

  - [x] **3.4.7 `annotation_legend_offsetbox_gallery`: a fixture omission and a
        double-converted mutation scale.** 2,026/289 px, RMSE 1.075 to
        1,287/244 px, RMSE 0.85; the arrow cluster at y[219:229] x[717:728] is
        now pixel-identical (max per-channel difference 1). Tolerances ratcheted
        to RMSE 0.95 / 1,450 px / 480 px.

        **The "no `ArrowStyle` field" premise was stale.** This entry claimed
        `core.AnnotationBboxOptions` could not express an arrow style and that
        closing it needed a public API change, a frozen-audit regen and an
        `api-freeze-delta.md` record. `ArrowStyle` and `ConnectionStyle` were
        already there; nothing on the exported surface moved.

        **The filled head was a fixture omission after all.** `plot.py` passes
        `arrowprops={"arrowstyle": "->"}`; the Go example set `Arrow: true` and
        never set `ArrowStyle`, so it fell through to the port's default `-|>`.
        Exactly the divergence 3.4.3 warns about — read the two sources side by
        side before blaming the core.

        **The core defect was the mutation scale, and it is double-converted.**
        `Axes.AnnotationBbox` hardcoded `ArrowHeadSize = 8`. matplotlib takes it
        from the annotation box's own font size, which is independent of the
        offset box content's font size and defaults to
        `rcParams["legend.fontsize"]` — but `offsetbox.AnnotationBbox.update_positions`
        stores `renderer.points_to_pixels(self.get_fontsize())`, i.e. a value
        already in **pixels**, into a slot `FancyArrowPatch` then multiplies by
        `dpi_cor = points_to_pixels(1)` a second time. At 100 dpi the reference
        arrow transmutes at mutation scale **19.29**, not 13.89. Fixing only the
        `8 → 10` default still left both caret arms ~2 px short, because the
        port converted once. `AnnotationBbox.arrowProxy` now applies the extra
        factor; `text.Annotation.update_positions` has no such second
        conversion, so it stays out of the shared arrow path. Verified by
        printing the transmuted head vertices on both sides — they agree to
        1e-4 once the scale matches.

        The two further clusters this entry used to list — 40-42 px vertical
        runs at x[715:716] and x[970:972] that `dx=-1` cancelled exactly — were
        legend frame edges, and 3.4.2 closed them. What remains on this case is
        text antialiasing: the largest cluster is now 434 px at y[609:629]
        x[133:156] (bottom-panel anchored text), all of it below amplitude 32
        except 18 px.

  - [ ] **3.4.8 `text_annotation_matrix`: text-bbox borders.** 1,625/490 px,
        RMSE 2.70 against 2.95 (1.09x) — re-measured after 3.4.7, which fixed
        the AnnotationBbox arrow mutation scale and took this case from
        2,082/487 px and RMSE 2.761. The residual concentrates on
        `FancyBboxPatch` borders behind annotation text — the largest cluster is
        a 24x3 horizontal strip at y[85:88] x[110:134], the bottom edge of an
        olive-bordered box — plus two arrow heads at x[447:463] that local
        offsets of `(-1,0)` and `(-2,+1)` only partly cancel. Split it: settle
        the box borders first (they are the bulk; the shared cause with 3.4.2
        that this used to assume is gone, but 3.4.2 did find that matplotlib
        forces `snap=True` on the legend's `FancyBboxPatch`, so check what snap
        state these annotation `FancyBboxPatch`es carry), then re-measure the
        arrows. **Read the two fixtures side by side first:**
        `test/parity/text_annotation_matrix/plot.py` passes
        `"linewidth": 1.25` on the `TextArea` box's `arrowprops` where
        `plot.go` leaves `ArrowWidth` at the 1.0 default — 3.4.7's fixture
        omission again, and it must be settled before opening a third arrow
        investigation. **Check the affine composition too:** 3.4.5 found
        that generic patches still evaluate their local-to-data affine and the
        data-to-pixel leg in sequence (`buildCachedDisplayPath`,
        `core/patch_paths.go`) where matplotlib collapses the pair into one
        matrix first. Bars were fixed there; every `Patch` still has it, and it
        moves a box border by a whole pixel whenever an edge lands on a `.5`
        device tie.

  - [ ] **3.4.9 `unstructured_showcase`: inline contour labels sit at different
        points along the contours.** 3,097/1,074 px, RMSE 2.634 against 3.20
        (1.21x, `ok`). The six largest clusters are 78-136 px blobs in
        x[584:693] y[144:378] and **no integer offset improves any of them**.
        Magnified, the reason is plain: the port and matplotlib label the same
        contour lines with the same glyphs at _different positions along the
        path_ — where the port writes "0.9" matplotlib writes "0.6", and the
        1.2 label sits a third of the way further along. This is
        `ContourLabeler.labels`' location search (`locate_label`, the
        path-length/spacing heuristic and the break inserted for the label gap),
        not text rendering. Expect this to be the largest piece of work in 3.4
        and the least likely to reduce to a rounding fix; if the search cannot
        be made faithful, record it as an accepted difference with the
        reasoning, which is a legitimate outcome for this phase.

  - [ ] **3.4.10 `legend_layout_matrix`: scatter handle geometry.** 585/198 px,
        RMSE 1.657 against 1.90 (1.15x) — re-measured after 3.4.2, which removed
        the frame and handle-edge residual and left this as the whole case: the
        four largest clusters, now 32-50 px each, are all inside y[94:113]
        x[80:118] — one legend entry's three-circle scatter handle, and they are
        4 of the 14 clusters left. `dx=+1` partly cancels each, and magnified
        the port's circles are slightly smaller and closer together than
        matplotlib's. Check the legend handle's marker sizing against
        `HandlerPathCollection` (`_default_update_prop` scales by
        `legend.markerscale`, and the sample x positions come from
        `handlelength`/`scatterpoints`, all in points).

  - [ ] **3.4.11 `colorbar_boundary_values`: the two extreme ticks.** 76/63 px,
        RMSE 1.555 against 3.10 (1.99x, `ok`) — the smallest residual in this
        group. All four clusters are at the colorbar's first and last tick:
        13-14 px vertical runs at x[501:502] y[43:56] and y[307:321], and 12-24
        px horizontal runs at x[503:515] on rows 56 and 307-309. Interior ticks
        are identical. So it is specifically the boundary ticks of a
        `BoundaryNorm` colorbar — check `Colorbar._ticker`'s handling of the
        extend regions and whether the first/last tick is drawn at the boundary
        or at the segment centre.

  - [ ] **3.4.12 `axes_control_surface`: accept and ratchet.** 893/29 px, RMSE
        0.594 against a 3.00 ceiling — already flagged `loose` at 5.05x slack,
        and the only case in this group whose gate is nowhere near binding. At
        amplitude 32 just 29 pixels survive, in 15 clusters of 2 px each,
        evenly spaced ~38.5 px apart along y[76:78] — the top ends of the major
        tick marks, one pixel of antialiasing apiece. 3.4.2 did not change it —
        its fixes were legend-local and this case's residual is on tick marks —
        so the disposition here is "accepted difference", and the real action is
        3.6 tightening the ceiling to its measured value.

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
  - [x] **3.5.1 The whole-suite off-by-one: the AGG plain blender.** Done
        2026-07-29. Not a case, a mechanism, and the largest single parity item
        left in Phase 3. Across all 190 reference cases **4,613,539 pixels
        differed by exactly 1** against 618,008 differing by more — deterministic,
        not noise.

        Matplotlib does not use stock AGG for compositing. `_backend_agg.h:113`
        builds its pixel format from `fixed_blender_rgba_plain`
        (`src/agg_workaround.h:52`), which exists "to workaround a bug in Agg SVN
        where the blending of RGBA32 pixels does not preserve enough precision".
        It keeps the composite in one fixed-point expression —
        `a = ((alpha + a) << 8) - alpha * a`, then
        `p[R] = (((cr << 8) - r) * alpha + (r << 8)) / a` — where stock
        `agg24-svn` (`extern/agg24-svn/include/agg_pixfmt_rgba.h:243`) routes it
        through multiply → `lerp` → demultiply and picks up `lerp`'s
        round-to-nearest term. `agg_go`'s `BlenderRGBA8Plain` is a faithful port
        of the SVN one, so the port rounded where matplotlib truncates, and the
        two also disagree on the real value because matplotlib divides by the
        *unshifted* combined alpha. Confirmed bit-exactly on the three dominant
        `stat_variants` fills before any code changed: step-filled blue
        `(107,158,230) α140` → port `(174,202,241)` vs reference `(173,201,241)`,
        stacked orange `(219,107,48) α204` → `(226,137,89)` vs `(226,136,89)`,
        stackplot blue `(51,140,191) α194` → `(100,168,206)` vs `(100,167,206)`.

        Fixed in `agg_go` by adding `BlenderRGBA8PlainFixed` beside the SVN one
        rather than replacing it — the SVN port is correct for what it ports, and
        other users depend on that fidelity. `Agg2D` selects it through
        `SetHighPrecisionPlainBlend`, exposed as
        `agg.NewContextForImageWithOptions`; `newAggSurface`
        (`backends/agg/surface.go`) opts in. The blender deliberately does not
        implement `PremulSrc`, so the pixfmt's standard fast paths cannot
        substitute lerp math for it, and it carries its own opaque + full-cover
        short-circuit because C++ AGG's `copy_or_blend_pix` never reaches
        `blend_pix` in that case and the formula is off by one without it.
        `TestTranslucentFillMatchesMatplotlibBlender` and
        `TestOpaqueFillIsCopiedNotBlended` pin the values, all read out of
        Matplotlib 3.10.9 rendering the same fills.

        Suite-wide: amplitude-1 pixels 4,613,539 → 2,933,607, amplitude>1
        618,008 → 507,176, byte-identical cases **4 → 25**, 166 cases improved
        against 19 that regressed (total RMSE gain 5.89 against 0.41 lost), and
        no case's peak amplitude grew by more than 2. The 19 regressions are all
        ±1 wobble on pre-existing defects, not new ones: in `polar_axes` and the
        `geo_*` cases the affected pixels are antialiased grid lines where the
        port is already 8-10 units off, so a corrected 1-LSB step can land either
        way. 48 cases ratcheted onto the new measurements at 1.25x headroom,
        which took the audit's `loose` verdicts from 44 to 26.

        **Two findings for 3.5's remaining list.** First, none of the dense
        residuals listed above were resolved by this — measured before and after,
        they are separate defects and stay on the list. Second, the 2.9M
        amplitude-1 pixels that survive are *not* leftovers of this bug: they sit
        in the paths deliberately left out of scope (the image blit, which
        `internal/agg2d/image.go` routes through the premultiplied pixfmt where
        matplotlib uses the plain fixed blender; `compositeClipSurface` in
        `backends/agg/agg_clip.go`; `blendAlphaMask` in
        `backends/agg/freetype_alpha_native.go`) — and at least one is a
        different bug entirely. `quad_mesh` did not move a single pixel and is
        the clean example: its 331,560 differing pixels are **opaque** viridis
        fills where the port reads one *below* the reference
        (`(52,97,141)` vs `(51,96,141)`), so no blender is involved. That is a
        colormap/float→uint8 quantization mismatch and wants its own subtask.

        The agg_go side shipped as **v0.4.0** and `go.mod` depends on the tag,
        so there is no `replace` directive — the goldens were re-verified
        byte-identical against the published module.

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

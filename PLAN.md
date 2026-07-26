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

**Done when:** native batch primitives and real GPU acceleration are available
under `skiacgo`/`skiagpu`, output is parity-checked, and per-mode capability
reporting reflects runtime behavior.

## Phase 2: Go-Idiomatic API Rework and `core/` Split

**Goal:** perform the single coordinated pre-v1.0 breaking pass: make the API
Go-idiomatic, split the `core/` package, and re-freeze the public surface.
Rendering must remain byte-identical to the current pre-break golden baseline.

After every stage, regenerate the frozen API with
`UPDATE_PUBLIC_API_AUDIT=1`, remap
`internal/examplecatalog/public_surface_parity.go`, regenerate
`docs/matplotlib-parity-status.md` with `go run ./cmd/paritystatusdoc`, and
update the coupled API/doc tests in the same commit.

### 2.1 Surface Tiering

- [x] Classify all 3,102 frozen symbols as keep, demote, or delete in a design
      document. Explicitly decide the fate of:
  - Python-style introspection (`Setp`, `Getp`, `GetpAll`, `Findobj`,
    `FindobjType`);
  - `*Units` variants that overlap the new error convention;
  - renderer-extension interfaces used only by backends.

### 2.2 Package Split

- [x] Move `core/axes3d*.go` and 3D projection files into `plot3d`
      (approximately 98 exported symbols and 7k lines).
- [x] Move tick locators, formatters, and date ticks into `ticker`, preserving
      Matplotlib's natural `ticker`/dates boundary where useful.
- [x] Move widget and selector implementations into `widgets`, beside the
      canvas/event layer.
- [ ] After each move, keep `go build ./...` and `just test` green, verify
      goldens are byte-identical, run the full API regeneration workflow,
      refresh `docs/large-file-decomposition.md`, and run
      `just large-file-audit`.

### 2.3 Idiomatic Conventions

- [x] Adopt one error convention: rejected plot input returns `(T, error)`;
      `diag.Warnf` remains only for accepted degradations. Fold redundant
      `*Units` variants into primary methods.
  - [x] Fold `PlotUnits` into a transactional, unit-capable `Plot`.
  - [x] Fold `ScatterUnits` into `Scatter`; reject conversion, shape, scalar-map,
        and extra-option errors without mutating axes or advancing the cycle.
  - [x] Fold `BarUnits` into `Bar`/`BarH`; preserve categorical-axis locator
        behavior and make vertical/horizontal rejection transactional.
  - [x] Fold `FillBetweenUnits` into `FillBetween`; validate all three converted
        inputs before constructing polygons or committing axis-unit state.
  - [x] Inventory the remaining warn-and-skip plotting entry points, convert
        rejected input to errors, and retain warnings only where an artist is
        accepted with a documented degradation.
- [x] Replace the 83 variadic option structs and 408 pointer-to-primitive
      fields with one consistent options model; extra option sets must be
      impossible or rejected. Replace raw-string enums with typed constants.
  - [x] Reject extra option values across the existing variadic surface as an
        intermediate safety step.
  - [x] Choose and document the final options representation, then migrate one
        representative line, collection, image, and annotation API.
  - [x] Migrate the remaining option families and replace option raw strings
        with typed constants.
- [x] Add `Figure.Save(path)`, `Figure.WriteTo(w, format)`, and
      `Figure.Image()`; replace the repeated backend-specific save boilerplate
      in examples.
- [x] Rename `GetX()` getters to `X()` (or an explicit `LookupX()` spelling
      where the noun conflicts with an exported type).
- [ ] Resolve exported mutable fields versus setter duplication consistently.
  - [x] Classify exported mutable fields by immutable configuration, observable
        state, and internal cache; record the intended ownership per family.
  - [x] Encapsulate fields that duplicate setters and update callers/tests.
  - [x] Re-run the API/doc freeze and add migration notes for every removal.
  - [ ] Close the two deferred families: move `PathCollection`'s
        `EdgeColorsFace` mirroring to read time, and fold
        `Line2D.MarkerFaceColor`/`MarkerEdgeColor` into their `*Spec` siblings
        and `Dashes`/`DashUnits` into one value type (see
        `docs/plans/phase2-mutable-fields.md`).
- [x] Document the concurrency contract for global rc state, registries, and
      figures; stop discarding pyplot errors.
- [ ] Consolidate duplicated alpha baking, option unpacking, and scalar-map
      resolution paths.
  - [x] Share alpha multiplication through `render.Color.WithAlphaMultiplier`.
  - [x] Share plot/plot3d scalar-map configuration through
        `PlotOptions.ScalarMapConfig`.
  - [x] Centralize single-option unpacking/extra-option rejection through
        `internal/optarg`.
  - [ ] Route remaining scalar-mappable artists through the shared resolver.

### 2.4 Re-freeze

- [x] Create the post-split checkpoint: regenerate `stable_public_api.json`,
      remap public-surface classifications, regenerate the parity-status
      document, add migration notes for every completed break, and draft the
      Phase 4 changelog section.
- [ ] Repeat the freeze after the remaining error/options/mutable-field work
      and treat that artifact as the final Phase 2 surface.

**Done when:** `core/` no longer owns plot3d/ticker/widgets; plot methods use
the chosen error and options conventions; no raw-string option enums remain;
the new figure save surface is used by examples; the API is re-frozen; and all
goldens remain byte-identical to the pre-break baseline.

**2026-07-25 checkpoint:** surface tiering, the `plot3d`/`ticker`/`dates`/
`widgets` moves, figure output, getter naming, concurrency documentation,
registry synchronization, example migration, API/parity remapping, migration
notes, and the changelog draft are complete. No golden/reference fixture
changed. The error convention is closed: all four `*Units` folds are done and
the warn-and-skip inventory is complete. Remaining Phase 2 work is the
options/raw-enum conversion, mutable-field cleanup, and the remaining
option/scalar-map consolidation paths. `Axes.PlotUnits` is now folded into the
transactional, unit-capable `Axes.Plot` method, which returns `(*Line2D, error)`
and rejects extra option values; `PlotDate` and the corresponding pyplot
wrappers propagate that error. `Axes.ScatterUnits` is likewise folded into the
transactional, unit-capable `Axes.Scatter`; scalar-map and per-point shape
validation happen before the scatter cycle advances, and `pyplot.Scatter`
propagates errors. `Axes.BarUnits` is now folded into transactional,
unit-capable `Axes.Bar`/`BarH`; shape, orientation, per-bar, and error-bar
validation happen before the property cycle advances, categorical position
locators remain intact, and the pyplot wrappers propagate errors.
`Axes.FillBetweenUnits` is folded into the transactional, unit-capable
`Axes.FillBetween`, which converts and validates x, y1, y2, and the `Where`
mask before committing axis units or adding a polygon. The warn-and-skip
inventory then converted the last five rejecting entry points — `FillBetweenX`,
`FillBetweenPlot`, `Hist`, `ErrorBar`, and `ImShowRGB`, plus
`ErrorBarContainer` and the matching pyplot wrappers — to `(T, error)`, leaving
`diag.Warnf` only for artists accepted with a documented degradation; the audit
and the retained-warning rationale are in
`docs/plans/phase2-warn-and-skip-inventory.md`. The options rework then began
with its intermediate safety step: `internal/optarg` is now the single place the
"at most one option value" rule lives, and every variadic option tail routes
through it. Entry points that already return an error report a
`*optarg.TooManyError`; the rest panic, because an extra option value can only
come from a literal at the call site. Internal helpers that received an
already-unpacked option set lost their variadic tails outright, so the frozen
public API is unchanged by that pass; the rationale, the enforcement table, and
the variadic parameters deliberately left alone are in
`docs/plans/phase2-extra-option-rejection.md`. The final options
representation is now chosen and documented in
`docs/plans/phase2-options-model.md`: entry points take exactly one options
value instead of a variadic tail, so extra option sets are a compile error, and
optional fields are the new `optional.Value[T]` instead of pointers or magic
zero values. One API per artist family is migrated end to end — `Axes.Stem`
(line), `Axes.HLines`/`Axes.VLines` (collection, via the new
`LineCollectionOptions` that stops using the `LineCollection` artist as its own
options), `Axes.ImShow` (image), and `Axes.Annotate` (annotation) — together
with their pyplot wrappers, with no golden fixture changes. The bulk migration
then closed the rest: all 205 variadic option tails and all 408
pointer-to-primitive option fields are gone across `core`, `pyplot`, `plot3d`,
and `widgets`, so an extra option set is a compile error everywhere and
`internal/optarg` is deleted. The raw-string option enums became defined string
types with named constants (`PlotOrientation`, `ColorbarExtend`,
`ColorbarLocation`, `ImageAspect`, `VectorPivot`, `ViolinSide` in
`core/option_enums.go`). Collapsing the two recurring merge shapes — the
prepared-defaults literal that was immediately overwritten, and the `supplied`
flag that guarded a no-op block — made the pass a net deletion; the four
pointer-cloning helpers went with it. Goldens and references are byte-identical
throughout. The mutable-field cleanup then gave every artist field exactly one
public writer. Auditing all 1750 exported non-option fields found 29 shadowed by
a `Set<Field>` that did more than assign, and the fix splits by what the setter
was compensating for. Where the companion was a hand-rolled optional — a
`faceColorSet`-style flag — the field widened to `optional.Value[T]` and the
flag was deleted (`Patch.FaceColor`/`EdgeColor`/`EdgeWidth`, `Line2D.GapColor`,
`PathCollection.OffsetCoords`), so the field stays exported and struct literals
keep working. Where the setter clamps, normalizes, or fires a callback, the
field was unexported behind a same-named reader (`Axes.Title`/`XLabel`/`YLabel`,
`Figure.SupTitle`/`SupXLabel`/`SupYLabel`, `Slider.Value`, `TextBox.Value`,
`RadioButtons.Active`, `RangeSlider.Low`/`High`). Four redundant spellings went:
the duplicate `SetSuptitle`/`SetSupxlabel`/`SetSupylabel` casings, the alias
`SetLocator` pair, `Axis.SetTickDirection` (replaced by `ParseTickDirection`,
closing a raw-string enum the typed-constant pass missed), and the misnamed
`SetMarkEvery` (now `SetMarkEverySpec`). `PathCollection`'s slice setters and
the `Line2D` dash/marker-spec pairs are deliberately deferred with reasons in
`docs/plans/phase2-mutable-fields.md`. Goldens stayed byte-identical. Core and
plot3d alpha multiplier paths share
`render.Color.WithAlphaMultiplier`, and 3D scalar maps derive their
configuration through
`core.PlotOptions.ScalarMapConfig`, with no golden fixture changes. `just test`
passes all packages, including the golden and Matplotlib-reference checks.

## Phase 3: Visual QA and Tolerance Closure

**Goal:** inspect every case whose tolerance can hide a visible divergence,
record its disposition, and hand Phase 4 a defensible frozen tolerance set.
Run after Phase 2.

- [ ] Review `widgets_gallery` and `animation_gallery` and replace their
      remaining loose RMSE allowances with measured, binding thresholds.
- [ ] Inspect every current `MaxRMSE >= 4` case side-by-side with Matplotlib.
      Fix real divergences; document acceptable raster/text differences and
      upstream differences; then ratchet each tolerance to measured output plus
      small headroom. The queue includes:
  - `skewt_basic`, `annotation_legend_offsetbox_gallery`, `text_bbox_styles`,
    `text_layout_gallery`, `mathtext_accents`, `mathtext_fractions`,
    `mathtext_inline_labels`, `projection_toolkit_gallery`,
    `specialty_depth`, `formatter_log_mathtext_labels`, `named_colors`,
    `stem_plot`, `stem_horizontal`, `lognorm_imshow`,
    `colorbar_variants_gallery`, `errorbar_capthick`, `patch_style_matrix`,
    `text_annotation_matrix`, `scatter_gallery`,
    `animation_subplots_frame`, `animation_line_frame`,
    `animation_scatter_frame`, `mplot3d_stem3d`, and
    `scale_logit_ticks`.
- [ ] Inspect dense low-PSNR cases where RMSE alone is weak:
      `mathtext_gallery`, `image_variants_gallery`,
      `triangulation_gallery`, and `pcolormesh_gouraud`; add binding PSNR
      floors where warranted.
- [ ] Commit a per-case disposition table (`fixed`, `documented exception`, or
      `upstream difference`) alongside the catalog and update every reviewed
      tolerance.

**Done when:** no catalog case has an effectively disabled gate, every case
with `MaxRMSE >= 4` has a written disposition, and Phase 4 receives the
ratcheted tolerance set.

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

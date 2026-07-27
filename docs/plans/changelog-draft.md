# Phase 4 Changelog Draft

This fragment records the coordinated pre-v1 API breaks made during Phase 2.
Phase 2 is closed, so the break list below is complete: it covers the package
split, the error convention, the options model, and the mutable-field cleanup.
It still needs release framing and a version heading before it seeds the v1.0
`CHANGELOG.md`; it is not a release announcement.

## Breaking changes

- Moved 3D axes and projection APIs from `core` to `plot3d`. Construct 3D axes
  with `plot3d.AddAxes` or `plot3d.NewAxes`.
- Moved numeric tick locators and formatters from `core` to `ticker`, and date
  conversion, epoch, locator, formatter, and `strftime` APIs to `dates`.
- Moved widgets, selectors, cursors, and their axes helpers from `core` to
  `widgets`.
- Removed the string-keyed Python introspection/property surface (`Setp`,
  `Getp`, `GetpAll`, `Findobj`, `FindobjType`, and `PropertyBag`). Use typed
  artist APIs and explicit ownership instead.
- Removed `render.CapabilityBridgeReporter`. Runtime backend capability
  reporting is typed through `backends.CapabilityStatus`.
- Moved `render.RendererModeReporter` to `backends.RendererModeReporter`.
- Renamed exported `GetX` accessors to Go-style nouns or explicit lookup
  names. See `docs/matplotlib-migration-notes.md` for the complete mapping.
- Renamed the renderer drawing method `Image(image, destination)` to
  `DrawImage(image, destination)`; raster exporters now expose their buffer
  through `Image()`.
- Folded `Axes.PlotUnits` into `Axes.Plot`. The primary method now accepts
  unit-capable slice values and returns `(*Line2D, error)` for rejected input;
  `PlotDate`, `pyplot.Plot`, and `pyplot.PlotDate` propagate the same error.
- Folded `Axes.ScatterUnits` into `Axes.Scatter`. The primary method and
  `pyplot.Scatter` now accept unit-capable slice values and return
  `(*Scatter2D, error)` with transactional rejection.
- Folded `Axes.BarUnits` into `Axes.Bar` and `Axes.BarH`. The primary methods
  and their pyplot wrappers now accept unit-capable slice values and return
  `(*Bar2D, error)` with transactional rejection while preserving categorical
  position-axis locators.
- Folded `Axes.FillBetweenUnits` into `Axes.FillBetween`. The primary method
  and `pyplot.FillBetween` now accept unit-capable slice values and return
  `(*Fill2D, error)`; all three inputs are validated before any axis-unit state
  is committed.
- Converted the remaining warn-and-skip plotting entry points to the error
  convention. `Axes.FillBetweenX`, `Axes.FillBetweenPlot`, `Axes.Hist`,
  `Axes.ErrorBar`, `Axes.ErrorBarContainer`, `Axes.ImShowRGB`, and the
  `pyplot.FillBetweenX`, `pyplot.Hist`, and `pyplot.ErrorBar` wrappers return
  `(T, error)` instead of emitting a diagnostic and returning a nil artist.
  Empty input, mismatched slice lengths, and extra option values are now
  rejected, and `Axes.ErrorBar` validates before the property cycle advances.
  `diag.Warnf` is reserved for artists accepted with a documented degradation;
  the audit is in `docs/plans/warn-and-skip-inventory.md`.
- Rejected extra option values across the whole variadic plotting surface. A
  call that supplies two or more option structs used to keep the first and
  discard the rest. Entry points that return an error now report a
  `*optarg.TooManyError`; the rest panic, because the extra value can only come
  from a literal at the call site. No signature changed, and calls that pass
  zero or one option value behave exactly as before.
- Adopted the final options model and migrated one API per artist family to it.
  `Axes.ImShow`, `Axes.Stem`, `Axes.Annotate`, `Axes.HLines`, `Axes.VLines`, and
  their pyplot wrappers now take exactly one options value instead of a variadic
  tail, so extra option sets are a compile error. Their optional fields moved
  from pointers and magic zero values to `optional.Value[T]`, which makes
  previously inexpressible requests — alpha 0, a zero-width arrow, an annotation
  offset of (0, 0), an explicit upper image origin against a `lower` rc — work.
  `Axes.HLines`/`Axes.VLines` take the new `core.LineCollectionOptions` rather
  than the `LineCollection` artist. See
  `docs/plans/options-model.md`.
- Completed the options migration across every remaining family. All 205
  variadic `...FooOptions` tails are gone: `core`, `pyplot`, `plot3d`, and
  `widgets` entry points each take exactly one options value, so a second one is
  a compile error everywhere. All 408 pointer-to-primitive option fields are now
  `optional.Value[T]`; the only pointers left in an options struct are
  references to live objects such as `*core.Figure` and `*log.Logger`. Fields
  whose zero value already was the default were demoted to plain values
  (`PlotOptions.DrawStyle`, `PlotOptions.MarkerFaceAlt`,
  `PlotOptions.MarkEverySpec`, `StairsOptions.Fill`, `StepOptions.Where`).
- Replaced the raw-string option enums with defined string types in `core`:
  `PlotOrientation`, `ColorbarExtend`, `ColorbarLocation`, `ImageAspect`,
  `VectorPivot`, and `ViolinSide`, each with named constants. String literals
  still compile; a plain `string` variable now needs an explicit conversion.
- Removed `internal/optarg` and the `*optarg.TooManyError` it produced. With no
  variadic tails left there is no extra option set to reject at run time, so
  code that matched on that error can drop the branch.
- Removed `core.ScalarMapConfig`'s `*float64` limits in favour of
  `optional.Value[float64]`. `PlotOptions.ScalarMapConfig` no longer hands out
  pointers that alias the caller's variables.
- Gave every artist field exactly one public writer. Twenty-nine exported fields
  were shadowed by a `Set<Field>` method that did more than assign, so the field
  write silently skipped part of the work. Where the setter was compensating for
  a hand-rolled optional, the field widened instead: `Patch.FaceColor`,
  `Patch.EdgeColor`, `Patch.EdgeWidth`, `Line2D.GapColor`, and
  `PathCollection.OffsetCoords` are now `optional.Value[T]`, and the companion
  `faceColorSet`/`edgeColorSet`/`edgeWidthSet`/`GapColorSet`/`offsetCoordsSet`
  flags are gone. These fields stay exported, so struct literals keep working —
  wrap the value in `optional.Of`.
- Encapsulated the artist fields whose setters clamp, normalize, or notify.
  `Axes.Title`, `XLabel`, `YLabel`, `Figure.SupTitle`, `SupXLabel`, `SupYLabel`,
  `Slider.Value`, `TextBox.Value`, `RadioButtons.Active`, `RangeSlider.Low`, and
  `RangeSlider.High` are unexported; a reader of the same name replaces each
  (`ax.Title()`), and the setter is the only writer. Direct writes used to skip
  `ensureRCTextDefaults` on the axes side and the clamping, caret repositioning,
  and on-change callback on the widget side.
- Removed the duplicate `Figure.SetSuptitle`, `SetSupxlabel`, and `SetSupylabel`
  spellings; the Go-cased `SetSupTitle`/`SetSupXLabel`/`SetSupYLabel` remain.
- Removed `Legend.SetLocator` and `AnchoredTextBox.SetLocator`, which were pure
  aliases for the exported `Locator` field. Assign the field.
- Removed `Axes.AddWidget` and the `core.WidgetArtist` interface. `Axes.Add`
  detects artists that implement the (now unexported) widget interface and
  routes them to the widget layer, so widgets are added like any other artist.
  A custom widget still joins that layer by implementing `Draw`, `Z`, `Bounds`,
  and `WidgetLayer`.
- Moved the widget axes helpers with the widgets themselves:
  `Figure.AddWidgetAxes`, `SubFigure.AddWidgetAxes`, and
  `SubplotSpec.AddWidgetAxes` are now `widgets.NewAxes`,
  `widgets.NewSubFigureAxes`, and `widgets.NewSubplotAxes`.
- Replaced `Axis.SetTickDirection(string) error` with
  `core.ParseTickDirection(string) (TickDirection, error)`; assign the result to
  the already-typed `Axis.TickDirection`. This closes a raw-string enum the
  typed-constant pass missed. `AxisArtist.SetTickDirection` is unchanged.
- Renamed `Line2D.SetMarkEvery` to `SetMarkEverySpec`, matching the
  `MarkEverySpec` field it actually writes; it never touched `MarkEvery`.
- Merged `Dashes` and `DashUnits` into a single `core.DashPattern` value on
  `Line2D` and `Patch`. Build one with `core.PixelDashes(...)` or
  `core.MatplotlibDashes(...)`; `pattern.Scaled(lineWidth)` returns the device
  lengths. Assigning the sequence alone used to leave the units at pixels and
  silently render a Matplotlib pattern at the wrong scale. The `DashUnits` type
  and its constants are unchanged, and `Line2D.SetDashes` keeps its signature.
- Removed `Line2D.MarkerFaceColor` and `Line2D.MarkerEdgeColor`. Both were read
  only while the neighboring `MarkerFaceSpec`/`MarkerEdgeSpec` was unset, so a
  color now goes through `core.ExplicitMarkerColor(c)`. A zero-alpha legacy color
  meant "inherit the line color", which `MarkerColorDefault` and
  `core.AutoMarkerColor()` already express. `SetMarkerFaceColor` and
  `SetMarkerEdgeColor` are unchanged.
- Resolved collection `EdgeColorsFace` (Matplotlib `edgecolors="face"`) when the
  stroke color is read rather than by mirroring `FaceColors` into `EdgeColors` at
  write time, on `PathCollection`, `PatchCollection`, and `QuadMesh`. Assigning
  `FaceColors` directly now matches `SetFaceColors`. Reading `EdgeColors` no
  longer shows the mirrored faces, and `SetEdgeColors` no longer clears
  `EdgeColorsFace` — clear it to make explicit edge colors win.

## Added

- Added `Figure.Save`, `Figure.WriteTo`, and `Figure.Image` as the common
  figure-output surface. Import `backends/all` or a specific output backend to
  register the desired formats.
- Added `render.Color.WithAlphaMultiplier` as the shared non-mutating color
  alpha composition primitive used by core and 3D artists.
- Added `core.PlotOptions.ScalarMapConfig` so core and 3D plotting paths use
  one typed scalar-map configuration conversion.
- Added the `optional` package. `optional.Value[T]` is the tri-state optional
  field used by the migrated option structs, replacing pointer fields and
  zero-value sentinels.
- Added an explicit concurrency contract for rc state, registries, pyplot
  state, figures, axes, artists, and renderers.
- Added synchronized backend, desktop-constructor, and color-sequence
  registries.

## Internal and documentation changes

- Classified every declaration in the pre-break 3,102-symbol API snapshot as
  keep, demote, or delete.
- Re-froze the public API after the package moves and intentional deletions.
- Remapped parity-catalog source ownership to the new packages without changing
  render behavior or golden images.
- Took the final Phase 2 freeze after the error, options, and mutable-field
  work: `test/testdata/public_api/stable_public_api.json` held 3,176 symbols
  across 29 packages, and that artifact is the surface Phase 3 and Phase 4 are
  measured against. Regenerating it, the parity-status document, and the
  public-surface classifications produced no diff, so the committed artifacts
  already matched the code when Phase 2 closed. The Phase 2 follow-ups then
  added the two shared contour scalar-map helpers, and Phase 3.3.4 added the
  `transform.AffineScale`/`IsAffineScale` pair, bringing the freeze to 3,180.
- Reconciled that freeze against the Phase 2.1 tiering decisions symbol by
  symbol in `docs/plans/api-freeze-delta.{md,json}`. All 19 `delete` rows are
  absent and the one `demote` row landed in `backends`; every one of the 402
  removed and 480 added names carries a category and, where it is a rename or a
  move, a replacement that the test verifies is frozen.

# Phase 4 Changelog Draft

This fragment records the coordinated pre-v1 API breaks completed so far
during Phase 2. It must be extended with the remaining error/options work
before it seeds the v1.0 `CHANGELOG.md`; it is not a release announcement.

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
  the audit is in `docs/plans/phase2-warn-and-skip-inventory.md`.

## Added

- Added `Figure.Save`, `Figure.WriteTo`, and `Figure.Image` as the common
  figure-output surface. Import `backends/all` or a specific output backend to
  register the desired formats.
- Added `render.Color.WithAlphaMultiplier` as the shared non-mutating color
  alpha composition primitive used by core and 3D artists.
- Added `core.PlotOptions.ScalarMapConfig` so core and 3D plotting paths use
  one typed scalar-map configuration conversion.
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

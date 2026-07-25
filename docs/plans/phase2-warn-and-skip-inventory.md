# Phase 2.3: warn-and-skip plotting entry point inventory

Phase 2.3 adopts a single error convention: a plotting call that **rejects**
its input returns `(T, error)` and leaves the axes untouched, while `diag.Warnf`
is reserved for calls that **accept** an artist with a documented degradation.

This document records the audit of every `diag.Warnf` site that a plotting call
could reach, the disposition of each, and the reasoning. It is the closing
artifact for the PLAN.md bullet "Inventory the remaining warn-and-skip plotting
entry points, convert rejected input to errors, and retain warnings only where
an artist is accepted with a documented degradation."

## Converted: rejected input now returns an error

Each of these sites warned and then returned a nil artist, so the caller had no
programmatic way to distinguish "rejected" from "nothing to draw". All five are
now error-returning, and all of them validate before the property cycle
advances, so a rejected call leaves the axes, its artists, and its cycle
unchanged.

| Entry point            | Former warning                                                                       | Rejections now reported as errors                                                                                                                                                            |
| ---------------------- | ------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Axes.FillBetweenX`    | `FillBetweenX: where length … skipping`                                              | nil axes, more than one `FillOptions`, empty `y`/`x1`/`x2`, unequal lengths, `Where` length mismatch                                                                                         |
| `Axes.FillBetweenPlot` | `FillBetween: where length … skipping`                                               | nil axes, more than one `FillOptions`, empty `x`/`y1`/`y2`, unequal lengths, `Where` length mismatch                                                                                         |
| `Axes.Hist`            | `Hist: weights length … skipping`                                                    | nil axes, more than one `HistOptions`, empty data, `Weights` length mismatch                                                                                                                 |
| `Axes.ErrorBar`        | `ErrorBar: error/limit arrays … skipping`, `ErrorBar: invalid errorevery … skipping` | nil axes, more than one `ErrorBarOptions`, empty `x`/`y`, error or limit arrays that are neither scalar nor per-point, negative or non-finite errors, invalid `ErrorEvery`/`ErrorEveryStart` |
| `Axes.ImShowRGB`       | `ImShowRGB: …; skipping`                                                             | nil axes, more than one `ImShowRGBOptions`, ragged or non-`(M,N,{1,3,4})` arrays, single-channel data that is not a finite rectangular matrix                                                |

Two adjacent signatures moved with them:

- `Axes.ErrorBarContainer` forwards the `Axes.ErrorBar` error rather than
  collapsing it to a nil container.
- `pyplot.FillBetweenX`, `pyplot.Hist`, and `pyplot.ErrorBar` propagate the
  error from the current axes, matching the already-converted `pyplot.Plot`,
  `pyplot.Scatter`, `pyplot.Bar`, and `pyplot.FillBetween`.

Two internal callers discard the new error deliberately, each with the reason
recorded at the call site:

- `Bar2D.addErrorBars` — `validateBarInput` already rejected every malformed
  error or limit array before the bar artist was constructed.
- `Axes.StackPlot` and `Axes.HistMulti` — they build equal-length, non-empty
  slices with no `Where` mask or weights, so the inner call cannot be rejected.

## Retained: artist accepted with a documented degradation

These warnings stay. In every case a real artist is added or drawn; the warning
records that some requested detail was approximated.

| Site                                    | Degradation                                                                                                      |
| --------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `core/container.go` `stemIsHorizontal`  | An unrecognized `Orientation` string falls back to vertical stems; the stem plot is still created.               |
| `core/collection_quadmesh.go`           | A renderer without Gouraud triangle support gets flat cell shading. Draw-time capability fallback, not input.    |
| `core/image.go`                         | A renderer without rotated-image support draws the image axis-aligned. Draw-time capability fallback, not input. |
| `core/image_rgba.go`                    | Out-of-range RGB(A) channel values are clipped to `[0,1]`. Matplotlib clips and warns here too.                  |
| `core/mathtext_warn.go`                 | An unknown mathtext command renders as literal text, matching Matplotlib.                                        |
| `core/text_bbox.go`                     | An unresolvable bbox style falls back to a square box.                                                           |
| `plot3d/axes.go`, `plot3d/bar_voxel.go` | `PlotSurface`/`Voxel` draw the simplified artist and point at the full-fidelity method.                          |

## Known divergences left for a later bullet

- `stemIsHorizontal` accepts an invalid orientation string where Matplotlib
  raises `ValueError`. The orientation argument is a raw-string enum, and the
  Phase 2.3 options bullet ("Replace raw-string enums with typed constants")
  makes the invalid value unrepresentable, which is a better fix than adding an
  error return now.
- `color/colormap.go` falls back to the default colormap for an unknown name
  where Matplotlib raises. This is a registry lookup rather than a plotting
  entry point, so it belongs with the colormap/registry surface rather than
  this pass.
- The `style/` rcParam warnings describe intentionally unsupported or unhonored
  settings; they are configuration diagnostics, not plot input.

## Entry points that still return a bare nil

Plotting methods outside this inventory (for example `Axes.Fill`,
`Axes.ImShow`, `Axes.Stem`, and the signal helpers) can still return a nil
artist for degenerate input without warning. They were never warn-and-skip
sites, so they are outside this bullet; the options rework and the final Phase
2.4 re-freeze are the point at which their signatures are settled.

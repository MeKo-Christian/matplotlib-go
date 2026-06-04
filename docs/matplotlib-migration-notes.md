# Matplotlib to Go Migration Notes

This project keeps Matplotlib's plotting model and visual behavior where it can,
but the Go API uses typed option structs instead of Python's dynamic keyword
and property-dict conventions.

## Phase 17.75.2 Axes Helpers

The 17.75.2 helpers cover common Matplotlib migration calls with intentionally
typed signatures:

- `Axes.Bxp` and `pyplot.Bxp` accept `[]core.BxpStat` plus
  `core.BxpOptions`, instead of Python's list of dictionaries and separate
  `boxprops`, `whiskerprops`, `capprops`, and related property dictionaries.
- `Axes.Violin` and `pyplot.Violin` accept `[]core.ViolinStat` plus
  `core.ViolinStatsOptions`, mirroring Matplotlib's precomputed-stat contract
  while keeping styling and visibility flags in a typed options value.
- `Axes.HLines`, `Axes.VLines`, `pyplot.HLines`, and `pyplot.VLines` accept
  slice inputs and a `core.LineCollection` style value. Single-value endpoints
  or extents broadcast across multiple positions like Matplotlib; masked-array
  handling and Python's full color/linestyle alias grammar remain out of scope.
- `Axes.Clabel`, `ContourSet.Clabel`, and `pyplot.Clabel` accept a
  `core.ClabelOptions` value for level filtering, formatter, font size, color,
  inline spacing, and manual label positions. GUI-driven manual placement and
  exact Python `Text` artist compatibility remain partial.

The parity fixture `axes_convenience_helpers` exercises these helpers against
Matplotlib 3.10.9 reference output.

### Source Alignment

The covered helpers were checked against the vendored Matplotlib 3.10.9 source
in `third_party/matplotlib`:

- `Axes.HLines` and `Axes.VLines` track
  `lib/matplotlib/axes/_axes.py` `Axes.hlines` and `Axes.vlines` by
  normalizing scalar-like endpoints/extents, creating data-space two-point
  segments, registering a `LineCollection`, and updating data limits through
  the collection path. The Go API accepts slices and one optional typed
  `LineCollection`; Python `data=`, masked arrays, arbitrary `**kwargs`, and
  color/linestyle alias normalization are intentionally not mirrored.
- `Axes.Bxp` tracks `Axes.bxp` by consuming precomputed median/quartile/whisker
  stats, deriving default positions, widths, and cap widths, drawing boxes,
  whiskers, caps, medians, optional means/fliers, and optionally managing tick
  labels. Go returns a `core.BxpContainer` of typed `Line2D` groups and does
  not implement Python's `patch_artist`, rcParam/property-dict merging, pending
  `vert` deprecation path, or dynamic legend/property alias behavior.
- `Axes.Violin` tracks `Axes.violin` by consuming precomputed `coords`/`vals`
  density stats, normalizing density to half-width, honoring
  vertical/horizontal orientation and `both`/`low`/`high` sides, and drawing
  body, mean, median, extrema, and quantile collections. Go keeps collection
  return values in `core.ViolinContainer` and leaves Python's exact return-dict
  key shape and rcParam color-cycle details partial.
- `Axes.Clabel`, `ContourSet.Clabel`, and `pyplot.Clabel` track
  `Axes.clabel` and `ContourLabeler.clabel`: the axes method delegates to the
  contour set, level filtering must match existing contour levels, formatters
  and colors are applied to generated labels, and manual iterable positions are
  supported. GUI/manual event-loop placement, `rightside_up`,
  `use_clabeltext`, and exact `Text` artist return semantics remain out of
  scope for the typed Go surface.

## Phase 17.75.3 Axes Option Breadth

The 17.75.3 option-breadth work covers visible high-use options through typed
Go option structs:

- `Axes.Hist` and `pyplot.Hist` support weighted samples, explicit histogram
  ranges, density normalization based on the in-range weighted total, and
  right-to-left cumulative density through `ReverseCumulative`.
- `Axes.Scatter` supports per-point sizes, per-point face/edge colors, scalar
  values mapped through a colormap/norm, and Matplotlib's default
  edgecolor-face behavior for scalar-mapped markers.
- `Axes.Bar`, `Axes.BarH`, and `Axes.BarLabel` support per-bar widths/colors,
  edge alignment, explicit baselines, and edge labels that use the stacked bar
  endpoint rather than only the segment height.
- `Axes.FillBetween` and `Axes.FillBetweenX` support `where` masks,
  interpolation at crossing boundaries, and `pre`/`post`/`mid` step expansion.
- `Axes.ErrorBar` supports `errorevery` stride/start selection while keeping
  the data line complete.
- Path collections and quad meshes expose typed mutable setters for visible
  scalar-map, offset, style, and rectilinear-edge changes.

These were checked against vendored Matplotlib 3.10.9 in
`third_party/matplotlib/lib/matplotlib/axes/_axes.py` and
`third_party/matplotlib/lib/matplotlib/collections.py`. The Go API keeps
explicit fields such as `core.HistOptions.Weights`,
`core.ScatterOptions.ScalarValues`, `core.BarOptions.Align`,
`core.FillOptions.Where`, and `core.ErrorBarOptions.ErrorEvery` instead of
Python's dynamic `**kwargs`, `data=`, property dictionaries, and overload
grammar. Multi-dataset histogram inputs, broad alias normalization, masked
array input objects, and exact Python container/property mutation semantics
remain partial.

The parity fixture `axes_option_breadth_17_75_3` exercises visible scatter,
bar-label, fill-between, and errorbar option behavior against Matplotlib 3.10.9
reference output. Collection and mesh mutable-setter behavior is covered by
focused unit tests and existing mesh/colorbar fixtures where it produces
visible output.

## Phase 17.75.4 mplot3d Depth and Clipping

The 17.75.4 depth/clipping audit aligns common static mplot3d helpers with
Matplotlib 3.10.9 while keeping the Go API typed:

- `Axes3D.Plot3D`, `Scatter3D`, `Wireframe`, `Quiver3D`, `Stem3D`, and
  `ErrorBar3D` support explicit-limit clipping through typed `AxLimClip`
  fields and reproject their geometry when view limits or camera state change.
- `Axes3D.Surface`, `Contour`, `Contourf`, `TriContour`, `TriContourf`,
  `Trisurf`, `Bar3D`, and `FillBetween3D` use projected collection z-ordering
  and face/polygon ordering where the renderer needs stable static output.
  Offset-plane contour clipping and typed `AxLimClip` options follow
  Matplotlib's `axlim_clip` intent.
- `Axes3D.Voxels` culls adjacent internal faces, sorts visible faces by
  projected depth, supports typed `AxLimClip`, and clears stale projected
  geometry when a view-limit reprojection removes a voxel.

Intentional differences remain: Go does not mirror Python's full positional
overload grammar, `data=` dispatch, masked-array machinery, GUI event methods,
or mutable `Poly3DCollection`/`Line3DCollection` internals. Projection and depth
ordering are implemented as deterministic pre-projected 2D artists, so rare
interpenetrating 3D geometry can still differ from Matplotlib's painter-order
heuristics.

## Phase 17.75.4 mplot3d Axis Defaults and View State

The 17.75.4 axis/view audit checked the Go `Axes3D` state against
`mpl_toolkits/mplot3d/axes3d.py` and `axis3d.py` in the vendored Matplotlib
3.10.9 source:

- View defaults match Matplotlib's static-rendering defaults: elevation `30`,
  azimuth `-60`, roll `0`, vertical axis `z`, camera distance `10`, and the
  4:4:3 box-aspect scale used by mplot3d.
- `Axes3D.SetViewInit` carries the full view state needed for roll and
  vertical-axis projection, while `SetProjectionType("persp"|"ortho")`
  mirrors Matplotlib's projection-type validation and focal-length behavior for
  static rendering.
- Explicit x/y/z view limits are kept in caller order so inverted 3D axes
  project like Matplotlib. Tick generation and `AxLimClip` checks still use
  the numeric inclusive range so inverted axes retain visible ticks, labels,
  and clipped artists.
- Pane selection, grid segments, axis lines, tick directions, pane colors,
  rc-derived grid/axis colors, line widths, and grid dash styling are aligned
  with the mplot3d frame logic where they affect static output.
- Tick-label and axis-label placement uses Matplotlib's rough point-to-data
  offset model, includes endpoint tick labels, keeps Unicode-minus formatting,
  and supports typed x/y/z tick-label visibility toggles.

Intentional differences remain: Go exposes typed setters and immutable option
values rather than Python's mutable `Axes3D` artist objects, shared-view axes,
`data=` dispatch, GUI mouse rotation/pan/zoom callbacks, or arbitrary keyword
mutation. `Axes3D.View()` reports the legacy elevation/azimuth/distance triple;
roll, vertical-axis, projection type, and focal length are controlled through
dedicated typed methods.

## Phase 17.75.4 mplot3d Scalar-Mappable Inventory

The 17.75.4 colormapping audit maps each public Go 3D helper to Matplotlib's
collection type and scalar-mappable behavior:

| Go helper | Matplotlib surface | Upstream mappable behavior | Current Go state |
| --- | --- | --- | --- |
| `Axes3D.Surface` / `PlotSurfaceGrid` | `Axes3D.plot_surface` -> `Poly3DCollection` | With `cmap`, `set_array(avg_z)`, `set_clim`, and `set_norm`; with `facecolors`, explicit per-face colors; otherwise a solid color, optionally shaded. | `PlotOptions.Colormap`, `Norm`, `VMin`, and `VMax` populate average-z `PolyCollection` scalar-map metadata; explicit `FaceColors`, edge-color behavior, alpha, and colorbar creation/update through collection setters are supported. |
| `Axes3D.Trisurf` | `Axes3D.plot_trisurf` -> `Poly3DCollection` | With `cmap`, scalar array is per-triangle average z; `norm`/`vmin`/`vmax` propagate. Without `cmap`, uses explicit/next color and optional shading. | Colormap, norm, clim metadata, per-triangle average-z arrays, explicit edge color, alpha, and colorbar creation/update through collection setters exist on the returned `PolyCollection`. |
| `Axes3D.Contour` / `TriContour` | `Axes3D.contour` / `tricontour` -> 3D contour collections | 2D contour set is converted to 3D line collections; level colors are scalar-mapped unless explicit colors override them. | Level colors, scalar arrays, scalar-map metadata, alpha, and colorbars are exposed; explicit colors clear scalar-map state like Matplotlib. |
| `Axes3D.Contourf` / `TriContourf` | `Axes3D.contourf` / `tricontourf` -> `Poly3DCollection` bands | Filled contour bands carry level-based scalar-map state unless explicit colors override them. | Filled bands expose colormap/norm/clim metadata, filled-level autoscaling arrays, alpha, and colorbars; explicit colors clear scalar-map state like Matplotlib. |
| `Axes3D.Scatter3D` | `Axes3D.scatter` -> `Path3DCollection` | Delegates to 2D `Axes.scatter`; numeric `c` remains scalar-mappable, then colors are depth-shaded and z-sorted. | Returned `Scatter2D` supports `ScalarValues`, `Colormap`, `Norm`, `VMin`, and `VMax`; scalar-mapped colors respect scatter alpha before depth shading and remain colorbar-compatible. |
| `Axes3D.Quiver3D`, `Wireframe`, `ErrorBar3D`, `Stem3D`, `Plot3D` | `Line3DCollection` or `Line2D`-derived artists | Primarily explicit line colors/kwargs; not scalar-mappable in the common mplot3d examples. | Typed color, alpha, width, and label options exist; scalar mapping is intentionally not treated as the default surface for these line helpers. |
| `Axes3D.Bar3D` | `Axes3D.bar3d` -> `Poly3DCollection` | Accepts single, per-bar, six-face, or per-face color arrays; shade/lightsource apply to facecolors. It is not a scalar-array mappable by default. | Typed color, alpha, width, shaded projected faces, and single/per-bar/six-face/`6*N` face-color variants are covered; it remains explicit-color-only by default. |
| `Axes3D.FillBetween3D` | `Axes3D.fill_between` -> `Poly3DCollection` | Accepts `facecolors`, optional shade, and forwards collection kwargs; not scalar-array mappable by default. | Typed color, edge color, alpha, mode, clipping, and non-scalar-mappable collection behavior are covered. |
| `Axes3D.Voxels` | `Axes3D.voxels` -> per-voxel `Poly3DCollection` dict | Accepts scalar or shaped facecolor/edgecolor arrays; shade/lightsource apply per voxel; not scalar-array mappable by default. | Typed per-voxel face/edge color maps, default shading, alpha propagation, visible-face sorting, and non-scalar-mappable collection behavior are covered. |

This inventory keeps scalar-array colorbars scoped to the helpers that
Matplotlib exposes as scalar mappables in normal static examples, while
explicit-color 3D helpers remain typed color surfaces unless a later task adds
dedicated scalar-array APIs.

### 3D scalar-mappable colorbar contract

Matplotlib 3.10.9 colorbars consume the public scalar-mappable surface rather
than private 3D projection data: `Colorbar` reads the mappable's `cmap`, `norm`,
clim, `get_array`, and artist alpha, with callback wiring for later
`changed()` updates. The mplot3d helpers preserve that contract by returning
scalar-mappable collection types where upstream examples expect colorbars:
`plot_surface` / `plot_trisurf` return `Poly3DCollection` instances that call
`set_array` with average face z values when `cmap` is present; `scatter`
returns a `Path3DCollection` from the 2D scatter path; contour helpers return
`ContourSet` values whose levels and color values drive colorbar boundaries.
`voxels`, `bar3d`, `fill_between`, quiver, stem, and line helpers use explicit
color collection or line surfaces by default rather than scalar arrays.

The Go colorbar path intentionally uses a narrower typed contract:
`Figure.AddColorbar` accepts `ScalarMappable`, reads `ScalarMap()` for colormap,
norm, and clim, and keeps the mappable handle so `syncColorbarMapping` can
refresh mutable clim/colormap changes. Shared collections additionally expose
`GetArray()` as a Matplotlib-style audit surface for scalar values, but
`GetArray()` is not required by `AddColorbar` itself. Therefore the 3D colorbar
integration audit treats collection-backed helpers as compatible when they
return a `ScalarMappable` with matching `ScalarMap()` metadata and, where the
underlying Matplotlib collection calls `set_array`, a matching `GetArray()`
shape for the colorbar-driving scalar values.

The helper-level audit currently classifies the Go 3D helpers this way:

| Helper group | Colorbar-compatible state | Notes |
| --- | --- | --- |
| `Surface` / `PlotSurfaceGrid` | yes | Returned `PolyCollection` exposes average-z `GetArray()`, cmap, norm, and clim metadata when `Colormap` is set. |
| `Trisurf` | yes | Returned `PolyCollection` exposes per-triangle average-z `GetArray()`, cmap, norm, and clim metadata when `Colormap` is set. |
| `Contour` / `TriContour` | yes | Returned `LineCollection` exposes level arrays and scalar-map metadata unless explicit colors disable the scalar map. |
| `Contourf` / `TriContourf` | yes | Returned `PolyCollection` exposes filled-band layer arrays and scalar-map metadata unless explicit colors disable the scalar map. |
| `Scatter3D` | yes | Returned `Scatter2D` exposes `ScalarMap()` and `GetArray()` directly, while projected `PathCollection` draw state keeps the depth-shaded mapped colors. |
| `Voxels`, `Bar3D`, `FillBetween3D`, `Quiver3D`, `Wireframe`, `ErrorBar3D`, `Stem3D`, `Plot3D` | no by default | These helpers follow Matplotlib's common explicit-color collection or line surfaces and are not scalar-array colorbar sources unless a future typed scalar-array API is added. |

Some explicit-color helpers return collection types that satisfy Go's generic
`ScalarMappable` interface because the shared collection base exposes
`ScalarMap()`. When those helpers have no scalar array or scalar-map metadata,
forcing `Figure.AddColorbar` on the returned collection creates only the
generic default `viridis` 0..1 colorbar. That is not treated as a supported
data-backed 3D colorbar; supported 3D colorbar sources are the helpers in the
`yes` rows above.

### 3D explicit-color and immutable omissions

The explicit-color-only 3D helpers intentionally do not grow scalar-array
colorbar APIs in 17.75.4. This follows the normal upstream construction paths
in `third_party/matplotlib/lib/mpl_toolkits/mplot3d/axes3d.py`: `plot`
forwards kwargs to 2D `Axes.plot`, `plot_wireframe` constructs a
`Line3DCollection`, `errorbar` constructs `Line3D` cap markers plus
`Line3DCollection` error segments, and `bar3d`, `voxels`, `quiver`, `stem`, and
`fill_between` use explicit color kwargs or collection face/edge colors by
default. Go keeps `Bar3D`, `Voxels`, `FillBetween3D`, `Quiver3D`, `Wireframe`,
`ErrorBar3D`, `Stem3D`, and `Plot3D` as typed explicit-color surfaces unless a
future task adds a dedicated scalar-array API for one of those helpers.

The mutable-state omissions are also intentional for this phase. Matplotlib's
`ScalarMappable` / `ColorizingArtist` callback lifecycle and
`Colorbar.update_normal` path are not mirrored directly. Go supports
post-creation 3D colorbar synchronization through explicit collection setters
and pull-based draw/layout sync; it does not promise automatic callback
registration, colorbar alpha propagation from 3D artists, mutable `Scatter3D`
scalar arrays, or persistence of manual collection-array overrides across later
3D reprojection.

### 3D scalar-mappable mutable update audit

Matplotlib's scalar-mappable update model is callback-driven. A
`ScalarMappable` / `ColorizingArtist` owns an array (`set_array` /
`get_array`), cmap, norm, and clim; changing cmap or norm calls `changed()`,
changing clim updates the norm and emits the norm's `changed` signal, and
`Colorbar` registers `mappable.callbacks.connect('changed',
Colorbar.update_normal)`. `Colorbar.update_normal` then pulls the mappable's
alpha, cmap, and norm, resets locator/formatter state only when the norm object
changes, redraws the colorbar, and keeps contour-specific line overlays in
sync. During draw, colorbar processing also calls `mappable.autoscale_None()`
when an array exists so unscaled norms can pick up array-derived limits.

The Go colorbar model is pull-based. Shared collection types expose typed
`SetArray`, `SetColormap`, `SetNorm`, and `SetCLim` methods that refresh their
own scalar-derived colors and scalar-map metadata. `Figure.AddColorbar` stores
the `ScalarMappable` handle, and `syncColorbarMapping` re-reads
`ScalarMap()` during layout/draw to update colorbar cmap, norm, and clim.
There is no callback registry on 3D mappables, no mutable `SetArray` API on the
returned `Scatter2D`, and colorbar alpha is currently independent from 3D
artist alpha. For 3D collection-backed helpers, mutable cmap/norm/clim updates
can follow the shared collection setters; for `Scatter3D`, array and mapping
state are treated as construction-time values unless a later typed scatter
mutation API is added.

Decision for 17.75.4: 3D colorbars support post-creation updates only through
the existing typed collection setters on the returned collection objects
(`SetArray`, `SetColormap`, `SetNorm`, and `SetCLim`) plus the pull-based
colorbar sync that runs during layout/draw. This covers collection-backed
`Surface`, `Trisurf`, `Contour`, `Contourf`, `TriContour`, and `TriContourf`
handles when callers intentionally mutate the returned collection. It does not
promise Matplotlib's callback lifecycle, alpha propagation into existing
colorbars, mutable scatter arrays, or persistence of manual collection-array
overrides across later 3D view/limit reprojection. Reprojection closures remain
owned by the original 3D helper inputs and may recompute projection-derived
arrays such as surface average-z, contour levels, filled-contour layer values,
or scatter visible sorted scalars.

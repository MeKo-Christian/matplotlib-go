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

## Phase 17.75.5 Color Conversion

`color.ToRGBA` covers the Matplotlib color forms that map cleanly to typed Go:
single-letter base colors, CSS4/X11 names, `tab:` Tableau colors, `xkcd:`
survey names, case-insensitive named colors except single-letter aliases,
`none`, `C0`-style color-cycle references, `#rgb`, `#rgba`, `#rrggbb`,
`#rrggbbaa`, grayscale strings in `[0,1]`, `render.Color`, Go
`image/color.Color`, and numeric RGB/RGBA slices or arrays.

The error surface is compared by failure category rather than exact Python
wording: alpha range, malformed hex, grayscale-string range, invalid RGBA
argument, sequence length, and RGBA channel range. Go diagnostics may be more
specific than Matplotlib's `Invalid RGBA argument` text, but they are kept in
the same broad conversion categories.

Python-only dynamic input forms are intentionally not mirrored by `ToRGBA`:
color-alpha tuples such as `("red", 0.5)`, `to_rgba_array` batch conversion
inputs such as lists of grayscale strings or alpha arrays, NumPy masked values
and masked arrays, and Matplotlib's scalar-array bad-value handling where a
NaN RGB component can pass through `to_rgba`. Use explicit typed Go alpha
options, numeric RGB/RGBA slices, and scalar-mappable image/collection APIs
instead.

The public parity ledger records this as a typed `colors.py` partial: `ToRGBA`
is the supported conversion entry point, while Python convenience helpers such
as `to_rgb`, `to_hex`, `same_color`, `is_color_like`, and `to_rgba_array`
remain intentionally omitted unless a later migration fixture needs dedicated
typed APIs.

## Phase 17.75.5 Norm Inventory

Current scalar-mappable normalization is exposed through the typed
`core.ScalarNormalizer` interface rather than Matplotlib's mutable `Normalize`
class hierarchy. The Phase 17.75.5 inventory maps the upstream `colors.py`
norm surface to Go as follows:

| Matplotlib norm | Go surface | Related axis scale | Colorbar route |
| --- | --- | --- | --- |
| `Normalize` | `core.Normalize` | `linear` | default linear scalar-map axis |
| `LogNorm` | `core.LogNorm` | `log` | log colorbar scale and log ticks |
| `SymLogNorm` | `core.SymLogNorm` | `symlog` | function colorbar scale through the norm inverse |
| `AsinhNorm` | `core.AsinhNorm` | `asinh` | asinh colorbar scale with linear-width metadata |
| `PowerNorm` | `core.PowerNorm` | none | function colorbar scale through the norm inverse |
| `TwoSlopeNorm` | `core.TwoSlopeNorm` | none | function colorbar scale through the norm inverse |
| `CenteredNorm` | `core.CenteredNorm` | none | function colorbar scale through the norm inverse |
| `BoundaryNorm` | `core.BoundaryNorm` | none | boundary ticks, boundaries, values, and extensions |
| `NoNorm` | `core.NoNorm` | none | index-style scalar map on a linear colorbar axis |
| `FuncNorm` | custom `core.ScalarNormalizer` implementations | `function`, `functionlog` scales exist for axes | accepted through the interface, with no concrete `FuncNorm` clone |

This keeps axes scales and color normalization deliberately separate: axis
`function`/`functionlog` scales do not automatically become color norms, and
custom color normalization should implement `ScalarNormalizer` directly.

## Phase 17.75.5 Norm Gap List

Supported parity fixtures currently exercise the norm behavior needed by
visible examples:

- `lognorm_imshow` covers `LogNorm` image mapping and log colorbar ticks.
- `twoslope_norm_image` covers diverging `TwoSlopeNorm` colorbar placement
  with an inverse-backed function scale.
- `boundarynorm_pcolormesh` and `colorbar_boundary_values` cover
  `BoundaryNorm`, explicit boundaries/values, and colorbar extensions.
- `asinh_norm_image` covers `AsinhNorm` image mapping and its colorbar scale.

No supported parity fixture currently requires `FuncNorm`, `MultiNorm`, or a
multi-stage normalization pipeline. `FuncNorm` remains represented by custom
`ScalarNormalizer` implementations until a fixture needs a concrete typed
constructor and callback policy. `MultiNorm` is not present as a Matplotlib
3.10.9 `colors.py` public class, so any later multi-stage norm support should
be driven by a visible example rather than added speculatively.

Known norm deviations to resolve before broadening the surface are Matplotlib's
mutable scalar-mappable callback lifecycle, exact `clip` edge behavior for
nonlinear norms beyond the covered fixtures, and richer boundary/under/over
interactions when a colorbar is updated after construction.

## Phase 17.75.5 FuncNorm Upstream Contract

In Matplotlib 3.10.9, `FuncNorm` is not handwritten normalization logic. It is
generated by `make_norm_from_scale` using `scale.FuncScale`, so its behavior is
the generic scale-backed norm contract:

- The constructor accepts a `functions` two-tuple plus `vmin`, `vmax`, and
  `clip`; the functions are forwarded to `scale.FuncScale`.
- The forward and inverse functions are both required, and the forward function
  must be monotonic.
- `__call__` autoscale-fills missing limits, optionally applies `clip`, maps
  values through the scale transform, normalizes transformed values between
  transformed `vmin` and `vmax`, and masks invalid transformed results.
- `inverse` rescales from `[0, 1]` through the inverse transform and requires a
  scaled norm.
- `autoscale_None` only considers transform-domain finite values.

The Go scale registry already has axis-level `function` and `functionlog`
scales with forward and inverse callbacks, but a color `FuncNorm` decision must
choose a typed `ScalarNormalizer` constructor, the callback shape, and how much
of Matplotlib's array-like, mask, `clip`, and autoscale behavior to mirror.

## Phase 17.75.5 FuncNorm Go Callback Shape

The current Go callback shape for arbitrary color normalization remains the
existing `ScalarNormalizer` interface:

- `Map(float64) float64`
- `Inverse(float64) (float64, bool)`
- `Autoscale([]float64) ScalarNormalizer`
- `Range() (float64, float64)`
- `Validate() error`
- `NormName() string`

This deliberately does not share the axis `transform.Scale` callback shape.
Axis scales map data coordinates into axes fraction space, while color norms
also carry scalar-map range, validation, autoscale, and colorbar identity
metadata. Until a supported fixture needs Matplotlib-compatible `FuncNorm`
construction, no concrete `FuncNorm` type is added; callers that need arbitrary
normalization should provide a typed `ScalarNormalizer` implementation.

## Phase 17.75.5 FuncNorm Omission Ledger

`FuncNorm` is an intentional omission as a concrete Go type. The affected
examples are currently none: no supported parity fixture requires it.
`asinh_norm_image`, `twoslope_norm_image`, `lognorm_imshow`, and
`boundarynorm_pcolormesh` are covered by concrete typed norms, and the other
supported colorbar examples use explicit `Normalize` or `BoundaryNorm`
behavior.

The rationale is API scope. Matplotlib's `FuncNorm` is array-like, transform
scale-backed, mask-aware, and callback-friendly; adding that surface without a
fixture would create a broad compatibility promise that the current typed API
does not need. The fallback recommendation is to implement `ScalarNormalizer`
directly for custom color normalization. Axis-level custom transforms continue
to use `transform.Scale` through `function` and `functionlog`, which remains
separate from scalar-mappable color normalization.

## Phase 17.75.5 Norm Public Surface Metadata

The public-surface ledger now gives norm behavior explicit rows instead of
leaving it inside the broad `colors.py` partial classification. `Normalize`,
`SymLogNorm`, `PowerNorm`, `TwoSlopeNorm`, `CenteredNorm`, `BoundaryNorm`,
`NoNorm`, and `AsinhNorm` are marked `idiomatic-equivalent`: Go exposes their
behavior as concrete `ScalarNormalizer` values rather than Matplotlib's mutable
normalizer classes, and the focused norm/scalar-mappable tests plus the
`asinh_norm_image`, `twoslope_norm_image`, `boundarynorm_pcolormesh`,
`colorbar_boundary_values`, and `colorbar_extensions` fixtures cover the
currently supported behavior.

`FuncNorm` and `make_norm_from_scale` are marked `intentional-omission` as
public Go constructors. The supported route for arbitrary color normalization
is still a caller-provided `ScalarNormalizer`, while axis-level custom
transforms continue to use `transform.Scale`. `LogNorm` remains covered by the
norm inventory and `lognorm_imshow`; Matplotlib 3.10.9 generates it through
`make_norm_from_scale`, so the committed public-surface extraction does not
carry a separate `colors.py:class:LogNorm` row.

`docs/matplotlib-parity-status.md` was regenerated after those row decisions.
The open broad `colors.py` rows now point to the explicit norm rows and no
longer list implemented normalizers as unresolved Phase 17.75.5 partials.

## Phase 17.75.5 Scalar-Mappable Norm Update Audit

Matplotlib `Colorizer` connects norm callbacks into a changed-event pipeline:
norm changes trigger the colorizer, scalar mappables forward `changed`, and
attached colorbars listen through `mappable.callbacks.connect`. `set_clim`
blocks norm callbacks, changes `vmin` and `vmax` together, then emits one norm
change. `Colorbar.update_normal` responds to mappable changes and resets
locator/formatter/scale state when the norm object changes.

Go uses pull-based scalar-map updates instead. Mappables expose `ScalarMap()`;
mutable collections and meshes expose `SetNorm`, `SetCLim`, `SetArray`, and
`SetColormap`; colorbars keep the mappable handle and refresh through
`syncColorbarMapping` during layout/draw. There is no callback registry for
norm changes. The current supported behavior is explicit mutation followed by
redraw, with colorbars pulling the latest norm and clim from the mappable.

## Phase 17.75.5 Scalar-Mappable Norm Update Decision

The supported Go update path is explicit mutation followed by redraw:
`SetArray`, `SetColormap`, `SetNorm`, and `SetCLim` update the mappable state
and mark it stale where the artist supports mutable scalar maps. A colorbar
stores the mappable handle, so during layout or drawing the colorbar pulls the
latest `ScalarMap()` state and updates its scale, limits, colormap, and norm
metadata.

Matplotlib-style callback registry remains intentionally omitted for this
surface. Callers should mutate the typed mappable and redraw the figure; they
should not expect norm objects to notify colorbars independently.

## Phase 17.75.5 LightSource Algorithm Inventory

Matplotlib's `colors.LightSource` is a 2D elevation-image lighting helper, not
the same code path as mplot3d collection face shading. Its constructor defaults
are `azdeg=315` and `altdeg=45`, with HSV blend limits
`hsv_min_val=0`, `hsv_max_val=1`, `hsv_min_sat=1`, and `hsv_max_sat=0`.
The `direction` property converts clockwise-from-north azimuth into
counterclockwise-from-east math coordinates with `90 - azdeg`, then combines
that azimuth with altitude as a unit vector.

`hillshade` takes a 2D elevation array and defaults `vert_exag=1`, `dx=1`,
`dy=1`, and `fraction=1`. The row spacing is inverted with `dy = -dy` to match
image/raster orientation, gradients are computed as
`np.gradient(vert_exag * elevation, dy, dx)`, normals are assembled as
`(-e_dx, -e_dy, 1)`, normalized, and passed to `shade_normals`. In
`shade_normals`, `fraction` multiplies the dot-product intensity before
Matplotlib rescales by the original intensity min/max when the range exceeds
`1e-6`, then clips the result to `[0, 1]`.

`shade` first colormaps data with an optional norm and defaults to
`blend_mode='overlay'`; `shade_rgb` consumes an existing RGB image and defaults
to `blend_mode='hsv'`. The built-in blend paths are `blend_overlay`,
`blend_soft_light`, and `blend_hsv`, with callable blend functions also
accepted upstream. Overlay uses the low/high branch formulas around
`rgb <= 0.5`, soft-light uses the pegtop formula
`2 * intensity * rgb + (1 - 2 * intensity) * rgb**2`, and HSV shifts
saturation/value toward the configured min/max HSV limits after converting the
hillshade intensity to `[-1, 1]`.

3D collection face shading is separate. Matplotlib's `art3d._shade_colors`
uses `LightSource(azdeg=225, altdeg=19.4712)` by default, maps the face-normal
dot product from `[-1, 1]` to `[0.3, 1]`, multiplies RGB by that shade, and
preserves alpha. The current Go 3D shading helper mirrors that mplot3d face
shading route; it does not implement the 2D `LightSource.hillshade`,
`shade`, or `shade_rgb` image-lighting API.

## Phase 17.75.5 LightSource Example Need List

No committed Python parity fixture imports `LightSource`, passes
`lightsource=`, or calls the 2D image-lighting helpers. No 2D image fixture
currently calls `hillshade`, `shade`, or `shade_rgb`; the image fixtures are
plain imshow/matshow/alpha/interpolation cases rather than shaded-relief image
examples.

The current references that mention shading are mplot3d examples, and they use
Matplotlib's separate collection-face shading path:

- `mplot3d_terrain`, `mplot3d_basic`, and `mplot3d_surface3d` call
  `plot_surface(..., cmap="viridis")` or another colormap; Matplotlib disables
  surface face shading when a colormap is present.
- `mplot3d_bar3d` and `mplot3d_voxels` rely on default mplot3d face shading
  for explicit colors, which the Go `shade3DFaceColor` path already mirrors.
- `mplot3d_trisurf3d` uses a colormap and therefore does not require
  LightSource-driven face shading; explicit-color trisurf shading is covered by
  focused Go tests for `shade3DFaceColor`.

This means the supported parity set does not require a broad `LightSource` API
or 2D shaded-image integration yet. The next implementation decision can either
keep LightSource as an explicit omission with this fixture audit, or add a
small static subset only if a new visual fixture exercises hillshade/blend
behavior.

## Phase 17.75.5 LightSource Hillshade Core Decision

`hillshade` remains intentionally omitted for the current Go public API. No
`core.LightSource` or `color.LightSource` type is added, and there is no
standalone grayscale hillshade constructor that promises Matplotlib's
`azdeg=315`, `altdeg=45`, `vert_exag=1`, `dx=1`, `dy=1`, and `fraction=1`
semantics.

The reason is fixture scope rather than algorithm complexity. No committed
parity fixture requires grayscale hillshade output, and the supported mplot3d
examples exercise collection-face shading instead of 2D elevation-image
lighting. `shade3DFaceColor` remains the supported 3D face-shading path for
solid-color bars, voxels, and explicit-color triangulated surfaces.

Revisit this decision when a shaded-relief image fixture is added. At that
point the implementation should start from the upstream audit above: inverted
image-row spacing, `np.gradient`-style derivative handling, normalized
`(-e_dx, -e_dy, 1)` normals, fraction scaling before min/max rescale, and final
`[0, 1]` clipping.

## Phase 17.75.5 LightSource RGB Blend Mode Decision

`shade` and `shade_rgb` remain intentionally omitted along with the
LightSource RGB blend-mode surface. That means there is no public Go equivalent
for `blend_overlay`, `blend_soft_light`, `blend_hsv`, or Matplotlib's callable
blend modes in the current API.

No committed parity fixture requires RGB shaded-relief blend output. Adding
only one of the upstream blend formulas would create an incomplete public
contract, while adding all modes would still be unverified without a visual
fixture that exercises the LightSource image-lighting path. For now, the Go
port should not expose a partial LightSource blend API.

Existing colormap lookup and mplot3d face shading remain unchanged by this
omission. Colormapped images continue to use normal scalar mapping, and
explicit-color 3D faces continue to use the separate `shade3DFaceColor`
implementation documented above.

## Phase 17.75.5 LightSource Image Path Integration Decision

Because the LightSource hillshade and RGB blend APIs are omitted, no
LightSource path is connected to `Image2D`, `imshow`, `matshow`, or
transformed-image rendering. The AGG image backend remains a scalar-image
renderer: it receives already-colormapped or scalar-mapped image data and does
not synthesize shaded-relief RGB output from an elevation array.

No static image fixture requires shaded-relief rendering. Keeping the image
path unchanged avoids coupling image resampling to unimplemented
hillshade/blend semantics and avoids introducing a renderer-only feature that
would have no public `LightSource` API or visual parity fixture. A future
shaded-relief fixture should add the typed lighting surface first, then connect
it through image creation before backend resampling.

## Phase 17.75.5 LightSource Surface Path Integration Decision

For mplot3d, no public `LightSource` object is connected to `Surface`,
`Trisurf`, `Bar3D`, or `Voxels`. The supported integration remains the
Matplotlib 3D face-shading subset already implemented in core:
`shade3DFaceColor` remains the supported mplot3d face-shading implementation,
using `LightSource(azdeg=225, altdeg=19.4712)`, mapping face-normal dot
products into `[0.3, 1]`, multiplying RGB, and preserving alpha.

Colormapped `plot_surface` and `plot_trisurf` paths keep shading disabled like
Matplotlib, while explicit-color trisurf, bar3d, and voxel paths keep using
the existing face-normal shading helper. This preserves the 17.75.4 3D
semantics and avoids changing colorbar/scalar-map behavior for colormapped
surfaces.

Do not add a `lightsource` option to these paths until a visual fixture needs
custom mplot3d light-source positioning. Such a fixture should compare against
`art3d._shade_colors`, not the 2D `colors.LightSource.shade_rgb` image path.

## Phase 17.75.5 LightSource Fixture Decision

No new LightSource or shaded-image visual triplet is added for this phase.
`hillshade`, `shade`, and `shade_rgb` reference fixtures are deferred because
the committed example set does not exercise 2D shaded-relief image lighting.

`mplot3d_terrain` remains an mplot3d surface fixture, not a 2D LightSource
fixture: its colormapped `plot_surface` path disables Matplotlib face shading,
and its solid-color trisurf path is covered by the existing mplot3d face-shade
tests. The existing 3D fixtures cover the separate face-shading path for bars,
voxels, and explicit-color surfaces without adding the 2D `LightSource` API.

Add a dedicated visual triplet before implementing the 2D LightSource API. The
triplet should call one of `hillshade`, `shade`, or `shade_rgb` directly so the
Go port has a concrete reference for gradient, fraction, blend-mode, alpha, and
mask behavior.

## Phase 17.75.5 LightSource Metadata Update

The public-surface ledger now marks `colors.py:class:LightSource` as
`intentional-omission`. The row is anchored to `mplot3d_terrain` only as a
fixture-audit reference: that case uses mplot3d surface rendering and does not
exercise `LightSource.hillshade`, `shade`, or `shade_rgb` image lighting.

`docs/matplotlib-parity-status.md` was regenerated after the metadata update,
so `LightSource` no longer appears as an open Phase 17.75.5 partial row. The
broad `colors.py` rows now point to explicit norm, dynamic norm-factory, and
LightSource rows; remaining open color-surface work is the Python-only dynamic
color input surface plus bivariate/multivariate colormap decisions.

## Phase 17.75.5 Bivar/Multivar Upstream API Inventory

Matplotlib 3.10.9 treats bivariate and multivariate colormaps as separate
public color-mapping surfaces rather than ordinary scalar colormap names.
`MultivarColormap` wraps two or more scalar colormaps and combines component
RGBA results with a `combination_mode`: `sRGB_add` sums RGB channels while
`sRGB_sub` subtracts the combined complement. It multiplies component alpha,
tracks transparent bad values, supports tuple-shaped input matching the number
of variates, validates alpha shape against the first input, optionally clips
RGB/alpha to `[0, 1]`, rejects `bytes=True` when clipping is disabled, and has
per-component resampling plus `with_extremes` support.

`BivarColormap` is a two-input lookup-table surface with `N` and `M`
quantization dimensions, `origin`, transparent bad color, magenta outside
color, and shape modes `square`, `circle`, `ignore`, and `circleignore`.
Its call path requires a two-part input, applies shape-specific clipping or
outside marking, maps float inputs by multiplying by `N`/`M` while treating
exact `1.0` as the last in-range index, indexes `_lut[X0, X1]`, applies bad and
outside colors, optionally returns bytes, and validates alpha shape against the
first input. It also exposes 1D component colormaps through `__getitem__`,
resampling/reversal/transposition helpers, circular display masking, and
`with_extremes` for bad/outside/shape/origin.

`SegmentedBivarColormap` builds a bivariate table from a `(k, l, 3)` patch by
supersampling into an `(N, N, 4)` LUT with bilinear image resampling.
`BivarColormapFromImage` accepts `(N, M, 3)` or `(N, M, 4)` lookup tables,
converts uint8 data to floats, and adds opaque alpha when only RGB is supplied.
`cm.py` registers generated families through `_bivar_colormaps` and
`_multivar_colormaps`; scalar-mappable integration is multi-variate because
the colormap input is a tuple of component arrays, not one normalized scalar
array.

## Phase 17.75.5 Bivar/Multivar Go Fit Assessment

The current Go color API is intentionally scalar. `color.Colormap` maps one
normalized `float64` to one `render.Color` through `At` / `AtValue`.
`ScalarMapInfo` stores one colormap name, one norm, and one scalar range, and
`ScalarMapInfo.Color` maps one scalar value plus alpha into a display color.
That model fits Matplotlib's scalar colormaps, norms, and scalar-mappable
colorbars.

Bivariate and multivariate colormaps do not fit the current scalar colormap
model as an overload. They require tuple or vector input, variate counts,
shape-specific clipping/outside behavior, multi-dimensional lookup tables,
component-colormap mixing, and different bad/outside/alpha semantics. They
also imply colorbar expectations beyond one scalar axis: a bivariate LUT needs
2D display semantics, while a multivariate colormap exposes multiple component
colormaps.

If support is added later, it should be a separate typed surface rather than a
hidden mode of `color.Colormap`. The scalar colormap path should stay stable
for existing `ScalarMappable` artists, and colorbar expectations would also
need a new contract before bivariate or multivariate mappables become public.

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

### 3D fixture coverage (17.75.4.5)

The 3D fixture sweep closed the inventory gaps with seven new parity triplets
(Go `test/parity/<id>/plot.go` + `plot.py`, Python `test/matplotlib_ref/plots`,
golden, and matplotlib reference): `mplot3d_errorbar3d`, `mplot3d_contour3d`,
`mplot3d_contourf3d`, `mplot3d_tricontour3d`, `mplot3d_tricontourf3d`,
`mplot3d_bar2d_zdir`, and `mplot3d_text3d`. Each keeps the Go and Python
sources structurally close and shares identical input data so only rendering
differences remain. The structured-contour cases reuse the
`get_test_data(0.25)` dual-Gaussian grid with explicit levels and matched
`vmin`/`vmax`; the triangulated-contour cases reuse the `mplot3d_trisurf3d`
polar fan point cloud and rely on auto-Delaunay so the Go `core.Triangulation`
mesh and matplotlib's qhull mesh agree. All seven join the existing `mplot3d_*`
family in `optionalVisualGoldenIDs` and use the shared 3D tolerance band
(`MinPSNR` 30, `MaxMeanAbs` 8-12, `MaxRMSE` 18).

Measured golden-vs-reference metrics: contour3d RMSE 3.17 / PSNR 52.0,
contourf3d RMSE 6.06 / PSNR 45.4, tricontour3d RMSE 4.62 / PSNR 52.0,
tricontourf3d RMSE 13.0 / PSNR 46.5, bar2d_zdir RMSE 1.09 / PSNR 56.8, and
text3d RMSE 14.79 / PSNR 45.3.

Residual rendering differences (all within the band):

- Filled 3D contour bands (`Contourf` / `TriContourf`) render marginally more
  transparent than Matplotlib's `Poly3DCollection`, so `mplot3d_contourf3d` and
  `mplot3d_tricontourf3d` carry the band's higher RMSE; the band positions,
  level colors, and depth order match.
- `mplot3d_text3d` sets explicit `SetXLim`/`SetYLim`/`SetZLim` in both the Go
  and Python sources because Go's `Text3D` expands the data limits while
  Matplotlib's `text` does not participate in 3D autoscaling. With shared
  limits the projection is identical; only flat (`zdir=None`) labels are used,
  since Go does not rotate 3D text along an axis direction.
- `mplot3d_bar2d_zdir` uses a single base color per plane (and fixed, not
  random, heights) because the Go plane-bar API takes one color rather than
  Matplotlib's per-bar color array; the per-plane depth ordering, alpha, and
  edges otherwise match.

## Phase 17.75.4 mplot3d Public-Surface Summary

The 17.75.4 closure brings the static Go `Axes3D` surface in line with the
Matplotlib 3.10.9 behavior used by the parity catalog. The implemented behavior
is covered by focused unit tests plus the `mplot3d_*` fixture family:
`mplot3d_plot3d`, `mplot3d_scatter3d`, `mplot3d_surface3d`, `mplot3d_wire3d`,
`mplot3d_trisurf3d`, `mplot3d_bar3d`, `mplot3d_voxels`, `mplot3d_quiver3d`,
`mplot3d_stem3d`, `mplot3d_fill_between3d`, `mplot3d_errorbar3d`,
`mplot3d_contour3d`, `mplot3d_contourf3d`, `mplot3d_tricontour3d`,
`mplot3d_tricontourf3d`, `mplot3d_bar2d_zdir`, and `mplot3d_text3d`.

For migration purposes, the supported static behavior is:

- Matplotlib-style view defaults, roll, vertical-axis projection, perspective
  and orthographic projection type, focal-length validation, explicit/inverted
  x/y/z limits, autoscale margins, pane/grid/tick styling, and 3D label/tick
  placement.
- Projected depth ordering, explicit-limit clipping, and view/limit
  reprojection for lines, markers, surfaces, contours, triangulated contours,
  bars, voxels, quiver, stems, error bars, and fill-between surfaces.
- Scalar-mappable metadata and colorbar compatibility for surface/trisurf,
  structured and triangulated contour/contourf, and scalar-colored scatter;
  explicit-color helpers remain explicit-color surfaces by default.
- Matplotlib-reference coverage for the previously weak or missing 3D fixture
  cases: structured contour, filled structured contour, triangulated contour,
  filled triangulated contour, 3D error bars, projected 2D bars, and 3D text.

Retained Go-specific differences are intentional and should be treated as API
differences, not fixture bugs:

- Go uses typed option structs and explicit setters instead of Matplotlib's
  dynamic keyword grammar, `data=` dispatch, mutable artist-property strings,
  and Python collection internals. Prefer typed `PlotOptions`, explicit
  x/y/z-limit setters, and returned collection handles.
- Colorbar updates are pull-based through `Figure.AddColorbar` and typed
  collection setters (`SetArray`, `SetColormap`, `SetNorm`, `SetCLim`) rather
  than Matplotlib's callback registry and `Colorbar.update_normal` lifecycle.
- 3D text currently renders flat projected text and participates in Go data
  limit expansion. Match Matplotlib's non-autoscaling text behavior by setting
  explicit 3D limits when text positions should not affect the plotted data
  range.
- GUI/event-only methods, Python overload aliases, masked-array machinery, and
  rare interpenetrating-geometry painter-order differences remain outside the
  static typed surface covered by 17.75.4.

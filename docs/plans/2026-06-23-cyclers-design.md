# Property Cyclers & bundled stylesheets

Roadmap item: _"Cyclers for `linestyle`/`marker`/`linewidth`; bundle the common
`.mplstyle` sheets (`seaborn-_`, `fivethirtyeight`, `bmh`, `Solarize_Light2`)
(`style/theme.go:23`)."\*

## Goal

Match matplotlib's `axes.prop_cycle`, which is a generic `Cycler` combining any
set of artist properties (not just color). A plot pulls one step from the shared
cycle per line and applies each cycled property only where the user has not set
it explicitly. Plus: ship matplotlib's standard `.mplstyle` library so users can
`style.MustTheme("fivethirtyeight")` etc.

## Decisions (confirmed with user)

1. **Full cycler algebra** — port `Cycler` faithfully (`+` concat, `*` product).
2. **`go:embed` the real `.mplstyle` files** rather than hand-translating.
3. **Unified cycle** — a single shared iterator; `NextColor` returns its color
   column, line-drawing sites additionally consume linestyle/marker/linewidth.

## Components

### `cycler/` (new top-level package)

Generic finite cycler.

```go
type Cycler struct { keys []string; rows []map[string]any }
func New(key string, values ...any) *Cycler
func (c *Cycler) Concat(o *Cycler) (*Cycler, error) // '+': equal len, disjoint keys
func (c *Cycler) Multiply(o *Cycler) *Cycler        // '*': cartesian product
func (c *Cycler) Len() int
func (c *Cycler) Keys() []string
func (c *Cycler) Row(i int) map[string]any          // i mod Len()
func (c *Cycler) ByKey(key string) []any
```

Values are `any`: `render.Color` for `color`, `string` for `linestyle`/`marker`,
`float64` for `linewidth`. Consumers type-assert per key. `+` errors on unequal
length or shared keys (matplotlib raises `ValueError`). `*` requires disjoint
keys.

### `style`

- `RC.PropCycle *cycler.Cycler` (nil ⇒ color-only; preserves today's behavior).
- `WithPropCycle(*cycler.Cycler) Option`; `WithColorCycle` unchanged.
- `RC.Palette()` returns the `color` column of `PropCycle` when present, else
  `ColorCycle`. `Apply`/round-trip clones `PropCycle`.
- `parseMPLColorCycle` → `parseMPLPropCycle`: parse `cycler(...)` with `+`/`*`,
  positional (`'color', [...]`) and keyword (`color=[...]`) forms, keys
  color / linestyle|ls / marker / linewidth|lw. Sets both `PropCycle` and the
  derived `ColorCycle`.

### `core`

- Keep exported `Axes.ColorCycle`/`PatchColorCycle` (`*color.ColorCycle`) — they
  still own the index, so no public Axes API change.
- Add unexported `Axes.lineCycle`/`patchCycle *cycler.Cycler` from `RC.PropCycle`.
- `LineProps{ Color; LineStyle string + ok; Marker MarkerType + ok; LineWidth + ok }`.
- `NextLineProps()`: `idx := ColorCycle.Index(); col := NextColor()`; pull
  ls/marker/lw from `lineCycle.Row(idx)`.
- Line sites (`Plot`, `plot_variants`, line-collection helpers, 3D lines) apply
  cycled ls→Dashes (`lineStyleToDashes`), marker (`MarkerTypeFromString`), lw —
  **only where the corresponding `opt.*` is nil**. Non-line `NextColor` sites are
  untouched (they advance the shared index and ignore extra props, which is
  correct — those artists do not cycle ls/marker).

### Stylesheets

Copy matplotlib 3.10.9 `mpl-data/stylelib/*.mplstyle` into `style/stylelib/`,
`//go:embed` them, register each at `init` via the existing `LoadMPLStyle`
parser + `RegisterTheme` — **register-if-absent**, so the hardcoded
`default`/`ggplot`/`dark_background`/`publication` themes (and their goldens) are
not overridden.

## Testing

- `cycler/cycler_test.go`: New, `+` (len/key errors), `*` product, `Row` wrap,
  `ByKey`.
- `style/mplstyle_test.go`: multi-key prop_cycle (`+`, `*`, keyword form),
  derived palette.
- `style/stylelib_test.go`: embedded themes registered (bmh, fivethirtyeight,
  solarize_light2, seaborn-v0_8…) with expected first cycle color; existing
  themes not clobbered.
- `core/*cycle_test.go`: successive `Plot` cycle ls/marker/lw; explicit opt
  overrides; color-only back-compat unchanged.

## Wrap-up

Regenerate frozen public-API JSON (`UPDATE_PUBLIC_API_AUDIT=1`) and parity-status
doc; `just fmt && just lint && just test`.

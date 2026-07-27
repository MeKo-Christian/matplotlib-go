# AGG Text Parity

This note keeps the AGG text parity model and validation loop in one place.
Use it when changing glyph rasterization, text measurement, tick/title
placement, or strict text fixtures.

## Reference Model

The source of truth is the vendored Matplotlib snapshot in
`third_party/matplotlib`:

- `lib/matplotlib/backends/backend_agg.py`: `RendererAgg.draw_text` prepares an
  `FT2Font`, lays out glyphs, then draws glyph bitmaps through
  `draw_text_image`.
- `src/ft2font.cpp`: `FT_Set_Char_Size` applies the DPI and hinting-factor
  setup used by Matplotlib's Agg renderer.
- `lib/matplotlib/text.py`: `_get_layout` combines run metrics with font-wide
  ascent, descent, and line gap. Pure ink bounds are not enough for vertical
  parity.
- `lib/matplotlib/mpl-data/matplotlibrc`: tick padding/alignment, title
  padding, font defaults, and text hinting defaults.

Local implementation points:

- `backends/agg/agg_text.go`, `backends/agg/text_raster.go`, and
  `backends/agg/freetype_native.go` own AGG text drawing and the pinned
  FreeType 2.6.1 link.
- `render/text_layout.go` and related `render/text_*` files own shared text
  layout and shaping helpers.
- `core/axis.go`, `core/text.go`, and `core/figure_draw.go` own tick, label,
  title, and artist placement.
- `test/golden_test.go`, `test/matplotlib_ref_test.go`, and
  `test/reference_compare_test.go` run the catalog-driven visual checks.

## Canary Cases

Keep these cases in the loop for text-rendering changes:

- `bar_basic_tick_labels`
- `bar_basic_title`
- `hist_strategies`
- `text_labels_strict`
- `title_strict`

The strict text cases run unconditionally (the historical
`RUN_OPTIONAL_VISUAL_TESTS` gate was removed earlier).

## Commands

Run the focused loop through the checked-in recipes:

```bash
just text-parity-backend
just text-parity-core
just text-parity-canaries
just text-parity-compare
```

Refresh only the text goldens after an intentional visual change:

```bash
just text-parity-golden
```

The equivalent catalog selectors are:

```bash
go test ./test -run 'TestMatplotlibRef/(bar_basic_tick_labels|bar_basic_title|hist_strategies|text_labels_strict|title_strict)$' -count=1 -v
go test ./test -run 'TestGolden/(bar_basic_tick_labels|bar_basic_title|hist_strategies|text_labels_strict|title_strict)$' -count=1 -update-golden -v
go test ./test -run 'TestReferenceCompare/(bar_basic_tick_labels|bar_basic_title|hist_strategies|text_labels_strict|title_strict)$' -count=1 -v
```

When a comparison fails, inspect the generated files under
`testdata/_artifacts/` or use:

```bash
just parity-viewer FILTER="bar_basic_tick_labels|bar_basic_title|hist_strategies|text_labels_strict|title_strict"
```

## Guardrails

- Prefer matching Matplotlib's pipeline over empirical offsets.
- Keep normal text on the bitmap glyph path; do not silently switch it to an
  outline-fill path.
- Do not regenerate unrelated goldens for a backend-text-only change.
- The default cgo AGG backend should link vendored FreeType 2.6.1. The
  `systemfreetype` tag is a compile fallback, not a parity target.

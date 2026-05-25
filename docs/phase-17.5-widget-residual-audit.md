# Phase 17.5 Widget Residual Audit

Date: 2026-05-25

This audit covers the `widgets_gallery` Matplotlib-compatible parity path after
splitting it from the Go-default user-facing showcase.

## Metrics

Before the Phase 17.5 compatibility path, `widgets_gallery` was listed in
`PLAN.md` as a residual offender at `TestMatplotlibRef` MeanAbs `6.41`.

Current focused verification after the 17.5.4 chrome-geometry pass:

| Check | Result |
| --- | --- |
| `TestMatplotlibRef/widgets_gallery` | PSNR `47.2 dB`, MeanAbs `0.79`, MaxDiff `255` |
| `TestReferenceCompare/widgets_gallery` | PSNR `47.24 dB`, MeanAbs `0.79`, RMSE `16.97`, MaxDiff `255` |
| `TestGolden/widgets_gallery` | MaxDiff `0`, MeanAbs `0.00`, PSNR `+Inf` |

Commands used:

```bash
rtk proxy go test -v ./test/ -run TestMatplotlibRef/widgets_gallery
rtk proxy go test -v ./test/ -run TestReferenceCompare/widgets_gallery
rtk proxy go test -v ./test/ -run TestGolden/widgets_gallery
```

## Residual Classification

**Layout:** Mostly closed for the parity fixture. The Go parity wrapper now uses
the same fixed figure axes rectangles as the Python reference instead of the
Go-default GridSpec showcase layout.

**Widget chrome:** Still the main intentional work item, but the first
Matplotlib-compatible geometry slice is complete. Constructor colors, panel
padding/radius/stroke, slider track/handle geometry, text-box panel/input
chrome, and check/radio marker geometry now route through the visual-style
policy. Remaining visible chrome residuals are mostly text placement: slider
label/value anchors, range-slider value text, and text-box label alignment.

**Selector geometry:** Mostly closed for the parity fixture. Static selector
regions in the parity wrapper are now rendered with patch/line primitives that
match the Python reference more closely than the interactive selector artists.
Further selector-widget chrome work should focus on the interactive selector
artists separately, not on this static parity fixture.

**Text metrics:** Minor but visible residuals remain in widget labels and value
text. Slider/range-slider labels and values do not yet match Matplotlib's
axes-coordinate text placement. Text-box value placement is closer after the
chrome pass, but label alignment remains Go-native.

**Cursor/multi-cursor behavior:** Not applicable to the static parity render
after the source-aligned wrapper change. The reference cursor lines are rendered
as ordinary `axvline`/`axhline` equivalents in the parity path. Interactive
cursor/multi-cursor behavior remains covered by the widget interaction tests.

**Renderer-boundary bugs:** Vertical text orientation is still visible in the
main y-axis label and should stay tracked under the broader renderer-boundary
work rather than being compensated in the widget fixture. Remaining antialiasing
differences are low-impact at the fixture level.

## Decisions

- Continue with core widget chrome changes for Matplotlib-compatible mode:
  slider/range-slider label/value anchors and the Matplotlib init line should
  move behind the style policy next.
- Do not change the Go-default showcase appearance while working on
  Matplotlib-compatible chrome, except where a shared geometry helper is proven
  to fix hit-testing correctness.
- Do not compensate for vertical text orientation inside `widgets_gallery`; keep
  that classified as renderer-boundary work.
- Treat the current static selector output as sufficient for this parity
  fixture. Interactive selector visuals can be refined later through focused
  selector tests.

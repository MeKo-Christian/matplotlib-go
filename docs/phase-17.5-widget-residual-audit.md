# Phase 17.5 Widget Residual Audit

Date: 2026-05-25

This audit covers the `widgets_gallery` Matplotlib-compatible parity path after
splitting it from the Go-default user-facing showcase.

## Metrics

Before the Phase 17.5 compatibility path, `widgets_gallery` was listed in
`PLAN.md` as a residual offender at `TestMatplotlibRef` MeanAbs `6.41`.

Current focused verification after completing the 17.5.4 chrome-policy pass:

| Check                                  | Result                                                       |
| -------------------------------------- | ------------------------------------------------------------ |
| `TestMatplotlibRef/widgets_gallery`    | PSNR `47.4 dB`, MeanAbs `0.73`, MaxDiff `255`                |
| `TestReferenceCompare/widgets_gallery` | PSNR `47.39 dB`, MeanAbs `0.73`, RMSE `15.81`, MaxDiff `255` |
| `TestGolden/widgets_gallery`           | MaxDiff `0`, MeanAbs `0.00`, PSNR `+Inf`                     |

Commands used:

```bash
rtk proxy go test -count=1 -v ./test/ -run 'TestGolden/widgets_gallery|TestMatplotlibRef/widgets_gallery|TestReferenceCompare/widgets_gallery'
```

## Residual Classification

**Layout:** Mostly closed for the parity fixture. The Go parity wrapper now uses
the same fixed figure axes rectangles as the Python reference instead of the
Go-default GridSpec showcase layout.

**Widget chrome:** Closed for the 17.5.4 parity pass. Constructor colors,
button hover face color, panel padding/radius/stroke, slider label/value
anchors, track/selection/handle geometry, the slider init line, range-slider
tuple value text, text-box label/value/caret anchors, and check/radio active
marker semantics now route through the visual-style policy.

**Selector geometry:** Mostly closed for the parity fixture. Static selector
regions in the parity wrapper are now rendered with patch/line primitives that
match the Python reference more closely than the interactive selector artists.
Further selector-widget chrome work should focus on the interactive selector
artists separately, not on this static parity fixture.

**Text metrics:** Minor residuals remain from renderer text metrics and glyph
placement, but widget label/value anchors now follow the Matplotlib-compatible
axes-coordinate layout instead of the Go-default internal layout.

**Cursor/multi-cursor behavior:** Not applicable to the static parity render
after the source-aligned wrapper change. The reference cursor lines are rendered
as ordinary `axvline`/`axhline` equivalents in the parity path. Interactive
cursor/multi-cursor behavior remains covered by the widget interaction tests.

**Renderer-boundary bugs:** Vertical text orientation is still visible in the
main y-axis label and should stay tracked under the broader renderer-boundary
work rather than being compensated in the widget fixture. Remaining antialiasing
differences are low-impact at the fixture level.

## Decisions

- Treat 17.5.4 widget chrome policy work as complete for this fixture. Future
  widget work should move to interaction/hit-testing coverage in 17.5.5 unless
  new visual regressions are found.
- Do not change the Go-default showcase appearance while working on
  Matplotlib-compatible chrome, except where a shared geometry helper is proven
  to fix hit-testing correctness.
- Do not compensate for vertical text orientation inside `widgets_gallery`; keep
  that classified as renderer-boundary work.
- Treat the current static selector output as sufficient for this parity
  fixture. Interactive selector visuals can be refined later through focused
  selector tests.

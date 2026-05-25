# ADR 0001: Desktop Interactive Backend Toolkit

- **Status:** Accepted
- **Date:** 2026-05-23
- **Phase / Task:** Phase 4 §4.2 — Desktop Interactive Backend
- **Related:** PLAN.md §4.2, §4.4; `canvas/dispatcher.go`, `canvas/scheduler.go`,
  `canvas/navigation.go`, `backends/agg`.

## Context

Phase 4.1 landed the event dispatcher, draw-idle scheduler, pan / zoom /
pick helpers, and hit testing for every artist family. The headless AGG
backend already produces the pixels we want to show, and `canvas.Navigation`
already mutates view limits and asks for a redraw. What is still missing
is a process that:

1. Opens an OS window sized to the figure.
2. Hands a fresh AGG `*image.RGBA` to that window on every `DrawIdle`
   tick.
3. Translates the toolkit's native mouse / key / resize / close events
   into `canvas.MouseEvent`, `canvas.KeyEvent`, `canvas.ResizeEvent`, and
   `canvas.CloseEvent` and emits them through the dispatcher.
4. Surfaces the standard `NavigationToolbar` actions (home, pan, zoom,
   save) without coupling the rest of matplotlib-go to one toolkit.

The shortlist from PLAN.md §4.2 is **Fyne, Ebiten, Gio, or a thin SDL2
binding**. Decision criteria, in order of weight:

1. **Pure-Go preference.** The repository builds without cgo on `purego`
   tags and runs in WASM (`canvas/wasm`, `cmd/wasm`); any dependency that
   forces cgo at link time would split the build matrix.
2. **AGG framebuffer embedding.** AGG produces an `*image.RGBA`. The host
   needs to upload that bitmap to a window surface every frame with
   minimal copying.
3. **Event fidelity.** We need press / release / move / scroll, key
   press / release with modifiers, resize, and close. Anything that
   coalesces or drops these events forces us to re-implement them.
4. **Embeddability.** The same toolkit should be usable both as a
   stand-alone figure window (`pyplot.Show()` equivalent) and as one
   element of a larger application — Matplotlib's `FigureCanvasQT` and
   `FigureCanvasTk` both support this.
5. **CI availability.** Tests must run in a headless container.

## Options considered

### Fyne (`fyne.io/fyne/v2`)

- Pros: mature widget toolkit; native menus and toolbars; large ecosystem;
  `canvas.NewImageFromImage` accepts `*image.RGBA` directly; reasonable
  documentation for embedding.
- Cons: requires cgo (OpenGL via `go-gl/glfw`) on every desktop platform,
  splitting our `purego` build story; redraw cadence is driven by widget
  invalidation rather than a tight tick loop, which fights §4.4's
  damage-region story; image refresh allocates / re-uploads the entire
  texture per frame.
- Embeddability: good — figures live inside any `fyne.Container`.
- CI: needs Xvfb or the test driver; doable but adds infrastructure.

### Ebiten (`github.com/hajimehoshi/ebiten/v2`)

- Pros: pure Go on Linux / Windows (purego on macOS); tight `Update` /
  `Draw` tick loop is a perfect host for §4.4's draw-idle scheduling;
  `ebiten.NewImageFromImage` plus `screen.DrawImage` is the minimum-cost
  blit; first-class touch / gamepad / clipboard fidelity.
- Cons: built around a single full-window game loop — it does not host
  itself inside larger GUI apps and provides no native widgets, so a
  toolbar must be hand-painted; the global `RunGame` model is hostile to
  multi-window figure managers; high-DPI scaling is per-frame manual.
- Embeddability: poor — designed to _own_ the window.
- CI: headless mode (`ebiten.RunGameWithOptions` with `SkipTaskbar` etc.)
  works but tests still need a display.

### Gio (`gioui.org`)

- Pros: pure Go across Linux / Windows / macOS / Android / iOS / WASM
  (no cgo on the supported platforms when using the Wayland / X11 / Win32
  backends bundled with Gio); explicit `app.Window` event loop maps 1:1
  to our `canvas.Dispatcher` (`pointer.Event` → mouse, `key.Event` →
  key, `app.FrameEvent` → draw); `paint.NewImageOp` blits an
  `*image.RGBA` with a single GPU upload per frame; supports multiple
  windows so multi-figure managers work; an immediate-mode UI layer
  means our toolbar can either be Gio widgets or AGG-painted bitmaps
  without changing the backend contract.
- Cons: smaller widget ecosystem than Fyne; immediate-mode mental model
  is unfamiliar to some contributors; styling defaults are minimal.
- Embeddability: excellent — a Gio `app.Window` can host any combination
  of figures plus app chrome.
- CI: `gio.app/headless` provides a software renderer that runs without
  a display server; this is what matches our existing headless test
  pattern.

### SDL2 binding (`github.com/veandco/go-sdl2`)

- Pros: well-understood event model; mature.
- Cons: cgo-only and links against the system `libSDL2`; pulls
  `libSDL2-dev` into every developer machine and CI image; we would
  still hand-roll a widget system for the toolbar; gives up the
  purego / WASM build matrix.
- Rejected on criterion 1 alone.

## Decision

**Adopt Gio (`gioui.org`) as the desktop interactive backend.**

It is the only option that satisfies all five criteria simultaneously:
pure-Go end-to-end with a headless test driver, a direct `*image.RGBA →
paint.NewImageOp` blit path, full event fidelity, multi-window
embeddability, and CI without system packages.

The chosen layering is:

```
backends/desktop/        // host-agnostic Backend interface + options
backends/desktop/gio/    // Gio adapter (deferred; see §4.2 follow-up)
canvas/toolbar.go        // NavigationToolbar abstraction (this commit)
```

The `Backend` and `NavigationToolbar` interfaces are deliberately
toolkit-agnostic so a future Qt, GTK, or WebGPU host can drop in
alongside `gio/` without touching `canvas/` consumers.

## Consequences

**Positive.**

- One canonical desktop backend instead of waiting on toolkit comparisons
  every refactor.
- Pure-Go build matrix stays intact; WASM and `purego` builds continue
  to work because Gio is an opt-in import behind `backends/desktop/gio`.
- `app.FrameEvent` is the natural call site for §4.4's draw-idle
  coalescing — we get the redraw scheduler for free instead of fighting
  a widget invalidation queue.
- Multi-figure managers and embedded figures are both supported by the
  same `Backend` constructor.

**Negative / trade-offs.**

- Toolbar visuals are bespoke. We pay an up-front cost to draw / lay out
  the home / pan / zoom / save buttons rather than inheriting a native
  toolbar. Matplotlib pays the same cost (`NavigationToolbar2` is
  hand-drawn in every backend), so this is alignment, not regression.
- Contributors unfamiliar with immediate-mode UI need a short on-ramp.
  We will keep the Gio surface area small — one window per figure, one
  event-pump goroutine, one frame producer — and document it inline.
- Gio's release cadence is independent of Go's; pin a known-good
  `gioui.org` version in `go.mod` and update it deliberately.

**Open questions deferred to the implementation PR.**

- Whether high-DPI scaling lives in the Gio adapter or in the AGG
  renderer's DPI knob.
- Whether the toolbar is painted by Gio widgets or by the AGG renderer
  itself. The `NavigationToolbar` interface accommodates either.
- Whether `pyplot.Show()` blocks like CPython's `plt.show()` or returns
  a handle. Matches `canvas.FigureManager.Show()` either way.

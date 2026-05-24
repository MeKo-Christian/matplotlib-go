# Interactive Backend Contracts

Interactive backends expose the same artist tree and event dispatcher as
headless rendering. Embedders should program against the shared `canvas`
interfaces first, then feature-detect optional capabilities.

## Common Surface

Every manager returns a `canvas.FigureCanvas` from `Canvas()`:

- `Figure()` returns the live `*core.Figure` / `*canvas.Figure`.
- `Draw()` repaints immediately and emits `canvas.EventDraw`.
- `Resize(width, height)` updates figure pixels and emits
  `canvas.EventResize`.
- `Connect(eventType, handler)` and `Disconnect(id)` manage dispatcher
  callbacks.
- `Close()` emits `canvas.EventClose`.

Interactive canvases that can schedule event-loop redraws also implement
`canvas.DrawIdleCanvas`. `DrawIdle()` coalesces repeated requests where the
backend has an event loop; headless canvases execute the pending draw
synchronously.

Backends with retained raster buffers may implement `canvas.BlitCanvas`:

- `CopyFromBBox(bbox)` captures a renderer buffer region.
- `RestoreRegion(region, bbox, offset)` restores captured pixels.
- `Blit(bbox)` publishes the damaged rectangle without a full figure draw.

WebAgg implements `BlitCanvas` when its renderer supports buffer regions. Gio
currently uses full-frame redraws; keeping desktop blitting out of the contract
avoids promising retained-buffer behavior before the animation phase hardens it.

## Event Payloads

Backends normalize GUI input into `canvas.Event`:

| Matplotlib event | Canvas event | Notes |
| --- | --- | --- |
| `draw_event` | `canvas.EventDraw` | Carries width and height. |
| `resize_event` | `canvas.EventResize` | Carries width and height. |
| `close_event` | `canvas.EventClose` | Emitted once per close path. |
| `button_press_event` | `canvas.EventMousePress` | Emits before `canvas.EventPick`; WebAgg preserves `DoubleClick`. |
| `button_release_event` | `canvas.EventMouseRelease` | Does not emit pick. |
| `motion_notify_event` | `canvas.EventMouseMove` | Used by navigation and axes hover tracking. |
| `figure_enter_event` | `canvas.EventFigureEnter` | WebAgg and Gio preserve position. |
| `figure_leave_event` | `canvas.EventFigureLeave` | Also clears axes hover state. |
| `axes_enter_event` | `canvas.EventAxesEnter` | Synthesized by `canvas.AxesHoverTracker`. |
| `axes_leave_event` | `canvas.EventAxesLeave` | Synthesized when the resolved axes changes or figure leaves. |
| `scroll_event` | `canvas.EventScroll` | `DeltaY` carries upstream scroll step for WebAgg. |
| `key_press_event` | `canvas.EventKeyPress` | Modifier prefixes are stripped from `Key` and carried in `Modifiers`. |
| `key_release_event` | `canvas.EventKeyRelease` | Same key normalization as press events. |
| `pick_event` | `canvas.EventPick` | Carries the original mouse payload plus selected artist metadata. |

Use `canvas.ResolveEventTarget(fig, position)` when synthesizing events outside
a backend. It returns the topmost axes under a display-pixel position and the
corresponding data coordinates when inversion succeeds.

## Backend Notes

WebAgg:

- Handles upstream-compatible JSON input names such as `button_press`,
  `motion_notify`, `scroll`, `key_press`, and `dblclick`.
- Broadcasts `navigate_mode`, `history_buttons`, `rubberband`, and PNG frame
  updates to connected clients.
- Supports `DrawIdle()` coalescing, stale-artist redraw scheduling, server-side
  save hooks, and blit frame broadcasts.

Gio:

- Maps Gio pointer and key events into the same canvas event payloads.
- Exposes `canvas.DrawIdleCanvas`; invalidation is delegated to the Gio window.
- Uses full-frame redraws for now. Desktop retained-buffer blitting is deferred
  until animation work needs it.

Headless:

- Provides the same `FigureCanvas`, `DrawIdleCanvas`, dispatcher, home/redraw,
  and save tool surface without GUI input.
- Is useful for tests and embedders that want event registration without a
  native or browser event loop.

## Switchable Embedding

`examples/embed/switchable` builds one figure and wires event callbacks through
`canvas.FigureCanvas`, then chooses `headless`, `gio`, or `webagg` at runtime:

```bash
go run ./examples/embed/switchable --backend=headless
go run ./examples/embed/switchable --backend=gio
go run ./examples/embed/switchable --backend=webagg --addr=:8080
```

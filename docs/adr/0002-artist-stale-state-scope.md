# ADR 0002: Artist Stale State Scope

- **Status:** Accepted
- **Date:** 2026-05-24
- **Topic:** Scope of artist stale-state propagation
- **Related:** `core/rasterization.go`,
  `core/lifecycle.go`, `transform/node.go`

## Context

The shared artist metadata mirrors important parts of
Matplotlib's `Artist` base class: visibility, alpha, labels, clipping, and
per-artist transforms. Those metadata setters now mark the local artist stale
through `ArtistRasterization.Stale()`.

Matplotlib also propagates stale state from artists to parents, eventually
marking the owning axes and figure dirty. This matters for interactive draw
scheduling and cached layout. The current Go architecture does not yet have a
single `Figure.stale` / `Axes.stale` lifecycle path:

- `Figure` and `Axes` do not expose stale state and do not consume artist stale
  flags during draw scheduling.
- `ArtistLifecycle` already exists as an opt-in helper for callback-based stale
  behavior, but most concrete artists do not embed it.
- `transform.TransformNode` already handles transform graph invalidation
  separately from artist metadata.
- Static rendering remains immediate: `DrawFigure` traverses the figure when
  asked rather than deciding whether to redraw from stale state.

Adding parent propagation now would require owner references or callback
registration for every figure-level and axes-level artist insertion path, plus
clear rules for when draw completion clears the stale tree. A partial
implementation would be easy to observe and hard to make Matplotlib-compatible.

## Decision

For v1.0, artist metadata stale state is intentionally **local artist state
only**.

`ArtistRasterization` setters mark the artist itself stale so inspection,
tests, and future interactive code can observe that metadata changed. They do
not automatically mark parent `Axes` or `Figure` values stale because those
parents do not yet own a coherent stale lifecycle.

`ArtistLifecycle` remains the opt-in callback mechanism for artists or helper
objects that need explicit stale notifications before a broader parent
lifecycle exists. `transform.TransformNode` remains the separate invalidation
mechanism for transform-cache dependencies.

## Consequences

**Positive.**

- The static rendering path stays simple and deterministic.
- Shared metadata setters have a consistent local observable effect.
- The project avoids a half-implemented parent dirty tree that could conflict
  with future interactive draw-idle scheduling.
- Future work can wire parent propagation deliberately once `Axes` and `Figure`
  have explicit stale semantics.

**Negative / trade-offs.**

- Setting artist metadata does not currently notify a parent figure manager or
  schedule an interactive redraw by itself.
- Code that needs callbacks must use `ArtistLifecycle` or a higher-level
  interactive manager rather than relying on `ArtistRasterization`.
- This is not full Matplotlib stale propagation parity yet.

## Follow-up

A future interactive lifecycle phase should define:

- `Figure.Stale()` / `Axes.Stale()` semantics.
- Owner registration when artists are added, removed, or moved between axes.
- Whether draw completion clears stale state automatically.
- How stale propagation interacts with `canvas.DrawIdle()` and transform graph
  invalidation.

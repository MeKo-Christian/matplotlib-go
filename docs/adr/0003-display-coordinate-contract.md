# ADR 0003: Display-coordinate contract and backend y-inversion ownership

## Status

Accepted (Phase 12.4G).

## Context

Matplotlib computes all _signed_ display-space geometry — connection-style
curvature (`arc3`), arrow-head normals, text rotation, annotation offsets — in a
**y-up display space** (origin at the bottom-left, +Y increases upward). Its Agg
backend converts that display space to the y-down device buffer only at
rasterization, by appending a flip affine to the path transform
(`src/_backend_agg.h::draw_path`: `trans *= scaling(1,-1); trans *=
translation(0,height)`). `backend_bases.RendererBase.flipy()` returns `True`,
which signals that _positions_ flip but glyph and image _bitmaps_ stay upright.

Historically this port baked the y-inversion into the data→pixel transform
(`transform.NewDisplayRectTransform` plus the `1 - frac.Y` flip in
`Axes.layout`), so core/renderer space was **y-down**. Matplotlib's
verbatim-ported signed-geometry formulas then produced mirrored output (e.g. an
`arc3` annotation arrow curved to the opposite side). Examples masked this with
sign workarounds (e.g. `OffsetY: pt(40)` against Python's `xytext=(72,-40)`).

## Decision

Core display space is **y-up with a bottom-left origin**, identical to
Matplotlib. Concretely:

- The data→display transform chain produces y-up display pixels: figure
  fractions map directly to pixels (no `1 - frac` flip), and
  `NewDisplayRectTransform` maps fraction `(0,1)` to `(dst.Min.Y, dst.Max.Y)`.
- **Every backend owns the device y-inversion at its own rasterization
  boundary.** A backend receives y-up display coordinates and converts to its
  native device frame:
  - Raster backends (AGG, gobasic) and SVG (native y-down, top-left) apply
    `y_device = H - y_display` to path/marker/clip/image **placement** while
    drawing glyph and image **bitmaps upright** (the `flipy()==True` rule).
  - PDF, PostScript, and PGF are natively y-up (bottom-left); they consume
    display coordinates directly. Any previous flip directive they emitted to
    compensate for the old y-down core is removed (becomes identity).
- Signed display-space formulas (`connectionArc3Path`, `normalForVector`, text
  rotation, annotation offsets) stay verbatim Matplotlib and are correct without
  per-site sign flips, because their inputs are now y-up.

## Consequences

- **Net-neutral invariant.** The flip only _moves_ from the transform layer to
  the backend. Ordinary (unsigned) content must rasterize byte-identically; only
  signed display geometry changes. This is the regression guard: golden diffs on
  unsigned content indicate a y-down assumption that still needs converting.
- Core helpers that position artists directly in display pixels (titles, ticks,
  spines, legends, colorbars, anchored/offset boxes) must express "up" as +Y and
  "top edge" as `rect.Max.Y`.
- Example sources become 1:1 ports of their Matplotlib originals; no per-case
  sign flips or alignment workarounds.
- The single conversion boundary per backend is the documented, source-backed
  place where the device inversion lives — no scattered ad-hoc negations.

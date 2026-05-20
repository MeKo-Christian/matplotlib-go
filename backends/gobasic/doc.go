// Package gobasic provides a pure Go renderer backend using golang.org/x/image/vector.
//
// This backend uses image.RGBA as the drawing surface and vector.Rasterizer for
// path filling and stroking. It is designed to be deterministic and work without
// CGO dependencies.
//
// GoBasic is a correctness fallback, not a pixel-identical Matplotlib renderer.
// Use AGG when exact visual parity, higher-quality rasterization, native image
// transforms, native hatching, or collection-specialized drawing is required.
//
// The GoBasic renderer supports:
//   - Fill and stroke operations, including alpha, line joins, line caps, and
//     dashed strokes through renderer-neutral stroke geometry.
//   - Rectangular and path clipping with Save/Restore state isolation.
//   - Basic text drawing, rotated text, vertical text, text measurement, and
//     text-to-path conversion through pure-Go font paths.
//   - Renderer-neutral path effects for strokes, shadows, patch effects,
//     ticked strokes, and text-as-path effects.
//   - DPI-aware text sizing via render.DPIAware.
//   - Raster image drawing and PNG export via image/png.
//   - Renderer-neutral fallbacks for marker, path-collection, quad-mesh, and
//     hatch rendering.
//
// Known fidelity limitations:
//   - Text metrics are approximate and do not expose font-wide metric or ink
//     bounds interfaces.
//   - Path rasterization uses the vector rasterizer's antialiasing behavior;
//     per-path AntialiasOn/AntialiasOff switches are not fidelity controls.
//   - Image transforms are not native; rotated images fall back to axis-aligned
//     drawing in core.
//   - Filter path effects do not use offscreen blur surfaces; unsupported
//     filters deterministically repaint the requested effect pass.
//   - Hatches and collection batches are handled by core fallbacks rather than
//     by backend-native batched render paths.
//   - Gouraud triangle shading, vector output, SVG export, GPU acceleration,
//     threading, and Matplotlib-level subpixel fidelity are unsupported.
//
// This backend is the default pure-Go renderer used by default.
package gobasic

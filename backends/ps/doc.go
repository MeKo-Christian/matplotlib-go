// Package ps implements a deterministic Level-2 PostScript/EPS backend.
//
// The backend emits vector paths, clips, background fills, hatches, text,
// RGBA raster images, transformed raster images, marker batches, and
// path-collection batches, and mixed raster/vector artist groups directly to
// Level-2 PostScript. It is reachable through core.SaveFig and registry
// extension dispatch for .ps and .eps files.
//
// Text defaults to render.PSFontPolicyPath, which converts resolved glyphs to
// filled outlines through the shared font manager. This mirrors the PDF
// backend's deterministic default where feasible without adding a separate
// Type 3/Type 42 font embedding pipeline. render.PSFontPolicyBase14 emits
// direct Helvetica text for simple/searchable output, but non-ASCII runes are
// replaced and arbitrary font keys are not embedded.
//
// Level-2 PostScript has no PDF-style alpha graphics state. Fully transparent
// fills, strokes, hatches, and text are skipped; partially transparent vector
// artists are emitted opaque using their RGB components. RGBA images, including
// mixed-mode rasterized artist groups, are pre-composited over white before
// encoding. Raster images are written as deterministic RGB colorimage
// procedures and repeated identical payloads are reused; JPEG passthrough is
// intentionally not exposed because PS colorimage consumes decoded sample data
// rather than PDF-style image XObjects with a DCTDecode filter.
package ps

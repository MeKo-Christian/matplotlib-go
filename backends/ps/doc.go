// Package ps implements a deterministic Level-2 PostScript/EPS backend.
//
// The backend is intentionally incremental: it emits vector paths, clips,
// background fills, hatches, basic text, RGBA raster images, transformed raster
// images, marker batches, and path-collection batches directly to PostScript.
// It is reachable through core.SaveFig and registry extension dispatch for .ps
// and .eps files. Font embedding and image resource reuse support are left to
// later 1.2 work so they can share semantics with the PDF backend instead of
// growing a separate policy surface.
package ps

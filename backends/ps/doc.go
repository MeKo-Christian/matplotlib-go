// Package ps implements a deterministic Level-2 PostScript/EPS backend.
//
// The backend is intentionally incremental: it emits vector paths, clips,
// background fills, hatches, basic text, RGBA raster images, and transformed
// raster images directly to PostScript, and is reachable through core.SaveFig
// and registry extension dispatch for .ps and .eps files. Font embedding and
// resource reuse support are left to later 1.2 work so they can share semantics
// with the PDF backend instead of growing a separate policy surface.
package ps

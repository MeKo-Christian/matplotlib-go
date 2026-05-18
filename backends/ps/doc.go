// Package ps implements a deterministic Level-2 PostScript/EPS backend.
//
// The backend is intentionally small in this first slice: it emits vector
// paths, clips, background fills, and basic text directly to PostScript, and is
// reachable through core.SaveFig and registry extension dispatch for .ps and
// .eps files. Image, hatch, and font embedding support are left to later 1.2
// work so they can share semantics with the PDF backend instead of growing a
// separate policy surface.
package ps

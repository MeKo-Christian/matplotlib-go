// Package render defines the backend-agnostic renderer contract that every
// drawing backend implements.
//
// The central type is [Renderer]: a small set of drawing verbs (paths, text,
// images, markers) that the artist tree in package core calls during a draw
// pass. Backends such as AGG, GoBasic, SVG, PDF, PostScript, PGF, and Skia
// each provide a Renderer implementation; optional capability interfaces
// (text drawing, transformed images, DPI-aware metrics, format export) let a
// backend advertise features beyond the base contract.
//
// Colors passed through this package use normalized sRGBA component values in
// the range [0,1] unless a specific backend documents a different contract.
// [Color] carries those components; [Color.Premultiply] and
// [Color.ToPremultipliedRGBA] convert to premultiplied forms for compositing.
//
// [NullRenderer] is a no-op Renderer that validates the push/pop balance of a
// draw pass without producing output; it is useful in tests and for measuring
// artist traversal without a real backend.
//
// Save options ([SaveOption], [SaveOptions]) describe format-specific export
// settings; they are validated against the chosen file extension so that
// unsupported combinations fail explicitly rather than being silently ignored.
package render

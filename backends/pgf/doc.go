// Package pgf implements a deterministic generator-only PGF/TikZ backend.
//
// The backend writes standalone pgfpicture snippets suitable for inclusion in
// LaTeX documents. CI keeps PGF as pure generator smoke output; local TeX
// compilation is an optional developer check because TeX installations are too
// environment-specific for the default test path.
//
// Draw-time text layout uses deterministic approximations for width, ascent,
// and descent so figures can be generated without invoking TeX. The exact
// TeX/font metrics are delegated to LaTeX when the snippet is compiled.
// This backend intentionally does not implement TeX metric extraction or font
// subsetting.
//
// Raster images and mixed raster/vector artist groups are emitted as
// self-contained PGF pixel rectangles so the output does not depend on sidecar
// files. That is deterministic and portable, but intentionally not compact for
// dense images.
//
// Filter path effects use the renderer-neutral fallback and repaint the
// requested effect pass directly. This backend does not expose native filtered
// path-effect support or offscreen filter support because PGF generation has no
// backend-local blur/soft-mask primitive that is portable across TeX drivers.
package pgf

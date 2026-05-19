// Package pgf implements a deterministic generator-only PGF/TikZ backend.
//
// The backend writes standalone pgfpicture snippets suitable for inclusion in
// LaTeX documents. CI keeps PGF as pure generator smoke output; local TeX
// compilation is an optional developer check because TeX installations are too
// environment-specific for the default test path.
//
// Raster images are emitted as self-contained PGF pixel rectangles so the
// output does not depend on sidecar files. That is deterministic and portable,
// but intentionally not compact for dense images.
package pgf

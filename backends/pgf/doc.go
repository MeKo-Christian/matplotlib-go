// Package pgf implements a deterministic generator-only PGF/TikZ backend.
//
// The backend writes standalone pgfpicture snippets suitable for inclusion in
// LaTeX documents. It intentionally does not invoke lualatex or any other TeX
// engine; external verification is left to a later policy decision in
// PLAN.md 1.2.
package pgf

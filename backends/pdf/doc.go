// Package pdf implements a native PDF renderer backend for matplotlib-go.
//
// The backend emits a single-page PDF document with deterministic object
// numbering and trailer offsets so byte-for-byte reproducibility is the
// default. It covers the core renderer contract: paths with stroke/fill/clip,
// raster images, mixed raster/vector artist groups, text-as-path output,
// optional embedded Type 0 / CIDFontType2 text output, native hatch tiling
// patterns, renderer-neutral pattern fills, axial/radial gradient fills, and
// mixed raster/vector output.
//
// Path effects are emitted through the renderer-neutral replay pipeline. Normal
// stroke/fill effects stay vector, and identity path-effect filters are captured
// as transparency-group Form XObjects. Native blurred path-effect filters
// intentionally
// use the mixed raster/vector fallback: PDF has no standard Gaussian-blur graphics
// operator, so claiming native vector support would be misleading and
// viewer-dependent.
//
// PDF-specific output options are carried by render.PDFOptions and can be
// passed through core.SavePDF, core.SaveFig, or the backends.Registry save
// dispatch. The default writes an empty /Info dictionary and uses text-as-path
// output for typography; render.WithPDFFontPolicy(render.PDFFontPolicyEmbed)
// emits real PDF text with embedded TrueType font programs.
package pdf

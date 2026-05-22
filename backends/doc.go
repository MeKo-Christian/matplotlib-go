// Package backends provides the renderer registry and backend-selection logic
// that connect the artist tree to a concrete drawing engine.
//
// Each backend (AGG, GoBasic, SVG, PDF, PostScript, PGF, Skia) registers
// itself in its init function via [Register], advertising a [Backend] name, a
// [Factory] that builds a [render.Renderer] from a [Config], and the set of
// [Capability] values it supports. Importing the convenience package
// backends/all registers every built-in backend for side effects.
//
// Callers normally do not name a backend directly. [GetBestBackend] picks the
// best available backend that satisfies a list of required capabilities, and
// [NewRendererFromEnv] honors the MATPLOTLIB_BACKEND environment variable
// while falling back to that selection. [Create] builds a renderer from an
// explicit [Backend]; [SimpleConfig] is a shorthand for the common
// width/height/background [Config].
//
// The Save* helpers ([SavePNG], [SaveSVG], [SavePDF], [SavePS], [SavePGF])
// dispatch a finished renderer to the matching on-disk format. Capability
// queries ([HasCapability], [SupportsRendererCapability],
// [VerifyRendererCapabilities]) let callers check feature support before
// relying on it.
package backends

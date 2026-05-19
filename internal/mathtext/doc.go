// Package mathtext parses and lays out Matplotlib-style MathText expressions
// for renderer-independent drawing.
//
// The package intentionally exposes a small API surface while it remains
// internal:
//   - Normalize and NormalizeDisplay provide Unicode fallback text.
//   - SplitDisplaySegments separates plain text from inline $...$ math.
//   - LayoutMathText and LayoutDisplay return flattened text runs and rule
//     rectangles that AGG, SVG, PDF, and other backends can consume.
//   - Cache and CacheConfig provide optional parsed/layout reuse. Layout cache
//     entries must be isolated by MeasurementKey because renderer text metrics
//     and font resolution affect final geometry.
//
// Promotion target: keep this internal API stable through the Phase 3 backend
// work, document any remaining cache/storage decisions, and either promote it
// to a top-level package or split it into its own module by 2026-07-31.
package mathtext

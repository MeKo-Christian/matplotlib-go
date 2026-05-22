// Package agg implements the AGG backend for matplotlib-go.
//
// Phase 8.1D task (AGG image/effects/buffer parity) status
// ---------------------------------------------------------
//
// This package is treated as the AGG reference backend and should track
// parity against:
//
//	third_party/matplotlib/lib/matplotlib/backends/backend_agg.py
//	third_party/matplotlib/src/_backend_agg.h
//
// Native coverage currently implemented
// - Path and marker drawing, including clip rect/path stacking
// - Linear/radial gradient fills and tiled pattern fills
// - Path collections, quad mesh, and Gouraud triangle rasterization
// - Image drawing and affine image drawing
// - Text raster and path rendering, text measurement, font metrics
// - PNG export
// - Direct RGBA buffer access through GetImage
// - `copy_from_bbox` / `restore_region` equivalent (`CopyFromBBox` / `RestoreRegion`)
// - `start_filter` / `stop_filter` equivalent (`StartFilter` / `StopFilter`)
//
// Intentionally unsupported for now
//   - ARGB string export helpers such as Matplotlib's `tostring_argb`.
//
// These intentional gaps should be revisited as part of ongoing 8.1 sub-tasks.
package agg

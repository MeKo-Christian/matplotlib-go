// Package svg implements a native SVG renderer backend for matplotlib-go.
//
// The backend emits vector paths, text, clip paths, transformed images, marker
// batches, path collections, hatch fills, and mixed raster/vector artist groups
// directly as SVG. Gouraud-shaded triangle meshes are intentionally not native
// yet; core code keeps using the renderer-neutral raster fallback for that case
// until an SVG gradient strategy is worth the added complexity.
//
// SVG-specific output options are carried by render.SVGOptions and can be
// passed through core.SaveSVG, core.SaveFig, or backends.Registry save dispatch.
// The default keeps native SVG text and writes an empty metadata block for
// deterministic output. Use render.WithSVGFontPolicy(render.SVGFontPolicyPath)
// for text-as-path output, render.WithSVGMetadata for sorted metadata entries,
// and render.WithSVGHashSalt to derive def IDs from SHA-256(salt+content)
// instead of registration-order counters.
package svg

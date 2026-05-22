// Package transform defines the coordinate-transform interfaces and the scale
// implementations that map data values to display space.
//
// A [T] is a 2D point transform; a [Separable] transform acts on the x and y
// axes independently and is the common case for Cartesian plots.
// [NewLinearAxis] builds a one-dimensional affine mapping, [NewSeparable] and
// [Blend] combine per-axis transforms, and [ChainSeparable] composes them.
// [NewRectTransform] and [NewUnitRectTransform] map one rectangle onto
// another — the basis for the data-to-axes-to-display pipeline.
//
// Non-linear axes are modeled as [Scale] values. Built-in scales (linear,
// log, symlog, asinh, logit, function) are created through [NewScale] using
// the [ScaleOption] functional options, and custom scales can be added with
// [RegisterScale]. [NewScaleTransform] adapts a pair of scales into a
// Separable transform.
//
// [TransformNode] and [CachedTransform] provide dependency-tracked caching so
// that an expensive transform is rebuilt only when one of its inputs changes.
package transform

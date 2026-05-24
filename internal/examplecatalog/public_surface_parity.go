package examplecatalog

import (
	"strconv"
	"strings"
)

// PublicSurfaceRow records one upstream Matplotlib public API or registry item.
type PublicSurfaceRow struct {
	ID     string `json:"id"`
	Module string `json:"module"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
}

// PublicSurfaceParityStatus records how a public upstream surface maps to this
// Go port.
type PublicSurfaceParityStatus string

const (
	PublicSurfaceDirectEquivalent    PublicSurfaceParityStatus = "direct-equivalent"
	PublicSurfaceIdiomaticEquivalent PublicSurfaceParityStatus = "idiomatic-equivalent"
	PublicSurfacePartial             PublicSurfaceParityStatus = "partial"
	PublicSurfaceNotStarted          PublicSurfaceParityStatus = "not-started"
	PublicSurfaceIntentionalOmission PublicSurfaceParityStatus = "intentional-omission"
)

// PublicSurfaceParity maps a single upstream API or registry row to local
// implementation and coverage context.
type PublicSurfaceParity struct {
	ID                string
	UpstreamID        string
	FeatureCoverageID string
	Status            PublicSurfaceParityStatus
	GoFiles           []string
	CatalogIDs        []string
	ExampleIDs        []string
	Note              string
}

type publicSurfaceParityRule struct {
	idPrefix          string
	module            string
	kind              string
	name              string
	featureCoverageID string
	status            PublicSurfaceParityStatus
	goFiles           []string
	catalogIDs        []string
	exampleIDs        []string
	note              string
}

var publicSurfaceParityOverrides = []PublicSurfaceParity{
	{
		ID:                "artist-class",
		UpstreamID:        "artist.py:class:Artist",
		FeatureCoverageID: "artist",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/artist.go", "core/lifecycle.go", "core/rasterization.go"},
		CatalogIDs:        []string{"artist_metadata", "basic_line", "patch_showcase", "mixed_raster_vector"},
		Note:              "Go has an Artist interface plus shared static rendering metadata for labels, visibility, alpha, clipping, transforms, in-layout, and local stale state. Matplotlib's dynamic property, callback, getp/setp, and parent-stale lifecycle surface remains intentionally partial.",
	},
	{
		ID:                "line2d-class",
		UpstreamID:        "lines.py:class:Line2D",
		FeatureCoverageID: "lines",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/line.go", "core/plot.go", "core/scatter.go"},
		CatalogIDs:        []string{"basic_line", "dashes", "joins_caps", "line2d_semantics", "line2d_markers", "scatter_marker_types"},
		ExampleIDs:        []string{"basic_line", "dashes"},
		Note:              "Static Line2D rendering covers strokes, dashes, joins, caps, mutable data, invalid-point breaks, gapcolor, markevery, integrated markers, legends, marker color sentinels, and half-fill behavior. The remaining partial status is the broad Python setter/getter alias surface and performance caches.",
	},
	{
		ID:                "star-marker",
		UpstreamID:        "markers.py:registry:marker:*",
		FeatureCoverageID: "lines",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"core/scatter.go"},
		CatalogIDs:        []string{"scatter_marker_types"},
		Note:              "Built-in marker registry parity includes the star marker and is covered by the marker grid fixture.",
	},
	{
		ID:                "ticker-locator-base",
		UpstreamID:        "ticker.py:class:Locator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go uses a small Locator interface instead of Matplotlib's mutable TickHelper/Locator inheritance surface.",
	},
	{
		ID:                "ticker-auto-locator",
		UpstreamID:        "ticker.py:class:AutoLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go AutoLocator delegates to MaxNLocator with Matplotlib-style default steps. Remaining 12.2B scope is exact upstream autolimit and very small/large span behavior.",
	},
	{
		ID:                "ticker-auto-minor-locator",
		UpstreamID:        "ticker.py:class:AutoMinorLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go MinorLinearLocator covers the visible AutoMinorLocator subdivision path. Remaining 12.2B scope is exact rcParam-driven auto subdivision selection.",
	},
	{
		ID:                "ticker-asinh-locator",
		UpstreamID:        "ticker.py:class:AsinhLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go AsinhLocator is installed for asinh scales and covers symmetric linear/log-like tick placement plus minor subs. Remaining 12.2B scope is exact upstream rounding preferences and catalog fixtures.",
	},
	{
		ID:                "ticker-fixed-locator",
		UpstreamID:        "ticker.py:class:FixedLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface", "units_categories"},
		Note:              "Go FixedLocator stores explicit tick locations and supports max-count subsampling around zero.",
	},
	{
		ID:                "ticker-index-locator",
		UpstreamID:        "ticker.py:class:IndexLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go IndexLocator covers base-plus-offset indexed tick placement.",
	},
	{
		ID:                "ticker-linear-locator",
		UpstreamID:        "ticker.py:class:LinearLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go LinearLocator covers exact-count linear ticks and preset expansion. Remaining 12.2B scope is Python mutable preset API compatibility.",
	},
	{
		ID:                "ticker-max-n-locator",
		UpstreamID:        "ticker.py:class:MaxNLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go MaxNLocator covers nbins, custom steps, integer relaxation, symmetric ranges, pruning, and degenerate ranges. Remaining 12.2B scope is very small/large spans and exact upstream offset/autolimit edge behavior.",
	},
	{
		ID:                "ticker-log-locator",
		UpstreamID:        "ticker.py:class:LogLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface", "lognorm_imshow"},
		Note:              "Go LogLocator covers base changes, explicit subs, auto/all subs modes, dense automatic minor suppression, and safe non-positive/inverted ranges. Remaining 12.2B scope is exact stride/numticks matching and catalog fixtures.",
	},
	{
		ID:                "ticker-logit-locator",
		UpstreamID:        "ticker.py:class:LogitLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go LogitLocator is installed for logit scales and covers major/minor probability ticks. Remaining 12.2B scope is exact upstream minor pruning and catalog fixtures.",
	},
	{
		ID:                "ticker-multiple-locator",
		UpstreamID:        "ticker.py:class:MultipleLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go MultipleLocator covers regular ticks with an optional offset.",
	},
	{
		ID:                "ticker-null-locator",
		UpstreamID:        "ticker.py:class:NullLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go NullLocator returns no tick locations.",
	},
	{
		ID:                "ticker-symmetrical-log-locator",
		UpstreamID:        "ticker.py:class:SymmetricalLogLocator",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go SymLogLocator is installed for symlog scales and covers negative, linear-threshold, and positive log regions. Remaining 12.2B scope is exact upstream minor stride behavior and catalog fixtures.",
	},
	{
		ID:                "ticker-log-formatter",
		UpstreamID:        "ticker.py:class:LogFormatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface", "lognorm_imshow"},
		Note:              "Go LogFormatter covers base-10 power labels, label-only-base behavior, and minor-label sparsity thresholds. Remaining 12.2C scope is formatter catalog fixtures.",
	},
	{
		ID:                "ticker-log-formatter-exponent",
		UpstreamID:        "ticker.py:class:LogFormatterExponent",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface", "lognorm_imshow"},
		Note:              "Go LogFormatterExponent covers exponent labels, label-only-base behavior, and minor-label sparsity thresholds. Remaining 12.2C scope is formatter catalog fixtures.",
	},
	{
		ID:                "ticker-log-formatter-mathtext",
		UpstreamID:        "ticker.py:class:LogFormatterMathtext",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface", "lognorm_imshow"},
		Note:              "Go LogFormatterMathText covers MathText powers, scientific notation, label-only-base behavior, and minor-label sparsity thresholds. Remaining 12.2C scope is formatter catalog fixtures.",
	},
	{
		ID:                "ticker-log-formatter-sci-notation",
		UpstreamID:        "ticker.py:class:LogFormatterSciNotation",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface", "lognorm_imshow"},
		Note:              "Go represents LogFormatterSciNotation as LogFormatterMathText with SciNotation enabled, including sparse minor-label behavior. Remaining 12.2C scope is formatter catalog fixtures.",
	},
	{
		ID:                "ticker-formatter-base",
		UpstreamID:        "ticker.py:class:Formatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go uses a Formatter interface plus optional IndexedFormatter for tick-index-aware labels instead of Matplotlib's mutable Formatter base class.",
	},
	{
		ID:                "ticker-logit-formatter",
		UpstreamID:        "ticker.py:class:LogitFormatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go LogitFormatter is installed for logit scales and covers major labels, half labels, minor suppression, and one-minus labels. Remaining 12.2C scope is exact overline MathText rendering and formatter catalog fixtures.",
	},
	{
		ID:                "ticker-scalar-formatter",
		UpstreamID:        "ticker.py:class:ScalarFormatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go ScalarFormatter covers fixed-minus, step-aware precision, scientific suppression, inclusive power limits, and MathText-style scientific labels in tick rendering. Remaining 12.2C scope is offset text and locale/no-locale policy.",
	},
	{
		ID:                "ticker-eng-formatter",
		UpstreamID:        "ticker.py:class:EngFormatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go EngFormatter covers SI prefix scaling, fixed-minus, unicode micro opt-in, rounding rollover, extreme SI prefixes, MathText-style number wrapping, and format_eng-style aliasing. Remaining 12.2C scope is exact default separator/places compatibility and offset behavior.",
	},
	{
		ID:                "ticker-percent-formatter",
		UpstreamID:        "ticker.py:class:PercentFormatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go PercentFormatter covers explicit xmax, fixed decimals, configured display-range auto decimals, symbols, and fixed-minus. Remaining 12.2C scope is exact zero-value defaults and TeX symbol escaping.",
	},
	{
		ID:                "ticker-fixed-formatter",
		UpstreamID:        "ticker.py:class:FixedFormatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface", "units_categories"},
		Note:              "Go FixedFormatter labels by tick index through IndexedFormatter. Python's runtime FixedFormatter/FixedLocator mismatch warning is intentionally omitted because Go callers wire typed locators/formatters explicitly.",
	},
	{
		ID:                "ticker-null-formatter",
		UpstreamID:        "ticker.py:class:NullFormatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go NullFormatter always returns an empty label.",
	},
	{
		ID:                "ticker-func-formatter",
		UpstreamID:        "ticker.py:class:FuncFormatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go FuncFormatter is a typed function callback over the tick value; Python's optional positional index argument is intentionally omitted.",
	},
	{
		ID:                "ticker-format-str-formatter",
		UpstreamID:        "ticker.py:class:FormatStrFormatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go FormatStrFormatter uses fmt.Sprintf-style formatting, matching the visible printf-style tick label use case.",
	},
	{
		ID:                "ticker-str-method-formatter",
		UpstreamID:        "ticker.py:class:StrMethodFormatter",
		FeatureCoverageID: "axis-ticker-scale",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go StrMethodFormatter supports the common {x:.Ng} / {x:.Nf} / {x:.Ne} subset used by tick labels; full Python format mini-language parity remains intentionally partial.",
	},
	{
		ID:                "lanczos-interpolation",
		UpstreamID:        "image.py:registry:interpolation:lanczos",
		FeatureCoverageID: "image",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"core/image.go", "backends/agg/interpolation.go"},
		CatalogIDs:        []string{"image_heatmap", "imshow_bilinear", "imshow_bicubic"},
		Note:              "AGG interpolation name resolution maps Matplotlib's Lanczos filter directly.",
	},
	{
		ID:                "transforms-transform-node",
		UpstreamID:        "transforms.py:class:TransformNode",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"transform/node.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go TransformNode provides explicit invalidation propagation and versioning instead of Matplotlib's broader parent/child invalidation graph.",
	},
	{
		ID:                "transforms-transform-base",
		UpstreamID:        "transforms.py:class:Transform",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"transform/transform.go", "transform/graph.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go uses a small transform interface with Apply/Invert behavior and concrete graph helpers rather than Matplotlib's dynamic Transform base class.",
	},
	{
		ID:                "transforms-affine-base",
		UpstreamID:        "transforms.py:class:AffineBase",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"internal/geom/geom.go", "transform/transform.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go represents affine transforms as geom.Affine plus graph adapters instead of Python's AffineBase inheritance surface.",
	},
	{
		ID:                "transforms-affine2d-base",
		UpstreamID:        "transforms.py:class:Affine2DBase",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"internal/geom/geom.go", "transform/transform.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go geom.Affine covers 2D affine application, composition, and inversion without Matplotlib's mutable matrix API.",
	},
	{
		ID:                "transforms-affine2d",
		UpstreamID:        "transforms.py:class:Affine2D",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"internal/geom/geom.go", "transform/transform.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go geom.Affine covers static affine math; Matplotlib's mutating rotate, scale, skew, and matrix convenience surface remains intentionally partial.",
	},
	{
		ID:                "transforms-identity-transform",
		UpstreamID:        "transforms.py:class:IdentityTransform",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"internal/geom/geom.go", "transform/transform.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Go geom.Identity and identity transform adapters cover identity application and inversion.",
	},
	{
		ID:                "transforms-affine-delta-transform",
		UpstreamID:        "transforms.py:class:AffineDeltaTransform",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"internal/geom/geom.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Go callers model delta transforms by using affine linear terms with zero translation; a separate wrapper class is intentionally unnecessary.",
	},
	{
		ID:                "transforms-bbox-base",
		UpstreamID:        "transforms.py:class:BboxBase",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"internal/geom/geom.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go geom.Rect provides BBox-style dimensions, containment, union, intersection, padding, expansion, anchoring, and transforms without Python's mutable base class.",
	},
	{
		ID:                "transforms-bbox",
		UpstreamID:        "transforms.py:class:Bbox",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"internal/geom/geom.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go geom.Rect covers static BBox geometry. Mutable point-array APIs, null extents, and live lockable BBox variants remain outside the current Go surface.",
	},
	{
		ID:                "transforms-bbox-transform",
		UpstreamID:        "transforms.py:class:BboxTransform",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"transform/graph.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Go NewRectTransform maps one rectangle into another.",
	},
	{
		ID:                "transforms-bbox-transform-from",
		UpstreamID:        "transforms.py:class:BboxTransformFrom",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"transform/graph.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Go NewRectTransform with a unit or caller-provided destination covers BboxTransformFrom use cases.",
	},
	{
		ID:                "transforms-bbox-transform-to",
		UpstreamID:        "transforms.py:class:BboxTransformTo",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"transform/graph.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Go NewUnitRectTransform and NewDisplayRectTransform cover BboxTransformTo use cases.",
	},
	{
		ID:                "transforms-blended-generic",
		UpstreamID:        "transforms.py:class:BlendedGenericTransform",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"transform/graph.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go Blend combines independent x and y transform components.",
	},
	{
		ID:                "transforms-blended-affine",
		UpstreamID:        "transforms.py:class:BlendedAffine2D",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"transform/graph.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go Blend covers the affine blended-transform path through separable axis transforms.",
	},
	{
		ID:                "transforms-composite-generic",
		UpstreamID:        "transforms.py:class:CompositeGenericTransform",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"transform/transform.go", "transform/graph.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go Chain and ChainSeparable compose transforms explicitly.",
	},
	{
		ID:                "transforms-composite-affine",
		UpstreamID:        "transforms.py:class:CompositeAffine2D",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"internal/geom/geom.go", "transform/transform.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Go affine composition is handled by geom.Affine.Mul and transform chaining.",
	},
	{
		ID:                "transforms-scaled-translation",
		UpstreamID:        "transforms.py:class:ScaledTranslation",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"transform/graph.go"},
		CatalogIDs:        []string{"annotation_composition"},
		Note:              "Go NewOffset applies device-space offsets after a base transform; DPI or physical-unit scaling is explicit at call sites.",
	},
	{
		ID:                "transforms-transform-wrapper",
		UpstreamID:        "transforms.py:class:TransformWrapper",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"transform/node.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Go CachedTransform and explicit transform assignment cover wrapper-style replacement with invalidation.",
	},
	{
		ID:                "transforms-transformed-path",
		UpstreamID:        "transforms.py:class:TransformedPath",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"transform/transformed_path.go", "internal/geom/geom.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go TransformedPath covers clone-safe transformed-path caching with dependency invalidation. Remaining 12.2G scope is affine/non-affine split behavior if a non-affine transform family is added.",
	},
	{
		ID:                "transforms-transformed-patch-path",
		UpstreamID:        "transforms.py:class:TransformedPatchPath",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"transform/transformed_path.go", "core/patch.go"},
		CatalogIDs:        []string{"annotation_composition"},
		Note:              "Go patch paths can use TransformedPath directly; a patch-specific subclass is intentionally omitted.",
	},
	{
		ID:                "transforms-transformed-bbox",
		UpstreamID:        "transforms.py:class:TransformedBbox",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"internal/geom/geom.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go geom.Rect.Transformed and InverseTransformed cover transformed BBox extents without a live wrapper object.",
	},
	{
		ID:                "transforms-lockable-bbox",
		UpstreamID:        "transforms.py:class:LockableBbox",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIntentionalOmission,
		GoFiles:           []string{"internal/geom/geom.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Mutable per-edge lock semantics are intentionally omitted; Go layout code uses explicit rect values.",
	},
	{
		ID:                "transforms-blended-transform-factory",
		UpstreamID:        "transforms.py:function:blended_transform_factory",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"transform/graph.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go Blend is the direct factory for blended transforms.",
	},
	{
		ID:                "transforms-composite-transform-factory",
		UpstreamID:        "transforms.py:function:composite_transform_factory",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceDirectEquivalent,
		GoFiles:           []string{"transform/transform.go", "transform/graph.go"},
		CatalogIDs:        []string{"transform_coordinates", "annotation_composition"},
		Note:              "Go Chain and ChainSeparable are the direct factories for composed transforms.",
	},
	{
		ID:                "transforms-offset-copy",
		UpstreamID:        "transforms.py:function:offset_copy",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIdiomaticEquivalent,
		GoFiles:           []string{"transform/graph.go"},
		CatalogIDs:        []string{"annotation_composition"},
		Note:              "Go NewOffset creates offset copies; physical unit conversion is explicit at call sites.",
	},
	{
		ID:                "transforms-interval-helpers",
		UpstreamID:        "transforms.py:function:interval_contains",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIntentionalOmission,
		GoFiles:           []string{"internal/geom/geom.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Standalone interval helpers are intentionally omitted; callers use explicit comparisons or Rect containment helpers.",
	},
	{
		ID:                "transforms-interval-open-helpers",
		UpstreamID:        "transforms.py:function:interval_contains_open",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIntentionalOmission,
		GoFiles:           []string{"internal/geom/geom.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Open-interval helpers are intentionally omitted; callers use explicit comparisons.",
	},
	{
		ID:                "transforms-nonsingular",
		UpstreamID:        "transforms.py:function:nonsingular",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/tick.go", "transform/scale_registry.go"},
		CatalogIDs:        []string{"axes_control_surface"},
		Note:              "Go handles nonsingular expansion locally in locator and scale code. A standalone public helper remains omitted unless shared callers need it.",
	},
	{
		ID:                "transforms-debug-constant",
		UpstreamID:        "transforms.py:constant:DEBUG",
		FeatureCoverageID: "transforms",
		Status:            PublicSurfaceIntentionalOmission,
		GoFiles:           []string{"transform/doc.go"},
		CatalogIDs:        []string{"transform_coordinates"},
		Note:              "Matplotlib's module-level DEBUG flag is intentionally omitted; Go tests and explicit errors cover diagnostics.",
	},
	{
		ID:                "pyplot-plot",
		UpstreamID:        "pyplot.py:function:plot",
		FeatureCoverageID: "pyplot-state",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"pyplot/pyplot.go", "core/plot.go"},
		CatalogIDs:        []string{"basic_line"},
		ExampleIDs:        []string{"basic_line"},
		Note:              "A stateful Go pyplot package exists for common migration paths, but it intentionally does not mirror the full Python overload surface.",
	},
	{
		ID:                "button-widget",
		UpstreamID:        "widgets.py:class:Button",
		FeatureCoverageID: "widgets-events-animation",
		Status:            PublicSurfacePartial,
		GoFiles:           []string{"core/widgets.go", "canvas/dispatcher.go"},
		Note:              "Static widget artists and event infrastructure exist; interactive widget workflows remain Phase 9C work.",
	},
	{
		ID:                "func-animation",
		UpstreamID:        "animation.py:class:FuncAnimation",
		FeatureCoverageID: "widgets-events-animation",
		Status:            PublicSurfaceNotStarted,
		GoFiles:           []string{"canvas/scheduler.go"},
		Note:              "Canvas scheduling exists, but Matplotlib-style FuncAnimation parity has not been implemented or demoed.",
	},
}

var publicSurfaceParityRules = []publicSurfaceParityRule{
	{
		idPrefix:          "artist",
		module:            "artist.py",
		featureCoverageID: "artist",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/artist.go", "core/lifecycle.go", "core/rasterization.go"},
		catalogIDs:        []string{"basic_line", "patch_showcase", "mixed_raster_vector"},
		note:              "Go keeps an interface-based Artist model; broad dynamic properties, inspection helpers, getp/setp, and callbacks remain partial.",
	},
	{
		idPrefix:          "axis",
		module:            "axis.py",
		featureCoverageID: "axis-ticker-scale",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/axis.go", "core/tick.go", "core/grid.go"},
		catalogIDs:        []string{"axes_control_surface", "units_dates", "units_categories", "skewt_basic"},
		exampleIDs:        []string{"axes_control_surface", "units_overview", "skewt_basic"},
		note:              "Axis, ticks, grid lines, mirrored axes, labels, scale defaults, and TickParams styling exist. Go intentionally keeps ticks axis-owned for v1.0; remaining partial scope is catalog fixture breadth and rare Python setter/callback surfaces.",
	},
	{
		idPrefix:          "ticker",
		module:            "ticker.py",
		featureCoverageID: "axis-ticker-scale",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/tick.go", "core/date_tick.go", "core/units.go"},
		catalogIDs:        []string{"axes_control_surface", "units_dates", "units_categories", "lognorm_imshow", "skewt_basic"},
		exampleIDs:        []string{"axes_control_surface", "units_overview", "skewt_basic"},
		note:              "Common locators and formatters plus symlog/asinh/logit/date/category families exist. Remaining 12.2B/C scope is AutoLocator/MaxNLocator edge semantics, dense log minor labels, ScalarFormatter/EngFormatter/PercentFormatter edge behavior, and catalog fixture breadth.",
	},
	{
		idPrefix:          "scale-registry-functionlog",
		module:            "scale.py",
		kind:              "registry:scale",
		name:              "functionlog",
		featureCoverageID: "axis-ticker-scale",
		status:            PublicSurfaceDirectEquivalent,
		goFiles:           []string{"transform/scale_registry.go"},
		catalogIDs:        []string{"axes_control_surface"},
		exampleIDs:        []string{"axes_control_surface"},
		note:              "The Go scale registry includes functionlog with caller-provided forward/inverse functions and log-family axis defaults.",
	},
	{
		idPrefix:          "scale",
		module:            "scale.py",
		featureCoverageID: "axis-ticker-scale",
		status:            PublicSurfacePartial,
		goFiles:           []string{"transform/scale_registry.go", "transform/transform.go", "core/axis.go"},
		catalogIDs:        []string{"axes_control_surface", "lognorm_imshow", "skewt_basic"},
		exampleIDs:        []string{"axes_control_surface", "skewt_basic"},
		note:              "Named scale construction covers linear, log, symlog, asinh, logit, function, and functionlog with scale-specific axis defaults and log-like non-positive handling. Remaining partial scope is upstream audit notes and catalog fixture breadth.",
	},
	{
		idPrefix:          "transforms",
		module:            "transforms.py",
		featureCoverageID: "transforms",
		status:            PublicSurfacePartial,
		goFiles:           []string{"transform/transform.go", "transform/graph.go", "transform/node.go", "internal/geom/geom.go"},
		catalogIDs:        []string{"transform_coordinates", "annotation_composition", "imshow_transformed"},
		exampleIDs:        []string{"annotation_composition"},
		note:              "Go has affine, scale, blended, chained, display-rect, offset, cached graph transforms, and BBox-style rect helpers. Remaining 12.2G scope is frozen snapshots, transformed-path cache helpers, anchored rect helpers, and visible path/bezier helpers.",
	},
	{
		idPrefix:          "line-style-registry",
		module:            "lines.py",
		kind:              "registry:linestyle",
		featureCoverageID: "lines",
		status:            PublicSurfaceDirectEquivalent,
		goFiles:           []string{"core/line.go"},
		catalogIDs:        []string{"dashes", "joins_caps"},
		exampleIDs:        []string{"dashes"},
		note:              "Matplotlib line-style aliases are represented by Go stroke and dash options and covered by dash fixtures.",
	},
	{
		idPrefix:          "lines",
		module:            "lines.py",
		featureCoverageID: "lines",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/line.go", "core/plot.go", "core/scatter.go"},
		catalogIDs:        []string{"basic_line", "dashes", "joins_caps", "line2d_semantics", "line2d_markers", "scatter_marker_types"},
		exampleIDs:        []string{"basic_line", "dashes"},
		note:              "Line drawing, dashes, joins, caps, mutable data, invalid data breaks, markevery, gapcolor, integrated marker drawing, and marker legends are covered; remaining partial status is Python API breadth and transformed-path performance caching.",
	},
	{
		idPrefix:          "marker-fillstyle-registry",
		module:            "markers.py",
		kind:              "registry:fillstyle",
		featureCoverageID: "lines",
		status:            PublicSurfaceDirectEquivalent,
		goFiles:           []string{"core/scatter.go"},
		catalogIDs:        []string{"scatter_marker_types", "line2d_markers"},
		note:              "Marker fill styles are modeled by MarkerStyle, including full, none, and split left/right/top/bottom rendering for Line2D and Scatter2D marker paths.",
	},
	{
		idPrefix:          "marker-registry",
		module:            "markers.py",
		kind:              "registry:marker",
		featureCoverageID: "lines",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/scatter.go"},
		catalogIDs:        []string{"scatter_marker_types", "line2d_markers"},
		note:              "Go has built-in marker constants, common string aliases, tuple markers, custom paths, mathtext/text markers, and Line2D integration. Remaining partial status is the exact Python object registry and rare alias breadth.",
	},
	{
		idPrefix:          "markers",
		module:            "markers.py",
		featureCoverageID: "lines",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/scatter.go"},
		catalogIDs:        []string{"scatter_marker_types", "line2d_markers"},
		note:              "MarkerStyle has an idiomatic Go equivalent for built-in, tuple, custom-path, mathtext, fillstyle, and half-fill workflows; the upstream mutable class surface is not cloned one-for-one.",
	},
	{
		idPrefix:          "collections",
		module:            "collections.py",
		featureCoverageID: "collections",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/collection.go", "core/mesh.go", "core/eventplot.go", "core/hexbin.go"},
		catalogIDs:        []string{"mixed_collection", "large_scatter", "quad_mesh", "gouraud_triangles", "specialty_artists"},
		exampleIDs:        []string{"specialty_artists"},
		note:              "Path, line, patch, poly, quad, event, and hexbin collections exist; specialized variants and mutable setter parity remain incomplete.",
	},
	{
		idPrefix:          "patch-style-registry",
		module:            "patches.py",
		kind:              "registry:",
		featureCoverageID: "patches",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/patch.go", "core/patch_extra.go", "core/arrow_patch.go"},
		catalogIDs:        []string{"patch_showcase"},
		note:              "Common patch, arrow, box, connection, and hatch behavior exists, but the upstream style registries need row-by-row closure.",
	},
	{
		idPrefix:          "patches",
		module:            "patches.py",
		featureCoverageID: "patches",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/patch.go", "core/patch_extra.go", "core/arrow_patch.go"},
		catalogIDs:        []string{"patch_showcase"},
		note:              "Core patch classes and extra shapes exist; full Python patch API and all specialized classes remain partial.",
	},
	{
		idPrefix:          "text",
		module:            "text.py",
		featureCoverageID: "text-annotation-legend",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/text.go", "core/mathtext.go", "core/arrow_patch.go"},
		catalogIDs:        []string{"text_labels_strict", "title_strict", "annotation_composition", "figure_labels_composition", "mathtext_inline_labels"},
		exampleIDs:        []string{"annotation_composition", "figure_labels_composition"},
		note:              "Text and annotation rendering exist, including rotation, MathText, arrows, and anchored labels; font property and coordinate model parity is partial.",
	},
	{
		idPrefix:          "legend",
		module:            "legend.py",
		featureCoverageID: "text-annotation-legend",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/legend.go"},
		catalogIDs:        []string{"multi_series_basic", "multi_series_color_cycle", "figure_labels_composition"},
		exampleIDs:        []string{"figure_labels_composition"},
		note:              "Static legend layout exists; draggable legends, handler maps, and broader proxy-artist parity remain partial.",
	},
	{
		idPrefix:          "offsetbox",
		module:            "offsetbox.py",
		featureCoverageID: "text-annotation-legend",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/anchored_text.go", "core/text.go"},
		catalogIDs:        []string{"annotation_composition", "figure_labels_composition"},
		exampleIDs:        []string{"annotation_composition", "figure_labels_composition"},
		note:              "Anchored text covers the main static offset-box use case; packing boxes, draggable boxes, offset images, and AnnotationBbox remain thin.",
	},
	{
		idPrefix:          "image-interpolation-registry",
		module:            "image.py",
		kind:              "registry:interpolation",
		featureCoverageID: "image",
		status:            PublicSurfaceDirectEquivalent,
		goFiles:           []string{"core/image.go", "backends/agg/interpolation.go"},
		catalogIDs:        []string{"image_heatmap", "imshow_bilinear", "imshow_bicubic"},
		exampleIDs:        []string{"image_heatmap"},
		note:              "The AGG backend resolves Matplotlib interpolation names to AGG image filters, including adaptive auto/antialiased handling.",
	},
	{
		idPrefix:          "image",
		module:            "image.py",
		featureCoverageID: "image",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/image.go", "core/image_api.go", "core/matrix_helpers.go"},
		catalogIDs:        []string{"image_heatmap", "imshow_clipped", "imshow_transformed", "image_alpha", "matshow_basic", "spy_marker", "spy_image", "arrays_showcase"},
		exampleIDs:        []string{"image_heatmap", "arrays_showcase"},
		note:              "imshow, matshow, spy, alpha, origin, extent, and transformed images exist; FigureImage, BboxImage, NonUniformImage, and IO helpers are incomplete.",
	},
	{
		idPrefix:          "colorbar",
		module:            "colorbar.py",
		featureCoverageID: "colorbar",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/colorbar.go", "core/scalar_mappable.go", "core/norm.go"},
		catalogIDs:        []string{"colorbar_composition", "boundarynorm_pcolormesh", "lognorm_imshow", "twoslope_norm_image", "colorbar_extensions"},
		exampleIDs:        []string{"colorbar_composition"},
		note:              "Scalar mappables and vertical colorbars exist; horizontal placement, custom ticks, boundaries, and gridspec helpers remain partial.",
	},
	{
		idPrefix:          "cm",
		module:            "cm.py",
		featureCoverageID: "colors-cm",
		status:            PublicSurfacePartial,
		goFiles:           []string{"color/colormap.go", "color/listed_colormaps.go"},
		catalogIDs:        []string{"colormap_diverging", "colormap_qualitative", "colormap_cyclic", "image_heatmap"},
		note:              "Colormap lookup, listed/segmented maps, reversals, and resampling exist; the Python ColormapRegistry surface is only partially mirrored.",
	},
	{
		idPrefix:          "colors",
		module:            "colors.py",
		featureCoverageID: "colors-cm",
		status:            PublicSurfacePartial,
		goFiles:           []string{"color/colormap.go", "color/listed_colormaps.go", "color/named_colors.go", "core/norm.go"},
		catalogIDs:        []string{"colormap_diverging", "colormap_qualitative", "colormap_cyclic", "named_colors", "lognorm_imshow", "twoslope_norm_image"},
		note:              "Named colors, colormaps, and common norms exist; advanced norms, LightSource, bivar/multivar colormaps, and conversion edge cases remain partial.",
	},
	{
		idPrefix:          "pyplot",
		module:            "pyplot.py",
		featureCoverageID: "pyplot-state",
		status:            PublicSurfacePartial,
		goFiles:           []string{"pyplot/pyplot.go", "canvas/canvas.go"},
		catalogIDs:        []string{"basic_line", "scatter_basic", "bar_basic"},
		exampleIDs:        []string{"basic_line", "scatter_basic", "bar_basic"},
		note:              "The Go pyplot package covers common stateful migration helpers but does not clone the full Python wrapper and overload surface.",
	},
	{
		idPrefix:          "backend-base",
		module:            "backend_bases.py",
		featureCoverageID: "renderer-backends",
		status:            PublicSurfaceIdiomaticEquivalent,
		goFiles:           []string{"render/render.go", "render/graphics_context.go", "render/extensions.go", "backends/registry.go", "canvas/canvas.go", "canvas/dispatcher.go"},
		catalogIDs:        []string{"basic_line", "mixed_raster_vector", "large_scatter", "clip_path_batch"},
		exampleIDs:        []string{"basic_line"},
		note:              "Renderer, canvas, events, and backend registration are split into Go packages rather than mirroring Matplotlib backend base classes directly.",
	},
	{
		idPrefix:          "backend-tool-registry",
		module:            "backend_tools.py",
		kind:              "registry:tool",
		featureCoverageID: "widgets-events-animation",
		status:            PublicSurfacePartial,
		goFiles:           []string{"canvas/tool.go", "canvas/navigation.go"},
		note:              "Navigation and tool infrastructure exists, but the full Matplotlib tool registry is not complete.",
	},
	{
		idPrefix:          "backend-tools",
		module:            "backend_tools.py",
		featureCoverageID: "widgets-events-animation",
		status:            PublicSurfacePartial,
		goFiles:           []string{"canvas/tool.go", "canvas/navigation.go", "canvas/dispatcher.go"},
		note:              "Go has an idiomatic canvas/tool split; many concrete Matplotlib tools are still missing or backend-specific.",
	},
	{
		idPrefix:          "widgets",
		module:            "widgets.py",
		featureCoverageID: "widgets-events-animation",
		status:            PublicSurfacePartial,
		goFiles:           []string{"core/widgets.go", "canvas/dispatcher.go", "canvas/picker.go"},
		note:              "Static widget artists and event dispatch exist; interactive selectors, sliders, buttons, and callbacks remain partial.",
	},
	{
		idPrefix:          "animation",
		module:            "animation.py",
		featureCoverageID: "widgets-events-animation",
		status:            PublicSurfaceNotStarted,
		goFiles:           []string{"canvas/scheduler.go"},
		note:              "Canvas scheduling exists, but Matplotlib animation classes, movie writers, and animation demos are not implemented.",
	},
}

// PublicSurfaceParityRows returns Phase 9B public-surface parity
// classifications.
func PublicSurfaceParityRows() []PublicSurfaceParity {
	out := make([]PublicSurfaceParity, len(publicSurfaceParityOverrides))
	copy(out, publicSurfaceParityOverrides)
	for i := range out {
		out[i] = clonePublicSurfaceParity(out[i])
	}
	return out
}

// PublicSurfaceParityRowsForSurface classifies an upstream public-surface
// inventory. Every returned row corresponds to one upstream row.
func PublicSurfaceParityRowsForSurface(rows []PublicSurfaceRow) []PublicSurfaceParity {
	out := make([]PublicSurfaceParity, 0, len(rows))
	for _, surface := range rows {
		row, ok := PublicSurfaceParityForRow(surface)
		if !ok {
			continue
		}
		out = append(out, row)
	}
	return out
}

// PublicSurfaceParityForRow classifies one upstream public-surface row.
func PublicSurfaceParityForRow(surface PublicSurfaceRow) (PublicSurfaceParity, bool) {
	if row, ok := LookupPublicSurfaceParityByUpstreamID(surface.ID); ok {
		return row, true
	}
	for _, rule := range publicSurfaceParityRules {
		if !rule.matches(surface) {
			continue
		}
		return rule.classification(surface), true
	}
	return PublicSurfaceParity{}, false
}

// LookupPublicSurfaceParityByUpstreamID finds a classification by upstream
// public-surface row ID.
func LookupPublicSurfaceParityByUpstreamID(upstreamID string) (PublicSurfaceParity, bool) {
	for _, row := range PublicSurfaceParityRows() {
		if row.UpstreamID == upstreamID {
			return clonePublicSurfaceParity(row), true
		}
	}
	return PublicSurfaceParity{}, false
}

func (r publicSurfaceParityRule) matches(surface PublicSurfaceRow) bool {
	if surface.Module != r.module {
		return false
	}
	if r.kind == "" {
		return r.name == "" || surface.Name == r.name
	}
	if surface.Kind != r.kind && (!strings.HasSuffix(r.kind, ":") || !strings.HasPrefix(surface.Kind, r.kind)) {
		return false
	}
	return r.name == "" || surface.Name == r.name
}

func (r publicSurfaceParityRule) classification(surface PublicSurfaceRow) PublicSurfaceParity {
	id := r.idPrefix + "-" + publicSurfaceIDSlug(surface.Kind+"-"+surface.Name)
	return PublicSurfaceParity{
		ID:                id,
		UpstreamID:        surface.ID,
		FeatureCoverageID: r.featureCoverageID,
		Status:            r.status,
		GoFiles:           append([]string(nil), r.goFiles...),
		CatalogIDs:        append([]string(nil), r.catalogIDs...),
		ExampleIDs:        append([]string(nil), r.exampleIDs...),
		Note:              r.note,
	}
}

func clonePublicSurfaceParity(row PublicSurfaceParity) PublicSurfaceParity {
	row.GoFiles = append([]string(nil), row.GoFiles...)
	row.CatalogIDs = append([]string(nil), row.CatalogIDs...)
	row.ExampleIDs = append([]string(nil), row.ExampleIDs...)
	return row
}

func publicSurfaceIDSlug(value string) string {
	original := value
	var canonical strings.Builder
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
			canonical.WriteByte('u')
			canonical.WriteRune(r + ('a' - 'A'))
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			canonical.WriteRune(r)
		default:
			canonical.WriteRune(r)
		}
	}
	value = canonical.String()
	value = strings.NewReplacer(
		" ", "space",
		":", "-",
		"_", "-",
		".", "dot",
		"-", "dash",
		">", "gt",
		"<", "lt",
		"[", "bracket",
		"]", "bracket",
		"|", "bar",
		"+", "plus",
		"*", "star",
		",", "comma",
		"(", "paren",
		")", "paren",
	).Replace(value)
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	slug := strings.Join(fields, "-")
	if slug == "" {
		slug = "item"
	}
	return slug + "-" + publicSurfaceIDChecksum(original)
}

func publicSurfaceIDChecksum(value string) string {
	var sum uint64
	for i, r := range value {
		sum += uint64(i+1) * uint64(r)
	}
	return strconv.FormatUint(sum, 36)
}

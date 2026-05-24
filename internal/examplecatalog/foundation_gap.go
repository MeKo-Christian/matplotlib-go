package examplecatalog

// GapDecision records how a Phase 9A.2 foundation API gap should be handled.
type GapDecision string

const (
	// GapDecisionImplement means the gap should be closed with public Go API or
	// core behavior changes.
	GapDecisionImplement GapDecision = "implement"
	// GapDecisionIdiomaticEquivalent means the Go port should expose equivalent
	// behavior through existing Go-style types or options instead of copying the
	// Python class shape.
	GapDecisionIdiomaticEquivalent GapDecision = "idiomatic-equivalent"
	// GapDecisionIntentionalOmission means the upstream surface is intentionally
	// not planned for this port.
	GapDecisionIntentionalOmission GapDecision = "intentional-omission"
)

// FoundationAPIGap records one Phase 9A.2 missing or thin foundational API area.
type FoundationAPIGap struct {
	ID                string
	CoverageID        string
	Title             string
	UpstreamModules   []string
	GoFiles           []string
	CurrentEquivalent string
	Gap               string
	Decision          GapDecision
	Rationale         string
}

var foundationAPIGaps = []FoundationAPIGap{
	{
		ID:              "artist-properties-callbacks",
		CoverageID:      "artist",
		Title:           "Artist properties, visibility, clipping, and callbacks",
		UpstreamModules: []string{"artist.py"},
		GoFiles:         []string{"core/artist.go", "core/lifecycle.go", "core/rasterization.go"},
		CurrentEquivalent: "Go has an Artist interface plus draw order, bounds, lifecycle propagation, " +
			"sticky edges, and rasterization helpers.",
		Gap: "Matplotlib's Artist property surface is broader: visible / alpha / clip / " +
			"transform / label setters, stale callbacks, inspector-style metadata, and " +
			"uniform get/set behavior are only partially modeled.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Keep the Go interface-based artist model, but add high-value shared mixins/options where " +
			"parity fixtures or user APIs need the behavior.",
	},
	{
		ID:              "artist-clipping-transform",
		CoverageID:      "artist",
		Title:           "Per-artist clipping and transform metadata",
		UpstreamModules: []string{"artist.py"},
		GoFiles:         []string{"core/artist.go", "core/lifecycle.go", "core/rasterization.go", "core/text.go", "core/picker.go"},
		CurrentEquivalent: "Shared artist metadata now carries visibility, alpha, labels, in-layout, " +
			"local stale state, explicit clip boxes / paths, clip-on behavior, and common static " +
			"coordinate/display transform overrides through draw traversal.",
		Gap: "The remaining scope is intentionally narrower than Matplotlib's Artist base class: " +
			"parent stale propagation, callback registries, animated draw integration, and the full " +
			"dynamic getp/setp property surface are not modeled as shared v1.0 behavior.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Static rendering parity for visibility, alpha, clipping, and transforms is covered by " +
			"shared metadata and catalog fixtures; lifecycle callbacks remain a separate interactive/API " +
			"surface rather than part of the static artist draw contract.",
	},
	{
		ID:              "ticker-formatter-catalog",
		CoverageID:      "axis-ticker-scale",
		Title:           "Locator and formatter catalog breadth",
		UpstreamModules: []string{"axis.py", "ticker.py", "scale.py", "dates.py", "category.py"},
		GoFiles:         []string{"core/axis.go", "core/tick.go", "core/date_tick.go", "core/units.go", "transform/scale_registry.go"},
		CurrentEquivalent: "Go has Axis, fixed / auto / log-family / symlog / asinh / logit / date / " +
			"category locators, formatter families for scalar, fixed, function, printf, str-method, " +
			"engineering, percent, log math text, logit, dates, categories, unit conversion hooks, " +
			"and scale-specific defaults including functionlog.",
		Gap: "The remaining 12.2 scope is split into concrete rows: 12.2B closed locator " +
			"edge semantics and catalog fixture coverage; 12.2C still tracks ScalarFormatter offset/scientific behavior, " +
			"log-family sparse minor labels, EngFormatter offset behavior, and formatter catalog " +
			"fixtures; 12.2D/E still track date/category and scale-default " +
			"catalog fixtures. The dates.py row-by-row audit is tracked in DateSurfaceAuditRows.",
		Decision: GapDecisionImplement,
		Rationale: "Tick labels and scale semantics are user-visible parity surfaces; each remaining " +
			"catalog gap is now tracked as an explicit 12.2 subtask rather than a broad audit bucket.",
	},
	{
		ID:              "tick-artist-model",
		CoverageID:      "axis-ticker-scale",
		Title:           "Explicit tick artist behavior",
		UpstreamModules: []string{"axis.py", "ticker.py"},
		GoFiles:         []string{"core/axis.go", "core/tick.go", "core/grid.go"},
		CurrentEquivalent: "Go draws ticks, minor ticks, grid lines, mirrored axes, and tick labels " +
			"from Axis state, with TickParams covering major/minor/both selection, axis selection, " +
			"length, width, colors, label size/color/rotation/pad/alignment, side visibility, " +
			"direction, reset, and grid styling.",
		Gap: "A Python-style Tick artist clone remains an explicit non-goal for v1.0 unless a " +
			"migration example needs per-tick object identity, callbacks, or artist-level stale state.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "The visible static tick behavior is exposed through Go axis-owned state and option " +
			"structs, which keeps the API typed while avoiding a dynamic Tick object hierarchy.",
	},
	{
		ID:              "transform-bbox-paths",
		CoverageID:      "transforms",
		Title:           "Transform, BBox, and transformed-path breadth",
		UpstreamModules: []string{"transforms.py", "path.py", "bezier.py"},
		GoFiles:         []string{"transform/transform.go", "transform/graph.go", "transform/node.go", "internal/geom/geom.go"},
		CurrentEquivalent: "Go has affine, separable, blended, chained, offset, display-rect, scale, " +
			"and graph-backed transforms plus rect/path primitives with BBox-style union, " +
			"intersection, expansion/padding, point containment, affine transformed bounds, and " +
			"inverse-transformed bounds, plus null-rectangle accumulation helpers.",
		Gap: "The remaining 12.2G scope is split into frozen transform snapshots, transformed-path " +
			"cache helpers with invalidation, anchored/null rect helpers for annotation/layout parity, " +
			"and remaining path/bezier helpers only when they affect visible clipping or layout behavior.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Preserve the lean transform graph, adding focused BBox/path helpers when annotation, " +
			"layout, clipping, or image-transform parity needs them.",
	},
	{
		ID:              "line2d-marker-data-semantics",
		CoverageID:      "lines",
		Title:           "Line2D marker, data, and transformed-path semantics",
		UpstreamModules: []string{"lines.py", "markers.py"},
		GoFiles:         []string{"core/line.go", "core/plot.go", "core/scatter.go"},
		CurrentEquivalent: "Go Line2D now combines stroked polylines with data markers, marker face / " +
			"edge colors, explicit auto/none color sentinels, point-based marker edge widths, " +
			"fillstyle and half-fill rendering, markevery forms, gapcolor, invalid-point line breaks, " +
			"data accessors, stale invalidation, legends, and custom / tuple / mathtext marker paths.",
		Gap: "The remaining scope is performance and Python-surface breadth rather than visible static " +
			"marker semantics: Matplotlib's transformed-path cache, sorted-data subslicing, and broad " +
			"dynamic setter/getter aliases are not cloned one-for-one.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Visible Line2D marker, data, and legend parity is now covered by focused fixtures; " +
			"future work should add caches or aliases only when profiling or migration examples need them.",
	},
	{
		ID:              "collection-variants-setters",
		CoverageID:      "collections",
		Title:           "Collection variants and setter surface",
		UpstreamModules: []string{"collections.py"},
		GoFiles:         []string{"core/collection.go", "core/mesh.go", "core/eventplot.go", "core/hexbin.go", "core/triangulation.go"},
		CurrentEquivalent: "Go has PathCollection, LineCollection, PatchCollection, PolyCollection, " +
			"QuadMesh, event collections, and hexbin collections; PColor uses the same rectilinear " +
			"QuadMesh implementation as PColorMesh.",
		Gap: "Specialized upstream collection classes and the broad mutable setter / scalar-mapping " +
			"surface are only partially represented. Matplotlib's PolyQuadMesh-specific pcolor behavior, " +
			"including masked-coordinate polygon dropping and per-cell hatch/linestyle flexibility, is " +
			"intentionally omitted until a visible fixture needs it.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Keep collection data structures compact, but add missing variants or setters when they " +
			"are required for public examples, scalar mappables, or backend-native batching.",
	},
	{
		ID:              "collection-scalar-mapping",
		CoverageID:      "collections",
		Title:           "Collection scalar-mappable updates and offset transforms",
		UpstreamModules: []string{"collections.py", "cm.py", "colors.py"},
		GoFiles:         []string{"core/collection.go", "core/scalar_mappable.go", "core/mesh.go", "core/scatter.go"},
		CurrentEquivalent: "Go collections carry scalar-mappable metadata; PathCollection, LineCollection, " +
			"PatchCollection, and QuadMesh expose Go-style SetArray, SetColormap, SetNorm, SetCLim, and " +
			"face-edge tracking helpers; PathCollection supports a separate offset coordinate transform.",
		Gap: "Matplotlib's broader draw-time scalar-mapping callbacks, QuadMesh array shape " +
			"variants beyond the supported flat/nearest cell and Gouraud vertex updates, and the full " +
			"mutable collection setter surface are still only partially represented.",
		Decision: GapDecisionImplement,
		Rationale: "Scalar-mappable collection updates are required for colorbar correctness and public " +
			"collection parity; obscure mutable setters can stay Go-idiomatic.",
	},
	{
		ID:              "patch-style-registries",
		CoverageID:      "patches",
		Title:           "Patch shapes, BoxStyle, ArrowStyle, and ConnectionStyle registries",
		UpstreamModules: []string{"patches.py", "hatch.py"},
		GoFiles:         []string{"core/patch.go", "core/patch_extra.go", "core/arrow_patch.go"},
		CurrentEquivalent: "Go has common patch shapes, hatch routing, FancyBboxPatch, FancyArrowPatch, " +
			"ConnectionPatch, and several extra patch classes.",
		Gap: "Full box style, arrow style, connection style, hatch-density, and specialized patch " +
			"catalog parity still needs enumeration against upstream registries.",
		Decision: GapDecisionImplement,
		Rationale: "These are enumerable public catalogs; missing entries should be implemented or tracked " +
			"as explicit intentional divergences.",
	},
	{
		ID:              "text-font-layout",
		CoverageID:      "text-annotation-legend",
		Title:           "Text font, layout, wrapping, and annotation breadth",
		UpstreamModules: []string{"text.py", "textpath.py", "legend.py", "legend_handler.py", "offsetbox.py"},
		GoFiles:         []string{"core/text.go", "core/mathtext.go", "core/legend.go", "core/anchored_text.go", "render/font_manager.go", "render/text_path.go"},
		CurrentEquivalent: "Go supports text, rotated text, annotations, text paths, MathText, TeX paths, " +
			"anchored text, legends, and renderer font resolution.",
		Gap: "Font property breadth, wrapping, multiline layout, annotation coordinate modes, " +
			"AnnotationBbox / OffsetBox families, legend handler maps, proxy artists, and draggable " +
			"legend behavior are thin or absent.",
		Decision: GapDecisionImplement,
		Rationale: "Text and legend behavior is heavily visible in parity images and migration examples; " +
			"missing surfaces should be implemented incrementally with fixture coverage.",
	},
	{
		ID:              "text-font-properties",
		CoverageID:      "text-annotation-legend",
		Title:           "Text font property surface",
		UpstreamModules: []string{"text.py", "font_manager.py", "textpath.py"},
		GoFiles:         []string{"core/text.go", "render/font_manager.go", "render/text_shaping.go", "render/text_path.go"},
		CurrentEquivalent: "Go text options expose size, color, MathText / TeX routing, and renderer-level " +
			"font resolution / shaping.",
		Gap: "Per-text font family, style, weight, stretch, variant, font features, language, " +
			"math font, parse_math, and usetex-style setters are not modeled as a cohesive " +
			"artist-level property set.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Expose a Go TextOptions font-property struct that maps to renderer font keys instead " +
			"of mirroring Python's dynamic setter catalog.",
	},
	{
		ID:              "annotation-coordinate-model",
		CoverageID:      "text-annotation-legend",
		Title:           "Annotation coordinate and clipping model",
		UpstreamModules: []string{"text.py", "patches.py"},
		GoFiles:         []string{"core/text.go", "core/arrow_patch.go", "transform/transform.go"},
		CurrentEquivalent: "Go annotations support text, arrows, arrow styles, connection styles, and " +
			"overlay drawing.",
		Gap: "Matplotlib's separate xycoords / textcoords, annotation clipping policy, " +
			"AnnotationBbox behavior, and tightbbox / window-extent interaction are only partially " +
			"represented.",
		Decision: GapDecisionImplement,
		Rationale: "Common annotation coordinate modes are widely used in Matplotlib examples and should " +
			"be implemented with transform-backed Go options.",
	},
	{
		ID:              "image-class-breadth",
		CoverageID:      "image",
		Title:           "Image artist class breadth and interpolation semantics",
		UpstreamModules: []string{"image.py"},
		GoFiles:         []string{"core/image.go", "core/image_api.go", "core/matrix_helpers.go", "backends/agg/interpolation.go"},
		CurrentEquivalent: "Go supports scalar matrix images, imshow-style options, matshow, spy, alpha, " +
			"origin / extent, interpolation modes, and transformed images.",
		Gap: "FigureImage, BboxImage, NonUniformImage, PcolorImage, pcolorfast, and the full " +
			"Matplotlib interpolation / antialias policy are not fully represented.",
		Decision: GapDecisionImplement,
		Rationale: "Image class and interpolation gaps directly affect visual parity and should be closed " +
			"where they map to real static rendering behavior.",
	},
	{
		ID:              "colorbar-orientation-ticks",
		CoverageID:      "colorbar",
		Title:           "Colorbar orientation, placement, tick, and boundary behavior",
		UpstreamModules: []string{"colorbar.py", "colorizer.py"},
		GoFiles:         []string{"core/colorbar.go", "core/scalar_mappable.go", "core/norm.go"},
		CurrentEquivalent: "Go has figure-level colorbars backed by ScalarMappable, colormap/norm routing, " +
			"vertical and horizontal placement, labels, extension patches, explicit tick lists, explicit " +
			"boundaries/values, uniform/proportional boundary spacing, drawedges, rectangular extensions, " +
			"shrink/anchor options, and mutable mappable synchronization.",
		Gap: "Custom formatter objects, gridspec-specific helpers, and " +
			"multi-parent colorbar placement remain intentionally partial in the current figure/axes model.",
		Decision: GapDecisionImplement,
		Rationale: "Colorbar rendering is common and already covered by parity fixtures; missing public " +
			"semantics should become catalog-visible work.",
	},
	{
		ID:              "colors-norms-lightsource",
		CoverageID:      "colors-cm",
		Title:           "Advanced colors, norms, and LightSource",
		UpstreamModules: []string{"colors.py", "cm.py", "_cm.py", "_cm_listed.py", "_color_data.py"},
		GoFiles:         []string{"color/colormap.go", "color/listed_colormaps.go", "color/named_colors.go", "core/norm.go"},
		CurrentEquivalent: "Go has named colors, listed and segmented colormaps, reversed/resampled " +
			"colormaps, and common norms including LogNorm, SymLogNorm, PowerNorm, TwoSlopeNorm, " +
			"CenteredNorm, BoundaryNorm, AsinhNorm, and NoNorm.",
		Gap: "FuncNorm is represented by custom ScalarNormalizer implementations; MultiNorm, " +
			"multivar/bivar colormaps, LightSource, and edge-case color conversion behavior remain " +
			"outside the narrow v1.0 surface until a visible fixture needs them.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "AsinhNorm is implemented because it improves image and colorbar parity directly; " +
			"less common multivariate color machinery does not fit the current single-scalar mapping API.",
	},
	{
		ID:              "pyplot-wrapper-surface",
		CoverageID:      "pyplot-state",
		Title:           "Pyplot wrapper and stateful migration surface",
		UpstreamModules: []string{"pyplot.py", "_pylab_helpers.py"},
		GoFiles:         []string{"pyplot/pyplot.go", "pyplot/pyplot_test.go", "canvas/canvas.go"},
		CurrentEquivalent: "Go has a pyplot package with figure/current-axes state, common plot wrappers, " +
			"rc helpers, savefig, show, and pause hooks.",
		Gap: "The wrapper surface is much smaller than upstream pyplot, especially for overloads, " +
			"state transitions, interactive mode, and many convenience functions.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Do not clone Python signatures wholesale; add high-value migration wrappers where they " +
			"reduce friction and delegate to the object-oriented Go API.",
	},
	{
		ID:              "backend-canvas-manager-lifecycle",
		CoverageID:      "renderer-backends",
		Title:           "Backend canvas, manager, tool, and event lifecycle",
		UpstreamModules: []string{"backend_bases.py", "backend_tools.py"},
		GoFiles:         []string{"render/render.go", "backends/registry.go", "canvas/canvas.go", "canvas/dispatcher.go", "canvas/tool.go"},
		CurrentEquivalent: "Go separates renderer contracts, backend registry, figure canvas, managers, " +
			"dispatchers, navigation, and tools across render/backends/canvas packages.",
		Gap: "Matplotlib's FigureCanvasBase, FigureManagerBase, ToolManager, toolbar, timer, " +
			"draw-event, and interactive lifecycle semantics are only partially mirrored.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Preserve the Go package split, while completing lifecycle semantics through the Phase 4 " +
			"interactive backend and event-loop work.",
	},
}

// FoundationAPIGapAudit returns the Phase 9A.2 foundational API gap decisions.
func FoundationAPIGapAudit() []FoundationAPIGap {
	out := make([]FoundationAPIGap, len(foundationAPIGaps))
	copy(out, foundationAPIGaps)
	for i := range out {
		out[i].UpstreamModules = append([]string(nil), out[i].UpstreamModules...)
		out[i].GoFiles = append([]string(nil), out[i].GoFiles...)
	}
	return out
}

// LookupFoundationAPIGap finds a Phase 9A.2 gap by stable ID.
func LookupFoundationAPIGap(id string) (FoundationAPIGap, bool) {
	for _, gap := range FoundationAPIGapAudit() {
		if gap.ID == id {
			return gap, true
		}
	}
	return FoundationAPIGap{}, false
}

// FoundationGapsForUpstreamModule returns gap rows that cite an upstream module.
func FoundationGapsForUpstreamModule(module string) []FoundationAPIGap {
	var out []FoundationAPIGap
	for _, gap := range FoundationAPIGapAudit() {
		for _, upstream := range gap.UpstreamModules {
			if upstream == module {
				out = append(out, gap)
				break
			}
		}
	}
	return out
}

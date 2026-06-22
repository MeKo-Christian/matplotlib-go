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
		GoFiles:         []string{"transform/transform.go", "transform/graph.go", "transform/node.go", "geom/geom.go"},
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
		GoFiles:         []string{"core/collection_common.go", "core/collection_path.go", "core/collection_line.go", "core/collection_patch.go", "core/collection_poly.go", "core/collection_quadmesh.go", "core/collection_fillbetween.go", "core/mesh.go", "core/eventplot.go", "core/hexbin.go", "core/triangulation.go"},
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
		GoFiles:         []string{"core/collection_common.go", "core/collection_path.go", "core/collection_line.go", "core/collection_patch.go", "core/collection_quadmesh.go", "core/scalar_mappable.go", "core/mesh.go", "core/scatter.go"},
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
		CurrentEquivalent: "Go has common patch shapes, hatch routing, FancyBboxPatch with the upstream " +
			"BoxStyle registry entries, renderer-neutral hatch geometry for the upstream hatch " +
			"character set, AGG shape hatches, vector-native shape hatch patterns, FancyArrowPatch, " +
			"DPI-correct Matplotlib-style FancyArrowPatch endpoint shrinking, patchA/patchB endpoint clipping, explicit ConnectionPatch shrink conversion, FancyArrowPatch / ConnectionPatch round cap/join defaults, FancyArrowPatch mutation-aspect arrow transmutation, ArrowStyle Simple / Fancy / Wedge quadratic-connection geometry, ArrowStyle Wedge shrink-factor behavior, ArrowStyle curve linewidth-projected line shortening and bracket scale overrides, ArrowStyle BarAB zero-length bracket defaults, ConnectionStyle Arc3 zero-rad quadratic paths, Arc defaults and rounded arm geometry, ConnectionStyle Bar angle projection, ConnectionPatch, and several extra patch classes.",
		Gap: "Remaining patch scope is Python helper/debug function parity, the broad mutable " +
			"Patch property grammar, and broader visual fixture closure beyond the focused " +
			"patch_showcase and patch_style_matrix registry cases.",
		Decision: GapDecisionImplement,
		Rationale: "These are enumerable public catalogs; missing entries should be implemented or tracked " +
			"as explicit intentional divergences.",
	},
	{
		ID:              "text-font-layout",
		CoverageID:      "text-annotation-legend",
		Title:           "Text layout, wrapping, and bounds",
		UpstreamModules: []string{"text.py", "textpath.py"},
		GoFiles:         []string{"core/text.go", "core/mathtext.go", "render/text_path.go"},
		CurrentEquivalent: "Go supports text, rotated text, multiline drawing, text bbox patches, " +
			"text paths, MathText, TeX paths, per-text font routing, and renderer font resolution. " +
			"Text and annotation drawing use shared artist-alpha metadata for visible text and arrow colors. " +
			"TextOptions.WrapWidth provides explicit display-pixel word wrapping, TextOptions.Wrap computes Matplotlib-style figure-box wrap widths, TextOptions.MultiAlignment separates per-line multiline alignment from block anchoring, TextOptions/AnnotationOptions.Linespacing controls multiline baseline advance from renderer font metrics and line gaps, and multiline text routes angle-aware draws through rotated renderer capabilities. " +
			"Multiline text also routes glyph paths through path effects. TextOptions.RotationMode supports Matplotlib-style anchor, xtick, and ytick rotation modes, TextVerticalAlign includes center_baseline, and text bbox patches, including multiline block bboxes, rotate with rotated text. " +
			"The text_annotation_matrix fixture gives focused coverage for multiline text, rotated " +
			"text, text bbox output, and structured font requests.",
		Gap: "After Phase 12.4C, remaining text layout scope is limited to fixture-specific " +
			"pixel residuals and text rasterization interactions rather than known broad " +
			"bounding-box, baseline, rotation-mode, multiline, wrapping, or bbox-patch gaps.",
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
		CurrentEquivalent: "Go text and annotation options expose size, color, FontKey overrides, " +
			"structured FontProperties for family/style/weight/stretch/variant/file/language/math-font requests, " +
			"OpenType feature toggles, MathText / TeX routing, per-artist ParseMath control, " +
			"and renderer-level font resolution / shaping.",
		Gap:      "Usetex-style per-artist setters and exact fontconfig stretch/variant scoring remain partial.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Expose a Go TextOptions font-property struct that maps to renderer font keys instead " +
			"of mirroring Python's dynamic setter catalog.",
	},
	{
		ID:              "annotation-coordinate-model",
		CoverageID:      "text-annotation-legend",
		Title:           "Annotation coordinate and clipping model",
		UpstreamModules: []string{"text.py", "patches.py"},
		GoFiles:         []string{"core/text.go", "core/annotation_box.go", "core/arrow_patch.go", "transform/transform.go"},
		CurrentEquivalent: "Go annotations support text, arrows, arrow styles, connection styles, " +
			"explicit annotation_clip behavior, Matplotlib's default data-only annotation_clip policy, " +
			"text bbox patches, rotated and multiline annotation text, separate annotated-point and " +
			"text-position coordinate spaces through TextPosition/TextCoords, static AnnotationBbox-style " +
			"text/image boxes, and overlay drawing.",
		Gap: "Remaining annotation scope is Matplotlib's dynamic setter/getter grammar and OffsetFrom's " +
			"live artist-bbox callable; tight layout/window-extent behavior is represented by the measured " +
			"text/annotation layout helpers and catalog fixtures rather than public Python get_window_extent methods.",
		Decision: GapDecisionImplement,
		Rationale: "Common annotation coordinate modes are widely used in Matplotlib examples and should " +
			"be implemented with transform-backed Go options.",
	},
	{
		ID:              "offsetbox-static-layout",
		CoverageID:      "text-annotation-legend",
		Title:           "Offset-box static layout",
		UpstreamModules: []string{"offsetbox.py"},
		GoFiles:         []string{"core/anchored_text.go", "core/anchored_drawing_area.go", "core/anchored_packer.go", "core/anchored_sizebar.go", "core/annotation_box.go", "core/text.go"},
		CurrentEquivalent: "Go has static equivalents for AnchoredText, AnchoredDrawingArea, " +
			"AnnotationBbox with TextArea / OffsetImage-style content, horizontal and vertical " +
			"packers for text / drawing / image children, and AnchoredSizeBar. The " +
			"text_annotation_matrix fixture gives focused coverage for those static offset-box " +
			"paths.",
		Gap: "Remaining offsetbox scope is exact Python mutable class hierarchy and arbitrary custom " +
			"child dispatch; draggable GUI-only boxes and the module DEBUG flag are explicit intentional " +
			"omissions from the static renderer-neutral v1.0 surface.",
		Decision: GapDecisionImplement,
		Rationale: "Static offset boxes affect annotations, legends, and axes-grid style labels; GUI-only " +
			"dragging should stay separate from renderer-neutral layout parity.",
	},
	{
		ID:              "legend-handler-layout",
		CoverageID:      "text-annotation-legend",
		Title:           "Legend handlers and static layout",
		UpstreamModules: []string{"legend.py", "legend_handler.py"},
		GoFiles:         []string{"core/legend.go"},
		CurrentEquivalent: "Go legends support static placement, titles, frame visibility, multi-column " +
			"layout, marker scaling, scatter sample counts, explicit proxy entries, typed per-artist " +
			"handler overrides, scalar-mapped collection samples, errorbar samples, combined stem samples, " +
			"and LegendBest avoidance for line/path intersections, collection offsets, image/rectangle " +
			"bboxes, patch paths, and annotation anchors. Figure-level legends collect across axes and " +
			"stack with figure labels. The legend_layout_matrix fixture gives focused coverage for those " +
			"handler and layout paths.",
		Gap: "Draggable legend behavior and exact arbitrary Python handler-map dispatch are explicit " +
			"typed-API divergences; static samples are represented through typed Go entries, collected " +
			"artist samples, and per-artist handler overrides.",
		Decision: GapDecisionImplement,
		Rationale: "Legend output is visible in many examples; handler parity should stay typed and " +
			"Go-style while matching static rendered samples where feasible.",
	},
	{
		ID:              "image-class-breadth",
		CoverageID:      "image",
		Title:           "Image artist class breadth and interpolation semantics",
		UpstreamModules: []string{"image.py"},
		GoFiles:         []string{"core/image.go", "core/image_api.go", "core/matrix_helpers.go", "backends/agg/interpolation.go"},
		CurrentEquivalent: "Go supports scalar matrix images, imshow-style options, matshow, spy, alpha, " +
			"origin / extent, the full Matplotlib interpolation registry in AGG, a focused " +
			"interpolation matrix fixture, transformed images, and a typed PColorFast alias " +
			"through the QuadMesh-backed PColorMesh path.",
		Gap: "FigureImage, BboxImage, NonUniformImage, PcolorImage, and image IO helpers now " +
			"have explicit public-surface classifications; remaining image breadth is exact " +
			"transformed-image resampling edge behavior.",
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
		GoFiles:         []string{"color/colormap.go", "color/listed_colormaps.go", "color/named_colors.go", "color/petroff.go", "core/norm.go"},
		CurrentEquivalent: "Go has named colors, listed and segmented colormaps, reversed/resampled " +
			"colormaps, the petroff10 color sequence, and common norms including FuncNorm, LogNorm, " +
			"SymLogNorm, PowerNorm, TwoSlopeNorm, CenteredNorm, BoundaryNorm, AsinhNorm, and NoNorm.",
		Gap: "MultiNorm, multivar/bivar colormaps, LightSource, and edge-case color conversion " +
			"behavior remain outside the narrow v1.0 surface until a visible fixture needs them. " +
			"MultiNorm in particular is a Matplotlib 3.11 feature (absent from the 3.10.9 reference) " +
			"that depends on the deferred multivariate-colormap machinery.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "AsinhNorm is implemented because it improves image and colorbar parity directly; " +
			"less common multivariate color machinery does not fit the current single-scalar mapping API.",
	},
	{
		ID:              "pyplot-wrapper-surface",
		CoverageID:      "pyplot-state",
		Title:           "Pyplot wrapper and stateful migration surface",
		UpstreamModules: []string{"pyplot.py", "_pylab_helpers.py"},
		GoFiles:         []string{"pyplot/pyplot.go", "pyplot/pyplot_registry_layout_test.go", "pyplot/pyplot_wrappers_test.go", "canvas/canvas.go"},
		CurrentEquivalent: "Go has a pyplot package with figure/current-axes state, common plot wrappers, " +
			"text, annotation, reference-line, span, axis-limit, and axis-scale wrappers, rc helpers, savefig, show, and pause hooks. " +
			"Phase 17.6.7 audited the wrapper surface against vendored pyplot.py/_pylab_helpers.py: state transitions, " +
			"interactive-mode hooks, and current figure/axes behavior all delegate to the object-oriented core API, and every " +
			"upstream row carries an explicit decision (locked by TestPyplotStateSurfaceRowsAreExplicitlyDecided).",
		Gap: "Intentional signature divergences remain by design: setter-only limit/tick helpers, label setters that return " +
			"no Text handle, omitted numeric figure registry and current-mappable shortcuts (sci/clim/set_cmap), omitted " +
			"GUI-blocking helpers, and typed options instead of Python overload/data=/property-dict grammar.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Do not clone Python signatures wholesale; the closed object-oriented surface is reachable through the " +
			"existing wrappers plus typed core.*Options, so 17.6.7 added no new public wrappers and documented every " +
			"intentional divergence in docs/matplotlib-migration-notes.md instead.",
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
			"and draw/resize/close lifecycle semantics are mirrored idiomatically; only GUI-toolkit-specific " +
			"behavior (live window events, cursor/status presentation, GUI tool widgets) is out of scope.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Phase 17.6.8 closed the backend lifecycle as a static-vs-GUI split: draw-event/resize/close " +
			"routing, timer start/stop/running, navigation, and the home/back/forward/pan/zoom/save tools are " +
			"idiomatic equivalents, while purely GUI tools (fullscreen, grid, help, copy, cursor, quit, " +
			"subplot-config, rubberband) are explicit intentional omissions.",
	},
	{
		ID:              "widget-interaction-scope",
		CoverageID:      "widgets-events-animation",
		Title:           "Widget and selector interaction scope",
		UpstreamModules: []string{"widgets.py"},
		GoFiles: []string{
			"core/widget_button.go",
			"core/widget_slider.go",
			"core/widget_rangeslider.go",
			"core/widget_checkbuttons.go",
			"core/widget_radiobuttons.go",
			"core/widget_textbox.go",
			"core/selectors_common.go",
			"core/widgets_common.go",
			"canvas/widget_interaction.go",
			"canvas/dispatcher.go",
		},
		CurrentEquivalent: "Go has widget artists and canvas event routing for buttons, sliders, " +
			"range sliders, check buttons, radio buttons, text boxes, span selectors, rectangle / ellipse " +
			"selectors, polygon selectors, lasso selectors, cursor, and multi-cursor helpers, with interaction " +
			"tests for mouse, keyboard, hover, callback ordering, active-state, handle behavior, disabled state, " +
			"and keyboard modifiers verified under both widget visual styles.",
		Gap: "Only GUI-only behavior remains out of scope: useblit, live cursor/status integration, " +
			"input-method editing, the SubplotTool/menu widgets, and browser-demo interaction.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Phase 17.6.8 completed the widget interaction edge cases (active-state, handles, disabled, " +
			"modifiers) with style-parameterized tests; remaining GUI-only paths are documented intentional " +
			"omissions, with browser-demo interaction deferred to Phase 19.",
	},
	{
		ID:              "animation-playback-writers",
		CoverageID:      "widgets-events-animation",
		Title:           "Animation playback and writer scope",
		UpstreamModules: []string{"animation.py"},
		GoFiles:         []string{"animation/animation.go", "animation/writers.go", "animation/writer_pillow.go", "animation/writer_apng.go", "animation/writer_html.go", "animation/writer_ffmpeg.go", "canvas/scheduler.go"},
		CurrentEquivalent: "Go has FuncAnimation- and ArtistAnimation-style stepping on top of the " +
			"canvas scheduler, including init callbacks, repeat handling, repeat-delay ticks, animated " +
			"artist tracking, blit-region hooks, and event-loop start/stop behavior, plus Save to a " +
			"deterministic GIF via the AbstractMovieWriter/GifWriter stack (image/gif), dependency-free " +
			"full-RGBA APNG output, self-contained HTML playback payloads for web demos, and optional MP4/WebM output through the -tags ffmpeg " +
			"writer, all covered by unit tests.",
		Gap: "Only ImageMagick-style external output remains out of scope; " +
			"default builds still report ErrWriterUnsupported for ffmpeg-backed MP4/WebM.",
		Decision: GapDecisionIdiomaticEquivalent,
		Rationale: "Phase 17.6.8 chose a pure-Go GifWriter (image/gif, no external tools) as the default writer scope; " +
			"Phase 6.2 adds dependency-free APNG and HTML writers plus opt-in ffmpeg shellout support behind a build tag and runtime executable detection.",
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

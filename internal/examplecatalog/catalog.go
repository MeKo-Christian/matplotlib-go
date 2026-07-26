package examplecatalog

// Case describes a plotting example that participates in parity testing.
//
// The catalog is the shared source of truth for the relationship between
// user-facing showcase examples, Go golden fixtures, Matplotlib references,
// and the curated web demo subset.
//
// Layout rules:
//   - Showcase: true        → body lives at examples/<id>/example.go (also
//     the GoPath for source display); test/parity
//     wrapper imports it.
//   - FixtureOnly: true     → body lives at test/parity/<id>/plot.go; no
//     examples/<id>/ entry. These are stress
//     fixtures, not user examples.
//   - neither               → body lives at test/parity/<id>/plot.go; not
//     surfaced in examples/. They participate in
//     parity testing but are not curated showcases.
type Case struct {
	ID          string
	Topic       string
	Title       string
	Description string
	Width       int
	Height      int
	DPI         int
	GoPath      string
	PythonPath  string
	Optional    bool
	FixtureOnly bool
	Showcase    bool
	WebDemoID   string

	// NativeBackend marks fixtures that intentionally exercise backend-native
	// renderer capabilities rather than renderer-neutral fallback paths.
	// NativeCapabilities names the backend capabilities the fixture covers.
	NativeBackend      string
	NativeCapabilities []string
	SVGGoldenFamily    string
	GoBasicSmokeFamily string
	// SkiaParityFamily groups cases consumed by the Skia-vs-AGG parity harness
	// (see backends/skia/parity_test.go). A non-empty value opts the case in;
	// the per-case MaxMeanAbs / MaxRMSE fields override the harness defaults
	// when non-zero. Families share the same naming convention as
	// SVGGoldenFamily / GoBasicSmokeFamily.
	SkiaParityFamily    string
	PickPointData       *[2]float64
	PickPointPixel      *[2]float64
	Pickable            bool
	NoPickReason        string
	SkipInteractivePan  bool
	SkipInteractiveZoom bool

	// MaxMeanAbs / MaxRMSE override the matplotlib reference-compare tolerance
	// defaults for this case. Zero means "use the package default". Only
	// consulted by tests in test/; harmless metadata otherwise.
	//
	// MaxRMSE bounds the L2 residual and MaxMeanAbs the L1 residual; the two are
	// genuinely independent, so a case can carry both. There is deliberately no
	// MinPSNR: imagecmp derives PSNR from RMSE as 20*log10(255/RMSE), so a PSNR
	// floor can only restate a MaxRMSE ceiling. The field existed until Phase
	// 3.1, when the uint8 overflow that made PSNR look independent was fixed;
	// every floor that was both binding and reachable was folded into MaxRMSE.
	// Neither metric localizes a residual — see docs/plans/phase3-tolerance-audit.md.
	MaxMeanAbs float64
	MaxRMSE    float64

	// MaxDiffPixels / MaxLargestCluster bound the *shape* of the residual that
	// MaxMeanAbs and MaxRMSE, being whole-image averages, cannot express: how
	// many pixels differ at all, and how big the largest 8-connected clump of
	// them is. Zero means "unbounded" (the amplitude gates still apply).
	//
	// This is the gate that catches a wholly misplaced glyph. basic_line's
	// residual is one y tick label drawn a pixel low — 139 pixels at
	// per-channel difference up to 249 — which still scores RMSE 2.49 against
	// a 2.8 allowance, because 139 pixels out of 230,400 cannot move an
	// average. A cluster ceiling sees it immediately. Added in Phase 3.2;
	// Phase 3.6 ratchets the remaining cases.
	MaxDiffPixels     int
	MaxLargestCluster int
}

const (
	DefaultWidth  = 640
	DefaultHeight = 360
	DefaultDPI    = 100
)

var cases = []Case{
	{ID: "basic_line", Topic: "lines", Title: "Basic Line", Description: "A minimal line plot with explicit limits, labels, and one Line2D artist.", Showcase: true, GoBasicSmokeFamily: "line", SkiaParityFamily: "line", MaxRMSE: 2.8, MaxDiffPixels: 180, MaxLargestCluster: 92},
	{ID: "basic_line_labels", Topic: "lines", Title: "Basic Line Labels", Description: "A minimal line plot with x and y axis labels for label layout parity checks.", Showcase: true, GoBasicSmokeFamily: "line", SkiaParityFamily: "line", MaxRMSE: 2.8, MaxDiffPixels: 180, MaxLargestCluster: 92},
	{ID: "joins_caps", Topic: "lines", Title: "Line Joins and Caps", MaxRMSE: 0.3},
	{ID: "dashes", Topic: "lines", Title: "Dash Patterns", Description: "Multiple line styles showing dash arrays, cap styles, and legend labeling.", Showcase: true, MaxRMSE: 1.6},
	{ID: "lines_markers_gallery", Topic: "lines", Title: "Line and Marker Style Gallery", Description: "A combined gallery of dash arrays, line joins and caps, a built-in marker grid with open-fill markers, and a multi-series legend.", WebDemoID: "lines", Showcase: true, Width: 840, Height: 620, MaxMeanAbs: 1.0, MaxRMSE: 2.8},
	{ID: "line2d_semantics", Topic: "lines", Title: "Line2D Semantics", FixtureOnly: true, MaxRMSE: 2.6, MaxDiffPixels: 380, MaxLargestCluster: 23},
	// MaxRMSE 3.3: MeanAbs 0.14 / PSNR ~52 dB; residual is sub-pixel marker-edge antialiasing under the 3.10.9 reference set.
	{ID: "line2d_markers", Topic: "lines", Title: "Line2D Markers", FixtureOnly: true, SkiaParityFamily: "markers", MaxRMSE: 3.3, MaxDiffPixels: 4300, MaxLargestCluster: 2600},
	{ID: "path_effects", Topic: "effects", Title: "Path Effects", FixtureOnly: true, MaxRMSE: 0.3, SkiaParityFamily: "effects"},
	{ID: "pattern_gradient_effects", Topic: "effects", Title: "Pattern and Gradient Effects", FixtureOnly: true, MaxMeanAbs: 1.0, MaxRMSE: 1.608, SkiaParityFamily: "effects"},
	// PSNR ~64 dB / MeanAbs ~0.01 / RMSE ~0.5: the sketch wiggle is RNG-, phase-
	// AND vertex-exact vs Matplotlib, and the figure background now matches too.
	// Two fixes got here: (1) enabling path simplification by default
	// (style.PathSimplify, matching path.simplify=True) — Matplotlib simplifies a
	// line before the sketch filter, so the segmentator + RNG must see the same
	// simplified polyline or the wiggle desyncs (7.7 -> 3.5 RMSE); (2) rendering the
	// figure background the way Matplotlib does — a NON-antialiased figure patch
	// (figure.patch, antialiased=False) composited over a transparent canvas, so the
	// xkcd wiggle perforates the border with the same 42 fully-transparent notch
	// pixels (3.5 -> 0.5 RMSE; the notches are byte-exact). The residual ~0.5 is the
	// irreducible sub-pixel AA floor on the sine curve. See core.drawSketchedFigurePatch.
	{ID: "sketch_xkcd", Topic: "effects", Title: "Sketch / xkcd Mode", Description: "A sine curve and a flat reference line under Matplotlib's xkcd sketch mode, exercising the global path.sketch perturbation on spines, ticks, and lines.", FixtureOnly: true, MaxMeanAbs: 0.05, MaxRMSE: 1.0},
	{ID: "scatter_basic", Topic: "scatter", Title: "Basic Scatter", Description: "A compact scatter plot with variable marker size, color, alpha, and axes labels.", Showcase: true, GoBasicSmokeFamily: "scatter", SkiaParityFamily: "scatter", MaxRMSE: 0.3},
	{ID: "scatter_marker_types", Topic: "scatter", Title: "Scatter Marker Types", MaxRMSE: 2.8},
	{ID: "scatter_advanced", Topic: "scatter", Title: "Advanced Scatter", MaxRMSE: 0.3},
	{ID: "scatter_gallery", Topic: "scatter", Title: "Advanced Scatter Gallery", Description: "A combined gallery of colormapped scalar mapping, variable marker size, alpha blending of overlapping markers, and multiple marker families.", WebDemoID: "scatter", Showcase: true, Width: 840, Height: 620, MaxMeanAbs: 1.0, MaxRMSE: 0.64},
	{ID: "bar_basic_frame", Topic: "bar", Title: "Bar Frame", MaxRMSE: 0.3},
	{ID: "bar_basic_ticks", Topic: "bar", Title: "Bar Ticks", MaxRMSE: 0.3},
	{ID: "bar_basic_tick_labels", Topic: "bar", Title: "Bar Tick Labels", MaxRMSE: 0.3},
	{ID: "bar_basic_title", Topic: "bar", Title: "Bar Title", MaxRMSE: 0.3},
	{ID: "bar_basic", Topic: "bar", Title: "Basic Bars", Description: "A categorical bar chart with tick labels, title text, and framed axes.", Showcase: true, SVGGoldenFamily: "bar", GoBasicSmokeFamily: "bar", SkiaParityFamily: "bar", MaxRMSE: 0.3},
	{ID: "bar_horizontal", Topic: "bar", Title: "Horizontal Bars", MaxRMSE: 0.3},
	{ID: "bar_grouped", Topic: "bar", Title: "Grouped Bars", MaxRMSE: 0.5},
	{ID: "bar_variants", Topic: "bar", Title: "Bar Variants Gallery", Description: "A combined gallery of vertical bars with tick labels, horizontal bars, grouped bars, and stacked bars with bar labels.", WebDemoID: "bars", Showcase: true, Width: 840, Height: 620, MaxMeanAbs: 1.5, MaxRMSE: 2.4},
	{ID: "bar_yerr", Topic: "bar", Title: "Bars with Error Bars", MaxMeanAbs: 0.02, MaxRMSE: 0.064},
	{ID: "fill_basic", Topic: "fill", Title: "Fill to Baseline", Description: "A filled region under a smooth curve, useful for area-chart and alpha blending checks.", Showcase: true, MaxRMSE: 0.5},
	{ID: "fill_between", Topic: "fill", Title: "Fill Between Curves", GoBasicSmokeFamily: "fill", SkiaParityFamily: "fill", MaxRMSE: 0.2},
	{ID: "fill_stacked", Topic: "fill", Title: "Stacked Fill", MaxRMSE: 0.2},
	{ID: "stackplot_streamgraph", Topic: "fill", Title: "Streamgraph (weighted_wiggle)", Description: "A stackplot using the weighted_wiggle baseline (streamgraph layout) with default property-cycle colors.", Optional: true, MaxMeanAbs: 2.0, MaxRMSE: 0.5},
	{ID: "fill_variants", Topic: "fill", Title: "Fill Variants Gallery", Description: "A combined gallery of fill_between, fill_betweenx, stacked fills, and translucent overlapping areas.", WebDemoID: "fills", Showcase: true, Width: 840, Height: 620, MaxMeanAbs: 1.5, MaxRMSE: 1.2},
	{ID: "errorbar_basic", Topic: "errorbar", Title: "Error Bars", Description: "Symmetric and asymmetric error bars with caps, marker styling, and legend output.", WebDemoID: "errorbars", Showcase: true, SVGGoldenFamily: "errorbar", GoBasicSmokeFamily: "errorbar", SkiaParityFamily: "errorbar", MaxMeanAbs: 0.06, MaxRMSE: 0.16},
	{ID: "multi_series_basic", Topic: "multi", Title: "Multiple Series", Description: "Several labeled lines sharing one axes, demonstrating color cycling and legends.", Showcase: true, SkiaParityFamily: "line", MaxRMSE: 0.8},
	{ID: "multi_series_color_cycle", Topic: "multi", Title: "Color Cycle", MaxRMSE: 0.6},
	// MaxRMSE 3.3: MeanAbs 0.12 / PSNR ~54 dB; residual is sub-pixel text/handle edge antialiasing under the 3.10.9 reference set.
	{ID: "legend_layout_matrix", Topic: "legend", Title: "Legend Layout Matrix", FixtureOnly: true, MaxRMSE: 3.3, MaxDiffPixels: 3300, MaxLargestCluster: 2900},
	{ID: "text_annotation_matrix", Topic: "annotation", Title: "Text Annotation Matrix", FixtureOnly: true, MaxRMSE: 4.2, MaxDiffPixels: 2800, MaxLargestCluster: 380},
	{ID: "hist_basic", Topic: "histogram", Title: "Histogram Counts", Description: "A deterministic histogram with count bins, labels, and default bar styling.", Showcase: true, SVGGoldenFamily: "hist", GoBasicSmokeFamily: "histogram", SkiaParityFamily: "histogram", MaxRMSE: 1.1},
	{ID: "hist_log", Topic: "histogram", Title: "Logarithmic Histogram", Description: "A count histogram on a logarithmic y axis, matching matplotlib hist(log=True).", Showcase: true, MaxRMSE: 1.1},
	{ID: "hist_density", Topic: "histogram", Title: "Histogram Density", MaxRMSE: 1.3},
	{ID: "hist_strategies", Topic: "histogram", Title: "Histogram Strategies", MaxRMSE: 1.608},
	{ID: "histogram_variants", Topic: "histogram", Title: "Histogram Variants Gallery", Description: "A combined gallery of count, density, cumulative, and overlapping probability histograms over deterministic samples.", WebDemoID: "histogram", Showcase: true, Width: 840, Height: 620, MaxMeanAbs: 1.5, MaxRMSE: 1.433},
	{ID: "boxplot_basic", Topic: "boxplot", Title: "Box Plot", Description: "Grouped box plots with whiskers, medians, outliers, and categorical labels.", Optional: true, Showcase: true, MaxMeanAbs: 2.0, MaxRMSE: 1.608},
	{ID: "boxplot_default", Topic: "boxplot", Title: "Box Plot Default Styling", Description: "Box plots with Matplotlib default styling: unfilled boxes, C1 medians, and unfilled-circle fliers (patch_artist=False).", Optional: true, MaxMeanAbs: 2.0, MaxRMSE: 1.608},
	{ID: "boxplot_single_series", Topic: "boxplot", Title: "Single-Series Box Plot", Description: "A focused parity fixture for the direct single-series BoxPlot API.", FixtureOnly: true, MaxMeanAbs: 0.02, MaxRMSE: 0.5},
	{ID: "axes_convenience_helpers", Topic: "axes", Title: "Axes Convenience Helpers", Description: "Precomputed bxp and violin stats, hlines/vlines broadcasting, and post-hoc clabel convenience helpers.", FixtureOnly: true, MaxRMSE: 2.1},
	{ID: "axes_log_plot_wrappers", Topic: "axes", Title: "Log Plot Wrappers", Description: "Focused visual coverage for SemilogX, SemilogY, and LogLog convenience plotting.", FixtureOnly: true, Width: 960, Height: 360, MaxMeanAbs: 0.02, MaxRMSE: 0.08},
	{ID: "axes_option_breadth", Topic: "axes", Title: "Axes Option Breadth", Description: "Scatter scalar styling, edge-aligned stacked bars with labels, fill-between masks/steps/interpolation, and errorevery sampling.", FixtureOnly: true, MaxRMSE: 0.6},
	{ID: "axes_secondary_y_twiny", Topic: "axes", Title: "Twin and Secondary Y Axes", Description: "Focused visual coverage for TwinY shared-y overlays and transformed SecondaryYAxis ticks.", FixtureOnly: true, Width: 760, Height: 400, MaxMeanAbs: 0.05, MaxRMSE: 0.8},
	{ID: "text_labels_strict", Topic: "text", Title: "Strict Text Labels", Optional: true, SVGGoldenFamily: "text_layout", GoBasicSmokeFamily: "text", SkiaParityFamily: "text", MaxRMSE: 0.3},
	{ID: "title_strict", Topic: "text", Title: "Strict Title", MaxRMSE: 0.3},
	{ID: "mathtext_basic", Topic: "mathtext", Title: "MathText Basic", FixtureOnly: true, SVGGoldenFamily: "mathtext", GoBasicSmokeFamily: "mathtext", SkiaParityFamily: "mathtext", MaxRMSE: 2.8, MaxDiffPixels: 300, MaxLargestCluster: 80},
	{ID: "mathtext_fractions", Topic: "mathtext", Title: "MathText Fractions", FixtureOnly: true, SkiaParityFamily: "mathtext", MaxRMSE: 4.5, MaxDiffPixels: 110, MaxLargestCluster: 45},
	{ID: "mathtext_integrals", Topic: "mathtext", Title: "MathText Operators", FixtureOnly: true, SkiaParityFamily: "mathtext", MaxRMSE: 0.3},
	{ID: "mathtext_matrices", Topic: "mathtext", Title: "MathText Matrices", FixtureOnly: true, SkiaParityFamily: "mathtext", MaxRMSE: 0.5},
	{ID: "mathtext_inline_labels", Topic: "mathtext", Title: "MathText Inline Labels", FixtureOnly: true, SkiaParityFamily: "mathtext", MaxRMSE: 1.608},
	{ID: "mathtext_accents", Topic: "mathtext", Title: "MathText Accents", FixtureOnly: true, SkiaParityFamily: "mathtext", MaxRMSE: 1.608, MaxDiffPixels: 10, MaxLargestCluster: 5},
	{ID: "mathtext_gallery", Topic: "mathtext", Title: "MathText Gallery", Description: "Fractions, roots, operators, fences, matrices, and inline MathText labels in one browsable figure.", Showcase: true, Width: 900, Height: 560, MaxRMSE: 2.6, MaxDiffPixels: 210, MaxLargestCluster: 110},
	{ID: "text_layout_gallery", Topic: "text", Title: "Text Layout Gallery", Description: "Alignment, rotation, multiline layout, wrapping, and text bbox styling in one gallery.", Showcase: true, Width: 900, Height: 560, MaxMeanAbs: 1.0, MaxRMSE: 5.0},
	{ID: "text_bbox_styles", Topic: "text", Title: "Text BBox Styles", Description: "Every FancyBboxPatch boxstyle (square, round, round4, circle, ellipse, sawtooth, roundtooth, arrows) applied behind text via the Matplotlib boxstyle spec bridge.", FixtureOnly: true, Width: 640, Height: 360, MaxRMSE: 1.608},
	{ID: "image_heatmap", Topic: "image", Title: "Heatmap Image", Description: "A gridded image plot with a colorbar and axis labels for matrix-style data.", WebDemoID: "heatmap", Showcase: true, SVGGoldenFamily: "image", GoBasicSmokeFamily: "image", SkiaParityFamily: "image", MaxRMSE: 1.3, MaxDiffPixels: 340, MaxLargestCluster: 340},
	{ID: "image_variants_gallery", Topic: "image", Title: "Image Variants Gallery", Description: "Side-by-side interpolation modes, alpha image overlays, MatShow ticks, and spy marker/image modes.", Showcase: true, Width: 1080, Height: 720, MaxRMSE: 2.7},
	{ID: "imshow_clipped", Topic: "image", Title: "Clipped Imshow", FixtureOnly: true, MaxRMSE: 3.6, MaxDiffPixels: 2900, MaxLargestCluster: 2900},
	{ID: "imshow_rgb", Topic: "image", Title: "RGB/RGBA Imshow", FixtureOnly: true, MaxRMSE: 1.608},
	{ID: "imshow_transformed", Topic: "image", Title: "Transformed Imshow", FixtureOnly: true, Width: 420, Height: 420, MaxRMSE: 2.9},
	{ID: "imshow_bilinear", Topic: "image", Title: "Bilinear Imshow", FixtureOnly: true, Width: 256, Height: 256, MaxRMSE: 1.2},
	{ID: "imshow_bicubic", Topic: "image", Title: "Bicubic Imshow", FixtureOnly: true, Width: 256, Height: 256, MaxRMSE: 1.6},
	{ID: "imshow_interpolation_matrix", Topic: "image", Title: "Imshow Interpolation Matrix", FixtureOnly: true, Width: 800, Height: 480, MaxRMSE: 3.3},
	{ID: "image_alpha", Topic: "image", Title: "Image Alpha", FixtureOnly: true, MaxRMSE: 0.7},
	{ID: "matshow_basic", Topic: "image", Title: "Matshow", FixtureOnly: true, MaxRMSE: 1.608, MaxDiffPixels: 25, MaxLargestCluster: 25},
	{ID: "spy_marker", Topic: "image", Title: "Spy Marker Mode", FixtureOnly: true, MaxRMSE: 0.5},
	{ID: "spy_image", Topic: "image", Title: "Spy Image Mode", FixtureOnly: true, MaxRMSE: 1.608},
	{ID: "colormap_diverging", Topic: "colormap", Title: "Diverging Colormap", FixtureOnly: true, MaxRMSE: 1.608, MaxDiffPixels: 330, MaxLargestCluster: 330},
	{ID: "colormap_qualitative", Topic: "colormap", Title: "Qualitative Colormap", FixtureOnly: true, MaxRMSE: 0.3},
	{ID: "colormap_cyclic", Topic: "colormap", Title: "Cyclic Colormap", FixtureOnly: true, MaxRMSE: 1.0},
	{ID: "colormap_families_gallery", Topic: "colormap", Title: "Colormap Family Gallery", Description: "Sequential, reversed, perceptual, diverging, qualitative, and cyclic colormap strips in one browsable figure.", Showcase: true, Width: 900, Height: 520, MaxRMSE: 0.5},
	{ID: "named_colors", Topic: "color", Title: "Named Colors", FixtureOnly: true, MaxMeanAbs: 0.60, MaxRMSE: 0.904},
	{ID: "named_colors_gallery", Topic: "color", Title: "Named Color Swatches", Description: "CSS4 names, Tableau tab colors, xkcd names, shorthand colors, grayscale strings, hex values, and RGBA tuples.", Showcase: true, Width: 900, Height: 520, MaxRMSE: 1.0},
	{ID: "axes_top_right_inverted", Topic: "axes", Title: "Top/Right Inverted Axes", Optional: true, MaxRMSE: 0.8},
	{ID: "axline_slope", Topic: "axes", Title: "Slope-Defined AxLine", Description: "A focused parity fixture for an infinite line defined by a point and slope.", FixtureOnly: true, MaxMeanAbs: 0.02, MaxRMSE: 0.6},
	// Twinned-axes frame ratchet: twinx/twiny now keep foreground frame spines
	// visible like Matplotlib; refreshed golden-vs-reference RMSE is 2.90.
	{ID: "axes_control_surface", Topic: "axes", Title: "Axes, Scales, and Twins", Optional: true, WebDemoID: "axes", Description: "Minor ticks, top/right axes, aspect controls, log scale, twin axes, and secondary axes.", Showcase: true, GoBasicSmokeFamily: "axes", MaxRMSE: 3.0, MaxDiffPixels: 1300, MaxLargestCluster: 360},
	{ID: "transform_coordinates", Topic: "axes", Title: "Transform Coordinates", Optional: true, MaxRMSE: 2.4, MaxDiffPixels: 330, MaxLargestCluster: 290},
	{ID: "transform_annotation_modes", Topic: "axes", Title: "Annotation Coordinate Modes", FixtureOnly: true, Width: 720, Height: 420, MaxRMSE: 2.9, MaxDiffPixels: 510, MaxLargestCluster: 180},
	{ID: "path_clipped_transformed", Topic: "axes", Title: "Clipped Transformed Path", FixtureOnly: true, Width: 720, Height: 420, MaxRMSE: 1.608},
	{ID: "layout_bbox_helpers", Topic: "axes", Title: "Layout BBox Helpers", FixtureOnly: true, Width: 720, Height: 420, MaxRMSE: 1.1},
	{ID: "formatter_engineering_labels", Topic: "axes", Title: "Engineering Formatter Labels", FixtureOnly: true, Width: 720, Height: 400, MaxRMSE: 3.0, MaxDiffPixels: 150, MaxLargestCluster: 35},
	{ID: "formatter_fixed_null_labels", Topic: "axes", Title: "Fixed and Null Formatter Labels", FixtureOnly: true, MaxRMSE: 0.7},
	{ID: "formatter_log_mathtext_labels", Topic: "axes", Title: "Log MathText Formatter Labels", FixtureOnly: true, MaxMeanAbs: 0.08, MaxRMSE: 0.303},
	{ID: "formatter_percent_labels", Topic: "axes", Title: "Percent Formatter Labels", FixtureOnly: true, MaxRMSE: 0.4},
	{ID: "formatter_scalar_scientific_labels", Topic: "axes", Title: "Scalar Scientific Formatter Labels", FixtureOnly: true, MaxRMSE: 3.8, MaxDiffPixels: 90, MaxLargestCluster: 88},
	{ID: "locator_fixed_index_labels", Topic: "axes", Title: "Fixed and Index Locator Labels", FixtureOnly: true, Width: 720, Height: 420, MaxMeanAbs: 0.05, MaxRMSE: 0.08},
	{ID: "locator_linear_labels", Topic: "axes", Title: "Linear Locator Labels", FixtureOnly: true, Width: 720, Height: 540, MaxMeanAbs: 0.05, MaxRMSE: 0.08},
	{ID: "locator_log_minor_threshold_labels", Topic: "axes", Title: "Log Locator Minor Labels", FixtureOnly: true, Width: 720, Height: 420, MaxMeanAbs: 0.20, MaxRMSE: 1.278},
	{ID: "locator_maxn_edge_labels", Topic: "axes", Title: "MaxN Locator Edge Labels", FixtureOnly: true, Width: 720, Height: 540, MaxMeanAbs: 0.02, MaxRMSE: 0.08},
	{ID: "scale_asinh_ticks", Topic: "axes", Title: "Asinh Scale Ticks", FixtureOnly: true, Width: 720, Height: 400, MaxMeanAbs: 0.05, MaxRMSE: 2.5},
	{ID: "scale_function_defaults", Topic: "axes", Title: "Function Scale Defaults", FixtureOnly: true, Width: 720, Height: 480, MaxMeanAbs: 0.05, MaxRMSE: 0.202},
	{ID: "scale_logit_ticks", Topic: "axes", Title: "Logit Scale Ticks", FixtureOnly: true, Width: 720, Height: 400, MaxMeanAbs: 0.10, MaxRMSE: 0.36},
	{ID: "scale_symlog_ticks", Topic: "axes", Title: "Symlog Scale Ticks", FixtureOnly: true, Width: 720, Height: 400, MaxMeanAbs: 0.05, MaxRMSE: 0.16},
	{ID: "spine_positions", Topic: "axes", Title: "Spine Position Modes", Description: "Axes-fraction and outward-point spine positions with their attached ticks and labels.", FixtureOnly: true, Width: 720, Height: 400, MaxRMSE: 1.0},
	{ID: "ticks_styling_surface", Topic: "axes", Title: "Tick Styling Surface", FixtureOnly: true, Width: 720, Height: 420, MaxMeanAbs: 0.30, MaxRMSE: 3.0},
	// Axes-patch draw-order ratchet: regenerated golden-vs-reference RMSE is 2.80.
	{ID: "ticks_scales_formatters_gallery", Topic: "axes", Title: "Ticks, Scales, and Formatters Gallery", Description: "A focused gallery covering major and minor locators, log and signed scales, formatter families, date labels, category labels, and custom units.", Optional: true, Width: 1320, Height: 900, Showcase: true, MaxMeanAbs: 0.10, MaxRMSE: 2.9},
	{ID: "artist_metadata", Topic: "artist", Title: "Artist Metadata", FixtureOnly: true, MaxRMSE: 2.2},
	{ID: "gridspec_composition", Topic: "composition", Title: "Figure Composition", WebDemoID: "composition", Description: "GridSpec spans, figure-level labels, figure legends, anchored text, and colorbars.", Showcase: true, GoBasicSmokeFamily: "layout", MaxRMSE: 0.8},
	{ID: "figure_labels_composition", Topic: "composition", Title: "Figure Labels", Description: "A multi-axes figure with shared figure title, x label, y label, and legend placement.", WebDemoID: "subplots", Showcase: true, MaxRMSE: 1.608},
	{ID: "colorbar_composition", Topic: "colorbar", Title: "Colorbar Composition", Description: "A composed figure that exercises image color mapping, shared colorbars, and layout spacing.", Showcase: true, GoBasicSmokeFamily: "colorbar", MaxRMSE: 0.8},
	{ID: "colorbar_variants_gallery", Topic: "colorbar", Title: "Colorbar Variants Gallery", Description: "LogNorm, TwoSlopeNorm, BoundaryNorm draw edges, and under/over extension colorbars with labels.", WebDemoID: "colorbars", Showcase: true, Width: 1040, Height: 720, MaxRMSE: 4.4},
	{ID: "annotation_composition", Topic: "annotation", Title: "Annotations", Description: "Text annotations, arrows, anchored labels, and mixed coordinate placement.", Showcase: true, GoBasicSmokeFamily: "annotation", MaxRMSE: 1.608},
	{ID: "annotation_legend_offsetbox_gallery", Topic: "annotation", Title: "Annotation, Legend, and Offset Box Gallery", Description: "Annotation arrows, bbox annotations, multi-column legends, proxy handles, anchored text, drawing areas, packers, image boxes, and size bars.", WebDemoID: "annotations", Showcase: true, Width: 1040, Height: 720, MaxRMSE: 5.0, MaxDiffPixels: 5100, MaxLargestCluster: 1800},
	{ID: "patch_showcase", Topic: "patches", Title: "Patch Showcase", Description: "Rectangles, circles, ellipses, polygons, path patches, fancy arrows, fancy boxes, hatches, alpha, fills, and strokes.", Optional: true, WebDemoID: "patches", Showcase: true, SVGGoldenFamily: "hatch_bars", GoBasicSmokeFamily: "patch_hatch", SkiaParityFamily: "patch_hatch", MaxRMSE: 1.608},
	{ID: "patch_style_matrix", Topic: "patches", Title: "Patch Style Matrix", FixtureOnly: true, MaxRMSE: 4.2, MaxDiffPixels: 3700, MaxLargestCluster: 83},
	{ID: "mesh_contour_tri", Topic: "mesh", Title: "Meshes and Contours", Optional: true, WebDemoID: "mesh", Description: "PColorMesh, contour/contourf, Hist2D, triplot, tripcolor, and tricontour.", Showcase: true, GoBasicSmokeFamily: "mesh", SkiaParityFamily: "mesh", MaxRMSE: 2.6},
	{ID: "contour_styles", Topic: "mesh", Title: "Contour Styles", Optional: true, Description: "Monochrome contour with negative_linestyles dashing and clabel format-string labels; contourf with extend and cycled hatches.", Width: 640, Height: 360, MaxRMSE: 2.8},
	// Axis-below ratchet: showcase grids now mirror Matplotlib set_axisbelow(True);
	// refreshed golden-vs-reference RMSE is 2.29.
	{ID: "plot_variants", Topic: "variants", Title: "Plot Variants", Optional: true, WebDemoID: "variants", Description: "Step, stairs, reference lines, spans, broken bars, and stacked bars.", Showcase: true, GoBasicSmokeFamily: "variants", MaxRMSE: 2.6},
	// Phase 2 ratchet: off-bin fixture signal avoids undefined phase residues;
	// regenerated golden-vs-reference RMSE is 1.15.
	{ID: "spectrum_variants", Topic: "signal", Title: "Spectrum Variants", FixtureOnly: true, GoBasicSmokeFamily: "signal", MaxMeanAbs: 0.10, MaxRMSE: 1.5},
	{ID: "psd_welch", Topic: "signal", Title: "Welch Power Spectral Density", Description: "Direct PSD coverage with overlapping Hann-windowed, mean-detrended segments and zero padding.", FixtureOnly: true, Width: 640, Height: 360, MaxMeanAbs: 0.05, MaxRMSE: 0.227},
	{ID: "specgram_psd", Topic: "signal", Title: "PSD Spectrogram", Description: "Direct Specgram coverage for an overlapping-window chirp rendered in decibels.", FixtureOnly: true, Width: 640, Height: 360, MaxMeanAbs: 0.05, MaxRMSE: 1.2, MaxDiffPixels: 550, MaxLargestCluster: 450},
	{ID: "cohere_welch", Topic: "signal", Title: "Welch Coherence", Description: "Direct Cohere coverage for partially shared deterministic signals.", FixtureOnly: true, Width: 640, Height: 360, MaxMeanAbs: 0.05, MaxRMSE: 0.227},
	{ID: "csd_welch", Topic: "signal", Title: "Welch Cross-Spectral Density", Description: "Direct CSD coverage with overlapping Hann-windowed, mean-detrended segments and zero padding.", FixtureOnly: true, Width: 640, Height: 360, MaxMeanAbs: 0.05, MaxRMSE: 0.227},
	{ID: "stat_variants", Topic: "statistics", Title: "Statistical Views", Optional: true, WebDemoID: "statistics", Description: "Box plots, violin plots, empirical CDFs, and stack plots.", Showcase: true, GoBasicSmokeFamily: "statistics", MaxMeanAbs: 0.35, MaxRMSE: 3.4, MaxDiffPixels: 1400, MaxLargestCluster: 530},
	{ID: "specialty_depth", Topic: "statistics", Title: "Specialty Depth", FixtureOnly: true, MaxRMSE: 1.608},
	{ID: "stem_plot", Topic: "specialty", Title: "Stem Plot", Optional: true, MaxRMSE: 1.608},
	// Phase 9 "misc artist kwargs" parity fixtures.
	{ID: "stem_horizontal", Topic: "specialty", Title: "Horizontal Stem", Description: "Stem plot with orientation=\"horizontal\": locs along y, heads along x from a vertical baseline.", FixtureOnly: true, MaxMeanAbs: 2.0, MaxRMSE: 1.608},
	{ID: "errorbar_capthick", Topic: "errorbar", Title: "Error Bars (capthick)", Description: "Error bars whose caps use a thicker line than the bars via the capthick kwarg.", FixtureOnly: true, MaxMeanAbs: 2.0, MaxRMSE: 1.608},
	{ID: "scatter_plotnonfinite", Topic: "scatter", Title: "Scatter (plotnonfinite)", Description: "Scatter with plotnonfinite=True: non-finite scalar values ride the colormap bad color while finite points map through viridis.", FixtureOnly: true, MaxMeanAbs: 2.0, MaxRMSE: 1.608},
	{ID: "linecollection_linestyle", Topic: "specialty", Title: "Line Styles", Description: "Per-item LineCollection string linestyles (solid/dashed/dashdot/dotted) via hlines.", FixtureOnly: true, MaxMeanAbs: 2.0, MaxRMSE: 1.608},
	// Table cell text now follows Matplotlib's Text advance alignment; refreshed
	// golden-vs-reference RMSE is 1.74.
	{ID: "specialty_artists", Topic: "specialty", Title: "Specialty Artists", Optional: true, WebDemoID: "specialty", Description: "Event plots, hexbin, pie charts, stem plots, tables, and Sankey-style flows.", Showcase: true, GoBasicSmokeFamily: "specialty", MaxRMSE: 2.0},
	// W3.10 ratchet: X-axis concise-date offset text now follows Matplotlib's
	// ticklabel-bounds placement rule, reducing the committed RMSE to 0.05.
	{ID: "date_concise_intraday_labels", Topic: "units", Title: "Concise Intraday Date Labels", FixtureOnly: true, Width: 720, Height: 360, MaxMeanAbs: 0.01, MaxRMSE: 0.057},
	{ID: "date_month_year_labels", Topic: "units", Title: "Month and Year Date Labels", FixtureOnly: true, Width: 720, Height: 360, MaxMeanAbs: 0.05, MaxRMSE: 0.08},
	{ID: "units_overview", Topic: "units", Title: "Dates and Categories", Optional: true, WebDemoID: "units", Description: "Time-aware axes, categorical bars, and horizontal categorical bars.", Showcase: true, GoBasicSmokeFamily: "units", MaxMeanAbs: 0.10, MaxRMSE: 0.101},
	{ID: "units_dates", Topic: "units", Title: "Date Units", Optional: true, MaxMeanAbs: 0.15, MaxRMSE: 1.5},
	{ID: "units_categories", Topic: "units", Title: "Category Units", Optional: true, MaxMeanAbs: 0.05, MaxRMSE: 2.0},
	{ID: "units_custom_converter", Topic: "units", Title: "Custom Unit Converter", Optional: true, MaxMeanAbs: 0.05, MaxRMSE: 0.101},
	{ID: "vector_fields", Topic: "vectors", Title: "Vector Fields", Optional: true, WebDemoID: "vectors", Description: "Quiver, quiver keys, barbs, streamplots, and grid-based vector input.", Showcase: true, GoBasicSmokeFamily: "vectors", MaxRMSE: 2.3},
	{ID: "quiver_autoscale", Topic: "vectors", Title: "Quiver Default Scale", FixtureOnly: true, Width: 640, Height: 360, MaxRMSE: 2.0},
	{ID: "polar_axes", Topic: "polar", Title: "Polar Wave", WebDemoID: "polar", Description: "A filled polar curve with custom radial and angular grid styling.", Showcase: true, SVGGoldenFamily: "clipped_polar", GoBasicSmokeFamily: "polar", SkiaParityFamily: "polar", MaxRMSE: 1.2},
	{ID: "geo_mollweide_axes", Topic: "geo", Title: "Projections and Insets", WebDemoID: "projections", Description: "Mollweide geo projection plus a zoomed inset axes.", Showcase: true, GoBasicSmokeFamily: "geo", MaxRMSE: 3.0},
	{ID: "geo_aitoff_axes", Topic: "geo", Title: "Aitoff Projection", Description: "An Aitoff equal-area projection with longitude wrapping and graticule rendering.", Optional: true, Showcase: true, MaxRMSE: 3.0},
	{ID: "geo_hammer_axes", Topic: "geo", Title: "Hammer Projection", Optional: true, MaxRMSE: 3.0},
	// MaxRMSE 1.3: MeanAbs 0.04 / PSNR ~60 dB; residual is a few sub-pixel graticule/label edge pixels under the 3.10.9 reference set.
	{ID: "geo_lambert_axes", Topic: "geo", Title: "Lambert Projection", Optional: true, MaxRMSE: 2.2},
	// MaxRMSE 0.6: MeanAbs 0.09 / PSNR ~58 dB; residual is isolated antialiased edge pixels on the closed radar polygons.
	{ID: "radar_basic", Topic: "radar", Title: "Radar Projection", Description: "A radar chart using polar projection plumbing with closed polygon series.", Optional: true, Showcase: true, MaxMeanAbs: 2.0, MaxRMSE: 0.9},
	{ID: "skewt_basic", Topic: "skewt", Title: "Skew-T Projection", Description: "A meteorological-style skew-T axes with transformed temperature grid lines.", Optional: true, Showcase: true, MaxRMSE: 1.608},
	{ID: "projection_toolkit_gallery", Topic: "projections", Title: "Projection and Toolkit Gallery", Description: "A grouped gallery covering polar, Mollweide, Aitoff, Hammer, Lambert, radar, skew-T, axisartist-style twin axes, and axes_grid1-style image grids.", Optional: true, WebDemoID: "toolkit", Width: 1320, Height: 900, Showcase: true, MaxRMSE: 4.9},
	{ID: "mplot3d_basic", Topic: "mplot3d", Title: "3D Toolkit Scaffold", Optional: true, GoBasicSmokeFamily: "mplot3d", MaxRMSE: 1.608},
	{ID: "mplot3d_terrain", Topic: "mplot3d", Title: "3D Terrain", Description: "A 3D surface terrain plot with depth sorting, pane styling, and projected axes.", Optional: true, Width: 900, Height: 640, Showcase: true, MaxRMSE: 1.3},
	{ID: "mplot3d_gallery", Topic: "mplot3d", Title: "mplot3d Gallery", Description: "A broad mplot3d gallery covering 3D lines, scatter, surface, wireframe, trisurf, bar3d, voxels, quiver3d, stem3d, and fill-between3d.", Optional: true, WebDemoID: "mplot3d", Width: 1320, Height: 840, Showcase: true, MaxRMSE: 1.608},
	{ID: "mplot3d_plot3d", Topic: "mplot3d", Title: "3D Plot", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.4},
	{ID: "mplot3d_scatter3d", Topic: "mplot3d", Title: "3D Scatter", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.608},
	{ID: "mplot3d_surface3d", Topic: "mplot3d", Title: "3D Surface", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.1},
	{ID: "mplot3d_wire3d", Topic: "mplot3d", Title: "3D Wireframe", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.5},
	{ID: "mplot3d_trisurf3d", Topic: "mplot3d", Title: "3D Triangulated Surface", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.608},
	{ID: "mplot3d_bar3d", Topic: "mplot3d", Title: "3D Bars", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.608},
	{ID: "mplot3d_voxels", Topic: "mplot3d", Title: "3D Voxels", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.608},
	{ID: "mplot3d_quiver3d", Topic: "mplot3d", Title: "3D Quiver", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.608},
	{ID: "mplot3d_contour3d", Topic: "mplot3d", Title: "3D Contour", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.608},
	{ID: "mplot3d_contourf3d", Topic: "mplot3d", Title: "3D Filled Contour", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.2},
	{ID: "mplot3d_tricontour3d", Topic: "mplot3d", Title: "3D Triangulated Contour", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 2.8},
	{ID: "mplot3d_tricontourf3d", Topic: "mplot3d", Title: "3D Triangulated Filled Contour", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.608},
	{ID: "mplot3d_bar2d_zdir", Topic: "mplot3d", Title: "3D Planar 2D Bars", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.2},
	{ID: "mplot3d_text3d", Topic: "mplot3d", Title: "3D Text Labels", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.608},
	{ID: "mplot3d_errorbar3d", Topic: "mplot3d", Title: "3D Error Bars", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 3.0},
	{ID: "mplot3d_stem3d", Topic: "mplot3d", Title: "3D Stem", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.608},
	{ID: "mplot3d_fill_between3d", Topic: "mplot3d", Title: "3D Fill Between", FixtureOnly: true, Width: 720, Height: 560, MaxRMSE: 1.608},
	{ID: "unstructured_showcase", Topic: "unstructured", Title: "Unstructured Showcase", Description: "Triangulated and irregular data views covering triplot, tripcolor, and contour variants.", Optional: true, Showcase: true, GoBasicSmokeFamily: "unstructured", MaxRMSE: 3.2, MaxDiffPixels: 3900, MaxLargestCluster: 490},
	{ID: "triangulation_gallery", Topic: "unstructured", Title: "Triangulation Gallery", Description: "A focused triangulation gallery covering triplot, tripcolor, tricontour, tricontourf, and masked pcolormesh behavior.", Optional: true, WebDemoID: "triangulation", Width: 1320, Height: 760, Showcase: true, MaxRMSE: 2.9},
	{ID: "arrays_showcase", Topic: "arrays", Title: "Matrix Helpers", Optional: true, WebDemoID: "matrix", Description: "MatShow, sparsity spy plots, annotated heatmaps, and colorbars.", Width: 1240, Height: 620, Showcase: true, GoBasicSmokeFamily: "matrix", MaxRMSE: 3.1},
	{ID: "widgets_gallery", Topic: "widgets", Title: "Widgets and Selectors", Description: "Buttons, sliders, range sliders, text boxes, check and radio buttons, selectors, cursor, and multi-cursor helpers.", Optional: true, Width: 1100, Height: 760, Showcase: true, GoBasicSmokeFamily: "widgets", MaxRMSE: 5.0},
	{ID: "animation_gallery", Topic: "animation", Title: "Animation Playback", Description: "FuncAnimation and ArtistAnimation setup with repeat, repeat-delay, blit, and explicit unsupported writer behavior.", Optional: true, WebDemoID: "animation", Width: 1000, Height: 560, Showcase: true, GoBasicSmokeFamily: "animation", MaxRMSE: 1.5},
	{ID: "animation_line_frame", Topic: "animation", Title: "Animated Line Frame", Description: "Deterministic frame-N capture of a traveling-wave FuncAnimation line plot.", FixtureOnly: true, Width: 640, Height: 360, MaxMeanAbs: 1.5, MaxRMSE: 1.433},
	{ID: "animation_scatter_frame", Topic: "animation", Title: "Animated Scatter Frame", Description: "Deterministic frame-N capture of an orbiting scatter/collection animation with scalar-mapped marker colors.", FixtureOnly: true, Width: 640, Height: 360, MaxMeanAbs: 1.5, MaxRMSE: 1.433},
	{ID: "animation_imshow_frame", Topic: "animation", Title: "Animated Heatmap Frame", Description: "Deterministic frame-N capture of a ripple imshow (heatmap) animation.", FixtureOnly: true, Width: 640, Height: 360, MaxMeanAbs: 1.5, MaxRMSE: 1.433},
	{ID: "animation_subplots_frame", Topic: "animation", Title: "Animated Subplots Frame", Description: "Deterministic frame-N capture of a two-panel line + heatmap subplot composition animation.", FixtureOnly: true, Width: 800, Height: 360, MaxMeanAbs: 1.5, MaxRMSE: 4.5, MaxDiffPixels: 230, MaxLargestCluster: 88},
	{ID: "axisartist_showcase", Topic: "axisartist", Title: "AxisArtist Showcase", Description: "Floating and axis-artist style axes with custom spine placement and labels.", Optional: true, WebDemoID: "axisartist", Showcase: true, GoBasicSmokeFamily: "axisartist", MaxRMSE: 2.3},
	{ID: "axes_grid1_showcase", Topic: "axes_grid1", Title: "Axes Grid1 Showcase", Description: "Axes divider, image-grid, inset, and anchored layout helpers in one composition.", Optional: true, WebDemoID: "axes_grid1", Showcase: true, GoBasicSmokeFamily: "axes_grid1", MaxMeanAbs: 0.16, MaxRMSE: 2.0},
	{ID: "pcolor_flat", Topic: "mesh", Title: "PColor Flat", FixtureOnly: true, MaxRMSE: 0.7},
	{ID: "pcolormesh_nearest", Topic: "mesh", Title: "PColorMesh Nearest", FixtureOnly: true, MaxRMSE: 0.8},
	{ID: "pcolormesh_gouraud", Topic: "mesh", Title: "PColorMesh Gouraud", FixtureOnly: true, SkiaParityFamily: "gouraud", MaxRMSE: 1.0},
	{ID: "pcolormesh_masked", Topic: "mesh", Title: "PColorMesh Masked", FixtureOnly: true, MaxRMSE: 0.7},
	{ID: "hist2d_weighted_density", Topic: "mesh", Title: "Hist2D Weighted Density", FixtureOnly: true, MaxRMSE: 0.9},
	{ID: "asinh_norm_image", Topic: "colorbar", Title: "AsinhNorm Image", FixtureOnly: true, MaxRMSE: 2.1},
	{ID: "boundarynorm_pcolormesh", Topic: "colorbar", Title: "BoundaryNorm PColorMesh", FixtureOnly: true, MaxRMSE: 0.9},
	{ID: "collection_mutable_scalarmap", Topic: "colorbar", Title: "Mutable Collection ScalarMap", FixtureOnly: true, MaxRMSE: 0.9},
	{ID: "colorbar_boundary_values", Topic: "colorbar", Title: "Boundary Colorbar Values", FixtureOnly: true, MaxRMSE: 3.1, MaxDiffPixels: 470, MaxLargestCluster: 380},
	{ID: "colorbar_horizontal_ticks", Topic: "colorbar", Title: "Horizontal Colorbar Ticks", FixtureOnly: true, MaxRMSE: 1.2},
	{ID: "lognorm_imshow", Topic: "colorbar", Title: "LogNorm Imshow", FixtureOnly: true, MaxRMSE: 4.4},
	{ID: "twoslope_norm_image", Topic: "colorbar", Title: "TwoSlopeNorm Image", FixtureOnly: true, MaxRMSE: 1.1},
	{ID: "colorbar_extensions", Topic: "colorbar", Title: "Colorbar Extensions", FixtureOnly: true, MaxRMSE: 1.1},
	{ID: "colorbar_extendfrac", Topic: "colorbar", Title: "Colorbar Extend Fraction", FixtureOnly: true, MaxMeanAbs: 2.0, MaxRMSE: 1.6},
	{ID: "colorbar_symlog_ticks", Topic: "colorbar", Title: "SymLogNorm Colorbar", FixtureOnly: true, MaxMeanAbs: 2.0, MaxRMSE: 1.6},
	{ID: "large_scatter", Topic: "raster", Title: "Large Scatter Batch", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"pathcollectionbatch"}, MaxMeanAbs: 0.5, MaxRMSE: 2.7},
	{ID: "mixed_collection", Topic: "raster", Title: "Mixed Path Collection", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"pathcollectionbatch"}, SVGGoldenFamily: "collection", GoBasicSmokeFamily: "collection", MaxMeanAbs: 0.5, MaxRMSE: 1.7},
	{ID: "mixed_raster_vector", Topic: "raster", Title: "Mixed Raster Vector Output", Description: "A polar mixed-output example with rasterized dense scatter points and vector-preserved line, text, axes, legend, SVG, and PDF artifacts.", Showcase: true, Width: 640, Height: 640, SVGGoldenFamily: "mixed_raster", MaxRMSE: 1.608},
	{ID: "quad_mesh", Topic: "raster", Title: "Quad Mesh Batch", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"quadmeshbatch"}, MaxMeanAbs: 1.0, MaxRMSE: 1.6, MaxDiffPixels: 88, MaxLargestCluster: 88},
	{ID: "gouraud_triangles", Topic: "raster", Title: "Gouraud Triangles", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"gouraudtrianglebatch"}, SkiaParityFamily: "gouraud", MaxMeanAbs: 1.0, MaxRMSE: 0.6},
	{ID: "clip_path_batch", Topic: "raster", Title: "Clip Path Batch", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"pathclip", "quadmeshbatch"}, MaxMeanAbs: 1.0, MaxRMSE: 1.433},
}

// Cases returns every cataloged parity example/fixture.
func Cases() []Case {
	out := make([]Case, len(cases))
	copy(out, cases)
	for i := range out {
		out[i].NativeCapabilities = append([]string(nil), out[i].NativeCapabilities...)
		applyDefaults(&out[i])
	}
	return out
}

// Lookup finds a parity case by canonical case ID.
func Lookup(id string) (Case, bool) {
	all := Cases()
	for i := range all {
		if all[i].ID == id {
			return all[i], true
		}
	}
	return Case{}, false
}

// WebDemos returns the curated web-demo subset in display order.
func WebDemos() []Case {
	all := Cases()
	var out []Case
	for i := range all {
		if all[i].WebDemoID != "" {
			out = append(out, all[i])
		}
	}
	return out
}

// LookupWebDemo finds a catalog case by browser demo ID.
func LookupWebDemo(id string) (Case, bool) {
	demos := WebDemos()
	for i := range demos {
		if demos[i].WebDemoID == id {
			return demos[i], true
		}
	}
	return Case{}, false
}

func applyDefaults(c *Case) {
	if c.ID != "" {
		if c.GoPath == "" {
			if c.Showcase {
				c.GoPath = "examples/" + c.ID + "/example.go"
			} else {
				// Non-showcase cases (parity-only and fixture-only alike) live
				// exclusively under test/parity/ and are not surfaced in examples/.
				c.GoPath = "test/parity/" + c.ID + "/plot.go"
			}
		}
		c.PythonPath = "test/parity/" + c.ID + "/plot.py"
	}
	if c.Width == 0 {
		c.Width = DefaultWidth
	}
	if c.Height == 0 {
		c.Height = DefaultHeight
	}
	if c.DPI == 0 {
		c.DPI = DefaultDPI
	}
	if c.Description == "" {
		c.Description = c.Title
	}
}

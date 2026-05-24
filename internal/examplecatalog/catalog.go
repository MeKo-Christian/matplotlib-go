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
	NativeBackend       string
	NativeCapabilities  []string
	SVGGoldenFamily     string
	GoBasicSmokeFamily  string
	PickPointData       *[2]float64
	PickPointPixel      *[2]float64
	Pickable            bool
	NoPickReason        string
	SkipInteractivePan  bool
	SkipInteractiveZoom bool

	// MinPSNR / MaxMeanAbs / MaxRMSE override the matplotlib reference-compare
	// tolerance defaults for this case. Zero means "use the package default".
	// Only consulted by tests in test/; harmless metadata otherwise.
	MinPSNR    float64
	MaxMeanAbs float64
	MaxRMSE    float64
}

const (
	DefaultWidth  = 640
	DefaultHeight = 360
	DefaultDPI    = 100
)

var cases = []Case{
	{ID: "basic_line", Topic: "lines", Title: "Basic Line", Description: "A minimal line plot with explicit limits, labels, and one Line2D artist.", Showcase: true, GoBasicSmokeFamily: "line"},
	{ID: "joins_caps", Topic: "lines", Title: "Line Joins and Caps"},
	{ID: "dashes", Topic: "lines", Title: "Dash Patterns", Description: "Multiple line styles showing dash arrays, cap styles, and legend labeling.", Showcase: true},
	{ID: "line2d_semantics", Topic: "lines", Title: "Line2D Semantics", FixtureOnly: true, MinPSNR: 34.0, MaxMeanAbs: 8.0},
	{ID: "line2d_markers", Topic: "lines", Title: "Line2D Markers", FixtureOnly: true, MinPSNR: 30.0, MaxMeanAbs: 10.0},
	{ID: "path_effects", Topic: "effects", Title: "Path Effects", FixtureOnly: true, MinPSNR: 18.0, MaxMeanAbs: 24.0, MaxRMSE: 42.0},
	{ID: "pattern_gradient_effects", Topic: "effects", Title: "Pattern and Gradient Effects", FixtureOnly: true, MinPSNR: 16.0, MaxMeanAbs: 28.0, MaxRMSE: 48.0},
	{ID: "scatter_basic", Topic: "scatter", Title: "Basic Scatter", Description: "A compact scatter plot with variable marker size, color, alpha, and axes labels.", Showcase: true, GoBasicSmokeFamily: "scatter"},
	{ID: "scatter_marker_types", Topic: "scatter", Title: "Scatter Marker Types"},
	{ID: "scatter_advanced", Topic: "scatter", Title: "Advanced Scatter"},
	{ID: "bar_basic_frame", Topic: "bar", Title: "Bar Frame"},
	{ID: "bar_basic_ticks", Topic: "bar", Title: "Bar Ticks"},
	{ID: "bar_basic_tick_labels", Topic: "bar", Title: "Bar Tick Labels"},
	{ID: "bar_basic_title", Topic: "bar", Title: "Bar Title"},
	{ID: "bar_basic", Topic: "bar", Title: "Basic Bars", Description: "A categorical bar chart with tick labels, title text, and framed axes.", Showcase: true, SVGGoldenFamily: "bar", GoBasicSmokeFamily: "bar"},
	{ID: "bar_horizontal", Topic: "bar", Title: "Horizontal Bars"},
	{ID: "bar_grouped", Topic: "bar", Title: "Grouped Bars"},
	{ID: "fill_basic", Topic: "fill", Title: "Fill to Baseline", Description: "A filled region under a smooth curve, useful for area-chart and alpha blending checks.", Showcase: true, MinPSNR: 45.0, MaxMeanAbs: 6.0},
	{ID: "fill_between", Topic: "fill", Title: "Fill Between Curves", GoBasicSmokeFamily: "fill"},
	{ID: "fill_stacked", Topic: "fill", Title: "Stacked Fill"},
	{ID: "errorbar_basic", Topic: "errorbar", Title: "Error Bars", Description: "Symmetric and asymmetric error bars with caps, marker styling, and legend output.", Showcase: true, SVGGoldenFamily: "errorbar", GoBasicSmokeFamily: "errorbar"},
	{ID: "multi_series_basic", Topic: "multi", Title: "Multiple Series", Description: "Several labeled lines sharing one axes, demonstrating color cycling and legends.", Showcase: true},
	{ID: "multi_series_color_cycle", Topic: "multi", Title: "Color Cycle"},
	{ID: "hist_basic", Topic: "histogram", Title: "Histogram Counts", Description: "A deterministic histogram with count bins, labels, and default bar styling.", Showcase: true, SVGGoldenFamily: "hist", GoBasicSmokeFamily: "histogram"},
	{ID: "hist_density", Topic: "histogram", Title: "Histogram Density"},
	{ID: "hist_strategies", Topic: "histogram", Title: "Histogram Strategies"},
	{ID: "boxplot_basic", Topic: "boxplot", Title: "Box Plot", Description: "Grouped box plots with whiskers, medians, outliers, and categorical labels.", Optional: true, Showcase: true, MinPSNR: 44.0, MaxMeanAbs: 2.0},
	{ID: "text_labels_strict", Topic: "text", Title: "Strict Text Labels", Optional: true, SVGGoldenFamily: "text_layout", GoBasicSmokeFamily: "text"},
	{ID: "title_strict", Topic: "text", Title: "Strict Title"},
	{ID: "mathtext_basic", Topic: "mathtext", Title: "MathText Basic", FixtureOnly: true, SVGGoldenFamily: "mathtext", GoBasicSmokeFamily: "mathtext"},
	{ID: "mathtext_fractions", Topic: "mathtext", Title: "MathText Fractions", FixtureOnly: true},
	{ID: "mathtext_integrals", Topic: "mathtext", Title: "MathText Operators", FixtureOnly: true},
	{ID: "mathtext_matrices", Topic: "mathtext", Title: "MathText Matrices", FixtureOnly: true},
	{ID: "mathtext_inline_labels", Topic: "mathtext", Title: "MathText Inline Labels", FixtureOnly: true},
	{ID: "image_heatmap", Topic: "image", Title: "Heatmap Image", Description: "A gridded image plot with a colorbar and axis labels for matrix-style data.", Showcase: true, SVGGoldenFamily: "image", GoBasicSmokeFamily: "image"},
	{ID: "imshow_clipped", Topic: "image", Title: "Clipped Imshow", FixtureOnly: true, MinPSNR: 30.0, MaxMeanAbs: 10.0},
	{ID: "imshow_transformed", Topic: "image", Title: "Transformed Imshow", FixtureOnly: true, Width: 420, Height: 420, MinPSNR: 24.0, MaxMeanAbs: 18.0, MaxRMSE: 30.0},
	{ID: "imshow_bilinear", Topic: "image", Title: "Bilinear Imshow", FixtureOnly: true, Width: 256, Height: 256, MinPSNR: 30.0, MaxMeanAbs: 16.0},
	{ID: "imshow_bicubic", Topic: "image", Title: "Bicubic Imshow", FixtureOnly: true, Width: 256, Height: 256, MinPSNR: 30.0, MaxMeanAbs: 16.0},
	{ID: "image_alpha", Topic: "image", Title: "Image Alpha", FixtureOnly: true, MinPSNR: 30.0, MaxMeanAbs: 16.0, MaxRMSE: 10.0},
	{ID: "matshow_basic", Topic: "image", Title: "Matshow", FixtureOnly: true, MinPSNR: 30.0, MaxMeanAbs: 10.0, MaxRMSE: 10.0},
	{ID: "spy_marker", Topic: "image", Title: "Spy Marker Mode", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 12.0},
	{ID: "spy_image", Topic: "image", Title: "Spy Image Mode", FixtureOnly: true, MinPSNR: 27.0, MaxMeanAbs: 22.0, MaxRMSE: 30.0},
	{ID: "colormap_diverging", Topic: "colormap", Title: "Diverging Colormap", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "colormap_qualitative", Topic: "colormap", Title: "Qualitative Colormap", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "colormap_cyclic", Topic: "colormap", Title: "Cyclic Colormap", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "named_colors", Topic: "color", Title: "Named Colors", FixtureOnly: true, MinPSNR: 35.0, MaxMeanAbs: 8.0},
	{ID: "axes_top_right_inverted", Topic: "axes", Title: "Top/Right Inverted Axes", Optional: true},
	{ID: "axes_control_surface", Topic: "axes", Title: "Axes, Scales, and Twins", Optional: true, WebDemoID: "axes", Description: "Minor ticks, top/right axes, aspect controls, log scale, twin axes, and secondary axes.", Showcase: true, GoBasicSmokeFamily: "axes", MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "transform_coordinates", Topic: "axes", Title: "Transform Coordinates", Optional: true, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "transform_annotation_modes", Topic: "axes", Title: "Annotation Coordinate Modes", FixtureOnly: true, Width: 720, Height: 420, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "path_clipped_transformed", Topic: "axes", Title: "Clipped Transformed Path", FixtureOnly: true, Width: 720, Height: 420, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "layout_bbox_helpers", Topic: "axes", Title: "Layout BBox Helpers", FixtureOnly: true, Width: 720, Height: 420, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "formatter_engineering_labels", Topic: "axes", Title: "Engineering Formatter Labels", FixtureOnly: true, Width: 720, Height: 400, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "formatter_fixed_null_labels", Topic: "axes", Title: "Fixed and Null Formatter Labels", FixtureOnly: true, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "formatter_log_mathtext_labels", Topic: "axes", Title: "Log MathText Formatter Labels", FixtureOnly: true, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "formatter_percent_labels", Topic: "axes", Title: "Percent Formatter Labels", FixtureOnly: true, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "formatter_scalar_scientific_labels", Topic: "axes", Title: "Scalar Scientific Formatter Labels", FixtureOnly: true, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "locator_fixed_index_labels", Topic: "axes", Title: "Fixed and Index Locator Labels", FixtureOnly: true, Width: 720, Height: 420, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "locator_linear_labels", Topic: "axes", Title: "Linear Locator Labels", FixtureOnly: true, Width: 720, Height: 540, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "locator_log_minor_threshold_labels", Topic: "axes", Title: "Log Locator Minor Labels", FixtureOnly: true, Width: 720, Height: 420, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "locator_maxn_edge_labels", Topic: "axes", Title: "MaxN Locator Edge Labels", FixtureOnly: true, Width: 720, Height: 540, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "scale_asinh_ticks", Topic: "axes", Title: "Asinh Scale Ticks", FixtureOnly: true, Width: 720, Height: 400, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "scale_function_defaults", Topic: "axes", Title: "Function Scale Defaults", FixtureOnly: true, Width: 720, Height: 480, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "scale_logit_ticks", Topic: "axes", Title: "Logit Scale Ticks", FixtureOnly: true, Width: 720, Height: 400, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "scale_symlog_ticks", Topic: "axes", Title: "Symlog Scale Ticks", FixtureOnly: true, Width: 720, Height: 400, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "ticks_styling_surface", Topic: "axes", Title: "Tick Styling Surface", FixtureOnly: true, Width: 720, Height: 420, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "artist_metadata", Topic: "artist", Title: "Artist Metadata", FixtureOnly: true, MinPSNR: 40.0, MaxMeanAbs: 4.0},
	{ID: "gridspec_composition", Topic: "composition", Title: "Figure Composition", WebDemoID: "composition", Description: "GridSpec spans, figure-level labels, figure legends, anchored text, and colorbars.", Showcase: true, GoBasicSmokeFamily: "layout", MinPSNR: 35.0, MaxMeanAbs: 8.0},
	{ID: "figure_labels_composition", Topic: "composition", Title: "Figure Labels", Description: "A multi-axes figure with shared figure title, x label, y label, and legend placement.", Showcase: true, MinPSNR: 32.0, MaxMeanAbs: 9.0},
	{ID: "colorbar_composition", Topic: "colorbar", Title: "Colorbar Composition", Description: "A composed figure that exercises image color mapping, shared colorbars, and layout spacing.", Showcase: true, GoBasicSmokeFamily: "colorbar", MinPSNR: 32.0, MaxMeanAbs: 16.0},
	{ID: "annotation_composition", Topic: "annotation", Title: "Annotations", Description: "Text annotations, arrows, anchored labels, and mixed coordinate placement.", Showcase: true, GoBasicSmokeFamily: "annotation", MinPSNR: 35.0, MaxMeanAbs: 7.0},
	{ID: "patch_showcase", Topic: "patches", Title: "Patch Showcase", Optional: true, SVGGoldenFamily: "hatch_bars", GoBasicSmokeFamily: "patch_hatch", MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "mesh_contour_tri", Topic: "mesh", Title: "Meshes and Contours", Optional: true, WebDemoID: "mesh", Description: "PColorMesh, contour/contourf, Hist2D, triplot, tripcolor, and tricontour.", Showcase: true, GoBasicSmokeFamily: "mesh", MinPSNR: 37.5, MaxMeanAbs: 7.5},
	{ID: "plot_variants", Topic: "variants", Title: "Plot Variants", Optional: true, WebDemoID: "variants", Description: "Step, stairs, reference lines, spans, broken bars, and stacked bars.", Showcase: true, GoBasicSmokeFamily: "variants", MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "spectrum_variants", Topic: "signal", Title: "Spectrum Variants", FixtureOnly: true, GoBasicSmokeFamily: "signal", MinPSNR: 35.0, MaxMeanAbs: 6.5, MaxRMSE: 10.0},
	{ID: "stat_variants", Topic: "statistics", Title: "Statistical Views", Optional: true, WebDemoID: "statistics", Description: "Box plots, violin plots, empirical CDFs, and stack plots.", Showcase: true, GoBasicSmokeFamily: "statistics", MinPSNR: 32.0, MaxMeanAbs: 9.0},
	{ID: "specialty_depth", Topic: "statistics", Title: "Specialty Depth", FixtureOnly: true, MinPSNR: 22.0, MaxMeanAbs: 20.0, MaxRMSE: 35.0},
	{ID: "stem_plot", Topic: "specialty", Title: "Stem Plot", Optional: true},
	{ID: "specialty_artists", Topic: "specialty", Title: "Specialty Artists", Optional: true, WebDemoID: "specialty", Description: "Event plots, hexbin, pie charts, stem plots, tables, and Sankey-style flows.", Showcase: true, GoBasicSmokeFamily: "specialty"},
	{ID: "date_concise_intraday_labels", Topic: "units", Title: "Concise Intraday Date Labels", FixtureOnly: true, Width: 720, Height: 360, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "date_month_year_labels", Topic: "units", Title: "Month and Year Date Labels", FixtureOnly: true, Width: 720, Height: 360, MinPSNR: 35.0, MaxMeanAbs: 6.5},
	{ID: "units_overview", Topic: "units", Title: "Dates and Categories", Optional: true, WebDemoID: "units", Description: "Time-aware axes, categorical bars, and horizontal categorical bars.", Showcase: true, GoBasicSmokeFamily: "units", MinPSNR: 43.5},
	{ID: "units_dates", Topic: "units", Title: "Date Units", Optional: true, MinPSNR: 45.0, MaxMeanAbs: 1.6},
	{ID: "units_categories", Topic: "units", Title: "Category Units", Optional: true, MinPSNR: 41.0, MaxMeanAbs: 3.2},
	{ID: "units_custom_converter", Topic: "units", Title: "Custom Unit Converter", Optional: true, MinPSNR: 40.0, MaxMeanAbs: 3.5},
	{ID: "vector_fields", Topic: "vectors", Title: "Vector Fields", Optional: true, WebDemoID: "vectors", Description: "Quiver, quiver keys, barbs, streamplots, and grid-based vector input.", Showcase: true, GoBasicSmokeFamily: "vectors", MinPSNR: 41.5, MaxMeanAbs: 3.0},
	{ID: "polar_axes", Topic: "polar", Title: "Polar Wave", WebDemoID: "polar", Description: "A filled polar curve with custom radial and angular grid styling.", Showcase: true, SVGGoldenFamily: "clipped_polar", GoBasicSmokeFamily: "polar", MinPSNR: 32.0, MaxMeanAbs: 9.0},
	{ID: "geo_mollweide_axes", Topic: "geo", Title: "Projections and Insets", WebDemoID: "projections", Description: "Mollweide geo projection plus a zoomed inset axes.", Showcase: true, GoBasicSmokeFamily: "geo", MinPSNR: 30.0, MaxMeanAbs: 12.0},
	{ID: "geo_aitoff_axes", Topic: "geo", Title: "Aitoff Projection", Description: "An Aitoff equal-area projection with longitude wrapping and graticule rendering.", Optional: true, Showcase: true, MinPSNR: 30.0, MaxMeanAbs: 12.0},
	{ID: "geo_hammer_axes", Topic: "geo", Title: "Hammer Projection", Optional: true, MinPSNR: 30.0, MaxMeanAbs: 12.0},
	{ID: "geo_lambert_axes", Topic: "geo", Title: "Lambert Projection", Optional: true, MinPSNR: 30.0, MaxMeanAbs: 12.0},
	{ID: "radar_basic", Topic: "radar", Title: "Radar Projection", Description: "A radar chart using polar projection plumbing with closed polygon series.", Optional: true, Showcase: true, MinPSNR: 45.0, MaxMeanAbs: 2.0},
	{ID: "skewt_basic", Topic: "skewt", Title: "Skew-T Projection", Description: "A meteorological-style skew-T axes with transformed temperature grid lines.", Optional: true, Showcase: true, MinPSNR: 24.0, MaxMeanAbs: 18.0},
	{ID: "mplot3d_basic", Topic: "mplot3d", Title: "3D Toolkit Scaffold", Optional: true, GoBasicSmokeFamily: "mplot3d", MinPSNR: 39.0, MaxMeanAbs: 5.0, MaxRMSE: 18.0},
	{ID: "mplot3d_terrain", Topic: "mplot3d", Title: "3D Terrain", Description: "A 3D surface terrain plot with depth sorting, pane styling, and projected axes.", Optional: true, Width: 900, Height: 640, Showcase: true, MinPSNR: 38.0, MaxMeanAbs: 5.0, MaxRMSE: 18.0},
	{ID: "mplot3d_plot3d", Topic: "mplot3d", Title: "3D Plot", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 38.0, MaxMeanAbs: 8.0, MaxRMSE: 18.0},
	{ID: "mplot3d_scatter3d", Topic: "mplot3d", Title: "3D Scatter", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 35.0, MaxMeanAbs: 8.0, MaxRMSE: 18.0},
	{ID: "mplot3d_surface3d", Topic: "mplot3d", Title: "3D Surface", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 35.0, MaxMeanAbs: 10.0, MaxRMSE: 18.0},
	{ID: "mplot3d_wire3d", Topic: "mplot3d", Title: "3D Wireframe", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 10.0, MaxRMSE: 18.0},
	{ID: "mplot3d_trisurf3d", Topic: "mplot3d", Title: "3D Triangulated Surface", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 12.0, MaxRMSE: 18.0},
	{ID: "mplot3d_bar3d", Topic: "mplot3d", Title: "3D Bars", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 8.0, MaxRMSE: 18.0},
	{ID: "mplot3d_voxels", Topic: "mplot3d", Title: "3D Voxels", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 12.0, MaxRMSE: 18.0},
	{ID: "mplot3d_quiver3d", Topic: "mplot3d", Title: "3D Quiver", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 10.0, MaxRMSE: 18.0},
	{ID: "mplot3d_stem3d", Topic: "mplot3d", Title: "3D Stem", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 30.0, MaxMeanAbs: 10.0, MaxRMSE: 18.0},
	{ID: "mplot3d_fill_between3d", Topic: "mplot3d", Title: "3D Fill Between", FixtureOnly: true, Width: 720, Height: 560, MinPSNR: 35.0, MaxMeanAbs: 10.0, MaxRMSE: 18.0},
	{ID: "unstructured_showcase", Topic: "unstructured", Title: "Unstructured Showcase", Description: "Triangulated and irregular data views covering triplot, tripcolor, and contour variants.", Optional: true, Showcase: true, GoBasicSmokeFamily: "unstructured", MinPSNR: 30.0, MaxMeanAbs: 10.0},
	{ID: "arrays_showcase", Topic: "arrays", Title: "Matrix Helpers", Optional: true, WebDemoID: "matrix", Description: "MatShow, sparsity spy plots, annotated heatmaps, and colorbars.", Width: 1240, Height: 620, Showcase: true, GoBasicSmokeFamily: "matrix", MinPSNR: 30.0, MaxMeanAbs: 10.0},
	{ID: "axisartist_showcase", Topic: "axisartist", Title: "AxisArtist Showcase", Description: "Floating and axis-artist style axes with custom spine placement and labels.", Optional: true, Showcase: true, GoBasicSmokeFamily: "axisartist", MinPSNR: 28.0, MaxMeanAbs: 12.0},
	{ID: "axes_grid1_showcase", Topic: "axes_grid1", Title: "Axes Grid1 Showcase", Description: "Axes divider, image-grid, inset, and anchored layout helpers in one composition.", Optional: true, Showcase: true, GoBasicSmokeFamily: "axes_grid1", MinPSNR: 28.0, MaxMeanAbs: 12.0},
	{ID: "pcolor_flat", Topic: "mesh", Title: "PColor Flat", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 15.0},
	{ID: "pcolormesh_nearest", Topic: "mesh", Title: "PColorMesh Nearest", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 15.0},
	{ID: "pcolormesh_gouraud", Topic: "mesh", Title: "PColorMesh Gouraud", FixtureOnly: true, MinPSNR: 20.0, MaxMeanAbs: 22.0, MaxRMSE: 30.0},
	{ID: "pcolormesh_masked", Topic: "mesh", Title: "PColorMesh Masked", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 15.0},
	{ID: "hist2d_weighted_density", Topic: "mesh", Title: "Hist2D Weighted Density", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0, MaxRMSE: 30.0},
	{ID: "boundarynorm_pcolormesh", Topic: "colorbar", Title: "BoundaryNorm PColorMesh", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "lognorm_imshow", Topic: "colorbar", Title: "LogNorm Imshow", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "twoslope_norm_image", Topic: "colorbar", Title: "TwoSlopeNorm Image", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "colorbar_extensions", Topic: "colorbar", Title: "Colorbar Extensions", FixtureOnly: true, MinPSNR: 28.0, MaxMeanAbs: 16.0},
	{ID: "large_scatter", Topic: "raster", Title: "Large Scatter Batch", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"pathcollectionbatch"}, MinPSNR: 55.0, MaxMeanAbs: 0.5, MaxRMSE: 4.0},
	{ID: "mixed_collection", Topic: "raster", Title: "Mixed Path Collection", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"pathcollectionbatch"}, SVGGoldenFamily: "collection", GoBasicSmokeFamily: "collection", MinPSNR: 60.0, MaxMeanAbs: 0.5, MaxRMSE: 2.0},
	{ID: "mixed_raster_vector", Topic: "raster", Title: "Mixed Raster Vector Output", FixtureOnly: true, Width: 640, Height: 640, SVGGoldenFamily: "mixed_raster", MinPSNR: 32.0, MaxMeanAbs: 9.0, MaxRMSE: 18.0},
	{ID: "quad_mesh", Topic: "raster", Title: "Quad Mesh Batch", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"quadmeshbatch"}, MinPSNR: 48.0, MaxMeanAbs: 1.0, MaxRMSE: 4.0},
	{ID: "gouraud_triangles", Topic: "raster", Title: "Gouraud Triangles", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"gouraudtrianglebatch"}, MinPSNR: 25.0, MaxMeanAbs: 18.0},
	{ID: "clip_path_batch", Topic: "raster", Title: "Clip Path Batch", FixtureOnly: true, NativeBackend: "agg", NativeCapabilities: []string{"pathclip", "quadmeshbatch"}, MinPSNR: 45.0, MaxMeanAbs: 1.0, MaxRMSE: 6.0},
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

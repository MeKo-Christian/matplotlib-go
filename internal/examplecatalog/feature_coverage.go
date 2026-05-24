package examplecatalog

// CoverageStatus records whether a Matplotlib feature area has local coverage.
type CoverageStatus string

const (
	// CoverageImplemented means the row has a direct Go implementation or
	// equivalent for the main public behavior being tracked.
	CoverageImplemented CoverageStatus = "implemented"
	// CoveragePartial means a Go equivalent exists, but important upstream
	// breadth or edge behavior is still missing.
	CoveragePartial CoverageStatus = "partial"
	// CoveragePending means the row is planned but not implemented yet.
	CoveragePending CoverageStatus = "pending"
	// CoverageOmitted means the row is intentionally not implemented.
	CoverageOmitted CoverageStatus = "omitted"
)

// BreadthStatus records whether examples exercise the breadth of a feature.
type BreadthStatus string

const (
	// BreadthBroad means current demos exercise the important variants.
	BreadthBroad BreadthStatus = "broad"
	// BreadthThin means coverage exists but does not yet show enough variants.
	BreadthThin BreadthStatus = "thin"
	// BreadthFixtureOnly means parity fixtures exist, but user-facing demos are
	// missing or too sparse.
	BreadthFixtureOnly BreadthStatus = "fixture-only"
	// BreadthPending means breadth coverage has not started.
	BreadthPending BreadthStatus = "pending"
	// BreadthNotApplicable means visual/demo breadth does not apply directly.
	BreadthNotApplicable BreadthStatus = "not-applicable"
)

// FeatureCoverage describes one Matplotlib feature area in the Phase 9A
// machine-readable coverage matrix.
type FeatureCoverage struct {
	ID                      string
	Title                   string
	UpstreamModules         []string
	UpstreamGalleryFamilies []string
	GoFiles                 []string
	CatalogIDs              []string
	ExampleIDs              []string
	WebDemoIDs              []string
	GoEquivalent            CoverageStatus
	ParityFixture           CoverageStatus
	UserShowcase            CoverageStatus
	BrowserDemo             CoverageStatus
	Breadth                 BreadthStatus
	IntentionalOmission     string
	Notes                   string
}

var featureCoverageRows = []FeatureCoverage{
	{
		ID:              "artist",
		Title:           "Artist Base Model",
		UpstreamModules: []string{"artist.py"},
		GoFiles:         []string{"core/artist.go", "core/lifecycle.go", "core/rasterization.go"},
		CatalogIDs:      []string{"basic_line", "patch_showcase", "mixed_raster_vector"},
		GoEquivalent:    CoveragePartial,
		ParityFixture:   CoverageImplemented,
		UserShowcase:    CoveragePartial,
		BrowserDemo:     CoveragePartial,
		Breadth:         BreadthThin,
		Notes:           "Core artist draw, z-order, bounds, rasterization, and lifecycle behavior exists; Matplotlib's broader property/setter/callback surface remains partial.",
	},
	{
		ID:                      "axes",
		Title:                   "Axes Plotting Surface",
		UpstreamModules:         []string{"axes/_axes.py", "axes/_base.py"},
		UpstreamGalleryFamilies: []string{"lines_bars_and_markers", "statistics", "images_contours_and_fields"},
		GoFiles:                 []string{"core/artist.go", "core/plot.go", "core/plot_variants.go", "core/histogram.go", "core/mesh.go", "core/contour.go"},
		CatalogIDs:              []string{"basic_line", "scatter_basic", "bar_basic", "hist_basic", "mesh_contour_tri", "stat_variants"},
		ExampleIDs:              []string{"basic_line", "scatter_basic", "bar_basic", "hist_basic", "mesh_contour_tri", "stat_variants"},
		WebDemoIDs:              []string{"axes", "mesh", "statistics"},
		GoEquivalent:            CoverageImplemented,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoverageImplemented,
		BrowserDemo:             CoverageImplemented,
		Breadth:                 BreadthBroad,
	},
	{
		ID:                      "figure-layout",
		Title:                   "Figure, Subplots, and Layout",
		UpstreamModules:         []string{"figure.py", "gridspec.py", "_tight_layout.py", "_constrained_layout.py", "layout_engine.py"},
		UpstreamGalleryFamilies: []string{"subplots_axes_and_figures"},
		GoFiles:                 []string{"core/artist.go", "core/gridspec.go", "core/subplots.go", "core/layout_engine.go", "core/figure_layout.go"},
		CatalogIDs:              []string{"gridspec_composition", "figure_labels_composition", "axes_grid1_showcase"},
		ExampleIDs:              []string{"gridspec_composition", "figure_labels_composition", "axes_grid1_showcase"},
		WebDemoIDs:              []string{"composition"},
		GoEquivalent:            CoverageImplemented,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoverageImplemented,
		BrowserDemo:             CoveragePartial,
		Breadth:                 BreadthBroad,
	},
	{
		ID:                      "axis-ticker-scale",
		Title:                   "Axis, Tickers, Formatters, and Scales",
		UpstreamModules:         []string{"axis.py", "ticker.py", "scale.py", "dates.py", "category.py"},
		UpstreamGalleryFamilies: []string{"ticks", "scales"},
		GoFiles:                 []string{"core/axis.go", "core/date_tick.go", "core/units.go", "transform/scale_registry.go"},
		CatalogIDs:              []string{"axes_control_surface", "units_dates", "units_categories", "lognorm_imshow", "skewt_basic"},
		ExampleIDs:              []string{"axes_control_surface", "units_overview", "skewt_basic"},
		WebDemoIDs:              []string{"axes", "units"},
		GoEquivalent:            CoveragePartial,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoveragePartial,
		BrowserDemo:             CoveragePartial,
		Breadth:                 BreadthThin,
		Notes: "Major/minor ticks, TickParams styling, date/category units, and linear/log/symlog/asinh/logit/function/functionlog scales exist. " +
			"The dates.py surface audit is tracked in DateSurfaceAuditRows. Remaining 12.2C/D/E scope is formatter edge cases and family-specific catalog fixtures.",
	},
	{
		ID:                      "transforms",
		Title:                   "Transforms and Coordinate Systems",
		UpstreamModules:         []string{"transforms.py", "bezier.py", "path.py"},
		UpstreamGalleryFamilies: []string{"subplots_axes_and_figures", "text_labels_and_annotations"},
		GoFiles:                 []string{"transform/transform.go", "transform/graph.go", "transform/node.go", "internal/geom/geom.go"},
		CatalogIDs:              []string{"transform_coordinates", "transform_annotation_modes", "path_clipped_transformed", "layout_bbox_helpers", "annotation_composition", "imshow_transformed"},
		ExampleIDs:              []string{"annotation_composition"},
		WebDemoIDs:              []string{"axes"},
		GoEquivalent:            CoveragePartial,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoveragePartial,
		BrowserDemo:             CoveragePartial,
		Breadth:                 BreadthThin,
		Notes: "TransformedPath caching, annotation coordinates, clipped transformed paths, BBox layout, and visible path interpolation/clipping/extents are covered; " +
			"simplification and curve splitting stay omitted until a visible parity target requires them.",
	},
	{
		ID:                      "lines",
		Title:                   "Lines, Dashes, Markers, Caps, and Joins",
		UpstreamModules:         []string{"lines.py", "markers.py"},
		UpstreamGalleryFamilies: []string{"lines_bars_and_markers"},
		GoFiles:                 []string{"core/line.go", "core/plot.go", "core/scatter.go"},
		CatalogIDs:              []string{"basic_line", "dashes", "joins_caps", "scatter_marker_types"},
		ExampleIDs:              []string{"basic_line", "dashes"},
		GoEquivalent:            CoverageImplemented,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoveragePartial,
		BrowserDemo:             CoveragePending,
		Breadth:                 BreadthThin,
	},
	{
		ID:                      "collections",
		Title:                   "Collections and Batched Drawing",
		UpstreamModules:         []string{"collections.py"},
		UpstreamGalleryFamilies: []string{"shapes_and_collections", "images_contours_and_fields"},
		GoFiles:                 []string{"core/collection.go", "core/mesh.go", "core/eventplot.go", "core/hexbin.go"},
		CatalogIDs:              []string{"mixed_collection", "large_scatter", "quad_mesh", "gouraud_triangles", "collection_mutable_scalarmap", "specialty_artists"},
		ExampleIDs:              []string{"specialty_artists"},
		WebDemoIDs:              []string{"specialty"},
		GoEquivalent:            CoveragePartial,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoveragePartial,
		BrowserDemo:             CoveragePartial,
		Breadth:                 BreadthThin,
	},
	{
		ID:                      "patches",
		Title:                   "Patches, Hatches, and Shape Styles",
		UpstreamModules:         []string{"patches.py", "hatch.py"},
		UpstreamGalleryFamilies: []string{"shapes_and_collections"},
		GoFiles:                 []string{"core/patch.go", "core/patch_extra.go", "core/arrow_patch.go"},
		CatalogIDs:              []string{"patch_showcase", "patch_style_matrix"},
		GoEquivalent:            CoveragePartial,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoveragePending,
		BrowserDemo:             CoveragePending,
		Breadth:                 BreadthFixtureOnly,
	},
	{
		ID:                      "text-annotation-legend",
		Title:                   "Text, Annotations, Legends, and Offset Boxes",
		UpstreamModules:         []string{"text.py", "legend.py", "legend_handler.py", "offsetbox.py", "textpath.py"},
		UpstreamGalleryFamilies: []string{"text_labels_and_annotations"},
		GoFiles:                 []string{"core/text.go", "core/annotation_box.go", "core/legend.go", "core/anchored_text.go", "core/anchored_drawing_area.go", "core/anchored_packer.go", "core/anchored_sizebar.go", "render/text_path.go"},
		CatalogIDs:              []string{"text_labels_strict", "title_strict", "annotation_composition", "figure_labels_composition", "mathtext_inline_labels", "legend_layout_matrix", "text_annotation_matrix"},
		ExampleIDs:              []string{"annotation_composition", "figure_labels_composition"},
		WebDemoIDs:              []string{"composition"},
		GoEquivalent:            CoveragePartial,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoveragePartial,
		BrowserDemo:             CoveragePartial,
		Breadth:                 BreadthThin,
	},
	{
		ID:                      "image",
		Title:                   "Images, Matrix Helpers, and Interpolation",
		UpstreamModules:         []string{"image.py"},
		UpstreamGalleryFamilies: []string{"images_contours_and_fields"},
		GoFiles:                 []string{"core/image.go", "core/image_api.go", "core/matrix_helpers.go"},
		CatalogIDs:              []string{"image_heatmap", "imshow_clipped", "imshow_transformed", "imshow_bilinear", "imshow_bicubic", "image_alpha", "matshow_basic", "spy_marker", "spy_image", "arrays_showcase"},
		ExampleIDs:              []string{"image_heatmap", "arrays_showcase"},
		WebDemoIDs:              []string{"matrix"},
		GoEquivalent:            CoverageImplemented,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoveragePartial,
		BrowserDemo:             CoveragePartial,
		Breadth:                 BreadthThin,
	},
	{
		ID:                      "colorbar",
		Title:                   "Scalar Mappables and Colorbars",
		UpstreamModules:         []string{"colorbar.py", "colorizer.py"},
		UpstreamGalleryFamilies: []string{"images_contours_and_fields"},
		GoFiles:                 []string{"core/colorbar.go", "core/scalar_mappable.go", "core/norm.go"},
		CatalogIDs:              []string{"colorbar_composition", "asinh_norm_image", "boundarynorm_pcolormesh", "collection_mutable_scalarmap", "colorbar_boundary_values", "colorbar_horizontal_ticks", "lognorm_imshow", "twoslope_norm_image", "colorbar_extensions"},
		ExampleIDs:              []string{"colorbar_composition"},
		WebDemoIDs:              []string{"composition", "matrix"},
		GoEquivalent:            CoveragePartial,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoveragePartial,
		BrowserDemo:             CoveragePartial,
		Breadth:                 BreadthThin,
	},
	{
		ID:                      "colors-cm",
		Title:                   "Colors, Colormaps, and Norms",
		UpstreamModules:         []string{"colors.py", "cm.py", "_cm.py", "_cm_listed.py", "_color_data.py"},
		UpstreamGalleryFamilies: []string{"color"},
		GoFiles:                 []string{"color/colormap.go", "color/listed_colormaps.go", "color/named_colors.go", "core/norm.go"},
		CatalogIDs:              []string{"colormap_diverging", "colormap_qualitative", "colormap_cyclic", "named_colors", "asinh_norm_image", "lognorm_imshow", "twoslope_norm_image"},
		GoEquivalent:            CoveragePartial,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoveragePending,
		BrowserDemo:             CoveragePending,
		Breadth:                 BreadthFixtureOnly,
	},
	{
		ID:              "pyplot-state",
		Title:           "Pyplot Stateful API",
		UpstreamModules: []string{"pyplot.py", "_pylab_helpers.py"},
		GoFiles:         []string{"pyplot/pyplot.go", "canvas/canvas.go"},
		CatalogIDs:      []string{"basic_line", "scatter_basic", "bar_basic"},
		ExampleIDs:      []string{"basic_line", "scatter_basic", "bar_basic"},
		GoEquivalent:    CoveragePartial,
		ParityFixture:   CoverageImplemented,
		UserShowcase:    CoveragePartial,
		BrowserDemo:     CoveragePending,
		Breadth:         BreadthThin,
	},
	{
		ID:                      "renderer-backends",
		Title:                   "Renderer Contract and Backends",
		UpstreamModules:         []string{"backend_bases.py", "backends/backend_agg.py", "backends/backend_svg.py", "backends/backend_pdf.py", "backends/backend_ps.py", "backends/backend_pgf.py"},
		UpstreamGalleryFamilies: []string{"misc"},
		GoFiles:                 []string{"render/render.go", "render/graphics_context.go", "render/extensions.go", "backends/registry.go"},
		CatalogIDs:              []string{"basic_line", "mixed_raster_vector", "large_scatter", "clip_path_batch"},
		ExampleIDs:              []string{"basic_line"},
		GoEquivalent:            CoverageImplemented,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoveragePartial,
		BrowserDemo:             CoveragePending,
		Breadth:                 BreadthThin,
	},
	{
		ID:                      "widgets-events-animation",
		Title:                   "Widgets, Events, and Animation",
		UpstreamModules:         []string{"widgets.py", "backend_bases.py", "backend_tools.py", "animation.py"},
		UpstreamGalleryFamilies: []string{"widgets", "event_handling", "animation", "user_interfaces"},
		GoFiles:                 []string{"core/widget_button.go", "core/widget_slider.go", "core/widget_rangeslider.go", "core/widget_checkbuttons.go", "core/widget_radiobuttons.go", "core/widget_textbox.go", "core/selectors.go", "core/widgets_common.go", "canvas/widget_interaction.go", "canvas/dispatcher.go", "canvas/navigation.go", "canvas/picker.go", "canvas/scheduler.go", "animation/animation.go"},
		GoEquivalent:            CoveragePartial,
		ParityFixture:           CoveragePending,
		UserShowcase:            CoveragePending,
		BrowserDemo:             CoveragePending,
		Breadth:                 BreadthPending,
		Notes:                   "Static widget artists, common selector workflows, event/navigation infrastructure, and FuncAnimation / ArtistAnimation-style playback exist; browser demos, writer scope, and GUI-specific edge behavior remain pending.",
	},
	{
		ID:                      "toolkits-projections",
		Title:                   "Toolkits and Projections",
		UpstreamModules:         []string{"projections/polar.py", "projections/geo.py", "mpl_toolkits/mplot3d", "mpl_toolkits/axisartist", "mpl_toolkits/axes_grid1"},
		UpstreamGalleryFamilies: []string{"mplot3d", "axisartist", "axes_grid1", "specialty_plots"},
		GoFiles:                 []string{"core/polar.go", "core/geo.go", "core/skew.go", "core/axes3d.go", "core/axis_artist.go", "core/image_grid.go"},
		CatalogIDs:              []string{"polar_axes", "geo_mollweide_axes", "geo_aitoff_axes", "geo_hammer_axes", "geo_lambert_axes", "radar_basic", "skewt_basic", "mplot3d_terrain", "axisartist_showcase", "axes_grid1_showcase"},
		ExampleIDs:              []string{"polar_axes", "geo_mollweide_axes", "geo_aitoff_axes", "radar_basic", "skewt_basic", "mplot3d_terrain", "axisartist_showcase", "axes_grid1_showcase"},
		WebDemoIDs:              []string{"polar", "projections"},
		GoEquivalent:            CoveragePartial,
		ParityFixture:           CoverageImplemented,
		UserShowcase:            CoverageImplemented,
		BrowserDemo:             CoveragePartial,
		Breadth:                 BreadthThin,
	},
}

// FeatureCoverageMatrix returns the Phase 9A feature coverage inventory.
func FeatureCoverageMatrix() []FeatureCoverage {
	out := make([]FeatureCoverage, len(featureCoverageRows))
	copy(out, featureCoverageRows)
	for i := range out {
		out[i].UpstreamModules = append([]string(nil), out[i].UpstreamModules...)
		out[i].UpstreamGalleryFamilies = append([]string(nil), out[i].UpstreamGalleryFamilies...)
		out[i].GoFiles = append([]string(nil), out[i].GoFiles...)
		out[i].CatalogIDs = append([]string(nil), out[i].CatalogIDs...)
		out[i].ExampleIDs = append([]string(nil), out[i].ExampleIDs...)
		out[i].WebDemoIDs = append([]string(nil), out[i].WebDemoIDs...)
	}
	return out
}

// LookupFeatureCoverage finds a coverage row by stable row ID.
func LookupFeatureCoverage(id string) (FeatureCoverage, bool) {
	for _, row := range FeatureCoverageMatrix() {
		if row.ID == id {
			return row, true
		}
	}
	return FeatureCoverage{}, false
}

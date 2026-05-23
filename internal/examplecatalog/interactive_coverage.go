package examplecatalog

// InteractiveCoverageRow records the representative case used to smoke-test
// one catalog topic through interactive backends.
type InteractiveCoverageRow struct {
	Topic            string
	RepresentativeID string
	WebAgg           bool
	Gio              bool
	Notes            string
}

var interactiveCoverage = []InteractiveCoverageRow{
	{Topic: "annotation", RepresentativeID: "annotation_composition", WebAgg: true, Gio: true},
	{Topic: "arrays", RepresentativeID: "arrays_showcase", WebAgg: true, Gio: true},
	{Topic: "axes", RepresentativeID: "axes_control_surface", WebAgg: true, Gio: true},
	{Topic: "axes_grid1", RepresentativeID: "axes_grid1_showcase", WebAgg: true, Gio: true},
	{Topic: "axisartist", RepresentativeID: "axisartist_showcase", WebAgg: true, Gio: true},
	{Topic: "bar", RepresentativeID: "bar_basic", WebAgg: true, Gio: true},
	{Topic: "boxplot", RepresentativeID: "boxplot_basic", WebAgg: true, Gio: true},
	{Topic: "color", RepresentativeID: "named_colors", WebAgg: true, Gio: true},
	{Topic: "colorbar", RepresentativeID: "colorbar_composition", WebAgg: true, Gio: true},
	{Topic: "colormap", RepresentativeID: "colormap_diverging", WebAgg: true, Gio: true},
	{Topic: "composition", RepresentativeID: "gridspec_composition", WebAgg: true, Gio: true},
	{Topic: "errorbar", RepresentativeID: "errorbar_basic", WebAgg: true, Gio: true},
	{Topic: "fill", RepresentativeID: "fill_between", WebAgg: true, Gio: true},
	{Topic: "geo", RepresentativeID: "geo_mollweide_axes", WebAgg: true, Gio: true},
	{Topic: "histogram", RepresentativeID: "hist_basic", WebAgg: true, Gio: true},
	{Topic: "image", RepresentativeID: "image_heatmap", WebAgg: true, Gio: true},
	{Topic: "lines", RepresentativeID: "basic_line", WebAgg: true, Gio: true},
	{Topic: "mathtext", RepresentativeID: "mathtext_basic", WebAgg: true, Gio: true},
	{Topic: "mesh", RepresentativeID: "mesh_contour_tri", WebAgg: true, Gio: true},
	{Topic: "mplot3d", RepresentativeID: "mplot3d_basic", WebAgg: true, Gio: true},
	{Topic: "multi", RepresentativeID: "multi_series_basic", WebAgg: true, Gio: true},
	{Topic: "patches", RepresentativeID: "patch_showcase", WebAgg: true, Gio: true},
	{Topic: "polar", RepresentativeID: "polar_axes", WebAgg: true, Gio: true},
	{Topic: "radar", RepresentativeID: "radar_basic", WebAgg: true, Gio: true},
	{Topic: "raster", RepresentativeID: "mixed_collection", WebAgg: true, Gio: true},
	{Topic: "scatter", RepresentativeID: "scatter_basic", WebAgg: true, Gio: true},
	{Topic: "signal", RepresentativeID: "spectrum_variants", WebAgg: true, Gio: true},
	{Topic: "skewt", RepresentativeID: "skewt_basic", WebAgg: true, Gio: true},
	{Topic: "specialty", RepresentativeID: "specialty_artists", WebAgg: true, Gio: true},
	{Topic: "statistics", RepresentativeID: "stat_variants", WebAgg: true, Gio: true},
	{Topic: "text", RepresentativeID: "text_labels_strict", WebAgg: true, Gio: true},
	{Topic: "units", RepresentativeID: "units_overview", WebAgg: true, Gio: true},
	{Topic: "unstructured", RepresentativeID: "unstructured_showcase", WebAgg: true, Gio: true},
	{Topic: "variants", RepresentativeID: "plot_variants", WebAgg: true, Gio: true},
	{Topic: "vectors", RepresentativeID: "vector_fields", WebAgg: true, Gio: true},
}

// InteractiveCoverageMatrix returns a copy of the Phase 4 interactive
// representative matrix.
func InteractiveCoverageMatrix() []InteractiveCoverageRow {
	out := make([]InteractiveCoverageRow, len(interactiveCoverage))
	copy(out, interactiveCoverage)
	return out
}

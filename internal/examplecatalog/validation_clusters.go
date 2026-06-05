package examplecatalog

// ValidationCluster describes a cross-fixture set used to validate parity
// fixes that touch shared rendering, layout, or transform behavior.
type ValidationCluster struct {
	ID          string
	Description string
	CaseIDs     []string
}

const (
	ValidationClusterLayoutText        = "layout-text"
	ValidationClusterImageMeshColorbar = "image-mesh-colorbar"
	ValidationClusterProjection3D      = "projection-3d"
	ValidationClusterPathStrokeMarker  = "path-stroke-marker"
	ValidationClusterMathText          = "mathtext"
	ValidationClusterStatsSpecialty    = "stats-specialty"
	ValidationClusterSignalUnitsVector = "signal-units-vector"
	ValidationClusterContour           = "contour-unstructured"
)

var validationClusters = []ValidationCluster{
	{
		ID:          ValidationClusterLayoutText,
		Description: "Layout, text metrics, figure labels, annotations, axes controls, and colorbar-adjacent composition.",
		CaseIDs: []string{
			"figure_labels_composition",
			"text_labels_strict",
			"axes_top_right_inverted",
			"axes_control_surface",
			"transform_coordinates",
			"annotation_composition",
			"legend_layout_matrix",
			"text_annotation_matrix",
			"colorbar_composition",
			"axisartist_showcase",
			"axes_grid1_showcase",
		},
	},
	{
		ID:          ValidationClusterImageMeshColorbar,
		Description: "Image sampling, mesh rasterization, norm/colorbar rendering, and related rasterizer behavior.",
		CaseIDs: []string{
			"image_heatmap",
			"imshow_clipped",
			"imshow_transformed",
			"spy_marker",
			"spy_image",
			"pcolor_flat",
			"pcolormesh_gouraud",
			"asinh_norm_image",
			"boundarynorm_pcolormesh",
			"collection_mutable_scalarmap",
			"colorbar_boundary_values",
			"colorbar_horizontal_ticks",
			"lognorm_imshow",
			"twoslope_norm_image",
			"colorbar_extensions",
			"colorbar_composition",
			"mesh_contour_tri",
			"arrays_showcase",
			"unstructured_showcase",
			"hist2d_weighted_density",
		},
	},
	{
		ID:          ValidationClusterProjection3D,
		Description: "3D toolkit, polar, geographic, radar, and skew projection behavior.",
		CaseIDs: []string{
			"mplot3d_basic",
			"mplot3d_terrain",
			"mplot3d_plot3d",
			"mplot3d_scatter3d",
			"mplot3d_surface3d",
			"mplot3d_wire3d",
			"mplot3d_trisurf3d",
			"mplot3d_bar3d",
			"mplot3d_voxels",
			"mplot3d_quiver3d",
			"mplot3d_errorbar3d",
			"mplot3d_stem3d",
			"mplot3d_fill_between3d",
			"mplot3d_contour3d",
			"mplot3d_contourf3d",
			"mplot3d_tricontour3d",
			"mplot3d_tricontourf3d",
			"mplot3d_bar2d_zdir",
			"mplot3d_text3d",
			"polar_axes",
			"geo_mollweide_axes",
			"geo_aitoff_axes",
			"geo_hammer_axes",
			"geo_lambert_axes",
			"radar_basic",
			"skewt_basic",
		},
	},
	{
		ID:          ValidationClusterPathStrokeMarker,
		Description: "Path construction, fills, strokes, markers, hatches, caps, joins, and rasterized edge behavior.",
		CaseIDs: []string{
			"fill_basic",
			"fill_stacked",
			"errorbar_basic",
			"boxplot_basic",
			"spy_marker",
			"patch_showcase",
			"patch_style_matrix",
			"plot_variants",
			"specialty_depth",
			"stem_plot",
			"units_dates",
			"vector_fields",
		},
	},
	{
		ID:          ValidationClusterMathText,
		Description: "MathText parsing, glyph metrics, stacked layout, large operators, delimiters, and inline text integration.",
		CaseIDs: []string{
			"mathtext_basic",
			"mathtext_fractions",
			"mathtext_integrals",
			"mathtext_matrices",
			"mathtext_inline_labels",
			"text_labels_strict",
		},
	},
	{
		ID:          ValidationClusterStatsSpecialty,
		Description: "Statistical and specialty artists with compound geometry, labels, legends, and table-like layout.",
		CaseIDs: []string{
			"boxplot_basic",
			"stat_variants",
			"specialty_depth",
			"stem_plot",
			"specialty_artists",
		},
	},
	{
		ID:          ValidationClusterSignalUnitsVector,
		Description: "Signal helpers, unit converters, date/category formatting, and vector-field glyph scaling.",
		CaseIDs: []string{
			"spectrum_variants",
			"units_overview",
			"units_dates",
			"units_categories",
			"units_custom_converter",
			"vector_fields",
		},
	},
	{
		ID:          ValidationClusterContour,
		Description: "Structured and unstructured contour/triangulation behavior, contour labels, and mixed mesh compositions.",
		CaseIDs: []string{
			"mesh_contour_tri",
			"unstructured_showcase",
			"arrays_showcase",
			"pcolor_flat",
			"pcolormesh_gouraud",
		},
	},
}

// ValidationClusters returns the named cross-fixture sets used by parity
// hardening.
func ValidationClusters() []ValidationCluster {
	out := make([]ValidationCluster, len(validationClusters))
	for i := range validationClusters {
		out[i] = validationClusters[i]
		out[i].CaseIDs = append([]string(nil), validationClusters[i].CaseIDs...)
	}
	return out
}

// LookupValidationCluster finds a validation cluster by ID.
func LookupValidationCluster(id string) (ValidationCluster, bool) {
	for _, cluster := range ValidationClusters() {
		if cluster.ID == id {
			return cluster, true
		}
	}
	return ValidationCluster{}, false
}
